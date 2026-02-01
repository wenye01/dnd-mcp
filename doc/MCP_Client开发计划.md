# MCP Client 开发计划

## 计划概述

本文档按照 **CODE_STANDARDS.md** 规范,制定 MCP Client 的持续集成开发计划。

### 开发原则

1. **持续集成**：每个阶段结束时都是可运行、可测试的完整程序
2. **增量开发**：从最小可用功能开始,逐步增加复杂度
3. **测试驱动**：每个功能完成后都有对应的单元测试和集成测试
4. **代码规范**：严格遵循 CODE_STANDARDS.md 的编码标准
5. **跨平台设计**：Windows 开发,保证后续可部署到 Linux/macOS

### 项目结构 (按规范)

```
mcp-client/
├── cmd/
│   └── mcp-client/
│       └── main.go                 # 程序入口
├── internal/
│   ├── api/                        # HTTP/WebSocket API层
│   │   ├── router.go              # 路由注册
│   │   ├── middleware/            # 中间件
│   │   │   ├── request_id.go
│   │   │   ├── logging.go
│   │   │   └── recovery.go
│   │   ├── handler/               # HTTP处理器
│   │   │   ├── session.go
│   │   │   ├── chat.go
│   │   │   └── query.go
│   │   └── websocket/             # WebSocket管理
│   │       ├── manager.go
│   │       ├── conn.go
│   │       └── hub.go
│   ├── orchestrator/              # 核心业务逻辑
│   │   ├── orchestrator.go        # 主控制器
│   │   ├── context_builder.go     # 上下文构建
│   │   ├── llm_manager.go         # LLM管理
│   │   ├── tool_coordinator.go    # 工具协调
│   │   └── response_generator.go  # 响应生成
│   ├── client/                    # 外部服务客户端
│   │   ├── llm/                   # LLM客户端
│   │   │   ├── client.go          # 接口定义
│   │   │   ├── openai.go          # OpenAI实现
│   │   │   └── mock.go            # 测试Mock
│   │   └── mcp/                   # MCP客户端
│   │       ├── client.go          # MCP客户端
│   │       ├── protocol.go        # 协议编解码
│   │       └── retry.go           # 重试逻辑
│   ├── store/                     # 数据持久化层
│   │   ├── message.go             # 消息存储
│   │   ├── session.go             # 会话存储
│   │   └── interface.go           # 存储接口
│   ├── cache/                     # 缓存管理层
│   │   └── redis.go               # Redis缓存
│   ├── models/                    # 数据模型定义
│   │   ├── session.go
│   │   ├── message.go
│   │   ├── chat.go
│   │   └── event.go
│   ├── dispatcher/                # 事件分发
│   │   ├── dispatcher.go          # 事件分发器
│   │   ├── queue.go               # 事件队列
│   │   └── broadcaster.go         # WebSocket广播
│   └── config/                    # 配置管理
│       ├── config.go              # 配置结构
│       └── loader.go              # 配置加载
├── pkg/                           # 公共库(可被外部导入)
│   └── errors/                    # 错误定义
│       └── errors.go
├── mock/                          # Mock服务器和工具
│   ├── mcp_server/                # Mock MCP Server
│   │   ├── main.go
│   │   ├── handlers.go
│   │   └── data.go
│   ├── llm_server/                # Mock LLM Server
│   │   ├── main.go
│   │   └── handlers.go
│   └── fixtures/                  # 测试数据
│       ├── mcp_responses.go
│       └── llm_responses.go
├── scripts/                       # 构建和部署脚本
│   ├── build.sh
│   └── deploy.sh
├── tests/                         # 集成测试
│   ├── integration/               # 集成测试
│   │   └── chat_test.go
│   └── testutil/                  # 测试工具
│       ├── setup.go
│       └── helpers.go
├── go.mod
├── go.sum
├── config.yaml                    # 配置文件
├── config.test.yaml               # 测试配置
├── Makefile                       # 构建脚本 (Linux/macOS)
├── build.bat                      # 构建脚本 (Windows)
├── scripts/                       # 构建和部署脚本
│   ├── build.bat                 # Windows 构建脚本
│   ├── build.sh                  # Linux/macOS 构建脚本
│   ├── test.bat                  # Windows 测试脚本
│   ├── test.sh                   # Linux/macOS 测试脚本
│   ├── setup.ps1                 # Windows 环境设置
│   └── setup.sh                  # Linux/macOS 环境设置
├── deploy/                        # 部署配置 (可选,用于跨平台部署)
│   ├── Dockerfile
│   └── docker-compose.yml
└── README.md
```

---

## 开发环境设置 (Windows)

### 必需软件

1. **Go 开发环境**
   - 下载: https://go.dev/dl/
   - 版本: Go 1.21 或更高
   - 安装后验证: `go version`

2. **PostgreSQL 数据库**
   - 下载: https://www.postgresql.org/download/windows/
   - 版本: PostgreSQL 15 或更高
   - 安装时记住设置的密码
   - 默认端口: 5432
   - 创建测试数据库: `createdb -U postgres mcp_test`

3. **Redis 缓存** (可选,用于缓存层)
   - 下载: https://github.com/microsoftarchive/redis/releases
   - 或使用 Memurai (Redis Windows 兼容版本)
   - 默认端口: 6379

4. **Git** (版本控制)
   - 下载: https://git-scm.com/download/win

### 可选软件 (用于跨平台部署)

- **Docker Desktop**: 用于后续 Linux/macOS 部署测试
- **WSL2**: Windows Subsystem for Linux (可选)

### 环境变量配置

**Windows PowerShell**:
```powershell
# 设置系统环境变量或用户环境变量
$env:GOOS="windows"
$env:GOARCH="amd64"

# 数据库配置
$env:DATABASE_URL="postgres://postgres:your-password@localhost:5432/mcp_test?sslmode=disable"

# Redis 配置 (如果使用)
$env:REDIS_URL="redis://localhost:6379"

# LLM 配置 (测试时使用 Mock)
$env:LLM_PROVIDER="mock"
$env:LLM_SERVER_URL="http://localhost:9001"

# MCP Server 配置 (测试时使用 Mock)
$env:MCP_SERVER_URL="http://localhost:9000"
```

**永久设置** (系统属性 → 高级 → 环境变量):
```
DATABASE_URL=postgres://postgres:your-password@localhost:5432/mcp_test?sslmode=disable
LLM_PROVIDER=mock
LLM_SERVER_URL=http://localhost:9001
MCP_SERVER_URL=http://localhost:9000
```

### 本地服务启动

**方式1: 手动启动服务**

1. **启动 PostgreSQL**:
   - 服务管理器中启动 PostgreSQL 服务
   - 或使用命令: `net start postgresql-x64-15`

2. **启动 Redis** (如果使用):
   ```powershell
   # 安装为服务后
   net start Redis
   ```

3. **启动 Mock MCP Server**:
   ```powershell
   cd mock\mcp_server
   go run main.go
   # 监听 :9000
   ```

4. **启动 Mock LLM Server**:
   ```powershell
   cd mock\llm_server
   go run main.go
   # 监听 :9001
   ```

**方式2: 使用启动脚本**

