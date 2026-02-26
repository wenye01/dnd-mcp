// Package rules provides D&D 5e game rule implementations
package rules

import (
	"github.com/dnd-mcp/server/internal/models"
)

// 死亡豁免规则
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points

// DeathSaveResult 死亡豁免结果
type DeathSaveResult string

const (
	// DeathSaveSuccess 豁免成功 (>= 10)
	DeathSaveSuccess DeathSaveResult = "success"
	// DeathSaveFailure 豁免失败 (< 10)
	DeathSaveFailure DeathSaveResult = "failure"
	// DeathSaveCriticalSuccess 大成功 (自然 20)
	DeathSaveCriticalSuccess DeathSaveResult = "critical_success"
	// DeathSaveCriticalFailure 大失败 (自然 1)
	DeathSaveCriticalFailure DeathSaveResult = "critical_failure"
)

// HPState HP状态
type HPState string

const (
	// HPStateNormal 正常状态
	HPStateNormal HPState = "normal"
	// HPStateUnconscious 昏迷状态 (HP = 0，未稳定)
	HPStateUnconscious HPState = "unconscious"
	// HPStateStable 稳定状态 (HP = 0，已稳定)
	HPStateStable HPState = "stable"
	// HPStateDead 死亡状态
	HPStateDead HPState = "dead"
)

// IsUnconscious 检查是否处于昏迷状态
// 规则: HP = 0 且未稳定时昏迷
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points
func IsUnconscious(hp *models.HP, deathSaves *models.DeathSaves) bool {
	if hp == nil {
		return false
	}

	// HP = 0 且未死亡且未稳定
	if hp.Current == 0 && !IsStable(deathSaves) && !isDeadByDeathSaves(deathSaves) {
		return true
	}

	return false
}

// IsDead 检查是否死亡
// 规则:
// 1. 死亡豁免失败 3 次
// 2. 受到 >= MaxHP 的溢出伤害 (instant death)
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Instant Death
func IsDead(hp *models.HP, deathSaves *models.DeathSaves) bool {
	// 死亡豁免失败 3 次
	if isDeadByDeathSaves(deathSaves) {
		return true
	}

	// 受到 >= MaxHP 的溢出伤害导致立即死亡
	// HP.Current 可以为负数，表示溢出伤害
	// 当溢出伤害 >= MaxHP 时，角色立即死亡
	if hp != nil && hp.Current < 0 {
		overflow := -hp.Current
		if overflow >= hp.Max {
			return true
		}
	}

	return false
}

// IsStable 检查是否处于稳定状态
// 规则: 死亡豁免成功 3 次
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Stabilizing
func IsStable(deathSaves *models.DeathSaves) bool {
	if deathSaves == nil {
		return false
	}
	return deathSaves.Successes >= 3
}

// isDeadByDeathSaves 检查是否因死亡豁免失败而死亡
func isDeadByDeathSaves(deathSaves *models.DeathSaves) bool {
	if deathSaves == nil {
		return false
	}
	return deathSaves.Failures >= 3
}

// GetHPState 获取当前HP状态
func GetHPState(hp *models.HP, deathSaves *models.DeathSaves) HPState {
	if IsDead(hp, deathSaves) {
		return HPStateDead
	}
	if IsStable(deathSaves) && hp != nil && hp.Current == 0 {
		return HPStateStable
	}
	if IsUnconscious(hp, deathSaves) {
		return HPStateUnconscious
	}
	return HPStateNormal
}

// MakeDeathSave 执行死亡豁免
// 规则:
// - 自然 20: 大成功，恢复 1 HP
// - 自然 1: 大失败，失败 +2
// - >= 10: 成功，成功 +1
// - < 10: 失败，失败 +1
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Death Saving Throws
func MakeDeathSave(deathSaves *models.DeathSaves, roll int) (DeathSaveResult, bool) {
	if deathSaves == nil {
		return DeathSaveFailure, false
	}

	// 自然 20: 大成功，恢复 1 HP
	if roll == 20 {
		return DeathSaveCriticalSuccess, true
	}

	// 自然 1: 大失败，失败 +2
	if roll == 1 {
		deathSaves.AddFailure()
		deathSaves.AddFailure()
		return DeathSaveCriticalFailure, isDeadByDeathSaves(deathSaves)
	}

	// >= 10: 成功
	if roll >= 10 {
		deathSaves.AddSuccess()
		return DeathSaveSuccess, IsStable(deathSaves)
	}

	// < 10: 失败
	deathSaves.AddFailure()
	return DeathSaveFailure, isDeadByDeathSaves(deathSaves)
}

// TakeDamageWhileUnconscious 昏迷状态下受到伤害
// 规则:
// - 受到伤害: 失败 +1
// - 暴击伤害: 失败 +2 (因为暴击对昏迷目标是自动命中且有额外伤害)
// - 受到 >= MaxHP 的伤害: 立即死亡
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Damage at 0 Hit Points
func TakeDamageWhileUnconscious(hp *models.HP, deathSaves *models.DeathSaves, damage int, isCrit bool) (died bool) {
	if hp == nil || deathSaves == nil {
		return false
	}

	// 检查是否造成溢出伤害导致立即死亡
	// 当 HP = 0 时，受到的伤害如果 >= MaxHP，立即死亡
	if damage >= hp.Max {
		// 立即死亡
		deathSaves.Failures = 3
		return true
	}

	// 暴击: 失败 +2
	if isCrit {
		deathSaves.AddFailure()
		deathSaves.AddFailure()
	} else {
		// 普通伤害: 失败 +1
		deathSaves.AddFailure()
	}

	// 检查是否死亡
	return isDeadByDeathSaves(deathSaves)
}

// HealFromUnconscious 从昏迷状态恢复
// 规则: 受到任何治疗，HP 恢复，脱离昏迷，重置死亡豁免
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points
func HealFromUnconscious(hp *models.HP, deathSaves *models.DeathSaves, healing int) int {
	if hp == nil || healing <= 0 {
		return 0
	}

	// 重置死亡豁免
	if deathSaves != nil {
		deathSaves.Reset()
	}

	// 恢复 HP
	return hp.Heal(healing)
}

// Stabilize 稳定角色（不通过死亡豁免的方式）
// 规则: 成功的 Medicine 检定 (DC 10) 或其他效果可以稳定角色
// 规则参考: PHB 第9章 - Dropping to 0 Hit Points / Stabilizing
func Stabilize(deathSaves *models.DeathSaves) {
	if deathSaves == nil {
		return
	}
	deathSaves.Successes = 3
	deathSaves.Failures = 0
}

// ResetDeathSaves 重置死亡豁免（用于 HP 恢复或稳定后再次受伤）
func ResetDeathSaves(deathSaves *models.DeathSaves) {
	if deathSaves == nil {
		return
	}
	deathSaves.Reset()
}
