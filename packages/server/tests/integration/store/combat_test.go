// Package store_test contains integration tests for the combat store
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
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getCombatTestModuleRoot returns the module root directory
func getCombatTestModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getCombatTestMigrationsPath returns the absolute path to migrations directory
func getCombatTestMigrationsPath() string {
	return filepath.Join(getCombatTestModuleRoot(), "internal", "store", "postgres", "migrations")
}

func getCombatTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getCombatTestEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getCombatTestEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getCombatTestEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getCombatTestEnvOrDefault("POSTGRES_TEST_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getCombatTestEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupCombatTestDB sets up the test database and returns stores and cleanup function
func setupCombatTestDB(t *testing.T) (*postgres.Client, *postgres.CampaignStore, *postgres.CharacterStore, *postgres.CombatStore, func()) {
	t.Helper()

	// Load .env file
	envPath := filepath.Join(getCombatTestModuleRoot(), ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	cfg := getCombatTestConfig()

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
	migrator := postgres.NewMigratorWithPath(client, getCombatTestMigrationsPath())
	err = migrator.Up(ctx)
	require.NoError(t, err, "Failed to run migrations")

	campaignStore := postgres.NewCampaignStore(client)
	characterStore := postgres.NewCharacterStore(client)
	combatStore := postgres.NewCombatStore(client)

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Pool().Exec(ctx, "DELETE FROM combats")
		client.Pool().Exec(ctx, "DELETE FROM characters")
		client.Pool().Exec(ctx, "DELETE FROM game_states")
		client.Pool().Exec(ctx, "DELETE FROM campaigns")
		// Note: We don't clean schema_migrations to avoid re-running migrations
		// Migrator.Up() is idempotent - it skips already applied migrations
		client.Close()
	}

	return client, campaignStore, characterStore, combatStore, cleanup
}

// createCombatTestCampaign creates a test campaign for combat tests
func createCombatTestCampaign(t *testing.T, ctx context.Context, campaignStore *postgres.CampaignStore, suffix string) *models.Campaign {
	campaign := models.NewCampaign(
		fmt.Sprintf("Combat Test Campaign %s", suffix),
		fmt.Sprintf("dm-combat-%s", suffix),
		"Test campaign for combat tests",
	)
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create test campaign")
	return campaign
}

// createCombatTestCharacter creates a test character for combat tests
func createCombatTestCharacter(t *testing.T, ctx context.Context, characterStore *postgres.CharacterStore, campaignID, name string, isNPC bool) *models.Character {
	character := models.NewCharacter(campaignID, name, isNPC)
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(10)
	if !isNPC {
		character.PlayerID = "player-" + name
	} else {
		character.NPCType = models.NPCTypeScripted
	}
	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create test character")
	return character
}

func TestCombatStore_Create(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "create")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero 1", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster 1", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 15
	combat.Participants[1].Initiative = 12

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	assert.NotEmpty(t, combat.ID, "ID should be set")
	assert.NotZero(t, combat.StartedAt, "StartedAt should be set")
	assert.Equal(t, models.CombatStatusActive, combat.Status)
	assert.Equal(t, 1, combat.Round)
	assert.Equal(t, 0, combat.TurnIndex)
	assert.Len(t, combat.Participants, 2)
}

func TestCombatStore_Create_WithFullData(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "full-data")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Full", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Full", true)

	// Create combat with full data
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 18
	combat.Participants[0].TempHP = 5
	combat.Participants[0].SetPosition(3, 4)
	combat.Participants[0].AddCondition("blessed", 10, "spell")
	combat.Participants[1].Initiative = 14
	combat.Participants[1].SetPosition(8, 6)
	combat.Round = 2
	combat.TurnIndex = 1
	combat.MapID = uuid.New().String()

	// Add log entries
	combat.AddLogEntry(char1.ID, "attack", char2.ID, "hit for 8 damage")
	combat.AddLogEntry(char2.ID, "attack", char1.ID, "missed")

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat with full data")

	// Retrieve and verify
	retrieved, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")

	assert.Equal(t, 2, retrieved.Round)
	assert.Equal(t, 1, retrieved.TurnIndex)
	assert.NotEmpty(t, retrieved.MapID)
	assert.Len(t, retrieved.Participants, 2)
	assert.Len(t, retrieved.Log, 2)

	// Check participant details
	assert.Equal(t, 18, retrieved.Participants[0].Initiative)
	assert.Equal(t, 5, retrieved.Participants[0].TempHP)
	assert.NotNil(t, retrieved.Participants[0].Position)
	assert.Equal(t, 3, retrieved.Participants[0].Position.X)
	assert.Equal(t, 4, retrieved.Participants[0].Position.Y)
	assert.True(t, retrieved.Participants[0].HasCondition("blessed"))
}

