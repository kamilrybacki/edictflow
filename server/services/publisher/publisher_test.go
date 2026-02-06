package publisher

import (
	"context"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/events"
)

func TestRedisPublisher_PublishRuleEvent(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Subscribe first
	teamID := "test-team"
	channel := events.ChannelForTeam(teamID)
	sub := client.Subscribe(ctx, channel)
	defer sub.Close()

	// Wait for subscription
	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Publish
	publisher := NewRedisPublisher(client)
	go func() {
		time.Sleep(100 * time.Millisecond)
		publisher.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-123", teamID)
	}()

	// Receive
	ch := sub.Channel()
	select {
	case msg := <-ch:
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != events.EventRuleUpdated {
			t.Errorf("wrong event type: got %s, want %s", event.Type, events.EventRuleUpdated)
		}
		if event.EntityID != "rule-123" {
			t.Errorf("wrong entity ID: got %s, want rule-123", event.EntityID)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}
