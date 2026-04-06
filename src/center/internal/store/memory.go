package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/models"

	pb "github.com/ofa/center/proto"
)

// MemoryStore implements StoreInterface using in-memory storage
type MemoryStore struct {
	agents    sync.Map // map[string]*models.Agent
	tasks     sync.Map // map[string]*models.Task
	messages  sync.Map // map[string]*models.Message
	online    sync.Map // map[string]bool - online status
	resources sync.Map // map[string]*models.ResourceUsage
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() (*MemoryStore, error) {
	return &MemoryStore{}, nil
}

// Close closes all connections
func (s *MemoryStore) Close() error {
	return nil
}

// ===== Agent Store =====

// SaveAgent saves or updates an agent
func (s *MemoryStore) SaveAgent(ctx context.Context, agent *models.Agent) error {
	s.agents.Store(agent.ID, agent)
	return nil
}

// GetAgent retrieves an agent by ID
func (s *MemoryStore) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	v, ok := s.agents.Load(id)
	if !ok {
		return nil, ErrNotFound
	}
	return v.(*models.Agent), nil
}

// ListAgents lists agents with filters
func (s *MemoryStore) ListAgents(ctx context.Context, agentType pb.AgentType, status pb.AgentStatus, page, pageSize int) ([]*models.Agent, int, error) {
	var agents []*models.Agent
	s.agents.Range(func(key, value interface{}) bool {
		agent := value.(*models.Agent)
		if (agentType == 0 || agent.Type == agentType) && (status == 0 || agent.Status == status) {
			agents = append(agents, agent)
		}
		return true
	})
	return agents, len(agents), nil
}

// DeleteAgent deletes an agent
func (s *MemoryStore) DeleteAgent(ctx context.Context, id string) error {
	s.agents.Delete(id)
	return nil
}

// ===== Task Store =====

// SaveTask saves or updates a task
func (s *MemoryStore) SaveTask(ctx context.Context, task *models.Task) error {
	s.tasks.Store(task.ID, task)
	return nil
}

// GetTask retrieves a task by ID
func (s *MemoryStore) GetTask(ctx context.Context, id string) (*models.Task, error) {
	v, ok := s.tasks.Load(id)
	if !ok {
		return nil, ErrNotFound
	}
	return v.(*models.Task), nil
}

// ListTasks lists tasks with filters
func (s *MemoryStore) ListTasks(ctx context.Context, status pb.TaskStatus, agentID string, page, pageSize int) ([]*models.Task, int, error) {
	var tasks []*models.Task
	s.tasks.Range(func(key, value interface{}) bool {
		task := value.(*models.Task)
		if (status == 0 || task.Status == status) && (agentID == "" || task.TargetAgent == agentID || task.SourceAgent == agentID) {
			tasks = append(tasks, task)
		}
		return true
	})
	return tasks, len(tasks), nil
}

// ===== Message Store =====

// SaveMessage saves a message
func (s *MemoryStore) SaveMessage(ctx context.Context, msg *models.Message) error {
	s.messages.Store(msg.ID, msg)
	return nil
}

// GetPendingMessages retrieves undelivered messages for an agent
func (s *MemoryStore) GetPendingMessages(ctx context.Context, agentID string) ([]*models.Message, error) {
	var messages []*models.Message
	s.messages.Range(func(key, value interface{}) bool {
		msg := value.(*models.Message)
		if msg.ToAgent == agentID {
			messages = append(messages, msg)
		}
		return true
	})
	return messages, nil
}

// MarkMessageDelivered marks a message as delivered
func (s *MemoryStore) MarkMessageDelivered(ctx context.Context, msgID string) error {
	s.messages.Delete(msgID)
	return nil
}

// ===== Cache Operations =====

// SetAgentOnline sets agent online status
func (s *MemoryStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	s.online.Store(agentID, true)
	return nil
}

// IsAgentOnline checks if agent is online
func (s *MemoryStore) IsAgentOnline(ctx context.Context, agentID string) bool {
	v, ok := s.online.Load(agentID)
	return ok && v.(bool)
}

// SetAgentResources stores agent resources
func (s *MemoryStore) SetAgentResources(ctx context.Context, agentID string, resources *models.ResourceUsage) error {
	s.resources.Store(agentID, resources)
	return nil
}

// GetAgentResources retrieves agent resources
func (s *MemoryStore) GetAgentResources(ctx context.Context, agentID string) (*models.ResourceUsage, error) {
	v, ok := s.resources.Load(agentID)
	if !ok {
		return nil, ErrNotFound
	}
	return v.(*models.ResourceUsage), nil
}

// ===== Factory Function =====

// StoreType defines the storage backend type
type StoreType string

const (
	StoreMemory     StoreType = "memory"
	StoreSQLite     StoreType = "sqlite"
	StorePostgreSQL StoreType = "postgres"
	StoreHybrid     StoreType = "hybrid"
)

// NewStore creates a store based on configuration
func NewStore(cfg *config.Config) (StoreInterface, error) {
	storeType := StoreType(cfg.Database.Type)
	if storeType == "" {
		// Check legacy Database field
		if cfg.Database.Database == "" || cfg.Database.Database == "memory" {
			storeType = StoreMemory
		} else if cfg.Database.Host != "" {
			storeType = StorePostgreSQL
		} else {
			storeType = StoreSQLite
		}
	}

	switch storeType {
	case StoreMemory, "":
		return NewMemoryStore()
	case StoreSQLite:
		return NewSQLiteStore(cfg)
	case StorePostgreSQL:
		return NewPostgreSQLStore(PostgreSQLConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Database,
			SSLMode:  "disable",
		})
	case StoreHybrid:
		return NewHybridStore(cfg)
	default:
		return nil, fmt.Errorf("unsupported store type: %s", storeType)
	}
}

// ===== Legacy Compatibility =====

// The following types and functions are kept for backward compatibility

// Store is an alias for MemoryStore (backward compatibility)
type Store = MemoryStore

// Error definitions
var ErrNotFound = &StoreError{Message: "not found"}

type StoreError struct {
	Message string
}

func (e *StoreError) Error() string {
	return e.Message
}