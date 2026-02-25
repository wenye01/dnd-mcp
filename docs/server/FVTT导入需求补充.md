# 高层次设计: MCP Server 游戏规则引擎

## 文档信息

- **版本**: v2.0
- **创建日期**: 2025-02-16
- **更新日期**: 2026-02-24
- **设计范围**: MCP Server 核心架构与模块设计（含 FVTT 导入支持）
- **状态**: 待细化
- **变更说明**: 新增 FVTT/UVTT 导入支持模块

---

## 1. 背景与目标

### 1.1 背景

DND MCP 项目采用 Client-Server 架构：

**Client 现状**：
- 已完成基础框架：会话管理、对话历史、WebSocket 推送
- 使用 Redis + PostgreSQL 存储
- 提供 HTTP API 供前端调用

**Server 开发状态**：
- ✅ M1 项目基础设施 - 已完成
- ✅ M2 战役管理 - 已完成
- 🚧 M3-M8 - 待开发

**FVTT 导入需求**：
- 大量用户已有 FVTT (Foundry VTT) 的角色卡和地图数据
- FVTT 格式虽为闭源，但 JSON 结构半公开，约 60-80% 字段可解析
- dnd5e 系统是 FVTT 最流行的游戏系统，覆盖主要用户群体

### 1.2 目标

**主要目标**：
- 设计一个独立、可复用的 D&D 5e 游戏规则引擎
- 通过 MCP 协议暴露游戏能力（Tools）
- 与 Client 形成清晰的职责边界
- **支持 FVTT/UVTT 格式导入，降低用户迁移成本**

**成功标准**：
- Server 不依赖 LLM，保持轻量
- 规则引擎可独立测试
- 支持 Client 的两种模式（完整模式/简化模式）
- **导入后的角色可直接参与战斗**
- **导入后的地图可正常显示和移动 Token**

### 1.3 范围

**包含**：
- 规则引擎层（角色、战斗、骰子、地图）
- 战斗系统（回合制、先攻、攻击、施法）
- 地图系统（大地图、战斗地图、移动）
- 骰子与检定系统
- **FVTT/UVTT 导入系统（地图、角色、物品）**

**不包含**：
- 规则可选/冲突处理
- 地图编辑功能
- 战斗 AI
- 状态效果系统
- **非 dnd5e 游戏系统的 FVTT 导入**
- **FVTT 光源/ActiveEffect 导入**

---

## 2. 架构与模块

### 2.1 技术选型

| 决策项 | 选择 | 理由 |
|--------|------|------|
| **语言** | Go 1.24+ | 与 Client 技术栈一致 |
| **协议** | MCP 协议 + HTTP 传输 | 标准 LLM 工具协议 |
| **存储** | PostgreSQL + JSONB | 关系型数据，灵活存储 |
| **LLM 依赖** | 无 | 保持简单 |
| **图片存储** | 数据库 JSONB (Base64) | 简化部署 |
| **导入策略** | 核心字段转换 + 保留原始 JSON | 不丢失数据 |

### 2.2 架构分层

```
┌─────────────────────────────────────────────────────────────────────┐
│                    MCP Server 架构分层 (v2.0)                        │
├─────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    MCP 适配层 (MCP Layer)                       │ │
│  │   新增: import_map, import_character, import_items Tools       │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                ▼                                     │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    业务服务层 (Service Layer)                   │ │
│  │   新增: ImportService (导入流程编排)                            │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                ▼                                     │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    导入模块 (Import Module) [新增]              │ │
│  │   Parser(格式解析) → Converter(数据转换) → Validator(校验)     │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                ▼                                     │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                    规则引擎层 + 存储层                          │ │
│  │   Character(扩展) / Map(扩展) / Token(扩展)                    │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 模块划分

| 模块 | 职责 | FVTT 影响 |
|------|------|-----------|
| **Campaign** | 战役生命周期管理 | ✅ 已完成 (M2) |
| **Character** | 角色数据管理（含 NPC） | 🔴 重大扩展 |
| **Combat** | 战斗状态和流程管理 | 无影响 |
| **Dice** | 骰子投掷和检定计算 | 无影响 |
| **Map** | 大地图和战斗地图管理 | 🟡 中等扩展 |
| **Token** | Token 显示和位置管理 | 🟡 中等扩展 |
| **Context** | 对话历史和上下文管理 | 无影响 |
| **Lookup** | 规则查询 | 无影响 |
| **Import** | 外部格式导入 [新增] | 新增模块 |

### 2.4 Import 模块内部结构

```
packages/server/internal/import/
├── interface.go           # 公共接口定义
├── service.go             # 导入服务（流程编排）
├── format/
│   ├── uvtt.go            # UVTT 格式定义
│   └── fvtt.go            # FVTT 格式定义
├── parser/
│   ├── parser.go          # Parser 接口和工厂
│   ├── uvtt_parser.go     # UVTT 解析器
│   └── fvtt_parser.go     # FVTT 解析器
├── converter/
│   ├── map_converter.go   # 地图转换器
│   ├── character_converter.go  # 角色转换器
│   └── item_converter.go  # 物品转换器
└── validator/
    └── validator.go       # 导入数据校验
