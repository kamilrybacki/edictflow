# Master-Worker Architecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor server into separate master (API) and worker (WebSocket) processes coordinated via Redis pub/sub.

**Architecture:** Masters handle REST API requests, publish rule change events to Redis. Workers handle WebSocket connections, subscribe to team-specific Redis channels, and push updates to connected agents. Both are stateless and horizontally scalable.

**Tech Stack:** Go 1.24, go-redis/v9, Chi router, pgxpool, gorilla/websocket

---

## Phase 1: Add Redis Infrastructure

### Task 1.1: Add go-redis dependency

**Files:**
- Modify: `server/go.mod`

**Step 1: Add go-redis dependency**

```bash
cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go get github.com/redis/go-redis/v9
```

**Step 2: Verify dependency added**

Run: `grep "go-redis" server/go.mod`
Expected: Line containing `github.com/redis/go-redis/v9`

**Step 3: Commit**

```bash
git add server/go.mod server/go.sum
git commit -m "deps: add go-redis/v9 for pub/sub coordination"
```

---

### Task 1.2: Add Redis configuration

**Files:**
- Modify: `server/configurator/settings.go`

**Step 1: Write the test**

Create: `server/configurator/settings_test.go`

```go
package configurator

import (
	"os"
	"testing"
)

func TestLoadSettings_RedisURL(t *testing.T) {
	// Test default value
	os.Unsetenv("REDIS_URL")
	settings := LoadSettings()
	if settings.RedisURL != "redis://localhost:6379/0" {
		t.Errorf("expected default redis URL, got %s", settings.RedisURL)
	}

	// Test custom value
	os.Setenv("REDIS_URL", "redis://custom:6380/1")
	defer os.Unsetenv("REDIS_URL")
	settings = LoadSettings()
	if settings.RedisURL != "redis://custom:6380/1" {
		t.Errorf("expected custom redis URL, got %s", settings.RedisURL)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./configurator/... -v -run TestLoadSettings_RedisURL`
Expected: FAIL - `settings.RedisURL undefined`

**Step 3: Add RedisURL to Settings struct**

Modify `server/configurator/settings.go`:

```go
type Settings struct {
	DatabaseURL string
	RedisURL    string
	ServerPort  string
	JWTSecret   string
	BaseURL     string
}

func LoadSettings() Settings {
	port := getEnv("SERVER_PORT", "8080")
	return Settings{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/claudeception?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ServerPort:  port,
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:"+port),
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./configurator/... -v -run TestLoadSettings_RedisURL`
Expected: PASS

**Step 5: Commit**

```bash
git add server/configurator/
git commit -m "feat(config): add RedisURL setting with default"
```

---

### Task 1.3: Create Redis client adapter

**Files:**
- Create: `server/adapters/redis/client.go`
- Create: `server/adapters/redis/client_test.go`

**Step 1: Write the interface and client**

Create `server/adapters/redis/client.go`:

```go
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis operations for pub/sub and caching
type Client struct {
	rdb *redis.Client
}

// NewClient creates a Redis client from URL
func NewClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	return &Client{rdb: rdb}, nil
}

// Ping checks the Redis connection
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Publish sends a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message []byte) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

// Subscribe returns a subscription to channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// Set stores a value with optional TTL
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	return c.rdb.Get(ctx, key).Bytes()
}

// Del removes keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Underlying returns the underlying redis.Client for advanced use
func (c *Client) Underlying() *redis.Client {
	return c.rdb
}
```

**Step 2: Write integration test**

Create `server/adapters/redis/client_test.go`:

```go
package redis

import (
	"context"
	"testing"
	"time"
)

func TestClient_PubSub(t *testing.T) {
	// Skip if no Redis available
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Test pub/sub
	sub := client.Subscribe(ctx, "test-channel")
	defer sub.Close()

	// Wait for subscription to be ready
	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to receive subscription confirmation: %v", err)
	}

	// Publish
	go func() {
		time.Sleep(100 * time.Millisecond)
		client.Publish(ctx, "test-channel", []byte("hello"))
	}()

	// Receive
	ch := sub.Channel()
	select {
	case msg := <-ch:
		if msg.Payload != "hello" {
			t.Errorf("expected 'hello', got '%s'", msg.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestClient_Cache(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	key := "test-key-" + time.Now().Format(time.RFC3339Nano)
	value := []byte("test-value")

	// Set
	if err := client.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected '%s', got '%s'", value, got)
	}

	// Del
	if err := client.Del(ctx, key); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}
```

