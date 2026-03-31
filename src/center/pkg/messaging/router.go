// Package messaging - 消息路由模块
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RouteType 路由类型
type RouteType string

const (
	RouteDirect   RouteType = "direct"   // 直接路由
	RouteViaCenter RouteType = "center"  // 通过Center转发
	RouteBroadcast RouteType = "broadcast" // 广播路由
	RouteBalanced  RouteType = "balanced"  // 负载均衡路由
)

// RouteRule 路由规则
type RouteRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Match       RouteMatch        `json:"match"`       // 匹配条件
	Action      RouteAction       `json:"action"`      // 路由动作
	Priority    int               `json:"priority"`    // 优先级
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata"`
}

// RouteMatch 路由匹配条件
type RouteMatch struct {
	MessageType  MessageType `json:"message_type,omitempty"`
	Action       string      `json:"action,omitempty"`
	FromAgent    string      `json:"from_agent,omitempty"`
	ToAgent      string      `json:"to_agent,omitempty"`
	Group        string      `json:"group,omitempty"`
	PriorityMin  MessagePriority `json:"priority_min,omitempty"`
	PriorityMax  MessagePriority `json:"priority_max,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// RouteAction 路由动作
type RouteAction struct {
	Type        RouteType `json:"type"`
	Target      string    `json:"target"`      // 目标AgentID/GroupID
	Targets     []string  `json:"targets"`     // 多目标
	Strategy    string    `json:"strategy"`    // 负载均衡策略
	Transform   string    `json:"transform"`   // 消息转换
	Delay       time.Duration `json:"delay"`   // 延迟
	RetryPolicy RetryPolicy `json:"retry_policy"` // 重试策略
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries  int           `json:"max_retries"`
	Interval    time.Duration `json:"interval"`
	Backoff     float64       `json:"backoff"`     // 退避系数
	MaxInterval time.Duration `json:"max_interval"`
}

// RouteEntry 路由表条目
type RouteEntry struct {
	Destination string    `json:"destination"` // 目标AgentID
	NextHop     string    `json:"next_hop"`    // 下一跳
	HopCount    int       `json:"hop_count"`   // 跳数
	Latency     time.Duration `json:"latency"` // 延迟
	Bandwidth   int64     `json:"bandwidth"`   // 带宽
	Cost        int       `json:"cost"`        // 路由代价
	LastUpdated time.Time `json:"last_updated"`
	Valid       bool      `json:"valid"`
}

// RouterConfig 路由器配置
type RouterConfig struct {
	MaxRoutes        int           `json:"max_routes"`        // 最大路由数
	RouteTimeout     time.Duration `json:"route_timeout"`     // 路由超时
	UpdateInterval   time.Duration `json:"update_interval"`   // 更新间隔
	EnableDiscovery  bool          `json:"enable_discovery"`  // 启用发现
	EnableForwarding bool          `json:"enable_forwarding"` // 启用转发
	MaxHops          int           `json:"max_hops"`          // 最大跳数
}

// MessageRouter 消息路由器
type MessageRouter struct {
	config      RouterConfig
	localAgentID string
	p2pManager  *P2PManager
	routes      map[string]*RouteEntry   // 目标 -> 路由条目
	rules       map[string]*RouteRule    // 规则ID -> 规则
	agentIndex  map[string][]string      // 能力 -> AgentID列表
	stats       *RouterStats
	mu          sync.RWMutex
}

// RouterStats 路由统计
type RouterStats struct {
	TotalRouted    int64 `json:"total_routed"`
	TotalForwarded int64 `json:"total_forwarded"`
	TotalDropped   int64 `json:"total_dropped"`
	TotalRetried   int64 `json:"total_retried"`
	AvgLatency     time.Duration `json:"avg_latency"`
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter(config RouterConfig, localAgentID string, p2pManager *P2PManager) *MessageRouter {
	return &MessageRouter{
		config:      config,
		localAgentID: localAgentID,
		p2pManager:  p2pManager,
		routes:      make(map[string]*RouteEntry),
		rules:       make(map[string]*RouteRule),
		agentIndex:  make(map[string][]string),
		stats:       &RouterStats{},
	}
}

// Initialize 初始化
func (mr *MessageRouter) Initialize() error {
	// 加载默认路由规则
	mr.loadDefaultRules()
	return nil
}

// loadDefaultRules 加载默认规则
func (mr *MessageRouter) loadDefaultRules() {
	rules := []RouteRule{
		{
			ID:       "rule-direct-message",
			Name:     "直接消息路由",
			Priority: 100,
			Enabled:  true,
			Match:    RouteMatch{},
			Action:   RouteAction{Type: RouteDirect},
		},
		{
			ID:       "rule-broadcast",
			Name:     "广播消息路由",
			Priority: 90,
			Enabled:  true,
			Match: RouteMatch{
				MessageType: TypeBroadcast,
			},
			Action: RouteAction{Type: RouteBroadcast},
		},
		{
			ID:       "rule-multicast",
			Name:     "组播消息路由",
			Priority: 90,
			Enabled:  true,
			Match: RouteMatch{
				MessageType: TypeMulticast,
			},
			Action: RouteAction{Type: RouteDirect},
		},
		{
			ID:       "rule-request-response",
			Name:     "请求-响应路由",
			Priority: 95,
			Enabled:  true,
			Match: RouteMatch{
				MessageType: TypeRequest,
			},
			Action: RouteAction{
				Type: RouteDirect,
				RetryPolicy: RetryPolicy{
					MaxRetries:  3,
					Interval:    time.Second,
					Backoff:     2.0,
					MaxInterval: 10 * time.Second,
				},
			},
		},
	}

	mr.mu.Lock()
	for _, rule := range rules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		mr.rules[rule.ID] = &rule
	}
	mr.mu.Unlock()
}

// Route 路由消息
func (mr *MessageRouter) Route(ctx context.Context, msg *Message) error {
	// 查找匹配的规则
	rule := mr.findMatchingRule(msg)
	if rule == nil {
		return fmt.Errorf("未找到匹配的路由规则")
	}

	// 执行路由动作
	start := time.Now()
	err := mr.executeAction(ctx, msg, rule)

	// 更新统计
	mr.mu.Lock()
	mr.stats.TotalRouted++
	if err != nil {
		mr.stats.TotalDropped++
	}
	mr.mu.Unlock()

	return err
}

// findMatchingRule 查找匹配的规则
func (mr *MessageRouter) findMatchingRule(msg *Message) *RouteRule {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var bestMatch *RouteRule
	bestPriority := -1

	for _, rule := range mr.rules {
		if !rule.Enabled {
			continue
		}

		if mr.matchRule(rule, msg) && rule.Priority > bestPriority {
			bestMatch = rule
			bestPriority = rule.Priority
		}
	}

	return bestMatch
}

// matchRule 检查是否匹配规则
func (mr *MessageRouter) matchRule(rule *RouteRule, msg *Message) bool {
	match := rule.Match

	// 检查消息类型
	if match.MessageType != "" && match.MessageType != msg.Type {
		return false
	}

	// 检查动作
	if match.Action != "" && match.Action != msg.Action {
		return false
	}

	// 检查发送者
	if match.FromAgent != "" && match.FromAgent != msg.From {
		return false
	}

	// 检查接收者
	if match.ToAgent != "" && match.ToAgent != msg.To {
		return false
	}

	// 检查优先级范围
	if msg.Priority < match.PriorityMin || msg.Priority > match.PriorityMax {
		if match.PriorityMin != 0 || match.PriorityMax != 0 {
			return false
		}
	}

	return true
}

// executeAction 执行路由动作
func (mr *MessageRouter) executeAction(ctx context.Context, msg *Message, rule *RouteRule) error {
	action := rule.Action

	switch action.Type {
	case RouteDirect:
		return mr.routeDirect(ctx, msg, action)
	case RouteViaCenter:
		return mr.routeViaCenter(ctx, msg, action)
	case RouteBroadcast:
		return mr.routeBroadcast(ctx, msg, action)
	case RouteBalanced:
		return mr.routeBalanced(ctx, msg, action)
	default:
		return fmt.Errorf("未知路由类型: %s", action.Type)
	}
}

// routeDirect 直接路由
func (mr *MessageRouter) routeDirect(ctx context.Context, msg *Message, action RouteAction) error {
	target := msg.To
	if action.Target != "" {
		target = action.Target
	}

	// 检查是否有直接路由
	mr.mu.RLock()
	route, hasRoute := mr.routes[target]
	mr.mu.RUnlock()

	if hasRoute && route.Valid {
		// 通过下一跳转发
		if route.NextHop != "" && route.NextHop != mr.localAgentID {
			msg.To = route.NextHop
			return mr.p2pManager.Send(route.NextHop, msg)
		}
	}

	// 直接发送
	return mr.p2pManager.Send(target, msg)
}

// routeViaCenter 通过Center转发
func (mr *MessageRouter) routeViaCenter(ctx context.Context, msg *Message, action RouteAction) error {
	// 将消息发送到Center进行转发
	// 实际实现中会调用Center的转发接口
	return fmt.Errorf("Center转发暂未实现")
}

// routeBroadcast 广播路由
func (mr *MessageRouter) routeBroadcast(ctx context.Context, msg *Message, action RouteAction) error {
	return mr.p2pManager.Broadcast(msg)
}

// routeBalanced 负载均衡路由
func (mr *MessageRouter) routeBalanced(ctx context.Context, msg *Message, action RouteAction) error {
	// 选择最优目标
	targets := action.Targets
	if len(targets) == 0 {
		return fmt.Errorf("负载均衡路由无目标")
	}

	// 简单轮询
	selected := targets[0]
	minLoad := int64(1<<63 - 1)

	mr.mu.RLock()
	for _, t := range targets {
		if route, ok := mr.routes[t]; ok && route.Valid {
			// 选择负载最低的
			if int64(route.Latency) < minLoad {
				minLoad = int64(route.Latency)
				selected = t
			}
		}
	}
	mr.mu.RUnlock()

	msg.To = selected
	return mr.p2pManager.Send(selected, msg)
}

// AddRoute 添加路由
func (mr *MessageRouter) AddRoute(destination string, nextHop string, cost int) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if len(mr.routes) >= mr.config.MaxRoutes {
		return fmt.Errorf("路由表已满")
	}

	mr.routes[destination] = &RouteEntry{
		Destination: destination,
		NextHop:     nextHop,
		HopCount:    1,
		Cost:        cost,
		LastUpdated: time.Now(),
		Valid:       true,
	}

	return nil
}

// RemoveRoute 删除路由
func (mr *MessageRouter) RemoveRoute(destination string) {
	mr.mu.Lock()
	delete(mr.routes, destination)
	mr.mu.Unlock()
}

// UpdateRoute 更新路由
func (mr *MessageRouter) UpdateRoute(destination string, entry *RouteEntry) {
	mr.mu.Lock()
	entry.LastUpdated = time.Now()
	mr.routes[destination] = entry
	mr.mu.Unlock()
}

// GetRoute 获取路由
func (mr *MessageRouter) GetRoute(destination string) (*RouteEntry, error) {
	mr.mu.RLock()
	route, ok := mr.routes[destination]
	mr.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("路由不存在: %s", destination)
	}
	return route, nil
}

// ListRoutes 列出所有路由
func (mr *MessageRouter) ListRoutes() []*RouteEntry {
	mr.mu.RLock()
	routes := make([]*RouteEntry, 0, len(mr.routes))
	for _, r := range mr.routes {
		routes = append(routes, r)
	}
	mr.mu.RUnlock()
	return routes
}

// AddRule 添加路由规则
func (mr *MessageRouter) AddRule(rule *RouteRule) error {
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	mr.mu.Lock()
	mr.rules[rule.ID] = rule
	mr.mu.Unlock()

	return nil
}

// RemoveRule 删除路由规则
func (mr *MessageRouter) RemoveRule(ruleID string) {
	mr.mu.Lock()
	delete(mr.rules, ruleID)
	mr.mu.Unlock()
}

// ListRules 列出所有规则
func (mr *MessageRouter) ListRules() []*RouteRule {
	mr.mu.RLock()
	rules := make([]*RouteRule, 0, len(mr.rules))
	for _, r := range mr.rules {
		rules = append(rules, r)
	}
	mr.mu.RUnlock()
	return rules
}

// RegisterAgent 注册Agent能力
func (mr *MessageRouter) RegisterAgent(agentID string, capabilities []string) {
	mr.mu.Lock()
	for _, cap := range capabilities {
		if mr.agentIndex[cap] == nil {
			mr.agentIndex[cap] = make([]string, 0)
		}
		// 避免重复
		found := false
		for _, id := range mr.agentIndex[cap] {
			if id == agentID {
				found = true
				break
			}
		}
		if !found {
			mr.agentIndex[cap] = append(mr.agentIndex[cap], agentID)
		}
	}
	mr.mu.Unlock()
}

// UnregisterAgent 注销Agent
func (mr *MessageRouter) UnregisterAgent(agentID string) {
	mr.mu.Lock()
	for cap, agents := range mr.agentIndex {
		newAgents := make([]string, 0)
		for _, id := range agents {
			if id != agentID {
				newAgents = append(newAgents, id)
			}
		}
		mr.agentIndex[cap] = newAgents
	}
	// 同时删除相关路由
	delete(mr.routes, agentID)
	mr.mu.Unlock()
}

// FindAgentsByCapability 根据能力查找Agent
func (mr *MessageRouter) FindAgentsByCapability(capability string) []string {
	mr.mu.RLock()
	agents := mr.agentIndex[capability]
	if agents == nil {
		agents = []string{}
	}
	result := make([]string, len(agents))
	copy(result, agents)
	mr.mu.RUnlock()
	return result
}

// RouteToCapability 路由到具备某能力的Agent
func (mr *MessageRouter) RouteToCapability(ctx context.Context, msg *Message, capability string) error {
	agents := mr.FindAgentsByCapability(capability)
	if len(agents) == 0 {
		return fmt.Errorf("未找到具备能力 %s 的Agent", capability)
	}

	// 选择延迟最低的
	var bestAgent string
	bestLatency := time.Duration(1<<63 - 1)

	mr.mu.RLock()
	for _, agent := range agents {
		if route, ok := mr.routes[agent]; ok && route.Valid {
			if route.Latency < bestLatency {
				bestLatency = route.Latency
				bestAgent = agent
			}
		}
	}
	mr.mu.RUnlock()

	if bestAgent == "" {
		bestAgent = agents[0] // 默认选择第一个
	}

	msg.To = bestAgent
	return mr.p2pManager.Send(bestAgent, msg)
}

// Forward 转发消息
func (mr *MessageRouter) Forward(ctx context.Context, msg *Message) error {
	if !mr.config.EnableForwarding {
		return fmt.Errorf("消息转发未启用")
	}

	// 检查跳数
	if msg.RetryCount > mr.config.MaxHops {
		return fmt.Errorf("超过最大跳数")
	}

	// 查找路由
	route, err := mr.GetRoute(msg.To)
	if err != nil {
		return err
	}

	if !route.Valid {
		return fmt.Errorf("路由无效")
	}

	// 增加跳数
	msg.RetryCount++

	// 转发到下一跳
	return mr.p2pManager.Send(route.NextHop, msg)
}

// DiscoverRoutes 发现路由
func (mr *MessageRouter) DiscoverRoutes(ctx context.Context) error {
	if !mr.config.EnableDiscovery {
		return nil
	}

	// 向所有对等节点发送路由查询
	peers := mr.p2pManager.GetConnectedPeers()

	for _, peerID := range peers {
		query := &Message{
			Type:    TypeRequest,
			Action:  "route_discover",
			From:    mr.localAgentID,
			To:      peerID,
			Payload: map[string]interface{}{"query": "routes"},
		}

		// 发送查询
		mr.p2pManager.Send(peerID, query)
	}

	return nil
}

// GetStatistics 获取统计信息
func (mr *MessageRouter) GetStatistics() map[string]interface{} {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	return map[string]interface{}{
		"total_routed":     mr.stats.TotalRouted,
		"total_forwarded":  mr.stats.TotalForwarded,
		"total_dropped":    mr.stats.TotalDropped,
		"total_retried":    mr.stats.TotalRetried,
		"avg_latency":      mr.stats.AvgLatency,
		"route_count":      len(mr.routes),
		"rule_count":       len(mr.rules),
		"agent_index_size": len(mr.agentIndex),
	}
}

// ExportRoutes 导出路由表
func (mr *MessageRouter) ExportRoutes() ([]byte, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	return json.MarshalIndent(mr.routes, "", "  ")
}

// ImportRoutes 导入路由表
func (mr *MessageRouter) ImportRoutes(data []byte) error {
	routes := make(map[string]*RouteEntry)
	if err := json.Unmarshal(data, &routes); err != nil {
		return err
	}

	mr.mu.Lock()
	for k, v := range routes {
		mr.routes[k] = v
	}
	mr.mu.Unlock()

	return nil
}