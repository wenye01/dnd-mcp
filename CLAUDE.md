# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 提供在此代码库中工作的指导。

## 项目概述

**DND MCP API** 是一个轻量级的有状态协调层，用于管理 D&D 游戏会话和消息。它提供 RESTful HTTP API 和 WebSocket 实时通信，以 Redis 为主存储，PostgreSQL 为备份。

- **语言**: Go 1.24+
- **主要开发环境**: Windows (PowerShell 脚本)
- **模块**: github.com/dnd-mcp/client
- **项目定位**: 纯后端 HTTP API 服务

## 核心架构

### 简化架构（当前实现）

项目使用简化架构（Handler → Service → Store → Models）以实现快速迭代：

```
Handler (HTTP) → Service → Store (Redis/PostgreSQL) → Models
```

**关键设计决策**: 项目使用适配器模式来桥接存储接口和持久化接口。适配器定义在 `cmd/api/main.go` 中。

### 存储策略

- **Redis**: 所有数据的主存储（会话、消息、系统元数据）
- **PostgreSQL**: 备份存储，定期持久化（默认每30秒）
- **数据隔离**: 集成测试使用 Redis DB 1，避免污染生产数据

### 核心组件

- **internal/models**: 领域模型（Session、Message）及业务逻辑方法
- **internal/store**: 存储接口和实现（Redis、PostgreSQL）
- **internal/api/handler**: HTTP 请求处理器
- **internal/service**: 业务逻辑层（SessionService）
- **internal/persistence**: Redis 和 PostgreSQL 之间的备份/恢复服务
- **internal/ws**: WebSocket 实时通信
- **internal/llm**: LLM 客户端集成（OpenAI 兼容）
- **internal/mcp**: MCP 协议客户端集成
- **internal/monitor**: 健康检查和系统统计监控
- **internal/api/dto**: 统一响应 DTO
- **internal/api/httperror**: 统一 HTTP 错误处理
- **pkg/config**: 从环境变量加载配置管理
- **pkg/logger**: 结构化日志（JSON/文本格式）
- **pkg/errors**: 应用级错误定义

## 构建和测试命令

### 构建 (Windows)

```powershell
# 构建项目
.\scripts\build.ps1

# 输出: bin/dnd-api.exe
```

### 测试 (Windows)

```powershell
# 运行所有测试（推荐 - 包括单元测试、集成测试和E2E测试）
.\scripts\test-all.ps1

# 从全新环境运行测试（单元 + 集成）
.\scripts\test-fresh.ps1

# 运行E2E测试（自动启动服务器）
.\scripts\test-e2e.ps1

# 快速测试（仅单元 + 集成）
.\scripts\test.ps1

# 仅运行单元测试
go test -v ./tests/unit/... -cover

# 运行集成测试（需要 Redis）
go test -v ./tests/integration/... -cover -timeout 30s

# 运行特定测试
go test -v ./tests/unit/service/... -run TestSessionCreate
```

### 环境设置

```powershell
# 设置开发环境
.\scripts\dev.ps1

# 重置环境（停止服务、清空 Redis、清理构建产物）
.\scripts\reset-env.ps1 -Force
```

### Redis 管理

```powershell
# 启动 Redis
.\scripts\start-redis.ps1

# 手动 Redis 连接（根据需要调整路径）
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING

# 清空所有 Redis 数据（警告：破坏性操作）
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" FLUSHALL

# 仅清空测试数据库（DB 1）
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" -n 1 FLUSHDB
```

### 运行服务器

```powershell
# 启动 API 服务器
.\bin\dnd-api.exe

# 使用环境变量
$env:LOG_LEVEL="debug"
$env:REDIS_HOST="localhost:6379"
.\bin\dnd-api.exe
```

## 配置

配置从环境变量加载（参见 `pkg/config/config.go`）：

**必需:**
- `REDIS_HOST`: Redis 服务器地址（默认: localhost:6379）

**可选:**
- `HTTP_HOST`: HTTP 服务器主机（默认: 0.0.0.0）
- `HTTP_PORT`: HTTP 服务器端口（默认: 8080）
- `LOG_LEVEL`: 日志级别（debug、info、warn、error）
- `DATABASE_URL`: PostgreSQL 连接字符串（用于持久化）

## 测试策略

项目使用三层测试方法：

### 1. 单元测试 (tests/unit/)
- 隔离测试业务逻辑
- 模拟外部依赖（Redis、PostgreSQL）
- 专注于服务层和模型
- 覆盖率: >85%

### 2. 集成测试 (tests/integration/)
- 使用真实 Redis 测试 API 端点
- 使用 HTTP 测试服务器测试处理器
- 测试数据隔离使用 Redis DB 1
- 通过 `tests/testutil/testutil.go` 自动清理

### 3. E2E测试 (tests/e2e/)
- 完整用户流程测试
- 对运行中的服务器进行真实 HTTP 调用
- 并发测试
- 错误情况测试

**重要提示**: 提交更改前务必运行 `.\scripts\test-all.ps1`。

## 代码规范

### 文件组织

