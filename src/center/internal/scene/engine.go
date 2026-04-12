package scene

import (
	"context"
	"sync"
	"time"
)

// SceneType defines the type of scene
type SceneType string

const (
	SceneRunning      SceneType = "running"      // 跑步场景
	SceneMeeting      SceneType = "meeting"      // 会议场景
	SceneHealthAlert  SceneType = "health_alert" // 健康异常场景
	SceneDriving      SceneType = "driving"      // 驾驶场景
	SceneSleeping     SceneType = "sleeping"     // 睡眠场景
	SceneExercise     SceneType = "exercise"     // 运动场景
	SceneWork         SceneType = "work"         // 工作场景
	SceneHome         SceneType = "home"         // 家庭场景
	SceneTravel       SceneType = "travel"       // 旅行场景
	SceneEntertainment SceneType = "entertainment" // 娱乐场景
)

// SceneState represents the current state of a scene
type SceneState struct {
	Type         SceneType          `json:"type"`
	IdentityID   string             `json:"identity_id"`
	AgentID      string             `json:"agent_id"`       // 检测到的 Agent
	StartTime    time.Time          `json:"start_time"`
	EndTime      *time.Time         `json:"end_time,omitempty"`
	Duration     time.Duration      `json:"duration"`
	Confidence   float64            `json:"confidence"`     // 检测置信度
	Active       bool               `json:"active"`
	Context      map[string]interface{} `json:"context"`     // 场景上下文
	TriggeredBy  string             `json:"triggered_by"`   // 触发源
	Affects      []string           `json:"affects"`        // 影响的设备列表
	Actions      []SceneAction      `json:"actions"`        // 执行的动作
}

// SceneAction represents an action to be executed in a scene
type SceneAction struct {
	ID           string             `json:"id"`
	Type         string             `json:"type"`           // notify, route, block, allow, alert
	TargetAgent  string             `json:"target_agent"`   // 目标 Agent
	Priority     int                `json:"priority"`       // 优先级
	Payload      map[string]interface{} `json:"payload"`
	Executed     bool               `json:"executed"`
	ExecutedAt   *time.Time         `json:"executed_at,omitempty"`
	Result       string             `json:"result,omitempty"`
}

