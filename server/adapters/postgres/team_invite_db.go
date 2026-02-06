package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrInviteNotFound = errors.New("invite not found")
	ErrInviteExpired  = errors.New("invite expired or max uses reached")
)

type TeamInviteDB struct {
	pool *pgxpool.Pool
}

func NewTeamInviteDB(pool *pgxpool.Pool) *TeamInviteDB {
	return &TeamInviteDB{pool: pool}
}

func (db *TeamInviteDB) Create(ctx context.Context, invite domain.TeamInvite) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO team_invites (id, team_id, code, max_uses, use_count, expires_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invite.ID, invite.TeamID, invite.Code, invite.MaxUses, invite.UseCount, invite.ExpiresAt, invite.CreatedBy, invite.CreatedAt)
	return err
}

func (db *TeamInviteDB) GetByCode(ctx context.Context, code string) (domain.TeamInvite, error) {
	var invite domain.TeamInvite
	err := db.pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE code = $1
	`, code).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	return invite, err
}

func (db *TeamInviteDB) GetByID(ctx context.Context, id string) (domain.TeamInvite, error) {
	var invite domain.TeamInvite
	err := db.pool.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE id = $1
	`, id).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	return invite, err
}

func (db *TeamInviteDB) ListByTeam(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE team_id = $1 AND expires_at > NOW() AND use_count < max_uses
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []domain.TeamInvite
	for rows.Next() {
		var invite domain.TeamInvite
		if err := rows.Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (db *TeamInviteDB) Delete(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `DELETE FROM team_invites WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}
	return nil
}

// IncrementUseCountAtomic atomically increments use_count and returns the updated invite.
// Uses SELECT FOR UPDATE to prevent race conditions.
// Returns ErrInviteExpired if the invite is expired or max uses reached.
func (db *TeamInviteDB) IncrementUseCountAtomic(ctx context.Context, code string) (domain.TeamInvite, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return domain.TeamInvite{}, err
	}
	defer tx.Rollback(ctx)

	var invite domain.TeamInvite
	err = tx.QueryRow(ctx, `
		SELECT id, team_id, code, max_uses, use_count, expires_at, created_by, created_at
		FROM team_invites
		WHERE code = $1
		FOR UPDATE
	`, code).Scan(&invite.ID, &invite.TeamID, &invite.Code, &invite.MaxUses, &invite.UseCount, &invite.ExpiresAt, &invite.CreatedBy, &invite.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TeamInvite{}, ErrInviteNotFound
	}
	if err != nil {
		return domain.TeamInvite{}, err
	}

	if !invite.IsValid() {
		return domain.TeamInvite{}, ErrInviteExpired
	}

	_, err = tx.Exec(ctx, `
		UPDATE team_invites SET use_count = use_count + 1 WHERE id = $1
	`, invite.ID)
	if err != nil {
		return domain.TeamInvite{}, err
	}

	invite.UseCount++

	if err := tx.Commit(ctx); err != nil {
		return domain.TeamInvite{}, err
	}

	return invite, nil
}
