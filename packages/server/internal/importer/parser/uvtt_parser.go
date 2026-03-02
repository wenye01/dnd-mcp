// Package parser provides parsers for various map import formats
package parser

import (
	"encoding/json"
	"fmt"

	"github.com/dnd-mcp/server/internal/importer/format"
)

// UVTTParser parses Universal VTT format files
type UVTTParser struct{}

// NewUVTTParser creates a new UVTT parser
func NewUVTTParser() *UVTTParser {
	return &UVTTParser{}
}

// CanParse checks if the data appears to be a UVTT file
func (p *UVTTParser) CanParse(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// UVTT files are JSON
	if data[0] != '{' {
		return false
	}

	// Try to parse as UVTT and check for required fields
	var uvtt map[string]interface{}
	if err := json.Unmarshal(data, &uvtt); err != nil {
		return false
	}

	// UVTT must have format and resolution fields
	_, hasFormat := uvtt["format"]
	_, hasResolution := uvtt["resolution"]

	return hasFormat && hasResolution
}

// Parse parses UVTT data into a structured format
func (p *UVTTParser) Parse(data []byte) (*format.ParseResult, error) {
	var uvtt format.UVTTData
	if err := json.Unmarshal(data, &uvtt); err != nil {
		return nil, format.NewParseError(format.FormatUVTT, "failed to parse JSON", err)
	}

	// Validate required fields
	if uvtt.Resolution.PixelsPerGrid <= 0 {
		return nil, format.NewParseError(format.FormatUVTT, "invalid pixels_per_grid", nil)
	}
	if uvtt.Resolution.MapSize.X <= 0 || uvtt.Resolution.MapSize.Y <= 0 {
		return nil, format.NewParseError(format.FormatUVTT, "invalid map dimensions", nil)
	}

	// Build warnings for optional features we might not fully support
	var warnings []string

	if len(uvtt.Lights) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d lights were parsed but may not be fully imported", len(uvtt.Lights)))
	}

	if uvtt.Image != "" && len(uvtt.Images) > 0 {
		warnings = append(warnings, "both embedded image and image references found; using embedded image")
	}

	return &format.ParseResult{
		Format:   format.FormatUVTT,
		Data:     &uvtt,
		Warnings: warnings,
	}, nil
}

// Format returns the format this parser handles
func (p *UVTTParser) Format() format.ImportFormat {
	return format.FormatUVTT
}

// ParseUVTT is a convenience function to parse UVTT data
func ParseUVTT(data []byte) (*format.UVTTData, error) {
	parser := NewUVTTParser()
	result, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}
	return result.Data.(*format.UVTTData), nil
}
