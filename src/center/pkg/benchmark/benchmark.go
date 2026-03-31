package benchmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/ofa/center/proto"
)

// BenchmarkConfig holds benchmark configuration
type BenchmarkConfig struct {
	CenterURL      string
	Concurrency    int
	TotalTasks     int
	TaskTimeout    time.Duration
	ReportInterval time.Duration
}

// DefaultBenchmarkConfig returns default configuration
func DefaultBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		CenterURL:      "http://localhost:8080",
		Concurrency:    10,
		TotalTasks:     1000,
		TaskTimeout:    30 * time.Second,
		ReportInterval: time.Second,
	}
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	TotalTasks      int64
	SuccessTasks    int64
	FailedTasks     int64
	TotalDuration   time.Duration
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	Throughput      float64 // tasks per second
	Errors          []error
	LatencyBuckets  map[string]int64 // latency distribution
}

// Benchmark runs performance tests
type Benchmark struct {
	config  *BenchmarkConfig
	client  *http.Client
	results *BenchmarkResult
}

// NewBenchmark creates a new benchmark runner
func NewBenchmark(config *BenchmarkConfig) *Benchmark {
	return &Benchmark{
		config: config,
		client: &http.Client{
			Timeout: config.TaskTimeout,
		},
		results: &BenchmarkResult{
			LatencyBuckets: make(map[string]int64),
		},
	}
}

// Run executes the benchmark
func (b *Benchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	start := time.Now()

	var wg sync.WaitGroup
	taskChan := make(chan int, b.config.TotalTasks)

	// Generate tasks
	for i := 0; i < b.config.TotalTasks; i++ {
		taskChan <- i
	}
	close(taskChan)

	// Track latencies
	var latencies sync.Map
	var successCount, failedCount int64

	// Start workers
	for i := 0; i < b.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range taskChan {
				latency, err := b.submitTask(ctx)
				if err != nil {
					atomic.AddInt64(&failedCount, 1)
					b.results.Errors = append(b.results.Errors, err)
					continue
				}

				atomic.AddInt64(&successCount, 1)
				latencyMs := latency.Milliseconds()
				bucket := b.getLatencyBucket(latencyMs)
				latencies.Store(bucket, atomic.AddInt64(&b.results.LatencyBuckets[bucket], 1))
			}
		}()
	}

	// Wait for completion
	wg.Wait()

	// Calculate results
	b.results.TotalTasks = int64(b.config.TotalTasks)
	b.results.SuccessTasks = successCount
	b.results.FailedTasks = failedCount
	b.results.TotalDuration = time.Since(start)
	b.results.Throughput = float64(successCount) / b.results.TotalDuration.Seconds()

	return b.results, nil
}

func (b *Benchmark) submitTask(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	// Submit task request
	req := map[string]interface{}{
		"skill_id": "text.process",
		"input":    `{"text":"hello","operation":"uppercase"}`,
	}

	reqBody, _ := json.Marshal(req)
	url := fmt.Sprintf("%s/api/v1/tasks", b.config.CenterURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	// Poll for completion
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}

		taskURL := fmt.Sprintf("%s/api/v1/tasks/%s", b.config.CenterURL, result.TaskID)
		taskReq, _ := http.NewRequestWithContext(ctx, "GET", taskURL, nil)

		taskResp, err := b.client.Do(taskReq)
		if err != nil {
			return 0, err
		}

		var task struct {
			Status int `json:"status"`
		}
		json.NewDecoder(taskResp.Body).Decode(&task)
		taskResp.Body.Close()

		// Task completed or failed
		if task.Status == int(pb.TaskStatus_TASK_STATUS_COMPLETED) ||
			task.Status == int(pb.TaskStatus_TASK_STATUS_FAILED) {
			return time.Since(start), nil
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Benchmark) getLatencyBucket(ms int64) string {
	switch {
	case ms < 10:
		return "0-10ms"
	case ms < 50:
		return "10-50ms"
	case ms < 100:
		return "50-100ms"
	case ms < 500:
		return "100-500ms"
	case ms < 1000:
		return "500ms-1s"
	default:
		return ">1s"
	}
}

// PrintResults prints benchmark results
func (r *BenchmarkResult) PrintResults() {
	fmt.Println("=== OFA Benchmark Results ===")
	fmt.Printf("Total Tasks:     %d\n", r.TotalTasks)
	fmt.Printf("Successful:      %d\n", r.SuccessTasks)
	fmt.Printf("Failed:          %d\n", r.FailedTasks)
	fmt.Printf("Total Duration:  %v\n", r.TotalDuration)
	fmt.Printf("Throughput:      %.2f tasks/sec\n", r.Throughput)
	fmt.Println()
	fmt.Println("Latency Distribution:")
	for bucket, count := range r.LatencyBuckets {
		fmt.Printf("  %s: %d\n", bucket, count)
	}
}

// LoadTest runs a simple load test
func LoadTest(url string, requests int, concurrency int) error {
	var success, failed int64
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for j := 0; j < requests/concurrency; j++ {
				resp, err := client.Get(url + "/health")
				if err != nil || resp.StatusCode != 200 {
					atomic.AddInt64(&failed, 1)
					if resp != nil {
						resp.Body.Close()
					}
					continue
				}
				resp.Body.Close()
				atomic.AddInt64(&success, 1)
			}
		}()
	}

	wg.Wait()

	duration := time.Since(start)
	total := success + failed
	throughput := float64(total) / duration.Seconds()

	log.Printf("Load Test Results:")
	log.Printf("  Total: %d, Success: %d, Failed: %d", total, success, failed)
	log.Printf("  Duration: %v", duration)
	log.Printf("  Throughput: %.2f req/sec", throughput)

	return nil
}