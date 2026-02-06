package postgres

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type ExceptionRequestDB struct {
	pool *pgxpool.Pool
}

func NewExceptionRequestDB(pool *pgxpool.Pool) *ExceptionRequestDB {
	return &ExceptionRequestDB{pool: pool}
}

func (db *ExceptionRequestDB) Create(ctx context.Context, er domain.ExceptionRequest) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO exception_requests (
			id, change_request_id, user_id, justification, exception_type,
			expires_at, status, created_at, resolved_at, resolved_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, er.ID, er.ChangeRequestID, er.UserID, er.Justification, er.ExceptionType,
		er.ExpiresAt, er.Status, er.CreatedAt, er.ResolvedAt, er.ResolvedByUserID)
	return err
}

func (db *ExceptionRequestDB) GetByID(ctx context.Context, id string) (*domain.ExceptionRequest, error) {
	var er domain.ExceptionRequest
	err := db.pool.QueryRow(ctx, `
		SELECT id, change_request_id, user_id, justification, exception_type,
			expires_at, status, created_at, resolved_at, resolved_by_user_id
		FROM exception_requests WHERE id = $1
	`, id).Scan(
		&er.ID, &er.ChangeRequestID, &er.UserID, &er.Justification, &er.ExceptionType,
		&er.ExpiresAt, &er.Status, &er.CreatedAt, &er.ResolvedAt, &er.ResolvedByUserID,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &er, nil
}

type ExceptionRequestFilter struct {
	Status          *domain.ExceptionRequestStatus
	ExceptionType   *domain.ExceptionType
	ChangeRequestID *string
	UserID          *string
	Limit           int
	Offset          int
}

func (db *ExceptionRequestDB) ListByTeam(ctx context.Context, teamID string, filter ExceptionRequestFilter) ([]domain.ExceptionRequest, error) {
	query := `
		SELECT er.id, er.change_request_id, er.user_id, er.justification, er.exception_type,
			er.expires_at, er.status, er.created_at, er.resolved_at, er.resolved_by_user_id
		FROM exception_requests er
		JOIN change_requests cr ON er.change_request_id = cr.id
		WHERE cr.team_id = $1
	`
	args := []interface{}{teamID}
	argIdx := 2

	if filter.Status != nil {
		query += ` AND er.status = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.ExceptionType != nil {
		query += ` AND er.exception_type = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.ExceptionType)
		argIdx++
	}
	if filter.ChangeRequestID != nil {
		query += ` AND er.change_request_id = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.ChangeRequestID)
		argIdx++
	}
	if filter.UserID != nil {
		query += ` AND er.user_id = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	query += ` ORDER BY er.created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT $` + strconv.Itoa(argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + strconv.Itoa(argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ExceptionRequest
	for rows.Next() {
		var er domain.ExceptionRequest
		if err := rows.Scan(
			&er.ID, &er.ChangeRequestID, &er.UserID, &er.Justification, &er.ExceptionType,
			&er.ExpiresAt, &er.Status, &er.CreatedAt, &er.ResolvedAt, &er.ResolvedByUserID,
		); err != nil {
			return nil, err
		}
		results = append(results, er)
	}
	return results, rows.Err()
}

func (db *ExceptionRequestDB) Update(ctx context.Context, er domain.ExceptionRequest) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE exception_requests SET
			expires_at = $2, status = $3, resolved_at = $4, resolved_by_user_id = $5
		WHERE id = $1
	`, er.ID, er.ExpiresAt, er.Status, er.ResolvedAt, er.ResolvedByUserID)
	return err
}

func (db *ExceptionRequestDB) FindActiveByUserRuleFile(ctx context.Context, userID, ruleID, filePath string) (*domain.ExceptionRequest, error) {
	var er domain.ExceptionRequest
	err := db.pool.QueryRow(ctx, `
		SELECT er.id, er.change_request_id, er.user_id, er.justification, er.exception_type,
			er.expires_at, er.status, er.created_at, er.resolved_at, er.resolved_by_user_id
		FROM exception_requests er
		JOIN change_requests cr ON er.change_request_id = cr.id
		WHERE er.user_id = $1
			AND cr.rule_id = $2
			AND cr.file_path = $3
			AND er.status = 'approved'
			AND (er.expires_at IS NULL OR er.expires_at > now())
		ORDER BY er.created_at DESC
		LIMIT 1
	`, userID, ruleID, filePath).Scan(
		&er.ID, &er.ChangeRequestID, &er.UserID, &er.Justification, &er.ExceptionType,
		&er.ExpiresAt, &er.Status, &er.CreatedAt, &er.ResolvedAt, &er.ResolvedByUserID,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &er, nil
}

func (db *ExceptionRequestDB) FindByChangeRequest(ctx context.Context, changeRequestID string) ([]domain.ExceptionRequest, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, change_request_id, user_id, justification, exception_type,
			expires_at, status, created_at, resolved_at, resolved_by_user_id
		FROM exception_requests
		WHERE change_request_id = $1
		ORDER BY created_at DESC
	`, changeRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.ExceptionRequest
	for rows.Next() {
		var er domain.ExceptionRequest
		if err := rows.Scan(
			&er.ID, &er.ChangeRequestID, &er.UserID, &er.Justification, &er.ExceptionType,
			&er.ExpiresAt, &er.Status, &er.CreatedAt, &er.ResolvedAt, &er.ResolvedByUserID,
		); err != nil {
			return nil, err
		}
		results = append(results, er)
	}
	return results, rows.Err()
}
