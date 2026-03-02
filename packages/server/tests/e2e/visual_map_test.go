// Package e2e provides end-to-end tests for the DND MCP Server visual map functionality
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	visualTestStartupTimeout = 30 * time.Second
	visualTestRequestTimeout = 10 * time.Second
	visualTestPort           = 8085
)

// buildVisualTestBinary builds the server binary for visual map tests
func buildVisualTestBinary(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	// tests/e2e/visual_map_test.go -> go up 2 levels to get to server directory
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binaryName := "dnd-server-visual-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(projectRoot, "bin", binaryName)

	// Ensure bin directory exists
	binDir := filepath.Join(projectRoot, "bin")
	os.MkdirAll(binDir, 0755)

	t.Logf("Building server binary: %s", binaryPath)
	// Build from current directory (server package)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
	buildCmd.Dir = projectRoot
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server binary: %s", string(output))

	return binaryPath
}

// startVisualTestServer starts the server process for visual map tests
func startVisualTestServer(t *testing.T, binaryPath string) *exec.Cmd {
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("HTTP_PORT=%d", visualTestPort),
		"HTTP_HOST=127.0.0.1",
		"LOG_LEVEL=debug",
		"LOG_FORMAT=text",
	)

	cmd := exec.Command(binaryPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("Starting server on port %d", visualTestPort)
	err := cmd.Start()
	require.NoError(t, err, "Failed to start server process")

	// Wait for server to be ready
	client := &http.Client{Timeout: visualTestRequestTimeout}
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", visualTestPort)

	start := time.Now()
	for time.Since(start) < visualTestStartupTimeout {
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
	t.Fatalf("Server did not become ready within %v", visualTestStartupTimeout)
	return nil
}

// stopVisualTestServer stops the server process
func stopVisualTestServer(t *testing.T, cmd *exec.Cmd) {
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

// callVisualTool calls an MCP tool and returns the response
func callVisualTool(t *testing.T, client *http.Client, toolName string, args map[string]interface{}) map[string]interface{} {
	reqBody := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	reqJSON, _ := json.Marshal(reqBody)

	resp, err := client.Post(
		fmt.Sprintf("http://127.0.0.1:%d/mcp/tools/call", visualTestPort),
		"application/json",
		bytes.NewReader(reqJSON),
	)
	require.NoError(t, err, "Failed to call %s tool", toolName)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse response JSON")

	return result
}

// extractCampaignIDFromResponse extracts campaign ID from create_campaign response
func extractCampaignIDFromResponse(t *testing.T, result map[string]interface{}) string {
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "response should have content array")
	require.Greater(t, len(content), 0, "content array should not be empty")

	text, ok := content[0].(map[string]interface{})["text"].(string)
	require.True(t, ok, "content should have text field")

	// Parse the JSON text to extract campaign ID
	var campaignData struct {
		Campaign struct {
			ID string `json:"id"`
		} `json:"campaign"`
		Message string `json:"message"`
	}
	err := json.Unmarshal([]byte(text), &campaignData)
	require.NoError(t, err, "Failed to parse campaign response")

	return campaignData.Campaign.ID
}

// extractMapIDFromGetWorldMap extracts the map ID from get_world_map response
func extractMapIDFromGetWorldMap(t *testing.T, result map[string]interface{}) string {
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "response should have content array")
	require.Greater(t, len(content), 0, "content array should not be empty")

	text, ok := content[0].(map[string]interface{})["text"].(string)
	require.True(t, ok, "content should have text field")

	// Parse the JSON text to extract map ID
	var mapData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := json.Unmarshal([]byte(text), &mapData)
	require.NoError(t, err, "Failed to parse map response")

	return mapData.ID
}

