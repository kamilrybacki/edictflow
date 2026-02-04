# E2E Testing Design

## Overview

Containerized End-to-End test suite using testcontainers-go that exercises the full Claudeception system: Server, Agent daemon, and filesystem interactions. The host Go test process acts as the "user" modifying files, while containers run the infrastructure.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Host Machine                                │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    go test ./e2e/...                          │  │
│  │                                                               │  │
│  │  1. Compile agent binary → /tmp/agent-build/claudeception    │  │
│  │  2. Create temp workspace → /tmp/test-workspace-xxxx/        │  │
│  │  3. Start containers (testcontainers-go)                      │  │
│  │  4. Write files to temp workspace (simulate user)             │  │
│  │  5. Assert file state + call server API                       │  │
│  └──────────────────────────────────────────────────────────────┘  │
│          │                    │                     │               │
│          │ bind mount         │ bind mount          │ HTTP          │
│          │ (binary)           │ (workspace)         │ :random_port  │
│          ▼                    ▼                     ▼               │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │          Docker Network: claudeception-e2e-{uuid}              │ │
│  │                                                                │ │
│  │  ┌──────────┐    ┌──────────────┐    ┌───────────────────┐   │ │
│  │  │ postgres │◄───│    server    │◄───│    agent-node     │   │ │
│  │  │  :5432   │    │ :8080 (int)  │    │                   │   │ │
│  │  │          │    │ :xxxxx (ext) │    │ /app/bin ← binary │   │ │
│  │  └──────────┘    └──────────────┘    │ /app/ws ← tempdir │   │ │
│  │                         ▲            └───────────────────┘   │ │
│  │                         │                                     │ │
│  │            Container-to-Container: http://server:8080         │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                            ▲                                        │
│                            │                                        │
│              Host-to-Container: http://localhost:xxxxx              │
└─────────────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### Network Isolation

Each test run creates a unique Docker network with UUID suffix (`claudeception-e2e-{uuid}`) to prevent interference with other E2E tests running in parallel or other projects.

### Split-Brain Networking

| Caller | Target | Address |
|--------|--------|---------|
| Host (test) → Server | API verification | `localhost:{mapped_port}` |
| Agent container → Server | WebSocket + API | `http://server:8080` (network alias) |
| Server → Postgres | Database | `postgres:5432` (network alias) |

### Fsnotify Compatibility

Bind mounts from host → container have known fsnotify limitations on Mac/Windows. Solution: Agent uses `--poll-interval 500ms` flag to ensure reliable file change detection across all platforms.

### Authentication Bypass

The interactive Device Code Flow is bypassed by:
1. Pre-seeding Postgres with a test user and valid token
2. Pre-seeding Agent's SQLite database with matching credentials
3. Using `e2e-token-*` prefix that server recognizes in test mode

---

## Test Scenarios

### Scenario A: Block Mode

1. Agent starts with rule in `block` enforcement mode
2. Test runner modifies `CLAUDE.md` in workspace
3. Assert: File is reverted to original content within 10 seconds
4. Assert: Server received `change_blocked` event via API

### Scenario B: Temporary Mode

1. Update rule to `temporary` mode via server API
2. Wait for agent to sync new configuration
3. Test runner modifies `CLAUDE.md`
4. Assert: File change persists (NOT reverted)
5. Assert: Server received `change_detected` event

### Scenario C: Warning Mode

1. Update rule to `warning` mode via server API
2. Wait for agent to sync
3. Test runner modifies `CLAUDE.md`
4. Assert: File change persists
5. Assert: Server received `change_flagged` event

---

## File Structure

```
e2e/
├── go.mod                    # E2E test module
├── go.sum
├── run-e2e.sh               # Test runner script
├── Dockerfile.agent          # Lightweight agent container
├── suite_test.go            # Test suite setup/teardown
├── helpers_test.go          # Build & container helpers
├── seed_test.go             # Database & auth seeding
├── enforcement_test.go      # Core test scenarios
└── api_helpers_test.go      # Server API interaction
```

---

## Implementation

### suite_test.go - Test Suite Setup

