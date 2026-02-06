package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/deviceauth"
)

type DeviceCodeDB struct {
	pool *pgxpool.Pool
}

func NewDeviceCodeDB(pool *pgxpool.Pool) *DeviceCodeDB {
	return &DeviceCodeDB{pool: pool}
}

func (r *DeviceCodeDB) Create(ctx context.Context, dc domain.DeviceCode) error {
	query := `
		INSERT INTO device_codes (device_code, user_code, client_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		dc.DeviceCode, dc.UserCode, dc.ClientID, dc.ExpiresAt, dc.CreatedAt)
	return err
}

func (r *DeviceCodeDB) GetByDeviceCode(ctx context.Context, deviceCode string) (domain.DeviceCode, error) {
	query := `
		SELECT device_code, user_code, user_id, client_id, expires_at, authorized_at, created_at
		FROM device_codes WHERE device_code = $1
	`
	var dc domain.DeviceCode
	err := r.pool.QueryRow(ctx, query, deviceCode).Scan(
		&dc.DeviceCode, &dc.UserCode, &dc.UserID, &dc.ClientID,
		&dc.ExpiresAt, &dc.AuthorizedAt, &dc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DeviceCode{}, deviceauth.ErrNotFound
	}
	return dc, err
}

func (r *DeviceCodeDB) GetByUserCode(ctx context.Context, userCode string) (domain.DeviceCode, error) {
	query := `
		SELECT device_code, user_code, user_id, client_id, expires_at, authorized_at, created_at
		FROM device_codes WHERE user_code = $1
	`
	var dc domain.DeviceCode
	err := r.pool.QueryRow(ctx, query, userCode).Scan(
		&dc.DeviceCode, &dc.UserCode, &dc.UserID, &dc.ClientID,
		&dc.ExpiresAt, &dc.AuthorizedAt, &dc.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DeviceCode{}, deviceauth.ErrNotFound
	}
	return dc, err
}

func (r *DeviceCodeDB) Authorize(ctx context.Context, deviceCode, userID string) error {
	query := `
		UPDATE device_codes
		SET user_id = $1, authorized_at = $2
		WHERE device_code = $3 AND user_id IS NULL
	`
	result, err := r.pool.Exec(ctx, query, userID, time.Now(), deviceCode)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return deviceauth.ErrNotFound
	}
	return nil
}

func (r *DeviceCodeDB) Delete(ctx context.Context, deviceCode string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM device_codes WHERE device_code = $1", deviceCode)
	return err
}

func (r *DeviceCodeDB) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM device_codes WHERE expires_at < $1", time.Now())
	return err
}
