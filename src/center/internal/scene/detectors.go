package scene

import (
	"context"
	"time"
)

// === Scene Detectors ===

// RunningDetector detects running scenes
type RunningDetector struct{}

// GetName returns detector name
func (d *RunningDetector) GetName() string {
	return "running_detector"
}

// GetSupportedScenes returns supported scene types
func (d *RunningDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneRunning}
}

// Detect detects running scene
func (d *RunningDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	// Check activity type
	activityType, ok := data["activity_type"].(string)
	if !ok {
		return nil, nil
	}

	if activityType != "running" {
		return nil, nil
	}

	// Get additional context
	confidence := 0.7
	context := map[string]interface{}{
		"activity_type": activityType,
		"agent_id":      agentID,
	}

	// Check duration
	if duration, ok := data["duration"].(float64); ok {
		context["duration"] = duration
		if duration > 60 {
			confidence = min(0.95, confidence + 0.2)
		}
	}

	// Check heart rate (elevated during running)
	if heartRate, ok := data["heart_rate"].(float64); ok {
		context["heart_rate"] = heartRate
		if heartRate > 100 && heartRate < 180 {
			confidence = min(0.95, confidence + 0.1)
		}
	}

	// Check step count
	if steps, ok := data["steps"].(float64); ok {
		context["steps"] = steps
		if steps > 0 {
			confidence = min(0.95, confidence + 0.05)
		}
	}

	// Check location (outdoor)
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "outdoor" {
			confidence = min(0.95, confidence + 0.1)
		}
	}

	scene := &SceneState{
		Type:       SceneRunning,
		AgentID:    agentID,
		Confidence: confidence,
		Active:     true,
		Context:    context,
		TriggeredBy: "running_detector",
	}

	// Define actions for running scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_route_to_phone",
			Type:        "route",
			TargetAgent: "", // Will be determined by router
			Priority:    10,
			Payload:     map[string]interface{}{
				"message_type": "running_notification",
				"route_to":     "phone",
			},
		},
		{
			ID:          "action_filter_messages",
			Type:        "filter",
			TargetAgent: "watch",
			Priority:    5,
			Payload:     map[string]interface{}{
				"filter_type": "urgent_only",
			},
		},
	}

	return scene, nil
}

// MeetingDetector detects meeting scenes
type MeetingDetector struct{}

// GetName returns detector name
func (d *MeetingDetector) GetName() string {
	return "meeting_detector"
}

// GetSupportedScenes returns supported scene types
func (d *MeetingDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneMeeting}
}

// Detect detects meeting scene
func (d *MeetingDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}

	// Check calendar event
	if calendarEvent, ok := data["calendar_event"].(string); ok {
		context["calendar_event"] = calendarEvent
		if containsMeetingKeywords(calendarEvent) {
			confidence += 0.4
		}
	}

	// Check location
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "会议室" || location == "办公室" || location == "meeting_room" {
			confidence += 0.3
		}
	}

	// Check audio environment (quiet)
	if audioLevel, ok := data["audio_level"].(float64); ok {
		context["audio_level"] = audioLevel
		if audioLevel < 30 { // Quiet environment
			confidence += 0.2
		}
	}

	// Check device type (likely phone in meeting)
	if deviceType, ok := data["device_type"].(string); ok {
		context["device_type"] = deviceType
		if deviceType == "phone" || deviceType == "tablet" {
			confidence += 0.1
		}
	}

	// Check time (work hours)
	if timeContext, ok := data["time"].(string); ok {
		context["time"] = timeContext
		if isWorkHours() {
			confidence += 0.1
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:       SceneMeeting,
		AgentID:    agentID,
		Confidence: min(0.95, confidence),
		Active:     true,
		Context:    context,
		TriggeredBy: "meeting_detector",
	}

	// Define actions for meeting scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_dnd_mode",
			Type:        "dnd",
			TargetAgent: agentID,
			Priority:    10,
			Payload:     map[string]interface{}{
				"dnd_duration": "meeting_end",
			},
		},
		{
			ID:          "action_notify_glasses",
			Type:        "notify",
			TargetAgent: "", // Glasses device
			Priority:    5,
			Payload:     map[string]interface{}{
				"notification_type": "meeting_reminder",
				"silent":            true,
			},
		},
		{
			ID:          "action_block_calls",
			Type:        "block",
			TargetAgent: "phone",
			Priority:    8,
			Payload:     map[string]interface{}{
				"block_type": "calls",
				"except":     []string{"urgent"},
			},
		},
	}

	return scene, nil
}

// HealthAlertDetector detects health alert scenes
type HealthAlertDetector struct{}

// GetName returns detector name
func (d *HealthAlertDetector) GetName() string {
	return "health_alert_detector"
}

// GetSupportedScenes returns supported scene types
func (d *HealthAlertDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneHealthAlert}
}

