//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamilrybacki/claudeception/server/adapters/postgres"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/claudeception/server/integration/testhelpers"
)

// testTeamService implements handlers.TeamService for integration tests
type testTeamService struct {
	teamDB *postgres.TeamDB
}

func (s *testTeamService) Create(ctx context.Context, name string) (domain.Team, error) {
	team := domain.NewTeam(name)
	if err := s.teamDB.CreateTeam(ctx, team); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}

func (s *testTeamService) GetByID(ctx context.Context, id string) (domain.Team, error) {
	return s.teamDB.GetTeam(ctx, id)
}

func (s *testTeamService) List(ctx context.Context) ([]domain.Team, error) {
	return s.teamDB.ListTeams(ctx)
}

func (s *testTeamService) Update(ctx context.Context, team domain.Team) error {
	return s.teamDB.UpdateTeam(ctx, team)
}

func (s *testTeamService) Delete(ctx context.Context, id string) error {
	return s.teamDB.DeleteTeam(ctx, id)
}

// testRuleService implements handlers.RuleService for integration tests
type testRuleService struct {
	ruleDB *postgres.RuleDB
}

func (s *testRuleService) Create(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error) {
	triggers := make([]domain.Trigger, len(req.Triggers))
	for i, t := range req.Triggers {
		triggers[i] = domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		}
	}

	rule := domain.NewRule(req.Name, domain.TargetLayer(req.TargetLayer), req.Content, triggers, req.TeamID)
	if err := s.ruleDB.CreateRule(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}

func (s *testRuleService) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	return s.ruleDB.GetRule(ctx, id)
}

func (s *testRuleService) ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	return s.ruleDB.ListRulesByTeam(ctx, teamID)
}

func (s *testRuleService) Update(ctx context.Context, rule domain.Rule) error {
	return s.ruleDB.UpdateRule(ctx, rule)
}

func (s *testRuleService) Delete(ctx context.Context, id string) error {
	return s.ruleDB.DeleteRule(ctx, id)
}

func setupTestRouter() *httptest.Server {
	teamDB := postgres.NewTeamDB(testPool)
	ruleDB := postgres.NewRuleDB(testPool)

	router := api.NewRouter(api.Config{
		JWTSecret:   testhelpers.TestJWTSecret,
		TeamService: &testTeamService{teamDB: teamDB},
		RuleService: &testRuleService{ruleDB: ruleDB},
	})

	return httptest.NewServer(router)
}

func TestAPI_HealthCheck(t *testing.T) {
	server := setupTestRouter()
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAPI_CreateTeam(t *testing.T) {
	resetDB(t)
	server := setupTestRouter()
	defer server.Close()

	authHeader, err := testhelpers.AuthHeader("user-1", "team-1")
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	// Create team
	body := `{"name": "API Test Team"}`
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/teams/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var team domain.Team
	if err := json.NewDecoder(resp.Body).Decode(&team); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if team.Name != "API Test Team" {
		t.Errorf("Expected team name 'API Test Team', got %s", team.Name)
	}
	if team.ID == "" {
		t.Error("Expected team ID to be set")
	}

	// Verify it's in the database
	ctx := context.Background()
	dbTeam, err := postgres.NewTeamDB(testPool).GetTeam(ctx, team.ID)
	if err != nil {
		t.Fatalf("Failed to get team from DB: %v", err)
	}
	if dbTeam.Name != team.Name {
		t.Errorf("DB team name mismatch: expected %s, got %s", team.Name, dbTeam.Name)
	}
}

func TestAPI_ListTeams(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	server := setupTestRouter()
	defer server.Close()

	// Create teams via fixtures
	for i := 0; i < 3; i++ {
		if _, err := testFixtures.CreateTeam(ctx, "Fixture Team"); err != nil {
			t.Fatalf("Failed to create fixture team: %v", err)
		}
	}

	authHeader, err := testhelpers.AuthHeader("user-1", "team-1")
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/teams/", nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var teams []domain.Team
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(teams) != 3 {
		t.Errorf("Expected 3 teams, got %d", len(teams))
	}
}

func TestAPI_GetTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	server := setupTestRouter()
	defer server.Close()

	// Create team via fixture
	team, err := testFixtures.CreateTeam(ctx, "Get Test Team")
	if err != nil {
		t.Fatalf("Failed to create fixture team: %v", err)
	}

	authHeader, err := testhelpers.AuthHeader("user-1", team.ID)
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/teams/"+team.ID, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var retrieved domain.Team
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrieved.ID != team.ID {
		t.Errorf("Expected team ID %s, got %s", team.ID, retrieved.ID)
	}
}

func TestAPI_DeleteTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	server := setupTestRouter()
	defer server.Close()

	team, err := testFixtures.CreateTeam(ctx, "Delete Test Team")
	if err != nil {
		t.Fatalf("Failed to create fixture team: %v", err)
	}

	authHeader, err := testhelpers.AuthHeader("user-1", team.ID)
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	req, _ := http.NewRequest("DELETE", server.URL+"/api/v1/teams/"+team.ID, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify deletion
	teamDB := postgres.NewTeamDB(testPool)
	_, err = teamDB.GetTeam(ctx, team.ID)
	if err == nil {
		t.Error("Expected team to be deleted")
	}
}

func TestAPI_CreateRule(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	server := setupTestRouter()
	defer server.Close()

	// Create team first
	team, err := testFixtures.CreateTeam(ctx, "Rule Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	authHeader, err := testhelpers.AuthHeader("user-1", team.ID)
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	body := `{
		"name": "API Test Rule",
		"target_layer": "project",
		"content": "Test rule content",
		"team_id": "` + team.ID + `",
		"triggers": [{"type": "path", "pattern": "*.go"}]
	}`

	req, _ := http.NewRequest("POST", server.URL+"/api/v1/rules/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var rule domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if rule.Name != "API Test Rule" {
		t.Errorf("Expected rule name 'API Test Rule', got %s", rule.Name)
	}
	if rule.TeamID != team.ID {
		t.Errorf("Expected team ID %s, got %s", team.ID, rule.TeamID)
	}
}

func TestAPI_ListRulesByTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()
	server := setupTestRouter()
	defer server.Close()

	team, err := testFixtures.CreateTeam(ctx, "Rule List Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Create rules
	for i := 0; i < 2; i++ {
		if _, err := testFixtures.CreateRule(ctx, "Test Rule", team.ID); err != nil {
			t.Fatalf("Failed to create rule: %v", err)
		}
	}

	authHeader, err := testhelpers.AuthHeader("user-1", team.ID)
	if err != nil {
		t.Fatalf("Failed to generate auth header: %v", err)
	}

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/rules/?team_id="+team.ID, nil)
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var rules []domain.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
}

func TestAPI_Unauthorized(t *testing.T) {
	resetDB(t)
	server := setupTestRouter()
	defer server.Close()

	// Request without auth header
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/teams/", nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAPI_InvalidToken(t *testing.T) {
	resetDB(t)
	server := setupTestRouter()
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/teams/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}
