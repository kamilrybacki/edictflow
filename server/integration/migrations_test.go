//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestMigrations_AllTablesExist(t *testing.T) {
	ctx := context.Background()

	expectedTables := []string{
		"teams",
		"users",
		"rules",
		"projects",
		"agents",
	}

	for _, tableName := range expectedTables {
		t.Run(tableName, func(t *testing.T) {
			var exists bool
			err := testPool.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)
			`, tableName).Scan(&exists)

			if err != nil {
				t.Fatalf("Failed to check table existence: %v", err)
			}

			if !exists {
				t.Errorf("Table %s does not exist", tableName)
			}
		})
	}
}

func TestMigrations_TeamsTableStructure(t *testing.T) {
	ctx := context.Background()

	expectedColumns := map[string]string{
		"id":         "uuid",
		"name":       "character varying",
		"settings":   "jsonb",
		"created_at": "timestamp with time zone",
	}

	for column, expectedType := range expectedColumns {
		t.Run(column, func(t *testing.T) {
			var dataType string
			err := testPool.QueryRow(ctx, `
				SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'teams' AND column_name = $1
			`, column).Scan(&dataType)

			if err != nil {
				t.Fatalf("Failed to get column info for %s: %v", column, err)
			}

			if dataType != expectedType {
				t.Errorf("Column %s: expected type %s, got %s", column, expectedType, dataType)
			}
		})
	}
}

func TestMigrations_RulesTableStructure(t *testing.T) {
	ctx := context.Background()

	expectedColumns := map[string]string{
		"id":              "uuid",
		"name":            "character varying",
		"content":         "text",
		"target_layer":    "character varying",
		"priority_weight": "integer",
		"triggers":        "jsonb",
		"team_id":         "uuid",
		"created_at":      "timestamp with time zone",
		"updated_at":      "timestamp with time zone",
	}

	for column, expectedType := range expectedColumns {
		t.Run(column, func(t *testing.T) {
			var dataType string
			err := testPool.QueryRow(ctx, `
				SELECT data_type
				FROM information_schema.columns
				WHERE table_name = 'rules' AND column_name = $1
			`, column).Scan(&dataType)

			if err != nil {
				t.Fatalf("Failed to get column info for %s: %v", column, err)
			}

			if dataType != expectedType {
				t.Errorf("Column %s: expected type %s, got %s", column, expectedType, dataType)
			}
		})
	}
}

func TestMigrations_IndexesExist(t *testing.T) {
	ctx := context.Background()

	expectedIndexes := []string{
		"idx_teams_name",
		"idx_rules_team_id",
		"idx_rules_target_layer",
	}

	for _, indexName := range expectedIndexes {
		t.Run(indexName, func(t *testing.T) {
			var exists bool
			err := testPool.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 FROM pg_indexes
					WHERE indexname = $1
				)
			`, indexName).Scan(&exists)

			if err != nil {
				t.Fatalf("Failed to check index existence: %v", err)
			}

			if !exists {
				t.Errorf("Index %s does not exist", indexName)
			}
		})
	}
}

func TestMigrations_ForeignKeyConstraints(t *testing.T) {
	ctx := context.Background()

	// Check that rules.team_id has a foreign key to teams.id
	var exists bool
	err := testPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.table_constraints tc
			JOIN information_schema.constraint_column_usage ccu
				ON tc.constraint_name = ccu.constraint_name
			WHERE tc.table_name = 'rules'
			AND tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_name = 'teams'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check foreign key: %v", err)
	}

	if !exists {
		t.Error("Foreign key from rules.team_id to teams.id does not exist")
	}
}
