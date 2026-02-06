package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/domain"
	"github.com/kamilrybacki/edictflow/server/services/teams"
)

// TeamDB implements teams.DB interface with PostgreSQL
type TeamDB struct {
	pool *pgxpool.Pool
}

// NewTeamDB creates a new TeamDB instance
func NewTeamDB(pool *pgxpool.Pool) *TeamDB {
	return &TeamDB{pool: pool}
}

// CreateTeam inserts a new team into the database
func (db *TeamDB) CreateTeam(ctx context.Context, team domain.Team) error {
	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx, `
		INSERT INTO teams (id, name, settings, created_at)
		VALUES ($1, $2, $3, $4)
	`, team.ID, team.Name, settingsJSON, team.CreatedAt)
	return err
}

// GetTeam retrieves a team by ID
func (db *TeamDB) GetTeam(ctx context.Context, id string) (domain.Team, error) {
	var team domain.Team
	var settingsJSON []byte

	err := db.pool.QueryRow(ctx, `
		SELECT id, name, settings, created_at
		FROM teams
		WHERE id = $1
	`, id).Scan(&team.ID, &team.Name, &settingsJSON, &team.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, teams.ErrTeamNotFound
		}
		return domain.Team{}, err
	}

	if err := json.Unmarshal(settingsJSON, &team.Settings); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

// ListTeams retrieves all teams
func (db *TeamDB) ListTeams(ctx context.Context) ([]domain.Team, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, name, settings, created_at
		FROM teams
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamsList []domain.Team
	for rows.Next() {
		var team domain.Team
		var settingsJSON []byte

		if err := rows.Scan(&team.ID, &team.Name, &settingsJSON, &team.CreatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(settingsJSON, &team.Settings); err != nil {
			return nil, err
		}

		teamsList = append(teamsList, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teamsList, nil
}

// UpdateTeam updates an existing team
func (db *TeamDB) UpdateTeam(ctx context.Context, team domain.Team) error {
	settingsJSON, err := json.Marshal(team.Settings)
	if err != nil {
		return err
	}

	result, err := db.pool.Exec(ctx, `
		UPDATE teams
		SET name = $2, settings = $3
		WHERE id = $1
	`, team.ID, team.Name, settingsJSON)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return teams.ErrTeamNotFound
	}

	return nil
}

// DeleteTeam removes a team by ID
func (db *TeamDB) DeleteTeam(ctx context.Context, id string) error {
	result, err := db.pool.Exec(ctx, `
		DELETE FROM teams
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return teams.ErrTeamNotFound
	}

	return nil
}
