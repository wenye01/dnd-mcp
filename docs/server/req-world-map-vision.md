# 大世界地图视觉理解需求规格

## 文档信息

- **创建日期**: 2026-03-01
- **状态**: 待审核
- **目标组件**: Server
- **需求类型**: 扩展/完善

---

## 1. 需求分析

### 1.1 需求类型

**类型**: 扩展 (Extension)

### 1.2 目标组件

**目标组件**: Server - MCP Server 地图系统

### 1.3 核心诉求

用户希望增强当前的大世界地图功能，使其具备类似 D&D Hub (https://www.dndhub.top/guide) 的能力：

1. **视觉理解能力**: LLM (DM) 能够"看到"地图图片，而不仅仅是依赖结构化的格子数据
2. **动态地图输入**: 支持直接上传图片作为大世界地图，而无需预先定义格子数据
3. **地点识别与标记**: 能够从地图图片中识别或手动添加地点（城镇、遗迹、营地等）
4. **玩家位置感知**: DM 能够清晰知道玩家当前在世界地图的位置
5. **剧情规划支持**: 基于地图和玩家位置，LLM 能够更好地规划后续剧情发展

### 1.4 现状 vs 期望

| 维度 | 当前设计 | 期望能力 |
|------|---------|---------|
| 地图输入 | 结构化格子数据 (Grid + Cells) | 直接图片上传 + 可选格子化 |
| LLM 地图感知 | 依赖结构化数据描述 | 直接视觉理解地图图片 |
| 地点标记 | 手动定义 Location | 视觉识别或手动标记 |
| DM 感知能力 | 纯文本数据 | 视觉 + 文本混合 |

---

## 2. 场景分析

### 2.1 场景对比

| 场景 | 当前行为 | 期望行为 |
|------|---------|---------|
| **上传世界地图** | 需要手动定义 Grid 和 Cells，为每个格子设置类型 | 直接上传一张地图图片，系统自动存储并可被视觉模型理解 |
| **DM 了解地图** | 调用 `get_world_map`，获取结构化格子和地点列表 | 调用 `get_world_map`，获取地图图片 + 地点列表 + 玩家位置 |
| **LLM 规划剧情** | 基于地点名称和坐标推断距离和关系 | 基于地图图片视觉信息 + 位置信息，理解地理关系 |
| **添加地点标记** | 手动指定坐标和名称 | 在地图图片上点击添加标记（前端）或通过描述添加（LLM） |

### 2.2 关键场景流

#### 场景 A: DM 上传新世界地图

```
用户 → 前端
  1. 上传地图图片 (JPG/PNG)
  2. 可选：输入地图名称、描述

前端 → Server
  1. create_world_map(campaign_id, image_url, name?, description?)

Server
  1. 创建 Map 记录 (type = "world")
  2. 存储图片 URL/引用
  3. 初始化空的 Locations 列表
  4. 返回 map_id

LLM (DM)
  1. 后续可调用 get_world_map 获取地图视觉信息
```

#### 场景 B: LLM 理解地图并规划剧情

```
Client → LLM
  1. 用户问："我们向北走会遇到什么？"
  2. Client 调用 get_world_map 获取地图

Server → Client
  1. 返回地图图片（或描述）
  2. 返回当前玩家位置
  3. 返回已知地点列表

LLM
  1. 视觉模型分析地图，识别北方的地理特征
  2. 结合地点信息，生成剧情建议
  3. 回复用户

Client → Server (可选)
  1. add_location(map_id, name, position, description) - LLM 发现新地点并添加
```

---

## 3. 功能需求

### 3.1 核心功能

#### F1: 地图图片存储
- **描述**: Map 模型支持关联地图图片
- **输入**: 图片 URL 或 Base64
- **输出**: Map 记录包含 Image 字段
- **优先级**: 高

#### F2: 视觉 MCP Tool
- **描述**: 提供 `analyze_map_image` Tool，返回地图的视觉描述
- **输入**: map_id
- **输出**: 地图的视觉描述文本（地形、主要区域、地理特征等）
- **优先级**: 高

#### F3: 地点标记管理
- **描述**: 支持在地图上添加/删除/查询地点标记
- **输入**: name, position (坐标或百分比), description
- **输出**: Location 记录
- **优先级**: 中

#### F4: 玩家位置追踪
- **描述**: GameState.PartyPosition 支持百分比坐标（适应不同尺寸的地图图片）
- **输入**: 新位置 (x, y 百分比)
- **输出**: 更新后的位置
- **优先级**: 中

### 3.2 扩展功能

#### F5: 地图图片预处理 (可选)
- **描述**: 自动检测地图边界、比例尺、图例等
- **依赖**: 高级视觉模型
- **优先级**: 低

#### F6: 地点自动识别 (可选)
- **描述**: 从地图图片中自动识别可能的地点标记
- **依赖**: OCR + 目标检测
- **优先级**: 低

---

## 4. 技术方案初步分析

### 4.1 方案选项

| 方案 | 描述 | 优点 | 缺点 | 复杂度 |
|------|------|------|------|--------|
| **A. 视觉 MCP Tool** | Server 提供 `analyze_map_image` Tool，调用外部视觉服务 | 无需在 Server 集成 LLM，保持架构清晰 | 需要外部视觉服务 | 低 |
| **B. VLLM 集成** | Server 直接集成 VLLM 提供视觉能力 | 性能更好，控制更精细 | 增加 Server 复杂度，违反"无 LLM 依赖"原则 | 高 |
| **C. 客户端视觉** | Client 负责调用视觉模型，Server 只存图片 | 完全符合 Server 设计理念 | 需要修改 Client 架构 | 中 |

### 4.2 推荐方案

**方案 A + C 混合**

1. **Server 职责**:
   - 存储地图图片引用 (URL)
   - 提供 `get_world_map` 返回图片 URL
   - 提供 `analyze_map_image` Tool 作为可选接口

2. **Client 职责**:
   - 获取地图图片 URL
   - 调用视觉模型（如 Claude、GPT-4V）分析地图
   - 将视觉描述作为上下文传递给 LLM

3. **可选扩展**:
   - Server 提供 `analyze_map_image` Tool，内部转发到视觉服务
   - 这样支持无视觉能力的简化 Client

### 4.3 架构影响

```
现有架构 (简化):
  Client → get_world_map → Server → 返回结构化数据

新架构 (扩展):
  Client → get_world_map → Server → 返回结构化数据 + 图片 URL
  Client → 视觉模型(图片) → 视觉描述
  Client → LLM(上下文+视觉描述) → 剧情规划

或 (使用 Server Tool):
  Client → analyze_map_image → Server → 视觉服务 → 视觉描述
  Client → LLM(上下文+视觉描述) → 剧情规划
```

---

## 5. 数据模型变更

### 5.1 Map 模型扩展

```go
// 现有 Map 模型需要添加的字段
type Map struct {
    // ... 现有字段 ...

    // 新增字段
    Image       *MapImage   `json:"image"`        // 地图图片
    ImagePrompt string      `json:"image_prompt"` // LLM 可用的图片描述（可缓存）
}

// MapImage 地图图片
type MapImage struct {
    URL     string  `json:"url"`      // 图片 URL（推荐，支持外部存储）
    Width   int     `json:"width"`    // 原始宽度（像素）
    Height  int     `json:"height"`   // 原始高度（像素）
    Format  string  `json:"format"`   // 格式 (jpg, png, webp)
}
```

### 5.2 Position 模型扩展

```go
// 现有 Position 支持绝对坐标，扩展支持百分比坐标
type Position struct {
    X     float64 `json:"x"`       // 绝对坐标（格子）或百分比 (0-100)
    Y     float64 `json:"y"`       // 绝对坐标（格子）或百分比 (0-100)
    Type  PositionType `json:"type"` // 坐标类型
}

type PositionType string

const (
    PositionTypeAbsolute  PositionType = "absolute"  // 绝对坐标（格子）
    PositionTypePercent   PositionType = "percent"   // 百分比坐标（图片）
)
```

### 5.3 新增 Tool 请求/响应

```go
// analyze_map_image Tool
type AnalyzeMapImageRequest struct {
    CampaignID string `json:"campaign_id"` // 必填
    MapID      string `json:"map_id"`      // 必填
}

type AnalyzeMapImageResponse struct {
    Description  string   `json:"description"`   // 地图视觉描述
    Features     []string `json:"features"`      // 识别的特征（山脉、河流等）
    Suggestions  []string `json:"suggestions"`   // 可能的地点建议
}
```

---

## 6. MCP Tools 变更

### 6.1 现有 Tools 修改

| Tool | 变更类型 | 变更说明 |
|------|----------|----------|
| `get_world_map` | 扩展响应 | 添加 Map.Image 和 ImageURL 字段 |
| `create_campaign` | 无变更 | 保持不变 |
| `move_to` | 扩展 | 支持百分比坐标 |

### 6.2 新增 Tools

| Tool | 描述 | 优先级 |
|------|------|--------|
| `analyze_map_image` | 分析地图图片，返回视觉描述 | 高 |
| `add_location` | 在地图上添加地点标记 | 中 |
| `update_location` | 更新地点信息 | 中 |
| `remove_location` | 删除地点标记 | 低 |

---

## 7. 影响范围评估

### 7.1 涉及的现有模块

| 模块 | 影响程度 | 变更说明 |
|------|----------|----------|
| `models/map.go` | 高 | 添加 Image 字段 |
| `models/game_state.go` | 中 | Position 支持百分比 |
| `api/tools/map.go` | 高 | 扩展响应，新增 Tools |
| `service/map.go` | 高 | 图片分析逻辑 |
| `store/postgres/map.go` | 中 | 存储 Image URL |
| `docs/server/详细设计.md` | 高 | 更新地图相关章节 |
| `docs/server/plan/M6-地图系统.md` | 高 | 添加新任务 |

### 7.2 需要修改的设计章节

1. **详细设计.md**:
   - 1.4 地图章节：添加 MapImage 定义
   - 2.5 地图/移动 Tools：添加新 Tools 定义
   - 3.4 地图系统流程：添加图片分析流程

2. **M6-地图系统.md**:
   - 添加新任务：T6-8 地图图片存储
   - 添加新任务：T6-9 视觉分析 Tool
   - 添加新任务：T6-10 地点标记管理

### 7.3 无影响的模块

- 骰子系统 (M4)
- 战斗系统 (M5)
- 角色系统 (M2, M3)
- 上下文管理

---

## 8. 实现建议

### 8.1 分阶段实现

**Phase 1: 基础图片存储** (M6 扩展)
- 扩展 Map 模型支持 Image
- 修改 `get_world_map` 返回图片 URL
- 前端负责显示图片

**Phase 2: 视觉 Tool** (新功能)
- 实现 `analyze_map_image` Tool
- 集成外部视觉服务（或返回描述让 Client 处理）

**Phase 3: 地点标记管理** (M6 扩展)
- 实现 `add_location`, `update_location`, `remove_location`
- 支持百分比坐标

**Phase 4: 高级功能** (可选)
- 自动地点识别
- 地图预处理

### 8.2 技术依赖

| 依赖 | 类型 | 说明 |
|------|------|------|
| 图片存储 | 外部 | 需要决定图片存储方案（本地/云存储） |
| 视觉模型 | 外部 | Claude、GPT-4V 或其他视觉 API |
| 图片处理库 | Go | 用于获取图片尺寸、格式等 |

### 8.3 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 图片存储成本 | 中 | 使用外部存储服务，支持 URL 引用 |
| 视觉模型调用成本 | 中 | 缓存分析结果，避免重复调用 |
| 性能影响 | 低 | 图片分析异步进行，不影响核心流程 |

---

## 9. 验收标准

### 9.1 功能验收

- [ ] 支持 `create_world_map` 上传图片
- [ ] `get_world_map` 返回图片 URL
- [ ] `analyze_map_image` 返回有效的视觉描述
- [ ] 地点标记支持百分比坐标
- [ ] 玩家位置支持百分比坐标

### 9.2 集成测试场景

```
1. 创建战役，上传世界地图图片
2. 调用 get_world_map，验证返回图片 URL
3. 调用 analyze_map_image，获取描述
4. 添加地点标记（百分比坐标）
5. 移动玩家到新位置（百分比坐标）
6. 验证所有数据正确存储和检索
```

---

## 10. 待确认问题

| 问题 | 影响 | 需要决策 |
|------|------|----------|
| 图片存储方案 | 高 | 本地存储 vs 云存储 (S3/OSS) |
| 视觉服务选择 | 高 | 集成到 Server vs 由 Client 负责 |
| 百分比坐标精度 | 中 | 使用 float32 还是 float64 |
| 图片大小限制 | 中 | 单张地图图片最大支持多少 MB |
| 缓存策略 | 低 | 视觉分析结果是否缓存，缓存多久 |

---

## 附录: D&D Hub 能力参考

### D&D Hub 核心功能

1. **地图操作**
   - 缩放、拖拽
   - 点击添加标记
   - 筛选标记（按类型）

2. **情报管理**
   - Wiki 记录人物、地点、线索
   - 线索板可视化推理
   - 时间轴

3. **地图与剧情**
   - 地点与剧情关联
   - 基于位置的线索发现

### 与本项目的对应

| D&D Hub 功能 | 本项目实现方式 |
|-------------|---------------|
| 地图显示 | 前端负责（图片 URL） |
| 地点标记 | Server 存储 Location |
| Wiki/线索 | 对话历史 + Campaign 描述 |
| 剧情关联 | LLM 基于地图和上下文生成 |
