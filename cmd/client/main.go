package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnd-mcp/client/internal/api"
	"github.com/dnd-mcp/client/internal/api/handler"
	"github.com/dnd-mcp/client/internal/client/websocket"
	"github.com/dnd-mcp/client/internal/config"
	"github.com/dnd-mcp/client/internal/store"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建存储层
	dataStore, err := store.NewPostgresStore(cfg.PostgreSQLURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dataStore.Close()

	log.Println("Database connected successfully")

	// 创建WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 创建处理器
	sessionHandler := handler.NewSessionHandler(dataStore)
	chatHandler := handler.NewChatHandler()
	wsHandler := handler.NewWebSocketHandler(hub)

	// 设置路由
	router := api.Router(sessionHandler, chatHandler, wsHandler)

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 等待中断信号优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭,等待5秒
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
