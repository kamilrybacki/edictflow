package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/auth"
)

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (string, domain.User, error)
	Login(ctx context.Context, req auth.LoginRequest) (string, domain.User, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

type AuthHandler struct {
	authService AuthService
	userService UserService
}

func NewAuthHandler(authService AuthService, userService UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	TeamID   string `json:"team_id,omitempty"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string           `json:"token"`
	User  AuthUserResponse `json:"user"`
}

// AuthUserResponse matches frontend User interface with camelCase
type AuthUserResponse struct {
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	Name         string   `json:"name"`
	AvatarURL    string   `json:"avatarUrl,omitempty"`
	AuthProvider string   `json:"authProvider"`
	TeamID       *string  `json:"teamId,omitempty"`
	Permissions  []string `json:"permissions"`
	IsActive     bool     `json:"isActive"`
	CreatedAt    string   `json:"createdAt"`
	LastLoginAt  *string  `json:"lastLoginAt,omitempty"`
}

type UserProfileResponse struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	TeamID      *string  `json:"team_id,omitempty"`
	Permissions []string `json:"permissions"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Register(r.Context(), auth.RegisterRequest{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
		TeamID:   req.TeamID,
	})
	if err != nil {
		if err.Error() == "email already registered" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  authUserToResponse(user),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, user, err := h.authService.Login(r.Context(), auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token: token,
		User:  authUserToResponse(user),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For simple JWT auth, logout is client-side (discard token)
	// Could implement token blacklist here if needed
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserProfileResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		TeamID:      user.TeamID,
		Permissions: user.Permissions,
	})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
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

	if err := h.userService.Update(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/me", h.GetProfile)
	r.Put("/me", h.UpdateProfile)
	r.Put("/me/password", h.UpdatePassword)
}

func authUserToResponse(user domain.User) AuthUserResponse {
	var lastLogin *string
	if user.LastLoginAt != nil {
		t := user.LastLoginAt.Format("2006-01-02T15:04:05Z")
		lastLogin = &t
	}
	return AuthUserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		AuthProvider: string(user.AuthProvider),
		TeamID:       user.TeamID,
		Permissions:  user.Permissions,
		IsActive:     user.IsActive,
		CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		LastLoginAt:  lastLogin,
	}
}
