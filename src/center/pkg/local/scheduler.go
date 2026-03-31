// Package local - 本地调度器
// 支持离线模式下的任务调度和执行
package local

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// OfflineLevel 离线能力等级
type OfflineLevel int

const (
	OfflineNone OfflineLevel = iota // 不支持离线
	OfflineL1                        // 完全离线 (本地执行)
	OfflineL2                        // 局域网协作
	OfflineL3                        // 弱网同步
	OfflineL4                        // 在线模式
)

// LocalScheduler 本地调度器
type LocalScheduler struct {
	level         OfflineLevel
	skills        map[string]*LocalSkill
	taskQueue     chan *LocalTask
	pendingTasks  map[string]*LocalTask
	completedTask map[string]*LocalTask
	cache         *OfflineCache
	mu            sync.RWMutex
	running       atomic.Bool
	workers       int
}

// LocalSkill 本地技能
type LocalSkill struct {
	ID             string
	Name           string
	Category       string
	OfflineCapable bool
	Handler        SkillHandler
}

// SkillHandler 技能处理函数
type SkillHandler func(ctx context.Context, input []byte) ([]byte, error)

// LocalTask 本地任务
type LocalTask struct {
	ID              string
	SkillID         string
	Input           []byte
	Output          []byte
	Status          TaskStatus
	Error           string
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time
	OfflineCapable  bool
	SyncPending     bool
	SyncWhenOnline  bool
}

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskCancelled
)

// OfflineCache 离线缓存
type OfflineCache struct {
	data      map[string][]byte
	timestamp map[string]time.Time
	mu        sync.RWMutex
	maxSize   int
}

// NewLocalScheduler 创建本地调度器
func NewLocalScheduler(level OfflineLevel, workers int) *LocalScheduler {
	return &LocalScheduler{
		level:         level,
		skills:        make(map[string]*LocalSkill),
		taskQueue:     make(chan *LocalTask, 1000),
		pendingTasks:  make(map[string]*LocalTask),
		completedTask: make(map[string]*LocalTask),
		cache:         NewOfflineCache(100 * 1024 * 1024), // 100MB
		workers:       workers,
	}
}

// Start 启动调度器
func (s *LocalScheduler) Start(ctx context.Context) error {
	if s.running.Swap(true) {
		return fmt.Errorf("scheduler already running")
	}

	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		go s.worker(ctx, i)
	}

	log.Printf("Local scheduler started with %d workers, level: %d", s.workers, s.level)
	return nil
}

// Stop 停止调度器
func (s *LocalScheduler) Stop() {
	s.running.Store(false)
}

// RegisterSkill 注册技能
func (s *LocalScheduler) RegisterSkill(skill *LocalSkill) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.skills[skill.ID]; exists {
		return fmt.Errorf("skill %s already registered", skill.ID)
	}

	s.skills[skill.ID] = skill
	log.Printf("Registered local skill: %s (offline: %v)", skill.ID, skill.OfflineCapable)
	return nil
}

// Execute 执行任务
func (s *LocalScheduler) Execute(ctx context.Context, skillID string, input []byte) (*LocalTask, error) {
	// 检查技能是否存在
	s.mu.RLock()
	skill, exists := s.skills[skillID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("skill %s not found", skillID)
	}

	// 检查离线能力
	if s.level == OfflineNone && !skill.OfflineCapable {
		return nil, fmt.Errorf("skill %s not available in offline mode", skillID)
	}

	// 创建任务
	task := &LocalTask{
		ID:              generateTaskID(),
		SkillID:         skillID,
		Input:           input,
		Status:          TaskPending,
		CreatedAt:       time.Now(),
		OfflineCapable:  skill.OfflineCapable,
		SyncWhenOnline:  s.level < OfflineL4,
	}

	// 加入队列
	select {
	case s.taskQueue <- task:
		s.mu.Lock()
		s.pendingTasks[task.ID] = task
		s.mu.Unlock()
		return task, nil
	default:
		return nil, fmt.Errorf("task queue full")
	}
}

