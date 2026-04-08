// Package sync - Redis 设备状态缓存 (v2.7.0)
//
// Center 使用 Redis 缓存设备在线状态、会话信息等。
// 实现高性能的设备状态查询和 Pub/Sub 通知。
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string        `json:"addr"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConn  int           `json:"min_idle_conn"`
	KeyPrefix    string        `json:"key_prefix"` // ofa:device:
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// DefaultRedisConfig 默认配置
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConn:  2,
		KeyPrefix:    "ofa:device:",
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// RedisDeviceCache Redis 设备缓存
type RedisDeviceCache struct {
	mu     sync.RWMutex
	client *redis.Client
	config RedisConfig
}

// NewRedisDeviceCache 创建 Redis 缓存
func NewRedisDeviceCache(config RedisConfig) (*RedisDeviceCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConn,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis device cache connected: %s", config.Addr)

	return &RedisDeviceCache{
		client: client,
		config: config,
	}, nil
}

// key 生成带前缀的 key
func (c *RedisDeviceCache) key(id string) string {
	return c.config.KeyPrefix + id
}

// === 设备状态缓存 ===

// SetDeviceOnline 设置设备在线状态
func (c *RedisDeviceCache) SetDeviceOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	key := c.key(agentID + ":online")
	return c.client.Set(ctx, key, time.Now().Unix(), ttl).Err()
}

// IsDeviceOnline 检查设备是否在线
func (c *RedisDeviceCache) IsDeviceOnline(ctx context.Context, agentID string) bool {
	key := c.key(agentID + ":online")
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		log.Printf("Redis error checking online status: %v", err)
		return false
	}
	return val != ""
}

// GetDeviceLastSeen 获取设备最后活跃时间
func (c *RedisDeviceCache) GetDeviceLastSeen(ctx context.Context, agentID string) (time.Time, error) {
	key := c.key(agentID + ":online")
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return time.Time{}, fmt.Errorf("device not found")
	}
	if err != nil {
		return time.Time{}, err
	}

	unix, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(unix, 0), nil
}

// === 设备信息缓存 ===

// CacheDeviceInfo 缓存设备信息
func (c *RedisDeviceCache) CacheDeviceInfo(ctx context.Context, device *DeviceInfo, ttl time.Duration) error {
	key := c.key(device.AgentID + ":info")
	data, err := json.Marshal(device)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedDevice 获取缓存的设备信息
func (c *RedisDeviceCache) GetCachedDevice(ctx context.Context, agentID string) (*DeviceInfo, error) {
	key := c.key(agentID + ":info")
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var device DeviceInfo
	if err := json.Unmarshal([]byte(data), &device); err != nil {
		return nil, err
	}
	return &device, nil
}

// === 同步状态缓存 ===

// CacheSyncVersion 缓存同步版本
func (c *RedisDeviceCache) CacheSyncVersion(ctx context.Context, agentID string, version int64) error {
	key := c.key(agentID + ":version")
	return c.client.Set(ctx, key, version, 0).Err() // 不过期
}

// GetCachedSyncVersion 获取缓存的同步版本
func (c *RedisDeviceCache) GetCachedSyncVersion(ctx context.Context, agentID string) (int64, error) {
	key := c.key(agentID + ":version")
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	version, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// === 身份关联缓存 ===

// CacheIdentityDevices 缓存身份关联的设备列表
func (c *RedisDeviceCache) CacheIdentityDevices(ctx context.Context, identityID string, agentIDs []string) error {
	key := c.key(identityID + ":devices")
	data, err := json.Marshal(agentIDs)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 5*time.Minute).Err()
}

// GetCachedIdentityDevices 获取缓存的身份设备列表
func (c *RedisDeviceCache) GetCachedIdentityDevices(ctx context.Context, identityID string) ([]string, error) {
	key := c.key(identityID + ":devices")
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var agentIDs []string
	if err := json.Unmarshal([]byte(data), &agentIDs); err != nil {
		return nil, err
	}
	return agentIDs, nil
}

// === Pub/Sub 通知 ===

// PublishSyncNotification 发布同步通知
func (c *RedisDeviceCache) PublishSyncNotification(ctx context.Context, identityID string, event SyncEvent) error {
	channel := c.key(identityID + ":sync")
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return c.client.Publish(ctx, channel, data).Err()
}

// SyncEvent 同步事件
type SyncEvent struct {
	Type      string    `json:"type"` // identity_update, memory_update, preference_update
	IdentityID string   `json:"identity_id"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // 来源设备
}

// SubscribeSyncEvents 订阅同步事件
func (c *RedisDeviceCache) SubscribeSyncEvents(ctx context.Context, identityID string) <-chan *SyncEvent {
	channel := c.key(identityID + ":sync")
	sub := c.client.Subscribe(ctx, channel)

	events := make(chan *SyncEvent, 100)
	go func() {
		defer close(events)
		ch := sub.Channel()
		for msg := range ch {
			var event SyncEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}
			events <- &event
		}
	}()

	return events
}

// === 统计缓存 ===

// IncrementDeviceCount 增加设备计数
func (c *RedisDeviceCache) IncrementDeviceCount(ctx context.Context, identityID string) error {
	key := c.key(identityID + ":device_count")
	return c.client.Incr(ctx, key).Err()
}

// DecrementDeviceCount 减少设备计数
func (c *RedisDeviceCache) DecrementDeviceCount(ctx context.Context, identityID string) error {
	key := c.key(identityID + ":device_count")
	return c.client.Decr(ctx, key).Err()
}

// GetCachedDeviceCount 获取缓存的设备计数
func (c *RedisDeviceCache) GetCachedDeviceCount(ctx context.Context, identityID string) (int64, error) {
	key := c.key(identityID + ":device_count")
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// === 清理 ===

// ClearDeviceCache 清除设备缓存
func (c *RedisDeviceCache) ClearDeviceCache(ctx context.Context, agentID string) error {
	keys := []string{
		c.key(agentID + ":online"),
		c.key(agentID + ":info"),
		c.key(agentID + ":version"),
	}
	return c.client.Del(ctx, keys...).Err()
}

// ClearIdentityCache 清除身份相关缓存
func (c *RedisDeviceCache) ClearIdentityCache(ctx context.Context, identityID string) error {
	keys := []string{
		c.key(identityID + ":devices"),
		c.key(identityID + ":device_count"),
	}
	return c.client.Del(ctx, keys...).Err()
}

// Close 关闭连接
func (c *RedisDeviceCache) Close() error {
	return c.client.Close()
}