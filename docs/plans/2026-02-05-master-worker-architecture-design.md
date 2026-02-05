# Master-Worker Architecture Design

## Overview

This document describes the architectural redesign of Claudeception to support horizontal scaling through a master-worker architecture with Redis coordination.

## Goals

- Enable horizontal scaling of both API and WebSocket layers
- Ensure high availability through stateless components
- Maintain low-latency rule propagation to agents
- Provide graceful degradation during partial failures

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Load Balancer                                │
│                    (e.g., nginx, Traefik, ALB)                       │
└──────────────┬────────────────────────────┬─────────────────────────┘
               │ HTTP/REST                  │ WebSocket
               ▼                            ▼
┌──────────────────────────┐    ┌──────────────────────────────────────┐
│      Master Servers      │    │           Sync Workers               │
│  (stateless, N replicas) │    │      (stateless, N replicas)         │
│                          │    │                                      │
│  • REST API (Chi)        │    │  • WebSocket connections (agents)    │
│  • Web UI backend        │    │  • Subscribe to Redis team channels  │
│  • CLI authentication    │    │  • Push updates to connected agents  │
│  • Rule CRUD operations  │    │  • Report agent health               │
└────────────┬─────────────┘    └──────────────┬───────────────────────┘
             │                                  │
             │         ┌────────────────┐       │
             └────────►│     Redis      │◄──────┘
                       │                │
                       │  • Pub/Sub     │
                       │  • Cache       │
                       └───────┬────────┘
                               │
                       ┌───────▼────────┐
                       │   PostgreSQL   │
                       │ (source of     │
                       │  truth)        │
                       └────────────────┘
