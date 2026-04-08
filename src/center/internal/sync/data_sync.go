package sync

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SyncOperation 同步操作类型
type SyncOperation string

const (
	SyncOpCreate  SyncOperation = "create"
	SyncOpUpdate  SyncOperation = "update"
	SyncOpDelete  SyncOperation = "delete"
	SyncOpMerge   SyncOperation = "merge"
)

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "pending"
	SyncStatusSyncing   SyncStatus = "syncing"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusConflict  SyncStatus = "conflict"
)

// DataConflictStrategy 数据冲突解决策略
type DataConflictStrategy string

const (
	DataConflictStrategyLastWrite  DataConflictStrategy = "last_write"   // 最后写入胜
	DataConflictStrategyFirstWrite DataConflictStrategy = "first_write"  // 最先写入胜
	DataConflictStrategyMerge      DataConflictStrategy = "merge"        // 合并
	DataConflictStrategyManual     DataConflictStrategy = "manual"       // 手动解决
	DataConflictStrategyServer     DataConflictStrategy = "server"       // 服务端优先
	DataConflictStrategyClient     DataConflictStrategy = "client"       // 客户端优先
)

// SyncRecord 同步记录
type SyncRecord struct {
	RecordID      string                 `json:"record_id"`
	IdentityID    string                 `json:"identity_id"`
	AgentID       string                 `json:"agent_id"`
	DataType      string                 `json:"data_type"`      // identity/memory/preference/behavior
	DataKey       string                 `json:"data_key"`       // 数据键
	Operation     SyncOperation          `json:"operation"`
	Version       int64                  `json:"version"`        // 数据版本号
	Timestamp     time.Time              `json:"timestamp"`
	OldValue      map[string]interface{} `json:"old_value,omitempty"`
	NewValue      map[string]interface{} `json:"new_value"`
	Checksum      string                 `json:"checksum"`       // 数据校验和
	Status        SyncStatus             `json:"status"`
	ConflictWith  string                 `json:"conflict_with,omitempty"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy    string                 `json:"resolved_by,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

// SyncBatch 同步批次
type SyncBatch struct {
	BatchID     string        `json:"batch_id"`
	IdentityID  string        `json:"identity_id"`
	AgentID     string        `json:"agent_id"`
	Records     []*SyncRecord `json:"records"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Status      SyncStatus    `json:"status"`
	TotalSize   int64         `json:"total_size"`
	Processed   int           `json:"processed"`
	Conflicts   int           `json:"conflicts"`
}

// DataVersion 数据版本
type DataVersion struct {
	IdentityID  string    `json:"identity_id"`
	DataType    string    `json:"data_type"`
	DataKey     string    `json:"data_key"`
	Version     int64     `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`
	Checksum    string    `json:"checksum"`
	IsDeleted   bool      `json:"is_deleted"`
}

// ConflictRecord 冲突记录
type ConflictRecord struct {
	ConflictID    string                 `json:"conflict_id"`
	IdentityID    string                 `json:"identity_id"`
	DataType      string                 `json:"data_type"`
	DataKey       string                 `json:"data_key"`
	ServerVersion *DataVersion           `json:"server_version"`
	ClientRecords []*SyncRecord          `json:"client_records"`
	Strategy      DataConflictStrategy   `json:"strategy"`
	Status        SyncStatus             `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	ResolvedValue map[string]interface{} `json:"resolved_value,omitempty"`
}

// DataSyncService 数据同步服务
type DataSyncService struct {
	mu               sync.RWMutex
	records          map[string]*SyncRecord           // recordID -> record
	versions         map[string]*DataVersion          // "identity_type_key" -> version
	pendingRecords   map[string][]*SyncRecord         // identityID -> pending records
	conflicts        map[string]*ConflictRecord       // conflictID -> conflict
	batches          map[string]*SyncBatch            // batchID -> batch
	syncQueues       map[string][]*SyncRecord         // identityID -> sync queue
	config           DataSyncConfig
	listeners        []SyncListener
	dataStore        map[string]map[string]interface{} // 简化的数据存储
}

// DataSyncConfig 同步配置
type DataSyncConfig struct {
	MaxBatchSize        int                  `json:"max_batch_size"`
	MaxRetryCount       int                  `json:"max_retry_count"`
	SyncInterval        time.Duration        `json:"sync_interval"`
	ConflictStrategy    DataConflictStrategy `json:"conflict_strategy"`
	EnableAutoMerge     bool                 `json:"enable_auto_merge"`
	VersionGapThreshold int64                `json:"version_gap_threshold"`
}

// DefaultDataSyncConfig 默认配置
func DefaultDataSyncConfig() DataSyncConfig {
	return DataSyncConfig{
		MaxBatchSize:        100,
		MaxRetryCount:       3,
		SyncInterval:        30 * time.Second,
		ConflictStrategy:    DataConflictStrategyLastWrite,
		EnableAutoMerge:     true,
		VersionGapThreshold: 10,
	}
}

// SyncListener 同步事件监听器
type SyncListener interface {
	OnSyncStarted(batch *SyncBatch)
	OnSyncCompleted(batch *SyncBatch)
	OnSyncFailed(batch *SyncBatch, err error)
	OnConflictDetected(conflict *ConflictRecord)
	OnConflictResolved(conflict *ConflictRecord)
	OnRecordSynced(record *SyncRecord)
}

// NewDataSyncService 创建同步服务
func NewDataSyncService(config DataSyncConfig) *DataSyncService {
	return &DataSyncService{
		records:        make(map[string]*SyncRecord),
		versions:       make(map[string]*DataVersion),
		pendingRecords: make(map[string][]*SyncRecord),
		conflicts:      make(map[string]*ConflictRecord),
		batches:        make(map[string]*SyncBatch),
		syncQueues:     make(map[string][]*SyncRecord),
		config:         config,
		listeners:      make([]SyncListener, 0),
		dataStore:      make(map[string]map[string]interface{}),
	}
}

// AddListener 添加监听器
func (s *DataSyncService) AddListener(listener SyncListener) {
	s.mu.Lock()
	s.listeners = append(s.listeners, listener)
	s.mu.Unlock()
}

// === 增量同步 ===

// CreateSyncRecord 创建同步记录
func (s *DataSyncService) CreateSyncRecord(identityID, agentID, dataType, dataKey string,
	operation SyncOperation, newValue map[string]interface{}) *SyncRecord {

	s.mu.Lock()
	defer s.mu.Unlock()

	recordID := generateRecordID(identityID, dataType, dataKey)
	versionKey := s.getVersionKey(identityID, dataType, dataKey)

	// 获取当前版本
	var currentVersion int64 = 0
	if v, exists := s.versions[versionKey]; exists {
		currentVersion = v.Version
	}

	// 计算校验和
	checksum := calculateChecksum(newValue)

	record := &SyncRecord{
		RecordID:   recordID,
		IdentityID: identityID,
		AgentID:    agentID,
		DataType:   dataType,
		DataKey:    dataKey,
		Operation:  operation,
		Version:    currentVersion + 1,
		Timestamp:  time.Now(),
		NewValue:   newValue,
		Checksum:   checksum,
		Status:     SyncStatusPending,
		RetryCount: 0,
	}

	s.records[recordID] = record
	s.pendingRecords[identityID] = append(s.pendingRecords[identityID], record)
	s.syncQueues[identityID] = append(s.syncQueues[identityID], record)

	return record
}

// GetPendingRecords 获取待同步记录
func (s *DataSyncService) GetPendingRecords(identityID string, limit int) []*SyncRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.pendingRecords[identityID]
	if len(records) == 0 {
		return nil
	}

	if limit > 0 && len(records) > limit {
		return records[:limit]
	}

	return records
}

