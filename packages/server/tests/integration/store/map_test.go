// Package store_test contains integration tests for the map store
package store_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store/postgres"
	"github.com/dnd-mcp/server/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMapTestModuleRoot returns the module root directory
func getMapTestModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getMapTestMigrationsPath returns the absolute path to migrations directory
func getMapTestMigrationsPath() string {
	return filepath.Join(getMapTestModuleRoot(), "internal", "store", "postgres", "migrations")
}

func getMapTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getMapTestEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getMapTestEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getMapTestEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getMapTestEnvOrDefault("POSTGRES_TEST_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getMapTestEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupMapTestDB sets up the test database and returns a client and cleanup function
func setupMapTestDB(t *testing.T) (*postgres.Client, *postgres.CampaignStore, *postgres.MapStore, func()) {
	t.Helper()

	cfg := getMapTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")

	// Check if database is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		t.Skipf("Database not available: %v", err)
	}

	// Run migrations
	migrator := postgres.NewMigratorWithPath(client, getMapTestMigrationsPath())
	err = migrator.Up(ctx)
	require.NoError(t, err, "Failed to run migrations")

	campaignStore := postgres.NewCampaignStore(client)
	mapStore := postgres.NewMapStore(client)

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Pool().Exec(ctx, "DELETE FROM maps")
		client.Pool().Exec(ctx, "DELETE FROM campaigns")
		client.Close()
	}

	return client, campaignStore, mapStore, cleanup
}

func TestMapStore_Create(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-001", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a world map
	worldMap := models.NewWorldMap(campaign.ID, "Forgotten Realms", 50, 50)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create world map")

	assert.NotEmpty(t, worldMap.ID, "ID should be set")
	assert.NotZero(t, worldMap.CreatedAt, "CreatedAt should be set")
	assert.NotZero(t, worldMap.UpdatedAt, "UpdatedAt should be set")
	assert.Equal(t, models.MapTypeWorld, worldMap.Type, "Type should be world")
	assert.Equal(t, campaign.ID, worldMap.CampaignID, "CampaignID should match")
}

func TestMapStore_Create_BattleMap(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-002", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a battle map
	battleMap := models.NewBattleMap(campaign.ID, "Dungeon Room 1", 20, 20, 5)
	err = mapStore.Create(ctx, battleMap)
	require.NoError(t, err, "Failed to create battle map")

	assert.Equal(t, models.MapTypeBattle, battleMap.Type, "Type should be battle")
	assert.Equal(t, 20, battleMap.Grid.Width, "Width should be 20")
	assert.Equal(t, 20, battleMap.Grid.Height, "Height should be 20")
	assert.Equal(t, 5, battleMap.Grid.CellSize, "CellSize should be 5")
}

func TestMapStore_Create_WithLocations(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-003", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a world map with locations
	worldMap := models.NewWorldMap(campaign.ID, "Faerun", 30, 30)

	// Add locations
	location1 := models.NewLocation("Waterdeep", "City of splendors", 10, 15)
	location2 := models.NewLocation("Baldur's Gate", "Gateway to the south", 20, 25)
	err = worldMap.AddLocation(*location1)
	require.NoError(t, err)
	err = worldMap.AddLocation(*location2)
	require.NoError(t, err)

	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create world map with locations")

	// Verify locations are preserved
	retrieved, err := mapStore.Get(ctx, worldMap.ID)
	require.NoError(t, err, "Failed to get map")
	assert.Len(t, retrieved.Locations, 2, "Should have 2 locations")
	assert.Equal(t, "Waterdeep", retrieved.Locations[0].Name)
	assert.Equal(t, "Baldur's Gate", retrieved.Locations[1].Name)
}