```

---

## 3. 交互流程

### 3.1 导入流程（新增）

```
Client                                    Server
   │                                         │
   │  import_character(campaign_id, data)    │
   │ ─────────────────────────────────────>  │
   │                           ┌─────────────────────────────────┐
   │                           │ Import Service:                 │
   │                           │  1. 格式检测 (FVTT/UVTT)        │
   │                           │  2. 解析为中间格式               │
   │                           │  3. 转换为内部模型               │
   │                           │  4. 校验 + 持久化               │
   │                           └─────────────────────────────────┘
   │  返回 { character, warnings }           │
   │ <─────────────────────────────────────  │
```

### 3.2 地图导入流程

```
Client                                    Server
   │                                         │
   │  import_map(campaign_id, data)          │
   │ ─────────────────────────────────────>  │
   │                           ┌─────────────────────────────────┐
   │                           │ 格式检测:                        │
   │                           │  UVTT: 检查 resolution 字段      │
   │                           │  FVTT: 检查 _id, grid 字段       │
   │                           │                                 │
   │                           │ UVTT 解析: image, walls, grid   │
   │                           │ FVTT 解析: background, walls,   │
   │                           │           tokens                │
   │                           └─────────────────────────────────┘
   │  返回 { map, detected_format, warnings }│
   │ <─────────────────────────────────────  │
```

---

## 4. 核心概念

### 4.1 术语定义

| 术语 | 定义 |
|------|------|
| **ImportFormat** | 导入格式类型：`uvtt`, `fvtt` |
| **ImportMeta** | 导入元数据：format, original_id, imported_at, raw_json |
| **中间格式** | Parser 输出的格式特定结构 |
| **字段映射** | 外部字段到内部字段的转换规则 |

### 4.2 关键决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 游戏系统范围 | 仅 dnd5e | 覆盖主要用户，工作量可控 |
| 导入策略 | 核心字段转换 + 保留原始 JSON | 不丢失数据，可后续扩展 |
| 图片存储 | 本地数据库（Base64） | 简化部署，数据集中管理 |
| 光源系统 | 暂不支持 | 非核心功能，复杂度高 |
| Parser/Converter 分离 | 是 | 便于扩展新格式 |
| 装备槽位系统 | 新增 | 匹配 D&D 装备规则 |

---

## 5. MCP Tools 变更

### 5.1 新增 Tools

| Tool | 描述 | 优先级 | 依赖 |
|------|------|--------|------|
| `import_map` | 导入地图（UVTT/FVTT Scene） | P0 | M6 |
| `import_character` | 导入角色（FVTT Actor） | P0 | M3 |
| `import_items` | 批量导入物品 | P1 | M3 |
| `export_character` | 导出角色为 FVTT 格式 | P3（可选） | M3 |

### 5.2 扩展 Tools

| Tool | 变更内容 |
|------|----------|
| `create_character` | 新增可选参数：Image, Experience, Currency, Spells, Features 等 |
| `update_character` | 新增可选参数：DeathSaves, EquipmentSlots, Inventory 等 |
| `get_character` | 响应扩展：返回完整扩展字段 |
| `get_battle_map` | 响应扩展：返回 Image, Walls |

---

## 6. 数据结构扩展

### 6.1 Character 扩展字段

| 字段 | 类型 | 说明 |
|------|------|------|
| Image | string | 角色图片 |
| Experience | int | XP 值 |
| DeathSaves | *DeathSaves | 死亡豁免 |
| Proficiency | int | 熟练加值 |
| Speed | *Speed | 移动速度（结构化，多类型） |
| Skills | map[string]*Skill | 技能（含熟练/专精） |
| Saves | map[string]*Save | 豁免（含熟练） |
| Currency | *Currency | 货币 |
| Equipment | *EquipmentSlots | 装备槽位（替代数组） |
| Spells | *Spellbook | 法术书 |
| Features | []Feature | 专长/特性 |
| Biography | *Biography | 传记 |
| Traits | *Traits | 特性/抗性等 |
| ImportMeta | *ImportMeta | 导入元数据 |

### 6.2 Map 扩展字段

| 字段 | 类型 | 说明 |
|------|------|------|
| Image | *MapImage | 地图图片（Base64/URL + 尺寸） |
| Walls | []Wall | 墙体数据（坐标 + 类型） |
| ImportMeta | *ImportMeta | 导入元数据 |

### 6.3 Token 扩展字段

| 字段 | 类型 | 说明 |
|------|------|------|
| Name | string | Token 名称 |
| Image | string | Token 图片 |
| Width/Height | int | 格子尺寸 |
| Rotation | int | 旋转角度 |
| Scale/Alpha | float64 | 缩放/透明度 |
| ActorLink | bool | 是否链接角色 |
| Disposition | int | 态度 |
| Hidden/Locked | bool | 隐藏/锁定 |
| Bar1/Bar2 | *TokenBar | 血条 |

### 6.4 ImportMeta 结构

```go
type ImportMeta struct {
    Format      string    `json:"format"`        // "fvtt" / "uvtt"
    OriginalID  string    `json:"original_id,omitempty"`
    ImportedAt  time.Time `json:"imported_at"`
    RawJSON     string    `json:"raw_json,omitempty"` // 压缩后的原始 JSON
}
```

---

## 7. 开发里程碑（更新）

### 7.1 里程碑总览

| 里程碑 | 主题 | 状态 | FVTT 影响 |
|--------|------|------|-----------|
| M1 | 项目基础设施 | ✅ 已完成 | 无 |
| M2 | 战役管理 | ✅ 已完成 | 无 |
| M3 | 角色管理 | 待开发 | 🔴 数据结构扩展 |
| M4 | 骰子系统 | 待开发 | 无 |
| M5 | 战斗系统 | 待开发 | 无 |
| M6 | 地图系统 | 待开发 | 🟡 数据结构扩展 |
| M7 | 上下文管理 | 待开发 | 无 |
| M8 | 规则查询 | 待开发 | 无 |
| **M9** | **导入功能** | **新增** | **新增模块** |

### 7.2 依赖关系

```
M1 ──> M2 ──> M3 ──> M4 ──> M5
       │             │
       │             └──> M6 ──> M9 (导入)
       │
       └──> M7

