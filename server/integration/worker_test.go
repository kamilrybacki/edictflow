package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/claudeception/server/adapters/redis"
	"github.com/kamilrybacki/claudeception/server/events"
	"github.com/kamilrybacki/claudeception/server/services/publisher"
	"github.com/kamilrybacki/claudeception/server/worker"
)

// TestWorkerHubScaling tests multiple worker hubs receiving the same events
func TestWorkerHubScaling(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Create multiple worker hubs (simulating multiple worker processes)
	hub1 := worker.NewHub(client)
	hub2 := worker.NewHub(client)
	go hub1.Run()
	go hub2.Run()
	defer hub1.Stop()
	defer hub2.Stop()

	// Register agents on different hubs
	teamID := "shared-team"
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)

	agent1 := &worker.AgentConn{
		ID:      "conn-1",
		AgentID: "agent-hub1",
		TeamID:  teamID,
		Send:    received1,
	}
	agent2 := &worker.AgentConn{
		ID:      "conn-2",
		AgentID: "agent-hub2",
		TeamID:  teamID,
		Send:    received2,
	}

	hub1.Register(agent1)
	hub2.Register(agent2)

	// Wait for subscriptions
	time.Sleep(300 * time.Millisecond)

	// Publish event - both hubs should receive via Redis
	pub := publisher.NewRedisPublisher(client)
	pub.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-123", teamID)

	// Both agents should receive
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case msg := <-received1:
			if len(msg) == 0 {
				t.Error("hub1 received empty message")
			}
		case <-time.After(3 * time.Second):
			t.Error("hub1 timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case msg := <-received2:
			if len(msg) == 0 {
				t.Error("hub2 received empty message")
			}
		case <-time.After(3 * time.Second):
			t.Error("hub2 timeout")
		}
	}()

	wg.Wait()
}

// TestWorkerHubAgentRegistrationUnregistration tests agent lifecycle
func TestWorkerHubAgentRegistrationUnregistration(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := worker.NewHub(client)
	go hub.Run()
	defer hub.Stop()

	// Register agents
	agents := make([]*worker.AgentConn, 5)
	for i := 0; i < 5; i++ {
		agents[i] = &worker.AgentConn{
			ID:      "conn-" + string(rune('0'+i)),
			AgentID: "agent-" + string(rune('0'+i)),
			TeamID:  "team-1",
			Send:    make(chan []byte, 256),
		}
		hub.Register(agents[i])
	}

	time.Sleep(200 * time.Millisecond)

	// Verify stats
	agentCount, teamCount, subCount := hub.Stats()
	if agentCount != 5 {
		t.Errorf("expected 5 agents, got %d", agentCount)
	}
	if teamCount != 1 {
		t.Errorf("expected 1 team, got %d", teamCount)
	}
	if subCount != 1 {
		t.Errorf("expected 1 subscription, got %d", subCount)
	}

	// Unregister some agents
	hub.Unregister(agents[0])
	hub.Unregister(agents[1])
	time.Sleep(100 * time.Millisecond)

	agentCount, _, _ = hub.Stats()
	if agentCount != 3 {
		t.Errorf("expected 3 agents after unregister, got %d", agentCount)
	}

	// Unregister all remaining - subscription should be removed
	hub.Unregister(agents[2])
	hub.Unregister(agents[3])
	hub.Unregister(agents[4])
	time.Sleep(100 * time.Millisecond)

	agentCount, teamCount, subCount = hub.Stats()
	if agentCount != 0 || teamCount != 0 || subCount != 0 {
		t.Errorf("expected all zeros, got agents=%d teams=%d subs=%d", agentCount, teamCount, subCount)
	}
}

// TestWorkerHubTeamIsolation tests that messages are isolated to teams
func TestWorkerHubTeamIsolation(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := worker.NewHub(client)
	go hub.Run()
	defer hub.Stop()

	// Agents in different teams
	teamA := make(chan []byte, 10)
	teamB := make(chan []byte, 10)

	agentA := &worker.AgentConn{
		ID:      "conn-a",
		AgentID: "agent-a",
		TeamID:  "team-a",
		Send:    teamA,
	}
	agentB := &worker.AgentConn{
		ID:      "conn-b",
		AgentID: "agent-b",
		TeamID:  "team-b",
		Send:    teamB,
	}

	hub.Register(agentA)
	hub.Register(agentB)
	time.Sleep(200 * time.Millisecond)

	// Publish to team-a only
	pub := publisher.NewRedisPublisher(client)
	pub.PublishRuleEvent(ctx, events.EventRuleCreated, "rule-for-a", "team-a")

	// Team A should receive
	select {
	case msg := <-teamA:
		if len(msg) == 0 {
			t.Error("team-a received empty message")
		}
	case <-time.After(2 * time.Second):
		t.Error("team-a should have received message")
	}

	// Team B should NOT receive
	select {
	case <-teamB:
		t.Error("team-b should NOT have received message for team-a")
	case <-time.After(500 * time.Millisecond):
		// Expected - no message
	}
}

// TestWorkerHubHighThroughput tests high message volume
func TestWorkerHubHighThroughput(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := worker.NewHub(client)
	go hub.Run()
	defer hub.Stop()

	received := make(chan []byte, 1000)
	agent := &worker.AgentConn{
		ID:      "conn-throughput",
		AgentID: "agent-throughput",
		TeamID:  "team-throughput",
		Send:    received,
	}
	hub.Register(agent)
	time.Sleep(200 * time.Millisecond)

	pub := publisher.NewRedisPublisher(client)

	// Send many messages
	messageCount := 100
	go func() {
		for i := 0; i < messageCount; i++ {
			pub.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-"+string(rune('0'+i%10)), "team-throughput")
		}
	}()

	// Count received messages
	receivedCount := 0
	timeout := time.After(5 * time.Second)

loop:
	for {
		select {
		case <-received:
			receivedCount++
			if receivedCount >= messageCount {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if receivedCount < messageCount*90/100 { // Allow 10% loss
		t.Errorf("expected ~%d messages, received %d", messageCount, receivedCount)
	}
}

// TestWorkerHubGracefulShutdown tests that hub stops without panic
func TestWorkerHubGracefulShutdown(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	hub := worker.NewHub(client)
	go hub.Run()

	// Register agent
	received := make(chan []byte, 10)
	agent := &worker.AgentConn{
		ID:      "conn-shutdown",
		AgentID: "agent-shutdown",
		TeamID:  "team-shutdown",
		Send:    received,
	}
	hub.Register(agent)
	time.Sleep(100 * time.Millisecond)

	// Verify agent is registered
	agents, _, _ := hub.Stats()
	if agents != 1 {
		t.Errorf("expected 1 agent before shutdown, got %d", agents)
	}

	// Stop hub - this should not panic
	hub.Stop()
	time.Sleep(200 * time.Millisecond)

	// Test passes if no panic occurred during shutdown
}
