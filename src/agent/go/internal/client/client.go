package client

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

	pb "github.com/ofa/agent/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient handles communication with the Center
type AgentClient struct {
	centerAddr string
	conn       *grpc.ClientConn
	stream     pb.AgentService_ConnectClient
	agentID    string
	token      string

	// Handlers
	taskHandler  TaskHandler
	messageHandler MessageHandler

	// State
	connected bool
	mu        sync.Mutex

	// Channels
	taskChan    chan *pb.TaskAssignment
	messageChan chan *pb.Message
	doneChan    chan struct{}
}

// TaskHandler handles task execution
type TaskHandler func(ctx context.Context, task *pb.TaskAssignment) (*pb.TaskResult, error)

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, msg *pb.Message) error

// Config holds agent configuration
type Config struct {
	CenterAddr      string
	Name            string
	Type            pb.AgentType
	DeviceInfo      *pb.DeviceInfo
	Capabilities    []*pb.Capability
	Metadata        map[string]string
	TaskHandler     TaskHandler
	MessageHandler  MessageHandler
}

// NewAgentClient creates a new agent client
func NewAgentClient(cfg *Config) (*AgentClient, error) {
	conn, err := grpc.Dial(cfg.CenterAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to center: %w", err)
	}

	client := &AgentClient{
		centerAddr:    cfg.CenterAddr,
		conn:          conn,
		taskHandler:   cfg.TaskHandler,
		messageHandler: cfg.MessageHandler,
		taskChan:      make(chan *pb.TaskAssignment, 10),
		messageChan:   make(chan *pb.Message, 100),
		doneChan:      make(chan struct{}),
	}

	return client, nil
}

// Connect connects to the Center and starts the main loop
func (c *AgentClient) Connect(ctx context.Context, cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	stream, err := pb.NewAgentServiceClient(c.conn).Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	c.stream = stream

	// Send registration
	msgID := generateID()
	registerReq := &pb.RegisterRequest{
		Name:         cfg.Name,
		Type:         cfg.Type,
		DeviceInfo:   cfg.DeviceInfo,
		Capabilities: cfg.Capabilities,
		Metadata:     cfg.Metadata,
	}

	if err := stream.Send(&pb.AgentMessage{
		MsgId:     msgID,
		Timestamp: time.Now().Unix(),
		Register:  registerReq,
	}); err != nil {
		return fmt.Errorf("failed to send register: %w", err)
	}

	// Wait for register response
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive register response: %w", err)
	}

	if resp.Register == nil || !resp.Register.Success {
		return fmt.Errorf("registration failed: %s", resp.Register.Error)
	}

	c.agentID = resp.Register.AgentId
	c.token = resp.Register.Token
	c.connected = true

	log.Printf("Agent registered: %s", c.agentID)

	// Start message receiver
	go c.receiveLoop(ctx)

	// Start heartbeat sender
	heartbeatInterval := time.Duration(resp.Register.HeartbeatIntervalMs) * time.Millisecond
	go c.heartbeatLoop(ctx, heartbeatInterval)

	// Start task processor
	go c.taskProcessor(ctx)

	// Start message processor
	go c.messageProcessor(ctx)

	return nil
}

// Disconnect disconnects from the Center
func (c *AgentClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		close(c.doneChan)
		c.stream.CloseSend()
		c.conn.Close()
		c.connected = false
	}
}

// GetAgentID returns the agent ID
func (c *AgentClient) GetAgentID() string {
	return c.agentID
}

// receiveLoop receives messages from the Center
func (c *AgentClient) receiveLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.doneChan:
			return
		default:
			msg, err := c.stream.Recv()
			if err == io.EOF {
				log.Println("Stream closed by server")
				c.connected = false
				return
			}
			if err != nil {
				log.Printf("Receive error: %v", err)
				continue
			}

			// Handle message
			if msg.Task != nil {
				c.taskChan <- msg.Task
			}
			if msg.Message != nil {
				c.messageChan <- msg.Message
			}
			if msg.CancelTask != nil {
				// Handle task cancellation
				log.Printf("Task cancelled: %s", msg.CancelTask.TaskId)
			}
		}
	}
}

// heartbeatLoop sends heartbeats to the Center
func (c *AgentClient) heartbeatLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.doneChan:
			return
		case <-ticker.C:
			if !c.connected {
				return
			}

			msgID := generateID()
			heartbeat := &pb.HeartbeatRequest{
				AgentId:      c.agentID,
				Status:       pb.AgentStatus_AGENT_STATUS_ONLINE,
				PendingTasks: int32(len(c.taskChan)),
			}

			c.stream.Send(&pb.AgentMessage{
				MsgId:     msgID,
				Timestamp: time.Now().Unix(),
				Heartbeat: heartbeat,
			})
		}
	}
}

// taskProcessor processes incoming tasks
func (c *AgentClient) taskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.doneChan:
			return
		case task := <-c.taskChan:
			if c.taskHandler == nil {
				log.Printf("No task handler, skipping task: %s", task.TaskId)
				continue
			}

			// Execute task
			result, err := c.taskHandler(ctx, task)
			if err != nil {
				result = &pb.TaskResult{
					TaskId: task.TaskId,
					Status: pb.TaskStatus_TASK_STATUS_FAILED,
					Error:  err.Error(),
				}
			}

			// Send result
			msgID := generateID()
			c.stream.Send(&pb.AgentMessage{
				MsgId:     msgID,
				Timestamp: time.Now().Unix(),
				TaskResult: result,
			})
		}
	}
}

// messageProcessor processes incoming messages
func (c *AgentClient) messageProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.doneChan:
			return
		case msg := <-c.messageChan:
			if c.messageHandler != nil {
				err := c.messageHandler(ctx, msg)
				// Send response
				msgID := generateID()
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				c.stream.Send(&pb.AgentMessage{
					MsgId: msgID,
					Timestamp: time.Now().Unix(),
					MessageResponse: &pb.MessageResponse{
						MsgId:   msg.MsgId,
						Success: err == nil,
						Error:   errStr,
					},
				})
			}
		}
	}
}

// SendMessage sends a message to another agent
func (c *AgentClient) SendMessage(ctx context.Context, toAgent, action string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := pb.NewMessageServiceClient(c.conn)
	_, err = client.SendMessage(ctx, &pb.SendMessageRequest{
		Message: &pb.Message{
			FromAgent: c.agentID,
			ToAgent:   toAgent,
			Action:    action,
			Payload:   data,
			Timestamp: time.Now().Unix(),
		},
	})
	return err
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}