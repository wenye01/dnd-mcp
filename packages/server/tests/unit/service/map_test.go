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

// MockMapStore is a mock implementation of MapStore
type MockMapStore struct {
	mock.Mock
}

func (m *MockMapStore) Create(ctx context.Context, gameMap *models.Map) error {
	args := m.Called(ctx, gameMap)
	return args.Error(0)
}

func (m *MockMapStore) Get(ctx context.Context, id string) (*models.Map, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Map), args.Error(1)
}

func (m *MockMapStore) GetByCampaign(ctx context.Context, campaignID string) ([]*models.Map, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Map), args.Error(1)
}

func (m *MockMapStore) Update(ctx context.Context, gameMap *models.Map) error {
	args := m.Called(ctx, gameMap)
	return args.Error(0)
}

func (m *MockMapStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMapStore) GetWorldMap(ctx context.Context, campaignID string) (*models.Map, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Map), args.Error(1)
}

func (m *MockMapStore) GetBattleMap(ctx context.Context, id string) (*models.Map, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Map), args.Error(1)
}

func (m *MockMapStore) GetByParent(ctx context.Context, parentID string) ([]*models.Map, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Map), args.Error(1)
}

// MockCampaignStoreForMap is a mock implementation of CampaignStoreForMap
type MockCampaignStoreForMap struct {
	mock.Mock
}

func (m *MockCampaignStoreForMap) Get(ctx context.Context, id string) (*models.Campaign, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Campaign), args.Error(1)
}

// MockGameStateStoreForMap is a mock implementation of GameStateStoreForMap
type MockGameStateStoreForMap struct {
	mock.Mock
}

func (m *MockGameStateStoreForMap) Get(ctx context.Context, campaignID string) (*models.GameState, error) {
	args := m.Called(ctx, campaignID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameState), args.Error(1)
}

func (m *MockGameStateStoreForMap) Update(ctx context.Context, gameState *models.GameState) error {
	args := m.Called(ctx, gameState)
	return args.Error(0)
}

func TestMapService_GetWorldMap(t *testing.T) {
	tests := []struct {
		name          string
		campaignID    string
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap)
		expectError   bool
		errorContains string
		validateMap   func(t *testing.T, gameMap *models.Map)
	}{
		{
			name:       "successful get existing world map",
			campaignID: "campaign-123",
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				c.On("Get", mock.Anything, "campaign-123").Return(&models.Campaign{
					ID:   "campaign-123",
					Name: "Test Campaign",
				}, nil)

				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
			},
			expectError: false,
			validateMap: func(t *testing.T, gameMap *models.Map) {
				assert.NotNil(t, gameMap)
				assert.True(t, gameMap.IsWorldMap())
				assert.Equal(t, 50, gameMap.Grid.Width)
				assert.Equal(t, 50, gameMap.Grid.Height)
			},
		},
		{
			name:       "create default world map if not exists",
			campaignID: "campaign-123",
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				c.On("Get", mock.Anything, "campaign-123").Return(&models.Campaign{
					ID:   "campaign-123",
					Name: "Test Campaign",
				}, nil)

				// First call returns not found
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(nil, errors.New("not found"))
				// Then create new map
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateMap: func(t *testing.T, gameMap *models.Map) {
				assert.NotNil(t, gameMap)
				assert.True(t, gameMap.IsWorldMap())
				assert.Equal(t, 50, gameMap.Grid.Width)
				assert.Equal(t, 50, gameMap.Grid.Height)
			},
		},
		{
			name:       "campaign not found",
			campaignID: "campaign-123",
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				c.On("Get", mock.Anything, "campaign-123").Return(nil, errors.New("campaign not found"))
			},
			expectError:   true,
			errorContains: "failed to get campaign",
		},
		{
			name:       "empty campaign ID",
			campaignID: "",
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			gameMap, err := svc.GetWorldMap(context.Background(), tt.campaignID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateMap != nil {
					tt.validateMap(t, gameMap)
				}
			}

			mapStore.AssertExpectations(t)
			campaignStore.AssertExpectations(t)
		})
	}
}

