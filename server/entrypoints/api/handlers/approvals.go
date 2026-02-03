package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/claudeception/server/services/approvals"
)

type ApprovalsService interface {
	SubmitRule(ctx context.Context, ruleID string) error
	ApproveRule(ctx context.Context, ruleID, userID, comment string) error
	RejectRule(ctx context.Context, ruleID, userID, comment string) error
	GetApprovalStatus(ctx context.Context, ruleID string) (approvals.ApprovalStatus, error)
	GetPendingRules(ctx context.Context, teamID string) ([]domain.Rule, error)
	GetPendingRulesByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error)
	ResetRule(ctx context.Context, ruleID string) error
}

type ApprovalsHandler struct {
	service ApprovalsService
}

func NewApprovalsHandler(service ApprovalsService) *ApprovalsHandler {
	return &ApprovalsHandler{service: service}
}

type ApprovalDecisionRequest struct {
	Comment string `json:"comment,omitempty"`
}

type ApprovalStatusResponse struct {
	RuleID        string                   `json:"rule_id"`
	Status        string                   `json:"status"`
	RequiredCount int                      `json:"required_count"`
	CurrentCount  int                      `json:"current_count"`
	Approvals     []ApprovalRecordResponse `json:"approvals"`
}

type ApprovalRecordResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name,omitempty"`
	Decision  string `json:"decision"`
	Comment   string `json:"comment,omitempty"`
	CreatedAt string `json:"created_at"`
}

type PendingRuleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TargetLayer string  `json:"target_layer"`
	TeamID      string  `json:"team_id"`
	CreatedBy   *string `json:"created_by,omitempty"`
	SubmittedAt string  `json:"submitted_at,omitempty"`
}

func (h *ApprovalsHandler) Submit(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, "rule id required", http.StatusBadRequest)
		return
	}

	if err := h.service.SubmitRule(r.Context(), ruleID); err != nil {
		switch err.Error() {
		case "rule not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "rule cannot be submitted in current state":
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, "rule id required", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ApprovalDecisionRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.ApproveRule(r.Context(), ruleID, userID, req.Comment); err != nil {
		switch err.Error() {
		case "rule not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "rule is not pending approval":
			http.Error(w, err.Error(), http.StatusConflict)
		case "user does not have permission to approve this rule":
			http.Error(w, err.Error(), http.StatusForbidden)
		case "user has already voted on this rule":
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) Reject(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, "rule id required", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ApprovalDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Comment == "" {
		http.Error(w, "comment required for rejection", http.StatusBadRequest)
		return
	}

	if err := h.service.RejectRule(r.Context(), ruleID, userID, req.Comment); err != nil {
		switch err.Error() {
		case "rule not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "rule is not pending approval":
			http.Error(w, err.Error(), http.StatusConflict)
		case "user does not have permission to approve this rule":
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, "rule id required", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetApprovalStatus(r.Context(), ruleID)
	if err != nil {
		if err.Error() == "rule not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ApprovalStatusResponse{
		RuleID:        status.RuleID,
		Status:        string(status.Status),
		RequiredCount: status.RequiredCount,
		CurrentCount:  status.CurrentCount,
	}

	for _, a := range status.Approvals {
		response.Approvals = append(response.Approvals, ApprovalRecordResponse{
			ID:        a.ID,
			UserID:    a.UserID,
			UserName:  a.UserName,
			Decision:  string(a.Decision),
			Comment:   a.Comment,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ApprovalsHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	teamID := middleware.GetTeamID(r.Context())
	scope := r.URL.Query().Get("scope")

	var rules []domain.Rule
	var err error

	if scope != "" {
		rules, err = h.service.GetPendingRulesByScope(r.Context(), domain.TargetLayer(scope))
	} else if teamID != "" {
		rules, err = h.service.GetPendingRules(r.Context(), teamID)
	} else {
		http.Error(w, "team_id or scope required", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []PendingRuleResponse
	for _, rule := range rules {
		resp := PendingRuleResponse{
			ID:          rule.ID,
			Name:        rule.Name,
			TargetLayer: string(rule.TargetLayer),
			TeamID:      rule.TeamID,
			CreatedBy:   rule.CreatedBy,
		}
		if rule.SubmittedAt != nil {
			resp.SubmittedAt = rule.SubmittedAt.Format("2006-01-02T15:04:05Z")
		}
		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ApprovalsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		http.Error(w, "rule id required", http.StatusBadRequest)
		return
	}

	if err := h.service.ResetRule(r.Context(), ruleID); err != nil {
		if err.Error() == "rule not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/pending", h.ListPending)
	r.Post("/rules/{ruleId}/submit", h.Submit)
	r.Get("/rules/{ruleId}", h.GetStatus)
	r.Post("/rules/{ruleId}/approve", h.Approve)
	r.Post("/rules/{ruleId}/reject", h.Reject)
	r.Post("/rules/{ruleId}/reset", h.Reset)
}
