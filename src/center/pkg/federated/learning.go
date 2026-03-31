// Package federated provides federated learning capabilities for OFA
// including model training tasks, data privacy protection,
// gradient aggregation, and model distribution.
package federated

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TrainingStatus defines training status
type TrainingStatus string

const (
	StatusPending    TrainingStatus = "pending"
	StatusRunning    TrainingStatus = "running"
	StatusCompleted  TrainingStatus = "completed"
	StatusFailed     TrainingStatus = "failed"
	StatusCancelled  TrainingStatus = "cancelled"
)

// AggregationStrategy defines aggregation strategies
type AggregationStrategy string

const (
	AggregationFedAvg  AggregationStrategy = "fedavg"   // Federated Averaging
	AggregationFedProx AggregationStrategy = "fedprox" // Federated Proximal
	AggregationFedSGD  AggregationStrategy = "fedsgd"  // Federated SGD
	AggregationSCAFFOLD AggregationStrategy = "scaffold" // SCAFFOLD
)

// PrivacyMethod defines privacy protection methods
type PrivacyMethod string

const (
	PrivacyNone          PrivacyMethod = "none"
	PrivacyDP            PrivacyMethod = "dp"            // Differential Privacy
	PrivacySecureAgg     PrivacyMethod = "secure_agg"    // Secure Aggregation
	PrivacyHE            PrivacyMethod = "he"            // Homomorphic Encryption
	PrivacyFLDP          PrivacyMethod = "fldp"          // Federated Learning with DP
)

// FederatedConfig holds federated learning configuration
type FederatedConfig struct {
	MinClients       int
	MaxClients       int
	MinRounds        int
	MaxRounds        int
	TargetAccuracy   float64
	Aggregation      AggregationStrategy
	PrivacyMethod    PrivacyMethod
	DPEpsilon        float64
	DPDelta          float64
	ClipNorm         float64
	ClientTimeout    time.Duration
	CheckpointDir    string
}

// DefaultFederatedConfig returns default configuration
func DefaultFederatedConfig() *FederatedConfig {
	return &FederatedConfig{
		MinClients:     3,
		MaxClients:     100,
		MinRounds:      1,
		MaxRounds:      100,
		TargetAccuracy: 0.95,
		Aggregation:    AggregationFedAvg,
		PrivacyMethod:  PrivacyNone,
		DPEpsilon:      1.0,
		DPDelta:        1e-5,
		ClipNorm:       1.0,
		ClientTimeout:  5 * time.Minute,
	}
}