func TestMapService_CreateBattleMap(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.CreateBattleMapRequest
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap)
		expectError   bool
		errorContains string
		validateMap   func(t *testing.T, gameMap *models.Map)
	}{
		{
			name: "successful create battle map",
			req: &service.CreateBattleMapRequest{
				CampaignID: "campaign-123",
				Name:       "Dungeon Room 1",
				Width:      20,
				Height:     20,
				CellSize:   5,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				c.On("Get", mock.Anything, "campaign-123").Return(&models.Campaign{
					ID:   "campaign-123",
					Name: "Test Campaign",
				}, nil)
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateMap: func(t *testing.T, gameMap *models.Map) {
				assert.NotNil(t, gameMap)
				assert.True(t, gameMap.IsBattleMap())
				assert.Equal(t, "Dungeon Room 1", gameMap.Name)
				assert.Equal(t, 20, gameMap.Grid.Width)
				assert.Equal(t, 20, gameMap.Grid.Height)
				assert.Equal(t, 5, gameMap.Grid.CellSize)
			},
		},
		{
			name: "empty campaign ID",
			req: &service.CreateBattleMapRequest{
				Name:     "Dungeon",
				Width:    20,
				Height:   20,
				CellSize: 5,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
		{
			name: "empty name",
			req: &service.CreateBattleMapRequest{
				CampaignID: "campaign-123",
				Width:      20,
				Height:     20,
				CellSize:   5,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "name is required",
		},
		{
			name: "invalid width (too large)",
			req: &service.CreateBattleMapRequest{
				CampaignID: "campaign-123",
				Name:       "Dungeon",
				Width:      500, // Exceeds MaxMapWidth
				Height:     20,
				CellSize:   5,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "width must be between",
		},
		{
			name: "invalid height (negative)",
			req: &service.CreateBattleMapRequest{
				CampaignID: "campaign-123",
				Name:       "Dungeon",
				Width:      20,
				Height:     -5,
				CellSize:   5,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "height must be between",
		},
		{
			name: "invalid cell size (zero)",
			req: &service.CreateBattleMapRequest{
				CampaignID: "campaign-123",
				Name:       "Dungeon",
				Width:      20,
				Height:     20,
				CellSize:   0,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "cell size must be positive",
		},
		{
			name: "campaign not found",
			req: &service.CreateBattleMapRequest{
				CampaignID: "nonexistent",
				Name:       "Dungeon",
				Width:      20,
				Height:     20,
				CellSize:   5,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				c.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("campaign not found"))
			},
			expectError:   true,
			errorContains: "failed to get campaign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			gameMap, err := svc.CreateBattleMap(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateMap != nil {
					tt.validateMap(t, gameMap)
				}
			}

			mapStore.AssertExpectations(t)
			campaignStore.AssertExpectations(t)
		})
	}
}

func TestMapService_CalculateTravelTime(t *testing.T) {
	tests := []struct {
		name            string
		fromX, fromY    int
		toX, toY        int
		travelMode      string
		expectedResult  *service.TravelResult
	}{
		{
			name:      "no movement (same position)",
			fromX:     0,
			fromY:     0,
			toX:       0,
			toY:       0,
			travelMode: "normal",
			expectedResult: &service.TravelResult{
				Distance:    0,
				Hours:       0,
				Days:        0,
				Pace:        "normal",
				Description: "Normal pace: 0 miles at 3 mph = 0 hours",
			},
		},
		{
			name:      "normal pace 10 miles",
			fromX:     0,
			fromY:     0,
			toX:       10,
			toY:       0,
			travelMode: "normal",
			expectedResult: &service.TravelResult{
				Distance:    10,
				Hours:       3, // 10 / 3 = 3.33, truncated to 3
				Days:        10.0 / 24.0,
				Pace:        "normal",
				Description: "Normal pace: 10 miles at 3 mph = 3 hours",
			},
		},
		{
			name:      "fast pace 30 miles",
			fromX:     0,
			fromY:     0,
			toX:       30,
			toY:       0,
			travelMode: "fast",
			expectedResult: &service.TravelResult{
				Distance:    30,
				Hours:       7, // 30 / 4 = 7.5, truncated to 7
				Days:        30.0 / 30.0,
				Pace:        "fast",
				Description: "Fast pace: 30 miles at 4 mph = 7 hours",
			},
		},
		{
			name:      "slow pace 12 miles",
			fromX:     0,
			fromY:     0,
			toX:       12,
			toY:       0,
			travelMode: "slow",
			expectedResult: &service.TravelResult{
				Distance:    12,
				Hours:       6, // 12 / 2 = 6
				Days:        12.0 / 12.0,
				Pace:        "slow",
				Description: "Slow pace: 12 miles at 2 mph = 6 hours",
			},
		},
		{
			name:      "diagonal movement (5,5) from (0,0)",
			fromX:     0,
			fromY:     0,
			toX:       5,
			toY:       5,
			travelMode: "normal",
			expectedResult: &service.TravelResult{
				Distance:    10, // Manhattan distance: 5 + 5
				Hours:       3, // 10 / 3 = 3.33, truncated to 3
				Days:        10.0 / 24.0,
				Pace:        "normal",
				Description: "Normal pace: 10 miles at 3 mph = 3 hours",
			},
		},
		{
			name:      "very short distance (1 mile)",
			fromX:     0,
			fromY:     0,
			toX:       1,
			toY:       0,
			travelMode: "normal",
			expectedResult: &service.TravelResult{
				Distance:    1,
				Hours:       1, // Minimum 1 hour if distance > 0
				Days:        1.0 / 24.0,
				Pace:        "normal",
				Description: "Normal pace: 1 miles at 3 mph = 1 hours",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			result := svc.CalculateTravelTime(tt.fromX, tt.fromY, tt.toX, tt.toY, tt.travelMode)

			assert.Equal(t, tt.expectedResult.Distance, result.Distance)
			assert.Equal(t, tt.expectedResult.Hours, result.Hours)
			assert.InDelta(t, tt.expectedResult.Days, result.Days, 0.01)
			assert.Equal(t, tt.expectedResult.Pace, result.Pace)
		})
	}
}

func TestMapService_MoveTo(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.MoveToRequest
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap, *MockGameStateStoreForMap)
		expectError   bool
		errorContains string
		validateState func(t *testing.T, result *service.MoveToResult)
	}{
		{
			name: "successful move",
			req: &service.MoveToRequest{
				CampaignID: "campaign-123",
				X:          5,
				Y:          10,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)

				gameState := models.NewGameState("campaign-123")
				s.On("Get", mock.Anything, "campaign-123").Return(gameState, nil)
				s.On("Update", mock.Anything, mock.AnythingOfType("*models.GameState")).Return(nil)
			},
			expectError: false,
			validateState: func(t *testing.T, result *service.MoveToResult) {
				state := result.GameState
				assert.NotNil(t, state.PartyPosition)
				assert.Equal(t, 5, state.PartyPosition.X)
				assert.Equal(t, 10, state.PartyPosition.Y)
				// Game time should have advanced (5 + 10 = 15 miles / 3 mph = 5 hours)
				// Default time is 8:00 AM, so should be 1:00 PM
				assert.Equal(t, 13, state.GameTime.Hour)
				// Check travel result
				assert.Equal(t, 15, result.TravelResult.Distance)
				assert.Equal(t, 5, result.TravelResult.Hours)
			},
		},
		{
			name: "empty campaign ID",
			req: &service.MoveToRequest{
				X: 5,
				Y: 10,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
		{
			name: "negative position",
			req: &service.MoveToRequest{
				CampaignID: "campaign-123",
				X:          -1,
				Y:          10,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				gameState := models.NewGameState("campaign-123")
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
				s.On("Get", mock.Anything, "campaign-123").Return(gameState, nil)
			},
			expectError:   true,
			errorContains: "position cannot be negative",
		},
		{
			name: "position out of bounds",
			req: &service.MoveToRequest{
				CampaignID: "campaign-123",
				X:          100,
				Y:          10,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				gameState := models.NewGameState("campaign-123")
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
				s.On("Get", mock.Anything, "campaign-123").Return(gameState, nil)
			},
			expectError:   true,
			errorContains: "out of map bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore, gameStateStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			result, err := svc.MoveTo(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateState != nil {
					tt.validateState(t, result)
				}
			}

			mapStore.AssertExpectations(t)
			gameStateStore.AssertExpectations(t)
		})
	}
}

