# Server 脚本说明

本目录包含 DND MCP Server 的部署和管理脚本。

## 环境要求

- **Go 1.24+**: 用于编译和运行服务
- **PostgreSQL 14+**: 本地安装的 PostgreSQL 数据库
  - 下载地址: https://www.postgresql.org/download/windows/
  - 安装后确保 `bin` 目录在 PATH 中或位于以下位置之一:
    - `C:\Program Files\PostgreSQL\18\bin`
    - `C:\Program Files\PostgreSQL\17\bin`
    - `C:\Program Files\PostgreSQL\16\bin`
    - `C:\Program Files\PostgreSQL\15\bin`
    - `C:\Program Files\PostgreSQL\14\bin`

## 脚本概览

| 脚本 | 用途 | 使用场景 |
|------|------|---------|
| `deploy.ps1` | 一键部署 | 首次部署或完整重新部署 |
| `stop.ps1` | 停止服务 | 停止运行中的服务 |
| `build.ps1` | 构建服务 | 仅编译二进制文件 |
| `start-postgres.ps1` | 管理 PostgreSQL | 启动/停止本地数据库 |
| `init-db.ps1` | 初始化数据库 | 创建数据库和用户 |
| `test.ps1` | 快速测试 | 运行单元测试和集成测试 |
| `test-all.ps1` | 完整测试 | 运行所有测试 |

## 快速开始

### 一键部署 Server

```powershell
# 完整部署（推荐首次使用）
.\scripts\deploy.ps1

# 跳过数据库设置（数据库已就绪）
.\scripts\deploy.ps1 -SkipDb

# 跳过构建（二进制已存在）
.\scripts\deploy.ps1 -SkipBuild

# 强制重建数据库
.\scripts\deploy.ps1 -Force

# 设置日志级别
.\scripts\deploy.ps1 -LogLevel debug
```

### 停止服务

```powershell
# 仅停止 Server
.\scripts\stop.ps1

# 停止并清理日志
.\scripts\stop.ps1 -Clean

# 停止所有服务（包括 Client、PostgreSQL、Redis）
.\scripts\stop.ps1 -All

# 仅停止数据库
.\scripts\stop.ps1 -Db
```

## 详细说明

### deploy.ps1 - 一键部署

自动化完成以下步骤：
1. 启动本地 PostgreSQL
2. 初始化数据库和用户
3. 构建服务二进制
4. 启动服务
5. 健康检查

**参数：**
- `-SkipDb`: 跳过数据库设置
- `-SkipBuild`: 跳过构建步骤
- `-Force`: 强制重建数据库
- `-LogLevel`: 日志级别（默认 info）

### stop.ps1 - 停止服务

**参数：**
- `-All`: 停止所有服务（Server、Client、PostgreSQL、Redis）
- `-Db`: 仅停止数据库
- `-Clean`: 清理日志文件

### start-postgres.ps1 - PostgreSQL 管理

**参数：**
- `-Action start`: 启动数据库（默认）
- `-Action stop`: 停止数据库
- `-Action status`: 查看状态
- `-Action reset`: 重置数据库（删除数据目录并重新初始化）
- `-Reset`: 启动时重置数据库

**示例：**
```powershell
# 启动 PostgreSQL
.\scripts\start-postgres.ps1

# 停止 PostgreSQL
.\scripts\start-postgres.ps1 -Action stop

# 查看状态
.\scripts\start-postgres.ps1 -Action status

# 完全重置（删除所有数据）
.\scripts\start-postgres.ps1 -Action reset
```

### init-db.ps1 - 数据库初始化

**参数：**
- `-PostgresHost`: 数据库主机（默认 localhost）
- `-PostgresPort`: 数据库端口（默认 5432）
- `-AdminUser`: 管理员用户（默认 postgres）
- `-AdminPassword`: 管理员密码（默认 postgres）
- `-DbName`: 数据库名（默认 dnd_server）
- `-DbUser`: 应用用户（默认 dnd）
- `-DbPassword`: 应用密码（默认 password）
- `-Force`: 强制重建数据库

