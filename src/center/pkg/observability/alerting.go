// Package observability provides alerting capabilities
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// AlertSeverity defines alert severity levels
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
	SeverityEmergency AlertSeverity = "emergency"
)

// AlertState defines alert states
type AlertState string

const (
	AlertFiring   AlertState = "firing"
	AlertResolved AlertState = "resolved"
	AlertSilenced AlertState = "silenced"
)

// Alert represents an alert
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Severity    AlertSeverity     `json:"severity"`
	State       AlertState        `json:"state"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	TenantID    string            `json:"tenant_id,omitempty"`
	FiredAt     time.Time         `json:"fired_at"`
	ResolvedAt  time.Time         `json:"resolved_at,omitempty"`
	LastEvalAt  time.Time         `json:"last_eval_at"`
	SilencedBy  string            `json:"silenced_by,omitempty"`
}

// AlertRule defines an alert rule
type AlertRule struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Expr         string        `json:"expr"`          // Expression to evaluate
	For          time.Duration `json:"for"`           // Duration before firing
	Severity     AlertSeverity `json:"severity"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	Enabled      bool          `json:"enabled"`
	TenantID     string        `json:"tenant_id,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// AlertRuleState tracks rule evaluation state
type AlertRuleState struct {
	Rule        *AlertRule
	PendingAt   time.Time
	FiringSince time.Time
	LastValue   float64
	Alerts      []*Alert
}

// AlertManager manages alerts and alert rules
type AlertManager struct {
	// Rules and alerts
	rules      sync.Map // map[string]*AlertRule
	ruleStates sync.Map // map[string]*AlertRuleState
	alerts     sync.Map // map[string]*Alert

	// Notification channels
	channels sync.Map // map[string]NotificationChannel

	// Silences
	silences sync.Map // map[string]*Silence

	// Metrics
	totalAlerts     int64
	firingAlerts    int64
	resolvedAlerts  int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NotificationChannel defines notification delivery
type NotificationChannel interface {
	Send(alert *Alert) error
	Name() string
}

// Silence represents an alert silence
type Silence struct {
	ID        string            `json:"id"`
	Matchers  map[string]string `json:"matchers"`  // Labels to match
	StartsAt  time.Time         `json:"starts_at"`
	EndsAt    time.Time         `json:"ends_at"`
	CreatedBy string            `json:"created_by"`
	Comment   string            `json:"comment"`
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &AlertManager{
		ctx:    ctx,
		cancel: cancel,
	}

	// Start evaluator
	go manager.evaluator()

	// Start silence cleaner
	go manager.silenceCleaner()

	return manager
}

// CreateRule creates an alert rule
func (m *AlertManager) CreateRule(rule *AlertRule) error {
	if rule.ID == "" {
		rule.ID = generateAlertRuleID(rule.Name)
	}

	rule.CreatedAt = time.Now()
	rule.Enabled = true

	m.rules.Store(rule.ID, rule)

	// Initialize rule state
	m.ruleStates.Store(rule.ID, &AlertRuleState{
		Rule:   rule,
		Alerts: make([]*Alert, 0),
	})

	return nil
}

// GetRule retrieves an alert rule
func (m *AlertManager) GetRule(ruleID string) (*AlertRule, error) {
	if v, ok := m.rules.Load(ruleID); ok {
		return v.(*AlertRule), nil
	}
	return nil, fmt.Errorf("rule not found: %s", ruleID)
}

// ListRules lists all alert rules
func (m *AlertManager) ListRules() []*AlertRule {
	var rules []*AlertRule
	m.rules.Range(func(key, value interface{}) bool {
		rules = append(rules, value.(*AlertRule))
		return true
	})
	return rules
}

// EnableRule enables an alert rule
func (m *AlertManager) EnableRule(ruleID string) error {
	rule, err := m.GetRule(ruleID)
	if err != nil {
		return err
	}

	rule.Enabled = true
	return nil
}

// DisableRule disables an alert rule
func (m *AlertManager) DisableRule(ruleID string) error {
	rule, err := m.GetRule(ruleID)
	if err != nil {
		return err
	}

	rule.Enabled = false
	return nil
}

// DeleteRule deletes an alert rule
func (m *AlertManager) DeleteRule(ruleID string) error {
	m.rules.Delete(ruleID)
	m.ruleStates.Delete(ruleID)
	return nil
}

// Evaluate evaluates an alert expression
func (m *AlertManager) Evaluate(expr string) (float64, error) {
	// Simplified expression evaluation
	// In production, would use proper expression parser
	switch expr {
	case "agent_count":
		return float64(10), nil
	case "task_success_rate":
		return 0.95, nil
	case "error_rate":
		return 0.02, nil
	case "cpu_usage":
		return 65.0, nil
	case "memory_usage":
		return 72.0, nil
	case "response_time":
		return 150.0, nil
	default:
		return 0, nil
	}
}

// evaluator periodically evaluates alert rules
func (m *AlertManager) evaluator() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.evaluateRules()
		}
	}
}

// evaluateRules evaluates all enabled rules
func (m *AlertManager) evaluateRules() {
	m.rules.Range(func(key, value interface{}) bool {
		rule := value.(*AlertRule)

		if !rule.Enabled {
			return true
		}

		state, ok := m.ruleStates.Load(rule.ID)
		if !ok {
			return true
		}

		ruleState := state.(*AlertRuleState)

		// Evaluate expression
		value, err := m.Evaluate(rule.Expr)
		if err != nil {
			log.Printf("Error evaluating rule %s: %v", rule.ID, err)
			return true
		}

		ruleState.LastValue = value
		rule.LastEvalAt = time.Now()

		// Check threshold
		firing := m.checkThreshold(rule, value)

		if firing {
			m.handleFiring(ruleState, value)
		} else {
			m.handleResolved(ruleState)
		}

		return true
	})
}

// checkThreshold checks if value crosses threshold
func (m *AlertManager) checkThreshold(rule *AlertRule, value float64) bool {
	// Parse expression for threshold
	// Simplified: assume threshold is 0.9 for rates, 90 for percentages
	threshold := 0.9
	if rule.Severity == SeverityCritical {
		threshold = 0.95
	}

	// Check if expression contains comparison
	if contains(rule.Expr, "rate") {
		return value < threshold
	}
	if contains(rule.Expr, "usage") || contains(rule.Expr, "time") {
		return value > threshold*100
	}

	return value > threshold
}

// handleFiring handles a firing alert
func (m *AlertManager) handleFiring(state *AlertRuleState, value float64) {
	rule := state.Rule

	// Check silence
	if m.isSilenced(rule) {
		return
	}

	if state.FiringSince.IsZero() {
		// Start pending period
		state.PendingAt = time.Now()
	}

	// Check if duration threshold met
	if time.Since(state.PendingAt) >= rule.For {
		if state.FiringSince.IsZero() {
			// Fire alert
			state.FiringSince = time.Now()

			alert := &Alert{
				ID:          generateAlertID(rule.Name),
				Name:        rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				State:       AlertFiring,
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
				Value:       value,
				Threshold:   0.9,
				TenantID:    rule.TenantID,
				FiredAt:     time.Now(),
				LastEvalAt:  time.Now(),
			}

			m.alerts.Store(alert.ID, alert)
			state.Alerts = append(state.Alerts, alert)
			m.totalAlerts++
			m.firingAlerts++

			// Send notifications
			m.sendNotifications(alert)

			log.Printf("Alert firing: %s (value=%.2f)", rule.Name, value)
		} else {
			// Update existing alert
			if len(state.Alerts) > 0 {
				alert := state.Alerts[len(state.Alerts)-1]
				alert.Value = value
				alert.LastEvalAt = time.Now()
			}
		}
	}
}

// handleResolved handles a resolved alert
func (m *AlertManager) handleResolved(state *AlertRuleState) {
	if state.FiringSince.IsZero() {
		// Not firing, nothing to resolve
		return
	}

	// Resolve the alert
	if len(state.Alerts) > 0 {
		alert := state.Alerts[len(state.Alerts)-1]
		alert.State = AlertResolved
		alert.ResolvedAt = time.Now()
		alert.LastEvalAt = time.Now()

		m.resolvedAlerts++
		m.firingAlerts--

		// Send resolution notification
		m.sendNotifications(alert)

		log.Printf("Alert resolved: %s", alert.Name)
	}

	// Reset firing state
	state.FiringSince = time.Time{}
	state.PendingAt = time.Time{}
}

// isSilenced checks if rule is silenced
func (m *AlertManager) isSilenced(rule *AlertRule) bool {
	var silenced bool

	m.silences.Range(func(key, value interface{}) bool {
		silence := value.(*Silence)

		// Check if silence is active
		now := time.Now()
		if now.Before(silence.StartsAt) || now.After(silence.EndsAt) {
			return true
		}

		// Check matchers
		for k, v := range silence.Matchers {
			if rule.Labels[k] != v {
				return true
			}
		}

		silenced = true
		return false
	})

	return silenced
}

// CreateSilence creates a silence
func (m *AlertManager) CreateSilence(silence *Silence) error {
	if silence.ID == "" {
		silence.ID = generateSilenceID()
	}

	m.silences.Store(silence.ID, silence)
	return nil
}

// DeleteSilence deletes a silence
func (m *AlertManager) DeleteSilence(silenceID string) error {
	m.silences.Delete(silenceID)
	return nil
}

// ListSilences lists all silences
func (m *AlertManager) ListSilences() []*Silence {
	var silences []*Silence
	m.silences.Range(func(key, value interface{}) bool {
		s := value.(*Silence)
		// Only return active silences
		now := time.Now()
		if now.After(s.StartsAt) && now.Before(s.EndsAt) {
			silences = append(silences, s)
		}
		return true
	})
	return silences
}

// RegisterChannel registers a notification channel
func (m *AlertManager) RegisterChannel(channel NotificationChannel) {
	m.channels.Store(channel.Name(), channel)
}

// sendNotifications sends alert to all channels
func (m *AlertManager) sendNotifications(alert *Alert) {
	m.channels.Range(func(key, value interface{}) bool {
		channel := value.(NotificationChannel)
		if err := channel.Send(alert); err != nil {
			log.Printf("Failed to send alert to %s: %v", channel.Name(), err)
		}
		return true
	})
}

// GetAlert retrieves an alert
func (m *AlertManager) GetAlert(alertID string) (*Alert, error) {
	if v, ok := m.alerts.Load(alertID); ok {
		return v.(*Alert), nil
	}
	return nil, fmt.Errorf("alert not found: %s", alertID)
}

// ListAlerts lists alerts
func (m *AlertManager) ListAlerts(state AlertState, tenantID string) []*Alert {
	var alerts []*Alert

	m.alerts.Range(func(key, value interface{}) bool {
		alert := value.(*Alert)

		if state != "" && alert.State != state {
			return true
		}
		if tenantID != "" && alert.TenantID != tenantID {
			return true
		}

		alerts = append(alerts, alert)
		return true
	})

	return alerts
}

// silenceCleaner periodically cleans up expired silences
func (m *AlertManager) silenceCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanSilences()
		}
	}
}

// cleanSilences removes expired silences
func (m *AlertManager) cleanSilences() {
	now := time.Now()

	m.silences.Range(func(key, value interface{}) bool {
		silence := value.(*Silence)
		if now.After(silence.EndsAt) {
			m.silences.Delete(key)
		}
		return true
	})
}

// GetStats returns alert manager statistics
func (m *AlertManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_alerts":    m.totalAlerts,
		"firing_alerts":   m.firingAlerts,
		"resolved_alerts": m.resolvedAlerts,
		"rules_count":     len(m.ListRules()),
		"silences_count":  len(m.ListSilences()),
	}
}

// Close closes the alert manager
func (m *AlertManager) Close() {
	m.cancel()
}

// DefaultAlertRules returns default alert rules
func DefaultAlertRules() []*AlertRule {
	return []*AlertRule{
		{
			Name:        "HighCPUUsage",
			Description: "CPU usage is above 90%",
			Expr:        "cpu_usage",
			For:         5 * time.Minute,
			Severity:    SeverityWarning,
			Labels:      map[string]string{"category": "system"},
		},
		{
			Name:        "HighMemoryUsage",
			Description: "Memory usage is above 90%",
			Expr:        "memory_usage",
			For:         5 * time.Minute,
			Severity:    SeverityWarning,
			Labels:      map[string]string{"category": "system"},
		},
		{
			Name:        "HighErrorRate",
			Description: "Error rate is above 5%",
			Expr:        "error_rate",
			For:         1 * time.Minute,
			Severity:    SeverityCritical,
			Labels:      map[string]string{"category": "application"},
		},
		{
			Name:        "LowSuccessRate",
			Description: "Task success rate is below 90%",
			Expr:        "task_success_rate",
			For:         5 * time.Minute,
			Severity:    SeverityWarning,
			Labels:      map[string]string{"category": "application"},
		},
		{
			Name:        "HighResponseTime",
			Description: "Response time is above 500ms",
			Expr:        "response_time",
			For:         3 * time.Minute,
			Severity:    SeverityWarning,
			Labels:      map[string]string{"category": "performance"},
		},
	}
}

// Helper functions

func generateAlertRuleID(name string) string {
	return fmt.Sprintf("rule-%s-%d", name, time.Now().UnixNano())
}

func generateAlertID(name string) string {
	return fmt.Sprintf("alert-%s-%d", name, time.Now().UnixNano())
}

func generateSilenceID() string {
	return fmt.Sprintf("silence-%d", time.Now().UnixNano())
}

// SlackNotifier sends alerts to Slack
type SlackNotifier struct {
	WebhookURL string
	Channel    string
}

// Send implements NotificationChannel
func (n *SlackNotifier) Send(alert *Alert) error {
	// Placeholder for Slack notification
	return nil
}

// Name implements NotificationChannel
func (n *SlackNotifier) Name() string {
	return "slack"
}

// EmailNotifier sends alerts via email
type EmailNotifier struct {
	SMTPServer string
	From       string
	To         []string
}

// Send implements NotificationChannel
func (n *EmailNotifier) Send(alert *Alert) error {
	// Placeholder for email notification
	return nil
}

// Name implements NotificationChannel
func (n *EmailNotifier) Name() string {
	return "email"
}

// WebhookNotifier sends alerts to webhook
type WebhookNotifier struct {
	URL    string
	secret string
}

// Send implements NotificationChannel
func (n *WebhookNotifier) Send(alert *Alert) error {
	// Placeholder for webhook notification
	data, _ := json.Marshal(alert)
	_ = data
	return nil
}

// Name implements NotificationChannel
func (n *WebhookNotifier) Name() string {
	return "webhook"
}

// PagerDutyNotifier sends alerts to PagerDuty
type PagerDutyNotifier struct {
	ServiceKey string
}

// Send implements NotificationChannel
func (n *PagerDutyNotifier) Send(alert *Alert) error {
	// Placeholder for PagerDuty notification
	return nil
}

// Name implements NotificationChannel
func (n *PagerDutyNotifier) Name() string {
	return "pagerduty"
}

// Math helper
func init() {
	_ = math.MaxFloat64
}