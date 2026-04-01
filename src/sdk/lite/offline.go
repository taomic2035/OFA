// Package lite - 离线支持(轻量级，适合手表/手环)
package lite

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfflineLevel 离线级别
type OfflineLevel int

const (
	OfflineL1 OfflineLevel = iota // 完全离线
	OfflineL2                     // 局域网
	OfflineL3                     // 弱网
	OfflineL4                     // 在线
)

// OfflineConfig 离线配置
type OfflineConfig struct {
	MaxLocalTasks  int   // 最大本地任务数(默认10，节省内存)
	MaxCacheSize   int   // 最大缓存条目数(默认50)
	MaxRetries     int   // 最大重试次数(默认2)
	AutoSync       bool  // 自动同步
	PowerSaveMode  bool  // 省电模式
}

// DefaultOfflineConfig 默认配置
func DefaultOfflineConfig() OfflineConfig {
	return OfflineConfig{
		MaxLocalTasks: 10,
		MaxCacheSize:  50,
		MaxRetries:    2,
		AutoSync:      true,
		PowerSaveMode: true,
	}
}

// LocalTask 本地任务(简化版)
type LocalTask struct {
	ID         string      `json:"id"`
	Skill      string      `json:"skill"`
	Action     string      `json:"action"`
	Params     interface{} `json:"params,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	Status     string      `json:"status"` // pending, running, done, failed
	Error      string      `json:"error,omitempty"`
	RetryCount int         `json:"retry_count"`
	CreatedAt  time.Time   `json:"created_at"`
	Synced     bool        `json:"synced"`
}

// CacheEntry 缓存条目(简化版)
type CacheEntry struct {
	Key       string      `json:"key"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at,omitempty"`
	Synced    bool        `json:"synced"`
}

// OfflineManager 离线管理器(轻量级)
type OfflineManager struct {
	level   OfflineLevel
	config  OfflineConfig
	agent   *LiteAgent

	// 本地任务队列
	tasks    map[string]*LocalTask
	taskList []string // 保持顺序

	// 数据缓存
	cache     map[string]*CacheEntry
	cacheList []string

	// 持久化路径
	dataPath string

	mu      sync.Mutex
	running bool
}

// NewOfflineManager 创建离线管理器
func NewOfflineManager(agent *LiteAgent, config OfflineConfig) *OfflineManager {
	return &OfflineManager{
		level:    OfflineL4,
		config:   config,
		agent:    agent,
		tasks:    make(map[string]*LocalTask),
		taskList: make([]string, 0, config.MaxLocalTasks),
		cache:    make(map[string]*CacheEntry),
		cacheList: make([]string, 0, config.MaxCacheSize),
	}
}

// SetDataPath 设置数据持久化路径
func (m *OfflineManager) SetDataPath(path string) {
	m.dataPath = path
	os.MkdirAll(path, 0755)
}

// SetLevel 设置离线级别
func (m *OfflineManager) SetLevel(level OfflineLevel) {
	m.mu.Lock()
	m.level = level
	m.mu.Unlock()
}

// GetLevel 获取离线级别
func (m *OfflineManager) GetLevel() OfflineLevel {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.level
}

// IsOnline 是否在线
func (m *OfflineManager) IsOnline() bool {
	return m.GetLevel() == OfflineL4
}

// SubmitTask 提交本地任务
func (m *OfflineManager) SubmitTask(skill, action string, params interface{}) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查队列大小
	if len(m.tasks) >= m.config.MaxLocalTasks {
		// 移除最旧的已完成任务
		m.evictOldTasks()
	}

	task := &LocalTask{
		ID:        generateLiteID(),
		Skill:     skill,
		Action:    action,
		Params:    params,
		Status:    "pending",
		CreatedAt: time.Now(),
		Synced:    false,
	}

	m.tasks[task.ID] = task
	m.taskList = append(m.taskList, task.ID)

	return task.ID
}

// GetTask 获取任务
func (m *OfflineManager) GetTask(id string) *LocalTask {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tasks[id]
}

// ExecutePending 执行待处理任务
func (m *OfflineManager) ExecutePending() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range m.taskList {
		task := m.tasks[id]
		if task == nil || task.Status != "pending" {
			continue
		}

		// 检查电池
		if m.agent != nil && m.config.PowerSaveMode {
			battery := m.agent.GetBatteryInfo()
			if battery != nil && battery.Level < 20 && !battery.Charging {
				continue // 低电量暂停执行
			}
		}

		// 执行任务
		task.Status = "running"
		if m.agent != nil {
			lt := &LiteTask{
				ID:     task.ID,
				Skill:  task.Skill,
				Action: task.Action,
				Params: task.Params,
			}
			result, err := m.agent.ExecuteTask(nil, lt)
			if err != nil {
				task.Status = "failed"
				task.Error = err.Error()
				task.RetryCount++
				if task.RetryCount <= m.config.MaxRetries {
					task.Status = "pending" // 允许重试
				}
			} else {
				task.Status = "done"
				task.Result = result
			}
		}
	}
}

