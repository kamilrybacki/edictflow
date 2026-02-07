package worker

import (
	"context"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/events"
	"github.com/kamilrybacki/edictflow/server/services/publisher"
)

func TestHub_RegisterUnregister(t *testing.T) {
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

	agents, _, subs = hub.Stats()
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

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

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
	_ = pub.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-123", "team-1")

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
