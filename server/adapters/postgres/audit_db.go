package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type AuditDB struct {
	pool *pgxpool.Pool
}

func NewAuditDB(pool *pgxpool.Pool) *AuditDB {
	return &AuditDB{pool: pool}
}

func (db *AuditDB) Create(ctx context.Context, entry domain.AuditEntry) error {
	changesJSON, _ := json.Marshal(entry.Changes)
	metadataJSON, _ := json.Marshal(entry.Metadata)

	_, err := db.pool.Exec(ctx, `
		INSERT INTO audit_entries (id, entity_type, entity_id, action, actor_id, changes, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, entry.ID, entry.EntityType, entry.EntityID, entry.Action, entry.ActorID, changesJSON, metadataJSON, entry.CreatedAt)
	return err
}

type AuditListParams struct {
	EntityType *domain.AuditEntityType
	EntityID   *string
	ActorID    *string
	Action     *domain.AuditAction
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

func (db *AuditDB) List(ctx context.Context, params AuditListParams) ([]domain.AuditEntry, int, error) {
	query := `SELECT id, entity_type, entity_id, action, actor_id, changes, metadata, created_at FROM audit_entries WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_entries WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if params.EntityType != nil {
		filter := fmt.Sprintf(` AND entity_type = $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.EntityType)
		argNum++
	}
	if params.EntityID != nil {
		filter := fmt.Sprintf(` AND entity_id = $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.EntityID)
		argNum++
	}
	if params.ActorID != nil {
		filter := fmt.Sprintf(` AND actor_id = $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.ActorID)
		argNum++
	}
	if params.Action != nil {
		filter := fmt.Sprintf(` AND action = $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.Action)
		argNum++
	}
	if params.From != nil {
		filter := fmt.Sprintf(` AND created_at >= $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.From)
		argNum++
	}
	if params.To != nil {
		filter := fmt.Sprintf(` AND created_at <= $%d`, argNum)
		query += filter
		countQuery += filter
		args = append(args, *params.To)
		argNum++
	}

	// Get total count
	var total int
	if err := db.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += ` ORDER BY created_at DESC`
	if params.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argNum)
		args = append(args, params.Limit)
		argNum++
	}
	if params.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argNum)
		args = append(args, params.Offset)
	}

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := make([]domain.AuditEntry, 0, 32) // Preallocate with reasonable capacity
	for rows.Next() {
		var entry domain.AuditEntry
		var changesJSON, metadataJSON []byte
		if err := rows.Scan(&entry.ID, &entry.EntityType, &entry.EntityID, &entry.Action, &entry.ActorID, &changesJSON, &metadataJSON, &entry.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal(changesJSON, &entry.Changes)
		_ = json.Unmarshal(metadataJSON, &entry.Metadata)
		entries = append(entries, entry)
	}
	return entries, total, rows.Err()
}

func (db *AuditDB) GetEntityHistory(ctx context.Context, entityType domain.AuditEntityType, entityID string) ([]domain.AuditEntry, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT ae.id, ae.entity_type, ae.entity_id, ae.action, ae.actor_id, u.name, ae.changes, ae.metadata, ae.created_at
		FROM audit_entries ae
		LEFT JOIN users u ON ae.actor_id = u.id
		WHERE ae.entity_type = $1 AND ae.entity_id = $2
		ORDER BY ae.created_at DESC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.AuditEntry, 0, 32) // Preallocate with reasonable capacity
	for rows.Next() {
		var entry domain.AuditEntry
		var actorName *string
		var changesJSON, metadataJSON []byte
		if err := rows.Scan(&entry.ID, &entry.EntityType, &entry.EntityID, &entry.Action, &entry.ActorID, &actorName, &changesJSON, &metadataJSON, &entry.CreatedAt); err != nil {
			return nil, err
		}
		if actorName != nil {
			entry.ActorName = *actorName
		}
		json.Unmarshal(changesJSON, &entry.Changes)
		json.Unmarshal(metadataJSON, &entry.Metadata)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
