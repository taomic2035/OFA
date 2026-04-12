// advanced_cache.go
// OFA Center Advanced Cache System (v9.2.0)

package cache

import (
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Key        string
	Value      interface{}
	ExpiresAt  time.Time
	AccessCount int
	LastAccess time.Time
	Size       int
}

// CachePolicy 缓存策略
type CachePolicy string

const (
	CachePolicyLFU  CachePolicy = "lfu"  // 最少使用
	CachePolicyLRU  CachePolicy = "lru"  // 最近最少使用
	CachePolicyTTL  CachePolicy = "ttl"  // 时间过期
	CachePolicyFIFO CachePolicy = "fifo" // 先进先出
)

// TieredCache 分层缓存
type TieredCache struct {
	l1        *LocalCache    // 本地缓存 (L1)
	l2        *RedisCache    // Redis缓存 (L2)
	config    TieredCacheConfig
	mu        sync.RWMutex
}

// TieredCacheConfig 分层缓存配置
type TieredCacheConfig struct {
	L1MaxSize     int
	L1TTL         time.Duration
	L2TTL         time.Duration
	L2Enabled     bool
	WritePolicy   WritePolicy   // 写策略
	ReadPolicy    ReadPolicy    // 读策略
}

// WritePolicy 写策略
type WritePolicy string

const (
	WritePolicyWriteThrough  WritePolicy = "write_through"  // 同时写入L1和L2
	WritePolicyWriteBack     WritePolicy = "write_back"     // 先写L1，异步写L2
	WritePolicyWriteAround   WritePolicy = "write_around"   // 只写L2，读取时更新L1
)

// ReadPolicy 读策略
type ReadPolicy string

const (
	ReadPolicyLookThrough   ReadPolicy = "look_through"   // L1 miss -> L2
	ReadPolicyLookAside     ReadPolicy = "look_aside"     // 同时查询L1和L2
)

// DefaultTieredCacheConfig 默认分层缓存配置
func DefaultTieredCacheConfig() TieredCacheConfig {
	return TieredCacheConfig{
		L1MaxSize:     1000,
		L1TTL:         5 * time.Minute,
		L2TTL:         30 * time.Minute,
		L2Enabled:     true,
		WritePolicy:   WritePolicyWriteThrough,
		ReadPolicy:    ReadPolicyLookThrough,
	}
}

// NewTieredCache 创建分层缓存
func NewTieredCache(config TieredCacheConfig) *TieredCache {
	tc := &TieredCache{
		config: config,
	}

	tc.l1 = NewLocalCache(config.L1MaxSize, config.L1TTL, CachePolicyLFU)

	if config.L2Enabled {
		tc.l2 = NewRedisCache(config.L2TTL)
	}

	return tc
}

// Get 获取缓存
func (tc *TieredCache) Get(key string) (interface{}, bool) {
	// 先查L1
	value, found := tc.l1.Get(key)
	if found {
		return value, true
	}

	// L1 miss，查L2
	if tc.l2 != nil {
		value, found = tc.l2.Get(key)
		if found {
			// 更新L1
			tc.l1.Set(key, value)
			return value, true
		}
	}

	return nil, false
}

// Set 设置缓存
func (tc *TieredCache) Set(key string, value interface{}) {
	switch tc.config.WritePolicy {
	case WritePolicyWriteThrough:
		// 同时写入L1和L2
		tc.l1.Set(key, value)
		if tc.l2 != nil {
			tc.l2.Set(key, value)
		}
	case WritePolicyWriteBack:
		// 先写L1，异步写L2
		tc.l1.Set(key, value)
		if tc.l2 != nil {
			go tc.l2.Set(key, value)
		}
	case WritePolicyWriteAround:
		// 只写L2
		if tc.l2 != nil {
			tc.l2.Set(key, value)
		}
	}
}

