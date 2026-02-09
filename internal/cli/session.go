// Package cli 提供命令行工具
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store/redis"
	"github.com/dnd-mcp/client/pkg/config"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "会话管理",
	Long:  `管理 D&D 游戏会话,包括创建、查询、更新和删除操作。`,
}

var (
	sessionName    string
	sessionCreator string
	mcpURL         string
	maxPlayers     int
	outputFormat   string
)

// sessionCreateCmd 创建会话命令
var sessionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新会话",
	Long:  `创建一个新的 D&D 游戏会话。`,
	Example: `
  dnd-client session create --name "我的战役" --creator "user-123" --mcp-url "http://localhost:9000"
  dnd-client session create --name "我的战役" --creator "user-123" --mcp-url "http://localhost:9000" --max-players 6`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 参数验证
		if sessionName == "" {
			return fmt.Errorf("--name 参数必填")
		}
		if sessionCreator == "" {
			return fmt.Errorf("--creator 参数必填")
		}
		if mcpURL == "" {
			return fmt.Errorf("--mcp-url 参数必填")
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

		// 创建会话存储
		sessionStore := redis.NewSessionStore(redisClient)

		// 创建会话对象
		session := models.NewSession(sessionName, sessionCreator, mcpURL)
		session.ID = uuid.New().String()
		session.WebSocketKey = "ws-" + uuid.New().String()
		if maxPlayers > 0 {
			session.MaxPlayers = maxPlayers
		}

		// 保存会话
		ctx := context.Background()
		if err := sessionStore.Create(ctx, session); err != nil {
			return fmt.Errorf("创建会话失败: %w", err)
		}

		// 输出结果
		printSession(session, outputFormat)
		return nil
	},
}

// sessionGetCmd 获取会话命令
var sessionGetCmd = &cobra.Command{
	Use:   "get <session-id>",
	Short: "获取会话详情",
	Long:  `获取指定会话的详细信息。`,
	Args:  cobra.ExactArgs(1),
	Example: `
  dnd-client session get 123e4567-e89b-12d3-a456-426614174000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]

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

		// 创建会话存储
		sessionStore := redis.NewSessionStore(redisClient)

		// 获取会话
		ctx := context.Background()
		session, err := sessionStore.Get(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("获取会话失败: %w", err)
		}

		// 输出结果
		printSession(session, outputFormat)
		return nil
	},
}

// sessionListCmd 列出会话命令
var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有会话",
	Long:  `列出所有活跃的 D&D 游戏会话。`,
	Example: `
  dnd-client session list`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// 创建会话存储
		sessionStore := redis.NewSessionStore(redisClient)

		// 获取会话列表
		ctx := context.Background()
		sessions, err := sessionStore.List(ctx)
		if err != nil {
			return fmt.Errorf("获取会话列表失败: %w", err)
		}

		// 输出结果
		printSessionList(sessions, outputFormat)
		return nil
	},
}

// sessionDeleteCmd 删除会话命令
var sessionDeleteCmd = &cobra.Command{
	Use:   "delete <session-id>",
	Short: "删除会话",
	Long:  `删除指定的会话(软删除)。`,
	Args:  cobra.ExactArgs(1),
	Example: `
  dnd-client session delete 123e4567-e89b-12d3-a456-426614174000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]

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

		// 创建会话存储
		sessionStore := redis.NewSessionStore(redisClient)

		// 删除会话
		ctx := context.Background()
		if err := sessionStore.Delete(ctx, sessionID); err != nil {
			return fmt.Errorf("删除会话失败: %w", err)
		}

		fmt.Printf("会话 %s 已删除\n", sessionID)
		return nil
	},
}

func init() {
	// 添加子命令
	sessionCmd.AddCommand(sessionCreateCmd)
	sessionCmd.AddCommand(sessionGetCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionDeleteCmd)

	// 创建命令参数
	sessionCreateCmd.Flags().StringVar(&sessionName, "name", "", "会话名称(必填)")
	sessionCreateCmd.Flags().StringVar(&sessionCreator, "creator", "", "创建者ID(必填)")
	sessionCreateCmd.Flags().StringVar(&mcpURL, "mcp-url", "", "MCP Server URL(必填)")
	sessionCreateCmd.Flags().IntVar(&maxPlayers, "max-players", 0, "最大玩家数(可选)")
	sessionCreateCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")

	// 获取和列表命令参数
	sessionGetCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")
	sessionListCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")
}

// printSession 打印会话信息
func printSession(session *models.Session, format string) {
	if format == "json" {
		data, _ := json.MarshalIndent(session, "", "  ")
		fmt.Println(string(data))
		return
	}

	// 表格格式输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID:\t", session.ID)
	fmt.Fprintln(w, "名称:\t", session.Name)
	fmt.Fprintln(w, "创建者:\t", session.CreatorID)
	fmt.Fprintln(w, "MCP Server:\t", session.MCPServerURL)
	fmt.Fprintln(w, "WebSocket Key:\t", session.WebSocketKey)
	fmt.Fprintln(w, "最大玩家数:\t", session.MaxPlayers)
	fmt.Fprintln(w, "状态:\t", session.Status)
	fmt.Fprintln(w, "创建时间:\t", session.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w, "更新时间:\t", session.UpdatedAt.Format("2006-01-02 15:04:05"))
	w.Flush()
	fmt.Println()
}

// printSessionList 打印会话列表
func printSessionList(sessions []*models.Session, format string) {
	if format == "json" {
		data, _ := json.MarshalIndent(sessions, "", "  ")
		fmt.Println(string(data))
		return
	}

	if len(sessions) == 0 {
		fmt.Println("没有找到任何会话")
		return
	}

	// 表格格式输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\t名称\t创建者\t状态\t创建时间")
	fmt.Fprintln(w, "----\t----\t----\t----\t----")
	for _, session := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			session.ID,
			session.Name,
			session.CreatorID,
			session.Status,
			session.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	w.Flush()
	fmt.Printf("共 %d 个会话\n", len(sessions))
}
