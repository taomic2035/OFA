package cluster

import (
	"context"
	"log"
	"sync"
	"time"
)

// FailoverState represents the current failover state
type FailoverState string

const (
	FailoverStateNormal    FailoverState = "normal"
	FailoverStateDegraded  FailoverState = "degraded"
	FailoverStateRecovery  FailoverState = "recovery"
	FailoverStateFailover  FailoverState = "failover"
)

// FailoverConfig holds failover configuration
type FailoverConfig struct {
	Enabled         bool
	FailoverDelay   time.Duration // Time before triggering failover
	RecoveryDelay   time.Duration // Time before marking recovery
	MaxRetryCount   int
	RetryInterval   time.Duration
	QuorumSize      int // Minimum nodes for quorum
}

// DefaultFailoverConfig returns default configuration
func DefaultFailoverConfig() *FailoverConfig {
	return &FailoverConfig{
		Enabled:       true,
		FailoverDelay: 10 * time.Second,
		RecoveryDelay: 30 * time.Second,
		MaxRetryCount: 3,
		RetryInterval: 5 * time.Second,
		QuorumSize:    2,
	}
}

// FailoverManager handles automatic failover and recovery
type FailoverManager struct {
	config    *FailoverConfig
	discovery *ServiceDiscovery
	lb        *LoadBalancer

	state      FailoverState
	stateSince time.Time

	// Failed nodes tracking
	failedNodes sync.Map // map[string]*FailoverNodeInfo

	// Pending tasks for failed nodes
	pendingTasks sync.Map // map[string][]PendingTask

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// FailoverNodeInfo tracks information about a failed node
type FailoverNodeInfo struct {
	NodeID        string
	FailedAt      time.Time
	LastAttempt   time.Time
	AttemptCount  int
	Recovered     bool
	OriginalState NodeState
}

// PendingTask represents a task pending on a failed node
type PendingTask struct {
	TaskID    string
	TaskType  string
	Data      []byte
	Priority  int
	CreatedAt time.Time
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(config *FailoverConfig, discovery *ServiceDiscovery, lb *LoadBalancer) (*FailoverManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	fm := &FailoverManager{
		config:    config,
		discovery: discovery,
		lb:        lb,
		state:     FailoverStateNormal,
		stateSince: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}

	return fm, nil
}

// Start begins failover monitoring
func (fm *FailoverManager) Start() {
	go fm.monitorLoop()
	go fm.recoveryLoop()
}

// Stop stops the failover manager
func (fm *FailoverManager) Stop() {
	fm.cancel()
}

// GetState returns current failover state
func (fm *FailoverManager) GetState() FailoverState {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.state
}

// monitorLoop monitors node health and triggers failover
func (fm *FailoverManager) monitorLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.checkNodes()
			fm.updateState()
		}
	}
}

// recoveryLoop attempts to recover failed nodes
func (fm *FailoverManager) recoveryLoop() {
	ticker := time.NewTicker(fm.config.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.attemptRecovery()
		}
	}
}

// checkNodes checks for failed nodes
func (fm *FailoverManager) checkNodes() {
	nodes := fm.discovery.GetAllNodes()

	for _, node := range nodes {
		if node.ID == fm.discovery.GetLocalNode().ID {
			continue // Skip self
		}

		// Check if node is offline or draining
		if node.State == NodeStateOffline || node.State == NodeStateDraining {
			fm.handleNodeFailure(node)
		} else if node.State == NodeStateOnline {
			// Check LastSeen
			if node.LastSeen.Before(time.Now().Add(-fm.config.FailoverDelay)) {
				fm.handleNodeFailure(node)
			}
		}
	}
}

// handleNodeFailure handles a node failure
func (fm *FailoverManager) handleNodeFailure(node *NodeInfo) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Check if already tracked
	if v, ok := fm.failedNodes.Load(node.ID); ok {
		info := v.(*FailoverNodeInfo)
		if info.Recovered {
			// Node failed again after recovery
			info.Recovered = false
			info.FailedAt = time.Now()
			info.AttemptCount = 0
		}
		return
	}

	// Track new failure
	info := &FailoverNodeInfo{
		NodeID:        node.ID,
		FailedAt:      time.Now(),
		OriginalState: node.State,
	}
	fm.failedNodes.Store(node.ID, info)

	log.Printf("Node %s failure detected", node.ID)

	// Trigger failover
	fm.triggerFailover(node.ID)
}

// triggerFailover triggers failover for a node
func (fm *FailoverManager) triggerFailover(nodeID string) {
	// Get pending tasks from the failed node
	if v, ok := fm.pendingTasks.Load(nodeID); ok {
		tasks := v.([]PendingTask)
		log.Printf("Redistributing %d pending tasks from failed node %s", len(tasks), nodeID)

		for _, task := range tasks {
			fm.redistributeTask(task)
		}
		fm.pendingTasks.Delete(nodeID)
	}

	// Update node state to draining
	nodeInfo, ok := fm.discovery.GetNode(nodeID)
	if ok {
		// Mark node as offline in discovery
		fm.discovery.UnregisterNode(nodeID)
	}
}

