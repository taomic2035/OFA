package auth

import (
	"sync"
	"testing"
	"time"
)

// === Agent Token Tests ===

func TestAgentTokenManagerCreation(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.config == nil {
		t.Error("Config should be set to default")
	}

	if manager.config.TokenLength != 32 {
		t.Errorf("Expected token length 32, got %d", manager.config.TokenLength)
	}
}

func TestGenerateAgentToken(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	token, tokenString, err := manager.GenerateToken(
		"agent-test-001",
		"identity-001",
		"Test Token",
		AgentTokenTypeAPIKey,
		[]string{"task:execute", "message:send"},
	)

	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == nil {
		t.Fatal("Token should not be nil")
	}

	if tokenString == "" {
		t.Error("Token string should not be empty")
	}

	if token.AgentID != "agent-test-001" {
		t.Errorf("Expected agent ID 'agent-test-001', got '%s'", token.AgentID)
	}

	if token.TokenType != AgentTokenTypeAPIKey {
		t.Errorf("Expected token type APIKey, got %s", token.TokenType)
	}

	if len(token.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(token.Permissions))
	}

	if !token.IsActive {
		t.Error("Token should be active")
	}

	if token.IsRevoked {
		t.Error("Token should not be revoked")
	}

	if token.ExpiresAt == nil {
		t.Error("APIKey token should have expiry")
	}
}

func TestValidateAgentToken(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	_, tokenString, err := manager.GenerateToken(
		"agent-validate-001",
		"",
		"Validate Test",
		AgentTokenTypeAPIKey,
		[]string{"task:execute"},
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	validated, err := manager.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if validated.AgentID != "agent-validate-001" {
		t.Errorf("Expected agent ID 'agent-validate-001', got '%s'", validated.AgentID)
	}

	// Test invalid token
	_, err = manager.ValidateToken("invalid-token-string")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}
}

func TestRevokeAgentToken(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	token, _, err := manager.GenerateToken(
		"agent-revoke-001",
		"",
		"Revoke Test",
		AgentTokenTypeAPIKey,
		[]string{"task:execute"},
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	err = manager.RevokeToken(token.ID, "admin")
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Check token is revoked
	revokedToken, err := manager.GetToken(token.ID)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if !revokedToken.IsRevoked {
		t.Error("Token should be revoked")
	}

	if revokedToken.RevokedBy != "admin" {
		t.Errorf("Expected revoked by 'admin', got '%s'", revokedToken.RevokedBy)
	}
}

func TestMaxTokensPerAgent(t *testing.T) {
	config := &AgentTokenConfig{
		MaxTokensPerAgent: 2,
		TokenLength:       32,
	}
	manager := NewAgentTokenManager(config)

	// Generate first token
	_, _, err := manager.GenerateToken("agent-max-001", "", "Token 1", AgentTokenTypeAPIKey, []string{})
	if err != nil {
		t.Fatalf("Failed to generate first token: %v", err)
	}

	// Generate second token
	_, _, err = manager.GenerateToken("agent-max-001", "", "Token 2", AgentTokenTypeAPIKey, []string{})
	if err != nil {
		t.Fatalf("Failed to generate second token: %v", err)
	}

	// Should fail for third token
	_, _, err = manager.GenerateToken("agent-max-001", "", "Token 3", AgentTokenTypeAPIKey, []string{})
	if err == nil {
		t.Error("Should fail when exceeding max tokens per agent")
	}
}

func TestTokenTypes(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	tests := []struct {
		name      string
		tokenType AgentTokenType
		checkFn   func(token *AgentToken) bool
	}{
		{
			name:      "APIKey",
			tokenType: AgentTokenTypeAPIKey,
			checkFn:   func(t *AgentToken) bool { return t.ExpiresAt != nil && t.ExpiresAt.Sub(t.CreatedAt) >= 30*24*time.Hour },
		},
		{
			name:      "Session",
			tokenType: AgentTokenTypeSession,
			checkFn:   func(t *AgentToken) bool { return t.ExpiresAt != nil && t.ExpiresAt.Sub(t.CreatedAt) >= 24*time.Hour },
		},
		{
			name:      "Temporary",
			tokenType: AgentTokenTypeTemporary,
			checkFn:   func(t *AgentToken) bool { return t.MaxUsage == 1 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _, err := manager.GenerateToken(
				"agent-type-"+tt.name,
				"",
				tt.name+" Test",
				tt.tokenType,
				[]string{},
			)
			if err != nil {
				t.Fatalf("Failed to generate %s token: %v", tt.name, err)
			}

			if !tt.checkFn(token) {
				t.Errorf("%s token validation failed", tt.name)
			}
		})
	}
}

func TestTokenUsageTracking(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	token, tokenString, err := manager.GenerateToken(
		"agent-usage-001",
		"",
		"Usage Test",
		AgentTokenTypeAPIKey,
		[]string{},
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Initial usage count should be 0
	if token.UsageCount != 0 {
		t.Errorf("Initial usage count should be 0, got %d", token.UsageCount)
	}

	// Validate token multiple times
	for i := 0; i < 5; i++ {
		_, err = manager.ValidateToken(tokenString)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}
	}

	// Check usage count
	updatedToken, err := manager.GetToken(token.ID)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Note: ValidateToken updates usage, so count should be > 0
	if updatedToken.UsageCount < 5 {
		t.Errorf("Expected usage count >= 5, got %d", updatedToken.UsageCount)
	}

	if updatedToken.LastUsedAt == nil {
		t.Error("LastUsedAt should be set")
	}
}

func TestGetAgentTokens(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	// Generate multiple tokens for same agent
	for i := 0; i < 3; i++ {
		_, _, err := manager.GenerateToken(
			"agent-list-001",
			"",
			"Token "+string(rune('0'+i)),
			AgentTokenTypeAPIKey,
			[]string{},
		)
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}
	}

	tokens, err := manager.GetAgentTokens("agent-list-001")
	if err != nil {
		t.Fatalf("Failed to get agent tokens: %v", err)
	}

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}
}