// TrainingTask represents a federated training task
type TrainingTask struct {
	ID              string             `json:"id"`
	ModelID         string             `json:"model_id"`
	ModelVersion    string             `json:"model_version"`
	Status          TrainingStatus     `json:"status"`
	Round           int                `json:"round"`
	TotalRounds     int                `json:"total_rounds"`
	Aggregation     AggregationStrategy `json:"aggregation"`
	PrivacyMethod   PrivacyMethod      `json:"privacy_method"`
	Config          FederatedConfig    `json:"config"`
	SelectedClients []string           `json:"selected_clients"`
	CompletedClients []string          `json:"completed_clients"`
	GlobalModel     []byte             `json:"global_model,omitempty"`
	Metrics         TrainingMetrics    `json:"metrics"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	CompletedAt     time.Time          `json:"completed_at,omitempty"`
}

// TrainingMetrics holds training metrics
type TrainingMetrics struct {
	GlobalLoss     float64   `json:"global_loss"`
	GlobalAccuracy float64   `json:"global_accuracy"`
	RoundLosses    []float64 `json:"round_losses"`
	RoundAccuracies []float64 `json:"round_accuracies"`
	TotalSamples   int       `json:"total_samples"`
	TotalClients   int       `json:"total_clients"`
}

// ClientUpdate represents a client's model update
type ClientUpdate struct {
	TaskID         string    `json:"task_id"`
	ClientID       string    `json:"client_id"`
	Round          int       `json:"round"`
	ModelWeights   []byte    `json:"model_weights"`
	SampleCount    int       `json:"sample_count"`
	LocalLoss      float64   `json:"local_loss"`
	LocalAccuracy  float64   `json:"local_accuracy"`
	TrainingTime   int64     `json:"training_time_ms"`
	SubmittedAt    time.Time `json:"submitted_at"`
	Signature      string    `json:"signature"` // For secure aggregation
}

// ClientInfo represents a federated client
type ClientInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Capabilities []string  `json:"capabilities"`
	DataSize     int       `json:"data_size"`
	LastSeen     time.Time `json:"last_seen"`
	Reliability  float64   `json:"reliability"` // 0-1 score
	Active       bool      `json:"active"`
}

// FederatedManager manages federated learning operations
type FederatedManager struct {
	config *FederatedConfig

	// Tasks
	tasks       sync.Map // map[string]*TrainingTask
	activeTasks sync.Map // map[string]*TrainingTask

	// Clients
	clients      sync.Map // map[string]*ClientInfo
	clientQueue  chan string

	// Updates
	pendingUpdates sync.Map // map[string][]*ClientUpdate

	// Model storage
	modelStore *ModelStore

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// ModelStore stores model versions
type ModelStore struct {
	dir string
	mu  sync.RWMutex
}

// NewFederatedManager creates a new federated learning manager
func NewFederatedManager(config *FederatedConfig) (*FederatedManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &FederatedManager{
		config:      config,
		clientQueue: make(chan string, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize model store
	if config.CheckpointDir != "" {
		manager.modelStore = &ModelStore{dir: config.CheckpointDir}
		os.MkdirAll(config.CheckpointDir, 0755)
	}

	return manager, nil
}

// Start starts the federated manager
func (fm *FederatedManager) Start() {
	go fm.taskScheduler()
	go fm.aggregationWorker()
	go fm.clientMonitor()
}

// Stop stops the federated manager
func (fm *FederatedManager) Stop() {
	fm.cancel()
}

// RegisterClient registers a new federated client
func (fm *FederatedManager) RegisterClient(info *ClientInfo) error {
	if info.ID == "" {
		return errors.New("client ID required")
	}

	info.LastSeen = time.Now()
	info.Active = true
	info.Reliability = 1.0

	fm.clients.Store(info.ID, info)

	log.Printf("Client registered: %s", info.ID)

	return nil
}

// UnregisterClient unregisters a client
func (fm *FederatedManager) UnregisterClient(clientID string) {
	fm.clients.Delete(clientID)
	log.Printf("Client unregistered: %s", clientID)
}

// GetClient returns client info
func (fm *FederatedManager) GetClient(clientID string) (*ClientInfo, bool) {
	if v, ok := fm.clients.Load(clientID); ok {
		return v.(*ClientInfo), true
	}
	return nil, false
}

// ListClients returns all registered clients
func (fm *FederatedManager) ListClients() []*ClientInfo {
	var clients []*ClientInfo
	fm.clients.Range(func(key, value interface{}) bool {
		clients = append(clients, value.(*ClientInfo))
		return true
	})
	return clients
}

// GetActiveClients returns active clients
func (fm *FederatedManager) GetActiveClients() []*ClientInfo {
	var clients []*ClientInfo
	fm.clients.Range(func(key, value interface{}) bool {
		client := value.(*ClientInfo)
		if client.Active {
			clients = append(clients, client)
		}
		return true
	})
	return clients
}

// CreateTask creates a new training task
func (fm *FederatedManager) CreateTask(modelID, modelVersion string, config *FederatedConfig) (*TrainingTask, error) {
	if config == nil {
		config = fm.config
	}

	taskID := generateTaskID()

	task := &TrainingTask{
		ID:           taskID,
		ModelID:      modelID,
		ModelVersion: modelVersion,
		Status:       StatusPending,
		Round:        0,
		TotalRounds:  config.MaxRounds,
		Aggregation:  config.Aggregation,
		PrivacyMethod: config.PrivacyMethod,
		Config:       *config,
		Metrics:      TrainingMetrics{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	fm.tasks.Store(taskID, task)

	log.Printf("Training task created: %s", taskID)

	return task, nil
}

// GetTask returns a training task
func (fm *FederatedManager) GetTask(taskID string) (*TrainingTask, bool) {
	if v, ok := fm.tasks.Load(taskID); ok {
		return v.(*TrainingTask), true
	}
	return nil, false
}

// StartTask starts a training task
func (fm *FederatedManager) StartTask(taskID string) error {
	task, ok := fm.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	fm.mu.Lock()
	if task.Status != StatusPending {
		fm.mu.Unlock()
		return fmt.Errorf("task already started")
	}
	task.Status = StatusRunning
	task.UpdatedAt = time.Now()
	fm.mu.Unlock()

	fm.activeTasks.Store(taskID, task)

	// Start first round
	go fm.startRound(task)

	log.Printf("Training task started: %s", taskID)

	return nil
}

// startRound starts a new training round
func (fm *FederatedManager) startRound(task *TrainingTask) {
	task.Round++
	task.SelectedClients = nil
	task.CompletedClients = nil

	// Select clients for this round
	clients := fm.selectClients(task.Config)
	if len(clients) < task.Config.MinClients {
		log.Printf("Not enough clients for task %s round %d", task.ID, task.Round)
		task.Status = StatusFailed
		return
	}

	task.SelectedClients = clients

	// Distribute global model to clients
	fm.distributeModel(task, clients)

	// Wait for client updates
	go fm.waitForUpdates(task)
}

// selectClients selects clients for training
func (fm *FederatedManager) selectClients(config FederatedConfig) []string {
	activeClients := fm.GetActiveClients()

	if len(activeClients) <= config.MaxClients {
		// Use all active clients
		var selected []string
		for _, c := range activeClients {
			selected = append(selected, c.ID)
		}
		return selected
	}

	// Random selection with reliability weighting
	selected := make(map[string]bool)
	for len(selected) < config.MaxClients {
		// Weight by reliability
		for _, client := range activeClients {
			if len(selected) >= config.MaxClients {
				break
			}
			if rand.Float64() < client.Reliability && !selected[client.ID] {
				selected[client.ID] = true
			}
		}
	}

	var result []string
	for id := range selected {
		result = append(result, id)
	}

	return result
}

// distributeModel distributes global model to clients
func (fm *FederatedManager) distributeModel(task *TrainingTask, clients []string) {
	// Placeholder for model distribution
	for _, clientID := range clients {
		// In production, would send model to client
		log.Printf("Distributing model to client %s for task %s round %d", clientID, task.ID, task.Round)
	}
}

// waitForUpdates waits for client updates
func (fm *FederatedManager) waitForUpdates(task *TrainingTask) {
	timeout := time.After(task.Config.ClientTimeout)

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-timeout:
			// Check if we have enough updates
			updates, ok := fm.pendingUpdates.Load(task.ID)
			if ok {
				updatesList := updates.([]*ClientUpdate)
				if len(updatesList) >= task.Config.MinClients {
					// Proceed with aggregation
					go fm.aggregate(task, updatesList)
					return
				}
			}

			log.Printf("Timeout waiting for updates for task %s round %d", task.ID, task.Round)
			// Retry or fail
			return
		}
	}
}

// SubmitUpdate submits a client update
func (fm *FederatedManager) SubmitUpdate(update *ClientUpdate) error {
	task, ok := fm.GetTask(update.TaskID)
	if !ok {
		return fmt.Errorf("task not found: %s", update.TaskID)
	}

	if task.Status != StatusRunning {
		return errors.New("task not running")
	}

	// Verify client is selected
	selected := false
	for _, c := range task.SelectedClients {
		if c == update.ClientID {
			selected = true
			break
		}
	}

	if !selected {
		return errors.New("client not selected for this round")
	}

	// Apply privacy protection if configured
	if task.PrivacyMethod != PrivacyNone {
		if err := fm.applyPrivacy(update, task); err != nil {
			return err
		}
	}

	// Store update
	update.SubmittedAt = time.Now()

	pending, _ := fm.pendingUpdates.LoadOrStore(update.TaskID, []*ClientUpdate{})
	pendingList := pending.([]*ClientUpdate)
	pendingList = append(pendingList, update)
	fm.pendingUpdates.Store(update.TaskID, pendingList)

	task.CompletedClients = append(task.CompletedClients, update.ClientID)

	// Update client reliability
	if client, ok := fm.GetClient(update.ClientID); ok {
		client.LastSeen = time.Now()
	}

	log.Printf("Update received from client %s for task %s round %d", update.ClientID, update.TaskID, update.Round)

	// Check if we have enough updates
	if len(pendingList) >= task.Config.MinClients {
		go fm.aggregate(task, pendingList)
	}

	return nil
}

// applyPrivacy applies privacy protection to update
func (fm *FederatedManager) applyPrivacy(update *ClientUpdate, task *TrainingTask) error {
	switch task.PrivacyMethod {
	case PrivacyDP:
		return fm.applyDP(update, task.Config.DPEpsilon, task.Config.DPDelta)
	case PrivacySecureAgg:
		return fm.applySecureAggregation(update)
	case PrivacyHE:
		return fm.applyHomomorphicEncryption(update)
	}
	return nil
}

// applyDP applies differential privacy
func (fm *FederatedManager) applyDP(update *ClientUpdate, epsilon, delta float64) error {
	// Placeholder for DP noise addition
	// In production, would add calibrated noise to gradients
	return nil
}

// applySecureAggregation applies secure aggregation
func (fm *FederatedManager) applySecureAggregation(update *ClientUpdate) error {
	// Placeholder for secure aggregation
	return nil
}

// applyHomomorphicEncryption applies homomorphic encryption
func (fm *FederatedManager) applyHomomorphicEncryption(update *ClientUpdate) error {
	// Placeholder for HE
	return nil
}

// aggregate aggregates client updates
func (fm *FederatedManager) aggregate(task *TrainingTask, updates []*ClientUpdate) {
	log.Printf("Aggregating %d updates for task %s round %d", len(updates), task.ID, task.Round)

	// Clear pending updates
	fm.pendingUpdates.Delete(task.ID)

	// Perform aggregation based on strategy
	var aggregated []byte
	var err error

	switch task.Aggregation {
	case AggregationFedAvg:
		aggregated, err = fm.fedAvgAggregate(updates)
	case AggregationFedProx:
		aggregated, err = fm.fedProxAggregate(updates)
	case AggregationFedSGD:
		aggregated, err = fm.fedSGDAggregate(updates)
	default:
		aggregated, err = fm.fedAvgAggregate(updates)
	}

	if err != nil {
		log.Printf("Aggregation failed: %v", err)
		task.Status = StatusFailed
		return
	}

	// Update global model
	task.GlobalModel = aggregated

	// Calculate metrics
	fm.calculateMetrics(task, updates)

	// Check termination conditions
	if task.Round >= task.TotalRounds || task.Metrics.GlobalAccuracy >= task.Config.TargetAccuracy {
		fm.completeTask(task)
		return
	}

	// Start next round
	time.Sleep(1 * time.Second)
	go fm.startRound(task)
}

// fedAvgAggregate performs federated averaging
func (fm *FederatedManager) fedAvgAggregate(updates []*ClientUpdate) ([]byte, error) {
	// Placeholder for FedAvg aggregation
	// In production, would average model weights weighted by sample count

	totalSamples := 0
	for _, u := range updates {
		totalSamples += u.SampleCount
	}

	// Simulate aggregated model
	result := []byte("aggregated_model_weights")

	return result, nil
}

// fedProxAggregate performs FedProx aggregation
func (fm *FederatedManager) fedProxAggregate(updates []*ClientUpdate) ([]byte, error) {
	// Placeholder for FedProx
	return fm.fedAvgAggregate(updates)
}

// fedSGDAggregate performs FedSGD aggregation
func (fm *FederatedManager) fedSGDAggregate(updates []*ClientUpdate) ([]byte, error) {
	// Placeholder for FedSGD
	return fm.fedAvgAggregate(updates)
}

// calculateMetrics calculates training metrics
func (fm *FederatedManager) calculateMetrics(task *TrainingTask, updates []*ClientUpdate) {
	var totalLoss, totalAcc float64
	var totalSamples int

	for _, u := range updates {
		totalLoss += u.LocalLoss * float64(u.SampleCount)
		totalAcc += u.LocalAccuracy * float64(u.SampleCount)
		totalSamples += u.SampleCount
	}

	if totalSamples > 0 {
		task.Metrics.GlobalLoss = totalLoss / float64(totalSamples)
		task.Metrics.GlobalAccuracy = totalAcc / float64(totalSamples)
	}

	task.Metrics.RoundLosses = append(task.Metrics.RoundLosses, task.Metrics.GlobalLoss)
	task.Metrics.RoundAccuracies = append(task.Metrics.RoundAccuracies, task.Metrics.GlobalAccuracy)
	task.Metrics.TotalSamples = totalSamples
	task.Metrics.TotalClients = len(updates)

	task.UpdatedAt = time.Now()
}

// completeTask completes a training task
func (fm *FederatedManager) completeTask(task *TrainingTask) {
	task.Status = StatusCompleted
	task.CompletedAt = time.Now()
	task.UpdatedAt = time.Now()

	fm.activeTasks.Delete(task.ID)

	// Save final model
	if fm.modelStore != nil {
		fm.modelStore.Save(task.ModelID, task.ModelVersion, task.GlobalModel)
	}

	log.Printf("Training task completed: %s (accuracy: %.4f)", task.ID, task.Metrics.GlobalAccuracy)
}

// CancelTask cancels a training task
func (fm *FederatedManager) CancelTask(taskID string) error {
	task, ok := fm.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	task.Status = StatusCancelled
	task.UpdatedAt = time.Now()

	fm.activeTasks.Delete(taskID)
	fm.pendingUpdates.Delete(taskID)

	return nil
}

// ListTasks lists all tasks
func (fm *FederatedManager) ListTasks() []*TrainingTask {
	var tasks []*TrainingTask
	fm.tasks.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*TrainingTask))
		return true
	})
	return tasks
}

// GetActiveTasks returns active tasks
func (fm *FederatedManager) GetActiveTasks() []*TrainingTask {
	var tasks []*TrainingTask
	fm.activeTasks.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*TrainingTask))
		return true
	})
	return tasks
}

// taskScheduler schedules training tasks
func (fm *FederatedManager) taskScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			// Check for pending tasks
			fm.tasks.Range(func(key, value interface{}) bool {
				task := value.(*TrainingTask)
				if task.Status == StatusPending {
					// Check if we have enough clients
					if len(fm.GetActiveClients()) >= task.Config.MinClients {
						fm.StartTask(task.ID)
					}
				}
				return true
			})
		}
	}
}

// aggregationWorker processes aggregation
func (fm *FederatedManager) aggregationWorker() {
	// Placeholder for aggregation worker
}

// clientMonitor monitors client health
func (fm *FederatedManager) clientMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			fm.clients.Range(func(key, value interface{}) bool {
				client := value.(*ClientInfo)
				// Mark inactive if not seen in 5 minutes
				if time.Since(client.LastSeen) > 5*time.Minute {
					client.Active = false
				}
				return true
			})
		}
	}
}

// GetStats returns federated learning statistics
func (fm *FederatedManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_clients":  len(fm.ListClients()),
		"active_clients": len(fm.GetActiveClients()),
		"total_tasks":    len(fm.ListTasks()),
		"active_tasks":   len(fm.GetActiveTasks()),
		"aggregation":    fm.config.Aggregation,
		"privacy_method": fm.config.PrivacyMethod,
	}
}

// ModelStore methods

// Save saves a model to storage
func (s *ModelStore) Save(modelID, version string, weights []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, fmt.Sprintf("%s_%s.model", modelID, version))
	return os.WriteFile(path, weights, 0644)
}

// Load loads a model from storage
func (s *ModelStore) Load(modelID, version string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dir, fmt.Sprintf("%s_%s.model", modelID, version))
	return os.ReadFile(path)
}

// generateTaskID generates a unique task ID
func generateTaskID() string {
	return fmt.Sprintf("fl-%d-%s", time.Now().Unix(), randomString(6))
}

// randomString generates a random string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}