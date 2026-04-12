// gateway.go
// OFA Center API Gateway (v9.1.0)

package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// GatewayMode Gateway模式
type GatewayMode string

const (
	GatewayModeStandard GatewayMode = "standard"
	GatewayModeStrict   GatewayMode = "strict"
	GatewayModeRelaxed  GatewayMode = "relaxed"
)

// RateLimitAlgorithm 限流算法
type RateLimitAlgorithm string

const (
	RateLimitTokenBucket   RateLimitAlgorithm = "token_bucket"
	RateLimitSlidingWindow RateLimitAlgorithm = "sliding_window"
	RateLimitFixedWindow   RateLimitAlgorithm = "fixed_window"
)

// APIGateway API Gateway
type APIGateway struct {
	router         *Router
	rateLimiter    RateLimiter
	circuitBreaker *CircuitBreaker
	responseCache  *ResponseCache
	middlewares    []Middleware
	config         GatewayConfig
	mu             sync.RWMutex
}

// GatewayConfig Gateway配置
type GatewayConfig struct {
	Mode               GatewayMode
	RateLimitAlgorithm RateLimitAlgorithm
	RateLimitRPS       int           // requests per second
	RateLimitBurst     int           // burst capacity
	CircuitBreakerConfig CircuitBreakerConfig
	CacheTTL           time.Duration
	CacheMaxSize       int
	EnableCache        bool
	EnableRateLimit    bool
	EnableCircuitBreaker bool
}

// DefaultGatewayConfig 默认Gateway配置
func DefaultGatewayConfig() GatewayConfig {
	return GatewayConfig{
		Mode:               GatewayModeStandard,
		RateLimitAlgorithm: RateLimitTokenBucket,
		RateLimitRPS:       100,
		RateLimitBurst:     20,
		CircuitBreakerConfig: DefaultCircuitBreakerConfig(),
		CacheTTL:           5 * time.Minute,
		CacheMaxSize:       1000,
		EnableCache:        true,
		EnableRateLimit:    true,
		EnableCircuitBreaker: true,
	}
}

// NewAPIGateway 创建API Gateway
func NewAPIGateway(config GatewayConfig) *APIGateway {
	gw := &APIGateway{
		router:      NewRouter(),
		middlewares: make([]Middleware, 0),
		config:      config,
	}

	if config.EnableRateLimit {
		switch config.RateLimitAlgorithm {
		case RateLimitTokenBucket:
			gw.rateLimiter = NewTokenBucketRateLimiter(config.RateLimitRPS, config.RateLimitBurst)
		case RateLimitSlidingWindow:
			gw.rateLimiter = NewSlidingWindowRateLimiter(config.RateLimitRPS, 1*time.Minute)
		default:
			gw.rateLimiter = NewTokenBucketRateLimiter(config.RateLimitRPS, config.RateLimitBurst)
		}
	}

	if config.EnableCircuitBreaker {
		gw.circuitBreaker = NewCircuitBreaker(config.CircuitBreakerConfig)
	}

	if config.EnableCache {
		gw.responseCache = NewResponseCache(config.CacheMaxSize, config.CacheTTL)
	}

	return gw
}

// Use 添加中间件
func (g *APIGateway) Use(middleware Middleware) {
	g.mu.Lock()
	g.middlewares = append(g.middlewares, middleware)
	g.mu.Unlock()
}

// Handle 处理请求
func (g *APIGateway) Handle(pattern string, handler http.Handler) {
	g.router.Register(pattern, handler)
}

// HandleFunc 处理请求函数
func (g *APIGateway) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.router.Register(pattern, handler)
}

