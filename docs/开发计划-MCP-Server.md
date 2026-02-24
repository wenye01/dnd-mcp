# 开发计划: MCP Server 游戏规则引擎

## 文档信息

- **版本**: v2.0
- **更新日期**: 2026-02-24
- **变更说明**: 新增 M9 导入功能里程碑，扩展 M3/M6 任务以支持 FVTT 导入

---

## 1. 概述

### 1.1 目标

将 MCP Server 详细设计书转化为任务驱动的增量开发计划。每个任务是完整的垂直切片，可独立开发、测试、验证。

### 1.2 输入

- 设计文档: [详细设计-MCP-Server.md](./详细设计-MCP-Server.md)
- 需求规格: [设计-MCP-Server-游戏规则引擎-v2.md](./设计-MCP-Server-游戏规则引擎-v2.md)
- 变更文档: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md)

### 1.3 约束

- 技术栈: Go 1.24+, PostgreSQL
- 架构模式: Handler → Service → Store → Models
- 测试要求: 单元测试 + 集成测试
- 包路径: `github.com/dnd-mcp/server`
- 图片存储: 数据库 JSONB (Base64)
- 导入策略: 核心字段转换 + 保留原始 JSON

---

## 2. 里程碑总览

| 里程碑 | 主题 | 任务数 | 核心交付 | FVTT 影响 |
|--------|------|--------|----------|-----------|
| M1 | 项目基础设施 | 4 | 可运行的服务框架、数据库连接 | 无 |
| M2 | 战役管理 | 5 | 战役 CRUD、游戏状态管理 | ✅ 已完成 |
| M3 | 角色管理 | 6 | 角色/NPC 创建和管理 | 🔴 数据结构扩展 |
| M4 | 骰子系统 | 3 | 骰子投掷、属性检定 | 无 |
| M5 | 战斗系统 | 6 | 完整战斗流程 | 无 |
| M6 | 地图系统 | 7 | 大地图和战斗地图 | 🟡 数据结构扩展 |
| M7 | 上下文管理 | 3 | 对话存储、上下文压缩 | 无 |
| M8 | 规则查询 | 3 | RAG 集成、规则查询 | 无 |
| **M9** | **导入功能** | **10** | **FVTT/UVTT 导入** | **新增** |

### 依赖关系

```
M1 ──> M2 ──> M3 ──> M4 ──> M5
       │             │
       │             └──> M6 ──> M9 (导入)
       │                    ↑
       └──> M7              │
                            │
M3 ─────────────────────────┘
M4 ──> M8 (可选依赖)
```

---

## 3. 任务依赖图

```
                           T1-1 (项目结构)
                               │
                   ┌───────────┼───────────┐
                   │           │           │
                   v           v           v
                T1-2        T1-3        T1-4
             (配置系统)   (数据库层)   (MCP框架)
                   │           │           │
                   └───────────┼───────────┘
                               │
                               v
                            T2-1 (战役模型)
                               │
                   ┌───────────┼───────────┐
                   │           │           │
                   v           v           v
               T2-2         T2-3       T2-4
            (战役存储)    (战役服务)  (战役Tools)
                   │           │           │
                   └───────────┼───────────┘
                               │
                               v
                       T2-5 (游戏状态)
                               │
                               v
                       T3-1 (角色模型)
                               │
                   ┌───────────┼───────────┐
                   │           │           │
                   v           v           v
               T3-2         T3-3       T3-4
            (角色存储)    (角色服务)  (角色Tools)
                   │           │           │
                   └───────────┼───────────┘
                               │
                               v
                           T3-5
                        (NPC管理)
                               │
                               v
                       T3-6 (角色扩展) ─────────────────┐
                               │                         │
                               v                         │
                       T4-1 (骰子模型)                   │
                               │                         │
                   ┌───────────┼───────────┐             │
                   │           │           │             │
                   v           v           v             │
               T4-2         T4-3                         │
            (骰子服务)    (骰子Tools)                    │
                               │                         │
                   ┌───────────┴───────────┐             │
                   │                       │             │
                   v                       v             │
               T5-1                     T6-1             │
            (战斗模型)               (地图模型)          │
               │                       │                 │
    ┌──────────┼──────────┐           ...                │
    │          │          │                             │
    v          v          v                             │
 T5-2      T5-3       T5-4                              │
(战斗存储) (战斗服务) (战斗Tools)                        │
    │          │          │                             │
    └──────────┼──────────┘                             │
               │                                        │
               v                                        │
          T5-5 (回合管理)                               │
               │                                        │
               v                                        │
          T5-6 (战斗结算)                               │
                                                        │
                               T6-1 ──> T6-7 (地图扩展)─┘
                                                        │
    ... (M7, M8 类似结构)                               │
                                                        │
                                                        v
                                               T9-1 (Import框架)
                                                        │
                                    ┌───────────────────┼───────────────────┐
                                    │                   │                   │
                                    v                   v                   v
                                T9-2               T9-3               T9-6
                            (UVTT Parser)     (FVTT Scene)        (FVTT Actor)
                                    │                   │                   │
                                    └───────┬───────────┘                   │
                                            │                               │
                                            v                               v
                                        T9-4                           T9-7
                                    (Map Converter)            (Character Converter)
                                            │                               │
                                            v                               v
                                        T9-5                           T9-8
                                    (import_map)              (import_character)
                                                                            │
                                                                            v
                                                                        T9-9
                                                                    (Item Converter)
                                                                            │
                                                                            v
                                                                        T9-10
                                                                    (import_items)
```

---

## 4. Milestone 1: 项目基础设施

### 4.1 目标

搭建可运行的 Server 项目框架，包含配置、数据库连接和 MCP 协议支持。

### 4.2 范围

- 包含: 项目结构、配置系统、数据库层、MCP 框架
- 不包含: 业务逻辑

### 4.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T1-1 | 创建项目结构 | - | S |
| T1-2 | 配置系统 | T1-1 | S |
| T1-3 | 数据库层 | T1-1 | M |
| T1-4 | MCP 框架 | T1-1 | M |

### 4.4 验收标准

- [ ] 项目可编译，无错误
- [ ] 数据库连接成功
- [ ] MCP 协议框架可响应基本请求
- [ ] 单元测试通过

---

## 5. 详细任务定义

### Task T1-1: 创建项目结构

**一句话描述**: 创建 packages/server 目录结构和基础文件。

#### 需求来源

- 设计文档: 第1轮 数据结构
- 架构模式: Handler → Service → Store → Models

#### 实现范围

**数据层**:
- 模型: 无（本任务只创建目录结构）

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 无

#### 约束条件

- 前置任务: 无
- 技术约束: Go 1.24+, 模块路径 `github.com/dnd-mcp/server`
- 边界条件: 遵循 Client 项目的目录结构模式

#### 验收标准

- [ ] 目录结构创建完成
- [ ] go.mod 初始化
- [ ] main.go 可编译

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/go.mod | Go 模块定义 |
| 新增 | packages/server/cmd/server/main.go | 服务入口 |
| 新增 | packages/server/internal/models/ | 模型目录 |
| 新增 | packages/server/internal/store/ | 存储目录 |
| 新增 | packages/server/internal/service/ | 服务目录 |
| 新增 | packages/server/internal/api/ | API 目录 |
| 新增 | packages/server/internal/rules/ | 规则引擎目录 |
| 新增 | packages/server/pkg/ | 公共包目录 |
| 新增 | packages/server/tests/ | 测试目录 |

---

### Task T1-2: 配置系统

**一句话描述**: 实现从环境变量加载配置的功能。

#### 需求来源

- 设计文档: 第4轮 存储与异常
- 参考: packages/client/pkg/config

#### 实现范围

**数据层**:
- 模型: `Config` 结构体

**业务层**:
- 服务: `Load()` 函数

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 测试配置加载和默认值

