package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the OFA Center
type Metrics struct {
	// Agent metrics
	AgentsTotal       prometheus.Gauge
	AgentsOnline      prometheus.Gauge
	AgentsOffline     prometheus.Gauge
	AgentsByType      *prometheus.GaugeVec

	// Task metrics
	TasksTotal        prometheus.Counter
	TasksCompleted    prometheus.Counter
	TasksFailed       prometheus.Counter
	TasksCancelled    prometheus.Counter
	TasksPending      prometheus.Gauge
	TasksRunning      prometheus.Gauge
	TaskDuration      prometheus.Histogram

	// Message metrics
	MessagesTotal     prometheus.Counter
	MessagesDelivered prometheus.Counter
	MessagesFailed    prometheus.Counter
	MessageLatency    prometheus.Histogram

	// System metrics
	RequestDuration   *prometheus.HistogramVec
	RequestsTotal    *prometheus.CounterVec
	GRPCConnections   prometheus.Gauge
	HealthCheckCount  prometheus.Counter

	registry *prometheus.Registry
	mu       sync.RWMutex
}

// NewMetrics creates a new Metrics instance with all collectors registered
func NewMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
	}

	// Agent metrics
	m.AgentsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_agents_total",
		Help: "Total number of registered agents",
	})
	m.AgentsOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_agents_online",
		Help: "Number of online agents",
	})
	m.AgentsOffline = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_agents_offline",
		Help: "Number of offline agents",
	})
	m.AgentsByType = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ofa_agents_by_type",
			Help: "Number of agents by type",
		},
		[]string{"type"},
	)

	// Task metrics
	m.TasksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_tasks_total",
		Help: "Total number of tasks submitted",
	})
	m.TasksCompleted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_tasks_completed_total",
		Help: "Total number of completed tasks",
	})
	m.TasksFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_tasks_failed_total",
		Help: "Total number of failed tasks",
	})
	m.TasksCancelled = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_tasks_cancelled_total",
		Help: "Total number of cancelled tasks",
	})
	m.TasksPending = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_tasks_pending",
		Help: "Number of pending tasks",
	})
	m.TasksRunning = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_tasks_running",
		Help: "Number of running tasks",
	})
	m.TaskDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ofa_task_duration_seconds",
		Help:    "Duration of task execution in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
	})

	// Message metrics
	m.MessagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_messages_total",
		Help: "Total number of messages sent",
	})
	m.MessagesDelivered = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_messages_delivered_total",
		Help: "Total number of messages delivered",
	})
	m.MessagesFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_messages_failed_total",
		Help: "Total number of messages failed",
	})
	m.MessageLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ofa_message_latency_seconds",
		Help:    "Message delivery latency in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
	})

	// System metrics
	m.RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ofa_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
		[]string{"method", "path", "status"},
	)
	m.RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ofa_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	m.GRPCConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ofa_grpc_connections",
		Help: "Number of active gRPC connections",
	})
	m.HealthCheckCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ofa_health_checks_total",
		Help: "Total number of health checks",
	})

	// Register all collectors
	m.registry.MustRegister(
		m.AgentsTotal, m.AgentsOnline, m.AgentsOffline, m.AgentsByType,
		m.TasksTotal, m.TasksCompleted, m.TasksFailed, m.TasksCancelled,
		m.TasksPending, m.TasksRunning, m.TaskDuration,
		m.MessagesTotal, m.MessagesDelivered, m.MessagesFailed, m.MessageLatency,
		m.RequestDuration, m.RequestsTotal, m.GRPCConnections, m.HealthCheckCount,
	)

	// Register default Go metrics
	m.registry.MustRegister(prometheus.NewGoCollector())
	m.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return m
}

// Handler returns the HTTP handler for Prometheus metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// UpdateAgentCount updates agent-related metrics
func (m *Metrics) UpdateAgentCount(total, online, offline int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AgentsTotal.Set(float64(total))
	m.AgentsOnline.Set(float64(online))
	m.AgentsOffline.Set(float64(offline))
}

// UpdateAgentByType updates the count of agents by type
func (m *Metrics) UpdateAgentByType(typeCounts map[string]int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all type gauges
	m.AgentsByType.Reset()

	for typ, count := range typeCounts {
		m.AgentsByType.WithLabelValues(typ).Set(float64(count))
	}
}

// IncrementTaskTotal increments the total tasks counter
func (m *Metrics) IncrementTaskTotal() {
	m.TasksTotal.Inc()
}

// IncrementTaskCompleted increments completed tasks counter and records duration
func (m *Metrics) IncrementTaskCompleted(duration time.Duration) {
	m.TasksCompleted.Inc()
	m.TaskDuration.Observe(duration.Seconds())
}

// IncrementTaskFailed increments failed tasks counter
func (m *Metrics) IncrementTaskFailed() {
	m.TasksFailed.Inc()
}

// IncrementTaskCancelled increments cancelled tasks counter
func (m *Metrics) IncrementTaskCancelled() {
	m.TasksCancelled.Inc()
}

// UpdateTaskQueue updates pending and running task counts
func (m *Metrics) UpdateTaskQueue(pending, running int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TasksPending.Set(float64(pending))
	m.TasksRunning.Set(float64(running))
}

// IncrementMessages increments message counters
func (m *Metrics) IncrementMessages(delivered bool, latency time.Duration) {
	m.MessagesTotal.Inc()
	if delivered {
		m.MessagesDelivered.Inc()
		m.MessageLatency.Observe(latency.Seconds())
	} else {
		m.MessagesFailed.Inc()
	}
}

// RecordRequest records HTTP request metrics
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, path, string(status)).Inc()
	m.RequestDuration.WithLabelValues(method, path, string(status)).Observe(duration.Seconds())
}

// IncrementHealthCheck increments health check counter
func (m *Metrics) IncrementHealthCheck() {
	m.HealthCheckCount.Inc()
}

// UpdateGRPCConnections updates gRPC connection count
func (m *Metrics) UpdateGRPCConnections(count int) {
	m.GRPCConnections.Set(float64(count))
}

// GetRegistry returns the Prometheus registry
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}