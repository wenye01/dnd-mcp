// Package converter provides converters for transforming parsed data to Map models
package converter

import (
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/google/uuid"
)

// MapConverter converts parsed map data to Map models
type MapConverter struct{}

// NewMapConverter creates a new map converter
func NewMapConverter() *MapConverter {
	return &MapConverter{}
}

// Format returns the format this converter handles
func (c *MapConverter) Format() format.ImportFormat {
	return format.FormatAuto // Handles multiple formats
}

// Convert converts parsed data to a Map model
// It detects the type of parsedData and dispatches to the appropriate conversion method
func (c *MapConverter) Convert(parsedData interface{}, opts format.ImportOptions) (*models.Map, error) {
	switch data := parsedData.(type) {
	case *format.UVTTData:
		return c.ConvertFromUVTT(data, opts)
	case *format.FVTTScene:
		return c.ConvertFromFVTTScene(data, opts)
	case map[string]interface{}:
		// Try to detect if it's FVTT Scene data
		if _, hasID := data["_id"]; hasID {
			if _, hasGrid := data["grid"]; hasGrid {
				// Likely FVTT Scene - marshal and unmarshal to convert
				jsonData, err := json.Marshal(data)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal scene data: %w", err)
				}
				var scene format.FVTTScene
				if err := json.Unmarshal(jsonData, &scene); err != nil {
					return nil, fmt.Errorf("failed to unmarshal scene data: %w", err)
				}
				return c.ConvertFromFVTTScene(&scene, opts)
			}
		}
		return nil, fmt.Errorf("unable to determine data format for conversion")
	default:
		return nil, fmt.Errorf("unsupported data type for conversion: %T", parsedData)
	}
}

// ConvertFromUVTT converts UVTT data to a Map model
func (c *MapConverter) ConvertFromUVTT(uvtt *format.UVTTData, opts format.ImportOptions) (*models.Map, error) {
	// Calculate grid dimensions
	gridWidth := uvtt.Resolution.MapSize.X
	gridHeight := uvtt.Resolution.MapSize.Y

	if gridWidth <= 0 || gridHeight <= 0 {
		return nil, fmt.Errorf("invalid map dimensions: %dx%d", gridWidth, gridHeight)
	}

	// Create the map using the constructor
	gameMap := models.NewBattleMap(opts.CampaignID, opts.Name, gridWidth, gridHeight, 5)

	// Convert walls if requested
	if opts.ImportWalls {
		walls := c.convertUVTTWalls(uvtt.Walls, uvtt.Resolution.PixelsPerGrid)
		for _, wall := range walls {
			if err := gameMap.Walls.Add(wall); err != nil {
				// Log warning but continue
				continue
			}
		}

		// Also convert portals as doors
		portalWalls := c.convertUVTTPortals(uvtt.Portals, uvtt.Resolution.PixelsPerGrid)
		for _, wall := range portalWalls {
			if err := gameMap.Walls.Add(wall); err != nil {
				continue
			}
		}
	}

	// Convert tokens if requested
	if opts.ImportTokens {
		tokens := c.convertUVTTTokens(uvtt.Tokens, uvtt.Resolution.PixelsPerGrid)
		for _, token := range tokens {
			if err := gameMap.AddToken(token); err != nil {
				continue
			}
		}
	}

	if err := gameMap.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return gameMap, nil
}

