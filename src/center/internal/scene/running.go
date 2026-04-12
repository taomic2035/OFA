package scene

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RunningSceneConfig holds configuration for running scene
type RunningSceneConfig struct {
	MinDuration          time.Duration `yaml:"min_duration"`           // 最小持续时间
	HeartRateThreshold   float64       `yaml:"heart_rate_threshold"`  // 心率阈值
	RouteToPhone         bool          `yaml:"route_to_phone"`        // 路由到手机
	FilterOnWatch        bool          `yaml:"filter_on_watch"`       // 手表端过滤
	NotifyOnStart        bool          `yaml:"notify_on_start"`       // 开始时通知
	NotifyOnEnd          bool          `yaml:"notify_on_end"`         // 结束时通知
}

// DefaultRunningSceneConfig returns default configuration
func DefaultRunningSceneConfig() *RunningSceneConfig {
	return &RunningSceneConfig{
		MinDuration:         60 * time.Second,
		HeartRateThreshold:  100,
		RouteToPhone:        true,
		FilterOnWatch:       true,
		NotifyOnStart:       true,
		NotifyOnEnd:         true,
	}
}

// RunningSceneOrchestrator orchestrates running scene across devices
type RunningSceneOrchestrator struct {
	config       *RunningSceneConfig
	engine       *SceneEngine
	activeRuns   sync.Map // identityID -> *RunningSession
	eventHandler RunningEventHandler
	mu           sync.RWMutex
}

// RunningSession represents an active running session
type RunningSession struct {
	IdentityID    string             `json:"identity_id"`
	WatchAgentID  string             `json:"watch_agent_id"`
	PhoneAgentID  string             `json:"phone_agent_id"`
	StartTime     time.Time          `json:"start_time"`
	EndTime       *time.Time         `json:"end_time,omitempty"`
	Duration      time.Duration      `json:"duration"`
	Distance      float64            `json:"distance"` // km
	Steps         int64              `json:"steps"`
	HeartRateAvg  float64            `json:"heart_rate_avg"`
	HeartRateMax  float64            `json:"heart_rate_max"`
	Calories      float64            `json:"calories"`
	RouteHistory  []RouteEvent       `json:"route_history"`
	Active        bool               `json:"active"`
}

// RouteEvent represents a routing event
type RouteEvent struct {
	Timestamp   time.Time          `json:"timestamp"`
	FromAgent   string             `json:"from_agent"`
	ToAgent     string             `json:"to_agent"`
	MessageType string             `json:"message_type"`
	Payload     map[string]interface{} `json:"payload"`
	Success     bool               `json:"success"`
}

// RunningEventHandler handles running scene events
type RunningEventHandler interface {
	OnRunningStart(session *RunningSession)
	OnRunningEnd(session *RunningSession)
	OnRouteToPhone(event *RouteEvent)
	OnWatchFilter(identityID, message string)
	OnHeartRateAlert(identityID string, heartRate float64)
}

// NewRunningSceneOrchestrator creates a new running scene orchestrator
func NewRunningSceneOrchestrator(engine *SceneEngine, config *RunningSceneConfig) *RunningSceneOrchestrator {
	if config == nil {
		config = DefaultRunningSceneConfig()
	}

	orch := &RunningSceneOrchestrator{
		config: config,
		engine: engine,
	}

	// Register listener
	engine.AddListener(orch)

	return orch
}

// SetEventHandler sets the event handler
func (o *RunningSceneOrchestrator) SetEventHandler(handler RunningEventHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.eventHandler = handler
}

// HandleRunningDetection processes running detection from watch
func (o *RunningSceneOrchestrator) HandleRunningDetection(ctx context.Context, identityID, watchAgentID string, data map[string]interface{}) error {
	// Validate data
	activityType, ok := data["activity_type"].(string)
	if !ok || activityType != "running" {
		return nil
	}

	// Get or create session
	session := o.getOrCreateSession(identityID, watchAgentID, data)

	// Update session metrics
	o.updateSessionMetrics(session, data)

	// Check if should route to phone
	if o.config.RouteToPhone && session.Duration >= o.config.MinDuration {
		err := o.routeToPhone(ctx, session)
		if err != nil {
			return err
		}
	}

	// Check heart rate threshold
	if heartRate, ok := data["heart_rate"].(float64); ok {
		if heartRate > o.config.HeartRateThreshold {
			o.handleHeartRateCheck(session, heartRate)
		}
	}

	// Filter on watch
	if o.config.FilterOnWatch {
		o.filterMessagesOnWatch(session)
	}

	return nil
}

