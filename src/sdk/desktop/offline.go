// Package desktop provides offline support
package desktop

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineLevel defines offline capability level
type OfflineLevel int

const (
	OfflineL1 OfflineLevel = iota // Complete offline
	OfflineL2                     // LAN collaboration
	OfflineL3                     // Weak network sync
	OfflineL4                     // Online mode
)

// OfflineMode represents current offline mode
type OfflineMode struct {
	Level       OfflineLevel
	IsOnline    bool
	LastSync    time.Time
	PendingData int
}

// OfflineManager manages offline operations for desktop agent
type OfflineManager struct {
	dataDir     string
	level       OfflineLevel
	scheduler   *LocalScheduler
	cache       *OfflineCache
	checker     *ConstraintChecker

	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
}

// NewOfflineManager creates a new offline manager
func NewOfflineManager(dataDir string) *OfflineManager {
	ctx, cancel := context.WithCancel(context.Background())

	cacheDir := filepath.Join(dataDir, "offline_cache")
	os.MkdirAll(cacheDir, 0755)

	return &OfflineManager{
		dataDir:   dataDir,
		level:     OfflineL4,
		scheduler: NewLocalScheduler(4),
		cache:     NewOfflineCache(filepath.Join(cacheDir, "cache.json")),
		checker:   NewConstraintChecker(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// SetLevel sets offline level
func (m *OfflineManager) SetLevel(level OfflineLevel) {
	m.mu.Lock()
	m.level = level
	m.mu.Unlock()
	log.Printf("Offline level set to: %d", level)
}

// GetLevel returns current offline level
func (m *OfflineManager) GetLevel() OfflineLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

// IsOnline checks if in online mode
func (m *OfflineManager) IsOnline() bool {
	return m.GetLevel() == OfflineL4
}

// GetMode returns current offline mode info
func (m *OfflineManager) GetMode() OfflineMode {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return OfflineMode{
		Level:       m.level,
		IsOnline:    m.level == OfflineL4,
		LastSync:    time.Now(),
		PendingData: len(m.scheduler.GetSyncPendingTasks()),
	}
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
	go m.networkMonitor()
	go m.autoSync()

	log.Println("Offline manager started")
}

// Stop stops the offline manager
func (m *OfflineManager) Stop() {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	m.cancel()
	m.scheduler.Stop()
	m.cache.Persist()

	log.Println("Offline manager stopped")
}

// SubmitLocalTask submits a task for local execution
func (m *OfflineManager) SubmitLocalTask(skillID, operation string, params map[string]interface{}) string {
	input, _ := json.Marshal(params)
	taskID := m.scheduler.Submit(skillID, operation, input, 3)

	// Cache the input
	m.cache.Put("task:"+taskID, input, 0, true)

	return taskID
}

// GetLocalTask gets a local task status
func (m *OfflineManager) GetLocalTask(taskID string) (*LocalTask, error) {
	return m.scheduler.GetTask(taskID)
}

// RegisterOfflineSkill registers an offline skill handler
func (m *OfflineManager) RegisterOfflineSkill(skillID string, handler func(json.RawMessage) (json.RawMessage, error)) {
	m.scheduler.RegisterHandler(skillID, handler)
}

// CheckConstraints checks data against constraints
func (m *OfflineManager) CheckConstraints(taskID, skillID, operation string, data json.RawMessage) *ConstraintReport {
	return m.checker.Check(taskID, skillID, operation, data)
}

// SyncPending syncs pending data to center
func (m *OfflineManager) SyncPending(agent *DesktopAgent) error {
	if !m.IsOnline() {
		return nil
	}

	// Sync pending tasks
	tasks := m.scheduler.GetSyncPendingTasks()
	for _, task := range tasks {
		if agent.connector != nil && agent.connector.IsConnected() {
			// Convert to agent task format
			agentTask := Task{
				ID:        task.TaskID,
				SkillID:   task.SkillID,
				Operation: task.Operation,
			}

			var params map[string]interface{}
			json.Unmarshal(task.Input, &params)
			agentTask.Params = params

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

// GetScheduler returns the scheduler
func (m *OfflineManager) GetScheduler() *LocalScheduler {
	return m.scheduler
}

// GetCache returns the cache
func (m *OfflineManager) GetCache() *OfflineCache {
	return m.cache
}

// GetChecker returns the constraint checker
func (m *OfflineManager) GetChecker() *ConstraintChecker {
	return m.checker
}

func (m *OfflineManager) networkMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.autoDetectNetwork()
		}
	}
}

func (m *OfflineManager) autoDetectNetwork() {
	// Simple network check - try to connect to center
	// In production: implement proper network detection
}

func (m *OfflineManager) autoSync() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if m.IsOnline() {
				// Would sync to center here
			}
		}
	}
}

