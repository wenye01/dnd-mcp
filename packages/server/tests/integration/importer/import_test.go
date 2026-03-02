// Package importer_test provides integration tests for the map import functionality
package importer_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnd-mcp/server/internal/importer"
	"github.com/dnd-mcp/server/internal/importer/converter"
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
	"github.com/dnd-mcp/server/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMapStore implements importer.MapStoreForImport for testing
type mockMapStore struct {
	maps map[string]*models.Map
}

func newMockMapStore() *mockMapStore {
	return &mockMapStore{
		maps: make(map[string]*models.Map),
	}
}

func (s *mockMapStore) Create(ctx context.Context, gameMap *models.Map) error {
	s.maps[gameMap.ID] = gameMap
	return nil
}

func (s *mockMapStore) Get(ctx context.Context, id string) (*models.Map, error) {
	if m, ok := s.maps[id]; ok {
		return m, nil
	}
	return nil, os.ErrNotExist
}

func (s *mockMapStore) Update(ctx context.Context, gameMap *models.Map) error {
	s.maps[gameMap.ID] = gameMap
	return nil
}

func setupImportService(t *testing.T) *importer.ImportService {
	store := newMockMapStore()
	service := importer.NewImportService(store)

	// Register parsers
	uvttParser := parser.NewUVTTParser()
	service.RegisterParser(uvttParser)

	fvttSceneParser := parser.NewFVTTSceneParser()
	service.RegisterParser(fvttSceneParser)

	// Register converter for multiple formats
	// MapConverter handles both UVTT and FVTT Scene formats
	mapConverter := converter.NewMapConverter()
	service.RegisterConverterForFormat(mapConverter, format.FormatUVTT)
	service.RegisterConverterForFormat(mapConverter, format.FormatFVTTScene)

	return service
}

func TestImportService_ImportUVTT(t *testing.T) {
	service := setupImportService(t)

	// Read sample UVTT file
	uvttPath := filepath.Join("..", "..", "testdata", "maps", "sample.uvtt")
	data, err := os.ReadFile(uvttPath)
	require.NoError(t, err, "Failed to read sample UVTT file")

	// Import the map
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Format:       format.FormatUVTT,
		Name:         "UVTT Test Map", // Provide a name
		ImportTokens: true,
		ImportWalls:  true,
	}

	result, err := service.Import(context.Background(), data, opts)
	require.NoError(t, err, "Failed to import UVTT")
	require.NotNil(t, result)

	// Verify the result
	assert.NotNil(t, result.Map, "Map should not be nil")
	assert.Equal(t, "test-campaign", result.Map.CampaignID)
	assert.NotEmpty(t, result.Map.ID, "Map ID should be generated")
	assert.NotEmpty(t, result.Map.Name, "Map name should be set")

	// Verify grid
	if result.Map.Grid != nil {
		assert.Greater(t, result.Map.Grid.Width, 0, "Grid width should be positive")
		assert.Greater(t, result.Map.Grid.Height, 0, "Grid height should be positive")
	}

	// Verify meta
	assert.NotNil(t, result.Meta, "Import meta should be set")
	assert.Equal(t, format.FormatUVTT, result.Meta.SourceFormat)
}

func TestImportService_ImportFVTTScene(t *testing.T) {
	service := setupImportService(t)

	// Read sample FVTT Scene file
	scenePath := filepath.Join("..", "..", "testdata", "maps", "sample_scene.json")
	data, err := os.ReadFile(scenePath)
	require.NoError(t, err, "Failed to read sample FVTT Scene file")

	// Import the map
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Format:       format.FormatFVTTScene,
		ImportTokens: true,
		ImportWalls:  true,
	}

	result, err := service.Import(context.Background(), data, opts)
	require.NoError(t, err, "Failed to import FVTT Scene")
	require.NotNil(t, result)

	// Verify the result
	assert.NotNil(t, result.Map, "Map should not be nil")
	assert.Equal(t, "test-campaign", result.Map.CampaignID)
	assert.NotEmpty(t, result.Map.ID, "Map ID should be generated")
	assert.NotEmpty(t, result.Map.Name, "Map name should be set")

	// Verify meta
	assert.NotNil(t, result.Meta, "Import meta should be set")
	assert.Equal(t, format.FormatFVTTScene, result.Meta.SourceFormat)
}

