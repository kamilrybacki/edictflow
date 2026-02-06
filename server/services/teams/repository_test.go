package teams_test

import (
	"context"
	"testing"

	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/teams"
)

type mockDB struct {
	teams map[string]domain.Team
}

func newMockDB() *mockDB {
	return &mockDB{teams: make(map[string]domain.Team)}
}

type mockInviteDB struct {
	invites map[string]domain.TeamInvite
}

func newMockInviteDB() *mockInviteDB {
	return &mockInviteDB{invites: make(map[string]domain.TeamInvite)}
}

func (m *mockInviteDB) Create(ctx context.Context, invite domain.TeamInvite) error {
	m.invites[invite.ID] = invite
	return nil
}

func (m *mockInviteDB) GetByCode(ctx context.Context, code string) (domain.TeamInvite, error) {
	for _, inv := range m.invites {
		if inv.Code == code {
			return inv, nil
		}
	}
	return domain.TeamInvite{}, teams.ErrTeamNotFound
}

func (m *mockInviteDB) GetByID(ctx context.Context, id string) (domain.TeamInvite, error) {
	inv, ok := m.invites[id]
	if !ok {
		return domain.TeamInvite{}, teams.ErrTeamNotFound
	}
	return inv, nil
}

func (m *mockInviteDB) ListByTeam(ctx context.Context, teamID string) ([]domain.TeamInvite, error) {
	var result []domain.TeamInvite
	for _, inv := range m.invites {
		if inv.TeamID == teamID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockInviteDB) Delete(ctx context.Context, id string) error {
	delete(m.invites, id)
	return nil
}

func (m *mockInviteDB) IncrementUseCountAtomic(ctx context.Context, code string) (domain.TeamInvite, error) {
	for id, inv := range m.invites {
		if inv.Code == code {
			inv.UseCount++
			m.invites[id] = inv
			return inv, nil
		}
	}
	return domain.TeamInvite{}, teams.ErrTeamNotFound
}

func (m *mockDB) CreateTeam(ctx context.Context, team domain.Team) error {
	m.teams[team.ID] = team
	return nil
}

func (m *mockDB) GetTeam(ctx context.Context, id string) (domain.Team, error) {
	team, ok := m.teams[id]
	if !ok {
		return domain.Team{}, teams.ErrTeamNotFound
	}
	return team, nil
}

func (m *mockDB) ListTeams(ctx context.Context) ([]domain.Team, error) {
	var result []domain.Team
	for _, t := range m.teams {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockDB) UpdateTeam(ctx context.Context, team domain.Team) error {
	m.teams[team.ID] = team
	return nil
}

func (m *mockDB) DeleteTeam(ctx context.Context, id string) error {
	delete(m.teams, id)
	return nil
}

func TestRepositoryCreateAndGet(t *testing.T) {
	db := newMockDB()
	inviteDB := newMockInviteDB()
	repo := teams.NewRepository(db, inviteDB)
	ctx := context.Background()

	team := domain.NewTeam("Engineering")
	err := repo.Create(ctx, team)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	got, err := repo.GetByID(ctx, team.ID)
	if err != nil {
		t.Fatalf("failed to get team: %v", err)
	}

	if got.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", got.Name)
	}
}

func TestRepositoryGetByIDReturnsErrorForMissing(t *testing.T) {
	db := newMockDB()
	inviteDB := newMockInviteDB()
	repo := teams.NewRepository(db, inviteDB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if err != teams.ErrTeamNotFound {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}
