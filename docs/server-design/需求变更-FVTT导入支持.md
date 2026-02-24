# 需求变更：FVTT 导入支持

## 文档信息

- **版本**: v1.0
- **创建日期**: 2026-02-24
- **变更类型**: 功能扩展
- **影响范围**: Server 数据结构、MCP Tools、存储层
- **状态**: 待评审

---

## 1. 变更背景

### 1.1 原始需求

Server 设计支持 UVTT 格式导入地图，角色和物品使用自定义数据结构。

### 1.2 新增需求

支持 FVTT (Foundry VTT) 格式导入：
- **地图**: 同时支持 UVTT 和 FVTT Scene
- **人物卡**: 支持 FVTT Actor (dnd5e 系统)
- **物品装备**: 支持 FVTT Item

### 1.3 设计决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 游戏系统 | 仅 dnd5e | 覆盖主要用户，工作量可控 |
| 导入策略 | 核心字段转换 + 保留原始 JSON | 不丢失数据，可后续扩展 |
| 图片存储 | 本地数据库 | 简化部署，数据集中管理 |
| 光源系统 | 暂不支持 | 非核心功能，后续扩展 |

---

## 2. 影响分析

### 2.1 模块影响范围

| 模块 | 影响 | 说明 |
|------|------|------|
| **Character** | 🔴 重大 | 数据结构大幅扩展 |
| **Map** | 🟡 中等 | 新增图片存储、墙体数据 |
| **Token** | 🟡 中等 | 扩展显示属性 |
| **Import** | 🟢 新增 | 新增导入模块 |
| **Combat** | 🟢 无 | 无影响 |
| **Dice** | 🟢 无 | 无影响 |
| **Campaign** | 🟢 无 | 无影响 |
| **Context** | 🟢 无 | 无影响 |
| **Lookup** | 🟢 无 | 无影响 |

### 2.2 数据结构变更汇总

| 实体 | 变更类型 | 字段数变化 |
|------|----------|-----------|
| Character | 扩展 | +25 字段 |
| Map | 扩展 | +3 字段 |
| Token | 扩展 | +10 字段 |
| Equipment | 重构 | 从简单结构改为槽位系统 |
| Item | 扩展 | +5 字段 |
| Spell | 新增 | 新实体 |
| Feature | 新增 | 新实体 |

### 2.3 MCP Tools 变更

| Tool | 变更 |
|------|------|
| `import_map` | 新增 |
| `import_character` | 新增 |
| `import_items` | 新增 |
| `export_character` | 新增（可选） |
| `create_character` | 参数扩展 |
| `update_character` | 参数扩展 |
| `get_character` | 响应扩展 |

---

## 3. Character 数据结构变更

### 3.1 原始结构（参考）

