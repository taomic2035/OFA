// Package iot - 离线支持
package iot

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineMode 离线模式
type OfflineMode int

const (
	OfflineDisabled OfflineMode = iota // 禁用离线
	OfflineCache    OfflineMode = iota // 缓存模式
	OfflineEdge     OfflineMode = iota // 边缘计算模式
)

// OfflineConfig 离线配置
type OfflineConfig struct {
	Mode           OfflineMode `json:"mode"`
	CacheSize      int         `json:"cache_size"`      // 最大缓存数
	CacheTTL       int         `json:"cache_ttl"`       // 缓存过期(秒)
	PersistPath    string      `json:"persist_path"`    // 持久化路径
	AutoReplay     bool        `json:"auto_replay"`     // 自动重放
	EdgeComputing  bool        `json:"edge_computing"`  // 边缘计算
}

// DefaultOfflineConfig 默认配置
func DefaultOfflineConfig() OfflineConfig {
	return OfflineConfig{
		Mode:          OfflineCache,
		CacheSize:     100,
		CacheTTL:      3600,
		AutoReplay:    true,
		EdgeComputing: false,
	}
}

// CachedMessage 缓存的消息
type CachedMessage struct {
	ID        string      `json:"id"`
	Topic     string      `json:"topic"`
	Payload   []byte      `json:"payload"`
	QoS       byte        `json:"qos"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	Sent      bool        `json:"sent"`
}

// OfflineCache 离线缓存
type OfflineCache struct {
	config    OfflineConfig
	messages  map[string]*CachedMessage
	msgOrder  []string
	dataPath  string

	mu sync.Mutex
}

// NewOfflineCache 创建离线缓存
func NewOfflineCache(config OfflineConfig) *OfflineCache {
	oc := &OfflineCache{
		config:   config,
		messages: make(map[string]*CachedMessage),
		msgOrder: make([]string, 0, config.CacheSize),
	}

	if config.PersistPath != "" {
		oc.dataPath = config.PersistPath
		os.MkdirAll(config.PersistPath, 0755)
		oc.load()
	}

	return oc
}

// Cache 缓存消息
func (oc *OfflineCache) Cache(topic string, payload []byte, qos byte) string {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	// 检查容量
	if len(oc.messages) >= oc.config.CacheSize {
		oc.evictOldest()
	}

	id := generateMsgID()
	msg := &CachedMessage{
		ID:        id,
		Topic:     topic,
		Payload:   payload,
		QoS:       qos,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(oc.config.CacheTTL) * time.Second),
		Sent:      false,
	}

	oc.messages[id] = msg
	oc.msgOrder = append(oc.msgOrder, id)

	return id
}

// Get 获取缓存消息
func (oc *OfflineCache) Get(id string) *CachedMessage {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	msg, ok := oc.messages[id]
	if !ok {
		return nil
	}

	if time.Now().After(msg.ExpiresAt) {
		delete(oc.messages, id)
		return nil
	}

	return msg
}

// GetPending 获取待发送消息
func (oc *OfflineCache) GetPending() []*CachedMessage {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	var pending []*CachedMessage
	now := time.Now()

	for _, id := range oc.msgOrder {
		if msg, ok := oc.messages[id]; ok && !msg.Sent && now.Before(msg.ExpiresAt) {
			pending = append(pending, msg)
		}
	}

	return pending
}

// MarkSent 标记已发送
func (oc *OfflineCache) MarkSent(id string) {
	oc.mu.Lock()
	if msg, ok := oc.messages[id]; ok {
		msg.Sent = true
	}
	oc.mu.Unlock()
}

// Clear 清除缓存
func (oc *OfflineCache) Clear() {
	oc.mu.Lock()
	oc.messages = make(map[string]*CachedMessage)
	oc.msgOrder = make([]string, 0, oc.config.CacheSize)
	oc.mu.Unlock()
}

// PurgeExpired 清除过期消息
func (oc *OfflineCache) PurgeExpired() int {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	now := time.Now()
	count := 0

	for id, msg := range oc.messages {
		if now.After(msg.ExpiresAt) {
			delete(oc.messages, id)
			count++
		}
	}

	// 重建顺序
	oc.msgOrder = make([]string, 0, len(oc.messages))
	for id := range oc.messages {
		oc.msgOrder = append(oc.msgOrder, id)
	}

	return count
}

// Persist 持久化
func (oc *OfflineCache) Persist() {
	if oc.dataPath == "" {
		return
	}

	oc.mu.Lock()
	defer oc.mu.Unlock()

	// 过滤未发送且未过期
	var pending []*CachedMessage
	now := time.Now()
	for _, msg := range oc.messages {
		if !msg.Sent && now.Before(msg.ExpiresAt) {
			pending = append(pending, msg)
		}
	}

	data, _ := json.Marshal(pending)
	path := filepath.Join(oc.dataPath, "offline_cache.json")
	os.WriteFile(path, data, 0644)
}

// Load 加载
func (oc *OfflineCache) load() {
	if oc.dataPath == "" {
		return
	}

	path := filepath.Join(oc.dataPath, "offline_cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var pending []*CachedMessage
	if json.Unmarshal(data, &pending) == nil {
		now := time.Now()
		for _, msg := range pending {
			if now.Before(msg.ExpiresAt) {
				oc.messages[msg.ID] = msg
				oc.msgOrder = append(oc.msgOrder, msg.ID)
			}
		}
	}
}

// Stats 统计
func (oc *OfflineCache) Stats() map[string]int {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	pending := 0
	for _, msg := range oc.messages {
		if !msg.Sent {
			pending++
		}
	}

	return map[string]int{
		"total":   len(oc.messages),
		"pending": pending,
	}
}

func (oc *OfflineCache) evictOldest() {
	if len(oc.msgOrder) == 0 {
		return
	}

	// 移除最旧的已发送消息
	for i, id := range oc.msgOrder {
		if msg, ok := oc.messages[id]; ok && msg.Sent {
			delete(oc.messages, id)
			oc.msgOrder = append(oc.msgOrder[:i], oc.msgOrder[i+1:]...)
			return
		}
	}

	// 如果没有已发送的，移除最旧的
	if len(oc.msgOrder) > 0 {
		id := oc.msgOrder[0]
		delete(oc.messages, id)
		oc.msgOrder = oc.msgOrder[1:]
	}
}

// OfflineManager 离线管理器
type OfflineManager struct {
	config  OfflineConfig
	cache   *OfflineCache
	agent   *IoTAgent

	// 边缘计算
	edgeRules map[string]EdgeRule

	mu      sync.Mutex
	running bool
}

// EdgeRule 边缘计算规则
type EdgeRule struct {
	ID         string
	Condition  func(map[string]interface{}) bool
	Action     func(map[string]interface{}) error
	Enabled    bool
}

// NewOfflineManager 创建离线管理器
func NewOfflineManager(agent *IoTAgent, config OfflineConfig) *OfflineManager {
	return &OfflineManager{
		config:    config,
		cache:     NewOfflineCache(config),
		agent:     agent,
		edgeRules: make(map[string]EdgeRule),
	}
}

// Start 启动
func (om *OfflineManager) Start() {
	om.mu.Lock()
	om.running = true
	om.mu.Unlock()

	// 重放缓存消息
	if om.config.AutoReplay {
		go om.replayCachedMessages()
	}

	log.Printf("Offline manager started (mode: %d)", om.config.Mode)
}

// Stop 停止
func (om *OfflineManager) Stop() {
	om.mu.Lock()
	om.running = false
	om.mu.Unlock()

	om.cache.Persist()
}

// CacheMessage 缓存消息
func (om *OfflineManager) CacheMessage(topic string, payload []byte, qos byte) string {
	return om.cache.Cache(topic, payload, qos)
}

// GetPendingMessages 获取待发送消息
func (om *OfflineManager) GetPendingMessages() []*CachedMessage {
	return om.cache.GetPending()
}

// MarkSent 标记已发送
func (om *OfflineManager) MarkSent(id string) {
	om.cache.MarkSent(id)
}

// AddEdgeRule 添加边缘规则
func (om *OfflineManager) AddEdgeRule(rule EdgeRule) {
	om.mu.Lock()
	om.edgeRules[rule.ID] = rule
	om.mu.Unlock()
}

// RemoveEdgeRule 移除边缘规则
func (om *OfflineManager) RemoveEdgeRule(id string) {
	om.mu.Lock()
	delete(om.edgeRules, id)
	om.mu.Unlock()
}

// ProcessEdge 处理边缘计算
func (om *OfflineManager) ProcessEdge(data map[string]interface{}) error {
	if !om.config.EdgeComputing {
		return nil
	}

	om.mu.Lock()
	defer om.mu.Unlock()

	for _, rule := range om.edgeRules {
		if !rule.Enabled {
			continue
		}

		if rule.Condition != nil && rule.Condition(data) {
			if rule.Action != nil {
				if err := rule.Action(data); err != nil {
					log.Printf("Edge rule %s failed: %v", rule.ID, err)
				}
			}
		}
	}

	return nil
}

func (om *OfflineManager) replayCachedMessages() {
	if om.agent == nil || om.agent.mqttClient == nil {
		return
	}

	// 等待连接
	for i := 0; i < 30; i++ {
		if om.agent.mqttClient.IsConnected() {
			break
		}
		time.Sleep(time.Second)
	}

	if !om.agent.mqttClient.IsConnected() {
		return
	}

	// 重放
	pending := om.cache.GetPending()
	for _, msg := range pending {
		if err := om.agent.mqttClient.Publish(msg.Topic, msg.QoS, msg.Payload); err != nil {
			log.Printf("Failed to replay message %s: %v", msg.ID, err)
		} else {
			om.cache.MarkSent(msg.ID)
		}
	}

	log.Printf("Replayed %d cached messages", len(pending))
}

func generateMsgID() string {
	return "msg-" + time.Now().Format("20060102150405")
}