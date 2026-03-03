// Package models provides domain models and business logic
package models

// PHB 附录A - Conditions 的 15 种状态类型
const (
	ConditionBlinded      = "blinded"
	ConditionCharmed      = "charmed"
	ConditionDeafened     = "deafened"
	ConditionFrightened   = "frightened"
	ConditionGrappled     = "grappled"
	ConditionIncapacitated = "incapacitated"
	ConditionInvisible    = "invisible"
	ConditionParalyzed    = "paralyzed"
	ConditionPetrified    = "petrified"
	ConditionPoisoned     = "poisoned"
	ConditionProne        = "prone"
	ConditionRestrained   = "restrained"
	ConditionStunned      = "stunned"
	ConditionUnconscious  = "unconscious"
	ConditionExhaustion   = "exhaustion"
)

// 所有有效的状态类型
var ValidConditionTypes = map[string]bool{
	ConditionBlinded:      true,
	ConditionCharmed:      true,
	ConditionDeafened:     true,
	ConditionFrightened:   true,
	ConditionGrappled:     true,
	ConditionIncapacitated: true,
	ConditionInvisible:    true,
	ConditionParalyzed:    true,
	ConditionPetrified:    true,
	ConditionPoisoned:     true,
	ConditionProne:        true,
	ConditionRestrained:   true,
	ConditionStunned:      true,
	ConditionUnconscious:  true,
	ConditionExhaustion:   true,
}

// IsValidConditionType 检查状态类型是否有效
func IsValidConditionType(conditionType string) bool {
	return ValidConditionTypes[conditionType]
}

// ConditionEffect 状态效果描述
// 规则参考: PHB 附录A - Conditions
type ConditionEffect struct {
	Effects       []string          `json:"effects"`       // 效果描述列表
	Disadvantages []string          `json:"disadvantages"` // 劣势检定列表
	Advantages    []string          `json:"advantages"`    // 优势检定列表
	Immunities    []string          `json:"immunities"`    // 免疫列表
	OtherEffects  map[string]string `json:"other_effects"` // 其他特殊效果
}

// 状态效果映射表
var conditionEffects = map[string]ConditionEffect{
	ConditionBlinded: {
		Effects: []string{
			"无法看见",
			"无法看见的攻击者对目标攻击具有优势",
			"目标的攻击检定具有劣势",
			"目标的察觉（Perception）检定依赖于视觉时具有劣势",
		},
		Disadvantages: []string{"attack_rolls", "perception_visual"},
		Advantages:    []string{"attack_against"},
	},

	ConditionCharmed: {
		Effects: []string{
			"魅惑者对目标有优势",
			"目标无法对魅惑者发动攻击",
			"魅惑者的社交类检定对目标有优势",
		},
		Advantages: []string{"social_checks"},
		OtherEffects: map[string]string{
			"cannot_attack_charmer": "true",
		},
	},

	ConditionDeafened: {
		Effects: []string{
			"无法听见",
			"依赖于听觉的察觉（Perception）检定具有劣势",
		},
		Disadvantages: []string{"perception_auditory"},
	},

	ConditionFrightened: {
		Effects: []string{
			"恐惧源对目标有优势",
			"目标的攻击检定具有劣势，除非恐惧源在其触及范围外",
			"目标必须使用其行动尽可能远离恐惧源",
		},
		Disadvantages: []string{"attack_rolls"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"must_flee": "true",
		},
	},

	ConditionGrappled: {
		Effects: []string{
			"速度变为0",
			"无法获得任何加值或状态带来的速度提升",
		},
		Disadvantages: []string{},
		OtherEffects: map[string]string{
			"speed_zero": "true",
		},
	},

	ConditionIncapacitated: {
		Effects: []string{
			"无法执行动作或反应",
		},
		OtherEffects: map[string]string{
			"cannot_act": "true",
			"cannot_cast_spells": "true",
		},
	},

	ConditionInvisible: {
		Effects: []string{
			"无法被看见",
			"无法看见的攻击者对目标的攻击具有劣势",
			"目标的攻击检定具有优势",
		},
		Disadvantages: []string{"attack_against"},
		Advantages:    []string{"attack_rolls"},
	},

	ConditionParalyzed: {
		Effects: []string{
			"无法移动或说话",
			"陷入失能（Incapacitated）状态",
			"对所有敏捷豁免具有劣势",
			"对能看见该目标的攻击者，攻击检定具有优势",
			"攻击命中自动造成暴击（Critical Hit）",
		},
		Disadvantages: []string{"dexterity_saves"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"auto_crit": "true",
			"speed_zero": "true",
			"incapacitated": "true",
		},
	},

	ConditionPetrified: {
		Effects: []string{
			"变为石头",
			"抵抗所有毒素和疲劳",
			"陷入失能（Incapacitated）状态",
			"无法移动或说话",
			"对所有力量和敏捷豁免具有劣势",
			"对能看见该目标的攻击者，攻击检定具有优势",
			"对毒素和疲劳免疫",
		},
		Disadvantages: []string{"strength_saves", "dexterity_saves"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"incapacitated": "true",
			"immune_poison": "true",
			"immune_exhaustion": "true",
		},
	},

	ConditionPoisoned: {
		Effects: []string{
			"所有攻击检定和属性检定具有劣势",
		},
		Disadvantages: []string{"attack_rolls", "ability_checks"},
	},

	ConditionProne: {
		Effects: []string{
			"仅能爬行，除非爬起否则速度减半",
			"对能看见该目标的攻击者，攻击检定具有劣势",
			"对5英尺内目标的攻击具有优势",
		},
		Disadvantages: []string{"attack_rolls"},
		Advantages:    []string{"attack_melee_within_5ft"},
		OtherEffects: map[string]string{
			"speed_halved": "true",
		},
	},

	ConditionRestrained: {
		Effects: []string{
			"速度变为0",
			"对所有敏捷豁免具有劣势",
			"对能看见该目标的攻击者，攻击检定具有优势",
		},
		Disadvantages: []string{"dexterity_saves"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"speed_zero": "true",
		},
	},

	ConditionStunned: {
		Effects: []string{
			"陷入失能（Incapacitated）状态",
			"无法移动或说话",
			"仅能反射性地做出反应",
			"对所有敏捷豁免具有劣势",
			"对能看见该目标的攻击者，攻击检定具有优势",
		},
		Disadvantages: []string{"dexterity_saves"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"incapacitated": "true",
		},
	},

	ConditionUnconscious: {
		Effects: []string{
			"陷入失能（Incapacitated）状态",
			"无法移动或说话",
			"对其所有的敏捷和力量豁免具有劣势",
			"对能看见该目标的攻击者，攻击检定具有优势",
			"攻击命中自动造成暴击（Critical Hit）",
		},
		Disadvantages: []string{"dexterity_saves", "strength_saves"},
		Advantages:    []string{"attack_against"},
		OtherEffects: map[string]string{
			"auto_crit": "true",
			"incapacitated": "true",
		},
	},

	ConditionExhaustion: {
		Effects: []string{
			"力竭有6个等级，特殊效果取决于等级",
		},
		OtherEffects: map[string]string{
			"level_1": "disadvantage_on_ability_checks",
			"level_2": "speed_halved",
			"level_3": "disadvantage_on_attack_rolls_and_saving_throws",
			"level_4": "hit_point_maximum_halved",
			"level_5": "speed_zero",
			"level_6": "death",
		},
	},
}

