# DND MCP Client

> 一个基于 Go 和 PostgreSQL 的 DND（龙与地下城）MCP (Model Context Protocol) 客户端实现，支持与 LLM 集成，提供完整的会话管理和消息处理功能。

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 特性

- 🎯 **完整的 MCP 协议实现** - 支持标准 Model Context Protocol
- 💬 **会话管理** - 多会话支持，每个会话独立管理
- 🤖 **LLM 集成** - 支持 OpenAI API，易于扩展其他提供商
- 🔄 **自动重试** - 内置智能重试机制，处理 429 和 5xx 错误
- 📊 **数据持久化** - PostgreSQL 存储，支持完整的 CRUD 操作
- 🧪 **完整测试** - 27+ 单元测试 + 5+ 集成测试
- 🚀 **一键部署** - 完整的自动化脚本，从零到生产
- 📈 **高并发** - 支持并发消息处理，线程安全

## 📋 目录

- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [核心功能](#核心功能)
- [开发指南](#开发指南)
- [测试](#测试)
- [部署](#部署)
- [文档](#文档)
- [贡献](#贡献)
- [许可证](#许可证)

## 🚀 快速开始

### 前置要求

- Go 1.25+
- PostgreSQL 14+
- PowerShell (Windows) 或 Bash (Linux/macOS)

### 一键启动

```powershell
# 1. 克隆项目
git clone https://github.com/your-org/dnd-mcp.git
cd dnd-mcp

# 2. 安装依赖
go mod download

# 3. 初始化数据库
.\scripts\init-database.ps1

# 4. 运行测试
.\scripts\test.ps1
```

**就这么简单！** 🎉

### 验证安装

```powershell
# 检查 Go
go version

# 检查 PostgreSQL
psql --version

# 运行快速测试
go test -v ./tests/unit/store -run TestPostgresStore_CreateSession
```

## 📁 项目结构

```
dnd-mcp/
├── cmd/                  # 主程序入口
│   └── server/           # HTTP 服务器
├── internal/             # 内部包（不对外暴露）
│   ├── api/             # API 处理器
│   │   └── handler/     # HTTP 请求处理
│   ├── client/          # 客户端实现
│   │   └── llm/         # LLM 客户端（OpenAI）
│   ├── models/          # 数据模型定义
│   └── store/           # 数据持久化层
├── tests/               # 测试代码
│   ├── unit/            # 单元测试
│   │   ├── store/       # Store 测试
│   │   ├── client/llm/  # LLM 测试
│   │   └── api/handler/ # Handler 测试
│   ├── integration/     # 集成测试
│   └── reports/         # 测试报告
├── scripts/             # 脚本工具
│   ├── migrate/         # 数据库迁移工具
│   ├── migrations/      # SQL 迁移文件
│   ├── test.ps1         # 测试脚本
│   ├── init-database.ps1 # 数据库初始化
│   └── drop-database.ps1 # 数据库清理
├── doc/                 # 项目文档
├── go.mod
├── go.sum
└── README.md
```

## 🎯 核心功能

### 1. 会话管理

```go
// 创建新会话
session := &models.Session{
    ID:           uuid.New(),
    CampaignName: "被遗忘的国度",
    Location:     "地下城入口",
    GameTime:     "Morning",
    State:        make(map[string]interface{}),
}
store.CreateSession(ctx, session)

// 获取会话
session, err := store.GetSession(ctx, sessionID)
```

### 2. 消息处理

```go
// 创建消息
message := &models.Message{
    ID:        uuid.New(),
    SessionID: sessionID,
    Role:      "user",
    Content:   "我要攻击那个哥布林",
    PlayerID:  "player-001",
}
store.CreateMessage(ctx, message)

// 获取消息历史
messages, err := store.GetMessages(ctx, sessionID, 100, 0)
```

### 3. LLM 集成

```go
// 创建 OpenAI 客户端
config := &llm.Config{
    APIKey:      "your-api-key",
    Model:       "gpt-4",
    Temperature: 0.7,
    MaxRetries:  3,
}
client := llm.NewOpenAIClient(config)

// 发送聊天请求
req := &llm.ChatCompletionRequest{
    Model:    "gpt-4",
    Messages: []llm.Message{
        {Role: "system", Content: "你是一个DND地下城主"},
        {Role: "user", Content: "我要投骰子"},
    },
}
resp, err := client.ChatCompletion(ctx, req)
```

### 4. API 处理

```go
// 创建 Chat Handler
handler := handler.NewChatHandler(llmClient, dataStore)

// 注册路由
router.POST("/api/sessions/:id/chat", handler.ChatMessage)
```

## 📖 开发指南

### 环境设置

```powershell
# 1. 克隆项目
git clone <repository-url>
cd dnd-mcp

# 2. 安装依赖
go mod download

# 3. 设置环境变量（可选）
$env:PGPASSWORD = "your_password"
$env:TEST_DB_PASSWORD = "your_password"
```

### 运行项目

```powershell
# 构建项目
go build -o bin/dnd-mcp.exe ./cmd/server

# 运行服务器
.\bin\dnd-mcp.exe
```

### 代码格式化

```powershell
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

## 🧪 测试

### 运行所有测试

```powershell
# 使用测试脚本（推荐）
.\scripts\test.ps1

# 或直接使用 go test
go test -v ./tests/unit/... ./tests/integration/...
```

### 测试覆盖

| 类型 | 数量 | 文件 |
|------|------|------|
| 单元测试 | 27 | `tests/unit/` |
| 集成测试 | 5 | `tests/integration/` |
| **总计** | **32** | |

### 查看测试报告

```powershell
# 测试报告
Get-Content tests\reports\*.txt

# 覆盖率报告（HTML）
start tests\reports\coverage.html
```

### 运行特定测试

```powershell
# 单元测试
go test -v ./tests/unit/...

# 集成测试
go test -v ./tests/integration/...

# 特定包
go test -v ./tests/unit/store

# 特定测试函数
go test -v ./tests/unit/store -run TestPostgresStore_CreateSession
```

## 🚀 部署

### 快速部署

```powershell
# 一键部署（全新环境）
.\scripts\clean-and-test.ps1
```

### 手动部署

```powershell
# 1. 初始化数据库
.\scripts\init-database.ps1

# 2. 运行迁移
go run scripts/migrate/main.go -action up -dsn "postgres://postgres:password@localhost:5432/dnd_mcp_test?sslmode=disable"

# 3. 运行测试
.\scripts\test.ps1
```

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PGPASSWORD` | `070831` | PostgreSQL 密码 |
| `TEST_DB_PASSWORD` | `070831` | 测试数据库密码 |
| `DATABASE_URL` | 自动生成 | 完整数据库连接字符串 |

## 📚 文档

- **[SETUP_GUIDE.md](SETUP_GUIDE.md)** - 完整的部署和设置指南
- **[QUICKSTART.md](QUICKSTART.md)** - 快速参考卡片
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - 部署和维护指南
- **[tests/README.md](tests/README.md)** - 测试指南
- **[doc/](doc/)** - 项目详细文档
  - [MCP_Client开发计划.md](doc/MCP_Client开发计划.md) - 开发计划
  - [MCP_Client设计.md](doc/MCP_Client设计.md) - 架构设计
  - [规范.md](doc/规范.md) - 编码规范

## 🛠️ 技术栈

- **语言**: Go 1.25+
- **数据库**: PostgreSQL 14+
- **Web 框架**: Gin
- **LLM**: OpenAI API
- **测试**: Testify
- **数据库驱动**: lib/pq

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

### 开发流程

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 遵循 [doc/规范.md](doc/规范.md)
- 所有代码必须通过 `go vet` 和 `go fmt`
- 新功能必须包含测试
- 测试覆盖率不低于 80%

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🔗 相关链接

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [OpenAI API 文档](https://platform.openai.com/docs/api-reference)
- [Gin Web 框架](https://gin-gonic.com/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)

## 💬 联系方式

- 项目主页: [https://github.com/your-org/dnd-mcp](https://github.com/your-org/dnd-mcp)
- 问题反馈: [GitHub Issues](https://github.com/your-org/dnd-mcp/issues)
- 邮件: your-email@example.com

## 🙏 致谢

感谢所有贡献者！

---

**⭐ 如果这个项目对你有帮助，请给个 Star！**

Made with ❤️ by DND MCP Team
