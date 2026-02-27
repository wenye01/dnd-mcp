package dice_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules/dice"
	"github.com/stretchr/testify/assert"
)

// MockRandomSource is a mock implementation of RandomSource for testing
type MockRandomSource struct {
	values []int
	index  int
}

// Intn returns the next value from the mock
func (m *MockRandomSource) Intn(n int) int {
	if m.index >= len(m.values) {
		return 0
	}
	val := m.values[m.index] % n
	m.index++
	return val
}

// NewMockRandomSource creates a new mock random source
func NewMockRandomSource(values ...int) *MockRandomSource {
	return &MockRandomSource{
		values: values,
		index:  0,
	}
}

func TestRoller_Roll(t *testing.T) {
	// Test with mock that always returns 0 (result should be 1)
	mock := NewMockRandomSource(0)
	roller := dice.NewRollerWithSource(mock)

	result := roller.Roll(20)
	assert.Equal(t, 1, result, "Roll(20) with mock returning 0 should be 1")
}

func TestRoller_RollMultiple(t *testing.T) {
	// Mock returns 0, 1, 2 which should give 1, 2, 3
	mock := NewMockRandomSource(0, 1, 2)
	roller := dice.NewRollerWithSource(mock)

	rolls := roller.RollMultiple(3, 20)
	assert.Len(t, rolls, 3)
	assert.Equal(t, []int{1, 2, 3}, rolls)
}

func TestRoller_RollFormula_Basic(t *testing.T) {
	tests := []struct {
		name        string
		formula     string
		mockValues  []int
		expectTotal int
		expectCrit  models.CritStatus
	}{
		{
			name:        "1d20 natural 20",
			formula:     "1d20",
			mockValues:  []int{19}, // 19 % 20 = 19, + 1 = 20
			expectTotal: 20,
			expectCrit:  models.CritStatusSuccess,
		},
		{
			name:        "1d20 natural 1",
			formula:     "1d20",
			mockValues:  []int{0}, // 0 % 20 = 0, + 1 = 1
			expectTotal: 1,
			expectCrit:  models.CritStatusFail,
		},
		{
			name:        "1d20+5",
			formula:     "1d20+5",
			mockValues:  []int{9}, // 10
			expectTotal: 15,
			expectCrit:  models.CritStatusNone,
		},
		{
			name:        "2d6",
			formula:     "2d6",
			mockValues:  []int{2, 3}, // 3, 4
			expectTotal: 7,
			expectCrit:  models.CritStatusNone,
		},
		{
			name:        "modifier only",
			formula:     "+5",
			mockValues:  []int{},
			expectTotal: 5,
			expectCrit:  models.CritStatusNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRandomSource(tt.mockValues...)
			roller := dice.NewRollerWithSource(mock)

			formula, err := dice.ParseFormula(tt.formula)
			assert.NoError(t, err)

			result := roller.RollFormula(formula)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.expectCrit, result.CritStatus)
		})
	}
}

func TestRoller_RollFormula_KeepHighest(t *testing.T) {
	// 4d6kh3: roll 4d6 keep highest 3
	// Mock values 0,1,2,3 -> rolls 1,2,3,4 -> keep 3 highest (2,3,4) -> total 9
	mock := NewMockRandomSource(0, 1, 2, 3)
	roller := dice.NewRollerWithSource(mock)

	formula, err := dice.ParseFormula("4d6kh3")
	assert.NoError(t, err)

	result := roller.RollFormula(formula)
	assert.Equal(t, 9, result.Total)
	assert.Len(t, result.Rolls, 3)
}

func TestRoller_RollFormula_KeepLowest(t *testing.T) {
	// 4d6kl2: roll 4d6 keep lowest 2
	// Mock values 0,1,2,3 -> rolls 1,2,3,4 -> keep 2 lowest (1,2) -> total 3
	mock := NewMockRandomSource(0, 1, 2, 3)
	roller := dice.NewRollerWithSource(mock)

	formula, err := dice.ParseFormula("4d6kl2")
	assert.NoError(t, err)

	result := roller.RollFormula(formula)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Rolls, 2)
}

func TestRoller_RollAdvantage(t *testing.T) {
	tests := []struct {
		name        string
		mockValues  []int
		expectRoll  int
		expectCrit  models.CritStatus
	}{
		{
			name:       "advantage takes higher",
			mockValues: []int{4, 14}, // rolls 5 and 15, keep 15
			expectRoll: 15,
			expectCrit: models.CritStatusNone,
		},
		{
			name:       "advantage with natural 20 on first die",
			mockValues: []int{19, 10}, // rolls 20 and 11, keep 20, crit!
			expectRoll: 20,
			expectCrit: models.CritStatusSuccess,
		},
		{
			name:       "advantage with natural 20 on second die",
			mockValues: []int{10, 19}, // rolls 11 and 20, keep 20, crit!
			expectRoll: 20,
			expectCrit: models.CritStatusSuccess,
		},
		{
			name:       "advantage both 1s is fumble",
			mockValues: []int{0, 0}, // rolls 1 and 1, fumble!
			expectRoll: 1,
			expectCrit: models.CritStatusFail,
		},
		{
			name:       "advantage one 1 is not fumble",
			mockValues: []int{0, 10}, // rolls 1 and 11, keep 11, not fumble
			expectRoll: 11,
			expectCrit: models.CritStatusNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRandomSource(tt.mockValues...)
			roller := dice.NewRollerWithSource(mock)

			result := roller.RollWithAdvantage(0)
			assert.Equal(t, tt.expectRoll, result.Rolls[0])
			assert.Equal(t, tt.expectCrit, result.CritStatus)
		})
	}
}

