// Package models 提供领域模型定义
package models

import (
	"time"
)

// Session 会话模型
// 注意: Session 与 Campaign 是同义词，Client 保留 Session 术语以保持 API 兼容性
// Server 使用 Campaign 术语，Client 内部会进行映射
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	CreatorID    string                 `json:"creator_id"`
	MCPServerURL string                 `json:"mcp_server_url"`
	WebSocketKey string                 `json:"websocket_key"`
	MaxPlayers   int                    `json:"max_players"`
	Settings     map[string]interface{} `json:"settings"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    time.Time              `json:"deleted_at,omitempty"` // 软删除时间
	Status       string                 `json:"status"`
}

// Campaign 是 Session 的类型别名，保持 API 兼容性
// 推荐在代码中使用 Campaign 术语，API 层会自动映射
type Campaign = Session

// CampaignID 返回 Campaign ID（等同于 Session.ID）
func (s *Session) CampaignID() string {
	return s.ID
}

// NewSession 创建新会话
func NewSession(name, creatorID, mcpServerURL string) *Session {
	now := time.Now()
	return &Session{
		Name:         name,
		CreatorID:    creatorID,
		MCPServerURL: mcpServerURL,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
		MaxPlayers:   4, // 默认最大4个玩家
		Settings:     make(map[string]interface{}),
	}
}

// IsActive 检查会话是否活跃
func (s *Session) IsActive() bool {
	return s.Status == "active"
}

// Archive 归档会话
func (s *Session) Archive() {
	s.Status = "archived"
	s.UpdatedAt = time.Now()
}

// UpdateSettings 更新设置
func (s *Session) UpdateSettings(settings map[string]interface{}) {
	if s.Settings == nil {
		s.Settings = make(map[string]interface{})
	}
	for k, v := range settings {
		s.Settings[k] = v
	}
	s.UpdatedAt = time.Now()
}
