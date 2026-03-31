package proto

// This file provides placeholder types until protobuf generation is complete.
// After running gen_proto.sh, the actual generated files will replace these.

// AgentType enum placeholder
type AgentType int32

const (
	AgentType_AGENT_TYPE_UNKNOWN AgentType = 0
	AgentType_AGENT_TYPE_FULL    AgentType = 1
	AgentType_AGENT_TYPE_MOBILE  AgentType = 2
	AgentType_AGENT_TYPE_LITE    AgentType = 3
	AgentType_AGENT_TYPE_IOT     AgentType = 4
	AgentType_AGENT_TYPE_EDGE    AgentType = 5
)

// AgentStatus enum placeholder
type AgentStatus int32

const (
	AgentStatus_AGENT_STATUS_UNKNOWN AgentStatus = 0
	AgentStatus_AGENT_STATUS_ONLINE  AgentStatus = 1
	AgentStatus_AGENT_STATUS_BUSY    AgentStatus = 2
	AgentStatus_AGENT_STATUS_IDLE    AgentStatus = 3
	AgentStatus_AGENT_STATUS_OFFLINE AgentStatus = 4
)

// TaskStatus enum placeholder
type TaskStatus int32

const (
	TaskStatus_TASK_STATUS_UNKNOWN   TaskStatus = 0
	TaskStatus_TASK_STATUS_PENDING   TaskStatus = 1
	TaskStatus_TASK_STATUS_RUNNING   TaskStatus = 2
	TaskStatus_TASK_STATUS_COMPLETED TaskStatus = 3
	TaskStatus_TASK_STATUS_FAILED    TaskStatus = 4
	TaskStatus_TASK_STATUS_CANCELLED TaskStatus = 5
	TaskStatus_TASK_STATUS_TIMEOUT   TaskStatus = 6
)

func (t TaskStatus) String() string {
	switch t {
	case TaskStatus_TASK_STATUS_PENDING:
		return "PENDING"
	case TaskStatus_TASK_STATUS_RUNNING:
		return "RUNNING"
	case TaskStatus_TASK_STATUS_COMPLETED:
		return "COMPLETED"
	case TaskStatus_TASK_STATUS_FAILED:
		return "FAILED"
	case TaskStatus_TASK_STATUS_CANCELLED:
		return "CANCELLED"
	case TaskStatus_TASK_STATUS_TIMEOUT:
		return "TIMEOUT"
	default:
		return "UNKNOWN"
	}
}

// DeviceInfo message placeholder
type DeviceInfo struct {
	Os           string
	OsVersion    string
	Model        string
	Manufacturer string
	TotalMemory  int64
	CpuCores     int32
	Arch         string
}

// Capability message placeholder
type Capability struct {
	Id       string
	Name     string
	Version  string
	Category string
	Metadata map[string]string
}

// ResourceUsage message placeholder
type ResourceUsage struct {
	CpuUsage       float64
	MemoryUsage    float64
	BatteryLevel   int32
	NetworkType    string
	NetworkLatencyMs int32
}

// AgentStatusInfo message placeholder
type AgentStatusInfo struct {
	AgentId       string
	Status        AgentStatus
	Resources     *ResourceUsage
	Capabilities  []*Capability
	LastSeen      int64
}

// Task message placeholder
type Task struct {
	TaskId       string
	ParentTaskId string
	SkillId      string
	Input        []byte
	Output       []byte
	Status       TaskStatus
	Priority     int32
	TargetAgent  string
	SourceAgent  string
	Error        string
	CreatedAt    int64
	StartedAt    int64
	CompletedAt  int64
	DurationMs   int64
	TimeoutMs    int64
	Metadata     map[string]string
}

// Message message placeholder
type Message struct {
	MsgId     string
	FromAgent string
	ToAgent   string
	Action    string
	Payload   []byte
	Timestamp int64
	Ttl       int32
	Headers   map[string]string
}

// RegisterRequest message placeholder
type RegisterRequest struct {
	AgentId      string
	Name         string
	Type         AgentType
	DeviceInfo   *DeviceInfo
	Capabilities []*Capability
	Metadata     map[string]string
}

// RegisterResponse message placeholder
type RegisterResponse struct {
	Success            bool
	AgentId            string
	Error              string
	Token              string
	HeartbeatIntervalMs int64
	Config             map[string]string
}

// HeartbeatRequest message placeholder
type HeartbeatRequest struct {
	AgentId      string
	Status       AgentStatus
	Resources    *ResourceUsage
	PendingTasks int32
}

// TaskAssignment message placeholder
type TaskAssignment struct {
	TaskId       string
	ParentTaskId string
	SkillId      string
	Input        []byte
	Priority     int32
	TimeoutMs    int64
	CreatedAt    int64
	Metadata     map[string]string
}

// TaskResult message placeholder
type TaskResult struct {
	TaskId     string
	Status     TaskStatus
	Output     []byte
	Error      string
	DurationMs int64
	Metadata   map[string]string
}

// AgentEvent message placeholder
type AgentEvent struct {
	EventType string
	Data      []byte
	Timestamp int64
}

// ConfigUpdate message placeholder
type ConfigUpdate struct {
	Config map[string]string
	Version int64
}

// SubmitTaskRequest message placeholder
type SubmitTaskRequest struct {
	SkillId      string
	Input        []byte
	TargetAgent  string
	Priority     int32
	TimeoutMs    int64
	CallbackUrl  string
	Metadata     map[string]string
}

