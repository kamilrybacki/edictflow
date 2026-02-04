package deviceauth

import (
	"context"

	"github.com/kamilrybacki/claudeception/server/domain"
)

type Repository interface {
	Create(ctx context.Context, dc domain.DeviceCode) error
	GetByDeviceCode(ctx context.Context, deviceCode string) (domain.DeviceCode, error)
	GetByUserCode(ctx context.Context, userCode string) (domain.DeviceCode, error)
	Authorize(ctx context.Context, deviceCode, userID string) error
	Delete(ctx context.Context, deviceCode string) error
}
