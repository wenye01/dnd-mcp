// Package rules provides D&D 5e game rule implementations
package rules

import (
	"github.com/dnd-mcp/server/internal/models"
)

// AbilityModifier calculates the D&D 5e ability modifier
// Formula: floor((score - 10) / 2)
func AbilityModifier(score int) int {
	// Go's integer division truncates toward zero, but we need floor division
	// For negative results, we need to adjust
	result := (score - 10) / 2
	// If score-10 is negative and odd, we need to round down further
	if (score-10) < 0 && (score-10)%2 != 0 {
		result--
	}
	return result
}

// GetStrengthModifier returns the Strength modifier
func GetStrengthModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Strength)
}

// GetDexterityModifier returns the Dexterity modifier
func GetDexterityModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Dexterity)
}

// GetConstitutionModifier returns the Constitution modifier
func GetConstitutionModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Constitution)
}

// GetIntelligenceModifier returns the Intelligence modifier
func GetIntelligenceModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Intelligence)
}

// GetWisdomModifier returns the Wisdom modifier
func GetWisdomModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Wisdom)
}

// GetCharismaModifier returns the Charisma modifier
func GetCharismaModifier(a *models.Abilities) int {
	if a == nil {
		return 0
	}
	return AbilityModifier(a.Charisma)
}

// AbilityName represents ability name constants
type AbilityName string

const (
	AbilityStrength     AbilityName = "strength"
	AbilityDexterity    AbilityName = "dexterity"
	AbilityConstitution AbilityName = "constitution"
	AbilityIntelligence AbilityName = "intelligence"
	AbilityWisdom       AbilityName = "wisdom"
	AbilityCharisma     AbilityName = "charisma"
)

// GetModifierByName returns the modifier for the given ability name
func GetModifierByName(a *models.Abilities, name AbilityName) int {
	if a == nil {
		return 0
	}

	switch name {
	case AbilityStrength:
		return GetStrengthModifier(a)
	case AbilityDexterity:
		return GetDexterityModifier(a)
	case AbilityConstitution:
		return GetConstitutionModifier(a)
	case AbilityIntelligence:
		return GetIntelligenceModifier(a)
	case AbilityWisdom:
		return GetWisdomModifier(a)
	case AbilityCharisma:
		return GetCharismaModifier(a)
	default:
		return 0
	}
}

// GetAbilityScoreByName returns the score for the given ability name
func GetAbilityScoreByName(a *models.Abilities, name AbilityName) int {
	if a == nil {
		return 0
	}

	switch name {
	case AbilityStrength:
		return a.Strength
	case AbilityDexterity:
		return a.Dexterity
	case AbilityConstitution:
		return a.Constitution
	case AbilityIntelligence:
		return a.Intelligence
	case AbilityWisdom:
		return a.Wisdom
	case AbilityCharisma:
		return a.Charisma
	default:
		return 0
	}
}

// ModifierTable returns a lookup table of ability score to modifier
func ModifierTable() map[int]int {
	return map[int]int{
		1:  -5,
		2:  -4,
		3:  -4,
		4:  -3,
		5:  -3,
		6:  -2,
		7:  -2,
		8:  -1,
		9:  -1,
		10: 0,
		11: 0,
		12: 1,
		13: 1,
		14: 2,
		15: 2,
		16: 3,
		17: 3,
		18: 4,
		19: 4,
		20: 5,
		21: 5,
		22: 6,
		23: 6,
		24: 7,
		25: 7,
		26: 8,
		27: 8,
		28: 9,
		29: 9,
		30: 10,
	}
}

// CalculatePassiveScore calculates passive perception or other passive scores
// Formula: 10 + modifier + bonuses
func CalculatePassiveScore(modifier int, bonuses ...int) int {
	total := 10 + modifier
	for _, b := range bonuses {
		total += b
	}
	return total
}

// CalculateBaseHP calculates base HP at level 1
// Formula: Hit Die max + Constitution modifier
func CalculateBaseHP(hitDieMax int, conModifier int) int {
	hp := hitDieMax + conModifier
	if hp < 1 {
		hp = 1 // Minimum 1 HP
	}
	return hp
}

// CalculateHPPerLevel calculates HP gained per level after level 1
// Formula: average of Hit Die (rounded up) + Constitution modifier
func CalculateHPPerLevel(hitDieMax int, conModifier int) int {
	average := (hitDieMax + 1) / 2 // Round up
	hp := average + conModifier
	if hp < 1 {
		hp = 1 // Minimum 1 HP per level
	}
	return hp
}