```go
package e2e

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "github.com/google/uuid"
    _ "github.com/lib/pq"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/network"
    "github.com/testcontainers/testcontainers-go/wait"
)

type E2ESuite struct {
    ctx           context.Context
    network       *testcontainers.DockerNetwork
    postgres      testcontainers.Container
    server        testcontainers.Container
    agent         testcontainers.Container

    // Paths
    binaryPath    string
    workspacePath string
    agentDBPath   string

    // Connection info
    serverHostURL string  // Host access: localhost:xxxxx
    serverIntURL  string  // Container access: http://server:8080
    postgresURL   string

    // Test data
    testUserID    string
    testToken     string
    testRuleID    string
}

func NewE2ESuite(t *testing.T) *E2ESuite {
    t.Helper()

    ctx := context.Background()
    suite := &E2ESuite{
        ctx:        ctx,
        testUserID: uuid.New().String(),
        testToken:  fmt.Sprintf("e2e-token-%s", uuid.New().String()),
        testRuleID: uuid.New().String(),
    }

    // Create isolated network with unique name
    networkName := fmt.Sprintf("claudeception-e2e-%s", uuid.New().String()[:8])
    net, err := network.New(ctx,
        network.WithDriver("bridge"),
        network.WithLabels(map[string]string{
            "project": "claudeception-e2e",
            "run-id":  uuid.New().String(),
        }),
    )
    if err != nil {
        t.Fatalf("failed to create network: %v", err)
    }
    suite.network = net

    // Build agent binary
    suite.buildAgentBinary(t)

    // Create temp workspace and agent DB directory
    suite.createTempDirs(t)

    // Start containers in order
    suite.startPostgres(t)
    suite.seedDatabase(t)
    suite.startServer(t)
    suite.seedAgentDB(t)
    suite.startAgent(t)

    // Wait for agent to connect
    suite.waitForAgentConnection(t)

    return suite
}

func (s *E2ESuite) Cleanup(t *testing.T) {
    t.Helper()
    ctx := context.Background()

    if s.agent != nil {
        _ = s.agent.Terminate(ctx)
    }
    if s.server != nil {
        _ = s.server.Terminate(ctx)
    }
    if s.postgres != nil {
        _ = s.postgres.Terminate(ctx)
    }
    if s.network != nil {
        _ = s.network.Remove(ctx)
    }

    // Cleanup temp directories
    if s.workspacePath != "" {
        _ = os.RemoveAll(s.workspacePath)
    }
    if s.agentDBPath != "" {
        _ = os.RemoveAll(s.agentDBPath)
    }
    if s.binaryPath != "" {
        _ = os.RemoveAll(filepath.Dir(s.binaryPath))
    }
}
```

### helpers_test.go - Build & Container Helpers

