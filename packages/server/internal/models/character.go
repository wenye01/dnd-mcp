package models

import (
	"time"
)

// NPCType NPC类型
type NPCType string

const (
	// NPCTypeScripted 剧本固定NPC
	NPCTypeScripted NPCType = "scripted"
	// NPCTypeGenerated LLM即兴创建NPC
	NPCTypeGenerated NPCType = "generated"
)

// Abilities 六大属性
type Abilities struct {
	Strength     int `json:"strength"`     // 力量
	Dexterity    int `json:"dexterity"`    // 敏捷
	Constitution int `json:"constitution"` // 体质
	Intelligence int `json:"intelligence"` // 智力
	Wisdom       int `json:"wisdom"`       // 感知
	Charisma     int `json:"charisma"`     // 魅力
}

// NewAbilities 创建默认属性值（标准数组：15, 14, 13, 12, 10, 8）
func NewAbilities() *Abilities {
	return &Abilities{
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       10,
		Charisma:     8,
	}
}

// Validate 验证属性值
func (a *Abilities) Validate() error {
	if a.Strength < 1 || a.Strength > 30 {
		return NewValidationError("strength", "must be between 1 and 30")
	}
	if a.Dexterity < 1 || a.Dexterity > 30 {
		return NewValidationError("dexterity", "must be between 1 and 30")
	}
	if a.Constitution < 1 || a.Constitution > 30 {
		return NewValidationError("constitution", "must be between 1 and 30")
	}
	if a.Intelligence < 1 || a.Intelligence > 30 {
		return NewValidationError("intelligence", "must be between 1 and 30")
	}
	if a.Wisdom < 1 || a.Wisdom > 30 {
		return NewValidationError("wisdom", "must be between 1 and 30")
	}
	if a.Charisma < 1 || a.Charisma > 30 {
		return NewValidationError("charisma", "must be between 1 and 30")
	}
	return nil
}

// HP 生命值
type HP struct {
	Current int `json:"current"` // 当前HP
	Max     int `json:"max"`     // 最大HP
	Temp    int `json:"temp"`    // 临时HP
}

// NewHP 创建生命值
func NewHP(max int) *HP {
	return &HP{
		Current: max,
		Max:     max,
		Temp:    0,
	}
}

// Validate 验证HP
func (h *HP) Validate() error {
	if h.Max < 1 {
		return NewValidationError("hp.max", "must be at least 1")
	}
	if h.Current < 0 {
		return NewValidationError("hp.current", "cannot be negative")
	}
	if h.Temp < 0 {
		return NewValidationError("hp.temp", "cannot be negative")
	}
	return nil
}

// IsDead 检查是否死亡
func (h *HP) IsDead() bool {
	return h.Current <= 0
}

// IsUnconscious 检查是否昏迷（HP为0但未死亡）
func (h *HP) IsUnconscious() bool {
	return h.Current == 0
}

// TakeDamage 受到伤害
func (h *HP) TakeDamage(damage int) int {
	// 先扣临时HP
	if h.Temp > 0 {
		if damage <= h.Temp {
			h.Temp -= damage
			return 0
		}
		damage -= h.Temp
		h.Temp = 0
	}

	// 再扣当前HP
	if damage > h.Current {
		overflow := damage - h.Current
		h.Current = 0
		return overflow
	}
	h.Current -= damage
	return 0
}

// Heal 恢复生命值
func (h *HP) Heal(amount int) int {
	healed := amount
	if h.Current+amount > h.Max {
		healed = h.Max - h.Current
		h.Current = h.Max
	} else {
		h.Current += amount
	}
	return healed
}

// AddTempHP 添加临时HP
func (h *HP) AddTempHP(amount int) {
	// 临时HP不叠加，取较大值
	if amount > h.Temp {
		h.Temp = amount
	}
}

