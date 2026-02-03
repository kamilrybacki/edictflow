package main

import (
	"context"

	"github.com/kamilrybacki/claudeception/server/adapters/postgres"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
)

// teamServiceImpl implements handlers.TeamService
type teamServiceImpl struct {
	db *postgres.TeamDB
}

var _ handlers.TeamService = (*teamServiceImpl)(nil)

func (s *teamServiceImpl) Create(ctx context.Context, name string) (domain.Team, error) {
	team := domain.NewTeam(name)
	if err := team.Validate(); err != nil {
		return domain.Team{}, err
	}
	if err := s.db.CreateTeam(ctx, team); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}

func (s *teamServiceImpl) GetByID(ctx context.Context, id string) (domain.Team, error) {
	return s.db.GetTeam(ctx, id)
}

func (s *teamServiceImpl) List(ctx context.Context) ([]domain.Team, error) {
	return s.db.ListTeams(ctx)
}

func (s *teamServiceImpl) Update(ctx context.Context, team domain.Team) error {
	return s.db.UpdateTeam(ctx, team)
}

func (s *teamServiceImpl) Delete(ctx context.Context, id string) error {
	return s.db.DeleteTeam(ctx, id)
}

// ruleServiceImpl implements handlers.RuleService
type ruleServiceImpl struct {
	db *postgres.RuleDB
}

var _ handlers.RuleService = (*ruleServiceImpl)(nil)

func (s *ruleServiceImpl) Create(ctx context.Context, req handlers.CreateRuleRequest) (domain.Rule, error) {
	triggers := make([]domain.Trigger, len(req.Triggers))
	for i, t := range req.Triggers {
		triggers[i] = domain.Trigger{
			Type:         domain.TriggerType(t.Type),
			Pattern:      t.Pattern,
			ContextTypes: t.ContextTypes,
			Tags:         t.Tags,
		}
	}

	rule := domain.NewRule(
		req.Name,
		domain.TargetLayer(req.TargetLayer),
		req.Content,
		triggers,
		req.TeamID,
	)

	if err := rule.Validate(); err != nil {
		return domain.Rule{}, err
	}

	if err := s.db.CreateRule(ctx, rule); err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

func (s *ruleServiceImpl) GetByID(ctx context.Context, id string) (domain.Rule, error) {
	return s.db.GetRule(ctx, id)
}

func (s *ruleServiceImpl) ListByTeam(ctx context.Context, teamID string) ([]domain.Rule, error) {
	return s.db.ListRulesByTeam(ctx, teamID)
}

func (s *ruleServiceImpl) Update(ctx context.Context, rule domain.Rule) error {
	return s.db.UpdateRule(ctx, rule)
}

func (s *ruleServiceImpl) Delete(ctx context.Context, id string) error {
	return s.db.DeleteRule(ctx, id)
}

func (s *ruleServiceImpl) ListByStatus(ctx context.Context, teamID string, status domain.RuleStatus) ([]domain.Rule, error) {
	return s.db.ListByStatus(ctx, teamID, status)
}
