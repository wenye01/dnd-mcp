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

// MockCampaignStore is a mock implementation of CampaignStore
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
	campaign, ok := m.campaigns[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return campaign, nil
}

func (m *MockCampaignStore) GetByIDAndDM(ctx context.Context, id, dmID string) (*models.Campaign, error) {
	campaign, ok := m.campaigns[id]
	if !ok || campaign.DMID != dmID {
		return nil, errors.New("not found")
	}
	return campaign, nil
}

func (m *MockCampaignStore) List(ctx context.Context, filter *store.CampaignFilter) ([]*models.Campaign, error) {
	var result []*models.Campaign
	for _, c := range m.campaigns {
		if filter != nil {
			if filter.Status != "" && c.Status != filter.Status {
				continue
			}
			if filter.DMID != "" && c.DMID != filter.DMID {
				continue
			}
			if !filter.IncludeDeleted && c.DeletedAt != nil {
				continue
			}
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *MockCampaignStore) Update(ctx context.Context, campaign *models.Campaign) error {
	if _, ok := m.campaigns[campaign.ID]; !ok {
		return errors.New("not found")
	}
	m.campaigns[campaign.ID] = campaign
	return nil
}

func (m *MockCampaignStore) Delete(ctx context.Context, id string) error {
	if campaign, ok := m.campaigns[id]; ok {
		now := models.CampaignStatusArchived
		campaign.Status = now
		return nil
	}
	return errors.New("not found")
}

func (m *MockCampaignStore) HardDelete(ctx context.Context, id string) error {
	delete(m.campaigns, id)
	return nil
}

func (m *MockCampaignStore) Count(ctx context.Context, filter *store.CampaignFilter) (int64, error) {
	campaigns, _ := m.List(ctx, filter)
	return int64(len(campaigns)), nil
}

// MockGameStateStore is a mock implementation of GameStateStore
type MockGameStateStore struct {
	gameStates map[string]*models.GameState
}

func NewMockGameStateStore() *MockGameStateStore {
	return &MockGameStateStore{
		gameStates: make(map[string]*models.GameState),
	}
}

func (m *MockGameStateStore) Create(ctx context.Context, gameState *models.GameState) error {
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *MockGameStateStore) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	gs, ok := m.gameStates[campaignID]
	if !ok {
		return nil, errors.New("not found")
	}
	return gs, nil
}

func (m *MockGameStateStore) GetByID(ctx context.Context, id string) (*models.GameState, error) {
	for _, gs := range m.gameStates {
		if gs.ID == id {
			return gs, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockGameStateStore) Update(ctx context.Context, gameState *models.GameState) error {
	if _, ok := m.gameStates[gameState.CampaignID]; !ok {
		return errors.New("not found")
	}
	m.gameStates[gameState.CampaignID] = gameState
	return nil
}

func (m *MockGameStateStore) Delete(ctx context.Context, campaignID string) error {
	delete(m.gameStates, campaignID)
	return nil
}

// Test fixtures
func setupCampaignService() (*service.CampaignService, *MockCampaignStore, *MockGameStateStore) {
	campaignStore := NewMockCampaignStore()
	gameStateStore := NewMockGameStateStore()
	svc := service.NewCampaignService(campaignStore, gameStateStore)
	return svc, campaignStore, gameStateStore
}

func TestCampaignService_CreateCampaign(t *testing.T) {
	svc, _, gsStore := setupCampaignService()
	ctx := context.Background()

	req := &service.CreateCampaignRequest{
		Name:        "Test Campaign",
		Description: "A test campaign",
		DMID:        "dm-001",
	}

	campaign, err := svc.CreateCampaign(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, campaign)

	assert.NotEmpty(t, campaign.ID)
	assert.Equal(t, "Test Campaign", campaign.Name)
	assert.Equal(t, "A test campaign", campaign.Description)
	assert.Equal(t, "dm-001", campaign.DMID)
	assert.Equal(t, models.CampaignStatusActive, campaign.Status)

	// Verify game state was created
	gs, err := gsStore.Get(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, campaign.ID, gs.CampaignID)
}

func TestCampaignService_CreateCampaign_WithSettings(t *testing.T) {
	svc, _, _ := setupCampaignService()
	ctx := context.Background()

	req := &service.CreateCampaignRequest{
		Name:        "Custom Campaign",
		Description: "",
		DMID:        "dm-002",
		Settings: &service.CampaignSettingsInput{
			MaxPlayers:    6,
			StartLevel:    3,
			Ruleset:       "dnd5e",
			ContextWindow: 30,
			HouseRules:    map[string]interface{}{"critical": "double"},
		},
	}

	campaign, err := svc.CreateCampaign(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, 6, campaign.Settings.MaxPlayers)
	assert.Equal(t, 3, campaign.Settings.StartLevel)
	assert.Equal(t, "dnd5e", campaign.Settings.Ruleset)
	assert.Equal(t, 30, campaign.Settings.ContextWindow)
}

func TestCampaignService_CreateCampaign_ValidationErrors(t *testing.T) {
	svc, _, _ := setupCampaignService()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *service.CreateCampaignRequest
		expectedErr string
	}{
		{
			name: "empty name",
			req: &service.CreateCampaignRequest{
				Name: "",
				DMID: "dm-001",
			},
			expectedErr: "campaign name is required",
		},
		{
			name: "empty DM ID",
			req: &service.CreateCampaignRequest{
				Name: "Test",
				DMID: "",
			},
			expectedErr: "DM ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateCampaign(ctx, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestCampaignService_GetCampaign(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Get Test", "dm-001", "Description")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	// Get the campaign
	retrieved, err := svc.GetCampaign(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, campaign.ID, retrieved.ID)
	assert.Equal(t, campaign.Name, retrieved.Name)
}

func TestCampaignService_GetCampaign_NotFound(t *testing.T) {
	svc, _, _ := setupCampaignService()
	ctx := context.Background()

	_, err := svc.GetCampaign(ctx, uuid.New().String())
	require.Error(t, err)
}

func TestCampaignService_GetCampaign_EmptyID(t *testing.T) {
	svc, _, _ := setupCampaignService()
	ctx := context.Background()

	_, err := svc.GetCampaign(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "campaign ID is required")
}

func TestCampaignService_GetCampaignByDM(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("DM Test", "dm-001", "")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	// Get by DM
	retrieved, err := svc.GetCampaignByDM(ctx, campaign.ID, "dm-001")
	require.NoError(t, err)
	assert.Equal(t, campaign.ID, retrieved.ID)

	// Wrong DM
	_, err = svc.GetCampaignByDM(ctx, campaign.ID, "wrong-dm")
	require.Error(t, err)
}

func TestCampaignService_ListCampaigns(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create multiple campaigns
	for i := 0; i < 5; i++ {
		campaign := models.NewCampaign(
			"Campaign "+string(rune('0'+i)),
			"dm-"+string(rune('0'+i)),
			"",
		)
		campaign.ID = uuid.New().String()
		cStore.Create(ctx, campaign)
	}

	// List all
	campaigns, err := svc.ListCampaigns(ctx, &service.ListCampaignsRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(campaigns), 5)
}

func TestCampaignService_ListCampaigns_WithFilter(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create campaigns with different DMs
	campaign1 := models.NewCampaign("Campaign 1", "dm-filter", "")
	campaign1.ID = uuid.New().String()
	cStore.Create(ctx, campaign1)

	campaign2 := models.NewCampaign("Campaign 2", "dm-other", "")
	campaign2.ID = uuid.New().String()
	cStore.Create(ctx, campaign2)

	// Filter by DM
	campaigns, err := svc.ListCampaigns(ctx, &service.ListCampaignsRequest{
		DMID: "dm-filter",
	})
	require.NoError(t, err)
	assert.Len(t, campaigns, 1)
	assert.Equal(t, "dm-filter", campaigns[0].DMID)
}

func TestCampaignService_UpdateCampaign(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Original Name", "dm-001", "Original description")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	// Update
	newName := "Updated Name"
	newDesc := "Updated description"
	updated, err := svc.UpdateCampaign(ctx, campaign.ID, &service.UpdateCampaignRequest{
		Name:        &newName,
		Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
}

func TestCampaignService_UpdateCampaign_Status(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create an active campaign
	campaign := models.NewCampaign("Status Test", "dm-001", "")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	// Pause campaign
	paused := models.CampaignStatusPaused
	updated, err := svc.UpdateCampaign(ctx, campaign.ID, &service.UpdateCampaignRequest{
		Status: &paused,
	})
	require.NoError(t, err)
	assert.Equal(t, models.CampaignStatusPaused, updated.Status)

	// Resume campaign
	active := models.CampaignStatusActive
	updated, err = svc.UpdateCampaign(ctx, campaign.ID, &service.UpdateCampaignRequest{
		Status: &active,
	})
	require.NoError(t, err)
	assert.Equal(t, models.CampaignStatusActive, updated.Status)
}

func TestCampaignService_UpdateCampaign_InvalidStatusTransition(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create an archived campaign
	campaign := models.NewCampaign("Archived Test", "dm-001", "")
	campaign.ID = uuid.New().String()
	campaign.Status = models.CampaignStatusArchived
	cStore.Create(ctx, campaign)

	// Try to activate (should fail)
	active := models.CampaignStatusActive
	_, err := svc.UpdateCampaign(ctx, campaign.ID, &service.UpdateCampaignRequest{
		Status: &active,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transition")
}

func TestCampaignService_UpdateCampaign_EmptyName(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	campaign := models.NewCampaign("Test", "dm-001", "")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	emptyName := ""
	_, err := svc.UpdateCampaign(ctx, campaign.ID, &service.UpdateCampaignRequest{
		Name: &emptyName,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "campaign name cannot be empty")
}

func TestCampaignService_DeleteCampaign(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	campaign := models.NewCampaign("Delete Test", "dm-001", "")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	err := svc.DeleteCampaign(ctx, campaign.ID)
	require.NoError(t, err)
}

func TestCampaignService_DeleteCampaign_NotFound(t *testing.T) {
	svc, _, _ := setupCampaignService()
	ctx := context.Background()

	err := svc.DeleteCampaign(ctx, uuid.New().String())
	require.Error(t, err)
}

func TestCampaignService_GetCampaignSummary(t *testing.T) {
	svc, cStore, gsStore := setupCampaignService()
	ctx := context.Background()

	campaign := models.NewCampaign("Summary Test", "dm-001", "Description")
	campaign.ID = uuid.New().String()
	cStore.Create(ctx, campaign)

	// Create game state
	gs := models.NewGameState(campaign.ID)
	gsStore.Create(ctx, gs)

	summary, err := svc.GetCampaignSummary(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, campaign.ID, summary.CampaignID)
	assert.Equal(t, "Summary Test", summary.CampaignName)
	assert.Equal(t, models.CampaignStatusActive, summary.Status)
	assert.NotNil(t, summary.GameTime)
}

func TestCampaignService_GetGameState(t *testing.T) {
	svc, _, gsStore := setupCampaignService()
	ctx := context.Background()

	campaignID := uuid.New().String()
	gs := models.NewGameState(campaignID)
	gsStore.Create(ctx, gs)

	retrieved, err := svc.GetGameState(ctx, campaignID)
	require.NoError(t, err)
	assert.Equal(t, campaignID, retrieved.CampaignID)
}

func TestCampaignService_CountCampaigns(t *testing.T) {
	svc, cStore, _ := setupCampaignService()
	ctx := context.Background()

	// Create campaigns
	for i := 0; i < 3; i++ {
		campaign := models.NewCampaign("Count "+string(rune('0'+i)), "dm-count", "")
		campaign.ID = uuid.New().String()
		cStore.Create(ctx, campaign)
	}

	count, err := svc.CountCampaigns(ctx, &service.ListCampaignsRequest{
		DMID: "dm-count",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

func TestServiceError(t *testing.T) {
	err := service.NewServiceError(service.ErrCodeInvalidInput, "test error")
	assert.Equal(t, service.ErrCodeInvalidInput, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, "test error", err.Error())

	assert.True(t, service.IsServiceError(err))
	assert.False(t, service.IsServiceError(errors.New("other error")))

	se := service.GetServiceError(err)
	require.NotNil(t, se)
	assert.Equal(t, service.ErrCodeInvalidInput, se.Code)

	se = service.GetServiceError(errors.New("other error"))
	assert.Nil(t, se)
}
