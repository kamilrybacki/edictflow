package events

import (
	"testing"
	"time"
)

func TestEvent_MarshalUnmarshal(t *testing.T) {
	original := Event{
		Type:      EventRuleUpdated,
		EntityID:  "rule-123",
		TeamID:    "team-456",
		Version:   12345,
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	parsed, err := UnmarshalEvent(data)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Type != original.Type {
		t.Errorf("type mismatch: got %s, want %s", parsed.Type, original.Type)
	}
	if parsed.EntityID != original.EntityID {
		t.Errorf("entity_id mismatch: got %s, want %s", parsed.EntityID, original.EntityID)
	}
	if parsed.TeamID != original.TeamID {
		t.Errorf("team_id mismatch: got %s, want %s", parsed.TeamID, original.TeamID)
	}
}

func TestChannelNaming(t *testing.T) {
	if got := ChannelForTeam("abc"); got != "team:abc:rules" {
		t.Errorf("got %s, want team:abc:rules", got)
	}
	if got := ChannelForAgent("xyz"); got != "agent:xyz:direct" {
		t.Errorf("got %s, want agent:xyz:direct", got)
	}
}
