package rules_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
)

func TestAbilityModifier(t *testing.T) {
	tests := []struct {
		score    int
		expected int
	}{
		{1, -5},
		{2, -4},
		{3, -4},
		{4, -3},
		{5, -3},
		{6, -2},
		{7, -2},
		{8, -1},
		{9, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{13, 1},
		{14, 2},
		{15, 2},
		{16, 3},
		{17, 3},
		{18, 4},
		{19, 4},
		{20, 5},
		{21, 5},
		{22, 6},
		{23, 6},
		{24, 7},
		{25, 7},
		{30, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.AbilityModifier(tt.score)
			if result != tt.expected {
				t.Errorf("AbilityModifier(%d) = %d, expected %d", tt.score, result, tt.expected)
			}
		})
	}
}

func TestGetStrengthModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetStrengthModifier(abilities)
	expected := 3
	if result != expected {
		t.Errorf("GetStrengthModifier() = %d, expected %d", result, expected)
	}

	// Test nil abilities
	result = rules.GetStrengthModifier(nil)
	if result != 0 {
		t.Errorf("GetStrengthModifier(nil) = %d, expected 0", result)
	}
}

func TestGetDexterityModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetDexterityModifier(abilities)
	expected := 2
	if result != expected {
		t.Errorf("GetDexterityModifier() = %d, expected %d", result, expected)
	}

	// Test nil abilities
	result = rules.GetDexterityModifier(nil)
	if result != 0 {
		t.Errorf("GetDexterityModifier(nil) = %d, expected 0", result)
	}
}

func TestGetConstitutionModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetConstitutionModifier(abilities)
	expected := 1
	if result != expected {
		t.Errorf("GetConstitutionModifier() = %d, expected %d", result, expected)
	}
}

func TestGetIntelligenceModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetIntelligenceModifier(abilities)
	expected := 0
	if result != expected {
		t.Errorf("GetIntelligenceModifier() = %d, expected %d", result, expected)
	}
}

func TestGetWisdomModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetWisdomModifier(abilities)
	expected := -1
	if result != expected {
		t.Errorf("GetWisdomModifier() = %d, expected %d", result, expected)
	}
}

func TestGetCharismaModifier(t *testing.T) {
	abilities := &models.Abilities{Strength: 16, Dexterity: 14, Constitution: 12, Intelligence: 10, Wisdom: 8, Charisma: 15}
	result := rules.GetCharismaModifier(abilities)
	expected := 2
	if result != expected {
		t.Errorf("GetCharismaModifier() = %d, expected %d", result, expected)
	}
}

func TestGetModifierByName(t *testing.T) {
	abilities := &models.Abilities{Strength: 18, Dexterity: 14, Constitution: 16, Intelligence: 12, Wisdom: 10, Charisma: 8}

	tests := []struct {
		name     rules.AbilityName
		expected int
	}{
		{rules.AbilityStrength, 4},
		{rules.AbilityDexterity, 2},
		{rules.AbilityConstitution, 3},
		{rules.AbilityIntelligence, 1},
		{rules.AbilityWisdom, 0},
		{rules.AbilityCharisma, -1},
		{rules.AbilityName("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			result := rules.GetModifierByName(abilities, tt.name)
			if result != tt.expected {
				t.Errorf("GetModifierByName(%s) = %d, expected %d", tt.name, result, tt.expected)
			}
		})
	}

	// Test nil abilities
	result := rules.GetModifierByName(nil, rules.AbilityStrength)
	if result != 0 {
		t.Errorf("GetModifierByName(nil, ...) = %d, expected 0", result)
	}
}

