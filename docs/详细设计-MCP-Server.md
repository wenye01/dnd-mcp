# MCP Server 详细设计书

## 文档信息

- **版本**: v1.0
- **创建日期**: 2025-02-16
- **基于**: 高层次设计 v1.1
- **状态**: 细化中

---

## 第1轮: 数据结构

> 定义所有核心实体和字段

### 1.1 战役 (Campaign)

```go
// Campaign 战役实体
type Campaign struct {
    ID          string                 `json:"id"`           // UUID
    Name        string                 `json:"name"`         // 战役名称
    Description string                 `json:"description"`  // 战役描述
    DMID        string                 `json:"dm_id"`        // DM（地下城主）用户ID
    Settings    *CampaignSettings      `json:"settings"`     // 战役设置
    Status      CampaignStatus         `json:"status"`       // 状态
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// CampaignStatus 战役状态
type CampaignStatus string

const (
    CampaignStatusActive   CampaignStatus = "active"    // 进行中
    CampaignStatusPaused   CampaignStatus = "paused"    // 暂停
    CampaignStatusFinished CampaignStatus = "finished"  // 已结束
    CampaignStatusArchived CampaignStatus = "archived"  // 已归档
)

// CampaignSettings 战役设置
type CampaignSettings struct {
    MaxPlayers    int                    `json:"max_players"`     // 最大玩家数，默认4
    StartLevel    int                    `json:"start_level"`     // 起始等级，默认1
    Ruleset       string                 `json:"ruleset"`         // 规则集，默认 "dnd5e"
    HouseRules    map[string]interface{} `json:"house_rules"`     // 房规
    ContextWindow int                    `json:"context_window"`  // 上下文窗口大小，默认20
}
```

### 1.2 角色 (Character)

```go
// Character 角色实体（玩家角色和NPC共用）
type Character struct {
    ID          string            `json:"id"`           // UUID
    CampaignID  string            `json:"campaign_id"`  // 所属战役ID
    Name        string            `json:"name"`         // 角色名称
    IsNPC       bool              `json:"is_npc"`       // 是否为NPC
    NPCType     NPCType           `json:"npc_type"`     // NPC类型（仅NPC使用）
    PlayerID    string            `json:"player_id"`    // 玩家ID（玩家角色使用）

    // 基础属性
    Race        string            `json:"race"`         // 种族
    Class       string            `json:"class"`        // 职业
    Level       int               `json:"level"`        // 等级
    Background  string            `json:"background"`   // 背景
    Alignment   string            `json:"alignment"`    // 阵营

    // 属性值
    Abilities   *Abilities        `json:"abilities"`    // 六大属性

    // 战斗属性
    HP          *HP               `json:"hp"`           // 生命值
    AC          int               `json:"ac"`           // 护甲等级
    Speed       int               `json:"speed"`        // 移动速度（英尺）
    Initiative  int               `json:"initiative"`   // 先攻加值

    // 技能和特长
    Skills      map[string]int    `json:"skills"`       // 技能加值
    Saves       map[string]int    `json:"saves"`        // 豁免加值

    // 装备和物品
    Equipment   []Equipment       `json:"equipment"`    // 装备列表
    Inventory   []Item            `json:"inventory"`    // 背包物品

    // 状态
    Conditions  []Condition       `json:"conditions"`   // 状态效果

    // 元数据
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// NPCType NPC类型
type NPCType string

const (
    NPCTypeScripted  NPCType = "scripted"   // 剧本固定NPC
    NPCTypeGenerated NPCType = "generated"  // LLM即兴创建NPC
)

// Abilities 六大属性
type Abilities struct {
    Strength     int `json:"strength"`      // 力量
    Dexterity    int `json:"dexterity"`     // 敏捷
    Constitution int `json:"constitution"`  // 体质
    Intelligence int `json:"intelligence"`  // 智力
    Wisdom       int `json:"wisdom"`        // 感知
    Charisma     int `json:"charisma"`      // 魅力
}

// HP 生命值
type HP struct {
    Current int `json:"current"`  // 当前HP
    Max     int `json:"max"`      // 最大HP
    Temp    int `json:"temp"`     // 临时HP
}

// Equipment 装备
type Equipment struct {
    ID       string `json:"id"`        // 物品ID
    Name     string `json:"name"`      // 物品名称
    Slot     string `json:"slot"`      // 装备槽位（main_hand, off_hand, armor, etc.）
    Bonus    int    `json:"bonus"`     // 加值
    Damage   string `json:"damage"`    // 伤害骰（武器）
    DamageType string `json:"damage_type"` // 伤害类型
}

// Item 物品
type Item struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Quantity int    `json:"quantity"`
}

// Condition 状态效果
type Condition struct {
    Type     string     `json:"type"`      // 状态类型（poisoned, paralyzed, etc.）
    Duration int        `json:"duration"`  // 持续回合数，-1表示永久
    Source   string     `json:"source"`    // 来源
}
```

### 1.3 战斗 (Combat)

```go
// Combat 战斗实体
type Combat struct {
    ID           string           `json:"id"`            // UUID
    CampaignID   string           `json:"campaign_id"`   // 所属战役ID
    Status       CombatStatus     `json:"status"`        // 战斗状态
    Round        int              `json:"round"`         // 当前回合数
    TurnIndex    int              `json:"turn_index"`    // 当前行动者索引
    Participants []Participant    `json:"participants"`  // 参战者列表
    MapID        string           `json:"map_id"`        // 战斗地图ID（可选）
    Log          []CombatLogEntry `json:"log"`           // 战斗日志
    StartedAt    time.Time        `json:"started_at"`
    EndedAt      *time.Time       `json:"ended_at,omitempty"`
}

// CombatStatus 战斗状态
type CombatStatus string

const (
    CombatStatusActive   CombatStatus = "active"    // 进行中
    CombatStatusFinished CombatStatus = "finished"  // 已结束
)

// Participant 参战者
type Participant struct {
    CharacterID  string          `json:"character_id"`  // 角色ID
    Initiative   int             `json:"initiative"`    // 先攻值
    HasActed     bool            `json:"has_acted"`     // 本回合是否已行动
    Position     *Position       `json:"position"`      // 战斗地图位置
    TempHP       int             `json:"temp_hp"`       // 临时HP（战斗中）
    Conditions   []Condition     `json:"conditions"`    // 战斗中的临时状态
}

// Position 位置
type Position struct {
    X int `json:"x"`  // 格子X坐标
    Y int `json:"y"`  // 格子Y坐标
}

// CombatLogEntry 战斗日志条目
type CombatLogEntry struct {
    Round     int       `json:"round"`       // 回合数
    ActorID   string    `json:"actor_id"`    // 行动者ID
    Action    string    `json:"action"`      // 动作类型（attack, spell, move, etc.）
    TargetID  string    `json:"target_id"`   // 目标ID（可选）
    Result    string    `json:"result"`      // 结果描述
    Timestamp time.Time `json:"timestamp"`
}
```

### 1.4 地图 (Map)

```go
// Map 地图实体
type Map struct {
    ID          string       `json:"id"`           // UUID
    CampaignID  string       `json:"campaign_id"`  // 所属战役ID
    Name        string       `json:"name"`         // 地图名称
    Type        MapType      `json:"type"`         // 地图类型
    Grid        *Grid        `json:"grid"`         // 格子系统
    Locations   []Location   `json:"locations"`    // 地点标记（大地图）
    Tokens      []Token      `json:"tokens"`       // Token列表（战斗地图）
    ParentID    string       `json:"parent_id"`    // 父地图ID（战斗地图关联的地点）
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

// MapType 地图类型
type MapType string

const (
    MapTypeWorld  MapType = "world"   // 大地图
    MapTypeBattle MapType = "battle"  // 战斗地图
)

// Grid 格子系统
type Grid struct {
    Width    int          `json:"width"`      // 宽度（格子数）
    Height   int          `json:"height"`     // 高度（格子数）
    CellSize int          `json:"cell_size"`  // 每格大小（游戏单位，如5英尺）
    Cells    [][]CellType `json:"cells"`      // 格子内容
}

// CellType 格子类型
type CellType string

const (
    CellTypeEmpty           CellType = "empty"             // 空地
    CellTypeWall            CellType = "wall"              // 墙壁
    CellTypeDifficult       CellType = "difficult"         // 困难地形
    CellTypeWater           CellType = "water"             // 水域
    CellTypeDoor            CellType = "door"              // 门
    // 大地图专用
    CellTypeRoad            CellType = "road"              // 道路
    CellTypeForest          CellType = "forest"            // 森林
    CellTypeMountain        CellType = "mountain"          // 山地
    CellTypeBuilding        CellType = "building"          // 建筑
)

// Location 地点标记
type Location struct {
    ID          string   `json:"id"`           // UUID
    Name        string   `json:"name"`         // 地点名称
    Description string   `json:"description"`  // 地点描述
    Position    Position `json:"position"`     // 地图位置
    BattleMapID string   `json:"battle_map_id"`// 关联的战斗地图ID
}

// Token 地图上的标记
type Token struct {
    ID           string   `json:"id"`            // UUID
    CharacterID  string   `json:"character_id"`  // 关联的角色ID
    Position     Position `json:"position"`      // 位置
    Size         TokenSize `json:"size"`         // 大小
    Visible      bool     `json:"visible"`       // 是否可见
}

// TokenSize Token大小
type TokenSize string

const (
    TokenSizeTiny   TokenSize = "tiny"    // 微型 (2.5x2.5英尺)
    TokenSizeSmall  TokenSize = "small"   // 小型 (5x5英尺)
    TokenSizeMedium TokenSize = "medium"  // 中型 (5x5英尺)
    TokenSizeLarge  TokenSize = "large"   // 大型 (10x10英尺)
    TokenSizeHuge   TokenSize = "huge"    // 超大型 (15x15英尺)
    TokenSizeGargantuan TokenSize = "gargantuan" // 巨型 (20x20英尺或更大)
)
```

