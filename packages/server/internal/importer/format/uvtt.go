// Package format defines types for UVTT format
package format

// UVTTData represents the root structure of a Universal VTT file
type UVTTData struct {
	// Format is the UVTT format version (typically 1.0 or 2)
	Format float64 `json:"format"`

	// Resolution contains map resolution information
	Resolution UVTTResolution `json:"resolution"`

	// Image contains base64-encoded image data (optional)
	Image string `json:"image,omitempty"`

	// Images is an array of image references (UVTT 2.x)
	Images []UVTTImage `json:"images,omitempty"`

	// Walls contains wall definitions
	Walls []UVTTWall `json:"walls,omitempty"`

	// Portals contains portal/door definitions
	Portals []UVTTPortal `json:"portals,omitempty"`

	// LineOfSight contains line of sight polygons
	LineOfSight []UVTTLineOfSight `json:"line_of_sight,omitempty"`

	// Lights contains light definitions
	Lights []UVTTLight `json:"lights,omitempty"`

	// Tokens contains token definitions
	Tokens []UVTTToken `json:"tokens,omitempty"`
}

// UVTTResolution contains resolution information
type UVTTResolution struct {
	// PixelsPerGrid is the number of pixels per grid square
	PixelsPerGrid int `json:"pixels_per_grid"`

	// MapSize contains the map dimensions in grid squares
	MapSize UVTTMapSize `json:"map_size"`

	// MapOrientation indicates the map orientation
	MapOrientation string `json:"map_orientation,omitempty"`
}

// UVTTMapSize represents map dimensions
type UVTTMapSize struct {
	// X is the width in grid squares
	X int `json:"x"`

	// Y is the height in grid squares
	Y int `json:"y"`
}

// UVTTImage represents an image reference
type UVTTImage struct {
	// ID is the image identifier
	ID string `json:"id"`

	// URL is the image URL or path
	URL string `json:"url,omitempty"`

	// Data is base64-encoded image data
	Data string `json:"data,omitempty"`
}

// UVTTWall represents a wall segment
type UVTTWall struct {
	// Bounds contains the wall bounding box
	Bounds UVTTBounds `json:"bounds"`

	// Door indicates if this wall is a door
	Door bool `json:"door,omitempty"`

	// Move indicates movement restriction (normal, none, etc.)
	Move string `json:"move,omitempty"`

	// Sense indicates sensing restriction
	Sense string `json:"sense,omitempty"`
}

// UVTTPortal represents a portal or door
type UVTTPortal struct {
	// Bounds contains the portal bounding box
	Bounds UVTTBounds `json:"bounds"`

	// Position indicates the portal position/direction
	Position string `json:"position,omitempty"`

	// Closed indicates if the portal is closed
	Closed bool `json:"closed,omitempty"`

	// Rotation is the portal rotation in degrees
	Rotation float64 `json:"rotation,omitempty"`
}

// UVTTBounds represents a bounding box
type UVTTBounds struct {
	// X is the left coordinate in pixels
	X int `json:"x"`

	// Y is the top coordinate in pixels
	Y int `json:"y"`

	// W is the width in pixels
	W int `json:"w"`

	// H is the height in pixels
	H int `json:"h"`
}

// UVTTLineOfSight represents a line of sight polygon
type UVTTLineOfSight struct {
	// Polygon contains the polygon vertices
	Polygon []UVTTPoint `json:"polygon"`

	// Restricted indicates if this area has restricted vision
	Restricted bool `json:"restricted,omitempty"`
}

// UVTTPoint represents a 2D point
type UVTTPoint struct {
	// X coordinate in pixels
	X int `json:"x"`

	// Y coordinate in pixels
	Y int `json:"y"`
}

// UVTTLight represents a light source
type UVTTLight struct {
	// Position contains the light position
	Position UVTTPoint `json:"position"`

	// Range is the light range in grid squares
	Range int `json:"range"`

	// Angle is the light angle in degrees (360 for omnidirectional)
	Angle int `json:"angle,omitempty"`

	// Color is the light color (hex string)
	Color string `json:"color,omitempty"`

	// Intensity is the light intensity (0-1)
	Intensity float64 `json:"intensity,omitempty"`
}

// UVTTToken represents a token on the map
type UVTTToken struct {
	// Name is the token name
	Name string `json:"name,omitempty"`

	// X is the x position in pixels
	X int `json:"x"`

	// Y is the y position in pixels
	Y int `json:"y"`

	// Size is the token size multiplier (1 = 1x1 grid)
	Size float64 `json:"size,omitempty"`

	// Image is the token image URL or data
	Image string `json:"image,omitempty"`

	// ID is a unique identifier
	ID string `json:"_id,omitempty"`

	// Hidden indicates if the token is hidden from players
	Hidden bool `json:"hidden,omitempty"`

	// Locked indicates if the token is locked
	Locked bool `json:"locked,omitempty"`
}
