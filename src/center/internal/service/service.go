package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ofa/center/internal/config"
	identity "github.com/ofa/center/internal/identity"
	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/internal/scheduler"
	"github.com/ofa/center/internal/store"

	pb "github.com/ofa/center/proto"
)

// CenterService is the main service orchestrating all components
type CenterService struct {
	cfg      *config.Config
	store    *store.Store
	scheduler *scheduler.Scheduler
	identity *identity.Service // v1.2.0: Identity Service

	// Active agent connections
	connections sync.Map // map[string]*models.AgentConnection

	// Channels for internal communication
	taskQueue   chan *models.Task
	messageChan chan *models.Message

	ctx    context.Context
	cancel context.CancelFunc
}

// NewCenterService creates a new Center service
func NewCenterService(ctx context.Context, cfg *config.Config) (*CenterService, error) {
	store, err := store.NewStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	service := &CenterService{
		cfg:        cfg,
		store:      store,
		taskQueue:  make(chan *models.Task, cfg.Scheduler.MaxConcurrent),
		messageChan: make(chan *models.Message, 1000),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize scheduler
	service.scheduler = scheduler.NewScheduler(store, cfg.Scheduler.DefaultStrategy)
	service.scheduler.SetTaskQueue(service.taskQueue)

	// v1.2.0: Initialize Identity Service
	identityStore := identity.NewMemoryStore() // Use memory store for now, can be FileStore
	service.identity = identity.NewService(identityStore)

	// Start background workers
	go service.taskDispatcher()
	go service.messageDispatcher()
	go service.agentMonitor()

	return service, nil
}

// Close closes the service
func (s *CenterService) Close() {
	s.cancel()
	s.scheduler.Stop()
	s.store.Close()
}

// ===== Accessor Methods =====

// GetStore returns the store instance
func (s *CenterService) GetStore() *store.Store {
	return s.store
}

// GetScheduler returns the scheduler instance
func (s *CenterService) GetScheduler() *scheduler.Scheduler {
	return s.scheduler
}

// GetIdentity returns the identity service instance (v1.2.0)
func (s *CenterService) GetIdentity() *identity.Service {
	return s.identity
}

// GetTaskQueue returns the task queue channel
func (s *CenterService) GetTaskQueue() chan *models.Task {
	return s.taskQueue
}

// GetConnections returns the connections map
func (s *CenterService) GetConnections() *sync.Map {
	return &s.connections
}

// ===== Agent Management =====

// RegisterAgent registers a new agent or reconnects an existing one
func (s *CenterService) RegisterAgent(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	agentID := req.AgentId
	if agentID == "" {
		agentID = generateID()
	}

	token := generateToken()

	agent := &models.Agent{
		ID:           agentID,
		Name:         req.Name,
		Type:         req.Type,
		Status:       pb.AgentStatus_AGENT_STATUS_ONLINE,
		DeviceInfo:   convertDeviceInfo(req.DeviceInfo),
		Capabilities: convertCapabilities(req.Capabilities),
		Metadata:     req.Metadata,
		Token:        token,
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := s.store.SaveAgent(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to save agent: %w", err)
	}

	// Set online in Redis
	s.store.SetAgentOnline(ctx, agentID, s.cfg.Agent.HeartbeatTimeout*2)

	// Register capabilities with scheduler
	s.scheduler.RegisterAgentCapabilities(agentID, agent.Capabilities)

	log.Printf("Agent registered: %s (%s)", agentID, req.Name)

	return &pb.RegisterResponse{
		Success:           true,
		AgentId:           agentID,
		Token:             token,
		HeartbeatIntervalMs: int64(s.cfg.Agent.HeartbeatInterval.Milliseconds()),
		Config:            getDefaultConfig(),
	}, nil
}

// HandleHeartbeat processes agent heartbeat
func (s *CenterService) HandleHeartbeat(ctx context.Context, req *pb.HeartbeatRequest) error {
	// Update agent status
	agent, err := s.store.GetAgent(ctx, req.AgentId)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	agent.Status = req.Status
	agent.LastSeen = time.Now()

	if err := s.store.SaveAgent(ctx, agent); err != nil {
		return err
	}

	// Update online status in Redis
	s.store.SetAgentOnline(ctx, req.AgentId, s.cfg.Agent.HeartbeatTimeout*2)

	// Update resources in Redis
	if req.Resources != nil {
		resources := convertResourceUsage(req.Resources)
		s.store.SetAgentResources(ctx, req.AgentId, resources)
	}

	// Update connection info
	conn, ok := s.connections.Load(req.AgentId)
	if ok {
		connection := conn.(*models.AgentConnection)
		connection.LastSeen = time.Now()
		connection.Status = req.Status
		connection.Resources = convertResourceUsage(req.Resources)
	}

	// Update scheduler agent load
	s.scheduler.UpdateAgentLoad(req.AgentId, int(req.PendingTasks))

	return nil
}

// AddConnection adds an agent stream connection
func (s *CenterService) AddConnection(agentID string, stream pb.AgentService_ConnectServer) {
	conn := &models.AgentConnection{
		AgentID:  agentID,
		Stream:   stream,
		LastSeen: time.Now(),
		Status:   pb.AgentStatus_AGENT_STATUS_ONLINE,
	}
	s.connections.Store(agentID, conn)
}

// RemoveConnection removes an agent stream connection
func (s *CenterService) RemoveConnection(agentID string) {
	s.connections.Delete(agentID)
}

// ===== Task Management =====

// SubmitTask submits a new task
func (s *CenterService) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error) {
	taskID := generateID()

	task := &models.Task{
		ID:          taskID,
		SkillID:     req.SkillId,
		Input:       req.Input,
		TargetAgent: req.TargetAgent,
		Priority:    req.Priority,
		TimeoutMS:   req.TimeoutMs,
		Metadata:    req.Metadata,
		Status:      pb.TaskStatus_TASK_STATUS_PENDING,
		CreatedAt:   time.Now(),
	}

	// Save to database
	if err := s.store.SaveTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// Enqueue for scheduling
	s.taskQueue <- task

	log.Printf("Task submitted: %s (skill: %s)", taskID, req.SkillId)

	return &pb.SubmitTaskResponse{
		Success: true,
		TaskId:  taskID,
	}, nil
}

// GetTaskStatus retrieves task status
func (s *CenterService) GetTaskStatus(ctx context.Context, req *pb.GetTaskStatusRequest) (*pb.GetTaskStatusResponse, error) {
	task, err := s.store.GetTask(ctx, req.TaskId)
	if err != nil {
		return &pb.GetTaskStatusResponse{
			Success: false,
			Error:   "Task not found",
		}, nil
	}

	return &pb.GetTaskStatusResponse{
		Success: true,
		Task:    convertTaskToProto(task),
	}, nil
}

// CancelTask cancels a task
func (s *CenterService) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	task, err := s.store.GetTask(ctx, req.TaskId)
	if err != nil {
		return &pb.CancelTaskResponse{
			Success: false,
			Error:   "Task not found",
		}, nil
	}

	// Check if task can be cancelled
	if task.Status == pb.TaskStatus_TASK_STATUS_COMPLETED ||
		task.Status == pb.TaskStatus_TASK_STATUS_FAILED ||
		task.Status == pb.TaskStatus_TASK_STATUS_CANCELLED {
		return &pb.CancelTaskResponse{
			Success: false,
			Error:   "Task already finished",
		}, nil
	}

	// Update status
	task.Status = pb.TaskStatus_TASK_STATUS_CANCELLED
	task.Error = req.Reason
	task.CompletedAt = time.Now()

	if err := s.store.SaveTask(ctx, task); err != nil {
		return nil, err
	}

	// Send cancel request to agent if running
	if task.TargetAgent != "" {
		s.SendCancelTaskRequest(task.TargetAgent, req)
	}

	return &pb.CancelTaskResponse{Success: true}, nil
}

