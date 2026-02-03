# DND MCP Client 详细设计文档

## 文档信息

- **版本**: v1.1
- **创建日期**: 2025-02-03
- **更新日期**: 2025-02-03
- **基于**: DND_MCP_架构设计_方案B修订.md
- **设计范围**: MCP Client 组件详细设计
- **主要更新**:
  - 添加持久化触发器设计（支持多种触发策略）
  - 简化 API 设计（删除分页、排序等复杂参数）

---

## 一、Client 架构概览

### 1.1 定位与职责

```
MCP Client = 轻量级有状态协调层

核心职责：
  ├─ 会话管理（创建、查询、删除）
  ├─ 对话消息管理（存储、检索、上下文构建）
  ├─ AI 对话编排（LLM 调用、工具执行）
  ├─ 实时通信（WebSocket 推送）
  ├─ 数据持久化（Redis → PostgreSQL 定期备份）
  └─ MCP Server 协调（工具调用、事件订阅）

非职责：
  └─ 游戏状态管理（由 MCP Server 负责）
```

### 1.2 技术栈

```
语言: Go 1.21+

核心依赖：
  ├─ Web 框架: gin-gonic/gin (HTTP) + gorilla/websocket (WebSocket)
  ├─ 数据库:
  │   ├─ Redis: go-redis/redis/v9 (主存储)
  │   └─ PostgreSQL: jackc/pgx/v5 (备份存储)
  ├─ MCP 协议: 自定义 MCP Client SDK
  ├─ LLM 集成: openai/openai-go (或兼容接口)
  └─ 配置管理: spf13/viper

开发工具：
  ├─ 依赖管理: go modules
  ├─ 代码生成: go generate (wire 依赖注入)
  └─ 测试: testing + testify
```

---

## 二、核心组件设计

### 2.1 组件架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Layer                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Session API  │  │  Chat API    │  │  System API  │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│  ┌────────────────────────────────────────────────────────────┐│
│  │                  SessionService                             ││
│  │   • CreateSession()                                         ││
│  │   • GetSession()                                            ││
│  │   • ListSessions()                                          ││
│  │   • DeleteSession()                                         ││
│  └────────────────────────────────────────────────────────────┘│
│  ┌────────────────────────────────────────────────────────────┐│
│  │                  ChatService                                ││
│  │   • SendMessage()                                           ││
│  │   • GetMessages()                                           ││
│  │   • StreamMessage() (SSE)                                   ││
│  └────────────────────────────────────────────────────────────┘│
│  ┌────────────────────────────────────────────────────────────┐│
│  │                  OrchestratorService                        ││
│  │   ┌──────────────────┐  ┌──────────────────┐              ││
│  │   │ Context Builder  │  │ Tool Coordinator │              ││
│  │   └──────────────────┘  └──────────────────┘              ││
│  │   ┌──────────────────┐  ┌──────────────────┐              ││
│  │   │ LLM Manager      │  │ Event Handler    │              ││
│  │   └──────────────────┘  └──────────────────┘              ││
│  └────────────────────────────────────────────────────────────┘│
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Data Layer                                 │
│  ┌────────────────────────────────────────────────────────────┐│
│  │                  Repository Interface                      ││
│  │   ┌────────────────┐  ┌────────────────┐                 ││
│  │   │SessionRepository│  │MessageRepository│                ││
│  │   └────────────────┘  └────────────────┘                 ││
│  └────────────────────────────────────────────────────────────┘│
│  ┌────────────────────────────────────────────────────────────┐│
│  │                  Store Implementation                       ││
│  │   ┌────────────────┐  ┌────────────────┐                 ││
│  │   │ Redis Store    │  │ Postgres Store │                 ││
│  │   │ (主存储)        │  │  (备份)        │                 ││
│  │   └────────────────┘  └────────────────┘                 ││
│  └────────────────────────────────────────────────────────────┘│
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Redis Client │  │ PGX Client   │  │ MCP Client   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ LLM Client   │  │ WS Manager   │  │ Event Bus    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 目录结构

```
dnd-client/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/
│   ├── api/                       # HTTP handlers
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   └── logging.go
│   │   ├── handler/
│   │   │   ├── session.go
│   │   │   ├── chat.go
│   │   │   ├── system.go
│   │   │   └── websocket.go
│   │   └── router.go
│   ├── service/                   # 业务逻辑层
│   │   ├── session.go
│   │   ├── chat.go
│   │   └── orchestrator/
│   │       ├── context_builder.go
│   │       ├── llm_manager.go
│   │       ├── tool_coordinator.go
│   │       └── event_handler.go
│   ├── repository/                # 数据访问接口
│   │   ├── session.go
│   │   └── message.go
│   ├── store/                     # 存储实现
│   │   ├── redis/
│   │   │   ├── client.go
│   │   │   ├── session.go
│   │   │   ├── message.go
│   │   │   └── system.go
│   │   └── postgres/
│   │       ├── client.go
│   │       ├── session.go
│   │       ├── message.go
│   │       └── migrations/
│   ├── domain/                    # 领域模型
│   │   ├── session.go
│   │   ├── message.go
│   │   └── events.go
│   ├── mcp/                       # MCP 协议集成
│   │   ├── client.go
│   │   ├── tools.go
│   │   └── transport.go
│   ├── llm/                       # LLM 集成
│   │   ├── client.go
│   │   ├── openai.go
│   │   └── types.go
│   ├── ws/                        # WebSocket 管理
│   │   ├── manager.go
│   │   ├── hub.go
│   │   └── connection.go
│   └── persistence/               # 持久化管理
│       ├── manager.go
│       ├── worker.go
│       └── trigger/                # 触发器策略
│           ├── trigger.go          # 触发器接口
│           ├── time.go             # 时间触发器
│           └── message.go          # 消息量触发器（预留）
├── pkg/
│   ├── config/                    # 配置管理
│   │   └── config.go
│   ├── logger/                    # 日志
│   │   └── logger.go
│   └── errors/                    # 错误定义
│       └── errors.go
├── configs/
│   ├── config.yaml
│   └── config.dev.yaml
├── deployments/
│   ├── docker-compose.yml
│   └── Dockerfile
├── scripts/
│   ├── build.sh
│   └── migrate.sh
├── go.mod
├── go.sum
└── README.md
```

---

## 三、执行流程设计

### 3.1 会话创建流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      会话创建流程                                │
└─────────────────────────────────────────────────────────────────┘

触发: POST /api/sessions

参与者:
  Frontend → Handler → SessionService → Repository → Store → MCP Server

流程步骤:

1. 接收请求
   ┌─────────────────────────────────────────────────────────────┐
   │ Handler.CreateSession                                       │
   │   输入: {                                                   │
   │     "name": "我的第一个战役",                                │
   │     "creator_id": "user-123",                               │
   │     "mcp_server_url": "http://localhost:8080",             │
   │     "settings": {                                           │
   │       "max_players": 5,                                     │
   │       "ruleset": "dnd5e"                                    │
   │     }                                                       │
   │   }                                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
2. 参数验证
   ┌─────────────────────────────────────────────────────────────┐
   │ • 验证必填字段                                               │
   │ • 验证 MCP Server URL 可访问性                               │
   │ • 验证 settings 格式                                         │
   │ • 生成 UUID (session_id)                                    │
   │ • 生成 WebSocket Key                                        │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 调用服务层
   ┌─────────────────────────────────────────────────────────────┐
   │ SessionService.CreateSession                                │
   │   session := &domain.Session{                               │
   │     ID:           uuid,                                    │
   │     Name:         req.Name,                                │
   │     CreatorID:    req.CreatorID,                           │
   │     MCPServerURL: req.MCPServerURL,                        │
   │     WebSocketKey: websocketKey,                            │
   │     CreatedAt:    time.Now(),                              │
   │   }                                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 保存到 Redis
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.CreateSession                                   │
   │   Pipeline:                                                 │
   │     HSET session:{uuid}                                    │
   │       id "uuid"                                            │
   │       name "我的第一个战役"                                  │
   │       creator_id "user-123"                                │
   │       mcp_server_url "http://localhost:8080"               │
   │       websocket_key "ws-xxx"                                │
   │       created_at "2025-02-03T10:00:00Z"                    │
   │                                                              │
   │     SADD sessions:all {uuid}                               │
   │                                                              │
   │   EXEC                                                      │
   │   耗时: <1ms                                                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 初始化 MCP Server 连接
   ┌─────────────────────────────────────────────────────────────┐
   │ MCPClient.Initialize                                       │
   │   • 连接到 MCP Server                                       │
   │   • 执行 initialize 握手                                     │
   │   • 订阅 Server 事件                                         │
   │   • 验证连接成功                                             │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 返回响应
   ┌─────────────────────────────────────────────────────────────┐
   │ Response:                                                   │
   │   {                                                         │
   │     "id": "uuid-xxx",                                       │
   │     "name": "我的第一个战役",                                │
   │     "creator_id": "user-123",                               │
   │     "mcp_server_url": "http://localhost:8080",             │
   │     "websocket_key": "ws-xxx",                              │
   │     "created_at": "2025-02-03T10:00:00Z",                  │
   │     "status": "active"                                      │
   │   }                                                         │
   │   HTTP Status: 201 Created                                  │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
