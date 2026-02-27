// Package service provides business logic layer implementations
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
	"github.com/dnd-mcp/server/internal/rules/dice"
)

// DiceService provides dice rolling and check functionality
// 规则参考: PHB 第7章 Ability Checks, 第9章 Combat
type DiceService struct {
	characterStore CharacterStore
	roller         *dice.Roller
}

// NewDiceService creates a new dice service
func NewDiceService(characterStore CharacterStore) *DiceService {
	return &DiceService{
		characterStore: characterStore,
		roller:         dice.NewRoller(),
	}
}

// NewDiceServiceWithRoller creates a new dice service with a custom roller (for testing)
func NewDiceServiceWithRoller(characterStore CharacterStore, roller *dice.Roller) *DiceService {
	return &DiceService{
		characterStore: characterStore,
		roller:         roller,
	}
}

// RollDiceRequest represents a dice roll request
type RollDiceRequest struct {
	Formula string `json:"formula"` // 骰子公式（如 "1d20+5", "2d6", "4d6kh3"）
}

// RollDiceResponse represents a dice roll response
type RollDiceResponse struct {
	Result *models.DiceResult `json:"result"` // 骰子结果
}

// RollDice rolls dice according to the given formula
// 规则参考: PHB 第7章 Ability Checks
func (s *DiceService) RollDice(ctx context.Context, req *RollDiceRequest) (*RollDiceResponse, error) {
	if req.Formula == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "formula is required")
	}

	// Parse formula with advantage/disadvantage support
	formula, err := dice.ParseAdvantageDisadvantage(req.Formula)
	if err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid formula: %v", err))
	}

	// Roll the dice
	result := s.roller.RollFormula(formula)

	return &RollDiceResponse{
		Result: result,
	}, nil
}

// RollCheckRequest represents an ability/skill check request
type RollCheckRequest struct {
	CharacterID string `json:"character_id"` // 角色ID
	Ability     string `json:"ability"`      // 属性（strength, dexterity, etc.）
	Skill       string `json:"skill"`        // 技能（可选，如 athletics, stealth）
	DC          int    `json:"dc"`           // 难度等级（可选）
	Advantage   bool   `json:"advantage"`    // 是否优势
	Disadvantage bool  `json:"disadvantage"` // 是否劣势
}

// RollCheckResponse represents an ability/skill check response
type RollCheckResponse struct {
	Result *models.CheckResult `json:"result"` // 检定结果
}

// RollCheck performs an ability or skill check
// 规则参考: PHB 第7章 Ability Checks
func (s *DiceService) RollCheck(ctx context.Context, req *RollCheckRequest) (*RollCheckResponse, error) {
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}
	if req.Ability == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "ability is required")
	}

	// Validate ability name
	ability := strings.ToLower(req.Ability)
	if !isValidAbility(ability) {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid ability: %s", req.Ability))
	}

	// Get character
	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Calculate modifier
	modifier := s.calculateCheckModifier(character, ability, req.Skill)

	// Roll d20 with advantage/disadvantage
	var diceResult *models.DiceResult
	if req.Advantage && !req.Disadvantage {
		diceResult = s.roller.RollWithAdvantage(modifier)
	} else if req.Disadvantage && !req.Advantage {
		diceResult = s.roller.RollWithDisadvantage(modifier)
	} else {
		// Normal roll or both advantage/disadvantage (they cancel out)
		diceResult = s.roller.RollD20(modifier)
	}

	// Create check result
	checkResult := models.NewCheckResult(diceResult, ability)
	if req.Skill != "" {
		checkResult.SetSkill(strings.ToLower(req.Skill))
	}
	if req.DC > 0 {
		checkResult.SetDC(req.DC)
	}

	return &RollCheckResponse{
		Result: checkResult,
	}, nil
}

// RollSaveRequest represents a saving throw request
type RollSaveRequest struct {
	CharacterID  string `json:"character_id"`  // 角色ID
	Ability      string `json:"ability"`       // 属性（strength, dexterity, etc.）
	DC           int    `json:"dc"`            // 难度等级（可选）
	Advantage    bool   `json:"advantage"`     // 是否优势
	Disadvantage bool   `json:"disadvantage"`  // 是否劣势
}

