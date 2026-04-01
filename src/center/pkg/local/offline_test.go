package local

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestOfflineIntegration 离线模式集成测试
func TestOfflineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("CompleteOfflineWorkflow", func(t *testing.T) {
		// 创建完全离线的调度器 (L1)
		scheduler := NewLocalScheduler(OfflineL1, 4)
		err := scheduler.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}
		defer scheduler.Stop()

		// 注册多个本地技能
		skills := []*LocalSkill{
			{
				ID:             "text.process",
				Name:           "Text Processor",
				Category:       "text",
				OfflineCapable: true,
				Handler: func(ctx context.Context, input []byte) ([]byte, error) {
					return append([]byte("PROCESSED: "), input...), nil
				},
			},
			{
				ID:             "json.format",
				Name:           "JSON Formatter",
				Category:       "json",
				OfflineCapable: true,
				Handler: func(ctx context.Context, input []byte) ([]byte, error) {
					var data interface{}
					if err := json.Unmarshal(input, &data); err != nil {
						return nil, err
					}
					return json.MarshalIndent(data, "", "  ")
				},
			},
			{
				ID:             "calculator",
				Name:           "Calculator",
				Category:       "math",
				OfflineCapable: true,
				Handler: func(ctx context.Context, input []byte) ([]byte, error) {
					var req struct {
						Op    string  `json:"op"`
						A, B  float64 `json:"a,b"`
					}
					if err := json.Unmarshal(input, &req); err != nil {
						return nil, err
					}
					var result float64
					switch req.Op {
					case "add":
						result = req.A + req.B
					case "sub":
						result = req.A - req.B
					case "mul":
						result = req.A * req.B
					case "div":
						if req.B == 0 {
							return nil, fmt.Errorf("division by zero")
						}
						result = req.A / req.B
					}
					return json.Marshal(result)
				},
			},
		}

		for _, skill := range skills {
			if err := scheduler.RegisterSkill(skill); err != nil {
				t.Fatalf("Failed to register skill %s: %v", skill.ID, err)
			}
		}

		// 并发执行多个任务
		var wg sync.WaitGroup
		results := make(chan *LocalTask, 10)

		// 执行文本处理
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, err := scheduler.ExecuteSync(ctx, "text.process", []byte("hello offline"))
			if err != nil {
				t.Errorf("text.process failed: %v", err)
				return
			}
			results <- task
		}()

		// 执行JSON格式化
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, err := scheduler.ExecuteSync(ctx, "json.format", []byte(`{"name":"test","value":123}`))
			if err != nil {
				t.Errorf("json.format failed: %v", err)
				return
			}
			results <- task
		}()

		// 执行计算
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, err := scheduler.ExecuteSync(ctx, "calculator", []byte(`{"op":"add","a":10,"b":20}`))
			if err != nil {
				t.Errorf("calculator failed: %v", err)
				return
			}
			results <- task
		}()

		wg.Wait()
		close(results)

		// 验证结果
		taskCount := 0
		for task := range results {
			taskCount++
			if task.Status != TaskCompleted {
				t.Errorf("Task %s status = %v, want Completed", task.SkillID, task.Status)
			}
		}

		if taskCount != 3 {
			t.Errorf("Expected 3 completed tasks, got %d", taskCount)
		}
	})

	t.Run("OfflineCacheIntegration", func(t *testing.T) {
		cache := NewOfflineCache(1024 * 1024) // 1MB

		// 模拟缓存场景
		testData := map[string][]byte{
			"task:result:1": []byte(`{"status":"completed","output":"result1"}`),
			"task:result:2": []byte(`{"status":"completed","output":"result2"}`),
			"skill:config":  []byte(`{"timeout":30,"retries":3}`),
		}

		// 写入缓存
		for key, value := range testData {
			cache.Set(key, value)
		}

		// 读取并验证
		for key, expected := range testData {
			value, exists := cache.Get(key)
			if !exists {
				t.Errorf("Cache key %s not found", key)
				continue
			}
			if string(value) != string(expected) {
				t.Errorf("Cache value for %s = %s, want %s", key, value, expected)
			}
		}

		// 检查缓存大小
		size := cache.Size()
		if size == 0 {
			t.Error("Cache size should not be zero")
		}

		// 清空缓存
		cache.Clear()
		if cache.Size() != 0 {
			t.Error("Cache should be empty after clear")
		}
	})

	t.Run("L3SyncMode", func(t *testing.T) {
		// 弱网同步模式测试
		scheduler := NewLocalScheduler(OfflineL3, 2)
		ctx := context.Background()
		err := scheduler.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start L3 scheduler: %v", err)
		}
		defer scheduler.Stop()

		scheduler.RegisterSkill(&LocalSkill{
			ID:             "sync.test",
			Name:           "Sync Test",
			OfflineCapable: true,
			Handler: func(ctx context.Context, input []byte) ([]byte, error) {
				return append([]byte("synced: "), input...), nil
			},
		})

		// 执行多个任务
		for i := 0; i < 5; i++ {
			_, err := scheduler.ExecuteSync(ctx, "sync.test", []byte("test"))
			if err != nil {
				t.Errorf("Task %d failed: %v", i, err)
			}
		}

		// 检查待同步队列
		pending := scheduler.GetPendingSyncTasks()
		if len(pending) != 5 {
			t.Errorf("Expected 5 pending sync tasks, got %d", len(pending))
		}

		// 标记部分已同步
		for i, task := range pending {
			if i < 3 {
				scheduler.MarkSynced(task.ID)
			}
		}

		// 验证剩余待同步
		remaining := scheduler.GetPendingSyncTasks()
		if len(remaining) != 2 {
			t.Errorf("Expected 2 remaining sync tasks, got %d", len(remaining))
		}
	})
}

