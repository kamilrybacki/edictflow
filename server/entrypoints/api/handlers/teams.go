package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrNotFound = errors.New("not found")

type TeamService interface {
	Create(ctx context.Context, name string) (domain.Team, error)
	GetByID(ctx context.Context, id string) (domain.Team, error)
	List(ctx context.Context) ([]domain.Team, error)
	Update(ctx context.Context, team domain.Team) error
	Delete(ctx context.Context, id string) error
	// Invite methods
	CreateInvite(ctx context.Context, teamID, createdBy string, maxUses, expiresInHours int) (domain.TeamInvite, error)
	ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error)
	DeleteInvite(ctx context.Context, teamID, inviteID string) error
}

type TeamsHandler struct {
	service TeamService
}

func NewTeamsHandler(service TeamService) *TeamsHandler {
	return &TeamsHandler{service: service}
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type UpdateTeamSettingsRequest struct {
	DriftThresholdMinutes *int `json:"drift_threshold_minutes,omitempty"`
	// InheritGlobalRules is no longer configurable - teams always inherit global rules
}

type CreateInviteRequest struct {
	MaxUses        int `json:"max_uses"`
	ExpiresInHours int `json:"expires_in_hours,omitempty"`
}

type InviteResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	MaxUses   int    `json:"max_uses"`
	UseCount  int    `json:"use_count"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

func inviteToResponse(invite domain.TeamInvite) InviteResponse {
	return InviteResponse{
		ID:        invite.ID,
		Code:      invite.Code,
		MaxUses:   invite.MaxUses,
		UseCount:  invite.UseCount,
		ExpiresAt: invite.ExpiresAt.Format(time.RFC3339),
		CreatedAt: invite.CreatedAt.Format(time.RFC3339),
	}
}

func (h *TeamsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	team, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func (h *TeamsHandler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *TeamsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	team, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *TeamsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeamsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	team, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req UpdateTeamSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DriftThresholdMinutes != nil {
		team.Settings.DriftThresholdMinutes = *req.DriftThresholdMinutes
	}
	// InheritGlobalRules is always true - not configurable

	if err := h.service.Update(r.Context(), team); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *TeamsHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}
	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 24
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	invite, err := h.service.CreateInvite(r.Context(), teamID, userID, req.MaxUses, req.ExpiresInHours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inviteToResponse(invite))
}

func (h *TeamsHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	invites, err := h.service.ListInvites(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []InviteResponse
	for _, invite := range invites {
		response = append(response, inviteToResponse(invite))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TeamsHandler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")
	inviteID := chi.URLParam(r, "inviteId")

	if err := h.service.DeleteInvite(r.Context(), teamID, inviteID); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "invite not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TeamsHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}/settings", h.UpdateSettings)
	r.Delete("/{id}", h.Delete)

	// Invite routes
	r.Post("/{id}/invites", h.CreateInvite)
	r.Get("/{id}/invites", h.ListInvites)
	r.Delete("/{id}/invites/{inviteId}", h.DeleteInvite)
}
