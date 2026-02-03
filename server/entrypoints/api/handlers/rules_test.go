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

type mockRuleService struct {
	rules map[string]domain.Rule
}

func newMockRuleService() *mockRuleService {
	return &mockRuleService{rules: make(map[string]domain.Rule)}
}

func (m *mockRuleService) Create(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error) {
	var triggers []domain.Trigger
	for _, t := range req.Triggers {
		triggers = append(triggers, domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		})
	}
	rule := domain.NewRule(req.Name, domain.TargetLayer(req.TargetLayer), req.Content, triggers, req.TeamID)
	m.rules[rule.ID] = rule
	return rule, nil
}

func (m *mockRuleService) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	rule, ok := m.rules[id]
	if !ok {
		return domain.Rule{}, handlers.ErrNotFound
	}
	return rule, nil
}

func (m *mockRuleService) ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		if r.TeamID == teamID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRuleService) Update(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleService) Delete(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func TestCreateRuleHandler(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc)

	body := `{
		"name": "React Standards",
		"target_layer": "project",
		"content": "# React\nUse hooks.",
		"team_id": "team-123",
		"triggers": [{"type": "path", "pattern": "**/frontend/**"}]
	}`
	req := httptest.NewRequest("POST", "/rules", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.Rule
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "React Standards" {
		t.Errorf("expected name 'React Standards', got '%s'", resp.Name)
	}
}
