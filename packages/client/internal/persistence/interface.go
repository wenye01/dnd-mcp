// Package persistence 提供持久化服务
// 在使用方定义接口，遵循规范要求
package persistence

import (
	"context"
	"github.com/dnd-mcp/client/internal/models"
)

// SessionWriter 会话写入接口（在使用方定义）
type SessionWriter interface {
	// Create 创建会话
	Create(ctx context.Context, session *models.Session) error

	// BatchCreate 批量创建会话
	BatchCreate(ctx context.Context, sessions []*models.Session) error

	// Update 更新会话
	Update(ctx context.Context, session *models.Session) error
}

// SessionReader 会话读取接口
type SessionReader interface {
	// Get 获取会话
	Get(ctx context.Context, id string) (*models.Session, error)

	// List 列出所有会话
	List(ctx context.Context) ([]*models.Session, error)

	// ListActive 列出活跃会话（软删除过滤）
	ListActive(ctx context.Context) ([]*models.Session, error)
}

// MessageWriter 消息写入接口
type MessageWriter interface {
	// Create 创建消息
	Create(ctx context.Context, message *models.Message) error

	// BatchCreate 批量创建消息
	BatchCreate(ctx context.Context, messages []*models.Message) error
}

// MessageReader 消息读取接口
type MessageReader interface {
	// Get 获取消息
	Get(ctx context.Context, sessionID, messageID string) (*models.Message, error)

	// List 获取消息列表
	List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error)

	// ListByRole 按角色获取消息
	ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error)
}
