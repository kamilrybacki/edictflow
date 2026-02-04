package domain

import "time"

type DeviceCode struct {
	DeviceCode   string     `json:"device_code"`
	UserCode     string     `json:"user_code"`
	UserID       *string    `json:"user_id,omitempty"`
	ClientID     string     `json:"client_id"`
	ExpiresAt    time.Time  `json:"expires_at"`
	AuthorizedAt *time.Time `json:"authorized_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (dc *DeviceCode) IsExpired() bool {
	return time.Now().After(dc.ExpiresAt)
}

func (dc *DeviceCode) IsAuthorized() bool {
	return dc.UserID != nil && dc.AuthorizedAt != nil
}
