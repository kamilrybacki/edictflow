# Claudeception Design Document

A team collaboration tool for managing layered CLAUDE.md configurations with a centralized web UI and local agents.

## Overview

**Claudeception** consists of three components:

1. **Central Server (Go)** - Hosts web UI and API, stores configurations in PostgreSQL, maintains WebSocket connections to agents
2. **Web UI (TypeScript/Next.js)** - Team management, rule editor, drift dashboard
3. **Local Agent (Go)** - Background daemon on each developer's machine, syncs configurations, validates drift

**Authentication:** OAuth (GitHub/GitLab/Google) as primary, with self-managed email/password fallback.

---

## CLAUDE.md Hierarchy

Claudeception manages the layered CLAUDE.md system as defined by Claude Code:

| Level | Path | Managed By | Agent Behavior |
|-------|------|------------|----------------|
| **Enterprise** | `/etc/claude-code/CLAUDE.md` | IT/Admin (requires root) | Read-only reference; admin pushes manually or via config management |
| **Global User** | `~/.claude/CLAUDE.md` | Claudeception agent | Agent writes team-wide personal standards |
| **Project** | `<repo>/CLAUDE.md` | Claudeception agent | Agent writes team project rules |
| **Project Local** | `<repo>/CLAUDE.local.md` | User (gitignored) | Agent can write user's personal overrides |

Files load in order, with each level able to override or extend previous ones.

---

## Rule System & Triggering

### Rule Definition

A rule consists of:
- **Name** - Human-readable identifier (e.g., "React Frontend Standards")
- **Content** - The CLAUDE.md content to inject
- **Target layer** - Which hierarchy level (enterprise/global/project/local)
- **Trigger conditions** - When this rule applies
- **Priority weight** - For ordering when multiple rules match

### Trigger Types (specificity order, highest priority first)

1. **Path patterns** - Glob patterns like `**/frontend/**/*.tsx`
2. **Context detection** - Auto-detected from project files (package.json, go.mod, Cargo.toml, etc.)
3. **Tags** - Explicit labels assigned to projects in the UI

### Conflict Resolution

When multiple rules match, they merge in specificity order. More specific rules override less specific ones. For same specificity, priority weight breaks ties.

---

## Local Agent Architecture

### Daemon Mode

The agent runs as a background process, started at login:
- Maintains persistent WebSocket connection to central server
- Watches filesystem for project directory changes
- Applies configurations when projects are opened/detected
- Sends heartbeat + status updates to server

### Project Detection

When the agent detects a project directory:
1. Scans for context markers (package.json, go.mod, etc.)
2. Checks path against configured path patterns
3. Looks up any explicit tags assigned in the UI
4. Requests matching rules from server (or uses cache if offline)
5. Merges rules by specificity and writes CLAUDE.md files to appropriate locations

### Drift Validation

Periodically (and on-demand via CLI), the agent:
- Compares local CLAUDE.md files against expected content
- Reports drift to central server (shown in dashboard)
- Optionally auto-fixes drift or prompts user

### CLI Commands

| Command | Description |
|---------|-------------|
| `claudeception start` | Start daemon |
| `claudeception stop` | Stop daemon |
| `claudeception status` | Show connection status, cached config age |
| `claudeception sync` | Force immediate sync |
| `claudeception validate [path]` | Check drift for a project |
| `claudeception apply [path]` | Write CLAUDE.md files for a project |
| `claudeception login` | Authenticate with central server |

### Offline Behavior

Cache stored in `~/.claudeception/cache/`. Agent continues applying cached rules when disconnected, marks status as "stale" after configurable threshold.

---

## Central Server & Data Model

### Core Data Models

```
Team
├── id, name, created_at
├── settings (JSON: priority order preferences, drift thresholds)
└── has many: Users, Rules, Projects

User
├── id, email, name, avatar_url
├── auth_provider (github | gitlab | google | local)
├── role (admin | member)
└── belongs to: Team

Rule
├── id, name, content (CLAUDE.md text)
├── target_layer (enterprise | global | project | local)
├── priority_weight (integer)
├── triggers (JSON array)
└── belongs to: Team

Project
├── id, path_pattern (for matching)
├── tags (array)
├── last_seen_at
└── belongs to: Team

Agent
├── id, machine_id, user_id
├── last_heartbeat, status (online | stale | offline)
├── cached_config_version
└── belongs to: User
```

---

## Web UI Design

### Navigation Structure

```
Dashboard (home)
├── Teams
│   ├── Team Settings
│   └── Members
├── Rules
│   ├── Rule List
│   └── Rule Editor
├── Projects
│   └── Project Details
├── Agents
│   └── Agent Status
└── Settings (personal)
```

### Rule Editor (dual-mode)

1. **Form Mode** (default)
   - Name field
   - Target layer dropdown (enterprise/global/project/local)
   - Priority weight slider
   - Trigger builder:
     - Path pattern input with glob syntax helper
     - Context type multi-select (Node, Go, Rust, Python, etc.)
     - Tag selector (from existing tags or create new)
   - Content editor with markdown preview and snippet library

2. **Raw Mode** (toggle for power users)
   - Single YAML/JSON textarea with full rule definition
   - Paste existing rules directly
   - Syntax validation with inline errors

### Drift Dashboard

- Table of agents with status indicators
- Expandable rows showing which files have drifted
- One-click "push fix" to sync agent, or "accept local" to update rule

---

## Project Structure

Follows Clean/Hexagonal architecture with domain at center, services around it, entrypoints on the outside.

