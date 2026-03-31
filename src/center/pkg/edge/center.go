// Package edge provides edge computing capabilities for OFA
// including edge Center deployment, local task processing,
// edge-cloud collaboration, and data preprocessing.
package edge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EdgeNodeType defines the type of edge node
type EdgeNodeType string

const (
	EdgeNodeGateway   EdgeNodeType = "gateway"   // Gateway edge node
	EdgeNodeCompute   EdgeNodeType = "compute"   // Compute edge node
	EdgeNodeStorage   EdgeNodeType = "storage"   // Storage edge node
	EdgeNodeInference EdgeNodeType = "inference" // AI inference edge node
)

// EdgeNodeState defines edge node state
type EdgeNodeState string

const (
	EdgeStateOnline      EdgeNodeState = "online"
	EdgeStateOffline     EdgeNodeState = "offline"
	EdgeStateSyncing     EdgeNodeState = "syncing"
	EdgeStateDegraded    EdgeNodeState = "degraded"
	EdgeStateMaintenance EdgeNodeState = "maintenance"
)

// EdgeNode represents an edge computing node
type EdgeNode struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          EdgeNodeType      `json:"type"`
	State         EdgeNodeState     `json:"state"`
	Address       string            `json:"address"`
	Port          int               `json:"port"`
	Region        string            `json:"region"`
	Zone          string            `json:"zone"`
	ParentCenter  string            `json:"parent_center"`  // Parent Center ID
	Capabilities  []EdgeCapability  `json:"capabilities"`
	Resources     EdgeResources     `json:"resources"`
	Config        EdgeConfig        `json:"config"`
	LastSync      time.Time         `json:"last_sync"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	StartedAt     time.Time         `json:"started_at"`
	Metadata      map[string]string `json:"metadata"`
}

// EdgeCapability defines an edge capability
type EdgeCapability struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`    // compute, storage, inference, network
	Operations  []string `json:"operations"`
	Priority    int      `json:"priority"`
	MaxLoad     int      `json:"max_load"`
	CurrentLoad int      `json:"current_load"`
}

// EdgeResources defines edge node resources
type EdgeResources struct {
	CPU          int     `json:"cpu"`           // Number of cores
	MemoryMB     int     `json:"memory_mb"`     // Memory in MB
	StorageGB    int     `json:"storage_gb"`    // Storage in GB
	GPU          int     `json:"gpu"`           // Number of GPUs
	GPUModel     string  `json:"gpu_model"`     // GPU model
	GPUMemoryMB  int     `json:"gpu_memory_mb"` // GPU memory in MB
	BandwidthMBps int    `json:"bandwidth_mbps"` // Network bandwidth
	CPUUsage     float64 `json:"cpu_usage"`     // Current CPU usage %
	MemoryUsage  float64 `json:"memory_usage"`  // Current memory usage %
	GPUUsage     float64 `json:"gpu_usage"`     // Current GPU usage %
}

// EdgeConfig defines edge node configuration
type EdgeConfig struct {
	SyncInterval      time.Duration `json:"sync_interval"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	OfflineMode       bool          `json:"offline_mode"`
	MaxOfflineTime    time.Duration `json:"max_offline_time"`
	CacheSize         int           `json:"cache_size"`
	BatchSize         int           `json:"batch_size"`
	PriorityThreshold int           `json:"priority_threshold"`
	PreprocessingEnabled bool        `json:"preprocessing_enabled"`
}

// DefaultEdgeConfig returns default edge configuration
func DefaultEdgeConfig() EdgeConfig {
	return EdgeConfig{
		SyncInterval:       30 * time.Second,
		HeartbeatInterval:  10 * time.Second,
		OfflineMode:        true,
		MaxOfflineTime:     1 * time.Hour,
		CacheSize:          1000,
		BatchSize:          100,
		PriorityThreshold:  50,
		PreprocessingEnabled: true,
	}
}

// EdgeConfig holds edge center configuration
type Config struct {
	NodeID       string
	NodeName     string
	NodeType     EdgeNodeType
	Address      string
	Port         int
	Region       string
	Zone         string
	ParentCenter string
	DataDir      string
	EdgeConfig   EdgeConfig
}

// EdgeCenter manages edge computing operations
type EdgeCenter struct {
	config   Config
	node     *EdgeNode
	client   *http.Client

	// Task management
	taskQueue   chan EdgeTask
	taskResults chan EdgeTaskResult

	// Local cache
	taskCache    sync.Map // map[string]*EdgeTask
	resultCache  sync.Map // map[string]*EdgeTaskResult
	modelCache   sync.Map // map[string]*CachedModel

	// Sync state
	syncState    EdgeSyncState
	pendingSync  sync.Map // map[string]interface{}

	// Preprocessing
	preprocessor *DataPreprocessor

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// EdgeSyncState represents sync state with parent Center
type EdgeSyncState struct {
	LastSync      time.Time
	SyncInProgress bool
	PendingTasks  int
	PendingResults int
	Connected     bool
}

// EdgeTask represents a task for edge processing
type EdgeTask struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`    // inference, compute, preprocess, sync
	SkillID      string                 `json:"skill_id"`
	Operation    string                 `json:"operation"`
	Params       map[string]interface{} `json:"params"`
	Priority     int                    `json:"priority"`
	Source       string                 `json:"source"`  // local or cloud
	CreatedAt    time.Time              `json:"created_at"`
	ProcessAfter time.Time              `json:"process_after,omitempty"` // For delayed processing
	Timeout      int                    `json:"timeout"`
	RequiresGPU  bool                   `json:"requires_gpu"`
	ModelID      string                 `json:"model_id,omitempty"`
}

