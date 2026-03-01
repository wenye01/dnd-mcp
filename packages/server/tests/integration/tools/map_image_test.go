// Package tools contains integration tests for map tools in Image mode
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

// setupMapToolsForImage creates a MapTools instance with character support for Image mode tests
func setupMapToolsForImage() (*tools.MapTools, *mcp.Registry, *MockMapStore, *MockCampaignStore, *MockGameStateStore, *MockCharacterStore) {
	mapStore := NewMockMapStore()
	campaignStore := NewMockCampaignStore()
	gameStateStore := NewMockGameStateStore()
	characterStore := NewMockCharacterStore()

	// Create a test campaign
	campaign := models.NewCampaign("Test Campaign", "dm-001", "A test campaign")
	campaign.ID = "campaign-001"
	campaignStore.Create(context.Background(), campaign)

	// Create a test game state
	gameState := models.NewGameState("campaign-001")
	gameStateStore.Create(context.Background(), gameState)

	mapServiceWithChars := service.NewMapServiceWithCharacters(mapStore, campaignStore, gameStateStore, characterStore)
	mapTools := tools.NewMapToolsWithCharacters(mapServiceWithChars)
	registry := mcp.NewRegistry()

	return mapTools, registry, mapStore, campaignStore, gameStateStore, characterStore
}

// TestMapTools_GetWorldMap_ImageMode tests getting world map in Image mode
func TestMapTools_GetWorldMap_ImageMode(t *testing.T) {
	ctx := context.Background()
	mapTools, registry, mapStore, campaignStore, gameStateStore, _ := setupMapToolsForImage()
	mapTools.Register(registry)

	// Create a test campaign
	campaign := models.NewCampaign("Visual Campaign", "dm-001", "A test visual campaign")
	campaign.ID = "campaign-visual-001"
	campaignStore.Create(ctx, campaign)

	// Create an Image mode world map
	worldMap := models.NewWorldMap("campaign-visual-001", "Visual World Map", 50, 50)
	worldMap.Mode = models.MapModeImage
	worldMap.ID = "world-visual-001"
	worldMap.Image = models.NewMapImage("https://example.com/map.jpg")
	worldMap.Image.Width = 1920
	worldMap.Image.Height = 1080
	mapStore.Create(ctx, worldMap)

	// Add visual locations
	vloc1 := models.NewVisualLocation("Town", "A small town", "town", 0.3, 0.4)
	vloc2 := models.NewVisualLocation("Forest", "Dark forest", "forest", 0.6, 0.7)
	worldMap.AddVisualLocation(*vloc1)
	worldMap.AddVisualLocation(*vloc2)
	mapStore.Update(ctx, worldMap)

	// Create game state with player marker
	gameState := models.NewGameState("campaign-visual-001")
	gameState.PlayerMarker = models.NewPlayerMarker(0.5, 0.5)
	gameState.PlayerMarker.CurrentScene = "Near the town"
	gameStateStore.Create(ctx, gameState)

	// Call get_world_map tool
	req := mcp.ToolRequest{
		ToolName: "get_world_map",
		Arguments: []byte(`{"campaign_id": "campaign-visual-001"}`),
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "image", result["mode"])
	assert.NotNil(t, result["map"])

	// Verify image data
	imageData, ok := result["image"].(map[string]interface{})
	assert.True(t, ok, "Should have image data")
	assert.Equal(t, "https://example.com/map.jpg", imageData["url"])
	assert.Equal(t, float64(1920), imageData["width"])
	assert.Equal(t, float64(1080), imageData["height"])

	// Verify visual locations
	visualLocations, ok := result["visual_locations"].([]interface{})
	assert.True(t, ok, "Should have visual_locations")
	assert.Len(t, visualLocations, 2)

	// Verify player marker
	playerMarker, ok := result["player_marker"].(map[string]interface{})
	assert.True(t, ok, "Should have player_marker")
	assert.Equal(t, 0.5, playerMarker["position_x"])
	assert.Equal(t, 0.5, playerMarker["position_y"])
	assert.Equal(t, "Near the town", playerMarker["current_scene"])

	// Verify Grid mode fields are not present for Image mode
	_, hasGrid := result["grid"]
	assert.False(t, hasGrid, "Should not have grid in Image mode")
	_, hasLocations := result["locations"]
	assert.False(t, hasLocations, "Should not have locations in Image mode")
}

