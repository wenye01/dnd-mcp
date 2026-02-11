// Package repository 提供数据访问接口定义
// 遵循"在使用方定义接口"原则，接口定义在这里，store 层实现这些接口
package repository

import (
	"context"
	"github.com/dnd-mcp/client/internal/models"
)

// SessionRepository 会话数据访问接口
// 定义会话的数据访问操作，由具体存储实现（如 Redis、PostgreSQL）
type SessionRepository interface {
	// Create 创建会话
	Create(ctx context.Context, session *models.Session) error

	// Get 获取会话详情
	Get(ctx context.Context, id string) (*models.Session, error)

	// List 列出所有会话
	List(ctx context.Context) ([]*models.Session, error)

	// Update 更新会话
	Update(ctx context.Context, session *models.Session) error

	// Delete 删除会话（软删除）
	Delete(ctx context.Context, id string) error

	// Count 统计会话数量
	Count(ctx context.Context) (int64, error)
}
