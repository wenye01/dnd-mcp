// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"

	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/gin-gonic/gin"
)

// AdminHandler 管理员API处理器
type AdminHandler struct {
	migrator   *persistence.Migrator
	backupSvc  *persistence.BackupService
	restoreSvc *persistence.RestoreService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(
	migrator *persistence.Migrator,
	backupSvc *persistence.BackupService,
	restoreSvc *persistence.RestoreService,
) *AdminHandler {
	return &AdminHandler{
		migrator:   migrator,
		backupSvc:  backupSvc,
		restoreSvc: restoreSvc,
	}
}

// Migrate 执行数据库迁移
// POST /api/admin/migrate
func (h *AdminHandler) Migrate(c *gin.Context) {
	if err := h.migrator.Up(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "执行迁移失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "数据库迁移成功",
	})
}

// MigrateStatus 查看迁移状态
// GET /api/admin/migrate/status
func (h *AdminHandler) MigrateStatus(c *gin.Context) {
	migrations, err := h.migrator.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取迁移状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"migrations": migrations,
	})
}

// BackupAll 备份所有数据
// POST /api/admin/backup/all
func (h *AdminHandler) BackupAll(c *gin.Context) {
	result, err := h.backupSvc.BackupAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "备份失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"session_count": result.SessionCount,
		"message_count": result.MessageCount,
		"duration_ms":   result.Duration.Milliseconds(),
		"start_time":    result.StartTime,
		"end_time":      result.EndTime,
	})
}

// BackupSession 备份指定会话
// POST /api/admin/backup/session/:id
func (h *AdminHandler) BackupSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id 参数必填",
		})
		return
	}

	result, err := h.backupSvc.BackupSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "备份失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"session_count": result.SessionCount,
		"message_count": result.MessageCount,
		"duration_ms":   result.Duration.Milliseconds(),
		"start_time":    result.StartTime,
		"end_time":      result.EndTime,
	})
}

// RestoreAll 恢复所有数据
// POST /api/admin/restore/all
func (h *AdminHandler) RestoreAll(c *gin.Context) {
	// 获取 force 参数
	forceStr := c.Query("force")
	force := forceStr == "true" || forceStr == "1"

	result, err := h.restoreSvc.RestoreAll(c.Request.Context(), force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "恢复失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"session_count": result.SessionCount,
		"message_count": result.MessageCount,
		"skipped_count": result.SkippedCount,
		"duration_ms":   result.Duration.Milliseconds(),
		"start_time":    result.StartTime,
		"end_time":      result.EndTime,
	})
}

// RestoreSession 恢复指定会话
// POST /api/admin/restore/session/:id
func (h *AdminHandler) RestoreSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id 参数必填",
		})
		return
	}

	// 获取 force 参数
	forceStr := c.Query("force")
	force := forceStr == "true" || forceStr == "1"

	result, err := h.restoreSvc.RestoreSession(c.Request.Context(), sessionID, force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "恢复失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"session_count": result.SessionCount,
		"message_count": result.MessageCount,
		"skipped_count": result.SkippedCount,
		"duration_ms":   result.Duration.Milliseconds(),
		"start_time":    result.StartTime,
		"end_time":      result.EndTime,
	})
}

// GetBackupHistory 查看备份记录（暂时返回空，后续可实现）
// GET /api/admin/backups
func (h *AdminHandler) GetBackupHistory(c *gin.Context) {
	// TODO: 实现备份历史记录功能
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "备份历史记录功能待实现",
		"backups": []interface{}{},
	})
}
