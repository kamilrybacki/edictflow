package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuditEntityType string

const (
	AuditEntityRule           AuditEntityType = "rule"
	AuditEntityUser           AuditEntityType = "user"
	AuditEntityRole           AuditEntityType = "role"
	AuditEntityTeam           AuditEntityType = "team"
	AuditEntityApprovalConfig AuditEntityType = "approval_config"
)

type AuditAction string

const (
	AuditActionCreated           AuditAction = "created"
	AuditActionUpdated           AuditAction = "updated"
	AuditActionDeleted           AuditAction = "deleted"
	AuditActionSubmitted         AuditAction = "submitted"
	AuditActionApproved          AuditAction = "approved"
	AuditActionRejected          AuditAction = "rejected"
	AuditActionDeactivated       AuditAction = "deactivated"
	AuditActionRoleAssigned      AuditAction = "role_assigned"
	AuditActionRoleRemoved       AuditAction = "role_removed"
	AuditActionPermissionAdded   AuditAction = "permission_added"
	AuditActionPermissionRemoved AuditAction = "permission_removed"
)

type ChangeValue struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

type AuditEntry struct {
	ID         string                  `json:"id"`
	EntityType AuditEntityType         `json:"entity_type"`
	EntityID   string                  `json:"entity_id"`
	Action     AuditAction             `json:"action"`
	ActorID    *string                 `json:"actor_id,omitempty"`
	ActorName  string                  `json:"actor_name,omitempty"`
	Changes    map[string]*ChangeValue `json:"changes,omitempty"`
	Metadata   map[string]interface{}  `json:"metadata,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
}

func NewAuditEntry(entityType AuditEntityType, entityID string, action AuditAction, actorID *string) AuditEntry {
	return AuditEntry{
		ID:         uuid.New().String(),
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		Changes:    make(map[string]*ChangeValue),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}
}

func (e *AuditEntry) AddChange(field string, oldVal, newVal interface{}) {
	e.Changes[field] = &ChangeValue{Old: oldVal, New: newVal}
}

func (e *AuditEntry) AddMetadata(key string, value interface{}) {
	e.Metadata[key] = value
}

func (e AuditEntry) Validate() error {
	if e.EntityType == "" {
		return errors.New("entity type cannot be empty")
	}
	if e.EntityID == "" {
		return errors.New("entity ID cannot be empty")
	}
	if e.Action == "" {
		return errors.New("action cannot be empty")
	}
	return nil
}