### 1.5 游戏状态 (GameState)

```go
// GameState 游戏状态
type GameState struct {
    ID              string    `json:"id"`              // 与CampaignID相同
    CampaignID      string    `json:"campaign_id"`     // 所属战役ID
    GameTime        *GameTime `json:"game_time"`       // 游戏时间
    PartyPosition   *Position `json:"party_position"`  // 队伍在大地图的位置
    CurrentMapID    string    `json:"current_map_id"`  // 当前所在地图ID
    CurrentMapType  MapType   `json:"current_map_type"`// 当前地图类型
    Weather         string    `json:"weather"`         // 天气
    ActiveCombatID  string    `json:"active_combat_id"`// 当前战斗ID（如果有）
    UpdatedAt       time.Time `json:"updated_at"`
}

// GameTime 游戏时间
type GameTime struct {
    Year    int    `json:"year"`     // 年
    Month   int    `json:"month"`    // 月
    Day     int    `json:"day"`      // 日
    Hour    int    `json:"hour"`     // 时
    Minute  int    `json:"minute"`   // 分
    Phase   string `json:"phase"`    // 时段（dawn, morning, noon, afternoon, dusk, night）
}
```

### 1.6 消息 (Message)

```go
// Message 对话消息
type Message struct {
    ID          string     `json:"id"`           // UUID
    CampaignID  string     `json:"campaign_id"`  // 所属战役ID
    Role        MessageRole `json:"role"`        // 角色（user, assistant, system）
    Content     string     `json:"content"`      // 消息内容
    PlayerID    string     `json:"player_id"`    // 玩家ID（user消息）
    ToolCalls   []ToolCall `json:"tool_calls"`   // 工具调用（assistant消息）
    CreatedAt   time.Time  `json:"created_at"`
}

// MessageRole 消息角色
type MessageRole string

const (
    MessageRoleUser      MessageRole = "user"
    MessageRoleAssistant MessageRole = "assistant"
    MessageRoleSystem    MessageRole = "system"
)

// ToolCall 工具调用
type ToolCall struct {
    ID        string                 `json:"id"`         // 调用ID
    Name      string                 `json:"name"`       // 工具名称
    Arguments map[string]interface{} `json:"arguments"`  // 参数
    Result    *ToolResult            `json:"result"`     // 执行结果
}

// ToolResult 工具执行结果
type ToolResult struct {
    Success bool                   `json:"success"`  // 是否成功
    Data    map[string]interface{} `json:"data"`     // 返回数据
    Error   string                 `json:"error"`    // 错误信息
}
```

### 1.7 骰子结果 (DiceResult)

```go
// DiceResult 骰子投掷结果
type DiceResult struct {
    Formula    string       `json:"formula"`     // 骰子公式（如 "1d20+5"）
    Rolls      []int        `json:"rolls"`       // 原始投掷值
    Modifier   int          `json:"modifier"`    // 修正值
    Total      int          `json:"total"`       // 总计
    CritStatus CritStatus   `json:"crit_status"` // 暴击状态
}

// CritStatus 暴击状态
type CritStatus string

const (
    CritStatusNone    CritStatus = "none"     // 普通
    CritStatusSuccess CritStatus = "critical" // 暴击（自然20）
    CritStatusFail    CritStatus = "fumble"   // 大失败（自然1）
)

// CheckResult 检定结果
type CheckResult struct {
    DiceResult  *DiceResult `json:"dice_result"`  // 骰子结果
    Ability     string      `json:"ability"`      // 检定属性
    Skill       string      `json:"skill"`        // 技能（可选）
    DC          int         `json:"dc"`           // 难度等级
    Success     bool        `json:"success"`      // 是否成功
    Margin      int         `json:"margin"`       // 成功/失败幅度
}

// AttackResult 攻击结果
type AttackResult struct {
    AttackRoll   *DiceResult `json:"attack_roll"`   // 攻击骰
    TargetAC     int         `json:"target_ac"`     // 目标AC
    Hit          bool        `json:"hit"`           // 是否命中
    Crit         bool        `json:"crit"`          // 是否暴击
    DamageRoll   *DiceResult `json:"damage_roll"`   // 伤害骰
    Damage       int         `json:"damage"`        // 总伤害
    DamageType   string      `json:"damage_type"`   // 伤害类型
    TargetHP     *HP         `json:"target_hp"`     // 目标剩余HP
    TargetDown   bool        `json:"target_down"`   // 目标是否倒地
}
```

### 1.8 上下文 (Context)

```go
// Context LLM上下文
type Context struct {
    CampaignID      string          `json:"campaign_id"`
    GameSummary     *GameSummary    `json:"game_summary"`     // 游戏状态摘要
    Messages        []Message       `json:"messages"`         // 对话历史（压缩后）
    RawMessageCount int             `json:"raw_message_count"`// 原始消息总数
    TokenEstimate   int             `json:"token_estimate"`   // 预估token数
}

// GameSummary 游戏状态摘要
type GameSummary struct {
    CampaignName    string        `json:"campaign_name"`    // 战役名称
    GameTime        *GameTime     `json:"game_time"`        // 游戏时间
    CurrentLocation string        `json:"current_location"` // 当前位置
    PartyMembers    []PartyMember `json:"party_members"`    // 队伍成员
    ActiveCombat    *CombatSummary `json:"active_combat"`   // 当前战斗（如果有）
    RecentEvents    []string      `json:"recent_events"`    // 最近事件
}

// PartyMember 队伍成员摘要
type PartyMember struct {
    CharacterID string `json:"character_id"`
    Name        string `json:"name"`
    Class       string `json:"class"`
    Level       int    `json:"level"`
    HP          *HP    `json:"hp"`
    Conditions  []string `json:"conditions"`
}

// CombatSummary 战斗摘要
type CombatSummary struct {
    Round        int      `json:"round"`
    CurrentTurn  string   `json:"current_turn"`  // 当前行动角色名
    TurnOrder    []string `json:"turn_order"`    // 回合顺序（角色名）
    Enemies      []string `json:"enemies"`       // 敌人列表
}
```

---

## 第2轮: 接口定义

> 定义所有 MCP Tools 的请求/响应

### 2.1 战役管理 Tools

#### 2.1.1 create_campaign

创建新战役。

**请求参数**:
```go
type CreateCampaignRequest struct {
    Name        string                 `json:"name"`         // 必填，战役名称
    Description string                 `json:"description"`  // 可选，战役描述
    DMID        string                 `json:"dm_id"`        // 必填，DM用户ID
    Settings    *CampaignSettings      `json:"settings"`     // 可选，战役设置
}
```

**响应**:
```go
type CreateCampaignResponse struct {
    Campaign *Campaign `json:"campaign"` // 创建的战役实体
}
```

**Tool 定义**:
```json
{
  "name": "create_campaign",
  "description": "创建一个新的D&D战役",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name": {"type": "string", "description": "战役名称"},
      "description": {"type": "string", "description": "战役描述"},
      "dm_id": {"type": "string", "description": "DM用户ID"},
      "settings": {"type": "object", "description": "战役设置"}
    },
    "required": ["name", "dm_id"]
  }
}
```

#### 2.1.2 get_campaign

获取战役详情。

**请求参数**:
```go
type GetCampaignRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetCampaignResponse struct {
    Campaign *Campaign `json:"campaign"` // 战役实体
}
```

#### 2.1.3 list_campaigns

列出所有战役。

**请求参数**:
```go
type ListCampaignsRequest struct {
    Status CampaignStatus `json:"status"` // 可选，按状态筛选
    DMID   string         `json:"dm_id"`  // 可选，按DM筛选
}
```

**响应**:
```go
type ListCampaignsResponse struct {
    Campaigns []*Campaign `json:"campaigns"` // 战役列表
}
```

#### 2.1.4 delete_campaign

删除战役（软删除）。

**请求参数**:
```go
type DeleteCampaignRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type DeleteCampaignResponse struct {
    Success bool `json:"success"` // 是否成功
}
```

#### 2.1.5 get_campaign_summary

获取战役摘要（用于 LLM 上下文）。

**请求参数**:
```go
type GetCampaignSummaryRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetCampaignSummaryResponse struct {
    Summary *GameSummary `json:"summary"` // 游戏状态摘要
}
```

---

### 2.2 角色管理 Tools

#### 2.2.1 create_character

创建角色或即兴 NPC。

**请求参数**:
```go
type CreateCharacterRequest struct {
    CampaignID  string      `json:"campaign_id"`  // 必填，所属战役ID
    Name        string      `json:"name"`         // 必填，角色名称
    IsNPC       bool        `json:"is_npc"`       // 可选，是否为NPC，默认false
    NPCType     NPCType     `json:"npc_type"`     // 可选，NPC类型
    PlayerID    string      `json:"player_id"`    // 可选，玩家ID（玩家角色必填）

    // 基础属性
    Race        string      `json:"race"`         // 必填，种族
    Class       string      `json:"class"`        // 必填，职业
    Level       int         `json:"level"`        // 可选，等级，默认1
    Background  string      `json:"background"`   // 可选，背景
    Alignment   string      `json:"alignment"`    // 可选，阵营

    // 属性值（可选，有默认生成规则）
    Abilities   *Abilities  `json:"abilities"`

    // 战斗属性（可选，根据属性值计算）
    HP          *HP         `json:"hp"`
    AC          int         `json:"ac"`
    Speed       int         `json:"speed"`

    // 技能和豁免（可选）
    Skills      map[string]int `json:"skills"`
    Saves       map[string]int `json:"saves"`
}
```

