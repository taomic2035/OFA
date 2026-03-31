package ofa

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/ofa/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentType defines the type of agent
type AgentType int32

const (
	AgentTypeUnknown AgentType = iota
	AgentTypeFull
	AgentTypeMobile
	AgentTypeLite
	AgentTypeIoT
	AgentTypeEdge
)

// Config holds agent configuration
type Config struct {
	CenterAddr   string
	Name         string
	Type         AgentType
	DeviceInfo   *DeviceInfo
	Metadata     map[string]string
}

// DeviceInfo holds device information
type DeviceInfo struct {
	OS           string
	OSVersion    string
	Model        string
	Manufacturer string
	TotalMemory  int64
	CPUCores     int32
	Arch         string
}

// ResourceUsage holds resource utilization
type ResourceUsage struct {
	CPUUsage       float64
	MemoryUsage    float64
	BatteryLevel   int32
	NetworkType    string
	NetworkLatency int32
}

// SkillHandler handles skill execution
type SkillHandler func(ctx context.Context, input []byte) ([]byte, error)

// Skill defines a skill
type Skill struct {
	ID       string
	Name     string
	Version  string
	Category string
	Handler  SkillHandler
}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *Message) error

// Message represents a message
type Message struct {
	MsgID     string
	FromAgent string
	ToAgent   string
	Action    string
	Payload   []byte
	Timestamp int64
}

// Agent represents an OFA agent
type Agent struct {
	config   *Config
	conn     *grpc.ClientConn
	stream   pb.AgentService_ConnectClient
	agentID  string
	token    string

	skills         sync.Map
	messageHandler MessageHandler

	taskChan    chan *pb.TaskAssignment
	messageChan chan *pb.Message
	doneChan    chan struct{}
	mu          sync.Mutex
	running     bool
}

// NewAgent creates a new agent
func NewAgent(cfg *Config) (*Agent, error) {
	conn, err := grpc.Dial(cfg.CenterAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Agent{
		config:      cfg,
		conn:        conn,
		taskChan:    make(chan *pb.TaskAssignment, 10),
		messageChan: make(chan *pb.Message, 100),
		doneChan:    make(chan struct{}),
	}, nil
}

// Connect connects to the Center
func (a *Agent) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	client := pb.NewAgentServiceClient(a.conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	a.stream = stream

	// Send registration
	regReq := &pb.RegisterRequest{
		Name:         a.config.Name,
		Type:         pb.AgentType(a.config.Type),
		DeviceInfo:   convertDeviceInfo(a.config.DeviceInfo),
		Capabilities: a.getCapabilities(),
		Metadata:     a.config.Metadata,
	}

	if err := stream.Send(&pb.AgentMessage{
		MsgId:     generateID(),
		Timestamp: time.Now().Unix(),
		Register:  regReq,
	}); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	// Wait for response
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	if resp.Register == nil || !resp.Register.Success {
		return fmt.Errorf("registration failed: %s", resp.Register.GetError())
	}

	a.agentID = resp.Register.AgentId
	a.token = resp.Register.Token
	a.running = true

	// Start background workers
	go a.receiveLoop(ctx)
	go a.heartbeatLoop(ctx, time.Duration(resp.Register.HeartbeatIntervalMs)*time.Millisecond)
	go a.taskProcessor(ctx)
	go a.messageProcessor(ctx)

	log.Printf("Agent connected: %s", a.agentID)
	return nil
}

// Disconnect disconnects from Center
func (a *Agent) Disconnect() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		close(a.doneChan)
		if a.stream != nil {
			a.stream.CloseSend()
		}
		if a.conn != nil {
			a.conn.Close()
		}
		a.running = false
	}
}

// ID returns the agent ID
func (a *Agent) ID() string {
	return a.agentID
}

// RegisterSkill registers a skill
func (a *Agent) RegisterSkill(skill *Skill) {
	a.skills.Store(skill.ID, skill)
}

// SetMessageHandler sets the message handler
func (a *Agent) SetMessageHandler(handler MessageHandler) {
	a.messageHandler = handler
}

