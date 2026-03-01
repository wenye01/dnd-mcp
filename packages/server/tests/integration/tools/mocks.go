// Package tools contains common mock implementations for testing
package tools

import (
	"context"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/service"
	"github.com/dnd-mcp/server/internal/store"
)

// MockMapStore for testing
type MockMapStore struct {
	maps map[string]*models.Map
}

func NewMockMapStore() *MockMapStore {
	return &MockMapStore{
		maps: make(map[string]*models.Map),
	}
}

func (m *MockMapStore) Create(ctx context.Context, gameMap *models.Map) error {
	m.maps[gameMap.ID] = gameMap
	return nil
}

func (m *MockMapStore) Get(ctx context.Context, id string) (*models.Map, error) {
	gameMap, ok := m.maps[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "map not found")
	}
	return gameMap, nil
}

func (m *MockMapStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error) {
	var result []*models.Map
	for _, gameMap := range m.maps {
		if gameMap.CampaignID == campaignID {
			result = append(result, gameMap)
		}
	}
	return result, nil
}

func (m *MockMapStore) Update(ctx context.Context, gameMap *models.Map) error {
	m.maps[gameMap.ID] = gameMap
	return nil
}

func (m *MockMapStore) Delete(ctx context.Context, id string) error {
	delete(m.maps, id)
	return nil
}

func (m *MockMapStore) GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error) {
	// Try to find an existing world map
	for _, gameMap := range m.maps {
		if gameMap.CampaignID == campaignID && gameMap.IsWorldMap() {
			return gameMap, nil
		}
	}
	// Return not found - service will create one
	return nil, service.NewServiceError(service.ErrCodeNotFound, "world map not found")
}

func (m *MockMapStore) GetBattleMap(ctx context.Context, id string) (*models.Map, error) {
	gameMap, ok := m.maps[id]
	if !ok || !gameMap.IsBattleMap() {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "battle map not found")
	}
	return gameMap, nil
}

func (m *MockMapStore) GetByParent(ctx context.Context, parentID string) ([]*models.Map, error) {
	var result []*models.Map
	for _, gameMap := range m.maps {
		if gameMap.ParentID == parentID {
			result = append(result, gameMap)
		}
	}
	return result, nil
}

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

// MockCharacterStore for testing
type MockCharacterStore struct {
	characters map[string]*models.Character
}

func NewMockCharacterStore() *MockCharacterStore {
	return &MockCharacterStore{
		characters: make(map[string]*models.Character),
	}
}

func (m *MockCharacterStore) Create(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStore) Get(ctx context.Context, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStore) GetByCampaignAndID(ctx context.Context, campaignID, id string) (*models.Character, error) {
	c, ok := m.characters[id]
	if !ok || c.CampaignID != campaignID {
		return nil, service.NewServiceError(service.ErrCodeNotFound, "character not found")
	}
	return c, nil
}

func (m *MockCharacterStore) List(ctx context.Context, filter *store.CharacterFilter) ([]*models.Character, error) {
	var result []*models.Character
	for _, c := range m.characters {
		if filter != nil && filter.CampaignID != "" && c.CampaignID != filter.CampaignID {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *MockCharacterStore) Update(ctx context.Context, character *models.Character) error {
	m.characters[character.ID] = character
	return nil
}

func (m *MockCharacterStore) Delete(ctx context.Context, id string) error {
	delete(m.characters, id)
	return nil
}

func (m *MockCharacterStore) Count(ctx context.Context, filter *store.CharacterFilter) (int64, error) {
	return int64(len(m.characters)), nil
}