**响应**:
```go
type CreateCharacterResponse struct {
    Character *Character `json:"character"` // 创建的角色实体
}
```

#### 2.2.2 get_character

获取角色详情。

**请求参数**:
```go
type GetCharacterRequest struct {
    CharacterID string `json:"character_id"` // 必填，角色ID
}
```

**响应**:
```go
type GetCharacterResponse struct {
    Character *Character `json:"character"` // 角色实体
}
```

#### 2.2.3 update_character

更新角色信息。

**请求参数**:
```go
type UpdateCharacterRequest struct {
    CharacterID string      `json:"character_id"` // 必填，角色ID
    Name        *string     `json:"name"`         // 可选
    HP          *HP         `json:"hp"`           // 可选
    AC          *int        `json:"ac"`           // 可选
    Conditions  []Condition `json:"conditions"`   // 可选，替换状态列表
    Equipment   []Equipment `json:"equipment"`    // 可选，替换装备列表
    Inventory   []Item      `json:"inventory"`    // 可选，替换背包
}
```

**响应**:
```go
type UpdateCharacterResponse struct {
    Character *Character `json:"character"` // 更新后的角色实体
}
```

#### 2.2.4 list_characters

列出角色。

**请求参数**:
```go
type ListCharactersRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
    IsNPC      *bool  `json:"is_npc"`      // 可选，筛选玩家角色或NPC
}
```

**响应**:
```go
type ListCharactersResponse struct {
    Characters []*Character `json:"characters"` // 角色列表
}
```

#### 2.2.5 delete_character

删除角色。

**请求参数**:
```go
type DeleteCharacterRequest struct {
    CharacterID string `json:"character_id"` // 必填，角色ID
}
```

**响应**:
```go
type DeleteCharacterResponse struct {
    Success bool `json:"success"` // 是否成功
}
```

---

### 2.3 骰子/检定 Tools

#### 2.3.1 roll_dice

投骰子（通用）。

**请求参数**:
```go
type RollDiceRequest struct {
    Formula string `json:"formula"` // 必填，骰子公式（如 "1d20+5", "2d6", "4d6kh3"）
}
```

**响应**:
```go
type RollDiceResponse struct {
    Result *DiceResult `json:"result"` // 骰子结果
}
```

#### 2.3.2 roll_check

属性/技能检定。

**请求参数**:
```go
type RollCheckRequest struct {
    CharacterID string `json:"character_id"` // 必填，角色ID
    Ability     string `json:"ability"`      // 必填，属性（strength, dexterity, etc.）
    Skill       string `json:"skill"`        // 可选，技能（stealth, perception, etc.）
    DC          int    `json:"dc"`           // 可选，难度等级
    Advantage   bool   `json:"advantage"`    // 可选，优势
    Disadvantage bool  `json:"disadvantage"` // 可选，劣势
}
```

**响应**:
```go
type RollCheckResponse struct {
    Result *CheckResult `json:"result"` // 检定结果
}
```

#### 2.3.3 roll_save

豁免检定。

**请求参数**:
```go
type RollSaveRequest struct {
    CharacterID string `json:"character_id"` // 必填，角色ID
    SaveType    string `json:"save_type"`    // 必填，豁免类型（strength, dexterity, constitution, etc.）
    DC          int    `json:"dc"`           // 可选，难度等级
}
```

**响应**:
```go
type RollSaveResponse struct {
    Result *CheckResult `json:"result"` // 检定结果
}
```

---

### 2.4 战斗操作 Tools

#### 2.4.1 start_combat

开始战斗。

**请求参数**:
```go
type StartCombatRequest struct {
    CampaignID      string   `json:"campaign_id"`      // 必填，战役ID
    ParticipantIDs  []string `json:"participant_ids"`  // 必填，参战角色ID列表
    MapID           string   `json:"map_id"`           // 可选，战斗地图ID
}
```

**响应**:
```go
type StartCombatResponse struct {
    Combat *Combat `json:"combat"` // 战斗状态（含先攻顺序）
}
```

#### 2.4.2 get_combat_state

获取战斗状态。

**请求参数**:
```go
type GetCombatStateRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetCombatStateResponse struct {
    Combat *Combat `json:"combat"` // 战斗状态
}
```

#### 2.4.3 attack

攻击。

**请求参数**:
```go
type AttackRequest struct {
    AttackerID string `json:"attacker_id"` // 必填，攻击者ID
    TargetID   string `json:"target_id"`   // 必填，目标ID
    WeaponID   string `json:"weapon_id"`   // 可选，武器ID（默认使用主手武器）
    Advantage  bool   `json:"advantage"`   // 可选，优势
    Disadvantage bool `json:"disadvantage"` // 可选，劣势
}
```

**响应**:
```go
type AttackResponse struct {
    Result   *AttackResult `json:"result"`   // 攻击结果
    Combat   *Combat       `json:"combat"`   // 更新后的战斗状态
}
```

#### 2.4.4 cast_spell

施法。

**请求参数**:
```go
type CastSpellRequest struct {
    CasterID   string   `json:"caster_id"`   // 必填，施法者ID
    SpellName  string   `json:"spell_name"`  // 必填，法术名称
    TargetIDs  []string `json:"target_ids"`  // 可选，目标ID列表
    SpellLevel int      `json:"spell_level"` // 可选，升环等级
}
```

**响应**:
```go
type CastSpellResponse struct {
    Success bool                   `json:"success"` // 是否成功
    Results []SpellTargetResult    `json:"results"` // 每个目标的结果
    Combat  *Combat                `json:"combat"`  // 更新后的战斗状态
}

type SpellTargetResult struct {
    TargetID  string       `json:"target_id"`
    Hit       bool         `json:"hit"`
    Damage    int          `json:"damage"`
    SaveMade  bool         `json:"save_made"`
    Effects   []string     `json:"effects"`
}
```

#### 2.4.5 end_turn

结束当前回合。

**请求参数**:
```go
type EndTurnRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type EndTurnResponse struct {
    Combat *Combat `json:"combat"` // 更新后的战斗状态
}
```

#### 2.4.6 end_combat

结束战斗。

**请求参数**:
```go
type EndCombatRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type EndCombatResponse struct {
    Summary *CombatSummaryResult `json:"summary"` // 战斗结算
}

type CombatSummaryResult struct {
    Rounds       int              `json:"rounds"`        // 总回合数
    Duration     string           `json:"duration"`      // 持续时间
    Participants []ParticipantSummary `json:"participants"` // 参战者统计
    Log          []CombatLogEntry `json:"log"`           // 完整战斗日志
}

type ParticipantSummary struct {
    CharacterID   string `json:"character_id"`
    Name          string `json:"name"`
    DamageDealt   int    `json:"damage_dealt"`
    DamageTaken   int    `json:"damage_taken"`
    HealingDone   int    `json:"healing_done"`
    FinalHP       int    `json:"final_hp"`
    Survived      bool   `json:"survived"`
}
```

---

### 2.5 地图/移动 Tools

#### 2.5.1 get_world_map

获取大地图。

**请求参数**:
```go
type GetWorldMapRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetWorldMapResponse struct {
    Map           *Map      `json:"map"`            // 地图数据
    PartyPosition *Position `json:"party_position"` // 队伍位置
    Locations     []Location `json:"locations"`     // 可进入的地点列表
}
```

#### 2.5.2 move_to

大地图移动。

**请求参数**:
```go
type MoveToRequest struct {
    CampaignID  string `json:"campaign_id"`  // 必填，战役ID
    LocationID  string `json:"location_id"`  // 必填，目标地点ID
    TravelMode  string `json:"travel_mode"`  // 可选，旅行方式（walk, horse, ship）
}
```

**响应**:
```go
type MoveToResponse struct {
    Success      bool      `json:"success"`       // 是否成功
    TimeElapsed  string    `json:"time_elapsed"`  // 消耗时间
    NewPosition  *Position `json:"new_position"`  // 新位置
    GameState    *GameState `json:"game_state"`   // 更新后的游戏状态
    Events       []string  `json:"events"`        // 途中事件（可选）
}
```

#### 2.5.3 enter_battle_map

进入战斗地图。

**请求参数**:
```go
type EnterBattleMapRequest struct {
    CampaignID  string `json:"campaign_id"`  // 必填，战役ID
    LocationID  string `json:"location_id"`  // 必填，地点ID
    BattleMapID string `json:"battle_map_id"` // 可选，指定战斗地图ID
}
```

**响应**:
```go
type EnterBattleMapResponse struct {
    Map        *Map      `json:"map"`        // 战斗地图数据
    Tokens     []Token   `json:"tokens"`     // 初始Token位置
    GameState  *GameState `json:"game_state"` // 更新后的游戏状态
}
```

#### 2.5.4 get_battle_map

获取战斗地图。

**请求参数**:
```go
type GetBattleMapRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetBattleMapResponse struct {
    Map    *Map    `json:"map"`    // 地图数据
    Tokens []Token `json:"tokens"` // 所有Token位置
}
```

