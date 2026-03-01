# Milestone 2: MCP 协议对接

> **状态**: ⏳ 待开发
> **依赖**: M1 (GLM 集成)
> **任务数**: 4
> **预计工作量**: 3-4 天

---

## 2.1 目标

实现 Client 与 Server 的 MCP 协议对接，使 Client 能够调用 Server 提供的工具。

---

## 2.2 范围

- **包含**:
  - MCP 客户端实现
  - Server 工具映射
  - LLM Tool Call → MCP Tool 转换
  - 结果处理和响应生成

- **不包含**:
  - Server 端实现（已完成 M1-M5）
  - 端到端测试（M3）

---

## 2.3 任务清单

| ID | 任务名称 | 依赖 | 复杂度 | 状态 |
|----|----------|------|--------|------|
| [T2-1](#task-t2-1-mcp-客户端实现) | MCP 客户端实现 | M1 | M | ⏳ |
| [T2-2](#task-t2-2-server-工具映射) | Server 工具映射 | T2-1 | M | ⏳ |
| [T2-3](#task-t2-3-对话编排服务) | 对话编排服务 | T2-2 | L | ⏳ |
| [T2-4](#task-t2-4-http-api-适配) | HTTP API 适配 | T2-3 | M | ⏳ |

---

## 2.4 验收标准

- [ ] Client 可连接 Server
- [ ] 可调用 Server 工具
- [ ] LLM Tool Call 正确转换为 MCP 调用
- [ ] 结果正确返回给 LLM 和用户
- [ ] 测试覆盖 > 80%

---

## 详细任务定义

### Task T2-1: MCP 客户端实现

**一句话描述**: 实现与 MCP Server 的通信

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-020

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 创建 `internal/mcp/client.go`
- [ ] 实现 stdio 传输
- [ ] 实现工具列表获取
- [ ] 实现工具调用

**接口层**:
- [ ] 定义 `MCPClient` 接口

**测试层**:
- [ ] 单元测试
- [ ] Mock Server 测试

#### 技术细节

```go
// internal/mcp/client.go

package mcp

import (
    "context"
    "encoding/json"
    "os/exec"
)

// Client MCP 客户端
type Client struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.Reader
    tools  []Tool
}

// Config MCP 客户端配置
type Config struct {
    ServerCommand string   // Server 启动命令
    ServerArgs    []string // Server 参数
}

// NewClient 创建客户端
func NewClient(config Config) (*Client, error) {
    // 启动 Server 子进程
    cmd := exec.Command(config.ServerCommand, config.ServerArgs...)

    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    if err := cmd.Start(); err != nil {
        return nil, err
    }

    return &Client{
        cmd:    cmd,
        stdin:  stdin,
        stdout: stdout,
    }, nil
}

// Initialize 初始化连接
func (c *Client) Initialize(ctx context.Context) error {
    // 发送 initialize 请求
    // 获取服务器能力
    // 获取工具列表
}

// ListTools 获取工具列表
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
    req := Request{
        Method: "tools/list",
    }
    resp, err := c.call(ctx, req)
    // 解析工具列表
}

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error) {
    req := Request{
        Method: "tools/call",
        Params: map[string]any{
            "name":      name,
            "arguments": args,
        },
    }
    resp, err := c.call(ctx, req)
    // 解析结果
}

// Close 关闭连接
func (c *Client) Close() error {
    return c.cmd.Process.Kill()
}
```

#### 约束条件
- 使用 stdio 传输（单机部署）
- 兼容 MCP 协议规范
- 支持超时控制

#### 验收标准
- [ ] 可启动 Server 子进程
- [ ] 可获取工具列表
- [ ] 可调用工具
- [ ] 测试覆盖 > 80%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/mcp/client.go` | MCP 客户端 |
| 新增 | `internal/mcp/types.go` | 类型定义 |
| 新增 | `internal/mcp/transport.go` | 传输层 |
| 新增 | `tests/unit/mcp/client_test.go` | 单元测试 |

---

### Task T2-2: Server 工具映射

**一句话描述**: 将 Server MCP Tools 映射为 LLM Tools

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-021

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 创建 `internal/mcp/tool_adapter.go`
- [ ] 实现 MCP Tool → LLM Tool 转换
- [ ] 实现参数格式转换

**接口层**:
- [ ] 提供 `GetLLMTools()` 方法

**测试层**:
- [ ] 转换测试

#### 技术细节

```go
// internal/mcp/tool_adapter.go

// ToolAdapter 工具适配器
type ToolAdapter struct {
    mcpClient *Client
}

// NewToolAdapter 创建适配器
func NewToolAdapter(mcpClient *Client) *ToolAdapter {
    return &ToolAdapter{mcpClient: mcpClient}
}

// GetLLMTools 获取 LLM 格式的工具列表
func (a *ToolAdapter) GetLLMTools() ([]glm.Tool, error) {
    mcpTools, err := a.mcpClient.ListTools(context.Background())
    if err != nil {
        return nil, err
    }

    var llmTools []glm.Tool
    for _, t := range mcpTools {
        llmTools = append(llmTools, glm.Tool{
            Type: "function",
            Function: glm.ToolFunction{
                Name:        t.Name,
                Description: t.Description,
                Parameters:  t.InputSchema,
            },
        })
    }
    return llmTools, nil
}

// ExecuteToolCall 执行 LLM 工具调用
func (a *ToolAdapter) ExecuteToolCall(ctx context.Context, call glm.ToolCall) (*glm.ToolResult, error) {
    // 解析参数
    var args map[string]any
    json.Unmarshal([]byte(call.Function.Arguments), &args)

    // 调用 MCP 工具
    result, err := a.mcpClient.CallTool(ctx, call.Function.Name, args)
    if err != nil {
        return &glm.ToolResult{
            ToolCallID: call.ID,
            Content:    err.Error(),
            IsError:    true,
        }, nil
    }

    return &glm.ToolResult{
        ToolCallID: call.ID,
        Content:    result.Content,
    }, nil
}
```

#### 约束条件
- 保持参数语义一致
- 错误信息友好

#### 验收标准
- [ ] 工具列表正确转换
- [ ] 参数正确传递
- [ ] 结果正确返回
- [ ] 测试覆盖 > 80%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/mcp/tool_adapter.go` | 工具适配器 |
| 新增 | `tests/unit/mcp/tool_adapter_test.go` | 单元测试 |

---

### Task T2-3: 对话编排服务

**一句话描述**: 实现完整的对话编排流程

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-022

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 创建 `internal/service/orchestrator.go`
- [ ] 实现对话流程管理
- [ ] 实现多轮工具调用
- [ ] 实现上下文管理

**接口层**:
- [ ] 提供 `ProcessMessage()` 方法

**测试层**:
- [ ] 编排流程测试

#### 技术细节

```go
// internal/service/orchestrator.go

// Orchestrator 对话编排服务
type Orchestrator struct {
    llmClient   *glm.Client
    toolAdapter *mcp.ToolAdapter
    maxTurns    int // 最大工具调用轮数
}

// ProcessMessage 处理用户消息
func (o *Orchestrator) ProcessMessage(ctx context.Context, sessionID, userMessage string) (*Response, error) {
    // 1. 获取上下文
    messages, err := o.buildContext(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    // 2. 添加用户消息
    messages = append(messages, glm.Message{
        Role:    "user",
        Content: userMessage,
    })

    // 3. 获取工具列表
    tools, _ := o.toolAdapter.GetLLMTools()

    // 4. 调用 LLM
    response, err := o.processWithTools(ctx, messages, tools)
    if err != nil {
        return nil, err
    }

    // 5. 保存消息
    o.saveMessages(ctx, sessionID, messages, response)

    return response, nil
}

// processWithTools 处理可能包含工具调用的响应
func (o *Orchestrator) processWithTools(ctx context.Context, messages []glm.Message, tools []glm.Tool) (*Response, error) {
    for turn := 0; turn < o.maxTurns; turn++ {
        // 调用 LLM
        resp, err := o.llmClient.Chat(ctx, &glm.ChatRequest{
            Model:    "glm-4.7-flash",
            Messages: messages,
            Tools:    tools,
        })
        if err != nil {
            return nil, err
        }

        // 检查是否有工具调用
        if len(resp.ToolCalls) == 0 {
            // 没有工具调用，返回最终响应
            return &Response{
                Content: resp.Content,
            }, nil
        }

        // 执行工具调用
        messages = append(messages, glm.Message{
            Role:      "assistant",
            Content:   resp.Content,
            ToolCalls: resp.ToolCalls,
        })

        for _, call := range resp.ToolCalls {
            result, _ := o.toolAdapter.ExecuteToolCall(ctx, call)
            messages = append(messages, glm.Message{
                Role:       "tool",
                ToolCallID: call.ID,
                Content:    result.Content,
            })
        }
    }

    return nil, errors.New("max turns exceeded")
}
```

#### 约束条件
- 限制最大工具调用轮数（防止死循环）
- 支持流式输出
- 支持取消

#### 验收标准
- [ ] 对话流程正确
- [ ] 多轮工具调用正常
- [ ] 上下文正确传递
- [ ] 测试覆盖 > 80%

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新增 | `internal/service/orchestrator.go` | 编排服务 |
| 新增 | `tests/unit/service/orchestrator_test.go` | 单元测试 |

---

### Task T2-4: HTTP API 适配

**一句话描述**: 适配现有 HTTP API 使用新的编排服务

#### 需求来源
- 设计文档: 整体架构设计.md
- 需求ID: REQ-030

#### 实现范围

**数据层**:
- [ ] 无

**业务层**:
- [ ] 修改 `internal/api/handler/message.go`
- [ ] 注入 Orchestrator

**接口层**:
- [ ] 保持现有 API 接口不变
- [ ] 添加流式响应支持

**测试层**:
- [ ] 集成测试更新

#### 技术细节

```go
// internal/api/handler/message.go

// MessageHandler 消息处理器
type MessageHandler struct {
    orchestrator *service.Orchestrator
    sessionStore store.SessionStore
}

// SendMessage 发送消息
func (h *MessageHandler) SendMessage(c *gin.Context) {
    sessionID := c.Param("id")

    var req SendMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 使用编排服务处理
    response, err := h.orchestrator.ProcessMessage(c.Request.Context(), sessionID, req.Content)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "content": response.Content,
    })
}

// SendMessageStream 流式发送消息
func (h *MessageHandler) SendMessageStream(c *gin.Context) {
    // SSE 流式响应
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")

    // 流式处理...
}
```

#### 约束条件
- 保持 API 向后兼容
- 支持流式响应（SSE）

#### 验收标准
- [ ] API 正常工作
- [ ] 流式响应正常
- [ ] 集成测试通过

#### 文件清单
| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `internal/api/handler/message.go` | 消息处理器 |
| 修改 | `cmd/api/main.go` | 依赖注入 |
| 新增 | `tests/integration/api/orchestrator_test.go` | 集成测试 |

---

## 测试命令

```powershell
# 单元测试
go test -v ./tests/unit/mcp/...
go test -v ./tests/unit/service/orchestrator_test.go

# 集成测试
go test -v ./tests/integration/...
```
