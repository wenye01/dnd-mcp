// Package importer_test provides unit tests for BUG-M6.5-001 fixes
// These tests verify the type compatibility fixes for FVTT Scene and UVTT imports
package importer_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dnd-mcp/server/internal/importer/converter"
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFIX1_UVTTDefaultNameGeneration verifies that UVTT import generates a default name when none is provided
// BUG-M6.5-001 P0-1: UVTT format has no name field, code should auto-generate
func TestFIX1_UVTTDefaultNameGeneration(t *testing.T) {
	c := converter.NewMapConverter()

	// Create sample UVTT data (without name field - UVTT format doesn't have one)
	uvttData := &format.UVTTData{
		Format: 2,
		Resolution: format.UVTTResolution{
			MapSize:        format.UVTTMapSize{X: 20, Y: 15},
			PixelsPerGrid:  70,
		},
		LineOfSight: []format.UVTTLineOfSight{},
		Portals:     []format.UVTTPortal{},
		Lights:      []format.UVTTLight{},
		Tokens:      []format.UVTTToken{},
		Walls:       []format.UVTTWall{},
	}

	// Import WITHOUT providing a name - this was causing validation error
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Name:         "", // Empty name - should auto-generate
		ImportTokens: true,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromUVTT(uvttData, opts)
	require.NoError(t, err, "UVTT import should succeed without name")
	require.NotNil(t, gameMap)

	// Verify auto-generated name
	assert.NotEmpty(t, gameMap.Name, "Name should be auto-generated")
	assert.Contains(t, gameMap.Name, "Imported Map", "Auto-generated name should have prefix")
	assert.Equal(t, models.MapTypeBattle, gameMap.Type)
	assert.NotNil(t, gameMap.Grid)
	assert.Equal(t, 20, gameMap.Grid.Width)
	assert.Equal(t, 15, gameMap.Grid.Height)

	t.Logf("Auto-generated name: %s", gameMap.Name)
}

// TestFIX1_UVTTWithProvidedName verifies that UVTT import uses provided name
func TestFIX1_UVTTWithProvidedName(t *testing.T) {
	c := converter.NewMapConverter()

	uvttData := &format.UVTTData{
		Format: 2,
		Resolution: format.UVTTResolution{
			MapSize:        format.UVTTMapSize{X: 10, Y: 10},
			PixelsPerGrid:  70,
		},
	}

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Name:         "My Custom Map Name",
		ImportTokens: false,
		ImportWalls:  false,
	}

	gameMap, err := c.ConvertFromUVTT(uvttData, opts)
	require.NoError(t, err)
	assert.Equal(t, "My Custom Map Name", gameMap.Name)
}

// TestFIX2_FVTTDrawingTypeString verifies that FVTTDrawing.Type can be string
// BUG-M6.5-001 P0-2: drawings[].type can be "t", "ghost" etc. (string), not just int
func TestFIX2_FVTTDrawingTypeString(t *testing.T) {
	// Real data from baileywiki-maps.db shows: "type":"t", "type":"ghost"
	jsonData := `{
		"_id": "drawing-001",
		"type": "t",
		"x": 100,
		"y": 200,
		"width": 300,
		"height": 150.5,
		"fill": "#ff0000"
	}`

	var drawing format.FVTTDrawing
	err := json.Unmarshal([]byte(jsonData), &drawing)
	require.NoError(t, err, "Should parse drawing with string type")

	assert.Equal(t, "drawing-001", drawing.ID)
	assert.Equal(t, "t", drawing.GetTypeString())
	assert.InDelta(t, 100.0, drawing.X, 0.001)
	assert.InDelta(t, 200.0, drawing.Y, 0.001)
	assert.InDelta(t, 300.0, drawing.Width, 0.001)
	assert.InDelta(t, 150.5, drawing.Height, 0.001)
}

// TestFIX2_FVTTDrawingTypeGhost verifies another string type variant
func TestFIX2_FVTTDrawingTypeGhost(t *testing.T) {
	jsonData := `{
		"_id": "drawing-002",
		"type": "ghost",
		"x": 0,
		"y": 0,
		"width": 100,
		"height": 100
	}`

	var drawing format.FVTTDrawing
	err := json.Unmarshal([]byte(jsonData), &drawing)
	require.NoError(t, err, "Should parse drawing with 'ghost' type")

	assert.Equal(t, "ghost", drawing.GetTypeString())
}

// TestFIX3_GlobalLightThresholdFloat64 verifies GlobalLightThreshold as float64
// BUG-M6.5-001 P0-3: globalLightThreshold can be 0.6 (float64), not just int
func TestFIX3_GlobalLightThresholdFloat64(t *testing.T) {
	// Real data from baileywiki-maps.db shows: "globalLightThreshold":0.6
	jsonData := `{
		"_id": "scene-001",
		"name": "Test Scene",
		"width": 2000,
		"height": 2000,
		"grid": 100,
		"globalLight": true,
		"globalLightThreshold": 0.6
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse scene with float64 globalLightThreshold")

	require.NotNil(t, scene.GlobalLightThreshold)
	assert.InDelta(t, 0.6, *scene.GlobalLightThreshold, 0.001)
}

// TestFIX3_GlobalLightThresholdNull verifies null handling
func TestFIX3_GlobalLightThresholdNull(t *testing.T) {
	jsonData := `{
		"_id": "scene-002",
		"name": "Test Scene",
		"width": 2000,
		"height": 2000,
		"grid": 100,
		"globalLightThreshold": null
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err)

	assert.Nil(t, scene.GlobalLightThreshold)
}

