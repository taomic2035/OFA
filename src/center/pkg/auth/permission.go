package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// Role represents a user/agent role with associated permissions
type Role struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Permissions []string          `json:"permissions"`
	Inherits    []string          `json:"inherits,omitempty"`  // Role IDs to inherit from
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	IsDefault   bool              `json:"is_default"`
	IsSystem    bool              `json:"is_system"`  // System roles cannot be deleted
}

// ResourcePermission defines permission for a specific resource
type ResourcePermission struct {
	ResourceType string   `json:"resource_type"`  // agent, task, identity, device, etc.
	ResourceID   string   `json:"resource_id,omitempty"`  // Specific resource or "*" for all
	Actions      []string `json:"actions"`  // read, write, delete, execute, etc.
	Conditions   map[string]interface{} `json:"conditions,omitempty"`  // Additional conditions
}

// PermissionSystem manages roles and permissions
type PermissionSystem struct {
	roles       sync.Map // roleID -> *Role
	roleAssign  sync.Map // agentID -> []roleID
	resources   sync.Map // agentID -> []ResourcePermission
	mu          sync.RWMutex
}

// NewPermissionSystem creates a new permission system
func NewPermissionSystem() *PermissionSystem {
	ps := &PermissionSystem{}
	ps.initializeDefaultRoles()
	return ps
}

// initializeDefaultRoles sets up system default roles
func (ps *PermissionSystem) initializeDefaultRoles() {
	now := time.Now()

	// System roles (cannot be modified or deleted)
	systemRoles := []Role{
		{
			ID:          "role_admin",
			Name:        "Admin",
			Description: "Full system administrator",
			Permissions: []string{"*"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsDefault:   false,
			IsSystem:    true,
		},
		{
			ID:          "role_agent",
			Name:        "Agent",
			Description: "Standard agent role",
			Permissions: []string{
				"agent:read:self",
				"agent:write:self",
				"task:read:self",
				"task:write:self",
				"identity:read:self",
				"identity:write:self",
				"sync:read:self",
				"sync:write:self",
			},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsDefault:   true,
			IsSystem:    true,
		},
		{
			ID:          "role_guest",
			Name:        "Guest",
			Description: "Limited guest access",
			Permissions: []string{
				"agent:read:self",
				"identity:read:self",
			},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsDefault:   false,
			IsSystem:    true,
		},
		{
			ID:          "role_service",
			Name:        "Service",
			Description: "Service account for integrations",
			Permissions: []string{
				"agent:read",
				"task:read",
				"task:write",
				"sync:read",
				"sync:write",
			},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsDefault:   false,
			IsSystem:    true,
		},
	}

	for _, role := range systemRoles {
		ps.roles.Store(role.ID, &role)
	}
}

// CreateRole creates a new custom role
func (ps *PermissionSystem) CreateRole(id, name, description string, permissions []string) (*Role, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check if role already exists
	if _, ok := ps.roles.Load(id); ok {
		return nil, errors.New("role already exists")
	}

	now := time.Now()
	role := &Role{
		ID:          id,
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDefault:   false,
		IsSystem:    false,
	}

	ps.roles.Store(id, role)
	return role, nil
}

// GetRole retrieves a role by ID
func (ps *PermissionSystem) GetRole(roleID string) (*Role, error) {
	r, ok := ps.roles.Load(roleID)
	if !ok {
		return nil, ErrRoleNotFound
	}
	return r.(*Role), nil
}

// UpdateRole updates a role (system roles cannot be updated)
func (ps *PermissionSystem) UpdateRole(roleID string, updates map[string]interface{}) error {
	r, ok := ps.roles.Load(roleID)
	if !ok {
		return ErrRoleNotFound
	}

	role := r.(*Role)
	if role.IsSystem {
		return errors.New("cannot modify system role")
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				role.Name = v
			}
		case "description":
			if v, ok := value.(string); ok {
				role.Description = v
			}
		case "permissions":
			if v, ok := value.([]string); ok {
				role.Permissions = v
			}
		case "inherits":
			if v, ok := value.([]string); ok {
				role.Inherits = v
			}
		}
	}

	role.UpdatedAt = time.Now()
	ps.roles.Store(roleID, role)
	return nil
}

// DeleteRole deletes a role (system roles cannot be deleted)
func (ps *PermissionSystem) DeleteRole(roleID string) error {
	r, ok := ps.roles.Load(roleID)
	if !ok {
		return ErrRoleNotFound
	}

	role := r.(*Role)
	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.roles.Delete(roleID)
	return nil
}

