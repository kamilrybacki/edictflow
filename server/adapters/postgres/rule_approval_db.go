package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type RuleApprovalDB struct {
	pool *pgxpool.Pool
}

func NewRuleApprovalDB(pool *pgxpool.Pool) *RuleApprovalDB {
	return &RuleApprovalDB{pool: pool}
}

func (db *RuleApprovalDB) Create(ctx context.Context, approval domain.RuleApproval) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO rule_approvals (id, rule_id, user_id, decision, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, approval.ID, approval.RuleID, approval.UserID, approval.Decision, approval.Comment, approval.CreatedAt)
	return err
}

func (db *RuleApprovalDB) ListByRule(ctx context.Context, ruleID string) ([]domain.RuleApproval, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT ra.id, ra.rule_id, ra.user_id, u.name, ra.decision, ra.comment, ra.created_at
		FROM rule_approvals ra
		LEFT JOIN users u ON ra.user_id = u.id
		WHERE ra.rule_id = $1
		ORDER BY ra.created_at DESC
	`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []domain.RuleApproval
	for rows.Next() {
		var a domain.RuleApproval
		var userName *string
		if err := rows.Scan(&a.ID, &a.RuleID, &a.UserID, &userName, &a.Decision, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		if userName != nil {
			a.UserName = *userName
		}
		approvals = append(approvals, a)
	}
	return approvals, rows.Err()
}

func (db *RuleApprovalDB) CountApprovals(ctx context.Context, ruleID string) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM rule_approvals WHERE rule_id = $1 AND decision = 'approved'
	`, ruleID).Scan(&count)
	return count, err
}

func (db *RuleApprovalDB) HasUserApproved(ctx context.Context, ruleID, userID string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM rule_approvals WHERE rule_id = $1 AND user_id = $2)
	`, ruleID, userID).Scan(&exists)
	return exists, err
}

func (db *RuleApprovalDB) DeleteByRule(ctx context.Context, ruleID string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM rule_approvals WHERE rule_id = $1`, ruleID)
	return err
}
