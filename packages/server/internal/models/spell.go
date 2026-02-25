package models

// Spellbook 法术书
// 规则参考: PHB 第10章 Spellcasting
type Spellbook struct {
	// 法术位
	Slots map[int]*SpellSlots `json:"slots,omitempty"` // 按等级的法术位

	// 已知法术（术士、邪术师等）
	KnownSpells map[int][]string `json:"known_spells,omitempty"` // 按等级的已知法术ID

	// 准备法术（法师、牧师等）
	PreparedSpells map[int][]string `json:"prepared_spells,omitempty"` // 按等级的准备法术ID

	// 所有法术（法师法术书、所有已知法术等）
	Spells map[string]*Spell `json:"spells,omitempty"` // 所有法术，key为法术ID

	// 施法属性
	SpellcastingAbility string `json:"spellcasting_ability,omitempty"` // 施法属性（intelligence, wisdom, charisma）

	// 施法专注
	ConcentrationSpell string `json:"concentration_spell,omitempty"` // 当前专注的法术ID
	ConcentrationRounds int   `json:"concentration_rounds,omitempty"` // 专注剩余回合数
}

// SpellSlots 法术位
// 规则参考: PHB 第11章 Spell Slots
type SpellSlots struct {
	Total     int `json:"total"`     // 总法术位
	Used      int `json:"used"`      // 已使用
	MaxLevel  int `json:"max_level"` // 最大可用等级（用于术士超魔等）
}

// NewSpellbook 创建空的法术书
func NewSpellbook() *Spellbook {
	return &Spellbook{
		Slots:          make(map[int]*SpellSlots),
		Spells:         make(map[string]*Spell),
		KnownSpells:    make(map[int][]string),
		PreparedSpells: make(map[int][]string),
	}
}

// NewSpellSlots 创建法术位
func NewSpellSlots(total int) *SpellSlots {
	return &SpellSlots{
		Total: total,
		Used:  0,
	}
}

// UseSlot 使用法术位
func (s *SpellSlots) UseSlot() bool {
	if s.Used < s.Total {
		s.Used++
		return true
	}
	return false
}

// UseSlotAtLevel 在指定等级使用法术位
func (s *SpellSlots) UseSlotAtLevel(level int) bool {
	if level <= 0 || s.Used >= s.Total {
		return false
	}
	s.Used++
	return true
}

// RestoreSlots 恢复法术位
func (s *SpellSlots) RestoreSlots(count int) {
	s.Used -= count
	if s.Used < 0 {
		s.Used = 0
	}
}

// RestoreAll 恢复所有法术位
func (s *SpellSlots) RestoreAll() {
	s.Used = 0
}

// Available 返回可用法术位数量
func (s *SpellSlots) Available() int {
	return s.Total - s.Used
}

// Validate 验证法术书
func (s *Spellbook) Validate() error {
	// 验证法术位
	for level, slots := range s.Slots {
		if level < 1 || level > 9 {
			return NewValidationError("spellbook.slots", "invalid spell level")
		}
		if slots.Total < 0 {
			return NewValidationError("spellbook.slots.total", "cannot be negative")
		}
		if slots.Used < 0 {
			return NewValidationError("spellbook.slots.used", "cannot be negative")
		}
	}
	return nil
}

// AddSpell 添加法术
func (s *Spellbook) AddSpell(spell *Spell) {
	if s.Spells == nil {
		s.Spells = make(map[string]*Spell)
	}
	s.Spells[spell.ID] = spell
}

// RemoveSpell 移除法术
func (s *Spellbook) RemoveSpell(spellID string) bool {
	if s.Spells == nil {
		return false
	}
	if _, exists := s.Spells[spellID]; exists {
		delete(s.Spells, spellID)
		return true
	}
	return false
}

// GetSpell 获取法术
func (s *Spellbook) GetSpell(spellID string) *Spell {
	if s.Spells == nil {
		return nil
	}
	return s.Spells[spellID]
}

// PrepareSpell 准备法术
func (s *Spellbook) PrepareSpell(spellID string, level int) {
	if s.PreparedSpells == nil {
		s.PreparedSpells = make(map[int][]string)
	}

	spells := s.PreparedSpells[level]
	for _, id := range spells {
		if id == spellID {
			return // 已经准备
		}
	}
	s.PreparedSpells[level] = append(spells, spellID)
}

// UnprepareSpell 取消准备法术
func (s *Spellbook) UnprepareSpell(spellID string, level int) bool {
	if s.PreparedSpells == nil {
		return false
	}

	spells := s.PreparedSpells[level]
	for i, id := range spells {
		if id == spellID {
			s.PreparedSpells[level] = append(spells[:i], spells[i+1:]...)
			return true
		}
	}
	return false
}