创建 `scripts\dev-start.bat`:
```batch
@echo off
echo Starting development environment...

start "PostgreSQL" net start postgresql-x64-15
start "Redis" net start Redis

timeout /t 2 /nobreak > nul

start "Mock MCP Server" cmd /k "cd /d %~dp0..\mock\mcp_server && go run main.go"
start "Mock LLM Server" cmd /k "cd /d %~dp0..\mock\llm_server && go run main.go"

echo All services started!
echo - Mock MCP Server: http://localhost:9000
echo - Mock LLM Server: http://localhost:9001
pause
```

创建 `scripts\dev-stop.bat`:
```batch
@echo off
echo Stopping development environment...

taskkill /F /IM go.exe > nul 2>&1

echo Stopping PostgreSQL...
net stop postgresql-x64-15

echo Stopping Redis...
net stop Redis

echo All services stopped!
pause
```

### 项目初始化

```powershell
# 1. 克隆或创建项目目录
cd C:\Code\dnd\mcp-client

# 2. 初始化 Go module
go mod init github.com/yourname/mcp-client

# 3. 安装依赖
go mod tidy

# 4. 创建数据库
# 使用 pgAdmin 或命令行
createdb -U postgres mcp_test

# 5. 运行数据库迁移
go run scripts/migrate/main.go

# 6. 启动开发环境
.\scripts\dev-start.bat
```

### IDE 推荐

1. **GoLand** (JetBrains) - 功能强大,收费
2. **VS Code** + Go 扩展 - 轻量级,免费
   - 扩展: Go (Google 官方)
   - 扩展: Docker (可选)
   - 扩展: PostgreSQL (可选)

### 常用命令

```powershell
# 构建项目
.\build.bat
# 或
go build -o bin\mcp-client.exe cmd\mcp-client\main.go

# 运行项目
go run cmd\mcp-client\main.go

# 运行所有测试
go test -v .\internal\...

# 运行特定测试
go test -v .\internal\store\...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html

# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

---

## Mock 实现方案

### 概述

由于当前只开发 MCP Client，所有外部依赖都需要通过 Mock 实现进行测试。

### 外部依赖清单

| 依赖组件 | 部署方式 | 说明 |
|---------|---------|------|
| **MCP Server** | Mock HTTP Server (本地进程) | 模拟工具调用和状态查询 |
| **OpenAI API** | Mock HTTP Server (本地进程) | 模拟LLM响应和工具调用 |
| **PostgreSQL** | 本地安装 | 真实数据库用于测试 |
| **Redis** | 本地安装 (可选) | 真实缓存用于测试 |
| **WebSocket客户端** | 测试代码 | 模拟前端连接 |

### Mock MCP Server 实现

**文件位置**: `mock/mcp_server/`

**功能需求**:

1. **状态管理接口**
   - `GET /state?session_id={id}` - 返回会话状态
   - 响应示例:
     ```json
     {
       "session_id": "test-session-001",
       "location": "地下城入口",
       "game_time": "2024-01-01T12:00:00Z",
       "characters": [
         {"id": "char-001", "name": "玩家A", "hp": 20, "hp_max": 20}
       ],
       "combat": null
     }
     ```

2. **工具调用接口**
   - `POST /tools/call` - 执行工具调用
   - 支持的工具:
     - `resolve_attack`: 攻击结算
     - `roll_dice`: 投骰子
     - `move_character`: 角色移动
     - `create_character`: 创建角色

3. **事件生成**
   - 每次工具调用返回事件列表
   - 支持的事件类型:
     - `dice.rolled`: 骰子投掷
     - `combat.attack_resolved`: 攻击结算
     - `character.hp_changed`: HP变化
     - `character.moved`: 角色移动

**实现代码结构**:

```go
// mock/mcp_server/main.go
package main

var (
    sessions = make(map[string]*SessionState)
    mu       sync.RWMutex
)

type SessionState struct {
    SessionID string
    Location  string
    GameTime  time.Time
    Characters []Character
    Combat    *CombatState
}

func main() {
    http.HandleFunc("/state", handleGetState)
    http.HandleFunc("/tools/call", handleToolCall)
    http.ListenAndServe(":9000", nil)
}

func handleGetState(w http.ResponseWriter, r *http.Request) {
    sessionID := r.URL.Query().Get("session_id")

    mu.RLock()
    session, exists := sessions[sessionID]
    mu.RUnlock()

    if !exists {
        // 返回默认空状态
        session = createDefaultSession(sessionID)
        mu.Lock()
        sessions[sessionID] = session
        mu.Unlock()
    }

    json.NewEncoder(w).Encode(session)
}

func handleToolCall(w http.ResponseWriter, r *http.Request) {
    var req ToolCallRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 根据工具名称执行不同逻辑
    result, events := executeTool(req)

    response := ToolCallResponse{
        Result:  result,
        Version: 1,
        Events:  events,
    }

    json.NewEncoder(w).Encode(response)
}
```

**测试数据**: `mock/fixtures/mcp_responses.go`

```go
package fixtures

var (
    // 默认会话状态
    DefaultSessionState = map[string]interface{}{
        "session_id": "test-session-001",
        "location": "地下城入口",
        "characters": []map[string]interface{}{
            {
                "id": "char-001",
                "name": "玩家A",
                "hp": 20,
                "hp_max": 20,
            },
        },
        "combat": nil,
    }

    // resolve_attack 工具响应
    AttackToolResponse = map[string]interface{}{
        "hit": true,
        "damage": 8,
        "attack_roll": 18,
        "damage_roll": "2d6",
    }

    // 攻击产生的事件
    AttackEvents = []map[string]interface{}{
        {
            "type": "dice.rolled",
            "data": map[string]interface{}{
                "roll_type": "attack",
                "result": 18,
                "modifier": 3,
            },
        },
        {
            "type": "combat.attack_resolved",
            "data": map[string]interface{}{
                "attacker": "char-001",
                "target": "goblin-001",
                "damage": 8,
            },
        },
    }
)
```

### Mock LLM Server 实现

**文件位置**: `mock/llm_server/`

**功能需求**:

1. **Chat Completion 接口**
   - `POST /v1/chat/completions` - 模拟OpenAI API
   - 支持流式和非流式响应
   - 支持工具调用 (`tool_calls`)

2. **响应模式**
   - **简单对话模式**: 只返回文本
   - **工具调用模式**: 返回 `tool_calls`
   - **混合模式**: 先文本,再工具调用

**实现代码结构**:

```go
// mock/llm_server/main.go
package main

type ChatCompletionRequest struct {
    Messages []Message `json:"messages"`
    Tools    []Tool    `json:"tools,omitempty"`
    Model    string    `json:"model"`
}

type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

func handleChatCompletion(w http.ResponseWriter, r *http.Request) {
    var req ChatCompletionRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 分析最后一条用户消息
    lastMessage := req.Messages[len(req.Messages)-1]

    // 根据消息内容决定响应模式
    response := generateResponse(lastMessage, req.Tools)

    json.NewEncoder(w).Encode(response)
}