7. (后台) 标记需要持久化
   ┌─────────────────────────────────────────────────────────────┐
   │ PersistenceManager.MarkDirty                                │
   │   • 将 session_id 加入 dirty 队列                            │
   │   • 等待下一次持久化周期                                     │
   └─────────────────────────────────────────────────────────────┘

总耗时: ~50ms (主要是 MCP Server 连接时间)
Redis 操作耗时: <1ms
```

### 3.2 聊天对话流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      聊天对话流程                                │
└─────────────────────────────────────────────────────────────────┘

触发: POST /api/sessions/{id}/chat

参与者:
  Frontend → Handler → ChatService → Orchestrator → LLM/MCP Server

流程步骤:

1. 接收用户消息
   ┌─────────────────────────────────────────────────────────────┐
   │ Handler.SendMessage                                         │
   │   输入: {                                                   │
   │     "session_id": "uuid-xxx",                               │
   │     "content": "我要攻击哥布林",                             │
   │     "player_id": "player-123",                              │
   │     "stream": false                                         │
   │   }                                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
2. 保存用户消息
   ┌─────────────────────────────────────────────────────────────┐
   │ MessageRepository.Save                                      │
   │   userMsg := &domain.Message{                               │
   │     ID:        newUUID(),                                  │
   │     SessionID: sessionID,                                  │
   │     Role:      "user",                                     │
   │     Content:   "我要攻击哥布林",                            │
   │     PlayerID:  "player-123",                               │
   │     CreatedAt: time.Now(),                                 │
   │   }                                                         │
   │   RedisStore.SaveMessage(userMsg)                          │
   │   操作: ZADD msg:{session_id} {timestamp} {json}           │
   │   耗时: <1ms                                                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 构建对话上下文
   ┌─────────────────────────────────────────────────────────────┐
   │ ContextBuilder.Build                                        │
   │   步骤:                                                     │
   │   a) 从 Redis 加载历史消息                                   │
   │      ZREVRANGE msg:{session_id} 0 49                        │
   │      获取最近 50 条消息                                      │
   │      耗时: <1ms                                             │
   │                                                              │
   │   b) 从 MCP Server 获取游戏状态摘要                           │
   │      MCPClient.GetSessionState(session_id)                  │
   │      返回: {                                                │
   │        "location": "幽暗森林",                               │
   │        "characters": ["勇者A", "哥布林"],                   │
   │        "game_time": "第3天 10:30"                           │
   │      }                                                      │
   │      耗时: 10-50ms                                          │
   │                                                              │
   │   c) 组装 LLM 上下文                                         │
   │      context := []llm.Message{                              │
   │        {                                                     │
   │          Role:    "system",                                │
   │          Content: systemPrompt + stateSummary,             │
   │        },                                                    │
   │        // 历史消息...                                        │
   │        {                                                     │
   │          Role:    "user",                                  │
   │          Content: "我要攻击哥布林",                         │
   │        },                                                    │
   │      }                                                       │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 调用 LLM
   ┌─────────────────────────────────────────────────────────────┐
   │ LLMManager.ChatCompletion                                   │
   │   request := llm.ChatCompletionRequest{                     │
   │     Model:    "gpt-4",                                     │
   │     Messages: context,                                      │
   │     Tools:    availableTools,                               │
   │   }                                                         │
   │   response := llmClient.Chat(ctx, request)                 │
   │   耗时: 1-3s                                                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 判断响应类型
   ┌─────────────────────────────────────────────────────────────┐
   │ 情况 A: 有 tool_calls (需要调用工具)                          │
   │                                                              │
   │   a) 转换为 MCP 工具调用                                      │
   │      ToolCoordinator.ConvertToMCP(tool_calls)               │
   │      转换:                                                   │
   │        {                                                     │
   │          "name": "resolve_attack",                          │
   │          "arguments": {                                     │
   │            "attacker": "player-123",                        │
   │            "target": "goblin-1",                            │
   │            "attack_type": "melee"                           │
   │          }                                                   │
   │        }                                                     │
   │                                                              │
   │   b) 调用 MCP Server 执行工具                                 │
   │      MCPClient.CallTool(session_id, tool_name, args)        │
   │      耗时: 10-100ms                                         │
   │                                                              │
   │   c) 保存 tool 消息到 Redis                                   │
   │      toolMsg := &domain.Message{                            │
   │        Role:      "tool",                                  │
   │        Content:   "",                                      │
   │        ToolCalls: tool_calls,                              │
   │      }                                                       │
   │      RedisStore.SaveMessage(toolMsg)                        │
   │                                                              │
   │   d) 保存 tool 响应消息到 Redis                               │
   │      toolResponse := &domain.Message{                       │
   │        Role:      "tool",                                  │
   │        Content:   result,                                  │
   │      }                                                       │
   │      RedisStore.SaveMessage(toolResponse)                   │
   │                                                              │
   │   e) 回到步骤 4 继续调用 LLM（带工具结果）                     │
   │                                                              │
   │ 情况 B: 无 tool_calls (纯文本响应)                            │
   │                                                              │
   │   保存助手响应到 Redis                                       │
   │   assistantMsg := &domain.Message{                          │
   │     Role:    "assistant",                                  │
   │     Content: response.Content,                             │
   │   }                                                         │
   │   RedisStore.SaveMessage(assistantMsg)                      │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 返回响应给前端
   ┌─────────────────────────────────────────────────────────────┐
   │ Response:                                                   │
   │   {                                                         │
   │     "id": "msg-uuid-xxx",                                   │
   │     "role": "assistant",                                   │
   │     "content": "你冲向哥布林，挥舞着手中的长剑...",          │
   │     "tool_calls": [...],                                   │
   │     "state_changes": {                                      │
   │       "combat": "active",                                  │
   │       "turn": "player-123"                                 │
   │     },                                                      │
   │     "created_at": "2025-02-03T10:05:30Z"                   │
   │   }                                                         │
   │   HTTP Status: 200 OK                                      │
   └─────────────────────────────────────────────────────────────┘

总耗时: ~2-5秒 (主要是 LLM 调用)
数据访问耗时: ~2ms (Redis + MCP Server)

多轮工具调用循环:
  用户消息 → LLM → tool_1 → Server → LLM → tool_2 → Server → LLM → 最终响应
```

### 3.3 数据持久化流程（后台）

```
┌─────────────────────────────────────────────────────────────────┐
│                   数据持久化流程（后台）                         │
└─────────────────────────────────────────────────────────────────┘

触发: 持久化触发器（支持多种策略）

当前实现: 时间触发器（每 30 秒）
预留扩展: 消息量触发器、手动触发器、组合触发器

触发器设计:

  interface PersistenceTrigger {
    // 判断是否应该触发持久化
    ShouldTrigger(ctx context.Context) (bool, error)
    // 重置触发器状态
    Reset(ctx context.Context) error
  }

  实现的触发器:
    • TimeTrigger: 时间间隔触发（当前使用）
      - 配置: interval = 30s
      - 逻辑: 每隔 30 秒触发一次

    • MessageCountTrigger: 消息数量触发（预留）
      - 配置: threshold = 100 条消息
      - 逻辑: 当新增消息达到阈值时触发

    • ManualTrigger: 手动触发（预留）
      - 配置: 通过 HTTP API 触发
      - 逻辑: 收到触发信号时立即执行

    • CompositeTrigger: 组合触发器（预留）
      - 配置: 多个触发器的组合
      - 逻辑: 任一触发器满足条件即触发

参与者:
  Trigger → PersistenceManager → Repository → PostgresStore

流程步骤:

1. 触发器检测
   ┌─────────────────────────────────────────────────────────────┐
   │ TimeTrigger.ShouldTrigger()                                │
   │   • 检查距离上次持久化的时间                                │
   │   • 如果 >= 30 秒，返回 true                                │
   │   • 否则返回 false                                         │
   │                                                              │
   │ 扩展: 未来可以同时注册多个触发器                             │
   │   • 任一触发器返回 true 即执行持久化                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼ (shouldTrigger = true)
2. 检查系统状态
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.GetSystemStatus                                │
   │   status = GET system:status                              │
   │   if status != "ready":                                    │
   │     跳过本次持久化                                          │
   │     return                                                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 设置持久化状态
   ┌─────────────────────────────────────────────────────────────┐
   │ SET system:status "persistence_in_progress"                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 获取所有会话
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.GetAllSessionIDs                               │
   │   sessionIDs = SMEMBERS sessions:all                       │
   │   返回: [uuid-1, uuid-2, uuid-3, ...]                      │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 遍历每个会话
   ┌─────────────────────────────────────────────────────────────┐
   │ 对每个 session_id:                                         │
   │   a) 读取会话元数据                                          │
   │      HGETALL session:{uuid}                                │
   │      返回: {id, name, creator_id, mcp_server_url, ...}      │
   │                                                              │
   │   b) 读取该会话的所有消息                                    │
   │      ZREVRANGE msg:{uuid} 0 -1 WITH SCORES                 │
   │      返回: [{message, score}, ...]                          │
   │                                                              │
   │   c) 写入 PostgreSQL                                        │
   │      PostgresStore.UpsertSession(session)                  │
   │      INSERT INTO client_sessions (...)                     │
   │      VALUES (...)                                          │
   │      ON CONFLICT (id) DO UPDATE SET ...                    │
   │                                                              │
   │      PostgresStore.BatchInsertMessages(messages)           │
   │      INSERT INTO client_messages (...)                     │
   │      VALUES (...)                                          │
   │      ON CONFLICT (id) DO NOTHING                           │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 更新系统状态
   ┌─────────────────────────────────────────────────────────────┐
   │ SET system:last_persistence {timestamp}                    │
   │ SET system:status "ready"                                  │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
7. 重置触发器
   ┌─────────────────────────────────────────────────────────────┐
   │ TimeTrigger.Reset()                                        │
   │   • 更新上次持久化时间戳                                    │
   │   • 准备下一次触发检测                                      │
   │                                                              │
   │ 扩展: 重置所有已注册的触发器                                 │
   │   • MessageCountTrigger: 重置消息计数                        │
   │   • ManualTrigger: 清除触发信号                             │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
8. 记录日志
   ┌─────────────────────────────────────────────────────────────┐
   │ logger.Info(                                              │
   │   "持久化完成",                                            │
   │     "trigger", "TimeTrigger",                             │
   │     "sessions", sessionCount,                             │
   │     "messages", messageCount,                             │
   │     "duration", duration.Since(start),                    │
   │ )                                                          │
   │   示例: "持久化完成: trigger=TimeTrigger, 5个会话, 1200条消息, 耗时450ms"
   └─────────────────────────────────────────────────────────────┘

总耗时: ~500ms (假设 5000 条消息)
性能优化:
  • 使用 PostgreSQL 批量插入 (COPY)
  • 使用 Pipeline 批量读取 Redis
  • 可以使用 goroutine 并行处理多个会话

扩展点:
  • 新增触发器: 实现 PersistenceTrigger 接口
  • 组合触发器: 注册多个触发器，任一满足即触发
  • 触发器配置: 通过配置文件动态配置触发器参数
```

