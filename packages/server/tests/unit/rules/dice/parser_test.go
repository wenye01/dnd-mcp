package dice_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules/dice"
)

func TestParseFormula_Basic(t *testing.T) {
	tests := []struct {
		formula     string
		expectCount int
		expectSides int
		expectMod   int
		expectKeepH int
		expectKeepL int
		expectErr   bool
	}{
		// Basic dice
		{"1d20", 1, 20, 0, 0, 0, false},
		{"2d6", 2, 6, 0, 0, 0, false},
		{"4d8", 4, 8, 0, 0, 0, false},
		{"d20", 1, 20, 0, 0, 0, false}, // default count
		{"D20", 1, 20, 0, 0, 0, false}, // case insensitive

		// With positive modifier
		{"1d20+5", 1, 20, 5, 0, 0, false},
		{"2d6+3", 2, 6, 3, 0, 0, false},
		{"d20+10", 1, 20, 10, 0, 0, false},

		// With negative modifier
		{"1d20-2", 1, 20, -2, 0, 0, false},
		{"2d6-1", 2, 6, -1, 0, 0, false},

		// Keep high
		{"4d6kh3", 4, 6, 0, 3, 0, false},
		{"2d20kh1", 2, 20, 0, 1, 0, false},

		// Keep low
		{"2d20kl1", 2, 20, 0, 0, 1, false},
		{"4d6kl2", 4, 6, 0, 0, 2, false},

		// Keep with modifier
		{"4d6kh3+2", 4, 6, 2, 3, 0, false},
		{"2d20kl1-1", 2, 20, -1, 0, 1, false},

		// Simple numbers (modifier only)
		{"+5", 0, 0, 5, 0, 0, false},
		{"-3", 0, 0, -3, 0, 0, false},
		{"10", 0, 0, 10, 0, 0, false},

		// Invalid formats
		{"", 0, 0, 0, 0, 0, true},        // empty
		{"invalid", 0, 0, 0, 0, 0, true}, // not a formula
		{"0d20", 0, 0, 0, 0, 0, true},    // zero dice
		{"101d20", 0, 0, 0, 0, 0, true},  // too many dice
		{"1d0", 0, 0, 0, 0, 0, true},     // zero sides
		{"1d1001", 0, 0, 0, 0, 0, true},  // too many sides
		{"4d6kh5", 0, 0, 0, 0, 0, true},  // keep more than rolled
		{"4d6kh0", 0, 0, 0, 0, 0, true},  // keep zero
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			result, err := dice.ParseFormula(tt.formula)

			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseFormula(%q) expected error, got nil", tt.formula)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFormula(%q) unexpected error: %v", tt.formula, err)
				return
			}

			if result.Count != tt.expectCount {
				t.Errorf("ParseFormula(%q) Count = %d, expected %d", tt.formula, result.Count, tt.expectCount)
			}
			if result.Sides != tt.expectSides {
				t.Errorf("ParseFormula(%q) Sides = %d, expected %d", tt.formula, result.Sides, tt.expectSides)
			}
			if result.Modifier != tt.expectMod {
				t.Errorf("ParseFormula(%q) Modifier = %d, expected %d", tt.formula, result.Modifier, tt.expectMod)
			}
			if result.KeepHigh != tt.expectKeepH {
				t.Errorf("ParseFormula(%q) KeepHigh = %d, expected %d", tt.formula, result.KeepHigh, tt.expectKeepH)
			}
			if result.KeepLow != tt.expectKeepL {
				t.Errorf("ParseFormula(%q) KeepLow = %d, expected %d", tt.formula, result.KeepLow, tt.expectKeepL)
			}
		})
	}
}

func TestParseAdvantageDisadvantage(t *testing.T) {
	tests := []struct {
		formula        string
		expectRollType models.RollType
		expectCount    int
		expectSides    int
	}{
		// Advantage keywords
		{"1d20 adv", models.RollTypeAdvantage, 1, 20},
		{"1d20 advantage", models.RollTypeAdvantage, 1, 20},
		{"d20 ADV", models.RollTypeAdvantage, 1, 20},

		// Disadvantage keywords
		{"1d20 dis", models.RollTypeDisadvantage, 1, 20},
		{"1d20 disadvantage", models.RollTypeDisadvantage, 1, 20},
		{"d20 DIS", models.RollTypeDisadvantage, 1, 20},

		// Implicit advantage (2d20kh1)
		{"2d20kh1", models.RollTypeAdvantage, 2, 20},

		// Implicit disadvantage (2d20kl1)
		{"2d20kl1", models.RollTypeDisadvantage, 2, 20},

		// Normal roll
		{"1d20", models.RollTypeNormal, 1, 20},
		{"2d6", models.RollTypeNormal, 2, 6},

		// With modifier
		{"1d20+5 adv", models.RollTypeAdvantage, 1, 20},
		{"1d20-2 dis", models.RollTypeDisadvantage, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			result, err := dice.ParseAdvantageDisadvantage(tt.formula)
			if err != nil {
				t.Errorf("ParseAdvantageDisadvantage(%q) unexpected error: %v", tt.formula, err)
				return
			}

			if result.RollType != tt.expectRollType {
				t.Errorf("ParseAdvantageDisadvantage(%q) RollType = %s, expected %s",
					tt.formula, result.RollType, tt.expectRollType)
			}
			if result.Count != tt.expectCount {
				t.Errorf("ParseAdvantageDisadvantage(%q) Count = %d, expected %d",
					tt.formula, result.Count, tt.expectCount)
			}
			if result.Sides != tt.expectSides {
				t.Errorf("ParseAdvantageDisadvantage(%q) Sides = %d, expected %d",
					tt.formula, result.Sides, tt.expectSides)
			}
		})
	}
}

