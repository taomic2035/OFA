package sync

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// === 场景感知路由 (v3.2.0) ===
//
// Center 是永远在线的灵魂载体，场景感知路由确保：
// - 消息根据场景自动路由到最合适的设备
// - 支持自定义路由规则
// - 支持设备能力匹配
// - 支持优先级和时间条件

// SceneType - 场景类型
type SceneType string

const (
	SceneUnknown  SceneType = "unknown"
	SceneIdle     SceneType = "idle"
	SceneRunning  SceneType = "running"
	SceneWalking  SceneType = "walking"
	SceneDriving  SceneType = "driving"
	SceneCycling  SceneType = "cycling"
	SceneMeeting  SceneType = "meeting"
	SceneWorking  SceneType = "working"
	SceneResting  SceneType = "resting"
	SceneSleeping SceneType = "sleeping"
	SceneGaming   SceneType = "gaming"
	SceneCooking  SceneType = "cooking"
)

// SceneMessageType - 场景路由消息类型
type SceneMessageType string

const (
	SceneMessageTypeCommand       SceneMessageType = "command"
	SceneMessageTypeNotification  SceneMessageType = "notification"
	SceneMessageTypeData          SceneMessageType = "data"
	SceneMessageTypeSync          SceneMessageType = "sync"
	SceneMessageTypeAlert         SceneMessageType = "alert"
	SceneMessageTypeHealth        SceneMessageType = "health"
	SceneMessageTypeSocial        SceneMessageType = "social"
	SceneMessageTypeSystem        SceneMessageType = "system"
)

