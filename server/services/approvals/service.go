package approvals

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrRuleNotFound         = errors.New("rule not found")
	ErrCannotSubmit         = errors.New("rule cannot be submitted in current state")
	ErrNotPending           = errors.New("rule is not pending approval")
	ErrNoApprovalPermission = errors.New("user does not have permission to approve this rule")
	ErrAlreadyVoted         = errors.New("user has already voted on this rule")
)

type RuleDB interface {
	GetRule(ctx context.Context, id string) (domain.Rule, error)
	UpdateStatus(ctx context.Context, rule domain.Rule) error
	ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error)
	ListPendingByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error)
}

type ApprovalDB interface {
	Create(ctx context.Context, approval domain.RuleApproval) error
	ListByRule(ctx context.Context, ruleID string) ([]domain.RuleApproval, error)
	CountApprovals(ctx context.Context, ruleID string) (int, error)
	HasUserApproved(ctx context.Context, ruleID, userID string) (bool, error)
	DeleteByRule(ctx context.Context, ruleID string) error
}

type ApprovalConfigDB interface {
	GetForScope(ctx context.Context, scope domain.TargetLayer, teamID *string) (domain.ApprovalConfig, error)
}

type RoleDB interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

type AuditLogger interface {
	LogApprovalAction(ctx context.Context, ruleID string, action domain.AuditAction, actorID *string, metadata map[string]interface{}) error
}

type Service struct {
	ruleDB     RuleDB
	approvalDB ApprovalDB
	configDB   ApprovalConfigDB
	roleDB     RoleDB
	auditLog   AuditLogger
}

func NewService(ruleDB RuleDB, approvalDB ApprovalDB, configDB ApprovalConfigDB, roleDB RoleDB) *Service {
	return &Service{
		ruleDB:     ruleDB,
		approvalDB: approvalDB,
		configDB:   configDB,
		roleDB:     roleDB,
	}
}

func (s *Service) WithAuditLogger(logger AuditLogger) *Service {
	s.auditLog = logger
	return s
}

type ApprovalStatus struct {
	RuleID        string                `json:"rule_id"`
	Status        domain.RuleStatus     `json:"status"`
	RequiredCount int                   `json:"required_count"`
	CurrentCount  int                   `json:"current_count"`
	Approvals     []domain.RuleApproval `json:"approvals"`
}

func (s *Service) SubmitRule(ctx context.Context, ruleID string) error {
	rule, err := s.ruleDB.GetRule(ctx, ruleID)
	if err != nil {
		return ErrRuleNotFound
	}

	if !rule.CanSubmit() {
		return ErrCannotSubmit
	}

	rule.Submit()
	if err := s.ruleDB.UpdateStatus(ctx, rule); err != nil {
		return err
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.LogApprovalAction(ctx, ruleID, domain.AuditActionSubmitted, rule.CreatedBy, map[string]interface{}{
			"rule_name":    rule.Name,
			"target_layer": string(rule.TargetLayer),
		})
	}

	return nil
}

func (s *Service) ApproveRule(ctx context.Context, ruleID, userID, comment string) error {
	rule, err := s.ruleDB.GetRule(ctx, ruleID)
	if err != nil {
		return ErrRuleNotFound
	}

	if rule.Status != domain.RuleStatusPending {
		return ErrNotPending
	}

	// Check user has permission
	if err := s.checkApprovalPermission(ctx, userID, rule.TargetLayer, rule.TeamID); err != nil {
		return err
	}

	// Check user hasn't already voted
	hasVoted, err := s.approvalDB.HasUserApproved(ctx, ruleID, userID)
	if err != nil {
		return err
	}
	if hasVoted {
		return ErrAlreadyVoted
	}

	// Record approval
	approval := domain.NewRuleApproval(ruleID, userID, domain.ApprovalDecisionApproved, comment)
	if err := s.approvalDB.Create(ctx, approval); err != nil {
		return err
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.LogApprovalAction(ctx, ruleID, domain.AuditActionApproved, &userID, map[string]interface{}{
			"comment":   comment,
			"rule_name": rule.Name,
		})
	}

	// Check if quorum is met
	return s.checkAndUpdateQuorum(ctx, rule)
}