func TestMapService_AddLocation(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.AddLocationRequest
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap)
		expectError   bool
		errorContains string
		validateLoc   func(t *testing.T, loc *models.Location)
	}{
		{
			name: "successful add location",
			req: &service.AddLocationRequest{
				CampaignID:  "campaign-123",
				Name:        "Village of Hommlet",
				Description: "A small village",
				X:           10,
				Y:           15,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateLoc: func(t *testing.T, loc *models.Location) {
				assert.NotNil(t, loc)
				assert.Equal(t, "Village of Hommlet", loc.Name)
				assert.Equal(t, "A small village", loc.Description)
				assert.Equal(t, 10, loc.Position.X)
				assert.Equal(t, 15, loc.Position.Y)
			},
		},
		{
			name: "empty campaign ID",
			req: &service.AddLocationRequest{
				Name: "Village",
				X:    10,
				Y:    15,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
		{
			name: "empty name",
			req: &service.AddLocationRequest{
				CampaignID: "campaign-123",
				X:          10,
				Y:          15,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "name is required",
		},
		{
			name: "position out of bounds",
			req: &service.AddLocationRequest{
				CampaignID:  "campaign-123",
				Name:        "Village",
				Description: "A village",
				X:           100,
				Y:           15,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
			},
			expectError:   true,
			errorContains: "out of map bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			loc, err := svc.AddLocation(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateLoc != nil {
					tt.validateLoc(t, loc)
				}
			}

			mapStore.AssertExpectations(t)
		})
	}
}

