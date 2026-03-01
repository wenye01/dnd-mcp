// Package tools contains integration tests for character tools
package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures
func setupCharacterTools() (*tools.CharacterTools, *mcp.Registry, *MockCharacterStore) {
	characterStore := NewMockCharacterStore()
	characterService := service.NewCharacterService(characterStore)
	characterTools := tools.NewCharacterTools(characterService)
	registry := mcp.NewRegistry()

	return characterTools, registry, characterStore
}

func TestCharacterTools_Register(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()

	// Register tools
	characterTools.Register(registry)

	// Verify all tools are registered
	assert.Equal(t, 5, registry.Count())

	for _, name := range tools.CharacterToolNames {
		assert.True(t, registry.Has(name), "Tool %s should be registered", name)
	}
}

func TestCharacterTools_CreateCharacter(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create player character
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"name":        "Aragorn",
		"player_id":   "player-001",
		"race":        "Human",
		"class":       "Ranger",
		"level":       5,
		"background":  "Outlander",
		"alignment":   "Neutral Good",
	})

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	character := result["character"].(map[string]interface{})
	assert.Equal(t, "Aragorn", character["name"])
	assert.Equal(t, "Human", character["race"])
	assert.Equal(t, "Ranger", character["class"])
	assert.Equal(t, float64(5), character["level"])
	assert.NotEmpty(t, character["id"])
}

func TestCharacterTools_CreateCharacter_NPC(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create NPC
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"name":        "Goblin Warrior",
		"is_npc":      true,
		"npc_type":    "scripted",
		"race":        "Goblin",
		"class":       "Fighter",
		"level":       1,
	})

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	character := result["character"].(map[string]interface{})
	assert.Equal(t, "Goblin Warrior", character["name"])
	assert.Equal(t, true, character["is_npc"])
	assert.Equal(t, "scripted", character["npc_type"])
}

func TestCharacterTools_CreateCharacter_MissingRequired(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Missing name
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"race":        "Human",
		"class":       "Fighter",
	})

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "character name is required")
}

func TestCharacterTools_CreateCharacter_PlayerRequiresPlayerID(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Player character without player_id
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"name":        "Test Hero",
		"race":        "Human",
		"class":       "Fighter",
	})

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "player ID is required for player characters")
}

func TestCharacterTools_CreateCharacter_WithAbilities(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create character with custom abilities
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
		"name":        "Strong Hero",
		"player_id":   "player-001",
		"race":        "Human",
		"class":       "Fighter",
		"abilities": map[string]interface{}{
			"strength":     18,
			"dexterity":    14,
			"constitution": 16,
			"intelligence": 10,
			"wisdom":       12,
			"charisma":     8,
		},
	})

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	character := result["character"].(map[string]interface{})
	abilities := character["abilities"].(map[string]interface{})
	assert.Equal(t, float64(18), abilities["strength"])
	assert.Equal(t, float64(16), abilities["constitution"])
}

func TestCharacterTools_GetCharacter(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create a character first
	character := models.NewCharacter("campaign-001", "Gandalf", false)
	character.ID = "char-001"
	character.PlayerID = "player-001"
	character.Race = "Human"
	character.Class = "Wizard"
	cStore.Create(ctx, character)

	// Get character
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	c := result["character"].(map[string]interface{})
	assert.Equal(t, "Gandalf", c["name"])
	assert.Equal(t, "char-001", c["id"])
}

func TestCharacterTools_GetCharacter_NotFound(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "non-existent",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCharacterTools_UpdateCharacter(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create a character first
	character := models.NewCharacter("campaign-001", "Legolas", false)
	character.ID = "char-002"
	character.PlayerID = "player-001"
	character.Race = "Elf"
	character.Class = "Ranger"
	character.Level = 3
	cStore.Create(ctx, character)

	// Update character
	newName := "Legolas Greenleaf"
	newLevel := 4
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-002",
		"name":         newName,
		"level":        newLevel,
	})

	req := mcp.ToolRequest{
		ToolName:  "update_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	c := result["character"].(map[string]interface{})
	assert.Equal(t, "Legolas Greenleaf", c["name"])
	assert.Equal(t, float64(4), c["level"])
}

