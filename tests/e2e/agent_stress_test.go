// e2e/agent_stress_test.go
// Stress tests for agent-server interactions
package e2e

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestAgentStressMultipleConnections tests many agents connecting simultaneously
func TestAgentStressMultipleConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Use all 5 agent accounts
	agentEmails := []string{
		"agent1@test.local",
		"agent2@test.local",
		"agent3@test.local",
		"agent4@test.local",
		"agent5@test.local",
	}

	tokens := make([]string, len(agentEmails))
	for i, email := range agentEmails {
		token, err := loginUser(email, testUserPassword)
		if err != nil {
			t.Fatalf("login failed for %s: %v", email, err)
		}
		tokens[i] = token
	}

	t.Run("AllAgentsConnectConcurrently", func(t *testing.T) {
		var wg sync.WaitGroup
		var successCount int32
		var failCount int32
		connections := make([]*websocket.Conn, len(tokens))
		var mu sync.Mutex

		for i, token := range tokens {
			wg.Add(1)
			go func(idx int, tok string) {
				defer wg.Done()
				conn, err := connectWebSocket(tok)
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					t.Logf("Agent %d failed to connect: %v", idx, err)
					return
				}
				atomic.AddInt32(&successCount, 1)
				mu.Lock()
				connections[idx] = conn
				mu.Unlock()
			}(i, token)
		}

		wg.Wait()

		// Clean up connections
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}

		if successCount != int32(len(tokens)) {
			t.Errorf("Expected %d successful connections, got %d", len(tokens), successCount)
		} else {
			t.Logf("All %d agents connected successfully", successCount)
		}
	})
}

// TestAgentStressRapidHeartbeats tests rapid heartbeat messages from multiple agents
func TestAgentStressRapidHeartbeats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Use 3 agents for this test
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

	t.Run("RapidHeartbeatsFromMultipleAgents", func(t *testing.T) {
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

		// Each agent sends 10 rapid heartbeats
		numHeartbeats := 10
		var wg sync.WaitGroup
		var totalSent int32
		var totalFailed int32

		for i, conn := range connections {
			wg.Add(1)
			go func(idx int, c *websocket.Conn) {
				defer wg.Done()
				for j := 0; j < numHeartbeats; j++ {
					heartbeat := Message{
						Type: "heartbeat",
						ID:   fmt.Sprintf("stress-hb-agent%d-%d", idx, j),
						Payload: HeartbeatPayload{
							Status:         "online",
							CachedVersion:  j,
							ActiveProjects: []string{fmt.Sprintf("/agent%d/project", idx)},
						},
					}
					if err := sendMessage(c, heartbeat); err != nil {
						atomic.AddInt32(&totalFailed, 1)
					} else {
						atomic.AddInt32(&totalSent, 1)
					}
					time.Sleep(50 * time.Millisecond) // Small delay between heartbeats
				}
			}(i, conn)
		}

		wg.Wait()

		expectedTotal := int32(len(tokens) * numHeartbeats)
		if totalSent != expectedTotal {
			t.Errorf("Expected %d heartbeats sent, got %d (failed: %d)", expectedTotal, totalSent, totalFailed)
		} else {
			t.Logf("Successfully sent %d heartbeats from %d agents", totalSent, len(tokens))
		}
	})
}

// TestAgentStressReconnectionCycle tests rapid connect/disconnect cycles
func TestAgentStressReconnectionCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("RapidReconnectionCycles", func(t *testing.T) {
		numCycles := 5
		var successfulCycles int

		for i := 0; i < numCycles; i++ {
			conn, err := connectWebSocket(token)
			if err != nil {
				t.Logf("Cycle %d: connection failed: %v", i, err)
				continue
			}

			// Send heartbeat
			heartbeat := Message{
				Type: "heartbeat",
				ID:   fmt.Sprintf("reconnect-cycle-%d", i),
				Payload: HeartbeatPayload{
					Status:         "online",
					CachedVersion:  i,
					ActiveProjects: []string{},
				},
			}
			if err := sendMessage(conn, heartbeat); err != nil {
				t.Logf("Cycle %d: heartbeat failed: %v", i, err)
				conn.Close()
				continue
			}

			// Close connection
			conn.Close()
			successfulCycles++

			// Brief pause between cycles
			time.Sleep(100 * time.Millisecond)
		}

		if successfulCycles != numCycles {
			t.Errorf("Expected %d successful cycles, got %d", numCycles, successfulCycles)
		} else {
			t.Logf("Completed %d reconnection cycles successfully", successfulCycles)
		}
	})
}

