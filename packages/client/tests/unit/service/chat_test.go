// Package service_test 提供 Service 层单元测试
package service_test

import (
	"context"
	"testing"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/mcp"
	"github.com/dnd-mcp/client/internal/server"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
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

	message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "这是一个测试响应", message.Content)

	mockLLMClient.AssertExpectations(t)
}

// TestChatService_SendMessage_LLMError 测试 LLM 错误
func TestChatService_SendMessage_LLMError(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望 - LLM 返回错误
	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(nil, assert.AnError)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "你好",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, message)

	mockLLMClient.AssertExpectations(t)
}

// TestChatService_SendMessage_SaveMessageError 测试保存消息错误
func TestChatService_SendMessage_SaveMessageError(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	// 设置返回错误
	mockServerClient.SetReturnError(true)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 测试 - 应该在获取上下文时失败
	req := &service.SendMessageRequest{
		Content:  "你好",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

	// 断言
	assert.Error(t, err)
	assert.Nil(t, message)
}

// TestChatService_SendMessage_WithContext 测试带上下文的消息发送
func TestChatService_SendMessage_WithContext(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	config := &service.ContextBuilderConfig{
		UseRawContext: true,
		MessageLimit:  10,
		IncludeCombat: true,
	}
	contextBuilder := service.NewContextBuilder(mockServerClient, config)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "响应内容",
				},
				FinishReason: "stop",
			},
		},
	}, nil)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "我想攻击哥布林",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, message)

	mockLLMClient.AssertExpectations(t)
}

// TestChatService_SendMessage_WithToolCalls 测试带工具调用的响应
func TestChatService_SendMessage_WithToolCalls(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望 - LLM 返回带工具调用的响应
	mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call-123",
							Type: "function",
							Function: llm.FunctionCall{
								Name:      "roll_dice",
								Arguments: `{"formula": "1d20+5"}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}, nil)

	// 设置 MCP 工具调用 Mock
	mockMCPClient.On("CallTool", mock.Anything, mock.Anything, "roll_dice", mock.Anything).Return(map[string]interface{}{
		"success": true,
		"result": map[string]interface{}{
			"total": 18,
		},
	}, nil)

	// 测试
	req := &service.SendMessageRequest{
		Content:  "我要投一个 d20",
		PlayerID: "player-123",
	}

	message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

	// 断言
	assert.NoError(t, err)
	assert.NotNil(t, message)

	mockLLMClient.AssertExpectations(t)
}

// TestContextBuilder_BuildContext 测试上下文构建
func TestContextBuilder_BuildContext(t *testing.T) {
	mockServerClient := server.NewMockClient()

	testCases := []struct {
		name      string
		config    *service.ContextBuilderConfig
		expectLen int // 期望至少有多少条消息
	}{
		{
			name:      "简化模式",
			config:    nil, // 使用默认配置
			expectLen: 2,   // system + user
		},
		{
			name: "完整模式",
			config: &service.ContextBuilderConfig{
				UseRawContext: true,
				MessageLimit:  20,
				IncludeCombat: true,
			},
			expectLen: 2,
		},
		{
			name: "不包含战斗",
			config: &service.ContextBuilderConfig{
				UseRawContext: false,
				MessageLimit:  10,
				IncludeCombat: false,
			},
			expectLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contextBuilder := service.NewContextBuilder(mockServerClient, tc.config)

			ctx := context.Background()
			messages, err := contextBuilder.BuildContext(ctx, "campaign-1", "test message")

			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(messages), tc.expectLen)

			// 验证第一条是 system 消息
			if len(messages) > 0 {
				assert.Equal(t, "system", messages[0].Role)
			}
		})
	}
}

// TestContextBuilder_BuildContext_Error 测试上下文构建错误
func TestContextBuilder_BuildContext_Error(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockServerClient.SetReturnError(true)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)

	ctx := context.Background()
	messages, err := contextBuilder.BuildContext(ctx, "campaign-1", "test message")

	assert.Error(t, err)
	assert.Nil(t, messages)
}

// TestContextBuilder_Config 测试配置
func TestContextBuilder_Config(t *testing.T) {
	mockServerClient := server.NewMockClient()

	// 测试默认配置
	cb1 := service.NewContextBuilder(mockServerClient, nil)
	assert.NotNil(t, cb1)

	// 测试自定义配置
	config := &service.ContextBuilderConfig{
		UseRawContext: true,
		MessageLimit:  50,
		IncludeCombat: false,
	}
	cb2 := service.NewContextBuilder(mockServerClient, config)
	assert.NotNil(t, cb2)
}

// TestChatService_MultipleMessages 测试多消息场景
func TestChatService_MultipleMessages(t *testing.T) {
	mockServerClient := server.NewMockClient()
	mockLLMClient := new(MockLLMClient)
	mockMCPClient := new(MockMCPClient)

	contextBuilder := service.NewContextBuilder(mockServerClient, nil)
	chatService := service.NewChatService(mockServerClient, mockLLMClient, mockMCPClient, contextBuilder)

	// 设置 Mock 期望
	for i := 0; i < 3; i++ {
		mockLLMClient.On("Chat", mock.Anything, mock.AnythingOfType("*llm.ChatRequest")).Return(&llm.ChatResponse{
			Choices: []llm.Choice{
				{
					Message: llm.Message{
						Role:    "assistant",
						Content: "响应",
					},
					FinishReason: "stop",
				},
			},
		}, nil).Once()
	}

	// 发送多条消息
	for i := 0; i < 3; i++ {
		req := &service.SendMessageRequest{
			Content:  "消息",
			PlayerID: "player-123",
		}

		message, err := chatService.SendMessage(context.Background(), "campaign-123", req)

		assert.NoError(t, err)
		assert.NotNil(t, message)
	}

	// 验证消息被保存
	messages := mockServerClient.GetMessages()
	assert.GreaterOrEqual(t, len(messages), 3)

	mockLLMClient.AssertExpectations(t)
}
