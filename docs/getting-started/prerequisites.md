# Prerequisites

Before installing Edictflow, ensure you have the following prerequisites met.

## Server Requirements

### For Docker Deployment (Recommended)

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| Docker | 20.10+ | 24.0+ |
| Docker Compose | 2.0+ | 2.20+ |
| RAM | 2 GB | 4 GB |
| Disk | 10 GB | 20 GB |
| CPU | 2 cores | 4 cores |

### For Manual Deployment

| Requirement | Version |
|-------------|---------|
| Go | 1.22+ |
| PostgreSQL | 14+ |
| Node.js | 20+ (for Web UI) |

## Agent Requirements

The agent is a lightweight Go binary with minimal requirements:

| Requirement | Details |
|-------------|---------|
| OS | Linux, macOS, Windows |
| Architecture | amd64, arm64 |
| RAM | 50 MB |
| Disk | 20 MB |
| Network | HTTPS/WSS to server |

## Network Requirements

### Server

The server needs the following network access:

| Port | Protocol | Purpose |
|------|----------|---------|
| 8080 | HTTP/HTTPS | REST API |
| 8080 | WS/WSS | WebSocket connections |
| 5432 | TCP | PostgreSQL (internal) |
| 3000 | HTTP | Web UI (optional) |

### Agent

The agent needs outbound access to:

- Your Edictflow server (HTTPS/WSS)
- No inbound ports required

## Authentication Prerequisites

### OAuth/SSO (Optional)

If using OAuth for authentication:

| Provider | Required |
|----------|----------|
| GitHub | OAuth App credentials |
| Google | OAuth 2.0 credentials |
| Custom OIDC | Provider configuration |

### Device Code Flow

The agent uses OAuth 2.0 Device Code flow for CLI authentication. Ensure your OAuth provider supports this flow.

## Development Prerequisites

For local development:

```bash
# Required
go install github.com/go-task/task/v3/cmd/task@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Optional (for linting)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Optional (for hot reload)
go install github.com/air-verse/air@latest
```

## Verification

Verify your environment is ready:

=== "Docker"

    ```bash
    docker --version
    docker compose version
    ```

=== "Manual"

    ```bash
    go version
    psql --version
    node --version
    ```

=== "Development"

    ```bash
    task --version
    migrate --version
    ```

## Next Steps

Once prerequisites are met:

- [Quick Start](quickstart.md) - Get running in 5 minutes
- [Server Deployment](../admin/deployment.md) - Production deployment guide
- [Agent Installation](../user/installation.md) - Install the agent
