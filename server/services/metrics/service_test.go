package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/adapters/splunk"
)

func TestNewSplunkService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		BufferSize:    50,
		FlushInterval: 100 * time.Millisecond,
		Hostname:      "test-host",
	}

	service := NewSplunkService(cfg)
	if service == nil {
		t.Fatal("expected non-nil service")
	}

	if service.bufferSize != 50 {
		t.Errorf("expected buffer size 50, got %d", service.bufferSize)
	}

	if service.hostname != "test-host" {
		t.Errorf("expected hostname test-host, got %s", service.hostname)
	}

	service.Close()
}

func TestRecordAPIRequest(t *testing.T) {
	var receivedEvents []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		mu.Lock()
		receivedEvents = append(receivedEvents, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL:     server.URL,
			Token:      "test-token",
			Source:     "test",
			SourceType: "test:metrics",
		},
		BufferSize:    10,
		FlushInterval: 1 * time.Hour,
		Hostname:      "test-host",
	})
	defer service.Close()

	service.RecordAPIRequest("GET", "/api/v1/rules", 200, 50*time.Millisecond, "user-123")
	service.Flush()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) == 0 {
		t.Fatal("expected at least one event")
	}

	var event splunk.Event
	lines := strings.Split(strings.TrimSpace(receivedEvents[0]), "\n")
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	eventData := event.Event
	if eventData["type"] != "api_request" {
		t.Errorf("expected type api_request, got %v", eventData["type"])
	}
	if eventData["method"] != "GET" {
		t.Errorf("expected method GET, got %v", eventData["method"])
	}
	if eventData["path"] != "/api/v1/rules" {
		t.Errorf("expected path /api/v1/rules, got %v", eventData["path"])
	}
}

func TestRecordHubStats(t *testing.T) {
	var receivedEvents []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		mu.Lock()
		receivedEvents = append(receivedEvents, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		BufferSize:    10,
		FlushInterval: 1 * time.Hour,
		Hostname:      "worker-1",
	})
	defer service.Close()

	service.RecordHubStats(10, 5, 25)
	service.Flush()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) == 0 {
		t.Fatal("expected at least one event")
	}

	var event splunk.Event
	lines := strings.Split(strings.TrimSpace(receivedEvents[0]), "\n")
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	eventData := event.Event
	if eventData["type"] != "hub_stats" {
		t.Errorf("expected type hub_stats, got %v", eventData["type"])
	}
	if int(eventData["agents"].(float64)) != 10 {
		t.Errorf("expected agents 10, got %v", eventData["agents"])
	}
}

func TestBufferFlush(t *testing.T) {
	flushCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		flushCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		BufferSize:    3,
		FlushInterval: 1 * time.Hour,
		Hostname:      "test-host",
	})
	defer service.Close()

	// Add 3 events to trigger buffer flush
	service.RecordAPIRequest("GET", "/test1", 200, 10*time.Millisecond, "user1")
	service.RecordAPIRequest("GET", "/test2", 200, 10*time.Millisecond, "user2")
	service.RecordAPIRequest("GET", "/test3", 200, 10*time.Millisecond, "user3")

	// Wait for async flush
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount == 0 {
		t.Error("expected at least one flush when buffer is full")
	}
}

func TestBackgroundFlusher(t *testing.T) {
	flushCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		flushCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		BufferSize:    100,
		FlushInterval: 50 * time.Millisecond,
		Hostname:      "test-host",
	})

	service.RecordAPIRequest("GET", "/test", 200, 10*time.Millisecond, "user1")

	// Wait for background flusher
	time.Sleep(150 * time.Millisecond)

	service.Close()

	mu.Lock()
	defer mu.Unlock()

	if flushCount == 0 {
		t.Error("expected background flusher to have run")
	}
}

func TestNoOpService(t *testing.T) {
	service := &NoOpService{}

	// These should not panic
	service.RecordAPIRequest("GET", "/test", 200, 10*time.Millisecond, "user1")
	service.RecordAPIError("GET", "/test", "not_found", "user1")
	service.RecordRedisEvent("publish", "channel1", true)
	service.RecordRedisPublish("channel1", "rule_updated", true, 5)
	service.RecordRedisSubscription("team:123:rules", "subscribe")
	service.RecordHubStats(10, 5, 25)
	service.RecordAgentConnection("agent-1", "team-1", "connected")
	service.RecordWebSocketMessage("inbound", "heartbeat", "agent-1", 64)
	service.RecordBroadcast("team-1", "rule_updated", 5)
	service.RecordDBQuery("SELECT", "rules", 5*time.Millisecond, true)
	service.RecordDBPoolStats(25, 5, 20, 50)
	service.RecordHealthCheck("postgres", "healthy", 2)
	service.RecordWorkerHeartbeat("worker-1", 10, 5)

	if err := service.Flush(); err != nil {
		t.Errorf("expected no error from NoOp Flush, got %v", err)
	}

	if err := service.Close(); err != nil {
		t.Errorf("expected no error from NoOp Close, got %v", err)
	}
}

func TestFlushEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request for empty buffer")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		FlushInterval: 1 * time.Hour,
		Hostname:      "test-host",
	})
	defer service.Close()

	err := service.Flush()
	if err != nil {
		t.Errorf("expected no error flushing empty buffer, got %v", err)
	}
}

func TestNewMetricsMethods(t *testing.T) {
	var receivedEvents []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		mu.Lock()
		receivedEvents = append(receivedEvents, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewSplunkService(Config{
		SplunkConfig: splunk.Config{
			HECURL: server.URL,
			Token:  "test-token",
		},
		BufferSize:    20,
		FlushInterval: 1 * time.Hour,
		Hostname:      "test-host",
	})
	defer service.Close()

	// Record various new metric types
	service.RecordRedisPublish("team:123:rules", "rule_updated", true, 5)
	service.RecordRedisSubscription("team:123:rules", "subscribe")
	service.RecordAgentConnection("agent-abc", "team-123", "connected")
	service.RecordWebSocketMessage("inbound", "heartbeat", "agent-abc", 64)
	service.RecordBroadcast("team-123", "rule_updated", 5)
	service.RecordDBPoolStats(25, 5, 20, 50)
	service.RecordHealthCheck("postgres", "healthy", 2)
	service.RecordWorkerHeartbeat("worker-1", 10, 5)

	service.Flush()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) == 0 {
		t.Fatal("expected events to be recorded")
	}

	// Verify the events contain expected types
	allEvents := strings.Join(receivedEvents, " ")
	expectedTypes := []string{
		"redis_publish",
		"redis_subscription",
		"agent_connection",
		"websocket_message",
		"broadcast",
		"db_pool_stats",
		"health_check",
		"worker_heartbeat",
	}

	for _, expectedType := range expectedTypes {
		if !strings.Contains(allEvents, expectedType) {
			t.Errorf("expected event type %s not found in events", expectedType)
		}
	}
}
