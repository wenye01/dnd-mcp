# Server 脚本说明

本目录包含 DND MCP Server 的部署和管理脚本。

## 脚本概览

| 脚本 | 用途 | 使用场景 |
|------|------|---------|
| `deploy.ps1` | 一键部署 | 首次部署或完整重新部署 |
| `stop.ps1` | 停止服务 | 停止运行中的服务 |
| `build.ps1` | 构建服务 | 仅编译二进制文件 |
| `start-postgres.ps1` | 管理 PostgreSQL | 启动/停止数据库容器 |
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

# 停止所有服务（包括数据库）
.\scripts\stop.ps1 -All

# 仅停止数据库
.\scripts\stop.ps1 -Db
```

## 详细说明

### deploy.ps1 - 一键部署

自动化完成以下步骤：
1. 启动 PostgreSQL 容器
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
- `-Db`: 仅停止数据库容器
- `-Clean`: 清理日志文件

### start-postgres.ps1 - PostgreSQL 管理

**参数：**
- `-Action start`: 启动容器（默认）
- `-Action stop`: 停止容器
- `-Action remove`: 删除容器
- `-Action status`: 查看状态
- `-Action reset`: 重置容器
- `-Reset`: 启动时重置容器

**示例：**
```powershell
# 启动 PostgreSQL
.\scripts\start-postgres.ps1

# 停止 PostgreSQL
.\scripts\start-postgres.ps1 -Action stop

# 查看状态
.\scripts\start-postgres.ps1 -Action status

# 完全重置
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

## 环境变量

Server 启动时使用的环境变量：

```powershell
# PostgreSQL 配置
$env:POSTGRES_HOST = "localhost"
$env:POSTGRES_PORT = "5432"
$env:POSTGRES_USER = "dnd"
$env:POSTGRES_PASSWORD = "password"
$env:POSTGRES_DBNAME = "dnd_server"
$env:POSTGRES_SSLMODE = "disable"

# HTTP 配置
$env:HTTP_HOST = "0.0.0.0"
$env:HTTP_PORT = "8081"

# 日志配置
$env:LOG_LEVEL = "info"
$env:LOG_FORMAT = "text"
```

## 常见问题

### 端口被占用

```powershell
# 检查端口占用
Get-NetTCPConnection -LocalPort 8081

# 终止占用进程
Stop-Process -Id <PID> -Force
```

### PostgreSQL 连接失败

```powershell
# 检查容器状态
docker ps -a --filter "name=dnd-postgres"

# 查看容器日志
docker logs dnd-postgres

# 重启容器
.\scripts\start-postgres.ps1 -Action reset
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

## Docker 容器信息

| 容器名 | 镜像 | 端口 | 用途 |
|--------|------|------|------|
| dnd-postgres | postgres:16-alpine | 5432 | Server 数据库 |
| dnd-redis | redis:7-alpine | 6379 | Client 缓存 |

```powershell
# 查看所有容器
docker ps -a

# 查看容器日志
docker logs dnd-postgres
docker logs dnd-redis

# 进入容器
docker exec -it dnd-postgres psql -U postgres
docker exec -it dnd-redis redis-cli
```
