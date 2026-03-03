// Package tools contains integration tests for context tools
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
func setupContextTools() (*tools.ContextTools, *mcp.Registry, *MockCharacterStoreForContext, *MockMessageStore, *MockGameStateStore, *MockCombatStore, *MockMapStore) {
	characterStore := NewMockCharacterStoreForContext()
	messageStore := NewMockMessageStore()
	gameStateStore := NewMockGameStateStore()
	combatStore := NewMockCombatStore()
	mapStore := NewMockMapStore()

	contextService := service.NewContextService(messageStore, characterStore, gameStateStore, combatStore, mapStore)
	contextTools := tools.NewContextTools(contextService)
	registry := mcp.NewRegistry()

	return contextTools, registry, characterStore, messageStore, gameStateStore, combatStore, mapStore
}

// MockCharacterStoreForContext is a mock for context tool testing
type MockCharacterStoreForContext struct {
	characters map[string]*models.Character
}

func NewMockCharacterStoreForContext() *MockCharacterStoreForContext {
	return &MockCharacterStoreForContext{
		characters: make(map[string]*models.Character),
	}
}

func (m *MockCharacterStoreForContext) Create(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForContext) Get(ctx context.Context, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForContext) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok || c.CampaignID != campaignID {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForContext) Update(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForContext) Delete(ctx context.Context, id string) error {
	delete(m.characters, id)
	return nil
}

func (m *MockCharacterStoreForContext) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return int64(len(m.characters)), nil
}

func (m *MockCharacterStoreForContext) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	var result []*models.Character
	for _, c := range m.characters {
		// Apply filters
		if filter.CampaignID != "" && c.CampaignID != filter.CampaignID {
			continue
		}
		if filter.IsNPC != nil && c.IsNPC != *filter.IsNPC {
			continue
		}
		if filter.PlayerID != "" && c.PlayerID != filter.PlayerID {
			continue
		}
		if filter.NPCType != "" && c.NPCType != filter.NPCType {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func createTestCampaignData(charStore *MockCharacterStoreForContext, gsStore *MockGameStateStore, combatStore *MockCombatStore, mapStore *MockMapStore, campaignID string) {
	// Create a character
	character := models.NewCharacter(campaignID, "Test Hero", false)
	character.ID = "char-001"
	character.PlayerID = "player-001"
	character.Race = "Human"
	character.Class = "Fighter"
	character.Level = 5
	character.Abilities = &models.Abilities{
		Strength:     16,
		Dexterity:    14,
		Constitution: 15,
		Intelligence: 10,
		Wisdom:       12,
		Charisma:     8,
	}
	character.HP = models.NewHP(45)
	charStore.Create(context.Background(), character)

	// Create game state
	gameState := models.NewGameState(campaignID)
	gameState.GameTime = &models.GameTime{
		Year:   1492,
		Month:  6,
		Day:    15,
		Hour:   10,
		Minute: 30,
		Phase:  "Morning",
	}
	gameState.PartyPosition = &models.Position{X: 100, Y: 200}
	gameState.Weather = "Sunny"
	gameState.CurrentMapID = "map-001"
	gsStore.Create(context.Background(), gameState)

	// Create map
	gameMap := models.NewMap(campaignID, "Test Map", models.MapTypeWorld, 100, 100, 1)
	gameMap.ID = "map-001"
	mapStore.Create(context.Background(), gameMap)

	// Create combat
	combat := models.NewCombat(campaignID, []string{"char-001", "goblin-001"})
	combat.ID = "combat-001"
	combat.Status = models.CombatStatusActive
	combat.Round = 2
	combat.TurnIndex = 1
	// Set initiatives
	for i := range combat.Participants {
		if combat.Participants[i].CharacterID == "char-001" {
			combat.Participants[i].Initiative = 15
		} else {
			combat.Participants[i].Initiative = 10
		}
	}
	combatStore.Create(context.Background(), combat)
}

func createContextMessages(msgStore *MockMessageStore, campaignID string, count int) {
	for i := 0; i < count; i++ {
		role := models.MessageRoleUser
		if i%2 == 1 {
			role = models.MessageRoleAssistant
		}
		msg := models.NewMessage(campaignID, role, fmt.Sprintf("Message %d", i+1))
		if role == models.MessageRoleUser {
			msg.PlayerID = "player-001"
		}
		msgStore.Create(context.Background(), msg)
	}
}

func TestContextTools_Register(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()

	// Register tools
	contextTools.Register(registry)

	// Verify all tools are registered
	assert.Equal(t, 3, registry.Count())

	for _, name := range tools.ContextToolNames {
		assert.True(t, registry.Has(name), "Tool %s should be registered", name)
	}
}

func TestContextTools_GetContext_Success(t *testing.T) {
	contextTools, registry, _, msgStore, gsStore, combatStore, mapStore := setupContextTools()
	charStore := NewMockCharacterStoreForContext()
	contextTools.Register(registry)

	campaignID := "campaign-001"
	createTestCampaignData(charStore, gsStore, combatStore, mapStore, campaignID)
	createContextMessages(msgStore, campaignID, 5)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":    campaignID,
		"message_limit":  10,
		"include_combat": true,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	contextData := result["context"].(map[string]interface{})
	assert.Equal(t, campaignID, contextData["campaign_id"])
	assert.NotEmpty(t, contextData["game_summary"])
	assert.NotEmpty(t, contextData["messages"])
	assert.Equal(t, float64(5), contextData["raw_message_count"])
}

func TestContextTools_GetContext_MessageLimit(t *testing.T) {
	contextTools, registry, _, msgStore, gsStore, combatStore, mapStore := setupContextTools()
	charStore := NewMockCharacterStoreForContext()
	contextTools.Register(registry)

	campaignID := "campaign-001"
	createTestCampaignData(charStore, gsStore, combatStore, mapStore, campaignID)
	createContextMessages(msgStore, campaignID, 30)

	ctx := context.Background()

	// Request only 10 messages
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":   campaignID,
		"message_limit": 10,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	contextData := result["context"].(map[string]interface{})
	// Raw count should be 30, but only 10 messages in result
	assert.Equal(t, float64(30), contextData["raw_message_count"])

	messages := contextData["messages"].([]interface{})
	assert.Len(t, messages, 10)
}

func TestContextTools_GetContext_NoCombat(t *testing.T) {
	contextTools, registry, _, msgStore, gsStore, combatStore, mapStore := setupContextTools()
	charStore := NewMockCharacterStoreForContext()
	contextTools.Register(registry)

	campaignID := "campaign-001"
	createTestCampaignData(charStore, gsStore, combatStore, mapStore, campaignID)
	createContextMessages(msgStore, campaignID, 5)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":    campaignID,
		"include_combat": false,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	contextData := result["context"].(map[string]interface{})
	gameSummary := contextData["game_summary"].(map[string]interface{})
	// Combat should not be included
	_, hasCombat := gameSummary["combat"]
	assert.False(t, hasCombat)
}

func TestContextTools_GetContext_MissingCampaignID(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})

	req := mcp.ToolRequest{
		ToolName:  "get_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "campaign ID is required")
}

func TestContextTools_GetRawContext_Success(t *testing.T) {
	contextTools, registry, charStore, msgStore, gsStore, combatStore, mapStore := setupContextTools()
	contextTools.Register(registry)

	campaignID := "campaign-001"
	createTestCampaignData(charStore, gsStore, combatStore, mapStore, campaignID)
	createContextMessages(msgStore, campaignID, 5)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": campaignID,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_raw_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result["campaign_id"])
	assert.NotEmpty(t, result["game_state"])
	assert.NotEmpty(t, result["characters"])
	assert.NotEmpty(t, result["combat"])
	assert.NotEmpty(t, result["map"])
	assert.NotEmpty(t, result["messages"])
	assert.Equal(t, float64(5), result["message_count"])
}

func TestContextTools_GetRawContext_EmptyCampaign(t *testing.T) {
	contextTools, registry, _, _, gsStore, _, _ := setupContextTools()
	contextTools.Register(registry)

	// Create an empty game state
	campaignID := "campaign-empty"
	gameState := models.NewGameState(campaignID)
	gsStore.Create(context.Background(), gameState)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": campaignID,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_raw_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	assert.Equal(t, campaignID, result["campaign_id"])
	assert.NotEmpty(t, result["game_state"])
	// Characters, combat, map, and messages might be empty
}

func TestContextTools_GetRawContext_MissingCampaignID(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})

	req := mcp.ToolRequest{
		ToolName:  "get_raw_context",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "campaign ID is required")
}

