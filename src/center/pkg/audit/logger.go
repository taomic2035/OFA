// Package audit provides security audit logging
package audit

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

// AuditEventType defines audit event types
type AuditEventType string

const (
	// Authentication events
	EventLogin         AuditEventType = "auth.login"
	EventLogout        AuditEventType = "auth.logout"
	EventLoginFailed   AuditEventType = "auth.login_failed"
	EventTokenRefresh  AuditEventType = "auth.token_refresh"

	// Authorization events
	EventAccessGranted  AuditEventType = "auth.access_granted"
	EventAccessDenied   AuditEventType = "auth.access_denied"
	EventRoleAssign     AuditEventType = "auth.role_assign"
	EventRoleRevoke     AuditEventType = "auth.role_revoke"

	// Resource events
	EventCreate AuditEventType = "resource.create"
	EventRead   AuditEventType = "resource.read"
	EventUpdate AuditEventType = "resource.update"
	EventDelete AuditEventType = "resource.delete"

	// System events
	EventConfigChange AuditEventType = "system.config_change"
	EventBackupCreate AuditEventType = "system.backup_create"
	EventBackupRestore AuditEventType = "system.backup_restore"
	EventFailover     AuditEventType = "system.failover"
	EventDeploy       AuditEventType = "system.deploy"

	// Security events
	EventSecurityAlert  AuditEventType = "security.alert"
	EventKeyGenerate    AuditEventType = "security.key_generate"
	EventKeyRotate      AuditEventType = "security.key_rotate"
	EventCertIssue      AuditEventType = "security.cert_issue"
	EventCertRevoke     AuditEventType = "security.cert_revoke"

	// Tenant events
	EventTenantCreate   AuditEventType = "tenant.create"
	EventTenantUpdate   AuditEventType = "tenant.update"
	EventTenantDelete   AuditEventType = "tenant.delete"
	EventTenantSuspend  AuditEventType = "tenant.suspend"

	// Agent events
	EventAgentRegister  AuditEventType = "agent.register"
	EventAgentUnregister AuditEventType = "agent.unregister"
	EventTaskSubmit     AuditEventType = "agent.task_submit"
	EventTaskComplete   AuditEventType = "agent.task_complete"
)

// AuditSeverity defines severity levels
type AuditSeverity string

const (
	SeverityInfo     AuditSeverity = "info"
	SeverityWarning  AuditSeverity = "warning"
	SeverityError    AuditSeverity = "error"
	SeverityCritical AuditSeverity = "critical"
)

// AuditEvent represents an audit event
type AuditEvent struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	EventType    AuditEventType         `json:"event_type"`
	Severity     AuditSeverity          `json:"severity"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Actor        ActorInfo              `json:"actor"`
	Target       TargetInfo             `json:"target,omitempty"`
	TenantID     string                 `json:"tenant_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	SourceIP     string                 `json:"source_ip,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Location     string                 `json:"location,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Changes      []ChangeRecord         `json:"changes,omitempty"`
	Status       string                 `json:"status"` // success, failure
	Error        string                 `json:"error,omitempty"`
	Duration     time.Duration          `json:"duration,omitempty"`
}

// ActorInfo holds information about the actor
type ActorInfo struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // user, service, agent, system
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

// TargetInfo holds information about the target
type TargetInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"` // agent, task, skill, model, etc.
	Name string `json:"name,omitempty"`
}

// ChangeRecord represents a change in an audit event
type ChangeRecord struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// AuditConfig holds audit configuration
type AuditConfig struct {
	Enabled       bool   `json:"enabled"`
	StoragePath   string `json:"storage_path"`
	RetentionDays int    `json:"retention_days"`
	MaxFileSize   int64  `json:"max_file_size_mb"`
	FlushInterval time.Duration `json:"flush_interval"`
	CompressOld   bool   `json:"compress_old"`
	Encrypt       bool   `json:"encrypt"`
}