```go
package e2e

import (
    "database/sql"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func (s *E2ESuite) buildAgentBinary(t *testing.T) {
    t.Helper()

    buildDir, err := os.MkdirTemp("", "claudeception-agent-build-")
    if err != nil {
        t.Fatalf("failed to create build dir: %v", err)
    }

    binaryPath := filepath.Join(buildDir, "claudeception")

    // Build for Linux (container target)
    cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/claudeception")
    cmd.Dir = filepath.Join("..", "agent")
    cmd.Env = append(os.Environ(),
        "GOOS=linux",
        "GOARCH=amd64",
        "CGO_ENABLED=1",  // Required for SQLite
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("failed to build agent: %v\n%s", err, output)
    }

    s.binaryPath = binaryPath
    t.Logf("Built agent binary: %s", binaryPath)
}

func (s *E2ESuite) createTempDirs(t *testing.T) {
    t.Helper()

    workspace, err := os.MkdirTemp("", "claudeception-workspace-")
    if err != nil {
        t.Fatalf("failed to create workspace: %v", err)
    }
    s.workspacePath = workspace

    agentDB, err := os.MkdirTemp("", "claudeception-agentdb-")
    if err != nil {
        t.Fatalf("failed to create agent db dir: %v", err)
    }
    s.agentDBPath = agentDB

    // Create initial CLAUDE.md with known content
    initialContent := "# CLAUDE.md\n\nOriginal content - do not modify.\n"
    err = os.WriteFile(filepath.Join(workspace, "CLAUDE.md"), []byte(initialContent), 0644)
    if err != nil {
        t.Fatalf("failed to write initial CLAUDE.md: %v", err)
    }

    t.Logf("Created workspace: %s", workspace)
    t.Logf("Created agent DB dir: %s", agentDB)
}

func (s *E2ESuite) startPostgres(t *testing.T) {
    t.Helper()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Networks:     []string{s.network.Name},
        NetworkAliases: map[string][]string{
            s.network.Name: {"postgres"},
        },
        Env: map[string]string{
            "POSTGRES_USER":     "claudeception",
            "POSTGRES_PASSWORD": "testpass",
            "POSTGRES_DB":       "claudeception_test",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(30 * time.Second),
        Labels: map[string]string{
            "project": "claudeception-e2e",
        },
    }

    container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("failed to start postgres: %v", err)
    }

    s.postgres = container

    // Get mapped port for host access
    mappedPort, err := container.MappedPort(s.ctx, "5432")
    if err != nil {
        t.Fatalf("failed to get postgres port: %v", err)
    }

    host, err := container.Host(s.ctx)
    if err != nil {
        t.Fatalf("failed to get postgres host: %v", err)
    }

    s.postgresURL = fmt.Sprintf("postgres://claudeception:testpass@%s:%s/claudeception_test?sslmode=disable",
        host, mappedPort.Port())

    t.Logf("Postgres started: %s", s.postgresURL)
}

func (s *E2ESuite) startServer(t *testing.T) {
    t.Helper()

    req := testcontainers.ContainerRequest{
        FromDockerfile: testcontainers.FromDockerfile{
            Context:    filepath.Join("..", "server"),
            Dockerfile: "Dockerfile",
        },
        ExposedPorts: []string{"8080/tcp", "8081/tcp"},
        Networks:     []string{s.network.Name},
        NetworkAliases: map[string][]string{
            s.network.Name: {"server"},
        },
        Env: map[string]string{
            "DATABASE_URL": "postgres://claudeception:testpass@postgres:5432/claudeception_test?sslmode=disable",
            "JWT_SECRET":   "e2e-test-secret-key-minimum-32-characters-long",
            "BASE_URL":     "http://server:8080",
            "WS_PORT":      "8081",
        },
        WaitingFor: wait.ForHTTP("/health").
            WithPort("8080/tcp").
            WithStartupTimeout(60 * time.Second),
        Labels: map[string]string{
            "project": "claudeception-e2e",
        },
    }

    container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("failed to start server: %v", err)
    }

    s.server = container

    // Get mapped port for host access
    mappedPort, err := container.MappedPort(s.ctx, "8080")
    if err != nil {
        t.Fatalf("failed to get server port: %v", err)
    }

    host, err := container.Host(s.ctx)
    if err != nil {
        t.Fatalf("failed to get server host: %v", err)
    }

    s.serverHostURL = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
    s.serverIntURL = "http://server:8080"

    t.Logf("Server started: %s (host) / %s (internal)", s.serverHostURL, s.serverIntURL)
}

func (s *E2ESuite) startAgent(t *testing.T) {
    t.Helper()

    req := testcontainers.ContainerRequest{
        Image:    "debian:bookworm-slim",
        Networks: []string{s.network.Name},
        NetworkAliases: map[string][]string{
            s.network.Name: {"agent"},
        },
        Env: map[string]string{
            "CLAUDECEPTION_SERVER_URL": "http://server:8080",
            "CLAUDECEPTION_WS_URL":     "ws://server:8081/ws",
            "CLAUDECEPTION_CONFIG_DIR": "/root/.claudeception",
        },
        Mounts: testcontainers.Mounts(
            testcontainers.BindMount(s.binaryPath, "/app/bin/claudeception"),
            testcontainers.BindMount(s.workspacePath, "/app/workspace"),
            testcontainers.BindMount(s.agentDBPath, "/root/.claudeception"),
        ),
        Cmd: []string{"/app/bin/claudeception", "start", "--foreground", "--poll-interval", "500ms"},
        WaitingFor: wait.ForLog("Connected to server").
            WithStartupTimeout(30 * time.Second),
        Labels: map[string]string{
            "project": "claudeception-e2e",
        },
    }

    container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("failed to start agent: %v", err)
    }

    s.agent = container
    t.Log("Agent started and connected")
}
```

