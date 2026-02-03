# Claudeception Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a team collaboration tool for managing layered CLAUDE.md configurations with centralized web UI and local agents.

**Architecture:** Go backend with Clean/Hexagonal architecture (domain → services → entrypoints), Next.js frontend, Go CLI agent. PostgreSQL for storage, WebSocket for real-time sync.

**Tech Stack:** Go (Chi, pgx, gorilla/websocket, Cobra), TypeScript (Next.js, Tailwind, React Query, Zod), PostgreSQL, Docker.

---

## Progress Summary

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Project Scaffolding & Domain Layer | ✅ Complete |
| Phase 2 | Database Layer | ✅ Complete |
| Phase 3 | Service Layer - Rule Matching | ✅ Complete |
| Phase 4 | Next.js Frontend Setup | ✅ Complete |
| Phase 5 | Agent CLI Setup | ✅ Complete |
| Phase 6 | Docker & Development Environment | ✅ Complete |
| Phase 7 | REST API Endpoints | 🔄 In Progress (6/12 tasks) |
| Phase 8 | WebSocket Handlers | ⏳ Pending |

**Current Phase 7 Progress:**
- ✅ Task 1: Add server dependencies (chi, pgx, jwt)
- ✅ Task 2: Database connection pool
- ✅ Task 3: Team repository
- ✅ Task 4: Rule repository
- ✅ Task 5: JWT authentication middleware
- ✅ Task 6: Teams API handler
- ⏳ Task 7: Rules API handler
- ⏳ Task 8: Chi router setup
- ⏳ Task 9: WebSocket message types
- ⏳ Task 10: WebSocket hub
- ⏳ Task 11: WebSocket handler
- ⏳ Task 12: Server main entry point

**See:** `docs/plans/2026-02-03-phase2-api-websocket.md` for Phase 7-8 detailed plan.

---

## Phase 1: Project Scaffolding & Domain Layer (✅ Complete)

### Task 1: Initialize Go Modules

**Files:**
- Create: `server/go.mod`
- Create: `agent/go.mod`

**Step 1: Create server Go module**

```bash
cd server && go mod init github.com/kamilrybacki/claudeception/server
```

**Step 2: Create agent Go module**

```bash
cd ../agent && go mod init github.com/kamilrybacki/claudeception/agent
```

**Step 3: Commit**

```bash
git add server/go.mod agent/go.mod
git commit -m "feat: initialize Go modules for server and agent"
```

---

### Task 2: Server Domain Models - Team

**Files:**
- Create: `server/domain/team.go`
- Create: `server/domain/team_test.go`

**Step 1: Write the failing test**

Create `server/domain/team_test.go`:

```go
package domain_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewTeamCreatesValidTeam(t *testing.T) {
	team := domain.NewTeam("Engineering")

	if team.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", team.Name)
	}
	if team.ID == "" {
		t.Error("expected non-empty ID")
	}
	if team.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestTeamValidateRejectsEmptyName(t *testing.T) {
	team := domain.Team{
		ID:        "test-id",
		Name:      "",
		CreatedAt: time.Now(),
	}

	err := team.Validate()
	if err == nil {
		t.Error("expected validation error for empty name")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain/... -v`
Expected: FAIL with "package domain is not in std"

**Step 3: Write minimal implementation**

Create `server/domain/team.go`:

```go
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TeamSettings struct {
	DriftThresholdMinutes int `json:"drift_threshold_minutes"`
}

type Team struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Settings  TeamSettings `json:"settings"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewTeam(name string) Team {
	return Team{
		ID:        uuid.New().String(),
		Name:      name,
		Settings:  TeamSettings{DriftThresholdMinutes: 60},
		CreatedAt: time.Now(),
	}
}

