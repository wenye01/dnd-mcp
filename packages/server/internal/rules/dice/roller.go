// Package dice provides dice rolling and formula parsing functionality
package dice

import (
	"math/rand"
	"sort"
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

// RandomSource defines an interface for random number generation
// This allows for deterministic testing by injecting a controlled random source
type RandomSource interface {
	Intn(n int) int
}

// DefaultRandomSource is the default implementation using math/rand
type DefaultRandomSource struct {
	rng *rand.Rand
}

// NewDefaultRandomSource creates a new default random source seeded with current time
func NewDefaultRandomSource() *DefaultRandomSource {
	return &DefaultRandomSource{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewSeededRandomSource creates a new random source with a specific seed (for testing)
func NewSeededRandomSource(seed int64) *DefaultRandomSource {
	return &DefaultRandomSource{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Intn returns a random integer in [0, n)
func (r *DefaultRandomSource) Intn(n int) int {
	return r.rng.Intn(n)
}

// Roller handles dice rolling operations
// 规则参考: PHB 第7章 Ability Checks, 第9章 Combat
type Roller struct {
	random RandomSource
}

// NewRoller creates a new Roller with the default random source
func NewRoller() *Roller {
	return &Roller{
		random: NewDefaultRandomSource(),
	}
}

// NewRollerWithSource creates a new Roller with a custom random source (for testing)
func NewRollerWithSource(source RandomSource) *Roller {
	return &Roller{
		random: source,
	}
}

// Roll rolls a single die with the given number of sides
// Returns a value in the range [1, sides]
func (r *Roller) Roll(sides int) int {
	if sides < 1 {
		return 0
	}
	return r.random.Intn(sides) + 1
}

// RollMultiple rolls multiple dice and returns all individual results
func (r *Roller) RollMultiple(count, sides int) []int {
	if count < 1 || sides < 1 {
		return nil
	}

	rolls := make([]int, count)
	for i := 0; i < count; i++ {
		rolls[i] = r.Roll(sides)
	}
	return rolls
}

// RollFormula rolls dice according to the given formula
// 规则参考: PHB 第7章 Ability Checks
func (r *Roller) RollFormula(formula *models.DiceFormula) *models.DiceResult {
	result := models.NewDiceResult(formula.Original)

	// Handle simple modifier-only formulas (no dice)
	if formula.Count == 0 || formula.Sides == 0 {
		result.SetModifier(formula.Modifier)
		return result
	}

	// Handle advantage/disadvantage
	if formula.RollType == models.RollTypeAdvantage {
		return r.rollAdvantage(formula)
	} else if formula.RollType == models.RollTypeDisadvantage {
		return r.rollDisadvantage(formula)
	}

	// Roll dice
	rolls := r.RollMultiple(formula.Count, formula.Sides)

	// Handle keep highest/lowest
	if formula.KeepHigh > 0 {
		rolls = keepHighest(rolls, formula.KeepHigh)
	} else if formula.KeepLow > 0 {
		rolls = keepLowest(rolls, formula.KeepLow)
	}

	result.SetRolls(rolls)
	result.SetModifier(formula.Modifier)

	// Check for critical hit/fumble on d20 rolls
	if formula.Sides == 20 && formula.Count == 1 && formula.KeepHigh == 0 && formula.KeepLow == 0 {
		result.CheckCritStatus()
	}

	return result
}

// rollAdvantage rolls 2d20 and takes the higher result
// 规则参考: PHB 第7章 Advantage and Disadvantage
func (r *Roller) rollAdvantage(formula *models.DiceFormula) *models.DiceResult {
	result := models.NewDiceFormula()
	result.Count = 2
	result.Sides = 20
	result.Modifier = formula.Modifier
	result.RollType = models.RollTypeAdvantage
	result.Original = formula.Original

	// Roll 2d20
	rolls := r.RollMultiple(2, 20)

	// Create dice result with both rolls for transparency
	diceResult := models.NewDiceResult(formula.Original)
	diceResult.SetRolls(rolls)
	diceResult.SetModifier(formula.Modifier)

	// Keep the higher roll
	keptRolls := keepHighest(rolls, 1)
	diceResult.SetRolls(keptRolls)

	// Check crit status on the kept roll (the higher one)
	// Note: In advantage, a natural 20 on either die counts as a crit
	// But we use the higher roll for the result
	if rolls[0] == 20 || rolls[1] == 20 {
		diceResult.CritStatus = models.CritStatusSuccess
	} else if rolls[0] == 1 && rolls[1] == 1 {
		// Both must be 1 for a fumble with advantage
		diceResult.CritStatus = models.CritStatusFail
	}

	return diceResult
}

// rollDisadvantage rolls 2d20 and takes the lower result
// 规则参考: PHB 第7章 Advantage and Disadvantage
func (r *Roller) rollDisadvantage(formula *models.DiceFormula) *models.DiceResult {
	result := models.NewDiceFormula()
	result.Count = 2
	result.Sides = 20
	result.Modifier = formula.Modifier
	result.RollType = models.RollTypeDisadvantage
	result.Original = formula.Original

	// Roll 2d20
	rolls := r.RollMultiple(2, 20)

	// Create dice result with both rolls for transparency
	diceResult := models.NewDiceResult(formula.Original)
	diceResult.SetRolls(rolls)
	diceResult.SetModifier(formula.Modifier)

	// Keep the lower roll
	keptRolls := keepLowest(rolls, 1)
	diceResult.SetRolls(keptRolls)

	// Check crit status on the kept roll (the lower one)
	// Note: In disadvantage, a natural 20 on both dice counts as a crit
	// But a natural 1 on either die counts as a fumble
	if rolls[0] == 20 && rolls[1] == 20 {
		diceResult.CritStatus = models.CritStatusSuccess
	} else if rolls[0] == 1 || rolls[1] == 1 {
		// Either being 1 causes fumble with disadvantage
		diceResult.CritStatus = models.CritStatusFail
	}

	return diceResult
}

// RollWithAdvantage rolls a d20 with advantage (2d20, take highest)
// This is a convenience method for common use case
func (r *Roller) RollWithAdvantage(modifier int) *models.DiceResult {
	formula := models.NewDiceFormula()
	formula.Count = 1
	formula.Sides = 20
	formula.Modifier = modifier
	formula.RollType = models.RollTypeAdvantage
	formula.Original = "1d20 advantage"
	return r.RollFormula(formula)
}

// RollWithDisadvantage rolls a d20 with disadvantage (2d20, take lowest)
// This is a convenience method for common use case
func (r *Roller) RollWithDisadvantage(modifier int) *models.DiceResult {
	formula := models.NewDiceFormula()
	formula.Count = 1
	formula.Sides = 20
	formula.Modifier = modifier
	formula.RollType = models.RollTypeDisadvantage
	formula.Original = "1d20 disadvantage"
	return r.RollFormula(formula)
}

// RollD20 rolls a single d20 with optional modifier
func (r *Roller) RollD20(modifier int) *models.DiceResult {
	formula := models.NewDiceFormula()
	formula.Count = 1
	formula.Sides = 20
	formula.Modifier = modifier
	formula.Original = "1d20"
	return r.RollFormula(formula)
}

// keepHighest keeps the n highest rolls from the slice
func keepHighest(rolls []int, n int) []int {
	if n >= len(rolls) {
		return rolls
	}
	if n <= 0 {
		return nil
	}

	// Sort in descending order
	sorted := make([]int, len(rolls))
	copy(sorted, rolls)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))

	return sorted[:n]
}

// keepLowest keeps the n lowest rolls from the slice
func keepLowest(rolls []int, n int) []int {
	if n >= len(rolls) {
		return rolls
	}
	if n <= 0 {
		return nil
	}

	// Sort in ascending order
	sorted := make([]int, len(rolls))
	copy(sorted, rolls)
	sort.Ints(sorted)

	return sorted[:n]
}

// Global roller instance for convenience functions
var defaultRoller = NewRoller()

// Roll rolls a single die using the default roller
func Roll(sides int) int {
	return defaultRoller.Roll(sides)
}

// RollMultiple rolls multiple dice using the default roller
func RollMultiple(count, sides int) []int {
	return defaultRoller.RollMultiple(count, sides)
}

// RollFormula rolls a formula using the default roller
func RollFormula(formula *models.DiceFormula) *models.DiceResult {
	return defaultRoller.RollFormula(formula)
}

// RollD20 rolls a d20 using the default roller
func RollD20(modifier int) *models.DiceResult {
	return defaultRoller.RollD20(modifier)
}

// RollWithAdvantage rolls with advantage using the default roller
func RollWithAdvantage(modifier int) *models.DiceResult {
	return defaultRoller.RollWithAdvantage(modifier)
}

// RollWithDisadvantage rolls with disadvantage using the default roller
func RollWithDisadvantage(modifier int) *models.DiceResult {
	return defaultRoller.RollWithDisadvantage(modifier)
}