```
claudeception/
├── server/                      # Go backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Entrypoint
│   ├── domain/                  # Core business models
│   │   ├── team.go
│   │   ├── user.go
│   │   ├── rule.go
│   │   ├── project.go
│   │   └── agent.go
│   ├── services/                # Business logic layer
│   │   ├── auth/                # Authentication service
│   │   ├── rules/               # Rule matching & merging
│   │   ├── sync/                # Agent synchronization
│   │   └── detection/           # Context detection logic
│   ├── entrypoints/             # External interfaces
│   │   ├── api/                 # REST handlers
│   │   └── ws/                  # WebSocket handlers
│   ├── common/
│   │   └── utils/
│   ├── configurator/            # DI container, settings
│   │   ├── container.go
│   │   └── settings.go
│   └── migrations/
│
├── web/                         # Next.js frontend
│   ├── app/
│   ├── components/
│   ├── domain/                  # TypeScript domain models
│   ├── services/                # API client, business logic
│   └── lib/
│
├── agent/                       # Go local agent
│   ├── cmd/
│   │   └── agent/
│   │       └── main.go
│   ├── domain/
│   ├── services/
│   │   ├── daemon/
│   │   ├── detector/
│   │   ├── sync/
│   │   └── cache/
│   ├── entrypoints/
│   │   └── cli/
│   └── configurator/
│
├── tests/                       # Shared test infrastructure
│   ├── assets/
│   │   └── mock_server/         # Containerized mock central server
│   ├── functional/
│   └── integration/
│
└── docs/
    ├── api/
    │   ├── openapi.yaml
    │   └── swagger-ui/
    ├── agent/
    │   └── cli-reference.md
    ├── components/
    │   └── storybook/
    └── reference/
        ├── server/
        ├── agent/
        └── web/
```

---

## Technology Stack

### Backend (Go)

- **Framework:** Chi or Gin for HTTP routing
- **WebSocket:** gorilla/websocket
- **Database:** pgx for PostgreSQL
- **Migrations:** golang-migrate
- **Auth:** goth for OAuth providers

### Frontend (TypeScript)

- **Framework:** Next.js (App Router)
- **Styling:** Tailwind CSS
- **State:** Zustand or React Query for server state
- **Forms:** React Hook Form + Zod validation
- **Editor:** Monaco or CodeMirror for raw mode

### Local Agent (Go)

- **CLI:** Cobra
- **Filesystem watch:** fsnotify
- **Config cache:** SQLite or flat JSON files
- **System service:** kardianos/service for cross-platform daemon

---

## Automatically Generated Documentation

| Component | Tool | Output |
|-----------|------|--------|
| Go API | swag | OpenAPI spec + Swagger UI |
| Go packages | godoc | API reference |
| Agent CLI | Cobra docs | CLI reference markdown |
| React components | Storybook | Component library |
| TypeScript | TypeDoc | API reference |

CI regenerates all docs on every PR and deploys to `/docs` route in web UI.

---

## Testing Strategy

### Go Components

| Level | Location | Dependencies |
|-------|----------|--------------|
| Unit | `*_test.go` alongside code | Mocked via interfaces |
| Integration | `tests/integration/` | Testcontainers |
| E2E | `tests/e2e/` | Full containerized stack |

### TypeScript Frontend

| Level | Tool | Purpose |
|-------|------|---------|
| Unit | Vitest | Component logic, utilities |
| Component | Testing Library | UI behavior |
| E2E | Playwright | Full browser flows |

---

## WebSocket Protocol

### Message Format

```json
{
  "type": "config_update",
  "id": "msg-uuid-here",
  "timestamp": "2024-01-15T10:30:00Z",
  "payload": { ... }
}
```

### Server → Agent Messages

| Type | Payload | Description |
|------|---------|-------------|
| `config_update` | `{ rules, version }` | Push updated rules |
| `sync_request` | `{ project_paths }` | Request re-validation |
| `ack` | `{ ref_id }` | Acknowledge agent message |

### Agent → Server Messages

| Type | Payload | Description |
|------|---------|-------------|
| `heartbeat` | `{ status, cached_version, active_projects }` | Periodic status |
| `drift_report` | `{ project_path, expected_hash, actual_hash, diff }` | File mismatch |
| `context_detected` | `{ project_path, detected_context, detected_tags }` | New project |
| `sync_complete` | `{ project_path, files_written }` | Confirms sync |

### Reconnection Strategy

1. Exponential backoff (1s, 2s, 4s... max 60s)
2. On reconnect, agent sends heartbeat with cached version
3. Server compares versions, sends config_update if stale

---

## Deployment & Distribution

### Server

| Option | Description |
|--------|-------------|
| Docker Compose | Single-node: server + PostgreSQL + web |
| Kubernetes | Helm chart for larger deployments |
| Binary + managed DB | Single Go binary, external PostgreSQL |

### Agent Distribution

| Platform | Format |
|----------|--------|
| macOS | Homebrew tap, `.pkg` installer |
| Linux | `.deb`, `.rpm`, standalone binary |
| Windows | `.msi` installer, Scoop package |

Cross-compiled via GoReleaser.

### Agent Configuration

`~/.claudeception/config.yaml`:
```yaml
server_url: https://claudeception.yourteam.com
auth_token: "eyJ..."
sync_interval: 30s
watch_paths:
  - ~/projects
  - ~/work
```

---

## Design Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Backend language | Go | Single binary distribution, good concurrency |
| Frontend | Next.js + TypeScript | Team familiarity, good DX |
| Database | PostgreSQL | Relational, proven, complex queries |
| Agent-server communication | WebSocket | Real-time push, immediate sync |
| Offline handling | Cache last-known config | Resilience, continues working |
| Rule conflicts | Specificity wins | Predictable, clear precedence |
| Trigger system | Path + context + tags | Flexible, covers all use cases |
| Authentication | OAuth + local fallback | Convenience + air-gapped support |
| Architecture | Clean/Hexagonal | Testable, maintainable per standards |