func TestMapService_UpdateLocation(t *testing.T) {
	tests := []struct {
		name          string
		campaignID    string
		locationID    string
		updates       service.LocationUpdate
		setupMocks    func(*MockMapStore)
		expectError   bool
		errorContains string
		validateLoc   func(t *testing.T, loc *models.Location)
	}{
		{
			name:       "successful update location name",
			campaignID: "campaign-123",
			locationID: "loc-123",
			updates: service.LocationUpdate{
				Name: strPtr("New Name"),
			},
			setupMocks: func(m *MockMapStore) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				location := models.NewLocation("Old Name", "Description", 10, 15)
				location.ID = "loc-123"
				worldMap.AddLocation(*location)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, "New Name", loc.Name)
			},
		},
		{
			name:       "successful update location position",
			campaignID: "campaign-123",
			locationID: "loc-123",
			updates: service.LocationUpdate{
				X: intPtr(20),
				Y: intPtr(25),
			},
			setupMocks: func(m *MockMapStore) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				location := models.NewLocation("Location", "Description", 10, 15)
				location.ID = "loc-123"
				worldMap.AddLocation(*location)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateLoc: func(t *testing.T, loc *models.Location) {
				assert.Equal(t, 20, loc.Position.X)
				assert.Equal(t, 25, loc.Position.Y)
			},
		},
		{
			name:       "empty campaign ID",
			campaignID: "",
			locationID: "loc-123",
			updates:    service.LocationUpdate{},
			setupMocks: func(m *MockMapStore) {},
			expectError: true,
			errorContains: "campaign ID is required",
		},
		{
			name:       "empty location ID",
			campaignID: "campaign-123",
			locationID: "",
			updates:    service.LocationUpdate{},
			setupMocks: func(m *MockMapStore) {},
			expectError: true,
			errorContains: "location ID is required",
		},
		{
			name:       "location not found",
			campaignID: "campaign-123",
			locationID: "nonexistent",
			updates:    service.LocationUpdate{},
			setupMocks: func(m *MockMapStore) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
			},
			expectError: true,
			errorContains: "location not found",
		},
		{
			name:       "new position out of bounds",
			campaignID: "campaign-123",
			locationID: "loc-123",
			updates: service.LocationUpdate{
				X: intPtr(100),
			},
			setupMocks: func(m *MockMapStore) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				location := models.NewLocation("Location", "Description", 10, 15)
				location.ID = "loc-123"
				worldMap.AddLocation(*location)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
			},
			expectError: true,
			errorContains: "out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			loc, err := svc.UpdateLocation(context.Background(), tt.campaignID, tt.locationID, tt.updates)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateLoc != nil {
					tt.validateLoc(t, loc)
				}
			}

			mapStore.AssertExpectations(t)
		})
	}
}

