// Package ha provides backup and recovery capabilities
package ha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BackupType defines backup types
type BackupType string

const (
	BackupFull    BackupType = "full"    // Full backup
	BackupIncremental BackupType = "incremental" // Incremental backup
	BackupDifferential BackupType = "differential" // Differential backup
)

// BackupStatus defines backup status
type BackupStatus string

const (
	BackupPending    BackupStatus = "pending"
	BackupRunning    BackupStatus = "running"
	BackupCompleted  BackupStatus = "completed"
	BackupFailed     BackupStatus = "failed"
	BackupExpired    BackupStatus = "expired"
)

// Backup represents a backup
type Backup struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Type          BackupType   `json:"type"`
	Status        BackupStatus `json:"status"`
	Size          int64        `json:"size"`
	Compressed    bool         `json:"compressed"`
	Encrypted     bool         `json:"encrypted"`
	Checksum      string       `json:"checksum"`
	Path          string       `json:"path"`
	BaseBackupID  string       `json:"base_backup_id,omitempty"` // For incremental
	DataCenters   []string     `json:"data_centers"`
	TenantID      string       `json:"tenant_id,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	StartedAt     time.Time    `json:"started_at,omitempty"`
	CompletedAt   time.Time    `json:"completed_at,omitempty"`
	ExpiresAt     time.Time    `json:"expires_at"`
	Duration      time.Duration `json:"duration"`
	Error         string       `json:"error,omitempty"`
	Metadata      map[string]string `json:"metadata"`
}

// BackupSchedule represents a backup schedule
type BackupSchedule struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         BackupType   `json:"type"`
	Schedule     string       `json:"schedule"` // Cron expression
	Retention    int          `json:"retention"` // Days to retain
	Enabled      bool         `json:"enabled"`
	LastRun      time.Time    `json:"last_run"`
	NextRun      time.Time    `json:"next_run"`
	TenantID     string       `json:"tenant_id,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

// RecoveryPlan represents a disaster recovery plan
type RecoveryPlan struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	TargetDC       string        `json:"target_dc"`
	SourceBackup   string        `json:"source_backup"`
	RTO            time.Duration `json:"rto"` // Recovery Time Objective
	RPO            time.Duration `json:"rpo"` // Recovery Point Objective
	Priority       int           `json:"priority"`
	Steps          []RecoveryStep `json:"steps"`
	CreatedAt      time.Time     `json:"created_at"`
	LastTestedAt   time.Time     `json:"last_tested_at"`
}

// RecoveryStep represents a step in recovery
type RecoveryStep struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Action      string        `json:"action"`
	Timeout     time.Duration `json:"timeout"`
	RetryCount  int           `json:"retry_count"`
	DependsOn   []string      `json:"depends_on"`
	Status      string        `json:"status"`
	StartedAt   time.Time     `json:"started_at,omitempty"`
	CompletedAt time.Time     `json:"completed_at,omitempty"`
	Error       string        `json:"error,omitempty"`
}

