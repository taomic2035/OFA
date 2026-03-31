package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType defines the type of JWT token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents custom JWT claims
type Claims struct {
	jwt.RegisteredClaims
	AgentID     string            `json:"agent_id,omitempty"`
	AgentType   int               `json:"agent_type,omitempty"`
	TokenType   TokenType         `json:"token_type"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Issuer        string        `yaml:"issuer"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		Issuer:        "ofa-center",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}
}

// JWTManager handles JWT token operations
type JWTManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	config     *JWTConfig
}

// NewJWTManager creates a new JWT manager with generated keys
func NewJWTManager(config *JWTConfig) (*JWTManager, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     config,
	}, nil
}

// NewJWTManagerWithKeys creates a JWT manager with existing keys
func NewJWTManagerWithKeys(config *JWTConfig, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) (*JWTManager, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     config,
	}, nil
}

// GenerateAccessToken generates an access token for an agent
func (m *JWTManager) GenerateAccessToken(agentID string, agentType int, permissions []string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   agentID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		AgentID:     agentID,
		AgentType:   agentType,
		TokenType:   TokenTypeAccess,
		Permissions: permissions,
	}

	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(m.privateKey)
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(agentID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   agentID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		AgentID:   agentID,
		TokenType: TokenTypeRefresh,
	}

	return jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(m.privateKey)
}

// ValidateToken validates a token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New("not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("not a refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (m *JWTManager) RefreshAccessToken(refreshToken string, permissions []string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	return m.GenerateAccessToken(claims.AgentID, claims.AgentType, permissions)
}

// GetPublicKey returns the public key for verification
func (m *JWTManager) GetPublicKey() ed25519.PublicKey {
	return m.publicKey
}

// GetPrivateKey returns the private key (for key export)
func (m *JWTManager) GetPrivateKey() ed25519.PrivateKey {
	return m.privateKey
}

// TokenPair holds access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

// GenerateTokenPair generates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(agentID string, agentType int, permissions []string) (*TokenPair, error) {
	accessToken, err := m.GenerateAccessToken(agentID, agentType, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := m.GenerateRefreshToken(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.config.AccessExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}