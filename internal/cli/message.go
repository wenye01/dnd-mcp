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

var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "消息管理",
	Long:  `管理会话消息,包括保存、查询等操作。`,
}

var (
	messageContent string
	messageRole    string
	messageLimit   int
	playerID       string
)

// messageSaveCmd 保存消息命令
var messageSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "保存消息",
	Long:  `保存一条新的消息到会话。`,
	Example: `
  dnd-client message save --session <session-id> --content "你好" --player-id "player-123"
  dnd-client message save --session <session-id> --content "系统通知" --role system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 参数验证
		sessionID, _ := cmd.Flags().GetString("session")
		if sessionID == "" {
			return fmt.Errorf("--session 参数必填")
		}
		if messageContent == "" {
			return fmt.Errorf("--content 参数必填")
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

		// 创建消息存储
		messageStore := redis.NewMessageStore(redisClient)

		// 确定消息角色
		role := messageRole
		if role == "" {
			role = "user" // 默认为用户消息
		}

		// 创建消息对象
		message := models.NewMessage(sessionID, role, messageContent)
		message.ID = uuid.New().String()
		if role == "user" && playerID != "" {
			message.PlayerID = playerID
		}

		// 保存消息
		ctx := context.Background()
		if err := messageStore.Create(ctx, message); err != nil {
			return fmt.Errorf("保存消息失败: %w", err)
		}

		// 输出结果
		printMessage(message, outputFormat)
		return nil
	},
}

// messageListCmd 列出消息命令
var messageListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出消息",
	Long:  `列出会话的消息列表。`,
	Example: `
  dnd-client message list --session <session-id>
  dnd-client message list --session <session-id> --limit 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 参数验证
		sessionID, _ := cmd.Flags().GetString("session")
		if sessionID == "" {
			return fmt.Errorf("--session 参数必填")
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

		// 创建消息存储
		messageStore := redis.NewMessageStore(redisClient)

		// 获取消息列表
		ctx := context.Background()
		limit := messageLimit
		if limit <= 0 {
			limit = 50 // 默认50条
		}

		messages, err := messageStore.List(ctx, sessionID, limit)
		if err != nil {
			return fmt.Errorf("获取消息列表失败: %w", err)
		}

		// 输出结果
		printMessageList(messages, outputFormat)
		return nil
	},
}

// messageGetCmd 获取消息命令
var messageGetCmd = &cobra.Command{
	Use:   "get <message-id>",
	Short: "获取消息详情",
	Long:  `获取指定消息的详细信息。`,
	Args:  cobra.ExactArgs(1),
	Example: `
  dnd-client message get msg-123 --session <session-id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		messageID := args[0]

		// 参数验证
		sessionID, _ := cmd.Flags().GetString("session")
		if sessionID == "" {
			return fmt.Errorf("--session 参数必填")
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

		// 创建消息存储
		messageStore := redis.NewMessageStore(redisClient)

		// 获取消息
		ctx := context.Background()
		message, err := messageStore.Get(ctx, sessionID, messageID)
		if err != nil {
			return fmt.Errorf("获取消息失败: %w", err)
		}

		// 输出结果
		printMessage(message, outputFormat)
		return nil
	},
}

func init() {
	// 添加子命令
	messageCmd.AddCommand(messageSaveCmd)
	messageCmd.AddCommand(messageListCmd)
	messageCmd.AddCommand(messageGetCmd)

	// 保存命令参数
	messageSaveCmd.Flags().String("session", "", "会话ID(必填)")
	messageSaveCmd.Flags().StringVar(&messageContent, "content", "", "消息内容(必填)")
	messageSaveCmd.Flags().StringVar(&messageRole, "role", "", "消息角色 (user|assistant|system|tool),默认为user")
	messageSaveCmd.Flags().StringVar(&playerID, "player-id", "", "玩家ID(用户消息可选)")
	messageSaveCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")

	// 列表命令参数
	messageListCmd.Flags().String("session", "", "会话ID(必填)")
	messageListCmd.Flags().IntVar(&messageLimit, "limit", 50, "返回消息数量,默认50")
	messageListCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")

	// 获取命令参数
	messageGetCmd.Flags().String("session", "", "会话ID(必填)")
	messageGetCmd.Flags().StringVar(&outputFormat, "output", "table", "输出格式 (table|json)")
}

// printMessage 打印消息信息
func printMessage(message *models.Message, format string) {
	if format == "json" {
		data, _ := json.MarshalIndent(message, "", "  ")
		fmt.Println(string(data))
		return
	}

	// 表格格式输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID:\t", message.ID)
	fmt.Fprintln(w, "会话ID:\t", message.SessionID)
	fmt.Fprintln(w, "角色:\t", message.Role)
	fmt.Fprintln(w, "内容:\t", message.Content)
	if message.PlayerID != "" {
		fmt.Fprintln(w, "玩家ID:\t", message.PlayerID)
	}
	fmt.Fprintln(w, "创建时间:\t", message.CreatedAt.Format("2006-01-02 15:04:05"))
	w.Flush()
	fmt.Println()
}

// printMessageList 打印消息列表
func printMessageList(messages []*models.Message, format string) {
	if format == "json" {
		data, _ := json.MarshalIndent(messages, "", "  ")
		fmt.Println(string(data))
		return
	}

	if len(messages) == 0 {
		fmt.Println("没有找到任何消息")
		return
	}

	// 表格格式输出
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\t角色\t内容\t创建时间")
	fmt.Fprintln(w, "----\t----\t----\t----")
	for _, message := range messages {
		// 截断过长的内容
		content := message.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			message.ID,
			message.Role,
			content,
			message.CreatedAt.Format("15:04:05"),
		)
	}
	w.Flush()
	fmt.Printf("共 %d 条消息\n", len(messages))
}
