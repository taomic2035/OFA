package proto

import (
	context "context"
)

// AgentServiceServer is the server API for AgentService.
type AgentServiceServer interface {
	Connect(AgentService_ConnectServer) error
	SubmitTask(context.Context, *SubmitTaskRequest) (*SubmitTaskResponse, error)
	GetTaskStatus(context.Context, *GetTaskStatusRequest) (*GetTaskStatusResponse, error)
	CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error)
	SubscribeTask(*SubscribeTaskRequest, AgentService_SubscribeTaskServer) error
	RegisterCapabilities(context.Context, *RegisterCapabilitiesRequest) (*RegisterCapabilitiesResponse, error)
	GetCapabilities(context.Context, *GetCapabilitiesRequest) (*GetCapabilitiesResponse, error)
}

// AgentService_ConnectServer is the server stream for Connect
type AgentService_ConnectServer interface {
	Send(*CenterMessage) error
	Recv() (*AgentMessage, error)
	Context() context.Context
}

// AgentService_SubscribeTaskServer is the server stream for SubscribeTask
type AgentService_SubscribeTaskServer interface {
	Send(*TaskEvent) error
	Context() context.Context
}

// UnimplementedAgentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAgentServiceServer struct{}

func (UnimplementedAgentServiceServer) Connect(AgentService_ConnectServer) error {
	return nil
}
func (UnimplementedAgentServiceServer) SubmitTask(context.Context, *SubmitTaskRequest) (*SubmitTaskResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) GetTaskStatus(context.Context, *GetTaskStatusRequest) (*GetTaskStatusResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) CancelTask(context.Context, *CancelTaskRequest) (*CancelTaskResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) SubscribeTask(*SubscribeTaskRequest, AgentService_SubscribeTaskServer) error {
	return nil
}
func (UnimplementedAgentServiceServer) RegisterCapabilities(context.Context, *RegisterCapabilitiesRequest) (*RegisterCapabilitiesResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) GetCapabilities(context.Context, *GetCapabilitiesRequest) (*GetCapabilitiesResponse, error) {
	return nil, nil
}

// RegisterAgentServiceServer registers the AgentService server
func RegisterAgentServiceServer(s interface{}, srv AgentServiceServer) {
	// Placeholder - actual registration happens via grpc.Server
}

// MessageServiceServer is the server API for MessageService.
type MessageServiceServer interface {
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	Broadcast(context.Context, *BroadcastRequest) (*BroadcastResponse, error)
	Multicast(context.Context, *MulticastRequest) (*MulticastResponse, error)
	Subscribe(*SubscribeMessageRequest, MessageService_SubscribeServer) error
}

// MessageService_SubscribeServer is the server stream for Subscribe
type MessageService_SubscribeServer interface {
	Send(*Message) error
	Context() context.Context
}

// UnimplementedMessageServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMessageServiceServer struct{}

func (UnimplementedMessageServiceServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, nil
}
func (UnimplementedMessageServiceServer) Broadcast(context.Context, *BroadcastRequest) (*BroadcastResponse, error) {
	return nil, nil
}
func (UnimplementedMessageServiceServer) Multicast(context.Context, *MulticastRequest) (*MulticastResponse, error) {
	return nil, nil
}
func (UnimplementedMessageServiceServer) Subscribe(*SubscribeMessageRequest, MessageService_SubscribeServer) error {
	return nil
}

// RegisterMessageServiceServer registers the MessageService server
func RegisterMessageServiceServer(s interface{}, srv MessageServiceServer) {
	// Placeholder
}