// AssignRole assigns roles to an agent
func (ps *PermissionSystem) AssignRole(agentID string, roleIDs []string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Validate all roles exist
	for _, roleID := range roleIDs {
		if _, ok := ps.roles.Load(roleID); !ok {
			return ErrRoleNotFound
		}
	}

	ps.roleAssign.Store(agentID, roleIDs)
	return nil
}

// AddRoleToAgent adds a role to an agent's existing roles
func (ps *PermissionSystem) AddRoleToAgent(agentID, roleID string) error {
	// Validate role exists
	if _, ok := ps.roles.Load(roleID); !ok {
		return ErrRoleNotFound
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	var roles []string
	if r, ok := ps.roleAssign.Load(agentID); ok {
		roles = r.([]string)
	}

	// Check if role already assigned
	for _, r := range roles {
		if r == roleID {
			return nil // Already assigned
		}
	}

	roles = append(roles, roleID)
	ps.roleAssign.Store(agentID, roles)
	return nil
}

// RemoveRoleFromAgent removes a role from an agent
func (ps *PermissionSystem) RemoveRoleFromAgent(agentID, roleID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	roles, ok := ps.roleAssign.Load(agentID)
	if !ok {
		return nil
	}

	roleList := roles.([]string)
	newRoles := []string{}
	for _, r := range roleList {
		if r != roleID {
			newRoles = append(newRoles, r)
		}
	}

	ps.roleAssign.Store(agentID, newRoles)
	return nil
}

// GetAgentRoles returns all roles assigned to an agent
func (ps *PermissionSystem) GetAgentRoles(agentID string) ([]*Role, error) {
	roles, ok := ps.roleAssign.Load(agentID)
	if !ok {
		// Return default role
		defaultRole, _ := ps.GetRole("role_agent")
		return []*Role{defaultRole}, nil
	}

	roleIDs := roles.([]string)
	var result []*Role
	for _, roleID := range roleIDs {
		if r, ok := ps.roles.Load(roleID); ok {
			result = append(result, r.(*Role))
		}
	}

	return result, nil
}

// GetEffectivePermissions returns all effective permissions for an agent (including inherited)
func (ps *PermissionSystem) GetEffectivePermissions(agentID string) []string {
	roles, err := ps.GetAgentRoles(agentID)
	if err != nil || len(roles) == 0 {
		return []string{}
	}

	permissions := make(map[string]bool)
	visitedRoles := make(map[string]bool)

	// Recursively collect permissions including inherited
	for _, role := range roles {
		ps.collectPermissions(role, permissions, visitedRoles)
	}

	var result []string
	for perm := range permissions {
		result = append(result, perm)
	}

	return result
}

// collectPermissions recursively collects permissions from roles
func (ps *PermissionSystem) collectPermissions(role *Role, permissions map[string]bool, visitedRoles map[string]bool) {
	if visitedRoles[role.ID] {
		return
	}
	visitedRoles[role.ID] = true

	// Add role's permissions
	for _, perm := range role.Permissions {
		permissions[perm] = true
	}

	// Add inherited permissions
	for _, inheritID := range role.Inherits {
		if r, ok := ps.roles.Load(inheritID); ok {
			ps.collectPermissions(r.(*Role), permissions, visitedRoles)
		}
	}
}

// CheckPermission checks if an agent has a specific permission
func (ps *PermissionSystem) CheckPermission(agentID, permission string) bool {
	permissions := ps.GetEffectivePermissions(agentID)

	for _, p := range permissions {
		if p == "*" || p == permission {
			return true
		}

		// Check wildcard patterns
		if ps.matchPermission(p, permission) {
			return true
		}
	}

	return false
}

// CheckResourcePermission checks resource-level permission
func (ps *PermissionSystem) CheckResourcePermission(agentID, resourceType, resourceID, action string) bool {
	// First check role-based permissions
	perm := resourceType + ":" + action
	if ps.CheckPermission(agentID, perm) {
		return true
	}

	// Check self permissions
	if resourceID == agentID {
		selfPerm := resourceType + ":" + action + ":self"
		if ps.CheckPermission(agentID, selfPerm) {
			return true
		}
	}

	// Check resource-specific permissions
	resPerms, ok := ps.resources.Load(agentID)
	if !ok {
		return false
	}

	for _, rp := range resPerms.([]ResourcePermission) {
		if rp.ResourceType == resourceType {
			if rp.ResourceID == "*" || rp.ResourceID == resourceID {
				for _, a := range rp.Actions {
					if a == action || a == "*" {
						return true
					}
				}
			}
		}
	}

	return false
}

// matchPermission matches permission patterns (e.g., "agent:*" matches "agent:read")
func (ps *PermissionSystem) matchPermission(pattern, permission string) bool {
	if pattern == "*" {
		return true
	}

	// Split permissions
	patternParts := splitPermission(pattern)
	permParts := splitPermission(permission)

	if len(patternParts) != len(permParts) {
		return false
	}

	for i, p := range patternParts {
		if p == "*" {
			continue
		}
		if p != permParts[i] {
			return false
		}
	}

	return true
}

// splitPermission splits a permission string into parts
func splitPermission(perm string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(perm); i++ {
		if perm[i] == ':' {
			result = append(result, perm[start:i])
			start = i + 1
		}
	}
	result = append(result, perm[start:])
	return result
}

// AssignResourcePermission assigns resource-level permission to an agent
func (ps *PermissionSystem) AssignResourcePermission(agentID string, rp ResourcePermission) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var perms []ResourcePermission
	if p, ok := ps.resources.Load(agentID); ok {
		perms = p.([]ResourcePermission)
	}

	perms = append(perms, rp)
	ps.resources.Store(agentID, perms)
	return nil
}

