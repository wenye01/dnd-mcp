// Package store 提供数据持久化接口
package store

import (
	"context"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/google/uuid"
)

// SessionStore 会话存储接口
type SessionStore interface {
	// CreateSession 创建新会话
	CreateSession(ctx context.Context, session *models.Session) error

	// GetSession 根据ID获取会话
	GetSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)

	// ListSessions 列出会话
	ListSessions(ctx context.Context, limit, offset int) ([]*models.Session, error)

	// UpdateSession 更新会话
	UpdateSession(ctx context.Context, session *models.Session) error

	// DeleteSession 删除会话(软删除)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}

// MessageStore 消息存储接口
type MessageStore interface {
	// CreateMessage 创建新消息
	CreateMessage(ctx context.Context, message *models.Message) error

	// GetMessages 获取指定会话的消息列表
	GetMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*models.Message, error)

	// GetRecentMessages 获取最近的消息(用于上下文构建)
	GetRecentMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.Message, error)
}

// Store 组合存储接口
type Store interface {
	SessionStore
	MessageStore
}