// TestFIX4_FVTTDrawingPoints verifies FVTTDrawing.Points as interface{}
// BUG-M6.5-001 P0-4: points can be various formats
func TestFIX4_FVTTDrawingPoints(t *testing.T) {
	jsonData := `{
		"_id": "drawing-003",
		"type": "polygon",
		"x": 0,
		"y": 0,
		"width": 200,
		"height": 200.5,
		"points": [[0, 0], [100, 50], [200, 0]]
	}`

	var drawing format.FVTTDrawing
	err := json.Unmarshal([]byte(jsonData), &drawing)
	require.NoError(t, err, "Should parse drawing with nested array points")

	assert.Equal(t, 200.5, drawing.Height)
	assert.Equal(t, 200, drawing.GetHeightInt())

	points := drawing.GetPointsAsPairs()
	assert.NotNil(t, points)
}

// TestFIX5_FVTTLightBrightFloat64 verifies FVTTLight.Bright/Dim as float64
// BUG-M6.5-001 P0-5: lights[].bright can be 4.26 (float64), not just int
func TestFIX5_FVTTLightBrightFloat64(t *testing.T) {
	// Real data from baileywiki-maps.db shows: "bright":4.26
	jsonData := `{
		"_id": "light-001",
		"x": 500,
		"y": 500,
		"dim": 8.5,
		"bright": 4.26,
		"angle": 360,
		"t": "l",
		"color": "#ffcc00",
		"alpha": 0.8
	}`

	var light format.FVTTLight
	err := json.Unmarshal([]byte(jsonData), &light)
	require.NoError(t, err, "Should parse light with float64 bright/dim")

	assert.Equal(t, "light-001", light.ID)
	assert.InDelta(t, 500.0, light.X, 0.001)
	assert.InDelta(t, 500.0, light.Y, 0.001)
	assert.InDelta(t, 8.5, light.Dim, 0.001)
	assert.InDelta(t, 4.26, light.Bright, 0.001)

	// Test helper methods
	assert.Equal(t, 8, light.GetDimInt())
	assert.Equal(t, 4, light.GetBrightInt())
}

// TestFIX6_FullSceneImport verifies complete scene import with all fixed types
// This is the integration test that uses realistic FVTT Scene data
func TestFIX6_FullSceneImport(t *testing.T) {
	// This JSON contains all the problematic field types from real FVTT data
	jsonData := `{
		"_id": "scene-full-test",
		"name": "Full Test Scene",
		"active": false,
		"navigation": true,
		"navOrder": 0,
		"width": 4000,
		"height": 3000,
		"padding": 0.25,
		"gridType": 1,
		"grid": 100,
		"gridDistance": 5,
		"gridUnits": "ft",
		"tokenVision": true,
		"globalLight": true,
		"globalLightThreshold": 0.6,
		"darkness": 0.5,
		"drawings": [
			{
				"_id": "draw-1",
				"type": "t",
				"x": 100,
				"y": 100,
				"width": 200,
				"height": 150.75,
				"fill": "#ff0000"
			},
			{
				"_id": "draw-2",
				"type": "ghost",
				"x": 0,
				"y": 0,
				"width": 100,
				"height": 100
			}
		],
		"lights": [
			{
				"_id": "light-1",
				"x": 500,
				"y": 500,
				"dim": 12.5,
				"bright": 6.25,
				"angle": 360,
				"t": "l",
				"color": "#ffffff",
				"alpha": 1.0
			}
		],
		"walls": [
			{
				"_id": "wall-1",
				"c": [[0, 0], [1000, 0]],
				"move": 0,
				"sense": 0,
				"door": 0
			}
		],
		"tokens": [],
		"permission": {"default": 0}
	}`

	// Parse the scene
	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse full scene without errors")

	// Verify scene fields
	assert.Equal(t, "scene-full-test", scene.ID)
	assert.Equal(t, "Full Test Scene", scene.Name)
	assert.Equal(t, 4000, scene.Width)
	assert.Equal(t, 3000, scene.Height)
	assert.Equal(t, 100, scene.Grid)

	// Verify globalLightThreshold (FIX-3)
	require.NotNil(t, scene.GlobalLightThreshold)
	assert.InDelta(t, 0.6, *scene.GlobalLightThreshold, 0.001)

	// Verify drawings (FIX-2, FIX-4)
	require.Len(t, scene.Drawings, 2)
	assert.Equal(t, "t", scene.Drawings[0].GetTypeString())
	assert.Equal(t, "ghost", scene.Drawings[1].GetTypeString())
	assert.Equal(t, 150.75, scene.Drawings[0].Height)

	// Verify lights (FIX-5)
	require.Len(t, scene.Lights, 1)
	assert.InDelta(t, 12.5, scene.Lights[0].Dim, 0.001)
	assert.InDelta(t, 6.25, scene.Lights[0].Bright, 0.001)

	// Now convert to Map model
	c := converter.NewMapConverter()
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: true,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert scene to map without errors")
	require.NotNil(t, gameMap)

	assert.Equal(t, "Full Test Scene", gameMap.Name)
	assert.NotNil(t, gameMap.Grid)
	assert.Equal(t, 40, gameMap.Grid.Width)  // 4000 / 100
	assert.Equal(t, 30, gameMap.Grid.Height) // 3000 / 100

	t.Logf("Successfully imported scene '%s' with grid %dx%d", gameMap.Name, gameMap.Grid.Width, gameMap.Grid.Height)
}

