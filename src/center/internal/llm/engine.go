// Package llm provides LLM Engine management (v8.0.0).
package llm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// LLMEngine manages LLM providers and conversations.
type LLMEngine struct {
	mu sync.RWMutex

	// Configuration
	config LLMEngineConfig

	// Providers
	providers map[string]LLMProvider
	primary   LLMProvider
	fallback  LLMProvider

	// Conversation manager
	conversationManager *ConversationManager

	// Intent parser
	intentParser IntentParser

	// Cache
	cache *responseCache
}

// LLMEngineConfig holds configuration for the LLM engine.
type LLMEngineConfig struct {
	// Primary provider name
	PrimaryProvider string `json:"primary_provider"`

	// Fallback provider name (optional)
	FallbackProvider string `json:"fallback_provider,omitempty"`

	// Enable caching
	EnableCache bool `json:"enable_cache"`

	// Cache size
	CacheSize int `json:"cache_size"`

	// Conversation config
	ConversationConfig ConversationConfig `json:"conversation_config"`

	// Enable intent parsing
	EnableIntentParsing bool `json:"enable_intent_parsing"`

	// Auto TTS response
	AutoTTS bool `json:"auto_tts"`

	// Personality integration
	EnablePersonalityIntegration bool `json:"enable_personality_integration"`
}

// DefaultLLMEngineConfig returns default configuration.
func DefaultLLMEngineConfig() LLMEngineConfig {
	return LLMEngineConfig{
		PrimaryProvider:              "claude",
		EnableCache:                  true,
		CacheSize:                    1000,
		ConversationConfig:           DefaultConversationConfig(),
		EnableIntentParsing:          true,
		AutoTTS:                      false,
		EnablePersonalityIntegration: true,
	}
}

// NewLLMEngine creates a new LLM engine.
func NewLLMEngine(config LLMEngineConfig) *LLMEngine {
	engine := &LLMEngine{
		config:    config,
		providers: make(map[string]LLMProvider),
	}

	// Initialize cache
	if config.EnableCache {
		engine.cache = newResponseCache(config.CacheSize)
	}

	// Initialize conversation manager
	engine.conversationManager = NewConversationManager(config.ConversationConfig)

	return engine
}

// RegisterProvider registers an LLM provider.
func (e *LLMEngine) RegisterProvider(provider LLMProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.providers[provider.Name()] = provider

	// Set as primary if matches config
	if provider.Name() == e.config.PrimaryProvider {
		e.primary = provider
	}

	// Set as fallback if matches config
	if provider.Name() == e.config.FallbackProvider {
		e.fallback = provider
	}
}

// SetIntentParser sets the intent parser.
func (e *LLMEngine) SetIntentParser(parser IntentParser) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.intentParser = parser
}

// Chat generates a response for a user message.
func (e *LLMEngine) Chat(ctx context.Context, identityID, userMessage string, personality *PersonalityContext) (*GenerateResponse, error) {
	// Get or create conversation
	conv := e.conversationManager.GetOrCreate(identityID)

	// Build system prompt with personality
	if personality != nil && e.config.EnablePersonalityIntegration {
		conv.SystemPrompt = personality.BuildSystemPrompt()
	}

	// Add user message
	userMsg := Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMsg)

	// Check cache
	if e.cache != nil {
		cached := e.cache.Get(identityID, userMessage)
		if cached != nil {
			return cached, nil
		}
	}

	// Build request
	req := &GenerateRequest{
		ConversationID: conv.ID,
		IdentityID:     identityID,
		Messages:       conv.Messages,
		SystemPrompt:   conv.SystemPrompt,
		MaxTokens:      e.config.ConversationConfig.MaxContextTokens / 2, // leave room for context
		Temperature:    0.7,
	}

	// Generate with primary provider
	response, err := e.generateWithProvider(ctx, e.primary, req)
	if err != nil && e.fallback != nil {
		// Try fallback
		response, err = e.generateWithProvider(ctx, e.fallback, req)
	}

	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Add assistant message to conversation
	assistantMsg := Message{
		Role:      "assistant",
		Content:   response.Content,
		Timestamp: time.Now(),
		Tokens:    response.TokenUsage.OutputTokens,
	}
	conv.Messages = append(conv.Messages, assistantMsg)

	// Update token count
	conv.TokenCount = response.TokenUsage.TotalTokens

	// Manage context window
	e.conversationManager.ManageContextWindow(conv)

	// Cache response
	if e.cache != nil {
		e.cache.Set(identityID, userMessage, response)
	}

	return response, nil
}