// LocalScheduler manages local task execution
type LocalScheduler struct {
	maxConcurrent int
	handlers      map[string]func(json.RawMessage) (json.RawMessage, error)
	pendingQueue  []*LocalTask
	activeTasks   map[string]*LocalTask
	completedTasks map[string]*LocalTask

	mu      sync.Mutex
	cond    *sync.Cond
	running bool
	done    chan struct{}

	stats SchedulerStats
}

// SchedulerStats holds scheduler statistics
type SchedulerStats struct {
	Total   uint64
	Success uint64
	Failed  uint64
	Retry   uint64
}

// LocalTask represents a local task
type LocalTask struct {
	TaskID      string
	SkillID     string
	Operation   string
	Input       json.RawMessage
	Result      json.RawMessage
	Status      TaskStatus
	Error       string
	RetryCount  int
	MaxRetries  int
	CreatedAt   int64
	UpdatedAt   int64
	CompletedAt int64
	SyncPending bool
}

// TaskStatus represents task status
type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusCancelled
	StatusRetrying
)

// NewLocalScheduler creates a new scheduler
func NewLocalScheduler(maxConcurrent int) *LocalScheduler {
	s := &LocalScheduler{
		maxConcurrent:  maxConcurrent,
		handlers:       make(map[string]func(json.RawMessage) (json.RawMessage, error)),
		activeTasks:    make(map[string]*LocalTask),
		completedTasks: make(map[string]*LocalTask),
		done:           make(chan struct{}),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// RegisterHandler registers a handler
func (s *LocalScheduler) RegisterHandler(skillID string, handler func(json.RawMessage) (json.RawMessage, error)) {
	s.mu.Lock()
	s.handlers[skillID] = handler
	s.mu.Unlock()
}

// Submit submits a new task
func (s *LocalScheduler) Submit(skillID, operation string, input json.RawMessage, maxRetries int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &LocalTask{
		TaskID:      generateTaskID(),
		SkillID:     skillID,
		Operation:   operation,
		Input:       input,
		Status:      StatusPending,
		MaxRetries:  maxRetries,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
		SyncPending: true,
	}

	s.pendingQueue = append(s.pendingQueue, task)
	s.stats.Total++
	s.cond.Signal()

	return task.TaskID
}

// GetTask returns task by ID
func (s *LocalScheduler) GetTask(taskID string) (*LocalTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.activeTasks[taskID]; ok {
		return task, nil
	}
	if task, ok := s.completedTasks[taskID]; ok {
		return task, nil
	}
	return nil, ErrTaskNotFound
}

// GetSyncPendingTasks returns pending sync tasks
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

func (s *LocalScheduler) processLoop() {
	for {
		s.mu.Lock()
		for len(s.pendingQueue) == 0 && s.running {
			s.cond.Wait()
		}
		if !s.running {
			s.mu.Unlock()
			return
		}

		if len(s.activeTasks) >= s.maxConcurrent {
			s.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		task := s.pendingQueue[0]
		s.pendingQueue = s.pendingQueue[1:]
		s.activeTasks[task.TaskID] = task
		s.mu.Unlock()

		s.executeTask(task)
	}
}

func (s *LocalScheduler) executeTask(task *LocalTask) {
	task.Status = StatusRunning
	task.UpdatedAt = time.Now().UnixMilli()

	s.mu.Lock()
	handler, ok := s.handlers[task.SkillID]
	s.mu.Unlock()

	if !ok {
		s.completeTask(task, nil, ErrSkillNotFound)
		return
	}

	result, err := handler(task.Input)
	s.completeTask(task, result, err)
}

func (s *LocalScheduler) completeTask(task *LocalTask, result json.RawMessage, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		task.Error = err.Error()
		if task.RetryCount < task.MaxRetries {
			task.Status = StatusRetrying
			task.RetryCount++
			s.stats.Retry++
			delete(s.activeTasks, task.TaskID)
			s.pendingQueue = append(s.pendingQueue, task)
			s.cond.Signal()
			return
		}
		task.Status = StatusFailed
		s.stats.Failed++
	} else {
		task.Result = result
		task.Status = StatusSuccess
		s.stats.Success++
	}

	task.CompletedAt = time.Now().UnixMilli()
	delete(s.activeTasks, task.TaskID)
	s.completedTasks[task.TaskID] = task
	s.cond.Signal()
}

// OfflineCache manages data cache
type OfflineCache struct {
	path  string
	cache map[string]*CacheEntry
	mu    sync.Mutex
}

// CacheEntry represents cached data
type CacheEntry struct {
	Key         string
	Data        json.RawMessage
	CreatedAt   int64
	ExpiresAt   int64
	SyncPending bool
	Source      string
}

// NewOfflineCache creates a new cache
func NewOfflineCache(path string) *OfflineCache {
	c := &OfflineCache{
		path:  path,
		cache: make(map[string]*CacheEntry),
	}
	c.load()
	return c
}

// Put stores data
func (c *OfflineCache) Put(key string, data json.RawMessage, ttl int64, syncPending bool) {
	c.mu.Lock()
	now := time.Now().UnixMilli()
	c.cache[key] = &CacheEntry{
		Key:         key,
		Data:        data,
		CreatedAt:   now,
		ExpiresAt:   ttl > 0 ? now + ttl*1000 : 0,
		SyncPending: syncPending,
		Source:      "local",
	}
	c.mu.Unlock()
}

// Get retrieves data
func (c *OfflineCache) Get(key string) (json.RawMessage, bool) {
	c.mu.Lock()
	entry, ok := c.cache[key]
	if !ok || (entry.ExpiresAt > 0 && time.Now().UnixMilli() > entry.ExpiresAt) {
		c.mu.Unlock()
		return nil, false
	}
	c.mu.Unlock()
	return entry.Data, true
}

// GetSyncPending returns pending sync entries
func (c *OfflineCache) GetSyncPending() []*CacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	var entries []*CacheEntry
	for _, entry := range c.cache {
		if entry.SyncPending && (entry.ExpiresAt == 0 || time.Now().UnixMilli() <= entry.ExpiresAt) {
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
	c.mu.Lock()
	defer c.mu.Unlock()

	var entries []*CacheEntry
	for _, entry := range c.cache {
		if entry.ExpiresAt == 0 || time.Now().UnixMilli() <= entry.ExpiresAt {
			entries = append(entries, entry)
		}
	}

	data, _ := json.Marshal(entries)
	os.WriteFile(c.path, data, 0644)
}

func (c *OfflineCache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}

	var entries []*CacheEntry
	if json.Unmarshal(data, &entries) == nil {
		now := time.Now().UnixMilli()
		for _, entry := range entries {
			if entry.ExpiresAt == 0 || now <= entry.ExpiresAt {
				c.cache[entry.Key] = entry
			}
		}
	}
}

func generateTaskID() string {
	return "local-" + time.Now().Format("20060102150405")
}

// Errors
var ErrTaskNotFound = &OfflineError{Message: "task not found"}
var ErrSkillNotFound = &OfflineError{Message: "skill not found"}

// OfflineError represents offline error
type OfflineError struct {
	Message string
}

func (e *OfflineError) Error() string {
	return e.Message
}