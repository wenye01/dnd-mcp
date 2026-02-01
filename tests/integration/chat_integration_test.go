package integration

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

	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/client/llm"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/testcontainers/testcontainers-go"
)

// ChatIntegrationTest 集成测试套件
type ChatIntegrationTest struct {
	db            *sql.DB
	dataStore     store.Store
	llmClient     llm.Client
	router        *gin.Engine
	server        *httptest.Server
	serverURL     string
	cleanup       func()
}

// SetupChatIntegrationTest 设置聊天集成测试
func SetupChatIntegrationTest(t *testing.T) *ChatIntegrationTest {
	gin.SetMode(gin.TestMode)

	// 设置测试数据库
	databaseURL := getTestDatabaseURL()
	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err, "Failed to connect to test database")

	// 运行迁移
	runTestMigrations(t, db)

	// 创建store
	dataStore, err := store.NewPostgresStore(databaseURL)
	require.NoError(t, err, "Failed to create store")

	// 创建mock LLM客户端（可以配置为使用真实的mock服务器）
	mockLLM := &MockLLMClientForIntegration{
		responses: make(map[string]string),
	}

	// 创建chat handler
	chatHandler := handler.NewChatHandler(mockLLM, dataStore)

	// 设置路由
	router := gin.New()
	router.POST("/api/sessions/:id/chat", chatHandler.ChatMessage)
	router.GET("/api/sessions", handler.NewSessionHandler(dataStore).ListSessions)
	router.POST("/api/sessions", handler.NewSessionHandler(dataStore).CreateSession)
	router.GET("/api/sessions/:id", handler.NewSessionHandler(dataStore).GetSession)

	// 创建测试服务器
	server := httptest.NewServer(router)

	cleanup := func() {
		server.Close()
		dataStore.Close()
		db.Close()
	}

	return &ChatIntegrationTest{
		db:        db,
		dataStore: dataStore,
		llmClient: mockLLM,
		router:    router,
		server:    server,
		serverURL: server.URL,
		cleanup:   cleanup,
	}
}

// TestChatIntegration_SimpleConversation 测试简单对话流程
func TestChatIntegration_SimpleConversation(t *testing.T) {
	test := SetupChatIntegrationTest(t)
	defer test.cleanup()

	// 设置mock响应
	mockLLM := test.llmClient.(*MockLLMClientForIntegration)
	mockLLM.SetResponse("你好，勇敢的冒险者！欢迎来到地下城。")

	// 1. 创建会话
	sessionID := test.createSession(t, "被遗忘的国度")

	// 2. 发送聊天消息
	resp := test.sendMessage(t, sessionID, "你好，地下城主")

	// 3. 验证响应
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.NotEmpty(t, result["response"])
	assert.NotNil(t, result["usage"])
	assert.Equal(t, "你好，勇敢的冒险者！欢迎来到地下城。", result["response"])

	// 4. 验证消息已保存到数据库
	messages := test.getMessages(t, sessionID)
	assert.Len(t, messages, 2) // 用户消息 + 助手响应
}

// TestChatIntegration_MultiTurnConversation 测试多轮对话
func TestChatIntegration_MultiTurnConversation(t *testing.T) {
	test := SetupChatIntegrationTest(t)
	defer test.cleanup()

	mockLLM := test.llmClient.(*MockLLMClientForIntegration)

	sessionID := test.createSession(t, "多轮对话测试")

	// 第一轮
	mockLLM.SetResponse("欢迎！你要做什么？")
	test.sendMessage(t, sessionID, "你好")
	assert.Len(t, test.getMessages(t, sessionID), 2)

	// 第二轮
	mockLLM.SetResponse("很好！继续探索吧。")
	test.sendMessage(t, sessionID, "我要探索地下城")
	assert.Len(t, test.getMessages(t, sessionID), 4)

	// 第三轮
	mockLLM.SetResponse("祝你好运！")
	test.sendMessage(t, sessionID, "再见")
	assert.Len(t, test.getMessages(t, sessionID), 6)
}

// TestChatIntegration_SessionNotFound 测试不存在的会话
func TestChatIntegration_SessionNotFound(t *testing.T) {
	test := SetupChatIntegrationTest(t)
	defer test.cleanup()

	// 使用不存在的UUID
	fakeSessionID := "550e8400-e29b-41d4-a716-446655449999"
	resp := test.sendMessage(t, fakeSessionID, "测试消息")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "session not found", result["error"])
}

