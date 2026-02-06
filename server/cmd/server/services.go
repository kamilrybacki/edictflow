package main

import (
	"context"
	"errors"

	"github.com/kamilrybacki/edictflow/server/adapters/postgres"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/entrypoints/api/handlers"
	"github.com/kamilrybacki/edictflow/server/services/merge"
)

var errInvalidPassword = errors.New("invalid password")

// teamServiceImpl implements handlers.TeamService and handlers.InviteService
type teamServiceImpl struct {
	db       *postgres.TeamDB
	inviteDB *postgres.TeamInviteDB
	userDB   *postgres.UserDB
}

var _ handlers.TeamService = (*teamServiceImpl)(nil)
var _ handlers.InviteService = (*teamServiceImpl)(nil)

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

// Invite methods for handlers.TeamService

func (s *teamServiceImpl) CreateInvite(ctx context.Context, teamID, createdBy string, maxUses, expiresInHours int) (domain.TeamInvite, error) {
	// Verify team exists
	if _, err := s.db.GetTeam(ctx, teamID); err != nil {
		return domain.TeamInvite{}, err
	}

	invite := domain.NewTeamInvite(teamID, createdBy, maxUses, expiresInHours)
	if err := s.inviteDB.Create(ctx, invite); err != nil {
		return domain.TeamInvite{}, err
	}
	return invite, nil
}

func (s *teamServiceImpl) ListInvites(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	return s.inviteDB.ListByTeam(ctx, teamID)
}

func (s *teamServiceImpl) DeleteInvite(ctx context.Context, teamID, inviteID string) error {
	// Verify invite belongs to team
	invite, err := s.inviteDB.GetByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.TeamID != teamID {
		return errors.New("invite not found")
	}
	return s.inviteDB.Delete(ctx, inviteID)
}

// JoinByCode implements handlers.InviteService
func (s *teamServiceImpl) JoinByCode(ctx context.Context, code, userID string) (domain.Team, error) {
	// Get user and check not already in team
	user, err := s.userDB.GetByID(ctx, userID)
	if err != nil {
		return domain.Team{}, err
	}
	if user.TeamID != nil {
		return domain.Team{}, errors.New("user already in a team")
	}

	// Use invite (atomic increment)
	invite, err := s.inviteDB.IncrementUseCountAtomic(ctx, code)
	if err != nil {
		return domain.Team{}, err
	}

	// Get team
	team, err := s.db.GetTeam(ctx, invite.TeamID)
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

func (s *ruleServiceImpl) ListGlobal(ctx context.Context) ([]domain.Rule, error) {
	return s.db.ListGlobalRules(ctx)
}

func (s *ruleServiceImpl) CreateGlobal(ctx context.Context, name, content string, description *string, force bool) (domain.Rule, error) {
	rule := domain.NewGlobalRule(name, content, force)
	rule.Description = description
	if err := rule.Validate(); err != nil {
		return domain.Rule{}, err
	}
	if err := s.db.CreateRule(ctx, rule); err != nil {
		return domain.Rule{}, err
	}
	return rule, nil
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
