// Package rules provides D&D 5e game rule implementations
package rules

import (
	"github.com/dnd-mcp/server/internal/models"
)

// EncumbranceConfig 负重规则配置
type EncumbranceConfig struct {
	// Enabled 是否启用负重规则（默认启用）
	Enabled bool
	// UseVariantRules 是否使用变体负重规则（PHB 第7章可选规则）
	// 变体规则下，超过 5×力量值会轻微超重，超过 10×力量值会严重超重
	UseVariantRules bool
	// SizeMultiplier 体型倍率覆盖（可选）
	SizeMultiplier map[models.Size]float64
}

// DefaultEncumbranceConfig 默认负重配置
func DefaultEncumbranceConfig() *EncumbranceConfig {
	return &EncumbranceConfig{
		Enabled:         true,
		UseVariantRules: false,
	}
}

// CalculateCapacity 计算负重能力
// 规则参考: PHB 第7章 - Lifting and Carrying
// 公式: 力量值 × 15（磅），根据体型修正
func CalculateCapacity(strength int, size models.Size) int {
	baseCapacity := strength * 15
	multiplier := models.SizeMultiplier(size)
	return int(float64(baseCapacity) * multiplier)
}

// CalculatePushDragLift 计算推/拖/举起重量
// 规则参考: PHB 第7章 - Push, Drag, or Lift
// 公式: 力量值 × 30（磅），根据体型修正
func CalculatePushDragLift(strength int, size models.Size) int {
	baseCapacity := strength * 30
	multiplier := models.SizeMultiplier(size)
	return int(float64(baseCapacity) * multiplier)
}

// CalculateEquipmentWeight 计算装备槽位中所有装备的总重量
func CalculateEquipmentWeight(slots *models.EquipmentSlots) float64 {
	if slots == nil {
		return 0
	}

	var total float64

	// 遍历所有装备槽位
	equipment := []*models.EquipmentItem{
		slots.MainHand,
		slots.OffHand,
		slots.Armor,
		slots.Shield,
		slots.Helmet,
		slots.Cloak,
		slots.Amulet,
		slots.Ring1,
		slots.Ring2,
		slots.Belt,
		slots.Boots,
		slots.Gloves,
		slots.Bracers,
	}

	for _, item := range equipment {
		if item != nil && item.Weight > 0 {
			total += item.Weight
		}
	}

	// 同调物品
	for _, item := range slots.Attunement {
		if item != nil && item.Weight > 0 {
			total += item.Weight
		}
	}

	return total
}

// CalculateInventoryWeight 计算背包中所有物品的总重量
func CalculateInventoryWeight(items []*models.InventoryItem) float64 {
	if items == nil {
		return 0
	}

	var total float64
	for _, item := range items {
		if item != nil {
			// 优先使用 TotalWeight，如果为 0 则计算
			if item.TotalWeight > 0 {
				total += item.TotalWeight
			} else if item.Weight > 0 {
				total += item.Weight * float64(item.Quantity)
			}
		}
	}

	return total
}

// CalculateLegacyEquipmentWeight 计算旧版 Equipment 列表的总重量（向后兼容）
func CalculateLegacyEquipmentWeight(equipment []models.Equipment) float64 {
	if equipment == nil {
		return 0
	}

	var total float64
	for _, item := range equipment {
		// 旧版 Equipment 没有重量字段，这里返回 0
		// 如果需要支持，需要扩展 Equipment 结构体
		_ = item
	}
	return total
}

// CalculateLegacyInventoryWeight 计算旧版 Item 列表的总重量（向后兼容）
func CalculateLegacyInventoryWeight(items []models.Item) float64 {
	if items == nil {
		return 0
	}

	var total float64
	for _, item := range items {
		// 旧版 Item 没有重量字段，这里返回 0
		_ = item
	}
	return total
}

// CalculateTotalWeight 计算角色总负重
// 包括：装备槽位 + 背包物品 + 货币重量
func CalculateTotalWeight(
	slots *models.EquipmentSlots,
	inventoryItems []*models.InventoryItem,
	legacyEquipment []models.Equipment,
	legacyInventory []models.Item,
	currency *models.Currency,
) float64 {
	total := CalculateEquipmentWeight(slots)
	total += CalculateInventoryWeight(inventoryItems)
	total += CalculateLegacyEquipmentWeight(legacyEquipment)
	total += CalculateLegacyInventoryWeight(legacyInventory)

	// 货币重量：50 枚硬币 = 1 磅（PHB 第5章）
	if currency != nil {
		coinCount := currency.PP + currency.GP + currency.EP + currency.SP + currency.CP
		total += float64(coinCount) / 50.0
	}

	return total
}

// CalculateEncumbrance 计算角色的负重状态
// 规则参考: PHB 第7章 - Lifting and Carrying
func CalculateEncumbrance(
	strength int,
	size models.Size,
	slots *models.EquipmentSlots,
	inventoryItems []*models.InventoryItem,
	legacyEquipment []models.Equipment,
	legacyInventory []models.Item,
	currency *models.Currency,
) *models.Encumbrance {
	enc := models.NewEncumbrance()

	// 计算负重能力
	capacity := CalculateCapacity(strength, size)
	pushDragLift := CalculatePushDragLift(strength, size)

	// 计算当前负重
	carried := CalculateTotalWeight(slots, inventoryItems, legacyEquipment, legacyInventory, currency)

	// 更新负重状态
	enc.UpdateEncumbrance(carried, capacity, pushDragLift)

	return enc
}

// IsEncumbered 判断是否超重
// 规则参考: PHB 第7章 - Lifting and Carrying
func IsEncumbered(carried float64, capacity int) bool {
	return carried > float64(capacity)
}

// EncumbranceLevel 超重等级
type EncumbranceLevel int

const (
	EncumbranceNone EncumbranceLevel = iota // 未超重
	EncumbranceLight                        // 轻微超重（变体规则）
	EncumbranceHeavy                        // 严重超重（变体规则）
	EncumbranceOverCapacity                 // 超过负重能力
)

// GetEncumbranceLevel 获取超重等级（支持变体规则）
// 规则参考: PHB 第7章 - Variant: Encumbrance
// 变体规则：
// - 轻微超重：超过 5×力量值，速度 -10
// - 严重超重：超过 10×力量值，速度 -20，劣势在属性检定
func GetEncumbranceLevel(carried float64, strength int, useVariantRules bool) EncumbranceLevel {
	if useVariantRules {
		lightThreshold := float64(strength * 5)
		heavyThreshold := float64(strength * 10)

		if carried > heavyThreshold {
			return EncumbranceHeavy
		}
		if carried > lightThreshold {
			return EncumbranceLight
		}
		return EncumbranceNone
	}

	// 简化规则：只区分超重和不超重
	capacity := float64(strength * 15)
	if carried > capacity {
		return EncumbranceOverCapacity
	}
	return EncumbranceNone
}

// GetEncumberedSpeed 获取超重后的速度
// 规则参考: PHB 第7章 - Variant: Encumbrance
func GetEncumberedSpeed(baseSpeed int, level EncumbranceLevel) int {
	switch level {
	case EncumbranceHeavy:
		// 严重超重：速度 -20，最低 5
		speed := baseSpeed - 20
		if speed < 5 {
			return 5
		}
		return speed
	case EncumbranceLight:
		// 轻微超重：速度 -10
		speed := baseSpeed - 10
		if speed < 5 {
			return 5
		}
		return speed
	case EncumbranceOverCapacity:
		// 简化规则超重：速度降为 5
		return 5
	default:
		return baseSpeed
	}
}
