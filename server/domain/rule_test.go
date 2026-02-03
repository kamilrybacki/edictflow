package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewRuleCreatesValidRule(t *testing.T) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "**/frontend/**"},
	}
	rule := domain.NewRule("React Standards", domain.TargetLayerProject, "# React\nUse hooks.", triggers, "team-123")

	if rule.Name != "React Standards" {
		t.Errorf("expected name 'React Standards', got '%s'", rule.Name)
	}
	if rule.TargetLayer != domain.TargetLayerProject {
		t.Errorf("expected target layer 'project', got '%s'", rule.TargetLayer)
	}
	if rule.PriorityWeight != 0 {
		t.Errorf("expected priority weight 0, got %d", rule.PriorityWeight)
	}
}

func TestTriggerSpecificityOrdersCorrectly(t *testing.T) {
	pathTrigger := domain.Trigger{Type: domain.TriggerTypePath, Pattern: "**/src/**"}
	contextTrigger := domain.Trigger{Type: domain.TriggerTypeContext, ContextTypes: []string{"node"}}
	tagTrigger := domain.Trigger{Type: domain.TriggerTypeTag, Tags: []string{"frontend"}}

	if pathTrigger.Specificity() <= contextTrigger.Specificity() {
		t.Error("path trigger should have higher specificity than context trigger")
	}
	if contextTrigger.Specificity() <= tagTrigger.Specificity() {
		t.Error("context trigger should have higher specificity than tag trigger")
	}
}

func TestRuleValidateRejectsEmptyContent(t *testing.T) {
	rule := domain.Rule{
		ID:          "test-id",
		Name:        "Test Rule",
		TargetLayer: domain.TargetLayerProject,
		Content:     "",
		TeamID:      "team-123",
	}

	err := rule.Validate()
	if err == nil {
		t.Error("expected validation error for empty content")
	}
}
