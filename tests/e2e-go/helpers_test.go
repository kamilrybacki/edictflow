// e2e/helpers_test.go
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	// Disable Ryuk (container cleanup daemon) - we handle cleanup manually
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	// Use host network for container communication
	os.Setenv("TESTCONTAINERS_HOST_OVERRIDE", "127.0.0.1")
}

// buildAgentBinary compiles the agent for Linux using Docker for cross-compilation
// This ensures CGO works correctly with sqlite3
func buildAgentBinary(t *testing.T) string {
	t.Helper()

	// Create temp directory for build output
	tmpDir, err := os.MkdirTemp("", "agent-build-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir for agent binary: %v", err)
	}
	binaryPath := filepath.Join(tmpDir, "agent")

	agentDir, err := filepath.Abs(filepath.Join("..", "agent"))
	if err != nil {
		t.Fatalf("Failed to get agent dir path: %v", err)
	}

	// Use Docker to build the agent for Linux with CGO support
	// This uses golang:1.23-alpine which includes musl for static linking
	// Note: Match this with the Go version in server/Dockerfile
	cmd := exec.Command("docker", "run", "--rm",
		"-v", agentDir+":/src",
		"-v", tmpDir+":/out",
		"-w", "/src",
		"golang:1.23-alpine",
		"sh", "-c", "apk add --no-cache gcc musl-dev sqlite-dev && CGO_ENABLED=1 go build -ldflags='-s -w -linkmode external -extldflags -static' -o /out/agent ./cmd/agent",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build agent binary: %v\nOutput: %s", err, output)
	}

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Agent binary was not created at %s", binaryPath)
	}

	// Make executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		t.Fatalf("Failed to chmod agent binary: %v", err)
	}

	t.Logf("Built agent binary at %s", binaryPath)
	return binaryPath
}

// createTempDirs creates temporary directories for workspace and agent DB
func createTempDirs(t *testing.T) (workspaceDir, agentDBDir string) {
	t.Helper()

	var err error
	workspaceDir, err = os.MkdirTemp("", "e2e-workspace-*")
	if err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	agentDBDir, err = os.MkdirTemp("", "e2e-agentdb-*")
	if err != nil {
		t.Fatalf("Failed to create agent DB dir: %v", err)
	}

	// Create initial CLAUDE.md
	claudeMDPath := filepath.Join(workspaceDir, "CLAUDE.md")
	originalContent := "# CLAUDE.md\n\nOriginal content - do not modify.\n"
	if err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create CLAUDE.md: %v", err)
	}

	t.Logf("Created workspace at %s", workspaceDir)
	t.Logf("Created agent DB dir at %s", agentDBDir)
	return workspaceDir, agentDBDir
}

// startPostgres starts a PostgreSQL container
func startPostgres(t *testing.T, ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, string, string) {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "edictflow",
			"POSTGRES_PASSWORD": "edictflow",
			"POSTGRES_DB":       "edictflow",
		},
		Networks:       []string{net.Name},
		NetworkAliases: map[string][]string{net.Name: {"postgres"}},
		WaitingFor:     wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get postgres host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get postgres mapped port: %v", err)
	}

	t.Logf("PostgreSQL running at %s:%s", host, mappedPort.Port())
	return container, host, mappedPort.Port()
}

// startServer builds and starts the server container
func startServer(t *testing.T, ctx context.Context, net *testcontainers.DockerNetwork, postgresHost string) (testcontainers.Container, string, string) {
	t.Helper()

	// Build from Dockerfile in ../server
	serverDir := filepath.Join("..", "server")

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    serverDir,
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_URL": "postgres://edictflow:edictflow@postgres:5432/edictflow?sslmode=disable",
			"JWT_SECRET":   "e2e-test-secret-do-not-use-in-production",
			"SERVER_PORT":  "8080",
		},
		Networks:       []string{net.Name},
		NetworkAliases: map[string][]string{net.Name: {"server"}},
		WaitingFor:     wait.ForHTTP("/health").WithPort("8080").WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start server container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get server host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("Failed to get server mapped port: %v", err)
	}

	hostURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	// For containers to reach the server, use host.docker.internal with the mapped port
	// This works when TESTCONTAINERS_HOST_OVERRIDE is set
	containerAccessibleURL := fmt.Sprintf("http://host.docker.internal:%s", mappedPort.Port())

	t.Logf("Server running at %s (container accessible: %s)", hostURL, containerAccessibleURL)
	return container, hostURL, containerAccessibleURL
}

// startAgent starts the agent container with bind mounts
func startAgent(t *testing.T, ctx context.Context, net *testcontainers.DockerNetwork, serverURL, workspaceDir, agentDBDir, agentBinaryPath string) testcontainers.Container {
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
		WaitingFor: wait.ForLog("Daemon running").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start agent container: %v", err)
	}

	t.Log("Agent container started")
	return container
}
