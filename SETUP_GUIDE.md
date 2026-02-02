# DND MCP Client - 完整部署指南

本指南提供从零开始在全新环境中部署和测试 DND MCP Client 的详细步骤。

## 目录

- [系统要求](#系统要求)
- [前置准备](#前置准备)
- [安装步骤](#安装步骤)
- [验证部署](#验证部署)
- [运行测试](#运行测试)
- [常见问题](#常见问题)
- [下一步](#下一步)

---

## 系统要求

### 必需软件

| 软件 | 最低版本 | 推荐版本 | 检查命令 |
|------|---------|---------|---------|
| Go | 1.25+ | 1.25+ | `go version` |
| PostgreSQL | 14+ | 16+ | `psql --version` |
| PowerShell | 5.1+ | 7+ | `$PSVersionTable` |

### 系统要求

- **操作系统**: Windows 10/11, Windows Server 2019+, Linux, macOS
- **内存**: 最低 2GB RAM，推荐 4GB+
- **磁盘**: 最低 500MB 可用空间
- **网络**: 需要访问 GitHub (下载 Go 依赖)

---

## 前置准备

### 步骤 1: 验证系统环境

#### 1.1 检查 Go 安装

打开 PowerShell/Terminal，执行：

```powershell
go version
```

**期望输出**:
```
go version go1.25.x windows/amd64
```

**如果未安装 Go**:
1. 访问 https://go.dev/dl/
2. 下载适合您系统的安装包
3. 运行安装程序，使用默认设置
4. 重启 PowerShell/Terminal
5. 再次运行 `go version` 验证

**验证 PATH**:
```powershell
go env GOPATH
```

期望输出一个有效路径（非空）。

#### 1.2 检查 PostgreSQL 安装

```powershell
psql --version
```

**期望输出**:
```
psql (PostgreSQL) 14.x 或更高
```

**如果未安装 PostgreSQL**:

**Windows**:
1. 访问 https://www.postgresql.org/download/windows/
2. 下载并运行安装程序
3. **重要**: 记住设置的密码（默认示例使用 `070831`）
4. 使用默认端口 `5432`
5. 确保 "pgAdmin 4" 和 "Command Line Tools" 已安装

**Linux (Ubuntu/Debian)**:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

**macOS**:
```bash
brew install postgresql@14
brew services start postgresql@14
```

#### 1.3 验证 PostgreSQL 连接

```powershell
psql -U postgres -d postgres -c "SELECT version();"
```

**期望输出**: PostgreSQL 版本信息

**如果提示密码错误**:
- 使用安装时设置的密码
- 如果忘记了密码，见 [常见问题](#常见问题)

### 步骤 2: 克隆项目

```powershell
# 克隆仓库（替换为实际仓库地址）
git clone https://github.com/your-org/dnd-mcp.git
cd dnd-mcp
```

**如果无法使用 Git**:
1. 下载项目 ZIP 文件
2. 解压到工作目录
3. 在 PowerShell 中进入项目目录

**验证项目结构**:
```powershell
ls
```

应该看到以下主要目录/文件：
```
internal/
tests/
scripts/
go.mod
go.sum
```

---

## 安装步骤

### 步骤 3: 安装 Go 依赖

```powershell
go mod download
```

**期望输出**: 无错误，可能有一些警告但不是致命错误

**验证**:
```powershell
go mod verify
```

期望输出: `all modules verified`

### 步骤 4: 配置环境变量（可选）

项目有内置默认值，但可以自定义：

**方法 1: 临时设置（当前会话）**:
```powershell
$env:PGPASSWORD = "your_password"
$env:TEST_DB_PASSWORD = "your_password"
```

**方法 2: 永久设置**:
```powershell
# 添加到系统环境变量
[System.Environment]::SetEnvironmentVariable('PGPASSWORD', 'your_password', 'User')
[System.Environment]::SetEnvironmentVariable('TEST_DB_PASSWORD', 'your_password', 'User')
```

**默认配置**:
- PostgreSQL 用户: `postgres`
- PostgreSQL 密码: `070831`
- 数据库名: `dnd_mcp_test`
- 端口: `5432`
- 主机: `localhost`

### 步骤 5: 初始化数据库

```powershell
.\scripts\init-database.ps1
```

**期望输出**:
```
========================================
Database Initialization Script
========================================

[1/5] Checking PostgreSQL connection...
[OK] PostgreSQL connection successful

[2/5] Dropping old database (if exists)...
[OK] Old database dropped

[3/5] Creating new database...
[OK] Database created successfully

[4/5] Running database migrations...
[OK] Migrations completed successfully

[5/5] Verifying database schema...
[OK] Table 'sessions' exists
[OK] Table 'messages' exists

========================================
Database Initialization Complete!
========================================

Database: dnd_mcp_test
Tables: sessions, messages

You can now run tests with:
  .\scripts\test.ps1
```

**如果失败**:
1. 检查 PostgreSQL 是否运行
2. 检查密码是否正确
3. 查看 [常见问题](#常见问题)

**手动验证数据库**:
```powershell
psql -U postgres -d dnd_mcp_test -c "\dt"
```

期望输出:
```
        List of relations
 Schema |   Name   | Type  |  Owner
--------+----------+-------+----------
 public | messages | table | postgres
 public | sessions | table | postgres
```

---

## 验证部署

### 步骤 6: 运行快速验证测试

```powershell
go test -v ./tests/unit/store -run TestPostgresStore_CreateSession
```

**期望输出**:
```
=== RUN   TestPostgresStore_CreateSession
--- PASS: TestPostgresStore_CreateSession (0.XXs)
PASS
ok      github.com/dnd-mcp/client/tests/unit/store    0.XXXs
```

### 步骤 7: 编译项目

```powershell
go build -o bin/dnd-mcp.exe ./cmd/server
```

**期望输出**: 无错误，生成了 `bin/dnd-mcp.exe`

**如果 `./cmd/server` 不存在**:
```powershell
# 编译所有包
go build ./...
```

期望输出: 无错误

---

## 运行测试

### 方法 1: 完整测试套件（推荐）

```powershell
.\scripts\test.ps1
```

这会运行：
1. 环境检查
2. 数据库设置
3. 单元测试（27 个测试）
4. 集成测试（5 个测试）
5. 竞态条件检测
6. 覆盖率报告生成

**期望输出（末尾）**:
```
========================================
Test Summary
========================================

[PASS] tests/unit/store tests
[PASS] tests/unit/client/llm tests
[PASS] tests/unit/api/handler tests
[PASS] integration tests
[PASS] No race conditions detected

========================================
All tests passed!
========================================
```

### 方法 2: 快速测试（不包含竞态检测）

```powershell
go test -v ./tests/unit/... ./tests/integration/...
```

### 方法 3: 只运行单元测试

```powershell
go test -v ./tests/unit/...
```

### 方法 4: 只运行集成测试

```powershell
go test -v ./tests/integration/...
```

### 查看测试报告

```powershell
# 测试日志
Get-Content tests\reports\*.txt

# 覆盖率报告（如果生成了）
# 在浏览器中打开
start tests\reports\coverage.html
```

---

## 常见问题

### 问题 1: "go: command not found"

**原因**: Go 未安装或 PATH 未配置

**解决方案**:
1. 重新安装 Go
2. 重启 PowerShell/Terminal
3. 验证: `go env GOPATH`

### 问题 2: "psql: command not found"

**原因**: PostgreSQL 未安装或 PATH 未配置

**解决方案**:
1. 验证 PostgreSQL 安装
2. 将 PostgreSQL bin 目录添加到 PATH
   - Windows: `C:\Program Files\PostgreSQL\16\bin`
   - Linux: `/usr/bin`
3. 重启 PowerShell/Terminal

### 问题 3: "connection refused" 或 "could not connect to server"

**原因**: PostgreSQL 未运行

**解决方案**:

**Windows**:
```powershell
# 检查服务
Get-Service -Name postgresql*

# 启动服务
Start-Service -Name postgresql-x16-16  # 根据实际版本调整
```

**Linux**:
```bash
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**macOS**:
```bash
brew services start postgresql@14
```

### 问题 4: "password authentication failed"

**原因**: 密码不正确

**解决方案**:

**重置 PostgreSQL 密码（Windows）**:
1. 打开 `pg_hba.conf`
   - 位置: `C:\Program Files\PostgreSQL\16\data\pg_hba.conf`
2. 找到 `IPv4 local connections` 行
3. 将 `md5` 改为 `trust`
4. 重启 PostgreSQL 服务
5. 连接: `psql -U postgres`
6. 重置密码: `ALTER USER postgres PASSWORD 'new_password';`
7. 恢复 `pg_hba.conf`，改为 `md5`
8. 重启服务

**重置 PostgreSQL 密码（Linux/macOS）**:
```bash
sudo -u postgres psql
ALTER USER postgres PASSWORD 'new_password';
\q
```

### 问题 5: "database already exists"

**原因**: 数据库已存在

**解决方案**:
```powershell
.\scripts\drop-database.ps1
.\scripts\init-database.ps1
```

或手动：
```powershell
psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS dnd_mcp_test;"
```

### 问题 6: 测试超时或失败

**可能原因**:
- 数据库连接问题
- 端口被占用
- 权限问题

**解决方案**:
```powershell
# 1. 检查数据库连接
psql -U postgres -d dnd_mcp_test -c "SELECT 1;"

# 2. 查看详细测试输出
go test -v ./tests/unit/store -run TestPostgresStore_CreateSession

# 3. 完全清理并重新开始
go clean -cache -testcache
.\scripts\drop-database.ps1
.\scripts\test.ps1
```

### 问题 7: "CGO_ENABLED=1" 错误（竞态检测）

**原因**: Windows 上需要启用 CGO

**解决方案**:
- 这是预期的行为，不影响测试
- 跳过竞态检测：注释掉 `test.ps1` 中的 `-race` 部分
- 或在 Linux/macOS 上运行竞态检测

### 问题 8: 权限错误（Access Denied）

**Windows**:
```powershell
# 以管理员身份运行 PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Linux/macOS**:
```bash
chmod +x scripts/*.ps1  # 如果使用 PowerShell Core
# 或直接使用
pwsh -ExecutionPolicy Bypass -File scripts/test.ps1
```

### 问题 9: 依赖下载失败

**原因**: 网络问题或 Go Proxy 问题

**解决方案**:
```powershell
# 设置 Go Proxy（中国大陆）
go env -w GOPROXY=https://goproxy.cn,direct

# 重新下载
go mod download
go mod verify
```

### 问题 10: 端口 5432 被占用

**检查端口使用**:
```powershell
netstat -ano | findstr :5432
```

**解决方案**:
1. 停止占用端口的服务
2. 或修改 PostgreSQL 配置使用其他端口
3. 更新项目连接字符串

---

## 下一步

### 验证成功标志

✅ 所有前置条件满足
✅ 项目克隆成功
✅ Go 依赖安装完成
✅ 数据库创建并迁移成功
✅ 快速验证测试通过
✅ 完整测试套件通过（32 个测试）
✅ 生成了测试报告

### 开始开发

**项目结构**:
```
dnd-mcp/
├── cmd/              # 主程序入口
├── internal/         # 内部包
│   ├── api/         # API 处理器
│   ├── client/      # LLM 客户端
│   ├── models/      # 数据模型
│   └── store/       # 数据持久化
├── tests/           # 测试代码
└── scripts/         # 脚本工具
```

**添加新功能**:
1. 在 `internal/` 中添加实现代码
2. 在 `tests/unit/` 中添加单元测试
3. 运行 `.\scripts\test.ps1` 验证
4. 提交代码

**阅读文档**:
- `DEPLOYMENT.md` - 部署指南（本文档的详细版）
- `tests/README.md` - 测试指南
- `MCP_Client开发计划.md` - 开发计划
- `MCP_Client设计.md` - 架构设计

**常用命令**:
```powershell
# 运行测试
.\scripts\test.ps1

# 重新初始化数据库
.\scripts\init-database.ps1

# 清理环境
go clean -cache -testcache

# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

### 获取帮助

**查看日志**:
```powershell
# 测试日志
Get-Content tests\reports\*.txt

# PostgreSQL 日志（Windows）
Get-Content "C:\Program Files\PostgreSQL\16\data\postgresql.log" -Tail 50
```

**启用调试模式**:
```powershell
# Go 测试详细输出
go test -v -cover ./tests/unit/...

# PostgreSQL 查询日志
# 编辑 postgresql.conf，设置:
# log_statement = 'all'
# 然后重启 PostgreSQL
```

---

## 检查清单

部署完成检查清单：

- [ ] Go 1.25+ 已安装并验证
- [ ] PostgreSQL 14+ 已安装并运行
- [ ] 项目已克隆到本地
- [ ] Go 依赖已下载（`go mod download`）
- [ ] 数据库 `dnd_mcp_test` 已创建
- [ ] 数据库迁移已运行（sessions 和 messages 表存在）
- [ ] 快速验证测试通过（至少 1 个测试）
- [ ] 完整测试套件通过（27 个单元测试 + 5 个集成测试）
- [ ] 理解项目目录结构
- [ ] 知道如何运行测试
- [ ] 阅读了相关文档

**如果所有项目都打勾，恭喜！您已成功部署 DND MCP Client！** 🎉

---

## 附录

### A. 完全卸载

如果需要完全清理环境：

```powershell
# 1. 删除数据库
.\scripts\drop-database.ps1

# 2. 清理 Go 缓存
go clean -cache -testcache -modcache

# 3. 删除项目目录
cd ..
Remove-Item -Recurse -Force dnd-mcp

# 4. （可选）卸载 PostgreSQL
# Windows: 使用"添加或删除程序"
# Linux: sudo apt remove postgresql
# macOS: brew uninstall postgresql
```

### B. 环境变量速查表

| 变量名 | 默认值 | 用途 |
|--------|--------|------|
| `PGPASSWORD` | `070831` | PostgreSQL 密码 |
| `TEST_DB_PASSWORD` | `070831` | 测试数据库密码 |
| `DATABASE_URL` | 自动生成 | 完整数据库连接字符串 |
| `GOPROXY` | `https://proxy.golang.org` | Go 模块代理 |

### C. 端口和服务

| 服务 | 默认端口 | 配置位置 |
|------|---------|---------|
| PostgreSQL | 5432 | `postgresql.conf` |
| (预留) API Server | 8080 | 代码中定义 |

### D. 有用的 SQL 命令

```sql
-- 查看所有数据库
\l

-- 连接到测试数据库
\c dnd_mcp_test

-- 查看所有表
\dt

-- 查看表结构
\d sessions
\d messages

-- 查看表数据
SELECT * FROM sessions LIMIT 10;
SELECT * FROM messages LIMIT 10;

-- 清空表数据
TRUNCATE messages, sessions CASCADE;

-- 删除表
DROP TABLE messages;
DROP TABLE sessions;

-- 删除数据库
DROP DATABASE dnd_mcp_test;
```

### E. 测试命令速查

```powershell
# 所有测试
.\scripts\test.ps1

# 单元测试
go test -v ./tests/unit/...

# 集成测试
go test -v ./tests/integration/...

# 特定包
go test -v ./tests/unit/store

# 特定测试
go test -v ./tests/unit/store -run TestPostgresStore_CreateSession

# 带覆盖率
go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out

# 详细输出
go test -v -cover ./tests/unit/... ./tests/integration/...

# 短输出（适合 CI）
go test ./tests/unit/... ./tests/integration/...
```

---

**文档版本**: 1.0
**最后更新**: 2025-02-02
**维护者**: DND MCP Team
