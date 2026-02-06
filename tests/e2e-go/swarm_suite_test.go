// e2e/swarm_suite_test.go
package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// AgentInstance represents a single agent in the swarm
type AgentInstance struct {
	UserID       string
	UserEmail    string
	AgentID      string
	Token        string
	Container    testcontainers.Container
	WorkspaceDir string
	AgentDBDir   string
}

// SwarmSuite holds all resources for a multi-agent E2E test run
type SwarmSuite struct {
	t *testing.T

	// Network for container communication
	network *testcontainers.DockerNetwork

	// Containers
	postgresContainer testcontainers.Container
	serverContainer   testcontainers.Container

	// Agent swarm
	agents []*AgentInstance
	mu     sync.RWMutex

	// Connection info
	postgresHost     string
	postgresPort     string
	serverHostURL    string
	serverNetworkURL string

	// Shared test data
	testTeamID string
	testRuleID string

	// Agent binary path (shared across all agents)
	agentBinaryPath string
}

// SwarmConfig configures the swarm test
type SwarmConfig struct {
	NumAgents int // Number of agent containers to spawn
}

// NewSwarmSuite creates and initializes a multi-agent E2E test environment
func NewSwarmSuite(t *testing.T, cfg SwarmConfig) *SwarmSuite {
	t.Helper()
	ctx := context.Background()

	if cfg.NumAgents < 1 {
		cfg.NumAgents = 3 // Default to 3 agents
	}

	s := &SwarmSuite{
		t:          t,
		testTeamID: uuid.New().String(),
		testRuleID: uuid.New().String(),
		agents:     make([]*AgentInstance, 0, cfg.NumAgents),
	}

	// Create network
	var err error
	s.network, err = network.New(ctx, network.WithCheckDuplicate(), network.WithDriver("bridge"))
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}

	// Build agent binary (shared by all agent containers)
	s.agentBinaryPath = buildAgentBinary(t)

	// Start postgres
	s.postgresContainer, s.postgresHost, s.postgresPort = startPostgres(t, ctx, s.network)

	// Run migrations
	runMigrations(t, s.postgresHost, s.postgresPort)

	// Seed shared team and rule
	s.seedTeamAndRule(t)

	// Create users and seed their data
	for i := 0; i < cfg.NumAgents; i++ {
		agent := s.createAgentInstance(t, i)
		s.agents = append(s.agents, agent)
	}

	// Start server (after seeding all users)
	s.serverContainer, s.serverHostURL, s.serverNetworkURL = startServer(t, ctx, s.network, s.postgresHost)

	// Start all agent containers concurrently
	s.startAgentContainers(t, ctx)

	return s
}

// seedTeamAndRule creates the shared team and rule in the database
func (s *SwarmSuite) seedTeamAndRule(t *testing.T) {
	t.Helper()

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable", s.postgresHost, s.postgresPort)
	db, err := openPostgres(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Insert team
	_, err = db.Exec(`
		INSERT INTO teams (id, name, settings)
		VALUES ($1, 'E2E Swarm Team', '{}')
	`, s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert team: %v", err)
	}

	// Insert rule with block enforcement mode
	triggersJSON := `[{"type":"path","pattern":"CLAUDE.md"}]`
	_, err = db.Exec(`
		INSERT INTO rules (id, name, content, target_layer, priority_weight, triggers, team_id, enforcement_mode, temporary_timeout_hours)
		VALUES ($1, 'Swarm Test Rule', '# Initial Rule Content', 'project', 100, $2, $3, 'block', 24)
	`, s.testRuleID, triggersJSON, s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert rule: %v", err)
	}

	t.Log("Seeded team and rule for swarm test")
}

// createAgentInstance creates a user, agent record, and prepares directories
func (s *SwarmSuite) createAgentInstance(t *testing.T, index int) *AgentInstance {
	t.Helper()

	userID := uuid.New().String()
	agentID := uuid.New().String()
	email := fmt.Sprintf("agent-%d@e2e-test.local", index)

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable", s.postgresHost, s.postgresPort)
	db, err := openPostgres(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Insert user
	_, err = db.Exec(`
		INSERT INTO users (id, email, name, auth_provider, team_id, is_active, email_verified)
		VALUES ($1, $2, $3, 'local', $4, true, true)
	`, userID, email, fmt.Sprintf("Agent %d", index), s.testTeamID)
	if err != nil {
		t.Fatalf("Failed to insert user %d: %v", index, err)
	}

	// Insert agent record
	_, err = db.Exec(`
		INSERT INTO agents (id, machine_id, user_id, status, cached_config_version)
		VALUES ($1, $2, $3, 'offline', 0)
	`, agentID, fmt.Sprintf("swarm-machine-%d", index), userID)
	if err != nil {
		t.Fatalf("Failed to insert agent %d: %v", index, err)
	}

	// Generate token
	token := generateTestToken(t, userID, s.testTeamID)

	// Create temp directories
	workspaceDir, err := os.MkdirTemp("", fmt.Sprintf("swarm-workspace-%d-*", index))
	if err != nil {
		t.Fatalf("Failed to create workspace dir for agent %d: %v", index, err)
	}

	agentDBDir, err := os.MkdirTemp("", fmt.Sprintf("swarm-agentdb-%d-*", index))
	if err != nil {
		t.Fatalf("Failed to create agent DB dir for agent %d: %v", index, err)
	}

	// Create initial CLAUDE.md
	claudeMDPath := filepath.Join(workspaceDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md for agent %d: %v", index, err)
	}

	// Seed agent SQLite database
	s.seedAgentSQLite(t, agentDBDir, token, userID, index)

	agent := &AgentInstance{
		UserID:       userID,
		UserEmail:    email,
		AgentID:      agentID,
		Token:        token,
		WorkspaceDir: workspaceDir,
		AgentDBDir:   agentDBDir,
	}

	t.Logf("Created agent instance %d: user=%s", index, email)
	return agent
}

// seedAgentSQLite creates the SQLite database for an agent
func (s *SwarmSuite) seedAgentSQLite(t *testing.T, agentDBDir, token, userID string, index int) {
	t.Helper()

	dbPath := filepath.Join(agentDBDir, "edictflow.db")
	db, err := openSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to open agent SQLite DB for agent %d: %v", index, err)
	}
	defer db.Close()

	// Create schema
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
		t.Fatalf("Failed to create agent schema for agent %d: %v", index, err)
	}

	// Insert auth
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	_, err = db.Exec(`
		INSERT INTO auth (id, access_token, refresh_token, expires_at, user_id, user_email, user_name)
		VALUES (1, ?, '', ?, ?, ?, ?)
	`, token, expiresAt, userID, fmt.Sprintf("agent-%d@e2e-test.local", index), fmt.Sprintf("Agent %d", index))
	if err != nil {
		t.Fatalf("Failed to insert auth for agent %d: %v", index, err)
	}

	// Insert cached rule
	_, err = db.Exec(`
		INSERT INTO cached_rules (id, name, content, target_layer, triggers, enforcement_mode, temporary_timeout_hours, version, cached_at)
		VALUES (?, 'Swarm Test Rule', '# Initial Rule Content', 'project', '[{"type":"path","pattern":"CLAUDE.md"}]', 'block', 24, 1, ?)
	`, s.testRuleID, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert cached rule for agent %d: %v", index, err)
	}

	// Insert watched project
	_, err = db.Exec(`
		INSERT INTO watched_projects (path, detected_context, detected_tags, last_sync_at)
		VALUES ('/workspace', 'e2e-swarm', '', ?)
	`, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert watched project for agent %d: %v", index, err)
	}
}

