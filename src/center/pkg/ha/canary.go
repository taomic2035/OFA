// Package ha provides canary deployment capabilities
package ha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// CanaryStatus defines canary deployment status
type CanaryStatus string

const (
	CanaryPending   CanaryStatus = "pending"
	CanaryRunning   CanaryStatus = "running"
	CanaryPromoting CanaryStatus = "promoting"
	CanaryCompleted CanaryStatus = "completed"
	CanaryRolledback CanaryStatus = "rolledback"
	CanaryFailed    CanaryStatus = "failed"
)

// CanaryStrategy defines canary deployment strategy
type CanaryStrategy string

const (
	StrategyPercentage CanaryStrategy = "percentage" // Traffic percentage
	StrategyDuration   CanaryStrategy = "duration"   // Time-based
	StrategyManual     CanaryStrategy = "manual"     // Manual approval
)

// CanaryDeployment represents a canary deployment
type CanaryDeployment struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Version         string           `json:"version"`
	TargetVersion   string           `json:"target_version"`
	Status          CanaryStatus     `json:"status"`
	Strategy        CanaryStrategy   `json:"strategy"`

	// Traffic configuration
	InitialTraffic  float64          `json:"initial_traffic"` // Initial traffic percentage
	CurrentTraffic  float64          `json:"current_traffic"` // Current traffic percentage
	TargetTraffic   float64          `json:"target_traffic"`  // Target traffic percentage
	IncrementStep   float64          `json:"increment_step"`  // Step size for increments
	IncrementInterval time.Duration  `json:"increment_interval"`

	// Duration configuration
	Duration        time.Duration    `json:"duration"`

	// Metrics thresholds
	SuccessThreshold float64         `json:"success_threshold"` // Success rate threshold
	LatencyThreshold time.Duration   `json:"latency_threshold"`
	ErrorThreshold   float64         `json:"error_threshold"`

	// Metrics
	Metrics         CanaryMetrics    `json:"metrics"`

	// Timing
	CreatedAt       time.Time        `json:"created_at"`
	StartedAt       time.Time        `json:"started_at,omitempty"`
	CompletedAt     time.Time        `json:"completed_at,omitempty"`

	// Rollback
	AutoRollback    bool             `json:"auto_rollback"`
	RollbackReason  string           `json:"rollback_reason,omitempty"`
	RolledBackAt    time.Time        `json:"rolled_back_at,omitempty"`

	// Approval
	RequireApproval bool             `json:"require_approval"`
	ApprovedBy      string           `json:"approved_by,omitempty"`
	ApprovedAt      time.Time        `json:"approved_at,omitempty"`

	// Data center
	DataCenter      string           `json:"data_center"`
	TenantID        string           `json:"tenant_id,omitempty"`
}

// CanaryMetrics holds canary deployment metrics
type CanaryMetrics struct {
	// Traffic
	RequestCount    int64   `json:"request_count"`
	CanaryRequests  int64   `json:"canary_requests"`

	// Success
	SuccessCount    int64   `json:"success_count"`
	FailureCount    int64   `json:"failure_count"`
	SuccessRate     float64 `json:"success_rate"`

	// Latency
	AvgLatency      time.Duration `json:"avg_latency"`
	P50Latency      time.Duration `json:"p50_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`

	// Errors
	ErrorRate       float64 `json:"error_rate"`
	Errors          []string `json:"errors"`

	// Comparison (vs stable)
	LatencyDiff     float64 `json:"latency_diff"` // Percentage difference
	SuccessRateDiff float64 `json:"success_rate_diff"`
}

// CanaryManager manages canary deployments
type CanaryManager struct {
	// Deployments
	deployments sync.Map // map[string]*CanaryDeployment
	active      *CanaryDeployment

	// Metrics collector
	collector MetricsCollector

	// Statistics
	totalDeployments    int64
	successfulPromotions int64
	rollbacks           int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// MetricsCollector collects metrics for canary analysis
type MetricsCollector interface {
	Collect(version string, duration time.Duration) (*CanaryMetrics, error)
}

// NewCanaryManager creates a new canary manager
func NewCanaryManager() *CanaryManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &CanaryManager{
		ctx:    ctx,
		cancel: cancel,
	}

	// Start monitor
	go manager.monitor()

	return manager
}