#### 2.5.5 move_token

移动 Token。

**请求参数**:
```go
type MoveTokenRequest struct {
    CampaignID string   `json:"campaign_id"` // 必填，战役ID
    TokenID    string   `json:"token_id"`    // 必填，Token ID
    Position   Position `json:"position"`    // 必填，目标位置
}
```

**响应**:
```go
type MoveTokenResponse struct {
    Success        bool     `json:"success"`         // 是否成功
    MovementUsed   int      `json:"movement_used"`   // 消耗移动力
    NewPosition    Position `json:"new_position"`    // 新位置
    TriggeredTiles []string `json:"triggered_tiles"` // 触发的地形效果
}
```

#### 2.5.6 exit_battle_map

离开战斗地图。

**请求参数**:
```go
type ExitBattleMapRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type ExitBattleMapResponse struct {
    GameState *GameState `json:"game_state"` // 更新后的游戏状态（回到大地图）
}
```

---

### 2.6 规则查询 Tools

#### 2.6.1 lookup_spell

查询法术。

**请求参数**:
```go
type LookupSpellRequest struct {
    SpellName string `json:"spell_name"` // 必填，法术名称
}
```

**响应**:
```go
type LookupSpellResponse struct {
    Found bool                   `json:"found"` // 是否找到
    Data  map[string]interface{} `json:"data"`  // 法术数据（来自RAG）
}
```

#### 2.6.2 lookup_item

查询物品。

**请求参数**:
```go
type LookupItemRequest struct {
    ItemName string `json:"item_name"` // 必填，物品名称
}
```

**响应**:
```go
type LookupItemResponse struct {
    Found bool                   `json:"found"` // 是否找到
    Data  map[string]interface{} `json:"data"`  // 物品数据（来自RAG）
}
```

#### 2.6.3 lookup_monster

查询怪物。

**请求参数**:
```go
type LookupMonsterRequest struct {
    MonsterName string `json:"monster_name"` // 必填，怪物名称
}
```

**响应**:
```go
type LookupMonsterResponse struct {
    Found bool                   `json:"found"` // 是否找到
    Data  map[string]interface{} `json:"data"`  // 怪物数据（来自RAG）
}
```

---

### 2.7 上下文管理 Tools

#### 2.7.1 get_context

获取压缩后上下文（简化模式）。

**请求参数**:
```go
type GetContextRequest struct {
    CampaignID    string `json:"campaign_id"`     // 必填，战役ID
    MessageLimit  int    `json:"message_limit"`   // 可选，消息数量限制，默认20
    IncludeCombat bool   `json:"include_combat"`  // 可选，是否包含战斗详情，默认true
}
```

**响应**:
```go
type GetContextResponse struct {
    Context *Context `json:"context"` // 压缩后的上下文
}
```

#### 2.7.2 get_raw_context

获取原始数据（完整模式）。

**请求参数**:
```go
type GetRawContextRequest struct {
    CampaignID string `json:"campaign_id"` // 必填，战役ID
}
```

**响应**:
```go
type GetRawContextResponse struct {
    Campaign    *Campaign    `json:"campaign"`     // 战役数据
    Characters  []*Character `json:"characters"`   // 所有角色
    GameState   *GameState   `json:"game_state"`   // 游戏状态
    Maps        []*Map       `json:"maps"`         // 所有地图
    Messages    []*Message   `json:"messages"`     // 所有对话历史
    ActiveCombat *Combat     `json:"active_combat"` // 当前战斗（如果有）
}
```

#### 2.7.3 save_message

保存对话消息。

**请求参数**:
```go
type SaveMessageRequest struct {
    CampaignID string     `json:"campaign_id"` // 必填，战役ID
    Role       MessageRole `json:"role"`       // 必填，消息角色
    Content    string     `json:"content"`     // 必填，消息内容
    PlayerID   string     `json:"player_id"`   // 可选，玩家ID（user消息）
    ToolCalls  []ToolCall `json:"tool_calls"`  // 可选，工具调用（assistant消息）
}
```

**响应**:
```go
type SaveMessageResponse struct {
    Message *Message `json:"message"` // 保存的消息实体
}
```

---

### 2.8 MCP Tool 清单汇总

| 分类 | Tool | 描述 |
|------|------|------|
| **战役** | create_campaign | 创建战役 |
| | get_campaign | 获取战役 |
| | list_campaigns | 列出战役 |
| | delete_campaign | 删除战役 |
| | get_campaign_summary | 获取战役摘要 |
| **角色** | create_character | 创建角色/NPC |
| | get_character | 获取角色 |
| | update_character | 更新角色 |
| | list_characters | 列出角色 |
| | delete_character | 删除角色 |
| **骰子** | roll_dice | 投骰子 |
| | roll_check | 属性/技能检定 |
| | roll_save | 豁免检定 |
| **战斗** | start_combat | 开始战斗 |
| | get_combat_state | 获取战斗状态 |
| | attack | 攻击 |
| | cast_spell | 施法 |
| | end_turn | 结束回合 |
| | end_combat | 结束战斗 |
| **地图** | get_world_map | 获取大地图 |
| | move_to | 大地图移动 |
| | enter_battle_map | 进入战斗地图 |
| | get_battle_map | 获取战斗地图 |
| | move_token | 移动Token |
| | exit_battle_map | 离开战斗地图 |
| **查询** | lookup_spell | 查询法术 |
| | lookup_item | 查询物品 |
| | lookup_monster | 查询怪物 |
| **上下文** | get_context | 获取压缩上下文 |
| | get_raw_context | 获取原始数据 |
| | save_message | 保存消息 |

**总计**: 30 个 MCP Tools

---

## 第3轮: 业务流程

> 完善处理逻辑和流程分支

### 3.1 战役管理流程

#### 3.1.1 创建战役流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                       create_campaign 流程                           │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ name 非空？
  │    ├─ dm_id 非空？
  │    └─ settings 合法？（max_players 1-10, start_level 1-20）
  │
  ├─ 2. 设置默认值
  │    ├─ Settings.MaxPlayers = 4（如果未设置）
  │    ├─ Settings.StartLevel = 1（如果未设置）
  │    ├─ Settings.Ruleset = "dnd5e"（如果未设置）
  │    └─ Settings.ContextWindow = 20（如果未设置）
  │
  ├─ 3. 生成 ID
  │    └─ Campaign.ID = UUID()
  │
  ├─ 4. 初始化关联数据
  │    ├─ 创建 GameState
  │    │    ├─ GameTime = 默认起始时间
  │    │    └─ CurrentMapType = "world"
  │    └─ 创建默认大地图（可选）
  │
  ├─ 5. 持久化
  │    ├─ 保存 Campaign
  │    └─ 保存 GameState
  │
  └─ 6. 返回结果
       └─ 返回 Campaign 实体
```

#### 3.1.2 获取战役摘要流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                    get_campaign_summary 流程                         │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 查询战役
  │    └─ 验证 Campaign 存在
  │
  ├─ 2. 查询 GameState
  │    └─ 获取当前游戏状态
  │
  ├─ 3. 查询角色
  │    ├─ 获取所有玩家角色（is_npc = false）
  │    └─ 构建 PartyMember 列表
  │
  ├─ 4. 查询战斗状态（如果有）
  │    ├─ 检查 ActiveCombatID
  │    └─ 如果存在，构建 CombatSummary
  │
  ├─ 5. 获取最近事件
  │    └─ 从战斗日志和消息中提取
  │
  └─ 6. 组装返回
       └─ 返回 GameSummary
```

---

### 3.2 角色管理流程

#### 3.2.1 创建角色流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                     create_character 流程                            │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ campaign_id 有效？
  │    ├─ name 非空？
  │    ├─ race 非空？
  │    ├─ class 非空？
  │    └─ 玩家角色需要 player_id
  │
  ├─ 2. 生成缺失属性（如果未提供）
  │    ├─ Abilities: 使用标准阵列或投骰生成
  │    ├─ HP.Max = 类基础HP + Constitution修正
  │    ├─ HP.Current = HP.Max
  │    ├─ AC = 基于装备和敏捷计算
  │    └─ Speed = 种族基础速度
  │
  ├─ 3. 计算衍生属性
  │    ├─ 技能加值（基于属性）
  │    └─ 豁免加值（基于职业）
  │
  ├─ 4. 生成 ID
  │    └─ Character.ID = UUID()
  │
  ├─ 5. 持久化
  │    └─ 保存 Character
  │
  └─ 6. 返回结果
       └─ 返回 Character 实体
```

#### 3.2.2 更新角色流程（HP变化）

```
┌─────────────────────────────────────────────────────────────────────┐
│                    update_character (HP) 流程                        │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 查询角色
  │    └─ 验证 Character 存在
  │
  ├─ 2. 应用 HP 变化
  │    ├─ 伤害: HP.Current -= damage
  │    │    ├─ 如果 HP.Current < 0 → HP.Current = 0
  │    │    └─ 如果 HP.Current = 0 → 标记死亡/昏迷
  │    │
  │    └─ 治疗: HP.Current += healing
  │         └─ 如果 HP.Current > HP.Max → HP.Current = HP.Max
  │
  ├─ 3. 处理临时 HP
  │    └─ Temp HP 不叠加，取较大值
  │
  ├─ 4. 持久化
  │    └─ 保存更新后的 Character
  │
  └─ 5. 返回结果
       └─ 返回更新后的 Character
