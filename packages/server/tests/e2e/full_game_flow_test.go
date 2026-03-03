// Package e2e provides end-to-end tests for complete D&D game flow
package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	flowTestStartupTimeout = 30 * time.Second
	flowTestRequestTimeout = 10 * time.Second
)

// ============================================
// Test Infrastructure - Server Management
// ============================================

// buildFlowTestBinary builds the server binary for flow tests
func buildFlowTestBinary(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binaryName := "dnd-server-flow-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(projectRoot, "bin", binaryName)

	t.Logf("Building server binary: %s", binaryPath)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
	buildCmd.Dir = projectRoot
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server binary: %s", string(output))

	return binaryPath
}

// startFlowTestServer starts the server process and waits for it to be ready
func startFlowTestServer(t *testing.T, binaryPath string, port int) *exec.Cmd {
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("HTTP_PORT=%d", port),
		"HTTP_HOST=127.0.0.1",
		"LOG_LEVEL=debug",
		"LOG_FORMAT=text",
	)

	cmd := exec.Command(binaryPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("Starting server on port %d", port)
	err := cmd.Start()
	require.NoError(t, err, "Failed to start server process")

	// Wait for server to be ready
	client := &http.Client{Timeout: flowTestRequestTimeout}
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	start := time.Now()
	for time.Since(start) < flowTestStartupTimeout {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("Server is ready after %v", time.Since(start))
				return cmd
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	cmd.Process.Kill()
	t.Fatalf("Server failed to start within %v", flowTestStartupTimeout)
	return nil
}

// stopFlowTestServer stops the server process
func stopFlowTestServer(t *testing.T, cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	t.Log("Stopping server...")
	cmd.Process.Signal(os.Interrupt)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		t.Log("Server didn't stop gracefully, killing...")
		cmd.Process.Kill()
	case err := <-done:
		if err != nil {
			t.Logf("Server exited with error: %v", err)
		} else {
			t.Log("Server stopped gracefully")
		}
	}
}

// ============================================
// Test Infrastructure - Tool Calling Helpers
// ============================================

// callFlowTool calls an MCP tool and returns the response
func callFlowTool(t *testing.T, client *http.Client, port int, toolName string, args map[string]interface{}) map[string]interface{} {
	reqBody := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	reqJSON, _ := json.Marshal(reqBody)

	resp, err := client.Post(
		fmt.Sprintf("http://127.0.0.1:%d/mcp/tools/call", port),
		"application/json",
		bytes.NewReader(reqJSON),
	)
	require.NoError(t, err, "Failed to call %s tool", toolName)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body from %s", toolName)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse response JSON from %s", toolName)

	return result
}

// extractID extracts an ID field from a tool response content
func extractID(t *testing.T, result map[string]interface{}, fieldPath ...string) string {
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "response should have content array")
	require.Greater(t, len(content), 0, "content array should not be empty")

	text, ok := content[0].(map[string]interface{})["text"].(string)
	require.True(t, ok, "content should have text field")

	var data map[string]interface{}
	err := json.Unmarshal([]byte(text), &data)
	require.NoError(t, err, "Failed to parse response content JSON")

	// Navigate through field path
	current := data
	for i, key := range fieldPath {
		if i == len(fieldPath)-1 {
			id, ok := current[key].(string)
			require.True(t, ok, "field %s should be a string", key)
			return id
		}
		current, ok = current[key].(map[string]interface{})
		require.True(t, ok, "field %s should be an object", key)
	}

	t.Fatal("reached end of fieldPath without finding ID")
	return ""
}

// assertNoError asserts that a tool response has no error
func assertNoError(t *testing.T, result map[string]interface{}, toolName string) {
	// isError is omitempty, so it might not be present for successful responses
	isError, ok := result["isError"].(bool)
	if ok && isError {
		require.Fail(t, fmt.Sprintf("%s should not return an error. Response: %v", toolName, result))
	}
	// If isError is not present or is false, the call was successful
}

