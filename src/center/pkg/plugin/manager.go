// Package plugin - 插件系统
// 提供插件生命周期管理和钩子机制
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PluginState 插件状态
type PluginState string

const (
	PluginStateDisabled  PluginState = "disabled"  // 已禁用
	PluginStateEnabled   PluginState = "enabled"   // 已启用
	PluginStateRunning   PluginState = "running"   // 运行中
	PluginStateError     PluginState = "error"     // 错误
	PluginStateUpgrading PluginState = "upgrading" // 升级中
)

// PluginType 插件类型
type PluginType string

const (
	PluginTypeSkill    PluginType = "skill"    // 技能插件
	PluginTypeAdapter  PluginType = "adapter"  // 适配器插件
	PluginTypeStorage  PluginType = "storage"  // 存储插件
	PluginTypeAuth     PluginType = "auth"     // 认证插件
	PluginTypeNotifier PluginType = "notifier" // 通知插件
	PluginTypeAnalyzer PluginType = "analyzer" // 分析插件
	PluginTypeCustom   PluginType = "custom"   // 自定义插件
)

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Type        PluginType        `json:"type"`
	Tags        []string          `json:"tags"`
	Homepage    string            `json:"homepage,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	License     string            `json:"license"`
	Icon        string            `json:"icon,omitempty"`
	Config      map[string]string `json:"config,omitempty"`
	Dependencies []Dependency     `json:"dependencies,omitempty"`
	Permissions []string          `json:"permissions"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Dependency 依赖
type Dependency struct {
	ID      string `json:"id"`
	Version string `json:"version"` // 语义化版本约束
	Required bool   `json:"required"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	ID         string                 `json:"id"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`   // 执行优先级
	Settings   map[string]interface{} `json:"settings"`   // 插件设置
	AutoStart  bool                   `json:"auto_start"` // 自动启动
	RestartPolicy RestartPolicy       `json:"restart_policy"`
}

// RestartPolicy 重启策略
type RestartPolicy string

const (
	RestartNever    RestartPolicy = "never"    // 从不重启
	RestartOnFailure RestartPolicy = "on_failure" // 失败时重启
	RestartAlways   RestartPolicy = "always"   // 总是重启
)

// PluginInstance 插件实例
type PluginInstance struct {
	Info       *PluginInfo       `json:"info"`
	Config     *PluginConfig     `json:"config"`
	State      PluginState       `json:"state"`
	LoadedAt   *time.Time        `json:"loaded_at,omitempty"`
	StartedAt  *time.Time        `json:"started_at,omitempty"`
	StoppedAt  *time.Time        `json:"stopped_at,omitempty"`
	Error      string            `json:"error,omitempty"`
	Stats      *PluginStats      `json:"stats,omitempty"`
}

// PluginStats 插件统计
type PluginStats struct {
	Invocations   int64         `json:"invocations"`
	Successes     int64         `json:"successes"`
	Failures      int64         `json:"failures"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	LastError     string        `json:"last_error,omitempty"`
	LastErrorTime *time.Time    `json:"last_error_time,omitempty"`
}

// Plugin 插件接口
type Plugin interface {
	// Info 返回插件信息
	Info() *PluginInfo

	// Init 初始化插件
	Init(ctx context.Context, config *PluginConfig) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// Execute 执行插件功能
	Execute(ctx context.Context, input interface{}) (interface{}, error)

	// Health 健康检查
	Health(ctx context.Context) error

	// Cleanup 清理资源
	Cleanup(ctx context.Context) error
}

// HookFunc 钩子函数
type HookFunc func(ctx context.Context, input interface{}) (interface{}, error)

// HookType 钩子类型
type HookType string

const (
	HookBeforeTask   HookType = "before_task"   // 任务执行前
	HookAfterTask    HookType = "after_task"    // 任务执行后
	HookBeforeSkill  HookType = "before_skill"  // 技能执行前
	HookAfterSkill   HookType = "after_skill"   // 技能执行后
	HookOnAgentJoin  HookType = "on_agent_join"  // Agent加入
	HookOnAgentLeave HookType = "on_agent_leave" // Agent离开
	HookOnError      HookType = "on_error"      // 错误发生
	HookOnEvent      HookType = "on_event"      // 事件发生
)

// Hook 钩子
type Hook struct {
	ID       string    `json:"id"`
	Type     HookType  `json:"type"`
	Priority int       `json:"priority"`
	Handler  HookFunc  `json:"-"`
	PluginID string    `json:"plugin_id"`
	Enabled  bool      `json:"enabled"`
}