func TestCharacterTools_UpdateCharacter_HP(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create a character first
	character := models.NewCharacter("campaign-001", "Fighter", false)
	character.ID = "char-hp"
	character.PlayerID = "player-001"
	character.HP = models.NewHP(20)
	cStore.Create(ctx, character)

	// Update HP
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "char-hp",
		"hp": map[string]interface{}{
			"current": 15,
			"max":     20,
			"temp":    5,
		},
	})

	req := mcp.ToolRequest{
		ToolName:  "update_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	c := result["character"].(map[string]interface{})
	hp := c["hp"].(map[string]interface{})
	assert.Equal(t, float64(15), hp["current"])
	assert.Equal(t, float64(20), hp["max"])
	assert.Equal(t, float64(5), hp["temp"])
}

func TestCharacterTools_ListCharacters(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create multiple characters
	for i := 0; i < 3; i++ {
		character := models.NewCharacter("campaign-001", "Player "+string(rune('0'+i)), false)
		character.ID = "player-" + string(rune('0'+i))
		character.PlayerID = "player-id-" + string(rune('0'+i))
		cStore.Create(ctx, character)
	}

	// Create an NPC
	npc := models.NewCharacter("campaign-001", "Shopkeeper", true)
	npc.ID = "npc-001"
	npc.NPCType = models.NPCTypeScripted
	cStore.Create(ctx, npc)

	// List all characters in campaign
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "list_characters",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	characters := result["characters"].([]interface{})
	assert.GreaterOrEqual(t, len(characters), 4)
}

func TestCharacterTools_ListCharacters_FilterByNPC(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create player characters
	for i := 0; i < 2; i++ {
		character := models.NewCharacter("campaign-002", "Player "+string(rune('0'+i)), false)
		character.ID = "pc-" + string(rune('0'+i))
		character.PlayerID = "player-id"
		cStore.Create(ctx, character)
	}

	// Create NPCs
	for i := 0; i < 3; i++ {
		npc := models.NewCharacter("campaign-002", "NPC "+string(rune('0'+i)), true)
		npc.ID = "npc-" + string(rune('0'+i))
		npc.NPCType = models.NPCTypeGenerated
		cStore.Create(ctx, npc)
	}

	// List only NPCs
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "campaign-002",
		"is_npc":      true,
	})

	req := mcp.ToolRequest{
		ToolName:  "list_characters",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	characters := result["characters"].([]interface{})
	assert.Equal(t, 3, len(characters))

	// Verify all are NPCs
	for _, c := range characters {
		charMap := c.(map[string]interface{})
		assert.Equal(t, true, charMap["is_npc"])
	}
}

func TestCharacterTools_DeleteCharacter(t *testing.T) {
	characterTools, registry, cStore := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	// Create a character
	character := models.NewCharacter("campaign-001", "To Delete", false)
	character.ID = "delete-id-001"
	character.PlayerID = "player-001"
	cStore.Create(ctx, character)

	// Delete character
	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "delete-id-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "delete_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	assert.True(t, result["success"].(bool))

	// Verify character is deleted
	_, exists := cStore.characters["delete-id-001"]
	assert.False(t, exists)
}

func TestCharacterTools_DeleteCharacter_NotFound(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"character_id": "non-existent",
	})

	req := mcp.ToolRequest{
		ToolName:  "delete_character",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCharacterTools_InvalidJSON(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	req := mcp.ToolRequest{
		ToolName:  "create_character",
		Arguments: []byte(`{invalid json}`),
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid arguments")
}

func TestCharacterTools_UnknownTool(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})
	req := mcp.ToolRequest{
		ToolName:  "unknown_character_tool",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "unknown tool")
}

func TestCharacterTools_ListTools(t *testing.T) {
	characterTools, registry, _ := setupCharacterTools()
	characterTools.Register(registry)

	toolList := registry.List()
	assert.Len(t, toolList, 5)

	// Verify tool definitions
	toolNames := make(map[string]bool)
	for _, tool := range toolList {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Description)
		assert.NotEmpty(t, tool.InputSchema.Type)
	}

	assert.True(t, toolNames["create_character"])
	assert.True(t, toolNames["get_character"])
	assert.True(t, toolNames["update_character"])
	assert.True(t, toolNames["list_characters"])
	assert.True(t, toolNames["delete_character"])
}