**Step 3: Run tests**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./adapters/redis/... -v`
Expected: PASS (or SKIP if Redis not running)

**Step 4: Commit**

```bash
git add server/adapters/redis/
git commit -m "feat(redis): add Redis client adapter with pub/sub and cache"
```

---

### Task 1.4: Add Redis to docker-compose

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Add Redis service**

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: claudeception
      POSTGRES_PASSWORD: claudeception
      POSTGRES_DB: claudeception
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U claudeception"]
      interval: 5s
      timeout: 5s
      retries: 5

  server:
    build:
      context: ./server
      dockerfile: Dockerfile
    environment:
      REDIS_URL: redis://redis:6379/0
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    depends_on:
      - server

volumes:
  pgdata:
```

**Step 2: Verify compose file is valid**

Run: `docker compose -f /Users/kamilrybacki/Projects/Personal/claudeception/docker-compose.yml config --quiet && echo "Valid"`
Expected: "Valid"

**Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "infra: add Redis service to docker-compose"
```

---

## Phase 2: Create Event Publisher

### Task 2.1: Define event types

**Files:**
- Create: `server/events/events.go`

**Step 1: Create event types**

```go
package events

import (
	"encoding/json"
	"time"
)

// EventType identifies the type of event
type EventType string

const (
	EventRuleCreated   EventType = "rule_created"
	EventRuleUpdated   EventType = "rule_updated"
	EventRuleDeleted   EventType = "rule_deleted"
	EventCategoryUpdated EventType = "category_updated"
	EventSyncRequired  EventType = "sync_required"
)

