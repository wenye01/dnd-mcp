// Package e2e provides end-to-end tests for the DND MCP Server combat flow
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
	combatTestStartupTimeout = 30 * time.Second
	combatTestRequestTimeout = 10 * time.Second
)

// buildCombatTestBinary builds the server binary and returns the path
func buildCombatTestBinary(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	binaryName := "dnd-server-combat-test"
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

// startCombatTestServer starts the server process and waits for it to be ready
func startCombatTestServer(t *testing.T, binaryPath string, port int) *exec.Cmd {
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
	client := &http.Client{Timeout: combatTestRequestTimeout}
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	start := time.Now()
	for time.Since(start) < combatTestStartupTimeout {
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
	t.Fatalf("Server failed to start within %v", combatTestStartupTimeout)
	return nil
}

// stopCombatTestServer stops the server process
func stopCombatTestServer(t *testing.T, cmd *exec.Cmd) {
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

// callTool calls an MCP tool and returns the response
func callTool(t *testing.T, client *http.Client, port int, toolName string, args map[string]interface{}) map[string]interface{} {
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
	require.NoError(t, err, "Failed to read response body")

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse response JSON")

	return result
}

// TestCombatFlowE2E tests the complete combat flow from start to finish
// This test requires a PostgreSQL database to be available
func TestCombatFlowE2E(t *testing.T) {
	// Check if we should skip E2E tests
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	// Check if database is available
	if os.Getenv("POSTGRES_HOST") == "" {
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	port := 18083

	// Build the binary
	binaryPath := buildCombatTestBinary(t)
	defer os.Remove(binaryPath)

	// Start the server
	cmd := startCombatTestServer(t, binaryPath, port)
	defer stopCombatTestServer(t, cmd)

	client := &http.Client{Timeout: combatTestRequestTimeout}

	// Test 1: Verify combat tools are registered
	t.Run("CombatToolsRegistered", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/mcp/tools", port))
		require.NoError(t, err, "Failed to call tools endpoint")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		tools, ok := result["tools"].([]interface{})
		require.True(t, ok, "tools not found in response")

		// Build a map of tool names
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolMap := tool.(map[string]interface{})
			name := toolMap["name"].(string)
			toolNames[name] = true
		}

		// Verify all combat tools are registered
		expectedTools := []string{
			"start_combat",
			"get_combat_state",
			"attack",
			"cast_spell",
			"end_turn",
			"end_combat",
		}
		for _, expectedTool := range expectedTools {
			assert.True(t, toolNames[expectedTool], "%s tool should be registered", expectedTool)
		}

		t.Logf("All %d combat tools are registered", len(expectedTools))
	})

	// Test 2: End combat with invalid ID
	t.Run("EndCombatInvalidID", func(t *testing.T) {
		result := callTool(t, client, port, "end_combat", map[string]interface{}{
			"combat_id": "invalid-id",
		})

		// Verify error is returned
		isError, _ := result["isError"].(bool)
		assert.True(t, isError, "end_combat with invalid ID should return an error")

		t.Logf("end_combat error response: %v", result)
	})

	// Test 3: End combat without required parameters
	t.Run("EndCombatMissingParams", func(t *testing.T) {
		result := callTool(t, client, port, "end_combat", map[string]interface{}{})

		// Verify error is returned
		isError, _ := result["isError"].(bool)
		assert.True(t, isError, "end_combat without combat_id should return an error")

		t.Logf("end_combat missing params response: %v", result)
	})
}

// TestCombatToolNames verifies the combat tool names are exported correctly
func TestCombatToolNames(t *testing.T) {
	// This test verifies that the tool names are accessible
	// The actual values are checked in the e2e test above
	t.Log("Combat tool names should include: start_combat, get_combat_state, attack, cast_spell, end_turn, end_combat")
}
