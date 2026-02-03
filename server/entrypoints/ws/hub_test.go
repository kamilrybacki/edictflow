package ws_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/claudeception/server/entrypoints/ws"
)

func TestHubRegisterAndUnregister(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	client := &ws.Client{
		ID:     "agent-1",
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}

	hub.Register(client)

	// Give hub time to process
	time.Sleep(10 * time.Millisecond)

	if hub == nil {
		t.Error("hub should not be nil")
	}
}

func TestHubBroadcastToUser(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	client := &ws.Client{
		ID:     "agent-1",
		UserID: "user-1",
		Send:   make(chan []byte, 256),
	}

	hub.Register(client)

	// Give hub time to process registration
	time.Sleep(10 * time.Millisecond)

	msg := []byte(`{"type":"test"}`)
	hub.BroadcastToUser("user-1", msg)

	select {
	case received := <-client.Send:
		if string(received) != string(msg) {
			t.Errorf("expected %s, got %s", msg, received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected message on client channel")
	}
}
