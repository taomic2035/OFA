package memory

import (
	"context"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

func TestRemember(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	memory, err := service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "今天去了公园散步", WithTags([]string{"休闲", "周末"}))
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	if memory.UserID != "user1" {
		t.Errorf("Expected userID 'user1', got '%s'", memory.UserID)
	}

	if memory.Type != models.MemoryTypeEpisodic {
		t.Errorf("Expected type 'episodic', got '%s'", memory.Type)
	}

	if memory.Content != "今天去了公园散步" {
		t.Errorf("Expected content '今天去了公园散步', got '%s'", memory.Content)
	}

	if len(memory.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(memory.Tags))
	}

	if memory.ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestRememberWithOptions(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	memory, err := service.Remember(ctx, "user1", models.MemoryTypeEmotional, "非常开心的一天",
		WithImportance(0.9),
		WithPriority(9),
		WithCategory("daily"),
		WithTags([]string{"开心", "重要"}),
		WithSource("manual"),
		WithEmotion("joy", 0.85),
	)

	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	if memory.Importance != 0.9 {
		t.Errorf("Expected importance 0.9, got %f", memory.Importance)
	}

	if memory.Priority != 9 {
		t.Errorf("Expected priority 9, got %d", memory.Priority)
	}

	if memory.Category != "daily" {
		t.Errorf("Expected category 'daily', got '%s'", memory.Category)
	}

	if memory.Emotion != "joy" {
		t.Errorf("Expected emotion 'joy', got '%s'", memory.Emotion)
	}
}

func TestRecallByID(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memory first
	created, _ := service.Remember(ctx, "user1", models.MemoryTypeFact, "我喜欢喝咖啡")

	// Recall by ID
	memory, err := service.RecallByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("RecallByID failed: %v", err)
	}

	if memory.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, memory.ID)
	}

	if memory.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", memory.AccessCount)
	}
}

func TestRecallByType(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create multiple memories
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "事件1")
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "事件2")
	service.Remember(ctx, "user1", models.MemoryTypeFact, "事实1")

	// Recall by type
	memories, err := service.RecallByType(ctx, "user1", models.MemoryTypeEpisodic, 10)
	if err != nil {
		t.Fatalf("RecallByType failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 episodic memories, got %d", len(memories))
	}
}

func TestRecallRecent(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create multiple memories
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "记忆1")
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "记忆2")
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "记忆3")

	// Recall recent
	memories, err := service.RecallRecent(ctx, "user1", 2)
	if err != nil {
		t.Fatalf("RecallRecent failed: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(memories))
	}
}

func TestRecallImportant(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memories with different importance
	service.Remember(ctx, "user1", models.MemoryTypeFact, "重要事实", WithImportance(0.9))
	service.Remember(ctx, "user1", models.MemoryTypeFact, "一般事实", WithImportance(0.5))
	service.Remember(ctx, "user1", models.MemoryTypeFact, "不重要事实", WithImportance(0.2))

	// Recall important
	memories, err := service.RecallImportant(ctx, "user1", 0.7, 10)
	if err != nil {
		t.Fatalf("RecallImportant failed: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("Expected 1 important memory, got %d", len(memories))
	}
}

func TestForget(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memory
	memory, _ := service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "测试记忆")

	// Forget
	err := service.Forget(ctx, memory.ID)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}

	// Verify deleted
	_, err = service.RecallByID(ctx, memory.ID)
	if err == nil {
		t.Error("Expected error after forget")
	}
}

func TestUpdateMemory(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memory
	memory, _ := service.Remember(ctx, "user1", models.MemoryTypeFact, "原始内容")

	// Update
	updates := map[string]interface{}{
		"content":    "更新内容",
		"importance": 0.8,
		"tags":       []string{"更新", "标签"},
	}

	updated, err := service.UpdateMemory(ctx, memory.ID, updates)
	if err != nil {
		t.Fatalf("UpdateMemory failed: %v", err)
	}

	if updated.Content != "更新内容" {
		t.Errorf("Expected content '更新内容', got '%s'", updated.Content)
	}

	if updated.Importance != 0.8 {
		t.Errorf("Expected importance 0.8, got %f", updated.Importance)
	}
}

func TestAssociateMemories(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memories
	m1, _ := service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "去餐厅吃饭")
	m2, _ := service.Remember(ctx, "user1", models.MemoryTypeFact, "我喜欢这家餐厅")

	// Associate
	err := service.Associate(ctx, m1.ID, m2.ID, "related_to", 0.8)
	if err != nil {
		t.Fatalf("Associate failed: %v", err)
	}

	// Get related
	related, err := service.RecallRelated(ctx, m1.ID, 10)
	if err != nil {
		t.Fatalf("RecallRelated failed: %v", err)
	}

	if len(related) != 1 {
		t.Errorf("Expected 1 related memory, got %d", len(related))
	}
}

