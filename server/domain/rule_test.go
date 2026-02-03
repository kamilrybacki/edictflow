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

func TestRuleStatus_Transitions(t *testing.T) {
	rule := domain.NewRule("Test", domain.TargetLayerGlobal, "content", nil, "team-1")

	// New rules start as draft
	if rule.Status != domain.RuleStatusDraft {
		t.Errorf("Expected new rule to be draft, got %s", rule.Status)
	}

	// Draft can be submitted
	if !rule.CanSubmit() {
		t.Error("Draft rule should be submittable")
	}

	// Submit the rule
	rule.Submit()
	if rule.Status != domain.RuleStatusPending {
		t.Errorf("Expected pending after submit, got %s", rule.Status)
	}
	if rule.SubmittedAt == nil {
		t.Error("SubmittedAt should be set after submit")
	}

	// Pending cannot be submitted again
	if rule.CanSubmit() {
		t.Error("Pending rule should not be submittable")
	}
}

func TestRule_Approve(t *testing.T) {
	rule := domain.NewRule("Test", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()

	rule.Approve()
	if rule.Status != domain.RuleStatusApproved {
		t.Errorf("Expected approved, got %s", rule.Status)
	}
	if rule.ApprovedAt == nil {
		t.Error("ApprovedAt should be set")
	}
}

func TestRule_Reject(t *testing.T) {
	rule := domain.NewRule("Test", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()

	rule.Reject()
	if rule.Status != domain.RuleStatusRejected {
		t.Errorf("Expected rejected, got %s", rule.Status)
	}
}

func TestRule_ResetToDraft(t *testing.T) {
	rule := domain.NewRule("Test", domain.TargetLayerGlobal, "content", nil, "team-1")
	rule.Submit()
	rule.Reject()

	// Rejected rules can be resubmitted
	if !rule.CanSubmit() {
		t.Error("Rejected rule should be submittable")
	}

	rule.ResetToDraft()
	if rule.Status != domain.RuleStatusDraft {
		t.Errorf("Expected draft after reset, got %s", rule.Status)
	}
	if rule.SubmittedAt != nil {
		t.Error("SubmittedAt should be nil after reset")
	}
}
