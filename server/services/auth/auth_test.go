package auth

import (
	"context"
	"testing"
	"time"

	"github.com/kamilrybacki/edictflow/server/domain"
)

type mockUserDB struct {
	users map[string]domain.User
}

func (m *mockUserDB) Create(ctx context.Context, user domain.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserDB) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return domain.User{}, ErrInvalidCredentials
}

func (m *mockUserDB) UpdateLastLogin(ctx context.Context, userID string) error {
	return nil
}

type mockRoleDB struct{}

func (m *mockRoleDB) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return []string{"create_rules"}, nil
}

func (m *mockRoleDB) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	return nil
}

func TestService_Register(t *testing.T) {
	svc := NewService(&mockUserDB{users: make(map[string]domain.User)}, &mockRoleDB{}, "test-secret", 24*time.Hour)

	token, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass1",
	})

	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if token == "" {
		t.Error("Expected token to be returned")
	}
}

func TestService_Register_WeakPassword(t *testing.T) {
	svc := NewService(&mockUserDB{users: make(map[string]domain.User)}, &mockRoleDB{}, "test-secret", 24*time.Hour)

	_, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "weak",
	})

	if err == nil {
		t.Error("Expected error for weak password")
	}
}

func TestService_Login(t *testing.T) {
	userDB := &mockUserDB{users: make(map[string]domain.User)}
	svc := NewService(userDB, &mockRoleDB{}, "test-secret", 24*time.Hour)

	// First register a user
	_, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass1",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Then login
	token, _, err := svc.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "ValidPass1",
	})

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Error("Expected token to be returned")
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	userDB := &mockUserDB{users: make(map[string]domain.User)}
	svc := NewService(userDB, &mockRoleDB{}, "test-secret", 24*time.Hour)

	// First register a user
	_, _, _ = svc.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass1",
	})

	// Then try wrong password
	_, _, err := svc.Login(context.Background(), LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPass1",
	})

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_ValidateToken(t *testing.T) {
	userDB := &mockUserDB{users: make(map[string]domain.User)}
	svc := NewService(userDB, &mockRoleDB{}, "test-secret", 24*time.Hour)

	token, _, _ := svc.Register(context.Background(), RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass1",
	})

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", claims.Email)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "create_rules" {
		t.Errorf("Expected permissions ['create_rules'], got %v", claims.Permissions)
	}
}
