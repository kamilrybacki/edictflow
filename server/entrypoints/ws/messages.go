package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	// Server -> Agent
	TypeConfigUpdate MessageType = "config_update"
	TypeSyncRequest  MessageType = "sync_request"
	TypeAck          MessageType = "ack"

	// Agent -> Server
	TypeHeartbeat       MessageType = "heartbeat"
	TypeDriftReport     MessageType = "drift_report"
	TypeContextDetected MessageType = "context_detected"
	TypeSyncComplete    MessageType = "sync_complete"
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

// Server -> Agent payloads
type ConfigUpdatePayload struct {
	Rules   []RulePayload `json:"rules"`
	Version int           `json:"version"`
}

type RulePayload struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Content     string          `json:"content"`
	TargetLayer string          `json:"target_layer"`
	Triggers    json.RawMessage `json:"triggers"`
}

type SyncRequestPayload struct {
	ProjectPaths []string `json:"project_paths"`
}

type AckPayload struct {
	RefID string `json:"ref_id"`
}

// Agent -> Server payloads
type HeartbeatPayload struct {
	Status         string   `json:"status"`
	CachedVersion  int      `json:"cached_version"`
	ActiveProjects []string `json:"active_projects"`
}

type DriftReportPayload struct {
	ProjectPath  string `json:"project_path"`
	ExpectedHash string `json:"expected_hash"`
	ActualHash   string `json:"actual_hash"`
	Diff         string `json:"diff"`
}

type ContextDetectedPayload struct {
	ProjectPath     string   `json:"project_path"`
	DetectedContext []string `json:"detected_context"`
	DetectedTags    []string `json:"detected_tags"`
}

type SyncCompletePayload struct {
	ProjectPath  string   `json:"project_path"`
	FilesWritten []string `json:"files_written"`
}
