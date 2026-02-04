package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kamilrybacki/claudeception/server/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
)

type UserDB interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
}

type RoleDB interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	AssignUserRole(ctx context.Context, userID, roleID string, assignedBy *string) error
}

type Service struct {
	userDB      UserDB
	roleDB      RoleDB
	jwtSecret   string
	tokenExpiry time.Duration
}

func NewService(userDB UserDB, roleDB RoleDB, jwtSecret string, tokenExpiry time.Duration) *Service {
	return &Service{
		userDB:      userDB,
		roleDB:      roleDB,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

type RegisterRequest struct {
	Email    string
	Name     string
	Password string
	TeamID   string
}

type LoginRequest struct {
	Email    string
	Password string
}

type Claims struct {
	jwt.RegisteredClaims
	Email       string   `json:"email"`
	TeamID      *string  `json:"team_id,omitempty"`
	Permissions []string `json:"permissions"`
}

const DefaultMemberRoleID = "b0000001-0000-0000-0000-000000000001"

func (s *Service) Register(ctx context.Context, req RegisterRequest) (string, error) {
	if err := domain.ValidatePassword(req.Password); err != nil {
		return "", err
	}

	user := domain.NewUserWithPassword(req.Email, req.Name, req.TeamID, nil)
	if err := user.SetPassword(req.Password); err != nil {
		return "", err
	}

	if err := user.Validate(); err != nil {
		return "", err
	}

	if err := s.userDB.Create(ctx, user); err != nil {
		return "", err
	}

	// Assign default Member role
	if err := s.roleDB.AssignUserRole(ctx, user.ID, DefaultMemberRoleID, nil); err != nil {
		return "", err
	}

	return s.generateToken(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	user, err := s.userDB.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", ErrInvalidCredentials
	}

	if !user.CheckPassword(req.Password) {
		return "", ErrInvalidCredentials
	}

	if err := s.userDB.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", err
	}

	return s.generateToken(ctx, user)
}

func (s *Service) generateToken(ctx context.Context, user domain.User) (string, error) {
	permissions, err := s.roleDB.GetUserPermissions(ctx, user.ID)
	if err != nil {
		return "", err
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email:       user.Email,
		TeamID:      user.TeamID,
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken generates a JWT token for the given userID.
// This is used by device auth flow when a device code is authorized.
func (s *Service) GenerateToken(userID string) (string, error) {
	ctx := context.Background()
	permissions, err := s.roleDB.GetUserPermissions(ctx, userID)
	if err != nil {
		return "", err
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Permissions: permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
