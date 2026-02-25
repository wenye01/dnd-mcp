package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCharacterStore is a mock implementation of CharacterStore
type MockCharacterStore struct {
	mock.Mock
}

func (m *MockCharacterStore) Create(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStore) Get(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStore) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	args := m.Called(ctx, campaignID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStore) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Character), args.Error(1)
}

func (m *MockCharacterStore) Update(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockCharacterStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCharacterStore) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// TestGetAbilityModifier tests the ability modifier calculation
// 规则参考: PHB 第7章 Ability Scores and Modifiers
func TestGetAbilityModifier(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected int
	}{
		{"Score 1", 1, -5},
		{"Score 2", 2, -4},
		{"Score 3", 3, -4},
		{"Score 4", 4, -3},
		{"Score 5", 5, -3},
		{"Score 6", 6, -2},
		{"Score 7", 7, -2},
		{"Score 8", 8, -1},
		{"Score 9", 9, -1},
		{"Score 10", 10, 0},
		{"Score 11", 11, 0},
		{"Score 12", 12, 1},
		{"Score 13", 13, 1},
		{"Score 14", 14, 2},
		{"Score 15", 15, 2},
		{"Score 16", 16, 3},
		{"Score 17", 17, 3},
		{"Score 18", 18, 4},
		{"Score 19", 19, 4},
		{"Score 20", 20, 5},
		{"Score 30", 30, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetAbilityModifier(tt.score)
			assert.Equal(t, tt.expected, result, "Ability modifier calculation for score %d", tt.score)
		})
	}
}

// TestCalculateUnarmoredAC tests the unarmored AC calculation
// 规则参考: PHB 第5章 Armor and Shields
func TestCalculateUnarmoredAC(t *testing.T) {
	tests := []struct {
		name          string
		dexterity     int
		expectedAC    int
	}{
		{"Dex 8 (mod -1)", 8, 9},
		{"Dex 10 (mod 0)", 10, 10},
		{"Dex 14 (mod +2)", 14, 12},
		{"Dex 16 (mod +3)", 16, 13},
		{"Dex 18 (mod +4)", 18, 14},
		{"Dex 20 (mod +5)", 20, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateUnarmoredAC(tt.dexterity)
			assert.Equal(t, tt.expectedAC, result)
		})
	}
}

// TestCreateCharacter tests character creation
func TestCreateCharacter(t *testing.T) {
	t.Run("create player character successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Aragorn",
			IsNPC:      false,
			PlayerID:   "player-1",
			Race:       "Human",
			Class:      "Ranger",
			Level:      5,
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.Equal(t, "Aragorn", character.Name)
		assert.Equal(t, "campaign-1", character.CampaignID)
		assert.Equal(t, "player-1", character.PlayerID)
		assert.False(t, character.IsNPC)
		assert.Equal(t, 5, character.Level)
		assert.NotEmpty(t, character.ID)
		mockStore.AssertExpectations(t)
	})

	t.Run("create NPC successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Shopkeeper",
			IsNPC:      true,
			NPCType:    models.NPCTypeScripted,
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.True(t, character.IsNPC)
		assert.Equal(t, models.NPCTypeScripted, character.NPCType)
		mockStore.AssertExpectations(t)
	})

	t.Run("create with custom abilities", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		abilities := &models.Abilities{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     8,
		}

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Strong Hero",
			IsNPC:      false,
			PlayerID:   "player-1",
			Abilities:  abilities,
		}

		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, character)
		assert.Equal(t, 16, character.Abilities.Strength)
		assert.Equal(t, 14, character.Abilities.Dexterity)

		// 验证 AC 计算（无甲 AC = 10 + 敏捷修正 = 10 + 2 = 12）
		assert.Equal(t, 12, character.AC)

		// 验证先攻（敏捷修正 = +2）
		assert.Equal(t, 2, character.Initiative)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail without campaign ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			Name:     "Test",
			IsNPC:    false,
			PlayerID: "player-1",
		}

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "campaign ID is required")
	})

	t.Run("fail without name", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			IsNPC:      false,
			PlayerID:   "player-1",
		}

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "character name is required")
	})

	t.Run("fail player character without player ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.CreateCharacterRequest{
			CampaignID: "campaign-1",
			Name:       "Test",
			IsNPC:      false,
		}

		character, err := charService.CreateCharacter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "player ID is required for player characters")
	})
}

