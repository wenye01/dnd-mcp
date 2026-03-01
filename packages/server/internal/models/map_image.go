// Package models provides map image model definitions
// 规则参考: Foundry VTT Scene/Image Data Structure
package models

import (
	"fmt"
	"regexp"
	"strings"
)

// MapImage represents an image or background for a map
// Used for visual representation of maps in the VTT
type MapImage struct {
	// URL is the location of the image file
	URL string `json:"url"`

	// Texture is the path to a Foundry VTT texture (if using built-in)
	Texture string `json:"texture,omitempty"`

	// OffsetX is the horizontal offset in pixels
	OffsetX int `json:"offset_x,omitempty"`

	// OffsetY is the vertical offset in pixels
	OffsetY int `json:"offset_y,omitempty"`

	// ScaleX is the horizontal scale factor (1.0 = normal)
	ScaleX float64 `json:"scale_x,omitempty"`

	// ScaleY is the vertical scale factor (1.0 = normal)
	ScaleY float64 `json:"scale_y,omitempty"`

	// Rotation is the rotation angle in degrees
	Rotation float64 `json:"rotation,omitempty"`

	// Width is the displayed width in pixels
	Width int `json:"width,omitempty"`

	// Height is the displayed height in pixels
	Height int `json:"height,omitempty"`

	// ZIndex controls rendering order (higher = on top)
	ZIndex int `json:"z_index,omitempty"`
}

// NewMapImage creates a new map image from a URL
func NewMapImage(url string) *MapImage {
	return &MapImage{
		URL:      url,
		ScaleX:   1.0,
		ScaleY:   1.0,
		Rotation: 0,
		ZIndex:   0,
	}
}

// Validate validates the map image
func (img *MapImage) Validate() error {
	if img.URL == "" && img.Texture == "" {
		return NewValidationError("map_image.url", "url or texture must be provided")
	}

	// Validate URL format if provided
	if img.URL != "" {
		if !isValidImageURL(img.URL) {
			return NewValidationError("map_image.url", "invalid URL format")
		}
	}

	// Validate scale factors
	if img.ScaleX <= 0 {
		return NewValidationError("map_image.scale_x", "must be positive")
	}
	if img.ScaleY <= 0 {
		return NewValidationError("map_image.scale_y", "must be positive")
	}

	// Validate dimensions
	if img.Width < 0 {
		return NewValidationError("map_image.width", "cannot be negative")
	}
	if img.Height < 0 {
		return NewValidationError("map_image.height", "cannot be negative")
	}

	// Validate rotation (-360 to 360 degrees)
	if img.Rotation < -360 || img.Rotation > 360 {
		return NewValidationError("map_image.rotation", "must be between -360 and 360 degrees")
	}

	return nil
}

// isValidImageURL checks if the URL is a valid image URL
func isValidImageURL(url string) bool {
	// Check for common image protocols or paths
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	if strings.HasPrefix(url, "/") || strings.HasPrefix(url, "./") {
		return true
	}
	if strings.HasPrefix(url, "data:image/") {
		return true
	}

	// Check for common image extensions
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"}
	for _, ext := range imageExtensions {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return true
		}
	}

	return false
}

// GetDisplaySize calculates the actual display size in pixels
func (img *MapImage) GetDisplaySize() (width, height int) {
	if img.Width > 0 && img.Height > 0 {
		return img.Width, img.Height
	}

	// Return default size if not specified
	return 800, 600
}

// IsDataURI returns true if the image is embedded as a data URI
func (img *MapImage) IsDataURI() bool {
	return strings.HasPrefix(img.URL, "data:image/")
}

