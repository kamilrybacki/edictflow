# Test Infrastructure

This directory contains files for running a dockerized test environment where you can manually test Edictflow as a user.

## Quick Start

```bash
# Start the infrastructure
task test:infra:up

# Shell into the user container
task test:infra:shell

# Inside the container, you're a developer with the agent installed
edictflow-agent login http://master:8080
edictflow-agent start --foreground
```

## What's Included

### Services

| Service | URL | Description |
|---------|-----|-------------|
| `redis` | (internal) | Redis pub/sub for real-time events |
| `db` | localhost:5433 | PostgreSQL database |
| `master` | localhost:8080 | Edictflow Master API server |
| `worker` | (internal) | Edictflow Worker for WebSocket connections |
| `web` | localhost:3000 | Web UI |
| `user` | (shell) | User simulation container |

### Test Data

The seed data includes:

- **Test Team**: "Test Team"
- **Test User**: developer@test.local
- **Test Rule (block)**: "Standard CLAUDE.md" - will revert changes
- **Test Rule (warning)**: "Guidelines (Warning)" - logs only

## Usage

### As a User

1. Start infrastructure: `task test:infra:up`
2. Shell in: `task test:infra:shell`
3. Login: `edictflow-agent login http://master:8080`
4. Start agent: `edictflow-agent start --foreground`
5. Edit file: `vim ~/workspace/CLAUDE.md`
6. Watch the agent respond based on enforcement mode

### Testing Enforcement Modes

**Block Mode** (default rule):
- Edit `~/workspace/CLAUDE.md`
- Watch it get reverted immediately

**Warning Mode**:
- Create `~/workspace/GUIDELINES.md`
- Edit it
- Check the server logs or Web UI for the flagged event

### Web UI

Access the Web UI at http://localhost:3000 to:
- View rules
- See change events
- Manage enforcement modes

## Commands

```bash
# Start infrastructure
task test:infra:up

# Shell into user container
task test:infra:shell

# View logs
task test:infra:logs

# Stop and cleanup
task test:infra:down

# Re-seed data
task test:infra:seed
```

## Files

| File | Purpose |
|------|---------|
| `Dockerfile.user` | User simulation container with agent |
| `seed-data.sql` | Test data for database |
| `generate-token.go` | Generate test JWT tokens |
| `README.md` | This file |

## Troubleshooting

### Agent won't connect

1. Check master is running: `curl http://master:8080/health`
2. Verify network: `ping master`
3. Check logs: `task test:infra:logs`

### Database issues

```bash
# Reset everything
task test:infra:down
task test:infra:up
```

### Port conflicts

If ports 3000, 5433, or 8080 are in use, stop other services or modify `docker-compose.test.yml`.
