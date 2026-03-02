package importer

import (
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/models"
)

// Parser defines the interface for parsing raw map data
type Parser interface {
	// CanParse checks if this parser can handle the given data
	CanParse(data []byte) bool

	// Parse parses the raw data into a format-specific structure
	Parse(data []byte) (*format.ParseResult, error)

	// Format returns the format this parser handles
	Format() format.ImportFormat
}

// Converter defines the interface for converting parsed data to Map model
type Converter interface {
	// Convert converts parsed data to a Map model
	Convert(parsedData interface{}, opts format.ImportOptions) (*models.Map, error)

	// Format returns the format this converter handles
	Format() format.ImportFormat
}

// Validator defines the interface for validating imported maps
type Validator interface {
	// Validate validates a map and returns any validation errors
	Validate(gameMap *models.Map) error

	// ValidateWithWarnings validates a map and returns warnings
	ValidateWithWarnings(gameMap *models.Map) ([]string, error)
}

// FormatDetector detects the format of map data
type FormatDetector interface {
	// Detect attempts to detect the format of the given data
	Detect(data []byte) format.ImportFormat
}

// ModuleParser defines the interface for parsing FVTT modules
type ModuleParser interface {
	// Open opens a module at the given path
	Open(modulePath string) error

	// Close closes the module and releases resources
	Close() error

	// ListScenes returns a list of scene names/IDs in the module
	ListScenes() ([]string, error)

	// GetScene retrieves a specific scene by name or ID
	GetScene(nameOrID string) (*format.ParseResult, error)

	// GetModuleInfo returns information about the module
	GetModuleInfo() (*ModuleInfo, error)
}

// ModuleInfo contains information about an FVTT module
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
