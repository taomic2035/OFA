package ofa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineLevel defines offline capability level
type OfflineLevel int

const (
	OfflineLevelL1 OfflineLevel = iota // Complete offline
	OfflineLevelL2                     // LAN collaboration
	OfflineLevelL3                     // Weak network sync
	OfflineLevelL4                     // Online mode
)

// LocalTaskStatus defines local task status
type LocalTaskStatus int

const (
	TaskStatusPending LocalTaskStatus = iota
	TaskStatusRunning
	TaskStatusSuccess
	TaskStatusFailed
	TaskStatusCancelled
	TaskStatusRetrying
)

// LocalTask represents a local task
type LocalTask struct {
	TaskID      string          `json:"taskId"`
	SkillID     string          `json:"skillId"`
	Operation   string          `json:"operation"`
	Input       json.RawMessage `json:"input"`
	Result      json.RawMessage `json:"result,omitempty"`
	Status      LocalTaskStatus `json:"status"`
	Error       string          `json:"error,omitempty"`
	RetryCount  int             `json:"retryCount"`
	MaxRetries  int             `json:"maxRetries"`
	CreatedAt   int64           `json:"createdAt"`
	UpdatedAt   int64           `json:"updatedAt"`
	CompletedAt int64           `json:"completedAt,omitempty"`
	SyncPending bool            `json:"syncPending"`
}

// CacheEntry represents cache entry
type CacheEntry struct {
	Key         string          `json:"key"`
	Data        json.RawMessage `json:"data"`
	CreatedAt   int64           `json:"createdAt"`
	ExpiresAt   int64           `json:"expiresAt"`
	SyncPending bool            `json:"syncPending"`
	Source      string          `json:"source"`
}

// LocalScheduler manages local task execution
type LocalScheduler struct {
	maxConcurrent int
	handlers      map[string]SkillHandler
	pendingQueue  []*LocalTask
	activeTasks   map[string]*LocalTask
	completedTasks map[string]*LocalTask
	onComplete    func(*LocalTask)

	mu     sync.Mutex
	cond   *sync.Cond
	running bool
	done   chan struct{}

	stats SchedulerStats
}

// SchedulerStats holds scheduler statistics
type SchedulerStats struct {
	TotalTasks   uint64
	SuccessTasks uint64
	FailedTasks  uint64
	RetryTasks   uint64
}

