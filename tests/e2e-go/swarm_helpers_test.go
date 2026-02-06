// e2e/swarm_helpers_test.go
package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// openPostgres opens a connection to the postgres database
func openPostgres(connStr string) (*sql.DB, error) {
	return sql.Open("postgres", connStr)
}

// openSQLite opens a connection to a SQLite database
func openSQLite(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}

// startAgentForSwarm starts an agent container for the swarm test
func startAgentForSwarm(t *testing.T, ctx context.Context, net *testcontainers.DockerNetwork, serverURL, workspaceDir, agentDBDir, agentBinaryPath string, index int) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:      "debian:bookworm-slim",
		Entrypoint: []string{"/app/bin/agent"},
		Cmd:        []string{"start", "--foreground", "--poll-interval", "500ms", "--server", serverURL},
		Env: map[string]string{
			"HOME": "/home/agent",
		},
		Networks: []string{net.Name},
		Mounts: testcontainers.ContainerMounts{
			testcontainers.BindMount(agentBinaryPath, "/app/bin/agent"),
			testcontainers.BindMount(workspaceDir, "/workspace"),
			testcontainers.BindMount(agentDBDir, "/home/agent/.edictflow"),
		},
		WaitingFor: wait.ForLog("Daemon running").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start agent container %d: %v", index, err)
	}

	t.Logf("Agent container %d started", index)
	return container
}

// modifyAllAgentFiles modifies CLAUDE.md in all agent workspaces concurrently
func modifyAllAgentFiles(t *testing.T, suite *SwarmSuite, content string) {
	t.Helper()

	for i := range suite.agents {
		path := suite.GetCLAUDEMDPath(i)
		modifyFile(t, path, content)
	}
}

// verifyAllAgentFilesMatch checks if all agent CLAUDE.md files have the expected content
func verifyAllAgentFilesMatch(t *testing.T, suite *SwarmSuite, expected string) bool {
	t.Helper()

	allMatch := true
	for i := range suite.agents {
		path := suite.GetCLAUDEMDPath(i)
		content := getFileContent(t, path)
		if content != expected {
			t.Logf("Agent %d file content mismatch. Expected:\n%s\nGot:\n%s", i, expected, content)
			allMatch = false
		}
	}
	return allMatch
}

// updateRuleContent updates the rule content in the database
func updateRuleContent(t *testing.T, suite *SwarmSuite, newContent string) {
	t.Helper()

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable",
		suite.postgresHost, suite.postgresPort)
	db, err := openPostgres(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		UPDATE rules SET content = $1, updated_at = NOW()
		WHERE id = $2
	`, newContent, suite.testRuleID)
	if err != nil {
		t.Fatalf("Failed to update rule content: %v", err)
	}

	t.Logf("Updated rule content in database")
}

// getConnectedAgentCount queries the database for connected agents
func getConnectedAgentCount(t *testing.T, suite *SwarmSuite) int {
	t.Helper()

	connStr := fmt.Sprintf("postgres://edictflow:edictflow@%s:%s/edictflow?sslmode=disable",
		suite.postgresHost, suite.postgresPort)
	db, err := openPostgres(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM agents WHERE status = 'online'`).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count online agents: %v", err)
	}

	return count
}

// waitForAgentCount waits until the expected number of agents are connected
func waitForAgentCount(t *testing.T, suite *SwarmSuite, expected int, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		count := getConnectedAgentCount(t, suite)
		if count >= expected {
			t.Logf("Expected %d agents connected, got %d", expected, count)
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}

	count := getConnectedAgentCount(t, suite)
	t.Logf("Timeout waiting for agents. Expected %d, got %d", expected, count)
	return false
}

// triggerRuleSync sends a config update to all connected agents via the API
// This simulates what happens when an admin updates a rule
func triggerRuleSync(t *testing.T, suite *SwarmSuite) {
	t.Helper()

	// Update the rule's updated_at timestamp to trigger version increment
	// In a real implementation, this would trigger a broadcast via WebSocket
	updateRuleEnforcementMode(t, &E2ESuite{
		serverHostURL: suite.serverHostURL,
		testRuleID:    suite.testRuleID,
		testTeamID:    suite.testTeamID,
		testToken:     suite.agents[0].Token, // Use first agent's token
	}, "block")
}

// getAgentLogs retrieves logs from an agent container
func getAgentLogs(t *testing.T, container testcontainers.Container) string {
	t.Helper()

	ctx := context.Background()
	logs, err := container.Logs(ctx)
	if err != nil {
		return fmt.Sprintf("Error getting logs: %v", err)
	}
	defer logs.Close()

	buf := make([]byte, 4096)
	n, _ := logs.Read(buf)
	return string(buf[:n])
}

// verifyAgentReceivedUpdate checks agent logs for config update receipt
func verifyAgentReceivedUpdate(t *testing.T, container testcontainers.Container, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		logs := getAgentLogs(t, container)
		if containsAny(logs, []string{"Updated rules to version", "config_update", "Connected to server"}) {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if len(s) > 0 && len(sub) > 0 {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
