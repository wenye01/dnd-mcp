package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMessageStoreForContext is a mock implementation of MessageStoreForContext
type MockMessageStoreForContext struct {
	messages map[string][]*models.Message
}

func NewMockMessageStoreForContext() *MockMessageStoreForContext {
	return &MockMessageStoreForContext{
		messages: make(map[string][]*models.Message),
	}
}

func (m *MockMessageStoreForContext) Create(ctx context.Context, message *models.Message) error {
	campaignMessages := m.messages[message.CampaignID]
	m.messages[message.CampaignID] = append(campaignMessages, message)
	return nil
}

func (m *MockMessageStoreForContext) ListByCampaign(ctx context.Context, campaignID string, limit int) ([]*models.Message, error) {
	messages, ok := m.messages[campaignID]
	if !ok {
		return []*models.Message{}, nil
	}
	return messages, nil
}

func (m *MockMessageStoreForContext) CountByCampaign(ctx context.Context, campaignID string) (int, error) {
	messages, ok := m.messages[campaignID]
	if !ok {
		return 0, nil
	}
	return len(messages), nil
}

// MockCharacterStoreForContext is a mock implementation for context service
type MockCharacterStoreForContext struct {
	characters map[string][]*models.Character
}

func NewMockCharacterStoreForContext() *MockCharacterStoreForContext {
	return &MockCharacterStoreForContext{
		characters: make(map[string][]*models.Character),
	}
}

func (m *MockCharacterStoreForContext) Create(ctx context.Context, character *models.Character) error {
	return nil
}

func (m *MockCharacterStoreForContext) Get(ctx context.Context, id string) (*models.Character, error) {
	return nil, errors.New("not found")
}

func (m *MockCharacterStoreForContext) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	return nil, errors.New("not found")
}

func (m *MockCharacterStoreForContext) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	if filter == nil {
		return []*models.Character{}, nil
	}

	characters, ok := m.characters[filter.CampaignID]
	if !ok {
		return []*models.Character{}, nil
	}

	// Filter by IsNPC if specified
	if filter.IsNPC != nil {
		result := make([]*models.Character, 0)
		for _, char := range characters {
			if char.IsNPC == *filter.IsNPC {
				result = append(result, char)
			}
		}
		return result, nil
	}

	return characters, nil
}

func (m *MockCharacterStoreForContext) Update(ctx context.Context, character *models.Character) error {
	return nil
}

func (m *MockCharacterStoreForContext) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockCharacterStoreForContext) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return 0, nil
}

// MockGameStateStoreForContext is a mock implementation for context service
type MockGameStateStoreForContext struct {
	gameStates map[string]*models.GameState
}

func NewMockGameStateStoreForContext() *MockGameStateStoreForContext {
	return &MockGameStateStoreForContext{
		gameStates: make(map[string]*models.GameState),
	}
}

func (m *MockGameStateStoreForContext) Create(ctx context.Context, gameState *models.GameState) error {
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *MockGameStateStoreForContext) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	gameState, ok := m.gameStates[campaignID]
	if !ok {
		return nil, errors.New("not found")
	}
	return gameState, nil
}

func (m *MockGameStateStoreForContext) GetByID(ctx context.Context, id string) (*models.GameState, error) {
	return nil, errors.New("not found")
}

func (m *MockGameStateStoreForContext) Update(ctx context.Context, gameState *models.GameState) error {
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *MockGameStateStoreForContext) Delete(ctx context.Context, campaignID string) error {
	delete(m.gameStates, campaignID)
	return nil
}

// MockCombatStoreForContext is a mock implementation for context service
type MockCombatStoreForContext struct {
	combats       map[string]*models.Combat
	activeCombats map[string]*models.Combat
}

func NewMockCombatStoreForContext() *MockCombatStoreForContext {
	return &MockCombatStoreForContext{
		combats:       make(map[string]*models.Combat),
		activeCombats: make(map[string]*models.Combat),
	}
}

