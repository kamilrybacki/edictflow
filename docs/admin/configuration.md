# Configuration

Configure Claudeception server using environment variables.

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string (master only) | `postgres://user:pass@host:5432/db?sslmode=disable` |
| `JWT_SECRET` | Secret for signing JWT tokens (min 32 chars) | `your-256-bit-secret-key-here` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379/0` |

### Master Process

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `SERVER_HOST` | `0.0.0.0` | HTTP server bind address |
| `READ_TIMEOUT` | `30s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

### Worker Process

| Variable | Default | Description |
|----------|---------|-------------|
| `WORKER_PORT` | `8081` | Worker WebSocket port |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection for pub/sub |
| `JWT_SECRET` | - | Must match master's JWT secret |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_EXPIRY` | `24h` | JWT token expiration |
| `REFRESH_EXPIRY` | `168h` | Refresh token expiration (7 days) |
| `OAUTH_GITHUB_CLIENT_ID` | - | GitHub OAuth client ID |
| `OAUTH_GITHUB_CLIENT_SECRET` | - | GitHub OAuth client secret |
| `OAUTH_GOOGLE_CLIENT_ID` | - | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | - | Google OAuth client secret |

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open database connections |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle database connections |
| `DB_CONN_MAX_LIFETIME` | `5m` | Maximum connection lifetime |

### Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |
| `REDIS_POOL_SIZE` | `10` | Connection pool size |
| `REDIS_READ_TIMEOUT` | `3s` | Redis read timeout |
| `REDIS_WRITE_TIMEOUT` | `3s` | Redis write timeout |

### WebSocket

| Variable | Default | Description |
|----------|---------|-------------|
| `WS_PING_INTERVAL` | `30s` | WebSocket ping interval |
| `WS_PONG_TIMEOUT` | `60s` | WebSocket pong timeout |
| `WS_WRITE_TIMEOUT` | `10s` | WebSocket write timeout |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, text) |

### Splunk Metrics

Claudeception can send metrics to Splunk HTTP Event Collector (HEC) for observability.

| Variable | Default | Description |
|----------|---------|-------------|
| `SPLUNK_ENABLED` | `false` | Enable Splunk metrics |
| `SPLUNK_HEC_URL` | - | HEC endpoint URL (e.g., `http://splunk:8088/services/collector/event`) |
| `SPLUNK_HEC_TOKEN` | - | HEC authentication token |
| `SPLUNK_SOURCE` | `claudeception` | Event source identifier |
| `SPLUNK_SOURCETYPE` | `claudeception:metrics` | Event sourcetype for indexing |
| `SPLUNK_INDEX` | `main` | Target Splunk index |
| `SPLUNK_SKIP_TLS_VERIFY` | `false` | Skip TLS certificate verification (dev only) |

**Metrics collected:**

*API Layer:*
- **api_request**: Method, path, status code, duration (ms), user ID
- **api_error**: Error types and affected endpoints

*Agent-Server Communication:*
- **agent_connection**: Agent ID, team ID, action (connected/disconnected)
- **websocket_message**: Direction (inbound/outbound), message type, agent ID, size (bytes)
- **broadcast**: Team ID, event type, recipient count
- **hub_stats**: Connected agents, teams, active subscriptions (periodic)

*Master-Worker Communication:*
- **redis_publish**: Channel, message type, success, latency (ms)
- **redis_subscription**: Channel, action (subscribe/unsubscribe)
- **worker_heartbeat**: Worker ID, agent count, team count (periodic)

*Infrastructure Health:*
- **health_check**: Component (postgres/redis), status (healthy/unhealthy), latency (ms)
- **db_pool_stats**: Total connections, acquired, idle, max (periodic)
- **db_query**: Operation type, table, duration (ms), success/failure

**Example Splunk configuration:**

```bash
SPLUNK_ENABLED=true
SPLUNK_HEC_URL=https://splunk.internal:8088/services/collector/event
SPLUNK_HEC_TOKEN=your-hec-token
SPLUNK_SOURCE=claudeception-prod
SPLUNK_SOURCETYPE=claudeception:metrics
SPLUNK_INDEX=observability
```

### AppDynamics RUM (Frontend)

The web UI supports AppDynamics Real User Monitoring for frontend observability.

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_APPDYNAMICS_ENABLED` | `false` | Enable AppDynamics RUM |
| `NEXT_PUBLIC_APPDYNAMICS_APP_KEY` | - | AppDynamics application key |
| `NEXT_PUBLIC_APPDYNAMICS_ADR_URL` | - | Self-hosted agent JavaScript URL |
| `NEXT_PUBLIC_APPDYNAMICS_BEACON_URL` | - | Beacon collector endpoint URL |

**Example AppDynamics configuration:**

```bash
NEXT_PUBLIC_APPDYNAMICS_ENABLED=true
NEXT_PUBLIC_APPDYNAMICS_APP_KEY=your-app-key
NEXT_PUBLIC_APPDYNAMICS_ADR_URL=https://cdn.appdynamics.com/adrum
NEXT_PUBLIC_APPDYNAMICS_BEACON_URL=https://col.eum-appdynamics.com
```

## Configuration File

Alternatively, use a configuration file at `/etc/claudeception/config.yaml`:

```yaml
server:
  port: 8080
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s

