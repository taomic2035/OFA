// metrics.go
// OFA Center Benchmark Metrics Collection (v9.0.0)

package benchmark

import (
	"sync"
	"time"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	requestCounts   map[string]int64
	latencySum      map[string]time.Duration
	latencyMin      map[string]time.Duration
	latencyMax      map[string]time.Duration
	errorCounts     map[string]int64
	mu              sync.Mutex
	startTime       time.Time
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestCounts: make(map[string]int64),
		latencySum:    make(map[string]time.Duration),
		latencyMin:    make(map[string]time.Duration),
		latencyMax:    make(map[string]time.Duration),
		errorCounts:   make(map[string]int64),
		startTime:     time.Now(),
	}
}

// RecordRequest 记录请求
func (m *MetricsCollector) RecordRequest(operation string, latency time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestCounts[operation]++

	if success {
		m.latencySum[operation] += latency

		// Update min/max
		currentMin := m.latencyMin[operation]
		if currentMin == 0 || latency < currentMin {
			m.latencyMin[operation] = latency
		}

		currentMax := m.latencyMax[operation]
		if latency > currentMax {
			m.latencyMax[operation] = latency
		}
	} else {
		m.errorCounts[operation]++
	}
}

// GetMetrics 获取指标
func (m *MetricsCollector) GetMetrics() map[string]OperationMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]OperationMetrics)
	duration := time.Since(m.startTime)

	for op, count := range m.requestCounts {
		avg := time.Duration(0)
		if count > 0 {
			avg = m.latencySum[op] / time.Duration(count)
		}

		result[op] = OperationMetrics{
			Operation:     op,
			RequestCount:  count,
			AvgLatency:    avg,
			MinLatency:    m.latencyMin[op],
			MaxLatency:    m.latencyMax[op],
			ErrorCount:    m.errorCounts[op],
			Throughput:    float64(count) / duration.Seconds(),
			ErrorRate:     float64(m.errorCounts[op]) / float64(count) * 100,
		}
	}

	return result
}

// OperationMetrics 操作指标
type OperationMetrics struct {
	Operation     string
	RequestCount  int64
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	ErrorCount    int64
	Throughput    float64
	ErrorRate     float64
}

// PerformanceThresholds 性能阈值
type PerformanceThresholds struct {
	MaxAvgLatency   time.Duration // 最大平均延迟
	MaxP99Latency   time.Duration // 最大 P99 延迟
	MinThroughput   float64       // 最小吞吐量
	MaxErrorRate    float64       // 最大错误率
}

// DefaultThresholds 默认性能阈值
func DefaultThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MaxAvgLatency: 100 * time.Millisecond,
		MaxP99Latency: 500 * time.Millisecond,
		MinThroughput: 100, // 100 req/s
		MaxErrorRate:  1.0, // 1%
	}
}

// CheckThresholds 检查性能阈值
func (r *BenchmarkResult) CheckThresholds(thresholds PerformanceThresholds) []string {
	warnings := make([]string, 0)

	if r.AvgLatency > thresholds.MaxAvgLatency {
		warnings = append(warnings, 
			"Average latency exceeds threshold: %v > %v", 
			r.AvgLatency, thresholds.MaxAvgLatency)
	}

	if r.P99Latency > thresholds.MaxP99Latency {
		warnings = append(warnings, 
			"P99 latency exceeds threshold: %v > %v",
			r.P99Latency, thresholds.MaxP99Latency)
	}

	if r.Throughput < thresholds.MinThroughput {
		warnings = append(warnings,
			"Throughput below threshold: %.2f < %.2f",
			r.Throughput, thresholds.MinThroughput)
	}

	if r.ErrorRate > thresholds.MaxErrorRate {
		warnings = append(warnings,
			"Error rate exceeds threshold: %.2f%% > %.2f%%",
			r.ErrorRate, thresholds.MaxErrorRate)
	}

	return warnings
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	Timestamp     time.Time
	Duration      time.Duration
	Benchmarks    []*BenchmarkResult
	Metrics       map[string]OperationMetrics
	Warnings      []string
	Passed        bool
}

// GenerateReport 生成性能报告
func GenerateReport(results []*BenchmarkResult) *PerformanceReport {
	thresholds := DefaultThresholds()
	allWarnings := make([]string, 0)
	allPassed := true

	for _, r := range results {
		warnings := r.CheckThresholds(thresholds)
		if len(warnings) > 0 {
			allPassed = false
			allWarnings = append(allWarnings, r.Name+": "+warnings...)
		}
	}

	return &PerformanceReport{
		Timestamp:  time.Now(),
		Duration:   time.Since(results[0].StartTime),
		Benchmarks: results,
		Warnings:   allWarnings,
		Passed:     allPassed,
	}
}

// Print 打印报告
func (r *PerformanceReport) Print() {
	println("\n========================================")
	println("Performance Report")
	println("========================================")
	println("Generated:", r.Timestamp.Format(time.RFC3339))

	for _, bench := range r.Benchmarks {
		bench.Print()
	}

	if len(r.Warnings) > 0 {
		println("\n⚠️ Warnings:")
		for _, w := range r.Warnings {
			println("  -", w)
		}
	}

	if r.Passed {
		println("\n✅ All performance thresholds passed")
	} else {
		println("\n❌ Some performance thresholds failed")
	}

	println("========================================")
}
