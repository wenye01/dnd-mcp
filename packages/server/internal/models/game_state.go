package models

import (
	"time"
)

// MapType 地图类型
type MapType string

const (
	// MapTypeWorld 大地图
	MapTypeWorld MapType = "world"
	// MapTypeBattle 战斗地图
	MapTypeBattle MapType = "battle"
)

// TimePhase 时段
type TimePhase string

const (
	// TimePhaseDawn 黎明
	TimePhaseDawn TimePhase = "dawn"
	// TimePhaseMorning 上午
	TimePhaseMorning TimePhase = "morning"
	// TimePhaseNoon 正午
	TimePhaseNoon TimePhase = "noon"
	// TimePhaseAfternoon 下午
	TimePhaseAfternoon TimePhase = "afternoon"
	// TimePhaseDusk 黄昏
	TimePhaseDusk TimePhase = "dusk"
	// TimePhaseNight 夜晚
	TimePhaseNight TimePhase = "night"
)

// GameTime 游戏时间
type GameTime struct {
	Year   int       `json:"year"`   // 年
	Month  int       `json:"month"`  // 月
	Day    int       `json:"day"`    // 日
	Hour   int       `json:"hour"`   // 时
	Minute int       `json:"minute"` // 分
	Phase  TimePhase `json:"phase"`  // 时段
}

// NewGameTime 创建默认游戏时间
func NewGameTime() *GameTime {
	return &GameTime{
		Year:   1,
		Month:  1,
		Day:    1,
		Hour:   8,
		Minute: 0,
		Phase:  TimePhaseMorning,
	}
}

// Validate 验证游戏时间
func (t *GameTime) Validate() error {
	if t.Year < 0 {
		return NewValidationError("year", "cannot be negative")
	}
	if t.Month < 1 || t.Month > 12 {
		return NewValidationError("month", "must be between 1 and 12")
	}
	if t.Day < 1 || t.Day > 31 {
		return NewValidationError("day", "must be between 1 and 31")
	}
	if t.Hour < 0 || t.Hour > 23 {
		return NewValidationError("hour", "must be between 0 and 23")
	}
	if t.Minute < 0 || t.Minute > 59 {
		return NewValidationError("minute", "must be between 0 and 59")
	}
	return nil
}

// AddMinutes 添加分钟数
func (t *GameTime) AddMinutes(minutes int) {
	t.Minute += minutes
	for t.Minute >= 60 {
		t.Minute -= 60
		t.Hour++
	}
	t.normalize()
}

// AddHours 添加小时数
func (t *GameTime) AddHours(hours int) {
	t.Hour += hours
	t.normalize()
}

// AddDays 添加天数
func (t *GameTime) AddDays(days int) {
	t.Day += days
	t.normalize()
}

// normalize 标准化时间
func (t *GameTime) normalize() {
	// 标准化分钟
	for t.Minute >= 60 {
		t.Minute -= 60
		t.Hour++
	}

	// 标准化小时
	for t.Hour >= 24 {
		t.Hour -= 24
		t.Day++
	}

	// 标准化天（简化处理，假设每月30天）
	for t.Day > 30 {
		t.Day -= 30
		t.Month++
	}

	// 标准化月
	for t.Month > 12 {
		t.Month -= 12
		t.Year++
	}

	// 更新时段
	t.updatePhase()
}

// updatePhase 根据小时更新时段
func (t *GameTime) updatePhase() {
	switch {
	case t.Hour >= 5 && t.Hour < 6:
		t.Phase = TimePhaseDawn
	case t.Hour >= 6 && t.Hour < 12:
		t.Phase = TimePhaseMorning
	case t.Hour >= 12 && t.Hour < 14:
		t.Phase = TimePhaseNoon
	case t.Hour >= 14 && t.Hour < 18:
		t.Phase = TimePhaseAfternoon
	case t.Hour >= 18 && t.Hour < 20:
		t.Phase = TimePhaseDusk
	default:
		t.Phase = TimePhaseNight
	}
}

// Position 位置
type Position struct {
	X int `json:"x"` // 格子X坐标
	Y int `json:"y"` // 格子Y坐标
}

// Validate 验证位置
func (p *Position) Validate() error {
	if p.X < 0 {
		return NewValidationError("x", "cannot be negative")
	}
	if p.Y < 0 {
		return NewValidationError("y", "cannot be negative")
	}
	return nil
}

