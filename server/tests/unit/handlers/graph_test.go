package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
)

type mockGraphTeamService struct {
	teams []domain.Team
}

func (m *mockGraphTeamService) List() ([]domain.Team, error) {
	return m.teams, nil
}

type mockGraphUserService struct {
	users []domain.User
}

func (m *mockGraphUserService) List(teamID string, activeOnly bool) ([]domain.User, error) {
	return m.users, nil
}

func (m *mockGraphUserService) CountByTeam(teamID string) (int, error) {
	count := 0
	for _, u := range m.users {
		if u.TeamID != nil && *u.TeamID == teamID {
			count++
		}
	}
	return count, nil
}

type mockGraphRuleService struct {
	rules []domain.Rule
}

func (m *mockGraphRuleService) ListAll() ([]domain.Rule, error) {
	return m.rules, nil
}

func TestGraphHandler_Get(t *testing.T) {
	teamID := "team-1"
	teams := []domain.Team{{ID: teamID, Name: "Platform"}}
	users := []domain.User{{ID: "user-1", Name: "Alice", Email: "alice@test.com", TeamID: &teamID}}
	rules := []domain.Rule{{
		ID:          "rule-1",
		Name:        "Test Rule",
		Status:      domain.RuleStatusApproved,
		TargetTeams: []string{teamID},
	}}

	handler := handlers.NewGraphHandler(
		&mockGraphTeamService{teams: teams},
		&mockGraphUserService{users: users},
		&mockGraphRuleService{rules: rules},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response handlers.GraphResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Teams) != 1 {
		t.Errorf("expected 1 team, got %d", len(response.Teams))
	}
	if len(response.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(response.Users))
	}
	if len(response.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(response.Rules))
	}
}
