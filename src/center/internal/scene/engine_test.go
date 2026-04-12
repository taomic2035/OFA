package scene

import (
	"context"
	"testing"
	"time"
)

// TestSceneEngineCreation tests scene engine creation
func TestSceneEngineCreation(t *testing.T) {
	engine := NewSceneEngine(nil)

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.config == nil {
		t.Error("Config should be set")
	}

	if engine.config.MinConfidence != 0.7 {
		t.Errorf("Default min confidence should be 0.7, got %f", engine.config.MinConfidence)
	}

	if len(engine.detectors) < 3 {
		t.Errorf("Should have at least 3 default detectors, got %d", len(engine.detectors))
	}

	if len(engine.handlers) < 3 {
		t.Errorf("Should have at least 3 default handlers, got %d", len(engine.handlers))
	}
}

// TestDefaultSceneEngineConfig tests default configuration
func TestDefaultSceneEngineConfig(t *testing.T) {
	config := DefaultSceneEngineConfig()

	if config.DetectionInterval != 5*time.Second {
		t.Errorf("Detection interval should be 5s, got %v", config.DetectionInterval)
	}

	if config.MaxActiveScenes != 10 {
		t.Errorf("Max active scenes should be 10, got %d", config.MaxActiveScenes)
	}

	if config.MinConfidence != 0.7 {
		t.Errorf("Min confidence should be 0.7, got %f", config.MinConfidence)
	}

	if config.SceneTimeout != 30*time.Minute {
		t.Errorf("Scene timeout should be 30min, got %v", config.SceneTimeout)
	}
}

// TestRunningDetector tests running scene detection
func TestRunningDetector(t *testing.T) {
	detector := &RunningDetector{}

	// Test basic running detection
	data := map[string]interface{}{
		"activity_type": "running",
		"duration":      120,
		"heart_rate":    140,
		"location":      "outdoor",
	}

	scene, err := detector.Detect(context.Background(), "agent-001", data)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if scene == nil {
		t.Fatal("Should detect running scene")
	}

	if scene.Type != SceneRunning {
		t.Errorf("Scene type should be running, got %s", scene.Type)
	}

	if scene.Confidence < 0.7 {
		t.Errorf("Confidence should be at least 0.7, got %f", scene.Confidence)
	}

	if len(scene.Actions) < 1 {
		t.Error("Should have at least one action")
	}

	// Test non-running activity
	data2 := map[string]interface{}{
		"activity_type": "walking",
	}

	scene2, err := detector.Detect(context.Background(), "agent-001", data2)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if scene2 != nil {
		t.Error("Should not detect running scene for walking")
	}
}

// TestMeetingDetector tests meeting scene detection
func TestMeetingDetector(t *testing.T) {
	detector := &MeetingDetector{}

	// Test meeting detection with calendar event
	data := map[string]interface{}{
		"calendar_event": "项目讨论会议",
		"location":       "会议室",
		"audio_level":    20,
		"device_type":    "phone",
	}

	scene, err := detector.Detect(context.Background(), "agent-002", data)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if scene == nil {
		t.Fatal("Should detect meeting scene")
	}

	if scene.Type != SceneMeeting {
		t.Errorf("Scene type should be meeting, got %s", scene.Type)
	}

	if scene.Confidence < 0.5 {
		t.Errorf("Confidence should be at least 0.5, got %f", scene.Confidence)
	}

	// Check DND action exists
	hasDND := false
	for _, action := range scene.Actions {
		if action.Type == "dnd" {
			hasDND = true
			break
		}
	}
	if !hasDND {
		t.Error("Meeting scene should have DND action")
	}

	// Test non-meeting scenario
	data2 := map[string]interface{}{
		"location": "户外",
	}

	scene2, err := detector.Detect(context.Background(), "agent-002", data2)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if scene2 != nil {
		t.Error("Should not detect meeting scene in outdoor location")
	}
}

// TestHealthAlertDetector tests health alert scene detection
func TestHealthAlertDetector(t *testing.T) {
	detector := &HealthAlertDetector{}

	// Test high heart rate alert
	data := map[string]interface{}{
		"heart_rate": 130,
		"agent_id":   "watch-001",
	}

	scene, err := detector.Detect(context.Background(), "watch-001", data)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	if scene == nil {
		t.Fatal("Should detect health alert scene")
	}

	if scene.Type != SceneHealthAlert {
		t.Errorf("Scene type should be health_alert, got %s", scene.Type)
	}

	if scene.Confidence < 0.8 {
		t.Errorf("Confidence should be at least 0.8 for high heart rate, got %f", scene.Confidence)
	}

	if scene.Context["alert_type"] != "high_heart_rate" {
		t.Errorf("Alert type should be high_heart_rate, got %s", scene.Context["alert_type"])
	}

	// Check broadcast action exists
	hasBroadcast := false
	for _, action := range scene.Actions {
		if action.Type == "alert" {
			hasBroadcast = true
			break
		}
	}
	if !hasBroadcast {
		t.Error("Health alert should have broadcast action")
	}

	// Test low heart rate alert
	data2 := map[string]interface{}{
		"heart_rate": 45,
	}

	scene2, _ := detector.Detect(context.Background(), "watch-001", data2)
	if scene2 == nil {
		t.Error("Should detect low heart rate alert")
	}

	// Test normal heart rate
	data3 := map[string]interface{}{
		"heart_rate": 70,
	}

	scene3, _ := detector.Detect(context.Background(), "watch-001", data3)
	if scene3 != nil {
		t.Error("Should not detect alert for normal heart rate")
	}

	// Test low oxygen
	data4 := map[string]interface{}{
		"oxygen_level": 92,
	}

	scene4, _ := detector.Detect(context.Background(), "watch-001", data4)
	if scene4 == nil {
		t.Error("Should detect low oxygen alert")
	}
}

