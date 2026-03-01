// Package tools contains integration tests for dice tools
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/rules/dice"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCharacterStoreForDice for dice tool testing
type MockCharacterStoreForDice struct {
	characters map[string]*models.Character
}

func NewMockCharacterStoreForDice() *MockCharacterStoreForDice {
	return &MockCharacterStoreForDice{
		characters: make(map[string]*models.Character),
	}
}

func (m *MockCharacterStoreForDice) Create(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForDice) Get(ctx context.Context, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForDice) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok || c.CampaignID != campaignID {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForDice) Update(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForDice) Delete(ctx context.Context, id string) error {
	delete(m.characters, id)
	return nil
}

func (m *MockCharacterStoreForDice) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return int64(len(m.characters)), nil
}

func (m *MockCharacterStoreForDice) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
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

// Test fixtures
func setupDiceTools() (*tools.DiceTools, *mcp.Registry, *MockCharacterStoreForDice) {
	characterStore := NewMockCharacterStoreForDice()
	roller := dice.NewRollerWithSource(dice.NewSeededRandomSource(42)) // Seeded for reproducibility
	diceService := service.NewDiceServiceWithRoller(characterStore, roller)
	diceTools := tools.NewDiceTools(diceService)
	registry := mcp.NewRegistry()

	return diceTools, registry, characterStore
}

func createTestCharacter(store *MockCharacterStoreForDice) *models.Character {
	character := models.NewCharacter("campaign-001", "Test Hero", false)
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
	character.SkillsDetail = map[string]*models.Skill{
		"athletics":  {Proficient: true},
		"perception": {Proficient: true},
	}
	character.SavesDetail = map[string]*models.Save{
		"strength":     {Proficient: true},
		"constitution": {Proficient: true},
	}
	character.HP = models.NewHP(45)
	store.Create(context.Background(), character)
	return character
}

func TestDiceTools_Register(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()

	// Register tools
	diceTools.Register(registry)

	// Verify all tools are registered
	assert.Equal(t, 3, registry.Count())

	for _, name := range tools.DiceToolNames {
		assert.True(t, registry.Has(name), "Tool %s should be registered", name)
	}
}

func TestDiceTools_RollDice_Simple(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Roll a d20
	args, _ := json.Marshal(map[string]interface{}{
		"formula": "1d20+5",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	diceResult := result["result"].(map[string]interface{})
	assert.Equal(t, "1d20+5", diceResult["formula"])
	assert.NotEmpty(t, diceResult["rolls"])

	// Verify total is roll + modifier
	rolls := diceResult["rolls"].([]interface{})
	modifier := int(diceResult["modifier"].(float64))
	total := int(diceResult["total"].(float64))

	rollSum := 0
	for _, r := range rolls {
		rollSum += int(r.(float64))
	}
	assert.Equal(t, rollSum+modifier, total)
}

func TestDiceTools_RollDice_MultipleDice(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Roll 2d6
	args, _ := json.Marshal(map[string]interface{}{
		"formula": "2d6",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	diceResult := result["result"].(map[string]interface{})
	rolls := diceResult["rolls"].([]interface{})
	assert.Len(t, rolls, 2)
}

func TestDiceTools_RollDice_KeepHighest(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Roll 4d6kh3 (ability score generation)
	args, _ := json.Marshal(map[string]interface{}{
		"formula": "4d6kh3",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	diceResult := result["result"].(map[string]interface{})
	rolls := diceResult["rolls"].([]interface{})
	// Should keep 3 highest rolls
	assert.Len(t, rolls, 3)
}

func TestDiceTools_RollDice_InvalidFormula(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Invalid formula
	args, _ := json.Marshal(map[string]interface{}{
		"formula": "invalid",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid formula")
}

func TestDiceTools_RollDice_MissingFormula(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Missing formula
	args, _ := json.Marshal(map[string]interface{}{})

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "formula is required")
}

func TestDiceTools_RollCheck_AbilityCheck(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll strength check
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "strength",
		"dc":           15,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "strength", checkResult["ability"])
	assert.Equal(t, float64(15), checkResult["dc"])

	// Verify success/failure is set
	_, hasSuccess := checkResult["success"]
	assert.True(t, hasSuccess)

	// Verify dice result exists
	diceResult := checkResult["dice_result"].(map[string]interface{})
	assert.NotEmpty(t, diceResult["rolls"])

	// Verify message
	assert.Contains(t, result["message"], "strength check")
}

func TestDiceTools_RollCheck_SkillCheck(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll athletics check (proficient skill)
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "strength",
		"skill":        "athletics",
		"dc":           10,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "athletics", checkResult["skill"])
	assert.Contains(t, result["message"], "athletics")
}

func TestDiceTools_RollCheck_WithAdvantage(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll with advantage
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "dexterity",
		"advantage":    true,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "dexterity", checkResult["ability"])
}

func TestDiceTools_RollCheck_WithDisadvantage(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll with disadvantage
	args, _ := json.Marshal(map[string]interface{}{
		"character_id":  "char-001",
		"ability":       "dexterity",
		"disadvantage":  true,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "dexterity", checkResult["ability"])
}

func TestDiceTools_RollCheck_CharacterNotFound(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "non-existent",
		"ability":      "strength",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "character not found")
}

func TestDiceTools_RollCheck_InvalidAbility(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "invalid_ability",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid ability")
}

func TestDiceTools_RollCheck_MissingRequired(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Missing ability
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_check",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "ability is required")
}

func TestDiceTools_RollSave_Success(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll constitution save (proficient)
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "constitution",
		"dc":           12,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_save",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "constitution", checkResult["ability"])
	assert.Equal(t, float64(12), checkResult["dc"])

	// Verify message
	assert.Contains(t, result["message"], "saving throw")
}

func TestDiceTools_RollSave_WithAdvantage(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	// Roll dexterity save with advantage
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "dexterity",
		"advantage":    true,
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_save",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	checkResult := result["result"].(map[string]interface{})
	assert.Equal(t, "dexterity", checkResult["ability"])
}

func TestDiceTools_RollSave_CharacterNotFound(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "non-existent",
		"ability":      "strength",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_save",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "character not found")
}

func TestDiceTools_RollSave_InvalidAbility(t *testing.T) {
	diceTools, registry, charStore := setupDiceTools()
	diceTools.Register(registry)
	createTestCharacter(charStore)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
		"ability":      "not_an_ability",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_save",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid ability")
}

func TestDiceTools_RollSave_MissingRequired(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	// Missing character_id
	args, _ := json.Marshal(map[string]interface{}{
		"ability": "strength",
	})

	req := mcp.ToolRequest{
		ToolName:  "roll_save",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "character_id is required")
}

func TestDiceTools_InvalidJSON(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	req := mcp.ToolRequest{
		ToolName:  "roll_dice",
		Arguments: []byte(`{invalid json}`),
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid arguments")
}

func TestDiceTools_UnknownTool(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})
	req := mcp.ToolRequest{
		ToolName:  "unknown_dice_tool",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "unknown tool")
}

func TestDiceTools_ListTools(t *testing.T) {
	diceTools, registry, _ := setupDiceTools()
	diceTools.Register(registry)

	toolList := registry.List()
	assert.Len(t, toolList, 3)

	// Verify tool definitions
	toolNames := make(map[string]bool)
	for _, tool := range toolList {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Description)
		assert.NotEmpty(t, tool.InputSchema.Type)
	}

	assert.True(t, toolNames["roll_dice"])
	assert.True(t, toolNames["roll_check"])
	assert.True(t, toolNames["roll_save"])
}