// IsSpellPrepared 检查法术是否已准备
func (s *Spellbook) IsSpellPrepared(spellID string) bool {
	if s.PreparedSpells == nil {
		return false
	}

	for _, spells := range s.PreparedSpells {
		for _, id := range spells {
			if id == spellID {
				return true
			}
		}
	}
	return false
}

// UseSlotAtLevel 使用指定等级的法术位
func (s *Spellbook) UseSlotAtLevel(level int) bool {
	if s.Slots == nil {
		return false
	}

	slots, exists := s.Slots[level]
	if !exists {
		return false
	}

	return slots.UseSlot()
}

// RestoreAllSlots 恢复所有法术位
func (s *Spellbook) RestoreAllSlots() {
	for _, slots := range s.Slots {
		slots.RestoreAll()
	}
}

// StartConcentration 开始专注
func (s *Spellbook) StartConcentration(spellID string, rounds int) {
	s.ConcentrationSpell = spellID
	s.ConcentrationRounds = rounds
}

// EndConcentration 结束专注
func (s *Spellbook) EndConcentration() {
	s.ConcentrationSpell = ""
	s.ConcentrationRounds = 0
}

// TickConcentration 推进专注
func (s *Spellbook) TickConcentration() bool {
	if s.ConcentrationRounds > 0 {
		s.ConcentrationRounds--
		if s.ConcentrationRounds == 0 {
			s.ConcentrationSpell = ""
			return true // 专注结束
		}
	}
	return false
}

// IsConcentrating 检查是否正在专注
func (s *Spellbook) IsConcentrating() bool {
	return s.ConcentrationSpell != ""
}

// Spell 法术
// 规则参考: PHB 第10-11章
type Spell struct {
	ID          string            `json:"id"`                    // 法术ID
	Name        string            `json:"name"`                  // 法术名称
	Level       int               `json:"level"`                 // 法术等级（0为戏法）
	School      SpellSchool       `json:"school"`                // 学派
	Ritual      bool              `json:"ritual,omitempty"`      // 是否为仪式法术
	CastingTime string            `json:"casting_time"`          // 施法时间
	Range       string            `json:"range"`                 // 射程
	Components  *SpellComponents  `json:"components,omitempty"`  // 成分
	Duration    string            `json:"duration"`              // 持续时间
	Concentration bool            `json:"concentration,omitempty"` // 是否需要专注
	Description string            `json:"description"`           // 描述
	HigherLevels string           `json:"higher_levels,omitempty"` // 升阶效果

	// 伤害/效果
	Damage       *SpellDamage     `json:"damage,omitempty"`       // 伤害
	Save         *SpellSave       `json:"save,omitempty"`         // 豁免
	Healing      *SpellHealing    `json:"healing,omitempty"`      // 治疗
	AreaOfEffect *AreaOfEffect    `json:"area_of_effect,omitempty"` // 范围效果

	// 元数据
	Classes      []string         `json:"classes,omitempty"`      // 可用职业
	Source       string           `json:"source,omitempty"`       // 来源书
	RawData      map[string]interface{} `json:"raw_data,omitempty"` // 原始数据
}

// SpellSchool 法术学派
// 规则参考: PHB 第10章
type SpellSchool string

const (
	SchoolAbjuration   SpellSchool = "abjuration"   // 防护
	SchoolConjuration  SpellSchool = "conjuration"  // 咒法
	SchoolDivination   SpellSchool = "divination"   // 预言
	SchoolEnchantment  SpellSchool = "enchantment"  // 附魔
	SchoolEvocation    SpellSchool = "evocation"    // 塑能
	SchoolIllusion     SpellSchool = "illusion"     // 幻术
	SchoolNecromancy   SpellSchool = "necromancy"   // 死灵
	SchoolTransmutation SpellSchool = "transmutation" // 变化
)

// SpellComponents 法术成分
// 规则参考: PHB 第10章 Components
type SpellComponents struct {
	Verbal      bool   `json:"verbal,omitempty"`      // 言语（V）
	Somatic     bool   `json:"somatic,omitempty"`     // 姿势（S）
	Material    bool   `json:"material,omitempty"`    // 材料（M）
	Materials   string `json:"materials,omitempty"`   // 具体材料描述
	GoldCost    int    `json:"gold_cost,omitempty"`   // 材料金币消耗
	Consumed    bool   `json:"consumed,omitempty"`    // 材料是否消耗
}

