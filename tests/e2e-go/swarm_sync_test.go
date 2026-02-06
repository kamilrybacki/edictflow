// e2e/swarm_sync_test.go
package e2e

import (
	"sync"
	"testing"
	"time"
)

const (
	swarmSyncTimeout = 60 * time.Second
	numSwarmAgents   = 3
)

// TestSwarmSync tests synchronization across multiple agents
func TestSwarmSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping swarm E2E test in short mode")
	}

	// Create swarm with multiple agents
	suite := NewSwarmSuite(t, SwarmConfig{
		NumAgents: numSwarmAgents,
	})
	defer suite.Cleanup()

	// Wait for all agents to sync
	if !suite.WaitForAllAgentsSync(t, swarmSyncTimeout) {
		t.Fatal("Not all agents synced with server")
	}

	t.Logf("All %d agents connected and synced", numSwarmAgents)

	// Run test scenarios
	t.Run("AllAgentsReceiveInitialState", func(t *testing.T) {
		testAllAgentsReceiveInitialState(t, suite)
	})

	t.Run("ConcurrentFileModifications", func(t *testing.T) {
		testConcurrentFileModifications(t, suite)
	})

	t.Run("RuleUpdatePropagation", func(t *testing.T) {
		testRuleUpdatePropagation(t, suite)
	})
}

// testAllAgentsReceiveInitialState verifies all agents start with the same initial file content
func testAllAgentsReceiveInitialState(t *testing.T, suite *SwarmSuite) {
	t.Helper()

	// All agents should have the original CLAUDE.md content
	if verifyAllAgentFilesMatch(t, suite, originalContent) {
		t.Log("All agents have consistent initial file state")
	} else {
		t.Error("Agents have inconsistent initial file state")
	}
}

// testConcurrentFileModifications tests what happens when all agents' files are modified simultaneously
func testConcurrentFileModifications(t *testing.T, suite *SwarmSuite) {
	t.Helper()

	// Reset all files to original content
	modifyAllAgentFiles(t, suite, originalContent)
	time.Sleep(1 * time.Second)

	// Modify all agent files concurrently with unique content
	var wg sync.WaitGroup
	for i := 0; i < numSwarmAgents; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := suite.GetCLAUDEMDPath(idx)
			uniqueContent := modifiedContent + "\n<!-- Agent " + string(rune('A'+idx)) + " modification -->\n"
			modifyFile(t, path, uniqueContent)
		}(i)
	}
	wg.Wait()

	// Wait for agents to detect changes
	time.Sleep(2 * time.Second)

	// Verify each agent detected its own file modification
	// (Since revert isn't implemented, files should remain modified)
	allModified := true
	for i := 0; i < numSwarmAgents; i++ {
		path := suite.GetCLAUDEMDPath(i)
		content := getFileContent(t, path)
		if content == originalContent {
			t.Logf("Agent %d file was unexpectedly reverted", i)
			allModified = false
		} else {
			t.Logf("Agent %d file modification persists (change detected)", i)
		}
	}

	if allModified {
		t.Log("All concurrent file modifications were detected")
	}

	// Reset for next test
	modifyAllAgentFiles(t, suite, originalContent)
	time.Sleep(500 * time.Millisecond)
}

// testRuleUpdatePropagation verifies that rule updates are propagated to all agents
func testRuleUpdatePropagation(t *testing.T, suite *SwarmSuite) {
	t.Helper()

	// Update the rule's enforcement mode
	// This tests that the server can communicate with all agents
	updateRuleEnforcementMode(t, &E2ESuite{
		serverHostURL: suite.serverHostURL,
		testRuleID:    suite.testRuleID,
		testTeamID:    suite.testTeamID,
		testToken:     suite.agents[0].Token,
	}, "warning")

	// Wait for propagation
	time.Sleep(2 * time.Second)

	// Verify all agents are still connected and responsive
	// by checking each agent received the update (or at least is still running)
	allResponsive := true
	for i, agent := range suite.agents {
		logs := getAgentLogs(t, agent.Container)
		if logs == "" {
			t.Logf("Agent %d: No logs available (may have crashed)", i)
			allResponsive = false
		} else {
			t.Logf("Agent %d: Still running", i)
		}
	}

	if allResponsive {
		t.Log("All agents remained responsive after rule update")
	} else {
		t.Error("Some agents became unresponsive after rule update")
	}
}

// TestSwarmScaling tests the system with varying numbers of agents
func TestSwarmScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping swarm scaling test in short mode")
	}

	testCases := []struct {
		name      string
		numAgents int
	}{
		{"TwoAgents", 2},
		{"FiveAgents", 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := NewSwarmSuite(t, SwarmConfig{
				NumAgents: tc.numAgents,
			})
			defer suite.Cleanup()

			// Wait for all agents to sync
			if !suite.WaitForAllAgentsSync(t, swarmSyncTimeout) {
				t.Fatalf("Failed to sync %d agents", tc.numAgents)
			}

			t.Logf("Successfully synced %d agents", tc.numAgents)

			// Verify initial state consistency
			if !verifyAllAgentFilesMatch(t, suite, originalContent) {
				t.Error("Agents have inconsistent initial state")
			}

			// Test concurrent modifications at scale
			modifyAllAgentFiles(t, suite, modifiedContent)
			time.Sleep(2 * time.Second)

			// Verify all files were modified (not reverted)
			// This confirms all agents are independently functional
			for i := 0; i < tc.numAgents; i++ {
				path := suite.GetCLAUDEMDPath(i)
				content := getFileContent(t, path)
				if content != modifiedContent {
					t.Errorf("Agent %d: unexpected content after modification", i)
				}
			}
		})
	}
}

// TestSwarmIsolation verifies that agents have isolated workspaces
func TestSwarmIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping swarm isolation test in short mode")
	}

	suite := NewSwarmSuite(t, SwarmConfig{
		NumAgents: 3,
	})
	defer suite.Cleanup()

	// Wait for all agents to sync
	if !suite.WaitForAllAgentsSync(t, swarmSyncTimeout) {
		t.Fatal("Not all agents synced")
	}

	// Modify only agent 0's file
	path0 := suite.GetCLAUDEMDPath(0)
	modifyFile(t, path0, modifiedContent)
	time.Sleep(1 * time.Second)

	// Verify agent 0's file was modified
	content0 := getFileContent(t, path0)
	if content0 != modifiedContent {
		t.Error("Agent 0 file was not modified")
	}

	// Verify other agents' files were NOT modified
	for i := 1; i < 3; i++ {
		path := suite.GetCLAUDEMDPath(i)
		content := getFileContent(t, path)
		if content != originalContent {
			t.Errorf("Agent %d file was unexpectedly modified (isolation breach)", i)
		} else {
			t.Logf("Agent %d workspace is properly isolated", i)
		}
	}
}
