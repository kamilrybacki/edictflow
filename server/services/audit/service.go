package audit

import (
	"context"
	"time"

	"github.com/kamilrybacki/edictflow/server/adapters/postgres"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type AuditDB interface {
	Create(ctx context.Context, entry domain.AuditEntry) error
	List(ctx context.Context, params postgres.AuditListParams) ([]domain.AuditEntry, int, error)
	GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error)
}

type Service struct {
	db AuditDB
}

func NewService(db AuditDB) *Service {
	return &Service{db: db}
}

type ListParams struct {
	EntityType *domain.AuditEntityType
	EntityID   *string
	ActorID    *string
	Action     *domain.AuditAction
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

func (s *Service) LogCreate(ctx context.Context, entityType domain.AuditEntityType, entityID string, actorID *string, metadata map[string]interface{}) error {
	entry := domain.NewAuditEntry(entityType, entityID, domain.AuditActionCreated, actorID)
	if metadata != nil {
		entry.Metadata = metadata
	}
	return s.db.Create(ctx, entry)
}

func (s *Service) LogUpdate(ctx context.Context, entityType domain.AuditEntityType, entityID string, actorID *string, changes map[string]*domain.ChangeValue, metadata map[string]interface{}) error {
	entry := domain.NewAuditEntry(entityType, entityID, domain.AuditActionUpdated, actorID)
	if changes != nil {
		entry.Changes = changes
	}
	if metadata != nil {
		entry.Metadata = metadata
	}
	return s.db.Create(ctx, entry)
}

func (s *Service) LogDelete(ctx context.Context, entityType domain.AuditEntityType, entityID string, actorID *string, metadata map[string]interface{}) error {
	entry := domain.NewAuditEntry(entityType, entityID, domain.AuditActionDeleted, actorID)
	if metadata != nil {
		entry.Metadata = metadata
	}
	return s.db.Create(ctx, entry)
}

func (s *Service) LogApprovalAction(ctx context.Context, ruleID string, action domain.AuditAction, actorID *string, metadata map[string]interface{}) error {
	entry := domain.NewAuditEntry(domain.AuditEntityRule, ruleID, action, actorID)
	if metadata != nil {
		entry.Metadata = metadata
	}
	return s.db.Create(ctx, entry)
}

func (s *Service) List(ctx context.Context, params ListParams) ([]domain.AuditEntry, int, error) {
	dbParams := postgres.AuditListParams{
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		ActorID:    params.ActorID,
		Action:     params.Action,
		From:       params.From,
		To:         params.To,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}
	return s.db.List(ctx, dbParams)
}

func (s *Service) GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error) {
	return s.db.GetEntityHistory(ctx, entityType, entityID)
}
