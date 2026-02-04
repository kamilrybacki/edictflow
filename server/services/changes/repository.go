package changes

import (
	"context"
	"time"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type ChangeRequestRepository interface {
	Create(ctx context.Context, cr domain.ChangeRequest) error
	GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error)
	ListByTeam(ctx context.Context, teamID string, filter ChangeRequestFilter) ([]domain.ChangeRequest, error)
	Update(ctx context.Context, cr domain.ChangeRequest) error
	FindExpiredTemporary(ctx context.Context, before time.Time) ([]domain.ChangeRequest, error)
	FindByAgentAndFile(ctx context.Context, agentID, filePath string) (*domain.ChangeRequest, error)
}

type ChangeRequestFilter struct {
	Status          *domain.ChangeRequestStatus
	EnforcementMode *domain.EnforcementMode
	RuleID          *string
	AgentID         *string
	UserID          *string
	Limit           int
	Offset          int
}

type RuleRepository interface {
	GetRule(ctx context.Context, id string) (domain.Rule, error)
}

type AgentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Agent, error)
}

type NotificationCreator interface {
	Create(ctx context.Context, n domain.Notification) error
}

type AuditLogger interface {
	Log(ctx context.Context, action domain.AuditAction, actorID *string, resourceType, resourceID string, metadata map[string]interface{}) error
}
