//go:build integration

package testhelpers

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations executes all up migrations from the migrations directory
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Collect and sort migration files
	var upMigrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	// Execute migrations in order
	for _, migration := range upMigrations {
		content, err := os.ReadFile(filepath.Join(migrationsDir, migration))
		if err != nil {
			return err
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return err
		}
	}

	return nil
}

// ResetDatabase truncates all tables and restarts sequences
func ResetDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	tables := []string{
		"role_permissions",
		"user_roles",
		"rules",
		"users",
		"projects",
		"agents",
		"teams",
		"categories",
	}

	// Delete non-seeded permissions (keep ones from migrations)
	_, _ = pool.Exec(ctx, "DELETE FROM permissions WHERE id NOT LIKE 'a0000001-%'")

	for _, table := range tables {
		if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE"); err != nil {
			// Table might not exist yet, ignore errors
			continue
		}
	}

	return nil
}

// DropAllTables drops all tables (used for migration testing)
func DropAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	return err
}
