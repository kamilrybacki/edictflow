package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/rules"
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
		INSERT INTO rules (
			id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force,
			status, enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20,
			$21, $22, $23, $24
		)
	`, rule.ID, rule.Name, rule.Content, rule.Description, rule.TargetLayer, rule.CategoryID,
		rule.PriorityWeight, rule.Overridable, rule.EffectiveStart, rule.EffectiveEnd,
		rule.TargetTeams, rule.TargetUsers, rule.Tags, triggersJSON, rule.TeamID, rule.Force,
		rule.Status, rule.EnforcementMode, rule.TemporaryTimeoutHours, rule.CreatedBy,
		rule.SubmittedAt, rule.ApprovedAt, rule.CreatedAt, rule.UpdatedAt)
	return err
}

// GetRule retrieves a rule by ID
func (db *RuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	var rule domain.Rule
	var triggersJSON []byte

	err := db.pool.QueryRow(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE id = $1
	`, id).Scan(
		&rule.ID, &rule.Name, &rule.Content, &rule.Description, &rule.TargetLayer, &rule.CategoryID,
		&rule.PriorityWeight, &rule.Overridable, &rule.EffectiveStart, &rule.EffectiveEnd,
		&rule.TargetTeams, &rule.TargetUsers, &rule.Tags, &triggersJSON, &rule.TeamID, &rule.Force, &rule.Status,
		&rule.EnforcementMode, &rule.TemporaryTimeoutHours, &rule.CreatedBy,
		&rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt,
	)

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
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE team_id = $1
		ORDER BY priority_weight DESC, created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// scanRules is a helper to scan multiple rules from rows
func (db *RuleDB) scanRules(rows pgx.Rows) ([]domain.Rule, error) {
	var rulesList []domain.Rule
	for rows.Next() {
		var rule domain.Rule
		var triggersJSON []byte

		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Content, &rule.Description, &rule.TargetLayer, &rule.CategoryID,
			&rule.PriorityWeight, &rule.Overridable, &rule.EffectiveStart, &rule.EffectiveEnd,
			&rule.TargetTeams, &rule.TargetUsers, &rule.Tags, &triggersJSON, &rule.TeamID, &rule.Force, &rule.Status,
			&rule.EnforcementMode, &rule.TemporaryTimeoutHours, &rule.CreatedBy,
			&rule.SubmittedAt, &rule.ApprovedAt, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
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
		SET name = $2, content = $3, description = $4, target_layer = $5, category_id = $6,
			priority_weight = $7, overridable = $8, effective_start = $9, effective_end = $10,
			target_teams = $11, target_users = $12, tags = $13, triggers = $14,
			enforcement_mode = $15, temporary_timeout_hours = $16, updated_at = $17
		WHERE id = $1
	`, rule.ID, rule.Name, rule.Content, rule.Description, rule.TargetLayer, rule.CategoryID,
		rule.PriorityWeight, rule.Overridable, rule.EffectiveStart, rule.EffectiveEnd,
		rule.TargetTeams, rule.TargetUsers, rule.Tags, triggersJSON,
		rule.EnforcementMode, rule.TemporaryTimeoutHours, rule.UpdatedAt)
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
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules WHERE team_id = $1 AND status = $2
		ORDER BY created_at DESC
	`, teamID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// ListPendingByScope retrieves all pending rules by target layer scope
func (db *RuleDB) ListPendingByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules WHERE target_layer = $1 AND status = 'pending'
		ORDER BY submitted_at ASC
	`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// GetRulesForMerge returns all approved rules for a given target layer, filtered by targeting
func (db *RuleDB) GetRulesForMerge(ctx context.Context, targetLayer domain.TargetLayer, userID string, teamIDs []string, teamInheritsGlobal bool) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE status = 'approved'
		  AND (
			  -- Global rules (team_id IS NULL)
			  (team_id IS NULL AND target_layer = $1 AND (force = true OR $4 = true))
			  OR
			  -- Team rules (existing logic)
			  (team_id IS NOT NULL AND target_layer = $1 AND (
				  (target_teams = '{}' AND target_users = '{}')
				  OR $2 = ANY(target_users)
				  OR target_teams && $3::uuid[]
			  ))
		  )
		ORDER BY
			CASE WHEN team_id IS NULL THEN 0 ELSE 1 END,
			force DESC,
			priority_weight DESC,
			name
	`, targetLayer, userID, teamIDs, teamInheritsGlobal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// ListByTargetLayer retrieves all rules for a specific target layer
func (db *RuleDB) ListByTargetLayer(ctx context.Context, targetLayer domain.TargetLayer) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE target_layer = $1 AND status = 'approved'
		ORDER BY priority_weight DESC, name
	`, targetLayer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// ListGlobalRules retrieves all global rules (team_id IS NULL)
func (db *RuleDB) ListGlobalRules(ctx context.Context) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		WHERE team_id IS NULL
		ORDER BY force DESC, priority_weight DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}

// ListAllRules retrieves all rules across all teams
func (db *RuleDB) ListAllRules(ctx context.Context) ([]domain.Rule, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, content, description, target_layer, category_id,
			priority_weight, overridable, effective_start, effective_end,
			target_teams, target_users, tags, triggers, team_id, force, status,
			enforcement_mode, temporary_timeout_hours, created_by,
			submitted_at, approved_at, created_at, updated_at
		FROM rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return db.scanRules(rows)
}
