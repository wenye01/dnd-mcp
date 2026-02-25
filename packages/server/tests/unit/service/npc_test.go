package service_test

import (
	"context"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNPCCreation tests NPC creation scenarios
// 规则参考: DMG 第4章 Designing NPCs
func TestNPCCreation(t *testing.T) {
	t.Run("create scripted NPC successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Barlimor Butterburr",
			IsNPC:      true,
			NPCType:    models.NPCTypeScripted,
			Race:       "Human",
			Class:      "Commoner",
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.True(t, character.IsNPC)
		assert.Equal(t, models.NPCTypeScripted, character.NPCType)
		assert.True(t, character.IsScriptedNPC())
		assert.False(t, character.IsGeneratedNPC())
		mockStore.AssertExpectations(t)
	})

	t.Run("create generated NPC successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Random Traveler",
			IsNPC:      true,
			NPCType:    models.NPCTypeGenerated,
			Race:       "Human",
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.True(t, character.IsNPC)
		assert.Equal(t, models.NPCTypeGenerated, character.NPCType)
		assert.False(t, character.IsScriptedNPC())
		assert.True(t, character.IsGeneratedNPC())
		mockStore.AssertExpectations(t)
	})

	t.Run("create NPC without type (defaults to empty)", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Generic NPC",
			IsNPC:      true,
			// NPCType not specified
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.True(t, character.IsNPC)
		assert.Equal(t, models.NPCType(""), character.NPCType)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail player character with NPC type", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Invalid Player",
			IsNPC:      false,
			PlayerID:   "player-1",
			NPCType:    models.NPCTypeScripted, // Should not be allowed
		}

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "player characters cannot have NPC type")
		mockStore.AssertExpectations(t)
	})

	t.Run("fail with invalid NPC type", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Invalid NPC",
			IsNPC:      true,
			NPCType:    models.NPCType("invalid_type"),
		}

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "invalid NPC type")
		mockStore.AssertExpectations(t)
	})

	t.Run("NPC does not require player ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Shopkeeper",
			IsNPC:      true,
			NPCType:    models.NPCTypeScripted,
			// PlayerID not specified - should be OK for NPCs
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.True(t, character.IsNPC)
		assert.Empty(t, character.PlayerID)
		mockStore.AssertExpectations(t)
	})
}

// TestNPCFiltering tests NPC filtering in list operations
func TestNPCFiltering(t *testing.T) {
	t.Run("list only NPCs", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedNPCs := []*models.Character{
			{ID: "npc-1", Name: "Shopkeeper", IsNPC: true, NPCType: models.NPCTypeScripted},
			{ID: "npc-2", Name: "Guard", IsNPC: true, NPCType: models.NPCTypeGenerated},
		}

		isNPC := true
		mockStore.On("List", mock.Anything, mock.MatchedBy(func(f *store.CharacterFilter) bool {
			return f.CampaignID == "campaign-1" && *f.IsNPC == true
		})).Return(expectedNPCs, nil)

		req := &service.ListCharactersRequest{
			CampaignID: "campaign-1",
			IsNPC:      &isNPC,
		}

		characters, err := charService.ListCharacters(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, characters, 2)
		for _, c := range characters {
			assert.True(t, c.IsNPC)
		}
		mockStore.AssertExpectations(t)
	})

	t.Run("list only player characters", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedPlayers := []*models.Character{
			{ID: "player-1", Name: "Aragorn", IsNPC: false, PlayerID: "user-1"},
			{ID: "player-2", Name: "Legolas", IsNPC: false, PlayerID: "user-2"},
		}

		isNPC := false
		mockStore.On("List", mock.Anything, mock.MatchedBy(func(f *store.CharacterFilter) bool {
			return f.CampaignID == "campaign-1" && *f.IsNPC == false
		})).Return(expectedPlayers, nil)

		req := &service.ListCharactersRequest{
			CampaignID: "campaign-1",
			IsNPC:      &isNPC,
		}

		characters, err := charService.ListCharacters(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, characters, 2)
		for _, c := range characters {
			assert.False(t, c.IsNPC)
		}
		mockStore.AssertExpectations(t)
	})

	t.Run("list NPCs by type", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedNPCs := []*models.Character{
			{ID: "npc-1", Name: "Story NPC", IsNPC: true, NPCType: models.NPCTypeScripted},
		}

		isNPC := true
		mockStore.On("List", mock.Anything, mock.MatchedBy(func(f *store.CharacterFilter) bool {
			return f.CampaignID == "campaign-1" &&
				*f.IsNPC == true &&
				f.NPCType == models.NPCTypeScripted
		})).Return(expectedNPCs, nil)

		req := &service.ListCharactersRequest{
			CampaignID: "campaign-1",
			IsNPC:      &isNPC,
			NPCType:    models.NPCTypeScripted,
		}

		characters, err := charService.ListCharacters(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, characters, 1)
		assert.Equal(t, models.NPCTypeScripted, characters[0].NPCType)
		mockStore.AssertExpectations(t)
	})
}

// TestNPCModelMethods tests NPC-related model methods
func TestNPCModelMethods(t *testing.T) {
	t.Run("IsPlayerCharacter returns correct value", func(t *testing.T) {
		playerChar := models.NewCharacter("campaign-1", "Player", false)
		playerChar.PlayerID = "player-1"
		assert.False(t, playerChar.IsNPC)
		assert.True(t, playerChar.IsPlayerCharacter())

		npc := models.NewCharacter("campaign-1", "NPC", true)
		assert.True(t, npc.IsNPC)
		assert.False(t, npc.IsPlayerCharacter())
	})

	t.Run("IsScriptedNPC returns correct value", func(t *testing.T) {
		scriptedNPC := models.NewCharacter("campaign-1", "Scripted", true)
		scriptedNPC.NPCType = models.NPCTypeScripted
		assert.True(t, scriptedNPC.IsScriptedNPC())
		assert.False(t, scriptedNPC.IsGeneratedNPC())

		generatedNPC := models.NewCharacter("campaign-1", "Generated", true)
		generatedNPC.NPCType = models.NPCTypeGenerated
		assert.False(t, generatedNPC.IsScriptedNPC())
		assert.True(t, generatedNPC.IsGeneratedNPC())

		noTypeNPC := models.NewCharacter("campaign-1", "NoType", true)
		assert.False(t, noTypeNPC.IsScriptedNPC())
		assert.False(t, noTypeNPC.IsGeneratedNPC())

		playerChar := models.NewCharacter("campaign-1", "Player", false)
		playerChar.PlayerID = "player-1"
		assert.False(t, playerChar.IsScriptedNPC())
		assert.False(t, playerChar.IsGeneratedNPC())
	})

	t.Run("NPC validation allows empty player ID", func(t *testing.T) {
		npc := models.NewCharacter("campaign-1", "NPC", true)
		npc.NPCType = models.NPCTypeScripted

		err := npc.Validate()
		assert.NoError(t, err)
	})

	t.Run("player character validation requires player ID", func(t *testing.T) {
		playerChar := models.NewCharacter("campaign-1", "Player", false)
		// PlayerID not set

		err := playerChar.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player_id")
	})
}

// TestNPCTypeConstants tests NPC type constants
func TestNPCTypeConstants(t *testing.T) {
	assert.Equal(t, models.NPCType("scripted"), models.NPCTypeScripted)
	assert.Equal(t, models.NPCType("generated"), models.NPCTypeGenerated)
}