#### 约束条件

- 前置任务: T1-1
- 技术约束: 使用环境变量，支持 .env 文件
- 边界条件: 处理缺失配置，使用合理默认值

#### 验收标准

- [ ] 实现完成: 代码已编写，无编译错误
- [ ] 测试通过: 单元测试通过
- [ ] 功能验证: 可从环境变量读取配置

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/pkg/config/config.go | 配置定义 |
| 新增 | packages/server/pkg/config/config_test.go | 单元测试 |

---

### Task T1-3: 数据库层

**一句话描述**: 实现 PostgreSQL 数据库连接和迁移。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 4.1 数据库设计

#### 实现范围

**数据层**:
- 模型: 无（数据库连接）
- 存储: `PostgresClient`, 迁移脚本

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 连接测试
- 集成测试: 迁移测试

#### 约束条件

- 前置任务: T1-1
- 技术约束: PostgreSQL, pgx 驱动
- 边界条件: 处理连接失败，支持重连

#### 验收标准

- [ ] 实现完成: 数据库连接池
- [ ] 测试通过: 连接测试通过
- [ ] 功能验证: 可执行数据库迁移

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/postgres/client.go | 数据库客户端 |
| 新增 | packages/server/internal/store/postgres/migrate.go | 迁移逻辑 |
| 新增 | packages/server/internal/store/postgres/migrations/ | 迁移文件 |
| 新增 | packages/server/tests/integration/store/postgres_test.go | 集成测试 |

---

### Task T1-4: MCP 框架

**一句话描述**: 实现 MCP 协议基础框架和 Tool 注册机制。

#### 需求来源

- 设计文档: 第2轮 接口定义 - MCP Tool 清单

#### 实现范围

**数据层**:
- 模型: `Tool`, `ToolRequest`, `ToolResponse`

**业务层**:
- 服务: `ToolRegistry`, `MCPServer`

**接口层**:
- 端点: HTTP 端点（MCP 协议）

**测试层**:
- 单元测试: Tool 注册测试
- 集成测试: HTTP 调用测试

#### 约束条件

- 前置任务: T1-1
- 技术约束: MCP 协议 + HTTP 传输
- 边界条件: 处理未知 Tool，参数验证

#### 验收标准

- [ ] 实现完成: MCP 框架可注册和调用 Tool
- [ ] 测试通过: 单元测试和集成测试通过
- [ ] 功能验证: curl 调用返回正确响应

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/mcp/server.go | MCP 服务器 |
| 新增 | packages/server/internal/mcp/registry.go | Tool 注册 |
| 新增 | packages/server/internal/mcp/types.go | MCP 类型定义 |
| 新增 | packages/server/tests/unit/mcp/registry_test.go | 单元测试 |
| 新增 | packages/server/tests/integration/mcp/server_test.go | 集成测试 |

---

## 6. Milestone 2: 战役管理

### 6.1 目标

实现战役的完整 CRUD 操作和游戏状态管理。

### 6.2 范围

- 包含: 战役创建、查询、更新、删除、摘要
- 不包含: 角色、战斗等其他实体

### 6.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T2-1 | 战役模型 | T1-3 | S |
| T2-2 | 战役存储 | T2-1 | M |
| T2-3 | 战役服务 | T2-2 | M |
| T2-4 | 战役 Tools | T2-3 | M |
| T2-5 | 游戏状态 | T2-4 | M |

### 6.4 验收标准

- [ ] 所有 5 个战役 Tools 实现并通过测试
- [ ] 战役级联删除正确
- [ ] 游戏状态自动创建

---

### Task T2-1: 战役模型

**一句话描述**: 定义 Campaign 和 GameState 数据结构。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.1 战役, 1.5 游戏状态

#### 实现范围

**数据层**:
- 模型: `Campaign`, `CampaignStatus`, `CampaignSettings`, `GameState`, `GameTime`
- 存储: 无

**业务层**:
- 服务: 无
- 规则: 默认值设置

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 模型方法测试

#### 约束条件

- 前置任务: T1-3
- 技术约束: UUID 使用 google/uuid
- 边界条件: 字段验证

#### 验收标准

- [ ] 实现完成: 所有结构体定义
- [ ] 测试通过: 默认值和方法测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/campaign.go | Campaign 模型 |
| 新增 | packages/server/internal/models/game_state.go | GameState 模型 |
| 新增 | packages/server/tests/unit/models/campaign_test.go | 单元测试 |

---

### Task T2-2: 战役存储

**一句话描述**: 实现战役的数据库存储操作。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 4.1.1 campaigns 表

#### 实现范围

**数据层**:
- 模型: 无（使用 T2-1 定义的模型）
- 存储: `CampaignStore` 接口和实现

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 集成测试: CRUD 操作测试

#### 约束条件

- 前置任务: T2-1
- 技术约束: PostgreSQL, JSONB 存储 settings
- 边界条件: 软删除、级联删除

#### 验收标准

- [ ] 实现完成: Create, Get, List, Update, Delete
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/interface.go | Store 接口定义 |
| 新增 | packages/server/internal/store/postgres/campaign.go | Campaign 存储 |
| 新增 | packages/server/tests/integration/store/campaign_test.go | 集成测试 |

---

### Task T2-3: 战役服务

**一句话描述**: 实现战役业务逻辑层。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.1 战役管理流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 使用 CampaignStore

**业务层**:
- 服务: `CampaignService`, 包含验证、默认值、关联创建
- 规则: 参数验证、状态转换

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 服务方法测试

#### 约束条件

- 前置任务: T2-2
- 技术约束: 依赖注入
- 边界条件: 名称验证、设置验证

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: 单元测试覆盖主要场景

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/campaign.go | Campaign 服务 |
| 新增 | packages/server/tests/unit/service/campaign_test.go | 单元测试 |

---

### Task T2-4: 战役 Tools

**一句话描述**: 实现 5 个战役相关的 MCP Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.1 战役管理 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 CampaignService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `CreateCampaignRequest`, `GetCampaignRequest` 等
- 响应: `CreateCampaignResponse`, `GetCampaignResponse` 等

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T2-3
- 技术约束: MCP 协议格式
- 边界条件: 参数验证、错误响应

#### 验收标准

- [ ] 实现完成: 5 个 Tools
- [ ] 测试通过: 集成测试通过
- [ ] 功能验证: curl 调用成功

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/campaign.go | 战役 Tools |
| 新增 | packages/server/tests/integration/tools/campaign_test.go | 集成测试 |

---

### Task T2-5: 游戏状态

**一句话描述**: 实现游戏状态存储和自动关联创建。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.5 游戏状态
- 设计文档: 第3轮 业务流程 - 3.1.1 创建战役流程

#### 实现范围

**数据层**:
- 模型: 使用 T2-1 定义的 GameState
- 存储: `GameStateStore`

**业务层**:
- 服务: GameState 创建与更新
- 规则: 自动创建规则

**接口层**:
- 端点: 无（通过其他 Tools 间接操作）

**测试层**:
- 集成测试: 状态创建和更新测试

#### 约束条件

- 前置任务: T2-4
- 技术约束: 与 Campaign 一对一关系
- 边界条件: 创建战役时自动创建 GameState

#### 验收标准

- [ ] 实现完成: GameStateStore
- [ ] 测试通过: 集成测试通过
- [ ] 功能验证: 创建战役时 GameState 自动创建

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/postgres/game_state.go | GameState 存储 |
| 新增 | packages/server/tests/integration/store/game_state_test.go | 集成测试 |

---

## 7. Milestone 3: 角色管理

### 7.1 目标

实现角色和 NPC 的创建、管理功能，包含 FVTT 导入所需的扩展字段。

### 7.2 范围

- 包含: 角色 CRUD、属性计算、NPC 管理、扩展字段（法术、装备槽位、货币等）
- 不包含: 战斗中的角色操作、FVTT 导入逻辑（在 M9）

