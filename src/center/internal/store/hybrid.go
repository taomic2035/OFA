package store

import (
	"context"
	"fmt"
	"time"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/models"
	"github.com/ofa/center/pkg/cache"

	pb "github.com/ofa/center/proto"
)

// HybridStore combines persistent storage (PostgreSQL/SQLite) with Redis cache
type HybridStore struct {
	persistent StoreInterface // PostgreSQL or SQLite for persistence
	cache      *RedisCacheStore // Redis for online status, resources, pub/sub
}

// NewHybridStore creates a hybrid store with PostgreSQL + Redis
func NewHybridStore(cfg *config.Config) (*HybridStore, error) {
	var persistent StoreInterface
	var err error

	// Initialize persistent storage
	switch cfg.Database.Type {
	case "postgres", "postgresql":
		persistent, err = NewPostgreSQLStore(PostgreSQLConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Database: cfg.Database.Database,
			SSLMode:  "disable", // Can be configured
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL store: %w", err)
		}
	case "sqlite":
		persistent, err = NewSQLiteStore(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQLite store: %w", err)
		}
	default:
		persistent, err = NewMemoryStore()
		if err != nil {
			return nil, fmt.Errorf("failed to create Memory store: %w", err)
		}
	}

	// Initialize Redis cache
	redisConfig := cache.RedisConfig{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 10,
		KeyPrefix: "ofa:center",
	}

	cacheStore, err := NewRedisCacheStore(redisConfig)
	if err != nil {
		// Non-fatal - continue without Redis
		fmt.Printf("Warning: Redis unavailable, using local cache: %v\n", err)
	}

	return &HybridStore{
		persistent: persistent,
		cache:      cacheStore,
	}, nil
}

// Close closes both stores
func (s *HybridStore) Close() error {
	if s.cache != nil {
		s.cache.Close()
	}
	return s.persistent.Close()
}

// === Agent Operations (Persistent) ===

// SaveAgent saves agent to persistent storage
func (s *HybridStore) SaveAgent(ctx context.Context, agent *models.Agent) error {
	return s.persistent.SaveAgent(ctx, agent)
}

// GetAgent retrieves agent from persistent storage
func (s *HybridStore) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	return s.persistent.GetAgent(ctx, id)
}

// ListAgents lists agents from persistent storage
func (s *HybridStore) ListAgents(ctx context.Context, agentType pb.AgentType, status pb.AgentStatus, page, pageSize int) ([]*models.Agent, int, error) {
	return s.persistent.ListAgents(ctx, agentType, status, page, pageSize)
}

// DeleteAgent deletes agent from persistent storage
func (s *HybridStore) DeleteAgent(ctx context.Context, id string) error {
	// Also clear cache
	if s.cache != nil {
		s.cache.Delete(ctx, "agent:online:"+id)
		s.cache.Delete(ctx, "agent:resources:"+id)
	}
	return s.persistent.DeleteAgent(ctx, id)
}

// === Task Operations (Persistent) ===

// SaveTask saves task to persistent storage
func (s *HybridStore) SaveTask(ctx context.Context, task *models.Task) error {
	return s.persistent.SaveTask(ctx, task)
}

// GetTask retrieves task from persistent storage
func (s *HybridStore) GetTask(ctx context.Context, id string) (*models.Task, error) {
	return s.persistent.GetTask(ctx, id)
}

// ListTasks lists tasks from persistent storage
func (s *HybridStore) ListTasks(ctx context.Context, status pb.TaskStatus, agentID string, page, pageSize int) ([]*models.Task, int, error) {
	return s.persistent.ListTasks(ctx, status, agentID, page, pageSize)
}

// === Message Operations (Persistent) ===

// SaveMessage saves message to persistent storage
func (s *HybridStore) SaveMessage(ctx context.Context, msg *models.Message) error {
	return s.persistent.SaveMessage(ctx, msg)
}

// GetPendingMessages retrieves pending messages from persistent storage
func (s *HybridStore) GetPendingMessages(ctx context.Context, agentID string) ([]*models.Message, error) {
	return s.persistent.GetPendingMessages(ctx, agentID)
}

// MarkMessageDelivered marks message as delivered
func (s *HybridStore) MarkMessageDelivered(ctx context.Context, msgID string) error {
	return s.persistent.MarkMessageDelivered(ctx, msgID)
}

// === Cache Operations (Redis) ===

// SetAgentOnline sets agent online status in Redis
func (s *HybridStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	if s.cache != nil {
		return s.cache.Set(ctx, "agent:online:"+agentID, true, ttl)
	}
	return s.persistent.SetAgentOnline(ctx, agentID, ttl)
}

// IsAgentOnline checks agent online status from Redis
func (s *HybridStore) IsAgentOnline(ctx context.Context, agentID string) bool {
	if s.cache != nil {
		exists, err := s.cache.Exists(ctx, "agent:online:"+agentID)
		if err == nil && exists {
			return true
		}
	}
	return s.persistent.IsAgentOnline(ctx, agentID)
}

// SetAgentResources stores agent resources in Redis
func (s *HybridStore) SetAgentResources(ctx context.Context, agentID string, resources *models.ResourceUsage) error {
	if s.cache != nil {
		return s.cache.Set(ctx, "agent:resources:"+agentID, resources, 5*time.Minute)
	}
	return s.persistent.SetAgentResources(ctx, agentID, resources)
}

// GetAgentResources retrieves agent resources from Redis
func (s *HybridStore) GetAgentResources(ctx context.Context, agentID string) (*models.ResourceUsage, error) {
	if s.cache != nil {
		v, err := s.cache.Get(ctx, "agent:resources:"+agentID)
		if err == nil {
			if res, ok := v.(*models.ResourceUsage); ok {
				return res, nil
			}
		}
	}
	return s.persistent.GetAgentResources(ctx, agentID)
}

// === Extended Cache Operations ===

// GetCache returns the Redis cache store for direct access
func (s *HybridStore) GetCache() *RedisCacheStore {
	return s.cache
}

// Publish publishes a message to Redis channel
func (s *HybridStore) Publish(ctx context.Context, channel string, message interface{}) error {
	if s.cache != nil {
		return s.cache.Publish(ctx, channel, message)
	}
	return fmt.Errorf("cache not available")
}

// Subscribe subscribes to a Redis channel
func (s *HybridStore) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	if s.cache != nil {
		return s.cache.Subscribe(ctx, channel)
	}
	return nil, fmt.Errorf("cache not available")
}

// GetCacheStatistics returns cache statistics
func (s *HybridStore) GetCacheStatistics(ctx context.Context) map[string]interface{} {
	if s.cache != nil {
		return s.cache.GetStatistics(ctx)
	}
	return map[string]interface{}{
		"cache_available": false,
	}
}