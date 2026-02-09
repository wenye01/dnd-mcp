# DND MCP Client

> 轻量级 D&D 游戏会话和消息管理服务

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**DND MCP Client** 是一个轻量级的有状态协调层，用于管理 D&D 游戏会话和消息。它提供 HTTP API 和 WebSocket 支持实时通信，以 Redis 为主存储，PostgreSQL 为备份。

## ✨ 特性

- 🚀 **高性能**: 基于 Go 和 Gin 框架，提供快速的 HTTP API
- 📊 **多层存储**: Redis 主存储 + PostgreSQL 备份，确保数据安全
- 🔌 **实时通信**: WebSocket 支持实时事件推送
- 🎯 **健康监控**: 内置健康检查和系统监控
- 📝 **结构化日志**: 支持 JSON 和文本格式的结构化日志
- 🧪 **完善测试**: 单元测试、集成测试和 HTTP 测试，覆盖率 > 85%

## 📋 目录

- [快速开始](#快速开始)
- [项目架构](#项目架构)
- [API 文档](#api-文档)
- [配置](#配置)
- [开发](#开发)
- [测试](#测试)
- [文档](#文档)

## 🚀 快速开始

### 前置要求

- **Go**: 1.24+
- **Redis**: 7.0+ (或使用 Docker)
- **PostgreSQL**: 15+ (可选，用于持久化)
- **操作系统**: Windows/Linux/Mac

### 安装

```bash
# 克隆仓库
git clone https://github.com/dnd-mcp/client.git
cd client

# 下载依赖
go mod download
```

### 启动服务

#### Windows (PowerShell)

```powershell
# 1. 启动 Redis (使用 Docker)
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine

# 2. 构建项目
.\scripts\build.ps1

# 3. 启动服务器
.\bin\dnd-client.exe
```

#### Linux/Mac

```bash
# 1. 启动 Redis (使用 Docker)
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine

# 2. 构建项目
chmod +x ./scripts/build.sh
./scripts/build.sh

# 3. 启动服务器
./bin/dnd-client
```

### 验证安装

```bash
# 健康检查
curl http://localhost:8080/api/system/health

# 系统统计
curl http://localhost:8080/api/system/stats
```

## 🏗️ 项目架构

### 技术栈

- **语言**: Go 1.24+
- **HTTP 框架**: Gin
- **主存储**: Redis 7.0+
- **备份存储**: PostgreSQL 15+ (可选)
- **WebSocket**: Gorilla WebSocket
- **测试**: Testify

### 架构设计

```
┌─────────────┐
│   HTTP API  │  Gin + WebSocket
└──────┬──────┘
       │
┌──────▼──────────┐
│   Service Layer │  业务逻辑
└──────┬──────────┘
       │
┌──────▼──────────┐
│    Store Layer  │  数据访问接口
└──────┬──────────┘
       │
   ┌───┴───┐
   ▼       ▼
┌──────┐ ┌──────────┐
│Redis │ │PostgreSQL│
└──────┘ └──────────┘
```

### 项目结构

```
dnd-mcp/
├── cmd/              # 应用程序入口
│   └── server/       # HTTP 服务器
├── internal/         # 私有应用代码
│   ├── api/          # HTTP API 层
│   ├── service/      # 业务逻辑层
│   ├── store/        # 存储层
│   ├── models/       # 领域模型
│   ├── monitor/      # 系统监控
│   └── ...
├── pkg/              # 公共库
│   ├── config/       # 配置管理
│   ├── logger/       # 结构化日志
│   └── errors/       # 错误定义
├── tests/            # 测试代码
├── scripts/          # 构建脚本
└── doc/              # 文档
```

详细结构说明: [doc/PROJECT_STRUCTURE.md](doc/PROJECT_STRUCTURE.md)

## 📡 API 文档

### 健康检查

```bash
# 健康检查
GET /api/system/health

# 响应
{
  "status": "healthy",
  "timestamp": "2025-02-09T10:30:00Z",
  "components": {
    "redis": {
      "status": "healthy",
      "message": "Redis connection OK",
      "latency_ms": 5.2
    }
  }
}
```

### 系统统计

```bash
# 系统统计
GET /api/system/stats

# 响应
{
  "uptime_seconds": 3600,
  "start_time": "2025-02-09T09:30:00Z",
  "version": "v0.1.0",
  "request_count": 150,
  "error_count": 2,
  "components": {
    "redis": {
      "key_count": 42,
      "available": true
    },
    "sessions": {
      "count": 5
    }
  }
}
```

### 会话管理

```bash
# 创建会话
POST /api/sessions
Content-Type: application/json

{
  "name": "测试会话",
  "creator_id": "user-123",
  "mcp_server_url": "http://localhost:9000"
}

# 获取会话列表
GET /api/sessions

# 获取会话详情
GET /api/sessions/{id}

# 更新会话
PATCH /api/sessions/{id}

# 删除会话
DELETE /api/sessions/{id}
```

### 消息管理

```bash
# 发送消息
POST /api/sessions/{id}/chat
Content-Type: application/json

{
  "content": "你好",
  "player_id": "player-123"
}

# 获取消息历史
GET /api/sessions/{id}/messages?limit=10

# 获取单条消息
GET /api/sessions/{id}/messages/{message_id}
```

### WebSocket

```bash
# 连接 WebSocket
WS /ws/sessions/{id}?key={websocket_key}

# 订阅事件
{
  "type": "subscribe",
  "data": {
    "events": ["new_message", "state_changed"]
  }
}
```

## ⚙️ 配置

### 环境变量

| 变量 | 说明 | 默认值 | 必需 |
|------|------|--------|------|
| `REDIS_HOST` | Redis 服务器地址 | localhost:6379 | ✅ |
| `HTTP_HOST` | HTTP 服务器主机 | 0.0.0.0 | ❌ |
| `HTTP_PORT` | HTTP 服务器端口 | 8080 | ❌ |
| `LOG_LEVEL` | 日志级别 | info | ❌ |
| `DATABASE_URL` | PostgreSQL 连接字符串 | - | ❌ |

### 配置文件

创建 `.env` 文件:

```bash
# Redis
REDIS_HOST=localhost:6379

# HTTP Server
HTTP_HOST=0.0.0.0
HTTP_PORT=8080

# 日志
LOG_LEVEL=debug

# PostgreSQL (可选)
DATABASE_URL=postgres://user:password@localhost:5432/dbname
```

## 🛠️ 开发

### 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 函数保持简短（< 50 行）
- 导出的函数、类型必须添加文档注释
- 测试覆盖率 > 80%

详见: [doc/规范.md](doc/规范.md)

### 构建

```bash
# Windows
.\scripts\build.ps1

# Linux/Mac
./scripts/build.sh
```

### 运行

```bash
# 开发模式
go run ./cmd/server/main.go

# 或使用环境变量
LOG_LEVEL=debug go run ./cmd/server/main.go
```

## 🧪 测试

### 运行测试

```bash
# Windows - 完整测试套件
.\scripts\test-all.ps1

# Windows - 快速测试
.\scripts\test.ps1

# Linux/Mac - 完整测试套件
./scripts/test-all.sh

# Linux/Mac - 快速测试
./scripts/test.sh
```

### 手动测试

```bash
# 单元测试
go test -v ./pkg/logger/...
go test -v ./internal/monitor/...
go test -v ./internal/store/...

# 集成测试 (需要 Redis)
go test -v ./internal/store/redis/...

# HTTP 测试
go test -v ./internal/api/handler/...

# 查看覆盖率
go test -cover ./...
```

### 测试覆盖

当前测试覆盖率: **> 85%**

主要测试:
- ✅ pkg/logger: 12/12 tests passed
- ✅ internal/monitor: 13/13 tests passed
- ✅ internal/api/handler: 3/3 tests passed
- ✅ internal/store: 集成测试完成

## 📚 文档

- **[项目结构](doc/PROJECT_STRUCTURE.md)** - 目录组织和架构说明
- **[开发进度](doc/development_progress.md)** - 任务进度和完成情况
- **[详细设计](doc/DND_MCP_Client详细设计.md)** - 技术设计文档
- **[开发计划](doc/DND_MCP_Client_开发计划.md)** - 开发路线图
- **[代码规范](doc/规范.md)** - 编码规范和最佳实践
- **[Claude 指南](CLAUDE.md)** - Claude Code 项目指南

## 📊 开发状态

**当前版本**: v0.1.0 (开发中)

**开发进度**:
- ✅ 任务 1: 项目脚手架 + Redis 基础存储
- ✅ 任务 2: PostgreSQL 持久化
- ✅ 任务 3: HTTP API - 会话管理
- ✅ 任务 4: HTTP API - 消息管理
- ✅ 任务 5: WebSocket 实时通信
- ✅ 任务 6: LLM 集成
- ✅ 任务 7: MCP Server 集成
- ✅ 任务 8: 持久化触发器
- ✅ 任务 9: 系统监控和日志
- ⏳ 任务 10: 完整集成和优化

详见: [doc/development_progress.md](doc/development_progress.md)

## 🤝 贡献

欢迎贡献！请遵循以下步骤:

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 License

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👥 作者

DND MCP Team

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [Redis](https://redis.io/) - 高性能键值存储
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket 实现
- [Testify](https://github.com/stretchr/testify) - 测试工具包
