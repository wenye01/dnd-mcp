// Package middleware 提供 HTTP 中间件
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/dnd-mcp/client/pkg/logger"
)

// Recovery 恢复中间件，捕获 panic
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				log.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
				)

				// 返回 500 错误
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "服务器内部错误",
						"details": gin.H{},
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
