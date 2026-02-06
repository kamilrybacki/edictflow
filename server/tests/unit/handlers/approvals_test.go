package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/tests/testutil"
)

func withUserContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
	return req.WithContext(ctx)
}

func withTeamContext(req *http.Request, teamID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.TeamIDContextKey, teamID)
	return req.WithContext(ctx)
}

func TestApprovalsHandler_Submit(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
	}{
		{
			name:   "successful submission",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "already pending",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "already approved",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusApproved
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "database error",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				m.SubmitFunc = func(ctx context.Context, ruleID string) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Post("/approvals/rules/{ruleId}/submit", h.Submit)

			req := httptest.NewRequest("POST", "/approvals/rules/"+tt.ruleID+"/submit", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestApprovalsHandler_Approve(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		userID         string
		body           string
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
	}{
		{
			name:   "successful approval",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":"LGTM"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "approval without comment",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			userID:         "approver-1",
			body:           `{"comment":"LGTM"}`,
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "not pending",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":"LGTM"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusDraft
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "missing user context",
			ruleID: "rule-1",
			userID: "", // No user
			body:   `{"comment":"LGTM"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "invalid JSON",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "self-approval attempt",
			ruleID: "rule-1",
			userID: "creator-1",
			body:   `{"comment":"Self approve"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				creatorID := "creator-1"
				rule.CreatedBy = &creatorID
				m.Rules["rule-1"] = rule
				m.ApproveFunc = func(ctx context.Context, ruleID, userID, comment string) error {
					if userID == "creator-1" {
						return errors.New("cannot approve own rule")
					}
					return nil
				}
			},
			expectedStatus: http.StatusInternalServerError, // Handler returns 500 for unrecognized errors
		},
		{
			name:   "duplicate approval",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":"Already approved"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
				m.ApproveFunc = func(ctx context.Context, ruleID, userID, comment string) error {
					return errors.New("user has already voted on this rule") // Must match exact handler error
				}
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Post("/approvals/rules/{ruleId}/approve", h.Approve)

			req := httptest.NewRequest("POST", "/approvals/rules/"+tt.ruleID+"/approve", bytes.NewBufferString(tt.body))
			if tt.userID != "" {
				req = withUserContext(req, tt.userID)
			}
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestApprovalsHandler_Reject(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		userID         string
		body           string
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
	}{
		{
			name:   "successful rejection with comment",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":"Needs improvement"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "rejection requires comment",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "rejection with empty comment",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":""}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			userID:         "approver-1",
			body:           `{"comment":"Needs improvement"}`,
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "not pending",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{"comment":"Needs improvement"}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusApproved
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "invalid JSON",
			ruleID: "rule-1",
			userID: "approver-1",
			body:   `{invalid}`,
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Post("/approvals/rules/{ruleId}/reject", h.Reject)

			req := httptest.NewRequest("POST", "/approvals/rules/"+tt.ruleID+"/reject", bytes.NewBufferString(tt.body))
			if tt.userID != "" {
				req = withUserContext(req, tt.userID)
			}
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestApprovalsHandler_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "get approval status",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.ApprovalStatusResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.RuleID != "rule-1" {
					t.Errorf("expected rule ID 'rule-1', got '%s'", resp.RuleID)
				}
				if resp.RequiredCount != 2 {
					t.Errorf("expected required count 2, got %d", resp.RequiredCount)
				}
			},
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "rule with approvals",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusPending
				m.Rules["rule-1"] = rule
				m.ApprovalRecords["rule-1"] = []domain.RuleApproval{
					{ID: "approval-1", UserID: "user-1", Decision: domain.ApprovalDecisionApproved},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp handlers.ApprovalStatusResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.CurrentCount != 1 {
					t.Errorf("expected current count 1, got %d", resp.CurrentCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Get("/approvals/rules/{ruleId}", h.GetStatus)

			req := httptest.NewRequest("GET", "/approvals/rules/"+tt.ruleID, nil)
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

func TestApprovalsHandler_ListPending(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		teamID         string // Team ID to set in context
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "list pending by scope",
			queryParams: "?scope=local",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.Status = domain.RuleStatusPending
				m.Rules[rule.ID] = rule
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.PendingRuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 1 {
					t.Errorf("expected 1 pending rule, got %d", len(resp))
				}
			},
		},
		{
			name:   "list pending by team",
			teamID: "team-1", // Team ID comes from context, not query param
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.Status = domain.RuleStatusPending
				m.Rules[rule.ID] = rule
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no pending rules",
			queryParams:    "?scope=local",
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.PendingRuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 0 {
					t.Errorf("expected 0 pending rules, got %d", len(resp))
				}
			},
		},
		{
			name:        "filter excludes non-pending",
			queryParams: "?scope=local",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule1 := domain.NewRule("Pending Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule1.Status = domain.RuleStatusPending
				m.Rules[rule1.ID] = rule1

				rule2 := domain.NewRule("Draft Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule2.Status = domain.RuleStatusDraft
				m.Rules[rule2.ID] = rule2
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp []handlers.PendingRuleResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if len(resp) != 1 {
					t.Errorf("expected 1 pending rule, got %d", len(resp))
				}
			},
		},
		{
			name:           "missing scope and team",
			queryParams:    "",
			teamID:         "", // No team in context
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Get("/approvals/pending", h.ListPending)

			req := httptest.NewRequest("GET", "/approvals/pending"+tt.queryParams, nil)
			if tt.teamID != "" {
				req = withTeamContext(req, tt.teamID)
			}
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

func TestApprovalsHandler_Reset(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		setupMock      func(*testutil.MockApprovalsService)
		expectedStatus int
	}{
		{
			name:   "successful reset",
			ruleID: "rule-1",
			setupMock: func(m *testutil.MockApprovalsService) {
				rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
				rule.ID = "rule-1"
				rule.Status = domain.RuleStatusRejected
				m.Rules["rule-1"] = rule
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "rule not found",
			ruleID:         "non-existent",
			setupMock:      func(m *testutil.MockApprovalsService) {},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testutil.NewMockApprovalsService()
			if tt.setupMock != nil {
				tt.setupMock(svc)
			}

			h := handlers.NewApprovalsHandler(svc)

			r := chi.NewRouter()
			r.Post("/approvals/rules/{ruleId}/reset", h.Reset)

			req := httptest.NewRequest("POST", "/approvals/rules/"+tt.ruleID+"/reset", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
