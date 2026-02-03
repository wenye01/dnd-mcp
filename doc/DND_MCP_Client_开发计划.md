# DND MCP Client 开发计划

## 文档信息

- **版本**: v1.0
- **创建日期**: 2025-02-03
- **基于**: DND_MCP_Client详细设计.md v1.1
- **开发原则**: 小步迭代、可运行、可测试

---

## 开发原则

### 核心理念

1. **最小可运行单位**：每个开发任务结束都必须是一个完整的、可运行的程序
2. **需求驱动**：以需求为最小单位，多个需求组成一个开发任务
3. **独立测试**：每个任务都可以独立测试，无需依赖外部服务（使用 Mock）
4. **测试先行**：每个需求都有明确的测试用例
5. **增量交付**：每个任务都能演示具体功能

### 测试策略

- **单元测试**：测试单个函数和类
- **集成测试**：测试多个组件协作（使用 Mock 替代外部服务）
- **HTTP 测试**：使用 httptest 测试 HTTP API
- **Mock 策略**：
  - LLM Client：使用 Mock 返回预设响应
  - MCP Server：使用 Mock Server 模拟 MCP 协议
  - 真实依赖：Redis、PostgreSQL（使用 testcontainers）
- 测试应该能从完全初始的状态开始

---

## 任务拆解

### 任务1：项目脚手架 + Redis 基础存储

**目标**：搭建项目基础架构，实现 Redis 存储会话和消息

**需求清单**：

1. 需求1.1：项目脚手架搭建
2. 需求1.2：Redis 连接和配置管理
3. 需求1.3：Redis 存储会话
4. 需求1.4：Redis 存储消息
5. 需求1.5：命令行工具测试

**可演示功能**：

```bash
# 创建会话
./bin/dnd-client session create --name "测试会话" --creator "user-123"

# 查看会话
./bin/dnd-client session get <session-id>

# 保存消息
./bin/dnd-client message save --session <session-id> --content "你好"

# 查看消息
./bin/dnd-client message list --session <session-id>
```

**测试用例**：

| 需求  | 测试用例                    | 测试类型  |
| --- | ----------------------- | ----- |
| 1.1 | 项目可以正常编译和运行             | 单元测试  |
| 1.2 | Redis 连接成功，Ping 返回 PONG | 集成测试  |
| 1.3 | 创建会话后能从 Redis 读取        | 集成测试  |
| 1.4 | 保存消息后能从 Redis 读取        | 集成测试  |
| 1.5 | CLI 命令执行正确              | 端到端测试 |

**验收标准**：

- ✅ 项目可以编译运行
- ✅ 使用 Redis 存储会话和消息
- ✅ 提供 CLI 工具测试所有功能
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（使用 Docker 启动）
- 无其他依赖

---

### 任务2：PostgreSQL 持久化

**目标**：实现 PostgreSQL 备份存储和数据恢复

**需求清单**：

1. 需求2.1：PostgreSQL 连接和数据库迁移
2. 需求2.2：从 Redis 备份会话到 PostgreSQL
3. 需求2.3：从 Redis 备份消息到 PostgreSQL
4. 需求2.4：从 PostgreSQL 恢复会话到 Redis
5. 需求2.5：从 PostgreSQL 恢复消息到 Redis
6. 需求2.6：命令行工具测试持久化

**可演示功能**：

```bash
# 备份到 PostgreSQL
./bin/dnd-client backup --all

# 从 PostgreSQL 恢复
./bin/dnd-client restore --all

# 查看备份记录
./bin/dnd-client backup list
```

**测试用例**：

| 需求  | 测试用例                  | 测试类型  |
| --- | --------------------- | ----- |
| 2.1 | PostgreSQL 连接成功，表结构正确 | 集成测试  |
| 2.2 | 会话正确备份到 PostgreSQL    | 集成测试  |
| 2.3 | 消息正确备份到 PostgreSQL    | 集成测试  |
| 2.4 | 会话从 PostgreSQL 正确恢复   | 集成测试  |
| 2.5 | 消息从 PostgreSQL 正确恢复   | 集成测试  |
| 2.6 | CLI 命令执行正确            | 端到端测试 |

**验收标准**：

- ✅ 数据库表结构正确创建
- ✅ 可以备份和恢复会话、消息
- ✅ 提供命令行工具测试
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis、PostgreSQL（使用 Docker 启动）
- 前置任务：任务1

---

### 任务3：HTTP API - 会话管理

**目标**：实现会话管理的 REST API

**需求清单**：

