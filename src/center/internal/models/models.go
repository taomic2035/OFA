package models

import (
	"time"

	pb "github.com/ofa/center/proto"
)

// Agent represents a registered agent in the system
type Agent struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          pb.AgentType      `json:"type"`
	Status        pb.AgentStatus    `json:"status"`
	DeviceInfo    *DeviceInfo       `json:"device_info,omitempty"`
	Capabilities  []Capability      `json:"capabilities"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Token         string            `json:"token,omitempty"`
	LastSeen      time.Time         `json:"last_seen"`
	LastHeartbeat time.Time         `json:"last_heartbeat,omitempty"` // Alias for LastSeen
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// DeviceInfo represents device hardware information
type DeviceInfo struct {
	OS           string `json:"os"`
	OSVersion    string `json:"os_version"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	TotalMemory  int64  `json:"total_memory"`
	CPUCores     int32  `json:"cpu_cores"`
	Arch         string `json:"arch"`
}

// Capability represents an agent's capability/skill
type Capability struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Category string            `json:"category"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ResourceUsage represents current resource utilization
type ResourceUsage struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	BatteryLevel   int32   `json:"battery_level"`
	NetworkType    string  `json:"network_type"`
	NetworkLatency int32   `json:"network_latency_ms"`
}

// Task represents a task in the system
type Task struct {
	ID            string            `json:"id"`
	ParentTaskID  string            `json:"parent_task_id,omitempty"`
	SkillID       string            `json:"skill_id"`
	Input         []byte            `json:"input"`
	Output        []byte            `json:"output,omitempty"`
	Status        pb.TaskStatus     `json:"status"`
	Priority      int32             `json:"priority"`
	TargetAgent   string            `json:"target_agent,omitempty"`
	SourceAgent   string            `json:"source_agent,omitempty"`
	Error         string            `json:"error,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	StartedAt     time.Time         `json:"started_at,omitempty"`
	CompletedAt   time.Time         `json:"completed_at,omitempty"`
	DurationMS    int64             `json:"duration_ms"`
	TimeoutMS     int64             `json:"timeout_ms"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Message represents an agent-to-agent message
type Message struct {
	ID        string            `json:"id"`
	FromAgent string            `json:"from_agent"`
	ToAgent   string            `json:"to_agent"`
	Type      MessageType       `json:"type"`
	Action    string            `json:"action"`
	Payload   []byte            `json:"payload"`
	Status    MessageStatus     `json:"status"`
	TTL       int32             `json:"ttl"`
	Headers   map[string]string `json:"headers,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Timestamp time.Time         `json:"timestamp"`
}

// MessageType defines message types
type MessageType int

const (
	MessageTypeNormal   MessageType = 0
	MessageTypeBroadcast MessageType = 1
	MessageTypeRequest  MessageType = 2
	MessageTypeResponse MessageType = 3
)

// MessageStatus defines message status
type MessageStatus int

const (
	MessageStatusPending   MessageStatus = 1
	MessageStatusDelivered MessageStatus = 2
	MessageStatusFailed    MessageStatus = 3
)

// AgentConnection represents an active agent connection
type AgentConnection struct {
	AgentID   string
	Stream    pb.AgentService_ConnectServer
	LastSeen  time.Time
	Status    pb.AgentStatus
	Resources *ResourceUsage
}

// TaskQueue represents a queue of pending tasks
type TaskQueue struct {
	Priority int32
	Task     *Task
}