func TestImportService_AutoDetectFormat(t *testing.T) {
	service := setupImportService(t)

	tests := []struct {
		name         string
		filePath     string
		expectedType format.ImportFormat
		mapName      string
	}{
		{
			name:         "UVTT file",
			filePath:     filepath.Join("..", "..", "testdata", "maps", "sample.uvtt"),
			expectedType: format.FormatUVTT,
			mapName:      "Auto-detected UVTT Map",
		},
		{
			name:         "FVTT Scene file",
			filePath:     filepath.Join("..", "..", "testdata", "maps", "sample_scene.json"),
			expectedType: format.FormatFVTTScene,
			mapName:      "", // FVTT Scene has name in data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.filePath)
			require.NoError(t, err, "Failed to read test file")

			opts := format.ImportOptions{
				CampaignID:   "test-campaign",
				Format:       format.FormatAuto, // Auto-detect
				Name:         tt.mapName,
				ImportTokens: true,
				ImportWalls:  true,
			}

			result, err := service.Import(context.Background(), data, opts)
			require.NoError(t, err, "Failed to import with auto-detect")
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedType, result.Meta.SourceFormat,
				"Should detect correct format")
		})
	}
}

func TestImportService_ImportAndSave(t *testing.T) {
	store := newMockMapStore()
	service := importer.NewImportService(store)

	// Register parsers and converters
	service.RegisterParser(parser.NewUVTTParser())
	mapConverter := converter.NewMapConverter()
	service.RegisterConverterForFormat(mapConverter, format.FormatUVTT)
	service.RegisterConverterForFormat(mapConverter, format.FormatFVTTScene)

	// Read sample UVTT file
	uvttPath := filepath.Join("..", "..", "testdata", "maps", "sample.uvtt")
	data, err := os.ReadFile(uvttPath)
	require.NoError(t, err, "Failed to read sample UVTT file")

	// Import and save
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Format:       format.FormatUVTT,
		Name:         "Imported Map", // Provide a name
		ImportTokens: true,
		ImportWalls:  true,
	}

	result, err := service.ImportAndSave(context.Background(), "test-campaign", data, opts)
	require.NoError(t, err, "Failed to import and save")
	require.NotNil(t, result)

	// Verify the map was saved
	assert.Equal(t, "test-campaign", result.Map.CampaignID)
	assert.Contains(t, store.maps, result.Map.ID, "Map should be saved in store")
}

func TestUVTTParser_Parse(t *testing.T) {
	p := parser.NewUVTTParser()

	// Read sample UVTT file
	uvttPath := filepath.Join("..", "..", "testdata", "maps", "sample.uvtt")
	data, err := os.ReadFile(uvttPath)
	require.NoError(t, err, "Failed to read sample UVTT file")

	// Check if parser can handle this data
	assert.True(t, p.CanParse(data), "Parser should recognize UVTT format")

	// Parse the data
	result, err := p.Parse(data)
	require.NoError(t, err, "Failed to parse UVTT")
	require.NotNil(t, result)

	assert.Equal(t, format.FormatUVTT, result.Format)
	assert.NotNil(t, result.Data)
}

func TestFVTTSceneParser_Parse(t *testing.T) {
	p := parser.NewFVTTSceneParser()

	// Read sample FVTT Scene file
	scenePath := filepath.Join("..", "..", "testdata", "maps", "sample_scene.json")
	data, err := os.ReadFile(scenePath)
	require.NoError(t, err, "Failed to read sample FVTT Scene file")

	// Check if parser can handle this data
	assert.True(t, p.CanParse(data), "Parser should recognize FVTT Scene format")

	// Parse the data
	result, err := p.Parse(data)
	require.NoError(t, err, "Failed to parse FVTT Scene")
	require.NotNil(t, result)

	assert.Equal(t, format.FormatFVTTScene, result.Format)
	assert.NotNil(t, result.Data)
}

func TestNDJSONParser_Open(t *testing.T) {
	p := parser.NewNDJSONParser()

	// Open the sample module
	modulePath := filepath.Join("..", "..", "testdata", "maps", "sample_module")
	err := p.Open(modulePath)
	require.NoError(t, err, "Failed to open sample module")
	defer p.Close()

	// List scenes
	scenes, err := p.ListScenes()
	require.NoError(t, err, "Failed to list scenes")
	assert.NotEmpty(t, scenes, "Module should have scenes")

	// Get module info
	info, err := p.GetModuleInfo()
	require.NoError(t, err, "Failed to get module info")
	assert.NotEmpty(t, info.Name, "Module should have a name")
	assert.NotEmpty(t, info.Title, "Module should have a title")
}

func TestNDJSONParser_GetScene(t *testing.T) {
	p := parser.NewNDJSONParser()

	// Open the sample module
	modulePath := filepath.Join("..", "..", "testdata", "maps", "sample_module")
	err := p.Open(modulePath)
	require.NoError(t, err, "Failed to open sample module")
	defer p.Close()

	// List scenes first
	scenes, err := p.ListScenes()
	require.NoError(t, err, "Failed to list scenes")
	require.NotEmpty(t, scenes, "Module should have scenes")

	// Get the first scene
	scene, err := p.GetScene(scenes[0])
	require.NoError(t, err, "Failed to get scene")
	require.NotNil(t, scene)

	assert.Equal(t, format.FormatFVTTScene, scene.Format)
	assert.NotNil(t, scene.Data)
}