// HandleRunningEnd processes running end event
func (o *RunningSceneOrchestrator) HandleRunningEnd(ctx context.Context, identityID, watchAgentID string, summary map[string]interface{}) error {
	session, ok := o.activeRuns.Load(identityID)
	if !ok {
		return nil
	}

	runSession := session.(*RunningSession)

	// End session
	now := time.Now()
	runSession.EndTime = &now
	runSession.Duration = now.Sub(runSession.StartTime)
	runSession.Active = false

	// Update final metrics
	if distance, ok := summary["distance"].(float64); ok {
		runSession.Distance = distance
	}
	if calories, ok := summary["calories"].(float64); ok {
		runSession.Calories = calories
	}

	// End scene in engine
	o.engine.EndScene(identityID, SceneRunning)

	// Notify end
	if o.config.NotifyOnEnd && o.eventHandler != nil {
		o.eventHandler.OnRunningEnd(runSession)
	}

	// Remove from active
	o.activeRuns.Delete(identityID)

	return nil
}

// RouteMessageToPhone routes a message to phone during running
func (o *RunningSceneOrchestrator) RouteMessageToPhone(ctx context.Context, identityID string, message map[string]interface{}) error {
	session, ok := o.activeRuns.Load(identityID)
	if !ok {
		return fmt.Errorf("no active running session")
	}

	runSession := session.(*RunningSession)

	// Create route event
	event := &RouteEvent{
		Timestamp:   time.Now(),
		FromAgent:   runSession.WatchAgentID,
		ToAgent:     runSession.PhoneAgentID,
		MessageType: "running_message",
		Payload:     message,
	}

	// Execute routing (integrate with CrossDeviceRouter)
	if o.eventHandler != nil {
		o.eventHandler.OnRouteToPhone(event)
	}

	// Record route history
	runSession.RouteHistory = append(runSession.RouteHistory, *event)

	return nil
}

// GetActiveRunning returns active running sessions
func (o *RunningSceneOrchestrator) GetActiveRunning(identityID string) *RunningSession {
	session, ok := o.activeRuns.Load(identityID)
	if !ok {
		return nil
	}
	return session.(*RunningSession)
}

// GetAllActiveRunning returns all active running sessions
func (o *RunningSceneOrchestrator) GetAllActiveRunning() []*RunningSession {
	var sessions []*RunningSession
	o.activeRuns.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*RunningSession))
		return true
	})
	return sessions
}

// SceneListener implementation

func (o *RunningSceneOrchestrator) OnSceneStart(scene *SceneState) {
	if scene.Type != SceneRunning {
		return
	}

	// Notify start
	if o.config.NotifyOnStart && o.eventHandler != nil {
		session := o.GetActiveRunning(scene.IdentityID)
		if session != nil {
			o.eventHandler.OnRunningStart(session)
		}
	}
}

func (o *RunningSceneOrchestrator) OnSceneEnd(scene *SceneState) {
	// Handled by HandleRunningEnd
}

func (o *RunningSceneOrchestrator) OnSceneAction(scene *SceneState, action *SceneAction) {
	if scene.Type != SceneRunning {
		return
	}

	// Handle specific actions
	switch action.Type {
	case "route":
		o.handleRouteAction(scene, action)
	case "filter":
		o.handleFilterAction(scene, action)
	}
}

// Private methods

func (o *RunningSceneOrchestrator) getOrCreateSession(identityID, watchAgentID string, data map[string]interface{}) *RunningSession {
	session, ok := o.activeRuns.Load(identityID)
	if ok {
		return session.(*RunningSession)
	}

	// Create new session
	newSession := &RunningSession{
		IdentityID:    identityID,
		WatchAgentID:  watchAgentID,
		StartTime:     time.Now(),
		Active:        true,
		RouteHistory:  []RouteEvent{},
	}

	// Get phone agent ID from context
	if phoneID, ok := data["phone_agent_id"].(string); ok {
		newSession.PhoneAgentID = phoneID
	}

	o.activeRuns.Store(identityID, newSession)

	// Notify start
	if o.config.NotifyOnStart && o.eventHandler != nil {
		o.eventHandler.OnRunningStart(newSession)
	}

	return newSession
}

