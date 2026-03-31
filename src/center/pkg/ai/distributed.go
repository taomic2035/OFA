// Package distributed provides distributed inference capabilities
// for multi-GPU and multi-node model inference.
package distributed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// DistributionStrategy defines distribution strategies
type DistributionStrategy string

const (
	StrategyDataParallel   DistributionStrategy = "data_parallel"   // Split input data
	StrategyModelParallel  DistributionStrategy = "model_parallel"  // Split model layers
	StrategyPipeline       DistributionStrategy = "pipeline"        // Pipeline parallelism
	StrategyTensorParallel DistributionStrategy = "tensor_parallel" // Tensor parallelism
)

// NodeRole defines node roles in distributed inference
type NodeRole string

const (
	RoleCoordinator NodeRole = "coordinator" // Coordinates inference
	RoleWorker      NodeRole = "worker"      // Executes inference
	RoleAggregator  NodeRole = "aggregator"  // Aggregates results
)

// InferenceNode represents a node in distributed inference
type InferenceNode struct {
	ID           string            `json:"id"`
	Address      string            `json:"address"`
	Port         int               `json:"port"`
	Role         NodeRole          `json:"role"`
	GPUCount     int               `json:"gpu_count"`
	GPUMemory    int               `json:"gpu_memory_mb"`
	ModelIDs     []string          `json:"model_ids"`
	State        NodeState         `json:"state"`
	Load         float64           `json:"load"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
	Capabilities map[string]string `json:"capabilities"`
}

// NodeState defines node states
type NodeState string

const (
	NodeStateOnline    NodeState = "online"
	NodeStateOffline   NodeState = "offline"
	NodeStateBusy      NodeState = "busy"
	NodeStateRecovering NodeState = "recovering"
)

// DistributedConfig holds distributed inference configuration
type DistributedConfig struct {
	Strategy          DistributionStrategy `json:"strategy"`
	MaxNodes          int                  `json:"max_nodes"`
	MinNodes          int                  `json:"min_nodes"`
	Timeout           time.Duration        `json:"timeout"`
	RetryCount        int                  `json:"retry_count"`
	BatchSize         int                  `json:"batch_size"`
	ReplicationFactor int                  `json:"replication_factor"`
}

// DefaultDistributedConfig returns default configuration
func DefaultDistributedConfig() *DistributedConfig {
	return &DistributedConfig{
		Strategy:          StrategyDataParallel,
		MaxNodes:          10,
		MinNodes:          1,
		Timeout:           30 * time.Second,
		RetryCount:        3,
		BatchSize:         32,
		ReplicationFactor: 2,
	}
}

// DistributedRequest represents a distributed inference request
type DistributedRequest struct {
	ID           string                 `json:"id"`
	ModelID      string                 `json:"model_id"`
	Input        interface{}            `json:"input"`
	Params       map[string]interface{} `json:"params"`
	Strategy     DistributionStrategy   `json:"strategy"`
	BatchSize    int                    `json:"batch_size"`
	Priority     int                    `json:"priority"`
	CreatedAt    time.Time              `json:"created_at"`
}

// DistributedResult represents distributed inference result
type DistributedResult struct {
	RequestID    string                 `json:"request_id"`
	Success      bool                   `json:"success"`
	Output       interface{}            `json:"output"`
	Error        string                 `json:"error,omitempty"`
	Duration     int64                  `json:"duration_ms"`
	NodeCount    int                    `json:"node_count"`
	NodeResults  []NodeResult           `json:"node_results"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NodeResult represents result from a single node
type NodeResult struct {
	NodeID    string      `json:"node_id"`
	Success   bool        `json:"success"`
	Output    interface{} `json:"output"`
	Error     string      `json:"error,omitempty"`
	Duration  int64       `json:"duration_ms"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
}

// DistributedManager manages distributed inference
type DistributedManager struct {
	config *DistributedConfig

	// Node registry
	nodes    sync.Map // map[string]*InferenceNode
	nodeList []*InferenceNode

	// Request handling
	requestQueue  chan *DistributedRequest
	resultChan    chan *DistributedResult

	// Model partitioning
	partitions sync.Map // map[string]*ModelPartition

	// Metrics
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalLatency    int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// ModelPartition represents model partition info
type ModelPartition struct {
	ModelID     string       `json:"model_id"`
	Strategy    DistributionStrategy `json:"strategy"`
	Partitions  []Partition  `json:"partitions"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Partition represents a single partition
type Partition struct {
	ID       string   `json:"id"`
	NodeIDs  []string `json:"node_ids"`
	LayerStart int    `json:"layer_start"`
	LayerEnd   int    `json:"layer_end"`
	Size      int64    `json:"size"`
}

// NewDistributedManager creates a new distributed inference manager
func NewDistributedManager(config *DistributedConfig) (*DistributedManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &DistributedManager{
		config:       config,
		requestQueue: make(chan *DistributedRequest, 1000),
		resultChan:   make(chan *DistributedResult, 1000),
		ctx:          ctx,
		cancel:       cancel,
	}

	return manager, nil
}

// Start starts the distributed manager
func (dm *DistributedManager) Start() {
	go dm.requestProcessor()
	go dm.nodeMonitor()
	go dm.loadBalancer()
}

// Stop stops the distributed manager
func (dm *DistributedManager) Stop() {
	dm.cancel()
}

// RegisterNode registers a new inference node
func (dm *DistributedManager) RegisterNode(node *InferenceNode) error {
	if node.ID == "" {
		return errors.New("node ID required")
	}

	node.LastHeartbeat = time.Now()
	node.State = NodeStateOnline

	dm.nodes.Store(node.ID, node)

	dm.mu.Lock()
	dm.nodeList = append(dm.nodeList, node)
	dm.mu.Unlock()

	log.Printf("Node registered: %s (GPUs: %d)", node.ID, node.GPUCount)

	return nil
}

// UnregisterNode unregisters a node
func (dm *DistributedManager) UnregisterNode(nodeID string) {
	dm.nodes.Delete(nodeID)

	dm.mu.Lock()
	for i, n := range dm.nodeList {
		if n.ID == nodeID {
			dm.nodeList = append(dm.nodeList[:i], dm.nodeList[i+1:]...)
			break
		}
	}
	dm.mu.Unlock()

	log.Printf("Node unregistered: %s", nodeID)
}

// GetNode returns node info
func (dm *DistributedManager) GetNode(nodeID string) (*InferenceNode, bool) {
	if v, ok := dm.nodes.Load(nodeID); ok {
		return v.(*InferenceNode), true
	}
	return nil, false
}

// ListNodes returns all registered nodes
func (dm *DistributedManager) ListNodes() []*InferenceNode {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.nodeList
}

// GetAvailableNodes returns available nodes for inference
func (dm *DistributedManager) GetAvailableNodes(modelID string) []*InferenceNode {
	var available []*InferenceNode

	dm.nodes.Range(func(key, value interface{}) bool {
		node := value.(*InferenceNode)
		if node.State == NodeStateOnline && node.Load < 0.8 {
			// Check if node has the model
			for _, m := range node.ModelIDs {
				if m == modelID {
					available = append(available, node)
					break
				}
			}
		}
		return true
	})

	return available
}

// Infer performs distributed inference
func (dm *DistributedManager) Infer(req *DistributedRequest) (*DistributedResult, error) {
	req.CreatedAt = time.Now()

	// Select strategy if not specified
	if req.Strategy == "" {
		req.Strategy = dm.config.Strategy
	}

	// Queue request
	select {
	case dm.requestQueue <- req:
	default:
		return nil, errors.New("request queue full")
	}

	// Wait for result (simplified)
	time.Sleep(10 * time.Millisecond)

	return &DistributedResult{
		RequestID: req.ID,
		Success:   true,
		Output:    "distributed_inference_result",
		NodeCount: 1,
	}, nil
}

// requestProcessor processes inference requests
func (dm *DistributedManager) requestProcessor() {
	for {
		select {
		case <-dm.ctx.Done():
			return
		case req := <-dm.requestQueue:
			go dm.processRequest(req)
		}
	}
}

// processRequest processes a single request
func (dm *DistributedManager) processRequest(req *DistributedRequest) {
	startTime := time.Now()
	result := &DistributedResult{
		RequestID: req.ID,
	}

	// Get available nodes
	nodes := dm.GetAvailableNodes(req.ModelID)
	if len(nodes) < dm.config.MinNodes {
		result.Success = false
		result.Error = fmt.Sprintf("insufficient nodes: %d < %d", len(nodes), dm.config.MinNodes)
		dm.resultChan <- result
		return
	}

	// Limit nodes
	if len(nodes) > dm.config.MaxNodes {
		nodes = nodes[:dm.config.MaxNodes]
	}

	// Execute based on strategy
	var outputs []NodeResult
	var err error

	switch req.Strategy {
	case StrategyDataParallel:
		outputs, err = dm.executeDataParallel(req, nodes)
	case StrategyModelParallel:
		outputs, err = dm.executeModelParallel(req, nodes)
	case StrategyPipeline:
		outputs, err = dm.executePipeline(req, nodes)
	default:
		outputs, err = dm.executeDataParallel(req, nodes)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	result.NodeCount = len(nodes)
	result.NodeResults = outputs

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		dm.mu.Lock()
		dm.failedRequests++
		dm.mu.Unlock()
	} else {
		result.Success = true
		result.Output = dm.aggregateOutputs(outputs)
		dm.mu.Lock()
		dm.successRequests++
		dm.totalLatency += result.Duration
		dm.mu.Unlock()
	}

	dm.mu.Lock()
	dm.totalRequests++
	dm.mu.Unlock()

	dm.resultChan <- result
}

// executeDataParallel executes data parallel inference
func (dm *DistributedManager) executeDataParallel(req *DistributedRequest, nodes []*InferenceNode) ([]NodeResult, error) {
	// Split input data across nodes
	inputs := dm.splitInput(req.Input, len(nodes))
	if len(inputs) == 0 {
		inputs = []interface{}{req.Input}
	}

	results := make([]NodeResult, len(nodes))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *InferenceNode) {
			defer wg.Done()

			start := time.Now()
			result := NodeResult{
				NodeID:    n.ID,
				StartTime: start,
			}

			// Execute inference on node
			output, err := dm.executeOnNode(n, req.ModelID, inputs[idx%len(inputs)], req.Params)
			result.EndTime = time.Now()
			result.Duration = time.Since(start).Milliseconds()

			if err != nil {
				result.Success = false
				result.Error = err.Error()
			} else {
				result.Success = true
				result.Output = output
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, node)
	}

	wg.Wait()

	// Check for failures
	failCount := 0
	for _, r := range results {
		if !r.Success {
			failCount++
		}
	}

	if failCount > len(results)/2 {
		return results, errors.New("majority of nodes failed")
	}

	return results, nil
}

// executeModelParallel executes model parallel inference
func (dm *DistributedManager) executeModelParallel(req *DistributedRequest, nodes []*InferenceNode) ([]NodeResult, error) {
	// Get model partition
	partition, ok := dm.partitions.Load(req.ModelID)
	if !ok {
		// Create default partition
		partition = dm.createModelPartition(req.ModelID, len(nodes))
	}

	p := partition.(*ModelPartition)
	results := make([]NodeResult, len(nodes))

	// Execute pipeline through partitions
	currentInput := req.Input
	for i, part := range p.Partitions {
		if i >= len(nodes) {
			break
		}

		node := nodes[i]
		start := time.Now()

		output, err := dm.executeOnNode(node, req.ModelID, currentInput, req.Params)

		results[i] = NodeResult{
			NodeID:    node.ID,
			StartTime: start,
			EndTime:   time.Now(),
			Duration:  time.Since(start).Milliseconds(),
			Success:   err == nil,
			Output:    output,
			Error:     errorToString(err),
		}

		if err != nil {
			return results, err
		}

		currentInput = output
	}

	return results, nil
}

// executePipeline executes pipeline parallel inference
func (dm *DistributedManager) executePipeline(req *DistributedRequest, nodes []*InferenceNode) ([]NodeResult, error) {
	// Similar to model parallel but with pipelining
	return dm.executeModelParallel(req, nodes)
}

// splitInput splits input data for parallel processing
func (dm *DistributedManager) splitInput(input interface{}, parts int) []interface{} {
	// Handle different input types
	switch v := input.(type) {
	case []interface{}:
		if len(v) < parts {
			return v
		}
		result := make([]interface{}, parts)
		chunkSize := len(v) / parts
		for i := 0; i < parts; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if i == parts-1 {
				end = len(v)
			}
			result[i] = v[start:end]
		}
		return result
	case string:
		// For text, split into chunks
		if len(v) < parts*100 {
			return []interface{}{v}
		}
		result := make([]interface{}, parts)
		chunkSize := len(v) / parts
		for i := 0; i < parts; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if i == parts-1 {
				end = len(v)
			}
			result[i] = v[start:end]
		}
		return result
	default:
		// For other types, return as single item
		return []interface{}{input}
	}
}

