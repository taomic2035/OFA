// Package ha provides high availability features
package ha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// DataCenterStatus defines data center status
type DataCenterStatus string

const (
	DCStatusActive   DataCenterStatus = "active"   // Fully operational
	DCStatusDegraded DataCenterStatus = "degraded" // Partially operational
	DCStatusFailover DataCenterStatus = "failover" // In failover mode
	DCStatusOffline  DataCenterStatus = "offline"  // Not available
)

// DataCenter represents a data center
type DataCenter struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	Zone         string            `json:"zone"`
	Status       DataCenterStatus  `json:"status"`
	Endpoint     string            `json:"endpoint"`
	APIToken     string            `json:"-"` // Not serialized
	Priority     int               `json:"priority"` // Lower = higher priority
	Capacity     int               `json:"capacity"` // Max capacity percentage
	Load         int               `json:"load"`     // Current load percentage
	Latency      time.Duration     `json:"latency"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	Enabled      bool              `json:"enabled"`
	IsPrimary    bool              `json:"is_primary"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
}

// FailoverConfig holds failover configuration
type FailoverConfig struct {
	// Health check
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	UnhealthyThreshold  int           `json:"unhealthy_threshold"` // Failures before unhealthy
	HealthyThreshold    int           `json:"healthy_threshold"`   // Successes before healthy

	// Failover
	AutoFailover       bool          `json:"auto_failover"`
	FailoverTimeout    time.Duration `json:"failover_timeout"`
	FailbackDelay      time.Duration `json:"failback_delay"` // Delay before failback
	ManualFailoverOnly bool          `json:"manual_failover_only"`

	// Traffic
	TrafficSplitMode string `json:"traffic_split_mode"` // active-active, active-passive
}

// DefaultFailoverConfig returns default failover configuration
func DefaultFailoverConfig() FailoverConfig {
	return FailoverConfig{
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		UnhealthyThreshold:  3,
		HealthyThreshold:    2,
		AutoFailover:        true,
		FailoverTimeout:     30 * time.Second,
		FailbackDelay:       5 * time.Minute,
		TrafficSplitMode:    "active-passive",
	}
}

// FailoverEvent represents a failover event
type FailoverEvent struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"` // failover, failback, switch
	FromDC       string           `json:"from_dc"`
	ToDC         string           `json:"to_dc"`
	Reason       string           `json:"reason"`
	TriggeredAt  time.Time        `json:"triggered_at"`
	CompletedAt  time.Time        `json:"completed_at,omitempty"`
	Status       string           `json:"status"` // initiated, completed, failed
	Error        string           `json:"error,omitempty"`
	Duration     time.Duration    `json:"duration"`
	TenantID     string           `json:"tenant_id,omitempty"`
}

// MultiDCManager manages multiple data centers
type MultiDCManager struct {
	config FailoverConfig

	// Data centers
	dataCenters sync.Map // map[string]*DataCenter
	dcList      []*DataCenter
	primaryDC   *DataCenter
	activeDC    *DataCenter

	// Health tracking
	healthStatus sync.Map // map[string]*HealthStatus

	// Failover events
	events    []*FailoverEvent
	eventLock sync.RWMutex

	// Statistics
	totalFailovers    int64
	successFailovers  int64
	totalDowntime     time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// HealthStatus tracks health of a data center
type HealthStatus struct {
	DCID            string    `json:"dc_id"`
	Healthy         bool      `json:"healthy"`
	ConsecFailures  int       `json:"consec_failures"`
	ConsecSuccesses int       `json:"consec_successes"`
	LastCheck       time.Time `json:"last_check"`
	LastHealthy     time.Time `json:"last_healthy"`
	LastUnhealthy   time.Time `json:"last_unhealthy"`
	ResponseTime    time.Duration `json:"response_time"`
}

// NewMultiDCManager creates a new multi-DC manager
func NewMultiDCManager(config FailoverConfig) *MultiDCManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &MultiDCManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start health checker
	go manager.healthChecker()

	// Start traffic monitor
	go manager.trafficMonitor()

	return manager
}

