// Package middleware 提供 HTTP 中间件
package middleware

import (
	"time"

	"github.com/dnd-mcp/client/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(startTime)

		// 记录日志
		log.Info("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency,
			"client_ip", c.ClientIP(),
		)
	}
}
