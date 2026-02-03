//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestDatabaseConnectivity(t *testing.T) {
	ctx := context.Background()

	// Test basic connectivity
	var result int
	err := testPool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("Failed to execute SELECT 1: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}
}

func TestDatabasePoolStats(t *testing.T) {
	stats := testPool.Stat()

	if stats.TotalConns() == 0 {
		t.Error("Expected at least one connection in the pool")
	}
}
