package rules

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/domain"
)

var ErrRuleNotFound = errors.New("rule not found")

type DB interface {
	CreateRule(ctx context.Context, rule domain.Rule) error
	GetRule(ctx context.Context, id string) (domain.Rule, error)
	ListRulesByTeam(ctx context.Context, teamID string) ([]domain.Rule, error)
	ListAllRules(ctx context.Context) ([]domain.Rule, error)
	UpdateRule(ctx context.Context, rule domain.Rule) error
	DeleteRule(ctx context.Context, id string) error
}

type Repository struct {
	db DB
}

func NewRepository(db DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, rule domain.Rule) error {
	return r.db.CreateRule(ctx, rule)
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	return r.db.GetRule(ctx, id)
}

func (r *Repository) ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	return r.db.ListRulesByTeam(ctx, teamID)
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.Rule, error) {
	return r.db.ListAllRules(ctx)
}

func (r *Repository) Update(ctx context.Context, rule domain.Rule) error {
	return r.db.UpdateRule(ctx, rule)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.db.DeleteRule(ctx, id)
}
