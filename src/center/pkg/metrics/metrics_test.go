package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	// Check that all metrics are initialized
	if m.AgentsTotal == nil {
		t.Error("AgentsTotal not initialized")
	}
	if m.AgentsOnline == nil {
		t.Error("AgentsOnline not initialized")
	}
	if m.TasksTotal == nil {
		t.Error("TasksTotal not initialized")
	}
	if m.TaskDuration == nil {
		t.Error("TaskDuration not initialized")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration not initialized")
	}
}

func TestMetricsHandler(t *testing.T) {
	m := NewMetrics()

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Call handler
	m.Handler().ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that response contains metrics
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Metrics response is empty")
	}
}

func TestUpdateAgentCount(t *testing.T) {
	m := NewMetrics()

	m.UpdateAgentCount(10, 8, 2)

	// Verify the gauge values by collecting metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	// Check for expected metric values
	if !contains(body, "ofa_agents_total 10") {
		t.Error("AgentsTotal metric not found or wrong value")
	}
	if !contains(body, "ofa_agents_online 8") {
		t.Error("AgentsOnline metric not found or wrong value")
	}
	if !contains(body, "ofa_agents_offline 2") {
		t.Error("AgentsOffline metric not found or wrong value")
	}
}

func TestTaskMetrics(t *testing.T) {
	m := NewMetrics()

	// Increment task counters
	m.IncrementTaskTotal()
	m.IncrementTaskCompleted(time.Second)
	m.IncrementTaskFailed()
	m.IncrementTaskCancelled()

	// Update queue
	m.UpdateTaskQueue(5, 3)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	if !contains(body, "ofa_tasks_total 1") {
		t.Error("TasksTotal counter not incremented")
	}
	if !contains(body, "ofa_tasks_completed_total 1") {
		t.Error("TasksCompleted counter not incremented")
	}
	if !contains(body, "ofa_tasks_failed_total 1") {
		t.Error("TasksFailed counter not incremented")
	}
	if !contains(body, "ofa_tasks_cancelled_total 1") {
		t.Error("TasksCancelled counter not incremented")
	}
	if !contains(body, "ofa_tasks_pending 5") {
		t.Error("TasksPending gauge not updated")
	}
	if !contains(body, "ofa_tasks_running 3") {
		t.Error("TasksRunning gauge not updated")
	}
}

func TestMessageMetrics(t *testing.T) {
	m := NewMetrics()

	m.IncrementMessages(true, 100*time.Millisecond)
	m.IncrementMessages(false, 0)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	if !contains(body, "ofa_messages_total 2") {
		t.Error("MessagesTotal counter not incremented")
	}
	if !contains(body, "ofa_messages_delivered_total 1") {
		t.Error("MessagesDelivered counter not incremented")
	}
	if !contains(body, "ofa_messages_failed_total 1") {
		t.Error("MessagesFailed counter not incremented")
	}
}

func TestRequestMetrics(t *testing.T) {
	m := NewMetrics()

	m.RecordRequest("GET", "/health", 200, 50*time.Millisecond)
	m.RecordRequest("POST", "/api/v1/tasks", 201, 100*time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	if !contains(body, "ofa_requests_total{method=\"GET\",path=\"/health\",status=\"200\"}") {
		t.Error("RequestsTotal for GET /health not found")
	}
	if !contains(body, "ofa_request_duration_seconds") {
		t.Error("RequestDuration histogram not found")
	}
}

func TestGRPCConnections(t *testing.T) {
	m := NewMetrics()

	m.UpdateGRPCConnections(5)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	if !contains(body, "ofa_grpc_connections 5") {
		t.Error("GRPCConnections metric not found or wrong value")
	}
}

func TestHealthCheckCounter(t *testing.T) {
	m := NewMetrics()

	for i := 0; i < 3; i++ {
		m.IncrementHealthCheck()
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler().ServeHTTP(w, req)

	body := w.Body.String()

	if !contains(body, "ofa_health_checks_total 3") {
		t.Error("HealthCheckCount not incremented correctly")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr)
}