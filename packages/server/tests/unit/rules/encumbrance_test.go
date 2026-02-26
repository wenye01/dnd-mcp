package rules_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
)

func TestSizeMultiplier(t *testing.T) {
	tests := []struct {
		size     models.Size
		expected float64
	}{
		{models.SizeTiny, 0.5},
		{models.SizeSmall, 0.5},
		{models.SizeMedium, 1.0},
		{models.SizeLarge, 2.0},
		{models.SizeHuge, 4.0},
		{models.SizeGargantuan, 8.0},
		{models.Size("unknown"), 1.0}, // 默认值
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			result := models.SizeMultiplier(tt.size)
			if result != tt.expected {
				t.Errorf("SizeMultiplier(%s) = %f, expected %f", tt.size, result, tt.expected)
			}
		})
	}
}

func TestCalculateCapacity(t *testing.T) {
	tests := []struct {
		name     string
		strength int
		size     models.Size
		expected int
	}{
		{"Medium STR 10", 10, models.SizeMedium, 150},
		{"Medium STR 15", 15, models.SizeMedium, 225},
		{"Medium STR 20", 20, models.SizeMedium, 300},
		{"Small STR 10", 10, models.SizeSmall, 75},   // × 0.5
		{"Large STR 10", 10, models.SizeLarge, 300},  // × 2
		{"Tiny STR 8", 8, models.SizeTiny, 60},       // 8 × 15 × 0.5
		{"Huge STR 20", 20, models.SizeHuge, 1200},   // 20 × 15 × 4
		{"Gargantuan STR 30", 30, models.SizeGargantuan, 3600}, // 30 × 15 × 8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CalculateCapacity(tt.strength, tt.size)
			if result != tt.expected {
				t.Errorf("CalculateCapacity(%d, %s) = %d, expected %d",
					tt.strength, tt.size, result, tt.expected)
			}
		})
	}
}

func TestCalculatePushDragLift(t *testing.T) {
	tests := []struct {
		name     string
		strength int
		size     models.Size
		expected int
	}{
		{"Medium STR 10", 10, models.SizeMedium, 300},
		{"Medium STR 15", 15, models.SizeMedium, 450},
		{"Small STR 10", 10, models.SizeSmall, 150},  // × 0.5
		{"Large STR 10", 10, models.SizeLarge, 600},  // × 2
		{"Huge STR 20", 20, models.SizeHuge, 2400},   // 20 × 30 × 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CalculatePushDragLift(tt.strength, tt.size)
			if result != tt.expected {
				t.Errorf("CalculatePushDragLift(%d, %s) = %d, expected %d",
					tt.strength, tt.size, result, tt.expected)
			}
		})
	}
}

func TestCalculateEquipmentWeight(t *testing.T) {
	tests := []struct {
		name     string
		slots    *models.EquipmentSlots
		expected float64
	}{
		{"nil slots", nil, 0},
		{"empty slots", &models.EquipmentSlots{}, 0},
		{"single weapon", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Longsword", Weight: 3},
		}, 3},
		{"multiple items", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Longsword", Weight: 3},
			Armor:    &models.EquipmentItem{Name: "Chain Mail", Weight: 55},
			Shield:   &models.EquipmentItem{Name: "Shield", Weight: 6},
		}, 64},
		{"with attunement", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Sword", Weight: 3},
			Attunement: []*models.EquipmentItem{
				{Name: "Ring", Weight: 1},
				{Name: "Amulet", Weight: 2},
			},
		}, 6},
		{"zero weight items", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Sword", Weight: 0},
			Armor:    &models.EquipmentItem{Name: "Robes", Weight: 0},
		}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CalculateEquipmentWeight(tt.slots)
			if result != tt.expected {
				t.Errorf("CalculateEquipmentWeight() = %f, expected %f", result, tt.expected)
			}
		})
	}
}

