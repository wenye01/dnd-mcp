// Package parser_test provides unit tests for FVTT Scene parser
package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFVTTSceneParser_CanParse(t *testing.T) {
	p := parser.NewFVTTSceneParser()

	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "valid scene JSON",
			data:     `{"_id":"test","name":"Test Scene","walls":[],"tokens":[]}`,
			expected: true,
		},
		{
			name:     "valid scene with tokens only",
			data:     `{"_id":"test","name":"Test Scene","tokens":[{"_id":"t1","x":0,"y":0}]}`,
			expected: true,
		},
		{
			name:     "valid scene with walls only",
			data:     `{"_id":"test","name":"Test Scene","walls":[{"_id":"w1","c":[[0,0],[100,100]]}]}`,
			expected: true,
		},
		{
			name:     "missing _id field",
			data:     `{"name":"Test Scene","walls":[],"tokens":[]}`,
			expected: false,
		},
		{
			name:     "missing name field",
			data:     `{"_id":"test","walls":[],"tokens":[]}`,
			expected: false,
		},
		{
			name:     "missing both walls and tokens",
			data:     `{"_id":"test","name":"Test Scene"}`,
			expected: false,
		},
		{
			name:     "empty JSON",
			data:     `{}`,
			expected: false,
		},
		{
			name:     "invalid JSON",
			data:     `not json`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.CanParse([]byte(tt.data))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFVTTSceneParser_Parse(t *testing.T) {
	p := parser.NewFVTTSceneParser()

	t.Run("valid minimal scene", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Test Scene",
			"width": 1000,
			"height": 800,
			"grid": 100,
			"gridType": 1,
			"gridDistance": 5,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, format.FormatFVTTScene, result.Format)
		assert.Empty(t, result.Warnings)

		scene, ok := result.Data.(*format.FVTTScene)
		require.True(t, ok)
		assert.Equal(t, "scene-001", scene.ID)
		assert.Equal(t, "Test Scene", scene.Name)
		assert.Equal(t, 1000, scene.Width)
		assert.Equal(t, 800, scene.Height)
		assert.Equal(t, 100, scene.Grid)
	})

	t.Run("scene with all fields", func(t *testing.T) {
		data := `{
			"_id": "scene-002",
			"name": "Complete Scene",
			"active": true,
			"navigation": true,
			"navOrder": 1,
			"navName": "Complete",
			"thumb": "thumbs/complete.png",
			"width": 2000,
			"height": 1500,
			"padding": 0.25,
			"backgroundColor": "#999999",
			"gridType": 1,
			"grid": 140,
			"gridDistance": 5,
			"gridUnits": "ft",
			"gridColor": "#000000",
			"gridAlpha": 0.2,
			"tokenVision": true,
			"fogExploration": true,
			"globalLight": false,
			"darkness": 0,
			"walls": [
				{
					"_id": "wall-001",
					"c": [[100, 100], [500, 100]],
					"move": 20,
					"sense": 20,
					"dir": 0,
					"door": 0,
					"ds": 0
				}
			],
			"tokens": [
				{
					"_id": "token-001",
					"name": "Player Token",
					"x": 700,
					"y": 700,
					"width": 70,
					"height": 70,
					"rotation": 0,
					"alpha": 1,
					"hidden": false,
					"locked": false
				}
			],
			"lights": [
				{
					"_id": "light-001",
					"x": 700,
					"y": 700,
					"dim": 10,
					"bright": 5,
					"angle": 360,
					"t": "l",
					"color": "#ffffff",
					"alpha": 0.5
				}
			],
			"tiles": [
				{
					"_id": "tile-001",
					"img": "maps/background.webp",
					"x": 0,
					"y": 0,
					"width": 2000,
					"height": 1500,
					"z": 0,
					"rotation": 0,
					"alpha": 1
				}
			]
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		require.NotNil(t, result)

		scene, ok := result.Data.(*format.FVTTScene)
		require.True(t, ok)

		assert.Equal(t, "scene-002", scene.ID)
		assert.Equal(t, "Complete Scene", scene.Name)
		assert.True(t, scene.Active)
		assert.True(t, scene.Navigation)
		assert.Equal(t, 1, scene.NavOrder)
		assert.Equal(t, "Complete", scene.NavName)
		assert.Equal(t, 2000, scene.Width)
		assert.Equal(t, 1500, scene.Height)
		assert.Equal(t, 140, scene.Grid)
		assert.Equal(t, 5, scene.GridDistance)
		assert.Equal(t, "ft", scene.GridUnits)
		assert.Len(t, scene.Walls, 1)
		assert.Len(t, scene.Tokens, 1)
		assert.Len(t, scene.Lights, 1)
		assert.Len(t, scene.Tiles, 1)
	})

	t.Run("missing required field _id", func(t *testing.T) {
		data := `{
			"name": "Test Scene",
			"width": 1000,
			"height": 800,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		assert.Error(t, err)
		assert.Nil(t, result)

		parseErr, ok := err.(*format.ParseError)
		require.True(t, ok)
		assert.Equal(t, format.FormatFVTTScene, parseErr.Format)
		assert.Contains(t, parseErr.Message, "_id")
	})

	t.Run("missing required field name", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"width": 1000,
			"height": 800,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		assert.Error(t, err)
		assert.Nil(t, result)

		parseErr, ok := err.(*format.ParseError)
		require.True(t, ok)
		assert.Contains(t, parseErr.Message, "name")
	})

	t.Run("invalid width", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Test Scene",
			"width": -100,
			"height": 800,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		data := `{invalid json}`

		result, err := p.Parse([]byte(data))
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestFVTTSceneParser_Format(t *testing.T) {
	p := parser.NewFVTTSceneParser()
	assert.Equal(t, format.FormatFVTTScene, p.Format())
}

func TestFVTTSceneParser_Warnings(t *testing.T) {
	p := parser.NewFVTTSceneParser()

	t.Run("hexagonal grid warning", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Hex Scene",
			"width": 1000,
			"height": 800,
			"gridType": 2,
			"grid": 100,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "hexagonal")
	})

	t.Run("very large scene warning", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Large Scene",
			"width": 5000,
			"height": 5000,
			"gridType": 1,
			"grid": 100,
			"gridDistance": 5,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "very large")
	})

	t.Run("tile without image warning", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Scene",
			"width": 1000,
			"height": 800,
			"gridType": 1,
			"grid": 100,
			"gridDistance": 5,
			"walls": [],
			"tokens": [],
			"tiles": [{"_id": "tile-001"}]
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "no image")
	})

	t.Run("unset grid distance warning", func(t *testing.T) {
		data := `{
			"_id": "scene-001",
			"name": "Scene",
			"width": 1000,
			"height": 800,
			"gridType": 1,
			"grid": 100,
			"gridDistance": 0,
			"walls": [],
			"tokens": []
		}`

		result, err := p.Parse([]byte(data))
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "grid distance")
	})
}

