# Milestone 1: GLM-4.7-Flash 集成

> **状态**: ⏳ 待开发
> **依赖**: 无
> **任务数**: 3
> **预计工作量**: 2-3 天

---

## 1.1 目标

将 GLM-4.7-Flash 集成到 MCP Client，实现：
1. LLM API 调用封装
2. 工具调用（Function Calling）支持
3. 流式输出支持

---

## 1.2 范围

- **包含**:
  - GLM API 客户端实现
  - OpenAI 兼容接口适配
  - 工具定义和调用处理
  - 流式响应处理
  - 配置管理

- **不包含**:
  - MCP 协议对接（M2）
  - Server 集成（M2）
  - 端到端测试（M3）

---

## 1.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 | 状态 |
|----|----------|------|--------|------|
| [T1-1](#task-t1-1-glm-api-客户端) | GLM API 客户端 | - | M | ⏳ |
| [T1-2](#task-t1-2-工具调用处理) | 工具调用处理 | T1-1 | M | ⏳ |
| [T1-3](#task-t1-3-流式输出支持) | 流式输出支持 | T1-1 | S | ⏳ |

---

## 1.4 验收标准

- [ ] 可成功调用 GLM-4.7-Flash API
- [ ] 工具调用正确生成和解析
- [ ] 流式输出正常工作
- [ ] 单元测试覆盖 > 80%
- [ ] 错误处理完善

---

## 详细任务定义

### Task T1-1: GLM API 客户端

**一句话描述**: 实现 GLM-4.7-Flash API 调用封装

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-010

#### 实现范围

**数据层**:
- [ ] 无（LLM 是外部服务）

**业务层**:
- [ ] 创建 `internal/llm/glm/client.go`
- [ ] 实现 `GLMClient` 结构体
- [ ] 实现 `Chat()` 方法
- [ ] 实现 `ChatStream()` 方法
- [ ] 添加配置结构 `GLMConfig`

**接口层**:
- [ ] 定义 `LLMClient` 接口（如不存在）
- [ ] GLMClient 实现接口

**测试层**:
- [ ] 单元测试: `tests/unit/llm/glm/client_test.go`
- [ ] Mock 服务器测试

#### 技术细节

```go
// internal/llm/glm/client.go

package glm

import (
    "context"
    "net/http"
)

// Config GLM 客户端配置
type Config struct {
    APIEndpoint string  // https://open.bigmodel.cn/api/paas/v4/chat/completions
    APIKey      string  // 从环境变量读取
    Model       string  // glm-4.7-flash
    MaxTokens   int     // 65536
    Temperature float64 // 1.0
}

// Client GLM API 客户端
type Client struct {
    config     Config
    httpClient *http.Client
}

// NewClient 创建客户端
func NewClient(config Config) *Client {
    return &Client{
        config:     config,
        httpClient: &http.Client{},
    }
}

// ChatRequest 请求结构
type ChatRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    Tools       []Tool    `json:"tools,omitempty"`
    ToolChoice  any       `json:"tool_choice,omitempty"`
    MaxTokens   int       `json:"max_tokens"`
    Temperature float64   `json:"temperature"`
    Stream      bool      `json:"stream,omitempty"`
    Thinking    *Thinking `json:"thinking,omitempty"`
}

// Chat 非流式调用
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    // 实现 API 调用
}

// ChatStream 流式调用
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
    // 实现流式调用
}
```

#### 约束条件
- 使用 OpenAI 兼容格式
- 支持自定义 API 端点（方便测试）
- API Key 不能硬编码

#### 验收标准
- [ ] 可成功调用 API
- [ ] 错误处理完善
- [ ] 测试覆盖 > 80%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/llm/glm/client.go` | GLM 客户端 |
| 新增 | `internal/llm/glm/types.go` | 类型定义 |
| 新增 | `internal/llm/glm/config.go` | 配置 |
| 新增 | `tests/unit/llm/glm/client_test.go` | 单元测试 |

---

### Task T1-2: 工具调用处理

**一句话描述**: 实现 Function Calling 工具调用处理

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-011

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 创建 `internal/llm/glm/tools.go`
- [ ] 实现 `Tool` 结构体
- [ ] 实现 `ToolCall` 解析
- [ ] 实现 `ToolResult` 格式化

**接口层**:
- [ ] 定义 `ToolExecutor` 接口
- [ ] 工具调用流程管理

**测试层**:
- [ ] 单元测试: `tests/unit/llm/glm/tools_test.go`

#### 技术细节

```go
// internal/llm/glm/tools.go

// Tool 工具定义
type Tool struct {
    Type     string       `json:"type"`     // "function"
    Function ToolFunction `json:"function"`
}

// ToolFunction 函数定义
type ToolFunction struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]any         `json:"parameters"`
}

// ToolCall 工具调用
type ToolCall struct {
    ID       string `json:"id"`
    Type     string `json:"type"`
    Function struct {
        Name      string `json:"name"`
        Arguments string `json:"arguments"` // JSON 字符串
    } `json:"function"`
}

// ToolResult 工具结果
type ToolResult struct {
    ToolCallID string `json:"tool_call_id"`
    Content    string `json:"content"`
    IsError    bool   `json:"is_error,omitempty"`
}

// ParseToolCalls 解析工具调用
func ParseToolCalls(response *ChatResponse) ([]ToolCall, error) {
    // 解析响应中的 tool_calls
}

// FormatToolResult 格式化工具结果为消息
func FormatToolResult(result *ToolResult) Message {
    return Message{
        Role:       "tool",
        ToolCallID: result.ToolCallID,
        Content:    result.Content,
    }
}
```

#### 约束条件
- 兼容 OpenAI Function Calling 格式
- 支持多个工具调用
- 错误结果需标记 `is_error`

#### 验收标准
- [ ] 工具定义格式正确
- [ ] 工具调用解析正确
- [ ] 结果格式化正确
- [ ] 测试覆盖 > 80%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/llm/glm/tools.go` | 工具处理 |
| 新增 | `tests/unit/llm/glm/tools_test.go` | 单元测试 |

---

### Task T1-3: 流式输出支持

**一句话描述**: 实现流式响应处理

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-012

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 在 `client.go` 中实现 `ChatStream()`
- [ ] 实现 SSE (Server-Sent Events) 解析
- [ ] 实现流式 chunk 处理

**接口层**:
- [ ] 提供 channel 接口

**测试层**:
- [ ] 流式调用测试

#### 技术细节

```go
// StreamChunk 流式响应块
type StreamChunk struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Choices []struct {
        Index        int          `json:"index"`
        Delta        Delta        `json:"delta"`
        FinishReason *string      `json:"finish_reason"`
    } `json:"choices"`
}

// Delta 增量内容
type Delta struct {
    Role             string     `json:"role,omitempty"`
    Content          string     `json:"content,omitempty"`
    ReasoningContent string     `json:"reasoning_content,omitempty"`
    ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

// ChatStream 流式调用
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error) {
    req.Stream = true

    // 创建 channel
    ch := make(chan StreamChunk, 100)

    go func() {
        defer close(ch)

        // 发起 SSE 请求
        // 解析并发送 chunks
    }()

    return ch, nil
}
```

#### 约束条件
- 使用 SSE 协议
- 正确处理连接断开
- 支持 context 取消

#### 验收标准
- [ ] 流式输出正常
- [ ] 可正确处理 delta
- [ ] 支持取消
- [ ] 测试覆盖 > 70%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/llm/glm/client.go` | 添加流式方法 |
| 新增 | `internal/llm/glm/stream.go` | 流式处理 |
| 新增 | `tests/unit/llm/glm/stream_test.go` | 流式测试 |

---

## 环境配置

```bash
# .env 或环境变量
GLM_API_KEY=your-api-key
GLM_API_ENDPOINT=https://open.bigmodel.cn/api/paas/v4/chat/completions
GLM_MODEL=glm-4.7-flash
GLM_MAX_TOKENS=65536
```

## 测试命令

```powershell
# 单元测试
go test -v ./tests/unit/llm/glm/...

# 集成测试（需要 API Key）
go test -v ./tests/integration/llm/... -tags=integration
```
