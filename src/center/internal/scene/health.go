package scene

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthAlertSceneConfig holds configuration for health alert scene
type HealthAlertSceneConfig struct {
	HighHeartRateThreshold   float64       `yaml:"high_heart_rate_threshold"`   // 高心率阈值
	LowHeartRateThreshold    float64       `yaml:"low_heart_rate_threshold"`    // 低心率阈值
	HighBloodPressureSystolic float64      `yaml:"high_bp_systolic"`            // 高血压收缩压阈值
	LowOxygenThreshold       float64       `yaml:"low_oxygen_threshold"`        // 低氧阈值
	HighTemperatureThreshold float64       `yaml:"high_temp_threshold"`         // 高温阈值
	BroadcastToAll           bool          `yaml:"broadcast_to_all"`            // 广播到所有设备
	LogToCenter              bool          `yaml:"log_to_center"`               // 记录到 Center
	NotifyEmergencyContact   bool          `yaml:"notify_emergency"`            // 通知紧急联系人
	AlertCooldown            time.Duration `yaml:"alert_cooldown"`              // 告警冷却时间
}

// DefaultHealthAlertSceneConfig returns default configuration
func DefaultHealthAlertSceneConfig() *HealthAlertSceneConfig {
	return &HealthAlertSceneConfig{
		HighHeartRateThreshold:   120,
		LowHeartRateThreshold:    50,
		HighBloodPressureSystolic: 140,
		LowOxygenThreshold:       95,
		HighTemperatureThreshold: 37.5,
		BroadcastToAll:           true,
		LogToCenter:              true,
		NotifyEmergencyContact:   true,
		AlertCooldown:            5 * time.Minute,
	}
}

// HealthAlertSceneOrchestrator orchestrates health alert scene across devices
type HealthAlertSceneOrchestrator struct {
	config        *HealthAlertSceneConfig
	engine        *SceneEngine
	activeAlerts  sync.Map // identityID -> *HealthAlertSession
	alertHistory  sync.Map // identityID -> []*HealthAlertSession
	eventHandler  HealthAlertEventHandler
	lastAlertTime sync.Map // identityID -> time.Time (for cooldown)
	mu            sync.RWMutex
}

// HealthAlertSession represents an active health alert session
type HealthAlertSession struct {
	IdentityID        string             `json:"identity_id"`
	WatchAgentID      string             `json:"watch_agent_id"`
	AlertType         string             `json:"alert_type"`        // high_heart_rate, low_heart_rate, low_oxygen, etc.
	Severity          string             `json:"severity"`          // high, medium, low
	StartTime         time.Time          `json:"start_time"`
	EndTime           *time.Time         `json:"end_time,omitempty"`
	Duration          time.Duration      `json:"duration"`
	Resolved          bool               `json:"resolved"`
	Value             float64            `json:"value"`             // 实际测量值
	Threshold         float64            `json:"threshold"`         // 阈值值
	Context           map[string]interface{} `json:"context"`       // 更多上下文数据
	BroadcastSent     bool               `json:"broadcast_sent"`
	CenterLogged      bool               `json:"center_logged"`
	EmergencyNotified bool               `json:"emergency_notified"`
	DevicesNotified   []string           `json:"devices_notified"`
}

// HealthAlertEventHandler handles health alert scene events
type HealthAlertEventHandler interface {
	OnHealthAlertStart(session *HealthAlertSession)
	OnHealthAlertEnd(session *HealthAlertSession)
	OnBroadcastAlert(session *HealthAlertSession, deviceIDs []string)
	OnCenterLog(session *HealthAlertSession)
	OnEmergencyNotify(session *HealthAlertSession, contact string)
	OnValueUpdate(identityID string, alertType string, value float64)
}

// NewHealthAlertSceneOrchestrator creates a new health alert scene orchestrator
func NewHealthAlertSceneOrchestrator(engine *SceneEngine, config *HealthAlertSceneConfig) *HealthAlertSceneOrchestrator {
	if config == nil {
		config = DefaultHealthAlertSceneConfig()
	}

	orch := &HealthAlertSceneOrchestrator{
		config: config,
		engine: engine,
	}

	// Register listener
	engine.AddListener(orch)

	return orch
}

// SetEventHandler sets the event handler
func (o *HealthAlertSceneOrchestrator) SetEventHandler(handler HealthAlertEventHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.eventHandler = handler
}

// HandleHealthAlertDetection processes health alert detection from watch
func (o *HealthAlertSceneOrchestrator) HandleHealthAlertDetection(ctx context.Context, identityID, watchAgentID string, data map[string]interface{}) error {
	// Check cooldown
	if !o.checkCooldown(identityID) {
		return nil // Skip due to cooldown
	}

	// Determine alert type and severity
	alertType, severity, value, threshold := o.analyzeHealthData(data)

	if alertType == "" {
		return nil // No alert condition
	}

	// Create alert session
	session := o.createAlertSession(identityID, watchAgentID, alertType, severity, value, threshold, data)

	// Store active alert
	o.activeAlerts.Store(identityID, session)

	// Broadcast to all devices
	if o.config.BroadcastToAll {
		o.broadcastAlert(session)
	}

	// Log to center
	if o.config.LogToCenter {
		o.logToCenter(session)
	}

	// Notify emergency contact
	if o.config.NotifyEmergencyContact && severity == "high" {
		o.notifyEmergency(session)
	}

	// Update last alert time
	o.lastAlertTime.Store(identityID, time.Now())

	return nil
}

