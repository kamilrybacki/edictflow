package roles

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
)

type mockRoleDB struct {
	roles           map[string]domain.RoleEntity
	rolePermissions map[string][]domain.Permission
	userRoles       map[string][]string // userID -> []roleID
}

func (m *mockRoleDB) Create(ctx context.Context, role domain.RoleEntity) error {
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleDB) GetByID(ctx context.Context, id string) (domain.RoleEntity, error) {
	if role, ok := m.roles[id]; ok {
		return role, nil
	}
	return domain.RoleEntity{}, ErrRoleNotFound
}

func (m *mockRoleDB) List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error) {
	var result []domain.RoleEntity
	for _, role := range m.roles {
		if teamID == nil || role.TeamID == nil || *role.TeamID == *teamID {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *mockRoleDB) Update(ctx context.Context, role domain.RoleEntity) error {
	if _, ok := m.roles[role.ID]; !ok {
		return ErrRoleNotFound
	}
	m.roles[role.ID] = role
	return nil
}

func (m *mockRoleDB) Delete(ctx context.Context, id string) error {
	if _, ok := m.roles[id]; !ok {
		return ErrRoleNotFound
	}
	delete(m.roles, id)
	return nil
}

func (m *mockRoleDB) GetPermissions(ctx context.Context, roleID string) ([]domain.Permission, error) {
	if perms, ok := m.rolePermissions[roleID]; ok {
		return perms, nil
	}
	return []domain.Permission{}, nil
}

func (m *mockRoleDB) AddPermission(ctx context.Context, roleID, permissionID string) error {
	return nil
}

func (m *mockRoleDB) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return nil
}

func (m *mockRoleDB) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	m.userRoles[userID] = append(m.userRoles[userID], roleID)
	return nil
}

func (m *mockRoleDB) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	return nil
}

func (m *mockRoleDB) GetUserRoles(ctx context.Context, userID string) ([]domain.RoleEntity, error) {
	var result []domain.RoleEntity
	for _, roleID := range m.userRoles[userID] {
		if role, ok := m.roles[roleID]; ok {
			result = append(result, role)
		}
	}
	return result, nil
}

func (m *mockRoleDB) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil
}

type mockPermissionDB struct {
	permissions map[string]domain.Permission
}

