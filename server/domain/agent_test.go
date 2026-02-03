package domain_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewAgentCreatesValidAgent(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")

	if agent.MachineID != "machine-abc123" {
		t.Errorf("expected machine ID 'machine-abc123', got '%s'", agent.MachineID)
	}
	if agent.Status != domain.AgentStatusOnline {
		t.Errorf("expected status 'online', got '%s'", agent.Status)
	}
}

func TestAgentIsStaleAfterThreshold(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")
	agent.LastHeartbeat = time.Now().Add(-2 * time.Hour)

	staleThreshold := 1 * time.Hour
	if !agent.IsStale(staleThreshold) {
		t.Error("expected agent to be stale after threshold")
	}
}

func TestAgentIsNotStaleWithinThreshold(t *testing.T) {
	agent := domain.NewAgent("machine-abc123", "user-456")
	agent.LastHeartbeat = time.Now().Add(-30 * time.Minute)

	staleThreshold := 1 * time.Hour
	if agent.IsStale(staleThreshold) {
		t.Error("expected agent NOT to be stale within threshold")
	}
}
