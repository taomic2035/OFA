package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// NodeState represents the state of a cluster node
type NodeState string

const (
	NodeStateOnline    NodeState = "online"
	NodeStateOffline   NodeState = "offline"
	NodeStateDraining  NodeState = "draining"
	NodeStateRecovering NodeState = "recovering"
)

// Node represents a cluster node
type Node struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	GRPCPort    int               `json:"grpc_port"`
	HTTPPort    int               `json:"http_port"`
	State       NodeState         `json:"state"`
	Load        int               `json:"load"`        // 0-100
	Priority    int               `json:"priority"`    // Higher = preferred
	Region      string            `json:"region"`
	Zone        string            `json:"zone"`
	LastSeen    time.Time         `json:"last_seen"`
	StartedAt   time.Time         `json:"started_at"`
	Metadata    map[string]string `json:"metadata"`
}

// NodeInfo contains runtime node information
type NodeInfo struct {
	Node
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	AgentCount   int     `json:"agent_count"`
	TaskCount    int     `json:"task_count"`
	ActiveTasks  int     `json:"active_tasks"`
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	NodeID       string
	NodeName     string
	Address      string
	GRPCPort     int
	HTTPPort     int
	Region       string
	Zone         string
	Priority     int
	Peers        []string // Initial peer addresses
	HeartbeatInt time.Duration
}

// DefaultClusterConfig returns default configuration
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		HeartbeatInt: 10 * time.Second,
	}
}

// ServiceDiscovery manages cluster node discovery
type ServiceDiscovery struct {
	config    *ClusterConfig
	localNode *Node
	nodes     sync.Map // map[string]*NodeInfo
	client    *http.Client

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewServiceDiscovery creates a new service discovery manager
func NewServiceDiscovery(config *ClusterConfig) (*ServiceDiscovery, error) {
	ctx, cancel := context.WithCancel(context.Background())

	sd := &ServiceDiscovery{
		config: config,
		localNode: &Node{
			ID:        config.NodeID,
			Name:      config.NodeName,
			Address:   config.Address,
			GRPCPort:  config.GRPCPort,
			HTTPPort:  config.HTTPPort,
			State:     NodeStateOnline,
			Priority:  config.Priority,
			Region:    config.Region,
			Zone:      config.Zone,
			StartedAt: time.Now(),
		},
		client: &http.Client{Timeout: 5 * time.Second},
		ctx:    ctx,
		cancel: cancel,
	}

	// Register self
	sd.nodes.Store(config.NodeID, &NodeInfo{Node: *sd.localNode})

	return sd, nil
}

// Start begins the discovery process
func (sd *ServiceDiscovery) Start() {
	go sd.heartbeatLoop()
	go sd.peerDiscoveryLoop()
}

// Stop stops the discovery process
func (sd *ServiceDiscovery) Stop() {
	sd.cancel()
}

// RegisterNode registers a node with the cluster
func (sd *ServiceDiscovery) RegisterNode(node *Node) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	node.LastSeen = time.Now()
	sd.nodes.Store(node.ID, &NodeInfo{Node: *node})
}

// UnregisterNode removes a node from the cluster
func (sd *ServiceDiscovery) UnregisterNode(nodeID string) {
	sd.nodes.Delete(nodeID)
}

// GetNode retrieves a node by ID
func (sd *ServiceDiscovery) GetNode(nodeID string) (*NodeInfo, bool) {
	if v, ok := sd.nodes.Load(nodeID); ok {
		return v.(*NodeInfo), true
	}
	return nil, false
}

// GetAllNodes returns all registered nodes
func (sd *ServiceDiscovery) GetAllNodes() []*NodeInfo {
	var nodes []*NodeInfo
	sd.nodes.Range(func(key, value interface{}) bool {
		nodes = append(nodes, value.(*NodeInfo))
		return true
	})
	return nodes
}

// GetActiveNodes returns all online nodes
func (sd *ServiceDiscovery) GetActiveNodes() []*NodeInfo {
	var nodes []*NodeInfo
	sd.nodes.Range(func(key, value interface{}) bool {
		info := value.(*NodeInfo)
		if info.State == NodeStateOnline {
			nodes = append(nodes, info)
		}
		return true
	})
	return nodes
}

// GetLocalNode returns the local node info
func (sd *ServiceDiscovery) GetLocalNode() *Node {
	return sd.localNode
}

// UpdateLocalNode updates local node information
func (sd *ServiceDiscovery) UpdateLocalNode(update func(*NodeInfo)) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if v, ok := sd.nodes.Load(sd.config.NodeID); ok {
		info := v.(*NodeInfo)
		update(info)
		info.LastSeen = time.Now()
	}
}