// ChatStream generates a streaming response.
func (e *LLMEngine) ChatStream(ctx context.Context, identityID, userMessage string, personality *PersonalityContext) (StreamReader, error) {
	// Get or create conversation
	conv := e.conversationManager.GetOrCreate(identityID)

	// Build system prompt with personality
	if personality != nil && e.config.EnablePersonalityIntegration {
		conv.SystemPrompt = personality.BuildSystemPrompt()
	}

	// Add user message
	userMsg := Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMsg)

	// Build request
	req := &GenerateRequest{
		ConversationID: conv.ID,
		IdentityID:     identityID,
		Messages:       conv.Messages,
		SystemPrompt:   conv.SystemPrompt,
		MaxTokens:      e.config.ConversationConfig.MaxContextTokens / 2,
		Temperature:    0.7,
	}

	// Use primary provider
	provider := e.primary
	if provider == nil {
		return nil, fmt.Errorf("no primary LLM provider registered")
	}

	if !provider.SupportsStreaming() {
		// Fallback to non-streaming
		response, err := e.Chat(ctx, identityID, userMessage, personality)
		if err != nil {
			return nil, err
		}
		return NewMockStreamReader(response), nil
	}

	return provider.GenerateStream(ctx, req)
}

// ParseIntent parses the user message to detect intent.
func (e *LLMEngine) ParseIntent(ctx context.Context, text string) (*Intent, error) {
	if e.intentParser == nil {
		return &Intent{
			Type:         IntentConversation,
			Name:         "chat",
			Confidence:   1.0,
			OriginalText: text,
		}, nil
	}

	return e.intentParser.Parse(ctx, text)
}

// GetConversation returns the conversation for an identity.
func (e *LLMEngine) GetConversation(identityID string) *Conversation {
	return e.conversationManager.Get(identityID)
}

// ClearConversation clears the conversation for an identity.
func (e *LLMEngine) ClearConversation(identityID string) {
	e.conversationManager.Clear(identityID)
}

// GetConversationHistory returns the conversation history.
func (e *LLMEngine) GetConversationHistory(identityID string, limit int) []Message {
	conv := e.conversationManager.Get(identityID)
	if conv == nil {
		return []Message{}
	}

	if limit > 0 && len(conv.Messages) > limit {
		return conv.Messages[len(conv.Messages)-limit:]
	}
	return conv.Messages
}

// generateWithProvider generates with a specific provider.
func (e *LLMEngine) generateWithProvider(ctx context.Context, provider LLMProvider, req *GenerateRequest) (*GenerateResponse, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	return provider.Generate(ctx, req)
}

// Close closes all providers.
func (e *LLMEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, provider := range e.providers {
		provider.Close()
	}

	return nil
}

// ConversationManager manages conversations.
type ConversationManager struct {
	mu sync.RWMutex

	config ConversationConfig

	// Conversations (identityID -> *Conversation)
	conversations sync.Map
}

// NewConversationManager creates a new conversation manager.
func NewConversationManager(config ConversationConfig) *ConversationManager {
	return &ConversationManager{
		config: config,
	}
}

