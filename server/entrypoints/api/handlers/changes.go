package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
)

type ChangeService interface {
	GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error)
	ListByTeam(ctx context.Context, teamID string, filter ChangeRequestFilter) ([]domain.ChangeRequest, error)
	Approve(ctx context.Context, id, approverUserID string) error
	Reject(ctx context.Context, id, approverUserID string) error
}

type ChangeRequestFilter struct {
	Status          *domain.ChangeRequestStatus
	EnforcementMode *domain.EnforcementMode
	RuleID          *string
	AgentID         *string
	UserID          *string
	Limit           int
	Offset          int
}

type ChangesHandler struct {
	service ChangeService
}

func NewChangesHandler(service ChangeService) *ChangesHandler {
	return &ChangesHandler{service: service}
}

type ChangeRequestResponse struct {
	ID               string  `json:"id"`
	RuleID           string  `json:"rule_id"`
	AgentID          string  `json:"agent_id"`
	UserID           string  `json:"user_id"`
	TeamID           string  `json:"team_id"`
	FilePath         string  `json:"file_path"`
	OriginalHash     string  `json:"original_hash"`
	ModifiedHash     string  `json:"modified_hash"`
	DiffContent      string  `json:"diff_content"`
	Status           string  `json:"status"`
	EnforcementMode  string  `json:"enforcement_mode"`
	TimeoutAt        *string `json:"timeout_at,omitempty"`
	CreatedAt        string  `json:"created_at"`
	ResolvedAt       *string `json:"resolved_at,omitempty"`
	ResolvedByUserID *string `json:"resolved_by_user_id,omitempty"`
}

func changeRequestToResponse(cr domain.ChangeRequest) ChangeRequestResponse {
	resp := ChangeRequestResponse{
		ID:               cr.ID,
		RuleID:           cr.RuleID,
		AgentID:          cr.AgentID,
		UserID:           cr.UserID,
		TeamID:           cr.TeamID,
		FilePath:         cr.FilePath,
		OriginalHash:     cr.OriginalHash,
		ModifiedHash:     cr.ModifiedHash,
		DiffContent:      cr.DiffContent,
		Status:           string(cr.Status),
		EnforcementMode:  string(cr.EnforcementMode),
		CreatedAt:        cr.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ResolvedByUserID: cr.ResolvedByUserID,
	}

	if cr.TimeoutAt != nil {
		t := cr.TimeoutAt.Format("2006-01-02T15:04:05Z")
		resp.TimeoutAt = &t
	}
	if cr.ResolvedAt != nil {
		t := cr.ResolvedAt.Format("2006-01-02T15:04:05Z")
		resp.ResolvedAt = &t
	}

	return resp
}

func (h *ChangesHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id query parameter required", http.StatusBadRequest)
		return
	}

	filter := ChangeRequestFilter{}
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ChangeRequestStatus(status)
		filter.Status = &s
	}
	if ruleID := r.URL.Query().Get("rule_id"); ruleID != "" {
		filter.RuleID = &ruleID
	}

	changes, err := h.service.ListByTeam(r.Context(), teamID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []ChangeRequestResponse
	for _, cr := range changes {
		response = append(response, changeRequestToResponse(cr))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *ChangesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cr, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if cr == nil {
		http.Error(w, "change request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(changeRequestToResponse(*cr))
}

func (h *ChangesHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Approve(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ChangesHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Reject(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ChangesHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/reject", h.Reject)
}
