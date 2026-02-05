package main

import (
	"context"
	"errors"

	"github.com/kamilrybacki/claudeception/server/adapters/postgres"
	"github.com/kamilrybacki/claudeception/server/domain"
	"github.com/kamilrybacki/claudeception/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/claudeception/server/services/merge"
)

var errInvalidPassword = errors.New("invalid password")

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
	db         *postgres.RuleDB
	categoryDB *postgres.CategoryDB
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

func (s *ruleServiceImpl) ListByTargetLayer(ctx context.Context, targetLayer domain.TargetLayer) ([]domain.Rule, error) {
	return s.db.ListByTargetLayer(ctx, targetLayer)
}

func (s *ruleServiceImpl) GetMergedContent(ctx context.Context, targetLayer domain.TargetLayer) (string, error) {
	rules, err := s.db.ListByTargetLayer(ctx, targetLayer)
	if err != nil {
		return "", err
	}

	categories, err := s.categoryDB.ListAll(ctx)
	if err != nil {
		return "", err
	}

	mergeSvc := merge.NewService()
	return mergeSvc.RenderManagedSection(rules, categories), nil
}

// categoryServiceImpl implements handlers.CategoryService
type categoryServiceImpl struct {
	db *postgres.CategoryDB
}

var _ handlers.CategoryService = (*categoryServiceImpl)(nil)

func (s *categoryServiceImpl) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	return s.db.Create(ctx, category)
}

func (s *categoryServiceImpl) GetByID(ctx context.Context, id string) (domain.Category, error) {
	return s.db.GetByID(ctx, id)
}

func (s *categoryServiceImpl) List(ctx context.Context, orgID *string) ([]domain.Category, error) {
	return s.db.List(ctx, orgID)
}

func (s *categoryServiceImpl) Update(ctx context.Context, category domain.Category) error {
	return s.db.Update(ctx, category)
}

func (s *categoryServiceImpl) Delete(ctx context.Context, id string) error {
	return s.db.Delete(ctx, id)
}

// userServiceImpl implements handlers.UserService
type userServiceImpl struct {
	db *postgres.UserDB
}

var _ handlers.UserService = (*userServiceImpl)(nil)

func (s *userServiceImpl) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.db.GetByID(ctx, id)
}

func (s *userServiceImpl) Update(ctx context.Context, user domain.User) error {
	return s.db.Update(ctx, user)
}

func (s *userServiceImpl) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.db.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.CheckPassword(oldPassword) {
		return errInvalidPassword
	}

	if err := user.SetPassword(newPassword); err != nil {
		return err
	}

	return s.db.UpdatePassword(ctx, userID, user.PasswordHash)
}
