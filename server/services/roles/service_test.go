package roles

import (
	"context"
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
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