// Helper functions for pointer creation
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestMapService_MoveToken(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.TokenMoveRequest
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap)
		expectError   bool
		errorContains string
		validateResult func(t *testing.T, result *service.TokenMoveResult)
	}{
		{
			name: "successful token move - simple movement",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				// Create a battle map
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				// Add a token
				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *service.TokenMoveResult) {
				assert.Equal(t, "token-001", result.Token.ID)
				assert.Equal(t, 5, result.Token.Position.X)
				assert.Equal(t, 0, result.Token.Position.Y)
				// 5 squares * 5 feet = 25 feet
				assert.Equal(t, 25, result.MovementUsed)
				assert.Equal(t, 5, result.RemainingSpeed) // 30 - 25 = 5
			},
		},
		{
			name: "token move with custom speed",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        4,
				ToY:        0,
				Speed:      intPtr(20),
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *service.TokenMoveResult) {
				// 4 squares * 5 feet = 20 feet
				assert.Equal(t, 20, result.MovementUsed)
				assert.Equal(t, 0, result.RemainingSpeed) // 20 - 20 = 0
			},
		},
		{
			name: "diagonal movement has increased cost",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        2,
				ToY:        2,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			},
			expectError: false,
			validateResult: func(t *testing.T, result *service.TokenMoveResult) {
				// Diagonal: 2+2=4 squares, with 1.5x multiplier = 6 squares * 5 = 30 feet
				assert.GreaterOrEqual(t, result.MovementUsed, 20) // Should cost more than straight 4 squares
				assert.LessOrEqual(t, result.MovementUsed, 30)     // But not more than available
			},
		},
		{
			name: "empty campaign ID",
			req: &service.TokenMoveRequest{
				MapID:   "map-001",
				TokenID: "token-001",
				ToX:     5,
				ToY:     0,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
		{
			name: "empty map ID",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				TokenID:    "token-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "map ID is required",
		},
		{
			name: "empty token ID",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap) {},
			expectError:   true,
			errorContains: "token ID is required",
		},
		{
			name: "map not found",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "nonexistent",
				TokenID:    "token-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				m.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("map not found"))
			},
			expectError:   true,
			errorContains: "failed to get map",
		},
		{
			name: "token not found on map",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "nonexistent-token",
				ToX:        5,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"
				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			},
			expectError:   true,
			errorContains: "token not found",
		},
		{
			name: "position out of bounds",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        25, // Beyond map width of 20
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			},
			expectError:   true,
			errorContains: "out of map bounds",
		},
		{
			name: "negative position",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        -1,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			},
			expectError:   true,
			errorContains: "out of map bounds",
		},
		{
			name: "insufficient movement",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        10,
				ToY:        0,
				Speed:      intPtr(20), // Not enough speed for 10 squares (50 feet)
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			},
			expectError:   true,
			errorContains: "insufficient movement",
		},
		{
			name: "not a battle map",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "world-001",
				TokenID:    "token-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				worldMap.ID = "world-001"
				m.On("Get", mock.Anything, "world-001").Return(worldMap, nil)
			},
			expectError:   true,
			errorContains: "only supported on battle maps",
		},
		{
			name: "wrong campaign",
			req: &service.TokenMoveRequest{
				CampaignID: "campaign-456",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        5,
				ToY:        0,
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap) {
				battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
				battleMap.ID = "map-001"

				token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
				token.ID = "token-001"
				battleMap.AddToken(*token)

				m.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			},
			expectError:   true,
			errorContains: "does not belong to the specified campaign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			result, err := svc.MoveToken(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			mapStore.AssertExpectations(t)
		})
	}
}

func TestMapService_MoveToken_DifficultTerrain(t *testing.T) {
	mapStore := new(MockMapStore)
	campaignStore := new(MockCampaignStoreForMap)
	gameStateStore := new(MockGameStateStoreForMap)

	// Create a battle map with difficult terrain
	battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
	battleMap.ID = "map-001"

	// Set some cells as difficult terrain
	battleMap.Grid.SetCell(1, 0, models.CellTypeDifficult)
	battleMap.Grid.SetCell(2, 0, models.CellTypeDifficult)

	token := models.NewToken("char-001", 0, 0, models.TokenSizeMedium)
	token.ID = "token-001"
	battleMap.AddToken(*token)

	mapStore.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
	mapStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)

	svc := service.NewMapService(mapStore, campaignStore, gameStateStore)

	req := &service.TokenMoveRequest{
		CampaignID: "campaign-123",
		MapID:      "map-001",
		TokenID:    "token-001",
		ToX:        3,
		ToY:        0,
	}

	result, err := svc.MoveToken(context.Background(), req)
	assert.NoError(t, err)

	// Movement should be more expensive due to difficult terrain
	// 3 squares * 5 feet = 15 feet, doubled for difficult terrain = 30 feet
	assert.GreaterOrEqual(t, result.MovementUsed, 15)
}

