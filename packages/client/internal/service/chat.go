// Package service 提供业务逻辑层实现
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/mcp"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/google/uuid"
)

// ChatServiceInterface 聊天服务接口
type ChatServiceInterface interface {
	// SendMessage 发送消息并获取 AI 响应
	SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*models.Message, error)
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content  string
	PlayerID string
}

// ChatService 聊天服务实现
type ChatService struct {
	messageStore   store.MessageStore
	sessionStore   store.SessionStore
	llmClient      llm.LLMClient
	mcpClient      mcp.MCPClient
	contextBuilder *ContextBuilder
}

// NewChatService 创建聊天服务
func NewChatService(
	messageStore store.MessageStore,
	sessionStore store.SessionStore,
	llmClient llm.LLMClient,
	mcpClient mcp.MCPClient,
	contextBuilder *ContextBuilder,
) *ChatService {
	return &ChatService{
		messageStore:   messageStore,
		sessionStore:   sessionStore,
		llmClient:      llmClient,
		mcpClient:      mcpClient,
		contextBuilder: contextBuilder,
	}
}

// SendMessage 发送消息并获取 AI 响应
func (s *ChatService) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) (*models.Message, error) {
	// 1. 验证会话是否存在
	_, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("会话不存在: %w", err)
	}

	// 2. 保存用户消息
	userMessage := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
		PlayerID:  req.PlayerID,
		CreatedAt: time.Now(),
	}
	if err := s.messageStore.Create(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 3. 构建对话上下文（包含历史消息）
	messages, err := s.contextBuilder.BuildContext(ctx, sessionID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("构建对话上下文失败: %w", err)
	}

	// 4. 调用 LLM
	llmResp, err := s.llmClient.Chat(ctx, &llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 5. 检查响应类型
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空响应")
	}

	choice := llmResp.Choices[0]

	// 6. 判断是否有 tool_calls
	if choice.FinishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		// 处理工具调用
		return s.handleToolCalls(ctx, sessionID, choice.Message.ToolCalls, req.Content)
	}

	// 7. 保存助手消息（纯文本响应）
	assistantMessage := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
		CreatedAt: time.Now(),
	}
	if err := s.messageStore.Create(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("保存助手消息失败: %w", err)
	}

	return assistantMessage, nil
}

// handleToolCalls 处理工具调用
func (s *ChatService) handleToolCalls(ctx context.Context, sessionID string, toolCalls []llm.ToolCall, originalUserMessage string) (*models.Message, error) {
	// 1. 保存 assistant 消息（包含 tool_calls）
	assistantMsg := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "",
		ToolCalls: convertLLMToolCalls(toolCalls),
		CreatedAt: time.Now(),
	}

	if err := s.messageStore.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("保存工具调用消息失败: %w", err)
	}

	// 2. 执行所有工具调用
	toolResults := make([]map[string]interface{}, len(toolCalls))
	for i, toolCall := range toolCalls {
		// 解析 arguments
		var args map[string]interface{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

		// 调用 MCP Server
		result, err := s.mcpClient.CallTool(ctx, sessionID, toolCall.Function.Name, args)
		if err != nil {
			return nil, fmt.Errorf("工具调用失败: %w", err)
		}

		toolResults[i] = result
	}

	// 3. 保存 tool 响应消息
	for i, toolCall := range toolCalls {
		toolMsg := &models.Message{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "tool",
			Content:   fmt.Sprintf("工具 %s 执行结果: %+v", toolCall.Function.Name, toolResults[i]),
			CreatedAt: time.Now(),
		}

		if err := s.messageStore.Create(ctx, toolMsg); err != nil {
			return nil, fmt.Errorf("保存工具响应失败: %w", err)
		}
	}

	// 4. 重新构建上下文（包含工具调用和结果）
	messages, err := s.contextBuilder.BuildContext(ctx, sessionID, originalUserMessage)
	if err != nil {
		return nil, fmt.Errorf("构建上下文失败: %w", err)
	}

	// 添加 tool_calls 消息
	messages = append(messages, llm.Message{
		Role:      "assistant",
		Content:   "",
		ToolCalls: toolCalls,
	})

	// 添加 tool 响应消息
	for i, result := range toolResults {
		messages = append(messages, llm.Message{
			Role:       "tool",
			Content:    fmt.Sprintf("%+v", result),
			ToolCallID: toolCalls[i].ID,
		})
	}

	// 5. 继续调用 LLM（传递工具结果）
	followupReq := &llm.ChatRequest{
		Model:       "gpt-4",
		Messages:    messages,
		Temperature: 0.7,
	}

	followupResp, err := s.llmClient.Chat(ctx, followupReq)
	if err != nil {
		return nil, fmt.Errorf("LLM 后续调用失败: %w", err)
	}

	// 6. 保存最终助手响应
	if len(followupResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空响应")
	}

	finalMsg := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      followupResp.Choices[0].Message.Role,
		Content:   followupResp.Choices[0].Message.Content,
		CreatedAt: time.Now(),
	}

	if err := s.messageStore.Create(ctx, finalMsg); err != nil {
		return nil, fmt.Errorf("保存最终响应失败: %w", err)
	}

	return finalMsg, nil
}

// convertLLMToolCalls 转换 LLM tool_calls 到模型
func convertLLMToolCalls(llmToolCalls []llm.ToolCall) []models.ToolCall {
	toolCalls := make([]models.ToolCall, len(llmToolCalls))
	for i, tc := range llmToolCalls {
		// 解析 arguments JSON string
		var args map[string]interface{}
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		toolCalls[i] = models.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		}
	}
	return toolCalls
}
