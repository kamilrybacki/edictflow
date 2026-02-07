package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrAttachmentNotFound = errors.New("attachment not found")
var ErrAttachmentExists = errors.New("attachment already exists for this rule and team")

type RuleAttachmentDB struct {
	pool *pgxpool.Pool
}

func NewRuleAttachmentDB(pool *pgxpool.Pool) *RuleAttachmentDB {
	return &RuleAttachmentDB{pool: pool}
}

func (db *RuleAttachmentDB) Create(ctx context.Context, attachment domain.RuleAttachment) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO rule_attachments (
			id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, attachment.ID, attachment.RuleID, attachment.TeamID, attachment.EnforcementMode,
		attachment.TemporaryTimeoutHours, attachment.Status, attachment.RequestedBy,
		attachment.ApprovedBy, attachment.CreatedAt, attachment.ApprovedAt)

	if err != nil && err.Error() == "ERROR: duplicate key value violates unique constraint \"rule_attachments_rule_id_team_id_key\" (SQLSTATE 23505)" {
		return ErrAttachmentExists
	}
	return err
}

func (db *RuleAttachmentDB) GetByID(ctx context.Context, id string) (domain.RuleAttachment, error) {
	var att domain.RuleAttachment
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE id = $1
	`, id).Scan(
		&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
		&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RuleAttachment{}, ErrAttachmentNotFound
	}
	return att, err
}

func (db *RuleAttachmentDB) GetByRuleAndTeam(ctx context.Context, ruleID, teamID string) (domain.RuleAttachment, error) {
	var att domain.RuleAttachment
	err := db.pool.QueryRow(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE rule_id = $1 AND team_id = $2
	`, ruleID, teamID).Scan(
		&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
		&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RuleAttachment{}, ErrAttachmentNotFound
	}
	return att, err
}

func (db *RuleAttachmentDB) ListByTeam(ctx context.Context, teamID string) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE team_id = $1
		ORDER BY created_at DESC
	`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) ListByRule(ctx context.Context, ruleID string) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE rule_id = $1
		ORDER BY created_at DESC
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) ListByStatus(ctx context.Context, status domain.AttachmentStatus) ([]domain.RuleAttachment, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, rule_id, team_id, enforcement_mode, temporary_timeout_hours,
			status, requested_by, approved_by, created_at, approved_at
		FROM rule_attachments WHERE status = $1
		ORDER BY created_at ASC
	`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return db.scanAttachments(rows)
}

func (db *RuleAttachmentDB) Update(ctx context.Context, attachment domain.RuleAttachment) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE rule_attachments
		SET enforcement_mode = $2, temporary_timeout_hours = $3, status = $4,
			approved_by = $5, approved_at = $6
		WHERE id = $1
	`, attachment.ID, attachment.EnforcementMode, attachment.TemporaryTimeoutHours,
		attachment.Status, attachment.ApprovedBy, attachment.ApprovedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (db *RuleAttachmentDB) Delete(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `DELETE FROM rule_attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

func (db *RuleAttachmentDB) scanAttachments(rows pgx.Rows) ([]domain.RuleAttachment, error) {
	var attachments []domain.RuleAttachment
	for rows.Next() {
		var att domain.RuleAttachment
		if err := rows.Scan(
			&att.ID, &att.RuleID, &att.TeamID, &att.EnforcementMode, &att.TemporaryTimeoutHours,
			&att.Status, &att.RequestedBy, &att.ApprovedBy, &att.CreatedAt, &att.ApprovedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, rows.Err()
}