```go
type Character struct {
    ID          string
    CampaignID  string
    Name        string
    IsNPC       bool
    NPCType     NPCType
    PlayerID    string

    Race        string
    Class       string
    Level       int
    Background  string
    Alignment   string

    Abilities   *Abilities
    HP          *HP
    AC          int
    Speed       int
    Initiative  int

    Skills      map[string]int
    Saves       map[string]int

    Equipment   []Equipment  // 简单数组
    Inventory   []Item       // 简单结构

    Conditions  []Condition

    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 3.2 变更后结构

```go
type Character struct {
    // === 基础字段（保持不变） ===
    ID          string
    CampaignID  string
    Name        string
    IsNPC       bool
    NPCType     NPCType
    PlayerID    string

    // === 扩展基础属性 ===
    Image       string       // 新增：角色图片
    Race        string
    Class       string
    Level       int
    Background  string
    Alignment   string
    Experience  int          // 新增：XP 值

    // === 属性值（保持不变） ===
    Abilities   *Abilities

    // === 战斗属性（扩展） ===
    HP          *HP
    AC          int
    Speed       *Speed       // 变更：int → *Speed 结构体
    Initiative  int
    DeathSaves  *DeathSaves  // 新增：死亡豁免

    // === 技能和豁免（扩展） ===
    Skills      map[string]*Skill    // 变更：int → *Skill 结构体
    Saves       map[string]*Save     // 变更：int → *Save 结构体
    Proficiency int                   // 新增：熟练加值

    // === 装备和物品（重构） ===
    Currency    *Currency            // 新增：货币
    Equipment   *EquipmentSlots      // 变更：数组 → 槽位结构
    Inventory   []InventoryItem      // 变更：结构扩展
    Spells      *Spellbook           // 新增：法术书
    Features    []Feature            // 新增：专长/特性

    // === 状态（保持不变） ===
    Conditions  []Condition

    // === 背景信息（新增） ===
    Biography   *Biography           // 新增：传记
    Traits      *Traits              // 新增：特性/抗性等

    // === 导入元数据（新增） ===
    ImportMeta  *ImportMeta          // 新增：导入来源信息

    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 3.3 新增子结构

#### 3.3.1 Speed（移动速度）

```go
type Speed struct {
    Walk    int  `json:"walk"`
    Burrow  int  `json:"burrow,omitempty"`  // 挖掘
    Climb   int  `json:"climb,omitempty"`   // 攀爬
    Fly     int  `json:"fly,omitempty"`     // 飞行
    Swim    int  `json:"swim,omitempty"`    // 游泳
    Hover   bool `json:"hover,omitempty"`   // 悬停
}
```

**映射来源**: FVTT `system.attributes.movement`

#### 3.3.2 DeathSaves（死亡豁免）

```go
type DeathSaves struct {
    Successes int `json:"successes"` // 成功次数 (0-3)
    Failures  int `json:"failures"`  // 失败次数 (0-3)
}
```

**映射来源**: FVTT `system.attributes.death`

#### 3.3.3 Skill（技能）- 扩展

```go
type Skill struct {
    Ability    string `json:"ability"`     // 关联属性 (str/dex/...)
    Bonus      int    `json:"bonus"`       // 总加值
    Proficient bool   `json:"proficient"`  // 是否熟练
    Expertise  bool   `json:"expertise"`   // 是否专精（双倍熟练）
}
```

**映射来源**: FVTT `system.skills.*`

#### 3.3.4 Save（豁免）- 扩展

```go
type Save struct {
    Bonus      int  `json:"bonus"`
    Proficient bool `json:"proficient"`
}
```

**映射来源**: FVTT `system.abilities.*.proficient` + 计算

#### 3.3.5 Currency（货币）

```go
type Currency struct {
    PP int `json:"pp"` // 铂金币
    GP int `json:"gp"` // 金币
    EP int `json:"ep"` // 电金币
    SP int `json:"sp"` // 银币
    CP int `json:"cp"` // 铜币
}
```

**映射来源**: FVTT `system.currency`

#### 3.3.6 EquipmentSlots（装备槽位）

```go
type EquipmentSlots struct {
    MainHand   *EquipmentItem `json:"main_hand,omitempty"`
    OffHand    *EquipmentItem `json:"off_hand,omitempty"`
    Head       *EquipmentItem `json:"head,omitempty"`
    Armor      *EquipmentItem `json:"armor,omitempty"`
    Cloak      *EquipmentItem `json:"cloak,omitempty"`
    Gloves     *EquipmentItem `json:"gloves,omitempty"`
    Boots      *EquipmentItem `json:"boots,omitempty"`
    Ring1      *EquipmentItem `json:"ring1,omitempty"`
    Ring2      *EquipmentItem `json:"ring2,omitempty"`
    Amulet     *EquipmentItem `json:"amulet,omitempty"`
    Belt       *EquipmentItem `json:"belt,omitempty"`
    Attuned    []*EquipmentItem `json:"attuned,omitempty"` // 已同调物品
}
```

**说明**: 替代原有的 `[]Equipment` 数组，支持标准装备槽位。

#### 3.3.7 EquipmentItem（装备物品）- 重构

```go
type EquipmentItem struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`          // weapon/equipment/loot
    Image        string                 `json:"image,omitempty"`
    Description  string                 `json:"description,omitempty"`
    Rarity       string                 `json:"rarity,omitempty"`
    Attunement   bool                   `json:"attunement,omitempty"`
    Attuned      bool                   `json:"attuned,omitempty"`
    Equipped     bool                   `json:"equipped"`

    // 武器属性
    Damage       *Damage                `json:"damage,omitempty"`
    Properties   []string               `json:"properties,omitempty"`
    Range        string                 `json:"range,omitempty"`

    // 护甲属性
    AC           *ArmorAC               `json:"ac,omitempty"`
    StrengthReq  int                    `json:"strength_req,omitempty"`
    StealthDis   bool                   `json:"stealth_disadvantage,omitempty"`

    // 原始数据
    RawSystem    map[string]interface{} `json:"raw_system,omitempty"`
}
```

**映射来源**: FVTT `items[type=weapon/equipment]`

#### 3.3.8 InventoryItem（背包物品）- 扩展

```go
type InventoryItem struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`         // consumable/loot/tool/...
    Image       string                 `json:"image,omitempty"`
    Description string                 `json:"description,omitempty"`
    Quantity    int                    `json:"quantity"`
    Weight      float64                `json:"weight,omitempty"`
    Uses        *ItemUses              `json:"uses,omitempty"`
    RawSystem   map[string]interface{} `json:"raw_system,omitempty"`
}
```

