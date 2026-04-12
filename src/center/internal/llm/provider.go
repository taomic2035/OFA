// Package llm provides LLM (Large Language Model) integration for intelligent conversations (v8.0.0).
//
// This package implements the core interaction layer that was missing in v5.x,
// enabling natural dialogue capabilities with Claude, GPT, and local models.
package llm

import (
	"context"
	"io"
	"time"
)

// LLMProvider is the interface for LLM providers (Claude, GPT, etc).
type LLMProvider interface {
	// Name returns the provider name (e.g., "claude", "gpt", "local").
	Name() string

	// Generate generates a response from the LLM.
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	// GenerateStream generates a streaming response from the LLM.
	GenerateStream(ctx context.Context, req *GenerateRequest) (StreamReader, error)

	// CountTokens counts the tokens in a message.
	CountTokens(text string) (int, error)

	// GetMaxTokens returns the maximum tokens this provider can handle.
	GetMaxTokens() int

	// SupportsStreaming returns whether this provider supports streaming.
	SupportsStreaming() bool

	// Close closes the provider connection.
	Close() error
}

// StreamReader reads streaming responses.
type StreamReader interface {
	// Read reads the next chunk of the response.
	Read() (*StreamChunk, error)

	// Close closes the stream.
	Close() error

	// Err returns any error that occurred during streaming.
	Err() error
}