// TriggerRule defines the rule for triggering a scene
type TriggerRule struct {
	ID           string             `json:"id"`
	SceneType    SceneType          `json:"scene_type"`
	Conditions   []TriggerCondition `json:"conditions"`
	Priority     int                `json:"priority"`
	Enabled      bool               `json:"enabled"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// TriggerCondition defines a condition for triggering
type TriggerCondition struct {
	Field        string             `json:"field"`          // 检测字段
	Operator     string             `json:"operator"`       // eq, gt, lt, in, contains
	Value        interface{}        `json:"value"`
	Duration     time.Duration      `json:"duration"`       // 持续时间要求
	Confidence   float64            `json:"confidence"`     // 置信度阈值
}

// SceneDetector defines the interface for scene detection
type SceneDetector interface {
	// Detect detects a scene from agent data
	Detect(ctx context.Context, agentID string, data map[string]interface{}) (*SceneState, error)

	// GetSupportedScenes returns supported scene types
	GetSupportedScenes() []SceneType

	// GetName returns detector name
	GetName() string
}

// SceneHandler defines the interface for handling scenes
type SceneHandler interface {
	// Handle handles a scene event
	Handle(ctx context.Context, scene *SceneState) error

	// GetSupportedScenes returns supported scene types
	GetSupportedScenes() []SceneType

	// GetName returns handler name
	GetName() string
}

// SceneEngineConfig holds configuration for scene engine
type SceneEngineConfig struct {
	DetectionInterval   time.Duration `yaml:"detection_interval"`
	MaxActiveScenes     int           `yaml:"max_active_scenes"`
	MinConfidence       float64       `yaml:"min_confidence"`
	SceneTimeout        time.Duration `yaml:"scene_timeout"`
	EnableBroadcast     bool          `yaml:"enable_broadcast"`
}

// DefaultSceneEngineConfig returns default configuration
func DefaultSceneEngineConfig() *SceneEngineConfig {
	return &SceneEngineConfig{
		DetectionInterval:   5 * time.Second,
		MaxActiveScenes:     10,
		MinConfidence:       0.7,
		SceneTimeout:        30 * time.Minute,
		EnableBroadcast:     true,
	}
}

// SceneEngine manages scene detection and handling
type SceneEngine struct {
	config       *SceneEngineConfig
	detectors    []SceneDetector
	handlers     []SceneHandler
	rules        sync.Map // ruleID -> *TriggerRule
	activeScenes sync.Map // identityID -> []*SceneState
	sceneHistory sync.Map // identityID -> []*SceneState
	listeners    []SceneListener
	mu           sync.RWMutex
}

// SceneListener listens for scene events
type SceneListener interface {
	OnSceneStart(scene *SceneState)
	OnSceneEnd(scene *SceneState)
	OnSceneAction(scene *SceneState, action *SceneAction)
}

// NewSceneEngine creates a new scene engine
func NewSceneEngine(config *SceneEngineConfig) *SceneEngine {
	if config == nil {
		config = DefaultSceneEngineConfig()
	}
	engine := &SceneEngine{
		config:    config,
		detectors: []SceneDetector{},
		handlers:  []SceneHandler{},
		listeners: []SceneListener{},
	}

	// Register default detectors
	engine.RegisterDetector(&RunningDetector{})
	engine.RegisterDetector(&MeetingDetector{})
	engine.RegisterDetector(&HealthAlertDetector{})

	// Register default handlers
	engine.RegisterHandler(&NotificationHandler{})
	engine.RegisterHandler(&RoutingHandler{})
	engine.RegisterHandler(&AlertHandler{})

	// Initialize default rules
	engine.initDefaultRules()

	return engine
}

// initDefaultRules initializes default trigger rules
func (e *SceneEngine) initDefaultRules() {
	now := time.Now()

	// Running scene rule
	e.AddRule(&TriggerRule{
		ID:        "rule_running_default",
		SceneType: SceneRunning,
		Conditions: []TriggerCondition{
			{Field: "activity_type", Operator: "eq", Value: "running"},
			{Field: "duration", Operator: "gt", Value: 60}, // 至少60秒
			{Confidence: 0.8},
		},
		Priority:  10,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})

	// Meeting scene rule
	e.AddRule(&TriggerRule{
		ID:        "rule_meeting_default",
		SceneType: SceneMeeting,
		Conditions: []TriggerCondition{
			{Field: "calendar_event", Operator: "contains", Value: "会议"},
			{Field: "location", Operator: "in", Value: []string{"会议室", "办公室"}},
			{Confidence: 0.75},
		},
		Priority:  8,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})

	// Health alert rule
	e.AddRule(&TriggerRule{
		ID:        "rule_health_alert_default",
		SceneType: SceneHealthAlert,
		Conditions: []TriggerCondition{
			{Field: "heart_rate", Operator: "gt", Value: 120},
			{Field: "heart_rate", Operator: "lt", Value: 50},
			{Duration: 30 * time.Second},
			{Confidence: 0.9},
		},
		Priority:  15, // 最高优先级
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// RegisterDetector registers a scene detector
func (e *SceneEngine) RegisterDetector(detector SceneDetector) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.detectors = append(e.detectors, detector)
}

// RegisterHandler registers a scene handler
func (e *SceneEngine) RegisterHandler(handler SceneHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers = append(e.handlers, handler)
}

// AddListener adds a scene listener
func (e *SceneEngine) AddListener(listener SceneListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

// AddRule adds a trigger rule
func (e *SceneEngine) AddRule(rule *TriggerRule) {
	e.rules.Store(rule.ID, rule)
}

// GetRule retrieves a rule by ID
func (e *SceneEngine) GetRule(ruleID string) (*TriggerRule, error) {
	r, ok := e.rules.Load(ruleID)
	if !ok {
		return nil, ErrRuleNotFound
	}
	return r.(*TriggerRule), nil
}

// DeleteRule deletes a rule
func (e *SceneEngine) DeleteRule(ruleID string) {
	e.rules.Delete(ruleID)
}

// ListRules lists all rules
func (e *SceneEngine) ListRules() []*TriggerRule {
	var rules []*TriggerRule
	e.rules.Range(func(key, value interface{}) bool {
		rules = append(rules, value.(*TriggerRule))
		return true
	})
	return rules
}

// DetectScene detects scenes from agent data
func (e *SceneEngine) DetectScene(ctx context.Context, identityID, agentID string, data map[string]interface{}) (*SceneState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var detectedScenes []*SceneState

	for _, detector := range e.detectors {
		scene, err := detector.Detect(ctx, agentID, data)
		if err != nil {
			continue
		}

		if scene != nil && scene.Confidence >= e.config.MinConfidence {
			scene.IdentityID = identityID
			scene.AgentID = agentID
			scene.StartTime = time.Now()
			scene.Active = true

			// Apply rules
			if e.applyRules(scene, data) {
				detectedScenes = append(detectedScenes, scene)
			}
		}
	}

	if len(detectedScenes) == 0 {
		return nil, nil
	}

	// Return highest priority scene
	var bestScene *SceneState
	for _, scene := range detectedScenes {
		if bestScene == nil || scene.Confidence > bestScene.Confidence {
			bestScene = scene
		}
	}

	// Store active scene
	e.storeActiveScene(identityID, bestScene)

	// Notify listeners
	for _, listener := range e.listeners {
		listener.OnSceneStart(bestScene)
	}

	return bestScene, nil
}

// applyRules applies trigger rules to a scene
func (e *SceneEngine) applyRules(scene *SceneState, data map[string]interface{}) bool {
	e.rules.Range(func(key, value interface{}) bool {
		rule := value.(*TriggerRule)

		if !rule.Enabled || rule.SceneType != scene.Type {
			return true
		}

		// Check conditions
		matched := true
		for _, cond := range rule.Conditions {
			if !e.checkCondition(cond, data) {
				matched = false
				break
			}
		}

		if matched {
			scene.Confidence = max(scene.Confidence, rule.Priority/20.0)
		}

		return true
	})

	return true
}

// checkCondition checks a trigger condition
func (e *SceneEngine) checkCondition(cond TriggerCondition, data map[string]interface{}) bool {
	value, ok := data[cond.Field]
	if !ok {
		return false
	}

	switch cond.Operator {
	case "eq":
		return value == cond.Value
	case "gt":
		if v, ok := value.(float64); ok {
			if cv, ok := cond.Value.(float64); ok {
				return v > cv
			}
			if cv, ok := cond.Value.(int); ok {
				return v > float64(cv)
			}
		}
	case "lt":
		if v, ok := value.(float64); ok {
			if cv, ok := cond.Value.(float64); ok {
				return v < cv
			}
			if cv, ok := cond.Value.(int); ok {
				return v < float64(cv)
			}
		}
	case "in":
		if arr, ok := cond.Value.([]string); ok {
			if v, ok := value.(string); ok {
				for _, a := range arr {
					if a == v {
						return true
					}
				}
			}
		}
	case "contains":
		if v, ok := value.(string); ok {
			if cv, ok := cond.Value.(string); ok {
				return containsString(v, cv)
			}
		}
	}

	return false
}

// HandleScene handles a detected scene
func (e *SceneEngine) HandleScene(ctx context.Context, scene *SceneState) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, handler := range e.handlers {
		supported := false
		for _, st := range handler.GetSupportedScenes() {
			if st == scene.Type {
				supported = true
				break
			}
		}

		if supported {
			err := handler.Handle(ctx, scene)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// EndScene ends an active scene
func (e *SceneEngine) EndScene(identityID string, sceneType SceneType) error {
	scenes, ok := e.activeScenes.Load(identityID)
	if !ok {
		return nil
	}

	sceneList := scenes.([]*SceneState)
	now := time.Now()

	for _, scene := range sceneList {
		if scene.Type == sceneType && scene.Active {
			scene.EndTime = &now
			scene.Duration = now.Sub(scene.StartTime)
			scene.Active = false

			// Notify listeners
			for _, listener := range e.listeners {
				listener.OnSceneEnd(scene)
			}

			// Store to history
			e.addToHistory(identityID, scene)
		}
	}

	return nil
}

// GetActiveScenes returns active scenes for an identity
func (e *SceneEngine) GetActiveScenes(identityID string) []*SceneState {
	scenes, ok := e.activeScenes.Load(identityID)
	if !ok {
		return []*SceneState{}
	}

	var active []*SceneState
	for _, scene := range scenes.([]*SceneState) {
		if scene.Active {
			active = append(active, scene)
		}
	}
	return active
}

// GetSceneHistory returns scene history for an identity
func (e *SceneEngine) GetSceneHistory(identityID string, limit int) []*SceneState {
	history, ok := e.sceneHistory.Load(identityID)
	if !ok {
		return []*SceneState{}
	}

	sceneHistory := history.([]*SceneState)
	if limit > 0 && len(sceneHistory) > limit {
		return sceneHistory[:limit]
	}
	return sceneHistory
}

// storeActiveScene stores an active scene
func (e *SceneEngine) storeActiveScene(identityID string, scene *SceneState) {
	var scenes []*SceneState
	if s, ok := e.activeScenes.Load(identityID); ok {
		scenes = s.([]*SceneState)
	}

	// Check max active scenes
	if len(scenes) >= e.config.MaxActiveScenes {
		// Remove oldest inactive scene
		for i, s := range scenes {
			if !s.Active {
				scenes = append(scenes[:i], scenes[i+1:]...)
				break
			}
		}
	}

	scenes = append(scenes, scene)
	e.activeScenes.Store(identityID, scenes)
}

// addToHistory adds scene to history
func (e *SceneEngine) addToHistory(identityID string, scene *SceneState) {
	var history []*SceneState
	if h, ok := e.sceneHistory.Load(identityID); ok {
		history = h.([]*SceneState)
	}

	history = append(history, scene)

	// Keep only last 100 scenes
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	e.sceneHistory.Store(identityID, history)
}

// CleanupExpiredScenes cleans up expired active scenes
func (e *SceneEngine) CleanupExpiredScenes() int {
	count := 0
	now := time.Now()

	e.activeScenes.Range(func(key, value interface{}) bool {
		scenes := value.([]*SceneState)
		for _, scene := range scenes {
			if scene.Active && now.Sub(scene.StartTime) > e.config.SceneTimeout {
				scene.Active = false
				scene.EndTime = &now
				scene.Duration = now.Sub(scene.StartTime)
				count++

				for _, listener := range e.listeners {
					listener.OnSceneEnd(scene)
				}
			}
		}
		return true
	})

	return count
}

// ExecuteAction executes a scene action
func (e *SceneEngine) ExecuteAction(ctx context.Context, scene *SceneState, action *SceneAction) error {
	now := time.Now()
	action.Executed = true
	action.ExecutedAt = &now

	// Notify listeners
	for _, listener := range e.listeners {
		listener.OnSceneAction(scene, action)
	}

	return nil
}

// GetStatistics returns scene engine statistics
func (e *SceneEngine) GetStatistics() map[string]interface{} {
	var totalActive, totalHistory int

	e.activeScenes.Range(func(key, value interface{}) bool {
		for _, scene := range value.([]*SceneState) {
			if scene.Active {
				totalActive++
			}
		}
		return true
	})

	e.sceneHistory.Range(func(key, value interface{}) bool {
		totalHistory += len(value.([]*SceneState))
		return true
	})

	var ruleCount int
	e.rules.Range(func(key, value interface{}) bool {
		ruleCount++
		return true
	})

	return map[string]interface{}{
		"active_scenes":    totalActive,
		"history_count":    totalHistory,
		"rule_count":       ruleCount,
		"detector_count":   len(e.detectors),
		"handler_count":    len(e.handlers),
		"min_confidence":   e.config.MinConfidence,
		"scene_timeout":    e.config.SceneTimeout,
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Error definitions

var ErrRuleNotFound = errorf("rule not found")

func errorf(msg string) error {
	return &SceneError{Message: msg}
}

type SceneError struct {
	Message string
}

func (e *SceneError) Error() string {
	return e.Message
}