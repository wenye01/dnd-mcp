package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules/dice"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCombatService_EndTurn tests the basic end turn functionality
func TestCombatService_EndTurn(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with 3 participants
	combat := models.NewCombat("campaign1", []string{"char1", "char2", "char3"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: false},
		{CharacterID: "char2", Initiative: 15, HasActed: false},
		{CharacterID: "char3", Initiative: 10, HasActed: false},
	}

	char2 := createTestCharacter("char2", "Rogue", "campaign1", 30, 14)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char2, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Combat.TurnIndex) // moved to second participant
	assert.False(t, resp.NewRound)
	assert.True(t, resp.Combat.Participants[0].HasActed) // first participant has acted
	assert.False(t, resp.Combat.Participants[1].HasActed) // second participant not yet
}

// TestCombatService_EndTurn_NewRound tests advancing to a new round
func TestCombatService_EndTurn_NewRound(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat at the end of round 1
	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: true},
		{CharacterID: "char2", Initiative: 10, HasActed: true},
	}
	combat.TurnIndex = 1 // Last participant's turn

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.NewRound)
	assert.Equal(t, 2, resp.Combat.Round)     // round increased
	assert.Equal(t, 0, resp.Combat.TurnIndex) // back to first participant

	// All participants should have their HasActed reset
	assert.False(t, resp.Combat.Participants[0].HasActed)
	assert.False(t, resp.Combat.Participants[1].HasActed)
}

// TestCombatService_EndTurn_ConditionExpires tests that conditions expire at turn end
func TestCombatService_EndTurn_ConditionExpires(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with conditions
	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{
			CharacterID: "char1",
			Initiative:  20,
			HasActed:    false,
			Conditions: []models.Condition{
				{Type: "poisoned", Duration: 1, Source: "trap"}, // expires this turn
				{Type: "blessed", Duration: 3, Source: "spell"}, // still has time
			},
		},
		{
			CharacterID: "char2",
			Initiative:  10,
			HasActed:    false,
			Conditions:  []models.Condition{},
		},
	}

	char2 := createTestCharacter("char2", "Rogue", "campaign1", 30, 14)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char2, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Check that poisoned condition expired
	assert.Len(t, resp.Combat.Participants[0].Conditions, 1)
	assert.Equal(t, "blessed", resp.Combat.Participants[0].Conditions[0].Type)
	assert.Equal(t, 2, resp.Combat.Participants[0].Conditions[0].Duration) // decremented

	// Check that combat log has condition expired entry
	var foundExpiredLog bool
	for _, entry := range resp.Combat.Log {
		if entry.Action == "condition_expired" && entry.Result == "condition poisoned expired" {
			foundExpiredLog = true
			break
		}
	}
	assert.True(t, foundExpiredLog, "should have log entry for expired condition")
}

// TestCombatService_EndTurn_NotActiveCombat tests that end turn fails for inactive combat
func TestCombatService_EndTurn_NotActiveCombat(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create finished combat
	combat := models.NewCombat("campaign1", []string{"char1"})
	combat.ID = "combat1"
	combat.End() // Mark as finished

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "combat is not active")
}

// TestCombatService_EndTurn_MultipleRounds tests multiple round transitions
func TestCombatService_EndTurn_MultipleRounds(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with 2 participants
	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: false},
		{CharacterID: "char2", Initiative: 10, HasActed: false},
	}

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)
	char2 := createTestCharacter("char2", "Rogue", "campaign1", 30, 14)

	// Setup for multiple advances
	callCount := 0
	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil).Run(func(args mock.Arguments) {
		callCount++
	})
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char2, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	// First advance: char1 -> char2
	resp1, err := svc.AdvanceTurn(context.Background(), "combat1")
	assert.NoError(t, err)
	assert.False(t, resp1.NewRound)
	assert.Equal(t, 1, resp1.Combat.TurnIndex)

	// Update combat for next call
	combat.TurnIndex = 1
	combat.Participants[0].HasActed = true

	// Second advance: char2 -> new round -> char1
	resp2, err := svc.AdvanceTurn(context.Background(), "combat1")
	assert.NoError(t, err)
	assert.True(t, resp2.NewRound)
	assert.Equal(t, 2, resp2.Combat.Round)
	assert.Equal(t, 0, resp2.Combat.TurnIndex)
}

