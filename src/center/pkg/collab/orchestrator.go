// Package collab - 任务编排器
// 0.9.0 Beta: 智能Agent协作
package collab

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Orchestrator 任务编排器
type Orchestrator struct {
	taskExecutor TaskExecutor
	stateTracker *StateTracker
	mu           sync.RWMutex
}

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *CollabTask) (*TaskResult, error)
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	AgentID   string                 `json:"agent_id"`
}

// StateTracker 状态追踪器
type StateTracker struct {
	taskStates map[string]TaskState
	mu         sync.RWMutex
}

// NewStateTracker 创建状态追踪器
func NewStateTracker() *StateTracker {
	return &StateTracker{
		taskStates: make(map[string]TaskState),
	}
}

// UpdateState 更新状态
func (t *StateTracker) UpdateState(taskID string, state TaskState) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.taskStates[taskID] = state
}

// GetState 获取状态
func (t *StateTracker) GetState(taskID string) TaskState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.taskStates[taskID]
}

// NewOrchestrator 创建编排器
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		taskExecutor: &DefaultTaskExecutor{},
		stateTracker: NewStateTracker(),
	}
}

// SetTaskExecutor 设置任务执行器
func (o *Orchestrator) SetTaskExecutor(executor TaskExecutor) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.taskExecutor = executor
}

// ExecuteSequential 顺序执行
func (o *Orchestrator) ExecuteSequential(ctx context.Context, collab *Collaboration) error {
	o.mu.RLock()
	executor := o.taskExecutor
	o.mu.RUnlock()

	// 按优先级排序任务
	tasks := sortTasksByPriority(collab.Tasks)

	for _, task := range collab.Tasks {
		// 检查依赖
		if !o.checkDependencies(ctx, collab, task) {
			task.State = TaskStateFailed
			return fmt.Errorf("任务 %s 依赖未满足", task.ID)
		}

		// 更新状态
		task.State = TaskStateRunning
		o.stateTracker.UpdateState(task.ID, TaskStateRunning)

		// 执行任务
		result, err := executor.ExecuteTask(ctx, task)
		if err != nil {
			task.State = TaskStateFailed
			o.stateTracker.UpdateState(task.ID, TaskStateFailed)
			return fmt.Errorf("任务 %s 执行失败: %w", task.ID, err)
		}

		// 更新结果
		task.Output = result.Output
		task.State = TaskStateCompleted
		o.stateTracker.UpdateState(task.ID, TaskStateCompleted)
	}

	return nil
}

// ExecuteParallel 并行执行
func (o *Orchestrator) ExecuteParallel(ctx context.Context, collab *Collaboration) error {
	o.mu.RLock()
	executor := o.taskExecutor
	o.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(collab.Tasks))

	for _, task := range collab.Tasks {
		wg.Add(1)
		go func(t *CollabTask) {
			defer wg.Done()

			t.State = TaskStateRunning
			o.stateTracker.UpdateState(t.ID, TaskStateRunning)

			result, err := executor.ExecuteTask(ctx, t)
			if err != nil {
				t.State = TaskStateFailed
				o.stateTracker.UpdateState(t.ID, TaskStateFailed)
				errChan <- fmt.Errorf("任务 %s 执行失败: %w", t.ID, err)
				return
			}

			t.Output = result.Output
			t.State = TaskStateCompleted
			o.stateTracker.UpdateState(t.ID, TaskStateCompleted)
		}(task)
	}

	wg.Wait()

	// 检查错误
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// ExecutePipeline 管道执行
func (o *Orchestrator) ExecutePipeline(ctx context.Context, collab *Collaboration) error {
	o.mu.RLock()
	executor := o.taskExecutor
	o.mu.RUnlock()

	// 构建管道
	pipeline := o.buildPipeline(collab.Tasks)

	// 初始输入
	input := make(map[string]interface{})
	for k, v := range collab.Goal {
		input[k] = v
	}

	// 管道阶段执行
	for stage, tasks := range pipeline {
		stageInput := input

		for _, task := range tasks {
			task.State = TaskStateRunning
			o.stateTracker.UpdateState(task.ID, TaskStateRunning)

			// 合并输入
			taskInput := mergeInputs(stageInput, task.Input)

			result, err := executor.ExecuteTask(ctx, task)
			if err != nil {
				task.State = TaskStateFailed
				o.stateTracker.UpdateState(task.ID, TaskStateFailed)
				return fmt.Errorf("管道阶段 %d 任务 %s 失败: %w", stage, task.ID, err)
			}

			task.Output = result.Output
			task.State = TaskStateCompleted
			o.stateTracker.UpdateState(task.ID, TaskStateCompleted)

			// 输出传递到下一阶段
			for k, v := range result.Output {
				input[k] = v
			}
		}
	}

	return nil
}

