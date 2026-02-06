// e2e/api_helpers_test.go
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// waitForFileContent polls a file until it contains the expected content or times out
func waitForFileContent(t *testing.T, path, expected string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	pollInterval := 200 * time.Millisecond

	for time.Now().Before(deadline) {
		content, err := os.ReadFile(path)
		if err == nil && strings.TrimSpace(string(content)) == strings.TrimSpace(expected) {
			return true
		}
		time.Sleep(pollInterval)
	}

	// Log final state for debugging
	content, err := os.ReadFile(path)
	if err != nil {
		t.Logf("File read error at timeout: %v", err)
	} else {
		t.Logf("File content at timeout: %s", string(content))
	}
	return false
}

// waitForFileChanged polls until file content differs from original
func waitForFileChanged(t *testing.T, path, original string, timeout time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	pollInterval := 200 * time.Millisecond

	for time.Now().Before(deadline) {
		content, err := os.ReadFile(path)
		if err == nil && strings.TrimSpace(string(content)) != strings.TrimSpace(original) {
			return true
		}
		time.Sleep(pollInterval)
	}
	return false
}

// updateRuleEnforcementMode updates the enforcement mode of a rule via the API
func updateRuleEnforcementMode(t *testing.T, s *E2ESuite, mode string) {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/rules/%s", s.serverHostURL, s.testRuleID)

	payload := map[string]interface{}{
		"enforcement_mode": mode,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.testToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to update rule enforcement mode: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	t.Logf("Updated rule enforcement mode to: %s", mode)
}

// waitForAgentSync waits for the agent to sync with the server
func waitForAgentSync(t *testing.T, container testcontainers.Container, timeout time.Duration) bool {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	var lastLogs string
	for time.Now().Before(deadline) {
		logs, err := container.Logs(ctx)
		if err == nil {
			logContent, _ := io.ReadAll(logs)
			logs.Close()
			lastLogs = string(logContent)
			if strings.Contains(lastLogs, "Updated rules to version") ||
				strings.Contains(lastLogs, "Connected to server") {
				t.Logf("Agent synced successfully")
				return true
			}
		}
		time.Sleep(pollInterval)
	}

	// Log the agent output for debugging
	t.Logf("Agent logs at timeout:\n%s", lastLogs)
	return false
}

// ChangeEvent represents a change event from the API
type ChangeEvent struct {
	ID           string `json:"id"`
	RuleID       string `json:"rule_id"`
	FilePath     string `json:"file_path"`
	EventType    string `json:"event_type"`
	OriginalHash string `json:"original_hash"`
	ModifiedHash string `json:"modified_hash"`
	CreatedAt    string `json:"created_at"`
}

// getLatestChangeEvent retrieves the most recent change event from the server
func getLatestChangeEvent(t *testing.T, s *E2ESuite) (*ChangeEvent, error) {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/changes?team_id=%s&limit=1", s.serverHostURL, s.testTeamID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.testToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get changes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get changes: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var events []ChangeEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(events) == 0 {
		return nil, nil
	}

	return &events[0], nil
}

// modifyFile writes new content to a file
func modifyFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to modify file %s: %v", path, err)
	}
	t.Logf("Modified file: %s", path)
}

// getFileContent reads and returns file content
func getFileContent(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}
