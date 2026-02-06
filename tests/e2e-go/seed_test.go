// e2e/seed_test.go
package e2e

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// e2eJWTSecret is the shared JWT secret for E2E tests
// Must match the value in helpers_test.go server container env
const e2eJWTSecret = "e2e-test-secret-do-not-use-in-production"

// runMigrations runs server migrations against Postgres
func runMigrations(t *testing.T, host, port string) {
	t.Helper()

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable", host, port)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Read and execute migration files
	migrationsDir := filepath.Join("..", "server", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("Failed to read migrations dir: %v", err)
	}

	// Sort entries to ensure migrations run in order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		// Only run "up" migrations
		if len(entry.Name()) < 7 || entry.Name()[len(entry.Name())-7:] != ".up.sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			t.Fatalf("Failed to read migration %s: %v", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("Failed to execute migration %s: %v", entry.Name(), err)
		}
		t.Logf("Applied migration: %s", entry.Name())
	}
}

// seedDatabase inserts test data into Postgres
func seedDatabase(t *testing.T, s *E2ESuite) string {
	t.Helper()

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable", s.postgresHost, s.postgresPort)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Insert team
	_, err = db.Exec(`
		INSERT INTO teams (id, name, settings)
		VALUES ($1, 'E2E Test Team', '{}')
	`, s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert team: %v", err)
	}

	// Insert user (with columns from migrations 002 and 009)
	_, err = db.Exec(`
		INSERT INTO users (id, email, name, auth_provider, team_id, is_active, email_verified)
		VALUES ($1, 'e2e-test@example.com', 'E2E Test User', 'local', $2, true, true)
	`, s.testUserID, s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert rule with block enforcement mode
	// Triggers must be an array of Trigger objects, not just strings
	triggersJSON := `[{"type":"path","pattern":"CLAUDE.md"}]`
	_, err = db.Exec(`
		INSERT INTO rules (id, name, content, target_layer, priority_weight, triggers, team_id, enforcement_mode, temporary_timeout_hours)
		VALUES ($1, 'E2E Test Rule', '# Test Rule Content', 'project', 100, $2, $3, 'block', 24)
	`, s.testRuleID, triggersJSON, s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert rule: %v", err)
	}

	// Insert agent
	_, err = db.Exec(`
		INSERT INTO agents (id, machine_id, user_id, status, cached_config_version)
		VALUES ($1, 'e2e-test-machine', $2, 'online', 1)
	`, s.testAgentID, s.testUserID)
	if err != nil {
		t.Fatalf("Failed to insert agent: %v", err)
	}

	t.Log("Seeded database with test data")

	// Generate test token
	return generateTestToken(t, s.testUserID, s.testTeamID)
}

// generateTestToken creates a valid JWT for API requests
func generateTestToken(t *testing.T, userID, teamID string) string {
	t.Helper()

	claims := jwt.MapClaims{
		"sub":     userID,
		"team_id": teamID,
		"email":   "e2e-test@example.com",
		"name":    "E2E Test User",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(e2eJWTSecret))
	if err != nil {
		t.Fatalf("Failed to sign JWT: %v", err)
	}

	t.Log("Generated test JWT token")
	return tokenString
}

// seedAgentDB creates the SQLite database for the agent
func seedAgentDB(t *testing.T, s *E2ESuite) {
	t.Helper()

	dbPath := filepath.Join(s.agentDBDir, "edictflow.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open agent SQLite DB: %v", err)
	}
	defer db.Close()

	// Create schema (matching agent/storage/migrations.go)
	schema := `
		CREATE TABLE IF NOT EXISTS auth (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			access_token TEXT NOT NULL,
			refresh_token TEXT,
			expires_at INTEGER NOT NULL,
			user_id TEXT NOT NULL,
			user_email TEXT NOT NULL,
			user_name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS message_queue (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ref_id TEXT UNIQUE NOT NULL,
			msg_type TEXT NOT NULL,
			payload TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			attempts INTEGER DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS cached_rules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			content TEXT NOT NULL,
			target_layer TEXT NOT NULL,
			triggers TEXT NOT NULL,
			enforcement_mode TEXT NOT NULL,
			temporary_timeout_hours INTEGER,
			version INTEGER NOT NULL,
			cached_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS watched_projects (
			path TEXT PRIMARY KEY,
			detected_context TEXT,
			detected_tags TEXT,
			last_sync_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS pending_changes (
			id TEXT PRIMARY KEY,
			rule_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			original_content TEXT NOT NULL,
			modified_content TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create agent schema: %v", err)
	}

	// Insert auth credentials
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	_, err = db.Exec(`
		INSERT INTO auth (id, access_token, refresh_token, expires_at, user_id, user_email, user_name)
		VALUES (1, ?, '', ?, ?, 'e2e-test@example.com', 'E2E Test User')
	`, s.testToken, expiresAt, s.testUserID)
	if err != nil {
		t.Fatalf("Failed to insert auth: %v", err)
	}

	// Insert cached rule
	_, err = db.Exec(`
		INSERT INTO cached_rules (id, name, content, target_layer, triggers, enforcement_mode, temporary_timeout_hours, version, cached_at)
		VALUES (?, 'E2E Test Rule', '# Test Rule Content', 'project', '["CLAUDE.md"]', 'block', 24, 1, ?)
	`, s.testRuleID, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert cached rule: %v", err)
	}

	// Insert watched project
	_, err = db.Exec(`
		INSERT INTO watched_projects (path, detected_context, detected_tags, last_sync_at)
		VALUES ('/workspace', 'e2e-test', '', ?)
	`, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert watched project: %v", err)
	}

	t.Logf("Seeded agent SQLite database at %s", dbPath)
}