### 3.4 WebSocket 实时通信流程

```
┌─────────────────────────────────────────────────────────────────┐
│                  WebSocket 实时通信流程                         │
└─────────────────────────────────────────────────────────────────┘

触发: 前端建立 WebSocket 连接

参与者:
  Frontend → WS Handler → WS Manager → Hub → Connection

流程步骤:

1. 建立 WebSocket 连接
   ┌─────────────────────────────────────────────────────────────┐
   │ Frontend:                                                  │
   │   ws://localhost:8080/ws/sessions/{session_id}?key=ws-xxx  │
   │                                                              │
   │ Handler:                                                   │
   │   • 验证 session_id 存在                                     │
   │   • 验证 websocket_key 匹配                                  │
   │   • 升级 HTTP → WebSocket                                   │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
2. 注册连接
   ┌─────────────────────────────────────────────────────────────┐
   │ WSManager.RegisterConnection                               │
   │   conn := &Connection{                                     │
   │     ID:        connID,                                    │
   │     SessionID: sessionID,                                 │
   │     PlayerID:  playerID,                                  │
   │     WebSocket: ws,                                         │
   │     Send:      make(chan []byte, 256),                    │
   │   }                                                         │
   │   Hub.Register <- conn                                     │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 启动读写 goroutine
   ┌─────────────────────────────────────────────────────────────┐
   │ go conn.readPump()   # 读取客户端消息                       │
   │ go conn.writePump()  # 写入消息到客户端                     │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 接收 MCP Server 事件
   ┌─────────────────────────────────────────────────────────────┐
   │ EventHandler.OnServerEvent                                 │
   │   事件类型:                                                  │
   │   • state_changed: 游戏状态变更                              │
   │   • combat_updated: 战斗状态更新                             │
   │   • character_moved: 角色移动                               │
   │   • dice_rolled: 骰子投掷结果                                │
   │                                                              │
   │   推送到前端:                                               │
   │     Hub.Broadcast <- Event{                                │
   │       SessionID: sessionID,                                │
   │       Type:      "state_changed",                          │
   │       Data:      eventData,                                │
   │     }                                                       │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 推送到前端
   ┌─────────────────────────────────────────────────────────────┐
   │ conn.writePump():                                         │
   │   select {                                                 │
   │   case message := <-conn.Send:                             │
   │     err := ws.WriteMessage(websocket.TextMessage, message) │
   │     if err != nil:                                         │
   │       Hub.Unregister <- conn                               │
   │     }                                                       │
   │   }                                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 客户端断开连接
   ┌─────────────────────────────────────────────────────────────┐
   │ conn.readPump() 检测到错误                                  │
   │   Hub.Unregister <- conn                                   │
   │   close(conn.Send)                                         │
   │   ws.Close()                                               │
   └─────────────────────────────────────────────────────────────┘

实时推送的消息类型:
  • 新消息通知
  • 游戏状态变更
  • 战斗事件
  • 系统通知
```

### 3.5 启动加载流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      启动加载流程                                │
└─────────────────────────────────────────────────────────────────┘

触发: 应用启动

参与者:
  Main → DataLoader → PostgresStore → RedisStore → HTTP Server

流程步骤:

1. 加载配置
   ┌─────────────────────────────────────────────────────────────┐
   │ config.Load()                                             │
   │   读取 configs/config.yaml                                │
   │   验证配置有效性                                            │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
2. 初始化依赖
   ┌─────────────────────────────────────────────────────────────┐
   │ • 初始化 Logger                                             │
   │ • 初始化 Redis Client                                       │
   │ • 初始化 PostgreSQL Client                                  │
   │ • 初始化 MCP Client                                         │
   │ • 初始化 LLM Client                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 设置启动状态
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.SetSystemStatus("starting")                     │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 加载会话数据
   ┌─────────────────────────────────────────────────────────────┐
   │ DataLoader.LoadSessions()                                 │
   │   a) 从 PostgreSQL 查询                                     │
   │      SELECT * FROM client_sessions                         │
   │      WHERE deleted_at IS NULL                              │
   │      ORDER BY created_at DESC                              │
   │                                                              │
   │   b) 批量写入 Redis                                         │
   │      Pipeline:                                              │
   │        对每个 session:                                      │
   │          HSET session:{uuid} {field} {value}               │
   │          SADD sessions:all {uuid}                          │
   │      Pipeline.Exec()                                       │
   │   耗时: ~100ms (假设 5 个会话)                              │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 加载消息数据
   ┌─────────────────────────────────────────────────────────────┐
   │ DataLoader.LoadMessages()                                 │
   │   a) 对每个 session_id 从 PostgreSQL 查询                   │
   │      SELECT * FROM client_messages                         │
   │      WHERE session_id = $1                                 │
   │      ORDER BY created_at ASC                               │
   │                                                              │
   │   b) 批量写入 Redis                                         │
   │      Pipeline:                                              │
   │        对每个 message:                                      │
   │          ZADD msg:{uuid} {timestamp} {json}                │
   │      Pipeline.Exec()                                       │
   │   耗时: ~500ms (假设 5000 条消息)                           │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 设置就绪状态
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.SetSystemStatus("ready")                        │
   │ RedisStore.SetSystemStartupTime(time.Now())                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
7. 启动后台服务
   ┌─────────────────────────────────────────────────────────────┐
   │ go PersistenceManager.Start()  # 每 30 秒持久化             │
   │ go HTTPServer.Serve()            # HTTP 服务                │
   │ go WSServer.Serve()              # WebSocket 服务           │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
8. 系统就绪
   ┌─────────────────────────────────────────────────────────────┐
   │ logger.Info("系统已启动，开始接收请求")                      │
   └─────────────────────────────────────────────────────────────┘

总耗时: <1秒
```

### 3.6 优雅关闭流程

```
┌─────────────────────────────────────────────────────────────────┐
│                     优雅关闭流程                                 │
└─────────────────────────────────────────────────────────────────┘

触发: SIGTERM / SIGINT 信号

参与者:
  OS → Main → HTTP Server → WS Manager → PersistenceManager

流程步骤:

1. 捕获信号
   ┌─────────────────────────────────────────────────────────────┐
   │ signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)    │
   │   <-sigChan                                                │
   │   logger.Info("收到关闭信号，开始优雅关闭...")                │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
2. 设置停止状态
   ┌─────────────────────────────────────────────────────────────┐
   │ RedisStore.SetSystemStatus("stopping")                     │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
3. 停止接收新请求
   ┌─────────────────────────────────────────────────────────────┐
   │ HTTPServer.Shutdown():                                    │
   │   • 停止监听新端口                                          │
   │   • 设置 ctx.WithTimeout(10s)                              │
   │   • 等待现有请求完成                                        │
   │   • 超时后强制关闭                                          │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
