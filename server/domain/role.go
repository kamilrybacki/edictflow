package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role represents an RBAC role with hierarchy and permissions.
type Role struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	HierarchyLevel int          `json:"hierarchy_level"`
	ParentRoleID   *string      `json:"parent_role_id,omitempty"`
	TeamID         *string      `json:"team_id,omitempty"`
	IsSystem       bool         `json:"is_system"`
	Permissions    []Permission `json:"permissions,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

func NewRole(name, description string, hierarchyLevel int, parentRoleID, teamID *string) Role {
	return Role{
		ID:             uuid.New().String(),
		Name:           name,
		Description:    description,
		HierarchyLevel: hierarchyLevel,
		ParentRoleID:   parentRoleID,
		TeamID:         teamID,
		IsSystem:       false,
		CreatedAt:      time.Now(),
	}
}

func (r Role) Validate() error {
	if r.Name == "" {
		return errors.New("role name cannot be empty")
	}
	if r.HierarchyLevel < 1 {
		return errors.New("hierarchy level must be at least 1")
	}
	return nil
}
