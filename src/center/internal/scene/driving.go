package scene

import (
	"context"
	"time"
)

// DrivingDetector detects driving scenes
type DrivingDetector struct{}

// GetName returns detector name
func (d *DrivingDetector) GetName() string {
	return "driving_detector"
}

// GetSupportedScenes returns supported scene types
func (d *DrivingDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneDriving}
}

// Detect detects driving scene from sensor data
func (d *DrivingDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}

	// Check activity type (most direct indicator)
	if activityType, ok := data["activity_type"].(string); ok {
		context["activity_type"] = activityType
		if activityType == "driving" || activityType == "in_vehicle" {
			confidence += 0.5
		}
	}

	// Check speed (vehicle speed typically > 20 km/h)
	if speed, ok := data["speed"].(float64); ok {
		context["speed"] = speed
		if speed > 20 && speed < 150 {
			confidence += 0.3
		}
		if speed > 5 && speed < 20 {
			confidence += 0.1 // Could be slow traffic
		}
	}

	// Check Bluetooth connection (car Bluetooth)
	if bluetoothDevices, ok := data["bluetooth_devices"].([]interface{}); ok {
		context["bluetooth_devices"] = bluetoothDevices
		for _, device := range bluetoothDevices {
			if dev, ok := device.(map[string]interface{}); ok {
				if name, ok := dev["name"].(string); ok {
					if containsCarKeywords(name) {
						confidence += 0.25
						context["car_connected"] = true
						break
					}
				}
				if devType, ok := dev["type"].(string); ok {
					if devType == "car_audio" || devType == "vehicle" {
						confidence += 0.25
						context["car_connected"] = true
						break
					}
				}
			}
		}
	}

	// Check location (on road)
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "road" || location == "highway" || location == "street" {
			confidence += 0.15
		}
	}

	// Check GPS coordinates changing rapidly
	if gpsMoving, ok := data["gps_moving"].(bool); ok {
		context["gps_moving"] = gpsMoving
		if gpsMoving {
			confidence += 0.1
		}
	}

	// Check audio environment (car noise level)
	if audioLevel, ok := data["audio_level"].(float64); ok {
		context["audio_level"] = audioLevel
		if audioLevel > 40 && audioLevel < 70 {
			confidence += 0.05 // Typical car interior noise
		}
	}

	// Check device type (watch less likely, phone more likely)
	if deviceType, ok := data["device_type"].(string); ok {
		context["device_type"] = deviceType
		if deviceType == "phone" {
			confidence += 0.1
		}
		if deviceType == "car_display" || deviceType == "navigation" {
			confidence += 0.3
		}
	}

	// Check navigation active
	if navigationActive, ok := data["navigation_active"].(bool); ok {
		context["navigation_active"] = navigationActive
		if navigationActive {
			confidence += 0.2
		}
	}

	// Check music playing (typical in car)
	if musicPlaying, ok := data["music_playing"].(bool); ok {
		context["music_playing"] = musicPlaying
		if musicPlaying {
			confidence += 0.05
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:        SceneDriving,
		AgentID:     agentID,
		Confidence:  min(0.95, confidence),
		Active:      true,
		Context:     context,
		TriggeredBy: "driving_detector",
	}

	// Define actions for driving scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_enable_car_mode",
			Type:        "mode",
			TargetAgent: agentID,
			Priority:    15,
			Payload:     map[string]interface{}{
				"mode":         "car",
				"auto_reply":   true,
				"voice_only":   true,
			},
		},
		{
			ID:          "action_route_to_car",
			Type:        "route",
			TargetAgent: "", // Car display device
			Priority:    10,
			Payload:     map[string]interface{}{
				"message_type": "driving_notification",
				"route_to":     "car_display",
				"fallback":     "phone_speaker",
			},
		},
		{
			ID:          "action_block_messages",
			Type:        "block",
			TargetAgent: "phone",
			Priority:    8,
			Payload:     map[string]interface{}{
				"block_type": "visual_notifications",
				"except":     []string{"navigation", "emergency"},
			},
		},
		{
			ID:          "action_voice_commands",
			Type:        "enable",
			TargetAgent: agentID,
			Priority:    12,
			Payload:     map[string]interface{}{
				"feature":      "voice_commands",
				"activation":   "auto",
				"hands_free":   true,
			},
		},
		{
			ID:          "action_call_handling",
			Type:        "config",
			TargetAgent: "phone",
			Priority:    10,
			Payload:     map[string]interface{}{
				"call_mode":    "auto_bluetooth",
				"reject_sms":   true,
				"auto_reply":   "正在驾驶，稍后回复",
			},
		},
	}

	return scene, nil
}

