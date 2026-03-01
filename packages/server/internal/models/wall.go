// Package models provides wall model definitions for map boundaries
// 规则参考: Foundry VTT Wall Data Structure
package models

import (
	"encoding/json"
	"fmt"
)

// WallType represents the type of wall
type WallType string

const (
	// WallTypeWall Normal wall that blocks movement
	WallTypeWall WallType = "wall"
	// WallTypeDoor Door that can be opened/closed
	WallTypeDoor WallType = "door"
	// WallTypeWindow Window that blocks movement but allows vision
	WallTypeWindow WallType = "window"
	// WallTypeTerrain Terrain boundary (difficult terrain)
	WallTypeTerrain WallType = "terrain"
)

// WallDoorState represents the state of a door
type WallDoorState string

const (
	// DoorStateClosed Door is closed (blocks movement)
	DoorStateClosed WallDoorState = "closed"
	// DoorStateOpen Door is open (allows movement)
	DoorStateOpen WallDoorState = "open"
	// DoorStateLocked Door is locked (requires action to open)
	DoorStateLocked WallDoorState = "locked"
)

// WallDirection represents the direction of a wall segment
type WallDirection string

const (
	// WallDirectionLeft Wall extends to the left
	WallDirectionLeft WallDirection = "left"
	// WallDirectionRight Wall extends to the right
	WallDirectionRight WallDirection = "right"
	// WallDirectionTop Wall extends upward
	WallDirectionTop WallDirection = "top"
	// WallDirectionBottom Wall extends downward
	WallDirectionBottom WallDirection = "bottom"
)

// Wall represents a wall or boundary on a map
// Used for Foundry VTT compatibility and tactical combat
// 规则参考: PHB 第9章 - Cover and Obstacles
type Wall struct {
	// ID is the unique identifier for this wall
	ID string `json:"id"`

	// Type is the type of wall (wall, door, window, terrain)
	Type WallType `json:"type"`

	// Coordinates are in grid units
	C  struct {
		X int `json:"x"` // X coordinate in grid units
		Y int `json:"y"` // Y coordinate in grid units
	} `json:"c"`

	// Bounds define the wall segment [x1, y1, x2, y2]
	Bounds []int `json:"bounds"`

	// Direction indicates which way the wall faces
	Direction WallDirection `json:"direction,omitempty"`

	// Door is specific data for door-type walls
	Door *WallDoor `json:"door,omitempty"`

	// Move restricts movement through this wall
	Move int `json:"move"` // 0 = blocks, 1 = difficult, 2 = allows

	// Sense restricts sensing through this wall (vision, hearing, etc.)
	Sense int `json:"sense"` // 0 = blocks, 1 = limited, 2 = allows

	// Lightweight indicates this is a temporary wall
	Lightweight bool `json:"lightweight,omitempty"`
}

// WallDoor contains door-specific properties
type WallDoor struct {
	// State is the current door state
	State WallDoorState `json:"state"`

	// Secret is true if this is a secret door
	Secret bool `json:"secret"`

	// DC is the difficulty class to find/open a secret door
	DC int `json:"dc,omitempty"`

	// LockedDC is the difficulty class to pick the lock
	LockedDC int `json:"locked_dc,omitempty"`
}

// NewWall creates a new wall between two points
// 规则参考: PHB 第9章 - Cover
func NewWall(id string, wallType WallType, x1, y1, x2, y2 int, moveRestrict, senseRestrict int) *Wall {
	return &Wall{
		ID:   id,
		Type: wallType,
		C: struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{X: x1, Y: y1},
		Bounds: []int{x1, y1, x2, y2},
		Move:   moveRestrict,
		Sense:  senseRestrict,
	}
}

// NewDoor creates a new door wall
func NewDoor(id string, x, y int, direction WallDirection) *Wall {
	wall := &Wall{
		ID:   id,
		Type: WallTypeDoor,
		C: struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{X: x, Y: y},
		Bounds:    []int{x, y, x + 1, y + 1}, // Default 1x1 door
		Direction: direction,
		Move:      0, // Blocks movement when closed
		Sense:     1, // Limits vision when closed
		Door: &WallDoor{
			State: DoorStateClosed,
			Secret: false,
		},
	}
	return wall
}

// Validate validates the wall data
func (w *Wall) Validate() error {
	if w.ID == "" {
		return NewValidationError("wall.id", "cannot be empty")
	}

	// Validate bounds
	if len(w.Bounds) != 4 {
		return NewValidationError("wall.bounds", "must have exactly 4 coordinates [x1,y1,x2,y2]")
	}

	x1, y1, x2, y2 := w.Bounds[0], w.Bounds[1], w.Bounds[2], w.Bounds[3]

	// Validate coordinates are non-negative
	if x1 < 0 || y1 < 0 || x2 < 0 || y2 < 0 {
		return NewValidationError("wall.bounds", "coordinates cannot be negative")
	}

	// Validate move restriction
	if w.Move < 0 || w.Move > 2 {
		return NewValidationError("wall.move", "must be 0 (blocks), 1 (difficult), or 2 (allows)")
	}

	// Validate sense restriction
	if w.Sense < 0 || w.Sense > 2 {
		return NewValidationError("wall.sense", "must be 0 (blocks), 1 (limited), or 2 (allows)")
	}

	// Validate door data if present
	if w.Type == WallTypeDoor && w.Door != nil {
		if err := w.Door.Validate(); err != nil {
			return NewValidationError("wall.door", err.Error())
		}
	}

	return nil
}

