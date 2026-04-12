// health.go
// OFA Center Health Check Framework (v9.3.0)

package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Name        string
	Status      HealthStatus
	Message     string
	LastCheck   time.Time
	Details     map[string]interface{}
	IsHealable  bool
	HealAttempt int
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status       HealthStatus
	Components   []ComponentHealth
	Timestamp    time.Time
	Version      string
	Uptime       time.Duration
	Alerts       []Alert
	Degradation  *DegradationStrategy
}

// HealthChecker 健康检查器
type HealthChecker struct {
	components  map[string]ComponentChecker
	alerts      *AlertManager
	degradation *DegradationManager
	mu          sync.RWMutex
	startTime   time.Time
	version     string
}

// ComponentChecker 组件检查器接口
type ComponentChecker interface {
	Name() string
	Check(ctx context.Context) (ComponentHealth, error)
	IsHealable() bool
	Heal(ctx context.Context) error
}

// HealableCheck 可自愈的检查器接口
type HealableCheck interface {
	ComponentChecker
	Heal(ctx context.Context) error
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		components:  make(map[string]ComponentChecker),
		alerts:      NewAlertManager(),
		degradation: NewDegradationManager(),
		startTime:   time.Now(),
		version:     version,
	}
}

// RegisterComponent 注册组件检查器
func (h *HealthChecker) RegisterComponent(checker ComponentChecker) {
	h.mu.Lock()
	h.components[checker.Name()] = checker
	h.mu.Unlock()
}

// Check 执行健康检查
func (h *HealthChecker) Check(ctx context.Context) *HealthCheckResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime),
		Components: make([]ComponentHealth, 0),
		Alerts:    h.alerts.GetPendingAlerts(),
	}

	overallStatus := HealthStatusHealthy

	for name, checker := range h.components {
		health, err := checker.Check(ctx)
		if err != nil {
			health.Status = HealthStatusUnhealthy
			health.Message = err.Error()
		}

		// 尝试自愈
		if health.Status == HealthStatusUnhealthy && health.IsHealable {
			if healable, ok := checker.(HealableCheck); ok {
				healErr := healable.Heal(ctx)
				health.HealAttempt++
				if healErr == nil {
					health.Status = HealthStatusHealthy
					health.Message = "Self-healed successfully"
				} else {
					health.Message = fmt.Sprintf("Heal failed: %v", healErr)
				}
			}
		}

		result.Components = append(result.Components, health)

		// 更新整体状态
		if health.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if health.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}

		// 发送告警
		if health.Status != HealthStatusHealthy {
			h.alerts.SendAlert(Alert{
				Component: name,
				Level:     AlertLevelWarning,
				Message:   health.Message,
				Time:      time.Now(),
			})
		}
	}

	result.Status = overallStatus

	if overallStatus != HealthStatusHealthy {
		result.Degradation = h.degradation.GetStrategy(overallStatus)
	}

	return result
}

// GetUptime 获取运行时间
func (h *HealthChecker) GetUptime() time.Duration {
	return time.Since(h.startTime)
}

// Alert 告警
type Alert struct {
	Component string
	Level     AlertLevel
	Message   string
	Time      time.Time
	Resolved  bool
}

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertManager 告警管理器
type AlertManager struct {
	alerts    []Alert
	listeners []AlertListener
	mu        sync.RWMutex
}

// AlertListener 告警监听器
type AlertListener func(alert Alert)

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:    make([]Alert, 0),
		listeners: make([]AlertListener, 0),
	}
}

// SendAlert 发送告警
func (a *AlertManager) SendAlert(alert Alert) {
	a.mu.Lock()
	a.alerts = append(a.alerts, alert)
	listeners := a.listeners
	a.mu.Unlock()

	for _, listener := range listeners {
		listener(alert)
	}
}

// AddListener 添加监听器
func (a *AlertManager) AddListener(listener AlertListener) {
	a.mu.Lock()
	a.listeners = append(a.listeners, listener)
	a.mu.Unlock()
}

