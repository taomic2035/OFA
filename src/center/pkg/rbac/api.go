package rbac

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// APIHandler handles RBAC API requests
type APIHandler struct {
	manager *RBACManager
}

// NewAPIHandler creates a new RBAC API handler
func NewAPIHandler(manager *RBACManager) *APIHandler {
	return &APIHandler{
		manager: manager,
	}
}

// RegisterRoutes registers RBAC API routes
func (h *APIHandler) RegisterRoutes(r *mux.Router) {
	// Role management
	r.HandleFunc("/api/v1/rbac/roles", h.ListRolesHandler).Methods("GET")
	r.HandleFunc("/api/v1/rbac/roles", h.CreateRoleHandler).Methods("POST")
	r.HandleFunc("/api/v1/rbac/roles/{id}", h.GetRoleHandler).Methods("GET")
	r.HandleFunc("/api/v1/rbac/roles/{id}", h.UpdateRoleHandler).Methods("PUT")
	r.HandleFunc("/api/v1/rbac/roles/{id}", h.DeleteRoleHandler).Methods("DELETE")

	// User management
	r.HandleFunc("/api/v1/rbac/users", h.ListUsersHandler).Methods("GET")
	r.HandleFunc("/api/v1/rbac/users", h.CreateUserHandler).Methods("POST")
	r.HandleFunc("/api/v1/rbac/users/{id}", h.GetUserHandler).Methods("GET")
	r.HandleFunc("/api/v1/rbac/users/{id}", h.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/api/v1/rbac/users/{id}", h.DeleteUserHandler).Methods("DELETE")

	// Role assignment
	r.HandleFunc("/api/v1/rbac/users/{id}/roles", h.GetUserRolesHandler).Methods("GET")
	r.HandleFunc("/api/v1/rbac/users/{id}/roles/{role_id}", h.AssignRoleHandler).Methods("POST")
	r.HandleFunc("/api/v1/rbac/users/{id}/roles/{role_id}", h.RemoveRoleHandler).Methods("DELETE")

	// Permission check
	r.HandleFunc("/api/v1/rbac/check", h.CheckPermissionHandler).Methods("POST")
	r.HandleFunc("/api/v1/rbac/users/{id}/permissions", h.GetUserPermissionsHandler).Methods("GET")
}

// ListRolesHandler handles list roles request
func (h *APIHandler) ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	roles := h.manager.ListRoles()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	})
}

// CreateRoleHandler handles create role request
func (h *APIHandler) CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	var role Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.manager.CreateRole(&role); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, role)
}

// GetRoleHandler handles get role request
func (h *APIHandler) GetRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	role, ok := h.manager.GetRole(roleID)
	if !ok {
		writeError(w, http.StatusNotFound, "Role not found")
		return
	}

	writeJSON(w, http.StatusOK, role)
}

// UpdateRoleHandler handles update role request
func (h *APIHandler) UpdateRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	var role Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	role.ID = roleID

	if err := h.manager.UpdateRole(&role); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, role)
}

// DeleteRoleHandler handles delete role request
func (h *APIHandler) DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roleID := vars["id"]

	if err := h.manager.DeleteRole(roleID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Role deleted",
		"id":      roleID,
	})
}

// ListUsersHandler handles list users request
func (h *APIHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := h.manager.ListUsers()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// CreateUserHandler handles create user request
func (h *APIHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.manager.CreateUser(&user); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// GetUserHandler handles get user request
func (h *APIHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, ok := h.manager.GetUser(userID)
	if !ok {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateUserHandler handles update user request
func (h *APIHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user.ID = userID

	if err := h.manager.UpdateUser(&user); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUserHandler handles delete user request
func (h *APIHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := h.manager.DeleteUser(userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "User deleted",
		"id":      userID,
	})
}

// GetUserRolesHandler handles get user roles request
func (h *APIHandler) GetUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	roles := h.manager.GetUserRoles(userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"roles":   roles,
		"count":   len(roles),
	})
}

// AssignRoleHandler handles assign role request
func (h *APIHandler) AssignRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	roleID := vars["role_id"]

	if err := h.manager.AssignRole(userID, roleID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Role assigned",
		"user_id":  userID,
		"role_id":  roleID,
	})
}

// RemoveRoleHandler handles remove role request
func (h *APIHandler) RemoveRoleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	roleID := vars["role_id"]

	if err := h.manager.RemoveRole(userID, roleID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Role removed",
		"user_id":  userID,
		"role_id":  roleID,
	})
}

// CheckPermissionHandler handles permission check request
func (h *APIHandler) CheckPermissionHandler(w http.ResponseWriter, r *http.Request) {
	var check PermissionCheck
	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	allowed := h.manager.CheckPermission(&check)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":   check.UserID,
		"resource":  check.Resource,
		"action":    check.Action,
		"allowed":   allowed,
	})
}

// GetUserPermissionsHandler handles get user permissions request
func (h *APIHandler) GetUserPermissionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	permissions := h.manager.GetUserPermissions(userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   strings.ToLower(strings.Replace(http.StatusText(status), " ", "_", -1)),
		"message": message,
		"code":    status,
	})
}