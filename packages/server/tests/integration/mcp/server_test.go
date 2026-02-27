package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *config.Config {
	return &config.Config{
		HTTP: config.HTTPConfig{
			Host:            "0.0.0.0",
			Port:            8081,
			ReadTimeout:     30,
			WriteTimeout:    30,
			ShutdownTimeout: 10,
			EnableCORS:      true,
		},
		Log: config.LogConfig{
			Level:  "debug",
			Format: "text",
		},
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "test",
			PoolSize: 5,
		},
		RAG: config.RAGConfig{
			Timeout: 30,
		},
	}
}

// setupTestServer creates a test server using the real routing configuration
// This ensures tests validate the actual server behavior
func setupTestServer(cfg *config.Config, server *mcp.Server) *httptest.Server {
	// Use the real handler from the server
	handler := server.Handler()
	return httptest.NewServer(handler)
}

func TestServer_Health(t *testing.T) {
	cfg := newTestConfig()
	server := mcp.NewServer(cfg)

	testServer := setupTestServer(cfg, server)
	defer testServer.Close()

	// Test health endpoint
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "healthy", result["status"])
}

func TestServer_Initialize(t *testing.T) {
	cfg := newTestConfig()
	server := mcp.NewServer(cfg)

	testServer := setupTestServer(cfg, server)
	defer testServer.Close()

	// Test initialize endpoint
	initReq := mcp.InitializeRequest{
		ProtocolVersion: "2024-11-05",
		ClientInfo: mcp.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	body, _ := json.Marshal(initReq)

	resp, err := http.Post(testServer.URL+"/mcp/initialize", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result mcp.InitializeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "2024-11-05", result.ProtocolVersion)
	assert.Equal(t, "dnd-mcp-server", result.ServerInfo.Name)
}

func TestServer_ListTools(t *testing.T) {
	cfg := newTestConfig()
	server := mcp.NewServer(cfg)

	// Register a test tool
	tool := mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: mcp.InputSchema{Type: "object"},
	}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse("ok")
	}
	server.RegisterTool(tool, handler)

	testServer := setupTestServer(cfg, server)
	defer testServer.Close()

	// Test list tools endpoint
	resp, err := http.Get(testServer.URL + "/mcp/tools")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result mcp.ListToolsResponse
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Len(t, result.Tools, 1)
	assert.Equal(t, "test_tool", result.Tools[0].Name)
}

func TestServer_CallTool(t *testing.T) {
	cfg := newTestConfig()
	server := mcp.NewServer(cfg)

	// Register a test tool
	tool := mcp.Tool{
		Name:        "echo",
		Description: "Echo tool",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"message": {Type: "string", Description: "Message to echo"},
			},
		},
	}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		var args struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(req.Arguments, &args); err != nil {
			return mcp.NewErrorResponse(err)
		}
		return mcp.NewTextResponse(args.Message)
	}
	server.RegisterTool(tool, handler)

	testServer := setupTestServer(cfg, server)
	defer testServer.Close()

	// Test call tool endpoint
	callReq := mcp.CallToolRequest{
		Name: "echo",
		Arguments: map[string]interface{}{
			"message": "Hello, World!",
		},
	}
	body, _ := json.Marshal(callReq)

	resp, err := http.Post(testServer.URL+"/mcp/tools/call", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result mcp.CallToolResponse
	json.NewDecoder(resp.Body).Decode(&result)
	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)
	assert.Equal(t, "Hello, World!", result.Content[0].Text)
}

func TestServer_CallUnknownTool(t *testing.T) {
	cfg := newTestConfig()
	server := mcp.NewServer(cfg)

	testServer := setupTestServer(cfg, server)
	defer testServer.Close()

	// Test call unknown tool
	callReq := mcp.CallToolRequest{
		Name: "unknown_tool",
	}
	body, _ := json.Marshal(callReq)

	resp, err := http.Post(testServer.URL+"/mcp/tools/call", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result mcp.CallToolResponse
	json.NewDecoder(resp.Body).Decode(&result)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].Text, "unknown tool")
}
