package sync

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// === 任务协同执行 (v3.3.0) ===
//
// Center 是永远在线的灵魂载体，任务协同执行确保：
// - 多设备协同完成任务
// - 任务拆分与分配
// - 结果聚合与合并
// - 失败转移与重试

// TaskStatus - 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusTimeout    TaskStatus = "timeout"
)

// TaskPriority - 任务优先级
type TaskPriority int

const (
	TaskPriorityLow      TaskPriority = 0
	TaskPriorityNormal   TaskPriority = 1
	TaskPriorityHigh     TaskPriority = 2
	TaskPriorityUrgent   TaskPriority = 3
)

// SubTask - 子任务
type SubTask struct {
	SubTaskID    string                 `json:"subtask_id"`
	ParentTaskID string                 `json:"parent_task_id"`
	AssignedTo   string                 `json:"assigned_to"`  // AgentID
	Type         string                 `json:"type"`
	Payload      map[string]interface{} `json:"payload"`
	Status       TaskStatus             `json:"status"`
	Priority     TaskPriority           `json:"priority"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	AssignedAt   *time.Time             `json:"assigned_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Timeout      time.Duration          `json:"timeout"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// CollaborativeTask - 协同任务
type CollaborativeTask struct {
	TaskID       string                 `json:"task_id"`
	IdentityID   string                 `json:"identity_id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Priority     TaskPriority           `json:"priority"`
	Status       TaskStatus             `json:"status"`
	Payload      map[string]interface{} `json:"payload"`

	// 拆分策略
	SplitStrategy SplitStrategy          `json:"split_strategy"`
	SubTasks      []*SubTask             `json:"subtasks"`

	// 聚合策略
	MergeStrategy MergeStrategy          `json:"merge_strategy"`
	Result        map[string]interface{} `json:"result,omitempty"`

	// 执行约束
	RequiredCapabilities []string        `json:"required_capabilities"`
	PreferredDevices     []string        `json:"preferred_devices"`
	MinDevices           int             `json:"min_devices"`
	MaxDevices           int             `json:"max_devices"`
	Timeout              time.Duration   `json:"timeout"`
	MaxRetries           int             `json:"max_retries"`

	// 时间信息
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`

	// 元数据
	Metadata     map[string]interface{} `json:"metadata"`
}

// SplitStrategy - 拆分策略
type SplitStrategy string

const (
	SplitNone      SplitStrategy = "none"      // 不拆分，单设备执行
	SplitParallel  SplitStrategy = "parallel"  // 并行拆分
	SplitSequence  SplitStrategy = "sequence"  // 顺序拆分
	SplitMapReduce SplitStrategy = "map_reduce" // MapReduce 模式
	SplitByDevice  SplitStrategy = "by_device" // 按设备能力拆分
)

// MergeStrategy - 聚合策略
type MergeStrategy string

const (
	MergeNone       MergeStrategy = "none"       // 不聚合
	MergeAll        MergeStrategy = "all"        // 收集所有结果
	MergeFirst      MergeStrategy = "first"      // 取第一个成功结果
	MergeConsensus  MergeStrategy = "consensus"  // 共识结果
	MergeAggregate  MergeStrategy = "aggregate"  // 聚合统计
	MergeBest       MergeStrategy = "best"       // 取最佳结果
)

// TaskSplitter - 任务拆分器
type TaskSplitter interface {
	Split(task *CollaborativeTask) ([]*SubTask, error)
}

// ResultMerger - 结果合并器
type ResultMerger interface {
	Merge(task *CollaborativeTask, results []*SubTask) (map[string]interface{}, error)
}

// TaskCoordinator - 任务协调器
type TaskCoordinator struct {
	mu sync.RWMutex

	// 任务存储
	tasks      map[string]*CollaborativeTask
	subTasks   map[string]*SubTask

	// 拆分器
	splitters  map[SplitStrategy]TaskSplitter

	// 合并器
	mergers    map[MergeStrategy]ResultMerger

	// 设备管理器
	deviceManager *DeviceManager

	// 消息总线
	messageBus *MessageBus

	// 状态管理器
	stateManager *StateSyncManager

	// 配置
	config TaskCoordinatorConfig

	// 任务监听器
	listeners []TaskListener
}

// TaskCoordinatorConfig - 配置
type TaskCoordinatorConfig struct {
	DefaultTimeout   time.Duration
	DefaultMaxRetries int
	MaxConcurrentTasks int
	TaskExpiration    time.Duration
}

// DefaultTaskCoordinatorConfig 默认配置
func DefaultTaskCoordinatorConfig() TaskCoordinatorConfig {
	return TaskCoordinatorConfig{
		DefaultTimeout:      30 * time.Minute,
		DefaultMaxRetries:   3,
		MaxConcurrentTasks:  100,
		TaskExpiration:      24 * time.Hour,
	}
}

// TaskListener - 任务监听器
type TaskListener interface {
	OnTaskCreated(task *CollaborativeTask)
	OnTaskAssigned(task *CollaborativeTask, agentID string)
	OnSubTaskCompleted(subTask *SubTask)
	OnTaskCompleted(task *CollaborativeTask)
	OnTaskFailed(task *CollaborativeTask, err error)
}

// NewTaskCoordinator 创建任务协调器
func NewTaskCoordinator(config TaskCoordinatorConfig) *TaskCoordinator {
	if config.DefaultTimeout == 0 {
		config = DefaultTaskCoordinatorConfig()
	}

	tc := &TaskCoordinator{
		tasks:     make(map[string]*CollaborativeTask),
		subTasks:  make(map[string]*SubTask),
		splitters: make(map[SplitStrategy]TaskSplitter),
		mergers:   make(map[MergeStrategy]ResultMerger),
		config:    config,
		listeners: make([]TaskListener, 0),
	}

	// 注册默认拆分器
	tc.registerDefaultSplitters()

	// 注册默认合并器
	tc.registerDefaultMergers()

	return tc
}

// SetDeviceManager 设置设备管理器
func (tc *TaskCoordinator) SetDeviceManager(dm *DeviceManager) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.deviceManager = dm
}

