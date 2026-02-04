# Scripts - 构建和测试脚本 (Windows环境)

本目录包含用于Windows环境的PowerShell开发和测试脚本。

## 快速开始

```powershell
# 运行完整测试套件 (推荐)
.\scripts\test-all.ps1

# 设置开发环境
.\scripts\dev.ps1

# 构建项目
.\scripts\build.ps1
```

## 脚本功能说明

### 🧪 test-all.ps1 - 统一测试脚本 (推荐使用)

**这是最重要的脚本**,运行完整的测试套件。

**功能**:
1. ✅ 环境清理
   - 停止所有运行的服务 (Redis、DND服务器)
   - 清空Redis数据库

2. 🔨 构建项目
   - 清理旧的构建文件
   - 编译新版本

3. 📋 运行单元测试
   - Service层测试 (tests/unit/service/) - 约9个测试
   - Model层测试 (tests/unit/models/)
   - 其他单元测试

4. 🔗 运行集成测试
   - API集成测试 (tests/integration/api/) - 约8个测试
   - Store集成测试 (tests/integration/store/)

5. 🌐 运行功能测试
   - 健康检查端点
   - 完整的CRUD操作:
     - 创建会话 (POST /api/sessions)
     - 获取会话 (GET /api/sessions/:id)
     - 列出会话 (GET /api/sessions)
     - 更新会话 (PATCH /api/sessions/:id)
     - 删除会话 (DELETE /api/sessions/:id)

6. 📊 结果汇总
   - 彩色输出显示测试结果
   - 详细的通过/失败信息

**输出示例**:
```
========================================
  DND MCP Complete Test Suite
========================================

=== Step 1/6: Stopping Services ===
✓ Redis stopped
✓ DND client server stopped

=== Step 4/6: Building Project ===
✓ Build successful

=== Step 5/6: Running Unit & Integration Tests ===
Running Service Layer Unit Tests...
✓ Service layer unit tests passed

Running API Integration Tests...
✓ API integration tests passed

========================================
  Test Results Summary
========================================

Unit Tests:           PASSED
Integration Tests:    PASSED
API Functional Tests: PASSED

Overall Result: ALL TESTS PASSED ✓
```

**使用方法**:
```powershell
# 运行所有测试
.\scripts\test-all.ps1

# 如果遇到执行策略问题
powershell -ExecutionPolicy Bypass -File .\scripts\test-all.ps1
```

---

### 🔨 build.ps1 - 构建脚本

**功能**: 从源代码编译项目

**构建步骤**:
1. 清理bin目录
2. 编译主程序: `cmd/client/main.go`
3. 输出到: `bin/dnd-client.exe`

**使用方法**:
```powershell
.\scripts\build.ps1
```

**适用场景**:
- 修改代码后需要重新编译
- 确保代码能正常编译
- 准备部署前构建

---

### 🛠️ dev.ps1 - 开发环境设置

**功能**: 一键设置完整的开发环境

**设置内容**:
1. 检查Go环境
2. 安装依赖 (`go mod tidy`)
3. 编译项目
4. 启动Redis (如果需要)
5. 验证环境配置

**使用方法**:
```powershell
.\scripts\dev.ps1
```

**适用场景**:
- 首次克隆项目后
- 更换开发机器时
- 重新设置开发环境

---

### 🔄 reset-env.ps1 - 环境重置脚本

**功能**: 将开发环境重置到初始状态

**重置内容**:
1. ⚠️ 停止所有Redis服务/进程
2. ⚠️ 清空所有Redis数据库 (FLUSHALL)
3. 🗑️ 删除构建产物 (bin/目录)
4. 🧹 清理测试缓存 (`go clean -testcache`)
5. 🗑️ 删除临时文件 (*.log, *.tmp, *.temp)
6. 🧹 清理Redis日志文件

**警告**:
- ⚠️ 此操作会删除所有Redis数据 (不可恢复!)
- ⚠️ 所有Redis进程将被停止
- ⚠️ 构建文件将被删除

**使用方法**:
```powershell
# 交互模式 (会询问确认)
.\scripts\reset-env.ps1

# 强制模式 (不询问确认)
.\scripts\reset-env.ps1 -Force

# 详细模式 (显示所有操作详情)
.\scripts\reset-env.ps1 -Verbose

# 组合使用
.\scripts\reset-env.ps1 -Force -Verbose
```

**适用场景**:
- 开始新的开发任务前
- 环境变得混乱时
- 运行测试前确保干净状态
- 部署前清理环境

---

### 🔴 start-redis.ps1 - Redis启动脚本

**功能**: 启动Redis服务器 (如果尚未运行)

**检查步骤**:
1. 检查Redis是否已在运行
2. 如果未运行,启动Redis服务
3. 验证Redis连接

**使用方法**:
```powershell
.\scripts\start-redis.ps1
```

**适用场景**:
- 单独启动Redis时
- 其他脚本未启动Redis时
- 手动测试Redis连接

---

### ⚡ test.ps1 - 快速测试脚本

**功能**: 运行基本的Go测试

**测试范围**:
- 单元测试
- 基本集成测试

**使用方法**:
```powershell
.\scripts\test.ps1
```

**适用场景**:
- 快速验证代码
- 不需要完整测试时
- 开发过程中的快速反馈

---

## 功能对比

