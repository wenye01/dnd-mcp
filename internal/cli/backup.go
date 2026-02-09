// Package cli 提供备份命令
package cli

import (
	"fmt"

	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/dnd-mcp/client/internal/store/postgres"
	"github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/spf13/cobra"
)

// backupCmd 备份命令
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "备份数据到 PostgreSQL",
	Long:  `从 Redis 备份会话和消息到 PostgreSQL 数据库`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 参数解析
		all, _ := cmd.Flags().GetBool("all")
		sessionID, _ := cmd.Flags().GetString("session")
		list, _ := cmd.Flags().GetBool("list")

		if list {
			return listBackups(cmd)
		}

		if all {
			return backupAll(cmd)
		}

		if sessionID != "" {
			return backupSession(cmd, sessionID)
		}

		return fmt.Errorf("请指定 --all 或 --session 参数，或使用 --list 查看备份记录")
	},
}

// backupAll 备份所有数据
func backupAll(cmd *cobra.Command) error {
	fmt.Println("开始备份所有数据...")

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
	if err := redisClient.Ping(ctx); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}
	fmt.Println("✓ Redis 连接正常")

	if err := postgresClient.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL 连接失败: %w", err)
	}
	fmt.Println("✓ PostgreSQL 连接正常")

	// 创建存储实例（Redis 存储同时实现了旧接口和 persistence 接口）
	redisSessionStore := redis.NewSessionStore(redisClient)
	redisMessageStore := redis.NewMessageStore(redisClient)
	postgresSessionStore := postgres.NewPostgresSessionStore(postgresClient)
	postgresMessageStore := postgres.NewPostgresMessageStore(postgresClient)

	// 创建备份服务
	backupSvc := persistence.NewBackupService(
		redisSessionStore.(persistence.SessionReader),
		redisMessageStore.(persistence.MessageReader),
		postgresSessionStore,
		postgresMessageStore,
	)

	// 执行备份
	result, err := backupSvc.BackupAll(ctx)
	if err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 显示结果
	fmt.Printf("\n备份摘要:\n")
	fmt.Printf("  会话数量: %d\n", result.SessionCount)
	fmt.Printf("  消息数量: %d\n", result.MessageCount)
	fmt.Printf("  耗时: %v\n", result.Duration)

	return nil
}

// backupSession 备份指定会话
func backupSession(cmd *cobra.Command, sessionID string) error {
	fmt.Printf("开始备份会话: %s\n", sessionID)

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
	if err := redisClient.Ping(ctx); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}
	fmt.Println("✓ Redis 连接正常")

	if err := postgresClient.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL 连接失败: %w", err)
	}
	fmt.Println("✓ PostgreSQL 连接正常")

	// 创建存储实例（Redis 存储同时实现了旧接口和 persistence 接口）
	redisSessionStore := redis.NewSessionStore(redisClient)
	redisMessageStore := redis.NewMessageStore(redisClient)
	postgresSessionStore := postgres.NewPostgresSessionStore(postgresClient)
	postgresMessageStore := postgres.NewPostgresMessageStore(postgresClient)

	// 创建备份服务
	backupSvc := persistence.NewBackupService(
		redisSessionStore.(persistence.SessionReader),
		redisMessageStore.(persistence.MessageReader),
		postgresSessionStore,
		postgresMessageStore,
	)

	// 执行备份
	result, err := backupSvc.BackupSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	// 显示结果
	fmt.Printf("\n备份摘要:\n")
	fmt.Printf("  会话: %s\n", sessionID)
	fmt.Printf("  消息数量: %d\n", result.MessageCount)
	fmt.Printf("  耗时: %v\n", result.Duration)

	return nil
}

// listBackups 查看备份记录
func listBackups(cmd *cobra.Command) error {
	fmt.Println("查看备份记录...")
	fmt.Println("\n功能开发中...")
	fmt.Println("提示: 可以查询 persistence_snapshots 表查看备份历史")
	return nil
}

func init() {
	backupCmd.Flags().BoolP("all", "a", false, "备份所有数据")
	backupCmd.Flags().StringP("session", "s", "", "备份指定会话")
	backupCmd.Flags().Bool("list", false, "查看备份记录")
}