**映射来源**: FVTT `items[type=consumable/loot/tool]`

#### 3.3.9 Spellbook（法术书）

```go
type Spellbook struct {
    Class       string    `json:"class,omitempty"`
    Ability     string    `json:"ability,omitempty"`    // 施法属性
    DC          int       `json:"dc,omitempty"`         // 法术DC
    AttackBonus int       `json:"attack_bonus,omitempty"`

    Slots       SpellSlots  `json:"slots"`             // 法术位
    Known       []*Spell    `json:"known,omitempty"`   // 已知法术
    Prepared    []*Spell    `json:"prepared,omitempty"`// 准备法术
}
```

**映射来源**: FVTT `items[type=spell]` + `system.spells`

#### 3.3.10 Spell（法术）

```go
type Spell struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Level       int                    `json:"level"`        // 0-9 (0=戏法)
    School      string                 `json:"school"`       // evocation/...
    Image       string                 `json:"image,omitempty"`
    Description string                 `json:"description,omitempty"`

    Casting     *SpellCasting          `json:"casting,omitempty"`
    Duration    string                 `json:"duration,omitempty"`
    Range       string                 `json:"range,omitempty"`
    Components  *SpellComponents       `json:"components,omitempty"`
    Damage      *SpellDamage           `json:"damage,omitempty"`

    Prepared    bool                   `json:"prepared"`
    RawSystem   map[string]interface{} `json:"raw_system,omitempty"`
}
```

**映射来源**: FVTT `items[type=spell]`

#### 3.3.11 Feature（专长/特性）

