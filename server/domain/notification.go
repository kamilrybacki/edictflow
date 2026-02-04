package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeChangeDetected     NotificationType = "change_detected"
	NotificationTypeApprovalRequired   NotificationType = "approval_required"
	NotificationTypeChangeApproved     NotificationType = "change_approved"
	NotificationTypeChangeRejected     NotificationType = "change_rejected"
	NotificationTypeChangeAutoReverted NotificationType = "change_auto_reverted"
	NotificationTypeExceptionGranted   NotificationType = "exception_granted"
	NotificationTypeExceptionDenied    NotificationType = "exception_denied"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeChangeDetected, NotificationTypeApprovalRequired,
		NotificationTypeChangeApproved, NotificationTypeChangeRejected,
		NotificationTypeChangeAutoReverted, NotificationTypeExceptionGranted,
		NotificationTypeExceptionDenied:
		return true
	}
	return false
}

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	TeamID    *string                `json:"team_id,omitempty"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func NewNotification(
	userID string,
	teamID *string,
	notificationType NotificationType,
	title, body string,
	metadata map[string]interface{},
) Notification {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		TeamID:    teamID,
		Type:      notificationType,
		Title:     title,
		Body:      body,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
}

func (n Notification) Validate() error {
	if n.UserID == "" {
		return errors.New("user_id cannot be empty")
	}
	if n.Title == "" {
		return errors.New("title cannot be empty")
	}
	if n.Body == "" {
		return errors.New("body cannot be empty")
	}
	if !n.Type.IsValid() {
		return errors.New("invalid notification type")
	}
	return nil
}

func (n *Notification) MarkRead() {
	now := time.Now()
	n.ReadAt = &now
}

func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}
