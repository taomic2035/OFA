package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/pkg/cache"
)

// RedisCacheStore implements CacheInterface using Redis
type RedisCacheStore struct {
	redis  *cache.RedisCache
	local  *LocalCacheLayer // L1 local cache fallback
	mu     sync.RWMutex
}

// LocalCacheLayer provides fallback when Redis is unavailable
type LocalCacheLayer struct {
	data     sync.Map // key -> value
	ttls     sync.Map // key -> expiryTime
	sets     sync.Map // key -> set of members
	channels map[string][]chan interface{}
	mu       sync.RWMutex
}

// NewRedisCacheStore creates a Redis-backed cache store
func NewRedisCacheStore(cfg cache.RedisConfig) (*RedisCacheStore, error) {
	redisCache := cache.NewRedisCache(cfg)

	// Try to connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisCache.Connect(ctx); err != nil {
		// Redis unavailable, use local fallback
		return &RedisCacheStore{
			redis: redisCache,
			local: &LocalCacheLayer{},
		}, nil
	}

	return &RedisCacheStore{
		redis: redisCache,
		local: &LocalCacheLayer{},
	}, nil
}

// === Key-Value Operations ===

// Set stores a value with TTL
func (s *RedisCacheStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.redis.IsConnected() {
		return s.redis.Set(ctx, key, value, ttl)
	}

	// Fallback to local cache
	s.local.data.Store(key, value)
	if ttl > 0 {
		s.local.ttls.Store(key, time.Now().Add(ttl))
	}
	return nil
}

// Get retrieves a value
func (s *RedisCacheStore) Get(ctx context.Context, key string) (interface{}, error) {
	// Check TTL expiry for local cache
	if expiry, ok := s.local.ttls.Load(key); ok {
		if time.Now().After(expiry.(time.Time)) {
			s.local.data.Delete(key)
			s.local.ttls.Delete(key)
			return nil, fmt.Errorf("key expired: %s", key)
		}
	}

	if s.redis.IsConnected() {
		return s.redis.Get(ctx, key)
	}

	// Fallback to local cache
	v, ok := s.local.data.Load(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return v, nil
}

// Delete removes a key
func (s *RedisCacheStore) Delete(ctx context.Context, key string) error {
	if s.redis.IsConnected() {
		return s.redis.Delete(ctx, key)
	}

	s.local.data.Delete(key)
	s.local.ttls.Delete(key)
	return nil
}

// Exists checks if a key exists
func (s *RedisCacheStore) Exists(ctx context.Context, key string) (bool, error) {
	// Check TTL expiry
	if expiry, ok := s.local.ttls.Load(key); ok {
		if time.Now().After(expiry.(time.Time)) {
			return false, nil
		}
	}

	if s.redis.IsConnected() {
		return s.redis.Exists(ctx, key), nil
	}

	_, ok := s.local.data.Load(key)
	return ok, nil
}

// === Set Operations ===

// SAdd adds members to a set
func (s *RedisCacheStore) SAdd(ctx context.Context, key string, members ...string) error {
	if s.redis.IsConnected() {
		client := s.redis.GetClient()
		return client.SAdd(ctx, key, members).Err()
	}

	// Fallback to local
	v, _ := s.local.sets.LoadOrStore(key, &sync.Map{})
	setMap := v.(*sync.Map)
	for _, m := range members {
		setMap.Store(m, true)
	}
	return nil
}

// SMembers returns all members of a set
func (s *RedisCacheStore) SMembers(ctx context.Context, key string) ([]string, error) {
	if s.redis.IsConnected() {
		client := s.redis.GetClient()
		return client.SMembers(ctx, key).Result()
	}

	// Fallback to local
	v, ok := s.local.sets.Load(key)
	if !ok {
		return []string{}, nil
	}

	setMap := v.(*sync.Map)
	var members []string
	setMap.Range(func(k, _ interface{}) bool {
		members = append(members, k.(string))
		return true
	})
	return members, nil
}

// SRem removes members from a set
func (s *RedisCacheStore) SRem(ctx context.Context, key string, members ...string) error {
	if s.redis.IsConnected() {
		client := s.redis.GetClient()
		return client.SRem(ctx, key, members).Err()
	}

	// Fallback to local
	v, ok := s.local.sets.Load(key)
	if !ok {
		return nil
	}

	setMap := v.(*sync.Map)
	for _, m := range members {
		setMap.Delete(m)
	}
	return nil
}

// === Pub/Sub Operations ===

// Publish publishes a message to a channel
func (s *RedisCacheStore) Publish(ctx context.Context, channel string, message interface{}) error {
	if s.redis.IsConnected() {
		client := s.redis.GetClient()
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		return client.Publish(ctx, channel, data).Err()
	}

	// Fallback to local pub/sub
	s.local.mu.Lock()
	chans, ok := s.local.channels[channel]
	if ok {
		for _, ch := range chans {
			select {
			case ch <- message:
			default:
				// Channel full, skip
			}
		}
	}
	s.local.mu.Unlock()
	return nil
}

// Subscribe subscribes to a channel
func (s *RedisCacheStore) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	if s.redis.IsConnected() {
		client := s.redis.GetClient()
		sub := client.Subscribe(ctx, channel)

		// Wait for confirmation
		_, err := sub.Receive(ctx)
		if err != nil {
			return nil, err
		}

		ch := make(chan interface{}, 100)
		go func() {
			defer close(ch)
			for {
				msg, err := sub.ReceiveMessage(ctx)
				if err != nil {
					return
				}

				var value interface{}
				json.Unmarshal([]byte(msg.Payload), &value)
				select {
				case ch <- value:
				case <-ctx.Done():
					return
				}
			}
		}()

		return ch, nil
	}

	// Fallback to local pub/sub
	s.local.mu.Lock()
	ch := make(chan interface{}, 100)
	s.local.channels[channel] = append(s.local.channels[channel], ch)
	s.local.mu.Unlock()

	return ch, nil
}

// Close closes connections
func (s *RedisCacheStore) Close() error {
	if s.redis != nil {
		s.redis.Disconnect()
	}
	return nil
}

// IsConnected checks Redis connection status
func (s *RedisCacheStore) IsConnected() bool {
	return s.redis != nil && s.redis.IsConnected()
}

// GetStatistics returns cache statistics
func (s *RedisCacheStore) GetStatistics(ctx context.Context) map[string]interface{} {
	if s.redis.IsConnected() {
		return s.redis.GetStatistics(ctx)
	}

	// Return local cache stats
	localCount := 0
	s.local.data.Range(func(_, _ interface{}) bool {
		localCount++
		return true
	})

	return map[string]interface{}{
		"redis_connected": false,
		"local_cache_size": localCount,
	}
}