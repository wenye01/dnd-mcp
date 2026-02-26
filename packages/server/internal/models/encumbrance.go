package models

// Size 体型类别
// 规则参考: PHB 第7章 Using Ability Scores / Size Categories
type Size string

const (
	SizeTiny      Size = "tiny"      // 微型
	SizeSmall     Size = "small"     // 小型
	SizeMedium    Size = "medium"    // 中型
	SizeLarge     Size = "large"     // 大型
	SizeHuge      Size = "huge"      // 超大型
	SizeGargantuan Size = "gargantuan" // 巨型
)

// SizeMultiplier 返回体型对应的负重倍率
// 规则参考: PHB 第7章 - Lifting and Carrying
func SizeMultiplier(size Size) float64 {
	switch size {
	case SizeTiny:
		return 0.5
	case SizeSmall:
		return 0.5
	case SizeMedium:
		return 1.0
	case SizeLarge:
		return 2.0
	case SizeHuge:
		return 4.0 // 推断：超大型的两倍
	case SizeGargantuan:
		return 8.0 // 推断：巨型的两倍
	default:
		return 1.0
	}
}

// Encumbrance 负重状态
// 规则参考: PHB 第7章 - Lifting and Carrying
type Encumbrance struct {
	Carried      float64 `json:"carried"`        // 当前负重（磅）
	Capacity     int     `json:"capacity"`       // 负重能力（磅）
	PushDragLift int     `json:"push_drag_lift"` // 推/拖/举起重量（磅）
	IsEncumbered bool    `json:"is_encumbered"`  // 是否超重
	// 变体规则字段（可选启用）
	IsHeavilyEncumbered bool `json:"is_heavily_encumbered,omitempty"` // 是否严重超重（变体规则）
}

// NewEncumbrance 创建负重状态
func NewEncumbrance() *Encumbrance {
	return &Encumbrance{}
}

// Validate 验证负重状态
func (e *Encumbrance) Validate() error {
	if e.Carried < 0 {
		return NewValidationError("encumbrance.carried", "cannot be negative")
	}
	if e.Capacity < 0 {
		return NewValidationError("encumbrance.capacity", "cannot be negative")
	}
	if e.PushDragLift < 0 {
		return NewValidationError("encumbrance.push_drag_lift", "cannot be negative")
	}
	return nil
}

// UpdateEncumbrance 更新负重状态
// 参数: carried 当前负重, capacity 负重能力, pushDragLift 推/拖/举起
func (e *Encumbrance) UpdateEncumbrance(carried float64, capacity, pushDragLift int) {
	e.Carried = carried
	e.Capacity = capacity
	e.PushDragLift = pushDragLift
	e.IsEncumbered = carried > float64(capacity)
	// 变体规则：严重超重是超过 2 倍负重能力
	e.IsHeavilyEncumbered = carried > float64(capacity*2)
}

// CanCarry 检查是否可以携带指定重量
func (e *Encumbrance) CanCarry(additionalWeight float64) bool {
	return e.Carried+additionalWeight <= float64(e.PushDragLift)
}

// CanPushDragLift 检查是否可以推/拖/举起指定重量
func (e *Encumbrance) CanPushDragLift(weight float64) bool {
	return weight <= float64(e.PushDragLift)
}

// GetRemainingCapacity 获取剩余负重能力
func (e *Encumbrance) GetRemainingCapacity() float64 {
	remaining := float64(e.Capacity) - e.Carried
	if remaining < 0 {
		return 0
	}
	return remaining
}
