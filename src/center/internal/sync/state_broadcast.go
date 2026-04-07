package sync

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// === 状态广播服务 (v3.1.0) ===
//
// Center 是永远在线的灵魂载体，状态广播确保：
// - 设备状态变更实时通知其他设备
// - 支持订阅指定类型的状态变更
// - 支持离线设备的状态缓存

// StateBroadcastConfig - 状态广播配置
type StateBroadcastConfig struct {
	// 广播间隔（批量广播）
	BroadcastInterval time.Duration
	// 状态缓存大小
	CacheSize int
	// 缓存过期时间
	CacheExpiration time.Duration
	// 是否广播电池状态
	BroadcastBattery bool
	// 是否广播网络状态
	BroadcastNetwork bool
	// 是否广播场景状态
	BroadcastScene bool
	// 是否广播位置状态
	BroadcastLocation bool
}

// DefaultStateBroadcastConfig 默认配置
func DefaultStateBroadcastConfig() StateBroadcastConfig {
	return StateBroadcastConfig{
		BroadcastInterval:  5 * time.Second,
		CacheSize:          1000,
		CacheExpiration:    10 * time.Minute,
		BroadcastBattery:   true,
		BroadcastNetwork:   true,
		BroadcastScene:     true,
		BroadcastLocation:  false, // 位置敏感，默认不广播
	}
}

// StateBroadcastService - 状态广播服务
type StateBroadcastService struct {
	mu sync.RWMutex

	// 配置
	config StateBroadcastConfig

	// 状态同步管理器
	stateManager *StateSyncManager

	// 消息总线
	messageBus *MessageBus

	// 设备管理器
	deviceManager *DeviceManager

	// 订阅管理
	subscriptions map[string]*StateSubscription // subscriptionId -> subscription

	// 身份订阅: identityId -> []subscriptionId
	identitySubscriptions map[string][]string

	// 类型订阅: changeType -> []subscriptionId
	typeSubscriptions map[StateChangeType][]string

	// 待广播队列
	pendingBroadcasts []*StateChange

	// 广播缓存（离线设备）
	broadcastCache map[string][]CachedBroadcast // agentId -> cached broadcasts

	// 停止信号
	stopChan chan struct{}
}

// StateSubscription - 状态订阅
type StateSubscription struct {
	SubscriptionID string           `json:"subscription_id"`
	SubscriberID   string           `json:"subscriber_id"`   // 订阅者设备ID
	IdentityID     string           `json:"identity_id"`     // 所属身份
	ChangeTypes    []StateChangeType `json:"change_types"`   // 关注的变更类型
	TargetDevices  []string         `json:"target_devices"`  // 关注的设备（可选，空表示全部）
	Active         bool             `json:"active"`
	CreatedAt      time.Time        `json:"created_at"`
}

// CachedBroadcast - 缓存的广播消息
type CachedBroadcast struct {
	Change      *StateChange `json:"change"`
	BroadcastAt time.Time    `json:"broadcast_at"`
	Delivered   bool         `json:"delivered"`
}

// NewStateBroadcastService 创建状态广播服务
func NewStateBroadcastService(config StateBroadcastConfig) *StateBroadcastService {
	if config.BroadcastInterval == 0 {
		config = DefaultStateBroadcastConfig()
	}

	return &StateBroadcastService{
		config:                config,
		subscriptions:         make(map[string]*StateSubscription),
		identitySubscriptions: make(map[string][]string),
		typeSubscriptions:     make(map[StateChangeType][]string),
		pendingBroadcasts:     make([]*StateChange, 0),
		broadcastCache:        make(map[string][]CachedBroadcast),
		stopChan:              make(chan struct{}),
	}
}

// SetStateManager 设置状态同步管理器
func (s *StateBroadcastService) SetStateManager(sm *StateSyncManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateManager = sm
}

// SetMessageBus 设置消息总线
func (s *StateBroadcastService) SetMessageBus(mb *MessageBus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messageBus = mb
}

// SetDeviceManager 设置设备管理器
func (s *StateBroadcastService) SetDeviceManager(dm *DeviceManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceManager = dm
}

// Start 启动广播服务
func (s *StateBroadcastService) Start() {
	go s.broadcastLoop()
	go s.cacheCleanupLoop()
	log.Println("State broadcast service started")
}

