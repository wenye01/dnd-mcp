// Package trigger 提供持久化触发器接口
package trigger

import "context"

// PersistenceTrigger 持久化触发器接口
type PersistenceTrigger interface {
	// ShouldTrigger 判断是否应该触发持久化
	ShouldTrigger(ctx context.Context) (bool, error)

	// Reset 重置触发器状态
	Reset(ctx context.Context) error

	// Name 返回触发器名称
	Name() string
}