func TestGetAbilityScoreByName(t *testing.T) {
	abilities := &models.Abilities{Strength: 18, Dexterity: 14, Constitution: 16, Intelligence: 12, Wisdom: 10, Charisma: 8}

	tests := []struct {
		name     rules.AbilityName
		expected int
	}{
		{rules.AbilityStrength, 18},
		{rules.AbilityDexterity, 14},
		{rules.AbilityConstitution, 16},
		{rules.AbilityIntelligence, 12},
		{rules.AbilityWisdom, 10},
		{rules.AbilityCharisma, 8},
		{rules.AbilityName("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			result := rules.GetAbilityScoreByName(abilities, tt.name)
			if result != tt.expected {
				t.Errorf("GetAbilityScoreByName(%s) = %d, expected %d", tt.name, result, tt.expected)
			}
		})
	}

	// Test nil abilities
	result := rules.GetAbilityScoreByName(nil, rules.AbilityStrength)
	if result != 0 {
		t.Errorf("GetAbilityScoreByName(nil, ...) = %d, expected 0", result)
	}
}

func TestModifierTable(t *testing.T) {
	table := rules.ModifierTable()

	// Verify some key values
	if table[10] != 0 {
		t.Errorf("ModifierTable()[10] = %d, expected 0", table[10])
	}
	if table[20] != 5 {
		t.Errorf("ModifierTable()[20] = %d, expected 5", table[20])
	}
	if table[1] != -5 {
		t.Errorf("ModifierTable()[1] = %d, expected -5", table[1])
	}
	if table[30] != 10 {
		t.Errorf("ModifierTable()[30] = %d, expected 10", table[30])
	}
}

func TestCalculatePassiveScore(t *testing.T) {
	tests := []struct {
		modifier int
		bonuses  []int
		expected int
	}{
		{0, nil, 10},
		{3, nil, 13},
		{3, []int{2}, 15},
		{3, []int{2, 1}, 16},
		{-1, nil, 9},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculatePassiveScore(tt.modifier, tt.bonuses...)
			if result != tt.expected {
				t.Errorf("CalculatePassiveScore(%d, %v) = %d, expected %d", tt.modifier, tt.bonuses, result, tt.expected)
			}
		})
	}
}

func TestCalculateBaseHP(t *testing.T) {
	tests := []struct {
		hitDieMax   int
		conModifier int
		expected    int
	}{
		{8, 3, 11},  // d8 + 3 CON
		{10, 2, 12}, // d10 + 2 CON
		{12, 4, 16}, // d12 + 4 CON
		{8, -2, 6},  // d8 - 2 CON
		{6, -5, 1},  // Minimum 1 HP
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateBaseHP(tt.hitDieMax, tt.conModifier)
			if result != tt.expected {
				t.Errorf("CalculateBaseHP(%d, %d) = %d, expected %d", tt.hitDieMax, tt.conModifier, result, tt.expected)
			}
		})
	}
}