func TestFVTTScene_ParseSampleFile(t *testing.T) {
	// Get the project root directory
	projectRoot := filepath.Join("..", "..", "..", "..")
	testDataPath := filepath.Join(projectRoot, "tests", "testdata", "maps", "sample_scene.json")

	// Check if file exists
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("sample_scene.json not found, skipping test")
		return
	}

	p := parser.NewFVTTSceneParser()

	// Read the test file
	data, err := os.ReadFile(testDataPath)
	require.NoError(t, err)

	// Parse the file
	result, err := p.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, result)

	scene, ok := result.Data.(*format.FVTTScene)
	require.True(t, ok)

	// Verify key fields
	assert.Equal(t, "sample-scene-001", scene.ID)
	assert.Equal(t, "Sample Test Room", scene.Name)
	assert.Equal(t, 1400, scene.Width)
	assert.Equal(t, 1400, scene.Height)
	assert.Equal(t, 140, scene.Grid)
	assert.Equal(t, 5, scene.GridDistance)
	assert.Equal(t, "ft", scene.GridUnits)

	// Verify content
	assert.Len(t, scene.Walls, 4)
	assert.Len(t, scene.Tokens, 1)
	assert.Len(t, scene.Lights, 1)
	assert.Len(t, scene.Tiles, 1)

	// Verify wall structure
	wall := scene.Walls[0]
	assert.Equal(t, "wall-001", wall.ID)
	assert.Len(t, wall.C, 2)
	assert.InDelta(t, 140.0, wall.C[0][0], 0.001)
	assert.InDelta(t, 140.0, wall.C[0][1], 0.001)
	assert.InDelta(t, 1260.0, wall.C[1][0], 0.001)
	assert.InDelta(t, 140.0, wall.C[1][1], 0.001)

	// Verify token structure
	token := scene.Tokens[0]
	assert.Equal(t, "token-001", token.ID)
	assert.Equal(t, "Test Token", token.Name)
	assert.Equal(t, 700, token.X)
	assert.Equal(t, 700, token.Y)
	assert.Equal(t, 70, token.Width)
	assert.Equal(t, 70, token.Height)

	// Verify tile structure
	tile := scene.Tiles[0]
	assert.Equal(t, "tile-001", tile.ID)
	assert.Equal(t, "maps/images/small.webp", tile.Image)
}

