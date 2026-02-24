# FVTT/UVTT 格式研究报告

> 研究日期: 2026-02-24
> 研究目的: 评估 FVTT 和 UVTT 格式的公开程度、兼容性，以及对 DND MCP 项目的参考价值

---

## 一、格式概述

### 1.1 FVTT (Foundry VTT) 格式

| 属性 | 说明 |
|------|------|
| **类型** | 闭源商业软件的内部数据格式 |
| **存储** | JSON 文件 (LevelDB 封装) |
| **公开程度** | 半公开 - API 文档可用，完整规范缺失 |
| **互操作性** | 低 - 仅限 Foundry VTT 生态 |
| **许可证** | 商业许可，代码闭源 |

### 1.2 UVTT (Universal VTT) 格式

| 属性 | 说明 |
|------|------|
| **类型** | 开放标准的地图导出格式 |
| **文件扩展名** | `.dd2vtt` 或 `.uvtt` |
| **来源** | Dungeondraft 等地图编辑工具 |
| **公开程度** | 完全公开 - 有规范文档 |
| **互操作性** | 高 - 跨平台通用 |

---

## 二、格式兼容性

### 2.1 UVTT ↔ FVTT 兼容关系

| 方向 | 兼容性 | 说明 |
|------|--------|------|
| UVTT → FVTT | ✅ 支持 | Foundry 内置 UVTT 导入器，可自动转换墙体、光源、格子 |
| FVTT → UVTT | ❌ 不支持 | Foundry 不提供原生 UVTT 导出功能 |

### 2.2 数据类型覆盖对比

| 数据类型 | UVTT | FVTT | 说明 |
|---------|------|------|------|
| 地图图像 | ✅ | ✅ | 背景图片 |
| 墙体/视线 | ✅ | ✅ | line_of_sight / Wall |
| 光源 | ✅ | ✅ | lights / AmbientLight |
| 格子系统 | ✅ | ✅ | grid / scene.grid |
| 传送门 | ✅ | ⚠️ | portals / 需转换 |
| **Token 位置** | ❌ | ✅ | Token 嵌入在 Scene 中 |
| **角色数据** | ❌ | ✅ | Actor 文档 |
| **物品/装备** | ❌ | ✅ | Item 文档 |
| **战斗追踪** | ❌ | ✅ | Combat 文档 |
| **日志/笔记** | ❌ | ✅ | JournalEntry 文档 |
| **宏脚本** | ❌ | ✅ | Macro 文档 |
| **音效/音乐** | ❌ | ✅ | Playlist 文档 |
| **战争迷雾** | ❌ | ✅ | FogExploration 文档 |
| **ActiveEffect** | ❌ | ✅ | 嵌入在 Actor/Item 中 |

---

## 三、FVTT 数据结构详解

### 3.1 主要文档类型 (Primary Documents)

存储在独立数据库表中，可通过 `game.actors`、`game.scenes` 等访问。

| 文档类型 | 说明 | 可存入 Compendium |
|---------|------|------------------|
| Actor | 角色 (PC/NPC/怪物) | ✅ |
| Cards | 卡牌组 | ✅ |
| Folder | 文件夹 (组织用) | ✅ |
| JournalEntry | 日志/笔记/知识条目 | ✅ |
| Item | 物品/装备/法术/专长 | ✅ |
| Macro | 宏脚本 | ✅ |
| Scene | 场景/地图 | ✅ |
| RollableTable | 随机表 | ✅ |
| ChatMessage | 聊天消息 | ❌ |
| Combat | 战斗追踪器 | ❌ |
| FogExploration | 战争迷雾探索数据 | ❌ |
| Playlist | 音乐/音效播放列表 | ❌ |
| Setting | 系统设置项 | ❌ |
| User | 用户账户 | ❌ |

### 3.2 Scene 嵌入文档 (Embedded Documents)

嵌入在 Scene 中，不存在于独立数据库。

| 嵌入类型 | 说明 | UVTT 对应 |
|---------|------|----------|
| AmbientLight | 环境光源 | ✅ lights |
| AmbientSound | 环境音效 | ❌ |
| Drawing | 绘图标注 | ❌ |
| MeasuredTemplate | 测量模板 (法术范围等) | ❌ |
| Note | 地图标注/笔记 | ❌ |
| Tile | 地图瓦片 (装饰层) | ❌ |
| **Token** | 角色/怪物标记 | ❌ |
| Wall | 墙体 (阻挡视线/移动) | ✅ line_of_sight |

### 3.3 Actor 相关嵌入文档

| 嵌入类型 | 说明 |
|---------|------|
| ActiveEffect | 临时/永久效果 (增益/减益) |
| Item | 物品 (嵌入在 Actor 中) |

