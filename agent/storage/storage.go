// agent/storage/storage.go
package storage

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(configDir, "edictflow.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".edictflow"), nil
}

func GetConfigDir() (string, error) {
	return getConfigDir()
}

// SaveServerURL saves the API server URL to config
func (s *Storage) SaveServerURL(url string) error {
	_, err := s.db.Exec(`
		INSERT INTO config (key, value) VALUES ('server_url', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, url)
	return err
}

// GetServerURL retrieves the saved API server URL
func (s *Storage) GetServerURL() (string, error) {
	var url string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = 'server_url'`).Scan(&url)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return url, err
}

// SaveWSServerURL saves the WebSocket server URL to config
func (s *Storage) SaveWSServerURL(url string) error {
	_, err := s.db.Exec(`
		INSERT INTO config (key, value) VALUES ('ws_server_url', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, url)
	return err
}

// GetWSServerURL retrieves the saved WebSocket server URL
func (s *Storage) GetWSServerURL() (string, error) {
	var url string
	err := s.db.QueryRow(`SELECT value FROM config WHERE key = 'ws_server_url'`).Scan(&url)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return url, err
}