// assertError asserts that a tool response has an error
func assertError(t *testing.T, result map[string]interface{}, toolName string) {
	isError, ok := result["isError"].(bool)
	require.True(t, ok && isError, "%s should return an error (isError should be true). Response: %v", toolName, result)
}

// listTools retrieves all registered tools from the server
func listTools(t *testing.T, client *http.Client, port int) map[string]bool {
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/mcp/tools", port))
	require.NoError(t, err, "Failed to call tools endpoint")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "tools not found in response")

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		name := toolMap["name"].(string)
		toolNames[name] = true
	}

	return toolNames
}

// ============================================
// Test Scenario 1: Campaign Preparation Flow
// Tools: create_campaign, create_character, get_world_map, move_to
// ============================================

// TestCampaignPreparationFlow tests the complete campaign setup flow
// This is Scenario 1 from T7.5-1
func TestCampaignPreparationFlow(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	// Setup database environment
	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18090

	binaryPath := buildFlowTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startFlowTestServer(t, binaryPath, port)
	defer stopFlowTestServer(t, cmd)

	client := &http.Client{Timeout: flowTestRequestTimeout}

	// Verify all expected tools are registered
	t.Run("VerifyToolsRegistered", func(t *testing.T) {
		tools := listTools(t, client, port)

		expectedTools := []string{
			"create_campaign",
			"get_campaign",
			"create_character",
			"get_character",
			"get_world_map",
			"move_to",
		}

		for _, tool := range expectedTools {
			assert.True(t, tools[tool], "%s tool should be registered", tool)
		}

		t.Logf("All %d expected tools are registered", len(expectedTools))
	})

	var campaignID, characterID string

	// Step 1: Create a campaign
	t.Run("Step1_CreateCampaign", func(t *testing.T) {
		result := callFlowTool(t, client, port, "create_campaign", map[string]interface{}{
			"name":        "The Lost Mine of Phandelver",
			"description": "A classic D&D 5e adventure for levels 1-5",
			"dm_id":       "test-dm-001",
			"settings": map[string]interface{}{
				"max_players":     5,
				"start_level":     1,
				"ruleset":         "5e",
				"context_window":  50,
			},
		})

		assertNoError(t, result, "create_campaign")
		campaignID = extractID(t, result, "campaign", "id")
		assert.NotEmpty(t, campaignID, "campaign ID should not be empty")

		t.Logf("Created campaign with ID: %s", campaignID)
	})

	// Step 2: Get campaign details
	t.Run("Step2_GetCampaign", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "get_campaign", map[string]interface{}{
			"campaign_id": campaignID,
		})

		assertNoError(t, result, "get_campaign")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		var data map[string]interface{}
		err := json.Unmarshal([]byte(text), &data)
		require.NoError(t, err)

		campaign, ok := data["campaign"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "The Lost Mine of Phandelver", campaign["name"])

		t.Logf("Retrieved campaign: %v", campaign["name"])
	})

	// Step 3: Create a player character
	t.Run("Step3_CreateCharacter", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "create_character", map[string]interface{}{
			"campaign_id": campaignID,
			"name":        "Gimble Halfshield",
			"race":        "Dwarf",
			"class":       "Cleric",
			"level":       1,
			"background":  "Acolyte",
			"player_id":   "player-001",
			"is_npc":      false,
			"stats": map[string]int{
				"strength":     14,
				"dexterity":    10,
				"constitution": 16,
				"intelligence": 12,
				"wisdom":       16,
				"charisma":     10,
			},
		})

		assertNoError(t, result, "create_character")
		characterID = extractID(t, result, "character", "id")
		assert.NotEmpty(t, characterID, "character ID should not be empty")

		t.Logf("Created character with ID: %s", characterID)
	})

	// Step 4: Get character details
	t.Run("Step4_GetCharacter", func(t *testing.T) {
		if characterID == "" {
			t.Skip("No character ID available")
		}

		result := callFlowTool(t, client, port, "get_character", map[string]interface{}{
			"character_id": characterID,
		})

		assertNoError(t, result, "get_character")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		var data map[string]interface{}
		err := json.Unmarshal([]byte(text), &data)
		require.NoError(t, err)

		character, ok := data["character"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Gimble Halfshield", character["name"])
		assert.Equal(t, "Dwarf", character["race"])

		t.Logf("Retrieved character: %v the %v", character["name"], character["class"])
	})

	// Step 5: Get world map (creates if missing)
	t.Run("Step5_GetWorldMap", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "get_world_map", map[string]interface{}{
			"campaign_id": campaignID,
		})

		assertNoError(t, result, "get_world_map")

		t.Logf("World map retrieved for campaign: %s", campaignID)
	})

	// Step 6: Move to a location (create a visual location first)
	t.Run("Step6_MoveToLocation", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		// First create a visual location
		createResult := callFlowTool(t, client, port, "create_visual_location", map[string]interface{}{
			"campaign_id": campaignID,
			"name":        "Phandalin",
			"position": map[string]interface{}{
				"x": 100,
				"y": 150,
			},
			"description": "A small frontier town",
			"icon_type":   "town",
		})

		// This might work or fail depending on implementation
		if isError, _ := createResult["isError"].(bool); !isError {
			t.Logf("Created visual location")
		}

		// Now try to move to that location
		result := callFlowTool(t, client, port, "move_to", map[string]interface{}{
			"campaign_id": campaignID,
			"location_id": "phandalin", // Using a location ID
		})

		// move_to might fail without proper location setup
		if isError, _ := result["isError"].(bool); isError {
			t.Logf("move_to returned error (expected without proper setup): %v", result)
		} else {
			t.Logf("Successfully moved to location")
		}
	})
}

