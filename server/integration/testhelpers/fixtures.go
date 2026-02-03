//go:build integration

package testhelpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/claudeception/server/domain"
)

// Fixtures provides test data creation utilities
type Fixtures struct {
	pool *pgxpool.Pool
}

// NewFixtures creates a new Fixtures instance
func NewFixtures(pool *pgxpool.Pool) *Fixtures {
	return &Fixtures{pool: pool}
}

// CreateTeam creates a team in the database and returns it
func (f *Fixtures) CreateTeam(ctx context.Context, name string) (domain.Team, error) {
	team := domain.Team{
		ID:        uuid.New().String(),
		Name:      name,
		Settings:  domain.TeamSettings{DriftThresholdMinutes: 60},
		CreatedAt: time.Now(),
	}

	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return domain.Team{}, err
	}

	_, err = f.pool.Exec(ctx, `
		INSERT INTO teams (id, name, settings, created_at)
		VALUES ($1, $2, $3, $4)
	`, team.ID, team.Name, settingsJSON, team.CreatedAt)
	if err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

// CreateRule creates a rule in the database and returns it
func (f *Fixtures) CreateRule(ctx context.Context, name, teamID string) (domain.Rule, error) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "*.go"},
	}

	rule := domain.Rule{
		ID:             uuid.New().String(),
		Name:           name,
		Content:        "Test rule content",
		TargetLayer:    domain.TargetLayerProject,
		PriorityWeight: 0,
		Triggers:       triggers,
		TeamID:         teamID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	triggersJSON, err := json.Marshal(rule.Triggers)
	if err != nil {
		return domain.Rule{}, err
	}

	_, err = f.pool.Exec(ctx, `
		INSERT INTO rules (id, name, content, target_layer, priority_weight, triggers, team_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, rule.ID, rule.Name, rule.Content, rule.TargetLayer, rule.PriorityWeight, triggersJSON, rule.TeamID, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

// CreateUser creates a user in the database and returns the ID
func (f *Fixtures) CreateUser(ctx context.Context, email, teamID string) (string, error) {
	userID := uuid.New().String()
	now := time.Now()

	_, err := f.pool.Exec(ctx, `
		INSERT INTO users (id, email, team_id, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, email, teamID, now)
	if err != nil {
		return "", err
	}

	return userID, nil
}