func TestTokenStatistics(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	// Generate tokens
	for i := 0; i < 5; i++ {
		_, _, err := manager.GenerateToken(
			"agent-stats-"+string(rune('0'+i)),
			"",
			"Stats Token",
			AgentTokenTypeAPIKey,
			[]string{},
		)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}
	}

	stats := manager.GetStatistics()

	if stats["total_tokens"] == nil {
		t.Error("Should have total_tokens")
	}

	if stats["active_tokens"] == nil {
		t.Error("Should have active_tokens")
	}
}

func TestFormatAgentToken(t *testing.T) {
	token := &AgentToken{
		ID:          "token-001",
		AgentID:     "agent-001",
		TokenType:   AgentTokenTypeAPIKey,
		KeyPrefix:   "abcd1234",
		Name:        "Test Token",
		Permissions: []string{"task:execute"},
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	formatted := FormatAgentToken(token)

	if formatted["id"] != "token-001" {
		t.Error("Formatted token should have ID")
	}

	if formatted["token_hash"] != nil {
		t.Error("Formatted token should NOT expose token_hash")
	}

	if formatted["key_prefix"] != "abcd1234" {
		t.Error("Formatted token should have key_prefix")
	}
}

func TestAgentTokenConcurrency(t *testing.T) {
	manager := NewAgentTokenManager(nil)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _, err := manager.GenerateToken(
				"agent-concurrent-"+string(rune('0'+idx)),
				"",
				"Concurrent Token",
				AgentTokenTypeAPIKey,
				[]string{},
			)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()

	select {
	case err := <-errors:
		t.Errorf("Concurrent token generation failed: %v", err)
	default:
		// All successful
	}
}

// === Permission System Tests ===

func TestPermissionSystemCreation(t *testing.T) {
	ps := NewPermissionSystem()

	if ps == nil {
		t.Fatal("Permission system should not be nil")
	}

	// Check default roles
	admin, err := ps.GetRole("role_admin")
	if err != nil {
		t.Errorf("Should have admin role: %v", err)
	}

	if admin.Name != "Admin" {
		t.Errorf("Admin role name mismatch: %s", admin.Name)
	}

	if !admin.IsSystem {
		t.Error("Admin role should be system role")
	}
}

func TestCreateRole(t *testing.T) {
	ps := NewPermissionSystem()

	role, err := ps.CreateRole(
		"role_custom",
		"Custom Role",
		"Custom role for testing",
		[]string{"task:read", "task:write"},
	)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	if role.ID != "role_custom" {
		t.Errorf("Expected role ID 'role_custom', got '%s'", role.ID)
	}

	if len(role.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(role.Permissions))
	}

	if role.IsSystem {
		t.Error("Custom role should not be system role")
	}
}