// SpellDamage 法术伤害
type SpellDamage struct {
	DamageType string `json:"damage_type,omitempty"` // 伤害类型
	BaseDamage string `json:"base_damage,omitempty"` // 基础伤害（如 "3d8"）
	SpellMod   bool   `json:"spell_mod,omitempty"`   // 是否加施法属性修正
	LevelScale []string `json:"level_scale,omitempty"` // 升阶伤害（每级增加）
}

// SpellSave 法术豁免
type SpellSave struct {
	Ability   string `json:"ability,omitempty"`   // 豁免属性
	DamageHalf bool  `json:"damage_half,omitempty"` // 成功是否减半伤害
	Effect     string `json:"effect,omitempty"`     // 豁免成功效果
}

// SpellHealing 法术治疗
type SpellHealing struct {
	BaseHealing string `json:"base_healing,omitempty"` // 基础治疗（如 "1d8+施法属性修正"）
	SpellMod    bool   `json:"spell_mod,omitempty"`    // 是否加施法属性修正
	LevelScale  []string `json:"level_scale,omitempty"` // 升阶治疗
}

// AreaOfEffect 范围效果
type AreaOfEffect struct {
	Type   string  `json:"type"`             // 类型：sphere, cylinder, cone, line, cube
	Size   float64 `json:"size"`             // 大小（英尺）
	Shape  string  `json:"shape,omitempty"`  // 形状描述
}

// Validate 验证法术
func (s *Spell) Validate() error {
	if s.Name == "" {
		return NewValidationError("spell.name", "cannot be empty")
	}
	if s.Level < 0 || s.Level > 9 {
		return NewValidationError("spell.level", "must be between 0 and 9")
	}
	return nil
}

// IsCantrip 是否为戏法
func (s *Spell) IsCantrip() bool {
	return s.Level == 0
}

// GetLevelName 获取法术等级名称
func (s *Spell) GetLevelName() string {
	if s.Level == 0 {
		return "Cantrip"
	}

	levelNames := []string{
		"", "1st", "2nd", "3rd", "4th", "5th",
		"6th", "7th", "8th", "9th",
	}

	if s.Level >= 1 && s.Level <= 9 {
		return levelNames[s.Level]
	}
	return ""
}

// Feature 专长/特性
// 规则参考: PHB 第6章 Customization Options
type Feature struct {
	ID          string                 `json:"id"`                    // 特性ID
	Name        string                 `json:"name"`                  // 特性名称
	Type        FeatureType            `json:"type"`                  // 特性类型
	Source      string                 `json:"source,omitempty"`      // 来源（种族/职业/专长名）
	Level       int                    `json:"level,omitempty"`       // 获得等级
	Description string                 `json:"description"`           // 描述
	Uses        int                    `json:"uses,omitempty"`        // 使用次数
	Used        int                    `json:"used,omitempty"`        // 已使用次数
	RestoreType string                 `json:"restore_type,omitempty"` // 恢复类型（short_rest, long_rest）
	Actions     []FeatureAction        `json:"actions,omitempty"`     // 关联动作
	RawData     map[string]interface{} `json:"raw_data,omitempty"`    // 原始数据
}

// FeatureType 特性类型
type FeatureType string

const (
	FeatureTypeRacial    FeatureType = "racial"    // 种族特性
	FeatureTypeClass     FeatureType = "class"     // 职业特性
	FeatureTypeFeat      FeatureType = "feat"      // 专长
	FeatureTypeBackground FeatureType = "background" // 背景特性
	FeatureTypeItem      FeatureType = "item"      // 物品特性
	FeatureTypeOther     FeatureType = "other"     // 其他
)

// FeatureAction 特性动作
type FeatureAction struct {
	Name        string `json:"name"`                  // 动作名称
	Type        string `json:"type"`                  // 动作类型（action, bonus_action, reaction）
	Description string `json:"description,omitempty"` // 动作描述
}

// Validate 验证特性
func (f *Feature) Validate() error {
	if f.Name == "" {
		return NewValidationError("feature.name", "cannot be empty")
	}
	return nil
}

// Use 使用特性（返回是否成功）
func (f *Feature) Use() bool {
	if f.Uses <= 0 {
		return true // 无限使用
	}
	if f.Used >= f.Uses {
		return false // 已用完
	}
	f.Used++
	return true
}

// Restore 恢复使用次数
func (f *Feature) Restore(restType string) bool {
	if f.RestoreType == restType || f.RestoreType == "any" {
		f.Used = 0
		return true
	}
	return false
}

// Available 返回剩余使用次数
func (f *Feature) Available() int {
	if f.Uses <= 0 {
		return -1 // 无限
	}
	return f.Uses - f.Used
}
