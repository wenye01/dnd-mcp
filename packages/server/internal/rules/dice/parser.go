// Package dice provides dice rolling and formula parsing functionality
package dice

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/dnd-mcp/server/internal/models"
)

// Parse errors
var (
	ErrEmptyFormula     = errors.New("empty formula")
	ErrInvalidFormat    = errors.New("invalid dice formula format")
	ErrInvalidDiceCount = errors.New("invalid dice count (must be 1-100)")
	ErrInvalidDiceSides = errors.New("invalid dice sides (must be 1-1000)")
	ErrInvalidModifier  = errors.New("invalid modifier value")
	ErrInvalidKeep      = errors.New("invalid keep value")
)

// Formula patterns
var (
	// Basic dice: NdS (e.g., 1d20, 2d6)
	// With modifier: NdS+M, NdS-M (e.g., 1d20+5, 2d6-1)
	// Keep high: NdSkhK (e.g., 4d6kh3)
	// Keep low: NdSklK (e.g., 2d20kl1)
	dicePattern = regexp.MustCompile(`^(\d+)?d(\d+)(kh|kl)?(\d+)?([+-]\d+)?$`)

	// Simple number (just a modifier, no dice)
	numberPattern = regexp.MustCompile(`^[+-]?\d+$`)
)

// ParseFormula parses a dice formula string into a DiceFormula struct
// Supported formats:
//   - Basic: NdS (e.g., "1d20", "2d6", "4d8")
//   - With modifier: NdS+M, NdS-M (e.g., "1d20+5", "2d6-1")
//   - Keep high: NdSkhK (e.g., "4d6kh3" - roll 4d6 keep highest 3)
//   - Keep low: NdSklK (e.g., "2d20kl1" - disadvantage)
//   - Simple number: just a modifier (e.g., "+5", "-2", "3")
func ParseFormula(formula string) (*models.DiceFormula, error) {
	formula = strings.TrimSpace(formula)
	formula = strings.ToLower(formula)

	if formula == "" {
		return nil, ErrEmptyFormula
	}

	// Check for simple number (just a modifier)
	if numberPattern.MatchString(formula) {
		modifier, err := strconv.Atoi(formula)
		if err != nil {
			return nil, ErrInvalidModifier
		}
		result := models.NewDiceFormula()
		result.Count = 0
		result.Sides = 0
		result.Modifier = modifier
		result.Original = formula
		return result, nil
	}

	// Match dice pattern
	matches := dicePattern.FindStringSubmatch(formula)
	if matches == nil {
		return nil, ErrInvalidFormat
	}

	result := models.NewDiceFormula()
	result.Original = formula

	// Parse dice count (default 1)
	if matches[1] != "" {
		count, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, ErrInvalidDiceCount
		}
		if count < 1 || count > 100 {
			return nil, ErrInvalidDiceCount
		}
		result.Count = count
	} else {
		result.Count = 1
	}

	// Parse dice sides
	sides, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, ErrInvalidDiceSides
	}
	if sides < 1 || sides > 1000 {
		return nil, ErrInvalidDiceSides
	}
	result.Sides = sides

	// Parse keep type (kh/kl)
	keepType := matches[3]
	keepValue := matches[4]

	if keepType != "" && keepValue != "" {
		keep, err := strconv.Atoi(keepValue)
		if err != nil {
			return nil, ErrInvalidKeep
		}
		if keep < 1 || keep > result.Count {
			return nil, ErrInvalidKeep
		}

		if keepType == "kh" {
			result.KeepHigh = keep
		} else if keepType == "kl" {
			result.KeepLow = keep
		}
	}

	// Parse modifier
	if matches[5] != "" {
		modifier, err := strconv.Atoi(matches[5])
		if err != nil {
			return nil, ErrInvalidModifier
		}
		result.Modifier = modifier
	}

	return result, nil
}

// MustParseFormula parses a dice formula string, panicking on error
func MustParseFormula(formula string) *models.DiceFormula {
	result, err := ParseFormula(formula)
	if err != nil {
		panic(err)
	}
	return result
}