// GetPendingTasks 获取待同步任务
func (m *OfflineManager) GetPendingTasks() []*LocalTask {
	m.mu.Lock()
	defer m.mu.Unlock()

	var tasks []*LocalTask
	for _, id := range m.taskList {
		if task := m.tasks[id]; task != nil && !task.Synced {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// MarkSynced 标记已同步
func (m *OfflineManager) MarkSynced(taskID string) {
	m.mu.Lock()
	if task := m.tasks[taskID]; task != nil {
		task.Synced = true
	}
	m.mu.Unlock()
}

// CachePut 缓存数据
func (m *OfflineManager) CachePut(key string, data interface{}, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查缓存大小
	if len(m.cache) >= m.config.MaxCacheSize {
		m.evictOldCache()
	}

	entry := &CacheEntry{
		Key:       key,
		Data:      data,
		CreatedAt: time.Now(),
		Synced:    false,
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	m.cache[key] = entry
	m.cacheList = append(m.cacheList, key)
}

// CacheGet 获取缓存
func (m *OfflineManager) CacheGet(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.cache[key]
	if !ok {
		return nil
	}

	// 检查过期
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		delete(m.cache, key)
		return nil
	}

	return entry.Data
}

// GetPendingCache 获取待同步缓存
func (m *OfflineManager) GetPendingCache() []*CacheEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var entries []*CacheEntry
	now := time.Now()

	for _, key := range m.cacheList {
		if entry := m.cache[key]; entry != nil && !entry.Synced {
			if entry.ExpiresAt.IsZero() || now.Before(entry.ExpiresAt) {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

// MarkCacheSynced 标记缓存已同步
func (m *OfflineManager) MarkCacheSynced(key string) {
	m.mu.Lock()
	if entry := m.cache[key]; entry != nil {
		entry.Synced = true
	}
	m.mu.Unlock()
}

// Sync 同步到服务器
func (m *OfflineManager) Sync() error {
	if !m.IsOnline() {
		return nil
	}

	// 同步任务
	tasks := m.GetPendingTasks()
	for _, task := range tasks {
		// 发送到服务器
		log.Printf("Syncing task %s", task.ID)
		task.Synced = true
	}

	// 同步缓存
	entries := m.GetPendingCache()
	for _, entry := range entries {
		log.Printf("Syncing cache %s", entry.Key)
		entry.Synced = true
	}

	return nil
}

// Persist 持久化数据
func (m *OfflineManager) Persist() {
	if m.dataPath == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 保存任务
	tasksPath := filepath.Join(m.dataPath, "tasks.json")
	tasksData, _ := json.Marshal(m.tasks)
	os.WriteFile(tasksPath, tasksData, 0644)

	// 保存缓存
	cachePath := filepath.Join(m.dataPath, "cache.json")
	cacheData, _ := json.Marshal(m.cache)
	os.WriteFile(cachePath, cacheData, 0644)
}

// Load 加载数据
func (m *OfflineManager) Load() {
	if m.dataPath == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载任务
	tasksPath := filepath.Join(m.dataPath, "tasks.json")
	if data, err := os.ReadFile(tasksPath); err == nil {
		json.Unmarshal(data, &m.tasks)
	}

	// 加载缓存
	cachePath := filepath.Join(m.dataPath, "cache.json")
	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &m.cache)
	}
}

// Clear 清理数据
func (m *OfflineManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks = make(map[string]*LocalTask)
	m.taskList = make([]string, 0, m.config.MaxLocalTasks)
	m.cache = make(map[string]*CacheEntry)
	m.cacheList = make([]string, 0, m.config.MaxCacheSize)
}

// Stats 统计信息
func (m *OfflineManager) Stats() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()

	pending := 0
	for _, task := range m.tasks {
		if !task.Synced {
			pending++
		}
	}

	return map[string]int{
		"tasks":       len(m.tasks),
		"pending":     pending,
		"cache":       len(m.cache),
		"cache_pending": len(m.GetPendingCache()),
	}
}

func (m *OfflineManager) evictOldTasks() {
	// 移除已同步的旧任务
	for i, id := range m.taskList {
		if task := m.tasks[id]; task != nil && task.Synced {
			delete(m.tasks, id)
			m.taskList = append(m.taskList[:i], m.taskList[i+1:]...)
			break
		}
	}
}

func (m *OfflineManager) evictOldCache() {
	// 移除过期或已同步的缓存
	now := time.Now()
	for i, key := range m.cacheList {
		if entry := m.cache[key]; entry != nil {
			if entry.Synced || (!entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt)) {
				delete(m.cache, key)
				m.cacheList = append(m.cacheList[:i], m.cacheList[i+1:]...)
				break
			}
		}
	}
}

func generateLiteID() string {
	return time.Now().Format("150405.000") // HHMMSS.mmm
}