// Package models_test provides unit tests for visual location models
package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dnd-mcp/server/internal/models"
)

func TestMapMode(t *testing.T) {
	// Test MapMode constants
	assert.Equal(t, models.MapMode("grid"), models.MapModeGrid)
	assert.Equal(t, models.MapMode("image"), models.MapModeImage)
}

func TestNewVisualLocation(t *testing.T) {
	loc := models.NewVisualLocation("Riverwood", "A peaceful village by the river", "town", 0.5, 0.3)

	assert.NotEmpty(t, loc.ID)
	assert.Equal(t, "Riverwood", loc.Name)
	assert.Equal(t, "A peaceful village by the river", loc.Description)
	assert.Equal(t, "town", loc.Type)
	assert.Equal(t, 0.5, loc.PositionX)
	assert.Equal(t, 0.3, loc.PositionY)
	assert.False(t, loc.IsConfirmed)
}

func TestVisualLocation_Validate(t *testing.T) {
	tests := []struct {
		name        string
		loc         *models.VisualLocation
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid location",
			loc:  models.NewVisualLocation("Test", "A test location", "town", 0.5, 0.5),
		},
		{
			name: "valid location with custom name only",
			loc: &models.VisualLocation{
				CustomName: "Custom",
				PositionX:  0.5,
				PositionY:  0.5,
			},
		},
		{
			name: "missing name",
			loc: &models.VisualLocation{
				PositionX: 0.5,
				PositionY: 0.5,
			},
			expectError: true,
			errorMsg:    "name or custom_name must be provided",
		},
		{
			name: "position_x out of range (negative)",
			loc: &models.VisualLocation{
				Name:      "Test",
				PositionX: -0.1,
				PositionY: 0.5,
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "position_x out of range (greater than 1)",
			loc: &models.VisualLocation{
				Name:      "Test",
				PositionX: 1.5,
				PositionY: 0.5,
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "position_y out of range",
			loc: &models.VisualLocation{
				Name:      "Test",
				PositionX: 0.5,
				PositionY: 1.5,
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "boundary values (0, 0)",
			loc: &models.VisualLocation{
				Name:      "Test",
				PositionX: 0,
				PositionY: 0,
			},
		},
		{
			name: "boundary values (1, 1)",
			loc: &models.VisualLocation{
				Name:      "Test",
				PositionX: 1,
				PositionY: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loc.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVisualLocation_GetDisplayName(t *testing.T) {
	// Without custom name
	loc1 := models.NewVisualLocation("Original", "Description", "town", 0.5, 0.5)
	assert.Equal(t, "Original", loc1.GetDisplayName())

	// With custom name
	loc1.SetCustomName("Custom")
	assert.Equal(t, "Custom", loc1.GetDisplayName())
}

func TestVisualLocation_Confirm(t *testing.T) {
	loc := models.NewVisualLocation("Test", "Description", "town", 0.5, 0.5)
	assert.False(t, loc.IsConfirmed)

	loc.Confirm()
	assert.True(t, loc.IsConfirmed)
}

func TestVisualLocation_SetBattleMapID(t *testing.T) {
	loc := models.NewVisualLocation("Test", "Description", "town", 0.5, 0.5)
	assert.Empty(t, loc.BattleMapID)

	loc.SetBattleMapID("battle-123")
	assert.Equal(t, "battle-123", loc.BattleMapID)
}

func TestRect_Validate(t *testing.T) {
	tests := []struct {
		name        string
		rect        *models.Rect
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid rect",
			rect: &models.Rect{X: 0.1, Y: 0.2, Width: 0.5, Height: 0.3},
		},
		{
			name: "x out of range",
			rect: &models.Rect{X: -0.1, Y: 0.2, Width: 0.5, Height: 0.3},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "y out of range",
			rect: &models.Rect{X: 0.1, Y: 1.5, Width: 0.5, Height: 0.3},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "width zero",
			rect: &models.Rect{X: 0.1, Y: 0.2, Width: 0, Height: 0.3},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "height negative",
			rect: &models.Rect{X: 0.1, Y: 0.2, Width: 0.5, Height: -0.1},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rect.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRect_Contains(t *testing.T) {
	rect := &models.Rect{X: 0.2, Y: 0.2, Width: 0.5, Height: 0.5}

	tests := []struct {
		name     string
		x        float64
		y        float64
		expected bool
	}{
		{"inside center", 0.4, 0.4, true},
		{"inside edge left", 0.2, 0.4, true},
		{"inside edge top", 0.4, 0.2, true},
		{"inside edge right", 0.7, 0.4, true},
		{"inside edge bottom", 0.4, 0.7, true},
		{"outside left", 0.1, 0.4, false},
		{"outside right", 0.8, 0.4, false},
		{"outside top", 0.4, 0.1, false},
		{"outside bottom", 0.4, 0.8, false},
		{"corner bottom right", 0.7, 0.7, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, rect.Contains(tt.x, tt.y))
		})
	}
}

func TestLandmark_Validate(t *testing.T) {
	tests := []struct {
		name        string
		landmark    *models.Landmark
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid landmark",
			landmark: &models.Landmark{Name: "Ancient Tower", Description: "A ruined tower", PositionX: 0.5, PositionY: 0.5},
		},
		{
			name:        "empty name",
			landmark:    &models.Landmark{PositionX: 0.5, PositionY: 0.5},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "position_x out of range",
			landmark:    &models.Landmark{Name: "Test", PositionX: 1.5, PositionY: 0.5},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.landmark.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTerrainArea_Validate(t *testing.T) {
	tests := []struct {
		name        string
		terrain     *models.TerrainArea
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid terrain",
			terrain: &models.TerrainArea{
				Type:        "forest",
				Description: "Dense forest",
				Bounds:      models.Rect{X: 0.1, Y: 0.1, Width: 0.3, Height: 0.3},
			},
		},
		{
			name: "empty type",
			terrain: &models.TerrainArea{
				Description: "Dense forest",
				Bounds:      models.Rect{X: 0.1, Y: 0.1, Width: 0.3, Height: 0.3},
			},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name: "invalid bounds",
			terrain: &models.TerrainArea{
				Type:        "forest",
				Description: "Dense forest",
				Bounds:      models.Rect{X: -0.1, Y: 0.1, Width: 0.3, Height: 0.3},
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.terrain.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVisualAnalysis_Validate(t *testing.T) {
	tests := []struct {
		name        string
		analysis    *models.VisualAnalysis
		expectError bool
	}{
		{
			name: "valid analysis",
			analysis: &models.VisualAnalysis{
				Summary: "A vast wilderness",
				Locations: []models.VisualLocation{
					*models.NewVisualLocation("Town", "A small town", "town", 0.5, 0.5),
				},
			},
		},
		{
			name: "valid empty analysis",
			analysis: &models.VisualAnalysis{
				Summary: "Empty map",
			},
		},
		{
			name: "invalid location in analysis",
			analysis: &models.VisualAnalysis{
				Summary: "A vast wilderness",
				Locations: []models.VisualLocation{
					{PositionX: 0.5, PositionY: 0.5}, // Missing name
				},
			},
			expectError: true,
		},
		{
			name: "invalid terrain in analysis",
			analysis: &models.VisualAnalysis{
				Summary: "A vast wilderness",
				Terrains: []models.TerrainArea{
					{Type: "", Description: "Invalid"}, // Missing type
				},
			},
			expectError: true,
		},
		{
			name: "invalid landmark in analysis",
			analysis: &models.VisualAnalysis{
				Summary: "A vast wilderness",
				Landmarks: []models.Landmark{
					{PositionX: 0.5, PositionY: 0.5}, // Missing name
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.analysis.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPlayerMarker(t *testing.T) {
	marker := models.NewPlayerMarker(0.5, 0.3)

	assert.Equal(t, 0.5, marker.PositionX)
	assert.Equal(t, 0.3, marker.PositionY)
	assert.NotEmpty(t, marker.UpdatedAt)
	assert.Empty(t, marker.CurrentScene)
}

func TestPlayerMarker_Validate(t *testing.T) {
	tests := []struct {
		name        string
		marker      *models.PlayerMarker
		expectError bool
		errorMsg    string
	}{
		{
			name:   "valid marker",
			marker: models.NewPlayerMarker(0.5, 0.5),
		},
		{
			name: "boundary values (0, 0)",
			marker: &models.PlayerMarker{
				PositionX: 0,
				PositionY: 0,
			},
		},
		{
			name: "boundary values (1, 1)",
			marker: &models.PlayerMarker{
				PositionX: 1,
				PositionY: 1,
			},
		},
		{
			name: "position_x negative",
			marker: &models.PlayerMarker{
				PositionX: -0.1,
				PositionY: 0.5,
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
		{
			name: "position_y greater than 1",
			marker: &models.PlayerMarker{
				PositionX: 0.5,
				PositionY: 1.5,
			},
			expectError: true,
			errorMsg:    "must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.marker.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlayerMarker_SetPosition(t *testing.T) {
	marker := models.NewPlayerMarker(0.5, 0.5)
	assert.NotEmpty(t, marker.UpdatedAt)

	marker.SetPosition(0.7, 0.3)

	assert.Equal(t, 0.7, marker.PositionX)
	assert.Equal(t, 0.3, marker.PositionY)
	assert.NotEmpty(t, marker.UpdatedAt)
}

func TestPlayerMarker_SetScene(t *testing.T) {
	marker := models.NewPlayerMarker(0.5, 0.5)
	assert.NotEmpty(t, marker.UpdatedAt)

	marker.SetScene("A dark forest clearing")

	assert.Equal(t, "A dark forest clearing", marker.CurrentScene)
	assert.NotEmpty(t, marker.UpdatedAt)
}

// MapImage Tests

func TestNewMapImage(t *testing.T) {
	img := models.NewMapImage("https://example.com/map.jpg")

	assert.Equal(t, "https://example.com/map.jpg", img.URL)
	assert.Equal(t, 1.0, img.ScaleX)
	assert.Equal(t, 1.0, img.ScaleY)
	assert.Equal(t, 0.0, img.Rotation)
	assert.Equal(t, 0, img.ZIndex)
}

func TestMapImage_Validate(t *testing.T) {
	tests := []struct {
		name        string
		image       *models.MapImage
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid image with URL",
			image: models.NewMapImage("https://example.com/map.jpg"),
		},
		{
			name: "valid image with texture",
			image: &models.MapImage{
				Texture:  "some_texture",
				ScaleX:   1.0,
				ScaleY:   1.0,
				Rotation: 0,
			},
		},
		{
			name: "valid image with relative path",
			image: models.NewMapImage("./images/map.png"),
		},
		{
			name: "valid image with absolute path",
			image: models.NewMapImage("/images/map.png"),
		},
		{
			name: "valid image with data URI",
			image: models.NewMapImage("data:image/png;base64,abc123"),
		},
		{
			name:        "missing URL and texture",
			image:       &models.MapImage{ScaleX: 1.0, ScaleY: 1.0},
			expectError: true,
			errorMsg:    "url or texture must be provided",
		},
		{
			name: "negative scale_x",
			image: &models.MapImage{
				URL:    "https://example.com/map.jpg",
				ScaleX: -1.0,
				ScaleY: 1.0,
			},
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name: "zero scale_y",
			image: &models.MapImage{
				URL:    "https://example.com/map.jpg",
				ScaleX: 1.0,
				ScaleY: 0,
			},
			expectError: true,
			errorMsg:    "must be positive",
		},
		{
			name: "negative width",
			image: &models.MapImage{
				URL:    "https://example.com/map.jpg",
				ScaleX: 1.0,
				ScaleY: 1.0,
				Width:  -1,
			},
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name: "negative height",
			image: &models.MapImage{
				URL:     "https://example.com/map.jpg",
				ScaleX:  1.0,
				ScaleY:  1.0,
				Height:  -1,
			},
			expectError: true,
			errorMsg:    "cannot be negative",
		},
		{
			name: "rotation too large",
			image: &models.MapImage{
				URL:      "https://example.com/map.jpg",
				ScaleX:   1.0,
				ScaleY:   1.0,
				Rotation: 400,
			},
			expectError: true,
			errorMsg:    "must be between -360 and 360",
		},
		{
			name: "rotation too small",
			image: &models.MapImage{
				URL:      "https://example.com/map.jpg",
				ScaleX:   1.0,
				ScaleY:   1.0,
				Rotation: -400,
			},
			expectError: true,
			errorMsg:    "must be between -360 and 360",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.image.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMapImage_GetDisplaySize(t *testing.T) {
	tests := []struct {
		name           string
		image          *models.MapImage
		expectedWidth  int
		expectedHeight int
	}{
		{
			name: "explicit dimensions",
			image: &models.MapImage{
				URL:    "https://example.com/map.jpg",
				Width:  1920,
				Height: 1080,
			},
			expectedWidth:  1920,
			expectedHeight: 1080,
		},
		{
			name:           "default dimensions",
			image:          models.NewMapImage("https://example.com/map.jpg"),
			expectedWidth:  800,
			expectedHeight: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := tt.image.GetDisplaySize()
			assert.Equal(t, tt.expectedWidth, width)
			assert.Equal(t, tt.expectedHeight, height)
		})
	}
}

func TestMapImage_IsDataURI(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"data URI png", "data:image/png;base64,abc123", true},
		{"data URI jpeg", "data:image/jpeg;base64,xyz789", true},
		{"https URL", "https://example.com/map.jpg", false},
		{"http URL", "http://example.com/map.jpg", false},
		{"relative path", "./images/map.png", false},
		{"absolute path", "/images/map.png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := models.NewMapImage(tt.url)
			assert.Equal(t, tt.expected, img.IsDataURI())
		})
	}
}

func TestMapImage_GetDataURIFormat(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"data URI png", "data:image/png;base64,abc123", "png"},
		{"data URI jpeg", "data:image/jpeg;base64,xyz789", "jpeg"},
		{"data URI webp", "data:image/webp;base64,def456", "webp"},
		{"https URL", "https://example.com/map.jpg", ""},
		{"empty URL", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := models.NewMapImage(tt.url)
			assert.Equal(t, tt.expected, img.GetDataURIFormat())
		})
	}
}

func TestMapImage_Clone(t *testing.T) {
	original := &models.MapImage{
		URL:      "https://example.com/map.jpg",
		Texture:  "texture1",
		OffsetX:  10,
		OffsetY:  20,
		ScaleX:   1.5,
		ScaleY:   1.2,
		Rotation: 45,
		Width:    1920,
		Height:   1080,
		ZIndex:   5,
	}

	cloned := original.Clone()

	// Verify values match
	assert.Equal(t, original.URL, cloned.URL)
	assert.Equal(t, original.Texture, cloned.Texture)
	assert.Equal(t, original.OffsetX, cloned.OffsetX)
	assert.Equal(t, original.OffsetY, cloned.OffsetY)
	assert.Equal(t, original.ScaleX, cloned.ScaleX)
	assert.Equal(t, original.ScaleY, cloned.ScaleY)
	assert.Equal(t, original.Rotation, cloned.Rotation)
	assert.Equal(t, original.Width, cloned.Width)
	assert.Equal(t, original.Height, cloned.Height)
	assert.Equal(t, original.ZIndex, cloned.ZIndex)

	// Verify it's a deep copy (different pointer)
	assert.NotSame(t, original, cloned)
}

// MapImages Tests

func TestMapImages_Add(t *testing.T) {
	var images models.MapImages

	img1 := models.NewMapImage("https://example.com/map1.jpg")
	err := images.Add(img1)
	assert.NoError(t, err)
	assert.Len(t, images, 1)

	img2 := models.NewMapImage("https://example.com/map2.jpg")
	err = images.Add(img2)
	assert.NoError(t, err)
	assert.Len(t, images, 2)
}

func TestMapImages_AddInvalid(t *testing.T) {
	var images models.MapImages

	// Try to add invalid image
	invalidImg := &models.MapImage{
		ScaleX: 1.0,
		ScaleY: 1.0,
		// Missing URL and Texture
	}
	err := images.Add(invalidImg)
	assert.Error(t, err)
	assert.Len(t, images, 0)
}

func TestMapImages_GetPrimary(t *testing.T) {
	images := models.MapImages{
		{URL: "map1.jpg", ZIndex: 5},
		{URL: "map2.jpg", ZIndex: 2},
		{URL: "map3.jpg", ZIndex: 8},
	}

	primary := images.GetPrimary()
	assert.NotNil(t, primary)
	assert.Equal(t, "map2.jpg", primary.URL)
	assert.Equal(t, 2, primary.ZIndex)
}

func TestMapImages_GetPrimaryEmpty(t *testing.T) {
	var images models.MapImages
	primary := images.GetPrimary()
	assert.Nil(t, primary)
}

func TestMapImages_GetOverlay(t *testing.T) {
	images := models.MapImages{
		{URL: "map1.jpg", ZIndex: 5},
		{URL: "map2.jpg", ZIndex: 2},
		{URL: "map3.jpg", ZIndex: 8},
	}

	overlay := images.GetOverlay()
	assert.NotNil(t, overlay)
	assert.Equal(t, "map3.jpg", overlay.URL)
	assert.Equal(t, 8, overlay.ZIndex)
}

func TestMapImages_GetOverlayEmpty(t *testing.T) {
	var images models.MapImages
	overlay := images.GetOverlay()
	assert.Nil(t, overlay)
}

func TestMapImages_Clone(t *testing.T) {
	original := models.MapImages{
		{URL: "map1.jpg", ZIndex: 1},
		{URL: "map2.jpg", ZIndex: 2},
	}

	cloned := original.Clone()

	assert.Len(t, cloned, 2)
	assert.Equal(t, original[0].URL, cloned[0].URL)
	assert.Equal(t, original[1].URL, cloned[1].URL)

	// Verify it's a deep copy
	cloned[0].URL = "modified.jpg"
	assert.Equal(t, "map1.jpg", original[0].URL)
	assert.Equal(t, "modified.jpg", cloned[0].URL)
}

// MapImportMeta Tests

func TestNewMapImportMeta(t *testing.T) {
	meta := models.NewMapImportMeta("scene-123", "My Scene")

	assert.Equal(t, "foundryvtt", meta.SourceSystem)
	assert.Equal(t, "10", meta.SourceVersion)
	assert.Equal(t, "scene-123", meta.OriginalID)
	assert.Equal(t, "My Scene", meta.OriginalName)
	assert.True(t, meta.AutoScale)
	assert.False(t, meta.GridForce)
}

func TestMapImportMeta_Validate(t *testing.T) {
	tests := []struct {
		name        string
		meta        *models.MapImportMeta
		expectError bool
	}{
		{
			name: "valid import meta",
			meta: models.NewMapImportMeta("scene-123", "My Scene"),
		},
		{
			name:        "empty source system",
			meta:        &models.MapImportMeta{SourceSystem: ""},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMapImportMeta_GetSourceDisplayName(t *testing.T) {
	tests := []struct {
		source    string
		expected  string
	}{
		{"foundryvtt", "Foundry VTT"},
		{"roll20", "Roll20"},
		{"dungeonforge", "Dungeon Forge"},
		{"custom", "Custom"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			meta := &models.MapImportMeta{SourceSystem: tt.source}
			assert.Equal(t, tt.expected, meta.GetSourceDisplayName())
		})
	}
}

func TestMapImportMeta_Clone(t *testing.T) {
	original := &models.MapImportMeta{
		SourceSystem:  "foundryvtt",
		SourceVersion: "10",
		OriginalID:    "scene-123",
		OriginalName:  "My Scene",
		AutoScale:     true,
		GridForce:     false,
	}

	cloned := original.Clone()

	assert.Equal(t, original.SourceSystem, cloned.SourceSystem)
	assert.Equal(t, original.SourceVersion, cloned.SourceVersion)
	assert.Equal(t, original.OriginalID, cloned.OriginalID)
	assert.Equal(t, original.OriginalName, cloned.OriginalName)
	assert.Equal(t, original.AutoScale, cloned.AutoScale)
	assert.Equal(t, original.GridForce, cloned.GridForce)

	// Verify it's a deep copy
	cloned.SourceSystem = "modified"
	assert.Equal(t, "foundryvtt", original.SourceSystem)
	assert.Equal(t, "modified", cloned.SourceSystem)
}