4. 关闭 WebSocket 连接
   ┌─────────────────────────────────────────────────────────────┐
   │ WSManager.Shutdown():                                     │
   │   • 停止接受新连接                                          │
   │   • 发送关闭帧到所有客户端                                   │
   │   • 等待客户端关闭连接（最多 5 秒）                           │
   │   • 关闭所有连接                                            │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
5. 执行最后持久化
   ┌─────────────────────────────────────────────────────────────┐
   │ PersistenceManager.Shutdown():                            │
   │   • 执行完整持久化流程                                       │
   │   • 验证持久化结果                                          │
   │   • 关闭 PostgreSQL 连接                                    │
   │   • 关闭 Redis 连接                                         │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
6. 关闭其他资源
   ┌─────────────────────────────────────────────────────────────┐
   │ • 关闭 MCP Client 连接                                      │
   │ • 关闭 LLM Client 连接                                      │
   │ • 刷新日志缓冲                                              │
   └─────────────────────────────────────────────────────────────┘
                │
                ▼
7. 进程退出
   ┌─────────────────────────────────────────────────────────────┐
   │ logger.Info("优雅关闭完成")                                  │
   │ os.Exit(0)                                                 │
   └─────────────────────────────────────────────────────────────┘

总耗时: <15秒
超时保护: 10 秒后强制退出
```

### 3.7 持久化触发器设计

```
┌─────────────────────────────────────────────────────────────────┐
│                    持久化触发器设计                              │
└─────────────────────────────────────────────────────────────────┘

设计目标:
  • 支持多种触发策略
  • 易于扩展新的触发器
  • 支持触发器组合
  • 配置化管理

触发器接口定义:

  type PersistenceTrigger interface {
    // ShouldTrigger 判断是否应该触发持久化
    ShouldTrigger(ctx context.Context) (bool, error)

    // Reset 重置触发器状态
    Reset(ctx context.Context) error

    // Name 返回触发器名称
    Name() string
  }

已实现的触发器:

1. TimeTrigger (时间触发器)
   ├─ 描述: 按固定时间间隔触发
   ├─ 配置参数:
   │   └─ interval: time.Duration (如 30s)
   ├─ 触发逻辑:
   │   └─ 距离上次持久化时间 >= interval
   └─ 使用场景:
       ├─ 当前默认使用
       └─ 适合数据量小、要求定时备份的场景

2. MessageCountTrigger (消息量触发器 - 预留)
   ├─ 描述: 当新增消息达到阈值时触发
   ├─ 配置参数:
   │   └─ threshold: int (如 100 条)
   ├─ 触发逻辑:
   │   └─ 自上次持久化以来新增消息数 >= threshold
   ├─ 实现要点:
   │   ├─ 每次保存消息时计数 +1
   │   ├─ 持久化完成后重置计数
   │   └─ 计数器存储在 Redis (persistence:msg_count)
   └─ 使用场景:
       ├─ 消息频繁时及时持久化
       └─ 防止消息丢失过多

3. ManualTrigger (手动触发器 - 预留)
   ├─ 描述: 通过 API 手动触发持久化
   ├─ 配置参数:
   │   └─ 无
   ├─ 触发逻辑:
   │   └─ 收到触发信号时立即返回 true
   ├─ 实现要点:
   │   ├─ 使用 channel 接收触发信号
   │   │   triggerChan chan struct{}
   │   ├─ 提供 HTTP API: POST /api/system/persistence/trigger
   │   └─ 管理员可随时触发
   └─ 使用场景:
       ├─ 关键操作后立即持久化
       ├─ 系统维护前备份
       └─ 调试和测试

4. CompositeTrigger (组合触发器 - 预留)
   ├─ 描述: 组合多个触发器，任一满足即触发
   ├─ 配置参数:
   │   └─ triggers: []PersistenceTrigger
   ├─ 触发逻辑:
   │   └─ 任一子触发器 ShouldTrigger() 返回 true
   ├─ 实现要点:
   │   │   func (c *CompositeTrigger) ShouldTrigger(ctx context.Context) (bool, error) {
   │   │     for _, trigger := range c.triggers {
   │   │       if shouldTrigger, err := trigger.ShouldTrigger(ctx); err != nil {
   │   │         return false, err
   │   │       } else if shouldTrigger {
   │   │         return true, nil
   │   │       }
   │   │     }
   │   │     return false, nil
   │   │   }
   │   └─ Reset 时重置所有子触发器
   └─ 使用场景:
       ├─ 时间 + 消息量双重保障
       ├─ 定时 + 手动混合模式
       └─ 灵活配置触发策略

触发器注册和管理:

  type TriggerManager struct {
    triggers []PersistenceTrigger
  }

  // 注册触发器
  func (m *TriggerManager) Register(trigger PersistenceTrigger)

  // 检查任一触发器是否满足条件
  func (m *TriggerManager) ShouldTrigger(ctx context.Context) (bool, error)

  // 重置所有触发器
  func (m *TriggerManager) ResetAll(ctx context.Context) error

  // 获取触发的触发器名称
  func (m *TriggerManager) GetTriggeredNames(ctx context.Context) []string

配置示例 (config.yaml):

  persistence:
    type: composite              # 组合触发器
    triggers:
      - type: time              # 时间触发器
        interval: 30s

      - type: message           # 消息量触发器
        threshold: 100

      # - type: manual          # 手动触发器（可选）

扩展新的触发器:

  1. 实现 PersistenceTrigger 接口
     type CustomTrigger struct {
       // 自定义字段
     }

     func (t *CustomTrigger) ShouldTrigger(ctx context.Context) (bool, error) {
       // 自定义触发逻辑
       return true, nil
     }

     func (t *CustomTrigger) Reset(ctx context.Context) error {
       // 重置逻辑
       return nil
     }

     func (t *CustomTrigger) Name() string {
       return "CustomTrigger"
     }

  2. 注册到 TriggerManager
     manager.Register(&CustomTrigger{...})

  3. 在配置文件中配置（如需）

当前实现:

  • 仅使用 TimeTrigger
  • 每 30 秒触发一次
  • 配置: PERSISTENCE_TYPE=time, PERSISTENCE_INTERVAL=30s

未来扩展:

  • 实现 MessageCountTrigger
  • 实现 ManualTrigger (添加 HTTP API)
  • 实现 CompositeTrigger 支持多触发器组合
  • 添加更多触发器类型（如: 内存使用量、时间段等）
```

---

## 四、组件交互设计

### 4.1 组件交互矩阵

```
┌─────────────────────────────────────────────────────────────────┐
│                    组件交互矩阵                                  │
└─────────────────────────────────────────────────────────────────┘

                    ┌─────────┐
                    │  前端    │
                    └────┬────┘
                         │ HTTP/WebSocket
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP Layer                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │SessionHandler│  │ ChatHandler  │  │WSHandler     │         │
│  └───────┬──────┘  └───────┬──────┘  └───────┬──────┘         │
└──────────┼──────────────────┼──────────────────┼───────────────┘
           │                  │                  │
           ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Service Layer                                │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │ SessionService   │  │  ChatService     │                    │
│  └────────┬─────────┘  └────────┬─────────┘                    │
│           │                     │                               │
│           │         ┌───────────▼──────────────┐               │
│           │         │  OrchestratorService    │               │
│           │         │  ┌────────────────────┐ │               │
│           │         │  │ ContextBuilder     │ │               │
│           │         │  └─────────┬──────────┘ │               │
│           │         │  ┌─────────▼──────────┐ │               │
│           │         │  │ LLMManager        │ │               │
│           │         │  └─────────┬──────────┘ │               │
│           │         │  ┌─────────▼──────────┐ │               │
│           │         │  │ ToolCoordinator    │ │               │
│           │         │  └─────────┬──────────┘ │               │
│           │         └───────────┼────────────┘               │
│           ▼                     ▼                               │
└──────────┼─────────────────────┼───────────────────────────────┘
           │                     │
           ▼                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Repository Layer                             │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │SessionRepository │  │MessageRepository │                    │
│  └────────┬─────────┘  └────────┬─────────┘                    │
└───────────┼────────────────────┼───────────────────────────────┘
            │                    │
            ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Store Layer                                  │
│  ┌──────────────────┐  ┌──────────────────┐                    │
│  │  RedisStore      │  │  PostgresStore   │                    │
│  │  (主存储)         │  │  (备份)          │                    │
│  └────────┬─────────┘  └────────┬─────────┘                    │
└───────────┼────────────────────┼───────────────────────────────┘
            │                    │
            ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│              Infrastructure Layer                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │  Redis   │  │PostgreSQL│  │   MCP    │  │   LLM    │       │
│  │  Client  │  │  Client  │  │  Client  │  │  Client  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 组件通信方式