1. 需求3.1：HTTP 服务器和路由设置
2. 需求3.2：创建会话 API (POST /api/sessions)
3. 需求3.3：获取会话详情 API (GET /api/sessions/{id})
4. 需求3.4：列出所有会话 API (GET /api/sessions)
5. 需求3.5：更新会话 API (PATCH /api/sessions/{id})
6. 需求3.6：删除会话 API (DELETE /api/sessions/{id})

**可演示功能**：

```bash
# 启动服务器
./bin/dnd-client server start

# 测试 API
curl http://localhost:8080/api/sessions
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"name":"测试会话","creator_id":"user-123","mcp_server_url":"http://localhost:9000"}'
```

**测试用例**：

| 需求  | 测试用例                     | 测试类型    |
| --- | ------------------------ | ------- |
| 3.1 | HTTP 服务器正常启动和监听          | 集成测试    |
| 3.2 | 创建会话返回 201，数据正确存入 Redis  | HTTP 测试 |
| 3.3 | 获取存在的会话返回 200，不存在的返回 404 | HTTP 测试 |
| 3.4 | 列出会话返回数组，状态过滤正确          | HTTP 测试 |
| 3.5 | 更新会话返回 200，数据正确更新        | HTTP 测试 |
| 3.6 | 删除会话返回 204，数据标记为删除       | HTTP 测试 |

**验收标准**：

- ✅ 所有 API 正常工作
- ✅ HTTP 状态码正确
- ✅ 请求参数验证正确
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（真实）、PostgreSQL（可选，用于持久化测试）
- 前置任务：任务1

---

### 任务4：HTTP API - 消息管理

**目标**：实现消息管理的 REST API

**需求清单**：

1. 需求4.1：发送消息 API (POST /api/sessions/{id}/chat)
2. 需求4.2：获取消息历史 API (GET /api/sessions/{id}/messages)
3. 需求4.3：获取单条消息 API (GET /api/sessions/{id}/messages/{msg_id})

**可演示功能**：

```bash
# 发送消息（使用 Mock LLM）
curl -X POST http://localhost:8080/api/sessions/{id}/chat \
  -H "Content-Type: application/json" \
  -d '{"content":"你好","player_id":"player-123"}'

# 返回 Mock AI 响应
{
  "id": "msg-xxx",
  "role": "assistant",
  "content": "这是 Mock LLM 的响应",
  "created_at": "2025-02-03T10:00:00Z"
}

# 获取消息历史
curl http://localhost:8080/api/sessions/{id}/messages?limit=10
```

**测试用例**：

| 需求  | 测试用例                         | 测试类型    |
| --- | ---------------------------- | ------- |
| 4.1 | 发送消息后保存到 Redis，返回 Mock AI 响应 | HTTP 测试 |
| 4.2 | 获取消息历史返回数组，limit 参数正确        | HTTP 测试 |
| 4.3 | 获取存在的消息返回 200，不存在的返回 404     | HTTP 测试 |

**Mock 策略**：

- LLM Client：使用 Mock 返回固定响应 `"这是 Mock LLM 的响应"`
- 不调用真实的 LLM API

**验收标准**：

- ✅ 所有 API 正常工作
- ✅ Mock LLM 返回预设响应
- ✅ 消息正确保存到 Redis
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（真实）
- Mock：LLM Client（Mock）
- 前置任务：任务1、任务3

---

### 任务5：WebSocket 实时通信

**目标**：实现 WebSocket 实时推送事件

**需求清单**：

1. 需求5.1：WebSocket 服务器和连接管理
2. 需求5.2：订阅和取消订阅事件
3. 需求5.3：广播新消息事件
4. 需求5.4：广播游戏状态变更事件（Mock）
5. 需求5.5：心跳机制

**可演示功能**：

```bash
# 使用 websocat 测试 WebSocket
websocat ws://localhost:8080/ws/sessions/{id}?key={ws-key}

# 发送订阅消息
{"type":"subscribe","data":{"events":["new_message","state_changed"]}}

# 接收服务器推送
{"type":"new_message","data":{"id":"msg-xxx","content":"新消息"}}
```

**测试用例**：

| 需求  | 测试用例                  | 测试类型 |
| --- | --------------------- | ---- |
| 5.1 | WebSocket 连接成功，验证 key | 集成测试 |
| 5.2 | 订阅后能接收指定事件，取消订阅后不再接收  | 集成测试 |
| 5.3 | 发送消息时广播到所有订阅的连接       | 集成测试 |
| 5.4 | 模拟状态变更并广播             | 集成测试 |
| 5.5 | 心跳 ping/pong 正常工作     | 集成测试 |