func TestMapConverter_ConvertFromUVTT(t *testing.T) {
	c := converter.NewMapConverter()

	// Create sample UVTT data
	uvttData := &format.UVTTData{
		Format: 2,
		Resolution: format.UVTTResolution{
			MapSize:        format.UVTTMapSize{X: 20, Y: 15},
			PixelsPerGrid:  70,
		},
		LineOfSight: []format.UVTTLineOfSight{},
		Portals:     []format.UVTTPortal{},
		Lights:      []format.UVTTLight{},
		Tokens:      []format.UVTTToken{},
		Walls:       []format.UVTTWall{},
	}

	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Name:         "Test Map",
		ImportTokens: true,
		ImportWalls:  true,
	}

	gameMap, err := c.ConvertFromUVTT(uvttData, opts)
	require.NoError(t, err, "Failed to convert UVTT data")
	require.NotNil(t, gameMap)

	assert.Equal(t, "Test Map", gameMap.Name)
	assert.Equal(t, models.MapTypeBattle, gameMap.Type)
	assert.NotNil(t, gameMap.Grid)
	assert.Equal(t, 20, gameMap.Grid.Width)
	assert.Equal(t, 15, gameMap.Grid.Height)
}

func TestFormatDetector_Detect(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		expectedType format.ImportFormat
	}{
		{
			name:         "UVTT format",
			filePath:     filepath.Join("..", "..", "testdata", "maps", "sample.uvtt"),
			expectedType: format.FormatUVTT,
		},
		{
			name:         "FVTT Scene format",
			filePath:     filepath.Join("..", "..", "testdata", "maps", "sample_scene.json"),
			expectedType: format.FormatFVTTScene,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.filePath)
			require.NoError(t, err, "Failed to read test file")

			detector := importer.NewDefaultFormatDetector()
			detected := detector.Detect(data)

			assert.Equal(t, tt.expectedType, detected,
				"Should detect correct format for %s", tt.name)
		})
	}
}

func TestImportOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    format.ImportOptions
		wantErr bool
	}{
		{
			name: "Valid options",
			opts: format.ImportOptions{
				CampaignID: "test-campaign",
				Format:     format.FormatUVTT,
			},
			wantErr: false,
		},
		{
			name: "Missing campaign ID",
			opts: format.ImportOptions{
				Format: format.FormatUVTT,
			},
			wantErr: true,
		},
		{
			name: "Invalid format",
			opts: format.ImportOptions{
				CampaignID: "test-campaign",
				Format:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "Negative scale",
			opts: format.ImportOptions{
				CampaignID: "test-campaign",
				Format:     format.FormatUVTT,
				Scale:      -1.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImportFromModule(t *testing.T) {
	store := newMockMapStore()
	service := importer.NewImportService(store)

	// Register parsers and converters
	service.RegisterParser(parser.NewFVTTSceneParser())
	mapConverter := converter.NewMapConverter()
	service.RegisterConverterForFormat(mapConverter, format.FormatFVTTScene)

	// Import from module
	modulePath := filepath.Join("..", "..", "testdata", "maps", "sample_module")
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		ImportTokens: true,
		ImportWalls:  true,
	}

	result, err := service.ImportFromModule(context.Background(), "test-campaign", modulePath, "", opts)
	require.NoError(t, err, "Failed to import from module")
	require.NotNil(t, result)

	// Verify result
	assert.NotEmpty(t, result.Maps, "Should import at least one map")
	assert.NotNil(t, result.ModuleInfo, "Module info should be set")

	for _, gameMap := range result.Maps {
		assert.Equal(t, "test-campaign", gameMap.CampaignID)
	}

	t.Logf("Imported %d maps from module '%s'", len(result.Maps), result.ModuleInfo.Title)
	t.Logf("Warnings: %v", result.Warnings)
}

// TestJSONSerialization tests that import results can be serialized to JSON
func TestJSONSerialization(t *testing.T) {
	service := setupImportService(t)

	// Read sample UVTT file
	uvttPath := filepath.Join("..", "..", "testdata", "maps", "sample.uvtt")
	data, err := os.ReadFile(uvttPath)
	require.NoError(t, err, "Failed to read sample UVTT file")

	// Import the map
	opts := format.ImportOptions{
		CampaignID:   "test-campaign",
		Format:       format.FormatUVTT,
		Name:         "Serialization Test Map", // Provide a name
		ImportTokens: true,
		ImportWalls:  true,
	}

	result, err := service.Import(context.Background(), data, opts)
	require.NoError(t, err, "Failed to import UVTT")

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err, "Failed to serialize result to JSON")

	// Verify JSON is valid
	var unmarshaled format.ImportResult
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Failed to unmarshal JSON")

	t.Logf("Serialized result:\n%s", string(jsonData))
}