func TestAssignRole(t *testing.T) {
	ps := NewPermissionSystem()

	err := ps.AssignRole("agent-001", []string{"role_agent", "role_guest"})
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	roles, err := ps.GetAgentRoles("agent-001")
	if err != nil {
		t.Fatalf("Failed to get agent roles: %v", err)
	}

	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
}

func TestGetEffectivePermissions(t *testing.T) {
	ps := NewPermissionSystem()

	// Create role with inheritance
	_, err := ps.CreateRole("role_inherit", "Inherit Role", "", []string{"task:admin"})
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// Assign role to agent
	err = ps.AssignRole("agent-perm-001", []string{"role_inherit", "role_agent"})
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	permissions := ps.GetEffectivePermissions("agent-perm-001")

	// Should include both role_inherit and role_agent permissions
	if len(permissions) < 1 {
		t.Error("Should have effective permissions")
	}
}

func TestCheckPermission(t *testing.T) {
	ps := NewPermissionSystem()

	// Admin has wildcard permission
	ps.AssignRole("agent-admin-001", []string{"role_admin"})

	if !ps.CheckPermission("agent-admin-001", "any:permission") {
		t.Error("Admin should have any permission")
	}

	// Agent has specific permissions
	ps.AssignRole("agent-specific-001", []string{"role_agent"})

	if !ps.CheckPermission("agent-specific-001", "agent:read:self") {
		t.Error("Agent should have self read permission")
	}

	if ps.CheckPermission("agent-specific-001", "agent:delete") {
		t.Error("Agent should NOT have delete permission")
	}
}

func TestCheckResourcePermission(t *testing.T) {
	ps := NewPermissionSystem()

	ps.AssignRole("agent-resource-001", []string{"role_agent"})

	// Self permission
	if !ps.CheckResourcePermission("agent-resource-001", "agent", "agent-resource-001", "read") {
		t.Error("Agent should be able to read own resource")
	}

	// Other resource
	if ps.CheckResourcePermission("agent-resource-001", "agent", "other-agent", "delete") {
		t.Error("Agent should NOT be able to delete other's resource")
	}
}

func TestAssignResourcePermission(t *testing.T) {
	ps := NewPermissionSystem()

	err := ps.AssignResourcePermission("agent-rp-001", ResourcePermission{
		ResourceType: "task",
		ResourceID:   "task-123",
		Actions:      []string{"read", "write"},
	})
	if err != nil {
		t.Fatalf("Failed to assign resource permission: %v", err)
	}

	if !ps.CheckResourcePermission("agent-rp-001", "task", "task-123", "read") {
		t.Error("Agent should have read permission for task-123")
	}

	if !ps.CheckResourcePermission("agent-rp-001", "task", "task-123", "write") {
		t.Error("Agent should have write permission for task-123")
	}
}

func TestUpdateRole(t *testing.T) {
	ps := NewPermissionSystem()

	ps.CreateRole("role_update", "Update Role", "", []string{"task:read"})

	err := ps.UpdateRole("role_update", map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
		"permissions": []string{"task:read", "task:write"},
	})
	if err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}

	role, _ := ps.GetRole("role_update")

	if role.Name != "Updated Name" {
		t.Errorf("Name should be updated")
	}

	if len(role.Permissions) != 2 {
		t.Errorf("Permissions should be updated")
	}
}

func TestUpdateSystemRole(t *testing.T) {
	ps := NewPermissionSystem()

	err := ps.UpdateRole("role_admin", map[string]interface{}{
		"name": "Modified Admin",
	})
	if err == nil {
		t.Error("Should NOT be able to update system role")
	}
}

func TestDeleteRole(t *testing.T) {
	ps := NewPermissionSystem()

	ps.CreateRole("role_delete", "Delete Role", "", []string{"task:read"})

	err := ps.DeleteRole("role_delete")
	if err != nil {
		t.Fatalf("Failed to delete role: %v", err)
	}

	_, err = ps.GetRole("role_delete")
	if err != ErrRoleNotFound {
		t.Errorf("Should return ErrRoleNotFound, got: %v", err)
	}
}

