package local

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// 测试用技能处理器
func testSkillHandler(ctx context.Context, input []byte) ([]byte, error) {
	return append([]byte("processed: "), input...), nil
}

func failingSkillHandler(ctx context.Context, input []byte) ([]byte, error) {
	return nil, errors.New("skill failed")
}

func slowSkillHandler(ctx context.Context, input []byte) ([]byte, error) {
	time.Sleep(100 * time.Millisecond)
	return []byte("slow result"), nil
}

func TestNewLocalScheduler(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 2)
	if scheduler == nil {
		t.Fatal("NewLocalScheduler() returned nil")
	}
	if scheduler.level != OfflineL1 {
		t.Errorf("level = %v, want %v", scheduler.level, OfflineL1)
	}
	if scheduler.workers != 2 {
		t.Errorf("workers = %v, want 2", scheduler.workers)
	}
}

func TestLocalScheduler_StartStop(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 2)
	ctx := context.Background()

	// 启动
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// 重复启动
	err = scheduler.Start(ctx)
	if err == nil {
		t.Error("second Start() should fail")
	}

	// 停止
	scheduler.Stop()
}

func TestLocalScheduler_RegisterSkill(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)

	skill := &LocalSkill{
		ID:             "test.skill",
		Name:           "Test Skill",
		Category:       "test",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	}

	err := scheduler.RegisterSkill(skill)
	if err != nil {
		t.Fatalf("RegisterSkill() error = %v", err)
	}

	// 重复注册
	err = scheduler.RegisterSkill(skill)
	if err == nil {
		t.Error("duplicate RegisterSkill() should fail")
	}
}

func TestLocalScheduler_Execute(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 2)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	// 注册技能
	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.echo",
		Name:           "Echo",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	})

	// 执行任务
	task, err := scheduler.Execute(ctx, "test.echo", []byte("hello"))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if task == nil {
		t.Fatal("task is nil")
	}
	if task.ID == "" {
		t.Error("task ID is empty")
	}
	if task.SkillID != "test.echo" {
		t.Errorf("SkillID = %v, want test.echo", task.SkillID)
	}
	if task.Status != TaskPending {
		t.Errorf("Status = %v, want Pending", task.Status)
	}
}

func TestLocalScheduler_ExecuteSync(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 2)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.sync",
		Name:           "Sync",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	})

	task, err := scheduler.ExecuteSync(ctx, "test.sync", []byte("test input"))
	if err != nil {
		t.Fatalf("ExecuteSync() error = %v", err)
	}

	if task.Status != TaskCompleted {
		t.Errorf("Status = %v, want Completed", task.Status)
	}
	if string(task.Output) != "processed: test input" {
		t.Errorf("Output = %s, want 'processed: test input'", string(task.Output))
	}
}

func TestLocalScheduler_ExecuteNonExistentSkill(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	_, err := scheduler.Execute(ctx, "non.existent", nil)
	if err == nil {
		t.Error("Execute() should fail for non-existent skill")
	}
}

func TestLocalScheduler_ExecuteFailingSkill(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.failing",
		Name:           "Failing",
		OfflineCapable: true,
		Handler:        failingSkillHandler,
	})

	task, err := scheduler.ExecuteSync(ctx, "test.failing", []byte("test"))
	if err != nil {
		t.Fatalf("ExecuteSync() error = %v", err)
	}

	if task.Status != TaskFailed {
		t.Errorf("Status = %v, want Failed", task.Status)
	}
	if task.Error == "" {
		t.Error("Error message is empty")
	}
}

func TestLocalScheduler_GetTask(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.get",
		Name:           "Get",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	})

	task, _ := scheduler.Execute(ctx, "test.get", []byte("test"))

	// 等待执行完成
	time.Sleep(50 * time.Millisecond)

	// 查询任务
	found, exists := scheduler.GetTask(task.ID)
	if !exists {
		t.Fatal("GetTask() task not found")
	}
	if found.ID != task.ID {
		t.Errorf("ID = %v, want %v", found.ID, task.ID)
	}
}

func TestLocalScheduler_OfflineMode(t *testing.T) {
	// 完全离线模式
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	skill := &LocalSkill{
		ID:             "offline.skill",
		Name:           "Offline Skill",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	}
	scheduler.RegisterSkill(skill)

	// 离线模式下应能执行
	task, err := scheduler.ExecuteSync(ctx, "offline.skill", []byte("offline test"))
	if err != nil {
		t.Fatalf("ExecuteSync() in offline mode error = %v", err)
	}
	if task.Status != TaskCompleted {
		t.Errorf("Status = %v, want Completed", task.Status)
	}
}

func TestLocalScheduler_SyncPending(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL3, 1) // 弱网同步模式
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.sync",
		Name:           "Sync",
		OfflineCapable: true,
		Handler:        testSkillHandler,
	})

	task, _ := scheduler.ExecuteSync(ctx, "test.sync", []byte("test"))

	// 弱网同步模式下，任务应标记为待同步
	if !task.SyncWhenOnline {
		t.Error("SyncWhenOnline should be true in L3 mode")
	}
	if !task.SyncPending {
		t.Error("SyncPending should be true after completion")
	}

	// 获取待同步任务
	pending := scheduler.GetPendingSyncTasks()
	if len(pending) == 0 {
		t.Error("GetPendingSyncTasks() should return pending tasks")
	}

	// 标记已同步
	scheduler.MarkSynced(task.ID)
	pending = scheduler.GetPendingSyncTasks()
	for _, p := range pending {
		if p.ID == task.ID {
			t.Error("Task should not be in pending list after MarkSynced")
		}
	}
}

func TestLocalScheduler_ConcurrentExecution(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 4)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "test.concurrent",
		Name:           "Concurrent",
		OfflineCapable: true,
		Handler:        slowSkillHandler,
	})

	var wg sync.WaitGroup
	taskCount := 10
	results := make([]*LocalTask, taskCount)

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task, err := scheduler.ExecuteSync(ctx, "test.concurrent", []byte("test"))
			if err != nil {
				t.Errorf("ExecuteSync() error = %v", err)
				return
			}
			results[idx] = task
		}(i)
	}

	wg.Wait()

	// 验证所有任务完成
	for i, task := range results {
		if task == nil {
			t.Errorf("task %d is nil", i)
			continue
		}
		if task.Status != TaskCompleted {
			t.Errorf("task %d Status = %v, want Completed", i, task.Status)
		}
	}
}

func TestOfflineCache(t *testing.T) {
	cache := NewOfflineCache(1024)

	// Set
	cache.Set("key1", []byte("value1"))
	cache.Set("key2", []byte("value2"))

	// Get
	val, exists := cache.Get("key1")
	if !exists {
		t.Fatal("key1 not found")
	}
	if string(val) != "value1" {
		t.Errorf("value = %s, want value1", string(val))
	}

	// Size
	size := cache.Size()
	if size != 12 { // "value1" + "value2" = 12 bytes
		t.Errorf("size = %d, want 12", size)
	}

	// Delete
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	if exists {
		t.Error("key1 should be deleted")
	}

	// Clear
	cache.Clear()
	size = cache.Size()
	if size != 0 {
		t.Errorf("size after clear = %d, want 0", size)
	}
}

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskPending, "pending"},
		{TaskRunning, "running"},
		{TaskCompleted, "completed"},
		{TaskFailed, "failed"},
		{TaskCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskStatus_MarshalJSON(t *testing.T) {
	status := TaskCompleted
	data, err := status.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(data) != `"completed"` {
		t.Errorf("MarshalJSON() = %s, want \"completed\"", string(data))
	}
}