// Package modelversion provides enhanced model version management
// with A/B testing, rollback, and version comparison capabilities.
package modelversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// VersionStatus defines model version status
type VersionStatus string

const (
	VersionStatusActive    VersionStatus = "active"    // Currently in use
	VersionStatusStaged    VersionStatus = "staged"    // Staged for deployment
	VersionStatusDeprecated VersionStatus = "deprecated" // Marked for removal
	VersionStatusArchived  VersionStatus = "archived"  // Archived
)

// ABTestStatus defines A/B test status
type ABTestStatus string

const (
	ABTestRunning   ABTestStatus = "running"
	ABTestCompleted ABTestStatus = "completed"
	ABTestCancelled ABTestStatus = "cancelled"
)

// ModelVersion represents a model version
type ModelVersion struct {
	ModelID       string            `json:"model_id"`
	Version       string            `json:"version"`
	Checksum      string            `json:"checksum"`
	Size          int64             `json:"size"`
	Path          string            `json:"path"`
	Status        VersionStatus     `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	ActivatedAt   time.Time         `json:"activated_at,omitempty"`
	DeprecatedAt  time.Time         `json:"deprecated_at,omitempty"`
	Metrics       VersionMetrics    `json:"metrics"`
	Tags          []string          `json:"tags"`
	Description   string            `json:"description"`
	ParentVersion string            `json:"parent_version,omitempty"`
	Metadata      map[string]string `json:"metadata"`
}

// VersionMetrics holds performance metrics for a version
type VersionMetrics struct {
	Accuracy     float64 `json:"accuracy"`
	Precision    float64 `json:"precision"`
	Recall       float64 `json:"recall"`
	F1Score      float64 `json:"f1_score"`
	LatencyP50   int64   `json:"latency_p50_ms"`
	LatencyP99   int64   `json:"latency_p99_ms"`
	Throughput   float64 `json:"throughput"` // requests per second
	MemoryUsage  int64   `json:"memory_usage_mb"`
	RequestCount int64   `json:"request_count"`
	ErrorRate    float64 `json:"error_rate"`
}

// ABTestConfig holds A/B test configuration
type ABTestConfig struct {
	ID            string        `json:"id"`
	ModelID       string        `json:"model_id"`
	VersionA      string        `json:"version_a"`
	VersionB      string        `json:"version_b"`
	TrafficSplit  float64       `json:"traffic_split"` // 0.0-1.0 for version B
	Duration      time.Duration `json:"duration"`
	MetricName    string        `json:"metric_name"`   // Primary metric to compare
	MinSampleSize int           `json:"min_sample_size"`
	Significance  float64       `json:"significance"` // Statistical significance threshold
	CreatedAt     time.Time     `json:"created_at"`
	StartedAt     time.Time     `json:"started_at,omitempty"`
	EndedAt       time.Time     `json:"ended_at,omitempty"`
}

// ABTestResult holds A/B test results
type ABTestResult struct {
	TestID       string         `json:"test_id"`
	Status       ABTestStatus   `json:"status"`
	Winner       string         `json:"winner"` // version_a, version_b, or none
	MetricsA     ABTestMetrics  `json:"metrics_a"`
	MetricsB     ABTestMetrics  `json:"metrics_b"`
	Confidence   float64        `json:"confidence"`
	SampleSizeA  int            `json:"sample_size_a"`
	SampleSizeB  int            `json:"sample_size_b"`
	Conclusion   string         `json:"conclusion"`
}

// ABTestMetrics holds metrics collected during A/B test
type ABTestMetrics struct {
	Mean          float64 `json:"mean"`
	StdDev        float64 `json:"std_dev"`
	Median        float64 `json:"median"`
	Percentile95  float64 `json:"percentile_95"`
	RequestCount  int     `json:"request_count"`
}

// VersionRegistry manages model versions
type VersionRegistry struct {
	storagePath string

	// Version storage
	versions sync.Map // map[string]map[string]*ModelVersion (modelID -> version -> ModelVersion)
	activeVersions sync.Map // map[string]string (modelID -> active version)

	// A/B testing
	abTests    sync.Map // map[string]*ABTestConfig
	abResults  sync.Map // map[string]*ABTestResult

	// Rollback history
	rollbackHistory sync.Map // map[string][]RollbackRecord

	mu sync.RWMutex
}

// RollbackRecord represents a rollback operation
type RollbackRecord struct {
	ID            string    `json:"id"`
	ModelID       string    `json:"model_id"`
	FromVersion   string    `json:"from_version"`
	ToVersion     string    `json:"to_version"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
	InitiatedBy   string    `json:"initiated_by"`
}

