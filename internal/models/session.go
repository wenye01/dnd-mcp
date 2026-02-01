package models

import (
	"time"

	"github.com/google/uuid"
)

// Session 游戏会话
type Session struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Version   int       `json:"version" db:"version"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// 游戏元数据
	GameTime      string `json:"game_time" db:"game_time"`
	Location      string `json:"location" db:"location"`
	CampaignName  string `json:"campaign_name" db:"campaign_name"`

	// 战斗状态
	Combat *CombatState `json:"combat,omitempty"`

	// 地图状态
	Map *MapState `json:"map,omitempty"`

	// 角色列表
	Characters []*Character `json:"characters,omitempty"`

	// 完整状态(用于存储到state字段)
	State map[string]interface{} `json:"state,omitempty" db:"state"`
}

// CombatState 战斗状态
type CombatState struct {
	Active          bool     `json:"active"`
	Round           int      `json:"round"`
	CurrentTurn     string   `json:"current_turn"` // character_id
	InitiativeOrder []string `json:"initiative_order"`
}

// MapState 地图状态
type MapState struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	GridSize int    `json:"grid_size"`
}

// NewSession 创建新会话
func NewSession(campaignName string) *Session {
	now := time.Now()
	return &Session{
		ID:           uuid.New(),
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
		CampaignName: campaignName,
		GameTime:     "Morning",
		Location:     "Unknown",
		Combat:       nil,
		Map:          nil,
		Characters:   make([]*Character, 0),
		State:        make(map[string]interface{}),
	}
}
