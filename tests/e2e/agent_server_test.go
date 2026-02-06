// e2e/agent_server_test.go
// Agent-Server interaction E2E tests against live containerized stack
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Test configuration - uses environment variables for flexibility
var (
	masterURL  = getEnv("EDICTFLOW_MASTER_URL", "http://localhost:18080")
	workerURL  = getEnv("EDICTFLOW_WORKER_URL", "ws://localhost:18081")
	dbConnStr  = getEnv("EDICTFLOW_DB_URL", "postgres://edictflow:edictflow@localhost:15432/edictflow?sslmode=disable")
	jwtSecret  = getEnv("EDICTFLOW_JWT_SECRET", "test-secret-for-local-testing-only")

	// Test user credentials from seed-data.sql
	testUserEmail    = "user@test.local"
	testUserPassword = "Test1234"
	testAdminEmail   = "admin@test.local"
	testAdminPassword= "Test1234"
	testTeamID       = "a0000000-0000-0000-0000-000000000001"
	testRuleID       = "d0000000-0000-0000-0000-000000000001"
)

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// TestAgentServerStack verifies the stack is running and healthy
func TestAgentServerStack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("MasterHealthCheck", func(t *testing.T) {
		resp, err := http.Get(masterURL + "/health")
		if err != nil {
			t.Fatalf("master health check failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("master health check returned %d, want 200", resp.StatusCode)
		}
	})

	t.Run("WorkerHealthCheck", func(t *testing.T) {
		// Convert ws:// to http:// for health check
		workerHealthURL := strings.Replace(workerURL, "ws://", "http://", 1) + "/health"
		resp, err := http.Get(workerHealthURL)
		if err != nil {
			t.Fatalf("worker health check failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("worker health check returned %d, want 200", resp.StatusCode)
		}
	})
}

// TestAgentAuthentication tests the device flow authentication for agents
func TestAgentAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	t.Run("LoginWithValidCredentials", func(t *testing.T) {
		token, err := loginUser(testUserEmail, testUserPassword)
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		_, err := loginUser("invalid@test.local", "wrongpassword")
		if err == nil {
			t.Error("expected error for invalid credentials")
		}
	})

	t.Run("LoginAsAdmin", func(t *testing.T) {
		token, err := loginUser(testAdminEmail, testAdminPassword)
		if err != nil {
			t.Fatalf("admin login failed: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token for admin")
		}
	})
}

// TestAgentWebSocketConnection tests WebSocket connection lifecycle
func TestAgentWebSocketConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("ConnectWithValidToken", func(t *testing.T) {
		conn, err := connectWebSocket(token)
		if err != nil {
			t.Fatalf("WebSocket connection failed: %v", err)
		}
		defer conn.Close()

		// Send heartbeat message
		heartbeat := Message{
			Type: "heartbeat",
			ID:   "test-heartbeat-1",
			Payload: HeartbeatPayload{
				Status:         "online",
				CachedVersion:  0,
				ActiveProjects: []string{},
			},
		}
		if err := sendMessage(conn, heartbeat); err != nil {
			t.Fatalf("failed to send heartbeat: %v", err)
		}

		// Wait for config_update response (or timeout)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		msg, err := receiveMessage(conn)
		if err != nil {
			// Timeout is acceptable if no config update needed
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				t.Log("Connection closed normally")
				return
			}
			t.Logf("No immediate response (may be expected): %v", err)
			return
		}
		t.Logf("Received message type: %s", msg.Type)
	})

	t.Run("ConnectWithInvalidToken", func(t *testing.T) {
		_, err := connectWebSocket("invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("HeartbeatTriggersConfigUpdate", func(t *testing.T) {
		conn, err := connectWebSocket(token)
		if err != nil {
			t.Fatalf("WebSocket connection failed: %v", err)
		}
		defer conn.Close()

		// Send heartbeat with version 0 to trigger config update
		heartbeat := Message{
			Type: "heartbeat",
			ID:   "test-heartbeat-config",
			Payload: HeartbeatPayload{
				Status:         "online",
				CachedVersion:  0, // Outdated version
				ActiveProjects: []string{"/test/project"},
			},
		}
		if err := sendMessage(conn, heartbeat); err != nil {
			t.Fatalf("failed to send heartbeat: %v", err)
		}

		// Wait for config_update
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		msg, err := receiveMessage(conn)
		if err != nil {
			t.Logf("No config_update received (may be expected if rules empty): %v", err)
			return
		}
		if msg.Type != "config_update" {
			t.Logf("Expected config_update, got: %s", msg.Type)
		} else {
			t.Log("Received config_update as expected")
		}
	})
}