**Mock 策略**：

- MCP Server 事件：使用 Mock 生成事件

**验收标准**：

- ✅ WebSocket 连接正常
- ✅ 事件订阅和推送正确
- ✅ 心跳机制正常
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（真实）
- Mock：事件生成器（Mock）
- 前置任务：任务3、任务4

---

### 任务6：LLM 集成（真实）

**目标**：集成真实的 LLM 服务（可选配置使用 Mock）

**需求清单**：

1. 需求6.1：LLM Client 配置和初始化
2. 需求6.2：调用 OpenAI API
3. 需求6.3：解析 LLM 响应和 tool_calls
4. 需求6.4：流式响应支持（可选）
5. 需求6.5：Mock 模式（配置切换）

**可演示功能**：

```bash
# 使用真实 LLM
LLM_PROVIDER=openai LLM_API_KEY=sk-xxx ./bin/dnd-client server start

# 使用 Mock LLM
LLM_PROVIDER=mock ./bin/dnd-client server start

# 发送消息
curl -X POST http://localhost:8080/api/sessions/{id}/chat \
  -d '{"content":"告诉我一个故事"}'

# 返回真实 LLM 响应
```

**测试用例**：

| 需求  | 测试用例                  | 测试类型     |
| --- | --------------------- | -------- |
| 6.1 | LLM Client 配置加载正确     | 单元测试     |
| 6.2 | 真实 API 调用成功（使用测试 key） | 集成测试     |
| 6.3 | tool_calls 正确解析       | 单元测试     |
| 6.4 | 流式响应正确处理              | 集成测试（可选） |
| 6.5 | Mock 模式正确切换           | 单元测试     |

**Mock 策略**：

- 提供 Mock LLM 实现，返回预设响应
- 通过环境变量 `LLM_PROVIDER=mock` 切换

**验收标准**：

- ✅ 支持 OpenAI API
- ✅ 支持流式响应（可选）
- ✅ 可以切换 Mock 模式
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（真实）、OpenAI API（可选）
- Mock：Mock LLM（内置）
- 前置任务：任务4

---

### 任务7：MCP Server 集成

**目标**：集成 MCP Server，实现工具调用和事件订阅

**需求清单**：

1. 需求7.1：MCP Client 初始化和握手
2. 需求7.2：调用 MCP Server 工具
3. 需求7.3：订阅 MCP Server 事件
4. 需求7.4：事件转换为 WebSocket 推送
5. 需求7.5：Mock MCP Server（用于测试）

**可演示功能**：

```bash
# 使用真实 MCP Server
./bin/dnd-client server start --mcp-url http://localhost:9000

# 使用 Mock MCP Server
./bin/dnd-client server start --mcp-url mock://

# 调用工具
curl -X POST http://localhost:8080/api/sessions/{id}/chat \
  -d '{"content":"投掷骰子"}'

# LLM 决定调用 roll_dice 工具
# MCP Server 返回结果
# WebSocket 推送事件
```

**测试用例**：

| 需求  | 测试用例                          | 测试类型 |
| --- | ----------------------------- | ---- |
| 7.1 | MCP 握手成功，session 初始化          | 集成测试 |
| 7.2 | 工具调用正确发送到 MCP Server          | 集成测试 |
| 7.3 | 事件订阅成功                        | 集成测试 |
| 7.4 | MCP Server 事件转换为 WebSocket 推送 | 集成测试 |
| 7.5 | Mock MCP Server 正确响应          | 集成测试 |

**Mock 策略**：

- 实现 Mock MCP Server（http handler）
- 模拟工具调用响应
- 模拟事件推送

**验收标准**：

- ✅ MCP 协议正确实现
- ✅ 工具调用正常工作
- ✅ 事件订阅和推送正常
- ✅ 提供 Mock MCP Server 用于测试
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis（真实）、MCP Server（可选，Mock 用于测试）
- Mock：Mock MCP Server（内置）
- 前置任务：任务5、任务6

---

### 任务8：持久化触发器

**目标**：实现可扩展的持久化触发器系统

**需求清单**：

1. 需求8.1：触发器接口定义
2. 需求8.2：TimeTrigger 实现（时间触发）
3. 需求8.3：MessageCountTrigger 实现（消息量触发）
4. 需求8.4：ManualTrigger 实现（手动触发）
5. 需求8.5：CompositeTrigger 实现（组合触发）
6. 需求8.6：触发器管理器
7. 需求8.7：配置化管理