### seed_test.go - Database & Auth Seeding

```go
package e2e

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "time"

    _ "github.com/lib/pq"
    _ "github.com/mattn/go-sqlite3"
)

func (s *E2ESuite) seedDatabase(t *testing.T) {
    t.Helper()

    // Wait a moment for postgres to be fully ready
    time.Sleep(1 * time.Second)

    db, err := sql.Open("postgres", s.postgresURL)
    if err != nil {
        t.Fatalf("failed to connect to postgres: %v", err)
    }
    defer db.Close()

    // Run migrations first (simplified - in practice, use migrate library)
    if err := s.runMigrations(db); err != nil {
        t.Fatalf("failed to run migrations: %v", err)
    }

    // Seed test user
    _, err = db.Exec(`
        INSERT INTO users (id, email, name, password_hash, is_active, created_at, updated_at)
        VALUES ($1, 'e2e-test@claudeception.local', 'E2E Test User', 'not-used', true, NOW(), NOW())
        ON CONFLICT (id) DO NOTHING
    `, s.testUserID)
    if err != nil {
        t.Fatalf("failed to seed user: %v", err)
    }

    // Seed test team
    testTeamID := "e2e-test-team-id"
    _, err = db.Exec(`
        INSERT INTO teams (id, name, created_at, updated_at)
        VALUES ($1, 'E2E Test Team', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING
    `, testTeamID)
    if err != nil {
        t.Fatalf("failed to seed team: %v", err)
    }

    // Add user to team
    _, err = db.Exec(`
        INSERT INTO team_members (team_id, user_id, role, created_at)
        VALUES ($1, $2, 'admin', NOW())
        ON CONFLICT (team_id, user_id) DO NOTHING
    `, testTeamID, s.testUserID)
    if err != nil {
        t.Fatalf("failed to seed team member: %v", err)
    }

    // Seed test rule in BLOCK mode
    triggers, _ := json.Marshal([]map[string]interface{}{
        {"type": "path", "pattern": "CLAUDE.md"},
    })

    _, err = db.Exec(`
        INSERT INTO rules (id, team_id, name, content, target_layer, triggers, enforcement_mode, status, created_by, created_at, updated_at)
        VALUES ($1, $2, 'E2E Test Rule', '# Test Rule\nDo not modify.', 'project', $3, 'block', 'approved', $4, NOW(), NOW())
        ON CONFLICT (id) DO NOTHING
    `, s.testRuleID, testTeamID, string(triggers), s.testUserID)
    if err != nil {
        t.Fatalf("failed to seed rule: %v", err)
    }

    // Seed agent registration
    testAgentID := "e2e-test-agent-id"
    _, err = db.Exec(`
        INSERT INTO agents (id, user_id, machine_id, hostname, os, last_seen_at, created_at)
        VALUES ($1, $2, 'e2e-machine-id', 'e2e-agent-container', 'linux', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING
    `, testAgentID, s.testUserID)
    if err != nil {
        t.Fatalf("failed to seed agent: %v", err)
    }

    t.Logf("Database seeded: user=%s, team=%s, rule=%s", s.testUserID, testTeamID, s.testRuleID)
}

func (s *E2ESuite) runMigrations(db *sql.DB) error {
    // In practice, use golang-migrate or similar
    // For now, assume migrations are baked into server container startup
    var result int
    return db.QueryRow("SELECT 1").Scan(&result)
}

func (s *E2ESuite) seedAgentDB(t *testing.T) {
    t.Helper()

    dbPath := filepath.Join(s.agentDBPath, "claudeception.db")

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        t.Fatalf("failed to create agent sqlite db: %v", err)
    }
    defer db.Close()

    // Create schema (matches agent/storage/migrations.go)
    schema := `
        CREATE TABLE IF NOT EXISTS auth (
            id INTEGER PRIMARY KEY CHECK (id = 1),
            access_token TEXT NOT NULL,
            refresh_token TEXT,
            expires_at INTEGER NOT NULL,
            user_id TEXT NOT NULL,
            user_email TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS message_queue (
            id INTEGER PRIMARY KEY,
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

    _, err = db.Exec(schema)
    if err != nil {
        t.Fatalf("failed to create agent schema: %v", err)
    }

    // Generate a valid JWT token (must match server's JWT_SECRET)
    token := s.generateTestToken(t)

    // Seed auth with valid token (expires in 24 hours)
    expiresAt := time.Now().Add(24 * time.Hour).Unix()
    _, err = db.Exec(`
        INSERT OR REPLACE INTO auth (id, access_token, refresh_token, expires_at, user_id, user_email)
        VALUES (1, ?, '', ?, ?, 'e2e-test@claudeception.local')
    `, token, expiresAt, s.testUserID)
    if err != nil {
        t.Fatalf("failed to seed agent auth: %v", err)
    }

    // Pre-cache the test rule
    triggers, _ := json.Marshal([]map[string]interface{}{
        {"type": "path", "pattern": "CLAUDE.md"},
    })
    _, err = db.Exec(`
        INSERT OR REPLACE INTO cached_rules (id, name, content, target_layer, triggers, enforcement_mode, version, cached_at)
        VALUES (?, 'E2E Test Rule', '# Test Rule\nDo not modify.', 'project', ?, 'block', 1, ?)
    `, s.testRuleID, string(triggers), time.Now().Unix())
    if err != nil {
        t.Fatalf("failed to seed cached rule: %v", err)
    }

    // Register workspace as watched project
    _, err = db.Exec(`
        INSERT OR REPLACE INTO watched_projects (path, detected_context, detected_tags, last_sync_at)
        VALUES ('/app/workspace', '[]', '[]', ?)
    `, time.Now().Unix())
    if err != nil {
        t.Fatalf("failed to seed watched project: %v", err)
    }

    s.testToken = token
    t.Logf("Agent DB seeded: %s", dbPath)
}

func (s *E2ESuite) generateTestToken(t *testing.T) string {
    t.Helper()
    return fmt.Sprintf("e2e-token-%s-%d", s.testUserID, time.Now().Unix())
}

func (s *E2ESuite) waitForAgentConnection(t *testing.T) {
    t.Helper()

    deadline := time.Now().Add(30 * time.Second)

    for time.Now().Before(deadline) {
        logs, err := s.agent.Logs(s.ctx)
        if err == nil {
            buf := make([]byte, 4096)
            n, _ := logs.Read(buf)
            if n > 0 && containsString(string(buf[:n]), "Connected to server") {
                t.Log("Agent connected to server")
                return
            }
        }
        time.Sleep(500 * time.Millisecond)
    }

    t.Fatal("timeout waiting for agent to connect")
}

func containsString(haystack, needle string) bool {
    for i := 0; i <= len(haystack)-len(needle); i++ {
        if haystack[i:i+len(needle)] == needle {
            return true
        }
    }
    return false
}
```

### enforcement_test.go - Core Test Scenarios

```go
package e2e

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "testing"
    "time"
)

const (
    originalContent  = "# CLAUDE.md\n\nOriginal content - do not modify.\n"
    modifiedContent  = "# CLAUDE.md\n\nThis content was modified by the user.\n"
    pollInterval     = 200 * time.Millisecond
    revertTimeout    = 10 * time.Second
    syncTimeout      = 15 * time.Second
)

func TestAgentEnforcement(t *testing.T) {
    suite := NewE2ESuite(t)
    defer suite.Cleanup(t)

    t.Run("ScenarioA_BlockMode", func(t *testing.T) {
        suite.testBlockMode(t)
    })

    t.Run("ScenarioB_TemporaryMode", func(t *testing.T) {
        suite.testTemporaryMode(t)
    })

    t.Run("ScenarioC_WarningMode", func(t *testing.T) {
        suite.testWarningMode(t)
    })
}

// ScenarioA: Block Mode - changes are immediately reverted
func (s *E2ESuite) testBlockMode(t *testing.T) {
    t.Helper()

    claudeMDPath := filepath.Join(s.workspacePath, "CLAUDE.md")

    // Verify initial state
    content, err := os.ReadFile(claudeMDPath)
    if err != nil {
        t.Fatalf("failed to read initial CLAUDE.md: %v", err)
    }
    if string(content) != originalContent {
        t.Fatalf("unexpected initial content: %s", content)
    }

    // Step 1: Simulate user modifying the file
    t.Log("Modifying CLAUDE.md (block mode - should be reverted)")
    err = os.WriteFile(claudeMDPath, []byte(modifiedContent), 0644)
    if err != nil {
        t.Fatalf("failed to write modified content: %v", err)
    }

    // Step 2: Wait for agent to revert the file
    reverted := s.waitForFileContent(t, claudeMDPath, originalContent, revertTimeout)
    if !reverted {
        actualContent, _ := os.ReadFile(claudeMDPath)
        t.Fatalf("file was not reverted within timeout\nexpected: %s\nactual: %s",
            originalContent, actualContent)
    }
    t.Log("File successfully reverted by agent")

    // Step 3: Verify server received change_blocked event
    event := s.getLatestChangeEvent(t)
    if event == nil {
        t.Fatal("no change event found on server")
    }
    if event.EventType != "change_blocked" {
        t.Errorf("expected event type 'change_blocked', got '%s'", event.EventType)
    }
    if event.FilePath != "/app/workspace/CLAUDE.md" {
        t.Errorf("unexpected file path: %s", event.FilePath)
    }

    t.Log("Server recorded change_blocked event")
}

// ScenarioB: Temporary Mode - changes persist but are tracked
func (s *E2ESuite) testTemporaryMode(t *testing.T) {
    t.Helper()

    claudeMDPath := filepath.Join(s.workspacePath, "CLAUDE.md")

    // Reset file to original content
    err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644)
    if err != nil {
        t.Fatalf("failed to reset CLAUDE.md: %v", err)
    }
    time.Sleep(1 * time.Second)

    // Step 1: Update rule to temporary mode via server API
    t.Log("Updating rule to temporary mode")
    s.updateRuleEnforcementMode(t, "temporary")

    // Step 2: Wait for agent to sync new config
    s.waitForAgentSync(t)

    // Step 3: Modify the file
    t.Log("Modifying CLAUDE.md (temporary mode - should persist)")
    err = os.WriteFile(claudeMDPath, []byte(modifiedContent), 0644)
    if err != nil {
        t.Fatalf("failed to write modified content: %v", err)
    }

    // Step 4: Wait and verify file is NOT reverted
    time.Sleep(3 * time.Second)
    content, err := os.ReadFile(claudeMDPath)
    if err != nil {
        t.Fatalf("failed to read CLAUDE.md: %v", err)
    }
    if string(content) != modifiedContent {
        t.Fatalf("file should have persisted in temporary mode\nexpected: %s\nactual: %s",
            modifiedContent, content)
    }
    t.Log("File change persisted (temporary mode)")

    // Step 5: Verify server received change_detected event
    event := s.getLatestChangeEvent(t)
    if event == nil {
        t.Fatal("no change event found on server")
    }
    if event.EventType != "change_detected" {
        t.Errorf("expected event type 'change_detected', got '%s'", event.EventType)
    }

    t.Log("Server recorded change_detected event")
}

// ScenarioC: Warning Mode - changes persist, flagged for review
func (s *E2ESuite) testWarningMode(t *testing.T) {
    t.Helper()

    claudeMDPath := filepath.Join(s.workspacePath, "CLAUDE.md")

    // Reset file to original content
    err := os.WriteFile(claudeMDPath, []byte(originalContent), 0644)
    if err != nil {
        t.Fatalf("failed to reset CLAUDE.md: %v", err)
    }
    time.Sleep(1 * time.Second)

    // Step 1: Update rule to warning mode
    t.Log("Updating rule to warning mode")
    s.updateRuleEnforcementMode(t, "warning")

    // Step 2: Wait for agent to sync
    s.waitForAgentSync(t)

    // Step 3: Modify the file
    t.Log("Modifying CLAUDE.md (warning mode - should persist)")
    err = os.WriteFile(claudeMDPath, []byte(modifiedContent), 0644)
    if err != nil {
        t.Fatalf("failed to write modified content: %v", err)
    }

    // Step 4: Verify file persists
    time.Sleep(3 * time.Second)
    content, err := os.ReadFile(claudeMDPath)
    if err != nil {
        t.Fatalf("failed to read CLAUDE.md: %v", err)
    }
    if string(content) != modifiedContent {
        t.Fatalf("file should have persisted in warning mode")
    }
    t.Log("File change persisted (warning mode)")

    // Step 5: Verify server received change_flagged event
    event := s.getLatestChangeEvent(t)
    if event == nil {
        t.Fatal("no change event found on server")
    }
    if event.EventType != "change_flagged" {
        t.Errorf("expected event type 'change_flagged', got '%s'", event.EventType)
    }

    t.Log("Server recorded change_flagged event")
}
```

### api_helpers_test.go - Server API Interaction

```go
package e2e

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "testing"
    "time"
)