// SetMessageBus 设置消息总线
func (tc *TaskCoordinator) SetMessageBus(mb *MessageBus) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.messageBus = mb
}

// SetStateManager 设置状态管理器
func (tc *TaskCoordinator) SetStateManager(sm *StateSyncManager) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.stateManager = sm
}

// === 任务创建与管理 ===

// CreateTask 创建任务
func (tc *TaskCoordinator) CreateTask(task *CollaborativeTask) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 设置默认值
	if task.TaskID == "" {
		task.TaskID = generateTaskID()
	}
	if task.Status == "" {
		task.Status = TaskStatusPending
	}
	if task.Timeout == 0 {
		task.Timeout = tc.config.DefaultTimeout
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = tc.config.DefaultMaxRetries
	}
	task.CreatedAt = time.Now()
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}

	// 存储任务
	tc.tasks[task.TaskID] = task

	// 通知监听器
	go tc.notifyTaskCreated(task)

	log.Printf("Task created: %s (type=%s, strategy=%s)", task.TaskID, task.Type, task.SplitStrategy)

	return nil
}

// GetTask 获取任务
func (tc *TaskCoordinator) GetTask(taskID string) *CollaborativeTask {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tasks[taskID]
}

// GetSubTask 获取子任务
func (tc *TaskCoordinator) GetSubTask(subTaskID string) *SubTask {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.subTasks[subTaskID]
}

// === 任务执行 ===

// StartTask 启动任务
func (tc *TaskCoordinator) StartTask(taskID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	task, exists := tc.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	if task.Status != TaskStatusPending {
		return ErrTaskAlreadyStarted
	}

	// 拆分任务
	subTasks, err := tc.splitTask(task)
	if err != nil {
		task.Status = TaskStatusFailed
		task.Metadata["error"] = err.Error()
		return err
	}

	task.SubTasks = subTasks
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now

	// 存储子任务
	for _, st := range subTasks {
		tc.subTasks[st.SubTaskID] = st
	}

	// 分配子任务到设备
	err = tc.assignSubTasks(task)
	if err != nil {
		log.Printf("Failed to assign subtasks: %v", err)
		return err
	}

	log.Printf("Task started: %s with %d subtasks", task.TaskID, len(subTasks))

	return nil
}

// splitTask 拆分任务
func (tc *TaskCoordinator) splitTask(task *CollaborativeTask) ([]*SubTask, error) {
	splitter, exists := tc.splitters[task.SplitStrategy]
	if !exists {
		// 默认不拆分
		return tc.createSingleSubTask(task), nil
	}

	return splitter.Split(task)
}

// createSingleSubTask 创建单一子任务
func (tc *TaskCoordinator) createSingleSubTask(task *CollaborativeTask) []*SubTask {
	return []*SubTask{{
		SubTaskID:    task.TaskID + "_0",
		ParentTaskID: task.TaskID,
		Type:         task.Type,
		Payload:      task.Payload,
		Status:       TaskStatusPending,
		Priority:     task.Priority,
		CreatedAt:    time.Now(),
		Timeout:      task.Timeout,
		MaxRetries:   task.MaxRetries,
	}}
}

