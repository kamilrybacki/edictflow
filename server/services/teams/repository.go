package teams

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrTeamNotFound = errors.New("team not found")

type DB interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeam(ctx context.Context, id string) (domain.Team, error)
	ListTeams(ctx context.Context) ([]domain.Team, error)
	UpdateTeam(ctx context.Context, team domain.Team) error
	DeleteTeam(ctx context.Context, id string) error
}

type InviteDB interface {
	Create(ctx context.Context, invite domain.TeamInvite) error
	GetByCode(ctx context.Context, code string) (domain.TeamInvite, error)
	GetByID(ctx context.Context, id string) (domain.TeamInvite, error)
	ListByTeam(ctx context.Context, teamID string) ([]domain.TeamInvite, error)
	Delete(ctx context.Context, id string) error
	IncrementUseCountAtomic(ctx context.Context, code string) (domain.TeamInvite, error)
}

type Repository struct {
	db       DB
	inviteDB InviteDB
}

func NewRepository(db DB, inviteDB InviteDB) *Repository {
	return &Repository{db: db, inviteDB: inviteDB}
}

func (r *Repository) Create(ctx context.Context, team domain.Team) error {
	return r.db.CreateTeam(ctx, team)
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Team, error) {
	return r.db.GetTeam(ctx, id)
}

func (r *Repository) List(ctx context.Context) ([]domain.Team, error) {
	return r.db.ListTeams(ctx)
}

func (r *Repository) Update(ctx context.Context, team domain.Team) error {
	return r.db.UpdateTeam(ctx, team)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.DeleteTeam(ctx, id)
}

// Invite methods

func (r *Repository) CreateInvite(ctx context.Context, invite domain.TeamInvite) error {
	return r.inviteDB.Create(ctx, invite)
}

func (r *Repository) GetInviteByCode(ctx context.Context, code string) (domain.TeamInvite, error) {
	return r.inviteDB.GetByCode(ctx, code)
}

func (r *Repository) GetInviteByID(ctx context.Context, id string) (domain.TeamInvite, error) {
	return r.inviteDB.GetByID(ctx, id)
}

func (r *Repository) ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	return r.inviteDB.ListByTeam(ctx, teamID)
}

func (r *Repository) DeleteInvite(ctx context.Context, id string) error {
	return r.inviteDB.Delete(ctx, id)
}

func (r *Repository) UseInvite(ctx context.Context, code string) (domain.TeamInvite, error) {
	return r.inviteDB.IncrementUseCountAtomic(ctx, code)
}
