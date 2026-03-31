// Package performance provides multi-level caching
package performance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CacheLevel defines cache levels
type CacheLevel int

const (
	CacheLevelL1 CacheLevel = iota // L1: In-memory cache (fastest, smallest)
	CacheLevelL2 CacheLevel = iota // L2: In-process cache
	CacheLevelL3 CacheLevel = iota // L3: Distributed cache (Redis)
)

// CacheItem represents a cached item
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	ExpiresAt  time.Time   `json:"expires_at"`
	CreatedAt  time.Time   `json:"created_at"`
	AccessCount int        `json:"access_count"`
	LastAccess time.Time   `json:"last_access"`
	Size       int64       `json:"size"`
	Level      CacheLevel  `json:"level"`
	TenantID   string      `json:"tenant_id,omitempty"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxSizeL1      int64         `json:"max_size_l1"`      // Max size in bytes for L1
	MaxSizeL2      int64         `json:"max_size_l2"`      // Max size in bytes for L2
	MaxSizeL3      int64         `json:"max_size_l3"`      // Max size in bytes for L3
	DefaultTTL     time.Duration `json:"default_ttl"`      // Default TTL
	CleanupInterval time.Duration `json:"cleanup_interval"` // Cleanup interval
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSizeL1:       10 * 1024 * 1024,   // 10MB
		MaxSizeL2:       100 * 1024 * 1024,  // 100MB
		MaxSizeL3:       1024 * 1024 * 1024, // 1GB
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

// MultiLevelCache provides multi-level caching
type MultiLevelCache struct {
	config CacheConfig

	// L1 Cache (hot data)
	l1 sync.Map // map[string]*CacheItem
	l1Size int64

	// L2 Cache (warm data)
	l2 sync.Map // map[string]*CacheItem
	l2Size int64

	// L3 Cache (cold data - Redis, etc.)
	l3 CacheBackend

	// Statistics
	hits    int64
	misses  int64
	sets    int64
	deletes int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// CacheBackend defines cache backend interface
type CacheBackend interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(config CacheConfig, l3 CacheBackend) *MultiLevelCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &MultiLevelCache{
		config: config,
		l3:     l3,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start cleanup routine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (c *MultiLevelCache) Get(key string) (interface{}, error) {
	// Try L1
	if v, ok := c.l1.Load(key); ok {
		item := v.(*CacheItem)
		if time.Now().Before(item.ExpiresAt) {
			c.hits++
			item.AccessCount++
			item.LastAccess = time.Now()
			return item.Value, nil
		}
		c.l1.Delete(key)
	}

	// Try L2
	if v, ok := c.l2.Load(key); ok {
		item := v.(*CacheItem)
		if time.Now().Before(item.ExpiresAt) {
			c.hits++
			item.AccessCount++
			item.LastAccess = time.Now()

			// Promote to L1
			c.promoteToL1(item)
			return item.Value, nil
		}
		c.l2.Delete(key)
	}

	// Try L3
	if c.l3 != nil {
		data, err := c.l3.Get(key)
		if err == nil {
			var item CacheItem
			if err := json.Unmarshal(data, &item); err == nil {
				c.hits++

				// Promote to L2
				c.promoteToL2(&item)
				return item.Value, nil
			}
		}
	}

	c.misses++
	return nil, errors.New("cache miss")
}

// Set stores a value in cache
func (c *MultiLevelCache) Set(key string, value interface{}, ttl time.Duration, tenantID string) error {
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	// Estimate size
	size := estimateSize(value)

	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Size:       size,
		TenantID:   tenantID,
	}

	// Determine which level to store in
	if size < c.config.MaxSizeL1/10 {
		// Small items go to L1
		c.storeL1(item)
	} else if size < c.config.MaxSizeL2/10 {
		// Medium items go to L2
		c.storeL2(item)
	}

	// Always store in L3 if available
	if c.l3 != nil {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		c.l3.Set(key, data, ttl)
	}

	c.sets++
	return nil
}

// Delete removes a value from cache
func (c *MultiLevelCache) Delete(key string) error {
	c.l1.Delete(key)
	c.l2.Delete(key)

	if c.l3 != nil {
		c.l3.Delete(key)
	}

	c.deletes++
	return nil
}

// storeL1 stores item in L1 cache
func (c *MultiLevelCache) storeL1(item *CacheItem) {
	// Evict if necessary
	for c.l1Size+item.Size > c.config.MaxSizeL1 {
		c.evictL1()
	}

	item.Level = CacheLevelL1
	c.l1.Store(item.Key, item)
	c.l1Size += item.Size
}

// storeL2 stores item in L2 cache
func (c *MultiLevelCache) storeL2(item *CacheItem) {
	// Evict if necessary
	for c.l2Size+item.Size > c.config.MaxSizeL2 {
		c.evictL2()
	}

	item.Level = CacheLevelL2
	c.l2.Store(item.Key, item)
	c.l2Size += item.Size
}

// promoteToL1 promotes item to L1
func (c *MultiLevelCache) promoteToL1(item *CacheItem) {
	// Remove from L2
	c.l2.Delete(item.Key)
	c.l2Size -= item.Size

	// Store in L1
	c.storeL1(item)
}

// promoteToL2 promotes item to L2
func (c *MultiLevelCache) promoteToL2(item *CacheItem) {
	c.storeL2(item)
}

// evictL1 evicts least recently used item from L1
func (c *MultiLevelCache) evictL1() {
	var oldestKey string
	var oldestTime time.Time
	var oldestSize int64

	c.l1.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key.(string)
			oldestTime = item.LastAccess
			oldestSize = item.Size
		}
		return true
	})

	if oldestKey != "" {
		c.l1.Delete(oldestKey)
		c.l1Size -= oldestSize

		// Demote to L2
		if v, ok := c.l1.Load(oldestKey); ok {
			c.storeL2(v.(*CacheItem))
		}
	}
}

// evictL2 evicts least recently used item from L2
func (c *MultiLevelCache) evictL2() {
	var oldestKey string
	var oldestTime time.Time
	var oldestSize int64

	c.l2.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key.(string)
			oldestTime = item.LastAccess
			oldestSize = item.Size
		}
		return true
	})

	if oldestKey != "" {
		c.l2.Delete(oldestKey)
		c.l2Size -= oldestSize
	}
}

// cleanup removes expired items
func (c *MultiLevelCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired items from all levels
func (c *MultiLevelCache) cleanupExpired() {
	now := time.Now()

	// Clean L1
	c.l1.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if now.After(item.ExpiresAt) {
			c.l1.Delete(key)
			c.l1Size -= item.Size
		}
		return true
	})

	// Clean L2
	c.l2.Range(func(key, value interface{}) bool {
		item := value.(*CacheItem)
		if now.After(item.ExpiresAt) {
			c.l2.Delete(key)
			c.l2Size -= item.Size
		}
		return true
	})
}

// GetStats returns cache statistics
func (c *MultiLevelCache) GetStats() map[string]interface{} {
	var hitRate float64
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":     c.hits,
		"misses":   c.misses,
		"sets":     c.sets,
		"deletes":  c.deletes,
		"hit_rate": fmt.Sprintf("%.2f%%", hitRate*100),
		"l1_size":  c.l1Size,
		"l2_size":  c.l2Size,
		"l1_items": countItems(&c.l1),
		"l2_items": countItems(&c.l2),
	}
}

// Clear clears all cache levels
func (c *MultiLevelCache) Clear() {
	c.l1 = sync.Map{}
	c.l2 = sync.Map{}
	c.l1Size = 0
	c.l2Size = 0
}

// Close closes the cache
func (c *MultiLevelCache) Close() {
	c.cancel()
}

// Helper functions

func estimateSize(value interface{}) int64 {
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

func countItems(m *sync.Map) int64 {
	var count int64
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// MemoryCacheBackend is an in-memory L3 backend
type MemoryCacheBackend struct {
	items sync.Map
	mu    sync.RWMutex
}

// NewMemoryCacheBackend creates a new memory cache backend
func NewMemoryCacheBackend() *MemoryCacheBackend {
	return &MemoryCacheBackend{}
}

// Get retrieves an item
func (b *MemoryCacheBackend) Get(key string) ([]byte, error) {
	if v, ok := b.items.Load(key); ok {
		return v.([]byte), nil
	}
	return nil, errors.New("not found")
}

// Set stores an item
func (b *MemoryCacheBackend) Set(key string, value []byte, ttl time.Duration) error {
	b.items.Store(key, value)
	return nil
}

// Delete removes an item
func (b *MemoryCacheBackend) Delete(key string) error {
	b.items.Delete(key)
	return nil
}

// Exists checks if item exists
func (b *MemoryCacheBackend) Exists(key string) (bool, error) {
	_, ok := b.items.Load(key)
	return ok, nil
}