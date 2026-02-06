# Local Setup

Set up a complete Edictflow development environment on your machine.

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.22+ | Server and agent |
| Docker | 20.10+ | Containers |
| Docker Compose | 2.0+ | Orchestration |
| Node.js | 20+ | Web UI |
| Task | 3.0+ | Build automation |

## Quick Setup

```bash
# Clone repository
git clone https://github.com/kamilrybacki/edictflow.git
cd edictflow

# Install Task runner
go install github.com/go-task/task/v3/cmd/task@latest

# Start all services
task dev
```

This starts:

- PostgreSQL database
- Server (API)
- Web UI

## Manual Setup

### 1. Install Dependencies

```bash
# Install development tools
task setup:deps

# This installs:
# - golang-migrate (database migrations)
# - golangci-lint (code linting)
```

### 2. Start Database

```bash
# Start PostgreSQL in Docker
task db:start

# Wait for it to be ready
task db:wait

# Run migrations
task db:migrate
```

### 3. Start Server

```bash
# Run server locally (connects to Docker PostgreSQL)
task dev:local:server

# Or with hot reload (requires air)
task dev:local:watch
```

### 4. Start Web UI

```bash
cd web
npm install
npm run dev
```

## Docker Development

### Start All Services

```bash
task dev
```

### View Logs

```bash
# All services
task logs

# Specific service
task logs:server
task logs:web
```

### Rebuild Services

```bash
# Rebuild all
task dev:rebuild

# Rebuild specific
task dev:rebuild:server
task dev:rebuild:web
```

### Stop Services

```bash
task down
```

## Database Management

### Connect to Database

```bash
task db:psql
```

### Run Migrations

```bash
# Apply all migrations
task db:migrate

# Rollback last migration
task db:migrate:down

# Reset database
task db:reset
```

### Create New Migration

```bash
task db:migrate:create -- add_new_table
```

This creates:

- `server/migrations/NNNNNN_add_new_table.up.sql`
- `server/migrations/NNNNNN_add_new_table.down.sql`

## Agent Development

### Build Agent

```bash
task agent:build
```

This creates `agent/agent` binary.

### Run Agent

```bash
cd agent

# Login (use local server)
./agent login http://localhost:8080

# Start in foreground for debugging
./agent start --foreground

# With polling mode
./agent start --foreground --poll-interval 500ms
```

### Test Agent

```bash
task agent:test
```

## Environment Variables

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | Dynamic | PostgreSQL connection |
| `SERVER_PORT` | 8080 | HTTP port |
| `JWT_SECRET` | dev-secret | JWT signing key |

### Web UI

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | http://localhost:8080 | API URL |

## IDE Setup

### VS Code

Recommended extensions:

- Go (golang.go)
- ESLint
- Prettier
- Docker
- GitLens

Settings (`.vscode/settings.json`):

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

### GoLand

1. Enable `gofmt` on save
2. Configure `golangci-lint` as external tool
3. Set Go modules integration

## Debugging

### Server

Use VS Code launch configuration:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/server/cmd/server",
      "env": {
        "DATABASE_URL": "postgres://edictflow:edictflow@localhost:5432/edictflow?sslmode=disable",
        "JWT_SECRET": "dev-secret",
        "SERVER_PORT": "8080"
      }
    }
  ]
}
```

### Agent

```json
{
  "name": "Debug Agent",
  "type": "go",
  "request": "launch",
  "mode": "auto",
  "program": "${workspaceFolder}/agent/cmd/agent",
  "args": ["start", "--foreground"]
}
```

### Web UI

Use browser DevTools or VS Code debugger for Next.js.

## Common Issues

### Port Already in Use

Edictflow automatically finds available ports. To see which ports are being used:

```bash
cat .db_port .server_port .web_port
```

To reset:

```bash
rm -f .db_port .server_port .web_port
task dev
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker compose ps

# Check logs
docker compose logs db

# Restart database
task db:restart
```

### Go Module Issues

```bash
cd server
go mod tidy
go mod download
```

### Web Build Fails

```bash
cd web
rm -rf node_modules .next
npm install
npm run dev
```

## Seeding Test Data

### Via API

After starting services, create test data:

```bash
# Get auth token (from web UI or device flow)
TOKEN="your-token"

# Create team
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "Test Team"}'

# Create rule
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Test Rule",
    "team_id": "team-uuid",
    "content": "# Test",
    "enforcement_mode": "warning",
    "triggers": [{"type": "path", "pattern": "CLAUDE.md"}]
  }'
```

### Via Database

```bash
task db:psql

-- Insert test data
INSERT INTO teams (id, name) VALUES (gen_random_uuid(), 'Engineering');
```

## Next Steps

- [Testing Guide](testing.md) - Learn about the testing strategy
- [Contributing Guide](contributing.md) - How to contribute
