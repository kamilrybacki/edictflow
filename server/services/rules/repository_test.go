package rules_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/rules"
)

// teamIDMatches checks if a *string TeamID matches a string value
func teamIDMatches(teamIDPtr *string, teamID string) bool {
	if teamIDPtr == nil {
		return teamID == ""
	}
	return *teamIDPtr == teamID
}

type mockDB struct {
	rules map[string]domain.Rule
}

func newMockDB() *mockDB {
	return &mockDB{rules: make(map[string]domain.Rule)}
}

func (m *mockDB) CreateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	rule, ok := m.rules[id]
	if !ok {
		return domain.Rule{}, rules.ErrRuleNotFound
	}
	return rule, nil
}

func (m *mockDB) ListRulesByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		if teamIDMatches(r.TeamID, teamID) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockDB) ListAllRules(ctx context.Context) ([]domain.Rule, error) {
	var result []domain.Rule
	for _, r := range m.rules {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockDB) UpdateRule(ctx context.Context, rule domain.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

func (m *mockDB) DeleteRule(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func TestRuleRepositoryCreateAndGet(t *testing.T) {
	db := newMockDB()
	repo := rules.NewRepository(db)
	ctx := context.Background()

	triggers := []domain.Trigger{{Type: domain.TriggerTypePath, Pattern: "**/src/**"}}
	rule := domain.NewRule("Test Rule", domain.TargetLayerProject, "# Content", triggers, "team-1")

	err := repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	got, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if got.Name != "Test Rule" {
		t.Errorf("expected name 'Test Rule', got '%s'", got.Name)
	}
}

func TestRuleRepositoryListByTeam(t *testing.T) {
	db := newMockDB()
	repo := rules.NewRepository(db)
	ctx := context.Background()

	rule1 := domain.NewRule("Rule 1", domain.TargetLayerProject, "# 1", nil, "team-1")
	rule2 := domain.NewRule("Rule 2", domain.TargetLayerProject, "# 2", nil, "team-1")
	rule3 := domain.NewRule("Rule 3", domain.TargetLayerProject, "# 3", nil, "team-2")

	_ = repo.Create(ctx, rule1)
	_ = repo.Create(ctx, rule2)
	_ = repo.Create(ctx, rule3)

	rules, err := repo.ListByTeam(ctx, "team-1")
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
	}

	if len(rules) != 2 {
		t.Errorf("expected 2 rules for team-1, got %d", len(rules))
	}
}
