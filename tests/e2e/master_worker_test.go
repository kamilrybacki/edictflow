package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestMasterWorkerE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start docker-compose
	comp, err := compose.NewDockerCompose("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("failed to create compose: %v", err)
	}

	t.Cleanup(func() {
		if err := comp.Down(ctx, testcontainers.RemoveOrphans(true)); err != nil {
			t.Logf("failed to stop compose: %v", err)
		}
	})

	err = comp.Up(ctx, compose.Wait(true))
	if err != nil {
		t.Fatalf("failed to start compose: %v", err)
	}

	// Wait for services to be ready
	time.Sleep(10 * time.Second)

	// Test master health
	masterURL := "http://localhost:8080/health"
	resp, err := http.Get(masterURL)
	if err != nil {
		t.Fatalf("master health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("master health check returned %d", resp.StatusCode)
	}

	// Test worker health
	workerURL := "http://localhost:8081/health"
	resp, err = http.Get(workerURL)
	if err != nil {
		t.Fatalf("worker health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("worker health check returned %d", resp.StatusCode)
	}

	// Test master API root endpoint
	apiURL := "http://localhost:8080/"
	resp, err = http.Get(apiURL)
	if err != nil {
		t.Fatalf("API root check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("API root returned %d", resp.StatusCode)
	}
}
