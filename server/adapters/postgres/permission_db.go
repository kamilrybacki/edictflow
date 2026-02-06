package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

type PermissionDB struct {
	pool *pgxpool.Pool
}

func NewPermissionDB(pool *pgxpool.Pool) *PermissionDB {
	return &PermissionDB{pool: pool}
}

func (db *PermissionDB) List(ctx context.Context) ([]domain.Permission, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, code, description, category, created_at
		FROM permissions ORDER BY category, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.Description, &p.Category, &p.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, rows.Err()
}

func (db *PermissionDB) GetByCode(ctx context.Context, code string) (domain.Permission, error) {
	var p domain.Permission
	err := db.pool.QueryRow(ctx, `
		SELECT id, code, description, category, created_at
		FROM permissions WHERE code = $1
	`, code).Scan(&p.ID, &p.Code, &p.Description, &p.Category, &p.CreatedAt)
	return p, err
}
