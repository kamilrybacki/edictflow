package attachments

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrNotFound = errors.New("attachment not found")
var ErrRuleNotApproved = errors.New("rule must be approved before attaching")
var ErrAlreadyAttached = errors.New("rule is already attached to this team")

type DB interface {
	Create(ctx context.Context, attachment domain.RuleAttachment) error
	GetByID(ctx context.Context, id string) (domain.RuleAttachment, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error)
	Update(ctx context.Context, attachment domain.RuleAttachment) error
	Delete(ctx context.Context, id string) error
}

type RuleDB interface {
	GetRule(ctx context.Context, id string) (domain.Rule, error)
}

type TeamDB interface {
	ListTeams(ctx context.Context) ([]domain.Team, error)
}

type Service struct {
	db     DB
	ruleDB RuleDB
	teamDB TeamDB
}

func NewService(db DB, ruleDB RuleDB, teamDB TeamDB) *Service {
	return &Service{db: db, ruleDB: ruleDB, teamDB: teamDB}
}

type AttachRequest struct {
	RuleID          string
	TeamID          string
	EnforcementMode domain.EnforcementMode
	TimeoutHours    int
	RequestedBy     string
}

func (s *Service) RequestAttachment(ctx context.Context, req AttachRequest) (domain.RuleAttachment, error) {
	// Verify rule exists and is approved
	rule, err := s.ruleDB.GetRule(ctx, req.RuleID)
	if err != nil {
		return domain.RuleAttachment{}, err
	}
	if rule.Status != domain.RuleStatusApproved {
		return domain.RuleAttachment{}, ErrRuleNotApproved
	}

	attachment := domain.NewRuleAttachment(req.RuleID, req.TeamID, req.EnforcementMode, req.RequestedBy)
	if req.TimeoutHours > 0 {
		attachment.TemporaryTimeoutHours = req.TimeoutHours
	}

	if err := s.db.Create(ctx, attachment); err != nil {
		return domain.RuleAttachment{}, err
	}
	return attachment, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	return s.db.GetByID(ctx, id)
}

func (s *Service) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	return s.db.ListByTeam(ctx, teamID)
}

func (s *Service) ApproveAttachment(ctx context.Context, id, approvedBy string) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.Approve(approvedBy)
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) RejectAttachment(ctx context.Context, id string) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.Reject()
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) UpdateEnforcement(ctx context.Context, id string, mode domain.EnforcementMode, timeoutHours int) (domain.RuleAttachment, error) {
	att, err := s.db.GetByID(ctx, id)
	if err != nil {
		return domain.RuleAttachment{}, err
	}

	att.UpdateEnforcement(mode, timeoutHours)
	if err := s.db.Update(ctx, att); err != nil {
		return domain.RuleAttachment{}, err
	}
	return att, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.db.Delete(ctx, id)
}

// AutoAttachEnterpriseRule creates approved attachments for all teams
func (s *Service) AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error {
	teams, err := s.teamDB.ListTeams(ctx)
	if err != nil {
		return err
	}

	for _, team := range teams {
		att := domain.NewApprovedAttachment(ruleID, team.ID, domain.EnforcementModeBlock, approvedBy)
		if err := s.db.Create(ctx, att); err != nil {
			// Ignore duplicate errors (rule might already be attached)
			continue
		}
	}
	return nil
}
