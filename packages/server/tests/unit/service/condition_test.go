package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCharacterStoreForCondition is a mock implementation of CharacterStoreForCondition
type MockCharacterStoreForCondition struct {
	mock.Mock
}

func (m *MockCharacterStoreForCondition) Get(ctx context.Context, id string) (*models.Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Character), args.Error(1)
}

func (m *MockCharacterStoreForCondition) Update(ctx context.Context, character *models.Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

// TestApplyCondition_Success tests successful condition application
func TestApplyCondition_Success(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return c.ID == "char-123" && len(c.Conditions) == 1
	})).Return(nil).Once()

	req := &service.ApplyConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
		Duration:      10,
		Source:        "spider bite",
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Applied)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, models.ConditionPoisoned, resp.Conditions[0].Type)
	assert.Equal(t, 10, resp.Conditions[0].Duration)
	mockStore.AssertExpectations(t)
}

// TestApplyCondition_DurationTracking tests condition duration tracking
func TestApplyCondition_DurationTracking(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"

	ctx := context.Background()

	// Test permanent condition (-1)
	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return len(c.Conditions) == 1 && c.Conditions[0].Duration == -1
	})).Return(nil).Once()

	req := &service.ApplyConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: models.ConditionBlinded,
		Duration:      -1,
		Source:        "spell",
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Applied)
	assert.Equal(t, -1, resp.Conditions[0].Duration)
	mockStore.AssertExpectations(t)
}

// TestApplyCondition_CharacterImmunity tests condition immunity
func TestApplyCondition_CharacterImmunity(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Traits = models.NewTraits()
	character.Traits.AddConditionImmunity(models.ConditionPoisoned)

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.ApplyConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
		Duration:      10,
		Source:        "spider bite",
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.NoError(t, err)
	assert.False(t, resp.Applied)
	assert.Contains(t, resp.Message, "immune")
	mockStore.AssertExpectations(t)
}

// TestApplyCondition_ExhaustionLevel tests exhaustion level handling
func TestApplyCondition_ExhaustionLevel(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return len(c.Conditions) == 1
	})).Return(nil).Once()

	req := &service.ApplyConditionRequest{
		CampaignID:      "campaign-123",
		CharacterID:     "char-123",
		ConditionType:   models.ConditionExhaustion,
		Duration:        -1,
		ExhaustionLevel: 2,
		Source:          "forced march",
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Applied)
	assert.Equal(t, models.ConditionExhaustion, resp.Conditions[0].Type)
	assert.Contains(t, resp.Conditions[0].Source, "Level 2")
	mockStore.AssertExpectations(t)
}

// TestApplyCondition_ExhaustionInvalidLevel tests invalid exhaustion level
func TestApplyCondition_ExhaustionInvalidLevel(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	ctx := context.Background()

	req := &service.ApplyConditionRequest{
		CampaignID:      "campaign-123",
		CharacterID:     "char-123",
		ConditionType:   models.ConditionExhaustion,
		ExhaustionLevel: 7, // Invalid: > 6
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestApplyCondition_UpdateExistingExhaustion tests updating to higher exhaustion level
func TestApplyCondition_UpdateExistingExhaustion(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{
			Type:     models.ConditionExhaustion,
			Duration: -1,
			Source:   "exhaustion (Level 1)",
		},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return len(c.Conditions) == 1 && models.ExtractExhaustionLevel(c.Conditions[0].Source) == 3
	})).Return(nil).Once()

	req := &service.ApplyConditionRequest{
		CampaignID:      "campaign-123",
		CharacterID:     "char-123",
		ConditionType:   models.ConditionExhaustion,
		Duration:        -1,
		ExhaustionLevel: 3,
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Applied)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, 3, models.ExtractExhaustionLevel(resp.Conditions[0].Source))
	mockStore.AssertExpectations(t)
}