// ExecuteMapReduce MapReduce执行
func (o *Orchestrator) ExecuteMapReduce(ctx context.Context, collab *Collaboration) error {
	o.mu.RLock()
	executor := o.taskExecutor
	o.mu.RUnlock()

	// 分离Map和Reduce任务
	mapTasks, reduceTasks := o.separateMapReduce(collab.Tasks)

	// Map阶段 - 并行执行
	mapResults := make(chan *TaskResult, len(mapTasks))
	var mapWg sync.WaitGroup

	for _, task := range mapTasks {
		mapWg.Add(1)
		go func(t *CollabTask) {
			defer mapWg.Done()

			t.State = TaskStateRunning
			o.stateTracker.UpdateState(t.ID, TaskStateRunning)

			result, err := executor.ExecuteTask(ctx, t)
			if err != nil {
				t.State = TaskStateFailed
				o.stateTracker.UpdateState(t.ID, TaskStateFailed)
				mapResults <- &TaskResult{TaskID: t.ID, Success: false, Error: err.Error()}
				return
			}

			t.Output = result.Output
			t.State = TaskStateCompleted
			o.stateTracker.UpdateState(t.ID, TaskStateCompleted)
			mapResults <- result
		}(task)
	}

	mapWg.Wait()
	close(mapResults)

	// 收集Map结果
	mapOutputs := make([]map[string]interface{}, 0)
	for result := range mapResults {
		if result.Success {
			mapOutputs = append(mapOutputs, result.Output)
		}
	}

	// Reduce阶段
	reduceInput := map[string]interface{}{
		"map_outputs": mapOutputs,
	}

	for _, task := range reduceTasks {
		task.Input = mergeInputs(reduceInput, task.Input)
		task.State = TaskStateRunning
		o.stateTracker.UpdateState(task.ID, TaskStateRunning)

		result, err := executor.ExecuteTask(ctx, task)
		if err != nil {
			task.State = TaskStateFailed
			o.stateTracker.UpdateState(task.ID, TaskStateFailed)
			return fmt.Errorf("Reduce任务 %s 失败: %w", task.ID, err)
		}

		task.Output = result.Output
		task.State = TaskStateCompleted
		o.stateTracker.UpdateState(task.ID, TaskStateCompleted)
	}

	return nil
}

// checkDependencies 检查依赖
func (o *Orchestrator) checkDependencies(ctx context.Context, collab *Collaboration, task *CollabTask) bool {
	for _, depID := range task.Dependencies {
		state := o.stateTracker.GetState(depID)
		if state != TaskStateCompleted {
			return false
		}
	}
	return true
}

// buildPipeline 构建管道
func (o *Orchestrator) buildPipeline(tasks []*CollabTask) map[int][]*CollabTask {
	pipeline := make(map[int][]*CollabTask)

	for _, task := range tasks {
		stage := task.Priority
		if pipeline[stage] == nil {
			pipeline[stage] = make([]*CollabTask, 0)
		}
		pipeline[stage] = append(pipeline[stage], task)
	}

	return pipeline
}

// separateMapReduce 分离Map和Reduce任务
func (o *Orchestrator) separateMapReduce(tasks []*CollabTask) ([]*CollabTask, []*CollabTask) {
	var mapTasks, reduceTasks []*CollabTask

	for _, task := range tasks {
		if task.Operation == "map" {
			mapTasks = append(mapTasks, task)
		} else if task.Operation == "reduce" {
			reduceTasks = append(reduceTasks, task)
		} else {
			// 默认为Map任务
			mapTasks = append(mapTasks, task)
		}
	}

	return mapTasks, reduceTasks
}

// sortTasksByPriority 按优先级排序
func sortTasksByPriority(tasks []*CollabTask) []*CollabTask {
	sorted := make([]*CollabTask, len(tasks))
	copy(sorted, tasks)

	// 简单排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// mergeInputs 合并输入
func mergeInputs(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overlay {
		result[k] = v
	}
	return result
}

// DefaultTaskExecutor 默认任务执行器
type DefaultTaskExecutor struct{}

// ExecuteTask 执行任务
func (e *DefaultTaskExecutor) ExecuteTask(ctx context.Context, task *CollabTask) (*TaskResult, error) {
	start := time.Now()

	// 模拟执行 - 实际实现会调用Agent
	output := make(map[string]interface{})
	output["task_id"] = task.ID
	output["operation"] = task.Operation
	output["status"] = "completed"

	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Output:   output,
		Duration: time.Since(start),
		AgentID:  task.AssignedTo,
	}, nil
}