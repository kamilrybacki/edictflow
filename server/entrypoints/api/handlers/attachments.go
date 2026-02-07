package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/services/attachments"
)

type AttachmentService interface {
	RequestAttachment(ctx context.Context, req attachments.AttachRequest) (domain.RuleAttachment, error)
	GetByID(ctx context.Context, id string) (domain.RuleAttachment, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error)
	ApproveAttachment(ctx context.Context, id, approvedBy string) (domain.RuleAttachment, error)
	RejectAttachment(ctx context.Context, id string) (domain.RuleAttachment, error)
	UpdateEnforcement(ctx context.Context, id string, mode domain.EnforcementMode, timeoutHours int) (domain.RuleAttachment, error)
	Delete(ctx context.Context, id string) error
}

type AttachmentsHandler struct {
	service AttachmentService
}

func NewAttachmentsHandler(service AttachmentService) *AttachmentsHandler {
	return &AttachmentsHandler{service: service}
}

type AttachmentResponse struct {
	ID                    string `json:"id"`
	RuleID                string `json:"ruleId"`
	TeamID                string `json:"teamId"`
	EnforcementMode       string `json:"enforcementMode"`
	TemporaryTimeoutHours int    `json:"temporaryTimeoutHours"`
	Status                string `json:"status"`
	RequestedBy           string `json:"requestedBy"`
	ApprovedBy            string `json:"approvedBy,omitempty"`
	CreatedAt             string `json:"createdAt"`
	ApprovedAt            string `json:"approvedAt,omitempty"`
}

func attachmentToResponse(att domain.RuleAttachment) AttachmentResponse {
	resp := AttachmentResponse{
		ID:                    att.ID,
		RuleID:                att.RuleID,
		TeamID:                att.TeamID,
		EnforcementMode:       string(att.EnforcementMode),
		TemporaryTimeoutHours: att.TemporaryTimeoutHours,
		Status:                string(att.Status),
		RequestedBy:           att.RequestedBy,
		CreatedAt:             att.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if att.ApprovedBy != nil {
		resp.ApprovedBy = *att.ApprovedBy
	}
	if att.ApprovedAt != nil {
		resp.ApprovedAt = att.ApprovedAt.Format("2006-01-02T15:04:05Z")
	}
	return resp
}

type CreateAttachmentRequest struct {
	RuleID          string `json:"rule_id"`
	EnforcementMode string `json:"enforcement_mode"`
	TimeoutHours    int    `json:"temporary_timeout_hours,omitempty"`
}

func (h *AttachmentsHandler) CreateForTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")
	userID := middleware.GetUserID(r.Context())

	var req CreateAttachmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mode := domain.EnforcementMode(req.EnforcementMode)
	if !mode.IsValid() {
		http.Error(w, "invalid enforcement_mode", http.StatusBadRequest)
		return
	}

	att, err := h.service.RequestAttachment(r.Context(), attachments.AttachRequest{
		RuleID:          req.RuleID,
		TeamID:          teamID,
		EnforcementMode: mode,
		TimeoutHours:    req.TimeoutHours,
		RequestedBy:     userID,
	})
	if err != nil {
		if errors.Is(err, attachments.ErrRuleNotApproved) {
			http.Error(w, "rule must be approved before attaching", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) ListByTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamId")

	atts, err := h.service.ListByTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []AttachmentResponse
	for _, att := range atts {
		response = append(response, attachmentToResponse(att))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AttachmentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	att, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

type UpdateAttachmentEnforcementRequest struct {
	EnforcementMode string `json:"enforcement_mode"`
	TimeoutHours    int    `json:"temporary_timeout_hours,omitempty"`
}

func (h *AttachmentsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateAttachmentEnforcementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mode := domain.EnforcementMode(req.EnforcementMode)
	if !mode.IsValid() {
		http.Error(w, "invalid enforcement_mode", http.StatusBadRequest)
		return
	}

	att, err := h.service.UpdateEnforcement(r.Context(), id, mode, req.TimeoutHours)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttachmentsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	att, err := h.service.ApproveAttachment(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	att, err := h.service.RejectAttachment(r.Context(), id)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachmentToResponse(att))
}

func (h *AttachmentsHandler) RegisterTeamRoutes(r chi.Router) {
	r.Post("/", h.CreateForTeam)
	r.Get("/", h.ListByTeam)
}

func (h *AttachmentsHandler) RegisterAttachmentRoutes(r chi.Router) {
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/reject", h.Reject)
}
