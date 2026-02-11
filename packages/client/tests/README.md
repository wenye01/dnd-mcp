# 测试文档

本项目的测试分为三个层次：

## 测试层次

### 1. 单元测试 (Unit Tests)
- **位置**: `tests/unit/`
- **依赖**: 无外部依赖
- **运行速度**: 快 (< 1秒)
- **覆盖内容**:
  - Models层测试 (`tests/unit/models/`)
  - Service层测试 (`tests/unit/service/`)

### 2. 集成测试 (Integration Tests)
- **位置**: `tests/integration/`
- **依赖**: Redis服务器
- **运行速度**: 中等 (~5秒)
- **覆盖内容**:
  - API Handler测试 (`tests/integration/api/`)
  - Redis存储测试 (`tests/integration/store/`)

### 3. 端到端测试 (E2E Tests)
- **位置**: `tests/e2e/`
- **依赖**: 运行中的HTTP服务器
- **运行速度**: 慢 (~30秒)
- **覆盖内容**:
  - 完整的用户流程
  - 真实HTTP调用
  - 并发测试
  - 错误情况测试

## 快速开始

### 运行所有测试 (推荐)

```powershell
# 使用统一测试脚本
.\scripts\test-all.ps1
```

### 运行特定类型的测试

```powershell
# 仅运行单元测试
go test -v ./tests/unit/... -cover

# 仅运行集成测试(需要Redis)
go test -v ./tests/integration/... -cover

# 仅运行E2E测试(需要运行中的服务器)
go test -v ./tests/e2e/... -timeout 60s
```

### 使用新的测试脚本

```powershell
# 从全新环境运行测试(单元+集成)
.\scripts\test-fresh.ps1

# 运行E2E测试(自动启动服务器)
.\scripts\test-e2e.ps1
```

## 测试环境要求

### 单元测试
- Go 1.24+
- 无其他要求

### 集成测试
- Go 1.24+
- Redis服务器运行在 localhost:6379
- 测试使用Redis DB 1(避免污染数据)

### E2E测试
- Go 1.24+
- Redis服务器运行在 localhost:6379
- HTTP服务器运行在 localhost:8080

## 测试工具

### testutil包

位置: `tests/testutil/testutil.go`

提供以下工具函数:

- `SetupIntegrationTest(t)` - 设置集成测试环境
- `SetupTestEnvironment(t)` - 设置测试环境(可配置)
- `CreateTestSession(data)` - 创建测试会话
- `CreateTestMessage(sessionID, content)` - 创建测试消息
- `FlushRedis()` - 清空Redis测试数据库
- `WaitForCondition(condition, timeout, msg)` - 等待条件满足
- `AssertRedisHasKey(key)` - 断言Redis包含key
- `AssertRedisHashField(key, field)` - 断言Hash包含字段

### 清理机制

测试工具提供自动清理:
- 测试结束后自动删除创建的会话和消息
- 自动关闭Redis连接
- 自动关闭PostgreSQL连接(如果使用)

## 编写测试

### 单元测试示例

```go
func TestSessionService_CreateSession(t *testing.T) {
    // 创建mock store
    mockStore := &MockSessionStore{}

    // 创建service
    service := NewSessionService(mockStore)

    // 测试
    session, err := service.CreateSession(ctx, &CreateSessionRequest{
        Name:       "测试会话",
        CreatorID:  "user-123",
        MCPServerURL: "http://localhost:9000",
    })

    // 断言
    require.NoError(t, err)
    assert.Equal(t, "测试会话", session.Name)
}
```

### 集成测试示例

```go
func TestAPI_CreateSession(t *testing.T) {
    if testing.Short() {
        t.Skip("跳过集成测试")
    }

    // 设置测试环境
    testCtx := testutil.SetupIntegrationTest(t)
    defer testCtx.Cleanup()

    // 创建测试服务器
    testServer := setupTestServer(testCtx.RedisClient)
    defer testServer.Close()

    // 测试
    resp := request(t, "POST", "/api/sessions", testData)

    // 断言
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

### E2E测试示例

```go
func TestE2E_FullUserFlow(t *testing.T) {
    // 等待服务器启动
    waitForServer(t)

    // 创建会话
    resp := request(t, "POST", "/api/sessions", sessionData)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // 获取会话
    sessionID := getSessionID(resp)
    resp = request(t, "GET", fmt.Sprintf("/api/sessions/%s", sessionID), nil)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // 更新会话
    // ...
}
```

## 测试覆盖率

查看测试覆盖率:

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 在浏览器中查看
go tool cover -html=coverage.out

# 查看覆盖率百分比
go tool cover -func=coverage.out
```

## 当前测试覆盖

- ✅ Models层: ~90%
- ✅ Service层: ~85%
- ✅ API Handlers: ~80%
- ✅ Monitor层: ~90%
- ✅ Logger层: ~95%

## 持续集成

在CI环境中运行测试:

```bash
# 运行所有测试(跳过E2E,因为CI环境可能没有运行的服务器)
go test -v ./tests/unit/... ./tests/integration/... -cover

# 或使用测试脚本
.\scripts\test-fresh.ps1 -SkipE2E
```

## 故障排除

### 测试失败: Redis连接被拒绝

```powershell
# 启动Redis
.\scripts\start-redis.ps1

# 或手动检查Redis
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING
```

### 测试失败: 端口被占用

```powershell
# 重置环境
.\scripts\reset-env.ps1 -Force
```

### 测试失败: 数据污染

```powershell
# 清空Redis
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" FLUSHALL

# 或清空特定数据库
"C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" -n 1 FLUSHDB
```

## 最佳实践

1. **编写测试前**:
   - 明确测试类型(单元/集成/E2E)
   - 选择合适的测试层次
   - 准备测试数据和mock

2. **编写测试时**:
   - 使用表驱动测试(table-driven tests)
   - 测试边界条件
   - 测试错误情况
   - 保持测试简洁

3. **运行测试**:
   - 提交前运行完整测试套件
   - 使用 `.\scripts\test-all.ps1` 确保一切正常
   - 检查测试覆盖率

4. **维护测试**:
   - 保持测试更新
   - 删除过时的测试
   - 重构重复的测试代码
   - 使用testutil工具函数

## 相关文档

- [测试脚本使用说明](../scripts/README.md)
- [代码规范](../doc/CODE_STANDARDS.md)
- [代码质量报告](../doc/CODE_QUALITY_REPORT.md)