```
┌─────────────────────────────────────────────────────────────────┐
│                   组件通信方式                                   │
└─────────────────────────────────────────────────────────────────┘

1. 同步调用 (函数调用)
   Handler → Service → Repository → Store
   特点:
   • 直接函数调用
   • 等待结果返回
   • 用于请求-响应模式

2. 异步通信 (Channel)
   Orchestrator → LLM Manager (streaming)
   MCP Server → Event Handler → WS Manager
   特点:
   • 使用 Go channel
   • 非阻塞通信
   • 用于实时推送

3. 定时触发 (Ticker)
   Timer → PersistenceManager
   特点:
   • time.Ticker
   • 固定间隔触发
   • 用于后台任务

4. 事件驱动 (Event Bus)
   MCP Server → Event Handler → Multiple Subscribers
   特点:
   • 发布-订阅模式
   • 一对多通信
   • 用于状态变更通知
```

### 4.3 关键交互流程

```
┌─────────────────────────────────────────────────────────────────┐
│                交互1: 用户聊天请求                               │
└─────────────────────────────────────────────────────────────────┘

Frontend
    │ POST /api/sessions/{id}/chat
    ▼
ChatHandler.SendMessage()
    │
    ├─→ ChatService.SendMessage()
    │       │
    │       ├─→ MessageRepository.Save() ──→ RedisStore.SaveMessage()
    │       │       (<1ms)
    │       │
    │       ├─→ OrchestratorService.ProcessMessage()
    │       │       │
    │       │       ├─→ ContextBuilder.Build()
    │       │       │       │
    │       │       │       ├─→ MessageRepository.GetHistory()
    │       │       │       │       └─→ RedisStore.GetMessages()
    │       │       │       │           (<1ms)
    │       │       │       │
    │       │       │       └─→ MCPClient.GetSessionState()
    │       │       │           (10-50ms)
    │       │       │
    │       │       ├─→ LLMManager.ChatCompletion()
    │       │       │       └─→ LLM Client (云端)
    │       │       │           (1-3s)
    │       │       │
    │       │       └─→ (如果有 tool_calls)
    │       │               ToolCoordinator.ExecuteTools()
    │       │                   │
    │       │                   └─→ MCPClient.CallTool()
    │       │                       (10-100ms)
    │       │
    │       └─→ MessageRepository.Save() ──→ RedisStore.SaveMessage()
    │
    └─→ Response JSON
            (<1ms)

总耗时: 2-5秒 (主要是 LLM)
```

```
┌─────────────────────────────────────────────────────────────────┐
│                交互2: MCP Server 事件推送                        │
└─────────────────────────────────────────────────────────────────┘

MCP Server
    │ 发送领域事件
    ▼
MCPClient.ReceiveEvent()
    │
    ├─→ EventHandler.OnEvent()
    │       │
    │       ├─→ 解析事件类型
    │       │
    │       ├─→ 转换为前端事件格式
    │       │
    │       └─→ WSManager.Broadcast()
    │               │
    │               └─→ Hub.Broadcast <- Event
    │                       │
    │                       ├─→ connection1.Send <- Event
    │                       ├─→ connection2.Send <- Event
    │                       └─→ connection3.Send <- Event
    │
    └─→ WebSocket 推送到前端
            (<1ms)

总耗时: <1ms
```

```
┌─────────────────────────────────────────────────────────────────┐
│                交互3: 数据持久化（后台）                          │
└─────────────────────────────────────────────────────────────────┘

Timer (每30秒)
    │
    ▼
PersistenceManager.Run()
    │
    ├─→ RedisStore.GetSystemStatus()
    │
    ├─→ RedisStore.GetAllSessionIDs()
    │
    ├─→ 对每个 session_id:
    │       │
    │       ├─→ RedisStore.GetSession() ──→ HGETALL session:{uuid}
    │       │
    │       ├─→ RedisStore.GetMessages() ──→ ZREVRANGE msg:{uuid}
    │       │
    │       ├─→ PostgresStore.UpsertSession()
    │       │       └─→ PostgreSQL: UPSERT client_sessions
    │       │
    │       └─→ PostgresStore.BatchInsertMessages()
    │               └─→ PostgreSQL: INSERT client_messages
    │
    ├─→ RedisStore.SetSystemStatus("ready")
    │
    └─→ logger.Info("持久化完成")

总耗时: ~500ms
```

---

## 五、对外 API 设计

### 5.1 REST API 定义

```
┌─────────────────────────────────────────────────────────────────┐
│                      REST API 设计                              │
└─────────────────────────────────────────────────────────────────┘

Base URL: http://localhost:8080/api
Content-Type: application/json
```

#### 5.1.1 会话管理 API

**1. 创建会话**

```
POST /api/sessions

描述: 创建一个新的游戏会话

请求体:
{
  "name": "我的第一个战役",           // 必填，会话名称
  "creator_id": "user-123",          // 必填，创建者ID
  "mcp_server_url": "http://localhost:9000",  // 必填，MCP Server URL
  "max_players": 5,                  // 可选，最大玩家数，默认4
  "settings": {                      // 可选，会话配置
    "ruleset": "dnd5e",             // 规则集
    "starting_level": 1,            // 起始等级
    "allow_characters": true        // 是否允许自定义角色
  }
}

响应: 201 Created
{
  "id": "session-uuid-xxx",         // 会话ID
  "name": "我的第一个战役",
  "creator_id": "user-123",
  "mcp_server_url": "http://localhost:9000",
  "websocket_key": "ws-key-xxx",    // WebSocket连接密钥
  "max_players": 5,
  "settings": {...},
  "created_at": "2025-02-03T10:00:00Z",
  "updated_at": "2025-02-03T10:00:00Z",
  "status": "active"                // active | archived
}

错误响应:
  400 Bad Request - 参数验证失败
  503 Service Unavailable - MCP Server 连接失败
```

**2. 获取会话详情**

```
GET /api/sessions/{session_id}

描述: 获取指定会话的详细信息

路径参数:
  session_id: 会话ID (UUID)

响应: 200 OK
{
  "id": "session-uuid-xxx",
  "name": "我的第一个战役",
  "creator_id": "user-123",
  "mcp_server_url": "http://localhost:9000",
  "max_players": 5,
  "current_players": 3,              // 当前玩家数
  "settings": {...},
  "created_at": "2025-02-03T10:00:00Z",
  "updated_at": "2025-02-03T10:05:00Z",
  "status": "active",
  "message_count": 120               // 消息总数
}

错误响应:
  404 Not Found - 会话不存在
```

**3. 列出所有会话**

```
GET /api/sessions

查询参数:
  status: 状态过滤（可选），active | archived，默认 active

示例:
  GET /api/sessions
  GET /api/sessions?status=archived

响应: 200 OK
[
  {
    "id": "session-uuid-1",
    "name": "我的第一个战役",
    "creator_id": "user-123",
    "current_players": 3,
    "message_count": 120,
    "created_at": "2025-02-03T10:00:00Z",
    "updated_at": "2025-02-03T10:05:00Z",
    "status": "active"
  },
  {
    "id": "session-uuid-2",
    "name": "第二个战役",
    "creator_id": "user-456",
    "current_players": 2,
    "message_count": 85,
    "created_at": "2025-02-02T15:30:00Z",
    "updated_at": "2025-02-03T09:20:00Z",
    "status": "active"
  }
]

说明:
  • 直接返回会话数组，无分页
  • 按 created_at 倒序排列
  • 适用于会话数量较少的场景（< 100 个）
```

**4. 更新会话**

```
PATCH /api/sessions/{session_id}

描述: 更新会话信息

请求体:
{
  "name": "新的会话名称",              // 可选
  "max_players": 6,                   // 可选
  "settings": {                       // 可选
    "ruleset": "dnd5e"
  }
}

响应: 200 OK
{
  "id": "session-uuid-xxx",
  "name": "新的会话名称",
  // ... 其他字段
}

错误响应:
  404 Not Found - 会话不存在
  400 Bad Request - 参数验证失败
```

**5. 删除会话**

```
DELETE /api/sessions/{session_id}

描述: 删除指定会话（软删除）

路径参数:
  session_id: 会话ID (UUID)

响应: 204 No Content

错误响应:
  404 Not Found - 会话不存在
```

#### 5.1.2 聊天消息 API

**6. 发送消息**

```
POST /api/sessions/{session_id}/chat

描述: 发送用户消息并获取 AI 响应

请求体:
{
  "content": "我要攻击哥布林",        // 必填，消息内容
  "player_id": "player-123",         // 必填，玩家ID
  "stream": false                    // 可选，是否使用流式响应，默认false
}

响应: 200 OK
{
  "id": "msg-uuid-xxx",
  "session_id": "session-uuid-xxx",
  "role": "assistant",              // user | assistant | system | tool
  "content": "你冲向哥布林，挥舞着手中的长剑...",
  "tool_calls": [                   // 如果有工具调用
    {
      "id": "call-xxx",
      "name": "resolve_attack",
      "arguments": {
        "attacker": "player-123",
        "target": "goblin-1",
        "attack_type": "melee"
      },
      "result": {                   // 工具执行结果
        "success": true,
        "damage": 8,
        "events": [...]
      }
    }
  ],
  "state_changes": {                // 状态变更摘要
    "combat": "active",
    "turn": "player-123"
  },
  "player_id": "player-123",
  "created_at": "2025-02-03T10:05:30Z"
}

错误响应:
  404 Not Found - 会话不存在
  400 Bad Request - 参数验证失败
  503 Service Unavailable - LLM 或 MCP Server 不可用
```

