package domain

import (
	"testing"
	"time"
)

func TestNewTeamInvite(t *testing.T) {
	teamID := "team-123"
	createdBy := "user-456"
	maxUses := 5
	expiresInHours := 24

	invite := NewTeamInvite(teamID, createdBy, maxUses, expiresInHours)

	if invite.TeamID != teamID {
		t.Errorf("expected TeamID %s, got %s", teamID, invite.TeamID)
	}
	if invite.CreatedBy != createdBy {
		t.Errorf("expected CreatedBy %s, got %s", createdBy, invite.CreatedBy)
	}
	if invite.MaxUses != maxUses {
		t.Errorf("expected MaxUses %d, got %d", maxUses, invite.MaxUses)
	}
	if invite.UseCount != 0 {
		t.Errorf("expected UseCount 0, got %d", invite.UseCount)
	}
	if len(invite.Code) != 8 {
		t.Errorf("expected Code length 8, got %d", len(invite.Code))
	}
	if invite.ID == "" {
		t.Error("expected ID to be set")
	}

	expectedExpiry := time.Now().Add(time.Duration(expiresInHours) * time.Hour)
	if invite.ExpiresAt.Before(expectedExpiry.Add(-time.Minute)) || invite.ExpiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expected ExpiresAt around %v, got %v", expectedExpiry, invite.ExpiresAt)
	}
}

func TestTeamInvite_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		invite   TeamInvite
		expected bool
	}{
		{
			name: "valid invite",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  2,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: true,
		},
		{
			name: "expired invite",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  2,
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: false,
		},
		{
			name: "max uses reached",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  5,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "max uses exceeded",
			invite: TeamInvite{
				MaxUses:   5,
				UseCount:  6,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.invite.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTeamInvite_IncrementUseCount(t *testing.T) {
	invite := TeamInvite{UseCount: 2}
	invite.IncrementUseCount()
	if invite.UseCount != 3 {
		t.Errorf("expected UseCount 3, got %d", invite.UseCount)
	}
}

func TestGenerateInviteCode(t *testing.T) {
	code := GenerateInviteCode()

	if len(code) != 8 {
		t.Errorf("expected code length 8, got %d", len(code))
	}

	// Check all characters are alphanumeric uppercase
	for _, c := range code {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			t.Errorf("unexpected character in code: %c", c)
		}
	}

	// Check uniqueness (generate 100 codes, all should be different)
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		c := GenerateInviteCode()
		if codes[c] {
			t.Errorf("duplicate code generated: %s", c)
		}
		codes[c] = true
	}
}
