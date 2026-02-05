package events

import (
	"encoding/json"
	"time"
)

// EventType identifies the type of event
type EventType string

const (
	EventRuleCreated     EventType = "rule_created"
	EventRuleUpdated     EventType = "rule_updated"
	EventRuleDeleted     EventType = "rule_deleted"
	EventCategoryUpdated EventType = "category_updated"
	EventSyncRequired    EventType = "sync_required"
)

// Event represents a change event published to Redis
type Event struct {
	Type      EventType `json:"event"`
	EntityID  string    `json:"entity_id"`
	TeamID    string    `json:"team_id"`
	Version   int64     `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, entityID, teamID string) Event {
	return Event{
		Type:      eventType,
		EntityID:  entityID,
		TeamID:    teamID,
		Version:   time.Now().UnixNano(),
		Timestamp: time.Now().UTC(),
	}
}

// Marshal serializes the event to JSON
func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEvent deserializes an event from JSON
func UnmarshalEvent(data []byte) (Event, error) {
	var e Event
	err := json.Unmarshal(data, &e)
	return e, err
}

// ChannelForTeam returns the Redis channel name for team rule updates
func ChannelForTeam(teamID string) string {
	return "team:" + teamID + ":rules"
}

// ChannelForTeamCategories returns the Redis channel for team category updates
func ChannelForTeamCategories(teamID string) string {
	return "team:" + teamID + ":categories"
}

// ChannelBroadcast returns the broadcast channel for all workers
const ChannelBroadcast = "broadcast:all"

// ChannelForAgent returns the channel for direct agent messages
func ChannelForAgent(agentID string) string {
	return "agent:" + agentID + ":direct"
}
