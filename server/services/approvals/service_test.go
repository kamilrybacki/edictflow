package approvals

import (
	"context"
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type mockRuleDB struct {
	rules map[string]domain.Rule
}

func (m *mockRuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	if rule, ok := m.rules[id]; ok {
		return rule, nil
	}
	return domain.Rule{}, ErrRuleNotFound
}

func (m *mockRuleDB) UpdateStatus(ctx context.Context, rule domain.Rule) error {
	if _, ok := m.rules[rule.ID]; !ok {
		return ErrRuleNotFound
	}
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleDB) ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.rules {
		if rule.TeamID == teamID && rule.Status == status {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (m *mockRuleDB) ListPendingByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, rule := range m.rules {
		if rule.TargetLayer == scope && rule.Status == domain.RuleStatusPending {
			result = append(result, rule)
		}
	}
	return result, nil
}

type mockApprovalDB struct {
	approvals map[string][]domain.RuleApproval
}

func (m *mockApprovalDB) Create(ctx context.Context, approval domain.RuleApproval) error {
	m.approvals[approval.RuleID] = append(m.approvals[approval.RuleID], approval)
	return nil
}

func (m *mockApprovalDB) ListByRule(ctx context.Context, ruleID string) ([]domain.RuleApproval, error) {
	return m.approvals[ruleID], nil
}

func (m *mockApprovalDB) CountApprovals(ctx context.Context, ruleID string) (int, error) {
	count := 0
	for _, a := range m.approvals[ruleID] {
		if a.Decision == domain.ApprovalDecisionApproved {
			count++
		}
	}
	return count, nil
}

func (m *mockApprovalDB) HasUserApproved(ctx context.Context, ruleID, userID string) (bool, error) {
	for _, a := range m.approvals[ruleID] {
		if a.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockApprovalDB) DeleteByRule(ctx context.Context, ruleID string) error {
	delete(m.approvals, ruleID)
	return nil
}

type mockApprovalConfigDB struct {
	configs map[domain.TargetLayer]domain.ApprovalConfig
}

func (m *mockApprovalConfigDB) GetForScope(ctx context.Context, scope domain.TargetLayer, teamID *string) (domain.ApprovalConfig, error) {
	if config, ok := m.configs[scope]; ok {
		return config, nil
	}
	return domain.ApprovalConfig{RequiredCount: 1, RequiredPermission: "approve_local"}, nil
}

type mockRoleDB struct {
	userPermissions map[string][]string
}

func (m *mockRoleDB) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return m.userPermissions[userID], nil
}

func newTestService() (*Service, *mockRuleDB, *mockApprovalDB) {
	ruleDB := &mockRuleDB{rules: make(map[string]domain.Rule)}
	approvalDB := &mockApprovalDB{approvals: make(map[string][]domain.RuleApproval)}
	configDB := &mockApprovalConfigDB{
		configs: map[domain.TargetLayer]domain.ApprovalConfig{
			domain.TargetLayerLocal:   {RequiredCount: 1, RequiredPermission: "approve_local"},
			domain.TargetLayerProject: {RequiredCount: 1, RequiredPermission: "approve_project"},
			domain.TargetLayerGlobal:  {RequiredCount: 2, RequiredPermission: "approve_global"},
		},
	}
	roleDB := &mockRoleDB{
		userPermissions: map[string][]string{
			"approver-1": {"approve_local", "approve_project", "approve_global"},
			"approver-2": {"approve_local", "approve_project", "approve_global"},
			"member-1":   {"create_rules"},
		},
	}
	svc := NewService(ruleDB, approvalDB, configDB, roleDB)
	return svc, ruleDB, approvalDB
}

func TestService_SubmitRule(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	ruleDB.rules[rule.ID] = rule

	err := svc.SubmitRule(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("SubmitRule() error = %v", err)
	}

	updated := ruleDB.rules[rule.ID]
	if updated.Status != domain.RuleStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", updated.Status)
	}
	if updated.SubmittedAt == nil {
		t.Error("Expected SubmittedAt to be set")
	}
}

func TestService_SubmitRule_AlreadyPending(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	err := svc.SubmitRule(context.Background(), rule.ID)
	if err != ErrCannotSubmit {
		t.Errorf("Expected ErrCannotSubmit, got %v", err)
	}
}

func TestService_ApproveRule(t *testing.T) {
	svc, ruleDB, approvalDB := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	err := svc.ApproveRule(context.Background(), rule.ID, "approver-1", "")
	if err != nil {
		t.Fatalf("ApproveRule() error = %v", err)
	}

	// Check approval was recorded
	if len(approvalDB.approvals[rule.ID]) != 1 {
		t.Errorf("Expected 1 approval, got %d", len(approvalDB.approvals[rule.ID]))
	}

	// Check rule status (local requires 1 approval, should be approved now)
	updated := ruleDB.rules[rule.ID]
	if updated.Status != domain.RuleStatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", updated.Status)
	}
}

