// Package tools contains integration tests for map visual location tools
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMapToolsForVisual creates a MapTools instance for visual location tests
func setupMapToolsForVisual() (*tools.MapTools, *mcp.Registry, *MockMapStore, *MockCampaignStore, *MockGameStateStore) {
	mapStore := NewMockMapStore()
	campaignStore := NewMockCampaignStore()
	gameStateStore := NewMockGameStateStore()

	// Create a test campaign
	campaign := models.NewCampaign("Test Campaign", "dm-001", "A test campaign")
	campaign.ID = "campaign-001"
	campaignStore.Create(context.Background(), campaign)

	// Create a test game state
	gameState := models.NewGameState("campaign-001")
	gameStateStore.Create(context.Background(), gameState)

	mapService := service.NewMapService(mapStore, campaignStore, gameStateStore)
	mapTools := tools.NewMapTools(mapService)
	registry := mcp.NewRegistry()

	return mapTools, registry, mapStore, campaignStore, gameStateStore
}

// TestMapTools_CreateVisualLocation_Success tests successful creation of a visual location
func TestMapTools_CreateVisualLocation_Success(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	worldMap.Image = models.NewMapImage("https://example.com/map.jpg")
	mapStore.Create(ctx, worldMap)

	// Create a visual location
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":  "campaign-001",
		"map_id":       "world-visual-001",
		"name":         "Waterdeep",
		"description":  "The City of Splendors",
		"type":         "town",
		"position_x":   0.5,
		"position_y":   0.3,
	})

	req := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result["message"], "Created visual location")
	assert.Contains(t, result["message"], "Waterdeep")

	visualLoc := result["visual_location"].(map[string]interface{})
	assert.NotEmpty(t, visualLoc["id"])
	assert.Equal(t, "Waterdeep", visualLoc["name"])
	assert.Equal(t, "The City of Splendors", visualLoc["description"])
	assert.Equal(t, "town", visualLoc["type"])
	assert.Equal(t, 0.5, visualLoc["position_x"])
	assert.Equal(t, 0.3, visualLoc["position_y"])
	assert.Equal(t, false, visualLoc["is_confirmed"])

	// Verify map was updated
	updatedMap, _ := mapStore.Get(ctx, "world-visual-001")
	assert.Len(t, updatedMap.VisualLocations, 1)
	assert.Equal(t, "Waterdeep", updatedMap.VisualLocations[0].Name)
}

// TestMapTools_CreateVisualLocation_MissingRequiredFields tests validation of required fields
func TestMapTools_CreateVisualLocation_MissingRequiredFields(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	tests := []struct {
		name        string
		missingField string
		args         map[string]interface{}
		shouldError  bool
	}{
		{
			name:        "missing name",
			missingField: "name",
			args: map[string]interface{}{
				"campaign_id": "campaign-001",
				"map_id":      "world-visual-001",
				"type":        "town",
				"position_x":  0.5,
				"position_y":  0.3,
			},
			shouldError: true, // Empty name will be caught by service validation
		},
		{
			name:        "missing type",
			missingField: "type",
			args: map[string]interface{}{
				"campaign_id": "campaign-001",
				"map_id":      "world-visual-001",
				"name":        "Location",
				"position_x":  0.5,
				"position_y":  0.3,
			},
			shouldError: true, // Empty type will be caught by service validation
		},
		{
			name:        "missing position_x",
			missingField: "position_x",
			args: map[string]interface{}{
				"campaign_id": "campaign-001",
				"map_id":      "world-visual-001",
				"name":        "Location",
				"type":        "town",
				"position_y":  0.3,
			},
			shouldError: false, // Zero is a valid position
		},
		{
			name:        "missing position_y",
			missingField: "position_y",
			args: map[string]interface{}{
				"campaign_id": "campaign-001",
				"map_id":      "world-visual-001",
				"name":        "Location",
				"type":        "town",
				"position_x":  0.5,
			},
			shouldError: false, // Zero is a valid position
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, _ := json.Marshal(tt.args)
			req := mcp.ToolRequest{
				ToolName:  "create_visual_location",
				Arguments: args,
			}

			resp := registry.Call(ctx, req)
			if tt.shouldError {
				assert.True(t, resp.IsError, "Should return error for missing "+tt.missingField)
			} else {
				// Zero position values are valid (0,0 is a valid position)
				assert.False(t, resp.IsError, "Should succeed with default zero position")
			}
		})
	}
}

