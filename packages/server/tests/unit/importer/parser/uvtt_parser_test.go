// Package parser_test provides tests for UVTT parser
package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUVTTParser_CanParse(t *testing.T) {
	p := parser.NewUVTTParser()

	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "valid UVTT format v2",
			data:     []byte(`{"format": 2, "resolution": {"pixels_per_grid": 140, "map_size": {"x": 10, "y": 10}}}`),
			expected: true,
		},
		{
			name:     "valid UVTT format v1",
			data:     []byte(`{"format": 1, "resolution": {"pixels_per_grid": 100, "map_size": {"x": 20, "y": 15}}}`),
			expected: true,
		},
		{
			name:     "UVTT with additional fields",
			data:     []byte(`{"format": 2, "resolution": {"pixels_per_grid": 140, "map_size": {"x": 10, "y": 10}}, "walls": [], "portals": []}`),
			expected: true,
		},
		{
			name:     "empty data",
			data:     []byte(""),
			expected: false,
		},
		{
			name:     "non-JSON data",
			data:     []byte("not json"),
			expected: false,
		},
		{
			name:     "JSON without format field",
			data:     []byte(`{"resolution": {"pixels_per_grid": 100}}`),
			expected: false,
		},
		{
			name:     "JSON without resolution field",
			data:     []byte(`{"format": 1}`),
			expected: false,
		},
		{
			name:     "array instead of object",
			data:     []byte(`[1, 2, 3]`),
			expected: false,
		},
		{
			name:     "binary data",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.CanParse(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUVTTParser_Parse(t *testing.T) {
	p := parser.NewUVTTParser()

	t.Run("valid UVTT v2 with all fields", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10},
				"map_orientation": "straight"
			},
			"line_of_sight": [
				{
					"polygon": [
						{"x": 140, "y": 140},
						{"x": 560, "y": 140},
						{"x": 560, "y": 560},
						{"x": 140, "y": 560}
					],
					"restricted": true
				}
			],
			"portals": [
				{
					"bounds": {"x": 140, "y": 560, "w": 100, "h": 10},
					"position": "N",
					"closed": false
				}
			],
			"walls": [
				{
					"bounds": {"x": 140, "y": 140, "w": 420, "h": 10},
					"door": false,
					"move": "normal"
				}
			]
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, format.FormatUVTT, result.Format)

		uvttData, ok := result.Data.(*format.UVTTData)
		require.True(t, ok)
		assert.Equal(t, 2.0, uvttData.Format)
		assert.Equal(t, 140, uvttData.Resolution.PixelsPerGrid)
		assert.Equal(t, 10, uvttData.Resolution.MapSize.X)
		assert.Equal(t, 10, uvttData.Resolution.MapSize.Y)
		assert.Len(t, uvttData.Walls, 1)
		assert.Len(t, uvttData.Portals, 1)
		assert.Len(t, uvttData.LineOfSight, 1)
	})

	t.Run("valid UVTT v1 minimal", func(t *testing.T) {
		data := []byte(`{
			"format": 1,
			"resolution": {
				"pixels_per_grid": 100,
				"map_size": {"x": 20, "y": 15}
			}
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, format.FormatUVTT, result.Format)

		uvttData, ok := result.Data.(*format.UVTTData)
		require.True(t, ok)
		assert.Equal(t, 1.0, uvttData.Format)
		assert.Equal(t, 100, uvttData.Resolution.PixelsPerGrid)
	})

	t.Run("UVTT with image data", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10}
			},
			"image": "base64_encoded_image_data"
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)

		uvttData, ok := result.Data.(*format.UVTTData)
		require.True(t, ok)
		assert.Equal(t, "base64_encoded_image_data", uvttData.Image)
	})

	t.Run("UVTT with tokens", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10}
			},
			"tokens": [
				{
					"name": "Goblin",
					"x": 140,
					"y": 140,
					"size": 1,
					"hidden": false
				}
			]
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)

		uvttData, ok := result.Data.(*format.UVTTData)
		require.True(t, ok)
		assert.Len(t, uvttData.Tokens, 1)
		assert.Equal(t, "Goblin", uvttData.Tokens[0].Name)
	})
}

func TestUVTTParser_Parse_Errors(t *testing.T) {
	p := parser.NewUVTTParser()

	tests := []struct {
		name        string
		data        []byte
		expectedErr string
	}{
		{
			name:        "invalid JSON",
			data:        []byte(`{invalid json}`),
			expectedErr: "failed to parse JSON",
		},
		{
			name:        "missing pixels_per_grid",
			data:        []byte(`{"format": 2, "resolution": {"map_size": {"x": 10, "y": 10}}}`),
			expectedErr: "invalid pixels_per_grid",
		},
		{
			name:        "zero pixels_per_grid",
			data:        []byte(`{"format": 2, "resolution": {"pixels_per_grid": 0, "map_size": {"x": 10, "y": 10}}}`),
			expectedErr: "invalid pixels_per_grid",
		},
		{
			name:        "negative pixels_per_grid",
			data:        []byte(`{"format": 2, "resolution": {"pixels_per_grid": -10, "map_size": {"x": 10, "y": 10}}}`),
			expectedErr: "invalid pixels_per_grid",
		},
		{
			name:        "missing map_size.x",
			data:        []byte(`{"format": 2, "resolution": {"pixels_per_grid": 100, "map_size": {"y": 10}}}`),
			expectedErr: "invalid map dimensions",
		},
		{
			name:        "zero map_size.x",
			data:        []byte(`{"format": 2, "resolution": {"pixels_per_grid": 100, "map_size": {"x": 0, "y": 10}}}`),
			expectedErr: "invalid map dimensions",
		},
		{
			name:        "zero map_size.y",
			data:        []byte(`{"format": 2, "resolution": {"pixels_per_grid": 100, "map_size": {"x": 10, "y": 0}}}`),
			expectedErr: "invalid map dimensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.data)
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestUVTTParser_Parse_Warnings(t *testing.T) {
	p := parser.NewUVTTParser()

	t.Run("warning for lights", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10}
			},
			"lights": [
				{"position": {"x": 100, "y": 100}, "range": 30}
			]
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "lights")
	})

	t.Run("warning for both image and images", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10}
			},
			"image": "base64_data",
			"images": [{"id": "img1", "data": "more_data"}]
		}`)

		result, err := p.Parse(data)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Warnings)
		assert.Contains(t, result.Warnings[0], "both embedded image and image references")
	})
}

