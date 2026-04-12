package scene

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MeetingSceneConfig holds configuration for meeting scene
type MeetingSceneConfig struct {
	DNDModeEnabled     bool          `yaml:"dnd_mode_enabled"`     // 勿扰模式
	DNDDefaultDuration time.Duration `yaml:"dnd_default_duration"` // 默认勿扰时长
	NotifyGlasses      bool          `yaml:"notify_glasses"`       // 通知眼镜
	BlockCalls         bool          `yaml:"block_calls"`          // 拒接来电
	BlockExceptUrgent  bool          `yaml:"block_except_urgent"`  // 仅允许紧急
	EndOnCalendar      bool          `yaml:"end_on_calendar"`      // 会议结束同步日历
}

// DefaultMeetingSceneConfig returns default configuration
func DefaultMeetingSceneConfig() *MeetingSceneConfig {
	return &MeetingSceneConfig{
		DNDModeEnabled:     true,
		DNDDefaultDuration: 1 * time.Hour,
		NotifyGlasses:      true,
		BlockCalls:         true,
		BlockExceptUrgent:  true,
		EndOnCalendar:      true,
	}
}

// MeetingSceneOrchestrator orchestrates meeting scene across devices
type MeetingSceneOrchestrator struct {
	config         *MeetingSceneConfig
	engine         *SceneEngine
	activeMeetings sync.Map // identityID -> *MeetingSession
	eventHandler   MeetingEventHandler
	mu             sync.RWMutex
}

// MeetingSession represents an active meeting session
type MeetingSession struct {
	IdentityID       string             `json:"identity_id"`
	PhoneAgentID     string             `json:"phone_agent_id"`
	GlassesAgentID   string             `json:"glasses_agent_id,omitempty"`
	MeetingTitle     string             `json:"meeting_title"`
	MeetingLocation  string             `json:"meeting_location,omitempty"`
	CalendarEventID  string             `json:"calendar_event_id,omitempty"`
	StartTime        time.Time          `json:"start_time"`
	EndTime          *time.Time         `json:"end_time,omitempty"`
	ScheduledEnd     *time.Time         `json:"scheduled_end,omitempty"`
	Duration         time.Duration      `json:"duration"`
	DNDActive        bool               `json:"dnd_active"`
	Participants     []string           `json:"participants,omitempty"`
	Active           bool               `json:"active"`
	BlockedCalls     int                `json:"blocked_calls"`
	AllowedMessages  int                `json:"allowed_messages"`
}

// MeetingEventHandler handles meeting scene events
type MeetingEventHandler interface {
	OnMeetingStart(session *MeetingSession)
	OnMeetingEnd(session *MeetingSession)
	OnDNDModeChange(identityID string, active bool)
	OnGlassesNotification(session *MeetingSession, message string)
	OnCallBlocked(identityID string, caller string)
	OnUrgentAllowed(identityID string, message map[string]interface{})
}

// NewMeetingSceneOrchestrator creates a new meeting scene orchestrator
func NewMeetingSceneOrchestrator(engine *SceneEngine, config *MeetingSceneConfig) *MeetingSceneOrchestrator {
	if config == nil {
		config = DefaultMeetingSceneConfig()
	}

	orch := &MeetingSceneOrchestrator{
		config: config,
		engine: engine,
	}

	// Register listener
	engine.AddListener(orch)

	return orch
}

// SetEventHandler sets the event handler
func (o *MeetingSceneOrchestrator) SetEventHandler(handler MeetingEventHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.eventHandler = handler
}

// HandleMeetingDetection processes meeting detection from phone
func (o *MeetingSceneOrchestrator) HandleMeetingDetection(ctx context.Context, identityID, phoneAgentID string, data map[string]interface{}) error {
	// Validate meeting context
	calendarEvent, ok := data["calendar_event"].(string)
	if !ok || calendarEvent == "" {
		return nil
	}

	// Get or create session
	session := o.getOrCreateSession(identityID, phoneAgentID, data)

	// Enable DND mode
	if o.config.DNDModeEnabled && !session.DNDActive {
		o.enableDNDMode(session)
	}

	// Block calls
	if o.config.BlockCalls {
		o.blockCalls(session)
	}

	// Notify glasses
	if o.config.NotifyGlasses && session.GlassesAgentID != "" {
		o.notifyGlasses(session)
	}

	return nil
}

