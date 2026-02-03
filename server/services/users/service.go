package users

import (
	"context"
	"errors"

	"github.com/kamilrybacki/claudeception/server/domain"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

type UserDB interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error)
	Deactivate(ctx context.Context, id string) error
}

type RoleDB interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserRoles(ctx context.Context, userID string) ([]domain.RoleEntity, error)
}

type Service struct {
	userDB UserDB
	roleDB RoleDB
}

func NewService(userDB UserDB, roleDB RoleDB) *Service {
	return &Service{
		userDB: userDB,
		roleDB: roleDB,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.userDB.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.userDB.GetByEmail(ctx, email)
}

func (s *Service) Update(ctx context.Context, user domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	return s.userDB.Update(ctx, user)
}

func (s *Service) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if !user.CheckPassword(oldPassword) {
		return ErrInvalidPassword
	}

	if err := domain.ValidatePassword(newPassword); err != nil {
		return err
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	return s.userDB.UpdatePassword(ctx, userID, user.PasswordHash)
}

func (s *Service) List(ctx context.Context, teamID *string, activeOnly bool) ([]domain.User, error) {
	return s.userDB.List(ctx, teamID, activeOnly)
}

func (s *Service) Deactivate(ctx context.Context, id string) error {
	return s.userDB.Deactivate(ctx, id)
}

func (s *Service) GetWithRolesAndPermissions(ctx context.Context, id string) (domain.User, error) {
	user, err := s.userDB.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	permissions, err := s.roleDB.GetUserPermissions(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user.Permissions = permissions

	roles, err := s.roleDB.GetUserRoles(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user.Roles = roles

	return user, nil
}