// ValidateFormula validates a dice formula string without returning the parsed result
func ValidateFormula(formula string) error {
	_, err := ParseFormula(formula)
	return err
}

// NormalizeFormula normalizes a dice formula string
// e.g., "D20" -> "1d20", "d6" -> "1d6"
func NormalizeFormula(formula string) (string, error) {
	parsed, err := ParseFormula(formula)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

// ParseAdvantageDisadvantage parses advantage/disadvantage notation
// Returns the roll type and cleans the formula
// Advantage formats: "1d20 adv", "1d20 advantage", "2d20kh1"
// Disadvantage formats: "1d20 dis", "1d20 disadvantage", "2d20kl1"
func ParseAdvantageDisadvantage(formula string) (*models.DiceFormula, error) {
	formula = strings.TrimSpace(formula)
	formulaLower := strings.ToLower(formula)

	// Check for explicit advantage/disadvantage keywords
	isAdvantage := strings.HasSuffix(formulaLower, " adv") ||
		strings.HasSuffix(formulaLower, " advantage")
	isDisadvantage := strings.HasSuffix(formulaLower, " dis") ||
		strings.HasSuffix(formulaLower, " disadvantage")

	// Remove keywords from formula
	if isAdvantage {
		formula = strings.TrimSuffix(formula, " adv")
		formula = strings.TrimSuffix(formula, " Adv")
		formula = strings.TrimSuffix(formula, " ADV")
		formula = strings.TrimSuffix(formula, " advantage")
		formula = strings.TrimSuffix(formula, " Advantage")
		formula = strings.TrimSuffix(formula, " ADVANTAGE")
		formula = strings.TrimSpace(formula)
	} else if isDisadvantage {
		formula = strings.TrimSuffix(formula, " dis")
		formula = strings.TrimSuffix(formula, " Dis")
		formula = strings.TrimSuffix(formula, " DIS")
		formula = strings.TrimSuffix(formula, " disadvantage")
		formula = strings.TrimSuffix(formula, " Disadvantage")
		formula = strings.TrimSuffix(formula, " DISADVANTAGE")
		formula = strings.TrimSpace(formula)
	}

	parsed, err := ParseFormula(formula)
	if err != nil {
		return nil, err
	}

	// Check for implicit advantage/disadvantage in formula
	if parsed.KeepHigh == 1 && parsed.Count == 2 {
		parsed.RollType = models.RollTypeAdvantage
		parsed.KeepHigh = 0 // Clear keep, handled as advantage
	} else if parsed.KeepLow == 1 && parsed.Count == 2 {
		parsed.RollType = models.RollTypeDisadvantage
		parsed.KeepLow = 0 // Clear keep, handled as disadvantage
	} else if isAdvantage {
		parsed.RollType = models.RollTypeAdvantage
	} else if isDisadvantage {
		parsed.RollType = models.RollTypeDisadvantage
	}

	return parsed, nil
}

// SplitCompoundFormula splits a compound formula into individual formulas
// e.g., "1d20+5 + 2d6" -> ["1d20+5", "2d6"]
// This is useful for complex damage formulas
func SplitCompoundFormula(formula string) []string {
	formula = strings.TrimSpace(formula)
	if formula == "" {
		return nil
	}

	// Simple splitting by "+" while preserving modifiers
	// This is a simplified implementation
	parts := strings.Split(formula, "+")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// CombineFormulas combines multiple dice formulas into one
// e.g., ["1d20", "+5"] -> "1d20+5"
func CombineFormulas(formulas []string) string {
	if len(formulas) == 0 {
		return ""
	}

	result := ""
	for i, formula := range formulas {
		formula = strings.TrimSpace(formula)
		if formula == "" {
			continue
		}

		// Add separator for subsequent parts
		if i > 0 && !strings.HasPrefix(formula, "-") && !strings.HasPrefix(formula, "+") {
			result += "+"
		}
		result += formula
	}
	return result
}