// Equipment 装备
type Equipment struct {
	ID         string `json:"id"`          // 物品ID
	Name       string `json:"name"`        // 物品名称
	Slot       string `json:"slot"`        // 装备槽位（main_hand, off_hand, armor, etc.）
	Bonus      int    `json:"bonus"`       // 加值
	Damage     string `json:"damage"`      // 伤害骰（武器）
	DamageType string `json:"damage_type"` // 伤害类型
}

// Validate 验证装备
func (e *Equipment) Validate() error {
	if e.Name == "" {
		return NewValidationError("equipment.name", "cannot be empty")
	}
	return nil
}

// Item 物品
type Item struct {
	ID       string `json:"id"`       // 物品ID
	Name     string `json:"name"`     // 物品名称
	Quantity int    `json:"quantity"` // 数量
}

// Validate 验证物品
func (i *Item) Validate() error {
	if i.Name == "" {
		return NewValidationError("item.name", "cannot be empty")
	}
	if i.Quantity < 0 {
		return NewValidationError("item.quantity", "cannot be negative")
	}
	return nil
}

// Condition 状态效果
type Condition struct {
	Type     string `json:"type"`     // 状态类型（poisoned, paralyzed, etc.）
	Duration int    `json:"duration"` // 持续回合数，-1表示永久
	Source   string `json:"source"`   // 来源
}

// Validate 验证状态效果
func (c *Condition) Validate() error {
	if c.Type == "" {
		return NewValidationError("condition.type", "cannot be empty")
	}
	if c.Duration < -1 {
		return NewValidationError("condition.duration", "must be -1 (permanent) or >= 0")
	}
	return nil
}

// IsPermanent 检查是否永久状态
func (c *Condition) IsPermanent() bool {
	return c.Duration == -1
}

// Tick 减少持续时间
func (c *Condition) Tick() bool {
	if c.Duration > 0 {
		c.Duration--
		return c.Duration == 0 // 返回是否已过期
	}
	return false
}

// Character 角色实体（玩家角色和NPC共用）
type Character struct {
	ID         string         `json:"id"`          // UUID
	CampaignID string         `json:"campaign_id"` // 所属战役ID
	Name       string         `json:"name"`        // 角色名称
	IsNPC      bool           `json:"is_npc"`      // 是否为NPC
	NPCType    NPCType        `json:"npc_type"`    // NPC类型（仅NPC使用）
	PlayerID   string         `json:"player_id"`   // 玩家ID（玩家角色使用）

	// 基础属性
	Race       string `json:"race"`       // 种族
	Class      string `json:"class"`      // 职业
	Level      int    `json:"level"`      // 等级
	Background string `json:"background"` // 背景
	Alignment  string `json:"alignment"`  // 阵营

	// 属性值
	Abilities *Abilities `json:"abilities"` // 六大属性

	// 战斗属性
	HP         *HP            `json:"hp"`         // 生命值
	AC         int            `json:"ac"`         // 护甲等级
	Speed      int            `json:"speed"`      // 移动速度（英尺）
	Initiative int            `json:"initiative"` // 先攻加值

	// 技能和特长
	Skills map[string]int `json:"skills"` // 技能加值
	Saves  map[string]int `json:"saves"`  // 豁免加值

	// 装备和物品
	Equipment []Equipment `json:"equipment"` // 装备列表
	Inventory []Item      `json:"inventory"` // 背包物品

	// 状态
	Conditions []Condition `json:"conditions"` // 状态效果

	// 元数据
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // 软删除时间
}

