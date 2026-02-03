package teams

import (
	"context"
	"errors"

	"github.com/kamilrybacki/claudeception/server/domain"
)

var ErrTeamNotFound = errors.New("team not found")

type DB interface {
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeam(ctx context.Context, id string) (domain.Team, error)
	ListTeams(ctx context.Context) ([]domain.Team, error)
	UpdateTeam(ctx context.Context, team domain.Team) error
	DeleteTeam(ctx context.Context, id string) error
}

type Repository struct {
	db DB
}

func NewRepository(db DB) *Repository {
	return &Repository{db: db}
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
