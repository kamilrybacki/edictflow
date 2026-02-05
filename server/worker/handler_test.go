package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/middleware"
)

func TestHandler_ServeHTTP_Unauthorized(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	hub := NewHub(client)
	go hub.Run()
	defer hub.Stop()

	handler := NewHandler(hub)

	// Create request without auth context
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandler_ServeHTTP_WithAuth(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := NewHub(client)
	go hub.Run()
	defer hub.Stop()

	handler := NewHandler(hub)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inject user ID into context
		ctx := context.WithValue(r.Context(), middleware.UserIDContextKey, "test-user")
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	// Connect via WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?team_id=test-team"
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v (response: %v)", err, resp)
	}
	defer ws.Close()

	// Wait for registration
	time.Sleep(100 * time.Millisecond)

	// Verify agent registered
	agents, teams, _ := hub.Stats()
	if agents != 1 {
		t.Errorf("expected 1 agent, got %d", agents)
	}
	if teams != 1 {
		t.Errorf("expected 1 team, got %d", teams)
	}
}

func TestHandler_HeartbeatMessage(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := NewHub(client)
	go hub.Run()
	defer hub.Stop()

	handler := NewHandler(hub)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserIDContextKey, "test-user")
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer ws.Close()

	// Send heartbeat with agent_id
	heartbeat := map[string]interface{}{
		"type": "heartbeat",
		"payload": map[string]string{
			"agent_id": "agent-123",
			"team_id":  "team-456",
		},
	}
	data, _ := json.Marshal(heartbeat)
	ws.WriteMessage(websocket.TextMessage, data)

	// Read ack
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read ack: %v", err)
	}

	var ack map[string]string
	json.Unmarshal(msg, &ack)
	if ack["type"] != "ack" {
		t.Errorf("expected ack, got %s", ack["type"])
	}
}