// GetConditionEffect 获取状态效果
func GetConditionEffect(conditionType string) ConditionEffect {
	if effect, ok := conditionEffects[conditionType]; ok {
		return effect
	}
	return ConditionEffect{
		Effects: []string{"未知状态效果"},
	}
}

// GetExhaustionEffect 获取力竭效果描述
// 规则参考: PHB 附录A - Exhaustion
func GetExhaustionEffect(level int) ConditionEffect {
	effects := make([]string, 0)
	disadvantages := make([]string, 0)
	otherEffects := make(map[string]string)

	switch level {
	case 1:
		effects = append(effects, "属性检定具有劣势")
		disadvantages = append(disadvantages, "ability_checks")
	case 2:
		effects = append(effects, "属性检定具有劣势", "速度减半")
		disadvantages = append(disadvantages, "ability_checks")
		otherEffects["speed_halved"] = "true"
	case 3:
		effects = append(effects, "属性检定具有劣势", "速度减半", "攻击检定和豁免具有劣势")
		disadvantages = append(disadvantages, "ability_checks", "attack_rolls", "saving_throws")
		otherEffects["speed_halved"] = "true"
	case 4:
		effects = append(effects, "属性检定具有劣势", "速度减半", "攻击检定和豁免具有劣势", "生命值上限减半")
		disadvantages = append(disadvantages, "ability_checks", "attack_rolls", "saving_throws")
		otherEffects["speed_halved"] = "true"
		otherEffects["hp_max_halved"] = "true"
	case 5:
		effects = append(effects, "属性检定具有劣势", "速度减半", "攻击检定和豁免具有劣势", "生命值上限减半", "速度变为0")
		disadvantages = append(disadvantages, "ability_checks", "attack_rolls", "saving_throws")
		otherEffects["speed_zero"] = "true"
		otherEffects["hp_max_halved"] = "true"
	case 6:
		effects = append(effects, "死亡")
		otherEffects["death"] = "true"
	}

	return ConditionEffect{
		Effects:       effects,
		Disadvantages: disadvantages,
		OtherEffects:  otherEffects,
	}
}

// ExtractExhaustionLevel 从 Source 字符串提取力竭等级
func ExtractExhaustionLevel(source string) int {
	level := 1
	// 从 "(Level X)" 格式提取等级
	if len(source) > 8 {
		for i := 0; i < len(source)-7; i++ {
			if source[i:i+7] == "(Level " {
				j := i + 7
				level = 0
				for j < len(source) && source[j] >= '0' && source[j] <= '9' {
					level = level*10 + int(source[j]-'0')
					j++
				}
				break
			}
		}
	}
	return level
}