// TestMapTools_GetWorldMap_GridMode tests getting world map in Grid mode (backward compatibility)
func TestMapTools_GetWorldMap_GridMode(t *testing.T) {
	ctx := context.Background()
	mapTools, registry, mapStore, campaignStore, gameStateStore, _ := setupMapToolsForImage()
	mapTools.Register(registry)

	// Create a test campaign
	campaign := models.NewCampaign("Grid Campaign", "dm-001", "A test grid campaign")
	campaign.ID = "campaign-grid-001"
	campaignStore.Create(ctx, campaign)

	// Create a Grid mode world map (default mode)
	worldMap := models.NewWorldMap("campaign-grid-001", "Grid World Map", 50, 50)
	worldMap.ID = "world-grid-001"
	worldMap.Mode = models.MapModeGrid
	mapStore.Create(ctx, worldMap)

	// Add locations
	loc1 := models.NewLocation("Town A", "First town", 10, 20)
	loc2 := models.NewLocation("Town B", "Second town", 30, 40)
	worldMap.AddLocation(*loc1)
	worldMap.AddLocation(*loc2)
	mapStore.Update(ctx, worldMap)

	// Create game state (no player marker in Grid mode)
	gameState := models.NewGameState("campaign-grid-001")
	gameStateStore.Create(ctx, gameState)

	// Call get_world_map tool
	req := mcp.ToolRequest{
		ToolName: "get_world_map",
		Arguments: []byte(`{"campaign_id": "campaign-grid-001"}`),
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "grid", result["mode"])
	assert.NotNil(t, result["map"])

	// Verify grid data
	gridData, ok := result["grid"].(map[string]interface{})
	assert.True(t, ok, "Should have grid data")
	assert.Equal(t, float64(50), gridData["width"])
	assert.Equal(t, float64(50), gridData["height"])
	assert.NotNil(t, gridData["cells"])

	// Verify locations
	locations, ok := result["locations"].([]interface{})
	assert.True(t, ok, "Should have locations")
	assert.Len(t, locations, 2)

	// Verify Image mode fields are not present for Grid mode
	_, hasImage := result["image"]
	assert.False(t, hasImage, "Should not have image in Grid mode")
	_, hasVisualLocations := result["visual_locations"]
	assert.False(t, hasVisualLocations, "Should not have visual_locations in Grid mode")
	_, hasPlayerMarker := result["player_marker"]
	assert.False(t, hasPlayerMarker, "Should not have player_marker in Grid mode")
}

// TestMapTools_GetWorldMap_ImageModeNoMarker tests Image mode without player marker
func TestMapTools_GetWorldMap_ImageModeNoMarker(t *testing.T) {
	ctx := context.Background()
	mapTools, registry, mapStore, campaignStore, gameStateStore, _ := setupMapToolsForImage()
	mapTools.Register(registry)

	// Create a test campaign
	campaign := models.NewCampaign("Visual Campaign", "dm-001", "A test visual campaign")
	campaign.ID = "campaign-visual-002"
	campaignStore.Create(ctx, campaign)

	// Create an Image mode world map without image
	worldMap := models.NewWorldMap("campaign-visual-002", "Visual World Map", 50, 50)
	worldMap.Mode = models.MapModeImage
	worldMap.ID = "world-visual-002"
	// No Image set
	mapStore.Create(ctx, worldMap)

	// Create game state without player marker
	gameState := models.NewGameState("campaign-visual-002")
	gameStateStore.Create(ctx, gameState)

	// Call get_world_map tool
	req := mcp.ToolRequest{
		ToolName: "get_world_map",
		Arguments: []byte(`{"campaign_id": "campaign-visual-002"}`),
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Response should not be an error")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "image", result["mode"])
	assert.Nil(t, result["image"], "Image should be nil when not set")
	assert.Nil(t, result["player_marker"], "Player marker should be nil when not set")
	assert.Empty(t, result["visual_locations"], "Visual locations should be empty")
}