**7. 获取消息历史**

```
GET /api/sessions/{session_id}/messages

描述: 获取会话的消息历史

查询参数:
  limit: 返回消息数量（可选），默认 50，最大 100
  role: 角色过滤（可选），user | assistant | system | tool
  since: 起始时间（可选），RFC3339 格式，如 "2025-02-03T10:00:00Z"

示例:
  GET /api/sessions/{id}/messages
  GET /api/sessions/{id}/messages?limit=20
  GET /api/sessions/{id}/messages?role=user&limit=10
  GET /api/sessions/{id}/messages?since=2025-02-03T10:00:00Z

响应: 200 OK
[
  {
    "id": "msg-uuid-1",
    "session_id": "session-uuid-xxx",
    "role": "user",
    "content": "我要攻击哥布林",
    "player_id": "player-123",
    "created_at": "2025-02-03T10:05:00Z"
  },
  {
    "id": "msg-uuid-2",
    "session_id": "session-uuid-xxx",
    "role": "assistant",
    "content": "你冲向哥布林...",
    "created_at": "2025-02-03T10:05:30Z"
  }
]

说明:
  • 直接返回消息数组，无分页
  • 按 created_at 升序排列（从旧到新）
  • 适用于快速加载最近的消息
```

**8. 获取单条消息**

```
GET /api/sessions/{session_id}/messages/{message_id}

描述: 获取指定消息的详细信息

路径参数:
  session_id: 会话ID
  message_id: 消息ID

响应: 200 OK
{
  "id": "msg-uuid-xxx",
  "session_id": "session-uuid-xxx",
  "role": "assistant",
  "content": "你冲向哥布林...",
  "tool_calls": [...],
  "created_at": "2025-02-03T10:05:30Z"
}
```

#### 5.1.3 系统 API

**9. 获取系统状态**

```
GET /api/system/health

描述: 获取系统健康状态

响应: 200 OK
{
  "status": "healthy",              // healthy | degrading | unhealthy
  "version": "1.0.0",
  "uptime": 3600,                   // 运行时长（秒）
  "components": {
    "redis": {
      "status": "healthy",
      "latency_ms": 0.5
    },
    "postgres": {
      "status": "healthy",
      "latency_ms": 5
    },
    "mcp_server": {
      "status": "healthy",
      "connected_sessions": 5
    },
    "llm": {
      "status": "healthy",
      "provider": "openai"
    }
  },
  "last_persistence": "2025-02-03T10:30:00Z"
}
```

**10. 获取系统统计**

```
GET /api/system/stats

描述: 获取系统统计信息

响应: 200 OK
{
  "sessions": {
    "total": 45,
    "active": 30,
    "archived": 15
  },
  "messages": {
    "total": 15000,
    "last_24h": 1200
  },
  "performance": {
    "avg_response_time_ms": 250,
    "p95_response_time_ms": 500,
    "p99_response_time_ms": 1000
  },
  "resources": {
    "memory_used_mb": 256,
    "goroutines": 150
  }
}
```

### 5.2 WebSocket API 定义

```
┌─────────────────────────────────────────────────────────────────┐
│                     WebSocket API 设计                          │
└─────────────────────────────────────────────────────────────────┘

WebSocket URL: ws://localhost:8080/ws/sessions/{session_id}?key={websocket_key}

连接参数:
  session_id: 会话ID
  key: WebSocket密钥（从创建会话API获取）
```

#### 5.2.1 客户端 → 服务器消息

**1. 订阅事件**

```json
{
  "type": "subscribe",
  "data": {
    "events": ["state_changed", "combat_updated", "character_moved"]
  }
}
```

**2. 取消订阅**

```json
{
  "type": "unsubscribe",
  "data": {
    "events": ["combat_updated"]
  }
}
```

**3. 心跳**

```json
{
  "type": "ping",
  "data": {
    "timestamp": "2025-02-03T10:00:00Z"
  }
}
```

#### 5.2.2 服务器 → 客户端消息

**1. 新消息通知**

```json
{
  "type": "new_message",
  "data": {
    "id": "msg-uuid-xxx",
    "session_id": "session-uuid-xxx",
    "role": "assistant",
    "content": "你冲向哥布林...",
    "created_at": "2025-02-03T10:05:30Z"
  }
}
```

**2. 游戏状态变更**

```json
{
  "type": "state_changed",
  "data": {
    "session_id": "session-uuid-xxx",
    "changes": {
      "location": "幽暗森林",
      "game_time": "第3天 10:30",
      "combat": "active"
    },
    "timestamp": "2025-02-03T10:05:00Z"
  }
}
```

**3. 战斗事件**

```json
{
  "type": "combat_updated",
  "data": {
    "session_id": "session-uuid-xxx",
    "combat_id": "combat-uuid-xxx",
    "action": "attack",
    "attacker": "player-123",
    "target": "goblin-1",
    "result": {
      "hit": true,
      "damage": 8
    },
    "timestamp": "2025-02-03T10:05:00Z"
  }
}
```

**4. 骰子投掷结果**

```json
{
  "type": "dice_rolled",
  "data": {
    "session_id": "session-uuid-xxx",
    "player_id": "player-123",
    "roll_type": "attack",
    "formula": "1d20+5",
    "result": 18,
    "critical": false,
    "timestamp": "2025-02-03T10:05:00Z"
  }
}
```

**5. 心跳响应**

```json
{
  "type": "pong",
  "data": {
    "timestamp": "2025-02-03T10:00:00Z"
  }
}
```

**6. 错误通知**

```json
{
  "type": "error",
  "data": {
    "code": "MCP_SERVER_ERROR",
    "message": "MCP Server 连接失败",
    "timestamp": "2025-02-03T10:00:00Z"
  }
}
```

#### 5.2.3 WebSocket 连接生命周期

```
1. 建立连接
   Client → Server: WebSocket 握手请求
   Server → Client: WebSocket 握手响应
   Server → Client: {type: "connected", data: {session_id, ...}}

2. 保持连接
   Client → Server: {type: "ping"} (每30秒)
   Server → Client: {type: "pong"}

3. 接收实时事件
   Server → Client: {type: "state_changed", ...}
   Server → Client: {type: "combat_updated", ...}

4. 关闭连接
   Client → Server: Close 帧
   Server → Client: Close 帧
```

### 5.3 API 错误码定义

```
┌─────────────────────────────────────────────────────────────────┐
│                      错误码定义                                  │
└─────────────────────────────────────────────────────────────────┘

HTTP 状态码:
  200 OK              - 请求成功
  201 Created         - 资源创建成功
  204 No Content      - 删除成功
  400 Bad Request     - 请求参数错误
  404 Not Found       - 资源不存在
  500 Internal Server Error - 服务器内部错误
  503 Service Unavailable - 服务不可用

业务错误码:

错误码格式: {SERVICE}_{ERROR_TYPE}_{SPECIFIC_ERROR}

示例:
  SESSION_NOT_FOUND       - 会话不存在
  MESSAGE_INVALID_CONTENT - 消息内容无效
  LLM_RATE_LIMIT_EXCEEDED - LLM 调用频率超限
  MCP_SERVER_UNAVAILABLE  - MCP Server 不可用
  REDIS_CONNECTION_FAILED - Redis 连接失败

错误响应格式:
{
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "指定的会话不存在",
    "details": {
      "session_id": "invalid-uuid"
    },
    "timestamp": "2025-02-03T10:00:00Z",
    "request_id": "req-uuid-xxx"
  }
}
```

---

## 六、数据库存储设计

### 6.1 Redis 数据结构

