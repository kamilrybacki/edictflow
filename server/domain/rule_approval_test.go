package domain

import (
	"testing"
)

func TestRuleApproval_Validate(t *testing.T) {
	tests := []struct {
		name     string
		approval RuleApproval
		wantErr  bool
	}{
		{
			name:     "valid approval",
			approval: RuleApproval{RuleID: "rule-1", UserID: "user-1", Decision: ApprovalDecisionApproved},
			wantErr:  false,
		},
		{
			name:     "valid rejection with comment",
			approval: RuleApproval{RuleID: "rule-1", UserID: "user-1", Decision: ApprovalDecisionRejected, Comment: "Needs work"},
			wantErr:  false,
		},
		{
			name:     "rejection without comment",
			approval: RuleApproval{RuleID: "rule-1", UserID: "user-1", Decision: ApprovalDecisionRejected},
			wantErr:  true,
		},
		{
			name:     "invalid decision",
			approval: RuleApproval{RuleID: "rule-1", UserID: "user-1", Decision: "invalid"},
			wantErr:  true,
		},
		{
			name:     "empty rule ID",
			approval: RuleApproval{RuleID: "", UserID: "user-1", Decision: ApprovalDecisionApproved},
			wantErr:  true,
		},
		{
			name:     "empty user ID",
			approval: RuleApproval{RuleID: "rule-1", UserID: "", Decision: ApprovalDecisionApproved},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.approval.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRuleApproval(t *testing.T) {
	approval := NewRuleApproval("rule-123", "user-456", ApprovalDecisionApproved, "")
	if approval.ID == "" {
		t.Error("Expected approval to have an ID")
	}
	if approval.RuleID != "rule-123" {
		t.Errorf("Expected RuleID 'rule-123', got %s", approval.RuleID)
	}
	if approval.Decision != ApprovalDecisionApproved {
		t.Errorf("Expected decision 'approved', got %s", approval.Decision)
	}
}