// ============================================
// Test Scenario 2: Exploration Flow
// Tools: roll_check, roll_save, save_message, get_context
// ============================================

// TestExplorationFlow tests the exploration phase flow
// This is Scenario 2 from T7.5-1
func TestExplorationFlow(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18091

	binaryPath := buildFlowTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startFlowTestServer(t, binaryPath, port)
	defer stopFlowTestServer(t, cmd)

	client := &http.Client{Timeout: flowTestRequestTimeout}

	var campaignID string

	// Setup: Create a campaign
	t.Run("Setup_CreateCampaign", func(t *testing.T) {
		result := callFlowTool(t, client, port, "create_campaign", map[string]interface{}{
			"name":        "Exploration Test Campaign",
			"description": "Testing exploration mechanics",
			"dm_id":       "test-dm-002",
		})

		assertNoError(t, result, "create_campaign")
		campaignID = extractID(t, result, "campaign", "id")
		t.Logf("Created campaign with ID: %s", campaignID)
	})

	// Step 1: Roll a skill check (Perception)
	t.Run("Step1_RollPerceptionCheck", func(t *testing.T) {
		result := callFlowTool(t, client, port, "roll_check", map[string]interface{}{
			"check_type": "perception",
			"modifier":   5,
			"reason":     "Searching for hidden traps",
		})

		assertNoError(t, result, "roll_check")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		t.Logf("Perception check result: %s", text)

		// Verify the result contains the expected information
		assert.Contains(t, text, "Perception")
		assert.Contains(t, text, "d20")
	})

	// Step 2: Roll a saving throw (Dexterity)
	t.Run("Step2_RollDexteritySave", func(t *testing.T) {
		result := callFlowTool(t, client, port, "roll_save", map[string]interface{}{
			"save_type": "dexterity",
			"modifier":  3,
			"reason":    "Avoiding a falling boulder",
		})

		assertNoError(t, result, "roll_save")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		t.Logf("Dexterity save result: %s", text)

		// Verify the result contains the expected information
		assert.Contains(t, text, "Dexterity")
		assert.Contains(t, text, "Saving Throw")
	})

	// Step 3: Save a message to the campaign history
	t.Run("Step3_SaveMessage", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "save_message", map[string]interface{}{
			"campaign_id": campaignID,
			"role":        "assistant",
			"content":     "The party discovers an ancient stone door covered in runes. A faint magical aura emanates from the keyhole.",
			"message_type": "narration",
		})

		assertNoError(t, result, "save_message")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		t.Logf("Saved message: %s", text)
	})

	// Step 4: Get context for the campaign
	t.Run("Step4_GetContext", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "get_context", map[string]interface{}{
			"campaign_id": campaignID,
			"include":     []string{"messages", "characters", "location"},
		})

		assertNoError(t, result, "get_context")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		var data map[string]interface{}
		err := json.Unmarshal([]byte(text), &data)
		require.NoError(t, err)

		// Verify context structure
		assert.Contains(t, data, "campaign_id")

		t.Logf("Retrieved context with fields: %v", getKeys(data))
	})

	// Step 5: Get raw context
	t.Run("Step5_GetRawContext", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "get_raw_context", map[string]interface{}{
			"campaign_id": campaignID,
		})

		assertNoError(t, result, "get_raw_context")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		t.Logf("Raw context length: %d characters", len(text))

		// Raw context should be longer than formatted context
		assert.Greater(t, len(text), 100, "Raw context should contain substantial data")
	})

	// Step 6: Roll multiple checks (Athletics, Arcana, Stealth)
	t.Run("Step6_RollMultipleChecks", func(t *testing.T) {
		checks := []struct {
			checkType string
			modifier  int
			reason    string
		}{
			{"athletics", 4, "Climbing a steep wall"},
			{"arcana", 6, "Deciphering magical runes"},
			{"stealth", 3, "Sneaking past guards"},
		}

		for _, check := range checks {
			result := callFlowTool(t, client, port, "roll_check", map[string]interface{}{
				"check_type": check.checkType,
				"modifier":   check.modifier,
				"reason":     check.reason,
			})

			assertNoError(t, result, "roll_check for "+check.checkType)

			content, ok := result["content"].([]interface{})
			require.True(t, ok)
			text, ok := content[0].(map[string]interface{})["text"].(string)
			require.True(t, ok)

			t.Logf("%s check result: %s", strings.Title(check.checkType), text)
		}
	})
}