// AuditLogger handles audit logging
type AuditLogger struct {
	config AuditConfig

	// Event buffer
	buffer    []*AuditEvent
	bufferMu  sync.Mutex

	// Storage
	events    []*AuditEvent
	eventsMu  sync.RWMutex

	// Statistics
	totalEvents     int64
	eventsByType    map[AuditEventType]int64
	eventsBySeverity map[AuditSeverity]int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditConfig) (*AuditLogger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if config.RetentionDays == 0 {
		config.RetentionDays = 90
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}

	logger := &AuditLogger{
		config:          config,
		buffer:          make([]*AuditEvent, 0),
		events:          make([]*AuditEvent, 0),
		eventsByType:    make(map[AuditEventType]int64),
		eventsBySeverity: make(map[AuditSeverity]int64),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Create storage directory
	if config.StoragePath != "" {
		os.MkdirAll(config.StoragePath, 0755)
	}

	// Start flusher
	go logger.flusher()

	// Start retention cleaner
	go logger.retentionCleaner()

	return logger, nil
}

// Log logs an audit event
func (l *AuditLogger) Log(event *AuditEvent) error {
	if !l.config.Enabled {
		return nil
	}

	// Set defaults
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}
	if event.Status == "" {
		event.Status = "success"
	}

	// Buffer event
	l.bufferMu.Lock()
	l.buffer = append(l.buffer, event)
	l.bufferMu.Unlock()

	// Update statistics
	l.mu.Lock()
	l.totalEvents++
	l.eventsByType[event.EventType]++
	l.eventsBySeverity[event.Severity]++
	l.mu.Unlock()

	return nil
}

// LogSimple logs a simple audit event
func (l *AuditLogger) LogSimple(eventType AuditEventType, action, resource, resourceID, actorID string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType:  eventType,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Actor: ActorInfo{
			ID:   actorID,
			Type: "user",
		},
		Details: details,
	}

	return l.Log(event)
}

// LogWithActor logs an event with full actor info
func (l *AuditLogger) LogWithActor(eventType AuditEventType, actor ActorInfo, resource, resourceID, action string, changes []ChangeRecord) error {
	event := &AuditEvent{
		EventType:  eventType,
		Actor:      actor,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
		Changes:    changes,
	}

	return l.Log(event)
}

// LogError logs an error event
func (l *AuditLogger) LogError(eventType AuditEventType, action, resource, error string, actorID string) error {
	event := &AuditEvent{
		EventType: eventType,
		Action:    action,
		Resource:  resource,
		Severity:  SeverityError,
		Status:    "failure",
		Error:     error,
		Actor: ActorInfo{
			ID:   actorID,
			Type: "user",
		},
	}

	return l.Log(event)
}

// flusher periodically flushes buffer to storage
func (l *AuditLogger) flusher() {
	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			l.flush()
			return
		case <-ticker.C:
			l.flush()
		}
	}
}

// flush flushes buffer to storage
func (l *AuditLogger) flush() {
	l.bufferMu.Lock()
	if len(l.buffer) == 0 {
		l.bufferMu.Unlock()
		return
	}

	events := l.buffer
	l.buffer = make([]*AuditEvent, 0)
	l.bufferMu.Unlock()

	// Add to memory
	l.eventsMu.Lock()
	l.events = append(l.events, events...)
	l.eventsMu.Unlock()

	// Write to disk
	if l.config.StoragePath != "" {
		l.writeToDisk(events)
	}
}

// writeToDisk writes events to disk
func (l *AuditLogger) writeToDisk(events []*AuditEvent) error {
	filename := fmt.Sprintf("audit-%s.json", time.Now().Format("2006-01-02"))
	path := filepath.Join(l.config.StoragePath, filename)

	// Read existing
	var existing []*AuditEvent
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &existing)
	}

	// Append new
	all := append(existing, events...)

	// Write
	data, err := json.Marshal(all)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// retentionCleaner removes old audit logs
func (l *AuditLogger) retentionCleaner() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.cleanOldLogs()
		}
	}
}

// cleanOldLogs removes logs older than retention period
func (l *AuditLogger) cleanOldLogs() {
	if l.config.StoragePath == "" {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -l.config.RetentionDays)

	// Clean memory
	l.eventsMu.Lock()
	var newEvents []*AuditEvent
	for _, e := range l.events {
		if e.Timestamp.After(cutoff) {
			newEvents = append(newEvents, e)
		}
	}
	l.events = newEvents
	l.eventsMu.Unlock()

	// Clean disk
	files, _ := filepath.Glob(filepath.Join(l.config.StoragePath, "audit-*.json"))
	for _, f := range files {
		info, err := os.Stat(f)
		if err == nil && info.ModTime().Before(cutoff) {
			os.Remove(f)
			log.Printf("Removed old audit log: %s", f)
		}
	}
}

// Query queries audit events
func (l *AuditLogger) Query(filter AuditFilter) ([]*AuditEvent, error) {
	l.eventsMu.RLock()
	defer l.eventsMu.RUnlock()

	var result []*AuditEvent

	for _, e := range l.events {
		if l.matchesFilter(e, filter) {
			result = append(result, e)
		}
	}

	// Apply limit
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[len(result)-filter.Limit:]
	}

	return result, nil
}

