# Architecture

Edictflow follows a hub-and-spoke architecture with a central server and distributed agents, designed for horizontal scalability.

## System Overview

```mermaid
graph TB
    subgraph "Control Plane"
        WebUI[Web UI<br/>React/Next.js]
        Master[Master API<br/>Go/Chi]
        Auth[Auth Service<br/>JWT + OAuth]
    end

    subgraph "Message Plane"
        Redis[(Redis<br/>Pub/Sub)]
    end

    subgraph "Worker Plane"
        W1[Worker 1<br/>WebSocket]
        W2[Worker 2<br/>WebSocket]
        WN[Worker N<br/>WebSocket]
    end

    subgraph "Data Plane"
        PG[(PostgreSQL<br/>Primary Store)]
    end

    subgraph "Agent Plane"
        A1[Agent 1<br/>SQLite Cache]
        A2[Agent 2<br/>SQLite Cache]
        AN[Agent N<br/>SQLite Cache]
    end

    WebUI --> Master
    Master --> Auth
    Master --> PG
    Master -->|Publish Events| Redis

    Redis -->|Subscribe| W1
    Redis -->|Subscribe| W2
    Redis -->|Subscribe| WN

    A1 <-->|WSS| W1
    A2 <-->|WSS| W2
    AN <-->|WSS| WN
```

## Master-Worker Architecture

Edictflow uses a master-worker architecture for horizontal scalability:

| Process | Responsibility | Stateless | Scales To |
|---------|----------------|-----------|-----------|
| **Master** | REST API, business logic, database operations | Yes | N instances |
| **Worker** | WebSocket connections, real-time updates | Yes | N instances |
| **Redis** | Event coordination, pub/sub | N/A | Cluster mode |

### Masters

Masters handle:
- All REST API requests
- Authentication and authorization
- Database read/write operations
- Publishing events to Redis on data changes

Masters are stateless and can be load-balanced with any strategy.

### Workers

Workers handle:
- WebSocket connections from agents
- Subscribing to Redis channels for team events
- Broadcasting updates to connected agents
- Health monitoring of agent connections

Workers are stateless - agents can connect to any worker and receive events for their team.

### Redis Coordination

Redis provides:
- Pub/sub channels for team-specific events
- Channel naming: `team:{team_id}:rules`, `team:{team_id}:categories`
- Broadcast channel for global events: `broadcast:all`
- Direct agent messaging: `agent:{agent_id}:direct`

## Components

### Server (Master Process)

The master process handles API operations:

| Component | Technology | Purpose |
|-----------|------------|---------|
| REST API | Go + Chi | CRUD operations for rules, users, teams |
| Auth Service | JWT + OAuth 2.0 | Authentication and authorization |
| Database | PostgreSQL | Persistent storage |
| Publisher | go-redis | Event publishing to Redis |

### Server (Worker Process)

The worker process handles real-time communication:

| Component | Technology | Purpose |
|-----------|------------|---------|
| WebSocket Hub | gorilla/websocket | Manage agent connections |
| Redis Subscriber | go-redis | Subscribe to team channels |
| Broadcaster | Go channels | Fan-out events to agents |

### Agent

The agent is a lightweight daemon that:

| Component | Technology | Purpose |
|-----------|------------|---------|
| CLI | Go + Cobra | User interaction |
| Daemon | Go | Background file monitoring |
| File Watcher | fsnotify | Detect file changes |
| Local Cache | SQLite | Offline operation support |
| WebSocket Client | gorilla/websocket | Server communication |

### Web UI

The web interface provides:

| Feature | Technology | Purpose |
|---------|------------|---------|
| Dashboard | React/Next.js | Overview and quick actions |
| Rule Editor | Monaco Editor | Edit CLAUDE.md content |
| User Management | React | Manage users and roles |
| Audit Log | React | View change history |
| Graph View | React Flow | Visualize organization hierarchy |
| Command Palette | React | Quick navigation (`Ctrl+K`) |

## Data Flow

### Rule Creation (with Master-Worker)

```mermaid
sequenceDiagram
    participant Admin as Admin
    participant Master as Master API
    participant DB as PostgreSQL
    participant Redis as Redis
    participant Worker as Workers
    participant Agent as Agents

    Admin->>Master: POST /api/v1/rules
    Master->>DB: Insert rule
    DB-->>Master: Rule created
    Master->>Redis: Publish rule_created event
    Master-->>Admin: 201 Created

    Redis-->>Worker: Event received
    Worker->>Agent: Push to team agents
    Agent->>Agent: Cache rule locally
```

### File Change Detection

```mermaid
sequenceDiagram
    participant User as Developer
    participant File as CLAUDE.md
    participant Agent as Agent
    participant Worker as Worker
    participant Redis as Redis
    participant Master as Master
    participant DB as PostgreSQL

    User->>File: Modify file
    File-->>Agent: Change detected (fsnotify)
    Agent->>Agent: Check enforcement mode

    alt Block Mode
        Agent->>File: Revert to cached content
        Agent->>Worker: Report change_blocked
    else Temporary Mode
        Agent->>Worker: Report change_detected
    else Warning Mode
        Agent->>Worker: Report change_flagged
    end

    Worker->>Redis: Publish event
    Redis-->>Master: Event received (optional)
    Master->>DB: Log to audit trail
```

