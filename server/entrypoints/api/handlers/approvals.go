package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/response"
	"github.com/kamilrybacki/edictflow/server/services/approvals"
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
	Description string  `json:"description,omitempty"`
	Content     string  `json:"content"`
	TargetLayer string  `json:"target_layer"`
	TeamID      *string `json:"team_id,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	SubmittedAt string  `json:"submitted_at,omitempty"`
}

func (h *ApprovalsHandler) handleApprovalError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, approvals.ErrRuleNotFound):
		response.NotFound(w, "rule not found")
	case errors.Is(err, approvals.ErrCannotSubmit):
		response.Conflict(w, "rule cannot be submitted in current state")
	case errors.Is(err, approvals.ErrNotPending):
		response.Conflict(w, "rule is not pending approval")
	case errors.Is(err, approvals.ErrNoApprovalPermission):
		response.Forbidden(w, "user does not have permission to approve this rule")
	case errors.Is(err, approvals.ErrAlreadyVoted):
		response.Conflict(w, "user has already voted on this rule")
	default:
		response.InternalError(w, "internal server error")
	}
}

func (h *ApprovalsHandler) Submit(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		response.BadRequest(w, "rule id required")
		return
	}

	if err := h.service.SubmitRule(r.Context(), ruleID); err != nil {
		h.handleApprovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		response.BadRequest(w, "rule id required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req ApprovalDecisionRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}
	}

	if err := h.service.ApproveRule(r.Context(), ruleID, userID, req.Comment); err != nil {
		h.handleApprovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) Reject(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		response.BadRequest(w, "rule id required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.Unauthorized(w, "unauthorized")
		return
	}

	var req ApprovalDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Comment == "" {
		response.ValidationError(w, "comment required for rejection")
		return
	}

	if err := h.service.RejectRule(r.Context(), ruleID, userID, req.Comment); err != nil {
		h.handleApprovalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApprovalsHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		response.BadRequest(w, "rule id required")
		return
	}

	status, err := h.service.GetApprovalStatus(r.Context(), ruleID)
	if err != nil {
		if errors.Is(err, approvals.ErrRuleNotFound) {
			response.NotFound(w, "rule not found")
			return
		}
		response.InternalError(w, "internal server error")
		return
	}

	resp := ApprovalStatusResponse{
		RuleID:        status.RuleID,
		Status:        string(status.Status),
		RequiredCount: status.RequiredCount,
		CurrentCount:  status.CurrentCount,
	}

	for _, a := range status.Approvals {
		resp.Approvals = append(resp.Approvals, ApprovalRecordResponse{
			ID:        a.ID,
			UserID:    a.UserID,
			UserName:  a.UserName,
			Decision:  string(a.Decision),
			Comment:   a.Comment,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	response.WriteSuccess(w, resp)
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
		response.BadRequest(w, "team_id or scope required")
		return
	}

	if err != nil {
		response.InternalError(w, "internal server error")
		return
	}

	var resp []PendingRuleResponse
	for _, rule := range rules {
		item := PendingRuleResponse{
			ID:          rule.ID,
			Name:        rule.Name,
			Content:     rule.Content,
			TargetLayer: string(rule.TargetLayer),
			TeamID:      rule.TeamID,
			CreatedBy:   rule.CreatedBy,
		}
		if rule.Description != nil {
			item.Description = *rule.Description
		}
		if rule.SubmittedAt != nil {
			item.SubmittedAt = rule.SubmittedAt.Format("2006-01-02T15:04:05Z")
		}
		resp = append(resp, item)
	}

	response.WriteSuccess(w, resp)
}

func (h *ApprovalsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if ruleID == "" {
		response.BadRequest(w, "rule id required")
		return
	}

	if err := h.service.ResetRule(r.Context(), ruleID); err != nil {
		if errors.Is(err, approvals.ErrRuleNotFound) {
			response.NotFound(w, "rule not found")
			return
		}
		response.InternalError(w, "internal server error")
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