// NewVersionRegistry creates a new version registry
func NewVersionRegistry(storagePath string) (*VersionRegistry, error) {
	os.MkdirAll(storagePath, 0755)

	registry := &VersionRegistry{
		storagePath: storagePath,
	}

	// Load existing data
	if err := registry.load(); err != nil {
		log.Printf("Warning: failed to load registry: %v", err)
	}

	return registry, nil
}

// RegisterVersion registers a new model version
func (r *VersionRegistry) RegisterVersion(version *ModelVersion) error {
	if version.ModelID == "" || version.Version == "" {
		return errors.New("model_id and version required")
	}

	version.CreatedAt = time.Now()
	version.Status = VersionStatusStaged

	// Calculate checksum if not provided
	if version.Checksum == "" {
		version.Checksum = r.calculateChecksum(version.Path)
	}

	// Store version
	versionMap, _ := r.versions.LoadOrStore(version.ModelID, &sync.Map{})
	vm := versionMap.(*sync.Map)
	vm.Store(version.Version, version)

	r.save()

	log.Printf("Registered version: %s@%s", version.ModelID, version.Version)

	return nil
}

// ActivateVersion activates a model version
func (r *VersionRegistry) ActivateVersion(modelID, version string) error {
	ver, err := r.GetVersion(modelID, version)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Deactivate current active version
	if currentActive, ok := r.activeVersions.Load(modelID); ok {
		if currentVer, err := r.GetVersion(modelID, currentActive.(string)); err == nil {
			currentVer.Status = VersionStatusDeprecated
			currentVer.DeprecatedAt = time.Now()
		}
	}

	// Activate new version
	ver.Status = VersionStatusActive
	ver.ActivatedAt = time.Now()
	r.activeVersions.Store(modelID, version)

	r.save()

	log.Printf("Activated version: %s@%s", modelID, version)

	return nil
}

// Rollback rolls back to a previous version
func (r *VersionRegistry) Rollback(modelID, targetVersion, reason string) error {
	currentVersion, ok := r.activeVersions.Load(modelID)
	if !ok {
		return fmt.Errorf("no active version for model: %s", modelID)
	}

	target, err := r.GetVersion(modelID, targetVersion)
	if err != nil {
		return err
	}

	if target.Status == VersionStatusDeprecated {
		target.Status = VersionStatusStaged
		target.DeprecatedAt = time.Time{}
	}

	// Create rollback record
	record := RollbackRecord{
		ID:          generateID(),
		ModelID:     modelID,
		FromVersion: currentVersion.(string),
		ToVersion:   targetVersion,
		Reason:      reason,
		Timestamp:   time.Now(),
	}

	// Store rollback record
	history, _ := r.rollbackHistory.LoadOrStore(modelID, []RollbackRecord{})
	historyList := history.([]RollbackRecord)
	historyList = append(historyList, record)
	r.rollbackHistory.Store(modelID, historyList)

	// Activate target version
	return r.ActivateVersion(modelID, targetVersion)
}