## Enforcement Modes

Edictflow supports three enforcement modes:

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Block** | Immediately revert unauthorized changes | Production configurations |
| **Temporary** | Allow changes, flag for review | Development/testing |
| **Warning** | Log changes without intervention | Monitoring/gradual rollout |

## Security Model

### Authentication

```mermaid
graph LR
    subgraph "Web UI"
        Login[Login Form]
        JWT1[JWT Token]
    end

    subgraph "Agent CLI"
        Device[Device Code Flow]
        JWT2[JWT Token]
    end

    subgraph "Server"
        Auth[Auth Service]
        Verify[Token Verification]
    end

    Login --> Auth
    Auth --> JWT1
    JWT1 --> Verify

    Device --> Auth
    Auth --> JWT2
    JWT2 --> Verify
```

### Authorization (RBAC)

Permissions are hierarchical:

```
super_admin
├── manage_users
├── manage_teams
├── manage_roles
└── admin
    ├── manage_rules
    ├── approve_changes
    └── user
        ├── view_rules
        ├── request_changes
        └── view_changes
```

## Offline Operation

Agents maintain local SQLite caches for resilience:

```mermaid
graph TB
    subgraph "Agent"
        Daemon[Daemon]
        SQLite[(SQLite Cache)]
        Watcher[File Watcher]
    end

    subgraph "Worker"
        WS[WebSocket]
    end

    Daemon <-->|Online| WS
    Daemon <-->|Offline| SQLite
    Watcher --> Daemon
```

When offline:

1. Agent uses cached rules
2. Changes are queued locally
3. Sync occurs on reconnection

## Scalability

### Horizontal Scaling

```mermaid
graph TB
    LB[Load Balancer]
    WSLB[WebSocket LB<br/>Sticky Sessions]

    subgraph "Master Pool"
        M1[Master 1]
        M2[Master 2]
        M3[Master N]
    end

    subgraph "Worker Pool"
        W1[Worker 1]
        W2[Worker 2]
        W3[Worker N]
    end

    subgraph "Coordination"
        Redis[(Redis)]
    end

    subgraph "Database"
        PG[(PostgreSQL<br/>Primary)]
        PGR[(PostgreSQL<br/>Replica)]
    end

    LB --> M1
    LB --> M2
    LB --> M3

    WSLB --> W1
    WSLB --> W2
    WSLB --> W3

    M1 --> Redis
    M2 --> Redis
    M3 --> Redis

    W1 --> Redis
    W2 --> Redis
    W3 --> Redis

    M1 --> PG
    M2 --> PG
    M3 --> PG

    PG --> PGR
```

### Scaling Guidelines

| Component | Scaling Strategy | Notes |
|-----------|-----------------|-------|
| Masters | Add instances behind load balancer | Any LB strategy works |
| Workers | Add instances, use sticky sessions for WebSocket | Agents auto-reconnect |
| Redis | Use Redis Cluster for high availability | Built-in pub/sub support |
| PostgreSQL | Read replicas for scaling reads | Primary for writes |

## Technology Stack

| Layer | Technology | Why |
|-------|------------|-----|
| Master | Go | Performance, single binary, goroutines |
| Worker | Go | Efficient WebSocket handling |
| Router | Chi | Lightweight, idiomatic Go |
| Pub/Sub | Redis | Fast, reliable, built-in pub/sub |
| Database | PostgreSQL | Reliability, JSON support, migrations |
| Agent DB | SQLite | Zero-config, embedded, reliable |
| Web UI | Next.js 16 | React 19 ecosystem, SSR, fast development |
| Graph Visualization | React Flow | Interactive node-based diagrams |
| Styling | Tailwind CSS 4 | Utility-first CSS framework |
| Auth | JWT + OAuth 2.0 | Stateless, standard protocols |
| Real-time | WebSocket | Bidirectional, low latency |
| File Watch | fsnotify | Cross-platform, efficient |

## Directory Structure

```
edictflow/
├── server/           # Go server
│   ├── cmd/
│   │   ├── master/   # Master API entrypoint
│   │   ├── worker/   # Worker WebSocket entrypoint
│   │   └── server/   # Legacy combined server
│   ├── entrypoints/  # HTTP handlers
│   ├── services/     # Business logic
│   │   └── publisher/ # Redis event publisher
│   ├── worker/       # Worker hub and handler
│   ├── adapters/
│   │   ├── postgres/ # Database layer
│   │   └── redis/    # Redis client
│   ├── events/       # Event types
│   └── migrations/   # SQL migrations
├── agent/            # Go agent
│   ├── cmd/          # Entry points
│   ├── entrypoints/  # CLI commands
│   ├── daemon/       # Background service
│   └── watcher/      # File monitoring
├── web/              # Next.js frontend
│   ├── src/          # React components
│   └── public/       # Static assets
├── e2e/              # E2E tests
└── docs/             # Documentation
```
