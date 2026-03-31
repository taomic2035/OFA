package cluster

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancerType defines load balancer types
type LoadBalancerType string

const (
	LoadBalancerRoundRobin   LoadBalancerType = "round_robin"
	LoadBalancerLeastConn    LoadBalancerType = "least_connections"
	LoadBalancerLeastLoad    LoadBalancerType = "least_load"
	LoadBalancerWeighted     LoadBalancerType = "weighted"
	LoadBalancerRegionAware  LoadBalancerType = "region_aware"
)

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	Type            LoadBalancerType
	LocalRegion     string
	HealthCheckInt  time.Duration
	FailThreshold   int           // Number of failures before marking unhealthy
	RecoverThreshold int          // Number of successes before marking healthy
	RetryCount      int
	RetryDelay      time.Duration
}

// DefaultLoadBalancerConfig returns default configuration
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		Type:             LoadBalancerLeastLoad,
		HealthCheckInt:   10 * time.Second,
		FailThreshold:    3,
		RecoverThreshold: 2,
		RetryCount:       3,
		RetryDelay:       100 * time.Millisecond,
	}
}

// NodeStatus tracks node health and connection status
type NodeStatus struct {
	NodeID       string
	Healthy      bool
	FailCount    int
	SuccessCount int
	Connections  int
	LastCheck    time.Time
	LastError    error
}

// LoadBalancer manages request distribution across nodes
type LoadBalancer struct {
	config    *LoadBalancerConfig
	discovery *ServiceDiscovery

	// Node status tracking
	status sync.Map // map[string]*NodeStatus

	// Round robin counter
	rrCounter uint64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(config *LoadBalancerConfig, discovery *ServiceDiscovery) (*LoadBalancer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	lb := &LoadBalancer{
		config:    config,
		discovery: discovery,
		ctx:       ctx,
		cancel:    cancel,
	}

	return lb, nil
}

// Start begins health checking
func (lb *LoadBalancer) Start() {
	go lb.healthCheckLoop()
}

// Stop stops the load balancer
func (lb *LoadBalancer) Stop() {
	lb.cancel()
}

// SelectNode selects a node for routing based on the configured strategy
func (lb *LoadBalancer) SelectNode() (*NodeInfo, error) {
	nodes := lb.getHealthyNodes()
	if len(nodes) == 0 {
		return nil, errors.New("no healthy nodes available")
	}

	switch lb.config.Type {
	case LoadBalancerRoundRobin:
		return lb.selectRoundRobin(nodes), nil
	case LoadBalancerLeastConn:
		return lb.selectLeastConnections(nodes), nil
	case LoadBalancerLeastLoad:
		return lb.selectLeastLoad(nodes), nil
	case LoadBalancerWeighted:
		return lb.selectWeighted(nodes), nil
	case LoadBalancerRegionAware:
		return lb.selectRegionAware(nodes), nil
	default:
		return lb.selectLeastLoad(nodes), nil
	}
}

// getHealthyNodes returns all healthy nodes
func (lb *LoadBalancer) getHealthyNodes() []*NodeInfo {
	var healthy []*NodeInfo
	nodes := lb.discovery.GetActiveNodes()

	for _, node := range nodes {
		if lb.isNodeHealthy(node.ID) {
			healthy = append(healthy, node)
		}
	}

	return healthy
}

// isNodeHealthy checks if a node is healthy
func (lb *LoadBalancer) isNodeHealthy(nodeID string) bool {
	if v, ok := lb.status.Load(nodeID); ok {
		return v.(*NodeStatus).Healthy
	}
	// New nodes are considered healthy initially
	return true
}

// selectRoundRobin selects nodes in round-robin order
func (lb *LoadBalancer) selectRoundRobin(nodes []*NodeInfo) *NodeInfo {
	idx := atomic.AddUint64(&lb.rrCounter, 1) % uint64(len(nodes))
	return nodes[idx]
}

// selectLeastConnections selects the node with least connections
func (lb *LoadBalancer) selectLeastConnections(nodes []*NodeInfo) *NodeInfo {
	selected := nodes[0]
	minConn := lb.getConnectionCount(selected.ID)

	for _, node := range nodes[1:] {
		conn := lb.getConnectionCount(node.ID)
		if conn < minConn {
			minConn = conn
			selected = node
		}
	}

	return selected
}

// selectLeastLoad selects the node with lowest load
func (lb *LoadBalancer) selectLeastLoad(nodes []*NodeInfo) *NodeInfo {
	selected := nodes[0]

	for _, node := range nodes[1:] {
		if node.Load < selected.Load {
			selected = node
		}
	}

	return selected
}

// selectWeighted selects based on priority weights
func (lb *LoadBalancer) selectWeighted(nodes []*NodeInfo) *NodeInfo {
	// Use priority as weight
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += node.Priority + 1 // +1 to ensure non-zero
	}

	// Simple weighted selection based on cumulative weight
	selected := nodes[0]
	maxWeight := 0

	for _, node := range nodes {
		weight := node.Priority + 1
		if weight > maxWeight {
			maxWeight = weight
			selected = node
		}
	}

	return selected
}

