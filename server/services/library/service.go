package library

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrRuleNotFound = errors.New("rule not found")
var ErrInvalidStatus = errors.New("rule is not in a valid status for this operation")

type DB interface {
	CreateRule(ctx context.Context, rule domain.Rule) error
	GetRule(ctx context.Context, id string) (domain.Rule, error)
	ListAllRules(ctx context.Context) ([]domain.Rule, error)
	UpdateRule(ctx context.Context, rule domain.Rule) error
	DeleteRule(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, rule domain.Rule) error
}

type AttachmentService interface {
	AutoAttachEnterpriseRule(ctx context.Context, ruleID, approvedBy string) error
}

type Service struct {
	db            DB
	attachmentSvc AttachmentService
}

func NewService(db DB, attachmentSvc AttachmentService) *Service {
	return &Service{db: db, attachmentSvc: attachmentSvc}
}

type CreateRequest struct {
	Name           string
	Content        string
	Description    string
	TargetLayer    domain.TargetLayer
	CategoryID     string
	PriorityWeight int
	Overridable    bool
	Tags           []string
	Triggers       []domain.Trigger
	CreatedBy      string
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.Rule, error) {
	rule := domain.NewLibraryRule(req.Name, req.TargetLayer, req.Content, req.Triggers, req.CreatedBy)

	if req.Description != "" {
		rule.Description = &req.Description
	}
	if req.CategoryID != "" {
		rule.CategoryID = &req.CategoryID
	}
	rule.PriorityWeight = req.PriorityWeight
	rule.Overridable = req.Overridable
	rule.Tags = req.Tags

	if err := rule.Validate(); err != nil {
		return domain.Rule{}, err
	}
	if err := s.db.CreateRule(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	return s.db.GetRule(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Rule, error) {
	return s.db.ListAllRules(ctx)
}

func (s *Service) Update(ctx context.Context, rule domain.Rule) error {
	existing, err := s.db.GetRule(ctx, rule.ID)
	if err != nil {
		return err
	}
	// Only draft or rejected rules can be edited
	if existing.Status != domain.RuleStatusDraft && existing.Status != domain.RuleStatusRejected {
		return ErrInvalidStatus
	}
	return s.db.UpdateRule(ctx, rule)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return err
	}
	// Only draft rules can be deleted
	if rule.Status != domain.RuleStatusDraft {
		return ErrInvalidStatus
	}
	return s.db.DeleteRule(ctx, id)
}

func (s *Service) Submit(ctx context.Context, id string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if !rule.CanSubmit() {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Submit()
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}

func (s *Service) Approve(ctx context.Context, id, approvedBy string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if rule.Status != domain.RuleStatusPending {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Approve()
	rule.ApprovedBy = &approvedBy
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}

	// Auto-attach enterprise rules to all teams
	if rule.IsEnterprise() && s.attachmentSvc != nil {
		if err := s.attachmentSvc.AutoAttachEnterpriseRule(ctx, rule.ID, approvedBy); err != nil {
			// Log but don't fail - rule is still approved
			// TODO: Add proper logging
		}
	}

	return rule, nil
}

func (s *Service) Reject(ctx context.Context, id string) (domain.Rule, error) {
	rule, err := s.db.GetRule(ctx, id)
	if err != nil {
		return domain.Rule{}, err
	}
	if rule.Status != domain.RuleStatusPending {
		return domain.Rule{}, ErrInvalidStatus
	}

	rule.Reject()
	if err := s.db.UpdateStatus(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
}
