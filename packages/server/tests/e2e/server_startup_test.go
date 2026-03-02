// Package e2e provides end-to-end tests for the DND MCP Server
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
	// Default timeout for server startup
	serverStartupTimeout = 30 * time.Second
	// Default timeout for HTTP requests
	httpRequestTimeout = 10 * time.Second
)

// ServerProcess represents a running server process
type ServerProcess struct {
	cmd    *exec.Cmd
	port   int
	binary string
}

// buildBinary builds the server binary and returns the path
func buildBinary(t *testing.T) string {
	// Get project root directory
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binaryName := "dnd-server-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(projectRoot, "bin", binaryName)

	// Build the binary
	t.Logf("Building server binary: %s", binaryPath)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/server")
	buildCmd.Dir = projectRoot
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build server binary: %s", string(output))

	return binaryPath
}

// startServer starts the server process and waits for it to be ready
func startServer(t *testing.T, binaryPath string, port int) *ServerProcess {
	// Set environment variables for the server
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("HTTP_PORT=%d", port),
		"HTTP_HOST=127.0.0.1",
		"LOG_LEVEL=debug",
		"LOG_FORMAT=text",
	)

	// Start the server process
	cmd := exec.Command(binaryPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.Logf("Starting server on port %d", port)
	err := cmd.Start()
	require.NoError(t, err, "Failed to start server process")

	sp := &ServerProcess{
		cmd:    cmd,
		port:   port,
		binary: binaryPath,
	}

	// Wait for server to be ready
	sp.waitForReady(t)

	return sp
}

// waitForReady waits for the server to respond to health checks
func (sp *ServerProcess) waitForReady(t *testing.T) {
	client := &http.Client{Timeout: httpRequestTimeout}
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", sp.port)

	start := time.Now()
	for time.Since(start) < serverStartupTimeout {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Logf("Server is ready after %v", time.Since(start))
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Server didn't start in time, kill the process
	sp.cmd.Process.Kill()
	t.Fatalf("Server failed to start within %v", serverStartupTimeout)
}

// Stop stops the server process
func (sp *ServerProcess) Stop(t *testing.T) {
	if sp.cmd == nil || sp.cmd.Process == nil {
		return
	}

	t.Log("Stopping server...")
	// Send interrupt signal to allow graceful shutdown
	if err := sp.cmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, kill the process
		sp.cmd.Process.Kill()
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- sp.cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		t.Log("Server didn't stop gracefully, killing...")
		sp.cmd.Process.Kill()
	case err := <-done:
		if err != nil {
			t.Logf("Server exited with error: %v", err)
		} else {
			t.Log("Server stopped gracefully")
		}
	}
}

// TestServerStartup tests that the server can start and respond to requests
// This test requires a PostgreSQL database to be available
func TestServerStartup(t *testing.T) {
	// Check if we should skip E2E tests
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E test (SKIP_E2E_TESTS=true)")
	}

	// Check if database is available
	if os.Getenv("POSTGRES_HOST") == "" {
		// Try default values
		os.Setenv("POSTGRES_HOST", "localhost")
		os.Setenv("POSTGRES_PORT", "5432")
		os.Setenv("POSTGRES_USER", "dnd")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("POSTGRES_DBNAME", "dnd_server")
	}

	// Use a random port to avoid conflicts
	port := 18081

	// Build the binary
	binaryPath := buildBinary(t)
	defer func() {
		// Clean up binary after test
		os.Remove(binaryPath)
	}()

	// Start the server
	server := startServer(t, binaryPath, port)
	defer server.Stop(t)

	// Test 1: Health check
	t.Run("HealthCheck", func(t *testing.T) {
		client := &http.Client{Timeout: httpRequestTimeout}
		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		require.NoError(t, err, "Failed to call health endpoint")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Equal(t, "healthy", result["status"])
		t.Logf("Health check passed: %s", string(body))
	})

	// Test 2: MCP Initialize
	t.Run("MCPInitialize", func(t *testing.T) {
		client := &http.Client{Timeout: httpRequestTimeout}

		initReq := map[string]interface{}{
			"protocol_version": "2024-11-05",
			"client_info": map[string]interface{}{
				"name":    "e2e-test-client",
				"version": "1.0.0",
			},
		}
		reqBody, _ := json.Marshal(initReq)

		resp, err := client.Post(
			fmt.Sprintf("http://127.0.0.1:%d/mcp/initialize", port),
			"application/json",
			bytes.NewReader(reqBody),
		)
		require.NoError(t, err, "Failed to call initialize endpoint")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		assert.Equal(t, "2024-11-05", result["protocolVersion"])
		serverInfo, ok := result["serverInfo"].(map[string]interface{})
		require.True(t, ok, "serverInfo not found in response")
		assert.Equal(t, "dnd-mcp-server", serverInfo["name"])
		t.Logf("MCP Initialize passed: %s", string(body))
	})

	// Test 3: List Tools
	t.Run("ListTools", func(t *testing.T) {
		client := &http.Client{Timeout: httpRequestTimeout}

		resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/mcp/tools", port))
		require.NoError(t, err, "Failed to call tools endpoint")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		require.NoError(t, err)

		// M1 has no tools implemented yet, so tools should be empty
		tools, ok := result["tools"].([]interface{})
		require.True(t, ok, "tools not found in response")
		// M1 has no tools, so this should be empty
		t.Logf("List Tools passed: %d tools registered", len(tools))
	})
}
