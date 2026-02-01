# 数据库配置说明

## 问题
在 PowerShell 中运行 `.\test.bat` 时提示输入数据库密码：
```
用户 postgres 的口令：
```

## 解决方案

### 方法 1：使用 .env 文件（推荐）

已为您创建了 `.env` 文件，包含正确的数据库密码：

```bash
# 测试数据库配置
TEST_DATABASE_URL=postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable
```

### 方法 2：修改 test.bat

`test.bat` 已更新，设置了正确的数据库密码：

```batch
set DB_PASSWORD=070831
set DATABASE_URL=postgres://postgres:%DB_PASSWORD%@localhost:5432/dnd_mcp_test?sslmode=disable
set PGPASSWORD=%DB_PASSWORD%
```

### 方法 3：使用环境变量（可选）

如果需要使用不同的密码，可以设置环境变量：

**PowerShell**:
```powershell
$env:POSTGRESQL_PASSWORD = "your-password"
```

**CMD**:
```cmd
set POSTGRESQL_PASSWORD=your-password
```

## 当前配置

### 数据库连接信息
- **主机**: localhost
- **端口**: 5432
- **用户**: postgres
- **密码**: 070831
- **测试数据库**: dnd_mcp_test

### test.bat 中的配置
```batch
set DB_PASSWORD=070831
set PGPASSWORD=%DB_PASSWORD%
```

## 如何修改密码

如果您的 PostgreSQL 密码不是 `070831`，请按以下步骤修改：

### 修改 test.bat
打开 `test.bat`，找到第 59 行：
```batch
set DB_PASSWORD=070831
```
将 `070831` 改为您的实际密码。

### 修改 .env 文件
打开 `.env` 文件，修改以下行：
```bash
POSTGRESQL_PASSWORD=your-password-here
TEST_DATABASE_URL=postgres://postgres:your-password-here@localhost:5432/dnd_mcp_test?sslmode=disable
```

## 验证配置

### 测试数据库连接
在 PowerShell 或 CMD 中运行：
```bash
psql -h localhost -U postgres -d postgres -c "SELECT 1"
```

如果配置正确，不应该提示输入密码。

### 创建测试数据库
```bash
createdb -U postgres dnd_mcp_test
```

### 运行测试
```powershell
.\test.bat
```

## 常见问题

### Q: 还是提示输入密码怎么办？
A: 确保 `test.bat` 中的 `PGPASSWORD` 环境变量设置正确。

### Q: 如何设置 PostgreSQL 密码？
A: 在 PostgreSQL 中运行：
```sql
ALTER USER postgres PASSWORD '070831';
```

### Q: 不想密码保存在文件中？
A: 可以设置 Windows 环境变量：
1. 右键"此电脑" → "属性"
2. "高级系统设置" → "环境变量"
3. 新建用户变量：
   - 变量名: `PGPASSWORD`
   - 变量值: `070831`

## 安全提示

⚠️ **重要**:
- `.env` 文件包含敏感信息，不应提交到 Git
- 确保 `.gitignore` 包含 `.env`
- 示例文件 `.env.example` 已创建，用于模板

## 下一步

现在可以运行测试：
```powershell
.\test.bat
```

应该不再提示输入密码了！