// TestFIX6_BatchImportSimulation simulates the batch import scenario
// Verifies that multiple scenes with various type combinations can be imported
func TestFIX6_BatchImportSimulation(t *testing.T) {
	c := converter.NewMapConverter()

	// Simulate multiple scenes with different problematic data combinations
	scenes := []string{
		// Scene 1: float globalLightThreshold
		`{"_id":"s1","name":"Scene 1","width":1000,"height":1000,"grid":100,"globalLightThreshold":0.5}`,
		// Scene 2: string drawing type
		`{"_id":"s2","name":"Scene 2","width":2000,"height":2000,"grid":100,"drawings":[{"_id":"d1","type":"t","x":0,"y":0,"width":100,"height":100}]}`,
		// Scene 3: float light bright/dim
		`{"_id":"s3","name":"Scene 3","width":3000,"height":3000,"grid":100,"lights":[{"_id":"l1","x":100,"y":100,"dim":5.5,"bright":2.75,"angle":360,"t":"l"}]}`,
		// Scene 4: all combined
		`{"_id":"s4","name":"Scene 4","width":4000,"height":4000,"grid":100,"globalLightThreshold":0.8,"drawings":[{"_id":"d2","type":"ghost","x":0,"y":0,"width":50,"height":50}],"lights":[{"_id":"l2","x":200,"y":200,"dim":10.0,"bright":5.0,"angle":360,"t":"l"}]}`,
	}

	opts := format.ImportOptions{
		CampaignID:   "batch-test-campaign",
		ImportTokens: false,
		ImportWalls:  false,
	}

	successCount := 0
	for i, sceneJSON := range scenes {
		var scene format.FVTTScene
		err := json.Unmarshal([]byte(sceneJSON), &scene)
		require.NoError(t, err, "Scene %d should parse", i+1)

		gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
		require.NoError(t, err, "Scene %d should convert", i+1)
		require.NotNil(t, gameMap)

		t.Logf("Scene %d: %s -> Map %dx%d", i+1, scene.Name, gameMap.Grid.Width, gameMap.Grid.Height)
		successCount++
	}

	assert.Equal(t, len(scenes), successCount, "All scenes should import successfully")
	t.Logf("Batch import: %d/%d scenes imported successfully", successCount, len(scenes))
}

