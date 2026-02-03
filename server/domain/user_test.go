package domain_test

import (
	"testing"

	"github.com/kamilrybacki/claudeception/server/domain"
)

func TestNewUserCreatesValidUser(t *testing.T) {
	user := domain.NewUser("alice@example.com", "Alice", domain.AuthProviderGitHub, "team-123")

	if user.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", user.Email)
	}
	if user.AuthProvider != domain.AuthProviderGitHub {
		t.Errorf("expected auth provider 'github', got '%s'", user.AuthProvider)
	}
	if !user.IsActive {
		t.Error("expected user to be active")
	}
	if !user.EmailVerified {
		t.Error("expected email to be verified")
	}
}

func TestUserValidateRejectsInvalidEmail(t *testing.T) {
	user := domain.User{
		ID:           "test-id",
		Email:        "not-an-email",
		Name:         "Test",
		AuthProvider: domain.AuthProviderGitHub,
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
	}

	err := user.Validate()
	if err == nil {
		t.Error("expected validation error for invalid auth provider")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Password1", false},
		{"too short", "Pass1", true},
		{"no uppercase", "password1", true},
		{"no lowercase", "PASSWORD1", true},
		{"no number", "Password", true},
		{"complex valid", "MyP@ssw0rd!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := domain.NewUserWithPassword("test@example.com", "Test User", "team-123", nil)
	err := user.SetPassword("ValidPass1")
	if err != nil {
		t.Errorf("SetPassword() unexpected error: %v", err)
	}
	if user.PasswordHash == "" {
		t.Error("Expected password hash to be set")
	}
	if !user.CheckPassword("ValidPass1") {
		t.Error("CheckPassword() should return true for correct password")
	}
	if user.CheckPassword("WrongPass1") {
		t.Error("CheckPassword() should return false for incorrect password")
	}
}

func TestNewUserWithPassword(t *testing.T) {
	creatorID := "creator-123"
	user := domain.NewUserWithPassword("test@example.com", "Test User", "team-123", &creatorID)

	if user.AuthProvider != domain.AuthProviderLocal {
		t.Errorf("expected auth provider 'local', got '%s'", user.AuthProvider)
	}
	if user.CreatedBy == nil || *user.CreatedBy != creatorID {
		t.Error("expected CreatedBy to be set")
	}
	if !user.IsActive {
		t.Error("expected user to be active")
	}
}

func TestUser_HasPermission(t *testing.T) {
	user := domain.User{
		Permissions: []string{"create_rules", "edit_rules"},
	}

	if !user.HasPermission("create_rules") {
		t.Error("expected HasPermission to return true for 'create_rules'")
	}
	if user.HasPermission("delete_rules") {
		t.Error("expected HasPermission to return false for 'delete_rules'")
	}
}
