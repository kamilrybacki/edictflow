//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kamilrybacki/edictflow/server/adapters/postgres"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/rules"
	"github.com/kamilrybacki/edictflow/server/services/teams"
)

// Team Repository Tests

func TestTeamRepository_CreateAndGetByID(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewTeamDB(testPool)
	repo := teams.NewRepository(db)

	team := domain.Team{
		ID:        uuid.New().String(),
		Name:      "Test Team",
		Settings:  domain.TeamSettings{DriftThresholdMinutes: 30},
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}

	// Create
	err := repo.Create(ctx, team)
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(ctx, team.ID)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}

	if retrieved.ID != team.ID {
		t.Errorf("ID mismatch: expected %s, got %s", team.ID, retrieved.ID)
	}
	if retrieved.Name != team.Name {
		t.Errorf("Name mismatch: expected %s, got %s", team.Name, retrieved.Name)
	}
	if retrieved.Settings.DriftThresholdMinutes != team.Settings.DriftThresholdMinutes {
		t.Errorf("Settings mismatch: expected %d, got %d",
			team.Settings.DriftThresholdMinutes, retrieved.Settings.DriftThresholdMinutes)
	}
}

func TestTeamRepository_ListTeams(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewTeamDB(testPool)
	repo := teams.NewRepository(db)

	// Create multiple teams
	teamNames := []string{"Team Alpha", "Team Beta", "Team Gamma"}
	for _, name := range teamNames {
		team := domain.NewTeam(name)
		if err := repo.Create(ctx, team); err != nil {
			t.Fatalf("Failed to create team %s: %v", name, err)
		}
	}

	// List all teams
	teamsList, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list teams: %v", err)
	}

	if len(teamsList) != len(teamNames) {
		t.Errorf("Expected %d teams, got %d", len(teamNames), len(teamsList))
	}
}

func TestTeamRepository_UpdateTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewTeamDB(testPool)
	repo := teams.NewRepository(db)

	// Create team
	team := domain.NewTeam("Original Name")
	if err := repo.Create(ctx, team); err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Update team
	team.Name = "Updated Name"
	team.Settings.DriftThresholdMinutes = 120
	if err := repo.Update(ctx, team); err != nil {
		t.Fatalf("Failed to update team: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, team.ID)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Name not updated: expected 'Updated Name', got %s", retrieved.Name)
	}
	if retrieved.Settings.DriftThresholdMinutes != 120 {
		t.Errorf("Settings not updated: expected 120, got %d", retrieved.Settings.DriftThresholdMinutes)
	}
}

func TestTeamRepository_DeleteTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewTeamDB(testPool)
	repo := teams.NewRepository(db)

	// Create team
	team := domain.NewTeam("Team to Delete")
	if err := repo.Create(ctx, team); err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Delete team
	if err := repo.Delete(ctx, team.ID); err != nil {
		t.Fatalf("Failed to delete team: %v", err)
	}

	// Verify deletion
	_, err := repo.GetByID(ctx, team.ID)
	if err != teams.ErrTeamNotFound {
		t.Errorf("Expected ErrTeamNotFound, got %v", err)
	}
}

func TestTeamRepository_GetNonExistent(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewTeamDB(testPool)
	repo := teams.NewRepository(db)

	_, err := repo.GetByID(ctx, uuid.New().String())
	if err != teams.ErrTeamNotFound {
		t.Errorf("Expected ErrTeamNotFound, got %v", err)
	}
}

// Rule Repository Tests

