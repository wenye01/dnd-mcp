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

// IsDeadByOverflow 检查是否因溢出伤害死亡
// 规则: 当溢出伤害 >= MaxHP 时立即死亡
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Instant Death
func (h *HP) IsDeadByOverflow() bool {
	if h.Current < 0 {
		overflow := -h.Current
		return overflow >= h.Max
	}
	return false
}

// IsAtZero 检查 HP 是否为 0
func (h *HP) IsAtZero() bool {
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
	Speed      int            `json:"speed"`      // 移动速度（英尺）- 向后兼容
	Initiative int            `json:"initiative"` // 先攻加值

	// 技能和特长
	Skills map[string]int `json:"skills"` // 技能加值 - 向后兼容
	Saves  map[string]int `json:"saves"`  // 豁免加值 - 向后兼容

	// 装备和物品
	Equipment []Equipment `json:"equipment"` // 装备列表 - 向后兼容
	Inventory []Item      `json:"inventory"` // 背包物品 - 向后兼容

	// 状态
	Conditions []Condition `json:"conditions"` // 状态效果

	// ============ 扩展字段（T3-6）============

	// 图片
	Image string `json:"image,omitempty"` // 角色图片（Base64或URL）

	// 经验值
	Experience int `json:"experience,omitempty"` // 经验值

	// 熟练加值（通常根据等级计算，但可覆盖）
	Proficiency int `json:"proficiency,omitempty"` // 熟练加值

	// 移动速度（结构化，支持多种移动类型）
	SpeedDetail *Speed `json:"speed_detail,omitempty"` // 详细移动速度

	// 死亡豁免
	DeathSaves *DeathSaves `json:"death_saves,omitempty"` // 死亡豁免

	// 技能（结构化，含熟练/专精）
	SkillsDetail map[string]*Skill `json:"skills_detail,omitempty"` // 详细技能

	// 豁免（结构化，含熟练）
	SavesDetail map[string]*Save `json:"saves_detail,omitempty"` // 详细豁免

	// 货币
	Currency *Currency `json:"currency,omitempty"` // 货币

	// 装备槽位（结构化）
	EquipmentSlots *EquipmentSlots `json:"equipment_slots,omitempty"` // 装备槽位

	// 背包物品（结构化）
	InventoryItems []*InventoryItem `json:"inventory_items,omitempty"` // 详细背包物品

	// 法术书
	Spellbook *Spellbook `json:"spellbook,omitempty"` // 法术书

	// 专长/特性
	Features []*Feature `json:"features,omitempty"` // 专长/特性列表

	// 传记
	Biography *Biography `json:"biography,omitempty"` // 传记

	// 特性/抗性/语言
	Traits *Traits `json:"traits,omitempty"` // 特性/抗性/语言

	// 导入元数据
	ImportMeta *ImportMeta `json:"import_meta,omitempty"` // 导入元数据

	// ============ 元数据 ============

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
// 规则: 死亡豁免失败 3 次或溢出伤害 >= MaxHP
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points
func (c *Character) IsDead() bool {
	if c.HP == nil {
		return false
	}

	// 检查溢出伤害导致立即死亡
	if c.HP.IsDeadByOverflow() {
		return true
	}

	// 检查死亡豁免失败 3 次
	if c.DeathSaves != nil && c.DeathSaves.Failures >= 3 {
		return true
	}

	return false
}

// IsUnconscious 检查是否昏迷
// 规则: HP = 0 且未死亡且未稳定
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points
func (c *Character) IsUnconscious() bool {
	if c.HP == nil {
		return false
	}

	// HP 不为 0 则不昏迷
	if !c.HP.IsAtZero() {
		return false
	}

	// 已死亡则不昏迷
	if c.IsDead() {
		return false
	}

	// 已稳定则不昏迷
	if c.IsStable() {
		return false
	}

	return true
}

// IsStable 检查是否处于稳定状态
// 规则: 死亡豁免成功 3 次
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points
func (c *Character) IsStable() bool {
	if c.DeathSaves == nil {
		return false
	}
	return c.DeathSaves.Successes >= 3
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

// ============ 扩展结构体定义 ============

// Speed 多类型移动速度
// 规则参考: PHB 第8章 Movement
type Speed struct {
	Walk    int  `json:"walk"`              // 步行速度
	Burrow  int  `json:"burrow,omitempty"`  // 挖掘速度
	Climb   int  `json:"climb,omitempty"`   // 攀爬速度
	Fly     int  `json:"fly,omitempty"`     // 飞行速度
	Hover   bool `json:"hover,omitempty"`   // 是否悬停
	Swim    int  `json:"swim,omitempty"`    // 游泳速度
	// 自定义速度
	Custom map[string]int `json:"custom,omitempty"` // 自定义速度类型
}

// NewSpeed 创建默认移动速度
func NewSpeed(walk int) *Speed {
	return &Speed{
		Walk:   walk,
		Custom: make(map[string]int),
	}
}

// Validate 验证移动速度
func (s *Speed) Validate() error {
	if s.Walk < 0 {
		return NewValidationError("speed.walk", "cannot be negative")
	}
	if s.Burrow < 0 {
		return NewValidationError("speed.burrow", "cannot be negative")
	}
	if s.Climb < 0 {
		return NewValidationError("speed.climb", "cannot be negative")
	}
	if s.Fly < 0 {
		return NewValidationError("speed.fly", "cannot be negative")
	}
	if s.Swim < 0 {
		return NewValidationError("speed.swim", "cannot be negative")
	}
	return nil
}

// GetSpeed 获取指定类型的速度
func (s *Speed) GetSpeed(speedType string) int {
	switch speedType {
	case "walk":
		return s.Walk
	case "burrow":
		return s.Burrow
	case "climb":
		return s.Climb
	case "fly":
		return s.Fly
	case "swim":
		return s.Swim
	default:
		if s.Custom != nil {
			return s.Custom[speedType]
		}
		return 0
	}
}

// SetSpeed 设置指定类型的速度
func (s *Speed) SetSpeed(speedType string, value int) {
	switch speedType {
	case "walk":
		s.Walk = value
	case "burrow":
		s.Burrow = value
	case "climb":
		s.Climb = value
	case "fly":
		s.Fly = value
	case "swim":
		s.Swim = value
	default:
		if s.Custom == nil {
			s.Custom = make(map[string]int)
		}
		s.Custom[speedType] = value
	}
}

// DeathSaves 死亡豁免
// 规则参考: PHB 第9章 Dropping to 0 Hit Points
type DeathSaves struct {
	Successes int `json:"successes"` // 成功次数（3次稳定）
	Failures  int `json:"failures"`  // 失败次数（3次死亡）
}

// NewDeathSaves 创建死亡豁免
func NewDeathSaves() *DeathSaves {
	return &DeathSaves{}
}

// Validate 验证死亡豁免
func (d *DeathSaves) Validate() error {
	if d.Successes < 0 || d.Successes > 3 {
		return NewValidationError("death_saves.successes", "must be between 0 and 3")
	}
	if d.Failures < 0 || d.Failures > 3 {
		return NewValidationError("death_saves.failures", "must be between 0 and 3")
	}
	return nil
}

// AddSuccess 添加成功
func (d *DeathSaves) AddSuccess() bool {
	if d.Successes < 3 {
		d.Successes++
		return true
	}
	return false
}

// AddFailure 添加失败
func (d *DeathSaves) AddFailure() bool {
	if d.Failures < 3 {
		d.Failures++
		return true
	}
	return false
}

// IsStable 是否稳定
func (d *DeathSaves) IsStable() bool {
	return d.Successes >= 3
}

// IsDead 是否死亡
func (d *DeathSaves) IsDead() bool {
	return d.Failures >= 3
}

// Reset 重置死亡豁免
func (d *DeathSaves) Reset() {
	d.Successes = 0
	d.Failures = 0
}

// Skill 技能（结构化）
// 规则参考: PHB 第7章 Using Ability Scores
type Skill struct {
	Ability   string `json:"ability"`              // 关联属性
	Bonus     int    `json:"bonus"`                // 总加值
	Proficient bool  `json:"proficient,omitempty"` // 是否熟练
	Expertise  bool  `json:"expertise,omitempty"`  // 是否专精（双倍熟练）
	// 半熟练（游荡者熟练）
	HalfProficient bool `json:"half_proficient,omitempty"`
	// 覆盖加值（忽略计算，直接使用）
	Override int `json:"override,omitempty"`
}

// Validate 验证技能
func (s *Skill) Validate() error {
	if s.Ability == "" {
		return NewValidationError("skill.ability", "cannot be empty")
	}
	return nil
}

// CalculateBonus 计算技能加值
func (s *Skill) CalculateBonus(abilityMod, proficiencyBonus int) int {
	// 如果有覆盖值，直接使用
	if s.Override != 0 {
		return s.Override
	}

	bonus := abilityMod

	if s.Expertise {
		bonus += proficiencyBonus * 2
	} else if s.Proficient {
		bonus += proficiencyBonus
	} else if s.HalfProficient {
		bonus += proficiencyBonus / 2
	}

	return bonus
}

// Save 豁免（结构化）
// 规则参考: PHB 第7章 Saving Throws
type Save struct {
	Bonus      int  `json:"bonus"`                // 总加值
	Proficient bool `json:"proficient,omitempty"` // 是否熟练
	// 覆盖加值
	Override int `json:"override,omitempty"`
}

// Validate 验证豁免
func (s *Save) Validate() error {
	return nil
}

// CalculateBonus 计算豁免加值
func (s *Save) CalculateBonus(abilityMod, proficiencyBonus int) int {
	// 如果有覆盖值，直接使用
	if s.Override != 0 {
		return s.Override
	}

	bonus := abilityMod

	if s.Proficient {
		bonus += proficiencyBonus
	}

	return bonus
}

// Biography 传记
// 规则参考: PHB 第4章 Personality and Background
type Biography struct {
	Age         string `json:"age,omitempty"`         // 年龄
	Height      string `json:"height,omitempty"`      // 身高
	Weight      string `json:"weight,omitempty"`      // 体重
	Eyes        string `json:"eyes,omitempty"`        // 眼睛
	Skin        string `json:"skin,omitempty"`        // 肤色
	Hair        string `json:"hair,omitempty"`        // 发色
	Appearance  string `json:"appearance,omitempty"`  // 外貌描述
	Backstory   string `json:"backstory,omitempty"`   // 背景故事
	PersonalityTraits string `json:"personality_traits,omitempty"` // 个性特征
	Ideals      string `json:"ideals,omitempty"`      // 理想
	Bonds       string `json:"bonds,omitempty"`       // 羁绊
	Flaws       string `json:"flaws,omitempty"`       // 缺点
	// 自定义字段
	Custom      map[string]string `json:"custom,omitempty"` // 自定义字段
}

// NewBiography 创建空传记
func NewBiography() *Biography {
	return &Biography{
		Custom: make(map[string]string),
	}
}

// Traits 特性/抗性/语言
type Traits struct {
	// 伤害抗性
	DamageResistances []string `json:"damage_resistances,omitempty"`
	// 伤害免疫
	DamageImmunities []string `json:"damage_immunities,omitempty"`
	// 状态免疫
	ConditionImmunities []string `json:"condition_immunities,omitempty"`
	// 伤害易伤
	DamageVulnerabilities []string `json:"damage_vulnerabilities,omitempty"`
	// 语言
	Languages []string `json:"languages,omitempty"`
	// 感官（黑暗视觉、微光视觉等）
	Senses map[string]int `json:"senses,omitempty"` // 类型 -> 距离（英尺）
	// 特殊特质
	SpecialTraits []string `json:"special_traits,omitempty"`
}

// NewTraits 创建空特性
func NewTraits() *Traits {
	return &Traits{
		DamageResistances:     make([]string, 0),
		DamageImmunities:      make([]string, 0),
		ConditionImmunities:   make([]string, 0),
		DamageVulnerabilities: make([]string, 0),
		Languages:             make([]string, 0),
		Senses:                make(map[string]int),
		SpecialTraits:         make([]string, 0),
	}
}

// AddResistance 添加伤害抗性
func (t *Traits) AddResistance(damageType string) {
	for _, r := range t.DamageResistances {
		if r == damageType {
			return
		}
	}
	t.DamageResistances = append(t.DamageResistances, damageType)
}

// RemoveResistance 移除伤害抗性
func (t *Traits) RemoveResistance(damageType string) bool {
	for i, r := range t.DamageResistances {
		if r == damageType {
			t.DamageResistances = append(t.DamageResistances[:i], t.DamageResistances[i+1:]...)
			return true
		}
	}
	return false
}

// HasResistance 检查是否有伤害抗性
func (t *Traits) HasResistance(damageType string) bool {
	for _, r := range t.DamageResistances {
		if r == damageType {
			return true
		}
	}
	return false
}

// AddImmunity 添加伤害免疫
func (t *Traits) AddImmunity(damageType string) {
	for _, r := range t.DamageImmunities {
		if r == damageType {
			return
		}
	}
	t.DamageImmunities = append(t.DamageImmunities, damageType)
}

// HasImmunity 检查是否有伤害免疫
func (t *Traits) HasImmunity(damageType string) bool {
	for _, r := range t.DamageImmunities {
		if r == damageType {
			return true
		}
	}
	return false
}

// AddConditionImmunity 添加状态免疫
func (t *Traits) AddConditionImmunity(condition string) {
	for _, r := range t.ConditionImmunities {
		if r == condition {
			return
		}
	}
	t.ConditionImmunities = append(t.ConditionImmunities, condition)
}

// HasConditionImmunity 检查是否有状态免疫
func (t *Traits) HasConditionImmunity(condition string) bool {
	for _, r := range t.ConditionImmunities {
		if r == condition {
			return true
		}
	}
	return false
}

// AddLanguage 添加语言
func (t *Traits) AddLanguage(language string) {
	for _, r := range t.Languages {
		if r == language {
			return
		}
	}
	t.Languages = append(t.Languages, language)
}

// HasLanguage 检查是否有语言
func (t *Traits) HasLanguage(language string) bool {
	for _, r := range t.Languages {
		if r == language {
			return true
		}
	}
	return false
}

// AddSense 添加感官
func (t *Traits) AddSense(senseType string, rangeFeet int) {
	if t.Senses == nil {
		t.Senses = make(map[string]int)
	}
	t.Senses[senseType] = rangeFeet
}

// HasSense 检查是否有感官
func (t *Traits) HasSense(senseType string) bool {
	if t.Senses == nil {
		return false
	}
	_, exists := t.Senses[senseType]
	return exists
}

// ============ 扩展方法 ============

// GetProficiencyBonus 获取熟练加值
// 规则参考: PHB 第7章 Proficiency Bonus
func (c *Character) GetProficiencyBonus() int {
	if c.Proficiency > 0 {
		return c.Proficiency
	}
	// 根据等级计算: 1-4: +2, 5-8: +3, 9-12: +4, 13-16: +5, 17-20: +6
	return 2 + ((c.Level - 1) / 4)
}

// SetProficiency 设置熟练加值
func (c *Character) SetProficiency(bonus int) {
	c.Proficiency = bonus
	c.UpdatedAt = time.Now()
}

// GetDetailedSpeed 获取详细移动速度
func (c *Character) GetDetailedSpeed() *Speed {
	if c.SpeedDetail == nil {
		c.SpeedDetail = NewSpeed(c.Speed)
	}
	return c.SpeedDetail
}

// GetDeathSaves 获取死亡豁免
func (c *Character) GetDeathSaves() *DeathSaves {
	if c.DeathSaves == nil {
		c.DeathSaves = NewDeathSaves()
	}
	return c.DeathSaves
}

// GetCurrency 获取货币
func (c *Character) GetCurrency() *Currency {
	if c.Currency == nil {
		c.Currency = NewCurrency()
	}
	return c.Currency
}

// GetEquipmentSlots 获取装备槽位
func (c *Character) GetEquipmentSlots() *EquipmentSlots {
	if c.EquipmentSlots == nil {
		c.EquipmentSlots = NewEquipmentSlots()
	}
	return c.EquipmentSlots
}

// GetSpellbook 获取法术书
func (c *Character) GetSpellbook() *Spellbook {
	if c.Spellbook == nil {
		c.Spellbook = NewSpellbook()
	}
	return c.Spellbook
}

// GetTraits 获取特性
func (c *Character) GetTraits() *Traits {
	if c.Traits == nil {
		c.Traits = NewTraits()
	}
	return c.Traits
}

// GetBiography 获取传记
func (c *Character) GetBiography() *Biography {
	if c.Biography == nil {
		c.Biography = NewBiography()
	}
	return c.Biography
}

// AddFeature 添加专长/特性
func (c *Character) AddFeature(feature *Feature) {
	if c.Features == nil {
		c.Features = make([]*Feature, 0)
	}
	c.Features = append(c.Features, feature)
	c.UpdatedAt = time.Now()
}

// RemoveFeature 移除专长/特性
func (c *Character) RemoveFeature(featureID string) bool {
	for i, f := range c.Features {
		if f.ID == featureID {
			c.Features = append(c.Features[:i], c.Features[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetFeature 获取专长/特性
func (c *Character) GetFeature(featureID string) *Feature {
	for _, f := range c.Features {
		if f.ID == featureID {
			return f
		}
	}
	return nil
}

// AddInventoryItem 添加详细背包物品
func (c *Character) AddInventoryItem(item *InventoryItem) {
	if c.InventoryItems == nil {
		c.InventoryItems = make([]*InventoryItem, 0)
	}
	c.InventoryItems = append(c.InventoryItems, item)
	c.UpdatedAt = time.Now()
}

// RemoveInventoryItem 移除详细背包物品
func (c *Character) RemoveInventoryItem(itemID string) bool {
	for i, item := range c.InventoryItems {
		if item.ID == itemID {
			c.InventoryItems = append(c.InventoryItems[:i], c.InventoryItems[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetInventoryItem 获取详细背包物品
func (c *Character) GetInventoryItem(itemID string) *InventoryItem {
	for _, item := range c.InventoryItems {
		if item.ID == itemID {
			return item
		}
	}
	return nil
}

// SetSkillDetail 设置详细技能
func (c *Character) SetSkillDetail(skillName string, skill *Skill) {
	if c.SkillsDetail == nil {
		c.SkillsDetail = make(map[string]*Skill)
	}
	c.SkillsDetail[skillName] = skill
	c.UpdatedAt = time.Now()
}

// GetSkillDetail 获取详细技能
func (c *Character) GetSkillDetail(skillName string) *Skill {
	if c.SkillsDetail == nil {
		return nil
	}
	return c.SkillsDetail[skillName]
}

// SetSaveDetail 设置详细豁免
func (c *Character) SetSaveDetail(saveName string, save *Save) {
	if c.SavesDetail == nil {
		c.SavesDetail = make(map[string]*Save)
	}
	c.SavesDetail[saveName] = save
	c.UpdatedAt = time.Now()
}

// GetSaveDetail 获取详细豁免
func (c *Character) GetSaveDetail(saveName string) *Save {
	if c.SavesDetail == nil {
		return nil
	}
	return c.SavesDetail[saveName]
}

// AddCurrency 添加货币
func (c *Character) AddCurrency(amount *Currency) {
	c.GetCurrency().Add(amount)
	c.UpdatedAt = time.Now()
}

// SubtractCurrency 扣除货币
func (c *Character) SubtractCurrency(amount *Currency) bool {
	success := c.GetCurrency().Subtract(amount)
	if success {
		c.UpdatedAt = time.Now()
	}
	return success
}

// IsImported 检查是否为导入角色
func (c *Character) IsImported() bool {
	return c.ImportMeta != nil
}

// SetImportMeta 设置导入元数据
func (c *Character) SetImportMeta(meta *ImportMeta) {
	c.ImportMeta = meta
	c.UpdatedAt = time.Now()
}
