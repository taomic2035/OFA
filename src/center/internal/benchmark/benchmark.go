// benchmark.go
// OFA Center Performance Benchmark Framework (v9.0.0)

package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// BenchmarkConfig 压测配置
type BenchmarkConfig struct {
	Name          string
	Duration      time.Duration
	Concurrency   int
	TargetRate    int // requests per second
	WarmupTime    time.Duration
	SampleInterval time.Duration
}

// BenchmarkResult 压测结果
type BenchmarkResult struct {
	Name           string
	Duration       time.Duration
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	AvgLatency     time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	P50Latency     time.Duration
	P90Latency     time.Duration
	P99Latency     time.Duration
	Throughput     float64 // requests per second
	ErrorRate      float64
	StartTime      time.Time
	EndTime        time.Time
}

// LatencySample 延迟样本
type LatencySample struct {
	Latency   time.Duration
	Success   bool
	Timestamp time.Time
}

// BenchmarkRunner 压测运行器
type BenchmarkRunner struct {
	config   BenchmarkConfig
	samples  []LatencySample
	mu       sync.Mutex
	start    time.Time
	wg       sync.WaitGroup
	stopCh   chan struct{}
}

// NewBenchmarkRunner 创建压测运行器
func NewBenchmarkRunner(config BenchmarkConfig) *BenchmarkRunner {
	return &BenchmarkRunner{
		config:  config,
		samples: make([]LatencySample, 0),
		stopCh:  make(chan struct{}),
	}
}

// Run 执行压测
func (r *BenchmarkRunner) Run(ctx context.Context, fn func() (time.Duration, error)) *BenchmarkResult {
	r.start = time.Now()
	r.samples = make([]LatencySample, 0)

	// Warmup phase
	if r.config.WarmupTime > 0 {
		fmt.Printf("[%s] Warmup phase (%v)\n", r.config.Name, r.config.WarmupTime)
		r.runWarmup(ctx, fn)
	}

	// Main benchmark phase
	fmt.Printf("[%s] Starting benchmark (%v, concurrency=%d)\n", 
		r.config.Name, r.config.Duration, r.config.Concurrency)

	// Start workers
	for i := 0; i < r.config.Concurrency; i++ {
		r.wg.Add(1)
		go r.worker(ctx, fn, i)
	}

	// Wait for duration
	select {
	case <-time.After(r.config.Duration):
	case <-ctx.Done():
	}

	// Stop workers
	close(r.stopCh)
	r.wg.Wait()

	// Calculate results
	return r.calculateResult()
}

func (r *BenchmarkRunner) runWarmup(ctx context.Context, fn func() (time.Duration, error)) {
	warmupCtx, cancel := context.WithTimeout(ctx, r.config.WarmupTime)
	defer cancel()

	for {
		select {
		case <-warmupCtx.Done():
			return
		default:
			latency, err := fn()
			if err == nil {
				// Don't record warmup samples
			}
		}
	}
}

func (r *BenchmarkRunner) worker(ctx context.Context, fn func() (time.Duration, error), id int) {
	defer r.wg.Done()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			latency, err := fn()
			sample := LatencySample{
				Latency:   latency,
				Success:   err == nil,
				Timestamp: time.Now(),
			}
			r.recordSample(sample)
		}
	}
}

func (r *BenchmarkRunner) recordSample(sample LatencySample) {
	r.mu.Lock()
	r.samples = append(r.samples, sample)
	r.mu.Unlock()
}

