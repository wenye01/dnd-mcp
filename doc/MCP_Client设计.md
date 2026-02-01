# MCP Client 详细设计

## 目录

- [一、整体架构](#一整体架构)
- [二、核心组件设计](#二核心组件设计)
- [三、Session管理机制](#三session管理机制)
- [四、数据结构定义](#四数据结构定义)
- [五、接口定义](#五接口定义)
- [六、数据流设计](#六数据流设计)
- [七、错误处理策略](#七错误处理策略)
- [八、配置管理](#八配置管理)
- [九、部署方案](#九部署方案)

---

## 一、整体架构

### 1.1 MCP Client 组件架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          【外部系统】                                         │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐              │
│  │   前端 UI    │      │  云端LLM     │      │  MCP Server  │              │
│  │  (Browser)   │      │  (OpenAI)    │      │  (规则引擎)   │              │
│  └──────┬───────┘      └──────┬───────┘      └──────┬───────┘              │
│         │                     │                      │                        │
│         │ HTTP/WebSocket      │ HTTPS               │ HTTP+JSON-RPC          │
│         │                     │                      │                        │
└─────────┼─────────────────────┼──────────────────────┼────────────────────────┘
          │                     │                      │
          ▼                     ▼                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MCP Client (无状态协调层)                            │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                        【API Gateway 层】                            │  │
│  │  ┌─────────────────────────────────────────────────────────────┐    │  │
│  │  │                    HTTP Router                              │    │  │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │    │  │
│  │  │  │Session Routes│  │ Chat Routes  │  │Query Routes  │      │    │  │
│  │  │  │  (CRUD)      │  │  (/chat)     │  │  (/state)    │      │    │  │
│  │  │  └──────────────┘  └──────────────┘  └──────────────┘      │    │  │
│  │  └─────────────────────────────────────────────────────────────┘    │  │
│  │                                                                      │  │
│  │  ┌─────────────────────────────────────────────────────────────┐    │  │
│  │  │                  WebSocket Manager                          │    │  │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │    │  │
│  │  │  │Connection Mgr│  │ Message Pump  │  │Broadcast     │      │    │  │
│  │  │  │  (注册/注销)  │  │  (读写泵)     │  │  (推送)       │      │    │  │
│  │  │  └──────────────┘  └──────────────┘  └──────────────┘      │    │  │
│  │  └─────────────────────────────────────────────────────────────┘    │  │
│  │                                                                      │  │
│  │  ┌─────────────────────────────────────────────────────────────┐    │  │
│  │  │                    Middleware Stack                         │    │  │
│  │  │  RequestID → Logging → Recovery → Auth → RateLimit           │    │  │
│  │  └─────────────────────────────────────────────────────────────┘    │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                     │
│                                    ▼                                     │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                       【Orchestrator 核心层】                        │  │
│  │                                                                     │  │
│  │  ┌─────────────────────────────────────────────────────────────┐   │  │
│  │  │                   主处理循环控制器                           │   │  │
│  │  │  ProcessChatMessage()                                       │   │  │
│  │  │  • 接收请求                                                 │   │  │
│  │  │  • 调用子组件                                               │   │  │
│  │  │  • 协调工作流                                               │   │  │
│  │  │  • 错误处理                                                 │   │  │
│  │  └─────────────────────────────────────────────────────────────┘   │  │
│  │                                                                     │  │
│  │  ┌────────────────┐      ┌────────────────┐                       │  │
│  │  │                │      │                │                       │  │
│  │  │Context Builder │      │  LLM Manager   │                       │  │
│  │  │┌──────────────┐│      │┌──────────────┐│                       │  │
│  │  ││Msg Loader    ││      ││LLM Client    ││                       │  │
│  │  ││State Fetcher ││      ││OpenAI Impl   ││                       │  │
│  │  ││System Prompt ││      ││Anthropic Impl││                       │  │
│  │  ││Token Budget  ││      ││Retry Logic   ││                       │  │
│  │  │└──────────────┘│      │└──────────────┘│                       │  │
│  │  └────────────────┘      └────────────────┘                       │  │
│  │                                                                  │  │
│  │  ┌────────────────┐      ┌────────────────┐                      │  │
│  │  │                │      │                │                      │  │
│  │  │Tool Coordinator│      │Response Gen    │                      │  │
│  │  │┌──────────────┐│      │┌──────────────┐│                      │  │
│  │  ││Format Convert││      ││Text Extract  ││                      │  │
│  │  ││MCP Caller    ││      ││State Extract ││                      │  │
│  │  ││Result Parser ││      ││Turn Info     ││                      │  │
│  │  ││Event Collect ││      ││Roll Check    ││                      │  │
│  │  │└──────────────┘│      │└──────────────┘│                      │  │
│  │  └────────────────┘      └────────────────┘                      │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                      【Data Access 层】                              │  │
│  │                                                                     │  │
│  │  ┌──────────────────────┐    ┌──────────────────────┐              │  │
│  │  │   Message Store      │    │    MCP Client        │              │  │
│  │  │ ┌──────────────────┐ │    │ ┌──────────────────┐ │              │  │
│  │  │ │ Database Adapter │ │    │ │ HTTP Client      │ │              │  │
│  │  │ │ Cache Layer      │ │    │ │ Protocol Codec   │ │              │  │
│  │  │ │ Query Builder    │ │    │ │ Retry Wrapper    │ │              │  │
│  │  │ └──────────────────┘ │    │ └──────────────────┘ │              │  │
│  │  └──────────────────────┘    └──────────────────────┘              │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                     │
│                                    ▼                                     │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    【Event Dispatch 层】                             │  │
│  │  ┌─────────────────────────────────────────────────────────────┐    │  │
│  │  │                    Event Dispatcher                         │    │  │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │    │  │
│  │  │  │Event Queue   │  │Filter        │  │Broadcaster   │      │    │  │
│  │  │  │  (缓冲区)     │  │  (过滤)       │  │  (广播)       │      │    │  │
│  │  │  └──────────────┘  └──────────────┘  └──────────────┘      │    │  │
│  │  └─────────────────────────────────────────────────────────────┘    │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                    │                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼ 推送事件
                          ┌────────────────┐
                          │   前端 UI      │
                          │  (WebSocket)   │
                          └────────────────┘
```

### 1.2 组件层次结构

```
MCP Client
│
├─【API Gateway 层】
│  ├─ HTTP Router
│  │  ├─ SessionHandler (会话CRUD)
│  │  ├─ ChatHandler (聊天核心)
│  │  ├─ QueryHandler (状态查询)
│  │  └─ SnapshotHandler (快照管理)
│  │
│  ├─ WebSocket Manager
│  │  ├─ ConnectionManager (连接管理)
│  │  ├─ ReadPump (读取泵)
│  │  ├─ WritePump (写入泵)
│  │  └─ Heartbeat (心跳保活)
│  │
│  └─ Middleware Stack
│     ├─ RequestIDMiddleware
│     ├─ LoggingMiddleware
│     ├─ RecoveryMiddleware
│     ├─ AuthMiddleware (可选)
│     └─ RateLimitMiddleware (可选)
│
├─【Orchestrator 核心层】
│  ├─ 主控制器
│  │  └─ ProcessChatMessage (主处理循环)
│  │
│  ├─ ContextBuilder (上下文构建器)
│  │  ├─ MessageLoader (消息加载器)
│  │  ├─ StateFetcher (状态获取器)
│  │  ├─ SystemPromptBuilder (系统提示构建器)
│  │  └─ TokenBudgetManager (Token预算管理器)
│  │
│  ├─ LLMManager (LLM交互管理器)
│  │  ├─ LLMClient (接口)
│  │  │  ├─ OpenAIClient
│  │  │  └─ AnthropicClient
│  │  ├─ RetryStrategy (重试策略)
│  │  └─ ErrorHandler (错误处理器)
│  │
│  ├─ ToolCoordinator (工具协调器)
│  │  ├─ FormatConverter (格式转换器)
│  │  ├─ MCPInvoker (MCP调用器)
│  │  ├─ ResultParser (结果解析器)
│  │  └─ EventCollector (事件收集器)
│  │
│  └─ ResponseGenerator (响应生成器)
│     ├─ TextExtractor (文本提取器)
│     ├─ StateChangeExtractor (状态变更提取器)
│     ├─ TurnInfoExtractor (回合信息提取器)
│     └─ RollRequirementChecker (投骰需求检查器)
│
├─【Data Access 层】
│  ├─ MessageStore (消息存储)
│  │  ├─ DatabaseAdapter (数据库适配器)
│  │  ├─ CacheLayer (缓存层)
│  │  └─ QueryBuilder (查询构建器)
│  │
│  └─ MCPClient (MCP客户端)
│     ├─ HTTPClient (HTTP客户端)
│     ├─ ProtocolCodec (协议编解码器)
│     └─ RetryWrapper (重试包装器)
│
└─【Event Dispatch 层】
   └─ EventDispatcher (事件分发器)
      ├─ EventQueue (事件队列)
      ├─ EventFilter (事件过滤器)
      └─ WebSocketBroadcaster (WebSocket广播器)
```

### 1.3 核心设计原则

#### 1.3.1 无状态设计

- **不持有业务状态**：所有会话状态由MCP Server管理
- **每次请求获取最新状态**：不缓存会话状态
- **可水平扩展**：支持多实例部署

#### 1.3.2 职责单一

- **只负责协调**：协调LLM和MCP Server交互
- **协议转换**：LLM格式 ↔ MCP Protocol格式
- **不实现业务逻辑**：由MCP Server处理
- **不做AI决策**：由LLM处理

#### 1.3.3 异步非阻塞

- **WebSocket事件推送异步处理**
- **LLM调用超时控制**
- **工具执行并发处理**

#### 1.3.4 容错设计

- **LLM调用失败重试**（最多3次）
- **MCP调用失败返回错误给LLM**
- **最大迭代次数保护**（10次）
- **panic恢复机制**

---

## 二、核心组件设计

### 2.1 API Gateway 层

#### 2.1.1 HTTP Router

**职责**：

- 接收前端HTTP请求
- 路由分发到对应的Handler
- 提取请求参数和session_id

**主要Handler**：

| Handler         | 路径                                  | 功能         |
| --------------- | ----------------------------------- | ---------- |
| SessionHandler  | `POST /api/sessions`                | 创建新会话      |
| SessionHandler  | `GET /api/sessions`                 | 列出会话       |
| SessionHandler  | `GET /api/sessions/{id}`            | 获取会话详情     |
| SessionHandler  | `DELETE /api/sessions/{id}`         | 删除会话       |
| **ChatHandler** | `POST /api/sessions/{id}/chat`      | **聊天核心接口** |
| QueryHandler    | `GET /api/sessions/{id}/state`      | 获取会话状态     |
| QueryHandler    | `GET /api/sessions/{id}/messages`   | 获取消息历史     |
| QueryHandler    | `GET /api/sessions/{id}/characters` | 获取角色列表     |
| QueryHandler    | `GET /api/sessions/{id}/combat`     | 获取战斗状态     |

#### 2.1.2 WebSocket Manager

**职责**：

- 管理WebSocket连接生命周期
- 实时推送事件到前端
- 维护连接状态

**核心功能**：

```
连接管理
├─ ConnectionManager: 注册/注销连接
├─ ReadPump: 读取前端消息（心跳等）
├─ WritePump: 写入事件到前端
└─ Heartbeat: Ping/Pong保活

事件推送
├─ Broadcast: 广播事件到指定session的所有连接
├─ EventFilter: 过滤不需要推送的事件
└─ QueueBuffer: 事件队列缓冲
```

#### 2.1.3 Middleware Stack

```
请求处理流程：
Request → RequestID → Logging → Recovery → Auth → RateLimit → Handler
              ↓          ↓         ↓        ↓        ↓         ↓
           生成ID    记录日志    panic恢复  认证    限流    业务处理
```

### 2.2 Orchestrator 核心层

#### 2.2.1 主处理循环控制器

**ProcessChatMessage 流程**：

```
1. 接收ChatRequest
   ├─ session_id: 必填
   ├─ message: 玩家输入
   └─ player_id: 可选

2. ContextBuilder 构建对话上下文
   ├─ MessageLoader: 加载最近50条消息
   ├─ StateFetcher: 从MCP Server获取状态摘要
   ├─ SystemPromptBuilder: 生成动态system prompt
   └─ TokenBudgetManager: Token预算管理

3. LLMManager 调用LLM 【循环开始】
   ├─ 发送对话上下文
   ├─ 解析tool_calls
   └─ 返回LLMResponse

4. 判断: tool_calls是否为空？
   ├─ 是 → 跳到步骤7
   └─ 否 → 继续步骤5

5. ToolCoordinator 执行工具调用
   ├─ FormatConverter: LLM格式 → MCP格式
   ├─ MCPInvoker: 调用MCP Server
   ├─ ResultParser: 解析结果和事件
   └─ EventCollector: 收集事件

6. EventDispatcher 推送事件到前端
   └─ WebSocketManager 广播事件

7. 判断: 是否达到最大迭代次数(10)？
   ├─ 是 → 返回错误
   └─ 否 → 回到步骤3（将tool results作为tool messages发送给LLM）

8. ResponseGenerator 生成最终响应
   ├─ TextExtractor: 提取LLM文本
   ├─ StateChangeExtractor: 提取状态变更
   ├─ TurnInfoExtractor: 提取回合信息
   └─ RollRequirementChecker: 检查是否需要投骰

9. 返回ChatResponse
```

#### 2.2.2 ContextBuilder (上下文构建器)

**职责**：

- 加载历史消息
- 获取当前状态摘要
- 生成动态system message
- 管理token预算

**System Prompt 包含内容**：

```
1. 角色定义（D&D DM）
2. 核心规则摘要
   - 骰子系统
   - 战斗规则
   - 优势/劣势
3. 当前场景信息
   - 地点
   - 游戏时间
   - 战斗状态
   - 在场角色
4. 可用工具列表
   - 工具名称
   - 参数说明
   - 返回值说明
5. 工作流程指引
```

#### 2.2.3 LLMManager (LLM交互管理器)

**支持的LLM Provider**：

- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude 3 Opus/Sonnet)

**重试策略**：

```
失败判断：
- 网络超时
- HTTP 429 (Rate Limit)
- HTTP 5xx (服务器错误)

重试逻辑：
- 最多重试3次
- 每次重试延迟递增（1s, 2s, 4s）
- 不可重试错误直接返回（如400参数错误）
```

#### 2.2.4 ToolCoordinator (工具协调器)

**格式转换**：

| LLM格式                                    | MCP格式                |
| ---------------------------------------- | -------------------- |
| `tool_call.function.name`                | `tool_name`          |
| `tool_call.function.arguments` (JSON字符串) | `arguments` (JSON对象) |
| `tool_call.id`                           | (不传递)                |

**执行流程**：

```
1. 接收LLM的tool_calls数组
2. 逐个转换格式
3. 并发调用MCP Server
4. 收集所有results和events
5. 返回ToolExecutionResult
```

#### 2.2.5 ResponseGenerator (响应生成器)

**提取内容**：

- **response**: LLM生成的叙述文本
- **state_changes**: 从tool results提取的状态变更摘要
- **turn_info**: 战斗回合信息（如果在战斗中）
- **requires_roll**: 是否需要玩家投骰（启发式判断）
- **usage**: Token使用统计

### 2.3 Data Access 层

#### 2.3.1 MessageStore (消息存储)

**存储结构**：

```
Message {
  id: UUID
  session_id: UUID  (索引)
  role: "system" | "user" | "assistant" | "tool"
  content: string
  tool_calls: array (可选)
  tool_call_id: string (可选，tool消息专用)
  created_at: timestamp
}
```

**查询接口**：

- `GetRecentMessages(session_id, limit)`: 获取最近N条消息
- `SaveMessage(session_id, message)`: 保存消息

#### 2.3.2 MCPClient (MCP客户端)

**调用接口**：

- `GetSessionState(session_id)`: 获取会话状态
- `CallTool(session_id, tool_name, arguments)`: 调用工具
- `CreateSessionState(session_id)`: 创建会话状态
- `DeleteSessionState(session_id)`: 删除会话状态

**协议格式**：HTTP + JSON-RPC 2.0

### 2.4 Event Dispatch 层

#### 2.4.1 EventDispatcher

**事件类型**：

| 分类        | 事件类型                      | 说明   |
| --------- | ------------------------- | ---- |
| dice      | `dice.rolled`             | 骰子投掷 |
| combat    | `combat.started`          | 战斗开始 |
| combat    | `combat.attack_resolved`  | 攻击结算 |
| combat    | `combat.character_downed` | 角色倒地 |
| character | `character.hp_changed`    | HP变化 |
| character | `character.moved`         | 角色移动 |
| map       | `map.door_opened`         | 门被打开 |

**推送流程**：

```
MCP Server → EventDispatcher → WebSocketManager → 前端UI
```

---

## 三、Session管理机制

### 3.1 Session 隔离机制

```
┌─────────────────────────────────────────────────────────────┐
│                  Session 隔离层次                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  1. API Gateway 层                                            │
│     ├─ 从请求中提取 session_id                               │
│     └─ 所有操作必须携带 session_id                           │
│                                                               │
│  2. Orchestrator 层                                           │
│     ├─ ContextBuilder: 按 session_id 加载消息历史            │
│     ├─ ToolCoordinator: 所有MCP调用携带 session_id            │
│     └─ EventDispatcher: 按 session_id 推送事件                │
│                                                               │
│  3. Data Access 层                                            │
│     ├─ MessageStore: 按 session_id 存储消息                  │
│     └─ MCPClient: 所有请求携带 session_id                     │
│                                                               │
│  4. MCP Server 层 ⭐ (真正的状态隔离)                         │
│     ├─ SessionManager: session_id 级别的分布式锁             │
│     ├─ StateStorage: 按 session_id 隔离存储                   │
│     └─ EventPublisher: 按 session_id 推送事件                 │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Session 数据结构

#### 3.2.1 Session 元数据

| 字段             | 类型        | 说明                          |
| -------------- | --------- | --------------------------- |
| id             | UUID      | 全局唯一标识                      |
| name           | string    | 战役名称（如"失落的地牢"）              |
| creator_id     | UUID      | 创建者ID                       |
| dm_user_id     | UUID      | DM的用户ID                     |
| status         | enum      | active / archived / deleted |
| max_players    | int       | 最大玩家数                       |
| players        | array     | 玩家列表                        |
| created_at     | timestamp | 创建时间                        |
| updated_at     | timestamp | 更新时间                        |
| last_access_at | timestamp | 最后访问时间                      |

#### 3.2.2 Session 状态

**状态流转**：

```
创建 → active → archived → deleted
  ↑        ↓           ↓
  └────────┴───────────┘
    (可以恢复)
```

**状态说明**：

- **active**: 进行中，可以正常交互
- **archived**: 已归档，保留状态但不可交互
- **deleted**: 已删除（软删除），可恢复，定期清理

### 3.3 Session 管理接口

| 接口                         | 方法     | 说明                |
| -------------------------- | ------ | ----------------- |
| /api/sessions              | POST   | 创建新会话             |
| /api/sessions              | GET    | 列出会话（支持user_id过滤） |
| /api/sessions/{id}         | GET    | 获取会话详情            |
| /api/sessions/{id}         | DELETE | 删除会话（软删除）         |
| /api/sessions/{id}/archive | POST   | 归档会话              |
| /api/sessions/{id}/restore | POST   | 恢复已归档的会话          |

### 3.4 多会话场景

**支持场景**：

- 一个用户可以创建多个战役会话
- 不同会话之间完全隔离
- 可以随时切换不同会话进行游戏

**示例**：

```
用户A:
  ├─ session_001: "失落的地牢" (active)
  ├─ session_002: "龙之城堡" (active)
  └─ session_003: "短篇冒险" (archived)

用户B:
  ├─ session_004: "新手村" (active)
  └─ session_005: "森林探险" (active)
```

### 3.5 Session 生命周期

```
1. 创建阶段
   前端 → POST /api/sessions
   {
     name: "失落的地牢",
     creator_id: "user_123",
     max_players: 5
   }

   ├─ 生成 UUID 作为 session_id
   ├─ 创建 Session 元数据
   ├─ 调用 MCP Server 创建游戏状态
   └─ 保存到数据库

   返回: {session_id: "uuid-xxx", name: "失落的地牢"}

2. 活跃阶段 (active)
   ┌─────────────────────────────────────────┐
   │ 用户通过 chat API 与游戏交互            │
   │ POST /api/sessions/{id}/chat            │
   │                                         │
   │ • 每次请求更新 last_access_at          │
   │ • 状态完全隔离在 MCP Server            │
   │ • 不同 session 互不影响                │
   └─────────────────────────────────────────┘

3. 归档阶段 (archived)
   POST /api/sessions/{id}/archive

   ├─ status: active → archived
   └─ MCP Server 状态保留，但不再接受新请求

4. 恢复阶段
   POST /api/sessions/{id}/restore

   ├─ status: archived → active
   └─ 可以继续交互

5. 删除阶段 (deleted - 软删除)
   DELETE /api/sessions/{id}

   ├─ status: any → deleted (软删除)
   └─ 异步删除 MCP Server 端状态

6. 物理删除（定期清理任务）
   Cron Job:
   ├─ 查询 status='deleted' 且 updated_at > 30天
   ├─ 从数据库删除记录
   └─ 调用 MCP Server 删除游戏状态
```

---

## 四、数据结构定义

### 4.1 核心数据结构

#### 4.1.1 ChatRequest (聊天请求)

| 字段           | 类型      | 必填  | 说明             |
| ------------ | ------- | --- | -------------- |
| session_id   | UUID    | ✅   | 会话ID           |
| message      | string  | ✅   | 玩家输入（1-2000字符） |
| player_id    | UUID    | ❌   | 玩家ID（可选）       |
| stream       | boolean | ❌   | 是否流式响应         |
| context_only | boolean | ❌   | 只获取上下文（不调用LLM） |

#### 4.1.2 ChatResponse (聊天响应)

| 字段            | 类型      | 说明          |
| ------------- | ------- | ----------- |
| response      | string  | LLM生成的叙述文本  |
| state_changes | object  | 状态变更摘要      |
| requires_roll | boolean | 是否需要玩家投骰    |
| turn_info     | object  | 当前回合信息（战斗中） |
| usage         | object  | Token使用统计   |

#### 4.1.3 Message (对话消息)

| 字段           | 类型        | 说明                                          |
| ------------ | --------- | ------------------------------------------- |
| role         | enum      | "system" \| "user" \| "assistant" \| "tool" |
| content      | string    | 消息内容                                        |
| tool_calls   | array     | 工具调用（assistant消息专用）                         |
| tool_call_id | string    | 工具调用ID（tool消息专用）                            |
| timestamp    | timestamp | 时间戳                                         |

#### 4.1.4 ToolCall (工具调用)

| 字段                 | 类型     | 说明          |
| ------------------ | ------ | ----------- |
| id                 | string | 工具调用ID      |
| type               | string | "function"  |
| function           | object | 函数调用        |
| function.name      | string | 工具名称        |
| function.arguments | string | 参数（JSON字符串） |

#### 4.1.5 LLMResponse (LLM响应)

| 字段            | 类型     | 说明                                 |
| ------------- | ------ | ---------------------------------- |
| content       | string | 生成文本                               |
| tool_calls    | array  | 工具调用列表                             |
| finish_reason | string | "stop" \| "tool_calls" \| "length" |
| usage         | object | Token使用统计                          |

#### 4.1.6 MCPToolCallRequest (MCP工具调用请求)

| 字段         | 类型     | 说明         |
| ---------- | ------ | ---------- |
| session_id | UUID   | 会话ID       |
| tool_name  | string | 工具名称       |
| arguments  | object | 参数对象       |
| version    | int    | 乐观锁版本号（可选） |

#### 4.1.7 MCPToolCallResponse (MCP工具调用响应)

| 字段      | 类型     | 说明       |
| ------- | ------ | -------- |
| result  | object | 工具执行结果   |
| version | int    | 新版本号     |
| events  | array  | 产生的事件    |
| error   | string | 错误信息（如有） |

#### 4.1.8 Event (事件)

| 字段         | 类型        | 说明   |
| ---------- | --------- | ---- |
| id         | UUID      | 事件ID |
| session_id | UUID      | 会话ID |
| type       | string    | 事件类型 |
| data       | object    | 事件数据 |
| timestamp  | timestamp | 时间戳  |

### 4.2 配置数据结构

#### 4.2.1 LLMConfig

| 字段          | 类型     | 说明                      |
| ----------- | ------ | ----------------------- |
| provider    | string | "openai" \| "anthropic" |
| api_key     | string | API密钥                   |
| model       | string | 模型名称                    |
| max_tokens  | int    | 单次请求最大token数            |
| temperature | float  | 温度参数（0-1）               |
| max_retries | int    | 最大重试次数                  |
| retry_delay | int    | 重试延迟（毫秒）                |
| timeout     | int    | 超时时间（秒）                 |

#### 4.2.2 MCPConfig

| 字段              | 类型     | 说明           |
| --------------- | ------ | ------------ |
| base_url        | string | MCP Server地址 |
| timeout         | int    | 超时时间（秒）      |
| max_connections | int    | 最大连接数        |

#### 4.2.3 OrchestratorConfig

| 字段                   | 类型  | 说明               |
| -------------------- | --- | ---------------- |
| max_iterations       | int | 最大工具调用循环次数（默认10） |
| context_window_size  | int | 上下文窗口大小          |
| message_history_size | int | 消息历史保留数量（默认50）   |
| token_budget         | int | Token预算（默认6000）  |

#### 4.2.4 WebSocketConfig

| 字段                | 类型  | 说明             |
| ----------------- | --- | -------------- |
| read_buffer_size  | int | 读缓冲区大小         |
| write_buffer_size | int | 写缓冲区大小         |
| ping_interval     | int | Ping间隔（秒，默认30） |
| pong_timeout      | int | Pong超时（秒，默认60） |

---

## 五、接口定义

### 5.1 HTTP API

#### 5.1.1 核心API：聊天

```
POST /api/sessions/{id}/chat

请求体:
{
  "message": "我要攻击那个哥布林",
  "player_id": "player-001"  // 可选
}

响应体:
{
  "response": "你冲向哥布林，挥剑砍中...",
  "state_changes": {
    "goblin_1": {"hp": 5, "hp_max": 13}
  },
  "requires_roll": false,
  "usage": {
    "prompt_tokens": 1200,
    "completion_tokens": 300,
    "total_tokens": 1500
  }
}
```

#### 5.1.2 会话管理API

```
创建会话:
POST /api/sessions
{
  "name": "失落的地牢",
  "creator_id": "user-123",
  "max_players": 5
}
→ {session_id: "uuid-xxx", name: "失落的地牢", ...}

列出会话:
GET /api/sessions?user_id=user-123&status=active&limit=20
→ {sessions: [...], total: 5}

获取会话详情:
GET /api/sessions/{id}
→ {id, name, status, players, ...}

删除会话:
DELETE /api/sessions/{id}
→ {"message": "Session deleted"}

归档会话:
POST /api/sessions/{id}/archive
→ {"message": "Session archived"}

恢复会话:
POST /api/sessions/{id}/restore
→ {"message": "Session restored"}
```

#### 5.1.3 查询API

```
获取会话状态:
GET /api/sessions/{id}/state
→ {
  session_id, location, game_time,
  combat: {...}, characters: [...]
}

获取消息历史:
GET /api/sessions/{id}/messages?limit=50
→ {messages: [...]}

获取角色列表:
GET /api/sessions/{id}/characters
→ {characters: [...]}

获取战斗状态:
GET /api/sessions/{id}/combat
→ {active: true, round: 3, current_turn: "..."}
```

### 5.2 WebSocket API

#### 5.2.1 连接

```
WS /api/sessions/{id}/ws

连接后立即推送:
{
  "type": "connected",
  "session_id": "uuid-xxx"
}
```

#### 5.2.2 事件推送

```
格式:
{
  "type": "event",
  "event": {
    "id": "event-uuid",
    "type": "combat.attack_resolved",
    "data": {...},
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### 5.3 MCP Protocol (Client → Server)

#### 5.3.1 工具调用

```
POST /tools/call

请求:
{
  "session_id": "uuid-xxx",
  "tool_name": "resolve_attack",
  "arguments": {
    "attacker_id": "...",
    "target_id": "...",
    "attack_type": "melee"
  },
  "version": 5  // 可选
}

响应:
{
  "result": {
    "hit": true,
    "damage": 8,
    "attack_roll": 18
  },
  "version": 6,
  "events": [...]
}
```

#### 5.3.2 获取状态

```
GET /state?session_id=uuid-xxx

响应:
{
  "session_id": "uuid-xxx",
  "location": "地下城",
  "characters": [...],
  "combat": {...}
}
```

---

## 六、数据流设计

### 6.1 聊天请求完整流程

```
1. 前端发起请求
   POST /api/sessions/sess-001/chat
   {message: "我要攻击那个哥布林"}

   ↓

2. API Gateway - HTTP Router
   ├─ RequestIDMiddleware: 生成 request_id
   ├─ LoggingMiddleware: 记录请求日志
   ├─ RecoveryMiddleware: panic恢复
   └─ ChatHandler: 接收请求

   ↓

3. Orchestrator - 主处理循环
   │
   ├─► 3.1 ContextBuilder (构建对话上下文)
   │    ├─ MessageLoader: 从MessageStore加载最近50条消息
   │    ├─ StateFetcher: 从MCP Server获取当前状态摘要
   │    ├─ SystemPromptBuilder: 动态生成system message
   │    └─ TokenBudgetManager: Token预算管理，裁剪历史消息
   │
   ├─► 3.2 LLMManager (调用LLM) 【循环开始】
   │    ├─ OpenAIClient: 发送ChatCompletion请求
   │    ├─ RetryStrategy: 失败自动重试（最多3次）
   │    └─ 返回: LLMResponse {content, tool_calls}
   │
   ├─► 3.3 判断: tool_calls是否为空？
   │    │
   │    ├─ 是 → 跳到 3.6
   │    │
   │    └─ 否 → 继续执行
   │
   ├─► 3.4 ToolCoordinator (执行工具调用)
   │    ├─ FormatConverter: LLM格式 → MCP格式
   │    ├─ MCPInvoker: 调用MCP Server
   │    │   POST http://mcp-server/tools/call
   │    │   {session_id, tool_name, arguments}
   │    ├─ ResultParser: 解析MCP响应
   │    │   ├─ 提取 result
   │    │   └─ 提取 events
   │    └─ EventCollector: 收集所有事件
   │
   ├─► 3.5 EventDispatcher (推送事件到前端)
   │    └─ WebSocketManager: 广播事件到前端
   │        ws://api/sessions/sess-001/ws
   │        {type: "event", event: {...}}
   │
   ├─► 3.6 判断: 是否达到最大迭代次数(10次)？
   │    │
   │    ├─ 是 → 返回错误: ErrMaxIterationsExceeded
   │    │
   │    └─ 否 → 继续循环 (回到3.2)
   │         将tool results作为tool messages发送给LLM
   │
   └─► 3.7 ResponseGenerator (生成最终响应)
        ├─ TextExtractor: 提取LLM生成的叙述文本
        ├─ StateChangeExtractor: 从tool results提取状态变更
        ├─ TurnInfoExtractor: 提取回合信息
        └─ RollRequirementChecker: 检查是否需要玩家投骰

   ↓

4. API Gateway - ChatHandler
   └─ 返回HTTP响应: ChatResponse
       {
         response: "你冲向哥布林，挥剑砍中...",
         state_changes: {goblin_1: {hp: 5, hp_max: 13}},
         requires_roll: false,
         usage: {prompt_tokens: 1200, ...}
       }

   ↓

5. 前端接收响应并更新UI
```

### 6.2 组件交互时序图

```
前端      API Gateway    Orchestrator     ContextBuilder   LLMManager    ToolCoordinator   MCP Server   EventDispatcher
 │            │              │                 │               │                │              │              │
 │ POST /chat  │              │                 │               │                │              │              │
 ├───────────>│              │                 │               │                │              │              │
 │            │  ProcessChat │                 │               │                │              │              │
 │            ├─────────────>│                 │               │                │              │              │
 │            │              │ Build(ctx)      │               │                │              │              │
 │            │              ├────────────────>│               │                │              │              │
 │            │              │                 │ Load msgs     │                │              │              │
 │            │              │                 ├───────────────┼──────────────>│              │              │
 │            │              │                 │ Get state     │                │              │              │
 │            │              │                 ├───────────────────────────────────────────────>│              │
 │            │              │                 │<──────────────────────────────────────────────┤              │
 │            │              │<────────────────┤               │                │              │              │
 │            │              │<────────────────────────────────┤                │              │              │
 │            │              │ Chat(ctx)       │               │                │              │              │
 │            │              ├─────────────────────────────────>│                │              │              │
 │            │              │                 │               │ Call LLM       │              │              │
 │            │              │                 │               ├───────────────>│              │              │
 │            │              │                 │               │                │              │              │
 │            │              │                 │               │<───────────────┤              │              │
 │            │              │<─────────────────────────────────┤                │              │              │
 │            │              │                 │               │ tool_calls?    │              │              │
 │            │              │ Execute(tools)  │               │                │              │              │
 │            │              ├─────────────────────────────────────────────────>│              │              │
 │            │              │                 │               │                │              │ /tools/call  │
 │            │              │                 │               │                │              ├─────────────>│
 │            │              │                 │               │                │              │              │
 │            │              │                 │               │                │<─────────────┤              │
 │            │              │<─────────────────────────────────────────────────┤                │              │
 │            │              │ Dispatch(events) │               │                │              │              │
 │            │              ├──────────────────────────────────────────────────────────────────>│              │
 │            │              │                 │               │                │              │              │
 │            │<───────────────────────────────────────────────────────────────────────────────┤              │
 │            │              │                 │               │                │              │              │
 │            │  Chat() again│ (with tool results)           │                │              │              │
 │            │              ├─────────────────────────────────>│                │              │              │
 │            │              │                 │               │                │              │              │
 │            │              │                 │               │<───────────────┤              │              │
 │            │              │                 │               │ No tool_calls  │              │              │
 │            │              │ Generate(resp)  │               │                │              │              │
 │            │              ├─────────────────────────────────────────────────>│              │              │
 │            │              │<─────────────────────────────────────────────────┤                │              │
 │            │<─────────────┤                 │               │                │              │              │
 │<───────────│              │                 │               │                │              │              │
 │  200 OK    │              │                 │               │                │              │              │
```

---

## 七、错误处理策略

### 7.1 错误分类

| 错误类型     | Code                    | HTTP状态 | 说明           |
| -------- | ----------------------- | ------ | ------------ |
| 会话不存在    | SESSION_NOT_FOUND       | 404    | session_id无效 |
| 版本冲突     | VERSION_CONFLICT        | 409    | 乐观锁冲突        |
| 最大迭代超限   | MAX_ITERATIONS_EXCEEDED | 500    | 工具调用循环超过10次  |
| 参数错误     | INVALID_REQUEST         | 400    | 请求参数无效       |
| LLM不可用   | LLM_UNAVAILABLE         | 503    | LLM服务异常      |
| MCP服务器错误 | MCP_SERVER_ERROR        | 502    | MCP Server异常 |
| 内部错误     | INTERNAL_ERROR          | 500    | 服务器内部错误      |

### 7.2 重试策略

**可重试错误**：

- 网络超时
- HTTP 429 (Rate Limit)
- HTTP 503 (Service Unavailable)
- HTTP 502 (Bad Gateway)

**不可重试错误**：

- HTTP 400 (Bad Request)
- HTTP 404 (Not Found)
- HTTP 409 (Conflict)

**重试配置**：

- 最大重试次数：3
- 重试延迟：递增（1s, 2s, 4s）

### 7.3 错误响应格式

```json
{
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session not found: uuid-xxx"
  }
}
```

---

## 八、配置管理

### 8.1 配置文件示例 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30
  write_timeout: 30
  shutdown_timeout: 10

llm:
  provider: "openai"
  api_key: "${LLM_API_KEY}"
  model: "gpt-4"
  max_tokens: 4096
  temperature: 0.7
  max_retries: 3
  retry_delay: 1000
  timeout: 30

mcp:
  base_url: "http://localhost:9000"
  timeout: 10
  max_connections: 100

orchestrator:
  max_iterations: 10
  context_window_size: 8192
  message_history_size: 50
  token_budget: 6000

websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  ping_interval: 30
  pong_timeout: 60
```

### 8.2 环境变量

| 变量名            | 说明           | 示例                    |
| -------------- | ------------ | --------------------- |
| LLM_API_KEY    | LLM API密钥    | sk-...                |
| MCP_SERVER_URL | MCP Server地址 | http://localhost:9000 |
| DATABASE_URL   | 数据库连接        | postgres://...        |
| REDIS_URL      | Redis连接      | redis://...           |

---

## 九、部署方案

### 9.1 Docker Compose 部署

```yaml
version: '3.8'

services:
  mcp-client:
    build: .
    ports:
      - "8080:8080"
    environment:
      - LLM_API_KEY=${LLM_API_KEY}
      - MCP_SERVER_URL=http://mcp-server:9000
    depends_on:
      - mcp-server
      - redis
    restart: unless-stopped

  mcp-server:
    image: mcp-server:latest
    ports:
      - "9000:9000"
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/mcp
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=mcp
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

### 9.2 水平扩展

```
            Nginx (Load Balancer)
                     │
        ┌────────────┼────────────┐
        │            │            │
   MCP Client   MCP Client   MCP Client
   (Instance 1)  (Instance 2)  (Instance 3)
        │            │            │
        └────────────┼────────────┘
                     │
            ┌─────────┴─────────┐
            │                   │
       MCP Server          Redis
      (有状态)            (共享缓存)
```

**关键点**：

- ✅ MCP Client 无状态，可任意扩展
- ✅ 通过 Nginx 负载均衡
- ✅ 共享 Redis 缓存
- ⚠️ MCP Server 有状态，单实例部署