// HandleMeetingEnd processes meeting end event
func (o *MeetingSceneOrchestrator) HandleMeetingEnd(ctx context.Context, identityID string) error {
	session, ok := o.activeMeetings.Load(identityID)
	if !ok {
		return nil
	}

	meetingSession := session.(*MeetingSession)

	// End session
	now := time.Now()
	meetingSession.EndTime = &now
	meetingSession.Duration = now.Sub(meetingSession.StartTime)
	meetingSession.Active = false

	// Disable DND mode
	if meetingSession.DNDActive {
		o.disableDNDMode(meetingSession)
	}

	// End scene in engine
	o.engine.EndScene(identityID, SceneMeeting)

	// Notify end
	if o.eventHandler != nil {
		o.eventHandler.OnMeetingEnd(meetingSession)
	}

	// Remove from active
	o.activeMeetings.Delete(identityID)

	return nil
}

// HandleIncomingCall handles incoming call during meeting
func (o *MeetingSceneOrchestrator) HandleIncomingCall(ctx context.Context, identityID, caller string, urgent bool) (bool, error) {
	session, ok := o.activeMeetings.Load(identityID)
	if !ok {
		return true, nil // No active meeting, allow call
	}

	meetingSession := session.(*MeetingSession)

	// Check if urgent
	if urgent && o.config.BlockExceptUrgent {
		// Allow urgent calls
		meetingSession.AllowedMessages++
		if o.eventHandler != nil {
			o.eventHandler.OnUrgentAllowed(identityID, map[string]interface{}{
				"type":   "call",
				"caller": caller,
				"urgent": true,
			})
		}
		return true, nil
	}

	// Block call
	if o.config.BlockCalls {
		meetingSession.BlockedCalls++
		if o.eventHandler != nil {
			o.eventHandler.OnCallBlocked(identityID, caller)
		}
		return false, nil
	}

	return true, nil
}

// HandleIncomingMessage handles incoming message during meeting
func (o *MeetingSceneOrchestrator) HandleIncomingMessage(ctx context.Context, identityID string, message map[string]interface{}) (bool, error) {
	session, ok := o.activeMeetings.Load(identityID)
	if !ok {
		return true, nil // No active meeting, allow message
	}

	// Check if urgent
	urgent, ok := message["urgent"].(bool)
	if urgent && o.config.BlockExceptUrgent {
		meetingSession := session.(*MeetingSession)
		meetingSession.AllowedMessages++

		if o.eventHandler != nil {
			o.eventHandler.OnUrgentAllowed(identityID, message)
		}
		return true, nil
	}

	// Non-urgent: route to glasses silently
	meetingSession := session.(*MeetingSession)
	if meetingSession.GlassesAgentID != "" {
		if o.eventHandler != nil {
			o.eventHandler.OnGlassesNotification(meetingSession, fmt.Sprintf("消息: %s", message["sender"]))
		}
	}

	return false, nil
}

// GetActiveMeeting returns active meeting session
func (o *MeetingSceneOrchestrator) GetActiveMeeting(identityID string) *MeetingSession {
	session, ok := o.activeMeetings.Load(identityID)
	if !ok {
		return nil
	}
	return session.(*MeetingSession)
}

// GetAllActiveMeetings returns all active meeting sessions
func (o *MeetingSceneOrchestrator) GetAllActiveMeetings() []*MeetingSession {
	var sessions []*MeetingSession
	o.activeMeetings.Range(func(key, value interface{}) bool {
		sessions = append(sessions, value.(*MeetingSession))
		return true
	})
	return sessions
}

// SceneListener implementation

func (o *MeetingSceneOrchestrator) OnSceneStart(scene *SceneState) {
	if scene.Type != SceneMeeting {
		return
	}

	// Notify start
	if o.eventHandler != nil {
		session := o.GetActiveMeeting(scene.IdentityID)
		if session != nil {
			o.eventHandler.OnMeetingStart(session)
		}
	}
}

func (o *MeetingSceneOrchestrator) OnSceneEnd(scene *SceneState) {
	// Handled by HandleMeetingEnd
}