func generateResponse(msg Message, tools []Tool) ChatCompletionResponse {
    content := strings.ToLower(msg.Content)

    // 简单对话: 不包含"攻击"、"投骰"等关键词
    if !containsKeywords(content, []string{"攻击", "投骰", "移动", "创建"}) {
        return ChatCompletionResponse{
            Choices: []Choice{
                {
                    Message: Message{
                        Role:    "assistant",
                        Content: "你好,冒险者!我是你的地下城主。",
                    },
                },
            },
        }
    }

    // 工具调用: 包含工具相关关键词
    if strings.Contains(content, "攻击") {
        return ChatCompletionResponse{
            Choices: []Choice{
                {
                    Message: Message{
                        Role:    "assistant",
                        Content: "",
                        ToolCalls: []ToolCall{
                            {
                                ID:   "call-001",
                                Type: "function",
                                Function: FunctionCall{
                                    Name: "resolve_attack",
                                    Arguments: `{"attacker_id":"char-001","target_id":"goblin-001","attack_type":"melee"}`,
                                },
                            },
                        },
                    },
                },
            },
        }
    }

    // 默认响应
    return simpleResponse("我理解了,请继续。")
}
```

**测试数据**: `mock/fixtures/llm_responses.go`

```go
package fixtures

var (
    // 简单对话响应
    SimpleChatResponse = ChatCompletionResponse{
        Choices: []Choice{
            {
                Message: Message{
                    Role:    "assistant",
                    Content: "你好,冒险者!有什么可以帮你的吗?",
                },
            },
        },
        Usage: Usage{
            PromptTokens:     50,
            CompletionTokens: 20,
            TotalTokens:      70,
        },
    }

    // 攻击工具调用响应
    AttackToolCallResponse = ChatCompletionResponse{
        Choices: []Choice{
            {
                Message: Message{
                    Role:    "assistant",
                    Content: "",
                    ToolCalls: []ToolCall{
                        {
                            ID:   "call-001",
                            Type: "function",
                            Function: FunctionCall{
                                Name:      "resolve_attack",
                                Arguments: `{"attacker_id":"char-001","target_id":"goblin-001","attack_type":"melee"}`,
                            },
                        },
                    },
                },
            },
        },
    }

    // 多工具调用响应
    MultiToolCallResponse = ChatCompletionResponse{
        Choices: []Choice{
            {
                Message: Message{
                    Role:    "assistant",
                    Content: "",
                    ToolCalls: []ToolCall{
                        {
                            ID:   "call-001",
                            Type: "function",
                            Function: FunctionCall{
                                Name:      "roll_dice",
                                Arguments: `{"dice_type":"d20","reason":"initiative"}`,
                            },
                        },
                        {
                            ID:   "call-002",
                            Type: "function",
                            Function: FunctionCall{
                                Name:      "resolve_attack",
                                Arguments: `{"attacker_id":"char-001","target_id":"goblin-001"}`,
                            },
                        },
                    },
                },
            },
        },
    }
)
```

### 测试工具函数

**文件位置**: `tests/testutil/`

```go
// tests/testutil/setup.go
package testutil

// SetupTestEnvironment 启动测试环境
func SetupTestEnvironment() *TestEnv {
    // 启动Mock MCP Server
    mcpServer := startMockMCPServer()

    // 启动Mock LLM Server
    llmServer := startMockLLMServer()

    // 启动PostgreSQL Docker容器
    db := startTestDatabase()

    // 启动Redis Docker容器
    cache := startTestRedis()

    return &TestEnv{
        MCPServerURL: mcpServer.URL,
        LLMServerURL: llmServer.URL,
        DB:           db,
        Cache:        cache,
    }
}

// CleanupTestEnvironment 清理测试环境
func CleanupTestEnvironment(env *TestEnv) {
    env.DB.Close()
    env.Cache.Close()
    env.MCPServer.Close()
    env.LLMServer.Close()
}

// CreateTestSession 创建测试会话
func CreateTestSession(db *sql.DB) *models.Session {
    session := &models.Session{
        ID:         uuid.New(),
        Name:       "测试会话",
        CreatorID:  uuid.New(),
        Status:     "active",
        MaxPlayers: 5,
        CreatedAt:  time.Now(),
    }

    // 保存到数据库
    err := session.Save(db)
    if err != nil {
        panic(err)
    }

    return session
}
```

### 测试环境启动脚本

**文件**: `scripts/test.bat` (Windows)

```batch
@echo off
echo ========================================
echo MCP Client Test Environment
echo ========================================

echo.
echo [1/4] Starting PostgreSQL...
net start postgresql-x64-15 > nul 2>&1
if %errorlevel% neq 0 (
    echo PostgreSQL already running or failed to start
) else (
    echo PostgreSQL started successfully
)

echo.
echo [2/4] Starting Redis (if enabled)...
net start Redis > nul 2>&1
if %errorlevel% neq 0 (
    echo Redis not installed or already running
) else (
    echo Redis started successfully
)

echo.
echo [3/4] Starting Mock MCP Server...
start "Mock MCP Server" cmd /k "cd /d %~dp0..\mock\mcp_server && echo Mock MCP Server running on http://localhost:9000 && go run main.go"
timeout /t 2 /nobreak > nul

echo.
echo [4/4] Starting Mock LLM Server...
start "Mock LLM Server" cmd /k "cd /d %~dp0..\mock\llm_server && echo Mock LLM Server running on http://localhost:9001 && go run main.go"
timeout /t 2 /nobreak > nul

echo.
echo ========================================
echo Test Environment Ready!
echo ========================================
echo - Mock MCP Server: http://localhost:9000
echo - Mock LLM Server: http://localhost:9001
echo - PostgreSQL: localhost:5432
echo - Redis: localhost:6379
echo.
echo Press any key to run tests...
pause > nul

echo.
echo Running tests...
go test -v ./tests/integration/...

echo.
echo Tests completed!
pause
```

**文件**: `scripts/test.sh` (Linux/macOS)

```bash
#!/bin/bash

echo "========================================"
echo "MCP Client Test Environment"
echo "========================================"

echo ""
echo "[1/4] Checking PostgreSQL..."
pg_isready -U postgres > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "PostgreSQL is running"
else
    echo "Please start PostgreSQL"
    exit 1
fi

echo ""
echo "[2/4] Checking Redis..."
redis-cli ping > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "Redis is running"
else
    echo "Redis not running (optional)"
fi

echo ""
echo "[3/4] Starting Mock MCP Server..."
cd mock/mcp_server
go run main.go &
MCP_PID=$!
echo "Mock MCP Server running on http://localhost:9000 (PID: $MCP_PID)"
cd ../..

sleep 2

echo ""
echo "[4/4] Starting Mock LLM Server..."
cd mock/llm_server
go run main.go &
LLM_PID=$!
echo "Mock LLM Server running on http://localhost:9001 (PID: $LLM_PID)"
cd ../..

sleep 2

echo ""
echo "========================================"
echo "Test Environment Ready!"
echo "========================================"
echo "- Mock MCP Server: http://localhost:9000"
echo "- Mock LLM Server: http://localhost:9001"
echo "- PostgreSQL: localhost:5432"
echo ""

echo "Running tests..."
go test -v ./tests/integration/...

echo ""
echo "Stopping Mock servers..."
kill $MCP_PID $LLM_PID 2>/dev/null

echo "Tests completed!"
```

### PowerShell 测试脚本 (高级)

**文件**: `scripts/test.ps1`

```powershell
#Requires -Version 5.1

param(
    [switch]$Verbose,
    [switch]$Coverage,
    [string]$TestPath = "./tests/integration/..."
)

