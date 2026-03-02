// Package format defines types and constants for map import formats
package format

import (
	"time"

	"github.com/dnd-mcp/server/internal/models"
)

// ImportFormat represents the supported import formats
type ImportFormat string

const (
	// FormatUVTT Universal VTT format (.uvtt)
	FormatUVTT ImportFormat = "uvtt"
	// FormatFVTTScene Foundry VTT Scene format (.json)
	FormatFVTTScene ImportFormat = "fvtt_scene"
	// FormatFVTTModule Foundry VTT Module format (LevelDB compendium)
	FormatFVTTModule ImportFormat = "fvtt_module"
	// FormatAuto Auto-detect format from data
	FormatAuto ImportFormat = "auto"
)

// String returns the string representation of the format
func (f ImportFormat) String() string {
	return string(f)
}

// ImportOptions defines options for map import
type ImportOptions struct {
	// Format specifies the import format (use "auto" for auto-detection)
	Format ImportFormat `json:"format"`

	// Name is the custom name for the imported map
	// If empty, uses the name from the source data
	Name string `json:"name,omitempty"`

	// CampaignID is the campaign to associate the map with
	CampaignID string `json:"campaign_id"`

	// ImportTokens indicates whether to import tokens
	ImportTokens bool `json:"import_tokens"`

	// ImportWalls indicates whether to import walls
	ImportWalls bool `json:"import_walls"`

	// ImportLights indicates whether to import lighting data
	ImportLights bool `json:"import_lights"`

	// Scale is the custom scale factor for the map (1.0 = original)
	Scale float64 `json:"scale,omitempty"`

	// OffsetX is the X offset in grid units
	OffsetX int `json:"offset_x,omitempty"`

	// OffsetY is the Y offset in grid units
	OffsetY int `json:"offset_y,omitempty"`

	// OverwriteExisting indicates whether to overwrite existing maps with the same ID
	OverwriteExisting bool `json:"overwrite_existing,omitempty"`
}

// Validate validates the import options
func (opts *ImportOptions) Validate() error {
	if opts.CampaignID == "" {
		return &ValidationError{Field: "campaign_id", Message: "campaign ID is required"}
	}

	// Validate format
	validFormats := map[ImportFormat]bool{
		FormatUVTT:       true,
		FormatFVTTScene:  true,
		FormatFVTTModule: true,
		FormatAuto:       true,
	}
	if !validFormats[opts.Format] {
		return &ValidationError{Field: "format", Message: "unsupported format: " + string(opts.Format)}
	}

	// Validate scale
	if opts.Scale < 0 {
		return &ValidationError{Field: "scale", Message: "scale cannot be negative"}
	}

	return nil
}

// ImportResult represents the result of a map import operation
type ImportResult struct {
	// Map is the imported map
	Map *models.Map `json:"map"`

	// Warnings contains non-fatal warnings that occurred during import
	Warnings []string `json:"warnings,omitempty"`

	// Skipped contains information about what was skipped during import
	Skipped *SkippedInfo `json:"skipped,omitempty"`

	// Meta contains import metadata
	Meta *ImportMeta `json:"meta,omitempty"`
}

// AddWarning adds a warning to the result
func (r *ImportResult) AddWarning(message string) {
	r.Warnings = append(r.Warnings, message)
}

// SkippedInfo contains information about what was skipped during import
type SkippedInfo struct {
	TokensCount  int `json:"tokens_count,omitempty"` // Number of tokens skipped
	WallsCount   int `json:"walls_count,omitempty"`  // Number of walls skipped
	LightsCount  int `json:"lights_count,omitempty"` // Number of lights skipped
	OtherCount   int `json:"other_count,omitempty"`  // Number of other items skipped
	TotalSkipped int `json:"total_skipped"`          // Total items skipped
}

