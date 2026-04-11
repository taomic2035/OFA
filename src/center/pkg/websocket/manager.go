package websocket

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrAgentNotConnected   = errors.New("agent not connected")
	ErrAgentAlreadyConnected = errors.New("agent already connected")
	ErrMaxConnections      = errors.New("maximum connections reached")
	ErrSessionNotFound     = errors.New("session not found")
)

// ConnectionManager manages WebSocket connections from agents
type ConnectionManager struct {
	mu sync.RWMutex

	// Active connections (agent_id -> *AgentConnection)
	connections sync.Map

	// Session mapping (session_id -> agent_id)
	sessions sync.Map

	// Connection state tracking
	states sync.Map // agent_id -> *ConnectionState

	// Configuration
	config ConnectionConfig

	// Message channels
	outboundChan chan *OutboundMessage

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// AgentConnection represents an active agent WebSocket connection
type AgentConnection struct {
	AgentID   string
	SessionID string
	IdentityID string
	Status    string

	// WebSocket connection
	conn WebSocketConn

	// Message handling
	sendQueue chan []byte
	receiveHandler func(*WebSocketMessage)

	// Timing
	lastSeen    time.Time
	registeredAt time.Time

	// Resources and capabilities
	Resources   ResourcePayload
	Capabilities []CapabilityPayload
	DeviceInfo  DeviceInfoPayload
}

// WebSocketConn is the interface for WebSocket connection
type WebSocketConn interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
	IsClosed() bool
}

// OutboundMessage is a message to be sent to an agent
type OutboundMessage struct {
	AgentID string
	Message *WebSocketMessage
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(config ConnectionConfig) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConnectionManager{
		config:       config,
		outboundChan: make(chan *OutboundMessage, config.MessageQueueSize),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the connection manager background workers
func (m *ConnectionManager) Start() {
	go m.outboundDispatcher()
	go m.healthMonitor()
	log.Printf("WebSocket ConnectionManager started")
}

// Stop stops the connection manager
func (m *ConnectionManager) Stop() {
	m.cancel()

	// Close all connections
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*AgentConnection)
		conn.Close()
		return true
	})

	log.Printf("WebSocket ConnectionManager stopped")
}

// === Connection Management ===

// RegisterAgent registers a new agent connection
func (m *ConnectionManager) RegisterAgent(wsConn WebSocketConn, payload *RegisterPayload) (*AgentConnection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check max connections
	connCount := 0
	m.connections.Range(func(_, _ interface{}) bool {
		connCount++
		return true
	})
	if connCount >= m.config.MaxConnections {
		return nil, ErrMaxConnections
	}

	// Check if already connected
	if existing, ok := m.connections.Load(payload.AgentID); ok {
		existingConn := existing.(*AgentConnection)
		// If old connection is stale, replace it
		if time.Since(existingConn.lastSeen) > m.config.HeartbeatTimeout {
			existingConn.Close()
		} else {
			return nil, ErrAgentAlreadyConnected
		}
	}

	// Create new connection
	sessionID := generateSessionID()
	conn := &AgentConnection{
		AgentID:      payload.AgentID,
		SessionID:    sessionID,
		IdentityID:   payload.IdentityID,
		Status:       "connected",
		conn:         wsConn,
		sendQueue:    make(chan []byte, m.config.MessageQueueSize),
		lastSeen:     time.Now(),
		registeredAt: time.Now(),
		Capabilities: payload.Capabilities,
		DeviceInfo:   payload.DeviceInfo,
	}

	// Store connection
	m.connections.Store(payload.AgentID, conn)
	m.sessions.Store(sessionID, payload.AgentID)

	// Store state
	state := &ConnectionState{
		AgentID:      payload.AgentID,
		SessionID:    sessionID,
		IdentityID:   payload.IdentityID,
		Status:       "connected",
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
		Capabilities: payload.Capabilities,
		DeviceInfo:   payload.DeviceInfo,
	}
	m.states.Store(payload.AgentID, state)

	// Start send worker
	go m.sendWorker(conn)

	log.Printf("Agent registered: %s (session: %s, identity: %s)",
		payload.AgentID, sessionID, payload.IdentityID)

	return conn, nil
}

// UnregisterAgent removes an agent connection
func (m *ConnectionManager) UnregisterAgent(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connections.Load(agentID); ok {
		agentConn := conn.(*AgentConnection)
		agentConn.Close()
		m.connections.Delete(agentID)
		m.sessions.Delete(agentConn.SessionID)

		// Update state
		if state, ok := m.states.Load(agentID); ok {
			connState := state.(*ConnectionState)
			connState.Status = "disconnected"
		}

		log.Printf("Agent unregistered: %s", agentID)
	}
}

// GetConnection retrieves a connection by agent ID
func (m *ConnectionManager) GetConnection(agentID string) (*AgentConnection, error) {
	if conn, ok := m.connections.Load(agentID); ok {
		return conn.(*AgentConnection), nil
	}
	return nil, ErrAgentNotConnected
}

// GetConnectionBySession retrieves a connection by session ID
func (m *ConnectionManager) GetConnectionBySession(sessionID string) (*AgentConnection, error) {
	agentID, ok := m.sessions.Load(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}
	return m.GetConnection(agentID.(string))
}

// GetState retrieves connection state by agent ID
func (m *ConnectionManager) GetState(agentID string) *ConnectionState {
	if state, ok := m.states.Load(agentID); ok {
		return state.(*ConnectionState)
	}
	return nil
}

// ListConnections returns all active connections
func (m *ConnectionManager) ListConnections() []*ConnectionState {
	var states []*ConnectionState
	m.states.Range(func(_, value interface{}) bool {
		state := value.(*ConnectionState)
		if state.Status == "connected" {
			states = append(states, state)
		}
		return true
	})
	return states
}

