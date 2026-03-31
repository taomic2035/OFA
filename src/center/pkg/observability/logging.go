// Package observability provides centralized logging and alerting
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel defines log levels
type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
	LogFatal LogLevel = "fatal"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       LogLevel          `json:"level"`
	Message     string            `json:"message"`
	Service     string            `json:"service"`
	Component   string            `json:"component,omitempty"`
	TraceID     string            `json:"trace_id,omitempty"`
	SpanID      string            `json:"span_id,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	RequestID   string            `json:"request_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string            `json:"error,omitempty"`
	Stack       string            `json:"stack,omitempty"`
	Duration    time.Duration     `json:"duration,omitempty"`
}

// LogAggregator aggregates logs from multiple sources
type LogAggregator struct {
	serviceName string
	outputPath  string

	// Log storage
	entries sync.Map // map[string][]*LogEntry (tenantID -> entries)
	allLogs []*LogEntry

	// Log buffering
	buffer     []*LogEntry
	bufferSize int
	flushInterval time.Duration

	// Log level filtering
	minLevel LogLevel

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewLogAggregator creates a new log aggregator
func NewLogAggregator(serviceName, outputPath string) *LogAggregator {
	ctx, cancel := context.WithCancel(context.Background())

	aggregator := &LogAggregator{
		serviceName:    serviceName,
		outputPath:     outputPath,
		bufferSize:     1000,
		flushInterval:  5 * time.Second,
		minLevel:       LogInfo,
		buffer:         make([]*LogEntry, 0),
		allLogs:        make([]*LogEntry, 0),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Create output directory
	if outputPath != "" {
		os.MkdirAll(outputPath, 0755)
	}

	// Start flusher
	go aggregator.flusher()

	return aggregator
}

// Log writes a log entry
func (a *LogAggregator) Log(level LogLevel, message string, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   a.serviceName,
		Fields:    fields,
	}

	a.addEntry(entry)
}

// LogWithContext writes a log entry with context
func (a *LogAggregator) LogWithContext(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   a.serviceName,
		Fields:    fields,
	}

	// Extract context values
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		entry.TraceID = traceID
	}
	if spanID, ok := ctx.Value("span_id").(string); ok {
		entry.SpanID = spanID
	}
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		entry.TenantID = tenantID
	}
	if requestID, ok := ctx.Value("request_id").(string); ok {
		entry.RequestID = requestID
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		entry.UserID = userID
	}

	a.addEntry(entry)
}

// Debug logs a debug message
func (a *LogAggregator) Debug(message string, fields map[string]interface{}) {
	a.Log(LogDebug, message, fields)
}

// Info logs an info message
func (a *LogAggregator) Info(message string, fields map[string]interface{}) {
	a.Log(LogInfo, message, fields)
}

// Warn logs a warning message
func (a *LogAggregator) Warn(message string, fields map[string]interface{}) {
	a.Log(LogWarn, message, fields)
}

// Error logs an error message
func (a *LogAggregator) Error(message string, err error, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogError,
		Message:   message,
		Service:   a.serviceName,
		Fields:    fields,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	a.addEntry(entry)
}

// Fatal logs a fatal message and exits
func (a *LogAggregator) Fatal(message string, err error, fields map[string]interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     LogFatal,
		Message:   message,
		Service:   a.serviceName,
		Fields:    fields,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	a.addEntry(entry)
	a.Flush()

	log.Fatalf("%s: %v", message, err)
}

// addEntry adds an entry to the aggregator
func (a *LogAggregator) addEntry(entry *LogEntry) {
	// Check log level
	if !a.shouldLog(entry.Level) {
		return
	}

	a.mu.Lock()
	a.buffer = append(a.buffer, entry)
	a.allLogs = append(a.allLogs, entry)

	// Store by tenant
	if entry.TenantID != "" {
		logs, _ := a.entries.LoadOrStore(entry.TenantID, []*LogEntry{})
		tenantLogs := logs.([]*LogEntry)
		tenantLogs = append(tenantLogs, entry)
		a.entries.Store(entry.TenantID, tenantLogs)
	}

	// Flush if buffer is full
	if len(a.buffer) >= a.bufferSize {
		go a.Flush()
	}
	a.mu.Unlock()

	// Also log to stdout
	a.logToStdout(entry)
}

// shouldLog checks if level should be logged
func (a *LogAggregator) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogDebug: 0,
		LogInfo:  1,
		LogWarn:  2,
		LogError: 3,
		LogFatal: 4,
	}

	return levels[level] >= levels[a.minLevel]
}