func (m *MockCombatStoreForContext) Create(ctx context.Context, combat *models.Combat) error {
	m.combats[combat.ID] = combat
	return nil
}

func (m *MockCombatStoreForContext) Get(ctx context.Context, id string) (*models.Combat, error) {
	combat, ok := m.combats[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return combat, nil
}

func (m *MockCombatStoreForContext) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error) {
	return []*models.Combat{}, nil
}

func (m *MockCombatStoreForContext) GetActive(ctx context.Context, campaignID string) (*models.Combat, error) {
	combat, ok := m.activeCombats[campaignID]
	if !ok {
		return nil, errors.New("not found")
	}
	return combat, nil
}

func (m *MockCombatStoreForContext) Update(ctx context.Context, combat *models.Combat) error {
	return nil
}

func (m *MockCombatStoreForContext) Delete(ctx context.Context, id string) error {
	return nil
}

// MockMapStoreForContext is a mock implementation for context service
type MockMapStoreForContext struct {
	maps map[string]*models.Map
}

func NewMockMapStoreForContext() *MockMapStoreForContext {
	return &MockMapStoreForContext{
		maps: make(map[string]*models.Map),
	}
}

func (m *MockMapStoreForContext) Create(ctx context.Context, gameMap *models.Map) error {
	m.maps[gameMap.ID] = gameMap
	return nil
}

func (m *MockMapStoreForContext) Get(ctx context.Context, id string) (*models.Map, error) {
	gameMap, ok := m.maps[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return gameMap, nil
}

func (m *MockMapStoreForContext) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error) {
	return []*models.Map{}, nil
}

func (m *MockMapStoreForContext) Update(ctx context.Context, gameMap *models.Map) error {
	return nil
}

func (m *MockMapStoreForContext) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockMapStoreForContext) GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error) {
	return nil, errors.New("not found")
}

func (m *MockMapStoreForContext) GetBattleMap(ctx context.Context, id string) (*models.Map, error) {
	return nil, errors.New("not found")
}

func (m *MockMapStoreForContext) GetByParent(ctx context.Context, parentID string) ([]*models.Map, error) {
	return []*models.Map{}, nil
}

// Test fixtures
func setupContextService() (*service.ContextService, *MockMessageStoreForContext, *MockCharacterStoreForContext, *MockGameStateStoreForContext, *MockCombatStoreForContext, *MockMapStoreForContext) {
	messageStore := NewMockMessageStoreForContext()
	characterStore := NewMockCharacterStoreForContext()
	gameStateStore := NewMockGameStateStoreForContext()
	combatStore := NewMockCombatStoreForContext()
	mapStore := NewMockMapStoreForContext()

	svc := service.NewContextService(messageStore, characterStore, gameStateStore, combatStore, mapStore)
	return svc, messageStore, characterStore, gameStateStore, combatStore, mapStore
}

