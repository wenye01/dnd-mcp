// Package tools contains integration tests for map update tools
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

// setupMapToolsForUpdate creates a MapTools instance for update tests
func setupMapToolsForUpdate() (*tools.MapTools, *mcp.Registry, *MockMapStore, *MockCampaignStore, *MockGameStateStore) {
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

func TestMapTools_UpdateVisualLocation_Success(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Create a visual location first
	createArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":  "campaign-001",
		"map_id":       "world-visual-001",
		"name":         "Waterdeep",
		"description":  "The City of Splendors",
		"type":         "town",
		"position_x":   0.5,
		"position_y":   0.3,
	})

	createReq := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: createArgs,
	}

	createResp := registry.Call(ctx, createReq)
	require.False(t, createResp.IsError)

	var createResult map[string]interface{}
	err := json.Unmarshal([]byte(createResp.Content[0].Text), &createResult)
	require.NoError(t, err)

	locationID := createResult["visual_location"].(map[string]interface{})["id"].(string)

	// Update the visual location with custom name
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":  "campaign-001",
		"map_id":       "world-visual-001",
		"location_id":  locationID,
		"custom_name":  "Deepwater",
		"description":  "The Crown of the North",
		"is_confirmed": true,
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.False(t, updateResp.IsError)

	var updateResult map[string]interface{}
	err = json.Unmarshal([]byte(updateResp.Content[0].Text), &updateResult)
	require.NoError(t, err)

	// Verify response
	visualLoc := updateResult["visual_location"].(map[string]interface{})
	assert.Equal(t, locationID, visualLoc["id"])
	assert.Equal(t, "Deepwater", visualLoc["custom_name"])
	assert.Equal(t, "The Crown of the North", visualLoc["description"])
	assert.Equal(t, true, visualLoc["is_confirmed"])
	assert.Equal(t, "Deepwater", visualLoc["display_name"]) // Custom name is used for display

	// Verify map was updated
	updatedMap, _ := mapStore.Get(ctx, "world-visual-001")
	assert.Len(t, updatedMap.VisualLocations, 1)
	assert.Equal(t, "Deepwater", updatedMap.VisualLocations[0].CustomName)
	assert.Equal(t, "The Crown of the North", updatedMap.VisualLocations[0].Description)
	assert.True(t, updatedMap.VisualLocations[0].IsConfirmed)
}

func TestMapTools_UpdateVisualLocation_PartialUpdate(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Create a visual location first
	createArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":  "campaign-001",
		"map_id":       "world-visual-001",
		"name":         "Location",
		"type":         "town",
		"position_x":   0.5,
		"position_y":   0.5,
	})

	createReq := mcp.ToolRequest{
		ToolName:  "create_visual_location",
		Arguments: createArgs,
	}

	createResp := registry.Call(ctx, createReq)
	require.False(t, createResp.IsError)

	var createResult map[string]interface{}
	err := json.Unmarshal([]byte(createResp.Content[0].Text), &createResult)
	require.NoError(t, err)

	locationID := createResult["visual_location"].(map[string]interface{})["id"].(string)

	// Update only custom name (partial update)
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"map_id":      "world-visual-001",
		"location_id": locationID,
		"custom_name": "Custom Location Name",
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.False(t, updateResp.IsError)

	// Verify only the custom name was updated
	updatedMap, _ := mapStore.Get(ctx, "world-visual-001")
	assert.Equal(t, "Custom Location Name", updatedMap.VisualLocations[0].CustomName)
	assert.Empty(t, updatedMap.VisualLocations[0].Description) // Original description is still empty
	assert.False(t, updatedMap.VisualLocations[0].IsConfirmed) // Original confirmation status is still false
}

func TestMapTools_UpdateVisualLocation_NotFound(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Try to update non-existent location
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"map_id":      "world-visual-001",
		"location_id": "nonexistent-id",
		"custom_name": "New Name",
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.True(t, updateResp.IsError)
	assert.Contains(t, updateResp.Content[0].Text, "visual location not found")
}

func TestMapTools_UpdateVisualLocation_MapNotFound(t *testing.T) {
	mapTools, registry, _, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Try to update location on non-existent map
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"map_id":      "nonexistent",
		"location_id": "location-id",
		"custom_name": "New Name",
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.True(t, updateResp.IsError)
}

func TestMapTools_UpdateVisualLocation_WrongCampaign(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map for campaign-001
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Create a visual location
	visualLoc := models.NewVisualLocation("Location", "desc", "town", 0.5, 0.5)
	worldMap.AddVisualLocation(*visualLoc)
	mapStore.Update(ctx, worldMap)

	// Try to update with different campaign ID
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-999", // Different campaign
		"map_id":      "world-visual-001",
		"location_id": visualLoc.ID,
		"custom_name": "New Name",
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.True(t, updateResp.IsError)
	assert.Contains(t, updateResp.Content[0].Text, "does not belong to the specified campaign")
}

func TestMapTools_UpdateVisualLocation_WithBattleMapID(t *testing.T) {
	mapTools, registry, mapStore, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	ctx := context.Background()

	// Create an image mode world map
	worldMap := models.NewWorldMap("campaign-001", "Visual World Map", 100, 100)
	worldMap.ID = "world-visual-001"
	worldMap.Mode = models.MapModeImage
	mapStore.Create(ctx, worldMap)

	// Create a visual location
	visualLoc := models.NewVisualLocation("Dungeon", "desc", "dungeon", 0.5, 0.5)
	worldMap.AddVisualLocation(*visualLoc)
	mapStore.Update(ctx, worldMap)

	// Update with battle map ID
	updateArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":   "campaign-001",
		"map_id":        "world-visual-001",
		"location_id":   visualLoc.ID,
		"battle_map_id": "battle-map-123",
	})

	updateReq := mcp.ToolRequest{
		ToolName:  "update_location",
		Arguments: updateArgs,
	}

	updateResp := registry.Call(ctx, updateReq)
	assert.False(t, updateResp.IsError)

	// Verify battle map ID was updated
	updatedMap, _ := mapStore.Get(ctx, "world-visual-001")
	assert.Equal(t, "battle-map-123", updatedMap.VisualLocations[0].BattleMapID)
}

func TestMapTools_RegisterIncludesUpdateLocation(t *testing.T) {
	mapTools, registry, _, _, _ := setupMapToolsForUpdate()
	mapTools.Register(registry)

	// Verify all tools are registered (should be 8 now with update_location)
	assert.Equal(t, 8, registry.Count())

	// Verify update_location is registered
	assert.True(t, registry.Has("update_location"))
}