// GetDataURIFormat returns the format of a data URI image (e.g., "png", "jpeg")
// Returns empty string if not a data URI
func (img *MapImage) GetDataURIFormat() string {
	if !img.IsDataURI() {
		return ""
	}

	// data:image/png;base64,xxxxx
	re := regexp.MustCompile(`data:image/([a-zA-Z]+);base64`)
	matches := re.FindStringSubmatch(img.URL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Clone creates a deep copy of the map image
func (img *MapImage) Clone() *MapImage {
	return &MapImage{
		URL:      img.URL,
		Texture:  img.Texture,
		OffsetX:  img.OffsetX,
		OffsetY:  img.OffsetY,
		ScaleX:   img.ScaleX,
		ScaleY:   img.ScaleY,
		Rotation: img.Rotation,
		Width:    img.Width,
		Height:   img.Height,
		ZIndex:   img.ZIndex,
	}
}

// MapImages represents a collection of map images (for layered backgrounds)
type MapImages []*MapImage

// Add adds an image to the collection
func (imgs *MapImages) Add(img *MapImage) error {
	if err := img.Validate(); err != nil {
		return err
	}
	*imgs = append(*imgs, img)
	return nil
}

// GetPrimary returns the primary image (lowest z-index)
func (imgs MapImages) GetPrimary() *MapImage {
	if len(imgs) == 0 {
		return nil
	}

	primary := imgs[0]
	for _, img := range imgs {
		if img.ZIndex < primary.ZIndex {
			primary = img
		}
	}
	return primary
}

// GetOverlay returns the overlay image (highest z-index)
func (imgs MapImages) GetOverlay() *MapImage {
	if len(imgs) == 0 {
		return nil
	}

	overlay := imgs[0]
	for _, img := range imgs {
		if img.ZIndex > overlay.ZIndex {
			overlay = img
		}
	}
	return overlay
}

// Validate validates all images
func (imgs MapImages) Validate() error {
	for i, img := range imgs {
		if err := img.Validate(); err != nil {
			return fmt.Errorf("image at index %d: %w", i, err)
		}
	}
	return nil
}

// Clone creates a deep copy of all images
func (imgs MapImages) Clone() MapImages {
	result := make(MapImages, len(imgs))
	for i, img := range imgs {
		result[i] = img.Clone()
	}
	return result
}

// MapImportMeta contains metadata about map import
// Used when importing maps from external sources like Foundry VTT
type MapImportMeta struct {
	// SourceSystem indicates where the map was imported from
	SourceSystem string `json:"source_system,omitempty"`

	// SourceVersion is the version of the source system
	SourceVersion string `json:"source_version,omitempty"`

	// ImportDate is when the map was imported
	ImportDate string `json:"import_date,omitempty"`

	// OriginalID is the ID of the map in the source system
	OriginalID string `json:"original_id,omitempty"`

	// OriginalName is the original name of the map
	OriginalName string `json:"original_name,omitempty"`

	// ImportedBy is the user who imported the map
	ImportedBy string `json:"imported_by,omitempty"`

	// AutoScale indicates if the map was auto-scaled during import
	AutoScale bool `json:"auto_scale,omitempty"`

	// GridForce indicates if the grid was forced to match during import
	GridForce bool `json:"grid_force,omitempty"`
}

// NewMapImportMeta creates import metadata for a Foundry VTT scene
func NewMapImportMeta(sceneID, sceneName string) *MapImportMeta {
	return &MapImportMeta{
		SourceSystem:  "foundryvtt",
		SourceVersion: "10",
		OriginalID:    sceneID,
		OriginalName:  sceneName,
		AutoScale:     true,
		GridForce:     false,
	}
}

// Validate validates import metadata
func (m *MapImportMeta) Validate() error {
	if m.SourceSystem == "" {
		return NewValidationError("map_import_meta.source_system", "cannot be empty")
	}
	return nil
}

// Clone creates a deep copy of import metadata
func (m *MapImportMeta) Clone() *MapImportMeta {
	return &MapImportMeta{
		SourceSystem:  m.SourceSystem,
		SourceVersion: m.SourceVersion,
		ImportDate:    m.ImportDate,
		OriginalID:    m.OriginalID,
		OriginalName:  m.OriginalName,
		ImportedBy:    m.ImportedBy,
		AutoScale:     m.AutoScale,
		GridForce:     m.GridForce,
	}
}

// GetSourceDisplayName returns a human-readable name for the source system
func (m *MapImportMeta) GetSourceDisplayName() string {
	switch m.SourceSystem {
	case "foundryvtt":
		return "Foundry VTT"
	case "roll20":
		return "Roll20"
	case "dungeonforge":
		return "Dungeon Forge"
	case "custom":
		return "Custom"
	default:
		return m.SourceSystem
	}
}
