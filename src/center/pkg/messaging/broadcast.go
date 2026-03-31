// Package messaging - 消息广播模块
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// BroadcastMode 广播模式
type BroadcastMode string

const (
	BroadcastAll      BroadcastMode = "all"      // 广播给所有Agent
	BroadcastOnline   BroadcastMode = "online"   // 仅广播给在线Agent
	BroadcastType     BroadcastMode = "type"     // 按类型广播
	BroadcastRegion   BroadcastMode = "region"   // 按区域广播
	BroadcastCapability BroadcastMode = "capability" // 按能力广播
	BroadcastExclude  BroadcastMode = "exclude"  // 排除某些Agent
)

// BroadcastScope 广播范围
type BroadcastScope struct {
	Mode        BroadcastMode `json:"mode"`
	Filter      string        `json:"filter,omitempty"`      // 过滤条件
	ExcludeList []string      `json:"exclude_list,omitempty"` // 排除列表
	IncludeList []string      `json:"include_list,omitempty"` // 包含列表
	MaxRecipients int         `json:"max_recipients,omitempty"` // 最大接收者数
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	ID           string          `json:"id"`
	Scope        BroadcastScope  `json:"scope"`
	Message      *Message        `json:"message"`
	Recipients   []string        `json:"recipients"`   // 实际接收者
	SentCount    int             `json:"sent_count"`   // 发送数量
	AckCount     int             `json:"ack_count"`    // 确认数量
	FailCount    int             `json:"fail_count"`   // 失败数量
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	Duration     time.Duration   `json:"duration"`
	Status       string          `json:"status"` // pending, in_progress, completed, failed
}

// BroadcastResult 广播结果
type BroadcastResult struct {
	ID            string          `json:"id"`
	BroadcastID   string          `json:"broadcast_id"`
	TotalSent     int             `json:"total_sent"`
	TotalAck      int             `json:"total_ack"`
	TotalFail     int             `json:"total_fail"`
	Results       []DeliveryResult `json:"results"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
	Duration      time.Duration   `json:"duration"`
}

// DeliveryResult 投递结果
type DeliveryResult struct {
	AgentID    string        `json:"agent_id"`
	Status     string        `json:"status"` // sent, acked, failed, timeout
	Error      string        `json:"error,omitempty"`
	Latency    time.Duration `json:"latency"`
	Timestamp  time.Time     `json:"timestamp"`
}

// MulticastGroup 组播组
type MulticastGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Members     []string  `json:"members"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MaxMembers  int       `json:"max_members"`
	IsPublic    bool      `json:"is_public"`
	Tags        []string  `json:"tags"`
}

// Subscription 订阅
type Subscription struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	AgentID     string    `json:"agent_id"`
	SubscribedAt time.Time `json:"subscribed_at"`
	Active      bool      `json:"active"`
	Filters     []string  `json:"filters"` // 消息过滤条件
}

// BroadcastConfig 广播配置
type BroadcastConfig struct {
	MaxRecipients   int           `json:"max_recipients"`   // 最大接收者
	Timeout         time.Duration `json:"timeout"`          // 超时时间
	RetryCount      int           `json:"retry_count"`      // 重试次数
	BatchSize       int           `json:"batch_size"`       // 批量大小
	BatchInterval   time.Duration `json:"batch_interval"`   // 批量间隔
	EnableAck       bool          `json:"enable_ack"`       // 启用确认
	MaxPending      int           `json:"max_pending"`      // 最大待发送数
}

// BroadcastManager 广播管理器
type BroadcastManager struct {
	config       BroadcastConfig
	p2pManager   *P2PManager
	router       *MessageRouter
	broadcasts   map[string]*BroadcastMessage
	groups       map[string]*MulticastGroup
	subscriptions map[string][]*Subscription // GroupID -> Subscriptions
	agentGroups  map[string][]string         // AgentID -> GroupIDs
	results      []*BroadcastResult
	pendingQueue chan *BroadcastMessage
	running      bool
	mu           sync.RWMutex
	wg           sync.WaitGroup
}

// NewBroadcastManager 创建广播管理器
func NewBroadcastManager(config BroadcastConfig, p2pManager *P2PManager, router *MessageRouter) *BroadcastManager {
	return &BroadcastManager{
		config:       config,
		p2pManager:   p2pManager,
		router:       router,
		broadcasts:   make(map[string]*BroadcastMessage),
		groups:       make(map[string]*MulticastGroup),
		subscriptions: make(map[string][]*Subscription),
		agentGroups:  make(map[string][]string),
		results:      make([]*BroadcastResult, 0),
		pendingQueue: make(chan *BroadcastMessage, config.MaxPending),
	}
}