// GenerateRequest represents a request to generate a response.
type GenerateRequest struct {
	// Conversation ID for context tracking
	ConversationID string `json:"conversation_id"`

	// Identity ID for personality integration
	IdentityID string `json:"identity_id"`

	// Messages in the conversation
	Messages []Message `json:"messages"`

	// System prompt (optional, overrides default)
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Max tokens to generate
	MaxTokens int `json:"max_tokens"`

	// Temperature for randomness (0-1)
	Temperature float64 `json:"temperature"`

	// Stop sequences (optional)
	StopSequences []string `json:"stop_sequences,omitempty"`

	// Metadata for tracking
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a message in the conversation.
type Message struct {
	// Role: "user", "assistant", "system"
	Role string `json:"role"`

	// Content of the message
	Content string `json:"content"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Token count (optional)
	Tokens int `json:"tokens,omitempty"`

	// Metadata (optional)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateResponse represents a response from the LLM.
type GenerateResponse struct {
	// Conversation ID
	ConversationID string `json:"conversation_id"`

	// Generated content
	Content string `json:"content"`

	// Role: "assistant"
	Role string `json:"role"`

	// Token usage
	TokenUsage TokenUsage `json:"token_usage"`

	// Finish reason: "stop", "length", "error"
	FinishReason string `json:"finish_reason"`

	// Model used
	Model string `json:"model"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StreamChunk represents a chunk of a streaming response.
type StreamChunk struct {
	// Conversation ID
	ConversationID string `json:"conversation_id"`

	// Content chunk
	Content string `json:"content"`

	// Is this the final chunk?
	IsFinal bool `json:"is_final"`

	// Finish reason (only on final chunk)
	FinishReason string `json:"finish_reason,omitempty"`

	// Token usage (only on final chunk)
	TokenUsage *TokenUsage `json:"token_usage,omitempty"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// TokenUsage represents token usage statistics.
type TokenUsage struct {
	// Input tokens
	InputTokens int `json:"input_tokens"`

	// Output tokens
	OutputTokens int `json:"output_tokens"`

	// Total tokens
	TotalTokens int `json:"total_tokens"`
}

// LLMProviderConfig holds configuration for an LLM provider.
type LLMProviderConfig struct {
	// Provider name
	Name string `json:"name"`

	// API key
	APIKey string `json:"api_key"`

	// API URL (optional, for custom endpoints)
	APIURL string `json:"api_url,omitempty"`

	// Model to use
	Model string `json:"model"`

	// Max tokens for this provider
	MaxTokens int `json:"max_tokens"`

	// Default temperature
	DefaultTemperature float64 `json:"default_temperature"`

	// Request timeout
	Timeout time.Duration `json:"timeout"`

	// Enable caching
	EnableCache bool `json:"enable_cache"`

	// Cache TTL
	CacheTTL time.Duration `json:"cache_ttl"`

	// Rate limiting
	MaxRequestsPerMinute int `json:"max_requests_per_minute"`

	// Retry configuration
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	RetryBackoff  float64       `json:"retry_backoff"`
}

// DefaultLLMProviderConfig returns default configuration.
func DefaultLLMProviderConfig() LLMProviderConfig {
	return LLMProviderConfig{
		MaxTokens:            4096,
		DefaultTemperature:   0.7,
		Timeout:              30 * time.Second,
		EnableCache:          true,
		CacheTTL:             5 * time.Minute,
		MaxRequestsPerMinute: 60,
		MaxRetries:           3,
		RetryDelay:           1 * time.Second,
		RetryBackoff:         2.0,
	}
}

// Conversation represents a conversation context.
type Conversation struct {
	// Conversation ID
	ID string `json:"id"`

	// Identity ID
	IdentityID string `json:"identity_id"`

	// Messages in the conversation
	Messages []Message `json:"messages"`

	// System prompt
	SystemPrompt string `json:"system_prompt"`

	// Created at
	CreatedAt time.Time `json:"created_at"`

	// Updated at
	UpdatedAt time.Time `json:"updated_at"`

	// Token count (total)
	TokenCount int `json:"token_count"`

	// Max tokens for context
	MaxContextTokens int `json:"max_context_tokens"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationConfig holds configuration for conversation management.
type ConversationConfig struct {
	// Max messages to keep in history
	MaxMessages int `json:"max_messages"`

	// Max tokens for context window
	MaxContextTokens int `json:"max_context_tokens"`

	// Enable memory integration
	EnableMemoryIntegration bool `json:"enable_memory_integration"`

	// Memory integration depth (how many recent memories to include)
	MemoryIntegrationDepth int `json:"memory_integration_depth"`

	// Auto-clear old conversations
	AutoClearAge time.Duration `json:"auto_clear_age"`

	// Conversation TTL
	ConversationTTL time.Duration `json:"conversation_ttl"`
}

// DefaultConversationConfig returns default conversation configuration.
func DefaultConversationConfig() ConversationConfig {
	return ConversationConfig{
		MaxMessages:             50,
		MaxContextTokens:        8192,
		EnableMemoryIntegration: true,
		MemoryIntegrationDepth:  5,
		AutoClearAge:            24 * time.Hour,
		ConversationTTL:         1 * time.Hour,
	}
}

// PersonalityContext represents personality integration for conversations.
type PersonalityContext struct {
	// Identity ID
	IdentityID string `json:"identity_id"`

	// Emotion context (from EmotionEngine)
	EmotionContext string `json:"emotion_context,omitempty"`

	// Philosophy context (from PhilosophyEngine)
	PhilosophyContext string `json:"philosophy_context,omitempty"`

	// Social context (from SocialIdentityEngine)
	SocialContext string `json:"social_context,omitempty"`

	// Culture context (from RegionalCultureEngine)
	CultureContext string `json:"culture_context,omitempty"`

	// LifeStage context (from LifeStageEngine)
	LifeStageContext string `json:"lifestage_context,omitempty"`

	// Relationship context (from RelationshipEngine)
	RelationshipContext string `json:"relationship_context,omitempty"`

	// Speech style (from SpeechContentEngine)
	SpeechStyle string `json:"speech_style,omitempty"`

	// Name for personalization
	Name string `json:"name,omitempty"`

	// Nickname
	Nickname string `json:"nickname,omitempty"`
}

// BuildSystemPrompt builds a system prompt from personality context.
func (p *PersonalityContext) BuildSystemPrompt() string {
	prompt := "You are a digital assistant with the following personality:\n\n"

	if p.Name != "" {
		prompt += "Name: " + p.Name + "\n"
	}
	if p.Nickname != "" {
		prompt += "Nickname: " + p.Nickname + "\n"
	}

	if p.EmotionContext != "" {
		prompt += "\nEmotional state: " + p.EmotionContext + "\n"
	}
	if p.PhilosophyContext != "" {
		prompt += "\nWorldview and values: " + p.PhilosophyContext + "\n"
	}
	if p.SocialContext != "" {
		prompt += "\nSocial identity: " + p.SocialContext + "\n"
	}
	if p.CultureContext != "" {
		prompt += "\nCultural background: " + p.CultureContext + "\n"
	}
	if p.LifeStageContext != "" {
		prompt += "\nLife stage: " + p.LifeStageContext + "\n"
	}
	if p.RelationshipContext != "" {
		prompt += "\nRelationship style: " + p.RelationshipContext + "\n"
	}
	if p.SpeechStyle != "" {
		prompt += "\nCommunication style: " + p.SpeechStyle + "\n"
	}

	prompt += "\nRespond in a way that reflects this personality while being helpful and natural."

	return prompt
}

// IntentType represents the type of intent detected.
type IntentType string

const (
	IntentQuestion    IntentType = "question"
	IntentCommand     IntentType = "command"
	IntentConversation IntentType = "conversation"
	IntentSkill       IntentType = "skill"
	IntentUnknown     IntentType = "unknown"
)

// Intent represents a parsed intent.
type Intent struct {
	// Intent type
	Type IntentType `json:"type"`

	// Intent name (e.g., "set_timer", "check_weather")
	Name string `json:"name"`

	// Confidence (0-1)
	Confidence float64 `json:"confidence"`

	// Entities extracted
	Entities []Entity `json:"entities"`

	// Original text
	OriginalText string `json:"original_text"`

	// Skill to route to (if applicable)
	SkillID string `json:"skill_id,omitempty"`

	// Skill parameters
	SkillParams map[string]interface{} `json:"skill_params,omitempty"`
}

// Entity represents an extracted entity.
type Entity struct {
	// Entity type (e.g., "time", "location", "person")
	Type string `json:"type"`

	// Entity value
	Value string `json:"value"`

	// Position in text
	Start int `json:"start"`
	End   int `json:"end"`

	// Confidence
	Confidence float64 `json:"confidence"`
}

// IntentParser parses user input to detect intents.
type IntentParser interface {
	// Parse parses the user input to detect intent.
	Parse(ctx context.Context, text string) (*Intent, error)

	// ParseWithContext parses with additional context.
	ParseWithContext(ctx context.Context, text string, context map[string]interface{}) (*Intent, error)

	// GetSupportedIntents returns the list of supported intents.
	GetSupportedIntents() []string
}