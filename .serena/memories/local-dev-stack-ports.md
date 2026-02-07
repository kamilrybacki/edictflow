# Local Dev Stack Port Configuration

When spinning up the local test stack, use dynamically configurable ports via environment variables.

## Default Ports (docker-compose.yml)

| Service | Env Variable | Default Port | Description |
|---------|--------------|--------------|-------------|
| PostgreSQL | `DB_PORT` | 5432 | Database |
| Redis | `REDIS_PORT` | 6380 | Cache/Pub-Sub |
| Master API | `API_PORT` | 8080 | Main API server |
| Worker | `WORKER_PORT` | 8081 | WebSocket/Agent handler |
| Web UI | `WEB_PORT` | 3000 | Next.js frontend |
| Legacy Server | `LEGACY_SERVER_PORT` | 8082 | Old monolith (legacy profile) |

## Usage

```bash
# Use default ports
docker-compose up -d

# Use custom ports (e.g., if defaults are in use)
API_PORT=9080 WORKER_PORT=9081 WEB_PORT=4000 docker-compose up -d
```

## Important Notes

1. **Worker port must be exposed** - The web UI fetches `/health` from the worker directly
2. **Single worker for dev** - Multiple workers would conflict on the host port; scaling is for production only
3. **NEXT_PUBLIC_* vars are build-time** - Web container needs rebuild if ports change significantly
4. **Always run migrations after fresh start**: `task dev:migrate`
5. **Seed test data**: `cat tests/infrastructure/seed-data.sql | docker exec -i edictflow-db-1 psql -U edictflow -d edictflow`

## Test Credentials

- `admin@test.local` / `Test1234` (Admin role)
- `user@test.local` / `Test1234` (Member role)
- `agent[1-5]@test.local` / `Test1234` (Auto-connected agents)