// ImportMeta contains metadata about the import operation
type ImportMeta struct {
	// SourceFormat is the detected or specified format
	SourceFormat ImportFormat `json:"source_format"`

	// SourceSystem is the system the map was imported from
	// (e.g., "foundryvtt", "roll20", "dungeonforge", "universalvtt")
	SourceSystem string `json:"source_system,omitempty"`

	// SourceVersion is the version of the source system
	SourceVersion string `json:"source_version,omitempty"`

	// ImportTimestamp is when the import was performed
	ImportTimestamp time.Time `json:"import_timestamp"`

	// OriginalID is the ID of the map in the source system
	OriginalID string `json:"original_id,omitempty"`

	// OriginalName is the original name of the map
	OriginalName string `json:"original_name,omitempty"`

	// ImportDuration is how long the import took
	ImportDuration time.Duration `json:"import_duration,omitempty"`

	// DataSize is the size of the imported data in bytes (alias for SourceSize)
	DataSize int64 `json:"data_size,omitempty"`

	// SourceSize is the size of the imported data in bytes
	SourceSize int64 `json:"source_size,omitempty"`

	// ImportTime is when the import was performed (alias for ImportTimestamp)
	ImportTime time.Time `json:"import_time,omitempty"`

	// SourceName is the original name of the map (alias for OriginalName)
	SourceName string `json:"source_name,omitempty"`

	// AutoScale indicates if auto-scaling was applied
	AutoScale bool `json:"auto_scale,omitempty"`

	// HasImage indicates if the imported map has an image
	HasImage bool `json:"has_image,omitempty"`

	// ImageDimensions contains the image dimensions if available
	ImageDimensions *ImageDimensions `json:"image_dimensions,omitempty"`
}

// NewImportMeta creates new import metadata
func NewImportMeta(format ImportFormat) *ImportMeta {
	return &ImportMeta{
		SourceFormat:    format,
		ImportTimestamp: time.Now(),
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ParseError represents an error during parsing
type ParseError struct {
	Format  ImportFormat `json:"format"`
	Message string       `json:"message"`
	Cause   error        `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.Cause != nil {
		return "parse error (" + e.Format.String() + "): " + e.Message + ": " + e.Cause.Error()
	}
	return "parse error (" + e.Format.String() + "): " + e.Message
}

// Unwrap returns the underlying cause
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new parse error
func NewParseError(format ImportFormat, message string, cause error) *ParseError {
	return &ParseError{
		Format:  format,
		Message: message,
		Cause:   cause,
	}
}

// ConvertError represents an error during conversion
type ConvertError struct {
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

// Error implements the error interface
func (e *ConvertError) Error() string {
	if e.Cause != nil {
		return "conversion error: " + e.Message + ": " + e.Cause.Error()
	}
	return "conversion error: " + e.Message
}

// Unwrap returns the underlying cause
func (e *ConvertError) Unwrap() error {
	return e.Cause
}

// NewConvertError creates a new conversion error
func NewConvertError(message string, cause error) *ConvertError {
	return &ConvertError{
		Message: message,
		Cause:   cause,
	}
}

// ParseResult represents the result of parsing import data
type ParseResult struct {
	// Data is the parsed data
	Data interface{} `json:"data"`

	// Warnings contains any warnings from parsing
	Warnings []string `json:"warnings,omitempty"`

	// Format is the detected format
	Format ImportFormat `json:"format"`
}

// ImageDimensions contains image dimension information
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ImportFromModuleResult represents the result of importing from a module
type ImportFromModuleResult struct {
	// Maps is the list of imported maps
	Maps []*models.Map `json:"maps"`

	// Warnings contains any warnings from the import
	Warnings []string `json:"warnings,omitempty"`

	// ModuleInfo contains information about the module
	ModuleInfo *ModuleInfo `json:"module_info,omitempty"`

	// Meta contains import metadata
	Meta *ImportMeta `json:"meta,omitempty"`
}

// ModuleInfo contains information about a module
type ModuleInfo struct {
	// Name is the module name
	Name string `json:"name"`

	// Title is the human-readable title
	Title string `json:"title"`

	// Description is the module description
	Description string `json:"description"`

	// Version is the module version
	Version string `json:"version"`

	// Author is the module author
	Author string `json:"author"`

	// System is the game system (e.g., "dnd5e")
	System string `json:"system"`

	// SceneCount is the number of scenes in the module
	SceneCount int `json:"scene_count"`
}

