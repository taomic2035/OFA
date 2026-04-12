package scene

import (
	"context"
	"time"
)

// ExerciseDetector detects exercise/fitness scenes
type ExerciseDetector struct{}

// GetName returns detector name
func (d *ExerciseDetector) GetName() string {
	return "exercise_detector"
}

// GetSupportedScenes returns supported scene types
func (d *ExerciseDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneExercise}
}

// Detect detects exercise scene from fitness data
func (d *ExerciseDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}
	exerciseType := "general"

	// Check activity type
	if activityType, ok := data["activity_type"].(string); ok {
		context["activity_type"] = activityType
		switch activityType {
		case "walking", "跑步", "cycling", "swimming", "yoga", "gym", "workout", "hiking", "fitness":
			confidence += 0.5
			exerciseType = activityType
		case "running":
			confidence += 0.5
			exerciseType = "running"
		}
	}

	// Check heart rate (elevated during exercise)
	if heartRate, ok := data["heart_rate"].(float64); ok {
		context["heart_rate"] = heartRate
		if heartRate > 80 && heartRate < 180 {
			confidence += 0.25
		}
		if heartRate > 100 && heartRate < 160 {
			confidence += 0.1 // Optimal exercise range
		}
	}

	// Check calories burned
	if calories, ok := data["calories"].(float64); ok {
		context["calories"] = calories
		if calories > 50 {
			confidence += 0.1
		}
		if calories > 200 {
			confidence += 0.15
		}
	}

	// Check duration
	if duration, ok := data["duration"].(float64); ok {
		context["duration"] = duration
		if duration > 300 { // 5 minutes
			confidence += 0.1
		}
		if duration > 1200 { // 20 minutes
			confidence += 0.15
		}
	}

	// Check steps
	if steps, ok := data["steps"].(float64); ok {
		context["steps"] = steps
		if steps > 100 {
			confidence += 0.05
		}
	}

	// Check location (gym, park, outdoor)
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "gym" || location == "健身房" || location == "park" || location == "outdoor" {
			confidence += 0.2
		}
	}

	// Check device type (watch is primary for fitness)
	if deviceType, ok := data["device_type"].(string); ok {
		context["device_type"] = deviceType
		if deviceType == "watch" || deviceType == "fitness_tracker" {
			confidence += 0.15
		}
	}

	// Check fitness app active
	if fitnessAppActive, ok := data["fitness_app_active"].(bool); ok {
		context["fitness_app_active"] = fitnessAppActive
		if fitnessAppActive {
			confidence += 0.2
		}
	}

	// Check music playing (common during exercise)
	if musicPlaying, ok := data["music_playing"].(bool); ok {
		context["music_playing"] = musicPlaying
		if musicPlaying {
			confidence += 0.05
		}
	}

	// Check workout intensity
	if intensity, ok := data["intensity"].(string); ok {
		context["intensity"] = intensity
		switch intensity {
		case "high", "高强度":
			confidence += 0.15
		case "medium", "中等":
			confidence += 0.1
		case "low", "低强度":
			confidence += 0.05
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:        SceneExercise,
		AgentID:     agentID,
		Confidence:  min(0.95, confidence),
		Active:      true,
		Context:     context,
		TriggeredBy: "exercise_detector",
	}

	// Define actions for exercise scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_track_workout",
			Type:        "track",
			TargetAgent: "watch",
			Priority:    15,
			Payload:     map[string]interface{}{
				"workout_type": exerciseType,
				"auto_record":  true,
			},
		},
		{
			ID:          "action_route_music",
			Type:        "route",
			TargetAgent: "", // Device playing music
			Priority:    10,
			Payload:     map[string]interface{}{
				"feature":      "music_control",
				"voice_control": true,
			},
		},
		{
			ID:          "action_filter_calls",
			Type:        "filter",
			TargetAgent: "phone",
			Priority:    8,
			Payload:     map[string]interface{}{
				"filter_type":    "non_urgent",
				"auto_reply":     true,
				"reply_message":  "正在运动中，稍后回复",
			},
		},
		{
			ID:          "action_health_monitor",
			Type:        "monitor",
			TargetAgent: "watch",
			Priority:    12,
			Payload:     map[string]interface{}{
				"monitor_heart_rate":  true,
				"monitor_calories":    true,
				"alert_threshold":     180, // Max heart rate alert
			},
		},
		{
			ID:          "action_suggest_rest",
			Type:        "suggest",
			TargetAgent: "watch",
			Priority:    5,
			Payload:     map[string]interface{}{
				"suggest_type":      "rest_interval",
				"check_fatigue":     true,
			},
		},
	}

	return scene, nil
}

// ExerciseHandler handles exercise scene actions
type ExerciseHandler struct{}

// GetName returns handler name
func (h *ExerciseHandler) GetName() string {
	return "exercise_handler"
}

// GetSupportedScenes returns supported scene types
func (h *ExerciseHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneExercise}
}

// Handle handles exercise scene
func (h *ExerciseHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		switch action.Type {
		case "track":
			// Start workout tracking
		case "route":
			// Configure music routing
		case "filter":
			// Filter non-urgent calls
		case "monitor":
			// Enable health monitoring
		case "suggest":
			// Suggest rest intervals
		}
	}
	return nil
}

// ExerciseSession represents an exercise session
type ExerciseSession struct {
	SessionID       string
	StartTime       time.Time
	EndTime         *time.Time
	Duration        time.Duration
	ExerciseType    string
	CaloriesBurned  float64
	AvgHeartRate    float64
	MaxHeartRate    float64
	Steps           int
	Distance        float64 // km
	Intensity       string
	DeviceUsed      string
	Metrics         []ExerciseMetric
	Alerts          []ExerciseAlert
}

