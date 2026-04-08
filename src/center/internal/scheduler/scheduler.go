// Package scheduler provides task scheduling functionality.
//
// Deprecated: As of v2.1.0, Center has transformed from a control center to a data center.
// Agents now make autonomous decisions and sync with Center proactively.
// This package is retained for backward compatibility but may be removed in future versions.
// New implementations should use Agent-side decision making with Center sync.
package scheduler

import (
	"sync"

	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/internal/store"

	pb "github.com/ofa/center/proto"
)

// Policy defines the scheduling policy interface
type Policy interface {
	Select(task *models.Task, agents []*AgentInfo) string
}

// AgentInfo contains agent information for scheduling
type AgentInfo struct {
	ID           string
	Type         pb.AgentType
	Status       pb.AgentStatus
	Capabilities []models.Capability
	Load         int
	CPUUsage     float64
	MemoryUsage  float64
	NetworkLat   int32
}

// Scheduler manages task scheduling
type Scheduler struct {
	store   store.StoreInterface
	policy  Policy
	agents  sync.Map // map[string]*AgentInfo
	queue   chan *models.Task
	stopCh  chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(store store.StoreInterface, defaultStrategy string) *Scheduler {
	policy := getPolicy(defaultStrategy)
	return &Scheduler{
		store:  store,
		policy: policy,
		stopCh: make(chan struct{}),
	}
}

// SetTaskQueue sets the task queue
func (s *Scheduler) SetTaskQueue(queue chan *models.Task) {
	s.queue = queue
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// RegisterAgentCapabilities registers an agent's capabilities
func (s *Scheduler) RegisterAgentCapabilities(agentID string, capabilities []models.Capability) {
	// Fetch agent type and status from store
	agent, err := s.store.GetAgent(nil, agentID)
	if err != nil {
		return
	}

	info := &AgentInfo{
		ID:           agentID,
		Type:         agent.Type,
		Status:       agent.Status,
		Capabilities: capabilities,
		Load:         0,
	}
	s.agents.Store(agentID, info)
}

// UnregisterAgent removes an agent from scheduling
func (s *Scheduler) UnregisterAgent(agentID string) {
	s.agents.Delete(agentID)
}

// UpdateAgentLoad updates an agent's current load
func (s *Scheduler) UpdateAgentLoad(agentID string, pendingTasks int) {
	info, ok := s.agents.Load(agentID)
	if ok {
		agentInfo := info.(*AgentInfo)
		agentInfo.Load = pendingTasks
	}
}

// UpdateAgentResources updates an agent's resources
func (s *Scheduler) UpdateAgentResources(agentID string, cpu, memory float64, latency int32) {
	info, ok := s.agents.Load(agentID)
	if ok {
		agentInfo := info.(*AgentInfo)
		agentInfo.CPUUsage = cpu
		agentInfo.MemoryUsage = memory
		agentInfo.NetworkLat = latency
	}
}

// Schedule selects the best agent for a task
func (s *Scheduler) Schedule(task *models.Task) string {
	// Get available agents
	agents := s.getAvailableAgents(task.SkillID)

	if len(agents) == 0 {
		return ""
	}

	// If target agent is specified, use it
	if task.TargetAgent != "" {
		if s.hasAgentCapability(task.TargetAgent, task.SkillID) {
			return task.TargetAgent
		}
		return ""
	}

	// Use policy to select agent
	return s.policy.Select(task, agents)
}

// getAvailableAgents returns agents that have the required capability
func (s *Scheduler) getAvailableAgents(skillID string) []*AgentInfo {
	var agents []*AgentInfo

	s.agents.Range(func(key, value interface{}) bool {
		info := value.(*AgentInfo)

		// Check if agent is online
		if info.Status != pb.AgentStatus_AGENT_STATUS_ONLINE &&
			info.Status != pb.AgentStatus_AGENT_STATUS_IDLE {
			return true
		}

		// Check if agent has the capability
		for _, cap := range info.Capabilities {
			if cap.ID == skillID {
				agents = append(agents, info)
				break
			}
		}

		return true
	})

	return agents
}

// hasAgentCapability checks if an agent has a specific capability
func (s *Scheduler) hasAgentCapability(agentID, skillID string) bool {
	info, ok := s.agents.Load(agentID)
	if !ok {
		return false
	}

	agentInfo := info.(*AgentInfo)
	for _, cap := range agentInfo.Capabilities {
		if cap.ID == skillID {
			return true
		}
	}
	return false
}

// ===== Scheduling Policies =====

// CapabilityFirstPolicy selects agents by capability match score
type CapabilityFirstPolicy struct{}

func (p *CapabilityFirstPolicy) Select(task *models.Task, agents []*AgentInfo) string {
	// For capability-first, prefer agents with exact match and lower load
	bestAgent := ""
	minLoad := int(1 << 30)

	for _, agent := range agents {
		if agent.Load < minLoad {
			minLoad = agent.Load
			bestAgent = agent.ID
		}
	}

	return bestAgent
}

// LoadBalancePolicy distributes tasks evenly
type LoadBalancePolicy struct{}

func (p *LoadBalancePolicy) Select(task *models.Task, agents []*AgentInfo) string {
	// Select agent with lowest load
	bestAgent := ""
	minLoad := int(1 << 30)

	for _, agent := range agents {
		if agent.Load < minLoad {
			minLoad = agent.Load
			bestAgent = agent.ID
		}
	}

	return bestAgent
}

// LatencyFirstPolicy selects agents with lowest latency
type LatencyFirstPolicy struct{}

func (p *LatencyFirstPolicy) Select(task *models.Task, agents []*AgentInfo) string {
	bestAgent := ""
	minLatency := int32(1 << 30)

	for _, agent := range agents {
		if agent.NetworkLat < minLatency {
			minLatency = agent.NetworkLat
			bestAgent = agent.ID
		}
	}

	return bestAgent
}

// PowerAwarePolicy selects agents with better battery/power state
type PowerAwarePolicy struct{}

func (p *PowerAwarePolicy) Select(task *models.Task, agents []*AgentInfo) string {
	// Prefer agents with better power state
	// Priority: FULL > EDGE > LITE > MOBILE > IOT
	typePower := map[pb.AgentType]int{
		pb.AgentType_AGENT_TYPE_FULL:   5,
		pb.AgentType_AGENT_TYPE_EDGE:   4,
		pb.AgentType_AGENT_TYPE_LITE:   3,
		pb.AgentType_AGENT_TYPE_MOBILE: 2,
		pb.AgentType_AGENT_TYPE_IOT:    1,
	}

	bestAgent := ""
	bestPower := -1

	for _, agent := range agents {
		power := typePower[agent.Type]
		if power > bestPower {
			bestPower = power
			bestAgent = agent.ID
		}
	}

	if bestAgent == "" && len(agents) > 0 {
		bestAgent = agents[0].ID
	}

	return bestAgent
}

// HybridPolicy combines multiple factors
type HybridPolicy struct{}

func (p *HybridPolicy) Select(task *models.Task, agents []*AgentInfo) string {
	bestAgent := ""
	bestScore := float64(-1 << 30)

	for _, agent := range agents {
		score := p.calculateScore(agent)
		if score > bestScore {
			bestScore = score
			bestAgent = agent.ID
		}
	}

	return bestAgent
}

func (p *HybridPolicy) calculateScore(agent *AgentInfo) float64 {
	// Score = -load * 10 - cpu * 5 - latency/100
	score := float64(-agent.Load * 10)
	score -= agent.CPUUsage * 5
	score -= float64(agent.NetworkLat) / 100

	// Bonus for full/edge agents (more capable)
	if agent.Type == pb.AgentType_AGENT_TYPE_FULL {
		score += 20
	} else if agent.Type == pb.AgentType_AGENT_TYPE_EDGE {
		score += 10
	}

	return score
}

func getPolicy(strategy string) Policy {
	switch strategy {
	case "capability_first":
		return &CapabilityFirstPolicy{}
	case "load_balance":
		return &LoadBalancePolicy{}
	case "latency_first":
		return &LatencyFirstPolicy{}
	case "power_aware":
		return &PowerAwarePolicy{}
	default:
		return &HybridPolicy{}
	}
}