// ClearResourcePermissions clears all resource permissions for an agent
func (ps *PermissionSystem) ClearResourcePermissions(agentID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.resources.Delete(agentID)
	return nil
}

// ListRoles returns all roles
func (ps *PermissionSystem) ListRoles() []*Role {
	var roles []*Role
	ps.roles.Range(func(key, value interface{}) bool {
		roles = append(roles, value.(*Role))
		return true
	})
	return roles
}

// PermissionMiddleware creates middleware for permission checking
type PermissionMiddleware struct {
	permissionSystem *PermissionSystem
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(ps *PermissionSystem) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionSystem: ps,
	}
}

// RequirePermission middleware checks for a specific permission
func (pm *PermissionMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			agentID := GetAgentID(ctx)

			if agentID == "" {
				http.Error(w, `{"error":"authentication required","code":101}`, http.StatusUnauthorized)
				return
			}

			if !pm.permissionSystem.CheckPermission(agentID, permission) {
				http.Error(w, `{"error":"permission denied: "+permission,"code":103}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireResourcePermission middleware checks for resource-level permission
func (pm *PermissionMiddleware) RequireResourcePermission(resourceType, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			agentID := GetAgentID(ctx)

			if agentID == "" {
				http.Error(w, `{"error":"authentication required","code":101}`, http.StatusUnauthorized)
				return
			}

			// Get resource ID from path parameter
			resourceID := getResourceIDFromPath(r.URL.Path)

			if !pm.permissionSystem.CheckResourcePermission(agentID, resourceType, resourceID, action) {
				http.Error(w, `{"error":"resource permission denied","code":103}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks for a specific role
func (pm *PermissionMiddleware) RequireRole(roleID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			agentID := GetAgentID(ctx)

			if agentID == "" {
				http.Error(w, `{"error":"authentication required","code":101}`, http.StatusUnauthorized)
				return
			}

			roles, err := pm.permissionSystem.GetAgentRoles(agentID)
			if err != nil {
				http.Error(w, `{"error":"role check failed","code":103}`, http.StatusForbidden)
				return
			}

			hasRole := false
			for _, role := range roles {
				if role.ID == roleID {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, `{"error":"required role not assigned","code":103}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware checks for admin role
func (pm *PermissionMiddleware) RequireAdmin() func(http.Handler) http.Handler {
	return pm.RequireRole("role_admin")
}

// getResourceIDFromPath extracts resource ID from URL path
func getResourceIDFromPath(path string) string {
	// Simple implementation - extract last segment after /
	parts := []string{}
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}

	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// Context helpers for permission system

// PermissionSystemKey is the context key for permission system
const PermissionSystemKey contextKey = "permission_system"

// WithPermissionSystem adds permission system to context
func WithPermissionSystem(ctx context.Context, ps *PermissionSystem) context.Context {
	return context.WithValue(ctx, PermissionSystemKey, ps)
}

// GetPermissionSystemFromContext retrieves permission system from context
func GetPermissionSystemFromContext(ctx context.Context) *PermissionSystem {
	if v := ctx.Value(PermissionSystemKey); v != nil {
		return v.(*PermissionSystem)
	}
	return nil
}

// Error definitions

var ErrRoleNotFound = errors.New("role not found")