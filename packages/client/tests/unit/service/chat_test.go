// Package service_test 提供 Service 层单元测试
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/mcp"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageStore Mock MessageStore 接口
type MockMessageStore struct {
	mock.Mock
}

func (m *MockMessageStore) Create(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageStore) Get(ctx context.Context, sessionID, messageID string) (*models.Message, error) {
	args := m.Called(ctx, sessionID, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageStore) List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageStore) ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, sessionID, role, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageStore) ListSince(ctx context.Context, sessionID string, since time.Time, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, sessionID, since, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

// MockSessionStoreForChat Mock SessionStore 接口（用于 ChatService 测试）
type MockSessionStoreForChat struct {
	mock.Mock
}

func (m *MockSessionStoreForChat) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionStoreForChat) Get(ctx context.Context, id string) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionStoreForChat) List(ctx context.Context) ([]*models.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionStoreForChat) Update(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionStoreForChat) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionStoreForChat) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockLLMClient Mock LLM 客户端接口
type MockLLMClient struct {
	mock.Mock
}

func (m *MockLLMClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.ChatResponse), args.Error(1)
}

// MockMCPClient Mock MCP 客户端接口
type MockMCPClient struct {
	mock.Mock
}

func (m *MockMCPClient) Initialize(ctx context.Context, sessionID, serverURL string) error {
	args := m.Called(ctx, sessionID, serverURL)
	return args.Error(0)
}

func (m *MockMCPClient) CallTool(ctx context.Context, sessionID string, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, sessionID, toolName, arguments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockMCPClient) SubscribeEvents(ctx context.Context, sessionID string, eventTypes []string) (<-chan mcp.Event, error) {
	args := m.Called(ctx, sessionID, eventTypes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan mcp.Event), args.Error(1)
}

func (m *MockMCPClient) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestChatService_SendMessage_Success 测试成功发送消息
func TestChatService_SendMessage_Success(t *testing.T) {
	// 创建 Mock
	mockSessionStore := new(MockSessionStoreForChat)
	mockMessageStore := new(MockMessageStore)
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockMessageStore, mockSessionStore)
	chatService := service.NewChatService(mockMessageStore, mockSessionStore, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	existingSession := &models.Session{
		ID:        "session-123",
		Name:      "测试会话",
		Status:    "active",
		CreatorID: "user-123",
	}
	mockSessionStore.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	mockMessageStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).Return(nil).Times(2)
	mockMessageStore.On("List", mock.Anything, "session-123", 50).Return([]*models.Message{}, nil)

	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "这是一个测试响应",
				},
				FinishReason: "stop",
			},
		},
	}, nil)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "你好",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "session-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "这是一个测试响应", message.Content)

	// 验证 Mock 调用
	mockSessionStore.AssertExpectations(t)
	mockMessageStore.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

// TestChatService_SendMessage_SessionNotFound 测试会话不存在
func TestChatService_SendMessage_SessionNotFound(t *testing.T) {
	// 创建 Mock
	mockSessionStore := new(MockSessionStoreForChat)
	mockMessageStore := new(MockMessageStore)
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockMessageStore, mockSessionStore)
	chatService := service.NewChatService(mockMessageStore, mockSessionStore, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	mockSessionStore.On("Get", mock.Anything, "non-existent").Return(nil, errors.ErrSessionNotFound)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "你好",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "non-existent", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Contains(t, err.Error(), "会话不存在")

	// 验证 Mock 调用
	mockSessionStore.AssertExpectations(t)
}

// TestChatService_SendMessage_LLMError 测试 LLM 调用失败
func TestChatService_SendMessage_LLMError(t *testing.T) {
	// 创建 Mock
	mockSessionStore := new(MockSessionStoreForChat)
	mockMessageStore := new(MockMessageStore)
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockMessageStore, mockSessionStore)
	chatService := service.NewChatService(mockMessageStore, mockSessionStore, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	existingSession := &models.Session{
		ID:        "session-123",
		Name:      "测试会话",
		Status:    "active",
		CreatorID: "user-123",
	}
	mockSessionStore.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	mockMessageStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).Return(nil).Once()
	mockMessageStore.On("List", mock.Anything, "session-123", 50).Return([]*models.Message{}, nil)

	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(nil, assert.AnError)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "你好",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "session-123", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Contains(t, err.Error(), "LLM 调用失败")

	// 验证 Mock 调用
	mockSessionStore.AssertExpectations(t)
	mockMessageStore.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
}