### 3.4 其他嵌入文档

| 嵌入类型 | 父文档 | 说明 |
|---------|--------|------|
| Combatant | Combat | 参战者 |
| JournalEntryPage | JournalEntry | 日志页面 |
| PlaylistSound | Playlist | 音轨 |
| TableResult | RollableTable | 随机表结果 |

---

## 四、可还原的 JSON 格式

### 4.1 通用字段 (所有 Document 共有)

```json
{
  "_id": "string (16位随机ID)",
  "name": "string",
  "type": "string (由游戏系统定义，如 character/npc/vehicle)",
  "img": "string (图片路径)",
  "folder": "string|null (所属文件夹ID)",
  "sort": "number (排序值)",
  "permission": {
    "default": "number (默认权限级别)",
    "userId": "number (用户特定权限)"
  },
  "flags": {
    "moduleKey": { /* 模块自定义数据 */ }
  }
}
```

**权限级别说明：**
- 0: None - 不可见
- 1: Limited - 有限可见
- 2: Observer - 可查看
- 3: Owner - 完全控制

### 4.2 Token 格式 (还原度 ~80%)

Token 是 Scene 的嵌入文档，表示地图上的角色/物体标记。

```json
{
  "_id": "string",
  "name": "string",

  "x": "number (格子坐标 X)",
  "y": "number (格子坐标 Y)",
  "elevation": "number (高度/层数)",
  "width": "number (占用格子宽度，默认1)",
  "height": "number (占用格子高度，默认1)",

  "img": "string (Token 图片路径)",
  "alpha": "number (透明度 0-1，默认1)",
  "hidden": "boolean (是否对玩家隐藏)",
  "locked": "boolean (是否锁定位置)",
  "rotation": "number (旋转角度，默认0)",
  "scale": "number (缩放比例，默认1)",
  "mirrorX": "boolean (水平翻转)",
  "mirrorY": "boolean (垂直翻转)",

  "actorId": "string (关联的 Actor ID)",
  "actorLink": "boolean (是否链接到源 Actor)",
  "actorData": { /* 未链接时的数据覆盖 (ActorDelta) */ },

  "dimSight": "number (昏暗视觉范围，格子数)",
  "brightSight": "number (明亮视觉范围，格子数)",
  "sightAngle": "number (视野角度 0-360)",

  "dimLight": "number (发出的昏暗光范围)",
  "brightLight": "number (发出的明亮光范围)",
  "lightAngle": "number (光照角度 0-360)",
  "lightColor": "string (光色 #RRGGBB)",
  "lightAlpha": "number (光强度 0-1)",
  "lightAnimation": { "type": "string", "speed": "number", "intensity": "number" },

  "disposition": "number (态度: -1敌对, 0中立, 1友好)",

  "displayBars": "number (何时显示血条: 0=不显示, 1=悬停, 2=总是, 3=控制时)",
  "bar1": { "attribute": "system.attributes.hp", "value": 10, "max": 20 },
  "bar2": { "attribute": "system.attributes.ac.value", "value": 15, "max": null },

  "flags": {}
}
```

**无法从文档还原的 Token 字段：**
- 完整的光照动画配置
- 某些视觉特效属性
- 特定版本新增字段

### 4.3 Item 格式 (还原度 ~60%)

Item 可以独立存在，也可以嵌入在 Actor 中。

```json
{
  "_id": "string",
  "name": "string",
  "type": "string (weapon/spell/equipment/feat/consumable/tool/loot/class/backpack)",
  "img": "string (图标路径)",
  "folder": "string|null",
  "sort": "number",

  "system": {
    /* ⚠️ 核心问题：此字段由游戏系统定义，没有统一规范！
     *
     * 以下是 dnd5e 系统的武器示例：
     */
    "description": {
      "value": "<p>一把精良的长剑...</p>",
      "chat": "",
      "unidentified": ""
    },
    "source": "Player's Handbook",
    "quantity": 1,
    "weight": 3,
    "price": { "value": 15, "denomination": "gp" },
    "attunement": 0,
    "equipped": true,
    "rarity": "common",
    "identified": true,
    "activation": { "type": "action", "cost": 1, "condition": "" },
    "duration": { "value": "", "units": "" },
    "cover": null,
    "crewed": false,
    "target": { "value": null, "width": null, "units": "", "type": "" },
    "range": { "value": 5, "long": null, "units": "ft" },
    "uses": { "value": null, "max": "", "per": null, "recovery": "" },
    "consume": { "type": "", "target": null, "amount": null },
    "damage": {
      "critical": { "bonus": ""},
      "includeBase": true,
      "parts": [ ["1d8", "slashing"] ],
      "versatile": "1d10"
    },
    "armor": { "value": null, "type": "weapon", "dex": null },
    "hp": { "value": null, "max": null, "dt": null, "conditions": "" },
    "speed": { "value": null, "conditions": "" },
    "strength": null,
    "stealth": false,
    "proficient": true,
    "properties": {
      "two": false,    // Two-Handed
      "v": true,       // Versatile
      "hvy": false,    // Heavy
      "fin": true,     // Finesse
      "lgt": false,    // Light
      "thr": false,    // Thrown
      "rch": false,    // Reach
      "amm": false,    // Ammunition
      "ret": false,    // Returning
      "spe": false,    // Special
      "mgc": false     // Magical
    },
    "ability": "str",
    "actionType": "mwak",
    "attackBonus": "",
    "chatFlavor": "",
    "critical": { "threshold": null, "damage": "" },
    "formula": ""
  },

  "effects": [
    /* ActiveEffect 数组 */
  ],

  "flags": {
    "core": { "sourceId": "Compendium.dnd5e.items.Item.xxx" },
    "exportSource": { "world": "my-world", "system": "dnd5e", "coreVersion": "12.xxx" }
  }
}
```