// GameState 游戏状态
type GameState struct {
	ID             string        `json:"id"`               // 与CampaignID相同
	CampaignID     string        `json:"campaign_id"`      // 所属战役ID
	GameTime       *GameTime     `json:"game_time"`        // 游戏时间
	PartyPosition  *Position     `json:"party_position"`   // 队伍在大地图的位置（Grid 模式）
	CurrentMapID   string        `json:"current_map_id"`   // 当前所在地图ID
	CurrentMapType MapType       `json:"current_map_type"` // 当前地图类型
	Weather        string        `json:"weather"`          // 天气
	ActiveCombatID string        `json:"active_combat_id"` // 当前战斗ID（如果有）
	PlayerMarker   *PlayerMarker `json:"player_marker,omitempty"` // 玩家标记位置（Image 模式）
	UpdatedAt      time.Time     `json:"updated_at"`
}

// NewGameState 创建新游戏状态
func NewGameState(campaignID string) *GameState {
	now := time.Now()
	return &GameState{
		ID:             campaignID, // ID 与 CampaignID 相同
		CampaignID:     campaignID,
		GameTime:       NewGameTime(),
		PartyPosition:  &Position{X: 0, Y: 0},
		CurrentMapType: MapTypeWorld,
		Weather:        "clear",
		UpdatedAt:      now,
	}
}

// Validate 验证游戏状态
func (s *GameState) Validate() error {
	if s.CampaignID == "" {
		return NewValidationError("campaign_id", "cannot be empty")
	}
	if s.GameTime != nil {
		if err := s.GameTime.Validate(); err != nil {
			return err
		}
	}
	if s.PartyPosition != nil {
		if err := s.PartyPosition.Validate(); err != nil {
			return err
		}
	}
	if s.PlayerMarker != nil {
		if err := s.PlayerMarker.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// IsInCombat 检查是否在战斗中
func (s *GameState) IsInCombat() bool {
	return s.ActiveCombatID != ""
}

// IsInBattleMap 检查是否在战斗地图中
func (s *GameState) IsInBattleMap() bool {
	return s.CurrentMapType == MapTypeBattle
}

// SetCombat 设置当前战斗
func (s *GameState) SetCombat(combatID string) {
	s.ActiveCombatID = combatID
	s.UpdatedAt = time.Now()
}

// ClearCombat 清除当前战斗
func (s *GameState) ClearCombat() {
	s.ActiveCombatID = ""
	s.UpdatedAt = time.Now()
}

// SetCurrentMap 设置当前地图
func (s *GameState) SetCurrentMap(mapID string, mapType MapType) {
	s.CurrentMapID = mapID
	s.CurrentMapType = mapType
	s.UpdatedAt = time.Now()
}

// SetPartyPosition 设置队伍位置
func (s *GameState) SetPartyPosition(pos *Position) error {
	if pos != nil {
		if err := pos.Validate(); err != nil {
			return err
		}
	}
	s.PartyPosition = pos
	s.UpdatedAt = time.Now()
	return nil
}

// SetWeather 设置天气
func (s *GameState) SetWeather(weather string) {
	s.Weather = weather
	s.UpdatedAt = time.Now()
}

// AdvanceTime 推进游戏时间（小时）
func (s *GameState) AdvanceTime(hours int) {
	if s.GameTime != nil {
		s.GameTime.AddHours(hours)
	}
	s.UpdatedAt = time.Now()
}

// AdvanceTimeMinutes 推进游戏时间（分钟）
func (s *GameState) AdvanceTimeMinutes(minutes int) {
	if s.GameTime != nil {
		s.GameTime.AddMinutes(minutes)
	}
	s.UpdatedAt = time.Now()
}

// PlayerMarker 玩家标记（用于 Image 模式的大地图）
type PlayerMarker struct {
	PositionX    float64 `json:"position_x"`              // X 坐标（0-1 归一化）
	PositionY    float64 `json:"position_y"`              // Y 坐标（0-1 归一化）
	CurrentScene string  `json:"current_scene,omitempty"` // 当前场景描述
	UpdatedAt    string  `json:"updated_at"`              // 更新时间（ISO 8601）
}

// NewPlayerMarker 创建新的玩家标记
func NewPlayerMarker(posX, posY float64) *PlayerMarker {
	return &PlayerMarker{
		PositionX: posX,
		PositionY: posY,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
}

// Validate 验证玩家标记
func (p *PlayerMarker) Validate() error {
	if p.PositionX < 0 || p.PositionX > 1 {
		return NewValidationError("player_marker.position_x", "must be between 0 and 1")
	}
	if p.PositionY < 0 || p.PositionY > 1 {
		return NewValidationError("player_marker.position_y", "must be between 0 and 1")
	}
	return nil
}

// SetPosition 设置位置
func (p *PlayerMarker) SetPosition(posX, posY float64) {
	p.PositionX = posX
	p.PositionY = posY
	p.UpdatedAt = time.Now().Format(time.RFC3339)
}

// SetScene 设置当前场景描述
func (p *PlayerMarker) SetScene(scene string) {
	p.CurrentScene = scene
	p.UpdatedAt = time.Now().Format(time.RFC3339)
}
