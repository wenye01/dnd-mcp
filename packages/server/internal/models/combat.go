package models

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// CombatStatus 战斗状态
type CombatStatus string

const (
	// CombatStatusActive 进行中
	CombatStatusActive CombatStatus = "active"
	// CombatStatusFinished 已结束
	CombatStatusFinished CombatStatus = "finished"
)

// Combat 战斗实体
// 规则参考: PHB 第9章 Combat
type Combat struct {
	ID           string           `json:"id"`            // UUID
	CampaignID   string           `json:"campaign_id"`   // 所属战役ID
	Status       CombatStatus     `json:"status"`        // 战斗状态
	Round        int              `json:"round"`         // 当前回合数
	TurnIndex    int              `json:"turn_index"`    // 当前行动者索引
	Participants []Participant    `json:"participants"`  // 参战者列表
	MapID        string           `json:"map_id"`        // 战斗地图ID（可选）
	Log          []CombatLogEntry `json:"log"`           // 战斗日志
	StartedAt    time.Time        `json:"started_at"`
	EndedAt      *time.Time       `json:"ended_at,omitempty"`
}

// NewCombat 创建新战斗
func NewCombat(campaignID string, participantIDs []string) *Combat {
	now := time.Now()
	participants := make([]Participant, 0, len(participantIDs))
	for _, id := range participantIDs {
		participants = append(participants, Participant{
			CharacterID: id,
			Initiative:  0,
			HasActed:    false,
			Conditions:  make([]Condition, 0),
		})
	}
	return &Combat{
		ID:           uuid.New().String(),
		CampaignID:   campaignID,
		Status:       CombatStatusActive,
		Round:        1,
		TurnIndex:    0,
		Participants: participants,
		Log:          make([]CombatLogEntry, 0),
		StartedAt:    now,
	}
}

// Validate 验证战斗数据
func (c *Combat) Validate() error {
	if c.CampaignID == "" {
		return NewValidationError("campaign_id", "cannot be empty")
	}
	if c.Round < 1 {
		return NewValidationError("round", "must be at least 1")
	}
	if c.TurnIndex < 0 {
		return NewValidationError("turn_index", "cannot be negative")
	}
	if len(c.Participants) == 0 {
		return NewValidationError("participants", "cannot be empty")
	}
	if c.TurnIndex >= len(c.Participants) {
		return NewValidationError("turn_index", "cannot exceed participants count")
	}
	// 验证每个参战者
	for i, p := range c.Participants {
		if err := p.Validate(); err != nil {
			return NewValidationError("participants", "invalid participant at index "+string(rune('0'+i))+": "+err.Error())
		}
	}
	return nil
}

// AddLogEntry 添加战斗日志
func (c *Combat) AddLogEntry(actorID, action, targetID, result string) {
	entry := CombatLogEntry{
		Round:     c.Round,
		ActorID:   actorID,
		Action:    action,
		TargetID:  targetID,
		Result:    result,
		Timestamp: time.Now(),
	}
	c.Log = append(c.Log, entry)
}

// GetCurrentParticipant 获取当前行动的参战者
func (c *Combat) GetCurrentParticipant() *Participant {
	if len(c.Participants) == 0 || c.TurnIndex >= len(c.Participants) {
		return nil
	}
	return &c.Participants[c.TurnIndex]
}

// AdvanceTurn 推进回合，返回是否进入新回合
func (c *Combat) AdvanceTurn() bool {
	if len(c.Participants) == 0 {
		return false
	}

	// 标记当前参战者已行动
	if c.TurnIndex < len(c.Participants) {
		c.Participants[c.TurnIndex].HasActed = true
	}

	// 推进到下一个参战者
	c.TurnIndex++

	// 检查是否需要进入新回合
	if c.TurnIndex >= len(c.Participants) {
		c.Round++
		c.TurnIndex = 0
		// 重置所有参战者的行动状态
		for i := range c.Participants {
			c.Participants[i].HasActed = false
		}
		return true
	}

	return false
}

