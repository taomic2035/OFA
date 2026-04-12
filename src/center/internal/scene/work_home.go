package scene

import (
	"context"
	"time"
)

// WorkDetector detects work/office scenes
type WorkDetector struct{}

// GetName returns detector name
func (d *WorkDetector) GetName() string {
	return "work_detector"
}

// GetSupportedScenes returns supported scene types
func (d *WorkDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneWork}
}

// Detect detects work scene from context
func (d *WorkDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}

	// Check location
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "office" || location == "办公室" || location == "work" || location == "公司" {
			confidence += 0.4
		}
	}

	// Check time (work hours)
	hour := time.Now().Hour()
	context["hour"] = hour
	if hour >= 9 && hour <= 18 {
		confidence += 0.2
	}
	if hour >= 9 && hour <= 12 {
		confidence += 0.1 // Morning work
	}
	if hour >= 14 && hour <= 17 {
		confidence += 0.1 // Afternoon work
	}

	// Check weekday
	weekday := time.Now().Weekday()
	context["weekday"] = weekday
	if weekday >= time.Monday && weekday <= time.Friday {
		confidence += 0.1
	}

	// Check calendar event (work related)
	if calendarEvent, ok := data["calendar_event"].(string); ok {
		context["calendar_event"] = calendarEvent
		if containsWorkKeywords(calendarEvent) {
			confidence += 0.2
		}
	}

	// Check app usage (work apps active)
	if appsActive, ok := data["apps_active"].([]interface{}); ok {
		context["apps_active"] = appsActive
		for _, app := range appsActive {
			if appName, ok := app.(string); ok {
				if isWorkApp(appName) {
					confidence += 0.15
					break
				}
			}
		}
	}

	// Check device type (laptop/desktop indicates work)
	if deviceType, ok := data["device_type"].(string); ok {
		context["device_type"] = deviceType
		if deviceType == "laptop" || deviceType == "desktop" {
			confidence += 0.3
		}
		if deviceType == "phone" {
			confidence += 0.05 // Could be work but less certain
		}
	}

	// Check keyboard/mouse activity (indicates computer work)
	if keyboardActivity, ok := data["keyboard_activity"].(float64); ok {
		context["keyboard_activity"] = keyboardActivity
		if keyboardActivity > 100 {
			confidence += 0.2
		}
	}

	// Check focus mode
	if focusMode, ok := data["focus_mode"].(bool); ok {
		context["focus_mode"] = focusMode
		if focusMode {
			confidence += 0.15
		}
	}

	// Check video call active (work meeting)
	if videoCall, ok := data["video_call"].(bool); ok {
		context["video_call"] = videoCall
		if videoCall {
			confidence += 0.25
		}
	}

	// Check network (office wifi)
	if networkName, ok := data["network_name"].(string); ok {
		context["network_name"] = networkName
		if containsWorkNetwork(networkName) {
			confidence += 0.2
		}
	}

	// Check Bluetooth devices (work headphones)
	if bluetoothDevices, ok := data["bluetooth_devices"].([]interface{}); ok {
		for _, device := range bluetoothDevices {
			if dev, ok := device.(map[string]interface{}); ok {
				if name, ok := dev["name"].(string); ok {
					if containsWorkKeywords(name) {
						confidence += 0.1
						break
					}
				}
			}
		}
	}

	// Check stress level (higher at work)
	if stressLevel, ok := data["stress_level"].(float64); ok {
		context["stress_level"] = stressLevel
		if stressLevel > 50 {
			confidence += 0.1
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:        SceneWork,
		AgentID:     agentID,
		Confidence:  min(0.95, confidence),
		Active:      true,
		Context:     context,
		TriggeredBy: "work_detector",
	}

	// Define actions for work scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_work_mode",
			Type:        "mode",
			TargetAgent: agentID,
			Priority:    15,
			Payload:     map[string]interface{}{
				"mode":              "work",
				"focus_notifications": true,
				"summarize_calls":    true,
			},
		},
		{
			ID:          "action_filter_social",
			Type:        "filter",
			TargetAgent: "phone",
			Priority:    10,
			Payload:     map[string]interface{}{
				"filter_type":     "social_apps",
				"app_list":        []string{"抖音", "微信朋友圈", "微博"},
				"except_urgent":   true,
			},
		},
		{
			ID:          "action_route_messages",
			Type:        "route",
			TargetAgent: "", // Desktop/laptop
			Priority:    8,
			Payload:     map[string]interface{}{
				"message_type":     "work_messages",
				"route_to":         "desktop",
				"summarize":        true,
			},
		},
		{
			ID:          "action_meeting_prep",
			Type:        "prepare",
			TargetAgent: agentID,
			Priority:    12,
			Payload:     map[string]interface{}{
				"calendar_check":    true,
				"meeting_reminders": true,
				"agenda_preparation": true,
			},
		},
		{
			ID:          "action_focus_timer",
			Type:        "config",
			TargetAgent: agentID,
			Priority:    5,
			Payload:     map[string]interface{}{
				"focus_duration":    25, // Pomodoro
				"break_duration":    5,
				"auto_start":        false,
			},
		},
	}

	return scene, nil
}