func TestCalculateInventoryWeight(t *testing.T) {
	tests := []struct {
		name     string
		items    []*models.InventoryItem
		expected float64
	}{
		{"nil items", nil, 0},
		{"empty items", []*models.InventoryItem{}, 0},
		{"single item", []*models.InventoryItem{
			{Name: "Rations", Quantity: 5, Weight: 2},
		}, 10}, // 2 × 5
		{"multiple items", []*models.InventoryItem{
			{Name: "Rations", Quantity: 5, Weight: 2},
			{Name: "Rope", Quantity: 1, Weight: 10},
			{Name: "Torch", Quantity: 10, Weight: 1},
		}, 30}, // 10 + 10 + 10
		{"with TotalWeight", []*models.InventoryItem{
			{Name: "Rations", Quantity: 5, Weight: 2, TotalWeight: 15}, // 使用 TotalWeight
		}, 15},
		{"nil item in slice", []*models.InventoryItem{nil, {Name: "Item", Quantity: 1, Weight: 5}}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CalculateInventoryWeight(tt.items)
			if result != tt.expected {
				t.Errorf("CalculateInventoryWeight() = %f, expected %f", result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalWeight(t *testing.T) {
	tests := []struct {
		name           string
		slots          *models.EquipmentSlots
		inventoryItems []*models.InventoryItem
		currency       *models.Currency
		expected       float64
	}{
		{"all empty", nil, nil, nil, 0},
		{"equipment only", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Sword", Weight: 3},
		}, nil, nil, 3},
		{"inventory only", nil, []*models.InventoryItem{
			{Name: "Rations", Quantity: 5, Weight: 2},
		}, nil, 10},
		{"currency only", nil, nil, &models.Currency{GP: 100}, 2}, // 100 / 50 = 2
		{"mixed", &models.EquipmentSlots{
			MainHand: &models.EquipmentItem{Name: "Sword", Weight: 3},
		}, []*models.InventoryItem{
			{Name: "Rations", Quantity: 5, Weight: 2},
		}, &models.Currency{GP: 50, SP: 25}, 14.5}, // 3 + 10 + (75/50) = 14.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.CalculateTotalWeight(
				tt.slots,
				tt.inventoryItems,
				nil, // legacyEquipment
				nil, // legacyInventory
				tt.currency,
			)
			if result != tt.expected {
				t.Errorf("CalculateTotalWeight() = %f, expected %f", result, tt.expected)
			}
		})
	}
}

func TestCalculateEncumbrance(t *testing.T) {
	// 测试完整的负重计算
	slots := &models.EquipmentSlots{
		MainHand: &models.EquipmentItem{Name: "Longsword", Weight: 3},
		Armor:    &models.EquipmentItem{Name: "Chain Mail", Weight: 55},
	}
	inventory := []*models.InventoryItem{
		{Name: "Rations", Quantity: 10, Weight: 2},
		{Name: "Rope", Quantity: 1, Weight: 10},
	}

	// 力量 15，中型生物
	enc := rules.CalculateEncumbrance(15, models.SizeMedium, slots, inventory, nil, nil, nil)

	// 负重能力 = 15 × 15 = 225
	if enc.Capacity != 225 {
		t.Errorf("Capacity = %d, expected 225", enc.Capacity)
	}

	// 推/拖/举起 = 15 × 30 = 450
	if enc.PushDragLift != 450 {
		t.Errorf("PushDragLift = %d, expected 450", enc.PushDragLift)
	}

	// 当前负重 = 3 + 55 + 20 + 10 = 88
	if enc.Carried != 88 {
		t.Errorf("Carried = %f, expected 88", enc.Carried)
	}

	// 未超重
	if enc.IsEncumbered {
		t.Error("Should not be encumbered")
	}
}

func TestIsEncumbered(t *testing.T) {
	tests := []struct {
		carried  float64
		capacity int
		expected bool
	}{
		{0, 150, false},
		{100, 150, false},
		{149, 150, false},
		{150, 150, false},
		{151, 150, true},
		{200, 150, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.IsEncumbered(tt.carried, tt.capacity)
			if result != tt.expected {
				t.Errorf("IsEncumbered(%f, %d) = %v, expected %v",
					tt.carried, tt.capacity, result, tt.expected)
			}
		})
	}
}