// RollSaveResponse represents a saving throw response
type RollSaveResponse struct {
	Result *models.CheckResult `json:"result"` // 检定结果
}

// RollSave performs a saving throw
// 规则参考: PHB 第7章 Saving Throws
func (s *DiceService) RollSave(ctx context.Context, req *RollSaveRequest) (*RollSaveResponse, error) {
	if req.CharacterID == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "character_id is required")
	}
	if req.Ability == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "ability is required")
	}

	// Validate ability name
	ability := strings.ToLower(req.Ability)
	if !isValidAbility(ability) {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid ability: %s", req.Ability))
	}

	// Get character
	character, err := s.characterStore.Get(ctx, req.CharacterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Calculate save modifier
	modifier := s.calculateSaveModifier(character, ability)

	// Roll d20 with advantage/disadvantage
	var diceResult *models.DiceResult
	if req.Advantage && !req.Disadvantage {
		diceResult = s.roller.RollWithAdvantage(modifier)
	} else if req.Disadvantage && !req.Advantage {
		diceResult = s.roller.RollWithDisadvantage(modifier)
	} else {
		// Normal roll or both advantage/disadvantage (they cancel out)
		diceResult = s.roller.RollD20(modifier)
	}

	// Create check result
	checkResult := models.NewCheckResult(diceResult, ability)
	if req.DC > 0 {
		checkResult.SetDC(req.DC)
	}

	return &RollSaveResponse{
		Result: checkResult,
	}, nil
}

// RollAttackRequest represents an attack roll request
type RollAttackRequest struct {
	CharacterID  string `json:"character_id"`  // 角色ID
	AttackBonus  int    `json:"attack_bonus"`  // 攻击加值
	TargetAC     int    `json:"target_ac"`     // 目标AC
	Advantage    bool   `json:"advantage"`     // 是否优势
	Disadvantage bool   `json:"disadvantage"`  // 是否劣势
}

// RollAttackResponse represents an attack roll response
type RollAttackResponse struct {
	AttackRoll *models.DiceResult `json:"attack_roll"` // 攻击骰
	TargetAC   int                `json:"target_ac"`   // 目标AC
	Hit        bool               `json:"hit"`         // 是否命中
	Crit       bool               `json:"crit"`        // 是否暴击
}

// RollAttack performs an attack roll
// 规则参考: PHB 第9章 Combat / Attack Rolls
func (s *DiceService) RollAttack(ctx context.Context, req *RollAttackRequest) (*RollAttackResponse, error) {
	if req.TargetAC < 0 {
		return nil, NewServiceError(ErrCodeInvalidInput, "target_ac cannot be negative")
	}

	// Roll d20 with advantage/disadvantage
	var diceResult *models.DiceResult
	if req.Advantage && !req.Disadvantage {
		diceResult = s.roller.RollWithAdvantage(req.AttackBonus)
	} else if req.Disadvantage && !req.Advantage {
		diceResult = s.roller.RollWithDisadvantage(req.AttackBonus)
	} else {
		// Normal roll or both advantage/disadvantage (they cancel out)
		diceResult = s.roller.RollD20(req.AttackBonus)
	}

	response := &RollAttackResponse{
		AttackRoll: diceResult,
		TargetAC:   req.TargetAC,
	}

	// Determine hit/crit
	// 规则参考: PHB 第9章 Combat / Critical Hits
	if diceResult.IsFumble() {
		// Natural 1 always misses
		response.Hit = false
		response.Crit = false
	} else if diceResult.IsCritical() {
		// Natural 20 always hits and is a crit
		response.Hit = true
		response.Crit = true
	} else {
		// Normal hit check
		response.Hit = diceResult.Total >= req.TargetAC
		response.Crit = false
	}

	return response, nil
}