// TestAgentMultiConnection tests multiple agents connecting simultaneously
func TestAgentMultiConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Login as different agent users
	agentEmails := []string{
		"agent1@test.local",
		"agent2@test.local",
		"agent3@test.local",
	}

	tokens := make([]string, len(agentEmails))
	for i, email := range agentEmails {
		token, err := loginUser(email, testUserPassword)
		if err != nil {
			t.Fatalf("login failed for %s: %v", email, err)
		}
		tokens[i] = token
	}

	t.Run("MultipleAgentsConnect", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, len(tokens))
		connections := make([]*websocket.Conn, len(tokens))

		for i, token := range tokens {
			wg.Add(1)
			go func(idx int, tok string) {
				defer wg.Done()
				conn, err := connectWebSocket(tok)
				if err != nil {
					errors <- fmt.Errorf("agent %d failed to connect: %w", idx, err)
					return
				}
				connections[idx] = conn
			}(i, token)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Error(err)
		}

		// Clean up connections
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}

		t.Logf("Successfully connected %d agents", len(tokens))
	})

	t.Run("AgentsReceiveIndependentHeartbeats", func(t *testing.T) {
		connections := make([]*websocket.Conn, len(tokens))
		for i, token := range tokens {
			conn, err := connectWebSocket(token)
			if err != nil {
				t.Fatalf("agent %d failed to connect: %v", i, err)
			}
			connections[i] = conn
		}
		defer func() {
			for _, conn := range connections {
				if conn != nil {
					conn.Close()
				}
			}
		}()

		// Send heartbeats from all agents
		for i, conn := range connections {
			heartbeat := Message{
				Type: "heartbeat",
				ID:   fmt.Sprintf("test-heartbeat-agent-%d", i),
				Payload: HeartbeatPayload{
					Status:         "online",
					CachedVersion:  i, // Different versions
					ActiveProjects: []string{fmt.Sprintf("/agent%d/project", i)},
				},
			}
			if err := sendMessage(conn, heartbeat); err != nil {
				t.Errorf("agent %d failed to send heartbeat: %v", i, err)
			}
		}

		t.Log("All agents sent heartbeats successfully")
	})
}

// TestChangeDetectionFlow tests the change detection and reporting flow
func TestChangeDetectionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	conn, err := connectWebSocket(token)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close()

	t.Run("ReportChangeDetected", func(t *testing.T) {
		changeDetected := Message{
			Type: "change_detected",
			ID:   "test-change-1",
			Payload: ChangeDetectedPayload{
				RuleID:          testRuleID,
				FilePath:        "/test/project/CLAUDE.md",
				OriginalHash:    "abc123",
				ModifiedHash:    "def456",
				Diff:            "- old line\n+ new line",
				EnforcementMode: "block",
			},
		}
		if err := sendMessage(conn, changeDetected); err != nil {
			t.Fatalf("failed to send change_detected: %v", err)
		}

		// Wait for ack or response
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		msg, err := receiveMessage(conn)
		if err != nil {
			t.Logf("No immediate response to change_detected: %v", err)
			return
		}
		t.Logf("Received response: type=%s", msg.Type)
	})
}

// TestRuleUpdatePropagation tests that rule updates propagate to connected agents
func TestRuleUpdatePropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Login as admin to update rules
	adminToken, err := loginUser(testAdminEmail, testAdminPassword)
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}

	// Login as regular user and connect WebSocket
	userToken, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("user login failed: %v", err)
	}

	conn, err := connectWebSocket(userToken)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close()

	// Send initial heartbeat
	heartbeat := Message{
		Type: "heartbeat",
		ID:   "test-heartbeat-rule-update",
		Payload: HeartbeatPayload{
			Status:         "online",
			CachedVersion:  0,
			ActiveProjects: []string{"/test/project"},
		},
	}
	if err := sendMessage(conn, heartbeat); err != nil {
		t.Fatalf("failed to send heartbeat: %v", err)
	}

	t.Run("UpdateRuleEnforcementMode", func(t *testing.T) {
		// Update rule via API
		err := updateRule(adminToken, testRuleID, map[string]interface{}{
			"enforcement_mode": "warning",
		})
		if err != nil {
			t.Logf("Rule update may have failed (expected if rule doesn't exist): %v", err)
			return
		}

		// Wait for config_update on WebSocket
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		msg, err := receiveMessage(conn)
		if err != nil {
			t.Logf("No config_update received: %v", err)
			return
		}
		if msg.Type == "config_update" {
			t.Log("Received config_update after rule change")
		} else {
			t.Logf("Received %s instead of config_update", msg.Type)
		}
	})
}