// SelectNode selects a node for task routing
func (sd *ServiceDiscovery) SelectNode(strategy SelectionStrategy) *NodeInfo {
	nodes := sd.GetActiveNodes()
	if len(nodes) == 0 {
		return nil
	}
	return strategy.Select(nodes)
}

// heartbeatLoop sends periodic heartbeats to peers
func (sd *ServiceDiscovery) heartbeatLoop() {
	ticker := time.NewTicker(sd.config.HeartbeatInt)
	defer ticker.Stop()

	for {
		select {
		case <-sd.ctx.Done():
			return
		case <-ticker.C:
			sd.sendHeartbeats()
		}
	}
}

// peerDiscoveryLoop discovers new peers
func (sd *ServiceDiscovery) peerDiscoveryLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial discovery
	sd.discoverPeers()

	for {
		select {
		case <-sd.ctx.Done():
			return
		case <-ticker.C:
			sd.discoverPeers()
		}
	}
}

// sendHeartbeats sends heartbeat to all known peers
func (sd *ServiceDiscovery) sendHeartbeats() {
	nodes := sd.GetAllNodes()

	for _, node := range nodes {
		if node.ID == sd.config.NodeID {
			continue // Skip self
		}

		go func(node *NodeInfo) {
			url := fmt.Sprintf("http://%s:%d/cluster/heartbeat", node.Address, node.HTTPPort)

			heartbeat := struct {
				NodeID    string    `json:"node_id"`
				State     NodeState `json:"state"`
				Load      int       `json:"load"`
				Timestamp time.Time `json:"timestamp"`
			}{
				NodeID:    sd.config.NodeID,
				State:     sd.localNode.State,
				Load:      sd.localNode.Load,
				Timestamp: time.Now(),
			}

			body, _ := json.Marshal(heartbeat)
			req, err := http.NewRequestWithContext(sd.ctx, "POST", url, nil)
			if err != nil {
				return
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
			defer req.Body.Close()

			_, err = sd.client.Do(req)
			if err != nil {
				log.Printf("Heartbeat failed to %s: %v", node.ID, err)
			}
		}(node)
	}
}

// discoverPeers discovers new peers from initial peer list
func (sd *ServiceDiscovery) discoverPeers() {
	for _, peerAddr := range sd.config.Peers {
		go func(addr string) {
			url := fmt.Sprintf("http://%s/cluster/nodes", addr)

			req, err := http.NewRequestWithContext(sd.ctx, "GET", url, nil)
			if err != nil {
				return
			}

			resp, err := sd.client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var nodes []*NodeInfo
			if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
				return
			}

			for _, node := range nodes {
				if node.ID != sd.config.NodeID {
					sd.RegisterNode(&node.Node)
				}
			}
		}(peerAddr)
	}
}

// SelectionStrategy defines node selection strategy
type SelectionStrategy interface {
	Select(nodes []*NodeInfo) *NodeInfo
}

// RoundRobinStrategy selects nodes in round-robin order
type RoundRobinStrategy struct {
	counter uint64
}

func (s *RoundRobinStrategy) Select(nodes []*NodeInfo) *NodeInfo {
	if len(nodes) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&s.counter, 1) % uint64(len(nodes))
	return nodes[idx]
}

// LeastLoadedStrategy selects the node with the lowest load
type LeastLoadedStrategy struct{}

func (s *LeastLoadedStrategy) Select(nodes []*NodeInfo) *NodeInfo {
	if len(nodes) == 0 {
		return nil
	}

	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.Load < selected.Load {
			selected = node
		}
	}
	return selected
}

// PriorityStrategy selects the highest priority node
type PriorityStrategy struct{}

func (s *PriorityStrategy) Select(nodes []*NodeInfo) *NodeInfo {
	if len(nodes) == 0 {
		return nil
	}

	selected := nodes[0]
	for _, node := range nodes[1:] {
		if node.Priority > selected.Priority {
			selected = node
		}
	}
	return selected
}

// RegionAwareStrategy prefers nodes in the same region
type RegionAwareStrategy struct {
	Region string
}

func (s *RegionAwareStrategy) Select(nodes []*NodeInfo) *NodeInfo {
	if len(nodes) == 0 {
		return nil
	}

	var sameRegion, otherRegion []*NodeInfo
	for _, node := range nodes {
		if node.Region == s.Region {
			sameRegion = append(sameRegion, node)
		} else {
			otherRegion = append(otherRegion, node)
		}
	}

	// Prefer same region, fall back to other regions
	if len(sameRegion) > 0 {
		leastLoaded := &LeastLoadedStrategy{}
		return leastLoaded.Select(sameRegion)
	}
	if len(otherRegion) > 0 {
		leastLoaded := &LeastLoadedStrategy{}
		return leastLoaded.Select(otherRegion)
	}
	return nil
}