func TestMapStore_Create_WithTokens(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-004", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a battle map with tokens
	battleMap := models.NewBattleMap(campaign.ID, "Throne Room", 15, 15, 5)

	// Add tokens
	token1 := models.NewToken("character-001", 5, 5, models.TokenSizeMedium)
	token2 := models.NewToken("character-002", 10, 10, models.TokenSizeLarge)
	err = battleMap.AddToken(*token1)
	require.NoError(t, err)
	err = battleMap.AddToken(*token2)
	require.NoError(t, err)

	err = mapStore.Create(ctx, battleMap)
	require.NoError(t, err, "Failed to create battle map with tokens")

	// Verify tokens are preserved
	retrieved, err := mapStore.Get(ctx, battleMap.ID)
	require.NoError(t, err, "Failed to get map")
	assert.Len(t, retrieved.Tokens, 2, "Should have 2 tokens")
	assert.Equal(t, "character-001", retrieved.Tokens[0].CharacterID)
	assert.Equal(t, models.TokenSizeMedium, retrieved.Tokens[0].Size)
}

func TestMapStore_Get(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-005", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a map
	worldMap := models.NewWorldMap(campaign.ID, "Get Test Map", 40, 40)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create map")

	// Get the map
	retrieved, err := mapStore.Get(ctx, worldMap.ID)
	require.NoError(t, err, "Failed to get map")

	assert.Equal(t, worldMap.ID, retrieved.ID)
	assert.Equal(t, worldMap.Name, retrieved.Name)
	assert.Equal(t, worldMap.Type, retrieved.Type)
	assert.Equal(t, worldMap.CampaignID, retrieved.CampaignID)
	assert.Equal(t, worldMap.Grid.Width, retrieved.Grid.Width)
	assert.Equal(t, worldMap.Grid.Height, retrieved.Grid.Height)
}

