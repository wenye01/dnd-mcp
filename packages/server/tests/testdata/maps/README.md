# Test Data for Map Import (M6.5)

测试数据用于 M6.5 地图导入功能的单元测试和集成测试。

## 目录结构

```
maps/
├── README.md                    # 本文件
├── sample.uvtt                  # 标准 UVTT 2.x 格式样本
├── sample_scene.json            # FVTT Scene 样本
├── sample_module/               # 最小化 FVTT 模块
│   ├── module.json              # 模块清单
│   └── packs/
│       └── scenes.db            # Compendium 数据库 (JSONL)
└── images/                      # 测试图片资源
    ├── small.webp               # 小图片 (~16KB)
    ├── medium.webp              # 中图片 (~121KB)
    └── large.webp               # 大图片 (~94KB)
```

## 文件说明

### sample.uvtt
Universal VTT 格式 2.x 样本文件，包含：
- `resolution`: 像素/网格配置 (140px/grid, 10x10 grids)
- `line_of_sight`: 视线阻挡多边形
- `portals`: 门/传送点
- `walls`: 墙壁定义

用途：测试 UVTT Parser (T6.5-2)

### sample_scene.json
Foundry VTT Scene 样本，包含：
- 基本场景配置 (_id, name, width, height, grid)
- 墙壁 (walls) - 4面墙
- Token (tokens) - 1个测试Token
- 光源 (lights) - 1个光源
- Tile (tiles) - 背景图

用途：测试 FVTT Scene Parser (T6.5-3)

### sample_module/
最小化的 FVTT 模块结构：
- `module.json`: 模块清单文件
- `packs/scenes.db`: JSONL 格式的场景数据库

用途：测试 LevelDB Compendium Parser (T6.5-4) 和 import_map_from_module Tool (T6.5-7)

### images/
测试图片资源，从 baileywiki-maps 复制：
- `small.webp`: ~16KB - stone-background.webp
- `medium.webp`: ~121KB - treetop1.webp
- `large.webp`: ~94KB - barn-lvl2.webp

用途：测试图片加载和处理

## 数据统计

- 总大小: ~256KB
- JSON 文件: 4个
- 图片文件: 3个
- UVTT 文件: 1个
- Compendium DB: 1个 (包含1条场景记录)

## 来源

图片和场景设计参考了 Baileywiki Free Maps Pack：
- 原始数据: `docs/data/baileywiki-maps/`
- 许可: 需确认原始许可条款
