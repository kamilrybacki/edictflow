//go:build integration

package testhelpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
)

// Fixtures provides test data creation utilities
type Fixtures struct {
	pool *pgxpool.Pool
}

// NewFixtures creates a new Fixtures instance
func NewFixtures(pool *pgxpool.Pool) *Fixtures {
	return &Fixtures{pool: pool}
}

// CreateTeam creates a team in the database and returns it
func (f *Fixtures) CreateTeam(ctx context.Context, name string) (domain.Team, error) {
	team := domain.Team{
		ID:        uuid.New().String(),
		Name:      name,
		Settings:  domain.TeamSettings{DriftThresholdMinutes: 60},
		CreatedAt: time.Now(),
	}

	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return domain.Team{}, err
	}

	_, err = f.pool.Exec(ctx, `
		INSERT INTO teams (id, name, settings, created_at)
		VALUES ($1, $2, $3, $4)
	`, team.ID, team.Name, settingsJSON, team.CreatedAt)
	if err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

// CreateRule creates a rule in the database and returns it
func (f *Fixtures) CreateRule(ctx context.Context, name, teamID string) (domain.Rule, error) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "*.go"},
	}

	rule := domain.Rule{
		ID:              uuid.New().String(),
		Name:            name,
		Content:         "Test rule content",
		TargetLayer:     domain.TargetLayerProject,
		PriorityWeight:  0,
		Triggers:        triggers,
		TeamID:          teamID,
		Status:          domain.RuleStatusDraft,
		EnforcementMode: domain.EnforcementModeBlock,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	triggersJSON, err := json.Marshal(rule.Triggers)
	if err != nil {
		return domain.Rule{}, err
	}

	_, err = f.pool.Exec(ctx, `
		INSERT INTO rules (id, name, content, target_layer, priority_weight, triggers, team_id, status, enforcement_mode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, rule.ID, rule.Name, rule.Content, rule.TargetLayer, rule.PriorityWeight, triggersJSON, rule.TeamID, rule.Status, rule.EnforcementMode, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

// CreateUser creates a user in the database and returns the user
func (f *Fixtures) CreateUser(ctx context.Context, email, teamID string) (domain.User, error) {
	user := domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         "Test User",
		TeamID:       &teamID,
		AuthProvider: domain.AuthProviderLocal,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}

	_, err := f.pool.Exec(ctx, `
		INSERT INTO users (id, email, name, team_id, auth_provider, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Email, user.Name, user.TeamID, user.AuthProvider, user.IsActive, user.CreatedAt)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// CreateRole creates a role in the database and returns it
func (f *Fixtures) CreateRole(ctx context.Context, name string, hierarchyLevel int, teamID *string) (domain.RoleEntity, error) {
	role := domain.RoleEntity{
		ID:             uuid.New().String(),
		Name:           name,
		Description:    "Test role description",
		HierarchyLevel: hierarchyLevel,
		TeamID:         teamID,
		IsSystem:       false,
		CreatedAt:      time.Now(),
	}

	_, err := f.pool.Exec(ctx, `
		INSERT INTO roles (id, name, description, hierarchy_level, team_id, is_system, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, role.ID, role.Name, role.Description, role.HierarchyLevel, role.TeamID, role.IsSystem, role.CreatedAt)
	if err != nil {
		return domain.RoleEntity{}, err
	}

	return role, nil
}

// CreatePermission creates a permission in the database and returns it
func (f *Fixtures) CreatePermission(ctx context.Context, code, description string, category domain.PermissionCategory) (domain.Permission, error) {
	perm := domain.Permission{
		ID:          uuid.New().String(),
		Code:        code,
		Description: description,
		Category:    category,
		CreatedAt:   time.Now(),
	}

	_, err := f.pool.Exec(ctx, `
		INSERT INTO permissions (id, code, description, category, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, perm.ID, perm.Code, perm.Description, perm.Category, perm.CreatedAt)
	if err != nil {
		return domain.Permission{}, err
	}

	return perm, nil
}

// AssignRoleToUser assigns a role to a user
func (f *Fixtures) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	_, err := f.pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, assigned_at)
		VALUES ($1, $2, NOW())
	`, userID, roleID)
	return err
}

// AddPermissionToRole adds a permission to a role
func (f *Fixtures) AddPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	_, err := f.pool.Exec(ctx, `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
	`, roleID, permissionID)
	return err
}

// CreateApprovalConfig creates an approval config
func (f *Fixtures) CreateApprovalConfig(ctx context.Context, targetLayer domain.TargetLayer, requiredApprovals int, teamID *string) error {
	id := uuid.New().String()
	_, err := f.pool.Exec(ctx, `
		INSERT INTO approval_configs (id, target_layer, required_approvals, team_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, id, targetLayer, requiredApprovals, teamID)
	return err
}

// CreateCategory creates a category in the database and returns it
func (f *Fixtures) CreateCategory(ctx context.Context, name string, isSystem bool) (domain.Category, error) {
	category := domain.Category{
		ID:           uuid.New().String(),
		Name:         name,
		IsSystem:     isSystem,
		DisplayOrder: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := f.pool.Exec(ctx, `
		INSERT INTO categories (id, name, is_system, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, category.ID, category.Name, category.IsSystem, category.DisplayOrder, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return domain.Category{}, err
	}

	return category, nil
}

// CreateRuleWithCategory creates a rule with a category in the database and returns it
func (f *Fixtures) CreateRuleWithCategory(ctx context.Context, name, teamID string, categoryID *string, targetLayer domain.TargetLayer, overridable bool) (domain.Rule, error) {
	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "*.go"},
	}

	rule := domain.Rule{
		ID:             uuid.New().String(),
		Name:           name,
		Content:        "Test rule content for " + name,
		TargetLayer:    targetLayer,
		CategoryID:     categoryID,
		PriorityWeight: 0,
		Overridable:    overridable,
		Triggers:       triggers,
		TeamID:         teamID,
		Status:         domain.RuleStatusApproved,
		EnforcementMode: domain.EnforcementModeBlock,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	triggersJSON, err := json.Marshal(rule.Triggers)
	if err != nil {
		return domain.Rule{}, err
	}

	_, err = f.pool.Exec(ctx, `
		INSERT INTO rules (id, name, content, target_layer, category_id, priority_weight, overridable, triggers, team_id, status, enforcement_mode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, rule.ID, rule.Name, rule.Content, rule.TargetLayer, rule.CategoryID, rule.PriorityWeight, rule.Overridable, triggersJSON, rule.TeamID, rule.Status, rule.EnforcementMode, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}