```
┌─────────────────────────────────────────────────────────────────┐
│                   Redis 数据结构设计                             │
└─────────────────────────────────────────────────────────────────┘

1. 会话元数据 (Hash)
   Key: session:{uuid}
   Type: Hash
   TTL: 永久
   Fields:
     • id: UUID
     • name: string (会话名称)
     • creator_id: string (创建者ID)
     • mcp_server_url: string (MCP Server URL)
     • websocket_key: string (WebSocket密钥)
     • max_players: integer (最大玩家数)
     • settings: JSON string (会话配置)
     • created_at: RFC3339 timestamp
     • updated_at: RFC3339 timestamp
     • status: string (active | archived)

   示例:
   HGETALL session:123e4567-e89b-12d3-a456-426614174000
   {
     "id": "123e4567-e89b-12d3-a456-426614174000",
     "name": "我的第一个战役",
     "creator_id": "user-123",
     "mcp_server_url": "http://localhost:9000",
     "websocket_key": "ws-key-abc123",
     "max_players": "5",
     "settings": "{\"ruleset\":\"dnd5e\"}",
     "created_at": "2025-02-03T10:00:00Z",
     "updated_at": "2025-02-03T10:00:00Z",
     "status": "active"
   }

2. 会话索引 (Set)
   Key: sessions:all
   Type: Set
   TTL: 永久
   Members: 所有活跃的 session UUID

   用途: 快速列出所有会话

   示例:
   SMEMBERS sessions:all
   ["123e4567-e89b-12d3-a456-426614174000", "789e0123-e89b-12d3-a456-426614174001"]

3. 对话消息 (Sorted Set)
   Key: msg:{session_uuid}
   Type: ZSET (有序集合)
   TTL: 永久
   Score: Unix timestamp (毫秒)
   Member: JSON序列化的消息对象

   消息结构:
   {
     "id": "msg-uuid-xxx",
     "session_id": "session-uuid-xxx",
     "role": "user | assistant | system | tool",
     "content": "消息内容",
     "tool_calls": [              // 可选，工具调用列表
       {
         "id": "call-xxx",
         "name": "resolve_attack",
         "arguments": {...}
       }
     ],
     "player_id": "player-123",   // 可选，玩家ID
     "created_at": "2025-02-03T10:00:00Z"
   }

   操作:
   • ZADD msg:{session_id} {timestamp} {json}  - 添加消息
   • ZREVRANGE msg:{session_id} 0 49          - 获取最近50条
   • ZCARD msg:{session_id}                   - 消息总数
   • ZRANGEBYSCORE msg:{session_id} {min} {max} - 按时间范围查询

   示例:
   ZADD msg:123e4567-e89b-12d3-a456-426614174000 1706926800000 '{"id":"msg-1","role":"user","content":"攻击哥布林"}'

4. 系统元数据 (String)
   Key: system:*
   Type: String
   TTL: 永久

   a. system:status
      值: starting | ready | persistence_in_progress | stopping

   b. system:last_persistence
      值: Unix timestamp (秒)

   c. system:startup_time
      值: Unix timestamp (秒)

   d. system:version
      值: "1.0.0"

   示例:
   SET system:status "ready"
   GET system:status
   > "ready"

5. WebSocket 连接索引 (Hash)
   Key: ws:connections:{session_uuid}
   Type: Hash
   TTL: 动态（连接断开时删除）
   Fields:
     • {connection_id}: JSON string (连接信息)

   用途: 跟踪活跃的 WebSocket 连接

   示例:
   HGETALL ws:connections:123e4567-e89b-12d3-a456-426614174000
   {
     "conn-1": "{\"player_id\":\"player-123\",\"connected_at\":\"2025-02-03T10:00:00Z\"}",
     "conn-2": "{\"player_id\":\"player-456\",\"connected_at\":\"2025-02-03T10:01:00Z\"}"
   }

6. 持久化标记 (Set)
   Key: persistence:dirty
   Type: Set
   TTL: 永久
   Members: 需要持久化的 session UUID

   用途: 标记自上次持久化以来有变更的会话

   示例:
   SADD persistence:dirty 123e4567-e89b-12d3-a456-426614174000

7. 速率限制 (String)
   Key: ratelimit:{user_id}:{endpoint}
   Type: String
   TTL: 动态（如60秒）
   Value: 请求计数

   用途: API 速率限制

   示例:
   SET ratelimit:user-123:chat 10 EX 60
   INCR ratelimit:user-123:chat
```

### 6.2 PostgreSQL 数据结构

```
┌─────────────────────────────────────────────────────────────────┐
│                 PostgreSQL 数据结构设计                          │
└─────────────────────────────────────────────────────────────────┘

1. client_sessions (会话元数据表)

   Columns:
     • id: UUID (PK)                           - 会话ID
     • created_at: TIMESTAMP (NOT NULL)         - 创建时间
     • updated_at: TIMESTAMP (NOT NULL)         - 更新时间
     • deleted_at: TIMESTAMP (NULLABLE)         - 删除时间（软删除）
     • name: VARCHAR(255) (NOT NULL)            - 会话名称
     • creator_id: VARCHAR(255) (NOT NULL)      - 创建者ID
     • mcp_server_url: VARCHAR(512) (NOT NULL)  - MCP Server URL
     • websocket_key: VARCHAR(255) (NOT NULL)   - WebSocket密钥
     • max_players: INTEGER (NULLABLE)          - 最大玩家数
     • settings: JSONB (NULLABLE)               - 会话配置
     • status: VARCHAR(20) (NOT NULL)           - 状态

   Indexes:
     • idx_sessions_created_at (created_at DESC)
     • idx_sessions_updated_at (updated_at DESC)
     • idx_sessions_deleted_at (deleted_at)
     • idx_sessions_creator_id (creator_id)
     • idx_sessions_status (status)

   Constraints:
     • PRIMARY KEY (id)
     • CHECK (deleted_at IS NULL OR updated_at >= deleted_at)

   SQL:
   CREATE TABLE client_sessions (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
     deleted_at TIMESTAMP,
     name VARCHAR(255) NOT NULL,
     creator_id VARCHAR(255) NOT NULL,
     mcp_server_url VARCHAR(512) NOT NULL,
     websocket_key VARCHAR(255) NOT NULL,
     max_players INTEGER,
     settings JSONB,
     status VARCHAR(20) NOT NULL DEFAULT 'active',
     CHECK (deleted_at IS NULL OR updated_at >= deleted_at)
   );

   CREATE INDEX idx_sessions_created_at ON client_sessions(created_at DESC);
   CREATE INDEX idx_sessions_updated_at ON client_sessions(updated_at DESC);
   CREATE INDEX idx_sessions_deleted_at ON client_sessions(deleted_at);
   CREATE INDEX idx_sessions_creator_id ON client_sessions(creator_id);
   CREATE INDEX idx_sessions_status ON client_sessions(status);

2. client_messages (对话消息表)

   Columns:
     • id: UUID (PK)                          - 消息ID
     • session_id: UUID (FK)                  - 会话ID
     • created_at: TIMESTAMP (NOT NULL)        - 创建时间
     • role: VARCHAR(20) (NOT NULL)           - 角色
     • content: TEXT (NULLABLE)               - 消息内容
     • tool_calls: JSONB (NULLABLE)           - 工具调用
     • player_id: VARCHAR(255) (NULLABLE)     - 玩家ID

   Indexes:
     • idx_messages_session_time (session_id, created_at DESC)
     • idx_messages_role (role)
     • idx_messages_player (player_id)

   Constraints:
     • PRIMARY KEY (id)
     • FOREIGN KEY (session_id)
       REFERENCES client_sessions(id)
       ON DELETE CASCADE
     • UNIQUE (id) - 防止重复插入
     • CHECK (role IN ('system', 'user', 'assistant', 'tool'))

   SQL:
   CREATE TABLE client_messages (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     session_id UUID NOT NULL,
     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
     role VARCHAR(20) NOT NULL,
     content TEXT,
     tool_calls JSONB,
     player_id VARCHAR(255),
     CONSTRAINT fk_session
       FOREIGN KEY (session_id)
       REFERENCES client_sessions(id)
       ON DELETE CASCADE,
     CONSTRAINT valid_role
       CHECK (role IN ('system', 'user', 'assistant', 'tool'))
   );

   CREATE INDEX idx_messages_session_time
     ON client_messages(session_id, created_at DESC);
   CREATE INDEX idx_messages_role
     ON client_messages(role);
   CREATE INDEX idx_messages_player
     ON client_messages(player_id);

3. persistence_snapshots (持久化快照表)

   Columns:
     • id: UUID (PK)                          - 快照ID
     • created_at: TIMESTAMP (NOT NULL)        - 创建时间
     • session_count: INTEGER (NOT NULL)       - 会话数量
     • message_count: INTEGER (NOT NULL)       - 消息数量
     • duration_ms: INTEGER (NOT NULL)         - 持久化耗时
     • status: VARCHAR(20) (NOT NULL)         - 状态

   Indexes:
     • idx_snapshots_created_at (created_at DESC)

   SQL:
   CREATE TABLE persistence_snapshots (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
     session_count INTEGER NOT NULL,
     message_count INTEGER NOT NULL,
     duration_ms INTEGER NOT NULL,
     status VARCHAR(20) NOT NULL
   );

   CREATE INDEX idx_snapshots_created_at
     ON persistence_snapshots(created_at DESC);

4. system_events (系统事件表)

   Columns:
     • id: UUID (PK)                          - 事件ID
     • created_at: TIMESTAMP (NOT NULL)        - 创建时间
     • event_type: VARCHAR(50) (NOT NULL)      - 事件类型
     • level: VARCHAR(20) (NOT NULL)          - 日志级别
     • message: TEXT (NOT NULL)               - 事件消息
     • data: JSONB (NULLABLE)                 - 附加数据

   Indexes:
     • idx_events_created_at (created_at DESC)
     • idx_events_type (event_type)
     • idx_events_level (level)

   SQL:
   CREATE TABLE system_events (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
     event_type VARCHAR(50) NOT NULL,
     level VARCHAR(20) NOT NULL,
     message TEXT NOT NULL,
     data JSONB
   );

   CREATE INDEX idx_events_created_at
     ON system_events(created_at DESC);
   CREATE INDEX idx_events_type
     ON system_events(event_type);
   CREATE INDEX idx_events_level
     ON system_events(level);

5. 数据库迁移

   迁移文件: /migrations/
   格式: {version}_{description}.sql

   示例:
   /migrations/
     001_initial_schema.sql
     002_add_persistence_snapshots.sql
     003_add_system_events.sql

   迁移工具: golang-migrate
   执行:
     migrate -path migrations -database "postgres://..." up
```