// ============================================
// Test Scenario 3: Combat Flow
// Tools: enter_battle_map, start_combat, attack, end_turn, end_combat, exit_battle_map
// ============================================

// TestCombatFlow tests the complete combat flow
// This is Scenario 3 from T7.5-1
func TestCombatFlow(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18092

	binaryPath := buildFlowTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startFlowTestServer(t, binaryPath, port)
	defer stopFlowTestServer(t, cmd)

	client := &http.Client{Timeout: flowTestRequestTimeout}

	var campaignID, characterID, combatID string

	// Setup: Create campaign and character
	t.Run("Setup_CreateCampaignAndCharacter", func(t *testing.T) {
		// Create campaign
		campaignResult := callFlowTool(t, client, port, "create_campaign", map[string]interface{}{
			"name":        "Combat Test Campaign",
			"description": "Testing combat mechanics",
			"dm_id":       "test-dm-003",
		})

		assertNoError(t, campaignResult, "create_campaign")
		campaignID = extractID(t, campaignResult, "campaign", "id")
		t.Logf("Created campaign with ID: %s", campaignID)

		// Create character
		charResult := callFlowTool(t, client, port, "create_character", map[string]interface{}{
			"campaign_id": campaignID,
			"name":        "Thorin Ironforge",
			"race":        "Dwarf",
			"class":       "Fighter",
			"level":       3,
			"player_id":   "player-003",
			"is_npc":      false,
			"stats": map[string]int{
				"strength":     16,
				"dexterity":    12,
				"constitution": 14,
				"intelligence": 10,
				"wisdom":       12,
				"charisma":     8,
			},
			"hp": map[string]interface{}{
				"current": 27,
				"max":     27,
			},
			"ac": 16,
		})

		assertNoError(t, charResult, "create_character")
		characterID = extractID(t, charResult, "character", "id")
		t.Logf("Created character with ID: %s", characterID)
	})

	// Step 1: Enter battle map
	t.Run("Step1_EnterBattleMap", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "enter_battle_map", map[string]interface{}{
			"campaign_id":       campaignID,
			"location_id":       "goblin_cave",
			"create_if_missing": true,
		})

		// This might work or fail depending on whether we have a valid location_id
		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("enter_battle_map returned error (may need proper location setup): %v", result)
		} else {
			t.Logf("Successfully entered battle map")
		}
	})

	// Step 2: Start combat
	t.Run("Step2_StartCombat", func(t *testing.T) {
		if campaignID == "" || characterID == "" {
			t.Skip("No campaign or character ID available")
		}

		result := callFlowTool(t, client, port, "start_combat", map[string]interface{}{
			"campaign_id": campaignID,
			"location_id": "goblin_cave",
			"combatants": []map[string]interface{}{
				{
					"character_id": characterID,
					"initiative":   0,
					"is_player":    true,
				},
				{
					"name":       "Goblin Scout",
					"initiative": 3,
					"is_player":  false,
					"stats": map[string]int{
						"ac":  12,
						"hp":  7,
					},
				},
			},
		})

		assertNoError(t, result, "start_combat")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		var data map[string]interface{}
		err := json.Unmarshal([]byte(text), &data)
		require.NoError(t, err)

		combat, ok := data["combat"].(map[string]interface{})
		if ok {
			combatID, ok = combat["id"].(string)
			if ok {
				t.Logf("Started combat with ID: %s", combatID)
			}
		}

		t.Logf("Combat started with %d combatants", len(data))
	})

	// Step 3: Get combat state
	t.Run("Step3_GetCombatState", func(t *testing.T) {
		if combatID == "" {
			t.Skip("No combat ID available")
		}

		result := callFlowTool(t, client, port, "get_combat_state", map[string]interface{}{
			"combat_id": combatID,
		})

		assertNoError(t, result, "get_combat_state")

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		text, ok := content[0].(map[string]interface{})["text"].(string)
		require.True(t, ok)

		t.Logf("Combat state: %s", text)
	})

	// Step 4: Perform an attack
	t.Run("Step4_Attack", func(t *testing.T) {
		if combatID == "" || characterID == "" {
			t.Skip("No combat or character ID available")
		}

		result := callFlowTool(t, client, port, "attack", map[string]interface{}{
			"combat_id":      combatID,
			"attacker_id":    characterID,
			"target_id":      "goblin_scout",
			"attack_type":    "melee",
			"damage_type":    "slashing",
			"to_hit_modifier": 5,
			"damage_dice":    "1d8+3",
		})

		// Attack might fail if target setup is incorrect
		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("Attack returned error: %v", result)
		} else {
			content, ok := result["content"].([]interface{})
			if ok {
				text, ok := content[0].(map[string]interface{})["text"].(string)
				if ok {
					t.Logf("Attack result: %s", text)
				}
			}
		}
	})

	// Step 5: End turn
	t.Run("Step5_EndTurn", func(t *testing.T) {
		if combatID == "" {
			t.Skip("No combat ID available")
		}

		result := callFlowTool(t, client, port, "end_turn", map[string]interface{}{
			"combat_id": combatID,
		})

		assertNoError(t, result, "end_turn")

		t.Logf("Turn ended")
	})

	// Step 6: End combat
	t.Run("Step6_EndCombat", func(t *testing.T) {
		if combatID == "" {
			t.Skip("No combat ID available")
		}

		result := callFlowTool(t, client, port, "end_combat", map[string]interface{}{
			"combat_id": combatID,
		})

		assertNoError(t, result, "end_combat")

		content, ok := result["content"].([]interface{})
		if ok && len(content) > 0 {
			text, ok := content[0].(map[string]interface{})["text"].(string)
			if ok {
				t.Logf("Combat ended: %s", text)
			}
		}

		t.Logf("Combat ended successfully")
	})

	// Step 7: Exit battle map
	t.Run("Step7_ExitBattleMap", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callFlowTool(t, client, port, "exit_battle_map", map[string]interface{}{
			"campaign_id":     campaignID,
			"keep_battle_map": false,
		})

		// This might fail if we never successfully entered a battle map
		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("exit_battle_map returned error: %v", result)
		} else {
			t.Logf("Successfully exited battle map")
		}
	})
}