// convertUVTTWalls converts UVTT walls to model walls
func (c *MapConverter) convertUVTTWalls(uvttWalls []format.UVTTWall, pixelsPerGrid int) []*models.Wall {
	walls := make([]*models.Wall, 0, len(uvttWalls))

	for _, w := range uvttWalls {
		// Convert pixel coordinates to grid coordinates
		x1 := w.Bounds.X / pixelsPerGrid
		y1 := w.Bounds.Y / pixelsPerGrid
		x2 := (w.Bounds.X + w.Bounds.W) / pixelsPerGrid
		y2 := (w.Bounds.Y + w.Bounds.H) / pixelsPerGrid

		// Skip zero-length walls
		if x1 == x2 && y1 == y2 {
			continue
		}

		// Determine wall type
		wallType := models.WallTypeWall
		if w.Door {
			wallType = models.WallTypeDoor
		}

		// Determine move/sense restrictions
		moveRestrict := 0 // blocks
		if w.Move == "none" {
			moveRestrict = 0
		} else if w.Move == "normal" {
			moveRestrict = 2 // allows
		}

		wall := models.NewWall(uuid.NewString(), wallType, x1, y1, x2, y2, moveRestrict, 0)

		// Add door data if it's a door
		if w.Door {
			wall.Door = &models.WallDoor{
				State:  models.DoorStateClosed,
				Secret: false,
			}
		}

		walls = append(walls, wall)
	}

	return walls
}

// convertUVTTPortals converts UVTT portals to model walls (doors)
func (c *MapConverter) convertUVTTPortals(portals []format.UVTTPortal, pixelsPerGrid int) []*models.Wall {
	walls := make([]*models.Wall, 0, len(portals))

	for _, p := range portals {
		// Convert pixel coordinates to grid coordinates
		x1 := p.Bounds.X / pixelsPerGrid
		y1 := p.Bounds.Y / pixelsPerGrid
		x2 := (p.Bounds.X + p.Bounds.W) / pixelsPerGrid
		y2 := (p.Bounds.Y + p.Bounds.H) / pixelsPerGrid

		// Skip zero-length
		if x1 == x2 && y1 == y2 {
			continue
		}

		// Determine move restriction based on closed state
		moveRestrict := 2 // allows by default
		if p.Closed {
			moveRestrict = 0 // blocks when closed
		}

		wall := models.NewWall(uuid.NewString(), models.WallTypeDoor, x1, y1, x2, y2, moveRestrict, 1)
		wall.Door = &models.WallDoor{
			State:  models.DoorStateClosed,
			Secret: false,
		}
		if !p.Closed {
			wall.Door.State = models.DoorStateOpen
		}

		walls = append(walls, wall)
	}

	return walls
}

// convertUVTTTokens converts UVTT tokens to model tokens
func (c *MapConverter) convertUVTTTokens(uvttTokens []format.UVTTToken, pixelsPerGrid int) []models.Token {
	tokens := make([]models.Token, 0, len(uvttTokens))

	for _, t := range uvttTokens {
		// Convert pixel coordinates to grid coordinates
		gridX := t.X / pixelsPerGrid
		gridY := t.Y / pixelsPerGrid

		// Determine token size
		size := models.TokenSizeMedium
		if t.Size > 1 {
			switch {
			case t.Size >= 4:
				size = models.TokenSizeGargantuan
			case t.Size >= 3:
				size = models.TokenSizeHuge
			case t.Size >= 2:
				size = models.TokenSizeLarge
			}
		}

		tokenID := t.ID
		if tokenID == "" {
			tokenID = uuid.NewString()
		}

		token := models.Token{
			ID:       tokenID,
			Name:     t.Name,
			Position: models.Position{X: gridX, Y: gridY},
			Size:     size,
			Hidden:   t.Hidden,
			Locked:   t.Locked,
		}

		tokens = append(tokens, token)
	}

	return tokens
}