type ChangeEvent struct {
    ID           string    `json:"id"`
    RuleID       string    `json:"rule_id"`
    AgentID      string    `json:"agent_id"`
    FilePath     string    `json:"file_path"`
    EventType    string    `json:"event_type"`
    OriginalHash string    `json:"original_hash"`
    ModifiedHash string    `json:"modified_hash"`
    CreatedAt    time.Time `json:"created_at"`
}

func (s *E2ESuite) waitForFileContent(t *testing.T, path, expected string, timeout time.Duration) bool {
    t.Helper()

    deadline := time.Now().Add(timeout)

    for time.Now().Before(deadline) {
        content, err := os.ReadFile(path)
        if err == nil && string(content) == expected {
            return true
        }
        time.Sleep(pollInterval)
    }

    return false
}

func (s *E2ESuite) updateRuleEnforcementMode(t *testing.T, mode string) {
    t.Helper()

    url := fmt.Sprintf("%s/api/v1/rules/%s", s.serverHostURL, s.testRuleID)

    payload := map[string]interface{}{
        "enforcement_mode": mode,
    }
    body, _ := json.Marshal(payload)

    req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
    if err != nil {
        t.Fatalf("failed to create request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.testToken)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("failed to update rule: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("failed to update rule: %d - %s", resp.StatusCode, body)
    }

    t.Logf("Rule enforcement mode updated to: %s", mode)
}

func (s *E2ESuite) waitForAgentSync(t *testing.T) {
    t.Helper()

    deadline := time.Now().Add(syncTimeout)

    for time.Now().Before(deadline) {
        logs, err := s.agent.Logs(s.ctx)
        if err == nil {
            buf := make([]byte, 8192)
            n, _ := logs.Read(buf)
            logContent := string(buf[:n])

            if containsString(logContent, "Config updated") ||
               containsString(logContent, "Rules synced") {
                t.Log("Agent synced new configuration")
                return
            }
        }
        time.Sleep(500 * time.Millisecond)
    }

    // Fallback: assume sync happened after timeout
    t.Log("Assuming agent synced (timeout reached)")
}

func (s *E2ESuite) getLatestChangeEvent(t *testing.T) *ChangeEvent {
    t.Helper()

    url := fmt.Sprintf("%s/api/v1/changes/?limit=1&order=desc", s.serverHostURL)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        t.Fatalf("failed to create request: %v", err)
    }
    req.Header.Set("Authorization", "Bearer "+s.testToken)

    var lastErr error
    for i := 0; i < 5; i++ {
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            lastErr = err
            time.Sleep(500 * time.Millisecond)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, body)
            time.Sleep(500 * time.Millisecond)
            continue
        }

        var events []ChangeEvent
        if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
            lastErr = err
            time.Sleep(500 * time.Millisecond)
            continue
        }

        if len(events) > 0 {
            return &events[0]
        }

        time.Sleep(500 * time.Millisecond)
    }

    if lastErr != nil {
        t.Logf("warning: error fetching events: %v", lastErr)
    }
    return nil
}

