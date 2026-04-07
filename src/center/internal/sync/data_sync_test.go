package sync

import (
	"testing"
	"time"
)

func TestDataSyncService(t *testing.T) {
	config := DefaultDataSyncConfig()
	service := NewDataSyncService(config)

	// 测试初始化
	if service.records == nil {
		t.Error("records map should be initialized")
	}
	if service.versions == nil {
		t.Error("versions map should be initialized")
	}
	if service.pendingRecords == nil {
		t.Error("pendingRecords map should be initialized")
	}
}

func TestDataSyncServiceCreateRecord(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	newValue := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	record := service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, newValue)

	if record == nil {
		t.Fatal("Record should not be nil")
	}

	// 验证记录属性
	if record.IdentityID != "identity-001" {
		t.Errorf("Expected identity ID 'identity-001', got '%s'", record.IdentityID)
	}
	if record.DataType != "memory" {
		t.Errorf("Expected data type 'memory', got '%s'", record.DataType)
	}
	if record.Operation != SyncOpCreate {
		t.Errorf("Expected operation 'create', got '%s'", record.Operation)
	}
	if record.Version != 1 {
		t.Errorf("Expected version 1, got %d", record.Version)
	}
	if record.Status != SyncStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", record.Status)
	}

	// 验证待处理队列
	pending := service.GetPendingRecords("identity-001", 10)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending record, got %d", len(pending))
	}
}

func TestDataSyncServiceCreateBatch(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建多个记录
	for i := 0; i < 5; i++ {
		service.CreateSyncRecord("identity-001", "agent-001", "memory",
			"key-"+string(rune('0'+i)), SyncOpCreate, map[string]interface{}{"index": i})
	}

	// 创建批次
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	if batch == nil {
		t.Fatal("Batch should not be nil")
	}

	if batch.IdentityID != "identity-001" {
		t.Errorf("Expected identity ID 'identity-001', got '%s'", batch.IdentityID)
	}
	if len(batch.Records) != 5 {
		t.Errorf("Expected 5 records in batch, got %d", len(batch.Records))
	}
	if batch.Status != SyncStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", batch.Status)
	}

	// 验证批次存储
	stored := service.batches[batch.BatchID]
	if stored == nil {
		t.Error("Batch should be stored")
	}
}

func TestDataSyncServiceProcessBatch(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建记录和批次
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"name": "test"})

	batch := service.CreateSyncBatch("identity-001", "agent-001")

	// 处理批次
	err := service.ProcessSyncBatch(batch.BatchID)
	if err != nil {
		t.Fatalf("ProcessSyncBatch failed: %v", err)
	}

	// 验证批次状态
	updated := service.batches[batch.BatchID]
	if updated.Status != SyncStatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", updated.Status)
	}
	if updated.Processed != 1 {
		t.Errorf("Expected 1 processed, got %d", updated.Processed)
	}

	// 验证记录状态
	record := service.records[batch.Records[0].RecordID]
	if record.Status != SyncStatusCompleted {
		t.Errorf("Expected record status 'completed', got '%s'", record.Status)
	}

	// 验证版本更新
	version := service.GetVersion("identity-001", "memory", "key-001")
	if version == nil {
		t.Fatal("Version should not be nil")
	}
	if version.Version != 1 {
		t.Errorf("Expected version 1, got %d", version.Version)
	}
}

func TestDataSyncServiceVersionManagement(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建并同步记录
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"v": 1})
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	// 更新数据
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpUpdate, map[string]interface{}{"v": 2})
	batch = service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	// 验证版本
	version := service.GetVersion("identity-001", "memory", "key-001")
	if version == nil {
		t.Fatal("Version should not be nil")
	}
	if version.Version != 2 {
		t.Errorf("Expected version 2, got %d", version.Version)
	}

	// 获取版本差异
	diff := service.GetVersionDiff("identity-001", 0)
	if len(diff) != 2 {
		t.Errorf("Expected 2 records in diff, got %d", len(diff))
	}
}

func TestDataSyncServiceConflict(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建并同步记录
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"name": "original", "checksum": "abc"})
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	// 模拟冲突：创建一个版本号相同的记录，但校验和不同
	conflictRecord := &SyncRecord{
		RecordID:   "record-conflict",
		IdentityID: "identity-001",
		AgentID:    "agent-002",
		DataType:   "memory",
		DataKey:    "key-001",
		Operation:  SyncOpUpdate,
		Version:    1, // 与服务端版本相同
		Timestamp:  time.Now(),
		NewValue:   map[string]interface{}{"name": "conflict"},
		Checksum:   "xyz", // 不同的校验和
		Status:     SyncStatusPending,
	}

	service.mu.Lock()
	service.records[conflictRecord.RecordID] = conflictRecord
	service.mu.Unlock()

	// 手动触发冲突检查
	conflict := service.checkConflict(conflictRecord)

	// 验证冲突
	if conflict == nil {
		t.Fatal("Conflict should be detected")
	}

	if conflict.DataType != "memory" {
		t.Errorf("Expected data type 'memory', got '%s'", conflict.DataType)
	}

	if conflict.ServerVersion.Version != 1 {
		t.Errorf("Expected server version 1, got %d", conflict.ServerVersion.Version)
	}
}

