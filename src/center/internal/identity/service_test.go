package identity

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

func TestCreateIdentity(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	req := &CreateIdentityRequest{
		Name:     "测试用户",
		Nickname: "小明",
		Gender:   "male",
		Location: "北京",
	}

	identity, err := service.CreateIdentity(ctx, req)
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	if identity.Name != "测试用户" {
		t.Errorf("Expected name '测试用户', got '%s'", identity.Name)
	}

	if identity.Nickname != "小明" {
		t.Errorf("Expected nickname '小明', got '%s'", identity.Nickname)
	}

	if identity.ID == "" {
		t.Error("Expected non-empty ID")
	}

	// Check default personality
	if identity.Personality == nil {
		t.Error("Expected default personality")
	}

	// Check default value system
	if identity.ValueSystem == nil {
		t.Error("Expected default value system")
	}
}

func TestGetIdentity(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Get
	identity, err := service.GetIdentity(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	if identity.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, identity.ID)
	}
}

func TestUpdateIdentity(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "原名称",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Update
	updates := map[string]interface{}{
		"name":     "新名称",
		"nickname": "新昵称",
		"location": "上海",
	}

	updated, err := service.UpdateIdentity(ctx, created.ID, updates)
	if err != nil {
		t.Fatalf("UpdateIdentity failed: %v", err)
	}

	if updated.Name != "新名称" {
		t.Errorf("Expected name '新名称', got '%s'", updated.Name)
	}

	if updated.Nickname != "新昵称" {
		t.Errorf("Expected nickname '新昵称', got '%s'", updated.Nickname)
	}
}

func TestDeleteIdentity(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Delete
	err = service.DeleteIdentity(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteIdentity failed: %v", err)
	}

	// Verify deleted
	_, err = service.GetIdentity(ctx, created.ID)
	if err == nil {
		t.Error("Expected error after delete")
	}
}

func TestUpdatePersonality(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Update personality
	updates := map[string]float64{
		"openness":          0.8,
		"conscientiousness": 0.7,
		"extraversion":      0.3,
		"custom_trait":      0.5,
	}

	personality, err := service.UpdatePersonality(ctx, created.ID, updates)
	if err != nil {
		t.Fatalf("UpdatePersonality failed: %v", err)
	}

	if personality.Openness != 0.8 {
		t.Errorf("Expected openness 0.8, got %f", personality.Openness)
	}

	if personality.CustomTraits["custom_trait"] != 0.5 {
		t.Errorf("Expected custom_trait 0.5, got %f", personality.CustomTraits["custom_trait"])
	}
}

func TestSetSpeakingTone(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Set speaking tone
	err = service.SetSpeakingTone(ctx, created.ID, "humorous", "brief", 0.8)
	if err != nil {
		t.Fatalf("SetSpeakingTone failed: %v", err)
	}

	// Verify
	identity, _ := service.GetIdentity(ctx, created.ID)
	if identity.Personality.SpeakingTone != "humorous" {
		t.Errorf("Expected speaking tone 'humorous', got '%s'", identity.Personality.SpeakingTone)
	}

	if identity.Personality.ResponseLength != "brief" {
		t.Errorf("Expected response length 'brief', got '%s'", identity.Personality.ResponseLength)
	}
}

func TestUpdateValueSystem(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Update value system
	updates := map[string]float64{
		"privacy":       0.9,
		"efficiency":    0.8,
		"risk_tolerance": 0.7,
		"custom_value":  0.5,
	}

	valueSystem, err := service.UpdateValueSystem(ctx, created.ID, updates)
	if err != nil {
		t.Fatalf("UpdateValueSystem failed: %v", err)
	}

	if valueSystem.Privacy != 0.9 {
		t.Errorf("Expected privacy 0.9, got %f", valueSystem.Privacy)
	}

	if valueSystem.CustomValues["custom_value"] != 0.5 {
		t.Errorf("Expected custom_value 0.5, got %f", valueSystem.CustomValues["custom_value"])
	}
}

func TestInterestManagement(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Add interest
	interest := models.Interest{
		Category:    "sports",
		Name:        "篮球",
		Level:       0.8,
		Keywords:    []string{"NBA", "运动"},
		Description: "喜欢打篮球",
	}

	err = service.AddInterest(ctx, created.ID, interest)
	if err != nil {
		t.Fatalf("AddInterest failed: %v", err)
	}

	// Get interests
	interests, err := service.GetInterests(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetInterests failed: %v", err)
	}

	if len(interests) != 1 {
		t.Errorf("Expected 1 interest, got %d", len(interests))
	}

	if interests[0].Name != "篮球" {
		t.Errorf("Expected interest name '篮球', got '%s'", interests[0].Name)
	}

	// Get by category
	sportsInterests, err := service.GetInterestsByCategory(ctx, created.ID, "sports")
	if err != nil {
		t.Fatalf("GetInterestsByCategory failed: %v", err)
	}

	if len(sportsInterests) != 1 {
		t.Errorf("Expected 1 sports interest, got %d", len(sportsInterests))
	}

	// Update level
	err = service.UpdateInterestLevel(ctx, created.ID, interests[0].ID, 0.9)
	if err != nil {
		t.Fatalf("UpdateInterestLevel failed: %v", err)
	}

	// Verify update
	interests, _ = service.GetInterests(ctx, created.ID)
	if interests[0].Level != 0.9 {
		t.Errorf("Expected level 0.9, got %f", interests[0].Level)
	}

	// Remove interest
	err = service.RemoveInterest(ctx, created.ID, interests[0].ID)
	if err != nil {
		t.Fatalf("RemoveInterest failed: %v", err)
	}

	// Verify removed
	interests, _ = service.GetInterests(ctx, created.ID)
	if len(interests) != 0 {
		t.Errorf("Expected 0 interests after removal, got %d", len(interests))
	}
}

