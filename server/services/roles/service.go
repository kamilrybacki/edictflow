package roles

import (
	"context"
	"errors"

	"github.com/kamilrybacki/claudeception/server/domain"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrCannotModifySystem = errors.New("cannot modify system role")
)

type RoleDB interface {
	Create(ctx context.Context, role domain.RoleEntity) error
	GetByID(ctx context.Context, id string) (domain.RoleEntity, error)
	List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error)
	Update(ctx context.Context, role domain.RoleEntity) error
	Delete(ctx context.Context, id string) error
	GetPermissions(ctx context.Context, roleID string) ([]domain.Permission, error)
	AddPermission(ctx context.Context, roleID, permissionID string) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error
	RemoveUserRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]domain.RoleEntity, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

type PermissionDB interface {
	List(ctx context.Context) ([]domain.Permission, error)
	GetByCode(ctx context.Context, code string) (domain.Permission, error)
}

type Service struct {
	roleDB       RoleDB
	permissionDB PermissionDB
}

func NewService(roleDB RoleDB, permissionDB PermissionDB) *Service {
	return &Service{
		roleDB:       roleDB,
		permissionDB: permissionDB,
	}
}

func (s *Service) Create(ctx context.Context, name, description string, hierarchyLevel int, parentRoleID, teamID *string) (domain.RoleEntity, error) {
	role := domain.NewRoleEntity(name, description, hierarchyLevel, parentRoleID, teamID)
	if err := role.Validate(); err != nil {
		return domain.RoleEntity{}, err
	}

	if err := s.roleDB.Create(ctx, role); err != nil {
		return domain.RoleEntity{}, err
	}

	return role, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.RoleEntity, error) {
	return s.roleDB.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, teamID *string) ([]domain.RoleEntity, error) {
	return s.roleDB.List(ctx, teamID)
}

func (s *Service) Update(ctx context.Context, role domain.RoleEntity) error {
	existing, err := s.roleDB.GetByID(ctx, role.ID)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		return ErrCannotModifySystem
	}

	if err := role.Validate(); err != nil {
		return err
	}

	return s.roleDB.Update(ctx, role)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	existing, err := s.roleDB.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		return ErrCannotModifySystem
	}

	return s.roleDB.Delete(ctx, id)
}

func (s *Service) GetPermissions(ctx context.Context, roleID string) ([]domain.Permission, error) {
	return s.roleDB.GetPermissions(ctx, roleID)
}

func (s *Service) AddPermission(ctx context.Context, roleID, permissionID string) error {
	_, err := s.roleDB.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	return s.roleDB.AddPermission(ctx, roleID, permissionID)
}

func (s *Service) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	_, err := s.roleDB.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	return s.roleDB.RemovePermission(ctx, roleID, permissionID)
}

func (s *Service) AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error {
	_, err := s.roleDB.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	return s.roleDB.AssignUserRole(ctx, userID, roleID, assignedBy)
}

func (s *Service) RemoveUserRole(ctx context.Context, userID, roleID string) error {
	return s.roleDB.RemoveUserRole(ctx, userID, roleID)
}

func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]domain.RoleEntity, error) {
	return s.roleDB.GetUserRoles(ctx, userID)
}

func (s *Service) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return s.roleDB.GetUserPermissions(ctx, userID)
}

func (s *Service) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return s.permissionDB.List(ctx)
}

func (s *Service) GetPermissionByCode(ctx context.Context, code string) (domain.Permission, error) {
	return s.permissionDB.GetByCode(ctx, code)
}

func (s *Service) GetRoleWithPermissions(ctx context.Context, id string) (domain.RoleEntity, error) {
	role, err := s.roleDB.GetByID(ctx, id)
	if err != nil {
		return domain.RoleEntity{}, err
	}

	permissions, err := s.roleDB.GetPermissions(ctx, id)
	if err != nil {
		return domain.RoleEntity{}, err
	}

	role.Permissions = permissions
	return role, nil
}