### 7.3 FVTT 影响说明

| 变更项 | 原设计 | 新设计 |
|--------|--------|--------|
| Speed | `int` | `*Speed` 结构体（多类型移动） |
| Skills | `map[string]int` | `map[string]*Skill`（含熟练/专精） |
| Saves | `map[string]int` | `map[string]*Save`（含熟练） |
| Equipment | `[]Equipment` | `*EquipmentSlots`（槽位系统） |
| Inventory | `[]Item` | `[]InventoryItem`（扩展结构） |
| 新增字段 | - | Image, Experience, DeathSaves, Proficiency, Currency, Spells, Features, Biography, Traits, ImportMeta |

### 7.4 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T3-1 | 角色模型（核心） | T2-5 | M |
| T3-2 | 角色存储 | T3-1 | M |
| T3-3 | 角色服务 | T3-2 | M |
| T3-4 | 角色 Tools | T3-3 | M |
| T3-5 | NPC 管理 | T3-4 | S |
| T3-6 | 角色扩展字段 | T3-5 | L |

### 7.5 验收标准

- [ ] 所有 5 个角色 Tools 实现并通过测试
- [ ] 属性修正计算正确
- [ ] NPC 创建和管理正常
- [ ] 扩展字段存储正确
- [ ] 装备槽位系统可用

---

### Task T3-1: 角色模型（核心）

**一句话描述**: 定义 Character 核心数据结构（基础字段）。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.2 角色
- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md)

#### 实现范围

**数据层**:
- 模型: `Character`（核心字段）, `NPCType`, `Abilities`, `HP`, `Condition`
- 存储: 无

**业务层**:
- 服务: 无
- 规则: 属性修正计算

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 模型方法、属性修正计算

#### 约束条件

- 前置任务: T2-5
- 技术约束: JSONB 存储复杂结构
- 边界条件: HP 范围验证

#### 验收标准

- [ ] 实现完成: 核心结构体定义
- [ ] 测试通过: 属性修正计算正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/character.go | Character 模型（核心） |
| 新增 | packages/server/internal/rules/ability.go | 属性修正规则 |
| 新增 | packages/server/tests/unit/rules/ability_test.go | 规则测试 |

---

### Task T3-2: 角色存储

**一句话描述**: 实现角色的数据库存储操作。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 4.1.1 characters 表

#### 实现范围

**数据层**:
- 模型: 无
- 存储: `CharacterStore` 接口和实现

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 集成测试: CRUD 操作测试

#### 约束条件

- 前置任务: T3-1
- 技术约束: 外键关联 Campaign
- 边界条件: is_npc 筛选

#### 验收标准

- [ ] 实现完成: Create, Get, Update, Delete, List
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/postgres/character.go | Character 存储 |
| 新增 | packages/server/tests/integration/store/character_test.go | 集成测试 |

---

### Task T3-3: 角色服务

**一句话描述**: 实现角色业务逻辑层。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.2 角色管理流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 使用 CharacterStore

**业务层**:
- 服务: `CharacterService`, 包含验证、默认值生成、HP 管理
- 规则: 属性生成、AC 计算、HP 变化

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 服务方法测试

#### 约束条件

- 前置任务: T3-2
- 技术约束: 属性自动生成规则
- 边界条件: HP 变化、临时 HP

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: HP 管理正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/character.go | Character 服务 |
| 新增 | packages/server/tests/unit/service/character_test.go | 单元测试 |

---

### Task T3-4: 角色 Tools

**一句话描述**: 实现 5 个角色相关的 MCP Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.2 角色管理 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 CharacterService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `CreateCharacterRequest`, `UpdateCharacterRequest` 等
- 响应: `CreateCharacterResponse`, `UpdateCharacterResponse` 等

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T3-3
- 技术约束: MCP 协议格式
- 边界条件: 参数验证、错误响应

#### 验收标准

- [ ] 实现完成: 5 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/character.go | 角色 Tools |
| 新增 | packages/server/tests/integration/tools/character_test.go | 集成测试 |

---

### Task T3-5: NPC 管理

**一句话描述**: 实现 NPC 类型的特殊处理。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.2 角色 (NPCType)
- 设计文档: 第2轮 接口定义 - 2.2.1 create_character

#### 实现范围

**数据层**:
- 模型: 无（使用现有模型）
- 存储: 无

**业务层**:
- 服务: NPC 类型处理逻辑
- 规则: scripted vs generated NPC 区分

**接口层**:
- 端点: 无（通过现有 Tools）

**测试层**:
- 单元测试: NPC 创建测试

#### 约束条件

- 前置任务: T3-4
- 技术约束: is_npc 和 npc_type 字段
- 边界条件: 玩家角色不能设置 npc_type

#### 验收标准

- [ ] 实现完成: NPC 类型逻辑
- [ ] 测试通过: 筛选功能正常

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/service/character.go | 添加 NPC 逻辑 |
| 新增 | packages/server/tests/unit/service/npc_test.go | NPC 测试 |

---

### Task T3-6: 角色扩展字段

**一句话描述**: 扩展 Character 模型以支持 FVTT 导入所需的完整字段。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第3章

#### 实现范围

**数据层**:
- 模型:
  - `Speed` - 多类型移动速度（walk/burrow/climb/fly/swim/hover）
  - `DeathSaves` - 死亡豁免（successes/failures）
  - `Skill` - 技能（ability/bonus/proficient/expertise）
  - `Save` - 豁免（bonus/proficient）
  - `Currency` - 货币（pp/gp/ep/sp/cp）
  - `EquipmentSlots` - 装备槽位（12个槽位 + 同调）
  - `EquipmentItem` - 装备物品（武器/护甲属性）
  - `InventoryItem` - 背包物品（数量/重量/用途）
  - `Spellbook` - 法术书（法术位/已知/准备）
  - `Spell` - 法术（等级/学派/施法/伤害）
  - `Feature` - 专长/特性
  - `Biography` - 传记
  - `Traits` - 特性/抗性/语言
  - `ImportMeta` - 导入元数据
- 存储: 更新数据库 Schema

**业务层**:
- 服务: 扩展 CharacterService 支持新字段
- 规则: 熟练加值计算、装备槽位验证

**接口层**:
- 端点: 扩展 create_character、update_character、get_character 参数/响应

**测试层**:
- 单元测试: 新字段模型方法
- 集成测试: 完整角色创建/更新

#### 约束条件

- 前置任务: T3-5
- 技术约束: JSONB 存储复杂嵌套结构，数据库迁移
- 边界条件: 装备槽位唯一性、法术位数量限制

#### 验收标准

- [ ] 实现完成: 所有扩展结构体定义
- [ ] 测试通过: 单元测试 + 集成测试通过
- [ ] 功能验证: 可创建包含扩展字段的角色
- [ ] 数据库迁移: 迁移脚本执行成功

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/models/character.go | 扩展 Character 模型 |
| 新增 | packages/server/internal/models/equipment.go | 装备槽位模型 |
| 新增 | packages/server/internal/models/spell.go | 法术模型 |
| 新增 | packages/server/internal/models/import_meta.go | 导入元数据 |
| 修改 | packages/server/internal/store/postgres/character.go | 更新存储逻辑 |
| 新增 | packages/server/internal/store/postgres/migrations/003_character_extend.sql | 数据库迁移 |
| 修改 | packages/server/internal/service/character.go | 扩展服务方法 |
| 修改 | packages/server/internal/api/tools/character.go | 扩展 Tools 参数 |
| 新增 | packages/server/tests/unit/models/equipment_test.go | 装备测试 |
| 新增 | packages/server/tests/unit/models/spell_test.go | 法术测试 |

---

## 8. Milestone 4: 骰子系统

### 8.1 目标

实现骰子投掷、属性检定和豁免检定功能。

### 8.2 范围

