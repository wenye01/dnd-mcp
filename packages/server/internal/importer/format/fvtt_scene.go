// Package format defines Foundry VTT Scene data structures
// 规则参考: Foundry VTT v10 Scene Data Format
package format

// FVTTScene represents a Foundry VTT Scene
// This is the main structure for Foundry VTT scene JSON files
type FVTTScene struct {
	// ID is the unique identifier for this scene
	ID string `json:"_id"`

	// Name is the display name of the scene
	Name string `json:"name"`

	// Active indicates if this is the currently active scene
	Active bool `json:"active"`

	// Navigation properties
	Navigation bool `json:"navigation"`
	NavOrder   int    `json:"navOrder"`
	NavName    string `json:"navName"`

	// Thumbnail image path
	Thumb string `json:"thumb"`

	// Scene dimensions in pixels
	Width  int `json:"width"`
	Height int `json:"height"`

	// Padding around the scene edges (percentage)
	Padding float64 `json:"padding"`

	// Initial view position
	Initial *FVTTInitialView `json:"initial"`

	// Background color
	BackgroundColor string `json:"backgroundColor"`

	// Grid configuration
	GridType   int     `json:"gridType"` // 0=none, 1=square, 2=hex
	Grid       int     `json:"grid"`     // Grid size in pixels
	ShiftX     int     `json:"shiftX"`   // Horizontal grid offset
	ShiftY     int     `json:"shiftY"`   // Vertical grid offset
	GridColor  string  `json:"gridColor"`
	GridAlpha  float64 `json:"gridAlpha"`
	GridUnits  string  `json:"gridUnits"`  // Distance unit (ft, m, etc.)
	GridDistance int   `json:"gridDistance"` // Distance per grid square

	// Vision settings
	TokenVision    bool    `json:"tokenVision"`
	FogExploration bool    `json:"fogExploration"`
	FogReset       int64   `json:"fogReset"`
	GlobalLight    bool    `json:"globalLight"`
	GlobalLightThreshold *int `json:"globalLightThreshold"`
	Darkness       float64 `json:"darkness"`

	// Scene content
	Drawings []FVTTDrawing `json:"drawings"`
	Tokens   []FVTTToken   `json:"tokens"`
	Lights   []FVTTLight   `json:"lights"`
	Notes    []FVTTNote    `json:"notes"`
	Sounds   []FVTTSound   `json:"sounds"`
	Templates[]FVTTTemplate `json:"templates"`
	Tiles    []FVTTTile    `json:"tiles"`
	Walls    []FVTTWall    `json:"walls"`

	// Audio playlist
	Playlist      interface{} `json:"playlist"`
	PlaylistSound interface{} `json:"playlistSound"`

	// Journal entry
	Journal interface{} `json:"journal"`

	// Weather effect
	Weather string `json:"weather"`

	// Folder organization
	Folder interface{} `json:"folder"`
	Sort   int `json:"sort"`

	// Permissions
	Permission FVTTPermission `json:"permission"`

	// Custom flags
	Flags map[string]interface{} `json:"flags"`
}

// FVTTInitialView represents the initial view position and zoom
type FVTTInitialView struct {
	X int `json:"x"`
	Y int `json:"y"`
	Scale float64 `json:"scale"`
}

// FVTTDrawing represents a drawing on the scene
type FVTTDrawing struct {
	ID     string `json:"_id"`
	Type   int    `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Points []int  `json:"points,omitempty"`
	Fill   string `json:"fill,omitempty"`
	Stroke string `json:"stroke,omitempty"`
}

// FVTTToken represents a token on the scene
type FVTTToken struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	DisplayName int    `json:"displayName"` // 0-40 (const.CONST_TOKEN_DISPLAY_MODES)
	ActorID     string `json:"actorId"`
	ActorLink   bool   `json:"actorLink"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Rotation    int    `json:"rotation"`
	Effects     []string `json:"effects"`
	Alpha       float64 `json:"alpha"`
	Hidden      bool    `json:"hidden"`
	Locked      bool    `json:"locked"`
	Disposition int    `json:"disposition,omitempty"` // -1=hostile, 0=neutral, 1=friendly
	DisplayBar  int    `json:"displayBar,omitempty"` // 0=none, 1=hover, 20=always
	Bar1        *FVTTTokenBar `json:"bar1,omitempty"`
	Bar2        *FVTTTokenBar `json:"bar2,omitempty"`
	Image       string `json:"img,omitempty"`
	Scale       int    `json:"scale,omitempty"`
	MirrorX     bool   `json:"mirrorX,omitempty"`
	MirrorY     bool   `json:"mirrorY,omitempty"`
	Flags       map[string]interface{} `json:"flags"`
}

// FVTTTokenBar represents a token's attribute bar
type FVTTTokenBar struct {
	Attribute string `json:"attribute"` // attributes.system.hp.value
	Value     int    `json:"value"`     // alias: attributes.system.hp.value
	Max       int    `json:"max"`       // alias: attributes.system.hp.max
}