// assignSubTasks 分配子任务
func (tc *TaskCoordinator) assignSubTasks(task *CollaborativeTask) error {
	if tc.deviceManager == nil {
		return ErrNoDeviceManager
	}

	// 获取可用设备
	devices := tc.deviceManager.GetActiveDevices(task.IdentityID)
	if len(devices) == 0 {
		return ErrNoAvailableDevices
	}

	// 过滤符合能力要求的设备
	var capableDevices []*DeviceInfo
	for _, d := range devices {
		if tc.hasCapabilities(d, task.RequiredCapabilities) {
			capableDevices = append(capableDevices, d)
		}
	}

	if len(capableDevices) < task.MinDevices {
		return ErrInsufficientDevices
	}

	// 分配子任务
	for i, st := range task.SubTasks {
		// 选择设备
		deviceIdx := i % len(capableDevices)
		device := capableDevices[deviceIdx]

		st.AssignedTo = device.AgentID
		st.AssignedAt = &[]time.Time{time.Now()}[0]
		st.Status = TaskStatusAssigned

		// 发送任务到设备
		if tc.messageBus != nil {
			tc.sendSubTaskToDevice(st)
		}

		// 通知监听器
		go tc.notifyTaskAssigned(task, device.AgentID)
	}

	return nil
}

// === 结果处理 ===

// ReportSubTaskResult 报告子任务结果
func (tc *TaskCoordinator) ReportSubTaskResult(subTaskID string, result map[string]interface{}, err error) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	subTask, exists := tc.subTasks[subTaskID]
	if !exists {
		return ErrSubTaskNotFound
	}

	task, exists := tc.tasks[subTask.ParentTaskID]
	if !exists {
		return ErrTaskNotFound
	}

	now := time.Now()

	if err != nil {
		subTask.Status = TaskStatusFailed
		subTask.Error = err.Error()

		// 检查是否需要重试
		if subTask.RetryCount < subTask.MaxRetries {
			subTask.RetryCount++
			subTask.Status = TaskStatusPending
			// 重新分配
			go tc.retrySubTask(subTask)
			return nil
		}
	} else {
		subTask.Status = TaskStatusCompleted
		subTask.Result = result
		subTask.CompletedAt = &now
	}

	// 通知监听器
	go tc.notifySubTaskCompleted(subTask)

	// 检查任务是否完成
	tc.checkTaskCompletion(task)

	return nil
}

// checkTaskCompletion 检查任务完成状态
func (tc *TaskCoordinator) checkTaskCompletion(task *CollaborativeTask) {
	completed := 0
	failed := 0

	for _, st := range task.SubTasks {
		switch st.Status {
		case TaskStatusCompleted:
			completed++
		case TaskStatusFailed:
			failed++
		}
	}

	// 所有子任务完成
	if completed+failed == len(task.SubTasks) {
		if failed == 0 {
			// 合并结果
			tc.mergeResults(task)

			task.Status = TaskStatusCompleted
			now := time.Now()
			task.CompletedAt = &now

			// 通知监听器
			go tc.notifyTaskCompleted(task)
		} else if completed == 0 {
			// 全部失败
			task.Status = TaskStatusFailed
			now := time.Now()
			task.CompletedAt = &now

			// 通知监听器
			go tc.notifyTaskFailed(task, ErrAllSubTasksFailed)
		} else {
			// 部分失败，检查是否可接受
			if task.MergeStrategy == MergeAll || task.MergeStrategy == MergeAggregate {
				tc.mergeResults(task)

				task.Status = TaskStatusCompleted
				now := time.Now()
				task.CompletedAt = &now

				go tc.notifyTaskCompleted(task)
			} else {
				task.Status = TaskStatusFailed
				now := time.Now()
				task.CompletedAt = &now

				go tc.notifyTaskFailed(task, ErrPartialFailure)
			}
		}
	}
}

// mergeResults 合并结果
func (tc *TaskCoordinator) mergeResults(task *CollaborativeTask) {
	merger, exists := tc.mergers[task.MergeStrategy]
	if !exists {
		// 默认收集所有结果
		task.Result = make(map[string]interface{})
		results := make([]map[string]interface{}, 0)
		for _, st := range task.SubTasks {
			if st.Status == TaskStatusCompleted && st.Result != nil {
				results = append(results, st.Result)
			}
		}
		task.Result["subtask_results"] = results
		return
	}

	result, err := merger.Merge(task, task.SubTasks)
	if err != nil {
		task.Metadata["merge_error"] = err.Error()
		task.Result = make(map[string]interface{})
		return
	}

	task.Result = result
}

