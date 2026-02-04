// Package persistence 提供持久化服务
package persistence

import (
	"context"
	"fmt"
	"time"
)

// BackupService 备份服务
// 从 Redis 备份数据到 PostgreSQL
type BackupService struct {
	redisSessionReader    SessionReader
	redisMessageReader     MessageReader
	postgresSessionWriter  SessionWriter
	postgresMessageWriter  MessageWriter
}

// BackupResult 备份结果
type BackupResult struct {
	SessionCount int       // 备份的会话数量
	MessageCount int       // 备份的消息数量
	Duration     time.Duration // 备份耗时
	StartTime    time.Time // 开始时间
	EndTime      time.Time // 结束时间
}

// NewBackupService 创建备份服务
func NewBackupService(
	redisSessionReader SessionReader,
	redisMessageReader MessageReader,
	postgresSessionWriter SessionWriter,
	postgresMessageWriter MessageWriter,
) *BackupService {
	return &BackupService{
		redisSessionReader:   redisSessionReader,
		redisMessageReader:    redisMessageReader,
		postgresSessionWriter: postgresSessionWriter,
		postgresMessageWriter: postgresMessageWriter,
	}
}

// BackupAll 备份所有数据
func (s *BackupService) BackupAll(ctx context.Context) (*BackupResult, error) {
	startTime := time.Now()
	fmt.Println("开始备份所有数据...")

	// 1. 备份会话
	fmt.Println("正在备份会话...")
	sessionCount, err := s.backupSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("备份会话失败: %w", err)
	}
	fmt.Printf("✓ 已备份 %d 个会话\n", sessionCount)

	// 2. 备份消息
	fmt.Println("正在备份消息...")
	messageCount, err := s.backupMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("备份消息失败: %w", err)
	}
	fmt.Printf("✓ 已备份 %d 条消息\n", messageCount)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	result := &BackupResult{
		SessionCount: sessionCount,
		MessageCount: messageCount,
		Duration:     duration,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	fmt.Printf("\n备份完成！耗时: %v\n", duration)
	return result, nil
}

// BackupSession 备份指定会话
func (s *BackupService) BackupSession(ctx context.Context, sessionID string) (*BackupResult, error) {
	startTime := time.Now()
	fmt.Printf("开始备份会话: %s\n", sessionID)

	// 1. 备份会话元数据
	fmt.Println("正在备份会话元数据...")
	session, err := s.redisSessionReader.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	if err := s.postgresSessionWriter.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("备份会话失败: %w", err)
	}
	fmt.Printf("✓ 已备份会话: %s\n", session.Name)

	// 2. 备份该会话的所有消息
	fmt.Println("正在备份会话消息...")
	messages, err := s.redisMessageReader.List(ctx, sessionID, 0) // 0 = 无限制
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	if len(messages) > 0 {
		if err := s.postgresMessageWriter.BatchCreate(ctx, messages); err != nil {
			return nil, fmt.Errorf("备份消息失败: %w", err)
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	result := &BackupResult{
		SessionCount: 1,
		MessageCount: len(messages),
		Duration:     duration,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	fmt.Printf("✓ 已备份 %d 条消息\n", len(messages))
	fmt.Printf("\n备份完成！耗时: %v\n", duration)
	return result, nil
}

// backupSessions 备份所有会话
func (s *BackupService) backupSessions(ctx context.Context) (int, error) {
	// 从 Redis 读取所有会话
	sessions, err := s.redisSessionReader.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("从 Redis 读取会话失败: %w", err)
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	// 批量写入 PostgreSQL
	if err := s.postgresSessionWriter.BatchCreate(ctx, sessions); err != nil {
		return 0, fmt.Errorf("写入 PostgreSQL 失败: %w", err)
	}

	return len(sessions), nil
}

// backupMessages 备份所有消息
func (s *BackupService) backupMessages(ctx context.Context) (int, error) {
	// 获取所有会话 ID
	sessions, err := s.redisSessionReader.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取会话列表失败: %w", err)
	}

	totalMessages := 0
	batchSize := 100 // 每批处理 100 条消息

	for _, session := range sessions {
		// 获取该会话的所有消息（分批处理）
		offset := 0
		for {
			messages, err := s.redisMessageReader.List(ctx, session.ID, batchSize)
			if err != nil {
				return 0, fmt.Errorf("读取会话 %s 的消息失败: %w", session.ID, err)
			}

			if len(messages) == 0 {
				break
			}

			// 批量写入 PostgreSQL
			if err := s.postgresMessageWriter.BatchCreate(ctx, messages); err != nil {
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