// Initialize 初始化
func (bm *BroadcastManager) Initialize() error {
	// 加载默认组
	bm.loadDefaultGroups()
	return nil
}

// loadDefaultGroups 加载默认组
func (bm *BroadcastManager) loadDefaultGroups() {
	groups := []MulticastGroup{
		{
			ID:          "all-agents",
			Name:        "所有Agent",
			Description: "包含所有注册的Agent",
			IsPublic:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "online-agents",
			Name:        "在线Agent",
			Description: "仅包含当前在线的Agent",
			IsPublic:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "full-agents",
			Name:        "Full Agent",
			Description: "PC/服务器类型Agent",
			IsPublic:    true,
			Tags:        []string{"type:full"},
			CreatedAt:   time.Now(),
		},
		{
			ID:          "mobile-agents",
			Name:        "Mobile Agent",
			Description: "手机/平板类型Agent",
			IsPublic:    true,
			Tags:        []string{"type:mobile"},
			CreatedAt:   time.Now(),
		},
		{
			ID:          "edge-agents",
			Name:        "Edge Agent",
			Description: "边缘网关类型Agent",
			IsPublic:    true,
			Tags:        []string{"type:edge"},
			CreatedAt:   time.Now(),
		},
	}

	bm.mu.Lock()
	for _, g := range groups {
		bm.groups[g.ID] = &g
	}
	bm.mu.Unlock()
}

// Start 启动广播管理器
func (bm *BroadcastManager) Start(ctx context.Context) error {
	bm.mu.Lock()
	bm.running = true
	bm.mu.Unlock()

	bm.wg.Add(1)
	go bm.processQueue(ctx)

	return nil
}

// Stop 停止广播管理器
func (bm *BroadcastManager) Stop() {
	bm.mu.Lock()
	bm.running = false
	bm.mu.Unlock()

	bm.wg.Wait()
	close(bm.pendingQueue)
}

// Broadcast 广播消息
func (bm *BroadcastManager) Broadcast(ctx context.Context, msg *Message, scope BroadcastScope) (*BroadcastMessage, error) {
	broadcast := &BroadcastMessage{
		ID:        fmt.Sprintf("bc-%d", time.Now().UnixNano()),
		Scope:     scope,
		Message:   msg,
		StartTime: time.Now(),
		Status:    "pending",
	}

	// 确定接收者
	recipients, err := bm.resolveRecipients(scope)
	if err != nil {
		return nil, err
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("没有接收者")
	}

	// 检查最大接收者限制
	if scope.MaxRecipients > 0 && len(recipients) > scope.MaxRecipients {
		recipients = recipients[:scope.MaxRecipients]
	}

	broadcast.Recipients = recipients
	broadcast.SentCount = len(recipients)

	// 入队处理
	select {
	case bm.pendingQueue <- broadcast:
	default:
		return nil, fmt.Errorf("广播队列已满")
	}

	return broadcast, nil
}

// resolveRecipients 解析接收者
func (bm *BroadcastManager) resolveRecipients(scope BroadcastScope) ([]string, error) {
	var recipients []string

	switch scope.Mode {
	case BroadcastAll:
		// 获取所有连接的Agent
		recipients = bm.p2pManager.GetConnectedPeers()

	case BroadcastOnline:
		// 仅在线Agent
		recipients = bm.p2pManager.GetConnectedPeers()

	case BroadcastType:
		// 按类型
		recipients = bm.router.FindAgentsByCapability(scope.Filter)

	case BroadcastCapability:
		// 按能力
		recipients = bm.router.FindAgentsByCapability(scope.Filter)

	case BroadcastExclude:
		// 排除某些Agent
		all := bm.p2pManager.GetConnectedPeers()
		excludeMap := make(map[string]bool)
		for _, id := range scope.ExcludeList {
			excludeMap[id] = true
		}
		for _, id := range all {
			if !excludeMap[id] {
				recipients = append(recipients, id)
			}
		}

	case BroadcastRegion:
		// 按区域 (简化实现)
		recipients = bm.p2pManager.GetConnectedPeers()

	default:
		return nil, fmt.Errorf("未知广播模式: %s", scope.Mode)
	}

	// 应用排除列表
	if len(scope.ExcludeList) > 0 {
		excludeMap := make(map[string]bool)
		for _, id := range scope.ExcludeList {
			excludeMap[id] = true
		}
		filtered := make([]string, 0)
		for _, id := range recipients {
			if !excludeMap[id] {
				filtered = append(filtered, id)
			}
		}
		recipients = filtered
	}

	// 应用包含列表
	if len(scope.IncludeList) > 0 {
		recipients = scope.IncludeList
	}

	return recipients, nil
}

