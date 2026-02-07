# Performance Optimizations

This document outlines the performance optimizations implemented in Edictflow to ensure fast, responsive operation at scale.

## Backend (Go Server)

### Database Layer

#### Connection Pool Configuration

The PostgreSQL connection pool is configured with sensible defaults for production workloads:

```go
// server/adapters/postgres/postgres.go
type PoolConfig struct {
    MaxConns          int32         // 25 - Maximum connections
    MinConns          int32         // 5  - Minimum idle connections
    MaxConnLifetime   time.Duration // 1 hour
    MaxConnIdleTime   time.Duration // 30 minutes
    HealthCheckPeriod time.Duration // 1 minute
}
```

#### Batch Queries (N+1 Prevention)

Batch lookup methods prevent N+1 query patterns:

```go
// server/adapters/postgres/user_db.go
func (db *UserDB) GetByIDs(ctx context.Context, ids []string) (map[string]domain.User, error) {
    rows, err := db.pool.Query(ctx, `
        SELECT id, email, name, ... FROM users WHERE id = ANY($1)
    `, ids)
    // ...
}
```

#### Slice Preallocation

All database query results use preallocated slices to reduce allocations:

```go
users := make([]domain.User, 0, 32) // Preallocate with reasonable capacity
```

Implemented in:
- `user_db.go`
- `rule_db.go`
- `notification_db.go`
- `audit_db.go`
- `team_db.go`
- `role_db.go`

#### Database Indexes

Key indexes are created for frequently queried columns:

```sql
-- agent/storage/migrations.go
CREATE INDEX IF NOT EXISTS idx_cached_rules_layer ON cached_rules(target_layer);
CREATE INDEX IF NOT EXISTS idx_message_queue_attempts ON message_queue(attempts);
CREATE INDEX IF NOT EXISTS idx_pending_changes_status ON pending_changes(status);
```

### Caching

#### Permission Cache

In-memory permission caching with TTL reduces database load for authorization checks:

```go
// server/entrypoints/api/middleware/permission.go
type Permission struct {
    cache    map[string]permissionCacheEntry
    cacheMu  sync.RWMutex
    cacheTTL time.Duration // 5 minutes
}
```

#### Response Cache

HTTP response caching uses FNV-1a hashing for fast, non-cryptographic cache key generation:

```go
// server/entrypoints/api/middleware/cache.go
h := fnv.New64a()
h.Write([]byte(strings.Join(parts, "|")))
return c.config.KeyPrefix + ":" + hex.EncodeToString(h.Sum(nil))
```

### Concurrency

#### Worker Pool

Bounded worker pools prevent unbounded goroutine growth for async operations:

```go
// server/common/workerpool/pool.go
type Pool struct {
    tasks   chan Task
    workers int
}

var DefaultAuditPool = New(4, 1000)  // Audit logging
var DefaultEventPool = New(4, 1000)  // Event publishing
var DefaultCachePool = New(2, 500)   // Cache writes
```

#### Lock Contention Reduction

WebSocket hub broadcasts copy client lists before iteration to reduce lock hold time:

```go
// server/entrypoints/ws/hub.go
func (h *Hub) BroadcastToAll(data []byte) {
    h.mu.RLock()
    clients := make([]*Client, 0, len(h.clients))
    for _, client := range h.clients {
        clients = append(clients, client)
    }
    h.mu.RUnlock()

    // Send without holding lock
    for _, client := range clients {
        select {
        case client.Send <- data:
        default:
        }
    }
}
```

### Rate Limiting

Atomic Lua script for sliding window rate limiting:

```go
// server/entrypoints/api/middleware/ratelimit.go
var rateLimitScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local max_requests = tonumber(ARGV[3])

redis.call('ZREMRANGEBYSCORE', key, '0', tostring(now - window))
local count = redis.call('ZCARD', key)

if count < max_requests then
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window + 1000)
end

return {count < max_requests and 1 or 0, count}
`)
```

### WebSocket Optimization

Larger buffer sizes for better throughput:

```go
// server/entrypoints/ws/handler.go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  4096,  // Increased from 1024
    WriteBufferSize: 4096,  // Increased from 1024
}
```

---

## Agent (Go CLI)

### File Hashing

Streaming file hashing to avoid loading entire files into memory:

```go
// agent/watcher/watcher.go
func hashFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}
```

### HTTP Connection Pooling

Shared HTTP client with connection pooling:

```go
// agent/auth/http.go
var sharedHTTPClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### Context-Aware Operations

Graceful shutdown support with context-aware goroutines:

```go
// agent/ws/client.go
func (c *Client) ConnectWithContext(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        // Connection logic with proper backoff
    }
}
```

