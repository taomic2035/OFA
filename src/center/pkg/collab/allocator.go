// Package collab - 任务分配器
package collab

import (
	"fmt"
	"sort"
	"sync"
)

// TaskAllocator 任务分配器
type TaskAllocator struct {
	agentRegistry *AgentRegistry
	policy        AllocationPolicy
	mu            sync.RWMutex
}

// AllocationPolicy 分配策略
type AllocationPolicy string

const (
	PolicyCapability  AllocationPolicy = "capability"  // 能力优先
	PolicyLoadBalance AllocationPolicy = "load_balance" // 负载均衡
	PolicyLatency     AllocationPolicy = "latency"     // 延迟优先
	PolicyCost        AllocationPolicy = "cost"        // 成本优先
	PolicyScore       AllocationPolicy = "score"       // 综合评分
)

// AgentRegistry Agent注册表
type AgentRegistry struct {
	agents map[string]*AgentInfo
	mu     sync.RWMutex
}

// AgentInfo Agent信息
type AgentInfo struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Capabilities []string          `json:"capabilities"`
	Load         int               `json:"load"`
	MaxLoad      int               `json:"max_load"`
	Score        float64           `json:"score"`
	Latency      int64             `json:"latency_ms"`
	Cost         float64           `json:"cost"`
	Status       string            `json:"status"`
	Metadata     map[string]string `json:"metadata"`
}

// NewAgentRegistry 创建Agent注册表
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentInfo),
	}
}

// Register 注册Agent
func (r *AgentRegistry) Register(info *AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[info.ID] = info
}

// Unregister 注销Agent
func (r *AgentRegistry) Unregister(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, agentID)
}

// Get 获取Agent信息
func (r *AgentRegistry) Get(agentID string) (*AgentInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.agents[agentID]
	return info, ok
}

// List 列出所有Agent
func (r *AgentRegistry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*AgentInfo, 0, len(r.agents))
	for _, info := range r.agents {
		list = append(list, info)
	}
	return list
}

// FindByCapability 按能力查找
func (r *AgentRegistry) FindByCapability(capability string) []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*AgentInfo, 0)
	for _, info := range r.agents {
		for _, cap := range info.Capabilities {
			if cap == capability {
				list = append(list, info)
				break
			}
		}
	}
	return list
}

// NewTaskAllocator 创建任务分配器
func NewTaskAllocator() *TaskAllocator {
	return &TaskAllocator{
		agentRegistry: NewAgentRegistry(),
		policy:        PolicyScore,
	}
}

// SetPolicy 设置分配策略
func (a *TaskAllocator) SetPolicy(policy AllocationPolicy) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.policy = policy
}

// SetAgentRegistry 设置Agent注册表
func (a *TaskAllocator) SetAgentRegistry(registry *AgentRegistry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.agentRegistry = registry
}

// Allocate 分配任务
func (a *TaskAllocator) Allocate(collab *Collaboration) ([]*TaskAssignment, error) {
	a.mu.RLock()
	registry := a.agentRegistry
	policy := a.policy
	a.mu.RUnlock()

	assignments := make([]*TaskAssignment, 0)

	for _, task := range collab.Tasks {
		// 检查是否已分配
		if task.AssignedTo != "" {
			assignments = append(assignments, &TaskAssignment{
				TaskID:  task.ID,
				AgentID: task.AssignedTo,
				Score:   1.0,
			})
			continue
		}

		// 查找合适的Agent
		agent, score, err := a.findBestAgent(registry, task, collab.Constraints, policy)
		if err != nil {
			return nil, fmt.Errorf("任务 %s 分配失败: %w", task.ID, err)
		}

		assignments = append(assignments, &TaskAssignment{
			TaskID:  task.ID,
			AgentID: agent.ID,
			Score:   score,
		})

		// 更新Agent负载
		registry.mu.Lock()
		agent.Load++
		registry.mu.Unlock()
	}

	return assignments, nil
}

// findBestAgent 查找最佳Agent
func (a *TaskAllocator) findBestAgent(registry *AgentRegistry, task *CollabTask, constraints *Constraints, policy AllocationPolicy) (*AgentInfo, float64, error) {
	// 查找具备技能的Agent
	candidates := registry.FindByCapability(task.SkillID)

	if len(candidates) == 0 {
		return nil, 0, fmt.Errorf("没有Agent具备技能 %s", task.SkillID)
	}

	// 应用约束过滤
	candidates = a.applyConstraints(candidates, constraints)

	if len(candidates) == 0 {
		return nil, 0, fmt.Errorf("没有Agent满足约束条件")
	}

	// 根据策略排序
	candidates = a.sortByPolicy(candidates, policy)

	// 选择最佳Agent
	for _, agent := range candidates {
		if agent.Load < agent.MaxLoad {
			return agent, a.calculateScore(agent, policy), nil
		}
	}

	return nil, 0, fmt.Errorf("所有Agent都已达到最大负载")
}

