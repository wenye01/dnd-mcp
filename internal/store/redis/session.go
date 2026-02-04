// Package redis 提供 Redis 存储实现
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/repository"
	"github.com/dnd-mcp/client/internal/persistence"
	"github.com/dnd-mcp/client/pkg/errors"
)

// sessionStore 会话存储实现
type sessionStore struct {
	client Client
}

// 确保 sessionStore 实现了 repository.SessionRepository 接口
var _ repository.SessionRepository = (*sessionStore)(nil)

// NewSessionStore 创建会话存储实例
// 返回 repository.SessionRepository 接口类型
func NewSessionStore(client Client) repository.SessionRepository {
	return &sessionStore{client: client}
}

// Create 创建会话
func (s *sessionStore) Create(ctx context.Context, session *models.Session) error {
	// 如果没有ID,生成一个
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	// 序列化 settings
	settingsJSON, err := json.Marshal(session.Settings)
	if err != nil {
		return fmt.Errorf("序列化 settings 失败: %w", err)
	}

	// 使用 Pipeline 批量操作
	pipe := s.client.Client().Pipeline()

	// 保存会话元数据到 Hash
	sessionKey := fmt.Sprintf("session:%s", session.ID)
	pipe.HSet(ctx, sessionKey, map[string]interface{}{
		"id":             session.ID,
		"name":           session.Name,
		"creator_id":     session.CreatorID,
		"mcp_server_url": session.MCPServerURL,
		"websocket_key":  session.WebSocketKey,
		"max_players":    session.MaxPlayers,
		"settings":       settingsJSON,
		"created_at":     session.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":     session.UpdatedAt.UTC().Format(time.RFC3339),
		"status":         session.Status,
	})

	// 添加到会话索引
	pipe.SAdd(ctx, "sessions:all", session.ID)

	// 执行 Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}

	return nil
}

