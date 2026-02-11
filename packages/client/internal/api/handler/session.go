// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/gin-gonic/gin"
)

// SessionResponse 会话响应
type SessionResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	CreatorID    string                 `json:"creator_id"`
	MCPServerURL string                 `json:"mcp_server_url"`
	WebSocketKey string                 `json:"websocket_key"`
	MaxPlayers   int                    `json:"max_players"`
	Settings     map[string]interface{} `json:"settings"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	Status       string                 `json:"status"`
}

// CreateSession 创建会话 Handler
func CreateSession(sessionService service.SessionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.CreateSessionRequest

		// 参数绑定和验证
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_REQUEST",
					"message": "请求参数错误",
					"details": gin.H{"fields": err.Error()},
				},
			})
			return
		}

		// 调用 Service
		session, err := sessionService.CreateSession(c.Request.Context(), &req)
		if err != nil {
			handleError(c, err)
			return
		}

		// 返回 201 Created
		c.JSON(http.StatusCreated, toSessionResponse(session))
	}
}

// GetSession 获取会话详情 Handler
func GetSession(sessionService service.SessionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		session, err := sessionService.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, toSessionResponse(session))
	}
}

// ListSessions 列出所有会话 Handler
func ListSessions(sessionService service.SessionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取查询参数
		status := c.DefaultQuery("status", "active")

		sessions, err := sessionService.ListSessions(c.Request.Context(), status)
		if err != nil {
			handleError(c, err)
			return
		}

		// 转换为响应格式
		response := make([]SessionResponse, len(sessions))
		for i, session := range sessions {
			response[i] = toSessionResponse(session)
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateSession 更新会话 Handler
func UpdateSession(sessionService service.SessionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		var req service.UpdateSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_REQUEST",
					"message": "请求参数错误",
					"details": gin.H{"fields": err.Error()},
				},
			})
			return
		}

		session, err := sessionService.UpdateSession(c.Request.Context(), sessionID, &req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, toSessionResponse(session))
	}
}

// DeleteSession 删除会话 Handler
func DeleteSession(sessionService service.SessionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		if err := sessionService.DeleteSession(c.Request.Context(), sessionID); err != nil {
			handleError(c, err)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// handleError 统一错误处理
func handleError(c *gin.Context, err error) {
	// 根据错误类型返回不同的 HTTP 状态码
	switch {
	case errors.Is(err, errors.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "SESSION_NOT_FOUND",
				"message": "会话不存在",
				"details": gin.H{},
			},
		})
	case errors.Is(err, errors.ErrInvalidMaxPlayers):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_MAX_PLAYERS",
				"message": "max_players 必须在 1-10 之间",
				"details": gin.H{},
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "服务器内部错误",
				"details": gin.H{},
			},
		})
	}
}

// toSessionResponse 转换为响应格式
func toSessionResponse(session *models.Session) SessionResponse {
	return SessionResponse{
		ID:           session.ID,
		Name:         session.Name,
		CreatorID:    session.CreatorID,
		MCPServerURL: session.MCPServerURL,
		WebSocketKey: session.WebSocketKey,
		MaxPlayers:   session.MaxPlayers,
		Settings:     session.Settings,
		CreatedAt:    session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    session.UpdatedAt.Format(time.RFC3339),
		Status:       session.Status,
	}
}
