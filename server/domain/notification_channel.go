package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelTypeEmail   ChannelType = "email"
	ChannelTypeWebhook ChannelType = "webhook"
)

func (t ChannelType) IsValid() bool {
	switch t {
	case ChannelTypeEmail, ChannelTypeWebhook:
		return true
	}
	return false
}

type NotificationChannel struct {
	ID          string                 `json:"id"`
	TeamID      string                 `json:"team_id"`
	ChannelType ChannelType            `json:"channel_type"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
}

func NewNotificationChannel(
	teamID string,
	channelType ChannelType,
	config map[string]interface{},
) NotificationChannel {
	return NotificationChannel{
		ID:          uuid.New().String(),
		TeamID:      teamID,
		ChannelType: channelType,
		Config:      config,
		Enabled:     true,
		CreatedAt:   time.Now(),
	}
}

func (nc NotificationChannel) Validate() error {
	if nc.TeamID == "" {
		return errors.New("team_id cannot be empty")
	}
	if !nc.ChannelType.IsValid() {
		return errors.New("invalid channel_type")
	}
	if nc.Config == nil {
		return errors.New("config cannot be nil")
	}
	return nc.validateConfig()
}

func (nc NotificationChannel) validateConfig() error {
	switch nc.ChannelType {
	case ChannelTypeEmail:
		return nc.validateEmailConfig()
	case ChannelTypeWebhook:
		return nc.validateWebhookConfig()
	}
	return nil
}

func (nc NotificationChannel) validateEmailConfig() error {
	if _, ok := nc.Config["recipients"]; !ok {
		return errors.New("email config must include recipients")
	}
	recipients, ok := nc.Config["recipients"].([]interface{})
	if !ok {
		return errors.New("recipients must be an array")
	}
	if len(recipients) == 0 {
		return errors.New("recipients cannot be empty")
	}
	return nil
}

func (nc NotificationChannel) validateWebhookConfig() error {
	if _, ok := nc.Config["url"]; !ok {
		return errors.New("webhook config must include url")
	}
	url, ok := nc.Config["url"].(string)
	if !ok || url == "" {
		return errors.New("url must be a non-empty string")
	}
	return nil
}

func (nc *NotificationChannel) Enable() {
	nc.Enabled = true
}

func (nc *NotificationChannel) Disable() {
	nc.Enabled = false
}

func (nc *NotificationChannel) UpdateConfig(config map[string]interface{}) {
	nc.Config = config
}

func (nc NotificationChannel) GetEmailRecipients() []string {
	if nc.ChannelType != ChannelTypeEmail {
		return nil
	}
	recipients, ok := nc.Config["recipients"].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(recipients))
	for _, r := range recipients {
		if s, ok := r.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func (nc NotificationChannel) GetWebhookURL() string {
	if nc.ChannelType != ChannelTypeWebhook {
		return ""
	}
	url, _ := nc.Config["url"].(string)
	return url
}

func (nc NotificationChannel) GetWebhookSecret() string {
	if nc.ChannelType != ChannelTypeWebhook {
		return ""
	}
	secret, _ := nc.Config["secret"].(string)
	return secret
}

func (nc NotificationChannel) GetEvents() []string {
	events, ok := nc.Config["events"].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(events))
	for _, e := range events {
		if s, ok := e.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func (nc NotificationChannel) ShouldNotifyFor(notificationType NotificationType) bool {
	events := nc.GetEvents()
	if len(events) == 0 {
		return true
	}
	for _, e := range events {
		if e == string(notificationType) {
			return true
		}
	}
	return false
}
