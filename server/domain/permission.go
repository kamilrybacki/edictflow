package domain

import (
	"errors"
	"time"
)

type PermissionCategory string

const (
	PermissionCategoryRules PermissionCategory = "rules"
	PermissionCategoryUsers PermissionCategory = "users"
	PermissionCategoryTeams PermissionCategory = "teams"
	PermissionCategoryAdmin PermissionCategory = "admin"
)

type Permission struct {
	ID          string             `json:"id"`
	Code        string             `json:"code"`
	Description string             `json:"description"`
	Category    PermissionCategory `json:"category"`
	CreatedAt   time.Time          `json:"created_at"`
}

func (p Permission) Validate() error {
	if p.Code == "" {
		return errors.New("permission code cannot be empty")
	}
	if !p.Category.IsValid() {
		return errors.New("invalid permission category")
	}
	return nil
}

func (c PermissionCategory) IsValid() bool {
	switch c {
	case PermissionCategoryRules, PermissionCategoryUsers, PermissionCategoryTeams, PermissionCategoryAdmin:
		return true
	}
	return false
}