**可演示功能**：

```bash
# 配置文件 config.yaml
persistence:
  type: composite
  triggers:
    - type: time
      interval: 30s
    - type: message
      threshold: 100

# 启动服务
./bin/dnd-client server start --config config.yaml

# 手动触发持久化
curl -X POST http://localhost:8080/api/system/persistence/trigger

# 查看持久化状态
curl http://localhost:8080/api/system/persistence/status
```

**测试用例**：

| 需求  | 测试用例                       | 测试类型 |
| --- | -------------------------- | ---- |
| 8.1 | 触发器接口定义正确                  | 单元测试 |
| 8.2 | TimeTrigger 每隔指定时间触发       | 单元测试 |
| 8.3 | MessageCountTrigger 达到阈值触发 | 单元测试 |
| 8.4 | ManualTrigger 收到信号触发       | 单元测试 |
| 8.5 | CompositeTrigger 任一满足即触发   | 单元测试 |
| 8.6 | 触发器管理器正确注册和调用              | 集成测试 |
| 8.7 | 配置文件正确加载                   | 集成测试 |

**验收标准**：

- ✅ 所有触发器正确实现
- ✅ 触发器可以组合使用
- ✅ 配置文件正确管理
- ✅ 测试覆盖率 > 80%

**依赖**：

- 外部：Redis、PostgreSQL（真实）
- 前置任务：任务2

---

### 任务9：系统监控和日志

**目标**：添加系统监控、健康检查和日志

**需求清单**：

1. 需求9.1：结构化日志
2. 需求9.2：健康检查 API (GET /api/system/health)
3. 需求9.3：系统统计 API (GET /api/system/stats)
4. 需求9.4：Prometheus 指标（可选）

**可演示功能**：

```bash
# 健康检查
curl http://localhost:8080/api/system/health

# 返回
{
  "status": "healthy",
  "components": {
    "redis": {"status": "healthy"},
    "postgres": {"status": "healthy"}
  }
}

# 系统统计
curl http://localhost:8080/api/system/stats

# Prometheus 指标
curl http://localhost:8080/metrics
```

**测试用例**：

| 需求  | 测试用例              | 测试类型     |
| --- | ----------------- | -------- |
| 9.1 | 日志格式正确，包含必要字段     | 单元测试     |
| 9.2 | 健康检查正确反映系统状态      | HTTP 测试  |
| 9.3 | 统计信息准确            | HTTP 测试  |
| 9.4 | Prometheus 指标正确导出 | 集成测试（可选） |

**验收标准**：

- ✅ 日志结构化且可检索
- ✅ 健康检查正确
- ✅ 统计信息准确
- ✅ 测试覆盖率 > 80%

**依赖**：

- 前置任务：所有之前任务

---

### 任务10：完整集成和优化

**目标**：完整集成所有功能，性能优化，文档完善

**需求清单**：

1. 需求10.1：端到端测试
2. 需求10.2：性能测试和优化
3. 需求10.3：Docker 镜像构建
4. 需求10.4：部署文档
5. 需求10.5：API 文档

**可演示功能**：

```bash
# 完整流程测试
# 1. 启动所有服务（Redis、PostgreSQL、Client、Mock MCP Server）
docker-compose up -d

# 2. 创建会话
curl -X POST http://localhost:8080/api/sessions ...

# 3. 发送消息
curl -X POST http://localhost:8080/api/sessions/{id}/chat ...

# 4. 接收 WebSocket 推送
websocat ws://localhost:8080/ws/sessions/{id}?key=...

# 5. 触发持久化
curl -X POST http://localhost:8080/api/system/persistence/trigger

# 6. 停止服务
docker-compose down

# 7. 重新启动，数据恢复
docker-compose up -d
```

**测试用例**：

| 需求   | 测试用例          | 测试类型  |
| ---- | ------------- | ----- |
| 10.1 | 完整用户流程测试      | 端到端测试 |
| 10.2 | 性能测试（响应时间、并发） | 性能测试  |
| 10.3 | Docker 镜像构建成功 | 集成测试  |
| 10.4 | 部署文档完整准确      | 文档审查  |
| 10.5 | API 文档完整准确    | 文档审查  |

**验收标准**：

- ✅ 端到端测试全部通过
- ✅ 性能满足要求（< 1ms Redis 访问，< 5s 聊天响应）
- ✅ Docker 镜像可以正常部署
- ✅ 文档完整
- ✅ 代码覆盖率 > 80%