// GetSyncQueue 获取同步队列
func (s *DataSyncService) GetSyncQueue(identityID string) []*SyncRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.syncQueues[identityID]
}

// CreateSyncBatch 创建同步批次
func (s *DataSyncService) CreateSyncBatch(identityID, agentID string) *SyncBatch {
	s.mu.Lock()
	defer s.mu.Unlock()

	pending := s.pendingRecords[identityID]
	if len(pending) == 0 {
		return nil
	}

	// 限制批次大小
	batchSize := s.config.MaxBatchSize
	if len(pending) < batchSize {
		batchSize = len(pending)
	}

	batchID := generateBatchID(identityID)
	batch := &SyncBatch{
		BatchID:    batchID,
		IdentityID: identityID,
		AgentID:    agentID,
		Records:    pending[:batchSize],
		StartTime:  time.Now(),
		Status:     SyncStatusPending,
		Processed:  0,
		Conflicts:  0,
	}

	s.batches[batchID] = batch

	// 从待处理队列移除
	s.pendingRecords[identityID] = pending[batchSize:]

	return batch
}

// ProcessSyncBatch 处理同步批次
func (s *DataSyncService) ProcessSyncBatch(batchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch := s.batches[batchID]
	if batch == nil {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	batch.Status = SyncStatusSyncing
	s.notifySyncStarted(batch)

	var batchErr error

	for _, record := range batch.Records {
		// 检查冲突
		if conflict := s.checkConflict(record); conflict != nil {
			batch.Conflicts++
			record.Status = SyncStatusConflict
			record.ConflictWith = conflict.ConflictID
			s.conflicts[conflict.ConflictID] = conflict
			s.notifyConflictDetected(conflict)
			continue
		}

		// 应用变更
		if err := s.applyRecord(record); err != nil {
			record.Status = SyncStatusFailed
			record.ErrorMessage = err.Error()
			batchErr = err
			continue
		}

		record.Status = SyncStatusCompleted
		batch.Processed++
		s.notifyRecordSynced(record)
	}

	// 更新批次状态
	now := time.Now()
	batch.EndTime = &now

	if batch.Conflicts > 0 {
		batch.Status = SyncStatusConflict
	} else if batchErr != nil {
		batch.Status = SyncStatusFailed
		s.notifySyncFailed(batch, batchErr)
	} else {
		batch.Status = SyncStatusCompleted
		s.notifySyncCompleted(batch)
	}

	return batchErr
}

// checkConflict 检查冲突
func (s *DataSyncService) checkConflict(record *SyncRecord) *ConflictRecord {
	versionKey := s.getVersionKey(record.IdentityID, record.DataType, record.DataKey)

	currentVersion, exists := s.versions[versionKey]
	if !exists {
		return nil // 没有现有数据，无冲突
	}

	// 检查版本差距
	if currentVersion.Version >= record.Version {
		// 版本不匹配，可能存在冲突
		if currentVersion.Checksum != record.Checksum {
			return &ConflictRecord{
				ConflictID:    generateConflictID(record),
				IdentityID:    record.IdentityID,
				DataType:      record.DataType,
				DataKey:       record.DataKey,
				ServerVersion: currentVersion,
				ClientRecords: []*SyncRecord{record},
				Strategy:      s.config.ConflictStrategy,
				Status:        SyncStatusConflict,
				CreatedAt:     time.Now(),
			}
		}
	}

	return nil
}

// applyRecord 应用记录
func (s *DataSyncService) applyRecord(record *SyncRecord) error {
	// 更新数据存储
	storeKey := record.DataType
	if s.dataStore[storeKey] == nil {
		s.dataStore[storeKey] = make(map[string]interface{})
	}

	switch record.Operation {
	case SyncOpCreate, SyncOpUpdate, SyncOpMerge:
		s.dataStore[storeKey][record.DataKey] = record.NewValue
	case SyncOpDelete:
		delete(s.dataStore[storeKey], record.DataKey)
	}

	// 更新版本
	versionKey := s.getVersionKey(record.IdentityID, record.DataType, record.DataKey)
	s.versions[versionKey] = &DataVersion{
		IdentityID: record.IdentityID,
		DataType:   record.DataType,
		DataKey:    record.DataKey,
		Version:    record.Version,
		UpdatedAt:  record.Timestamp,
		UpdatedBy:  record.AgentID,
		Checksum:   record.Checksum,
		IsDeleted:  record.Operation == SyncOpDelete,
	}

	return nil
}

// === 冲突解决 ===

// ResolveConflict 解决冲突
func (s *DataSyncService) ResolveConflict(conflictID string, strategy DataConflictStrategy,
	resolvedValue map[string]interface{}) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	conflict := s.conflicts[conflictID]
	if conflict == nil {
		return fmt.Errorf("conflict not found: %s", conflictID)
	}

	if len(resolvedValue) == 0 {
		// 自动合并
		if s.config.EnableAutoMerge {
			resolvedValue = s.autoMerge(conflict)
		}
	}

	if len(resolvedValue) == 0 {
		return fmt.Errorf("no resolved value provided")
	}

	// 创建合并记录
	mergeRecord := &SyncRecord{
		RecordID:   generateRecordID(conflict.IdentityID, conflict.DataType, conflict.DataKey),
		IdentityID: conflict.IdentityID,
		AgentID:    "conflict_resolver",
		DataType:   conflict.DataType,
		DataKey:    conflict.DataKey,
		Operation:  SyncOpMerge,
		Version:    conflict.ServerVersion.Version + 1,
		Timestamp:  time.Now(),
		NewValue:   resolvedValue,
		Checksum:   calculateChecksum(resolvedValue),
		Status:     SyncStatusCompleted,
	}

	// 应用合并结果
	if err := s.applyRecord(mergeRecord); err != nil {
		return err
	}

	// 更新冲突状态
	now := time.Now()
	conflict.Status = SyncStatusCompleted
	conflict.ResolvedAt = &now
	conflict.ResolvedValue = resolvedValue

	// 更新客户端记录状态
	for _, clientRecord := range conflict.ClientRecords {
		clientRecord.Status = SyncStatusCompleted
	}

	s.notifyConflictResolved(conflict)

	return nil
}

