package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageRole 消息角色
type MessageRole string

const (
	// MessageRoleUser 用户消息
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant AI 助手消息
	MessageRoleAssistant MessageRole = "assistant"
	// MessageRoleSystem 系统消息
	MessageRoleSystem MessageRole = "system"
)

// Message 对话消息
type Message struct {
	ID         string      `json:"id"`          // UUID
	CampaignID string      `json:"campaign_id"` // 所属战役ID
	Role       MessageRole `json:"role"`        // 角色（user, assistant, system）
	Content    string      `json:"content"`     // 消息内容
	PlayerID   string      `json:"player_id"`   // 玩家ID（user消息）
	ToolCalls  []ToolCall  `json:"tool_calls"`  // 工具调用（assistant消息）
	CreatedAt  time.Time   `json:"created_at"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID        string                 `json:"id"`        // 调用ID
	Name      string                 `json:"name"`      // 工具名称
	Arguments map[string]interface{} `json:"arguments"` // 参数
	Result    *ToolResult            `json:"result"`    // 执行结果
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success bool                   `json:"success"` // 是否成功
	Data    map[string]interface{} `json:"data"`    // 返回数据
	Error   string                 `json:"error"`   // 错误信息
}

// NewMessage 创建新消息
func NewMessage(campaignID string, role MessageRole, content string) *Message {
	return &Message{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Role:       role,
		Content:    content,
		CreatedAt:  time.Now(),
	}
}

// NewUserMessage 创建用户消息
func NewUserMessage(campaignID, playerID, content string) *Message {
	msg := NewMessage(campaignID, MessageRoleUser, content)
	msg.PlayerID = playerID
	return msg
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(campaignID, content string, toolCalls []ToolCall) *Message {
	msg := NewMessage(campaignID, MessageRoleAssistant, content)
	msg.ToolCalls = toolCalls
	return msg
}

// Validate 验证消息
func (m *Message) Validate() error {
	if m.CampaignID == "" {
		return NewValidationError("campaign_id", "cannot be empty")
	}

	// 验证角色
	switch m.Role {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem:
		// 有效角色
	default:
		return NewValidationError("role", "must be one of: user, assistant, system")
	}

	// 验证内容长度（1-100000字符）
	if len(m.Content) < 1 {
		return NewValidationError("content", "cannot be empty")
	}
	if len(m.Content) > 100000 {
		return NewValidationError("content", "cannot exceed 100000 characters")
	}

	// user 消息必须有 player_id
	if m.Role == MessageRoleUser && m.PlayerID == "" {
		return NewValidationError("player_id", "is required for user messages")
	}

	// 验证 tool_calls（仅 assistant 消息可以有 tool_calls）
	if len(m.ToolCalls) > 0 && m.Role != MessageRoleAssistant {
		return NewValidationError("tool_calls", "only allowed for assistant messages")
	}

	// 验证每个 tool_call
	for i, tc := range m.ToolCalls {
		if tc.ID == "" {
			return NewValidationError("tool_calls["+string(rune(i))+"].id", "cannot be empty")
		}
		if tc.Name == "" {
			return NewValidationError("tool_calls["+string(rune(i))+"].name", "cannot be empty")
		}
	}

	return nil
}

// AddToolCall 添加工具调用
func (m *Message) AddToolCall(id, name string, arguments map[string]interface{}) {
	m.ToolCalls = append(m.ToolCalls, ToolCall{
		ID:        id,
		Name:      name,
		Arguments: arguments,
	})
}

// SetToolResult 设置工具执行结果
func (m *Message) SetToolResult(toolCallID string, result *ToolResult) {
	for i := range m.ToolCalls {
		if m.ToolCalls[i].ID == toolCallID {
			m.ToolCalls[i].Result = result
			break
		}
	}
}

// HasToolCalls 检查是否有工具调用
func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

// IsFromPlayer 检查是否来自玩家
func (m *Message) IsFromPlayer() bool {
	return m.Role == MessageRoleUser && m.PlayerID != ""
}

// IsSystem 检查是否为系统消息
func (m *Message) IsSystem() bool {
	return m.Role == MessageRoleSystem
}