// executeOnNode executes inference on a specific node
func (dm *DistributedManager) executeOnNode(node *InferenceNode, modelID string, input interface{}, params map[string]interface{}) (interface{}, error) {
	// Placeholder for actual node communication
	// In production, would use gRPC or HTTP

	// Simulate inference
	result := map[string]interface{}{
		"node_id":  node.ID,
		"model_id": modelID,
		"input":    input,
		"output":   "node_inference_result",
	}

	return result, nil
}

// aggregateOutputs aggregates outputs from multiple nodes
func (dm *DistributedManager) aggregateOutputs(results []NodeResult) interface{} {
	if len(results) == 0 {
		return nil
	}

	if len(results) == 1 {
		return results[0].Output
	}

	// Collect successful outputs
	var outputs []interface{}
	for _, r := range results {
		if r.Success {
			outputs = append(outputs, r.Output)
		}
	}

	// Merge outputs based on type
	if len(outputs) == 0 {
		return nil
	}

	// For maps, merge them
	if m, ok := outputs[0].(map[string]interface{}); ok {
		merged := make(map[string]interface{})
		for k, v := range m {
			merged[k] = v
		}
		for i := 1; i < len(outputs); i++ {
			if m2, ok := outputs[i].(map[string]interface{}); ok {
				for k, v := range m2 {
					if _, exists := merged[k]; !exists {
						merged[k] = v
					} else if arr, ok := merged[k].([]interface{}); ok {
						if arr2, ok := v.([]interface{}); ok {
							merged[k] = append(arr, arr2...)
						}
					}
				}
			}
		}
		return merged
	}

	// For arrays, concatenate
	if arr, ok := outputs[0].([]interface{}); ok {
		result := make([]interface{}, len(arr))
		copy(result, arr)
		for i := 1; i < len(outputs); i++ {
			if arr2, ok := outputs[i].([]interface{}); ok {
				result = append(result, arr2...)
			}
		}
		return result
	}

	// Default: return first output
	return outputs[0]
}

