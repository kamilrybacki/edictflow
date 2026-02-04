package notifications

import (
	"context"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n domain.Notification) error
	CreateBulk(ctx context.Context, notifications []domain.Notification) error
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID string, filter NotificationFilter) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type NotificationFilter struct {
	Type   *domain.NotificationType
	Unread *bool
	TeamID *string
	Limit  int
	Offset int
}

type NotificationChannelRepository interface {
	Create(ctx context.Context, nc domain.NotificationChannel) error
	GetByID(ctx context.Context, id string) (*domain.NotificationChannel, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.NotificationChannel, error)
	ListEnabledByTeam(ctx context.Context, teamID string) ([]domain.NotificationChannel, error)
	Update(ctx context.Context, nc domain.NotificationChannel) error
	Delete(ctx context.Context, id string) error
}
