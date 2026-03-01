// Package models 提供领域模型定义
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CellType 格子类型
type CellType string

const (
	// CellTypeEmpty 空地
	CellTypeEmpty CellType = "empty"
	// CellTypeWall 墙壁
	CellTypeWall CellType = "wall"
	// CellTypeDifficult 困难地形
	CellTypeDifficult CellType = "difficult"
	// CellTypeWater 水域
	CellTypeWater CellType = "water"
	// CellTypeDoor 门
	CellTypeDoor CellType = "door"
	// CellTypeRoad 道路（大地图专用）
	CellTypeRoad CellType = "road"
	// CellTypeForest 森林（大地图专用）
	CellTypeForest CellType = "forest"
	// CellTypeMountain 山地（大地图专用）
	CellTypeMountain CellType = "mountain"
	// CellTypeBuilding 建筑（大地图专用）
	CellTypeBuilding CellType = "building"
)

// TokenSize Token大小
// 规则参考: PHB 第9章 - Size and Space
type TokenSize string

const (
	// TokenSizeTiny 微型 (2.5x2.5英尺)
	TokenSizeTiny TokenSize = "tiny"
	// TokenSizeSmall 小型 (5x5英尺)
	TokenSizeSmall TokenSize = "small"
	// TokenSizeMedium 中型 (5x5英尺)
	TokenSizeMedium TokenSize = "medium"
	// TokenSizeLarge 大型 (10x10英尺)
	TokenSizeLarge TokenSize = "large"
	// TokenSizeHuge 超大型 (15x15英尺)
	TokenSizeHuge TokenSize = "huge"
	// TokenSizeGargantuan 巨型 (20x20英尺或更大)
	TokenSizeGargantuan TokenSize = "gargantuan"
)

// TokenDisposition represents the token's disposition/attitude
// Used for Foundry VTT compatibility
type TokenDisposition string

const (
	// DispositionFriendly Friendly token (green border)
	DispositionFriendly TokenDisposition = "friendly"
	// DispositionNeutral Neutral token (yellow border)
	DispositionNeutral TokenDisposition = "neutral"
	// DispositionHostile Hostile token (red border)
	DispositionHostile TokenDisposition = "hostile"
	// DispositionSecret Secret disposition (hidden until revealed)
	DispositionSecret TokenDisposition = "secret"
)

// 最大地图尺寸限制
const (
	MaxMapWidth  = 200
	MaxMapHeight = 200
)

// GetTokenSizeInFeet 获取Token大小对应的英尺数
// 规则参考: PHB 第9章 - Size and Space
func GetTokenSizeInFeet(size TokenSize) int {
	switch size {
	case TokenSizeTiny:
		return 2
	case TokenSizeSmall, TokenSizeMedium:
		return 5
	case TokenSizeLarge:
		return 10
	case TokenSizeHuge:
		return 15
	case TokenSizeGargantuan:
		return 20
	default:
		return 5 // 默认为中型
	}
}

// GetTokenSizeInGrids 获取Token大小占用的格子数
// 规则参考: PHB 第9章 - Size and Space (每格5英尺)
func GetTokenSizeInGrids(size TokenSize) int {
	feet := GetTokenSizeInFeet(size)
	return feet / 5
}

// Grid 格子系统
type Grid struct {
	Width    int          `json:"width"`      // 宽度（格子数）
	Height   int          `json:"height"`     // 高度（格子数）
	CellSize int          `json:"cell_size"`  // 每格大小（游戏单位，如5英尺）
	Cells    [][]CellType `json:"cells"`      // 格子内容
}

// NewGrid 创建新格子
func NewGrid(width, height, cellSize int) *Grid {
	cells := make([][]CellType, height)
	for i := range cells {
		cells[i] = make([]CellType, width)
		for j := range cells[i] {
			cells[i][j] = CellTypeEmpty
		}
	}
	return &Grid{
		Width:    width,
		Height:   height,
		CellSize: cellSize,
		Cells:    cells,
	}
}