func TestMapStore_Get_NotFound(t *testing.T) {
	_, _, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := mapStore.Get(ctx, "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMapStore_GetByCampaign(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-006", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create multiple maps for the campaign
	worldMap := models.NewWorldMap(campaign.ID, "World", 50, 50)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err)

	battleMap1 := models.NewBattleMap(campaign.ID, "Room 1", 10, 10, 5)
	battleMap1.ParentID = worldMap.ID
	err = mapStore.Create(ctx, battleMap1)
	require.NoError(t, err)

	battleMap2 := models.NewBattleMap(campaign.ID, "Room 2", 15, 15, 5)
	battleMap2.ParentID = worldMap.ID
	err = mapStore.Create(ctx, battleMap2)
	require.NoError(t, err)

	// Get all maps for the campaign
	maps, err := mapStore.GetByCampaign(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get maps by campaign")
	assert.Len(t, maps, 3, "Should have 3 maps")
}

func TestMapStore_Update(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-007", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a map
	worldMap := models.NewWorldMap(campaign.ID, "Original Name", 30, 30)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create map")

	// Update the map
	worldMap.Name = "Updated Name"
	// Add a location
	location := models.NewLocation("New Location", "A new place", 10, 10)
	err = worldMap.AddLocation(*location)
	require.NoError(t, err)

	err = mapStore.Update(ctx, worldMap)
	require.NoError(t, err, "Failed to update map")

	// Verify update
	retrieved, err := mapStore.Get(ctx, worldMap.ID)
	require.NoError(t, err, "Failed to get updated map")
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Len(t, retrieved.Locations, 1, "Should have 1 location")
	assert.Equal(t, "New Location", retrieved.Locations[0].Name)
}

func TestMapStore_Update_NotFound(t *testing.T) {
	_, _, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	worldMap := models.NewWorldMap("campaign-999", "Not Found", 10, 10)
	worldMap.ID = "00000000-0000-0000-0000-000000000000"

	err := mapStore.Update(ctx, worldMap)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMapStore_Delete(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Test Campaign", "dm-008", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a map
	worldMap := models.NewWorldMap(campaign.ID, "To Delete", 20, 20)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create map")

	// Delete the map
	err = mapStore.Delete(ctx, worldMap.ID)
	require.NoError(t, err, "Failed to delete map")

	// Verify deletion
	_, err = mapStore.Get(ctx, worldMap.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")
}

func TestMapStore_Delete_NotFound(t *testing.T) {
	_, _, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := mapStore.Delete(ctx, "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMapStore_GetWorldMap(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-009", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a world map
	worldMap := models.NewWorldMap(campaign.ID, "Main World", 60, 60)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create world map")

	// Get the world map
	retrieved, err := mapStore.GetWorldMap(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get world map")
	assert.Equal(t, worldMap.ID, retrieved.ID)
	assert.Equal(t, models.MapTypeWorld, retrieved.Type)
	assert.Empty(t, retrieved.ParentID, "World map should not have parent")
}

func TestMapStore_GetWorldMap_NotFound(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign without a world map
	campaign := models.NewCampaign("Test Campaign", "dm-010", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Try to get world map
	_, err = mapStore.GetWorldMap(ctx, campaign.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound when no world map exists")
}

func TestMapStore_GetBattleMap(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-011", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a battle map
	battleMap := models.NewBattleMap(campaign.ID, "Arena", 25, 25, 5)
	err = mapStore.Create(ctx, battleMap)
	require.NoError(t, err, "Failed to create battle map")

	// Get the battle map
	retrieved, err := mapStore.GetBattleMap(ctx, battleMap.ID)
	require.NoError(t, err, "Failed to get battle map")
	assert.Equal(t, battleMap.ID, retrieved.ID)
	assert.Equal(t, models.MapTypeBattle, retrieved.Type)
}

func TestMapStore_GetBattleMap_NotFound(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-012", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Try to get non-existent battle map
	_, err = mapStore.GetBattleMap(ctx, "00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMapStore_GetBattleMap_WrongType(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-013", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a world map
	worldMap := models.NewWorldMap(campaign.ID, "World", 40, 40)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create world map")

	// Try to get as battle map
	_, err = mapStore.GetBattleMap(ctx, worldMap.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound when map is not a battle map")
}

func TestMapStore_GetByParent(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-014", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a world map
	worldMap := models.NewWorldMap(campaign.ID, "World", 50, 50)
	err = mapStore.Create(ctx, worldMap)
	require.NoError(t, err, "Failed to create world map")

	// Create battle maps with parent
	for i := 0; i < 3; i++ {
		battleMap := models.NewBattleMap(campaign.ID, fmt.Sprintf("Dungeon Room %d", i+1), 15, 15, 5)
		battleMap.ParentID = worldMap.ID
		err = mapStore.Create(ctx, battleMap)
		require.NoError(t, err)
	}

	// Get battle maps by parent
	childMaps, err := mapStore.GetByParent(ctx, worldMap.ID)
	require.NoError(t, err, "Failed to get maps by parent")
	assert.Len(t, childMaps, 3, "Should have 3 child maps")

	// Verify all are battle maps
	for _, m := range childMaps {
		assert.Equal(t, models.MapTypeBattle, m.Type, "Child should be battle map")
		assert.Equal(t, worldMap.ID, m.ParentID, "ParentID should match")
	}
}

func TestMapStore_GetByParent_Empty(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-015", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Try to get children of non-existent parent
	childMaps, err := mapStore.GetByParent(ctx, "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err, "Should not error on empty result")
	assert.Len(t, childMaps, 0, "Should return empty slice")
}

func TestMapStore_Grid_CellTypes(t *testing.T) {
	_, campaignStore, mapStore, cleanup := setupMapTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Test Campaign", "dm-016", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a battle map with different cell types
	battleMap := models.NewBattleMap(campaign.ID, "Complex Map", 10, 10, 5)

	// Set different cell types
	battleMap.Grid.SetCell(0, 0, models.CellTypeWall)
	battleMap.Grid.SetCell(1, 1, models.CellTypeDoor)
	battleMap.Grid.SetCell(2, 2, models.CellTypeWater)
	battleMap.Grid.SetCell(3, 3, models.CellTypeDifficult)

	err = mapStore.Create(ctx, battleMap)
	require.NoError(t, err, "Failed to create map with custom cells")

	// Verify cell types are preserved
	retrieved, err := mapStore.Get(ctx, battleMap.ID)
	require.NoError(t, err, "Failed to get map")
	assert.Equal(t, models.CellTypeWall, retrieved.Grid.GetCell(0, 0))
	assert.Equal(t, models.CellTypeDoor, retrieved.Grid.GetCell(1, 1))
	assert.Equal(t, models.CellTypeWater, retrieved.Grid.GetCell(2, 2))
	assert.Equal(t, models.CellTypeDifficult, retrieved.Grid.GetCell(3, 3))
}