// Event represents a change event published to Redis
type Event struct {
	Type      EventType `json:"event"`
	EntityID  string    `json:"entity_id"`
	TeamID    string    `json:"team_id"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, entityID, teamID string) Event {
	return Event{
		Type:      eventType,
		EntityID:  entityID,
		TeamID:    teamID,
		Version:   time.Now().UnixNano(),
		Timestamp: time.Now().UTC(),
	}
}

// Marshal serializes the event to JSON
func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEvent deserializes an event from JSON
func UnmarshalEvent(data []byte) (Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return e, err
}

// ChannelForTeam returns the Redis channel name for team rule updates
func ChannelForTeam(teamID string) string {
	return "team:" + teamID + ":rules"
}

// ChannelForTeamCategories returns the Redis channel for team category updates
func ChannelForTeamCategories(teamID string) string {
	return "team:" + teamID + ":categories"
}

// ChannelBroadcast returns the broadcast channel for all workers
const ChannelBroadcast = "broadcast:all"

// ChannelForAgent returns the channel for direct agent messages
func ChannelForAgent(agentID string) string {
	return "agent:" + agentID + ":direct"
}
```

**Step 2: Write tests for marshaling**

Create `server/events/events_test.go`:

```go
package events

import (
	"testing"
	"time"
)

func TestEvent_MarshalUnmarshal(t *testing.T) {
	original := Event{
		Type:      EventRuleUpdated,
		EntityID:  "rule-123",
		TeamID:    "team-456",
		Version:   12345,
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	parsed, err := UnmarshalEvent(data)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Type != original.Type {
		t.Errorf("type mismatch: got %s, want %s", parsed.Type, original.Type)
	}
	if parsed.EntityID != original.EntityID {
		t.Errorf("entity_id mismatch: got %s, want %s", parsed.EntityID, original.EntityID)
	}
	if parsed.TeamID != original.TeamID {
		t.Errorf("team_id mismatch: got %s, want %s", parsed.TeamID, original.TeamID)
	}
}

func TestChannelNaming(t *testing.T) {
	if got := ChannelForTeam("abc"); got != "team:abc:rules" {
		t.Errorf("got %s, want team:abc:rules", got)
	}
	if got := ChannelForAgent("xyz"); got != "agent:xyz:direct" {
		t.Errorf("got %s, want agent:xyz:direct", got)
	}
}
```

**Step 3: Run tests**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./events/... -v`
Expected: PASS

**Step 4: Commit**

```bash
git add server/events/
git commit -m "feat(events): add event types and channel naming for Redis pub/sub"
```

---

### Task 2.2: Create event publisher service

**Files:**
- Create: `server/services/publisher/publisher.go`
- Create: `server/services/publisher/publisher_test.go`

**Step 1: Create publisher interface and implementation**

```go
package publisher

import (
	"context"
	"log"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
)

// Publisher broadcasts events to Redis channels
type Publisher interface {
	PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error
	PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error
	PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error
	PublishToAgent(ctx context.Context, agentID string, data []byte) error
}

// RedisPublisher implements Publisher using Redis
type RedisPublisher struct {
	client *redisAdapter.Client
}

// NewRedisPublisher creates a new publisher
func NewRedisPublisher(client *redisAdapter.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

// PublishRuleEvent publishes a rule change event
func (p *RedisPublisher) PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error {
	event := events.NewEvent(eventType, ruleID, teamID)
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	channel := events.ChannelForTeam(teamID)
	log.Printf("Publishing %s to %s: rule=%s", eventType, channel, ruleID)
	return p.client.Publish(ctx, channel, data)
}

// PublishCategoryEvent publishes a category change event
func (p *RedisPublisher) PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error {
	event := events.NewEvent(eventType, categoryID, teamID)
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	channel := events.ChannelForTeamCategories(teamID)
	log.Printf("Publishing %s to %s: category=%s", eventType, channel, categoryID)
	return p.client.Publish(ctx, channel, data)
}

// PublishBroadcast publishes to all workers
func (p *RedisPublisher) PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error {
	event := events.NewEvent(eventType, message, "")
	data, err := event.Marshal()
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, events.ChannelBroadcast, data)
}

// PublishToAgent publishes directly to an agent
func (p *RedisPublisher) PublishToAgent(ctx context.Context, agentID string, data []byte) error {
	channel := events.ChannelForAgent(agentID)
	return p.client.Publish(ctx, channel, data)
}

// NoOpPublisher is a no-op implementation for testing or when Redis is disabled
type NoOpPublisher struct{}

func (p *NoOpPublisher) PublishRuleEvent(ctx context.Context, eventType events.EventType, ruleID, teamID string) error {
	return nil
}

func (p *NoOpPublisher) PublishCategoryEvent(ctx context.Context, eventType events.EventType, categoryID, teamID string) error {
	return nil
}

func (p *NoOpPublisher) PublishBroadcast(ctx context.Context, eventType events.EventType, message string) error {
	return nil
}

func (p *NoOpPublisher) PublishToAgent(ctx context.Context, agentID string, data []byte) error {
	return nil
}
```

**Step 2: Create test**

```go
package publisher

import (
	"context"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
)

func TestRedisPublisher_PublishRuleEvent(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Subscribe first
	teamID := "test-team"
	channel := events.ChannelForTeam(teamID)
	sub := client.Subscribe(ctx, channel)
	defer sub.Close()

	// Wait for subscription
	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Publish
	publisher := NewRedisPublisher(client)
	go func() {
		time.Sleep(100 * time.Millisecond)
		publisher.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-123", teamID)
	}()

	// Receive
	ch := sub.Channel()
	select {
	case msg := <-ch:
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != events.EventRuleUpdated {
			t.Errorf("wrong event type: got %s, want %s", event.Type, events.EventRuleUpdated)
		}
		if event.EntityID != "rule-123" {
			t.Errorf("wrong entity ID: got %s, want rule-123", event.EntityID)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}
```

**Step 3: Run tests**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./services/publisher/... -v`
Expected: PASS (or SKIP if no Redis)

**Step 4: Commit**

```bash
git add server/services/publisher/
git commit -m "feat(publisher): add Redis event publisher service"
```

---

### Task 2.3: Integrate publisher into rule handlers

**Files:**
- Modify: `server/entrypoints/api/handlers/rules.go`
- Modify: `server/cmd/server/main.go`

**Step 1: Add publisher to RulesHandler**

Modify `server/entrypoints/api/handlers/rules.go` to add Publisher field:

```go
// Add to imports
import (
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
)

// Add to RulesHandler struct
type RulesHandler struct {
	service   RuleService
	publisher publisher.Publisher
}

// Modify NewRulesHandler
func NewRulesHandler(service RuleService, pub publisher.Publisher) *RulesHandler {
	return &RulesHandler{
		service:   service,
		publisher: pub,
	}
}
```

**Step 2: Publish events after rule mutations**

In `Create` handler, after successful creation:

```go
// After: return JSON(w, http.StatusCreated, rule)
// Add before the return:
if h.publisher != nil {
	go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleCreated, rule.ID, rule.TeamID)
}
```

In `Update` handler, after successful update:

```go
if h.publisher != nil {
	go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleUpdated, rule.ID, rule.TeamID)
}
```

In `Delete` handler, after successful deletion:

```go
if h.publisher != nil {
	// Need to get teamID before delete - adjust handler to fetch rule first
	go h.publisher.PublishRuleEvent(r.Context(), events.EventRuleDeleted, ruleID, teamID)
}
```

**Step 3: Update main.go to pass publisher**

```go
// In main.go, after Redis client creation:
var pub publisher.Publisher
if redisClient != nil {
	pub = publisher.NewRedisPublisher(redisClient)
} else {
	pub = &publisher.NoOpPublisher{}
}

// Pass to handlers via router config
```

**Step 4: Run existing tests to ensure no regressions**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./entrypoints/api/handlers/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/entrypoints/api/handlers/rules.go server/cmd/server/main.go
git commit -m "feat(handlers): publish Redis events on rule mutations"
```

---

## Phase 3: Create Worker Process

### Task 3.1: Create worker hub with Redis subscription

**Files:**
- Create: `server/worker/hub.go`
- Create: `server/worker/hub_test.go`

**Step 1: Create worker hub**

```go
package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/redis/go-redis/v9"
)

// AgentConn represents a connected agent
type AgentConn struct {
	ID      string
	UserID  string
	AgentID string
	TeamID  string
	Send    chan []byte
	conn    *websocket.Conn
}

// Hub manages agent connections and Redis subscriptions
type Hub struct {
	redisClient *redisAdapter.Client

	// Team -> agents mapping
	teamAgents map[string]map[*AgentConn]struct{}

	// Agent ID -> connection
	agents map[string]*AgentConn

	// Active Redis subscriptions per team
	subscriptions map[string]*redis.PubSub

	// Channels for goroutine communication
	register   chan *AgentConn
	unregister chan *AgentConn

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHub creates a new worker hub
func NewHub(redisClient *redisAdapter.Client) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		redisClient:   redisClient,
		teamAgents:    make(map[string]map[*AgentConn]struct{}),
		agents:        make(map[string]*AgentConn),
		subscriptions: make(map[string]*redis.PubSub),
		register:      make(chan *AgentConn),
		unregister:    make(chan *AgentConn),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			h.cleanup()
			return

		case agent := <-h.register:
			h.handleRegister(agent)

		case agent := <-h.unregister:
			h.handleUnregister(agent)
		}
	}
}

