# DND MCP

> D&D 5e 游戏的 MCP (Model Context Protocol) 实现项目

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**DND MCP** 是一个 LLM 驱动的 D&D 5e 游戏系统，采用 Client-Server 架构，提供完整的游戏状态管理、AI 对话和实时通信功能。

## 项目结构

```
dnd-mcp/
├── README.md              # 项目介绍（本文件）
├── CLAUDE.md              # Claude Code 开发指南
├── go.work                # Go workspace 配置
├── docs/                  # 设计文档
│   ├── client/            # Client 设计文档
│   ├── server/            # Server 设计文档
│   │   ├── plan/          # 开发计划（按里程碑拆分）
│   │   └── *.md           # 设计文档
│   ├── research/          # 研究想法和探索
│   ├── dnd5e规则书/       # D&D 5e 规则书资源
│   └── *.md               # 整体设计文档
├── packages/              # 核心组件包
│   ├── client/            # MCP Client - 会话和消息管理
│   └── server/            # MCP Server - 游戏状态管理
└── scripts/               # 构建和测试脚本
```

## 组件说明

### MCP Client (`packages/client/`)

**职责**: 会话管理、消息存储、AI 对话编排

- RESTful API 和 WebSocket 实时通信
- Redis 主存储 + PostgreSQL 备份
- LLM 集成（OpenAI 兼容）和 MCP 工具调用

[详细文档 →](packages/client/doc/README.md)

### MCP Server (`packages/server/`)

**职责**: 游戏状态管理、规则引擎、工具实现

- 战役和角色管理（玩家角色/NPC）
- 战斗系统（先攻、回合、伤害计算）
- 地图系统（大地图和战斗地图）
- 骰子投掷和规则查询
- 严格遵循 D&D 5e 官方规则书

[详细文档 →](docs/server/plan/README.md)

## 快速开始

### 前置要求

- **Go**: 1.24+
- **Redis**: 7.0+
- **PostgreSQL**: 15+ (可选)

### 启动 Client (Windows)

```powershell
cd packages/client
.\scripts\dev.ps1          # 设置开发环境
.\scripts\build.ps1        # 构建
.\bin\dnd-api.exe          # 运行
```

服务将在 `http://localhost:8080` 启动

### 启动 Server (Windows)

```powershell
cd packages/server
.\scripts\dev.ps1          # 设置开发环境
.\scripts\build.ps1        # 构建
.\bin\dnd-server.exe       # 运行
```

服务将在 `http://localhost:9000` 启动

## 架构设计

```
┌─────────────┐     HTTP/WebSocket      ┌─────────────┐
│   前端界面   │ ←──────────────────────→ │  MCP Client │
└─────────────┘                          └──────┬──────┘
                                               │
                                               │ MCP Protocol
                                               │
                                        ┌──────▼──────┐
                          ┌────────────→│  MCP Server │←────────────┐
                          │             └─────────────┘             │
                          │                    │                    │
                   ┌──────▼──────┐      ┌──────▼──────┐      ┌──────▼──────┐
                   │    Redis    │      │  PostgreSQL │      │  D&D 5e规则 │
                   │  (主存储)   │      │   (备份)    │      │  (规则书)   │
                   └─────────────┘      └─────────────┘      └─────────────┘
```

**数据流**:
1. 前端 → Client: HTTP/WebSocket 请求
2. Client → Server: MCP 工具调用
3. Server: 执行游戏规则（严格遵循 D&D 5e 规则书）
4. Server → Redis/PostgreSQL: 状态持久化

## 开发文档

| 文档 | 说明 |
|------|------|
| [整体架构设计](docs/整体架构设计.md) | Client + Server 技术架构 |
| [使用指南](docs/使用指南.md) | API 使用说明 |
| [代码规范](docs/代码规范.md) | 编码规范 |
| [Server 开发计划](docs/server/plan/README.md) | 里程碑和任务详情 |
| [D&D 5e 规则查询](docs/dnd5e规则书/查询指南.md) | 规则书索引 |

## 开发状态

### MCP Client (packages/client/)

✅ **已完成** - 会话管理、消息存储、LLM 集成、WebSocket

### MCP Server (packages/server/)

| 里程碑 | 主题 | 状态 |
|--------|------|------|
| M1 | 项目基础设施 | ✅ 已完成 |
| M2 | 战役管理 | ✅ 已完成 |
| M3 | 角色管理 | 🚧 进行中 |
| M4 | 骰子系统 | ⏳ 待开发 |
| M5 | 战斗系统 | ⏳ 待开发 |
| M6 | 地图系统 | ⏳ 待开发 |
| M7 | 上下文管理 | ⏳ 待开发 |
| M8 | 规则查询 | ⏳ 待开发 |
| M9 | 导入功能 | ⏳ 待开发 |

详见 [Server 开发计划](docs/server/plan/README.md)

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 规则书声明

本项目涉及 D&D 5e 规则的实现严格遵循官方规则书（玩家手册 PHB、城主指南 DMG、怪物图鉴 MM）。规则书资源位于 `docs/dnd5e规则书/` 目录，仅供内部开发参考。
