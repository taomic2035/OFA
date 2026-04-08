// Package cache - Redis分布式缓存
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string        `json:"addr"`          // Redis服务器地址
	Password     string        `json:"password"`      // 密码
	DB           int           `json:"db"`            // 数据库编号
	PoolSize     int           `json:"pool_size"`     // 连接池大小
	MinIdleConn  int           `json:"min_idle_conn"` // 最小空闲连接
	MaxRetries   int           `json:"max_retries"`   // 最大重试次数
	DialTimeout  time.Duration `json:"dial_timeout"`  // 连接超时
	ReadTimeout  time.Duration `json:"read_timeout"`  // 读超时
	WriteTimeout time.Duration `json:"write_timeout"` // 写超时
	PoolTimeout  time.Duration `json:"pool_timeout"`  // 连接池超时
	IdleTimeout  time.Duration `json:"idle_timeout"`  // 空闲连接超时
	KeyPrefix    string        `json:"key_prefix"`    // 键前缀
}

// CacheItem 缓存项
type CacheItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Hits      int64       `json:"hits"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
	Tags      []string    `json:"tags,omitempty"`
}

// Session 会话数据
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys     int64         `json:"total_keys"`
	TotalHits     int64         `json:"total_hits"`
	TotalMisses   int64         `json:"total_misses"`
	HitRate       float64       `json:"hit_rate"`
	TotalBytes    int64         `json:"total_bytes"`
	ExpiredKeys   int64         `json:"expired_keys"`
	EvictedKeys   int64         `json:"evicted_keys"`
	Connections   int           `json:"connections"`
	IdleConns     int           `json:"idle_conns"`
	AvgLatency    time.Duration `json:"avg_latency"`
	LastStatsTime time.Time     `json:"last_stats_time"`
}

// RedisCache Redis缓存管理器
type RedisCache struct {
	config    RedisConfig
	client    *redis.Client
	local     *LocalCache // 本地缓存层(L1)
	stats     *CacheStats
	mu        sync.RWMutex
	running   bool
}

// LocalCache 本地缓存层(L1)
type LocalCache struct {
	items    map[string]*CacheItem
	maxSize  int
	ttl      time.Duration
	mu       sync.RWMutex
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(config RedisConfig) *RedisCache {
	// 默认配置
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConn == 0 {
		config.MinIdleConn = 5
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	return &RedisCache{
		config: config,
		local:  NewLocalCache(1000, 5*time.Minute),
		stats:  &CacheStats{},
	}
}

// NewLocalCache 创建本地缓存
func NewLocalCache(maxSize int, ttl time.Duration) *LocalCache {
	return &LocalCache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Connect 连接Redis
func (rc *RedisCache) Connect(ctx context.Context) error {
	rc.client = redis.NewClient(&redis.Options{
		Addr:         rc.config.Addr,
		Password:     rc.config.Password,
		DB:           rc.config.DB,
		PoolSize:     rc.config.PoolSize,
		MinIdleConns: rc.config.MinIdleConn,
		MaxRetries:   rc.config.MaxRetries,
		DialTimeout:  rc.config.DialTimeout,
		ReadTimeout:  rc.config.ReadTimeout,
		WriteTimeout: rc.config.WriteTimeout,
		PoolTimeout:  rc.config.PoolTimeout,
		ConnMaxIdleTime: rc.config.IdleTimeout,
	})

	// 测试连接
	_, err := rc.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	rc.mu.Lock()
	rc.running = true
	rc.mu.Unlock()

	// 启动后台清理
	go rc.cleanupLoop(ctx)

	return nil
}

// Disconnect 断开连接
func (rc *RedisCache) Disconnect() {
	rc.mu.Lock()
	rc.running = false
	rc.mu.Unlock()

	if rc.client != nil {
		rc.client.Close()
	}
}

// Get 获取缓存
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := rc.prefixKey(key)

	// 先查本地缓存(L1)
	if item := rc.local.Get(fullKey); item != nil {
		rc.mu.Lock()
		rc.stats.TotalHits++
		rc.mu.Unlock()
		return item.Value, nil
	}

	// 查Redis(L2)
	data, err := rc.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		rc.mu.Lock()
		rc.stats.TotalMisses++
		rc.mu.Unlock()
		return nil, fmt.Errorf("缓存不存在: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("读取缓存失败: %w", err)
	}

	// 反序列化
	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	// 存入本地缓存
	rc.local.Set(fullKey, value, rc.local.ttl)

	rc.mu.Lock()
	rc.stats.TotalHits++
	rc.mu.Unlock()

	return value, nil
}

// Set 设置缓存
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := rc.prefixKey(key)

	// 序列化
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 存入Redis
	if ttl == 0 {
		ttl = 30 * time.Minute
	}
	err = rc.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("写入Redis失败: %w", err)
	}

	// 存入本地缓存
	rc.local.Set(fullKey, value, ttl)

	rc.mu.Lock()
	rc.stats.TotalKeys++
	rc.mu.Unlock()

	return nil
}

// Delete 删除缓存
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.prefixKey(key)

	// 从Redis删除
	err := rc.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	// 从本地缓存删除
	rc.local.Delete(fullKey)

	rc.mu.Lock()
	rc.stats.TotalKeys--
	rc.mu.Unlock()

	return nil
}

// Exists 检查是否存在
func (rc *RedisCache) Exists(ctx context.Context, key string) bool {
	fullKey := rc.prefixKey(key)

	// 先查本地
	if rc.local.Exists(fullKey) {
		return true
	}

	// 查Redis
	count, err := rc.client.Exists(ctx, fullKey).Result()
	return err == nil && count > 0
}

// GetMulti 批量获取
func (rc *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = rc.prefixKey(k)
	}

	// 批量查询Redis
	vals, err := rc.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("批量获取失败: %w", err)
	}

	for i, val := range vals {
		if val == nil {
			continue
		}
		var value interface{}
		switch v := val.(type) {
		case string:
			json.Unmarshal([]byte(v), &value)
		default:
			value = v
		}
		results[keys[i]] = value
	}

	return results, nil
}

// SetMulti 批量设置
func (rc *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := rc.client.Pipeline()
	for key, value := range items {
		fullKey := rc.prefixKey(key)
		data, err := json.Marshal(value)
		if err != nil {
			continue
		}
		pipe.Set(ctx, fullKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量设置失败: %w", err)
	}

	rc.mu.Lock()
	rc.stats.TotalKeys += int64(len(items))
	rc.mu.Unlock()

	return nil
}

// DeleteMulti 批量删除
func (rc *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	fullKeys := make([]string, len(keys))
	for i, k := range keys {
		fullKeys[i] = rc.prefixKey(k)
	}

	err := rc.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("批量删除失败: %w", err)
	}

	rc.mu.Lock()
	rc.stats.TotalKeys -= int64(len(keys))
	rc.mu.Unlock()

	return nil
}

// Increment 计数器增加
func (rc *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := rc.prefixKey(key)
	return rc.client.IncrBy(ctx, fullKey, delta).Result()
}

// Decrement 计数器减少
func (rc *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := rc.prefixKey(key)
	return rc.client.DecrBy(ctx, fullKey, delta).Result()
}

// SetNX 设置(不存在时)
func (rc *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	fullKey := rc.prefixKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return rc.client.SetNX(ctx, fullKey, data, ttl).Result()
}

// GetSet 获取并设置新值
func (rc *RedisCache) GetSet(ctx context.Context, key string, value interface{}) (interface{}, error) {
	fullKey := rc.prefixKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	oldData, err := rc.client.GetSet(ctx, fullKey, data).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var oldValue interface{}
	json.Unmarshal([]byte(oldData), &oldValue)
	return oldValue, nil
}

// Expire 设置过期时间
func (rc *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := rc.prefixKey(key)
	return rc.client.Expire(ctx, fullKey, ttl).Err()
}

// TTL 获取剩余过期时间
func (rc *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := rc.prefixKey(key)
	return rc.client.TTL(ctx, fullKey).Result()
}

// === 会话管理 ===

// CreateSession 创建会话
func (rc *RedisCache) CreateSession(ctx context.Context, session *Session) error {
	if session.ID == "" {
		session.ID = fmt.Sprintf("sess-%d", time.Now().UnixNano())
	}
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	key := rc.sessionKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := session.ExpiresAt.Sub(session.CreatedAt)
	return rc.client.Set(ctx, key, data, ttl).Err()
}

// GetSession 获取会话
func (rc *RedisCache) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := rc.sessionKey(sessionID)
	data, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("会话不存在: %s", sessionID)
	}
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSession 更新会话
func (rc *RedisCache) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) error {
	session, err := rc.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// 应用更新
	for k, v := range updates {
		session.Data[k] = v
	}
	session.UpdatedAt = time.Now()

	key := rc.sessionKey(sessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := session.ExpiresAt.Sub(time.Now())
	return rc.client.Set(ctx, key, data, ttl).Err()
}

// DeleteSession 删除会话
func (rc *RedisCache) DeleteSession(ctx context.Context, sessionID string) error {
	key := rc.sessionKey(sessionID)
	return rc.client.Del(ctx, key).Err()
}

// RefreshSession 刷新会话
func (rc *RedisCache) RefreshSession(ctx context.Context, sessionID string, extend time.Duration) error {
	key := rc.sessionKey(sessionID)
	session, err := rc.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = session.ExpiresAt.Add(extend)
	session.UpdatedAt = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := session.ExpiresAt.Sub(time.Now())
	return rc.client.Set(ctx, key, data, ttl).Err()
}

// === 消息缓存 ===

// CacheMessage 缓存消息
func (rc *RedisCache) CacheMessage(ctx context.Context, msgID string, data []byte, ttl time.Duration) error {
	key := rc.msgKey(msgID)
	return rc.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedMessage 获取缓存消息
func (rc *RedisCache) GetCachedMessage(ctx context.Context, msgID string) ([]byte, error) {
	key := rc.msgKey(msgID)
	data, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("消息缓存不存在: %s", msgID)
	}
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// CachePendingMessages 缓存待发送消息列表
func (rc *RedisCache) CachePendingMessages(ctx context.Context, agentID string, msgIDs []string) error {
	key := rc.pendingKey(agentID)
	data, err := json.Marshal(msgIDs)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetPendingMessages 获取待发送消息列表
func (rc *RedisCache) GetPendingMessages(ctx context.Context, agentID string) ([]string, error) {
	key := rc.pendingKey(agentID)
	data, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var msgIDs []string
	json.Unmarshal([]byte(data), &msgIDs)
	return msgIDs, nil
}

// === 辅助方法 ===

func (rc *RedisCache) prefixKey(key string) string {
	if rc.config.KeyPrefix != "" {
		return rc.config.KeyPrefix + ":" + key
	}
	return key
}

func (rc *RedisCache) sessionKey(sessionID string) string {
	return rc.prefixKey("session:" + sessionID)
}

func (rc *RedisCache) msgKey(msgID string) string {
	return rc.prefixKey("msg:" + msgID)
}

func (rc *RedisCache) pendingKey(agentID string) string {
	return rc.prefixKey("pending:" + agentID)
}

// cleanupLoop 清理循环
func (rc *RedisCache) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rc.local.cleanup()
		}
	}
}

// GetStatistics 获取统计信息
func (rc *RedisCache) GetStatistics(ctx context.Context) map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	hitRate := 0.0
	total := rc.stats.TotalHits + rc.stats.TotalMisses
	if total > 0 {
		hitRate = float64(rc.stats.TotalHits) / float64(total)
	}

	// Redis信息
	var connInfo map[string]interface{}
	if rc.client != nil {
		info, err := rc.client.Info(ctx, "stats").Result()
		if err == nil {
			connInfo = parseRedisInfo(info)
		}
	}

	return map[string]interface{}{
		"total_keys":   rc.stats.TotalKeys,
		"total_hits":   rc.stats.TotalHits,
		"total_misses": rc.stats.TotalMisses,
		"hit_rate":     hitRate,
		"expired_keys": rc.stats.ExpiredKeys,
		"evicted_keys": rc.stats.EvictedKeys,
		"avg_latency":  rc.stats.AvgLatency,
		"redis_info":   connInfo,
		"local_cache": map[string]interface{}{
			"size":     len(rc.local.items),
			"max_size": rc.local.maxSize,
		},
	}
}

// parseRedisInfo 解析Redis INFO
func parseRedisInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// Clear 清空缓存
func (rc *RedisCache) Clear(ctx context.Context) error {
	// 清空本地缓存
	rc.local.Clear()

	// 清空Redis (谨慎使用)
	// rc.client.FlushDB(ctx)

	rc.mu.Lock()
	rc.stats = &CacheStats{}
	rc.mu.Unlock()

	return nil
}

// === 本地缓存方法 ===

func (lc *LocalCache) Get(key string) *CacheItem {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	item, ok := lc.items[key]
	if !ok {
		return nil
	}

	// 检查过期
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return nil
	}

	item.Hits++
	return item
}

func (lc *LocalCache) Set(key string, value interface{}, ttl time.Duration) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// 检查容量
	if len(lc.items) >= lc.maxSize {
		lc.evictOldest()
	}

	expiresAt := time.Now().Add(ttl)
	lc.items[key] = &CacheItem{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		CreatedAt: time.Now(),
		ExpiresAt: &expiresAt,
	}
}

func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.items, key)
}

func (lc *LocalCache) Exists(key string) bool {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	item, ok := lc.items[key]
	if !ok {
		return false
	}
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return false
	}
	return true
}

func (lc *LocalCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for k, item := range lc.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(lc.items, oldestKey)
	}
}

func (lc *LocalCache) cleanup() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	now := time.Now()
	for k, item := range lc.items {
		if item.ExpiresAt != nil && now.After(*item.ExpiresAt) {
			delete(lc.items, k)
		}
	}
}

func (lc *LocalCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.items = make(map[string]*CacheItem)
}

// IsConnected 检查连接状态
func (rc *RedisCache) IsConnected() bool {
	if rc.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := rc.client.Ping(ctx).Result()
	return err == nil
}

// GetClient 获取Redis客户端
func (rc *RedisCache) GetClient() *redis.Client {
	return rc.client
}