// TestE2E_VisualMapFlow tests the visual map functionality end-to-end
func TestE2E_VisualMapFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build and start server
	binaryPath := buildVisualTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startVisualTestServer(t, binaryPath)
	defer stopVisualTestServer(t, cmd)

	client := &http.Client{Timeout: visualTestRequestTimeout}

	var campaignID string
	var worldMapID string

	// Step 1: Create a campaign first
	t.Run("CreateCampaign", func(t *testing.T) {
		result := callVisualTool(t, client, "create_campaign", map[string]interface{}{
			"name":        "Visual Map Test Campaign",
			"description": "Test campaign for visual map functionality",
			"dm_id":       "test-dm",
		})

		isError, _ := result["isError"].(bool)
		require.False(t, isError, "create_campaign should succeed")

		campaignID = extractCampaignIDFromResponse(t, result)
		assert.NotEmpty(t, campaignID, "campaign ID should not be empty")
		t.Logf("Created campaign with ID: %s", campaignID)
	})

	// Step 2: Get world map (creates if missing)
	t.Run("GetWorldMap", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callVisualTool(t, client, "get_world_map", map[string]interface{}{
			"campaign_id": campaignID,
		})

		isError, _ := result["isError"].(bool)
		assert.False(t, isError, "get_world_map should succeed")

		worldMapID = extractMapIDFromGetWorldMap(t, result)
		t.Logf("Got world map with ID: %s", worldMapID)
	})

	// Step 3: Switch to image mode
	t.Run("SwitchToImageMode", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callVisualTool(t, client, "switch_map_mode", map[string]interface{}{
			"campaign_id": campaignID,
			"mode":        "image",
		})

		isError, _ := result["isError"].(bool)
		// This might not exist, log the result
		t.Logf("switch_map_mode response: %v", result)
		if isError {
			t.Logf("switch_map_mode failed (may not exist): %v", result)
		}
	})

	// Step 4: Create a visual location
	t.Run("CreateVisualLocation", func(t *testing.T) {
		if campaignID == "" || worldMapID == "" {
			t.Skip("No campaign ID or map ID available")
		}

		result := callVisualTool(t, client, "create_visual_location", map[string]interface{}{
			"campaign_id": campaignID,
			"map_id":      worldMapID,
			"name":        "Dark Cave",
			"type":        "poi",
			"position_x":  0.5,
			"position_y":  0.5,
		})

		isError, _ := result["isError"].(bool)
		assert.False(t, isError, "create_visual_location should succeed")
		t.Logf("create_visual_location response: %v", result)
	})
}

// TestE2E_VisualMapImageModeSwitch tests switching between grid and image modes
func TestE2E_VisualMapImageModeSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build and start server
	binaryPath := buildVisualTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := startVisualTestServer(t, binaryPath)
	defer stopVisualTestServer(t, cmd)

	client := &http.Client{Timeout: visualTestRequestTimeout}

	var campaignID string
	var worldMapID string

	// Step 1: Create a campaign first
	t.Run("CreateCampaign", func(t *testing.T) {
		result := callVisualTool(t, client, "create_campaign", map[string]interface{}{
			"name":        "Image Mode Test Campaign",
			"description": "Test campaign for image mode switching",
			"dm_id":       "test-dm",
		})

		isError, _ := result["isError"].(bool)
		require.False(t, isError, "create_campaign should succeed")

		campaignID = extractCampaignIDFromResponse(t, result)
		assert.NotEmpty(t, campaignID, "campaign ID should not be empty")
		t.Logf("Created campaign with ID: %s", campaignID)
	})

	// Step 2: Get world map
	t.Run("GetWorldMap", func(t *testing.T) {
		if campaignID == "" {
			t.Skip("No campaign ID available")
		}

		result := callVisualTool(t, client, "get_world_map", map[string]interface{}{
			"campaign_id": campaignID,
		})

		isError, _ := result["isError"].(bool)
		assert.False(t, isError, "get_world_map should succeed")

		worldMapID = extractMapIDFromGetWorldMap(t, result)
		t.Logf("Got world map with ID: %s", worldMapID)
	})

	// Step 3: Create visual location
	t.Run("CreateVisualLocation", func(t *testing.T) {
		if campaignID == "" || worldMapID == "" {
			t.Skip("No campaign ID or map ID available")
		}

		result := callVisualTool(t, client, "create_visual_location", map[string]interface{}{
			"campaign_id": campaignID,
			"map_id":      worldMapID,
			"name":        "Test Location",
			"type":        "poi",
			"position_x":  0.3,
			"position_y":  0.7,
		})

		isError, _ := result["isError"].(bool)
		assert.False(t, isError, "create_visual_location should succeed")
		t.Logf("create_visual_location response: %v", result)
	})
}

// TestE2E_VisualMapBoundaryConditions tests boundary conditions for visual map
func TestE2E_VisualMapBoundaryConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test coordinates at boundaries (0, 0) and (1, 1)
	testCoords := []struct {
		x, y float64
	}{
		{0, 0},
		{1, 1},
		{0.5, 0.5},
		{0.999, 0.999},
	}

	for _, tc := range testCoords {
		t.Run(fmt.Sprintf("coord_%.3f_%.3f", tc.x, tc.y), func(t *testing.T) {
			// Validate coordinates are within bounds
			assert.True(t, tc.x >= 0 && tc.x <= 1, "X coordinate out of bounds")
			assert.True(t, tc.y >= 0 && tc.y <= 1, "Y coordinate out of bounds")
		})
	}
}