// logToStdout logs to stdout
func (a *LogAggregator) logToStdout(entry *LogEntry) {
	msg := fmt.Sprintf("[%s] %s %s: %s",
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		entry.Level,
		entry.Service,
		entry.Message,
	)

	if entry.Error != "" {
		msg += fmt.Sprintf(" error=%s", entry.Error)
	}

	if entry.TraceID != "" {
		msg += fmt.Sprintf(" trace=%s", entry.TraceID)
	}

	if entry.TenantID != "" {
		msg += fmt.Sprintf(" tenant=%s", entry.TenantID)
	}

	log.Println(msg)
}

// Flush flushes buffer to disk
func (a *LogAggregator) Flush() error {
	if a.outputPath == "" {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.buffer) == 0 {
		return nil
	}

	// Create filename based on date
	filename := fmt.Sprintf("logs-%s.json", time.Now().Format("2006-01-02"))
	path := filepath.Join(a.outputPath, filename)

	// Read existing file
	var existing []*LogEntry
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &existing)
	}

	// Append new logs
	allLogs := append(existing, a.buffer...)

	// Write to file
	data, err := json.Marshal(allLogs)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Clear buffer
	a.buffer = make([]*LogEntry, 0)

	return nil
}

// flusher periodically flushes logs
func (a *LogAggregator) flusher() {
	ticker := time.NewTicker(a.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			a.Flush()
			return
		case <-ticker.C:
			a.Flush()
		}
	}
}

// Query queries logs with filters
func (a *LogAggregator) Query(filter LogFilter) ([]*LogEntry, error) {
	var results []*LogEntry

	for _, entry := range a.allLogs {
		if a.matchesFilter(entry, filter) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// LogFilter defines log query filters
type LogFilter struct {
	Level     LogLevel   `json:"level,omitempty"`
	Service   string     `json:"service,omitempty"`
	TenantID  string     `json:"tenant_id,omitempty"`
	TraceID   string     `json:"trace_id,omitempty"`
	StartTime time.Time  `json:"start_time,omitempty"`
	EndTime   time.Time  `json:"end_time,omitempty"`
	Message   string     `json:"message,omitempty"` // Contains
	Limit     int        `json:"limit,omitempty"`
}

// matchesFilter checks if entry matches filter
func (a *LogAggregator) matchesFilter(entry *LogEntry, filter LogFilter) bool {
	if filter.Level != "" && entry.Level != filter.Level {
		return false
	}
	if filter.Service != "" && entry.Service != filter.Service {
		return false
	}
	if filter.TenantID != "" && entry.TenantID != filter.TenantID {
		return false
	}
	if filter.TraceID != "" && entry.TraceID != filter.TraceID {
		return false
	}
	if !filter.StartTime.IsZero() && entry.Timestamp.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && entry.Timestamp.After(filter.EndTime) {
		return false
	}
	if filter.Message != "" && !contains(entry.Message, filter.Message) {
		return false
	}
	return true
}

// GetTenantLogs retrieves logs for a tenant
func (a *LogAggregator) GetTenantLogs(tenantID string, limit int) []*LogEntry {
	if v, ok := a.entries.Load(tenantID); ok {
		logs := v.([]*LogEntry)
		if limit > 0 && len(logs) > limit {
			return logs[len(logs)-limit:]
		}
		return logs
	}
	return nil
}

// GetStats returns logging statistics
func (a *LogAggregator) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Count by level
	countByLevel := make(map[LogLevel]int)
	for _, entry := range a.allLogs {
		countByLevel[entry.Level]++
	}

	// Count by tenant
	countByTenant := make(map[string]int)
	a.entries.Range(func(key, value interface{}) bool {
		countByTenant[key.(string)] = len(value.([]*LogEntry))
		return true
	})

	return map[string]interface{}{
		"service_name":    a.serviceName,
		"total_entries":   len(a.allLogs),
		"buffer_size":     len(a.buffer),
		"count_by_level":  countByLevel,
		"count_by_tenant": countByTenant,
		"min_level":       a.minLevel,
	}
}

// SetMinLevel sets minimum log level
func (a *LogAggregator) SetMinLevel(level LogLevel) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.minLevel = level
}

// Close closes the aggregator
func (a *LogAggregator) Close() {
	a.cancel()
	a.Flush()
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}