package domain_test

import (
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
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
	teamID := "team-123"
	rule := domain.Rule{
		ID:          "test-id",
		Name:        "Test Rule",
		TargetLayer: domain.TargetLayerProject,
		Content:     "",
		TeamID:      &teamID,
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

func strPtr(s string) *string {
	return &s
}

func TestRule_IsGlobal(t *testing.T) {
	tests := []struct {
		name   string
		teamID *string
		want   bool
	}{
		{
			name:   "nil team_id is global",
			teamID: nil,
			want:   true,
		},
		{
			name:   "empty team_id is global",
			teamID: strPtr(""),
			want:   true,
		},
		{
			name:   "non-empty team_id is not global",
			teamID: strPtr("team-123"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := domain.Rule{TeamID: tt.teamID}
			if got := rule.IsGlobal(); got != tt.want {
				t.Errorf("Rule.IsGlobal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_ValidateOverrideConflict(t *testing.T) {
	tests := []struct {
		name        string
		rule        domain.Rule
		higherRules []domain.Rule
		wantErr     bool
	}{
		{
			name: "no conflict - higher rule is overridable",
			rule: domain.Rule{
				Name:        "Project security rule",
				TargetLayer: domain.TargetLayerProject,
				CategoryID:  strPtr("cat-1"),
			},
			higherRules: []domain.Rule{
				{
					Name:        "Enterprise security rule",
					TargetLayer: domain.TargetLayerEnterprise,
					CategoryID:  strPtr("cat-1"),
					Overridable: true,
				},
			},
			wantErr: false,
		},
		{
			name: "conflict - higher rule not overridable, same category",
			rule: domain.Rule{
				Name:        "Project security rule",
				TargetLayer: domain.TargetLayerProject,
				CategoryID:  strPtr("cat-1"),
			},
			higherRules: []domain.Rule{
				{
					Name:        "Enterprise security rule",
					TargetLayer: domain.TargetLayerEnterprise,
					CategoryID:  strPtr("cat-1"),
					Overridable: false,
				},
			},
			wantErr: true,
		},
		{
			name: "no conflict - different categories",
			rule: domain.Rule{
				Name:        "Project testing rule",
				TargetLayer: domain.TargetLayerProject,
				CategoryID:  strPtr("cat-2"),
			},
			higherRules: []domain.Rule{
				{
					Name:        "Enterprise security rule",
					TargetLayer: domain.TargetLayerEnterprise,
					CategoryID:  strPtr("cat-1"),
					Overridable: false,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.ValidateOverrideConflict(tt.higherRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.ValidateOverrideConflict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGlobalRule(t *testing.T) {
	rule := domain.NewGlobalRule("Security Policy", "Never hardcode secrets", true)

	if !rule.IsGlobal() {
		t.Error("expected global rule to have nil TeamID")
	}
	if rule.TargetLayer != domain.TargetLayerOrganization {
		t.Errorf("expected organization layer, got %s", rule.TargetLayer)
	}
	if !rule.Force {
		t.Error("expected force to be true")
	}
}

func TestNewRule_WithTeamID(t *testing.T) {
	triggers := []domain.Trigger{}
	rule := domain.NewRule("Team Rule", domain.TargetLayerProject, "content", triggers, "team-123")

	if rule.IsGlobal() {
		t.Error("expected team rule to have TeamID set")
	}
	if rule.TeamID == nil || *rule.TeamID != "team-123" {
		t.Errorf("expected TeamID 'team-123', got %v", rule.TeamID)
	}
}

func TestRule_IsEffective(t *testing.T) {
	import_time := func() time.Time { return time.Now() }
	now := import_time()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name string
		rule domain.Rule
		want bool
	}{
		{
			name: "no dates - always effective",
			rule: domain.Rule{Name: "test"},
			want: true,
		},
		{
			name: "start in past, no end - effective",
			rule: domain.Rule{
				Name:           "test",
				EffectiveStart: &past,
			},
			want: true,
		},
		{
			name: "start in future - not effective",
			rule: domain.Rule{
				Name:           "test",
				EffectiveStart: &future,
			},
			want: false,
		},
		{
			name: "end in past - not effective",
			rule: domain.Rule{
				Name:         "test",
				EffectiveEnd: &past,
			},
			want: false,
		},
		{
			name: "start in past, end in future - effective",
			rule: domain.Rule{
				Name:           "test",
				EffectiveStart: &past,
				EffectiveEnd:   &future,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.IsEffective(); got != tt.want {
				t.Errorf("Rule.IsEffective() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_ValidateForce(t *testing.T) {
	tests := []struct {
		name    string
		rule    domain.Rule
		wantErr bool
	}{
		{
			name: "global rule with force=true, enterprise layer - valid",
			rule: domain.Rule{
				Name:        "Global Policy",
				Content:     "some content",
				TargetLayer: domain.TargetLayerEnterprise,
				TeamID:      nil,
				Force:       true,
			},
			wantErr: false,
		},
		{
			name: "global rule with force=false, enterprise layer - valid",
			rule: domain.Rule{
				Name:        "Global Policy",
				Content:     "some content",
				TargetLayer: domain.TargetLayerEnterprise,
				TeamID:      nil,
				Force:       false,
			},
			wantErr: false,
		},
		{
			name: "team rule with force=true - invalid",
			rule: domain.Rule{
				Name:        "Team Policy",
				Content:     "some content",
				TargetLayer: domain.TargetLayerProject,
				TeamID:      strPtr("team-123"),
				Force:       true,
			},
			wantErr: true,
		},
		{
			name: "global rule with project layer - invalid",
			rule: domain.Rule{
				Name:        "Global Policy",
				Content:     "some content",
				TargetLayer: domain.TargetLayerProject,
				TeamID:      nil,
				Force:       false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