func (m *mockPermissionDB) List(ctx context.Context) ([]domain.Permission, error) {
	var result []domain.Permission
	for _, p := range m.permissions {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPermissionDB) GetByCode(ctx context.Context, code string) (domain.Permission, error) {
	for _, p := range m.permissions {
		if p.Code == code {
			return p, nil
		}
	}
	return domain.Permission{}, ErrPermissionNotFound
}

func newMockRoleDB() *mockRoleDB {
	return &mockRoleDB{
		roles:           make(map[string]domain.RoleEntity),
		rolePermissions: make(map[string][]domain.Permission),
		userRoles:       make(map[string][]string),
	}
}

func TestService_Create(t *testing.T) {
	roleDB := newMockRoleDB()
	svc := NewService(roleDB, &mockPermissionDB{})

	role, err := svc.Create(context.Background(), "Manager", "Team manager role", 50, nil, nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if role.Name != "Manager" {
		t.Errorf("Expected name 'Manager', got '%s'", role.Name)
	}
	if role.HierarchyLevel != 50 {
		t.Errorf("Expected hierarchy level 50, got %d", role.HierarchyLevel)
	}
}

func TestService_GetByID(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	role, err := svc.GetByID(context.Background(), "role-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if role.Name != "Admin" {
		t.Errorf("Expected name 'Admin', got '%s'", role.Name)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	roleDB.roles["role-2"] = domain.RoleEntity{ID: "role-2", Name: "Member", HierarchyLevel: 1}
	svc := NewService(roleDB, &mockPermissionDB{})

	roles, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
}

func TestService_AssignUserRole(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.AssignUserRole(context.Background(), "user-1", "role-1", nil)
	if err != nil {
		t.Fatalf("AssignUserRole() error = %v", err)
	}

	if len(roleDB.userRoles["user-1"]) != 1 {
		t.Errorf("Expected user to have 1 role assigned")
	}
}

func TestService_GetUserRoles(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	roleDB.userRoles["user-1"] = []string{"role-1"}
	svc := NewService(roleDB, &mockPermissionDB{})

	roles, err := svc.GetUserRoles(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUserRoles() error = %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
}

func TestService_ListPermissions(t *testing.T) {
	permDB := &mockPermissionDB{
		permissions: map[string]domain.Permission{
			"perm-1": {ID: "perm-1", Code: "create_rules", Category: domain.PermissionCategoryRules},
		},
	}
	svc := NewService(newMockRoleDB(), permDB)

	perms, err := svc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("ListPermissions() error = %v", err)
	}
	if len(perms) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(perms))
	}
}

func TestService_Update(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	role := domain.RoleEntity{ID: "role-1", Name: "Super Admin", HierarchyLevel: 100}
	err := svc.Update(context.Background(), role)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if roleDB.roles["role-1"].Name != "Super Admin" {
		t.Errorf("Expected name 'Super Admin', got '%s'", roleDB.roles["role-1"].Name)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	role := domain.RoleEntity{ID: "nonexistent", Name: "Admin", HierarchyLevel: 100}
	err := svc.Update(context.Background(), role)
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_Update_SystemRole(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "System Admin", HierarchyLevel: 100, IsSystem: true}
	svc := NewService(roleDB, &mockPermissionDB{})

	role := domain.RoleEntity{ID: "role-1", Name: "Modified", HierarchyLevel: 100}
	err := svc.Update(context.Background(), role)
	if err != ErrCannotModifySystem {
		t.Errorf("Expected ErrCannotModifySystem, got %v", err)
	}
}

func TestService_Delete(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.Delete(context.Background(), "role-1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, ok := roleDB.roles["role-1"]; ok {
		t.Error("Expected role to be deleted")
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	err := svc.Delete(context.Background(), "nonexistent")
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_Delete_SystemRole(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "System Admin", HierarchyLevel: 100, IsSystem: true}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.Delete(context.Background(), "role-1")
	if err != ErrCannotModifySystem {
		t.Errorf("Expected ErrCannotModifySystem, got %v", err)
	}
}

func TestService_AddPermission(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.AddPermission(context.Background(), "role-1", "perm-1")
	if err != nil {
		t.Fatalf("AddPermission() error = %v", err)
	}
}

func TestService_AddPermission_RoleNotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	err := svc.AddPermission(context.Background(), "nonexistent", "perm-1")
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_RemovePermission(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.RemovePermission(context.Background(), "role-1", "perm-1")
	if err != nil {
		t.Fatalf("RemovePermission() error = %v", err)
	}
}

func TestService_RemovePermission_RoleNotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	err := svc.RemovePermission(context.Background(), "nonexistent", "perm-1")
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_RemoveUserRole(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	roleDB.userRoles["user-1"] = []string{"role-1"}
	svc := NewService(roleDB, &mockPermissionDB{})

	err := svc.RemoveUserRole(context.Background(), "user-1", "role-1")
	if err != nil {
		t.Fatalf("RemoveUserRole() error = %v", err)
	}
}

func TestService_GetPermissions(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	roleDB.rolePermissions["role-1"] = []domain.Permission{
		{ID: "perm-1", Code: "create_rules"},
	}
	svc := NewService(roleDB, &mockPermissionDB{})

	perms, err := svc.GetPermissions(context.Background(), "role-1")
	if err != nil {
		t.Fatalf("GetPermissions() error = %v", err)
	}
	if len(perms) != 1 {
		t.Errorf("Expected 1 permission, got %d", len(perms))
	}
}

func TestService_GetUserPermissions(t *testing.T) {
	roleDB := newMockRoleDB()
	svc := NewService(roleDB, &mockPermissionDB{})

	perms, err := svc.GetUserPermissions(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUserPermissions() error = %v", err)
	}
	if perms == nil {
		t.Error("Expected non-nil slice")
	}
}

func TestService_GetPermissionByCode(t *testing.T) {
	permDB := &mockPermissionDB{
		permissions: map[string]domain.Permission{
			"perm-1": {ID: "perm-1", Code: "create_rules", Category: domain.PermissionCategoryRules},
		},
	}
	svc := NewService(newMockRoleDB(), permDB)

	perm, err := svc.GetPermissionByCode(context.Background(), "create_rules")
	if err != nil {
		t.Fatalf("GetPermissionByCode() error = %v", err)
	}
	if perm.Code != "create_rules" {
		t.Errorf("Expected code 'create_rules', got '%s'", perm.Code)
	}
}

func TestService_GetPermissionByCode_NotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{permissions: make(map[string]domain.Permission)})

	_, err := svc.GetPermissionByCode(context.Background(), "nonexistent")
	if err != ErrPermissionNotFound {
		t.Errorf("Expected ErrPermissionNotFound, got %v", err)
	}
}

func TestService_GetRoleWithPermissions(t *testing.T) {
	roleDB := newMockRoleDB()
	roleDB.roles["role-1"] = domain.RoleEntity{ID: "role-1", Name: "Admin", HierarchyLevel: 100}
	roleDB.rolePermissions["role-1"] = []domain.Permission{
		{ID: "perm-1", Code: "create_rules"},
		{ID: "perm-2", Code: "delete_rules"},
	}
	svc := NewService(roleDB, &mockPermissionDB{})

	role, err := svc.GetRoleWithPermissions(context.Background(), "role-1")
	if err != nil {
		t.Fatalf("GetRoleWithPermissions() error = %v", err)
	}
	if role.Name != "Admin" {
		t.Errorf("Expected name 'Admin', got '%s'", role.Name)
	}
	if len(role.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(role.Permissions))
	}
}

func TestService_GetRoleWithPermissions_NotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	_, err := svc.GetRoleWithPermissions(context.Background(), "nonexistent")
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_AssignUserRole_RoleNotFound(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	err := svc.AssignUserRole(context.Background(), "user-1", "nonexistent", nil)
	if err != ErrRoleNotFound {
		t.Errorf("Expected ErrRoleNotFound, got %v", err)
	}
}

func TestService_Create_ValidationError(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	// Empty name should fail validation
	_, err := svc.Create(context.Background(), "", "Description", 50, nil, nil)
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
}

func TestService_Create_InvalidHierarchyLevel(t *testing.T) {
	svc := NewService(newMockRoleDB(), &mockPermissionDB{})

	// Hierarchy level < 1 should fail
	_, err := svc.Create(context.Background(), "Test", "Description", 0, nil, nil)
	if err == nil {
		t.Error("Expected validation error for invalid hierarchy level")
	}
}