- 包含: 骰子公式解析、投掷、检定
- 不包含: 战斗中的攻击骰（在 M5）

### 8.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T4-1 | 骰子模型 | T3-5 | M |
| T4-2 | 骰子服务 | T4-1 | M |
| T4-3 | 骰子 Tools | T4-2 | M |

### 8.4 验收标准

- [ ] 所有 3 个骰子 Tools 实现并通过测试
- [ ] 骰子公式解析正确
- [ ] 暴击/大失败检测正确

---

### Task T4-1: 骰子模型

**一句话描述**: 定义骰子结果和检定结果数据结构。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.7 骰子结果
- 设计文档: 第3轮 业务流程 - 3.6 骰子系统流程

#### 实现范围

**数据层**:
- 模型: `DiceResult`, `CritStatus`, `CheckResult`
- 存储: 无

**业务层**:
- 服务: 无
- 规则: 公式解析器

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 公式解析测试

#### 约束条件

- 前置任务: T3-5
- 技术约束: 支持多种公式格式
- 边界条件: 无效公式、骰子数量限制

#### 验收标准

- [ ] 实现完成: 所有结构体和解析器
- [ ] 测试通过: 公式解析正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/dice.go | 骰子模型 |
| 新增 | packages/server/internal/rules/dice/parser.go | 公式解析 |
| 新增 | packages/server/tests/unit/rules/dice/parser_test.go | 解析测试 |

---

### Task T4-2: 骰子服务

**一句话描述**: 实现骰子投掷和检定逻辑。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.6.2 检定流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 无

**业务层**:
- 服务: `DiceService`, 包含投掷、检定、豁免
- 规则: 优势/劣势、暴击检测、属性修正

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 投掷、检定测试

#### 约束条件

- 前置任务: T4-1
- 技术约束: 随机数生成可测试
- 边界条件: 角色不存在

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: 暴击检测正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/dice.go | 骰子服务 |
| 新增 | packages/server/internal/rules/dice/roller.go | 投掷逻辑 |
| 新增 | packages/server/tests/unit/service/dice_test.go | 单元测试 |

---

### Task T4-3: 骰子 Tools

**一句话描述**: 实现 3 个骰子相关的 MCP Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.3 骰子/检定 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 DiceService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `RollDiceRequest`, `RollCheckRequest`, `RollSaveRequest`
- 响应: `RollDiceResponse`, `RollCheckResponse`, `RollSaveResponse`

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T4-2
- 技术约束: MCP 协议格式
- 边界条件: 参数验证、角色验证

#### 验收标准

- [ ] 实现完成: 3 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/dice.go | 骰子 Tools |
| 新增 | packages/server/tests/integration/tools/dice_test.go | 集成测试 |

---

## 9. Milestone 5: 战斗系统

### 9.1 目标

实现完整的战斗流程，包括开始、攻击、施法、回合管理和结束。

### 9.2 范围

- 包含: 战斗初始化、攻击、施法、回合推进、战斗结算
- 不包含: 战斗地图移动（在 M6）

### 9.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T5-1 | 战斗模型 | T4-3 | M |
| T5-2 | 战斗存储 | T5-1 | M |
| T5-3 | 战斗服务 | T5-2 | L |
| T5-4 | 战斗 Tools | T5-3 | M |
| T5-5 | 回合管理 | T5-4 | M |
| T5-6 | 战斗结算 | T5-5 | M |

### 9.4 验收标准

- [ ] 所有 6 个战斗 Tools 实现并通过测试
- [ ] 先攻顺序正确
- [ ] 攻击命中/伤害计算正确
- [ ] E2E 战斗流程测试通过

---

### Task T5-1: 战斗模型

**一句话描述**: 定义 Combat 及相关数据结构。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.3 战斗

#### 实现范围

**数据层**:
- 模型: `Combat`, `CombatStatus`, `Participant`, `Position`, `CombatLogEntry`
- 存储: 无

**业务层**:
- 服务: 无
- 规则: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 模型方法测试

#### 约束条件

- 前置任务: T4-3
- 技术约束: JSONB 存储复杂结构
- 边界条件: 参战者数量限制

#### 验收标准

- [ ] 实现完成: 所有结构体定义
- [ ] 测试通过: 模型测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/combat.go | Combat 模型 |
| 新增 | packages/server/tests/unit/models/combat_test.go | 单元测试 |

---

### Task T5-2: 战斗存储

**一句话描述**: 实现战斗的数据库存储操作。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 4.1.1 combats 表

#### 实现范围

**数据层**:
- 模型: 无
- 存储: `CombatStore` 接口和实现

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 集成测试: CRUD 操作测试

#### 约束条件

- 前置任务: T5-1
- 技术约束: 外键关联 Campaign
- 边界条件: 状态筛选

#### 验收标准

- [ ] 实现完成: Create, Get, Update, GetActive
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/postgres/combat.go | Combat 存储 |
| 新增 | packages/server/tests/integration/store/combat_test.go | 集成测试 |

---

### Task T5-3: 战斗服务

**一句话描述**: 实现战斗业务逻辑层。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.3 战斗系统流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 使用 CombatStore, CharacterStore

**业务层**:
- 服务: `CombatService`, 包含开始战斗、攻击、施法
- 规则: 先攻投掷、命中判定、伤害计算

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 服务方法测试

#### 约束条件

- 前置任务: T5-2
- 技术约束: 事务处理
- 边界条件: 回合验证、目标验证

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: 攻击逻辑正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/combat.go | Combat 服务 |
| 新增 | packages/server/internal/rules/combat/attack.go | 攻击规则 |
| 新增 | packages/server/tests/unit/service/combat_test.go | 单元测试 |

---

### Task T5-4: 战斗 Tools

**一句话描述**: 实现 4 个战斗相关的 MCP Tools（start, get_state, attack, cast_spell）。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.4 战斗操作 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 CombatService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `StartCombatRequest`, `AttackRequest` 等
- 响应: `StartCombatResponse`, `AttackResponse` 等

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T5-3
- 技术约束: MCP 协议格式
- 边界条件: 参数验证、回合验证

#### 验收标准

- [ ] 实现完成: 4 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/combat.go | 战斗 Tools |
| 新增 | packages/server/tests/integration/tools/combat_test.go | 集成测试 |

---

### Task T5-5: 回合管理

**一句话描述**: 实现回合推进逻辑和 end_turn Tool。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.3.3 结束回合流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 无

**业务层**:
- 服务: 回合推进、状态效果处理
- 规则: 状态效果持续时间

**接口层**:
- 端点: end_turn Tool
- 请求: `EndTurnRequest`
- 响应: `EndTurnResponse`

**测试层**:
- 单元测试: 回合推进测试
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T5-4
- 技术约束: 回合循环逻辑
- 边界条件: 最后一人结束回合

#### 验收标准

- [ ] 实现完成: end_turn Tool
- [ ] 测试通过: 回合循环正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/service/combat.go | 添加回合逻辑 |
| 修改 | packages/server/internal/api/tools/combat.go | 添加 end_turn |
| 新增 | packages/server/tests/unit/service/turn_test.go | 回合测试 |

---

### Task T5-6: 战斗结算

**一句话描述**: 实现战斗结束和统计逻辑。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.3.4 结束战斗流程

#### 实现范围

**数据层**:
- 模型: `CombatSummaryResult`, `ParticipantSummary`
- 存储: 无

**业务层**:
- 服务: 战斗统计计算
- 规则: 无

**接口层**:
- 端点: end_combat Tool
- 请求: `EndCombatRequest`
- 响应: `EndCombatResponse`

**测试层**:
- 单元测试: 统计计算测试
- E2E 测试: 完整战斗流程

#### 约束条件

- 前置任务: T5-5
- 技术约束: 无
- 边界条件: 战斗日志解析

#### 验收标准

