//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/adapters/splunk"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api"
	"github.com/kamilrybacki/edictflow/server/services/metrics"
)

// mockHECServer creates a test server that records received events
type mockHECServer struct {
	server *httptest.Server
	events []splunk.Event
	mu     sync.Mutex
}

func newMockHECServer() *mockHECServer {
	m := &mockHECServer{
		events: make([]splunk.Event, 0),
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/services/collector/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"text":"HEC is healthy","code":17}`))
			return
		}

		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)

		lines := strings.Split(strings.TrimSpace(string(body)), "\n")
		m.mu.Lock()
		for _, line := range lines {
			if line == "" {
				continue
			}
			var event splunk.Event
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				m.events = append(m.events, event)
			}
		}
		m.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"Success","code":0}`))
	}))

	return m
}

func (m *mockHECServer) URL() string {
	return m.server.URL
}

func (m *mockHECServer) Close() {
	m.server.Close()
}

func (m *mockHECServer) Events() []splunk.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	events := make([]splunk.Event, len(m.events))
	copy(events, m.events)
	return events
}

func (m *mockHECServer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = m.events[:0]
}

func TestMetricsIntegration_APIRequestsGenerateMetrics(t *testing.T) {
	mockHEC := newMockHECServer()
	defer mockHEC.Close()

	metricsService := metrics.NewSplunkService(metrics.Config{
		SplunkConfig: splunk.Config{
			HECURL:     mockHEC.URL(),
			Token:      "test-token",
			Source:     "edictflow-test",
			SourceType: "edictflow:metrics",
			Index:      "test",
		},
		BufferSize:    1,
		FlushInterval: 100 * time.Millisecond,
		Hostname:      "test-host",
	})
	defer metricsService.Close()

	router := api.NewRouter(api.Config{
		JWTSecret:      "test-secret",
		MetricsService: metricsService,
	})

	// Make a request to the health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// Wait for flush
	time.Sleep(200 * time.Millisecond)
	metricsService.Flush()
	time.Sleep(100 * time.Millisecond)

	events := mockHEC.Events()
	if len(events) == 0 {
		t.Fatal("expected at least one metric event")
	}

	// Find the API request event
	var found bool
	for _, event := range events {
		if event.Event["type"] == "api_request" {
			found = true
			if event.Event["method"] != "GET" {
				t.Errorf("expected method GET, got %v", event.Event["method"])
			}
			if event.Event["path"] != "/health" {
				t.Errorf("expected path /health, got %v", event.Event["path"])
			}
			if int(event.Event["status_code"].(float64)) != 200 {
				t.Errorf("expected status_code 200, got %v", event.Event["status_code"])
			}
			break
		}
	}

	if !found {
		t.Error("did not find api_request event in recorded events")
	}
}

func TestMetricsIntegration_BufferingAndBatchSending(t *testing.T) {
	mockHEC := newMockHECServer()
	defer mockHEC.Close()

	metricsService := metrics.NewSplunkService(metrics.Config{
		SplunkConfig: splunk.Config{
			HECURL: mockHEC.URL(),
			Token:  "test-token",
		},
		BufferSize:    5,
		FlushInterval: 1 * time.Hour,
		Hostname:      "test-host",
	})
	defer metricsService.Close()

	// Record 4 events (below buffer threshold)
	for i := 0; i < 4; i++ {
		metricsService.RecordAPIRequest("GET", "/test", 200, 10*time.Millisecond, "user1")
	}

	// Should not have flushed yet
	time.Sleep(50 * time.Millisecond)
	if len(mockHEC.Events()) > 0 {
		t.Error("events should not be sent before buffer is full")
	}

	// Add 5th event to trigger flush
	metricsService.RecordAPIRequest("GET", "/test", 200, 10*time.Millisecond, "user1")

	// Wait for async flush
	time.Sleep(200 * time.Millisecond)

	events := mockHEC.Events()
	if len(events) != 5 {
		t.Errorf("expected 5 events after buffer flush, got %d", len(events))
	}
}

func TestMetricsIntegration_HubStats(t *testing.T) {
	mockHEC := newMockHECServer()
	defer mockHEC.Close()

	metricsService := metrics.NewSplunkService(metrics.Config{
		SplunkConfig: splunk.Config{
			HECURL:     mockHEC.URL(),
			Token:      "test-token",
			Source:     "edictflow-worker",
			SourceType: "edictflow:metrics",
		},
		BufferSize:    1,
		FlushInterval: 1 * time.Hour,
		Hostname:      "worker-1",
	})
	defer metricsService.Close()

	// Record hub stats
	metricsService.RecordHubStats(10, 5, 25)

	// Wait for async flush (buffer size is 1)
	time.Sleep(200 * time.Millisecond)

	events := mockHEC.Events()
	if len(events) == 0 {
		t.Fatal("expected hub_stats event")
	}

	event := events[0]
	if event.Event["type"] != "hub_stats" {
		t.Errorf("expected type hub_stats, got %v", event.Event["type"])
	}
	if int(event.Event["agents"].(float64)) != 10 {
		t.Errorf("expected agents 10, got %v", event.Event["agents"])
	}
	if int(event.Event["teams"].(float64)) != 5 {
		t.Errorf("expected teams 5, got %v", event.Event["teams"])
	}
	if int(event.Event["subscriptions"].(float64)) != 25 {
		t.Errorf("expected subscriptions 25, got %v", event.Event["subscriptions"])
	}
	if event.Host != "worker-1" {
		t.Errorf("expected host worker-1, got %v", event.Host)
	}
}

func TestMetricsIntegration_SplunkClientPing(t *testing.T) {
	mockHEC := newMockHECServer()
	defer mockHEC.Close()

	client := splunk.NewClient(splunk.Config{
		HECURL: mockHEC.URL(),
		Token:  "test-token",
	})

	err := client.Ping(context.Background())
	if err != nil {
		t.Errorf("expected successful ping, got error: %v", err)
	}
}

func TestMetricsIntegration_NoOpServiceDoesNothing(t *testing.T) {
	mockHEC := newMockHECServer()
	defer mockHEC.Close()

	noOpService := &metrics.NoOpService{}

	// Record various metrics
	noOpService.RecordAPIRequest("GET", "/test", 200, 10*time.Millisecond, "user1")
	noOpService.RecordAPIError("GET", "/test", "not_found", "user1")
	noOpService.RecordRedisEvent("publish", "channel1", true)
	noOpService.RecordHubStats(10, 5, 25)
	noOpService.RecordDBQuery("SELECT", "rules", 5*time.Millisecond, true)
	noOpService.Flush()
	noOpService.Close()

	// NoOp should not send anything
	if len(mockHEC.Events()) > 0 {
		t.Error("NoOpService should not send any events")
	}
}
