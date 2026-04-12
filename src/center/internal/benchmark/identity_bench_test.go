package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// IdentityBenchmark 身份操作压测
func IdentityBenchmark(baseURL string, config BenchmarkConfig) *BenchmarkResult {
	runner := NewBenchmarkRunner(config)
	ctx := context.Background()

	fn := func() (time.Duration, error) {
		start := time.Now()

		// Create identity
		identityReq := map[string]interface{}{
			"name": fmt.Sprintf("BenchmarkUser_%d", time.Now().UnixNano()),
			"personality": map[string]interface{}{
				"openness": 0.5,
				"conscientiousness": 0.5,
				"extraversion": 0.5,
				"agreeableness": 0.5,
				"neuroticism": 0.5,
			},
		}

		body, _ := json.Marshal(identityReq)
		resp, err := http.Post(
			baseURL+"/api/v1/identity",
			"application/json",
			bytes.NewReader(body),
		)

		if err != nil {
			return time.Since(start), err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return time.Since(start), fmt.Errorf("status code: %d", resp.StatusCode)
		}

		// Parse response
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		// Get identity
		if id, ok := result["id"].(string); ok {
			getResp, err := http.Get(baseURL + "/api/v1/identity/" + id)
			if err != nil {
				return time.Since(start), err
			}
			getResp.Body.Close()
		}

		return time.Since(start), nil
	}

	return runner.Run(ctx, fn)
}

// IdentityGetBenchmark 身份获取压测
func IdentityGetBenchmark(baseURL string, identityID string, config BenchmarkConfig) *BenchmarkResult {
	runner := NewBenchmarkRunner(config)
	ctx := context.Background()

	fn := func() (time.Duration, error) {
		start := time.Now()

		resp, err := http.Get(baseURL + "/api/v1/identity/" + identityID)
		if err != nil {
			return time.Since(start), err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return time.Since(start), fmt.Errorf("status code: %d", resp.StatusCode)
		}

		// Read body
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		return time.Since(start), nil
	}

	return runner.Run(ctx, fn)
}

// BehaviorReportBenchmark 行为上报压测
func BehaviorReportBenchmark(baseURL string, config BenchmarkConfig) *BenchmarkResult {
	runner := NewBenchmarkRunner(config)
	ctx := context.Background()

	fn := func() (time.Duration, error) {
		start := time.Now()

		behaviorReq := map[string]interface{}{
			"type": "decision",
			"context": map[string]interface{}{
				"impulse_purchase": true,
				"timestamp": time.Now().Unix(),
			},
		}

		body, _ := json.Marshal(behaviorReq)
		resp, err := http.Post(
			baseURL+"/api/v1/behavior",
			"application/json",
			bytes.NewReader(body),
		)

		if err != nil {
			return time.Since(start), err
		}
		defer resp.Body.Close()

		return time.Since(start), nil
	}

	return runner.Run(ctx, fn)
}

func TestIdentityBenchmark(t *testing.T) {
	baseURL := "http://localhost:8080"

	config := BenchmarkConfig{
		Name:        "Identity_Operations_Test",
		Duration:    5 * time.Second,
		Concurrency: 10,
		WarmupTime:  2 * time.Second,
	}

	result := IdentityBenchmark(baseURL, config)

	t.Logf("Total operations: %d", result.TotalRequests)
	t.Logf("Throughput: %.2f ops/s", result.Throughput)
	t.Logf("P99 latency: %v", result.P99Latency)

	if result.TotalRequests < 10 {
		t.Skip("Not enough operations completed")
	}
}

func TestBehaviorReportBenchmark(t *testing.T) {
	baseURL := "http://localhost:8080"

	config := BenchmarkConfig{
		Name:        "Behavior_Report_Test",
		Duration:    5 * time.Second,
		Concurrency: 20,
		WarmupTime:  2 * time.Second,
	}

	result := BehaviorReportBenchmark(baseURL, config)

	t.Logf("Total reports: %d", result.TotalRequests)
	t.Logf("Throughput: %.2f reports/s", result.Throughput)
}

// bytes.NewReader import
import "bytes"