func TestVoiceProfile(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Get default voice profile
	profile, err := service.GetVoiceProfile(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVoiceProfile failed: %v", err)
	}

	if profile == nil {
		t.Error("Expected default voice profile")
	}

	// Update voice profile
	newProfile := &models.VoiceProfile{
		VoiceType:    "clone",
		Pitch:        1.2,
		Speed:        0.9,
		Volume:       0.7,
		Tone:         "energetic",
		EmotionLevel: 0.8,
	}

	err = service.UpdateVoiceProfile(ctx, created.ID, newProfile)
	if err != nil {
		t.Fatalf("UpdateVoiceProfile failed: %v", err)
	}

	// Verify update
	profile, _ = service.GetVoiceProfile(ctx, created.ID)
	if profile.Pitch != 1.2 {
		t.Errorf("Expected pitch 1.2, got %f", profile.Pitch)
	}

	if profile.Tone != "energetic" {
		t.Errorf("Expected tone 'energetic', got '%s'", profile.Tone)
	}
}

func TestWritingStyle(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Get default writing style
	style, err := service.GetWritingStyle(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetWritingStyle failed: %v", err)
	}

	if style == nil {
		t.Error("Expected default writing style")
	}

	// Update writing style
	newStyle := &models.WritingStyle{
		Formality:         0.7,
		Verbosity:         0.4,
		Humor:             0.6,
		UseEmoji:          false,
		SignaturePhrase:   "Peace out",
		PreferredGreeting: "Hey",
	}

	err = service.UpdateWritingStyle(ctx, created.ID, newStyle)
	if err != nil {
		t.Fatalf("UpdateWritingStyle failed: %v", err)
	}

	// Verify update
	style, _ = service.GetWritingStyle(ctx, created.ID)
	if style.Formality != 0.7 {
		t.Errorf("Expected formality 0.7, got %f", style.Formality)
	}

	if style.SignaturePhrase != "Peace out" {
		t.Errorf("Expected signature phrase 'Peace out', got '%s'", style.SignaturePhrase)
	}
}

func TestInferPersonality(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Create behavior observations
	observations := []models.BehaviorObservation{
		{
			Type:    "decision",
			Outcome: "novel_trying",
			Context: map[string]interface{}{},
		},
		{
			Type:    "interaction",
			Outcome: "groupChats",
			Context: map[string]interface{}{},
		},
		{
			Type:    "activity",
			Outcome: "exploring_new",
			Context: map[string]interface{}{},
		},
	}

	// Infer personality
	personality, err := service.InferPersonalityFromBehavior(ctx, created.ID, observations)
	if err != nil {
		t.Fatalf("InferPersonalityFromBehavior failed: %v", err)
	}

	// Should have increased openness and extraversion
	if personality.Openness <= 0.5 {
		t.Errorf("Expected increased openness from novel_trying, got %f", personality.Openness)
	}

	if personality.Extraversion <= 0.5 {
		t.Errorf("Expected increased extraversion from groupChats, got %f", personality.Extraversion)
	}
}

func TestGetDecisionContext(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create first
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name: "测试",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Add some interests
	service.AddInterest(ctx, created.ID, models.Interest{
		Category: "tech",
		Name:     "编程",
		Level:    0.9,
	})

	// Get decision context
	decisionCtx, err := service.GetDecisionContext(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetDecisionContext failed: %v", err)
	}

	if decisionCtx.UserID != created.ID {
		t.Errorf("Expected user ID '%s', got '%s'", created.ID, decisionCtx.UserID)
	}

	if decisionCtx.Personality == nil {
		t.Error("Expected personality in decision context")
	}

	if decisionCtx.ValueSystem == nil {
		t.Error("Expected value system in decision context")
	}

	if len(decisionCtx.Interests) != 1 {
		t.Errorf("Expected 1 interest, got %d", len(decisionCtx.Interests))
	}

	if len(decisionCtx.ValuePriority) == 0 {
		t.Error("Expected value priority list")
	}
}

func TestListIdentities(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	// Create multiple identities
	for i := 0; i < 5; i++ {
		service.CreateIdentity(ctx, &CreateIdentityRequest{
			Name: "用户" + string(rune('A'+i)),
		})
	}

	// List with pagination
	identities, total, err := service.ListIdentities(ctx, 1, 3)
	if err != nil {
		t.Fatalf("ListIdentities failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(identities) != 3 {
		t.Errorf("Expected 3 identities on page 1, got %d", len(identities))
	}

	// Second page
	identities, _, err = service.ListIdentities(ctx, 2, 3)
	if err != nil {
		t.Fatalf("ListIdentities page 2 failed: %v", err)
	}

	if len(identities) != 2 {
		t.Errorf("Expected 2 identities on page 2, got %d", len(identities))
	}
}

func TestFileStore(t *testing.T) {
	// Use temp directory
	tempDir := t.TempDir()

	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	service := NewService(store)
	ctx := context.Background()

	// Create
	created, err := service.CreateIdentity(ctx, &CreateIdentityRequest{
		Name:     "文件测试",
		Nickname: "FileTest",
	})
	if err != nil {
		t.Fatalf("CreateIdentity failed: %v", err)
	}

	// Verify file exists
	files, _ := os.ReadDir(tempDir)
	if len(files) != 1 {
		t.Errorf("Expected 1 file in directory, got %d", len(files))
	}

	// Get
	identity, err := service.GetIdentity(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}

	if identity.Name != "文件测试" {
		t.Errorf("Expected name '文件测试', got '%s'", identity.Name)
	}
}