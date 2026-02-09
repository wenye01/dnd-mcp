// Package redis 提供 Redis 存储实现
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/dnd-mcp/client/internal/store"
	"github.com/dnd-mcp/client/pkg/errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// messageStore 消息存储实现
type messageStore struct {
	client Client
}

// NewMessageStore 创建消息存储实例
func NewMessageStore(client Client) store.MessageStore {
	return &messageStore{client: client}
}

// Create 保存消息
func (m *messageStore) Create(ctx context.Context, message *models.Message) error {
	// 如果没有ID,生成一个
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// 序列化消息
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 使用时间戳(毫秒)作为 score
	score := message.CreatedAt.UnixMilli()

	// 保存到 Sorted Set
	key := fmt.Sprintf("msg:%s", message.SessionID)
	err = m.client.Client().ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: messageJSON,
	}).Err()

	if err != nil {
		return fmt.Errorf("保存消息失败: %w", err)
	}

	return nil
}

// Get 获取消息
func (m *messageStore) Get(ctx context.Context, sessionID, messageID string) (*models.Message, error) {
	// 获取会话的所有消息
	key := fmt.Sprintf("msg:%s", sessionID)
	messages, err := m.client.Client().ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取消息失败: %w", err)
	}

	// 查找目标消息
	for _, msgJSON := range messages {
		var message models.Message
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			continue // 跳过解析失败的消息
		}

		if message.ID == messageID {
			return &message, nil
		}
	}

	return nil, errors.Wrap(errors.ErrMessageNotFound, fmt.Sprintf("message id: %s", messageID))
}

// List 获取会话消息列表
func (m *messageStore) List(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	key := fmt.Sprintf("msg:%s", sessionID)

	// 使用 ZREVRANGE 获取最近的消息(倒序,从新到旧)
	// 注意: ZRevRange 的 stop 参数是包含的,所以是 limit-1
	stop := int64(limit - 1)
	if limit <= 0 {
		stop = -1 // 获取全部
	}

	messagesJSON, err := m.client.Client().ZRevRange(ctx, key, 0, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("获取消息列表失败: %w", err)
	}

	// 如果没有消息,返回空列表
	if len(messagesJSON) == 0 {
		return []*models.Message{}, nil
	}

	// 解析消息
	messages := make([]*models.Message, 0, len(messagesJSON))
	for _, msgJSON := range messagesJSON {
		var message models.Message
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			continue // 跳过解析失败的消息
		}
		messages = append(messages, &message)
	}

	// 反转列表,使其按时间正序排列(从旧到新)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ListByRole 按角色获取消息
func (m *messageStore) ListByRole(ctx context.Context, sessionID, role string, limit int) ([]*models.Message, error) {
	// 获取所有消息
	allMessages, err := m.List(ctx, sessionID, -1) // 获取全部
	if err != nil {
		return nil, err
	}

	// 过滤指定角色的消息
	filtered := make([]*models.Message, 0)
	for _, msg := range allMessages {
		if msg.Role == role {
			filtered = append(filtered, msg)
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

// ListSince 按时间范围获取消息
func (m *messageStore) ListSince(ctx context.Context, sessionID string, since time.Time, limit int) ([]*models.Message, error) {
	key := fmt.Sprintf("msg:%s", sessionID)

	// 使用时间戳作为最小 score
	minScore := float64(since.UnixMilli())
	maxScore := float64(time.Now().UnixMilli())

	// 使用 ZRANGEBYSCORE 获取指定时间范围的消息
	messagesJSON, err := m.client.Client().ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", int64(minScore)),
		Max: fmt.Sprintf("%d", int64(maxScore)),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("获取消息列表失败: %w", err)
	}

	// 如果没有消息,返回空列表
	if len(messagesJSON) == 0 {
		return []*models.Message{}, nil
	}

	// 解析消息
	messages := make([]*models.Message, 0, len(messagesJSON))
	for _, msgJSON := range messagesJSON {
		var message models.Message
		if err := json.Unmarshal([]byte(msgJSON), &message); err != nil {
			continue // 跳过解析失败的消息
		}
		messages = append(messages, &message)
		if limit > 0 && len(messages) >= limit {
			break
		}
	}

	return messages, nil
}

// 确保实现了接口
var _ store.MessageStore = (*messageStore)(nil)
var _ persistence.MessageWriter = (*messageStore)(nil)
var _ persistence.MessageReader = (*messageStore)(nil)

// BatchCreate 批量创建消息（实现 persistence.MessageWriter）
func (m *messageStore) BatchCreate(ctx context.Context, messages []*models.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 使用 Pipeline 批量操作
	pipe := m.client.Client().Pipeline()

	for _, message := range messages {
		// 如果没有ID,生成一个
		if message.ID == "" {
			message.ID = uuid.New().String()
		}

		// 序列化消息
		messageJSON, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("序列化消息失败: %w", err)
		}

		// 使用时间戳(毫秒)作为 score
		score := message.CreatedAt.UnixMilli()

		// 保存到 Sorted Set
		key := fmt.Sprintf("msg:%s", message.SessionID)
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(score),
			Member: messageJSON,
		})
	}

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量创建消息失败: %w", err)
	}

	return nil
}
