package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AttachmentStatus string

const (
	AttachmentStatusPending  AttachmentStatus = "pending"
	AttachmentStatusApproved AttachmentStatus = "approved"
	AttachmentStatusRejected AttachmentStatus = "rejected"
)

func (s AttachmentStatus) IsValid() bool {
	switch s {
	case AttachmentStatusPending, AttachmentStatusApproved, AttachmentStatusRejected:
		return true
	}
	return false
}

type RuleAttachment struct {
	ID                    string           `json:"id"`
	RuleID                string           `json:"rule_id"`
	TeamID                string           `json:"team_id"`
	EnforcementMode       EnforcementMode  `json:"enforcement_mode"`
	TemporaryTimeoutHours int              `json:"temporary_timeout_hours"`
	Status                AttachmentStatus `json:"status"`
	RequestedBy           string           `json:"requested_by"`
	ApprovedBy            *string          `json:"approved_by,omitempty"`
	CreatedAt             time.Time        `json:"created_at"`
	ApprovedAt            *time.Time       `json:"approved_at,omitempty"`
}

func NewRuleAttachment(ruleID, teamID string, enforcementMode EnforcementMode, requestedBy string) RuleAttachment {
	return RuleAttachment{
		ID:                    uuid.New().String(),
		RuleID:                ruleID,
		TeamID:                teamID,
		EnforcementMode:       enforcementMode,
		TemporaryTimeoutHours: 24,
		Status:                AttachmentStatusPending,
		RequestedBy:           requestedBy,
		CreatedAt:             time.Now(),
	}
}

func NewApprovedAttachment(ruleID, teamID string, enforcementMode EnforcementMode, approvedBy string) RuleAttachment {
	now := time.Now()
	return RuleAttachment{
		ID:                    uuid.New().String(),
		RuleID:                ruleID,
		TeamID:                teamID,
		EnforcementMode:       enforcementMode,
		TemporaryTimeoutHours: 24,
		Status:                AttachmentStatusApproved,
		RequestedBy:           approvedBy,
		ApprovedBy:            &approvedBy,
		CreatedAt:             now,
		ApprovedAt:            &now,
	}
}

func (a RuleAttachment) Validate() error {
	if a.RuleID == "" {
		return errors.New("rule ID cannot be empty")
	}
	if a.TeamID == "" {
		return errors.New("team ID cannot be empty")
	}
	if !a.EnforcementMode.IsValid() {
		return errors.New("invalid enforcement mode")
	}
	return nil
}

func (a *RuleAttachment) Approve(approvedBy string) {
	a.Status = AttachmentStatusApproved
	a.ApprovedBy = &approvedBy
	now := time.Now()
	a.ApprovedAt = &now
}

func (a *RuleAttachment) Reject() {
	a.Status = AttachmentStatusRejected
}

func (a *RuleAttachment) UpdateEnforcement(mode EnforcementMode, timeoutHours int) {
	a.EnforcementMode = mode
	if mode == EnforcementModeTemporary {
		a.TemporaryTimeoutHours = timeoutHours
	}
}