func (h *Hub) handleRegister(agent *AgentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to agents map
	if agent.AgentID != "" {
		h.agents[agent.AgentID] = agent
	}

	// Add to team mapping
	if agent.TeamID != "" {
		if h.teamAgents[agent.TeamID] == nil {
			h.teamAgents[agent.TeamID] = make(map[*AgentConn]struct{})
		}
		h.teamAgents[agent.TeamID][agent] = struct{}{}

		// Subscribe to team channel if first agent for this team
		if len(h.teamAgents[agent.TeamID]) == 1 {
			h.subscribeToTeam(agent.TeamID)
		}
	}

	log.Printf("Agent registered: id=%s team=%s (total: %d)", agent.AgentID, agent.TeamID, len(h.agents))
}

func (h *Hub) handleUnregister(agent *AgentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove from agents map
	if agent.AgentID != "" {
		delete(h.agents, agent.AgentID)
	}

	// Remove from team mapping
	if agent.TeamID != "" {
		if agents, ok := h.teamAgents[agent.TeamID]; ok {
			delete(agents, agent)

			// Unsubscribe if no agents left for this team
			if len(agents) == 0 {
				h.unsubscribeFromTeam(agent.TeamID)
				delete(h.teamAgents, agent.TeamID)
			}
		}
	}

	close(agent.Send)
	log.Printf("Agent unregistered: id=%s (remaining: %d)", agent.AgentID, len(h.agents))
}

