// cache_test.go
// OFA Center Advanced Cache Tests (v9.2.0)

package cache

import (
	"testing"
	"time"
)

func TestNewTieredCache(t *testing.T) {
	config := DefaultTieredCacheConfig()
	tc := NewTieredCache(config)

	if tc == nil {
		t.Fatal("TieredCache should not be nil")
	}
	if tc.l1 == nil {
		t.Error("L1 cache should not be nil")
	}
	if tc.l2 == nil {
		t.Error("L2 cache should not be nil")
	}
}

func TestTieredCacheGetSet(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())

	// 设置缓存
	tc.Set("key1", "value1")

	// 获取缓存
	value, found := tc.Get("key1")
	if !found {
		t.Error("Cache should be found")
	}
	if value != "value1" {
		t.Errorf("Value should be 'value1', got '%v'", value)
	}
}

func TestTieredCacheL1Miss(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())

	// 直接在L2设置缓存
	tc.l2.Set("key1", "value1")

	// L1应该miss，L2应该hit
	value, found := tc.Get("key1")
	if !found {
		t.Error("Cache should be found in L2")
	}
	if value != "value1" {
		t.Errorf("Value should be 'value1'")
	}

	// L1应该被更新
	_, l1Found := tc.l1.Get("key1")
	if !l1Found {
		t.Error("L1 should be updated after L2 hit")
	}
}

func TestTieredCacheDelete(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())

	tc.Set("key1", "value1")
	tc.Delete("key1")

	_, found := tc.Get("key1")
	if found {
		t.Error("Cache should be deleted")
	}
}

func TestTieredCacheClear(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())

	tc.Set("key1", "value1")
	tc.Set("key2", "value2")
	tc.Clear()

	stats := tc.Stats()
	if stats.L1Size != 0 {
		t.Errorf("L1 should be empty after clear")
	}
}

func TestTieredCacheStats(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())

	tc.Set("key1", "value1")
	tc.Get("key1") // hit
	tc.Get("key2") // miss

	stats := tc.Stats()
	if stats.L1Hits != 1 {
		t.Errorf("L1Hits should be 1, got %d", stats.L1Hits)
	}
	if stats.L1Misses != 1 {
		t.Errorf("L1Misses should be 1, got %d", stats.L1Misses)
	}
}

func TestLocalCacheLFU(t *testing.T) {
	lc := NewLocalCache(3, 5*time.Minute, CachePolicyLFU)

	lc.Set("key1", "value1")
	lc.Set("key2", "value2")
	lc.Set("key3", "value3")

	// 访问key1和key2多次
	for i := 0; i < 5; i++ {
		lc.Get("key1")
		lc.Get("key2")
	}
	// key3访问次数最少

	// 添加新键，应该驱逐key3
	lc.Set("key4", "value4")

	_, found := lc.Get("key3")
	if found {
		t.Error("key3 should be evicted (LFU)")
	}
}

func TestLocalCacheLRU(t *testing.T) {
	lc := NewLocalCache(3, 5*time.Minute, CachePolicyLRU)

	lc.Set("key1", "value1")
	lc.Set("key2", "value2")
	lc.Set("key3", "value3")

	// 最近访问key1和key2
	time.Sleep(10 * time.Millisecond)
	lc.Get("key1")
	time.Sleep(10 * time.Millisecond)
	lc.Get("key2")

	// key3最近访问时间最早

	// 添加新键，应该驱逐key3
	lc.Set("key4", "value4")

	_, found := lc.Get("key3")
	if found {
		t.Error("key3 should be evicted (LRU)")
	}
}

func TestLocalCacheFIFO(t *testing.T) {
	lc := NewLocalCache(3, 5*time.Minute, CachePolicyFIFO)

	lc.Set("key1", "value1")
	lc.Set("key2", "value2")
	lc.Set("key3", "value3")

	// key1最早加入

	// 添加新键，应该驱逐key1
	lc.Set("key4", "value4")

	_, found := lc.Get("key1")
	if found {
		t.Error("key1 should be evicted (FIFO)")
	}
}

func TestLocalCacheTTL(t *testing.T) {
	lc := NewLocalCache(100, 100*time.Millisecond, CachePolicyLFU)

	lc.Set("key1", "value1")

	// 立即获取应该成功
	_, found := lc.Get("key1")
	if !found {
		t.Error("Cache should exist immediately")
	}

	// 等待过期
	time.Sleep(200 * time.Millisecond)

	// 应该过期
	_, found = lc.Get("key1")
	if found {
		t.Error("Cache should be expired")
	}
}

func TestLocalCacheStats(t *testing.T) {
	lc := NewLocalCache(100, 5*time.Minute, CachePolicyLFU)

	lc.Set("key1", "value1")
	lc.Get("key1") // hit
	lc.Get("key2") // miss

	stats := lc.Stats()
	if stats.Hits != 1 {
		t.Errorf("Hits should be 1, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses should be 1, got %d", stats.Misses)
	}
}