// Delete 删除缓存
func (tc *TieredCache) Delete(key string) {
	tc.l1.Delete(key)
	if tc.l2 != nil {
		tc.l2.Delete(key)
	}
}

// Clear 清除所有缓存
func (tc *TieredCache) Clear() {
	tc.l1.Clear()
	if tc.l2 != nil {
		tc.l2.Clear()
	}
}

// Stats 获取缓存统计
func (tc *TieredCache) Stats() CacheStats {
	l1Stats := tc.l1.Stats()
	stats := CacheStats{
		L1Hits:    l1Stats.Hits,
		L1Misses:  l1Stats.Misses,
		L1Size:    l1Stats.Size,
		L1MaxSize: l1Stats.MaxSize,
	}

	if tc.l2 != nil {
		l2Stats := tc.l2.Stats()
		stats.L2Hits = l2Stats.Hits
		stats.L2Misses = l2Stats.Misses
	}

	stats.HitRate = float64(stats.L1Hits) / float64(stats.L1Hits+stats.L1Misses+1)

	return stats
}

// CacheStats 缓存统计
type CacheStats struct {
	L1Hits    int64
	L1Misses  int64
	L2Hits    int64
	L2Misses  int64
	L1Size    int
	L1MaxSize int
	HitRate   float64
}

// LocalCache 本地缓存
type LocalCache struct {
	maxSize   int
	ttl       time.Duration
	policy    CachePolicy
	items     map[string]*CacheItem
	order     []string // 用于FIFO/LRU顺序
	hits      int64
	misses    int64
	mu        sync.RWMutex
}

// NewLocalCache 创建本地缓存
func NewLocalCache(maxSize int, ttl time.Duration, policy CachePolicy) *LocalCache {
	lc := &LocalCache{
		maxSize: maxSize,
		ttl:     ttl,
		policy:  policy,
		items:   make(map[string]*CacheItem),
		order:   make([]string, 0),
	}

	// 启动过期清理
	go lc.cleanupExpired()

	return lc
}

// Get 获取缓存
func (lc *LocalCache) Get(key string) (interface{}, bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	item, found := lc.items[key]
	if !found {
		lc.misses++
		return nil, false
	}

	// 检查过期
	if time.Now().After(item.ExpiresAt) {
		lc.deleteInternal(key)
		lc.misses++
		return nil, false
	}

	// 更新访问信息
	item.AccessCount++
	item.LastAccess = time.Now()

	lc.hits++
	return item.Value, true
}

// Set 设置缓存
func (lc *LocalCache) Set(key string, value interface{}) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// 检查是否需要驱逐
	if len(lc.items) >= lc.maxSize {
		lc.evict()
	}

	item := &CacheItem{
		Key:         key,
		Value:       value,
		ExpiresAt:   time.Now().Add(lc.ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	lc.items[key] = item
	lc.order = append(lc.order, key)
}

// Delete 删除缓存
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	lc.deleteInternal(key)
	lc.mu.Unlock()
}

// deleteInternal 内部删除方法（不加锁）
func (lc *LocalCache) deleteInternal(key string) {
	delete(lc.items, key)
	// 从order中移除
	for i, k := range lc.order {
		if k == key {
			lc.order = append(lc.order[:i], lc.order[i+1:]...)
			break
		}
	}
}

// evict 驱逐缓存
func (lc *LocalCache) evict() {
	if len(lc.items) == 0 {
		return
	}

	var evictKey string

	switch lc.policy {
	case CachePolicyLFU:
		// 驱逐访问次数最少的
		minAccess := -1
		for key, item := range lc.items {
			if minAccess == -1 || item.AccessCount < minAccess {
				minAccess = item.AccessCount
				evictKey = key
			}
		}
	case CachePolicyLRU:
		// 驱逐最近访问时间最早的
	 oldestAccess := time.Now().Add(24 * time.Hour)
		for key, item := range lc.items {
			if item.LastAccess.Before(oldestAccess) {
				oldestAccess = item.LastAccess
				evictKey = key
			}
		}
	case CachePolicyFIFO:
		// 驱逐最早加入的
		if len(lc.order) > 0 {
			evictKey = lc.order[0]
		}
	case CachePolicyTTL:
		// 驱逐即将过期的
		nearestExpiry := time.Now().Add(24 * time.Hour)
		for key, item := range lc.items {
			if item.ExpiresAt.Before(nearestExpiry) {
				nearestExpiry = item.ExpiresAt
				evictKey = key
			}
		}
	}

	if evictKey != "" {
		lc.deleteInternal(evictKey)
	}
}

