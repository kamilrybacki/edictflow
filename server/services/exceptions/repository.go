package exceptions

import (
	"context"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type ExceptionRequestRepository interface {
	Create(ctx context.Context, er domain.ExceptionRequest) error
	GetByID(ctx context.Context, id string) (*domain.ExceptionRequest, error)
	ListByTeam(ctx context.Context, teamID string, filter ExceptionRequestFilter) ([]domain.ExceptionRequest, error)
	Update(ctx context.Context, er domain.ExceptionRequest) error
	FindActiveByUserRuleFile(ctx context.Context, userID, ruleID, filePath string) (*domain.ExceptionRequest, error)
}

type ExceptionRequestFilter struct {
	Status          *domain.ExceptionRequestStatus
	ExceptionType   *domain.ExceptionType
	ChangeRequestID *string
	UserID          *string
	Limit           int
	Offset          int
}

type ChangeRequestRepository interface {
	GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error)
	Update(ctx context.Context, cr domain.ChangeRequest) error
}

type NotificationCreator interface {
	Create(ctx context.Context, n domain.Notification) error
}

type AuditLogger interface {
	Log(ctx context.Context, action domain.AuditAction, actorID *string, resourceType, resourceID string, metadata map[string]interface{}) error
}

type WebSocketNotifier interface {
	BroadcastToAgent(agentID string, msgType string, payload interface{}) error
}
