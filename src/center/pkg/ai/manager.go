// Package ai provides AI capabilities for OFA including
// model inference, distributed inference, model version management,
// and GPU scheduling.
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ModelType defines AI model types
type ModelType string

const (
	ModelTypeLLM      ModelType = "llm"       // Large Language Model
	ModelTypeCV       ModelType = "cv"        // Computer Vision
	ModelTypeASR      ModelType = "asr"       // Automatic Speech Recognition
	ModelTypeTTS      ModelType = "tts"       // Text to Speech
	ModelTypeEmbedding ModelType = "embedding" // Embedding Model
	ModelTypeDiffusion ModelType = "diffusion" // Diffusion Model
)

// ModelState defines model states
type ModelState string

const (
	ModelStateLoading   ModelState = "loading"
	ModelStateReady     ModelState = "ready"
	ModelStateBusy      ModelState = "busy"
	ModelStateError     ModelState = "error"
	ModelStateUnloading ModelState = "unloading"
)

// ModelFormat defines model file formats
type ModelFormat string

const (
	FormatONNX    ModelFormat = "onnx"
	FormatPyTorch ModelFormat = "pytorch"
	FormatTensorRT ModelFormat = "tensorrt"
	FormatGGML    ModelFormat = "ggml"
	FormatSafetensors ModelFormat = "safetensors"
)

