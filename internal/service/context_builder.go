// Package service 提供业务逻辑服务
package service

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/client/internal/llm"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/store"
)

// ContextBuilder 上下文构建器
// 负责构建 LLM 对话所需的完整上下文，包括系统提示词和历史消息
type ContextBuilder struct {
	messageStore store.MessageStore
	sessionStore store.SessionStore
}

// NewContextBuilder 创建上下文构建器
func NewContextBuilder(messageStore store.MessageStore, sessionStore store.SessionStore) *ContextBuilder {
	return &ContextBuilder{
		messageStore: messageStore,
		sessionStore: sessionStore,
	}
}

// BuildContext 构建对话上下文
// 返回完整的 LLM 消息列表，包括系统提示词、历史消息和当前用户消息
func (b *ContextBuilder) BuildContext(ctx context.Context, sessionID, userMessage string) ([]llm.Message, error) {
	//
	// === 配置区：快速调整上下文构建策略 ===
	//
	// 1. System Prompt - 定义 AI 的角色和行为
	systemPrompt := b.buildSystemPrompt()

	// 2. 历史消息加载策略
	maxHistory := 50           // 最多加载多少条历史消息（默认 50 条）
	includeRoles := []string{} // 包含哪些角色的消息（空数组表示全部）
	excludeRoles := []string{} // 排除哪些角色的消息（如 ["system"]）

	//
	// === 执行区 ===
	//

	// 1. 构建 System 消息
	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// 2. 加载历史消息
	history, err := b.messageStore.List(ctx, sessionID, maxHistory)
	if err != nil {
		return nil, fmt.Errorf("加载历史消息失败: %w", err)
	}

	// 3. 应用过滤规则并添加历史消息
	for _, msg := range history {
		// 角色过滤
		if len(includeRoles) > 0 && !contains(includeRoles, msg.Role) {
			continue
		}
		if len(excludeRoles) > 0 && contains(excludeRoles, msg.Role) {
			continue
		}

		// 添加到上下文
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 4. 添加当前用户消息
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	return messages, nil
}

// buildSystemPrompt 构建 System Prompt
// 定义 AI 的角色（DND 游戏的 DM）
func (b *ContextBuilder) buildSystemPrompt() string {
	return "你是 DND 游戏的 DM(地下城主)。负责描述场景、扮演 NPC、判断规则、推进剧情。"
}

// loadHistory 加载历史消息（保留用于未来扩展）
// 未来可以根据不同的策略加载历史消息，如：
// - 按时间范围加载（如最近 24 小时）
// - 按消息类型加载
// - 智能截断（根据 Token 数量）
// - 摘要压缩（将旧消息压缩为摘要）
func (b *ContextBuilder) loadHistory(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	return b.messageStore.List(ctx, sessionID, limit)
}

// contains 检查字符串数组是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