func TestDataSyncServiceResolveConflict(t *testing.T) {
	config := DefaultDataSyncConfig()
	config.ConflictStrategy = ConflictStrategyMerge
	service := NewDataSyncService(config)

	// 创建初始数据
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"name": "original"})
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	// 创建冲突记录
	conflict := &ConflictRecord{
		ConflictID: "conflict-001",
		IdentityID: "identity-001",
		DataType:   "memory",
		DataKey:    "key-001",
		ServerVersion: &DataVersion{
			IdentityID: "identity-001",
			DataType:   "memory",
			DataKey:    "key-001",
			Version:    1,
			Checksum:   "abc",
		},
		ClientRecords: []*SyncRecord{
			{
				RecordID:   "record-client",
				IdentityID: "identity-001",
				AgentID:    "agent-002",
				DataType:   "memory",
				DataKey:    "key-001",
				NewValue:   map[string]interface{}{"name": "client-value"},
				Timestamp:  time.Now(),
			},
		},
		Strategy:  ConflictStrategyMerge,
		Status:    SyncStatusConflict,
		CreatedAt: time.Now(),
	}

	service.mu.Lock()
	service.conflicts[conflict.ConflictID] = conflict
	service.mu.Unlock()

	// 解决冲突
	resolvedValue := map[string]interface{}{"name": "resolved"}
	err := service.ResolveConflict(conflict.ConflictID, ConflictStrategyMerge, resolvedValue)
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}

	// 验证冲突状态
	updated := service.conflicts[conflict.ConflictID]
	if updated.Status != SyncStatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", updated.Status)
	}
	if updated.ResolvedValue["name"] != "resolved" {
		t.Error("Resolved value should be set")
	}
}

func TestDataSyncServiceAutoMerge(t *testing.T) {
	config := DefaultDataSyncConfig()
	config.ConflictStrategy = ConflictStrategyClient
	service := NewDataSyncService(config)

	conflict := &ConflictRecord{
		ConflictID: "conflict-001",
		IdentityID: "identity-001",
		DataType:   "memory",
		DataKey:    "key-001",
		ServerVersion: &DataVersion{
			Version: 1,
		},
		ClientRecords: []*SyncRecord{
			{
				NewValue: map[string]interface{}{"name": "client-value"},
			},
		},
		Strategy: ConflictStrategyClient,
	}

	merged := service.autoMerge(conflict)

	if merged == nil {
		t.Fatal("Auto merge should return a value")
	}

	if merged["name"] != "client-value" {
		t.Errorf("Expected 'client-value', got '%v'", merged["name"])
	}
}

func TestDataSyncServiceDeepMerge(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	conflict := &ConflictRecord{
		ConflictID: "conflict-001",
		IdentityID: "identity-001",
		DataType:   "memory",
		DataKey:    "key-001",
		ServerVersion: &DataVersion{
			Version: 1,
		},
		ClientRecords: []*SyncRecord{
			{
				NewValue: map[string]interface{}{
					"name":  "client",
					"value": 100,
				},
			},
		},
		Strategy: ConflictStrategyMerge,
	}

	merged := service.deepMerge(conflict)

	if merged == nil {
		t.Fatal("Deep merge should return a value")
	}

	if merged["merged"] != true {
		t.Error("Merged flag should be set")
	}
}

func TestDataSyncServiceRetry(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建失败的记录
	record := service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"name": "test"})

	// 标记为失败
	service.mu.Lock()
	record.Status = SyncStatusFailed
	record.ErrorMessage = "test error"
	service.mu.Unlock()

	// 清空待处理队列
	service.mu.Lock()
	service.pendingRecords["identity-001"] = nil
	service.mu.Unlock()

	// 重试
	retried := service.RetryFailedRecords("identity-001")
	if retried != 1 {
		t.Errorf("Expected 1 retried record, got %d", retried)
	}

	// 验证记录状态
	service.mu.RLock()
	updated := service.records[record.RecordID]
	service.mu.RUnlock()

	if updated.Status != SyncStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", updated.Status)
	}
	if updated.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", updated.RetryCount)
	}
}