// SendMessage sends a message to another agent
func (a *Agent) SendMessage(ctx context.Context, toAgent, action string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := pb.NewMessageServiceClient(a.conn)
	_, err = client.SendMessage(ctx, &pb.SendMessageRequest{
		Message: &pb.Message{
			FromAgent: a.agentID,
			ToAgent:   toAgent,
			Action:    action,
			Payload:   data,
			Timestamp: time.Now().Unix(),
		},
	})
	return err
}

// UpdateResources updates resource usage
func (a *Agent) UpdateResources(resources *ResourceUsage) {
	// Will be sent with next heartbeat
}

// Internal methods

func (a *Agent) getCapabilities() []*pb.Capability {
	var caps []*pb.Capability
	a.skills.Range(func(key, value interface{}) bool {
		skill := value.(*Skill)
		caps = append(caps, &pb.Capability{
			Id:       skill.ID,
			Name:     skill.Name,
			Version:  skill.Version,
			Category: skill.Category,
		})
		return true
	})
	return caps
}

func (a *Agent) receiveLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.doneChan:
			return
		default:
			msg, err := a.stream.Recv()
			if err == io.EOF {
				log.Println("Stream closed")
				return
			}
			if err != nil {
				log.Printf("Receive error: %v", err)
				continue
			}

			if msg.Task != nil {
				a.taskChan <- msg.Task
			}
			if msg.Message != nil {
				a.messageChan <- msg.Message
			}
		}
	}
}

func (a *Agent) heartbeatLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.doneChan:
			return
		case <-ticker.C:
			a.stream.Send(&pb.AgentMessage{
				MsgId:     generateID(),
				Timestamp: time.Now().Unix(),
				Heartbeat: &pb.HeartbeatRequest{
					AgentId:      a.agentID,
					Status:       1, // Online
					PendingTasks: int32(len(a.taskChan)),
				},
			})
		}
	}
}

func (a *Agent) taskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.doneChan:
			return
		case task := <-a.taskChan:
			skill, ok := a.skills.Load(task.SkillId)
			if !ok {
				a.sendTaskResult(task.TaskId, nil, fmt.Errorf("skill not found: %s", task.SkillId))
				continue
			}

			s := skill.(*Skill)
			output, err := s.Handler(ctx, task.Input)
			a.sendTaskResult(task.TaskId, output, err)
		}
	}
}

func (a *Agent) messageProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.doneChan:
			return
		case msg := <-a.messageChan:
			if a.messageHandler != nil {
				err := a.messageHandler(ctx, &Message{
					MsgID:     msg.MsgId,
					FromAgent: msg.FromAgent,
					ToAgent:   msg.ToAgent,
					Action:    msg.Action,
					Payload:   msg.Payload,
					Timestamp: msg.Timestamp,
				})

				a.stream.Send(&pb.AgentMessage{
					MsgId:     generateID(),
					Timestamp: time.Now().Unix(),
					MessageResponse: &pb.MessageResponse{
						MsgId:   msg.MsgId,
						Success: err == nil,
						Error:   errStr(err),
					},
				})
			}
		}
	}
}

func (a *Agent) sendTaskResult(taskID string, output []byte, err error) {
	result := &pb.TaskResult{
		TaskId: taskID,
		Output: output,
	}
	if err != nil {
		result.Status = 4 // Failed
		result.Error = err.Error()
	} else {
		result.Status = 3 // Completed
	}

	a.stream.Send(&pb.AgentMessage{
		MsgId:     generateID(),
		Timestamp: time.Now().Unix(),
		TaskResult: result,
	})
}

func convertDeviceInfo(d *DeviceInfo) *pb.DeviceInfo {
	if d == nil {
		return nil
	}
	return &pb.DeviceInfo{
		Os:           d.OS,
		OsVersion:    d.OSVersion,
		Model:        d.Model,
		Manufacturer: d.Manufacturer,
		TotalMemory:  d.TotalMemory,
		CpuCores:     d.CPUCores,
		Arch:         d.Arch,
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}