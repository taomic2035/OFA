package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// AgentTokenType defines the type of agent token
type AgentTokenType string

const (
	AgentTokenTypeAPIKey    AgentTokenType = "api_key"     // Long-lived API key for devices
	AgentTokenTypeSession   AgentTokenType = "session"     // Short-lived session token
	AgentTokenTypeTemporary AgentTokenType = "temporary"   // One-time use token
)

// AgentToken represents a token for agent authentication
type AgentToken struct {
	ID           string          `json:"id"`
	AgentID      string          `json:"agent_id"`
	IdentityID   string          `json:"identity_id,omitempty"`
	TokenType    AgentTokenType  `json:"token_type"`
	TokenHash    string          `json:"token_hash"`    // SHA256 hash of the actual token
	KeyPrefix    string          `json:"key_prefix"`    // First 8 chars for identification
	Name         string          `json:"name"`          // Human-readable name
	Description  string          `json:"description,omitempty"`
	Permissions  []string        `json:"permissions"`
	Scopes       []string        `json:"scopes,omitempty"`  // Resource-level scopes
	CreatedAt    time.Time       `json:"created_at"`
	ExpiresAt    *time.Time      `json:"expires_at,omitempty"`
	LastUsedAt   *time.Time      `json:"last_used_at,omitempty"`
	UsageCount   int64           `json:"usage_count"`
	MaxUsage     int64           `json:"max_usage,omitempty"`  // 0 = unlimited
	IsActive     bool            `json:"is_active"`
	IsRevoked    bool            `json:"is_revoked"`
	RevokedAt    *time.Time      `json:"revoked_at,omitempty"`
	RevokedBy    string          `json:"revoked_by,omitempty"`
	DeviceInfo   map[string]string `json:"device_info,omitempty"`
	RateLimit    *RateLimit      `json:"rate_limit,omitempty"`
	LastMinute   time.Time       `json:"last_minute,omitempty"` // For rate limit tracking
	LastHour     time.Time       `json:"last_hour,omitempty"`
	LastDay      time.Time       `json:"last_day,omitempty"`
}

// RateLimit defines rate limiting configuration for a token
type RateLimit struct {
	RequestsPerMinute int   `json:"requests_per_minute"`
	RequestsPerHour   int   `json:"requests_per_hour"`
	RequestsPerDay    int   `json:"requests_per_day"`
	BurstSize         int   `json:"burst_size"`
}

// AgentTokenConfig holds configuration for agent token manager
type AgentTokenConfig struct {
	DefaultExpiry       time.Duration `yaml:"default_expiry"`
	MaxTokensPerAgent   int           `yaml:"max_tokens_per_agent"`
	TokenLength         int           `yaml:"token_length"`
	EnableRateLimit     bool          `yaml:"enable_rate_limit"`
	DefaultRateLimit    *RateLimit    `yaml:"default_rate_limit"`
}

// DefaultAgentTokenConfig returns default configuration
func DefaultAgentTokenConfig() *AgentTokenConfig {
	return &AgentTokenConfig{
		DefaultExpiry:     30 * 24 * time.Hour, // 30 days
		MaxTokensPerAgent: 5,
		TokenLength:       32,
		EnableRateLimit:   true,
		DefaultRateLimit: &RateLimit{
			RequestsPerMinute: 60,
			RequestsPerHour:   1000,
			RequestsPerDay:    10000,
			BurstSize:         10,
		},
	}
}

// AgentTokenManager manages agent tokens
type AgentTokenManager struct {
	config    *AgentTokenConfig
	tokens    sync.Map // tokenID -> *AgentToken
	byAgent   sync.Map // agentID -> []string (token IDs)
	byPrefix  sync.Map // keyPrefix -> *AgentToken (for quick lookup)
	usage     sync.Map // tokenID -> *UsageTracker
	mu        sync.RWMutex
}

// UsageTracker tracks token usage for rate limiting
type UsageTracker struct {
	MinuteCount  int64
	HourCount    int64
	DayCount     int64
	LastMinute   time.Time
	LastHour     time.Time
	LastDay      time.Time
	TotalCount   int64
	mu           sync.Mutex
}

// NewAgentTokenManager creates a new agent token manager
func NewAgentTokenManager(config *AgentTokenConfig) *AgentTokenManager {
	if config == nil {
		config = DefaultAgentTokenConfig()
	}
	return &AgentTokenManager{
		config: config,
	}
}

