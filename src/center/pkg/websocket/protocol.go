// Package websocket provides WebSocket communication between Center and Agents.
//
// This module implements real-time bidirectional communication for:
// - Agent registration and heartbeat
// - State synchronization and push notifications
// - Task assignment and result reporting
// - Identity and emotion state updates
package websocket

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Agent -> Center messages
	MsgTypeRegister     MessageType = "register"      // Agent registration
	MsgTypeHeartbeat    MessageType = "heartbeat"     // Agent heartbeat
	MsgTypeSyncRequest  MessageType = "sync_request"  // Sync request
	MsgTypeTaskResult   MessageType = "task_result"   // Task execution result
	MsgTypeBehaviorReport MessageType = "behavior"    // Behavior observation report
	MsgTypeDisconnect   MessageType = "disconnect"    // Agent disconnecting

	// Center -> Agent messages
	MsgTypeRegisterAck  MessageType = "register_ack"  // Registration acknowledgment
	MsgTypeStateUpdate  MessageType = "state_update"  // State push notification
	MsgTypeTaskAssign   MessageType = "task_assign"   // Task assignment
	MsgTypeSyncResponse MessageType = "sync_response" // Sync response
	MsgTypeConfigUpdate MessageType = "config_update" // Configuration update
	MsgTypeEmotionUpdate MessageType = "emotion_update" // Emotion state update

	// Error messages
	MsgTypeError        MessageType = "error"         // Error response
)

// WebSocketMessage is the base message structure
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
	MessageID string      `json:"message_id"`
	Payload   interface{} `json:"payload"`
}

// RegisterPayload is the agent registration payload
type RegisterPayload struct {
	AgentID      string            `json:"agent_id"`
	AgentName    string            `json:"agent_name"`
	AgentType    string            `json:"agent_type"`
	DeviceInfo   DeviceInfoPayload `json:"device_info"`
	Capabilities []CapabilityPayload `json:"capabilities"`
	IdentityID   string            `json:"identity_id"` // Bound identity
	Token        string            `json:"token"`       // Authentication token
}

// DeviceInfoPayload contains device information
type DeviceInfoPayload struct {
	OS           string `json:"os"`
	OSVersion    string `json:"os_version"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	TotalMemory  int64  `json:"total_memory"`
	CPUCores     int    `json:"cpu_cores"`
	Arch         string `json:"arch"`
}

// CapabilityPayload describes an agent capability
type CapabilityPayload struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Category string            `json:"category"`
	Metadata map[string]string `json:"metadata"`
}

// RegisterAckPayload is the registration acknowledgment
type RegisterAckPayload struct {
	Success           bool   `json:"success"`
	AgentID           string `json:"agent_id"`
	SessionID         string `json:"session_id"`
	HeartbeatInterval int    `json:"heartbeat_interval_ms"`
	Error             string `json:"error,omitempty"`
	Config            map[string]string `json:"config"`
}

// HeartbeatPayload is the agent heartbeat payload
type HeartbeatPayload struct {
	AgentID     string            `json:"agent_id"`
	Status      string            `json:"status"`
	Resources   ResourcePayload   `json:"resources"`
	PendingTasks int              `json:"pending_tasks"`
	Timestamp   int64             `json:"timestamp"`
}

// ResourcePayload contains agent resource usage
type ResourcePayload struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	BatteryLevel   int     `json:"battery_level"`
	NetworkType    string  `json:"network_type"`
	NetworkLatency int     `json:"network_latency_ms"`
}

// StateUpdatePayload is the state push notification payload
type StateUpdatePayload struct {
	IdentityID string                 `json:"identity_id"`
	UpdateType string                 `json:"update_type"` // identity, emotion, device, sync
	Data       map[string]interface{} `json:"data"`
	Version    int64                  `json:"version"`
	Timestamp  int64                  `json:"timestamp"`
}

// TaskAssignPayload is the task assignment payload
type TaskAssignPayload struct {
	TaskID       string                 `json:"task_id"`
	ParentTaskID string                 `json:"parent_task_id,omitempty"`
	SkillID      string                 `json:"skill_id"`
	Input        map[string]interface{} `json:"input"`
	Priority     int                    `json:"priority"`
	TimeoutMS    int                    `json:"timeout_ms"`
	CreatedAt    int64                  `json:"created_at"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
}