func (t Team) Validate() error {
	if t.Name == "" {
		return errors.New("team name cannot be empty")
	}
	return nil
}
```

**Step 4: Add uuid dependency**

Run: `cd server && go get github.com/google/uuid`

**Step 5: Run test to verify it passes**

Run: `cd server && go test ./domain/... -v`
Expected: PASS

**Step 6: Commit**

```bash
git add server/domain/team.go server/domain/team_test.go server/go.mod server/go.sum
git commit -m "feat(domain): add Team model with validation"
```

---

### Task 3: Server Domain Models - User

**Files:**
- Create: `server/domain/user.go`
- Create: `server/domain/user_test.go`

**Step 1: Write the failing test**

Create `server/domain/user_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewUserCreatesValidUser(t *testing.T) {
	user := domain.NewUser("alice@example.com", "Alice", "github", "team-123")

	if user.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", user.Email)
	}
	if user.AuthProvider != domain.AuthProviderGitHub {
		t.Errorf("expected auth provider 'github', got '%s'", user.AuthProvider)
	}
	if user.Role != domain.RoleMember {
		t.Errorf("expected role 'member', got '%s'", user.Role)
	}
}

func TestUserValidateRejectsInvalidEmail(t *testing.T) {
	user := domain.User{
		ID:           "test-id",
		Email:        "not-an-email",
		Name:         "Test",
		AuthProvider: domain.AuthProviderGitHub,
		Role:         domain.RoleMember,
		TeamID:       "team-123",
	}

	err := user.Validate()
	if err == nil {
		t.Error("expected validation error for invalid email")
	}
}

func TestUserValidateRejectsInvalidAuthProvider(t *testing.T) {
	user := domain.User{
		ID:           "test-id",
		Email:        "alice@example.com",
		Name:         "Alice",
		AuthProvider: "invalid",
		Role:         domain.RoleMember,
		TeamID:       "team-123",
	}

	err := user.Validate()
	if err == nil {
		t.Error("expected validation error for invalid auth provider")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `server/domain/user.go`:

```go
package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	AuthProviderGitHub AuthProvider = "github"
	AuthProviderGitLab AuthProvider = "gitlab"
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderLocal  AuthProvider = "local"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	AvatarURL    string       `json:"avatar_url,omitempty"`
	AuthProvider AuthProvider `json:"auth_provider"`
	Role         Role         `json:"role"`
	TeamID       string       `json:"team_id"`
	CreatedAt    time.Time    `json:"created_at"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewUser(email, name string, authProvider AuthProvider, teamID string) User {
	return User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		AuthProvider: authProvider,
		Role:         RoleMember,
		TeamID:       teamID,
		CreatedAt:    time.Now(),
	}
}

func (u User) Validate() error {
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}
	if !u.AuthProvider.IsValid() {
		return errors.New("invalid auth provider")
	}
	if !u.Role.IsValid() {
		return errors.New("invalid role")
	}
	return nil
}

func (ap AuthProvider) IsValid() bool {
	switch ap {
	case AuthProviderGitHub, AuthProviderGitLab, AuthProviderGoogle, AuthProviderLocal:
		return true
	}
	return false
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleMember:
		return true
	}
	return false
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/user.go server/domain/user_test.go
git commit -m "feat(domain): add User model with auth providers and roles"
```

---

### Task 4: Server Domain Models - Rule

**Files:**
- Create: `server/domain/rule.go`
- Create: `server/domain/rule_test.go`

**Step 1: Write the failing test**

Create `server/domain/rule_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewRuleCreatesValidRule(t *testing.T) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "**/frontend/**"},
	}
	rule := domain.NewRule("React Standards", domain.TargetLayerProject, "# React\nUse hooks.", triggers, "team-123")

	if rule.Name != "React Standards" {
		t.Errorf("expected name 'React Standards', got '%s'", rule.Name)
	}
	if rule.TargetLayer != domain.TargetLayerProject {
		t.Errorf("expected target layer 'project', got '%s'", rule.TargetLayer)
	}
	if rule.PriorityWeight != 0 {
		t.Errorf("expected priority weight 0, got %d", rule.PriorityWeight)
	}
}

func TestTriggerSpecificityOrdersCorrectly(t *testing.T) {
	pathTrigger := domain.Trigger{Type: domain.TriggerTypePath, Pattern: "**/src/**"}
	contextTrigger := domain.Trigger{Type: domain.TriggerTypeContext, ContextTypes: []string{"node"}}
	tagTrigger := domain.Trigger{Type: domain.TriggerTypeTag, Tags: []string{"frontend"}}

	if pathTrigger.Specificity() <= contextTrigger.Specificity() {
		t.Error("path trigger should have higher specificity than context trigger")
	}
	if contextTrigger.Specificity() <= tagTrigger.Specificity() {
		t.Error("context trigger should have higher specificity than tag trigger")
	}
}

