package auth

import (
	"testing"
	"time"
)

func TestNewJWTManager(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.publicKey == nil {
		t.Error("Public key not generated")
	}

	if manager.privateKey == nil {
		t.Error("Private key not generated")
	}
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	permissions := []string{"task:execute", "message:send"}
	token, err := manager.GenerateAccessToken("agent-123", 2, permissions)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("Token is empty")
	}

	claims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.AgentID != "agent-123" {
		t.Errorf("Expected agent ID 'agent-123', got '%s'", claims.AgentID)
	}

	if claims.AgentType != 2 {
		t.Errorf("Expected agent type 2, got %d", claims.AgentType)
	}

	if claims.TokenType != TokenTypeAccess {
		t.Errorf("Expected token type '%s', got '%s'", TokenTypeAccess, claims.TokenType)
	}

	if len(claims.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	token, err := manager.GenerateRefreshToken("agent-456")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	claims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if claims.AgentID != "agent-456" {
		t.Errorf("Expected agent ID 'agent-456', got '%s'", claims.AgentID)
	}

	if claims.TokenType != TokenTypeRefresh {
		t.Errorf("Expected token type '%s', got '%s'", TokenTypeRefresh, claims.TokenType)
	}
}

func TestValidateAccessTokenWithRefreshToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	refreshToken, err := manager.GenerateRefreshToken("agent-789")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	_, err = manager.ValidateAccessToken(refreshToken)
	if err == nil {
		t.Error("Should fail when validating refresh token as access token")
	}
}

func TestValidateRefreshTokenWithAccessToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	accessToken, err := manager.GenerateAccessToken("agent-000", 1, nil)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	_, err = manager.ValidateRefreshToken(accessToken)
	if err == nil {
		t.Error("Should fail when validating access token as refresh token")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	pair, err := manager.GenerateTokenPair("agent-pair", 3, []string{"task:execute"})
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if pair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if pair.ExpiresIn <= 0 {
		t.Error("Expires in should be positive")
	}

	if pair.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", pair.TokenType)
	}
}

func TestRefreshAccessToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	refreshToken, err := manager.GenerateRefreshToken("agent-refresh")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	permissions := []string{"task:execute"}
	newAccessToken, err := manager.RefreshAccessToken(refreshToken, permissions)
	if err != nil {
		t.Fatalf("Failed to refresh access token: %v", err)
	}

	claims, err := manager.ValidateAccessToken(newAccessToken)
	if err != nil {
		t.Fatalf("Failed to validate new access token: %v", err)
	}

	if claims.AgentID != "agent-refresh" {
		t.Errorf("Expected agent ID 'agent-refresh', got '%s'", claims.AgentID)
	}
}

func TestInvalidToken(t *testing.T) {
	manager, err := NewJWTManager(nil)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	_, err = manager.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Should fail with invalid token")
	}
}

func TestCustomConfig(t *testing.T) {
	config := &JWTConfig{
		Issuer:        "custom-issuer",
		AccessExpiry:  30 * time.Minute,
		RefreshExpiry: 48 * time.Hour,
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("Failed to create JWT manager: %v", err)
	}

	if manager.config.Issuer != "custom-issuer" {
		t.Errorf("Expected issuer 'custom-issuer', got '%s'", manager.config.Issuer)
	}
}