func TestContextService_GetContext_Basic(t *testing.T) {
	svc, msgStore, charStore, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gameState.GameTime = &models.GameTime{
		Year: 1492, Month: 5, Day: 15, Hour: 14, Minute: 30,
		Phase: models.TimePhaseAfternoon,
	}
	gameState.Weather = "Rainy"
	gsStore.Create(ctx, gameState)

	// Set up characters
	fighter := models.NewCharacter(campaignID, "Aragorn", false)
	fighter.Class = "Fighter"
	fighter.Level = 5
	fighter.HP = &models.HP{Current: 45, Max: 50}
	mage := models.NewCharacter(campaignID, "Gandalf", false)
	mage.Class = "Wizard"
	mage.Level = 10
	mage.HP = &models.HP{Current: 55, Max: 80} // 55/80 = 68.75% -> Lightly wounded
	charStore.characters[campaignID] = append(charStore.characters[campaignID], fighter, mage)

	// Set up messages
	for i := 0; i < 25; i++ {
		msg := models.NewMessage(campaignID, models.MessageRoleUser, "Test message")
		msgStore.Create(ctx, msg)
	}

	// Get context with limit of 10
	result, err := svc.GetContext(ctx, campaignID, 10, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, campaignID, result.CampaignID)
	assert.NotNil(t, result.GameSummary)

	// Check game summary
	assert.Contains(t, result.GameSummary.Time, "1492")
	assert.Contains(t, result.GameSummary.Time, "5")
	assert.Equal(t, "Rainy", result.GameSummary.Weather)
	assert.False(t, result.GameSummary.InCombat)

	// Check party members
	assert.Len(t, result.GameSummary.Party, 2)
	assert.Equal(t, "Aragorn", result.GameSummary.Party[0].Name)
	assert.Equal(t, "Fighter", result.GameSummary.Party[0].Class)
	assert.Equal(t, "Healthy", result.GameSummary.Party[0].HP)
	assert.Equal(t, "Gandalf", result.GameSummary.Party[1].Name)
	assert.Equal(t, "Wizard", result.GameSummary.Party[1].Class)
	assert.Equal(t, "Lightly wounded", result.GameSummary.Party[1].HP)

	// Check sliding window
	assert.Equal(t, 25, result.RawMessageCount)
	assert.Len(t, result.Messages, 10)

	// Check token estimate
	assert.Greater(t, result.TokenEstimate, 0)
}

func TestContextService_GetContext_NoCompressionNeeded(t *testing.T) {
	svc, msgStore, _, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gsStore.Create(ctx, gameState)

	// Add only 5 messages (less than default window)
	for i := 0; i < 5; i++ {
		msg := models.NewMessage(campaignID, models.MessageRoleUser, "Test message")
		_ = msgStore.Create(ctx, msg)
	}

	// Get context with default window
	result, err := svc.GetContext(ctx, campaignID, 0, false)
	require.NoError(t, err)

	assert.Equal(t, 5, result.RawMessageCount)
	assert.Len(t, result.Messages, 5) // All messages returned, no compression
}

func TestContextService_GetContext_WithCombat(t *testing.T) {
	svc, _, charStore, gsStore, combatStore, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state in combat
	gameState := models.NewGameState(campaignID)
	gameState.ActiveCombatID = "combat-001"
	gsStore.Create(ctx, gameState)

	// Set up combat
	combat := models.NewCombat(campaignID, []string{"char-001", "char-002"})
	combat.Round = 3
	combat.TurnIndex = 1
	combatStore.combats[combat.ID] = combat
	combatStore.activeCombats[campaignID] = combat

	// Set up characters matching combat participants
	char1 := models.NewCharacter(campaignID, "Hero1", false)
	char1.ID = "char-001"
	char1.Class = "Fighter"
	char1.HP = &models.HP{Current: 30, Max: 50}
	char2 := models.NewCharacter(campaignID, "Hero2", false)
	char2.ID = "char-002"
	char2.Class = "Cleric"
	char2.HP = &models.HP{Current: 25, Max: 35}
	charStore.characters[campaignID] = append(charStore.characters[campaignID], char1, char2)

	// Get context with combat
	result, err := svc.GetContext(ctx, campaignID, 10, true)
	require.NoError(t, err)

	assert.True(t, result.GameSummary.InCombat)
	assert.NotNil(t, result.GameSummary.Combat)
	assert.Equal(t, 3, result.GameSummary.Combat.Round)
	assert.Equal(t, 1, result.GameSummary.Combat.TurnIndex)
	assert.Len(t, result.GameSummary.Combat.Participants, 2)
}

func TestContextService_GetContext_InvalidCampaignID(t *testing.T) {
	svc, _, _, _, _, _ := setupContextService()
	ctx := context.Background()

	_, err := svc.GetContext(ctx, "", 10, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "campaign ID is required")
}

