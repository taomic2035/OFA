package scheduler

import (
	"testing"

	"github.com/ofa/center/internal/models"
	pb "github.com/ofa/center/proto"
)

func TestCapabilityFirstPolicy(t *testing.T) {
	policy := &CapabilityFirstPolicy{}

	agents := []*AgentInfo{
		{ID: "agent-1", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 5, Capabilities: []models.Capability{{ID: "skill-1"}}},
		{ID: "agent-2", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 2, Capabilities: []models.Capability{{ID: "skill-1"}}},
		{ID: "agent-3", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 10, Capabilities: []models.Capability{{ID: "skill-1"}}},
	}

	task := &models.Task{SkillID: "skill-1"}

	selected := policy.Select(task, agents)

	// Should select agent with lowest load
	if selected != "agent-2" {
		t.Errorf("Expected agent-2 (lowest load), got %s", selected)
	}
}

func TestLoadBalancePolicy(t *testing.T) {
	policy := &LoadBalancePolicy{}

	agents := []*AgentInfo{
		{ID: "agent-1", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 5},
		{ID: "agent-2", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 1},
		{ID: "agent-3", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 3},
	}

	task := &models.Task{SkillID: "skill-1"}

	selected := policy.Select(task, agents)

	// Should select agent with lowest load
	if selected != "agent-2" {
		t.Errorf("Expected agent-2 (lowest load), got %s", selected)
	}
}

func TestLatencyFirstPolicy(t *testing.T) {
	policy := &LatencyFirstPolicy{}

	agents := []*AgentInfo{
		{ID: "agent-1", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, NetworkLat: 100},
		{ID: "agent-2", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, NetworkLat: 20},
		{ID: "agent-3", Status: pb.AgentStatus_AGENT_STATUS_ONLINE, NetworkLat: 50},
	}

	task := &models.Task{SkillID: "skill-1"}

	selected := policy.Select(task, agents)

	// Should select agent with lowest latency
	if selected != "agent-2" {
		t.Errorf("Expected agent-2 (lowest latency), got %s", selected)
	}
}

func TestHybridPolicy(t *testing.T) {
	policy := &HybridPolicy{}

	agents := []*AgentInfo{
		{ID: "agent-1", Type: pb.AgentType_AGENT_TYPE_MOBILE, Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 2, CPUUsage: 30, NetworkLat: 50},
		{ID: "agent-2", Type: pb.AgentType_AGENT_TYPE_FULL, Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 5, CPUUsage: 20, NetworkLat: 30},
		{ID: "agent-3", Type: pb.AgentType_AGENT_TYPE_EDGE, Status: pb.AgentStatus_AGENT_STATUS_ONLINE, Load: 1, CPUUsage: 40, NetworkLat: 100},
	}

	task := &models.Task{SkillID: "skill-1"}

	selected := policy.Select(task, agents)

	// agent-2 has highest score: -50 - 100 - 0.3 + 20 = -130.3
	// agent-3: -10 - 200 - 1 + 10 = -201
	// agent-1: -20 - 150 - 0.5 = -170.5
	// So agent-2 should be selected (has the highest score)
	if selected != "agent-2" {
		t.Errorf("Expected agent-2 (highest hybrid score), got %s", selected)
	}
}

func TestPowerAwarePolicy(t *testing.T) {
	policy := &PowerAwarePolicy{}

	agents := []*AgentInfo{
		{ID: "agent-1", Type: pb.AgentType_AGENT_TYPE_MOBILE, Status: pb.AgentStatus_AGENT_STATUS_ONLINE},
		{ID: "agent-2", Type: pb.AgentType_AGENT_TYPE_FULL, Status: pb.AgentStatus_AGENT_STATUS_ONLINE},
		{ID: "agent-3", Type: pb.AgentType_AGENT_TYPE_IOT, Status: pb.AgentStatus_AGENT_STATUS_ONLINE},
	}

	task := &models.Task{SkillID: "skill-1"}

	selected := policy.Select(task, agents)

	// Should select full agent (best power)
	if selected != "agent-2" {
		t.Errorf("Expected agent-2 (full type), got %s", selected)
	}
}

func TestGetPolicy(t *testing.T) {
	tests := []struct {
		strategy   string
		expectType string
	}{
		{"capability_first", "*scheduler.CapabilityFirstPolicy"},
		{"load_balance", "*scheduler.LoadBalancePolicy"},
		{"latency_first", "*scheduler.LatencyFirstPolicy"},
		{"power_aware", "*scheduler.PowerAwarePolicy"},
		{"hybrid", "*scheduler.HybridPolicy"},
		{"unknown", "*scheduler.HybridPolicy"},
	}

	for _, tt := range tests {
		policy := getPolicy(tt.strategy)
		if policy == nil {
			t.Errorf("Policy for %s should not be nil", tt.strategy)
		}
	}
}