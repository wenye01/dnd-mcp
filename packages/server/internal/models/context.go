// Package models 提供领域模型定义
package models

import "time"

// Context LLM上下文
type Context struct {
	CampaignID      string        `json:"campaign_id"`       // 战役ID
	GameSummary     *GameSummary  `json:"game_summary"`      // 游戏状态摘要
	Messages        []Message     `json:"messages"`          // 对话历史（压缩后）
	RawMessageCount int           `json:"raw_message_count"` // 原始消息总数
	TokenEstimate   int           `json:"token_estimate"`    // 预估token数
	CreatedAt       time.Time     `json:"created_at"`
}

// NewContext 创建新上下文
func NewContext(campaignID string) *Context {
	return &Context{
		CampaignID: campaignID,
		Messages:   make([]Message, 0),
		CreatedAt:  time.Now(),
	}
}

// GetRawContextResponse 原始上下文响应
type GetRawContextResponse struct {
	CampaignID  string      `json:"campaign_id"`
	GameState   *GameState  `json:"game_state,omitempty"`   // 游戏状态
	Characters  []*Character `json:"characters,omitempty"`   // 队伍成员
	Combat      *Combat     `json:"combat,omitempty"`        // 当前战斗
	Map         *Map        `json:"map,omitempty"`           // 当前地图
	Messages    []*Message  `json:"messages"`               // 完整消息列表
	MessageCount int        `json:"message_count"`
}

// GameSummary 游戏状态摘要
type GameSummary struct {
	Time     string        `json:"time"`     // 游戏时间描述
	Location string        `json:"location"` // 当前位置描述
	Weather  string        `json:"weather"`  // 天气
	InCombat bool          `json:"in_combat"` // 是否在战斗中
	Party    []PartyMember `json:"party"`    // 队伍成员
	Combat   *CombatSummary `json:"combat,omitempty"` // 战斗摘要（如果适用）
}

// PartyMember 队伍成员摘要
type PartyMember struct {
	ID    string `json:"id"`    // 角色ID
	Name  string `json:"name"`  // 角色名称
	HP    string `json:"hp"`    // HP状态描述
	Class string `json:"class"` // 职业
}

// CombatSummary 战斗摘要
type CombatSummary struct {
	Round        int      `json:"round"`        // 当前回合
	TurnIndex    int      `json:"turn_index"`   // 当前行动者索引
	Participants []string `json:"participants"` // 参战者名称列表
}