func TestContextService_GetRawContext(t *testing.T) {
	svc, msgStore, charStore, gsStore, combatStore, mapStore := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gameState.CurrentMapID = "map-001"
	gsStore.Create(ctx, gameState)

	// Set up characters
	char := models.NewCharacter(campaignID, "Hero", false)
	charStore.characters[campaignID] = append(charStore.characters[campaignID], char)

	// Set up combat
	combat := models.NewCombat(campaignID, []string{"char-001"})
	combatStore.combats[combat.ID] = combat
	combatStore.activeCombats[campaignID] = combat

	// Set up map
	gameMap := models.NewWorldMap(campaignID, "Test Map", 100, 100)
	gameMap.ID = "map-001"
	mapStore.maps[gameMap.ID] = gameMap

	// Set up messages
	msg := models.NewMessage(campaignID, models.MessageRoleUser, "Test message")
	msgStore.Create(ctx, msg)

	// Get raw context
	result, err := svc.GetRawContext(ctx, campaignID)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result.CampaignID)
	assert.NotNil(t, result.GameState)
	assert.NotNil(t, result.Characters)
	assert.Len(t, result.Characters, 1)
	assert.NotNil(t, result.Combat)
	assert.NotNil(t, result.Map)
	assert.NotNil(t, result.Messages)
	assert.Equal(t, 1, result.MessageCount)
}

func TestContextService_GetRawContext_InvalidCampaignID(t *testing.T) {
	svc, _, _, _, _, _ := setupContextService()
	ctx := context.Background()

	_, err := svc.GetRawContext(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "campaign ID is required")
}

func TestContextService_SaveMessage(t *testing.T) {
	svc, msgStore, _, _, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Create valid message with player_id (required for user messages)
	msg := models.NewUserMessage(campaignID, "player-001", "Hello, world!")
	err := svc.SaveMessage(ctx, msg)
	require.NoError(t, err)

	// Verify message was saved
	messages, _ := msgStore.ListByCampaign(ctx, campaignID, 0)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Hello, world!", messages[0].Content)
}

func TestContextService_SaveMessage_InvalidMessage(t *testing.T) {
	svc, _, _, _, _, _ := setupContextService()
	ctx := context.Background()

	// Nil message
	err := svc.SaveMessage(ctx, nil)
	assert.Error(t, err)

	// Invalid message
	msg := &models.Message{
		CampaignID: "",
		Role:       models.MessageRoleUser,
		Content:    "test",
	}
	err = svc.SaveMessage(ctx, msg)
	assert.Error(t, err)
}

func TestContextService_SetDefaultWindow(t *testing.T) {
	svc, msgStore, _, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gsStore.Create(ctx, gameState)

	// Add 30 messages
	for i := 0; i < 30; i++ {
		msg := models.NewMessage(campaignID, models.MessageRoleUser, "Test message")
		msgStore.Create(ctx, msg)
	}

	// Change default window to 5
	svc.SetDefaultWindow(5)

	// Get context with default window (0 means use default)
	result, err := svc.GetContext(ctx, campaignID, 0, false)
	require.NoError(t, err)

	assert.Equal(t, 30, result.RawMessageCount)
	assert.Len(t, result.Messages, 5) // Should use new default of 5
}

func TestContextService_FormatContextForLLM(t *testing.T) {
	svc, _, _, _, _, _ := setupContextService()

	ctx := &models.Context{
		CampaignID: uuid.New().String(),
		GameSummary: &models.GameSummary{
			Time:     "Year 1492, Month 5, Day 15",
			Location: "Forest of Shadows",
			Weather:  "Rainy",
			InCombat: false,
			Party: []models.PartyMember{
				{ID: "char1", Name: "Aragorn", Class: "Fighter", HP: "Healthy"},
			},
		},
		Messages: []models.Message{
			{Role: models.MessageRoleUser, Content: "Hello"},
			{Role: models.MessageRoleAssistant, Content: "Hi there!"},
		},
		RawMessageCount: 2,
		TokenEstimate:   10,
	}

	formatted := svc.FormatContextForLLM(ctx)
	assert.Contains(t, formatted, "=== Game State Summary ===")
	assert.Contains(t, formatted, "Year 1492, Month 5, Day 15")
	assert.Contains(t, formatted, "Forest of Shadows")
	assert.Contains(t, formatted, "Rainy")
	assert.Contains(t, formatted, "Aragorn")
	assert.Contains(t, formatted, "Fighter")
	assert.Contains(t, formatted, "Hello")
	assert.Contains(t, formatted, "Hi there!")
}