func TestValidateFormula(t *testing.T) {
	tests := []struct {
		formula  string
		expectOK bool
	}{
		{"1d20", true},
		{"2d6+3", true},
		{"4d6kh3", true},
		{"", false},
		{"invalid", false},
		{"0d20", false},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			err := dice.ValidateFormula(tt.formula)
			if tt.expectOK && err != nil {
				t.Errorf("ValidateFormula(%q) expected OK, got error: %v", tt.formula, err)
			}
			if !tt.expectOK && err == nil {
				t.Errorf("ValidateFormula(%q) expected error, got nil", tt.formula)
			}
		})
	}
}

func TestDiceFormula_IsKeepRoll(t *testing.T) {
	tests := []struct {
		formula  string
		expected bool
	}{
		{"4d6kh3", true},
		{"2d20kl1", true},
		{"1d20", false},
		{"2d6+3", false},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			result, err := dice.ParseFormula(tt.formula)
			if err != nil {
				t.Fatalf("ParseFormula(%q) error: %v", tt.formula, err)
			}

			if result.IsKeepRoll() != tt.expected {
				t.Errorf("IsKeepRoll() = %v, expected %v", result.IsKeepRoll(), tt.expected)
			}
		})
	}
}

func TestDiceFormula_AdvantageDisadvantage(t *testing.T) {
	tests := []struct {
		formula     string
		isAdvantage bool
		isDisadv    bool
	}{
		{"1d20 adv", true, false},
		{"1d20 dis", false, true},
		{"1d20", false, false},
		{"2d20kh1", true, false},
		{"2d20kl1", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			result, err := dice.ParseAdvantageDisadvantage(tt.formula)
			if err != nil {
				t.Fatalf("ParseAdvantageDisadvantage(%q) error: %v", tt.formula, err)
			}

			if result.IsAdvantage() != tt.isAdvantage {
				t.Errorf("IsAdvantage() = %v, expected %v", result.IsAdvantage(), tt.isAdvantage)
			}
			if result.IsDisadvantage() != tt.isDisadv {
				t.Errorf("IsDisadvantage() = %v, expected %v", result.IsDisadvantage(), tt.isDisadv)
			}
		})
	}
}

func TestSplitCompoundFormula(t *testing.T) {
	tests := []struct {
		formula  string
		expected []string
	}{
		{"1d20+5", []string{"1d20", "5"}},
		{"1d20+2d6", []string{"1d20", "2d6"}},
		{"1d20 + 2d6 + 3", []string{"1d20", "2d6", "3"}},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			result := dice.SplitCompoundFormula(tt.formula)

			if len(result) != len(tt.expected) {
				t.Errorf("SplitCompoundFormula(%q) = %v, expected %v", tt.formula, result, tt.expected)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("SplitCompoundFormula(%q)[%d] = %q, expected %q", tt.formula, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestCombineFormulas(t *testing.T) {
	tests := []struct {
		formulas []string
		expected string
	}{
		{[]string{"1d20", "5"}, "1d20+5"},
		{[]string{"1d20", "-2"}, "1d20-2"},
		{[]string{"1d20", "+5"}, "1d20+5"},
		{[]string{}, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := dice.CombineFormulas(tt.formulas)
			if result != tt.expected {
				t.Errorf("CombineFormulas(%v) = %q, expected %q", tt.formulas, result, tt.expected)
			}
		})
	}
}

func TestMustParseFormula(t *testing.T) {
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParseFormula panicked unexpectedly: %v", r)
		}
	}()

	result := dice.MustParseFormula("1d20+5")
	if result.Count != 1 || result.Sides != 20 || result.Modifier != 5 {
		t.Errorf("MustParseFormula returned unexpected result: %+v", result)
	}
}

func TestMustParseFormula_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseFormula expected to panic for invalid formula")
		}
	}()

	dice.MustParseFormula("invalid")
}