// GetOrCreate gets or creates a conversation for an identity.
func (m *ConversationManager) GetOrCreate(identityID string) *Conversation {
	conv, ok := m.conversations.Load(identityID)
	if ok {
		return conv.(*Conversation)
	}

	// Create new conversation
	id := generateConversationID()
	newConv := &Conversation{
		ID:               id,
		IdentityID:       identityID,
		Messages:         []Message{},
		SystemPrompt:     "",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		TokenCount:       0,
		MaxContextTokens: m.config.MaxContextTokens,
		Metadata:         make(map[string]interface{}),
	}

	m.conversations.Store(identityID, newConv)
	return newConv
}

// Get returns the conversation for an identity.
func (m *ConversationManager) Get(identityID string) *Conversation {
	conv, ok := m.conversations.Load(identityID)
	if !ok {
		return nil
	}
	return conv.(*Conversation)
}

// Clear clears the conversation for an identity.
func (m *ConversationManager) Clear(identityID string) {
	m.conversations.Delete(identityID)
}

// ManageContextWindow manages the context window size.
func (m *ConversationManager) ManageContextWindow(conv *Conversation) {
	// Remove old messages if exceeds max
	if len(conv.Messages) > m.config.MaxMessages {
		// Keep system prompt and recent messages
		conv.Messages = conv.Messages[len(conv.Messages)-m.config.MaxMessages:]
	}

	// Check token count
	if conv.TokenCount > conv.MaxContextTokens {
		// Remove older messages to reduce tokens
		m.trimMessages(conv)
	}

	conv.UpdatedAt = time.Now()
}

// trimMessages trims messages to fit within token limit.
func (m *ConversationManager) trimMessages(conv *Conversation) {
	// Simple implementation: remove oldest messages
	targetTokens := conv.MaxContextTokens * 80 / 100 // keep 80%

	for len(conv.Messages) > 2 && conv.TokenCount > targetTokens {
		// Remove oldest user/assistant pair
		if len(conv.Messages) >= 2 {
			removedTokens := conv.Messages[0].Tokens + conv.Messages[1].Tokens
			conv.Messages = conv.Messages[2:]
			conv.TokenCount -= removedTokens
		}
	}
}

// responseCache caches LLM responses.
type responseCache struct {
	mu sync.RWMutex

	maxSize int
	cache   map[string]*GenerateResponse
}

// newResponseCache creates a new response cache.
func newResponseCache(maxSize int) *responseCache {
	return &responseCache{
		maxSize: maxSize,
		cache:   make(map[string]*GenerateResponse),
	}
}

// Get gets a cached response.
func (c *responseCache) Get(identityID, message string) *GenerateResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := identityID + ":" + message
	return c.cache[key]
}

// Set sets a cached response.
func (c *responseCache) Set(identityID, message string, response *GenerateResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check size limit
	if len(c.cache) >= c.maxSize {
		// Remove oldest entry (simple eviction)
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}

	key := identityID + ":" + message
	c.cache[key] = response
}

// MockStreamReader is a mock stream reader for providers that don't support streaming.
type MockStreamReader struct {
	response *GenerateResponse
	done     bool
}

// NewMockStreamReader creates a mock stream reader.
func NewMockStreamReader(response *GenerateResponse) *MockStreamReader {
	return &MockStreamReader{
		response: response,
		done:     false,
	}
}

// Read reads the next chunk.
func (r *MockStreamReader) Read() (*StreamChunk, error) {
	if r.done {
		return nil, io.EOF
	}

	r.done = true
	return &StreamChunk{
		ConversationID: r.response.ConversationID,
		Content:        r.response.Content,
		IsFinal:        true,
		FinishReason:   r.response.FinishReason,
		TokenUsage:     &r.response.TokenUsage,
		Timestamp:      r.response.Timestamp,
	}, nil
}

// Close closes the stream.
func (r *MockStreamReader) Close() error {
	return nil
}

// Err returns any error.
func (r *MockStreamReader) Err() error {
	return nil
}

// generateConversationID generates a unique conversation ID.
func generateConversationID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "conv_" + hex.EncodeToString(b)
}