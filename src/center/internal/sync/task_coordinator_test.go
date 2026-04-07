package sync

import (
	"testing"
	"time"
)

func TestTaskCoordinator(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	// 测试默认拆分器和合并器初始化
	if len(tc.splitters) == 0 {
		t.Error("Expected default splitters to be initialized")
	}
	if len(tc.mergers) == 0 {
		t.Error("Expected default mergers to be initialized")
	}
}

func TestTaskCoordinatorCreateTask(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	task := &CollaborativeTask{
		TaskID:     "task-001",
		IdentityID: "identity-001",
		Type:       "test_task",
		Payload:    map[string]interface{}{"key": "value"},
	}

	err := tc.CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	// 验证任务已存储
	stored := tc.GetTask("task-001")
	if stored == nil {
		t.Fatal("Task not stored")
	}

	if stored.Status != TaskStatusPending {
		t.Errorf("Expected pending status, got %s", stored.Status)
	}
}

func TestTaskCoordinatorGetTaskStats(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	// 创建多个任务
	for i := 0; i < 5; i++ {
		task := &CollaborativeTask{
			TaskID:     "task-" + string(rune('0'+i)),
			IdentityID: "identity-001",
			Type:       "test_task",
		}
		tc.CreateTask(task)
	}

	stats := tc.GetTaskStats("identity-001")

	if stats.TotalTasks != 5 {
		t.Errorf("Expected 5 tasks, got %d", stats.TotalTasks)
	}

	if stats.PendingTasks != 5 {
		t.Errorf("Expected 5 pending tasks, got %d", stats.PendingTasks)
	}
}

func TestTaskCoordinatorAutoTaskID(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	task := &CollaborativeTask{
		IdentityID: "identity-001",
		Type:       "test_task",
	}

	tc.CreateTask(task)

	if task.TaskID == "" {
		t.Error("TaskID should be auto-generated")
	}
}

func TestTaskCoordinatorDefaultValues(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	task := &CollaborativeTask{
		IdentityID: "identity-001",
		Type:       "test_task",
	}

	tc.CreateTask(task)

	if task.Timeout != tc.config.DefaultTimeout {
		t.Errorf("Expected default timeout, got %v", task.Timeout)
	}

	if task.MaxRetries != tc.config.DefaultMaxRetries {
		t.Errorf("Expected default max retries, got %d", task.MaxRetries)
	}
}

func TestTaskCoordinatorListener(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	created := false
	listener := &TestTaskListener{
		onCreated: func(task *CollaborativeTask) {
			created = true
		},
	}

	tc.AddListener(listener)

	task := &CollaborativeTask{
		TaskID:     "task-listener",
		IdentityID: "identity-001",
		Type:       "test_task",
	}

	tc.CreateTask(task)

	// 等待异步通知
	time.Sleep(100 * time.Millisecond)

	if !created {
		t.Error("Listener should have been notified")
	}
}

func TestSubTask(t *testing.T) {
	subTask := &SubTask{
		SubTaskID:    "sub-001",
		ParentTaskID: "task-001",
		Type:         "test_subtask",
		Payload:      map[string]interface{}{"key": "value"},
	}

	if subTask.Status != TaskStatusPending {
		t.Error("Default status should be pending")
	}
}

func TestCollaborativeTaskJSON(t *testing.T) {
	task := &CollaborativeTask{
		TaskID:       "task-json",
		IdentityID:   "identity-001",
		Type:         "test_task",
		Description:  "Test task for JSON",
		Priority:     TaskPriorityHigh,
		Status:       TaskStatusPending,
		SplitStrategy: SplitParallel,
		MergeStrategy: MergeAll,
		Payload:      map[string]interface{}{"key": "value"},
	}

	// 序列化
	data, err := task.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := CollaborativeTaskFromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.TaskID != task.TaskID {
		t.Error("TaskID mismatch after JSON roundtrip")
	}
	if parsed.Type != task.Type {
		t.Error("Type mismatch after JSON roundtrip")
	}
	if parsed.Priority != task.Priority {
		t.Error("Priority mismatch after JSON roundtrip")
	}
}

func TestSplitStrategies(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	// 测试并行拆分
	task := &CollaborativeTask{
		TaskID:        "task-parallel",
		IdentityID:    "identity-001",
		Type:          "test_task",
		SplitStrategy: SplitParallel,
		Payload:       map[string]interface{}{"data": "test"},
	}

	subTasks, err := tc.splitTask(task)
	if err != nil {
		t.Fatalf("SplitTask failed: %v", err)
	}

	if len(subTasks) == 0 {
		t.Error("Expected at least one subtask")
	}

	for _, st := range subTasks {
		if st.ParentTaskID != task.TaskID {
			t.Error("ParentTaskID mismatch")
		}
	}
}

