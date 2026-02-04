// Package cli 提供命令行工具
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/dnd-mcp/client/internal/api"
	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/repository"
	"github.com/dnd-mcp/client/internal/service"
	redisstore "github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/internal/ws"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/dnd-mcp/client/pkg/logger"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动 HTTP API 服务器",
	Long:  `启动 HTTP API 服务器，提供 RESTful API 接口`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. 加载配置
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 2. 初始化日志
		logger.Info("初始化日志", "level", cfg.Log.Level)

		// 3. 初始化 Redis Store Client
		redisStoreClient, err := redisstore.NewClient(&cfg.Redis)
		if err != nil {
			return fmt.Errorf("初始化 Redis 失败: %w", err)
		}
		defer redisStoreClient.Client().Close()

		// 测试 Redis 连接
		ctx := context.Background()
		if err := redisStoreClient.Client().Ping(ctx).Err(); err != nil {
			return fmt.Errorf("Redis 连接失败: %w", err)
		}
		logger.Info("✓ Redis 连接成功")

		// 4. 初始化 Store 层（简化架构：Handler 直接调用 Store）
		sessionStore := redisstore.NewSessionStore(redisStoreClient)
		messageStore := redisstore.NewMessageStore(redisStoreClient)

		// 5. 初始化 Mock LLM Client
		llmClient := llm.NewMockLLMClient()

		// 6. 初始化 Service 层（仅用于会话管理）
		var sessionRepo repository.SessionRepository = sessionStore
		sessionService := service.NewSessionService(sessionRepo)

		// 7. 初始化 WebSocket Hub（任务五）
		hub := ws.NewHub()
		logger.Info("✓ WebSocket Hub 初始化成功")

		// 8. 启动 Mock 事件生成器（可选，仅用于测试）
		// mockGenerator := ws.NewMockEventGenerator(hub)
		// go mockGenerator.Start()
		// logger.Info("✓ Mock 事件生成器已启动（每 10 秒生成事件）")

		// 9. 创建 HTTP 服务器
		server := api.NewServer(cfg, sessionService, sessionStore, messageStore, llmClient, hub)

		// 10. 启动服务器（后台）
		go func() {
			if err := server.Start(); err != nil {
				logger.Fatal("HTTP 服务器启动失败", "error", err)
			}
		}()

		logger.Info("HTTP 服务器已启动",
			"addr", fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		)

		// 11. 等待退出信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("收到关闭信号，正在优雅关闭...")

		// 12. 关闭 WebSocket Hub
		hub.Shutdown()
		logger.Info("✓ WebSocket Hub 已关闭")

		// 13. 优雅关闭 HTTP 服务器
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("HTTP 服务器关闭失败", "error", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