func TestRuleRepository_CreateAndGetByID(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	// First create a team (foreign key requirement)
	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	db := postgres.NewRuleDB(testPool)
	repo := rules.NewRepository(db)

	triggers := []domain.Trigger{
		{Type: domain.TriggerTypePath, Pattern: "*.go"},
		{Type: domain.TriggerTypeContext, ContextTypes: []string{"debug"}},
	}

	rule := domain.Rule{
		ID:              uuid.New().String(),
		Name:            "Test Rule",
		Content:         "This is test rule content",
		TargetLayer:     domain.TargetLayerProject,
		PriorityWeight:  10,
		Triggers:        triggers,
		TeamID:          team.ID,
		Status:          domain.RuleStatusDraft,
		EnforcementMode: domain.EnforcementModeBlock,
		CreatedAt:       time.Now().Truncate(time.Microsecond),
		UpdatedAt:       time.Now().Truncate(time.Microsecond),
	}

	// Create
	err = repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.ID != rule.ID {
		t.Errorf("ID mismatch: expected %s, got %s", rule.ID, retrieved.ID)
	}
	if retrieved.Name != rule.Name {
		t.Errorf("Name mismatch: expected %s, got %s", rule.Name, retrieved.Name)
	}
	if retrieved.Content != rule.Content {
		t.Errorf("Content mismatch: expected %s, got %s", rule.Content, retrieved.Content)
	}
	if retrieved.TargetLayer != rule.TargetLayer {
		t.Errorf("TargetLayer mismatch: expected %s, got %s", rule.TargetLayer, retrieved.TargetLayer)
	}
	if len(retrieved.Triggers) != len(rule.Triggers) {
		t.Errorf("Triggers count mismatch: expected %d, got %d", len(rule.Triggers), len(retrieved.Triggers))
	}
}

func TestRuleRepository_ListByTeam(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	// Create two teams
	team1, err := testFixtures.CreateTeam(ctx, "Team 1")
	if err != nil {
		t.Fatalf("Failed to create team1: %v", err)
	}

	team2, err := testFixtures.CreateTeam(ctx, "Team 2")
	if err != nil {
		t.Fatalf("Failed to create team2: %v", err)
	}

	// Create rules for team1
	for i := 0; i < 3; i++ {
		if _, err := testFixtures.CreateRule(ctx, "Team1 Rule", team1.ID); err != nil {
			t.Fatalf("Failed to create rule for team1: %v", err)
		}
	}

	// Create rules for team2
	for i := 0; i < 2; i++ {
		if _, err := testFixtures.CreateRule(ctx, "Team2 Rule", team2.ID); err != nil {
			t.Fatalf("Failed to create rule for team2: %v", err)
		}
	}

	db := postgres.NewRuleDB(testPool)
	repo := rules.NewRepository(db)

	// List rules for team1
	team1Rules, err := repo.ListByTeam(ctx, team1.ID)
	if err != nil {
		t.Fatalf("Failed to list rules for team1: %v", err)
	}

	if len(team1Rules) != 3 {
		t.Errorf("Expected 3 rules for team1, got %d", len(team1Rules))
	}

	// List rules for team2
	team2Rules, err := repo.ListByTeam(ctx, team2.ID)
	if err != nil {
		t.Fatalf("Failed to list rules for team2: %v", err)
	}

	if len(team2Rules) != 2 {
		t.Errorf("Expected 2 rules for team2, got %d", len(team2Rules))
	}
}

func TestRuleRepository_UpdateRule(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	rule, err := testFixtures.CreateRule(ctx, "Original Rule", team.ID)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	db := postgres.NewRuleDB(testPool)
	repo := rules.NewRepository(db)

	// Update rule
	rule.Name = "Updated Rule"
	rule.Content = "Updated content"
	rule.PriorityWeight = 50
	rule.UpdatedAt = time.Now().Truncate(time.Microsecond)

	if err := repo.Update(ctx, rule); err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.Name != "Updated Rule" {
		t.Errorf("Name not updated: expected 'Updated Rule', got %s", retrieved.Name)
	}
	if retrieved.Content != "Updated content" {
		t.Errorf("Content not updated: expected 'Updated content', got %s", retrieved.Content)
	}
	if retrieved.PriorityWeight != 50 {
		t.Errorf("PriorityWeight not updated: expected 50, got %d", retrieved.PriorityWeight)
	}
}

func TestRuleRepository_DeleteRule(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	rule, err := testFixtures.CreateRule(ctx, "Rule to Delete", team.ID)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	db := postgres.NewRuleDB(testPool)
	repo := rules.NewRepository(db)

	// Delete rule
	if err := repo.Delete(ctx, rule.ID); err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, rule.ID)
	if err != rules.ErrRuleNotFound {
		t.Errorf("Expected ErrRuleNotFound, got %v", err)
	}
}