func TestFVTTScene_HelperMethods(t *testing.T) {
	data := `{
		"_id": "scene-001",
		"name": "Test Scene",
		"width": 1400,
		"height": 1050,
		"gridType": 1,
		"grid": 140,
		"gridDistance": 5,
		"gridUnits": "ft",
		"walls": [],
		"tokens": []
	}`

	p := parser.NewFVTTSceneParser()
	result, err := p.Parse([]byte(data))
	require.NoError(t, err)

	scene := result.Data.(*format.FVTTScene)

	// Test helper methods
	assert.Equal(t, 140, scene.GetGridSizeInPixels())
	assert.Equal(t, "ft", scene.GetGridUnits())
	assert.Equal(t, 5, scene.GetGridDistance())
	width, height := scene.GetDimensionsInGrid()
	assert.Equal(t, 10, width) // 1400/140 = 10
	assert.Equal(t, 7, height) // 1050/140 = 7.5 -> truncated

	assert.True(t, scene.IsSquareGrid())
	assert.False(t, scene.IsHexGrid())
	assert.True(t, scene.HasGrid())
}

func TestFVTTWall_HelperMethods(t *testing.T) {
	tests := []struct {
		name       string
		wall       format.FVTTWall
		wallType   string
		isDoor     bool
		blocksMove bool
		blocksVision bool
		doorState  string
	}{
		{
			name: "normal wall",
			wall: format.FVTTWall{Door: 0, Move: 0, Sense: 0},
			wallType:   "wall",
			isDoor:     false,
			blocksMove: true,
			blocksVision: true,
			doorState:  "none",
		},
		{
			name: "closed door",
			wall: format.FVTTWall{Door: 1, DS: 0, Move: 0, Sense: 1},
			wallType:   "door",
			isDoor:     true,
			blocksMove: true,
			blocksVision: true,
			doorState:  "closed",
		},
		{
			name: "open door",
			wall: format.FVTTWall{Door: 1, DS: 1, Move: 2, Sense: 2},
			wallType:   "door",
			isDoor:     true,
			blocksMove: false,
			blocksVision: false,
			doorState:  "open",
		},
		{
			name: "secret door",
			wall: format.FVTTWall{Door: 2, DS: 0, Move: 0, Sense: 0},
			wallType:   "secret_door",
			isDoor:     true,
			blocksMove: true,
			blocksVision: true,
			doorState:  "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wallType, tt.wall.GetWallType())
			assert.Equal(t, tt.isDoor, tt.wall.IsDoor())
			assert.Equal(t, tt.blocksMove, tt.wall.BlocksMovement())
			assert.Equal(t, tt.blocksVision, tt.wall.BlocksVision())
			assert.Equal(t, tt.doorState, tt.wall.GetDoorState())
		})
	}
}