// CalculateTotalHP calculates total HP for a character
func CalculateTotalHP(hitDieMax int, conModifier int, level int) int {
	if level < 1 {
		level = 1
	}

	// Level 1: max hit die + con modifier
	total := CalculateBaseHP(hitDieMax, conModifier)

	// Additional levels: average hit die + con modifier
	for i := 2; i <= level; i++ {
		total += CalculateHPPerLevel(hitDieMax, conModifier)
	}

	return total
}

// CalculateBaseAC calculates base AC without armor
// Formula: 10 + Dexterity modifier
func CalculateBaseAC(dexModifier int) int {
	return 10 + dexModifier
}

// CalculateACWithArmor calculates AC with armor
// For light armor: AC + Dex modifier (no limit)
// For medium armor: AC + Dex modifier (max +2)
// For heavy armor: AC (no Dex bonus)
func CalculateACWithArmor(baseAC int, dexModifier int, armorType ArmorType) int {
	switch armorType {
	case ArmorTypeLight:
		return baseAC + dexModifier
	case ArmorTypeMedium:
		bonus := dexModifier
		if bonus > 2 {
			bonus = 2
		}
		return baseAC + bonus
	case ArmorTypeHeavy:
		return baseAC
	default:
		return baseAC + dexModifier
	}
}

// ArmorType represents armor type
type ArmorType string

const (
	ArmorTypeNone   ArmorType = "none"
	ArmorTypeLight  ArmorType = "light"
	ArmorTypeMedium ArmorType = "medium"
	ArmorTypeHeavy  ArmorType = "heavy"
)

// CalculateInitiative calculates initiative modifier
// Base: Dexterity modifier
func CalculateInitiative(dexModifier int, bonuses ...int) int {
	total := dexModifier
	for _, b := range bonuses {
		total += b
	}
	return total
}

// SkillAbilityMapping maps skills to their associated abilities
var SkillAbilityMapping = map[string]AbilityName{
	// Strength skills
	"athletics": AbilityStrength,

	// Dexterity skills
	"acrobatics":  AbilityDexterity,
	"sleight_of_hand": AbilityDexterity,
	"stealth":     AbilityDexterity,

	// Intelligence skills
	"arcana":     AbilityIntelligence,
	"history":    AbilityIntelligence,
	"investigation": AbilityIntelligence,
	"nature":     AbilityIntelligence,
	"religion":   AbilityIntelligence,

	// Wisdom skills
	"animal_handling": AbilityWisdom,
	"insight":     AbilityWisdom,
	"medicine":    AbilityWisdom,
	"perception":  AbilityWisdom,
	"survival":    AbilityWisdom,

	// Charisma skills
	"deception":  AbilityCharisma,
	"intimidation": AbilityCharisma,
	"performance": AbilityCharisma,
	"persuasion": AbilityCharisma,
}

// GetSkillAbility returns the ability associated with a skill
func GetSkillAbility(skill string) AbilityName {
	if ability, ok := SkillAbilityMapping[skill]; ok {
		return ability
	}
	return AbilityStrength // Default to Strength if unknown
}

// CalculateSkillBonus calculates total skill bonus
// Formula: ability modifier + proficiency bonus (if proficient) + bonuses
func CalculateSkillBonus(abilityModifier int, proficient bool, proficiencyBonus int, bonuses ...int) int {
	total := abilityModifier
	if proficient {
		total += proficiencyBonus
	}
	for _, b := range bonuses {
		total += b
	}
	return total
}

// ProficiencyBonusTable returns proficiency bonus by level
func ProficiencyBonusTable() map[int]int {
	return map[int]int{
		1:  2,
		2:  2,
		3:  2,
		4:  2,
		5:  3,
		6:  3,
		7:  3,
		8:  3,
		9:  4,
		10: 4,
		11: 4,
		12: 4,
		13: 5,
		14: 5,
		15: 5,
		16: 5,
		17: 6,
		18: 6,
		19: 6,
		20: 6,
	}
}

// GetProficiencyBonus returns proficiency bonus for a given level
func GetProficiencyBonus(level int) int {
	if level < 1 {
		level = 1
	}
	if level > 20 {
		level = 20
	}

	bonus := 2
	bonus += ((level - 1) / 4)
	return bonus
}

// SaveAbilityMapping maps saves to their abilities
var SaveAbilityMapping = map[string]AbilityName{
	"strength":     AbilityStrength,
	"dexterity":    AbilityDexterity,
	"constitution": AbilityConstitution,
	"intelligence": AbilityIntelligence,
	"wisdom":       AbilityWisdom,
	"charisma":     AbilityCharisma,
}

// GetSaveAbility returns the ability for a save type
func GetSaveAbility(save string) AbilityName {
	if ability, ok := SaveAbilityMapping[save]; ok {
		return ability
	}
	return AbilityStrength // Default
}
