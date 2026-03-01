// Package repository 提供数据访问接口定义
// 遵循"在使用方定义接口"原则，接口定义在这里，store 层实现这些接口
package repository

import (
	"context"
	"time"

	"github.com/dnd-mcp/client/internal/models"
)

// MessageRepository 消息数据访问接口
// 定义消息的数据访问操作，由具体存储实现（如 Redis、PostgreSQL）
type MessageRepository interface {
	// Create 保存消息
	Create(ctx context.Context, message *models.Message) error

	// Get 获取消息
	Get(ctx context.Context, sessionID, messageID string) (*models.Message, error)

	// List 获取消息列表
	List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error)

	// ListSince 按时间范围获取消息
	ListSince(ctx context.Context, sessionID string, since time.Time, limit int) ([]*models.Message, error)
}

// 注意: store.MessageStore 包含 MessageRepository 的所有方法，
// 因此任何实现 store.MessageStore 的类型都自动实现 MessageRepository。
// 编译时检查在 store 层完成（参见 internal/store/redis/message.go）