// GetPendingAlerts 获取待处理告警
func (a *AlertManager) GetPendingAlerts() []Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	pending := make([]Alert, 0)
	for _, alert := range a.alerts {
		if !alert.Resolved {
			pending = append(pending, alert)
		}
	}
	return pending
}

// ResolveAlert 解决告警
func (a *AlertManager) ResolveAlert(component string) {
	a.mu.Lock()
	for i := range a.alerts {
		if a.alerts[i].Component == component && !a.alerts[i].Resolved {
			a.alerts[i].Resolved = true
		}
	}
	a.mu.Unlock()
}

// DegradationStrategy 降级策略
type DegradationStrategy struct {
	Name        string
	Description string
	Actions     []DegradationAction
}

// DegradationAction 降级动作
type DegradationAction struct {
	Component string
	Action    string
	Enabled   bool
}

// DegradationManager 降级管理器
type DegradationManager struct {
	strategies map[HealthStatus]*DegradationStrategy
	mu         sync.RWMutex
}

// NewDegradationManager 创建降级管理器
func NewDegradationManager() *DegradationManager {
	return &DegradationManager{
		strategies: map[HealthStatus]*DegradationStrategy{
			HealthStatusDegraded: &DegradationStrategy{
				Name:        "degraded_mode",
				Description: "Enable fallback behaviors for degraded state",
				Actions: []DegradationAction{
					{Component: "cache", Action: "use_l1_only", Enabled: true},
					{Component: "websocket", Action: "reduce_connections", Enabled: true},
					{Component: "llm", Action: "use_cached_responses", Enabled: true},
				},
			},
			HealthStatusUnhealthy: &DegradationStrategy{
				Name:        "emergency_mode",
				Description: "Minimal functionality to prevent total failure",
				Actions: []DegradationAction{
					{Component: "cache", Action: "disable_l2", Enabled: true},
					{Component: "websocket", Action: "reject_new_connections", Enabled: true},
					{Component: "llm", Action: "disable_streaming", Enabled: true},
					{Component: "sync", Action: "use_local_only", Enabled: true},
				},
			},
		},
	}
}

// GetStrategy 获取降级策略
func (d *DegradationManager) GetStrategy(status HealthStatus) *DegradationStrategy {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.strategies[status]
}

// RegisterStrategy 注册降级策略
func (d *DegradationManager) RegisterStrategy(status HealthStatus, strategy *DegradationStrategy) {
	d.mu.Lock()
	d.strategies[status] = strategy
	d.mu.Unlock()
}

// DatabaseHealthChecker 数据库健康检查器
type DatabaseHealthChecker struct {
	name     string
	pingFunc func(ctx context.Context) error
}

func NewDatabaseHealthChecker(name string, pingFunc func(ctx context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{name: name, pingFunc: pingFunc}
}

func (d *DatabaseHealthChecker) Name() string { return d.name }
func (d *DatabaseHealthChecker) IsHealable() bool { return false }

func (d *DatabaseHealthChecker) Check(ctx context.Context) (ComponentHealth, error) {
	err := d.pingFunc(ctx)
	status := HealthStatusHealthy
	message := "Connection OK"

	if err != nil {
		status = HealthStatusUnhealthy
		message = err.Error()
	}

	return ComponentHealth{
		Name:      d.name,
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		IsHealable: false,
	}, nil
}

func (d *DatabaseHealthChecker) Heal(ctx context.Context) error { return nil }

// RedisHealthChecker Redis健康检查器
type RedisHealthChecker struct {
	pingFunc func(ctx context.Context) error
}

func NewRedisHealthChecker(pingFunc func(ctx context.Context) error) *RedisHealthChecker {
	return &RedisHealthChecker{pingFunc: pingFunc}
}

func (r *RedisHealthChecker) Name() string { return "redis" }
func (r *RedisHealthChecker) IsHealable() bool { return true }

func (r *RedisHealthChecker) Check(ctx context.Context) (ComponentHealth, error) {
	err := r.pingFunc(ctx)
	status := HealthStatusHealthy
	message := "Redis OK"

	if err != nil {
		status = HealthStatusUnhealthy
		message = err.Error()
	}

	return ComponentHealth{
		Name:       "redis",
		Status:     status,
		Message:    message,
		LastCheck:  time.Now(),
		IsHealable: true,
	}, nil
}

func (r *RedisHealthChecker) Heal(ctx context.Context) error {
	// 尝试重新连接
	return r.pingFunc(ctx)
}

// WebSocketHealthChecker WebSocket健康检查器
type WebSocketHealthChecker struct {
	maxConnections int
	getCountFunc   func() int
}

func NewWebSocketHealthChecker(maxConnections int, getCountFunc func() int) *WebSocketHealthChecker {
	return &WebSocketHealthChecker{maxConnections: maxConnections, getCountFunc: getCountFunc}
}

func (w *WebSocketHealthChecker) Name() string { return "websocket" }
func (w *WebSocketHealthChecker) IsHealable() bool { return false }

func (w *WebSocketHealthChecker) Check(ctx context.Context) (ComponentHealth, error) {
	count := w.getCountFunc()
	status := HealthStatusHealthy
	message := fmt.Sprintf("Connections: %d/%d", count, w.maxConnections)

	if count > w.maxConnections*90/100 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("High connection count: %d/%d", count, w.maxConnections)
	}

	return ComponentHealth{
		Name:       "websocket",
		Status:     status,
		Message:    message,
		LastCheck:  time.Now(),
		Details:    map[string]interface{}{"connections": count, "max": w.maxConnections},
		IsHealable: false,
	}, nil
}