// autoMerge 自动合并
func (s *DataSyncService) autoMerge(conflict *ConflictRecord) map[string]interface{} {
	strategy := conflict.Strategy
	if strategy == "" {
		strategy = s.config.ConflictStrategy
	}

	switch strategy {
	case DataConflictStrategyServer:
		// 服务端优先
		if conflict.ServerVersion != nil {
			return map[string]interface{}{
				"version": conflict.ServerVersion.Version,
				"source":  "server",
			}
		}

	case DataConflictStrategyClient:
		// 客户端优先
		if len(conflict.ClientRecords) > 0 {
			return conflict.ClientRecords[0].NewValue
		}

	case DataConflictStrategyLastWrite:
		// 最后写入胜
		if conflict.ServerVersion != nil && len(conflict.ClientRecords) > 0 {
			if conflict.ClientRecords[0].Timestamp.After(conflict.ServerVersion.UpdatedAt) {
				return conflict.ClientRecords[0].NewValue
			}
		}

	case DataConflictStrategyFirstWrite:
		// 最先写入胜
		if conflict.ServerVersion != nil {
			return map[string]interface{}{
				"version": conflict.ServerVersion.Version,
			}
		}

	case DataConflictStrategyMerge:
		// 深度合并
		return s.deepMerge(conflict)
	}

	return nil
}