func (h *Hub) subscribeToTeam(teamID string) {
	channel := events.ChannelForTeam(teamID)
	sub := h.redisClient.Subscribe(h.ctx, channel)
	h.subscriptions[teamID] = sub

	go h.listenToSubscription(teamID, sub)
	log.Printf("Subscribed to team channel: %s", channel)
}

func (h *Hub) unsubscribeFromTeam(teamID string) {
	if sub, ok := h.subscriptions[teamID]; ok {
		sub.Close()
		delete(h.subscriptions, teamID)
		log.Printf("Unsubscribed from team channel: team:%s:rules", teamID)
	}
}

func (h *Hub) listenToSubscription(teamID string, sub *redis.PubSub) {
	ch := sub.Channel()
	for msg := range ch {
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		h.broadcastToTeam(teamID, event)
	}
}

func (h *Hub) broadcastToTeam(teamID string, event events.Event) {
	h.mu.RLock()
	agents := h.teamAgents[teamID]
	h.mu.RUnlock()

	// Convert event to WebSocket message format
	wsMsg := map[string]interface{}{
		"type":      "config_update",
		"event":     event.Type,
		"entity_id": event.EntityID,
		"version":   event.Version,
		"timestamp": event.Timestamp,
	}
	data, _ := json.Marshal(wsMsg)

	for agent := range agents {
		select {
		case agent.Send <- data:
		default:
			// Buffer full, skip
		}
	}

	log.Printf("Broadcast %s to %d agents in team %s", event.Type, len(agents), teamID)
}

// Register adds an agent to the hub
func (h *Hub) Register(agent *AgentConn) {
	h.register <- agent
}

// Unregister removes an agent from the hub
func (h *Hub) Unregister(agent *AgentConn) {
	h.unregister <- agent
}

// Stop gracefully shuts down the hub
func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all subscriptions
	for _, sub := range h.subscriptions {
		sub.Close()
	}

	// Close all agent connections
	for _, agent := range h.agents {
		close(agent.Send)
	}
}

// Stats returns hub statistics
func (h *Hub) Stats() (agents int, teams int, subscriptions int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.agents), len(h.teamAgents), len(h.subscriptions)
}
```

**Step 2: Write test**

```go
package worker

import (
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
)

func TestHub_RegisterUnregister(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	hub := NewHub(client)
	go hub.Run()
	defer hub.Stop()

	agent := &AgentConn{
		ID:      "conn-1",
		AgentID: "agent-1",
		TeamID:  "team-1",
		Send:    make(chan []byte, 256),
	}

	hub.Register(agent)
	time.Sleep(100 * time.Millisecond)

	agents, teams, subs := hub.Stats()
	if agents != 1 {
		t.Errorf("expected 1 agent, got %d", agents)
	}
	if teams != 1 {
		t.Errorf("expected 1 team, got %d", teams)
	}
	if subs != 1 {
		t.Errorf("expected 1 subscription, got %d", subs)
	}

	hub.Unregister(agent)
	time.Sleep(100 * time.Millisecond)

	agents, teams, subs = hub.Stats()
	if agents != 0 {
		t.Errorf("expected 0 agents, got %d", agents)
	}
}