// HandleTaskResult processes task execution result
func (s *CenterService) HandleTaskResult(ctx context.Context, result *pb.TaskResult) error {
	task, err := s.store.GetTask(ctx, result.TaskId)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	task.Status = result.Status
	task.Output = result.Output
	task.Error = result.Error
	task.DurationMS = result.DurationMs
	task.CompletedAt = time.Now()

	if err := s.store.SaveTask(ctx, task); err != nil {
		return err
	}

	log.Printf("Task completed: %s (status: %s, duration: %dms)",
		result.TaskId, result.Status.String(), result.DurationMs)

	return nil
}

// ===== Message Management =====

// SendMessage sends a point-to-point message
func (s *CenterService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	msgID := generateID()

	msg := &models.Message{
		ID:        msgID,
		FromAgent: req.Message.FromAgent,
		ToAgent:   req.Message.ToAgent,
		Action:    req.Message.Action,
		Payload:   req.Message.Payload,
		Timestamp: time.Now(),
		TTL:       req.Message.Ttl,
		Headers:   req.Message.Headers,
	}

	// Save message
	if err := s.store.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	// If agent is online, send immediately
	if s.store.IsAgentOnline(ctx, msg.ToAgent) {
		s.messageChan <- msg
	}

	return &pb.SendMessageResponse{
		Success: true,
		MsgId:   msgID,
	}, nil
}

