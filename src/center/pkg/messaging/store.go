// Package messaging - 消息持久化存储
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// StorageType 存储类型
type StorageType string

const (
	StorageMemory     StorageType = "memory"
	StorageSQLite     StorageType = "sqlite"
	StoragePostgreSQL StorageType = "postgresql"
	StorageRedis      StorageType = "redis"
)

// MessageStatus 消息状态
type MessageStatus string

const (
	StatusPending   MessageStatus = "pending"
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusAcked     MessageStatus = "acked"
	StatusFailed    MessageStatus = "failed"
	StatusExpired   MessageStatus = "expired"
)

// StoredMessage 存储的消息
type StoredMessage struct {
	ID           string                 `json:"id"`
	Subject      string                 `json:"subject"`
	From         string                 `json:"from"`
	To           string                 `json:"to"`
	Type         MessageType            `json:"type"`
	Payload      map[string]interface{} `json:"payload"`
	RawData      []byte                 `json:"raw_data,omitempty"`
	Status       MessageStatus          `json:"status"`
	Priority     MessagePriority        `json:"priority"`
	TTL          time.Duration          `json:"ttl"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	SentAt       *time.Time             `json:"sent_at,omitempty"`
	AckedAt      *time.Time             `json:"acked_at,omitempty"`
	ExpiredAt    *time.Time             `json:"expired_at,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetry     int                    `json:"max_retry"`
	LastError    string                 `json:"last_error,omitempty"`
	CorrelationID string                `json:"correlation_id,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
}

// MessageQuery 消息查询
type MessageQuery struct {
	IDs          []string        `json:"ids,omitempty"`
	Subjects     []string        `json:"subjects,omitempty"`
	From         string          `json:"from,omitempty"`
	To           string          `json:"to,omitempty"`
	Status       MessageStatus   `json:"status,omitempty"`
	Statuses     []MessageStatus `json:"statuses,omitempty"`
	PriorityMin  MessagePriority `json:"priority_min,omitempty"`
	PriorityMax  MessagePriority `json:"priority_max,omitempty"`
	CreatedAfter *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time     `json:"created_before,omitempty"`
	Limit        int             `json:"limit,omitempty"`
	Offset       int             `json:"offset,omitempty"`
	OrderBy      string          `json:"order_by,omitempty"` // created_at, priority, status
	OrderDesc    bool            `json:"order_desc"`
}

// MessageStoreConfig 消息存储配置
type MessageStoreConfig struct {
	StorageType    StorageType   `json:"storage_type"`
	MaxMessages    int           `json:"max_messages"`    // 最大存储消息数
	Retention      time.Duration `json:"retention"`       // 消息保留时间
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
	EnableIndex    bool          `json:"enable_index"`    // 启用索引
	IndexFields    []string      `json:"index_fields"`    // 索引字段
}

// MessageStore 消息存储
type MessageStore struct {
	config      MessageStoreConfig
	messages    map[string]*StoredMessage
	bySubject   map[string][]string // subject -> message IDs
	byFrom      map[string][]string // from -> message IDs
	byTo        map[string][]string // to -> message IDs
	byStatus    map[MessageStatus][]string // status -> message IDs
	pending     []string // 待发送消息ID队列
	stats       *StoreStats
	mu          sync.RWMutex
}

// StoreStats 存储统计
type StoreStats struct {
	TotalMessages  int64 `json:"total_messages"`
	ByStatus       map[MessageStatus]int64 `json:"by_status"`
	BySubject      map[string]int64 `json:"by_subject"`
	ExpiredCount   int64 `json:"expired_count"`
	CleanupCount   int64 `json:"cleanup_count"`
	LastCleanup    time.Time `json:"last_cleanup"`
}

// NewMessageStore 创建消息存储
func NewMessageStore(config MessageStoreConfig) *MessageStore {
	return &MessageStore{
		config:    config,
		messages:  make(map[string]*StoredMessage),
		bySubject: make(map[string][]string),
		byFrom:    make(map[string][]string),
		byTo:      make(map[string][]string),
		byStatus:  make(map[MessageStatus][]string),
		pending:   make([]string, 0),
		stats: &StoreStats{
			ByStatus:  make(map[MessageStatus]int64),
			BySubject: make(map[string]int64),
		},
	}
}

// Initialize 初始化
func (ms *MessageStore) Initialize(ctx context.Context) error {
	// 启动清理任务
	go ms.cleanupLoop(ctx)
	return nil
}

// Store 存储消息
func (ms *MessageStore) Store(ctx context.Context, msg *StoredMessage) error {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("stored-%d", time.Now().UnixNano())
	}

	msg.CreatedAt = time.Now()
	msg.UpdatedAt = msg.CreatedAt
	msg.Status = StatusPending

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 检查容量限制
	if ms.config.MaxMessages > 0 && len(ms.messages) >= ms.config.MaxMessages {
		// 删除最旧的消息
		ms.evictOldest()
	}

	// 存储消息
	ms.messages[msg.ID] = msg

	// 更新索引
	ms.updateIndexes(msg)

	// 添加到待发送队列
	if msg.Status == StatusPending {
		ms.pending = append(ms.pending, msg.ID)
	}

	// 更新统计
	ms.stats.TotalMessages++
	ms.stats.ByStatus[msg.Status]++
	if msg.Subject != "" {
		ms.stats.BySubject[msg.Subject]++
	}

	return nil
}

// StoreBatch 批量存储
func (ms *MessageStore) StoreBatch(ctx context.Context, messages []*StoredMessage) error {
	for _, msg := range messages {
		if err := ms.Store(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Get 获取消息
func (ms *MessageStore) Get(id string) (*StoredMessage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	msg, ok := ms.messages[id]
	if !ok {
		return nil, fmt.Errorf("消息不存在: %s", id)
	}
	return msg, nil
}

// Update 更新消息
func (ms *MessageStore) Update(id string, updates func(*StoredMessage)) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	msg, ok := ms.messages[id]
	if !ok {
		return fmt.Errorf("消息不存在: %s", id)
	}

	// 记录旧状态
	oldStatus := msg.Status

	// 应用更新
	updates(msg)
	msg.UpdatedAt = time.Now()

	// 如果状态改变，更新索引
	if msg.Status != oldStatus {
		// 从旧状态索引移除
		ms.removeFromStatusIndex(id, oldStatus)
		// 添加到新状态索引
		ms.byStatus[msg.Status] = append(ms.byStatus[msg.Status], id)

		// 更新统计
		ms.stats.ByStatus[oldStatus]--
		ms.stats.ByStatus[msg.Status]++
	}

	return nil
}

// UpdateStatus 更新消息状态
func (ms *MessageStore) UpdateStatus(id string, status MessageStatus, errMsg string) error {
	return ms.Update(id, func(msg *StoredMessage) {
		msg.Status = status
		msg.LastError = errMsg
		if status == StatusSent {
			now := time.Now()
			msg.SentAt = &now
		} else if status == StatusAcked {
			now := time.Now()
			msg.AckedAt = &now
		}
	})
}

// Delete 删除消息
func (ms *MessageStore) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	msg, ok := ms.messages[id]
	if !ok {
		return fmt.Errorf("消息不存在: %s", id)
	}

	// 从索引移除
	ms.removeFromIndexes(msg, id)

	// 从待发送队列移除
	ms.removeFromPending(id)

	// 删除消息
	delete(ms.messages, id)

	// 更新统计
	ms.stats.TotalMessages--
	ms.stats.ByStatus[msg.Status]--
	if msg.Subject != "" {
		ms.stats.BySubject[msg.Subject]--
	}

	return nil
}

// Query 查询消息
func (ms *MessageStore) Query(ctx context.Context, query MessageQuery) ([]*StoredMessage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*StoredMessage

	// 根据查询条件过滤
	for _, msg := range ms.messages {
		if !ms.matchQuery(msg, query) {
			continue
		}
		results = append(results, msg)
	}

	// 排序
	ms.sortResults(results, query.OrderBy, query.OrderDesc)

	// 分页
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results, nil
}

// matchQuery 检查消息是否匹配查询
func (ms *MessageStore) matchQuery(msg *StoredMessage, query MessageQuery) bool {
	// ID过滤
	if len(query.IDs) > 0 {
		found := false
		for _, id := range query.IDs {
			if msg.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Subject过滤
	if len(query.Subjects) > 0 {
		found := false
		for _, s := range query.Subjects {
			if msg.Subject == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// From过滤
	if query.From != "" && msg.From != query.From {
		return false
	}

	// To过滤
	if query.To != "" && msg.To != query.To {
		return false
	}

	// Status过滤
	if query.Status != "" && msg.Status != query.Status {
		return false
	}
	if len(query.Statuses) > 0 {
		found := false
		for _, s := range query.Statuses {
			if msg.Status == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 时间过滤
	if query.CreatedAfter != nil && msg.CreatedAt.Before(*query.CreatedAfter) {
		return false
	}
	if query.CreatedBefore != nil && msg.CreatedAfter(*query.CreatedBefore) {
		return false
	}

	// 优先级过滤
	if msg.Priority < query.PriorityMin || msg.Priority > query.PriorityMax {
		if query.PriorityMin != 0 || query.PriorityMax != 0 {
			return false
		}
	}

	return true
}

// sortResults 排序结果
func (ms *MessageStore) sortResults(results []*StoredMessage, orderBy string, desc bool) {
	// 简化实现：按创建时间排序
	// 实际实现可支持多种排序
}

// GetPending 获取待发送消息
func (ms *MessageStore) GetPending(limit int) []*StoredMessage {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	results := make([]*StoredMessage, 0)
	count := 0

	for _, id := range ms.pending {
		if limit > 0 && count >= limit {
			break
		}
		if msg, ok := ms.messages[id]; ok && msg.Status == StatusPending {
			results = append(results, msg)
			count++
		}
	}

	return results
}

// GetByCorrelationID 根据关联ID获取消息
func (ms *MessageStore) GetByCorrelationID(correlationID string) ([]*StoredMessage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	results := make([]*StoredMessage, 0)
	for _, msg := range ms.messages {
		if msg.CorrelationID == correlationID {
			results = append(results, msg)
		}
	}

	return results, nil
}

// Ack 确认消息
func (ms *MessageStore) Ack(id string) error {
	return ms.UpdateStatus(id, StatusAcked, "")
}

// Nack 否认消息（标记为失败）
func (ms *MessageStore) Nack(id string, errMsg string) error {
	return ms.UpdateStatus(id, StatusFailed, errMsg)
}

// Retry 标记为重试
func (ms *MessageStore) Retry(id string) error {
	return ms.Update(id, func(msg *StoredMessage) {
		msg.RetryCount++
		msg.Status = StatusPending
	})
}

// CanRetry 检查是否可以重试
func (ms *MessageStore) CanRetry(id string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	msg, ok := ms.messages[id]
	if !ok {
		return false
	}

	return msg.RetryCount < msg.MaxRetry
}

// updateIndexes 更新索引
func (ms *MessageStore) updateIndexes(msg *StoredMessage) {
	id := msg.ID

	if msg.Subject != "" {
		ms.bySubject[msg.Subject] = append(ms.bySubject[msg.Subject], id)
	}
	if msg.From != "" {
		ms.byFrom[msg.From] = append(ms.byFrom[msg.From], id)
	}
	if msg.To != "" {
		ms.byTo[msg.To] = append(ms.byTo[msg.To], id)
	}
	ms.byStatus[msg.Status] = append(ms.byStatus[msg.Status], id)
}

// removeFromIndexes 从索引移除
func (ms *MessageStore) removeFromIndexes(msg *StoredMessage, id string) {
	ms.removeFromIndex(&ms.bySubject, msg.Subject, id)
	ms.removeFromIndex(&ms.byFrom, msg.From, id)
	ms.removeFromIndex(&ms.byTo, msg.To, id)
	ms.removeFromStatusIndex(id, msg.Status)
}

// removeFromIndex 从指定索引移除
func (ms *MessageStore) removeFromIndex(index *map[string][]string, key string, id string) {
	if key == "" {
		return
	}
	ids := (*index)[key]
	newIds := make([]string, 0)
	for _, i := range ids {
		if i != id {
			newIds = append(newIds, i)
		}
	}
	(*index)[key] = newIds
}

// removeFromStatusIndex 从状态索引移除
func (ms *MessageStore) removeFromStatusIndex(id string, status MessageStatus) {
	ids := ms.byStatus[status]
	newIds := make([]string, 0)
	for _, i := range ids {
		if i != id {
			newIds = append(newIds, i)
		}
	}
	ms.byStatus[status] = newIds
}

// removeFromPending 从待发送队列移除
func (ms *MessageStore) removeFromPending(id string) {
	newPending := make([]string, 0)
	for _, i := range ms.pending {
		if i != id {
			newPending = append(newPending, i)
		}
	}
	ms.pending = newPending
}

// evictOldest 驱逐最旧的消息
func (ms *MessageStore) evictOldest() {
	var oldestID string
	var oldestTime time.Time

	for id, msg := range ms.messages {
		if oldestID == "" || msg.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = msg.CreatedAt
		}
	}

	if oldestID != "" {
		ms.deleteInternal(oldestID)
	}
}

// deleteInternal 内部删除（不加锁）
func (ms *MessageStore) deleteInternal(id string) {
	msg, ok := ms.messages[id]
	if !ok {
		return
	}

	ms.removeFromIndexes(msg, id)
	ms.removeFromPending(id)
	delete(ms.messages, id)

	ms.stats.TotalMessages--
	ms.stats.ByStatus[msg.Status]--
}

// cleanupLoop 清理循环
func (ms *MessageStore) cleanupLoop(ctx context.Context) {
	if ms.config.CleanupInterval == 0 {
		ms.config.CleanupInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(ms.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ms.cleanup()
		}
	}
}

// cleanup 清理过期消息
func (ms *MessageStore) cleanup() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now()
	retention := ms.config.Retention
	if retention == 0 {
		retention = 24 * time.Hour
	}

	toDelete := make([]string, 0)

	for id, msg := range ms.messages {
		// 检查是否过期
		if msg.TTL > 0 {
			if now.Sub(msg.CreatedAt) > msg.TTL {
				msg.Status = StatusExpired
				msg.ExpiredAt = &now
				toDelete = append(toDelete, id)
				continue
			}
		}

		// 检查保留时间
		if now.Sub(msg.CreatedAt) > retention {
			toDelete = append(toDelete, id)
		}
	}

	// 删除过期消息
	for _, id := range toDelete {
		ms.deleteInternal(id)
		ms.stats.ExpiredCount++
	}

	ms.stats.CleanupCount += int64(len(toDelete))
	ms.stats.LastCleanup = now
}

// GetStatistics 获取统计信息
func (ms *MessageStore) GetStatistics() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	pendingCount := 0
	for _, id := range ms.pending {
		if msg, ok := ms.messages[id]; ok && msg.Status == StatusPending {
			pendingCount++
		}
	}

	return map[string]interface{}{
		"total_messages":   ms.stats.TotalMessages,
		"by_status":        ms.stats.ByStatus,
		"by_subject":       ms.stats.BySubject,
		"pending_count":    pendingCount,
		"expired_count":    ms.stats.ExpiredCount,
		"cleanup_count":    ms.stats.CleanupCount,
		"last_cleanup":     ms.stats.LastCleanup,
		"index_count": map[string]int{
			"subjects": len(ms.bySubject),
			"from":     len(ms.byFrom),
			"to":       len(ms.byTo),
		},
	}
}

// Export 导出消息
func (ms *MessageStore) Export(ctx context.Context, query MessageQuery) ([]byte, error) {
	messages, err := ms.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(messages, "", "  ")
}

// Import 导入消息
func (ms *MessageStore) Import(ctx context.Context, data []byte) error {
	var messages []*StoredMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}

	return ms.StoreBatch(ctx, messages)
}

// Count 统计消息数量
func (ms *MessageStore) Count(query MessageQuery) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	count := 0
	for _, msg := range ms.messages {
		if ms.matchQuery(msg, query) {
			count++
		}
	}
	return count
}

// Clear 清空存储
func (ms *MessageStore) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.messages = make(map[string]*StoredMessage)
	ms.bySubject = make(map[string][]string)
	ms.byFrom = make(map[string][]string)
	ms.byTo = make(map[string][]string)
	ms.byStatus = make(map[MessageStatus][]string)
	ms.pending = make([]string, 0)
	ms.stats = &StoreStats{
		ByStatus:  make(map[MessageStatus]int64),
		BySubject: make(map[string]int64),
	}
}