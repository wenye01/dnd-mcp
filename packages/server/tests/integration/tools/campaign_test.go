// Package tools_test contains integration tests for campaign tools
package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dnd-mcp/server/internal/api/tools"
	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCampaignStore for testing
type MockCampaignStore struct {
	campaigns map[string]*models.Campaign
}

func NewMockCampaignStore() *MockCampaignStore {
	return &MockCampaignStore{
		campaigns: make(map[string]*models.Campaign),
	}
}

func (m *MockCampaignStore) Create(ctx context.Context, campaign *models.Campaign) error {
	m.campaigns[campaign.ID] = campaign
	return nil
}

func (m *MockCampaignStore) Get(ctx context.Context, id string) (*models.Campaign, error) {
	c, ok := m.campaigns[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "campaign not found")
	}
	return c, nil
}

func (m *MockCampaignStore) GetByIDAndDM(ctx context.Context, id, dmID string) (*models.Campaign, error) {
	c, ok := m.campaigns[id]
	if !ok || c.DMID != dmID {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "campaign not found")
	}
	return c, nil
}

func (m *MockCampaignStore) List(ctx context.Context, filter *store.CampaignFilter) ([]*models.Campaign, error) {
	var result []*models.Campaign
	for _, c := range m.campaigns {
		result = append(result, c)
	}
	return result, nil
}

func (m *MockCampaignStore) Update(ctx context.Context, campaign *models.Campaign) error {
	m.campaigns[campaign.ID] = campaign
	return nil
}

func (m *MockCampaignStore) Delete(ctx context.Context, id string) error {
	delete(m.campaigns, id)
	return nil
}

func (m *MockCampaignStore) HardDelete(ctx context.Context, id string) error {
	delete(m.campaigns, id)
	return nil
}

func (m *MockCampaignStore) Count(ctx context.Context, filter *store.CampaignFilter) (int64, error) {
	return int64(len(m.campaigns)), nil
}

// MockGameStateStore for testing
type MockGameStateStore struct {
	gameStates map[string]*models.GameState
}

func NewMockGameStateStore() *MockGameStateStore {
	return &MockGameStateStore{
		gameStates: make(map[string]*models.GameState),
	}
}

func (m *MockGameStateStore) Create(ctx context.Context, gs *models.GameState) error {
	m.gameStates[gs.CampaignID] = gs
	return nil
}

func (m *MockGameStateStore) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	gs, ok := m.gameStates[campaignID]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "game state not found")
	}
	return gs, nil
}

func (m *MockGameStateStore) GetByID(ctx context.Context, id string) (*models.GameState, error) {
	for _, gs := range m.gameStates {
		if gs.ID == id {
			return gs, nil
		}
	}
	return nil, service.NewServiceError(service.ErrCodeNotFound, "game state not found")
}

func (m *MockGameStateStore) Update(ctx context.Context, gs *models.GameState) error {
	m.gameStates[gs.CampaignID] = gs
	return nil
}

func (m *MockGameStateStore) Delete(ctx context.Context, campaignID string) error {
	delete(m.gameStates, campaignID)
	return nil
}

// Test fixtures
func setupCampaignTools() (*tools.CampaignTools, *mcp.Registry, *MockCampaignStore) {
	campaignStore := NewMockCampaignStore()
	gameStateStore := NewMockGameStateStore()
	campaignService := service.NewCampaignService(campaignStore, gameStateStore)
	campaignTools := tools.NewCampaignTools(campaignService)
	registry := mcp.NewRegistry()

	return campaignTools, registry, campaignStore
}

func TestCampaignTools_Register(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()

	// Register tools
	campaignTools.Register(registry)

	// Verify all tools are registered
	assert.Equal(t, 5, registry.Count())

	for _, name := range tools.CampaignToolNames {
		assert.True(t, registry.Has(name), "Tool %s should be registered", name)
	}
}