// NewLocalScheduler creates a new local scheduler
func NewLocalScheduler(maxConcurrent int) *LocalScheduler {
	s := &LocalScheduler{
		maxConcurrent:  maxConcurrent,
		handlers:       make(map[string]SkillHandler),
		activeTasks:    make(map[string]*LocalTask),
		completedTasks: make(map[string]*LocalTask),
		done:           make(chan struct{}),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// RegisterHandler registers a task handler
func (s *LocalScheduler) RegisterHandler(skillID string, handler SkillHandler) {
	s.mu.Lock()
	s.handlers[skillID] = handler
	s.mu.Unlock()
}

// Submit submits a new task
func (s *LocalScheduler) Submit(skillID, operation string, input json.RawMessage, maxRetries int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &LocalTask{
		TaskID:      generateID(),
		SkillID:     skillID,
		Operation:   operation,
		Input:       input,
		Status:      TaskStatusPending,
		MaxRetries:  maxRetries,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		SyncPending: true,
	}

	s.pendingQueue = append(s.pendingQueue, task)
	s.stats.TotalTasks++
	s.cond.Signal()

	return task.TaskID
}

// Cancel cancels a task
func (s *LocalScheduler) Cancel(taskID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check pending queue
	for i, task := range s.pendingQueue {
		if task.TaskID == taskID {
			task.Status = TaskStatusCancelled
			task.UpdatedAt = time.Now().UnixMilli()
			s.completedTasks[taskID] = task
			s.pendingQueue = append(s.pendingQueue[:i], s.pendingQueue[i+1:]...)
			return true
		}
	}

	// Check active tasks
	if task, ok := s.activeTasks[taskID]; ok {
		task.Status = TaskStatusCancelled
		return true
	}

	return false
}

// GetTask gets task by ID
func (s *LocalScheduler) GetTask(taskID string) (*LocalTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.activeTasks[taskID]; ok {
		return task, nil
	}

	if task, ok := s.completedTasks[taskID]; ok {
		return task, nil
	}

	return nil, fmt.Errorf("task not found: %s", taskID)
}

// GetSyncPendingTasks returns tasks pending sync
func (s *LocalScheduler) GetSyncPendingTasks() []*LocalTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	var tasks []*LocalTask
	for _, task := range s.completedTasks {
		if task.SyncPending {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// OnComplete sets completion callback
func (s *LocalScheduler) OnComplete(callback func(*LocalTask)) {
	s.mu.Lock()
	s.onComplete = callback
	s.mu.Unlock()
}

// Start starts the scheduler
func (s *LocalScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go s.processLoop()
}

// Stop stops the scheduler
func (s *LocalScheduler) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	s.cond.Broadcast()
	close(s.done)
}

// Stats returns scheduler statistics
func (s *LocalScheduler) Stats() SchedulerStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

func (s *LocalScheduler) processLoop() {
	for {
		s.mu.Lock()

		// Wait for tasks or stop signal
		for len(s.pendingQueue) == 0 && s.running {
			s.cond.Wait()
		}

		if !s.running {
			s.mu.Unlock()
			return
		}

		// Check concurrency limit
		if len(s.activeTasks) >= s.maxConcurrent {
			s.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Get next task
		task := s.pendingQueue[0]
		s.pendingQueue = s.pendingQueue[1:]
		s.activeTasks[task.TaskID] = task
		s.mu.Unlock()

		s.executeTask(task)
	}
}

func (s *LocalScheduler) executeTask(task *LocalTask) {
	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now().UnixMilli()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	handler, ok := s.handlers[task.SkillID]
	if !ok {
		s.completeTask(task, nil, fmt.Errorf("no handler for skill: %s", task.SkillID))
		return
	}

	output, err := handler(ctx, task.Input)
	s.completeTask(task, output, err)
}

func (s *LocalScheduler) completeTask(task *LocalTask, output []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		task.Error = err.Error()
		task.UpdatedAt = time.Now().UnixMilli()

		if task.RetryCount < task.MaxRetries {
			task.Status = TaskStatusRetrying
			task.RetryCount++
			s.stats.RetryTasks++
			s.activeTasks[task.TaskID] = task // Keep active for retry
			go func() {
				time.Sleep(time.Second)
				s.mu.Lock()
				delete(s.activeTasks, task.TaskID)
				s.pendingQueue = append(s.pendingQueue, task)
				s.cond.Signal()
				s.mu.Unlock()
			}()
			return
		}

		task.Status = TaskStatusFailed
		task.CompletedAt = time.Now().UnixMilli()
		s.stats.FailedTasks++
	} else {
		task.Result = output
		task.Status = TaskStatusSuccess
		task.CompletedAt = time.Now().UnixMilli()
		s.stats.SuccessTasks++
	}

	delete(s.activeTasks, task.TaskID)
	s.completedTasks[task.TaskID] = task

	if s.onComplete != nil {
		go s.onComplete(task)
	}

	s.cond.Signal()
}

// OfflineCache manages offline data cache
type OfflineCache struct {
	dbPath string
	cache  map[string]*CacheEntry
	mu     sync.Mutex
}

// NewOfflineCache creates a new offline cache
func NewOfflineCache(dbPath string) *OfflineCache {
	c := &OfflineCache{
		dbPath: dbPath,
		cache:  make(map[string]*CacheEntry),
	}
	c.load()
	return c
}

// Put stores data in cache
func (c *OfflineCache) Put(key string, data json.RawMessage, ttlSeconds int64, syncPending bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()
	entry := &CacheEntry{
		Key:         key,
		Data:        data,
		CreatedAt:   now,
		ExpiresAt:   ttlSeconds > 0 ? now + ttlSeconds*1000 : 0,
		SyncPending: syncPending,
		Source:      "local",
	}
	c.cache[key] = entry
}

// Get retrieves data from cache
func (c *OfflineCache) Get(key string) (json.RawMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	if entry.ExpiresAt > 0 && time.Now().UnixMilli() > entry.ExpiresAt {
		return nil, false
	}

	return entry.Data, true
}

// Remove removes data from cache
func (c *OfflineCache) Remove(key string) {
	c.mu.Lock()
	delete(c.cache, key)
	c.mu.Unlock()
}

// ClearExpired clears expired entries
func (c *OfflineCache) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()
	for key, entry := range c.cache {
		if entry.ExpiresAt > 0 && now > entry.ExpiresAt {
			delete(c.cache, key)
		}
	}
}

// GetSyncPending returns entries pending sync
func (c *OfflineCache) GetSyncPending() []*CacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	var entries []*CacheEntry
	now := time.Now().UnixMilli()
	for _, entry := range c.cache {
		if entry.SyncPending && (entry.ExpiresAt == 0 || now <= entry.ExpiresAt) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// MarkSynced marks entry as synced
func (c *OfflineCache) MarkSynced(key string) {
	c.mu.Lock()
	if entry, ok := c.cache[key]; ok {
		entry.SyncPending = false
	}
	c.mu.Unlock()
}

// Persist saves cache to disk
func (c *OfflineCache) Persist() {
	if c.dbPath == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(c.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create cache directory: %v", err)
		return
	}

	// Serialize cache
	var entries []*CacheEntry
	for _, entry := range c.cache {
		if entry.ExpiresAt == 0 || time.Now().UnixMilli() <= entry.ExpiresAt {
			entries = append(entries, entry)
		}
	}

	data, err := json.Marshal(entries)
	if err != nil {
		log.Printf("Failed to marshal cache: %v", err)
		return
	}

	if err := os.WriteFile(c.dbPath, data, 0644); err != nil {
		log.Printf("Failed to write cache: %v", err)
	}
}

func (c *OfflineCache) load() {
	if c.dbPath == "" {
		return
	}

	data, err := os.ReadFile(c.dbPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read cache: %v", err)
		}
		return
	}

	var entries []*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Printf("Failed to unmarshal cache: %v", err)
		return
	}

	now := time.Now().UnixMilli()
	for _, entry := range entries {
		if entry.ExpiresAt == 0 || now <= entry.ExpiresAt {
			c.cache[entry.Key] = entry
		}
	}
}

// OfflineManager manages offline operations
type OfflineManager struct {
	level    OfflineLevel
	scheduler *LocalScheduler
	cache    *OfflineCache
	running  bool
	done     chan struct{}
	mu       sync.Mutex
}

// NewOfflineManager creates a new offline manager
func NewOfflineManager(cachePath string) *OfflineManager {
	return &OfflineManager{
		level:     OfflineLevelL4,
		scheduler: NewLocalScheduler(4),
		cache:     NewOfflineCache(cachePath),
		done:      make(chan struct{}),
	}
}

// SetLevel sets offline level
func (m *OfflineManager) SetLevel(level OfflineLevel) {
	m.mu.Lock()
	m.level = level
	m.mu.Unlock()
}

// Level returns current offline level
func (m *OfflineManager) Level() OfflineLevel {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.level
}

// IsOnline checks if online mode
func (m *OfflineManager) IsOnline() bool {
	return m.Level() == OfflineLevelL4
}

// Scheduler returns the scheduler
func (m *OfflineManager) Scheduler() *LocalScheduler {
	return m.scheduler
}

// Cache returns the cache
func (m *OfflineManager) Cache() *OfflineCache {
	return m.cache
}

// SubmitTask submits a local task
func (m *OfflineManager) SubmitTask(skillID, operation string, input json.RawMessage) string {
	taskID := m.scheduler.Submit(skillID, operation, input, 3)

	// Cache the input
	m.cache.Put("task:"+taskID, input, 0, true)

	return taskID
}

// RegisterOfflineSkill registers an offline skill handler
func (m *OfflineManager) RegisterOfflineSkill(skillID string, handler SkillHandler) {
	m.scheduler.RegisterHandler(skillID, handler)
}

// Start starts the offline manager
func (m *OfflineManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.scheduler.Start()
	go m.monitorLoop()
}

// Stop stops the offline manager
func (m *OfflineManager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.scheduler.Stop()
	close(m.done)
	m.cache.Persist()
}

// Sync syncs pending data to center
func (m *OfflineManager) Sync(ctx context.Context, agent *Agent) error {
	// Sync pending tasks
	tasks := m.scheduler.GetSyncPendingTasks()
	for _, task := range tasks {
		if agent != nil && agent.running {
			// Send to center
			log.Printf("Syncing task %s to center", task.TaskID)
			task.SyncPending = false
		}
	}

	// Sync pending cache entries
	entries := m.cache.GetSyncPending()
	for _, entry := range entries {
		log.Printf("Syncing cache key %s to center", entry.Key)
		m.cache.MarkSynced(entry.Key)
	}

	return nil
}

func (m *OfflineManager) monitorLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.autoSwitch()
			if m.IsOnline() {
				// Would sync to center here
			}
		}
	}
}

func (m *OfflineManager) autoSwitch() {
	// Auto-detect network and switch mode
	// In production: check connectivity to center
}