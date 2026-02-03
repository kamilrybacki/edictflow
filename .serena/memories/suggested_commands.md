# Suggested Commands for claudeception

## Testing

```bash
# Run unit tests only
cd server && go test -v -short ./...

# Run integration tests (requires Docker)
cd server && go test -v -tags=integration ./integration/...

# Run all tests
cd server && make test

# Run a specific integration test
cd server && go test -v -tags=integration -run TestTeamRepository_CreateAndGetByID ./integration/...
```

## Build

```bash
cd server && go build -o server ./cmd/server
```

## Development

```bash
# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

## Project Structure

- `server/` - Go backend
  - `adapters/postgres/` - PostgreSQL implementations of repository interfaces
  - `cmd/server/` - Main entry point
  - `common/db/` - Database pool utilities
  - `configurator/` - Settings/configuration
  - `domain/` - Domain models (Team, Rule, User, Agent, Project)
  - `entrypoints/api/` - HTTP API (chi router, handlers, middleware)
  - `entrypoints/ws/` - WebSocket handling
  - `integration/` - Integration tests with testcontainers
  - `migrations/` - SQL migrations
  - `services/` - Business logic (teams, rules)
