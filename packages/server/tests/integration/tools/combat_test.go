// Package tools contains integration tests for combat tools
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

// Mock stores for combat testing

type MockCombatStore struct {
	combats map[string]*models.Combat
}

func NewMockCombatStore() *MockCombatStore {
	return &MockCombatStore{
		combats: make(map[string]*models.Combat),
	}
}

func (m *MockCombatStore) Create(ctx context.Context, combat *models.Combat) error {
	m.combats[combat.ID] = combat
	return nil
}

func (m *MockCombatStore) Get(ctx context.Context, id string) (*models.Combat, error) {
	c, ok := m.combats[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "combat not found")
	}
	return c, nil
}

func (m *MockCombatStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Combat, error) {
	var result []*models.Combat
	for _, c := range m.combats {
		if c.CampaignID == campaignID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockCombatStore) GetActive(ctx context.Context, campaignID string) (*models.Combat, error) {
	for _, c := range m.combats {
		if c.CampaignID == campaignID && c.IsActive() {
			return c, nil
		}
	}
	return nil, service.NewServiceError(service.ErrCodeNotFound, "no active combat")
}

func (m *MockCombatStore) Update(ctx context.Context, combat *models.Combat) error {
	m.combats[combat.ID] = combat
	return nil
}

func (m *MockCombatStore) Delete(ctx context.Context, id string) error {
	delete(m.combats, id)
	return nil
}

type MockCharacterStoreForCombat struct {
	characters map[string]*models.Character
}

func NewMockCharacterStoreForCombat() *MockCharacterStoreForCombat {
	return &MockCharacterStoreForCombat{
		characters: make(map[string]*models.Character),
	}
}

func (m *MockCharacterStoreForCombat) Create(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForCombat) Get(ctx context.Context, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForCombat) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok || c.CampaignID != campaignID {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStoreForCombat) Update(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStoreForCombat) Delete(ctx context.Context, id string) error {
	delete(m.characters, id)
	return nil
}

func (m *MockCharacterStoreForCombat) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return int64(len(m.characters)), nil
}

func (m *MockCharacterStoreForCombat) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	var result []*models.Character
	for _, c := range m.characters {
		if filter.CampaignID != "" && c.CampaignID != filter.CampaignID {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

type MockCampaignStoreForCombat struct {
	campaigns map[string]*models.Campaign
}

func NewMockCampaignStoreForCombat() *MockCampaignStoreForCombat {
	return &MockCampaignStoreForCombat{
		campaigns: make(map[string]*models.Campaign),
	}
}

func (m *MockCampaignStoreForCombat) Create(ctx context.Context, campaign *models.Campaign) error {
	m.campaigns[campaign.ID] = campaign
	return nil
}

func (m *MockCampaignStoreForCombat) Get(ctx context.Context, id string) (*models.Campaign, error) {
	c, ok := m.campaigns[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "campaign not found")
	}
	return c, nil
}

// Test fixtures
func setupCombatTools() (*tools.CombatTools, *mcp.Registry, *MockCombatStore, *MockCharacterStoreForCombat, *MockCampaignStoreForCombat) {
	combatStore := NewMockCombatStore()
	characterStore := NewMockCharacterStoreForCombat()
	campaignStore := NewMockCampaignStoreForCombat()

	roller := dice.NewRollerWithSource(dice.NewSeededRandomSource(42))
	diceService := service.NewDiceServiceWithRoller(characterStore, roller)
	combatService := service.NewCombatServiceWithRoller(combatStore, characterStore, campaignStore, diceService, roller)
	combatTools := tools.NewCombatTools(combatService)
	registry := mcp.NewRegistry()

	return combatTools, registry, combatStore, characterStore, campaignStore
}

func createTestCampaign(store *MockCampaignStoreForCombat) *models.Campaign {
	campaign := models.NewCampaign("Test Campaign", "dm-001", "Test campaign description for combat tests")
	campaign.ID = "campaign-001"
	store.campaigns[campaign.ID] = campaign
	return campaign
}

func createTestCharacterForCombat(store *MockCharacterStoreForCombat, id, name, campaignID string) *models.Character {
	character := models.NewCharacter(campaignID, name, false)
	character.ID = id
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
	character.AC = 16
	store.characters[character.ID] = character
	return character
}

func TestCombatTools_Register(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()

	// Register tools
	combatTools.Register(registry)

	// Verify all tools are registered
	assert.Equal(t, 6, registry.Count())

	for _, name := range tools.CombatToolNames {
		assert.True(t, registry.Has(name), "Tool %s should be registered", name)
	}
}

func TestCombatTools_StartCombat(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})

	req := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Should not error: %s", resp.Content[0].Text)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	combat := result["combat"].(map[string]interface{})
	assert.Equal(t, "campaign-001", combat["campaign_id"])
	assert.Equal(t, float64(1), combat["round"])
	assert.True(t, combat["active"].(bool))

	participants := combat["participants"].([]interface{})
	assert.Len(t, participants, 2)

	assert.Contains(t, result["message"], "Combat started")
}