// ManagementServiceServer is the server API for ManagementService.
type ManagementServiceServer interface {
	ListAgents(context.Context, *ListAgentsRequest) (*ListAgentsResponse, error)
	GetAgent(context.Context, *GetAgentRequest) (*GetAgentResponse, error)
	DeleteAgent(context.Context, *DeleteAgentRequest) (*DeleteAgentResponse, error)
	ListTasks(context.Context, *ListTasksRequest) (*ListTasksResponse, error)
	GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error)
	ListSkills(context.Context, *ListSkillsRequest) (*ListSkillsResponse, error)
	InstallSkill(context.Context, *InstallSkillRequest) (*InstallSkillResponse, error)
	GetSystemInfo(context.Context, *GetSystemInfoRequest) (*GetSystemInfoResponse, error)
	GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error)
}

// UnimplementedManagementServiceServer must be embedded to have forward compatible implementations.
type UnimplementedManagementServiceServer struct{}

func (UnimplementedManagementServiceServer) ListAgents(context.Context, *ListAgentsRequest) (*ListAgentsResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) GetAgent(context.Context, *GetAgentRequest) (*GetAgentResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) DeleteAgent(context.Context, *DeleteAgentRequest) (*DeleteAgentResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) ListTasks(context.Context, *ListTasksRequest) (*ListTasksResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) GetTask(context.Context, *GetTaskRequest) (*GetTaskResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) ListSkills(context.Context, *ListSkillsRequest) (*ListSkillsResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) InstallSkill(context.Context, *InstallSkillRequest) (*InstallSkillResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) GetSystemInfo(context.Context, *GetSystemInfoRequest) (*GetSystemInfoResponse, error) {
	return nil, nil
}
func (UnimplementedManagementServiceServer) GetMetrics(context.Context, *GetMetricsRequest) (*GetMetricsResponse, error) {
	return nil, nil
}

// RegisterManagementServiceServer registers the ManagementService server
func RegisterManagementServiceServer(s interface{}, srv ManagementServiceServer) {
	// Placeholder
}

// ===== Client Interfaces =====

// AgentServiceClient is the client API for AgentService.
type AgentServiceClient interface {
	Connect(ctx context.Context) (AgentService_ConnectClient, error)
	SubmitTask(ctx context.Context, in *SubmitTaskRequest) (*SubmitTaskResponse, error)
	GetTaskStatus(ctx context.Context, in *GetTaskStatusRequest) (*GetTaskStatusResponse, error)
	CancelTask(ctx context.Context, in *CancelTaskRequest) (*CancelTaskResponse, error)
	SubscribeTask(ctx context.Context, in *SubscribeTaskRequest) (AgentService_SubscribeTaskClient, error)
	RegisterCapabilities(ctx context.Context, in *RegisterCapabilitiesRequest) (*RegisterCapabilitiesResponse, error)
	GetCapabilities(ctx context.Context, in *GetCapabilitiesRequest) (*GetCapabilitiesResponse, error)
}

type AgentService_ConnectClient interface {
	Send(*AgentMessage) error
	Recv() (*CenterMessage, error)
	CloseSend() error
	Context() context.Context
}

type AgentService_SubscribeTaskClient interface {
	Recv() (*TaskEvent, error)
	Context() context.Context
}

// NewAgentServiceClient creates a new AgentService client
func NewAgentServiceClient(cc interface{}) AgentServiceClient {
	return &unimplementedAgentServiceClient{}
}

type unimplementedAgentServiceClient struct{}

func (c *unimplementedAgentServiceClient) Connect(ctx context.Context) (AgentService_ConnectClient, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) SubmitTask(ctx context.Context, in *SubmitTaskRequest) (*SubmitTaskResponse, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) GetTaskStatus(ctx context.Context, in *GetTaskStatusRequest) (*GetTaskStatusResponse, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) CancelTask(ctx context.Context, in *CancelTaskRequest) (*CancelTaskResponse, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) SubscribeTask(ctx context.Context, in *SubscribeTaskRequest) (AgentService_SubscribeTaskClient, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) RegisterCapabilities(ctx context.Context, in *RegisterCapabilitiesRequest) (*RegisterCapabilitiesResponse, error) {
	return nil, nil
}
func (c *unimplementedAgentServiceClient) GetCapabilities(ctx context.Context, in *GetCapabilitiesRequest) (*GetCapabilitiesResponse, error) {
	return nil, nil
}