function Start-ServiceIfNotRunning {
    param([string]$ServiceName)

    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service -and $service.Status -ne 'Running') {
        Write-Host "Starting $ServiceName..." -ForegroundColor Yellow
        Start-Service -Name $ServiceName
        Start-Sleep -Seconds 2
    }
    return $service.Status -eq 'Running'
}

function Start-MockServer {
    param(
        [string]$Name,
        [string]$Path,
        [string]$Port
    )

    Write-Host "Starting $Name on port $Port..." -ForegroundColor Cyan

    $process = Start-Process -FilePath "go" `
        -ArgumentList "run", "main.go" `
        -WorkingDirectory $Path `
        -WindowStyle Normal `
        -PassThru

    Start-Sleep -Seconds 2
    return $process
}

# Main
Write-Host "========================================" -ForegroundColor Green
Write-Host "MCP Client Test Environment" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Start PostgreSQL
Write-Host "[1/4] Checking PostgreSQL..." -ForegroundColor Yellow
$pgRunning = Start-ServiceIfNotRunning -ServiceName "postgresql-x64-15"
if ($pgRunning) {
    Write-Host "PostgreSQL is running" -ForegroundColor Green
} else {
    Write-Host "Failed to start PostgreSQL" -ForegroundColor Red
    exit 1
}

# Start Redis (optional)
Write-Host ""
Write-Host "[2/4] Checking Redis..." -ForegroundColor Yellow
$redisRunning = Start-ServiceIfNotRunning -ServiceName "Redis"
if ($redisRunning) {
    Write-Host "Redis is running" -ForegroundColor Green
} else {
    Write-Host "Redis not installed (optional)" -ForegroundColor Gray
}

# Start Mock MCP Server
Write-Host ""
Write-Host "[3/4] Starting Mock MCP Server..." -ForegroundColor Yellow
$mcpProcess = Start-MockServer -Name "Mock MCP Server" -Path "mock\mcp_server" -Port "9000"
Write-Host "Mock MCP Server started (PID: $($mcpProcess.Id))" -ForegroundColor Green

# Start Mock LLM Server
Write-Host ""
Write-Host "[4/4] Starting Mock LLM Server..." -ForegroundColor Yellow
$llmProcess = Start-MockServer -Name "Mock LLM Server" -Path "mock\llm_server" -Port "9001"
Write-Host "Mock LLM Server started (PID: $($llmProcess.Id))" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Running Tests..." -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

# Run tests
$testArgs = @("test", "-v")
if ($Verbose) { $testArgs += "-verbose" }
if ($Coverage) { $testArgs += "-coverprofile=coverage.out" }
$testArgs += $TestPath

& go @testArgs

$testExitCode = $LASTEXITCODE

# Cleanup
Write-Host ""
Write-Host "Stopping Mock servers..." -ForegroundColor Yellow
Stop-Process -Id $mcpProcess.Id -Force -ErrorAction SilentlyContinue
Stop-Process -Id $llmProcess.Id -Force -ErrorAction SilentlyContinue

Write-Host ""
if ($testExitCode -eq 0) {
    Write-Host "All tests passed!" -ForegroundColor Green
} else {
    Write-Host "Some tests failed!" -ForegroundColor Red
}

exit $testExitCode
```

---

## 开发阶段划分

### 阶段 0: 项目初始化 (预计 1 天)

**目标**：搭建项目基础结构,配置开发环境

**任务清单**：

1. **创建项目结构**
   - 初始化 Go module
   - 创建目录结构
   - 配置 `.gitignore`

2. **配置管理**
   - 实现 `internal/config/config.go`
   - 实现 `internal/config/loader.go`
   - 创建 `config.yaml` 模板

3. **基础中间件**
   - Request ID 生成器
   - 日志中间件
   - panic 恢复中间件

4. **错误定义**
   - 实现 `pkg/errors/errors.go`
   - 定义所有业务错误码

**可交付物**：
- ✅ 项目能够启动并读取配置文件
- ✅ HTTP server 能够启动(空路由)
- ✅ 请求日志正常输出
- ✅ 配置热加载(可选)

**测试要求**：
- [ ] 配置加载单元测试
- [ ] 中间件单元测试
- [ ] 项目启动健康检查

---

### 阶段 1: 最小会话管理系统 (预计 2-3 天)

**目标**：实现基础的会话管理功能,不涉及 LLM 和 MCP

**任务清单**：

1. **数据模型定义** (`internal/models/`)
   - `session.go`: Session 结构体
   - `message.go`: Message 结构体
   - 实现 JSON 序列化/反序列化

2. **存储层实现** (`internal/store/`)
   - 定义存储接口 `interface.go`
   - 实现 PostgreSQL 适配器
   - 实现 Redis 缓存层
   - 数据库迁移脚本

3. **Session Handler** (`internal/api/handler/session.go`)
   - `POST /api/sessions` - 创建会话
   - `GET /api/sessions` - 列出会话
   - `GET /api/sessions/{id}` - 获取会话详情
   - `DELETE /api/sessions/{id}` - 删除会话(软删除)

4. **基础路由** (`internal/api/router.go`)
   - 注册 session 路由
   - 应用中间件链

**可交付物**：
- ✅ 完整的会话 CRUD API
- ✅ 会话数据持久化到数据库
- ✅ 支持多会话创建和查询
- ✅ 软删除机制

**API 测试**：
```bash
# 创建会话
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"name":"失落的地牢","creator_id":"user-123","max_players":5}'

# 查询会话列表
curl http://localhost:8080/api/sessions?user_id=user-123

# 获取会话详情
curl http://localhost:8080/api/sessions/{session_id}

# 删除会话
curl -X DELETE http://localhost:8080/api/sessions/{session_id}
```

**测试要求**：
- [ ] 存储层单元测试 (Mock 数据库)
- [ ] Handler 集成测试 (测试数据库)
- [ ] API 端到端测试
- [ ] 并发创建会话测试

---

### 阶段 2: MCP 客户端实现 + Mock MCP Server (预计 3-4 天)

**目标**：实现与 MCP Server 的通信,同时实现 Mock MCP Server 用于测试

**任务清单**：

1. **Mock MCP Server 实现** (`mock/mcp_server/`)
   - `main.go`: HTTP 服务器启动
   - `handlers.go`:
     - `handleGetState`: 返回会话状态
     - `handleToolCall`: 执行工具调用
     - `handleCreateSession`: 创建会话
   - `data.go`:
     - SessionState 结构体
     - 内存存储 (map[string]*SessionState)
     - 默认测试数据生成

2. **MCP 协议编解码** (`internal/client/mcp/protocol.go`)
   - 实现 JSON-RPC 2.0 协议封装
   - 请求/响应结构定义
   - 错误处理

3. **MCP HTTP 客户端** (`internal/client/mcp/client.go`)
   - 实现 MCPClient 接口
   - HTTP 请求封装
   - 连接池管理

4. **重试机制** (`internal/client/mcp/retry.go`)
   - 可重试错误判断
   - 指数退避重试
   - 最大重试次数限制

5. **测试数据 Fixtures** (`mock/fixtures/mcp_responses.go`)
   - DefaultSessionState: 默认会话状态
   - AttackToolResponse: 攻击工具响应
   - AttackEvents: 攻击事件列表
   - 各种工具的预设响应

6. **Query Handler** (`internal/api/handler/query.go`)
   - `GET /api/sessions/{id}/state` - 查询会话状态
   - `GET /api/sessions/{id}/characters` - 查询角色列表
   - `GET /api/sessions/{id}/combat` - 查询战斗状态

**可交付物**：
- ✅ Mock MCP Server 能够独立运行
- ✅ MCP 客户端能够调用 Mock MCP Server
- ✅ 支持获取会话状态
- ✅ 支持调用工具(暂不用)
- ✅ 错误处理和重试机制

**API 测试**：
```bash
# 查询会话状态 (通过 MCP Server)
curl http://localhost:8080/api/sessions/{session_id}/state