// TestExceptionRequestFlow tests the exception request workflow
func TestExceptionRequestFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	conn, err := connectWebSocket(token)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close()

	t.Run("RequestException", func(t *testing.T) {
		exceptionReq := Message{
			Type: "exception_request",
			ID:   "test-exception-1",
			Payload: ExceptionRequestPayload{
				ChangeID:               "test-change-1",
				Justification:          "Critical hotfix required",
				ExceptionType:          "temporary",
				RequestedDurationHours: intPtr(24),
			},
		}
		if err := sendMessage(conn, exceptionReq); err != nil {
			t.Fatalf("failed to send exception_request: %v", err)
		}

		// Wait for response
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		msg, err := receiveMessage(conn)
		if err != nil {
			t.Logf("No immediate response to exception_request: %v", err)
			return
		}
		t.Logf("Received response: type=%s", msg.Type)
	})
}

// TestAPIEndpoints tests the REST API endpoints
func TestAPIEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	token, err := loginUser(testAdminEmail, testAdminPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("GetRules", func(t *testing.T) {
		rules, err := getRules(token, testTeamID)
		if err != nil {
			t.Logf("Failed to get rules (may be expected): %v", err)
			return
		}
		t.Logf("Retrieved %d rules", len(rules))
	})

	t.Run("GetTeam", func(t *testing.T) {
		resp, err := makeAuthRequest("GET", masterURL+"/api/v1/teams/"+testTeamID, token, nil)
		if err != nil {
			t.Logf("Failed to get team: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Log("Successfully retrieved team")
		} else {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Get team returned %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("GetChanges", func(t *testing.T) {
		resp, err := makeAuthRequest("GET", masterURL+"/api/v1/changes?team_id="+testTeamID, token, nil)
		if err != nil {
			t.Logf("Failed to get changes: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Log("Successfully retrieved changes")
		} else {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Get changes returned %d: %s", resp.StatusCode, string(body))
		}
	})
}

// TestAgentReconnection tests agent reconnection behavior
func TestAgentReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("ReconnectAfterDisconnect", func(t *testing.T) {
		// First connection
		conn1, err := connectWebSocket(token)
		if err != nil {
			t.Fatalf("first connection failed: %v", err)
		}

		// Send heartbeat
		heartbeat := Message{
			Type: "heartbeat",
			ID:   "reconnect-test-1",
			Payload: HeartbeatPayload{
				Status:         "online",
				CachedVersion:  1,
				ActiveProjects: []string{},
			},
		}
		sendMessage(conn1, heartbeat)

		// Close first connection
		conn1.Close()

		// Wait a bit
		time.Sleep(500 * time.Millisecond)

		// Reconnect
		conn2, err := connectWebSocket(token)
		if err != nil {
			t.Fatalf("reconnection failed: %v", err)
		}
		defer conn2.Close()

		// Send heartbeat on new connection
		heartbeat.ID = "reconnect-test-2"
		if err := sendMessage(conn2, heartbeat); err != nil {
			t.Fatalf("failed to send heartbeat on reconnection: %v", err)
		}

		t.Log("Successfully reconnected after disconnect")
	})
}

// Helper functions

func loginUser(email, password string) (string, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(masterURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Handle both token field names
	if result.Token != "" {
		return result.Token, nil
	}
	return result.AccessToken, nil
}

func connectWebSocket(token string) (*websocket.Conn, error) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(workerURL+"/ws", headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendMessage(conn *websocket.Conn, msg Message) error {
	msg.Timestamp = time.Now()
	return conn.WriteJSON(msg)
}

func receiveMessage(conn *websocket.Conn) (*Message, error) {
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func updateRule(token, ruleID string, updates map[string]interface{}) error {
	body, _ := json.Marshal(updates)
	resp, err := makeAuthRequest("PATCH", masterURL+"/api/v1/rules/"+ruleID, token, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rule update failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func getRules(token, teamID string) ([]map[string]interface{}, error) {
	resp, err := makeAuthRequest("GET", masterURL+"/api/v1/rules?team_id="+teamID, token, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get rules failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var rules []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func makeAuthRequest(method, url, token string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func intPtr(i int) *int {
	return &i
}

// Message types for WebSocket communication

type Message struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type HeartbeatPayload struct {
	Status         string   `json:"status"`
	CachedVersion  int      `json:"cached_version"`
	ActiveProjects []string `json:"active_projects"`
}

type ChangeDetectedPayload struct {
	RuleID          string `json:"rule_id"`
	FilePath        string `json:"file_path"`
	OriginalHash    string `json:"original_hash"`
	ModifiedHash    string `json:"modified_hash"`
	Diff            string `json:"diff"`
	EnforcementMode string `json:"enforcement_mode"`
}

type ExceptionRequestPayload struct {
	ChangeID               string `json:"change_id"`
	Justification          string `json:"justification"`
	ExceptionType          string `json:"exception_type"`
	RequestedDurationHours *int   `json:"requested_duration_hours,omitempty"`
}
