package store

import (
	"context"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"

	pb "github.com/ofa/center/proto"
)

func TestMemoryStore(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("Failed to create memory store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test agent operations
	agent := &models.Agent{
		ID:        "test-agent-1",
		Name:      "Test Agent",
		Type:      pb.AgentType_AGENT_TYPE_FULL,
		Status:    pb.AgentStatus_AGENT_STATUS_ONLINE,
		CreatedAt: time.Now(),
	}

	err = store.SaveAgent(ctx, agent)
	if err != nil {
		t.Fatalf("Failed to save agent: %v", err)
	}

	retrieved, err := store.GetAgent(ctx, "test-agent-1")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("Expected ID %s, got %s", agent.ID, retrieved.ID)
	}

	// Test list agents
	agents, total, err := store.ListAgents(ctx, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to list agents: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 agent, got %d", total)
	}

	if len(agents) != 1 {
		t.Errorf("Expected 1 agent in list, got %d", len(agents))
	}

	// Test task operations
	task := &models.Task{
		ID:          "test-task-1",
		SkillID:     "text.process",
		TargetAgent: "test-agent-1",
		Status:      pb.TaskStatus_TASK_STATUS_PENDING,
		Priority:    5,
		CreatedAt:   time.Now(),
	}

	err = store.SaveTask(ctx, task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	retrievedTask, err := store.GetTask(ctx, "test-task-1")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	// Test delete agent
	err = store.DeleteAgent(ctx, "test-agent-1")
	if err != nil {
		t.Fatalf("Failed to delete agent: %v", err)
	}

	_, err = store.GetAgent(ctx, "test-agent-1")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStoreMessages(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("Failed to create memory store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test message operations
	msg := &models.Message{
		ID:        "test-msg-1",
		FromAgent: "agent-1",
		ToAgent:   "agent-2",
		Action:    "test.action",
		Payload:   []byte(`{"test":"data"}`),
		Timestamp: time.Now(),
	}

	err = store.SaveMessage(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to save message: %v", err)
	}

	messages, err := store.GetPendingMessages(ctx, "agent-2")
	if err != nil {
		t.Fatalf("Failed to get pending messages: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 pending message, got %d", len(messages))
	}

	// Test mark delivered
	err = store.MarkMessageDelivered(ctx, "test-msg-1")
	if err != nil {
		t.Fatalf("Failed to mark message delivered: %v", err)
	}
}

func TestMemoryStoreCache(t *testing.T) {
	store, err := NewMemoryStore()
	if err != nil {
		t.Fatalf("Failed to create memory store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test online status
	err = store.SetAgentOnline(ctx, "agent-1", 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to set agent online: %v", err)
	}

	if !store.IsAgentOnline(ctx, "agent-1") {
		t.Error("Expected agent to be online")
	}

	// Test resources
	resources := &models.ResourceUsage{
		CPUUsage:       45.5,
		MemoryUsage:    60.2,
		BatteryLevel:   80,
		NetworkType:    "wifi",
		NetworkLatency: 10,
	}

	err = store.SetAgentResources(ctx, "agent-1", resources)
	if err != nil {
		t.Fatalf("Failed to set agent resources: %v", err)
	}

	retrieved, err := store.GetAgentResources(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Failed to get agent resources: %v", err)
	}

	if retrieved.CPUUsage != 45.5 {
		t.Errorf("Expected CPU usage 45.5, got %f", retrieved.CPUUsage)
	}
}