func TestRoller_RollDisadvantage(t *testing.T) {
	tests := []struct {
		name        string
		mockValues  []int
		expectRoll  int
		expectCrit  models.CritStatus
	}{
		{
			name:       "disadvantage takes lower",
			mockValues: []int{14, 4}, // rolls 15 and 5, keep 5
			expectRoll: 5,
			expectCrit: models.CritStatusNone,
		},
		{
			name:       "disadvantage both 20s is crit",
			mockValues: []int{19, 19}, // rolls 20 and 20, crit!
			expectRoll: 20,
			expectCrit: models.CritStatusSuccess,
		},
		{
			name:       "disadvantage one 20 is not crit",
			mockValues: []int{19, 10}, // rolls 20 and 11, keep 11, not crit
			expectRoll: 11,
			expectCrit: models.CritStatusNone,
		},
		{
			name:       "disadvantage with natural 1 on first die",
			mockValues: []int{0, 10}, // rolls 1 and 11, keep 1, fumble!
			expectRoll: 1,
			expectCrit: models.CritStatusFail,
		},
		{
			name:       "disadvantage with natural 1 on second die",
			mockValues: []int{10, 0}, // rolls 11 and 1, keep 1, fumble!
			expectRoll: 1,
			expectCrit: models.CritStatusFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRandomSource(tt.mockValues...)
			roller := dice.NewRollerWithSource(mock)

			result := roller.RollWithDisadvantage(0)
			assert.Equal(t, tt.expectRoll, result.Rolls[0])
			assert.Equal(t, tt.expectCrit, result.CritStatus)
		})
	}
}

func TestRoller_RollD20(t *testing.T) {
	mock := NewMockRandomSource(9) // returns 10
	roller := dice.NewRollerWithSource(mock)

	result := roller.RollD20(5)
	assert.Equal(t, 10, result.Rolls[0])
	assert.Equal(t, 5, result.Modifier)
	assert.Equal(t, 15, result.Total)
}

func TestRoller_InvalidInputs(t *testing.T) {
	roller := dice.NewRoller()

	// Roll with 0 sides
	result := roller.Roll(0)
	assert.Equal(t, 0, result)

	// Roll with negative sides
	result = roller.Roll(-5)
	assert.Equal(t, 0, result)

	// RollMultiple with 0 count
	rolls := roller.RollMultiple(0, 20)
	assert.Nil(t, rolls)

	// RollMultiple with negative count
	rolls = roller.RollMultiple(-1, 20)
	assert.Nil(t, rolls)
}

func TestRoller_ModifierWithAdvantage(t *testing.T) {
	mock := NewMockRandomSource(4, 14) // rolls 5 and 15, keep 15
	roller := dice.NewRollerWithSource(mock)

	result := roller.RollWithAdvantage(3)
	assert.Equal(t, 15, result.Rolls[0])
	assert.Equal(t, 3, result.Modifier)
	assert.Equal(t, 18, result.Total)
}

func TestRoller_ModifierWithDisadvantage(t *testing.T) {
	mock := NewMockRandomSource(14, 4) // rolls 15 and 5, keep 5
	roller := dice.NewRollerWithSource(mock)

	result := roller.RollWithDisadvantage(-2)
	assert.Equal(t, 5, result.Rolls[0])
	assert.Equal(t, -2, result.Modifier)
	assert.Equal(t, 3, result.Total)
}

func TestDefaultRoller(t *testing.T) {
	// Test that default roller functions work
	result := dice.Roll(20)
	assert.GreaterOrEqual(t, result, 1)
	assert.LessOrEqual(t, result, 20)

	rolls := dice.RollMultiple(3, 6)
	assert.Len(t, rolls, 3)
	for _, r := range rolls {
		assert.GreaterOrEqual(t, r, 1)
		assert.LessOrEqual(t, r, 6)
	}
}

func TestNewSeededRandomSource(t *testing.T) {
	// Same seed should produce same sequence
	source1 := dice.NewSeededRandomSource(42)
	source2 := dice.NewSeededRandomSource(42)

	roller1 := dice.NewRollerWithSource(source1)
	roller2 := dice.NewRollerWithSource(source2)

	for i := 0; i < 10; i++ {
		r1 := roller1.Roll(20)
		r2 := roller2.Roll(20)
		assert.Equal(t, r1, r2, "Seeded rollers should produce same sequence")
	}
}
