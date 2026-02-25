package models_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewSpellbook(t *testing.T) {
	sb := models.NewSpellbook()

	if sb == nil {
		t.Fatal("expected spellbook to be created")
	}
	if sb.Slots == nil {
		t.Error("expected slots map to be initialized")
	}
	if sb.Spells == nil {
		t.Error("expected spells map to be initialized")
	}
	if sb.KnownSpells == nil {
		t.Error("expected known spells map to be initialized")
	}
	if sb.PreparedSpells == nil {
		t.Error("expected prepared spells map to be initialized")
	}
}

func TestNewSpellSlots(t *testing.T) {
	slots := models.NewSpellSlots(4)

	if slots == nil {
		t.Fatal("expected spell slots to be created")
	}
	if slots.Total != 4 {
		t.Errorf("expected 4 total slots, got %d", slots.Total)
	}
	if slots.Used != 0 {
		t.Errorf("expected 0 used slots, got %d", slots.Used)
	}
}

func TestSpellSlots_UseRestore(t *testing.T) {
	slots := models.NewSpellSlots(4)

	// Use slots
	if !slots.UseSlot() {
		t.Error("expected to use slot")
	}
	if slots.Used != 1 {
		t.Errorf("expected 1 used slot, got %d", slots.Used)
	}

	// Use all slots
	slots.UseSlot()
	slots.UseSlot()
	slots.UseSlot()
	if slots.UseSlot() {
		t.Error("expected to fail using more than total")
	}

	// Check available
	if slots.Available() != 0 {
		t.Errorf("expected 0 available slots, got %d", slots.Available())
	}

	// Restore some
	slots.RestoreSlots(2)
	if slots.Available() != 2 {
		t.Errorf("expected 2 available slots after restore, got %d", slots.Available())
	}

	// Restore all
	slots.RestoreAll()
	if slots.Available() != 4 {
		t.Errorf("expected 4 available slots after full restore, got %d", slots.Available())
	}
}

func TestSpellbook_Spells(t *testing.T) {
	sb := models.NewSpellbook()

	spell := &models.Spell{
		ID:          "fireball",
		Name:        "Fireball",
		Level:       3,
		School:      models.SchoolEvocation,
		CastingTime: "1 action",
		Range:       "150 feet",
		Duration:    "Instantaneous",
		Description: "A bright streak flashes from your pointing finger...",
	}

	// Add spell
	sb.AddSpell(spell)
	if len(sb.Spells) != 1 {
		t.Error("expected 1 spell in spellbook")
	}

	// Get spell
	got := sb.GetSpell("fireball")
	if got == nil || got.Name != "Fireball" {
		t.Error("expected to get Fireball spell")
	}

	// Remove spell
	if !sb.RemoveSpell("fireball") {
		t.Error("expected to remove spell")
	}
	if len(sb.Spells) != 0 {
		t.Error("expected 0 spells after removal")
	}

	// Remove non-existent
	if sb.RemoveSpell("nonexistent") {
		t.Error("expected to fail removing non-existent spell")
	}
}

func TestSpellbook_PrepareSpells(t *testing.T) {
	sb := models.NewSpellbook()

	// Add spells
	sb.AddSpell(&models.Spell{ID: "magic-missile", Name: "Magic Missile", Level: 1})
	sb.AddSpell(&models.Spell{ID: "shield", Name: "Shield", Level: 1})
	sb.AddSpell(&models.Spell{ID: "fireball", Name: "Fireball", Level: 3})

	// Prepare spells
	sb.PrepareSpell("magic-missile", 1)
	sb.PrepareSpell("fireball", 3)

	if !sb.IsSpellPrepared("magic-missile") {
		t.Error("expected magic missile to be prepared")
	}
	if !sb.IsSpellPrepared("fireball") {
		t.Error("expected fireball to be prepared")
	}
	if sb.IsSpellPrepared("shield") {
		t.Error("expected shield to not be prepared")
	}

	// Unprepare spell
	if !sb.UnprepareSpell("magic-missile", 1) {
		t.Error("expected to unprepare spell")
	}
	if sb.IsSpellPrepared("magic-missile") {
		t.Error("expected magic missile to be unprepared")
	}

	// Unprepare non-existent
	if sb.UnprepareSpell("nonexistent", 1) {
		t.Error("expected to fail unpreparing non-existent spell")
	}
}

func TestSpellbook_Slots(t *testing.T) {
	sb := models.NewSpellbook()

	// Set up spell slots
	sb.Slots[1] = models.NewSpellSlots(4)
	sb.Slots[2] = models.NewSpellSlots(3)
	sb.Slots[3] = models.NewSpellSlots(2)

	// Use slot
	if !sb.UseSlotAtLevel(1) {
		t.Error("expected to use level 1 slot")
	}
	if sb.Slots[1].Used != 1 {
		t.Errorf("expected 1 used level 1 slot, got %d", sb.Slots[1].Used)
	}

	// Use non-existent slot
	if sb.UseSlotAtLevel(9) {
		t.Error("expected to fail using non-existent level 9 slot")
	}

	// Restore all
	sb.RestoreAllSlots()
	if sb.Slots[1].Used != 0 || sb.Slots[2].Used != 0 || sb.Slots[3].Used != 0 {
		t.Error("expected all slots to be restored")
	}
}