// Validate 验证格子
func (g *Grid) Validate() error {
	if g.Width <= 0 || g.Width > MaxMapWidth {
		return NewValidationError("grid.width", "must be between 1 and "+string(rune('0'+MaxMapWidth)))
	}
	if g.Height <= 0 || g.Height > MaxMapHeight {
		return NewValidationError("grid.height", "must be between 1 and "+string(rune('0'+MaxMapHeight)))
	}
	if g.CellSize <= 0 {
		return NewValidationError("grid.cell_size", "must be positive")
	}
	if len(g.Cells) != g.Height {
		return NewValidationError("grid.cells", "height mismatch")
	}
	for i, row := range g.Cells {
		if len(row) != g.Width {
			return NewValidationError("grid.cells", "width mismatch at row "+string(rune('0'+i)))
		}
	}
	return nil
}

// GetCell 获取指定位置的格子类型
func (g *Grid) GetCell(x, y int) CellType {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return CellTypeEmpty
	}
	return g.Cells[y][x]
}

// SetCell 设置指定位置的格子类型
func (g *Grid) SetCell(x, y int, cellType CellType) bool {
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return false
	}
	g.Cells[y][x] = cellType
	return true
}

// IsWalkable 检查指定位置是否可通行
func (g *Grid) IsWalkable(x, y int) bool {
	cell := g.GetCell(x, y)
	return cell != CellTypeWall
}

// IsDifficultTerrain 检查指定位置是否为困难地形
// 规则参考: PHB 第8章 - Difficult Terrain
func (g *Grid) IsDifficultTerrain(x, y int) bool {
	cell := g.GetCell(x, y)
	return cell == CellTypeDifficult || cell == CellTypeWater || cell == CellTypeForest || cell == CellTypeMountain
}

// Location 地点标记（用于大地图）
type Location struct {
	ID          string   `json:"id"`           // UUID
	Name        string   `json:"name"`         // 地点名称
	Description string   `json:"description"`  // 地点描述
	Position    Position `json:"position"`     // 地图位置
	BattleMapID string   `json:"battle_map_id"`// 关联的战斗地图ID
}

// NewLocation 创建新地点
func NewLocation(name, description string, x, y int) *Location {
	return &Location{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Position:    Position{X: x, Y: y},
	}
}

// Validate 验证地点
func (l *Location) Validate() error {
	if l.Name == "" {
		return NewValidationError("location.name", "cannot be empty")
	}
	if err := l.Position.Validate(); err != nil {
		return err
	}
	return nil
}

// Token 地图上的标记（用于战斗地图）
type Token struct {
	ID          string   `json:"id"`           // UUID
	CharacterID string   `json:"character_id"` // 关联的角色ID
	Position    Position `json:"position"`     // 位置
	Size        TokenSize `json:"size"`        // 大小
	Visible     bool     `json:"visible"`      // 是否可见

	// Extended fields for FVTT compatibility
	Name         string       `json:"name,omitempty"`         // Token显示名称
	Image        *MapImage    `json:"image,omitempty"`        // Token图片
	Width        int          `json:"width,omitempty"`        // 显示宽度（像素）
	Height       int          `json:"height,omitempty"`       // 显示高度（像素）
	Rotation     float64      `json:"rotation,omitempty"`     // 旋转角度（度）
	Scale        float64      `json:"scale,omitempty"`        // 缩放比例
	Alpha        float64      `json:"alpha,omitempty"`        // 透明度（0-1）
	ActorLink    string       `json:"actor_link,omitempty"`  // 关联Actor ID
	Disposition  TokenDisposition `json:"disposition,omitempty"` // 态度（敌对/友好/中立）
	Hidden       bool         `json:"hidden,omitempty"`      // 对特定用户隐藏
	Locked       bool         `json:"locked,omitempty"`      // 锁定位置
	Bar1         *TokenBar    `json:"bar1,omitempty"`        // 主属性条（HP）
	Bar2         *TokenBar    `json:"bar2,omitempty"`        // 次属性条
}

// NewToken 创建新Token
func NewToken(characterID string, x, y int, size TokenSize) *Token {
	return &Token{
		ID:          uuid.New().String(),
		CharacterID: characterID,
		Position:    Position{X: x, Y: y},
		Size:        size,
		Visible:     true,
		// Extended fields with defaults
		Name:        "",
		Rotation:    0,
		Scale:       1.0,
		Alpha:       1.0,
		Disposition: DispositionNeutral,
		Hidden:      false,
		Locked:      false,
	}
}

