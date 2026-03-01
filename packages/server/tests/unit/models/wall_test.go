// Package models_test provides unit tests for wall models
package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dnd-mcp/server/internal/models"
)

func TestNewWall(t *testing.T) {
	wall := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)

	assert.Equal(t, "wall-001", wall.ID)
	assert.Equal(t, models.WallTypeWall, wall.Type)
	assert.Equal(t, 0, wall.Move) // Blocks movement
	assert.Equal(t, []int{0, 0, 10, 0}, wall.Bounds)
}

func TestNewDoor(t *testing.T) {
	door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)

	assert.Equal(t, "door-001", door.ID)
	assert.Equal(t, models.WallTypeDoor, door.Type)
	assert.Equal(t, models.WallDirectionLeft, door.Direction)
	assert.NotNil(t, door.Door)
	assert.Equal(t, models.DoorStateClosed, door.Door.State)
	assert.False(t, door.Door.Secret)
	assert.True(t, door.IsBlocking()) // Closed doors block movement
}

func TestWallValidate(t *testing.T) {
	tests := []struct {
		name        string
		wall        *models.Wall
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid wall",
			wall: models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0),
		},
		{
			name:        "empty ID",
			wall:        models.NewWall("", models.WallTypeWall, 0, 0, 10, 0, 0, 0),
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name: "valid bounds",
			wall: models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 5, 0, 0),
		},
		{
			name:        "negative coordinates",
			wall:        models.NewWall("wall-001", models.WallTypeWall, -1, 0, 10, 0, 0, 0),
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name:        "invalid move restriction",
			wall:        func() *models.Wall {
				w := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 3, 0)
				return w
			}(),
			expectError: true,
			errorMsg:    "must be 0 (blocks), 1 (difficult), or 2 (allows)",
		},
		{
			name:        "invalid sense restriction",
			wall:        func() *models.Wall {
				w := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, -1)
				return w
			}(),
			expectError: true,
			errorMsg:    "must be 0 (blocks), 1 (limited), or 2 (allows)",
		},
		{
			name: "valid door",
			wall: models.NewDoor("door-001", 5, 5, models.WallDirectionLeft),
		},
		{
			name: "secret door with valid DC",
			wall: func() *models.Wall {
				door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)
				door.Door.Secret = true
				door.Door.DC = 15
				return door
			}(),
		},
		{
			name: "secret door with invalid DC",
			wall: func() *models.Wall {
				door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)
				door.Door.Secret = true
				door.Door.DC = 35 // Too high
				return door
			}(),
			expectError: true,
			errorMsg:    "must be between 0 and 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.wall.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWallIsBlocking(t *testing.T) {
	tests := []struct {
		name     string
		move     int
		expected bool
	}{
		{
			name:     "blocks movement",
			move:     0,
			expected: true,
		},
		{
			name:     "difficult terrain",
			move:     1,
			expected: false,
		},
		{
			name:     "allows movement",
			move:     2,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wall := &models.Wall{ID: "wall-001", Move: tt.move}
			assert.Equal(t, tt.expected, wall.IsBlocking())
		})
	}
}

func TestWallBlocksVision(t *testing.T) {
	tests := []struct {
		name     string
		sense    int
		expected bool
	}{
		{
			name:     "blocks vision",
			sense:    0,
			expected: true,
		},
		{
			name:     "limits vision",
			sense:    1,
			expected: false,
		},
		{
			name:     "allows vision",
			sense:    2,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wall := &models.Wall{ID: "wall-001", Sense: tt.sense}
			assert.Equal(t, tt.expected, wall.BlocksVision())
		})
	}
}

func TestDoorOpenClose(t *testing.T) {
	door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)

	// Initially closed
	assert.True(t, door.IsBlocking())
	assert.False(t, door.IsOpen())

	// Open the door
	err := door.Open()
	require.NoError(t, err)
	assert.True(t, door.IsOpen())
	assert.False(t, door.IsBlocking())
	assert.Equal(t, models.DoorStateOpen, door.Door.State)

	// Close the door
	err = door.Close()
	require.NoError(t, err)
	assert.False(t, door.IsOpen())
	assert.True(t, door.IsBlocking())
	assert.Equal(t, models.DoorStateClosed, door.Door.State)
}

func TestDoorLocked(t *testing.T) {
	door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)
	door.Door.State = models.DoorStateLocked

	// Cannot open locked door
	err := door.Open()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "locked")

	// Unlock by setting state to closed
	door.Door.State = models.DoorStateClosed
	err = door.Open()
	assert.NoError(t, err)
	assert.True(t, door.IsOpen())
}

