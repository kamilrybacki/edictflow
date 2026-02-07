package integration

import (
	"context"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/events"
	"github.com/kamilrybacki/edictflow/server/services/publisher"
	"github.com/kamilrybacki/edictflow/server/worker"
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

func TestPubSubMultipleAgents(t *testing.T) {
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

	// Create multiple agents in same team
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)
	agent1 := &worker.AgentConn{
		ID:      "conn-1",
		AgentID: "agent-1",
		TeamID:  "shared-team",
		Send:    received1,
	}
	agent2 := &worker.AgentConn{
		ID:      "conn-2",
		AgentID: "agent-2",
		TeamID:  "shared-team",
		Send:    received2,
	}
	hub.Register(agent1)
	hub.Register(agent2)

	time.Sleep(200 * time.Millisecond)

	// Verify stats - 2 agents, 1 team, 1 subscription
	agents, teams, subs := hub.Stats()
	if agents != 2 || teams != 1 || subs != 1 {
		t.Errorf("unexpected stats: agents=%d, teams=%d, subs=%d", agents, teams, subs)
	}

	// Publish
	pub := publisher.NewRedisPublisher(client)
	_ = pub.PublishRuleEvent(ctx, events.EventRuleCreated, "rule-456", "shared-team")

	// Both agents should receive
	for i, ch := range []chan []byte{received1, received2} {
		select {
		case msg := <-ch:
			if len(msg) == 0 {
				t.Errorf("agent %d received empty message", i+1)
			}
		case <-time.After(3 * time.Second):
			t.Errorf("timeout waiting for message on agent %d", i+1)
		}
	}
}

func TestPubSubDifferentTeams(t *testing.T) {
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

	// Create agents in different teams
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)
	agent1 := &worker.AgentConn{
		ID:      "conn-1",
		AgentID: "agent-1",
		TeamID:  "team-a",
		Send:    received1,
	}
	agent2 := &worker.AgentConn{
		ID:      "conn-2",
		AgentID: "agent-2",
		TeamID:  "team-b",
		Send:    received2,
	}
	hub.Register(agent1)
	hub.Register(agent2)

	time.Sleep(200 * time.Millisecond)

	// Verify stats - 2 agents, 2 teams, 2 subscriptions
	agents, teams, subs := hub.Stats()
	if agents != 2 || teams != 2 || subs != 2 {
		t.Errorf("unexpected stats: agents=%d, teams=%d, subs=%d", agents, teams, subs)
	}

	// Publish to team-a only
	pub := publisher.NewRedisPublisher(client)
	_ = pub.PublishRuleEvent(ctx, events.EventRuleDeleted, "rule-789", "team-a")

	// Agent 1 should receive
	select {
	case msg := <-received1:
		if len(msg) == 0 {
			t.Error("agent 1 received empty message")
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message on agent 1")
	}

	// Agent 2 should NOT receive (different team)
	select {
	case <-received2:
		t.Error("agent 2 should not have received message for team-a")
	case <-time.After(500 * time.Millisecond):
		// Expected - no message
	}
}
