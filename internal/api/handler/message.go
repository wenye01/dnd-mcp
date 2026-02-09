// Package handler 提供 HTTP 请求处理器
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/mcp"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	messageStore   store.MessageStore
	sessionStore   store.SessionStore
	llmClient      llm.LLMClient
	mcpClient      mcp.MCPClient
	contextBuilder *service.ContextBuilder
	wsHub          *ws.Hub
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(
	messageStore store.MessageStore,
	sessionStore store.SessionStore,
	llmClient llm.LLMClient,
	mcpClient mcp.MCPClient,
	contextBuilder *service.ContextBuilder,
	wsHub *ws.Hub,
) *MessageHandler {
	return &MessageHandler{
		messageStore:   messageStore,
		sessionStore:   sessionStore,
		llmClient:      llmClient,
		mcpClient:      mcpClient,
		contextBuilder: contextBuilder,
		wsHub:          wsHub,
	}
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Content  string `json:"content" binding:"required"`
	PlayerID string `json:"player_id" binding:"required"`
	Stream   bool   `json:"stream"`
}

// SendMessage 发送消息并获取 AI 响应
func (h *MessageHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("id")

	// 验证会话是否存在
	_, err := h.sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "SESSION_NOT_FOUND", "message": "会话不存在"}})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "请求参数错误: " + err.Error()}})
		return
	}

	// 1. 保存用户消息
	userMessage := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
		PlayerID:  req.PlayerID,
		CreatedAt: time.Now(),
	}
	if err := h.messageStore.Create(c.Request.Context(), userMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存用户消息失败"}})
		return
	}

	// 2. 构建对话上下文（包含历史消息）
	messages, err := h.contextBuilder.BuildContext(c.Request.Context(), sessionID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "构建对话上下文失败"}})
		return
	}

	// 3. 调用 LLM
	llmResp, err := h.llmClient.Chat(c.Request.Context(), &llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LLM_ERROR", "message": "LLM 调用失败: " + err.Error()}})
		return
	}

	// 4. 检查响应类型
	if len(llmResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LLM_ERROR", "message": "LLM 返回空响应"}})
		return
	}

	choice := llmResp.Choices[0]

	// 5. 判断是否有 tool_calls
	if choice.FinishReason == "tool_calls" && len(choice.Message.ToolCalls) > 0 {
		// 处理工具调用
		h.handleToolCalls(c, sessionID, choice.Message.ToolCalls, req.Content)
		return
	}

	// 6. 保存助手消息（纯文本响应）
	assistantMessage := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      choice.Message.Role,
		Content:   choice.Message.Content,
		CreatedAt: time.Now(),
	}
	if err := h.messageStore.Create(c.Request.Context(), assistantMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存助手消息失败"}})
		return
	}

	// 7. 返回助手响应
	c.JSON(http.StatusOK, assistantMessage)
}

// handleToolCalls 处理工具调用
func (h *MessageHandler) handleToolCalls(c *gin.Context, sessionID string, toolCalls []llm.ToolCall, originalUserMessage string) {
	ctx := c.Request.Context()

	// 1. 保存 assistant 消息（包含 tool_calls）
	assistantMsg := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      "assistant",
		Content:   "",
		ToolCalls: convertLLMToolCalls(toolCalls),
		CreatedAt: time.Now(),
	}

	if err := h.messageStore.Create(ctx, assistantMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存工具调用消息失败"}})
		return
	}

	// 2. 执行所有工具调用
	toolResults := make([]map[string]interface{}, len(toolCalls))
	for i, toolCall := range toolCalls {
		// 解析 arguments
		var args map[string]interface{}
		json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

		// 调用 MCP Server
		result, err := h.mcpClient.CallTool(ctx, sessionID, toolCall.Function.Name, args)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "TOOL_EXECUTION_ERROR",
					"message": fmt.Sprintf("工具调用失败: %s", err.Error()),
				},
			})
			return
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

		if err := h.messageStore.Create(ctx, toolMsg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存工具响应失败"}})
			return
		}
	}

	// 4. 重新构建上下文（包含工具调用和结果）
	messages, err := h.contextBuilder.BuildContext(ctx, sessionID, originalUserMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "构建上下文失败"}})
		return
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

	followupResp, err := h.llmClient.Chat(ctx, followupReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LLM_ERROR", "message": "LLM 后续调用失败"}})
		return
	}

	// 6. 保存最终助手响应
	if len(followupResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LLM_ERROR", "message": "LLM 返回空响应"}})
		return
	}

	finalMsg := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      followupResp.Choices[0].Message.Role,
		Content:   followupResp.Choices[0].Message.Content,
		CreatedAt: time.Now(),
	}

	if err := h.messageStore.Create(ctx, finalMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存最终响应失败"}})
		return
	}

	// 7. 返回最终响应
	c.JSON(http.StatusOK, finalMsg)
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

// GetMessages 获取消息历史
func (h *MessageHandler) GetMessages(c *gin.Context) {
	sessionID := c.Param("id")

	// 验证会话是否存在
	_, err := h.sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "SESSION_NOT_FOUND", "message": "会话不存在"}})
		return
	}

	// 解析 limit 参数
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				l = 100
			}
			limit = l
		}
	}

	messages, err := h.messageStore.List(c.Request.Context(), sessionID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "获取消息列表失败"}})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetMessage 获取单条消息
func (h *MessageHandler) GetMessage(c *gin.Context) {
	sessionID := c.Param("id")
	messageID := c.Param("messageId")

	// 验证会话是否存在
	_, err := h.sessionStore.Get(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "SESSION_NOT_FOUND", "message": "会话不存在"}})
		return
	}

	message, err := h.messageStore.Get(c.Request.Context(), sessionID, messageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "MESSAGE_NOT_FOUND", "message": "消息不存在"}})
		return
	}

	c.JSON(http.StatusOK, message)
}
