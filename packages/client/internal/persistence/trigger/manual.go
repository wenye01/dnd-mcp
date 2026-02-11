// Package trigger 提供手动触发器
package trigger

import (
	"context"
)

// ManualTrigger 手动触发器
type ManualTrigger struct {
	signal chan struct{} // 触发信号通道
}

// NewManualTrigger 创建手动触发器
func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{
		signal: make(chan struct{}, 1), // 缓冲大小 1,避免阻塞
	}
}

// Trigger 发送触发信号
func (m *ManualTrigger) Trigger() {
	select {
	case m.signal <- struct{}{}:
		// 信号已发送
	default:
		// 已有待处理的信号,忽略
	}
}

// ShouldTrigger 判断是否应该触发
func (m *ManualTrigger) ShouldTrigger(ctx context.Context) (bool, error) {
	select {
	case <-m.signal:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil
	}
}

// Reset 重置触发器
func (m *ManualTrigger) Reset(ctx context.Context) error {
	// 清空通道中的信号
	select {
	case <-m.signal:
	default:
	}
	return nil
}

// Name 返回触发器名称
func (m *ManualTrigger) Name() string {
	return "ManualTrigger"
}