```go
type Feature struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`         // feat/class/race/background
    Source      string                 `json:"source,omitempty"`
    Level       int                    `json:"level,omitempty"`
    Description string                 `json:"description,omitempty"`
    Image       string                 `json:"image,omitempty"`
    RawSystem   map[string]interface{} `json:"raw_system,omitempty"`
}
```

**映射来源**: FVTT `items[type=feat/class/race/background]`

#### 3.3.12 Biography（传记）

```go
type Biography struct {
    Value   string `json:"value"`             // 详细背景
    Public  string `json:"public,omitempty"`  // 对外公开的描述
}
```

**映射来源**: FVTT `system.details.biography`

#### 3.3.13 Traits（特性）

```go
type Traits struct {
    PersonalityTraits string `json:"personality_traits,omitempty"`
    Ideals            string `json:"ideals,omitempty"`
    Bonds             string `json:"bonds,omitempty"`
    Flaws             string `json:"flaws,omitempty"`

    Size              string   `json:"size,omitempty"`
    Languages         []string `json:"languages,omitempty"`
    DamageResistances []string `json:"damage_resistances,omitempty"`
    DamageImmunities  []string `json:"damage_immunities,omitempty"`
    DamageVulnerabilities []string `json:"damage_vulnerabilities,omitempty"`
    ConditionImmunities []string `json:"condition_immunities,omitempty"`
}
```

**映射来源**: FVTT `system.traits`

#### 3.3.14 ImportMeta（导入元数据）

```go
type ImportMeta struct {
    Format      string    `json:"format"`        // "fvtt" / "uvtt"
    Version     string    `json:"version,omitempty"`
    OriginalID  string    `json:"original_id,omitempty"`
    ImportedAt  time.Time `json:"imported_at"`
    RawJSON     string    `json:"raw_json,omitempty"` // 压缩后的原始 JSON
}
```

---

## 4. Map 数据结构变更

### 4.1 原始结构（参考）

```go
type Map struct {
    ID          string
    CampaignID  string
    Name        string
    Type        MapType
    Grid        *Grid
    Locations   []Location
    Tokens      []Token
    ParentID    string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.2 变更后结构

```go
type Map struct {
    // === 基础字段（保持不变） ===
    ID          string
    CampaignID  string
    Name        string
    Type        MapType

    // === 新增：图片数据 ===
    Image       *MapImage       // 新增

    // === 格子系统（保持不变） ===
    Grid        *Grid

    // === 地点和标记（保持不变） ===
    Locations   []Location
    Tokens      []Token

    // === 新增：墙体数据 ===
    Walls       []Wall          // 新增

    ParentID    string

    // === 新增：导入元数据 ===
    ImportMeta  *ImportMeta     // 新增

    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.3 新增子结构

#### 4.3.1 MapImage（地图图片）

```go
type MapImage struct {
    Data     string `json:"data,omitempty"`     // Base64 编码图片数据
    URL      string `json:"url,omitempty"`      // 或者外部 URL
    Width    int    `json:"width"`              // 原始宽度（像素）
    Height   int    `json:"height"`             // 原始高度（像素）
    Format   string `json:"format,omitempty"`   // png/jpeg/webp
}
```

**说明**: 存储地图图片数据，支持 Base64 或 URL。

**设计决策变更**: 原设计"图片前端管理"，现改为"本地数据库保存图片"。

#### 4.3.2 Wall（墙体）

```go
type Wall struct {
    ID       string  `json:"id"`
    X1       float64 `json:"x1"`
    Y1       float64 `json:"y1"`
    X2       float64 `json:"x2"`
    Y2       float64 `json:"y2"`
    Door     int     `json:"door"`      // 0=墙, 1=门, 2=暗门
    Move     int     `json:"move"`      // 0=阻挡, 1=可通过
    Sight    int     `json:"sight"`     // 0=阻挡视线, 1=透明
}
```

**映射来源**: FVTT `walls`

---

## 5. Token 数据结构变更

### 5.1 原始结构（参考）

```go
type Token struct {
    ID           string
    CharacterID  string
    Position     Position
    Size         TokenSize
    Visible      bool
}
```

### 5.2 变更后结构

```go
type Token struct {
    // === 基础字段（保持不变） ===
    ID           string
    CharacterID  string
    Name         string              // 新增
    Position     Position
    Size         TokenSize
    Visible      bool

    // === 新增：显示属性 ===
    Image        string              // 新增：Token 图片
    Width        int                 // 新增：格子宽度
    Height       int                 // 新增：格子高度
    Rotation     int                 // 新增
    Scale        float64             // 新增
    Alpha        float64             // 新增：透明度

    // === 新增：关联属性 ===
    ActorLink    bool                // 新增：是否链接到角色
    Disposition  int                 // 新增：-1敌对, 0中立, 1友好
    Hidden       bool                // 新增
    Locked       bool                // 新增

    // === 新增：血条 ===
    Bar1         *TokenBar           // 新增
    Bar2         *TokenBar           // 新增
}
```

### 5.3 新增子结构

#### 5.3.1 TokenBar（Token 血条）

```go
type TokenBar struct {
    Attribute string `json:"attribute"` // 关联属性路径
    Value     int    `json:"value"`
    Max       int    `json:"max"`
}
```

---

## 6. 新增 MCP Tools

### 6.1 import_map

**描述**: 导入地图（支持 UVTT 和 FVTT Scene）

**请求**:
```go
type ImportMapRequest struct {
    CampaignID string `json:"campaign_id"` // 必填
    Data       string `json:"data"`        // 必填，JSON 或 Base64
    Format     string `json:"format"`      // 可选: "uvtt" / "fvtt"，自动检测
    Name       string `json:"name"`        // 可选，覆盖名称
}
```

**响应**:
```go
type ImportMapResponse struct {
    Map      *Map     `json:"map"`
    Format   string   `json:"detected_format"`
    Warnings []string `json:"warnings"`
}
```

### 6.2 import_character

**描述**: 导入角色（支持 FVTT Actor）

**请求**:
```go
type ImportCharacterRequest struct {
    CampaignID string `json:"campaign_id"` // 必填
    Data       string `json:"data"`        // 必填，FVTT Actor JSON
    Name       string `json:"name"`        // 可选，覆盖名称
    AsNPC      bool   `json:"as_npc"`      // 可选，强制作为 NPC 导入
}
```

**响应**:
```go
type ImportCharacterResponse struct {
    Character *Character `json:"character"`
    Warnings  []string   `json:"warnings"`
    Skipped   []string   `json:"skipped"`    // 未导入的字段列表
}
```

### 6.3 import_items

**描述**: 批量导入物品

**请求**:
```go
type ImportItemsRequest struct {
    CampaignID  string `json:"campaign_id"`
    CharacterID string `json:"character_id,omitempty"` // 添加到指定角色
    Data        string `json:"data"`                   // FVTT Item JSON 数组
}
```

**响应**:
```go
type ImportItemsResponse struct {
    Items    []InventoryItem `json:"items"`
    Warnings []string        `json:"warnings"`
}
```

### 6.4 export_character（可选）

**描述**: 导出角色为 FVTT 格式

**请求**:
```go
type ExportCharacterRequest struct {
    CharacterID string `json:"character_id"`
    Format      string `json:"format"` // "fvtt"
}
```

**响应**:
```go
type ExportCharacterResponse struct {
    Data     string   `json:"data"`     // FVTT Actor JSON
    Warnings []string `json:"warnings"` // 无法导出的字段
}
```

---

## 7. 现有 MCP Tools 变更

### 7.1 create_character

**变更**: 参数扩展，支持更多字段

**新增可选参数**:
```go
type CreateCharacterRequest struct {
    // ... 现有字段 ...

    // 新增可选参数
    Image       string
    Experience  int
    Currency    *Currency
    Biography   *Biography
    Traits      *Traits
    Spells      *Spellbook
    Features    []Feature
}
```

### 7.2 update_character

**变更**: 参数扩展，支持更新新增字段

**新增可选参数**:
```go
type UpdateCharacterRequest struct {
    // ... 现有字段 ...

    // 新增可选参数
    Image       *string
    Experience  *int
    Currency    *Currency
    DeathSaves  *DeathSaves
    Biography   *Biography
    Traits      *Traits
    Spells      *Spellbook
    Features    []Feature
    Equipment   *EquipmentSlots  // 替代原有 Equipment
    Inventory   []InventoryItem  // 替代原有 Inventory
}
```

---

## 8. 数据库表结构变更

### 8.1 characters 表

```sql
-- 变更说明：扩展字段，使用 JSONB 存储复杂结构

CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_npc BOOLEAN DEFAULT FALSE,
    npc_type VARCHAR(50),
    player_id VARCHAR(255),

    -- 新增：角色图片
    image TEXT,

    -- 基础属性
    race VARCHAR(100) NOT NULL,
    class VARCHAR(100) NOT NULL,
    level INT DEFAULT 1,
    background VARCHAR(255),
    alignment VARCHAR(50),
    experience INT DEFAULT 0,              -- 新增

    -- 属性值和战斗属性
    abilities JSONB NOT NULL,
    hp JSONB NOT NULL,
    ac INT NOT NULL,
    speed JSONB NOT NULL,                  -- 变更：INT → JSONB
    initiative INT DEFAULT 0,
    death_saves JSONB,                     -- 新增

    -- 技能和豁免
    skills JSONB DEFAULT '{}',             -- 结构变更
    saves JSONB DEFAULT '{}',              -- 结构变更
    proficiency INT DEFAULT 2,             -- 新增

    -- 装备和物品
    currency JSONB,                        -- 新增
    equipment JSONB DEFAULT '{}',          -- 结构变更
    inventory JSONB DEFAULT '[]',          -- 结构变更
    spells JSONB,                          -- 新增
    features JSONB DEFAULT '[]',           -- 新增

    -- 状态
    conditions JSONB DEFAULT '[]',

    -- 背景信息（新增）
    biography JSONB,
    traits JSONB,

    -- 导入元数据（新增）
    import_meta JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 8.2 maps 表

```sql
-- 变更说明：新增图片、墙体、导入元数据字段

CREATE TABLE maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,

    -- 新增：图片数据
    image JSONB,

    grid JSONB NOT NULL,
    locations JSONB DEFAULT '[]',
    tokens JSONB DEFAULT '[]',

    -- 新增：墙体数据
    walls JSONB DEFAULT '[]',

    parent_id UUID REFERENCES maps(id) ON DELETE SET NULL,

    -- 新增：导入元数据
    import_meta JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## 9. FVTT dnd5e 字段映射表

### 9.1 Actor → Character

| FVTT 字段路径 | 内部字段 | 转换规则 |
|--------------|----------|----------|
| `_id` | `ImportMeta.OriginalID` | 直接映射 |
| `name` | `Name` | 直接映射 |
| `type` | `IsNPC` | `character` → false, 其他 → true |
| `img` | `Image` | 存储为 Base64 |
| `system.abilities.str.value` | `Abilities.Strength` | 直接映射 |
| `system.abilities.dex.value` | `Abilities.Dexterity` | 直接映射 |
| `system.abilities.con.value` | `Abilities.Constitution` | 直接映射 |
| `system.abilities.int.value` | `Abilities.Intelligence` | 直接映射 |
| `system.abilities.wis.value` | `Abilities.Wisdom` | 直接映射 |
| `system.abilities.cha.value` | `Abilities.Charisma` | 直接映射 |
| `system.attributes.hp` | `HP` | 映射 value/max/temp |
| `system.attributes.ac.value` | `AC` | 直接映射 |
| `system.attributes.movement` | `Speed` | 映射各移动类型 |
| `system.attributes.init.value` | `Initiative` | value 字段 |
| `system.attributes.death` | `DeathSaves` | success/failure |
| `system.details.xp.value` | `Experience` | 直接映射，计算 Level |
| `system.details.race` | `Race` | 直接映射 |
| `system.details.background` | `Background` | 直接映射 |
| `system.details.alignment` | `Alignment` | 直接映射 |
| `system.details.biography` | `Biography` | value/public |
| `system.skills.*` | `Skills` | 转换为 map |
| `system.traits.*` | `Traits` | 语言、抗性等 |
| `system.currency` | `Currency` | pp/gp/ep/sp/cp |
| `system.prof` | `Proficiency` | 直接映射 |
| `items[type=class]` | `Class`, `Level` | 提取职业和等级 |
| `items[type=weapon]` | `EquipmentSlots.MainHand` | 转换为 EquipmentItem |
| `items[type=equipment]` | `EquipmentSlots.*` | 按 armor.type 分配槽位 |
| `items[type=consumable]` | `Inventory` | 转换为 InventoryItem |
| `items[type=loot]` | `Inventory` | 转换为 InventoryItem |
| `items[type=tool]` | `Inventory` | 转换为 InventoryItem |
| `items[type=spell]` | `Spellbook.Known` | 转换为 Spell |
| `items[type=feat]` | `Features` | 转换为 Feature |
| `items[type=background]` | `Background` + `Features` | 提取特性 |
| `items[type=race]` | `Race` + `Features` | 提取种族特性 |

### 9.2 Scene → Map

| FVTT 字段路径 | 内部字段 | 转换规则 |
|--------------|----------|----------|
| `_id` | `ImportMeta.OriginalID` | 直接映射 |
| `name` | `Name` | 直接映射 |
| `background` | `Image` | 转换为 MapImage |
| `width` | `Image.Width` | 直接映射 |
| `height` | `Image.Height` | 直接映射 |
| `grid.size` | `Grid.CellSize` | 像素→游戏单位 |
| `grid.distance` | `Grid.CellSize` | 每格距离（英尺） |
| `walls` | `Walls` | 转换为 Wall 数组 |
| `tokens` | `Tokens` | 转换为 Token 数组 |

---

## 10. 新增模块

### 10.1 Import 模块

```
packages/server/
└── internal/
    └── import/                    # 新增导入模块
        ├── interface.go           # 导入器接口
        ├── format/                # 格式定义
        │   ├── uvtt.go
        │   └── fvtt.go
        ├── parser/                # 解析器
        │   ├── uvtt_parser.go
        │   └── fvtt_parser.go
        ├── converter/             # 转换器
        │   ├── map_converter.go
        │   ├── character_converter.go
        │   └── item_converter.go
        └── service.go             # 导入服务
```

### 10.2 接口定义

```go
// ImportFormat 导入格式类型
type ImportFormat string

const (
    FormatUVTT ImportFormat = "uvtt"
    FormatFVTT ImportFormat = "fvtt"
)

// Parser 解析器接口
type Parser interface {
    CanParse(data []byte) bool
    Parse(data []byte) (interface{}, error)
}

// Converter 转换器接口
type Converter interface {
    Convert(parsed interface{}, opts ImportOptions) (interface{}, []string, error)
    Validate(parsed interface{}) []string
}
```

---

## 11. 实现计划

### 11.1 里程碑划分

| 里程碑 | 内容 | 优先级 |
|--------|------|--------|
| **M-Import-1** | Character 核心字段导入 | P0 |
| **M-Import-2** | Map 基础导入（UVTT + FVTT） | P0 |
| **M-Import-3** | Equipment 槽位系统 | P1 |
| **M-Import-4** | Inventory 和 Currency | P1 |
| **M-Import-5** | Spellbook 和 Spells | P1 |
| **M-Import-6** | Features | P2 |
| **M-Import-7** | Token 扩展属性 | P2 |
| **M-Import-8** | Character 导出 | P3 |

### 11.2 依赖关系

```
M-Import-1 (Character 核心)
    │
    ├──→ M-Import-3 (Equipment)
    │
    ├──→ M-Import-4 (Inventory)
    │
    └──→ M-Import-5 (Spellbook)

M-Import-2 (Map 基础)
    │
    └──→ M-Import-7 (Token)
```

---

## 12. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| FVTT 格式变化 | 解析失败 | 保留原始 JSON，版本检测 |
| dnd5e system 字段结构复杂 | 映射不完整 | 核心字段优先，其他保留原始 |
| 图片存储占用空间 | 数据库膨胀 | 压缩存储，可选外部存储 |
| 向后兼容性 | 旧数据格式不兼容 | 数据迁移脚本，版本控制 |

---

## 13. 待确认事项

- [ ] 是否需要支持 FVTT 导出功能？
- [ ] 图片存储大小限制？
- [ ] 是否需要支持批量导入（多个角色/地图）？
- [ ] 导入失败时的回滚策略？

---

## 附录：参考资料

- [FVTT-UVTT 格式研究报告](./research/FVTT-UVTT-格式研究报告.md)
- [详细设计-MCP-Server](./详细设计-MCP-Server.md)
- [设计-MCP-Server-游戏规则引擎](./设计-MCP-Server-游戏规则引擎.md)