// RoutingRule - 路由规则
type RoutingRule struct {
	RuleID       string            `json:"rule_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Priority     int               `json:"priority"`      // 规则优先级，高优先级先匹配

	// 匹配条件
	Scenes       []SceneType       `json:"scenes"`        // 匹配的场景（空表示全部）
	MessageTypes []SceneMessageType `json:"message_types"` // 匹配的消息类型（空表示全部）
	DeviceTypes  []string          `json:"device_types"`  // 目标设备类型
	Conditions   []RuleCondition   `json:"conditions"`    // 附加条件

	// 路由动作
	Action       RoutingAction     `json:"action"`        // 路由动作
	TargetDevice string            `json:"target_device"` // 指定目标设备（可选）
	TargetRole   string            `json:"target_role"`   // 目标设备角色（primary/secondary）

	// 元数据
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// RuleCondition - 规则条件
type RuleCondition struct {
	Field    string      `json:"field"`     // 字段名
	Operator string      `json:"operator"`  // 操作符: eq, ne, gt, lt, gte, lte, in, contains
	Value    interface{} `json:"value"`     // 比较值
}

// RoutingAction - 路由动作
type RoutingAction string

const (
	ActionRoute       RoutingAction = "route"       // 路由到指定设备
	ActionBroadcast   RoutingAction = "broadcast"   // 广播到所有设备
	ActionPrioritize  RoutingAction = "prioritize"  // 按优先级路由
	ActionCapability  RoutingAction = "capability"  // 按能力匹配
	ActionFallback    RoutingAction = "fallback"    // 回退策略
	ActionDelay       RoutingAction = "delay"       // 延迟投递
	ActionFilter      RoutingAction = "filter"      // 过滤不投递
)

// RoutingContext - 路由上下文
type RoutingContext struct {
	IdentityID    string                 `json:"identity_id"`
	FromAgent     string                 `json:"from_agent"`
	MessageType   SceneMessageType       `json:"message_type"`
	Scene         SceneType              `json:"scene"`
	Priority      int                    `json:"priority"`
	Payload       map[string]interface{} `json:"payload"`
	DeviceStates  []*DeviceState         `json:"device_states"`
	Timestamp     time.Time              `json:"timestamp"`
}

// RoutingResult - 路由结果
type RoutingResult struct {
	TargetAgents  []string       `json:"target_agents"`
	Action        RoutingAction  `json:"action"`
	Reason        string         `json:"reason"`
	MatchedRule   *RoutingRule   `json:"matched_rule,omitempty"`
	Delay         time.Duration  `json:"delay,omitempty"`
}

// SceneRouter - 场景路由器
type SceneRouter struct {
	mu sync.RWMutex

	// 路由规则
	rules []*RoutingRule

	// 默认规则
	defaultRule *RoutingRule

	// 状态同步管理器
	stateManager *StateSyncManager

	// 设备管理器
	deviceManager *DeviceManager

	// 消息总线
	messageBus *MessageBus

	// 配置
	config SceneRouterConfig

	// 路由历史
	routingHistory []*RoutingRecord
}

// SceneRouterConfig - 路由器配置
type SceneRouterConfig struct {
	// 最大历史记录
	MaxHistorySize int
	// 默认路由动作
	DefaultAction RoutingAction
	// 是否启用智能路由
	SmartRouting bool
	// 低电量阈值
	LowBatteryThreshold int
}

// DefaultSceneRouterConfig 默认配置
func DefaultSceneRouterConfig() SceneRouterConfig {
	return SceneRouterConfig{
		MaxHistorySize:       1000,
		DefaultAction:        ActionPrioritize,
		SmartRouting:         true,
		LowBatteryThreshold:  20,
	}
}

// RoutingRecord - 路由记录
type RoutingRecord struct {
	Timestamp   time.Time        `json:"timestamp"`
	IdentityID  string           `json:"identity_id"`
	FromAgent   string           `json:"from_agent"`
	ToAgents    []string         `json:"to_agents"`
	MessageType SceneMessageType `json:"message_type"`
	Scene       SceneType        `json:"scene"`
	Action      RoutingAction    `json:"action"`
	RuleID      string           `json:"rule_id"`
}

// NewSceneRouter 创建场景路由器
func NewSceneRouter(config SceneRouterConfig) *SceneRouter {
	if config.MaxHistorySize == 0 {
		config = DefaultSceneRouterConfig()
	}

	router := &SceneRouter{
		rules:          make([]*RoutingRule, 0),
		config:         config,
		routingHistory: make([]*RoutingRecord, 0),
	}

	// 初始化默认规则
	router.initDefaultRules()

	return router
}

// SetStateManager 设置状态同步管理器
func (r *SceneRouter) SetStateManager(sm *StateSyncManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stateManager = sm
}

// SetDeviceManager 设置设备管理器
func (r *SceneRouter) SetDeviceManager(dm *DeviceManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deviceManager = dm
}

// SetMessageBus 设置消息总线
func (r *SceneRouter) SetMessageBus(mb *MessageBus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messageBus = mb
}

// === 规则管理 ===

// AddRule 添加路由规则
func (r *SceneRouter) AddRule(rule *RoutingRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	if rule.RuleID == "" {
		rule.RuleID = generateRuleID()
	}

	r.rules = append(r.rules, rule)

	// 按优先级排序
	r.sortRules()

	log.Printf("Routing rule added: %s (priority=%d)", rule.RuleID, rule.Priority)
}

// RemoveRule 移除路由规则
func (r *SceneRouter) RemoveRule(ruleID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rule := range r.rules {
		if rule.RuleID == ruleID {
			r.rules = append(r.rules[:i], r.rules[i+1:]...)
			log.Printf("Routing rule removed: %s", ruleID)
			return
		}
	}
}

// UpdateRule 更新路由规则
func (r *SceneRouter) UpdateRule(rule *RoutingRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.rules {
		if existing.RuleID == rule.RuleID {
			rule.UpdatedAt = time.Now()
			r.rules[i] = rule
			r.sortRules()
			log.Printf("Routing rule updated: %s", rule.RuleID)
			return
		}
	}
}

// GetRule 获取路由规则
func (r *SceneRouter) GetRule(ruleID string) *RoutingRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if rule.RuleID == ruleID {
			return rule
		}
	}
	return nil
}

// GetRules 获取所有规则
func (r *SceneRouter) GetRules() []*RoutingRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]*RoutingRule, len(r.rules))
	copy(rules, r.rules)
	return rules
}

// === 路由决策 ===

// Route 执行路由决策
func (r *SceneRouter) Route(ctx *RoutingContext) *RoutingResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 获取设备状态
	if r.stateManager != nil {
		ctx.DeviceStates = r.stateManager.GetIdentityDeviceStates(ctx.IdentityID)
	}

	// 匹配规则
	var matchedRule *RoutingRule
	for _, rule := range r.rules {
		if !rule.Enabled {
			continue
		}

		if r.matchRule(rule, ctx) {
			matchedRule = rule
			break
		}
	}

	// 如果没有匹配规则，使用默认策略
	if matchedRule == nil {
		return r.defaultRoute(ctx)
	}

	// 执行路由动作
	result := r.executeAction(matchedRule, ctx)
	result.MatchedRule = matchedRule

	// 记录路由历史
	r.recordRouting(ctx, result, matchedRule.RuleID)

	return result
}

// matchRule 匹配规则
func (r *SceneRouter) matchRule(rule *RoutingRule, ctx *RoutingContext) bool {
	// 匹配场景
	if len(rule.Scenes) > 0 {
		matched := false
		for _, scene := range rule.Scenes {
			if scene == ctx.Scene {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 匹配消息类型
	if len(rule.MessageTypes) > 0 {
		matched := false
		for _, mt := range rule.MessageTypes {
			if mt == ctx.MessageType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 匹配附加条件
	for _, cond := range rule.Conditions {
		if !r.matchCondition(cond, ctx) {
			return false
		}
	}

	return true
}

// matchCondition 匹配条件
func (r *SceneRouter) matchCondition(cond RuleCondition, ctx *RoutingContext) bool {
	var value interface{}
	found := false

	// 从上下文获取字段值
	switch cond.Field {
	case "priority":
		value = ctx.Priority
		found = true
	case "scene":
		value = ctx.Scene
		found = true
	case "message_type":
		value = ctx.MessageType
		found = true
	default:
		// 从 payload 获取
		if ctx.Payload != nil {
			value, found = ctx.Payload[cond.Field]
		}
	}

	if !found {
		return false
	}

	// 执行比较
	switch cond.Operator {
	case "eq":
		return value == cond.Value
	case "ne":
		return value != cond.Value
	case "gt":
		if v, ok := toFloat(value); ok {
			if c, ok := toFloat(cond.Value); ok {
				return v > c
			}
		}
	case "lt":
		if v, ok := toFloat(value); ok {
			if c, ok := toFloat(cond.Value); ok {
				return v < c
			}
		}
	case "gte":
		if v, ok := toFloat(value); ok {
			if c, ok := toFloat(cond.Value); ok {
				return v >= c
			}
		}
	case "lte":
		if v, ok := toFloat(value); ok {
			if c, ok := toFloat(cond.Value); ok {
				return v <= c
			}
		}
	case "in":
		if arr, ok := cond.Value.([]interface{}); ok {
			for _, item := range arr {
				if value == item {
					return true
				}
			}
		}
		return false
	case "contains":
		if str, ok := value.(string); ok {
			if substr, ok := cond.Value.(string); ok {
				return contains(str, substr)
			}
		}
	}

	return false
}

// executeAction 执行路由动作
func (r *SceneRouter) executeAction(rule *RoutingRule, ctx *RoutingContext) *RoutingResult {
	result := &RoutingResult{
		Action: rule.Action,
	}

	switch rule.Action {
	case ActionRoute:
		// 路由到指定设备
		if rule.TargetDevice != "" {
			result.TargetAgents = []string{rule.TargetDevice}
			result.Reason = "specified by rule"
		} else if rule.TargetRole == "primary" {
			// 路由到主设备
			if r.deviceManager != nil {
				primary := r.deviceManager.GetPrimaryDevice(ctx.IdentityID)
				if primary != nil {
					result.TargetAgents = []string{primary.AgentID}
					result.Reason = "primary device"
				}
			}
		}
		if len(result.TargetAgents) == 0 {
			// 回退到优先级最高设备
			result = r.prioritizeRoute(ctx)
		}

	case ActionBroadcast:
		// 广播到所有设备
		result.TargetAgents = r.getOnlineAgents(ctx)
		result.Reason = "broadcast to all devices"

	case ActionPrioritize:
		// 按优先级路由
		result = r.prioritizeRoute(ctx)

	case ActionCapability:
		// 按能力匹配
		result = r.capabilityRoute(rule.DeviceTypes, ctx)

	case ActionFallback:
		// 回退策略
		result = r.fallbackRoute(ctx)

	case ActionFilter:
		// 过滤不投递
		result.TargetAgents = []string{}
		result.Reason = "filtered by rule"

	case ActionDelay:
		// 延迟投递
		result = r.prioritizeRoute(ctx)
		result.Delay = 5 * time.Minute // 默认延迟5分钟
		result.Reason = "delayed delivery"
	}

	return result
}

// prioritizeRoute 优先级路由
func (r *SceneRouter) prioritizeRoute(ctx *RoutingContext) *RoutingResult {
	result := &RoutingResult{
		Action: ActionPrioritize,
	}

	// 获取设备并按优先级排序
	if r.deviceManager != nil {
		devices := r.deviceManager.GetDevicesByPriority(ctx.IdentityID)
		for _, device := range devices {
			if device.Status == DeviceStatusOnline {
				// 排除发送者
				if device.AgentID != ctx.FromAgent {
					result.TargetAgents = append(result.TargetAgents, device.AgentID)
				}
			}
		}
	}

	// 智能路由优化
	if r.config.SmartRouting && len(result.TargetAgents) > 0 {
		result.TargetAgents = r.optimizeTargets(ctx, result.TargetAgents)
	}

	if len(result.TargetAgents) > 0 {
		result.Reason = "prioritized by device priority"
	} else {
		result.Reason = "no available device"
	}

	return result
}

// capabilityRoute 能力匹配路由
func (r *SceneRouter) capabilityRoute(deviceTypes []string, ctx *RoutingContext) *RoutingResult {
	result := &RoutingResult{
		Action: ActionCapability,
	}

	// 按设备类型筛选
	for _, state := range ctx.DeviceStates {
		if !state.Online {
			continue
		}

		// 排除发送者
		if state.AgentID == ctx.FromAgent {
			continue
		}

		// 匹配设备类型
		if len(deviceTypes) > 0 {
			matched := false
			for _, dt := range deviceTypes {
				if state.DeviceType == dt {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		result.TargetAgents = append(result.TargetAgents, state.AgentID)
	}

	result.Reason = "matched by capability"
	return result
}

// fallbackRoute 回退路由
func (r *SceneRouter) fallbackRoute(ctx *RoutingContext) *RoutingResult {
	result := &RoutingResult{
		Action: ActionFallback,
	}

	// 先尝试优先级路由
	priorityResult := r.prioritizeRoute(ctx)
	if len(priorityResult.TargetAgents) > 0 {
		return priorityResult
	}

	// 如果没有在线设备，选择离线设备（消息将被缓存）
	if r.deviceManager != nil {
		devices := r.deviceManager.GetDevicesByIdentity(ctx.IdentityID)
		if len(devices) > 0 {
			result.TargetAgents = []string{devices[0].AgentID}
			result.Reason = "fallback to offline device"
		}
	}

	return result
}

// optimizeTargets 智能优化目标设备
func (r *SceneRouter) optimizeTargets(ctx *RoutingContext, targets []string) []string {
	if len(targets) == 0 {
		return targets
	}

	var optimized []string

	for _, agentID := range targets {
		// 查找设备状态
		var state *DeviceState
		for _, s := range ctx.DeviceStates {
			if s.AgentID == agentID {
				state = s
				break
			}
		}

		if state == nil {
			continue
		}

		// 场景感知优化
		switch ctx.Scene {
		case SceneRunning, SceneWalking:
			// 运动场景：优先手表，过滤手机
			if state.DeviceType == "watch" {
				optimized = append([]string{agentID}, optimized...) // 优先
			} else if state.DeviceType != "mobile" {
				optimized = append(optimized, agentID)
			}
			// 跳过手机

		case SceneDriving:
			// 驾驶场景：优先车载设备或手表
			if state.DeviceType == "car" || state.DeviceType == "watch" {
				optimized = append([]string{agentID}, optimized...)
			} else {
				optimized = append(optimized, agentID)
			}

		case SceneMeeting, SceneWorking:
			// 会议/工作场景：优先电脑或平板
			if state.DeviceType == "desktop" || state.DeviceType == "tablet" {
				optimized = append([]string{agentID}, optimized...)
			} else {
				optimized = append(optimized, agentID)
			}

		case SceneSleeping:
			// 睡眠场景：仅手表
			if state.DeviceType == "watch" {
				optimized = append(optimized, agentID)
			}

		default:
			// 默认：所有设备
			optimized = append(optimized, agentID)
		}

		// 低电量过滤
		if state.BatteryLevel < r.config.LowBatteryThreshold && !state.Charging {
			// 低电量设备降低优先级
			if len(optimized) > 1 {
				// 移到最后
				for i, id := range optimized {
					if id == agentID {
						optimized = append(optimized[:i], optimized[i+1:]...)
						optimized = append(optimized, agentID)
						break
					}
				}
			}
		}
	}

	// 如果优化后没有目标，返回原始目标
	if len(optimized) == 0 {
		return targets
	}

	return optimized
}

// defaultRoute 默认路由
func (r *SceneRouter) defaultRoute(ctx *RoutingContext) *RoutingResult {
	return &RoutingResult{
		TargetAgents: r.getOnlineAgents(ctx),
		Action:       r.config.DefaultAction,
		Reason:       "default routing",
	}
}

// getOnlineAgents 获取在线设备列表
func (r *SceneRouter) getOnlineAgents(ctx *RoutingContext) []string {
	var agents []string
	for _, state := range ctx.DeviceStates {
		if state.Online && state.AgentID != ctx.FromAgent {
			agents = append(agents, state.AgentID)
		}
	}
	return agents
}

// === 辅助方法 ===

func (r *SceneRouter) sortRules() {
	// 按优先级降序排序
	for i := 0; i < len(r.rules); i++ {
		for j := i + 1; j < len(r.rules); j++ {
			if r.rules[j].Priority > r.rules[i].Priority {
				r.rules[i], r.rules[j] = r.rules[j], r.rules[i]
			}
		}
	}
}

func (r *SceneRouter) initDefaultRules() {
	// 跑步场景路由规则
	r.rules = append(r.rules, &RoutingRule{
		RuleID:       "rule-running-watch",
		Name:         "Running Scene - Route to Watch",
		Priority:     100,
		Scenes:       []SceneType{SceneRunning, SceneWalking},
		MessageTypes: []SceneMessageType{SceneMessageTypeNotification, SceneMessageTypeHealth, SceneMessageTypeSocial},
		DeviceTypes:  []string{"watch"},
		Action:       ActionRoute,
		Enabled:      true,
	})

	// 驾驶场景路由规则
	r.rules = append(r.rules, &RoutingRule{
		RuleID:       "rule-driving-safe",
		Name:         "Driving Scene - Safe Routing",
		Priority:     100,
		Scenes:       []SceneType{SceneDriving},
		MessageTypes: []SceneMessageType{SceneMessageTypeNotification, SceneMessageTypeSocial},
		Action:       ActionDelay,
		Enabled:      true,
	})

	// 会议场景路由规则
	r.rules = append(r.rules, &RoutingRule{
		RuleID:       "rule-meeting-filter",
		Name:         "Meeting Scene - Filter Social",
		Priority:     90,
		Scenes:       []SceneType{SceneMeeting},
		MessageTypes: []SceneMessageType{SceneMessageTypeSocial},
		Conditions:   []RuleCondition{{Field: "priority", Operator: "lt", Value: 2}},
		Action:       ActionFilter,
		Enabled:      true,
	})

	// 健康告警路由规则
	r.rules = append(r.rules, &RoutingRule{
		RuleID:       "rule-health-alert",
		Name:         "Health Alert - Route to All",
		Priority:     200,
		MessageTypes: []SceneMessageType{SceneMessageTypeHealth, SceneMessageTypeAlert},
		Action:       ActionBroadcast,
		Enabled:      true,
	})

	// 睡眠场景路由规则
	r.rules = append(r.rules, &RoutingRule{
		RuleID:       "rule-sleeping-quiet",
		Name:         "Sleeping Scene - Quiet Mode",
		Priority:     95,
		Scenes:       []SceneType{SceneSleeping},
		MessageTypes: []SceneMessageType{SceneMessageTypeNotification, SceneMessageTypeSocial},
		Conditions:   []RuleCondition{{Field: "priority", Operator: "lt", Value: 3}},
		Action:       ActionFilter,
		Enabled:      true,
	})

	// 默认规则
	r.defaultRule = &RoutingRule{
		RuleID:   "rule-default",
		Name:     "Default Rule",
		Priority: 0,
		Action:   ActionPrioritize,
		Enabled:  true,
	}
}

func (r *SceneRouter) recordRouting(ctx *RoutingContext, result *RoutingResult, ruleID string) {
	record := &RoutingRecord{
		Timestamp:   time.Now(),
		IdentityID:  ctx.IdentityID,
		FromAgent:   ctx.FromAgent,
		ToAgents:    result.TargetAgents,
		MessageType: ctx.MessageType,
		Scene:       ctx.Scene,
		Action:      result.Action,
		RuleID:      ruleID,
	}

	r.routingHistory = append(r.routingHistory, record)

	// 限制历史记录大小
	if len(r.routingHistory) > r.config.MaxHistorySize {
		r.routingHistory = r.routingHistory[len(r.routingHistory)-r.config.MaxHistorySize:]
	}
}

// GetRoutingHistory 获取路由历史
func (r *SceneRouter) GetRoutingHistory(identityID string, limit int) []*RoutingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var history []*RoutingRecord
	for i := len(r.routingHistory) - 1; i >= 0 && len(history) < limit; i-- {
		record := r.routingHistory[i]
		if identityID == "" || record.IdentityID == identityID {
			history = append(history, record)
		}
	}

	return history
}

// === 统计 ===

// RoutingStats 路由统计
type RoutingStats struct {
	TotalRoutings   int            `json:"total_routings"`
	ByAction        map[string]int `json:"by_action"`
	ByScene         map[string]int `json:"by_scene"`
	ByMessageType   map[string]int `json:"by_message_type"`
	TotalRules      int            `json:"total_rules"`
	EnabledRules    int            `json:"enabled_rules"`
}

// GetStats 获取路由统计
func (r *SceneRouter) GetStats() *RoutingStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &RoutingStats{
		ByAction:      make(map[string]int),
		ByScene:       make(map[string]int),
		ByMessageType: make(map[string]int),
	}

	stats.TotalRoutings = len(r.routingHistory)

	for _, record := range r.routingHistory {
		stats.ByAction[string(record.Action)]++
		stats.ByScene[string(record.Scene)]++
		stats.ByMessageType[string(record.MessageType)]++
	}

	stats.TotalRules = len(r.rules)
	for _, rule := range r.rules {
		if rule.Enabled {
			stats.EnabledRules++
		}
	}

	return stats
}

// ToJSON 序列化规则
func (rule *RoutingRule) ToJSON() ([]byte, error) {
	return json.Marshal(rule)
}

// FromJSON 解析规则
func RoutingRuleFromJSON(data []byte) (*RoutingRule, error) {
	var rule RoutingRule
	err := json.Unmarshal(data, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// === 辅助函数 ===

func generateRuleID() string {
	return "rule_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}