- **cmd/**: 仅包含应用程序入口点（main.go）
- **internal/**: 私有应用代码
- **pkg/**: 可被外部项目导入的公共库
- **tests/**: 所有测试代码（单元测试、集成测试、E2E测试）
- **scripts/**: PowerShell 构建和测试脚本
- **doc/**: 设计和开发文档

### 命名规范

- **包名**: 小写单词（例如：`store`、`models`、`handler`）
- **文件名**: 小写加下划线（例如：`session_store.go`、`message_handler.go`）
- **接口**: `<action>er` 或描述性名词（例如：`SessionStore`、`MessageStore`）
- **函数**: 导出函数用 PascalCase，内部函数用 camelCase
- **常量**: UPPER_SNAKE_CASE

### 错误处理

- 使用 `pkg/errors` 中的预定义错误（例如：`errors.ErrSessionNotFound`）
- 使用 `errors.Wrap()` 或 `errors.Wrapf()` 包装错误并添加上下文
- 始终处理错误，不要忽略它们
- 返回错误，不要在底层记录日志

### 代码风格

- 遵循标准 Go 格式化（`gofmt`）
- 优先使用早期返回以避免深层嵌套
- 保持函数在 50 行以内
- 限制函数参数在 4 个以内（更多时使用结构体）
- 为导出的函数、类型和复杂逻辑添加注释

详见 `doc/代码规范.md`。

## 开发工作流

### 开始开发

1. 重置环境: `.\scripts\reset-env.ps1 -Force`
2. 运行测试: `.\scripts\test-all.ps1`
3. 开始编码

### 进行更改

1. 编辑源文件
2. 运行快速测试: `.\scripts\test.ps1`
3. 构建: `.\scripts\build.ps1`
4. 运行完整测试: `.\scripts\test-all.ps1`

### 提交前

1. 运行完整测试套件: `.\scripts\test-all.ps1`
2. 确保所有测试通过
3. 格式化代码: `gofmt -w .`
4. 如果依赖项发生变化，运行 `go mod tidy`

## 重要实现细节

### Redis 数据结构

- **会话**: `HSET session:{uuid}` + `SADD sessions:all {uuid}`
- **消息**: `ZADD msg:{session_id} {timestamp} {json}`
- **系统状态**: `SET system:status {status}`

### PostgreSQL 模式

- **client_sessions**: 会话元数据，支持软删除
- **client_messages**: 消息，带有到会话的外键
- 表由 `internal/persistence/migrate.go` 创建/管理

### 适配器模式

代码库使用适配器类型来桥接存储接口和持久化接口：
- `redisSessionReaderAdapter`、`redisSessionWriterAdapter`
- `redisMessageReaderAdapter`、`redisMessageWriterAdapter`

这些适配器定义在 `cmd/server/main.go` 中。

### 监控系统

监控系统使用可插拔架构：
- **健康检查器**: 实现 `HealthChecker` 接口
- **统计收集器**: 实现 `StatsCollector` 接口
- 预实现的 Redis 和会话检查器/收集器在 `internal/monitor/` 中可用

### 响应 DTO 和错误处理

- **DTOs**: 使用 `internal/api/dto` 包进行统一的 API 响应
- **Errors**: 使用 `internal/api/httperror` 包进行一致的 HTTP 错误响应
- 自动错误分类和 HTTP 状态码映射

## API 端点

### 会话
- `POST /api/sessions` - 创建会话
- `GET /api/sessions` - 列出所有会话
- `GET /api/sessions/:id` - 获取会话详情
- `PATCH /api/sessions/:id` - 更新会话
- `DELETE /api/sessions/:id` - 删除会话

### 消息
- `POST /api/sessions/:id/chat` - 发送消息
- `GET /api/sessions/:id/messages` - 列出消息

### 系统
- `GET /api/system/health` - 健康检查
- `GET /api/system/stats` - 系统统计

### WebSocket
- `GET /ws/sessions/:id` - 实时更新的 WebSocket 连接

## 故障排除

### 端口已被占用
```powershell
# 查找并终止使用端口 8080 的进程
Get-NetTCPConnection -LocalPort 8080 | Select-Object -ExpandProperty OwningProcess
Stop-Process -Id <PID> -Force

# 或使用重置脚本
.\scripts\reset-env.ps1 -Force
```

### Redis 连接问题
```powershell
# 检查 Redis 是否运行
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING

# 启动 Redis
.\scripts\start-redis.ps1
```

### 测试失败
```powershell
# 清除测试缓存并重试
go clean -testcache
.\scripts\test-all.ps1
```

## 需要理解的关键文件

1. **cmd/api/main.go** - 应用程序引导和依赖注入
2. **internal/models/session.go** - 会话领域模型
3. **internal/models/message.go** - 消息领域模型
4. **internal/store/interface.go** - 存储接口定义
5. **internal/store/redis/session.go** - Redis 会话存储实现
6. **internal/store/redis/message.go** - Redis 消息存储实现
7. **internal/api/handler/session.go** - 会话 HTTP 处理器
8. **internal/service/session.go** - 会话业务逻辑
9. **pkg/config/config.go** - 配置管理
10. **pkg/errors/errors.go** - 错误定义

## 文档

- **doc/README.md** - 文档导航
- **doc/系统详细设计.md** - 详细设计文档（完整架构）
- **doc/代码规范.md** - 代码规范和最佳实践
- **doc/使用指南.md** - 用户指南和 API 文档
- **scripts/README.md** - 脚本文档
- **tests/README.md** - 测试文档
- **README.md** - 项目概述和快速入门

## 模块信息

- **模块路径**: github.com/dnd-mcp/client
- **Go 版本**: 1.24.0
- **主要依赖**:
  - gin-gonic/gin (HTTP 框架)
  - redis/go-redis/v9 (Redis 客户端)
  - jackc/pgx/v5 (PostgreSQL 驱动)
  - gorilla/websocket (WebSocket 支持)
  - google/uuid (UUID 生成)
  - stretchr/testify (测试工具)
