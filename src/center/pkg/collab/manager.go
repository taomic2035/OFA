// Package collab - Agent协作系统
// 支持顺序、并行、管道、MapReduce等多种协作模式
package collab

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CollaborationType 协作类型
type CollaborationType string

const (
	CollabTypeSequential  CollaborationType = "sequential"  // 顺序执行
	CollabTypeParallel    CollaborationType = "parallel"    // 并行执行
	CollabTypePipeline    CollaborationType = "pipeline"    // 管道执行
	CollabTypeMapReduce   CollaborationType = "map_reduce"  // MapReduce
	CollabTypeConsensus   CollaborationType = "consensus"   // 共识
	CollabTypeAuction     CollaborationType = "auction"     // 拍卖
	CollabTypeNegotiation CollaborationType = "negotiation" // 协商
)

// Collaboration 协作定义
type Collaboration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        CollaborationType `json:"type"`
	Description string            `json:"description"`
	Goal        string            `json:"goal"`
	Tasks       []*CollabTask     `json:"tasks"`
	Agents      []*AgentRole      `json:"agents"`
	Constraints *Constraints      `json:"constraints"`
	State       CollabState       `json:"state"`
	Result      *CollabResult     `json:"result,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Timeout     time.Duration     `json:"timeout"`
}

// CollabTask 协作任务
type CollabTask struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	SkillID      string                 `json:"skill_id"`
	Operation    string                 `json:"operation"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output,omitempty"`
	Dependencies []string               `json:"dependencies"`
	AssignedTo   string                 `json:"assigned_to"`
	State        TaskState              `json:"state"`
	Priority     int                    `json:"priority"`
	Timeout      time.Duration          `json:"timeout"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// TaskState 任务状态
type TaskState string

const (
	TaskStatePending   TaskState = "pending"
	TaskStateAssigned  TaskState = "assigned"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
	TaskStateFailed    TaskState = "failed"
	TaskStateCancelled TaskState = "cancelled"
)

// AgentRole Agent角色
type AgentRole struct {
	AgentID     string            `json:"agent_id"`
	Role        string            `json:"role"`
	Capabilities []string         `json:"capabilities"`
	MaxTasks    int               `json:"max_tasks"`
	CurrentLoad int               `json:"current_load"`
	Priority    int               `json:"priority"`
	Metadata    map[string]string `json:"metadata"`
}

// Constraints 约束条件
type Constraints struct {
	MaxDuration   time.Duration `json:"max_duration"`
	MinAgents     int           `json:"min_agents"`
	MaxAgents     int           `json:"max_agents"`
	RequiredSkills []string     `json:"required_skills"`
	ExcludeAgents []string      `json:"exclude_agents"`
	PreferLocal   bool          `json:"prefer_local"`
	Budget        float64       `json:"budget"`
}

// CollabState 协作状态
type CollabState string

const (
	CollabStateCreated    CollabState = "created"
	CollabStatePlanning   CollabState = "planning"
	CollabStateRunning    CollabState = "running"
	CollabStateCompleted  CollabState = "completed"
	CollabStateFailed     CollabState = "failed"
	CollabStateCancelled  CollabState = "cancelled"
)

// CollabResult 协作结果
type CollabResult struct {
	Success      bool                   `json:"success"`
	Output       map[string]interface{} `json:"output"`
	TasksTotal   int                    `json:"tasks_total"`
	TasksSuccess int                    `json:"tasks_success"`
	TasksFailed  int                    `json:"tasks_failed"`
	Duration     time.Duration          `json:"duration"`
	Cost         float64                `json:"cost"`
	Errors       []string               `json:"errors,omitempty"`
}

// CollaborationManager 协作管理器
type CollaborationManager struct {
	collaborations map[string]*Collaboration
	orchestrator   *Orchestrator
	negotiator     *Negotiator
	allocator      *TaskAllocator
	aggregator     *ResultAggregator
	eventBus       *EventBus
	stats          *ManagerStats
	mu             sync.RWMutex
}

// ManagerStats 管理器统计
type ManagerStats struct {
	TotalCollabs      int64 `json:"total_collabs"`
	SuccessfulCollabs int64 `json:"successful_collabs"`
	FailedCollabs     int64 `json:"failed_collabs"`
	TotalTasks        int64 `json:"total_tasks"`
	TasksCompleted    int64 `json:"tasks_completed"`
	AvgDuration       int64 `json:"avg_duration_ms"`
}

// NewCollaborationManager 创建协作管理器
func NewCollaborationManager() *CollaborationManager {
	return &CollaborationManager{
		collaborations: make(map[string]*Collaboration),
		orchestrator:   NewOrchestrator(),
		negotiator:     NewNegotiator(),
		allocator:      NewTaskAllocator(),
		aggregator:     NewResultAggregator(),
		eventBus:       NewEventBus(),
		stats:          &ManagerStats{},
	}
}

// CreateCollaboration 创建协作
func (m *CollaborationManager) CreateCollaboration(ctx context.Context, req *CreateCollabRequest) (*Collaboration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	collab := &Collaboration{
		ID:          generateID(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Goal:        req.Goal,
		Tasks:       req.Tasks,
		Agents:      make([]*AgentRole, 0),
		Constraints: req.Constraints,
		State:       CollabStateCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Timeout:     req.Timeout,
	}

	m.collaborations[collab.ID] = collab
	m.stats.TotalCollabs++

	// 发布事件
	m.eventBus.Publish(&CollabEvent{
		Type:        "created",
		CollabID:    collab.ID,
		Timestamp:   time.Now(),
	})

	return collab, nil
}

// StartCollaboration 启动协作
func (m *CollaborationManager) StartCollaboration(ctx context.Context, collabID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	collab, ok := m.collaborations[collabID]
	if !ok {
		return fmt.Errorf("协作不存在: %s", collabID)
	}

	// 规划阶段
	collab.State = CollabStatePlanning
	if err := m.planCollaboration(ctx, collab); err != nil {
		collab.State = CollabStateFailed
		m.stats.FailedCollabs++
		return err
	}

	// 执行阶段
	collab.State = CollabStateRunning
	m.eventBus.Publish(&CollabEvent{
		Type:        "started",
		CollabID:    collab.ID,
		Timestamp:   time.Now(),
	})

	go m.executeCollaboration(context.Background(), collab)

	return nil
}

// planCollaboration 规划协作
func (m *CollaborationManager) planCollaboration(ctx context.Context, collab *Collaboration) error {
	// 分配任务
	assignments, err := m.allocator.Allocate(collab)
	if err != nil {
		return fmt.Errorf("任务分配失败: %w", err)
	}

	// 更新任务分配
	for _, assignment := range assignments {
		for _, task := range collab.Tasks {
			if task.ID == assignment.TaskID {
				task.AssignedTo = assignment.AgentID
				task.State = TaskStateAssigned
			}
		}

		// 更新Agent角色
		found := false
		for _, agent := range collab.Agents {
			if agent.AgentID == assignment.AgentID {
				agent.CurrentLoad++
				found = true
				break
			}
		}

		if !found {
			collab.Agents = append(collab.Agents, &AgentRole{
				AgentID:     assignment.AgentID,
				CurrentLoad: 1,
			})
		}
	}

	return nil
}

// executeCollaboration 执行协作
func (m *CollaborationManager) executeCollaboration(ctx context.Context, collab *Collaboration) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if collab.Result != nil {
			collab.Result.Duration = duration
		}

		// 更新统计
		m.mu.Lock()
		if collab.State == CollabStateCompleted {
			m.stats.SuccessfulCollabs++
		} else {
			m.stats.FailedCollabs++
		}
		m.mu.Unlock()
	}()

	// 根据协作类型执行
	var err error
	switch collab.Type {
	case CollabTypeSequential:
		err = m.orchestrator.ExecuteSequential(ctx, collab)
	case CollabTypeParallel:
		err = m.orchestrator.ExecuteParallel(ctx, collab)
	case CollabTypePipeline:
		err = m.orchestrator.ExecutePipeline(ctx, collab)
	case CollabTypeMapReduce:
		err = m.orchestrator.ExecuteMapReduce(ctx, collab)
	default:
		err = m.orchestrator.ExecuteSequential(ctx, collab)
	}

	// 聚合结果
	collab.Result = m.aggregator.Aggregate(collab)

	if err != nil {
		collab.State = CollabStateFailed
		collab.Result.Success = false
		collab.Result.Errors = append(collab.Result.Errors, err.Error())
	} else {
		collab.State = CollabStateCompleted
		collab.Result.Success = true
	}

	collab.UpdatedAt = time.Now()

	// 发布完成事件
	m.eventBus.Publish(&CollabEvent{
		Type:        "completed",
		CollabID:    collab.ID,
		Timestamp:   time.Now(),
		Result:      collab.Result,
	})
}

// GetCollaboration 获取协作
func (m *CollaborationManager) GetCollaboration(id string) (*Collaboration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collab, ok := m.collaborations[id]
	if !ok {
		return nil, fmt.Errorf("协作不存在: %s", id)
	}

	return collab, nil
}

// CancelCollaboration 取消协作
func (m *CollaborationManager) CancelCollaboration(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	collab, ok := m.collaborations[id]
	if !ok {
		return fmt.Errorf("协作不存在: %s", id)
	}

	collab.State = CollabStateCancelled
	collab.UpdatedAt = time.Now()

	// 取消所有运行中的任务
	for _, task := range collab.Tasks {
		if task.State == TaskStateRunning || task.State == TaskStatePending {
			task.State = TaskStateCancelled
		}
	}

	m.eventBus.Publish(&CollabEvent{
		Type:        "cancelled",
		CollabID:    collab.ID,
		Timestamp:   time.Now(),
	})

	return nil
}

// ListCollaborations 列出协作
func (m *CollaborationManager) ListCollaborations() []*Collaboration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Collaboration, 0, len(m.collaborations))
	for _, collab := range m.collaborations {
		list = append(list, collab)
	}
	return list
}

// GetStats 获取统计
func (m *CollaborationManager) GetStats() *ManagerStats {
	return m.stats
}

// Subscribe 订阅事件
func (m *CollaborationManager) Subscribe(handler EventHandler) {
	m.eventBus.Subscribe(handler)
}

// CreateCollabRequest 创建协作请求
type CreateCollabRequest struct {
	Name        string            `json:"name"`
	Type        CollaborationType `json:"type"`
	Description string            `json:"description"`
	Goal        string            `json:"goal"`
	Tasks       []*CollabTask     `json:"tasks"`
	Constraints *Constraints      `json:"constraints"`
	Timeout     time.Duration     `json:"timeout"`
}

// generateID 生成ID
func generateID() string {
	return fmt.Sprintf("collab-%d", time.Now().UnixNano())
}

// CollabEvent 协作事件
type CollabEvent struct {
	Type      string         `json:"type"`
	CollabID  string         `json:"collab_id"`
	Timestamp time.Time      `json:"timestamp"`
	TaskID    string         `json:"task_id,omitempty"`
	AgentID   string         `json:"agent_id,omitempty"`
	Result    *CollabResult  `json:"result,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(event *CollabEvent)

// EventBus 事件总线
type EventBus struct {
	handlers []EventHandler
	mu       sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make([]EventHandler, 0),
	}
}

// Subscribe 订阅
func (b *EventBus) Subscribe(handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// Publish 发布
func (b *EventBus) Publish(event *CollabEvent) {
	b.mu.RLock()
	handlers := make([]EventHandler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

// TaskAssignment 任务分配
type TaskAssignment struct {
	TaskID  string `json:"task_id"`
	AgentID string `json:"agent_id"`
	Score   float64 `json:"score"`
}