// CreateDeployment creates a new canary deployment
func (m *CanaryManager) CreateDeployment(deployment *CanaryDeployment) error {
	if deployment.ID == "" {
		deployment.ID = generateDeploymentID()
	}

	deployment.Status = CanaryPending
	deployment.CreatedAt = time.Now()

	// Set defaults
	if deployment.InitialTraffic == 0 {
		deployment.InitialTraffic = 5.0 // 5% initial traffic
	}
	if deployment.TargetTraffic == 0 {
		deployment.TargetTraffic = 100.0
	}
	if deployment.IncrementStep == 0 {
		deployment.IncrementStep = 10.0
	}
	if deployment.IncrementInterval == 0 {
		deployment.IncrementInterval = 5 * time.Minute
	}
	if deployment.SuccessThreshold == 0 {
		deployment.SuccessThreshold = 0.99 // 99% success rate
	}
	if deployment.Duration == 0 {
		deployment.Duration = 30 * time.Minute
	}

	m.deployments.Store(deployment.ID, deployment)
	m.totalDeployments++

	return nil
}

// StartDeployment starts a canary deployment
func (m *CanaryManager) StartDeployment(deploymentID string) error {
	deployment, ok := m.deployments.Load(deploymentID)
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	d := deployment.(*CanaryDeployment)

	m.mu.Lock()
	defer m.mu.Unlock()

	if d.Status != CanaryPending {
		return fmt.Errorf("deployment not in pending state: %s", d.Status)
	}

	// Check approval if required
	if d.RequireApproval && d.ApprovedBy == "" {
		return errors.New("deployment requires approval")
	}

	d.Status = CanaryRunning
	d.StartedAt = time.Now()
	d.CurrentTraffic = d.InitialTraffic

	m.active = d

	log.Printf("Canary deployment started: %s (%s -> %s, initial traffic: %.1f%%)",
		d.ID, d.Version, d.TargetVersion, d.CurrentTraffic)

	return nil
}

// ApproveDeployment approves a deployment
func (m *CanaryManager) ApproveDeployment(deploymentID, approvedBy string) error {
	deployment, ok := m.deployments.Load(deploymentID)
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	d := deployment.(*CanaryDeployment)

	d.ApprovedBy = approvedBy
	d.ApprovedAt = time.Now()

	log.Printf("Canary deployment approved: %s by %s", deploymentID, approvedBy)

	return nil
}

// monitor monitors active canary deployments
func (m *CanaryManager) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkDeployment()
		}
	}
}

// checkDeployment checks the active deployment
func (m *CanaryManager) checkDeployment() {
	m.mu.RLock()
	if m.active == nil || m.active.Status != CanaryRunning {
		m.mu.RUnlock()
		return
	}
	deployment := m.active
	m.mu.RUnlock()

	// Collect metrics
	metrics := m.collectMetrics(deployment)
	deployment.Metrics = metrics

	// Check thresholds
	if m.shouldRollback(deployment, metrics) {
		m.RollbackDeployment(deployment.ID, "threshold_exceeded")
		return
	}

	// Check if should increment traffic
	if m.shouldIncrement(deployment) {
		m.incrementTraffic(deployment)
	}

	// Check if should complete
	if m.shouldComplete(deployment) {
		m.PromoteDeployment(deployment.ID)
	}
}

// collectMetrics collects metrics for a deployment
func (m *CanaryManager) collectMetrics(deployment *CanaryDeployment) CanaryMetrics {
	// Simulate metrics collection
	// In production, would query actual metrics

	requestCount := int64(rand.Intn(10000) + 1000)
	successCount := int64(float64(requestCount) * (0.95 + rand.Float64()*0.04))
	failureCount := requestCount - successCount

	return CanaryMetrics{
		RequestCount:    requestCount,
		CanaryRequests:  int64(float64(requestCount) * deployment.CurrentTraffic / 100),
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		SuccessRate:     float64(successCount) / float64(requestCount),
		AvgLatency:      time.Duration(rand.Intn(100)+50) * time.Millisecond,
		P50Latency:      time.Duration(rand.Intn(80)+40) * time.Millisecond,
		P95Latency:      time.Duration(rand.Intn(150)+100) * time.Millisecond,
		P99Latency:      time.Duration(rand.Intn(200)+150) * time.Millisecond,
		ErrorRate:       float64(failureCount) / float64(requestCount),
		LatencyDiff:     (rand.Float64() - 0.5) * 10, // -5% to +5%
		SuccessRateDiff: (rand.Float64() - 0.5) * 2,  // -1% to +1%
	}
}