// cleanupExpired 清理过期缓存
func (lc *LocalCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		lc.mu.Lock()
		now := time.Now()
		for key, item := range lc.items {
			if now.After(item.ExpiresAt) {
				lc.deleteInternal(key)
			}
		}
		lc.mu.Unlock()
	}
}

// Clear 清除所有缓存
func (lc *LocalCache) Clear() {
	lc.mu.Lock()
	lc.items = make(map[string]*CacheItem)
	lc.order = make([]string, 0)
	lc.mu.Unlock()
}

// Stats 获取缓存统计
func (lc *LocalCache) Stats() LocalCacheStats {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return LocalCacheStats{
		Hits:    lc.hits,
		Misses:  lc.misses,
		Size:    len(lc.items),
		MaxSize: lc.maxSize,
	}
}

// LocalCacheStats 本地缓存统计
type LocalCacheStats struct {
	Hits    int64
	Misses  int64
	Size    int
	MaxSize int
}

// RedisCache Redis缓存模拟
type RedisCache struct {
	ttl    time.Duration
	items  map[string]*CacheItem
	hits   int64
	misses int64
	mu     sync.RWMutex
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(ttl time.Duration) *RedisCache {
	return &RedisCache{
		ttl:   ttl,
		items: make(map[string]*CacheItem),
	}
}

// Get 获取缓存
func (rc *RedisCache) Get(key string) (interface{}, bool) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	item, found := rc.items[key]
	if !found {
		rc.misses++
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		delete(rc.items, key)
		rc.misses++
		return nil, false
	}

	rc.hits++
	return item.Value, true
}

// Set 设置缓存
func (rc *RedisCache) Set(key string, value interface{}) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.items[key] = &CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(rc.ttl),
	}
}

// Delete 删除缓存
func (rc *RedisCache) Delete(key string) {
	rc.mu.Lock()
	delete(rc.items, key)
	rc.mu.Unlock()
}

// Clear 清除所有缓存
func (rc *RedisCache) Clear() {
	rc.mu.Lock()
	rc.items = make(map[string]*CacheItem)
	rc.mu.Unlock()
}

// Stats 获取缓存统计
func (rc *RedisCache) Stats() LocalCacheStats {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return LocalCacheStats{
		Hits:   rc.hits,
		Misses: rc.misses,
		Size:   len(rc.items),
	}
}

// HotKeyCache 热点数据缓存
type HotKeyCache struct {
	baseCache   *TieredCache
	hotKeys     map[string]bool
	threshold   int           // 热点阈值
	preloadKeys []string      // 预加载键
	mu          sync.RWMutex
}

// NewHotKeyCache 创建热点数据缓存
func NewHotKeyCache(baseCache *TieredCache, threshold int) *HotKeyCache {
	return &HotKeyCache{
		baseCache: baseCache,
		hotKeys:   make(map[string]bool),
		threshold: threshold,
	}
}

// Get 获取缓存（自动识别热点）
func (hc *HotKeyCache) Get(key string) (interface{}, bool) {
	value, found := hc.baseCache.Get(key)

	if found {
		// 更新热点状态
		hc.mu.Lock()
		stats := hc.baseCache.l1.Stats()
		// 简化的热点判断：如果命中次数超过阈值
		if stats.Hits > int64(hc.threshold) {
			hc.hotKeys[key] = true
		}
		hc.mu.Unlock()
	}

	return value, found
}