func TestDeleteSystemRole(t *testing.T) {
	ps := NewPermissionSystem()

	err := ps.DeleteRole("role_admin")
	if err == nil {
		t.Error("Should NOT be able to delete system role")
	}
}

func TestPermissionMatching(t *testing.T) {
	ps := NewPermissionSystem()

	// Test via CheckPermission which uses matchPermission internally
	ps.AssignRole("agent-match-001", []string{"role_agent"})
	ps.AddRoleToAgent("agent-match-001", "role_admin")

	// Admin should have wildcard permission
	if !ps.CheckPermission("agent-match-001", "any:thing") {
		t.Error("Admin wildcard should match any permission")
	}
}

func TestListRoles(t *testing.T) {
	ps := NewPermissionSystem()

	ps.CreateRole("role_list_1", "List Role 1", "", []string{})
	ps.CreateRole("role_list_2", "List Role 2", "", []string{})

	roles := ps.ListRoles()

	// Should include system roles + custom roles
	if len(roles) < 6 { // 4 system + 2 custom
		t.Errorf("Expected at least 6 roles, got %d", len(roles))
	}
}

func TestAddRoleToAgent(t *testing.T) {
	ps := NewPermissionSystem()

	ps.AssignRole("agent-add-role-001", []string{"role_agent"})

	err := ps.AddRoleToAgent("agent-add-role-001", "role_guest")
	if err != nil {
		t.Fatalf("Failed to add role: %v", err)
	}

	roles, _ := ps.GetAgentRoles("agent-add-role-001")
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles after adding, got %d", len(roles))
	}
}

func TestRemoveRoleFromAgent(t *testing.T) {
	ps := NewPermissionSystem()

	ps.AssignRole("agent-remove-role-001", []string{"role_agent", "role_guest"})

	err := ps.RemoveRoleFromAgent("agent-remove-role-001", "role_guest")
	if err != nil {
		t.Fatalf("Failed to remove role: %v", err)
	}

	roles, _ := ps.GetAgentRoles("agent-remove-role-001")
	if len(roles) != 1 {
		t.Errorf("Expected 1 role after removing, got %d", len(roles))
	}
}

func TestPermissionConcurrency(t *testing.T) {
	ps := NewPermissionSystem()

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			agentID := "agent-concurrent-" + string(rune('0'+idx))
			ps.AssignRole(agentID, []string{"role_agent"})
			if !ps.CheckPermission(agentID, "agent:read:self") {
				errChan <- nil
			}
		}(i)

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			roleID := "role-concurrent-" + string(rune('0'+idx))
			_, err := ps.CreateRole(roleID, "Concurrent Role", "", []string{"test:perm"})
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()

	select {
	case err := <-errChan:
		t.Errorf("Concurrent operation failed: %v", err)
	default:
		// All successful
	}
}

// === Rate Limit Tests ===

func TestRateLimitConfig(t *testing.T) {
	config := &AgentTokenConfig{
		EnableRateLimit: true,
		DefaultRateLimit: &RateLimit{
			RequestsPerMinute: 60,
			RequestsPerHour:   1000,
			RequestsPerDay:    10000,
			BurstSize:         10,
		},
	}

	manager := NewAgentTokenManager(config)

	token, tokenString, err := manager.GenerateToken(
		"agent-ratelimit-001",
		"",
		"Rate Limit Test",
		AgentTokenTypeAPIKey,
		[]string{},
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token.RateLimit == nil {
		t.Error("Token should have rate limit config")
	}

	if token.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("Expected 60 requests per minute, got %d", token.RateLimit.RequestsPerMinute)
	}
}

func TestCleanupExpiredTokens(t *testing.T) {
	manager := NewAgentTokenManager(&AgentTokenConfig{
		DefaultExpiry:     1 * time.Millisecond, // Very short expiry
		MaxTokensPerAgent: 100,
		TokenLength:       32,
	})

	// Generate token that will expire quickly
	_, _, err := manager.GenerateToken(
		"agent-cleanup-001",
		"",
		"Cleanup Test",
		AgentTokenTypeAPIKey,
		[]string{},
	)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	count := manager.CleanupExpiredTokens()
	if count < 1 {
		t.Error("Should cleanup at least 1 expired token")
	}
}