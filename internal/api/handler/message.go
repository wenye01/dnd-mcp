// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	messageStore store.MessageStore
	sessionStore store.SessionStore
	llmClient    llm.LLMClient
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(messageStore store.MessageStore, sessionStore store.SessionStore, llmClient llm.LLMClient) *MessageHandler {
	return &MessageHandler{
		messageStore: messageStore,
		sessionStore: sessionStore,
		llmClient:    llmClient,
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

	// 保存用户消息
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

	// 调用 LLM
	llmResp, err := h.llmClient.Chat(c.Request.Context(), &llm.ChatRequest{
		Messages: []llm.Message{{Role: "user", Content: req.Content}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "LLM_ERROR", "message": "LLM 调用失败: " + err.Error()}})
		return
	}

	// 保存助手消息
	assistantMessage := &models.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      llmResp.Message.Role,
		Content:   llmResp.Message.Content,
		CreatedAt: time.Now(),
	}
	if err := h.messageStore.Create(c.Request.Context(), assistantMessage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "保存助手消息失败"}})
		return
	}

	c.JSON(http.StatusOK, assistantMessage)
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
