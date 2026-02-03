package domain

import (
	"time"

	"github.com/google/uuid"
)

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusStale   AgentStatus = "stale"
	AgentStatusOffline AgentStatus = "offline"
)

type Agent struct {
	ID                  string      `json:"id"`
	MachineID           string      `json:"machine_id"`
	UserID              string      `json:"user_id"`
	Status              AgentStatus `json:"status"`
	LastHeartbeat       time.Time   `json:"last_heartbeat"`
	CachedConfigVersion int         `json:"cached_config_version"`
	CreatedAt           time.Time   `json:"created_at"`
}

func NewAgent(machineID, userID string) Agent {
	now := time.Now()
	return Agent{
		ID:                  uuid.New().String(),
		MachineID:           machineID,
		UserID:              userID,
		Status:              AgentStatusOnline,
		LastHeartbeat:       now,
		CachedConfigVersion: 0,
		CreatedAt:           now,
	}
}

func (a Agent) IsStale(threshold time.Duration) bool {
	return time.Since(a.LastHeartbeat) > threshold
}

func (a *Agent) UpdateHeartbeat() {
	a.LastHeartbeat = time.Now()
	a.Status = AgentStatusOnline
}

func (a *Agent) MarkStale() {
	a.Status = AgentStatusStale
}

func (a *Agent) MarkOffline() {
	a.Status = AgentStatusOffline
}
