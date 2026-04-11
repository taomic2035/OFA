package cache

import (
	"context"
	"testing"
	"time"
)

// TestLocalCachePerformance 测试本地缓存性能
func TestLocalCachePerformance(t *testing.T) {
	cache := NewLocalCache(10000, 5*time.Minute)

	// 写入性能测试
	t.Run("SetPerformance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 10000; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(key, fmt.Sprintf("value_%d", i), 5*time.Minute)
		}
		duration := time.Since(start)

		opsPerSec := float64(10000) / duration.Seconds()
		t.Logf("Set性能: %d 操作耗时 %v, %.2f ops/sec", 10000, duration, opsPerSec)

		if opsPerSec < 100000 {
			t.Logf("警告: Set性能低于100000 ops/sec")
		}
	})

	//读取性能测试
	t.Run("GetPerformance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 10000; i++ {
			key := fmt.Sprintf("key_%d", i)
			item := cache.Get(key)
			if item == nil {
				t.Errorf("缓存项 %s 不存在", key)
			}
		}
		duration := time.Since(start)

		opsPerSec := float64(10000) / duration.Seconds()
		t.Logf("Get性能: %d 操作耗时 %v, %.2f ops/sec", 10000, duration, opsPerSec)

		if opsPerSec < 100000 {
			t.Logf("警告: Get性能低于100000 ops/sec")
		}
	})

	// 并发读写测试
	t.Run("ConcurrentPerformance", func(t *testing.T) {
		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < 100; i++ {
			wg.Add(2)

			// 写协程
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					key := fmt.Sprintf("concurrent_%d_%d", id, j)
					cache.Set(key, j, 5*time.Minute)
				}
			}(i)

			// 读协程
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					key := fmt.Sprintf("concurrent_%d_%d", id, j)
					cache.Get(key)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		totalOps := 100 * 100 * 2 // 100个协程,每个100次,读写各一次
		opsPerSec := float64(totalOps) / duration.Seconds()
		t.Logf("并发性能: %d 操作耗时 %v, %.2f ops/sec", totalOps, duration, opsPerSec)
	})
}

// BenchmarkLocalCacheSet 本地缓存写入基准测试
func BenchmarkLocalCacheSet(b *testing.B) {
	cache := NewLocalCache(b.N, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(key, i, 5*time.Minute)
	}
}

// BenchmarkLocalCacheGet 本地缓存读取基准测试
func BenchmarkLocalCacheGet(b *testing.B) {
	cache := NewLocalCache(b.N, 5*time.Minute)

	// 预填充
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Set(key, i, 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		cache.Get(key)
	}
}

// BenchmarkLocalCacheConcurrent 本地缓存并发基准测试
func BenchmarkLocalCacheConcurrent(b *testing.B) {
	cache := NewLocalCache(10000, 5*time.Minute)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("parallel_key_%d", i%10000)
			if i%2 == 0 {
				cache.Set(key, i, 5*time.Minute)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

// BenchmarkLocalCacheEviction 淘汰机制基准测试
func BenchmarkLocalCacheEviction(b *testing.B) {
	cache := NewLocalCache(1000, 5*time.Minute) // 小容量触发淘汰

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("evict_key_%d", i)
		cache.Set(key, i, 5*time.Minute)
	}
}

// TestCacheHitRate 测试缓存命中率
func TestCacheHitRate(t *testing.T) {
	cache := NewLocalCache(100, 5*time.Minute)

	// 预填充
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("hit_key_%d", i)
		cache.Set(key, i, 5*time.Minute)
	}

	hits := 0
	misses := 0

	// 测试命中率
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("hit_key_%d", i)
		if cache.Get(key) != nil {
			hits++
		} else {
			misses++
		}
	}

	hitRate := float64(hits) / float64(hits+misses) * 100
	t.Logf("命中率: %.2f%% (命中%d, 未命中%d)", hitRate, hits, misses)

	if hitRate < 40 {
		t.Errorf("命中率过低: %.2f%%", hitRate)
	}
}