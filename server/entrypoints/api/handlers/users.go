package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type UsersService interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error)
	Update(ctx context.Context, user domain.User) error
	Deactivate(ctx context.Context, id string) error
	GetWithRolesAndPermissions(ctx context.Context, id string) (domain.User, error)
	LeaveTeam(ctx context.Context, userID string) error
}

type UsersHandler struct {
	service UsersService
}

func NewUsersHandler(service UsersService) *UsersHandler {
	return &UsersHandler{service: service}
}

type UserResponse struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	AvatarURL     string   `json:"avatar_url,omitempty"`
	AuthProvider  string   `json:"auth_provider"`
	TeamID        *string  `json:"team_id,omitempty"`
	EmailVerified bool     `json:"email_verified"`
	IsActive      bool     `json:"is_active"`
	Permissions   []string `json:"permissions,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

type UserDetailResponse struct {
	UserResponse
	Roles []RoleResponse `json:"roles,omitempty"`
}

type RoleResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	HierarchyLevel int    `json:"hierarchy_level"`
	IsSystem       bool   `json:"is_system"`
}

type UpdateUserRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

func userToResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		AvatarURL:     user.AvatarURL,
		AuthProvider:  string(user.AuthProvider),
		TeamID:        user.TeamID,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		Permissions:   user.Permissions,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func userToDetailResponse(user domain.User) UserDetailResponse {
	resp := UserDetailResponse{
		UserResponse: userToResponse(user),
	}
	for _, role := range user.Roles {
		resp.Roles = append(resp.Roles, RoleResponse{
			ID:             role.ID,
			Name:           role.Name,
			Description:    role.Description,
			HierarchyLevel: role.HierarchyLevel,
			IsSystem:       role.IsSystem,
		})
	}
	return resp
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	var teamIDPtr *string
	if teamID != "" {
		teamIDPtr = &teamID
	}

	users, err := h.service.List(r.Context(), teamIDPtr, activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, userToResponse(user))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UsersHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetWithRolesAndPermissions(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userToDetailResponse(user))
}

func (h *UsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := h.service.Update(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	if err := h.service.Deactivate(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) LeaveTeam(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.LeaveTeam(r.Context(), userID); err != nil {
		if err.Error() == "user is not in a team" {
			http.Error(w, "not in a team", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UsersHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Deactivate)
	r.Post("/me/leave-team", h.LeaveTeam)
}
