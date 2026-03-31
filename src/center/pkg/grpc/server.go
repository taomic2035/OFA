package grpc

import (
	"context"
	"log"
	"net"

	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/internal/service"
	pb "github.com/ofa/center/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server implements the gRPC services
type Server struct {
	service *service.CenterService
	server  *grpc.Server
}

// NewServer creates a new gRPC server
func NewServer(service *service.CenterService) *Server {
	return &Server{
		service: service,
	}
}

// Start starts the gRPC server
func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	// Register services
	pb.RegisterAgentServiceServer(s.server, s)
	pb.RegisterMessageServiceServer(s.server, s)
	pb.RegisterManagementServiceServer(s.server, s)

	log.Printf("gRPC server listening on %s", address)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// ===== AgentService =====

// Connect handles bidirectional streaming connection
func (s *Server) Connect(stream pb.AgentService_ConnectServer) error {
	var agentID string

	for {
		msg, err := stream.Recv()
		if err != nil {
			if agentID != "" {
				s.service.RemoveConnection(agentID)
				log.Printf("Agent disconnected: %s", agentID)
			}
			return err
		}

		// Handle different message types
		if msg.Register != nil {
			// Register agent
			resp, err := s.service.RegisterAgent(stream.Context(), msg.Register)
			if err != nil {
				stream.Send(&pb.CenterMessage{
					MsgId:     msg.MsgId,
					Timestamp: msg.Timestamp,
					Register:  &pb.RegisterResponse{Success: false, Error: err.Error()},
				})
				continue
			}

			agentID = resp.AgentId
			s.service.AddConnection(agentID, stream)

			stream.Send(&pb.CenterMessage{
				MsgId:     msg.MsgId,
				Timestamp: msg.Timestamp,
				Register:  resp,
			})

		} else if msg.Heartbeat != nil {
			// Handle heartbeat
			err := s.service.HandleHeartbeat(stream.Context(), msg.Heartbeat)
			if err != nil {
				log.Printf("Heartbeat error for %s: %v", msg.Heartbeat.AgentId, err)
			}

		} else if msg.TaskResult != nil {
			// Handle task result
			err := s.service.HandleTaskResult(stream.Context(), msg.TaskResult)
			if err != nil {
				log.Printf("Task result error for %s: %v", msg.TaskResult.TaskId, err)
			}

		} else if msg.Event != nil {
			// Handle agent event
			log.Printf("Agent event from %s: %s", agentID, msg.Event.EventType)

		} else if msg.MessageResponse != nil {
			// Handle message response
			log.Printf("Message response: %s (success: %v)", msg.MessageResponse.MsgId, msg.MessageResponse.Success)
		}
	}
}

// SubmitTask handles task submission
func (s *Server) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error) {
	return s.service.SubmitTask(ctx, req)
}

// GetTaskStatus handles task status query
func (s *Server) GetTaskStatus(ctx context.Context, req *pb.GetTaskStatusRequest) (*pb.GetTaskStatusResponse, error) {
	return s.service.GetTaskStatus(ctx, req)
}

// CancelTask handles task cancellation
func (s *Server) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	return s.service.CancelTask(ctx, req)
}

// SubscribeTask handles task subscription
func (s *Server) SubscribeTask(req *pb.SubscribeTaskRequest, stream pb.AgentService_SubscribeTaskServer) error {
	// Stream task events to client
	for {
		task, err := s.service.GetStore().GetTask(stream.Context(), req.TaskId)
		if err != nil {
			break
		}

		event := &pb.TaskEvent{
			TaskId:    task.ID,
			Task:      convertTaskToProto(task),
			Timestamp: task.CreatedAt.Unix(),
		}

		if err := stream.Send(event); err != nil {
			return err
		}

		// Stop streaming if task is finished
		if task.Status == pb.TaskStatus_TASK_STATUS_COMPLETED ||
			task.Status == pb.TaskStatus_TASK_STATUS_FAILED ||
			task.Status == pb.TaskStatus_TASK_STATUS_CANCELLED {
			break
		}
	}

	return nil
}

