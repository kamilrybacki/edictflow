package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ApprovalConfig struct {
	ID                 string      `json:"id"`
	Scope              TargetLayer `json:"scope"`
	RequiredPermission string      `json:"required_permission"`
	RequiredCount      int         `json:"required_count"`
	TeamID             *string     `json:"team_id,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
}

func NewApprovalConfig(scope TargetLayer, requiredPermission string, requiredCount int, teamID *string) ApprovalConfig {
	return ApprovalConfig{
		ID:                 uuid.New().String(),
		Scope:              scope,
		RequiredPermission: requiredPermission,
		RequiredCount:      requiredCount,
		TeamID:             teamID,
		CreatedAt:          time.Now(),
	}
}

func (c ApprovalConfig) Validate() error {
	if !c.Scope.IsValid() {
		return errors.New("invalid scope")
	}
	if c.RequiredPermission == "" {
		return errors.New("required permission cannot be empty")
	}
	if c.RequiredCount < 1 {
		return errors.New("required count must be at least 1")
	}
	return nil
}

func (c ApprovalConfig) CanOverrideWith(newCount int) bool {
	return newCount >= c.RequiredCount
}

func (c ApprovalConfig) IsGlobal() bool {
	return c.TeamID == nil
}
