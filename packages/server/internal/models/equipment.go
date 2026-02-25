package models

import (
	"time"
)

// EquipmentSlot 装备槽位类型
// 规则参考: PHB 第5章 Equipment
type EquipmentSlot string

const (
	SlotMainHand   EquipmentSlot = "main_hand"   // 主手
	SlotOffHand    EquipmentSlot = "off_hand"    // 副手
	SlotArmor      EquipmentSlot = "armor"       // 护甲
	SlotShield     EquipmentSlot = "shield"      // 盾牌
	SlotHelmet     EquipmentSlot = "helmet"      // 头盔
	SlotCloak      EquipmentSlot = "cloak"       // 披风
	SlotAmulet     EquipmentSlot = "amulet"      // 护符
	SlotRing1      EquipmentSlot = "ring1"       // 戒指1
	SlotRing2      EquipmentSlot = "ring2"       // 戒指2
	SlotBelt       EquipmentSlot = "belt"        // 腰带
	SlotBoots      EquipmentSlot = "boots"       // 靴子
	SlotGloves     EquipmentSlot = "gloves"      // 手套
	SlotBracers    EquipmentSlot = "bracers"     // 护腕
	SlotAttunement EquipmentSlot = "attunement"  // 同调物品槽
)

// AllEquipmentSlots 返回所有装备槽位
func AllEquipmentSlots() []EquipmentSlot {
	return []EquipmentSlot{
		SlotMainHand, SlotOffHand, SlotArmor, SlotShield,
		SlotHelmet, SlotCloak, SlotAmulet, SlotRing1, SlotRing2,
		SlotBelt, SlotBoots, SlotGloves, SlotBracers, SlotAttunement,
	}
}

// EquipmentSlots 装备槽位集合
// 规则参考: PHB 第5章 Equipment, DMG 第7章 Magic Items
type EquipmentSlots struct {
	MainHand   *EquipmentItem `json:"main_hand,omitempty"`   // 主手武器
	OffHand    *EquipmentItem `json:"off_hand,omitempty"`    // 副手/盾牌
	Armor      *EquipmentItem `json:"armor,omitempty"`       // 护甲
	Shield     *EquipmentItem `json:"shield,omitempty"`      // 盾牌（独立于护甲）
	Helmet     *EquipmentItem `json:"helmet,omitempty"`      // 头盔
	Cloak      *EquipmentItem `json:"cloak,omitempty"`       // 披风/斗篷
	Amulet     *EquipmentItem `json:"amulet,omitempty"`      // 护符/项链
	Ring1      *EquipmentItem `json:"ring1,omitempty"`       // 戒指1
	Ring2      *EquipmentItem `json:"ring2,omitempty"`       // 戒指2
	Belt       *EquipmentItem `json:"belt,omitempty"`        // 腰带
	Boots      *EquipmentItem `json:"boots,omitempty"`       // 靴子
	Gloves     *EquipmentItem `json:"gloves,omitempty"`      // 手套
	Bracers    *EquipmentItem `json:"bracers,omitempty"`     // 护腕
	Attunement []*EquipmentItem `json:"attunement,omitempty"` // 同调物品（最多3个）
}

// NewEquipmentSlots 创建空的装备槽位
func NewEquipmentSlots() *EquipmentSlots {
	return &EquipmentSlots{
		Attunement: make([]*EquipmentItem, 0),
	}
}

// GetSlot 获取指定槽位的装备
func (e *EquipmentSlots) GetSlot(slot EquipmentSlot) *EquipmentItem {
	switch slot {
	case SlotMainHand:
		return e.MainHand
	case SlotOffHand:
		return e.OffHand
	case SlotArmor:
		return e.Armor
	case SlotShield:
		return e.Shield
	case SlotHelmet:
		return e.Helmet
	case SlotCloak:
		return e.Cloak
	case SlotAmulet:
		return e.Amulet
	case SlotRing1:
		return e.Ring1
	case SlotRing2:
		return e.Ring2
	case SlotBelt:
		return e.Belt
	case SlotBoots:
		return e.Boots
	case SlotGloves:
		return e.Gloves
	case SlotBracers:
		return e.Bracers
	default:
		return nil
	}
}

// SetSlot 设置指定槽位的装备
func (e *EquipmentSlots) SetSlot(slot EquipmentSlot, item *EquipmentItem) {
	switch slot {
	case SlotMainHand:
		e.MainHand = item
	case SlotOffHand:
		e.OffHand = item
	case SlotArmor:
		e.Armor = item
	case SlotShield:
		e.Shield = item
	case SlotHelmet:
		e.Helmet = item
	case SlotCloak:
		e.Cloak = item
	case SlotAmulet:
		e.Amulet = item
	case SlotRing1:
		e.Ring1 = item
	case SlotRing2:
		e.Ring2 = item
	case SlotBelt:
		e.Belt = item
	case SlotBoots:
		e.Boots = item
	case SlotGloves:
		e.Gloves = item
	case SlotBracers:
		e.Bracers = item
	}
}

