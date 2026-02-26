package rules_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules"
)

func TestIsUnconscious(t *testing.T) {
	tests := []struct {
		name       string
		hp         *models.HP
		deathSaves *models.DeathSaves
		expected   bool
	}{
		{
			name:       "HP=0, no death saves - unconscious",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: nil,
			expected:   true,
		},
		{
			name:       "HP=0, death saves at 0 - unconscious",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 0, Failures: 0},
			expected:   true,
		},
		{
			name:       "HP=5 - not unconscious",
			hp:         &models.HP{Current: 5, Max: 10},
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "HP=0, stable (3 successes) - not unconscious",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 3, Failures: 0},
			expected:   false,
		},
		{
			name:       "HP=0, dead (3 failures) - not unconscious",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 0, Failures: 3},
			expected:   false,
		},
		{
			name:       "HP=-5 (overflow death) - not unconscious",
			hp:         &models.HP{Current: -5, Max: 10},
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "nil HP - not unconscious",
			hp:         nil,
			deathSaves: nil,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.IsUnconscious(tt.hp, tt.deathSaves)
			if result != tt.expected {
				t.Errorf("IsUnconscious() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsDead(t *testing.T) {
	tests := []struct {
		name       string
		hp         *models.HP
		deathSaves *models.DeathSaves
		expected   bool
	}{
		{
			name:       "HP=5 - not dead",
			hp:         &models.HP{Current: 5, Max: 10},
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "HP=0 - not dead",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "HP=0, 3 failures - dead",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 0, Failures: 3},
			expected:   true,
		},
		{
			name:       "HP=-10 (overflow = MaxHP) - dead",
			hp:         &models.HP{Current: -10, Max: 10},
			deathSaves: nil,
			expected:   true,
		},
		{
			name:       "HP=-15 (overflow > MaxHP) - dead",
			hp:         &models.HP{Current: -15, Max: 10},
			deathSaves: nil,
			expected:   true,
		},
		{
			name:       "HP=-5 (overflow < MaxHP) - not dead",
			hp:         &models.HP{Current: -5, Max: 10},
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "nil HP - not dead",
			hp:         nil,
			deathSaves: nil,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.IsDead(tt.hp, tt.deathSaves)
			if result != tt.expected {
				t.Errorf("IsDead() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsStable(t *testing.T) {
	tests := []struct {
		name       string
		deathSaves *models.DeathSaves
		expected   bool
	}{
		{
			name:       "nil death saves - not stable",
			deathSaves: nil,
			expected:   false,
		},
		{
			name:       "0 successes - not stable",
			deathSaves: &models.DeathSaves{Successes: 0, Failures: 0},
			expected:   false,
		},
		{
			name:       "1 success - not stable",
			deathSaves: &models.DeathSaves{Successes: 1, Failures: 0},
			expected:   false,
		},
		{
			name:       "2 successes - not stable",
			deathSaves: &models.DeathSaves{Successes: 2, Failures: 0},
			expected:   false,
		},
		{
			name:       "3 successes - stable",
			deathSaves: &models.DeathSaves{Successes: 3, Failures: 0},
			expected:   true,
		},
		{
			name:       "3 successes, 2 failures - stable",
			deathSaves: &models.DeathSaves{Successes: 3, Failures: 2},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.IsStable(tt.deathSaves)
			if result != tt.expected {
				t.Errorf("IsStable() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetHPState(t *testing.T) {
	tests := []struct {
		name       string
		hp         *models.HP
		deathSaves *models.DeathSaves
		expected   rules.HPState
	}{
		{
			name:       "HP=5 - normal",
			hp:         &models.HP{Current: 5, Max: 10},
			deathSaves: nil,
			expected:   rules.HPStateNormal,
		},
		{
			name:       "HP=0, no saves - unconscious",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: nil,
			expected:   rules.HPStateUnconscious,
		},
		{
			name:       "HP=0, 3 successes - stable",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 3, Failures: 0},
			expected:   rules.HPStateStable,
		},
		{
			name:       "HP=0, 3 failures - dead",
			hp:         &models.HP{Current: 0, Max: 10},
			deathSaves: &models.DeathSaves{Successes: 0, Failures: 3},
			expected:   rules.HPStateDead,
		},
		{
			name:       "HP=-10 (overflow = MaxHP) - dead",
			hp:         &models.HP{Current: -10, Max: 10},
			deathSaves: nil,
			expected:   rules.HPStateDead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.GetHPState(tt.hp, tt.deathSaves)
			if result != tt.expected {
				t.Errorf("GetHPState() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMakeDeathSave(t *testing.T) {
	tests := []struct {
		name           string
		deathSaves     *models.DeathSaves
		roll           int
		expectedResult rules.DeathSaveResult
		expectedDead   bool
		expectedStable bool
	}{
		{
			name:           "Natural 20 - critical success",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 0},
			roll:           20,
			expectedResult: rules.DeathSaveCriticalSuccess,
			expectedDead:   false,
			expectedStable: false,
		},
		{
			name:           "Natural 1 - critical failure (adds 2 failures)",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 0},
			roll:           1,
			expectedResult: rules.DeathSaveCriticalFailure,
			expectedDead:   false,
			expectedStable: false,
		},
		{
			name:           "Natural 1 with 1 existing failure - dead",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 1},
			roll:           1,
			expectedResult: rules.DeathSaveCriticalFailure,
			expectedDead:   true,
			expectedStable: false,
		},
		{
			name:           "Roll 10 - success",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 0},
			roll:           10,
			expectedResult: rules.DeathSaveSuccess,
			expectedDead:   false,
			expectedStable: false,
		},
		{
			name:           "Roll 15 - success",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 0},
			roll:           15,
			expectedResult: rules.DeathSaveSuccess,
			expectedDead:   false,
			expectedStable: false,
		},
		{
			name:           "Roll 10 with 2 successes - stable",
			deathSaves:     &models.DeathSaves{Successes: 2, Failures: 0},
			roll:           10,
			expectedResult: rules.DeathSaveSuccess,
			expectedDead:   false,
			expectedStable: true,
		},
		{
			name:           "Roll 9 - failure",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 0},
			roll:           9,
			expectedResult: rules.DeathSaveFailure,
			expectedDead:   false,
			expectedStable: false,
		},
		{
			name:           "Roll 5 with 2 failures - dead",
			deathSaves:     &models.DeathSaves{Successes: 0, Failures: 2},
			roll:           5,
			expectedResult: rules.DeathSaveFailure,
			expectedDead:   true,
			expectedStable: false,
		},
		{
			name:           "nil death saves",
			deathSaves:     nil,
			roll:           10,
			expectedResult: rules.DeathSaveFailure,
			expectedDead:   false,
			expectedStable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, stableOrDead := rules.MakeDeathSave(tt.deathSaves, tt.roll)

			if result != tt.expectedResult {
				t.Errorf("MakeDeathSave() result = %v, expected %v", result, tt.expectedResult)
			}

			if tt.expectedStable && !stableOrDead {
				t.Errorf("MakeDeathSave() expected stable to be true")
			}

			if tt.expectedDead && !stableOrDead {
				t.Errorf("MakeDeathSave() expected dead to be true")
			}
		})
	}
}

func TestTakeDamageWhileUnconscious(t *testing.T) {
	tests := []struct {
		name         string
		hp           *models.HP
		deathSaves   *models.DeathSaves
		damage       int
		isCrit       bool
		expectedDead bool
	}{
		{
			name:         "Normal damage adds 1 failure",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 0},
			damage:       5,
			isCrit:       false,
			expectedDead: false,
		},
		{
			name:         "Normal damage with 2 existing failures - dead",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 2},
			damage:       5,
			isCrit:       false,
			expectedDead: true,
		},
		{
			name:         "Crit damage adds 2 failures",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 0},
			damage:       5,
			isCrit:       true,
			expectedDead: false,
		},
		{
			name:         "Crit damage with 1 existing failure - dead",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 1},
			damage:       5,
			isCrit:       true,
			expectedDead: true,
		},
		{
			name:         "Damage >= MaxHP - instant death",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 0},
			damage:       10,
			isCrit:       false,
			expectedDead: true,
		},
		{
			name:         "Damage > MaxHP - instant death",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 0},
			damage:       15,
			isCrit:       false,
			expectedDead: true,
		},
		{
			name:         "nil HP",
			hp:           nil,
			deathSaves:   &models.DeathSaves{Successes: 0, Failures: 0},
			damage:       5,
			isCrit:       false,
			expectedDead: false,
		},
		{
			name:         "nil death saves",
			hp:           &models.HP{Current: 0, Max: 10},
			deathSaves:   nil,
			damage:       5,
			isCrit:       false,
			expectedDead: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.TakeDamageWhileUnconscious(tt.hp, tt.deathSaves, tt.damage, tt.isCrit)
			if result != tt.expectedDead {
				t.Errorf("TakeDamageWhileUnconscious() = %v, expected %v", result, tt.expectedDead)
			}
		})
	}
}