func TestMapService_MoveToken_LargeToken(t *testing.T) {
	mapStore := new(MockMapStore)
	campaignStore := new(MockCampaignStoreForMap)
	gameStateStore := new(MockGameStateStoreForMap)

	// Create a battle map
	battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
	battleMap.ID = "map-001"

	// Add a Large token (2x2 squares)
	token := models.NewToken("char-001", 0, 0, models.TokenSizeLarge)
	token.ID = "token-001"
	battleMap.AddToken(*token)

	mapStore.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
	mapStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)

	svc := service.NewMapService(mapStore, campaignStore, gameStateStore)

	req := &service.TokenMoveRequest{
		CampaignID: "campaign-123",
		MapID:      "map-001",
		TokenID:    "token-001",
		ToX:        5,
		ToY:        0,
	}

	result, err := svc.MoveToken(context.Background(), req)
	assert.NoError(t, err)

	// Large token should move to (5, 0) - its top-left corner
	assert.Equal(t, 5, result.Token.Position.X)
	assert.Equal(t, 0, result.Token.Position.Y)
}

func TestMapService_MoveToken_TokenSizeConstraints(t *testing.T) {
	tests := []struct {
		name          string
		tokenSize     models.TokenSize
		fromX, fromY  int
		toX, toY      int
		expectError   bool
		errorContains string
		speed         *int // Optional custom speed to avoid insufficient movement errors
	}{
		{
			name:      "tiny token fits anywhere",
			tokenSize: models.TokenSizeTiny,
			fromX:     0,
			fromY:     0,
			toX:       19,
			toY:       19,
			expectError: false,
			speed:     intPtr(500), // High speed to avoid insufficient movement error
		},
		{
			name:      "medium token can move to edge",
			tokenSize: models.TokenSizeMedium,
			fromX:     0,
			fromY:     0,
			toX:       19,
			toY:       19,
			expectError: false, // Medium (1x1) can fit at (19,19) on a 20x20 map
			speed:     intPtr(500), // High speed for long distance
		},
		{
			name:      "medium token cannot go beyond edge",
			tokenSize: models.TokenSizeMedium,
			fromX:     0,
			fromY:     0,
			toX:       20, // Beyond map width of 20
			toY:       19,
			expectError: true,
			errorContains: "out of map bounds",
		},
		{
			name:      "large token needs 2x2 space - exceeds edge",
			tokenSize: models.TokenSizeLarge,
			fromX:     0,
			fromY:     0,
			toX:       19, // Beyond edge: 19 + 2 = 21 > 20
			toY:       19,
			expectError: true,
			errorContains: "out of map bounds",
			speed:     intPtr(500), // High speed to reach boundary check
		},
		{
			name:      "large token can fit at valid edge position",
			tokenSize: models.TokenSizeLarge,
			fromX:     0,
			fromY:     0,
			toX:       18,
			toY:       17, // Valid: 18+2 = 20, 17+2 = 19 <= 20
			expectError: false,
			speed:     intPtr(500), // High speed for long distance
		},
		{
			name:      "medium token short movement",
			tokenSize: models.TokenSizeMedium,
			fromX:     5,
			fromY:     5,
			toX:       10,
			toY:       5, // Straight line: 5 squares * 5 = 25 feet
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			battleMap := models.NewBattleMap("campaign-123", "Test Battle", 20, 20, 5)
			battleMap.ID = "map-001"

			token := models.NewToken("char-001", tt.fromX, tt.fromY, tt.tokenSize)
			token.ID = "token-001"
			battleMap.AddToken(*token)

			mapStore.On("Get", mock.Anything, "map-001").Return(battleMap, nil)
			if !tt.expectError {
				mapStore.On("Update", mock.Anything, mock.AnythingOfType("*models.Map")).Return(nil)
			}

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)

			req := &service.TokenMoveRequest{
				CampaignID: "campaign-123",
				MapID:      "map-001",
				TokenID:    "token-001",
				ToX:        tt.toX,
				ToY:        tt.toY,
				Speed:      tt.speed,
			}

			result, err := svc.MoveToken(context.Background(), req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mapStore.AssertExpectations(t)
		})
	}
}

