package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/client/llm"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMClient mock LLM客户端用于测试
type MockLLMClient struct {
	response     string
	toolCalls    []llm.ToolCall
	shouldError  bool
	errorMessage string
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	if m.shouldError {
		return nil, &LLMError{Message: m.errorMessage}
	}

	return &llm.ChatCompletionResponse{
		ID:     "test-response-001",
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:      "assistant",
					Content:   m.response,
					ToolCalls: m.toolCalls,
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     50,
			CompletionTokens: 30,
			TotalTokens:      80,
		},
	}, nil
}

func (m *MockLLMClient) StreamCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (<-chan llm.StreamChunk, error) {
	// 返回一个空 channel，表示流式响应未实现
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

// LLMError 自定义LLM错误
type LLMError struct {
	Message string
}

func (e *LLMError) Error() string {
	return e.Message
}

// setupTestChatHandler 设置测试chat handler
func setupTestChatHandler(t *testing.T) (*ChatHandler, *sql.DB, func()) {
	// 使用测试数据库
	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)

	// 运行迁移
	runMigrations(t, db)

	// 创建store
	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)

	// 创建mock LLM客户端
	mockLLM := &MockLLMClient{
		response: "你好，冒险者！有什么可以帮你的吗？",
	}

	handler := NewChatHandler(mockLLM, dataStore)

	cleanup := func() {
		dataStore.Close()
		db.Close()
	}

	return handler, db, cleanup
}

// TestChatHandler_ChatMessage_Success 测试成功的聊天消息
func TestChatHandler_ChatMessage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, cleanup := setupTestChatHandler(t)
	defer cleanup()

	// 创建测试会话
	ctx := context.Background()
	sessionID := createTestSession(t, db, ctx)

	// 创建测试路由
	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	// 构造请求
	reqBody := map[string]string{
		"message": "你好，地下城主",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID+"/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	// 记录响应
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["response"])
	assert.NotNil(t, resp["usage"])
}

// TestChatHandler_ChatMessage_SessionNotFound 测试会话不存在的错误
func TestChatHandler_ChatMessage_SessionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, cleanup := setupTestChatHandler(t)
	defer cleanup()

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	// 使用不存在的会话ID
	reqBody := map[string]string{
		"message": "测试消息",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/550e8400-e29b-41d4-a716-446655449999/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "session not found", resp["error"])
}

// TestChatHandler_ChatMessage_InvalidUUID 测试无效的UUID
func TestChatHandler_ChatMessage_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _, cleanup := setupTestChatHandler(t)
	defer cleanup()

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message": "测试消息",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/invalid-uuid/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "invalid session id", resp["error"])
}

// TestChatHandler_ChatMessage_MissingMessage 测试缺少消息体
func TestChatHandler_ChatMessage_MissingMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, cleanup := setupTestChatHandler(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := createTestSession(t, db, ctx)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	// 空请求体
	reqBody := map[string]string{}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID+"/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestChatHandler_ChatMessage_ToolCalls 测试工具调用响应
func TestChatHandler_ChatMessage_ToolCalls(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 使用测试数据库
	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	// 运行迁移
	runMigrations(t, db)

	// 创建store
	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	// 创建返回工具调用的mock LLM客户端
	mockLLM := &MockLLMClient{
		response: "",
		toolCalls: []llm.ToolCall{
			{
				ID:   "call_001",
				Type: "function",
				Function: llm.FunctionCall{
					Name:      "roll_dice",
					Arguments: `{"dice_type":"d20","modifier":5}`,
				},
			},
		},
	}

	handler := NewChatHandler(mockLLM, dataStore)

	ctx := context.Background()
	sessionID := createTestSession(t, db, ctx)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message": "我要投骰子",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID+"/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["tool_calls"])
}

// TestChatHandler_ChatMessage_PlayerID 测试带玩家ID的消息
func TestChatHandler_ChatMessage_PlayerID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, cleanup := setupTestChatHandler(t)
	defer cleanup()

	ctx := context.Background()
	sessionID := createTestSession(t, db, ctx)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message":  "我要攻击",
		"player_id": "player-001",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID+"/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 验证消息已保存到数据库
	var playerID string
	query := "SELECT player_id FROM messages WHERE session_id = $1 ORDER BY created_at DESC LIMIT 1"
	err := db.QueryRowContext(ctx, query, sessionID).Scan(&playerID)
	require.NoError(t, err)
	assert.Equal(t, "player-001", playerID)
}

// Helper functions

func getTestDatabaseURL() string {
	// 从环境变量读取数据库密码，如果没有则使用默认值
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "070831" // 默认密码
	}
	return fmt.Sprintf("postgres://postgres:%s@localhost:5432/dnd_mcp_test?sslmode=disable", password)
}

func runMigrations(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			game_time VARCHAR(50),
			location VARCHAR(255),
			campaign_name VARCHAR(255) NOT NULL,
			state JSONB,
			deleted_at TIMESTAMP
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY,
			session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			player_id UUID,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)
}

func createTestSession(t *testing.T, db *sql.DB, ctx context.Context) string {
	sessionID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO sessions (id, version, created_at, updated_at, campaign_name, location, state)
		VALUES ($1, 1, $2, $3, $4, $5, $6)
	`, sessionID, time.Now(), time.Now(), "测试战役", "测试地点", "{}")
	require.NoError(t, err)
	return sessionID
}
