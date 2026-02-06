package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/audit"
)

type AuditService interface {
	List(ctx context.Context, params audit.ListParams) ([]domain.AuditEntry, int, error)
	GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error)
}

type AuditHandler struct {
	service AuditService
}

func NewAuditHandler(service AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

type AuditEntryResponse struct {
	ID         string                        `json:"id"`
	EntityType string                        `json:"entity_type"`
	EntityID   string                        `json:"entity_id"`
	Action     string                        `json:"action"`
	ActorID    *string                       `json:"actor_id,omitempty"`
	ActorName  string                        `json:"actor_name,omitempty"`
	Changes    map[string]*domain.ChangeValue `json:"changes,omitempty"`
	Metadata   map[string]interface{}        `json:"metadata,omitempty"`
	CreatedAt  string                        `json:"created_at"`
}

type AuditListResponse struct {
	Entries []AuditEntryResponse `json:"entries"`
	Total   int                  `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
}

func auditEntryToResponse(entry domain.AuditEntry) AuditEntryResponse {
	return AuditEntryResponse{
		ID:         entry.ID,
		EntityType: string(entry.EntityType),
		EntityID:   entry.EntityID,
		Action:     string(entry.Action),
		ActorID:    entry.ActorID,
		ActorName:  entry.ActorName,
		Changes:    entry.Changes,
		Metadata:   entry.Metadata,
		CreatedAt:  entry.CreatedAt.Format(time.RFC3339),
	}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	params := audit.ListParams{}

	// Parse query parameters
	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		et := domain.AuditEntityType(entityType)
		params.EntityType = &et
	}
	if entityID := r.URL.Query().Get("entity_id"); entityID != "" {
		params.EntityID = &entityID
	}
	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		params.ActorID = &actorID
	}
	if action := r.URL.Query().Get("action"); action != "" {
		a := domain.AuditAction(action)
		params.Action = &a
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			params.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			params.To = &t
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
		}
	} else {
		params.Limit = 50 // Default limit
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	entries, total, err := h.service.List(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := AuditListResponse{
		Entries: make([]AuditEntryResponse, 0, len(entries)),
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}
	for _, entry := range entries {
		response.Entries = append(response.Entries, auditEntryToResponse(entry))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuditHandler) GetEntityHistory(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")

	if entityType == "" || entityID == "" {
		http.Error(w, "entity type and id required", http.StatusBadRequest)
		return
	}

	entries, err := h.service.GetEntityHistory(r.Context(), domain.AuditEntityType(entityType), entityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]AuditEntryResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, auditEntryToResponse(entry))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/entity/{entityType}/{entityId}", h.GetEntityHistory)
}
