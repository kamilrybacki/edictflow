package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/library"
)

type LibraryService interface {
	Create(ctx context.Context, req library.CreateRequest) (domain.Rule, error)
	GetByID(ctx context.Context, id string) (domain.Rule, error)
	List(ctx context.Context) ([]domain.Rule, error)
	Update(ctx context.Context, rule domain.Rule) error
	Delete(ctx context.Context, id string) error
	Submit(ctx context.Context, id string) (domain.Rule, error)
	Approve(ctx context.Context, id, approvedBy string) (domain.Rule, error)
	Reject(ctx context.Context, id string) (domain.Rule, error)
}

type LibraryHandler struct {
	service LibraryService
}

func NewLibraryHandler(service LibraryService) *LibraryHandler {
	return &LibraryHandler{service: service}
}

type CreateLibraryRuleRequest struct {
	Name           string           `json:"name"`
	Content        string           `json:"content"`
	Description    string           `json:"description,omitempty"`
	TargetLayer    string           `json:"target_layer"`
	CategoryID     string           `json:"category_id,omitempty"`
	PriorityWeight int              `json:"priority_weight"`
	Overridable    bool             `json:"overridable"`
	Tags           []string         `json:"tags,omitempty"`
	Triggers       []TriggerRequest `json:"triggers,omitempty"`
}

func (h *LibraryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateLibraryRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	var triggers []domain.Trigger
	for _, t := range req.Triggers {
		triggers = append(triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}

	rule, err := h.service.Create(r.Context(), library.CreateRequest{
		Name:           req.Name,
		Content:        req.Content,
		Description:    req.Description,
		TargetLayer:    domain.TargetLayer(req.TargetLayer),
		CategoryID:     req.CategoryID,
		PriorityWeight: req.PriorityWeight,
		Overridable:    req.Overridable,
		Tags:           req.Tags,
		Triggers:       triggers,
		CreatedBy:      userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []RuleResponse
	for _, rule := range rules {
		response = append(response, ruleToResponse(rule))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LibraryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req CreateLibraryRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rule.Name = req.Name
	rule.Content = req.Content
	if req.Description != "" {
		rule.Description = &req.Description
	}
	rule.TargetLayer = domain.TargetLayer(req.TargetLayer)
	if req.CategoryID != "" {
		rule.CategoryID = &req.CategoryID
	}
	rule.PriorityWeight = req.PriorityWeight
	rule.Overridable = req.Overridable
	rule.Tags = req.Tags

	rule.Triggers = nil
	for _, t := range req.Triggers {
		rule.Triggers = append(rule.Triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}

	if err := h.service.Update(r.Context(), rule); err != nil {
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "can only edit draft or rejected rules", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "can only delete draft rules", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LibraryHandler) Submit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.Submit(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "rule cannot be submitted in current status", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	rule, err := h.service.Approve(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "only pending rules can be approved", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rule, err := h.service.Reject(r.Context(), id)
	if err != nil {
		if errors.Is(err, library.ErrRuleNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, library.ErrInvalidStatus) {
			http.Error(w, "only pending rules can be rejected", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *LibraryHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/submit", h.Submit)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/reject", h.Reject)
}