func (s *Service) RejectRule(ctx context.Context, ruleID, userID, comment string) error {
	rule, err := s.ruleDB.GetRule(ctx, ruleID)
	if err != nil {
		return ErrRuleNotFound
	}

	if rule.Status != domain.RuleStatusPending {
		return ErrNotPending
	}

	// Check user has permission
	if err := s.checkApprovalPermission(ctx, userID, rule.TargetLayer, rule.TeamID); err != nil {
		return err
	}

	// Record rejection
	approval := domain.NewRuleApproval(ruleID, userID, domain.ApprovalDecisionRejected, comment)
	if err := s.approvalDB.Create(ctx, approval); err != nil {
		return err
	}

	// Reject rule immediately (any rejection rejects the rule)
	rule.Reject()
	if err := s.ruleDB.UpdateStatus(ctx, rule); err != nil {
		return err
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.LogApprovalAction(ctx, ruleID, domain.AuditActionRejected, &userID, map[string]interface{}{
			"comment":   comment,
			"rule_name": rule.Name,
		})
	}

	return nil
}

func (s *Service) GetApprovalStatus(ctx context.Context, ruleID string) (ApprovalStatus, error) {
	rule, err := s.ruleDB.GetRule(ctx, ruleID)
	if err != nil {
		return ApprovalStatus{}, ErrRuleNotFound
	}

	config, err := s.configDB.GetForScope(ctx, rule.TargetLayer, rule.TeamID)
	if err != nil {
		return ApprovalStatus{}, err
	}

	approvals, err := s.approvalDB.ListByRule(ctx, ruleID)
	if err != nil {
		return ApprovalStatus{}, err
	}

	currentCount, err := s.approvalDB.CountApprovals(ctx, ruleID)
	if err != nil {
		return ApprovalStatus{}, err
	}

	return ApprovalStatus{
		RuleID:        ruleID,
		Status:        rule.Status,
		RequiredCount: config.RequiredCount,
		CurrentCount:  currentCount,
		Approvals:     approvals,
	}, nil
}

func (s *Service) GetPendingRules(ctx context.Context, teamID string) ([]domain.Rule, error) {
	return s.ruleDB.ListByStatus(ctx, teamID, domain.RuleStatusPending)
}

func (s *Service) GetPendingRulesByScope(ctx context.Context, scope domain.TargetLayer) ([]domain.Rule, error) {
	return s.ruleDB.ListPendingByScope(ctx, scope)
}

func (s *Service) ResetRule(ctx context.Context, ruleID string) error {
	rule, err := s.ruleDB.GetRule(ctx, ruleID)
	if err != nil {
		return ErrRuleNotFound
	}

	// Clear all approvals
	if err := s.approvalDB.DeleteByRule(ctx, ruleID); err != nil {
		return err
	}

	// Reset to draft
	rule.ResetToDraft()
	return s.ruleDB.UpdateStatus(ctx, rule)
}

func (s *Service) checkApprovalPermission(ctx context.Context, userID string, scope domain.TargetLayer, teamID *string) error {
	config, err := s.configDB.GetForScope(ctx, scope, teamID)
	if err != nil {
		return err
	}

	permissions, err := s.roleDB.GetUserPermissions(ctx, userID)
	if err != nil {
		return err
	}

	for _, p := range permissions {
		if p == config.RequiredPermission {
			return nil
		}
	}

	return ErrNoApprovalPermission
}

func (s *Service) checkAndUpdateQuorum(ctx context.Context, rule domain.Rule) error {
	config, err := s.configDB.GetForScope(ctx, rule.TargetLayer, rule.TeamID)
	if err != nil {
		return err
	}

	count, err := s.approvalDB.CountApprovals(ctx, rule.ID)
	if err != nil {
		return err
	}

	if count >= config.RequiredCount {
		rule.Approve()
		return s.ruleDB.UpdateStatus(ctx, rule)
	}

	return nil
}