// applyConstraints 应用约束
func (a *TaskAllocator) applyConstraints(candidates []*AgentInfo, constraints *Constraints) []*AgentInfo {
	if constraints == nil {
		return candidates
	}

	result := make([]*AgentInfo, 0)

	for _, agent := range candidates {
		// 排除指定Agent
		excluded := false
		for _, excludeID := range constraints.ExcludeAgents {
			if agent.ID == excludeID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// 检查必需技能
		if len(constraints.RequiredSkills) > 0 {
			hasAll := true
			for _, skill := range constraints.RequiredSkills {
				found := false
				for _, cap := range agent.Capabilities {
					if cap == skill {
						found = true
						break
					}
				}
				if !found {
					hasAll = false
					break
				}
			}
			if !hasAll {
				continue
			}
		}

		// 检查预算
		if constraints.Budget > 0 && agent.Cost > constraints.Budget {
			continue
		}

		result = append(result, agent)
	}

	return result
}

// sortByPolicy 按策略排序
func (a *TaskAllocator) sortByPolicy(candidates []*AgentInfo, policy AllocationPolicy) []*AgentInfo {
	sorted := make([]*AgentInfo, len(candidates))
	copy(sorted, candidates)

	switch policy {
	case PolicyCapability:
		// 能力数量优先
		sort.Slice(sorted, func(i, j int) bool {
			return len(sorted[i].Capabilities) > len(sorted[j].Capabilities)
		})

	case PolicyLoadBalance:
		// 负载最低优先
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Load < sorted[j].Load
		})

	case PolicyLatency:
		// 延迟最低优先
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Latency < sorted[j].Latency
		})

	case PolicyCost:
		// 成本最低优先
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Cost < sorted[j].Cost
		})

	case PolicyScore:
		// 综合评分最高优先
		sort.Slice(sorted, func(i, j int) bool {
			return a.calculateScore(sorted[i], policy) > a.calculateScore(sorted[j], policy)
		})
	}

	return sorted
}

// calculateScore 计算评分
func (a *TaskAllocator) calculateScore(agent *AgentInfo, policy AllocationPolicy) float64 {
	switch policy {
	case PolicyCapability:
		return float64(len(agent.Capabilities))

	case PolicyLoadBalance:
		if agent.MaxLoad > 0 {
			return 1.0 - float64(agent.Load)/float64(agent.MaxLoad)
		}
		return 1.0

	case PolicyLatency:
		if agent.Latency > 0 {
			return 1.0 / float64(agent.Latency)
		}
		return 1.0

	case PolicyCost:
		if agent.Cost > 0 {
			return 1.0 / agent.Cost
		}
		return 1.0

	case PolicyScore:
		// 综合评分
		loadScore := 1.0 - float64(agent.Load)/float64(agent.MaxLoad+1)
		capScore := float64(len(agent.Capabilities)) / 10.0
		latScore := 1.0 / float64(agent.Latency+1)
		return (loadScore * 0.3 + capScore * 0.3 + latScore * 0.2 + agent.Score * 0.2)

	default:
		return agent.Score
	}
}

// Reallocate 重新分配
func (a *TaskAllocator) Reallocate(taskID string, reason string) (*TaskAssignment, error) {
	a.mu.RLock()
	registry := a.agentRegistry
	a.mu.RUnlock()

	// 查找负载最低的Agent
	candidates := registry.List()
	if len(candidates) == 0 {
		return nil, fmt.Errorf("没有可用Agent")
	}

	// 按负载排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Load < candidates[j].Load
	})

	agent := candidates[0]
	return &TaskAssignment{
		TaskID:  taskID,
		AgentID: agent.ID,
		Score:   a.calculateScore(agent, PolicyLoadBalance),
	}, nil
}

// GetAgentLoad 获取Agent负载
func (a *TaskAllocator) GetAgentLoad(agentID string) int {
	a.mu.RLock()
	registry := a.agentRegistry
	a.mu.RUnlock()

	info, ok := registry.Get(agentID)
	if !ok {
		return 0
	}
	return info.Load
}

// UpdateAgentScore 更新Agent评分
func (a *TaskAllocator) UpdateAgentScore(agentID string, score float64) {
	a.mu.RLock()
	registry := a.agentRegistry
	a.mu.RUnlock()

	registry.mu.Lock()
	defer registry.mu.Unlock()
	if info, ok := registry.agents[agentID]; ok {
		info.Score = score
	}
}