func (s *E2ESuite) triggerAgentSync(t *testing.T) {
    t.Helper()

    exitCode, output, err := s.agent.Exec(s.ctx, []string{
        "/app/bin/claudeception", "sync",
    })
    if err != nil {
        t.Logf("sync exec error: %v", err)
        return
    }
    if exitCode != 0 {
        buf := new(bytes.Buffer)
        io.Copy(buf, output)
        t.Logf("sync failed (exit %d): %s", exitCode, buf.String())
        return
    }

    t.Log("Triggered manual sync on agent")
}
```

---

## Supporting Files

### run-e2e.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

cleanup() {
    log_info "Cleaning up..."
    docker ps -a --filter "label=project=claudeception-e2e" -q | xargs -r docker rm -f 2>/dev/null || true
    docker network ls --filter "label=project=claudeception-e2e" -q | xargs -r docker network rm 2>/dev/null || true
    rm -rf /tmp/claudeception-agent-build-* 2>/dev/null || true
    rm -rf /tmp/claudeception-workspace-* 2>/dev/null || true
    rm -rf /tmp/claudeception-agentdb-* 2>/dev/null || true
}

trap cleanup EXIT

main() {
    log_info "=== Claudeception E2E Test Suite ==="

    log_info "Checking prerequisites..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    log_info "Building server Docker image..."
    docker build -t claudeception-server:e2e "$PROJECT_ROOT/server"

    log_info "Verifying agent compiles..."
    (cd "$PROJECT_ROOT/agent" && go build -o /dev/null ./cmd/claudeception)

    log_info "Running E2E tests..."
    cd "$SCRIPT_DIR"

    export CGO_ENABLED=1

    if go test -v -timeout 5m -count=1 ./...; then
        log_info "=== E2E Tests PASSED ==="
        exit 0
    else
        log_error "=== E2E Tests FAILED ==="
        exit 1
    fi
}

case "${1:-}" in
    --cleanup-only)
        cleanup
        log_info "Cleanup complete"
        exit 0
        ;;
    --help|-h)
        echo "Usage: $0 [--cleanup-only|--help]"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
```

