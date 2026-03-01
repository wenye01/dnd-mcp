package movement_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/rules/movement"
	"github.com/stretchr/testify/assert"
)

func TestGetTravelSpeed(t *testing.T) {
	tests := []struct {
		name  string
		pace  movement.TravelPace
		expected movement.TravelSpeed
	}{
		{
			name: "fast pace",
			pace: movement.PaceFast,
			expected: movement.TravelSpeed{
				MilesPerHour: 4,
				MilesPerDay:  30,
				Effect:       "Disadvantage on Wisdom (Perception) checks",
			},
		},
		{
			name: "normal pace",
			pace: movement.PaceNormal,
			expected: movement.TravelSpeed{
				MilesPerHour: 3,
				MilesPerDay:  24,
				Effect:       "No special effects",
			},
		},
		{
			name: "slow pace",
			pace: movement.PaceSlow,
			expected: movement.TravelSpeed{
				MilesPerHour: 2,
				MilesPerDay:  12,
				Effect:       "Can use Stealth while traveling",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.GetTravelSpeed(tt.pace)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateTravelTime(t *testing.T) {
	tests := []struct {
		name            string
		distanceInMiles int
		pace            movement.TravelPace
		expectedHours   int
		expectedDays    float64
	}{
		{
			name:            "24 miles normal pace",
			distanceInMiles: 24,
			pace:            movement.PaceNormal,
			expectedHours:   8, // 24 / 3 = 8
			expectedDays:    1.0, // 24 / 24 = 1
		},
		{
			name:            "30 miles fast pace",
			distanceInMiles: 30,
			pace:            movement.PaceFast,
			expectedHours:   7, // 30 / 4 = 7.5, truncated
			expectedDays:    1.0, // 30 / 30 = 1
		},
		{
			name:            "12 miles slow pace",
			distanceInMiles: 12,
			pace:            movement.PaceSlow,
			expectedHours:   6, // 12 / 2 = 6
			expectedDays:    1.0, // 12 / 12 = 1
		},
		{
			name:            "1 mile normal pace (minimum 1 hour)",
			distanceInMiles: 1,
			pace:            movement.PaceNormal,
			expectedHours:   1, // Minimum 1 hour
			expectedDays:    1.0 / 24.0,
		},
		{
			name:            "0 miles",
			distanceInMiles: 0,
			pace:            movement.PaceNormal,
			expectedHours:   0,
			expectedDays:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.CalculateTravelTime(tt.distanceInMiles, tt.pace)
			assert.Equal(t, tt.distanceInMiles, result.Distance)
			assert.Equal(t, tt.expectedHours, result.Hours)
			assert.InDelta(t, tt.expectedDays, result.Days, 0.01)
			assert.Equal(t, tt.pace, result.Pace)
		})
	}
}

func TestCalculateDistanceInGrids(t *testing.T) {
	tests := []struct {
		name           string
		fromX, fromY   int
		toX, toY       int
		expectedDist   int
	}{
		{
			name:         "same position",
			fromX:        0,
			fromY:        0,
			toX:          0,
			toY:          0,
			expectedDist: 0,
		},
		{
			name:         "horizontal only",
			fromX:        0,
			fromY:        0,
			toX:          10,
			toY:          0,
			expectedDist: 10,
		},
		{
			name:         "vertical only",
			fromX:        0,
			fromY:        0,
			toX:          0,
			toY:          10,
			expectedDist: 10,
		},
		{
			name:         "diagonal (Manhattan distance)",
			fromX:        0,
			fromY:        0,
			toX:          5,
			toY:          5,
			expectedDist: 10, // 5 + 5
		},
		{
			name:         "L-shaped path",
			fromX:        0,
			fromY:        0,
			toX:          7,
			toY:          3,
			expectedDist: 10, // 7 + 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.CalculateDistanceInGrids(tt.fromX, tt.fromY, tt.toX, tt.toY)
			assert.Equal(t, tt.expectedDist, result)
		})
	}
}

func TestCalculateStraightLineDistance(t *testing.T) {
	tests := []struct {
		name           string
		fromX, fromY   int
		toX, toY       int
		expectedDist   float64
	}{
		{
			name:         "same position",
			fromX:        0,
			fromY:        0,
			toX:          0,
			toY:          0,
			expectedDist: 0,
		},
		{
			name:         "horizontal 10 units",
			fromX:        0,
			fromY:        0,
			toX:          10,
			toY:          0,
			expectedDist: 10,
		},
		{
			name:         "vertical 10 units",
			fromX:        0,
			fromY:        0,
			toX:          0,
			toY:          10,
			expectedDist: 10,
		},
		{
			name:         "diagonal 5,5 (Pythagorean)",
			fromX:        0,
			fromY:        0,
			toX:          5,
			toY:          5,
			expectedDist: 7.07, // sqrt(50) ≈ 7.07
		},
		{
			name:         "3-4-5 triangle",
			fromX:        0,
			fromY:        0,
			toX:          3,
			toY:          4,
			expectedDist: 5, // sqrt(9 + 16) = 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.CalculateStraightLineDistance(tt.fromX, tt.fromY, tt.toX, tt.toY)
			assert.InDelta(t, tt.expectedDist, result, 0.1)
		})
	}
}

func TestCalculateTravelTimeWithTerrain(t *testing.T) {
	tests := []struct {
		name                    string
		distanceInMiles         int
		pace                    movement.TravelPace
		difficultTerrainMiles   int
		expectedEffectiveDist   int
		expectedHours           int
	}{
		{
			name:                  "10 miles with no difficult terrain",
			distanceInMiles:       10,
			pace:                  movement.PaceNormal,
			difficultTerrainMiles: 0,
			expectedEffectiveDist: 10,
			expectedHours:         3, // 10 / 3
		},
		{
			name:                  "10 miles with 2 miles difficult terrain",
			distanceInMiles:       10,
			pace:                  movement.PaceNormal,
			difficultTerrainMiles: 2,
			expectedEffectiveDist: 14, // 10 + (2 * 2) = 14
			expectedHours:         4, // 14 / 3 = 4.67, truncated
		},
		{
			name:                  "5 miles all difficult terrain",
			distanceInMiles:       5,
			pace:                  movement.PaceFast,
			difficultTerrainMiles: 5,
			expectedEffectiveDist: 15, // 5 + (5 * 2) = 15
			expectedHours:         3, // 15 / 4 = 3.75, truncated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.CalculateTravelTimeWithTerrain(tt.distanceInMiles, tt.pace, tt.difficultTerrainMiles)
			assert.Equal(t, tt.distanceInMiles, result.Distance)
			assert.Equal(t, tt.expectedHours, result.Hours)
			assert.Equal(t, tt.pace, result.Pace)
		})
	}
}