// Get 获取会话
func (s *sessionStore) Get(ctx context.Context, id string) (*models.Session, error) {
	sessionKey := fmt.Sprintf("session:%s", id)

	// 从 Hash 获取所有字段
	data, err := s.client.Client().HGetAll(ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 检查会话是否存在
	if len(data) == 0 {
		return nil, errors.Wrap(errors.ErrSessionNotFound, fmt.Sprintf("session id: %s", id))
	}

	// 解析数据
	session, err := s.parseSession(data)
	if err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	return session, nil
}

// List 列出所有会话
func (s *sessionStore) List(ctx context.Context) ([]*models.Session, error) {
	// 获取所有会话 ID
	sessionIDs, err := s.client.Client().SMembers(ctx, "sessions:all").Result()
	if err != nil {
		return nil, fmt.Errorf("获取会话列表失败: %w", err)
	}

	// 如果没有会话,返回空列表
	if len(sessionIDs) == 0 {
		return []*models.Session{}, nil
	}

	// 批量获取会话数据
	sessions := make([]*models.Session, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		session, err := s.Get(ctx, id)
		if err != nil {
			// 跳过错误的会话,继续处理其他会话
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Update 更新会话
func (s *sessionStore) Update(ctx context.Context, session *models.Session) error {
	// 检查会话是否存在
	_, err := s.Get(ctx, session.ID)
	if err != nil {
		return err
	}

	// 更新 updated_at
	session.UpdatedAt = time.Now()

	// 序列化 settings
	settingsJSON, err := json.Marshal(session.Settings)
	if err != nil {
		return fmt.Errorf("序列化 settings 失败: %w", err)
	}

	// 更新 Hash
	sessionKey := fmt.Sprintf("session:%s", session.ID)
	_, err = s.client.Client().HSet(ctx, sessionKey, map[string]interface{}{
		"name":           session.Name,
		"max_players":    session.MaxPlayers,
		"settings":       settingsJSON,
		"updated_at":     session.UpdatedAt.UTC().Format(time.RFC3339),
		"status":         session.Status,
	}).Result()

	if err != nil {
		return fmt.Errorf("更新会话失败: %w", err)
	}

	return nil
}

// Delete 删除会话(软删除)
func (s *sessionStore) Delete(ctx context.Context, id string) error {
	// 检查会话是否存在
	session, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// 使用 Pipeline 批量操作
	pipe := s.client.Client().Pipeline()

	// 从索引中移除
	pipe.SRem(ctx, "sessions:all", id)

	// 删除会话数据
	sessionKey := fmt.Sprintf("session:%s", id)
	pipe.Del(ctx, sessionKey)

	// 执行 Pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}

	_ = session // 避免未使用变量警告
	return nil
}

// parseSession 从 Redis Hash 数据解析会话
func (s *sessionStore) parseSession(data map[string]string) (*models.Session, error) {
	// 解析时间
	createdAt, err := time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		return nil, fmt.Errorf("解析 created_at 失败: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, data["updated_at"])
	if err != nil {
		return nil, fmt.Errorf("解析 updated_at 失败: %w", err)
	}

	// 解析 settings
	var settings map[string]interface{}
	if data["settings"] != "" {
		err = json.Unmarshal([]byte(data["settings"]), &settings)
		if err != nil {
			return nil, fmt.Errorf("解析 settings 失败: %w", err)
		}
	}

	// 构建会话对象
	session := &models.Session{
		ID:           data["id"],
		Name:         data["name"],
		CreatorID:    data["creator_id"],
		MCPServerURL: data["mcp_server_url"],
		WebSocketKey: data["websocket_key"],
		Settings:     settings,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		Status:       data["status"],
	}

	// 解析 max_players (可选字段)
	if maxPlayers, ok := data["max_players"]; ok && maxPlayers != "" {
		_, err := fmt.Sscanf(maxPlayers, "%d", &session.MaxPlayers)
		if err != nil {
			session.MaxPlayers = 4 // 默认值
		}
	}

	return session, nil
}

// 确保实现了接口
var _ repository.SessionRepository = (*sessionStore)(nil)
var _ persistence.SessionWriter = (*sessionStore)(nil)
var _ persistence.SessionReader = (*sessionStore)(nil)

// BatchCreate 批量创建会话（实现 persistence.SessionWriter）
func (s *sessionStore) BatchCreate(ctx context.Context, sessions []*models.Session) error {
	if len(sessions) == 0 {
		return nil
	}

	// 使用 Pipeline 批量操作
	pipe := s.client.Client().Pipeline()

	for _, session := range sessions {
		// 如果没有ID,生成一个
		if session.ID == "" {
			session.ID = uuid.New().String()
		}

		// 序列化 settings
		settingsJSON, err := json.Marshal(session.Settings)
		if err != nil {
			return fmt.Errorf("序列化 settings 失败: %w", err)
		}

		// 保存会话元数据到 Hash
		sessionKey := fmt.Sprintf("session:%s", session.ID)
		pipe.HSet(ctx, sessionKey, map[string]interface{}{
			"id":             session.ID,
			"name":           session.Name,
			"creator_id":     session.CreatorID,
			"mcp_server_url": session.MCPServerURL,
			"websocket_key":  session.WebSocketKey,
			"max_players":    session.MaxPlayers,
			"settings":       settingsJSON,
			"created_at":     session.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":     session.UpdatedAt.UTC().Format(time.RFC3339),
			"status":         session.Status,
		})

		// 添加到会话索引
		pipe.SAdd(ctx, "sessions:all", session.ID)
	}

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量创建会话失败: %w", err)
	}

	return nil
}

// ListActive 列出活跃会话（实现 persistence.SessionReader）
// 在 Redis 中，所有在 sessions:all 集合中的都是活跃会话
func (s *sessionStore) ListActive(ctx context.Context) ([]*models.Session, error) {
	// 目前 Redis 没有软删除机制，ListActive 和 List 行为一致
	// 如果将来需要软删除，可以在 session hash 中添加 deleted_at 字段
	return s.List(ctx)
}