func TestCalculateHPPerLevel(t *testing.T) {
	tests := []struct {
		hitDieMax   int
		conModifier int
		expected    int
	}{
		{8, 3, 7},   // (8+1)/2 + 3 = 4 + 3 = 7
		{10, 2, 7},  // (10+1)/2 + 2 = 5 + 2 = 7
		{12, 4, 10}, // (12+1)/2 + 4 = 6 + 4 = 10
		{8, -2, 2},  // (8+1)/2 - 2 = 4 - 2 = 2
		{6, -5, 1},  // Minimum 1 HP
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateHPPerLevel(tt.hitDieMax, tt.conModifier)
			if result != tt.expected {
				t.Errorf("CalculateHPPerLevel(%d, %d) = %d, expected %d", tt.hitDieMax, tt.conModifier, result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalHP(t *testing.T) {
	// Fighter with d10 hit die, +3 CON modifier, level 5
	// Level 1: 10 + 3 = 13
	// Level 2-5: (10+1)/2 + 3 = 5 + 3 = 8 each
	// Total: 13 + 8*4 = 45
	result := rules.CalculateTotalHP(10, 3, 5)
	expected := 45
	if result != expected {
		t.Errorf("CalculateTotalHP(10, 3, 5) = %d, expected %d", result, expected)
	}

	// Wizard with d6 hit die, +1 CON modifier, level 1
	// Level 1: 6 + 1 = 7
	result = rules.CalculateTotalHP(6, 1, 1)
	expected = 7
	if result != expected {
		t.Errorf("CalculateTotalHP(6, 1, 1) = %d, expected %d", result, expected)
	}

	// Level < 1 should default to 1
	result = rules.CalculateTotalHP(8, 2, 0)
	expected = 10
	if result != expected {
		t.Errorf("CalculateTotalHP(8, 2, 0) = %d, expected %d", result, expected)
	}
}

func TestCalculateBaseAC(t *testing.T) {
	tests := []struct {
		dexModifier int
		expected    int
	}{
		{0, 10},
		{3, 13},
		{-1, 9},
		{5, 15},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateBaseAC(tt.dexModifier)
			if result != tt.expected {
				t.Errorf("CalculateBaseAC(%d) = %d, expected %d", tt.dexModifier, result, tt.expected)
			}
		})
	}
}

func TestCalculateACWithArmor(t *testing.T) {
	tests := []struct {
		baseAC      int
		dexModifier int
		armorType   rules.ArmorType
		expected    int
	}{
		// Light armor - full Dex bonus
		{11, 3, rules.ArmorTypeLight, 14},
		{12, 5, rules.ArmorTypeLight, 17},

		// Medium armor - Dex bonus max +2
		{13, 3, rules.ArmorTypeMedium, 15},
		{14, 5, rules.ArmorTypeMedium, 16},
		{14, 1, rules.ArmorTypeMedium, 15},

		// Heavy armor - no Dex bonus
		{16, 3, rules.ArmorTypeHeavy, 16},
		{18, 5, rules.ArmorTypeHeavy, 18},
		{18, -1, rules.ArmorTypeHeavy, 18},

		// No armor - full Dex bonus (same as light)
		{10, 3, rules.ArmorTypeNone, 13},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateACWithArmor(tt.baseAC, tt.dexModifier, tt.armorType)
			if result != tt.expected {
				t.Errorf("CalculateACWithArmor(%d, %d, %s) = %d, expected %d",
					tt.baseAC, tt.dexModifier, tt.armorType, result, tt.expected)
			}
		})
	}
}

func TestCalculateInitiative(t *testing.T) {
	tests := []struct {
		dexModifier int
		bonuses     []int
		expected    int
	}{
		{0, nil, 0},
		{3, nil, 3},
		{3, []int{2}, 5},
		{3, []int{2, 1}, 6},
		{-1, nil, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateInitiative(tt.dexModifier, tt.bonuses...)
			if result != tt.expected {
				t.Errorf("CalculateInitiative(%d, %v) = %d, expected %d", tt.dexModifier, tt.bonuses, result, tt.expected)
			}
		})
	}
}

func TestSkillAbilityMapping(t *testing.T) {
	tests := []struct {
		skill    string
		expected rules.AbilityName
	}{
		{"athletics", rules.AbilityStrength},
		{"acrobatics", rules.AbilityDexterity},
		{"sleight_of_hand", rules.AbilityDexterity},
		{"stealth", rules.AbilityDexterity},
		{"arcana", rules.AbilityIntelligence},
		{"history", rules.AbilityIntelligence},
		{"investigation", rules.AbilityIntelligence},
		{"nature", rules.AbilityIntelligence},
		{"religion", rules.AbilityIntelligence},
		{"animal_handling", rules.AbilityWisdom},
		{"insight", rules.AbilityWisdom},
		{"medicine", rules.AbilityWisdom},
		{"perception", rules.AbilityWisdom},
		{"survival", rules.AbilityWisdom},
		{"deception", rules.AbilityCharisma},
		{"intimidation", rules.AbilityCharisma},
		{"performance", rules.AbilityCharisma},
		{"persuasion", rules.AbilityCharisma},
	}

	for _, tt := range tests {
		t.Run(tt.skill, func(t *testing.T) {
			result := rules.GetSkillAbility(tt.skill)
			if result != tt.expected {
				t.Errorf("GetSkillAbility(%s) = %s, expected %s", tt.skill, result, tt.expected)
			}
		})
	}

	// Unknown skill should return default
	result := rules.GetSkillAbility("unknown_skill")
	if result != rules.AbilityStrength {
		t.Errorf("GetSkillAbility(unknown_skill) = %s, expected %s", result, rules.AbilityStrength)
	}
}

