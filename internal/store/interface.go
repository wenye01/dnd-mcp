// Package store 提供数据存储接口定义
package store

import (
	"context"
	"time"

	"github.com/dnd-mcp/client/internal/models"
)

// SessionStore 会话存储接口
type SessionStore interface {
	// Create 创建会话
	Create(ctx context.Context, session *models.Session) error

	// Get 获取会话
	Get(ctx context.Context, id string) (*models.Session, error)

	// List 列出所有会话
	List(ctx context.Context) ([]*models.Session, error)

	// Update 更新会话
	Update(ctx context.Context, session *models.Session) error

	// Delete 删除会话(软删除)
	Delete(ctx context.Context, id string) error
}

// MessageStore 消息存储接口
type MessageStore interface {
	// Create 保存消息
	Create(ctx context.Context, message *models.Message) error

	// Get 获取消息
	Get(ctx context.Context, sessionID, messageID string) (*models.Message, error)

	// List 获取会话消息列表
	List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error)

	// ListByRole 按角色获取消息
	ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error)

	// ListSince 按时间范围获取消息
	ListSince(ctx context.Context, sessionID string, since time.Time, limit int) ([]*models.Message, error)
}