// SubmitTaskResponse message placeholder
type SubmitTaskResponse struct {
	Success bool
	TaskId  string
	Error   string
}

// GetTaskStatusRequest message placeholder
type GetTaskStatusRequest struct {
	TaskId string
}

// GetTaskStatusResponse message placeholder
type GetTaskStatusResponse struct {
	Success bool
	Task    *Task
	Error   string
}

// CancelTaskRequest message placeholder
type CancelTaskRequest struct {
	TaskId string
	Reason string
}

// CancelTaskResponse message placeholder
type CancelTaskResponse struct {
	Success bool
	Error   string
}

// SubscribeTaskRequest message placeholder
type SubscribeTaskRequest struct {
	TaskId string
}

// TaskEvent message placeholder
type TaskEvent struct {
	TaskId     string
	EventType  string
	Task       *Task
	Timestamp  int64
}

// RegisterCapabilitiesRequest message placeholder
type RegisterCapabilitiesRequest struct {
	AgentId      string
	Capabilities []*Capability
}

// RegisterCapabilitiesResponse message placeholder
type RegisterCapabilitiesResponse struct {
	Success bool
	Error   string
}

// GetCapabilitiesRequest message placeholder
type GetCapabilitiesRequest struct {
	AgentId string
}

// GetCapabilitiesResponse message placeholder
type GetCapabilitiesResponse struct {
	Capabilities []*Capability
}

// SendMessageRequest message placeholder
type SendMessageRequest struct {
	Message    *Message
	RequireAck bool
	TimeoutMs  int64
}

// SendMessageResponse message placeholder
type SendMessageResponse struct {
	Success bool
	MsgId   string
	Error   string
}

// BroadcastRequest message placeholder
type BroadcastRequest struct {
	FromAgent string
	Action    string
	Payload   []byte
	Ttl       int32
}

// BroadcastResponse message placeholder
type BroadcastResponse struct {
	Success        bool
	DeliveredCount int32
	Error          string
}

// MulticastRequest message placeholder
type MulticastRequest struct {
	FromAgent string
	ToAgents  []string
	Action    string
	Payload   []byte
	Ttl       int32
}

// MulticastResponse message placeholder
type MulticastResponse struct {
	Success        bool
	DeliveredCount int32
	Error          string
}

// SubscribeMessageRequest message placeholder
type SubscribeMessageRequest struct {
	AgentId string
	Actions []string
}

// MessageResponse message placeholder
type MessageResponse struct {
	MsgId   string
	Success bool
	Error   string
}

// ListAgentsRequest message placeholder
type ListAgentsRequest struct {
	Type     AgentType
	Status   AgentStatus
	Page     int32
	PageSize int32
}

// ListAgentsResponse message placeholder
type ListAgentsResponse struct {
	Agents   []*AgentStatusInfo
	Total    int32
	Page     int32
	PageSize int32
}

// GetAgentRequest message placeholder
type GetAgentRequest struct {
	AgentId string
}

// GetAgentResponse message placeholder
type GetAgentResponse struct {
	Success bool
	Agent   *AgentStatusInfo
	Error   string
}

// DeleteAgentRequest message placeholder
type DeleteAgentRequest struct {
	AgentId string
}

// DeleteAgentResponse message placeholder
type DeleteAgentResponse struct {
	Success bool
	Error   string
}

// ListTasksRequest message placeholder
type ListTasksRequest struct {
	Status   TaskStatus
	AgentId  string
	Page     int32
	PageSize int32
}

// ListTasksResponse message placeholder
type ListTasksResponse struct {
	Tasks    []*Task
	Total    int32
	Page     int32
	PageSize int32
}

// GetTaskRequest message placeholder
type GetTaskRequest struct {
	TaskId string
}

// GetTaskResponse message placeholder
type GetTaskResponse struct {
	Success bool
	Task    *Task
	Error   string
}

// ListSkillsRequest message placeholder
type ListSkillsRequest struct {
	Category string
}

// ListSkillsResponse message placeholder
type ListSkillsResponse struct {
	Skills []*Capability
}

// InstallSkillRequest message placeholder
type InstallSkillRequest struct {
	AgentId  string
	SkillId  string
	Version  string
}

// InstallSkillResponse message placeholder
type InstallSkillResponse struct {
	Success bool
	Error   string
}

// GetSystemInfoRequest message placeholder
type GetSystemInfoRequest struct{}

// GetSystemInfoResponse message placeholder
type GetSystemInfoResponse struct {
	Version      string
	UptimeSeconds int64
	AgentCount   int32
	TaskCount    int32
	Info         map[string]string
}

// GetMetricsRequest message placeholder
type GetMetricsRequest struct {
	Names []string
}

// GetMetricsResponse message placeholder
type GetMetricsResponse struct {
	Metrics map[string]float64
}

// AgentMessage message placeholder (stream)
type AgentMessage struct {
	MsgId     string
	Timestamp int64
	Register  *RegisterRequest
	Heartbeat *HeartbeatRequest
	TaskResult *TaskResult
	Event     *AgentEvent
	MessageResponse *MessageResponse
}

// CenterMessage message placeholder (stream)
type CenterMessage struct {
	MsgId      string
	Timestamp  int64
	Register   *RegisterResponse
	Task       *TaskAssignment
	Config     *ConfigUpdate
	Message    *Message
	CancelTask *CancelTaskRequest
}