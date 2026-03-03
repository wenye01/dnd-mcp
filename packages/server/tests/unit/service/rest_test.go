// Package service provides unit tests for the service layer
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ErrCharacterNotFound = errors.New("character not found")
	ErrGameStateNotFound  = errors.New("game state not found")
)

// mockCharacterStore is a mock implementation of CharacterStore for testing
type mockCharacterStore struct {
	characters map[string]*models.Character
}

func newMockCharacterStore() *mockCharacterStore {
	return &mockCharacterStore{
		characters: make(map[string]*models.Character),
	}
}

func (m *mockCharacterStore) Create(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *mockCharacterStore) Get(ctx context.Context, id string) (*models.Character, error) {
	char, ok := m.characters[id]
	if !ok {
		return nil, ErrCharacterNotFound
	}
	return char, nil
}

func (m *mockCharacterStore) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	char, ok := m.characters[id]
	if !ok {
		return nil, ErrCharacterNotFound
	}
	if char.CampaignID != campaignID {
		return nil, ErrCharacterNotFound
	}
	return char, nil
}

func (m *mockCharacterStore) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	result := make([]*models.Character, 0)
	for _, char := range m.characters {
		if filter.CampaignID != "" && char.CampaignID != filter.CampaignID {
			continue
		}
		if filter.IsNPC != nil {
			if char.IsNPC != *filter.IsNPC {
				continue
			}
		}
		result = append(result, char)
	}
	return result, nil
}

func (m *mockCharacterStore) Update(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *mockCharacterStore) Delete(ctx context.Context, id string) error {
	delete(m.characters, id)
	return nil
}

func (m *mockCharacterStore) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return int64(len(m.characters)), nil
}

// mockGameStateStore is a mock implementation of GameStateStore for testing
type mockGameStateStore struct {
	gameStates map[string]*models.GameState
}

func newMockGameStateStore() *mockGameStateStore {
	return &mockGameStateStore{
		gameStates: make(map[string]*models.GameState),
	}
}

func (m *mockGameStateStore) Create(ctx context.Context, gameState *models.GameState) error {
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *mockGameStateStore) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	state, ok := m.gameStates[campaignID]
	if !ok {
		return nil, ErrGameStateNotFound
	}
	return state, nil
}

func (m *mockGameStateStore) GetByID(ctx context.Context, id string) (*models.GameState, error) {
	for _, state := range m.gameStates {
		if state.ID == id {
			return state, nil
		}
	}
	return nil, ErrGameStateNotFound
}

func (m *mockGameStateStore) Update(ctx context.Context, gameState *models.GameState) error {
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *mockGameStateStore) Delete(ctx context.Context, campaignID string) error {
	delete(m.gameStates, campaignID)
	return nil
}

// Helper function to create a test character for rest tests
func createTestCharacterForRest(campaignID, id, name string, level int, class string, currentHP, maxHP int) *models.Character {
	char := models.NewCharacter(campaignID, name, false)
	char.ID = id
	char.Level = level
	char.Class = class
	char.HP = &models.HP{
		Current: currentHP,
		Max:     maxHP,
		Temp:    0,
	}
	char.Abilities = &models.Abilities{
		Strength:     10,
		Dexterity:    10,
		Constitution: 14, // +2 modifier
		Intelligence: 10,
		Wisdom:       10,
		Charisma:     10,
	}
	return char
}

