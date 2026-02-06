package domain

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type TeamInvite struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Code      string    `json:"code"`
	MaxUses   int       `json:"max_uses"`
	UseCount  int       `json:"use_count"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes I, O, 0, 1 for readability

func GenerateInviteCode() string {
	code := make([]byte, 8)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeChars))))
		code[i] = inviteCodeChars[n.Int64()]
	}
	return string(code)
}

func NewTeamInvite(teamID, createdBy string, maxUses, expiresInHours int) TeamInvite {
	return TeamInvite{
		ID:        uuid.New().String(),
		TeamID:    teamID,
		Code:      GenerateInviteCode(),
		MaxUses:   maxUses,
		UseCount:  0,
		ExpiresAt: time.Now().Add(time.Duration(expiresInHours) * time.Hour),
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}
}

func (i *TeamInvite) IsValid() bool {
	return i.UseCount < i.MaxUses && time.Now().Before(i.ExpiresAt)
}

func (i *TeamInvite) IncrementUseCount() {
	i.UseCount++
}
