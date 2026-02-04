package worker

import (
	"context"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type ChangeRequestExpirer interface {
	HandleExpiredTemporary(ctx context.Context) ([]domain.ChangeRequest, error)
}

type TimeoutChecker struct {
	changeService ChangeRequestExpirer
}

func NewTimeoutChecker(changeService ChangeRequestExpirer) *TimeoutChecker {
	return &TimeoutChecker{
		changeService: changeService,
	}
}

func (tc *TimeoutChecker) HandleExpiredTemporary(ctx context.Context) error {
	_, err := tc.changeService.HandleExpiredTemporary(ctx)
	return err
}