```

---

### 3.3 战斗系统流程

#### 3.3.1 开始战斗流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                      start_combat 流程                               │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ campaign_id 有效？
  │    ├─ participant_ids 非空？
  │    └─ 所有角色存在？
  │
  ├─ 2. 检查前置条件
  │    └─ 战役中没有进行中的战斗
  │
  ├─ 3. 投先攻
  │    ├─ 对每个参与者:
  │    │    ├─ roll = 1d20
  │    │    ├─ initiative = roll + character.Initiative
  │    │    └─ 创建 Participant
  │    │
  │    └─ 按 initiative 降序排序
  │
  ├─ 4. 初始化战斗状态
  │    ├─ Combat.Round = 1
  │    ├─ Combat.TurnIndex = 0
  │    ├─ Combat.Status = "active"
  │    └─ 设置起始位置（如果有地图）
  │
  ├─ 5. 持久化
  │    ├─ 保存 Combat
  │    └─ 更新 GameState.ActiveCombatID
  │
  └─ 6. 返回结果
       └─ 返回 Combat 实体（含先攻顺序）
```

#### 3.3.2 攻击流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                         attack 流程                                  │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ attacker_id 有效？
  │    ├─ target_id 有效？
  │    └─ 是否当前攻击者的回合？
  │
  ├─ 2. 获取战斗状态
  │    ├─ 获取当前 Combat
  │    └─ 验证 attacker 和 target 都是参战者
  │
  ├─ 3. 投攻击骰
  │    ├─ roll = 1d20
  │    ├─ 如果 advantage: roll = max(1d20, 1d20)
  │    ├─ 如果 disadvantage: roll = min(1d20, 1d20)
  │    ├─ attack_bonus = proficiency + ability_modifier
  │    └─ attack_total = roll + attack_bonus
  │
  ├─ 4. 判定命中
  │    ├─ 如果 roll == 20 → 自动命中 + 暴击
  │    ├─ 如果 roll == 1 → 自动失手
  │    ├─ 如果 attack_total >= target.AC → 命中
  │    └─ 否则 → 失手
  │
  ├─ 5. 投伤害骰（如果命中）
  │    ├─ damage_roll = weapon_damage_dice
  │    ├─ 如果暴击: damage_roll *= 2
  │    ├─ damage_bonus = ability_modifier
  │    └─ total_damage = damage_roll + damage_bonus
  │
  ├─ 6. 应用伤害
  │    ├─ 先扣临时HP
  │    │    └─ target.HP.Temp -= damage
  │    ├─ 临时HP不足时扣当前HP
  │    │    └─ target.HP.Current -= remaining_damage
  │    └─ 更新 Character
  │
  ├─ 7. 记录战斗日志
  │    └─ 添加 CombatLogEntry
  │
  ├─ 8. 持久化
  │    ├─ 保存 Combat
  │    └─ 保存更新的 Character
  │
  └─ 9. 返回结果
       ├─ AttackResult
       └─ 更新后的 Combat
```

#### 3.3.3 结束回合流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                       end_turn 流程                                  │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 获取战斗状态
  │    └─ 验证战斗进行中
  │
  ├─ 2. 处理当前回合结束效果
  │    ├─ 处理持续伤害/治疗
  │    ├─ 减少状态效果持续时间
  │    └─ 清除过期状态
  │
  ├─ 3. 推进回合
  │    ├─ TurnIndex++
  │    │
  │    └─ 如果 TurnIndex >= len(Participants):
  │         ├─ TurnIndex = 0
  │         ├─ Round++
  │         └─ 重置所有 HasActed = false
  │
  ├─ 4. 处理新回合开始效果
  │    └─ 触发新回合事件（如恢复）
  │
  ├─ 5. 持久化
  │    └─ 保存 Combat
  │
  └─ 6. 返回结果
       └─ 返回更新后的 Combat
```

#### 3.3.4 结束战斗流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                       end_combat 流程                                │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 获取战斗状态
  │    └─ 验证战斗进行中
  │
  ├─ 2. 生成战斗统计
  │    ├─ 计算总回合数
  │    ├─ 计算持续时间
  │    └─ 统计每个参战者的:
  │         ├─ 造成伤害
  │         ├─ 承受伤害
  │         ├─ 治疗量
  │         └─ 最终状态
  │
  ├─ 3. 更新战斗状态
  │    ├─ Combat.Status = "finished"
  │    ├─ Combat.EndedAt = now()
  │    └─ GameState.ActiveCombatID = ""
  │
  ├─ 4. 持久化
  │    ├─ 保存 Combat
  │    └─ 保存 GameState
  │
  └─ 5. 返回结果
       └─ 返回 CombatSummaryResult
```

---

### 3.4 地图系统流程

#### 3.4.1 大地图移动流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                        move_to 流程                                  │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ campaign_id 有效？
  │    ├─ location_id 有效？
  │    └─ 当前不在战斗地图中
  │
  ├─ 2. 计算旅行
  │    ├─ 获取当前位置
  │    ├─ 获取目标位置
  │    ├─ 计算距离（格子或预定义）
  │    └─ 根据 travel_mode 计算时间
  │         ├─ walk: 3英里/小时
  │         ├─ horse: 8英里/小时
  │         └─ ship: 根据航程
  │
  ├─ 3. 推进游戏时间
  │    └─ GameState.GameTime += travel_time
  │
  ├─ 4. 更新队伍位置
  │    └─ GameState.PartyPosition = target_location
  │
  ├─ 5. 检查遭遇（可选）
  │    ├─ 投 d20
  │    └─ 如果触发 → 添加到 Events
  │
  ├─ 6. 持久化
  │    └─ 保存 GameState
  │
  └─ 7. 返回结果
       ├─ MoveToResponse
       └─ 包含消耗时间、新位置、事件
```

#### 3.4.2 进入战斗地图流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                   enter_battle_map 流程                              │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ campaign_id 有效？
  │    ├─ location_id 有效？
  │    └─ location 有关联的 battle_map？
  │
  ├─ 2. 获取或创建战斗地图
  │    ├─ 如果指定 battle_map_id → 加载
  │    └─ 否则 → 加载 location 关联的地图
  │
  ├─ 3. 放置角色 Token
  │    ├─ 获取所有玩家角色
  │    ├─ 为每个角色创建 Token
  │    └─ 放置在入口或默认位置
  │
  ├─ 4. 更新游戏状态
  │    ├─ GameState.CurrentMapID = battle_map_id
  │    ├─ GameState.CurrentMapType = "battle"
  │    └─ 更新地图的 Tokens
  │
  ├─ 5. 持久化
  │    ├─ 保存 Map
  │    └─ 保存 GameState
  │
  └─ 6. 返回结果
       ├─ EnterBattleMapResponse
       └─ 包含地图数据和 Token 位置
```

#### 3.4.3 移动 Token 流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                      move_token 流程                                 │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ token 有效？
  │    ├─ 目标位置在地图范围内？
  │    └─ 目标位置非墙壁？
  │
  ├─ 2. 计算移动
  │    ├─ 计算起点到终点的格子数
  │    ├─ 检查角色移动速度
  │    └─ 困难地形消耗双倍移动力
  │
  ├─ 3. 检查是否可移动
  │    ├─ 移动力足够？
  │    └─ 路径无阻挡？（简化：直线检查）
  │
  ├─ 4. 更新 Token 位置
  │    ├─ Token.Position = new_position
  │    └─ 记录消耗的移动力
  │
  ├─ 5. 触发地形效果
  │    ├─ 检查目标格子类型
  │    └─ 添加到 TriggeredTiles
  │
  ├─ 6. 持久化
  │    └─ 保存 Map
  │
  └─ 7. 返回结果
       └─ MoveTokenResponse
```

---

### 3.5 上下文管理流程

#### 3.5.1 获取压缩上下文流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                      get_context 流程                                │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 获取战役和游戏状态
  │    ├─ Campaign
  │    └─ GameState
  │
  ├─ 2. 构建 GameSummary
  │    ├─ CampaignName
  │    ├─ GameTime
  │    ├─ CurrentLocation
  │    ├─ PartyMembers（玩家角色摘要）
  │    ├─ ActiveCombat（如果有）
  │    └─ RecentEvents
  │
  ├─ 3. 获取对话历史
  │    ├─ 查询最近 N 条消息（默认 20）
  │    └─ 按 CreatedAt 排序
  │
  ├─ 4. 计算 Token 估算
  │    └─ 简单估算：字符数 / 4
  │
  └─ 5. 返回结果
       └─ Context
```

#### 3.5.2 保存消息流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                     save_message 流程                                │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 参数验证
  │    ├─ campaign_id 有效？
  │    ├─ role 有效？（user, assistant, system）
  │    └─ content 非空？
  │
  ├─ 2. 生成 ID
  │    └─ Message.ID = UUID()
  │
  ├─ 3. 设置时间戳
  │    └─ Message.CreatedAt = now()
  │
  ├─ 4. 持久化
  │    └─ 保存 Message
  │
  └─ 5. 返回结果
       └─ 返回 Message 实体
```

---

### 3.6 骰子系统流程

#### 3.6.1 骰子公式解析

```
支持的骰子公式格式:
  ├─ 基本格式: NdS（N个S面骰）
  │    例: 1d20, 2d6, 4d8
  │
  ├─ 带修正: NdS+M 或 NdS-M
  │    例: 1d20+5, 2d6-1
  │
  ├─ 保留最高: NdSkhK（保留最高K个）
  │    例: 4d6kh3（投4个d6保留最高3个）
  │
  ├─ 保留最低: NdSklK（保留最低K个）
  │    例: 2d20kl1（劣势）
  │
  └─ 优势/劣势（特殊处理）
       ├─ 优势: max(1d20, 1d20)
       └─ 劣势: min(1d20, 1d20)
```