**`system` 字段的特殊性：**
- 由每个游戏系统 (dnd5e, pf2e, wfrp4e, etc.) 自行定义
- 不同系统的结构完全不同
- 需要查看具体系统的 `template.json` 文件

### 4.4 Actor 格式 (还原度 ~70%)

```json
{
  "_id": "string",
  "name": "string",
  "type": "string (character/npc/vehicle - 由系统定义)",
  "img": "string (角色图片)",
  "folder": "string|null",
  "sort": "number",
  "prototypeToken": { /* 默认 Token 配置，结构同 Token */ },

  "system": {
    /* ⚠️ 由游戏系统定义
     * dnd5e 角色示例：
     */
    "abilities": {
      "str": { "value": 16, "proficient": 0, "bonuses": { "check": "", "save": "" } },
      "dex": { "value": 14, "proficient": 1, "bonuses": { "check": "", "save": "" } },
      "con": { "value": 12, "proficient": 0, "bonuses": { "check": "", "save": "" } },
      "int": { "value": 10, "proficient": 0, "bonuses": { "check": "", "save": "" } },
      "wis": { "value": 8,  "proficient": 0, "bonuses": { "check": "", "save": "" } },
      "cha": { "value": 13, "proficient": 0, "bonuses": { "check": "", "save": "" } }
    },
    "attributes": {
      "ac": { "flat": null, "calc": "default", "formula": "", "value": 14 },
      "hp": { "value": 24, "max": 24, "temp": 0, "tempmax": 0 },
      "init": { "ability": "dex", "bonus": 0, "value": 2 },
      "movement": { "burrow": 0, "climb": 0, "fly": 0, "swim": 0, "walk": 30, "units": "ft", "hover": false }
    },
    "details": {
      "biography": { "value": "", "public": "" },
      "alignment": "Lawful Good",
      "race": "Human",
      "background": "Soldier",
      "level": 3,
      "xp": { "value": 900, "min": 0, "max": 1800 }
    },
    "traits": {
      "size": "med",
      "di": { "value": [], "bypass": [], "custom": "" },  // Damage Immunity
      "dr": { "value": [], "bypass": [], "custom": "" },  // Damage Resistance
      "dv": { "value": [], "bypass": [], "custom": "" },  // Damage Vulnerability
      "ci": { "value": [], "custom": "" },                // Condition Immunity
      "languages": { "value": ["common"], "custom": "" }
    },
    "skills": {
      "acr": { "ability": "dex", "bonus": 0, "value": 2, "bonuses": { "passive": "", "check": "" } },
      // ... 其他技能
    },
    "prof": 2,  // Proficiency Bonus
    "currency": { "pp": 0, "gp": 50, "ep": 0, "sp": 10, "cp": 5 }
  },

  "items": [ /* 嵌入的 Item 数组 */ ],
  "effects": [ /* ActiveEffect 数组 */ ],
  "flags": {}
}
```

---

## 五、无法从文档还原的内容

| 缺失内容 | 影响 | 解决方法 |
|---------|------|---------|
| `system` 字段完整结构 | 无法自动解析游戏数据 | 查看具体系统的 template.json |
| 默认值定义 | 创建文档时需猜测默认值 | 导出真实数据对比 |
| 验证规则 | 必填/可选字段不明确 | 测试验证 |
| 运行时派生字段 | 某些字段是计算得出 | 分析源数据 |
| 版本差异 | v10/v11/v12 格式变化 | 按版本分别处理 |
| 光照/动画完整配置 | 高级视觉效果属性 | 实验导出 |