// Broadcast broadcasts a message to all agents
func (s *CenterService) Broadcast(ctx context.Context, req *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	count := 0

	s.connections.Range(func(key, value interface{}) bool {
		agentID := key.(string)
		if agentID == req.FromAgent {
			return true // Skip sender
		}

		msgID := generateID()
		msg := &models.Message{
			ID:        msgID,
			FromAgent: req.FromAgent,
			ToAgent:   agentID,
			Action:    req.Action,
			Payload:   req.Payload,
			Timestamp: time.Now(),
			TTL:       req.Ttl,
		}

		s.messageChan <- msg
		count++

		return true
	})

	return &pb.BroadcastResponse{
		Success:        true,
		DeliveredCount: int32(count),
	}, nil
}

// Multicast sends a message to specific agents
func (s *CenterService) Multicast(ctx context.Context, req *pb.MulticastRequest) (*pb.MulticastResponse, error) {
	count := 0

	for _, agentID := range req.ToAgents {
		if !s.store.IsAgentOnline(ctx, agentID) {
			continue
		}

		msgID := generateID()
		msg := &models.Message{
			ID:        msgID,
			FromAgent: req.FromAgent,
			ToAgent:   agentID,
			Action:    req.Action,
			Payload:   req.Payload,
			Timestamp: time.Now(),
			TTL:       req.Ttl,
		}

		s.messageChan <- msg
		count++
	}

	return &pb.MulticastResponse{
		Success:        true,
		DeliveredCount: int32(count),
	}, nil
}

// ===== Management APIs =====

// ListAgents lists all agents
func (s *CenterService) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	agents, total, err := s.store.ListAgents(ctx, req.Type, req.Status, int(page), int(pageSize))
	if err != nil {
		return nil, err
	}

	return &pb.ListAgentsResponse{
		Agents:   convertAgentsToProto(agents),
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetAgent gets a specific agent
func (s *CenterService) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.GetAgentResponse, error) {
	agent, err := s.store.GetAgent(ctx, req.AgentId)
	if err != nil {
		return &pb.GetAgentResponse{
			Success: false,
			Error:   "Agent not found",
		}, nil
	}

	return &pb.GetAgentResponse{
		Success: true,
		Agent:   convertAgentToProto(agent),
	}, nil
}

// DeleteAgent deletes an agent
func (s *CenterService) DeleteAgent(ctx context.Context, req *pb.DeleteAgentRequest) (*pb.DeleteAgentResponse, error) {
	// Check if agent is connected
	if _, ok := s.connections.Load(req.AgentId); ok {
		return &pb.DeleteAgentResponse{
			Success: false,
			Error:   "Agent is still connected",
		}, nil
	}

	if err := s.store.DeleteAgent(ctx, req.AgentId); err != nil {
		return nil, err
	}

	return &pb.DeleteAgentResponse{Success: true}, nil
}