// WorkHandler handles work scene actions
type WorkHandler struct{}

// GetName returns handler name
func (h *WorkHandler) GetName() string {
	return "work_handler"
}

// GetSupportedScenes returns supported scene types
func (h *WorkHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneWork}
}

// Handle handles work scene
func (h *WorkHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		switch action.Type {
		case "mode":
			// Enable work mode
		case "filter":
			// Filter social apps
		case "route":
			// Route work messages to desktop
		case "prepare":
			// Prepare for meetings
		case "config":
			// Configure focus timer
		}
	}
	return nil
}

// WorkSession represents a work session
type WorkSession struct {
	SessionID         string
	StartTime         time.Time
	EndTime           *time.Time
	Duration          time.Duration
	Location          string
	FocusPeriods      []FocusPeriod
	Meetings          []WorkMeeting
	TasksCompleted    []string
	TasksInterrupted  []string
	BreakCount        int
	TotalFocusTime    time.Duration
	ProductivityScore float64 // 0-100
	DistractionEvents []DistractionEvent
}

// FocusPeriod represents a focused work period
type FocusPeriod struct {
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Task        string
	Interrupted bool
}

// WorkMeeting represents a work meeting
type WorkMeeting struct {
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Title       string
	Participants []string
	Notes       string
}

// DistractionEvent represents a distraction during work
type DistractionEvent struct {
	Timestamp   time.Time
	Type        string // "social_media", "notification", "call", "email"
	Duration    time.Duration
	Source      string
}

// HomeDetector detects home/personal scenes
type HomeDetector struct{}

// GetName returns detector name
func (d *HomeDetector) GetName() string {
	return "home_detector"
}

// GetSupportedScenes returns supported scene types
func (d *HomeDetector) GetSupportedScenes() []SceneType {
	return []SceneType{SceneHome}
}

