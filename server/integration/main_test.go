//go:build integration

package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamilrybacki/edictflow/server/integration/testhelpers"
)

var (
	testPool      *pgxpool.Pool
	testContainer *testhelpers.PostgresContainer
	testFixtures  *testhelpers.Fixtures
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container
	container, err := testhelpers.StartPostgres(ctx)
	if err != nil {
		log.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	testContainer = container

	// Create connection pool
	pool, err := container.NewPool(ctx)
	if err != nil {
		container.Stop(ctx)
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	testPool = pool

	// Find migrations directory
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		pool.Close()
		container.Stop(ctx)
		log.Fatalf("Failed to find migrations directory")
	}

	// Run migrations
	if err := testhelpers.RunMigrations(ctx, pool, migrationsDir); err != nil {
		pool.Close()
		container.Stop(ctx)
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create fixtures helper
	testFixtures = testhelpers.NewFixtures(pool)

	// Run tests
	code := m.Run()

	// Cleanup
	pool.Close()
	container.Stop(ctx)

	os.Exit(code)
}

// findMigrationsDir locates the migrations directory
func findMigrationsDir() string {
	// Try relative paths from the integration test directory
	candidates := []string{
		"../migrations",
		"./migrations",
		"../../migrations",
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for _, candidate := range candidates {
		path := filepath.Join(cwd, candidate)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}

	return ""
}

// resetDB clears all data between tests
func resetDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if err := testhelpers.ResetDatabase(ctx, testPool); err != nil {
		t.Fatalf("Failed to reset database: %v", err)
	}
}