- [ ] 实现完成: end_combat Tool
- [ ] 测试通过: 统计正确
- [ ] E2E: 完整战斗流程通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/service/combat.go | 添加结算逻辑 |
| 修改 | packages/server/internal/api/tools/combat.go | 添加 end_combat |
| 新增 | packages/server/tests/e2e/combat_flow_test.go | E2E 测试 |

---

## 10. Milestone 6: 地图系统

### 10.1 目标

实现大地图和战斗地图的完整功能，包含 FVTT 导入所需的扩展字段。

### 10.2 范围

- 包含: 地图数据、地点、Token、移动、扩展字段（图片、墙体）
- 不包含: 地图编辑

### 10.3 FVTT 影响说明

| 变更项 | 原设计 | 新设计 |
|--------|--------|--------|
| Map.Image | 无 | `*MapImage`（Base64/URL + 尺寸） |
| Map.Walls | 无 | `[]Wall`（墙体数据） |
| Map.ImportMeta | 无 | `*ImportMeta`（导入元数据） |
| Token 扩展 | 基础字段 | +Name, Image, Width, Height, Rotation, Scale, Alpha, ActorLink, Disposition, Hidden, Locked, Bar1, Bar2 |

### 10.4 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T6-1 | 地图模型（核心） | T4-3 | M |
| T6-2 | 地图存储 | T6-1 | M |
| T6-3 | 地图服务 | T6-2 | M |
| T6-4 | 地图 Tools | T6-3 | M |
| T6-5 | Token 移动 | T6-4 | M |
| T6-6 | 地图切换 | T6-5 | M |
| T6-7 | 地图扩展字段 | T6-6 | M |

### 10.5 验收标准

- [ ] 所有 6 个地图 Tools 实现并通过测试
- [ ] 大地图移动正确
- [ ] 战斗地图 Token 移动正确
- [ ] 地图切换正确
- [ ] 扩展字段存储正确

---

### Task T6-1: 地图模型

**一句话描述**: 定义 Map 及相关数据结构。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.4 地图

#### 实现范围

**数据层**:
- 模型: `Map`, `MapType`, `Grid`, `CellType`, `Location`, `Token`, `TokenSize`
- 存储: 无

**业务层**:
- 服务: 无
- 规则: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 模型方法测试

#### 约束条件

- 前置任务: T4-3
- 技术约束: JSONB 存储格子数据
- 边界条件: 地图尺寸限制

#### 验收标准

- [ ] 实现完成: 所有结构体定义
- [ ] 测试通过: 模型测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/map.go | Map 模型 |
| 新增 | packages/server/tests/unit/models/map_test.go | 单元测试 |

---

### Task T6-2: 地图存储

**一句话描述**: 实现地图的数据库存储操作。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 4.1.1 maps 表

#### 实现范围

**数据层**:
- 模型: 无
- 存储: `MapStore` 接口和实现

**业务层**:
- 服务: 无

**接口层**:
- 端点: 无

**测试层**:
- 集成测试: CRUD 操作测试

#### 约束条件

- 前置任务: T6-1
- 技术约束: 自引用 parent_id
- 边界条件: 类型筛选

#### 验收标准

- [ ] 实现完成: Create, Get, Update, GetWorld, GetBattle
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/store/postgres/map.go | Map 存储 |
| 新增 | packages/server/tests/integration/store/map_test.go | 集成测试 |

---

### Task T6-3: 地图服务

**一句话描述**: 实现地图业务逻辑层。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.4 地图系统流程

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 使用 MapStore, GameStateStore

**业务层**:
- 服务: `MapService`, 包含获取地图、计算旅行
- 规则: 旅行时间计算、移动消耗

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 服务方法测试

#### 约束条件

- 前置任务: T6-2
- 技术约束: 游戏时间推进
- 边界条件: 位置验证

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: 旅行计算正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/map.go | Map 服务 |
| 新增 | packages/server/internal/rules/movement/travel.go | 旅行规则 |
| 新增 | packages/server/tests/unit/service/map_test.go | 单元测试 |

---

### Task T6-4: 地图 Tools

**一句话描述**: 实现 get_world_map 和 move_to Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.5 地图/移动 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 MapService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `GetWorldMapRequest`, `MoveToRequest`
- 响应: `GetWorldMapResponse`, `MoveToResponse`

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T6-3
- 技术约束: MCP 协议格式
- 边界条件: 参数验证

#### 验收标准

- [ ] 实现完成: 2 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/map.go | 地图 Tools (1/3) |
| 新增 | packages/server/tests/integration/tools/map_test.go | 集成测试 |

---

### Task T6-5: Token 移动

**一句话描述**: 实现战斗地图中的 Token 移动。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.4.3 移动 Token 流程

#### 实现范围

**数据层**:
- 模型: `MoveTokenRequest`, `MoveTokenResponse`
- 存储: 无

**业务层**:
- 服务: Token 移动验证和执行
- 规则: 移动力消耗、困难地形

**接口层**:
- 端点: move_token Tool
- 请求: `MoveTokenRequest`
- 响应: `MoveTokenResponse`

**测试层**:
- 单元测试: 移动规则测试
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T6-4
- 技术约束: 格子计算
- 边界条件: 墙壁阻挡、移动力不足

#### 验收标准

- [ ] 实现完成: move_token Tool
- [ ] 测试通过: 移动规则正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/service/map.go | 添加 Token 移动 |
| 修改 | packages/server/internal/api/tools/map.go | 添加 move_token |
| 新增 | packages/server/internal/rules/movement/token.go | Token 移动规则 |

---

### Task T6-6: 地图切换

**一句话描述**: 实现大地图和战斗地图的切换。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.4.2 进入战斗地图流程

#### 实现范围

**数据层**:
- 模型: `EnterBattleMapRequest/Response`, `ExitBattleMapRequest/Response`, `GetBattleMapRequest/Response`
- 存储: 无

**业务层**:
- 服务: 地图切换逻辑
- 规则: Token 放置

**接口层**:
- 端点: enter_battle_map, get_battle_map, exit_battle_map Tools
- 请求: 各自请求结构体
- 响应: 各自响应结构体

**测试层**:
- 集成测试: 切换流程测试
- E2E 测试: 冒险流程

#### 约束条件

- 前置任务: T6-5
- 技术约束: GameState 更新
- 边界条件: 当前地图类型检查

#### 验收标准

- [ ] 实现完成: 3 个 Tools
- [ ] 测试通过: 切换流程正确
- [ ] E2E: 冒险流程通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/service/map.go | 添加切换逻辑 |
| 修改 | packages/server/internal/api/tools/map.go | 添加切换 Tools |
| 新增 | packages/server/tests/e2e/adventure_flow_test.go | E2E 测试 |

---

### Task T6-7: 地图扩展字段

**一句话描述**: 扩展 Map 和 Token 模型以支持 FVTT 导入所需的完整字段。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第4章、第5章

#### 实现范围

**数据层**:
- 模型:
  - `MapImage` - 地图图片（Base64/URL + 尺寸 + 格式）
  - `Wall` - 墙体（坐标 + 类型 + 属性）
  - `TokenBar` - Token 血条（attribute + value + max）
  - 扩展 `Map`: Image, Walls, ImportMeta
  - 扩展 `Token`: Name, Image, Width, Height, Rotation, Scale, Alpha, ActorLink, Disposition, Hidden, Locked, Bar1, Bar2
- 存储: 更新数据库 Schema

**业务层**:
- 服务: 扩展 MapService 支持新字段
- 规则: 墙体阻挡验证

**接口层**:
- 端点: 扩展 get_battle_map 响应

**测试层**:
- 单元测试: 新字段模型方法
- 集成测试: 完整地图创建/更新

#### 约束条件

- 前置任务: T6-6
- 技术约束: JSONB 存储图片数据，数据库迁移
- 边界条件: 图片大小限制、墙体坐标验证

