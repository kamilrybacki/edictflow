package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
)

type RuleService interface {
	Create(ctx context.Context, req CreateRuleRequest) (domain.Rule, error)
	GetByID(ctx context.Context, id string) (domain.Rule, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error)
	ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error)
	ListByTargetLayer(ctx context.Context, targetLayer domain.TargetLayer) ([]domain.Rule, error)
	Update(ctx context.Context, rule domain.Rule) error
	Delete(ctx context.Context, id string) error
	GetMergedContent(ctx context.Context, targetLayer domain.TargetLayer) (string, error)
}

type RulesHandler struct {
	service   RuleService
	publisher publisher.Publisher
}

func NewRulesHandler(service RuleService, pub publisher.Publisher) *RulesHandler {
	return &RulesHandler{
		service:   service,
		publisher: pub,
	}
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

type RuleResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Content        string            `json:"content"`
	TargetLayer    string            `json:"target_layer"`
	PriorityWeight int               `json:"priority_weight"`
	Triggers       []TriggerResponse `json:"triggers"`
	TeamID         string            `json:"team_id"`
	Status         string            `json:"status"`
	CreatedBy      *string           `json:"created_by,omitempty"`
	SubmittedAt    string            `json:"submitted_at,omitempty"`
	ApprovedAt     string            `json:"approved_at,omitempty"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type TriggerResponse struct {
	Type         string   `json:"type"`
	Pattern      string   `json:"pattern,omitempty"`
	ContextTypes []string `json:"context_types,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

func ruleToResponse(rule domain.Rule) RuleResponse {
	resp := RuleResponse{
		ID:             rule.ID,
		Name:           rule.Name,
		Content:        rule.Content,
		TargetLayer:    string(rule.TargetLayer),
		PriorityWeight: rule.PriorityWeight,
		TeamID:         rule.TeamID,
		Status:         string(rule.Status),
		CreatedBy:      rule.CreatedBy,
		CreatedAt:      rule.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      rule.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if rule.SubmittedAt != nil {
		resp.SubmittedAt = rule.SubmittedAt.Format("2006-01-02T15:04:05Z")
	}
	if rule.ApprovedAt != nil {
		resp.ApprovedAt = rule.ApprovedAt.Format("2006-01-02T15:04:05Z")
	}

	for _, t := range rule.Triggers {
		resp.Triggers = append(resp.Triggers, TriggerResponse{
			Type:         string(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}

	return resp
}

func (h *RulesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Set created_by from authenticated user
	userID := middleware.GetUserID(r.Context())
	if userID != "" {
		// Note: CreateRuleRequest doesn't have CreatedBy, the service should handle this
	}

	rule, err := h.service.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish event asynchronously
	if h.publisher != nil {
		go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleCreated, rule.ID, rule.TeamID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ruleToResponse(rule))
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
	json.NewEncoder(w).Encode(ruleToResponse(rule))
}

func (h *RulesHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id query parameter required", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")

	var rules []domain.Rule
	var err error

	if status != "" {
		rules, err = h.service.ListByStatus(r.Context(), teamID, domain.RuleStatus(status))
	} else {
		rules, err = h.service.ListByTeam(r.Context(), teamID)
	}

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

func (h *RulesHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	// Only allow editing draft or rejected rules
	if rule.Status != domain.RuleStatusDraft && rule.Status != domain.RuleStatusRejected {
		http.Error(w, "can only edit draft or rejected rules", http.StatusConflict)
		return
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	rule.Name = req.Name
	rule.Content = req.Content
	rule.TargetLayer = domain.TargetLayer(req.TargetLayer)

	// Convert triggers
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish event asynchronously
	if h.publisher != nil {
		go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleUpdated, rule.ID, rule.TeamID)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	// Only allow deleting draft rules
	if rule.Status != domain.RuleStatusDraft {
		http.Error(w, "can only delete draft rules", http.StatusConflict)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish event asynchronously
	if h.publisher != nil {
		go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleDeleted, id, rule.TeamID)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RulesHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.ListByTeam)
	r.Get("/merged", h.GetMerged)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Patch("/{id}", h.UpdateEnforcement)
	r.Delete("/{id}", h.Delete)
}

// GetMerged returns the merged CLAUDE.md content for a target layer
func (h *RulesHandler) GetMerged(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		http.Error(w, "level query parameter required", http.StatusBadRequest)
		return
	}

	targetLayer := domain.TargetLayer(level)
	if !targetLayer.IsValid() {
		http.Error(w, "invalid level: must be enterprise, user, or project", http.StatusBadRequest)
		return
	}

	content, err := h.service.GetMergedContent(r.Context(), targetLayer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write([]byte(content))
}

type UpdateEnforcementRequest struct {
	EnforcementMode       string `json:"enforcement_mode"`
	TemporaryTimeoutHours *int   `json:"temporary_timeout_hours,omitempty"`
}

func (h *RulesHandler) UpdateEnforcement(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateEnforcementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Update enforcement mode
	if req.EnforcementMode != "" {
		mode := domain.EnforcementMode(req.EnforcementMode)
		if !mode.IsValid() {
			http.Error(w, "invalid enforcement_mode", http.StatusBadRequest)
			return
		}
		rule.EnforcementMode = mode
	}

	if req.TemporaryTimeoutHours != nil {
		rule.TemporaryTimeoutHours = *req.TemporaryTimeoutHours
	}

	if err := h.service.Update(r.Context(), rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish event asynchronously
	if h.publisher != nil {
		go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleUpdated, rule.ID, rule.TeamID)
	}

	w.WriteHeader(http.StatusNoContent)
}
