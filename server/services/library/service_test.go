package library_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/library"
)

type mockRuleDB struct {
	rules map[string]domain.Rule
}

func newMockRuleDB() *mockRuleDB {
	return &mockRuleDB{rules: make(map[string]domain.Rule)}
}

func (m *mockRuleDB) CreateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	if r, ok := m.rules[id]; ok {
		return r, nil
	}
	return domain.Rule{}, library.ErrRuleNotFound
}

func (m *mockRuleDB) ListAllRules(ctx context.Context) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockRuleDB) UpdateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockRuleDB) DeleteRule(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func (m *mockRuleDB) UpdateStatus(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

type mockAttachmentService struct {
	called bool
}

func (m *mockAttachmentService) AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error {
	m.called = true
	return nil
}

func TestLibraryService_Create(t *testing.T) {
	db := newMockRuleDB()
	svc := library.NewService(db, nil)

	rule, err := svc.Create(context.Background(), library.CreateRequest{
		Name:        "Test Rule",
		Content:     "Test content",
		TargetLayer: domain.TargetLayerOrganization,
		CreatedBy:   "user-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Name != "Test Rule" {
		t.Errorf("expected name 'Test Rule', got '%s'", rule.Name)
	}
	if rule.Status != domain.RuleStatusDraft {
		t.Errorf("expected draft status, got '%s'", rule.Status)
	}
}

func TestLibraryService_ApproveEnterpriseRule(t *testing.T) {
	db := newMockRuleDB()
	attSvc := &mockAttachmentService{}
	svc := library.NewService(db, attSvc)

	// Create and submit enterprise rule
	rule, _ := svc.Create(context.Background(), library.CreateRequest{
		Name:        "Enterprise Policy",
		Content:     "All teams must...",
		TargetLayer: domain.TargetLayerOrganization,
		CreatedBy:   "user-1",
	})
	svc.Submit(context.Background(), rule.ID)

	// Approve
	approved, err := svc.Approve(context.Background(), rule.ID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.RuleStatusApproved {
		t.Errorf("expected approved status, got '%s'", approved.Status)
	}
	if !attSvc.called {
		t.Error("expected AutoAttachEnterpriseRule to be called")
	}
}