// retrySubTask 重试子任务
func (tc *TaskCoordinator) retrySubTask(subTask *SubTask) {
	// 简化处理：直接发送到原设备
	if tc.messageBus != nil {
		tc.sendSubTaskToDevice(subTask)
	}
}

// sendSubTaskToDevice 发送子任务到设备
func (tc *TaskCoordinator) sendSubTaskToDevice(st *SubTask) {
	if tc.messageBus == nil {
		return
	}

	msg := &Message{
		ID:         generateMessageID(),
		FromAgent:  "center",
		ToAgent:    st.AssignedTo,
		Type:       MessageTypeCommand,
		Priority:   MessagePriority(st.Priority),
		Payload:    subTaskToMap(st),
		CreatedAt:  time.Now(),
	}

	tc.messageBus.Send(msg)
}

// === 默认拆分器与合并器 ===

func (tc *TaskCoordinator) registerDefaultSplitters() {
	// 并行拆分
	tc.splitters[SplitParallel] = &ParallelSplitter{}

	// 按设备拆分
	tc.splitters[SplitByDevice] = &DeviceSplitter{}
}

func (tc *TaskCoordinator) registerDefaultMergers() {
	// 收集所有
	tc.mergers[MergeAll] = &AllResultMerger{}

	// 取第一个
	tc.mergers[MergeFirst] = &FirstResultMerger{}

	// 聚合统计
	tc.mergers[MergeAggregate] = &AggregateMerger{}
}

// === 监听器管理 ===

func (tc *TaskCoordinator) AddListener(listener TaskListener) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.listeners = append(tc.listeners, listener)
}

func (tc *TaskCoordinator) notifyTaskCreated(task *CollaborativeTask) {
	for _, l := range tc.listeners {
		l.OnTaskCreated(task)
	}
}

func (tc *TaskCoordinator) notifyTaskAssigned(task *CollaborativeTask, agentID string) {
	for _, l := range tc.listeners {
		l.OnTaskAssigned(task, agentID)
	}
}

func (tc *TaskCoordinator) notifySubTaskCompleted(subTask *SubTask) {
	for _, l := range tc.listeners {
		l.OnSubTaskCompleted(subTask)
	}
}

func (tc *TaskCoordinator) notifyTaskCompleted(task *CollaborativeTask) {
	for _, l := range tc.listeners {
		l.OnTaskCompleted(task)
	}
}

func (tc *TaskCoordinator) notifyTaskFailed(task *CollaborativeTask, err error) {
	for _, l := range tc.listeners {
		l.OnTaskFailed(task, err)
	}
}

// === 辅助方法 ===

func (tc *TaskCoordinator) hasCapabilities(device *DeviceInfo, required []string) bool {
	if len(required) == 0 {
		return true
	}

	capMap := make(map[string]bool)
	for _, cap := range device.Capabilities {
		capMap[cap] = true
	}

	for _, req := range required {
		if !capMap[req] {
			return false
		}
	}

	return true
}

// GetTaskStats 获取任务统计
func (tc *TaskCoordinator) GetTaskStats(identityID string) *TaskStats {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	stats := &TaskStats{}

	for _, task := range tc.tasks {
		if identityID != "" && task.IdentityID != identityID {
			continue
		}

		stats.TotalTasks++
		switch task.Status {
		case TaskStatusPending:
			stats.PendingTasks++
		case TaskStatusRunning:
			stats.RunningTasks++
		case TaskStatusCompleted:
			stats.CompletedTasks++
		case TaskStatusFailed:
			stats.FailedTasks++
		}
		stats.TotalSubTasks += len(task.SubTasks)
	}

	return stats
}

// TaskStats 任务统计
type TaskStats struct {
	TotalTasks      int `json:"total_tasks"`
	PendingTasks    int `json:"pending_tasks"`
	RunningTasks    int `json:"running_tasks"`
	CompletedTasks  int `json:"completed_tasks"`
	FailedTasks     int `json:"failed_tasks"`
	TotalSubTasks   int `json:"total_subtasks"`
}

// === 错误定义 ===