### Async Notifications

Non-blocking desktop notifications:

```go
// agent/notify/notify.go
func notifyAsync(title, message string) {
    go func() {
        if err := beeep.Notify(title, message, ""); err != nil {
            log.Printf("Notification failed: %v", err)
        }
    }()
}
```

### Singleton Storage

Shared storage instance to avoid multiple database connections:

```go
// agent/storage/storage.go
var (
    instance     *Storage
    instanceOnce sync.Once
    instanceErr  error
)

func GetShared() (*Storage, error) {
    instanceOnce.Do(func() {
        instance, instanceErr = New()
    })
    return instance, instanceErr
}
```

---

## Frontend (Next.js/React)

### React.memo

Frequently re-rendered components are memoized to prevent unnecessary re-renders:

| Component | Location |
|-----------|----------|
| `RuleCard` | `components/dashboard/RuleCard.tsx` |
| `StatCard` | `components/dashboard/StatCard.tsx` |
| `TeamCard` | `components/dashboard/TeamCard.tsx` |
| `ActivityFeed` | `components/dashboard/ActivityFeed.tsx` |
| `RuleHierarchy` | `components/dashboard/RuleHierarchy.tsx` |
| `ChangeRequestTable` | `components/ChangeRequestTable.tsx` |
| `ChannelList` | `components/ChannelList.tsx` |
| `DiffViewer` | `components/DiffViewer.tsx` |

### Code Splitting

Heavy modal components are lazy-loaded with `next/dynamic`:

```typescript
// app/page.tsx
const RuleEditor = dynamic(
  () => import('@/components/RuleEditor').then(mod => ({ default: mod.RuleEditor })),
  { loading: () => <Spinner />, ssr: false }
);

const AgentListModal = dynamic(
  () => import('@/components/dashboard/AgentListModal').then(mod => ({ default: mod.AgentListModal })),
  { ssr: false }
);

const RuleHistoryPanel = dynamic(
  () => import('@/components/dashboard/RuleHistoryPanel').then(mod => ({ default: mod.RuleHistoryPanel })),
  { ssr: false }
);
```

### useMemo / useCallback

Expensive computations and callback functions are memoized:

```typescript
// app/page.tsx
const filteredRules = useMemo(
  () => selectedLayer ? rules.filter(r => r.targetLayer === selectedLayer) : rules,
  [rules, selectedLayer]
);

const { pendingApprovals, activeRules, blockedRules } = useMemo(() => ({
  pendingApprovals: rules.filter(r => r.status === 'pending').length,
  activeRules: rules.filter(r => r.status === 'approved').length,
  blockedRules: rules.filter(r => r.enforcementMode === 'block').length,
}), [rules]);

const highlightRule = useCallback((ruleId: string) => {
  setHighlightedRuleId(ruleId);
  setTimeout(() => setHighlightedRuleId(undefined), 1500);
}, []);
```

### Memoized List Items

List components extract items as separate memoized components with `useCallback` handlers:

```typescript
// components/ChangeRequestTable.tsx
const ChangeRequestRow = memo(function ChangeRequestRow({ change, onApprove, onReject }) {
  const handleApprove = useCallback(() => onApprove?.(change.id), [onApprove, change.id]);
  const handleReject = useCallback(() => onReject?.(change.id), [onReject, change.id]);
  // ...
});
```

### Expensive Operations

Expensive operations are memoized to run only when inputs change:

```typescript
// components/DiffViewer.tsx
const parsedLines = useMemo(() => parseDiff(diff), [diff]);
```

---

## Performance Monitoring

### Recommended Tools

| Tool | Purpose |
|------|---------|
| `pprof` | Go CPU/memory profiling |
| `go tool trace` | Go execution tracing |
| React DevTools Profiler | React render profiling |
| Lighthouse | Web performance auditing |
| `wrk` / `hey` | HTTP load testing |

### Key Metrics

- **API Response Time:** < 100ms p95
- **Database Query Time:** < 10ms p95
- **WebSocket Latency:** < 50ms
- **Frontend TTI:** < 2s
- **Frontend FCP:** < 1s

---

## Best Practices

### Backend

1. Always use batch queries when fetching related entities
2. Preallocate slices with reasonable capacity
3. Use worker pools for async operations
4. Cache frequently accessed data with appropriate TTL
5. Use context for cancellation and timeouts

### Frontend

1. Wrap list item components with `React.memo`
2. Use `useMemo` for expensive computations
3. Use `useCallback` for callbacks passed to child components
4. Lazy load modals and heavy components with `next/dynamic`
5. Avoid inline object/array creation in JSX props
