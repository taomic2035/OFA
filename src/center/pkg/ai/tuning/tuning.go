// Package tuning provides automatic hyperparameter tuning capabilities
package tuning

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TuningStrategy defines tuning strategies
type TuningStrategy string

const (
	StrategyGridSearch   TuningStrategy = "grid_search"   // Exhaustive grid search
	StrategyRandomSearch TuningStrategy = "random_search" // Random search
	StrategyBayesian     TuningStrategy = "bayesian"      // Bayesian optimization
	StrategyHyperband    TuningStrategy = "hyperband"     // Hyperband
	StrategyBOHB         TuningStrategy = "bohb"          // Bayesian + Hyperband
)

// ParameterType defines parameter types
type ParameterType string

const (
	ParamInt    ParameterType = "int"
	ParamFloat  ParameterType = "float"
	ParamString ParameterType = "string"
	ParamBool   ParameterType = "bool"
	ParamCategorical ParameterType = "categorical"
)

// SearchSpace defines the search space for a parameter
type SearchSpace struct {
	Name     string        `json:"name"`
	Type     ParameterType `json:"type"`
	Min      interface{}   `json:"min,omitempty"`      // For int/float
	Max      interface{}   `json:"max,omitempty"`      // For int/float
	Step     interface{}   `json:"step,omitempty"`     // For int/float
	Choices  []interface{} `json:"choices,omitempty"`  // For categorical
	LogScale bool          `json:"log_scale,omitempty"` // Use log scale
}

