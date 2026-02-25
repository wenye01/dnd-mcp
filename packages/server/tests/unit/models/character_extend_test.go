package models_test

import (
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewSpeed(t *testing.T) {
	speed := models.NewSpeed(30)

	if speed == nil {
		t.Fatal("expected speed to be created")
	}
	if speed.Walk != 30 {
		t.Errorf("expected walk speed 30, got %d", speed.Walk)
	}
	if speed.Custom == nil {
		t.Error("expected custom map to be initialized")
	}
}

func TestSpeed_GetSet(t *testing.T) {
	speed := models.NewSpeed(30)

	// Test built-in speeds
	speed.SetSpeed("walk", 35)
	if speed.GetSpeed("walk") != 35 {
		t.Errorf("expected walk speed 35, got %d", speed.GetSpeed("walk"))
	}

	speed.SetSpeed("fly", 60)
	if speed.GetSpeed("fly") != 60 {
		t.Errorf("expected fly speed 60, got %d", speed.GetSpeed("fly"))
	}

	speed.SetSpeed("swim", 30)
	if speed.GetSpeed("swim") != 30 {
		t.Errorf("expected swim speed 30, got %d", speed.GetSpeed("swim"))
	}

	// Test custom speed
	speed.SetSpeed("teleport", 100)
	if speed.GetSpeed("teleport") != 100 {
		t.Errorf("expected teleport speed 100, got %d", speed.GetSpeed("teleport"))
	}

	// Test non-existent speed
	if speed.GetSpeed("nonexistent") != 0 {
		t.Error("expected 0 for non-existent speed")
	}
}

func TestSpeed_Validate(t *testing.T) {
	speed := models.NewSpeed(30)

	// Valid speed
	if err := speed.Validate(); err != nil {
		t.Errorf("expected valid speed, got: %v", err)
	}

	// Invalid - negative walk
	speed.Walk = -1
	if err := speed.Validate(); err == nil {
		t.Error("expected validation error for negative walk speed")
	}

	// Invalid - negative fly
	speed.Walk = 30
	speed.Fly = -1
	if err := speed.Validate(); err == nil {
		t.Error("expected validation error for negative fly speed")
	}
}

func TestDeathSaves(t *testing.T) {
	ds := models.NewDeathSaves()

	if ds == nil {
		t.Fatal("expected death saves to be created")
	}

	// Test initial state
	if ds.IsStable() || ds.IsDead() {
		t.Error("expected new death saves to be neither stable nor dead")
	}

	// Test adding successes
	ds.AddSuccess()
	ds.AddSuccess()
	if ds.Successes != 2 {
		t.Errorf("expected 2 successes, got %d", ds.Successes)
	}
	if ds.IsStable() {
		t.Error("expected not stable with only 2 successes")
	}

	ds.AddSuccess()
	if !ds.IsStable() {
		t.Error("expected stable with 3 successes")
	}

	// Test reset
	ds.Reset()
	if ds.Successes != 0 || ds.Failures != 0 {
		t.Error("expected reset to clear all")
	}

	// Test failures
	ds.AddFailure()
	ds.AddFailure()
	ds.AddFailure()
	if !ds.IsDead() {
		t.Error("expected dead with 3 failures")
	}
}

func TestDeathSaves_Validate(t *testing.T) {
	ds := models.NewDeathSaves()

	// Valid
	if err := ds.Validate(); err != nil {
		t.Errorf("expected valid death saves, got: %v", err)
	}

	// Invalid - too many successes
	ds.Successes = 4
	if err := ds.Validate(); err == nil {
		t.Error("expected validation error for too many successes")
	}

	// Invalid - too many failures
	ds.Successes = 0
	ds.Failures = 4
	if err := ds.Validate(); err == nil {
		t.Error("expected validation error for too many failures")
	}

	// Invalid - negative
	ds.Failures = -1
	if err := ds.Validate(); err == nil {
		t.Error("expected validation error for negative failures")
	}
}

func TestSkillDetail_CalculateBonus(t *testing.T) {
	tests := []struct {
		name             string
		skill            *models.Skill
		abilityMod       int
		proficiencyBonus int
		expected         int
	}{
		{
			name:             "no proficiency",
			skill:            &models.Skill{Ability: "dexterity"},
			abilityMod:       3,
			proficiencyBonus: 2,
			expected:         3,
		},
		{
			name:             "proficient",
			skill:            &models.Skill{Ability: "dexterity", Proficient: true},
			abilityMod:       3,
			proficiencyBonus: 2,
			expected:         5,
		},
		{
			name:             "expertise",
			skill:            &models.Skill{Ability: "dexterity", Proficient: true, Expertise: true},
			abilityMod:       3,
			proficiencyBonus: 2,
			expected:         7,
		},
		{
			name:             "override",
			skill:            &models.Skill{Ability: "dexterity", Override: 10},
			abilityMod:       3,
			proficiencyBonus: 2,
			expected:         10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.CalculateBonus(tt.abilityMod, tt.proficiencyBonus)
			if got != tt.expected {
				t.Errorf("expected bonus %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestSkill_Validate(t *testing.T) {
	// Valid skill
	skill := &models.Skill{Ability: "dexterity"}
	if err := skill.Validate(); err != nil {
		t.Errorf("expected valid skill, got: %v", err)
	}

	// Invalid - empty ability
	skill = &models.Skill{Ability: ""}
	if err := skill.Validate(); err == nil {
		t.Error("expected validation error for empty ability")
	}
}

func TestSave_CalculateBonus(t *testing.T) {
	tests := []struct {
		name             string
		save             *models.Save
		abilityMod       int
		proficiencyBonus int
		expected         int
	}{
		{
			name:             "no proficiency",
			save:             &models.Save{},
			abilityMod:       2,
			proficiencyBonus: 3,
			expected:         2,
		},
		{
			name:             "proficient",
			save:             &models.Save{Proficient: true},
			abilityMod:       2,
			proficiencyBonus: 3,
			expected:         5,
		},
		{
			name:             "override",
			save:             &models.Save{Override: 8},
			abilityMod:       2,
			proficiencyBonus: 3,
			expected:         8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.save.CalculateBonus(tt.abilityMod, tt.proficiencyBonus)
			if got != tt.expected {
				t.Errorf("expected bonus %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestBiography(t *testing.T) {
	bio := models.NewBiography()

	if bio == nil {
		t.Fatal("expected biography to be created")
	}
	if bio.Custom == nil {
		t.Error("expected custom map to be initialized")
	}

	// Set fields
	bio.Age = "25"
	bio.Height = "6'0\""
	bio.Weight = "180 lbs"
	bio.Eyes = "Blue"
	bio.Backstory = "A brave adventurer..."

	if bio.Age != "25" {
		t.Errorf("expected age '25', got '%s'", bio.Age)
	}
}

func TestTraits(t *testing.T) {
	traits := models.NewTraits()

	if traits == nil {
		t.Fatal("expected traits to be created")
	}

	// Test resistances
	traits.AddResistance("fire")
	if !traits.HasResistance("fire") {
		t.Error("expected fire resistance")
	}
	traits.AddResistance("fire") // Duplicate, should not add again
	if len(traits.DamageResistances) != 1 {
		t.Errorf("expected 1 resistance, got %d", len(traits.DamageResistances))
	}
	traits.RemoveResistance("fire")
	if traits.HasResistance("fire") {
		t.Error("expected no fire resistance after removal")
	}

	// Test immunities
	traits.AddImmunity("poison")
	if !traits.HasImmunity("poison") {
		t.Error("expected poison immunity")
	}

	// Test condition immunities
	traits.AddConditionImmunity("poisoned")
	if !traits.HasConditionImmunity("poisoned") {
		t.Error("expected poisoned condition immunity")
	}

	// Test languages
	traits.AddLanguage("Common")
	traits.AddLanguage("Elvish")
	if !traits.HasLanguage("Common") || !traits.HasLanguage("Elvish") {
		t.Error("expected Common and Elvish languages")
	}

	// Test senses
	traits.AddSense("darkvision", 60)
	if !traits.HasSense("darkvision") {
		t.Error("expected darkvision sense")
	}
	if traits.Senses["darkvision"] != 60 {
		t.Errorf("expected darkvision 60, got %d", traits.Senses["darkvision"])
	}
}

func TestCharacter_ExtendedFields(t *testing.T) {
	char := models.NewCharacter("campaign-001", "Test Hero", false)
	char.PlayerID = "player-001"

	// Test Image
	char.Image = "base64encodedimage"
	if char.Image != "base64encodedimage" {
		t.Error("expected image to be set")
	}

	// Test Experience
	char.Experience = 3000
	if char.Experience != 3000 {
		t.Error("expected experience to be set")
	}

	// Test Proficiency
	char.Proficiency = 3
	if char.Proficiency != 3 {
		t.Error("expected proficiency to be set")
	}

	// Test GetProficiencyBonus
	char.Proficiency = 0 // Auto-calculate
	bonus := char.GetProficiencyBonus()
	if bonus != 2 { // Level 1 character
		t.Errorf("expected proficiency bonus 2 for level 1, got %d", bonus)
	}

	// Test SpeedDetail
	speedDetail := char.GetDetailedSpeed()
	if speedDetail == nil {
		t.Fatal("expected speed detail to be created")
	}
	speedDetail.Fly = 60
	if char.SpeedDetail.Fly != 60 {
		t.Error("expected fly speed to be set")
	}

	// Test DeathSaves
	ds := char.GetDeathSaves()
	if ds == nil {
		t.Fatal("expected death saves to be created")
	}
	ds.AddSuccess()
	if char.DeathSaves.Successes != 1 {
		t.Error("expected 1 success")
	}

	// Test Currency
	currency := char.GetCurrency()
	if currency == nil {
		t.Fatal("expected currency to be created")
	}
	currency.GP = 100
	if char.Currency.GP != 100 {
		t.Error("expected 100gp")
	}

	// Test EquipmentSlots
	eqSlots := char.GetEquipmentSlots()
	if eqSlots == nil {
		t.Fatal("expected equipment slots to be created")
	}

	// Test Spellbook
	spellbook := char.GetSpellbook()
	if spellbook == nil {
		t.Fatal("expected spellbook to be created")
	}

	// Test Traits
	traits := char.GetTraits()
	if traits == nil {
		t.Fatal("expected traits to be created")
	}
	traits.AddLanguage("Common")
	if !char.Traits.HasLanguage("Common") {
		t.Error("expected Common language")
	}

	// Test Biography
	bio := char.GetBiography()
	if bio == nil {
		t.Fatal("expected biography to be created")
	}
	bio.Age = "25"
	if char.Biography.Age != "25" {
		t.Error("expected age to be set")
	}

	// Test Features
	feature := &models.Feature{
		ID:          "feat-001",
		Name:        "Great Weapon Master",
		Type:        models.FeatureTypeFeat,
		Description: "You've learned to put the weight of a weapon to your advantage...",
	}
	char.AddFeature(feature)
	if len(char.Features) != 1 {
		t.Error("expected 1 feature")
	}
	if char.GetFeature("feat-001") == nil {
		t.Error("expected to find feature")
	}
	char.RemoveFeature("feat-001")
	if len(char.Features) != 0 {
		t.Error("expected 0 features after removal")
	}

	// Test InventoryItems
	item := &models.InventoryItem{
		ID:       "item-001",
		Name:     "Health Potion",
		Quantity: 5,
	}
	char.AddInventoryItem(item)
	if len(char.InventoryItems) != 1 {
		t.Error("expected 1 inventory item")
	}
	if char.GetInventoryItem("item-001") == nil {
		t.Error("expected to find inventory item")
	}
	char.RemoveInventoryItem("item-001")
	if len(char.InventoryItems) != 0 {
		t.Error("expected 0 inventory items after removal")
	}

	// Test SkillsDetail
	char.SetSkillDetail("stealth", &models.Skill{Ability: "dexterity", Proficient: true})
	if char.GetSkillDetail("stealth") == nil {
		t.Error("expected to find skill detail")
	}

	// Test SavesDetail
	char.SetSaveDetail("constitution", &models.Save{Proficient: true})
	if char.GetSaveDetail("constitution") == nil {
		t.Error("expected to find save detail")
	}

	// Test Currency operations
	char.AddCurrency(&models.Currency{GP: 50})
	if char.Currency.GP != 150 {
		t.Errorf("expected 150gp after adding 50, got %d", char.Currency.GP)
	}

	success := char.SubtractCurrency(&models.Currency{GP: 30})
	if !success {
		t.Errorf("expected 120gp after subtracting 30, got %d", char.Currency.GP)
	}

	// Test Currency Subtract insufficient
	success = char.SubtractCurrency(&models.Currency{GP: 200})
	if success {
		t.Error("expected subtract to fail for insufficient funds")
	}

	// Test ImportMeta
	char.SetImportMeta(&models.ImportMeta{
		Format:     "fvtt",
		OriginalID: "actor-001",
		ImportedAt: time.Now(),
	})
	if !char.IsImported() {
		t.Error("expected character to be imported")
	}
}
