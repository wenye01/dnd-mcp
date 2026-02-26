package models_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewDiceResult(t *testing.T) {
	result := models.NewDiceResult("1d20+5")

	if result.Formula != "1d20+5" {
		t.Errorf("Formula = %q, expected %q", result.Formula, "1d20+5")
	}
	if len(result.Rolls) != 0 {
		t.Errorf("Rolls should be empty initially, got %v", result.Rolls)
	}
	if result.Modifier != 0 {
		t.Errorf("Modifier = %d, expected 0", result.Modifier)
	}
	if result.Total != 0 {
		t.Errorf("Total = %d, expected 0", result.Total)
	}
	if result.CritStatus != models.CritStatusNone {
		t.Errorf("CritStatus = %s, expected none", result.CritStatus)
	}
}

func TestDiceResult_AddRoll(t *testing.T) {
	result := models.NewDiceResult("2d6")

	result.AddRoll(3)
	if len(result.Rolls) != 1 || result.Rolls[0] != 3 {
		t.Errorf("AddRoll(3) failed, Rolls = %v", result.Rolls)
	}
	if result.Total != 3 {
		t.Errorf("Total = %d, expected 3", result.Total)
	}

	result.AddRoll(4)
	if len(result.Rolls) != 2 {
		t.Errorf("AddRoll(4) failed, Rolls = %v", result.Rolls)
	}
	if result.Total != 7 {
		t.Errorf("Total = %d, expected 7", result.Total)
	}
}

func TestDiceResult_SetRolls(t *testing.T) {
	result := models.NewDiceResult("3d6")
	result.SetModifier(2)

	rolls := []int{4, 5, 6}
	result.SetRolls(rolls)

	if len(result.Rolls) != 3 {
		t.Errorf("SetRolls failed, Rolls = %v", result.Rolls)
	}
	// Total should be 4+5+6+2 = 17
	if result.Total != 17 {
		t.Errorf("Total = %d, expected 17", result.Total)
	}
}

func TestDiceResult_SetModifier(t *testing.T) {
	result := models.NewDiceResult("1d20+5")
	result.AddRoll(15)

	result.SetModifier(5)
	if result.Modifier != 5 {
		t.Errorf("Modifier = %d, expected 5", result.Modifier)
	}
	// Total should be 15+5 = 20
	if result.Total != 20 {
		t.Errorf("Total = %d, expected 20", result.Total)
	}

	result.SetModifier(-2)
	if result.Modifier != -2 {
		t.Errorf("Modifier = %d, expected -2", result.Modifier)
	}
	// Total should be 15-2 = 13
	if result.Total != 13 {
		t.Errorf("Total = %d, expected 13", result.Total)
	}
}

func TestDiceResult_CheckCritStatus(t *testing.T) {
	tests := []struct {
		roll     int
		expected models.CritStatus
	}{
		{1, models.CritStatusFail},
		{20, models.CritStatusSuccess},
		{10, models.CritStatusNone},
		{15, models.CritStatusNone},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := models.NewDiceResult("1d20")
			result.AddRoll(tt.roll)
			result.CheckCritStatus()

			if result.CritStatus != tt.expected {
				t.Errorf("CheckCritStatus for roll %d = %s, expected %s",
					tt.roll, result.CritStatus, tt.expected)
			}
		})
	}
}

func TestDiceResult_CheckCritStatus_MultipleRolls(t *testing.T) {
	// Multiple rolls should not trigger crit status (only for single d20)
	result := models.NewDiceResult("2d6")
	result.AddRoll(20)
	result.AddRoll(20)
	result.CheckCritStatus()

	if result.CritStatus != models.CritStatusNone {
		t.Errorf("CheckCritStatus for multiple rolls should be none, got %s", result.CritStatus)
	}
}

func TestDiceResult_IsCritical(t *testing.T) {
	result := models.NewDiceResult("1d20")
	result.AddRoll(20)
	result.CheckCritStatus()

	if !result.IsCritical() {
		t.Error("IsCritical() should return true for natural 20")
	}

	result2 := models.NewDiceResult("1d20")
	result2.AddRoll(15)
	result2.CheckCritStatus()

	if result2.IsCritical() {
		t.Error("IsCritical() should return false for non-20 roll")
	}
}

func TestDiceResult_IsFumble(t *testing.T) {
	result := models.NewDiceResult("1d20")
	result.AddRoll(1)
	result.CheckCritStatus()

	if !result.IsFumble() {
		t.Error("IsFumble() should return true for natural 1")
	}

	result2 := models.NewDiceResult("1d20")
	result2.AddRoll(15)
	result2.CheckCritStatus()

	if result2.IsFumble() {
		t.Error("IsFumble() should return false for non-1 roll")
	}
}

func TestNewCheckResult(t *testing.T) {
	diceResult := models.NewDiceResult("1d20")
	diceResult.AddRoll(15)
	diceResult.SetModifier(3)

	checkResult := models.NewCheckResult(diceResult, "strength")

	if checkResult.DiceResult != diceResult {
		t.Error("DiceResult not set correctly")
	}
	if checkResult.Ability != "strength" {
		t.Errorf("Ability = %q, expected %q", checkResult.Ability, "strength")
	}
	if checkResult.Skill != "" {
		t.Errorf("Skill should be empty, got %q", checkResult.Skill)
	}
	if checkResult.DC != 0 {
		t.Errorf("DC should be 0, got %d", checkResult.DC)
	}
}

