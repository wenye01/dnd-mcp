package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/dnd-mcp/client/internal/client/llm"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	llmClient llm.Client
	store     store.Store // 用于保存消息
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(llmClient llm.Client, store store.Store) *ChatHandler {
	return &ChatHandler{
		llmClient: llmClient,
		store:     store,
	}
}

// ChatMessage 处理聊天消息
// POST /api/sessions/:id/chat
func (h *ChatHandler) ChatMessage(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req struct {
		Message     string `json:"message" binding:"required"`
		PlayerID    string `json:"player_id,omitempty"`
		ContextOnly bool   `json:"context_only,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查会话是否存在
	_, err = h.store.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 保存用户消息到数据库
	userMsg := &models.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		CreatedAt: time.Now(),
		Role:      "user",
		Content:   req.Message,
		PlayerID:  req.PlayerID,
	}

	if err := h.store.CreateMessage(c.Request.Context(), userMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save message"})
		return
	}

	// 构建LLM请求
	llmReq := &llm.ChatCompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "你是一个DND(龙与地下城)的地下城主DM。请用友好、有趣的方式与玩家互动。暂时不要调用任何工具。",
			},
			{
				Role:    "user",
				Content: req.Message,
			},
		},
		Model:       "gpt-4",
		Temperature: 0.7,
	}

	// 调用LLM
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	llmResp, err := h.llmClient.ChatCompletion(ctx, llmReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get LLM response"})
		return
	}

	// 提取响应内容
	var responseText string
	var toolCalls []llm.ToolCall

	if len(llmResp.Choices) > 0 {
		choice := llmResp.Choices[0]
		responseText = choice.Message.Content
		toolCalls = choice.Message.ToolCalls
	}

	// 保存助手响应到数据库
	assistantMsg := &models.Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		CreatedAt: time.Now(),
		Role:      "assistant",
		Content:   responseText,
	}

	if err := h.store.CreateMessage(c.Request.Context(), assistantMsg); err != nil {
		// 日志记录错误,但不中断响应
		// log.Printf("Failed to save assistant message: %v", err)
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"response": responseText,
		"usage":    llmResp.Usage,
		"tool_calls": func() []llm.ToolCall {
			if len(toolCalls) > 0 {
				return toolCalls
			}
			return nil
		}(),
		"state_changes": nil, // 暂时不处理状态变更
		"requires_roll": false, // 暂时不处理骰子
	})
}