# 响应示例
{
  "session_id": "uuid-xxx",
  "location": "地下城入口",
  "game_time": "2024-01-01T12:00:00Z",
  "characters": [...],
  "combat": null
}
```

**测试要求**：
- [ ] Mock MCP Server 单元测试
- [ ] MCP 客户端单元测试 (使用 Mock HTTP)
- [ ] 协议编解码测试
- [ ] 重试策略测试
- [ ] 与 Mock MCP Server 集成测试
- [ ] Query API 端到端测试 (使用 Mock MCP Server)

**测试流程**：
```bash
# 1. 启动 Mock MCP Server
cd mock/mcp_server
go run main.go
# 监听 :9000

# 2. 在另一个终端运行测试
DATABASE_URL=postgres://test:test@localhost:5433/mcp_test \
MCP_SERVER_URL=http://localhost:9000 \
go test -v ./internal/client/mcp/... ./internal/api/handler/...
```

---

### 阶段 3: LLM 客户端实现 + Mock LLM Server (预计 2-3 天)

**目标**：实现 LLM 调用功能,同时实现 Mock LLM Server 用于测试

**任务清单**：

1. **Mock LLM Server 实现** (`mock/llm_server/`)
   - `main.go`: HTTP 服务器启动(模拟OpenAI API格式)
   - `handlers.go`:
     - `handleChatCompletion`: 处理聊天请求
     - `handleStreamCompletion`: 处理流式请求(可选)
     - `generateResponse`: 根据消息内容生成响应
   - 智能响应逻辑:
     - 简单对话: 返回文本
     - 包含"攻击": 返回 resolve_attack tool_call
     - 包含"投骰": 返回 roll_dice tool_call
     - 包含"移动": 返回 move_character tool_call

2. **LLM 接口定义** (`internal/client/llm/client.go`)
   - 定义 LLMClient 接口
   - ChatCompletion 方法
   - StreamCompletion 方法(可选)

3. **OpenAI 实现** (`internal/client/llm/openai.go`)
   - 实现 OpenAI API 调用
   - Chat Completion 请求/响应
   - Token 使用统计

4. **重试和错误处理** (`internal/client/llm/retry.go`)
   - 可重试错误 (429, 5xx)
   - 不可重试错误 (400, 401)
   - 指数退避重试

5. **测试数据 Fixtures** (`mock/fixtures/llm_responses.go`)
   - SimpleChatResponse: 简单对话响应
   - AttackToolCallResponse: 攻击工具调用
   - MultiToolCallResponse: 多工具调用
   - 各种场景的预设响应

6. **简单 Chat Handler** (`internal/api/handler/chat.go`)
   - `POST /api/sessions/{id}/chat` - 基础聊天
   - 暂时不调用工具,只返回 LLM 文本响应
   - 支持配置使用真实OpenAI或Mock服务器

**可交付物**：
- ✅ Mock LLM Server 能够独立运行
- ✅ 能够调用 Mock LLM Server (测试环境)
- ✅ 能够调用真实 OpenAI API (生产环境)
- ✅ 支持 Chat Completion
- ✅ 错误处理和重试
- ✅ 简单的聊天 API (无工具调用)

**API 测试**：
```bash
# 简单聊天 (暂不调用工具)
curl -X POST http://localhost:8080/api/sessions/{session_id}/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"你好"}'

# 响应示例
{
  "response": "你好,冒险者!我是你的地下城主。",
  "usage": {
    "prompt_tokens": 50,
    "completion_tokens": 20,
    "total_tokens": 70
  }
}
```

**测试要求**：
- [ ] Mock LLM Server 单元测试
- [ ] LLM 客户端单元测试 (使用 Mock LLM Server)
- [ ] OpenAI API 调用测试 (使用真实测试 key, 可选)
- [ ] 重试机制测试
- [ ] Chat API 集成测试 (使用 Mock LLM Server)

**测试流程**：
```bash
# 1. 启动 Mock LLM Server
cd mock/llm_server
go run main.go
# 监听 :9001, 模拟 OpenAI API 格式

# 2. 配置使用 Mock 服务器
export LLM_PROVIDER=mock
export LLM_SERVER_URL=http://localhost:9001

# 3. 运行测试
DATABASE_URL=postgres://test:test@localhost:5433/mcp_test \
LLM_SERVER_URL=http://localhost:9001 \
go test -v ./internal/client/llm/... ./internal/api/handler/...
```

**配置支持** (`config.test.yaml`):
```yaml
llm:
  provider: "mock"  # 或 "openai"
  api_key: "test-key"
  model: "gpt-4"
  base_url: "http://localhost:9001"  # Mock服务器地址
  max_retries: 3
  timeout: 30
```

---

### 阶段 4: 消息存储和上下文构建 (预计 2-3 天)

**目标**：实现对话历史存储和上下文构建器

**任务清单**：

1. **消息存储增强** (`internal/store/message.go`)
   - 保存消息到数据库
   - 按时间戳查询历史消息
   - 按 session_id 索引
   - Token 计算(可选)

2. **ContextBuilder 实现** (`internal/orchestrator/context_builder.go`)
   - MessageLoader: 加载最近 50 条消息
   - StateFetcher: 从 MCP Server 获取状态
   - SystemPromptBuilder: 构建系统提示
   - TokenBudgetManager: Token 预算管理

3. **Chat Handler 增强**
   - 加载历史消息
   - 构建完整对话上下文
   - 传递给 LLM

**可交付物**：
- ✅ 对话历史持久化
- ✅ LLM 能够看到历史消息
- ✅ 系统提示动态生成
- ✅ 支持多轮对话

**测试要求**：
- [ ] 消息存储单元测试
- [ ] ContextBuilder 单元测试
- [ ] 多轮对话集成测试
- [ ] Token 预算管理测试

---

### 阶段 5: 工具调用循环 (预计 3-4 天)

**目标**：实现 LLM 工具调用和 MCP 执行的完整循环,使用 Mock 服务器进行完整测试

**任务清单**：

1. **ToolCoordinator** (`internal/orchestrator/tool_coordinator.go`)
   - FormatConverter: LLM 格式 → MCP 格式
   - MCPInvoker: 调用 MCP 工具
   - ResultParser: 解析 MCP 响应
   - EventCollector: 收集事件

2. **主处理循环** (`internal/orchestrator/orchestrator.go`)
   - ProcessChatMessage 核心逻辑
   - LLM → Tool → LLM 循环
   - 最大迭代次数限制(10次)
   - 工具结果作为 tool message 发送

3. **ResponseGenerator** (`internal/orchestrator/response_generator.go`)
   - TextExtractor: 提取最终文本
   - StateChangeExtractor: 提取状态变更
   - TurnInfoExtractor: 提取回合信息

4. **Chat Handler 完善**
   - 完整的聊天流程
   - 支持工具调用
   - 返回完整响应结构

5. **集成测试场景** (`tests/integration/chat_test.go`)
   - 场景1: 简单对话(无工具调用)
   - 场景2: 单次工具调用(攻击)
   - 场景3: 多次工具调用(投骰+攻击)
   - 场景4: 工具调用循环超限

**可交付物**：
- ✅ 完整的聊天功能 (带工具调用)
- ✅ Mock LLM Server 可以返回 tool_calls
- ✅ Mock MCP Server 可以处理工具调用
- ✅ LLM 可以调用 MCP 工具
- ✅ 工具结果返回给 LLM
- ✅ 多轮工具调用支持
- ✅ 最大迭代保护

**API 测试**：
```bash
# 完整聊天 (带工具调用)
curl -X POST http://localhost:8080/api/sessions/{session_id}/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"我要攻击那个哥布林"}'