// RollDamageRequest represents a damage roll request
type RollDamageRequest struct {
	Formula  string `json:"formula"`   // 伤害公式（如 "1d8+3", "2d6"）
	Crit     bool   `json:"crit"`      // 是否暴击（骰子翻倍）
}

// RollDamageResponse represents a damage roll response
type RollDamageResponse struct {
	Result *models.DiceResult `json:"result"` // 伤害骰结果
}

// RollDamage rolls damage dice
// 规则参考: PHB 第9章 Combat / Damage Rolls
func (s *DiceService) RollDamage(ctx context.Context, req *RollDamageRequest) (*RollDamageResponse, error) {
	if req.Formula == "" {
		return nil, NewServiceError(ErrCodeInvalidInput, "formula is required")
	}

	// Parse formula
	formula, err := dice.ParseFormula(req.Formula)
	if err != nil {
		return nil, NewServiceError(ErrCodeInvalidInput, fmt.Sprintf("invalid formula: %v", err))
	}

	// Double dice count for critical hits (but not modifier)
	// 规则参考: PHB 第9章 Combat / Critical Hits
	if req.Crit {
		formula.Count *= 2
		formula.Original = fmt.Sprintf("%s (critical)", req.Formula)
	}

	// Roll the dice
	result := s.roller.RollFormula(formula)

	return &RollDamageResponse{
		Result: result,
	}, nil
}

// calculateCheckModifier calculates the total modifier for an ability/skill check
// 规则参考: PHB 第7章 Ability Checks
func (s *DiceService) calculateCheckModifier(character *models.Character, ability string, skill string) int {
	var modifier int

	// Get ability modifier
	if character.Abilities != nil {
		modifier = rules.GetModifierByName(character.Abilities, rules.AbilityName(ability))
	}

	// Add skill proficiency bonus if applicable
	if skill != "" {
		skillBonus := s.getSkillBonus(character, skill, modifier)
		if skillBonus != 0 {
			return skillBonus
		}
	}

	return modifier
}

// getSkillBonus gets the total skill bonus including proficiency
func (s *DiceService) getSkillBonus(character *models.Character, skill string, abilityMod int) int {
	skill = strings.ToLower(skill)

	// Check detailed skills first
	if character.SkillsDetail != nil {
		if skillDetail, ok := character.SkillsDetail[skill]; ok {
			proficiencyBonus := character.GetProficiencyBonus()
			return skillDetail.CalculateBonus(abilityMod, proficiencyBonus)
		}
	}

	// Fall back to simple skill bonus map
	if character.Skills != nil {
		if bonus, ok := character.Skills[skill]; ok {
			return bonus
		}
	}

	// No skill bonus, return just ability modifier
	return abilityMod
}

// calculateSaveModifier calculates the total modifier for a saving throw
// 规则参考: PHB 第7章 Saving Throws
func (s *DiceService) calculateSaveModifier(character *models.Character, ability string) int {
	var modifier int

	// Get ability modifier
	if character.Abilities != nil {
		modifier = rules.GetModifierByName(character.Abilities, rules.AbilityName(ability))
	}

	// Check detailed saves first
	if character.SavesDetail != nil {
		if saveDetail, ok := character.SavesDetail[ability]; ok {
			proficiencyBonus := character.GetProficiencyBonus()
			return saveDetail.CalculateBonus(modifier, proficiencyBonus)
		}
	}

	// Fall back to simple save bonus map
	if character.Saves != nil {
		if bonus, ok := character.Saves[ability]; ok {
			return bonus
		}
	}

	return modifier
}

// isValidAbility checks if the given string is a valid ability name
func isValidAbility(ability string) bool {
	switch ability {
	case "strength", "dexterity", "constitution", "intelligence", "wisdom", "charisma":
		return true
	default:
		return false
	}
}

// GetSkillDCGuide returns common DC values for reference
// 规则参考: PHB 第8章 Ability Checks / Typical Difficulty Classes
func GetSkillDCGuide() map[string]int {
	return map[string]int{
		"very_easy":   5,
		"easy":        10,
		"medium":      15,
		"hard":        20,
		"very_hard":   25,
		"nearly_impossible": 30,
	}
}