func TestGetEncumbranceLevel(t *testing.T) {
	strength := 10

	tests := []struct {
		name             string
		carried          float64
		useVariantRules  bool
		expected         rules.EncumbranceLevel
	}{
		// 简化规则
		{"simplified - not encumbered", 100, false, rules.EncumbranceNone},
		{"simplified - at capacity", 150, false, rules.EncumbranceNone},
		{"simplified - over capacity", 151, false, rules.EncumbranceOverCapacity},

		// 变体规则
		{"variant - not encumbered", 40, true, rules.EncumbranceNone},
		{"variant - at light threshold", 50, true, rules.EncumbranceNone}, // 边界值
		{"variant - light encumbrance", 60, true, rules.EncumbranceLight},
		{"variant - at heavy threshold", 100, true, rules.EncumbranceLight}, // 边界值
		{"variant - heavy encumbrance", 110, true, rules.EncumbranceHeavy},
		{"variant - very heavy", 200, true, rules.EncumbranceHeavy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.GetEncumbranceLevel(tt.carried, strength, tt.useVariantRules)
			if result != tt.expected {
				t.Errorf("GetEncumbranceLevel(%f, %d, %v) = %v, expected %v",
					tt.carried, strength, tt.useVariantRules, result, tt.expected)
			}
		})
	}
}

func TestGetEncumberedSpeed(t *testing.T) {
	tests := []struct {
		name       string
		baseSpeed  int
		level      rules.EncumbranceLevel
		expected   int
	}{
		{"no encumbrance", 30, rules.EncumbranceNone, 30},
		{"light - speed 30", 30, rules.EncumbranceLight, 20},
		{"heavy - speed 30", 30, rules.EncumbranceHeavy, 10},
		{"over capacity - speed 30", 30, rules.EncumbranceOverCapacity, 5},
		{"light - speed 25", 25, rules.EncumbranceLight, 15},
		{"light - speed 20", 20, rules.EncumbranceLight, 10},
		{"light - speed 15", 15, rules.EncumbranceLight, 5},
		{"light - speed 10", 10, rules.EncumbranceLight, 5}, // 最低 5
		{"heavy - speed 20", 20, rules.EncumbranceHeavy, 5},  // 20 - 20 = 0 → 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.GetEncumberedSpeed(tt.baseSpeed, tt.level)
			if result != tt.expected {
				t.Errorf("GetEncumberedSpeed(%d, %v) = %d, expected %d",
					tt.baseSpeed, tt.level, result, tt.expected)
			}
		})
	}
}

// 测试 Encumbrance 模型方法
func TestEncumbranceModel(t *testing.T) {
	t.Run("UpdateEncumbrance", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(100, 150, 300)

		if enc.Carried != 100 {
			t.Errorf("Carried = %f, expected 100", enc.Carried)
		}
		if enc.Capacity != 150 {
			t.Errorf("Capacity = %d, expected 150", enc.Capacity)
		}
		if enc.PushDragLift != 300 {
			t.Errorf("PushDragLift = %d, expected 300", enc.PushDragLift)
		}
		if enc.IsEncumbered {
			t.Error("Should not be encumbered")
		}
	})

	t.Run("IsEncumbered - over capacity", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(200, 150, 300)

		if !enc.IsEncumbered {
			t.Error("Should be encumbered")
		}
	})

	t.Run("IsHeavilyEncumbered", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(250, 100, 300)

		if !enc.IsHeavilyEncumbered {
			t.Error("Should be heavily encumbered (over 2× capacity)")
		}
	})

	t.Run("CanCarry", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(100, 150, 300)

		if !enc.CanCarry(50) {
			t.Error("Should be able to carry 50 more (100 + 50 < 300)")
		}
		if enc.CanCarry(250) {
			t.Error("Should not be able to carry 250 more (100 + 250 > 300)")
		}
	})

	t.Run("CanPushDragLift", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(100, 150, 300)

		if !enc.CanPushDragLift(300) {
			t.Error("Should be able to push/drag/lift 300")
		}
		if enc.CanPushDragLift(301) {
			t.Error("Should not be able to push/drag/lift 301")
		}
	})

	t.Run("GetRemainingCapacity", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(100, 150, 300)

		remaining := enc.GetRemainingCapacity()
		if remaining != 50 {
			t.Errorf("RemainingCapacity = %f, expected 50", remaining)
		}

		// 超重情况
		enc.UpdateEncumbrance(200, 150, 300)
		remaining = enc.GetRemainingCapacity()
		if remaining != 0 {
			t.Errorf("RemainingCapacity = %f, expected 0 (when over capacity)", remaining)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		enc := models.NewEncumbrance()
		enc.UpdateEncumbrance(100, 150, 300)

		if err := enc.Validate(); err != nil {
			t.Errorf("Validation should pass: %v", err)
		}

		// 负数测试
		enc.Carried = -1
		if err := enc.Validate(); err == nil {
			t.Error("Validation should fail for negative carried")
		}
	})
}