**依赖**：

- 前置任务：所有之前任务

---

## 开发顺序建议

### 推荐顺序（按依赖关系）

```
任务1: 项目脚手架 + Redis 基础存储
  ↓
任务2: PostgreSQL 持久化
  ↓
任务3: HTTP API - 会话管理
  ↓
任务4: HTTP API - 消息管理 (使用 Mock LLM)
  ↓
任务5: WebSocket 实时通信 (使用 Mock 事件)
  ↓
任务6: LLM 集成 (可选真实 LLM，支持 Mock)
  ↓
任务7: MCP Server 集成 (使用 Mock MCP Server)
  ↓
任务8: 持久化触发器
  ↓
任务9: 系统监控和日志
  ↓
任务10: 完整集成和优化
```

### 并行开发（可同时进行的任务）

- 任务6 和 任务7 可以并行开发（都基于 Mock）
- 任务8 和 任务9 可以并行开发
- 任务10 必须在所有任务完成后进行

---

## 每个任务的标准交付物

### 代码交付

### ### 测试交付

注意测试交付的代码应该在tests文件夹

1. **单元测试**：测试单个函数和类
   
   - 命名：`{filename}_test.go`
   - 运行：`go test ./internal/{module} -v`

2. **集成测试**：测试多个组件协作
   
   - 命名：`{filename}_integration_test.go`
   - 运行：`go test ./internal/{module} -tags=integration -v`

3. **HTTP 测试**：测试 HTTP API
   
   - 使用 `httptest`
   - 运行：`go test ./internal/api -v`

4. **端到端测试**：完整流程测试
   
   - 命名：`e2e_test.go`
   - 运行：`go test ./e2e -v`

### 文档交付

1. **README.md**：项目介绍和快速开始
2. **API.md**：API 文档
3. **DEVELOPMENT.md**：开发指南
4. **CHANGELOG.md**：变更日志

---

## 测试策略详细说明

### Mock 实现

#### 1. LLM Client Mock

```go
// internal/llm/mock.go
type MockLLMClient struct {
    responses map[string]string // 预设响应
}

func (m *MockLLMClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
    return &ChatCompletionResponse{
        Content: m.responses[req.Messages[len(req.Messages)-1].Content],
    }, nil
}
```

#### 2. MCP Server Mock

```go
// internal/mcp/mock_server.go
func StartMockMCPServer(t *testing.T) *httptest.Server {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 模拟 MCP 协议响应
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "result": map[string]interface{}{
                "success": true,
                "data": "mock result",
            },
        })
    }))
    return server
}
```

### 测试数据库

```go
// 使用 testcontainers
func setupTestDB(t *testing.T) (*sql.DB, func()) {
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
    )
    if err != nil {
        t.Fatal(err)
    }

    connStr, _ := container.ConnectionString(ctx)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        t.Fatal(err)
    }

    // 运行迁移
    runMigrations(db)

    cleanup := func() {
        db.Close()
        container.Terminate(ctx)
    }

    return db, cleanup
}
```

---

## 里程碑和时间估算

| 里程碑           | 任务   | 预估时间 | 状态  |
| ------------- | ---- | ---- | --- |
| M1: 基础设施      | 任务1  | 2天   | ⬜   |
| M2: 数据持久化     | 任务2  | 2天   | ⬜   |
| M3: REST API  | 任务3  | 2天   | ⬜   |
| M4: 消息 API    | 任务4  | 2天   | ⬜   |
| M5: WebSocket | 任务5  | 2天   | ⬜   |
| M6: LLM 集成    | 任务6  | 2天   | ⬜   |
| M7: MCP 集成    | 任务7  | 3天   | ⬜   |
| M8: 触发器       | 任务8  | 2天   | ⬜   |
| M9: 监控        | 任务9  | 1天   | ⬜   |
| M10: 完整集成     | 任务10 | 3天   | ⬜   |

**总计**：约 23 天

---

# 

## 总结

本开发计划遵循以下原则：

1. ✅ **小步迭代**：10 个任务，每个任务 1-3 天
2. ✅ **可运行**：每个任务结束都是完整的、可运行的程序
3. ✅ **可测试**：每个任务都有完整的测试，使用 Mock 隔离外部依赖
4. ✅ **增量交付**：从简单到复杂，逐步构建完整系统
5. ✅ **独立测试**：即使没有真实的 MCP Server 和 LLM，也可以使用 Mock 进行集成测试

每个任务完成后，都有一个可演示的里程碑，确保开发进度和质量可见！
