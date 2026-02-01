package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	// TODO: 添加orchestrator依赖
}

// NewChatHandler 创建聊天处理器
func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
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

	// TODO: 调用orchestrator处理消息
	_ = sessionID

	c.JSON(http.StatusOK, gin.H{
		"response": "AI response placeholder",
		"state_changes": gin.H{
			"example": "data",
		},
		"requires_roll": false,
		"turn_info":     nil,
	})
}