func TestCombatStore_Get(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "get")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Get", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Get", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 16
	combat.Participants[1].Initiative = 11

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// Get the combat
	retrieved, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")

	assert.Equal(t, combat.ID, retrieved.ID)
	assert.Equal(t, combat.CampaignID, retrieved.CampaignID)
	assert.Equal(t, combat.Status, retrieved.Status)
	assert.Equal(t, combat.Round, retrieved.Round)
	assert.Equal(t, combat.TurnIndex, retrieved.TurnIndex)
	assert.Len(t, retrieved.Participants, 2)
}

func TestCombatStore_Get_NotFound(t *testing.T) {
	_, _, _, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := combatStore.Get(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCombatStore_GetByCampaign(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaigns
	campaign1 := createCombatTestCampaign(t, ctx, campaignStore, "by-campaign-1")
	campaign2 := createCombatTestCampaign(t, ctx, campaignStore, "by-campaign-2")

	// Create characters for campaign 1
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign1.ID, "Hero 1", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign1.ID, "Monster 1", true)

	// Create characters for campaign 2
	char3 := createCombatTestCharacter(t, ctx, characterStore, campaign2.ID, "Hero 2", false)
	char4 := createCombatTestCharacter(t, ctx, characterStore, campaign2.ID, "Monster 2", true)

	// Create combats for campaign 1
	combat1 := models.NewCombat(campaign1.ID, []string{char1.ID, char2.ID})
	err := combatStore.Create(ctx, combat1)
	require.NoError(t, err)

	combat2 := models.NewCombat(campaign1.ID, []string{char1.ID, char2.ID})
	err = combatStore.Create(ctx, combat2)
	require.NoError(t, err)

	// Create combat for campaign 2
	combat3 := models.NewCombat(campaign2.ID, []string{char3.ID, char4.ID})
	err = combatStore.Create(ctx, combat3)
	require.NoError(t, err)

	// Get combats for campaign 1
	combats, err := combatStore.GetByCampaign(ctx, campaign1.ID)
	require.NoError(t, err, "Failed to get combats by campaign")
	assert.Len(t, combats, 2, "Should have 2 combats for campaign 1")

	// Get combats for campaign 2
	combats, err = combatStore.GetByCampaign(ctx, campaign2.ID)
	require.NoError(t, err, "Failed to get combats by campaign")
	assert.Len(t, combats, 1, "Should have 1 combat for campaign 2")
}

func TestCombatStore_GetActive(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "active")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Active", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Active", true)

	// Create active combat
	activeCombat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	activeCombat.Participants[0].Initiative = 15
	activeCombat.Participants[1].Initiative = 12

	err := combatStore.Create(ctx, activeCombat)
	require.NoError(t, err, "Failed to create active combat")

	// Get active combat
	retrieved, err := combatStore.GetActive(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get active combat")
	assert.Equal(t, activeCombat.ID, retrieved.ID)
	assert.Equal(t, models.CombatStatusActive, retrieved.Status)
}

func TestCombatStore_GetActive_NoActiveCombat(t *testing.T) {
	_, campaignStore, _, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign without any combats
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "no-active")

	// Get active combat (should return ErrNotFound)
	_, err := combatStore.GetActive(ctx, campaign.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound when no active combat")
}

func TestCombatStore_GetActive_OnlyOne(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "only-one")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Only", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Only", true)

	// Create first active combat
	combat1 := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	err := combatStore.Create(ctx, combat1)
	require.NoError(t, err)

	// End the first combat
	combat1.End()
	err = combatStore.Update(ctx, combat1)
	require.NoError(t, err)

	// Create second active combat
	combat2 := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	err = combatStore.Create(ctx, combat2)
	require.NoError(t, err)

	// Get active combat - should return the second one
	retrieved, err := combatStore.GetActive(ctx, campaign.ID)
	require.NoError(t, err, "Failed to get active combat")
	assert.Equal(t, combat2.ID, retrieved.ID, "Should return the second active combat")
	assert.Equal(t, models.CombatStatusActive, retrieved.Status)

	// Verify first combat is finished
	retrieved1, err := combatStore.Get(ctx, combat1.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CombatStatusFinished, retrieved1.Status)
	assert.NotNil(t, retrieved1.EndedAt)
}

