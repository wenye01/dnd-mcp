// Package service 提供业务逻辑服务
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/server"
)

// ContextBuilderConfig ContextBuilder 配置
type ContextBuilderConfig struct {
	// UseRawContext 是否使用原始上下文（完整模式）
	UseRawContext bool
	// MessageLimit 消息限制（简化模式）
	MessageLimit int
	// IncludeCombat 是否包含战斗信息
	IncludeCombat bool
}

// DefaultContextBuilderConfig 默认配置
var DefaultContextBuilderConfig = ContextBuilderConfig{
	UseRawContext: false,
	MessageLimit:  20,
	IncludeCombat: true,
}

// ContextBuilder 上下文构建器
// 负责构建 LLM 对话所需的完整上下文
// 从 Server 获取上下文，而非本地存储
type ContextBuilder struct {
	serverClient server.ServerClient
	config       ContextBuilderConfig
}

// NewContextBuilder 创建上下文构建器
func NewContextBuilder(serverClient server.ServerClient, config *ContextBuilderConfig) *ContextBuilder {
	cfg := DefaultContextBuilderConfig
	if config != nil {
		cfg = *config
	}
	return &ContextBuilder{
		serverClient: serverClient,
		config:       cfg,
	}
}

// BuildContext 构建对话上下文
// 从 Server 获取上下文，转换为 LLM 消息格式
func (b *ContextBuilder) BuildContext(ctx context.Context, campaignID, userMessage string) ([]llm.Message, error) {
	// 1. 从 Server 获取上下文
	var messages []llm.Message
	var err error

	if b.config.UseRawContext {
		// 完整模式：获取原始数据
		messages, err = b.buildFromRawContext(ctx, campaignID, userMessage)
	} else {
		// 简化模式：使用 Server 压缩后的上下文
		messages, err = b.buildFromContext(ctx, campaignID, userMessage)
	}

	if err != nil {
		return nil, err
	}

	return messages, nil
}

