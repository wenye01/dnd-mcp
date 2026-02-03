// Package models 提供领域模型定义
package models

import (
	"time"

	"github.com/google/uuid"
)

// Message 消息模型
type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	PlayerID  string    `json:"player_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NewMessage 创建新消息
func NewMessage(sessionID, role, content string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// NewUserMessage 创建用户消息
func NewUserMessage(sessionID, content, playerID string) *Message {
	return &Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
		PlayerID:  playerID,
		CreatedAt: time.Now(),
	}
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage(sessionID, content string) *Message {
	return &Message{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// NewSystemMessage 创建系统消息
func NewSystemMessage(sessionID, content string) *Message {
	return &Message{
		SessionID: sessionID,
		Role:      "system",
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// HasToolCalls 检查消息是否包含工具调用
func (m *Message) HasToolCalls() bool {
	return len(m.ToolCalls) > 0
}

// AddToolCall 添加工具调用
func (m *Message) AddToolCall(toolCall ToolCall) {
	m.ToolCalls = append(m.ToolCalls, toolCall)
}