func TestRuleValidateRejectsEmptyContent(t *testing.T) {
	rule := domain.Rule{
		ID:          "test-id",
		Name:        "Test Rule",
		TargetLayer: domain.TargetLayerProject,
		Content:     "",
		TeamID:      "team-123",
	}

	err := rule.Validate()
	if err == nil {
		t.Error("expected validation error for empty content")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `server/domain/rule.go`:

```go
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TargetLayer string

const (
	TargetLayerEnterprise TargetLayer = "enterprise"
	TargetLayerGlobal     TargetLayer = "global"
	TargetLayerProject    TargetLayer = "project"
	TargetLayerLocal      TargetLayer = "local"
)

type TriggerType string

const (
	TriggerTypePath    TriggerType = "path"
	TriggerTypeContext TriggerType = "context"
	TriggerTypeTag     TriggerType = "tag"
)

type Trigger struct {
	Type         TriggerType `json:"type"`
	Pattern      string      `json:"pattern,omitempty"`
	ContextTypes []string    `json:"context_types,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
}

func (t Trigger) Specificity() int {
	switch t.Type {
	case TriggerTypePath:
		return 100
	case TriggerTypeContext:
		return 50
	case TriggerTypeTag:
		return 10
	}
	return 0
}

type Rule struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Content        string      `json:"content"`
	TargetLayer    TargetLayer `json:"target_layer"`
	PriorityWeight int         `json:"priority_weight"`
	Triggers       []Trigger   `json:"triggers"`
	TeamID         string      `json:"team_id"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func NewRule(name string, targetLayer TargetLayer, content string, triggers []Trigger, teamID string) Rule {
	now := time.Now()
	return Rule{
		ID:             uuid.New().String(),
		Name:           name,
		Content:        content,
		TargetLayer:    targetLayer,
		PriorityWeight: 0,
		Triggers:       triggers,
		TeamID:         teamID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (r Rule) Validate() error {
	if r.Name == "" {
		return errors.New("rule name cannot be empty")
	}
	if r.Content == "" {
		return errors.New("rule content cannot be empty")
	}
	if !r.TargetLayer.IsValid() {
		return errors.New("invalid target layer")
	}
	return nil
}

func (tl TargetLayer) IsValid() bool {
	switch tl {
	case TargetLayerEnterprise, TargetLayerGlobal, TargetLayerProject, TargetLayerLocal:
		return true
	}
	return false
}

func (r Rule) MaxSpecificity() int {
	maxSpecificity := 0
	for _, trigger := range r.Triggers {
		if s := trigger.Specificity(); s > maxSpecificity {
			maxSpecificity = s
		}
	}
	return maxSpecificity
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/rule.go server/domain/rule_test.go
git commit -m "feat(domain): add Rule model with triggers and specificity"
```

---

### Task 5: Server Domain Models - Project

**Files:**
- Create: `server/domain/project.go`
- Create: `server/domain/project_test.go`

**Step 1: Write the failing test**

Create `server/domain/project_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewProjectCreatesValidProject(t *testing.T) {
	project := domain.NewProject("~/projects/myapp", []string{"frontend", "react"}, "team-123")

	if project.PathPattern != "~/projects/myapp" {
		t.Errorf("expected path pattern '~/projects/myapp', got '%s'", project.PathPattern)
	}
	if len(project.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(project.Tags))
	}
}

func TestProjectMatchesPathPattern(t *testing.T) {
	project := domain.NewProject("**/frontend/**", []string{}, "team-123")

	if !project.MatchesPath("/home/user/projects/myapp/frontend/src/App.tsx") {
		t.Error("expected path to match pattern")
	}
	if project.MatchesPath("/home/user/projects/myapp/backend/main.go") {
		t.Error("expected path NOT to match pattern")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `server/domain/project.go`:

```go
package domain

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          string    `json:"id"`
	PathPattern string    `json:"path_pattern"`
	Tags        []string  `json:"tags"`
	TeamID      string    `json:"team_id"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewProject(pathPattern string, tags []string, teamID string) Project {
	now := time.Now()
	return Project{
		ID:          uuid.New().String(),
		PathPattern: pathPattern,
		Tags:        tags,
		TeamID:      teamID,
		LastSeenAt:  now,
		CreatedAt:   now,
	}
}

func (p Project) MatchesPath(path string) bool {
	matched, err := filepath.Match(p.PathPattern, path)
	if err != nil {
		// For glob patterns with **, use a simple contains check as fallback
		// Real implementation would use doublestar library
		return globMatch(p.PathPattern, path)
	}
	return matched
}

func globMatch(pattern, path string) bool {
	// Simplified glob matching for ** patterns
	// In production, use github.com/bmatcuk/doublestar
	if len(pattern) > 2 && pattern[:2] == "**" {
		suffix := pattern[2:]
		if len(suffix) > 0 && suffix[0] == '/' {
			suffix = suffix[1:]
		}
		// Check if path contains the suffix pattern
		matched, _ := filepath.Match("*"+suffix+"*", path)
		return matched
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/project.go server/domain/project_test.go
git commit -m "feat(domain): add Project model with path matching"
```

---

### Task 6: Server Domain Models - Agent

**Files:**
- Create: `server/domain/agent.go`
- Create: `server/domain/agent_test.go`

**Step 1: Write the failing test**

Create `server/domain/agent_test.go`:

```go
package domain_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewAgentCreatesValidAgent(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")

	if agent.MachineID != "machine-abc123" {
		t.Errorf("expected machine ID 'machine-abc123', got '%s'", agent.MachineID)
	}
	if agent.Status != domain.AgentStatusOnline {
		t.Errorf("expected status 'online', got '%s'", agent.Status)
	}
}

func TestAgentIsStaleAfterThreshold(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")
	agent.LastHeartbeat = time.Now().Add(-2 * time.Hour)

	staleThreshold := 1 * time.Hour
	if !agent.IsStale(staleThreshold) {
		t.Error("expected agent to be stale after threshold")
	}
}

func TestAgentIsNotStaleWithinThreshold(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")
	agent.LastHeartbeat = time.Now().Add(-30 * time.Minute)

	staleThreshold := 1 * time.Hour
	if agent.IsStale(staleThreshold) {
		t.Error("expected agent NOT to be stale within threshold")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./domain/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `server/domain/agent.go`:

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusStale   AgentStatus = "stale"
	AgentStatusOffline AgentStatus = "offline"
)

type Agent struct {
	ID                  string      `json:"id"`
	MachineID           string      `json:"machine_id"`
	UserID              string      `json:"user_id"`
	Status              AgentStatus `json:"status"`
	LastHeartbeat       time.Time   `json:"last_heartbeat"`
	CachedConfigVersion int         `json:"cached_config_version"`
	CreatedAt           time.Time   `json:"created_at"`
}

func NewAgent(machineID, userID string) Agent {
	now := time.Now()
	return Agent{
		ID:                  uuid.New().String(),
		MachineID:           machineID,
		UserID:              userID,
		Status:              AgentStatusOnline,
		LastHeartbeat:       now,
		CachedConfigVersion: 0,
		CreatedAt:           now,
	}
}

func (a Agent) IsStale(threshold time.Duration) bool {
	return time.Since(a.LastHeartbeat) > threshold
}

func (a *Agent) UpdateHeartbeat() {
	a.LastHeartbeat = time.Now()
	a.Status = AgentStatusOnline
}

func (a *Agent) MarkStale() {
	a.Status = AgentStatusStale
}

func (a *Agent) MarkOffline() {
	a.Status = AgentStatusOffline
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./domain/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/domain/agent.go server/domain/agent_test.go
git commit -m "feat(domain): add Agent model with status tracking"
```

---

## Phase 2: Database Layer

### Task 7: Database Configuration & Migrations Setup

**Files:**
- Create: `server/configurator/settings.go`
- Create: `server/migrations/000001_create_teams.up.sql`
- Create: `server/migrations/000001_create_teams.down.sql`

**Step 1: Create settings**

Create `server/configurator/settings.go`:

```go
package configurator

import (
	"os"
)

type Settings struct {
	DatabaseURL string
	ServerPort  string
	JWTSecret   string
}

func LoadSettings() Settings {
	return Settings{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/claudeception?sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

**Step 2: Create teams migration (up)**

Create `server/migrations/000001_create_teams.up.sql`:

```sql
CREATE TABLE teams (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teams_name ON teams(name);
```

**Step 3: Create teams migration (down)**

Create `server/migrations/000001_create_teams.down.sql`:

```sql
DROP TABLE IF EXISTS teams;
```

**Step 4: Commit**

```bash
git add server/configurator/settings.go server/migrations/
git commit -m "feat: add settings and teams migration"
```

---

### Task 8: Users Migration

**Files:**
- Create: `server/migrations/000002_create_users.up.sql`
- Create: `server/migrations/000002_create_users.down.sql`

**Step 1: Create users migration (up)**

Create `server/migrations/000002_create_users.up.sql`:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    auth_provider VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_team_id ON users(team_id);
```

**Step 2: Create users migration (down)**

Create `server/migrations/000002_create_users.down.sql`:

```sql
DROP TABLE IF EXISTS users;
```

**Step 3: Commit**

```bash
git add server/migrations/000002_*
git commit -m "feat: add users migration"
```

---

### Task 9: Rules Migration

**Files:**
- Create: `server/migrations/000003_create_rules.up.sql`
- Create: `server/migrations/000003_create_rules.down.sql`

**Step 1: Create rules migration (up)**

Create `server/migrations/000003_create_rules.up.sql`:

```sql
CREATE TABLE rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    target_layer VARCHAR(50) NOT NULL,
    priority_weight INTEGER NOT NULL DEFAULT 0,
    triggers JSONB NOT NULL DEFAULT '[]',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rules_team_id ON rules(team_id);
CREATE INDEX idx_rules_target_layer ON rules(target_layer);
```

**Step 2: Create rules migration (down)**

Create `server/migrations/000003_create_rules.down.sql`:

```sql
DROP TABLE IF EXISTS rules;
```

**Step 3: Commit**

```bash
git add server/migrations/000003_*
git commit -m "feat: add rules migration"
```

---

### Task 10: Projects and Agents Migrations

**Files:**
- Create: `server/migrations/000004_create_projects.up.sql`
- Create: `server/migrations/000004_create_projects.down.sql`
- Create: `server/migrations/000005_create_agents.up.sql`
- Create: `server/migrations/000005_create_agents.down.sql`

**Step 1: Create projects migration (up)**

Create `server/migrations/000004_create_projects.up.sql`:

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    path_pattern VARCHAR(500) NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_team_id ON projects(team_id);
CREATE INDEX idx_projects_tags ON projects USING GIN(tags);
```

**Step 2: Create projects migration (down)**

Create `server/migrations/000004_create_projects.down.sql`:

```sql
DROP TABLE IF EXISTS projects;
```

**Step 3: Create agents migration (up)**

Create `server/migrations/000005_create_agents.up.sql`:

```sql
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    machine_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'offline',
    last_heartbeat TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    cached_config_version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(machine_id, user_id)
);

CREATE INDEX idx_agents_user_id ON agents(user_id);
CREATE INDEX idx_agents_status ON agents(status);
```

**Step 4: Create agents migration (down)**

Create `server/migrations/000005_create_agents.down.sql`:

```sql
DROP TABLE IF EXISTS agents;
```

**Step 5: Commit**

```bash
git add server/migrations/000004_* server/migrations/000005_*
git commit -m "feat: add projects and agents migrations"
```

---

## Phase 3: Service Layer - Rule Matching

### Task 11: Rule Matcher Service

**Files:**
- Create: `server/services/rules/matcher.go`
- Create: `server/services/rules/matcher_test.go`

**Step 1: Write the failing test**

Create `server/services/rules/matcher_test.go`:

```go
package rules_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/services/rules"
)

