package store

import (
	"context"
	"time"

	"github.com/ofa/center/internal/models"

	pb "github.com/ofa/center/proto"
)

// StoreInterface defines the storage interface
type StoreInterface interface {
	// Lifecycle
	Close() error

	// Agent operations
	SaveAgent(ctx context.Context, agent *models.Agent) error
	GetAgent(ctx context.Context, id string) (*models.Agent, error)
	ListAgents(ctx context.Context, agentType pb.AgentType, status pb.AgentStatus, page, pageSize int) ([]*models.Agent, int, error)
	DeleteAgent(ctx context.Context, id string) error

	// Task operations
	SaveTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	ListTasks(ctx context.Context, status pb.TaskStatus, agentID string, page, pageSize int) ([]*models.Task, int, error)

	// Message operations
	SaveMessage(ctx context.Context, msg *models.Message) error
	GetPendingMessages(ctx context.Context, agentID string) ([]*models.Message, error)
	MarkMessageDelivered(ctx context.Context, msgID string) error

	// Cache operations (Redis-like)
	SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error
	IsAgentOnline(ctx context.Context, agentID string) bool
	SetAgentResources(ctx context.Context, agentID string, resources *models.ResourceUsage) error
	GetAgentResources(ctx context.Context, agentID string) (*models.ResourceUsage, error)
}

// CacheInterface defines the cache interface for Redis-like operations
type CacheInterface interface {
	// Key-Value operations
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Set operations (for agent groups, etc.)
	SAdd(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...string) error

	// Pub/Sub
	Publish(ctx context.Context, channel string, message interface{}) error
	Subscribe(ctx context.Context, channel string) (<-chan interface{}, error)

	// Lifecycle
	Close() error
}