// TestChatIntegration_MultipleSessions 测试多个会话
func TestChatIntegration_MultipleSessions(t *testing.T) {
	test := SetupChatIntegrationTest(t)
	defer test.cleanup()

	mockLLM := test.llmClient.(*MockLLMClientForIntegration)

	// 创建多个会话
	session1 := test.createSession(t, "会话1")
	session2 := test.createSession(t, "会话2")
	session3 := test.createSession(t, "会话3")

	// 向不同会话发送消息
	mockLLM.SetResponse("响应1")
	test.sendMessage(t, session1, "消息1")

	mockLLM.SetResponse("响应2")
	test.sendMessage(t, session2, "消息2")

	mockLLM.SetResponse("响应3")
	test.sendMessage(t, session3, "消息3")

	// 验证每个会话的消息数
	assert.Len(t, test.getMessages(t, session1), 2)
	assert.Len(t, test.getMessages(t, session2), 2)
	assert.Len(t, test.getMessages(t, session3), 2)
}

// TestChatIntegration_ConcurrentMessages 测试并发消息
func TestChatIntegration_ConcurrentMessages(t *testing.T) {
	test := SetupChatIntegrationTest(t)
	defer test.cleanup()

	mockLLM := test.llmClient.(*MockLLMClientForIntegration)
	mockLLM.SetResponse("收到")

	sessionID := test.createSession(t, "并发测试")

	// 并发发送多条消息
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			test.sendMessage(t, sessionID, fmt.Sprintf("并发消息%d", index))
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证所有消息都已保存
	messages := test.getMessages(t, sessionID)
	assert.Len(t, messages, concurrency*2) // 每条消息包括用户和助手响应
}

// MockLLMClientForIntegration 用于集成测试的Mock LLM客户端
type MockLLMClientForIntegration struct {
	responses map[string]string
	defaultResponse string
}

func (m *MockLLMClientForIntegration) SetResponse(response string) {
	m.defaultResponse = response
}

func (m *MockLLMClientForIntegration) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	// 返回设置的响应
	response := m.defaultResponse
	if response == "" {
		response = "默认响应"
	}

	return &llm.ChatCompletionResponse{
		ID:     fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object: "chat.completion",
		Model:  req.Model,
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: response,
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

func (m *MockLLMClientForIntegration) StreamCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (<-chan llm.StreamChunk, error) {
	// 返回一个空 channel，表示流式响应未实现
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

// Helper methods for ChatIntegrationTest

func (test *ChatIntegrationTest) createSession(t *testing.T, campaignName string) string {
	ctx := context.Background()
	sessionID := uuid.New().String()

	_, err := test.db.ExecContext(ctx, `
		INSERT INTO sessions (id, version, created_at, updated_at, campaign_name, location, state)
		VALUES ($1, 1, $2, $3, $4, $5, $6)
	`, sessionID, time.Now(), time.Now(), campaignName, "测试地点", "{}")
	require.NoError(t, err)

	return sessionID
}

func (test *ChatIntegrationTest) sendMessage(t *testing.T, sessionID, message string) *http.Response {
	reqBody := map[string]string{
		"message": message,
	}
	bodyJSON, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/sessions/%s/chat", test.serverURL, sessionID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyJSON))
	require.NoError(t, err)

	return resp
}

func (test *ChatIntegrationTest) getMessages(t *testing.T, sessionID string) []map[string]interface{} {
	ctx := context.Background()
	query := `
		SELECT id, role, content, player_id, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := test.db.QueryContext(ctx, query, sessionID)
	require.NoError(t, err)
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, role, content string
		var playerID *string
		var createdAt time.Time

		err = rows.Scan(&id, &role, &content, &playerID, &createdAt)
		require.NoError(t, err)

		msg := map[string]interface{}{
			"id":         id,
			"role":       role,
			"content":    content,
			"player_id":  playerID,
			"created_at": createdAt,
		}
		messages = append(messages, msg)
	}

	require.NoError(t, rows.Err())
	return messages
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

func runTestMigrations(t *testing.T, db *sql.DB) {
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
