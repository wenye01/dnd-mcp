// Package store_test contains integration tests for the character store
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

// getCharacterTestModuleRoot returns the module root directory
func getCharacterTestModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getCharacterTestMigrationsPath returns the absolute path to migrations directory
func getCharacterTestMigrationsPath() string {
	return filepath.Join(getCharacterTestModuleRoot(), "internal", "store", "postgres", "migrations")
}

func getCharacterTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getCharacterTestEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getCharacterTestEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getCharacterTestEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getCharacterTestEnvOrDefault("POSTGRES_TEST_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getCharacterTestEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupCharacterTestDB sets up the test database and returns stores and cleanup function
func setupCharacterTestDB(t *testing.T) (*postgres.Client, *postgres.CampaignStore, *postgres.CharacterStore, func()) {
	t.Helper()

	// Load .env file
	envPath := filepath.Join(getCharacterTestModuleRoot(), ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	cfg := getCharacterTestConfig()

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
	migrator := postgres.NewMigratorWithPath(client, getCharacterTestMigrationsPath())
	err = migrator.Up(ctx)
	require.NoError(t, err, "Failed to run migrations")

	campaignStore := postgres.NewCampaignStore(client)
	characterStore := postgres.NewCharacterStore(client)

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Pool().Exec(ctx, "DELETE FROM characters")
		client.Pool().Exec(ctx, "DELETE FROM game_states")
		client.Pool().Exec(ctx, "DELETE FROM campaigns")
		// Note: We don't clean schema_migrations to avoid re-running migrations
		// Migrator.Up() is idempotent - it skips already applied migrations
		client.Close()
	}

	return client, campaignStore, characterStore, cleanup
}

// createTestCampaign creates a test campaign for character tests
func createTestCampaign(t *testing.T, ctx context.Context, campaignStore *postgres.CampaignStore, suffix string) *models.Campaign {
	campaign := models.NewCampaign(
		fmt.Sprintf("Character Test Campaign %s", suffix),
		fmt.Sprintf("dm-char-%s", suffix),
		"Test campaign for character tests",
	)
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create test campaign")
	return campaign
}

func TestCharacterStore_Create(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign
	campaign := createTestCampaign(t, ctx, campaignStore, "create")

	// Create player character
	character := models.NewCharacter(campaign.ID, "Test Hero", false)
	character.PlayerID = "player-001"
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(10)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create character")

	assert.NotEmpty(t, character.ID, "ID should be set")
	assert.NotZero(t, character.CreatedAt, "CreatedAt should be set")
	assert.NotZero(t, character.UpdatedAt, "UpdatedAt should be set")
	assert.False(t, character.IsNPC, "IsNPC should be false")
}

func TestCharacterStore_Create_NPC(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign
	campaign := createTestCampaign(t, ctx, campaignStore, "npc-create")

	// Create NPC
	character := models.NewCharacter(campaign.ID, "Test NPC", true)
	character.NPCType = models.NPCTypeScripted
	character.Race = "Goblin"
	character.Class = "Warrior"
	character.HP = models.NewHP(7)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create NPC")

	assert.True(t, character.IsNPC, "IsNPC should be true")
	assert.Equal(t, models.NPCTypeScripted, character.NPCType, "NPCType should be scripted")
}

func TestCharacterStore_Create_WithFullData(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign
	campaign := createTestCampaign(t, ctx, campaignStore, "full-data")

	// Create character with full data
	character := models.NewCharacter(campaign.ID, "Full Character", false)
	character.PlayerID = "player-002"
	character.Race = "Elf"
	character.Class = "Wizard"
	character.Level = 5
	character.Background = "Sage"
	character.Alignment = "Lawful Good"
	character.Abilities = &models.Abilities{
		Strength:     8,
		Dexterity:    14,
		Constitution: 12,
		Intelligence: 16,
		Wisdom:       12,
		Charisma:     10,
	}
	character.HP = &models.HP{Current: 24, Max: 24, Temp: 5}
	character.AC = 13
	character.Speed = 30
	character.Initiative = 2
	character.Skills = map[string]int{"arcana": 7, "history": 5}
	character.Saves = map[string]int{"intelligence": 4, "wisdom": 1}
	character.Equipment = []models.Equipment{
		{ID: "eq-1", Name: "Quarterstaff", Slot: "main_hand", Bonus: 0, Damage: "1d6", DamageType: "bludgeoning"},
	}
	character.Inventory = []models.Item{
		{ID: "item-1", Name: "Spellbook", Quantity: 1},
	}
	character.Conditions = []models.Condition{
		{Type: "blessed", Duration: 10, Source: "spell"},
	}

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create character with full data")

	// Retrieve and verify
	retrieved, err := characterStore.Get(ctx, character.ID)
	require.NoError(t, err, "Failed to get character")

	assert.Equal(t, "Elf", retrieved.Race)
	assert.Equal(t, "Wizard", retrieved.Class)
	assert.Equal(t, 5, retrieved.Level)
	assert.Equal(t, "Sage", retrieved.Background)
	assert.Equal(t, "Lawful Good", retrieved.Alignment)
	assert.Equal(t, 8, retrieved.Abilities.Strength)
	assert.Equal(t, 16, retrieved.Abilities.Intelligence)
	assert.Equal(t, 24, retrieved.HP.Max)
	assert.Equal(t, 5, retrieved.HP.Temp)
	assert.Equal(t, 13, retrieved.AC)
	assert.Equal(t, 7, retrieved.Skills["arcana"])
	assert.Len(t, retrieved.Equipment, 1)
	assert.Len(t, retrieved.Inventory, 1)
	assert.Len(t, retrieved.Conditions, 1)
}

func TestCharacterStore_Get(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and character
	campaign := createTestCampaign(t, ctx, campaignStore, "get")
	character := models.NewCharacter(campaign.ID, "Get Test Character", false)
	character.PlayerID = "player-003"
	character.Race = "Dwarf"
	character.Class = "Cleric"
	character.HP = models.NewHP(12)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create character")

	// Get the character
	retrieved, err := characterStore.Get(ctx, character.ID)
	require.NoError(t, err, "Failed to get character")

	assert.Equal(t, character.ID, retrieved.ID)
	assert.Equal(t, character.Name, retrieved.Name)
	assert.Equal(t, character.CampaignID, retrieved.CampaignID)
	assert.Equal(t, character.Race, retrieved.Race)
	assert.Equal(t, character.Class, retrieved.Class)
}

func TestCharacterStore_Get_NotFound(t *testing.T) {
	_, _, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := characterStore.Get(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCharacterStore_GetByCampaignAndID(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign and character
	campaign := createTestCampaign(t, ctx, campaignStore, "get-by-campaign")
	character := models.NewCharacter(campaign.ID, "Campaign Character", false)
	character.PlayerID = "player-004"
	character.Race = "Halfling"
	character.Class = "Rogue"
	character.HP = models.NewHP(8)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create character")

	// Get by campaign and ID
	retrieved, err := characterStore.GetByCampaignAndID(ctx, campaign.ID, character.ID)
	require.NoError(t, err, "Failed to get character")
	assert.Equal(t, character.ID, retrieved.ID)

	// Try wrong campaign
	_, err = characterStore.GetByCampaignAndID(ctx, uuid.New().String(), character.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound for wrong campaign")
}

func TestCharacterStore_List(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test campaign
	campaign := createTestCampaign(t, ctx, campaignStore, "list")

	// Create multiple characters
	for i := 0; i < 5; i++ {
		character := models.NewCharacter(
			campaign.ID,
			fmt.Sprintf("Character %d", i),
			i%2 == 0, // Alternate between player and NPC
		)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		if !character.IsNPC {
			character.PlayerID = fmt.Sprintf("player-%d", i)
		} else {
			character.NPCType = models.NPCTypeGenerated
		}

		err := characterStore.Create(ctx, character)
		require.NoError(t, err, "Failed to create character")
	}

	// List all
	characters, err := characterStore.List(ctx, nil)
	require.NoError(t, err, "Failed to list characters")
	assert.GreaterOrEqual(t, len(characters), 5, "Should have at least 5 characters")
}

func TestCharacterStore_List_WithFilter_CampaignID(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two campaigns
	campaign1 := createTestCampaign(t, ctx, campaignStore, "filter-1")
	campaign2 := createTestCampaign(t, ctx, campaignStore, "filter-2")

	// Create characters for campaign 1
	for i := 0; i < 3; i++ {
		character := models.NewCharacter(campaign1.ID, fmt.Sprintf("C1 Char %d", i), false)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		character.PlayerID = fmt.Sprintf("player-c1-%d", i)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Create characters for campaign 2
	for i := 0; i < 2; i++ {
		character := models.NewCharacter(campaign2.ID, fmt.Sprintf("C2 Char %d", i), false)
		character.Race = "Elf"
		character.Class = "Wizard"
		character.HP = models.NewHP(6)
		character.PlayerID = fmt.Sprintf("player-c2-%d", i)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Filter by campaign 1
	filter := &store.CharacterFilter{CampaignID: campaign1.ID}
	characters, err := characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 3, "Should have 3 characters for campaign 1")

	// Filter by campaign 2
	filter = &store.CharacterFilter{CampaignID: campaign2.ID}
	characters, err = characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 2, "Should have 2 characters for campaign 2")
}

func TestCharacterStore_List_WithFilter_IsNPC(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "npc-filter")

	// Create 3 player characters
	for i := 0; i < 3; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Player %d", i), false)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		character.PlayerID = fmt.Sprintf("player-%d", i)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Create 2 NPCs
	for i := 0; i < 2; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("NPC %d", i), true)
		character.Race = "Goblin"
		character.Class = "Warrior"
		character.HP = models.NewHP(7)
		character.NPCType = models.NPCTypeScripted
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Filter for player characters only
	isNPCFalse := false
	filter := &store.CharacterFilter{CampaignID: campaign.ID, IsNPC: &isNPCFalse}
	characters, err := characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 3, "Should have 3 player characters")

	// Filter for NPCs only
	isNPCTrue := true
	filter = &store.CharacterFilter{CampaignID: campaign.ID, IsNPC: &isNPCTrue}
	characters, err = characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 2, "Should have 2 NPCs")
}