// HandleHealthAlertResolved processes health alert resolution
func (o *HealthAlertSceneOrchestrator) HandleHealthAlertResolved(ctx context.Context, identityID string, resolution map[string]interface{}) error {
	session, ok := o.activeAlerts.Load(identityID)
	if !ok {
		return nil
	}

	alertSession := session.(*HealthAlertSession)

	// End alert
	now := time.Now()
	alertSession.EndTime = &now
	alertSession.Duration = now.Sub(alertSession.StartTime)
	alertSession.Resolved = true

	// End scene in engine
	o.engine.EndScene(identityID, SceneHealthAlert)

	// Notify end
	if o.eventHandler != nil {
		o.eventHandler.OnHealthAlertEnd(alertSession)
	}

	// Add to history
	o.addToHistory(identityID, alertSession)

	// Remove from active
	o.activeAlerts.Delete(identityID)

	return nil
}

// HandleHealthValueUpdate processes health value update during alert
func (o *HealthAlertSceneOrchestrator) HandleHealthValueUpdate(ctx context.Context, identityID string, dataType string, value float64) error {
	session, ok := o.activeAlerts.Load(identityID)
	if !ok {
		return nil // No active alert
	}

	alertSession := session.(*HealthAlertSession)

	// Update value
	alertSession.Value = value

	// Check if resolved
	if o.isValueResolved(alertSession.AlertType, value) {
		// Auto-resolve
		o.HandleHealthAlertResolved(ctx, identityID, map[string]interface{}{
			"auto_resolve": true,
			"value":        value,
		})
	}

	// Notify update
	if o.eventHandler != nil {
		o.eventHandler.OnValueUpdate(identityID, dataType, value)
	}

	return nil
}

// GetActiveAlert returns active alert session
func (o *HealthAlertSceneOrchestrator) GetActiveAlert(identityID string) *HealthAlertSession {
	session, ok := o.activeAlerts.Load(identityID)
	if !ok {
		return nil
	}
	return session.(*HealthAlertSession)
}

// GetAllActiveAlerts returns all active alert sessions
func (o *HealthAlertSceneOrchestrator) GetAllActiveAlerts() []*HealthAlertSession {
	var sessions []*HealthAlertSession
	o.activeAlerts.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*HealthAlertSession))
		return true
	})
	return sessions
}

// GetAlertHistory returns alert history for an identity
func (o *HealthAlertSceneOrchestrator) GetAlertHistory(identityID string, limit int) []*HealthAlertSession {
	history, ok := o.alertHistory.Load(identityID)
	if !ok {
		return []*HealthAlertSession{}
	}

	alertHistory := history.([]*HealthAlertSession)
	if limit > 0 && len(alertHistory) > limit {
		return alertHistory[:limit]
	}
	return alertHistory
}

// SceneListener implementation

func (o *HealthAlertSceneOrchestrator) OnSceneStart(scene *SceneState) {
	if scene.Type != SceneHealthAlert {
		return
	}

	// Notify start
	if o.eventHandler != nil {
		session := o.GetActiveAlert(scene.IdentityID)
		if session != nil {
			o.eventHandler.OnHealthAlertStart(session)
		}
	}
}

func (o *HealthAlertSceneOrchestrator) OnSceneEnd(scene *SceneState) {
	// Handled by HandleHealthAlertResolved
}

func (o *HealthAlertSceneOrchestrator) OnSceneAction(scene *SceneState, action *SceneAction) {
	if scene.Type != SceneHealthAlert {
		return
	}

	switch action.Type {
	case "alert":
		o.handleAlertAction(scene, action)
	case "notify":
		o.handleNotifyAction(scene, action)
	case "log":
		o.handleLogAction(scene, action)
	}
}

// Private methods

func (o *HealthAlertSceneOrchestrator) checkCooldown(identityID string) bool {
	lastTime, ok := o.lastAlertTime.Load(identityID)
	if !ok {
		return true // No previous alert
	}

	return time.Since(lastTime.(time.Time)) >= o.config.AlertCooldown
}