func TestCombatTools_StartCombat_MissingCampaign(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "non-existent",
		"participant_ids": []string{"char-001"},
	})

	req := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "campaign not found")
}

func TestCombatTools_StartCombat_NoParticipants(t *testing.T) {
	combatTools, registry, _, _, campStore := setupCombatTools()
	combatTools.Register(registry)

	createTestCampaign(campStore)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{},
	})

	req := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "at least one participant")
}

func TestCombatTools_GetCombatState(t *testing.T) {
	combatTools, registry, combatStore, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get combat state
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id": combatID,
	})

	req := mcp.ToolRequest{
		ToolName:  "get_combat_state",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	combat := result["combat"].(map[string]interface{})
	assert.Equal(t, combatID, combat["id"])
	assert.Equal(t, "campaign-001", combat["campaign_id"])
	assert.True(t, combat["active"].(bool))

	// Verify combat was retrieved from store
	_, exists := combatStore.combats[combatID]
	assert.True(t, exists)
}

func TestCombatTools_GetCombatState_NotFound(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"combat_id": "non-existent",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_combat_state",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "combat not found")
}

func TestCombatTools_Attack(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current attacker (first in initiative order)
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	attackerID := firstParticipant["character_id"].(string)

	// Find target (the other participant)
	var targetID string
	for _, p := range participants {
		pm := p.(map[string]interface{})
		if pm["character_id"].(string) != attackerID {
			targetID = pm["character_id"].(string)
			break
		}
	}

	// Attack
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"attacker_id": attackerID,
		"target_id":   targetID,
	})

	req := mcp.ToolRequest{
		ToolName:  "attack",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Should not error: %s", resp.Content[0].Text)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	attackResult := result["result"].(map[string]interface{})
	assert.Contains(t, attackResult, "hit")
	assert.Contains(t, attackResult, "attack_roll")
	assert.Contains(t, attackResult, "target_ac")
	assert.Contains(t, result, "message")
}

func TestCombatTools_Attack_NotTheirTurn(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current attacker (first in initiative order)
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	attackerID := firstParticipant["character_id"].(string)

	// Find target (the other participant) - try to attack with them instead
	var wrongAttackerID string
	for _, p := range participants {
		pm := p.(map[string]interface{})
		if pm["character_id"].(string) != attackerID {
			wrongAttackerID = pm["character_id"].(string)
			break
		}
	}

	// Attack with wrong attacker
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"attacker_id": wrongAttackerID,
		"target_id":   attackerID,
	})

	req := mcp.ToolRequest{
		ToolName:  "attack",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "not attacker's turn")
}

func TestCombatTools_Attack_WithAdvantage(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current attacker
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	attackerID := firstParticipant["character_id"].(string)

	var targetID string
	for _, p := range participants {
		pm := p.(map[string]interface{})
		if pm["character_id"].(string) != attackerID {
			targetID = pm["character_id"].(string)
			break
		}
	}

	// Attack with advantage
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"attacker_id": attackerID,
		"target_id":   targetID,
		"advantage":   true,
	})

	req := mcp.ToolRequest{
		ToolName:  "attack",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)
}

func TestCombatTools_CastSpell(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current caster
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	casterID := firstParticipant["character_id"].(string)

	var targetID string
	for _, p := range participants {
		pm := p.(map[string]interface{})
		if pm["character_id"].(string) != casterID {
			targetID = pm["character_id"].(string)
			break
		}
	}

	// Cast spell
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"caster_id":   casterID,
		"spell_name":  "Fireball",
		"target_ids":  []string{targetID},
		"damage":      "8d6",
		"damage_type": "fire",
	})

	req := mcp.ToolRequest{
		ToolName:  "cast_spell",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError, "Should not error: %s", resp.Content[0].Text)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	spellResult := result["result"].(map[string]interface{})
	assert.Equal(t, "Fireball", spellResult["spell_name"])
	assert.Equal(t, "fire", spellResult["damage_type"])
	assert.Contains(t, result, "message")
}

