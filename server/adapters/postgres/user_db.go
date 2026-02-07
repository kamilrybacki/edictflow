package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")

type UserDB struct {
	pool *pgxpool.Pool
}

func NewUserDB(pool *pgxpool.Pool) *UserDB {
	return &UserDB{pool: pool}
}

func (db *UserDB) Create(ctx context.Context, user domain.User) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, avatar_url, auth_provider, team_id, created_by, email_verified, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, user.ID, user.Email, user.Name, user.PasswordHash, user.AvatarURL, user.AuthProvider, user.TeamID, user.CreatedBy, user.EmailVerified, user.IsActive, user.CreatedAt)

	if err != nil && err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
		return ErrEmailExists
	}
	return err
}

func (db *UserDB) GetByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User
	err := db.pool.QueryRow(ctx, `
		SELECT id, email, name, COALESCE(password_hash, ''), COALESCE(avatar_url, ''), auth_provider, team_id, created_by, COALESCE(email_verified, false), COALESCE(is_active, true), last_login_at, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AvatarURL, &user.AuthProvider, &user.TeamID, &user.CreatedBy, &user.EmailVerified, &user.IsActive, &user.LastLoginAt, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ErrUserNotFound
	}
	return user, err
}

func (db *UserDB) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := db.pool.QueryRow(ctx, `
		SELECT id, email, name, COALESCE(password_hash, ''), COALESCE(avatar_url, ''), auth_provider, team_id, created_by, COALESCE(email_verified, false), COALESCE(is_active, true), last_login_at, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AvatarURL, &user.AuthProvider, &user.TeamID, &user.CreatedBy, &user.EmailVerified, &user.IsActive, &user.LastLoginAt, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ErrUserNotFound
	}
	return user, err
}

func (db *UserDB) Update(ctx context.Context, user domain.User) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE users SET name = $2, avatar_url = $3, team_id = $4, email_verified = $5, is_active = $6, last_login_at = $7
		WHERE id = $1
	`, user.ID, user.Name, user.AvatarURL, user.TeamID, user.EmailVerified, user.IsActive, user.LastLoginAt)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (db *UserDB) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE users SET password_hash = $2 WHERE id = $1
	`, userID, passwordHash)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (db *UserDB) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE users SET last_login_at = NOW() WHERE id = $1
	`, userID)
	return err
}

func (db *UserDB) List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
	query := `SELECT id, email, name, COALESCE(avatar_url, ''), auth_provider, team_id, email_verified, is_active, last_login_at, created_at FROM users WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if teamID != nil {
		query += fmt.Sprintf(` AND team_id = $%d`, argNum)
		args = append(args, *teamID)
		argNum++
	}
	if activeOnly {
		query += ` AND is_active = true`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.AuthProvider, &user.TeamID, &user.EmailVerified, &user.IsActive, &user.LastLoginAt, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (db *UserDB) Deactivate(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `UPDATE users SET is_active = false WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