func TestRuleRepository_CascadeDeleteOnTeamDelete(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	rule, err := testFixtures.CreateRule(ctx, "Rule", team.ID)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	teamDB := postgres.NewTeamDB(testPool)
	teamRepo := teams.NewRepository(teamDB)

	ruleDB := postgres.NewRuleDB(testPool)
	ruleRepo := rules.NewRepository(ruleDB)

	// Delete team (should cascade delete rules)
	if err := teamRepo.Delete(ctx, team.ID); err != nil {
		t.Fatalf("Failed to delete team: %v", err)
	}

	// Verify rule was also deleted
	_, err = ruleRepo.GetByID(ctx, rule.ID)
	if err != rules.ErrRuleNotFound {
		t.Errorf("Expected ErrRuleNotFound after cascade delete, got %v", err)
	}
}

// Role Repository Tests

func TestRoleRepository_CreateAndGetByID(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	db := postgres.NewRoleDB(testPool)

	role := domain.NewRole("Test Admin", "Test admin role", 100, nil, nil)

	// Create
	if err := db.Create(ctx, role); err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// Get by ID
	retrieved, err := db.GetByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("Failed to get role: %v", err)
	}

	if retrieved.ID != role.ID {
		t.Errorf("ID mismatch: expected %s, got %s", role.ID, retrieved.ID)
	}
	if retrieved.Name != role.Name {
		t.Errorf("Name mismatch: expected %s, got %s", role.Name, retrieved.Name)
	}
	if retrieved.HierarchyLevel != role.HierarchyLevel {
		t.Errorf("HierarchyLevel mismatch: expected %d, got %d", role.HierarchyLevel, retrieved.HierarchyLevel)
	}
}

