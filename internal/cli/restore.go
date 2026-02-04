// Package cli 提供恢复命令
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/dnd-mcp/client/internal/store/postgres"
	"github.com/dnd-mcp/client/pkg/config"
)

// restoreCmd 恢复命令
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "从 PostgreSQL 恢复数据",
	Long:  `从 PostgreSQL 恢复会话和消息到 Redis 数据库`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 参数解析
		all, _ := cmd.Flags().GetBool("all")
		sessionID, _ := cmd.Flags().GetString("session")
		force, _ := cmd.Flags().GetBool("force")

		if all {
			return restoreAll(cmd, force)
		}

		if sessionID != "" {
			return restoreSession(cmd, sessionID, force)
		}

		return fmt.Errorf("请指定 --all 或 --session 参数")
	},
}

// restoreAll 恢复所有数据
func restoreAll(cmd *cobra.Command, force bool) error {
	fmt.Println("开始恢复所有数据...")

	if force {
		fmt.Println("⚠️  警告: 使用 --force 参数，将覆盖 Redis 中已存在的数据")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		return fmt.Errorf("创建 Redis 客户端失败: %w", err)
	}
	defer redisClient.Close()

	// 创建 PostgreSQL 客户端
	postgresClient, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		return fmt.Errorf("创建 PostgreSQL 客户端失败: %w", err)
	}
	defer postgresClient.Close()

	// 测试连接
	fmt.Println("测试数据库连接...")
	ctx := cmd.Context()
	if err := postgresClient.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL 连接失败: %w", err)
	}
	fmt.Println("✓ PostgreSQL 连接正常")

	if err := redisClient.Ping(ctx); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}
	fmt.Println("✓ Redis 连接正常")

	// 创建存储实例（Redis 存储同时实现了旧接口和 persistence 接口）
	postgresSessionStore := postgres.NewPostgresSessionStore(postgresClient)
	postgresMessageStore := postgres.NewPostgresMessageStore(postgresClient)
	redisSessionStore := redis.NewSessionStore(redisClient)
	redisMessageStore := redis.NewMessageStore(redisClient)

	// 创建恢复服务
	restoreSvc := persistence.NewRestoreService(
		postgresSessionStore,
		postgresMessageStore,
		redisSessionStore.(persistence.SessionReader),
		redisSessionStore.(persistence.SessionWriter),
		redisMessageStore.(persistence.MessageWriter),
	)

	// 执行恢复
	result, err := restoreSvc.RestoreAll(ctx, force)
	if err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	// 显示结果
	fmt.Printf("\n恢复摘要:\n")
	fmt.Printf("  会话数量: %d\n", result.SessionCount)
	if result.SkippedCount > 0 {
		fmt.Printf("  跳过数量: %d (已存在)\n", result.SkippedCount)
	}
	fmt.Printf("  消息数量: %d\n", result.MessageCount)
	fmt.Printf("  耗时: %v\n", result.Duration)

	return nil
}

// restoreSession 恢复指定会话
func restoreSession(cmd *cobra.Command, sessionID string, force bool) error {
	fmt.Printf("开始恢复会话: %s\n", sessionID)

	if force {
		fmt.Println("⚠️  警告: 使用 --force 参数，将覆盖 Redis 中已存在的会话")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 创建 Redis 客户端
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		return fmt.Errorf("创建 Redis 客户端失败: %w", err)
	}
	defer redisClient.Close()

	// 创建 PostgreSQL 客户端
	postgresClient, err := postgres.NewClient(cfg.Postgres)
	if err != nil {
		return fmt.Errorf("创建 PostgreSQL 客户端失败: %w", err)
	}
	defer postgresClient.Close()

	// 测试连接
	fmt.Println("测试数据库连接...")
	ctx := cmd.Context()
	if err := postgresClient.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL 连接失败: %w", err)
	}
	fmt.Println("✓ PostgreSQL 连接正常")

	if err := redisClient.Ping(ctx); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}
	fmt.Println("✓ Redis 连接正常")

	// 创建存储实例（Redis 存储同时实现了旧接口和 persistence 接口）
	postgresSessionStore := postgres.NewPostgresSessionStore(postgresClient)
	postgresMessageStore := postgres.NewPostgresMessageStore(postgresClient)
	redisSessionStore := redis.NewSessionStore(redisClient)
	redisMessageStore := redis.NewMessageStore(redisClient)

	// 创建恢复服务
	restoreSvc := persistence.NewRestoreService(
		postgresSessionStore,
		postgresMessageStore,
		redisSessionStore.(persistence.SessionReader),
		redisSessionStore.(persistence.SessionWriter),
		redisMessageStore.(persistence.MessageWriter),
	)

	// 执行恢复
	result, err := restoreSvc.RestoreSession(ctx, sessionID, force)
	if err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	// 显示结果
	fmt.Printf("\n恢复摘要:\n")
	fmt.Printf("  会话: %s\n", sessionID)
	if result.SkippedCount > 0 {
		fmt.Printf("  状态: 已跳过（已存在）\n")
	} else {
		fmt.Printf("  消息数量: %d\n", result.MessageCount)
	}
	fmt.Printf("  耗时: %v\n", result.Duration)

	return nil
}

func init() {
	restoreCmd.Flags().BoolP("all", "a", false, "恢复所有数据")
	restoreCmd.Flags().StringP("session", "s", "", "恢复指定会话")
	restoreCmd.Flags().Bool("force", false, "强制覆盖已存在的数据")
}
