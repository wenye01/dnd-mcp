// Package combat provides D&D 5e combat rule implementations
package combat

import (
	"strings"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
	"github.com/dnd-mcp/server/internal/rules/dice"
)

// AttackResult 攻击结果
// 规则参考: PHB 第9章 - Making an Attack
type AttackResult struct {
	AttackRoll   *models.DiceResult `json:"attack_roll"`   // 攻击骰
	TargetAC     int                `json:"target_ac"`     // 目标AC
	Hit          bool               `json:"hit"`           // 是否命中
	Crit         bool               `json:"crit"`          // 是否暴击
	DamageRoll   *models.DiceResult `json:"damage_roll"`   // 伤害骰
	Damage       int                `json:"damage"`        // 总伤害
	DamageType   string             `json:"damage_type"`   // 伤害类型
	TargetHP     *models.HP         `json:"target_hp"`     // 目标剩余HP
	TargetDown   bool               `json:"target_down"`   // 目标是否倒地
}

// SpellResult 法术结果
// 规则参考: PHB 第10-11章 - Spellcasting
type SpellResult struct {
	SpellID    string             `json:"spell_id"`    // 法术ID
	SpellName  string             `json:"spell_name"`  // 法术名称
	Level      int                `json:"level"`       // 施法等级
	CasterID   string             `json:"caster_id"`   // 施法者ID
	TargetIDs  []string           `json:"target_ids"`  // 目标ID列表
	DamageType string             `json:"damage_type"` // 伤害类型（如果有）
	Damage     int                `json:"damage"`      // 总伤害/治疗量
	IsHealing  bool               `json:"is_healing"`  // 是否为治疗
	Results    []SpellTargetResult `json:"results"`    // 各目标结果
}

// SpellTargetResult 法术目标结果
type SpellTargetResult struct {
	TargetID  string             `json:"target_id"`  // 目标ID
	Hit       bool               `json:"hit"`        // 是否命中（需要攻击检定的法术）
	Damage    int                `json:"damage"`     // 伤害/治疗量
	TargetHP  *models.HP         `json:"target_hp"`  // 目标剩余HP
}

// ResolveAttack 执行攻击
// 规则参考: PHB 第9章 - Making an Attack
func ResolveAttack(attacker, target *models.Character, weapon *models.EquipmentItem, advantage, disadvantage bool, roller *dice.Roller) *AttackResult {
	result := &AttackResult{
		TargetAC:   GetArmorClass(target),
		DamageType: "slashing", // 默认挥砍伤害
	}

	// 获取武器伤害类型
	if weapon != nil && weapon.DamageType != "" {
		result.DamageType = weapon.DamageType
	}

	// 计算攻击加值
	attackBonus := GetAttackBonus(attacker, weapon)

	// 投攻击骰
	// 规则参考: PHB 第9章 - Attack Rolls
	if advantage && !disadvantage {
		result.AttackRoll = roller.RollWithAdvantage(attackBonus)
	} else if disadvantage && !advantage {
		result.AttackRoll = roller.RollWithDisadvantage(attackBonus)
	} else {
		result.AttackRoll = roller.RollD20(attackBonus)
	}

	// 判定命中
	// 规则参考: PHB 第9章 - Attack Rolls / Critical Hits
	if result.AttackRoll.IsFumble() {
		// 自然 1: 自动失手
		result.Hit = false
		result.Crit = false
	} else if result.AttackRoll.IsCritical() {
		// 自然 20: 自动命中，暴击
		result.Hit = true
		result.Crit = true
	} else {
		// 普通命中判定
		result.Hit = result.AttackRoll.Total >= result.TargetAC
		result.Crit = false
	}

	// 如果命中，计算伤害
	if result.Hit {
		result.DamageRoll = RollDamage(weapon, result.Crit, roller)
		result.Damage = result.DamageRoll.Total

		// 应用伤害到目标
		if target.HP != nil {
			// 复制 HP 以避免修改原对象
			hpCopy := *target.HP
			overflow := hpCopy.TakeDamage(result.Damage)
			result.TargetHP = &hpCopy
			result.TargetDown = hpCopy.IsAtZero() || overflow > 0
		}
	}

	return result
}

// GetAttackBonus 获取攻击加值
// 规则参考: PHB 第9章 - Attack Rolls
// 公式: d20 + 属性调整值 + 熟练加值（如果熟练）
func GetAttackBonus(attacker *models.Character, weapon *models.EquipmentItem) int {
	if attacker == nil {
		return 0
	}

	bonus := 0

	// 获取属性调整值
	// 规则参考: PHB 第9章 - Modified by Ability Scores
	abilityMod := getWeaponAbilityModifier(attacker, weapon)
	bonus += abilityMod

	// 添加熟练加值（假设所有武器都熟练）
	// 规则参考: PHB 第9章 - Proficiency Bonus
	bonus += attacker.GetProficiencyBonus()

	// 添加武器魔法加值
	if weapon != nil && weapon.MagicBonus != 0 {
		bonus += weapon.MagicBonus
	}

	return bonus
}