// ListTasks lists tasks
func (s *CenterService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	page := req.Page
	if page == 0 {
		page = 1
	}

	tasks, total, err := s.store.ListTasks(ctx, req.Status, req.AgentId, int(page), int(pageSize))
	if err != nil {
		return nil, err
	}

	return &pb.ListTasksResponse{
		Tasks:    convertTasksToProto(tasks),
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetTask gets a specific task
func (s *CenterService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	task, err := s.store.GetTask(ctx, req.TaskId)
	if err != nil {
		return &pb.GetTaskResponse{
			Success: false,
			Error:   "Task not found",
		}, nil
	}

	return &pb.GetTaskResponse{
		Success: true,
		Task:    convertTaskToProto(task),
	}, nil
}

// GetSystemInfo returns system information
func (s *CenterService) GetSystemInfo(ctx context.Context, req *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	agentCount := 0
	s.connections.Range(func(_, _ interface{}) bool {
		agentCount++
		return true
	})

	return &pb.GetSystemInfoResponse{
		Version:     s.cfg.Server.Version,
		AgentCount:  int32(agentCount),
		TaskCount:   int32(len(s.taskQueue)),
		UptimeSeconds: int64(time.Since(time.Now()).Seconds()),
	}, nil
}

// GetCapabilities returns capabilities for an agent or all agents
func (s *CenterService) GetCapabilities(ctx context.Context, req *pb.GetCapabilitiesRequest) (*pb.GetCapabilitiesResponse, error) {
	if req.AgentId != "" {
		agent, err := s.store.GetAgent(ctx, req.AgentId)
		if err != nil {
			return &pb.GetCapabilitiesResponse{}, nil
		}

		caps := make([]*pb.Capability, len(agent.Capabilities))
		for i, c := range agent.Capabilities {
			caps[i] = &pb.Capability{
				Id:       c.ID,
				Name:     c.Name,
				Version:  c.Version,
				Category: c.Category,
				Metadata: c.Metadata,
			}
		}
		return &pb.GetCapabilitiesResponse{Capabilities: caps}, nil
	}

	// Return all capabilities from all agents
	var allCaps []*pb.Capability
	agents, _, _ := s.store.ListAgents(ctx, 0, 0, 1, 100)
	for _, agent := range agents {
		for _, c := range agent.Capabilities {
			allCaps = append(allCaps, &pb.Capability{
				Id:       c.ID,
				Name:     c.Name,
				Version:  c.Version,
				Category: c.Category,
				Metadata: c.Metadata,
			})
		}
	}

	return &pb.GetCapabilitiesResponse{Capabilities: allCaps}, nil
}

// GetMetrics returns system metrics
func (s *CenterService) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	metrics := map[string]float64{
		"agents_online":   0,
		"tasks_pending":   float64(len(s.taskQueue)),
		"tasks_completed": 0,
		"tasks_failed":    0,
	}

	s.connections.Range(func(_, _ interface{}) bool {
		metrics["agents_online"]++
		return true
	})

	return &pb.GetMetricsResponse{Metrics: metrics}, nil
}

// ===== Background Workers =====

// taskDispatcher dispatches tasks to agents
func (s *CenterService) taskDispatcher() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.taskQueue:
			// Schedule task
			agentID := s.scheduler.Schedule(task)

			if agentID == "" {
				// No available agent, retry later
				go func() {
					time.Sleep(5 * time.Second)
					s.taskQueue <- task
				}()
				continue
			}

			// Update task
			task.TargetAgent = agentID
			task.Status = pb.TaskStatus_TASK_STATUS_RUNNING
			task.StartedAt = time.Now()
			s.store.SaveTask(s.ctx, task)

			// Send task assignment to agent
			assignment := &pb.TaskAssignment{
				TaskId:       task.ID,
				ParentTaskId: task.ParentTaskID,
				SkillId:      task.SkillID,
				Input:        task.Input,
				Priority:     task.Priority,
				TimeoutMs:    task.TimeoutMS,
				CreatedAt:    task.CreatedAt.Unix(),
				Metadata:     task.Metadata,
			}

			s.SendTaskAssignment(agentID, assignment)
		}
	}
}

// messageDispatcher dispatches messages to agents
func (s *CenterService) messageDispatcher() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.messageChan:
			conn, ok := s.connections.Load(msg.ToAgent)
			if !ok {
				// Agent offline, save as pending
				continue
			}

			connection := conn.(*models.AgentConnection)
			centerMsg := &pb.CenterMessage{
				MsgId:     msg.ID,
				Timestamp: msg.Timestamp.Unix(),
				Message: &pb.Message{
					MsgId:     msg.ID,
					FromAgent: msg.FromAgent,
					ToAgent:   msg.ToAgent,
					Action:    msg.Action,
					Payload:   msg.Payload,
					Timestamp: msg.Timestamp.Unix(),
					Ttl:       msg.TTL,
					Headers:   msg.Headers,
				},
			}

			if err := connection.Stream.Send(centerMsg); err != nil {
				log.Printf("Failed to send message to %s: %v", msg.ToAgent, err)
			}
		}
	}
}

// agentMonitor monitors agent health
func (s *CenterService) agentMonitor() {
	ticker := time.NewTicker(s.cfg.Agent.HeartbeatTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.connections.Range(func(key, value interface{}) bool {
				agentID := key.(string)
				conn := value.(*models.AgentConnection)

				if time.Since(conn.LastSeen) > s.cfg.Agent.HeartbeatTimeout {
					// Agent timed out
					log.Printf("Agent timed out: %s", agentID)
					s.RemoveConnection(agentID)
					s.scheduler.UnregisterAgent(agentID)
				}

				return true
			})
		}
	}
}

// SendTaskAssignment sends task to agent
func (s *CenterService) SendTaskAssignment(agentID string, assignment *pb.TaskAssignment) error {
	conn, ok := s.connections.Load(agentID)
	if !ok {
		return fmt.Errorf("agent not connected: %s", agentID)
	}

	connection := conn.(*models.AgentConnection)
	centerMsg := &pb.CenterMessage{
		MsgId:     generateID(),
		Timestamp: time.Now().Unix(),
		Task:      assignment,
	}

	return connection.Stream.Send(centerMsg)
}

// SendCancelTaskRequest sends cancel request to agent
func (s *CenterService) SendCancelTaskRequest(agentID string, req *pb.CancelTaskRequest) {
	conn, ok := s.connections.Load(agentID)
	if !ok {
		return
	}

	connection := conn.(*models.AgentConnection)
	centerMsg := &pb.CenterMessage{
		MsgId:        generateID(),
		Timestamp:    time.Now().Unix(),
		CancelTask:   req,
	}

	connection.Stream.Send(centerMsg)
}