// ConnectionCount returns the number of active connections
func (m *ConnectionManager) ConnectionCount() int {
	count := 0
	m.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// === Heartbeat Management ===

// HandleHeartbeat processes a heartbeat from an agent
func (m *ConnectionManager) HandleHeartbeat(agentID string, payload *HeartbeatPayload) error {
	conn, err := m.GetConnection(agentID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	conn.lastSeen = time.Now()
	conn.Status = payload.Status
	conn.Resources = payload.Resources
	conn.mu.Unlock()

	// Update state
	if state, ok := m.states.Load(agentID); ok {
		connState := state.(*ConnectionState)
		connState.LastSeen = time.Now()
		connState.Status = payload.Status
	}

	return nil
}

// IsAgentOnline checks if an agent is online
func (m *ConnectionManager) IsAgentOnline(agentID string) bool {
	conn, err := m.GetConnection(agentID)
	if err != nil {
		return false
	}
	return time.Since(conn.lastSeen) < m.config.HeartbeatTimeout
}

// === Message Sending ===

// SendMessage sends a message to an agent
func (m *ConnectionManager) SendMessage(agentID string, msg *WebSocketMessage) error {
	conn, err := m.GetConnection(agentID)
	if err != nil {
		return err
	}

	data, err := EncodeMessage(msg)
	if err != nil {
		return err
	}

	select {
	case conn.sendQueue <- data:
		return nil
	default:
		return errors.New("send queue full")
	}
}

// BroadcastMessage sends a message to all connected agents
func (m *ConnectionManager) BroadcastMessage(msg *WebSocketMessage) int {
	count := 0
	m.connections.Range(func(key, value interface{}) bool {
		agentID := key.(string)
		if m.SendMessage(agentID, msg) == nil {
			count++
		}
		return true
	})
	return count
}

// SendToIdentity sends a message to all agents bound to an identity
func (m *ConnectionManager) SendToIdentity(identityID string, msg *WebSocketMessage) int {
	count := 0
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*AgentConnection)
		if conn.IdentityID == identityID {
			if m.SendMessage(conn.AgentID, msg) == nil {
				count++
			}
		}
		return true
	})
	return count
}

// === Background Workers ===

// sendWorker handles sending messages to an agent
func (m *ConnectionManager) sendWorker(conn *AgentConnection) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case data := <-conn.sendQueue:
			if err := conn.conn.Send(data); err != nil {
				log.Printf("Failed to send to agent %s: %v", conn.AgentID, err)
				// Connection might be broken, mark for cleanup
				return
			}
		}
	}
}

// outboundDispatcher handles outbound message queue
func (m *ConnectionManager) outboundDispatcher() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case outbound := <-m.outboundChan:
			if err := m.SendMessage(outbound.AgentID, outbound.Message); err != nil {
				log.Printf("Failed to dispatch to %s: %v", outbound.AgentID, err)
			}
		}
	}
}

// healthMonitor monitors agent health and removes stale connections
func (m *ConnectionManager) healthMonitor() {
	ticker := time.NewTicker(m.config.HeartbeatTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAgentHealth()
		}
	}
}

// checkAgentHealth checks all agent connections for timeout
func (m *ConnectionManager) checkAgentHealth() {
	m.connections.Range(func(key, value interface{}) bool {
		agentID := key.(string)
		conn := value.(*AgentConnection)

		if time.Since(conn.lastSeen) > m.config.HeartbeatTimeout {
			log.Printf("Agent %s timed out (last seen: %v)", agentID, conn.lastSeen)
			m.UnregisterAgent(agentID)
		}
		return true
	})
}

// === Identity Binding ===

// BindIdentity binds an agent to an identity
func (m *ConnectionManager) BindIdentity(agentID, identityID string) error {
	conn, err := m.GetConnection(agentID)
	if err != nil {
		return err
	}

	conn.mu.Lock()
	conn.IdentityID = identityID
	conn.mu.Unlock()

	// Update state
	if state, ok := m.states.Load(agentID); ok {
		connState := state.(*ConnectionState)
		connState.IdentityID = identityID
	}

	// Notify agent
	msg := NewMessage(MsgTypeConfigUpdate, &ConfigUpdatePayload{
		AgentID: agentID,
		Config: map[string]string{
			"identity_id": identityID,
		},
	})
	return m.SendMessage(agentID, msg)
}

// UnbindIdentity unbinds an agent from its identity
func (m *ConnectionManager) UnbindIdentity(agentID string) error {
	return m.BindIdentity(agentID, "")
}

// GetAgentsForIdentity returns all agents bound to an identity
func (m *ConnectionManager) GetAgentsForIdentity(identityID string) []string {
	var agents []string
	m.connections.Range(func(key, value interface{}) bool {
		conn := value.(*AgentConnection)
		if conn.IdentityID == identityID {
			agents = append(agents, conn.AgentID)
		}
		return true
	})
	return agents
}

// === Helper Functions ===

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return "sess_" + time.Now().Format("20060102150405") + "_" + randomString(12)
}

// Close closes an agent connection
func (c *AgentConnection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.sendQueue)
}

// SetReceiveHandler sets the message receive handler
func (c *AgentConnection) SetReceiveHandler(handler func(*WebSocketMessage)) {
	c.receiveHandler = handler
}

// GetLastSeen returns the last seen time
func (c *AgentConnection) GetLastSeen() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastSeen
}

// GetIdentityID returns the bound identity ID
func (c *AgentConnection) GetIdentityID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IdentityID
}