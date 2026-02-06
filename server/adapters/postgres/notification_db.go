package postgres

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/notifications"
)

type NotificationDB struct {
	pool *pgxpool.Pool
}

func NewNotificationDB(pool *pgxpool.Pool) *NotificationDB {
	return &NotificationDB{pool: pool}
}

func (db *NotificationDB) Create(ctx context.Context, n domain.Notification) error {
	metadataJSON, err := json.Marshal(n.Metadata)
	if err != nil {
		return err
	}
	_, err = db.pool.Exec(ctx, `
		INSERT INTO notifications (id, user_id, team_id, type, title, body, metadata, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, n.ID, n.UserID, n.TeamID, n.Type, n.Title, n.Body, metadataJSON, n.ReadAt, n.CreatedAt)
	return err
}

func (db *NotificationDB) CreateBulk(ctx context.Context, notifications []domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, n := range notifications {
		metadataJSON, err := json.Marshal(n.Metadata)
		if err != nil {
			return err
		}
		batch.Queue(`
			INSERT INTO notifications (id, user_id, team_id, type, title, body, metadata, read_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, n.ID, n.UserID, n.TeamID, n.Type, n.Title, n.Body, metadataJSON, n.ReadAt, n.CreatedAt)
	}

	results := db.pool.SendBatch(ctx, batch)
	defer results.Close()

	for range notifications {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (db *NotificationDB) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	var n domain.Notification
	var metadataJSON []byte
	err := db.pool.QueryRow(ctx, `
		SELECT id, user_id, team_id, type, title, body, metadata, read_at, created_at
		FROM notifications WHERE id = $1
	`, id).Scan(&n.ID, &n.UserID, &n.TeamID, &n.Type, &n.Title, &n.Body, &metadataJSON, &n.ReadAt, &n.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
		return nil, err
	}
	return &n, nil
}

func (db *NotificationDB) ListByUser(ctx context.Context, userID string, filter notifications.NotificationFilter) ([]domain.Notification, error) {
	query := `
		SELECT id, user_id, team_id, type, title, body, metadata, read_at, created_at
		FROM notifications WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIdx := 2

	if filter.Type != nil {
		query += ` AND type = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.Type)
		argIdx++
	}
	if filter.Unread != nil && *filter.Unread {
		query += ` AND read_at IS NULL`
	}
	if filter.TeamID != nil {
		query += ` AND team_id = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.TeamID)
		argIdx++
	}

	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT $` + strconv.Itoa(argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += ` OFFSET $` + strconv.Itoa(argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var metadataJSON []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.TeamID, &n.Type, &n.Title, &n.Body, &metadataJSON, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
			return nil, err
		}
		results = append(results, n)
	}
	return results, rows.Err()
}

func (db *NotificationDB) MarkRead(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE notifications SET read_at = now() WHERE id = $1 AND read_at IS NULL
	`, id)
	return err
}

func (db *NotificationDB) MarkAllRead(ctx context.Context, userID string) error {
	_, err := db.pool.Exec(ctx, `
		UPDATE notifications SET read_at = now() WHERE user_id = $1 AND read_at IS NULL
	`, userID)
	return err
}

func (db *NotificationDB) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL
	`, userID).Scan(&count)
	return count, err
}

func (db *NotificationDB) DeleteByUser(ctx context.Context, userID string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM notifications WHERE user_id = $1`, userID)
	return err
}