#### 3.6.2 检定流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                      roll_check 流程                                 │
└─────────────────────────────────────────────────────────────────────┘

开始
  │
  ├─ 1. 获取角色
  │    └─ 查询 Character
  │
  ├─ 2. 计算检定加值
  │    ├─ ability_modifier = 计算属性修正
  │    ├─ skill_bonus = Skills[skill]（如果有）
  │    └─ total_bonus = ability_modifier + skill_bonus
  │
  ├─ 3. 投骰
  │    ├─ 如果 advantage: roll = max(1d20, 1d20)
  │    ├─ 如果 disadvantage: roll = min(1d20, 1d20)
  │    └─ 否则: roll = 1d20
  │
  ├─ 4. 计算结果
  │    ├─ total = roll + total_bonus
  │    ├─ 判定暴击/大失败
  │    └─ 如果有 DC: success = total >= DC
  │
  └─ 5. 返回结果
       └─ CheckResult
```

---

### 3.7 业务规则汇总

#### 3.7.1 属性修正计算

```
属性值 → 修正值
  1     → -5
  2-3   → -4
  4-5   → -3
  6-7   → -2
  8-9   → -1
  10-11 → 0
  12-13 → +1
  14-15 → +2
  16-17 → +3
  18-19 → +4
  20-21 → +5
  ...   → (value - 10) / 2 (向下取整)
```

#### 3.7.2 熟练加值（按等级）

```
等级     熟练加值
 1-4     +2
 5-8     +3
 9-12    +4
 13-16   +5
 17-20   +6
```

#### 3.7.3 AC 计算规则

```
基础 AC 计算:
  ├─ 无护甲: 10 + DEX修正
  ├─ 轻甲: 护甲AC + DEX修正
  ├─ 中甲: 护甲AC + DEX修正（上限+2）
  └─ 重甲: 护甲AC（无DEX加值）

护盾: +2 AC（如果装备）
```

#### 3.7.4 战斗规则

```
先攻:
  ├─ 基础: 1d20 + DEX修正
  └─ 优势/劣势适用

攻击:
  ├─ 命中: 1d20 + 熟练加值 + 属性修正 >= 目标AC
  ├─ 暴击: 自然20，伤害骰翻倍
  └─ 大失败: 自然1，自动失手

伤害:
  ├─ 武器伤害 + 属性修正
  └─ 暴击时骰子翻倍（不含修正）

临时HP:
  ├─ 先消耗临时HP
  └─ 临时HP不足时扣当前HP

倒地/死亡:
  ├─ HP = 0 时倒地
  └─ 需要死亡豁免（后续扩展）