func TestCombatStore_Update(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "update")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Update", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Update", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 15
	combat.Participants[1].Initiative = 12

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// Update combat - advance turn
	combat.AdvanceTurn()
	combat.AddLogEntry(char1.ID, "attack", char2.ID, "hit for 6 damage")

	err = combatStore.Update(ctx, combat)
	require.NoError(t, err, "Failed to update combat")

	// Verify update
	retrieved, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")
	assert.Equal(t, 1, retrieved.TurnIndex)
	assert.Len(t, retrieved.Log, 1)
}

func TestCombatStore_Update_EndCombat(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "end")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero End", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster End", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 15
	combat.Participants[1].Initiative = 12

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// End combat
	combat.End()

	err = combatStore.Update(ctx, combat)
	require.NoError(t, err, "Failed to update combat")

	// Verify update
	retrieved, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")
	assert.Equal(t, models.CombatStatusFinished, retrieved.Status)
	assert.NotNil(t, retrieved.EndedAt)
}

func TestCombatStore_Update_NotFound(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign and character to have valid IDs
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "update-notfound")
	char := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero NF", false)

	// Try to update non-existent combat
	combat := models.NewCombat(campaign.ID, []string{char.ID})
	combat.ID = uuid.New().String()

	err := combatStore.Update(ctx, combat)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCombatStore_Delete(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "delete")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Delete", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Delete", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// Delete combat
	err = combatStore.Delete(ctx, combat.ID)
	require.NoError(t, err, "Failed to delete combat")

	// Verify deletion
	_, err = combatStore.Get(ctx, combat.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")
}

func TestCombatStore_Delete_NotFound(t *testing.T) {
	_, _, _, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := combatStore.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCombatStore_ParticipantWithConditions(t *testing.T) {
	_, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "conditions")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Cond", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Cond", true)

	// Create combat with conditions
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})
	combat.Participants[0].Initiative = 15
	combat.Participants[0].AddCondition("poisoned", 10, "trap")
	combat.Participants[0].AddCondition("blessed", 5, "spell")
	combat.Participants[1].Initiative = 12
	combat.Participants[1].AddCondition("prone", -1, "shove")

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// Retrieve and verify conditions
	retrieved, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")

	assert.Len(t, retrieved.Participants[0].Conditions, 2)
	assert.True(t, retrieved.Participants[0].HasCondition("poisoned"))
	assert.True(t, retrieved.Participants[0].HasCondition("blessed"))
	assert.Len(t, retrieved.Participants[1].Conditions, 1)
	assert.True(t, retrieved.Participants[1].HasCondition("prone"))

	// Update conditions - tick them
	expired := retrieved.Participants[0].TickConditions()
	assert.Len(t, expired, 0) // No conditions should expire yet

	err = combatStore.Update(ctx, retrieved)
	require.NoError(t, err, "Failed to update combat")

	// Verify conditions still exist after update
	retrieved2, err := combatStore.Get(ctx, combat.ID)
	require.NoError(t, err, "Failed to get combat")
	assert.Len(t, retrieved2.Participants[0].Conditions, 2)
}

func TestCombatStore_CascadeDelete(t *testing.T) {
	client, campaignStore, characterStore, combatStore, cleanup := setupCombatTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and characters
	campaign := createCombatTestCampaign(t, ctx, campaignStore, "cascade")
	char1 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Hero Cascade", false)
	char2 := createCombatTestCharacter(t, ctx, characterStore, campaign.ID, "Monster Cascade", true)

	// Create combat
	combat := models.NewCombat(campaign.ID, []string{char1.ID, char2.ID})

	err := combatStore.Create(ctx, combat)
	require.NoError(t, err, "Failed to create combat")

	// Delete campaign (should cascade delete combat)
	_, err = client.Pool().Exec(ctx, "DELETE FROM campaigns WHERE id = $1", campaign.ID)
	require.NoError(t, err)

	// Verify combat is deleted
	_, err = combatStore.Get(ctx, combat.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Combat should be deleted when campaign is deleted")
}