func TestCharacterStore_List_WithFilter_PlayerID(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "player-filter")

	// Create characters for different players
	playerIDs := []string{"player-a", "player-b", "player-a"}
	for i, playerID := range playerIDs {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Char %d", i), false)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		character.PlayerID = playerID
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Filter by player-a
	filter := &store.CharacterFilter{CampaignID: campaign.ID, PlayerID: "player-a"}
	characters, err := characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 2, "Should have 2 characters for player-a")

	// Filter by player-b
	filter = &store.CharacterFilter{CampaignID: campaign.ID, PlayerID: "player-b"}
	characters, err = characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 1, "Should have 1 character for player-b")
}

func TestCharacterStore_List_WithFilter_NPCType(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "npctype-filter")

	// Create scripted NPCs
	for i := 0; i < 2; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Scripted NPC %d", i), true)
		character.Race = "Human"
		character.Class = "Commoner"
		character.HP = models.NewHP(4)
		character.NPCType = models.NPCTypeScripted
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Create generated NPCs
	for i := 0; i < 3; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Generated NPC %d", i), true)
		character.Race = "Bandit"
		character.Class = "Rogue"
		character.HP = models.NewHP(8)
		character.NPCType = models.NPCTypeGenerated
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Filter by scripted NPC type
	filter := &store.CharacterFilter{CampaignID: campaign.ID, NPCType: models.NPCTypeScripted}
	characters, err := characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 2, "Should have 2 scripted NPCs")

	// Filter by generated NPC type
	filter = &store.CharacterFilter{CampaignID: campaign.ID, NPCType: models.NPCTypeGenerated}
	characters, err = characterStore.List(ctx, filter)
	require.NoError(t, err, "Failed to list characters")
	assert.Len(t, characters, 3, "Should have 3 generated NPCs")
}