func TestHealFromUnconscious(t *testing.T) {
	tests := []struct {
		name           string
		hp             *models.HP
		deathSaves     *models.DeathSaves
		healing        int
		expectedHealed int
		expectedHP     int
	}{
		{
			name:           "Heal 5 HP",
			hp:             &models.HP{Current: 0, Max: 10},
			deathSaves:     &models.DeathSaves{Successes: 2, Failures: 1},
			healing:        5,
			expectedHealed: 5,
			expectedHP:     5,
		},
		{
			name:           "Heal to max HP",
			hp:             &models.HP{Current: 0, Max: 10},
			deathSaves:     &models.DeathSaves{Successes: 2, Failures: 1},
			healing:        15,
			expectedHealed: 10,
			expectedHP:     10,
		},
		{
			name:           "Zero healing",
			hp:             &models.HP{Current: 0, Max: 10},
			deathSaves:     &models.DeathSaves{Successes: 2, Failures: 1},
			healing:        0,
			expectedHealed: 0,
			expectedHP:     0,
		},
		{
			name:           "nil HP",
			hp:             nil,
			deathSaves:     &models.DeathSaves{Successes: 2, Failures: 1},
			healing:        5,
			expectedHealed: 0,
			expectedHP:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.HealFromUnconscious(tt.hp, tt.deathSaves, tt.healing)
			if result != tt.expectedHealed {
				t.Errorf("HealFromUnconscious() healed = %d, expected %d", result, tt.expectedHealed)
			}
			if tt.hp != nil && tt.hp.Current != tt.expectedHP {
				t.Errorf("HealFromUnconscious() HP = %d, expected %d", tt.hp.Current, tt.expectedHP)
			}
			// Check death saves reset
			if tt.healing > 0 && tt.hp != nil && tt.deathSaves != nil {
				if tt.deathSaves.Successes != 0 || tt.deathSaves.Failures != 0 {
					t.Errorf("HealFromUnconscious() death saves not reset")
				}
			}
		})
	}
}

