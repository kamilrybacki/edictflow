package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/middleware"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/response"
	"github.com/kamilrybacki/edictflow/server/services/approvals"
)

// teamIDMatches checks if a *string TeamID matches a string value
func teamIDMatches(teamIDPtr *string, teamID string) bool {
	if teamIDPtr == nil {
		return teamID == ""
	}
	return *teamIDPtr == teamID
}

type mockApprovalsService struct {
	rules           map[string]domain.Rule
	approvalRecords map[string][]domain.RuleApproval
}

func newMockApprovalsService() *mockApprovalsService {
	return &mockApprovalsService{
		rules:           make(map[string]domain.Rule),
		approvalRecords: make(map[string][]domain.RuleApproval),
	}
}

func (m *mockApprovalsService) SubmitRule(ctx context.Context, ruleID string) error {
	rule, ok := m.rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusDraft {
		return approvals.ErrCannotSubmit
	}
	rule.Status = domain.RuleStatusPending
	m.rules[ruleID] = rule
	return nil
}

func (m *mockApprovalsService) ApproveRule(ctx context.Context, ruleID, userID, comment string) error {
	rule, ok := m.rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusPending {
		return approvals.ErrNotPending
	}
	m.approvalRecords[ruleID] = append(m.approvalRecords[ruleID], domain.RuleApproval{
		ID:        "approval-1",
		RuleID:    ruleID,
		UserID:    userID,
		Decision:  domain.ApprovalDecisionApproved,
		Comment:   comment,
		CreatedAt: time.Now(),
	})
	return nil
}

func (m *mockApprovalsService) RejectRule(ctx context.Context, ruleID, userID, comment string) error {
	rule, ok := m.rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	if rule.Status != domain.RuleStatusPending {
		return approvals.ErrNotPending
	}
	rule.Status = domain.RuleStatusRejected
	m.rules[ruleID] = rule
	return nil
}

func (m *mockApprovalsService) GetApprovalStatus(ctx context.Context, ruleID string) (approvals.ApprovalStatus, error) {
	if _, ok := m.rules[ruleID]; !ok {
		return approvals.ApprovalStatus{}, approvals.ErrRuleNotFound
	}
	return approvals.ApprovalStatus{
		RuleID:        ruleID,
		Status:        m.rules[ruleID].Status,
		RequiredCount: 2,
		CurrentCount:  len(m.approvalRecords[ruleID]),
		Approvals:     m.approvalRecords[ruleID],
	}, nil
}

func (m *mockApprovalsService) GetPendingRules(ctx context.Context, teamID string) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.rules {
		if rule.Status == domain.RuleStatusPending && teamIDMatches(rule.TeamID, teamID) {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (m *mockApprovalsService) GetPendingRulesByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.rules {
		if rule.Status == domain.RuleStatusPending && rule.TargetLayer == scope {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (m *mockApprovalsService) ResetRule(ctx context.Context, ruleID string) error {
	rule, ok := m.rules[ruleID]
	if !ok {
		return approvals.ErrRuleNotFound
	}
	rule.Status = domain.RuleStatusDraft
	m.rules[ruleID] = rule
	delete(m.approvalRecords, ruleID)
	return nil
}

func withUserContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
	return req.WithContext(ctx)
}

func TestApprovalsHandler_Submit(t *testing.T) {
	svc := newMockApprovalsService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	svc.rules[rule.ID] = rule

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Post("/approvals/rules/{ruleId}/submit", h.Submit)

	req := httptest.NewRequest("POST", "/approvals/rules/"+rule.ID+"/submit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if svc.rules[rule.ID].Status != domain.RuleStatusPending {
		t.Errorf("expected status 'pending', got '%s'", svc.rules[rule.ID].Status)
	}
}

func TestApprovalsHandler_Submit_NotFound(t *testing.T) {
	svc := newMockApprovalsService()
	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Post("/approvals/rules/{ruleId}/submit", h.Submit)

	req := httptest.NewRequest("POST", "/approvals/rules/non-existent/submit", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestApprovalsHandler_Approve(t *testing.T) {
	svc := newMockApprovalsService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Status = domain.RuleStatusPending
	svc.rules[rule.ID] = rule

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Post("/approvals/rules/{ruleId}/approve", h.Approve)

	body := `{"comment":"LGTM"}`
	req := httptest.NewRequest("POST", "/approvals/rules/"+rule.ID+"/approve", bytes.NewBufferString(body))
	req = withUserContext(req, "approver-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(svc.approvalRecords[rule.ID]) != 1 {
		t.Error("expected approval to be recorded")
	}
}

func TestApprovalsHandler_Reject(t *testing.T) {
	svc := newMockApprovalsService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Status = domain.RuleStatusPending
	svc.rules[rule.ID] = rule

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Post("/approvals/rules/{ruleId}/reject", h.Reject)

	body := `{"comment":"Needs improvement"}`
	req := httptest.NewRequest("POST", "/approvals/rules/"+rule.ID+"/reject", bytes.NewBufferString(body))
	req = withUserContext(req, "approver-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if svc.rules[rule.ID].Status != domain.RuleStatusRejected {
		t.Errorf("expected status 'rejected', got '%s'", svc.rules[rule.ID].Status)
	}
}

func TestApprovalsHandler_Reject_RequiresComment(t *testing.T) {
	svc := newMockApprovalsService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Status = domain.RuleStatusPending
	svc.rules[rule.ID] = rule

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Post("/approvals/rules/{ruleId}/reject", h.Reject)

	body := `{}`
	req := httptest.NewRequest("POST", "/approvals/rules/"+rule.ID+"/reject", bytes.NewBufferString(body))
	req = withUserContext(req, "approver-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestApprovalsHandler_GetStatus(t *testing.T) {
	svc := newMockApprovalsService()
	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Status = domain.RuleStatusPending
	svc.rules[rule.ID] = rule

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Get("/approvals/rules/{ruleId}", h.GetStatus)

	req := httptest.NewRequest("GET", "/approvals/rules/"+rule.ID, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var apiResp response.APIResponse
	json.NewDecoder(rec.Body).Decode(&apiResp)

	if !apiResp.Success {
		t.Errorf("expected success to be true")
	}

	// Convert data to ApprovalStatusResponse
	dataBytes, _ := json.Marshal(apiResp.Data)
	var resp handlers.ApprovalStatusResponse
	json.Unmarshal(dataBytes, &resp)

	if resp.RuleID != rule.ID {
		t.Errorf("expected rule ID '%s', got '%s'", rule.ID, resp.RuleID)
	}
	if resp.RequiredCount != 2 {
		t.Errorf("expected required count 2, got %d", resp.RequiredCount)
	}
}

func TestApprovalsHandler_ListPending(t *testing.T) {
	svc := newMockApprovalsService()
	rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
	rule1.Status = domain.RuleStatusPending
	svc.rules[rule1.ID] = rule1

	rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
	svc.rules[rule2.ID] = rule2 // Draft, not pending

	h := handlers.NewApprovalsHandler(svc)

	r := chi.NewRouter()
	r.Get("/approvals/pending", h.ListPending)

	req := httptest.NewRequest("GET", "/approvals/pending?scope=local", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var apiResp response.APIResponse
	json.NewDecoder(rec.Body).Decode(&apiResp)

	if !apiResp.Success {
		t.Errorf("expected success to be true")
	}

	// Convert data to []PendingRuleResponse
	dataBytes, _ := json.Marshal(apiResp.Data)
	var resp []handlers.PendingRuleResponse
	json.Unmarshal(dataBytes, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 pending rule, got %d", len(resp))
	}
}