# 响应示例
{
  "response": "你冲向哥布林,挥剑砍中,造成 8 点伤害!",
  "state_changes": {
    "goblin_1": {"hp": 5, "hp_max": 13}
  },
  "requires_roll": false,
  "usage": {...}
}
```

**测试要求**：
- [ ] ToolCoordinator 单元测试 (Mock MCP Server)
- [ ] 主处理循环单元测试 (Mock LLM + Mock MCP)
- [ ] 工具调用集成测试 (使用 Mock 服务器)
- [ ] 最大迭代保护测试
- [ ] 端到端聊天测试 (使用 Mock 服务器)

**完整测试流程**：
```bash
# 1. 启动测试环境 (包含所有Mock服务器)
make test-mocks

# 2. 运行集成测试
make test-with-mocks

# 3. 停止Mock服务器
make stop-mocks
```

**集成测试示例** (`tests/integration/chat_test.go`):
```go
func TestChatIntegration_SimpleConversation(t *testing.T) {
    // 使用Mock服务器
    env := setup.TestEnvironment()
    defer testutil.CleanupTestEnvironment(env)

    // 创建测试会话
    session := testutil.CreateTestSession(env.DB)

    // 发送简单消息
    req := ChatRequest{
        SessionID: session.ID,
        Message:   "你好",
    }

    resp, err := HandleChat(req, env)
    assert.NoError(t, err)
    assert.Contains(t, resp.Response, "你好,冒险者")
    assert.Empty(t, resp.StateChanges) // 无工具调用
}

func TestChatIntegration_ToolCalling(t *testing.T) {
    env := testutil.SetupTestEnvironment()
    defer testutil.CleanupTestEnvironment(env)

    session := testutil.CreateTestSession(env.DB)

    // 发送攻击消息
    req := ChatRequest{
        SessionID: session.ID,
        Message:   "我要攻击那个哥布林",
    }

    resp, err := HandleChat(req, env)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.StateChanges) // 有工具调用
    assert.Contains(t, resp.Response, "哥布林")
}

func TestChatIntegration_MultiToolCalling(t *testing.T) {
    env := testutil.SetupTestEnvironment()
    defer testutil.CleanupTestEnvironment(env)

    session := testutil.CreateTestSession(env.DB)

    // 发送需要多次工具调用的消息
    req := ChatRequest{
        SessionID: session.ID,
        Message:   "我先投骰子决定先攻，然后攻击哥布林",
    }

    resp, err := HandleChat(req, env)
    assert.NoError(t, err)
    // 应该调用两次工具: roll_dice + resolve_attack
}
```

---

### 阶段 6: WebSocket 事件推送 (预计 3-4 天)

**目标**：实现实时事件推送功能

**任务清单**：

1. **EventDispatcher** (`internal/dispatcher/`)
   - dispatcher.go: 事件分发器
   - queue.go: 事件队列(缓冲)
   - broadcaster.go: WebSocket 广播

2. **WebSocket Manager** (`internal/api/websocket/`)
   - manager.go: 连接管理
   - conn.go: 连接封装
   - hub.go: 连接集线器
   - ReadPump/WritePump: 读写泵
   - 心跳机制

3. **事件收集和推送**
   - ToolCoordinator 收集 MCP 事件
   - EventDispatcher 推送到前端
   - 按 session_id 过滤

4. **WebSocket 路由**
   - `WS /api/sessions/{id}/ws`
   - 连接认证(可选)

**可交付物**：
- ✅ WebSocket 连接管理
- ✅ 实时事件推送
- ✅ 心跳保活机制
- ✅ 按 session 隔离推送

**测试**：
```javascript
// 前端测试代码
const ws = new WebSocket('ws://localhost:8080/api/sessions/{session_id}/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data.event);
};

// 预期事件: dice.rolled, combat.attack_resolved, character.hp_changed
```

**测试要求**：
- [ ] WebSocket 连接测试
- [ ] 事件推送单元测试
- [ ] 心跳机制测试
- [ ] 并发连接测试
- [ ] 事件过滤测试

---

### 阶段 7: 高级功能和优化 (预计 2-3 天)

**目标**：完善系统功能,添加优化和增强

**任务清单**：

1. **会话生命周期完善**
   - 归档会话 `POST /api/sessions/{id}/archive`
   - 恢复会话 `POST /api/sessions/{id}/restore`
   - 定期清理软删除会话 (Cron Job)

2. **流式响应(可选)**
   - SSE (Server-Sent Events)
   - LLM 流式输出

3. **认证和授权(可选)**
   - JWT 认证
   - API Key 认证
   - 权限控制

4. **限流和防护**
   - Rate Limit 中间件
   - 请求配额
   - 防止滥用

5. **监控和指标**
   - Prometheus 指标
   - 请求延迟监控
   - LLM 调用统计
   - MCP 调用统计

6. **性能优化**
   - 数据库查询优化
   - 缓存优化
   - 连接池调优
   - pprof 性能分析

**可交付物**：
- ✅ 完整的会话管理
- ✅ 流式响应(可选)
- ✅ 认证授权(可选)
- ✅ 监控指标
- ✅ 性能优化

---

### 阶段 8: 跨平台部署和文档 (预计 1-2 天)

**目标**：准备跨平台部署,完善文档

**任务清单**：

1. **跨平台构建支持**
   - Windows 构建: `build.bat`
   - Linux/macOS 构建: `Makefile`
   - 交叉编译脚本 (可选)
   - 平台特定二进制文件

2. **Docker 部署** (可选,用于 Linux/macOS)
   - 编写 Dockerfile
   - 多阶段构建优化
   - Docker Compose 配置
   - 包含依赖服务 (PostgreSQL, Redis)

3. **部署脚本**
   - `scripts/build.bat`: Windows 构建脚本
   - `scripts/build.sh`: Linux/macOS 构建脚本
   - `scripts/deploy.bat`: Windows 部署脚本
   - `scripts/deploy.sh`: Linux/macOS 部署脚本
   - 数据库迁移脚本

4. **文档完善**
   - API 文档 (Swagger/OpenAPI)
   - Windows 部署文档
   - Linux/macOS 部署文档
   - 运维文档
   - 故障排查指南

5. **CI/CD 配置** (可选)
   - GitHub Actions / GitLab CI
   - 自动测试 (多平台)
   - 自动构建 (Windows + Linux)
   - 自动部署

**可交付物**：
- ✅ Windows 可执行文件 (.exe)
- ✅ Linux/macOS 二进制文件
- ✅ Docker 镜像 (可选)
- ✅ 跨平台部署脚本
- ✅ 完整文档
- ✅ CI/CD 流程(可选)

---

## 开发流程规范

### 日常开发流程

1. **拉取最新代码**
   ```powershell
   # Windows PowerShell
   git pull origin main

   # Linux/macOS
   git pull origin main
   ```

2. **创建功能分支**
   ```powershell
   git checkout -b feature/feature-name
   ```

3. **编写代码**
   - 遵循 CODE_STANDARDS.md
   - 编写单元测试
   - 运行 `go fmt` 和 `go vet`

4. **本地测试**
   ```powershell
   # Windows
   .\scripts\test.bat

   # Linux/macOS
   make test
   # 或
   ./scripts/test.sh

   # 运行特定测试
   go test ./internal/store/...
   ```

5. **提交代码**
   ```powershell
   git add .
   git commit -m "feat: 添加消息存储功能"
   ```

6. **推送并创建 PR**
   ```powershell
   git push origin feature/feature-name
   ```

7. **代码审查**
   - 至少一人审查
   - 检查代码规范
   - 检查测试覆盖

8. **合并到主分支**
   - 通过 CI 检查
   - 合并到 main
   - 删除功能分支

### 测试策略

**单元测试**：
- 每个包都有对应的 `_test.go` 文件
- 使用表驱动测试
- Mock 外部依赖 (数据库、HTTP 客户端)
- 核心逻辑覆盖率 ≥ 80%

**集成测试**：
- 测试组件间交互
- 使用测试数据库
- 使用 Mock MCP/LLM 服务
- 端到端 API 测试

**性能测试** (可选)：
- Benchmark 测试
- 压力测试
- 并发测试

### 构建脚本

#### Windows 构建脚本 (`build.bat`)

```batch
@echo off
setlocal EnableDelayedExpansion