// ExerciseMetric represents a metric snapshot during exercise
type ExerciseMetric struct {
	Timestamp   time.Time
	HeartRate   int
	Calories    float64
	Steps       int
	Distance    float64
}

// ExerciseAlert represents an alert during exercise
type ExerciseAlert struct {
	Type        string // "heart_rate_high", "dehydration", "fatigue", "goal_reached"
	Timestamp   time.Time
	Value       float64
	Message     string
}

// SleepDetector detects sleep/rest scenes
type SleepDetector struct{}

// GetName returns detector name
func (d *SleepDetector) GetName() string {
	return "sleep_detector"
}

// GetSupportedScenes returns supported scene types
func (d *SleepDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneSleeping}
}

// Detect detects sleep scene
func (d *SleepDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}

	// Check activity type
	if activityType, ok := data["activity_type"].(string); ok {
		context["activity_type"] = activityType
		if activityType == "sleep" || activityType == "sleeping" || activityType == "rest" {
			confidence += 0.5
		}
	}

	// Check time of day (night time sleep)
	hour := time.Now().Hour()
	context["hour"] = hour
	if hour >= 22 || hour <= 6 {
		confidence += 0.2
	} else if hour >= 13 && hour <= 15 {
		confidence += 0.1 // Nap time
	}

	// Check heart rate (low during sleep)
	if heartRate, ok := data["heart_rate"].(float64); ok {
		context["heart_rate"] = heartRate
		if heartRate < 70 && heartRate > 40 {
			confidence += 0.25
		}
		if heartRate < 60 {
			confidence += 0.15
		}
	}

	// Check movement (minimal during sleep)
	if movement, ok := data["movement"].(float64); ok {
		context["movement"] = movement
		if movement < 10 {
			confidence += 0.3
		}
	}

	// Check light level (dark during sleep)
	if lightLevel, ok := data["light_level"].(float64); ok {
		context["light_level"] = lightLevel
		if lightLevel < 10 {
			confidence += 0.15
		}
	}

	// Check device usage (no usage during sleep)
	if deviceUsage, ok := data["device_usage"].(bool); ok {
		context["device_usage"] = deviceUsage
		if !deviceUsage {
			confidence += 0.1
		}
	}

	// Check location (bedroom)
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "bedroom" || location == "卧室" || location == "home" {
			confidence += 0.1
		}
	}

	// Check screen off
	if screenOff, ok := data["screen_off"].(bool); ok {
		context["screen_off"] = screenOff
		if screenOff {
			confidence += 0.15
		}
	}

	// Check device position (flat, not moving)
	if devicePosition, ok := data["device_position"].(string); ok {
		context["device_position"] = devicePosition
		if devicePosition == "flat" || devicePosition == "still" {
			confidence += 0.1
		}
	}

	// Check breathing rate
	if breathingRate, ok := data["breathing_rate"].(float64); ok {
		context["breathing_rate"] = breathingRate
		if breathingRate < 20 && breathingRate > 10 {
			confidence += 0.1
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:        SceneSleeping,
		AgentID:     agentID,
		Confidence:  min(0.95, confidence),
		Active:      true,
		Context:     context,
		TriggeredBy: "sleep_detector",
	}

	// Define actions for sleep scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_dnd_mode",
			Type:        "dnd",
			TargetAgent: "phone",
			Priority:    15,
			Payload:     map[string]interface{}{
				"dnd_level":    "silent",
				"except":       []string{"alarm", "emergency"},
			},
		},
		{
			ID:          "action_block_all",
			Type:        "block",
			TargetAgent: "*",
			Priority:    12,
			Payload:     map[string]interface{}{
				"block_type":   "all_notifications",
				"except":       []string{"alarm", "emergency"},
			},
		},
		{
			ID:          "action_track_sleep",
			Type:        "track",
			TargetAgent: "watch",
			Priority:    10,
			Payload:     map[string]interface{}{
				"track_sleep":       true,
				"track_heart_rate":  true,
				"track_breathing":   true,
			},
		},
		{
			ID:          "action_adjust_alarm",
			Type:        "config",
			TargetAgent: "phone",
			Priority:    8,
			Payload:     map[string]interface{}{
				"smart_alarm":   true,
				"wake_range":    30, // minutes before alarm
			},
		},
		{
			ID:          "action_health_check",
			Type:        "monitor",
			TargetAgent: "watch",
			Priority:    5,
			Payload:     map[string]interface{}{
				"check_apnea":      true,
				"check_irregular":  true,
			},
		},
	}

	return scene, nil
}

// SleepHandler handles sleep scene actions
type SleepHandler struct{}

// GetName returns handler name
func (h *SleepHandler) GetName() string {
	return "sleep_handler"
}

// GetSupportedScenes returns supported scene types
func (h *SleepHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneSleeping}
}

// Handle handles sleep scene
func (h *SleepHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		switch action.Type {
		case "dnd":
			// Enable do not disturb
		case "block":
			// Block all notifications
		case "track":
			// Start sleep tracking
		case "config":
			// Configure smart alarm
		case "monitor":
			// Monitor sleep health
		}
	}
	return nil
}

// SleepSession represents a sleep session
type SleepSession struct {
	SessionID       string
	StartTime       time.Time
	EndTime         *time.Time
	Duration        time.Duration
	SleepQuality    float64 // 0-100
	DeepSleep       time.Duration
	LightSleep      time.Duration
	REMSleep        time.Duration
	AvgHeartRate    float64
	MinHeartRate    float64
	BreathingRate   float64
	MovementCount   int
	ApneaEvents     int
	Awakenings      []SleepAwakening
}

// SleepAwakening represents an awakening during sleep
type SleepAwakening struct {
	Timestamp   time.Time
	Duration    time.Duration
	Reason      string // "noise", "movement", "unknown"
}