// TestShortRest_Success tests successful short rest
func TestShortRest_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character with hit dice
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 15, 45)
	character.HitDice = models.NewHitDice(5, 10) // 5d10 hit dice (Fighter)
	mockCharStore.characters[character.ID] = character

	// Execute
	req := &service.ShortRestRequest{
		CampaignID:     "campaign-1",
		CharacterID:    "char-1",
		HitDiceToSpend: 2,
	}

	result, err := restService.ShortRest(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "char-1", result.CharacterID)
	assert.Equal(t, "Test Fighter", result.CharacterName)
	assert.Equal(t, 2, result.HitDiceSpent)
	assert.Greater(t, result.HPHealed, 0)
	assert.LessOrEqual(t, result.HPHealed, 24) // 2 * (10/2 + 1 + 2) = 2 * 8 = 16 max
	assert.Equal(t, 3, result.HitDiceRemaining) // 5 - 2 = 3
	assert.Equal(t, 2, result.ConstitutionMod)  // CON 14 = +2

	// Verify character was updated
	updatedChar := mockCharStore.characters["char-1"]
	assert.Equal(t, 3, updatedChar.HitDice.Available())
	assert.Greater(t, updatedChar.HP.Current, 15)
}

// TestShortRest_NoHitDice tests short rest with no hit dice available
func TestShortRest_NoHitDice(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character with no hit dice remaining
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 15, 45)
	character.HitDice = &models.HitDice{
		Total:   5,
		Current: 0, // No hit dice remaining
		DieSize: 10,
	}
	mockCharStore.characters[character.ID] = character

	// Execute
	req := &service.ShortRestRequest{
		CampaignID:     "campaign-1",
		CharacterID:    "char-1",
		HitDiceToSpend: 1,
	}

	_, err := restService.ShortRest(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no hit dice available")
}

// TestShortRest_TooManyHitDice tests short rest requesting more hit dice than available
func TestShortRest_TooManyHitDice(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character with limited hit dice
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 15, 45)
	character.HitDice = &models.HitDice{
		Total:   5,
		Current: 2, // Only 2 hit dice remaining
		DieSize: 10,
	}
	mockCharStore.characters[character.ID] = character

	// Execute
	req := &service.ShortRestRequest{
		CampaignID:     "campaign-1",
		CharacterID:    "char-1",
		HitDiceToSpend: 5, // Request 5, only have 2
	}

	result, err := restService.ShortRest(ctx, req)

	// Assert - should spend what's available
	require.NoError(t, err)
	assert.Equal(t, 2, result.HitDiceSpent)
	assert.Equal(t, 0, result.HitDiceRemaining)
}

// TestLongRest_Success tests successful long rest
func TestLongRest_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 10, 45)
	character.HitDice = &models.HitDice{
		Total:   5,
		Current: 2, // Only 2 remaining before rest
		DieSize: 10,
	}
	character.Spellbook = models.NewSpellbook()
	character.Spellbook.Slots = map[int]*models.SpellSlots{
		1: {Total: 4, Used: 3}, // 1st level: 4 total, 3 used
		2: {Total: 3, Used: 2}, // 2nd level: 3 total, 2 used
	}
	mockCharStore.characters[character.ID] = character

	// Create game state
	gameState := models.NewGameState("campaign-1")
	mockGameStateStore.gameStates[gameState.CampaignID] = gameState

	// Execute
	req := &service.LongRestRequest{
		CampaignID:  "campaign-1",
		CharacterID: "char-1",
	}

	result, err := restService.LongRest(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "char-1", result.CharacterID)
	assert.Equal(t, "Test Fighter", result.CharacterName)
	assert.Equal(t, 35, result.HPHealed)            // 45 - 10 = 35
	assert.Equal(t, 45, result.HPCurrent)           // Full HP
	assert.Equal(t, 45, result.HPMax)
	assert.True(t, result.SpellSlotsRestored)
	assert.Greater(t, result.HitDiceRestored, 0)    // Should restore at least 1 (half of 5 = 3)
	assert.Greater(t, result.HitDiceRemaining, 2)   // Should have more than before

	// Verify spell slots were restored
	updatedChar := mockCharStore.characters["char-1"]
	assert.Equal(t, 0, updatedChar.Spellbook.Slots[1].Used)
	assert.Equal(t, 0, updatedChar.Spellbook.Slots[2].Used)

	// Verify game time advanced
	updatedState := mockGameStateStore.gameStates["campaign-1"]
	assert.GreaterOrEqual(t, updatedState.GameTime.Hour, 8) // At least 8 hours passed
}

