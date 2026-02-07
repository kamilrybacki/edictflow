package exceptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrExceptionRequestNotFound = errors.New("exception request not found")
	ErrExceptionNotPending      = errors.New("exception request is not pending")
	ErrChangeRequestNotFound    = errors.New("change request not found")
	ErrChangeNotRejected        = errors.New("change request must be rejected to create exception")
)

type Service struct {
	exceptionRepo ExceptionRequestRepository
	changeRepo    ChangeRequestRepository
	notifier      NotificationCreator
	auditLog      AuditLogger
	wsNotifier    WebSocketNotifier
}

func NewService(
	exceptionRepo ExceptionRequestRepository,
	changeRepo ChangeRequestRepository,
) *Service {
	return &Service{
		exceptionRepo: exceptionRepo,
		changeRepo:    changeRepo,
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

type CreateExceptionRequest struct {
	ChangeRequestID       string
	UserID                string
	Justification         string
	ExceptionType         domain.ExceptionType
	RequestedDurationHours *int
}

func (s *Service) Create(ctx context.Context, req CreateExceptionRequest) (*domain.ExceptionRequest, error) {
	// Validate change request exists and is rejected
	cr, err := s.changeRepo.GetByID(ctx, req.ChangeRequestID)
	if err != nil {
		return nil, err
	}
	if cr == nil {
		return nil, ErrChangeRequestNotFound
	}

	if cr.Status != domain.ChangeRequestStatusRejected && cr.Status != domain.ChangeRequestStatusAutoReverted {
		return nil, ErrChangeNotRejected
	}

	er := domain.NewExceptionRequest(
		req.ChangeRequestID,
		req.UserID,
		req.Justification,
		req.ExceptionType,
	)

	if err := s.exceptionRepo.Create(ctx, er); err != nil {
		return nil, err
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.Log(ctx, domain.AuditActionCreated, &req.UserID, "exception_request", er.ID, map[string]interface{}{
			"change_request_id": req.ChangeRequestID,
			"exception_type":    string(req.ExceptionType),
		})
	}

	return &er, nil
}

func (s *Service) Approve(ctx context.Context, id, approverUserID string, expiresAt *time.Time) error {
	er, err := s.exceptionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if er == nil {
		return ErrExceptionRequestNotFound
	}

	if !er.IsPending() {
		return ErrExceptionNotPending
	}

	er.Approve(approverUserID, expiresAt)
	if err := s.exceptionRepo.Update(ctx, *er); err != nil {
		return err
	}

	// Update the change request to exception_granted
	cr, err := s.changeRepo.GetByID(ctx, er.ChangeRequestID)
	if err != nil {
		return err
	}
	if cr != nil {
		cr.GrantException()
		_ = s.changeRepo.Update(ctx, *cr)

		// Notify agent via WebSocket
		if s.wsNotifier != nil {
			_ = s.wsNotifier.BroadcastToAgent(cr.AgentID, "exception_granted", map[string]interface{}{
				"change_id":    cr.ID,
				"exception_id": er.ID,
				"expires_at":   expiresAt,
			})
		}

		// Notify user
		if s.notifier != nil {
			n := domain.NewNotification(
				er.UserID,
				&cr.TeamID,
				domain.NotificationTypeExceptionGranted,
				"Exception granted",
				fmt.Sprintf("Your exception request for %s has been approved", cr.FilePath),
				map[string]interface{}{
					"exception_id":      er.ID,
					"change_request_id": cr.ID,
					"expires_at":        expiresAt,
				},
			)
			_ = s.notifier.Create(ctx, n)
		}
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.Log(ctx, domain.AuditActionApproved, &approverUserID, "exception_request", id, map[string]interface{}{
			"expires_at": expiresAt,
		})
	}

	return nil
}

func (s *Service) Deny(ctx context.Context, id, approverUserID string) error {
	er, err := s.exceptionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if er == nil {
		return ErrExceptionRequestNotFound
	}

	if !er.IsPending() {
		return ErrExceptionNotPending
	}

	er.Deny(approverUserID)
	if err := s.exceptionRepo.Update(ctx, *er); err != nil {
		return err
	}

	// Get change request for notification
	cr, err := s.changeRepo.GetByID(ctx, er.ChangeRequestID)
	if err == nil && cr != nil {
		// Notify agent via WebSocket
		if s.wsNotifier != nil {
			_ = s.wsNotifier.BroadcastToAgent(cr.AgentID, "exception_denied", map[string]interface{}{
				"change_id":    cr.ID,
				"exception_id": er.ID,
			})
		}

		// Notify user
		if s.notifier != nil {
			n := domain.NewNotification(
				er.UserID,
				&cr.TeamID,
				domain.NotificationTypeExceptionDenied,
				"Exception denied",
				fmt.Sprintf("Your exception request for %s has been denied", cr.FilePath),
				map[string]interface{}{
					"exception_id":      er.ID,
					"change_request_id": cr.ID,
				},
			)
			_ = s.notifier.Create(ctx, n)
		}
	}

	// Log audit event
	if s.auditLog != nil {
		_ = s.auditLog.Log(ctx, domain.AuditActionRejected, &approverUserID, "exception_request", id, nil)
	}

	return nil
}

func (s *Service) HasActiveException(ctx context.Context, userID, ruleID, filePath string) (bool, error) {
	er, err := s.exceptionRepo.FindActiveByUserRuleFile(ctx, userID, ruleID, filePath)
	if err != nil {
		return false, err
	}
	return er != nil && er.IsActive(), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.ExceptionRequest, error) {
	return s.exceptionRepo.GetByID(ctx, id)
}

func (s *Service) ListByTeam(ctx context.Context, teamID string, filter ExceptionRequestFilter) ([]domain.ExceptionRequest, error) {
	return s.exceptionRepo.ListByTeam(ctx, teamID, filter)
}