#### 验收标准

- [ ] 实现完成: 所有扩展结构体定义
- [ ] 测试通过: 单元测试 + 集成测试通过
- [ ] 功能验证: 可创建包含图片和墙体的地图
- [ ] 数据库迁移: 迁移脚本执行成功

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/models/map.go | 扩展 Map 模型 |
| 新增 | packages/server/internal/models/wall.go | 墙体模型 |
| 新增 | packages/server/internal/models/token_bar.go | Token 血条模型 |
| 修改 | packages/server/internal/store/postgres/map.go | 更新存储逻辑 |
| 新增 | packages/server/internal/store/postgres/migrations/004_map_extend.sql | 数据库迁移 |
| 修改 | packages/server/internal/service/map.go | 扩展服务方法 |
| 修改 | packages/server/internal/api/tools/map.go | 扩展 Tools 响应 |
| 新增 | packages/server/tests/unit/models/wall_test.go | 墙体测试 |

---

## 11. Milestone 7: 上下文管理

### 11.1 目标

实现对话存储和上下文压缩功能。

### 11.2 范围

- 包含: 消息存储、上下文获取、滑动窗口压缩
- 不包含: 高级压缩（摘要策略）

### 11.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T7-1 | 消息模型 | T2-5 | S |
| T7-2 | 上下文服务 | T7-1 | M |
| T7-3 | 上下文 Tools | T7-2 | M |

### 11.4 验收标准

- [ ] 所有 3 个上下文 Tools 实现并通过测试
- [ ] 滑动窗口压缩正确
- [ ] 原始数据获取正确

---

### Task T7-1: 消息模型

**一句话描述**: 定义 Message 数据结构和存储。

#### 需求来源

- 设计文档: 第1轮 数据结构 - 1.6 消息

#### 实现范围

**数据层**:
- 模型: `Message`, `MessageRole`, `ToolCall`, `ToolResult`
- 存储: `MessageStore`

**业务层**:
- 服务: 无
- 规则: 无

**接口层**:
- 端点: 无

**测试层**:
- 集成测试: 存储测试

#### 约束条件

- 前置任务: T2-5
- 技术约束: JSONB 存储 tool_calls
- 边界条件: 消息长度限制

#### 验收标准

- [ ] 实现完成: 模型和存储
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/message.go | Message 模型 |
| 新增 | packages/server/internal/store/postgres/message.go | Message 存储 |
| 新增 | packages/server/tests/integration/store/message_test.go | 集成测试 |

---

### Task T7-2: 上下文服务

**一句话描述**: 实现上下文构建和压缩逻辑。

#### 需求来源

- 设计文档: 第3轮 业务流程 - 3.5 上下文管理流程

#### 实现范围

**数据层**:
- 模型: `Context`, `GameSummary`, `PartyMember`, `CombatSummary`
- 存储: 使用 MessageStore, GameStateStore

**业务层**:
- 服务: `ContextService`, 包含获取上下文、压缩
- 规则: 滑动窗口（默认 20 条）

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 压缩逻辑测试

#### 约束条件

- 前置任务: T7-1
- 技术约束: Token 估算
- 边界条件: 消息数量不足窗口大小

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: 压缩正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/context.go | Context 模型 |
| 新增 | packages/server/internal/service/context.go | Context 服务 |
| 新增 | packages/server/tests/unit/service/context_test.go | 单元测试 |

---

### Task T7-3: 上下文 Tools

**一句话描述**: 实现 3 个上下文相关的 MCP Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.7 上下文管理 Tools

#### 实现范围

**数据层**:
- 模型: 请求/响应结构体
- 存储: 无

**业务层**:
- 服务: 调用 ContextService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `GetContextRequest`, `GetRawContextRequest`, `SaveMessageRequest`
- 响应: `GetContextResponse`, `GetRawContextResponse`, `SaveMessageResponse`

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T7-2
- 技术约束: MCP 协议格式
- 边界条件: 参数验证

#### 验收标准

- [ ] 实现完成: 3 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/context.go | 上下文 Tools |
| 新增 | packages/server/tests/integration/tools/context_test.go | 集成测试 |

---

## 12. Milestone 8: 规则查询

### 12.1 目标

实现 RAG 集成和规则查询功能。

### 12.2 范围

- 包含: 法术、物品、怪物查询
- 不包含: RAG 数据导入

### 12.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T8-1 | 查询模型 | T1-4 | S |
| T8-2 | 查询服务 | T8-1 | M |
| T8-3 | 查询 Tools | T8-2 | M |

### 12.4 验收标准

- [ ] 所有 3 个查询 Tools 实现并通过测试
- [ ] RAG 服务集成正确

---

### Task T8-1: 查询模型

**一句话描述**: 定义查询请求和响应结构。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.6 规则查询 Tools

#### 实现范围

**数据层**:
- 模型: 查询请求/响应结构体
- 存储: 无（外部 RAG）

**业务层**:
- 服务: 无
- 规则: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 无

#### 约束条件

- 前置任务: T1-4
- 技术约束: RAG 接口定义
- 边界条件: 无

#### 验收标准

- [ ] 实现完成: 结构体定义

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/models/lookup.go | 查询模型 |

---

### Task T8-2: 查询服务

**一句话描述**: 实现 RAG 集成和查询逻辑。

#### 需求来源

- 设计文档: 第4轮 存储与异常 - 错误码 L501

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 无

**业务层**:
- 服务: `LookupService`, RAG 客户端调用
- 规则: 无

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 服务测试（Mock RAG）

#### 约束条件

- 前置任务: T8-1
- 技术约束: HTTP 调用 RAG 服务
- 边界条件: RAG 服务不可用

#### 验收标准

- [ ] 实现完成: 所有服务方法
- [ ] 测试通过: Mock 测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/service/lookup.go | 查询服务 |
| 新增 | packages/server/internal/rag/client.go | RAG 客户端接口 |
| 新增 | packages/server/tests/unit/service/lookup_test.go | 单元测试 |

---

### Task T8-3: 查询 Tools

**一句话描述**: 实现 3 个规则查询 MCP Tools。

#### 需求来源

- 设计文档: 第2轮 接口定义 - 2.6 规则查询 Tools

#### 实现范围

**数据层**:
- 模型: 无
- 存储: 无

**业务层**:
- 服务: 调用 LookupService
- 规则: 无

**接口层**:
- 端点: MCP Tool 端点
- 请求: `LookupSpellRequest`, `LookupItemRequest`, `LookupMonsterRequest`
- 响应: `LookupSpellResponse`, `LookupItemResponse`, `LookupMonsterResponse`

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T8-2
- 技术约束: MCP 协议格式
- 边界条件: 查询为空、服务不可用

#### 验收标准

- [ ] 实现完成: 3 个 Tools
- [ ] 测试通过: 集成测试通过

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/lookup.go | 查询 Tools |
| 新增 | packages/server/tests/integration/tools/lookup_test.go | 集成测试 |

---

## 13. Milestone 9: 导入功能

### 13.1 目标

实现 FVTT/UVTT 格式导入功能，支持地图和角色数据的迁移。

### 13.2 范围

- 包含: UVTT 地图导入、FVTT Scene 导入、FVTT Actor 导入、FVTT Item 导入
- 不包含: FVTT 导出、光源系统、ActiveEffect、非 dnd5e 游戏系统

### 13.3 模块结构

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

### 13.4 任务清单

| ID | 任务名称 | 依赖 | 复杂度 |
|----|----------|------|--------|
| T9-1 | Import 模块框架 | T3-6, T6-7 | M |
| T9-2 | UVTT Parser | T9-1 | M |
| T9-3 | FVTT Scene Parser | T9-1 | M |
| T9-4 | Map Converter | T9-2, T9-3 | M |
| T9-5 | import_map Tool | T9-4 | M |
| T9-6 | FVTT Actor Parser | T9-1 | L |
| T9-7 | Character Converter | T9-6, T3-6 | L |
| T9-8 | import_character Tool | T9-7 | M |
| T9-9 | Item Converter | T9-6 | M |
| T9-10 | import_items Tool | T9-9 | M |