// TestOfflineLevelCapabilities 测试不同离线等级的能力
func TestOfflineLevelCapabilities(t *testing.T) {
	tests := []struct {
		name  string
		level OfflineLevel
	}{
		{"L1 - 完全离线", OfflineL1},
		{"L2 - 局域网协作", OfflineL2},
		{"L3 - 弱网同步", OfflineL3},
		{"L4 - 在线模式", OfflineL4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewLocalScheduler(tt.level, 2)
			if scheduler.level != tt.level {
				t.Errorf("OfflineLevel = %v, want %v", scheduler.level, tt.level)
			}
		})
	}
}

// TestOfflineSkillRegistration 测试离线技能注册
func TestOfflineSkillRegistration(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	// 注册离线技能
	offlineSkill := &LocalSkill{
		ID:             "offline.only",
		Name:           "Offline Only Skill",
		Category:       "offline",
		OfflineCapable: true,
		Handler: func(ctx context.Context, input []byte) ([]byte, error) {
			return []byte("offline executed"), nil
		},
	}

	if err := scheduler.RegisterSkill(offlineSkill); err != nil {
		t.Fatalf("Failed to register offline skill: %v", err)
	}

	// 验证技能列表
	skills := scheduler.skills
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
	}
}

// TestConcurrentOfflineTasks 测试并发离线任务
func TestConcurrentOfflineTasks(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 8)
	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// 注册一个耗时的技能
	scheduler.RegisterSkill(&LocalSkill{
		ID:             "slow.task",
		Name:           "Slow Task",
		OfflineCapable: true,
		Handler: func(ctx context.Context, input []byte) ([]byte, error) {
			time.Sleep(50 * time.Millisecond)
			return append([]byte("done: "), input...), nil
		},
	})

	// 并发执行100个任务
	const taskCount = 100
	var wg sync.WaitGroup
	completed := make(chan struct{}, taskCount)

	start := time.Now()
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			task, err := scheduler.ExecuteSync(ctx, "slow.task", []byte("test"))
			if err != nil {
				t.Errorf("Task failed: %v", err)
				return
			}
			if task.Status == TaskCompleted {
				completed <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(completed)

	duration := time.Since(start)

	// 统计完成任务数
	count := 0
	for range completed {
		count++
	}

	if count != taskCount {
		t.Errorf("Completed %d tasks, want %d", count, taskCount)
	}

	// 验证并发执行（8个worker应该比顺序执行快）
	// 顺序执行: 100 * 50ms = 5s
	// 并发执行应该显著更快
	if duration > 2*time.Second {
		t.Errorf("Concurrent execution took %v, expected faster", duration)
	}

	t.Logf("Executed %d tasks in %v with 8 workers", taskCount, duration)
}

// TestTaskPersistence 测试任务持久化（模拟）
func TestTaskPersistence(t *testing.T) {
	scheduler := NewLocalScheduler(OfflineL1, 1)
	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop()

	scheduler.RegisterSkill(&LocalSkill{
		ID:             "persist.test",
		Name:           "Persist Test",
		OfflineCapable: true,
		Handler: func(ctx context.Context, input []byte) ([]byte, error) {
			return input, nil
		},
	})

	// 执行任务
	task, err := scheduler.ExecuteSync(ctx, "persist.test", []byte("test data"))
	if err != nil {
		t.Fatalf("Task execution failed: %v", err)
	}

	// 验证可以查询到任务
	found, exists := scheduler.GetTask(task.ID)
	if !exists {
		t.Error("Task should exist in task store")
	}

	if found.ID != task.ID {
		t.Errorf("Found task ID = %s, want %s", found.ID, task.ID)
	}

	// 验证任务状态
	if found.Status != TaskCompleted {
		t.Errorf("Task status = %v, want Completed", found.Status)
	}
}