func TestWallAdd(t *testing.T) {
	var walls models.Walls

	wall1 := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)
	wall2 := models.NewWall("wall-002", models.WallTypeWall, 0, 5, 10, 5, 0, 0)

	// Add walls
	err := walls.Add(wall1)
	require.NoError(t, err)
	err = walls.Add(wall2)
	require.NoError(t, err)

	assert.Len(t, walls, 2)
}

func TestWallRemove(t *testing.T) {
	var walls models.Walls

	wall1 := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)
	wall2 := models.NewWall("wall-002", models.WallTypeWall, 0, 5, 10, 5, 0, 0)

	walls.Add(wall1)
	walls.Add(wall2)

	// Remove wall1
	removed := walls.Remove("wall-001")
	assert.True(t, removed)
	assert.Len(t, walls, 1)

	// Try to remove again
	removed = walls.Remove("wall-001")
	assert.False(t, removed)
	assert.Len(t, walls, 1)
}

func TestWallGet(t *testing.T) {
	var walls models.Walls

	wall1 := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)
	wall2 := models.NewWall("wall-002", models.WallTypeWall, 0, 5, 10, 5, 0, 0)

	walls.Add(wall1)
	walls.Add(wall2)

	// Get existing wall
	found := walls.Get("wall-001")
	assert.NotNil(t, found)
	assert.Equal(t, "wall-001", found.ID)

	// Get non-existent wall
	notFound := walls.Get("wall-999")
	assert.Nil(t, notFound)
}

func TestWallGetWallsAtPoint(t *testing.T) {
	var walls models.Walls

	// Add a horizontal wall from (0,0) to (10,0)
	wall1 := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)
	walls.Add(wall1)

	// Add a vertical wall from (5,0) to (5,10)
	wall2 := models.NewWall("wall-002", models.WallTypeWall, 5, 0, 5, 10, 0, 0)
	walls.Add(wall2)

	// Point on horizontal wall (not at intersection)
	atHorizontal := walls.GetWallsAtPoint(3, 0)
	assert.Len(t, atHorizontal, 1)
	assert.Equal(t, "wall-001", atHorizontal[0].ID)

	// Point at intersection (both walls cross at (5,0))
	atIntersection := walls.GetWallsAtPoint(5, 0)
	assert.Len(t, atIntersection, 2)

	// Point on vertical wall (not at intersection)
	atVertical := walls.GetWallsAtPoint(5, 5)
	assert.Len(t, atVertical, 1)
	assert.Equal(t, "wall-002", atVertical[0].ID)

	// Point not on any wall
	notOnAny := walls.GetWallsAtPoint(20, 20)
	assert.Len(t, notOnAny, 0)
}

func TestWallGetDoors(t *testing.T) {
	var walls models.Walls

	wall := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)
	door := models.NewDoor("door-001", 5, 5, models.WallDirectionLeft)
	secretDoor := models.NewDoor("door-002", 10, 10, models.WallDirectionRight)
	secretDoor.Door.Secret = true

	walls.Add(wall)
	walls.Add(door)
	walls.Add(secretDoor)

	// Get all doors
	doors := walls.GetDoors()
	assert.Len(t, doors, 2)

	// Get secret doors
	secretDoors := walls.GetSecretDoors()
	assert.Len(t, secretDoors, 1)
	assert.Equal(t, "door-002", secretDoors[0].ID)
}

func TestWallValidateAll(t *testing.T) {
	var walls models.Walls

	// Add valid walls
	walls.Add(models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0))
	walls.Add(models.NewDoor("door-001", 5, 5, models.WallDirectionLeft))

	assert.NoError(t, walls.Validate())

	// Try to add invalid wall (should fail on Add, not on Validate)
	badWall := models.NewWall("bad-wall", models.WallTypeWall, 0, 0, 10, 0, 5, 0)
	err := walls.Add(badWall)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 0 (blocks), 1 (difficult), or 2 (allows)")
}

func TestWallJSON(t *testing.T) {
	wall := models.NewWall("wall-001", models.WallTypeWall, 0, 0, 10, 0, 0, 0)

	// To JSON
	jsonStr, err := wall.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, jsonStr, "wall-001")

	// From JSON
	parsedWall, err := models.WallFromJSON(jsonStr)
	require.NoError(t, err)
	assert.Equal(t, wall.ID, parsedWall.ID)
	assert.Equal(t, wall.Type, parsedWall.Type)
}
