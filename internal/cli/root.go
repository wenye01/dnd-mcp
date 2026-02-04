// Package cli 提供命令行工具
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dnd-client",
	Short: "DND MCP Client - D&D 游戏会话管理工具",
	Long: `DND MCP Client 是一个轻量级的有状态协调层,
用于管理 D&D 游戏会话和消息。`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(messageCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(serverCmd)
}
