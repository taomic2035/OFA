// Package sync - 设备消息总线 (v3.0.0)
//
// Center 是永远在线的灵魂载体，消息总线提供：
// - Center 到设备的消息推送
// - 设备到设备的消息路由
// - 离线消息存储和投递
// - 消息优先级和过期管理
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// === 消息类型定义 ===

// MessageType 消息类型
type MessageType string

const (
	MessageTypeCommand    MessageType = "command"    // 命令消息
	MessageTypeNotification MessageType = "notification" // 通知消息
	MessageTypeData       MessageType = "data"       // 数据消息
	MessageTypeSync       MessageType = "sync"       // 同步消息
	MessageTypeAck        MessageType = "ack"        // 确认消息
	MessageTypeHeartbeat  MessageType = "heartbeat"  // 心跳消息
)

// MessagePriority 消息优先级
type MessagePriority int

const (
	PriorityLow      MessagePriority = 0
	PriorityNormal   MessagePriority = 1
	PriorityHigh     MessagePriority = 2
	PriorityUrgent   MessagePriority = 3
)

// MessageStatus 消息状态
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"   // 待发送
	MessageStatusSent      MessageStatus = "sent"      // 已发送
	MessageStatusDelivered MessageStatus = "delivered" // 已送达
	MessageStatusAcked     MessageStatus = "acked"     // 已确认
	MessageStatusFailed    MessageStatus = "failed"    // 发送失败
	MessageStatusExpired   MessageStatus = "expired"   // 已过期
)