// RegisterDataCenter registers a data center
func (m *MultiDCManager) RegisterDataCenter(dc *DataCenter) error {
	if dc.ID == "" {
		return errors.New("data center ID required")
	}

	dc.CreatedAt = time.Now()
	dc.LastHeartbeat = time.Now()

	// Set as primary if first DC
	m.mu.Lock()
	if len(m.dcList) == 0 {
		dc.IsPrimary = true
		dc.Status = DCStatusActive
		m.primaryDC = dc
		m.activeDC = dc
	}
	m.dcList = append(m.dcList, dc)
	m.mu.Unlock()

	m.dataCenters.Store(dc.ID, dc)

	// Initialize health status
	m.healthStatus.Store(dc.ID, &HealthStatus{
		DCID:        dc.ID,
		Healthy:     true,
		LastCheck:   time.Now(),
		LastHealthy: time.Now(),
	})

	log.Printf("Data center registered: %s (region: %s, priority: %d)", dc.ID, dc.Region, dc.Priority)

	return nil
}

// UnregisterDataCenter unregisters a data center
func (m *MultiDCManager) UnregisterDataCenter(dcID string) error {
	dc, ok := m.dataCenters.Load(dcID)
	if !ok {
		return fmt.Errorf("data center not found: %s", dcID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from list
	for i, d := range m.dcList {
		if d.ID == dcID {
			m.dcList = append(m.dcList[:i], m.dcList[i+1:]...)
			break
		}
	}

	m.dataCenters.Delete(dcID)
	m.healthStatus.Delete(dcID)

	// Handle if it was primary or active
	if dc.(*DataCenter).IsPrimary {
		// Elect new primary
		m.electNewPrimary()
	}

	if m.activeDC != nil && m.activeDC.ID == dcID {
		// Failover to another DC
		m.performFailover(dcID, "data_center_removed")
	}

	return nil
}

// GetDataCenter retrieves a data center
func (m *MultiDCManager) GetDataCenter(dcID string) (*DataCenter, error) {
	if v, ok := m.dataCenters.Load(dcID); ok {
		return v.(*DataCenter), nil
	}
	return nil, fmt.Errorf("data center not found: %s", dcID)
}

// ListDataCenters lists all data centers
func (m *MultiDCManager) ListDataCenters() []*DataCenter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dcList
}

// GetActiveDataCenter returns the active data center
func (m *MultiDCManager) GetActiveDataCenter() *DataCenter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeDC
}

// GetPrimaryDataCenter returns the primary data center
func (m *MultiDCManager) GetPrimaryDataCenter() *DataCenter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primaryDC
}

// healthChecker periodically checks health of data centers
func (m *MultiDCManager) healthChecker() {
	ticker := time.NewTicker(m.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAllHealth()
		}
	}
}

// checkAllHealth checks health of all data centers
func (m *MultiDCManager) checkAllHealth() {
	m.dataCenters.Range(func(key, value interface{}) bool {
		dc := value.(*DataCenter)

		if !dc.Enabled {
			return true
		}

		healthy, latency := m.checkHealth(dc)

		// Update health status
		status, ok := m.healthStatus.Load(dc.ID)
		if !ok {
			return true
		}

		hs := status.(*HealthStatus)
		hs.LastCheck = time.Now()
		hs.ResponseTime = latency

		if healthy {
			hs.ConsecFailures = 0
			hs.ConsecSuccesses++
			dc.Latency = latency
			dc.LastHeartbeat = time.Now()

			if hs.ConsecSuccesses >= m.config.HealthyThreshold && !hs.Healthy {
				hs.Healthy = true
				hs.LastHealthy = time.Now()
				dc.Status = DCStatusActive
				log.Printf("Data center %s is now healthy", dc.ID)
			}
		} else {
			hs.ConsecSuccesses = 0
			hs.ConsecFailures++

			if hs.ConsecFailures >= m.config.UnhealthyThreshold && hs.Healthy {
				hs.Healthy = false
				hs.LastUnhealthy = time.Now()
				dc.Status = DCStatusDegraded
				log.Printf("Data center %s is now unhealthy", dc.ID)

				// Trigger failover if this is the active DC
				if m.activeDC != nil && m.activeDC.ID == dc.ID {
					if m.config.AutoFailover {
						go m.triggerFailover(dc.ID, "health_check_failed")
					}
				}
			}
		}

		return true
	})
}

