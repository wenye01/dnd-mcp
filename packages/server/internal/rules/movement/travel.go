// Package movement provides D&D 5e movement and travel rules
// 规则参考: PHB 第8章 - Adventuring / Movement
package movement

import (
	"fmt"
)

// TravelPace represents different travel paces
// 规则参考: PHB 第8章 - Travel Pace
type TravelPace string

const (
	// PaceFast Fast pace: 4 miles/hour, 30 miles/day
	// Disadvantage on Perception checks
	PaceFast TravelPace = "fast"
	// PaceNormal Normal pace: 3 miles/hour, 24 miles/day
	PaceNormal TravelPace = "normal"
	// PaceSlow Slow pace: 2 miles/hour, 12 miles/day
	// Can use Stealth while traveling
	PaceSlow TravelPace = "slow"
)

// TravelSpeed represents travel speed in miles per hour and per day
// 规则参考: PHB 第8章 Table: Travel Pace
type TravelSpeed struct {
	MilesPerHour int     `json:"miles_per_hour"`
	MilesPerDay  int     `json:"miles_per_day"`
	Effect       string  `json:"effect"` // Description of effects
}

// GetTravelSpeed returns the travel speed for a given pace
// 规则参考: PHB 第8章 - Travel Pace
func GetTravelSpeed(pace TravelPace) TravelSpeed {
	switch pace {
	case PaceFast:
		return TravelSpeed{
			MilesPerHour: 4,
			MilesPerDay:  30,
			Effect:       "Disadvantage on Wisdom (Perception) checks",
		}
	case PaceSlow:
		return TravelSpeed{
			MilesPerHour: 2,
			MilesPerDay:  12,
			Effect:       "Can use Stealth while traveling",
		}
	default:
		return TravelSpeed{
			MilesPerHour: 3,
			MilesPerDay:  24,
			Effect:       "No special effects",
		}
	}
}

// TravelTimeResult represents the result of travel time calculation
type TravelTimeResult struct {
	Distance     int     `json:"distance"`      // Distance in miles
	Hours        int     `json:"hours"`         // Travel time in hours
	Days         float64 `json:"days"`          // Travel time in days
	Pace         TravelPace `json:"pace"`       // Travel pace used
	Speed        TravelSpeed `json:"speed"`     // Speed details
	Description  string  `json:"description"`   // Human-readable description
}

// CalculateTravelTime calculates travel time for a given distance and pace
// 规则参考: PHB 第8章 - Travel Pace
func CalculateTravelTime(distanceInMiles int, pace TravelPace) TravelTimeResult {
	speed := GetTravelSpeed(pace)

	// Calculate hours (at least 1 hour if distance > 0)
	hours := distanceInMiles / speed.MilesPerHour
	if distanceInMiles > 0 && hours == 0 {
		hours = 1
	}

	// Calculate days
	days := float64(distanceInMiles) / float64(speed.MilesPerDay)

	return TravelTimeResult{
		Distance:    distanceInMiles,
		Hours:       hours,
		Days:        days,
		Pace:        pace,
		Speed:       speed,
		Description: fmt.Sprintf("%d miles at %s pace (%d mph) = %d hours (~%.1f days)",
			distanceInMiles, pace, speed.MilesPerHour, hours, days),
	}
}

// CalculateDistanceInGrids calculates the distance between two grid positions
// Uses Manhattan distance for overland travel (horizontal + vertical movement)
// For combat, straight-line distance should be used instead
func CalculateDistanceInGrids(fromX, fromY, toX, toY int) int {
	// Manhattan distance for overland travel
	dx := toX - fromX
	if dx < 0 {
		dx = -dx
	}
	dy := toY - fromY
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// CalculateStraightLineDistance calculates straight-line distance between two points
// Uses the Pythagorean theorem: sqrt((x2-x1)² + (y2-y1)²)
// This is used for range calculations in combat
func CalculateStraightLineDistance(fromX, fromY, toX, toY int) float64 {
	dx := float64(toX - fromX)
	dy := float64(toY - fromY)
	return sqrt(dx*dx + dy*dy)
}

// Simple sqrt implementation for distance calculation
func sqrt(x float64) float64 {
	// Newton-Raphson method
	z := 1.0
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// DifficultTerrainMultiplier represents the movement cost for difficult terrain
// 规则参考: PHB 第8章 - Difficult Terrain
// "Every foot of movement in difficult terrain costs 1 extra foot."
// In terms of grids, each square costs 2 squares of movement.
const DifficultTerrainMultiplier = 2

// CalculateTravelTimeWithTerrain calculates travel time considering difficult terrain
// 规则参考: PHB 第8章 - Difficult Terrain
func CalculateTravelTimeWithTerrain(distanceInMiles int, pace TravelPace, difficultTerrainMiles int) TravelTimeResult {
	// Calculate effective distance (difficult terrain counts double)
	effectiveDistance := distanceInMiles + (difficultTerrainMiles * DifficultTerrainMultiplier)

	speed := GetTravelSpeed(pace)

	// Calculate hours
	hours := effectiveDistance / speed.MilesPerHour
	if effectiveDistance > 0 && hours == 0 {
		hours = 1
	}

	// Calculate days
	days := float64(effectiveDistance) / float64(speed.MilesPerDay)

	effectDesc := speed.Effect
	if difficultTerrainMiles > 0 {
		effectDesc += fmt.Sprintf(", %d miles of difficult terrain", difficultTerrainMiles)
	}

	return TravelTimeResult{
		Distance:    distanceInMiles,
		Hours:       hours,
		Days:        days,
		Pace:        pace,
		Speed:       speed,
		Description: fmt.Sprintf("%d miles (%d effective) at %s pace = %d hours (~%.1f days)",
			distanceInMiles, effectiveDistance, pace, hours, days),
	}
}

// MovementSpeed represents a character's speed in feet per round
// 规则参考: PHB 第8章 - Movement and Position
type MovementSpeed struct {
	Walk       int `json:"walk"`       // Normal speed (e.g., 30 feet)
	Burrow     int `json:"burrow"`     // Burrowing speed
	Climb      int `json:"climb"`      // Climbing speed
	Fly        int `json:"fly"`        // Flying speed
	Swim       int `json:"swim"`       // Swimming speed
}

// GetSpeedsForCharacter returns typical speeds for a character
// 规则参考: PHB 第8章 - Speed
// Most races have 30 feet speed, some have 25 (dwarves) or 35 (wood elves)
func GetSpeedsForCharacter(baseSpeed int) MovementSpeed {
	if baseSpeed <= 0 {
		baseSpeed = 30 // Default speed
	}

	return MovementSpeed{
		Walk:   baseSpeed,
		Burrow: 0,
		Climb:  0, // Climbing requires Athletics check at half speed unless special ability
		Fly:    0,
		Swim:   0, // Swimming requires Athletics check at half speed unless special ability
	}
}

// GetMovementInGrids converts feet to number of grids (5-foot squares)
// 规则参考: PHB 第8章 - Movement and Position
// "In a normal grid, each 1-inch square represents 5 feet."
func GetMovementInGrids(speedInFeet int) int {
	return speedInFeet / 5
}

// CalculateMovementCost calculates movement cost for a path
// Returns total movement cost considering difficult terrain
// 规则参考: PHB 第8章 - Difficult Terrain
func CalculateMovementCost(normalSquares, difficultSquares int) int {
	// Each difficult terrain square costs double
	return normalSquares + (difficultSquares * DifficultTerrainMultiplier)
}
