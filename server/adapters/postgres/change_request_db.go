package postgres

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type ChangeRequestDB struct {
	pool *pgxpool.Pool
}

func NewChangeRequestDB(pool *pgxpool.Pool) *ChangeRequestDB {
	return &ChangeRequestDB{pool: pool}
}

func (db *ChangeRequestDB) Create(ctx context.Context, cr domain.ChangeRequest) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO change_requests (
			id, rule_id, agent_id, user_id, team_id, file_path,
			original_hash, modified_hash, diff_content, status,
			enforcement_mode, timeout_at, created_at, resolved_at, resolved_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, cr.ID, cr.RuleID, cr.AgentID, cr.UserID, cr.TeamID, cr.FilePath,
		cr.OriginalHash, cr.ModifiedHash, cr.DiffContent, cr.Status,
		cr.EnforcementMode, cr.TimeoutAt, cr.CreatedAt, cr.ResolvedAt, cr.ResolvedByUserID)
	return err
}

func (db *ChangeRequestDB) GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error) {
	var cr domain.ChangeRequest
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, agent_id, user_id, team_id, file_path,
			original_hash, modified_hash, diff_content, status,
			enforcement_mode, timeout_at, created_at, resolved_at, resolved_by_user_id
		FROM change_requests WHERE id = $1
	`, id).Scan(
		&cr.ID, &cr.RuleID, &cr.AgentID, &cr.UserID, &cr.TeamID, &cr.FilePath,
		&cr.OriginalHash, &cr.ModifiedHash, &cr.DiffContent, &cr.Status,
		&cr.EnforcementMode, &cr.TimeoutAt, &cr.CreatedAt, &cr.ResolvedAt, &cr.ResolvedByUserID,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

type ChangeRequestFilter struct {
	Status          *domain.ChangeRequestStatus
	EnforcementMode *domain.EnforcementMode
	RuleID          *string
	AgentID         *string
	UserID          *string
	Limit           int
	Offset          int
}

func (db *ChangeRequestDB) ListByTeam(ctx context.Context, teamID string, filter ChangeRequestFilter) ([]domain.ChangeRequest, error) {
	query := `
		SELECT id, rule_id, agent_id, user_id, team_id, file_path,
			original_hash, modified_hash, diff_content, status,
			enforcement_mode, timeout_at, created_at, resolved_at, resolved_by_user_id
		FROM change_requests WHERE team_id = $1
	`
	args := []interface{}{teamID}
	argIdx := 2

	if filter.Status != nil {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.EnforcementMode != nil {
		query += ` AND enforcement_mode = $` + itoa(argIdx)
		args = append(args, *filter.EnforcementMode)
		argIdx++
	}
	if filter.RuleID != nil {
		query += ` AND rule_id = $` + itoa(argIdx)
		args = append(args, *filter.RuleID)
		argIdx++
	}
	if filter.AgentID != nil {
		query += ` AND agent_id = $` + itoa(argIdx)
		args = append(args, *filter.AgentID)
		argIdx++
	}
	if filter.UserID != nil {
		query += ` AND user_id = $` + itoa(argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT $` + itoa(argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + itoa(argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ChangeRequest
	for rows.Next() {
		var cr domain.ChangeRequest
		if err := rows.Scan(
			&cr.ID, &cr.RuleID, &cr.AgentID, &cr.UserID, &cr.TeamID, &cr.FilePath,
			&cr.OriginalHash, &cr.ModifiedHash, &cr.DiffContent, &cr.Status,
			&cr.EnforcementMode, &cr.TimeoutAt, &cr.CreatedAt, &cr.ResolvedAt, &cr.ResolvedByUserID,
		); err != nil {
			return nil, err
		}
		results = append(results, cr)
	}
	return results, rows.Err()
}

func (db *ChangeRequestDB) Update(ctx context.Context, cr domain.ChangeRequest) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE change_requests SET
			modified_hash = $2, diff_content = $3, status = $4,
			timeout_at = $5, resolved_at = $6, resolved_by_user_id = $7
		WHERE id = $1
	`, cr.ID, cr.ModifiedHash, cr.DiffContent, cr.Status,
		cr.TimeoutAt, cr.ResolvedAt, cr.ResolvedByUserID)
	return err
}

func (db *ChangeRequestDB) FindExpiredTemporary(ctx context.Context, before time.Time) ([]domain.ChangeRequest, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, agent_id, user_id, team_id, file_path,
			original_hash, modified_hash, diff_content, status,
			enforcement_mode, timeout_at, created_at, resolved_at, resolved_by_user_id
		FROM change_requests
		WHERE status = 'pending'
			AND enforcement_mode = 'temporary'
			AND timeout_at IS NOT NULL
			AND timeout_at < $1
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ChangeRequest
	for rows.Next() {
		var cr domain.ChangeRequest
		if err := rows.Scan(
			&cr.ID, &cr.RuleID, &cr.AgentID, &cr.UserID, &cr.TeamID, &cr.FilePath,
			&cr.OriginalHash, &cr.ModifiedHash, &cr.DiffContent, &cr.Status,
			&cr.EnforcementMode, &cr.TimeoutAt, &cr.CreatedAt, &cr.ResolvedAt, &cr.ResolvedByUserID,
		); err != nil {
			return nil, err
		}
		results = append(results, cr)
	}
	return results, rows.Err()
}

func (db *ChangeRequestDB) FindByAgentAndFile(ctx context.Context, agentID, filePath string) (*domain.ChangeRequest, error) {
	var cr domain.ChangeRequest
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, agent_id, user_id, team_id, file_path,
			original_hash, modified_hash, diff_content, status,
			enforcement_mode, timeout_at, created_at, resolved_at, resolved_by_user_id
		FROM change_requests
		WHERE agent_id = $1 AND file_path = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`, agentID, filePath).Scan(
		&cr.ID, &cr.RuleID, &cr.AgentID, &cr.UserID, &cr.TeamID, &cr.FilePath,
		&cr.OriginalHash, &cr.ModifiedHash, &cr.DiffContent, &cr.Status,
		&cr.EnforcementMode, &cr.TimeoutAt, &cr.CreatedAt, &cr.ResolvedAt, &cr.ResolvedByUserID,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (db *ChangeRequestDB) CountByTeamAndStatus(ctx context.Context, teamID string, status domain.ChangeRequestStatus) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM change_requests WHERE team_id = $1 AND status = $2
	`, teamID, status).Scan(&count)
	return count, err
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