func TestStabilize(t *testing.T) {
	tests := []struct {
		name                 string
		deathSaves           *models.DeathSaves
		expectedSuccesses    int
		expectedFailures     int
	}{
		{
			name:                 "Stabilize with existing saves",
			deathSaves:           &models.DeathSaves{Successes: 1, Failures: 2},
			expectedSuccesses:    3,
			expectedFailures:     0,
		},
		{
			name:                 "Stabilize with no saves",
			deathSaves:           &models.DeathSaves{Successes: 0, Failures: 0},
			expectedSuccesses:    3,
			expectedFailures:     0,
		},
		{
			name:                 "nil death saves",
			deathSaves:           nil,
			expectedSuccesses:    0,
			expectedFailures:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules.Stabilize(tt.deathSaves)
			if tt.deathSaves != nil {
				if tt.deathSaves.Successes != tt.expectedSuccesses {
					t.Errorf("Stabilize() successes = %d, expected %d", tt.deathSaves.Successes, tt.expectedSuccesses)
				}
				if tt.deathSaves.Failures != tt.expectedFailures {
					t.Errorf("Stabilize() failures = %d, expected %d", tt.deathSaves.Failures, tt.expectedFailures)
				}
			}
		})
	}
}

func TestResetDeathSaves(t *testing.T) {
	deathSaves := &models.DeathSaves{Successes: 2, Failures: 1}
	rules.ResetDeathSaves(deathSaves)

	if deathSaves.Successes != 0 {
		t.Errorf("ResetDeathSaves() successes = %d, expected 0", deathSaves.Successes)
	}
	if deathSaves.Failures != 0 {
		t.Errorf("ResetDeathSaves() failures = %d, expected 0", deathSaves.Failures)
	}

	// Test nil
	rules.ResetDeathSaves(nil) // Should not panic
}