// TestRemoveCondition_Success tests successful condition removal
func TestRemoveCondition_Success(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{
			Type:     models.ConditionPoisoned,
			Duration: 10,
			Source:   "spider bite",
		},
		{
			Type:     models.ConditionBlinded,
			Duration: 5,
			Source:   "dust",
		},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return len(c.Conditions) == 1
	})).Return(nil).Once()

	req := &service.RemoveConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.RemoveCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Removed)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, models.ConditionBlinded, resp.Conditions[0].Type)
	mockStore.AssertExpectations(t)
}

// TestRemoveCondition_RemoveAll tests removing all conditions
func TestRemoveCondition_RemoveAll(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{Type: models.ConditionPoisoned, Duration: 10},
		{Type: models.ConditionBlinded, Duration: 5},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()
	mockStore.On("Update", ctx, mock.MatchedBy(func(c *models.Character) bool {
		return len(c.Conditions) == 0
	})).Return(nil).Once()

	req := &service.RemoveConditionRequest{
		CampaignID:  "campaign-123",
		CharacterID: "char-123",
		RemoveAll:   true,
	}

	resp, err := conditionService.RemoveCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.Removed)
	assert.Equal(t, 0, len(resp.Conditions))
	mockStore.AssertExpectations(t)
}

// TestRemoveCondition_NoConditions tests removing when character has no conditions
func TestRemoveCondition_NoConditions(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.RemoveConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.RemoveCondition(ctx, req)

	assert.NoError(t, err)
	assert.False(t, resp.Removed)
	mockStore.AssertExpectations(t)
}

// TestGetConditionEffects_Poisoned tests poisoned condition effects
func TestGetConditionEffects_Poisoned(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{
			Type:     models.ConditionPoisoned,
			Duration: 10,
			Source:   "spider bite",
		},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.GetConditionEffectsRequest{
		CampaignID:  "campaign-123",
		CharacterID: "char-123",
	}

	resp, err := conditionService.GetConditionEffects(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, models.ConditionPoisoned, resp.Conditions[0].Condition)
	assert.Contains(t, resp.TotalDisadvantage, "attack_rolls")
	assert.Contains(t, resp.TotalDisadvantage, "ability_checks")
	mockStore.AssertExpectations(t)
}

// TestGetConditionEffects_Prone tests prone condition effects
func TestGetConditionEffects_Prone(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{
			Type:     models.ConditionProne,
			Duration: 1,
			Source:   "knocked down",
		},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.GetConditionEffectsRequest{
		CampaignID:  "campaign-123",
		CharacterID: "char-123",
	}

	resp, err := conditionService.GetConditionEffects(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, models.ConditionProne, resp.Conditions[0].Condition)
	assert.Contains(t, resp.TotalDisadvantage, "attack_rolls")
	assert.Contains(t, resp.TotalAdvantage, "attack_melee_within_5ft")
	mockStore.AssertExpectations(t)
}

// TestGetConditionEffects_Paralyzed tests paralyzed condition effects
func TestGetConditionEffects_Paralyzed(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{
			Type:     models.ConditionParalyzed,
			Duration: 5,
			Source:   "hold person spell",
		},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.GetConditionEffectsRequest{
		CampaignID:  "campaign-123",
		CharacterID: "char-123",
	}

	resp, err := conditionService.GetConditionEffects(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Conditions))
	assert.Equal(t, models.ConditionParalyzed, resp.Conditions[0].Condition)
	assert.True(t, resp.CannotAct)
	assert.Contains(t, resp.TotalDisadvantage, "dexterity_saves")
	assert.Contains(t, resp.TotalAdvantage, "attack_against")
	mockStore.AssertExpectations(t)
}

// TestGetConditionEffects_MultipleConditions tests multiple conditions combined
func TestGetConditionEffects_MultipleConditions(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{Type: models.ConditionPoisoned, Duration: 10},
		{Type: models.ConditionProne, Duration: 1},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.GetConditionEffectsRequest{
		CampaignID:  "campaign-123",
		CharacterID: "char-123",
	}

	resp, err := conditionService.GetConditionEffects(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Conditions))
	// Both poisoned and prone give disadvantage on attack rolls
	assert.Contains(t, resp.TotalDisadvantage, "attack_rolls")
	// Prone gives advantage on melee attacks within 5ft
	assert.Contains(t, resp.TotalAdvantage, "attack_melee_within_5ft")
	mockStore.AssertExpectations(t)
}