func TestContextService_GetContext_EmptyCampaign(t *testing.T) {
	svc, _, _, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state but no characters or messages
	gameState := models.NewGameState(campaignID)
	gsStore.Create(ctx, gameState)

	// Get context
	result, err := svc.GetContext(ctx, campaignID, 10, false)
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.NotNil(t, result.GameSummary)
	assert.Empty(t, result.GameSummary.Party) // No party members
	assert.Empty(t, result.Messages)          // No messages
	assert.Equal(t, 0, result.RawMessageCount)
}

func TestContextService_GetContext_NPCsExcludedFromParty(t *testing.T) {
	svc, _, charStore, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gsStore.Create(ctx, gameState)

	// Set up characters with mix of PCs and NPCs
	pc := models.NewCharacter(campaignID, "Hero", false)
	pc.Class = "Fighter"
	pc.HP = &models.HP{Current: 50, Max: 50}

	npc := models.NewCharacter(campaignID, "Goblin", true)
	npc.NPCType = models.NPCTypeGenerated
	npc.Class = "Monster"
	npc.HP = &models.HP{Current: 10, Max: 10}

	charStore.characters[campaignID] = append(charStore.characters[campaignID], pc, npc)

	// Get context
	result, err := svc.GetContext(ctx, campaignID, 10, false)
	require.NoError(t, err)

	// Only PC should be in party summary
	assert.Len(t, result.GameSummary.Party, 1)
	assert.Equal(t, "Hero", result.GameSummary.Party[0].Name)
}

func TestContextService_HPStatusDescription(t *testing.T) {
	svc, _, charStore, gsStore, _, _ := setupContextService()
	ctx := context.Background()
	campaignID := uuid.New().String()

	// Set up game state
	gameState := models.NewGameState(campaignID)
	gsStore.Create(ctx, gameState)

	// Set up characters with different HP states
	healthy := models.NewCharacter(campaignID, "HealthyHero", false)
	healthy.HP = &models.HP{Current: 50, Max: 50}

	wounded := models.NewCharacter(campaignID, "WoundedHero", false)
	wounded.HP = &models.HP{Current: 35, Max: 50}

	critical := models.NewCharacter(campaignID, "CriticalHero", false)
	critical.HP = &models.HP{Current: 10, Max: 50}

	zeroHP := models.NewCharacter(campaignID, "DownedHero", false)
	zeroHP.HP = &models.HP{Current: 0, Max: 50}

	charStore.characters[campaignID] = append(charStore.characters[campaignID],
		healthy, wounded, critical, zeroHP)

	// Get context
	result, err := svc.GetContext(ctx, campaignID, 10, false)
	require.NoError(t, err)

	assert.Len(t, result.GameSummary.Party, 4)

	// Find each character's HP status
	hpStatus := make(map[string]string)
	for _, member := range result.GameSummary.Party {
		hpStatus[member.Name] = member.HP
	}

	assert.Equal(t, "Healthy", hpStatus["HealthyHero"])
	assert.Equal(t, "Lightly wounded", hpStatus["WoundedHero"])
	assert.Equal(t, "Critical", hpStatus["CriticalHero"])
	assert.Equal(t, "Unconscious/Dead", hpStatus["DownedHero"])
}
