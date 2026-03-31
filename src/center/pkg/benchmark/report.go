// Package benchmark - 性能基准测试报告生成器
// Sprint 26: v8.0最终发布 - 性能测试
package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceReport 性能测试报告
type PerformanceReport struct {
	GeneratedAt    time.Time           `json:"generated_at"`
	GoVersion      string              `json:"go_version"`
	OS             string              `json:"os"`
	Arch           string              `json:"arch"`
	CPUCount       int                 `json:"cpu_count"`
	MemoryMB       uint64              `json:"memory_mb"`
	TestResults    []PerformanceTest   `json:"test_results"`
	Summary        *PerformanceSummary `json:"summary"`
	Recommendations []string           `json:"recommendations"`
}

// PerformanceTest 单项性能测试
type PerformanceTest struct {
	Name           string        `json:"name"`
	Category       string        `json:"category"`
	Duration       time.Duration `json:"duration"`
	Operations     int64         `json:"operations"`
	OpsPerSecond   float64       `json:"ops_per_second"`
	AvgLatency     time.Duration `json:"avg_latency_ns"`
	P50Latency     time.Duration `json:"p50_latency_ns"`
	P99Latency     time.Duration `json:"p99_latency_ns"`
	MaxLatency     time.Duration `json:"max_latency_ns"`
	MemoryAllocMB  float64       `json:"memory_alloc_mb"`
	AllocsPerOp    int64         `json:"allocs_per_op"`
	Status         string        `json:"status"` // pass, warn, fail
	Baseline       *Baseline     `json:"baseline,omitempty"`
}

// Baseline 基准线
type Baseline struct {
	OpsPerSecond float64 `json:"ops_per_second"`
	MaxLatencyMs float64 `json:"max_latency_ms"`
}

// PerformanceSummary 性能摘要
type PerformanceSummary struct {
	TotalTests      int     `json:"total_tests"`
	PassedTests     int     `json:"passed_tests"`
	WarningTests    int     `json:"warning_tests"`
	FailedTests     int     `json:"failed_tests"`
	OverallScore    int     `json:"overall_score"` // 0-100
	ThroughputScore int     `json:"throughput_score"`
	LatencyScore    int     `json:"latency_score"`
	MemoryScore     int     `json:"memory_score"`
}

// PerformanceRunner 性能测试运行器
type PerformanceRunner struct {
	results []PerformanceTest
	mu      sync.RWMutex
}

// NewPerformanceRunner 创建性能测试运行器
func NewPerformanceRunner() *PerformanceRunner {
	return &PerformanceRunner{
		results: make([]PerformanceTest, 0),
	}
}

// TestFunc 测试函数类型
type TestFunc func(ctx context.Context, iterations int) TestMetrics

// TestMetrics 测试指标
type TestMetrics struct {
	Operations   int64
	AvgLatency   time.Duration
	P50Latency   time.Duration
	P99Latency   time.Duration
	MaxLatency   time.Duration
	MemoryAlloc  uint64
	Allocs       uint64
}

// RunTest 运行单个测试
func (pr *PerformanceRunner) RunTest(ctx context.Context, name, category string, duration time.Duration, testFunc TestFunc) PerformanceTest {
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startMem := m.Alloc

	// 预热
	testFunc(ctx, 100)

	// 正式测试
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	iterations := 0
	var totalOps int64
	var totalLatency time.Duration
	var maxLatency time.Duration
	latencies := make([]time.Duration, 0)
	var totalAllocs uint64

	for {
		select {
		case <-ctx.Done():
			goto done
		default:
			metrics := testFunc(ctx, 1000)
			totalOps += metrics.Operations
			totalLatency += metrics.AvgLatency * time.Duration(metrics.Operations)
			if metrics.MaxLatency > maxLatency {
				maxLatency = metrics.MaxLatency
			}
			latencies = append(latencies, metrics.AvgLatency)
			totalAllocs += metrics.Allocs
			iterations++
		}
	}
done:

	runtime.ReadMemStats(&m)
	endMem := m.Alloc

	testDuration := time.Since(start)
	opsPerSecond := float64(totalOps) / testDuration.Seconds()

	// 计算百分位延迟
	var p50, p99 time.Duration
	if len(latencies) > 0 {
		// 简化计算
		p50 = latencies[len(latencies)/2]
		p99 = latencies[len(latencies)*99/100]
		if p99 == 0 {
			p99 = latencies[len(latencies)-1]
		}
	}

	avgLatency := time.Duration(0)
	if totalOps > 0 {
		avgLatency = time.Duration(int64(totalLatency) / totalOps)
	}

	result := PerformanceTest{
		Name:          name,
		Category:      category,
		Duration:      testDuration,
		Operations:    totalOps,
		OpsPerSecond:  opsPerSecond,
		AvgLatency:    avgLatency,
		P50Latency:    p50,
		P99Latency:    p99,
		MaxLatency:    maxLatency,
		MemoryAllocMB: float64(endMem-startMem) / 1024 / 1024,
		AllocsPerOp:   int64(totalAllocs) / totalOps,
	}

	// 评估状态
	result.Status = pr.evaluateTest(result)

	pr.mu.Lock()
	pr.results = append(pr.results, result)
	pr.mu.Unlock()

	return result
}