func TestDataSyncServiceStats(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	// 创建一些记录
	for i := 0; i < 3; i++ {
		service.CreateSyncRecord("identity-001", "agent-001", "memory",
			"key-"+string(rune('0'+i)), SyncOpCreate, map[string]interface{}{"i": i})
	}

	// 创建批次并处理
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	// 获取统计
	stats := service.GetSyncStats("identity-001")

	if stats.PendingRecords != 0 {
		t.Errorf("Expected 0 pending records, got %d", stats.PendingRecords)
	}
	if stats.CompletedBatches != 1 {
		t.Errorf("Expected 1 completed batch, got %d", stats.CompletedBatches)
	}
}

func TestSyncRecordJSON(t *testing.T) {
	record := &SyncRecord{
		RecordID:   "record-001",
		IdentityID: "identity-001",
		AgentID:    "agent-001",
		DataType:   "memory",
		DataKey:    "key-001",
		Operation:  SyncOpCreate,
		Version:    1,
		Timestamp:  time.Now(),
		NewValue:   map[string]interface{}{"name": "test"},
		Checksum:   "abc",
		Status:     SyncStatusCompleted,
	}

	// 序列化
	data, err := record.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := SyncRecordFromJSON(data)
	if err != nil {
		t.Fatalf("SyncRecordFromJSON failed: %v", err)
	}

	if parsed.RecordID != record.RecordID {
		t.Error("RecordID mismatch after JSON roundtrip")
	}
	if parsed.DataType != record.DataType {
		t.Error("DataType mismatch after JSON roundtrip")
	}
}

func TestConflictRecordJSON(t *testing.T) {
	conflict := &ConflictRecord{
		ConflictID: "conflict-001",
		IdentityID: "identity-001",
		DataType:   "memory",
		DataKey:    "key-001",
		Strategy:   ConflictStrategyMerge,
		Status:     SyncStatusConflict,
		CreatedAt:  time.Now(),
	}

	// 序列化
	data, err := conflict.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := ConflictRecordFromJSON(data)
	if err != nil {
		t.Fatalf("ConflictRecordFromJSON failed: %v", err)
	}

	if parsed.ConflictID != conflict.ConflictID {
		t.Error("ConflictID mismatch after JSON roundtrip")
	}
	if parsed.Strategy != conflict.Strategy {
		t.Error("Strategy mismatch after JSON roundtrip")
	}
}

func TestDataSyncServiceListener(t *testing.T) {
	service := NewDataSyncService(DefaultDataSyncConfig())

	started := false
	completed := false
	recordSynced := false

	listener := &TestSyncListener{
		onStarted: func(b *SyncBatch) {
			started = true
		},
		onCompleted: func(b *SyncBatch) {
			completed = true
		},
		onRecordSynced: func(r *SyncRecord) {
			recordSynced = true
		},
	}

	service.AddListener(listener)

	// 创建并处理批次
	service.CreateSyncRecord("identity-001", "agent-001", "memory", "key-001",
		SyncOpCreate, map[string]interface{}{"name": "test"})
	batch := service.CreateSyncBatch("identity-001", "agent-001")
	service.ProcessSyncBatch(batch.BatchID)

	time.Sleep(50 * time.Millisecond)

	if !started {
		t.Error("Listener should have been notified on sync start")
	}
	if !completed {
		t.Error("Listener should have been notified on sync complete")
	}
	if !recordSynced {
		t.Error("Listener should have been notified on record synced")
	}
}

// 测试监听器
type TestSyncListener struct {
	onStarted         func(b *SyncBatch)
	onCompleted       func(b *SyncBatch)
	onFailed          func(b *SyncBatch, err error)
	onConflict        func(c *ConflictRecord)
	onConflictResolved func(c *ConflictRecord)
	onRecordSynced    func(r *SyncRecord)
}

func (l *TestSyncListener) OnSyncStarted(b *SyncBatch) {
	if l.onStarted != nil {
		l.onStarted(b)
	}
}

func (l *TestSyncListener) OnSyncCompleted(b *SyncBatch) {
	if l.onCompleted != nil {
		l.onCompleted(b)
	}
}

func (l *TestSyncListener) OnSyncFailed(b *SyncBatch, err error) {
	if l.onFailed != nil {
		l.onFailed(b, err)
	}
}

func (l *TestSyncListener) OnConflictDetected(c *ConflictRecord) {
	if l.onConflict != nil {
		l.onConflict(c)
	}
}

func (l *TestSyncListener) OnConflictResolved(c *ConflictRecord) {
	if l.onConflictResolved != nil {
		l.onConflictResolved(c)
	}
}

func (l *TestSyncListener) OnRecordSynced(r *SyncRecord) {
	if l.onRecordSynced != nil {
		l.onRecordSynced(r)
	}
}