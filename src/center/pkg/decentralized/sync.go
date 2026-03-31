// Package decentralized - 数据同步
// Sprint 33: v9.0 去中心化增强
package decentralized

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SyncManager 同步管理器
type SyncManager struct {
	syncMode    SyncMode
	syncTasks   map[string]*SyncTask
	syncQueue   chan *SyncTask
	state       *SyncState
	conflictRes *ConflictResolver
	mu          sync.RWMutex
}

// SyncMode 同步模式
type SyncMode string

const (
	SyncModeImmediate SyncMode = "immediate" // 立即同步
	SyncModeBatch     SyncMode = "batch"     // 批量同步
	SyncModeScheduled SyncMode = "scheduled" // 定时同步
	SyncModeOnDemand  SyncMode = "on_demand" // 按需同步
)

// SyncTask 同步任务
type SyncTask struct {
	ID           string        `json:"id"`
	Type         SyncType      `json:"type"`
	SourceNode   string        `json:"source_node"`
	TargetNodes  []string      `json:"target_nodes"`
	DataID       string        `json:"data_id"`
	Status       SyncStatus    `json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	StartedAt    time.Time     `json:"started_at,omitempty"`
	CompletedAt  time.Time     `json:"completed_at,omitempty"`
	Progress     float64       `json:"progress"`
	Error        string        `json:"error,omitempty"`
	RetryCount   int           `json:"retry_count"`
	MaxRetries   int           `json:"max_retries"`
	Timeout      time.Duration `json:"timeout"`
}

// SyncType 同步类型
type SyncType string

const (
	SyncTypeFull      SyncType = "full"      // 全量同步
	SyncTypeIncrement SyncType = "increment" // 增量同步
	SyncTypeDelta     SyncType = "delta"     // 差异同步
	SyncTypeMerkl     SyncType = "merkle"    // Merkle树同步
)

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "pending"
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusComplete  SyncStatus = "complete"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusCancelled SyncStatus = "cancelled"
)

// SyncState 同步状态
type SyncState struct {
	LastSyncTime time.Time `json:"last_sync_time"`
	LastSyncID   string    `json:"last_sync_id"`
	Version      int64     `json:"version"`
	PendingSyncs int       `json:"pending_syncs"`
}

// NewSyncManager 创建同步管理器
func NewSyncManager() *SyncManager {
	sm := &SyncManager{
		syncMode:    SyncModeBatch,
		syncTasks:   make(map[string]*SyncTask),
		syncQueue:   make(chan *SyncTask, 100),
		state:       &SyncState{},
		conflictRes: NewConflictResolver(),
	}

	// 启动同步工作器
	go sm.syncWorker()

	return sm
}

// SetSyncMode 设置同步模式
func (sm *SyncManager) SetSyncMode(mode SyncMode) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.syncMode = mode
}

// StartSync 启动同步
func (sm *SyncManager) StartSync(ctx context.Context, nodeID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 创建同步任务
	task := &SyncTask{
		ID:          generateSyncID(),
		Type:        SyncTypeIncrement,
		SourceNode:  nodeID,
		TargetNodes: []string{"all"},
		Status:      SyncStatusPending,
		CreatedAt:   time.Now(),
		Progress:    0,
		MaxRetries:  3,
		Timeout:     30 * time.Second,
	}

	sm.syncTasks[task.ID] = task
	sm.syncQueue <- task

	sm.state.LastSyncTime = time.Now()
	sm.state.LastSyncID = task.ID

	return nil
}

// CreateSyncTask 创建同步任务
func (sm *SyncManager) CreateSyncTask(ctx context.Context, syncType SyncType, sourceNode string, targetNodes []string, dataID string) (*SyncTask, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task := &SyncTask{
		ID:          generateSyncID(),
		Type:        syncType,
		SourceNode:  sourceNode,
		TargetNodes: targetNodes,
		DataID:      dataID,
		Status:      SyncStatusPending,
		CreatedAt:   time.Now(),
		Progress:    0,
		MaxRetries:  3,
		Timeout:     30 * time.Second,
	}

	sm.syncTasks[task.ID] = task
	sm.syncQueue <- task

	return task, nil
}

// syncWorker 同步工作器
func (sm *SyncManager) syncWorker() {
	for task := range sm.syncQueue {
		sm.executeSync(task)
	}
}

// executeSync 执行同步
func (sm *SyncManager) executeSync(task *SyncTask) {
	sm.mu.Lock()
	task.Status = SyncStatusRunning
	task.StartedAt = time.Now()
	sm.mu.Unlock()

	// 执行同步逻辑
	for i, targetNode := range task.TargetNodes {
		// 模拟同步
		time.Sleep(100 * time.Millisecond)

		sm.mu.Lock()
		task.Progress = float64(i+1) / float64(len(task.TargetNodes)) * 100
		sm.mu.Unlock()
	}

	// 完成
	sm.mu.Lock()
	task.Status = SyncStatusComplete
	task.CompletedAt = time.Now()
	task.Progress = 100
	sm.state.Version++
	sm.mu.Unlock()
}

// CancelSync 取消同步
func (sm *SyncManager) CancelSync(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.syncTasks[taskID]
	if !ok {
		return fmt.Errorf("同步任务不存在: %s", taskID)
	}

	if task.Status == SyncStatusRunning {
		return fmt.Errorf("同步任务正在运行，无法取消")
	}

	task.Status = SyncStatusCancelled
	return nil
}

// RetrySync 重试同步
func (sm *SyncManager) RetrySync(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.syncTasks[taskID]
	if !ok {
		return fmt.Errorf("同步任务不存在: %s", taskID)
	}

	if task.RetryCount >= task.MaxRetries {
		return fmt.Errorf("重试次数已达上限")
	}

	task.RetryCount++
	task.Status = SyncStatusPending
	task.Error = ""

	// 重新加入队列
	select {
	case sm.syncQueue <- task:
	default:
		return fmt.Errorf("同步队列已满")
	}

	return nil
}

// GetSyncTask 获取同步任务
func (sm *SyncManager) GetSyncTask(taskID string) (*SyncTask, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	task, ok := sm.syncTasks[taskID]
	if !ok {
		return nil, fmt.Errorf("同步任务不存在: %s", taskID)
	}
	return task, nil
}

// ListSyncTasks 列出同步任务
func (sm *SyncManager) ListSyncTasks(status SyncStatus) []*SyncTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	list := make([]*SyncTask, 0)
	for _, task := range sm.syncTasks {
		if status == "" || task.Status == status {
			list = append(list, task)
		}
	}
	return list
}

// ResolveConflict 解决冲突
func (sm *SyncManager) ResolveConflict(conflictType ConflictType, data map[string]interface{}) (map[string]interface{}, error) {
	return sm.conflictRes.Resolve(conflictType, []string{}, data)
}

// GetState 获取同步状态
func (sm *SyncManager) GetState() *SyncState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state := *sm.state

	// 计算待处理同步数
	pending := 0
	for _, task := range sm.syncTasks {
		if task.Status == SyncStatusPending || task.Status == SyncStatusRunning {
			pending++
		}
	}
	state.PendingSyncs = pending

	return &state
}

// generateSyncID 生成同步ID
func generateSyncID() string {
	return fmt.Sprintf("sync-%d", time.Now().UnixNano())
}

// SyncStats 同步统计
type SyncStats struct {
	TotalTasks      int           `json:"total_tasks"`
	CompletedTasks  int           `json:"completed_tasks"`
	FailedTasks     int           `json:"failed_tasks"`
	PendingTasks    int           `json:"pending_tasks"`
	AvgDuration     time.Duration `json:"avg_duration"`
	TotalDataSynced int64         `json:"total_data_synced"`
	SuccessRate     float64       `json:"success_rate"`
}

// GetStats 获取统计
func (sm *SyncManager) GetStats() *SyncStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := &SyncStats{
		TotalTasks: len(sm.syncTasks),
	}

	if len(sm.syncTasks) == 0 {
		return stats
	}

	var totalDuration time.Duration
	successCount := 0

	for _, task := range sm.syncTasks {
		switch task.Status {
		case SyncStatusComplete:
			stats.CompletedTasks++
			successCount++
			if !task.CompletedAt.IsZero() && !task.StartedAt.IsZero() {
				totalDuration += task.CompletedAt.Sub(task.StartedAt)
			}
		case SyncStatusFailed:
			stats.FailedTasks++
		case SyncStatusPending, SyncStatusRunning:
			stats.PendingTasks++
		}
	}

	if stats.CompletedTasks > 0 {
		stats.AvgDuration = totalDuration / time.Duration(stats.CompletedTasks)
	}

	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalTasks)
	}

	return stats
}

// ScheduleSync 调度同步
func (sm *SyncManager) ScheduleSync(ctx context.Context, interval time.Duration, syncType SyncType) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sm.StartSync(ctx, "scheduler")
			case <-ctx.Done():
				return
			}
		}
	}()
}

// WaitForSync 等待同步完成
func (sm *SyncManager) WaitForSync(ctx context.Context, taskID string) error {
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			sm.mu.RLock()
			task, ok := sm.syncTasks[taskID]
			sm.mu.RUnlock()

			if !ok {
				return fmt.Errorf("同步任务不存在: %s", taskID)
			}

			switch task.Status {
			case SyncStatusComplete:
				return nil
			case SyncStatusFailed:
				return fmt.Errorf("同步失败: %s", task.Error)
			case SyncStatusCancelled:
				return fmt.Errorf("同步已取消")
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}