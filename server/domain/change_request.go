package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChangeRequestStatus string

const (
	ChangeRequestStatusPending          ChangeRequestStatus = "pending"
	ChangeRequestStatusApproved         ChangeRequestStatus = "approved"
	ChangeRequestStatusRejected         ChangeRequestStatus = "rejected"
	ChangeRequestStatusAutoReverted     ChangeRequestStatus = "auto_reverted"
	ChangeRequestStatusExceptionGranted ChangeRequestStatus = "exception_granted"
)

func (s ChangeRequestStatus) IsValid() bool {
	switch s {
	case ChangeRequestStatusPending, ChangeRequestStatusApproved,
		ChangeRequestStatusRejected, ChangeRequestStatusAutoReverted,
		ChangeRequestStatusExceptionGranted:
		return true
	}
	return false
}

type ChangeRequest struct {
	ID               string              `json:"id"`
	RuleID           string              `json:"rule_id"`
	AgentID          string              `json:"agent_id"`
	UserID           string              `json:"user_id"`
	TeamID           string              `json:"team_id"`
	FilePath         string              `json:"file_path"`
	OriginalHash     string              `json:"original_hash"`
	ModifiedHash     string              `json:"modified_hash"`
	DiffContent      string              `json:"diff_content"`
	Status           ChangeRequestStatus `json:"status"`
	EnforcementMode  EnforcementMode     `json:"enforcement_mode"`
	TimeoutAt        *time.Time          `json:"timeout_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	ResolvedAt       *time.Time          `json:"resolved_at,omitempty"`
	ResolvedByUserID *string             `json:"resolved_by_user_id,omitempty"`
}

func NewChangeRequest(
	ruleID, agentID, userID, teamID string,
	filePath, originalHash, modifiedHash, diffContent string,
	enforcementMode EnforcementMode,
	timeoutAt *time.Time,
) ChangeRequest {
	return ChangeRequest{
		ID:              uuid.New().String(),
		RuleID:          ruleID,
		AgentID:         agentID,
		UserID:          userID,
		TeamID:          teamID,
		FilePath:        filePath,
		OriginalHash:    originalHash,
		ModifiedHash:    modifiedHash,
		DiffContent:     diffContent,
		Status:          ChangeRequestStatusPending,
		EnforcementMode: enforcementMode,
		TimeoutAt:       timeoutAt,
		CreatedAt:       time.Now(),
	}
}

func (cr ChangeRequest) Validate() error {
	if cr.RuleID == "" {
		return errors.New("rule_id cannot be empty")
	}
	if cr.AgentID == "" {
		return errors.New("agent_id cannot be empty")
	}
	if cr.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if cr.TeamID == "" {
		return errors.New("team_id cannot be empty")
	}
	if cr.FilePath == "" {
		return errors.New("file_path cannot be empty")
	}
	if cr.OriginalHash == "" {
		return errors.New("original_hash cannot be empty")
	}
	if cr.ModifiedHash == "" {
		return errors.New("modified_hash cannot be empty")
	}
	if !cr.Status.IsValid() {
		return errors.New("invalid status")
	}
	if !cr.EnforcementMode.IsValid() {
		return errors.New("invalid enforcement_mode")
	}
	return nil
}

func (cr *ChangeRequest) Approve(approverUserID string) {
	cr.Status = ChangeRequestStatusApproved
	now := time.Now()
	cr.ResolvedAt = &now
	cr.ResolvedByUserID = &approverUserID
}

func (cr *ChangeRequest) Reject(approverUserID string) {
	cr.Status = ChangeRequestStatusRejected
	now := time.Now()
	cr.ResolvedAt = &now
	cr.ResolvedByUserID = &approverUserID
}

func (cr *ChangeRequest) AutoRevert() {
	cr.Status = ChangeRequestStatusAutoReverted
	now := time.Now()
	cr.ResolvedAt = &now
}

func (cr *ChangeRequest) GrantException() {
	cr.Status = ChangeRequestStatusExceptionGranted
	now := time.Now()
	cr.ResolvedAt = &now
}

func (cr *ChangeRequest) IsPending() bool {
	return cr.Status == ChangeRequestStatusPending
}

func (cr *ChangeRequest) IsExpired() bool {
	if cr.TimeoutAt == nil {
		return false
	}
	return time.Now().After(*cr.TimeoutAt)
}

func (cr *ChangeRequest) UpdateDiff(newHash, newDiff string) {
	cr.ModifiedHash = newHash
	cr.DiffContent = newDiff
}