// Detect detects home scene from context
func (d *HomeDetector) Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error) {
	confidence := 0.0
	context := map[string]interface{}{
		"agent_id": agentID,
	}

	// Check location
	if location, ok := data["location"].(string); ok {
		context["location"] = location
		if location == "home" || location == "家" || location == "公寓" || location == "apartment" {
			confidence += 0.5
		}
	}

	// Check time (evening/weekend home time)
	hour := time.Now().Hour()
	context["hour"] = hour
	if hour >= 18 && hour <= 23 {
		confidence += 0.2
	}
	if hour >= 6 && hour <= 9 {
		confidence += 0.1 // Morning at home
	}

	// Check weekday (weekends at home)
	weekday := time.Now().Weekday()
	context["weekday"] = weekday
	if weekday == time.Saturday || weekday == time.Sunday {
		confidence += 0.3
	}

	// Check network (home wifi)
	if networkName, ok := data["network_name"].(string); ok {
		context["network_name"] = networkName
		if containsHomeNetwork(networkName) {
			confidence += 0.25
		}
	}

	// Check device type (TV, home speaker)
	if deviceType, ok := data["device_type"].(string); ok {
		context["device_type"] = deviceType
		if deviceType == "tv" || deviceType == "smart_speaker" || deviceType == "tablet" {
			confidence += 0.2
		}
	}

	// Check app usage (entertainment apps)
	if appsActive, ok := data["apps_active"].([]interface{}); ok {
		context["apps_active"] = appsActive
		for _, app := range appsActive {
			if appName, ok := app.(string); ok {
				if isHomeApp(appName) {
					confidence += 0.15
					break
				}
			}
		}
	}

	// Check video watching
	if videoWatching, ok := data["video_watching"].(bool); ok {
		context["video_watching"] = videoWatching
		if videoWatching {
			confidence += 0.1
		}
	}

	// Check family members nearby
	if familyNearby, ok := data["family_nearby"].(bool); ok {
		context["family_nearby"] = familyNearby
		if familyNearby {
			confidence += 0.15
		}
	}

	// Check relaxation indicators
	if relaxing, ok := data["relaxing"].(bool); ok {
		context["relaxing"] = relaxing
		if relaxing {
			confidence += 0.1
		}
	}

	// Check stress level (lower at home)
	if stressLevel, ok := data["stress_level"].(float64); ok {
		context["stress_level"] = stressLevel
		if stressLevel < 30 {
			confidence += 0.1
		}
	}

	if confidence < 0.5 {
		return nil, nil
	}

	scene := &SceneState{
		Type:        SceneHome,
		AgentID:     agentID,
		Confidence:  min(0.95, confidence),
		Active:      true,
		Context:     context,
		TriggeredBy: "home_detector",
	}

	// Define actions for home scene
	scene.Actions = []SceneAction{
		{
			ID:          "action_home_mode",
			Type:        "mode",
			TargetAgent: agentID,
			Priority:    15,
			Payload:     map[string]interface{}{
				"mode":               "home",
				"relaxed_filtering":  true,
				"family_priority":    true,
			},
		},
		{
			ID:          "action_route_to_tv",
			Type:        "route",
			TargetAgent: "", // TV device
			Priority:    8,
			Payload:     map[string]interface{}{
				"content_type":     "media",
				"route_to":         "tv",
				"fallback":         "phone",
			},
		},
		{
			ID:          "action_home_controls",
			Type:        "config",
			TargetAgent: "*", // Smart home devices
			Priority:    5,
			Payload:     map[string]interface{}{
				"smart_home":        true,
				"voice_control":     true,
			},
		},
		{
			ID:          "action_family_sync",
			Type:        "sync",
			TargetAgent: "center",
			Priority:    10,
			Payload:     map[string]interface{}{
				"family_members":    true,
				"shared_calendar":   true,
				"meal_planning":     true,
			},
		},
	}

	return scene, nil
}

// HomeHandler handles home scene actions
type HomeHandler struct{}

// GetName returns handler name
func (h *HomeHandler) GetName() string {
	return "home_handler"
}

// GetSupportedScenes returns supported scene types
func (h *HomeHandler) GetSupportedScenes() []SceneType {
	return []SceneType{SceneHome}
}

// Handle handles home scene
func (h *HomeHandler) Handle(ctx context.Context, scene *SceneState) error {
	for _, action := range scene.Actions {
		switch action.Type {
		case "mode":
			// Enable home mode
		case "route":
			// Route content to TV
		case "config":
			// Configure smart home
		case "sync":
			// Sync with family
		}
	}
	return nil
}

// Helper functions for work detection

func containsWorkKeywords(text string) bool {
	keywords := []string{"工作", "会议", "report", "meeting", "project", "deadline", "客户", "presentation", "汇报", "讨论", "任务"}
	for _, kw := range keywords {
		if containsIgnoreCase(text, kw) {
			return true
		}
	}
	return false
}

func isWorkApp(appName string) bool {
	workApps := []string{"Slack", "Teams", "Zoom", "Outlook", "Gmail", "钉钉", "企业微信", "飞书", "Notion", "Trello", "Jira", "VS Code", "IDE", "Excel", "Word", "PowerPoint", "Figma", "Sketch"}
	for _, app := range workApps {
		if containsIgnoreCase(appName, app) {
			return true
		}
	}
	return false
}

func containsWorkNetwork(networkName string) bool {
	keywords := []string{"office", "公司", "corp", "enterprise", "guest", "work"}
	for _, kw := range keywords {
		if containsIgnoreCase(networkName, kw) {
			return true
		}
	}
	return false
}

// Helper functions for home detection

func containsHomeNetwork(networkName string) bool {
	keywords := []string{"home", "家", "family", "公寓", "apartment", "residence", "xiaomi", "huawei_home"}
	for _, kw := range keywords {
		if containsIgnoreCase(networkName, kw) {
			return true
		}
	}
	return false
}

func isHomeApp(appName string) bool {
	homeApps := []string{"Netflix", "YouTube", "爱奇艺", "腾讯视频", "优酷", "Bilibili", "抖音", "小红书", "微博", "QQ音乐", "网易云音乐", "Spotify", "TikTok", "爱奇艺", "芒果TV"}
	for _, app := range homeApps {
		if containsIgnoreCase(appName, app) {
			return true
		}
	}
	return false
}