// TestChatService_SendMessage_WithToolCalls 测试带工具调用的消息
func TestChatService_SendMessage_WithToolCalls(t *testing.T) {
	// 创建 Mock
	mockSessionStore := new(MockSessionStoreForChat)
	mockMessageStore := new(MockMessageStore)
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockMessageStore, mockSessionStore)
	chatService := service.NewChatService(mockMessageStore, mockSessionStore, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	existingSession := &models.Session{
		ID:        "session-123",
		Name:      "测试会话",
		Status:    "active",
		CreatorID: "user-123",
	}
	mockSessionStore.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	// 第一次调用：创建用户消息
	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "user"
	})).Return(nil).Once()

	// 获取历史消息
	mockMessageStore.On("List", mock.Anything, "session-123", 50).Return([]*models.Message{}, nil)

	// 第一次 LLM 调用返回工具调用
	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_001",
							Type: "function",
							Function: llm.FunctionCall{
								Name:      "roll_dice",
								Arguments: `{"formula":"1d20+5"}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}, nil).Once()

	// 创建 assistant 消息（包含 tool_calls）
	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "assistant" && len(m.ToolCalls) > 0
	})).Return(nil).Once()

	// MCP 工具调用
	mockMCPClient.On("CallTool", mock.Anything, "session-123", "roll_dice", mock.Anything).Return(map[string]interface{}{
		"success": true,
		"result": map[string]interface{}{
			"total": 18,
		},
	}, nil).Once()

	// 创建 tool 响应消息
	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "tool"
	})).Return(nil).Once()

	// 再次获取历史消息
	mockMessageStore.On("List", mock.Anything, "session-123", 50).Return([]*models.Message{}, nil)

	// 第二次 LLM 调用（后续请求）
	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "投掷完成！结果是 18。",
				},
				FinishReason: "stop",
			},
		},
	}, nil).Once()

	// 创建最终响应消息
	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "assistant" && m.Content == "投掷完成！结果是 18。"
	})).Return(nil).Once()

	// 测试
	req := &service.SendMessageRequest{
		Content:  "投掷 d20",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "session-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "投掷完成！结果是 18。", message.Content)

	// 验证 Mock 调用
	mockSessionStore.AssertExpectations(t)
	mockMessageStore.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
	mockMCPClient.AssertExpectations(t)
}

// TestChatService_SendMessage_ToolCallError 测试工具调用失败
func TestChatService_SendMessage_ToolCallError(t *testing.T) {
	// 创建 Mock
	mockSessionStore := new(MockSessionStoreForChat)
	mockMessageStore := new(MockMessageStore)
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockMessageStore, mockSessionStore)
	chatService := service.NewChatService(mockMessageStore, mockSessionStore, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	existingSession := &models.Session{
		ID:        "session-123",
		Name:      "测试会话",
		Status:    "active",
		CreatorID: "user-123",
	}
	mockSessionStore.On("Get", mock.Anything, "session-123").Return(existingSession, nil)

	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "user"
	})).Return(nil).Once()

	mockMessageStore.On("List", mock.Anything, "session-123", 50).Return([]*models.Message{}, nil)

	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_001",
							Type: "function",
							Function: llm.FunctionCall{
								Name:      "roll_dice",
								Arguments: `{"formula":"1d20+5"}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}, nil).Once()

	mockMessageStore.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Message) bool {
		return m.Role == "assistant" && len(m.ToolCalls) > 0
	})).Return(nil).Once()

	// MCP 工具调用失败
	mockMCPClient.On("CallTool", mock.Anything, "session-123", "roll_dice", mock.Anything).Return(nil, assert.AnError).Once()

	// 测试
	req := &service.SendMessageRequest{
		Content:  "投掷 d20",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "session-123", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Contains(t, err.Error(), "工具调用失败")

	// 验证 Mock 调用
	mockSessionStore.AssertExpectations(t)
	mockMessageStore.AssertExpectations(t)
	mockLLMClient.AssertExpectations(t)
	mockMCPClient.AssertExpectations(t)
}
