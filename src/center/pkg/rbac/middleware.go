package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	Manager         *RBACManager
	PublicPaths     []string // Paths that don't require authentication
	SkipPaths       []string // Paths that skip RBAC check
	DefaultRole     string   // Default role for authenticated users without roles
	ContextKeyUser  string   // Context key for user ID
	ContextKeyRoles string   // Context key for roles
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig(manager *RBACManager) *MiddlewareConfig {
	return &MiddlewareConfig{
		Manager:         manager,
		PublicPaths:     []string{"/health", "/metrics", "/auth/login", "/auth/register"},
		SkipPaths:       []string{},
		DefaultRole:     "viewer",
		ContextKeyUser:  "user_id",
		ContextKeyRoles: "roles",
	}
}

// AuthMiddleware creates authentication middleware
func AuthMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check public paths
			for _, path := range config.PublicPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract user ID from header or token
			userID := extractUserID(r)
			if userID == "" {
				writeUnauthorized(w, "Missing authentication")
				return
			}

			// Get user
			user, ok := config.Manager.GetUser(userID)
			if !ok || !user.Active {
				writeUnauthorized(w, "User not found or inactive")
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), config.ContextKeyUser, userID)
			ctx = context.WithValue(ctx, config.ContextKeyRoles, user.Roles)

			// Assign default role if user has no roles
			if len(user.Roles) == 0 && config.DefaultRole != "" {
				config.Manager.AssignRole(userID, config.DefaultRole)
				ctx = context.WithValue(ctx, config.ContextKeyRoles, []string{config.DefaultRole})
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RBACMiddleware creates RBAC permission check middleware
func RBACMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check skip paths
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Get user from context
			userID, ok := getUserFromContext(r.Context(), config.ContextKeyUser)
			if !ok {
				writeUnauthorized(w, "User not in context")
				return
			}

			// Determine resource and action from request
			resource, action := parseRequest(r)

			// Check permission
			check := &PermissionCheck{
				UserID:   userID,
				Resource: resource,
				Action:   action,
				Context:  extractContext(r),
			}

			if !config.Manager.CheckPermission(check) {
				writeForbidden(w, userID, resource, action)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminMiddleware creates admin-only middleware
func AdminMiddleware(config *MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getUserFromContext(r.Context(), config.ContextKeyUser)
			if !ok {
				writeUnauthorized(w, "User not in context")
				return
			}

			if !config.Manager.IsAdmin(userID) {
				writeForbidden(w, userID, "*", ActionAdmin)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware creates role-specific middleware
func RoleMiddleware(config *MiddlewareConfig, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getUserFromContext(r.Context(), config.ContextKeyUser)
			if !ok {
				writeUnauthorized(w, "User not in context")
				return
			}

			if !config.Manager.HasRole(userID, requiredRole) {
				writeForbidden(w, userID, "role", ActionRead)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ResourceOwnerMiddleware creates middleware that checks resource ownership
func ResourceOwnerMiddleware(config *MiddlewareConfig, resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getUserFromContext(r.Context(), config.ContextKeyUser)
			if !ok {
				writeUnauthorized(w, "User not in context")
				return
			}

			// Admin always has access
			if config.Manager.IsAdmin(userID) {
				next.ServeHTTP(w, r)
				return
			}

			// Get resource ID from URL
			resourceID := extractResourceID(r)
			if resourceID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check ownership
			ownerID := r.Header.Get("X-Resource-Owner")
			if ownerID != userID {
				writeForbidden(w, userID, resource, ActionRead)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractUserID extracts user ID from request
func extractUserID(r *http.Request) string {
	// Check Authorization header (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// In real implementation, would validate JWT token
		// For now, use token as user ID for simplicity
		return token
	}

	// Check X-User-ID header (for internal services)
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		return userID
	}

	// Check query parameter
	userID = r.URL.Query().Get("user_id")
	if userID != "" {
		return userID
	}

	return ""
}

// parseRequest determines resource and action from HTTP request
func parseRequest(r *http.Request) (string, Action) {
	path := r.URL.Path
	method := r.Method

	// Parse resource from path
	// Example: /api/v1/agents -> resource: "agents"
	// Example: /api/v1/tasks/123 -> resource: "tasks"
	var resource string
	pathParts := strings.Split(path, "/")
	for i, part := range pathParts {
		if part == "api" || part == "v1" || part == "" {
			continue
		}
		// Resource is usually the first non-empty part after api/v1
		if i > 2 && resource == "" {
			resource = part
			break
		}
	}

	// Default resource if not found
	if resource == "" {
		resource = "system"
	}

	// Map HTTP method to action
	action := methodToAction(method)

	return resource, action
}

// methodToAction maps HTTP method to permission action
func methodToAction(method string) Action {
	switch method {
	case "GET":
		return ActionRead
	case "POST":
		return ActionCreate
	case "PUT", "PATCH":
		return ActionUpdate
	case "DELETE":
		return ActionDelete
	default:
		return ActionRead
	}
}

// extractContext extracts additional context from request
func extractContext(r *http.Request) map[string]string {
	ctx := make(map[string]string)

	// Extract common context values
	ctx["owner"] = r.Header.Get("X-Resource-Owner")
	ctx["region"] = r.Header.Get("X-Region")
	ctx["agent_type"] = r.Header.Get("X-Agent-Type")
	ctx["priority"] = r.Header.Get("X-Priority")

	// Extract resource ID
	ctx["resource_id"] = extractResourceID(r)

	return ctx
}

// extractResourceID extracts resource ID from URL path
func extractResourceID(r *http.Request) string {
	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	// Resource ID is usually the last numeric/UUID part
	for i := len(pathParts) - 1; i >= 0; i-- {
		if pathParts[i] != "" && !strings.HasPrefix(pathParts[i], "api") {
			return pathParts[i]
		}
	}

	return ""
}

// getUserFromContext gets user ID from context
func getUserFromContext(ctx context.Context, key string) (string, bool) {
	if v := ctx.Value(key); v != nil {
		if userID, ok := v.(string); ok {
			return userID, true
		}
	}
	return "", false
}

// writeUnauthorized writes unauthorized response
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"error":  "unauthorized",
		"message": message,
		"code":    401,
	}
	json.NewEncoder(w).Encode(response)
}

// writeForbidden writes forbidden response
func writeForbidden(w http.ResponseWriter, userID, resource string, action Action) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	response := map[string]interface{}{
		"error":   "forbidden",
		"message": fmt.Sprintf("User %s does not have permission to %s on %s", userID, action, resource),
		"code":    403,
		"user_id": userID,
		"resource": resource,
		"action":  action,
	}
	json.NewEncoder(w).Encode(response)
}

// PermissionError represents a permission error
type PermissionError struct {
	UserID   string
	Resource string
	Action   Action
	Message  string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: user %s cannot %s on %s", e.UserID, e.Action, e.Resource)
}

// NewPermissionError creates a permission error
func NewPermissionError(userID, resource string, action Action) error {
	return &PermissionError{
		UserID:   userID,
		Resource: resource,
		Action:   action,
		Message:  fmt.Sprintf("Permission denied for %s on %s", action, resource),
	}
}

// CheckPermissionHandler is a convenience function for handlers
func CheckPermissionHandler(config *MiddlewareConfig, userID, resource string, action Action) error {
	check := &PermissionCheck{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}

	if !config.Manager.CheckPermission(check) {
		return NewPermissionError(userID, resource, action)
	}

	return nil
}

// RequirePermission is a helper for permission checks
func RequirePermission(manager *RBACManager, userID, resource string, action Action) error {
	check := &PermissionCheck{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}

	if !manager.CheckPermission(check) {
		return errors.New("permission denied")
	}

	return nil
}

// RequireAdmin is a helper for admin checks
func RequireAdmin(manager *RBACManager, userID string) error {
	if !manager.IsAdmin(userID) {
		return errors.New("admin permission required")
	}
	return nil
}

// RequireRole is a helper for role checks
func RequireRole(manager *RBACManager, userID, roleID string) error {
	if !manager.HasRole(userID, roleID) {
		return fmt.Errorf("role %s required", roleID)
	}
	return nil
}