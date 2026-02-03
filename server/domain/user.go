package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	AuthProviderGitHub AuthProvider = "github"
	AuthProviderGitLab AuthProvider = "gitlab"
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderLocal  AuthProvider = "local"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	AvatarURL    string       `json:"avatar_url,omitempty"`
	AuthProvider AuthProvider `json:"auth_provider"`
	Role         Role         `json:"role"`
	TeamID       string       `json:"team_id"`
	CreatedAt    time.Time    `json:"created_at"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func NewUser(email, name string, authProvider AuthProvider, teamID string) User {
	return User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		AuthProvider: authProvider,
		Role:         RoleMember,
		TeamID:       teamID,
		CreatedAt:    time.Now(),
	}
}

func (u User) Validate() error {
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}
	if !u.AuthProvider.IsValid() {
		return errors.New("invalid auth provider")
	}
	if !u.Role.IsValid() {
		return errors.New("invalid role")
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

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleMember:
		return true
	}
	return false
}