| 脚本 | 清理环境 | 构建项目 | 单元测试 | 集成测试 | 功能测试 | 适用场景 |
|------|---------|---------|---------|---------|---------|---------|
| **test-all.ps1** | ✅ | ✅ | ✅ | ✅ | ✅ | 完整测试 (推荐) |
| build.ps1 | ❌ | ✅ | ❌ | ❌ | ❌ | 仅构建 |
| dev.ps1 | ❌ | ✅ | ❌ | ❌ | ❌ | 环境设置 |
| reset-env.ps1 | ✅ | ❌ | ❌ | ❌ | ❌ | 环境清理 |
| start-redis.ps1 | ❌ | ❌ | ❌ | ❌ | ❌ | 启动Redis |
| test.ps1 | ❌ | ❌ | ✅ | ✅ | ❌ | 快速测试 |

## 特性

### ✅ 无编码问题
- 使用英文输出
- UTF-8编码声明
- 适用于所有Windows系统

### ✅ 彩色输出
- 🟢 绿色 - 成功操作
- 🔴 红色 - 错误信息
- 🟡 黄色 - 警告/标题
- 🔵 蓝色 - 信息提示

### ✅ 错误处理
- 优雅的失败处理
- 清晰的错误信息
- 自动清理资源

### ✅ 自动化
- 自动停止冲突服务
- 自动清理测试数据
- 自动启动依赖服务

## 常见问题

### PowerShell执行策略问题

**错误**: "无法加载文件,因为在此系统上禁止运行脚本"

**解决方案**:
```powershell
# 临时解决 (推荐)
powershell -ExecutionPolicy Bypass -File .\scripts\test-all.ps1

# 永久解决
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 端口被占用

**错误**: "bind: Only one usage of each socket address"

**解决方案**:
```powershell
# 方法1: 使用重置脚本
.\scripts\reset-env.ps1 -Force

# 方法2: 手动查找并停止进程
Get-NetTCPConnection -LocalPort 8080 | Select-Object -ExpandProperty OwningProcess
Stop-Process -Id <PID> -Force
```

### Redis连接失败

**错误**: "Redis connection failed"

**解决方案**:
```powershell
# 检查Redis是否运行
& "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING

# 启动Redis
.\scripts\start-redis.ps1
```

### 找不到go命令

**错误**: 'go' 不是内部或外部命令

**解决方案**:
1. 从 https://golang.org/dl/ 安装Go
2. 将Go添加到系统PATH
3. 重启PowerShell

## 测试覆盖

### 单元测试 (~9个测试)
- ✅ Service层业务逻辑
- ✅ 参数验证
- ✅ 错误处理
- ✅ 边界条件测试

### 集成测试 (~8个测试)
- ✅ HTTP请求处理
- ✅ 响应格式化
- ✅ 错误响应
- ✅ 输入验证

### 功能测试 (~6个测试)
- ✅ 健康检查端点
- ✅ 创建会话
- ✅ 获取会话
- ✅ 列出会话
- ✅ 更新会话
- ✅ 删除会话

**总计**: 约23个测试用例

## 使用示例

### 示例1: 完整开发工作流

```powershell
# 1. 重置环境 (干净开始)
.\scripts\reset-env.ps1 -Force

# 2. 运行所有测试
.\scripts\test-all.ps1

# 3. 如果测试通过,启动服务器
.\bin\dnd-client.exe server
```

### 示例2: 快速开发循环

```powershell
# 1. 修改代码

# 2. 重新构建
.\scripts\build.ps1

# 3. 快速测试
.\scripts\test.ps1

# 4. 启动服务器
.\bin\dnd-client.exe server
```

### 示例3: 提交代码前

```powershell
# 运行完整测试确保一切正常
.\scripts\test-all.ps1

# 如果全部通过,提交代码
git add .
git commit -m "your message"
```

## 目录结构

```
scripts/
├── test-all.ps1       # 🧪 统一测试脚本 (推荐使用)
├── build.ps1          # 🔨 构建项目
├── dev.ps1            # 🛠️ 开发环境设置
├── reset-env.ps1      # 🔄 重置环境
├── start-redis.ps1    # 🔴 启动Redis
├── test.ps1           # ⚡ 快速测试
├── README.md          # 📖 本文档
└── migrations/        # 📁 数据库迁移文件
```

## 最佳实践

### 1. 开发前
```powershell
# 清理环境,确保干净状态
.\scripts\reset-env.ps1 -Force
```

### 2. 开发中
```powershell
# 使用快速测试
.\scripts\test.ps1
```

### 3. 提交前
```powershell
# 运行完整测试
.\scripts\test-all.ps1
```

### 4. 遇到问题时
```powershell
# 先重置环境
.\scripts\reset-env.ps1 -Force

# 重新设置环境
.\scripts\dev.ps1

# 运行测试
.\scripts\test-all.ps1
```

## 支持与文档

如有问题或疑问,请参考:
- 📄 项目文档: `doc/`
- 📊 测试报告: `doc/测试报告-任务三.md`
- 📋 开发指南: `doc/开发任务三_修订版.md`
- 🔧 脚本整理说明: `doc/Scripts整理说明.md`

## 版本历史

- **v1.0** (2026-02-04)
  - 初始版本
  - 统一测试脚本
  - 完整的文档

---

**提示**: 推荐使用 `test-all.ps1` 作为主要测试脚本,它提供了最完整的测试覆盖和最好的用户体验。