func TestCheckResult_SetSkill(t *testing.T) {
	diceResult := models.NewDiceResult("1d20")
	checkResult := models.NewCheckResult(diceResult, "dexterity")

	checkResult.SetSkill("stealth")

	if checkResult.Skill != "stealth" {
		t.Errorf("Skill = %q, expected %q", checkResult.Skill, "stealth")
	}
}

func TestCheckResult_SetDC(t *testing.T) {
	diceResult := models.NewDiceResult("1d20")
	diceResult.AddRoll(15)
	diceResult.SetModifier(3) // Total = 18

	checkResult := models.NewCheckResult(diceResult, "strength")
	checkResult.SetDC(15)

	if checkResult.DC != 15 {
		t.Errorf("DC = %d, expected 15", checkResult.DC)
	}
	if !checkResult.Success {
		t.Error("Success should be true (18 >= 15)")
	}
	if checkResult.Margin != 3 {
		t.Errorf("Margin = %d, expected 3", checkResult.Margin)
	}
}

func TestCheckResult_Evaluate(t *testing.T) {
	tests := []struct {
		roll       int
		modifier   int
		dc         int
		expectOK   bool
		expectMarg int
	}{
		{15, 3, 15, true, 3},   // 18 vs DC 15 -> success by 3
		{10, 2, 15, false, -3}, // 12 vs DC 15 -> failure by 3
		{15, 0, 15, true, 0},   // 15 vs DC 15 -> success by 0
		{20, 0, 25, false, -5}, // 20 vs DC 25 -> failure by 5
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			diceResult := models.NewDiceResult("1d20")
			diceResult.AddRoll(tt.roll)
			diceResult.SetModifier(tt.modifier)

			checkResult := models.NewCheckResult(diceResult, "strength")
			checkResult.SetDC(tt.dc)

			if checkResult.Success != tt.expectOK {
				t.Errorf("Success = %v, expected %v (total %d vs DC %d)",
					checkResult.Success, tt.expectOK, diceResult.Total, tt.dc)
			}
			if checkResult.Margin != tt.expectMarg {
				t.Errorf("Margin = %d, expected %d", checkResult.Margin, tt.expectMarg)
			}
		})
	}
}

func TestCheckResult_IsSuccess_IsFailure(t *testing.T) {
	diceResult := models.NewDiceResult("1d20")
	diceResult.AddRoll(15)
	diceResult.SetModifier(3)

	checkResult := models.NewCheckResult(diceResult, "strength")
	checkResult.SetDC(15)

	if !checkResult.IsSuccess() {
		t.Error("IsSuccess() should return true")
	}
	if checkResult.IsFailure() {
		t.Error("IsFailure() should return false")
	}

	// Failure case
	diceResult2 := models.NewDiceResult("1d20")
	diceResult2.AddRoll(10)
	diceResult2.SetModifier(0)

	checkResult2 := models.NewCheckResult(diceResult2, "strength")
	checkResult2.SetDC(15)

	if checkResult2.IsSuccess() {
		t.Error("IsSuccess() should return false")
	}
	if !checkResult2.IsFailure() {
		t.Error("IsFailure() should return true")
	}
}

func TestNewDiceFormula(t *testing.T) {
	formula := models.NewDiceFormula()

	if formula.Count != 1 {
		t.Errorf("Count = %d, expected 1", formula.Count)
	}
	if formula.Sides != 20 {
		t.Errorf("Sides = %d, expected 20", formula.Sides)
	}
	if formula.Modifier != 0 {
		t.Errorf("Modifier = %d, expected 0", formula.Modifier)
	}
	if formula.KeepHigh != 0 {
		t.Errorf("KeepHigh = %d, expected 0", formula.KeepHigh)
	}
	if formula.KeepLow != 0 {
		t.Errorf("KeepLow = %d, expected 0", formula.KeepLow)
	}
	if formula.RollType != models.RollTypeNormal {
		t.Errorf("RollType = %s, expected normal", formula.RollType)
	}
}

func TestDiceFormula_IsKeepRoll(t *testing.T) {
	tests := []struct {
		formula  *models.DiceFormula
		expected bool
	}{
		{&models.DiceFormula{KeepHigh: 3}, true},
		{&models.DiceFormula{KeepLow: 1}, true},
		{&models.DiceFormula{KeepHigh: 0, KeepLow: 0}, false},
	}

	for i, tt := range tests {
		if tt.formula.IsKeepRoll() != tt.expected {
			t.Errorf("Test %d: IsKeepRoll() = %v, expected %v", i, tt.formula.IsKeepRoll(), tt.expected)
		}
	}
}

func TestDiceFormula_RollTypeChecks(t *testing.T) {
	// Advantage
	advFormula := &models.DiceFormula{RollType: models.RollTypeAdvantage}
	if !advFormula.IsAdvantage() {
		t.Error("IsAdvantage() should return true")
	}
	if advFormula.IsDisadvantage() {
		t.Error("IsDisadvantage() should return false")
	}

	// Disadvantage
	disFormula := &models.DiceFormula{RollType: models.RollTypeDisadvantage}
	if disFormula.IsAdvantage() {
		t.Error("IsAdvantage() should return false")
	}
	if !disFormula.IsDisadvantage() {
		t.Error("IsDisadvantage() should return true")
	}

	// Normal
	normalFormula := &models.DiceFormula{RollType: models.RollTypeNormal}
	if normalFormula.IsAdvantage() {
		t.Error("IsAdvantage() should return false")
	}
	if normalFormula.IsDisadvantage() {
		t.Error("IsDisadvantage() should return false")
	}
}
