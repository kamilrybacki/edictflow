package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
)

type mockTeamService struct {
	teams map[string]domain.Team
}

func newMockTeamService() *mockTeamService {
	return &mockTeamService{teams: make(map[string]domain.Team)}
}

func (m *mockTeamService) Create(ctx context.Context, name string) (domain.Team, error) {
	team := domain.NewTeam(name)
	m.teams[team.ID] = team
	return team, nil
}

func (m *mockTeamService) GetByID(ctx context.Context, id string) (domain.Team, error) {
	team, ok := m.teams[id]
	if !ok {
		return domain.Team{}, handlers.ErrNotFound
	}
	return team, nil
}

func (m *mockTeamService) List(ctx context.Context) ([]domain.Team, error) {
	var result []domain.Team
	for _, t := range m.teams {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockTeamService) Update(ctx context.Context, team domain.Team) error {
	m.teams[team.ID] = team
	return nil
}

func (m *mockTeamService) Delete(ctx context.Context, id string) error {
	delete(m.teams, id)
	return nil
}

func TestCreateTeamHandler(t *testing.T) {
	svc := newMockTeamService()
	h := handlers.NewTeamsHandler(svc)

	body := `{"name": "Engineering"}`
	req := httptest.NewRequest("POST", "/teams", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp domain.Team
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", resp.Name)
	}
}

func TestListTeamsHandler(t *testing.T) {
	svc := newMockTeamService()
	svc.Create(context.Background(), "Team 1")
	svc.Create(context.Background(), "Team 2")

	h := handlers.NewTeamsHandler(svc)

	req := httptest.NewRequest("GET", "/teams", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp []domain.Team
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 teams, got %d", len(resp))
	}
}