// processQueue 处理广播队列
func (bm *BroadcastManager) processQueue(ctx context.Context) {
	defer bm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case broadcast, ok := <-bm.pendingQueue:
			if !ok {
				return
			}
			bm.executeBroadcast(ctx, broadcast)
		}
	}
}

// executeBroadcast 执行广播
func (bm *BroadcastManager) executeBroadcast(ctx context.Context, broadcast *BroadcastMessage) {
	broadcast.Status = "in_progress"

	// 保存广播记录
	bm.mu.Lock()
	bm.broadcasts[broadcast.ID] = broadcast
	bm.mu.Unlock()

	// 准备结果
	result := &BroadcastResult{
		ID:          fmt.Sprintf("result-%d", time.Now().UnixNano()),
		BroadcastID: broadcast.ID,
		StartTime:   time.Now(),
		Results:     make([]DeliveryResult, 0),
	}

	// 批量发送
	batchSize := bm.config.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	for i := 0; i < len(broadcast.Recipients); i += batchSize {
		end := i + batchSize
		if end > len(broadcast.Recipients) {
			end = len(broadcast.Recipients)
		}

		batch := broadcast.Recipients[i:end]

		// 并行发送批次
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, agentID := range batch {
			wg.Add(1)
			go func(aid string) {
				defer wg.Done()

				start := time.Now()
				err := bm.p2pManager.Send(aid, broadcast.Message)
				latency := time.Since(start)

				deliveryResult := DeliveryResult{
					AgentID:   aid,
					Timestamp: time.Now(),
					Latency:   latency,
				}

				if err != nil {
					deliveryResult.Status = "failed"
					deliveryResult.Error = err.Error()
					mu.Lock()
					result.TotalFail++
					broadcast.FailCount++
					mu.Unlock()
				} else {
					deliveryResult.Status = "sent"
					mu.Lock()
					result.TotalSent++
					broadcast.SentCount++
					mu.Unlock()
				}

				mu.Lock()
				result.Results = append(result.Results, deliveryResult)
				mu.Unlock()
			}(agentID)
		}

		wg.Wait()

		// 批次间等待
		if i+batchSize < len(broadcast.Recipients) {
			time.Sleep(bm.config.BatchInterval)
		}
	}

	broadcast.EndTime = time.Now()
	broadcast.Duration = broadcast.EndTime.Sub(broadcast.StartTime)
	broadcast.Status = "completed"

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// 保存结果
	bm.mu.Lock()
	bm.results = append(bm.results, result)
	bm.mu.Unlock()
}

// Multicast 组播消息
func (bm *BroadcastManager) Multicast(ctx context.Context, groupID string, msg *Message) (*BroadcastResult, error) {
	bm.mu.RLock()
	group, ok := bm.groups[groupID]
	bm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("组不存在: %s", groupID)
	}

	scope := BroadcastScope{
		Mode:        BroadcastAll,
		IncludeList: group.Members,
	}

	broadcast, err := bm.Broadcast(ctx, msg, scope)
	if err != nil {
		return nil, err
	}

	// 等待完成或超时
	select {
	case <-time.After(bm.config.Timeout):
		return nil, fmt.Errorf("组播超时")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 返回结果
	bm.mu.RLock()
	var result *BroadcastResult
	for _, r := range bm.results {
		if r.BroadcastID == broadcast.ID {
			result = r
			break
		}
	}
	bm.mu.RUnlock()

	return result, nil
}

// CreateGroup 创建组播组
func (bm *BroadcastManager) CreateGroup(group *MulticastGroup) error {
	if group.ID == "" {
		group.ID = fmt.Sprintf("group-%d", time.Now().UnixNano())
	}
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	bm.mu.Lock()
	bm.groups[group.ID] = group
	bm.mu.Unlock()

	return nil
}

// DeleteGroup 删除组播组
func (bm *BroadcastManager) DeleteGroup(groupID string) error {
	bm.mu.Lock()
	delete(bm.groups, groupID)
	delete(bm.subscriptions, groupID)
	// 清理agentGroups
	for agentID, groups := range bm.agentGroups {
		newGroups := make([]string, 0)
		for _, g := range groups {
			if g != groupID {
				newGroups = append(newGroups, g)
			}
		}
		bm.agentGroups[agentID] = newGroups
	}
	bm.mu.Unlock()

	return nil
}

// AddMember 添加组成员
func (bm *BroadcastManager) AddMember(groupID string, agentID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	group, ok := bm.groups[groupID]
	if !ok {
		return fmt.Errorf("组不存在: %s", groupID)
	}

	if group.MaxMembers > 0 && len(group.Members) >= group.MaxMembers {
		return fmt.Errorf("组已满")
	}

	// 检查是否已在组中
	for _, m := range group.Members {
		if m == agentID {
			return nil // 已存在
		}
	}

	group.Members = append(group.Members, agentID)
	group.UpdatedAt = time.Now()

	// 更新agentGroups
	bm.agentGroups[agentID] = append(bm.agentGroups[agentID], groupID)

	return nil
}

