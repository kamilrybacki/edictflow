package domain

import (
	"testing"
	"time"
)

func TestDeviceCode_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"not expired", time.Now().Add(time.Hour), false},
		{"expired", time.Now().Add(-time.Hour), true},
		{"just expired", time.Now().Add(-time.Second), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := DeviceCode{ExpiresAt: tt.expiresAt}
			if got := dc.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceCode_IsAuthorized(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		userID       *string
		authorizedAt *time.Time
		want         bool
	}{
		{"not authorized", nil, nil, false},
		{"authorized", ptr("user-123"), &now, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := DeviceCode{UserID: tt.userID, AuthorizedAt: tt.authorizedAt}
			if got := dc.IsAuthorized(); got != tt.want {
				t.Errorf("IsAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr(s string) *string { return &s }