// TaskResultPayload is the task execution result
type TaskResultPayload struct {
	TaskID     string                 `json:"task_id"`
	Status     string                 `json:"status"` // success, failed, cancelled
	Output     map[string]interface{} `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMS int                    `json:"duration_ms"`
	Timestamp  int64                  `json:"timestamp"`
}

// SyncRequestPayload is the synchronization request
type SyncRequestPayload struct {
	IdentityID string                 `json:"identity_id"`
	DataType   string                 `json:"data_type"` // identity, memories, preferences, behaviors
	Version    int64                  `json:"version"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// SyncResponsePayload is the synchronization response
type SyncResponsePayload struct {
	Success    bool                   `json:"success"`
	IdentityID string                 `json:"identity_id"`
	DataType   string                 `json:"data_type"`
	Version    int64                  `json:"version"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Conflict   bool                   `json:"conflict,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// BehaviorReportPayload is the behavior observation report
type BehaviorReportPayload struct {
	IdentityID string                 `json:"identity_id"`
	BehaviorType string               `json:"behavior_type"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   int64                  `json:"timestamp"`
}

// EmotionUpdatePayload is the emotion state update
type EmotionUpdatePayload struct {
	IdentityID      string  `json:"identity_id"`
	DominantEmotion string  `json:"dominant_emotion"`
	Intensity       float64 `json:"intensity"`
	Mood            string  `json:"mood"`
	Timestamp       int64   `json:"timestamp"`
}

// ConfigUpdatePayload is the configuration update
type ConfigUpdatePayload struct {
	AgentID string            `json:"agent_id"`
	Config  map[string]string `json:"config"`
}

// ErrorPayload is the error response payload
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// === Helper Functions ===

// NewMessage creates a new WebSocket message
func NewMessage(msgType MessageType, payload interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      msgType,
		Timestamp: time.Now().UnixMilli(),
		MessageID: generateMessageID(),
		Payload:   payload,
	}
}

// EncodeMessage encodes a message to JSON bytes
func EncodeMessage(msg *WebSocketMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// DecodeMessage decodes JSON bytes to a message
func DecodeMessage(data []byte) (*WebSocketMessage, error) {
	msg := &WebSocketMessage{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// DecodePayload decodes the payload into a specific type
func DecodePayload(msg *WebSocketMessage, target interface{}) error {
	data, err := json.Marshal(msg.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return "msg_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// === Connection State ===

// ConnectionState represents the state of an agent connection
type ConnectionState struct {
	AgentID     string
	SessionID   string
	IdentityID  string
	Status      string // connected, disconnected, timeout
	LastSeen    time.Time
	RegisteredAt time.Time
	Capabilities []CapabilityPayload
	DeviceInfo   DeviceInfoPayload
}

// ConnectionConfig holds WebSocket configuration
type ConnectionConfig struct {
	// Heartbeat settings
	HeartbeatInterval   time.Duration // How often agent sends heartbeat
	HeartbeatTimeout    time.Duration // How long before agent is considered timed out

	// Connection settings
	MaxConnections      int           // Maximum concurrent connections
	ReadBufferSize      int           // WebSocket read buffer size
	WriteBufferSize     int           // WebSocket write buffer size

	// Message settings
	MaxMessageSize      int64         // Maximum message size in bytes
	MessageQueueSize    int           // Message queue size per connection
}

// DefaultConnectionConfig returns default configuration
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		HeartbeatInterval:   30 * time.Second,
		HeartbeatTimeout:    90 * time.Second,
		MaxConnections:      1000,
		ReadBufferSize:      1024,
		WriteBufferSize:     1024,
		MaxMessageSize:      64 * 1024, // 64KB
		MessageQueueSize:    100,
	}
}