# Database Setup

Edictflow uses PostgreSQL as its primary database. This guide covers setup, migrations, and maintenance.

## PostgreSQL Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| Version | 14 | 16 |
| Storage | 1 GB | 10 GB |
| RAM | 512 MB | 2 GB |

## Docker Setup

### Using Docker Compose

The default `docker-compose.yml` includes PostgreSQL:

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: edictflow
      POSTGRES_PASSWORD: edictflow
      POSTGRES_DB: edictflow
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

Start with:

```bash
docker compose up -d db
```

### Standalone Container

```bash
docker run -d \
  --name edictflow-db \
  -e POSTGRES_USER=edictflow \
  -e POSTGRES_PASSWORD=your-secure-password \
  -e POSTGRES_DB=edictflow \
  -v edictflow-data:/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:16-alpine
```

## Manual Setup

### Installation

=== "Ubuntu/Debian"

    ```bash
    sudo apt update
    sudo apt install postgresql-16
    ```

=== "RHEL/CentOS"

    ```bash
    sudo dnf install postgresql16-server
    sudo postgresql-setup --initdb
    sudo systemctl start postgresql
    ```

=== "macOS"

    ```bash
    brew install postgresql@16
    brew services start postgresql@16
    ```

### Create Database

```bash
# Connect as postgres user
sudo -u postgres psql

# Create user and database
CREATE USER edictflow WITH PASSWORD 'your-secure-password';
CREATE DATABASE edictflow OWNER edictflow;
GRANT ALL PRIVILEGES ON DATABASE edictflow TO edictflow;

# Enable UUID extension
\c edictflow
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

## Connection String

Format:

```
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=MODE
```

Examples:

```bash
# Local development
postgres://edictflow:edictflow@localhost:5432/edictflow?sslmode=disable

# Production with SSL
postgres://edictflow:password@db.example.com:5432/edictflow?sslmode=require

# With connection pool
postgres://edictflow:password@db.example.com:5432/edictflow?sslmode=require&pool_max_conns=25
```

## Migrations

Edictflow uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Install migrate CLI

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Run Migrations

Migrations run automatically on server startup. To run manually:

```bash
# Using Taskfile
task db:migrate

# Using migrate directly
migrate -path server/migrations -database "$DATABASE_URL" up
```

### Migration Commands

| Command | Description |
|---------|-------------|
| `task db:migrate` | Run all pending migrations |
| `task db:migrate:down` | Rollback last migration |
| `task db:migrate:reset` | Rollback all migrations |
| `task db:migrate:create -- name` | Create new migration |

### Create New Migration

```bash
task db:migrate:create -- add_new_feature
```

This creates:

- `server/migrations/NNNNNN_add_new_feature.up.sql`
- `server/migrations/NNNNNN_add_new_feature.down.sql`

### Migration Best Practices

1. **Always provide down migrations** for rollback capability
2. **Test migrations** on a copy of production data
3. **Keep migrations small** - one change per migration
4. **Never modify** existing migrations after deployment

## Schema Overview

```mermaid
erDiagram
    teams ||--o{ users : has
    teams ||--o{ rules : has
    teams ||--o{ agents : has
    users ||--o{ change_events : creates
    rules ||--o{ change_events : triggers
    agents ||--o{ change_events : reports
    users ||--o{ approvals : creates
    change_events ||--o{ approvals : has
    roles ||--o{ users : assigned_to

    teams {
        uuid id PK
        string name
        timestamp created_at
    }

    users {
        uuid id PK
        uuid team_id FK
        string email
        string name
        string auth_provider
        uuid role_id FK
    }

    rules {
        uuid id PK
        uuid team_id FK
        string name
        text content
        string enforcement_mode
        jsonb triggers
    }

    agents {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        string hostname
        timestamp last_seen
    }

    change_events {
        uuid id PK
        uuid rule_id FK
        uuid agent_id FK
        uuid user_id FK
        string event_type
        text old_content
        text new_content
        timestamp created_at
    }

    roles {
        uuid id PK
        string name
        jsonb permissions
    }

    approvals {
        uuid id PK
        uuid change_id FK
        uuid user_id FK
        string status
        timestamp created_at
    }
```

## Backup and Restore

### Backup

```bash
# Using pg_dump
pg_dump -h localhost -U edictflow -d edictflow -F c -f backup.dump

# Using Docker
docker exec edictflow-db pg_dump -U edictflow edictflow > backup.sql
```

### Automated Backups

Create a backup script `/opt/edictflow/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR=/var/backups/edictflow
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup
pg_dump -h localhost -U edictflow -d edictflow -F c -f "$BACKUP_DIR/backup_$DATE.dump"

# Remove old backups
find "$BACKUP_DIR" -name "backup_*.dump" -mtime +$RETENTION_DAYS -delete
```

Schedule with cron:

```bash
0 2 * * * /opt/edictflow/backup.sh
```

### Restore

```bash
# Using pg_restore
pg_restore -h localhost -U edictflow -d edictflow -c backup.dump

# Using Docker
docker exec -i edictflow-db psql -U edictflow edictflow < backup.sql
```

## Performance Tuning

### Connection Pooling

Configure connection pool in the server:

```bash
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

### PostgreSQL Configuration

For production, tune `postgresql.conf`:

```ini
# Memory
shared_buffers = 256MB
effective_cache_size = 768MB
work_mem = 16MB

# Connections
max_connections = 100

# Write-Ahead Log
wal_buffers = 16MB
checkpoint_completion_target = 0.9

# Query Planning
random_page_cost = 1.1
effective_io_concurrency = 200
```

### Indexes

Key indexes are created by migrations. For custom queries, add indexes:

```sql
-- Example: Speed up change event queries
CREATE INDEX CONCURRENTLY idx_change_events_team_created
ON change_events (team_id, created_at DESC);
```

## Monitoring

### Connection Status

```sql
SELECT state, count(*)
FROM pg_stat_activity
WHERE datname = 'edictflow'
GROUP BY state;
```

### Table Sizes

```sql
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(tablename::regclass)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(tablename::regclass) DESC;
```

### Slow Queries

Enable slow query logging:

```ini
log_min_duration_statement = 1000  # Log queries > 1 second
```

## Troubleshooting

### Connection Refused

```bash
# Check PostgreSQL is running
systemctl status postgresql

# Check port is listening
ss -tlnp | grep 5432

# Check pg_hba.conf allows connection
```

### Authentication Failed

```bash
# Verify user exists
sudo -u postgres psql -c "SELECT usename FROM pg_user"

# Reset password
sudo -u postgres psql -c "ALTER USER edictflow PASSWORD 'new-password'"
```

### Migration Failed

```bash
# Check current version
migrate -path server/migrations -database "$DATABASE_URL" version

# Fix dirty state
migrate -path server/migrations -database "$DATABASE_URL" force VERSION

# Then retry
migrate -path server/migrations -database "$DATABASE_URL" up
```
