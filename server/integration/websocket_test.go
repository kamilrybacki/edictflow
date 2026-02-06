//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/entrypoints/ws"
	"github.com/kamilrybacki/edictflow/server/integration/testhelpers"
)

// testMessageHandler implements ws.MessageHandler for integration tests
type testMessageHandler struct {
	mu                sync.Mutex
	heartbeats        []ws.HeartbeatPayload
	driftReports      []ws.DriftReportPayload
	contextsDetected  []ws.ContextDetectedPayload
	syncsComplete     []ws.SyncCompletePayload
}

func newTestMessageHandler() *testMessageHandler {
	return &testMessageHandler{
		heartbeats:       make([]ws.HeartbeatPayload, 0),
		driftReports:     make([]ws.DriftReportPayload, 0),
		contextsDetected: make([]ws.ContextDetectedPayload, 0),
		syncsComplete:    make([]ws.SyncCompletePayload, 0),
	}
}

func (h *testMessageHandler) HandleHeartbeat(client *ws.Client, payload ws.HeartbeatPayload) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.heartbeats = append(h.heartbeats, payload)
	return nil
}

func (h *testMessageHandler) HandleDriftReport(client *ws.Client, payload ws.DriftReportPayload) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.driftReports = append(h.driftReports, payload)
	return nil
}

func (h *testMessageHandler) HandleContextDetected(client *ws.Client, payload ws.ContextDetectedPayload) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.contextsDetected = append(h.contextsDetected, payload)
	return nil
}

func (h *testMessageHandler) HandleSyncComplete(client *ws.Client, payload ws.SyncCompletePayload) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.syncsComplete = append(h.syncsComplete, payload)
	return nil
}

func (h *testMessageHandler) HandleChangeDetected(client *ws.Client, payload ws.ChangeDetectedPayload) error {
	return nil
}

func (h *testMessageHandler) HandleChangeUpdated(client *ws.Client, payload ws.ChangeUpdatedPayload) error {
	return nil
}

func (h *testMessageHandler) HandleExceptionRequest(client *ws.Client, payload ws.ExceptionRequestPayload) error {
	return nil
}

func (h *testMessageHandler) HandleRevertComplete(client *ws.Client, payload ws.RevertCompletePayload) error {
	return nil
}

func setupWebSocketServer(messageHandler ws.MessageHandler) (*httptest.Server, *ws.Hub) {
	hub := ws.NewHub()
	go hub.Run()

	handler := ws.NewHandler(hub, messageHandler)

	auth := middleware.NewAuth(testhelpers.TestJWTSecret)

	r := chi.NewRouter()
	r.With(auth.Middleware).Handle("/ws", handler)

	return httptest.NewServer(r), hub
}

func connectWebSocket(t *testing.T, serverURL, userID, teamID string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"

	token, err := testhelpers.GenerateTestToken(userID, teamID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			t.Fatalf("WebSocket dial failed with status %d: %v", resp.StatusCode, err)
		}
		t.Fatalf("WebSocket dial failed: %v", err)
	}

	return conn
}

func TestWebSocket_ConnectWithValidToken(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, _ := setupWebSocketServer(messageHandler)
	defer server.Close()

	conn := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn.Close()

	// Connection successful if we reach here
	t.Log("WebSocket connection established successfully")
}