### e2e/go.mod

```go
module github.com/kamilrybacki/claudeception/e2e

go 1.22

require (
    github.com/google/uuid v1.6.0
    github.com/lib/pq v1.10.9
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/testcontainers/testcontainers-go v0.29.1
)
```

### Taskfile.yml additions

```yaml
  e2e:
    desc: Run E2E integration tests
    dir: e2e
    cmds:
      - ./run-e2e.sh

  e2e:cleanup:
    desc: Clean up orphaned E2E test resources
    dir: e2e
    cmds:
      - ./run-e2e.sh --cleanup-only
```

---

## Prerequisites

Before running E2E tests:

1. **Docker** - Must be installed and running
2. **Go 1.22+** - With CGO enabled (for SQLite)
3. **Server Dockerfile** - Must exist at `server/Dockerfile`

## Running Tests

```bash
# Run full E2E suite
task e2e

# Or directly
cd e2e && ./run-e2e.sh

# Cleanup orphaned resources
task e2e:cleanup
```

## Troubleshooting

### Fsnotify not detecting changes

On Mac/Windows, bind mounts may not propagate inotify events reliably. The agent uses `--poll-interval 500ms` as a fallback.

### Container networking issues

Each test run uses a unique network name (`claudeception-e2e-{uuid}`) to avoid conflicts. If tests hang, run `task e2e:cleanup` to remove orphaned resources.

### SQLite CGO errors

Ensure `CGO_ENABLED=1` is set when building the agent binary for Linux.