func TestMergeStrategies(t *testing.T) {
	tc := NewTaskCoordinator(DefaultTaskCoordinatorConfig())

	// 创建带有完成子任务的任务
	task := &CollaborativeTask{
		TaskID:        "task-merge",
		IdentityID:    "identity-001",
		Type:          "test_task",
		MergeStrategy: MergeAll,
		SubTasks: []*SubTask{
			{
				SubTaskID:   "sub-1",
				Status:      TaskStatusCompleted,
				Result:      map[string]interface{}{"value": 1},
				CompletedAt: &[]time.Time{time.Now()}[0],
				AssignedAt:  &[]time.Time{time.Now().Add(-time.Minute)}[0],
			},
			{
				SubTaskID:   "sub-2",
				Status:      TaskStatusCompleted,
				Result:      map[string]interface{}{"value": 2},
				CompletedAt: &[]time.Time{time.Now()}[0],
				AssignedAt:  &[]time.Time{time.Now().Add(-time.Minute)}[0],
			},
		},
	}

	// 测试合并
	tc.mergeResults(task)

	if task.Result == nil {
		t.Fatal("Result should not be nil after merge")
	}

	results, ok := task.Result["subtask_results"]
	if !ok {
		t.Error("Expected subtask_results in result")
	}

	if results == nil {
		t.Error("subtask_results should not be nil")
	}
}

func TestAllResultMerger(t *testing.T) {
	merger := &AllResultMerger{}

	task := &CollaborativeTask{
		TaskID: "task-all",
	}

	subTasks := []*SubTask{
		{
			SubTaskID: "sub-1",
			Status:    TaskStatusCompleted,
			Result:    map[string]interface{}{"a": 1},
		},
		{
			SubTaskID: "sub-2",
			Status:    TaskStatusCompleted,
			Result:    map[string]interface{}{"b": 2},
		},
		{
			SubTaskID: "sub-3",
			Status:    TaskStatusFailed,
			Result:    nil,
		},
	}

	result, err := merger.Merge(task, subTasks)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	count, ok := result["count"].(int)
	if !ok || count != 2 {
		t.Errorf("Expected count=2, got %v", result["count"])
	}
}

func TestFirstResultMerger(t *testing.T) {
	merger := &FirstResultMerger{}

	task := &CollaborativeTask{
		TaskID: "task-first",
	}

	subTasks := []*SubTask{
		{
			SubTaskID: "sub-1",
			Status:    TaskStatusCompleted,
			Result:    map[string]interface{}{"first": true},
		},
		{
			SubTaskID: "sub-2",
			Status:    TaskStatusCompleted,
			Result:    map[string]interface{}{"first": false},
		},
	}

	result, err := merger.Merge(task, subTasks)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if result["first"] != true {
		t.Error("Expected first result")
	}
}

func TestAggregateMerger(t *testing.T) {
	merger := &AggregateMerger{}

	task := &CollaborativeTask{
		TaskID: "task-aggregate",
	}

	now := time.Now()
	subTasks := []*SubTask{
		{
			SubTaskID:   "sub-1",
			Status:      TaskStatusCompleted,
			CompletedAt: &[]time.Time{now}[0],
			AssignedAt:  &[]time.Time{now.Add(-time.Minute)}[0],
		},
		{
			SubTaskID:   "sub-2",
			Status:      TaskStatusCompleted,
			CompletedAt: &[]time.Time{now}[0],
			AssignedAt:  &[]time.Time{now.Add(-2 * time.Minute)}[0],
		},
		{
			SubTaskID: "sub-3",
			Status:    TaskStatusFailed,
		},
	}

	result, err := merger.Merge(task, subTasks)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if result["completed"].(int) != 2 {
		t.Errorf("Expected completed=2, got %v", result["completed"])
	}

	if result["failed"].(int) != 1 {
		t.Errorf("Expected failed=1, got %v", result["failed"])
	}

	if result["total"].(int) != 3 {
		t.Errorf("Expected total=3, got %v", result["total"])
	}
}

func TestTaskErrors(t *testing.T) {
	// 测试错误定义
	if ErrTaskNotFound.Code != "task_not_found" {
		t.Error("Wrong error code for ErrTaskNotFound")
	}

	if ErrNoAvailableDevices.Code != "no_available_devices" {
		t.Error("Wrong error code for ErrNoAvailableDevices")
	}

	if ErrTaskNotFound.Error() != "Task not found" {
		t.Error("Wrong error message")
	}
}

// 测试监听器
type TestTaskListener struct {
	onCreated      func(task *CollaborativeTask)
	onAssigned     func(task *CollaborativeTask, agentID string)
	onSubCompleted func(subTask *SubTask)
	onCompleted    func(task *CollaborativeTask)
	onFailed       func(task *CollaborativeTask, err error)
}

func (l *TestTaskListener) OnTaskCreated(task *CollaborativeTask) {
	if l.onCreated != nil {
		l.onCreated(task)
	}
}

func (l *TestTaskListener) OnTaskAssigned(task *CollaborativeTask, agentID string) {
	if l.onAssigned != nil {
		l.onAssigned(task, agentID)
	}
}

func (l *TestTaskListener) OnSubTaskCompleted(subTask *SubTask) {
	if l.onSubCompleted != nil {
		l.onSubCompleted(subTask)
	}
}

func (l *TestTaskListener) OnTaskCompleted(task *CollaborativeTask) {
	if l.onCompleted != nil {
		l.onCompleted(task)
	}
}

func (l *TestTaskListener) OnTaskFailed(task *CollaborativeTask, err error) {
	if l.onFailed != nil {
		l.onFailed(task, err)
	}
}