// EdgeTaskResult represents task execution result
type EdgeTaskResult struct {
	TaskID     string                 `json:"task_id"`
	Success    bool                   `json:"success"`
	Data       interface{}            `json:"data"`
	Error      string                 `json:"error,omitempty"`
	ProcessedAt time.Time             `json:"processed_at"`
	Duration   int64                  `json:"duration_ms"`
	ProcessedBy string                `json:"processed_by"` // local or cloud
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CachedModel represents a cached AI model
type CachedModel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	SizeMB      int       `json:"size_mb"`
	CachedAt    time.Time `json:"cached_at"`
	LastUsed    time.Time `json:"last_used"`
	UseCount    int       `json:"use_count"`
	DownloadURL string    `json:"download_url"`
	Checksum    string    `json:"checksum"`
}

// NewEdgeCenter creates a new edge center
func NewEdgeCenter(config Config) (*EdgeCenter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	edgeCenter := &EdgeCenter{
		config:      config,
		client:      &http.Client{Timeout: 30 * time.Second},
		taskQueue:   make(chan EdgeTask, config.EdgeConfig.CacheSize),
		taskResults: make(chan EdgeTaskResult, config.EdgeConfig.CacheSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize node
	edgeCenter.node = &EdgeNode{
		ID:           config.NodeID,
		Name:         config.NodeName,
		Type:         config.NodeType,
		State:        EdgeStateOnline,
		Address:      config.Address,
		Port:         config.Port,
		Region:       config.Region,
		Zone:         config.Zone,
		ParentCenter: config.ParentCenter,
		Config:       config.EdgeConfig,
		StartedAt:    time.Now(),
		Metadata:     make(map[string]string),
	}

	// Initialize preprocessor
	edgeCenter.preprocessor = NewDataPreprocessor(config.DataDir)

	// Load cached data
	if err := edgeCenter.loadCache(); err != nil {
		log.Printf("Warning: failed to load cache: %v", err)
	}

	return edgeCenter, nil
}

// Start starts the edge center
func (ec *EdgeCenter) Start() error {
	log.Printf("Starting Edge Center: %s", ec.config.NodeID)

	// Start task processor
	go ec.taskProcessor()

	// Start heartbeat to parent
	go ec.heartbeatLoop()

	// Start sync with parent
	go ec.syncLoop()

	// Start resource monitor
	go ec.resourceMonitor()

	// Start cache cleaner
	go ec.cacheCleaner()

	return nil
}

// Stop stops the edge center
func (ec *EdgeCenter) Stop() error {
	log.Println("Stopping Edge Center...")
	ec.cancel()
	ec.saveCache()
	return nil
}

// taskProcessor processes local tasks
func (ec *EdgeCenter) taskProcessor() {
	for {
		select {
		case <-ec.ctx.Done():
			return
		case task := <-ec.taskQueue:
			go ec.processTask(task)
		}
	}
}

// processTask processes a single task
func (ec *EdgeCenter) processTask(task EdgeTask) {
	startTime := time.Now()
	result := EdgeTaskResult{
		TaskID:      task.ID,
		ProcessedAt: time.Now(),
		ProcessedBy: "local",
	}

	// Check if should process locally or forward to cloud
	if !ec.shouldProcessLocally(task) {
		// Forward to parent Center
		forwardResult, err := ec.forwardToCloud(task)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Cloud forwarding failed: %v", err)
			result.ProcessedBy = "cloud"
		} else {
			result = *forwardResult
		}
		ec.taskResults <- result
		return
	}

	// Process locally
	var data interface{}
	var err error

	switch task.Type {
	case "inference":
		data, err = ec.processInference(task)
	case "compute":
		data, err = ec.processCompute(task)
	case "preprocess":
		data, err = ec.preprocessor.Process(task.Params)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}

	result.Duration = time.Since(startTime).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
		result.Data = data
	}

	// Cache result
	ec.resultCache.Store(task.ID, &result)

	// Send result
	ec.taskResults <- result
}

