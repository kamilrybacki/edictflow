package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
)

type ExceptionService interface {
	GetByID(ctx context.Context, id string) (*domain.ExceptionRequest, error)
	ListByTeam(ctx context.Context, teamID string, filter ExceptionRequestFilter) ([]domain.ExceptionRequest, error)
	Create(ctx context.Context, req CreateExceptionServiceRequest) (*domain.ExceptionRequest, error)
	Approve(ctx context.Context, id, approverUserID string, expiresAt *time.Time) error
	Deny(ctx context.Context, id, approverUserID string) error
}

type ExceptionRequestFilter struct {
	Status          *domain.ExceptionRequestStatus
	ExceptionType   *domain.ExceptionType
	ChangeRequestID *string
	UserID          *string
	Limit           int
	Offset          int
}

type CreateExceptionServiceRequest struct {
	ChangeRequestID        string
	UserID                 string
	Justification          string
	ExceptionType          domain.ExceptionType
	RequestedDurationHours *int
}

type ExceptionsHandler struct {
	service ExceptionService
}

func NewExceptionsHandler(service ExceptionService) *ExceptionsHandler {
	return &ExceptionsHandler{service: service}
}

type CreateExceptionAPIRequest struct {
	ChangeRequestID        string `json:"change_request_id"`
	Justification          string `json:"justification"`
	ExceptionType          string `json:"exception_type"`
	RequestedDurationHours *int   `json:"requested_duration_hours,omitempty"`
}

type ApproveExceptionRequest struct {
	ExpiresAt *string `json:"expires_at,omitempty"`
}

type ExceptionRequestResponse struct {
	ID               string  `json:"id"`
	ChangeRequestID  string  `json:"change_request_id"`
	UserID           string  `json:"user_id"`
	Justification    string  `json:"justification"`
	ExceptionType    string  `json:"exception_type"`
	ExpiresAt        *string `json:"expires_at,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	ResolvedAt       *string `json:"resolved_at,omitempty"`
	ResolvedByUserID *string `json:"resolved_by_user_id,omitempty"`
}

func exceptionRequestToResponse(er domain.ExceptionRequest) ExceptionRequestResponse {
	resp := ExceptionRequestResponse{
		ID:               er.ID,
		ChangeRequestID:  er.ChangeRequestID,
		UserID:           er.UserID,
		Justification:    er.Justification,
		ExceptionType:    string(er.ExceptionType),
		Status:           string(er.Status),
		CreatedAt:        er.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ResolvedByUserID: er.ResolvedByUserID,
	}

	if er.ExpiresAt != nil {
		t := er.ExpiresAt.Format("2006-01-02T15:04:05Z")
		resp.ExpiresAt = &t
	}
	if er.ResolvedAt != nil {
		t := er.ResolvedAt.Format("2006-01-02T15:04:05Z")
		resp.ResolvedAt = &t
	}

	return resp
}

func (h *ExceptionsHandler) List(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id query parameter required", http.StatusBadRequest)
		return
	}

	filter := ExceptionRequestFilter{}
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ExceptionRequestStatus(status)
		filter.Status = &s
	}
	if changeRequestID := r.URL.Query().Get("change_request_id"); changeRequestID != "" {
		filter.ChangeRequestID = &changeRequestID
	}

	exceptions, err := h.service.ListByTeam(r.Context(), teamID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []ExceptionRequestResponse
	for _, er := range exceptions {
		response = append(response, exceptionRequestToResponse(er))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ExceptionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateExceptionAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	er, err := h.service.Create(r.Context(), CreateExceptionServiceRequest{
		ChangeRequestID:        req.ChangeRequestID,
		UserID:                 userID,
		Justification:          req.Justification,
		ExceptionType:          domain.ExceptionType(req.ExceptionType),
		RequestedDurationHours: req.RequestedDurationHours,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exceptionRequestToResponse(*er))
}

func (h *ExceptionsHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req ApproveExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK
		req = ApproveExceptionRequest{}
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			http.Error(w, "invalid expires_at format", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	if err := h.service.Approve(r.Context(), id, userID, expiresAt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExceptionsHandler) Deny(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Deny(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExceptionsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Post("/{id}/approve", h.Approve)
	r.Post("/{id}/deny", h.Deny)
}