// TestSceneEngineDetectScene tests full scene detection
func TestSceneEngineDetectScene(t *testing.T) {
	engine := NewSceneEngine(nil)

	// Running scene
	data := map[string]interface{}{
		"activity_type": "running",
		"duration":      180,
		"heart_rate":    150,
	}

	scene, err := engine.DetectScene(context.Background(), "user-001", "watch-001", data)
	if err != nil {
		t.Fatalf("DetectScene failed: %v", err)
	}

	if scene == nil {
		t.Fatal("Should detect scene")
	}

	if scene.IdentityID != "user-001" {
		t.Errorf("IdentityID should be user-001, got %s", scene.IdentityID)
	}

	if scene.AgentID != "watch-001" {
		t.Errorf("AgentID should be watch-001, got %s", scene.AgentID)
	}
}

// TestSceneEngineRules tests trigger rules
func TestSceneEngineRules(t *testing.T) {
	engine := NewSceneEngine(nil)

	// List default rules
	rules := engine.ListRules()
	if len(rules) < 3 {
		t.Errorf("Should have at least 3 default rules, got %d", len(rules))
	}

	// Add custom rule
	now := time.Now()
	customRule := &TriggerRule{
		ID:        "rule_custom_test",
		SceneType: SceneExercise,
		Conditions: []TriggerCondition{
			{Field: "activity", Operator: "eq", Value: "gym"},
		},
		Priority:  5,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	engine.AddRule(customRule)

	// Retrieve rule
	retrieved, err := engine.GetRule("rule_custom_test")
	if err != nil {
		t.Fatalf("GetRule failed: %v", err)
	}

	if retrieved.SceneType != SceneExercise {
		t.Errorf("Scene type should be exercise, got %s", retrieved.SceneType)
	}

	// Delete rule
	engine.DeleteRule("rule_custom_test")

	_, err = engine.GetRule("rule_custom_test")
	if err != ErrRuleNotFound {
		t.Errorf("Should return ErrRuleNotFound, got %v", err)
	}
}

// TestSceneEngineActiveScenes tests active scene management
func TestSceneEngineActiveScenes(t *testing.T) {
	engine := NewSceneEngine(nil)

	// Detect a scene
	data := map[string]interface{}{
		"activity_type": "running",
	}

	scene, _ := engine.DetectScene(context.Background(), "user-001", "watch-001", data)
	if scene == nil {
		t.Skip("Skipping as scene not detected")
	}

	// Get active scenes
	active := engine.GetActiveScenes("user-001")
	if len(active) < 1 {
		t.Error("Should have at least one active scene")
	}

	// End scene
	engine.EndScene("user-001", SceneRunning)

	active2 := engine.GetActiveScenes("user-001")
	for _, s := range active2 {
		if s.Type == SceneRunning && s.Active {
			t.Error("Running scene should be inactive after EndScene")
		}
	}
}

// TestSceneEngineHistory tests scene history
func TestSceneEngineHistory(t *testing.T) {
	engine := NewSceneEngine(nil)

	// Detect and end a scene
	data := map[string]interface{}{
		"heart_rate": 130,
	}

	scene, _ := engine.DetectScene(context.Background(), "user-002", "watch-002", data)
	if scene == nil {
		t.Skip("Skipping as scene not detected")
	}

	engine.EndScene("user-002", SceneHealthAlert)

	// Get history
	history := engine.GetSceneHistory("user-002", 10)
	if len(history) < 1 {
		t.Error("Should have scene history")
	}

	for _, h := range history {
		if h.EndTime == nil {
			t.Error("Ended scenes should have EndTime")
		}
	}
}

// TestSceneEngineStatistics tests statistics
func TestSceneEngineStatistics(t *testing.T) {
	engine := NewSceneEngine(nil)

	stats := engine.GetStatistics()

	if stats["detector_count"] == nil {
		t.Error("Should have detector_count")
	}

	if stats["handler_count"] == nil {
		t.Error("Should have handler_count")
	}

	if stats["rule_count"] == nil {
		t.Error("Should have rule_count")
	}
}

// TestSceneEngineCleanup tests expired scene cleanup
func TestSceneEngineCleanup(t *testing.T) {
	config := &SceneEngineConfig{
		SceneTimeout:    1 * time.Second,
		MinConfidence:   0.7,
		MaxActiveScenes: 10,
	}

	engine := NewSceneEngine(config)

	// Detect scene
	data := map[string]interface{}{
		"activity_type": "running",
	}

	engine.DetectScene(context.Background(), "user-003", "watch-003", data)

	// Wait for timeout
	time.Sleep(2 * time.Second)

	// Cleanup
	count := engine.CleanupExpiredScenes()
	if count < 1 {
		t.Error("Should cleanup at least one expired scene")
	}
}

// TestConditionEvaluation tests condition evaluation
func TestConditionEvaluation(t *testing.T) {
	engine := NewSceneEngine(nil)

	tests := []struct {
		name     string
		cond     TriggerCondition
		data     map[string]interface{}
		expected bool
	}{
		{
			name: "eq_match",
			cond: TriggerCondition{Field: "type", Operator: "eq", Value: "running"},
			data: map[string]interface{}{"type": "running"},
			expected: true,
		},
		{
			name: "eq_no_match",
			cond: TriggerCondition{Field: "type", Operator: "eq", Value: "running"},
			data: map[string]interface{}{"type": "walking"},
			expected: false,
		},
		{
			name: "gt_match",
			cond: TriggerCondition{Field: "value", Operator: "gt", Value: 100},
			data: map[string]interface{}{"value": 150},
			expected: true,
		},
		{
			name: "gt_no_match",
			cond: TriggerCondition{Field: "value", Operator: "gt", Value: 100},
			data: map[string]interface{}{"value": 50},
			expected: false,
		},
		{
			name: "lt_match",
			cond: TriggerCondition{Field: "value", Operator: "lt", Value: 100},
			data: map[string]interface{}{"value": 50},
			expected: true,
		},
		{
			name: "in_match",
			cond: TriggerCondition{Field: "location", Operator: "in", Value: []string{"会议室", "办公室"}},
			data: map[string]interface{}{"location": "会议室"},
			expected: true,
		},
		{
			name: "contains_match",
			cond: TriggerCondition{Field: "event", Operator: "contains", Value: "会议"},
			data: map[string]interface{}{"event": "项目讨论会议"},
			expected: true,
		},
		{
			name: "field_missing",
			cond: TriggerCondition{Field: "type", Operator: "eq", Value: "running"},
			data: map[string]interface{}{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.checkCondition(tt.cond, tt.data)
			if result != tt.expected {
				t.Errorf("Condition check failed: expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSceneListener tests scene listener
func TestSceneListener(t *testing.T) {
	engine := NewSceneEngine(nil)

	listener := &MockSceneListener{}
	engine.AddListener(listener)

	// Detect scene
	data := map[string]interface{}{
		"activity_type": "running",
	}

	engine.DetectScene(context.Background(), "user-004", "watch-004", data)

	// Check listener was called
	if listener.StartCount < 1 {
		t.Error("Listener should be called on scene start")
	}
}

// MockSceneListener for testing
type MockSceneListener struct {
	StartCount  int
	EndCount    int
	ActionCount int
}

func (l *MockSceneListener) OnSceneStart(scene *SceneState) {
	l.StartCount++
}

func (l *MockSceneListener) OnSceneEnd(scene *SceneState) {
	l.EndCount++
}

func (l *MockSceneListener) OnSceneAction(scene *SceneState, action *SceneAction) {
	l.ActionCount++
}

// TestSceneState tests scene state structure
func TestSceneState(t *testing.T) {
	now := time.Now()
	scene := &SceneState{
		Type:       SceneRunning,
		IdentityID: "user-001",
		AgentID:    "watch-001",
		StartTime:  now,
		Active:     true,
		Confidence: 0.85,
		Context: map[string]interface{}{
			"heart_rate": 140,
		},
	}

	if scene.Type != SceneRunning {
		t.Errorf("Type mismatch")
	}

	if !scene.Active {
		t.Error("Should be active")
	}

	if scene.Confidence != 0.85 {
		t.Errorf("Confidence mismatch")
	}
}

// TestSceneAction tests scene action structure
func TestSceneAction(t *testing.T) {
	action := &SceneAction{
		ID:          "action-001",
		Type:        "notify",
		TargetAgent: "phone",
		Priority:    10,
		Payload: map[string]interface{}{
			"message": "Test",
		},
	}

	if action.ID != "action-001" {
		t.Error("ID mismatch")
	}

	if action.Executed {
		t.Error("Should not be executed initially")
	}
}

// BenchmarkSceneDetection benchmarks scene detection
func BenchmarkSceneDetection(b *testing.B) {
	engine := NewSceneEngine(nil)
	data := map[string]interface{}{
		"activity_type": "running",
		"duration":      180,
		"heart_rate":    150,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.DetectScene(context.Background(), "user-001", "watch-001", data)
	}
}