REM 默认目标
set TARGET=build

REM 解析命令行参数
if "%1" neq "" set TARGET=%1

REM 构建
if "%TARGET%"=="build" (
    echo Building MCP Client...
    if not exist bin mkdir bin
    go build -o bin\mcp-client.exe cmd\mcp-client\main.go
    if %errorlevel% equ 0 (
        echo Build successful: bin\mcp-client.exe
    ) else (
        echo Build failed!
        exit /b 1
    )
)

REM 运行
if "%TARGET%"=="run" (
    echo Running MCP Client...
    go run cmd\mcp-client\main.go
)

REM 测试
if "%TARGET%"=="test" (
    echo Running tests...
    go test -v -race -cover ./internal/...
)

REM 集成测试
if "%TARGET%"=="test-integration" (
    echo Running integration tests...
    call scripts\test.bat
)

REM 覆盖率
if "%TARGET%"=="coverage" (
    echo Generating coverage report...
    go test -coverprofile=coverage.out ./internal/...
    go tool cover -html=coverage.out -o coverage.html
    echo Coverage report: coverage.html
)

REM 代码检查
if "%TARGET%"=="lint" (
    echo Running code checks...
    go vet ./...
)

REM 格式化
if "%TARGET%"=="fmt" (
    echo Formatting code...
    go fmt ./...
)

REM 清理
if "%TARGET%"=="clean" (
    echo Cleaning build files...
    if exist bin rmdir /s /q bin
    if exist coverage.out del /q coverage.out
    if exist coverage.html del /q coverage.html
)

REM 帮助
if "%TARGET%"=="help" (
    echo Available targets:
    echo   build           - Build the application
    echo   run             - Run the application
    echo   test            - Run unit tests
    echo   test-integration - Run integration tests
    echo   coverage        - Generate coverage report
    echo   lint            - Run code checks
    echo   fmt             - Format code
    echo   clean           - Clean build files
    echo   help            - Show this help
)

endlocal
```

#### Linux/macOS Makefile

```makefile
.PHONY: help build run test test-integration coverage lint fmt clean

# 默认目标
.DEFAULT_GOAL := help

## build: 构建程序
build:
	@echo "Building MCP Client..."
	@mkdir -p bin
	@go build -o bin/mcp-client cmd/mcp-client/main.go
	@echo "Build successful: bin/mcp-client"

## run: 运行程序
run:
	@echo "Running MCP Client..."
	@go run cmd/mcp-client/main.go

## test: 运行单元测试
test:
	@echo "Running tests..."
	@go test -v -race -cover ./internal/...

## test-integration: 运行集成测试
test-integration:
	@echo "Running integration tests..."
	@./scripts/test.sh

## coverage: 生成测试覆盖率报告
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: 代码检查
lint:
	@echo "Running code checks..."
	@go vet ./...

## fmt: 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## clean: 清理构建文件
clean:
	@echo "Cleaning build files..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

## help: 显示帮助信息
help:
	@echo "Available targets:"
	@grep -E '^## ' ${MAKEFILE_LIST} | sed 's/## /  /'
```

#### 使用方式

**Windows**:
```powershell
# 构建
.\build.bat build

# 运行
.\build.bat run

# 测试
.\build.bat test

# 集成测试
.\build.bat test-integration

# 查看所有命令
.\build.bat help
```

**Linux/macOS**:
```bash
# 构建
make build

# 运行
make run

# 测试
make test

# 集成测试
make test-integration