// NewCharacter 创建新角色
func NewCharacter(campaignID, name string, isNPC bool) *Character {
	now := time.Now()
	return &Character{
		CampaignID: campaignID,
		Name:       name,
		IsNPC:      isNPC,
		Level:      1,
		Abilities:  NewAbilities(),
		AC:         10,
		Speed:      30,
		Skills:     make(map[string]int),
		Saves:      make(map[string]int),
		Equipment:  make([]Equipment, 0),
		Inventory:  make([]Item, 0),
		Conditions: make([]Condition, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate 验证角色数据
func (c *Character) Validate() error {
	if c.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}
	if c.CampaignID == "" {
		return NewValidationError("campaign_id", "cannot be empty")
	}
	if !c.IsNPC && c.PlayerID == "" {
		return NewValidationError("player_id", "cannot be empty for player characters")
	}
	if c.Level < 1 || c.Level > 20 {
		return NewValidationError("level", "must be between 1 and 20")
	}
	if c.Abilities != nil {
		if err := c.Abilities.Validate(); err != nil {
			return err
		}
	}
	if c.HP != nil {
		if err := c.HP.Validate(); err != nil {
			return err
		}
	}
	if c.AC < 0 {
		return NewValidationError("ac", "cannot be negative")
	}
	if c.Speed < 0 {
		return NewValidationError("speed", "cannot be negative")
	}
	return nil
}

// IsPlayerCharacter 检查是否为玩家角色
func (c *Character) IsPlayerCharacter() bool {
	return !c.IsNPC
}

// IsScriptedNPC 检查是否为剧本NPC
func (c *Character) IsScriptedNPC() bool {
	return c.IsNPC && c.NPCType == NPCTypeScripted
}

// IsGeneratedNPC 检查是否为即兴NPC
func (c *Character) IsGeneratedNPC() bool {
	return c.IsNPC && c.NPCType == NPCTypeGenerated
}

// IsDead 检查是否死亡
func (c *Character) IsDead() bool {
	return c.HP != nil && c.HP.IsDead()
}

// IsUnconscious 检查是否昏迷
func (c *Character) IsUnconscious() bool {
	return c.HP != nil && c.HP.IsUnconscious()
}

// HasCondition 检查是否有特定状态
func (c *Character) HasCondition(conditionType string) bool {
	for _, cond := range c.Conditions {
		if cond.Type == conditionType {
			return true
		}
	}
	return false
}

// AddCondition 添加状态效果
func (c *Character) AddCondition(conditionType string, duration int, source string) {
	// 检查是否已有该状态
	for i, cond := range c.Conditions {
		if cond.Type == conditionType {
			// 如果新持续时间更长或为永久，更新持续时间
			if duration == -1 || (cond.Duration != -1 && duration > cond.Duration) {
				c.Conditions[i].Duration = duration
			}
			c.UpdatedAt = time.Now()
			return
		}
	}

	// 添加新状态
	c.Conditions = append(c.Conditions, Condition{
		Type:     conditionType,
		Duration: duration,
		Source:   source,
	})
	c.UpdatedAt = time.Now()
}

// RemoveCondition 移除状态效果
func (c *Character) RemoveCondition(conditionType string) bool {
	for i, cond := range c.Conditions {
		if cond.Type == conditionType {
			c.Conditions = append(c.Conditions[:i], c.Conditions[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// TickConditions 推进所有状态的持续时间
func (c *Character) TickConditions() []string {
	expired := make([]string, 0)
	newConditions := make([]Condition, 0)

	for i, cond := range c.Conditions {
		if cond.Tick() {
			expired = append(expired, cond.Type)
		} else {
			newConditions = append(newConditions, c.Conditions[i])
		}
	}

	if len(expired) > 0 {
		c.Conditions = newConditions
		c.UpdatedAt = time.Now()
	}

	return expired
}

// TakeDamage 受到伤害
func (c *Character) TakeDamage(damage int) int {
	if c.HP == nil {
		return damage
	}
	overflow := c.HP.TakeDamage(damage)
	c.UpdatedAt = time.Now()
	return overflow
}

// Heal 恢复生命值
func (c *Character) Heal(amount int) int {
	if c.HP == nil {
		return 0
	}
	healed := c.HP.Heal(amount)
	c.UpdatedAt = time.Now()
	return healed
}

// AddItem 添加物品到背包
func (c *Character) AddItem(id, name string, quantity int) {
	// 检查是否已有该物品
	for i, item := range c.Inventory {
		if item.ID == id {
			c.Inventory[i].Quantity += quantity
			c.UpdatedAt = time.Now()
			return
		}
	}

	// 添加新物品
	c.Inventory = append(c.Inventory, Item{
		ID:       id,
		Name:     name,
		Quantity: quantity,
	})
	c.UpdatedAt = time.Now()
}

// RemoveItem 从背包移除物品
func (c *Character) RemoveItem(id string, quantity int) bool {
	for i, item := range c.Inventory {
		if item.ID == id {
			if item.Quantity <= quantity {
				// 全部移除
				c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			} else {
				// 部分移除
				c.Inventory[i].Quantity -= quantity
			}
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// Equip 装备物品
func (c *Character) Equip(equipment Equipment) error {
	if err := equipment.Validate(); err != nil {
		return err
	}

	// 检查槽位是否已有装备
	for i, eq := range c.Equipment {
		if eq.Slot == equipment.Slot {
			c.Equipment[i] = equipment
			c.UpdatedAt = time.Now()
			return nil
		}
	}

	// 添加新装备
	c.Equipment = append(c.Equipment, equipment)
	c.UpdatedAt = time.Now()
	return nil
}

// Unequip 卸下装备
func (c *Character) Unequip(slot string) *Equipment {
	for i, eq := range c.Equipment {
		if eq.Slot == slot {
			equipment := c.Equipment[i]
			c.Equipment = append(c.Equipment[:i], c.Equipment[i+1:]...)
			c.UpdatedAt = time.Now()
			return &equipment
		}
	}
	return nil
}

// GetEquipmentBySlot 获取指定槽位的装备
func (c *Character) GetEquipmentBySlot(slot string) *Equipment {
	for i, eq := range c.Equipment {
		if eq.Slot == slot {
			return &c.Equipment[i]
		}
	}
	return nil
}

// SetAbilities 设置属性值
func (c *Character) SetAbilities(abilities *Abilities) error {
	if abilities != nil {
		if err := abilities.Validate(); err != nil {
			return err
		}
	}
	c.Abilities = abilities
	c.UpdatedAt = time.Now()
	return nil
}

// SetHP 设置生命值
func (c *Character) SetHP(hp *HP) error {
	if hp != nil {
		if err := hp.Validate(); err != nil {
			return err
		}
	}
	c.HP = hp
	c.UpdatedAt = time.Now()
	return nil
}

// SetSkillBonus 设置技能加值
func (c *Character) SetSkillBonus(skill string, bonus int) {
	if c.Skills == nil {
		c.Skills = make(map[string]int)
	}
	c.Skills[skill] = bonus
	c.UpdatedAt = time.Now()
}

// SetSaveBonus 设置豁免加值
func (c *Character) SetSaveBonus(save string, bonus int) {
	if c.Saves == nil {
		c.Saves = make(map[string]int)
	}
	c.Saves[save] = bonus
	c.UpdatedAt = time.Now()
}

// GetSkillBonus 获取技能加值
func (c *Character) GetSkillBonus(skill string) int {
	if c.Skills == nil {
		return 0
	}
	return c.Skills[skill]
}

// GetSaveBonus 获取豁免加值
func (c *Character) GetSaveBonus(save string) int {
	if c.Saves == nil {
		return 0
	}
	return c.Saves[save]
}

// UpdateBasicInfo 更新基础信息
func (c *Character) UpdateBasicInfo(race, class, background, alignment string) {
	if race != "" {
		c.Race = race
	}
	if class != "" {
		c.Class = class
	}
	if background != "" {
		c.Background = background
	}
	if alignment != "" {
		c.Alignment = alignment
	}
	c.UpdatedAt = time.Now()
}

// SetLevel 设置等级
func (c *Character) SetLevel(level int) error {
	if level < 1 || level > 20 {
		return NewValidationError("level", "must be between 1 and 20")
	}
	c.Level = level
	c.UpdatedAt = time.Now()
	return nil
}
