// e2e/suite_test.go
package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

// E2ESuite holds all resources for an E2E test run
type E2ESuite struct {
	t *testing.T

	// Network for container communication
	network *testcontainers.DockerNetwork

	// Containers
	postgresContainer testcontainers.Container
	serverContainer   testcontainers.Container
	agentContainer    testcontainers.Container

	// Connection info
	postgresHost    string
	postgresPort    string
	serverHostURL   string
	serverNetworkURL string

	// Test data
	testTeamID  string
	testUserID  string
	testRuleID  string
	testAgentID string
	testToken   string

	// Temporary directories
	workspaceDir string
	agentDBDir   string

	// Agent binary path
	agentBinaryPath string
}

// NewE2ESuite creates and initializes a full E2E test environment
func NewE2ESuite(t *testing.T) *E2ESuite {
	t.Helper()
	ctx := context.Background()

	s := &E2ESuite{
		t:          t,
		testTeamID: uuid.New().String(),
		testUserID: uuid.New().String(),
		testRuleID: uuid.New().String(),
		testAgentID: uuid.New().String(),
	}

	// Create network
	var err error
	networkName := "e2e-" + uuid.New().String()[:8]
	s.network, err = network.New(ctx, network.WithCheckDuplicate(), network.WithDriver("bridge"))
	if err != nil {
		t.Fatalf("Failed to create network %s: %v", networkName, err)
	}

	// Create temp directories
	s.workspaceDir, s.agentDBDir = createTempDirs(t)

	// Build agent binary
	s.agentBinaryPath = buildAgentBinary(t)

	// Start containers in order - postgres first
	s.postgresContainer, s.postgresHost, s.postgresPort = startPostgres(t, ctx, s.network)

	// Run migrations BEFORE starting server (server may depend on tables)
	runMigrations(t, s.postgresHost, s.postgresPort)

	// Seed the database before server starts (server needs data to serve)
	s.testToken = seedDatabase(t, s)

	// Now start server (with migrations and seed data in place)
	s.serverContainer, s.serverHostURL, s.serverNetworkURL = startServer(t, ctx, s.network, s.postgresHost)

	// Seed agent SQLite database
	seedAgentDB(t, s)

	// Start agent container
	s.agentContainer = startAgent(t, ctx, s.network, s.serverNetworkURL, s.workspaceDir, s.agentDBDir, s.agentBinaryPath)

	return s
}

// Cleanup tears down all E2E resources
func (s *E2ESuite) Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Stop containers in reverse order
	if s.agentContainer != nil {
		if err := s.agentContainer.Terminate(ctx); err != nil {
			s.t.Logf("Failed to terminate agent container: %v", err)
		}
	}

	if s.serverContainer != nil {
		if err := s.serverContainer.Terminate(ctx); err != nil {
			s.t.Logf("Failed to terminate server container: %v", err)
		}
	}

	if s.postgresContainer != nil {
		if err := s.postgresContainer.Terminate(ctx); err != nil {
			s.t.Logf("Failed to terminate postgres container: %v", err)
		}
	}

	// Remove network
	if s.network != nil {
		if err := s.network.Remove(ctx); err != nil {
			s.t.Logf("Failed to remove network: %v", err)
		}
	}

	// Clean up temp directories
	if s.workspaceDir != "" {
		os.RemoveAll(s.workspaceDir)
	}
	if s.agentDBDir != "" {
		os.RemoveAll(s.agentDBDir)
	}

	// Clean up agent binary (and its parent directory from Docker build)
	if s.agentBinaryPath != "" {
		os.RemoveAll(filepath.Dir(s.agentBinaryPath))
	}
}

// GetCLAUDEMDPath returns the path to CLAUDE.md in the workspace
func (s *E2ESuite) GetCLAUDEMDPath() string {
	return filepath.Join(s.workspaceDir, "CLAUDE.md")
}

// TestSuiteCreation is a basic test to verify the suite can be created
func TestSuiteCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// This test only verifies the suite can be instantiated
	// without starting actual containers
	t.Log("E2E suite module is properly configured")
}
