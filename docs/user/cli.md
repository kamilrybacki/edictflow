# CLI Commands

Complete reference for the Edictflow agent CLI.

## Global Flags

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help for any command |
| `--version`, `-v` | Show agent version |
| `--config` | Path to config file |
| `--data-dir` | Override data directory |
| `--verbose` | Enable verbose output |

## Commands

### login

Authenticate with a Edictflow server.

```bash
edictflow-agent login <server-url>
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `server-url` | Yes | Server URL (e.g., `https://edictflow.example.com`) |

**Example:**

```bash
edictflow-agent login https://edictflow.yourcompany.com
```

**Flow:**

1. Agent requests a device code from the server
2. Displays code and verification URL
3. Opens browser automatically (or prompts to open manually)
4. Waits for authentication to complete
5. Stores credentials locally

**Output:**

```
Please open the following URL in your browser:
  https://edictflow.yourcompany.com/device

Enter this code: ABC-123

Waiting for authentication...
Authentication successful!
Logged in as: user@example.com
```

---

### logout

Remove stored credentials.

```bash
edictflow-agent logout
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force`, `-f` | Skip confirmation prompt |

---

### start

Start the agent daemon.

```bash
edictflow-agent start [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--foreground`, `-f` | `false` | Run in foreground (don't daemonize) |
| `--poll-interval` | `0` | Poll interval for file watching (e.g., `500ms`) |
| `--server` | - | Override server URL |

**Examples:**

```bash
# Start as background daemon
edictflow-agent start

# Run in foreground (for debugging)
edictflow-agent start --foreground

# Use polling mode (for containers/network filesystems)
edictflow-agent start --poll-interval 500ms
```

**Notes:**

- The agent connects to the server via WebSocket
- Rules are synced immediately on connection
- File watching begins after rules are received
- Use `--poll-interval` when fsnotify is unreliable (containers, NFS)

---

### stop

Stop the running agent daemon.

```bash
edictflow-agent stop
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force`, `-f` | Force stop (SIGKILL) |
| `--timeout` | Graceful shutdown timeout (default: 10s) |

---

### restart

Restart the agent daemon.

```bash
edictflow-agent restart
```

Equivalent to `stop` followed by `start`.

---

### status

Show agent status.

```bash
edictflow-agent status [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

**Output:**

```
Agent Status: Running
  PID: 12345
  Uptime: 2 hours 15 minutes

Server: https://edictflow.yourcompany.com
  Connected: Yes
  Latency: 45ms

User: user@example.com
  Team: Engineering

Rules: 3 active
  - Standard CLAUDE.md (block)
  - Project Guidelines (warning)
  - Security Rules (temporary)

Last Sync: 2 minutes ago
Next Sync: in 28 minutes
```

**JSON Output:**

```json
{
  "status": "running",
  "pid": 12345,
  "uptime_seconds": 8100,
  "server": {
    "url": "https://edictflow.yourcompany.com",
    "connected": true,
    "latency_ms": 45
  },
  "user": {
    "email": "user@example.com",
    "team": "Engineering"
  },
  "rules": [
    {
      "id": "rule-uuid",
      "name": "Standard CLAUDE.md",
      "enforcement": "block"
    }
  ],
  "last_sync": "2024-01-15T14:30:00Z"
}
```

---

### sync

Manually sync rules with the server.

```bash
edictflow-agent sync [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Force full sync (ignore cache) |

**Output:**

```
Syncing with server...
  Downloaded 3 rules
  Updated: Standard CLAUDE.md
  Unchanged: 2 rules
Sync complete.
```

---

### validate

Validate local CLAUDE.md against current rules.

```bash
edictflow-agent validate [path] [flags]
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | No | Path to validate (default: current directory) |

**Flags:**

| Flag | Description |
|------|-------------|
| `--fix` | Apply fixes for violations |
| `--json` | Output as JSON |

**Output:**

```
Validating /path/to/project...

CLAUDE.md
  ✓ Matches rule: Standard CLAUDE.md

src/CLAUDE.md
  ✗ Violation: Content differs from rule
    Expected: # CLAUDE.md ...
    Actual:   # Modified content ...

Validation: 1 passed, 1 failed
```

---

### rules

List active rules.

```bash
edictflow-agent rules [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |
| `--all` | Include inactive rules |

**Output:**

```
Active Rules (3):

  Standard CLAUDE.md
    ID: rule-uuid-1
    Enforcement: block
    Triggers: CLAUDE.md, **/CLAUDE.md
    Team: Engineering

  Project Guidelines
    ID: rule-uuid-2
    Enforcement: warning
    Triggers: .claude/guidelines.md
    Team: Engineering

  Security Rules
    ID: rule-uuid-3
    Enforcement: temporary
    Triggers: .claude/security.md
    Team: Security
```

---

### changes

List recent change events.

```bash
edictflow-agent changes [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--limit`, `-n` | 10 | Number of changes to show |
| `--type` | all | Filter by type (blocked, detected, flagged) |
| `--json` | false | Output as JSON |

**Output:**

```
Recent Changes:

  2024-01-15 14:30:00  BLOCKED
    File: /project/CLAUDE.md
    Rule: Standard CLAUDE.md

  2024-01-15 12:15:00  DETECTED
    File: /project/src/CLAUDE.md
    Rule: Project Guidelines

  2024-01-15 10:00:00  FLAGGED
    File: /project/.claude/security.md
    Rule: Security Rules
```

---

### version

Show version information.

```bash
edictflow-agent version
```

**Output:**

```
edictflow-agent version 1.0.0
  Built: 2024-01-15T10:30:00Z
  Commit: abc1234def5678
  Go: go1.22.0
  OS/Arch: darwin/arm64
```

---

### update

Update the agent to the latest version.

```bash
edictflow-agent update [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--check` | Only check for updates, don't install |
| `--version` | Update to specific version |

**Output:**

```
Current version: 1.0.0
Latest version: 1.1.0

Downloading update...
Installing...
Update complete! Restart the agent to apply.
```

---

### config

Manage agent configuration.

```bash
edictflow-agent config <subcommand>
```

**Subcommands:**

#### config show

Display current configuration.

```bash
edictflow-agent config show
```

#### config set

Set a configuration value.

```bash
edictflow-agent config set <key> <value>
```

**Examples:**

```bash
edictflow-agent config set log.level debug
edictflow-agent config set sync.interval 5m
```

#### config reset

Reset configuration to defaults.

```bash
edictflow-agent config reset
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `EDICTFLOW_SERVER` | Override server URL |
| `EDICTFLOW_DATA_DIR` | Override data directory |
| `EDICTFLOW_LOG_LEVEL` | Log level (debug, info, warn, error) |
| `EDICTFLOW_NO_COLOR` | Disable colored output |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication required |
| 3 | Connection failed |
| 4 | Validation failed |

## Shell Completion

### Bash

```bash
edictflow-agent completion bash > /etc/bash_completion.d/edictflow-agent
```

### Zsh

```bash
edictflow-agent completion zsh > "${fpath[1]}/_edictflow-agent"
```

### Fish

```bash
edictflow-agent completion fish > ~/.config/fish/completions/edictflow-agent.fish
```