func TestContextTools_SaveMessage_UserMessage(t *testing.T) {
	contextTools, registry, _, msgStore, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"role":        "user",
		"content":     "Hello, this is a test message",
		"player_id":   "player-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	assert.Equal(t, "campaign-001", message["campaign_id"])
	assert.Equal(t, "user", message["role"])
	assert.Equal(t, "Hello, this is a test message", message["content"])
	assert.Equal(t, "player-001", message["player_id"])

	// Verify message was stored
	messages, _ := msgStore.ListByCampaign(ctx, "campaign-001", 0)
	assert.Len(t, messages, 1)
}

func TestContextTools_SaveMessage_AssistantMessage(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	toolCalls := []map[string]interface{}{
		{
			"id":   "tc-001",
			"name": "roll_dice",
			"arguments": map[string]interface{}{
				"formula": "1d20+5",
			},
		},
	}

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"role":        "assistant",
		"content":     "I rolled the dice for you",
		"tool_calls":  toolCalls,
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	assert.Equal(t, "assistant", message["role"])
	assert.NotEmpty(t, message["tool_calls"])

	toolCallsResult := message["tool_calls"].([]interface{})
	assert.Len(t, toolCallsResult, 1)
}

func TestContextTools_SaveMessage_SystemMessage(t *testing.T) {
	contextTools, registry, _, msgStore, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"role":        "system",
		"content":     "System initialization complete",
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	message := result["message"].(map[string]interface{})
	assert.Equal(t, "system", message["role"])

	// Verify message was stored
	messages, _ := msgStore.ListByCampaign(ctx, "campaign-001", 0)
	assert.Len(t, messages, 1)
}

func TestContextTools_SaveMessage_InvalidRole(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"role":        "invalid_role",
		"content":     "Test message",
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid message")
}

func TestContextTools_SaveMessage_MissingRequired(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	// Missing role
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"content":     "Test message",
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "role")
}

func TestContextTools_SaveMessage_UserWithoutPlayerID(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"role":        "user",
		"content":     "Test message",
	})

	req := mcp.ToolRequest{
		ToolName:  "save_message",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "player_id")
}

func TestContextTools_InvalidJSON(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	ctx := context.Background()

	req := mcp.ToolRequest{
		ToolName:  "get_context",
		Arguments: []byte(`{invalid json}`),
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid arguments")
}

func TestContextTools_ListTools(t *testing.T) {
	contextTools, registry, _, _, _, _, _ := setupContextTools()
	contextTools.Register(registry)

	toolList := registry.List()
	assert.Len(t, toolList, 3)

	// Verify tool definitions
	toolNames := make(map[string]bool)
	for _, tool := range toolList {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Description)
		assert.NotEmpty(t, tool.InputSchema.Type)
	}

	assert.True(t, toolNames["get_context"])
	assert.True(t, toolNames["get_raw_context"])
	assert.True(t, toolNames["save_message"])
}
