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
	repo := teams.NewRepository(db)
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
	repo := teams.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if err != teams.ErrTeamNotFound {
		t.Errorf("expected ErrTeamNotFound, got %v", err)
	}
}
