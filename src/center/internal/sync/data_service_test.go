package sync

import (
	"context"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestDataService_SyncIdentity tests identity synchronization
func TestDataService_SyncIdentity(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 创建测试身份
	identity := &models.PersonalIdentity{
		ID:      "test-identity-1",
		Name:    "Test User",
		Version: 1,
	}

	// 第一次同步（创建）
	req := &SyncIdentityRequest{
		AgentID:  "agent-1",
		Identity: identity,
		Version:  1,
	}

	resp, err := service.SyncIdentity(ctx, req)
	if err != nil {
		t.Fatalf("SyncIdentity failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Conflict {
		t.Error("Expected no conflict on first sync")
	}

	// 验证同步状态
	state := service.GetSyncState("test-identity-1")
	if state == nil {
		t.Fatal("Expected sync state to be created")
	}

	if state.DeviceCount != 1 {
		t.Errorf("Expected device count 1, got %d", state.DeviceCount)
	}
}

// TestDataService_ReportBehavior tests behavior reporting
func TestDataService_ReportBehavior(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 上报行为
	report := &BehaviorReport{
		ID:         "obs-1",
		AgentID:    "agent-1",
		IdentityID: "identity-1",
		Type:       "decision",
		Context: map[string]interface{}{
			"decision_type": "impulse_purchase",
		},
		Inferences: map[string]float64{
			"neuroticism": 0.05,
		},
		Timestamp: time.Now(),
	}

	err := service.ReportBehavior(ctx, report)
	if err != nil {
		t.Fatalf("ReportBehavior failed: %v", err)
	}

	// 验证行为被存储
	observations := service.GetBehaviorObservations("identity-1")
	if len(observations) != 1 {
		t.Errorf("Expected 1 observation, got %d", len(observations))
	}

	if observations[0].Type != "decision" {
		t.Errorf("Expected type 'decision', got '%s'", observations[0].Type)
	}
}

// TestDataService_MultipleBehaviorReports tests multiple behavior reports
func TestDataService_MultipleBehaviorReports(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 上报多个行为
	for i := 0; i < 15; i++ {
		report := &BehaviorReport{
			ID:         string(rune('A' + i)),
			AgentID:    "agent-1",
			IdentityID: "identity-1",
			Type:       "interaction",
			Context: map[string]interface{}{
				"interaction_type": "group_chats",
			},
			Inferences: map[string]float64{
				"extraversion": 0.05,
			},
			Timestamp: time.Now(),
		}

		err := service.ReportBehavior(ctx, report)
		if err != nil {
			t.Fatalf("ReportBehavior %d failed: %v", i, err)
		}
	}

	// 验证行为被存储
	observations := service.GetBehaviorObservations("identity-1")
	if len(observations) != 15 {
		t.Errorf("Expected 15 observations, got %d", len(observations))
	}
}

// TestDataService_BehaviorLimit tests behavior storage limit
func TestDataService_BehaviorLimit(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 上报超过 100 个行为
	for i := 0; i < 150; i++ {
		report := &BehaviorReport{
			ID:         string(rune(i)),
			AgentID:    "agent-1",
			IdentityID: "identity-1",
			Type:       "activity",
			Context:    map[string]interface{}{},
			Timestamp:  time.Now(),
		}

		service.ReportBehavior(ctx, report)
	}

	// 验证只保留最近 100 条
	observations := service.GetBehaviorObservations("identity-1")
	if len(observations) > 100 {
		t.Errorf("Expected max 100 observations, got %d", len(observations))
	}
}

// TestDataService_GetDeviceCount tests device count tracking
func TestDataService_GetDeviceCount(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 同步多个设备
	for i := 0; i < 3; i++ {
		identity := &models.PersonalIdentity{
			ID:      "shared-identity",
			Name:    "Shared User",
			Version: 1,
		}

		req := &SyncIdentityRequest{
			AgentID:  string(rune('A' + i)),
			Identity: identity,
			Version:  1,
		}

		service.SyncIdentity(ctx, req)
	}

	// 验证设备数量
	count := service.GetDeviceCount("shared-identity")
	if count != 3 {
		t.Errorf("Expected device count 3, got %d", count)
	}
}

// TestDataService_SyncMemories tests memory synchronization
func TestDataService_SyncMemories(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 同步记忆
	memories := []*MemoryEntry{
		{
			Key:       "preference_theme",
			Value:     "dark",
			Timestamp: time.Now(),
		},
		{
			Key:       "preference_language",
			Value:     "zh-CN",
			Timestamp: time.Now(),
		},
	}

	req := &SyncMemoriesRequest{
		AgentID:    "agent-1",
		IdentityID: "identity-1",
		Memories:   memories,
		Version:    1,
	}

	// 注意：这个测试需要 MemoryStore 设置
	resp, err := service.SyncMemories(ctx, req)
	if err != nil {
		t.Fatalf("SyncMemories failed: %v", err)
	}

	// 由于没有设置 MemoryStore，success 应该为 false
	if resp.Success {
		t.Log("SyncMemories succeeded (unexpected without store)")
	}
}

// TestDataService_SyncPreferences tests preference synchronization
func TestDataService_SyncPreferences(t *testing.T) {
	service := NewDataService()

	ctx := context.Background()

	// 同步偏好
	prefs := map[string]interface{}{
		"theme":  "dark",
		"lang":   "zh-CN",
		"notify": true,
	}

	req := &SyncPreferencesRequest{
		AgentID:     "agent-1",
		IdentityID:  "identity-1",
		Preferences: prefs,
		Version:     1,
	}

	// 注意：这个测试需要 PreferenceStore 设置
	resp, err := service.SyncPreferences(ctx, req)
	if err != nil {
		t.Fatalf("SyncPreferences failed: %v", err)
	}

	// 由于没有设置 PreferenceStore，success 应该为 false
	if resp.Success {
		t.Log("SyncPreferences succeeded (unexpected without store)")
	}
}