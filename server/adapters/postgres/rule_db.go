package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/services/rules"
)

// RuleDB implements rules.DB interface with PostgreSQL
type RuleDB struct {
	pool *pgxpool.Pool
}

// NewRuleDB creates a new RuleDB instance
func NewRuleDB(pool *pgxpool.Pool) *RuleDB {
	return &RuleDB{pool: pool}
}

// CreateRule inserts a new rule into the database
func (db *RuleDB) CreateRule(ctx context.Context, rule domain.Rule) error {
	triggersJSON, err := json.Marshal(rule.Triggers)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx, `
		INSERT INTO rules (id, name, content, target_layer, priority_weight, triggers, team_id, status, created_by, submitted_at, approved_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, rule.ID, rule.Name, rule.Content, rule.TargetLayer, rule.PriorityWeight, triggersJSON, rule.TeamID, rule.Status, rule.CreatedBy, rule.SubmittedAt, rule.ApprovedAt, rule.CreatedAt, rule.UpdatedAt)
	return err
}

// GetRule retrieves a rule by ID
func (db *RuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	var rule domain.Rule
	var triggersJSON []byte

	err := db.pool.QueryRow(ctx, `
		SELECT id, name, content, target_layer, priority_weight, triggers, team_id, status, created_by, submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE id = $1
	`, id).Scan(&rule.ID, &rule.Name, &rule.Content, &rule.TargetLayer, &rule.PriorityWeight, &triggersJSON, &rule.TeamID, &rule.Status, &rule.CreatedBy, &rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Rule{}, rules.ErrRuleNotFound
		}
		return domain.Rule{}, err
	}

	if err := json.Unmarshal(triggersJSON, &rule.Triggers); err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

// ListRulesByTeam retrieves all rules for a team
func (db *RuleDB) ListRulesByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, target_layer, priority_weight, triggers, team_id, status, created_by, submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE team_id = $1
		ORDER BY priority_weight DESC, created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rulesList []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var triggersJSON []byte

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Content, &rule.TargetLayer, &rule.PriorityWeight, &triggersJSON, &rule.TeamID, &rule.Status, &rule.CreatedBy, &rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(triggersJSON, &rule.Triggers); err != nil {
			return nil, err
		}

		rulesList = append(rulesList, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rulesList, nil
}

// UpdateRule updates an existing rule
func (db *RuleDB) UpdateRule(ctx context.Context, rule domain.Rule) error {
	triggersJSON, err := json.Marshal(rule.Triggers)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx, `
		UPDATE rules
		SET name = $2, content = $3, target_layer = $4, priority_weight = $5, triggers = $6, updated_at = $7
		WHERE id = $1
	`, rule.ID, rule.Name, rule.Content, rule.TargetLayer, rule.PriorityWeight, triggersJSON, rule.UpdatedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return rules.ErrRuleNotFound
	}

	return nil
}

// DeleteRule removes a rule by ID
func (db *RuleDB) DeleteRule(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `
		DELETE FROM rules
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return rules.ErrRuleNotFound
	}

	return nil
}

// UpdateStatus updates the status and related timestamps of a rule
func (db *RuleDB) UpdateStatus(ctx context.Context, rule domain.Rule) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE rules SET status = $2, submitted_at = $3, approved_at = $4, updated_at = $5
		WHERE id = $1
	`, rule.ID, rule.Status, rule.SubmittedAt, rule.ApprovedAt, rule.UpdatedAt)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return rules.ErrRuleNotFound
	}
	return nil
}

// ListByStatus retrieves all rules for a team with a specific status
func (db *RuleDB) ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, target_layer, priority_weight, triggers, team_id, status, created_by, submitted_at, approved_at, created_at, updated_at
		FROM rules WHERE team_id = $1 AND status = $2
		ORDER BY created_at DESC
	`, teamID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rulesList []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var triggersJSON []byte

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Content, &rule.TargetLayer, &rule.PriorityWeight, &triggersJSON, &rule.TeamID, &rule.Status, &rule.CreatedBy, &rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(triggersJSON, &rule.Triggers); err != nil {
			return nil, err
		}

		rulesList = append(rulesList, rule)
	}

	return rulesList, rows.Err()
}

// ListPendingByScope retrieves all pending rules by target layer scope
func (db *RuleDB) ListPendingByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, target_layer, priority_weight, triggers, team_id, status, created_by, submitted_at, approved_at, created_at, updated_at
		FROM rules WHERE target_layer = $1 AND status = 'pending'
		ORDER BY submitted_at ASC
	`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rulesList []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var triggersJSON []byte

		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Content, &rule.TargetLayer, &rule.PriorityWeight, &triggersJSON, &rule.TeamID, &rule.Status, &rule.CreatedBy, &rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(triggersJSON, &rule.Triggers); err != nil {
			return nil, err
		}

		rulesList = append(rulesList, rule)
	}

	return rulesList, rows.Err()
}