// AddAttunementItem 添加同调物品
// 规则参考: DMG 第7章 - 最多同时同调3个物品
func (e *EquipmentSlots) AddAttunementItem(item *EquipmentItem) bool {
	if len(e.Attunement) >= 3 {
		return false
	}
	e.Attunement = append(e.Attunement, item)
	return true
}

// RemoveAttunementItem 移除同调物品
func (e *EquipmentSlots) RemoveAttunementItem(itemID string) bool {
	for i, item := range e.Attunement {
		if item.ID == itemID {
			e.Attunement = append(e.Attunement[:i], e.Attunement[i+1:]...)
			return true
		}
	}
	return false
}

// Validate 验证装备槽位
func (e *EquipmentSlots) Validate() error {
	// 验证同调物品数量
	if len(e.Attunement) > 3 {
		return NewValidationError("attunement", "cannot have more than 3 attuned items")
	}

	// 验证各槽位物品
	slots := []*EquipmentItem{
		e.MainHand, e.OffHand, e.Armor, e.Shield,
		e.Helmet, e.Cloak, e.Amulet, e.Ring1, e.Ring2,
		e.Belt, e.Boots, e.Gloves, e.Bracers,
	}
	for _, item := range slots {
		if item != nil {
			if err := item.Validate(); err != nil {
				return err
			}
		}
	}

	// 验证同调物品
	for _, item := range e.Attunement {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// EquipmentItem 装备物品
// 规则参考: PHB 第5章 Equipment
type EquipmentItem struct {
	ID          string         `json:"id"`                    // 物品ID
	Name        string         `json:"name"`                  // 物品名称
	Type        EquipmentType  `json:"type"`                  // 装备类型
	Subtype     string         `json:"subtype,omitempty"`     // 子类型（武器类型、护甲类型等）
	Rarity      ItemRarity     `json:"rarity,omitempty"`      // 稀有度
	RequiresAttunement bool    `json:"requires_attunement,omitempty"` // 是否需要同调
	Description string         `json:"description,omitempty"` // 描述

	// 武器属性
	Damage      string         `json:"damage,omitempty"`      // 伤害骰（如 "1d8"）
	DamageType  string         `json:"damage_type,omitempty"` // 伤害类型
	Range       string         `json:"range,omitempty"`       // 射程
	Properties  []string       `json:"properties,omitempty"`  // 武器属性（如 versatile, finesse）

	// 护甲属性
	AC          int            `json:"ac,omitempty"`          // 护甲等级
	ACBonus     int            `json:"ac_bonus,omitempty"`    // AC加值
	MaxDexBonus int            `json:"max_dex_bonus,omitempty"` // 敏捷加值上限（-1表示无限制）
	StealthDisadvantage bool   `json:"stealth_disadvantage,omitempty"` // 是否导致隐匿劣势

	// 其他属性
	Weight      float64        `json:"weight,omitempty"`      // 重量（磅）
	Value       int            `json:"value,omitempty"`       // 价值（铜币）
	MagicBonus  int            `json:"magic_bonus,omitempty"` // 魔法加值

	// 原始数据（用于FVTT导入）
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
}

// EquipmentType 装备类型
type EquipmentType string

const (
	EquipmentTypeWeapon    EquipmentType = "weapon"    // 武器
	EquipmentTypeArmor     EquipmentType = "armor"     // 护甲
	EquipmentTypeShield    EquipmentType = "shield"    // 盾牌
	EquipmentTypeAccessory EquipmentType = "accessory" // 饰品
)

// ItemRarity 物品稀有度
// 规则参考: DMG 第7章 Magic Items
type ItemRarity string

const (
	RarityCommon    ItemRarity = "common"    // 普通
	RarityUncommon  ItemRarity = "uncommon"  // 非普通
	RarityRare      ItemRarity = "rare"      // 稀有
	RarityVeryRare  ItemRarity = "very_rare" // 极稀有
	RarityLegendary ItemRarity = "legendary" // 传说
	RarityArtifact  ItemRarity = "artifact"  // 圣物
)

// Validate 验证装备物品
func (e *EquipmentItem) Validate() error {
	if e.Name == "" {
		return NewValidationError("equipment.name", "cannot be empty")
	}
	return nil
}

// InventoryItem 背包物品
// 规则参考: PHB 第7章 Between Adventures
type InventoryItem struct {
	ID          string                 `json:"id"`                    // 物品ID
	Name        string                 `json:"name"`                  // 物品名称
	Quantity    int                    `json:"quantity"`              // 数量
	Weight      float64                `json:"weight,omitempty"`      // 单个重量（磅）
	TotalWeight float64                `json:"total_weight,omitempty"` // 总重量（自动计算）
	Description string                 `json:"description,omitempty"` // 描述
	Usage       string                 `json:"usage,omitempty"`       // 用途（如 "1 action"）
	Charges     int                    `json:"charges,omitempty"`     // 充能次数
	MaxCharges  int                    `json:"max_charges,omitempty"` // 最大充能
	ItemType    string                 `json:"item_type,omitempty"`   // 物品类型（consumable, tool, etc.）
	Rarity      ItemRarity             `json:"rarity,omitempty"`      // 稀有度
	Value       int                    `json:"value,omitempty"`       // 单个价值（铜币）
	RawData     map[string]interface{} `json:"raw_data,omitempty"`    // 原始数据
}

// Validate 验证背包物品
func (i *InventoryItem) Validate() error {
	if i.Name == "" {
		return NewValidationError("inventory_item.name", "cannot be empty")
	}
	if i.Quantity < 0 {
		return NewValidationError("inventory_item.quantity", "cannot be negative")
	}
	return nil
}

// CalculateTotalWeight 计算总重量
func (i *InventoryItem) CalculateTotalWeight() float64 {
	return i.Weight * float64(i.Quantity)
}

// UseCharge 使用一次充能
func (i *InventoryItem) UseCharge() bool {
	if i.Charges > 0 {
		i.Charges--
		return true
	}
	return false
}

// Recharge 恢复充能
func (i *InventoryItem) Recharge(amount int) {
	if i.MaxCharges > 0 {
		i.Charges += amount
		if i.Charges > i.MaxCharges {
			i.Charges = i.MaxCharges
		}
	}
}

// Currency 货币
// 规则参考: PHB 第5章 Equipment - Currency
type Currency struct {
	PP int `json:"pp"` // 铂金币
	GP int `json:"gp"` // 金币
	EP int `json:"ep"` // 电金币
	SP int `json:"sp"` // 银币
	CP int `json:"cp"` // 铜币
}

// NewCurrency 创建默认货币（0）
func NewCurrency() *Currency {
	return &Currency{}
}

// Validate 验证货币
func (c *Currency) Validate() error {
	if c.PP < 0 {
		return NewValidationError("currency.pp", "cannot be negative")
	}
	if c.GP < 0 {
		return NewValidationError("currency.gp", "cannot be negative")
	}
	if c.EP < 0 {
		return NewValidationError("currency.ep", "cannot be negative")
	}
	if c.SP < 0 {
		return NewValidationError("currency.sp", "cannot be negative")
	}
	if c.CP < 0 {
		return NewValidationError("currency.cp", "cannot be negative")
	}
	return nil
}

// ToCopper 转换为铜币
// 汇率: 1pp = 10gp = 100sp = 1000cp, 1ep = 5sp = 50cp
func (c *Currency) ToCopper() int {
	return c.PP*1000 + c.GP*100 + c.EP*50 + c.SP*10 + c.CP
}

// FromCopper 从铜币转换
func (c *Currency) FromCopper(copper int) {
	if copper < 0 {
		copper = 0
	}

	c.PP = copper / 1000
	copper %= 1000

	c.GP = copper / 100
	copper %= 100

	c.EP = copper / 50
	copper %= 50

	c.SP = copper / 10
	c.CP = copper % 10
}

// Add 添加货币
func (c *Currency) Add(other *Currency) {
	c.PP += other.PP
	c.GP += other.GP
	c.EP += other.EP
	c.SP += other.SP
	c.CP += other.CP
}

// Subtract 扣除货币（返回是否足够）
func (c *Currency) Subtract(other *Currency) bool {
	// 先转换为铜币计算
	totalCopper := c.ToCopper()
	otherCopper := other.ToCopper()

	if totalCopper < otherCopper {
		return false
	}

	// 扣除并转换回来
	c.FromCopper(totalCopper - otherCopper)
	return true
}

// ImportMeta 导入元数据
type ImportMeta struct {
	Format     string    `json:"format"`               // 导入格式: "fvtt", "uvtt"
	OriginalID string    `json:"original_id,omitempty"` // 原始ID
	ImportedAt time.Time `json:"imported_at"`          // 导入时间
	RawJSON    string    `json:"raw_json,omitempty"`   // 压缩后的原始JSON
	Version    string    `json:"version,omitempty"`    // 数据版本
	Source     string    `json:"source,omitempty"`     // 来源标识
}

// NewImportMeta 创建导入元数据
func NewImportMeta(format, originalID string) *ImportMeta {
	return &ImportMeta{
		Format:     format,
		OriginalID: originalID,
		ImportedAt: time.Now(),
	}
}

// Validate 验证导入元数据
func (m *ImportMeta) Validate() error {
	if m.Format == "" {
		return NewValidationError("import_meta.format", "cannot be empty")
	}
	return nil
}