func (w *WebSocketHealthChecker) Heal(ctx context.Context) error { return nil }

// MemoryHealthChecker 内存健康检查器
type MemoryHealthChecker struct {
	maxMemoryMB   int
	getMemoryFunc func() int
}

func NewMemoryHealthChecker(maxMemoryMB int, getMemoryFunc func() int) *MemoryHealthChecker {
	return &MemoryHealthChecker{maxMemoryMB: maxMemoryMB, getMemoryFunc: getMemoryFunc}
}

func (m *MemoryHealthChecker) Name() string { return "memory" }
func (m *MemoryHealthChecker) IsHealable() bool { return true }

func (m *MemoryHealthChecker) Check(ctx context.Context) (ComponentHealth, error) {
	mem := m.getMemoryFunc()
	status := HealthStatusHealthy
	message := fmt.Sprintf("Memory: %dMB/%dMB", mem, m.maxMemoryMB)

	if mem > m.maxMemoryMB*80/100 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("High memory: %dMB/%dMB", mem, m.maxMemoryMB)
	}
	if mem > m.maxMemoryMB*95/100 {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Critical memory: %dMB/%dMB", mem, m.maxMemoryMB)
	}

	return ComponentHealth{
		Name:       "memory",
		Status:     status,
		Message:    message,
		LastCheck:  time.Now(),
		Details:    map[string]interface{}{"used_mb": mem, "max_mb": m.maxMemoryMB},
		IsHealable: true,
	}, nil
}

func (m *MemoryHealthChecker) Heal(ctx context.Context) error {
	// 尝试清理内存 (GC)
	// runtime.GC()
	return nil
}

// LLMHealthChecker LLM健康检查器
type LLMHealthChecker struct {
	checkFunc func(ctx context.Context) error
}

func NewLLMHealthChecker(checkFunc func(ctx context.Context) error) *LLMHealthChecker {
	return &LLMHealthChecker{checkFunc: checkFunc}
}

func (l *LLMHealthChecker) Name() string { return "llm" }
func (l *LLMHealthChecker) IsHealable() bool { return false }

func (l *LLMHealthChecker) Check(ctx context.Context) (ComponentHealth, error) {
	err := l.checkFunc(ctx)
	status := HealthStatusHealthy
	message := "LLM OK"

	if err != nil {
		status = HealthStatusDegraded
		message = err.Error()
	}

	return ComponentHealth{
		Name:       "llm",
		Status:     status,
		Message:    message,
		LastCheck:  time.Now(),
		IsHealable: false,
	}, nil
}

func (l *LLMHealthChecker) Heal(ctx context.Context) error { return nil }