// TestAgentStressConcurrentMessages tests concurrent message sending from all agents
func TestAgentStressConcurrentMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// Use all 5 agents
	agentEmails := []string{
		"agent1@test.local",
		"agent2@test.local",
		"agent3@test.local",
		"agent4@test.local",
		"agent5@test.local",
	}

	tokens := make([]string, len(agentEmails))
	for i, email := range agentEmails {
		token, err := loginUser(email, testUserPassword)
		if err != nil {
			t.Fatalf("login failed for %s: %v", email, err)
		}
		tokens[i] = token
	}

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

	t.Run("ConcurrentChangeDetectedMessages", func(t *testing.T) {
		var wg sync.WaitGroup
		var successCount int32

		// All agents send change_detected simultaneously
		for i, conn := range connections {
			wg.Add(1)
			go func(idx int, c *websocket.Conn) {
				defer wg.Done()
				changeDetected := Message{
					Type: "change_detected",
					ID:   fmt.Sprintf("stress-change-agent%d", idx),
					Payload: ChangeDetectedPayload{
						RuleID:          testRuleID,
						FilePath:        fmt.Sprintf("/agent%d/project/CLAUDE.md", idx),
						OriginalHash:    fmt.Sprintf("original-%d", idx),
						ModifiedHash:    fmt.Sprintf("modified-%d", idx),
						Diff:            "- old\n+ new",
						EnforcementMode: "block",
					},
				}
				if err := sendMessage(c, changeDetected); err == nil {
					atomic.AddInt32(&successCount, 1)
				}
			}(i, conn)
		}

		wg.Wait()

		if successCount != int32(len(tokens)) {
			t.Errorf("Expected %d change_detected messages, sent %d", len(tokens), successCount)
		} else {
			t.Logf("All %d agents sent change_detected messages concurrently", successCount)
		}
	})

	t.Run("ConcurrentMixedMessages", func(t *testing.T) {
		var wg sync.WaitGroup
		var heartbeatsSent int32
		var changesSent int32

		// Each agent sends alternating heartbeats and change_detected messages
		for i, conn := range connections {
			wg.Add(1)
			go func(idx int, c *websocket.Conn) {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					if j%2 == 0 {
						heartbeat := Message{
							Type: "heartbeat",
							ID:   fmt.Sprintf("mixed-hb-%d-%d", idx, j),
							Payload: HeartbeatPayload{
								Status:         "online",
								CachedVersion:  j,
								ActiveProjects: []string{},
							},
						}
						if err := sendMessage(c, heartbeat); err == nil {
							atomic.AddInt32(&heartbeatsSent, 1)
						}
					} else {
						change := Message{
							Type: "change_detected",
							ID:   fmt.Sprintf("mixed-change-%d-%d", idx, j),
							Payload: ChangeDetectedPayload{
								RuleID:          testRuleID,
								FilePath:        fmt.Sprintf("/agent%d/project/file%d.txt", idx, j),
								OriginalHash:    "abc",
								ModifiedHash:    "def",
								Diff:            "diff",
								EnforcementMode: "warning",
							},
						}
						if err := sendMessage(c, change); err == nil {
							atomic.AddInt32(&changesSent, 1)
						}
					}
					time.Sleep(20 * time.Millisecond)
				}
			}(i, conn)
		}

		wg.Wait()

		t.Logf("Sent %d heartbeats and %d change_detected messages", heartbeatsSent, changesSent)
	})
}

// TestAgentStressConnectionDuration tests holding connections open for extended period
func TestAgentStressConnectionDuration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	token, err := loginUser(testUserEmail, testUserPassword)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Run("LongLivedConnection", func(t *testing.T) {
		conn, err := connectWebSocket(token)
		if err != nil {
			t.Fatalf("connection failed: %v", err)
		}
		defer conn.Close()

		// Keep connection alive for 5 seconds with periodic heartbeats
		duration := 5 * time.Second
		interval := 1 * time.Second
		heartbeatCount := 0

		start := time.Now()
		for time.Since(start) < duration {
			heartbeat := Message{
				Type: "heartbeat",
				ID:   fmt.Sprintf("long-lived-%d", heartbeatCount),
				Payload: HeartbeatPayload{
					Status:         "online",
					CachedVersion:  heartbeatCount,
					ActiveProjects: []string{"/test/project"},
				},
			}
			if err := sendMessage(conn, heartbeat); err != nil {
				t.Errorf("heartbeat %d failed: %v", heartbeatCount, err)
				break
			}
			heartbeatCount++
			time.Sleep(interval)
		}

		t.Logf("Connection held for %v, sent %d heartbeats", time.Since(start), heartbeatCount)
	})
}
