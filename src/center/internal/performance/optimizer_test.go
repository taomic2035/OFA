package performance

import (
	"context"
	"testing"
	"time"
)

// 性能优化测试

func TestConnectionPool(t *testing.T) {
	pool := NewConnectionPool(10)

	// 测试获取连接
	conn := pool.Get()
	if conn == nil {
		t.Error("Failed to get connection")
	}
	if !conn.InUse {
		t.Error("Connection should be in use")
	}

	// 测试释放连接
	pool.Release(conn)
	if conn.InUse {
		t.Error("Connection should not be in use after release")
	}
}

func TestConnectionPoolExhaustion(t *testing.T) {
	pool := NewConnectionPool(5)

	// 获取超过池大小的连接
	connections := make([]*PooledConnection, 10)
	for i := 0; i < 10; i++ {
		connections[i] = pool.Get()
	}

	// 所有连接都应该可用（超出部分创建新连接）
	for i, conn := range connections {
		if conn == nil {
			t.Errorf("Connection %d is nil", i)
		}
	}

	// 释放所有连接
	for _, conn := range connections {
		pool.Release(conn)
	}
}

func TestCacheManager(t *testing.T) {
	cache := NewCacheManager(100, 5*time.Minute)

	// 测试设置和获取
	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok {
		t.Error("Failed to get cached value")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// 测试缓存统计
	stats := cache.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
}

func TestCacheManagerTTL(t *testing.T) {
	cache := NewCacheManager(100, 100*time.Millisecond)

	cache.Set("key1", "value1")

	// 立即获取应该成功
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Error("Immediate get should succeed")
	}

	// 等待TTL过期
	time.Sleep(150 * time.Millisecond)

	// 过期后获取应该失败
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expired key should not be found")
	}

	// 应该记录为miss
	stats := cache.GetStats()
	if stats.Misses < 1 {
		t.Errorf("Expected at least 1 miss, got %d", stats.Misses)
	}
}

func TestBatchProcessor(t *testing.T) {
	processed := 0
	handler := func(batch []BatchItem) error {
		processed += len(batch)
		return nil
	}

	bp := NewBatchProcessor(5, 50*time.Millisecond, handler)

	// 添加少于batch size的项目
	for i := 0; i < 3; i++ {
		bp.Add(BatchItem{ID: string(i), Data: i})
	}

	// 等待超时触发处理
	time.Sleep(100 * time.Millisecond)

	if processed != 3 {
		t.Errorf("Expected 3 processed, got %d", processed)
	}
}

func TestBatchProcessorFullBatch(t *testing.T) {
	processed := 0
	handler := func(batch []BatchItem) error {
		processed += len(batch)
		return nil
	}

	bp := NewBatchProcessor(5, 1*time.Second, handler)

	// 添加等于batch size的项目
	for i := 0; i < 5; i++ {
		bp.Add(BatchItem{ID: string(i), Data: i})
	}

	// 应该立即处理，不需要等待
	time.Sleep(10 * time.Millisecond)

	if processed != 5 {
		t.Errorf("Expected 5 processed immediately, got %d", processed)
	}
}

func TestAsyncWorkerPool(t *testing.T) {
	pool := NewAsyncWorkerPool(5)

	completed := 0
	var mu sync.Mutex

	// 提交多个任务
	for i := 0; i < 20; i++ {
		pool.Submit(func() {
			mu.Lock()
			completed++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)
	pool.Stop()

	mu.Lock()
	if completed != 20 {
		t.Errorf("Expected 20 completed tasks, got %d", completed)
	}
	mu.Unlock()
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	// 记录延迟
	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	for _, lat := range latencies {
		collector.RecordLatency(lat)
	}

	// 记录吞吐量
	collector.RecordThroughput(100.0)
	collector.RecordThroughput(150.0)
	collector.RecordThroughput(200.0)

	// 记录错误
	collector.RecordError(2)
	collector.RecordError(3)

	// 获取报告
	report := collector.GetReport()

	if report.SampleCount != 5 {
		t.Errorf("Expected 5 samples, got %d", report.SampleCount)
	}

	expectedAvg := (10 + 20 + 30 + 40 + 50) / 5 * time.Millisecond
	if report.AvgLatency != expectedAvg {
		t.Errorf("Expected avg %v, got %v", expectedAvg, report.AvgLatency)
	}

	if report.MinLatency != 10*time.Millisecond {
		t.Errorf("Expected min 10ms, got %v", report.MinLatency)
	}

	if report.MaxLatency != 50*time.Millisecond {
		t.Errorf("Expected max 50ms, got %v", report.MaxLatency)
	}

	if report.ErrorCount != 5 {
		t.Errorf("Expected 5 errors, got %d", report.ErrorCount)
	}
}

func TestPerformanceOptimizer(t *testing.T) {
	config := DefaultOptimizerConfig()
	optimizer := NewPerformanceOptimizer(config)

	// 初始化
	err := optimizer.Initialize()
	if err != nil {
		t.Errorf("Failed to initialize optimizer: %v", err)
	}

	// 验证组件初始化
	if optimizer.poolManager == nil {
		t.Error("Connection pool should be initialized")
	}

	if optimizer.cacheManager == nil {
		t.Error("Cache manager should be initialized")
	}
}

func BenchmarkConnectionPoolGet(b *testing.B) {
	pool := NewConnectionPool(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn := pool.Get()
			pool.Release(conn)
		}
	})
}

func BenchmarkCacheManagerGet(b *testing.B) {
	cache := NewCacheManager(1000, 5*time.Minute)

	// 预填充缓存
	for i := 0; i < 100; i++ {
		cache.Set("key"+string(i), "value"+string(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get("key" + string(i%100))
			i++
		}
	})
}

func BenchmarkBatchProcessorAdd(b *testing.B) {
	handler := func(batch []BatchItem) error {
		return nil
	}

	bp := NewBatchProcessor(100, 10*time.Millisecond, handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp.Add(BatchItem{ID: string(i), Data: i})
	}
}

func BenchmarkAsyncWorkerPool(b *testing.B) {
	pool := NewAsyncWorkerPool(20)

	var wg sync.WaitGroup

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		pool.Submit(func() {
			time.Sleep(1 * time.Millisecond)
			wg.Done()
		})
	}

	wg.Wait()
	pool.Stop()
}