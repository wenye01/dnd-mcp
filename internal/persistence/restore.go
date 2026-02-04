// Package persistence 提供持久化服务
package persistence

import (
	"context"
	"fmt"
	"time"
)

// RestoreService 恢复服务
// 从 PostgreSQL 恢复数据到 Redis
type RestoreService struct {
	postgresSessionReader SessionReader
	postgresMessageReader  MessageReader
	redisSessionReader    SessionReader // 添加 Reader 用于检查是否存在
	redisSessionWriter     SessionWriter
	redisMessageWriter      MessageWriter
}

// RestoreResult 恢复结果
type RestoreResult struct {
	SessionCount int       // 恢复的会话数量
	MessageCount int       // 恢复的消息数量
	SkippedCount int       // 跳过的数量（已存在）
	Duration     time.Duration // 恢复耗时
	StartTime    time.Time // 开始时间
	EndTime      time.Time // 结束时间
}

// NewRestoreService 创建恢复服务
func NewRestoreService(
	postgresSessionReader SessionReader,
	postgresMessageReader MessageReader,
	redisSessionReader SessionReader,
	redisSessionWriter SessionWriter,
	redisMessageWriter MessageWriter,
) *RestoreService {
	return &RestoreService{
		postgresSessionReader: postgresSessionReader,
		postgresMessageReader: postgresMessageReader,
		redisSessionReader:    redisSessionReader,
		redisSessionWriter:    redisSessionWriter,
		redisMessageWriter:    redisMessageWriter,
	}
}

// RestoreAll 恢复所有数据
func (s *RestoreService) RestoreAll(ctx context.Context, force bool) (*RestoreResult, error) {
	startTime := time.Now()
	fmt.Println("开始恢复所有数据...")

	// 1. 恢复会话
	fmt.Println("正在恢复会话...")
	sessionCount, skippedCount, err := s.restoreSessions(ctx, force)
	if err != nil {
		return nil, fmt.Errorf("恢复会话失败: %w", err)
	}
	fmt.Printf("✓ 已恢复 %d 个会话", sessionCount)
	if skippedCount > 0 {
		fmt.Printf("（跳过 %d 个已存在的会话）", skippedCount)
	}
	fmt.Println()

	// 2. 恢复消息
	fmt.Println("正在恢复消息...")
	messageCount, err := s.restoreMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("恢复消息失败: %w", err)
	}
	fmt.Printf("✓ 已恢复 %d 条消息\n", messageCount)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	result := &RestoreResult{
		SessionCount: sessionCount,
		MessageCount: messageCount,
		SkippedCount: skippedCount,
		Duration:     duration,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	fmt.Printf("\n恢复完成！耗时: %v\n", duration)
	return result, nil
}

// RestoreSession 恢复指定会话
func (s *RestoreService) RestoreSession(ctx context.Context, sessionID string, force bool) (*RestoreResult, error) {
	startTime := time.Now()
	fmt.Printf("开始恢复会话: %s\n", sessionID)

	// 1. 检查是否已存在
	if !force {
		_, err := s.redisSessionReader.Get(ctx, sessionID)
		if err == nil {
			fmt.Printf("会话已存在，跳过恢复（使用 --force 强制覆盖）\n")
			return &RestoreResult{
				SessionCount: 0,
				SkippedCount: 1,
				Duration:     time.Since(startTime),
				StartTime:    startTime,
				EndTime:      time.Now(),
			}, nil
		}
	}

	// 2. 从 PostgreSQL 读取会话
	fmt.Println("正在从 PostgreSQL 读取会话...")
	session, err := s.postgresSessionReader.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("从 PostgreSQL 获取会话失败: %w", err)
	}

	// 3. 写入 Redis
	if err := s.redisSessionWriter.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("恢复会话失败: %w", err)
	}
	fmt.Printf("✓ 已恢复会话: %s\n", session.Name)

	// 4. 恢复该会话的所有消息
	fmt.Println("正在恢复会话消息...")
	messages, err := s.postgresMessageReader.List(ctx, sessionID, 0) // 0 = 无限制
	if err != nil {
		return nil, fmt.Errorf("从 PostgreSQL 获取消息失败: %w", err)
	}

	if len(messages) > 0 {
		// 批量恢复消息（消息不需要检查 force，因为使用 ON CONFLICT DO NOTHING）
		if err := s.redisMessageWriter.BatchCreate(ctx, messages); err != nil {
			return nil, fmt.Errorf("恢复消息失败: %w", err)
		}
		fmt.Printf("✓ 已恢复 %d 条消息\n", len(messages))
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	result := &RestoreResult{
		SessionCount: 1,
		MessageCount: len(messages),
		SkippedCount: 0,
		Duration:     duration,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	fmt.Printf("\n恢复完成！耗时: %v\n", duration)
	return result, nil
}

// restoreSessions 恢复所有会话
func (s *RestoreService) restoreSessions(ctx context.Context, force bool) (int, int, error) {
	// 从 PostgreSQL 读取所有活跃会话（过滤软删除）
	sessions, err := s.postgresSessionReader.ListActive(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("从 PostgreSQL 读取会话失败: %w", err)
	}

	if len(sessions) == 0 {
		return 0, 0, nil
	}

	restored := 0
	skipped := 0

	for _, session := range sessions {
		// 检查是否已存在
		if !force {
			_, err := s.redisSessionReader.Get(ctx, session.ID)
			if err == nil {
				skipped++
				continue
			}
		}

		// 写入 Redis
		if err := s.redisSessionWriter.Create(ctx, session); err != nil {
			return 0, 0, fmt.Errorf("恢复会话 %s 失败: %w", session.ID, err)
		}

		restored++
	}

	return restored, skipped, nil
}

// restoreMessages 恢复所有消息
func (s *RestoreService) restoreMessages(ctx context.Context) (int, error) {
	// 获取所有会话 ID
	sessions, err := s.postgresSessionReader.ListActive(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取会话列表失败: %w", err)
	}

	totalMessages := 0
	batchSize := 100 // 每批处理 100 条消息

	for _, session := range sessions {
		// 获取该会话的所有消息（分批处理）
		offset := 0
		for {
			messages, err := s.postgresMessageReader.List(ctx, session.ID, batchSize)
			if err != nil {
				return 0, fmt.Errorf("读取会话 %s 的消息失败: %w", session.ID, err)
			}

			if len(messages) == 0 {
				break
			}

			// 批量写入 Redis
			if err := s.redisMessageWriter.BatchCreate(ctx, messages); err != nil {
				return 0, fmt.Errorf("写入会话 %s 的消息失败: %w", session.ID, err)
			}

			totalMessages += len(messages)
			offset += len(messages)

			if len(messages) < batchSize {
				break // 没有更多消息了
			}
		}
	}

	return totalMessages, nil
}
