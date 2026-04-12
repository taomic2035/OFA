package llm

import (
	"context"
	"testing"
	"time"
)

func TestGenerateRequestDefaults(t *testing.T) {
	req := &GenerateRequest{
		Message: "Hello",
	}

	if req.MaxTokens == 0 {
		t.Error("MaxTokens should have default value")
	}

	if req.Temperature == 0 {
		t.Error("Temperature should have default value")
	}
}

func TestConversation(t *testing.T) {
	conv := NewConversation("test-session")

	// Add message
	conv.AddMessage("user", "Hello")

	if len(conv.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(conv.Messages))
	}

	// Add response
	conv.AddMessage("assistant", "Hi there!")

	if len(conv.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(conv.Messages))
	}

	// Check history limit
	for i := 0; i < 100; i++ {
		conv.AddMessage("user", "message "+string(rune(i)))
	}

	if len(conv.Messages) > MaxConversationHistory*2 {
		t.Errorf("Conversation history exceeds limit")
	}
}

func TestPersonalityContextBuildSystemPrompt(t *testing.T) {
	ctx := &PersonalityContext{
		Personality: Personality{
			Openness:      0.8,
			Conscientiousness: 0.7,
			Extraversion:  0.5,
			Agreeableness: 0.9,
			Neuroticism:   0.3,
		},
		ValueSystem: ValueSystem{
			Privacy:   0.8,
			Efficiency: 0.7,
			Health:    0.9,
			Family:    0.8,
			Career:    0.6,
		},
	}

	prompt := ctx.BuildSystemPrompt()

	if prompt == "" {
		t.Error("System prompt should not be empty")
	}

	// Check personality traits mentioned
	if !containsSubstring(prompt, "openness") {
		t.Error("System prompt should mention openness")
	}

	if !containsSubstring(prompt, "extraversion") {
		t.Error("System prompt should mention extraversion")
	}
}

func TestStreamChunk(t *testing.T) {
	chunk := &StreamChunk{
		Content: "Hello",
		Delta:   " world",
		Index:   0,
		Finished: false,
	}

	if chunk.Content == "" && chunk.Delta == "" {
		t.Error("Chunk should have content or delta")
	}

	// Finished chunk
	finishedChunk := &StreamChunk{
		Finished: true,
	}

	if !finishedChunk.Finished {
		t.Error("Finished flag should be true")
	}
}

func TestConversationManager(t *testing.T) {
	manager := NewConversationManager()

	// Get new conversation
	conv1 := manager.GetConversation("session-1")
	if conv1 == nil {
		t.Error("Should create new conversation")
	}

	// Get same conversation again
	conv2 := manager.GetConversation("session-1")
	if conv2 != conv1 {
		t.Error("Should return same conversation for same session")
	}

	// Get different conversation
	conv3 := manager.GetConversation("session-2")
	if conv3 == conv1 {
		t.Error("Should create new conversation for different session")
	}

	// Clear conversation
	manager.ClearConversation("session-1")
	conv4 := manager.GetConversation("session-1")
	if len(conv4.Messages) != 0 {
		t.Error("Cleared conversation should have no messages")
	}
}

func TestLLMEngineConfig(t *testing.T) {
	config := &LLMConfig{
		Provider:    "claude",
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	if config.Provider == "" {
		t.Error("Provider should be specified")
	}

	if config.MaxTokens < 1 {
		t.Error("MaxTokens should be positive")
	}

	if config.Temperature < 0 || config.Temperature > 1 {
		t.Error("Temperature should be between 0 and 1")
	}
}

// Mock provider for testing
type MockProvider struct {
	name string
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	return &GenerateResponse{
		Content:   "Mock response",
		Model:     m.name,
		TokensUsed: 100,
	}, nil
}

func (m *MockProvider) GenerateStream(ctx context.Context, req *GenerateRequest) (StreamReader, error) {
	return &MockStreamReader{}, nil
}

func (m *MockProvider) CountTokens(text string) (int, error) {
	return len(text), nil
}

func (m *MockProvider) GetMaxTokens() int {
	return 4096
}

func (m *MockProvider) SupportsStreaming() bool {
	return true
}

func (m *MockProvider) Close() error {
	return nil
}

type MockStreamReader struct {
	chunks []StreamChunk
	index  int
}

func (m *MockStreamReader) Read() (*StreamChunk, error) {
	if m.index >= len(m.chunks) {
		return &StreamChunk{Finished: true}, nil
	}
	chunk := m.chunks[m.index]
	m.index++
	return &chunk, nil
}

func (m *MockStreamReader) Close() error {
	return nil
}

func TestMockProvider(t *testing.T) {
	provider := &MockProvider{name: "mock"}

	ctx := context.Background()
	req := &GenerateRequest{Message: "Test"}

	resp, err := provider.Generate(ctx, req)
	if err != nil {
		t.Errorf("Generate should not error: %v", err)
	}

	if resp.Content != "Mock response" {
		t.Errorf("Expected 'Mock response', got '%s'", resp.Content)
	}

	// Test stream
	reader, err := provider.GenerateStream(ctx, req)
	if err != nil {
		t.Errorf("GenerateStream should not error: %v", err)
	}
	defer reader.Close()
}

func TestRateLimit(t *testing.T) {
	limiter := NewRateLimiter(10, time.Second) // 10 requests per second

	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Errorf("Request %d should be allowed within limit", i)
		}
	}

	// 11th request should be denied
	if limiter.Allow() {
		t.Error("Request exceeding limit should be denied")
	}
}

func TestResponseCache(t *testing.T) {
	cache := NewResponseCache(100, 5*time.Minute)

	key := "test-key"
	response := &GenerateResponse{
		Content:   "Cached response",
		Model:     "claude",
		TokensUsed: 50,
	}

	// Set cache
	cache.Set(key, response)

	// Get cache
 cached, found := cache.Get(key)
 if !found {
  t.Error("Should find cached response")
 }

 if cached.Content != "Cached response" {
  t.Errorf("Expected 'Cached response', got '%s'", cached.Content)
 }

 // Clear cache
 cache.Clear()

 _, foundAfterClear := cache.Get(key)
 if foundAfterClear {
  t.Error("Should not find cached response after clear")
 }
}

func containsSubstring(s, substr string) bool {
 return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
  (len(s) > 0 && len(substr) > 0 && (s[:len(substr)] == substr || 
   containsSubstring(s[1:], substr))))
}