func TestCharacterStore_List_Pagination(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "pagination")

	// Create 10 characters
	for i := 0; i < 10; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Page Char %d", i), false)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		character.PlayerID = fmt.Sprintf("player-page-%d", i)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Get first page
	filter := &store.CharacterFilter{CampaignID: campaign.ID, Limit: 5, Offset: 0}
	page1, err := characterStore.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, page1, 5, "First page should have 5 characters")

	// Get second page
	filter = &store.CharacterFilter{CampaignID: campaign.ID, Limit: 5, Offset: 5}
	page2, err := characterStore.List(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, page2, 5, "Second page should have 5 characters")

	// Verify different characters on each page
	assert.NotEqual(t, page1[0].ID, page2[0].ID, "Pages should have different characters")
}

func TestCharacterStore_Update(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "update")

	// Create character
	character := models.NewCharacter(campaign.ID, "Update Test", false)
	character.PlayerID = "player-update"
	character.Race = "Human"
	character.Class = "Fighter"
	character.Level = 1
	character.HP = models.NewHP(10)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err, "Failed to create character")

	// Update character
	character.Level = 5
	character.Class = "Fighter (Champion)"
	character.HP.Max = 44
	character.HP.Current = 44
	character.AC = 16
	character.AddCondition("blessed", 10, "spell")

	err = characterStore.Update(ctx, character)
	require.NoError(t, err, "Failed to update character")

	// Verify update
	retrieved, err := characterStore.Get(ctx, character.ID)
	require.NoError(t, err, "Failed to get character")
	assert.Equal(t, 5, retrieved.Level)
	assert.Equal(t, "Fighter (Champion)", retrieved.Class)
	assert.Equal(t, 44, retrieved.HP.Max)
	assert.Equal(t, 16, retrieved.AC)
	assert.True(t, retrieved.HasCondition("blessed"))
}