// MessageServiceClient is the client API for MessageService.
type MessageServiceClient interface {
	SendMessage(ctx context.Context, in *SendMessageRequest) (*SendMessageResponse, error)
	Broadcast(ctx context.Context, in *BroadcastRequest) (*BroadcastResponse, error)
	Multicast(ctx context.Context, in *MulticastRequest) (*MulticastResponse, error)
	Subscribe(ctx context.Context, in *SubscribeMessageRequest) (MessageService_SubscribeClient, error)
}

type MessageService_SubscribeClient interface {
	Recv() (*Message, error)
	Context() context.Context
}

// NewMessageServiceClient creates a new MessageService client
func NewMessageServiceClient(cc interface{}) MessageServiceClient {
	return &unimplementedMessageServiceClient{}
}

type unimplementedMessageServiceClient struct{}

func (c *unimplementedMessageServiceClient) SendMessage(ctx context.Context, in *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, nil
}
func (c *unimplementedMessageServiceClient) Broadcast(ctx context.Context, in *BroadcastRequest) (*BroadcastResponse, error) {
	return nil, nil
}
func (c *unimplementedMessageServiceClient) Multicast(ctx context.Context, in *MulticastRequest) (*MulticastResponse, error) {
	return nil, nil
}
func (c *unimplementedMessageServiceClient) Subscribe(ctx context.Context, in *SubscribeMessageRequest) (MessageService_SubscribeClient, error) {
	return nil, nil
}

// ManagementServiceClient is the client API for ManagementService.
type ManagementServiceClient interface {
	ListAgents(ctx context.Context, in *ListAgentsRequest) (*ListAgentsResponse, error)
	GetAgent(ctx context.Context, in *GetAgentRequest) (*GetAgentResponse, error)
	DeleteAgent(ctx context.Context, in *DeleteAgentRequest) (*DeleteAgentResponse, error)
	ListTasks(ctx context.Context, in *ListTasksRequest) (*ListTasksResponse, error)
	GetTask(ctx context.Context, in *GetTaskRequest) (*GetTaskResponse, error)
	ListSkills(ctx context.Context, in *ListSkillsRequest) (*ListSkillsResponse, error)
	InstallSkill(ctx context.Context, in *InstallSkillRequest) (*InstallSkillResponse, error)
	GetSystemInfo(ctx context.Context, in *GetSystemInfoRequest) (*GetSystemInfoResponse, error)
	GetMetrics(ctx context.Context, in *GetMetricsRequest) (*GetMetricsResponse, error)
}

// NewManagementServiceClient creates a new ManagementService client
func NewManagementServiceClient(cc interface{}) ManagementServiceClient {
	return &unimplementedManagementServiceClient{}
}

type unimplementedManagementServiceClient struct{}

func (c *unimplementedManagementServiceClient) ListAgents(ctx context.Context, in *ListAgentsRequest) (*ListAgentsResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) GetAgent(ctx context.Context, in *GetAgentRequest) (*GetAgentResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) DeleteAgent(ctx context.Context, in *DeleteAgentRequest) (*DeleteAgentResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) ListTasks(ctx context.Context, in *ListTasksRequest) (*ListTasksResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) GetTask(ctx context.Context, in *GetTaskRequest) (*GetTaskResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) ListSkills(ctx context.Context, in *ListSkillsRequest) (*ListSkillsResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) InstallSkill(ctx context.Context, in *InstallSkillRequest) (*InstallSkillResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) GetSystemInfo(ctx context.Context, in *GetSystemInfoRequest) (*GetSystemInfoResponse, error) {
	return nil, nil
}
func (c *unimplementedManagementServiceClient) GetMetrics(ctx context.Context, in *GetMetricsRequest) (*GetMetricsResponse, error) {
	return nil, nil
}