// ============================================
// Test Scenario 4: Spellcasting Flow
// Tools: cast_spell with damage/healing effects
// ============================================

// TestSpellcastingFlow tests the spellcasting mechanics
// This is Scenario 4 from T7.5-1
func TestSpellcastingFlow(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18093

	binaryPath := buildFlowTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startFlowTestServer(t, binaryPath, port)
	defer stopFlowTestServer(t, cmd)

	client := &http.Client{Timeout: flowTestRequestTimeout}

	var campaignID, characterID, combatID string

	// Setup: Create campaign, character, and combat
	t.Run("Setup_CreateForSpellcasting", func(t *testing.T) {
		// Create campaign
		campaignResult := callFlowTool(t, client, port, "create_campaign", map[string]interface{}{
			"name":        "Spellcasting Test Campaign",
			"description": "Testing spellcasting mechanics",
			"dm_id":       "test-dm-004",
		})

		assertNoError(t, campaignResult, "create_campaign")
		campaignID = extractID(t, campaignResult, "campaign", "id")

		// Create spellcaster character
		charResult := callFlowTool(t, client, port, "create_character", map[string]interface{}{
			"campaign_id": campaignID,
			"name":        "Elara Moonwhisper",
			"race":        "Elf",
			"class":       "Wizard",
			"level":       3,
			"player_id":   "player-004",
			"is_npc":      false,
			"stats": map[string]int{
				"strength":     8,
				"dexterity":    14,
				"constitution": 12,
				"intelligence": 18,
				"wisdom":       12,
				"charisma":     10,
			},
			"hp": map[string]interface{}{
				"current": 20,
				"max":     20,
			},
			"ac": 12,
		})

		assertNoError(t, charResult, "create_character")
		characterID = extractID(t, charResult, "character", "id")

		// Start combat for spellcasting
		combatResult := callFlowTool(t, client, port, "start_combat", map[string]interface{}{
			"campaign_id": campaignID,
			"location_id": "test_arena",
			"combatants": []map[string]interface{}{
				{
					"character_id": characterID,
					"initiative":   5,
					"is_player":    true,
				},
				{
					"name":      "Training Dummy",
					"initiative": 0,
					"is_player": false,
					"stats": map[string]int{
						"ac": 10,
						"hp": 30,
					},
				},
			},
		})

		if isError, _ := combatResult["isError"].(bool); !isError {
			content := combatResult["content"].([]interface{})
			text := content[0].(map[string]interface{})["text"].(string)
			var data map[string]interface{}
			json.Unmarshal([]byte(text), &data)
			if combat, ok := data["combat"].(map[string]interface{}); ok {
				combatID, _ = combat["id"].(string)
			}
		}

		t.Logf("Setup complete - Campaign: %s, Character: %s, Combat: %s",
			campaignID, characterID, combatID)
	})

	// Step 1: Cast a damage spell (Firebolt)
	t.Run("Step1_CastDamageSpell", func(t *testing.T) {
		if combatID == "" || characterID == "" {
			t.Skip("No combat or character ID available")
		}

		result := callFlowTool(t, client, port, "cast_spell", map[string]interface{}{
			"combat_id":   combatID,
			"caster_id":   characterID,
			"spell_name":  "Firebolt",
			"spell_level": 1,
			"target_id":   "training_dummy",
			"damage_dice": "2d10",
			"damage_type": "fire",
		})

		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("cast_spell returned error: %v", result)
		} else {
			content, ok := result["content"].([]interface{})
			if ok && len(content) > 0 {
				text, ok := content[0].(map[string]interface{})["text"].(string)
				if ok {
					t.Logf("Firebolt result: %s", text)

					// Verify damage information is present
					assert.Contains(t, text, "damage", "Result should mention damage")
				}
			}
		}
	})

	// Step 2: Cast a healing spell (Cure Wounds)
	t.Run("Step2_CastHealingSpell", func(t *testing.T) {
		if combatID == "" || characterID == "" {
			t.Skip("No combat or character ID available")
		}

		// First, we need to create an ally that can be healed
		// For this test, we'll try to heal the caster
		result := callFlowTool(t, client, port, "cast_spell", map[string]interface{}{
			"combat_id":     combatID,
			"caster_id":     characterID,
			"spell_name":    "Cure Wounds",
			"spell_level":   1,
			"target_id":     characterID,
			"healing_dice":  "1d8+4",
			"healing_type":  "healing",
		})

		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("cast_spell (healing) returned error: %v", result)
		} else {
			content, ok := result["content"].([]interface{})
			if ok && len(content) > 0 {
				text, ok := content[0].(map[string]interface{})["text"].(string)
				if ok {
					t.Logf("Cure Wounds result: %s", text)

					// Verify healing information is present
					assert.Contains(t, text, "heal", "Result should mention healing")
				}
			}
		}
	})

	// Step 3: Cast an area-of-effect spell (Burning Hands)
	t.Run("Step3_CastAoESpell", func(t *testing.T) {
		if combatID == "" || characterID == "" {
			t.Skip("No combat or character ID available")
		}

		result := callFlowTool(t, client, port, "cast_spell", map[string]interface{}{
			"combat_id":      combatID,
			"caster_id":      characterID,
			"spell_name":     "Burning Hands",
			"spell_level":    1,
			"damage_dice":    "3d6",
			"damage_type":    "fire",
			"is_aoe":         true,
			"aoe_shape":      "cone",
			"aoe_size_feet":  15,
			"target_ids":     []string{"training_dummy"},
		})

		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("cast_spell (AoE) returned error: %v", result)
		} else {
			content, ok := result["content"].([]interface{})
			if ok && len(content) > 0 {
				text, ok := content[0].(map[string]interface{})["text"].(string)
				if ok {
					t.Logf("Burning Hands result: %s", text)
				}
			}
		}
	})

	// Step 4: Cast a spell with a saving throw (Sleep)
	t.Run("Step4_CastSavingThrowSpell", func(t *testing.T) {
		if combatID == "" || characterID == "" {
			t.Skip("No combat or character ID available")
		}

		result := callFlowTool(t, client, port, "cast_spell", map[string]interface{}{
			"combat_id":       combatID,
			"caster_id":       characterID,
			"spell_name":      "Sleep",
			"spell_level":     1,
			"save_dc":         13,
			"save_type":       "wisdom",
			"effect_duration": 60, // seconds
		})

		isError, _ := result["isError"].(bool)
		if isError {
			t.Logf("cast_spell (saving throw) returned error: %v", result)
		} else {
			content, ok := result["content"].([]interface{})
			if ok && len(content) > 0 {
				text, ok := content[0].(map[string]interface{})["text"].(string)
				if ok {
					t.Logf("Sleep result: %s", text)
				}
			}
		}
	})

	// Cleanup: End combat
	t.Run("Cleanup_EndCombat", func(t *testing.T) {
		if combatID != "" {
			result := callFlowTool(t, client, port, "end_combat", map[string]interface{}{
				"combat_id": combatID,
			})

			if isError, _ := result["isError"].(bool); !isError {
				t.Logf("Combat ended successfully")
			}
		}
	})
}

