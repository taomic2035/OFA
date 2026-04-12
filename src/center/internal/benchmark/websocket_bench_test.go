package benchmark

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketBenchmark WebSocket 连接压测
func WebSocketBenchmark(addr string, config BenchmarkConfig) *BenchmarkResult {
	runner := NewBenchmarkRunner(config)

	ctx := context.Background()

	fn := func() (time.Duration, error) {
		start := time.Now()
		
		u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
		conn, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			return time.Since(start), err
		}
		defer conn.Close()

		// Send register message
		msg := map[string]interface{}{
			"type": "Register",
			"payload": map[string]interface{}{
				"agent_id": fmt.Sprintf("bench_agent_%d", time.Now().UnixNano()),
				"device_type": "benchmark",
			},
			"timestamp": time.Now().Unix(),
		}

		if err := conn.WriteJSON(msg); err != nil {
			return time.Since(start), err
		}

		// Read response
		_, _, err = conn.ReadMessage()
		if err != nil {
			return time.Since(start), err
		}

		return time.Since(start), nil
	}

	return runner.Run(ctx, fn)
}

// WebSocketMessageBenchmark WebSocket 消息发送压测
func WebSocketMessageBenchmark(addr string, config BenchmarkConfig) *BenchmarkResult {
	runner := NewBenchmarkRunner(config)
	ctx := context.Background()

	// Pre-connect
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	conn, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return &BenchmarkResult{Name: config.Name, FailureCount: int64(config.Concurrency)}
	}
	defer conn.Close()

	// Register
	registerMsg := map[string]interface{}{
		"type": "Register",
		"payload": map[string]interface{}{
			"agent_id": "bench_message_agent",
			"device_type": "benchmark",
		},
	}
	conn.WriteJSON(registerMsg)
	conn.ReadMessage()

	fn := func() (time.Duration, error) {
		start := time.Now()

		// Send heartbeat
		msg := map[string]interface{}{
			"type": "Heartbeat",
			"payload": map[string]interface{}{
				"agent_id": "bench_message_agent",
				"status": "online",
			},
		}

		if err := conn.WriteJSON(msg); err != nil {
			return time.Since(start), err
		}

		return time.Since(start), nil
	}

	return runner.Run(ctx, fn)
}

func TestWebSocketBenchmark(t *testing.T) {
	// Skip if no server available
	addr := "localhost:8080"

	config := BenchmarkConfig{
		Name:        "WebSocket_Connection_Test",
		Duration:    5 * time.Second,
		Concurrency: 5,
		WarmupTime:  2 * time.Second,
	}

	result := WebSocketBenchmark(addr, config)

	t.Logf("Total requests: %d", result.TotalRequests)
	t.Logf("Throughput: %.2f req/s", result.Throughput)
	t.Logf("Avg latency: %v", result.AvgLatency)

	// Assertions
	if result.TotalRequests < 10 {
		t.Skip("Not enough requests completed (server may not be available)")
	}

	if result.ErrorRate > 50 {
		t.Errorf("Error rate too high: %.2f%%", result.ErrorRate)
	}
}

func TestWebSocketMessageBenchmark(t *testing.T) {
	addr := "localhost:8080"

	config := BenchmarkConfig{
		Name:        "WebSocket_Message_Test",
		Duration:    5 * time.Second,
		Concurrency: 3,
		WarmupTime:  2 * time.Second,
	}

	result := WebSocketMessageBenchmark(addr, config)

	t.Logf("Total messages: %d", result.TotalRequests)
	t.Logf("Throughput: %.2f msg/s", result.Throughput)

	if result.AvgLatency > 100*time.Millisecond {
		t.Logf("Warning: Average latency > 100ms: %v", result.AvgLatency)
	}
}
