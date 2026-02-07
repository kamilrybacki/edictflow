package publisher

import (
	"context"
	"sync"
	"testing"
	"time"

	redisAdapter "github.com/kamilrybacki/edictflow/server/adapters/redis"
	"github.com/kamilrybacki/edictflow/server/events"
)

func TestRedisPublisher_PublishCategoryEvent(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	teamID := "test-team-cat"
	channel := events.ChannelForTeamCategories(teamID)
	sub := client.Subscribe(ctx, channel)
	defer sub.Close()

	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	publisher := NewRedisPublisher(client)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = publisher.PublishCategoryEvent(ctx, events.EventCategoryUpdated, "cat-123", teamID)
	}()

	ch := sub.Channel()
	select {
	case msg := <-ch:
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != events.EventCategoryUpdated {
			t.Errorf("wrong event type: got %s, want %s", event.Type, events.EventCategoryUpdated)
		}
		if event.EntityID != "cat-123" {
			t.Errorf("wrong entity ID: got %s, want cat-123", event.EntityID)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for event")
	}
}

func TestRedisPublisher_PublishBroadcast(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	sub := client.Subscribe(ctx, events.ChannelBroadcast)
	defer sub.Close()

	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	publisher := NewRedisPublisher(client)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = publisher.PublishBroadcast(ctx, events.EventSyncRequired, "sync now")
	}()

	ch := sub.Channel()
	select {
	case msg := <-ch:
		event, err := events.UnmarshalEvent([]byte(msg.Payload))
		if err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		if event.Type != events.EventSyncRequired {
			t.Errorf("wrong event type: got %s", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

func TestRedisPublisher_PublishToAgent(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	agentID := "agent-direct-test"
	channel := events.ChannelForAgent(agentID)
	sub := client.Subscribe(ctx, channel)
	defer sub.Close()

	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	publisher := NewRedisPublisher(client)
	testData := []byte(`{"command":"refresh"}`)
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = publisher.PublishToAgent(ctx, agentID, testData)
	}()

	ch := sub.Channel()
	select {
	case msg := <-ch:
		if msg.Payload != string(testData) {
			t.Errorf("wrong payload: got %s, want %s", msg.Payload, testData)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for agent message")
	}
}

func TestRedisPublisher_MultipleTeams(t *testing.T) {
	client, err := redisAdapter.NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Subscribe to two team channels
	sub1 := client.Subscribe(ctx, events.ChannelForTeam("team-a"))
	defer sub1.Close()
	sub2 := client.Subscribe(ctx, events.ChannelForTeam("team-b"))
	defer sub2.Close()

	_, _ = sub1.Receive(ctx)
	_, _ = sub2.Receive(ctx)

	publisher := NewRedisPublisher(client)

	// Publish to both teams
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = publisher.PublishRuleEvent(ctx, events.EventRuleCreated, "rule-a", "team-a")
		_ = publisher.PublishRuleEvent(ctx, events.EventRuleUpdated, "rule-b", "team-b")
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ch := sub1.Channel()
		select {
		case msg := <-ch:
			event, _ := events.UnmarshalEvent([]byte(msg.Payload))
			if event.TeamID != "team-a" {
				t.Errorf("team-a got wrong team: %s", event.TeamID)
			}
		case <-time.After(2 * time.Second):
			t.Error("team-a timeout")
		}
	}()

	go func() {
		defer wg.Done()
		ch := sub2.Channel()
		select {
		case msg := <-ch:
			event, _ := events.UnmarshalEvent([]byte(msg.Payload))
			if event.TeamID != "team-b" {
				t.Errorf("team-b got wrong team: %s", event.TeamID)
			}
		case <-time.After(2 * time.Second):
			t.Error("team-b timeout")
		}
	}()

	wg.Wait()
}

func TestNoOpPublisher_AllMethods(t *testing.T) {
	pub := &NoOpPublisher{}
	ctx := context.Background()

	// All methods should return nil
	if err := pub.PublishRuleEvent(ctx, events.EventRuleCreated, "id", "team"); err != nil {
		t.Errorf("PublishRuleEvent should return nil, got %v", err)
	}
	if err := pub.PublishCategoryEvent(ctx, events.EventCategoryUpdated, "id", "team"); err != nil {
		t.Errorf("PublishCategoryEvent should return nil, got %v", err)
	}
	if err := pub.PublishBroadcast(ctx, events.EventSyncRequired, "msg"); err != nil {
		t.Errorf("PublishBroadcast should return nil, got %v", err)
	}
	if err := pub.PublishToAgent(ctx, "agent", []byte("data")); err != nil {
		t.Errorf("PublishToAgent should return nil, got %v", err)
	}
}