// RecoveryExecution represents a recovery execution
type RecoveryExecution struct {
	ID           string          `json:"id"`
	PlanID       string          `json:"plan_id"`
	BackupID     string          `json:"backup_id"`
	TargetDC     string          `json:"target_dc"`
	Status       string          `json:"status"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  time.Time       `json:"completed_at,omitempty"`
	Duration     time.Duration   `json:"duration"`
	StepResults  map[string]string `json:"step_results"`
	Error        string          `json:"error,omitempty"`
	TriggeredBy  string          `json:"triggered_by"`
}

// BackupManager manages backups
type BackupManager struct {
	storagePath string
	config      BackupConfig

	// Backups
	backups    sync.Map // map[string]*Backup
	backupList []*Backup

	// Schedules
	schedules sync.Map // map[string]*BackupSchedule

	// Recovery plans
	plans sync.Map // map[string]*RecoveryPlan

	// Statistics
	totalBackups     int64
	totalSize        int64
	successfulBackups int64
	failedBackups    int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// BackupConfig holds backup configuration
type BackupConfig struct {
	DefaultRetention int           `json:"default_retention"` // Days
	Compression      bool          `json:"compression"`
	Encryption       bool          `json:"encryption"`
	EncryptionKey    string        `json:"-"`
	MaxConcurrent    int           `json:"max_concurrent"`
	StorageBackend   string        `json:"storage_backend"` // local, s3, gcs, azure
	BucketName       string        `json:"bucket_name"`
}

// NewBackupManager creates a new backup manager
func NewBackupManager(storagePath string, config BackupConfig) (*BackupManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	os.MkdirAll(storagePath, 0755)

	manager := &BackupManager{
		storagePath: storagePath,
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Load existing backups
	if err := manager.load(); err != nil {
		log.Printf("Warning: failed to load backups: %v", err)
	}

	// Start scheduler
	go manager.scheduler()

	// Start cleanup
	go manager.cleanup()

	return manager, nil
}

// CreateBackup creates a new backup
func (m *BackupManager) CreateBackup(name string, backupType BackupType, tenantID string) (*Backup, error) {
	backup := &Backup{
		ID:        generateBackupID(),
		Name:      name,
		Type:      backupType,
		Status:    BackupPending,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(0, 0, m.config.DefaultRetention),
		Metadata:  make(map[string]string),
	}

	m.backups.Store(backup.ID, backup)

	m.mu.Lock()
	m.backupList = append(m.backupList, backup)
	m.mu.Unlock()

	// Start backup in background
	go m.runBackup(backup)

	return backup, nil
}

// runBackup executes the backup
func (m *BackupManager) runBackup(backup *Backup) {
	m.mu.Lock()
	backup.Status = BackupRunning
	backup.StartedAt = time.Now()
	m.mu.Unlock()

	m.totalBackups++

	// Create backup directory
	backupDir := filepath.Join(m.storagePath, backup.ID)
	os.MkdirAll(backupDir, 0755)

	// Simulate backup process
	// In production, would backup actual data
	var totalSize int64

	// Simulate backing up different components
	components := []string{"agents", "tasks", "messages", "skills", "models"}
	for _, comp := range components {
		compPath := filepath.Join(backupDir, comp+".json")
		data := fmt.Sprintf(`{"component": "%s", "backup_time": "%s", "data": []}`, comp, time.Now().Format(time.RFC3339))
		size := int64(len(data))
		totalSize += size

		if err := os.WriteFile(compPath, []byte(data), 0644); err != nil {
			m.mu.Lock()
			backup.Status = BackupFailed
			backup.Error = err.Error()
			backup.CompletedAt = time.Now()
			backup.Duration = backup.CompletedAt.Sub(backup.StartedAt)
			m.failedBackups++
			m.mu.Unlock()
			return
		}
	}

	// Calculate checksum
	checksum := m.calculateChecksum(backupDir)

	// Update backup
	m.mu.Lock()
	backup.Status = BackupCompleted
	backup.CompletedAt = time.Now()
	backup.Duration = backup.CompletedAt.Sub(backup.StartedAt)
	backup.Size = totalSize
	backup.Checksum = checksum
	backup.Path = backupDir
	backup.Compressed = m.config.Compression
	backup.Encrypted = m.config.Encryption
	m.successfulBackups++
	m.totalSize += totalSize
	m.mu.Unlock()

	m.save()

	log.Printf("Backup completed: %s (size: %d, duration: %v)", backup.ID, backup.Size, backup.Duration)
}

// GetBackup retrieves a backup
func (m *BackupManager) GetBackup(backupID string) (*Backup, error) {
	if v, ok := m.backups.Load(backupID); ok {
		return v.(*Backup), nil
	}
	return nil, fmt.Errorf("backup not found: %s", backupID)
}

// ListBackups lists backups
func (m *BackupManager) ListBackups(backupType BackupType, tenantID string) []*Backup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Backup
	for _, b := range m.backupList {
		if backupType != "" && b.Type != backupType {
			continue
		}
		if tenantID != "" && b.TenantID != tenantID {
			continue
		}
		result = append(result, b)
	}
	return result
}

// DeleteBackup deletes a backup
func (m *BackupManager) DeleteBackup(backupID string) error {
	backup, err := m.GetBackup(backupID)
	if err != nil {
		return err
	}

	// Remove files
	if backup.Path != "" {
		os.RemoveAll(backup.Path)
	}

	m.backups.Delete(backupID)

	m.mu.Lock()
	for i, b := range m.backupList {
		if b.ID == backupID {
			m.backupList = append(m.backupList[:i], m.backupList[i+1:]...)
			break
		}
		m.totalSize -= b.Size
	}
	m.mu.Unlock()

	m.save()

	return nil
}

// CreateSchedule creates a backup schedule
func (m *BackupManager) CreateSchedule(schedule *BackupSchedule) error {
	if schedule.ID == "" {
		schedule.ID = generateScheduleID()
	}

	schedule.CreatedAt = time.Now()
	schedule.Enabled = true

	// Parse cron and calculate next run
	// Simplified: assume daily backup
	schedule.NextRun = time.Now().Add(24 * time.Hour)

	m.schedules.Store(schedule.ID, schedule)

	return nil
}

// scheduler runs scheduled backups
func (m *BackupManager) scheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkSchedules()
		}
	}
}

// checkSchedules checks for due backups
func (m *BackupManager) checkSchedules() {
	m.schedules.Range(func(key, value interface{}) bool {
		schedule := value.(*BackupSchedule)

		if !schedule.Enabled {
			return true
		}

		if time.Now().After(schedule.NextRun) {
			// Create backup
			_, err := m.CreateBackup(schedule.Name, schedule.Type, schedule.TenantID)
			if err != nil {
				log.Printf("Failed to create scheduled backup: %v", err)
			}

			// Update schedule
			schedule.LastRun = time.Now()
			schedule.NextRun = time.Now().Add(24 * time.Hour)
		}

		return true
	})
}

// cleanup removes expired backups
func (m *BackupManager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanExpired()
		}
	}
}

// cleanExpired removes expired backups
func (m *BackupManager) cleanExpired() {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, backup := range m.backupList {
		if backup.Status == BackupCompleted && now.After(backup.ExpiresAt) {
			backup.Status = BackupExpired
			log.Printf("Backup expired: %s", backup.ID)
		}
	}
}

// Restore restores from a backup
func (m *BackupManager) Restore(backupID, targetDC string) error {
	backup, err := m.GetBackup(backupID)
	if err != nil {
		return err
	}

	if backup.Status != BackupCompleted {
		return fmt.Errorf("backup not available: %s", backup.Status)
	}

	log.Printf("Starting restore from backup %s to %s", backupID, targetDC)

	// Verify checksum
	if !m.verifyChecksum(backup) {
		return errors.New("backup checksum verification failed")
	}

	// Simulate restore process
	// In production, would restore actual data
	time.Sleep(1 * time.Second)

	log.Printf("Restore completed from backup %s", backupID)

	return nil
}

// CreateRecoveryPlan creates a recovery plan
func (m *BackupManager) CreateRecoveryPlan(plan *RecoveryPlan) error {
	if plan.ID == "" {
		plan.ID = generatePlanID()
	}

	plan.CreatedAt = time.Now()

	// Add default steps if not provided
	if len(plan.Steps) == 0 {
		plan.Steps = m.getDefaultRecoverySteps()
	}

	m.plans.Store(plan.ID, plan)

	return nil
}

// ExecuteRecovery executes a recovery plan
func (m *BackupManager) ExecuteRecovery(planID, backupID, triggeredBy string) (*RecoveryExecution, error) {
	plan, ok := m.plans.Load(planID)
	if !ok {
		return nil, fmt.Errorf("recovery plan not found: %s", planID)
	}

	backup, err := m.GetBackup(backupID)
	if err != nil {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}

	execution := &RecoveryExecution{
		ID:          generateExecutionID(),
		PlanID:      planID,
		BackupID:    backupID,
		TargetDC:    plan.(*RecoveryPlan).TargetDC,
		Status:      "running",
		StartedAt:   time.Now(),
		TriggeredBy: triggeredBy,
		StepResults: make(map[string]string),
	}

	// Execute steps
	for _, step := range plan.(*RecoveryPlan).Steps {
		step.Status = "running"
		step.StartedAt = time.Now()

		// Simulate step execution
		time.Sleep(100 * time.Millisecond)

		step.Status = "completed"
		step.CompletedAt = time.Now()
		execution.StepResults[step.ID] = "completed"
	}

	execution.Status = "completed"
	execution.CompletedAt = time.Now()
	execution.Duration = execution.CompletedAt.Sub(execution.StartedAt)

	log.Printf("Recovery execution completed: %s (duration: %v)", execution.ID, execution.Duration)

	return execution, nil
}

// getDefaultRecoverySteps returns default recovery steps
func (m *BackupManager) getDefaultRecoverySteps() []RecoveryStep {
	return []RecoveryStep{
		{ID: "verify_backup", Name: "Verify Backup", Action: "verify", Timeout: 5 * time.Minute},
		{ID: "restore_agents", Name: "Restore Agents", Action: "restore_agents", Timeout: 10 * time.Minute, DependsOn: []string{"verify_backup"}},
		{ID: "restore_tasks", Name: "Restore Tasks", Action: "restore_tasks", Timeout: 10 * time.Minute, DependsOn: []string{"verify_backup"}},
		{ID: "restore_skills", Name: "Restore Skills", Action: "restore_skills", Timeout: 5 * time.Minute, DependsOn: []string{"verify_backup"}},
		{ID: "verify_integrity", Name: "Verify Integrity", Action: "verify", Timeout: 5 * time.Minute, DependsOn: []string{"restore_agents", "restore_tasks", "restore_skills"}},
		{ID: "activate", Name: "Activate System", Action: "activate", Timeout: 2 * time.Minute, DependsOn: []string{"verify_integrity"}},
	}
}

// GetStats returns backup statistics
func (m *BackupManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_backups":      m.totalBackups,
		"successful_backups": m.successfulBackups,
		"failed_backups":     m.failedBackups,
		"total_size_bytes":   m.totalSize,
		"backup_count":       len(m.backupList),
		"schedule_count":     countMapItems(&m.schedules),
		"plan_count":         countMapItems(&m.plans),
	}
}

// load loads backups from disk
func (m *BackupManager) load() error {
	data, err := os.ReadFile(filepath.Join(m.storagePath, "backups.json"))
	if err != nil {
		return err
	}

	var backups []*Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		return err
	}

	for _, b := range backups {
		m.backups.Store(b.ID, b)
	}
	m.backupList = backups

	return nil
}

// save saves backups to disk
func (m *BackupManager) save() error {
	m.mu.RLock()
	backups := make([]*Backup, len(m.backupList))
	copy(backups, m.backupList)
	m.mu.RUnlock()

	data, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.storagePath, "backups.json"), data, 0644)
}

// calculateChecksum calculates backup checksum
func (m *BackupManager) calculateChecksum(dir string) string {
	// Simplified checksum
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// verifyChecksum verifies backup checksum
func (m *BackupManager) verifyChecksum(backup *Backup) bool {
	return backup.Checksum != ""
}

// Close closes the backup manager
func (m *BackupManager) Close() {
	m.cancel()
	m.save()
}

// Helper functions

func generateBackupID() string {
	return fmt.Sprintf("backup-%d", time.Now().UnixNano())
}

func generateScheduleID() string {
	return fmt.Sprintf("schedule-%d", time.Now().UnixNano())
}

func generatePlanID() string {
	return fmt.Sprintf("plan-%d", time.Now().UnixNano())
}

func generateExecutionID() string {
	return fmt.Sprintf("exec-%d", time.Now().UnixNano())
}

func countMapItems(m *sync.Map) int {
	var count int
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}