# 查看所有命令
make help
```

---

## 持续集成检查清单

每个阶段完成前,必须确认以下检查项:

### 代码质量
- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 检查
- [ ] 代码通过 `golangci-lint` 检查(可选)
- [ ] 无 TODO 或 FIXME (或已记录到问题追踪)
- [ ] 所有导出函数有注释
- [ ] 错误处理完整

### 测试覆盖
- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 所有测试通过 (`make test`)
- [ ] 集成测试通过 (`make test-integration`)
- [ ] 无竞态条件 (`go test -race`)

### 功能验证
- [ ] 功能按需求实现
- [ ] API 端到端测试通过
- [ ] 错误场景测试通过
- [ ] 边界条件测试通过

### 文档更新
- [ ] API 文档更新
- [ ] 配置文档更新
- [ ] 部署文档更新
- [ ] 变更日志更新

### 性能和安全
- [ ] 无内存泄漏
- [ ] 无 SQL 注入风险
- [ ] 无敏感信息泄露
- [ ] 资源正确释放 (连接、文件等)

---

## 预计时间线

| 阶段 | 任务 | 预计天数 | 累计天数 |
|----|----|-------|--------|
| 0  | 项目初始化 | 1 | 1 |
| 1  | 会话管理 | 2-3 | 3-4 |
| 2  | MCP 客户端 | 3-4 | 6-8 |
| 3  | LLM 客户端 | 2-3 | 8-11 |
| 4  | 消息存储和上下文 | 2-3 | 10-14 |
| 5  | 工具调用循环 | 3-4 | 13-18 |
| 6  | WebSocket 事件推送 | 3-4 | 16-22 |
| 7  | 高级功能和优化 | 2-3 | 18-25 |
| 8  | 部署和文档 | 1-2 | 19-27 |

**总计**：19-27 天 (约 4-5 周)

---

## 风险和依赖

### 外部依赖

1. **MCP Server**：需要 MCP Server 可用于集成测试
   - 缓解措施：使用 Mock Server

2. **OpenAI API**：需要 API Key 和配额
   - 缓解措施：使用 Mock 响应进行测试

3. **PostgreSQL/Redis**：需要数据库环境
   - 缓解措施：Docker Compose 自动启动

### 技术风险

1. **LLM 调用成本**：开发和测试期间会产生费用
   - 缓解措施：优先使用 Mock,使用较便宜的模型

2. **工具调用循环复杂度**：可能出现无限循环
   - 缓解措施：最大迭代次数限制,充分测试

3. **WebSocket 并发**：高并发下性能问题
   - 缓解措施：压力测试,性能优化

### 进度风险

1. **需求变更**：开发过程中需求可能调整
   - 缓解措施：灵活规划,预留缓冲时间

2. **技术难点**：某些技术实现可能比预期复杂
   - 缓解措施：提前技术预研,分阶段实现

---

## Mock 实现总结

### 为什么需要 Mock？

在当前开发场景中,**只开发 MCP Client 一个组件**,其他所有组件都尚未实现。因此必须通过 Mock 来模拟这些外部依赖:

1. **MCP Server** - 规则引擎,负责游戏状态管理
2. **LLM 服务** (OpenAI/Claude) - AI决策引擎
3. **前端 UI** - 用户交互界面

### Mock 架构全景

```
┌─────────────────────────────────────────────────────────────┐
│                  本地开发环境 (Windows)                        │
│                                                               │
│  独立进程:                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │Mock LLM Server│  │Mock MCP Server│  │PostgreSQL    │      │
│  │  localhost:  │  │  localhost:  │  │  localhost:  │      │
│  │  :9001       │  │  :9000       │  │  :5432       │      │
│  │  (go run)   │  │  (go run)   │  │  (服务)      │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                  │               │
│         │ HTTP            │ HTTP+JSON-RPC    │ SQL           │
│         │                 │                  │               │
└─────────┼─────────────────┼──────────────────┼───────────────┘
          │                 │                  │
          ▼                 ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                      MCP Client (被测系统)                    │
│                                                               │
│  internal/client/llm/  → 调用 Mock LLM Server                │
│  internal/client/mcp/  → 调用 Mock MCP Server                │
│  internal/store/       → 连接本地数据库                       │
│  internal/api/handler/ → HTTP API测试                        │
│                                                               │
└─────────────────────────────────────────────────────────────┘
          ▲
          │ HTTP 请求
          │
┌─────────┴───────────────────────────────────────────────────┐
│                    集成测试                                   │
│  tests/integration/chat_test.go                              │
│  或 scripts/test.bat (自动启动Mock服务器)                      │
│  - 启动Mock服务器 (go run)                                    │
│  - 创建测试数据                                                │
│  - 调用API端点                                                │
│  - 验证响应                                                    │
│  - 清理测试环境                                                │
└─────────────────────────────────────────────────────────────┘

跨平台部署 (可选):
┌─────────────────────────────────────────────────────────────┐
│              Docker 容器 (Linux/macOS)                         │
│  使用 deploy/docker-compose.yml 部署                          │
└─────────────────────────────────────────────────────────────┘
```

### Mock 实现的关键特性

#### 1. Mock LLM Server
- **智能响应**: 根据消息内容返回不同类型的响应
- **工具调用模拟**: 支持 `tool_calls` 格式
- **流式响应**: 可选支持SSE流式输出
- **错误模拟**: 可以模拟网络错误、超时等异常情况

**响应决策逻辑**:
```go
if 消息包含 "攻击" {
    return tool_call("resolve_attack")
}
if 消息包含 "投骰" {
    return tool_call("roll_dice")
}
if 消息包含 "移动" {
    return tool_call("move_character")
}
return text_response("简单对话")
```

#### 2. Mock MCP Server
- **状态管理**: 内存存储会话状态
- **工具执行**: 模拟各种工具的执行结果
- **事件生成**: 每次工具调用返回事件列表
- **版本控制**: 支持乐观锁版本号

**支持的工具**:
- `resolve_attack` - 攻击结算
- `roll_dice` - 投骰子
- `move_character` - 角色移动
- `create_character` - 创建角色
- `get_state` - 获取状态

### 测试覆盖矩阵

| 测试场景 | Mock LLM | Mock MCP | 测试数据库 | 测试类型 |
|---------|----------|----------|-----------|---------|
| 会话 CRUD | - | - | ✅ | 单元测试 |
| MCP客户端调用 | - | ✅ | ✅ | 集成测试 |
| LLM客户端调用 | ✅ | - | ✅ | 集成测试 |
| 简单对话 | ✅ | - | ✅ | 集成测试 |
| 工具调用 | ✅ | ✅ | ✅ | 集成测试 |
| 多轮工具调用 | ✅ | ✅ | ✅ | 集成测试 |
| WebSocket事件 | - | ✅ | ✅ | 集成测试 |
| 完整聊天流程 | ✅ | ✅ | ✅ | E2E测试 |

### Mock vs 真实服务切换

通过环境变量和配置文件,轻松切换Mock和真实服务:

**开发/测试环境** (`config.test.yaml`):
```yaml
llm:
  provider: "mock"
  base_url: "http://localhost:9001"

mcp:
  base_url: "http://localhost:9000"
```

**生产环境** (`config.yaml`):
```yaml
llm:
  provider: "openai"
  api_key: "${LLM_API_KEY}"
  base_url: "https://api.openai.com"

mcp:
  base_url: "${MCP_SERVER_URL}"
```

### Mock 开发优先级

**阶段 0-1**: 不需要Mock (只涉及数据库和API层)

**阶段 2**: **必须实现 Mock MCP Server**
- 实现基础HTTP接口
- 实现 `GET /state` 和 `POST /tools/call`
- 提供测试数据

**阶段 3**: **必须实现 Mock LLM Server**
- 实现OpenAI兼容接口
- 实现 `POST /v1/chat/completions`
- 支持工具调用响应

**阶段 4-8**: 扩展Mock服务器功能
- 增加更多工具支持
- 优化响应逻辑
- 添加错误模拟

### 成功标准

**Mock服务器必须满足**:
1. ✅ 能够独立启动和运行
2. ✅ 提供与真实服务兼容的API接口
3. ✅ 返回符合规范的数据格式
4. ✅ 支持并发请求
5. ✅ 有完善的错误处理
6. ✅ 易于配置和扩展

**测试必须满足**:
1. ✅ 所有测试使用Mock服务器
2. ✅ 测试覆盖率 ≥ 80%
3. ✅ 集成测试可一键运行
4. ✅ 测试结果稳定可靠
5. ✅ 测试执行时间合理 (<5分钟)

---

## 总结

本开发计划遵循以下核心原则:

1. **持续集成**：每个阶段都是可运行、可测试的完整程序
2. **代码规范**：严格遵循 CODE_STANDARDS.md 规范
3. **测试驱动**：单元测试 + 集成测试全覆盖
4. **增量开发**：从最小功能开始,逐步完善
5. **文档同步**：代码和文档同步更新

按照此计划执行,可以保证:
- ✅ 每个阶段结束都有可交付的软件
- ✅ 代码质量和测试覆盖率达标
- ✅ 团队协作顺畅
- ✅ 风险可控
- ✅ 最终交付高质量产品
