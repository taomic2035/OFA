package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Role represents a user role
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Inherits    []string     `json:"inherits"` // Role IDs to inherit from
}

// Permission represents a permission
type Permission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Resource    string   `json:"resource"`    // agents, tasks, messages, skills, workflows
	Action      Action   `json:"action"`      // create, read, update, delete, execute
	Conditions  []Condition `json:"conditions"` // Optional conditions
}

// Action represents permission action
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionExecute Action = "execute"
	ActionAdmin  Action = "admin"
)

// Condition represents a permission condition
type Condition struct {
	Type      ConditionType `json:"type"`    // owner, region, agent_type
	Value     string        `json:"value"`
}

// ConditionType defines condition types
type ConditionType string

const (
	ConditionOwner     ConditionType = "owner"
	ConditionRegion    ConditionType = "region"
	ConditionAgentType ConditionType = "agent_type"
	ConditionPriority  ConditionType = "priority"
)

// User represents a user with roles
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`     // Role IDs
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID     string
	Resource   string
	Action     Action
	ResourceID string            // Optional specific resource ID
	Context    map[string]string // Additional context for conditions
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	StoragePath string
	CacheEnabled bool
	CacheDuration time.Duration
}

// DefaultRBACConfig returns default configuration
func DefaultRBACConfig() *RBACConfig {
	return &RBACConfig{
		StoragePath:   "rbac",
		CacheEnabled:  true,
		CacheDuration: 5 * time.Minute,
	}
}

// RBACManager manages role-based access control
type RBACManager struct {
	config *RBACConfig

	// Storage
	roles sync.Map // map[string]*Role
	users sync.Map // map[string]*User

	// Cache for permission checks
	permissionCache sync.Map // map[string]*cachedPermission

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// cachedPermission represents a cached permission result
type cachedPermission struct {
	allowed   bool
	cachedAt  time.Time
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(config *RBACConfig) (*RBACManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &RBACManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Load existing data
	if err := manager.loadRoles(); err != nil {
		// Create default roles if not exists
		manager.createDefaultRoles()
	}

	if err := manager.loadUsers(); err != nil {
		// Ignore error - users may not exist yet
	}

	return manager, nil
}

// Start begins RBAC operations
func (m *RBACManager) Start() {
	go m.cacheCleanLoop()
}

// Stop stops the RBAC manager
func (m *RBACManager) Stop() {
	m.cancel()
}

// loadRoles loads roles from storage
func (m *RBACManager) loadRoles() error {
	if m.config.StoragePath == "" {
		return nil
	}

	path := filepath.Join(m.config.StoragePath, "roles.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var roles []*Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return err
	}

	for _, role := range roles {
		m.roles.Store(role.ID, role)
	}

	return nil
}

// loadUsers loads users from storage
func (m *RBACManager) loadUsers() error {
	if m.config.StoragePath == "" {
		return nil
	}

	path := filepath.Join(m.config.StoragePath, "users.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	for _, user := range users {
		m.users.Store(user.ID, user)
	}

	return nil
}

// createDefaultRoles creates default system roles
func (m *RBACManager) createDefaultRoles() {
	now := time.Now()

	// Admin role - full access
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system access",
		Permissions: []Permission{
			{ID: "admin-all", Name: "All Access", Resource: "*", Action: ActionAdmin},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.roles.Store("admin", adminRole)

	// Operator role - manage agents and tasks
	operatorRole := &Role{
		ID:          "operator",
		Name:        "Operator",
		Description: "Manage agents and tasks",
		Permissions: []Permission{
			{ID: "agent-read", Name: "Read Agents", Resource: "agents", Action: ActionRead},
			{ID: "agent-create", Name: "Create Agents", Resource: "agents", Action: ActionCreate},
			{ID: "agent-update", Name: "Update Agents", Resource: "agents", Action: ActionUpdate},
			{ID: "agent-delete", Name: "Delete Agents", Resource: "agents", Action: ActionDelete},
			{ID: "task-read", Name: "Read Tasks", Resource: "tasks", Action: ActionRead},
			{ID: "task-create", Name: "Create Tasks", Resource: "tasks", Action: ActionCreate},
			{ID: "task-execute", Name: "Execute Tasks", Resource: "tasks", Action: ActionExecute},
			{ID: "message-read", Name: "Read Messages", Resource: "messages", Action: ActionRead},
			{ID: "message-create", Name: "Create Messages", Resource: "messages", Action: ActionCreate},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.roles.Store("operator", operatorRole)

	// Developer role - develop and test skills
	devRole := &Role{
		ID:          "developer",
		Name:        "Developer",
		Description: "Develop and test skills",
		Permissions: []Permission{
			{ID: "skill-read", Name: "Read Skills", Resource: "skills", Action: ActionRead},
			{ID: "skill-create", Name: "Create Skills", Resource: "skills", Action: ActionCreate},
			{ID: "skill-update", Name: "Update Skills", Resource: "skills", Action: ActionUpdate},
			{ID: "task-read", Name: "Read Tasks", Resource: "tasks", Action: ActionRead},
			{ID: "task-execute", Name: "Execute Tasks", Resource: "tasks", Action: ActionExecute},
			{ID: "workflow-read", Name: "Read Workflows", Resource: "workflows", Action: ActionRead},
			{ID: "workflow-create", Name: "Create Workflows", Resource: "workflows", Action: ActionCreate},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.roles.Store("developer", devRole)

	// Viewer role - read-only access
	viewerRole := &Role{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access",
		Permissions: []Permission{
			{ID: "agent-read", Name: "Read Agents", Resource: "agents", Action: ActionRead},
			{ID: "task-read", Name: "Read Tasks", Resource: "tasks", Action: ActionRead},
			{ID: "message-read", Name: "Read Messages", Resource: "messages", Action: ActionRead},
			{ID: "skill-read", Name: "Read Skills", Resource: "skills", Action: ActionRead},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.roles.Store("viewer", viewerRole)

	// Agent role - for agent entities
	agentRole := &Role{
		ID:          "agent",
		Name:        "Agent",
		Description: "Agent entity role",
		Permissions: []Permission{
			{ID: "task-read", Name: "Read Tasks", Resource: "tasks", Action: ActionRead},
			{ID: "task-execute", Name: "Execute Tasks", Resource: "tasks", Action: ActionExecute},
			{ID: "message-read", Name: "Read Messages", Resource: "messages", Action: ActionRead},
			{ID: "message-create", Name: "Create Messages", Resource: "messages", Action: ActionCreate},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.roles.Store("agent", agentRole)

	m.saveRoles()
}

// saveRoles saves roles to storage
func (m *RBACManager) saveRoles() error {
	if m.config.StoragePath == "" {
		return nil
	}

	// Ensure directory exists
	os.MkdirAll(m.config.StoragePath, 0755)

	var roles []*Role
	m.roles.Range(func(key, value interface{}) bool {
		roles = append(roles, value.(*Role))
		return true
	})

	data, err := json.Marshal(roles)
	if err != nil {
		return err
	}

	path := filepath.Join(m.config.StoragePath, "roles.json")
	return os.WriteFile(path, data, 0644)
}

// saveUsers saves users to storage
func (m *RBACManager) saveUsers() error {
	if m.config.StoragePath == "" {
		return nil
	}

	os.MkdirAll(m.config.StoragePath, 0755)

	var users []*User
	m.users.Range(func(key, value interface{}) bool {
		users = append(users, value.(*User))
		return true
	})

	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	path := filepath.Join(m.config.StoragePath, "users.json")
	return os.WriteFile(path, data, 0644)
}

// CreateRole creates a new role
func (m *RBACManager) CreateRole(role *Role) error {
	if role.ID == "" {
		return errors.New("role ID required")
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	m.roles.Store(role.ID, role)
	m.saveRoles()

	// Invalidate cache
	m.invalidateCache()

	return nil
}

// UpdateRole updates an existing role
func (m *RBACManager) UpdateRole(role *Role) error {
	if _, ok := m.roles.Load(role.ID); !ok {
		return fmt.Errorf("role not found: %s", role.ID)
	}

	role.UpdatedAt = time.Now()
	m.roles.Store(role.ID, role)
	m.saveRoles()
	m.invalidateCache()

	return nil
}

// DeleteRole deletes a role
func (m *RBACManager) DeleteRole(roleID string) error {
	m.roles.Delete(roleID)
	m.saveRoles()
	m.invalidateCache()
	return nil
}

// GetRole returns a role by ID
func (m *RBACManager) GetRole(roleID string) (*Role, bool) {
	if v, ok := m.roles.Load(roleID); ok {
		return v.(*Role), true
	}
	return nil, false
}

// ListRoles returns all roles
func (m *RBACManager) ListRoles() []*Role {
	var roles []*Role
	m.roles.Range(func(key, value interface{}) bool {
		roles = append(roles, value.(*Role))
		return true
	})
	return roles
}

// CreateUser creates a new user
func (m *RBACManager) CreateUser(user *User) error {
	if user.ID == "" {
		return errors.New("user ID required")
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	m.users.Store(user.ID, user)
	m.saveUsers()
	m.invalidateCache()

	return nil
}

// UpdateUser updates an existing user
func (m *RBACManager) UpdateUser(user *User) error {
	if _, ok := m.users.Load(user.ID); !ok {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	user.UpdatedAt = time.Now()
	m.users.Store(user.ID, user)
	m.saveUsers()
	m.invalidateCache()

	return nil
}

// DeleteUser deletes a user
func (m *RBACManager) DeleteUser(userID string) error {
	m.users.Delete(userID)
	m.saveUsers()
	m.invalidateCache()
	return nil
}

// GetUser returns a user by ID
func (m *RBACManager) GetUser(userID string) (*User, bool) {
	if v, ok := m.users.Load(userID); ok {
		return v.(*User), true
	}
	return nil, false
}

// AssignRole assigns a role to a user
func (m *RBACManager) AssignRole(userID, roleID string) error {
	user, ok := m.GetUser(userID)
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	role, ok := m.GetRole(roleID)
	if !ok {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if already assigned
	for _, r := range user.Roles {
		if r == roleID {
			return nil // Already assigned
		}
	}

	user.Roles = append(user.Roles, roleID)
	user.UpdatedAt = time.Now()
	m.users.Store(userID, user)
	m.saveUsers()
	m.invalidateCache()

	return nil
}

// RemoveRole removes a role from a user
func (m *RBACManager) RemoveRole(userID, roleID string) error {
	user, ok := m.GetUser(userID)
	if !ok {
		return fmt.Errorf("user not found: %s", userID)
	}

	for i, r := range user.Roles {
		if r == roleID {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			user.UpdatedAt = time.Now()
			m.users.Store(userID, user)
			m.saveUsers()
			m.invalidateCache()
			return nil
		}
	}

	return nil
}

// GetUserRoles returns all roles for a user
func (m *RBACManager) GetUserRoles(userID string) []*Role {
	user, ok := m.GetUser(userID)
	if !ok {
		return nil
	}

	var roles []*Role
	for _, roleID := range user.Roles {
		if role, ok := m.GetRole(roleID); ok {
			roles = append(roles, role)
		}
	}

	return roles
}

// CheckPermission checks if a user has a permission
func (m *RBACManager) CheckPermission(check *PermissionCheck) bool {
	// Check cache
	if m.config.CacheEnabled {
		cacheKey := fmt.Sprintf("%s:%s:%s", check.UserID, check.Resource, check.Action)
		if v, ok := m.permissionCache.Load(cacheKey); ok {
			cached := v.(*cachedPermission)
			if time.Since(cached.cachedAt) < m.config.CacheDuration {
				return cached.allowed
			}
		}
	}

	allowed := m.checkPermissionInternal(check)

	// Cache result
	if m.config.CacheEnabled {
		cacheKey := fmt.Sprintf("%s:%s:%s", check.UserID, check.Resource, check.Action)
		m.permissionCache.Store(cacheKey, &cachedPermission{
			allowed:  allowed,
			cachedAt: time.Now(),
		})
	}

	return allowed
}

// checkPermissionInternal performs the actual permission check
func (m *RBACManager) checkPermissionInternal(check *PermissionCheck) bool {
	user, ok := m.GetUser(check.UserID)
	if !ok || !user.Active {
		return false
	}

	// Get all permissions from user's roles
	permissions := m.collectPermissions(user)

	for _, perm := range permissions {
		if m.matchPermission(perm, check) {
			return true
		}
	}

	return false
}

// collectPermissions collects all permissions from user's roles
func (m *RBACManager) collectPermissions(user *User) []Permission {
	var permissions []Permission
	visitedRoles := make(map[string]bool)

	for _, roleID := range user.Roles {
		m.collectRolePermissions(roleID, &permissions, visitedRoles)
	}

	return permissions
}

// collectRolePermissions recursively collects permissions from a role
func (m *RBACManager) collectRolePermissions(roleID string, permissions *[]Permission, visited map[string]bool) {
	if visited[roleID] {
		return // Prevent circular inheritance
	}
	visited[roleID] = true

	role, ok := m.GetRole(roleID)
	if !ok {
		return
	}

	// Add direct permissions
	*permissions = append(*permissions, role.Permissions...)

	// Add inherited permissions
	for _, inheritedID := range role.Inherits {
		m.collectRolePermissions(inheritedID, permissions, visited)
	}
}

// matchPermission checks if a permission matches the check
func (m *RBACManager) matchPermission(perm Permission, check *PermissionCheck) bool {
	// Check resource match
	if perm.Resource != "*" && perm.Resource != check.Resource {
		return false
	}

	// Check action match
	if perm.Action != ActionAdmin && perm.Action != check.Action {
		return false
	}

	// Check conditions
	for _, cond := range perm.Conditions {
		if !m.checkCondition(cond, check) {
			return false
		}
	}

	return true
}

// checkCondition checks a permission condition
func (m *RBACManager) checkCondition(cond Condition, check *PermissionCheck) bool {
	switch cond.Type {
	case ConditionOwner:
		if check.Context["owner"] != check.UserID {
			return false
		}
	case ConditionRegion:
		if check.Context["region"] != cond.Value {
			return false
		}
	case ConditionAgentType:
		if check.Context["agent_type"] != cond.Value {
			return false
		}
	case ConditionPriority:
		// Priority condition - check if priority is sufficient
		priority := check.Context["priority"]
		if priority < cond.Value {
			return false
		}
	}
	return true
}

// invalidateCache invalidates the permission cache
func (m *RBACManager) invalidateCache() {
	m.permissionCache = sync.Map{}
}

// cacheCleanLoop periodically cleans expired cache entries
func (m *RBACManager) cacheCleanLoop() {
	ticker := time.NewTicker(m.config.CacheDuration)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanExpiredCache()
		}
	}
}

// cleanExpiredCache removes expired cache entries
func (m *RBACManager) cleanExpiredCache() {
	m.permissionCache.Range(func(key, value interface{}) bool {
		cached := value.(*cachedPermission)
		if time.Since(cached.cachedAt) > m.config.CacheDuration {
			m.permissionCache.Delete(key)
		}
		return true
	})
}

// ListUsers returns all users
func (m *RBACManager) ListUsers() []*User {
	var users []*User
	m.users.Range(func(key, value interface{}) bool {
		users = append(users, value.(*User))
		return true
	})
	return users
}

// GetUserPermissions returns all permissions for a user
func (m *RBACManager) GetUserPermissions(userID string) []Permission {
	user, ok := m.GetUser(userID)
	if !ok {
		return nil
	}
	return m.collectPermissions(user)
}

// HasRole checks if user has a specific role
func (m *RBACManager) HasRole(userID, roleID string) bool {
	user, ok := m.GetUser(userID)
	if !ok {
		return false
	}

	for _, r := range user.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// IsAdmin checks if user has admin role
func (m *RBACManager) IsAdmin(userID string) bool {
	return m.HasRole(userID, "admin")
}