// Detect detects health alert scene
func (d *HealthAlertDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}
	alertType := ""

	// Check heart rate
	if heartRate, ok := data["heart_rate"].(float64); ok {
		context["heart_rate"] = heartRate

		// Abnormal high heart rate
		if heartRate > 120 {
			confidence = 0.9
			alertType = "high_heart_rate"
			context["alert_type"] = alertType
			context["severity"] = "high"
		}

		// Abnormal low heart rate
		if heartRate < 50 {
			confidence = 0.9
			alertType = "low_heart_rate"
			context["alert_type"] = alertType
			context["severity"] = "high"
		}

		// Moderate concern
		if heartRate > 100 && heartRate <= 120 {
			confidence = 0.6
			alertType = "elevated_heart_rate"
			context["alert_type"] = alertType
			context["severity"] = "medium"
		}
	}

	// Check blood pressure
	if bp, ok := data["blood_pressure"].(map[string]interface{}); ok {
		context["blood_pressure"] = bp
		if systolic, ok := bp["systolic"].(float64); ok {
			if systolic > 140 {
				confidence = max(confidence, 0.8)
				alertType = "high_blood_pressure"
				context["alert_type"] = alertType
				context["severity"] = "high"
			}
		}
	}

	// Check temperature
	if temperature, ok := data["temperature"].(float64); ok {
		context["temperature"] = temperature
		if temperature > 37.5 {
			confidence = max(confidence, 0.7)
			alertType = "high_temperature"
			context["alert_type"] = alertType
			context["severity"] = "medium"
		}
	}

	// Check oxygen level
	if oxygen, ok := data["oxygen_level"].(float64); ok {
		context["oxygen_level"] = oxygen
		if oxygen < 95 {
			confidence = max(confidence, 0.85)
			alertType = "low_oxygen"
			context["alert_type"] = alertType
			context["severity"] = "high"
		}
	}

	// Check stress level
	if stress, ok := data["stress_level"].(float64); ok {
		context["stress_level"] = stress
		if stress > 80 {
			confidence = max(confidence, 0.6)
			alertType = "high_stress"
			context["alert_type"] = alertType
			context["severity"] = "medium"
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:       SceneHealthAlert,
		AgentID:    agentID,
		Confidence: confidence,
		Active:     true,
		Context:    context,
		TriggeredBy: "health_alert_detector",
	}

	// Define actions for health alert scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_broadcast_alert",
			Type:        "alert",
			TargetAgent: "*", // Broadcast to all devices
			Priority:    15,
			Payload:     map[string]interface{}{
				"alert_type":    alertType,
				"severity":      context["severity"],
				"broadcast":     true,
			},
		},
		{
			ID:          "action_notify_center",
			Type:        "notify",
			TargetAgent: "center",
			Priority:    10,
			Payload:     map[string]interface{}{
				"notification_type": "health_alert",
				"data":              context,
			},
		},
		{
			ID:          "action_log_event",
			Type:        "log",
			TargetAgent: "center",
			Priority:    5,
			Payload:     map[string]interface{}{
				"event_type": "health_alert",
				"timestamp":  time.Now(),
				"data":       context,
			},
		},
	}

	return scene, nil
}

// === Scene Handlers ===

// NotificationHandler handles notification actions
type NotificationHandler struct{}

// GetName returns handler name
func (h *NotificationHandler) GetName() string {
	return "notification_handler"
}

// GetSupportedScenes returns supported scene types
func (h *NotificationHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneRunning, SceneMeeting, SceneHealthAlert}
}

// Handle handles scene notifications
func (h *NotificationHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		if action.Type == "notify" || action.Type == "alert" {
			// Execute notification
			// This will be integrated with WebSocket broadcaster
		}
	}
	return nil
}

// RoutingHandler handles routing actions
type RoutingHandler struct{}

// GetName returns handler name
func (h *RoutingHandler) GetName() string {
	return "routing_handler"
}

// GetSupportedScenes returns supported scene types
func (h *RoutingHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneRunning, SceneMeeting}
}

// Handle handles scene routing
func (h *RoutingHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		if action.Type == "route" || action.Type == "filter" {
			// Execute routing
			// This will be integrated with CrossDeviceRouter
		}
	}
	return nil
}

// AlertHandler handles alert actions
type AlertHandler struct{}

// GetName returns handler name
func (h *AlertHandler) GetName() string {
	return "alert_handler"
}

// GetSupportedScenes returns supported scene types
func (h *AlertHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneHealthAlert}
}

// Handle handles scene alerts
func (h *AlertHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		if action.Type == "alert" {
			// Execute alert broadcast
			// This will be integrated with WebSocket broadcaster
		}
	}
	return nil
}

// === Helper Functions ===

func containsMeetingKeywords(text string) bool {
	keywords := []string{"会议", "meeting", "讨论", "discussion", "汇报", "report", "面试", "interview"}
	for _, kw := range keywords {
		if containsString(text, kw) {
			return true
		}
	}
	return false
}

func isWorkHours() bool {
	now := time.Now()
	hour := now.Hour()
	// Work hours: 9:00 - 18:00
	return hour >= 9 && hour <= 18
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}