// GetArmorClass 获取护甲等级
// 规则参考: PHB 第5章 - Armor Class
func GetArmorClass(character *models.Character) int {
	if character == nil {
		return 10
	}

	// 如果角色有预设 AC，直接使用
	if character.AC > 0 {
		return character.AC
	}

	// 计算 AC
	// 规则: AC = 10 + 敏捷调整值（无甲）
	dexMod := 0
	if character.Abilities != nil {
		dexMod = rules.GetDexterityModifier(character.Abilities)
	}

	baseAC := 10 + dexMod

	// 检查是否穿着护甲
	if character.EquipmentSlots != nil && character.EquipmentSlots.Armor != nil {
		armor := character.EquipmentSlots.Armor
		baseAC = armor.AC

		// 根据护甲类型添加敏捷加值
		// 规则参考: PHB 第5章 - Armor
		armorType := getArmorType(armor.Subtype)
		switch armorType {
		case "light":
			// 轻甲：完整敏捷加值
			baseAC += dexMod
		case "medium":
			// 中甲：敏捷加值最大 +2
			if dexMod > 2 {
				dexMod = 2
			}
			baseAC += dexMod
		case "heavy":
			// 重甲：无敏捷加值
			// 不添加任何敏捷加值
		default:
			// 未知类型，按轻甲处理
			baseAC += dexMod
		}

		// 添加护甲魔法加值
		if armor.MagicBonus != 0 {
			baseAC += armor.MagicBonus
		}
	}

	// 检查是否装备盾牌
	if character.EquipmentSlots != nil && character.EquipmentSlots.Shield != nil {
		shield := character.EquipmentSlots.Shield
		baseAC += shield.AC // 盾牌通常提供 +2 AC
		if shield.MagicBonus != 0 {
			baseAC += shield.MagicBonus
		}
	}

	// 添加其他 AC 加值（如护符、戒指等）
	if character.EquipmentSlots != nil {
		baseAC += getACBonusFromAccessories(character.EquipmentSlots)
	}

	return baseAC
}

// RollDamage 投伤害骰
// 规则参考: PHB 第9章 - Damage Rolls
func RollDamage(weapon *models.EquipmentItem, isCrit bool, roller *dice.Roller) *models.DiceResult {
	// 默认伤害：1d8（徒手攻击使用力量）
	damageFormula := "1d8"

	if weapon != nil {
		if weapon.Damage != "" {
			damageFormula = weapon.Damage
		}
	}

	// 解析伤害公式
	formula, err := dice.ParseFormula(damageFormula)
	if err != nil {
		// 解析失败，使用默认值
		formula = models.NewDiceFormula()
		formula.Count = 1
		formula.Sides = 8
		formula.Original = "1d8"
	}

	// 暴击时骰子数量翻倍
	// 规则参考: PHB 第9章 - Critical Hits
	if isCrit {
		formula.Count *= 2
		formula.Original = damageFormula + " (critical)"
	}

	result := roller.RollFormula(formula)
	return result
}

// GetInitiative 获取先攻值
// 规则参考: PHB 第9章 - Initiative
func GetInitiative(character *models.Character, roller *dice.Roller) int {
	if character == nil {
		return roller.RollD20(0).Total
	}

	// 先攻 = 1d20 + 敏捷调整值
	dexMod := 0
	if character.Abilities != nil {
		dexMod = rules.GetDexterityModifier(character.Abilities)
	}

	// 角色可能有先攻加值
	initBonus := character.Initiative
	if initBonus == 0 {
		initBonus = dexMod
	} else {
		// Initiative 字段已经包含敏捷调整值，直接使用
	}

	return roller.RollD20(initBonus).Total
}

// getWeaponAbilityModifier 获取武器使用的属性调整值
// 规则参考: PHB 第5章 - Weapons / 第9章 - Attack Rolls
func getWeaponAbilityModifier(character *models.Character, weapon *models.EquipmentItem) int {
	if character.Abilities == nil {
		return 0
	}

	strMod := rules.GetStrengthModifier(character.Abilities)
	dexMod := rules.GetDexterityModifier(character.Abilities)

	// 没有武器时使用力量（徒手）
	if weapon == nil {
		return strMod
	}

	// 检查武器属性
	isFinesse := false
	isRanged := false

	for _, prop := range weapon.Properties {
		propLower := strings.ToLower(prop)
		if propLower == "finesse" {
			isFinesse = true
		}
		if propLower == "ranged" || propLower == "thrown" {
			isRanged = true
		}
	}

	// 根据武器类型决定使用的属性
	// 规则参考: PHB 第5章 - Weapon Properties
	if isRanged {
		// 远程武器使用敏捷
		return dexMod
	}

	if isFinesse {
		// 灵巧武器可以选择力量或敏捷（取较高值）
		if dexMod > strMod {
			return dexMod
		}
		return strMod
	}

	// 默认使用力量（近战武器）
	return strMod
}

// getArmorType 获取护甲类型
func getArmorType(subtype string) string {
	subtypeLower := strings.ToLower(subtype)
	switch subtypeLower {
	case "light", "padded", "leather", "studded leather":
		return "light"
	case "medium", "hide", "chain shirt", "scale mail", "breastplate", "half plate":
		return "medium"
	case "heavy", "ring mail", "chain mail", "splint", "plate":
		return "heavy"
	default:
		return "light" // 默认按轻甲处理
	}
}

// getACBonusFromAccessories 获取饰品的 AC 加值
func getACBonusFromAccessories(slots *models.EquipmentSlots) int {
	bonus := 0

	// 检查各种饰品槽位
	accessories := []*models.EquipmentItem{
		slots.Amulet,
		slots.Ring1,
		slots.Ring2,
		slots.Cloak,
		slots.Bracers,
	}

	for _, item := range accessories {
		if item != nil && item.ACBonus != 0 {
			bonus += item.ACBonus
		}
	}

	return bonus
}