// Message 消息定义
type Message struct {
	ID           string                 `json:"id"`
	FromAgent    string                 `json:"from_agent"`
	ToAgent      string                 `json:"to_agent"`
	IdentityID   string                 `json:"identity_id"`
	Type         MessageType            `json:"type"`
	Priority     MessagePriority        `json:"priority"`
	Payload      map[string]interface{} `json:"payload"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty"`
	AckedAt      *time.Time             `json:"acked_at,omitempty"`
	Status       MessageStatus          `json:"status"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// IsExpired 检查消息是否过期
func (m *Message) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

// ShouldRetry 检查是否应该重试
func (m *Message) ShouldRetry() bool {
	return m.Status == MessageStatusFailed &&
		m.RetryCount < m.MaxRetries &&
		!m.IsExpired()
}

// ToJSON 序列化消息
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 反序列化消息
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// === 消息总线 ===

// MessageBus 消息总线
type MessageBus struct {
	mu sync.RWMutex

	// 离线消息存储：agentId -> messages
	offlineMessages map[string][]*Message

	// 消息索引：messageId -> message
	messageIndex map[string]*Message

	// 待确认消息：agentId -> messageId -> sentTime
	pendingAcks map[string]map[string]time.Time

	// 配置
	config MessageBusConfig

	// 监听器
	listeners []MessageListener

	// 连接状态：agentId -> connected
	connectedAgents map[string]bool
}

// MessageBusConfig 消息总线配置
type MessageBusConfig struct {
	// 离线消息最大数量
	MaxOfflineMessages int `json:"max_offline_messages"`
	// 消息默认过期时间
	DefaultTTL time.Duration `json:"default_ttl"`
	// 最大重试次数
	MaxRetries int `json:"max_retries"`
	// 确认超时时间
	AckTimeout time.Duration `json:"ack_timeout"`
	// 清理间隔
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// DefaultMessageBusConfig 默认配置
func DefaultMessageBusConfig() MessageBusConfig {
	return MessageBusConfig{
		MaxOfflineMessages: 100,
		DefaultTTL:         24 * time.Hour,
		MaxRetries:         3,
		AckTimeout:         30 * time.Second,
		CleanupInterval:    5 * time.Minute,
	}
}

// MessageListener 消息监听器
type MessageListener interface {
	OnMessageSent(message *Message)
	OnMessageDelivered(message *Message)
	OnMessageAcked(message *Message)
	OnMessageFailed(message *Message, err error)
	OnAgentConnected(agentID string)
	OnAgentDisconnected(agentID string)
}

// NewMessageBus 创建消息总线
func NewMessageBus(config MessageBusConfig) *MessageBus {
	if config.MaxOfflineMessages == 0 {
		config = DefaultMessageBusConfig()
	}

	bus := &MessageBus{
		offlineMessages: make(map[string][]*Message),
		messageIndex:    make(map[string]*Message),
		pendingAcks:     make(map[string]map[string]time.Time),
		connectedAgents: make(map[string]bool),
		config:          config,
		listeners:       make([]MessageListener, 0),
	}

	// 启动清理协程
	go bus.cleanupLoop()

	return bus
}

// === 发送消息 ===

// SendMessage 发送消息
func (bus *MessageBus) SendMessage(ctx context.Context, msg *Message) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	// 设置默认值
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.ExpiresAt == nil && bus.config.DefaultTTL > 0 {
		exp := msg.CreatedAt.Add(bus.config.DefaultTTL)
		msg.ExpiresAt = &exp
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = bus.config.MaxRetries
	}
	msg.Status = MessageStatusPending

	// 检查是否过期
	if msg.IsExpired() {
		msg.Status = MessageStatusExpired
		return fmt.Errorf("message expired")
	}

	// 索引消息
	bus.messageIndex[msg.ID] = msg

	// 检查目标设备是否在线
	if bus.isAgentConnectedLocked(msg.ToAgent) {
		// 在线，直接发送
		return bus.sendToAgentLocked(msg)
	}

	// 离线，存储消息
	return bus.storeOfflineMessageLocked(msg)
}

// BroadcastMessage 广播消息到身份的所有设备
func (bus *MessageBus) BroadcastMessage(ctx context.Context, fromAgent, identityID string, msg *Message) ([]string, error) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	// 获取身份的所有设备
	// 这里需要与 DeviceManager 集成，暂时简化处理
	// 实际实现需要获取 identityID 对应的所有 agentID

	// 返回发送到的设备列表
	sentTo := make([]string, 0)

	// TODO: 从 DeviceManager 获取设备列表
	// agents := bus.deviceManager.GetDevicesByIdentity(identityID)

	log.Printf("Broadcast message from %s to identity %s", fromAgent, identityID)

	return sentTo, nil
}

// sendToAgentLocked 发送消息到在线设备（已持有锁）
func (bus *MessageBus) sendToAgentLocked(msg *Message) error {
	msg.Status = MessageStatusSent

	// 添加到待确认列表
	if bus.pendingAcks[msg.ToAgent] == nil {
		bus.pendingAcks[msg.ToAgent] = make(map[string]time.Time)
	}
	bus.pendingAcks[msg.ToAgent][msg.ID] = time.Now()

	// 触发监听器
	go bus.notifyMessageSent(msg)

	log.Printf("Message sent: %s -> %s, type=%s, priority=%d",
		msg.FromAgent, msg.ToAgent, msg.Type, msg.Priority)

	return nil
}

// storeOfflineMessageLocked 存储离线消息（已持有锁）
func (bus *MessageBus) storeOfflineMessageLocked(msg *Message) error {
	// 检查离线消息数量限制
	messages := bus.offlineMessages[msg.ToAgent]
	if len(messages) >= bus.config.MaxOfflineMessages {
		// 移除最旧的低优先级消息
		bus.removeOldestLowPriorityMessage(msg.ToAgent)
	}

	bus.offlineMessages[msg.ToAgent] = append(bus.offlineMessages[msg.ToAgent], msg)

	log.Printf("Message stored offline: %s -> %s", msg.FromAgent, msg.ToAgent)

	return nil
}

// === 接收消息 ===

// GetOfflineMessages 获取离线消息
func (bus *MessageBus) GetOfflineMessages(agentID string) []*Message {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	messages := bus.offlineMessages[agentID]
	if len(messages) == 0 {
		return nil
	}

	// 过滤过期消息
	valid := make([]*Message, 0, len(messages))
	for _, msg := range messages {
		if !msg.IsExpired() {
			msg.Status = MessageStatusDelivered
			now := time.Now()
			msg.DeliveredAt = &now
			valid = append(valid, msg)
		}
	}

	// 清空离线消息
	delete(bus.offlineMessages, agentID)

	log.Printf("Delivered %d offline messages to %s", len(valid), agentID)

	return valid
}

// AckMessage 确认消息
func (bus *MessageBus) AckMessage(agentID, messageID string) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	msg, exists := bus.messageIndex[messageID]
	if !exists {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// 更新状态
	msg.Status = MessageStatusAcked
	now := time.Now()
	msg.AckedAt = &now

	// 从待确认列表移除
	if pending, ok := bus.pendingAcks[agentID]; ok {
		delete(pending, messageID)
	}

	// 触发监听器
	go bus.notifyMessageAcked(msg)

	log.Printf("Message acked: %s by %s", messageID, agentID)

	return nil
}

// === 连接管理 ===

// AgentConnected 设备上线
func (bus *MessageBus) AgentConnected(agentID string) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.connectedAgents[agentID] = true

	// 触发监听器
	go bus.notifyAgentConnected(agentID)

	log.Printf("Agent connected: %s", agentID)
}

// AgentDisconnected 设备离线
func (bus *MessageBus) AgentDisconnected(agentID string) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.connectedAgents[agentID] = false

	// 触发监听器
	go bus.notifyAgentDisconnected(agentID)

	log.Printf("Agent disconnected: %s", agentID)
}

// IsAgentConnected 检查设备是否在线
func (bus *MessageBus) IsAgentConnected(agentID string) bool {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return bus.connectedAgents[agentID]
}

func (bus *MessageBus) isAgentConnectedLocked(agentID string) bool {
	return bus.connectedAgents[agentID]
}

// === 消息查询 ===

// GetMessage 获取消息
func (bus *MessageBus) GetMessage(messageID string) *Message {
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	return bus.messageIndex[messageID]
}

// GetPendingMessages 获取待确认消息
func (bus *MessageBus) GetPendingMessages(agentID string) []*Message {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	pending, ok := bus.pendingAcks[agentID]
	if !ok {
		return nil
	}

	messages := make([]*Message, 0, len(pending))
	for msgID := range pending {
		if msg, exists := bus.messageIndex[msgID]; exists {
			messages = append(messages, msg)
		}
	}

	return messages
}

// === 监听器管理 ===

// AddListener 添加监听器
func (bus *MessageBus) AddListener(listener MessageListener) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.listeners = append(bus.listeners, listener)
}

func (bus *MessageBus) notifyMessageSent(msg *Message) {
	for _, listener := range bus.listeners {
		listener.OnMessageSent(msg)
	}
}

func (bus *MessageBus) notifyMessageAcked(msg *Message) {
	for _, listener := range bus.listeners {
		listener.OnMessageAcked(msg)
	}
}

func (bus *MessageBus) notifyAgentConnected(agentID string) {
	for _, listener := range bus.listeners {
		listener.OnAgentConnected(agentID)
	}
}

func (bus *MessageBus) notifyAgentDisconnected(agentID string) {
	for _, listener := range bus.listeners {
		listener.OnAgentDisconnected(agentID)
	}
}

// === 清理和维护 ===

// cleanupLoop 清理循环
func (bus *MessageBus) cleanupLoop() {
	ticker := time.NewTicker(bus.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		bus.cleanup()
	}
}

// cleanup 清理过期消息
func (bus *MessageBus) cleanup() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	now := time.Now()

	// 清理过期消息
	for id, msg := range bus.messageIndex {
		if msg.IsExpired() {
			msg.Status = MessageStatusExpired
			delete(bus.messageIndex, id)
		}
	}

	// 清理超时的待确认消息
	for _, pending := range bus.pendingAcks {
		for msgID, sentTime := range pending {
			if now.Sub(sentTime) > bus.config.AckTimeout {
				if msg, exists := bus.messageIndex[msgID]; exists {
					msg.RetryCount++
					if msg.ShouldRetry() {
						// 重试
						go bus.retryMessage(msg)
					} else {
						msg.Status = MessageStatusFailed
					}
				}
				delete(pending, msgID)
			}
		}
	}

	// 清理离线消息中的过期消息
	for agentID, messages := range bus.offlineMessages {
		var valid []*Message
		for _, msg := range messages {
			if !msg.IsExpired() {
				valid = append(valid, msg)
			}
		}
		bus.offlineMessages[agentID] = valid
	}

	log.Printf("Message bus cleanup completed")
}

// retryMessage 重试发送消息
func (bus *MessageBus) retryMessage(msg *Message) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.isAgentConnectedLocked(msg.ToAgent) {
		bus.sendToAgentLocked(msg)
	}
}

// removeOldestLowPriorityMessage 移除最旧的低优先级消息
func (bus *MessageBus) removeOldestLowPriorityMessage(agentID string) {
	messages := bus.offlineMessages[agentID]
	if len(messages) == 0 {
		return
	}

	// 找到最旧的低优先级消息
	oldestIdx := -1
	var oldestTime time.Time

	for i, msg := range messages {
		if msg.Priority <= PriorityLow {
			if oldestIdx == -1 || msg.CreatedAt.Before(oldestTime) {
				oldestIdx = i
				oldestTime = msg.CreatedAt
			}
		}
	}

	if oldestIdx >= 0 {
		// 移除消息
		bus.offlineMessages[agentID] = append(
			messages[:oldestIdx],
			messages[oldestIdx+1:]...,
		)
		delete(bus.messageIndex, messages[oldestIdx].ID)
	}
}

// === 统计 ===

// GetStats 获取消息总线统计
func (bus *MessageBus) GetStats() MessageBusStats {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	stats := MessageBusStats{
		TotalMessages:    len(bus.messageIndex),
		ConnectedAgents:  0,
		OfflineMessages:  0,
		PendingAcks:      0,
	}

	for _, connected := range bus.connectedAgents {
		if connected {
			stats.ConnectedAgents++
		}
	}

	for _, messages := range bus.offlineMessages {
		stats.OfflineMessages += len(messages)
	}

	for _, pending := range bus.pendingAcks {
		stats.PendingAcks += len(pending)
	}

	return stats
}

// MessageBusStats 消息总线统计
type MessageBusStats struct {
	TotalMessages   int `json:"total_messages"`
	ConnectedAgents int `json:"connected_agents"`
	OfflineMessages int `json:"offline_messages"`
	PendingAcks     int `json:"pending_acks"`
}

// === 辅助函数 ===

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}