func TestService_ApproveRule_QuorumNotMet(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	// First approval
	err := svc.ApproveRule(context.Background(), rule.ID, "approver-1", "")
	if err != nil {
		t.Fatalf("ApproveRule() error = %v", err)
	}

	// Should still be pending (global requires 2)
	updated := ruleDB.rules[rule.ID]
	if updated.Status != domain.RuleStatusPending {
		t.Errorf("Expected status 'pending', got '%s'", updated.Status)
	}

	// Second approval
	err = svc.ApproveRule(context.Background(), rule.ID, "approver-2", "")
	if err != nil {
		t.Fatalf("ApproveRule() error = %v", err)
	}

	// Now should be approved
	updated = ruleDB.rules[rule.ID]
	if updated.Status != domain.RuleStatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", updated.Status)
	}
}

func TestService_ApproveRule_NoPermission(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	err := svc.ApproveRule(context.Background(), rule.ID, "member-1", "")
	if err != ErrNoApprovalPermission {
		t.Errorf("Expected ErrNoApprovalPermission, got %v", err)
	}
}

func TestService_ApproveRule_AlreadyVoted(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	// First vote
	_ = svc.ApproveRule(context.Background(), rule.ID, "approver-1", "")

	// Try to vote again
	err := svc.ApproveRule(context.Background(), rule.ID, "approver-1", "")
	if err != ErrAlreadyVoted {
		t.Errorf("Expected ErrAlreadyVoted, got %v", err)
	}
}

func TestService_RejectRule(t *testing.T) {
	svc, ruleDB, approvalDB := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerLocal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	err := svc.RejectRule(context.Background(), rule.ID, "approver-1", "Needs improvement")
	if err != nil {
		t.Fatalf("RejectRule() error = %v", err)
	}

	// Check rejection was recorded
	if len(approvalDB.approvals[rule.ID]) != 1 {
		t.Errorf("Expected 1 approval record, got %d", len(approvalDB.approvals[rule.ID]))
	}

	// Check rule is rejected
	updated := ruleDB.rules[rule.ID]
	if updated.Status != domain.RuleStatusRejected {
		t.Errorf("Expected status 'rejected', got '%s'", updated.Status)
	}
}

func TestService_GetApprovalStatus(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule := domain.NewRule("Test Rule", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()
	ruleDB.rules[rule.ID] = rule

	_ = svc.ApproveRule(context.Background(), rule.ID, "approver-1", "")

	status, err := svc.GetApprovalStatus(context.Background(), rule.ID)
	if err != nil {
		t.Fatalf("GetApprovalStatus() error = %v", err)
	}

	if status.RequiredCount != 2 {
		t.Errorf("Expected RequiredCount 2, got %d", status.RequiredCount)
	}
	if status.CurrentCount != 1 {
		t.Errorf("Expected CurrentCount 1, got %d", status.CurrentCount)
	}
	if len(status.Approvals) != 1 {
		t.Errorf("Expected 1 approval, got %d", len(status.Approvals))
	}
}

func TestService_GetPendingRules(t *testing.T) {
	svc, ruleDB, _ := newTestService()

	rule1 := domain.NewRule("Rule 1", domain.TargetLayerLocal, "content", nil, "team-1")
	rule1.Submit()
	ruleDB.rules[rule1.ID] = rule1

	rule2 := domain.NewRule("Rule 2", domain.TargetLayerLocal, "content", nil, "team-1")
	ruleDB.rules[rule2.ID] = rule2 // Still draft

	rules, err := svc.GetPendingRules(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("GetPendingRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 pending rule, got %d", len(rules))
	}
}
