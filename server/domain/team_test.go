package domain_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
)

func TestNewTeamCreatesValidTeam(t *testing.T) {
	team := domain.NewTeam("Engineering")

	if team.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", team.Name)
	}
	if team.ID == "" {
		t.Error("expected non-empty ID")
	}
	if team.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestTeamValidateRejectsEmptyName(t *testing.T) {
	team := domain.Team{
		ID:        "test-id",
		Name:      "",
		CreatedAt: time.Now(),
	}

	err := team.Validate()
	if err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestNewTeam_DefaultsInheritGlobalRules(t *testing.T) {
	team := domain.NewTeam("Engineering")

	if !team.Settings.InheritGlobalRules {
		t.Error("expected InheritGlobalRules to default to true")
	}
}