// Stop 停止广播服务
func (s *StateBroadcastService) Stop() {
	close(s.stopChan)
	log.Println("State broadcast service stopped")
}

// === 订阅管理 ===

// Subscribe 订阅状态变更
func (s *StateBroadcastService) Subscribe(subscriberID, identityID string, changeTypes []StateChangeType, targetDevices []string) *StateSubscription {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriptionID := generateSubscriptionID(subscriberID)

	sub := &StateSubscription{
		SubscriptionID: subscriptionID,
		SubscriberID:   subscriberID,
		IdentityID:     identityID,
		ChangeTypes:    changeTypes,
		TargetDevices:  targetDevices,
		Active:         true,
		CreatedAt:      time.Now(),
	}

	s.subscriptions[subscriptionID] = sub
	s.identitySubscriptions[identityID] = append(s.identitySubscriptions[identityID], subscriptionID)

	// 按类型注册
	for _, ct := range changeTypes {
		s.typeSubscriptions[ct] = append(s.typeSubscriptions[ct], subscriptionID)
	}

	log.Printf("State subscription created: %s for %s, types=%v", subscriptionID, subscriberID, changeTypes)
	return sub
}

// Unsubscribe 取消订阅
func (s *StateBroadcastService) Unsubscribe(subscriptionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, exists := s.subscriptions[subscriptionID]
	if !exists {
		return
	}

	// 从身份列表移除
	var newIdentitySubs []string
	for _, sid := range s.identitySubscriptions[sub.IdentityID] {
		if sid != subscriptionID {
			newIdentitySubs = append(newIdentitySubs, sid)
		}
	}
	s.identitySubscriptions[sub.IdentityID] = newIdentitySubs

	// 从类型列表移除
	for _, ct := range sub.ChangeTypes {
		var newTypeSubs []string
		for _, sid := range s.typeSubscriptions[ct] {
			if sid != subscriptionID {
				newTypeSubs = append(newTypeSubs, sid)
			}
		}
		s.typeSubscriptions[ct] = newTypeSubs
	}

	// 删除订阅
	delete(s.subscriptions, subscriptionID)

	log.Printf("State subscription removed: %s", subscriptionID)
}

// GetSubscription 获取订阅
func (s *StateBroadcastService) GetSubscription(subscriptionID string) *StateSubscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subscriptions[subscriptionID]
}

// GetIdentitySubscriptions 获取身份下的订阅
func (s *StateBroadcastService) GetIdentitySubscriptions(identityID string) []*StateSubscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var subs []*StateSubscription
	for _, sid := range s.identitySubscriptions[identityID] {
		if sub := s.subscriptions[sid]; sub != nil && sub.Active {
			subs = append(subs, sub)
		}
	}
	return subs
}

// === 状态广播 ===

// BroadcastChange 广播状态变更
func (s *StateBroadcastService) BroadcastChange(change *StateChange) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否应该广播此变更类型
	if !s.shouldBroadcast(change.ChangeType) {
		return
	}

	// 加入待广播队列
	s.pendingBroadcasts = append(s.pendingBroadcasts, change)
}

// BroadcastImmediate 立即广播（紧急状态）
func (s *StateBroadcastService) BroadcastImmediate(change *StateChange) {
	s.mu.Lock()
	s.mu.Unlock()

	if !s.shouldBroadcast(change.ChangeType) {
		return
	}

	// 立即广播到订阅者
	s.broadcastToSubscribers(change)
}

// shouldBroadcast 检查是否应该广播
func (s *StateBroadcastService) shouldBroadcast(changeType StateChangeType) bool {
	switch changeType {
	case StateChangeOnline, StateChangeOffline:
		return true // 状态变更总是广播
	case StateChangeBattery:
		return s.config.BroadcastBattery
	case StateChangeNetwork:
		return s.config.BroadcastNetwork
	case StateChangeScene:
		return s.config.BroadcastScene
	case StateChangeLocation:
		return s.config.BroadcastLocation
	case StateChangeFull:
		return true
	default:
		return true
	}
}

// broadcastLoop 广播循环
func (s *StateBroadcastService) broadcastLoop() {
	ticker := time.NewTicker(s.config.BroadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.flushBroadcasts()
		}
	}
}

