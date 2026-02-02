# 测试指南

本文档说明如何在新环境中设置并运行测试。

## 前置要求

1. **Go 1.25+** - 编程语言
2. **PostgreSQL 14+** - 数据库
3. **PowerShell** - Windows环境（或使用bash脚本）

## 环境设置

### 1. 安装PostgreSQL

确保PostgreSQL已安装并运行在localhost:5432。

```bash
# Windows使用默认安装
# 默认用户: postgres
# 默认密码: 070831 (可在脚本中修改)
```

### 2. 创建测试数据库

```bash
# 使用psql创建测试数据库
psql -U postgres -c "CREATE DATABASE dnd_mcp_test;"
```

### 3. 安装Go依赖

```bash
go mod download
```

## 运行测试

### Windows PowerShell

```powershell
# 运行完整测试套件
.\scripts\test.ps1

# 或者分步运行
go test ./internal/store/...
go test ./internal/client/llm/...
go test ./internal/api/handler/...
go test ./tests/integration/...
```

### Linux/Mac (Bash)

```bash
# 设置环境变量
export TEST_DB_PASSWORD="070831"
export PGPASSWORD="070831"

# 运行测试
go test ./internal/...
go test ./tests/integration/...
```

## 数据库迁移

测试脚本会自动运行数据库迁移，但也可以手动运行：

```bash
# 创建schema
go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up

# 删除schema
go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action down
```

## 测试结构

```
tests/
├── integration/           # 集成测试
│   └── chat_integration_test.go
├── reports/              # 测试报告（自动生成）
│   ├── coverage.html     # 覆盖率报告（HTML）
│   ├── coverage.out      # 覆盖率数据
│   ├── store_tests.txt   # Store测试输出
│   ├── llm_tests.txt     # LLM测试输出
│   ├── handler_tests.txt # Handler测试输出
│   ├── integration_tests.txt  # 集成测试输出
│   └── race_tests.txt    # 竞态检测输出
└── README.md             # 本文档
```

## 环境变量

- `TEST_DB_PASSWORD` - 数据库密码（默认: 070831）
- `DATABASE_URL` - 完整数据库连接字符串（可选）
- `PGPASSWORD` - PostgreSQL密码（用于psql命令）

## 在全新环境中一键部署

```powershell
# Windows PowerShell
# 1. 克隆仓库
git clone <repository-url>
cd dnd-mcp

# 2. 创建测试数据库
psql -U postgres -c "CREATE DATABASE dnd_mcp_test;"

# 3. 运行测试脚本（会自动完成所有设置）
.\scripts\test.ps1
```

## 测试输出说明

测试运行后，会在`tests/reports/`目录生成以下文件：

- `coverage.html` - 在浏览器中打开查看覆盖率报告
- `*_tests.txt` - 各模块的详细测试输出
- `race_tests.txt` - 竞态条件检测结果

## 故障排查

### 数据库连接失败

```bash
# 检查PostgreSQL是否运行
psql -U postgres -c "SELECT 1;"

# 检查数据库是否存在
psql -U postgres -l | grep dnd_mcp_test
```

### Schema错误

```bash
# 重新创建schema
go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action down
go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up
```

### 端口冲突

如果PostgreSQL不在默认端口5432，修改连接字符串中的端口号。

## 持续集成

测试脚本适用于CI/CD环境，只需确保：

1. PostgreSQL服务可用
2. 设置正确的环境变量
3. 运行`.\scripts\test.ps1`

## 许可

与主项目相同。
