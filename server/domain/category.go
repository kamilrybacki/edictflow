package domain

import (
	"errors"
	"time"
)

type Category struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	IsSystem     bool      `json:"is_system"`
	OrgID        *string   `json:"org_id,omitempty"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("category name is required")
	}
	if len(c.Name) > 100 {
		return errors.New("category name must be 100 characters or less")
	}
	return nil
}