---

## 六、实际获取完整格式的方法

### 方法 1: 导出真实数据 (推荐)

```
在 Foundry VTT 中：
1. 打开 Actor/Item 目录
2. 右键目标文档
3. 选择 "Export Data"
4. 保存 JSON 文件
5. 分析实际字段结构
```

### 方法 2: 查看游戏系统源码

各游戏系统的 `template.json` 定义了 `system` 字段结构：

| 游戏系统 | 仓库地址 |
|---------|---------|
| D&D 5e | https://gitlab.com/foundrynet/dnd5e |
| Pathfinder 2e | https://gitlab.com/foundrynet/pf2e |
| Warhammer 4e | https://github.com/moo-man/WFRP4e-FoundryVTT |

### 方法 3: 读取数据库文件

```
文件位置: FoundryVTT/Data/worlds/{world-name}/
结构:
├── actors.db/      # LevelDB 存储
├── scenes.db/
├── items.db/
├── journal.db/
└── ...

每个 .db 目录包含 JSON 格式的文档数据
```

### 方法 4: 查看 API 文档

- 官方 API: https://foundryvtt.com/api/v12/
- 社区 Wiki: https://foundryvtt.wiki/

---

## 七、对 DND MCP 项目的建议

### 7.1 格式支持优先级

| 优先级 | 格式 | 策略 | 理由 |
|-------|------|------|------|
| **P0** | 项目自有格式 | 完整定义 JSON Schema | 数据主权，格式稳定 |
| **P1** | UVTT | 支持导入 | 公开标准，地图数据足够 |
| **P2** | FVTT Token | 支持按需导入 | 核心 Token 数据可解析 |
| **P2** | FVTT Item (基础) | 导入 name/type/img | 基础信息可获取 |
| **P3** | FVTT Item (system) | 按游戏系统适配 | 需要大量工作，格式不稳定 |
| **P3** | FVTT 完整数据 | 不追求兼容 | 格式不公开，ROI 低 |

### 7.2 推荐的项目数据结构

参考 FVTT 但自主定义：

```go
// Token 数据结构
type Token struct {
    ID          string    `json:"_id"`
    Name        string    `json:"name"`
    X, Y        float64   `json:"x,y"`
    Width       float64   `json:"width"`
    Height      float64   `json:"height"`
    Image       string    `json:"img"`
    Hidden      bool      `json:"hidden"`
    ActorID     string    `json:"actorId,omitempty"`
    Vision      Vision    `json:"vision"`
    Light       Light     `json:"light"`
    // ... 项目特有字段
}

// Item 数据结构 - 简化版，不绑定具体游戏系统
type Item struct {
    ID          string                 `json:"_id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Image       string                 `json:"img"`
    Description string                 `json:"description"`
    Properties  map[string]interface{} `json:"properties"` // 灵活存储
    // ... 项目特有字段
}
```

### 7.3 导入工具设计

```
FVTT Import Pipeline:
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ FVTT JSON   │ --> │ Converter   │ --> │ Project DB  │
│ (导出文件)   │     │ (按类型转换) │     │ (自有格式)  │
└─────────────┘     └─────────────┘     └─────────────┘

转换策略:
- Token: 80% 字段可直接映射
- Item: 基础字段映射 + system 字段保留原 JSON
- Actor: 同 Item
- Scene: 墙体/光源可从 UVTT 导入，Token 需单独处理
```

---

## 八、结论

| 问题 | 答案 |
|------|------|
| **FVTT 格式是否公开？** | 半公开 - API 有文档，完整规范缺失 |
| **能否从文档完整还原？** | 不能 - 约 60-80% 还原度 |
| **Token 格式能否还原？** | 基本可以 - 核心字段有文档 (~80%) |
| **Item 格式能否还原？** | 部分可以 - `system` 字段需游戏系统定义 (~60%) |
| **UVTT/FVTT 是否兼容？** | 单向兼容 - UVTT 可导入 FVTT，反之不行 |
| **项目应支持哪种？** | 优先 UVTT (地图) + 自有格式 (其他数据) |

---

## 九、参考资源

### 官方资源
- Foundry VTT 官网: https://foundryvtt.com/
- API 文档: https://foundryvtt.com/api/v12/
- 许可证: https://foundryvtt.com/article/license/

### 社区资源
- Foundry VTT Wiki: https://foundryvtt.wiki/
- D&D 5e 系统源码: https://gitlab.com/foundrynet/dnd5e

### UVTT 相关
- UVTT 规范: https://foundryvtt.com/article/universal-vtt/
- Dungeondraft (UVTT 导出工具): https://cartographyassets.itch.io/dungeondraft
