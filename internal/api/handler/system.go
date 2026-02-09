// Package handler 提供 HTTP 请求处理器
package handler

import (
	"context"
	"net/http"

	"github.com/dnd-mcp/client/internal/monitor"
	"github.com/gin-gonic/gin"
)

// PersistenceTriggerer 持久化触发器接口
type PersistenceTriggerer interface {
	Trigger(ctx context.Context) error
}

// SystemHandler 系统处理器
type SystemHandler struct {
	persistenceTriggerer PersistenceTriggerer
	healthMonitor        *monitor.HealthMonitor
	statsMonitor         *monitor.StatsMonitor
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(persistenceTriggerer PersistenceTriggerer, healthMonitor *monitor.HealthMonitor, statsMonitor *monitor.StatsMonitor) *SystemHandler {
	return &SystemHandler{
		persistenceTriggerer: persistenceTriggerer,
		healthMonitor:        healthMonitor,
		statsMonitor:         statsMonitor,
	}
}

// Health 健康检查
func (h *SystemHandler) Health(c *gin.Context) {
	if h.healthMonitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "health monitor not configured",
		})
		return
	}

	health := h.healthMonitor.Check(c.Request.Context())

	// 根据健康状态返回适当的HTTP状态码
	statusCode := http.StatusOK
	if health.Status == monitor.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == monitor.HealthStatusDegraded {
		statusCode = 207 // Multi-status
	}

	c.JSON(statusCode, health)
}

// Stats 系统统计
func (h *SystemHandler) Stats(c *gin.Context) {
	if h.statsMonitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "stats monitor not configured",
		})
		return
	}

	stats := h.statsMonitor.Collect(c.Request.Context())
	c.JSON(http.StatusOK, stats)
}

// TriggerPersistence 手动触发持久化
func (h *SystemHandler) TriggerPersistence(c *gin.Context) {
	if h.persistenceTriggerer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": gin.H{
				"code":    "PERSISTENCE_NOT_AVAILABLE",
				"message": "持久化管理器未配置",
			},
		})
		return
	}

	err := h.persistenceTriggerer.Trigger(c.Request.Context())
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
