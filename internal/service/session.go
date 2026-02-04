// Package service 提供业务逻辑层实现
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/dnd-mcp/client/internal/models"
	"github.com/dnd-mcp/client/internal/repository"
	"github.com/dnd-mcp/client/pkg/errors"
)

// SessionServiceInterface 会话服务接口
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	ListSessions(ctx context.Context, status string) ([]*models.Session, error)
	UpdateSession(ctx context.Context, sessionID string, req *UpdateSessionRequest) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// SessionService 会话业务逻辑
type SessionService struct {
	sessionRepo repository.SessionRepository
}

// NewSessionService 创建会话服务
func NewSessionService(sessionRepo repository.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Name         string                 `json:"name" binding:"required"`
	CreatorID    string                 `json:"creator_id" binding:"required"`
	MCPServerURL string                 `json:"mcp_server_url" binding:"required,url"`
	MaxPlayers   int                    `json:"max_players"`
	Settings     map[string]interface{} `json:"settings"`
}

// CreateSession 创建会话
func (s *SessionService) CreateSession(ctx context.Context, req *CreateSessionRequest) (*models.Session, error) {
	// 参数验证
	if req.MaxPlayers < 1 || req.MaxPlayers > 10 {
		return nil, errors.ErrInvalidMaxPlayers
	}

	// 默认值
	if req.MaxPlayers == 0 {
		req.MaxPlayers = 4
	}

	// 生成会话ID和WebSocket密钥
	sessionID := uuid.New().String()
	wsKey := s.generateWebSocketKey()

	session := &models.Session{
		ID:           sessionID,
		Name:         req.Name,
		CreatorID:    req.CreatorID,
		MCPServerURL: req.MCPServerURL,
		WebSocketKey: wsKey,
		MaxPlayers:   req.MaxPlayers,
		Settings:     req.Settings,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       "active",
	}

	// 保存到存储
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// GetSession 获取会话详情
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// ListSessions 列出所有会话
func (s *SessionService) ListSessions(ctx context.Context, status string) ([]*models.Session, error) {
	sessions, err := s.sessionRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// 过滤状态
	if status != "" && status != "all" {
		var filtered []*models.Session
		for _, session := range sessions {
			if session.Status == status {
				filtered = append(filtered, session)
			}
		}
		return filtered, nil
	}

	return sessions, nil
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Name       *string
	MaxPlayers *int
	Settings   map[string]interface{}
}

// UpdateSession 更新会话
func (s *SessionService) UpdateSession(ctx context.Context, sessionID string, req *UpdateSessionRequest) (*models.Session, error) {
	// 获取现有会话
	session, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 部分更新
	if req.Name != nil {
		session.Name = *req.Name
	}
	if req.MaxPlayers != nil {
		if *req.MaxPlayers < 1 || *req.MaxPlayers > 10 {
			return nil, errors.ErrInvalidMaxPlayers
		}
		session.MaxPlayers = *req.MaxPlayers
	}
	if req.Settings != nil {
		session.Settings = req.Settings
	}
	session.UpdatedAt = time.Now()

	// 保存更新
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("更新会话失败: %w", err)
	}

	return session, nil
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	// 检查会话是否存在
	_, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// 软删除
	if err := s.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}

	return nil
}

// generateWebSocketKey 生成 WebSocket 密钥
func (s *SessionService) generateWebSocketKey() string {
	return "ws-" + uuid.New().String()
}
