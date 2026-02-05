package redis

import (
	"context"
	"testing"
	"time"
)

func TestClient_PubSub(t *testing.T) {
	// Skip if no Redis available
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	// Test pub/sub
	sub := client.Subscribe(ctx, "test-channel")
	defer sub.Close()

	// Wait for subscription to be ready
	_, err = sub.Receive(ctx)
	if err != nil {
		t.Fatalf("failed to receive subscription confirmation: %v", err)
	}

	// Publish
	go func() {
		time.Sleep(100 * time.Millisecond)
		client.Publish(ctx, "test-channel", []byte("hello"))
	}()

	// Receive
	ch := sub.Channel()
	select {
	case msg := <-ch:
		if msg.Payload != "hello" {
			t.Errorf("expected 'hello', got '%s'", msg.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

func TestClient_Cache(t *testing.T) {
	client, err := NewClient("redis://localhost:6379/0")
	if err != nil {
		t.Skip("Redis not available:", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skip("Redis not available:", err)
	}

	key := "test-key-" + time.Now().Format(time.RFC3339Nano)
	value := []byte("test-value")

	// Set
	if err := client.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Get
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected '%s', got '%s'", value, got)
	}

	// Del
	if err := client.Del(ctx, key); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}