// deepMerge 深度合并
func (s *DataSyncService) deepMerge(conflict *ConflictRecord) map[string]interface{} {
	result := make(map[string]interface{})

	// 从服务端版本开始
	if conflict.ServerVersion != nil {
		// 获取服务端数据（这里简化处理）
		result["server_version"] = conflict.ServerVersion.Version
	}

	// 合并客户端数据
	for _, clientRecord := range conflict.ClientRecords {
		for k, v := range clientRecord.NewValue {
			// 简单的合并策略：客户端值覆盖服务端值
			if _, exists := result[k]; !exists || k != "server_version" {
				result[k] = v
			}
		}
	}

	result["merged"] = true
	result["merged_at"] = time.Now().Format(time.RFC3339)

	return result
}

// GetConflicts 获取冲突列表
func (s *DataSyncService) GetConflicts(identityID string) []*ConflictRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conflicts := make([]*ConflictRecord, 0)
	for _, c := range s.conflicts {
		if c.IdentityID == identityID && c.Status == SyncStatusConflict {
			conflicts = append(conflicts, c)
		}
	}
	return conflicts
}

// === 版本管理 ===

// GetVersion 获取数据版本
func (s *DataSyncService) GetVersion(identityID, dataType, dataKey string) *DataVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versionKey := s.getVersionKey(identityID, dataType, dataKey)
	return s.versions[versionKey]
}

// GetVersions 获取身份的所有版本
func (s *DataSyncService) GetVersions(identityID string) []*DataVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions := make([]*DataVersion, 0)
	for _, v := range s.versions {
		if v.IdentityID == identityID {
			versions = append(versions, v)
		}
	}
	return versions
}

// GetVersionDiff 获取版本差异
func (s *DataSyncService) GetVersionDiff(identityID string, sinceVersion int64) []*SyncRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]*SyncRecord, 0)
	for _, r := range s.records {
		if r.IdentityID == identityID && r.Version > sinceVersion && r.Status == SyncStatusCompleted {
			records = append(records, r)
		}
	}
	return records
}

// === 重试机制 ===