func TestHub_ReceivesBroadcast(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	hub := NewHub(client)
	go hub.Run()
	defer hub.Stop()

	agent := &AgentConn{
		ID:      "conn-1",
		AgentID: "agent-1",
		TeamID:  "team-1",
		Send:    make(chan []byte, 256),
	}

	hub.Register(agent)
	time.Sleep(200 * time.Millisecond) // Wait for subscription

	// Publish event
	pub := publisher.NewRedisPublisher(client)
	pub.PublishRuleEvent(t.Context(), events.EventRuleUpdated, "rule-123", "team-1")

	// Wait for message
	select {
	case msg := <-agent.Send:
		if len(msg) == 0 {
			t.Error("received empty message")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}
```

**Step 3: Run tests**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./worker/... -v`
Expected: PASS (or SKIP if no Redis)

**Step 4: Commit**

```bash
git add server/worker/
git commit -m "feat(worker): add worker hub with Redis pub/sub integration"
```

---

### Task 3.2: Create worker WebSocket handler

**Files:**
- Create: `server/worker/handler.go`

**Step 1: Create handler**

```go
package worker

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure for production
	},
}

// Handler handles WebSocket connections for workers
type Handler struct {
	hub *Hub
}

// NewHandler creates a new worker WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeHTTP upgrades to WebSocket and manages the connection
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get team ID from query or header
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		teamID = r.Header.Get("X-Team-ID")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	agent := &AgentConn{
		ID:     uuid.New().String(),
		UserID: userID,
		TeamID: teamID,
		Send:   make(chan []byte, 256),
		conn:   conn,
	}

	h.hub.Register(agent)

	go h.writePump(agent)
	go h.readPump(agent)
}

func (h *Handler) readPump(agent *AgentConn) {
	defer func() {
		h.hub.Unregister(agent)
		agent.conn.Close()
	}()

	agent.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	agent.conn.SetPongHandler(func(string) error {
		agent.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := agent.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.handleMessage(agent, data)
	}
}

func (h *Handler) handleMessage(agent *AgentConn, data []byte) {
	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Invalid message: %v", err)
		return
	}

	switch msg.Type {
	case "heartbeat":
		var payload struct {
			AgentID string `json:"agent_id"`
			TeamID  string `json:"team_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			// Update agent info if provided
			if payload.AgentID != "" && agent.AgentID == "" {
				agent.AgentID = payload.AgentID
			}
			if payload.TeamID != "" && agent.TeamID != payload.TeamID {
				// Re-register with new team
				h.hub.Unregister(agent)
				agent.TeamID = payload.TeamID
				agent.Send = make(chan []byte, 256)
				h.hub.Register(agent)
			}
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}

	// Send ack
	ack, _ := json.Marshal(map[string]string{"type": "ack"})
	agent.Send <- ack
}

func (h *Handler) writePump(agent *AgentConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		agent.conn.Close()
	}()

	for {
		select {
		case message, ok := <-agent.Send:
			agent.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				agent.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := agent.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			agent.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := agent.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

**Step 2: Commit**

```bash
git add server/worker/handler.go
git commit -m "feat(worker): add WebSocket handler for worker process"
```

---

### Task 3.3: Create worker entrypoint

**Files:**
- Create: `server/cmd/worker/main.go`

**Step 1: Create worker main**

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/configurator"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/claudeception/server/worker"
)

func main() {
	settings := configurator.LoadSettings()
	ctx := context.Background()

	// Initialize Redis
	redisClient, err := redisAdapter.NewClient(settings.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize worker hub
	hub := worker.NewHub(redisClient)
	go hub.Run()

	// Create router
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}))

	// Health endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		agents, teams, subs := hub.Stats()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","agents":` + itoa(agents) + `,"teams":` + itoa(teams) + `,"subscriptions":` + itoa(subs) + `}`))
	})

	// WebSocket endpoint with auth
	auth := middleware.NewAuth(settings.JWTSecret)
	wsHandler := worker.NewHandler(hub)
	router.With(auth.Middleware).Get("/ws", wsHandler.ServeHTTP)

	// Get worker port (different from API port)
	port := getEnv("WORKER_PORT", "8081")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down worker...")

		hub.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Worker forced to shutdown: %v", err)
		}

		close(done)
	}()

	log.Printf("Worker starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Worker error: %v", err)
	}

	<-done
	log.Println("Worker stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func itoa(n int) string {
	return string(rune('0' + n%10)) // Simple for small numbers
}
```

**Step 2: Fix itoa helper (proper implementation)**

Replace the simple itoa with:

```go
import "strconv"

// Remove the itoa function, use strconv.Itoa instead:
w.Write([]byte(`{"status":"ok","agents":` + strconv.Itoa(agents) + `,"teams":` + strconv.Itoa(teams) + `,"subscriptions":` + strconv.Itoa(subs) + `}`))
```

**Step 3: Build to verify**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go build ./cmd/worker/...`
Expected: No errors

**Step 4: Commit**

```bash
git add server/cmd/worker/
git commit -m "feat(worker): add worker process entrypoint"
```

---

### Task 3.4: Create master entrypoint (API only)

**Files:**
- Rename: `server/cmd/server/main.go` → `server/cmd/master/main.go`
- Remove WebSocket from master

**Step 1: Copy and modify for master**

