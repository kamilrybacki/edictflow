package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TeamSettings struct {
	DriftThresholdMinutes int `json:"drift_threshold_minutes"`
}

type Team struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Settings  TeamSettings `json:"settings"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewTeam(name string) Team {
	return Team{
		ID:        uuid.New().String(),
		Name:      name,
		Settings:  TeamSettings{DriftThresholdMinutes: 60},
		CreatedAt: time.Now(),
	}
}

func (t Team) Validate() error {
	if t.Name == "" {
		return errors.New("team name cannot be empty")
	}
	return nil
}
