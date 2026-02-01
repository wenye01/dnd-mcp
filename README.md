# Scripts 使用说明

本目录包含项目的构建和测试脚本。

## 📁 目录结构

```
scripts/
├── build.bat         # Windows 构建脚本
├── test.ps1          # PowerShell 测试脚本
├── migrate/          # 数据库迁移工具
└── README.md         # 本文件
```

## 🚀 快速开始

### PowerShell 环境

所有脚本都设计为在 PowerShell 环境中直接运行。

#### 1. 构建项目

```powershell
# 从项目根目录运行
.\scripts\build.bat build
```

输出：
```
Building MCP Client...
Build successful: bin\dnd-mcp-client.exe
```

#### 2. 运行测试

```powershell
# 运行完整测试套件
.\scripts\test.ps1
```

#### 3. 其他常用命令

```powershell
# 运行应用
.\scripts\build.bat run

# 清理构建文件
.\scripts\build.bat clean

# 代码检查
.\scripts\build.bat lint

# 格式化代码
.\scripts\build.bat fmt

# 管理依赖
.\scripts\build.bat deps
```

## 📋 build.bat 命令参考

### 可用命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `build` | 构建应用 | `.\scripts\build.bat build` |
| `run` | 运行应用 | `.\scripts\build.bat run` |
| `test` | 运行单元测试 | `.\scripts\build.bat test` |
| `migrate-up` | 执行数据库迁移 | `.\scripts\build.bat migrate-up` |
| `migrate-down` | 回滚数据库迁移 | `.\scripts\build.bat migrate-down` |
| `deps` | 下载依赖 | `.\scripts\build.bat deps` |
| `lint` | 代码检查 | `.\scripts\build.bat lint` |
| `fmt` | 格式化代码 | `.\scripts\build.bat fmt` |
| `clean` | 清理构建文件 | `.\scripts\build.bat clean` |
| `help` | 显示帮助信息 | `.\scripts\build.bat help` |

### 使用示例

```powershell
# 构建并运行
.\scripts\build.bat build
.\scripts\build.bat run

# 运行测试
.\scripts\build.bat test

# 清理并重新构建
.\scripts\build.bat clean
.\scripts\build.bat build
```

## 🧪 test.ps1 测试脚本

### 功能

test.ps1 是一个完整的测试脚本，提供以下功能：

1. ✅ 环境检查（Go、数据库）
2. ✅ 清理旧测试数据
3. ✅ 运行单元测试
4. ✅ 运行集成测试
5. ✅ 生成覆盖率报告

### 使用方法

```powershell
# 运行所有测试
.\scripts\test.ps1

# 查看测试输出
Get-Content tests\reports\test_output.txt

# 查看覆盖率报告
# 报告生成在: tests\reports\coverage.html
```

### 测试报告位置

测试报告会保存在 `tests/reports/` 目录：

```
tests/reports/
├── test_output.txt       # 完整测试输出
├── test_report.txt       # 测试报告摘要
├── coverage.out          # 覆盖率数据
└── coverage.html         # HTML 覆盖率报告
```

### 环境要求

- ✅ Go 1.24+
- ✅ PostgreSQL 数据库（可选，用于集成测试）
- ✅ PowerShell 环境

### 数据库配置

默认测试数据库配置：
- 主机: `localhost`
- 端口: `5432`
- 用户: `postgres`
- 密码: `070831`
- 数据库: `dnd_mcp_test`

如需修改，请编辑 `test.ps1` 中的环境变量设置。

## 🔧 数据库迁移

### 执行迁移

```powershell
# 向上迁移（创建表结构）
.\scripts\build.bat migrate-up

# 向下迁移（删除表结构）
.\scripts\build.bat migrate-down
```

### 迁移文件位置

迁移文件位于项目根目录的 `migrations/` 文件夹：

```
migrations/
├── 001_initial_schema.up.sql
└── 001_initial_schema.down.sql
```

## ⚠️ 注意事项

### 执行策略

如果在运行 PowerShell 脚本时遇到执行策略错误：

```powershell
# 临时允许脚本执行
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# 然后运行脚本
.\scripts\test.ps1
```

### 工作目录

所有脚本都会自动切换到项目根目录执行，因此可以从任何位置运行：

```powershell
# 从项目根目录运行
.\scripts\build.bat build
.\scripts\test.ps1

# 从 scripts 目录运行也可以
cd scripts
.\build.bat build
.\test.ps1
```

## 📝 常见问题

### Q1: 提示 "无法加载文件，因为在此系统上禁止运行脚本"

**解决方案：**
```powershell
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

### Q2: 构建失败，提示找不到 Go

**解决方案：**
确保 Go 已安装并在 PATH 中：
```powershell
go version
```

### Q3: 数据库连接失败

**解决方案：**
1. 确保 PostgreSQL 正在运行
2. 检查密码配置（test.ps1 第 12-13 行）
3. 确认数据库服务可访问

## 🎯 最佳实践

1. **构建前先清理**
   ```powershell
   .\scripts\build.bat clean
   .\scripts\build.bat build
   ```

2. **提交代码前运行测试**
   ```powershell
   .\scripts\test.ps1
   ```

3. **定期运行代码检查**
   ```powershell
   .\scripts\build.bat lint
   .\scripts\build.bat fmt
   ```

4. **查看覆盖率报告**
   ```powershell
   # 运行测试后打开 HTML 报告
   start tests\reports\coverage.html
   ```

## 📚 相关文档

- [项目文档](../doc/)
- [开发计划](../doc/MCP_Client开发计划.md)
- [测试报告](../tests/reports/)

---

**最后更新**: 2026-02-01