// shouldProcessLocally determines if task should be processed locally
func (ec *EdgeCenter) shouldProcessLocally(task EdgeTask) bool {
	// Check if requires GPU and we have GPU
	if task.RequiresGPU && ec.node.Resources.GPU == 0 {
		return false
	}

	// Check if model is available locally
	if task.ModelID != "" {
		if _, ok := ec.modelCache.Load(task.ModelID); !ok {
			return false
		}
	}

	// Check resources
	if ec.node.Resources.CPUUsage > 80 || ec.node.Resources.MemoryUsage > 80 {
		if task.Priority < ec.config.EdgeConfig.PriorityThreshold {
			return false
		}
	}

	// Check connectivity
	if !ec.syncState.Connected && !ec.config.EdgeConfig.OfflineMode {
		return true // Must process locally
	}

	return true
}

// processInference processes AI inference task
func (ec *EdgeCenter) processInference(task EdgeTask) (interface{}, error) {
	// Get model
	model, ok := ec.modelCache.Load(task.ModelID)
	if !ok {
		return nil, fmt.Errorf("model not found: %s", task.ModelID)
	}

	cachedModel := model.(*CachedModel)

	// Update usage
	cachedModel.LastUsed = time.Now()
	cachedModel.UseCount++

	// Execute inference (placeholder - would call actual inference engine)
	result := map[string]interface{}{
		"model_id":  task.ModelID,
		"operation": task.Operation,
		"params":    task.Params,
		"result":    "inference_result_placeholder",
	}

	return result, nil
}

// processCompute processes compute task
func (ec *EdgeCenter) processCompute(task EdgeTask) (interface{}, error) {
	// Placeholder for compute task processing
	return map[string]interface{}{
		"task_id":  task.ID,
		"skill_id": task.SkillID,
		"status":   "completed",
	}, nil
}

// forwardToCloud forwards task to parent Center
func (ec *EdgeCenter) forwardToCloud(task EdgeTask) (*EdgeTaskResult, error) {
	// Placeholder for cloud forwarding
	return &EdgeTaskResult{
		TaskID:      task.ID,
		Success:     true,
		Data:        "forwarded_to_cloud",
		ProcessedBy: "cloud",
	}, nil
}

// heartbeatLoop sends heartbeats to parent Center
func (ec *EdgeCenter) heartbeatLoop() {
	ticker := time.NewTicker(ec.config.EdgeConfig.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ticker.C:
			ec.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends heartbeat to parent Center
func (ec *EdgeCenter) sendHeartbeat() {
	ec.mu.Lock()
	ec.node.LastHeartbeat = time.Now()
	ec.mu.Unlock()

	// Placeholder for actual heartbeat sending
}

// syncLoop synchronizes with parent Center
func (ec *EdgeCenter) syncLoop() {
	ticker := time.NewTicker(ec.config.EdgeConfig.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ticker.C:
			ec.syncWithParent()
		}
	}
}

// syncWithParent synchronizes data with parent Center
func (ec *EdgeCenter) syncWithParent() {
	ec.mu.Lock()
	if ec.syncState.SyncInProgress {
		ec.mu.Unlock()
		return
	}
	ec.syncState.SyncInProgress = true
	ec.mu.Unlock()

	defer func() {
		ec.mu.Lock()
		ec.syncState.SyncInProgress = false
		ec.syncState.LastSync = time.Now()
		ec.mu.Unlock()
	}()

	// Placeholder for actual sync
}

// resourceMonitor monitors node resources
func (ec *EdgeCenter) resourceMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ticker.C:
			ec.updateResources()
		}
	}
}

// updateResources updates resource metrics
func (ec *EdgeCenter) updateResources() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Placeholder for actual resource monitoring
	// In production, would query actual system resources
}

// cacheCleaner periodically cleans old cache entries
func (ec *EdgeCenter) cacheCleaner() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ticker.C:
			ec.cleanCache()
		}
	}
}