func (r *BenchmarkRunner) calculateResult() *BenchmarkResult {
	if len(r.samples) == 0 {
		return &BenchmarkResult{Name: r.config.Name}
	}

	// Sort samples by latency
	sorted := make([]time.Duration, len(r.samples))
	successCount := 0
	minLatency := time.Duration(1<<63 - 1)
	maxLatency := time.Duration(0)
	totalLatency := time.Duration(0)

	for i, s := range r.samples {
		if s.Success {
			successCount++
			totalLatency += s.Latency
			if s.Latency < minLatency {
				minLatency = s.Latency
			}
			if s.Latency > maxLatency {
				maxLatency = s.Latency
			}
		}
		sorted[i] = s.Latency
	}

	// Sort for percentile calculation
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	duration := time.Since(r.start)
	totalRequests := len(r.samples)
	failureCount := totalRequests - successCount

	// Calculate percentiles
	p50 := sorted[len(sorted)*50/100]
	p90 := sorted[len(sorted)*90/100]
	p99 := sorted[len(sorted)*99/100]

	avgLatency := time.Duration(0)
	if successCount > 0 {
		avgLatency = totalLatency / time.Duration(successCount)
	}

	throughput := float64(totalRequests) / duration.Seconds()
	errorRate := float64(failureCount) / float64(totalRequests) * 100

	return &BenchmarkResult{
		Name:          r.config.Name,
		Duration:      duration,
		TotalRequests: int64(totalRequests),
		SuccessCount:  int64(successCount),
		FailureCount:  int64(failureCount),
		AvgLatency:    avgLatency,
		MinLatency:    minLatency,
		MaxLatency:    maxLatency,
		P50Latency:    p50,
		P90Latency:    p90,
		P99Latency:    p99,
		Throughput:    throughput,
		ErrorRate:     errorRate,
		StartTime:     r.start,
		EndTime:       time.Now(),
	}
}

// PrintResult 打印压测结果
func (r *BenchmarkResult) Print() {
	fmt.Println("\n========================================")
	fmt.Printf("Benchmark Results: %s\n", r.Name)
	fmt.Println("========================================")
	fmt.Printf("Duration:       %v\n", r.Duration)
	fmt.Printf("Total Requests: %d\n", r.TotalRequests)
	fmt.Printf("Success:        %d\n", r.SuccessCount)
	fmt.Printf("Failures:       %d\n", r.FailureCount)
	fmt.Printf("Error Rate:     %.2f%%\n", r.ErrorRate)
	fmt.Printf("Throughput:     %.2f req/s\n", r.Throughput)
	fmt.Println("\nLatency Distribution:")
	fmt.Printf("  Average:      %v\n", r.AvgLatency)
	fmt.Printf("  Min:          %v\n", r.MinLatency)
	fmt.Printf("  Max:          %v\n", r.MaxLatency)
	fmt.Printf("  P50:          %v\n", r.P50Latency)
	fmt.Printf("  P90:          %v\n", r.P90Latency)
	fmt.Printf("  P99:          %v\n", r.P99Latency)
	fmt.Println("========================================")
}

// JSON 返回 JSON 格式结果
func (r *BenchmarkResult) JSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// BenchmarkSuite 压测套件
type BenchmarkSuite struct {
	configs  []BenchmarkConfig
	results  []*BenchmarkResult
}

// NewBenchmarkSuite 创建压测套件
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		configs: make([]BenchmarkConfig, 0),
		results: make([]*BenchmarkResult, 0),
	}
}

// AddBenchmark 添加压测配置
func (s *BenchmarkSuite) AddBenchmark(config BenchmarkConfig) {
	s.configs = append(s.configs, config)
}

// RunAll 执行所有压测
func (s *BenchmarkSuite) RunAll(ctx context.Context, fn func(name string) func() (time.Duration, error)) {
	for _, config := range s.configs {
		runner := NewBenchmarkRunner(config)
		result := runner.Run(ctx, fn(config.Name))
		s.results = append(s.results, result)
		result.Print()
	}
}

// Summary 返回汇总结果
func (s *BenchmarkSuite) Summary() string {
	summary := "\n========================================\n"
	summary += "Benchmark Suite Summary\n"
	summary += "========================================\n"

	for _, r := range s.results {
		summary += fmt.Sprintf("%s: %.2f req/s, P99=%v, Error=%.2f%%\n",
			r.Name, r.Throughput, r.P99Latency, r.ErrorRate)
	}

	return summary
}

// DefaultBenchmarkConfigs 默认压测配置
func DefaultBenchmarkConfigs() []BenchmarkConfig {
	return []BenchmarkConfig{
		{
			Name:         "WebSocket_Connection",
			Duration:     30 * time.Second,
			Concurrency:  10,
			WarmupTime:   5 * time.Second,
		},
		{
			Name:         "Identity_Operations",
			Duration:     30 * time.Second,
			Concurrency:  20,
			WarmupTime:   5 * time.Second,
		},
		{
			Name:         "Chat_Request",
			Duration:     60 * time.Second,
			Concurrency:  5,
			WarmupTime:   10 * time.Second,
		},
		{
			Name:         "Scene_Detection",
			Duration:     30 * time.Second,
			Concurrency:  15,
			WarmupTime:   5 * time.Second,
		},
	}
}