database:
  url: postgres://user:pass@host:5432/db
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

auth:
  jwt_secret: your-secret-key
  jwt_expiry: 24h
  refresh_expiry: 168h

oauth:
  github:
    client_id: your-client-id
    client_secret: your-client-secret
  google:
    client_id: your-client-id
    client_secret: your-client-secret

websocket:
  ping_interval: 30s
  pong_timeout: 60s

logging:
  level: info
  format: json
```

Environment variables take precedence over config file values.

## JWT Secret Generation

Generate a secure JWT secret:

```bash
# Using OpenSSL
openssl rand -base64 32

# Using /dev/urandom
head -c 32 /dev/urandom | base64
```

!!! danger "Security"
    Never commit JWT secrets to version control. Use environment variables or secrets management.

## OAuth Configuration

### GitHub OAuth

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Create a new OAuth App:
   - **Homepage URL**: `https://app.yourdomain.com`
   - **Authorization callback URL**: `https://api.yourdomain.com/auth/github/callback`
3. Copy Client ID and Client Secret
4. Set environment variables:
   ```bash
   OAUTH_GITHUB_CLIENT_ID=your-client-id
   OAUTH_GITHUB_CLIENT_SECRET=your-client-secret
   ```

### Google OAuth

1. Go to Google Cloud Console → APIs & Services → Credentials
2. Create OAuth 2.0 Client ID:
   - **Application type**: Web application
   - **Authorized redirect URIs**: `https://api.yourdomain.com/auth/google/callback`
3. Copy Client ID and Client Secret
4. Set environment variables:
   ```bash
   OAUTH_GOOGLE_CLIENT_ID=your-client-id
   OAUTH_GOOGLE_CLIENT_SECRET=your-client-secret
   ```

### Custom OIDC

For custom OpenID Connect providers:

1. Configure your OIDC provider with:
   - **Redirect URI**: `https://api.yourdomain.com/auth/oidc/callback`
   - **Scopes**: `openid`, `email`, `profile`
2. Set environment variables:
   ```bash
   OAUTH_OIDC_ISSUER=https://your-idp.example.com
   OAUTH_OIDC_CLIENT_ID=your-client-id
   OAUTH_OIDC_CLIENT_SECRET=your-client-secret
   ```

Supported OIDC providers include:

- Okta
- Auth0
- Keycloak
- Azure AD
- Any OIDC-compliant provider

## Device Code Flow (Agent CLI)

For agent CLI authentication using device code flow:

### GitHub

Enable device flow in your OAuth App settings.

### Google

Device code flow is enabled by default for OAuth applications.

## CORS Configuration

Configure CORS for the Web UI:

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ORIGINS` | `*` | Allowed origins (comma-separated) |
| `CORS_METHODS` | `GET,POST,PUT,PATCH,DELETE` | Allowed methods |
| `CORS_HEADERS` | `Authorization,Content-Type` | Allowed headers |

For production:

```bash
CORS_ORIGINS=https://app.yourdomain.com,https://admin.yourdomain.com
```

## Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_REQUESTS` | `100` | Requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window duration |
| `RATE_LIMIT_BURST` | `10` | Burst allowance |

## Example Configurations

### Development (Combined Server)

```bash
DATABASE_URL=postgres://claudeception:claudeception@localhost:5432/claudeception?sslmode=disable
REDIS_URL=redis://localhost:6379/0
JWT_SECRET=dev-secret-change-in-production
SERVER_PORT=8080
LOG_LEVEL=debug
LOG_FORMAT=text
```

### Production (Master)

```bash
DATABASE_URL=postgres://user:${DB_PASSWORD}@db.internal:5432/claudeception?sslmode=require
REDIS_URL=redis://redis.internal:6379/0
JWT_SECRET=${JWT_SECRET}
SERVER_PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json
CORS_ORIGINS=https://app.yourdomain.com
OAUTH_GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
OAUTH_GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

### Production (Worker)

```bash
REDIS_URL=redis://redis.internal:6379/0
JWT_SECRET=${JWT_SECRET}
WORKER_PORT=8081
LOG_LEVEL=info
LOG_FORMAT=json
```

## Validation

On startup, the server validates configuration:

- `DATABASE_URL` is set and valid
- `JWT_SECRET` is at least 32 characters
- Port numbers are valid
- Durations are parseable

Invalid configuration causes the server to exit with an error message.
