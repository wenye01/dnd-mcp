// Package api 提供 HTTP API 服务器
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/api/middleware"
	"github.com/dnd-mcp/client/internal/server"
	"github.com/dnd-mcp/client/internal/service"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/internal/ws"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/dnd-mcp/client/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器
type Server struct {
	config         *config.Config
	httpServer     *http.Server
	router         *gin.Engine
	sessionService *service.SessionService
	chatService    service.ChatServiceInterface
	sessionStore   store.SessionStore
	messageStore   store.MessageStore
	serverClient   server.ServerClient
	hub            *ws.Hub
	systemHandler  *handler.SystemHandler
}

// NewServer 创建 HTTP 服务器
func NewServer(
	cfg *config.Config,
	sessionService *service.SessionService,
	chatService service.ChatServiceInterface,
	sessionStore store.SessionStore,
	messageStore store.MessageStore,
	serverClient server.ServerClient,
	hub *ws.Hub,
	systemHandler *handler.SystemHandler,
) *Server {
	// 设置 Gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &Server{
		config:         cfg,
		router:         router,
		sessionService: sessionService,
		chatService:    chatService,
		sessionStore:   sessionStore,
		messageStore:   messageStore,
		serverClient:   serverClient,
		hub:            hub,
		systemHandler:  systemHandler,
	}

	// 设置中间件
	server.setupMiddleware()

	// 设置路由
	server.setupRoutes()

	return server
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger(logger.Instance))
	s.router.Use(middleware.Recovery(logger.Instance))
	if s.config.HTTP.EnableCORS {
		s.router.Use(middleware.CORS())
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 创建消息处理器
	messageHandler := handler.NewMessageHandler(s.chatService, s.sessionStore, s.messageStore, s.serverClient, s.hub)

	// 创建 WebSocket 处理器
	wsHandler := handler.NewWSHandler(s.hub, s.sessionStore)

	// API 路由组
	api := s.router.Group("/api")
	{
		// 会话路由（旧路由，保持兼容）
		sessions := api.Group("/sessions")
		{
			sessions.POST("", handler.CreateSession(s.sessionService))
			sessions.GET("", handler.ListSessions(s.sessionService))

			// 会话详情路由 (使用 :id 参数)
			sessions.GET("/:id", handler.GetSession(s.sessionService))
			sessions.PATCH("/:id", handler.UpdateSession(s.sessionService))
			sessions.DELETE("/:id", handler.DeleteSession(s.sessionService))

			// 消息路由（任务四）- 使用 :id 参数保持一致
			sessions.POST("/:id/chat", messageHandler.SendMessage)
			sessions.GET("/:id/messages", messageHandler.GetMessages)
			sessions.GET("/:id/messages/:messageId", messageHandler.GetMessage)

			// WebSocket 广播测试路由（仅用于测试）
			sessions.POST("/:id/broadcast", wsHandler.BroadcastMessage)
		}

		// Campaign 路由（新路由，推荐使用）
		// 内部映射到相同的 handler，保持兼容性
		campaigns := api.Group("/campaigns")
		{
			campaigns.POST("", handler.CreateSession(s.sessionService))
			campaigns.GET("", handler.ListSessions(s.sessionService))

			// Campaign 详情路由
			campaigns.GET("/:id", handler.GetSession(s.sessionService))
			campaigns.PATCH("/:id", handler.UpdateSession(s.sessionService))
			campaigns.DELETE("/:id", handler.DeleteSession(s.sessionService))

			// 消息路由
			campaigns.POST("/:id/chat", messageHandler.SendMessage)
			campaigns.GET("/:id/messages", messageHandler.GetMessages)
			campaigns.GET("/:id/messages/:messageId", messageHandler.GetMessage)

			// WebSocket 广播测试路由
			campaigns.POST("/:id/broadcast", wsHandler.BroadcastMessage)
		}

		// 系统路由（任务八：持久化触发器，任务九：健康检查和统计）
		system := api.Group("/system")
		{
			system.GET("/health", s.systemHandler.Health)
			system.GET("/stats", s.systemHandler.Stats)
			system.POST("/persistence/trigger", s.systemHandler.TriggerPersistence)
		}
	}

	// WebSocket 路由（任务五）- 支持 sessions 和 campaigns 两种路径
	s.router.GET("/ws/sessions/:id", wsHandler.HandleWebSocket)
	s.router.GET("/ws/campaigns/:id", wsHandler.HandleWebSocket)

	// 测试路由（仅用于开发测试）
	test := s.router.Group("/test")
	{
		test.GET("/ws/connections", wsHandler.GetConnectionsInfo)
		test.POST("/ws/broadcast", wsHandler.BroadcastTestMessage)
	}
}

// Start 启动 HTTP 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.HTTP.WriteTimeout) * time.Second,
	}

	logger.Info("HTTP 服务器启动", "addr", addr, "timeout", s.config.HTTP.ReadTimeout)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP 服务器启动失败: %w", err)
	}

	return nil
}

// Shutdown 优雅关闭 HTTP 服务器
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("HTTP 服务器正在关闭...")

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.HTTP.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP 服务器关闭失败: %w", err)
	}

	logger.Info("HTTP 服务器已关闭")
	return nil
}