// buildFromContext 使用简化模式构建上下文
func (b *ContextBuilder) buildFromContext(ctx context.Context, campaignID, userMessage string) ([]llm.Message, error) {
	// 从 Server 获取压缩后的上下文
	serverCtx, err := b.serverClient.GetContext(ctx, campaignID, b.config.MessageLimit, b.config.IncludeCombat)
	if err != nil {
		return nil, fmt.Errorf("获取上下文失败: %w", err)
	}

	// 1. 构建 System 消息
	messages := []llm.Message{
		{
			Role:    "system",
			Content: b.buildSystemPrompt(serverCtx),
		},
	}

	// 2. 添加历史消息
	for _, msg := range serverCtx.Messages {
		messages = append(messages, llm.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// 3. 添加当前用户消息
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	return messages, nil
}

// buildFromRawContext 使用完整模式构建上下文
func (b *ContextBuilder) buildFromRawContext(ctx context.Context, campaignID, userMessage string) ([]llm.Message, error) {
	// 从 Server 获取原始上下文
	rawCtx, err := b.serverClient.GetRawContext(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("获取原始上下文失败: %w", err)
	}

	// 1. 构建 System 消息（包含完整的游戏状态）
	messages := []llm.Message{
		{
			Role:    "system",
			Content: b.buildSystemPromptFromRaw(rawCtx),
		},
	}

	// 2. 添加历史消息（应用滑动窗口）
	messageLimit := b.config.MessageLimit
	if messageLimit <= 0 {
		messageLimit = 50 // 默认 50 条
	}

	startIdx := 0
	if len(rawCtx.Messages) > messageLimit {
		startIdx = len(rawCtx.Messages) - messageLimit
	}

	for i := startIdx; i < len(rawCtx.Messages); i++ {
		msg := rawCtx.Messages[i]
		messages = append(messages, llm.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// 3. 添加当前用户消息
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	return messages, nil
}

// buildSystemPrompt 构建 System Prompt（简化模式）
func (b *ContextBuilder) buildSystemPrompt(ctx *server.Context) string {
	var sb strings.Builder

	// 基础角色定义
	sb.WriteString("你是 DND 游戏的 DM(地下城主)。负责描述场景、扮演 NPC、判断规则、推进剧情。\n\n")

	// 游戏状态摘要
	if ctx.GameSummary != nil {
		sb.WriteString("=== 当前游戏状态 ===\n")
		sb.WriteString(fmt.Sprintf("地点: %s\n", ctx.GameSummary.Location))
		sb.WriteString(fmt.Sprintf("时间: %s\n", ctx.GameSummary.Time))

		if ctx.GameSummary.InCombat {
			sb.WriteString("状态: 战斗中\n")
			if ctx.GameSummary.Combat != nil {
				sb.WriteString(fmt.Sprintf("回合: %d\n", ctx.GameSummary.Combat.Round))
			}
		}
		sb.WriteString("\n")
	}

	// 统计信息
	sb.WriteString(fmt.Sprintf("对话历史: %d 条消息（显示最近 %d 条）\n",
		ctx.RawMessageCount, len(ctx.Messages)))

	return sb.String()
}

// buildSystemPromptFromRaw 构建 System Prompt（完整模式）
func (b *ContextBuilder) buildSystemPromptFromRaw(rawCtx *server.RawContext) string {
	var sb strings.Builder

	// 基础角色定义
	sb.WriteString("你是 DND 游戏的 DM(地下城主)。负责描述场景、扮演 NPC、判断规则、推进剧情。\n\n")

	// 游戏状态
	if rawCtx.GameState != nil {
		sb.WriteString("=== 游戏状态 ===\n")
		sb.WriteString(fmt.Sprintf("地点: %s\n", rawCtx.GameState.Location))
		sb.WriteString(fmt.Sprintf("时间: %s\n", rawCtx.GameState.GameTime))
		sb.WriteString("\n")
	}

	// 队伍成员
	if len(rawCtx.Characters) > 0 {
		sb.WriteString("=== 队伍成员 ===\n")
		for _, char := range rawCtx.Characters {
			if char.Type == "pc" {
				hpPercent := float64(0)
				if char.MaxHP > 0 {
					hpPercent = float64(char.HP) / float64(char.MaxHP) * 100
				}
				hpStatus := "健康"
				if hpPercent == 0 {
					hpStatus = "昏迷/死亡"
				} else if hpPercent < 25 {
					hpStatus = "危急"
				} else if hpPercent < 50 {
					hpStatus = "重伤"
				} else if hpPercent < 75 {
					hpStatus = "轻伤"
				}

				// 获取职业信息
				classInfo := ""
				if len(char.Classes) > 0 {
					classInfo = fmt.Sprintf("%s Lv.%d", char.Classes[0].Name, char.Classes[0].Level)
				}

				sb.WriteString(fmt.Sprintf("- %s (%s): HP %d/%d [%s]\n",
					char.Name, classInfo, char.HP, char.MaxHP, hpStatus))
			}
		}
		sb.WriteString("\n")
	}

	// 战斗状态
	if rawCtx.Combat != nil && rawCtx.Combat.Active {
		sb.WriteString("=== 战斗状态 ===\n")
		sb.WriteString(fmt.Sprintf("回合: %d\n", rawCtx.Combat.Round))
		sb.WriteString(fmt.Sprintf("当前行动: 第 %d 位\n", rawCtx.Combat.Turn))
		sb.WriteString("\n")
	}

	// 地图信息
	if rawCtx.Map != nil {
		sb.WriteString("=== 当前地图 ===\n")
		sb.WriteString(fmt.Sprintf("名称: %s\n", rawCtx.Map.Name))
		sb.WriteString(fmt.Sprintf("大小: %dx%d 格\n", rawCtx.Map.Width, rawCtx.Map.Height))
		sb.WriteString("\n")
	}

	return sb.String()
}

// SetUseRawContext 设置是否使用原始上下文
func (b *ContextBuilder) SetUseRawContext(useRawContext bool) {
	b.config.UseRawContext = useRawContext
}

// SetMessageLimit 设置消息限制
func (b *ContextBuilder) SetMessageLimit(limit int) {
	b.config.MessageLimit = limit
}

// SetIncludeCombat 设置是否包含战斗信息
func (b *ContextBuilder) SetIncludeCombat(include bool) {
	b.config.IncludeCombat = include
}
