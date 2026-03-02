package importer_test

import (
	"testing"

	"github.com/dnd-mcp/server/internal/importer"
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportFormatString(t *testing.T) {
	tests := []struct {
		name     string
		format   format.ImportFormat
		expected string
	}{
		{
			name:     "UVTT format",
			format:   format.FormatUVTT,
			expected: "uvtt",
		},
		{
			name:     "FVTT Scene format",
			format:   format.FormatFVTTScene,
			expected: "fvtt_scene",
		},
		{
			name:     "FVTT Module format",
			format:   format.FormatFVTTModule,
			expected: "fvtt_module",
		},
		{
			name:     "Auto format",
			format:   format.FormatAuto,
			expected: "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.String())
		})
	}
}

func TestImportOptionsValidation(t *testing.T) {
	tests := []struct {
		name      string
		options   format.ImportOptions
		wantErr   bool
		errField  string
		errMsg    string
	}{
		{
			name: "valid options with auto format",
			options: format.ImportOptions{
				Format:       format.FormatAuto,
				CampaignID:   "campaign-123",
				ImportTokens: true,
				ImportWalls:  true,
			},
			wantErr: false,
		},
		{
			name: "valid options with UVTT format",
			options: format.ImportOptions{
				Format:       format.FormatUVTT,
				CampaignID:   "campaign-123",
				Name:         "Test Map",
				ImportTokens: false,
				ImportWalls:  false,
			},
			wantErr: false,
		},
		{
			name: "missing campaign ID",
			options: format.ImportOptions{
				Format:       format.FormatAuto,
				CampaignID:   "",
				ImportTokens: true,
			},
			wantErr:  true,
			errField: "campaign_id",
			errMsg:   "campaign ID is required",
		},
		{
			name: "unsupported format",
			options: format.ImportOptions{
				Format:       format.ImportFormat("unknown"),
				CampaignID:   "campaign-123",
				ImportTokens: true,
			},
			wantErr:  true,
			errField: "format",
			errMsg:   "unsupported format",
		},
		{
			name: "negative scale",
			options: format.ImportOptions{
				Format:     format.FormatAuto,
				CampaignID: "campaign-123",
				Scale:      -1.0,
			},
			wantErr:  true,
			errField: "scale",
			errMsg:   "scale cannot be negative",
		},
		{
			name: "zero scale is valid",
			options: format.ImportOptions{
				Format:     format.FormatAuto,
				CampaignID: "campaign-123",
				Scale:      0,
			},
			wantErr: false,
		},
		{
			name: "positive scale is valid",
			options: format.ImportOptions{
				Format:     format.FormatAuto,
				CampaignID: "campaign-123",
				Scale:      2.5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

			if tt.wantErr {
				require.Error(t, err)
				validationErr, ok := err.(*format.ValidationError)
				require.True(t, ok, "error should be ValidationError")
				assert.Equal(t, tt.errField, validationErr.Field)
				assert.Contains(t, validationErr.Message, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestImportMeta(t *testing.T) {
	t.Run("NewImportMeta creates valid metadata", func(t *testing.T) {
		meta := format.NewImportMeta(format.FormatUVTT)

		assert.Equal(t, format.FormatUVTT, meta.SourceFormat)
		assert.False(t, meta.ImportTimestamp.IsZero())
	})

	t.Run("ImportResult_AddWarning", func(t *testing.T) {
		result := &format.ImportResult{
			Warnings: make([]string, 0),
		}

		result.AddWarning("first warning")
		result.AddWarning("second warning")

		assert.Len(t, result.Warnings, 2)
		assert.Contains(t, result.Warnings, "first warning")
		assert.Contains(t, result.Warnings, "second warning")
	})
}

func TestValidationError(t *testing.T) {
	err := format.NewValidationError("test_field", "test message")

	assert.Equal(t, "test_field", err.Field)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, "test_field: test message", err.Error())
}

func TestParseError(t *testing.T) {
	t.Run("ParseError without cause", func(t *testing.T) {
		err := format.NewParseError(format.FormatUVTT, "invalid data", nil)

		assert.Equal(t, format.FormatUVTT, err.Format)
		assert.Equal(t, "invalid data", err.Message)
		assert.Nil(t, err.Cause)
		assert.Contains(t, err.Error(), "uvtt")
		assert.Contains(t, err.Error(), "invalid data")
	})

	t.Run("ParseError with cause", func(t *testing.T) {
		cause := assert.AnError
		err := format.NewParseError(format.FormatFVTTScene, "parse failed", cause)

		assert.Equal(t, format.FormatFVTTScene, err.Format)
		assert.Equal(t, "parse failed", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Contains(t, err.Error(), "fvtt_scene")
		assert.Contains(t, err.Error(), "parse failed")
		assert.Contains(t, err.Error(), cause.Error())
	})

	t.Run("ParseError unwrap", func(t *testing.T) {
		cause := assert.AnError
		err := format.NewParseError(format.FormatUVTT, "test", cause)

		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestConvertError(t *testing.T) {
	t.Run("ConvertError without cause", func(t *testing.T) {
		err := format.NewConvertError("conversion failed", nil)

		assert.Equal(t, "conversion failed", err.Message)
		assert.Nil(t, err.Cause)
		assert.Contains(t, err.Error(), "conversion failed")
	})

	t.Run("ConvertError with cause", func(t *testing.T) {
		cause := assert.AnError
		err := format.NewConvertError("invalid map data", cause)

		assert.Equal(t, "invalid map data", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Contains(t, err.Error(), "invalid map data")
		assert.Contains(t, err.Error(), cause.Error())
	})

	t.Run("ConvertError unwrap", func(t *testing.T) {
		cause := assert.AnError
		err := format.NewConvertError("test", cause)

		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestSkippedInfo(t *testing.T) {
	skipped := &format.SkippedInfo{
		TokensCount:  5,
		WallsCount:   10,
		LightsCount:  3,
		OtherCount:   2,
		TotalSkipped: 20,
	}

	assert.Equal(t, 5, skipped.TokensCount)
	assert.Equal(t, 10, skipped.WallsCount)
	assert.Equal(t, 3, skipped.LightsCount)
	assert.Equal(t, 2, skipped.OtherCount)
	assert.Equal(t, 20, skipped.TotalSkipped)
}

func TestFormatDetector(t *testing.T) {
	detector := &importer.DefaultFormatDetector{}

	tests := []struct {
		name        string
		data        []byte
		wantFormat  format.ImportFormat
		description string
	}{
		{
			name:        "UVTT format detection",
			data:        []byte(`{"format": "universalvtt", "resolution": {"map_size": {"width": 20, "height": 20}}}`),
			wantFormat:  format.FormatUVTT,
			description: "Should detect Universal VTT format",
		},
		{
			name:        "FVTT Scene format detection",
			data:        []byte(`{"_id": "scene123", "name": "Test Scene", "grid": {"size": 100}}`),
			wantFormat:  format.FormatFVTTScene,
			description: "Should detect Foundry VTT scene format",
		},
		{
			name:        "empty data",
			data:        []byte(""),
			wantFormat:  format.FormatAuto,
			description: "Should return auto for empty data",
		},
		{
			name:        "unknown JSON format",
			data:        []byte(`{"foo": "bar", "baz": "qux"}`),
			wantFormat:  format.FormatAuto,
			description: "Should return auto for unrecognized format",
		},
		{
			name:        "binary data (LevelDB)",
			data:        []byte{0x00, 0x00, 0x01, 0x02, 0x03},
			wantFormat:  format.FormatFVTTModule,
			description: "Should detect binary as possible LevelDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFormat := detector.Detect(tt.data)
			assert.Equal(t, tt.wantFormat, gotFormat, tt.description)
		})
	}
}
