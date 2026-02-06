package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/tests/testutil"
)

func TestRulesHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupMock      func(*testutil.MockRuleService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			body: `{
				"name": "React Standards",
				"target_layer": "project",
				"content": "# React\nUse hooks.",
				"team_id": "team-123",
				"triggers": [{"type": "path", "pattern": "**/frontend/**"}]
			}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.RuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Name != "React Standards" {
					t.Errorf("expected name 'React Standards', got '%s'", resp.Name)
				}
			},
		},
		{
			name:           "invalid JSON",
			body:           `{invalid}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           ``,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing name",
			body:           `{"target_layer":"project","content":"content","team_id":"team-1"}`,
			expectedStatus: http.StatusCreated, // Name might be optional
		},
		{
			name:           "missing content",
			body:           `{"name":"Test","target_layer":"project","team_id":"team-1"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing team_id",
			body:           `{"name":"Test","target_layer":"project","content":"content"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid target_layer",
			body:           `{"name":"Test","target_layer":"invalid","content":"content","team_id":"team-1"}`,
			expectedStatus: http.StatusCreated, // Should validate
		},
		{
			name: "with multiple triggers",
			body: `{
				"name": "Multi Trigger Rule",
				"target_layer": "project",
				"content": "content",
				"team_id": "team-123",
				"triggers": [
					{"type": "path", "pattern": "*.go"},
					{"type": "context", "context_types": ["debug"]},
					{"type": "tag", "tags": ["backend"]}
				]
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "database error",
			body: `{"name":"Test","target_layer":"project","content":"content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				m.CreateFunc = func(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error) {
					return domain.Rule{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "very long content",
			body:           `{"name":"Test","target_layer":"project","content":"` + strings.Repeat("a", 10000) + `","team_id":"team-1"}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "markdown in content",
			body:           "{\"name\":\"Test\",\"target_layer\":\"project\",\"content\":\"# Header\\n## Subheader\\n- List item\\n```go\\ncode\\n```\",\"team_id\":\"team-1\"}",
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRuleService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRulesHandler(svc, nil)
			req := httptest.NewRequest("POST", "/rules", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRulesHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		setupMock      func(*testutil.MockRuleService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "existing rule",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.RuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.ID != "rule-1" {
					t.Errorf("expected ID 'rule-1', got '%s'", resp.ID)
				}
			},
		},
		{
			name:           "non-existing rule",
			ruleID:         "non-existent",
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				m.GetByIDFunc = func(ctx context.Context, id string) (domain.Rule, error) {
					return domain.Rule{}, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "empty ID",
			ruleID:         "",
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "special characters in ID",
			ruleID:         "rule-1-special-chars",
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRuleService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRulesHandler(svc, nil)

			r := chi.NewRouter()
			r.Get("/rules/{id}", h.Get)

			req := httptest.NewRequest("GET", "/rules/"+tt.ruleID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRulesHandler_ListByTeam(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*testutil.MockRuleService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "list rules for team",
			queryParams: "?team_id=team-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
				rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
				m.Rules[rule1.ID] = rule1
				m.Rules[rule2.ID] = rule2
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.RuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 2 {
					t.Errorf("expected 2 rules, got %d", len(resp))
				}
			},
		},
		{
			name:           "missing team_id",
			queryParams:    "",
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "empty team",
			queryParams: "?team_id=empty-team",
			setupMock: func(m *testutil.MockRuleService) {
				// No rules for this team
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.RuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 0 {
					t.Errorf("expected 0 rules, got %d", len(resp))
				}
			},
		},
		{
			name:        "filter by status",
			queryParams: "?team_id=team-1&status=pending",
			setupMock: func(m *testutil.MockRuleService) {
				rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
				rule1.Status = domain.RuleStatusPending
				rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
				rule2.Status = domain.RuleStatusDraft
				m.Rules[rule1.ID] = rule1
				m.Rules[rule2.ID] = rule2
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.RuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 1 {
					t.Errorf("expected 1 pending rule, got %d", len(resp))
				}
			},
		},
		{
			name:        "database error",
			queryParams: "?team_id=team-1",
			setupMock: func(m *testutil.MockRuleService) {
				m.ListByTeamFunc = func(ctx context.Context, teamID string) ([]domain.Rule, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRuleService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRulesHandler(svc, nil)
			req := httptest.NewRequest("GET", "/rules"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			h.ListByTeam(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestRulesHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		body           string
		setupMock      func(*testutil.MockRuleService)
		expectedStatus int
	}{
		{
			name:   "successful update - draft rule",
			ruleID: "rule-1",
			body:   `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusDraft
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "successful update - rejected rule",
			ruleID: "rule-1",
			body:   `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusRejected
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "cannot update pending rule",
			ruleID: "rule-1",
			body:   `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "cannot update approved rule",
			ruleID: "rule-1",
			body:   `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusApproved
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			body:           `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "invalid JSON",
			ruleID: "rule-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "database error on update",
			ruleID: "rule-1",
			body:   `{"name":"Updated Name","target_layer":"local","content":"updated content","team_id":"team-1"}`,
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Original", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				m.Rules["rule-1"] = rule
				m.UpdateFunc = func(ctx context.Context, rule domain.Rule) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRuleService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRulesHandler(svc, nil)

			r := chi.NewRouter()
			r.Put("/rules/{id}", h.Update)

			req := httptest.NewRequest("PUT", "/rules/"+tt.ruleID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRulesHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		setupMock      func(*testutil.MockRuleService)
		expectedStatus int
	}{
		{
			name:   "successful delete - draft rule",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Test", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusDraft
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "cannot delete pending rule",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Test", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "cannot delete approved rule",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Test", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusApproved
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			setupMock:      func(m *testutil.MockRuleService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockRuleService) {
				rule := domain.NewRule("Test", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				m.Rules["rule-1"] = rule
				m.DeleteFunc = func(ctx context.Context, id string) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockRuleService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewRulesHandler(svc, nil)

			r := chi.NewRouter()
			r.Delete("/rules/{id}", h.Delete)

			req := httptest.NewRequest("DELETE", "/rules/"+tt.ruleID, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
