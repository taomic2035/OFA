package performance

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceConfig 性能测试配置
type PerformanceConfig struct {
	ConcurrentUsers   int           // 并发用户数
	RequestsPerUser   int           // 每用户请求数
	Duration          time.Duration // 测试持续时间
	WarmupDuration    time.Duration //预热时间
	ReportInterval    time.Duration // 报告间隔
	TargetLatencyMs   int64         // 目标延迟(ms)
	TargetThroughput  float64       // 目标吞吐量(req/s)
}

// DefaultPerformanceConfig 默认配置
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		ConcurrentUsers:   10,
		RequestsPerUser:   100,
		Duration:          30 * time.Second,
		WarmupDuration:    5 * time.Second,
		ReportInterval:    1 * time.Second,
		TargetLatencyMs:   100,
		TargetThroughput:  1000,
	}
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TotalRequests    int64         // 总请求数
	SuccessRequests  int64         // 成功请求数
	FailedRequests   int64         // 失败请求数
	TotalDuration    time.Duration // 总持续时间
	AvgLatency       time.Duration // 平均延迟
	MinLatency       time.Duration // 最小延迟
	MaxLatency       time.Duration // 最大延迟
	P50Latency       time.Duration // 50分位延迟
	P90Latency       time.Duration // 90分位延迟
	P99Latency       time.Duration // 99分位延迟
	Throughput       float64       // 吞吐量(req/s)
	ErrorRate        float64       // 错误率
	MemoryUsageMB    float64       // 内存使用(MB)
	CPUUsagePercent  float64       // CPU使用率(%)
	ConnectionCount  int           // 连接数
}

// LatencyRecorder 延迟记录器
type LatencyRecorder struct {
	latencies []time.Duration
	mu        sync.Mutex
}

// NewLatencyRecorder 创建延迟记录器
func NewLatencyRecorder() *LatencyRecorder {
	return &LatencyRecorder{
		latencies: make([]time.Duration, 0, 10000),
	}
}

// Record 记录延迟
func (lr *LatencyRecorder) Record(latency time.Duration) {
	lr.mu.Lock()
	lr.mu.Unlock()
	lr.latencies = append(lr.latencies, latency)
}

// CalculatePercentiles 计算百分位延迟
func (lr *LatencyRecorder) CalculatePercentiles() (p50, p90, p99 time.Duration) {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if len(lr.latencies) == 0 {
		return 0, 0, 0
	}

	// 排序
	sorted := make([]time.Duration, len(lr.latencies))
	copy(sorted, lr.latencies)
	sortLatencies(sorted)

	n := len(sorted)
	p50 = sorted[n/2]
	p90 = sorted[n*90/100]
	p99 = sorted[n*99/100]

	return p50, p90, p99
}

// sortLatencies 排序延迟数组
func sortLatencies(d []time.Duration) {
	// 简单快速排序
	for i := 0; i < len(d); i++ {
		for j := i + 1; j < len(d); j++ {
			if d[j] < d[i] {
				d[i], d[j] = d[j], d[i]
			}
		}
	}
}

// GetStats 获取统计信息
func (lr *LatencyRecorder) GetStats() (avg, min, max time.Duration) {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	if len(lr.latencies) == 0 {
		return 0, 0, 0
	}

	var sum time.Duration
	min = lr.latencies[0]
	max = lr.latencies[0]

	for _, l := range lr.latencies {
		sum += l
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}

	avg = sum / time.Duration(len(lr.latencies))
	return avg, min, max
}

// Clear 清空记录
func (lr *LatencyRecorder) Clear() {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	lr.latencies = make([]time.Duration, 0, 10000)
}

// PerformanceTest 性能测试框架
type PerformanceTest struct {
	config    *PerformanceConfig
	recorder  *LatencyRecorder
	results   *PerformanceMetrics
	stopChan  chan struct{}
	running   bool
	mu        sync.RWMutex
}

// NewPerformanceTest 创建性能测试
func NewPerformanceTest(config *PerformanceConfig) *PerformanceTest {
	return &PerformanceTest{
		config:   config,
		recorder: NewLatencyRecorder(),
		results:  &PerformanceMetrics{},
		stopChan: make(chan struct{}),
	}
}

// Run 运行性能测试
func (pt *PerformanceTest) Run(ctx context.Context, executor func() (time.Duration, error)) (*PerformanceMetrics, error) {
	pt.mu.Lock()
	pt.running = true
	pt.mu.Unlock()

	start := time.Now()
	var successCount, failedCount int64

	// 预热阶段
warmupCtx, warmupCancel := context.WithTimeout(ctx, pt.config.WarmupDuration)
	pt.runWarmup(warmupCtx, executor)
	warmupCancel()

	// 主测试阶段
	var wg sync.WaitGroup
	for i := 0; i < pt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < pt.config.RequestsPerUser; j++ {
				select {
				case <-pt.stopChan:
					return
				case <-ctx.Done():
					return
				default:
				}

				latency, err := executor()
				if err != nil {
					atomic.AddInt64(&failedCount, 1)
					continue
				}

				atomic.AddInt64(&successCount, 1)
				pt.recorder.Record(latency)
			}
		}()
	}

	// 等待完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		close(pt.stopChan)
		wg.Wait()
	case <-time.After(pt.config.Duration):
		close(pt.stopChan)
		wg.Wait()
	}

	// 计算结果
	pt.calculateResults(start, successCount, failedCount)

	pt.mu.Lock()
	pt.running = false
	pt.mu.Unlock()

	return pt.results, nil
}

