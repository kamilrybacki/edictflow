package users

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
)

type mockUserDB struct {
	users map[string]domain.User
}

func (m *mockUserDB) GetByID(ctx context.Context, id string) (domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return domain.User{}, ErrUserNotFound
}

func (m *mockUserDB) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, ErrUserNotFound
}

func (m *mockUserDB) Update(ctx context.Context, user domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return ErrUserNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserDB) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if user, ok := m.users[userID]; ok {
		user.PasswordHash = passwordHash
		m.users[userID] = user
		return nil
	}
	return ErrUserNotFound
}

func (m *mockUserDB) List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
	var result []domain.User
	for _, user := range m.users {
		if activeOnly && !user.IsActive {
			continue
		}
		if teamID != nil && (user.TeamID == nil || *user.TeamID != *teamID) {
			continue
		}
		result = append(result, user)
	}
	return result, nil
}

func (m *mockUserDB) Deactivate(ctx context.Context, id string) error {
	if user, ok := m.users[id]; ok {
		user.IsActive = false
		m.users[id] = user
		return nil
	}
	return ErrUserNotFound
}

func (m *mockUserDB) GetByIDs(ctx context.Context, ids []string) (map[string]domain.User, error) {
	result := make(map[string]domain.User)
	for _, id := range ids {
		if user, ok := m.users[id]; ok {
			result[id] = user
		}
	}
	return result, nil
}

type mockRoleDB struct {
	userPermissions map[string][]string
	userRoles       map[string][]domain.Role
}

func (m *mockRoleDB) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	if perms, ok := m.userPermissions[userID]; ok {
		return perms, nil
	}
	return []string{}, nil
}

func (m *mockRoleDB) GetUserRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	if roles, ok := m.userRoles[userID]; ok {
		return roles, nil
	}
	return []domain.Role{}, nil
}

func TestService_GetByID(t *testing.T) {
	teamID := "team-123"
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", Name: "Test User", TeamID: &teamID, IsActive: true},
		},
	}
	roleDB := &mockRoleDB{
		userPermissions: map[string][]string{"user-1": {"create_rules"}},
		userRoles:       map[string][]domain.Role{"user-1": {{ID: "role-1", Name: "Member"}}},
	}

	svc := NewService(userDB, roleDB)

	user, err := svc.GetByID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewService(&mockUserDB{users: map[string]domain.User{}}, &mockRoleDB{})

	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestService_Update(t *testing.T) {
	teamID := "team-123"
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", Name: "Old Name", TeamID: &teamID, IsActive: true, AuthProvider: domain.AuthProviderLocal},
		},
	}
	svc := NewService(userDB, &mockRoleDB{})

	user := userDB.users["user-1"]
	user.Name = "New Name"

	err := svc.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated := userDB.users["user-1"]
	if updated.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", updated.Name)
	}
}

func TestService_UpdatePassword(t *testing.T) {
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", IsActive: true},
		},
	}
	svc := NewService(userDB, &mockRoleDB{})

	// Set initial password
	user := userDB.users["user-1"]
	_ = user.SetPassword("OldPass123")
	userDB.users["user-1"] = user

	err := svc.UpdatePassword(context.Background(), "user-1", "OldPass123", "NewPass456")
	if err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}
}

func TestService_UpdatePassword_WrongOld(t *testing.T) {
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", IsActive: true},
		},
	}
	svc := NewService(userDB, &mockRoleDB{})

	user := userDB.users["user-1"]
	_ = user.SetPassword("OldPass123")
	userDB.users["user-1"] = user

	err := svc.UpdatePassword(context.Background(), "user-1", "WrongPass1", "NewPass456")
	if err != ErrInvalidPassword {
		t.Errorf("Expected ErrInvalidPassword, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	teamID := "team-123"
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "a@example.com", TeamID: &teamID, IsActive: true},
			"user-2": {ID: "user-2", Email: "b@example.com", TeamID: &teamID, IsActive: false},
		},
	}
	svc := NewService(userDB, &mockRoleDB{})

	// List active only
	users, err := svc.List(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 active user, got %d", len(users))
	}
}

func TestService_Deactivate(t *testing.T) {
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", IsActive: true},
		},
	}
	svc := NewService(userDB, &mockRoleDB{})

	err := svc.Deactivate(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}

	if userDB.users["user-1"].IsActive {
		t.Error("Expected user to be deactivated")
	}
}

func TestService_GetWithRolesAndPermissions(t *testing.T) {
	teamID := "team-123"
	userDB := &mockUserDB{
		users: map[string]domain.User{
			"user-1": {ID: "user-1", Email: "test@example.com", TeamID: &teamID, IsActive: true},
		},
	}
	roleDB := &mockRoleDB{
		userPermissions: map[string][]string{"user-1": {"create_rules", "edit_rules"}},
		userRoles:       map[string][]domain.Role{"user-1": {{ID: "role-1", Name: "Member", HierarchyLevel: 1}}},
	}
	svc := NewService(userDB, roleDB)

	user, err := svc.GetWithRolesAndPermissions(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetWithRolesAndPermissions() error = %v", err)
	}
	if len(user.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(user.Permissions))
	}
	if len(user.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(user.Roles))
	}
}