// RegisterCapabilities handles capability registration
func (s *Server) RegisterCapabilities(ctx context.Context, req *pb.RegisterCapabilitiesRequest) (*pb.RegisterCapabilitiesResponse, error) {
	agent, err := s.service.GetStore().GetAgent(ctx, req.AgentId)
	if err != nil {
		return &pb.RegisterCapabilitiesResponse{
			Success: false,
			Error:   "Agent not found",
		}, nil
	}

	// Update capabilities
	caps := make([]models.Capability, len(req.Capabilities))
	for i, c := range req.Capabilities {
		caps[i] = models.Capability{
			ID:       c.Id,
			Name:     c.Name,
			Version:  c.Version,
			Category: c.Category,
			Metadata: c.Metadata,
		}
	}
	agent.Capabilities = caps

	if err := s.service.GetStore().SaveAgent(ctx, agent); err != nil {
		return &pb.RegisterCapabilitiesResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Update scheduler
	s.service.GetScheduler().RegisterAgentCapabilities(req.AgentId, caps)

	return &pb.RegisterCapabilitiesResponse{Success: true}, nil
}

// GetCapabilities handles capability query
func (s *Server) GetCapabilities(ctx context.Context, req *pb.GetCapabilitiesRequest) (*pb.GetCapabilitiesResponse, error) {
	if req.AgentId != "" {
		agent, err := s.service.GetStore().GetAgent(ctx, req.AgentId)
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
	agents, _, _ := s.service.GetStore().ListAgents(ctx, 0, 0, 1, 100)
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

// ===== MessageService =====

// SendMessage handles point-to-point message
func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	return s.service.SendMessage(ctx, req)
}

// Broadcast handles broadcast message
func (s *Server) Broadcast(ctx context.Context, req *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	return s.service.Broadcast(ctx, req)
}

// Multicast handles multicast message
func (s *Server) Multicast(ctx context.Context, req *pb.MulticastRequest) (*pb.MulticastResponse, error) {
	return s.service.Multicast(ctx, req)
}

// Subscribe handles message subscription
func (s *Server) Subscribe(req *pb.SubscribeMessageRequest, stream pb.MessageService_SubscribeServer) error {
	// Stream messages to subscriber
	msgs, err := s.service.GetStore().GetPendingMessages(stream.Context(), req.AgentId)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if len(req.Actions) > 0 {
			// Filter by actions
			found := false
			for _, action := range req.Actions {
				if msg.Action == action {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		protoMsg := &pb.Message{
			MsgId:     msg.ID,
			FromAgent: msg.FromAgent,
			ToAgent:   msg.ToAgent,
			Action:    msg.Action,
			Payload:   msg.Payload,
			Timestamp: msg.Timestamp.Unix(),
			Ttl:       msg.TTL,
			Headers:   msg.Headers,
		}

		if err := stream.Send(protoMsg); err != nil {
			return err
		}

		s.service.GetStore().MarkMessageDelivered(stream.Context(), msg.ID)
	}

	return nil
}

// ===== ManagementService =====

// ListAgents handles agent listing
func (s *Server) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	return s.service.ListAgents(ctx, req)
}

// GetAgent handles agent query
func (s *Server) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.GetAgentResponse, error) {
	return s.service.GetAgent(ctx, req)
}

// DeleteAgent handles agent deletion
func (s *Server) DeleteAgent(ctx context.Context, req *pb.DeleteAgentRequest) (*pb.DeleteAgentResponse, error) {
	return s.service.DeleteAgent(ctx, req)
}

// ListTasks handles task listing
func (s *Server) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	return s.service.ListTasks(ctx, req)
}

// GetTask handles task query
func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	return s.service.GetTask(ctx, req)
}

// ListSkills handles skill listing
func (s *Server) ListSkills(ctx context.Context, req *pb.ListSkillsRequest) (*pb.ListSkillsResponse, error) {
	// Get all capabilities
	resp, err := s.GetCapabilities(ctx, &pb.GetCapabilitiesRequest{})
	if err != nil {
		return nil, err
	}

	// Filter by category if specified
	var skills []*pb.Capability
	for _, cap := range resp.Capabilities {
		if req.Category == "" || cap.Category == req.Category {
			skills = append(skills, cap)
		}
	}

	return &pb.ListSkillsResponse{Skills: skills}, nil
}

// InstallSkill handles skill installation
func (s *Server) InstallSkill(ctx context.Context, req *pb.InstallSkillRequest) (*pb.InstallSkillResponse, error) {
	// This would typically involve sending a skill install command to the agent
	// For now, return success
	return &pb.InstallSkillResponse{Success: true}, nil
}

// GetSystemInfo handles system info query
func (s *Server) GetSystemInfo(ctx context.Context, req *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	return s.service.GetSystemInfo(ctx, req)
}

// GetMetrics handles metrics query
func (s *Server) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	return s.service.GetMetrics(ctx, req)
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