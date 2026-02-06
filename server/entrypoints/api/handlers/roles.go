package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type RolesService interface {
	Create(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.RoleEntity, error)
	GetByID(ctx context.Context, id string) (domain.RoleEntity, error)
	List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error)
	Update(ctx context.Context, role domain.RoleEntity) error
	Delete(ctx context.Context, id string) error
	GetRoleWithPermissions(ctx context.Context, id string) (domain.RoleEntity, error)
	AddPermission(ctx context.Context, roleID, permissionID string) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error
	RemoveUserRole(ctx context.Context, userID, roleID string) error
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
}

type RolesHandler struct {
	service RolesService
}

func NewRolesHandler(service RolesService) *RolesHandler {
	return &RolesHandler{service: service}
}

type CreateRoleRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	HierarchyLevel int     `json:"hierarchy_level"`
	ParentRoleID   *string `json:"parent_role_id,omitempty"`
	TeamID         *string `json:"team_id,omitempty"`
}

type UpdateRoleRequest struct {
	Name           string  `json:"name,omitempty"`
	Description    string  `json:"description,omitempty"`
	HierarchyLevel int     `json:"hierarchy_level,omitempty"`
	ParentRoleID   *string `json:"parent_role_id,omitempty"`
}

type RoleDetailResponse struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	HierarchyLevel int                  `json:"hierarchy_level"`
	ParentRoleID   *string              `json:"parent_role_id,omitempty"`
	TeamID         *string              `json:"team_id,omitempty"`
	IsSystem       bool                 `json:"is_system"`
	Permissions    []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt      string               `json:"created_at"`
}

type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type RolePermissionRequest struct {
	PermissionID string `json:"permission_id"`
}

type UserRoleRequest struct {
	UserID string `json:"user_id"`
}

func roleToResponse(role domain.RoleEntity) RoleResponse {
	return RoleResponse{
		ID:             role.ID,
		Name:           role.Name,
		Description:    role.Description,
		HierarchyLevel: role.HierarchyLevel,
		IsSystem:       role.IsSystem,
	}
}

func roleToDetailResponse(role domain.RoleEntity) RoleDetailResponse {
	resp := RoleDetailResponse{
		ID:             role.ID,
		Name:           role.Name,
		Description:    role.Description,
		HierarchyLevel: role.HierarchyLevel,
		ParentRoleID:   role.ParentRoleID,
		TeamID:         role.TeamID,
		IsSystem:       role.IsSystem,
		CreatedAt:      role.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	for _, p := range role.Permissions {
		resp.Permissions = append(resp.Permissions, PermissionResponse{
			ID:          p.ID,
			Code:        p.Code,
			Description: p.Description,
			Category:    string(p.Category),
		})
	}
	return resp
}

func (h *RolesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	role, err := h.service.Create(r.Context(), req.Name, req.Description, req.HierarchyLevel, req.ParentRoleID, req.TeamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(roleToResponse(role))
}

func (h *RolesHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")

	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	roles, err := h.service.List(r.Context(), teamIDPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []RoleResponse
	for _, role := range roles {
		response = append(response, roleToResponse(role))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RolesHandler) Get(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	role, err := h.service.GetRoleWithPermissions(r.Context(), roleID)
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roleToDetailResponse(role))
}

func (h *RolesHandler) Update(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	role, err := h.service.GetByID(r.Context(), roleID)
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.HierarchyLevel > 0 {
		role.HierarchyLevel = req.HierarchyLevel
	}
	if req.ParentRoleID != nil {
		role.ParentRoleID = req.ParentRoleID
	}

	if err := h.service.Update(r.Context(), role); err != nil {
		if err.Error() == "cannot modify system role" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), roleID); err != nil {
		if err.Error() == "cannot modify system role" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) AddPermission(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	var req RolePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddPermission(r.Context(), roleID, req.PermissionID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	permissionID := chi.URLParam(r, "permissionId")
	if roleID == "" || permissionID == "" {
		http.Error(w, "role id and permission id required", http.StatusBadRequest)
		return
	}

	if err := h.service.RemovePermission(r.Context(), roleID, permissionID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	if roleID == "" {
		http.Error(w, "role id required", http.StatusBadRequest)
		return
	}

	var req UserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AssignUserRole(r.Context(), req.UserID, roleID, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	if roleID == "" || userID == "" {
		http.Error(w, "role id and user id required", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveUserRole(r.Context(), userID, roleID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RolesHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.ListPermissions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []PermissionResponse
	for _, p := range permissions {
		response = append(response, PermissionResponse{
			ID:          p.ID,
			Code:        p.Code,
			Description: p.Description,
			Category:    string(p.Category),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RolesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/permissions", h.ListPermissions)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/permissions", h.AddPermission)
	r.Delete("/{id}/permissions/{permissionId}", h.RemovePermission)
	r.Post("/{id}/users", h.AssignUser)
	r.Delete("/{id}/users/{userId}", h.RemoveUser)
}