// RemoveMember 移除组成员
func (bm *BroadcastManager) RemoveMember(groupID string, agentID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	group, ok := bm.groups[groupID]
	if !ok {
		return fmt.Errorf("组不存在: %s", groupID)
	}

	newMembers := make([]string, 0)
	for _, m := range group.Members {
		if m != agentID {
			newMembers = append(newMembers, m)
		}
	}
	group.Members = newMembers
	group.UpdatedAt = time.Now()

	// 更新agentGroups
	newGroups := make([]string, 0)
	for _, g := range bm.agentGroups[agentID] {
		if g != groupID {
			newGroups = append(newGroups, g)
		}
	}
	bm.agentGroups[agentID] = newGroups

	return nil
}

// Subscribe 订阅组
func (bm *BroadcastManager) Subscribe(groupID string, agentID string, filters []string) error {
	sub := &Subscription{
		ID:           fmt.Sprintf("sub-%d", time.Now().UnixNano()),
		GroupID:      groupID,
		AgentID:      agentID,
		SubscribedAt: time.Now(),
		Active:       true,
		Filters:      filters,
	}

	bm.mu.Lock()
	bm.subscriptions[groupID] = append(bm.subscriptions[groupID], sub)
	bm.mu.Unlock()

	return nil
}

// Unsubscribe 取消订阅
func (bm *BroadcastManager) Unsubscribe(groupID string, agentID string) error {
	bm.mu.Lock()
	subs := bm.subscriptions[groupID]
	newSubs := make([]*Subscription, 0)
	for _, s := range subs {
		if s.AgentID != agentID {
			newSubs = append(newSubs, s)
		}
	}
	bm.subscriptions[groupID] = newSubs
	bm.mu.Unlock()

	return nil
}

// GetGroup 获取组
func (bm *BroadcastManager) GetGroup(groupID string) (*MulticastGroup, error) {
	bm.mu.RLock()
	group, ok := bm.groups[groupID]
	bm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("组不存在: %s", groupID)
	}
	return group, nil
}

// ListGroups 列出所有组
func (bm *BroadcastManager) ListGroups() []*MulticastGroup {
	bm.mu.RLock()
	groups := make([]*MulticastGroup, 0, len(bm.groups))
	for _, g := range bm.groups {
		groups = append(groups, g)
	}
	bm.mu.RUnlock()
	return groups
}

// GetAgentGroups 获取Agent所属的组
func (bm *BroadcastManager) GetAgentGroups(agentID string) []string {
	bm.mu.RLock()
	groups := bm.agentGroups[agentID]
	if groups == nil {
		groups = []string{}
	}
	result := make([]string, len(groups))
	copy(result, groups)
	bm.mu.RUnlock()
	return result
}

// GetBroadcast 获取广播记录
func (bm *BroadcastManager) GetBroadcast(broadcastID string) (*BroadcastMessage, error) {
	bm.mu.RLock()
	broadcast, ok := bm.broadcasts[broadcastID]
	bm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("广播不存在: %s", broadcastID)
	}
	return broadcast, nil
}

// GetResults 获取广播结果
func (bm *BroadcastManager) GetResults(limit int) []*BroadcastResult {
	bm.mu.RLock()
	size := len(bm.results)
	if limit > size || limit <= 0 {
		limit = size
	}
	results := bm.results[size-limit:]
	bm.mu.RUnlock()
	return results
}

// GetStatistics 获取统计信息
func (bm *BroadcastManager) GetStatistics() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	totalSent := 0
	totalAck := 0
	totalFail := 0

	for _, b := range bm.broadcasts {
		totalSent += b.SentCount
		totalAck += b.AckCount
		totalFail += b.FailCount
	}

	return map[string]interface{}{
		"total_broadcasts": len(bm.broadcasts),
		"total_groups":     len(bm.groups),
		"total_sent":       totalSent,
		"total_ack":        totalAck,
		"total_fail":       totalFail,
		"running":          bm.running,
	}
}

// ExportGroups 导出组配置
func (bm *BroadcastManager) ExportGroups() ([]byte, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return json.MarshalIndent(bm.groups, "", "  ")
}

// ImportGroups 导入组配置
func (bm *BroadcastManager) ImportGroups(data []byte) error {
	groups := make(map[string]*MulticastGroup)
	if err := json.Unmarshal(data, &groups); err != nil {
		return err
	}

	bm.mu.Lock()
	for k, v := range groups {
		bm.groups[k] = v
	}
	bm.mu.Unlock()

	return nil
}