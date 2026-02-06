package changes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrChangeRequestNotFound = errors.New("change request not found")
	ErrChangeNotPending      = errors.New("change request is not pending")
	ErrRuleNotFound          = errors.New("rule not found")
	ErrAgentNotFound         = errors.New("agent not found")
)

type Service struct {
	changeRepo   ChangeRequestRepository
	ruleRepo     RuleRepository
	agentRepo    AgentRepository
	notifier     NotificationCreator
	auditLog     AuditLogger
	wsNotifier   WebSocketNotifier
}

type WebSocketNotifier interface {
	BroadcastToAgent(agentID string, msgType string, payload interface{}) error
}

func NewService(
	changeRepo ChangeRequestRepository,
	ruleRepo RuleRepository,
	agentRepo AgentRepository,
) *Service {
	return &Service{
		changeRepo: changeRepo,
		ruleRepo:   ruleRepo,
		agentRepo:  agentRepo,
	}
}

func (s *Service) WithNotifier(notifier NotificationCreator) *Service {
	s.notifier = notifier
	return s
}

func (s *Service) WithAuditLogger(logger AuditLogger) *Service {
	s.auditLog = logger
	return s
}

func (s *Service) WithWebSocketNotifier(wsNotifier WebSocketNotifier) *Service {
	s.wsNotifier = wsNotifier
	return s
}

type AgentChangePayload struct {
	RuleID          string
	AgentID         string
	UserID          string
	TeamID          string
	FilePath        string
	OriginalHash    string
	ModifiedHash    string
	Diff            string
	EnforcementMode domain.EnforcementMode
}