// ===== Helper Functions =====

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getDefaultConfig() map[string]string {
	return map[string]string{
		"log_level":     "info",
		"max_tasks":     "100",
		"retry_count":   "3",
	}
}

func convertDeviceInfo(d *pb.DeviceInfo) *models.DeviceInfo {
	if d == nil {
		return nil
	}
	return &models.DeviceInfo{
		OS:           d.Os,
		OSVersion:    d.OsVersion,
		Model:        d.Model,
		Manufacturer: d.Manufacturer,
		TotalMemory:  d.TotalMemory,
		CPUCores:     d.CpuCores,
		Arch:         d.Arch,
	}
}

func convertCapabilities(c []*pb.Capability) []models.Capability {
	result := make([]models.Capability, len(c))
	for i, cap := range c {
		result[i] = models.Capability{
			ID:       cap.Id,
			Name:     cap.Name,
			Version:  cap.Version,
			Category: cap.Category,
			Metadata: cap.Metadata,
		}
	}
	return result
}

func convertResourceUsage(r *pb.ResourceUsage) *models.ResourceUsage {
	if r == nil {
		return nil
	}
	return &models.ResourceUsage{
		CPUUsage:       r.CpuUsage,
		MemoryUsage:    r.MemoryUsage,
		BatteryLevel:   r.BatteryLevel,
		NetworkType:    r.NetworkType,
		NetworkLatency: r.NetworkLatencyMs,
	}
}

func convertTaskToProto(t *models.Task) *pb.Task {
	return &pb.Task{
		TaskId:       t.ID,
		ParentTaskId: t.ParentTaskID,
		SkillId:      t.SkillID,
		Input:        t.Input,
		Output:       t.Output,
		Status:       t.Status,
		Priority:     t.Priority,
		TargetAgent:  t.TargetAgent,
		SourceAgent:  t.SourceAgent,
		Error:        t.Error,
		CreatedAt:    t.CreatedAt.Unix(),
		StartedAt:    t.StartedAt.Unix(),
		CompletedAt:  t.CompletedAt.Unix(),
		DurationMs:   t.DurationMS,
		TimeoutMs:    t.TimeoutMS,
		Metadata:     t.Metadata,
	}
}

func convertAgentsToProto(agents []*models.Agent) []*pb.AgentStatusInfo {
	result := make([]*pb.AgentStatusInfo, len(agents))
	for i, a := range agents {
		result[i] = convertAgentToProto(a)
	}
	return result
}

func convertAgentToProto(a *models.Agent) *pb.AgentStatusInfo {
	caps := make([]*pb.Capability, len(a.Capabilities))
	for i, c := range a.Capabilities {
		caps[i] = &pb.Capability{
			Id:       c.ID,
			Name:     c.Name,
			Version:  c.Version,
			Category: c.Category,
			Metadata: c.Metadata,
		}
	}

	return &pb.AgentStatusInfo{
		AgentId:     a.ID,
		Status:      a.Status,
		Capabilities: caps,
		LastSeen:    a.LastSeen.Unix(),
	}
}

func convertTasksToProto(tasks []*models.Task) []*pb.Task {
	result := make([]*pb.Task, len(tasks))
	for i, t := range tasks {
		result[i] = convertTaskToProto(t)
	}
	return result
}

// ===== Identity API Methods (v1.2.0) =====