func TestCalculateSkillBonus(t *testing.T) {
	tests := []struct {
		abilityModifier  int
		proficient       bool
		proficiencyBonus int
		bonuses          []int
		expected         int
	}{
		{3, false, 2, nil, 3},     // Not proficient
		{3, true, 2, nil, 5},      // Proficient
		{3, true, 2, []int{2}, 7}, // Proficient with bonus
		{-1, true, 2, nil, 1},     // Negative modifier, proficient
		{0, false, 2, nil, 0},     // Zero modifier, not proficient
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.CalculateSkillBonus(tt.abilityModifier, tt.proficient, tt.proficiencyBonus, tt.bonuses...)
			if result != tt.expected {
				t.Errorf("CalculateSkillBonus(%d, %v, %d, %v) = %d, expected %d",
					tt.abilityModifier, tt.proficient, tt.proficiencyBonus, tt.bonuses, result, tt.expected)
			}
		})
	}
}

func TestGetProficiencyBonus(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 2},
		{4, 2},
		{5, 3},
		{8, 3},
		{9, 4},
		{12, 4},
		{13, 5},
		{16, 5},
		{17, 6},
		{20, 6},
		{0, 2},  // < 1 defaults to 1
		{25, 6}, // > 20 defaults to 20
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := rules.GetProficiencyBonus(tt.level)
			if result != tt.expected {
				t.Errorf("GetProficiencyBonus(%d) = %d, expected %d", tt.level, result, tt.expected)
			}
		})
	}
}

func TestProficiencyBonusTable(t *testing.T) {
	table := rules.ProficiencyBonusTable()

	// Verify all levels 1-20 are present
	for i := 1; i <= 20; i++ {
		if _, ok := table[i]; !ok {
			t.Errorf("ProficiencyBonusTable() missing level %d", i)
		}
	}

	// Verify some key values
	if table[1] != 2 {
		t.Errorf("ProficiencyBonusTable()[1] = %d, expected 2", table[1])
	}
	if table[5] != 3 {
		t.Errorf("ProficiencyBonusTable()[5] = %d, expected 3", table[5])
	}
	if table[9] != 4 {
		t.Errorf("ProficiencyBonusTable()[9] = %d, expected 4", table[9])
	}
	if table[17] != 6 {
		t.Errorf("ProficiencyBonusTable()[17] = %d, expected 6", table[17])
	}
}

func TestSaveAbilityMapping(t *testing.T) {
	tests := []struct {
		save     string
		expected rules.AbilityName
	}{
		{"strength", rules.AbilityStrength},
		{"dexterity", rules.AbilityDexterity},
		{"constitution", rules.AbilityConstitution},
		{"intelligence", rules.AbilityIntelligence},
		{"wisdom", rules.AbilityWisdom},
		{"charisma", rules.AbilityCharisma},
	}

	for _, tt := range tests {
		t.Run(tt.save, func(t *testing.T) {
			result := rules.GetSaveAbility(tt.save)
			if result != tt.expected {
				t.Errorf("GetSaveAbility(%s) = %s, expected %s", tt.save, result, tt.expected)
			}
		})
	}

	// Unknown save should return default
	result := rules.GetSaveAbility("unknown_save")
	if result != rules.AbilityStrength {
		t.Errorf("GetSaveAbility(unknown_save) = %s, expected %s", result, rules.AbilityStrength)
	}
}
