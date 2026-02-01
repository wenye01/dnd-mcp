package handler

import (
	"net/http"
	"strconv"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	store store.Store
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(store store.Store) *SessionHandler {
	return &SessionHandler{
		store: store,
	}
}

// CreateSession 创建新会话
// POST /api/sessions
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req struct {
		CampaignName string   `json:"campaign_name" binding:"required"`
		Players      []string `json:"players"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建新会话
	session := models.NewSession(req.CampaignName)

	// 保存到数据库
	if err := h.store.CreateSession(c.Request.Context(), session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            session.ID,
		"campaign_name": session.CampaignName,
		"game_time":     session.GameTime,
		"location":      session.Location,
		"created_at":    session.CreatedAt,
	})
}

// GetSession 获取会话信息
// GET /api/sessions/:id
func (h *SessionHandler) GetSession(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// 从数据库获取会话
	session, err := h.store.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ListSessions 列出所有会话
// GET /api/sessions
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// 解析分页参数
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// 从数据库获取会话列表
	sessions, err := h.store.ListSessions(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
		"total":    len(sessions),
	})
}

// ResumeSession 恢复会话
// POST /api/sessions/:id/resume
func (h *SessionHandler) ResumeSession(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// 获取会话
	session, err := h.store.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 获取消息历史
	messages, err := h.store.GetRecentMessages(c.Request.Context(), sessionID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
		"messages": messages,
	})
}

// DeleteSession 删除会话
// DELETE /api/sessions/:id
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	id := c.Param("id")
	sessionID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// 从数据库删除会话
	if err := h.store.DeleteSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}
