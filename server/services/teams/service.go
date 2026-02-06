package teams

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var (
	ErrUserAlreadyInTeam = errors.New("user already in a team")
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite expired or max uses reached")
)

type UserDB interface {
	GetByID(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
}

type Service struct {
	repo   *Repository
	userDB UserDB
}

func NewService(repo *Repository, userDB UserDB) *Service {
	return &Service{repo: repo, userDB: userDB}
}

func (s *Service) Create(ctx context.Context, name string) (domain.Team, error) {
	team := domain.NewTeam(name)
	if err := team.Validate(); err != nil {
		return domain.Team{}, err
	}
	if err := s.repo.Create(ctx, team); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (domain.Team, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]domain.Team, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, team domain.Team) error {
	return s.repo.Update(ctx, team)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Invite methods

func (s *Service) CreateInvite(ctx context.Context, teamID, createdBy string, maxUses, expiresInHours int) (domain.TeamInvite, error) {
	// Verify team exists
	if _, err := s.repo.GetByID(ctx, teamID); err != nil {
		return domain.TeamInvite{}, err
	}

	invite := domain.NewTeamInvite(teamID, createdBy, maxUses, expiresInHours)
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return domain.TeamInvite{}, err
	}
	return invite, nil
}

func (s *Service) ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	return s.repo.ListInvites(ctx, teamID)
}

func (s *Service) DeleteInvite(ctx context.Context, teamID, inviteID string) error {
	// Verify invite belongs to team
	invite, err := s.repo.GetInviteByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.TeamID != teamID {
		return ErrInviteNotFound
	}
	return s.repo.DeleteInvite(ctx, inviteID)
}

func (s *Service) JoinByCode(ctx context.Context, code, userID string) (domain.Team, error) {
	// Get user and check not already in team
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return domain.Team{}, err
	}
	if user.TeamID != nil {
		return domain.Team{}, ErrUserAlreadyInTeam
	}

	// Use invite (atomic increment)
	invite, err := s.repo.UseInvite(ctx, code)
	if err != nil {
		return domain.Team{}, err
	}

	// Get team
	team, err := s.repo.GetByID(ctx, invite.TeamID)
	if err != nil {
		return domain.Team{}, err
	}

	// Update user's team
	user.TeamID = &team.ID
	if err := s.userDB.Update(ctx, user); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}