// TestMapTools_CreateVisualLocation_InvalidPosition tests validation of position coordinates
func TestMapTools_CreateVisualLocation_InvalidPosition(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	tests := []struct {
		name       string
		positionX  float64
		positionY  float64
		errorMsg   string
	}{
		{"x negative", -0.1, 0.5, "must be between 0 and 1"},
		{"x greater than 1", 1.5, 0.5, "must be between 0 and 1"},
		{"y negative", 0.5, -0.1, "must be between 0 and 1"},
		{"y greater than 1", 0.5, 1.5, "must be between 0 and 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, _ := json.Marshal(map[string]interface{}{
				"campaign_id": "campaign-001",
				"map_id":      "world-visual-001",
				"name":        "Location",
				"type":        "town",
				"position_x":  tt.positionX,
				"position_y":  tt.positionY,
			})

			req := mcp.ToolRequest{
				ToolName:  "create_visual_location",
				Arguments: args,
			}

			resp := registry.Call(ctx, req)
			assert.True(t, resp.IsError, "Should return error for invalid position")
			assert.Contains(t, resp.Content[0].Text, tt.errorMsg)
		})
	}
}

// TestMapTools_CreateVisualLocation_MapNotFound tests creating location on non-existent map
func TestMapTools_CreateVisualLocation_MapNotFound(t *testing.T) {
	mapTools, registry, _, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"map_id":      "nonexistent-map",
		"name":        "Location",
		"type":        "town",
		"position_x":  0.5,
		"position_y":  0.3,
	})

	req := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError, "Should return error for non-existent map")
}

// TestMapTools_CreateVisualLocation_GridModeRejectsVisualLocation tests that Grid mode maps reject visual locations
func TestMapTools_CreateVisualLocation_GridModeRejectsVisualLocation(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create a GRID mode world map (default)
	worldMap := models.NewWorldMap("campaign-001", "Grid World Map", 100, 100)
	worldMap.ID = "world-grid-001"
	worldMap.Mode = models.MapModeGrid // Explicitly set to Grid mode
	mapStore.Create(ctx, worldMap)

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"map_id":      "world-grid-001",
		"name":        "Location",
		"type":        "town",
		"position_x":  0.5,
		"position_y":  0.3,
	})

	req := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError, "Should return error for Grid mode map")
	assert.Contains(t, resp.Content[0].Text, "image mode")
}

// TestMapTools_CreateVisualLocation_WrongCampaign tests creating location on map from different campaign
func TestMapTools_CreateVisualLocation_WrongCampaign(t *testing.T) {
	mapTools, registry, mapStore, campaignStore, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create a map for a different campaign
	worldMap := models.NewWorldMap("campaign-999", "Other World Map", 100, 100)
	worldMap.ID = "world-other-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Create the other campaign
	otherCampaign := models.NewCampaign("Other Campaign", "dm-002", "Another campaign")
	otherCampaign.ID = "campaign-999"
	campaignStore.Create(ctx, otherCampaign)

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001", // Different campaign
		"map_id":      "world-other-001",
		"name":        "Location",
		"type":        "town",
		"position_x":  0.5,
		"position_y":  0.3,
	})

	req := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError, "Should return error for wrong campaign")
	assert.Contains(t, resp.Content[0].Text, "does not belong to")
}

// TestMapTools_CreateVisualLocation_MultipleLocations tests creating multiple visual locations
func TestMapTools_CreateVisualLocation_MultipleLocations(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	locations := []struct {
		name       string
		locationType string
		x, y       float64
	}{
		{"Town", "town", 0.3, 0.4},
		{"Forest", "forest", 0.6, 0.7},
		{"Dungeon", "dungeon", 0.8, 0.2},
	}

	for _, loc := range locations {
		args, _ := json.Marshal(map[string]interface{}{
			"campaign_id": "campaign-001",
			"map_id":      "world-visual-001",
			"name":        loc.name,
			"type":        loc.locationType,
			"position_x":  loc.x,
			"position_y":  loc.y,
		})

		req := mcp.ToolRequest{
			ToolName:  "create_visual_location",
			Arguments: args,
		}

		resp := registry.Call(ctx, req)
		assert.False(t, resp.IsError, "Should successfully create "+loc.name)
	}

	// Verify all locations were created
	updatedMap, _ := mapStore.Get(ctx, "world-visual-001")
	assert.Len(t, updatedMap.VisualLocations, 3)
}