// TestLongRest_WithFeatures tests long rest with features that restore
func TestLongRest_WithFeatures(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character with features
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 10, 45)
	character.HitDice = models.NewHitDice(5, 10)
	character.Features = []*models.Feature{
		{
			ID:          "action-surge",
			Name:        "Action Surge",
			Type:        models.FeatureTypeClass,
			Uses:        1,
			Used:        1,
			RestoreType: "short_rest", // Should restore on short rest only
		},
		{
			ID:          "second-wind",
			Name:        "Second Wind",
			Type:        models.FeatureTypeClass,
			Uses:        1,
			Used:        1,
			RestoreType: "long_rest", // Should restore on long rest
		},
		{
			ID:          "luck",
			Name:        "Lucky",
			Type:        models.FeatureTypeFeat,
			Uses:        3,
			Used:        2,
			RestoreType: "long_rest", // Should restore on long rest
		},
	}
	mockCharStore.characters[character.ID] = character

	// Create game state
	gameState := models.NewGameState("campaign-1")
	mockGameStateStore.gameStates[gameState.CampaignID] = gameState

	// Execute
	req := &service.LongRestRequest{
		CampaignID:  "campaign-1",
		CharacterID: "char-1",
	}

	result, err := restService.LongRest(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Only long_rest features should be restored
	assert.Len(t, result.FeaturesRestored, 2)
	assert.Contains(t, result.FeaturesRestored, "Second Wind")
	assert.Contains(t, result.FeaturesRestored, "Lucky")
	assert.NotContains(t, result.FeaturesRestored, "Action Surge")

	// Verify features were actually restored
	updatedChar := mockCharStore.characters["char-1"]
	for _, feature := range updatedChar.Features {
		if feature.Name == "Second Wind" || feature.Name == "Lucky" {
			assert.Equal(t, 0, feature.Used)
		}
	}
}

// TestLongRest_NoSpellbook tests long rest for character without spellbook
func TestLongRest_NoSpellbook(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character without spellbook (Fighter)
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 10, 45)
	character.HitDice = models.NewHitDice(5, 10)
	mockCharStore.characters[character.ID] = character

	// Execute
	req := &service.LongRestRequest{
		CampaignID:  "campaign-1",
		CharacterID: "char-1",
	}

	result, err := restService.LongRest(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.SpellSlotsRestored) // No spellbook to restore
	assert.Equal(t, 45, result.HPCurrent)       // Full HP
}

// TestLongRest_InvalidCharacter tests long rest with non-existent character
func TestLongRest_InvalidCharacter(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Execute
	req := &service.LongRestRequest{
		CampaignID:  "campaign-1",
		CharacterID: "non-existent",
	}

	_, err := restService.LongRest(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get character")
}

// TestPartyLongRest_Success tests successful party long rest
func TestPartyLongRest_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test characters
	char1 := createTestCharacterForRest("campaign-1", "char-1", "Fighter", 5, "fighter", 20, 45)
	char1.HitDice = models.NewHitDice(5, 10)
	char1.PlayerID = "player-1"

	char2 := createTestCharacterForRest("campaign-1", "char-2", "Cleric", 3, "cleric", 10, 27)
	char2.HitDice = models.NewHitDice(3, 8)
	char2.PlayerID = "player-2"

	// Create an NPC (should not be included in party rest)
	char3 := createTestCharacterForRest("campaign-1", "char-3", "Goblin", 1, "goblin", 5, 7)
	char3.IsNPC = true
	char3.NPCType = models.NPCTypeGenerated

	mockCharStore.characters[char1.ID] = char1
	mockCharStore.characters[char2.ID] = char2
	mockCharStore.characters[char3.ID] = char3

	// Create game state
	gameState := models.NewGameState("campaign-1")
	mockGameStateStore.gameStates[gameState.CampaignID] = gameState

	// Execute
	gameStateResult, results, err := restService.PartyLongRest(ctx, "campaign-1")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gameStateResult)
	assert.Len(t, results, 2) // Only 2 player characters

	// Check both characters were rested
	characterIDs := make([]string, 0, 2)
	for _, result := range results {
		characterIDs = append(characterIDs, result.CharacterID)
		assert.Equal(t, result.HPMax, result.HPCurrent) // Full HP
	}
	assert.Contains(t, characterIDs, "char-1")
	assert.Contains(t, characterIDs, "char-2")
	assert.NotContains(t, characterIDs, "char-3") // NPC should not be included
}