M3 ─────────────────────────────────┘
M6 ─────────────────────────────────┘
```

### 7.3 M9: 导入功能里程碑

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T9-1 | Import 模块框架 | M6 | M |
| T9-2 | UVTT Parser | T9-1 | M |
| T9-3 | FVTT Scene Parser | T9-1 | M |
| T9-4 | Map Converter | T9-2, T9-3 | M |
| T9-5 | import_map Tool | T9-4 | M |
| T9-6 | FVTT Actor Parser | T9-1 | L |
| T9-7 | Character Converter | T9-6, M3 | L |
| T9-8 | import_character Tool | T9-7 | M |
| T9-9 | Item Converter | T9-6 | M |
| T9-10 | import_items Tool | T9-9 | M |

---

## 8. 设计约束

### 8.1 技术约束

- Go 语言实现
- PostgreSQL 数据库
- MCP 协议 + HTTP 传输
- Server 不依赖 LLM
- 图片存储使用 JSONB (Base64)

### 8.2 业务约束

- 最终支持完整 5e 规则
- **FVTT 导入仅支持 dnd5e 系统**
- **暂不支持 FVTT 光源/ActiveEffect 导入**

---

## 9. 扩展性考虑

| 扩展点 | 说明 |
|--------|------|
| FVTT 光源导入 | 扩展 Map 结构支持光源 |
| FVTT 导出 | 将内部数据导出为 FVTT 格式 |
| 其他游戏系统 | pf2e, wfrp4e 等（新增 Parser） |

---

## 10. 完成检查

- [x] 能否列出所有需要细化的模块？ → 10 个模块（含 Import）
- [x] 每个模块的职责是否清晰？ → 已定义职责和边界
- [x] FVTT 导入设计是否完整？ → 新增模块 + 数据结构扩展
- [x] 与现有里程碑的关系是否清晰？ → M9 新增，M3/M6 扩展

---

## 附录 A: FVTT dnd5e 字段映射速查

### Actor → Character

| FVTT 路径 | 内部字段 | 复杂度 |
|-----------|----------|--------|
| `name` | `Name` | 简单 |
| `system.abilities.*.value` | `Abilities.*` | 简单 |
| `system.attributes.hp` | `HP` | 简单 |
| `system.attributes.ac.value` | `AC` | 简单 |
| `system.attributes.movement` | `Speed` | 中等 |
| `system.skills.*` | `Skills` | 中等 |
| `system.currency` | `Currency` | 简单 |
| `items[type=class]` | `Class`, `Level` | 中等 |
| `items[type=weapon]` | `EquipmentSlots` | 复杂 |
| `items[type=spell]` | `Spellbook` | 复杂 |

### Scene → Map

| FVTT 路径 | 内部字段 | 复杂度 |
|-----------|----------|--------|
| `name` | `Name` | 简单 |
| `background` | `Image` | 中等 |
| `grid.size` | `Grid.CellSize` | 简单 |
| `walls` | `Walls` | 中等 |
| `tokens` | `Tokens` | 中等 |

---

## 附录 B: 修订历史

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| v1.0 | 2025-02-16 | 初始版本 |
| v2.0 | 2026-02-24 | 新增 FVTT/UVTT 导入支持设计 |