// flushBroadcasts 执行批量广播
func (s *StateBroadcastService) flushBroadcasts() {
	s.mu.Lock()
	broadcasts := s.pendingBroadcasts
	s.pendingBroadcasts = make([]*StateChange, 0)
	s.mu.Unlock()

	if len(broadcasts) == 0 {
		return
	}

	// 按身份分组广播
	for _, change := range broadcasts {
		s.broadcastToSubscribers(change)
	}
}

// broadcastToSubscribers 广播到订阅者
func (s *StateBroadcastService) broadcastToSubscribers(change *StateChange) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.messageBus == nil {
		return
	}

	// 获取相关订阅
	var targetSubs []string

	// 按身份获取订阅
	if subs := s.identitySubscriptions[change.IdentityID]; len(subs) > 0 {
		targetSubs = append(targetSubs, subs...)
	}

	// 按类型获取订阅
	if subs := s.typeSubscriptions[change.ChangeType]; len(subs) > 0 {
		for _, sid := range subs {
			// 检查是否已包含
			found := false
			for _, ts := range targetSubs {
				if ts == sid {
					found = true
					break
				}
			}
			if !found {
				targetSubs = append(targetSubs, sid)
			}
		}
	}

	// 广播到每个订阅者
	for _, sid := range targetSubs {
		sub := s.subscriptions[sid]
		if sub == nil || !sub.Active {
			continue
		}

		// 检查订阅者是否在线
		if s.deviceManager != nil && !s.deviceManager.IsDeviceOnline(sub.SubscriberID) {
			// 缓存到离线队列
			s.cacheForOffline(sub.SubscriberID, change)
			continue
		}

		// 检查是否关注此设备
		if len(sub.TargetDevices) > 0 {
			targeted := false
			for _, target := range sub.TargetDevices {
				if target == change.AgentID {
					targeted = true
					break
				}
			}
			if !targeted {
				continue
			}
		}

		// 发送消息
		msg := &Message{
			ID:         generateMessageID(),
			FromAgent:  "center",
			ToAgent:    sub.SubscriberID,
			IdentityID: change.IdentityID,
			Type:       MessageTypeNotification,
			Priority:   s.getChangePriority(change.ChangeType),
			Payload:    change.ToMap(),
			CreatedAt:  time.Now(),
		}

		s.messageBus.Send(msg)
	}
}

// getChangePriority 获取变更优先级
func (s *StateBroadcastService) getChangePriority(changeType StateChangeType) int {
	switch changeType {
	case StateChangeOnline, StateChangeOffline:
		return MessagePriorityHigh
	case StateChangeBattery:
		return MessagePriorityNormal
	case StateChangeNetwork:
		return MessagePriorityNormal
	case StateChangeScene:
		return MessagePriorityNormal
	case StateChangeLocation:
		return MessagePriorityLow
	default:
		return MessagePriorityNormal
	}
}

// === 离线缓存 ===

// cacheForOffline 缓存离线设备消息
func (s *StateBroadcastService) cacheForOffline(agentID string, change *StateChange) {
	cache := s.broadcastCache[agentID]

	cached := CachedBroadcast{
		Change:      change,
		BroadcastAt: time.Now(),
		Delivered:   false,
	}

	cache = append(cache, cached)

	// 限制缓存大小
	if len(cache) > s.config.CacheSize {
		cache = cache[len(cache)-s.config.CacheSize:]
	}

	s.broadcastCache[agentID] = cache
}

// DeliverCachedBroadcasts 投递缓存的广播
func (s *StateBroadcastService) DeliverCachedBroadcasts(agentID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cache := s.broadcastCache[agentID]
	if len(cache) == 0 {
		return 0
	}

	if s.messageBus == nil {
		return 0
	}

	// 检查设备是否在线
	if s.deviceManager != nil && !s.deviceManager.IsDeviceOnline(agentID) {
		return 0
	}

	count := 0
	for _, cached := range cache {
		if !cached.Delivered && time.Since(cached.BroadcastAt) < s.config.CacheExpiration {
			msg := &Message{
				ID:         generateMessageID(),
				FromAgent:  "center",
				ToAgent:    agentID,
				IdentityID: cached.Change.IdentityID,
				Type:       MessageTypeNotification,
				Priority:   MessagePriorityNormal,
				Payload:    cached.Change.ToMap(),
				CreatedAt:  time.Now(),
				Metadata:   map[string]string{"cached": "true", "original_time": cached.BroadcastAt.Format(time.RFC3339)},
			}

			s.messageBus.Send(msg)
			cached.Delivered = true
			count++
		}
	}

	// 清理已投递或过期的缓存
	var newCache []CachedBroadcast
	for _, cached := range cache {
		if !cached.Delivered && time.Since(cached.BroadcastAt) < s.config.CacheExpiration {
			newCache = append(newCache, cached)
		}
	}
	s.broadcastCache[agentID] = newCache

	log.Printf("Delivered %d cached broadcasts to %s", count, agentID)
	return count
}