// TestGetCharacter tests character retrieval
func TestGetCharacter(t *testing.T) {
	t.Run("get character successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedChar := &models.Character{
			ID:         "char-1",
			CampaignID: "campaign-1",
			Name:       "Test Character",
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(expectedChar, nil)

		character, err := charService.GetCharacter(context.Background(), "char-1")

		assert.NoError(t, err)
		assert.Equal(t, "char-1", character.ID)
		assert.Equal(t, "Test Character", character.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail without ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		character, err := charService.GetCharacter(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "character ID is required")
	})
}

// TestGetCharacterByCampaign tests character retrieval by campaign ID and character ID
func TestGetCharacterByCampaign(t *testing.T) {
	t.Run("get character by campaign successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedChar := &models.Character{
			ID:         "char-1",
			CampaignID: "campaign-1",
			Name:       "Test Character",
		}

		mockStore.On("GetByCampaignAndID", mock.Anything, "campaign-1", "char-1").Return(expectedChar, nil)

		character, err := charService.GetCharacterByCampaign(context.Background(), "campaign-1", "char-1")

		assert.NoError(t, err)
		assert.Equal(t, "char-1", character.ID)
		assert.Equal(t, "campaign-1", character.CampaignID)
		assert.Equal(t, "Test Character", character.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail without campaign ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		character, err := charService.GetCharacterByCampaign(context.Background(), "", "char-1")

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "campaign ID is required")
	})

	t.Run("fail without character ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		character, err := charService.GetCharacterByCampaign(context.Background(), "campaign-1", "")

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "character ID is required")
	})

	t.Run("fail when character not found in campaign", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		mockStore.On("GetByCampaignAndID", mock.Anything, "campaign-1", "char-999").Return(nil, errors.New("not found"))

		character, err := charService.GetCharacterByCampaign(context.Background(), "campaign-1", "char-999")

		assert.Error(t, err)
		assert.Nil(t, character)
		mockStore.AssertExpectations(t)
	})
}