// TestApplyCondition_InvalidConditionType tests invalid condition type
func TestApplyCondition_InvalidConditionType(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	ctx := context.Background()

	req := &service.ApplyConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "char-123",
		ConditionType: "invalid_condition",
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestApplyCondition_MissingCharacterID tests missing character ID
func TestApplyCondition_MissingCharacterID(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	ctx := context.Background()

	req := &service.ApplyConditionRequest{
		CampaignID:    "campaign-123",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.ApplyCondition(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestRemoveCondition_CharacterNotFound tests removing condition from non-existent character
func TestRemoveCondition_CharacterNotFound(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	ctx := context.Background()

	mockStore.On("Get", ctx, "nonexistent").Return(nil, errors.New("character not found")).Once()

	req := &service.RemoveConditionRequest{
		CampaignID:    "campaign-123",
		CharacterID:   "nonexistent",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.RemoveCondition(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockStore.AssertExpectations(t)
}

// TestHasCondition_Success tests checking if character has condition
func TestHasCondition_Success(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"
	character.Conditions = []models.Condition{
		{Type: models.ConditionPoisoned, Duration: 10},
	}

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.HasConditionRequest{
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.HasCondition(ctx, req)

	assert.NoError(t, err)
	assert.True(t, resp.HasCondition)
	assert.Contains(t, resp.Message, "does have")
	mockStore.AssertExpectations(t)
}

// TestHasCondition_NotFound tests checking for condition character doesn't have
func TestHasCondition_NotFound(t *testing.T) {
	mockStore := new(MockCharacterStoreForCondition)
	conditionService := service.NewConditionService(mockStore)

	character := models.NewCharacter("campaign-123", "Test Hero", false)
	character.ID = "char-123"

	ctx := context.Background()

	mockStore.On("Get", ctx, "char-123").Return(character, nil).Once()

	req := &service.HasConditionRequest{
		CharacterID:   "char-123",
		ConditionType: models.ConditionPoisoned,
	}

	resp, err := conditionService.HasCondition(ctx, req)

	assert.NoError(t, err)
	assert.False(t, resp.HasCondition)
	assert.Contains(t, resp.Message, "does not have")
	mockStore.AssertExpectations(t)
}

// TestConditionEffectModels tests the condition effect model functions
func TestConditionEffectModels(t *testing.T) {
	// Test IsValidConditionType
	assert.True(t, models.IsValidConditionType(models.ConditionPoisoned))
	assert.True(t, models.IsValidConditionType(models.ConditionExhaustion))
	assert.False(t, models.IsValidConditionType("invalid"))

	// Test GetConditionEffect for poisoned
	effect := models.GetConditionEffect(models.ConditionPoisoned)
	assert.Contains(t, effect.Disadvantages, "attack_rolls")
	assert.Contains(t, effect.Disadvantages, "ability_checks")

	// Test GetConditionEffect for prone
	effect = models.GetConditionEffect(models.ConditionProne)
	assert.Contains(t, effect.Disadvantages, "attack_rolls")
	assert.Contains(t, effect.Advantages, "attack_melee_within_5ft")

	// Test GetExhaustionEffect
	exhaustionLevel2 := models.GetExhaustionEffect(2)
	assert.Contains(t, exhaustionLevel2.Disadvantages, "ability_checks")
	assert.Equal(t, "true", exhaustionLevel2.OtherEffects["speed_halved"])

	exhaustionLevel6 := models.GetExhaustionEffect(6)
	assert.Equal(t, "true", exhaustionLevel6.OtherEffects["death"])

	// Test ExtractExhaustionLevel
	level := models.ExtractExhaustionLevel("exhaustion (Level 3)")
	assert.Equal(t, 3, level)

	level = models.ExtractExhaustionLevel("exhaustion")
	assert.Equal(t, 1, level)
}
