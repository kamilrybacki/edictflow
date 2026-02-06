package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
)

type NotificationService interface {
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	ListForUser(ctx context.Context, userID string, filter NotificationFilterParams) ([]domain.Notification, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context, userID string) error
}

type NotificationFilterParams struct {
	Type   *domain.NotificationType
	Unread *bool
	TeamID *string
	Limit  int
	Offset int
}

type NotificationsHandler struct {
	service NotificationService
}

func NewNotificationsHandler(service NotificationService) *NotificationsHandler {
	return &NotificationsHandler{service: service}
}

type NotificationResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	TeamID    *string                `json:"team_id,omitempty"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ReadAt    *string                `json:"read_at,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

func notificationToResponse(n domain.Notification) NotificationResponse {
	resp := NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		TeamID:    n.TeamID,
		Type:      string(n.Type),
		Title:     n.Title,
		Body:      n.Body,
		Metadata:  n.Metadata,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if n.ReadAt != nil {
		t := n.ReadAt.Format("2006-01-02T15:04:05Z")
		resp.ReadAt = &t
	}

	return resp
}

type UnreadCountResponse struct {
	Count int `json:"count"`
}

func (h *NotificationsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	filter := NotificationFilterParams{}
	if unread := r.URL.Query().Get("unread"); unread == "true" {
		b := true
		filter.Unread = &b
	}
	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		filter.TeamID = &teamID
	}

	notifications, err := h.service.ListForUser(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []NotificationResponse
	for _, n := range notifications {
		response = append(response, notificationToResponse(n))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *NotificationsHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UnreadCountResponse{Count: count})
}

func (h *NotificationsHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.MarkRead(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationsHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/unread-count", h.GetUnreadCount)
	r.Post("/{id}/read", h.MarkRead)
	r.Post("/read-all", h.MarkAllRead)
}
