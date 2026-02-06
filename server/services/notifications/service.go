package notifications

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrChannelNotFound      = errors.New("notification channel not found")
)

type Service struct {
	notificationRepo NotificationRepository
	channelRepo      NotificationChannelRepository
	dispatcher       *Dispatcher
}

func NewService(
	notificationRepo NotificationRepository,
	channelRepo NotificationChannelRepository,
) *Service {
	return &Service{
		notificationRepo: notificationRepo,
		channelRepo:      channelRepo,
	}
}

func (s *Service) WithDispatcher(dispatcher *Dispatcher) *Service {
	s.dispatcher = dispatcher
	return s
}

func (s *Service) Create(ctx context.Context, n domain.Notification) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if err := s.notificationRepo.Create(ctx, n); err != nil {
		return err
	}

	// Dispatch to external channels if team is specified
	if n.TeamID != nil {
		go s.dispatchToChannels(ctx, *n.TeamID, n)
	}

	return nil
}

func (s *Service) CreateBulk(ctx context.Context, notifications []domain.Notification) error {
	for _, n := range notifications {
		if err := n.Validate(); err != nil {
			return err
		}
	}
	return s.notificationRepo.CreateBulk(ctx, notifications)
}

func (s *Service) ListForUser(ctx context.Context, userID string, filter NotificationFilter) ([]domain.Notification, error) {
	return s.notificationRepo.ListByUser(ctx, userID, filter)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, id string) error {
	return s.notificationRepo.MarkRead(ctx, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllRead(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	return s.notificationRepo.GetByID(ctx, id)
}

func (s *Service) dispatchToChannels(ctx context.Context, teamID string, n domain.Notification) {
	if s.dispatcher == nil {
		return
	}

	channels, err := s.channelRepo.ListEnabledByTeam(ctx, teamID)
	if err != nil {
		return
	}

	for _, ch := range channels {
		if ch.ShouldNotifyFor(n.Type) {
			s.dispatcher.Dispatch(ch, n)
		}
	}
}

// Channel management methods

func (s *Service) CreateChannel(ctx context.Context, nc domain.NotificationChannel) error {
	if err := nc.Validate(); err != nil {
		return err
	}
	return s.channelRepo.Create(ctx, nc)
}

func (s *Service) GetChannel(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	return s.channelRepo.GetByID(ctx, id)
}

func (s *Service) ListChannels(ctx context.Context, teamID string) ([]domain.NotificationChannel, error) {
	return s.channelRepo.ListByTeam(ctx, teamID)
}

func (s *Service) UpdateChannel(ctx context.Context, nc domain.NotificationChannel) error {
	if err := nc.Validate(); err != nil {
		return err
	}
	return s.channelRepo.Update(ctx, nc)
}

func (s *Service) DeleteChannel(ctx context.Context, id string) error {
	return s.channelRepo.Delete(ctx, id)
}

func (s *Service) TestChannel(ctx context.Context, id string) error {
	ch, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if ch == nil {
		return ErrChannelNotFound
	}

	if s.dispatcher == nil {
		return errors.New("dispatcher not configured")
	}

	testNotification := domain.NewNotification(
		"",
		&ch.TeamID,
		domain.NotificationTypeChangeDetected,
		"Test notification",
		"This is a test notification from Edictflow",
		map[string]interface{}{"test": true},
	)

	return s.dispatcher.DispatchSync(*ch, testNotification)
}
