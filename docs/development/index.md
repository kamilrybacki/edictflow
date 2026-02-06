# Development

Welcome to the Edictflow development guide. This section covers setting up a development environment, testing, and contributing to the project.

## Quick Navigation

<div class="grid" markdown>

<div class="card" markdown>

### [Local Setup](local-setup.md)

Set up a development environment on your machine.

</div>

<div class="card" markdown>

### [Testing](testing.md)

Run tests, write tests, and understand the testing strategy.

</div>

<div class="card" markdown>

### [Contributing](contributing.md)

How to contribute to Edictflow.

</div>

</div>

## Project Structure

```
edictflow/
├── server/                 # Go backend server
│   ├── cmd/server/         # Server entry point
│   ├── entrypoints/        # HTTP handlers
│   │   └── api/
│   │       ├── handlers/   # Request handlers
│   │       └── middleware/ # Middleware
│   ├── services/           # Business logic
│   │   ├── rules/          # Rules service
│   │   ├── changes/        # Changes service
│   │   └── roles/          # RBAC service
│   ├── storage/            # Database layer
│   │   └── postgres/       # PostgreSQL implementation
│   ├── migrations/         # SQL migrations
│   └── integration/        # Integration tests
│
├── agent/                  # Go CLI agent
│   ├── cmd/agent/          # Agent entry point
│   ├── entrypoints/cli/    # CLI commands
│   ├── daemon/             # Background service
│   ├── watcher/            # File monitoring
│   └── storage/            # SQLite layer
│
├── web/                    # Next.js frontend
│   ├── src/
│   │   ├── app/            # Next.js app router
│   │   ├── components/     # React components
│   │   └── lib/            # Utilities
│   └── public/             # Static assets
│
├── e2e/                    # E2E tests
│   ├── suite_test.go       # Test suite
│   ├── helpers_test.go     # Container helpers
│   └── enforcement_test.go # Enforcement tests
│
├── docs/                   # Documentation (this site)
├── Taskfile.yml            # Task automation
└── docker-compose.yml      # Local development
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Server | Go 1.22+ | Backend API |
| Router | Chi | HTTP routing |
| Database | PostgreSQL 16 | Primary storage |
| Migrations | golang-migrate | Schema management |
| Agent | Go 1.22+ | CLI and daemon |
| Agent DB | SQLite | Local cache |
| Frontend | Next.js 14 | Web UI |
| E2E Tests | testcontainers-go | Container-based testing |
| Docs | MkDocs Material | Documentation |

## Getting Started

1. **Clone the repository:**
   ```bash
   git clone https://github.com/kamilrybacki/edictflow.git
   cd edictflow
   ```

2. **Install Task:**
   ```bash
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

3. **Start development environment:**
   ```bash
   task dev
   ```

4. **Access services:**
   - Web UI: http://localhost:3000
   - API: http://localhost:8080
   - PostgreSQL: localhost:5432

## Available Tasks

Run `task --list` to see all available tasks:

```
task: Available tasks for this project:
* build:              Build all components
* build:docker:       Build all Docker images
* build:server:       Build the server binary
* check:              Run all code quality checks
* clean:              Clean build artifacts
* clean:all:          Clean everything
* db:migrate:         Run all up migrations
* db:migrate:create:  Create a new migration
* db:migrate:down:    Rollback last migration
* db:psql:            Open psql shell
* db:reset:           Reset database
* db:start:           Start PostgreSQL
* db:stop:            Stop PostgreSQL
* dev:                Start all services
* dev:rebuild:        Rebuild and restart
* down:               Stop all services
* e2e:                Run E2E tests
* fmt:                Format Go code
* lint:               Run linter
* logs:               View logs
* test:               Run all tests
* test:integration:   Run integration tests
* test:unit:          Run unit tests
```

## Development Workflow

### Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make changes to the code

3. Run tests:
   ```bash
   task test
   ```

4. Run linter:
   ```bash
   task lint
   ```

5. Commit changes:
   ```bash
   git add .
   git commit -m "feat: add my feature"
   ```

6. Push and create PR:
   ```bash
   git push -u origin feature/my-feature
   ```

### Hot Reload

For faster development:

- **Server:** Changes require rebuild with `task dev:rebuild:server`
- **Web:** Next.js hot reloads automatically
- **Agent:** Build with `task agent:build`

## Code Style

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write table-driven tests

### TypeScript

- Use TypeScript strict mode
- Follow ESLint configuration
- Use Prettier for formatting

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: fix bug
docs: update documentation
refactor: refactor code
test: add tests
chore: maintenance tasks
```

## Need Help?

- [GitHub Issues](https://github.com/kamilrybacki/edictflow/issues)
- [Contributing Guide](contributing.md)
