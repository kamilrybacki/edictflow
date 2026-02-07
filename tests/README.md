# Tests

This directory contains all test suites for Edictflow.

## Structure

```
tests/
├── e2e-go/           # Go-based E2E tests (testcontainers)
├── e2e-playwright/   # Browser-based E2E tests (Playwright)
├── infrastructure/   # Test infrastructure (Docker, seed data)
├── integration/      # Integration tests
└── unit/             # Unit tests
```

## Running Tests

### E2E Tests (Go - Agent/Server)

```bash
# Run all E2E tests
task e2e

# Run swarm synchronization tests
task e2e:swarm

# Run all E2E tests including swarm
task e2e:all

# Clean up orphaned test resources
task e2e:cleanup
```

### E2E Tests (Playwright - Web UI)

First, start the test infrastructure:

```bash
task test:infra:up
```

Then run the Playwright tests:

```bash
# Run Playwright tests
task e2e:playwright

# Run Playwright tests in headed mode (with browser visible)
task e2e:playwright:headed
```

### Test Infrastructure

The test infrastructure provides a complete environment for manual and automated testing:

```bash
# Start test infrastructure (PostgreSQL, Server, Web, User container)
task test:infra:up

# Seed test data
task test:infra:seed

# Shell into user container (simulate developer)
task test:infra:shell

# View logs
task test:infra:logs

# Stop and cleanup
task test:infra:down
```

### Test Accounts

| Email | Password | Role | Team | Notes |
|-------|----------|------|------|-------|
| alex.rivera@test.local | Test1234 | Admin | Engineering | Primary E2E test account |
| jordan.kim@test.local | Test1234 | Member | Engineering | Secondary test user |
| sarah.chen@test.local | Test1234 | Member | Design | Cross-team testing |
| admin@example.com | Password123 | Admin | Engineering | UI test account |

## Test Categories

### e2e-go/
Go-based end-to-end tests using testcontainers-go. These tests spin up actual containers for PostgreSQL, Server, and Agent to test the full system integration.

- `enforcement_test.go` - Tests rule enforcement modes (block, warning, temporary)
- `swarm_sync_test.go` - Tests multi-agent synchronization
- `suite_test.go` - Test suite setup and teardown

### e2e-playwright/ (web/e2e/)
Browser-based end-to-end tests using Playwright. These tests verify the web UI functionality.

**100+ E2E tests** covering:

| File | Tests | Coverage |
|------|-------|----------|
| `home.spec.ts` | 7 | Basic pages, authentication, responsive design |
| `user-flows.spec.ts` | 11 | Complete user journeys, team/rule management |
| `admin-features.spec.ts` | 12 | Admin functionality, RBAC |
| `graph-view.spec.ts` | 18 | Graph View page, node interactions, filtering |
| `approvals.spec.ts` | 17 | Approval workflows, status management |
| `changes.spec.ts` | 20 | Change detection, filtering, details |

### infrastructure/
Resources for running tests:

- `Dockerfile.user` - User simulation container
- `seed-data.sql` - Test data for PostgreSQL
- `setup.sh` - Environment setup script
