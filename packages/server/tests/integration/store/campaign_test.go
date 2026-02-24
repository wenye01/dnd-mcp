// Package store_test contains integration tests for the campaign store
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
	"github.com/dnd-mcp/server/internal/store"
	"github.com/dnd-mcp/server/internal/store/postgres"
	"github.com/dnd-mcp/server/pkg/config"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getCampaignTestModuleRoot returns the module root directory
func getCampaignTestModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getCampaignTestMigrationsPath returns the absolute path to migrations directory
func getCampaignTestMigrationsPath() string {
	return filepath.Join(getCampaignTestModuleRoot(), "internal", "store", "postgres", "migrations")
}

func getCampaignTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getCampaignTestEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getCampaignTestEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getCampaignTestEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getCampaignTestEnvOrDefault("POSTGRES_TEST_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getCampaignTestEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestDB sets up the test database and returns a client and cleanup function
func setupTestDB(t *testing.T) (*postgres.Client, *postgres.CampaignStore, *postgres.GameStateStore, func()) {
	t.Helper()

	// Load .env file
	envPath := filepath.Join(getCampaignTestModuleRoot(), ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	cfg := getCampaignTestConfig()

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
	migrator := postgres.NewMigratorWithPath(client, getCampaignTestMigrationsPath())
	err = migrator.Up(ctx)
	require.NoError(t, err, "Failed to run migrations")

	campaignStore := postgres.NewCampaignStore(client)
	gameStateStore := postgres.NewGameStateStore(client)

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Pool().Exec(ctx, "DELETE FROM game_states")
		client.Pool().Exec(ctx, "DELETE FROM campaigns")
		client.Close()
	}

	return client, campaignStore, gameStateStore, cleanup
}

func TestCampaignStore_Create(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := models.NewCampaign("Test Campaign", "dm-001", "Test description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	assert.NotEmpty(t, campaign.ID, "ID should be set")
	assert.NotZero(t, campaign.CreatedAt, "CreatedAt should be set")
	assert.NotZero(t, campaign.UpdatedAt, "UpdatedAt should be set")
	assert.Equal(t, models.CampaignStatusActive, campaign.Status, "Status should be active")
}

func TestCampaignStore_Create_WithSettings(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	settings := &models.CampaignSettings{
		MaxPlayers:    6,
		StartLevel:    3,
		Ruleset:       "dnd5e",
		ContextWindow: 30,
		HouseRules:    map[string]interface{}{"critical": "double"},
	}

	campaign := models.NewCampaign("Custom Campaign", "dm-002", "")
	campaign.Settings = settings

	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Retrieve and verify
	retrieved, err := campaignStore.Get(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get campaign")

	assert.Equal(t, 6, retrieved.Settings.MaxPlayers)
	assert.Equal(t, 3, retrieved.Settings.StartLevel)
	assert.Equal(t, "dnd5e", retrieved.Settings.Ruleset)
	assert.Equal(t, 30, retrieved.Settings.ContextWindow)
}

func TestCampaignStore_Get(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Get Test", "dm-003", "Description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Get the campaign
	retrieved, err := campaignStore.Get(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get campaign")

	assert.Equal(t, campaign.ID, retrieved.ID)
	assert.Equal(t, campaign.Name, retrieved.Name)
	assert.Equal(t, campaign.Description, retrieved.Description)
	assert.Equal(t, campaign.DMID, retrieved.DMID)
	assert.Equal(t, campaign.Status, retrieved.Status)
}

func TestCampaignStore_Get_NotFound(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := campaignStore.Get(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCampaignStore_GetByIDAndDM(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("DM Test", "dm-004", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Get by ID and DM
	retrieved, err := campaignStore.GetByIDAndDM(ctx, campaign.ID, "dm-004")
	require.NoError(t, err, "Failed to get campaign")
	assert.Equal(t, campaign.ID, retrieved.ID)

	// Try wrong DM
	_, err = campaignStore.GetByIDAndDM(ctx, campaign.ID, "wrong-dm")
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound for wrong DM")
}

func TestCampaignStore_List(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple campaigns
	for i := 0; i < 5; i++ {
		campaign := models.NewCampaign(
			fmt.Sprintf("Campaign %d", i),
			fmt.Sprintf("dm-%d", i),
			"",
		)
		err := campaignStore.Create(ctx, campaign)
		require.NoError(t, err, "Failed to create campaign")
	}

	// List all
	campaigns, err := campaignStore.List(ctx, nil)
	require.NoError(t, err, "Failed to list campaigns")
	assert.GreaterOrEqual(t, len(campaigns), 5, "Should have at least 5 campaigns")
}

func TestCampaignStore_List_WithFilter(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaigns with different DMs and statuses
	campaign1 := models.NewCampaign("Campaign 1", "dm-filter", "")
	err := campaignStore.Create(ctx, campaign1)
	require.NoError(t, err)

	campaign2 := models.NewCampaign("Campaign 2", "dm-filter", "")
	campaign2.Status = models.CampaignStatusPaused
	err = campaignStore.Create(ctx, campaign2)
	require.NoError(t, err)

	campaign3 := models.NewCampaign("Campaign 3", "dm-other", "")
	err = campaignStore.Create(ctx, campaign3)
	require.NoError(t, err)

	// Filter by DM
	filter := &store.CampaignFilter{DMID: "dm-filter"}
	campaigns, err := campaignStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list campaigns")
	assert.Len(t, campaigns, 2, "Should have 2 campaigns for dm-filter")

	// Filter by status
	filter = &store.CampaignFilter{Status: models.CampaignStatusPaused}
	campaigns, err = campaignStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list campaigns")
	assert.GreaterOrEqual(t, len(campaigns), 1, "Should have at least 1 paused campaign")
}

func TestCampaignStore_List_Pagination(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create 10 campaigns
	for i := 0; i < 10; i++ {
		campaign := models.NewCampaign(
			fmt.Sprintf("Page Campaign %d", i),
			fmt.Sprintf("dm-page-%d", i),
			"",
		)
		err := campaignStore.Create(ctx, campaign)
		require.NoError(t, err)
	}

	// Get first page
	filter := &store.CampaignFilter{Limit: 5, Offset: 0}
	page1, err := campaignStore.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, page1, 5, "First page should have 5 campaigns")

	// Get second page
	filter = &store.CampaignFilter{Limit: 5, Offset: 5}
	page2, err := campaignStore.List(ctx, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(page2), 5, "Second page should have at least 5 campaigns")
}

func TestCampaignStore_Update(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Update Test", "dm-005", "Original description")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Update campaign
	campaign.Description = "Updated description"
	campaign.Status = models.CampaignStatusPaused
	err = campaignStore.Update(ctx, campaign)
	require.NoError(t, err, "Failed to update campaign")

	// Verify update
	retrieved, err := campaignStore.Get(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, models.CampaignStatusPaused, retrieved.Status)
}

func TestCampaignStore_Update_NotFound(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := models.NewCampaign("Not Found", "dm-006", "")
	campaign.ID = uuid.New().String()

	err := campaignStore.Update(ctx, campaign)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCampaignStore_Delete(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Delete Test", "dm-007", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Delete campaign
	err = campaignStore.Delete(ctx, campaign.ID)
	require.NoError(t, err, "Failed to delete campaign")

	// Verify soft delete (should not be found)
	_, err = campaignStore.Get(ctx, campaign.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")

	// Verify soft delete with include deleted
	filter := &store.CampaignFilter{IncludeDeleted: true}
	campaigns, err := campaignStore.List(ctx, filter)
	require.NoError(t, err)

	found := false
	for _, c := range campaigns {
		if c.ID == campaign.ID {
			found = true
			assert.Equal(t, models.CampaignStatusArchived, c.Status)
			assert.NotNil(t, c.DeletedAt)
			break
		}
	}
	assert.True(t, found, "Campaign should be found when including deleted")
}

func TestCampaignStore_Delete_NotFound(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := campaignStore.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCampaignStore_HardDelete(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Hard Delete Test", "dm-008", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Hard delete campaign
	err = campaignStore.HardDelete(ctx, campaign.ID)
	require.NoError(t, err, "Failed to hard delete campaign")

	// Verify not found even with include deleted
	filter := &store.CampaignFilter{IncludeDeleted: true}
	campaigns, err := campaignStore.List(ctx, filter)
	require.NoError(t, err)

	for _, c := range campaigns {
		assert.NotEqual(t, campaign.ID, c.ID, "Campaign should be permanently deleted")
	}
}

func TestCampaignStore_Count(t *testing.T) {
	_, campaignStore, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaigns
	for i := 0; i < 3; i++ {
		campaign := models.NewCampaign(
			fmt.Sprintf("Count Campaign %d", i),
			fmt.Sprintf("dm-count-%d", i),
			"",
		)
		err := campaignStore.Create(ctx, campaign)
		require.NoError(t, err)
	}

	// Count all
	count, err := campaignStore.Count(ctx, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3), "Should have at least 3 campaigns")

	// Count with filter
	filter := &store.CampaignFilter{DMID: "dm-count-0"}
	count, err = campaignStore.Count(ctx, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1), "Should have at least 1 campaign for dm-count-0")
}

// GameState Store Tests

func TestGameStateStore_Create(t *testing.T) {
	_, campaignStore, gameStateStore, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("GS Test", "dm-gs-001", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create game state
	gameState := models.NewGameState(campaign.ID)
	err = gameStateStore.Create(ctx, gameState)
	require.NoError(t, err, "Failed to create game state")

	assert.Equal(t, campaign.ID, gameState.ID)
	assert.Equal(t, campaign.ID, gameState.CampaignID)
	assert.NotZero(t, gameState.UpdatedAt)
}

func TestGameStateStore_Get(t *testing.T) {
	_, campaignStore, gameStateStore, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaign and game state
	campaign := models.NewCampaign("GS Get Test", "dm-gs-002", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	gameState := models.NewGameState(campaign.ID)
	err = gameStateStore.Create(ctx, gameState)
	require.NoError(t, err)

	// Get game state
	retrieved, err := gameStateStore.Get(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get game state")

	assert.Equal(t, campaign.ID, retrieved.CampaignID)
	assert.NotNil(t, retrieved.GameTime)
	assert.NotNil(t, retrieved.PartyPosition)
	assert.Equal(t, models.MapTypeWorld, retrieved.CurrentMapType)
}

func TestGameStateStore_Get_NotFound(t *testing.T) {
	_, _, gameStateStore, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := gameStateStore.Get(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestGameStateStore_Update(t *testing.T) {
	client, campaignStore, gameStateStore, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaign and game state
	campaign := models.NewCampaign("GS Update Test", "dm-gs-003", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	gameState := models.NewGameState(campaign.ID)
	err = gameStateStore.Create(ctx, gameState)
	require.NoError(t, err)

	// Insert a mock combat record to test ActiveCombatID FK constraint
	combatID := uuid.New().String()
	_, err = client.Pool().Exec(ctx, `
		INSERT INTO combats (id, campaign_id, status, round, turn_index, participants)
		VALUES ($1, $2, 'active', 1, 0, '[]'::jsonb)
	`, combatID, campaign.ID)
	require.NoError(t, err, "Failed to insert mock combat")

	// Update game state
	gameState.Weather = "rain"
	gameState.AdvanceTime(5)
	gameState.CurrentMapType = models.MapTypeBattle
	gameState.ActiveCombatID = combatID

	err = gameStateStore.Update(ctx, gameState)
	require.NoError(t, err, "Failed to update game state")

	// Verify update
	retrieved, err := gameStateStore.Get(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, "rain", retrieved.Weather)
	assert.Equal(t, models.MapTypeBattle, retrieved.CurrentMapType)
	assert.Equal(t, combatID, retrieved.ActiveCombatID)
}

func TestGameStateStore_Delete(t *testing.T) {
	_, campaignStore, gameStateStore, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaign and game state
	campaign := models.NewCampaign("GS Delete Test", "dm-gs-004", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	gameState := models.NewGameState(campaign.ID)
	err = gameStateStore.Create(ctx, gameState)
	require.NoError(t, err)

	// Delete game state
	err = gameStateStore.Delete(ctx, campaign.ID)
	require.NoError(t, err, "Failed to delete game state")

	// Verify deletion
	_, err = gameStateStore.Get(ctx, campaign.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")
}