// ============================================
// Helper Functions
// ============================================

// getKeys extracts keys from a map for logging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ============================================
// Integration Issues Report Test
// ============================================

// TestIntegrationIssuesReport runs all scenarios and compiles a report
// This test can be used as the basis for T7.5-6 integration fixes
func TestIntegrationIssuesReport(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18094

	binaryPath := buildFlowTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startFlowTestServer(t, binaryPath, port)
	defer stopFlowTestServer(t, cmd)

	client := &http.Client{Timeout: flowTestRequestTimeout}

	// Get all available tools
	t.Run("Report_AvailableTools", func(t *testing.T) {
		tools := listTools(t, client, port)

		// Count tools by category
		categories := map[string][]string{
			"Campaign":  {},
			"Character": {},
			"Combat":    {},
			"Map":       {},
			"Dice":      {},
			"Context":   {},
			"Import":    {},
		}

		for tool := range tools {
			switch {
			case strings.HasPrefix(tool, "campaign") || strings.HasPrefix(tool, "list_campaign") || strings.HasPrefix(tool, "delete_campaign") || strings.HasPrefix(tool, "get_campaign"):
				categories["Campaign"] = append(categories["Campaign"], tool)
			case strings.HasPrefix(tool, "character") || strings.HasPrefix(tool, "create_character") || strings.HasPrefix(tool, "list_character") || strings.HasPrefix(tool, "delete_character") || strings.HasPrefix(tool, "get_character") || strings.HasPrefix(tool, "update_character"):
				categories["Character"] = append(categories["Character"], tool)
			case strings.HasPrefix(tool, "combat") || tool == "attack" || tool == "cast_spell" || tool == "start_combat" || tool == "end_combat" || tool == "end_turn" || tool == "get_combat":
				categories["Combat"] = append(categories["Combat"], tool)
			case strings.Contains(tool, "map") || strings.Contains(tool, "move") || strings.Contains(tool, "location") || strings.Contains(tool, "visual"):
				categories["Map"] = append(categories["Map"], tool)
			case strings.HasPrefix(tool, "roll"):
				categories["Dice"] = append(categories["Dice"], tool)
			case strings.Contains(tool, "context") || strings.Contains(tool, "message") || strings.Contains(tool, "save_message"):
				categories["Context"] = append(categories["Context"], tool)
			case strings.Contains(tool, "import"):
				categories["Import"] = append(categories["Import"], tool)
			}
		}

		t.Log("=== Available Tools Report ===")
		totalTools := 0
		for category, toolList := range categories {
			t.Logf("%s: %d tools - %v", category, len(toolList), toolList)
			totalTools += len(toolList)
		}
		t.Logf("Total: %d tools available", totalTools)

		// Verify we have at least 25 tools as per acceptance criteria
		assert.GreaterOrEqual(t, totalTools, 25, "Should have at least 25 tools implemented")
	})

	// Test tool interdependencies
	t.Run("Report_ToolInterdependencies", func(t *testing.T) {
		// Test that campaign is required for character creation
		t.Run("CharacterWithoutCampaign", func(t *testing.T) {
			result := callFlowTool(t, client, port, "create_character", map[string]interface{}{
				"campaign_id": "non-existent-campaign",
				"name":        "Test Character",
				"race":        "Human",
				"class":       "Fighter",
				"level":       1,
				"is_npc":      false,
			})

			// Should fail or create with a warning
			isError, _ := result["isError"].(bool)
			t.Logf("Character creation without valid campaign: isError=%v", isError)
		})

		// Test that combat requires proper setup
		t.Run("CombatWithoutSetup", func(t *testing.T) {
			result := callFlowTool(t, client, port, "get_combat_state", map[string]interface{}{
				"combat_id": "non-existent-combat",
			})

			// Should fail
			assertError(t, result, "get_combat_state")
			t.Logf("Correctly fails when combat doesn't exist")
		})
	})

	t.Log("=== Integration Issues Report Complete ===")
}