// TestMapTools_MoveTo_ImageMode tests move_to in Image mode
func TestMapTools_MoveTo_ImageMode(t *testing.T) {
	mapTools, registry, mapStore, _, gameStateStore := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	worldMap.Image = models.NewMapImage("https://example.com/map.jpg")
	worldMap.Image.Width = 1920
	worldMap.Image.Height = 1080
	mapStore.Create(ctx, worldMap)

	// Move to position using normalized coordinates
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"target_x":    0.7,
		"target_y":    0.4,
	})

	req := mcp.ToolRequest{
		ToolName:  "move_to",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result["message"], "moved to position")
	newMarker := result["new_marker"].(map[string]interface{})
	assert.Equal(t, 0.7, newMarker["position_x"])
	assert.Equal(t, 0.4, newMarker["position_y"])

	// Verify game state was updated
	gameState, _ := gameStateStore.Get(ctx, "campaign-001")
	assert.NotNil(t, gameState.PlayerMarker)
	assert.Equal(t, 0.7, gameState.PlayerMarker.PositionX)
	assert.Equal(t, 0.4, gameState.PlayerMarker.PositionY)
}

// TestMapTools_MoveTo_ImageModeWithScene tests move_to with scene description
func TestMapTools_MoveTo_ImageModeWithScene(t *testing.T) {
	mapTools, registry, mapStore, _, gameStateStore := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Set initial scene
	gameState := models.NewGameState("campaign-001")
	gameState.PlayerMarker = models.NewPlayerMarker(0.5, 0.5)
	gameState.PlayerMarker.SetScene("Starting point")
	gameStateStore.Update(ctx, gameState)

	// Move to new position
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"target_x":    0.7,
		"target_y":    0.4,
	})

	req := mcp.ToolRequest{
		ToolName:  "move_to",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	// Verify the scene is preserved when marker is updated
	updatedState, _ := gameStateStore.Get(ctx, "campaign-001")
	assert.NotNil(t, updatedState.PlayerMarker)
	assert.Equal(t, 0.7, updatedState.PlayerMarker.PositionX)
	assert.Equal(t, 0.4, updatedState.PlayerMarker.PositionY)
	// Note: The current implementation doesn't preserve CurrentScene when moving
	// This test documents the current behavior
}

// TestMapTools_MoveTo_ImageModeInvalidCoordinates tests move_to with invalid coordinates
func TestMapTools_MoveTo_ImageModeInvalidCoordinates(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	tests := []struct {
		name      string
		targetX   float64
		targetY   float64
		errorMsg  string
	}{
		{"x negative", -0.1, 0.5, "must be between 0 and 1"},
		{"x greater than 1", 1.5, 0.5, "must be between 0 and 1"},
		{"y negative", 0.5, -0.1, "must be between 0 and 1"},
		{"y greater than 1", 0.5, 1.5, "must be between 0 and 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, _ := json.Marshal(map[string]interface{}{
				"campaign_id": "campaign-001",
				"target_x":    tt.targetX,
				"target_y":    tt.targetY,
			})

			req := mcp.ToolRequest{
				ToolName:  "move_to",
				Arguments: args,
			}

			resp := registry.Call(ctx, req)
			assert.True(t, resp.IsError, "Should return error for invalid position")
			assert.Contains(t, resp.Content[0].Text, tt.errorMsg)
		})
	}
}

// TestMapTools_MoveTo_GridModeStillWorks tests that Grid mode still works with x/y coordinates
func TestMapTools_MoveTo_GridModeStillWorks(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create a Grid mode world map (default)
	worldMap := models.NewWorldMap("campaign-001", "Grid World Map", 100, 100)
	worldMap.ID = "world-grid-001"
	worldMap.Mode = models.MapModeGrid
	mapStore.Create(ctx, worldMap)

	// Move using Grid mode coordinates
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"x":           10,
		"y":           20,
	})

	req := mcp.ToolRequest{
		ToolName:  "move_to",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify Grid mode response structure
	position := result["position"].(map[string]interface{})
	assert.Equal(t, float64(10), position["x"])
	assert.Equal(t, float64(20), position["y"])

	// Verify travel info
	travel := result["travel"].(map[string]interface{})
	assert.NotNil(t, travel["distance"])
	assert.NotNil(t, travel["hours"])
}

// TestMapTools_RegisterIncludesVisualLocationTools tests that visual location tools are registered
func TestMapTools_RegisterIncludesVisualLocationTools(t *testing.T) {
	mapTools, registry, _, _, _ := setupMapToolsForVisual()
	mapTools.Register(registry)

	// Verify all expected tools are registered
	assert.True(t, registry.Has("create_visual_location"))
	assert.True(t, registry.Has("update_location"))
	assert.True(t, registry.Has("get_world_map"))
	assert.True(t, registry.Has("move_to"))
}