// GenerateToken generates a new agent token
func (m *AgentTokenManager) GenerateToken(agentID, identityID, name string, tokenType AgentTokenType, permissions []string) (*AgentToken, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check max tokens per agent
	if tokens, ok := m.byAgent.Load(agentID); ok {
		tokenIDs := tokens.([]string)
		activeCount := 0
		for _, tid := range tokenIDs {
			if t, ok := m.tokens.Load(tid); ok {
				if token := t.(*AgentToken); token.IsActive && !token.IsRevoked {
					activeCount++
				}
			}
		}
		if activeCount >= m.config.MaxTokensPerAgent {
			return nil, "", errors.New("max tokens per agent exceeded")
		}
	}

	// Generate token string
	tokenBytes := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenString := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Create token record
	now := time.Now()
	tokenID := generateTokenID()
	keyPrefix := tokenString[:8]

	token := &AgentToken{
		ID:          tokenID,
		AgentID:     agentID,
		IdentityID:  identityID,
		TokenType:   tokenType,
		TokenHash:   hashToken(tokenString),
		KeyPrefix:   keyPrefix,
		Name:        name,
		Permissions: permissions,
		CreatedAt:   now,
		IsActive:    true,
		IsRevoked:   false,
		RateLimit:   m.config.DefaultRateLimit,
	}

	// Set expiry based on type
	switch tokenType {
	case AgentTokenTypeAPIKey:
		expiry := now.Add(m.config.DefaultExpiry)
		token.ExpiresAt = &expiry
	case AgentTokenTypeSession:
		expiry := now.Add(24 * time.Hour)
		token.ExpiresAt = &expiry
	case AgentTokenTypeTemporary:
		expiry := now.Add(1 * time.Hour)
		token.ExpiresAt = &expiry
		token.MaxUsage = 1
	}

	// Store token
	m.tokens.Store(tokenID, token)
	m.byPrefix.Store(keyPrefix, token)

	// Store by agent
	var agentTokens []string
	if tokens, ok := m.byAgent.Load(agentID); ok {
		agentTokens = tokens.([]string)
	}
	agentTokens = append(agentTokens, tokenID)
	m.byAgent.Store(agentID, agentTokens)

	// Initialize usage tracker
	m.usage.Store(tokenID, &UsageTracker{
		LastMinute: now,
		LastHour:   now,
		LastDay:    now,
	})

	return token, tokenString, nil
}

// ValidateToken validates an agent token string
func (m *AgentTokenManager) ValidateToken(tokenString string) (*AgentToken, error) {
	keyPrefix := tokenString[:8]
	tokenHash := hashToken(tokenString)

	// Find token by prefix
	t, ok := m.byPrefix.Load(keyPrefix)
	if !ok {
		return nil, ErrInvalidToken
	}

	token := t.(*AgentToken)

	// Verify hash
	if token.TokenHash != tokenHash {
		return nil, ErrInvalidToken
	}

	// Check if active
	if !token.IsActive {
		return nil, ErrTokenInactive
	}

	// Check if revoked
	if token.IsRevoked {
		return nil, ErrTokenRevoked
	}

	// Check expiry
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Check usage limit
	if token.MaxUsage > 0 && token.UsageCount >= token.MaxUsage {
		return nil, ErrTokenUsageExceeded
	}

	// Check rate limit
	if m.config.EnableRateLimit && token.RateLimit != nil {
		if err := m.checkRateLimit(token.ID, token.RateLimit); err != nil {
			return nil, err
		}
	}

	// Update usage
	m.updateUsage(token)

	return token, nil
}

// ValidateTokenByID validates a token by its ID (for admin operations)
func (m *AgentTokenManager) ValidateTokenByID(tokenID string) (*AgentToken, error) {
	t, ok := m.tokens.Load(tokenID)
	if !ok {
		return nil, ErrTokenNotFound
	}

	token := t.(*AgentToken)
	if !token.IsActive || token.IsRevoked {
		return nil, ErrTokenInactive
	}

	return token, nil
}

// RevokeToken revokes a token
func (m *AgentTokenManager) RevokeToken(tokenID, revokedBy string) error {
	t, ok := m.tokens.Load(tokenID)
	if !ok {
		return ErrTokenNotFound
	}

	token := t.(*AgentToken)
	now := time.Now()
	token.IsRevoked = true
	token.IsActive = false
	token.RevokedAt = &now
	token.RevokedBy = revokedBy

	m.tokens.Store(tokenID, token)

	return nil
}

// RevokeAllAgentTokens revokes all tokens for an agent
func (m *AgentTokenManager) RevokeAllAgentTokens(agentID, revokedBy string) error {
	tokens, ok := m.byAgent.Load(agentID)
	if !ok {
		return nil
	}

	tokenIDs := tokens.([]string)
	for _, tid := range tokenIDs {
		m.RevokeToken(tid, revokedBy)
	}

	return nil
}

// GetAgentTokens returns all tokens for an agent
func (m *AgentTokenManager) GetAgentTokens(agentID string) ([]*AgentToken, error) {
	tokens, ok := m.byAgent.Load(agentID)
	if !ok {
		return []*AgentToken{}, nil
	}

	tokenIDs := tokens.([]string)
	var result []*AgentToken
	for _, tid := range tokenIDs {
		if t, ok := m.tokens.Load(tid); ok {
			result = append(result, t.(*AgentToken))
		}
	}

	return result, nil
}

// GetToken returns a specific token by ID
func (m *AgentTokenManager) GetToken(tokenID string) (*AgentToken, error) {
	t, ok := m.tokens.Load(tokenID)
	if !ok {
		return nil, ErrTokenNotFound
	}
	return t.(*AgentToken), nil
}

