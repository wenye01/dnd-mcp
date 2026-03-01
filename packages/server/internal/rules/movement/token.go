// Package movement provides D&D 5e movement rules for combat tokens
// 规则参考: PHB 第9章 - Movement and Position
package movement

import (
	"github.com/dnd-mcp/server/internal/models"
)

// TokenMovementResult represents the result of a token movement
type TokenMovementResult struct {
	// Token is the moved token (with updated position)
	Token *models.Token `json:"token"`
	// MovementUsed is the amount of movement consumed (in feet)
	MovementUsed int `json:"movement_used"`
	// RemainingSpeed is the remaining movement for this turn (in feet)
	RemainingSpeed int `json:"remaining_speed"`
	// Path is the actual path taken (including intermediate positions)
	Path []models.Position `json:"path"`
	// DifficultTerrainCount is the number of difficult terrain squares traversed
	DifficultTerrainCount int `json:"difficult_terrain_count"`
}

// PathPosition represents a position in a movement path
type PathPosition struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	IsDiagonal bool `json:"is_diagonal"`
	IsDifficult bool `json:"is_difficult"`
}

// CalculateTokenMovementCost calculates the movement cost for a token moving between two positions
// 规则参考: PHB 第9章 - Movement and Position
// - Normal movement: 1 square = 5 feet
// - Difficult terrain: Each square costs 1 extra foot of movement
// - Diagonal movement: Every 2nd diagonal counts as 2 squares (simplified: 1.5x cost)
func CalculateTokenMovementCost(fromX, fromY, toX, toY, cellSize int, isDifficultTerrain bool) int {
	// Calculate base movement (Manhattan distance for simplicity)
	dx := toX - fromX
	if dx < 0 {
		dx = -dx
	}
	dy := toY - fromY
	if dy < 0 {
		dy = -dy
	}

	// Check if diagonal
	isDiagonal := (dx > 0 && dy > 0)

	// Calculate squares moved (for simplicity, use Manhattan distance)
	squares := dx + dy
	if squares == 0 {
		return 0
	}

	// Base movement in feet
	movement := squares * cellSize

	// Diagonal movement penalty (simplified: +50% cost)
	// 规则参考: PHB 第9章 - "When you move diagonally, every 2nd diagonal square counts as 2 squares"
	// For simplicity, we add 50% to diagonal movement
	if isDiagonal {
		movement = (movement * 3) / 2
	}

	// Difficult terrain doubles the cost
	// 规则参考: PHB 第8章 - "Every foot of movement in difficult terrain costs 1 extra foot"
	if isDifficultTerrain {
		movement *= 2
	}

	return movement
}

// CanTokenOccupySpace checks if a token can occupy a given space
// 规则参考: PHB 第9章 - Size and Space
// - A creature can move through a space occupied by a creature 2 or more sizes smaller
// - A creature cannot move through a space occupied by a creature of equal or larger size
func CanTokenOccupySpace(movingToken *models.Token, otherToken *models.Token) bool {
	// No other token, space is free
	if otherToken == nil {
		return true
	}

	// Same token, obviously can occupy
	if movingToken.ID == otherToken.ID {
		return true
	}

	// Compare sizes
	movingSize := getSizeRank(movingToken.Size)
	otherSize := getSizeRank(otherToken.Size)

	// Can move through spaces 2 or more sizes smaller
	return movingSize >= otherSize+2
}

// getSizeRank returns a numeric rank for token size comparison
func getSizeRank(size models.TokenSize) int {
	switch size {
	case models.TokenSizeTiny:
		return 0
	case models.TokenSizeSmall:
		return 1
	case models.TokenSizeMedium:
		return 2
	case models.TokenSizeLarge:
		return 3
	case models.TokenSizeHuge:
		return 4
	case models.TokenSizeGargantuan:
		return 5
	default:
		return 2 // Default to medium
	}
}

// CalculatePathCost calculates the total movement cost for a path
// Returns the cost in feet and the number of difficult terrain squares
func CalculatePathCost(path []PathPosition, cellSize int) (int, int) {
	totalCost := 0
	difficultCount := 0

	for i, pos := range path {
		if i == 0 {
			continue // Skip starting position
		}

		cost := 1 * cellSize // Base cost per square

		// Diagonal penalty
		if pos.IsDiagonal {
			cost = (cost * 3) / 2
		}

		// Difficult terrain penalty
		if pos.IsDifficult {
			cost *= 2
			difficultCount++
		}

		totalCost += cost
	}

	return totalCost, difficultCount
}

// GetTokenOccupiedSquares returns all squares occupied by a token
// 规则参考: PHB 第9章 - Size and Space
func GetTokenOccupiedSquares(token *models.Token) []models.Position {
	size := token.GetSizeInGrids()
	squares := make([]models.Position, 0, size*size)

	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			squares = append(squares, models.Position{
				X: token.Position.X + dx,
				Y: token.Position.Y + dy,
			})
		}
	}

	return squares
}

// ValidateTokenPosition checks if a token can be placed at the given position
func ValidateTokenPosition(token *models.Token, toX, toY, mapWidth, mapHeight int, grid *models.Grid) error {
	size := token.GetSizeInGrids()

	// Check map bounds
	if toX < 0 || toY < 0 || toX+size > mapWidth || toY+size > mapHeight {
		return NewMovementError("position_out_of_bounds", "token position is out of map bounds")
	}

	// Check for walls (if grid provided)
	if grid != nil {
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				cell := grid.GetCell(toX+dx, toY+dy)
				if cell == models.CellTypeWall {
					return NewMovementError("wall_blocking", "token cannot occupy a wall space")
				}
			}
		}
	}

	return nil
}

// MovementError represents an error during token movement
type MovementError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *MovementError) Error() string {
	return e.Message
}

// NewMovementError creates a new movement error
func NewMovementError(code, message string) *MovementError {
	return &MovementError{
		Code:    code,
		Message: message,
	}
}

// IsMovementError checks if an error is a MovementError
func IsMovementError(err error) bool {
	_, ok := err.(*MovementError)
	return ok
}