// GetVersion returns a specific version
func (r *VersionRegistry) GetVersion(modelID, version string) (*ModelVersion, error) {
	versionMap, ok := r.versions.Load(modelID)
	if !ok {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	vm := versionMap.(*sync.Map)
	ver, ok := vm.Load(version)
	if !ok {
		return nil, fmt.Errorf("version not found: %s@%s", modelID, version)
	}

	return ver.(*ModelVersion), nil
}

// GetActiveVersion returns the active version for a model
func (r *VersionRegistry) GetActiveVersion(modelID string) (*ModelVersion, error) {
	version, ok := r.activeVersions.Load(modelID)
	if !ok {
		return nil, fmt.Errorf("no active version for model: %s", modelID)
	}

	return r.GetVersion(modelID, version.(string))
}

// ListVersions lists all versions for a model
func (r *VersionRegistry) ListVersions(modelID string) ([]*ModelVersion, error) {
	versionMap, ok := r.versions.Load(modelID)
	if !ok {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	var versions []*ModelVersion
	vm := versionMap.(*sync.Map)
	vm.Range(func(key, value interface{}) bool {
		versions = append(versions, value.(*ModelVersion))
		return true
	})

	return versions, nil
}

// DeprecateVersion marks a version as deprecated
func (r *VersionRegistry) DeprecateVersion(modelID, version string) error {
	ver, err := r.GetVersion(modelID, version)
	if err != nil {
		return err
	}

	if ver.Status == VersionStatusActive {
		return errors.New("cannot deprecate active version")
	}

	ver.Status = VersionStatusDeprecated
	ver.DeprecatedAt = time.Now()
	r.save()

	return nil
}

// ArchiveVersion archives a version
func (r *VersionRegistry) ArchiveVersion(modelID, version string) error {
	ver, err := r.GetVersion(modelID, version)
	if err != nil {
		return err
	}

	ver.Status = VersionStatusArchived
	r.save()

	return nil
}

// CompareVersions compares two versions
func (r *VersionRegistry) CompareVersions(modelID, versionA, versionB string) (*VersionComparison, error) {
	verA, err := r.GetVersion(modelID, versionA)
	if err != nil {
		return nil, err
	}

	verB, err := r.GetVersion(modelID, versionB)
	if err != nil {
		return nil, err
	}

	comparison := &VersionComparison{
		ModelID:   modelID,
		VersionA:  versionA,
		VersionB:  versionB,
		Timestamp: time.Now(),
	}

	// Compare metrics
	comparison.MetricsDiff = r.compareMetrics(&verA.Metrics, &verB.Metrics)

	// Determine winner based on accuracy (can be customized)
	if verA.Metrics.Accuracy > verB.Metrics.Accuracy {
		comparison.Winner = versionA
		comparison.AccuracyDiff = verA.Metrics.Accuracy - verB.Metrics.Accuracy
	} else if verB.Metrics.Accuracy > verA.Metrics.Accuracy {
		comparison.Winner = versionB
		comparison.AccuracyDiff = verB.Metrics.Accuracy - verA.Metrics.Accuracy
	} else {
		comparison.Winner = "tie"
		comparison.AccuracyDiff = 0
	}

	return comparison, nil
}

// VersionComparison represents version comparison result
type VersionComparison struct {
	ModelID      string            `json:"model_id"`
	VersionA     string            `json:"version_a"`
	VersionB     string            `json:"version_b"`
	Winner       string            `json:"winner"`
	AccuracyDiff float64           `json:"accuracy_diff"`
	MetricsDiff  map[string]float64 `json:"metrics_diff"`
	Timestamp    time.Time         `json:"timestamp"`
}

// compareMetrics compares metrics between versions
func (r *VersionRegistry) compareMetrics(a, b *VersionMetrics) map[string]float64 {
	diff := make(map[string]float64)
	diff["accuracy"] = a.Accuracy - b.Accuracy
	diff["precision"] = a.Precision - b.Precision
	diff["recall"] = a.Recall - b.Recall
	diff["f1_score"] = a.F1Score - b.F1Score
	diff["latency_p50"] = float64(a.LatencyP50 - b.LatencyP50)
	diff["latency_p99"] = float64(a.LatencyP99 - b.LatencyP99)
	diff["throughput"] = a.Throughput - b.Throughput
	diff["error_rate"] = a.ErrorRate - b.ErrorRate
	return diff
}

// CreateABTest creates an A/B test
func (r *VersionRegistry) CreateABTest(config *ABTestConfig) error {
	if config.ID == "" {
		config.ID = generateID()
	}
	config.CreatedAt = time.Now()

	// Validate versions exist
	if _, err := r.GetVersion(config.ModelID, config.VersionA); err != nil {
		return fmt.Errorf("version_a not found: %v", err)
	}
	if _, err := r.GetVersion(config.ModelID, config.VersionB); err != nil {
		return fmt.Errorf("version_b not found: %v", err)
	}

	r.abTests.Store(config.ID, config)

	return nil
}

// StartABTest starts an A/B test
func (r *VersionRegistry) StartABTest(testID string) error {
	test, ok := r.abTests.Load(testID)
	if !ok {
		return fmt.Errorf("test not found: %s", testID)
	}

	t := test.(*ABTestConfig)
	t.StartedAt = time.Now()

	return nil
}

// EndABTest ends an A/B test
func (r *VersionRegistry) EndABTest(testID string, winner string) error {
	test, ok := r.abTests.Load(testID)
	if !ok {
		return fmt.Errorf("test not found: %s", testID)
	}

	t := test.(*ABTestConfig)
	t.EndedAt = time.Now()

	// Store result
	result := &ABTestResult{
		TestID: testID,
		Status: ABTestCompleted,
		Winner: winner,
	}
	r.abResults.Store(testID, result)

	// Activate winning version
	if winner == "version_b" {
		r.ActivateVersion(t.ModelID, t.VersionB)
	}

	return nil
}

// GetABTest returns A/B test info
func (r *VersionRegistry) GetABTest(testID string) (*ABTestConfig, bool) {
	if v, ok := r.abTests.Load(testID); ok {
		return v.(*ABTestConfig), true
	}
	return nil, false
}

// GetABTestResult returns A/B test result
func (r *VersionRegistry) GetABTestResult(testID string) (*ABTestResult, bool) {
	if v, ok := r.abResults.Load(testID); ok {
		return v.(*ABTestResult), true
	}
	return nil, false
}

// GetRollbackHistory returns rollback history for a model
func (r *VersionRegistry) GetRollbackHistory(modelID string) []RollbackRecord {
	if v, ok := r.rollbackHistory.Load(modelID); ok {
		return v.([]RollbackRecord)
	}
	return nil
}

// UpdateVersionMetrics updates metrics for a version
func (r *VersionRegistry) UpdateVersionMetrics(modelID, version string, metrics VersionMetrics) error {
	ver, err := r.GetVersion(modelID, version)
	if err != nil {
		return err
	}

	ver.Metrics = metrics
	r.save()

	return nil
}

// load loads registry from disk
func (r *VersionRegistry) load() error {
	if r.storagePath == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(r.storagePath, "versions.json"))
	if err != nil {
		return err
	}

	var savedData struct {
		Versions       map[string]map[string]*ModelVersion `json:"versions"`
		ActiveVersions map[string]string                   `json:"active_versions"`
	}

	if err := json.Unmarshal(data, &savedData); err != nil {
		return err
	}

	for modelID, versions := range savedData.Versions {
		vm := &sync.Map{}
		for v, ver := range versions {
			vm.Store(v, ver)
		}
		r.versions.Store(modelID, vm)
	}

	for modelID, version := range savedData.ActiveVersions {
		r.activeVersions.Store(modelID, version)
	}

	return nil
}

// save saves registry to disk
func (r *VersionRegistry) save() error {
	if r.storagePath == "" {
		return nil
	}

	versions := make(map[string]map[string]*ModelVersion)
	r.versions.Range(func(key, value interface{}) bool {
		modelID := key.(string)
		vm := value.(*sync.Map)
		versions[modelID] = make(map[string]*ModelVersion)
		vm.Range(func(k, v interface{}) bool {
			versions[modelID][k.(string)] = v.(*ModelVersion)
			return true
		})
		return true
	})

	activeVersions := make(map[string]string)
	r.activeVersions.Range(func(key, value interface{}) bool {
		activeVersions[key.(string)] = value.(string)
		return true
	})

	savedData := struct {
		Versions       map[string]map[string]*ModelVersion `json:"versions"`
		ActiveVersions map[string]string                   `json:"active_versions"`
	}{
		Versions:       versions,
		ActiveVersions: activeVersions,
	}

	data, err := json.Marshal(savedData)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(r.storagePath, "versions.json"), data, 0644)
}

// calculateChecksum calculates model checksum
func (r *VersionRegistry) calculateChecksum(path string) string {
	// Placeholder - would use actual checksum calculation
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}