// startAgentContainers starts all agent containers concurrently
func (s *SwarmSuite) startAgentContainers(t *testing.T, ctx context.Context) {
	t.Helper()

	var wg sync.WaitGroup
	errors := make(chan error, len(s.agents))

	for i, agent := range s.agents {
		wg.Add(1)
		go func(idx int, a *AgentInstance) {
			defer wg.Done()

			container := startAgentForSwarm(t, ctx, s.network, s.serverNetworkURL,
				a.WorkspaceDir, a.AgentDBDir, s.agentBinaryPath, idx)

			s.mu.Lock()
			a.Container = container
			s.mu.Unlock()
		}(i, agent)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Fatalf("Failed to start agent container: %v", err)
		}
	}

	t.Logf("Started %d agent containers", len(s.agents))
}

// Cleanup tears down all swarm resources
func (s *SwarmSuite) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Stop all agent containers concurrently
	var wg sync.WaitGroup
	for _, agent := range s.agents {
		wg.Add(1)
		go func(a *AgentInstance) {
			defer wg.Done()
			if a.Container != nil {
				a.Container.Terminate(ctx)
			}
			if a.WorkspaceDir != "" {
				os.RemoveAll(a.WorkspaceDir)
			}
			if a.AgentDBDir != "" {
				os.RemoveAll(a.AgentDBDir)
			}
		}(agent)
	}
	wg.Wait()

	// Stop server and postgres
	if s.serverContainer != nil {
		s.serverContainer.Terminate(ctx)
	}
	if s.postgresContainer != nil {
		s.postgresContainer.Terminate(ctx)
	}

	// Remove network
	if s.network != nil {
		s.network.Remove(ctx)
	}

	// Clean up agent binary
	if s.agentBinaryPath != "" {
		os.RemoveAll(filepath.Dir(s.agentBinaryPath))
	}
}

// GetAgent returns an agent by index
func (s *SwarmSuite) GetAgent(index int) *AgentInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index >= 0 && index < len(s.agents) {
		return s.agents[index]
	}
	return nil
}

// GetCLAUDEMDPath returns the CLAUDE.md path for a specific agent
func (s *SwarmSuite) GetCLAUDEMDPath(agentIndex int) string {
	agent := s.GetAgent(agentIndex)
	if agent != nil {
		return filepath.Join(agent.WorkspaceDir, "CLAUDE.md")
	}
	return ""
}

// WaitForAllAgentsSync waits for all agents to connect and sync
func (s *SwarmSuite) WaitForAllAgentsSync(t *testing.T, timeout time.Duration) bool {
	t.Helper()

	var wg sync.WaitGroup
	results := make(chan bool, len(s.agents))

	for i, agent := range s.agents {
		wg.Add(1)
		go func(idx int, a *AgentInstance) {
			defer wg.Done()
			synced := waitForAgentSync(t, a.Container, timeout)
			if synced {
				t.Logf("Agent %d synced successfully", idx)
			} else {
				t.Logf("Agent %d failed to sync", idx)
			}
			results <- synced
		}(i, agent)
	}

	wg.Wait()
	close(results)

	allSynced := true
	for synced := range results {
		if !synced {
			allSynced = false
		}
	}

	return allSynced
}
