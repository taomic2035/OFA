// Package performance provides connection pooling
package performance

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	// Pool sizing
	InitialSize     int `json:"initial_size"`     // Initial connections
	MaxSize         int `json:"max_size"`         // Maximum connections
	MinIdle         int `json:"min_idle"`         // Minimum idle connections
	MaxIdle         int `json:"max_idle"`         // Maximum idle connections

	// Timeouts
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`     // Max idle time
	MaxLifetime     time.Duration `json:"max_lifetime"`     // Max connection lifetime

	// Health check
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	PingTimeout         time.Duration `json:"ping_timeout"`

	// Circuit breaker
	CircuitBreakerThreshold int `json:"circuit_breaker_threshold"` // Failures before open
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"` // Time before retry
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		InitialSize:             5,
		MaxSize:                 100,
		MinIdle:                 5,
		MaxIdle:                 20,
		ConnectTimeout:          5 * time.Second,
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		IdleTimeout:             5 * time.Minute,
		MaxLifetime:             30 * time.Minute,
		HealthCheckInterval:     30 * time.Second,
		PingTimeout:             1 * time.Second,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// Connection represents a pooled connection
type Connection struct {
	ID          string
	Conn        net.Conn
	CreatedAt   time.Time
	LastUsedAt  time.Time
	IsHealthy   bool
	IsInUse     bool
	Pool        *ConnectionPool

	mu sync.RWMutex
}

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	config PoolConfig
	dialer func() (net.Conn, error)

	// Connections
	allConns   sync.Map // map[string]*Connection
	idleConns  chan *Connection

	// Counters
	totalConns     int64
	idleConnsCount int64
	activeConns    int64
	waitCount      int64

	// Circuit breaker state
	failures       int64
	lastFailure    time.Time
	circuitOpen    bool

	// Statistics
	totalRequests  int64
	successRequests int64
	failedRequests int64
	totalWaitTime  int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config PoolConfig, dialer func() (net.Conn, error)) (*ConnectionPool, error) {
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.InitialSize <= 0 {
		config.InitialSize = 5
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:     config,
		dialer:     dialer,
		idleConns:  make(chan *Connection, config.MaxIdle),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize connections
	for i := 0; i < config.InitialSize; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			pool.Close()
			return nil, err
		}
		pool.putIdle(conn)
	}

	// Start health checker
	go pool.healthChecker()

	// Start idle reaper
	go pool.idleReaper()

	return pool, nil
}

// Get retrieves a connection from the pool
func (p *ConnectionPool) Get() (*Connection, error) {
	p.totalRequests++

	// Check circuit breaker
	if p.isCircuitOpen() {
		p.failedRequests++
		return nil, errors.New("circuit breaker open")
	}

	start := time.Now()

	// Try to get from idle pool
	select {
	case conn := <-p.idleConns:
		p.idleConnsCount--

		// Check if connection is still valid
		if !p.isConnectionValid(conn) {
			p.removeConnection(conn)
			return p.Get()
		}

		conn.markInUse()
		p.activeConns++
		p.totalWaitTime += int64(time.Since(start))
		return conn, nil
	default:
	}

	// Create new connection if under limit
	if atomic.LoadInt64(&p.totalConns) < int64(p.config.MaxSize) {
		conn, err := p.createConnection()
		if err != nil {
			p.recordFailure()
			return nil, err
		}

		conn.markInUse()
		p.activeConns++
		p.totalWaitTime += int64(time.Since(start))
		return conn, nil
	}

	// Wait for available connection
	p.waitCount++

	select {
	case conn := <-p.idleConns:
		p.idleConnsCount--

		if !p.isConnectionValid(conn) {
			p.removeConnection(conn)
			p.waitCount--
			return p.Get()
		}

		conn.markInUse()
		p.activeConns++
		p.totalWaitTime += int64(time.Since(start))
		return conn, nil
	case <-time.After(p.config.ConnectTimeout):
		p.waitCount--
		p.failedRequests++
		return nil, errors.New("connection pool exhausted")
	}
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn *Connection) {
	conn.markIdle()
	p.activeConns--

	// Check if connection is still valid
	if !p.isConnectionValid(conn) {
		p.removeConnection(conn)
		return
	}

	// Check max lifetime
	if time.Since(conn.CreatedAt) > p.config.MaxLifetime {
		p.removeConnection(conn)
		return
	}

	// Return to idle pool
	p.putIdle(conn)
}

// createConnection creates a new connection
func (p *ConnectionPool) createConnection() (*Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadInt64(&p.totalConns) >= int64(p.config.MaxSize) {
		return nil, errors.New("max connections reached")
	}

	netConn, err := p.dialer()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		ID:         generateConnID(),
		Conn:       netConn,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		IsHealthy:  true,
		Pool:       p,
	}

	p.allConns.Store(conn.ID, conn)
	p.totalConns++

	return conn, nil
}

// putIdle puts connection in idle pool
func (p *ConnectionPool) putIdle(conn *Connection) {
	select {
	case p.idleConns <- conn:
		p.idleConnsCount++
	default:
		// Idle pool full, close connection
		p.removeConnection(conn)
	}
}

// removeConnection removes a connection from the pool
func (p *ConnectionPool) removeConnection(conn *Connection) {
	p.allConns.Delete(conn.ID)
	conn.Conn.Close()
	atomic.AddInt64(&p.totalConns, -1)
}

