package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey type for context keys
type contextKey string

const (
	// AgentIDKey is the context key for agent ID
	AgentIDKey contextKey = "agent_id"
	// AgentTypeKey is the context key for agent type
	AgentTypeKey contextKey = "agent_type"
	// PermissionsKey is the context key for permissions
	PermissionsKey contextKey = "permissions"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtManager *JWTManager
	skipPaths  map[string]bool
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		skipPaths: map[string]bool{
			"/health":          true,
			"/metrics":         true,
			"/api/v1/auth/login":  true,
			"/api/v1/auth/register": true,
		},
	}
}

// SkipPath adds a path to skip authentication
func (m *AuthMiddleware) SkipPath(path string) {
	m.skipPaths[path] = true
}

// Middleware returns the HTTP middleware function
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain paths
		if m.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header","code":102}`, http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"invalid authorization header format","code":102}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token","code":102}`, http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, AgentIDKey, claims.AgentID)
		ctx = context.WithValue(ctx, AgentTypeKey, claims.AgentType)
		ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAgentID extracts agent ID from context
func GetAgentID(ctx context.Context) string {
	if v := ctx.Value(AgentIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetAgentType extracts agent type from context
func GetAgentType(ctx context.Context) int {
	if v := ctx.Value(AgentTypeKey); v != nil {
		return v.(int)
	}
	return 0
}

// GetPermissions extracts permissions from context
func GetPermissions(ctx context.Context) []string {
	if v := ctx.Value(PermissionsKey); v != nil {
		return v.([]string)
	}
	return nil
}

// HasPermission checks if the context has a specific permission
func HasPermission(ctx context.Context, permission string) bool {
	permissions := GetPermissions(ctx)
	for _, p := range permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// RequirePermission middleware checks for a specific permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasPermission(r.Context(), permission) {
				http.Error(w, `{"error":"permission denied","code":103}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}