Create `server/cmd/master/main.go` based on existing `cmd/server/main.go` but:
- Remove WebSocket hub and handler
- Add Redis client and publisher
- Keep all API routes

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kamilrybacki/claudeception/server/adapters/postgres"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/configurator"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api"
	"github.com/kamilrybacki/claudeception/server/services/approvals"
	"github.com/kamilrybacki/claudeception/server/services/auth"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
)

func main() {
	settings := configurator.LoadSettings()
	ctx := context.Background()

	// Initialize database
	pool, err := postgres.NewPool(ctx, settings.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize Redis (optional - graceful degradation)
	var pub publisher.Publisher
	redisClient, err := redisAdapter.NewClient(settings.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis not available, events will not be published: %v", err)
		pub = &publisher.NoOpPublisher{}
	} else {
		defer redisClient.Close()
		if err := redisClient.Ping(ctx); err != nil {
			log.Printf("Warning: Redis ping failed: %v", err)
			pub = &publisher.NoOpPublisher{}
		} else {
			log.Println("Connected to Redis")
			pub = publisher.NewRedisPublisher(redisClient)
		}
	}

	// Initialize repositories
	teamDB := postgres.NewTeamDB(pool)
	ruleDB := postgres.NewRuleDB(pool)
	categoryDB := postgres.NewCategoryDB(pool)
	userDB := postgres.NewUserDB(pool)
	roleDB := postgres.NewRoleDB(pool)
	approvalDB := postgres.NewRuleApprovalDB(pool)
	approvalConfigDB := postgres.NewApprovalConfigDB(pool)

	// Create services
	teamService := &teamServiceImpl{db: teamDB}
	ruleService := &ruleServiceImpl{db: ruleDB, categoryDB: categoryDB}
	categoryService := &categoryServiceImpl{db: categoryDB}
	userService := &userServiceImpl{db: userDB}
	authService := auth.NewService(userDB, roleDB, settings.JWTSecret, 24*time.Hour)
	approvalsService := approvals.NewService(ruleDB, approvalDB, approvalConfigDB, roleDB)

	// Create router (no WebSocket)
	router := api.NewRouter(api.Config{
		JWTSecret:        settings.JWTSecret,
		TeamService:      teamService,
		RuleService:      ruleService,
		CategoryService:  categoryService,
		AuthService:      authService,
		UserService:      userService,
		ApprovalsService: approvalsService,
		Publisher:        pub, // Pass publisher to router
	})

	server := &http.Server{
		Addr:         ":" + settings.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down master...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Master forced to shutdown: %v", err)
		}

		close(done)
	}()

	log.Printf("Master API starting on port %s", settings.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Master error: %v", err)
	}

	<-done
	log.Println("Master stopped")
}
```

**Step 2: Copy services.go to master directory**

```bash
cp server/cmd/server/services.go server/cmd/master/services.go
```

**Step 3: Update api.Config to include Publisher**

Modify `server/entrypoints/api/router.go` to add Publisher field to Config.

**Step 4: Build to verify**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go build ./cmd/master/...`
Expected: No errors

**Step 5: Commit**

```bash
git add server/cmd/master/
git commit -m "feat(master): add master process entrypoint (API only)"
```

---

## Phase 4: Update Docker Configuration

### Task 4.1: Create multi-stage Dockerfile

**Files:**
- Modify: `server/Dockerfile`

**Step 1: Update Dockerfile for multi-target build**

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY ../pkg ../pkg
RUN go mod download

# Copy source
COPY . .

# Build master
RUN CGO_ENABLED=0 GOOS=linux go build -o /master ./cmd/master

# Build worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Master image
FROM alpine:3.19 AS master
RUN apk --no-cache add ca-certificates
COPY --from=builder /master /master
EXPOSE 8080
CMD ["/master"]

