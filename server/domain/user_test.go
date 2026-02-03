package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewUserCreatesValidUser(t *testing.T) {
	user := domain.NewUser("alice@example.com", "Alice", "github", "team-123")

	if user.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", user.Email)
	}
	if user.AuthProvider != domain.AuthProviderGitHub {
		t.Errorf("expected auth provider 'github', got '%s'", user.AuthProvider)
	}
	if user.Role != domain.RoleMember {
		t.Errorf("expected role 'member', got '%s'", user.Role)
	}
}

func TestUserValidateRejectsInvalidEmail(t *testing.T) {
	user := domain.User{
		ID:           "test-id",
		Email:        "not-an-email",
		Name:         "Test",
		AuthProvider: domain.AuthProviderGitHub,
		Role:         domain.RoleMember,
		TeamID:       "team-123",
	}

	err := user.Validate()
	if err == nil {
		t.Error("expected validation error for invalid email")
	}
}

func TestUserValidateRejectsInvalidAuthProvider(t *testing.T) {
	user := domain.User{
		ID:           "test-id",
		Email:        "alice@example.com",
		Name:         "Alice",
		AuthProvider: "invalid",
		Role:         domain.RoleMember,
		TeamID:       "team-123",
	}

	err := user.Validate()
	if err == nil {
		t.Error("expected validation error for invalid auth provider")
	}
}
