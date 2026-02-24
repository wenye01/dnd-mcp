package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewAbilities(t *testing.T) {
	abilities := models.NewAbilities()

	// Standard array: 15, 14, 13, 12, 10, 8
	if abilities.Strength != 15 {
		t.Errorf("expected Strength to be 15, got %d", abilities.Strength)
	}
	if abilities.Dexterity != 14 {
		t.Errorf("expected Dexterity to be 14, got %d", abilities.Dexterity)
	}
	if abilities.Constitution != 13 {
		t.Errorf("expected Constitution to be 13, got %d", abilities.Constitution)
	}
	if abilities.Intelligence != 12 {
		t.Errorf("expected Intelligence to be 12, got %d", abilities.Intelligence)
	}
	if abilities.Wisdom != 10 {
		t.Errorf("expected Wisdom to be 10, got %d", abilities.Wisdom)
	}
	if abilities.Charisma != 8 {
		t.Errorf("expected Charisma to be 8, got %d", abilities.Charisma)
	}
}

func TestAbilities_Validate(t *testing.T) {
	tests := []struct {
		name     string
		abilities *models.Abilities
		wantErr  bool
		errField string
	}{
		{
			name:     "valid abilities",
			abilities: models.NewAbilities(),
			wantErr:  false,
		},
		{
			name: "strength too low",
			abilities: &models.Abilities{Strength: 0, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
			wantErr:  true,
			errField: "strength",
		},
		{
			name: "strength too high",
			abilities: &models.Abilities{Strength: 31, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
			wantErr:  true,
			errField: "strength",
		},
		{
			name: "dexterity invalid",
			abilities: &models.Abilities{Strength: 10, Dexterity: 0, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
			wantErr:  true,
			errField: "dexterity",
		},
		{
			name: "constitution invalid",
			abilities: &models.Abilities{Strength: 10, Dexterity: 10, Constitution: 31, Intelligence: 10, Wisdom: 10, Charisma: 10},
			wantErr:  true,
			errField: "constitution",
		},
		{
			name: "intelligence invalid",
			abilities: &models.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 0, Wisdom: 10, Charisma: 10},
			wantErr:  true,
			errField: "intelligence",
		},
		{
			name: "wisdom invalid",
			abilities: &models.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 31, Charisma: 10},
			wantErr:  true,
			errField: "wisdom",
		},
		{
			name: "charisma invalid",
			abilities: &models.Abilities{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 0},
			wantErr:  true,
			errField: "charisma",
		},
		{
			name: "boundary values",
			abilities: &models.Abilities{Strength: 1, Dexterity: 30, Constitution: 1, Intelligence: 30, Wisdom: 1, Charisma: 30},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.abilities.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				} else {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestNewHP(t *testing.T) {
	hp := models.NewHP(20)

	if hp.Max != 20 {
		t.Errorf("expected Max to be 20, got %d", hp.Max)
	}
	if hp.Current != 20 {
		t.Errorf("expected Current to be 20, got %d", hp.Current)
	}
	if hp.Temp != 0 {
		t.Errorf("expected Temp to be 0, got %d", hp.Temp)
	}
}

func TestHP_Validate(t *testing.T) {
	tests := []struct {
		name     string
		hp       *models.HP
		wantErr  bool
		errField string
	}{
		{
			name:    "valid hp",
			hp:      models.NewHP(20),
			wantErr: false,
		},
		{
			name:     "max too low",
			hp:       &models.HP{Max: 0, Current: 0, Temp: 0},
			wantErr:  true,
			errField: "hp.max",
		},
		{
			name:     "current negative",
			hp:       &models.HP{Max: 10, Current: -1, Temp: 0},
			wantErr:  true,
			errField: "hp.current",
		},
		{
			name:     "temp negative",
			hp:       &models.HP{Max: 10, Current: 5, Temp: -1},
			wantErr:  true,
			errField: "hp.temp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHP_IsDead(t *testing.T) {
	hp := models.NewHP(10)

	if hp.IsDead() {
		t.Error("expected HP not to be dead at full health")
	}

	// Note: Current implementation considers HP <= 0 as dead
	hp.Current = 0
	if !hp.IsDead() {
		t.Error("expected HP to be dead at 0 HP (current implementation)")
	}

	hp.Current = -1
	if !hp.IsDead() {
		t.Error("expected HP to be dead when negative")
	}
}

func TestHP_IsUnconscious(t *testing.T) {
	hp := models.NewHP(10)

	if hp.IsUnconscious() {
		t.Error("expected HP not to be unconscious at full health")
	}

	hp.Current = 0
	if !hp.IsUnconscious() {
		t.Error("expected HP to be unconscious at 0 HP")
	}

	hp.Current = -1
	if hp.IsUnconscious() {
		t.Error("expected HP not to be unconscious when dead")
	}
}

func TestHP_TakeDamage(t *testing.T) {
	tests := []struct {
		name           string
		hp             *models.HP
		damage         int
		expectedCurr   int
		expectedTemp   int
		expectedOverflow int
	}{
		{
			name:           "normal damage",
			hp:             models.NewHP(20),
			damage:         5,
			expectedCurr:   15,
			expectedTemp:   0,
			expectedOverflow: 0,
		},
		{
			name:           "damage with temp hp",
			hp:             &models.HP{Max: 20, Current: 20, Temp: 5},
			damage:         3,
			expectedCurr:   20,
			expectedTemp:   2,
			expectedOverflow: 0,
		},
		{
			name:           "damage exceeds temp hp",
			hp:             &models.HP{Max: 20, Current: 20, Temp: 5},
			damage:         8,
			expectedCurr:   17,
			expectedTemp:   0,
			expectedOverflow: 0,
		},
		{
			name:           "damage exceeds current hp",
			hp:             models.NewHP(10),
			damage:         15,
			expectedCurr:   0,
			expectedTemp:   0,
			expectedOverflow: 5,
		},
		{
			name:           "exact damage to zero",
			hp:             models.NewHP(10),
			damage:         10,
			expectedCurr:   0,
			expectedTemp:   0,
			expectedOverflow: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overflow := tt.hp.TakeDamage(tt.damage)

			if tt.hp.Current != tt.expectedCurr {
				t.Errorf("expected Current to be %d, got %d", tt.expectedCurr, tt.hp.Current)
			}
			if tt.hp.Temp != tt.expectedTemp {
				t.Errorf("expected Temp to be %d, got %d", tt.expectedTemp, tt.hp.Temp)
			}
			if overflow != tt.expectedOverflow {
				t.Errorf("expected overflow to be %d, got %d", tt.expectedOverflow, overflow)
			}
		})
	}
}

func TestHP_Heal(t *testing.T) {
	hp := models.NewHP(20)
	hp.Current = 10

	healed := hp.Heal(5)
	if healed != 5 {
		t.Errorf("expected healed to be 5, got %d", healed)
	}
	if hp.Current != 15 {
		t.Errorf("expected Current to be 15, got %d", hp.Current)
	}

	// Heal beyond max
	healed = hp.Heal(10)
	if healed != 5 {
		t.Errorf("expected healed to be 5 (capped at max), got %d", healed)
	}
	if hp.Current != 20 {
		t.Errorf("expected Current to be 20, got %d", hp.Current)
	}
}

func TestHP_AddTempHP(t *testing.T) {
	hp := models.NewHP(20)

	// Add temp HP
	hp.AddTempHP(5)
	if hp.Temp != 5 {
		t.Errorf("expected Temp to be 5, got %d", hp.Temp)
	}

	// Lower temp HP doesn't replace
	hp.AddTempHP(3)
	if hp.Temp != 5 {
		t.Errorf("expected Temp to remain 5, got %d", hp.Temp)
	}

	// Higher temp HP replaces
	hp.AddTempHP(10)
	if hp.Temp != 10 {
		t.Errorf("expected Temp to be 10, got %d", hp.Temp)
	}
}

func TestEquipment_Validate(t *testing.T) {
	tests := []struct {
		name     string
		eq       *models.Equipment
		wantErr  bool
	}{
		{
			name:    "valid equipment",
			eq:      &models.Equipment{ID: "1", Name: "Sword", Slot: "main_hand"},
			wantErr: false,
		},
		{
			name:    "empty name",
			eq:      &models.Equipment{ID: "1", Name: "", Slot: "main_hand"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.eq.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestItem_Validate(t *testing.T) {
	tests := []struct {
		name     string
		item     *models.Item
		wantErr  bool
	}{
		{
			name:    "valid item",
			item:    &models.Item{ID: "1", Name: "Potion", Quantity: 1},
			wantErr: false,
		},
		{
			name:    "empty name",
			item:    &models.Item{ID: "1", Name: "", Quantity: 1},
			wantErr: true,
		},
		{
			name:    "negative quantity",
			item:    &models.Item{ID: "1", Name: "Potion", Quantity: -1},
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

func TestCondition_Validate(t *testing.T) {
	tests := []struct {
		name     string
		cond     *models.Condition
		wantErr  bool
	}{
		{
			name:    "valid condition",
			cond:    &models.Condition{Type: "poisoned", Duration: 10, Source: "trap"},
			wantErr: false,
		},
		{
			name:    "permanent condition",
			cond:    &models.Condition{Type: "cursed", Duration: -1, Source: "spell"},
			wantErr: false,
		},
		{
			name:    "empty type",
			cond:    &models.Condition{Type: "", Duration: 10, Source: "trap"},
			wantErr: true,
		},
		{
			name:    "invalid duration",
			cond:    &models.Condition{Type: "poisoned", Duration: -2, Source: "trap"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cond.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCondition_IsPermanent(t *testing.T) {
	perm := &models.Condition{Type: "cursed", Duration: -1}
	if !perm.IsPermanent() {
		t.Error("expected condition to be permanent")
	}

	temp := &models.Condition{Type: "poisoned", Duration: 10}
	if temp.IsPermanent() {
		t.Error("expected condition to not be permanent")
	}
}

func TestCondition_Tick(t *testing.T) {
	cond := &models.Condition{Type: "poisoned", Duration: 3}

	// Tick 1
	if cond.Tick() {
		t.Error("expected condition not to expire at duration 2")
	}
	if cond.Duration != 2 {
		t.Errorf("expected duration to be 2, got %d", cond.Duration)
	}

	// Tick 2
	if cond.Tick() {
		t.Error("expected condition not to expire at duration 1")
	}

	// Tick 3 - should expire
	if !cond.Tick() {
		t.Error("expected condition to expire at duration 0")
	}

	// Permanent condition
	perm := &models.Condition{Type: "cursed", Duration: -1}
	if perm.Tick() {
		t.Error("expected permanent condition not to expire")
	}
	if perm.Duration != -1 {
		t.Errorf("expected permanent duration to remain -1, got %d", perm.Duration)
	}
}

func TestNewCharacter(t *testing.T) {
	character := models.NewCharacter("campaign-001", "Test Hero", false)

	if character.Name != "Test Hero" {
		t.Errorf("expected Name to be 'Test Hero', got %s", character.Name)
	}
	if character.CampaignID != "campaign-001" {
		t.Errorf("expected CampaignID to be 'campaign-001', got %s", character.CampaignID)
	}
	if character.IsNPC {
		t.Error("expected IsNPC to be false")
	}
	if character.Level != 1 {
		t.Errorf("expected Level to be 1, got %d", character.Level)
	}
	if character.Abilities == nil {
		t.Error("expected Abilities to be initialized")
	}
	if character.Skills == nil {
		t.Error("expected Skills to be initialized")
	}
	if character.Saves == nil {
		t.Error("expected Saves to be initialized")
	}
	if character.Equipment == nil {
		t.Error("expected Equipment to be initialized")
	}
	if character.Inventory == nil {
		t.Error("expected Inventory to be initialized")
	}
	if character.Conditions == nil {
		t.Error("expected Conditions to be initialized")
	}
}

func TestNewCharacter_NPC(t *testing.T) {
	character := models.NewCharacter("campaign-001", "Goblin Guard", true)

	if !character.IsNPC {
		t.Error("expected IsNPC to be true")
	}
}

func TestCharacter_Validate(t *testing.T) {
	tests := []struct {
		name      string
		character *models.Character
		wantErr   bool
		errField  string
	}{
		{
			name:      "valid player character",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "Hero", false); c.PlayerID = "player-001"; return c }(),
			wantErr:   false,
		},
		{
			name:      "valid NPC",
			character: models.NewCharacter("camp-001", "Goblin", true),
			wantErr:   false,
		},
		{
			name:      "empty name",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "", false); c.PlayerID = "player-001"; return c }(),
			wantErr:   true,
			errField:  "name",
		},
		{
			name:      "empty campaign ID",
			character: func() *models.Character { c := models.NewCharacter("", "Hero", false); c.PlayerID = "player-001"; return c }(),
			wantErr:   true,
			errField:  "campaign_id",
		},
		{
			name:      "player without player ID",
			character: models.NewCharacter("camp-001", "Hero", false),
			wantErr:   true,
			errField:  "player_id",
		},
		{
			name:      "level too low",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "Hero", false); c.PlayerID = "player-001"; c.Level = 0; return c }(),
			wantErr:   true,
			errField:  "level",
		},
		{
			name:      "level too high",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "Hero", false); c.PlayerID = "player-001"; c.Level = 21; return c }(),
			wantErr:   true,
			errField:  "level",
		},
		{
			name:      "negative AC",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "Hero", false); c.PlayerID = "player-001"; c.AC = -1; return c }(),
			wantErr:   true,
			errField:  "ac",
		},
		{
			name:      "negative speed",
			character: func() *models.Character { c := models.NewCharacter("camp-001", "Hero", false); c.PlayerID = "player-001"; c.Speed = -1; return c }(),
			wantErr:   true,
			errField:  "speed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.character.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if verr, ok := err.(*models.ValidationError); ok {
					if verr.Field != tt.errField {
						t.Errorf("expected error field '%s', got '%s'", tt.errField, verr.Field)
					}
				}
			}
		})
	}
}

func TestCharacter_TypeChecks(t *testing.T) {
	player := models.NewCharacter("camp-001", "Hero", false)
	player.PlayerID = "player-001"

	npcScripted := models.NewCharacter("camp-001", "NPC", true)
	npcScripted.NPCType = models.NPCTypeScripted

	npcGenerated := models.NewCharacter("camp-001", "NPC", true)
	npcGenerated.NPCType = models.NPCTypeGenerated

	if !player.IsPlayerCharacter() {
		t.Error("expected player to be player character")
	}
	if player.IsScriptedNPC() || player.IsGeneratedNPC() {
		t.Error("expected player to not be NPC")
	}

	if !npcScripted.IsScriptedNPC() {
		t.Error("expected NPC to be scripted NPC")
	}
	if npcScripted.IsGeneratedNPC() {
		t.Error("expected scripted NPC to not be generated")
	}

	if !npcGenerated.IsGeneratedNPC() {
		t.Error("expected NPC to be generated NPC")
	}
	if npcGenerated.IsScriptedNPC() {
		t.Error("expected generated NPC to not be scripted")
	}
}

func TestCharacter_DeadAndUnconscious(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// No HP set
	if character.IsDead() || character.IsUnconscious() {
		t.Error("expected character without HP to not be dead or unconscious")
	}

	// Set HP
	character.HP = models.NewHP(10)

	if character.IsDead() || character.IsUnconscious() {
		t.Error("expected healthy character to not be dead or unconscious")
	}

	// HP = 0: Both IsDead() and IsUnconscious() return true in current implementation
	// Note: This reflects the current HP.IsDead() implementation which returns true for Current <= 0
	character.HP.Current = 0
	if !character.IsDead() {
		t.Error("expected character to be dead at 0 HP (current implementation)")
	}
	if !character.IsUnconscious() {
		t.Error("expected character to be unconscious at 0 HP")
	}

	// HP < 0: Dead
	character.HP.Current = -1
	if !character.IsDead() {
		t.Error("expected character to be dead")
	}
}

func TestCharacter_Conditions(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// No conditions
	if character.HasCondition("poisoned") {
		t.Error("expected character to not have condition")
	}

	// Add condition
	character.AddCondition("poisoned", 10, "trap")
	if !character.HasCondition("poisoned") {
		t.Error("expected character to have poisoned condition")
	}

	// Add same condition with longer duration
	character.AddCondition("poisoned", 20, "spell")
	cond := character.Conditions[0]
	if cond.Duration != 20 {
		t.Errorf("expected duration to be updated to 20, got %d", cond.Duration)
	}

	// Remove condition
	if !character.RemoveCondition("poisoned") {
		t.Error("expected RemoveCondition to return true")
	}
	if character.HasCondition("poisoned") {
		t.Error("expected condition to be removed")
	}

	// Remove non-existent condition
	if character.RemoveCondition("nonexistent") {
		t.Error("expected RemoveCondition to return false for non-existent condition")
	}
}

func TestCharacter_TickConditions(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Add conditions with different durations
	character.AddCondition("poisoned", 1, "trap")   // Will expire
	character.AddCondition("cursed", -1, "spell")   // Permanent
	character.AddCondition("blessed", 3, "spell")   // Won't expire

	oldUpdatedAt := character.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	expired := character.TickConditions()

	if len(expired) != 1 || expired[0] != "poisoned" {
		t.Errorf("expected expired to be ['poisoned'], got %v", expired)
	}
	if len(character.Conditions) != 2 {
		t.Errorf("expected 2 remaining conditions, got %d", len(character.Conditions))
	}
	if !character.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestCharacter_DamageAndHeal(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"
	character.HP = models.NewHP(20)

	// Take damage
	overflow := character.TakeDamage(5)
	if overflow != 0 {
		t.Errorf("expected overflow to be 0, got %d", overflow)
	}
	if character.HP.Current != 15 {
		t.Errorf("expected HP to be 15, got %d", character.HP.Current)
	}

	// Heal
	healed := character.Heal(3)
	if healed != 3 {
		t.Errorf("expected healed to be 3, got %d", healed)
	}
	if character.HP.Current != 18 {
		t.Errorf("expected HP to be 18, got %d", character.HP.Current)
	}

	// Heal beyond max
	healed = character.Heal(10)
	if healed != 2 {
		t.Errorf("expected healed to be 2 (capped), got %d", healed)
	}
	if character.HP.Current != 20 {
		t.Errorf("expected HP to be 20 (max), got %d", character.HP.Current)
	}
}

func TestCharacter_Inventory(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Add item
	character.AddItem("potion-001", "Health Potion", 2)
	if len(character.Inventory) != 1 {
		t.Errorf("expected 1 item, got %d", len(character.Inventory))
	}
	if character.Inventory[0].Quantity != 2 {
		t.Errorf("expected quantity to be 2, got %d", character.Inventory[0].Quantity)
	}

	// Add same item (stack)
	character.AddItem("potion-001", "Health Potion", 3)
	if len(character.Inventory) != 1 {
		t.Errorf("expected 1 item (stacked), got %d", len(character.Inventory))
	}
	if character.Inventory[0].Quantity != 5 {
		t.Errorf("expected quantity to be 5, got %d", character.Inventory[0].Quantity)
	}

	// Remove partial
	if !character.RemoveItem("potion-001", 2) {
		t.Error("expected RemoveItem to return true")
	}
	if character.Inventory[0].Quantity != 3 {
		t.Errorf("expected quantity to be 3, got %d", character.Inventory[0].Quantity)
	}

	// Remove all
	if !character.RemoveItem("potion-001", 3) {
		t.Error("expected RemoveItem to return true")
	}
	if len(character.Inventory) != 0 {
		t.Errorf("expected 0 items, got %d", len(character.Inventory))
	}

	// Remove non-existent
	if character.RemoveItem("nonexistent", 1) {
		t.Error("expected RemoveItem to return false for non-existent item")
	}
}

func TestCharacter_Equipment(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Equip item
	sword := models.Equipment{ID: "sword-001", Name: "Longsword", Slot: "main_hand", Damage: "1d8"}
	err := character.Equip(sword)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(character.Equipment) != 1 {
		t.Errorf("expected 1 equipment, got %d", len(character.Equipment))
	}

	// Get equipment by slot
	eq := character.GetEquipmentBySlot("main_hand")
	if eq == nil || eq.Name != "Longsword" {
		t.Error("expected to get Longsword from main_hand slot")
	}

	// Replace equipment in same slot
	betterSword := models.Equipment{ID: "sword-002", Name: "Better Longsword", Slot: "main_hand", Damage: "1d10"}
	err = character.Equip(betterSword)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(character.Equipment) != 1 {
		t.Errorf("expected 1 equipment (replaced), got %d", len(character.Equipment))
	}

	// Unequip
	removed := character.Unequip("main_hand")
	if removed == nil || removed.Name != "Better Longsword" {
		t.Error("expected to unequip Better Longsword")
	}
	if len(character.Equipment) != 0 {
		t.Errorf("expected 0 equipment, got %d", len(character.Equipment))
	}

	// Unequip non-existent
	removed = character.Unequip("off_hand")
	if removed != nil {
		t.Error("expected nil for non-existent slot")
	}

	// Equip invalid
	invalidEq := models.Equipment{ID: "", Name: "", Slot: "main_hand"}
	err = character.Equip(invalidEq)
	if err == nil {
		t.Error("expected error for invalid equipment")
	}
}

func TestCharacter_SetAbilities(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Set valid abilities
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 13}
	err := character.SetAbilities(abilities)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if character.Abilities.Strength != 16 {
		t.Errorf("expected Strength to be 16, got %d", character.Abilities.Strength)
	}

	// Set invalid abilities
	invalidAbilities := &models.Abilities{Strength: 0, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10}
	err = character.SetAbilities(invalidAbilities)
	if err == nil {
		t.Error("expected error for invalid abilities")
	}

	// Set nil abilities
	err = character.SetAbilities(nil)
	if err != nil {
		t.Errorf("unexpected error for nil abilities: %v", err)
	}
	if character.Abilities != nil {
		t.Error("expected abilities to be nil")
	}
}

func TestCharacter_SetHP(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Set valid HP
	hp := models.NewHP(30)
	err := character.SetHP(hp)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if character.HP.Max != 30 {
		t.Errorf("expected HP.Max to be 30, got %d", character.HP.Max)
	}

	// Set invalid HP
	invalidHP := &models.HP{Max: 0, Current: 0, Temp: 0}
	err = character.SetHP(invalidHP)
	if err == nil {
		t.Error("expected error for invalid HP")
	}
}

func TestCharacter_SkillsAndSaves(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Set skill bonus
	character.SetSkillBonus("stealth", 5)
	if character.GetSkillBonus("stealth") != 5 {
		t.Errorf("expected stealth bonus to be 5, got %d", character.GetSkillBonus("stealth"))
	}

	// Get non-existent skill
	if character.GetSkillBonus("nonexistent") != 0 {
		t.Error("expected 0 for non-existent skill")
	}

	// Set save bonus
	character.SetSaveBonus("constitution", 4)
	if character.GetSaveBonus("constitution") != 4 {
		t.Errorf("expected constitution save to be 4, got %d", character.GetSaveBonus("constitution"))
	}

	// Get non-existent save
	if character.GetSaveBonus("nonexistent") != 0 {
		t.Error("expected 0 for non-existent save")
	}
}

func TestCharacter_UpdateBasicInfo(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	oldUpdatedAt := character.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	character.UpdateBasicInfo("Elf", "Wizard", "Sage", "Lawful Good")

	if character.Race != "Elf" {
		t.Errorf("expected Race to be Elf, got %s", character.Race)
	}
	if character.Class != "Wizard" {
		t.Errorf("expected Class to be Wizard, got %s", character.Class)
	}
	if character.Background != "Sage" {
		t.Errorf("expected Background to be Sage, got %s", character.Background)
	}
	if character.Alignment != "Lawful Good" {
		t.Errorf("expected Alignment to be Lawful Good, got %s", character.Alignment)
	}
	if !character.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}

	// Partial update (empty strings should not update)
	character.UpdateBasicInfo("", "Sorcerer", "", "")
	if character.Race != "Elf" {
		t.Errorf("expected Race to remain Elf, got %s", character.Race)
	}
	if character.Class != "Sorcerer" {
		t.Errorf("expected Class to be Sorcerer, got %s", character.Class)
	}
}

func TestCharacter_SetLevel(t *testing.T) {
	character := models.NewCharacter("camp-001", "Hero", false)
	character.PlayerID = "player-001"

	// Valid level
	err := character.SetLevel(5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if character.Level != 5 {
		t.Errorf("expected Level to be 5, got %d", character.Level)
	}

	// Invalid levels
	err = character.SetLevel(0)
	if err == nil {
		t.Error("expected error for level 0")
	}

	err = character.SetLevel(21)
	if err == nil {
		t.Error("expected error for level 21")
	}
}