func TestMapService_MoveToLocation(t *testing.T) {
	tests := []struct {
		name          string
		req           *service.MoveToLocationRequest
		setupMocks    func(*MockMapStore, *MockCampaignStoreForMap, *MockGameStateStoreForMap)
		expectError   bool
		errorContains string
		validateState func(t *testing.T, gameState *models.GameState, location *models.Location)
	}{
		{
			name: "successful move to location",
			req: &service.MoveToLocationRequest{
				CampaignID: "campaign-123",
				LocationID: "loc-123",
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				location := models.NewLocation("Village", "A small village", 10, 15)
				location.ID = "loc-123"
				worldMap.AddLocation(*location)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)

				gameState := models.NewGameState("campaign-123")
				s.On("Get", mock.Anything, "campaign-123").Return(gameState, nil)
				s.On("Update", mock.Anything, mock.AnythingOfType("*models.GameState")).Return(nil)
			},
			expectError: false,
			validateState: func(t *testing.T, gameState *models.GameState, location *models.Location) {
				assert.NotNil(t, gameState.PartyPosition)
				assert.Equal(t, 10, gameState.PartyPosition.X)
				assert.Equal(t, 15, gameState.PartyPosition.Y)
				assert.Equal(t, "Village", location.Name)
			},
		},
		{
			name: "empty campaign ID",
			req: &service.MoveToLocationRequest{
				LocationID: "loc-123",
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {},
			expectError:   true,
			errorContains: "campaign ID is required",
		},
		{
			name: "empty location ID",
			req: &service.MoveToLocationRequest{
				CampaignID: "campaign-123",
			},
			setupMocks:    func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {},
			expectError:   true,
			errorContains: "location ID is required",
		},
		{
			name: "location not found",
			req: &service.MoveToLocationRequest{
				CampaignID: "campaign-123",
				LocationID: "nonexistent",
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)
			},
			expectError:   true,
			errorContains: "location not found",
		},
		{
			name: "calculates travel time correctly",
			req: &service.MoveToLocationRequest{
				CampaignID: "campaign-123",
				LocationID: "loc-123",
			},
			setupMocks: func(m *MockMapStore, c *MockCampaignStoreForMap, s *MockGameStateStoreForMap) {
				worldMap := models.NewWorldMap("campaign-123", "World Map", 50, 50)
				location := models.NewLocation("Village", "A small village", 15, 20)
				location.ID = "loc-123"
				worldMap.AddLocation(*location)
				m.On("GetWorldMap", mock.Anything, "campaign-123").Return(worldMap, nil)

				gameState := models.NewGameState("campaign-123")
				s.On("Get", mock.Anything, "campaign-123").Return(gameState, nil)
				s.On("Update", mock.Anything, mock.AnythingOfType("*models.GameState")).Return(nil)
			},
			expectError: false,
			validateState: func(t *testing.T, gameState *models.GameState, location *models.Location) {
				// Starting from (0,0), traveling to (15,20) = 35 miles
				// At normal pace (3 mph): 35/3 = 11.67 hours
				// Game time should advance by 11 hours
				assert.Equal(t, 19, gameState.GameTime.Hour) // 8 AM + 11 hours = 7 PM
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapStore := new(MockMapStore)
			campaignStore := new(MockCampaignStoreForMap)
			gameStateStore := new(MockGameStateStoreForMap)

			tt.setupMocks(mapStore, campaignStore, gameStateStore)

			svc := service.NewMapService(mapStore, campaignStore, gameStateStore)
			gameState, location, err := svc.MoveToLocation(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gameState)
				assert.NotNil(t, location)
				if tt.validateState != nil {
					tt.validateState(t, gameState, location)
				}
			}

			mapStore.AssertExpectations(t)
			gameStateStore.AssertExpectations(t)
		})
	}
}
