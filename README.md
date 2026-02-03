# DND MCP Client

D&D MCP Client 是一个轻量级的有状态协调层,用于管理 D&D 游戏会话和消息。

## 项目状态

当前版本: v0.1.0 (开发中)

开发进度详见: [doc/development_progress.md](doc/development_progress.md)

## 快速开始

### 前置要求

- Go 1.21+
- Redis 7.0+
- Docker Desktop (Windows)

### Windows 环境

#### 安装依赖

```powershell
# 克隆仓库
git clone https://github.com/dnd-mcp/client.git
cd client

# 下载依赖
go mod download
```

#### 启动 Redis

```powershell
# 使用 PowerShell 脚本启动 Redis
.\scripts\start-redis.ps1

# 或手动启动
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine
```

#### 构建项目

```powershell
# 使用 PowerShell 脚本构建
.\scripts\build.ps1

# 或手动构建
go build ./cmd/client
```

#### 运行测试

```powershell
# 使用 PowerShell 脚本测试
.\scripts\test.ps1

# 或手动测试
go test ./tests/unit/... -v
```

#### 使用 CLI 工具

```powershell
# 查看帮助
.\client.exe --help

# 创建会话
.\client.exe session create --name "测试会话" --creator "user-123" --mcp-url "http://localhost:9000"

# 查看会话列表
.\client.exe session list

# 保存消息
.\client.exe message save --session <session-id> --content "你好" --player-id "player-123"
```

### Linux/Mac 环境

#### 安装依赖

```bash
# 克隆仓库
git clone https://github.com/dnd-mcp/client.git
cd client

# 下载依赖
go mod download
```

#### 启动 Redis

```bash
# 使用 Docker 启动 Redis
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine
```

#### 构建项目

```bash
# 使用脚本构建
chmod +x ./scripts/build.sh
./scripts/build.sh

# 或手动构建
go build -o bin/dnd-client ./cmd/client
```

#### 运行测试

```bash
# 使用脚本测试
chmod +x ./scripts/test.sh
./scripts/test.sh

# 或手动测试
go test ./tests/unit/... -v
```

## 开发文档

- [规范](doc/规范.md) - 代码规范
- [详细设计](doc/DND_MCP_Client详细设计.md) - 技术设计文档
- [开发计划](doc/DND_MCP_Client_开发计划.md) - 开发路线图
- [开发任务一](doc/开发任务一.md) - 当前开发任务

## 测试

### Windows (PowerShell)

```powershell
# 运行所有测试
.\scripts\test.ps1

# 运行单元测试
go test ./tests/unit/... -v

# 运行集成测试 (需要 Redis)
$env:INTEGRATION_TEST="1"
go test ./tests/integration/... -tags=integration -v

# 查看测试覆盖率
go test ./tests/unit/... -cover
```

### Linux/Mac

```bash
# 运行所有测试
./scripts/test.sh

# 运行单元测试
go test ./tests/unit/... -v

# 运行集成测试
INTEGRATION_TEST=1 go test ./tests/integration/... -tags=integration -v

# 查看测试覆盖率
go test ./... -cover
```

## 项目结构

```
├── cmd/           # 主程序入口
├── internal/      # 私有代码
│   ├── models/    # 领域模型
│   ├── store/     # 数据存储层
│   ├── config/    # 配置管理
│   └── cli/       # 命令行工具
├── pkg/           # 公共库
├── tests/         # 测试文件
├── scripts/       # 构建脚本
└── doc/           # 开发文档
```

## License

MIT
