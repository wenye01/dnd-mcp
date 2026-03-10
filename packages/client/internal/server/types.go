// Package server 提供 Server API 客户端的类型定义
package server

import "time"

// Context 压缩后的上下文（与 Server 端 models.Context 对齐）
type Context struct {
	CampaignID      string        `json:"campaign_id"`
	GameSummary     *GameSummary  `json:"game_summary"`
	Messages        []Message     `json:"messages"`
	RawMessageCount int           `json:"raw_message_count"`
	TokenEstimate   int           `json:"token_estimate"`
	CreatedAt       time.Time     `json:"created_at"`
}

// RawContext 原始上下文（完整模式，与 Server 端 models.GetRawContextResponse 对齐）
type RawContext struct {
	CampaignID  string      `json:"campaign_id"`
	GameState   *GameState  `json:"game_state,omitempty"`
	Characters  []*Character `json:"characters,omitempty"`
	Combat      *Combat     `json:"combat,omitempty"`
	Map         *Map        `json:"map,omitempty"`
	Messages    []*Message  `json:"messages"`
	MessageCount int        `json:"message_count"`
}

// Message 对话消息（与 Server 端 models.Message 对齐）
type Message struct {
	ID         string      `json:"id"`
	CampaignID string      `json:"campaign_id"`
	Role       MessageRole `json:"role"`
	Content    string      `json:"content"`
	PlayerID   string      `json:"player_id"`
	ToolCalls  []ToolCall  `json:"tool_calls"`
	CreatedAt  time.Time   `json:"created_at"`
}

// MessageRole 消息角色（与 Server 端 models.MessageRole 对齐）
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// ToolCall 工具调用（与 Server 端 models.ToolCall 对齐）
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Result    *ToolResult            `json:"result,omitempty"`
}

// ToolResult 工具执行结果（与 Server 端 models.ToolResult 对齐）
type ToolResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error"`
}

// GameSummary 游戏状态摘要（与 Server 端 models.GameSummary 对齐）
type GameSummary struct {
	Time      string         `json:"time"`
	Location  string         `json:"location"`
	Weather   string         `json:"weather"`
	InCombat  bool           `json:"in_combat"`
	Party     []PartyMember  `json:"party"`
	Combat    *CombatSummary `json:"combat,omitempty"`
}

// PartyMember 队伍成员摘要
type PartyMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	HP    string `json:"hp"`
	Class string `json:"class"`
}

// CombatSummary 战斗摘要
type CombatSummary struct {
	Round        int      `json:"round"`
	TurnIndex    int      `json:"turn_index"`
	Participants []string `json:"participants"`
}

// GameState 游戏状态
type GameState struct {
	Location      string    `json:"location"`
	GameTime      string    `json:"game_time"`
	LastRestTime  time.Time `json:"last_rest_time"`
	ShortRests    int       `json:"short_rests"`
	LongRests     int       `json:"long_rests"`
	ActiveEffects []Effect  `json:"active_effects"`
}

// Character 角色/角色
type Character struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"` // pc, npc, enemy
	HP              int              `json:"hp"`
	MaxHP           int              `json:"max_hp"`
	AC              int              `json:"ac"`
	Initiative      int              `json:"initiative"`
	Speed           int              `json:"speed"`
	Stats           CharacterStats   `json:"stats"`
	Saves           []string         `json:"saves"`
	Skills          []string         `json:"skills"`
	Actions         []Action         `json:"actions"`
	Reactions       []Reaction       `json:"reactions"`
	Conditions      []string         `json:"conditions"`
	Effects         []Effect         `json:"effects"`
	DeathSaves      *DeathSaves      `json:"death_saves,omitempty"`
	Level           int              `json:"level,omitempty"`
	Classes         []ClassInfo      `json:"classes,omitempty"`
}

// CharacterStats 角色属性
type CharacterStats struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// ClassInfo 职业信息
type ClassInfo struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// Action 动作
type Action struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // attack, spell, ability, item
	Description string `json:"description"`
}

// Reaction 反应
type Reaction struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Effect 效果
type Effect struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // condition, buff, debuff
	Duration   int       `json:"duration"` // rounds or seconds
	ExpiresAt  time.Time `json:"expires_at"`
	Source     string    `json:"source"`
	Modifier   int       `json:"modifier,omitempty"`
}

// DeathSaves 死亡豁免
type DeathSaves struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

// Combat 战斗状态
type Combat struct {
	ID           string         `json:"id"`
	Active       bool           `json:"active"`
	Round        int            `json:"round"`
	Turn         int            `json:"turn"`
	Initiative   []Initiative   `json:"initiative"`
	CurrentActor string         `json:"current_actor"`
	StartedAt    time.Time      `json:"started_at"`
}

// Initiative 先攻
type Initiative struct {
	CharacterID string `json:"character_id"`
	Value       int    `json:"value"`
}

// Map 地图数据（与 Server 端 models.Map 对齐）
type Map struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Scale  int     `json:"scale"`
	Tokens []Token `json:"tokens"`
	Walls  []Wall  `json:"walls"`
	Lights []Light `json:"lights"`
}

// Token 标记（角色/物体）
type Token struct {
	ID          string  `json:"id"`
	CharacterID string  `json:"character_id"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Visible     bool    `json:"visible"`
}

// Wall 墙壁/障碍物
type Wall struct {
	ID     string `json:"id"`
	X1     int    `json:"x1"`
	Y1     int    `json:"y1"`
	X2     int    `json:"x2"`
	Y2     int    `json:"y2"`
	Type   string `json:"type"` // wall, door, window
	Open   bool   `json:"open"`
}

// Light 光源
type Light struct {
	ID       string `json:"id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Radius   int    `json:"radius"`
	Bright   int    `json:"bright"`
	Dim      int    `json:"dim"`
	Color    string `json:"color"`
	Animated bool   `json:"animated"`
}