func (o *RunningSceneOrchestrator) updateSessionMetrics(session *RunningSession, data map[string]interface{}) {
	// Update duration
	session.Duration = time.Since(session.StartTime)

	// Update steps
	if steps, ok := data["steps"].(int64); ok {
		session.Steps = steps
	}

	// Update heart rate
	if heartRate, ok := data["heart_rate"].(float64); ok {
		// Update max
		if session.HeartRateMax == 0 || heartRate > session.HeartRateMax {
			session.HeartRateMax = heartRate
		}

		// Update average (simple rolling average)
		if session.HeartRateAvg == 0 {
			session.HeartRateAvg = heartRate
		} else {
			session.HeartRateAvg = (session.HeartRateAvg * 0.9) + (heartRate * 0.1)
		}
	}

	// Update distance
	if distance, ok := data["distance"].(float64); ok {
		session.Distance = distance
	}
}

func (o *RunningSceneOrchestrator) routeToPhone(ctx context.Context, session *RunningSession) error {
	if session.PhoneAgentID == "" {
		return nil // No phone registered
	}

	event := &RouteEvent{
		Timestamp:   time.Now(),
		FromAgent:   session.WatchAgentID,
		ToAgent:     session.PhoneAgentID,
		MessageType: "running_status",
		Payload: map[string]interface{}{
			"duration":     session.Duration,
			"distance":     session.Distance,
			"heart_rate":   session.HeartRateAvg,
			"steps":        session.Steps,
			"start_time":   session.StartTime,
		},
	}

	if o.eventHandler != nil {
		o.eventHandler.OnRouteToPhone(event)
	}

	session.RouteHistory = append(session.RouteHistory, *event)

	return nil
}

func (o *RunningSceneOrchestrator) handleHeartRateCheck(session *RunningSession, heartRate float64) {
	// Alert if heart rate is too high
	if heartRate > 160 {
		if o.eventHandler != nil {
			o.eventHandler.OnHeartRateAlert(session.IdentityID, heartRate)
		}
	}
}

func (o *RunningSceneOrchestrator) filterMessagesOnWatch(session *RunningSession) {
	// Filter non-urgent messages on watch
	if o.eventHandler != nil {
		o.eventHandler.OnWatchFilter(session.IdentityID, "filter_urgent_only")
	}
}

func (o *RunningSceneOrchestrator) handleRouteAction(scene *SceneState, action *SceneAction) {
	// Execute route action
	if o.eventHandler != nil {
		event := &RouteEvent{
			Timestamp:   time.Now(),
			FromAgent:   scene.AgentID,
			ToAgent:     action.TargetAgent,
			MessageType: action.Payload["message_type"].(string),
			Payload:     action.Payload,
		}
		o.eventHandler.OnRouteToPhone(event)
	}
}

func (o *RunningSceneOrchestrator) handleFilterAction(scene *SceneState, action *SceneAction) {
	// Execute filter action
	if o.eventHandler != nil {
		o.eventHandler.OnWatchFilter(scene.IdentityID, "filter_urgent_only")
	}
}

// GetStatistics returns running scene statistics
func (o *RunningSceneOrchestrator) GetStatistics() map[string]interface{} {
	var activeCount int
	var totalDuration time.Duration
	var totalDistance float64

	o.activeRuns.Range(func(key, value interface{}) bool {
		session := value.(*RunningSession)
		if session.Active {
			activeCount++
			totalDuration += session.Duration
			totalDistance += session.Distance
		}
		return true
	})

	return map[string]interface{}{
		"active_sessions": activeCount,
		"total_duration":  totalDuration,
		"total_distance":  totalDistance,
		"min_duration":    o.config.MinDuration,
		"heart_threshold": o.config.HeartRateThreshold,
		"route_to_phone":  o.config.RouteToPhone,
		"filter_on_watch": o.config.FilterOnWatch,
	}
}