func TestUVTTParser_Format(t *testing.T) {
	p := parser.NewUVTTParser()
	assert.Equal(t, format.FormatUVTT, p.Format())
}

func TestParseUVTT(t *testing.T) {
	t.Run("convenience function returns correct data", func(t *testing.T) {
		data := []byte(`{
			"format": 2,
			"resolution": {
				"pixels_per_grid": 140,
				"map_size": {"x": 10, "y": 10}
			}
		}`)

		uvttData, err := parser.ParseUVTT(data)
		require.NoError(t, err)
		assert.NotNil(t, uvttData)
		assert.Equal(t, 2.0, uvttData.Format)
	})

	t.Run("convenience function returns error on invalid data", func(t *testing.T) {
		data := []byte(`invalid json`)

		uvttData, err := parser.ParseUVTT(data)
		assert.Error(t, err)
		assert.Nil(t, uvttData)
	})
}

func TestUVTTParser_Parse_SampleFile(t *testing.T) {
	// Get the project root directory
	projectRoot := filepath.Join("..", "..", "..", "..")
	testdataPath := filepath.Join(projectRoot, "tests", "testdata", "maps", "sample.uvtt")

	// Check if sample file exists
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skip("sample.uvtt test file not found")
		return
	}

	data, err := os.ReadFile(testdataPath)
	require.NoError(t, err)

	p := parser.NewUVTTParser()

	// Verify CanParse
	assert.True(t, p.CanParse(data), "sample.uvtt should be parsable")

	// Parse the file
	result, err := p.Parse(data)
	require.NoError(t, err)

	assert.Equal(t, format.FormatUVTT, result.Format)
	assert.NotNil(t, result.Data)

	uvttData, ok := result.Data.(*format.UVTTData)
	require.True(t, ok)

	// Verify expected data from sample.uvtt
	assert.Equal(t, 2.0, uvttData.Format)
	assert.Equal(t, 140, uvttData.Resolution.PixelsPerGrid)
	assert.Equal(t, 10, uvttData.Resolution.MapSize.X)
	assert.Equal(t, 10, uvttData.Resolution.MapSize.Y)
}

// Benchmark tests

func BenchmarkUVTTParser_CanParse(b *testing.B) {
	p := parser.NewUVTTParser()
	data := []byte(`{"format": 2, "resolution": {"pixels_per_grid": 140, "map_size": {"x": 10, "y": 10}}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.CanParse(data)
	}
}

func BenchmarkUVTTParser_Parse(b *testing.B) {
	p := parser.NewUVTTParser()
	data := []byte(`{
		"format": 2,
		"resolution": {
			"pixels_per_grid": 140,
			"map_size": {"x": 10, "y": 10},
			"map_orientation": "straight"
		},
		"walls": [
			{"bounds": {"x": 140, "y": 140, "w": 420, "h": 10}, "door": false, "move": "normal"}
		],
		"portals": [],
		"line_of_sight": []
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(data)
	}
}