// redistributeTask redistributes a task to a healthy node
func (fm *FailoverManager) redistributeTask(task PendingTask) {
	if !fm.config.Enabled {
		return
	}

	// Select a healthy node
	targetNode, err := fm.lb.SelectNode()
	if err != nil {
		log.Printf("No healthy nodes to redistribute task %s", task.TaskID)
		return
	}

	log.Printf("Redistributing task %s to node %s", task.TaskID, targetNode.ID)

	// In real implementation, would send task to target node
	// fm.sendTaskToNode(targetNode, task)
}

// updateState updates the overall failover state
func (fm *FailoverManager) updateState() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	activeCount := len(fm.discovery.GetActiveNodes())
	failedCount := 0

	fm.failedNodes.Range(func(key, value interface{}) bool {
		if !value.(*FailoverNodeInfo).Recovered {
			failedCount++
		}
		return true
	})

	// Determine state based on node counts
	if activeCount < fm.config.QuorumSize {
		fm.state = FailoverStateDegraded
	} else if failedCount > 0 {
		if fm.state == FailoverStateNormal {
			fm.state = FailoverStateFailover
		}
	} else {
		if fm.state != FailoverStateNormal {
			fm.state = FailoverStateRecovery
		}
	}

	// Reset to normal if recovery delay passed
	if fm.state == FailoverStateRecovery {
		if time.Since(fm.stateSince) > fm.config.RecoveryDelay {
			fm.state = FailoverStateNormal
			fm.stateSince = time.Now()
			log.Println("Failover recovery complete, returning to normal state")
		}
	}
}

// attemptRecovery attempts to recover failed nodes
func (fm *FailoverManager) attemptRecovery() {
	fm.failedNodes.Range(func(key, value interface{}) bool {
		info := value.(*FailoverNodeInfo)

		if info.Recovered {
			return true
		}

		if info.AttemptCount >= fm.config.MaxRetryCount {
			log.Printf("Node %s max retry attempts reached, removing from tracking", info.NodeID)
			fm.failedNodes.Delete(info.NodeID)
			return true
		}

		// Attempt to reconnect
		if time.Since(info.LastAttempt) > fm.config.RetryInterval {
			info.AttemptCount++
			info.LastAttempt = time.Now()

			// Try to ping the node
			if fm.pingNode(info.NodeID) {
				info.Recovered = true
				log.Printf("Node %s recovered after %d attempts", info.NodeID, info.AttemptCount)

				// Re-register node
				nodeInfo, ok := fm.discovery.GetNode(info.NodeID)
				if ok {
					nodeInfo.State = NodeStateOnline
					nodeInfo.LastSeen = time.Now()
					fm.discovery.RegisterNode(&nodeInfo.Node)
				}
			}
		}

		return true
	})
}

// pingNode attempts to ping a node to check if it's alive
func (fm *FailoverManager) pingNode(nodeID string) bool {
	// In real implementation, would send a ping request
	// For now, just check if we can get node info
	nodeInfo, ok := fm.discovery.GetNode(nodeID)
	if !ok {
		return false
	}

	// Check if node has responded recently
	return nodeInfo.LastSeen.After(time.Now().Add(-fm.config.RetryInterval))
}

// AddPendingTask adds a pending task for tracking
func (fm *FailoverManager) AddPendingTask(nodeID string, task PendingTask) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	var tasks []PendingTask
	if v, ok := fm.pendingTasks.Load(nodeID); ok {
		tasks = v.([]PendingTask)
	}
	tasks = append(tasks, task)
	fm.pendingTasks.Store(nodeID, tasks)
}

// RemovePendingTask removes a completed task
func (fm *FailoverManager) RemovePendingTask(nodeID, taskID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if v, ok := fm.pendingTasks.Load(nodeID); ok {
		tasks := v.([]PendingTask)
		for i, task := range tasks {
			if task.TaskID == taskID {
				tasks = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
		fm.pendingTasks.Store(nodeID, tasks)
	}
}

// GetFailedNodes returns list of failed node IDs
func (fm *FailoverManager) GetFailedNodes() []string {
	var failed []string
	fm.failedNodes.Range(func(key, value interface{}) bool {
		if !value.(*FailoverNodeInfo).Recovered {
			failed = append(failed, key.(string))
		}
		return true
	})
	return failed
}

// IsNodeFailed checks if a node is marked as failed
func (fm *FailoverManager) IsNodeFailed(nodeID string) bool {
	if v, ok := fm.failedNodes.Load(nodeID); ok {
		return !v.(*FailoverNodeInfo).Recovered
	}
	return false
}

// GetPendingCount returns total pending tasks count
func (fm *FailoverManager) GetPendingCount() int {
	total := 0
	fm.pendingTasks.Range(func(key, value interface{}) bool {
		total += len(value.([]PendingTask))
		return true
	})
	return total
}