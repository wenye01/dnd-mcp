// Package persistence 提供持久化管理器
package persistence

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/persistence/trigger"
	"github.com/dnd-mcp/client/internal/store"
)

// PostgresStore PostgreSQL存储接口（简化版）
type PostgresStore interface {
	UpsertSession(ctx context.Context, session *models.Session) error
	BatchInsertMessages(ctx context.Context, messages []*models.Message) error
}

// Manager 持久化管理器
type Manager struct {
	trigger       trigger.PersistenceTrigger // 触发器
	sessionStore  store.SessionStore         // 会话存储
	messageStore  store.MessageStore         // 消息存储
	postgresStore PostgresStore              // PostgreSQL 存储（可选）
}

// NewManager 创建持久化管理器
func NewManager(
	trig trigger.PersistenceTrigger,
	sessionStore store.SessionStore,
	messageStore store.MessageStore,
	postgresStore PostgresStore,
) *Manager {
	return &Manager{
		trigger:       trig,
		sessionStore:  sessionStore,
		messageStore:  messageStore,
		postgresStore: postgresStore,
	}
}

// Start 启动持久化管理器
func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // 每秒检查一次
	defer ticker.Stop()

	log.Printf("[持久化] 管理器已启动,触发器: %s", m.trigger.Name())

	for {
		select {
		case <-ctx.Done():
			log.Println("[持久化] 管理器已停止")
			return

		case <-ticker.C:
			// 检查是否应该触发持久化
			shouldTrigger, err := m.trigger.ShouldTrigger(ctx)
			if err != nil {
				log.Printf("[持久化] 检查触发器失败: %v", err)
				continue
			}

			if shouldTrigger {
				log.Printf("[持久化] 触发器满足条件: %s", m.trigger.Name())
				m.persist(ctx)

				// 重置触发器
				if err := m.trigger.Reset(ctx); err != nil {
					log.Printf("[持久化] 重置触发器失败: %v", err)
				}
			}
		}
	}
}

// Trigger 手动触发持久化
func (m *Manager) Trigger(ctx context.Context) error {
	log.Println("[持久化] 手动触发持久化")
	return m.persist(ctx)
}

// persist 执行持久化
func (m *Manager) persist(ctx context.Context) error {
	start := time.Now()

	log.Println("[持久化] 开始持久化...")

	// 简化实现: 仅记录日志,不实际连接PostgreSQL
	// 实际项目中需要检查PostgreSQL是否可用

	// 获取所有会话
	sessions, err := m.sessionStore.List(ctx)
	if err != nil {
		return fmt.Errorf("获取会话列表失败: %w", err)
	}

	log.Printf("[持久化] 找到 %d 个会话", len(sessions))

	sessionCount := 0
	messageCount := 0

	for _, session := range sessions {
		// 备份会话到 PostgreSQL (如果可用)
		if m.postgresStore != nil {
			if err := m.postgresStore.UpsertSession(ctx, session); err != nil {
				log.Printf("[持久化] 备份会话 %s 失败: %v", session.ID, err)
				continue
			}
		}

		sessionCount++

		// 读取消息
		messages, err := m.messageStore.List(ctx, session.ID, 100)
		if err != nil {
			log.Printf("[持久化] 读取会话 %s 的消息失败: %v", session.ID, err)
			continue
		}

		// 备份消息到 PostgreSQL (如果可用)
		if m.postgresStore != nil {
			if err := m.postgresStore.BatchInsertMessages(ctx, messages); err != nil {
				log.Printf("[持久化] 备份会话 %s 的消息失败: %v", session.ID, err)
				continue
			}
		}

		messageCount += len(messages)
	}

	duration := time.Since(start)

	log.Printf("[持久化] 完成: %d 个会话, %d 条消息, 耗时 %s",
		sessionCount, messageCount, duration)

	return nil
}