// checkHealth performs health check on a data center
func (m *MultiDCManager) checkHealth(dc *DataCenter) (bool, time.Duration) {
	start := time.Now()

	// Simulate health check
	// In production, would make actual HTTP/gRPC call
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	// Simulate occasional failures
	healthy := rand.Float32() > 0.05 // 95% success rate

	latency := time.Since(start)

	return healthy, latency
}

// triggerFailover triggers a failover
func (m *MultiDCManager) triggerFailover(fromDCID, reason string) {
	event := &FailoverEvent{
		ID:          generateEventID(),
		Type:        "failover",
		FromDC:      fromDCID,
		Reason:      reason,
		TriggeredAt: time.Now(),
		Status:      "initiated",
	}

	m.recordEvent(event)

	// Find target DC
	targetDC := m.selectFailoverTarget(fromDCID)
	if targetDC == nil {
		event.Status = "failed"
		event.Error = "no available data center for failover"
		event.CompletedAt = time.Now()
		log.Printf("Failover failed: %s", event.Error)
		return
	}

	event.ToDC = targetDC.ID

	// Perform failover
	if err := m.performFailover(fromDCID, reason); err != nil {
		event.Status = "failed"
		event.Error = err.Error()
	} else {
		event.Status = "completed"
		m.successFailovers++
	}

	event.CompletedAt = time.Now()
	event.Duration = event.CompletedAt.Sub(event.TriggeredAt)

	m.totalFailovers++

	log.Printf("Failover %s: %s -> %s (duration: %v)", event.Status, fromDCID, event.ToDC, event.Duration)
}

// selectFailoverTarget selects a target for failover
func (m *MultiDCManager) selectFailoverTarget(excludeDCID string) *DataCenter {
	var candidates []*DataCenter

	m.dataCenters.Range(func(key, value interface{}) bool {
		dc := value.(*DataCenter)

		if dc.ID == excludeDCID || !dc.Enabled {
			return true
		}

		// Check health
		if hs, ok := m.healthStatus.Load(dc.ID); ok {
			if hs.(*HealthStatus).Healthy {
				candidates = append(candidates, dc)
			}
		}

		return true
	})

	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority
	// In production, would use proper sorting
	return candidates[0]
}

// performFailover performs the actual failover
func (m *MultiDCManager) performFailover(fromDCID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find new active DC
	var newActive *DataCenter
	for _, dc := range m.dcList {
		if dc.ID == fromDCID {
			continue
		}

		if hs, ok := m.healthStatus.Load(dc.ID); ok && hs.(*HealthStatus).Healthy && dc.Enabled {
			if newActive == nil || dc.Priority < newActive.Priority {
				newActive = dc
			}
		}
	}

	if newActive == nil {
		return errors.New("no healthy data center available")
	}

	// Mark old DC as failed over
	if oldDC, ok := m.dataCenters.Load(fromDCID); ok {
		oldDC.(*DataCenter).Status = DCStatusFailover
	}

	// Set new active
	m.activeDC = newActive
	newActive.Status = DCStatusActive

	return nil
}