// selectRegionAware prefers local region nodes
func (lb *LoadBalancer) selectRegionAware(nodes []*NodeInfo) *NodeInfo {
	var localNodes, remoteNodes []*NodeInfo

	for _, node := range nodes {
		if node.Region == lb.config.LocalRegion {
			localNodes = append(localNodes, node)
		} else {
			remoteNodes = append(remoteNodes, node)
		}
	}

	// Prefer local region with least load
	if len(localNodes) > 0 {
		return lb.selectLeastLoad(localNodes)
	}

	// Fall back to remote with least load
	if len(remoteNodes) > 0 {
		return lb.selectLeastLoad(remoteNodes)
	}

	return nil
}

// getConnectionCount returns connection count for a node
func (lb *LoadBalancer) getConnectionCount(nodeID string) int {
	if v, ok := lb.status.Load(nodeID); ok {
		return v.(*NodeStatus).Connections
	}
	return 0
}

// IncrementConnection increments connection count
func (lb *LoadBalancer) IncrementConnection(nodeID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	status := lb.getOrCreateStatus(nodeID)
	status.Connections++
}

// DecrementConnection decrements connection count
func (lb *LoadBalancer) DecrementConnection(nodeID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	status := lb.getOrCreateStatus(nodeID)
	if status.Connections > 0 {
		status.Connections--
	}
}

// RecordSuccess records a successful request
func (lb *LoadBalancer) RecordSuccess(nodeID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	status := lb.getOrCreateStatus(nodeID)
	status.FailCount = 0
	status.SuccessCount++
	status.LastCheck = time.Now()

	// Check if node should be marked healthy
	if !status.Healthy && status.SuccessCount >= lb.config.RecoverThreshold {
		status.Healthy = true
		log.Printf("Node %s recovered", nodeID)
	}
}

// RecordFailure records a failed request
func (lb *LoadBalancer) RecordFailure(nodeID string, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	status := lb.getOrCreateStatus(nodeID)
	status.FailCount++
	status.SuccessCount = 0
	status.LastError = err
	status.LastCheck = time.Now()

	// Check if node should be marked unhealthy
	if status.Healthy && status.FailCount >= lb.config.FailThreshold {
		status.Healthy = false
		log.Printf("Node %s marked unhealthy: %v", nodeID, err)
	}
}

// getOrCreateStatus gets or creates node status
func (lb *LoadBalancer) getOrCreateStatus(nodeID string) *NodeStatus {
	if v, ok := lb.status.Load(nodeID); ok {
		return v.(*NodeStatus)
	}

	status := &NodeStatus{
		NodeID:  nodeID,
		Healthy: true,
	}
	lb.status.Store(nodeID, status)
	return status
}

// healthCheckLoop performs periodic health checks
func (lb *LoadBalancer) healthCheckLoop() {
	ticker := time.NewTicker(lb.config.HealthCheckInt)
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case <-ticker.C:
			lb.checkAllNodes()
		}
	}
}

// checkAllNodes checks health of all known nodes
func (lb *LoadBalancer) checkAllNodes() {
	nodes := lb.discovery.GetAllNodes()

	for _, node := range nodes {
		if node.ID == lb.discovery.GetLocalNode().ID {
			continue // Skip self
		}

		go lb.checkNodeHealth(node)
	}
}

// checkNodeHealth performs a health check on a node
func (lb *LoadBalancer) checkNodeHealth(node *NodeInfo) {
	// Use discovery's heartbeat mechanism
	// If heartbeat succeeds, record success
	// Otherwise, record failure
	healthy := node.State == NodeStateOnline && node.LastSeen.After(time.Now().Add(-2 * lb.config.HealthCheckInt))

	if healthy {
		lb.RecordSuccess(node.ID)
	} else {
		lb.RecordFailure(node.ID, errors.New("node not responding"))
	}
}

// GetNodeStatus returns status for all nodes
func (lb *LoadBalancer) GetNodeStatus() map[string]*NodeStatus {
	result := make(map[string]*NodeStatus)
	lb.status.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*NodeStatus)
		return true
	})
	return result
}

// GetHealthyCount returns the number of healthy nodes
func (lb *LoadBalancer) GetHealthyCount() int {
	count := 0
	lb.status.Range(func(key, value interface{}) bool {
		if value.(*NodeStatus).Healthy {
			count++
		}
		return true
	})
	return count
}