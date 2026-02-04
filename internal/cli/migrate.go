// Package cli 提供数据库迁移命令
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dnd-mcp/client/internal/store/postgres"
	"github.com/dnd-mcp/client/pkg/config"
)

// migrateCmd 迁移命令
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
	Long:  `执行数据库迁移操作，包括执行迁移、回滚迁移和查看迁移状态`,
}

// migrateUpCmd 执行迁移命令
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行待执行的迁移",
	Long:  `执行所有待执行的数据库迁移，将数据库更新到最新版本`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 创建 PostgreSQL 客户端
		client, err := postgres.NewClient(cfg.Postgres)
		if err != nil {
			return fmt.Errorf("创建 PostgreSQL 客户端失败: %w", err)
		}
		defer client.Close()

		// 创建迁移器
		migrator := postgres.NewMigrator(client)

		// 执行迁移
		fmt.Println("正在执行数据库迁移...")
		if err := migrator.Up(cmd.Context()); err != nil {
			return fmt.Errorf("执行迁移失败: %w", err)
		}

		return nil
	},
}

// migrateDownCmd 回滚迁移命令
var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "回滚最后一次迁移",
	Long:  `回滚最后一次应用的数据库迁移`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 创建 PostgreSQL 客户端
		client, err := postgres.NewClient(cfg.Postgres)
		if err != nil {
			return fmt.Errorf("创建 PostgreSQL 客户端失败: %w", err)
		}
		defer client.Close()

		// 创建迁移器
		migrator := postgres.NewMigrator(client)

		// 回滚迁移
		fmt.Println("正在回滚最后一次迁移...")
		if err := migrator.Down(cmd.Context()); err != nil {
			return fmt.Errorf("回滚迁移失败: %w", err)
		}

		return nil
	},
}

// migrateStatusCmd 查看迁移状态命令
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Long:  `查看数据库迁移的当前状态和待执行的迁移`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 创建 PostgreSQL 客户端
		client, err := postgres.NewClient(cfg.Postgres)
		if err != nil {
			return fmt.Errorf("创建 PostgreSQL 客户端失败: %w", err)
		}
		defer client.Close()

		// 创建迁移器
		migrator := postgres.NewMigrator(client)

		// 获取迁移状态
		migrations, err := migrator.Status(cmd.Context())
		if err != nil {
			return fmt.Errorf("获取迁移状态失败: %w", err)
		}

		// 显示迁移状态
		fmt.Println("\n数据库迁移状态:")
		fmt.Println("================")
		for _, migration := range migrations {
			status := "  [待执行]"
			if migration.Applied {
				status = "  [已应用]"
			}
			fmt.Printf("%s %06d - %s\n", status, migration.Version, migration.Name)
			if migration.Applied {
				fmt.Printf("       应用时间: %s\n", migration.AppliedAt.Format("2006-01-02 15:04:05"))
			}
		}

		// 显示当前版本
		current, _ := migrator.GetCurrentVersion(context.Background())
		latest, _ := migrator.GetLatestVersion(context.Background())
		fmt.Printf("\n当前版本: %d\n", current)
		fmt.Printf("最新版本: %d\n", latest)

		// 检查是否最新
		if current == latest {
			fmt.Println("\n✓ 数据库已是最新版本")
		} else {
			fmt.Printf("\n✗ 数据库不是最新版本，有 %d 个待执行的迁移\n", latest-current)
		}

		return nil
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}