// DrivingHandler handles driving scene actions
type DrivingHandler struct{}

// GetName returns handler name
func (h *DrivingHandler) GetName() string {
	return "driving_handler"
}

// GetSupportedScenes returns supported scene types
func (h *DrivingHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneDriving}
}

// Handle handles driving scene
func (h *DrivingHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		switch action.Type {
		case "mode":
			// Enable car mode on device
		case "route":
			// Route notifications to car display
		case "block":
			// Block visual notifications on phone
		case "enable":
			// Enable voice commands
		case "config":
			// Configure call handling
		}
	}
	return nil
}

// DrivingSceneOrchestrator orchestrates driving scene across devices
type DrivingSceneOrchestrator struct {
	engine      *SceneEngine
	activeDrive *DrivingSession
	mu          sync.Mutex
}

// DrivingSession represents an active driving session
type DrivingSession struct {
	SessionID     string
	StartTime     time.Time
	EndTime       *time.Time
	Duration      time.Duration
	StartLocation string
	EndLocation   string
	Distance      float64 // km
	Route         []LocationPoint
	PhoneUsed     bool    // Did user use phone while driving
	VoiceCommands int     // Number of voice commands used
	MusicPlayed   bool
	NavigationUsed bool
	Alerts        []DrivingAlert
}

// LocationPoint represents a GPS coordinate
type LocationPoint struct {
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

// DrivingAlert represents an alert during driving
type DrivingAlert struct {
	Type        string // "speed_warning", "fatigue", "emergency_call"
	Timestamp   time.Time
	Message     string
	Severity    string // "low", "medium", "high"
	Handled     bool
}

// NewDrivingSceneOrchestrator creates a new driving orchestrator
func NewDrivingSceneOrchestrator(engine *SceneEngine) *DrivingSceneOrchestrator {
	return &DrivingSceneOrchestrator{
		engine: engine,
	}
}

// OnSceneStart handles driving scene start
func (o *DrivingSceneOrchestrator) OnSceneStart(scene *SceneState) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.activeDrive = &DrivingSession{
		SessionID:     generateSessionID(),
		StartTime:     scene.StartTime,
		StartLocation: getLocationFromContext(scene.Context),
	}
}

// OnSceneEnd handles driving scene end
func (o *DrivingSceneOrchestrator) OnSceneEnd(scene *SceneState) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.activeDrive != nil {
		now := time.Now()
		o.activeDrive.EndTime = &now
		o.activeDrive.Duration = now.Sub(o.activeDrive.StartTime)
		o.activeDrive.EndLocation = getLocationFromContext(scene.Context)

		// Calculate statistics
		// Store session to history
	}
}

// OnSceneAction handles driving scene action
func (o *DrivingSceneOrchestrator) OnSceneAction(scene *SceneState, action *SceneAction) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.activeDrive == nil {
		return
	}

	switch action.Type {
	case "voice_commands":
		o.activeDrive.VoiceCommands++
	case "navigation":
		o.activeDrive.NavigationUsed = true
	case "music":
		o.activeDrive.MusicPlayed = true
	case "alert":
		alert := DrivingAlert{
			Type:      action.Payload["alert_type"].(string),
			Timestamp: time.Now(),
			Message:   action.Payload["message"].(string),
			Severity:  action.Payload["severity"].(string),
		}
		o.activeDrive.Alerts = append(o.activeDrive.Alerts, alert)
	}
}

// GetDrivingSession returns current driving session
func (o *DrivingSceneOrchestrator) GetDrivingSession() *DrivingSession {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.activeDrive
}

// UpdateRoute updates driving route
func (o *DrivingSceneOrchestrator) UpdateRoute(point LocationPoint) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.activeDrive != nil {
		o.activeDrive.Route = append(o.activeDrive.Route, point)
	}
}

// Helper functions for driving detection

func containsCarKeywords(name string) bool {
	keywords := []string{"car", "auto", "vehicle", "BMW", "Mercedes", "Tesla", "Ford", "Honda", "Toyota", "Audi", "Volvo", "Nissan", "Bluetooth Audio", "车载", "汽车"}
	for _, kw := range keywords {
		if containsIgnoreCase(name, kw) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(sLower) >= len(substrLower) && sLower[:len(substrLower)] == substrLower
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

func getLocationFromContext(ctx map[string]interface{}) string {
	if location, ok := ctx["location"].(string); ok {
		return location
	}
	return ""
}

func generateSessionID() string {
	return "drive_" + time.Now().Format("20060102150405")
}

// init registers driving detector
func init() {
	// This will be called when package is imported
}