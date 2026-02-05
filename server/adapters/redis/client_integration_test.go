package redis

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestClient_PubSubMultipleSubscribers(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	channel := "test-multi-sub-" + time.Now().Format(time.RFC3339Nano)

	// Create multiple subscribers
	sub1 := client.Subscribe(ctx, channel)
	defer sub1.Close()
	sub2 := client.Subscribe(ctx, channel)
	defer sub2.Close()

	// Wait for subscriptions
	sub1.Receive(ctx)
	sub2.Receive(ctx)

	// Publish message
	go func() {
		time.Sleep(100 * time.Millisecond)
		client.Publish(ctx, channel, []byte("broadcast-message"))
	}()

	// Both should receive
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ch := sub1.Channel()
		select {
		case msg := <-ch:
			if msg.Payload != "broadcast-message" {
				t.Errorf("sub1: expected 'broadcast-message', got '%s'", msg.Payload)
			}
		case <-time.After(2 * time.Second):
			t.Error("sub1: timeout")
		}
	}()

	go func() {
		defer wg.Done()
		ch := sub2.Channel()
		select {
		case msg := <-ch:
			if msg.Payload != "broadcast-message" {
				t.Errorf("sub2: expected 'broadcast-message', got '%s'", msg.Payload)
			}
		case <-time.After(2 * time.Second):
			t.Error("sub2: timeout")
		}
	}()

	wg.Wait()
}

func TestClient_PubSubPatternSubscription(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Subscribe to pattern
	sub := client.Underlying().PSubscribe(ctx, "team:*:rules")
	defer sub.Close()

	// Wait for subscription
	sub.Receive(ctx)

	// Publish to different team channels
	go func() {
		time.Sleep(100 * time.Millisecond)
		client.Publish(ctx, "team:team-1:rules", []byte("msg1"))
		client.Publish(ctx, "team:team-2:rules", []byte("msg2"))
	}()

	// Should receive both
	ch := sub.Channel()
	received := 0
	timeout := time.After(2 * time.Second)

	for received < 2 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("timeout waiting for messages, received %d/2", received)
		}
	}
}

func TestClient_SetGetWithTTL(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	key := "test-ttl-" + time.Now().Format(time.RFC3339Nano)
	value := []byte("test-value")

	// Set with short TTL
	if err := client.Set(ctx, key, value, 500*time.Millisecond); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get immediately
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected '%s', got '%s'", value, got)
	}

	// Wait for TTL
	time.Sleep(600 * time.Millisecond)

	// Should be expired
	_, err = client.Get(ctx, key)
	if err == nil {
		t.Error("expected key to be expired")
	}
}

func TestClient_ConnectionPooling(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent-test-" + time.Now().Format(time.RFC3339Nano) + "-" + string(rune('0'+n%10))
			client.Set(ctx, key, []byte("value"), time.Minute)
			client.Get(ctx, key)
			client.Del(ctx, key)
		}(i)
	}
	wg.Wait()

	// Should complete without errors
}