// UpdateToken updates token metadata
func (m *AgentTokenManager) UpdateToken(tokenID string, updates map[string]interface{}) error {
	t, ok := m.tokens.Load(tokenID)
	if !ok {
		return ErrTokenNotFound
	}

	token := t.(*AgentToken)

	// Apply updates
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				token.Name = v
			}
		case "description":
			if v, ok := value.(string); ok {
				token.Description = v
			}
		case "permissions":
			if v, ok := value.([]string); ok {
				token.Permissions = v
			}
		case "scopes":
			if v, ok := value.([]string); ok {
				token.Scopes = v
			}
		case "is_active":
			if v, ok := value.(bool); ok {
				token.IsActive = v
			}
		case "rate_limit":
			if v, ok := value.(*RateLimit); ok {
				token.RateLimit = v
			}
		}
	}

	m.tokens.Store(tokenID, token)
	return nil
}

// ExtendTokenExpiry extends token expiry time
func (m *AgentTokenManager) ExtendTokenExpiry(tokenID string, duration time.Duration) error {
	t, ok := m.tokens.Load(tokenID)
	if !ok {
		return ErrTokenNotFound
	}

	token := t.(*AgentToken)
	if token.IsRevoked {
		return ErrTokenRevoked
	}

	newExpiry := time.Now().Add(duration)
	token.ExpiresAt = &newExpiry
	m.tokens.Store(tokenID, token)

	return nil
}

// CleanupExpiredTokens removes expired tokens
func (m *AgentTokenManager) CleanupExpiredTokens() int {
	count := 0
	now := time.Now()

	m.tokens.Range(func(key, value interface{}) bool {
		token := value.(*AgentToken)
		if token.ExpiresAt != nil && now.After(*token.ExpiresAt) {
			token.IsActive = false
			m.tokens.Store(key, token)
			count++
		}
		return true
	})

	return count
}

// GetStatistics returns token statistics
func (m *AgentTokenManager) GetStatistics() map[string]interface{} {
	var total, active, revoked, expired int

	m.tokens.Range(func(key, value interface{}) bool {
		token := value.(*AgentToken)
		total++
		if token.IsActive && !token.IsRevoked {
			if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
				expired++
			} else {
				active++
			}
		}
		if token.IsRevoked {
			revoked++
		}
		return true
	})

	return map[string]interface{}{
		"total_tokens":   total,
		"active_tokens":  active,
		"revoked_tokens": revoked,
		"expired_tokens": expired,
		"max_per_agent":  m.config.MaxTokensPerAgent,
	}
}

// checkRateLimit checks if the token is within rate limits
func (m *AgentTokenManager) checkRateLimit(tokenID string, limit *RateLimit) error {
	u, ok := m.usage.Load(tokenID)
	if !ok {
		return nil
	}

	usage := u.(*UsageTracker)
	usage.mu.Lock()
	defer usage.mu.Unlock()

	now := time.Now()

	// Reset counters based on time windows
	if now.Sub(usage.LastMinute) >= time.Minute {
		usage.MinuteCount = 0
		usage.LastMinute = now
	}
	if now.Sub(usage.LastHour) >= time.Hour {
		usage.HourCount = 0
		usage.LastHour = now
	}
	if now.Sub(usage.LastDay) >= 24*time.Hour {
		usage.DayCount = 0
		usage.LastDay = now
	}

	// Check limits
	if usage.MinuteCount >= int64(limit.RequestsPerMinute) {
		return ErrRateLimitExceeded
	}
	if usage.HourCount >= int64(limit.RequestsPerHour) {
		return ErrRateLimitExceeded
	}
	if usage.DayCount >= int64(limit.RequestsPerDay) {
		return ErrRateLimitExceeded
	}

	return nil
}

// updateUsage updates token usage statistics
func (m *AgentTokenManager) updateUsage(token *AgentToken) {
	u, ok := m.usage.Load(token.ID)
	if !ok {
		return
	}

	usage := u.(*UsageTracker)
	usage.mu.Lock()
	defer usage.mu.Unlock()

	now := time.Now()
	usage.MinuteCount++
	usage.HourCount++
	usage.DayCount++
	usage.TotalCount++
	usage.LastMinute = now

	token.UsageCount++
	token.LastUsedAt = &now
	m.tokens.Store(token.ID, token)
}

// Helper functions

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Error definitions

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenInactive      = errors.New("token inactive")
	ErrTokenUsageExceeded = errors.New("token usage exceeded")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)

// FormatAgentToken formats the token for display (never show full token)
func FormatAgentToken(token *AgentToken) map[string]interface{} {
	return map[string]interface{}{
		"id":           token.ID,
		"agent_id":     token.AgentID,
		"identity_id":  token.IdentityID,
		"type":         token.TokenType,
		"key_prefix":   token.KeyPrefix,
		"name":         token.Name,
		"description":  token.Description,
		"permissions":  token.Permissions,
		"scopes":       token.Scopes,
		"created_at":   token.CreatedAt,
		"expires_at":   token.ExpiresAt,
		"last_used_at": token.LastUsedAt,
		"usage_count":  token.UsageCount,
		"is_active":    token.IsActive,
		"is_revoked":   token.IsRevoked,
	}
}