```

**Key points:**

- Masters and workers are separate deployable units
- Both are stateless - scale horizontally by adding replicas
- Redis coordinates communication between masters and workers
- PostgreSQL remains the single source of truth

## Component Responsibilities

### Master Servers

Masters handle all **control plane** operations.

**API Endpoints (existing, unchanged):**

- `/api/rules/*` - CRUD for rules
- `/api/users/*` - User management
- `/api/teams/*` - Team management
- `/api/auth/*` - Authentication (JWT)
- `/api/approvals/*` - Change approvals

**New behavior on write operations:**

When a rule is created/updated/deleted:

1. Persist to PostgreSQL (as today)
2. Publish to Redis channel `team:{team_id}:rules` with payload:
   ```json
   {
     "event": "rule_updated",
     "team_id": "123",
     "rule_id": "456",
     "version": 42,
     "timestamp": "2026-02-05T10:00:00Z"
   }
   ```
3. Invalidate Redis cache for that team's rules

**What masters don't do:**

- No WebSocket handling (moved to workers)
- No direct agent communication

### Sync Workers

Workers handle all **data plane** operations.

**WebSocket connection lifecycle:**

1. **Agent connects** → Worker authenticates JWT, extracts team ID
2. **Subscribe to Redis** → Worker subscribes to `team:{team_id}:rules` (if not already subscribed for another agent on same team)
3. **Initial sync** → Worker fetches current rules from Redis cache (or PostgreSQL fallback), sends to agent
4. **Ongoing** → When Redis publishes update, worker pushes to all connected agents for that team
5. **Agent disconnects** → Unsubscribe from team channel if no other agents on that team

**Worker internal state (in-memory only):**

```go
type Worker struct {
    // Map: team_id -> set of connected agent connections
    teamAgents map[string]map[*AgentConn]struct{}

    // Map: agent_id -> connection (for targeted messages)
    agents map[string]*AgentConn

    // Redis pub/sub subscriptions
    subscriptions map[string]*redis.PubSub  // team_id -> subscription
}
```

**Health & metrics:**

- Expose `/health` endpoint for load balancer
- Report connected agent count to Redis (for observability)
- Prometheus metrics: connections, messages sent, latency

## Redis Channel Design

### Channel Naming Convention

| Channel | Purpose |
|---------|---------|
| `team:{id}:rules` | Rule changes for a team |
| `team:{id}:categories` | Category changes for a team |
| `broadcast:all` | System-wide announcements (maintenance, version updates) |
| `agent:{id}:direct` | Targeted messages to specific agent (approval responses) |

### Message Flow Example

Admin updates a rule:

```
1. Admin → Web UI → Master API
   POST /api/rules/456 { "content": "..." }

2. Master → PostgreSQL
   UPDATE rules SET content = '...' WHERE id = 456

3. Master → Redis PUBLISH
   Channel: team:123:rules
   Payload: { "event": "rule_updated", "rule_id": "456", "version": 43 }

4. Redis → All subscribed Workers
   (Workers subscribed to team:123:rules receive message)

5. Worker → Connected Agents
   Worker looks up teamAgents["123"], sends update to each
```

### Message Payload Format

```json
{
  "event": "rule_updated|rule_deleted|category_updated|sync_required",
  "entity_id": "456",
  "version": 43,
  "timestamp": "2026-02-05T10:00:00Z"
}
```

Workers fetch full data from cache/DB on `sync_required`, otherwise use event for incremental updates.

## Redis Caching Strategy

### What to Cache

| Key Pattern | Data | TTL |
|-------------|------|-----|
| `cache:team:{id}:rules` | Serialized rules for team | 5 min |
| `cache:team:{id}:categories` | Categories for team | 5 min |
| `cache:rule:{id}` | Single rule (for targeted fetches) | 5 min |

### Cache Invalidation

On write operations, Master does:

1. Write to PostgreSQL
2. Delete cache key (`DEL cache:team:{id}:rules`)
3. Publish event to Redis channel

Workers on cache miss:

1. Check Redis cache first
2. If miss → fetch from PostgreSQL
3. Populate cache with result

### Rationale for Delete-on-Write

- Simpler - avoids race conditions between cache and DB
- Workers lazy-load on next request
- Short TTL means stale data window is small

### Cache Sizing Estimate

- 100 teams × ~50KB rules each = ~5MB
- Fits comfortably in smallest Redis instance

## Failure Handling & Recovery

### Worker Crashes

- Load balancer detects failed health check, stops routing
- Agents experience disconnect, reconnect to any healthy worker
- New worker subscribes to agent's team channel, sends full sync
- No data loss - PostgreSQL is source of truth

### Master Crashes

- Load balancer routes to other masters
- No impact on connected agents (workers still running)
- API requests continue on healthy masters

### Redis Crashes

- Workers lose pub/sub subscriptions
- Workers detect disconnect, enter degraded mode:
  - Continue serving connected agents
  - Poll PostgreSQL on interval (fallback, 30s)
- On Redis recovery, workers resubscribe
- Consider Redis Sentinel or Redis Cluster for HA

### PostgreSQL Crashes

- Masters return 503 for write operations
- Workers continue serving cached data
- Agents remain connected, receive no new updates
- On recovery, system resumes normally

### Network Partition (Master Can't Reach Redis)

- Master writes to PostgreSQL succeed
- Publish to Redis fails → log error, continue
- Workers don't see update immediately
- Cache TTL expiration eventually causes refresh

## Deployment

### Container Structure

```
claudeception/
├── server/
│   ├── cmd/
│   │   ├── master/main.go    # New entrypoint
│   │   └── worker/main.go    # New entrypoint
```

Both share domain, adapters, and services - only entrypoints differ.

### Docker Compose (Development)

```yaml
services:
  master:
    build: { context: ./server, target: master }
    replicas: 2
    ports: ["8080:8080"]
    depends_on: [postgres, redis]

  worker:
    build: { context: ./server, target: worker }
    replicas: 3
    ports: ["8081:8081"]
    depends_on: [postgres, redis]

  redis:
    image: redis:7-alpine

  postgres:
    image: postgres:16-alpine
```

### Scaling Guidelines

| Component | Scale trigger | Typical ratio |
|-----------|--------------|---------------|
| Master | API request rate, CPU | 2-4 replicas |
| Worker | Connected agents, memory | 1 worker per ~5,000 agents |
| Redis | Pub/sub throughput | Single instance to start, Cluster if >100k msg/s |

### Load Balancer Configuration

- Masters: Round-robin, `/health` checks
- Workers: Round-robin for WebSocket upgrades (no sticky sessions needed)

## Migration Path

### Phase 1: Add Redis (No Breaking Changes)

- Add Redis to docker-compose
- Add go-redis dependency
- Implement Redis cache layer in adapters
- Masters publish events but hub still handles WebSocket
- Deploy and validate caching works

### Phase 2: Extract Worker Entrypoint

- Create `cmd/worker/main.go` with WebSocket hub + Redis subscriber
- Create `cmd/master/main.go` without WebSocket (API only)
- Both run side-by-side, sharing same PostgreSQL
- Worker subscribes to Redis channels
- Test with single worker instance

### Phase 3: Switch Traffic

- Route WebSocket connections to workers via load balancer
- Route HTTP/API to masters
- Remove WebSocket code from master
- Validate agents reconnect and sync correctly

### Phase 4: Scale Out

- Add replicas of both master and worker
- Load test to establish baseline capacity
- Set up monitoring and alerting

### Rollback Plan

Each phase can be rolled back independently. Phase 1-2 are additive. Phase 3 rollback: re-enable WebSocket on master, route traffic back.