// Validate 验证Token
func (t *Token) Validate() error {
	if t.CharacterID == "" {
		return NewValidationError("token.character_id", "cannot be empty")
	}
	if err := t.Position.Validate(); err != nil {
		return err
	}

	// Validate extended fields
	if t.Width < 0 {
		return NewValidationError("token.width", "cannot be negative")
	}
	if t.Height < 0 {
		return NewValidationError("token.height", "cannot be negative")
	}
	if t.Scale <= 0 {
		return NewValidationError("token.scale", "must be positive")
	}
	if t.Alpha < 0 || t.Alpha > 1 {
		return NewValidationError("token.alpha", "must be between 0 and 1")
	}
	if t.Rotation < -360 || t.Rotation > 360 {
		return NewValidationError("token.rotation", "must be between -360 and 360 degrees")
	}

	// Validate bars
	if t.Bar1 != nil {
		if err := t.Bar1.Validate(); err != nil {
			return NewValidationError("token.bar1", err.Error())
		}
	}
	if t.Bar2 != nil {
		if err := t.Bar2.Validate(); err != nil {
			return NewValidationError("token.bar2", err.Error())
		}
	}

	// Validate image if present
	if t.Image != nil {
		if err := t.Image.Validate(); err != nil {
			return NewValidationError("token.image", err.Error())
		}
	}

	return nil
}

// GetSizeInFeet 获取Token大小对应的英尺数
func (t *Token) GetSizeInFeet() int {
	return GetTokenSizeInFeet(t.Size)
}

// GetSizeInGrids 获取Token大小占用的格子数
func (t *Token) GetSizeInGrids() int {
	return GetTokenSizeInGrids(t.Size)
}

// SetPosition 设置Token位置
func (t *Token) SetPosition(x, y int) {
	t.Position = Position{X: x, Y: y}
}

// Map 地图实体
type Map struct {
	ID         string     `json:"id"`          // UUID
	CampaignID string     `json:"campaign_id"` // 所属战役ID
	Name       string     `json:"name"`        // 地图名称
	Type       MapType    `json:"type"`        // 地图类型
	Mode       MapMode    `json:"mode"`        // 地图模式（grid/image）
	Grid       *Grid      `json:"grid"`        // 格子系统
	Locations  []Location `json:"locations"`   // 地点标记（大地图 Grid 模式）
	Tokens     []Token    `json:"tokens"`      // Token列表（战斗地图）
	ParentID   string     `json:"parent_id"`   // 父地图ID（战斗地图关联的地点）
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Extended fields for FVTT compatibility
	Image           *MapImage        `json:"image,omitempty"`             // 地图背景图片
	Walls           Walls            `json:"walls,omitempty"`             // 墙壁列表
	ImportMeta      *MapImportMeta   `json:"import_meta,omitempty"`       // 导入元数据
	VisualLocations []VisualLocation `json:"visual_locations,omitempty"`  // 视觉识别的地点（Image 模式）
}