# Worker image
FROM alpine:3.19 AS worker
RUN apk --no-cache add ca-certificates
COPY --from=builder /worker /worker
EXPOSE 8081
CMD ["/worker"]
```

**Step 2: Build both targets**

Run: `docker build --target master -t claudeception-master ./server`
Run: `docker build --target worker -t claudeception-worker ./server`
Expected: Both succeed

**Step 3: Commit**

```bash
git add server/Dockerfile
git commit -m "build: add multi-target Dockerfile for master and worker"
```

---

### Task 4.2: Update docker-compose for master/worker

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Update compose file**

```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: claudeception
      POSTGRES_PASSWORD: claudeception
      POSTGRES_DB: claudeception
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U claudeception"]
      interval: 5s
      timeout: 5s
      retries: 5

  master:
    build:
      context: ./server
      target: master
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://claudeception:claudeception@db:5432/claudeception?sslmode=disable
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: dev-secret-change-in-production
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy

  worker:
    build:
      context: ./server
      target: worker
    ports:
      - "8081:8081"
    environment:
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: dev-secret-change-in-production
      WORKER_PORT: "8081"
    depends_on:
      redis:
        condition: service_healthy
    deploy:
      replicas: 2

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://master:8080
      NEXT_PUBLIC_WS_URL: ws://worker:8081
    depends_on:
      - master
      - worker

volumes:
  pgdata:
```

**Step 2: Validate compose file**

Run: `docker compose config --quiet && echo "Valid"`
Expected: "Valid"

**Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "infra: update docker-compose for master/worker architecture"
```

---

## Phase 5: Integration Testing

### Task 5.1: Add integration test for pub/sub flow

**Files:**
- Create: `server/integration/pubsub_test.go`

**Step 1: Create integration test**

```go
package integration

import (
	"context"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
	"github.com/kamilrybacki/claudeception/server/worker"
)

func TestPubSubIntegration(t *testing.T) {
	// Skip if Redis not available
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Start worker hub
	hub := worker.NewHub(client)
	go hub.Run()
	defer hub.Stop()

	// Create mock agent
	received := make(chan []byte, 10)
	agent := &worker.AgentConn{
		ID:      "test-conn",
		AgentID: "test-agent",
		TeamID:  "test-team",
		Send:    received,
	}
	hub.Register(agent)

	// Wait for subscription
	time.Sleep(200 * time.Millisecond)

	// Publish event (simulating master)
	pub := publisher.NewRedisPublisher(client)
	err = pub.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-123", "test-team")
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// Verify agent received update
	select {
	case msg := <-received:
		if len(msg) == 0 {
			t.Error("received empty message")
		}
		t.Logf("Agent received: %s", string(msg))
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for message")
	}

	// Verify stats
	agents, teams, subs := hub.Stats()
	if agents != 1 || teams != 1 || subs != 1 {
		t.Errorf("unexpected stats: agents=%d, teams=%d, subs=%d", agents, teams, subs)
	}
}
```

**Step 2: Run test**

Run: `cd /Users/kamilrybacki/Projects/Personal/claudeception/server && go test ./integration/... -v -run TestPubSubIntegration`
Expected: PASS (or SKIP if no Redis)

**Step 3: Commit**

```bash
git add server/integration/pubsub_test.go
git commit -m "test: add pub/sub integration test for master-worker flow"
```

---

### Task 5.2: Add E2E test with docker-compose

**Files:**
- Create: `tests/e2e/master_worker_test.go`

**Step 1: Create E2E test using testcontainers**

```go
package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestMasterWorkerE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start docker-compose
	comp, err := compose.NewDockerCompose("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("failed to create compose: %v", err)
	}

	t.Cleanup(func() {
		if err := comp.Down(ctx, testcontainers.RemoveOrphans(true)); err != nil {
			t.Logf("failed to stop compose: %v", err)
		}
	})

	err = comp.Up(ctx, compose.Wait(true))
	if err != nil {
		t.Fatalf("failed to start compose: %v", err)
	}

	// Wait for services
	time.Sleep(5 * time.Second)

	// Test master health
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("master health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("master health check returned %d", resp.StatusCode)
	}

	// Test worker health
	resp, err = http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("worker health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("worker health check returned %d", resp.StatusCode)
	}
}
```

**Step 2: Commit**

```bash
git add tests/e2e/
git commit -m "test: add E2E test for master-worker architecture"
```

---

## Summary

This plan implements the master-worker architecture in 5 phases:

1. **Phase 1** (Tasks 1.1-1.4): Add Redis infrastructure
2. **Phase 2** (Tasks 2.1-2.3): Create event publisher
3. **Phase 3** (Tasks 3.1-3.4): Create worker process
4. **Phase 4** (Tasks 4.1-4.2): Update Docker configuration
5. **Phase 5** (Tasks 5.1-5.2): Integration testing

Each task follows TDD with explicit file paths, code, and verification steps.