// SortParticipantsByInitiative 按先攻值排序参战者（从高到低）
// 规则参考: PHB 第9章 Initiative
func (c *Combat) SortParticipantsByInitiative() {
	sort.Slice(c.Participants, func(i, j int) bool {
		// 先攻值高的排前面
		return c.Participants[i].Initiative > c.Participants[j].Initiative
	})
}

// GetParticipantByCharacterID 获取指定参战者
func (c *Combat) GetParticipantByCharacterID(characterID string) *Participant {
	for i := range c.Participants {
		if c.Participants[i].CharacterID == characterID {
			return &c.Participants[i]
		}
	}
	return nil
}

// End 结束战斗
func (c *Combat) End() {
	c.Status = CombatStatusFinished
	now := time.Now()
	c.EndedAt = &now
}

// IsFinished 检查战斗是否已结束
func (c *Combat) IsFinished() bool {
	return c.Status == CombatStatusFinished
}

// IsActive 检查战斗是否进行中
func (c *Combat) IsActive() bool {
	return c.Status == CombatStatusActive
}

// Participant 参战者
type Participant struct {
	CharacterID string       `json:"character_id"`  // 角色ID
	Initiative  int          `json:"initiative"`    // 先攻值
	HasActed    bool         `json:"has_acted"`     // 本回合是否已行动
	Position    *Position    `json:"position"`      // 战斗地图位置
	TempHP      int          `json:"temp_hp"`       // 临时HP（战斗中）
	Conditions  []Condition  `json:"conditions"`    // 战斗中的临时状态
}

// Validate 验证参战者数据
func (p *Participant) Validate() error {
	if p.CharacterID == "" {
		return NewValidationError("character_id", "cannot be empty")
	}
	if p.Initiative < -5 || p.Initiative > 30 {
		// 正常范围检查，D&D 5e 先攻值通常在 -5 到 30 之间
		// 但不强制报错，仅作为合理性检查
	}
	if p.TempHP < 0 {
		return NewValidationError("temp_hp", "cannot be negative")
	}
	if p.Position != nil {
		if err := p.Position.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SetPosition 设置位置
func (p *Participant) SetPosition(x, y int) {
	p.Position = &Position{X: x, Y: y}
}

// AddCondition 添加战斗中的临时状态
func (p *Participant) AddCondition(conditionType string, duration int, source string) {
	// 检查是否已有该状态
	for i, cond := range p.Conditions {
		if cond.Type == conditionType {
			// 如果新持续时间更长或为永久，更新持续时间
			if duration == -1 || (cond.Duration != -1 && duration > cond.Duration) {
				p.Conditions[i].Duration = duration
			}
			return
		}
	}

	// 添加新状态
	p.Conditions = append(p.Conditions, Condition{
		Type:     conditionType,
		Duration: duration,
		Source:   source,
	})
}

// RemoveCondition 移除战斗中的临时状态
func (p *Participant) RemoveCondition(conditionType string) bool {
	for i, cond := range p.Conditions {
		if cond.Type == conditionType {
			p.Conditions = append(p.Conditions[:i], p.Conditions[i+1:]...)
			return true
		}
	}
	return false
}

// HasCondition 检查是否有特定状态
func (p *Participant) HasCondition(conditionType string) bool {
	for _, cond := range p.Conditions {
		if cond.Type == conditionType {
			return true
		}
	}
	return false
}

// TickConditions 推进所有状态的持续时间
func (p *Participant) TickConditions() []string {
	expired := make([]string, 0)
	newConditions := make([]Condition, 0)

	for i := range p.Conditions {
		if p.Conditions[i].Tick() {
			expired = append(expired, p.Conditions[i].Type)
		} else {
			newConditions = append(newConditions, p.Conditions[i])
		}
	}

	if len(expired) > 0 {
		p.Conditions = newConditions
	}

	return expired
}

// CombatLogEntry 战斗日志条目
type CombatLogEntry struct {
	Round     int       `json:"round"`       // 回合数
	ActorID   string    `json:"actor_id"`    // 行动者ID
	Action    string    `json:"action"`      // 动作类型（attack, spell, move, etc.）
	TargetID  string    `json:"target_id"`   // 目标ID（可选）
	Result    string    `json:"result"`      // 结果描述
	Timestamp time.Time `json:"timestamp"`
}