// Event 插件事件
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Source    string      `json:"source"` // 插件ID
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// PluginManager 插件管理器
type PluginManager struct {
	plugins    map[string]*PluginInstance
	pluginsMu  sync.RWMutex

	registry   *PluginRegistry
	loader     *PluginLoader
	hooks      map[HookType][]*Hook
	hooksMu    sync.RWMutex

	events     chan *Event
	eventSubs  map[string][]EventHandler
	eventsMu   sync.RWMutex

	config     ManagerConfig
	stats      *ManagerStats
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	PluginDir      string        `json:"plugin_dir"`
	MaxPlugins     int           `json:"max_plugins"`
	EnableHotLoad  bool          `json:"enable_hot_load"`
	AutoStart      bool          `json:"auto_start"`
	DefaultPolicy  RestartPolicy `json:"default_policy"`
	HealthCheckInt time.Duration `json:"health_check_interval"`
}

// DefaultManagerConfig 默认配置
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		PluginDir:      "./plugins",
		MaxPlugins:     100,
		EnableHotLoad:  true,
		AutoStart:      true,
		DefaultPolicy:  RestartOnFailure,
		HealthCheckInt: 30 * time.Second,
	}
}

// ManagerStats 管理器统计
type ManagerStats struct {
	TotalPlugins   int `json:"total_plugins"`
	RunningPlugins int `json:"running_plugins"`
	EnabledPlugins int `json:"enabled_plugins"`
	ErrorPlugins   int `json:"error_plugins"`
	HookCalls      int64 `json:"hook_calls"`
	EventPublished int64 `json:"event_published"`
}

// NewPluginManager 创建插件管理器
func NewPluginManager(config ManagerConfig) *PluginManager {
	return &PluginManager{
		plugins:   make(map[string]*PluginInstance),
		registry:  NewPluginRegistry(),
		loader:    NewPluginLoader(config.PluginDir),
		hooks:     make(map[HookType][]*Hook),
		events:    make(chan *Event, 1000),
		eventSubs: make(map[string][]EventHandler),
		config:    config,
		stats:     &ManagerStats{},
	}
}

// Register 注册插件
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	info := plugin.Info()
	if info.ID == "" {
		return fmt.Errorf("插件ID不能为空")
	}

	if _, exists := pm.plugins[info.ID]; exists {
		return fmt.Errorf("插件已存在: %s", info.ID)
	}

	if len(pm.plugins) >= pm.config.MaxPlugins {
		return fmt.Errorf("插件数量达到上限: %d", pm.config.MaxPlugins)
	}

	instance := &PluginInstance{
		Info:  info,
		Config: &PluginConfig{
			ID:        info.ID,
			Enabled:   false,
			Priority:  100,
			AutoStart: pm.config.AutoStart,
			RestartPolicy: pm.config.DefaultPolicy,
			Settings:  make(map[string]interface{}),
		},
		State: PluginStateDisabled,
		Stats: &PluginStats{},
	}

	pm.plugins[info.ID] = instance
	pm.stats.TotalPlugins++

	return nil
}

// Unregister 注销插件
func (pm *PluginManager) Unregister(pluginID string) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	instance, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	// 停止插件
	if instance.State == PluginStateRunning {
		ctx := context.Background()
		pm.stopPlugin(ctx, instance)
	}

	delete(pm.plugins, pluginID)
	pm.stats.TotalPlugins--

	return nil
}

// Enable 启用插件
func (pm *PluginManager) Enable(pluginID string) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	instance, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	instance.Config.Enabled = true
	instance.State = PluginStateEnabled
	pm.stats.EnabledPlugins++

	return nil
}

// Disable 禁用插件
func (pm *PluginManager) Disable(pluginID string) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	instance, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	if instance.State == PluginStateRunning {
		ctx := context.Background()
		pm.stopPlugin(ctx, instance)
	}

	instance.Config.Enabled = false
	instance.State = PluginStateDisabled
	pm.stats.EnabledPlugins--

	return nil
}

// Start 启动插件
func (pm *PluginManager) Start(ctx context.Context, pluginID string) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	instance, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	return pm.startPlugin(ctx, instance)
}

// startPlugin 内部启动插件
func (pm *PluginManager) startPlugin(ctx context.Context, instance *PluginInstance) error {
	if instance.State == PluginStateRunning {
		return nil
	}

	now := time.Now()
	instance.StartedAt = &now
	instance.State = PluginStateRunning
	pm.stats.RunningPlugins++

	return nil
}

// Stop 停止插件
func (pm *PluginManager) Stop(ctx context.Context, pluginID string) error {
	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	instance, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件不存在: %s", pluginID)
	}

	return pm.stopPlugin(ctx, instance)
}

// stopPlugin 内部停止插件
func (pm *PluginManager) stopPlugin(ctx context.Context, instance *PluginInstance) error {
	if instance.State != PluginStateRunning {
		return nil
	}

	now := time.Now()
	instance.StoppedAt = &now
	instance.State = PluginStateEnabled
	pm.stats.RunningPlugins--

	return nil
}

