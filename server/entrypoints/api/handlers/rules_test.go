package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
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

func (m *mockRuleService) ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		if r.TeamID == teamID && r.Status == status {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRuleService) ListByTargetLayer(ctx context.Context, targetLayer domain.TargetLayer) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		if r.TargetLayer == targetLayer {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRuleService) GetMergedContent(ctx context.Context, targetLayer domain.TargetLayer) (string, error) {
	return "<!-- MANAGED BY CLAUDECEPTION -->\n<!-- END CLAUDECEPTION -->", nil
}

func TestCreateRuleHandler(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

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

	var resp handlers.RuleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "React Standards" {
		t.Errorf("expected name 'React Standards', got '%s'", resp.Name)
	}
}

func TestCreateRuleHandler_InvalidBody(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

	req := httptest.NewRequest("POST", "/rules", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRulesHandler_Get(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Get("/rules/{id}", h.Get)

	req := httptest.NewRequest("GET", "/rules/"+rule.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp handlers.RuleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.ID != rule.ID {
		t.Errorf("expected ID '%s', got '%s'", rule.ID, resp.ID)
	}
}

func TestRulesHandler_Get_NotFound(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Get("/rules/{id}", h.Get)

	req := httptest.NewRequest("GET", "/rules/non-existent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRulesHandler_ListByTeam(t *testing.T) {
	svc := newMockRuleService()
	rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
	rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
	rule3 := domain.NewRule("Rule 3", domain.TargetLayerLocal, "content", nil, "team-2")
	svc.rules[rule1.ID] = rule1
	svc.rules[rule2.ID] = rule2
	svc.rules[rule3.ID] = rule3

	h := handlers.NewRulesHandler(svc, nil)

	req := httptest.NewRequest("GET", "/rules?team_id=team-1", nil)
	rec := httptest.NewRecorder()

	h.ListByTeam(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.RuleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 rules for team-1, got %d", len(resp))
	}
}

func TestRulesHandler_ListByTeam_MissingTeamID(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

	req := httptest.NewRequest("GET", "/rules", nil)
	rec := httptest.NewRecorder()

	h.ListByTeam(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRulesHandler_ListByStatus(t *testing.T) {
	svc := newMockRuleService()
	rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
	rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
	rule2.Status = domain.RuleStatusPending
	svc.rules[rule1.ID] = rule1
	svc.rules[rule2.ID] = rule2

	h := handlers.NewRulesHandler(svc, nil)

	req := httptest.NewRequest("GET", "/rules?team_id=team-1&status=pending", nil)
	rec := httptest.NewRecorder()

	h.ListByTeam(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []handlers.RuleResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 pending rule, got %d", len(resp))
	}
	if len(resp) > 0 && resp[0].Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", resp[0].Status)
	}
}

func TestRulesHandler_Update(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Original Name", domain.TargetLayerLocal, "original content", nil, "team-1")
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Put("/rules/{id}", h.Update)

	body := `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`
	req := httptest.NewRequest("PUT", "/rules/"+rule.ID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if svc.rules[rule.ID].Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", svc.rules[rule.ID].Name)
	}
}

func TestRulesHandler_Update_NotFound(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Put("/rules/{id}", h.Update)

	body := `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`
	req := httptest.NewRequest("PUT", "/rules/non-existent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRulesHandler_Update_NonDraftRule(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Original Name", domain.TargetLayerLocal, "original content", nil, "team-1")
	rule.Status = domain.RuleStatusPending // Not draft
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Put("/rules/{id}", h.Update)

	body := `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`
	req := httptest.NewRequest("PUT", "/rules/"+rule.ID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestRulesHandler_Update_RejectedRule(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Original Name", domain.TargetLayerLocal, "original content", nil, "team-1")
	rule.Status = domain.RuleStatusRejected // Rejected rules can be edited
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Put("/rules/{id}", h.Update)

	body := `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`
	req := httptest.NewRequest("PUT", "/rules/"+rule.ID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRulesHandler_Delete(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Delete("/rules/{id}", h.Delete)

	req := httptest.NewRequest("DELETE", "/rules/"+rule.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if _, ok := svc.rules[rule.ID]; ok {
		t.Error("expected rule to be deleted")
	}
}

func TestRulesHandler_Delete_NotFound(t *testing.T) {
	svc := newMockRuleService()
	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Delete("/rules/{id}", h.Delete)

	req := httptest.NewRequest("DELETE", "/rules/non-existent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRulesHandler_Delete_NonDraftRule(t *testing.T) {
	svc := newMockRuleService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Status = domain.RuleStatusApproved // Can't delete approved rules
	svc.rules[rule.ID] = rule

	h := handlers.NewRulesHandler(svc, nil)

	r := chi.NewRouter()
	r.Delete("/rules/{id}", h.Delete)

	req := httptest.NewRequest("DELETE", "/rules/"+rule.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}