func TestCharacterStore_Update_NotFound(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Need to create a campaign first for the character
	campaign := createTestCampaign(t, ctx, campaignStore, "update-notfound")

	character := models.NewCharacter(campaign.ID, "Not Found", false)
	character.ID = uuid.New().String()
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(10)

	err := characterStore.Update(ctx, character)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCharacterStore_Update_HPDamage(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "hp-damage")

	// Create character
	character := models.NewCharacter(campaign.ID, "HP Test", false)
	character.PlayerID = "player-hp"
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(20)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err)

	// Take damage
	overflow := character.TakeDamage(8)
	assert.Equal(t, 0, overflow)
	assert.Equal(t, 12, character.HP.Current)

	err = characterStore.Update(ctx, character)
	require.NoError(t, err)

	// Verify
	retrieved, err := characterStore.Get(ctx, character.ID)
	require.NoError(t, err)
	assert.Equal(t, 12, retrieved.HP.Current)
	assert.Equal(t, 20, retrieved.HP.Max)
}

func TestCharacterStore_Delete(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "delete")

	// Create character
	character := models.NewCharacter(campaign.ID, "Delete Test", false)
	character.PlayerID = "player-delete"
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(10)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err)

	// Delete character
	err = characterStore.Delete(ctx, character.ID)
	require.NoError(t, err, "Failed to delete character")

	// Verify deletion
	_, err = characterStore.Get(ctx, character.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")
}

func TestCharacterStore_Delete_NotFound(t *testing.T) {
	_, _, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := characterStore.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestCharacterStore_Count(t *testing.T) {
	_, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	campaign := createTestCampaign(t, ctx, campaignStore, "count")

	// Create 3 player characters and 2 NPCs
	for i := 0; i < 3; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("Player %d", i), false)
		character.Race = "Human"
		character.Class = "Fighter"
		character.HP = models.NewHP(10)
		character.PlayerID = fmt.Sprintf("player-count-%d", i)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		character := models.NewCharacter(campaign.ID, fmt.Sprintf("NPC %d", i), true)
		character.Race = "Goblin"
		character.Class = "Warrior"
		character.HP = models.NewHP(7)
		err := characterStore.Create(ctx, character)
		require.NoError(t, err)
	}

	// Count all in campaign
	count, err := characterStore.Count(ctx, &store.CharacterFilter{CampaignID: campaign.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count, "Should have 5 characters")

	// Count player characters only
	isNPCFalse := false
	count, err = characterStore.Count(ctx, &store.CharacterFilter{CampaignID: campaign.ID, IsNPC: &isNPCFalse})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should have 3 player characters")

	// Count NPCs only
	isNPCTrue := true
	count, err = characterStore.Count(ctx, &store.CharacterFilter{CampaignID: campaign.ID, IsNPC: &isNPCTrue})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "Should have 2 NPCs")
}

func TestCharacterStore_CascadeDelete(t *testing.T) {
	client, campaignStore, characterStore, cleanup := setupCharacterTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaign
	campaign := createTestCampaign(t, ctx, campaignStore, "cascade")

	// Create character
	character := models.NewCharacter(campaign.ID, "Cascade Test", false)
	character.PlayerID = "player-cascade"
	character.Race = "Human"
	character.Class = "Fighter"
	character.HP = models.NewHP(10)

	err := characterStore.Create(ctx, character)
	require.NoError(t, err)

	// Delete campaign (should cascade delete character)
	err = client.Pool().QueryRow(ctx, "DELETE FROM campaigns WHERE id = $1", campaign.ID).Scan()
	// Note: This might return no rows, which is fine for DELETE

	// Verify character is deleted
	_, err = characterStore.Get(ctx, character.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Character should be deleted when campaign is deleted")
}