// TestCombatService_EndTurn_EmptyCombat tests end turn with no participants
func TestCombatService_EndTurn_EmptyCombat(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with empty participants (edge case)
	combat := &models.Combat{
		ID:           "combat1",
		CampaignID:   "campaign1",
		Status:       models.CombatStatusActive,
		Round:        1,
		TurnIndex:    0,
		Participants: []models.Participant{},
		Log:          []models.CombatLogEntry{},
	}

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	// Should not error, but NewRound should be false (no participants to advance)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.NewRound)
}

// TestCombatService_EndTurn_MissingCombatID tests validation
func TestCombatService_EndTurn_MissingCombatID(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "combat ID is required")
}

// TestCombatService_EndTurn_CombatNotFound tests combat not found error
func TestCombatService_EndTurn_CombatNotFound(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	mockCombatStore.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get combat")
}

// TestCombatService_EndTurn_PermanentCondition tests that permanent conditions don't expire
func TestCombatService_EndTurn_PermanentCondition(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat with permanent condition (duration -1)
	combat := models.NewCombat("campaign1", []string{"char1", "char2"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{
			CharacterID: "char1",
			Initiative:  20,
			HasActed:    false,
			Conditions: []models.Condition{
				{Type: "blinded", Duration: -1, Source: "curse"}, // permanent
			},
		},
		{
			CharacterID: "char2",
			Initiative:  10,
			HasActed:    false,
			Conditions:  []models.Condition{},
		},
	}

	char2 := createTestCharacter("char2", "Rogue", "campaign1", 30, 14)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char2").Return(char2, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Permanent condition should still exist
	assert.Len(t, resp.Combat.Participants[0].Conditions, 1)
	assert.Equal(t, "blinded", resp.Combat.Participants[0].Conditions[0].Type)
	assert.Equal(t, -1, resp.Combat.Participants[0].Conditions[0].Duration)
}

// TestCombatService_EndTurn_LogEntry tests that log entries are created
func TestCombatService_EndTurn_LogEntry(t *testing.T) {
	mockCombatStore := NewMockCombatStore()
	mockCampaignStore := new(MockCampaignStoreForCombat)
	mockCharacterStore := new(MockCharacterStoreForCombat)
	mockDiceStore := new(MockCharacterStoreForDice)
	roller := dice.NewRoller()
	diceSvc := service.NewDiceServiceWithRoller(mockDiceStore, roller)

	// Create combat at end of round
	combat := models.NewCombat("campaign1", []string{"char1"})
	combat.ID = "combat1"
	combat.Participants = []models.Participant{
		{CharacterID: "char1", Initiative: 20, HasActed: true},
	}
	combat.TurnIndex = 0

	char1 := createTestCharacter("char1", "Fighter", "campaign1", 45, 16)

	mockCombatStore.On("Get", mock.Anything, "combat1").Return(combat, nil)
	mockCharacterStore.On("Get", mock.Anything, "char1").Return(char1, nil)
	mockCombatStore.On("Update", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewCombatServiceWithRoller(
		mockCombatStore,
		mockCharacterStore,
		mockCampaignStore,
		diceSvc,
		roller,
	)

	resp, err := svc.AdvanceTurn(context.Background(), "combat1")

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Check for new round log entry
	var foundNewRoundLog bool
	for _, entry := range resp.Combat.Log {
		if entry.Action == "new_round" {
			foundNewRoundLog = true
			assert.Contains(t, entry.Result, "Round 2 begins")
			break
		}
	}
	assert.True(t, foundNewRoundLog, "should have log entry for new round")
}