func TestCampaignTools_CreateCampaign(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Create campaign
	args, _ := json.Marshal(map[string]interface{}{
		"name":        "Test Campaign",
		"description": "A test campaign",
		"dm_id":       "dm-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "create_campaign",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	// Parse response
	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	campaign := result["campaign"].(map[string]interface{})
	assert.Equal(t, "Test Campaign", campaign["name"])
	assert.Equal(t, "dm-001", campaign["dm_id"])
	assert.NotEmpty(t, campaign["id"])
}

func TestCampaignTools_CreateCampaign_MissingRequired(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Missing name
	args, _ := json.Marshal(map[string]interface{}{
		"dm_id": "dm-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "create_campaign",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "campaign name is required")
}

func TestCampaignTools_GetCampaign(t *testing.T) {
	campaignTools, registry, cStore := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Get Test", "dm-001", "Description")
	campaign.ID = "test-id-001"
	cStore.Create(ctx, campaign)

	// Get campaign
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "test-id-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_campaign",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	c := result["campaign"].(map[string]interface{})
	assert.Equal(t, "Get Test", c["name"])
}

func TestCampaignTools_GetCampaign_NotFound(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "non-existent",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_campaign",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
}

func TestCampaignTools_ListCampaigns(t *testing.T) {
	campaignTools, registry, cStore := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Create multiple campaigns
	for i := 0; i < 3; i++ {
		campaign := models.NewCampaign("Campaign "+string(rune('0'+i)), "dm-001", "")
		campaign.ID = "id-" + string(rune('0'+i))
		cStore.Create(ctx, campaign)
	}

	// List campaigns
	args, _ := json.Marshal(map[string]interface{}{})

	req := mcp.ToolRequest{
		ToolName:  "list_campaigns",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	campaigns := result["campaigns"].([]interface{})
	assert.GreaterOrEqual(t, len(campaigns), 3)
}

func TestCampaignTools_DeleteCampaign(t *testing.T) {
	campaignTools, registry, cStore := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Delete Test", "dm-001", "")
	campaign.ID = "delete-id-001"
	cStore.Create(ctx, campaign)

	// Delete campaign
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "delete-id-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "delete_campaign",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	assert.True(t, result["success"].(bool))
}

func TestCampaignTools_GetCampaignSummary(t *testing.T) {
	campaignTools, registry, cStore := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Summary Test", "dm-001", "A campaign for summary test")
	campaign.ID = "summary-id-001"
	cStore.Create(ctx, campaign)

	// Get summary
	args, _ := json.Marshal(map[string]interface{}{
		"campaign_id": "summary-id-001",
	})

	req := mcp.ToolRequest{
		ToolName:  "get_campaign_summary",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.False(t, resp.IsError)

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	require.NoError(t, err)

	summary := result["summary"].(map[string]interface{})
	assert.Equal(t, "Summary Test", summary["campaign_name"])
	assert.Equal(t, "summary-id-001", summary["campaign_id"])
}

func TestCampaignTools_InvalidJSON(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	req := mcp.ToolRequest{
		ToolName:  "create_campaign",
		Arguments: []byte(`{invalid json}`),
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "invalid arguments")
}

func TestCampaignTools_UnknownTool(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	ctx := context.Background()

	args, _ := json.Marshal(map[string]interface{}{})
	req := mcp.ToolRequest{
		ToolName:  "unknown_tool",
		Arguments: args,
	}

	resp := registry.Call(ctx, req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "unknown tool")
}

func TestCampaignTools_ListTools(t *testing.T) {
	campaignTools, registry, _ := setupCampaignTools()
	campaignTools.Register(registry)

	toolList := registry.List()
	assert.Len(t, toolList, 5)

	// Verify tool definitions
	toolNames := make(map[string]bool)
	for _, tool := range toolList {
		toolNames[tool.Name] = true
		assert.NotEmpty(t, tool.Description)
		assert.NotEmpty(t, tool.InputSchema.Type)
	}

	assert.True(t, toolNames["create_campaign"])
	assert.True(t, toolNames["get_campaign"])
	assert.True(t, toolNames["list_campaigns"])
	assert.True(t, toolNames["delete_campaign"])
	assert.True(t, toolNames["get_campaign_summary"])
}