// TestUpdateCharacter tests character update
func TestUpdateCharacter(t *testing.T) {
	t.Run("update character name", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		// 使用 NewCharacter 创建有效的角色
		existingChar := models.NewCharacter("campaign-1", "Old Name", true)

		newName := "New Name"
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Name: &newName,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "New Name", character.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("update level", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Level = 1

		newLevel := 5
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Level: &newLevel,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 5, character.Level)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail invalid level", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Level = 1

		invalidLevel := 25 // D&D 5e max level is 20
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			Level: &invalidLevel,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "invalid level")
		mockStore.AssertExpectations(t)
	})

	t.Run("fail without character ID", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		req := &service.UpdateCharacterRequest{}

		character, err := charService.UpdateCharacter(context.Background(), "", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "character ID is required")
	})

	t.Run("fail with empty name", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		emptyName := ""
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			Name: &emptyName,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "character name cannot be empty")
		mockStore.AssertExpectations(t)
	})

	t.Run("update race", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Race = "Human"

		newRace := "Elf"
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Race: &newRace,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Elf", character.Race)
		mockStore.AssertExpectations(t)
	})

	t.Run("update class", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Class = "Fighter"

		newClass := "Wizard"
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Class: &newClass,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Wizard", character.Class)
		mockStore.AssertExpectations(t)
	})

	t.Run("update background", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newBackground := "Noble"
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Background: &newBackground,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Noble", character.Background)
		mockStore.AssertExpectations(t)
	})

	t.Run("update alignment", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newAlignment := "Lawful Good"
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Alignment: &newAlignment,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Lawful Good", character.Alignment)
		mockStore.AssertExpectations(t)
	})

	t.Run("update abilities", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newAbilities := &models.Abilities{
			Strength:     18,
			Dexterity:    16,
			Constitution: 14,
			Intelligence: 12,
			Wisdom:       10,
			Charisma:     8,
		}
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Abilities: newAbilities,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 18, character.Abilities.Strength)
		assert.Equal(t, 16, character.Abilities.Dexterity)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail with invalid abilities", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		invalidAbilities := &models.Abilities{
			Strength: 50, // Invalid: max is 30
		}
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			Abilities: invalidAbilities,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "invalid abilities")
		mockStore.AssertExpectations(t)
	})

	t.Run("update HP", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newHP := &models.HP{Current: 25, Max: 30, Temp: 5}
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			HP: newHP,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 25, character.HP.Current)
		assert.Equal(t, 30, character.HP.Max)
		assert.Equal(t, 5, character.HP.Temp)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail with invalid HP", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		invalidHP := &models.HP{Current: 10, Max: 0, Temp: 0} // Invalid: max must be at least 1
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			HP: invalidHP,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "invalid HP")
		mockStore.AssertExpectations(t)
	})

	t.Run("update AC", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.AC = 10

		newAC := 18
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			AC: &newAC,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 18, character.AC)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail with negative AC", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		negativeAC := -5
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			AC: &negativeAC,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "AC cannot be negative")
		mockStore.AssertExpectations(t)
	})

	t.Run("update speed", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Speed = 30

		newSpeed := 25
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Speed: &newSpeed,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 25, character.Speed)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail with negative speed", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		negativeSpeed := -10
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		req := &service.UpdateCharacterRequest{
			Speed: &negativeSpeed,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.Error(t, err)
		assert.Nil(t, character)
		assert.Contains(t, err.Error(), "speed cannot be negative")
		mockStore.AssertExpectations(t)
	})

	t.Run("update initiative", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Initiative = 0

		newInitiative := 3
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Initiative: &newInitiative,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 3, character.Initiative)
		mockStore.AssertExpectations(t)
	})

	t.Run("update skills", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newSkills := map[string]int{
			"perception":  5,
			"stealth":     7,
			"investigation": 3,
		}
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Skills: newSkills,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 5, character.Skills["perception"])
		assert.Equal(t, 7, character.Skills["stealth"])
		assert.Equal(t, 3, character.Skills["investigation"])
		mockStore.AssertExpectations(t)
	})

	t.Run("update saves", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)

		newSaves := map[string]int{
			"dexterity":  5,
			"constitution": 3,
		}
		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Saves: newSaves,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 5, character.Saves["dexterity"])
		assert.Equal(t, 3, character.Saves["constitution"])
		mockStore.AssertExpectations(t)
	})

	t.Run("update multiple fields at once", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Race = "Human"
		existingChar.Class = "Fighter"
		existingChar.Level = 1

		newName := "Updated Hero"
		newRace := "Elf"
		newClass := "Wizard"
		newLevel := 5
		newAC := 15

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{
			Name:  &newName,
			Race:  &newRace,
			Class: &newClass,
			Level: &newLevel,
			AC:    &newAC,
		}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Hero", character.Name)
		assert.Equal(t, "Elf", character.Race)
		assert.Equal(t, "Wizard", character.Class)
		assert.Equal(t, 5, character.Level)
		assert.Equal(t, 15, character.AC)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty request does not modify character", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := models.NewCharacter("campaign-1", "Test", true)
		existingChar.Race = "Human"
		existingChar.Class = "Fighter"
		existingChar.Level = 5

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.UpdateCharacterRequest{}

		character, err := charService.UpdateCharacter(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, "Test", character.Name)
		assert.Equal(t, "Human", character.Race)
		assert.Equal(t, "Fighter", character.Class)
		assert.Equal(t, 5, character.Level)
		mockStore.AssertExpectations(t)
	})
}

