package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/claudeception/server/domain"
)

type RuleService interface {
	Create(ctx context.Context, req CreateRuleRequest) (domain.Rule, error)
	GetByID(ctx context.Context, id string) (domain.Rule, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error)
	Update(ctx context.Context, rule domain.Rule) error
	Delete(ctx context.Context, id string) error
}

type RulesHandler struct {
	service RuleService
}

func NewRulesHandler(service RuleService) *RulesHandler {
	return &RulesHandler{service: service}
}

type TriggerRequest struct {
	Type         string   `json:"type"`
	Pattern      string   `json:"pattern,omitempty"`
	ContextTypes []string `json:"context_types,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type CreateRuleRequest struct {
	Name        string           `json:"name"`
	TargetLayer string           `json:"target_layer"`
	Content     string           `json:"content"`
	TeamID      string           `json:"team_id"`
	Triggers    []TriggerRequest `json:"triggers"`
}

func (h *RulesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rule, err := h.service.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (h *RulesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rule, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

func (h *RulesHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id query parameter required", http.StatusBadRequest)
		return
	}

	rules, err := h.service.ListByTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func (h *RulesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.ListByTeam)
	r.Get("/{id}", h.Get)
	r.Delete("/{id}", h.Delete)
}
