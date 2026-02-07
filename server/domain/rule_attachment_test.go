package domain_test

import (
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
)

func TestNewRuleAttachment(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-123", "team-456", domain.EnforcementModeBlock, "user-789")

	if attachment.RuleID != "rule-123" {
		t.Errorf("expected RuleID 'rule-123', got '%s'", attachment.RuleID)
	}
	if attachment.TeamID != "team-456" {
		t.Errorf("expected TeamID 'team-456', got '%s'", attachment.TeamID)
	}
	if attachment.EnforcementMode != domain.EnforcementModeBlock {
		t.Errorf("expected EnforcementMode 'block', got '%s'", attachment.EnforcementMode)
	}
	if attachment.Status != domain.AttachmentStatusPending {
		t.Errorf("expected Status 'pending', got '%s'", attachment.Status)
	}
	if attachment.RequestedBy != "user-789" {
		t.Errorf("expected RequestedBy 'user-789', got '%s'", attachment.RequestedBy)
	}
}

func TestRuleAttachmentValidate(t *testing.T) {
	tests := []struct {
		name       string
		attachment domain.RuleAttachment
		wantErr    bool
	}{
		{
			name: "valid attachment",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				TeamID:          "team-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: false,
		},
		{
			name: "missing rule ID",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				TeamID:          "team-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
		{
			name: "missing team ID",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				EnforcementMode: domain.EnforcementModeBlock,
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
		{
			name: "invalid enforcement mode",
			attachment: domain.RuleAttachment{
				ID:              "att-1",
				RuleID:          "rule-1",
				TeamID:          "team-1",
				EnforcementMode: "invalid",
				RequestedBy:     "user-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attachment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RuleAttachment.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleAttachmentApprove(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-1", "team-1", domain.EnforcementModeBlock, "user-1")

	attachment.Approve("admin-1")

	if attachment.Status != domain.AttachmentStatusApproved {
		t.Errorf("expected status 'approved', got '%s'", attachment.Status)
	}
	if attachment.ApprovedBy == nil || *attachment.ApprovedBy != "admin-1" {
		t.Error("expected ApprovedBy to be set")
	}
	if attachment.ApprovedAt == nil {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestRuleAttachmentReject(t *testing.T) {
	attachment := domain.NewRuleAttachment("rule-1", "team-1", domain.EnforcementModeBlock, "user-1")

	attachment.Reject()

	if attachment.Status != domain.AttachmentStatusRejected {
		t.Errorf("expected status 'rejected', got '%s'", attachment.Status)
	}
}