// Execute 执行插件
func (pm *PluginManager) Execute(ctx context.Context, pluginID string, input interface{}) (interface{}, error) {
	pm.pluginsMu.RLock()
	instance, exists := pm.plugins[pluginID]
	pm.pluginsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("插件不存在: %s", pluginID)
	}

	if instance.State != PluginStateRunning {
		return nil, fmt.Errorf("插件未运行: %s (状态: %s)", pluginID, instance.State)
	}

	// 更新统计
	instance.Stats.Invocations++

	start := time.Now()

	// 执行前置钩子
	result, err := pm.executeHooks(ctx, HookBeforeSkill, input)
	if err != nil {
		instance.Stats.Failures++
		instance.Stats.LastError = err.Error()
		instance.Stats.LastErrorTime = &start
		return nil, err
	}

	// 执行后置钩子
	_, err = pm.executeHooks(ctx, HookAfterSkill, result)
	if err != nil {
		instance.Stats.Failures++
		return result, err
	}

	instance.Stats.Successes++
	instance.Stats.TotalDuration += time.Since(start)
	if instance.Stats.Invocations > 0 {
		instance.Stats.AvgDuration = time.Duration(int64(instance.Stats.TotalDuration) / instance.Stats.Invocations)
	}

	return result, nil
}

// RegisterHook 注册钩子
func (pm *PluginManager) RegisterHook(hookType HookType, hook *Hook) error {
	pm.hooksMu.Lock()
	defer pm.hooksMu.Unlock()

	if hook.ID == "" {
		return fmt.Errorf("钩子ID不能为空")
	}

	hook.Enabled = true
	pm.hooks[hookType] = append(pm.hooks[hookType], hook)

	return nil
}

// UnregisterHook 注销钩子
func (pm *PluginManager) UnregisterHook(hookType HookType, hookID string) {
	pm.hooksMu.Lock()
	defer pm.hooksMu.Unlock()

	hooks := pm.hooks[hookType]
	newHooks := make([]*Hook, 0)
	for _, h := range hooks {
		if h.ID != hookID {
			newHooks = append(newHooks, h)
		}
	}
	pm.hooks[hookType] = newHooks
}

// executeHooks 执行钩子
func (pm *PluginManager) executeHooks(ctx context.Context, hookType HookType, input interface{}) (interface{}, error) {
	pm.hooksMu.RLock()
	hooks := pm.hooks[hookType]
	pm.hooksMu.RUnlock()

	var result interface{} = input
	var err error

	// 按优先级排序并执行
	for _, hook := range hooks {
		if !hook.Enabled {
			continue
		}
		result, err = hook.Handler(ctx, result)
		pm.stats.HookCalls++
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// PublishEvent 发布事件
func (pm *PluginManager) PublishEvent(event *Event) {
	select {
	case pm.events <- event:
		pm.stats.EventPublished++
	default:
		// 事件队列满，丢弃
	}
}

// SubscribeEvent 订阅事件
func (pm *PluginManager) SubscribeEvent(eventType string, handler EventHandler) {
	pm.eventsMu.Lock()
	defer pm.eventsMu.Unlock()

	pm.eventSubs[eventType] = append(pm.eventSubs[eventType], handler)
}

// GetPlugin 获取插件
func (pm *PluginManager) GetPlugin(pluginID string) (*PluginInstance, bool) {
	pm.pluginsMu.RLock()
	defer pm.pluginsMu.RUnlock()

	instance, exists := pm.plugins[pluginID]
	return instance, exists
}

// ListPlugins 列出插件
func (pm *PluginManager) ListPlugins() []*PluginInstance {
	pm.pluginsMu.RLock()
	defer pm.pluginsMu.RUnlock()

	plugins := make([]*PluginInstance, 0, len(pm.plugins))
	for _, instance := range pm.plugins {
		plugins = append(plugins, instance)
	}
	return plugins
}

// GetStats 获取统计
func (pm *PluginManager) GetStats() *ManagerStats {
	return pm.stats
}

// ExportConfig 导出配置
func (pm *PluginManager) ExportConfig() ([]byte, error) {
	pm.pluginsMu.RLock()
	defer pm.pluginsMu.RUnlock()

	config := struct {
		Plugins []*PluginConfig `json:"plugins"`
	}{
		Plugins: make([]*PluginConfig, 0),
	}

	for _, instance := range pm.plugins {
		config.Plugins = append(config.Plugins, instance.Config)
	}

	return json.MarshalIndent(config, "", "  ")
}

// ImportConfig 导入配置
func (pm *PluginManager) ImportConfig(data []byte) error {
	var config struct {
		Plugins []*PluginConfig `json:"plugins"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	pm.pluginsMu.Lock()
	defer pm.pluginsMu.Unlock()

	for _, cfg := range config.Plugins {
		if instance, exists := pm.plugins[cfg.ID]; exists {
			instance.Config = cfg
		}
	}

	return nil
}