func (o *MeetingSceneOrchestrator) OnSceneAction(scene *SceneState, action *SceneAction) {
	if scene.Type != SceneMeeting {
		return
	}

	switch action.Type {
	case "dnd":
		o.handleDNDAction(scene, action)
	case "notify":
		o.handleNotifyAction(scene, action)
	case "block":
		o.handleBlockAction(scene, action)
	}
}

// Private methods

func (o *MeetingSceneOrchestrator) getOrCreateSession(identityID, phoneAgentID string, data map[string]interface{}) *MeetingSession {
	session, ok := o.activeMeetings.Load(identityID)
	if ok {
		return session.(*MeetingSession)
	}

	// Create new session
	newSession := &MeetingSession{
		IdentityID:   identityID,
		PhoneAgentID: phoneAgentID,
		StartTime:    time.Now(),
		Active:       true,
		DNDActive:    false,
	}

	// Get meeting details
	if title, ok := data["calendar_event"].(string); ok {
		newSession.MeetingTitle = title
	}
	if location, ok := data["location"].(string); ok {
		newSession.MeetingLocation = location
	}
	if eventID, ok := data["calendar_event_id"].(string); ok {
		newSession.CalendarEventID = eventID
	}
	if glassesID, ok := data["glasses_agent_id"].(string); ok {
		newSession.GlassesAgentID = glassesID
	}
	if scheduledEnd, ok := data["scheduled_end"].(time.Time); ok {
		newSession.ScheduledEnd = &scheduledEnd
	}

	o.activeMeetings.Store(identityID, newSession)

	return newSession
}

func (o *MeetingSceneOrchestrator) enableDNDMode(session *MeetingSession) {
	session.DNDActive = true

	if o.eventHandler != nil {
		o.eventHandler.OnDNDModeChange(session.IdentityID, true)
	}
}

func (o *MeetingSceneOrchestrator) disableDNDMode(session *MeetingSession) {
	session.DNDActive = false

	if o.eventHandler != nil {
		o.eventHandler.OnDNDModeChange(session.IdentityID, false)
	}
}

func (o *MeetingSceneOrchestrator) blockCalls(session *MeetingSession) {
	// Call blocking is handled by HandleIncomingCall
}

func (o *MeetingSceneOrchestrator) notifyGlasses(session *MeetingSession) {
	if o.eventHandler != nil {
		message := fmt.Sprintf("会议开始: %s", session.MeetingTitle)
		o.eventHandler.OnGlassesNotification(session, message)
	}
}

func (o *MeetingSceneOrchestrator) handleDNDAction(scene *SceneState, action *SceneAction) {
	session := o.GetActiveMeeting(scene.IdentityID)
	if session != nil {
		if !session.DNDActive {
			o.enableDNDMode(session)
		}
	}
}

func (o *MeetingSceneOrchestrator) handleNotifyAction(scene *SceneState, action *SceneAction) {
	session := o.GetActiveMeeting(scene.IdentityID)
	if session != nil && o.eventHandler != nil {
		message := "会议提醒"
		o.eventHandler.OnGlassesNotification(session, message)
	}
}

func (o *MeetingSceneOrchestrator) handleBlockAction(scene *SceneState, action *SceneAction) {
	// Block actions are handled by HandleIncomingCall/HandleIncomingMessage
}

// GetStatistics returns meeting scene statistics
func (o *MeetingSceneOrchestrator) GetStatistics() map[string]interface{} {
	var activeCount int
	var totalBlockedCalls int
	var totalAllowedMessages int

	o.activeMeetings.Range(func(key, value interface{}) bool {
		session := value.(*MeetingSession)
		if session.Active {
			activeCount++
			totalBlockedCalls += session.BlockedCalls
			totalAllowedMessages += session.AllowedMessages
		}
		return true
	})

	return map[string]interface{}{
		"active_meetings":   activeCount,
		"blocked_calls":    totalBlockedCalls,
		"allowed_messages": totalAllowedMessages,
		"dnd_enabled":      o.config.DNDModeEnabled,
		"block_calls":      o.config.BlockCalls,
		"notify_glasses":   o.config.NotifyGlasses,
	}
}