### 13.5 验收标准

- [ ] import_map Tool 实现并通过测试（UVTT + FVTT Scene）
- [ ] import_character Tool 实现并通过测试（FVTT Actor）
- [ ] import_items Tool 实现并通过测试（FVTT Items）
- [ ] 导入后角色可参与战斗
- [ ] 导入后地图可正常显示和移动 Token
- [ ] 警告信息正确返回（未支持字段）

---

### Task T9-1: Import 模块框架

**一句话描述**: 创建 Import 模块的基础框架和接口定义。

#### 需求来源

- 设计文档: [设计-MCP-Server-游戏规则引擎-v2.md](./设计-MCP-Server-游戏规则引擎-v2.md) - 第2.4节

#### 实现范围

**数据层**:
- 模型: `ImportFormat`, `ImportOptions`, `ImportResult`
- 存储: 无

**业务层**:
- 服务: `ImportService`（流程编排）
- 规则: 格式检测

**接口层**:
- 端点: 无（框架层）

**测试层**:
- 单元测试: 格式检测测试

#### 约束条件

- 前置任务: T3-6, T6-7
- 技术约束: Parser/Converter 分离架构
- 边界条件: 格式自动检测

#### 验收标准

- [ ] 实现完成: 模块目录结构和接口定义
- [ ] 测试通过: 格式检测逻辑正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/interface.go | 接口定义 |
| 新增 | packages/server/internal/import/service.go | 导入服务 |
| 新增 | packages/server/internal/import/format/uvtt.go | UVTT 格式定义 |
| 新增 | packages/server/internal/import/format/fvtt.go | FVTT 格式定义 |
| 新增 | packages/server/tests/unit/import/format_test.go | 格式测试 |

---

### Task T9-2: UVTT Parser

**一句话描述**: 实现 UVTT 格式解析器。

#### 需求来源

- 研究文档: [FVTT-UVTT-格式研究报告.md](./research/FVTT-UVTT-格式研究报告.md)

#### 实现范围

**数据层**:
- 模型: `UVTTData`, `UVTTWall`, `UVTTLineOfSight`
- 存储: 无

**业务层**:
- 服务: `UVTTParser.CanParse()`, `UVTTParser.Parse()`
- 规则: resolution 字段检测

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 解析测试
- 集成测试: 真实 UVTT 文件解析

#### 约束条件

- 前置任务: T9-1
- 技术约束: UVTT 2.x 格式
- 边界条件: 无效格式、缺失字段

#### 验收标准

- [ ] 实现完成: UVTT 解析器
- [ ] 测试通过: 真实 UVTT 文件解析正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/parser/parser.go | Parser 接口 |
| 新增 | packages/server/internal/import/parser/uvtt_parser.go | UVTT 解析器 |
| 新增 | packages/server/tests/unit/import/uvtt_parser_test.go | 单元测试 |
| 新增 | packages/server/tests/testdata/sample.uvtt | 测试数据 |

---

### Task T9-3: FVTT Scene Parser

**一句话描述**: 实现 FVTT Scene 格式解析器。

#### 需求来源

- 研究文档: [FVTT-UVTT-格式研究报告.md](./research/FVTT-UVTT-格式研究报告.md)

#### 实现范围

**数据层**:
- 模型: `FVTTScene`, `FVTTFolder`, `FVTTJournal`
- 存储: 无

**业务层**:
- 服务: `FVTTSceneParser.CanParse()`, `FVTTSceneParser.Parse()`
- 规则: _id, grid 字段检测

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 解析测试
- 集成测试: 真实 FVTT Scene 文件解析

#### 约束条件

- 前置任务: T9-1
- 技术约束: FVTT Scene JSON 格式
- 边界条件: 无效格式、缺失字段

#### 验收标准

- [ ] 实现完成: FVTT Scene 解析器
- [ ] 测试通过: 真实 FVTT Scene 文件解析正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/parser/fvtt_scene_parser.go | FVTT Scene 解析器 |
| 新增 | packages/server/tests/unit/import/fvtt_scene_parser_test.go | 单元测试 |
| 新增 | packages/server/tests/testdata/sample_scene.json | 测试数据 |

---

### Task T9-4: Map Converter

**一句话描述**: 实现地图数据转换器（UVTT + FVTT Scene → Map）。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第9.2节

#### 实现范围

**数据层**:
- 模型: 无（使用现有 Map 模型）
- 存储: 无

**业务层**:
- 服务: `MapConverter.Convert()`, `MapConverter.Validate()`
- 规则: 字段映射、坐标转换

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 字段映射测试
- 集成测试: 完整转换流程

#### 约束条件

- 前置任务: T9-2, T9-3
- 技术约束: 使用 T6-7 定义的扩展 Map 模型
- 边界条件: 坐标系统转换

#### 验收标准

- [ ] 实现完成: Map Converter
- [ ] 测试通过: UVTT 和 FVTT Scene 转换正确
- [ ] 功能验证: 转换后地图可正常使用

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/converter/map_converter.go | 地图转换器 |
| 新增 | packages/server/internal/import/validator/validator.go | 数据校验 |
| 新增 | packages/server/tests/unit/import/map_converter_test.go | 单元测试 |

---

### Task T9-5: import_map Tool

**一句话描述**: 实现 import_map MCP Tool。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第6.1节

#### 实现范围

**数据层**:
- 模型: `ImportMapRequest`, `ImportMapResponse`
- 存储: 调用 MapStore

**业务层**:
- 服务: 调用 ImportService.ImportMap()
- 规则: 格式自动检测

**接口层**:
- 端点: import_map Tool
- 请求: campaign_id, data, format(可选), name(可选)
- 响应: map, detected_format, warnings

**测试层**:
- 集成测试: Tool 调用测试
- E2E 测试: 完整导入流程

#### 约束条件

- 前置任务: T9-4
- 技术约束: MCP 协议格式
- 边界条件: 大文件处理、图片存储

#### 验收标准

- [ ] 实现完成: import_map Tool
- [ ] 测试通过: UVTT 和 FVTT Scene 导入成功
- [ ] 功能验证: 导入后地图可正常显示和移动 Token

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/api/tools/import.go | 导入 Tools |
| 修改 | packages/server/internal/import/service.go | 添加 ImportMap 方法 |
| 新增 | packages/server/tests/integration/tools/import_map_test.go | 集成测试 |
| 新增 | packages/server/tests/e2e/import_flow_test.go | E2E 测试 |

---

### Task T9-6: FVTT Actor Parser

**一句话描述**: 实现 FVTT Actor 格式解析器。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第9.1节
- 研究文档: [FVTT-UVTT-格式研究报告.md](./research/FVTT-UVTT-格式研究报告.md)

#### 实现范围

**数据层**:
- 模型: `FVTTActor`, `FVTTAbilities`, `FVTTAttributes`, `FVTTDetails`, `FVTTSkills`, `FVTTTraits`, `FVTTCurrency`
- 存储: 无

**业务层**:
- 服务: `FVTTActorParser.CanParse()`, `FVTTActorParser.Parse()`
- 规则: _id, system.abilities 字段检测

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 解析测试
- 集成测试: 真实 FVTT Actor 文件解析

#### 约束条件

- 前置任务: T9-1
- 技术约束: FVTT dnd5e system 格式
- 边界条件: 无效格式、非 dnd5e 系统

#### 验收标准