var (
	ErrTaskNotFound       = &TaskError{Code: "task_not_found", Message: "Task not found"}
	ErrTaskAlreadyStarted = &TaskError{Code: "task_already_started", Message: "Task already started"}
	ErrSubTaskNotFound    = &TaskError{Code: "subtask_not_found", Message: "Subtask not found"}
	ErrNoDeviceManager    = &TaskError{Code: "no_device_manager", Message: "No device manager configured"}
	ErrNoAvailableDevices = &TaskError{Code: "no_available_devices", Message: "No available devices"}
	ErrInsufficientDevices = &TaskError{Code: "insufficient_devices", Message: "Insufficient devices"}
	ErrAllSubTasksFailed  = &TaskError{Code: "all_subtasks_failed", Message: "All subtasks failed"}
	ErrPartialFailure     = &TaskError{Code: "partial_failure", Message: "Partial failure"}
)

// TaskError 任务错误
type TaskError struct {
	Code    string
	Message string
}

func (e *TaskError) Error() string {
	return e.Message
}

// === 拆分器实现 ===

// ParallelSplitter 并行拆分器
type ParallelSplitter struct{}

func (s *ParallelSplitter) Split(task *CollaborativeTask) ([]*SubTask, error) {
	// 简单并行拆分：每个设备一个子任务
	// 实际应用中根据任务类型具体实现
	return []*SubTask{{
		SubTaskID:    task.TaskID + "_0",
		ParentTaskID: task.TaskID,
		Type:         task.Type,
		Payload:      task.Payload,
		Status:       TaskStatusPending,
		Priority:     task.Priority,
		CreatedAt:    time.Now(),
		Timeout:      task.Timeout,
		MaxRetries:   task.MaxRetries,
	}}, nil
}

// DeviceSplitter 按设备拆分
type DeviceSplitter struct{}

func (s *DeviceSplitter) Split(task *CollaborativeTask) ([]*SubTask, error) {
	// 根据设备数量拆分
	// 实际应用中根据任务类型和设备能力具体实现
	return []*SubTask{{
		SubTaskID:    task.TaskID + "_0",
		ParentTaskID: task.TaskID,
		Type:         task.Type,
		Payload:      task.Payload,
		Status:       TaskStatusPending,
		Priority:     task.Priority,
		CreatedAt:    time.Now(),
		Timeout:      task.Timeout,
		MaxRetries:   task.MaxRetries,
	}}, nil
}

// === 合并器实现 ===

// AllResultMerger 收集所有结果
type AllResultMerger struct{}

func (m *AllResultMerger) Merge(task *CollaborativeTask, subTasks []*SubTask) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	results := make([]map[string]interface{}, 0)

	for _, st := range subTasks {
		if st.Status == TaskStatusCompleted && st.Result != nil {
			results = append(results, st.Result)
		}
	}

	result["subtask_results"] = results
	result["count"] = len(results)

	return result, nil
}

// FirstResultMerger 取第一个结果
type FirstResultMerger struct{}

func (m *FirstResultMerger) Merge(task *CollaborativeTask, subTasks []*SubTask) (map[string]interface{}, error) {
	for _, st := range subTasks {
		if st.Status == TaskStatusCompleted && st.Result != nil {
			return st.Result, nil
		}
	}
	return nil, ErrAllSubTasksFailed
}

// AggregateMerger 聚合统计
type AggregateMerger struct{}

func (m *AggregateMerger) Merge(task *CollaborativeTask, subTasks []*SubTask) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	completed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, st := range subTasks {
		if st.Status == TaskStatusCompleted {
			completed++
			if st.CompletedAt != nil && st.AssignedAt != nil {
				totalDuration += st.CompletedAt.Sub(*st.AssignedAt)
			}
		} else if st.Status == TaskStatusFailed {
			failed++
		}
	}

	result["completed"] = completed
	result["failed"] = failed
	result["total"] = len(subTasks)
	if completed > 0 {
		result["avg_duration_ms"] = totalDuration.Milliseconds() / int64(completed)
	}

	return result, nil
}

// === 辅助函数 ===

func generateTaskID() string {
	return "task_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func subTaskToMap(st *SubTask) map[string]interface{} {
	m := make(map[string]interface{})
	m["subtask_id"] = st.SubTaskID
	m["parent_task_id"] = st.ParentTaskID
	m["type"] = st.Type
	m["payload"] = st.Payload
	m["priority"] = st.Priority
	m["timeout_ms"] = st.Timeout.Milliseconds()
	m["max_retries"] = st.MaxRetries
	return m
}

// ToJSON 序列化任务
func (t *CollaborativeTask) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON 解析任务
func CollaborativeTaskFromJSON(data []byte) (*CollaborativeTask, error) {
	var task CollaborativeTask
	err := json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}