func (s *Service) CreateFromAgent(ctx context.Context, payload AgentChangePayload) (*domain.ChangeRequest, error) {
	// Validate rule exists
	rule, err := s.ruleRepo.GetRule(ctx, payload.RuleID)
	if err != nil {
		return nil, ErrRuleNotFound
	}

	// Validate agent exists
	agent, err := s.agentRepo.GetByID(ctx, payload.AgentID)
	if err != nil || agent == nil {
		return nil, ErrAgentNotFound
	}

	// Check for existing pending request for same file
	existing, err := s.changeRepo.FindByAgentAndFile(ctx, payload.AgentID, payload.FilePath)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Update existing request instead of creating new one
		existing.UpdateDiff(payload.ModifiedHash, payload.Diff)
		if err := s.changeRepo.Update(ctx, *existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Calculate timeout for temporary mode
	var timeoutAt *time.Time
	if payload.EnforcementMode == domain.EnforcementModeTemporary {
		t := time.Now().Add(time.Duration(rule.TemporaryTimeoutHours) * time.Hour)
		timeoutAt = &t
	}

	cr := domain.NewChangeRequest(
		payload.RuleID,
		payload.AgentID,
		payload.UserID,
		payload.TeamID,
		payload.FilePath,
		payload.OriginalHash,
		payload.ModifiedHash,
		payload.Diff,
		payload.EnforcementMode,
		timeoutAt,
	)

	if err := s.changeRepo.Create(ctx, cr); err != nil {
		return nil, err
	}

	// Log audit event
	if s.auditLog != nil {
		s.auditLog.Log(ctx, domain.AuditActionCreated, &payload.UserID, "change_request", cr.ID, map[string]interface{}{
			"rule_id":          payload.RuleID,
			"file_path":        payload.FilePath,
			"enforcement_mode": string(payload.EnforcementMode),
		})
	}

	// Create notification for admins
	if s.notifier != nil {
		n := domain.NewNotification(
			payload.UserID,
			&payload.TeamID,
			domain.NotificationTypeApprovalRequired,
			"New change request pending approval",
			fmt.Sprintf("A change to %s requires approval", payload.FilePath),
			map[string]interface{}{
				"change_request_id": cr.ID,
				"file_path":         payload.FilePath,
			},
		)
		s.notifier.Create(ctx, n)
	}

	return &cr, nil
}

func (s *Service) Approve(ctx context.Context, id, approverUserID string) error {
	cr, err := s.changeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cr == nil {
		return ErrChangeRequestNotFound
	}

	if !cr.IsPending() {
		return ErrChangeNotPending
	}

	cr.Approve(approverUserID)
	if err := s.changeRepo.Update(ctx, *cr); err != nil {
		return err
	}

	// Log audit event
	if s.auditLog != nil {
		s.auditLog.Log(ctx, domain.AuditActionApproved, &approverUserID, "change_request", id, map[string]interface{}{
			"file_path": cr.FilePath,
		})
	}

	// Notify agent via WebSocket
	if s.wsNotifier != nil {
		s.wsNotifier.BroadcastToAgent(cr.AgentID, "change_approved", map[string]interface{}{
			"change_id": cr.ID,
			"rule_id":   cr.RuleID,
		})
	}

	// Notify user
	if s.notifier != nil {
		n := domain.NewNotification(
			cr.UserID,
			&cr.TeamID,
			domain.NotificationTypeChangeApproved,
			"Change approved",
			fmt.Sprintf("Your change to %s has been approved", cr.FilePath),
			map[string]interface{}{
				"change_request_id": cr.ID,
			},
		)
		s.notifier.Create(ctx, n)
	}

	return nil
}

func (s *Service) Reject(ctx context.Context, id, approverUserID string) error {
	cr, err := s.changeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cr == nil {
		return ErrChangeRequestNotFound
	}

	if !cr.IsPending() {
		return ErrChangeNotPending
	}

	cr.Reject(approverUserID)
	if err := s.changeRepo.Update(ctx, *cr); err != nil {
		return err
	}

	// Log audit event
	if s.auditLog != nil {
		s.auditLog.Log(ctx, domain.AuditActionRejected, &approverUserID, "change_request", id, map[string]interface{}{
			"file_path": cr.FilePath,
		})
	}

	// Notify agent via WebSocket to revert
	if s.wsNotifier != nil {
		s.wsNotifier.BroadcastToAgent(cr.AgentID, "change_rejected", map[string]interface{}{
			"change_id":      cr.ID,
			"rule_id":        cr.RuleID,
			"revert_to_hash": cr.OriginalHash,
		})
	}

	// Notify user
	if s.notifier != nil {
		n := domain.NewNotification(
			cr.UserID,
			&cr.TeamID,
			domain.NotificationTypeChangeRejected,
			"Change rejected",
			fmt.Sprintf("Your change to %s has been rejected", cr.FilePath),
			map[string]interface{}{
				"change_request_id": cr.ID,
			},
		)
		s.notifier.Create(ctx, n)
	}

	return nil
}

func (s *Service) HandleExpiredTemporary(ctx context.Context) ([]domain.ChangeRequest, error) {
	expired, err := s.changeRepo.FindExpiredTemporary(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	var processed []domain.ChangeRequest
	for _, cr := range expired {
		cr.AutoRevert()
		if err := s.changeRepo.Update(ctx, cr); err != nil {
			continue
		}
		processed = append(processed, cr)

		// Notify agent via WebSocket to revert
		if s.wsNotifier != nil {
			s.wsNotifier.BroadcastToAgent(cr.AgentID, "change_rejected", map[string]interface{}{
				"change_id":      cr.ID,
				"rule_id":        cr.RuleID,
				"revert_to_hash": cr.OriginalHash,
			})
		}

		// Notify user
		if s.notifier != nil {
			n := domain.NewNotification(
				cr.UserID,
				&cr.TeamID,
				domain.NotificationTypeChangeAutoReverted,
				"Change auto-reverted",
				fmt.Sprintf("Your change to %s was reverted (approval timeout)", cr.FilePath),
				map[string]interface{}{
					"change_request_id": cr.ID,
				},
			)
			s.notifier.Create(ctx, n)
		}
	}

	return processed, nil
}

func (s *Service) UpdateFromAgent(ctx context.Context, id, newHash, newDiff string) error {
	cr, err := s.changeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cr == nil {
		return ErrChangeRequestNotFound
	}

	if !cr.IsPending() {
		return ErrChangeNotPending
	}

	cr.UpdateDiff(newHash, newDiff)
	return s.changeRepo.Update(ctx, *cr)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.ChangeRequest, error) {
	return s.changeRepo.GetByID(ctx, id)
}

func (s *Service) ListByTeam(ctx context.Context, teamID string, filter ChangeRequestFilter) ([]domain.ChangeRequest, error) {
	return s.changeRepo.ListByTeam(ctx, teamID, filter)
}

func (s *Service) GrantException(ctx context.Context, id string) error {
	cr, err := s.changeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cr == nil {
		return ErrChangeRequestNotFound
	}

	cr.GrantException()
	return s.changeRepo.Update(ctx, *cr)
}
