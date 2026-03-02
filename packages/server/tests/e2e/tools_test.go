// Package e2e provides end-to-end tests for the DND MCP Server tools
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
	toolsTestStartupTimeout = 30 * time.Second
	toolsTestRequestTimeout = 10 * time.Second
)

// buildToolsTestBinary builds the server binary and returns the path
func buildToolsTestBinary(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binaryName := "dnd-server-tools-test"
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

// startToolsTestServer starts the server process and waits for it to be ready
func startToolsTestServer(t *testing.T, binaryPath string, port int) *exec.Cmd {
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
	client := &http.Client{Timeout: toolsTestRequestTimeout}
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)

	start := time.Now()
	for time.Since(start) < toolsTestStartupTimeout {
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
	t.Fatalf("Server failed to start within %v", toolsTestStartupTimeout)
	return nil
}

// stopToolsTestServer stops the server process
func stopToolsTestServer(t *testing.T, cmd *exec.Cmd) {
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

// TestDiceToolsE2E tests that dice tools are registered and functional
// This test requires a PostgreSQL database to be available
func TestDiceToolsE2E(t *testing.T) {
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

	port := 18082

	// Build the binary
	binaryPath := buildToolsTestBinary(t)
	defer os.Remove(binaryPath)

	// Start the server
	cmd := startToolsTestServer(t, binaryPath, port)
	defer stopToolsTestServer(t, cmd)

	client := &http.Client{Timeout: toolsTestRequestTimeout}

	// Test 1: Verify dice tools are registered
	t.Run("DiceToolsRegistered", func(t *testing.T) {
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

		// Verify all three dice tools are registered
		assert.True(t, toolNames["roll_dice"], "roll_dice tool should be registered")
		assert.True(t, toolNames["roll_check"], "roll_check tool should be registered")
		assert.True(t, toolNames["roll_save"], "roll_save tool should be registered")

		t.Logf("All 3 dice tools are registered: roll_dice, roll_check, roll_save")
	})

	// Test 2: Call roll_dice tool
	t.Run("RollDice", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "roll_dice",
			"arguments": map[string]interface{}{
				"formula": "1d20+5",
			},
		}
		reqJSON, _ := json.Marshal(reqBody)

		resp, err := client.Post(
			fmt.Sprintf("http://127.0.0.1:%d/mcp/tools/call", port),
			"application/json",
			bytes.NewReader(reqJSON),
		)
		require.NoError(t, err, "Failed to call roll_dice tool")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		// Verify no error
		isError, _ := result["isError"].(bool)
		assert.False(t, isError, "roll_dice should not return an error")

		// Verify content exists
		content, ok := result["content"].([]interface{})
		require.True(t, ok, "content not found in response")
		require.Len(t, content, 1, "should have one content item")

		t.Logf("roll_dice response: %s", string(body))
	})

	// Test 3: Call roll_dice with invalid formula
	t.Run("RollDiceInvalidFormula", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "roll_dice",
			"arguments": map[string]interface{}{
				"formula": "invalid",
			},
		}
		reqJSON, _ := json.Marshal(reqBody)

		resp, err := client.Post(
			fmt.Sprintf("http://127.0.0.1:%d/mcp/tools/call", port),
			"application/json",
			bytes.NewReader(reqJSON),
		)
		require.NoError(t, err, "Failed to call roll_dice tool")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		// Verify error is returned
		isError, _ := result["isError"].(bool)
		assert.True(t, isError, "roll_dice with invalid formula should return an error")

		t.Logf("roll_dice error response: %s", string(body))
	})

	// Test 4: Verify unknown tool returns error
	t.Run("UnknownTool", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "unknown_tool",
			"arguments": map[string]interface{}{},
		}
		reqJSON, _ := json.Marshal(reqBody)

		resp, err := client.Post(
			fmt.Sprintf("http://127.0.0.1:%d/mcp/tools/call", port),
			"application/json",
			bytes.NewReader(reqJSON),
		)
		require.NoError(t, err, "Failed to call tools endpoint")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		// Verify error is returned
		isError, _ := result["isError"].(bool)
		assert.True(t, isError, "unknown tool should return an error")

		t.Logf("Unknown tool error response: %s", string(body))
	})
}