func TestWebSocket_RejectWithoutToken(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, _ := setupWebSocketServer(messageHandler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	dialer := websocket.Dialer{}
	_, resp, err := dialer.Dial(wsURL, nil)

	if err == nil {
		t.Fatal("Expected connection to be rejected without token")
	}

	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestWebSocket_SendHeartbeatReceiveAck(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, _ := setupWebSocketServer(messageHandler)
	defer server.Close()

	conn := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn.Close()

	// Send heartbeat
	heartbeat := ws.HeartbeatPayload{
		Status:         "active",
		CachedVersion:  1,
		ActiveProjects: []string{"/path/to/project"},
	}

	msg, err := ws.NewMessage(ws.TypeHeartbeat, heartbeat)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	msgBytes, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read ack response
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var ackMsg ws.Message
	if err := json.Unmarshal(data, &ackMsg); err != nil {
		t.Fatalf("Failed to parse ack message: %v", err)
	}

	if ackMsg.Type != ws.TypeAck {
		t.Errorf("Expected message type 'ack', got %s", ackMsg.Type)
	}

	var ackPayload ws.AckPayload
	if err := json.Unmarshal(ackMsg.Payload, &ackPayload); err != nil {
		t.Fatalf("Failed to parse ack payload: %v", err)
	}

	if ackPayload.RefID != msg.ID {
		t.Errorf("Expected ack RefID %s, got %s", msg.ID, ackPayload.RefID)
	}

	// Verify handler received the heartbeat
	time.Sleep(100 * time.Millisecond) // Give handler time to process
	messageHandler.mu.Lock()
	defer messageHandler.mu.Unlock()

	if len(messageHandler.heartbeats) != 1 {
		t.Errorf("Expected 1 heartbeat, got %d", len(messageHandler.heartbeats))
	}
	if messageHandler.heartbeats[0].Status != "active" {
		t.Errorf("Expected status 'active', got %s", messageHandler.heartbeats[0].Status)
	}
}

func TestWebSocket_SendDriftReport(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, _ := setupWebSocketServer(messageHandler)
	defer server.Close()

	conn := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn.Close()

	driftReport := ws.DriftReportPayload{
		ProjectPath:  "/path/to/project",
		ExpectedHash: "abc123",
		ActualHash:   "def456",
		Diff:         "some diff content",
	}

	msg, err := ws.NewMessage(ws.TypeDriftReport, driftReport)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	msgBytes, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read ack
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Verify handler received the drift report
	time.Sleep(100 * time.Millisecond)
	messageHandler.mu.Lock()
	defer messageHandler.mu.Unlock()

	if len(messageHandler.driftReports) != 1 {
		t.Errorf("Expected 1 drift report, got %d", len(messageHandler.driftReports))
	}
	if messageHandler.driftReports[0].ProjectPath != "/path/to/project" {
		t.Errorf("Expected project path '/path/to/project', got %s", messageHandler.driftReports[0].ProjectPath)
	}
}

func TestWebSocket_HubBroadcastToUser(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, hub := setupWebSocketServer(messageHandler)
	defer server.Close()

	// Connect two clients with same user
	conn1 := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn1.Close()

	conn2 := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn2.Close()

	// Connect one client with different user
	conn3 := connectWebSocket(t, server.URL, "user-2", "team-1")
	defer conn3.Close()

	// Give hub time to register clients
	time.Sleep(200 * time.Millisecond)

	// Broadcast to user-1
	broadcastMsg := []byte(`{"type":"test_broadcast","data":"hello"}`)
	hub.BroadcastToUser("user-1", broadcastMsg)

	// Set read deadline for all connections
	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
	conn3.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

	// conn1 and conn2 should receive the message
	_, data1, err1 := conn1.ReadMessage()
	if err1 != nil {
		t.Errorf("conn1 failed to receive broadcast: %v", err1)
	} else if string(data1) != string(broadcastMsg) {
		t.Errorf("conn1 received wrong data: %s", string(data1))
	}

	_, data2, err2 := conn2.ReadMessage()
	if err2 != nil {
		t.Errorf("conn2 failed to receive broadcast: %v", err2)
	} else if string(data2) != string(broadcastMsg) {
		t.Errorf("conn2 received wrong data: %s", string(data2))
	}

	// conn3 should NOT receive the message (different user)
	_, _, err3 := conn3.ReadMessage()
	if err3 == nil {
		t.Error("conn3 should not have received the broadcast")
	}
}

func TestWebSocket_MultipleMessageTypes(t *testing.T) {
	messageHandler := newTestMessageHandler()
	server, _ := setupWebSocketServer(messageHandler)
	defer server.Close()

	conn := connectWebSocket(t, server.URL, "user-1", "team-1")
	defer conn.Close()

	// Send multiple message types
	messages := []struct {
		msgType ws.MessageType
		payload interface{}
	}{
		{ws.TypeHeartbeat, ws.HeartbeatPayload{Status: "active"}},
		{ws.TypeContextDetected, ws.ContextDetectedPayload{
			ProjectPath:     "/project",
			DetectedContext: []string{"frontend"},
			DetectedTags:    []string{"react"},
		}},
		{ws.TypeSyncComplete, ws.SyncCompletePayload{
			ProjectPath:  "/project",
			FilesWritten: []string{"file1.md", "file2.md"},
		}},
	}

	for _, m := range messages {
		msg, err := ws.NewMessage(m.msgType, m.payload)
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}

		msgBytes, _ := json.Marshal(msg)
		if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Read ack
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			t.Fatalf("Failed to read ack: %v", err)
		}
	}

	// Verify all messages were handled
	time.Sleep(200 * time.Millisecond)
	messageHandler.mu.Lock()
	defer messageHandler.mu.Unlock()

	if len(messageHandler.heartbeats) != 1 {
		t.Errorf("Expected 1 heartbeat, got %d", len(messageHandler.heartbeats))
	}
	if len(messageHandler.contextsDetected) != 1 {
		t.Errorf("Expected 1 context detected, got %d", len(messageHandler.contextsDetected))
	}
	if len(messageHandler.syncsComplete) != 1 {
		t.Errorf("Expected 1 sync complete, got %d", len(messageHandler.syncsComplete))
	}
}