// AuditFilter defines audit query filters
type AuditFilter struct {
	EventType  AuditEventType `json:"event_type,omitempty"`
	Severity   AuditSeverity  `json:"severity,omitempty"`
	ActorID    string         `json:"actor_id,omitempty"`
	Resource   string         `json:"resource,omitempty"`
	ResourceID string         `json:"resource_id,omitempty"`
	TenantID   string         `json:"tenant_id,omitempty"`
	Status     string         `json:"status,omitempty"`
	StartTime  time.Time      `json:"start_time,omitempty"`
	EndTime    time.Time      `json:"end_time,omitempty"`
	Limit      int            `json:"limit,omitempty"`
}

// matchesFilter checks if event matches filter
func (l *AuditLogger) matchesFilter(event *AuditEvent, filter AuditFilter) bool {
	if filter.EventType != "" && event.EventType != filter.EventType {
		return false
	}
	if filter.Severity != "" && event.Severity != filter.Severity {
		return false
	}
	if filter.ActorID != "" && event.Actor.ID != filter.ActorID {
		return false
	}
	if filter.Resource != "" && event.Resource != filter.Resource {
		return false
	}
	if filter.ResourceID != "" && event.ResourceID != filter.ResourceID {
		return false
	}
	if filter.TenantID != "" && event.TenantID != filter.TenantID {
		return false
	}
	if filter.Status != "" && event.Status != filter.Status {
		return false
	}
	if !filter.StartTime.IsZero() && event.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && event.Timestamp.After(filter.EndTime) {
		return false
	}
	return true
}

// GetStats returns audit statistics
func (l *AuditLogger) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	l.eventsMu.RLock()
	eventCount := len(l.events)
	l.eventsMu.RUnlock()

	return map[string]interface{}{
		"enabled":           l.config.Enabled,
		"total_events":      l.totalEvents,
		"events_in_memory":  eventCount,
		"events_by_type":    l.eventsByType,
		"events_by_severity": l.eventsBySeverity,
		"retention_days":    l.config.RetentionDays,
	}
}

// GenerateReport generates an audit report
func (l *AuditLogger) GenerateReport(startTime, endTime time.Time, tenantID string) (*AuditReport, error) {
	filter := AuditFilter{
		StartTime: startTime,
		EndTime:   endTime,
		TenantID:  tenantID,
	}

	events, err := l.Query(filter)
	if err != nil {
		return nil, err
	}

	report := &AuditReport{
		StartTime:     startTime,
		EndTime:       endTime,
		TenantID:      tenantID,
		TotalEvents:   len(events),
		GeneratedAt:   time.Now(),
		ByType:        make(map[AuditEventType]int),
		BySeverity:    make(map[AuditSeverity]int),
		ByActor:       make(map[string]int),
		FailedActions: make([]FailedAction, 0),
	}

	for _, e := range events {
		report.ByType[e.EventType]++
		report.BySeverity[e.Severity]++
		report.ByActor[e.Actor.ID]++

		if e.Status == "failure" {
			report.FailedActions = append(report.FailedActions, FailedAction{
				Timestamp:  e.Timestamp,
				EventType:  e.EventType,
				Action:     e.Action,
				Resource:   e.Resource,
				Error:      e.Error,
			})
		}
	}

	return report, nil
}

// AuditReport represents an audit report
type AuditReport struct {
	StartTime     time.Time                 `json:"start_time"`
	EndTime       time.Time                 `json:"end_time"`
	TenantID      string                    `json:"tenant_id,omitempty"`
	TotalEvents   int                       `json:"total_events"`
	GeneratedAt   time.Time                 `json:"generated_at"`
	ByType        map[AuditEventType]int    `json:"by_type"`
	BySeverity    map[AuditSeverity]int     `json:"by_severity"`
	ByActor       map[string]int            `json:"by_actor"`
	FailedActions []FailedAction            `json:"failed_actions"`
}

// FailedAction represents a failed action
type FailedAction struct {
	Timestamp time.Time       `json:"timestamp"`
	EventType AuditEventType  `json:"event_type"`
	Action    string          `json:"action"`
	Resource  string          `json:"resource"`
	Error     string          `json:"error"`
}

// Close closes the audit logger
func (l *AuditLogger) Close() {
	l.cancel()
	l.flush()
}

// Helper function
func generateEventID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}