// Failback fails back to primary
func (m *MultiDCManager) Failback() error {
	if m.primaryDC == nil {
		return errors.New("no primary data center configured")
	}

	// Check if primary is healthy
	hs, ok := m.healthStatus.Load(m.primaryDC.ID)
	if !ok || !hs.(*HealthStatus).Healthy {
		return errors.New("primary data center is not healthy")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	event := &FailoverEvent{
		ID:          generateEventID(),
		Type:        "failback",
		FromDC:      m.activeDC.ID,
		ToDC:        m.primaryDC.ID,
		Reason:      "manual_failback",
		TriggeredAt: time.Now(),
	}

	// Perform failback
	oldActive := m.activeDC
	m.activeDC = m.primaryDC
	m.primaryDC.Status = DCStatusActive

	if oldActive != nil && oldActive.ID != m.primaryDC.ID {
		oldActive.Status = DCStatusActive // Keep available
	}

	event.Status = "completed"
	event.CompletedAt = time.Now()
	event.Duration = event.CompletedAt.Sub(event.TriggeredAt)

	m.recordEvent(event)

	log.Printf("Failback completed: %s -> %s", event.FromDC, event.ToDC)

	return nil
}

// ManualFailover performs manual failover
func (m *MultiDCManager) ManualFailover(targetDCID string) error {
	targetDC, err := m.GetDataCenter(targetDCID)
	if err != nil {
		return err
	}

	// Check if target is healthy
	hs, ok := m.healthStatus.Load(targetDCID)
	if !ok || !hs.(*HealthStatus).Healthy {
		return fmt.Errorf("target data center %s is not healthy", targetDCID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	event := &FailoverEvent{
		ID:          generateEventID(),
		Type:        "manual_switch",
		FromDC:      m.activeDC.ID,
		ToDC:        targetDCID,
		Reason:      "manual_failover",
		TriggeredAt: time.Now(),
	}

	// Perform switch
	if m.activeDC != nil {
		m.activeDC.Status = DCStatusActive
	}
	m.activeDC = targetDC
	targetDC.Status = DCStatusActive

	event.Status = "completed"
	event.CompletedAt = time.Now()
	event.Duration = event.CompletedAt.Sub(event.TriggeredAt)

	m.recordEvent(event)

	log.Printf("Manual failover completed: %s -> %s", event.FromDC, event.ToDC)

	return nil
}

// electNewPrimary elects a new primary data center
func (m *MultiDCManager) electNewPrimary() {
	var newPrimary *DataCenter

	for _, dc := range m.dcList {
		if !dc.Enabled {
			continue
		}

		if hs, ok := m.healthStatus.Load(dc.ID); ok && hs.(*HealthStatus).Healthy {
			if newPrimary == nil || dc.Priority < newPrimary.Priority {
				newPrimary = dc
			}
		}
	}

	if newPrimary != nil {
		m.primaryDC = newPrimary
		newPrimary.IsPrimary = true
		log.Printf("New primary elected: %s", newPrimary.ID)
	}
}

// trafficMonitor monitors traffic distribution
func (m *MultiDCManager) trafficMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateLoad()
		}
	}
}

// updateLoad updates load on data centers
func (m *MultiDCManager) updateLoad() {
	m.dataCenters.Range(func(key, value interface{}) bool {
		dc := value.(*DataCenter)
		// Simulate load update
		dc.Load = rand.Intn(100)
		return true
	})
}

// recordEvent records a failover event
func (m *MultiDCManager) recordEvent(event *FailoverEvent) {
	m.eventLock.Lock()
	defer m.eventLock.Unlock()

	m.events = append(m.events, event)

	// Keep last 100 events
	if len(m.events) > 100 {
		m.events = m.events[len(m.events)-100:]
	}
}

// GetEvents returns failover events
func (m *MultiDCManager) GetEvents(limit int) []*FailoverEvent {
	m.eventLock.RLock()
	defer m.eventLock.RUnlock()

	if limit <= 0 || limit > len(m.events) {
		limit = len(m.events)
	}

	result := make([]*FailoverEvent, limit)
	copy(result, m.events[len(m.events)-limit:])
	return result
}

// GetStats returns statistics
func (m *MultiDCManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthyCount := 0
	m.healthStatus.Range(func(key, value interface{}) bool {
		if value.(*HealthStatus).Healthy {
			healthyCount++
		}
		return true
	})

	return map[string]interface{}{
		"total_data_centers":   len(m.dcList),
		"healthy_count":        healthyCount,
		"primary_dc":           m.primaryDC.ID,
		"active_dc":            m.activeDC.ID,
		"total_failovers":      m.totalFailovers,
		"successful_failovers": m.successFailovers,
		"total_downtime":       m.totalDowntime.String(),
		"auto_failover":        m.config.AutoFailover,
	}
}

// Close closes the manager
func (m *MultiDCManager) Close() {
	m.cancel()
}

// Helper function
func generateEventID() string {
	return fmt.Sprintf("event-%d", time.Now().UnixNano())
}