package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ApprovalDecision string

const (
	ApprovalDecisionApproved ApprovalDecision = "approved"
	ApprovalDecisionRejected ApprovalDecision = "rejected"
)

type RuleApproval struct {
	ID        string           `json:"id"`
	RuleID    string           `json:"rule_id"`
	UserID    string           `json:"user_id"`
	UserName  string           `json:"user_name,omitempty"`
	Decision  ApprovalDecision `json:"decision"`
	Comment   string           `json:"comment,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewRuleApproval(ruleID, userID string, decision ApprovalDecision, comment string) RuleApproval {
	return RuleApproval{
		ID:        uuid.New().String(),
		RuleID:    ruleID,
		UserID:    userID,
		Decision:  decision,
		Comment:   comment,
		CreatedAt: time.Now(),
	}
}

func (a RuleApproval) Validate() error {
	if a.RuleID == "" {
		return errors.New("rule ID cannot be empty")
	}
	if a.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if !a.Decision.IsValid() {
		return errors.New("invalid approval decision")
	}
	if a.Decision == ApprovalDecisionRejected && a.Comment == "" {
		return errors.New("rejection requires a comment")
	}
	return nil
}

func (d ApprovalDecision) IsValid() bool {
	switch d {
	case ApprovalDecisionApproved, ApprovalDecisionRejected:
		return true
	}
	return false
}