- [ ] 实现完成: FVTT Actor 解析器
- [ ] 测试通过: 真实 FVTT Actor 文件解析正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/parser/fvtt_actor_parser.go | FVTT Actor 解析器 |
| 新增 | packages/server/internal/import/format/fvtt_actor.go | FVTT Actor 格式定义 |
| 新增 | packages/server/tests/unit/import/fvtt_actor_parser_test.go | 单元测试 |
| 新增 | packages/server/tests/testdata/sample_actor.json | 测试数据 |

---

### Task T9-7: Character Converter

**一句话描述**: 实现角色数据转换器（FVTT Actor → Character）。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第9.1节

#### 实现范围

**数据层**:
- 模型: 无（使用 T3-6 定义的扩展 Character 模型）
- 存储: 无

**业务层**:
- 服务: `CharacterConverter.Convert()`, `CharacterConverter.Validate()`
- 规则:
  - abilities 映射
  - skills 转换（熟练/专精）
  - equipment 槽位分配
  - spells 法术书填充
  - ImportMeta 生成

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 字段映射测试
- 集成测试: 完整转换流程

#### 约束条件

- 前置任务: T9-6, T3-6
- 技术约束: 使用扩展 Character 模型
- 边界条件: 复杂装备映射、法术位计算

#### 验收标准

- [ ] 实现完成: Character Converter
- [ ] 测试通过: FVTT Actor 转换正确
- [ ] 功能验证: 转换后角色可参与战斗

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/converter/character_converter.go | 角色转换器 |
| 新增 | packages/server/tests/unit/import/character_converter_test.go | 单元测试 |

---

### Task T9-8: import_character Tool

**一句话描述**: 实现 import_character MCP Tool。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第6.2节

#### 实现范围

**数据层**:
- 模型: `ImportCharacterRequest`, `ImportCharacterResponse`
- 存储: 调用 CharacterStore

**业务层**:
- 服务: 调用 ImportService.ImportCharacter()
- 规则: NPC 类型判断

**接口层**:
- 端点: import_character Tool
- 请求: campaign_id, data, name(可选), as_npc(可选)
- 响应: character, warnings, skipped

**测试层**:
- 集成测试: Tool 调用测试
- E2E 测试: 完整导入流程

#### 约束条件

- 前置任务: T9-7
- 技术约束: MCP 协议格式
- 边界条件: 大文件处理、图片存储

#### 验收标准

- [ ] 实现完成: import_character Tool
- [ ] 测试通过: FVTT Actor 导入成功
- [ ] 功能验证: 导入后角色可参与战斗

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/api/tools/import.go | 添加 import_character |
| 修改 | packages/server/internal/import/service.go | 添加 ImportCharacter 方法 |
| 新增 | packages/server/tests/integration/tools/import_character_test.go | 集成测试 |

---

### Task T9-9: Item Converter

**一句话描述**: 实现物品数据转换器（FVTT Item → InventoryItem/EquipmentItem）。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第9.1节

#### 实现范围

**数据层**:
- 模型: 无（使用 T3-6 定义的 InventoryItem/EquipmentItem 模型）
- 存储: 无

**业务层**:
- 服务: `ItemConverter.Convert()`, `ItemConverter.ConvertBatch()`
- 规则:
  - weapon → EquipmentItem
  - equipment → EquipmentItem（按 armor.type 分配槽位）
  - consumable/loot/tool → InventoryItem

**接口层**:
- 端点: 无

**测试层**:
- 单元测试: 字段映射测试
- 集成测试: 批量转换流程

#### 约束条件

- 前置任务: T9-6
- 技术约束: FVTT Item 类型映射
- 边界条件: 未知物品类型

#### 验收标准

- [ ] 实现完成: Item Converter
- [ ] 测试通过: FVTT Item 转换正确

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | packages/server/internal/import/converter/item_converter.go | 物品转换器 |
| 新增 | packages/server/tests/unit/import/item_converter_test.go | 单元测试 |

---

### Task T9-10: import_items Tool

**一句话描述**: 实现 import_items MCP Tool。

#### 需求来源

- FVTT 变更: [需求变更-FVTT导入支持.md](./server-design/需求变更-FVTT导入支持.md) - 第6.3节

#### 实现范围

**数据层**:
- 模型: `ImportItemsRequest`, `ImportItemsResponse`
- 存储: 调用 CharacterStore（添加到指定角色）

**业务层**:
- 服务: 调用 ImportService.ImportItems()
- 规则: 批量处理

**接口层**:
- 端点: import_items Tool
- 请求: campaign_id, character_id(可选), data
- 响应: items, warnings

**测试层**:
- 集成测试: Tool 调用测试

#### 约束条件

- 前置任务: T9-9
- 技术约束: MCP 协议格式
- 边界条件: 批量处理、部分失败

#### 验收标准

- [ ] 实现完成: import_items Tool
- [ ] 测试通过: FVTT Items 批量导入成功

#### 文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | packages/server/internal/api/tools/import.go | 添加 import_items |
| 修改 | packages/server/internal/import/service.go | 添加 ImportItems 方法 |
| 新增 | packages/server/tests/integration/tools/import_items_test.go | 集成测试 |

---

## 14. 风险评估

| 风险 | 影响任务 | 概率 | 影响 | 缓解措施 |
|------|----------|------|------|----------|
| 骰子公式解析复杂 | T4-1, T4-2 | 中 | 中 | 先实现基本格式，逐步扩展 |
| 战斗规则复杂 | T5-3 | 高 | 高 | 分步实现，先核心后扩展 |
| RAG 服务依赖 | T8-2, T8-3 | 低 | 中 | Mock 接口，解耦依赖 |
| 并发控制复杂 | T5-3, T5-5 | 中 | 高 | 使用数据库行锁，短事务 |
| FVTT 格式变化 | T9-2, T9-3, T9-6 | 中 | 中 | 保留原始 JSON，版本检测 |
| dnd5e 字段结构复杂 | T9-7 | 高 | 高 | 核心字段优先，其他保留原始 |
| 图片存储占用空间 | T9-5, T9-8 | 中 | 中 | 压缩存储，可选外部存储 |
| 向后兼容性 | T3-6, T6-7 | 低 | 高 | 数据迁移脚本，版本控制 |

---

## 15. 执行建议

### 15.1 开发流程

1. 按里程碑顺序执行
2. 每完成一个任务运行测试验证
3. 每完成一个里程碑进行集成验收

### 15.2 验证命令

```powershell
# 运行单个任务的测试
go test -v ./tests/unit/models/...
go test -v ./tests/integration/store/...

# 运行里程碑的所有测试
go test -v ./tests/...

# E2E 测试
go test -v ./tests/e2e/... -tags=e2e

# 覆盖率
go test -cover ./...
```

### 15.3 提交策略

- 每完成一个任务提交一次
- 提交信息格式: `feat(M{里程碑}): 完成 T{任务ID} {任务名称}`

示例:
```
feat(M1): 完成 T1-1 创建项目结构
feat(M2): 完成 T2-4 战役 Tools
feat(M3): 完成 T3-6 角色扩展字段
feat(M5): 完成 T5-6 战斗结算
feat(M9): 完成 T9-8 import_character Tool
```

---

## 16. 完成检查

计划完成后，确认以下问题可回答：

- [x] 每个任务是否是完整的垂直切片？ **是**
- [x] 每个任务是否可独立验证？ **是**
- [x] 任务依赖是否清晰？ **是**
- [x] 里程碑验收标准是否明确？ **是**
- [x] 文件清单是否完整？ **是**
- [x] FVTT 导入任务是否完整？ **是（M9 含 10 个任务）**

---

**计划状态**: ✅ 已更新，包含 FVTT 导入支持

**总计**: 9 个里程碑，47 个任务，34 个 MCP Tools

**变更记录**:
- v1.0: 初始版本（8 个里程碑，35 个任务）
- v2.0: 新增 M9 导入功能，扩展 M3/M6 任务以支持 FVTT 导入（+12 任务，+4 Tools）