// Validate validates door-specific data
func (d *WallDoor) Validate() error {
	if d.State != DoorStateClosed && d.State != DoorStateOpen && d.State != DoorStateLocked {
		return NewValidationError("door.state", "must be closed, open, or locked")
	}

	if d.DC < 0 || d.DC > 30 {
		return NewValidationError("door.dc", "must be between 0 and 30")
	}

	if d.LockedDC < 0 || d.LockedDC > 30 {
		return NewValidationError("door.locked_dc", "must be between 0 and 30")
	}

	return nil
}

// IsBlocking returns true if the wall blocks movement
func (w *Wall) IsBlocking() bool {
	return w.Move == 0
}

// IsDifficult returns true if the wall makes movement difficult
func (w *Wall) IsDifficult() bool {
	return w.Move == 1
}

// IsSecret returns true if this is a secret door
func (w *Wall) IsSecret() bool {
	return w.Type == WallTypeDoor && w.Door != nil && w.Door.Secret
}

// BlocksVision returns true if the wall blocks vision
func (w *Wall) BlocksVision() bool {
	return w.Sense == 0
}

// IsOpen returns true if a door is open
func (w *Wall) IsOpen() bool {
	return w.Type == WallTypeDoor && w.Door != nil && w.Door.State == DoorStateOpen
}

// Open opens a door wall
func (w *Wall) Open() error {
	if w.Type != WallTypeDoor {
		return fmt.Errorf("cannot open: not a door")
	}
	if w.Door == nil {
		return fmt.Errorf("cannot open: door data is missing")
	}
	if w.Door.State == DoorStateLocked {
		return fmt.Errorf("cannot open: door is locked")
	}

	w.Door.State = DoorStateOpen
	w.Move = 2 // Allow movement
	w.Sense = 2 // Allow vision
	return nil
}

// Close closes a door wall
func (w *Wall) Close() error {
	if w.Type != WallTypeDoor {
		return fmt.Errorf("cannot close: not a door")
	}
	if w.Door == nil {
		return fmt.Errorf("cannot close: door data is missing")
	}

	w.Door.State = DoorStateClosed
	w.Move = 0 // Block movement
	w.Sense = 1 // Limit vision
	return nil
}

// ToJSON converts the wall to JSON
func (w *Wall) ToJSON() (string, error) {
	data, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WallFromJSON parses a wall from JSON
func WallFromJSON(jsonStr string) (*Wall, error) {
	var wall Wall
	err := json.Unmarshal([]byte(jsonStr), &wall)
	if err != nil {
		return nil, err
	}
	return &wall, nil
}

// Walls represents a collection of walls with helper methods
type Walls []*Wall

// Add adds a wall to the collection
func (ws *Walls) Add(wall *Wall) error {
	if err := wall.Validate(); err != nil {
		return err
	}
	*ws = append(*ws, wall)
	return nil
}

// Remove removes a wall by ID
func (ws *Walls) Remove(id string) bool {
	for i, wall := range *ws {
		if wall.ID == id {
			*ws = append((*ws)[:i], (*ws)[i+1:]...)
			return true
		}
	}
	return false
}

// Get retrieves a wall by ID
func (ws *Walls) Get(id string) *Wall {
	for _, wall := range *ws {
		if wall.ID == id {
			return wall
		}
	}
	return nil
}

// GetWallsAtPoint returns all walls at a given point
func (ws *Walls) GetWallsAtPoint(x, y int) Walls {
	var result Walls
	for _, wall := range *ws {
		x1, y1, x2, y2 := wall.Bounds[0], wall.Bounds[1], wall.Bounds[2], wall.Bounds[3]
		// Check if point is on the wall segment
		if (x >= x1 && x <= x2) || (x >= x2 && x <= x1) {
			if (y >= y1 && y <= y2) || (y >= y2 && y <= y1) {
				result = append(result, wall)
			}
		}
	}
	return result
}

// GetBlockingWallsAtPoint returns walls that block movement at a given point
func (ws *Walls) GetBlockingWallsAtPoint(x, y int) Walls {
	var result Walls
	for _, wall := range ws.GetWallsAtPoint(x, y) {
		if wall.IsBlocking() {
			result = append(result, wall)
		}
	}
	return result
}

// GetDoors returns all door-type walls
func (ws *Walls) GetDoors() Walls {
	var result Walls
	for _, wall := range *ws {
		if wall.Type == WallTypeDoor {
			result = append(result, wall)
		}
	}
	return result
}

// GetSecretDoors returns all secret doors
func (ws *Walls) GetSecretDoors() Walls {
	var result Walls
	for _, wall := range *ws {
		if wall.IsSecret() {
			result = append(result, wall)
		}
	}
	return result
}

// Validate validates all walls in the collection
func (ws *Walls) Validate() error {
	for i, wall := range *ws {
		if err := wall.Validate(); err != nil {
			return fmt.Errorf("wall at index %d: %w", i, err)
		}
	}
	return nil
}