```

---

## 第4轮: 存储与异常

> 存储方案、错误码、边界处理

### 4.1 数据库设计

#### 4.1.1 表结构

**campaigns 表**
```sql
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    dm_id VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_campaigns_dm_id ON campaigns(dm_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
```

**characters 表**
```sql
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_npc BOOLEAN DEFAULT FALSE,
    npc_type VARCHAR(50),
    player_id VARCHAR(255),

    -- 基础属性
    race VARCHAR(100) NOT NULL,
    class VARCHAR(100) NOT NULL,
    level INT DEFAULT 1,
    background VARCHAR(255),
    alignment VARCHAR(50),

    -- 属性值和战斗属性
    abilities JSONB NOT NULL,
    hp JSONB NOT NULL,
    ac INT NOT NULL,
    speed INT DEFAULT 30,
    initiative INT DEFAULT 0,

    -- 技能和豁免
    skills JSONB DEFAULT '{}',
    saves JSONB DEFAULT '{}',

    -- 装备和物品
    equipment JSONB DEFAULT '[]',
    inventory JSONB DEFAULT '[]',

    -- 状态
    conditions JSONB DEFAULT '[]',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_characters_campaign_id ON characters(campaign_id);
CREATE INDEX idx_characters_is_npc ON characters(is_npc);
CREATE INDEX idx_characters_player_id ON characters(player_id);
```

**combats 表**
```sql
CREATE TABLE combats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'active',
    round INT DEFAULT 1,
    turn_index INT DEFAULT 0,
    participants JSONB NOT NULL,
    map_id UUID,
    log JSONB DEFAULT '[]',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_combats_campaign_id ON combats(campaign_id);
CREATE INDEX idx_combats_status ON combats(status);
```

**maps 表**
```sql
CREATE TABLE maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    grid JSONB NOT NULL,
    locations JSONB DEFAULT '[]',
    tokens JSONB DEFAULT '[]',
    parent_id UUID REFERENCES maps(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_maps_campaign_id ON maps(campaign_id);
CREATE INDEX idx_maps_type ON maps(type);
CREATE INDEX idx_maps_parent_id ON maps(parent_id);
```

**game_states 表**
```sql
CREATE TABLE game_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE UNIQUE,
    game_time JSONB NOT NULL,
    party_position JSONB,
    current_map_id UUID REFERENCES maps(id),
    current_map_type VARCHAR(50) DEFAULT 'world',
    weather VARCHAR(50),
    active_combat_id UUID REFERENCES combats(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_game_states_campaign_id ON game_states(campaign_id);
```

**messages 表**
```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    player_id VARCHAR(255),
    tool_calls JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_campaign_id ON messages(campaign_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
```

#### 4.1.2 数据关系图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        数据库关系图                                   │
└─────────────────────────────────────────────────────────────────────┘

                    ┌──────────────┐
                    │  campaigns   │
                    │──────────────│
                    │ id (PK)      │
                    │ name         │
                    │ dm_id        │
                    │ settings     │
                    └──────┬───────┘
                           │
           ┌───────────────┼───────────────┬───────────────┐
           │               │               │               |
           ▼               ▼               ▼               ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │ characters   │ │   combats    │ │    maps      │ │ game_states  │
    │──────────────│ │──────────────│ │──────────────│ │──────────────│
    │ id (PK)      │ │ id (PK)      │ │ id (PK)      │ │ id (PK)      │
    │ campaign_id  │ │ campaign_id  │ │ campaign_id  │ │ campaign_id  │
    │ name         │ │ participants │ │ type         │ │ game_time    │
    │ is_npc       │ │ round        │ │ grid         │ │ party_pos    │
    │ abilities    │ │ log          │ │ tokens       │ │ active_combat│
    └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
                           │                                      │
                           └──────────────────────────────────────┘

    ┌──────────────┐
    │  messages    │
    │──────────────│
    │ id (PK)      │
    │ campaign_id  │
    │ role         │
    │ content      │
    │ tool_calls   │
    └──────────────┘
```

---

### 4.2 错误码定义

#### 4.2.1 错误码格式

```
格式: [模块][类型][序号]
  - 模块: C(战役), H(角色), B(战斗), M(地图), D(骰子), X(上下文), L(查询)
  - 类型: 1(参数错误), 2(不存在), 3(状态错误), 4(权限错误), 5(系统错误)
  - 序号: 01-99

示例: C201 = 战役模块-不存在-01
```

#### 4.2.2 错误码列表

**通用错误 (G)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| G001 | ErrInvalidRequest | 400 | 请求参数无效 |
| G002 | ErrUnauthorized | 401 | 未授权 |
| G003 | ErrForbidden | 403 | 禁止访问 |
| G004 | ErrNotFound | 404 | 资源不存在 |
| G005 | ErrInternalServer | 500 | 服务器内部错误 |

**战役模块 (C)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| C101 | ErrCampaignNameEmpty | 400 | 战役名称为空 |
| C102 | ErrCampaignInvalidSettings | 400 | 战役设置无效 |
| C201 | ErrCampaignNotFound | 404 | 战役不存在 |
| C301 | ErrCampaignAlreadyActive | 400 | 战役已激活 |
| C302 | ErrCampaignAlreadyFinished | 400 | 战役已结束 |

**角色模块 (H)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| H101 | ErrCharacterNameEmpty | 400 | 角色名称为空 |
| H102 | ErrCharacterInvalidRace | 400 | 无效种族 |
| H103 | ErrCharacterInvalidClass | 400 | 无效职业 |
| H104 | ErrCharacterInvalidLevel | 400 | 无效等级(1-20) |
| H201 | ErrCharacterNotFound | 404 | 角色不存在 |
| H202 | ErrCharacterNotInCampaign | 400 | 角色不属于该战役 |
| H301 | ErrCharacterDead | 400 | 角色已死亡 |

**战斗模块 (B)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| B101 | ErrCombatNoParticipants | 400 | 无参战者 |
| B102 | ErrCombatInvalidParticipant | 400 | 无效参战者ID |
| B201 | ErrCombatNotFound | 404 | 战斗不存在 |
| B301 | ErrCombatNotActive | 400 | 战斗未进行中 |
| B302 | ErrCombatAlreadyActive | 400 | 已有进行中的战斗 |
| B303 | ErrCombatNotYourTurn | 400 | 不是你的回合 |
| B304 | ErrCombatTargetNotParticipant | 400 | 目标不是参战者 |

**地图模块 (M)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| M101 | ErrMapInvalidPosition | 400 | 无效位置坐标 |
| M102 | ErrMapPositionOutOfRange | 400 | 位置超出地图范围 |
| M103 | ErrMapPositionBlocked | 400 | 位置被阻挡 |
| M104 | ErrMapInsufficientMovement | 400 | 移动力不足 |
| M201 | ErrMapNotFound | 404 | 地图不存在 |
| M202 | ErrLocationNotFound | 404 | 地点不存在 |
| M203 | ErrTokenNotFound | 404 | Token不存在 |
| M301 | ErrMapNotInBattleMap | 400 | 当前不在战斗地图中 |
| M302 | ErrMapAlreadyInBattleMap | 400 | 当前已在战斗地图中 |

**骰子模块 (D)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| D101 | ErrDiceInvalidFormula | 400 | 无效骰子公式 |
| D102 | ErrDiceTooManyDice | 400 | 骰子数量过多(>100) |
| D103 | ErrDiceTooManySides | 400 | 骰子面数过多(>1000) |
| H201 | ErrCharacterNotFound | 404 | 角色不存在(检定用) |

**上下文模块 (X)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| X101 | ErrContextMessageEmpty | 400 | 消息内容为空 |
| X102 | ErrContextInvalidRole | 400 | 无效消息角色 |
| C201 | ErrCampaignNotFound | 404 | 战役不存在 |

**查询模块 (L)**
| 错误码 | 名称 | HTTP状态 | 描述 |
|--------|------|----------|------|
| L101 | ErrLookupQueryEmpty | 400 | 查询内容为空 |
| L501 | ErrLookupServiceUnavailable | 503 | RAG服务不可用 |

#### 4.2.3 错误响应格式

```go
// ErrorResponse 统一错误响应
type ErrorResponse struct {
    Code      string `json:"code"`       // 错误码
    Message   string `json:"message"`    // 错误消息
    Details   string `json:"details"`    // 详细信息（调试用）
    RequestID string `json:"request_id"` // 请求ID
}
```

---

### 4.3 边界处理

#### 4.3.1 输入验证

| 字段 | 验证规则 |
|------|----------|
| Campaign.Name | 非空，1-255字符 |
| Campaign.MaxPlayers | 1-10 |
| Campaign.StartLevel | 1-20 |
| Character.Name | 非空，1-255字符 |
| Character.Level | 1-20 |
| Character.Abilities.* | 1-30 |
| Character.AC | 1-30 |
| Character.Speed | 0-120 |
| Combat.Participants | 2-50个 |
| Map.Grid.Width/Height | 1-1000 |
| Position.X/Y | >= 0 |
| Message.Content | 非空，1-100000字符 |
| Dice.Formula | 符合骰子公式格式 |

#### 4.3.2 并发控制

```
战斗操作锁:
  ├─ 每个 Campaign 一个读写锁
  ├─ 读操作: get_combat_state, get_battle_map
  └─ 写操作: start_combat, attack, end_turn, end_combat

角色操作锁:
  ├─ 每个 Character 一个锁
  └─ 用于: update_character, attack（修改HP）

策略:
  ├─ 使用数据库行锁（SELECT FOR UPDATE）处理高并发
  └─ 短事务，快速释放
```

#### 4.3.3 数据一致性

```
事务边界:
  ├─ create_campaign: Campaign + GameState（单事务）
  ├─ start_combat: Combat + GameState（单事务）
  ├─ attack: Combat + Character（单事务）
  ├─ end_combat: Combat + GameState（单事务）
  └─ move_to: GameState（单事务）

级联删除:
  ├─ 删除 Campaign 级联删除所有关联数据
  └─ 删除 Character 从 Combat.Participants 中移除
```

#### 4.3.4 数据限制

| 资源 | 限制 | 处理方式 |
|------|------|----------|
| 每个战役角色数 | 100 | 创建时检查，超出拒绝 |
| 每场战斗参战者 | 50 | 开始战斗时检查 |
| 每个战役地图数 | 50 | 创建时检查 |
| 对话历史总数 | 10000 | 超出自动归档旧消息 |
| 战斗日志条数 | 1000 | 超出自动截断旧日志 |
| 单条消息长度 | 100KB | 截断或拒绝 |

---

### 4.4 日志规范

#### 4.4.1 日志级别

| 级别 | 场景 |
|------|------|
| DEBUG | 详细调试信息（开发环境） |
| INFO | 正常操作（创建、更新、删除） |
| WARN | 可恢复的异常（验证失败、重试） |
| ERROR | 操作失败（业务错误、系统错误） |

#### 4.4.2 日志格式

```json
{
  "timestamp": "2025-02-16T10:30:00Z",
  "level": "INFO",
  "request_id": "req-abc123",
  "campaign_id": "camp-xyz",
  "tool": "attack",
  "message": "Attack executed",
  "details": {
    "attacker_id": "char-001",
    "target_id": "char-002",
    "damage": 12,
    "hit": true
  }
}
```

#### 4.4.3 审计日志

需要记录的关键操作:
- 创建/删除战役
- 创建/删除角色
- 开始/结束战斗
- 所有骰子投掷结果（防作弊）

---

## 第5轮: 测试规范

> 测试场景和验收标准

### 5.1 测试策略

#### 5.1.1 测试金字塔

```
                    ┌─────────────────┐
                    │    E2E 测试     │  少量
                    │  (完整流程)      │  关键场景
                    ├─────────────────┤
                    │   集成测试       │  中等
                    │  (服务+存储)     │  API契约
                    ├─────────────────┤
                    │    单元测试      │  大量
                    │ (规则引擎+逻辑)  │  覆盖率>80%
                    └─────────────────┘
```

#### 5.1.2 测试目录结构

```
packages/server/
└── tests/
    ├── unit/                    # 单元测试
    │   ├── rules/              # 规则引擎测试
    │   │   ├── dice_test.go
    │   │   ├── ability_test.go
    │   │   ├── combat_test.go
    │   │   └── movement_test.go
    │   ├── service/            # 服务层测试
    │   │   ├── campaign_test.go
    │   │   ├── character_test.go
    │   │   ├── combat_test.go
    │   │   └── context_test.go
    │   └── models/             # 模型测试
    │       └── models_test.go
    ├── integration/             # 集成测试
    │   ├── store/              # 存储层测试
    │   │   ├── campaign_store_test.go
    │   │   ├── character_store_test.go
    │   │   └── combat_store_test.go
    │   └── api/                # API集成测试
    │       └── mcp_tools_test.go
    └── e2e/                     # 端到端测试
        ├── combat_flow_test.go
        └── adventure_flow_test.go
```

---

### 5.2 单元测试

#### 5.2.1 骰子系统测试

| 测试场景 | 输入 | 预期输出 |
|----------|------|----------|
| 基本骰子 | 1d20 | 总和1-20 |
| 多个骰子 | 2d6 | 总和2-12 |
| 带修正 | 1d20+5 | 总和6-25 |
| 负修正 | 1d20-3 | 总和-2到17 |
| 保留最高 | 4d6kh3 | 3个值的总和 |
| 暴击检测 | 1d20 | 自然20标记为critical |
| 大失败检测 | 1d20 | 自然1标记为fumble |
| 无效公式 | "abc" | 返回错误 |
| 骰子过多 | "1000d6" | 返回错误 |

#### 5.2.2 属性修正测试

| 测试场景 | 输入 | 预期输出 |
|----------|------|----------|
| 属性1 | 1 | -5 |
| 属性10 | 10 | 0 |
| 属性14 | 14 | +2 |
| 属性20 | 20 | +5 |
| 属性3 | 3 | -4 |

#### 5.2.3 战斗规则测试

| 测试场景 | 输入 | 预期输出 |
|----------|------|----------|
| 攻击命中 | roll=15, attack_bonus=5, AC=18 | 命中 |
| 攻击失手 | roll=10, attack_bonus=5, AC=18 | 失手 |
| 暴击命中 | roll=20 | 自动命中+暴击 |
| 大失败 | roll=1 | 自动失手 |
| 优势取高 | advantage=true | 取两个骰子中高的 |
| 劣势取低 | disadvantage=true | 取两个骰子中低的 |
| 伤害计算 | weapon="1d8+3", crit=false | 伤害4-11 |
| 暴击伤害 | weapon="1d8+3", crit=true | 伤害5-19（骰子翻倍） |

#### 5.2.4 移动规则测试

| 测试场景 | 输入 | 预期输出 |
|----------|------|----------|
| 普通移动 | speed=30, distance=5格 | 成功 |
| 超速移动 | speed=30, distance=7格 | 失败 |
| 困难地形 | speed=30, difficult=true, distance=3格 | 消耗30移动力 |
| 墙壁阻挡 | target=wall | 失败 |

---

### 5.3 集成测试

#### 5.3.1 战役管理测试

| 测试场景 | 步骤 | 预期结果 |
|----------|------|----------|
| 创建战役 | 调用create_campaign | 返回Campaign，ID非空 |
| 获取战役 | 调用get_campaign | 返回正确的Campaign |
| 列出战役 | 调用list_campaigns | 返回战役列表 |
| 删除战役 | 调用delete_campaign | 成功，后续查询返回404 |
| 级联删除 | 删除战役后查询角色 | 角色也被删除 |

#### 5.3.2 角色管理测试

| 测试场景 | 步骤 | 预期结果 |
|----------|------|----------|
| 创建玩家角色 | 调用create_character(is_npc=false) | 角色创建成功 |
| 创建NPC | 调用create_character(is_npc=true) | NPC创建成功 |
| 筛选NPC | 调用list_characters(is_npc=true) | 只返回NPC |
| 更新HP | 调用update_character(hp) | HP正确更新 |
| HP降为0 | 伤害使HP=0 | 角色标记为倒地 |

#### 5.3.3 战斗系统测试

| 测试场景 | 步骤 | 预期结果 |
|----------|------|----------|
| 开始战斗 | 调用start_combat | 先攻顺序已排序 |
| 重复开始 | 战斗进行中再次调用 | 返回错误 |
| 攻击流程 | 调用attack | 返回命中/失手结果 |
| 回合推进 | 调用end_turn | TurnIndex+1 |
| 回合循环 | 最后一人end_turn | Round+1, TurnIndex=0 |
| 结束战斗 | 调用end_combat | 返回战斗统计 |

#### 5.3.4 上下文管理测试

| 测试场景 | 步骤 | 预期结果 |
|----------|------|----------|
| 保存消息 | 调用save_message | 消息保存成功 |
| 获取上下文 | 调用get_context | 返回压缩后的上下文 |
| 滑动窗口 | 保存30条消息后获取 | 只返回最近20条 |
| 原始数据 | 调用get_raw_context | 返回所有数据 |

---

### 5.4 E2E测试

#### 5.4.1 完整战斗流程

```
场景: 队伍遭遇哥布林

步骤:
1. 创建战役
2. 创建2个玩家角色（战士、法师）
3. 创建1个NPC哥布林
4. 开始战斗（3个参战者）
5. 战士攻击哥布林（命中）
6. 结束战士回合
7. 哥布林攻击战士（失手）
8. 结束哥布林回合
9. 法师施法攻击哥布林
10. 哥布林HP降为0
11. 结束战斗
12. 验证战斗统计

验证点:
- 先攻顺序正确
- HP变化正确
- 回合推进正确
- 战斗日志完整
```

#### 5.4.2 冒险流程

```
场景: 队伍从城镇旅行到地牢

步骤:
1. 创建战役
2. 创建玩家角色
3. 获取大地图
4. 移动到地牢入口
5. 验证游戏时间推进
6. 进入战斗地图
7. 移动角色Token
8. 离开战斗地图
9. 返回大地图

验证点:
- 位置更新正确
- 游戏时间推进
- 地图切换正确
```

---

### 5.5 性能测试

#### 5.5.1 基准测试

| 测试项 | 目标 |
|--------|------|
| 骰子投掷 | < 1ms |
| 攻击计算 | < 5ms |
| 上下文获取（100条消息） | < 50ms |
| 战斗状态获取 | < 20ms |

#### 5.5.2 负载测试

| 测试项 | 条件 | 目标 |
|--------|------|------|
| 并发请求 | 10个并发攻击 | 响应时间 < 100ms |
| 大量消息 | 1000条消息的战役 | get_context < 200ms |
| 复杂战斗 | 20个参战者 | start_combat < 500ms |

---

### 5.6 测试覆盖率目标

| 模块 | 单元测试 | 集成测试 | 总覆盖率 |
|------|----------|----------|----------|
| Dice | 90% | - | 90% |
| Rules | 85% | - | 85% |
| Service | 70% | 80% | 85% |
| Store | - | 90% | 90% |
| MCP Adapter | 60% | 80% | 80% |
| **整体** | - | - | **80%** |

---

### 5.7 验收标准

#### 5.7.1 功能验收

- [ ] 所有30个MCP Tools实现并通过测试
- [ ] 骰子系统支持所有公式格式
- [ ] 战斗流程完整（开始→攻击→回合→结束）
- [ ] 地图系统完整（大地图↔战斗地图）
- [ ] 上下文管理正确（压缩、存储）

#### 5.7.2 质量验收

- [ ] 单元测试覆盖率 >= 80%
- [ ] 所有集成测试通过
- [ ] 所有E2E测试通过
- [ ] 无高优先级代码问题（静态分析）

#### 5.7.3 文档验收

- [ ] API文档完整（所有Tools有描述）
- [ ] 错误码文档完整
- [ ] 部署文档完整

---

## 设计审视

> 执行设计一致性、完整性、冲突检测

### 审视报告

#### 一、设计一致性检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 追溯完整性 | ✅ 通过 | 所有细化项可追溯到高层次设计 |
| 范围正确性 | ✅ 通过 | 无超出原始设计范围的添加 |
| 遗漏检查 | ✅ 通过 | 高层次设计要求均已覆盖 |

**追溯矩阵**:

| 高层次设计要求 | 详细设计位置 |
|----------------|--------------|
| 9个模块 | 第1轮: 8个数据结构 + Dice模块 |
| 30个MCP Tools | 第2轮: 全部定义 |
| 4个交互流程 | 第3轮: 全部细化 |
| 存储方案 | 第4轮: 6个表定义 |
| 错误处理 | 第4轮: 30+错误码 |
| 测试规范 | 第5轮: 完整测试策略 |

#### 二、模块间一致性检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 数据结构一致性 | ✅ 通过 | 跨模块引用的类型一致 |
| 接口契约匹配 | ✅ 通过 | 请求/响应与数据结构匹配 |
| 数据流闭环 | ✅ 通过 | 所有数据流有始有终 |
| ID格式统一 | ✅ 通过 | 全部使用UUID |

**跨模块数据结构验证**:

| 引用关系 | 数据类型 | 一致性 |
|----------|----------|--------|
| Character.CampaignID → Campaign.ID | UUID | ✅ |
| Combat.CampaignID → Campaign.ID | UUID | ✅ |
| Map.CampaignID → Campaign.ID | UUID | ✅ |
| Message.CampaignID → Campaign.ID | UUID | ✅ |
| GameState.CampaignID → Campaign.ID | UUID | ✅ |
| Combat.MapID → Map.ID | UUID | ✅ |
| GameState.ActiveCombatID → Combat.ID | UUID | ✅ |
| Token.CharacterID → Character.ID | UUID | ✅ |
| Participant.CharacterID → Character.ID | UUID | ✅ |
| Location.BattleMapID → Map.ID | UUID | ✅ |

#### 三、完整性检查

| 检查项 | 状态 | 说明 |
|--------|------|------|
| API完整性 | ✅ 通过 | 每个Tool有请求/响应定义 |
| 存储完整性 | ✅ 通过 | 每个实体有表定义 |
| 错误处理完整性 | ✅ 通过 | 每个模块有错误码 |
| 测试覆盖完整性 | ✅ 通过 | 每个模块有测试规范 |

**API与存储映射**:

| 实体 | 表 | CRUD Tools |
|------|-----|------------|
| Campaign | campaigns | create/get/list/delete/get_summary |
| Character | characters | create/get/update/list/delete |
| Combat | combats | start/get_state/attack/cast_spell/end_turn/end |
| Map | maps | get_world/enter_battle/get_battle/exit_battle |
| Message | messages | save/get_context/get_raw_context |
| GameState | game_states | (通过其他Tools间接操作) |

#### 四、冲突检测

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 字段类型冲突 | ✅ 无冲突 | 同名字段类型一致 |
| 错误码冲突 | ✅ 无冲突 | 所有错误码唯一 |
| API路径冲突 | ⚠️ 不适用 | MCP协议无路径概念 |
| 存储键冲突 | ✅ 无冲突 | 表名和索引无冲突 |

**错误码唯一性验证**:

| 模块 | 错误码范围 | 冲突 |
|------|------------|------|
| 通用(G) | G001-G005 | 无 |
| 战役(C) | C101-C302 | 无 |
| 角色(H) | H101-H301 | 无 |
| 战斗(B) | B101-B304 | 无 |
| 地图(M) | M101-M302 | 无 |
| 骰子(D) | D101-D103 | 无 |
| 上下文(X) | X101-X102 | 无 |
| 查询(L) | L101, L501 | 无 |

#### 五、发现的问题与建议

**无重大问题发现**

**优化建议**:

1. **性能优化**: 考虑为高频查询（get_combat_state）添加缓存
2. **扩展预留**: GameState 表可预留更多字段用于未来扩展
3. **日志增强**: 建议添加结构化日志字段便于查询分析

#### 六、审视结论

| 结论 | 说明 |
|------|------|
| 设计完整性 | ✅ 完整 |
| 设计一致性 | ✅ 一致 |
| 实现可行性 | ✅ 可行 |
| 下一步 | 可开始实现 |

---

## 附录

### A. 快速参考

#### A.1 MCP Tools 快速索引

```
战役: create_campaign, get_campaign, list_campaigns, delete_campaign, get_campaign_summary
角色: create_character, get_character, update_character, list_characters, delete_character
骰子: roll_dice, roll_check, roll_save
战斗: start_combat, get_combat_state, attack, cast_spell, end_turn, end_combat
地图: get_world_map, move_to, enter_battle_map, get_battle_map, move_token, exit_battle_map
查询: lookup_spell, lookup_item, lookup_monster
上下文: get_context, get_raw_context, save_message
```

#### A.2 数据库表快速索引

```
campaigns    - 战役主表
characters   - 角色表（含NPC）
combats      - 战斗表
maps         - 地图表
game_states  - 游戏状态表
messages     - 对话消息表
```

#### A.3 错误码快速索引

```
G - 通用错误
C - 战役模块
H - 角色模块
B - 战斗模块
M - 地图模块
D - 骰子模块
X - 上下文模块
L - 查询模块
```

### B. 修订历史

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| v1.0 | 2025-02-16 | 初始版本，完成5轮细化 |

---

**设计文档状态**: ✅ 已完成，可开始实现