// TestChangeHP tests HP management
// 规则参考: PHB 第9章 Damage and Healing
func TestChangeHP(t *testing.T) {
	t.Run("take damage", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 20, Max: 20, Temp: 0},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			Damage: 5,
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 15, character.HP.Current)
		assert.Equal(t, 20, character.HP.Max)
		mockStore.AssertExpectations(t)
	})

	t.Run("heal damage", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 10, Max: 20, Temp: 0},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			Healing: 5,
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 15, character.HP.Current)
		mockStore.AssertExpectations(t)
	})

	t.Run("heal cannot exceed max HP", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 18, Max: 20, Temp: 0},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			Healing: 10,
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 20, character.HP.Current) // Capped at max
		mockStore.AssertExpectations(t)
	})

	t.Run("damage absorbed by temp HP first", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 20, Max: 20, Temp: 5},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			Damage: 3,
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 2, character.HP.Temp)     // 5 - 3 = 2
		assert.Equal(t, 20, character.HP.Current) // Current HP unchanged
		mockStore.AssertExpectations(t)
	})

	t.Run("temp HP does not stack", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 20, Max: 20, Temp: 10},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			TempHP: 5, // Less than current temp HP
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 10, character.HP.Temp) // Keeps higher value
		mockStore.AssertExpectations(t)
	})

	t.Run("increase max HP", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			HP:   &models.HP{Current: 20, Max: 20, Temp: 0},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		req := &service.HPChangeRequest{
			MaxHPBonus: 5,
		}

		character, err := charService.ChangeHP(context.Background(), "char-1", req)

		assert.NoError(t, err)
		assert.Equal(t, 25, character.HP.Max)
		assert.Equal(t, 25, character.HP.Current) // Current also increases
		mockStore.AssertExpectations(t)
	})
}

// TestConditionManagement tests condition management
// 规则参考: PHB 附录A Conditions
func TestConditionManagement(t *testing.T) {
	t.Run("add condition", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:         "char-1",
			Name:       "Test",
			Conditions: []models.Condition{},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.AddCondition(context.Background(), "char-1", "poisoned", 10, "spell")

		assert.NoError(t, err)
		assert.Len(t, character.Conditions, 1)
		assert.Equal(t, "poisoned", character.Conditions[0].Type)
		assert.Equal(t, 10, character.Conditions[0].Duration)
		mockStore.AssertExpectations(t)
	})

	t.Run("remove condition", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
			Conditions: []models.Condition{
				{Type: "poisoned", Duration: 10, Source: "spell"},
			},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Character")).Return(nil)

		character, err := charService.RemoveCondition(context.Background(), "char-1", "poisoned")

		assert.NoError(t, err)
		assert.Len(t, character.Conditions, 0)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail to remove non-existent condition", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:         "char-1",
			Name:       "Test",
			Conditions: []models.Condition{},
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)

		character, err := charService.RemoveCondition(context.Background(), "char-1", "poisoned")

		assert.Error(t, err)
		assert.Nil(t, character)
		mockStore.AssertExpectations(t)
	})
}

// TestListCharacters tests character listing
func TestListCharacters(t *testing.T) {
	t.Run("list characters with filter", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		expectedChars := []*models.Character{
			{ID: "char-1", Name: "Character 1"},
			{ID: "char-2", Name: "Character 2"},
		}

		isNPC := false
		mockStore.On("List", mock.Anything, mock.MatchedBy(func(f *store.CharacterFilter) bool {
			return f.CampaignID == "campaign-1" && *f.IsNPC == false
		})).Return(expectedChars, nil)

		req := &service.ListCharactersRequest{
			CampaignID: "campaign-1",
			IsNPC:      &isNPC,
		}

		characters, err := charService.ListCharacters(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, characters, 2)
		mockStore.AssertExpectations(t)
	})
}

// TestDeleteCharacter tests character deletion
func TestDeleteCharacter(t *testing.T) {
	t.Run("delete character successfully", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		existingChar := &models.Character{
			ID:   "char-1",
			Name: "Test",
		}

		mockStore.On("Get", mock.Anything, "char-1").Return(existingChar, nil)
		mockStore.On("Delete", mock.Anything, "char-1").Return(nil)

		err := charService.DeleteCharacter(context.Background(), "char-1")

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("fail to delete non-existent character", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		mockStore.On("Get", mock.Anything, "char-1").Return(nil, errors.New("not found"))

		err := charService.DeleteCharacter(context.Background(), "char-1")

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}

// TestCountCharacters tests character counting
func TestCountCharacters(t *testing.T) {
	t.Run("count characters", func(t *testing.T) {
		mockStore := new(MockCharacterStore)
		charService := service.NewCharacterService(mockStore)

		mockStore.On("Count", mock.Anything, mock.AnythingOfType("*store.CharacterFilter")).Return(int64(5), nil)

		req := &service.ListCharactersRequest{
			CampaignID: "campaign-1",
		}

		count, err := charService.CountCharacters(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockStore.AssertExpectations(t)
	})
}