// HyperParameter represents a hyperparameter value
type HyperParameter struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// TrialConfig represents a trial configuration
type TrialConfig struct {
	ID          string          `json:"id"`
	Parameters  []HyperParameter `json:"parameters"`
	Score       float64         `json:"score"`
	Status      TrialStatus     `json:"status"`
	Duration    int64           `json:"duration_ms"`
	Error       string          `json:"error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt time.Time       `json:"completed_at,omitempty"`
}

// TrialStatus defines trial status
type TrialStatus string

const (
	TrialPending   TrialStatus = "pending"
	TrialRunning   TrialStatus = "running"
	TrialCompleted TrialStatus = "completed"
	TrialFailed    TrialStatus = "failed"
	TrialPruned    TrialStatus = "pruned"
)

// TuningConfig holds tuning configuration
type TuningConfig struct {
	Strategy       TuningStrategy `json:"strategy"`
	MaxTrials      int            `json:"max_trials"`
	MaxConcurrency int            `json:"max_concurrency"`
	Timeout        time.Duration  `json:"timeout"`
	Direction      string         `json:"direction"` // minimize or maximize
	EarlyStopping  bool           `json:"early_stopping"`
	Patience       int            `json:"patience"`
	SavePath       string         `json:"save_path"`
}

// DefaultTuningConfig returns default configuration
func DefaultTuningConfig() *TuningConfig {
	return &TuningConfig{
		Strategy:       StrategyBayesian,
		MaxTrials:      100,
		MaxConcurrency: 4,
		Timeout:        1 * time.Hour,
		Direction:      "minimize",
		EarlyStopping:  true,
		Patience:       10,
	}
}

// TuningJob represents a hyperparameter tuning job
type TuningJob struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	ModelID      string         `json:"model_id"`
	SearchSpaces []SearchSpace  `json:"search_spaces"`
	Config       TuningConfig   `json:"config"`
	Trials       []*TrialConfig `json:"trials"`
	BestTrial    *TrialConfig   `json:"best_trial,omitempty"`
	Status       JobStatus      `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	StartedAt    time.Time      `json:"started_at,omitempty"`
	CompletedAt  time.Time      `json:"completed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// JobStatus defines tuning job status
type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"
)

// TrialEvaluator evaluates a trial
type TrialEvaluator interface {
	Evaluate(params map[string]interface{}) (float64, error)
}

// AutoTuner manages hyperparameter tuning
type AutoTuner struct {
	jobs     sync.Map // map[string]*TuningJob
	evaluator TrialEvaluator

	// Bayesian optimization state
	surrogateModel sync.Map // map[string]*GaussianProcess

	// Results storage
	resultsPath string

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewAutoTuner creates a new auto tuner
func NewAutoTuner(resultsPath string) *AutoTuner {
	ctx, cancel := context.WithCancel(context.Background())

	os.MkdirAll(resultsPath, 0755)

	return &AutoTuner{
		resultsPath: resultsPath,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// SetEvaluator sets the trial evaluator
func (t *AutoTuner) SetEvaluator(evaluator TrialEvaluator) {
	t.evaluator = evaluator
}

// CreateJob creates a new tuning job
func (t *AutoTuner) CreateJob(name, modelID string, searchSpaces []SearchSpace, config *TuningConfig) (*TuningJob, error) {
	if config == nil {
		config = DefaultTuningConfig()
	}

	// Validate search spaces
	if len(searchSpaces) == 0 {
		return nil, errors.New("search spaces required")
	}

	job := &TuningJob{
		ID:           generateJobID(),
		Name:         name,
		ModelID:      modelID,
		SearchSpaces: searchSpaces,
		Config:       *config,
		Trials:       make([]*TrialConfig, 0),
		Status:       JobPending,
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	t.jobs.Store(job.ID, job)

	return job, nil
}

// StartJob starts a tuning job
func (t *AutoTuner) StartJob(jobID string) error {
	job, ok := t.GetJob(jobID)
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	t.mu.Lock()
	if job.Status != JobPending {
		t.mu.Unlock()
		return errors.New("job already started")
	}
	job.Status = JobRunning
	job.StartedAt = time.Now()
	t.mu.Unlock()

	// Run tuning based on strategy
	go t.runTuning(job)

	return nil
}

// runTuning runs the tuning process
func (t *AutoTuner) runTuning(job *TuningJob) {
	defer func() {
		t.mu.Lock()
		if job.Status == JobRunning {
			job.Status = JobCompleted
		}
		job.CompletedAt = time.Now()
		t.mu.Unlock()
		t.saveJob(job)
	}()

	for i := 0; i < job.Config.MaxTrials; i++ {
		select {
		case <-t.ctx.Done():
			job.Status = JobCancelled
			return
		default:
		}

		// Generate trial config
		trial := t.generateTrial(job, i)
		trial.Status = TrialRunning
		trial.CreatedAt = time.Now()

		t.mu.Lock()
		job.Trials = append(job.Trials, trial)
		t.mu.Unlock()

		// Run trial
		startTime := time.Now()
		score, err := t.runTrial(job, trial)
		trial.Duration = time.Since(startTime).Milliseconds()
		trial.CompletedAt = time.Now()

		if err != nil {
			trial.Status = TrialFailed
			trial.Error = err.Error()
		} else {
			trial.Status = TrialCompleted
			trial.Score = score
		}

		// Update best trial
		t.updateBestTrial(job, trial)

		// Check early stopping
		if job.Config.EarlyStopping && t.shouldStopEarly(job) {
			log.Printf("Early stopping triggered for job %s", job.ID)
			break
		}
	}

	t.saveJob(job)
}

// generateTrial generates a trial configuration
func (t *AutoTuner) generateTrial(job *TuningJob, iteration int) *TrialConfig {
	params := make([]HyperParameter, len(job.SearchSpaces))

	switch job.Config.Strategy {
	case StrategyGridSearch:
		params = t.generateGridSearch(job.SearchSpaces, iteration)
	case StrategyRandomSearch:
		params = t.generateRandomSearch(job.SearchSpaces)
	case StrategyBayesian:
		params = t.generateBayesian(job.SearchSpaces, job.Trials)
	case StrategyHyperband:
		params = t.generateRandomSearch(job.SearchSpaces) // Simplified
	case StrategyBOHB:
		params = t.generateBayesian(job.SearchSpaces, job.Trials)
	default:
		params = t.generateRandomSearch(job.SearchSpaces)
	}

	return &TrialConfig{
		ID:         generateTrialID(job.ID, iteration),
		Parameters: params,
		Status:     TrialPending,
	}
}

// generateGridSearch generates grid search parameters
func (t *AutoTuner) generateGridSearch(spaces []SearchSpace, iteration int) []HyperParameter {
	params := make([]HyperParameter, len(spaces))

	// Simplified grid search
	for i, space := range spaces {
		params[i] = HyperParameter{
			Name:  space.Name,
			Value: t.sampleFromSpace(&space),
		}
	}

	return params
}

// generateRandomSearch generates random search parameters
func (t *AutoTuner) generateRandomSearch(spaces []SearchSpace) []HyperParameter {
	params := make([]HyperParameter, len(spaces))

	for i, space := range spaces {
		params[i] = HyperParameter{
			Name:  space.Name,
			Value: t.sampleFromSpace(&space),
		}
	}

	return params
}

// generateBayesian generates Bayesian optimization parameters
func (t *AutoTuner) generateBayesian(spaces []SearchSpace, pastTrials []*TrialConfig) []HyperParameter {
	// If not enough data, use random
	if len(pastTrials) < 5 {
		return t.generateRandomSearch(spaces)
	}

	// Simple Bayesian-like approach: favor regions with good scores
	// In production, would use Gaussian Process
	params := make([]HyperParameter, len(spaces))

	// Find best trial
	var bestTrial *TrialConfig
	for _, trial := range pastTrials {
		if trial.Status == TrialCompleted {
			if bestTrial == nil || trial.Score < bestTrial.Score {
				bestTrial = trial
			}
		}
	}

	// Sample around best trial with some exploration
	for i, space := range spaces {
		if bestTrial != nil && rand.Float64() < 0.7 {
			// Exploit: use best value with small perturbation
			bestValue := t.getParamValue(bestTrial.Parameters, space.Name)
			params[i] = HyperParameter{
				Name:  space.Name,
				Value: t.perturbValue(bestValue, &space),
			}
		} else {
			// Explore: random sample
			params[i] = HyperParameter{
				Name:  space.Name,
				Value: t.sampleFromSpace(&space),
			}
		}
	}

	return params
}

// sampleFromSpace samples a value from search space
func (t *AutoTuner) sampleFromSpace(space *SearchSpace) interface{} {
	switch space.Type {
	case ParamInt:
		min := space.Min.(int)
		max := space.Max.(int)
		return min + rand.Intn(max-min+1)
	case ParamFloat:
		min := space.Min.(float64)
		max := space.Max.(float64)
		if space.LogScale {
			// Log-uniform sampling
			logMin := math.Log(min)
			logMax := math.Log(max)
			return math.Exp(logMin + rand.Float64()*(logMax-logMin))
		}
		return min + rand.Float64()*(max-min)
	case ParamString:
		if len(space.Choices) > 0 {
			return space.Choices[rand.Intn(len(space.Choices))]
		}
		return ""
	case ParamBool:
		return rand.Intn(2) == 1
	case ParamCategorical:
		if len(space.Choices) > 0 {
			return space.Choices[rand.Intn(len(space.Choices))]
		}
		return nil
	}
	return nil
}

// perturbValue perturbs a value slightly
func (t *AutoTuner) perturbValue(value interface{}, space *SearchSpace) interface{} {
	switch space.Type {
	case ParamInt:
		v := value.(int)
		min := space.Min.(int)
		max := space.Max.(int)
		delta := int(float64(max-min) * 0.1) // 10% perturbation
		v = v + rand.Intn(2*delta+1) - delta
		if v < min {
			v = min
		}
		if v > max {
			v = max
		}
		return v
	case ParamFloat:
		v := value.(float64)
		min := space.Min.(float64)
		max := space.Max.(float64)
		delta := (max - min) * 0.1
		v = v + (rand.Float64()*2-1)*delta
		if v < min {
			v = min
		}
		if v > max {
			v = max
		}
		return v
	default:
		// For categorical, randomly choose from neighbors
		if len(space.Choices) > 1 {
			// Find current index and pick neighbor
			for i, c := range space.Choices {
				if c == value {
					// Pick random neighbor
					choices := []int{i - 1, i + 1}
					validChoices := []int{}
					for _, c := range choices {
						if c >= 0 && c < len(space.Choices) {
							validChoices = append(validChoices, c)
						}
					}
					if len(validChoices) > 0 {
						return space.Choices[validChoices[rand.Intn(len(validChoices))]]
					}
				}
			}
		}
		return value
	}
}

// getParamValue gets parameter value by name
func (t *AutoTuner) getParamValue(params []HyperParameter, name string) interface{} {
	for _, p := range params {
		if p.Name == name {
			return p.Value
		}
	}
	return nil
}

// runTrial runs a single trial
func (t *AutoTuner) runTrial(job *TuningJob, trial *TrialConfig) (float64, error) {
	if t.evaluator == nil {
		return 0, errors.New("no evaluator configured")
	}

	params := make(map[string]interface{})
	for _, p := range trial.Parameters {
		params[p.Name] = p.Value
	}

	return t.evaluator.Evaluate(params)
}

// updateBestTrial updates the best trial
func (t *AutoTuner) updateBestTrial(job *TuningJob, trial *TrialConfig) {
	if trial.Status != TrialCompleted {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if job.BestTrial == nil {
		job.BestTrial = trial
		return
	}

	// Compare based on direction
	if job.Config.Direction == "minimize" {
		if trial.Score < job.BestTrial.Score {
			job.BestTrial = trial
		}
	} else {
		if trial.Score > job.BestTrial.Score {
			job.BestTrial = trial
		}
	}
}

// shouldStopEarly checks if early stopping should trigger
func (t *AutoTuner) shouldStopEarly(job *TuningJob) bool {
	if len(job.Trials) < job.Config.Patience {
		return false
	}

	// Check if no improvement in last N trials
	bestScore := job.BestTrial.Score
	noImprovement := 0

	for i := len(job.Trials) - 1; i >= 0 && noImprovement < job.Config.Patience; i-- {
		trial := job.Trials[i]
		if trial.Status == TrialCompleted {
			if job.Config.Direction == "minimize" {
				if trial.Score >= bestScore {
					noImprovement++
				} else {
					break
				}
			} else {
				if trial.Score <= bestScore {
					noImprovement++
				} else {
					break
				}
			}
		}
	}

	return noImprovement >= job.Config.Patience
}

// CancelJob cancels a tuning job
func (t *AutoTuner) CancelJob(jobID string) error {
	job, ok := t.GetJob(jobID)
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	t.mu.Lock()
	job.Status = JobCancelled
	t.mu.Unlock()

	return nil
}

// GetJob gets a tuning job
func (t *AutoTuner) GetJob(jobID string) (*TuningJob, bool) {
	if v, ok := t.jobs.Load(jobID); ok {
		return v.(*TuningJob), true
	}
	return nil, false
}

// ListJobs lists all tuning jobs
func (t *AutoTuner) ListJobs() []*TuningJob {
	var jobs []*TuningJob
	t.jobs.Range(func(key, value interface{}) bool {
		jobs = append(jobs, value.(*TuningJob))
		return true
	})
	return jobs
}

// saveJob saves job to disk
func (t *AutoTuner) saveJob(job *TuningJob) error {
	if t.resultsPath == "" {
		return nil
	}

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	path := filepath.Join(t.resultsPath, job.ID+".json")
	return os.WriteFile(path, data, 0644)
}

// GetBestParameters returns best parameters from a job
func (t *AutoTuner) GetBestParameters(jobID string) (map[string]interface{}, error) {
	job, ok := t.GetJob(jobID)
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	if job.BestTrial == nil {
		return nil, errors.New("no completed trials")
	}

	params := make(map[string]interface{})
	for _, p := range job.BestTrial.Parameters {
		params[p.Name] = p.Value
	}

	return params, nil
}

// GetJobStats returns job statistics
func (t *AutoTuner) GetJobStats(jobID string) (map[string]interface{}, error) {
	job, ok := t.GetJob(jobID)
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	completed := 0
	failed := 0
	totalDuration := int64(0)

	for _, trial := range job.Trials {
		if trial.Status == TrialCompleted {
			completed++
			totalDuration += trial.Duration
		} else if trial.Status == TrialFailed {
			failed++
		}
	}

	avgDuration := int64(0)
	if completed > 0 {
		avgDuration = totalDuration / int64(completed)
	}

	return map[string]interface{}{
		"job_id":           job.ID,
		"status":           job.Status,
		"total_trials":     len(job.Trials),
		"completed_trials": completed,
		"failed_trials":    failed,
		"avg_duration_ms":  avgDuration,
		"best_score":       job.BestTrial.Score,
	}, nil
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("tune-%d", time.Now().UnixNano())
}

// generateTrialID generates a unique trial ID
func generateTrialID(jobID string, iteration int) string {
	return fmt.Sprintf("%s-trial-%d", jobID, iteration)
}