// FVTTLight represents a light source on the scene
type FVTTLight struct {
	ID     string `json:"_id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Dim    int    `json:"dim"`    // Dim radius
	Bright int    `json:"bright"` // Bright radius
	Angle  int    `json:"angle"`
	T      string `json:"t"`      // Type: "l"=local, "g"=global
	Color  string `json:"color"`
	Alpha  float64 `json:"alpha"`
	Animation *FVTTLightAnimation `json:"animation,omitempty"`
}

// FVTTLightAnimation represents light animation settings
type FVTTLightAnimation struct {
	Type    string  `json:"type"`
	Speed   int     `json:"speed"`
	Intensity float64 `json:"intensity"`
}

// FVTTNote represents a note on the scene
type FVTTNote struct {
	ID       string `json:"_id"`
	EntryID  string `json:"entryId"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Icon     string `json:"icon"`
	IconSize int    `json:"iconSize"`
	Text     string `json:"text"`
	FontSize int    `json:"fontSize"`
}

// FVTTSound represents a sound on the scene
type FVTTSound struct {
	ID     string `json:"_id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Path   string `json:"path"`
	Repeat bool   `json:"repeat"`
	Volume int    `json:"volume"`
	Echo   bool   `json:"echo"`
}

// FVTTTemplate represents a measured template (AoE)
type FVTTTemplate struct {
	ID     string `json:"_id"`
	Type   string `json:"type"` // circle, ray, cone, rect
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Angle  float64 `json:"angle"`
	Color  string `json:"color"`
}

// FVTTTile represents a tile on the scene
type FVTTTile struct {
	ID          string `json:"_id"`
	Image       string `json:"img"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Z           int    `json:"z"`
	Rotation    int    `json:"rotation"`
	Alpha       float64 `json:"alpha"`
	Hidden      bool `json:"hidden"`
	Locked      bool `json:"locked"`
	Overhead    bool `json:"overhead"`
	DisplayState int `json:"displayState"` // 0=play, 1=hover, 2=inspect
	RO          *FVTTRoof `json:"ro,omitempty"`
}

// FVTTRoof represents roof data for overhead tiles
type FVTTRoof struct {
	Enabled  bool     `json:"enabled"`
	Closed   bool     `json:"closed"`
	Secure   bool     `json:"secure"`
	Thresh   float64  `json:"thresh"`
	Sound    int      `json:"sound"`
	Tint     string   `json:"tint"`
	AdvCam   bool     `json:"advCam"`
}

// FVTTWall represents a wall on the scene
// In Foundry VTT, walls are defined by two endpoints [x1, y1, x2, y2]
type FVTTWall struct {
	ID    string      `json:"_id"`
	C     [][]int     `json:"c"` // [[x1, y1], [x2, y2]]
	Move  int         `json:"move"` // Movement restriction (0=block, 1=difficult, 2=allow)
	Sense int         `json:"sense"` // Sensing restriction (0=block, 1=limited, 2=allow)
	Dir   int         `json:"dir"` // Direction (bitmask: 1=left, 2=both, 4=right)
	Door  int         `json:"door"` // Door type (0=none, 1=door, 2=secret)
	DS    int         `json:"ds"` // Door state (0=closed, 1=open, 2=locked)
	Flags map[string]interface{} `json:"flags"`
}

// FVTTPermission represents permission settings
type FVTTPermission struct {
	Default int `json:"default"`
}

// GetGridSizeInPixels returns the grid size in pixels
func (s *FVTTScene) GetGridSizeInPixels() int {
	return s.Grid
}

// GetGridUnits returns the grid distance unit
func (s *FVTTScene) GetGridUnits() string {
	if s.GridUnits == "" {
		return "ft"
	}
	return s.GridUnits
}

// GetGridDistance returns the distance per grid square
func (s *FVTTScene) GetGridDistance() int {
	if s.GridDistance == 0 {
		return 5 // Default 5 feet per square
	}
	return int(s.GridDistance)
}

// GetDimensionsInGrid returns the scene dimensions in grid squares
func (s *FVTTScene) GetDimensionsInGrid() (width, height int) {
	if s.Grid > 0 {
		width = s.Width / s.Grid
		height = s.Height / s.Grid
	}
	return
}

// IsSquareGrid returns true if the scene uses a square grid
func (s *FVTTScene) IsSquareGrid() bool {
	return s.GridType == 1
}

// IsHexGrid returns true if the scene uses a hexagonal grid
func (s *FVTTScene) IsHexGrid() bool {
	return s.GridType == 2
}

// HasGrid returns true if the scene has any grid
func (s *FVTTScene) HasGrid() bool {
	return s.GridType > 0
}

// GetWallType determines the type of wall from its properties
func (w *FVTTWall) GetWallType() string {
	if w.Door == 1 {
		return "door"
	}
	if w.Door == 2 {
		return "secret_door"
	}
	return "wall"
}

// IsDoor returns true if this wall is a door
func (w *FVTTWall) IsDoor() bool {
	return w.Door > 0
}

// BlocksMovement returns true if the wall blocks movement
func (w *FVTTWall) BlocksMovement() bool {
	if w.IsDoor() && w.DS == 1 { // Open door
		return false
	}
	return w.Move == 0
}

// BlocksVision returns true if the wall blocks vision
// Closed doors block vision, open doors do not
func (w *FVTTWall) BlocksVision() bool {
	if w.IsDoor() {
		// Closed or locked doors block vision
		return w.DS != 1
	}
	return w.Sense == 0
}

// GetDoorState returns the door state as a string
func (w *FVTTWall) GetDoorState() string {
	if w.Door == 0 {
		return "none"
	}
	switch w.DS {
	case 0:
		return "closed"
	case 1:
		return "open"
	case 2:
		return "locked"
	default:
		return "closed"
	}
}