// NewMap 创建新地图
func NewMap(campaignID, name string, mapType MapType, width, height, cellSize int) *Map {
	now := time.Now()
	m := &Map{
		CampaignID: campaignID,
		Name:       name,
		Type:       mapType,
		Mode:       MapModeGrid, // 默认为 Grid 模式
		Grid:       NewGrid(width, height, cellSize),
		Locations:  make([]Location, 0),
		Tokens:     make([]Token, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	m.ID = uuid.New().String()
	return m
}

// NewWorldMap 创建大地图
func NewWorldMap(campaignID, name string, width, height int) *Map {
	return NewMap(campaignID, name, MapTypeWorld, width, height, 1) // 大地图每格为单位区域
}

// NewBattleMap 创建战斗地图
func NewBattleMap(campaignID, name string, width, height, cellSize int) *Map {
	return NewMap(campaignID, name, MapTypeBattle, width, height, cellSize)
}

// Validate 验证地图数据
func (m *Map) Validate() error {
	if m.Name == "" {
		return NewValidationError("name", "cannot be empty")
	}
	if m.CampaignID == "" {
		return NewValidationError("campaign_id", "cannot be empty")
	}
	if m.Grid != nil {
		if err := m.Grid.Validate(); err != nil {
			return err
		}
	}
	// 验证地点
	for i, loc := range m.Locations {
		if err := loc.Validate(); err != nil {
			return NewValidationError("locations", "invalid location at index "+string(rune('0'+i))+": "+err.Error())
		}
		// 检查地点位置是否在格子范围内
		if m.Grid != nil {
			if loc.Position.X < 0 || loc.Position.X >= m.Grid.Width ||
				loc.Position.Y < 0 || loc.Position.Y >= m.Grid.Height {
				return NewValidationError("locations", "location at index "+string(rune('0'+i))+" is out of bounds")
			}
		}
	}
	// 验证Token
	for i, token := range m.Tokens {
		if err := token.Validate(); err != nil {
			return NewValidationError("tokens", "invalid token at index "+string(rune('0'+i))+": "+err.Error())
		}
		// 检查Token位置是否在格子范围内
		if m.Grid != nil {
			sizeInGrids := token.GetSizeInGrids()
			if token.Position.X < 0 || token.Position.X+sizeInGrids > m.Grid.Width ||
				token.Position.Y < 0 || token.Position.Y+sizeInGrids > m.Grid.Height {
				return NewValidationError("tokens", "token at index "+string(rune('0'+i))+" is out of bounds")
			}
		}
	}
	return nil
}

// IsWorldMap 检查是否为大地图
func (m *Map) IsWorldMap() bool {
	return m.Type == MapTypeWorld
}

// IsBattleMap 检查是否为战斗地图
func (m *Map) IsBattleMap() bool {
	return m.Type == MapTypeBattle
}

// AddLocation 添加地点
func (m *Map) AddLocation(location Location) error {
	if err := location.Validate(); err != nil {
		return err
	}
	m.Locations = append(m.Locations, location)
	m.UpdatedAt = time.Now()
	return nil
}

// RemoveLocation 移除地点
func (m *Map) RemoveLocation(locationID string) bool {
	for i, loc := range m.Locations {
		if loc.ID == locationID {
			m.Locations = append(m.Locations[:i], m.Locations[i+1:]...)
			m.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetLocation 获取地点
func (m *Map) GetLocation(locationID string) *Location {
	for i := range m.Locations {
		if m.Locations[i].ID == locationID {
			return &m.Locations[i]
		}
	}
	return nil
}

// GetLocationAtPosition 获取指定位置的地点
func (m *Map) GetLocationAtPosition(x, y int) *Location {
	for i := range m.Locations {
		if m.Locations[i].Position.X == x && m.Locations[i].Position.Y == y {
			return &m.Locations[i]
		}
	}
	return nil
}

// AddToken 添加Token
func (m *Map) AddToken(token Token) error {
	if err := token.Validate(); err != nil {
		return err
	}
	m.Tokens = append(m.Tokens, token)
	m.UpdatedAt = time.Now()
	return nil
}

// RemoveToken 移除Token
func (m *Map) RemoveToken(tokenID string) bool {
	for i, token := range m.Tokens {
		if token.ID == tokenID {
			m.Tokens = append(m.Tokens[:i], m.Tokens[i+1:]...)
			m.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetToken 获取Token
func (m *Map) GetToken(tokenID string) *Token {
	for i := range m.Tokens {
		if m.Tokens[i].ID == tokenID {
			return &m.Tokens[i]
		}
	}
	return nil
}

// GetTokenByCharacterID 获取角色的Token
func (m *Map) GetTokenByCharacterID(characterID string) *Token {
	for i := range m.Tokens {
		if m.Tokens[i].CharacterID == characterID {
			return &m.Tokens[i]
		}
	}
	return nil
}

// GetTokensAtPosition 获取指定位置的所有Token
func (m *Map) GetTokensAtPosition(x, y int) []Token {
	tokens := make([]Token, 0)
	for i := range m.Tokens {
		token := &m.Tokens[i]
		size := token.GetSizeInGrids()
		// 检查Token是否覆盖该位置
		if x >= token.Position.X && x < token.Position.X+size &&
			y >= token.Position.Y && y < token.Position.Y+size {
			tokens = append(tokens, *token)
		}
	}
	return tokens
}

// AddVisualLocation 添加视觉识别的地点
func (m *Map) AddVisualLocation(location VisualLocation) error {
	if err := location.Validate(); err != nil {
		return err
	}
	m.VisualLocations = append(m.VisualLocations, location)
	m.UpdatedAt = time.Now()
	return nil
}

// RemoveVisualLocation 移除视觉地点
func (m *Map) RemoveVisualLocation(locationID string) bool {
	for i, loc := range m.VisualLocations {
		if loc.ID == locationID {
			m.VisualLocations = append(m.VisualLocations[:i], m.VisualLocations[i+1:]...)
			m.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetVisualLocation 获取视觉地点
func (m *Map) GetVisualLocation(locationID string) *VisualLocation {
	for i := range m.VisualLocations {
		if m.VisualLocations[i].ID == locationID {
			return &m.VisualLocations[i]
		}
	}
	return nil
}

// GetVisualLocationAtPosition 获取指定位置的视觉地点（使用容差）
func (m *Map) GetVisualLocationAtPosition(x, y float64, tolerance float64) *VisualLocation {
	for i := range m.VisualLocations {
		loc := &m.VisualLocations[i]
		dx := loc.PositionX - x
		dy := loc.PositionY - y
		if dx*dx+dy*dy <= tolerance*tolerance {
			return loc
		}
	}
	return nil
}

// GetConfirmedVisualLocations 获取已确认的视觉地点
func (m *Map) GetConfirmedVisualLocations() []VisualLocation {
	result := make([]VisualLocation, 0)
	for _, loc := range m.VisualLocations {
		if loc.IsConfirmed {
			result = append(result, loc)
		}
	}
	return result
}

// GetUnconfirmedVisualLocations 获取未确认的视觉地点
func (m *Map) GetUnconfirmedVisualLocations() []VisualLocation {
	result := make([]VisualLocation, 0)
	for _, loc := range m.VisualLocations {
		if !loc.IsConfirmed {
			result = append(result, loc)
		}
	}
	return result
}

// MapMode 地图模式
type MapMode string

const (
	// MapModeGrid 格子模式（默认）
	MapModeGrid MapMode = "grid"
	// MapModeImage 图片模式
	MapModeImage MapMode = "image"
)

// VisualLocation 视觉识别的地点
// 用于 Image 模式的大地图，使用归一化坐标
type VisualLocation struct {
	ID          string  `json:"id"`                    // UUID
	Name        string  `json:"name"`                  // 视觉识别的名称
	Description string  `json:"description"`           // AI 生成的描述
	Type        string  `json:"type"`                  // town, dungeon, forest, mountain, etc.
	PositionX   float64 `json:"position_x"`            // X 坐标（0-1 归一化）
	PositionY   float64 `json:"position_y"`            // Y 坐标（0-1 归一化）
	BattleMapID string  `json:"battle_map_id,omitempty"` // 关联的战斗地图
	CustomName  string  `json:"custom_name,omitempty"`   // DM 自定义名称
	IsConfirmed bool    `json:"is_confirmed"`            // DM 是否已确认
}

// NewVisualLocation 创建新的视觉地点
func NewVisualLocation(name, description, locationType string, posX, posY float64) *VisualLocation {
	return &VisualLocation{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Type:        locationType,
		PositionX:   posX,
		PositionY:   posY,
		IsConfirmed: false,
	}
}

// Validate 验证视觉地点
func (v *VisualLocation) Validate() error {
	if v.Name == "" && v.CustomName == "" {
		return NewValidationError("visual_location.name", "name or custom_name must be provided")
	}
	if v.PositionX < 0 || v.PositionX > 1 {
		return NewValidationError("visual_location.position_x", "must be between 0 and 1")
	}
	if v.PositionY < 0 || v.PositionY > 1 {
		return NewValidationError("visual_location.position_y", "must be between 0 and 1")
	}
	return nil
}

// GetDisplayName 获取显示名称（优先使用自定义名称）
func (v *VisualLocation) GetDisplayName() string {
	if v.CustomName != "" {
		return v.CustomName
	}
	return v.Name
}

// Confirm 确认地点
func (v *VisualLocation) Confirm() {
	v.IsConfirmed = true
}

// SetCustomName 设置自定义名称
func (v *VisualLocation) SetCustomName(name string) {
	v.CustomName = name
}

// SetBattleMapID 设置关联的战斗地图
func (v *VisualLocation) SetBattleMapID(mapID string) {
	v.BattleMapID = mapID
}

// TerrainArea 地形区域
type TerrainArea struct {
	Type        string `json:"type"`        // forest, desert, water, etc.
	Description string `json:"description"` // 描述
	Bounds      Rect   `json:"bounds"`      // 区域边界
}

// Validate 验证地形区域
func (t *TerrainArea) Validate() error {
	if t.Type == "" {
		return NewValidationError("terrain_area.type", "cannot be empty")
	}
	return t.Bounds.Validate()
}

// Landmark 标志物
type Landmark struct {
	Name        string  `json:"name"`        // 标志物名称
	Description string  `json:"description"` // 描述
	PositionX   float64 `json:"position_x"`  // X 坐标（0-1 归一化）
	PositionY   float64 `json:"position_y"`  // Y 坐标（0-1 归一化）
}

// Validate 验证标志物
func (l *Landmark) Validate() error {
	if l.Name == "" {
		return NewValidationError("landmark.name", "cannot be empty")
	}
	if l.PositionX < 0 || l.PositionX > 1 {
		return NewValidationError("landmark.position_x", "must be between 0 and 1")
	}
	if l.PositionY < 0 || l.PositionY > 1 {
		return NewValidationError("landmark.position_y", "must be between 0 and 1")
	}
	return nil
}

// Rect 矩形区域
type Rect struct {
	X      float64 `json:"x"`      // 左上角 X 坐标（0-1 归一化）
	Y      float64 `json:"y"`      // 左上角 Y 坐标（0-1 归一化）
	Width  float64 `json:"width"`  // 宽度（0-1 归一化）
	Height float64 `json:"height"` // 高度（0-1 归一化）
}

// Validate 验证矩形区域
func (r *Rect) Validate() error {
	if r.X < 0 || r.X > 1 {
		return NewValidationError("rect.x", "must be between 0 and 1")
	}
	if r.Y < 0 || r.Y > 1 {
		return NewValidationError("rect.y", "must be between 0 and 1")
	}
	if r.Width <= 0 || r.Width > 1 {
		return NewValidationError("rect.width", "must be between 0 and 1")
	}
	if r.Height <= 0 || r.Height > 1 {
		return NewValidationError("rect.height", "must be between 0 and 1")
	}
	return nil
}

// Contains 检查点是否在矩形内
func (r *Rect) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.Width && y >= r.Y && y <= r.Y+r.Height
}

// VisualAnalysis 视觉分析结果
type VisualAnalysis struct {
	Summary   string            `json:"summary"`   // 地图整体描述
	Locations []VisualLocation  `json:"locations"` // 识别的地点
	Terrains  []TerrainArea     `json:"terrains"`  // 识别的地形区域
	Landmarks []Landmark        `json:"landmarks"` // 标志性建筑/物体
}

// Validate 验证视觉分析结果
func (v *VisualAnalysis) Validate() error {
	for i, loc := range v.Locations {
		if err := loc.Validate(); err != nil {
			return NewValidationError("visual_analysis.locations", fmt.Sprintf("invalid location at index %d: %v", i, err))
		}
	}
	for i, terrain := range v.Terrains {
		if err := terrain.Validate(); err != nil {
			return NewValidationError("visual_analysis.terrains", fmt.Sprintf("invalid terrain at index %d: %v", i, err))
		}
	}
	for i, landmark := range v.Landmarks {
		if err := landmark.Validate(); err != nil {
			return NewValidationError("visual_analysis.landmarks", fmt.Sprintf("invalid landmark at index %d: %v", i, err))
		}
	}
	return nil
}
