package models

import (
	"time"

	"github.com/google/uuid"
)

// Message 对话消息
type Message struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	SessionID uuid.UUID              `json:"session_id" db:"session_id"`
	Role      string                 `json:"role" db:"role"` // user|assistant|system|tool
	Content   string                 `json:"content" db:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty" db:"tool_calls"`
	PlayerID  string                 `json:"player_id,omitempty" db:"player_id"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// NewMessage 创建新消息
func NewMessage(sessionID uuid.UUID, role, content string) *Message {
	return &Message{
		ID:        uuid.New(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		ToolCalls: make([]ToolCall, 0),
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}
