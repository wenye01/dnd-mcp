package api

import (
	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/gin-gonic/gin"
)

// Router 设置路由
func Router(
	sessionHandler *handler.SessionHandler,
	chatHandler *handler.ChatHandler,
	wsHandler *handler.WebSocketHandler,
) *gin.Engine {
	r := gin.Default()

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 会话管理
		sessions := api.Group("/sessions")
		{
			sessions.POST("", sessionHandler.CreateSession)
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSession)
			sessions.POST("/:id/resume", sessionHandler.ResumeSession)
			sessions.DELETE("/:id", sessionHandler.DeleteSession)

			// 聊天
			sessions.POST("/:id/chat", chatHandler.ChatMessage)

			// WebSocket
			sessions.GET("/:id/ws", wsHandler.HandleWebSocket)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
