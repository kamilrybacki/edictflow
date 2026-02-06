package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrCategoryNotFound = errors.New("category not found")

// CategoryDB implements category database operations
type CategoryDB struct {
	pool *pgxpool.Pool
}

// NewCategoryDB creates a new CategoryDB instance
func NewCategoryDB(pool *pgxpool.Pool) *CategoryDB {
	return &CategoryDB{pool: pool}
}

// Create inserts a new category into the database
func (db *CategoryDB) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	query := `
		INSERT INTO categories (name, is_system, org_id, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, is_system, org_id, display_order, created_at, updated_at
	`

	var result domain.Category
	err := db.pool.QueryRow(ctx, query,
		category.Name,
		category.IsSystem,
		category.OrgID,
		category.DisplayOrder,
	).Scan(
		&result.ID,
		&result.Name,
		&result.IsSystem,
		&result.OrgID,
		&result.DisplayOrder,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return domain.Category{}, err
	}

	return result, nil
}

// GetByID retrieves a category by ID
func (db *CategoryDB) GetByID(ctx context.Context, id string) (domain.Category, error) {
	query := `
		SELECT id, name, is_system, org_id, display_order, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var result domain.Category
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Name,
		&result.IsSystem,
		&result.OrgID,
		&result.DisplayOrder,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, ErrCategoryNotFound
		}
		return domain.Category{}, err
	}

	return result, nil
}

// List retrieves all categories (system + org-specific)
func (db *CategoryDB) List(ctx context.Context, orgID *string) ([]domain.Category, error) {
	query := `
		SELECT id, name, is_system, org_id, display_order, created_at, updated_at
		FROM categories
		WHERE is_system = TRUE OR org_id = $1
		ORDER BY display_order, name
	`

	rows, err := db.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.IsSystem,
			&c.OrgID,
			&c.DisplayOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// ListAll retrieves all categories regardless of org
func (db *CategoryDB) ListAll(ctx context.Context) ([]domain.Category, error) {
	query := `
		SELECT id, name, is_system, org_id, display_order, created_at, updated_at
		FROM categories
		ORDER BY display_order, name
	`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.IsSystem,
			&c.OrgID,
			&c.DisplayOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// Update modifies an existing category (only non-system categories)
func (db *CategoryDB) Update(ctx context.Context, category domain.Category) error {
	query := `
		UPDATE categories
		SET name = $1, display_order = $2, updated_at = NOW()
		WHERE id = $3 AND is_system = FALSE
	`

	result, err := db.pool.Exec(ctx, query, category.Name, category.DisplayOrder, category.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// Delete removes a category by ID (only non-system categories)
func (db *CategoryDB) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = $1 AND is_system = FALSE`

	result, err := db.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