func TestRoleRepository_ListRoles(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	// Create roles
	role1, err := testFixtures.CreateRole(ctx, "Admin", 100, nil)
	if err != nil {
		t.Fatalf("Failed to create role1: %v", err)
	}

	role2, err := testFixtures.CreateRole(ctx, "Member", 50, nil)
	if err != nil {
		t.Fatalf("Failed to create role2: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// List all roles (nil team ID gets global roles)
	roles, err := db.List(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to list roles: %v", err)
	}

	if len(roles) < 2 {
		t.Errorf("Expected at least 2 roles, got %d", len(roles))
	}

	// Verify ordering by hierarchy level
	var foundAdmin, foundMember bool
	for _, r := range roles {
		if r.ID == role1.ID {
			foundAdmin = true
		}
		if r.ID == role2.ID {
			foundMember = true
		}
	}
	if !foundAdmin || !foundMember {
		t.Error("Expected to find both created roles")
	}
}

func TestRoleRepository_UpdateRole(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	role, err := testFixtures.CreateRole(ctx, "Original Name", 50, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Update role
	role.Name = "Updated Name"
	role.Description = "Updated description"
	if err := db.Update(ctx, role); err != nil {
		t.Fatalf("Failed to update role: %v", err)
	}

	// Verify update
	retrieved, err := db.GetByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("Failed to get role: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Name not updated: expected 'Updated Name', got %s", retrieved.Name)
	}
}

func TestRoleRepository_DeleteRole(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	role, err := testFixtures.CreateRole(ctx, "Role to Delete", 50, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Delete role
	if err := db.Delete(ctx, role.ID); err != nil {
		t.Fatalf("Failed to delete role: %v", err)
	}

	// Verify deletion
	_, err = db.GetByID(ctx, role.ID)
	if err != postgres.ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestRoleRepository_AssignUserRole(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	user, err := testFixtures.CreateUser(ctx, "test@example.com", team.ID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role, err := testFixtures.CreateRole(ctx, "Admin", 100, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Assign role to user
	if err := db.AssignUserRole(ctx, user.ID, role.ID, nil); err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Verify assignment
	userRoles, err := db.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user roles: %v", err)
	}

	if len(userRoles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(userRoles))
	}
	if userRoles[0].ID != role.ID {
		t.Errorf("Expected role ID %s, got %s", role.ID, userRoles[0].ID)
	}
}

func TestRoleRepository_RemoveUserRole(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	user, err := testFixtures.CreateUser(ctx, "test@example.com", team.ID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role, err := testFixtures.CreateRole(ctx, "Admin", 100, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Assign and then remove
	if err := db.AssignUserRole(ctx, user.ID, role.ID, nil); err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	if err := db.RemoveUserRole(ctx, user.ID, role.ID); err != nil {
		t.Fatalf("Failed to remove role: %v", err)
	}

	// Verify removal
	userRoles, err := db.GetUserRoles(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user roles: %v", err)
	}

	if len(userRoles) != 0 {
		t.Errorf("Expected 0 roles after removal, got %d", len(userRoles))
	}
}

func TestRoleRepository_RolePermissions(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	role, err := testFixtures.CreateRole(ctx, "Admin", 100, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	perm1, err := testFixtures.CreatePermission(ctx, "rules:read", "Read rules", domain.PermissionCategoryRules)
	if err != nil {
		t.Fatalf("Failed to create permission1: %v", err)
	}

	perm2, err := testFixtures.CreatePermission(ctx, "rules:write", "Write rules", domain.PermissionCategoryRules)
	if err != nil {
		t.Fatalf("Failed to create permission2: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Add permissions
	if err := db.AddPermission(ctx, role.ID, perm1.ID); err != nil {
		t.Fatalf("Failed to add permission1: %v", err)
	}
	if err := db.AddPermission(ctx, role.ID, perm2.ID); err != nil {
		t.Fatalf("Failed to add permission2: %v", err)
	}

	// Get permissions
	perms, err := db.GetPermissions(ctx, role.ID)
	if err != nil {
		t.Fatalf("Failed to get permissions: %v", err)
	}

	if len(perms) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(perms))
	}

	// Remove a permission
	if err := db.RemovePermission(ctx, role.ID, perm1.ID); err != nil {
		t.Fatalf("Failed to remove permission: %v", err)
	}

	// Verify removal
	perms, err = db.GetPermissions(ctx, role.ID)
	if err != nil {
		t.Fatalf("Failed to get permissions: %v", err)
	}

	if len(perms) != 1 {
		t.Errorf("Expected 1 permission after removal, got %d", len(perms))
	}
}

func TestRoleRepository_GetUserPermissions(t *testing.T) {
	resetDB(t)
	ctx := context.Background()

	team, err := testFixtures.CreateTeam(ctx, "Test Team")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	user, err := testFixtures.CreateUser(ctx, "test@example.com", team.ID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	role, err := testFixtures.CreateRole(ctx, "Admin", 100, nil)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	perm1, err := testFixtures.CreatePermission(ctx, "test:rules:read", "Read rules", domain.PermissionCategoryRules)
	if err != nil {
		t.Fatalf("Failed to create permission1: %v", err)
	}

	perm2, err := testFixtures.CreatePermission(ctx, "test:users:read", "Read users", domain.PermissionCategoryUsers)
	if err != nil {
		t.Fatalf("Failed to create permission2: %v", err)
	}

	db := postgres.NewRoleDB(testPool)

	// Add permissions to role
	if err := db.AddPermission(ctx, role.ID, perm1.ID); err != nil {
		t.Fatalf("Failed to add permission1: %v", err)
	}
	if err := db.AddPermission(ctx, role.ID, perm2.ID); err != nil {
		t.Fatalf("Failed to add permission2: %v", err)
	}

	// Assign role to user
	if err := db.AssignUserRole(ctx, user.ID, role.ID, nil); err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Get user permissions
	userPerms, err := db.GetUserPermissions(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user permissions: %v", err)
	}

	if len(userPerms) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(userPerms))
	}

	// Verify permission codes
	permCodes := make(map[string]bool)
	for _, p := range userPerms {
		permCodes[p] = true
	}

	if !permCodes["test:rules:read"] {
		t.Error("Expected test:rules:read permission")
	}
	if !permCodes["test:users:read"] {
		t.Error("Expected test:users:read permission")
	}
}