// isConnectionValid checks if connection is valid
func (p *ConnectionPool) isConnectionValid(conn *Connection) bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()

	if !conn.IsHealthy {
		return false
	}

	// Check idle timeout
	if time.Since(conn.LastUsedAt) > p.config.IdleTimeout {
		return false
	}

	return true
}

// isCircuitOpen checks if circuit breaker is open
func (p *ConnectionPool) isCircuitOpen() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.circuitOpen {
		return false
	}

	// Check if timeout has passed
	if time.Since(p.lastFailure) > p.config.CircuitBreakerTimeout {
		p.circuitOpen = false
		p.failures = 0
		return false
	}

	return true
}

// recordFailure records a failure for circuit breaker
func (p *ConnectionPool) recordFailure() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.failures++
	p.lastFailure = time.Now()

	if p.failures >= int64(p.config.CircuitBreakerThreshold) {
		p.circuitOpen = true
	}
}

// recordSuccess records a success
func (p *ConnectionPool) recordSuccess() {
	p.mu.Lock()
	p.failures = 0
	p.mu.Unlock()
	p.successRequests++
}

// healthChecker periodically checks connection health
func (p *ConnectionPool) healthChecker() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.checkHealth()
		}
	}
}

// checkHealth checks health of idle connections
func (p *ConnectionPool) checkHealth() {
	p.allConns.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)

		conn.mu.RLock()
		inUse := conn.IsInUse
		conn.mu.RUnlock()

		if !inUse {
			// Ping connection
			if !p.ping(conn) {
				conn.mu.Lock()
				conn.IsHealthy = false
				conn.mu.Unlock()
			}
		}

		return true
	})
}

// ping checks if connection is alive
func (p *ConnectionPool) ping(conn *Connection) bool {
	if conn.Conn == nil {
		return false
	}

	// Set deadline
	deadline := time.Now().Add(p.config.PingTimeout)
	if err := conn.Conn.SetReadDeadline(deadline); err != nil {
		return false
	}

	// Try to read one byte
	oneByte := make([]byte, 1)
	_, err := conn.Conn.Read(oneByte)
	if err != nil {
		return false
	}

	return true
}

// idleReaper removes idle connections that have been idle too long
func (p *ConnectionPool) idleReaper() {
	ticker := time.NewTicker(p.config.IdleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.reapIdle()
		}
	}
}

// reapIdle removes expired idle connections
func (p *ConnectionPool) reapIdle() {
	now := time.Now()

	p.allConns.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)

		conn.mu.RLock()
		inUse := conn.IsInUse
		lastUsed := conn.LastUsedAt
		createdAt := conn.CreatedAt
		conn.mu.RUnlock()

		if !inUse {
			// Check idle timeout
			if now.Sub(lastUsed) > p.config.IdleTimeout {
				p.removeConnection(conn)
			}

			// Check max lifetime
			if now.Sub(createdAt) > p.config.MaxLifetime {
				p.removeConnection(conn)
			}
		}

		return true
	})
}

// GetStats returns pool statistics
func (p *ConnectionPool) GetStats() map[string]interface{} {
	var avgWaitTime int64
	if p.totalRequests > 0 {
		avgWaitTime = p.totalWaitTime / p.totalRequests
	}

	var successRate float64
	if p.totalRequests > 0 {
		successRate = float64(p.successRequests) / float64(p.totalRequests)
	}

	return map[string]interface{}{
		"total_connections":  atomic.LoadInt64(&p.totalConns),
		"idle_connections":   atomic.LoadInt64(&p.idleConnsCount),
		"active_connections": atomic.LoadInt64(&p.activeConns),
		"waiting_requests":   atomic.LoadInt64(&p.waitCount),
		"total_requests":     p.totalRequests,
		"successful_requests": p.successRequests,
		"failed_requests":    p.failedRequests,
		"success_rate":       fmt.Sprintf("%.2f%%", successRate*100),
		"avg_wait_time_ms":   avgWaitTime / 1e6,
		"circuit_open":       p.circuitOpen,
		"failures":           p.failures,
	}
}

// Close closes the pool
func (p *ConnectionPool) Close() {
	p.cancel()

	p.allConns.Range(func(key, value interface{}) bool {
		conn := value.(*Connection)
		conn.Conn.Close()
		return true
	})
}

// Connection methods

func (c *Connection) markInUse() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsInUse = true
	c.LastUsedAt = time.Now()
}

func (c *Connection) markIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsInUse = false
	c.LastUsedAt = time.Now()
}

func (c *Connection) Release() {
	if c.Pool != nil {
		c.Pool.Put(c)
	}
}

func generateConnID() string {
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}

// GRPCPool wraps connection pool for gRPC
type GRPCPool struct {
	pool *ConnectionPool
}

// NewGRPCPool creates a gRPC connection pool
func NewGRPCPool(addr string, config PoolConfig) (*GRPCPool, error) {
	dialer := func() (net.Conn, error) {
		return net.DialTimeout("tcp", addr, config.ConnectTimeout)
	}

	pool, err := NewConnectionPool(config, dialer)
	if err != nil {
		return nil, err
	}

	return &GRPCPool{pool: pool}, nil
}

// Get gets a connection
func (p *GRPCPool) Get() (*Connection, error) {
	return p.pool.Get()
}

// Close closes the pool
func (p *GRPCPool) Close() {
	p.pool.Close()
}