// RetryFailedRecords 重试失败的记录
func (s *DataSyncService) RetryFailedRecords(identityID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	retried := 0
	for _, record := range s.records {
		if record.IdentityID == identityID &&
			record.Status == SyncStatusFailed &&
			record.RetryCount < s.config.MaxRetryCount {

			record.Status = SyncStatusPending
			record.RetryCount++
			record.ErrorMessage = ""
			s.pendingRecords[identityID] = append(s.pendingRecords[identityID], record)
			retried++
		}
	}
	return retried
}

// === 统计信息 ===

// GetSyncStats 获取同步统计
func (s *DataSyncService) GetSyncStats(identityID string) *SyncStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SyncStats{
		IdentityID:      identityID,
		PendingRecords:  len(s.pendingRecords[identityID]),
		QueuedRecords:   len(s.syncQueues[identityID]),
		PendingConflicts: 0,
		CompletedBatches: 0,
		FailedBatches:    0,
	}

	for _, c := range s.conflicts {
		if c.IdentityID == identityID && c.Status == SyncStatusConflict {
			stats.PendingConflicts++
		}
	}

	for _, b := range s.batches {
		if b.IdentityID == identityID {
			if b.Status == SyncStatusCompleted {
				stats.CompletedBatches++
			} else if b.Status == SyncStatusFailed {
				stats.FailedBatches++
			}
		}
	}

	return stats
}

// SyncStats 同步统计
type SyncStats struct {
	IdentityID       string `json:"identity_id"`
	PendingRecords   int    `json:"pending_records"`
	QueuedRecords    int    `json:"queued_records"`
	PendingConflicts int    `json:"pending_conflicts"`
	CompletedBatches int    `json:"completed_batches"`
	FailedBatches    int    `json:"failed_batches"`
}

// === JSON 序列化 ===

// ToJSON 序列化同步记录
func (r *SyncRecord) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// SyncRecordFromJSON 从 JSON 解析
func SyncRecordFromJSON(data []byte) (*SyncRecord, error) {
	record := &SyncRecord{}
	if err := json.Unmarshal(data, record); err != nil {
		return nil, err
	}
	return record, nil
}

// ToJSON 序列化冲突记录
func (c *ConflictRecord) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// ConflictRecordFromJSON 从 JSON 解析
func ConflictRecordFromJSON(data []byte) (*ConflictRecord, error) {
	conflict := &ConflictRecord{}
	if err := json.Unmarshal(data, conflict); err != nil {
		return nil, err
	}
	return conflict, nil
}

// === 辅助方法 ===

func (s *DataSyncService) getVersionKey(identityID, dataType, dataKey string) string {
	return fmt.Sprintf("%s_%s_%s", identityID, dataType, dataKey)
}

func generateRecordID(identityID, dataType, dataKey string) string {
	return fmt.Sprintf("record-%s-%s-%s-%d", identityID, dataType, dataKey, time.Now().UnixNano())
}

func generateBatchID(identityID string) string {
	return fmt.Sprintf("batch-%s-%d", identityID, time.Now().UnixNano())
}

func generateConflictID(record *SyncRecord) string {
	return fmt.Sprintf("conflict-%s-%d", record.RecordID, time.Now().UnixNano())
}

func calculateChecksum(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	// 简化的校验和计算
	return fmt.Sprintf("%x", len(data))
}

// === 监听器通知 ===

func (s *DataSyncService) notifySyncStarted(batch *SyncBatch) {
	for _, l := range s.listeners {
		l.OnSyncStarted(batch)
	}
}

func (s *DataSyncService) notifySyncCompleted(batch *SyncBatch) {
	for _, l := range s.listeners {
		l.OnSyncCompleted(batch)
	}
}

func (s *DataSyncService) notifySyncFailed(batch *SyncBatch, err error) {
	for _, l := range s.listeners {
		l.OnSyncFailed(batch, err)
	}
}

func (s *DataSyncService) notifyConflictDetected(conflict *ConflictRecord) {
	for _, l := range s.listeners {
		l.OnConflictDetected(conflict)
	}
}

func (s *DataSyncService) notifyConflictResolved(conflict *ConflictRecord) {
	for _, l := range s.listeners {
		l.OnConflictResolved(conflict)
	}
}

func (s *DataSyncService) notifyRecordSynced(record *SyncRecord) {
	for _, l := range s.listeners {
		l.OnRecordSynced(record)
	}
}