// ConvertFromFVTTScene converts FVTT Scene data to a Map model
func (c *MapConverter) ConvertFromFVTTScene(scene *format.FVTTScene, opts format.ImportOptions) (*models.Map, error) {
	// Calculate grid dimensions
	gridWidth, gridHeight := scene.GetDimensionsInGrid()
	if gridWidth <= 0 || gridHeight <= 0 {
		// Fallback to pixel dimensions if grid calculation fails
		if scene.Grid > 0 {
			gridWidth = scene.Width / scene.Grid
			gridHeight = scene.Height / scene.Grid
		}
	}

	if gridWidth <= 0 || gridHeight <= 0 {
		return nil, fmt.Errorf("invalid map dimensions: %dx%d", gridWidth, gridHeight)
	}

	// Create the map
	name := opts.Name
	if name == "" {
		name = scene.Name
	}

	gameMap := models.NewBattleMap(opts.CampaignID, name, gridWidth, gridHeight, scene.GetGridDistance())

	// Convert walls if requested
	if opts.ImportWalls {
		walls := c.convertFVTTWalls(scene.Walls, scene.Grid)
		for _, wall := range walls {
			if err := gameMap.Walls.Add(wall); err != nil {
				continue
			}
		}
	}

	// Convert tokens if requested
	if opts.ImportTokens {
		tokens := c.convertFVTTTokens(scene.Tokens, scene.Grid)
		for _, token := range tokens {
			if err := gameMap.AddToken(token); err != nil {
				continue
			}
		}
	}

	if err := gameMap.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return gameMap, nil
}

// convertFVTTWalls converts FVTT walls to model walls
func (c *MapConverter) convertFVTTWalls(fvttWalls []format.FVTTWall, gridSize int) []*models.Wall {
	walls := make([]*models.Wall, 0, len(fvttWalls))

	for _, w := range fvttWalls {
		// FVTT walls have coordinates in format [[x1, y1], [x2, y2]]
		if len(w.C) < 2 || len(w.C[0]) < 2 || len(w.C[1]) < 2 {
			continue
		}

		// Convert pixel coordinates to grid coordinates
		x1 := w.C[0][0] / gridSize
		y1 := w.C[0][1] / gridSize
		x2 := w.C[1][0] / gridSize
		y2 := w.C[1][1] / gridSize

		// Skip zero-length walls
		if x1 == x2 && y1 == y2 {
			continue
		}

		// Determine wall type
		wallType := models.WallTypeWall
		if w.Door == 1 {
			wallType = models.WallTypeDoor
		} else if w.Door == 2 {
			wallType = models.WallTypeDoor // Secret door
		}

		wall := models.NewWall(w.ID, wallType, x1, y1, x2, y2, w.Move, w.Sense)

		if wall.ID == "" {
			wall.ID = uuid.NewString()
		}

		// Add door data if it's a door
		if w.Door > 0 {
			state := models.DoorStateClosed
			if w.DS == 1 {
				state = models.DoorStateOpen
			} else if w.DS == 2 {
				state = models.DoorStateLocked
			}

			wall.Door = &models.WallDoor{
				State:  state,
				Secret: w.Door == 2,
			}
		}

		walls = append(walls, wall)
	}

	return walls
}

// convertFVTTTokens converts FVTT tokens to model tokens
func (c *MapConverter) convertFVTTTokens(fvttTokens []format.FVTTToken, gridSize int) []models.Token {
	tokens := make([]models.Token, 0, len(fvttTokens))

	for _, t := range fvttTokens {
		// Convert pixel coordinates to grid coordinates
		gridX := t.X / gridSize
		gridY := t.Y / gridSize

		// Determine token size from dimensions
		size := models.TokenSizeMedium
		if t.Width > gridSize {
			switch {
			case t.Width >= gridSize*4:
				size = models.TokenSizeGargantuan
			case t.Width >= gridSize*3:
				size = models.TokenSizeHuge
			case t.Width >= gridSize*2:
				size = models.TokenSizeLarge
			}
		}

		tokenID := t.ID
		if tokenID == "" {
			tokenID = uuid.NewString()
		}

		token := models.Token{
			ID:       tokenID,
			Name:     t.Name,
			Position: models.Position{X: gridX, Y: gridY},
			Size:     size,
			Hidden:   t.Hidden,
			Locked:   t.Locked,
		}

		tokens = append(tokens, token)
	}

	return tokens
}