// ExecuteSync 同步执行任务
func (s *LocalScheduler) ExecuteSync(ctx context.Context, skillID string, input []byte) (*LocalTask, error) {
	task, err := s.Execute(ctx, skillID, input)
	if err != nil {
		return nil, err
	}

	// 等待完成
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			s.mu.RLock()
			completed, exists := s.completedTask[task.ID]
			s.mu.RUnlock()

			if exists {
				return completed, nil
			}
		}
	}
}

// GetTask 获取任务状态
func (s *LocalScheduler) GetTask(taskID string) (*LocalTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if task, exists := s.pendingTasks[taskID]; exists {
		return task, true
	}
	if task, exists := s.completedTask[taskID]; exists {
		return task, true
	}
	return nil, false
}

// GetPendingSyncTasks 获取待同步的任务
func (s *LocalScheduler) GetPendingSyncTasks() []*LocalTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*LocalTask
	for _, task := range s.completedTask {
		if task.SyncPending {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// MarkSynced 标记已同步
func (s *LocalScheduler) MarkSynced(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.completedTask[taskID]; exists {
		task.SyncPending = false
	}
}

// worker 工作协程
func (s *LocalScheduler) worker(ctx context.Context, id int) {
	log.Printf("Local worker %d started", id)

	for s.running.Load() {
		select {
		case <-ctx.Done():
			return
		case task := <-s.taskQueue:
			s.executeTask(ctx, task)
		}
	}
}

// executeTask 执行任务
func (s *LocalScheduler) executeTask(ctx context.Context, task *LocalTask) {
	s.mu.RLock()
	skill := s.skills[task.SkillID]
	s.mu.RUnlock()

	if skill == nil {
		task.Status = TaskFailed
		task.Error = fmt.Sprintf("skill %s not found", task.SkillID)
		s.moveToCompleted(task)
		return
	}

	// 更新状态
	task.Status = TaskRunning
	task.StartedAt = time.Now()

	// 执行
	output, err := skill.Handler(ctx, task.Input)
	task.CompletedAt = time.Now()

	if err != nil {
		task.Status = TaskFailed
		task.Error = err.Error()
	} else {
		task.Status = TaskCompleted
		task.Output = output
	}

	// 标记待同步
	if task.SyncWhenOnline {
		task.SyncPending = true
	}

	s.moveToCompleted(task)
}

// moveToCompleted 移动到完成队列
func (s *LocalScheduler) moveToCompleted(task *LocalTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pendingTasks, task.ID)
	s.completedTask[task.ID] = task

	// 限制完成队列大小
	if len(s.completedTask) > 10000 {
		// 删除最旧的
		var oldest *LocalTask
		for _, t := range s.completedTask {
			if oldest == nil || t.CompletedAt.Before(oldest.CompletedAt) {
				oldest = t
			}
		}
		if oldest != nil {
			delete(s.completedTask, oldest.ID)
		}
	}
}

// === 离线缓存 ===

// NewOfflineCache 创建离线缓存
func NewOfflineCache(maxSize int) *OfflineCache {
	return &OfflineCache{
		data:      make(map[string][]byte),
		timestamp: make(map[string]time.Time),
		maxSize:   maxSize,
	}
}

// Set 设置缓存
func (c *OfflineCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
	c.timestamp[key] = time.Now()
}

// Get 获取缓存
func (c *OfflineCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, exists := c.data[key]
	return data, exists
}

// Delete 删除缓存
func (c *OfflineCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	delete(c.timestamp, key)
}

// Clear 清空缓存
func (c *OfflineCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string][]byte)
	c.timestamp = make(map[string]time.Time)
}

// Size 获取缓存大小
func (c *OfflineCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := 0
	for _, v := range c.data {
		total += len(v)
	}
	return total
}

// === 辅助函数 ===

func generateTaskID() string {
	return fmt.Sprintf("local-%d", time.Now().UnixNano())
}

func (s TaskStatus) String() string {
	switch s {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func (s TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}