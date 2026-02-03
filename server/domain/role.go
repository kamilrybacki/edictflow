package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// RoleEntity represents an RBAC role with hierarchy and permissions
// Named RoleEntity to avoid conflict with legacy Role type in user.go
// TODO: Rename to Role after user.go is updated in Task 2.3
type RoleEntity struct {
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

func NewRoleEntity(name, description string, hierarchyLevel int, parentRoleID, teamID *string) RoleEntity {
	return RoleEntity{
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

func (r RoleEntity) Validate() error {
	if r.Name == "" {
		return errors.New("role name cannot be empty")
	}
	if r.HierarchyLevel < 1 {
		return errors.New("hierarchy level must be at least 1")
	}
	return nil
}
