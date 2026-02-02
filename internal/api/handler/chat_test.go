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
	response    string
	toolCalls   []llm.ToolCall
	shouldError bool
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	if m.shouldError {
		return nil, &LLMError{Message: "LLM error"}
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

// getTestDatabaseURL 获取测试数据库URL
func getTestDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "070831"
	}
	return fmt.Sprintf("postgres://postgres:%s@localhost:5432/dnd_mcp_test?sslmode=disable", password)
}

// createTestSession 创建测试会话
func createTestSession(t *testing.T, db *sql.DB, ctx context.Context) string {
	sessionID := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO sessions (id, version, created_at, updated_at, campaign_name, location, state)
		VALUES ($1, 1, $2, $3, $4, $5, $6)
	`, sessionID, time.Now(), time.Now(), "测试战役", "测试地点", "{}")
	require.NoError(t, err)
	return sessionID
}

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB) {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, "DELETE FROM messages WHERE session_id IN (SELECT id FROM sessions WHERE campaign_name = '测试战役')")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM sessions WHERE campaign_name = '测试战役'")
	require.NoError(t, err)
}

// TestChatHandler_ChatMessage_Success 测试成功的聊天消息
func TestChatHandler_ChatMessage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	cleanupTestData(t, db)

	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	mockLLM := &MockLLMClient{
		response: "你好，冒险者！有什么可以帮你的吗？",
	}

	handler := NewChatHandler(mockLLM, dataStore)

	sessionID := createTestSession(t, db, ctx)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message": "你好，地下城主",
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

	assert.NotEmpty(t, resp["response"])
	assert.NotEmpty(t, resp["usage"])
}

// TestChatHandler_ChatMessage_SessionNotFound 测试会话不存在
func TestChatHandler_ChatMessage_SessionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	databaseURL := getTestDatabaseURL()
	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	mockLLM := &MockLLMClient{
		response: "test",
	}

	handler := NewChatHandler(mockLLM, dataStore)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message": "test",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/"+uuid.New().String()+"/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestChatHandler_ChatMessage_InvalidUUID 测试无效UUID
func TestChatHandler_ChatMessage_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	databaseURL := getTestDatabaseURL()
	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	mockLLM := &MockLLMClient{
		response: "test",
	}

	handler := NewChatHandler(mockLLM, dataStore)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

	reqBody := map[string]string{
		"message": "test",
	}
	bodyJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/sessions/invalid-uuid/chat", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestChatHandler_ChatMessage_MissingMessage 测试缺少消息
func TestChatHandler_ChatMessage_MissingMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	cleanupTestData(t, db)

	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	mockLLM := &MockLLMClient{
		response: "test",
	}

	handler := NewChatHandler(mockLLM, dataStore)

	sessionID := createTestSession(t, db, ctx)

	router := gin.New()
	router.POST("/api/sessions/:id/chat", handler.ChatMessage)

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

	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	cleanupTestData(t, db)

	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

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

	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	cleanupTestData(t, db)

	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err)
	defer dataStore.Close()

	mockLLM := &MockLLMClient{
		response: "你好，玩家！",
	}

	handler := NewChatHandler(mockLLM, dataStore)

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
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages WHERE session_id = $1", sessionID).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)

	// 验证player_id已保存
	var playerID sql.NullString
	query := `SELECT player_id FROM messages WHERE session_id = $1 AND role = 'user' ORDER BY created_at DESC LIMIT 1`
	err = db.QueryRowContext(ctx, query, sessionID).Scan(&playerID)
	require.NoError(t, err)

	require.True(t, playerID.Valid, "player_id should not be NULL")
	assert.Equal(t, "player-001", playerID.String, "player_id should match the provided value")
}