// ServeHTTP 实现http.Handler接口
func (g *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 应用中间件链
	handler := g.router
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		handler = g.middlewares[i](handler)
	}

	// 速率限制检查
	if g.rateLimiter != nil {
		if !g.rateLimiter.Allow(r) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	// 熔断器检查
	if g.circuitBreaker != nil {
		state := g.circuitBreaker.GetState()
		if state == CircuitStateOpen {
			http.Error(w, "Service unavailable (circuit breaker open)", http.StatusServiceUnavailable)
			return
		}
	}

	// 响应缓存检查
	if g.responseCache != nil {
		cached, found := g.responseCache.Get(r)
		if found {
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached)
			return
		}
		w.Header().Set("X-Cache", "MISS")
	}

	// 记录请求开始时间
	start := time.Now()

	// 包装ResponseWriter以捕获响应
	wrapped := &responseWriterWrapper{ResponseWriter: w, body: make([]byte, 0)}

	// 执行请求
	handler.ServeHTTP(wrapped, r)

	// 更新熔断器状态
	if g.circuitBreaker != nil {
		duration := time.Since(start)
		if wrapped.statusCode >= 500 {
			g.circuitBreaker.RecordFailure()
		} else {
			g.circuitBreaker.RecordSuccess(duration)
		}
	}

	// 缓存成功响应
	if g.responseCache != nil && wrapped.statusCode < 500 && wrapped.statusCode >= 200 {
		g.responseCache.Set(r, wrapped.body)
	}
}

// responseWriterWrapper 包装ResponseWriter
type responseWriterWrapper struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Middleware 中间件函数类型
type Middleware func(http.Handler) http.Handler

// Router 路由器
type Router struct {
	routes map[string]http.Handler
	mu     sync.RWMutex
}

// NewRouter 创建路由器
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]http.Handler),
	}
}

// Register 注册路由
func (r *Router) Register(pattern string, handler http.Handler) {
	r.mu.Lock()
	r.routes[pattern] = handler
	r.mu.Unlock()
}

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	handler, ok := r.routes[req.URL.Path]
	r.mu.RUnlock()

	if ok {
		handler.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(r *http.Request) bool
	Reset()
}

// TokenBucketRateLimiter 令牌桶限流器
type TokenBucketRateLimiter struct {
	rate     int       // 令牌生成速率 (tokens/second)
	capacity int       // 桶容量
	tokens   float64   // 当前令牌数
	lastTime time.Time // 上次更新时间
	mu       sync.Mutex
}

// NewTokenBucketRateLimiter 创建令牌桶限流器
func NewTokenBucketRateLimiter(rate, capacity int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (t *TokenBucketRateLimiter) Allow(r *http.Request) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
 elapsed := now.Sub(t.lastTime).Seconds()
	t.tokens += elapsed * float64(t.rate)
	if t.tokens > float64(t.capacity) {
		t.tokens = float64(t.capacity)
	}
	t.lastTime = now

	if t.tokens >= 1.0 {
		t.tokens -= 1.0
		return true
	}
	return false
}

// Reset 重置限流器
func (t *TokenBucketRateLimiter) Reset() {
	t.mu.Lock()
	t.tokens = float64(t.capacity)
	t.lastTime = time.Now()
	t.mu.Unlock()
}

// SlidingWindowRateLimiter 滑动窗口限流器
type SlidingWindowRateLimiter struct {
	rate    int
	window  time.Duration
	requests map[string]*windowCounter
	mu      sync.Mutex
}

// windowCounter 窗口计数器
type windowCounter struct {
	count    int
	windowStart time.Time
}

// NewSlidingWindowRateLimiter 创建滑动窗口限流器
func NewSlidingWindowRateLimiter(rate int, window time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		rate:     rate,
		window:   window,
		requests: make(map[string]*windowCounter),
	}
}

// Allow 检查是否允许请求
func (s *SlidingWindowRateLimiter) Allow(r *http.Request) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := r.URL.Path
	now := time.Now()

	counter, ok := s.requests[key]
	if !ok {
		counter = &windowCounter{count: 0, windowStart: now}
		s.requests[key] = counter
	}

	// 检查窗口是否过期
	if now.Sub(counter.windowStart) > s.window {
		counter.count = 0
		counter.windowStart = now
	}

	if counter.count < s.rate {
		counter.count++
		return true
	}
	return false
}

// Reset 重置限流器
func (s *SlidingWindowRateLimiter) Reset() {
	s.mu.Lock()
	s.requests = make(map[string]*windowCounter)
	s.mu.Unlock()
}