// shouldRollback determines if deployment should be rolled back
func (m *CanaryManager) shouldRollback(deployment *CanaryDeployment, metrics CanaryMetrics) bool {
	if !deployment.AutoRollback {
		return false
	}

	// Check success rate
	if metrics.SuccessRate < deployment.SuccessThreshold {
		log.Printf("Canary rollback triggered: success rate %.2f%% < threshold %.2f%%",
			metrics.SuccessRate*100, deployment.SuccessThreshold*100)
		return true
	}

	// Check latency
	if deployment.LatencyThreshold > 0 && metrics.AvgLatency > deployment.LatencyThreshold {
		log.Printf("Canary rollback triggered: latency %v > threshold %v",
			metrics.AvgLatency, deployment.LatencyThreshold)
		return true
	}

	// Check error rate
	if deployment.ErrorThreshold > 0 && metrics.ErrorRate > deployment.ErrorThreshold {
		log.Printf("Canary rollback triggered: error rate %.2f%% > threshold %.2f%%",
			metrics.ErrorRate*100, deployment.ErrorThreshold*100)
		return true
	}

	return false
}

// shouldIncrement determines if traffic should be incremented
func (m *CanaryManager) shouldIncrement(deployment *CanaryDeployment) bool {
	if deployment.Strategy == StrategyManual {
		return false
	}

	// Check time since last increment
	return time.Since(deployment.StartedAt) >= deployment.IncrementInterval
}

// shouldComplete determines if deployment should complete
func (m *CanaryManager) shouldComplete(deployment *CanaryDeployment) bool {
	return deployment.CurrentTraffic >= deployment.TargetTraffic
}

// incrementTraffic increments traffic for a deployment
func (m *CanaryManager) incrementTraffic(deployment *CanaryDeployment) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newTraffic := math.Min(deployment.CurrentTraffic+deployment.IncrementStep, deployment.TargetTraffic)
	log.Printf("Canary traffic increment: %.1f%% -> %.1f%%", deployment.CurrentTraffic, newTraffic)
	deployment.CurrentTraffic = newTraffic
}

// PromoteDeployment promotes a canary deployment
func (m *CanaryManager) PromoteDeployment(deploymentID string) error {
	deployment, ok := m.deployments.Load(deploymentID)
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	d := deployment.(*CanaryDeployment)

	m.mu.Lock()
	d.Status = CanaryPromoting
	m.mu.Unlock()

	// Simulate promotion process
	time.Sleep(1 * time.Second)

	m.mu.Lock()
	d.Status = CanaryCompleted
	d.CurrentTraffic = 100
	d.CompletedAt = time.Now()
	m.active = nil
	m.successfulPromotions++
	m.mu.Unlock()

	log.Printf("Canary deployment promoted: %s", deploymentID)

	return nil
}

// RollbackDeployment rolls back a canary deployment
func (m *CanaryManager) RollbackDeployment(deploymentID, reason string) error {
	deployment, ok := m.deployments.Load(deploymentID)
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentID)
	}

	d := deployment.(*CanaryDeployment)

	m.mu.Lock()
	d.Status = CanaryRolledback
	d.RollbackReason = reason
	d.RolledBackAt = time.Now()
	d.CurrentTraffic = 0
	m.active = nil
	m.rollbacks++
	m.mu.Unlock()

	log.Printf("Canary deployment rolled back: %s (reason: %s)", deploymentID, reason)

	return nil
}

// GetDeployment retrieves a deployment
func (m *CanaryManager) GetDeployment(deploymentID string) (*CanaryDeployment, error) {
	if v, ok := m.deployments.Load(deploymentID); ok {
		return v.(*CanaryDeployment), nil
	}
	return nil, fmt.Errorf("deployment not found: %s", deploymentID)
}

// ListDeployments lists deployments
func (m *CanaryManager) ListDeployments(status CanaryStatus) []*CanaryDeployment {
	var result []*CanaryDeployment

	m.deployments.Range(func(key, value interface{}) bool {
		d := value.(*CanaryDeployment)
		if status == "" || d.Status == status {
			result = append(result, d)
		}
		return true
	})

	return result
}

// GetActiveDeployment returns the active deployment
func (m *CanaryManager) GetActiveDeployment() *CanaryDeployment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
}

// GetStats returns statistics
func (m *CanaryManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_deployments":      m.totalDeployments,
		"successful_promotions":  m.successfulPromotions,
		"rollbacks":              m.rollbacks,
		"active_deployment":      m.active != nil,
		"success_rate":           fmt.Sprintf("%.2f%%", float64(m.successfulPromotions)/float64(m.totalDeployments)*100),
	}
}

// Close closes the manager
func (m *CanaryManager) Close() {
	m.cancel()
}

// Helper function
func generateDeploymentID() string {
	return fmt.Sprintf("canary-%d", time.Now().UnixNano())
}