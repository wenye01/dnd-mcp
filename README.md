# DND MCP

> D&D 游戏的 MCP (Model Context Protocol) 实现项目

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**DND MCP** 是一个基于 MCP 协议的 D&D 龙与地下城游戏系统，提供完整的游戏状态管理、AI 对话和实时通信功能。

## 项目结构

```
dnd-mcp/
├── README.md              # 项目介绍（本文件）
├── go.work                # Go workspace 配置
├── docs/                  # 整体设计文档
│   ├── 系统详细设计.md
│   ├── 使用指南.md
│   └── 代码规范.md
├── packages/              # 核心组件包
│   ├── client/            # MCP Client - 会话和消息管理
│   └── server/            # MCP Server - 游戏状态管理
├── deployments/           # 部署配置
└── scripts/               # 共享脚本
```

## 组件说明

### MCP Client (`packages/client/`)

**职责**: 会话管理、消息存储、AI 对话编排

- 提供RESTful API和WebSocket实时通信
- Redis 主存储 + PostgreSQL 备份
- LLM 集成和 MCP 工具调用

[详细文档 →](packages/client/README.md)

### MCP Server (`packages/server/`)

**职责**: 游戏状态管理、规则引擎、工具实现

- 管理游戏状态（角色、战斗、场景）
- 提供 MCP 工具接口
- 实现游戏规则逻辑

[详细文档 →](packages/server/README.md)

## 快速开始

### 前置要求

- **Go**: 1.21+
- **Redis**: 7.0+
- **PostgreSQL**: 15+ (可选)

### 启动 Client

```bash
cd packages/client
go mod download
go run cmd/api/main.go
```

服务将在 `http://localhost:8080` 启动

### 启动 Server

```bash
cd packages/server
go mod download
go run cmd/server/main.go
```

服务将在 `http://localhost:9000` 启动

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 只启动 client
docker-compose up client

# 只启动 server
docker-compose up server
```

## 架构设计

```
┌─────────────┐     HTTP/WebSocket      ┌─────────────┐
│   前端界面   │ ←──────────────────────→ │  MCP Client │
└─────────────┘                          └──────┬──────┘
                                               │
                                               │ MCP Protocol
                                               │
┌─────────────┐                          ┌──────▼──────┐
│  PostgreSQL │ ←──────────────────────→ │  MCP Server │
└─────────────┘     数据持久化           └─────────────┘
```

**数据流**:
1. 前端 → Client: HTTP/WebSocket 请求
2. Client → Server: MCP 工具调用
3. Server → PostgreSQL: 状态持久化

## 开发文档

- [架构设计](docs/系统详细设计.md) - 整体技术架构
- [使用指南](docs/使用指南.md) - API 使用说明
- [代码规范](docs/代码规范.md) - 编码规范

## 开发状态

- ✅ MCP Client: 基础功能完成
- ⏳ MCP Server: 开发中

## 许可证

MIT License - 详见 [LICENSE](LICENSE)

## 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)
