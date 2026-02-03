package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/claudeception/server/domain"
)

var ErrRoleNotFound = errors.New("role not found")

type RoleDB struct {
	pool *pgxpool.Pool
}

func NewRoleDB(pool *pgxpool.Pool) *RoleDB {
	return &RoleDB{pool: pool}
}

func (db *RoleDB) Create(ctx context.Context, role domain.RoleEntity) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO roles (id, name, description, hierarchy_level, parent_role_id, team_id, is_system, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, role.ID, role.Name, role.Description, role.HierarchyLevel, role.ParentRoleID, role.TeamID, role.IsSystem, role.CreatedAt)
	return err
}

func (db *RoleDB) GetByID(ctx context.Context, id string) (domain.RoleEntity, error) {
	var role domain.RoleEntity
	err := db.pool.QueryRow(ctx, `
		SELECT id, name, description, hierarchy_level, parent_role_id, team_id, is_system, created_at
		FROM roles WHERE id = $1
	`, id).Scan(&role.ID, &role.Name, &role.Description, &role.HierarchyLevel, &role.ParentRoleID, &role.TeamID, &role.IsSystem, &role.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RoleEntity{}, ErrRoleNotFound
	}
	return role, err
}

func (db *RoleDB) List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error) {
	query := `SELECT id, name, description, hierarchy_level, parent_role_id, team_id, is_system, created_at FROM roles WHERE team_id IS NULL`
	args := []interface{}{}

	if teamID != nil {
		query += ` OR team_id = $1`
		args = append(args, *teamID)
	}
	query += ` ORDER BY hierarchy_level DESC, name`

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.RoleEntity
	for rows.Next() {
		var role domain.RoleEntity
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.HierarchyLevel, &role.ParentRoleID, &role.TeamID, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (db *RoleDB) Update(ctx context.Context, role domain.RoleEntity) error {
	result, err := db.pool.Exec(ctx, `
		UPDATE roles SET name = $2, description = $3, hierarchy_level = $4, parent_role_id = $5
		WHERE id = $1 AND is_system = false
	`, role.ID, role.Name, role.Description, role.HierarchyLevel, role.ParentRoleID)

	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (db *RoleDB) Delete(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1 AND is_system = false`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (db *RoleDB) GetPermissions(ctx context.Context, roleID string) ([]domain.Permission, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT p.id, p.code, p.description, p.category, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.category, p.code
	`, roleID)
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

func (db *RoleDB) AddPermission(ctx context.Context, roleID, permissionID string) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, roleID, permissionID)
	return err
}

func (db *RoleDB) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2
	`, roleID, permissionID)
	return err
}

func (db *RoleDB) GetUserRoles(ctx context.Context, userID string) ([]domain.RoleEntity, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT r.id, r.name, r.description, r.hierarchy_level, r.parent_role_id, r.team_id, r.is_system, r.created_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.hierarchy_level DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.RoleEntity
	for rows.Next() {
		var role domain.RoleEntity
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.HierarchyLevel, &role.ParentRoleID, &role.TeamID, &role.IsSystem, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (db *RoleDB) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, userID, roleID, assignedBy)
	return err
}

func (db *RoleDB) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2
	`, userID, roleID)
	return err
}

func (db *RoleDB) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT DISTINCT p.code
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}
	return permissions, rows.Err()
}