func TestGetSpeedsForCharacter(t *testing.T) {
	tests := []struct {
		name          string
		baseSpeed     int
		expectedSpeed movement.MovementSpeed
	}{
		{
			name:      "standard human speed",
			baseSpeed: 30,
			expectedSpeed: movement.MovementSpeed{
				Walk: 30,
				Burrow: 0,
				Climb: 0,
				Fly: 0,
				Swim: 0,
			},
		},
		{
			name:      "dwarf speed",
			baseSpeed: 25,
			expectedSpeed: movement.MovementSpeed{
				Walk: 25,
				Burrow: 0,
				Climb: 0,
				Fly: 0,
				Swim: 0,
			},
		},
		{
			name:      "wood elf speed",
			baseSpeed: 35,
			expectedSpeed: movement.MovementSpeed{
				Walk: 35,
				Burrow: 0,
				Climb: 0,
				Fly: 0,
				Swim: 0,
			},
		},
		{
			name:      "zero or negative defaults to 30",
			baseSpeed: 0,
			expectedSpeed: movement.MovementSpeed{
				Walk: 30,
				Burrow: 0,
				Climb: 0,
				Fly: 0,
				Swim: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.GetSpeedsForCharacter(tt.baseSpeed)
			assert.Equal(t, tt.expectedSpeed, result)
		})
	}
}

func TestGetMovementInGrids(t *testing.T) {
	tests := []struct {
		name            string
		speedInFeet     int
		expectedGrids   int
	}{
		{
			name:          "30 feet speed",
			speedInFeet:   30,
			expectedGrids: 6, // 30 / 5 = 6
		},
		{
			name:          "25 feet speed",
			speedInFeet:   25,
			expectedGrids: 5, // 25 / 5 = 5
		},
		{
			name:          "35 feet speed",
			speedInFeet:   35,
			expectedGrids: 7, // 35 / 5 = 7
		},
		{
			name:          "40 feet speed (monk)",
			speedInFeet:   40,
			expectedGrids: 8, // 40 / 5 = 8
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.GetMovementInGrids(tt.speedInFeet)
			assert.Equal(t, tt.expectedGrids, result)
		})
	}
}

func TestCalculateMovementCost(t *testing.T) {
	tests := []struct {
		name               string
		normalSquares      int
		difficultSquares   int
		expectedCost       int
	}{
		{
			name:             "no difficult terrain",
			normalSquares:    10,
			difficultSquares: 0,
			expectedCost:     10,
		},
		{
			name:             "half difficult terrain",
			normalSquares:    5,
			difficultSquares: 5,
			expectedCost:     15, // 5 + (5 * 2)
		},
		{
			name:             "all difficult terrain",
			normalSquares:    0,
			difficultSquares: 6,
			expectedCost:     12, // 6 * 2
		},
		{
			name:             "mixed terrain",
			normalSquares:    3,
			difficultSquares: 4,
			expectedCost:     11, // 3 + (4 * 2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := movement.CalculateMovementCost(tt.normalSquares, tt.difficultSquares)
			assert.Equal(t, tt.expectedCost, result)
		})
	}
}

func TestDifficultTerrainMultiplier(t *testing.T) {
	// This test ensures the multiplier constant is correct
	// 规则参考: PHB 第8章 - Difficult Terrain
	// "Every foot of movement in difficult terrain costs 1 extra foot."
	assert.Equal(t, 2, movement.DifficultTerrainMultiplier,
		"Difficult terrain should cost 2x movement")
}
