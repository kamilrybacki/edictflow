package attachments_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/attachments"
)

type mockAttachmentDB struct {
	attachments map[string]domain.RuleAttachment
}

func newMockDB() *mockAttachmentDB {
	return &mockAttachmentDB{attachments: make(map[string]domain.RuleAttachment)}
}

func (m *mockAttachmentDB) Create(ctx context.Context, att domain.RuleAttachment) error {
	m.attachments[att.ID] = att
	return nil
}

func (m *mockAttachmentDB) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	if att, ok := m.attachments[id]; ok {
		return att, nil
	}
	return domain.RuleAttachment{}, attachments.ErrNotFound
}

func (m *mockAttachmentDB) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	var result []domain.RuleAttachment
	for _, att := range m.attachments {
		if att.TeamID == teamID {
			result = append(result, att)
		}
	}
	return result, nil
}

func (m *mockAttachmentDB) Update(ctx context.Context, att domain.RuleAttachment) error {
	m.attachments[att.ID] = att
	return nil
}

func (m *mockAttachmentDB) Delete(ctx context.Context, id string) error {
	delete(m.attachments, id)
	return nil
}

type mockRuleDB struct{}

func (m *mockRuleDB) GetRule(ctx context.Context, id string) (domain.Rule, error) {
	return domain.Rule{
		ID:          id,
		Status:      domain.RuleStatusApproved,
		TargetLayer: domain.TargetLayerTeam,
	}, nil
}

type mockTeamDB struct{}

func (m *mockTeamDB) ListAllTeams(ctx context.Context) ([]domain.Team, error) {
	return []domain.Team{
		{ID: "team-1"},
		{ID: "team-2"},
	}, nil
}

func TestService_RequestAttachment(t *testing.T) {
	db := newMockDB()
	svc := attachments.NewService(db, &mockRuleDB{}, &mockTeamDB{})

	att, err := svc.RequestAttachment(context.Background(), attachments.AttachRequest{
		RuleID:          "rule-1",
		TeamID:          "team-1",
		EnforcementMode: domain.EnforcementModeBlock,
		RequestedBy:     "user-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if att.Status != domain.AttachmentStatusPending {
		t.Errorf("expected pending status, got %s", att.Status)
	}
}

func TestService_ApproveAttachment(t *testing.T) {
	db := newMockDB()
	svc := attachments.NewService(db, &mockRuleDB{}, &mockTeamDB{})

	att, _ := svc.RequestAttachment(context.Background(), attachments.AttachRequest{
		RuleID:          "rule-1",
		TeamID:          "team-1",
		EnforcementMode: domain.EnforcementModeBlock,
		RequestedBy:     "user-1",
	})

	approved, err := svc.ApproveAttachment(context.Background(), att.ID, "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if approved.Status != domain.AttachmentStatusApproved {
		t.Errorf("expected approved status, got %s", approved.Status)
	}
}