// Set 设置缓存
func (hc *HotKeyCache) Set(key string, value interface{}) {
	hc.baseCache.Set(key, value)
}

// Preload 预加载热点数据
func (hc *HotKeyCache) Preload(keys []string, loader func(key string) (interface{}, error)) {
	for _, key := range keys {
		value, err := loader(key)
		if err == nil {
			hc.Set(key, value)
			hc.mu.Lock()
			hc.hotKeys[key] = true
			hc.mu.Unlock()
		}
	}
}

// GetHotKeys 获取热点键列表
func (hc *HotKeyCache) GetHotKeys() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	keys := make([]string, 0)
	for key := range hc.hotKeys {
		keys = append(keys, key)
	}
	return keys
}

// IsHotKey 检查是否为热点键
func (hc *HotKeyCache) IsHotKey(key string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.hotKeys[key]
}

// RequestCache 按请求路径的缓存
type RequestCache struct {
	cache    *TieredCache
	patterns map[string]bool // 可缓存的路径模式
	mu       sync.RWMutex
}

// NewRequestCache 创建请求缓存
func NewRequestCache(cache *TieredCache) *RequestCache {
	return &RequestCache{
		cache:    cache,
		patterns: make(map[string]bool),
	}
}

// AddCacheablePattern 添加可缓存模式
func (rc *RequestCache) AddCacheablePattern(pattern string) {
	rc.mu.Lock()
	rc.patterns[pattern] = true
	rc.mu.Unlock()
}

// IsCacheable 检查路径是否可缓存
func (rc *RequestCache) IsCacheable(path string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.patterns[path]
}

// Get 获取请求缓存
func (rc *RequestCache) Get(method, path string) (interface{}, bool) {
	key := method + ":" + path
	return rc.cache.Get(key)
}

// Set 设置请求缓存
func (rc *RequestCache) Set(method, path string, value interface{}) {
	if rc.IsCacheable(path) {
		key := method + ":" + path
		rc.cache.Set(key, value)
	}
}

// IdentityCache 身份专用缓存
type IdentityCache struct {
	cache    *TieredCache
	profiles map[string]IdentityProfile
	mu       sync.RWMutex
}

// IdentityProfile 身份缓存项
type IdentityProfile struct {
	ID        string
	Name      string
	Personality interface{}
	Values      interface{}
	ExpiresAt   time.Time
}

// NewIdentityCache 创建身份缓存
func NewIdentityCache(cache *TieredCache) *IdentityCache {
	return &IdentityCache{
		cache:    cache,
		profiles: make(map[string]IdentityProfile),
	}
}

// GetIdentity 获取身份缓存
func (ic *IdentityCache) GetIdentity(id string) (*IdentityProfile, bool) {
	value, found := ic.cache.Get("identity:" + id)
	if found {
		profile, ok := value.(IdentityProfile)
		if ok {
			return &profile, true
		}
	}
	return nil, false
}

// SetIdentity 设置身份缓存
func (ic *IdentityCache) SetIdentity(profile IdentityProfile) {
	ic.cache.Set("identity:"+profile.ID, profile)
	ic.mu.Lock()
	ic.profiles[profile.ID] = profile
	ic.mu.Unlock()
}

// InvalidateIdentity 使身份缓存失效
func (ic *IdentityCache) InvalidateIdentity(id string) {
	ic.cache.Delete("identity:" + id)
	ic.mu.Lock()
	delete(ic.profiles, id)
	ic.mu.Unlock()
}

// GetAllIdentityIDs 获取所有身份ID
func (ic *IdentityCache) GetAllIdentityIDs() []string {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	ids := make([]string, 0)
	for id := range ic.profiles {
		ids = append(ids, id)
	}
	return ids
}