// Package trigger 提供时间触发器
package trigger

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TimeTrigger 时间间隔触发器
type TimeTrigger struct {
	interval    time.Duration // 触发间隔
	lastTrigger time.Time     // 上次触发时间
	mu          sync.RWMutex  // 读写锁
}

// NewTimeTrigger 创建时间触发器
func NewTimeTrigger(interval time.Duration) *TimeTrigger {
	return &TimeTrigger{
		interval:    interval,
		lastTrigger: time.Time{}, // 零值表示未触发过
	}
}

// ShouldTrigger 判断是否应该触发
func (t *TimeTrigger) ShouldTrigger(ctx context.Context) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 首次触发
	if t.lastTrigger.IsZero() {
		return true, nil
	}

	// 检查是否超过间隔
	elapsed := time.Since(t.lastTrigger)
	return elapsed >= t.interval, nil
}

// Reset 重置触发器
func (t *TimeTrigger) Reset(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastTrigger = time.Now()
	return nil
}

// Name 返回触发器名称
func (t *TimeTrigger) Name() string {
	return fmt.Sprintf("TimeTrigger(interval=%s)", t.interval)
}