### 6.3 数据一致性保证

```
┌─────────────────────────────────────────────────────────────────┐
│                  数据一致性保证策略                              │
└─────────────────────────────────────────────────────────────────┘

1. Redis 作为主存储
   • 所有读写操作先访问 Redis
   • 保证 <1ms 响应时间
   • 使用 Pipeline 批量操作

2. PostgreSQL 作为备份
   • 定期从 Redis 持久化到 PG
   • 使用 UPSERT 避免重复
   • 使用事务保证原子性

3. 持久化策略
   a. 定期持久化（每30秒）
      • 遍历所有会话
      • 批量写入 PostgreSQL
      • 记录持久化快照

   b. 关键操作立即持久化
      • 会话创建
      • 会话删除
      • 立即标记为 dirty

4. 数据恢复
   a. 系统启动时
      • 从 PostgreSQL 加载数据
      • 写入 Redis
      • 验证数据完整性

   b. 崩溃恢复
      • 加载上一次持久化点
      • 可能丢失 30 秒数据

5. 数据校验
   a. 定期校验
      • 对比 Redis 和 PG 数据量
      • 检测不一致情况

   b. 修复策略
      • 以 Redis 为准（主存储）
      • 重新持久化到 PG

6. 幂等性保证
   • 所有写操作支持幂等
   • 使用 UUID 唯一标识
   • ON CONFLICT DO NOTHING
```

---

## 七、性能优化策略

### 7.1 Redis 优化

```
1. 使用 Pipeline 批量操作
   • 减少网络往返次数
   • 一次性发送多个命令

2. 合理设置 TTL
   • 临时数据设置过期时间
   • 避免内存无限增长

3. 使用连接池
   • 复用连接
   • 减少连接建立开销

4. 压缩存储
   • JSON 数据压缩
   • 减少 Redis 内存占用
```

### 7.2 PostgreSQL 优化

```
1. 批量插入
   • 使用 COPY 命令
   • 或使用 VALUES (), (), ()

2. 索引优化
   • 为查询字段建立索引
   • 定期 ANALYZE 更新统计信息

3. 连接池
   • 使用 pgx 连接池
   • 复用数据库连接

4. 分区表
   • 按时间分区 client_messages
   • 提升查询性能
```

### 7.3 应用层优化

```
1. 并行处理
   • 使用 goroutine 并行处理会话
   • 加快持久化速度

2. 缓存策略
   • 缓存会话元数据
   • 缓存 MCP Server 状态

3. 流式响应
   • 支持 SSE 流式返回
   • 提升 LLM 响应体验

4. 限流保护
   • API 速率限制
   • 防止资源耗尽
```

---

## 八、监控和日志

### 8.1 监控指标

```
1. 系统指标
   • CPU 使用率
   • 内存使用量
   • Goroutine 数量
   • GC 暂停时间

2. 业务指标
   • 请求总数
   • 响应时间 (P50, P95, P99)
   • 错误率
   • 活跃会话数

3. 数据库指标
   • Redis 命令耗时
   • PostgreSQL 查询耗时
   • 连接池使用率

4. 外部服务指标
   • LLM 调用次数和耗时
   • MCP Server 调用次数和耗时
```

### 8.2 日志规范

```
1. 日志级别
   • Debug: 详细调试信息
   • Info: 一般信息
   • Warn: 警告信息
   • Error: 错误信息
   • Fatal: 致命错误

2. 日志格式
   {
     "level": "info",
     "time": "2025-02-03T10:00:00Z",
     "msg": "会话创建成功",
     "session_id": "xxx",
     "duration_ms": 50
   }

3. 结构化日志
   • 使用 JSON 格式
   • 包含请求 ID
   • 包含关键业务字段

4. 日志聚合
   • 集中收集到日志系统
   • 支持检索和分析
```

---

## 九、部署配置

### 9.1 环境变量

```
# 应用配置
APP_NAME=dnd-client
APP_VERSION=1.0.0
APP_ENV=production

# HTTP 服务
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
HTTP_READ_TIMEOUT=30s
HTTP_WRITE_TIMEOUT=30s

# WebSocket
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_PING_PERIOD=30s
WS_PONG_WAIT=10s

# Redis
REDIS_HOST=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5

# PostgreSQL
POSTGRES_HOST=localhost:5432
POSTGRES_USER=dnd
POSTGRES_PASSWORD=password
POSTGRES_DBNAME=dnd_client
POSTGRES_SSLMODE=disable
POSTGRES_POOL_SIZE=5
POSTGRES_MAX_CONN_LIFETIME=1h
POSTGRES_MAX_CONN_IDLETIME=30m

# LLM
LLM_PROVIDER=openai
LLM_API_KEY=sk-xxx
LLM_MODEL=gpt-4
LLM_MAX_TOKENS=4096
LLM_TEMPERATURE=0.7
LLM_TIMEOUT=30s

# 持久化触发器配置
PERSISTENCE_TYPE=time              # 触发器类型: time | message | manual | composite
PERSISTENCE_INTERVAL=30s           # 时间触发器间隔
PERSISTENCE_MESSAGE_THRESHOLD=100  # 消息量触发器阈值（预留）
PERSISTENCE_BATCH_SIZE=100         # 批量处理大小
PERSISTENCE_TIMEOUT=10s            # 持久化超时时间

# 日志
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

### 9.2 Docker Compose

```yaml
version: '3.8'

services:
  dnd-client:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - REDIS_HOST=redis:6379
      - POSTGRES_HOST=postgres:5432
    depends_on:
      - redis
      - postgres
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=dnd
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=dnd_client
    volumes:
      - postgres-data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  redis-data:
  postgres-data:
```

---

## 十、测试策略

### 10.1 单元测试

```
1. Repository 层测试
   • Mock Redis 和 PostgreSQL
   • 测试数据访问逻辑

2. Service 层测试
   • Mock Repository
   • 测试业务逻辑

3. Handler 层测试
   • 使用 httptest
   • 测试 HTTP 接口
```

### 10.2 集成测试

```
1. 数据库集成测试
   • 使用 testcontainers
   • 启动真实的 Redis 和 PostgreSQL

2. MCP Server 集成测试
   • Mock MCP Server
   • 测试协议交互

3. 端到端测试
   • 测试完整流程
   • 从 HTTP 请求到数据库
```

### 10.3 性能测试

```
1. 基准测试
   • 使用 Go benchmark
   • 测试关键函数性能

2. 压力测试
   • 使用 vegeta 或 k6
   • 模拟并发请求

3. 负载测试
   • 测试系统容量
   • 找出性能瓶颈
```

---

## 总结

本文档详细描述了 DND MCP Client 的设计，包括:

1. **执行流程**: 7 个核心流程的详细步骤（新增持久化触发器设计）
2. **组件交互**: 清晰的架构层次和通信方式
3. **API 设计**: 简洁的 REST API 和 WebSocket API 设计
4. **数据存储**: Redis 和 PostgreSQL 数据结构设计

核心特点:

- ✅ 轻量级有状态设计
- ✅ Redis 主存储 (<1ms 响应)
- ✅ PostgreSQL 定期备份
- ✅ **灵活的持久化触发机制**（支持时间、消息量、手动、组合触发）
- ✅ **简洁的 API 设计**（无分页，适合中小规模场景）
- ✅ 实时 WebSocket 推送
- ✅ 优雅的故障处理
- ✅ 易于扩展的架构设计

适用场景:

- 单实例部署
- 4-5 个并发会话
- 可接受 30 秒数据丢失
- 追求快速响应时间
- API 调用量中等（不需要复杂分页）

设计亮点:

1. **持久化触发器设计**
   
   - 接口化设计，易于扩展
   - 支持多种触发策略（时间、消息量、手动、组合）
   - 当前实现：时间触发器（30秒）
   - 预留扩展：消息量触发器、手动触发器
   - 配置化管理，灵活切换触发策略

2. **API 设计简化**
   
   - 删除复杂的分页、排序参数
   - 直接返回数组，结构简单
   - 适合中小规模场景（< 100 个会话）
   - 提升开发效率和易用性

3. **可扩展性**
   
   - 触发器接口预留扩展点
   - 新增触发器无需修改核心逻辑
   - 支持触发器组合，灵活配置

版本说明:

- v1.1: 添加持久化触发器设计，简化 API 参数
- v1.0: 初始版本
