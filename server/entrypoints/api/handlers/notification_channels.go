package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type NotificationChannelService interface {
	GetChannel(ctx context.Context, id string) (*domain.NotificationChannel, error)
	ListChannels(ctx context.Context, teamID string) ([]domain.NotificationChannel, error)
	CreateChannel(ctx context.Context, nc domain.NotificationChannel) error
	UpdateChannel(ctx context.Context, nc domain.NotificationChannel) error
	DeleteChannel(ctx context.Context, id string) error
	TestChannel(ctx context.Context, id string) error
}

type NotificationChannelsHandler struct {
	service NotificationChannelService
}

func NewNotificationChannelsHandler(service NotificationChannelService) *NotificationChannelsHandler {
	return &NotificationChannelsHandler{service: service}
}

type CreateChannelRequest struct {
	TeamID      string                 `json:"team_id"`
	ChannelType string                 `json:"channel_type"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

type UpdateChannelRequest struct {
	ChannelType string                 `json:"channel_type"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

type NotificationChannelResponse struct {
	ID          string                 `json:"id"`
	TeamID      string                 `json:"team_id"`
	ChannelType string                 `json:"channel_type"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   string                 `json:"created_at"`
}

func notificationChannelToResponse(nc domain.NotificationChannel) NotificationChannelResponse {
	return NotificationChannelResponse{
		ID:          nc.ID,
		TeamID:      nc.TeamID,
		ChannelType: string(nc.ChannelType),
		Config:      nc.Config,
		Enabled:     nc.Enabled,
		CreatedAt:   nc.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *NotificationChannelsHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id query parameter required", http.StatusBadRequest)
		return
	}

	channels, err := h.service.ListChannels(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []NotificationChannelResponse
	for _, nc := range channels {
		response = append(response, notificationChannelToResponse(nc))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *NotificationChannelsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	nc := domain.NewNotificationChannel(
		req.TeamID,
		domain.ChannelType(req.ChannelType),
		req.Config,
	)
	nc.Enabled = req.Enabled

	if err := h.service.CreateChannel(r.Context(), nc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(notificationChannelToResponse(nc))
}

func (h *NotificationChannelsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.service.GetChannel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "notification channel not found", http.StatusNotFound)
		return
	}

	var req UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	existing.ChannelType = domain.ChannelType(req.ChannelType)
	existing.Config = req.Config
	existing.Enabled = req.Enabled

	if err := h.service.UpdateChannel(r.Context(), *existing); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationChannelsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteChannel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationChannelsHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.TestChannel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationChannelsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/test", h.Test)
}
