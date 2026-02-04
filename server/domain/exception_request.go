package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ExceptionRequestStatus string

const (
	ExceptionRequestStatusPending  ExceptionRequestStatus = "pending"
	ExceptionRequestStatusApproved ExceptionRequestStatus = "approved"
	ExceptionRequestStatusDenied   ExceptionRequestStatus = "denied"
)

func (s ExceptionRequestStatus) IsValid() bool {
	switch s {
	case ExceptionRequestStatusPending, ExceptionRequestStatusApproved, ExceptionRequestStatusDenied:
		return true
	}
	return false
}

type ExceptionType string

const (
	ExceptionTypeTimeLimited ExceptionType = "time_limited"
	ExceptionTypePermanent   ExceptionType = "permanent"
)

func (t ExceptionType) IsValid() bool {
	switch t {
	case ExceptionTypeTimeLimited, ExceptionTypePermanent:
		return true
	}
	return false
}

type ExceptionRequest struct {
	ID               string                 `json:"id"`
	ChangeRequestID  string                 `json:"change_request_id"`
	UserID           string                 `json:"user_id"`
	Justification    string                 `json:"justification"`
	ExceptionType    ExceptionType          `json:"exception_type"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	Status           ExceptionRequestStatus `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
	ResolvedByUserID *string                `json:"resolved_by_user_id,omitempty"`
}

func NewExceptionRequest(
	changeRequestID, userID, justification string,
	exceptionType ExceptionType,
) ExceptionRequest {
	return ExceptionRequest{
		ID:              uuid.New().String(),
		ChangeRequestID: changeRequestID,
		UserID:          userID,
		Justification:   justification,
		ExceptionType:   exceptionType,
		Status:          ExceptionRequestStatusPending,
		CreatedAt:       time.Now(),
	}
}

func (er ExceptionRequest) Validate() error {
	if er.ChangeRequestID == "" {
		return errors.New("change_request_id cannot be empty")
	}
	if er.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if er.Justification == "" {
		return errors.New("justification cannot be empty")
	}
	if !er.ExceptionType.IsValid() {
		return errors.New("invalid exception_type")
	}
	if !er.Status.IsValid() {
		return errors.New("invalid status")
	}
	return nil
}

func (er *ExceptionRequest) Approve(approverUserID string, expiresAt *time.Time) {
	er.Status = ExceptionRequestStatusApproved
	now := time.Now()
	er.ResolvedAt = &now
	er.ResolvedByUserID = &approverUserID
	er.ExpiresAt = expiresAt
}

func (er *ExceptionRequest) Deny(approverUserID string) {
	er.Status = ExceptionRequestStatusDenied
	now := time.Now()
	er.ResolvedAt = &now
	er.ResolvedByUserID = &approverUserID
}

func (er *ExceptionRequest) IsPending() bool {
	return er.Status == ExceptionRequestStatusPending
}

func (er *ExceptionRequest) IsActive() bool {
	if er.Status != ExceptionRequestStatusApproved {
		return false
	}
	if er.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(*er.ExpiresAt)
}
