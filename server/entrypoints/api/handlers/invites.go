package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type InviteService interface {
	JoinByCode(ctx context.Context, code, userID string) (domain.Team, error)
}

type InvitesHandler struct {
	service InviteService
}

func NewInvitesHandler(service InviteService) *InvitesHandler {
	return &InvitesHandler{service: service}
}

type JoinTeamResponse struct {
	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`
}

func (h *InvitesHandler) Join(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		http.Error(w, "invite code required", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	team, err := h.service.JoinByCode(r.Context(), code, userID)
	if err != nil {
		switch err.Error() {
		case "invite not found", "invite expired or max uses reached":
			http.Error(w, "invite not found or expired", http.StatusNotFound)
		case "user already in a team":
			http.Error(w, "leave current team first", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JoinTeamResponse{
		TeamID:   team.ID,
		TeamName: team.Name,
	})
}

func (h *InvitesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/{code}/join", h.Join)
}
