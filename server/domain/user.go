package domain

import (
	"errors"
	"regexp"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthProvider string

const (
	AuthProviderGitHub AuthProvider = "github"
	AuthProviderGitLab AuthProvider = "gitlab"
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderLocal  AuthProvider = "local"
)

type User struct {
	ID            string        `json:"id"`
	Email         string        `json:"email"`
	Name          string        `json:"name"`
	PasswordHash  string        `json:"-"`
	AvatarURL     string        `json:"avatar_url,omitempty"`
	AuthProvider  AuthProvider  `json:"auth_provider"`
	TeamID        *string       `json:"team_id,omitempty"`
	CreatedBy     *string       `json:"created_by,omitempty"`
	EmailVerified bool          `json:"email_verified"`
	IsActive      bool          `json:"is_active"`
	LastLoginAt   *time.Time    `json:"last_login_at,omitempty"`
	Roles         []RoleEntity  `json:"roles,omitempty"`
	Permissions   []string      `json:"permissions,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewUser(email, name string, authProvider AuthProvider, teamID string) User {
	tid := teamID
	return User{
		ID:            uuid.New().String(),
		Email:         email,
		Name:          name,
		AuthProvider:  authProvider,
		TeamID:        &tid,
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Now(),
	}
}

func NewUserWithPassword(email, name string, teamID string, createdBy *string) User {
	var tid *string
	if teamID != "" {
		tid = &teamID
	}
	return User{
		ID:            uuid.New().String(),
		Email:         email,
		Name:          name,
		AuthProvider:  AuthProviderLocal,
		TeamID:        tid,
		CreatedBy:     createdBy,
		EmailVerified: true,
		IsActive:      true,
		CreatedAt:     time.Now(),
	}
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

func (u *User) HasPermission(permissionCode string) bool {
	for _, p := range u.Permissions {
		if p == permissionCode {
			return true
		}
	}
	return false
}

func (u User) Validate() error {
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}
	if u.Name == "" {
		return errors.New("name cannot be empty")
	}
	if !u.AuthProvider.IsValid() {
		return errors.New("invalid auth provider")
	}
	return nil
}

func (ap AuthProvider) IsValid() bool {
	switch ap {
	case AuthProviderGitHub, AuthProviderGitLab, AuthProviderGoogle, AuthProviderLocal:
		return true
	}
	return false
}
