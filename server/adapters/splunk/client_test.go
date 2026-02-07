package splunk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	cfg := Config{
		HECURL:     "http://localhost:8088/services/collector/event",
		Token:      "test-token",
		Source:     "test-source",
		SourceType: "test-sourcetype",
		Index:      "test-index",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.config.HECURL != cfg.HECURL {
		t.Errorf("expected HECURL %s, got %s", cfg.HECURL, client.config.HECURL)
	}
}

func TestSend(t *testing.T) {
	var receivedBody []byte
	var receivedAuthHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		receivedBody = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"Success","code":0}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL:     server.URL,
		Token:      "test-token",
		Source:     "test-source",
		SourceType: "test-sourcetype",
		Index:      "test-index",
	})

	event := Event{
		Host: "test-host",
		Event: map[string]interface{}{
			"message": "test message",
			"level":   "info",
		},
	}

	err := client.Send(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedAuthHeader != "Splunk test-token" {
		t.Errorf("expected auth header 'Splunk test-token', got '%s'", receivedAuthHeader)
	}

	var received Event
	if err := json.Unmarshal(receivedBody, &received); err != nil {
		t.Fatalf("failed to unmarshal received body: %v", err)
	}

	if received.Source != "test-source" {
		t.Errorf("expected source 'test-source', got '%s'", received.Source)
	}
	if received.SourceType != "test-sourcetype" {
		t.Errorf("expected sourcetype 'test-sourcetype', got '%s'", received.SourceType)
	}
	if received.Index != "test-index" {
		t.Errorf("expected index 'test-index', got '%s'", received.Index)
	}
	if received.Time == 0 {
		t.Error("expected time to be set")
	}
}

func TestSendBatch(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		receivedBody = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"Success","code":0}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL:     server.URL,
		Token:      "test-token",
		Source:     "test-source",
		SourceType: "test-sourcetype",
		Index:      "test-index",
	})

	events := []Event{
		{Host: "host1", Event: map[string]interface{}{"msg": "event1"}},
		{Host: "host2", Event: map[string]interface{}{"msg": "event2"}},
	}

	err := client.SendBatch(context.Background(), events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(receivedBody)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestSendBatchEmpty(t *testing.T) {
	client := NewClient(Config{
		HECURL: "http://localhost:8088",
		Token:  "test-token",
	})

	err := client.SendBatch(context.Background(), []Event{})
	if err != nil {
		t.Fatalf("expected no error for empty batch, got: %v", err)
	}
}

func TestSendError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"text":"Invalid data format","code":6}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL: server.URL,
		Token:  "test-token",
	})

	event := Event{
		Event: map[string]interface{}{"msg": "test"},
	}

	err := client.Send(context.Background(), event)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain status code 400, got: %v", err)
	}
}

func TestPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/collector/health" {
			t.Errorf("expected path /services/collector/health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"HEC is healthy","code":17}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL: server.URL,
		Token:  "test-token",
	})

	err := client.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPingError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"text":"HEC is unhealthy"}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL: server.URL,
		Token:  "test-token",
	})

	err := client.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for unhealthy HEC")
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(Config{
		HECURL: server.URL,
		Token:  "test-token",
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Send(ctx, Event{Event: map[string]interface{}{"msg": "test"}})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