// createModelPartition creates a model partition
func (dm *DistributedManager) createModelPartition(modelID string, nodeCount int) *ModelPartition {
	partitions := make([]Partition, nodeCount)

	for i := 0; i < nodeCount; i++ {
		partitions[i] = Partition{
			ID:         fmt.Sprintf("partition-%d", i),
			LayerStart: i * 10,
			LayerEnd:   (i + 1) * 10,
			Size:       1000000,
		}
	}

	mp := &ModelPartition{
		ModelID:    modelID,
		Strategy:   dm.config.Strategy,
		Partitions: partitions,
		CreatedAt:  time.Now(),
	}

	dm.partitions.Store(modelID, mp)

	return mp
}

// nodeMonitor monitors node health
func (dm *DistributedManager) nodeMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.checkNodes()
		}
	}
}

// checkNodes checks node health
func (dm *DistributedManager) checkNodes() {
	dm.nodes.Range(func(key, value interface{}) bool {
		node := value.(*InferenceNode)

		// Check heartbeat timeout
		if time.Since(node.LastHeartbeat) > 30*time.Second {
			node.State = NodeStateOffline
			log.Printf("Node %s marked offline", node.ID)
		}

		return true
	})
}

// loadBalancer balances load across nodes
func (dm *DistributedManager) loadBalancer() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.balanceLoad()
		}
	}
}

// balanceLoad rebalances load across nodes
func (dm *DistributedManager) balanceLoad() {
	// Placeholder for load balancing logic
}

// GetStats returns distributed inference statistics
func (dm *DistributedManager) GetStats() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var avgLatency int64
	if dm.successRequests > 0 {
		avgLatency = dm.totalLatency / dm.successRequests
	}

	return map[string]interface{}{
		"total_requests":    dm.totalRequests,
		"success_requests":  dm.successRequests,
		"failed_requests":   dm.failedRequests,
		"avg_latency_ms":    avgLatency,
		"node_count":        len(dm.nodeList),
		"strategy":          dm.config.Strategy,
	}
}

// errorToString converts error to string
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}