func TestSpellbook_Concentration(t *testing.T) {
	sb := models.NewSpellbook()

	// Not concentrating initially
	if sb.IsConcentrating() {
		t.Error("expected not to be concentrating initially")
	}

	// Start concentration
	sb.StartConcentration("bless", 10)
	if !sb.IsConcentrating() {
		t.Error("expected to be concentrating")
	}
	if sb.ConcentrationSpell != "bless" {
		t.Errorf("expected concentration on 'bless', got '%s'", sb.ConcentrationSpell)
	}

	// Tick concentration
	if sb.TickConcentration() {
		t.Error("expected concentration not to end after 1 tick")
	}
	if sb.ConcentrationRounds != 9 {
		t.Errorf("expected 9 rounds remaining, got %d", sb.ConcentrationRounds)
	}

	// End concentration
	sb.EndConcentration()
	if sb.IsConcentrating() {
		t.Error("expected not to be concentrating after end")
	}
}

func TestSpellbook_Validate(t *testing.T) {
	sb := models.NewSpellbook()

	// Valid empty spellbook
	if err := sb.Validate(); err != nil {
		t.Errorf("expected valid spellbook, got: %v", err)
	}

	// Valid with slots
	sb.Slots[1] = models.NewSpellSlots(4)
	if err := sb.Validate(); err != nil {
		t.Errorf("expected valid spellbook with slots, got: %v", err)
	}

	// Invalid - level 0
	sb.Slots[0] = models.NewSpellSlots(1)
	if err := sb.Validate(); err == nil {
		t.Error("expected validation error for level 0")
	}

	// Invalid - level 10
	delete(sb.Slots, 0)
	sb.Slots[10] = models.NewSpellSlots(1)
	if err := sb.Validate(); err == nil {
		t.Error("expected validation error for level 10")
	}
}

func TestSpell_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spell   *models.Spell
		wantErr bool
	}{
		{
			name: "valid spell",
			spell: &models.Spell{
				Name:        "Fireball",
				Level:       3,
				School:      models.SchoolEvocation,
				CastingTime: "1 action",
				Range:       "150 feet",
				Duration:    "Instantaneous",
				Description: "A bright streak...",
			},
			wantErr: false,
		},
		{
			name: "valid cantrip",
			spell: &models.Spell{
				Name:        "Magic Missile",
				Level:       0,
				School:      models.SchoolEvocation,
				CastingTime: "1 action",
				Range:       "120 feet",
				Duration:    "Instantaneous",
				Description: "You create three glowing darts...",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			spell: &models.Spell{
				Name:  "",
				Level: 1,
			},
			wantErr: true,
		},
		{
			name: "level too high",
			spell: &models.Spell{
				Name:  "Wish",
				Level: 10,
			},
			wantErr: true,
		},
		{
			name: "level negative",
			spell: &models.Spell{
				Name:  "Invalid",
				Level: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spell.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSpell_IsCantrip(t *testing.T) {
	cantrip := &models.Spell{Level: 0}
	if !cantrip.IsCantrip() {
		t.Error("expected level 0 spell to be cantrip")
	}

	spell := &models.Spell{Level: 1}
	if spell.IsCantrip() {
		t.Error("expected level 1 spell not to be cantrip")
	}
}

func TestSpell_GetLevelName(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{0, "Cantrip"},
		{1, "1st"},
		{2, "2nd"},
		{3, "3rd"},
		{4, "4th"},
		{5, "5th"},
		{6, "6th"},
		{7, "7th"},
		{8, "8th"},
		{9, "9th"},
		{10, ""}, // Invalid
		{-1, ""}, // Invalid
	}

	for _, tt := range tests {
		spell := &models.Spell{Level: tt.level}
		if got := spell.GetLevelName(); got != tt.expected {
			t.Errorf("level %d: expected '%s', got '%s'", tt.level, tt.expected, got)
		}
	}
}

func TestFeature_Validate(t *testing.T) {
	tests := []struct {
		name    string
		feature *models.Feature
		wantErr bool
	}{
		{
			name: "valid feature",
			feature: &models.Feature{
				Name:        "Great Weapon Master",
				Type:        models.FeatureTypeFeat,
				Description: "You've learned to put the weight of a weapon to your advantage...",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			feature: &models.Feature{
				Name: "",
				Type: models.FeatureTypeFeat,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.feature.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeature_UseRestore(t *testing.T) {
	feature := &models.Feature{
		Name:        "Action Surge",
		Type:        models.FeatureTypeClass,
		Description: "You can push yourself beyond your normal limits...",
		Uses:        1,
		RestoreType: "short_rest",
	}

	// Use feature
	if !feature.Use() {
		t.Error("expected to use feature")
	}
	if feature.Used != 1 {
		t.Errorf("expected 1 used, got %d", feature.Used)
	}

	// Use again - should fail
	if feature.Use() {
		t.Error("expected to fail using feature again")
	}

	// Check available
	if feature.Available() != 0 {
		t.Errorf("expected 0 available, got %d", feature.Available())
	}

	// Restore with wrong type
	if feature.Restore("long_rest") {
		t.Error("expected to fail restore with wrong type")
	}

	// Restore with correct type
	if !feature.Restore("short_rest") {
		t.Error("expected to restore with correct type")
	}
	if feature.Available() != 1 {
		t.Errorf("expected 1 available after restore, got %d", feature.Available())
	}
}

func TestFeature_UnlimitedUse(t *testing.T) {
	feature := &models.Feature{
		Name:        "Darkvision",
		Type:        models.FeatureTypeRacial,
		Description: "You have superior vision in dark and dim conditions...",
		Uses:        0, // Unlimited
	}

	// Always usable
	if !feature.Use() {
		t.Error("expected to use unlimited feature")
	}
	if !feature.Use() {
		t.Error("expected to use unlimited feature again")
	}

	// Available returns -1 for unlimited
	if feature.Available() != -1 {
		t.Errorf("expected -1 for unlimited available, got %d", feature.Available())
	}
}