// cleanCache removes old cache entries
func (ec *EdgeCenter) cleanCache() {
	maxAge := ec.config.EdgeConfig.MaxOfflineTime

	// Clean result cache
	ec.resultCache.Range(func(key, value interface{}) bool {
		result := value.(*EdgeTaskResult)
		if time.Since(result.ProcessedAt) > maxAge {
			ec.resultCache.Delete(key)
		}
		return true
	})
}

// loadCache loads cached data from disk
func (ec *EdgeCenter) loadCache() error {
	if ec.config.DataDir == "" {
		return nil
	}

	cachePath := filepath.Join(ec.config.DataDir, "edge_cache.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return err
	}

	var cache struct {
		Tasks   map[string]*EdgeTask       `json:"tasks"`
		Results map[string]*EdgeTaskResult `json:"results"`
		Models  map[string]*CachedModel    `json:"models"`
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}

	for k, v := range cache.Tasks {
		ec.taskCache.Store(k, v)
	}
	for k, v := range cache.Results {
		ec.resultCache.Store(k, v)
	}
	for k, v := range cache.Models {
		ec.modelCache.Store(k, v)
	}

	return nil
}

// saveCache saves cache to disk
func (ec *EdgeCenter) saveCache() error {
	if ec.config.DataDir == "" {
		return nil
	}

	os.MkdirAll(ec.config.DataDir, 0755)

	cache := struct {
		Tasks   map[string]*EdgeTask       `json:"tasks"`
		Results map[string]*EdgeTaskResult `json:"results"`
		Models  map[string]*CachedModel    `json:"models"`
	}{
		Tasks:   make(map[string]*EdgeTask),
		Results: make(map[string]*EdgeTaskResult),
		Models:  make(map[string]*CachedModel),
	}

	ec.taskCache.Range(func(key, value interface{}) bool {
		cache.Tasks[key.(string)] = value.(*EdgeTask)
		return true
	})
	ec.resultCache.Range(func(key, value interface{}) bool {
		cache.Results[key.(string)] = value.(*EdgeTaskResult)
		return true
	})
	ec.modelCache.Range(func(key, value interface{}) bool {
		cache.Models[key.(string)] = value.(*CachedModel)
		return true
	})

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	cachePath := filepath.Join(ec.config.DataDir, "edge_cache.json")
	return os.WriteFile(cachePath, data, 0644)
}

// SubmitTask submits a task for processing
func (ec *EdgeCenter) SubmitTask(task EdgeTask) error {
	task.CreatedAt = time.Now()

	// Cache task
	ec.taskCache.Store(task.ID, &task)

	// Queue for processing
	select {
	case ec.taskQueue <- task:
		return nil
	default:
		return errors.New("task queue full")
	}
}

// GetTaskResult gets task result
func (ec *EdgeCenter) GetTaskResult(taskID string) (*EdgeTaskResult, bool) {
	if v, ok := ec.resultCache.Load(taskID); ok {
		return v.(*EdgeTaskResult), true
	}
	return nil, false
}

// CacheModel caches a model locally
func (ec *EdgeCenter) CacheModel(model *CachedModel) error {
	ec.modelCache.Store(model.ID, model)

	// Download model if URL provided
	if model.DownloadURL != "" {
		go ec.downloadModel(model)
	}

	return nil
}

// downloadModel downloads model from URL
func (ec *EdgeCenter) downloadModel(model *CachedModel) {
	// Placeholder for actual model download
	log.Printf("Downloading model: %s", model.ID)
}

// GetNodeInfo returns current node info
func (ec *EdgeCenter) GetNodeInfo() *EdgeNode {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.node
}

// GetSyncState returns current sync state
func (ec *EdgeCenter) GetSyncState() EdgeSyncState {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.syncState
}

// GetStats returns edge center statistics
func (ec *EdgeCenter) GetStats() map[string]interface{} {
	taskCount := 0
	resultCount := 0
	modelCount := 0

	ec.taskCache.Range(func(key, value interface{}) bool {
		taskCount++
		return true
	})
	ec.resultCache.Range(func(key, value interface{}) bool {
		resultCount++
		return true
	})
	ec.modelCache.Range(func(key, value interface{}) bool {
		modelCount++
		return true
	})

	return map[string]interface{}{
		"node_id":      ec.config.NodeID,
		"node_name":    ec.config.NodeName,
		"node_type":    ec.config.NodeType,
		"state":        ec.node.State,
		"cached_tasks": taskCount,
		"cached_results": resultCount,
		"cached_models": modelCount,
		"sync_state":   ec.syncState,
		"resources":    ec.node.Resources,
	}
}