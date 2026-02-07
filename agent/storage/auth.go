// agent/storage/auth.go
package storage

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNotLoggedIn = errors.New("not logged in")

type AuthInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string
	UserEmail    string
	UserName     string
	TeamID       string
}

func (s *Storage) SaveAuth(auth AuthInfo) error {
	query := `
		INSERT OR REPLACE INTO auth (id, access_token, refresh_token, expires_at, user_id, user_email, user_name, team_id)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, auth.AccessToken, auth.RefreshToken, auth.ExpiresAt.Unix(), auth.UserID, auth.UserEmail, auth.UserName, auth.TeamID)
	return err
}

func (s *Storage) GetAuth() (AuthInfo, error) {
	query := `SELECT access_token, refresh_token, expires_at, user_id, user_email, user_name, COALESCE(team_id, '') FROM auth WHERE id = 1`
	var auth AuthInfo
	var expiresAt int64
	err := s.db.QueryRow(query).Scan(&auth.AccessToken, &auth.RefreshToken, &expiresAt, &auth.UserID, &auth.UserEmail, &auth.UserName, &auth.TeamID)
	if err == sql.ErrNoRows {
		return AuthInfo{}, ErrNotLoggedIn
	}
	if err != nil {
		return AuthInfo{}, err
	}
	auth.ExpiresAt = time.Unix(expiresAt, 0)
	return auth, nil
}

func (s *Storage) ClearAuth() error {
	_, err := s.db.Exec("DELETE FROM auth")
	return err
}

func (s *Storage) IsLoggedIn() bool {
	auth, err := s.GetAuth()
	if err != nil {
		return false
	}
	return time.Now().Before(auth.ExpiresAt)
}