// TestShortRest_EmptyHitDice tests short rest creating hit dice if not present
func TestShortRest_EmptyHitDice(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	// Create test character without hit dice (should be auto-created)
	character := createTestCharacterForRest("campaign-1", "char-1", "Test Fighter", 5, "fighter", 15, 45)
	character.HitDice = nil // No hit dice set
	mockCharStore.characters[character.ID] = character

	// Execute
	req := &service.ShortRestRequest{
		CampaignID:     "campaign-1",
		CharacterID:    "char-1",
		HitDiceToSpend: 1,
	}

	result, err := restService.ShortRest(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 4, result.HitDiceRemaining) // Should have 5 - 1 = 4

	// Verify hit dice were created
	updatedChar := mockCharStore.characters["char-1"]
	assert.NotNil(t, updatedChar.HitDice)
	assert.Equal(t, 5, updatedChar.HitDice.Total)
	assert.Equal(t, 10, updatedChar.HitDice.DieSize) // Fighter uses d10
}

// TestShortRest_DifferentClasses tests short rest with different class hit dice
func TestShortRest_DifferentClasses(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockCharStore := newMockCharacterStore()
	mockGameStateStore := newMockGameStateStore()
	restService := service.NewRestService(mockCharStore, mockGameStateStore)

	testCases := []struct {
		class       string
		dieSize     int
		expectedMin int // Minimum HP healed (with +2 CON)
		expectedMax int // Maximum HP healed (with +2 CON)
	}{
		{"wizard", 6, 3, 6},   // d6: avg(4) + 2 = 6, min(1) + 2 = 3
		{"sorcerer", 6, 3, 6},
		{"bard", 8, 4, 7},     // d8: avg(5) + 2 = 7, min(1) + 2 = 3 (rounded to 4)
		{"cleric", 8, 4, 7},
		{"rogue", 8, 4, 7},
		{"fighter", 10, 5, 8}, // d10: avg(6) + 2 = 8, min(1) + 2 = 3 (rounded to 5)
		{"paladin", 10, 5, 8},
		{"ranger", 10, 5, 8},
		{"barbarian", 12, 6, 9}, // d12: avg(7) + 2 = 9, min(1) + 2 = 3 (rounded to 6)
	}

	for _, tc := range testCases {
		t.Run(tc.class, func(t *testing.T) {
			// Create test character
			character := createTestCharacterForRest("campaign-1", "char-"+tc.class, "Test "+tc.class, 3, tc.class, 10, 30)
			character.HitDice = models.NewHitDice(3, tc.dieSize)
			mockCharStore.characters[character.ID] = character

			// Execute
			req := &service.ShortRestRequest{
				CampaignID:     "campaign-1",
				CharacterID:    character.ID,
				HitDiceToSpend: 1,
			}

			result, err := restService.ShortRest(ctx, req)

			// Assert
			require.NoError(t, err)
			assert.GreaterOrEqual(t, result.HPHealed, tc.expectedMin)
			assert.LessOrEqual(t, result.HPHealed, tc.expectedMax)

			// Clean up for next test
			delete(mockCharStore.characters, character.ID)
		})
	}
}
