// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/dnd-mcp/client/internal/persistence"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	persistenceManager *persistence.Manager
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(persistenceManager *persistence.Manager) *SystemHandler {
	return &SystemHandler{
		persistenceManager: persistenceManager,
	}
}

// TriggerPersistence 手动触发持久化
func (h *SystemHandler) TriggerPersistence(c *gin.Context) {
	err := h.persistenceManager.Trigger(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "PERSISTENCE_ERROR",
				"message": "持久化失败: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "持久化已触发",
	})
}
