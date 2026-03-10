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
	"github.com/dnd-mcp/client/internal/server"
	"github.com/google/uuid"
)

// 注意: ChatService 现在通过 ServerClient 与 Server 通信
// 不再直接依赖本地存储

// ChatServiceInterface 聊天服务接口
type ChatServiceInterface interface {
	// SendMessage 发送消息并获取 AI 响应
	SendMessage(ctx context.Context, campaignID string, req *SendMessageRequest) (*models.Message, error)
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content  string
	PlayerID string
}

// ChatService 聊天服务实现
// 通过 Server API 保存消息，而非本地存储
type ChatService struct {
	serverClient   server.ServerClient
	llmClient      llm.LLMClient
	mcpClient      mcp.MCPClient
	contextBuilder *ContextBuilder
}

// NewChatService 创建聊天服务
func NewChatService(
	serverClient server.ServerClient,
	llmClient llm.LLMClient,
	mcpClient mcp.MCPClient,
	contextBuilder *ContextBuilder,
) *ChatService {
	return &ChatService{
		serverClient:   serverClient,
		llmClient:      llmClient,
		mcpClient:      mcpClient,
		contextBuilder: contextBuilder,
	}
}

// SendMessage 发送消息并获取 AI 响应
func (s *ChatService) SendMessage(ctx context.Context, campaignID string, req *SendMessageRequest) (*models.Message, error) {
	// 1. 保存用户消息到 Server
	userMsg := &server.Message{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Role:       server.MessageRoleUser,
		Content:    req.Content,
		PlayerID:   req.PlayerID,
		CreatedAt:  time.Now(),
	}
	if err := s.serverClient.SaveMessage(ctx, campaignID, userMsg); err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 2. 构建对话上下文（从 Server 获取）
	messages, err := s.contextBuilder.BuildContext(ctx, campaignID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("构建对话上下文失败: %w", err)
	}

	// 3. 调用 LLM
	llmResp, err := s.llmClient.Chat(ctx, &llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	// 4. 检查响应类型
	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空响应")
	}

	choice := llmResp.Choices[0]

	// 5. 判断是否有 tool_calls
	if choice.FinishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		// 处理工具调用
		return s.handleToolCalls(ctx, campaignID, choice.Message.ToolCalls, req.Content)
	}

	// 6. 保存助手消息到 Server
	assistantMsg := &server.Message{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Role:       server.MessageRoleAssistant,
		Content:    choice.Message.Content,
		CreatedAt:  time.Now(),
	}
	if err := s.serverClient.SaveMessage(ctx, campaignID, assistantMsg); err != nil {
		return nil, fmt.Errorf("保存助手消息失败: %w", err)
	}

	// 转换为本地 models.Message 以保持接口兼容
	return &models.Message{
		ID:        assistantMsg.ID,
		SessionID: campaignID,
		Role:      string(assistantMsg.Role),
		Content:   assistantMsg.Content,
		CreatedAt: assistantMsg.CreatedAt,
	}, nil
}

// handleToolCalls 处理工具调用
func (s *ChatService) handleToolCalls(ctx context.Context, campaignID string, toolCalls []llm.ToolCall, originalUserMessage string) (*models.Message, error) {
	// 1. 保存 assistant 消息（包含 tool_calls）到 Server
	assistantMsg := &server.Message{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Role:       server.MessageRoleAssistant,
		Content:    "",
		ToolCalls:  convertLLMToolCallsToServer(toolCalls),
		CreatedAt:  time.Now(),
	}

	if err := s.serverClient.SaveMessage(ctx, campaignID, assistantMsg); err != nil {
		return nil, fmt.Errorf("保存工具调用消息失败: %w", err)
	}

	// 2. 执行所有工具调用
	toolResults := make([]map[string]interface{}, len(toolCalls))
	for i, toolCall := range toolCalls {
		// 解析 arguments
		var args map[string]interface{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

		// 通过 ServerClient 调用工具
		result, err := s.serverClient.CallTool(ctx, campaignID, toolCall.Function.Name, args)
		if err != nil {
			// 尝试使用 MCP 客户端作为备用
			result, err = s.mcpClient.CallTool(ctx, campaignID, toolCall.Function.Name, args)
			if err != nil {
				return nil, fmt.Errorf("工具调用失败: %w", err)
			}
		}

		toolResults[i] = result
	}

	// 3. 保存 tool 响应消息到 Server
	for i, toolCall := range toolCalls {
		toolMsg := &server.Message{
			ID:         uuid.New().String(),
			CampaignID: campaignID,
			Role:       "tool",
			Content:    fmt.Sprintf("工具 %s 执行结果: %+v", toolCall.Function.Name, toolResults[i]),
			CreatedAt:  time.Now(),
		}

		if err := s.serverClient.SaveMessage(ctx, campaignID, toolMsg); err != nil {
			return nil, fmt.Errorf("保存工具响应失败: %w", err)
		}
	}

	// 4. 重新构建上下文（包含工具调用和结果）
	messages, err := s.contextBuilder.BuildContext(ctx, campaignID, originalUserMessage)
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

	// 6. 保存最终助手响应到 Server
	if len(followupResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空响应")
	}

	finalMsg := &server.Message{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Role:       server.MessageRoleAssistant,
		Content:    followupResp.Choices[0].Message.Content,
		CreatedAt:  time.Now(),
	}

	if err := s.serverClient.SaveMessage(ctx, campaignID, finalMsg); err != nil {
		return nil, fmt.Errorf("保存最终响应失败: %w", err)
	}

	// 转换为本地 models.Message 以保持接口兼容
	return &models.Message{
		ID:        finalMsg.ID,
		SessionID: campaignID,
		Role:      string(finalMsg.Role),
		Content:   finalMsg.Content,
		CreatedAt: finalMsg.CreatedAt,
	}, nil
}

// convertLLMToolCallsToServer 转换 LLM tool_calls 到 Server 格式
func convertLLMToolCallsToServer(llmToolCalls []llm.ToolCall) []server.ToolCall {
	toolCalls := make([]server.ToolCall, len(llmToolCalls))
	for i, tc := range llmToolCalls {
		// 解析 arguments JSON string
		var args map[string]interface{}
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		toolCalls[i] = server.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		}
	}
	return toolCalls
}

// convertLLMToolCalls 转换 LLM tool_calls 到模型（保留兼容性）
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