// cacheCleanupLoop 缓存清理循环
func (s *StateBroadcastService) cacheCleanupLoop() {
	ticker := time.NewTicker(s.config.CacheExpiration / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cleanExpiredCache()
		}
	}
}

// cleanExpiredCache 清理过期缓存
func (s *StateBroadcastService) cleanExpiredCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for agentID, cache := range s.broadcastCache {
		var newCache []CachedBroadcast
		for _, cached := range cache {
			if time.Since(cached.BroadcastAt) < s.config.CacheExpiration {
				newCache = append(newCache, cached)
			}
		}
		if len(newCache) == 0 {
			delete(s.broadcastCache, agentID)
		} else {
			s.broadcastCache[agentID] = newCache
		}
	}
}

// GetCachedCount 获取缓存消息数量
func (s *StateBroadcastService) GetCachedCount(agentID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.broadcastCache[agentID])
}

// === 全局状态广播 ===

// BroadcastAllStates 广播所有设备状态
func (s *StateBroadcastService) BroadcastAllStates(identityID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.stateManager == nil || s.messageBus == nil {
		return
	}

	// 获取所有设备状态
	states := s.stateManager.GetIdentityDeviceStates(identityID)

	// 广播到在线订阅者
	for _, sid := range s.identitySubscriptions[identityID] {
		sub := s.subscriptions[sid]
		if sub == nil || !sub.Active {
			continue
		}

		if s.deviceManager != nil && !s.deviceManager.IsDeviceOnline(sub.SubscriberID) {
			continue
		}

		// 发送完整状态列表
		payload := map[string]interface{}{
			"type":       "full_state_sync",
			"identity_id": identityID,
			"states":     make([]map[string]interface{}, 0),
		}

		for _, state := range states {
			payload["states"] = append(payload["states"].([]map[string]interface{}), state.ToMap())
		}

		msg := &Message{
			ID:         generateMessageID(),
			FromAgent:  "center",
			ToAgent:    sub.SubscriberID,
			IdentityID: identityID,
			Type:       MessageTypeSync,
			Priority:   MessagePriorityHigh,
			Payload:    payload,
			CreatedAt:  time.Now(),
		}

		s.messageBus.Send(msg)
	}
}

// === 辅助函数 ===

func generateSubscriptionID(subscriberID string) string {
	return "sub_" + subscriberID + "_" + time.Now().Format("20060102150405")
}

// === 统计 ===

// BroadcastStats 广播统计
type BroadcastStats struct {
	TotalSubscriptions   int            `json:"total_subscriptions"`
	ActiveSubscriptions  int            `json:"active_subscriptions"`
	PendingBroadcasts    int            `json:"pending_broadcasts"`
	CachedBroadcasts     int            `json:"cached_broadcasts"`
	ByChangeType         map[string]int `json:"by_change_type"`
	OfflineDevicesCache  int            `json:"offline_devices_cache"`
}

// GetStats 获取广播统计
func (s *StateBroadcastService) GetStats() *BroadcastStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &BroadcastStats{
		ByChangeType: make(map[string]int),
	}

	stats.TotalSubscriptions = len(s.subscriptions)
	for _, sub := range s.subscriptions {
		if sub.Active {
			stats.ActiveSubscriptions++
		}
	}

	stats.PendingBroadcasts = len(s.pendingBroadcasts)

	for agentID, cache := range s.broadcastCache {
		stats.CachedBroadcasts += len(cache)
		if len(cache) > 0 {
			stats.OfflineDevicesCache++
		}
	}

	for ct, subs := range s.typeSubscriptions {
		stats.ByChangeType[string(ct)] = len(subs)
	}

	return stats
}

// ToJSON 序列化订阅信息
func (s *StateSubscription) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}