func TestCombatTools_CastSpell_Healing(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current caster
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	casterID := firstParticipant["character_id"].(string)

	// Cast healing spell
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"caster_id":   casterID,
		"spell_name":  "Cure Wounds",
		"target_ids":  []string{casterID},
		"damage":      "1d8",
		"is_healing":  true,
	})

	req := mcp.ToolRequest{
		ToolName:  "cast_spell",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	spellResult := result["result"].(map[string]interface{})
	assert.Equal(t, "Cure Wounds", spellResult["spell_name"])
	assert.True(t, spellResult["is_healing"].(bool))
	assert.Contains(t, result["message"], "healed")
}

func TestCombatTools_CastSpell_NotTheirTurn(t *testing.T) {
	combatTools, registry, _, charStore, campStore := setupCombatTools()
	combatTools.Register(registry)

	// Create test data
	createTestCampaign(campStore)
	char1 := createTestCharacterForCombat(charStore, "char-001", "Hero", "campaign-001")
	char2 := createTestCharacterForCombat(charStore, "char-002", "Villain", "campaign-001")

	// Start combat first
	ctx := context.Background()
	startArgs, _ := json.Marshal(map[string]interface{}{
		"campaign_id":     "campaign-001",
		"participant_ids": []string{char1.ID, char2.ID},
	})
	startReq := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: startArgs,
	}
	startResp := registry.Call(ctx, startReq)
	require.False(t, startResp.IsError)

	var startResult map[string]interface{}
	json.Unmarshal([]byte(startResp.Content[0].Text), &startResult)
	combatData := startResult["combat"].(map[string]interface{})
	combatID := combatData["id"].(string)

	// Get current caster
	participants := combatData["participants"].([]interface{})
	firstParticipant := participants[0].(map[string]interface{})
	casterID := firstParticipant["character_id"].(string)

	// Find wrong caster
	var wrongCasterID string
	for _, p := range participants {
		pm := p.(map[string]interface{})
		if pm["character_id"].(string) != casterID {
			wrongCasterID = pm["character_id"].(string)
			break
		}
	}

	// Try to cast with wrong caster
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   combatID,
		"caster_id":   wrongCasterID,
		"spell_name":  "Fireball",
		"target_ids":  []string{casterID},
	})

	req := mcp.ToolRequest{
		ToolName:  "cast_spell",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "not caster's turn")
}

func TestCombatTools_InvalidJSON(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	req := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: []byte(`{invalid json}`),
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid arguments")
}

func TestCombatTools_UnknownTool(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})
	req := mcp.ToolRequest{
		ToolName:  "unknown_combat_tool",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "unknown tool")
}

func TestCombatTools_ListTools(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	toolList := registry.List()
	assert.Len(t, toolList, 6)

	// Verify tool definitions
	toolNames := make(map[string]bool)
	for _, tool := range toolList {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Description)
		assert.NotEmpty(t, tool.InputSchema.Type)
	}

	assert.True(t, toolNames["start_combat"])
	assert.True(t, toolNames["get_combat_state"])
	assert.True(t, toolNames["attack"])
	assert.True(t, toolNames["cast_spell"])
	assert.True(t, toolNames["end_turn"])
	assert.True(t, toolNames["end_combat"])
}

func TestCombatTools_StartCombat_MissingRequired(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	// Missing campaign_id
	args, _ := json.Marshal(map[string]interface{}{
		"participant_ids": []string{"char-001"},
	})

	req := mcp.ToolRequest{
		ToolName:  "start_combat",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCombatTools_GetCombatState_MissingRequired(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	// Missing combat_id
	args, _ := json.Marshal(map[string]interface{}{})

	req := mcp.ToolRequest{
		ToolName:  "get_combat_state",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCombatTools_Attack_MissingRequired(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	// Missing target_id
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   "combat-001",
		"attacker_id": "char-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "attack",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCombatTools_CastSpell_MissingRequired(t *testing.T) {
	combatTools, registry, _, _, _ := setupCombatTools()
	combatTools.Register(registry)

	ctx := context.Background()

	// Missing target_ids
	args, _ := json.Marshal(map[string]interface{}{
		"combat_id":   "combat-001",
		"caster_id":   "char-001",
		"spell_name":  "Fireball",
	})

	req := mcp.ToolRequest{
		ToolName:  "cast_spell",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}