// TestBUG_M65_002_FVTTWallMoveSenseMapping verifies FVTT wall move/sense value mapping
// BUG-M6.5-002 P0-1: FVTT uses 0/10/20 for move/sense, model expects 0/1/2
func TestBUG_M65_002_FVTTWallMoveSenseMapping(t *testing.T) {
	c := converter.NewMapConverter()

	// Test scene with walls having different move/sense values
	jsonData := `{
		"_id": "scene-wall-test",
		"name": "Wall Move/Sense Test",
		"width": 2000,
		"height": 2000,
		"grid": 100,
		"walls": [
			{
				"_id": "wall-block",
				"c": [[0, 0], [100, 0]],
				"move": 0,
				"sense": 0,
				"door": 0
			},
			{
				"_id": "wall-difficult",
				"c": [[0, 100], [100, 100]],
				"move": 10,
				"sense": 10,
				"door": 0
			},
			{
				"_id": "wall-allow",
				"c": [[0, 200], [100, 200]],
				"move": 20,
				"sense": 20,
				"door": 0
			},
			{
				"_id": "wall-high-allow",
				"c": [[0, 300], [100, 300]],
				"move": 30,
				"sense": 30,
				"door": 0
			}
		]
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse scene with various wall move/sense values")

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: false,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert scene without errors")
	require.NotNil(t, gameMap)

	// Verify walls were converted
	require.Equal(t, 4, len(gameMap.Walls), "Should have 4 walls")

	// Test the mapping of move/sense values
	// Wall values in model: 0=block, 1=difficult/limited, 2=allow
	walls := gameMap.Walls

	// Find and verify each wall
	wallMap := make(map[string]*models.Wall)
	for _, w := range walls {
		wallMap[w.ID] = w
	}

	// Wall with move=0, sense=0 should map to 0, 0 (blocks)
	if wall, ok := wallMap["wall-block"]; ok {
		assert.Equal(t, 0, wall.Move, "move=0 should map to 0 (blocks)")
		assert.Equal(t, 0, wall.Sense, "sense=0 should map to 0 (blocks)")
	} else {
		t.Error("wall-block not found")
	}

	// Wall with move=10, sense=10 should map to 1, 1 (difficult/limited)
	if wall, ok := wallMap["wall-difficult"]; ok {
		assert.Equal(t, 1, wall.Move, "move=10 should map to 1 (difficult)")
		assert.Equal(t, 1, wall.Sense, "sense=10 should map to 1 (limited)")
	} else {
		t.Error("wall-difficult not found")
	}

	// Wall with move=20, sense=20 should map to 2, 2 (allows)
	if wall, ok := wallMap["wall-allow"]; ok {
		assert.Equal(t, 2, wall.Move, "move=20 should map to 2 (allows)")
		assert.Equal(t, 2, wall.Sense, "sense=20 should map to 2 (allows)")
	} else {
		t.Error("wall-allow not found")
	}

	// Wall with move=30, sense=30 should also map to 2, 2 (allows)
	if wall, ok := wallMap["wall-high-allow"]; ok {
		assert.Equal(t, 2, wall.Move, "move=30 should map to 2 (allows)")
		assert.Equal(t, 2, wall.Sense, "sense=30 should map to 2 (allows)")
	} else {
		t.Error("wall-high-allow not found")
	}

	t.Logf("Wall move/sense mapping verified for all 4 walls")
}

// TestBUG_M65_002_FVTTTokenCharacterID verifies FVTT token CharacterID setting
// BUG-M6.5-002 P0-2: Token CharacterID should be set from ActorID
func TestBUG_M65_002_FVTTTokenCharacterID(t *testing.T) {
	c := converter.NewMapConverter()

	// Test scene with tokens - some with ActorID, some without
	jsonData := `{
		"_id": "scene-token-test",
		"name": "Token CharacterID Test",
		"width": 2000,
		"height": 2000,
		"grid": 100,
		"tokens": [
			{
				"_id": "token-with-actor",
				"name": "Goblin Warrior",
				"actorId": "actor-goblin-001",
				"x": 500,
				"y": 500,
				"width": 100,
				"height": 100
			},
			{
				"_id": "token-no-actor",
				"name": "Placeholder Token",
				"actorId": "",
				"x": 700,
				"y": 500,
				"width": 100,
				"height": 100
			},
			{
				"_id": "",
				"name": "Token No ID",
				"actorId": "actor-noid-001",
				"x": 900,
				"y": 500,
				"width": 100,
				"height": 100
			}
		]
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse scene with tokens")

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: true,
		ImportWalls:  false,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert scene without errors")
	require.NotNil(t, gameMap)

	// Verify tokens were converted
	tokens := gameMap.Tokens
	require.Len(t, tokens, 3, "Should have 3 tokens")

	// Create a map for easier lookup
	tokenMap := make(map[string]models.Token)
	for _, token := range tokens {
		tokenMap[token.Name] = token
	}

	// Token with ActorID should use ActorID as CharacterID
	if token, ok := tokenMap["Goblin Warrior"]; ok {
		assert.Equal(t, "actor-goblin-001", token.CharacterID, "Token with ActorID should have CharacterID set from ActorID")
		assert.NotEmpty(t, token.ID, "Token should have ID")
		t.Logf("Token 'Goblin Warrior': ID=%s, CharacterID=%s", token.ID, token.CharacterID)
	} else {
		t.Error("Token 'Goblin Warrior' not found")
	}

	// Token without ActorID should use token ID as CharacterID
	if token, ok := tokenMap["Placeholder Token"]; ok {
		assert.Equal(t, "token-no-actor", token.CharacterID, "Token without ActorID should use token ID as CharacterID")
		assert.Equal(t, "token-no-actor", token.ID, "Token ID should match")
		t.Logf("Token 'Placeholder Token': ID=%s, CharacterID=%s", token.ID, token.CharacterID)
	} else {
		t.Error("Token 'Placeholder Token' not found")
	}

	// Token with no token ID but with ActorID should use ActorID
	if token, ok := tokenMap["Token No ID"]; ok {
		assert.Equal(t, "actor-noid-001", token.CharacterID, "Token with ActorID but no token ID should use ActorID")
		assert.NotEmpty(t, token.ID, "Token should have auto-generated ID")
		t.Logf("Token 'Token No ID': ID=%s, CharacterID=%s", token.ID, token.CharacterID)
	} else {
		t.Error("Token 'Token No ID' not found")
	}

	t.Logf("Token CharacterID setting verified for all 3 tokens")
}

// TestBUG_M65_002_Integration verifies both fixes work together
func TestBUG_M65_002_Integration(t *testing.T) {
	c := converter.NewMapConverter()

	// Complete scene with walls and tokens
	jsonData := `{
		"_id": "scene-integration-test",
		"name": "Integration Test Scene",
		"width": 3000,
		"height": 3000,
		"grid": 100,
		"gridDistance": 5,
		"gridUnits": "ft",
		"walls": [
			{
				"_id": "wall-1",
				"c": [[0, 0], [1000, 0]],
				"move": 0,
				"sense": 0,
				"door": 0
			},
			{
				"_id": "door-1",
				"c": [[500, 0], [600, 0]],
				"move": 10,
				"sense": 10,
				"door": 1,
				"ds": 0
			}
		],
		"tokens": [
			{
				"_id": "hero-1",
				"name": "Hero Character",
				"actorId": "actor-hero-001",
				"x": 1500,
				"y": 1500,
				"width": 100,
				"height": 100
			}
		]
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err)

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: true,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err)
	require.NotNil(t, gameMap)

	// Verify walls
	assert.Equal(t, 2, len(gameMap.Walls), "Should have 2 walls")

	// Verify tokens
	tokens := gameMap.Tokens
	require.Len(t, tokens, 1, "Should have 1 token")
	assert.Equal(t, "actor-hero-001", tokens[0].CharacterID, "Token should have CharacterID from ActorID")

	t.Logf("Integration test passed: walls=%d, tokens=%d", len(gameMap.Walls), len(tokens))
}

// TestFIX7_FVTTDrawingFloatCoords verifies FVTTDrawing coordinates as float64
// BUG-M6.5-003 P0-1: drawings[].x/y/width can be float64 (e.g., 1234.56)
func TestFIX7_FVTTDrawingFloatCoords(t *testing.T) {
	// Real data from Baileywiki Maps 39 scene
	jsonData := `{
		"_id": "drawing-float-coords",
		"type": "t",
		"x": 1234.56,
		"y": 2345.67,
		"width": 637.5,
		"height": 425.25
	}`

	var drawing format.FVTTDrawing
	err := json.Unmarshal([]byte(jsonData), &drawing)
	require.NoError(t, err, "Should parse drawing with float64 coordinates")

	assert.Equal(t, "drawing-float-coords", drawing.ID)
	assert.InDelta(t, 1234.56, drawing.X, 0.001)
	assert.InDelta(t, 2345.67, drawing.Y, 0.001)
	assert.InDelta(t, 637.5, drawing.Width, 0.001)
	assert.InDelta(t, 425.25, drawing.Height, 0.001)

	// Test helper methods
	assert.Equal(t, 1234, drawing.GetXInt())
	assert.Equal(t, 2345, drawing.GetYInt())
	assert.Equal(t, 637, drawing.GetWidthInt())
	assert.Equal(t, 425, drawing.GetHeightInt())

	t.Logf("Drawing float coords: x=%d, y=%d, width=%d, height=%d",
		drawing.GetXInt(), drawing.GetYInt(), drawing.GetWidthInt(), drawing.GetHeightInt())
}

// TestFIX8_FVTTLightFloatCoords verifies FVTTLight coordinates as float64
// BUG-M6.5-003 P0-2: lights[].x/y can be float64 (e.g., 3971.6345)
func TestFIX8_FVTTLightFloatCoords(t *testing.T) {
	// Real data from Baileywiki Maps 39 scene
	jsonData := `{
		"_id": "light-float-coords",
		"x": 3971.6345,
		"y": 2847.2134,
		"dim": 15.75,
		"bright": 7.875,
		"angle": 360,
		"t": "l",
		"color": "#ffcc00",
		"alpha": 0.5
	}`

	var light format.FVTTLight
	err := json.Unmarshal([]byte(jsonData), &light)
	require.NoError(t, err, "Should parse light with float64 coordinates")

	assert.Equal(t, "light-float-coords", light.ID)
	assert.InDelta(t, 3971.6345, light.X, 0.0001)
	assert.InDelta(t, 2847.2134, light.Y, 0.0001)
	assert.InDelta(t, 15.75, light.Dim, 0.001)
	assert.InDelta(t, 7.875, light.Bright, 0.001)

	// Test helper methods
	assert.Equal(t, 3971, light.GetXInt())
	assert.Equal(t, 2847, light.GetYInt())
	assert.Equal(t, 15, light.GetDimInt())
	assert.Equal(t, 7, light.GetBrightInt())

	t.Logf("Light float coords: x=%d, y=%d, dim=%d, bright=%d",
		light.GetXInt(), light.GetYInt(), light.GetDimInt(), light.GetBrightInt())
}

// TestFIX9_FVTTSoundFloatVolume verifies FVTTSound volume as float64
// BUG-M6.5-003 P0-3: sounds[].volume can be float64 (e.g., 0.85)
func TestFIX9_FVTTSoundFloatVolume(t *testing.T) {
	// Real data from Baileywiki Maps 39 scene
	jsonData := `{
		"_id": "sound-float-volume",
		"x": 1000,
		"y": 1000,
		"path": "sounds/ambient.mp3",
		"repeat": true,
		"volume": 0.85,
		"echo": false
	}`

	var sound format.FVTTSound
	err := json.Unmarshal([]byte(jsonData), &sound)
	require.NoError(t, err, "Should parse sound with float64 volume")

	assert.Equal(t, "sound-float-volume", sound.ID)
	assert.InDelta(t, 0.85, sound.Volume, 0.001)

	// Test helper method - volume should be 0-100 range
	assert.Equal(t, 85, sound.GetVolumeInt())

	t.Logf("Sound float volume: %f -> %d", sound.Volume, sound.GetVolumeInt())
}

// TestFIX9_FVTTSoundVolumeEdgeCases tests edge cases for volume conversion
func TestFIX9_FVTTSoundVolumeEdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		volume         float64
		expectedVolume int
	}{
		{"zero volume", 0.0, 0},
		{"half volume", 0.5, 50},
		{"full volume", 1.0, 100},
		{"quiet", 0.25, 25},
		{"loud", 0.9, 90},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sound := format.FVTTSound{
				ID:     "test-" + tc.name,
				Volume: tc.volume,
			}
			assert.Equal(t, tc.expectedVolume, sound.GetVolumeInt(),
				"Volume %f should convert to %d", tc.volume, tc.expectedVolume)
		})
	}
}

// TestFIX10_FVTTWallFloatCoords verifies FVTTWall coordinates as float64
// BUG-M6.5-003 P0-4: walls[].c can have float64 coordinates (e.g., [[140.5, 140.0], [1260.75, 140.0]])
func TestFIX10_FVTTWallFloatCoords(t *testing.T) {
	// Real data from Baileywiki Maps 39 scene
	jsonData := `{
		"_id": "wall-float-coords",
		"c": [[140.5, 140.0], [1260.75, 140.0]],
		"move": 0,
		"sense": 0,
		"door": 0
	}`

	var wall format.FVTTWall
	err := json.Unmarshal([]byte(jsonData), &wall)
	require.NoError(t, err, "Should parse wall with float64 coordinates")

	assert.Equal(t, "wall-float-coords", wall.ID)
	require.Len(t, wall.C, 2, "Wall should have 2 coordinate pairs")
	require.Len(t, wall.C[0], 2, "First coordinate pair should have 2 values")
	require.Len(t, wall.C[1], 2, "Second coordinate pair should have 2 values")

	// Verify coordinates
	assert.InDelta(t, 140.5, wall.C[0][0], 0.001)
	assert.InDelta(t, 140.0, wall.C[0][1], 0.001)
	assert.InDelta(t, 1260.75, wall.C[1][0], 0.001)
	assert.InDelta(t, 140.0, wall.C[1][1], 0.001)

	t.Logf("Wall float coords: [[%.2f, %.2f], [%.2f, %.2f]]",
		wall.C[0][0], wall.C[0][1], wall.C[1][0], wall.C[1][1])
}

// TestFIX10_FVTTWallConversionWithFloatCoords verifies wall conversion with float coordinates
func TestFIX10_FVTTWallConversionWithFloatCoords(t *testing.T) {
	c := converter.NewMapConverter()

	// Test scene with walls having float coordinates
	jsonData := `{
		"_id": "scene-float-wall-test",
		"name": "Float Wall Test",
		"width": 2000,
		"height": 2000,
		"grid": 70,
		"walls": [
			{
				"_id": "wall-1",
				"c": [[140.5, 140.0], [1260.75, 140.0]],
				"move": 0,
				"sense": 0,
				"door": 0
			},
			{
				"_id": "wall-2",
				"c": [[0.0, 0.0], [70.0, 0.0]],
				"move": 20,
				"sense": 20,
				"door": 0
			},
			{
				"_id": "door-1",
				"c": [[350.5, 700.25], [420.75, 700.25]],
				"move": 10,
				"sense": 10,
				"door": 1,
				"ds": 0
			}
		]
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse scene with float coordinate walls")

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: false,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert scene with float coordinate walls")
	require.NotNil(t, gameMap)

	// Verify walls were converted
	walls := gameMap.Walls
	require.Len(t, walls, 3, "Should have 3 walls")

	// Create a map for easier lookup
	wallMap := make(map[string]*models.Wall)
	for _, w := range walls {
		wallMap[w.ID] = w
	}

	// Verify wall-1: [140.5, 140.0] -> [1260.75, 140.0] with grid 70
	// Expected grid coords: x1=2, y1=2, x2=18, y2=2
	if wall, ok := wallMap["wall-1"]; ok {
		assert.Equal(t, 2, wall.Bounds[0], "wall-1 X1 should be 2 (140.5/70)")
		assert.Equal(t, 2, wall.Bounds[1], "wall-1 Y1 should be 2 (140.0/70)")
		assert.Equal(t, 18, wall.Bounds[2], "wall-1 X2 should be 18 (1260.75/70)")
		assert.Equal(t, 2, wall.Bounds[3], "wall-1 Y2 should be 2 (140.0/70)")
		assert.Equal(t, 0, wall.Move, "wall-1 should block movement")
		assert.Equal(t, 0, wall.Sense, "wall-1 should block sense")
		t.Logf("wall-1: grid coords [%d,%d] -> [%d,%d]", wall.Bounds[0], wall.Bounds[1], wall.Bounds[2], wall.Bounds[3])
	} else {
		t.Error("wall-1 not found")
	}

	// Verify wall-2: [0.0, 0.0] -> [70.0, 0.0] with grid 70
	// Expected grid coords: x1=0, y1=0, x2=1, y2=0
	if wall, ok := wallMap["wall-2"]; ok {
		assert.Equal(t, 0, wall.Bounds[0], "wall-2 X1 should be 0")
		assert.Equal(t, 0, wall.Bounds[1], "wall-2 Y1 should be 0")
		assert.Equal(t, 1, wall.Bounds[2], "wall-2 X2 should be 1 (70.0/70)")
		assert.Equal(t, 0, wall.Bounds[3], "wall-2 Y2 should be 0")
		assert.Equal(t, 2, wall.Move, "wall-2 should allow movement")
		assert.Equal(t, 2, wall.Sense, "wall-2 should allow sense")
	} else {
		t.Error("wall-2 not found")
	}

	// Verify door-1: [350.5, 700.25] -> [420.75, 700.25] with grid 70
	// Expected grid coords: x1=5, y1=10, x2=6, y2=10
	if wall, ok := wallMap["door-1"]; ok {
		assert.Equal(t, 5, wall.Bounds[0], "door-1 X1 should be 5 (350.5/70)")
		assert.Equal(t, 10, wall.Bounds[1], "door-1 Y1 should be 10 (700.25/70)")
		assert.Equal(t, 6, wall.Bounds[2], "door-1 X2 should be 6 (420.75/70)")
		assert.Equal(t, 10, wall.Bounds[3], "door-1 Y2 should be 10 (700.25/70)")
		assert.NotNil(t, wall.Door, "door-1 should have door data")
		assert.Equal(t, models.DoorStateClosed, wall.Door.State, "door-1 should be closed")
		t.Logf("door-1: grid coords [%d,%d] -> [%d,%d], door state=%s",
			wall.Bounds[0], wall.Bounds[1], wall.Bounds[2], wall.Bounds[3], wall.Door.State)
	} else {
		t.Error("door-1 not found")
	}

	t.Logf("Float coordinate wall conversion verified for all 3 walls")
}

// TestBUG_M65_003_RealisticFloatScene tests realistic float data from actual FVTT module
// BUG-M6.5-003: Tests all float type fixes with real-world data from Baileywiki Maps 39
func TestBUG_M65_003_RealisticFloatScene(t *testing.T) {
	// Load the realistic float scene from testdata
	sceneData, err := os.ReadFile("../../testdata/maps/realistic_float_scene.json")
	require.NoError(t, err, "Should be able to read realistic_float_scene.json")

	// Parse the scene
	var scene format.FVTTScene
	err = json.Unmarshal(sceneData, &scene)
	require.NoError(t, err, "Should parse realistic float scene without errors")

	// Verify scene metadata
	assert.Equal(t, "realistic-float-scene-001", scene.ID)
	assert.Equal(t, "Realistic Float Test Scene (Baileywiki Maps 39 Data)", scene.Name)
	assert.Equal(t, 4375, scene.Width)
	assert.Equal(t, 3281, scene.Height)
	assert.Equal(t, 70, scene.Grid)

	// Verify drawings with float coordinates (BUG-M6.5-003 P0-1)
	require.Len(t, scene.Drawings, 2, "Should have 2 drawings")

	// Drawing 1: x=1234.56, y=2345.67, width=637.5, height=425.25
	assert.Equal(t, "drawing-float-001", scene.Drawings[0].ID)
	assert.InDelta(t, 1234.56, scene.Drawings[0].X, 0.001)
	assert.InDelta(t, 2345.67, scene.Drawings[0].Y, 0.001)
	assert.InDelta(t, 637.5, scene.Drawings[0].Width, 0.001)
	assert.InDelta(t, 425.25, scene.Drawings[0].Height, 0.001)
	assert.Equal(t, 1234, scene.Drawings[0].GetXInt())
	assert.Equal(t, 2345, scene.Drawings[0].GetYInt())
	assert.Equal(t, 637, scene.Drawings[0].GetWidthInt())
	assert.Equal(t, 425, scene.Drawings[0].GetHeightInt())

	// Drawing 2: x=500.75, y=750.25, width=200.5, height=150.75
	assert.Equal(t, "drawing-float-002", scene.Drawings[1].ID)
	assert.InDelta(t, 500.75, scene.Drawings[1].X, 0.001)
	assert.InDelta(t, 750.25, scene.Drawings[1].Y, 0.001)
	assert.InDelta(t, 200.5, scene.Drawings[1].Width, 0.001)
	assert.InDelta(t, 150.75, scene.Drawings[1].Height, 0.001)

	// Verify lights with float coordinates (BUG-M6.5-003 P0-2)
	require.Len(t, scene.Lights, 2, "Should have 2 lights")

	// Light 1: x=3971.63, y=2847.21, dim=15.75, bright=7.875
	assert.Equal(t, "light-float-001", scene.Lights[0].ID)
	assert.InDelta(t, 3971.63, scene.Lights[0].X, 0.001)
	assert.InDelta(t, 2847.21, scene.Lights[0].Y, 0.001)
	assert.InDelta(t, 15.75, scene.Lights[0].Dim, 0.001)
	assert.InDelta(t, 7.875, scene.Lights[0].Bright, 0.001)
	assert.Equal(t, 3971, scene.Lights[0].GetXInt())
	assert.Equal(t, 2847, scene.Lights[0].GetYInt())
	assert.Equal(t, 15, scene.Lights[0].GetDimInt())
	assert.Equal(t, 7, scene.Lights[0].GetBrightInt())

	// Light 2: x=1050.5, y=700.25, dim=10.5, bright=5.25
	assert.Equal(t, "light-float-002", scene.Lights[1].ID)
	assert.InDelta(t, 1050.5, scene.Lights[1].X, 0.001)
	assert.InDelta(t, 700.25, scene.Lights[1].Y, 0.001)

	// Verify sounds with float volume (BUG-M6.5-003 P0-3)
	require.Len(t, scene.Sounds, 2, "Should have 2 sounds")

	// Sound 1: volume=0.85
	assert.Equal(t, "sound-float-001", scene.Sounds[0].ID)
	assert.InDelta(t, 0.85, scene.Sounds[0].Volume, 0.001)
	assert.Equal(t, 85, scene.Sounds[0].GetVolumeInt())

	// Sound 2: volume=0.65
	assert.Equal(t, "sound-float-002", scene.Sounds[1].ID)
	assert.InDelta(t, 0.65, scene.Sounds[1].Volume, 0.001)
	assert.Equal(t, 65, scene.Sounds[1].GetVolumeInt())

	// Verify walls with float coordinates (BUG-M6.5-003 P0-4)
	require.Len(t, scene.Walls, 4, "Should have 4 walls")

	// Wall 1: c=[[140.5, 140.0], [1260.75, 140.0]]
	assert.Equal(t, "wall-float-001", scene.Walls[0].ID)
	require.Len(t, scene.Walls[0].C, 2, "Wall should have 2 coordinate pairs")
	assert.InDelta(t, 140.5, scene.Walls[0].C[0][0], 0.001)
	assert.InDelta(t, 140.0, scene.Walls[0].C[0][1], 0.001)
	assert.InDelta(t, 1260.75, scene.Walls[0].C[1][0], 0.001)
	assert.InDelta(t, 140.0, scene.Walls[0].C[1][1], 0.001)

	// Wall 2: c=[[140.0, 140.5], [140.0, 1260.75]]
	assert.Equal(t, "wall-float-002", scene.Walls[1].ID)
	assert.InDelta(t, 140.0, scene.Walls[1].C[0][0], 0.001)
	assert.InDelta(t, 140.5, scene.Walls[1].C[0][1], 0.001)
	assert.InDelta(t, 140.0, scene.Walls[1].C[1][0], 0.001)
	assert.InDelta(t, 1260.75, scene.Walls[1].C[1][1], 0.001)

	// Door: c=[[350.5, 700.25], [420.75, 700.25]]
	assert.Equal(t, "door-float-001", scene.Walls[2].ID)
	assert.InDelta(t, 350.5, scene.Walls[2].C[0][0], 0.001)
	assert.InDelta(t, 700.25, scene.Walls[2].C[0][1], 0.001)
	assert.InDelta(t, 420.75, scene.Walls[2].C[1][0], 0.001)
	assert.InDelta(t, 700.25, scene.Walls[2].C[1][1], 0.001)

	// Secret door: c=[[2100.25, 1400.5], [2100.75, 1540.0]]
	assert.Equal(t, "wall-secret-001", scene.Walls[3].ID)
	assert.InDelta(t, 2100.25, scene.Walls[3].C[0][0], 0.001)
	assert.InDelta(t, 1400.5, scene.Walls[3].C[0][1], 0.001)
	assert.InDelta(t, 2100.75, scene.Walls[3].C[1][0], 0.001)
	assert.InDelta(t, 1540.0, scene.Walls[3].C[1][1], 0.001)

	t.Logf("Realistic float scene parsed successfully:")
	t.Logf("  Scene: %s (%dx%d)", scene.Name, scene.Width, scene.Height)
	t.Logf("  Drawings: %d (all with float coordinates)", len(scene.Drawings))
	t.Logf("  Lights: %d (all with float coordinates)", len(scene.Lights))
	t.Logf("  Sounds: %d (all with float volume)", len(scene.Sounds))
	t.Logf("  Walls: %d (all with float coordinates)", len(scene.Walls))
}

// TestBUG_M65_003_ConversionWithRealisticFloatData tests conversion with realistic float data
func TestBUG_M65_003_ConversionWithRealisticFloatData(t *testing.T) {
	// Load the realistic float scene from testdata
	sceneData, err := os.ReadFile("../../testdata/maps/realistic_float_scene.json")
	require.NoError(t, err, "Should be able to read realistic_float_scene.json")

	var scene format.FVTTScene
	err = json.Unmarshal(sceneData, &scene)
	require.NoError(t, err, "Should parse realistic float scene without errors")

	// Convert to Map model
	c := converter.NewMapConverter()
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: true,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert realistic float scene without errors")
	require.NotNil(t, gameMap)

	// Verify basic map info
	assert.Equal(t, "Realistic Float Test Scene (Baileywiki Maps 39 Data)", gameMap.Name)
	assert.NotNil(t, gameMap.Grid)
	assert.Equal(t, 62, gameMap.Grid.Width)  // 4375 / 70 = 62.46 -> 62
	assert.Equal(t, 46, gameMap.Grid.Height) // 3281 / 70 = 46.87 -> 46

	// Verify tokens were converted
	require.Len(t, gameMap.Tokens, 1, "Should have 1 token")
	assert.Equal(t, "Goblin Warrior", gameMap.Tokens[0].Name)
	assert.Equal(t, "actor-goblin-001", gameMap.Tokens[0].CharacterID)

	// Verify walls were converted with correct grid coordinates
	walls := gameMap.Walls
	require.Len(t, walls, 4, "Should have 4 walls")

	// Create a map for easier lookup
	wallMap := make(map[string]*models.Wall)
	for _, w := range walls {
		wallMap[w.ID] = w
	}

	// Verify wall-float-001: [140.5, 140.0] -> [1260.75, 140.0] with grid 70
	// Expected grid coords: x1=2, y1=2, x2=18, y2=2
	if wall, ok := wallMap["wall-float-001"]; ok {
		assert.Equal(t, 2, wall.Bounds[0], "wall-float-001 X1 should be 2 (140.5/70)")
		assert.Equal(t, 2, wall.Bounds[1], "wall-float-001 Y1 should be 2 (140.0/70)")
		assert.Equal(t, 18, wall.Bounds[2], "wall-float-001 X2 should be 18 (1260.75/70)")
		assert.Equal(t, 2, wall.Bounds[3], "wall-float-001 Y2 should be 2")
	}

	// Verify door-float-001: [350.5, 700.25] -> [420.75, 700.25] with grid 70
	// Expected grid coords: x1=5, y1=10, x2=6, y2=10
	if door, ok := wallMap["door-float-001"]; ok {
		assert.Equal(t, 5, door.Bounds[0], "door-float-001 X1 should be 5 (350.5/70)")
		assert.Equal(t, 10, door.Bounds[1], "door-float-001 Y1 should be 10 (700.25/70)")
		assert.Equal(t, 6, door.Bounds[2], "door-float-001 X2 should be 6 (420.75/70)")
		assert.Equal(t, 10, door.Bounds[3], "door-float-001 Y2 should be 10")
		assert.NotNil(t, door.Door, "door-float-001 should have door data")
	}

	// Verify wall-secret-001: [2100.25, 1400.5] -> [2100.75, 1540.0] with grid 70
	// Expected grid coords: x1=30, y1=20, x2=30, y2=22
	if wall, ok := wallMap["wall-secret-001"]; ok {
		assert.Equal(t, 30, wall.Bounds[0], "wall-secret-001 X1 should be 30")
		assert.Equal(t, 20, wall.Bounds[1], "wall-secret-001 Y1 should be 20")
		assert.Equal(t, 30, wall.Bounds[2], "wall-secret-001 X2 should be 30")
		assert.Equal(t, 22, wall.Bounds[3], "wall-secret-001 Y2 should be 22")
	}

	t.Logf("Realistic float scene conversion successful:")
	t.Logf("  Map: %s (grid %dx%d)", gameMap.Name, gameMap.Grid.Width, gameMap.Grid.Height)
	t.Logf("  Tokens: %d", len(gameMap.Tokens))
	t.Logf("  Walls: %d", len(gameMap.Walls))
}

// TestBUG_M65_003_Integration verifies all float type fixes work together
func TestBUG_M65_003_Integration(t *testing.T) {
	c := converter.NewMapConverter()

	// Complete scene with all float type issues from real data
	jsonData := `{
		"_id": "scene-float-integration",
		"name": "Float Type Integration Test",
		"width": 4000,
		"height": 3000,
		"grid": 70,
		"drawings": [
			{
				"_id": "draw-1",
				"type": "t",
				"x": 1234.56,
				"y": 2345.67,
				"width": 637.5,
				"height": 425.25
			}
		],
		"lights": [
			{
				"_id": "light-1",
				"x": 3971.6345,
				"y": 2847.2134,
				"dim": 15.75,
				"bright": 7.875,
				"angle": 360,
				"t": "l",
				"color": "#ffcc00",
				"alpha": 0.5
			}
		],
		"sounds": [
			{
				"_id": "sound-1",
				"x": 1000,
				"y": 1000,
				"path": "sounds/ambient.mp3",
				"repeat": true,
				"volume": 0.85,
				"echo": false
			}
		],
		"walls": [
			{
				"_id": "wall-1",
				"c": [[140.5, 140.0], [1260.75, 140.0]],
				"move": 0,
				"sense": 0,
				"door": 0
			}
		],
		"tokens": []
	}`

	var scene format.FVTTScene
	err := json.Unmarshal([]byte(jsonData), &scene)
	require.NoError(t, err, "Should parse scene with all float types")

	// Verify all types parsed correctly
	assert.InDelta(t, 1234.56, scene.Drawings[0].X, 0.001)
	assert.InDelta(t, 3971.6345, scene.Lights[0].X, 0.0001)
	assert.InDelta(t, 0.85, scene.Sounds[0].Volume, 0.001)
	assert.InDelta(t, 140.5, scene.Walls[0].C[0][0], 0.001)

	// Convert to map
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: false,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromFVTTScene(&scene, opts)
	require.NoError(t, err, "Should convert scene with all float types")
	require.NotNil(t, gameMap)

	assert.Equal(t, "Float Type Integration Test", gameMap.Name)
	assert.NotNil(t, gameMap.Grid)
	assert.Equal(t, 57, gameMap.Grid.Width)  // 4000 / 70
	assert.Equal(t, 42, gameMap.Grid.Height) // 3000 / 70

	t.Logf("Float type integration test passed: grid=%dx%d, walls=%d",
		gameMap.Grid.Width, gameMap.Grid.Height, len(gameMap.Walls))
}