func TestConsolidate(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memories with different importance
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "重要记忆", WithImportance(0.9))
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "普通记忆", WithImportance(0.5))
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "不重要记忆", WithImportance(0.1))

	// Consolidate
	result, err := service.Consolidate(ctx, "user1")
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	// Low importance memory should be forgotten
	if len(result.Forgotten) == 0 {
		t.Error("Expected some memories to be forgotten")
	}
}

func TestGetStats(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create various memories
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "事件1", WithCategory("daily"))
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "事件2", WithCategory("daily"))
	service.Remember(ctx, "user1", models.MemoryTypeFact, "事实1", WithCategory("personal"))
	service.Remember(ctx, "user1", models.MemoryTypeFact, "事实2", WithCategory("work"))

	// Get stats
	stats, err := service.GetStats(ctx, "user1")
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalCount != 4 {
		t.Errorf("Expected total count 4, got %d", stats.TotalCount)
	}

	if stats.CountByType[models.MemoryTypeEpisodic] != 2 {
		t.Errorf("Expected 2 episodic memories, got %d", stats.CountByType[models.MemoryTypeEpisodic])
	}

	if stats.CountByType[models.MemoryTypeFact] != 2 {
		t.Errorf("Expected 2 fact memories, got %d", stats.CountByType[models.MemoryTypeFact])
	}
}

func TestMemoryDecay(t *testing.T) {
	memory := models.NewMemory("user1", models.MemoryTypeEpisodic, "测试记忆")
	memory.Importance = 0.8
	memory.DecayFactor = 1.0

	// Decay
	memory.Decay(0.1)

	if memory.DecayFactor >= 1.0 {
		t.Error("Expected decay factor to decrease")
	}

	effective := memory.GetEffectiveImportance()
	if effective >= 0.8 {
		t.Error("Expected effective importance to be lower than original")
	}
}

func TestMemoryAccess(t *testing.T) {
	memory := models.NewMemory("user1", models.MemoryTypeEpisodic, "测试记忆")

	// Access multiple times
	memory.Access()
	memory.Access()
	memory.Access()

	if memory.AccessCount != 3 {
		t.Errorf("Expected access count 3, got %d", memory.AccessCount)
	}
}

func TestMemoryPromotion(t *testing.T) {
	memory := models.NewMemory("user1", models.MemoryTypeEpisodic, "测试记忆")
	memory.Layer = models.MemoryLayerL1

	// Promote to L2
	promoted := memory.Promote()
	if !promoted {
		t.Error("Expected promotion to succeed")
	}
	if memory.Layer != models.MemoryLayerL2 {
		t.Errorf("Expected layer L2, got %s", memory.Layer)
	}

	// Promote to L3
	promoted = memory.Promote()
	if !promoted {
		t.Error("Expected promotion to succeed")
	}
	if memory.Layer != models.MemoryLayerL3 {
		t.Errorf("Expected layer L3, got %s", memory.Layer)
	}

	// Can't promote further
	promoted = memory.Promote()
	if promoted {
		t.Error("Expected promotion to fail for L3")
	}
}

func TestMemoryDemotion(t *testing.T) {
	memory := models.NewMemory("user1", models.MemoryTypeEpisodic, "测试记忆")
	memory.Layer = models.MemoryLayerL3

	// Demote to L2
	demoted := memory.Demote()
	if !demoted {
		t.Error("Expected demotion to succeed")
	}
	if memory.Layer != models.MemoryLayerL2 {
		t.Errorf("Expected layer L2, got %s", memory.Layer)
	}

	// Demote to L1
	demoted = memory.Demote()
	if !demoted {
		t.Error("Expected demotion to succeed")
	}
	if memory.Layer != models.MemoryLayerL1 {
		t.Errorf("Expected layer L1, got %s", memory.Layer)
	}

	// Can't demote further
	demoted = memory.Demote()
	if demoted {
		t.Error("Expected demotion to fail for L1")
	}
}

func TestMemoryTags(t *testing.T) {
	memory := models.NewMemory("user1", models.MemoryTypeEpisodic, "测试记忆")

	memory.AddTag("tag1")
	memory.AddTag("tag2")
	memory.AddTag("tag1") // Duplicate

	if len(memory.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(memory.Tags))
	}
}

func TestRecallByTime(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, nil)
	ctx := context.Background()

	// Create memories
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "记忆1")
	time.Sleep(10 * time.Millisecond)
	service.Remember(ctx, "user1", models.MemoryTypeEpisodic, "记忆2")

	// Get all memories
	memories, _ := service.RecallRecent(ctx, "user1", 10)
	if len(memories) < 2 {
		t.Fatalf("Need at least 2 memories")
	}

	// Query by time range
	start := memories[1].Timestamp
	end := memories[0].Timestamp.Add(time.Second)

	result, err := service.RecallByTime(ctx, "user1", start, end, 10)
	if err != nil {
		t.Fatalf("RecallByTime failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 memories in time range, got %d", len(result))
	}
}