**示例：**
```powershell
# 使用默认配置初始化
.\scripts\init-db.ps1

# 自定义配置
.\scripts\init-db.ps1 -DbName "my_dnd" -DbUser "myuser" -DbPassword "mypass"

# 强制重建
.\scripts\init-db.ps1 -Force
```

## 配置

### 环境变量

Server 启动时使用的环境变量：

```powershell
# PostgreSQL 配置
$env:POSTGRES_HOST = "localhost"
$env:POSTGRES_PORT = "5432"
$env:POSTGRES_USER = "postgres"
$env:POSTGRES_PASSWORD = "postgres"
$env:POSTGRES_DBNAME = "dnd_server"
$env:POSTGRES_SSLMODE = "disable"

# HTTP 配置
$env:HTTP_HOST = "0.0.0.0"
$env:HTTP_PORT = "8081"

# 日志配置
$env:LOG_LEVEL = "info"
$env:LOG_FORMAT = "text"
```

### .env 文件

可以在 `packages/server/.env` 文件中配置环境变量，deploy.ps1 会自动加载：

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DBNAME=dnd_server
LOG_LEVEL=debug
```

## 常见问题

### PostgreSQL 未安装

脚本会在以下位置搜索 PostgreSQL：
- `C:\Program Files\PostgreSQL\18\bin`
- `C:\Program Files\PostgreSQL\17\bin`
- `C:\Program Files\PostgreSQL\16\bin`
- `C:\Program Files\PostgreSQL\15\bin`
- `C:\Program Files\PostgreSQL\14\bin`
- `C:\Tools\pgsql\bin`

如果未找到，请从 https://www.postgresql.org/download/windows/ 下载安装。

### 端口被占用

```powershell
# 检查端口占用
Get-NetTCPConnection -LocalPort 8081

# 终止占用进程
Stop-Process -Id <PID> -Force
```

### PostgreSQL 连接失败

```powershell
# 检查 PostgreSQL 是否运行
Get-Process -Name postgres

# 查看 PostgreSQL 状态
.\scripts\start-postgres.ps1 -Action status

# 重启 PostgreSQL
.\scripts\start-postgres.ps1 -Action stop
.\scripts\start-postgres.ps1 -Action start
```

### 构建失败

```powershell
# 清理并重新下载依赖
go clean -modcache
go mod tidy
go mod download

# 重新构建
.\scripts\build.ps1
```

### 查看日志

```powershell
# 查看服务日志
Get-Content server.log -Tail 50

# 查看错误日志
Get-Content server-error.log

# 实时监控
Get-Content server.log -Wait
```

### 重置数据库

```powershell
# 方法1: 使用 deploy.ps1
.\scripts\deploy.ps1 -Force

# 方法2: 使用 start-postgres.ps1
.\scripts\start-postgres.ps1 -Action reset

# 方法3: 手动删除数据目录
Remove-Item -Recurse -Force "$env:APPDATA\dnd-mcp\postgres-data"
.\scripts\start-postgres.ps1
```

## 与 Client 配合使用

如果需要部署完整的集成环境（Client + Server）：

```powershell
# 1. 部署 Server（包含 PostgreSQL）
cd packages/server
.\scripts\deploy.ps1

# 2. 启动 Redis（Client 依赖）
cd ../client
.\scripts\start-redis.ps1

# 3. 构建并启动 Client
.\scripts\build.ps1
$env:REDIS_HOST = "localhost:6379"
$env:HTTP_PORT = "8080"
Start-Process -FilePath ".\bin\dnd-client.exe" -RedirectStandardOutput "client.log"
```

## 数据存储位置

| 数据 | 位置 |
|------|------|
| PostgreSQL 数据 | `%APPDATA%\dnd-mcp\postgres-data` |
| Server 日志 | `packages/server/server.log` |
| Server 错误日志 | `packages/server/server-error.log` |
| 构建产物 | `packages/server/bin/` |

## 手动 PSQL 连接

```powershell
# 使用 psql 连接数据库
& "C:\Program Files\PostgreSQL\17\bin\psql.exe" -h localhost -p 5432 -U postgres -d dnd_server

# 或者如果 PostgreSQL bin 在 PATH 中
psql -h localhost -p 5432 -U postgres -d dnd_server
```