// CreateIdentity creates a new personal identity
func (s *CenterService) CreateIdentity(ctx context.Context, req *pb.CreateIdentityRequest) (*pb.CreateIdentityResponse, error) {
	createReq := &identity.CreateIdentityRequest{
		ID:         req.Id,
		Name:       req.Name,
		Nickname:   req.Nickname,
		Avatar:     req.Avatar,
		Birthday:   time.Unix(req.Birthday, 0),
		Gender:     req.Gender,
		Location:   req.Location,
		Occupation: req.Occupation,
		Languages:  req.Languages,
		Timezone:   req.Timezone,
	}

	ident, err := s.identity.CreateIdentity(ctx, createReq)
	if err != nil {
		return &pb.CreateIdentityResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.CreateIdentityResponse{
		Success:  true,
		Identity: convertIdentityToProto(ident),
	}, nil
}

// GetIdentity retrieves a personal identity
func (s *CenterService) GetIdentity(ctx context.Context, req *pb.GetIdentityRequest) (*pb.GetIdentityResponse, error) {
	ident, err := s.identity.GetIdentity(ctx, req.Id)
	if err != nil {
		return &pb.GetIdentityResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetIdentityResponse{
		Success:  true,
		Identity: convertIdentityToProto(ident),
	}, nil
}

// UpdateIdentity updates a personal identity
func (s *CenterService) UpdateIdentity(ctx context.Context, req *pb.UpdateIdentityRequest) (*pb.UpdateIdentityResponse, error) {
	ident, err := s.identity.UpdateIdentity(ctx, req.Id, req.Updates)
	if err != nil {
		return &pb.UpdateIdentityResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdateIdentityResponse{
		Success:  true,
		Identity: convertIdentityToProto(ident),
	}, nil
}

// DeleteIdentity deletes a personal identity
func (s *CenterService) DeleteIdentity(ctx context.Context, req *pb.DeleteIdentityRequest) (*pb.DeleteIdentityResponse, error) {
	err := s.identity.DeleteIdentity(ctx, req.Id)
	if err != nil {
		return &pb.DeleteIdentityResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.DeleteIdentityResponse{Success: true}, nil
}

// ListIdentities lists all identities
func (s *CenterService) ListIdentities(ctx context.Context, req *pb.ListIdentitiesRequest) (*pb.ListIdentitiesResponse, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}

	idents, total, err := s.identity.ListIdentities(ctx, int(page), int(pageSize))
	if err != nil {
		return &pb.ListIdentitiesResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.ListIdentitiesResponse{
		Success:    true,
		Identities: convertIdentitiesToProto(idents),
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// UpdatePersonality updates personality traits
func (s *CenterService) UpdatePersonality(ctx context.Context, req *pb.UpdatePersonalityRequest) (*pb.UpdatePersonalityResponse, error) {
	personality, err := s.identity.UpdatePersonality(ctx, req.Id, req.Updates)
	if err != nil {
		return &pb.UpdatePersonalityResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdatePersonalityResponse{
		Success:    true,
		Personality: convertPersonalityToProto(personality),
	}, nil
}

// SetSpeakingTone sets speaking tone preferences
func (s *CenterService) SetSpeakingTone(ctx context.Context, req *pb.SetSpeakingToneRequest) (*pb.SetSpeakingToneResponse, error) {
	err := s.identity.SetSpeakingTone(ctx, req.Id, req.Tone, req.ResponseLength, req.EmojiUsage)
	if err != nil {
		return &pb.SetSpeakingToneResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SetSpeakingToneResponse{Success: true}, nil
}

// UpdateValueSystem updates value system
func (s *CenterService) UpdateValueSystem(ctx context.Context, req *pb.UpdateValueSystemRequest) (*pb.UpdateValueSystemResponse, error) {
	valueSystem, err := s.identity.UpdateValueSystem(ctx, req.Id, req.Updates)
	if err != nil {
		return &pb.UpdateValueSystemResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdateValueSystemResponse{
		Success:     true,
		ValueSystem: convertValueSystemToProto(valueSystem),
	}, nil
}

// AddInterest adds an interest
func (s *CenterService) AddInterest(ctx context.Context, req *pb.AddInterestRequest) (*pb.AddInterestResponse, error) {
	interest := convertProtoToInterest(req.Interest)
	err := s.identity.AddInterest(ctx, req.Id, interest)
	if err != nil {
		return &pb.AddInterestResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.AddInterestResponse{Success: true}, nil
}

// RemoveInterest removes an interest
func (s *CenterService) RemoveInterest(ctx context.Context, req *pb.RemoveInterestRequest) (*pb.RemoveInterestResponse, error) {
	err := s.identity.RemoveInterest(ctx, req.Id, req.InterestId)
	if err != nil {
		return &pb.RemoveInterestResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.RemoveInterestResponse{Success: true}, nil
}

// GetInterests retrieves interests
func (s *CenterService) GetInterests(ctx context.Context, req *pb.GetInterestsRequest) (*pb.GetInterestsResponse, error) {
	var interests []models.Interest
	var err error

	if req.Category != "" {
		interests, err = s.identity.GetInterestsByCategory(ctx, req.Id, req.Category)
	} else {
		interests, err = s.identity.GetInterests(ctx, req.Id)
	}

	if err != nil {
		return &pb.GetInterestsResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetInterestsResponse{
		Success:   true,
		Interests: convertInterestsToProto(interests),
	}, nil
}

// GetVoiceProfile retrieves voice profile
func (s *CenterService) GetVoiceProfile(ctx context.Context, req *pb.GetVoiceProfileRequest) (*pb.GetVoiceProfileResponse, error) {
	profile, err := s.identity.GetVoiceProfile(ctx, req.Id)
	if err != nil {
		return &pb.GetVoiceProfileResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetVoiceProfileResponse{
		Success:     true,
		VoiceProfile: convertVoiceProfileToProto(profile),
	}, nil
}

// UpdateVoiceProfile updates voice profile
func (s *CenterService) UpdateVoiceProfile(ctx context.Context, req *pb.UpdateVoiceProfileRequest) (*pb.UpdateVoiceProfileResponse, error) {
	profile := convertProtoToVoiceProfile(req.Profile)
	err := s.identity.UpdateVoiceProfile(ctx, req.Id, profile)
	if err != nil {
		return &pb.UpdateVoiceProfileResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdateVoiceProfileResponse{Success: true}, nil
}

// GetWritingStyle retrieves writing style
func (s *CenterService) GetWritingStyle(ctx context.Context, req *pb.GetWritingStyleRequest) (*pb.GetWritingStyleResponse, error) {
	style, err := s.identity.GetWritingStyle(ctx, req.Id)
	if err != nil {
		return &pb.GetWritingStyleResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetWritingStyleResponse{
		Success:      true,
		WritingStyle: convertWritingStyleToProto(style),
	}, nil
}

// UpdateWritingStyle updates writing style
func (s *CenterService) UpdateWritingStyle(ctx context.Context, req *pb.UpdateWritingStyleRequest) (*pb.UpdateWritingStyleResponse, error) {
	style := convertProtoToWritingStyle(req.Style)
	err := s.identity.UpdateWritingStyle(ctx, req.Id, style)
	if err != nil {
		return &pb.UpdateWritingStyleResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.UpdateWritingStyleResponse{Success: true}, nil
}

// GetDecisionContext retrieves decision context for AI decision engine
func (s *CenterService) GetDecisionContext(ctx context.Context, req *pb.GetDecisionContextRequest) (*pb.GetDecisionContextResponse, error) {
	decisionCtx, err := s.identity.GetDecisionContext(ctx, req.Id)
	if err != nil {
		return &pb.GetDecisionContextResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetDecisionContextResponse{
		Success:         true,
		DecisionContext: convertDecisionContextToProto(decisionCtx),
	}, nil
}

// ===== Identity Conversion Helpers =====

func convertIdentityToProto(i *models.PersonalIdentity) *pb.PersonalIdentity {
	if i == nil {
		return nil
	}

	return &pb.PersonalIdentity{
		Id:         i.ID,
		Name:       i.Name,
		Nickname:   i.Nickname,
		Avatar:     i.Avatar,
		Birthday:   i.Birthday.Unix(),
		Gender:     i.Gender,
		Location:   i.Location,
		Occupation: i.Occupation,
		Languages:  i.Languages,
		Timezone:   i.Timezone,
		Personality:    convertPersonalityToProto(i.Personality),
		ValueSystem:    convertValueSystemToProto(i.ValueSystem),
		Interests:      convertInterestsToProto(i.Interests),
		VoiceProfile:   convertVoiceProfileToProto(i.VoiceProfile),
		WritingStyle:   convertWritingStyleToProto(i.WritingStyle),
		CreatedAt:      i.CreatedAt.Unix(),
		UpdatedAt:      i.UpdatedAt.Unix(),
	}
}

func convertIdentitiesToProto(idents []*models.PersonalIdentity) []*pb.PersonalIdentity {
	result := make([]*pb.PersonalIdentity, len(idents))
	for i, ident := range idents {
		result[i] = convertIdentityToProto(ident)
	}
	return result
}

func convertPersonalityToProto(p *models.Personality) *pb.Personality {
	if p == nil {
		return nil
	}
	return &pb.Personality{
		Openness:          p.Openness,
		Conscientiousness: p.Conscientiousness,
		Extraversion:      p.Extraversion,
		Agreeableness:     p.Agreeableness,
		Neuroticism:       p.Neuroticism,
		CustomTraits:      p.CustomTraits,
		SpeakingTone:      p.SpeakingTone,
		ResponseLength:    p.ResponseLength,
		EmojiUsage:        p.EmojiUsage,
		Summary:           p.Summary,
	}
}

func convertValueSystemToProto(v *models.ValueSystem) *pb.ValueSystem {
	if v == nil {
		return nil
	}
	return &pb.ValueSystem{
		Privacy:        v.Privacy,
		Efficiency:     v.Efficiency,
		Health:         v.Health,
		Family:         v.Family,
		Career:         v.Career,
		Entertainment:  v.Entertainment,
		Learning:       v.Learning,
		Social:         v.Social,
		Finance:        v.Finance,
		Environment:    v.Environment,
		RiskTolerance:  v.RiskTolerance,
		Impulsiveness:  v.Impulsiveness,
		Patience:       v.Patience,
		CustomValues:   v.CustomValues,
		Summary:        v.Summary,
	}
}

func convertInterestsToProto(interests []models.Interest) []*pb.Interest {
	result := make([]*pb.Interest, len(interests))
	for i, interest := range interests {
		result[i] = &pb.Interest{
			Id:          interest.ID,
			Category:    interest.Category,
			Name:        interest.Name,
			Level:       interest.Level,
			Keywords:    interest.Keywords,
			Description: interest.Description,
			Since:       interest.Since.Unix(),
			LastActive:  interest.LastActive.Unix(),
		}
	}
	return result
}

func convertVoiceProfileToProto(v *models.VoiceProfile) *pb.VoiceProfile {
	if v == nil {
		return nil
	}
	return &pb.VoiceProfile{
		Id:               v.ID,
		VoiceType:        v.VoiceType,
		PresetVoiceId:    v.PresetVoiceID,
		CloneReferenceId: v.CloneReferenceID,
		Pitch:            v.Pitch,
		Speed:            v.Speed,
		Volume:           v.Volume,
		Tone:             v.Tone,
		Accent:           v.Accent,
		EmotionLevel:     v.EmotionLevel,
		PausePattern:     v.PausePattern,
		EmphasisStyle:    v.EmphasisStyle,
		CreatedAt:        v.CreatedAt.Unix(),
		UpdatedAt:        v.UpdatedAt.Unix(),
	}
}

func convertWritingStyleToProto(w *models.WritingStyle) *pb.WritingStyle {
	if w == nil {
		return nil
	}
	return &pb.WritingStyle{
		Formality:         w.Formality,
		Verbosity:         w.Verbosity,
		Humor:             w.Humor,
		Technicality:      w.Technicality,
		UseEmoji:          w.UseEmoji,
		UseGifs:           w.UseGIFs,
		UseMarkdown:       w.UseMarkdown,
		SignaturePhrase:   w.SignaturePhrase,
		FrequentWords:     w.FrequentWords,
		AvoidWords:        w.AvoidWords,
		PreferredGreeting: w.PreferredGreeting,
		PreferredClosing:  w.PreferredClosing,
	}
}

func convertDecisionContextToProto(d *identity.DecisionContext) *pb.DecisionContext {
	if d == nil {
		return nil
	}

	interests := make([]*pb.Interest, len(d.Interests))
	for i, interest := range d.Interests {
		interests[i] = &pb.Interest{
			Id:          interest.ID,
			Category:    interest.Category,
			Name:        interest.Name,
			Level:       interest.Level,
			Keywords:    interest.Keywords,
			Description: interest.Description,
			Since:       interest.Since.Unix(),
			LastActive:  interest.LastActive.Unix(),
		}
	}

	return &pb.DecisionContext{
		UserId:         d.UserID,
		Personality:    convertPersonalityToProto(d.Personality),
		ValueSystem:    convertValueSystemToProto(d.ValueSystem),
		Interests:      interests,
		SpeakingTone:   d.SpeakingTone,
		ResponseLength: d.ResponseLength,
		ValuePriority:  d.ValuePriority,
	}
}

// Proto to Model conversions

func convertProtoToInterest(i *pb.Interest) models.Interest {
	if i == nil {
		return models.Interest{}
	}
	return models.Interest{
		ID:          i.Id,
		Category:    i.Category,
		Name:        i.Name,
		Level:       i.Level,
		Keywords:    i.Keywords,
		Description: i.Description,
		Since:       time.Unix(i.Since, 0),
		LastActive:  time.Unix(i.LastActive, 0),
	}
}

func convertProtoToVoiceProfile(v *pb.VoiceProfile) *models.VoiceProfile {
	if v == nil {
		return nil
	}
	return &models.VoiceProfile{
		ID:               v.Id,
		VoiceType:        v.VoiceType,
		PresetVoiceID:    v.PresetVoiceId,
		CloneReferenceID: v.CloneReferenceId,
		Pitch:            v.Pitch,
		Speed:            v.Speed,
		Volume:           v.Volume,
		Tone:             v.Tone,
		Accent:           v.Accent,
		EmotionLevel:     v.EmotionLevel,
		PausePattern:     v.PausePattern,
		EmphasisStyle:    v.EmphasisStyle,
		CreatedAt:        time.Unix(v.CreatedAt, 0),
		UpdatedAt:        time.Unix(v.UpdatedAt, 0),
	}
}

func convertProtoToWritingStyle(w *pb.WritingStyle) *models.WritingStyle {
	if w == nil {
		return nil
	}
	return &models.WritingStyle{
		Formality:         w.Formality,
		Verbosity:         w.Verbosity,
		Humor:             w.Humor,
		Technicality:      w.Technicality,
		UseEmoji:          w.UseEmoji,
		UseGIFs:           w.UseGifs,
		UseMarkdown:       w.UseMarkdown,
		SignaturePhrase:   w.SignaturePhrase,
		FrequentWords:     w.FrequentWords,
		AvoidWords:        w.AvoidWords,
		PreferredGreeting: w.PreferredGreeting,
		PreferredClosing:  w.PreferredClosing,
	}
}