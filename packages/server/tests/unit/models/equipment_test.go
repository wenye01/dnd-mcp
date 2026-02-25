package models_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewEquipmentSlots(t *testing.T) {
	slots := models.NewEquipmentSlots()

	if slots == nil {
		t.Fatal("expected equipment slots to be created")
	}
	if slots.Attunement == nil {
		t.Error("expected attunement to be initialized")
	}
}

func TestEquipmentSlots_GetSetSlot(t *testing.T) {
	slots := models.NewEquipmentSlots()

	// Test setting main hand
	item := &models.EquipmentItem{
		ID:   "sword-001",
		Name: "Longsword",
		Type: models.EquipmentTypeWeapon,
	}
	slots.SetSlot(models.SlotMainHand, item)

	// Test getting main hand
	got := slots.GetSlot(models.SlotMainHand)
	if got == nil || got.Name != "Longsword" {
		t.Error("expected to get Longsword from main hand slot")
	}

	// Test getting empty slot
	got = slots.GetSlot(models.SlotOffHand)
	if got != nil {
		t.Error("expected nil for empty slot")
	}
}

func TestEquipmentSlots_Attunement(t *testing.T) {
	slots := models.NewEquipmentSlots()

	// Test adding attunement items
	item1 := &models.EquipmentItem{ID: "item-1", Name: "Magic Sword"}
	item2 := &models.EquipmentItem{ID: "item-2", Name: "Magic Ring"}
	item3 := &models.EquipmentItem{ID: "item-3", Name: "Magic Amulet"}
	item4 := &models.EquipmentItem{ID: "item-4", Name: "Extra Item"}

	if !slots.AddAttunementItem(item1) {
		t.Error("expected to add first attunement item")
	}
	if !slots.AddAttunementItem(item2) {
		t.Error("expected to add second attunement item")
	}
	if !slots.AddAttunementItem(item3) {
		t.Error("expected to add third attunement item")
	}

	// Should not add 4th item (max 3)
	if slots.AddAttunementItem(item4) {
		t.Error("expected to fail adding 4th attunement item")
	}

	if len(slots.Attunement) != 3 {
		t.Errorf("expected 3 attunement items, got %d", len(slots.Attunement))
	}

	// Test removing attunement item
	if !slots.RemoveAttunementItem("item-2") {
		t.Error("expected to remove attunement item")
	}
	if len(slots.Attunement) != 2 {
		t.Errorf("expected 2 attunement items after removal, got %d", len(slots.Attunement))
	}

	// Test removing non-existent item
	if slots.RemoveAttunementItem("non-existent") {
		t.Error("expected to fail removing non-existent item")
	}
}

func TestEquipmentSlots_Validate(t *testing.T) {
	slots := models.NewEquipmentSlots()

	// Valid empty slots
	if err := slots.Validate(); err != nil {
		t.Errorf("expected empty slots to be valid, got: %v", err)
	}

	// Valid with item
	slots.MainHand = &models.EquipmentItem{Name: "Sword", Type: models.EquipmentTypeWeapon}
	if err := slots.Validate(); err != nil {
		t.Errorf("expected slots with valid item to be valid, got: %v", err)
	}

	// Invalid item
	slots.MainHand = &models.EquipmentItem{Name: "", Type: models.EquipmentTypeWeapon}
	if err := slots.Validate(); err == nil {
		t.Error("expected validation error for item with empty name")
	}

	// Too many attunement items
	slots.MainHand = nil
	slots.Attunement = make([]*models.EquipmentItem, 4)
	for i := range slots.Attunement {
		slots.Attunement[i] = &models.EquipmentItem{Name: "Item", Type: models.EquipmentTypeAccessory}
	}
	if err := slots.Validate(); err == nil {
		t.Error("expected validation error for more than 3 attunement items")
	}
}

func TestEquipmentItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    *models.EquipmentItem
		wantErr bool
	}{
		{
			name: "valid weapon",
			item: &models.EquipmentItem{
				Name:       "Longsword",
				Type:       models.EquipmentTypeWeapon,
				Damage:     "1d8",
				DamageType: "slashing",
			},
			wantErr: false,
		},
		{
			name: "valid armor",
			item: &models.EquipmentItem{
				Name:  "Chain Mail",
				Type:  models.EquipmentTypeArmor,
				AC:    16,
				Weight: 55,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			item: &models.EquipmentItem{
				Name: "",
				Type: models.EquipmentTypeWeapon,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInventoryItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    *models.InventoryItem
		wantErr bool
	}{
		{
			name: "valid item",
			item: &models.InventoryItem{
				Name:     "Health Potion",
				Quantity: 5,
				Weight:   0.5,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			item: &models.InventoryItem{
				Name:     "",
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "negative quantity",
			item: &models.InventoryItem{
				Name:     "Potion",
				Quantity: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInventoryItem_CalculateTotalWeight(t *testing.T) {
	item := &models.InventoryItem{
		Name:     "Arrows",
		Quantity: 20,
		Weight:   0.05,
	}

	expected := 1.0 // 20 * 0.05
	if got := item.CalculateTotalWeight(); got != expected {
		t.Errorf("expected total weight %f, got %f", expected, got)
	}
}

func TestInventoryItem_Charges(t *testing.T) {
	item := &models.InventoryItem{
		Name:       "Wand of Magic Missiles",
		Charges:    7,
		MaxCharges: 7,
	}

	// Use charges
	if !item.UseCharge() {
		t.Error("expected to use charge")
	}
	if item.Charges != 6 {
		t.Errorf("expected 6 charges, got %d", item.Charges)
	}

	// Recharge
	item.Recharge(3)
	if item.Charges != 7 {
		t.Errorf("expected 7 charges after recharge, got %d", item.Charges)
	}

	// Recharge beyond max
	item.Recharge(10)
	if item.Charges != 7 {
		t.Errorf("expected charges capped at max (7), got %d", item.Charges)
	}
}

func TestCurrency(t *testing.T) {
	currency := models.NewCurrency()

	if currency == nil {
		t.Fatal("expected currency to be created")
	}

	// Test validation
	if err := currency.Validate(); err != nil {
		t.Errorf("expected new currency to be valid, got: %v", err)
	}

	// Test ToCopper
	currency.GP = 10
	currency.SP = 5
	expected := 1000 + 50 // 10gp + 5sp in copper
	if got := currency.ToCopper(); got != expected {
		t.Errorf("expected %d copper, got %d", expected, got)
	}

	// Test FromCopper
	// 1050cp = 1pp + 50cp = 1pp 1ep (since 50cp = 1ep)
	currency2 := models.NewCurrency()
	currency2.FromCopper(1050)
	if currency2.PP != 1 || currency2.EP != 1 {
		t.Errorf("expected 1pp 1ep, got %dpp %dep", currency2.PP, currency2.EP)
	}

	// Test Add
	c1 := &models.Currency{GP: 10}
	c2 := &models.Currency{GP: 5, SP: 20}
	c1.Add(c2)
	if c1.GP != 15 || c1.SP != 20 {
		t.Errorf("expected 15gp 20sp after add, got %dgp %dsp", c1.GP, c1.SP)
	}

	// Test Subtract
	// 20gp = 2000cp, 5gp = 500cp, result = 1500cp
	// 1500cp = 1pp + 500cp = 1pp 5gp (since 1000cp = 1pp)
	c3 := &models.Currency{GP: 20}
	c4 := &models.Currency{GP: 5}
	if !c3.Subtract(c4) {
		t.Error("expected subtract to succeed")
	}
	// After subtract, 1500cp converts to 1pp 5gp
	if c3.PP != 1 || c3.GP != 5 {
		t.Errorf("expected 1pp 5gp after subtract, got %dpp %dgp", c3.PP, c3.GP)
	}

	// Test Subtract insufficient
	c5 := &models.Currency{GP: 5}
	c6 := &models.Currency{GP: 10}
	if c5.Subtract(c6) {
		t.Error("expected subtract to fail for insufficient funds")
	}
}

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name     string
		currency *models.Currency
		wantErr  bool
		errField string
	}{
		{
			name:     "valid currency",
			currency: &models.Currency{GP: 10, SP: 5},
			wantErr:  false,
		},
		{
			name:     "negative pp",
			currency: &models.Currency{PP: -1},
			wantErr:  true,
			errField: "currency.pp",
		},
		{
			name:     "negative gp",
			currency: &models.Currency{GP: -1},
			wantErr:  true,
			errField: "currency.gp",
		},
		{
			name:     "negative sp",
			currency: &models.Currency{SP: -1},
			wantErr:  true,
			errField: "currency.sp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImportMeta(t *testing.T) {
	meta := models.NewImportMeta("fvtt", "actor-001")

	if meta == nil {
		t.Fatal("expected import meta to be created")
	}
	if meta.Format != "fvtt" {
		t.Errorf("expected format 'fvtt', got '%s'", meta.Format)
	}
	if meta.OriginalID != "actor-001" {
		t.Errorf("expected original ID 'actor-001', got '%s'", meta.OriginalID)
	}
	if meta.ImportedAt.IsZero() {
		t.Error("expected ImportedAt to be set")
	}

	// Test validation
	if err := meta.Validate(); err != nil {
		t.Errorf("expected valid import meta, got: %v", err)
	}

	// Test invalid (empty format)
	invalidMeta := &models.ImportMeta{Format: ""}
	if err := invalidMeta.Validate(); err == nil {
		t.Error("expected validation error for empty format")
	}
}

func TestAllEquipmentSlots(t *testing.T) {
	slots := models.AllEquipmentSlots()

	expectedSlots := []models.EquipmentSlot{
		models.SlotMainHand, models.SlotOffHand, models.SlotArmor, models.SlotShield,
		models.SlotHelmet, models.SlotCloak, models.SlotAmulet, models.SlotRing1, models.SlotRing2,
		models.SlotBelt, models.SlotBoots, models.SlotGloves, models.SlotBracers, models.SlotAttunement,
	}

	if len(slots) != len(expectedSlots) {
		t.Errorf("expected %d slots, got %d", len(expectedSlots), len(slots))
	}

	for i, expected := range expectedSlots {
		if slots[i] != expected {
			t.Errorf("expected slot %d to be '%s', got '%s'", i, expected, slots[i])
		}
	}
}