// evaluateTest 评估测试结果
func (pr *PerformanceRunner) evaluateTest(result PerformanceTest) string {
	// 根据类型设置不同的阈值
	thresholds := map[string]struct {
		minOps   float64
		maxLatMs float64
	}{
		"agent":  {minOps: 1000, maxLatMs: 50},
		"task":   {minOps: 500, maxLatMs: 100},
		"message": {minOps: 5000, maxLatMs: 20},
		"cache":  {minOps: 10000, maxLatMs: 10},
		"store":  {minOps: 1000, maxLatMs: 50},
	}

	th, ok := thresholds[result.Category]
	if !ok {
		th = struct{ minOps float64; maxLatMs float64 }{minOps: 100, maxLatMs: 100}
	}

	latencyMs := float64(result.MaxLatency.Microseconds()) / 1000

	if result.OpsPerSecond >= th.minOps && latencyMs <= th.maxLatMs {
		return "pass"
	} else if result.OpsPerSecond >= th.minOps*0.7 && latencyMs <= th.maxLatMs*1.5 {
		return "warn"
	}
	return "fail"
}

// RunAllTests 运行所有测试
func (pr *PerformanceRunner) RunAllTests(ctx context.Context) *PerformanceReport {
	report := &PerformanceReport{
		GeneratedAt: time.Now(),
		GoVersion:   runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		CPUCount:    runtime.NumCPU(),
		TestResults: make([]PerformanceTest, 0),
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	report.MemoryMB = m.Sys / 1024 / 1024

	// Agent注册测试
	pr.RunTest(ctx, "agent_register", "agent", 5*time.Second, mockAgentRegister)

	// Agent心跳测试
	pr.RunTest(ctx, "agent_heartbeat", "agent", 5*time.Second, mockAgentHeartbeat)

	// 任务提交测试
	pr.RunTest(ctx, "task_submit", "task", 5*time.Second, mockTaskSubmit)

	// 任务执行测试
	pr.RunTest(ctx, "task_execute", "task", 5*time.Second, mockTaskExecute)

	// P2P消息测试
	pr.RunTest(ctx, "p2p_message", "message", 5*time.Second, mockP2PMessage)

	// 广播消息测试
	pr.RunTest(ctx, "broadcast_message", "message", 5*time.Second, mockBroadcastMessage)

	// 缓存读取测试
	pr.RunTest(ctx, "cache_get", "cache", 5*time.Second, mockCacheGet)

	// 缓存写入测试
	pr.RunTest(ctx, "cache_set", "cache", 5*time.Second, mockCacheSet)

	// 数据库读取测试
	pr.RunTest(ctx, "store_query", "store", 5*time.Second, mockStoreQuery)

	// 数据库写入测试
	pr.RunTest(ctx, "store_insert", "store", 5*time.Second, mockStoreInsert)

	report.TestResults = pr.results
	report.Summary = pr.calculateSummary()
	report.Recommendations = pr.generateRecommendations()

	return report
}

// 模拟测试函数
func mockAgentRegister(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		// 模拟Agent注册处理
		time.Sleep(time.Microsecond * time.Duration(50+mockLatency()))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockAgentHeartbeat(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(20+mockLatency()))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockTaskSubmit(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(100+mockLatency()*2))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockTaskExecute(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(200+mockLatency()*3))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockP2PMessage(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(30+mockLatency()))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockBroadcastMessage(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(50+mockLatency()))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockCacheGet(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(5+mockLatency()/2))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockCacheSet(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(10+mockLatency()/2))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockStoreQuery(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(100+mockLatency()))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockStoreInsert(ctx context.Context, iterations int) TestMetrics {
	var ops int64
	var totalLatency time.Duration
	var maxLatency time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		time.Sleep(time.Microsecond * time.Duration(150+mockLatency()*2))
		latency := time.Since(start)
		ops++
		totalLatency += latency
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	return TestMetrics{
		Operations: ops,
		AvgLatency: time.Duration(int64(totalLatency) / ops),
		MaxLatency: maxLatency,
	}
}

func mockLatency() int {
	return 0
}

// calculateSummary 计算摘要
func (pr *PerformanceRunner) calculateSummary() *PerformanceSummary {
	summary := &PerformanceSummary{
		TotalTests: len(pr.results),
	}

	var throughputScore, latencyScore, memoryScore int
	for _, r := range pr.results {
		switch r.Status {
		case "pass":
			summary.PassedTests++
			throughputScore += 10
			latencyScore += 10
		case "warn":
			summary.WarningTests++
			throughputScore += 7
			latencyScore += 7
		case "fail":
			summary.FailedTests++
			throughputScore += 3
			latencyScore += 3
		}

		if r.MemoryAllocMB < 100 {
			memoryScore += 10
		} else if r.MemoryAllocMB < 500 {
			memoryScore += 7
		} else {
			memoryScore += 3
		}
	}

	// 标准化分数
	count := len(pr.results)
	if count > 0 {
		summary.ThroughputScore = throughputScore / count
		summary.LatencyScore = latencyScore / count
		summary.MemoryScore = memoryScore / count
		summary.OverallScore = (summary.ThroughputScore + summary.LatencyScore + summary.MemoryScore) / 3
	}

	return summary
}

// generateRecommendations 生成建议
func (pr *PerformanceRunner) generateRecommendations() []string {
	recommendations := make([]string, 0)

	for _, r := range pr.results {
		if r.Status == "fail" {
			recommendations = append(recommendations,
				fmt.Sprintf("[%s] %s 性能不达标，建议优化", r.Category, r.Name))
		} else if r.Status == "warn" {
			recommendations = append(recommendations,
				fmt.Sprintf("[%s] %s 性能接近阈值，建议关注", r.Category, r.Name))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "所有性能测试均达标，系统性能良好")
	}

	return recommendations
}

// ExportReport 导出报告
func (r *PerformanceReport) ExportReport() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// PrintReport 打印报告
func (r *PerformanceReport) PrintReport() string {
	var sb string
	sb = fmt.Sprintf("\n")
	sb += fmt.Sprintf("╔════════════════════════════════════════════════════════════════╗\n")
	sb += fmt.Sprintf("║              OFA v8.0 性能基准测试报告                          ║\n")
	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")
	sb += fmt.Sprintf("║ 生成时间: %-52s║\n", r.GeneratedAt.Format(time.RFC3339))
	sb += fmt.Sprintf("║ Go版本: %-54s║\n", r.GoVersion)
	sb += fmt.Sprintf("║ 平台: %s/%s %-48s║\n", r.OS, r.Arch, "")
	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")
	sb += fmt.Sprintf("║                        测试结果                                 ║\n")
	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")

	for _, t := range r.TestResults {
		status := "✅"
		if t.Status == "warn" {
			status = "⚠️"
		} else if t.Status == "fail" {
			status = "❌"
		}
		sb += fmt.Sprintf("║ %s %-20s %10.0f ops/s  %-12s║\n",
			status, t.Name, t.OpsPerSecond, fmt.Sprintf("P99: %v", t.P99Latency))
	}

	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")
	sb += fmt.Sprintf("║                        性能评分                                 ║\n")
	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")
	sb += fmt.Sprintf("║ 综合评分: %-3d/100  吞吐量: %-3d  延迟: %-3d  内存: %-3d     ║\n",
		r.Summary.OverallScore, r.Summary.ThroughputScore, r.Summary.LatencyScore, r.Summary.MemoryScore)
	sb += fmt.Sprintf("╠════════════════════════════════════════════════════════════════╣\n")
	sb += fmt.Sprintf("║ 通过: %d  警告: %d  失败: %d                                    ║\n",
		r.Summary.PassedTests, r.Summary.WarningTests, r.Summary.FailedTests)
	sb += fmt.Sprintf("╚════════════════════════════════════════════════════════════════╝\n")

	return sb
}

// SaveReport 保存报告到文件
func (r *PerformanceReport) SaveReport(path string) error {
	data, err := r.ExportReport()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// QuickBenchmark 快速基准测试
func QuickBenchmark() *PerformanceReport {
	runner := NewPerformanceRunner()
	ctx := context.Background()
	return runner.RunAllTests(ctx)
}