// runWarmup 运行预热
func (pt *PerformanceTest) runWarmup(ctx context.Context, executor func() (time.Duration, error)) {
	for i := 0; i < pt.config.ConcurrentUsers/2; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				executor()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
}

// calculateResults 计算结果
func (pt *PerformanceTest) calculateResults(start time.Time, success, failed int64) {
	avg, min, max := pt.recorder.GetStats()
	p50, p90, p99 := pt.recorder.CalculatePercentiles()

	duration := time.Since(start)
	total := success + failed
	throughput := float64(success) / duration.Seconds()
	errorRate := float64(failed) / float64(total) * 100

	pt.results = &PerformanceMetrics{
		TotalRequests:   total,
		SuccessRequests: success,
		FailedRequests:  failed,
		TotalDuration:   duration,
		AvgLatency:      avg,
		MinLatency:      min,
		MaxLatency:      max,
		P50Latency:      p50,
		P90Latency:      p90,
		P99Latency:      p99,
		Throughput:      throughput,
		ErrorRate:       errorRate,
	}
}

// Stop 停止测试
func (pt *PerformanceTest) Stop() {
	pt.mu.RLock()
	if pt.running {
		close(pt.stopChan)
	}
	pt.mu.RUnlock()
}

// GetResults 获取结果
func (pt *PerformanceTest) GetResults() *PerformanceMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.results
}

// PrintReport 打印报告
func (pm *PerformanceMetrics) PrintReport() {
	fmt.Println("=== OFA 性能测试报告 ===")
	fmt.Println()
	fmt.Println("请求统计:")
	fmt.Printf("  总请求数:     %d\n", pm.TotalRequests)
	fmt.Printf("  成功请求数:   %d\n", pm.SuccessRequests)
	fmt.Printf("  失败请求数:   %d\n", pm.FailedRequests)
	fmt.Printf("  错误率:       %.2f%%\n", pm.ErrorRate)
	fmt.Println()
	fmt.Println("延迟统计:")
	fmt.Printf("  平均延迟:     %v\n", pm.AvgLatency)
	fmt.Printf("  最小延迟:     %v\n", pm.MinLatency)
	fmt.Printf("  最大延迟:     %v\n", pm.MaxLatency)
	fmt.Printf("  P50延迟:      %v\n", pm.P50Latency)
	fmt.Printf("  P90延迟:      %v\n", pm.P90Latency)
	fmt.Printf("  P99延迟:      %v\n", pm.P99Latency)
	fmt.Println()
	fmt.Println("吞吐量:")
	fmt.Printf("  吞吐量:       %.2f req/s\n", pm.Throughput)
	fmt.Printf("  总持续时间:   %v\n", pm.TotalDuration)
	fmt.Println()
	fmt.Println("资源使用:")
	fmt.Printf("  内存使用:     %.2f MB\n", pm.MemoryUsageMB)
	fmt.Printf("  CPU使用率:    %.2f%%\n", pm.CPUUsagePercent)
	fmt.Printf("  连接数:       %d\n", pm.ConnectionCount)
}

// CheckTargets 检查是否达到目标
func (pm *PerformanceMetrics) CheckTargets(targetLatencyMs int64, targetThroughput float64) []string {
	var issues []string

	avgMs := pm.AvgLatency.Milliseconds()
	if avgMs > targetLatencyMs {
		issues = append(issues, fmt.Sprintf("平均延迟 %dms 超过目标 %dms", avgMs, targetLatencyMs))
	}

	if pm.Throughput < targetThroughput {
		issues = append(issues, fmt.Sprintf("吞吐量 %.2f req/s 低于目标 %.2f req/s", pm.Throughput, targetThroughput))
	}

	if pm.ErrorRate > 1.0 {
		issues = append(issues, fmt.Sprintf("错误率 %.2f%% 过高", pm.ErrorRate))
	}

	return issues
}

// StressTest 压力测试
func StressTest(config *PerformanceConfig, executor func() (time.Duration, error)) (*PerformanceMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	pt := NewPerformanceTest(config)
	return pt.Run(ctx, executor)
}

// RampTest 渐进压力测试
func RampTest(baseConfig *PerformanceConfig, maxUsers int, step int, executor func() (time.Duration, error)) []*PerformanceMetrics {
	var results []*PerformanceMetrics

	for users := baseConfig.ConcurrentUsers; users <= maxUsers; users += step {
		config := *baseConfig
		config.ConcurrentUsers = users
		config.Duration = 10 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
		pt := NewPerformanceTest(&config)
		metrics, err := pt.Run(ctx, executor)
		cancel()

		if err != nil {
			fmt.Printf("用户数 %d 测试失败: %v\n", users, err)
			continue
		}

		results = append(results, metrics)
		fmt.Printf("用户数 %d: 吞吐量 %.2f req/s, 平均延迟 %v\n", users, metrics.Throughput, metrics.AvgLatency)
	}

	return results
}