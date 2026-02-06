package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrApprovalConfigNotFound = errors.New("approval config not found")

type ApprovalConfigDB struct {
	pool *pgxpool.Pool
}

func NewApprovalConfigDB(pool *pgxpool.Pool) *ApprovalConfigDB {
	return &ApprovalConfigDB{pool: pool}
}

func (db *ApprovalConfigDB) GetForScope(ctx context.Context, scope domain.TargetLayer, teamID *string) (domain.ApprovalConfig, error) {
	// First try team-specific config
	if teamID != nil {
		var config domain.ApprovalConfig
		err := db.pool.QueryRow(ctx, `
			SELECT id, scope, required_permission, required_count, team_id, created_at
			FROM approval_configs WHERE scope = $1 AND team_id = $2
		`, scope, *teamID).Scan(&config.ID, &config.Scope, &config.RequiredPermission, &config.RequiredCount, &config.TeamID, &config.CreatedAt)

		if err == nil {
			return config, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.ApprovalConfig{}, err
		}
	}

	// Fall back to global default
	var config domain.ApprovalConfig
	err := db.pool.QueryRow(ctx, `
		SELECT id, scope, required_permission, required_count, team_id, created_at
		FROM approval_configs WHERE scope = $1 AND team_id IS NULL
	`, scope).Scan(&config.ID, &config.Scope, &config.RequiredPermission, &config.RequiredCount, &config.TeamID, &config.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ApprovalConfig{}, ErrApprovalConfigNotFound
	}
	return config, err
}

func (db *ApprovalConfigDB) List(ctx context.Context) ([]domain.ApprovalConfig, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, scope, required_permission, required_count, team_id, created_at
		FROM approval_configs ORDER BY scope, team_id NULLS FIRST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.ApprovalConfig
	for rows.Next() {
		var config domain.ApprovalConfig
		if err := rows.Scan(&config.ID, &config.Scope, &config.RequiredPermission, &config.RequiredCount, &config.TeamID, &config.CreatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

func (db *ApprovalConfigDB) Upsert(ctx context.Context, config domain.ApprovalConfig) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO approval_configs (id, scope, required_permission, required_count, team_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (scope, team_id) DO UPDATE SET
			required_permission = EXCLUDED.required_permission,
			required_count = EXCLUDED.required_count
	`, config.ID, config.Scope, config.RequiredPermission, config.RequiredCount, config.TeamID, config.CreatedAt)
	return err
}

func (db *ApprovalConfigDB) DeleteTeamOverride(ctx context.Context, scope domain.TargetLayer, teamID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM approval_configs WHERE scope = $1 AND team_id = $2
	`, scope, teamID)
	return err
}
