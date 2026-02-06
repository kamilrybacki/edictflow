package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type NotificationChannelDB struct {
	pool *pgxpool.Pool
}

func NewNotificationChannelDB(pool *pgxpool.Pool) *NotificationChannelDB {
	return &NotificationChannelDB{pool: pool}
}

func (db *NotificationChannelDB) Create(ctx context.Context, nc domain.NotificationChannel) error {
	configJSON, err := json.Marshal(nc.Config)
	if err != nil {
		return err
	}
	_, err = db.pool.Exec(ctx, `
		INSERT INTO notification_channels (id, team_id, channel_type, config, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, nc.ID, nc.TeamID, nc.ChannelType, configJSON, nc.Enabled, nc.CreatedAt)
	return err
}

func (db *NotificationChannelDB) GetByID(ctx context.Context, id string) (*domain.NotificationChannel, error) {
	var nc domain.NotificationChannel
	var configJSON []byte
	err := db.pool.QueryRow(ctx, `
		SELECT id, team_id, channel_type, config, enabled, created_at
		FROM notification_channels WHERE id = $1
	`, id).Scan(&nc.ID, &nc.TeamID, &nc.ChannelType, &configJSON, &nc.Enabled, &nc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &nc.Config); err != nil {
		return nil, err
	}
	return &nc, nil
}

func (db *NotificationChannelDB) ListByTeam(ctx context.Context, teamID string) ([]domain.NotificationChannel, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, team_id, channel_type, config, enabled, created_at
		FROM notification_channels WHERE team_id = $1
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NotificationChannel
	for rows.Next() {
		var nc domain.NotificationChannel
		var configJSON []byte
		if err := rows.Scan(&nc.ID, &nc.TeamID, &nc.ChannelType, &configJSON, &nc.Enabled, &nc.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &nc.Config); err != nil {
			return nil, err
		}
		results = append(results, nc)
	}
	return results, rows.Err()
}

func (db *NotificationChannelDB) ListEnabledByTeam(ctx context.Context, teamID string) ([]domain.NotificationChannel, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, team_id, channel_type, config, enabled, created_at
		FROM notification_channels WHERE team_id = $1 AND enabled = true
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NotificationChannel
	for rows.Next() {
		var nc domain.NotificationChannel
		var configJSON []byte
		if err := rows.Scan(&nc.ID, &nc.TeamID, &nc.ChannelType, &configJSON, &nc.Enabled, &nc.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &nc.Config); err != nil {
			return nil, err
		}
		results = append(results, nc)
	}
	return results, rows.Err()
}

func (db *NotificationChannelDB) Update(ctx context.Context, nc domain.NotificationChannel) error {
	configJSON, err := json.Marshal(nc.Config)
	if err != nil {
		return err
	}
	_, err = db.pool.Exec(ctx, `
		UPDATE notification_channels SET
			channel_type = $2, config = $3, enabled = $4
		WHERE id = $1
	`, nc.ID, nc.ChannelType, configJSON, nc.Enabled)
	return err
}

func (db *NotificationChannelDB) Delete(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM notification_channels WHERE id = $1`, id)
	return err
}

func (db *NotificationChannelDB) CountByTeam(ctx context.Context, teamID string) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notification_channels WHERE team_id = $1
	`, teamID).Scan(&count)
	return count, err
}