// ModelInfo represents model information
type ModelInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         ModelType         `json:"type"`
	Format       ModelFormat       `json:"format"`
	State        ModelState        `json:"state"`
	Path         string            `json:"path"`
	SizeMB       int               `json:"size_mb"`
	GPUMemoryMB  int               `json:"gpu_memory_mb"`
	MaxTokens    int               `json:"max_tokens"`
	InputTypes   []string          `json:"input_types"`
	OutputTypes  []string          `json:"output_types"`
	Operations   []string          `json:"operations"`
	LoadedAt     time.Time         `json:"loaded_at"`
	LastUsed     time.Time         `json:"last_used"`
	UseCount     int               `json:"use_count"`
	AvgLatency   int64             `json:"avg_latency_ms"`
	Device       string            `json:"device"`    // cpu, cuda:0, cuda:1, etc.
	Metadata     map[string]string `json:"metadata"`
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	ID          string                 `json:"id"`
	ModelID     string                 `json:"model_id"`
	Operation   string                 `json:"operation"`
	Input       interface{}            `json:"input"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Timeout     int                    `json:"timeout"`
	CreatedAt   time.Time              `json:"created_at"`
}

// InferenceResult represents inference result
type InferenceResult struct {
	RequestID   string                 `json:"request_id"`
	ModelID     string                 `json:"model_id"`
	Success     bool                   `json:"success"`
	Output      interface{}            `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    int64                  `json:"duration_ms"`
	TokensIn    int                    `json:"tokens_in"`
	TokensOut   int                    `json:"tokens_out"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	MemoryMB    int     `json:"memory_mb"`
	MemoryUsed  int     `json:"memory_used"`
	MemoryFree  int     `json:"memory_free"`
	Utilization float64 `json:"utilization"`
	Temperature float64 `json:"temperature"`
}

// AIConfig holds AI module configuration
type AIConfig struct {
	ModelDir        string
	CacheDir        string
	MaxGPUModels    int
	MaxCPUModels    int
	DefaultDevice   string
	EnableGPU       bool
	PreloadModels   []string
}

// DefaultAIConfig returns default configuration
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		MaxGPUModels:  2,
		MaxCPUModels:  5,
		DefaultDevice: "cpu",
		EnableGPU:     true,
	}
}

// AIManager manages AI models and inference
type AIManager struct {
	config *AIConfig

	// Model registry
	models    sync.Map // map[string]*ModelInfo
	instances sync.Map // map[string]*ModelInstance

	// GPU management
	gpus []GPUInfo

	// Inference queue
	requestQueue  chan *InferenceRequest
	resultChan    chan *InferenceResult

	// Metrics
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalLatency    int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// ModelInstance represents a loaded model instance
type ModelInstance struct {
	Info     *ModelInfo
	Handler  ModelHandler
	Device   string
	Loaded   bool
}

// ModelHandler handles model inference
type ModelHandler interface {
	Load(modelPath string, device string) error
	Unload() error
	Infer(input interface{}, params map[string]interface{}) (interface{}, error)
	GetInfo() *ModelInfo
}

// NewAIManager creates a new AI manager
func NewAIManager(config *AIConfig) (*AIManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &AIManager{
		config:       config,
		requestQueue: make(chan *InferenceRequest, 1000),
		resultChan:   make(chan *InferenceResult, 1000),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Detect GPUs
	manager.detectGPUs()

	// Load model registry
	manager.loadModelRegistry()

	// Preload specified models
	for _, modelID := range config.PreloadModels {
		if err := manager.LoadModel(modelID, ""); err != nil {
			log.Printf("Failed to preload model %s: %v", modelID, err)
		}
	}

	return manager, nil
}

// Start starts the AI manager
func (am *AIManager) Start() {
	// Start inference processor
	go am.inferenceProcessor()

	// Start GPU monitor
	go am.gpuMonitor()

	// Start model auto-loader
	go am.modelManager()
}

// Stop stops the AI manager
func (am *AIManager) Stop() {
	am.cancel()

	// Unload all models
	am.instances.Range(func(key, value interface{}) bool {
		instance := value.(*ModelInstance)
		if instance.Loaded && instance.Handler != nil {
			instance.Handler.Unload()
		}
		return true
	})
}

// detectGPUs detects available GPUs
func (am *AIManager) detectGPUs() {
	// Placeholder for GPU detection
	// In production, would use CUDA/ROCm libraries
	am.gpus = []GPUInfo{}

	if am.config.EnableGPU {
		// Simulate GPU detection
		am.gpus = append(am.gpus, GPUInfo{
			ID:         0,
			Name:       "GPU-0",
			MemoryMB:   8192,
			MemoryUsed: 0,
			MemoryFree: 8192,
		})
	}
}

// loadModelRegistry loads model registry from disk
func (am *AIManager) loadModelRegistry() error {
	if am.config.ModelDir == "" {
		return nil
	}

	registryPath := filepath.Join(am.config.ModelDir, "models.json")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil // Registry doesn't exist yet
	}

	var models []*ModelInfo
	if err := json.Unmarshal(data, &models); err != nil {
		return err
	}

	for _, model := range models {
		am.models.Store(model.ID, model)
	}

	return nil
}

// saveModelRegistry saves model registry to disk
func (am *AIManager) saveModelRegistry() error {
	if am.config.ModelDir == "" {
		return nil
	}

	os.MkdirAll(am.config.ModelDir, 0755)

	var models []*ModelInfo
	am.models.Range(func(key, value interface{}) bool {
		models = append(models, value.(*ModelInfo))
		return true
	})

	data, err := json.Marshal(models)
	if err != nil {
		return err
	}

	registryPath := filepath.Join(am.config.ModelDir, "models.json")
	return os.WriteFile(registryPath, data, 0644)
}

// RegisterModel registers a new model
func (am *AIManager) RegisterModel(info *ModelInfo) error {
	if info.ID == "" {
		return errors.New("model ID required")
	}

	info.State = ModelStateReady
	info.LoadedAt = time.Time{}

	am.models.Store(info.ID, info)
	am.saveModelRegistry()

	return nil
}

// LoadModel loads a model into memory
func (am *AIManager) LoadModel(modelID, device string) error {
	info, ok := am.getModelInfo(modelID)
	if !ok {
		return fmt.Errorf("model not found: %s", modelID)
	}

	// Check if already loaded
	if instance, ok := am.instances.Load(modelID); ok {
		if instance.(*ModelInstance).Loaded {
			return nil // Already loaded
		}
	}

	// Determine device
	if device == "" {
		device = am.selectDevice(info)
	}

	// Check GPU memory if using GPU
	if device != "cpu" && info.GPUMemoryMB > 0 {
		if !am.hasGPUMemory(info.GPUMemoryMB, device) {
			// Fallback to CPU or return error
			device = "cpu"
		}
	}

	// Create model instance
	instance := &ModelInstance{
		Info:   info,
		Device: device,
	}

	// Create model handler
	handler, err := am.createModelHandler(info)
	if err != nil {
		return err
	}

	// Load model
	instance.Handler = handler
	if err := handler.Load(info.Path, device); err != nil {
		return err
	}

	instance.Loaded = true
	info.State = ModelStateReady
	info.LoadedAt = time.Now()
	info.Device = device

	am.instances.Store(modelID, instance)

	// Update GPU memory usage
	if device != "cpu" {
		am.updateGPUMemory(device, info.GPUMemoryMB)
	}

	log.Printf("Model %s loaded on %s", modelID, device)

	return nil
}

// UnloadModel unloads a model from memory
func (am *AIManager) UnloadModel(modelID string) error {
	instance, ok := am.instances.Load(modelID)
	if !ok {
		return nil // Not loaded
	}

	ins := instance.(*ModelInstance)
	if ins.Handler != nil {
		if err := ins.Handler.Unload(); err != nil {
			return err
		}
	}

	// Update GPU memory
	if ins.Device != "cpu" && ins.Info.GPUMemoryMB > 0 {
		am.updateGPUMemory(ins.Device, -ins.Info.GPUMemoryMB)
	}

	am.instances.Delete(modelID)
	ins.Info.State = ModelStateReady
	ins.Info.Device = ""

	log.Printf("Model %s unloaded", modelID)

	return nil
}

// getModelInfo gets model info
func (am *AIManager) getModelInfo(modelID string) (*ModelInfo, bool) {
	if v, ok := am.models.Load(modelID); ok {
		return v.(*ModelInfo), true
	}
	return nil, false
}

// selectDevice selects the best device for a model
func (am *AIManager) selectDevice(info *ModelInfo) string {
	if !am.config.EnableGPU || len(am.gpus) == 0 {
		return "cpu"
	}

	// Find GPU with enough memory
	for _, gpu := range am.gpus {
		if gpu.MemoryFree >= info.GPUMemoryMB {
			return fmt.Sprintf("cuda:%d", gpu.ID)
		}
	}

	return "cpu"
}

// hasGPUMemory checks if GPU has enough memory
func (am *AIManager) hasGPUMemory(requiredMB int, device string) bool {
	for _, gpu := range am.gpus {
		if fmt.Sprintf("cuda:%d", gpu.ID) == device {
			return gpu.MemoryFree >= requiredMB
		}
	}
	return false
}

// updateGPUMemory updates GPU memory usage
func (am *AIManager) updateGPUMemory(device string, deltaMB int) {
	for i := range am.gpus {
		if fmt.Sprintf("cuda:%d", am.gpus[i].ID) == device {
			am.gpus[i].MemoryUsed += deltaMB
			am.gpus[i].MemoryFree -= deltaMB
			break
		}
	}
}

// createModelHandler creates a model handler
func (am *AIManager) createModelHandler(info *ModelInfo) (ModelHandler, error) {
	switch info.Format {
	case FormatONNX:
		return NewONNXHandler(info), nil
	case FormatGGML:
		return NewGGMLHandler(info), nil
	default:
		return NewGenericHandler(info), nil
	}
}

// Inference performs model inference
func (am *AIManager) Inference(req *InferenceRequest) (*InferenceResult, error) {
	req.CreatedAt = time.Now()

	// Queue request
	select {
	case am.requestQueue <- req:
	default:
		return nil, errors.New("inference queue full")
	}

	// Wait for result (simplified - in production would use callbacks)
	time.Sleep(10 * time.Millisecond)

	// Check for result
	select {
	case result := <-am.resultChan:
		return result, nil
	default:
		// Return pending status
		return &InferenceResult{
			RequestID: req.ID,
			ModelID:   req.ModelID,
			Success:   false,
			Error:     "inference pending",
		}, nil
	}
}

// inferenceProcessor processes inference requests
func (am *AIManager) inferenceProcessor() {
	for {
		select {
		case <-am.ctx.Done():
			return
		case req := <-am.requestQueue:
			go am.processInference(req)
		}
	}
}

// processInference processes a single inference request
func (am *AIManager) processInference(req *InferenceRequest) {
	startTime := time.Now()
	result := &InferenceResult{
		RequestID: req.ID,
		ModelID:   req.ModelID,
	}

	// Get model instance
	instance, ok := am.instances.Load(req.ModelID)
	if !ok {
		// Try to load model
		if err := am.LoadModel(req.ModelID, ""); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Model not loaded: %v", err)
			am.resultChan <- result
			return
		}
		instance, _ = am.instances.Load(req.ModelID)
	}

	ins := instance.(*ModelInstance)

	// Check if model is busy
	am.mu.Lock()
	if ins.Info.State == ModelStateBusy {
		am.mu.Unlock()
		// Re-queue request
		go func() {
			time.Sleep(100 * time.Millisecond)
			am.requestQueue <- req
		}()
		return
	}
	ins.Info.State = ModelStateBusy
	am.mu.Unlock()

	defer func() {
		am.mu.Lock()
		ins.Info.State = ModelStateReady
		ins.Info.LastUsed = time.Now()
		ins.Info.UseCount++
		am.mu.Unlock()
	}()

	// Execute inference
	output, err := ins.Handler.Infer(req.Input, req.Parameters)
	result.Duration = time.Since(startTime).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		am.mu.Lock()
		am.failedRequests++
		am.mu.Unlock()
	} else {
		result.Success = true
		result.Output = output
		am.mu.Lock()
		am.successRequests++
		am.totalLatency += result.Duration
		am.mu.Unlock()
	}

	am.mu.Lock()
	am.totalRequests++
	am.mu.Unlock()

	am.resultChan <- result
}

// gpuMonitor monitors GPU utilization
func (am *AIManager) gpuMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.updateGPUStats()
		}
	}
}

// updateGPUStats updates GPU statistics
func (am *AIManager) updateGPUStats() {
	// Placeholder for GPU stats update
	// In production, would use NVML or similar
}

// modelManager manages model lifecycle
func (am *AIManager) modelManager() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.manageModels()
		}
	}
}

// manageModels manages loaded models
func (am *AIManager) manageModels() {
	// Unload unused models to free GPU memory
	am.instances.Range(func(key, value interface{}) bool {
		instance := value.(*ModelInstance)

		// Unload models not used in last 30 minutes
		if time.Since(instance.Info.LastUsed) > 30*time.Minute && instance.Info.UseCount > 10 {
			if instance.Device != "cpu" {
				am.UnloadModel(instance.Info.ID)
			}
		}

		return true
	})
}

// GetModel returns model info
func (am *AIManager) GetModel(modelID string) (*ModelInfo, bool) {
	return am.getModelInfo(modelID)
}

// ListModels returns all registered models
func (am *AIManager) ListModels() []*ModelInfo {
	var models []*ModelInfo
	am.models.Range(func(key, value interface{}) bool {
		models = append(models, value.(*ModelInfo))
		return true
	})
	return models
}

// GetLoadedModels returns loaded models
func (am *AIManager) GetLoadedModels() []*ModelInfo {
	var models []*ModelInfo
	am.instances.Range(func(key, value interface{}) bool {
		instance := value.(*ModelInstance)
		models = append(models, instance.Info)
		return true
	})
	return models
}

// GetGPUInfo returns GPU information
func (am *AIManager) GetGPUInfo() []GPUInfo {
	return am.gpus
}

// GetStats returns AI manager statistics
func (am *AIManager) GetStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var avgLatency int64
	if am.successRequests > 0 {
		avgLatency = am.totalLatency / am.successRequests
	}

	return map[string]interface{}{
		"total_requests":    am.totalRequests,
		"success_requests":  am.successRequests,
		"failed_requests":   am.failedRequests,
		"avg_latency_ms":    avgLatency,
		"loaded_models":     len(am.GetLoadedModels()),
		"registered_models": len(am.ListModels()),
		"gpus":              am.gpus,
	}
}