func TestRedisCache(t *testing.T) {
	rc := NewRedisCache(5 * time.Minute)

	rc.Set("key1", "value1")

	value, found := rc.Get("key1")
	if !found {
		t.Error("Cache should be found")
	}
	if value != "value1" {
		t.Errorf("Value should be 'value1'")
	}

	rc.Delete("key1")
	_, found = rc.Get("key1")
	if found {
		t.Error("Cache should be deleted")
	}
}

func TestRedisCacheExpiration(t *testing.T) {
	rc := NewRedisCache(100 * time.Millisecond)

	rc.Set("key1", "value1")

	// 等待过期
	time.Sleep(200 * time.Millisecond)

	_, found := rc.Get("key1")
	if found {
		t.Error("Cache should be expired")
	}
}

func TestHotKeyCache(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())
	hc := NewHotKeyCache(tc, 3)

	// 设置缓存
	hc.Set("hotkey1", "value1")

	// 多次访问使其成为热点
	for i := 0; i < 5; i++ {
		hc.Get("hotkey1")
	}

	// 检查热点状态
	if !hc.IsHotKey("hotkey1") {
		t.Error("hotkey1 should be hot after multiple accesses")
	}
}

func TestHotKeyCachePreload(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())
	hc := NewHotKeyCache(tc, 3)

	// 预加载热点数据
	keys := []string{"preload1", "preload2"}
	hc.Preload(keys, func(key string) (interface{}, error) {
		return "preloaded_" + key, nil
	})

	// 检查预加载
	for _, key := range keys {
		if !hc.IsHotKey(key) {
			t.Errorf("Preloaded key %s should be hot", key)
		}
	}
}

func TestRequestCache(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())
	rc := NewRequestCache(tc)

	// 添加可缓存模式
	rc.AddCacheablePattern("/api/v1/identity")
	rc.AddCacheablePattern("/api/v1/profile")

	// 设置可缓存请求
	rc.Set("GET", "/api/v1/identity", "identity_data")

	// 应该能获取到
	value, found := rc.Get("GET", "/api/v1/identity")
	if !found {
		t.Error("Cache should be found")
	}
	if value != "identity_data" {
		t.Errorf("Value should be 'identity_data'")
	}

	// 不可缓存的路径不应该被缓存
	rc.Set("GET", "/api/v1/uncacheable", "data")
	_, found = rc.Get("GET", "/api/v1/uncacheable")
	if found {
		t.Error("Uncacheable path should not be cached")
	}
}

func TestIdentityCache(t *testing.T) {
	tc := NewTieredCache(DefaultTieredCacheConfig())
	ic := NewIdentityCache(tc)

	// 设置身份缓存
	profile := IdentityProfile{
		ID:      "user1",
		Name:    "Test User",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	ic.SetIdentity(profile)

	// 获取身份缓存
	p, found := ic.GetIdentity("user1")
	if !found {
		t.Error("Identity should be found")
	}
	if p.Name != "Test User" {
		t.Errorf("Name should be 'Test User'")
	}

	// 使身份失效
	ic.InvalidateIdentity("user1")
	_, found = ic.GetIdentity("user1")
	if found {
		t.Error("Identity should be invalidated")
	}
}

func TestWritePolicyWriteThrough(t *testing.T) {
	config := DefaultTieredCacheConfig()
	config.WritePolicy = WritePolicyWriteThrough
	tc := NewTieredCache(config)

	tc.Set("key1", "value1")

	// L1和L2应该都有数据
	_, l1Found := tc.l1.Get("key1")
	_, l2Found := tc.l2.Get("key1")

	if !l1Found {
		t.Error("L1 should have data (write-through)")
	}
	if !l2Found {
		t.Error("L2 should have data (write-through)")
	}
}

func TestWritePolicyWriteAround(t *testing.T) {
	config := DefaultTieredCacheConfig()
	config.WritePolicy = WritePolicyWriteAround
	tc := NewTieredCache(config)

	tc.Set("key1", "value1")

	// 只有L2应该有数据
	_, l1Found := tc.l1.Get("key1")
	_, l2Found := tc.l2.Get("key1")

	if l1Found {
		t.Error("L1 should not have data (write-around)")
	}
	if !l2Found {
		t.Error("L2 should have data (write-around)")
	}

	// 读取后L1应该被更新
	tc.Get("key1")
	_, l1FoundAfter := tc.l1.Get("key1")
	if !l1FoundAfter {
		t.Error("L1 should be updated after read (write-around)")
	}
}

func TestTieredCacheWithoutL2(t *testing.T) {
	config := DefaultTieredCacheConfig()
	config.L2Enabled = false
	tc := NewTieredCache(config)

	if tc.l2 != nil {
		t.Error("L2 should be nil when disabled")
	}

	tc.Set("key1", "value1")
	_, found := tc.Get("key1")
	if !found {
		t.Error("Cache should still work without L2")
	}
}