func (o *HealthAlertSceneOrchestrator) analyzeHealthData(data map[string]interface{}) (alertType, severity string, value, threshold float64) {
	// Check heart rate
	if heartRate, ok := data["heart_rate"].(float64); ok {
		if heartRate > o.config.HighHeartRateThreshold {
			return "high_heart_rate", "high", heartRate, o.config.HighHeartRateThreshold
		}
		if heartRate < o.config.LowHeartRateThreshold {
			return "low_heart_rate", "high", heartRate, o.config.LowHeartRateThreshold
		}
		if heartRate > 100 && heartRate <= o.config.HighHeartRateThreshold {
			return "elevated_heart_rate", "medium", heartRate, 100
		}
	}

	// Check oxygen level
	if oxygen, ok := data["oxygen_level"].(float64); ok {
		if oxygen < o.config.LowOxygenThreshold {
			return "low_oxygen", "high", oxygen, o.config.LowOxygenThreshold
		}
	}

	// Check temperature
	if temperature, ok := data["temperature"].(float64); ok {
		if temperature > o.config.HighTemperatureThreshold {
			return "high_temperature", "medium", temperature, o.config.HighTemperatureThreshold
		}
	}

	// Check blood pressure
	if bp, ok := data["blood_pressure"].(map[string]interface{}); ok {
		if systolic, ok := bp["systolic"].(float64); ok {
			if systolic > o.config.HighBloodPressureSystolic {
				return "high_blood_pressure", "high", systolic, o.config.HighBloodPressureSystolic
			}
		}
	}

	return "", "", 0, 0
}

func (o *HealthAlertSceneOrchestrator) createAlertSession(identityID, watchAgentID, alertType, severity string, value, threshold float64, data map[string]interface{}) *HealthAlertSession {
	return &HealthAlertSession{
		IdentityID:   identityID,
		WatchAgentID: watchAgentID,
		AlertType:    alertType,
		Severity:     severity,
		StartTime:    time.Now(),
		Value:        value,
		Threshold:    threshold,
		Context:      data,
		Resolved:     false,
	}
}

func (o *HealthAlertSceneOrchestrator) broadcastAlert(session *HealthAlertSession) {
	session.BroadcastSent = true

	// Get all devices for identity (integrate with device manager)
	var deviceIDs []string
	// deviceIDs = ... from device manager

	session.DevicesNotified = deviceIDs

	if o.eventHandler != nil {
		o.eventHandler.OnBroadcastAlert(session, deviceIDs)
	}
}

func (o *HealthAlertSceneOrchestrator) logToCenter(session *HealthAlertSession) {
	session.CenterLogged = true

	if o.eventHandler != nil {
		o.eventHandler.OnCenterLog(session)
	}
}

func (o *HealthAlertSceneOrchestrator) notifyEmergency(session *HealthAlertSession) {
	session.EmergencyNotified = true

	// Get emergency contact (integrate with identity profile)
	contact := "emergency_contact" // ... from identity profile

	if o.eventHandler != nil {
		o.eventHandler.OnEmergencyNotify(session, contact)
	}
}

func (o *HealthAlertSceneOrchestrator) isValueResolved(alertType string, value float64) bool {
	switch alertType {
	case "high_heart_rate":
		return value <= 100
	case "low_heart_rate":
		return value >= 60
	case "low_oxygen":
		return value >= 95
	case "high_temperature":
		return value <= 37.0
	case "high_blood_pressure":
		return value <= 130
	}
	return false
}

func (o *HealthAlertSceneOrchestrator) addToHistory(identityID string, session *HealthAlertSession) {
	var history []*HealthAlertSession
	if h, ok := o.alertHistory.Load(identityID); ok {
		history = h.([]*HealthAlertSession)
	}

	history = append(history, session)

	// Keep only last 50 alerts
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	o.alertHistory.Store(identityID, history)
}

func (o *HealthAlertSceneOrchestrator) handleAlertAction(scene *SceneState, action *SceneAction) {
	session := o.GetActiveAlert(scene.IdentityID)
	if session != nil && !session.BroadcastSent {
		o.broadcastAlert(session)
	}
}

func (o *HealthAlertSceneOrchestrator) handleNotifyAction(scene *SceneState, action *SceneAction) {
	// Notify center
	session := o.GetActiveAlert(scene.IdentityID)
	if session != nil && !session.CenterLogged {
		o.logToCenter(session)
	}
}

func (o *HealthAlertSceneOrchestrator) handleLogAction(scene *SceneState, action *SceneAction) {
	session := o.GetActiveAlert(scene.IdentityID)
	if session != nil {
		o.logToCenter(session)
	}
}

// GetStatistics returns health alert scene statistics
func (o *HealthAlertSceneOrchestrator) GetStatistics() map[string]interface{} {
	var activeCount int
	var historyCount int
	var severityCounts = map[string]int{
		"high":   0,
		"medium": 0,
		"low":    0,
	}

	o.activeAlerts.Range(func(key, value interface{}) bool {
		session := value.(*HealthAlertSession)
		if !session.Resolved {
			activeCount++
			severityCounts[session.Severity]++
		}
		return true
	})

	o.alertHistory.Range(func(key, value interface{}) bool {
		historyCount += len(value.([]*HealthAlertSession))
		return true
	})

	return map[string]interface{}{
		"active_alerts":  activeCount,
		"history_count":  historyCount,
		"severity_high":  severityCounts["high"],
		"severity_medium": severityCounts["medium"],
		"severity_low":   severityCounts["low"],
		"hr_high_threshold": o.config.HighHeartRateThreshold,
		"hr_low_threshold":  o.config.LowHeartRateThreshold,
		"oxygen_threshold":  o.config.LowOxygenThreshold,
		"broadcast_enabled": o.config.BroadcastToAll,
		"cooldown":         o.config.AlertCooldown,
	}
}