// ResponseCache 响应缓存
type ResponseCache struct {
	maxSize int
	ttl     time.Duration
	cache   map[string]*cachedResponse
	mu      sync.RWMutex
}

// cachedResponse 缓存的响应
type cachedResponse struct {
	body      []byte
	expiresAt time.Time
}

// NewResponseCache 创建响应缓存
func NewResponseCache(maxSize int, ttl time.Duration) *ResponseCache {
	return &ResponseCache{
		maxSize: maxSize,
		ttl:     ttl,
		cache:   make(map[string]*cachedResponse),
	}
}

// Get 获取缓存响应
func (c *ResponseCache) Get(r *http.Request) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(r)
	cached, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(cached.expiresAt) {
		return nil, false
	}

	return cached.body, true
}

// Set 设置缓存响应
func (c *ResponseCache) Set(r *http.Request, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小
	if len(c.cache) >= c.maxSize {
		// 删除过期条目
		for key, cached := range c.cache {
			if time.Now().After(cached.expiresAt) {
				delete(c.cache, key)
			}
		}
		// 如果还是满了，删除最早的条目
		if len(c.cache) >= c.maxSize {
			var oldestKey string
			oldestTime := time.Now().Add(24 * time.Hour)
			for key, cached := range c.cache {
				if cached.expiresAt.Before(oldestTime) {
					oldestTime = cached.expiresAt
					oldestKey = key
				}
			}
			if oldestKey != "" {
				delete(c.cache, oldestKey)
			}
		}
	}

	key := c.cacheKey(r)
	c.cache[key] = &cachedResponse{
		body:      body,
	 expiresAt: time.Now().Add(c.ttl),
	}
}

// cacheKey 生成缓存键
func (c *ResponseCache) cacheKey(r *http.Request) string {
	return r.Method + ":" + r.URL.Path + ":" + r.URL.RawQuery
}

// Clear 清除缓存
func (c *ResponseCache) Clear() {
	c.mu.Lock()
	c.cache = make(map[string]*cachedResponse)
	c.mu.Unlock()
}

// CircuitState 熔断器状态
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold   int           // 失败阈值
	SuccessThreshold   int           // 成功阈值 (半开状态)
	Timeout            time.Duration // 开启状态超时时间
	MinRequestCount    int           // 最小请求计数
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            30 * time.Second,
		MinRequestCount:    10,
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitState
	failureCount    int
	successCount    int
	requestCount    int
	lastFailureTime time.Time
	mu              sync.Mutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitStateClosed,
	}
}

// GetState 获取当前状态
func (c *CircuitBreaker) GetState() CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否需要从开启状态转为半开状态
	if c.state == CircuitStateOpen {
		if time.Since(c.lastFailureTime) > c.config.Timeout {
			c.state = CircuitStateHalfOpen
			c.successCount = 0
		}
	}

	return c.state
}

// RecordSuccess 记录成功
func (c *CircuitBreaker) RecordSuccess(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestCount++

	switch c.state {
	case CircuitStateClosed:
		// 正常状态，重置失败计数
		c.failureCount = 0
	case CircuitStateHalfOpen:
		c.successCount++
		if c.successCount >= c.config.SuccessThreshold {
			c.state = CircuitStateClosed
			c.failureCount = 0
			c.requestCount = 0
		}
	}
}

// RecordFailure 记录失败
func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.failureCount++
	c.requestCount++
	c.lastFailureTime = time.Now()

	switch c.state {
	case CircuitStateClosed:
		if c.requestCount >= c.config.MinRequestCount && c.failureCount >= c.config.FailureThreshold {
			c.state = CircuitStateOpen
		}
	case CircuitStateHalfOpen:
		c.state = CircuitStateOpen
		c.successCount = 0
	}
}

// Reset 重置熔断器
func (c *CircuitBreaker) Reset() {
	c.mu.Lock()
	c.state = CircuitStateClosed
	c.failureCount = 0
	c.successCount = 0
	c.requestCount = 0
	c.mu.Unlock()
}

// RequestLogger 请求日志中间件
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			fmt.Printf("[%s] %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, duration)
		})
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(cb *CircuitBreaker) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state := cb.GetState()
			if state == CircuitStateOpen {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}