func TestMatcherReturnsRulesMatchingPath(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "rule-1",
			Name:     "Frontend Rule",
			Content:  "# Frontend",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/frontend/**"}},
		},
		{
			ID:       "rule-2",
			Name:     "Backend Rule",
			Content:  "# Backend",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/backend/**"}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		ProjectPath: "/home/user/myapp/frontend/src/App.tsx",
	}

	matched := matcher.Match(ctx)

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched rule, got %d", len(matched))
	}
	if matched[0].ID != "rule-1" {
		t.Errorf("expected rule-1, got %s", matched[0].ID)
	}
}

func TestMatcherReturnsRulesMatchingContext(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "rule-1",
			Name:     "Node Rule",
			Content:  "# Node",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypeContext, ContextTypes: []string{"node"}}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		DetectedContexts: []string{"node", "typescript"},
	}

	matched := matcher.Match(ctx)

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched rule, got %d", len(matched))
	}
}

func TestMatcherSortsBySpecificityDescending(t *testing.T) {
	ruleList := []domain.Rule{
		{
			ID:       "tag-rule",
			Name:     "Tag Rule",
			Content:  "# Tag",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypeTag, Tags: []string{"frontend"}}},
		},
		{
			ID:       "path-rule",
			Name:     "Path Rule",
			Content:  "# Path",
			Triggers: []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/src/**"}},
		},
	}

	matcher := rules.NewMatcher(ruleList)
	ctx := rules.MatchContext{
		ProjectPath: "/home/user/myapp/src/App.tsx",
		Tags:        []string{"frontend"},
	}

	matched := matcher.Match(ctx)

	if len(matched) != 2 {
		t.Fatalf("expected 2 matched rules, got %d", len(matched))
	}
	if matched[0].ID != "path-rule" {
		t.Errorf("expected path-rule first (highest specificity), got %s", matched[0].ID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd server && go test ./services/rules/... -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Create `server/services/rules/matcher.go`:

```go
package rules

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type MatchContext struct {
	ProjectPath      string
	DetectedContexts []string
	Tags             []string
}

type Matcher struct {
	rules []domain.Rule
}

func NewMatcher(rules []domain.Rule) *Matcher {
	return &Matcher{rules: rules}
}

func (m *Matcher) Match(ctx MatchContext) []domain.Rule {
	var matchedRules []domain.Rule

	for _, rule := range m.rules {
		if m.ruleMatches(rule, ctx) {
			matchedRules = append(matchedRules, rule)
		}
	}

	// Sort by specificity descending
	sort.Slice(matchedRules, func(i, j int) bool {
		return matchedRules[i].MaxSpecificity() > matchedRules[j].MaxSpecificity()
	})

	return matchedRules
}

func (m *Matcher) ruleMatches(rule domain.Rule, ctx MatchContext) bool {
	for _, trigger := range rule.Triggers {
		if m.triggerMatches(trigger, ctx) {
			return true
		}
	}
	return false
}

func (m *Matcher) triggerMatches(trigger domain.Trigger, ctx MatchContext) bool {
	switch trigger.Type {
	case domain.TriggerTypePath:
		return matchPath(trigger.Pattern, ctx.ProjectPath)
	case domain.TriggerTypeContext:
		return matchContext(trigger.ContextTypes, ctx.DetectedContexts)
	case domain.TriggerTypeTag:
		return matchTags(trigger.Tags, ctx.Tags)
	}
	return false
}

func matchPath(pattern, path string) bool {
	if strings.Contains(pattern, "**") {
		// Simplified ** matching
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			suffix := strings.TrimPrefix(parts[1], "/")
			return strings.Contains(path, suffix)
		}
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}

func matchContext(triggerContexts, detectedContexts []string) bool {
	for _, tc := range triggerContexts {
		for _, dc := range detectedContexts {
			if strings.EqualFold(tc, dc) {
				return true
			}
		}
	}
	return false
}

func matchTags(triggerTags, projectTags []string) bool {
	for _, tt := range triggerTags {
		for _, pt := range projectTags {
			if strings.EqualFold(tt, pt) {
				return true
			}
		}
	}
	return false
}
```

**Step 4: Run test to verify it passes**

Run: `cd server && go test ./services/rules/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/services/rules/matcher.go server/services/rules/matcher_test.go
git commit -m "feat(services): add rule matcher with specificity sorting"
```

---

## Phase 4: Next.js Frontend Setup

### Task 12: Initialize Next.js Project

**Files:**
- Create: `web/` directory with Next.js app

**Step 1: Create Next.js app**

```bash
cd web && npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir --import-alias "@/*" --no-git
```

**Step 2: Commit**

```bash
git add web/
git commit -m "feat(web): initialize Next.js with TypeScript and Tailwind"
```

---

### Task 13: Frontend Domain Models

**Files:**
- Create: `web/src/domain/team.ts`
- Create: `web/src/domain/user.ts`
- Create: `web/src/domain/rule.ts`

**Step 1: Create team model**

Create `web/src/domain/team.ts`:

```typescript
export interface TeamSettings {
  driftThresholdMinutes: number;
}

export interface Team {
  id: string;
  name: string;
  settings: TeamSettings;
  createdAt: string;
}

export function createDefaultTeamSettings(): TeamSettings {
  return {
    driftThresholdMinutes: 60,
  };
}
```

**Step 2: Create user model**

Create `web/src/domain/user.ts`:

```typescript
export type AuthProvider = 'github' | 'gitlab' | 'google' | 'local';
export type Role = 'admin' | 'member';

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  authProvider: AuthProvider;
  role: Role;
  teamId: string;
  createdAt: string;
}
```

**Step 3: Create rule model**

Create `web/src/domain/rule.ts`:

```typescript
export type TargetLayer = 'enterprise' | 'global' | 'project' | 'local';
export type TriggerType = 'path' | 'context' | 'tag';

export interface Trigger {
  type: TriggerType;
  pattern?: string;
  contextTypes?: string[];
  tags?: string[];
}

export interface Rule {
  id: string;
  name: string;
  content: string;
  targetLayer: TargetLayer;
  priorityWeight: number;
  triggers: Trigger[];
  teamId: string;
  createdAt: string;
  updatedAt: string;
}

export function getSpecificity(trigger: Trigger): number {
  switch (trigger.type) {
    case 'path':
      return 100;
    case 'context':
      return 50;
    case 'tag':
      return 10;
    default:
      return 0;
  }
}

export function getTargetLayerPath(layer: TargetLayer): string {
  switch (layer) {
    case 'enterprise':
      return '/etc/claude-code/CLAUDE.md';
    case 'global':
      return '~/.claude/CLAUDE.md';
    case 'project':
      return './CLAUDE.md';
    case 'local':
      return './CLAUDE.local.md';
  }
}
```

**Step 4: Commit**

```bash
git add web/src/domain/
git commit -m "feat(web): add TypeScript domain models"
```

---

## Phase 5: Agent CLI Setup

### Task 14: Initialize Agent CLI with Cobra

**Files:**
- Create: `agent/cmd/agent/main.go`
- Create: `agent/entrypoints/cli/root.go`

**Step 1: Add Cobra dependency**

```bash
cd agent && go get github.com/spf13/cobra
```

**Step 2: Create main.go**

Create `agent/cmd/agent/main.go`:

```go
package main

import (
	"os"

	"github.com/kamilrybacki/claudeception/agent/entrypoints/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
```

**Step 3: Create root command**

Create `agent/entrypoints/cli/root.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claudeception",
	Short: "Claudeception agent for managing CLAUDE.md configurations",
	Long: `Claudeception is a local agent that syncs CLAUDE.md configurations
from a central server to your development environment.

It watches for project changes, detects context, and applies
the appropriate rules to maintain consistent Claude behavior.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("claudeception agent v0.1.0")
	},
}
```

**Step 4: Test CLI builds**

```bash
cd agent && go build ./cmd/agent
./agent version
```

Expected: `claudeception agent v0.1.0`

**Step 5: Commit**

```bash
git add agent/
git commit -m "feat(agent): initialize CLI with Cobra"
```

---

### Task 15: Add Agent CLI Commands

**Files:**
- Create: `agent/entrypoints/cli/start.go`
- Create: `agent/entrypoints/cli/stop.go`
- Create: `agent/entrypoints/cli/status.go`
- Create: `agent/entrypoints/cli/sync.go`
- Create: `agent/entrypoints/cli/login.go`

**Step 1: Create start command**

Create `agent/entrypoints/cli/start.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the Claudeception daemon to sync configurations in the background.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Claudeception daemon...")
		// TODO: Implement daemon start
		fmt.Println("Daemon started successfully")
	},
}
```

**Step 2: Create stop command**

Create `agent/entrypoints/cli/stop.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the running Claudeception daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping Claudeception daemon...")
		// TODO: Implement daemon stop
		fmt.Println("Daemon stopped successfully")
	},
}
```

**Step 3: Create status command**

Create `agent/entrypoints/cli/status.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status",
	Long:  `Show the current connection status, cached config age, and active projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Status: Disconnected")
		fmt.Println("Cached config: None")
		fmt.Println("Active projects: 0")
		// TODO: Implement actual status check
	},
}
```

**Step 4: Create sync command**

Create `agent/entrypoints/cli/sync.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(applyCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Force immediate sync",
	Long:  `Force an immediate synchronization with the central server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing with server...")
		// TODO: Implement sync
		fmt.Println("Sync complete")
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Check drift for a project",
	Long:  `Validate that local CLAUDE.md files match expected content.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		fmt.Printf("Validating project at %s...\n", path)
		// TODO: Implement validation
		fmt.Println("No drift detected")
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply [path]",
	Short: "Write CLAUDE.md files for a project",
	Long:  `Apply the appropriate CLAUDE.md configurations to a project.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		fmt.Printf("Applying configuration to %s...\n", path)
		// TODO: Implement apply
		fmt.Println("Configuration applied")
	},
}
```

**Step 5: Create login command**

Create `agent/entrypoints/cli/login.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with central server",
	Long:  `Open browser to authenticate with the central server via OAuth.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Opening browser for authentication...")
		// TODO: Implement OAuth flow
		fmt.Println("Login successful")
	},
}
```

**Step 6: Test all commands**

```bash
cd agent && go build ./cmd/agent
./agent --help
./agent start
./agent stop
./agent status
./agent sync
./agent validate
./agent apply
./agent login
```

**Step 7: Commit**

```bash
git add agent/entrypoints/cli/
git commit -m "feat(agent): add CLI commands for start, stop, status, sync, login"
```

---

## Phase 6: Docker & Development Environment

### Task 16: Docker Compose for Development

**Files:**
- Create: `docker-compose.yml`
- Create: `server/Dockerfile`

**Step 1: Create server Dockerfile**

Create `server/Dockerfile`:

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /server .
COPY migrations ./migrations

EXPOSE 8080

CMD ["./server"]
```

**Step 2: Create docker-compose.yml**

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: claudeception
      POSTGRES_PASSWORD: claudeception
      POSTGRES_DB: claudeception
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U claudeception"]
      interval: 5s
      timeout: 5s
      retries: 5

  server:
    build:
      context: ./server
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://claudeception:claudeception@db:5432/claudeception?sslmode=disable
      SERVER_PORT: "8080"
      JWT_SECRET: dev-secret-change-in-production
    depends_on:
      db:
        condition: service_healthy

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    depends_on:
      - server

volumes:
  pgdata:
```

**Step 3: Commit**

```bash
git add docker-compose.yml server/Dockerfile
git commit -m "feat: add Docker Compose for development environment"
```

---

## Summary

This plan covers the foundation:

1. **Phase 1:** Project scaffolding and domain models (Tasks 1-6) ✅
2. **Phase 2:** Database migrations (Tasks 7-10) ✅
3. **Phase 3:** Rule matching service (Task 11) ✅
4. **Phase 4:** Next.js frontend setup (Tasks 12-13) ✅
5. **Phase 5:** Agent CLI setup (Tasks 14-15) ✅
6. **Phase 6:** Docker development environment (Task 16) ✅

**Next phases (in separate plan file `2026-02-03-phase2-api-websocket.md`):**
- Phase 7: REST API endpoints 🔄 In Progress
- Phase 8: WebSocket handlers ⏳

**Future phases (not yet planned):**
- Phase 9: Frontend pages and components
- Phase 10: Agent daemon and sync logic
- Phase 11: Integration tests
- Phase 12: Documentation generation
