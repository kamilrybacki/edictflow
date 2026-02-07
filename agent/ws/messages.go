// agent/ws/messages.go
package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	// Server -> Agent
	TypeConfigUpdate     MessageType = "config_update"
	TypeSyncRequest      MessageType = "sync_request"
	TypeAck              MessageType = "ack"
	TypeChangeApproved   MessageType = "change_approved"
	TypeChangeRejected   MessageType = "change_rejected"
	TypeExceptionGranted MessageType = "exception_granted"
	TypeExceptionDenied  MessageType = "exception_denied"

	// Agent -> Server
	TypeHeartbeat        MessageType = "heartbeat"
	TypeDriftReport      MessageType = "drift_report"
	TypeContextDetected  MessageType = "context_detected"
	TypeSyncComplete     MessageType = "sync_complete"
	TypeChangeDetected   MessageType = "change_detected"
	TypeChangeUpdated    MessageType = "change_updated"
	TypeExceptionRequest MessageType = "exception_request"
	TypeRevertComplete   MessageType = "revert_complete"
)

type Message struct {
	Type      MessageType     `json:"type"`
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

func NewMessage(msgType MessageType, payload interface{}) (Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Type:      msgType,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}

type HeartbeatPayload struct {
	Status         string   `json:"status"`
	CachedVersion  int      `json:"cached_version"`
	ActiveProjects []string `json:"active_projects"`
	AgentID        string   `json:"agent_id,omitempty"`
	TeamID         string   `json:"team_id,omitempty"`
	Hostname       string   `json:"hostname,omitempty"`
	Version        string   `json:"version,omitempty"`
	OS             string   `json:"os,omitempty"`
	ConnectedAt    string   `json:"connected_at,omitempty"`
}

type ConfigUpdatePayload struct {
	Rules   []RulePayload `json:"rules"`
	Version int           `json:"version"`
}

type RulePayload struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Content               string          `json:"content"`
	TargetLayer           string          `json:"target_layer"`
	Triggers              json.RawMessage `json:"triggers"`
	EnforcementMode       string          `json:"enforcement_mode"`
	TemporaryTimeoutHours int             `json:"temporary_timeout_hours"`
}

type AckPayload struct {
	RefID string `json:"ref_id"`
}

type ChangeDetectedPayload struct {
	RuleID          string `json:"rule_id"`
	FilePath        string `json:"file_path"`
	OriginalHash    string `json:"original_hash"`
	ModifiedHash    string `json:"modified_hash"`
	Diff            string `json:"diff"`
	EnforcementMode string `json:"enforcement_mode"`
}

type ChangeApprovedPayload struct {
	ChangeID string `json:"change_id"`
	RuleID   string `json:"rule_id"`
}

type ChangeRejectedPayload struct {
	ChangeID     string `json:"change_id"`
	RuleID       string `json:"rule_id"`
	RevertToHash string `json:"revert_to_hash"`
}
