---
name: dev-team
description: 启动开发 Team，按设计文档和开发计划进行开发。支持里程碑模式和单任务模式。测试失败必须修复，不允许作弊。
allowed-tools: Read, Glob, Grep, Write, Edit, Bash
argument-hint: <mode> <target> [task]
---

# 开发 Team

组织 Developer、Whitebox Tester、Blackbox Tester 协同工作，实现代码并通过双重测试验证。

## 核心原则

### 原则一：零容忍作弊

**测试失败必须修复代码，绝不能绕过测试。**

```
❌ 严格禁止：
- 跳过失败的测试
- 修改测试使其通过
- 使用 Mock 替代真实测试
- 伪造测试结果
- 降低测试标准

✅ 必须做到：
- 测试失败时修复代码
- 分析问题根本原因
- 本地验证后重新提交
- 保持测试的真实性
```

### 原则二：必须集成到主函数

**代码实现了不等于完成了，必须集成到 main.go 中才能真正使用。**

```
❌ 严格禁止：
- 只实现功能模块，不注册到 main.go
- 在 main.go 中只留 TODO 注释
- 功能代码存在但无法通过 API 访问
- 测试绕过 main.go 直接测试模块
- 多人同时修改 main.go 造成冲突

✅ 必须做到：
- 任务完成时只实现功能代码，暂不修改 main.go
- 里程碑所有任务完成后，统一集成到 main.go
- 集成工作由 Team Lead 协调，避免多人冲突
- E2E 测试在集成完成后验证功能可通过 API 访问
```

#### 集成时机与流程

```
┌──────────────────────────────────────────────────────────────────┐
│                        里程碑执行流程                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  任务 T1: 实现 → 单元测试 → 集成测试 (不涉及 main.go)              │
│     ↓                                                            │
│  任务 T2: 实现 → 单元测试 → 集成测试 (不涉及 main.go)              │
│     ↓                                                            │
│  任务 T3: 实现 → 单元测试 → 集成测试 (不涉及 main.go)              │
│     ↓                                                            │
│  ─────────────────────────────────────────────                   │
│  【里程碑集成阶段】由 Team Lead 统一执行                           │
│     ↓                                                            │
│  1. 汇总所有任务需要集成的组件                                     │
│  2. 统一更新 main.go（一次性修改）                                 │
│  3. 构建并验证编译通过                                            │
│     ↓                                                            │
│  E2E 测试：验证所有功能可通过 API 访问                             │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

#### 集成检查清单

| 类型 | 集成点 | 验证时机 | 验证方式 |
|------|--------|----------|----------|
| **Tool** | `main.go` 中注册 | 里程碑集成后 | `GET /mcp/tools` 返回该工具 |
| **Service** | `main.go` 中初始化 | 里程碑集成后 | 依赖该服务的功能可正常调用 |
| **Route** | `main.go` 中注册 | 里程碑集成后 | HTTP 请求返回正确响应 |
| **Middleware** | `main.go` 中挂载 | 里程碑集成后 | 中间件效果可见（如日志、认证） |
| **Store** | `main.go` 中连接 | 里程碑集成后 | 数据读写正常工作 |

#### 集成需求收集

```
每个任务完成时，Developer 需要在任务报告中声明集成需求：

## 任务完成报告

### 功能代码
- [x] internal/api/tools/dice.go - DiceTools 实现

### 单元测试
- [x] tests/unit/service/dice_test.go - 通过

### 集成测试
- [x] tests/integration/dice_test.go - 通过

### 【集成需求】(里程碑集成阶段使用)
需要在 main.go 中添加：
```go
// Step X: Initialize DiceService
diceService := service.NewDiceService()

// Step Y: Register DiceTools
diceTools := tools.NewDiceTools(diceService)
diceTools.Register(server.Registry())
```

依赖关系：
- DiceService 无外部依赖
- DiceTools 依赖 DiceService
```

## 目标组件

支持两种目标组件的开发：

| 组件 | 计划文档 | 设计文档 | 工作目录 |
|------|----------|----------|----------|
| **server** | `docs/server/plan/README.md` | `docs/server/详细设计.md` | `packages/server/` |
| **client** | `docs/client/重构计划.md` | `docs/client/详细设计.md` | `packages/client/` |

## 模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `milestone` | 执行整个里程碑 | `/dev-team milestone server M1` |
| `task` | 执行单个任务 | `/dev-team task server T1-1` |
| `continue` | 继续上次中断的任务 | `/dev-team continue` |

### 使用示例

```bash
# Server 开发
/dev-team milestone server M1     # 执行 M1 里程碑所有任务
/dev-team task server T1-1        # 执行单个任务 T1-1

# Client 开发
/dev-team milestone client M1     # 执行 M1 里程碑（如有）
/dev-team task client Step1       # 执行单个任务 Step1
```

## Team 结构

```
┌─────────────────────────────────────────┐
│           Dev Team Lead                  │
│   (任务分配 + 冲突避免 + 修复监督)         │
│   (集成验证 + 预防"实现了未集成"问题)      │
└───────────────┬─────────────────────────┘
                │
                ▼
        ┌───────────────┐
        │   Developer   │
        │   实现代码     │
        │   集成到main   │  ← 新增：必须集成
        │   修复问题     │
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Whitebox      │
        │ Tester        │
        │ 单元+集成测试  │
        │ 反作弊检查    │
        │ 集成代码检查   │  ← 新增：检查 main.go
        └───────┬───────┘
                │
        ┌───────┴───────┐
        │               │
      通过            失败
        │               │
        ▼               ▼
┌───────────────┐ ┌───────────────┐
│ Blackbox      │ │ 通知 Developer│
│ Tester        │ │ 修复问题      │
│ E2E测试       │ └───────┬───────┘
│ 服务验证      │         │
│ API可达验证   │  ← 新增：验证可通过API访问
└───────────────┘         │
                          ▼
                  ┌───────────────┐
                  │ 重新测试      │
                  │ (最多3次)     │
                  └───────────────┘
```

## 组件配置

根据目标组件（server/client）使用不同的配置：

### Server 配置

```yaml
# Server 组件配置
component: server
paths:
  plan: docs/server/plan/README.md
  design: docs/server/详细设计.md
  source: packages/server/
  tests:
    unit: packages/server/tests/unit/
    integration: packages/server/tests/integration/
    e2e: packages/server/tests/e2e/
scripts:
  build: packages/server/scripts/build.ps1
  test: packages/server/scripts/test-all.ps1
  test_unit: go test -v ./packages/server/tests/unit/...
  test_integration: go test -v ./packages/server/tests/integration/...
  test_e2e: packages/server/scripts/test-e2e.ps1
binary: packages/server/bin/dnd-server.exe
```

### Client 配置

```yaml
# Client 组件配置
component: client
paths:
  plan: docs/client/重构计划.md
  design: docs/client/详细设计.md
  source: packages/client/
  tests:
    unit: packages/client/tests/unit/
    integration: packages/client/tests/integration/
    e2e: packages/client/tests/e2e/
scripts:
  build: packages/client/scripts/build.ps1
  test: packages/client/scripts/test-all.ps1
  test_unit: go test -v ./packages/client/tests/unit/...
  test_integration: go test -v ./packages/client/tests/integration/...
  test_e2e: packages/client/scripts/test-e2e.ps1
binary: packages/client/bin/dnd-api.exe
```

### 配置选择规则

```
1. 解析 $ARGUMENTS 中的目标组件（server/client）
2. 加载对应组件的配置
3. 所有后续操作使用组件特定路径
4. 测试脚本使用组件特定命令
```

## 测试-修复循环

```
完整的测试-修复流程：

1. Developer 完成代码
2. Whitebox Tester 执行测试
3. 如果测试失败：
   a. Tester 生成详细的失败报告
   b. Tester 通知 Developer
   c. Developer 分析问题
   d. Developer 修复代码
   e. Developer 本地验证
   f. 重新提交测试
   g. 重复直到通过（最多3次）
4. 测试通过后进入下一阶段
5. Blackbox Tester 同理
```

## 测试失败报告格式

```markdown
## 测试失败报告

### 任务: T1-2

### 失败的测试
- 测试文件: tests/integration/session_test.go
- 测试函数: TestSessionCreate
- 行号: 45

### 失败详情
```
=== RUN   TestSessionCreate
    session_test.go:45: 期望状态码 201，实际 500
    session_test.go:46: 错误信息: "名称不能为空"
```

### 期望行为
创建会话时，如果名称为空，应该返回 400 Bad Request，
而不是 500 Internal Server Error。

### 实际行为
返回了 500 错误，说明服务端未正确处理边界条件。

### 相关代码
- 文件: internal/service/session.go
- 函数: CreateSession
- 可能行号: 30-50

### 修复建议
1. 在 CreateSession 开头添加输入验证
2. 返回 400 而非 500
3. 添加对应的单元测试

### 要求
请修复代码使测试通过。
注意：你不能跳过此测试或修改测试条件。
```

## Developer 修复流程

```markdown
## 修复响应

### 问题分析
1. 问题定位: internal/service/session.go:35
2. 根本原因: 未验证空名称输入
3. 影响范围: 仅影响会话创建

### 修复方案
在 CreateSession 方法开头添加名称验证：
```go
if req.Name == "" {
    return nil, errors.ErrInvalidInput
}
```

### 修复实施
已完成代码修改。

### 本地验证
```
go test -v ./tests/unit/service/session_test.go
=== RUN   TestCreateSession
--- PASS: TestCreateSession (0.00s)
PASS
```

请重新执行测试验证。
```

## 反作弊检查

### Whitebox Tester 检查项

```
在执行测试前，检查：

1. 无 t.Skip() 在失败测试上
   grep -r "t\.Skip()" tests/

2. 无手动构建路由
   grep -r "setupRouter" tests/integration/

3. 无空测试函数
   grep -r "func Test.*{\s*}" tests/

4. 所有断言有效
   检查测试是否包含真实断言

5. 检查 Developer 提交了集成需求声明
   - 任务报告是否包含【集成需求】部分？
   - 集成代码示例是否正确？

发现问题 → 拒绝测试 → 报告作弊行为
```

#### 集成需求声明检查

```
Whitebox Tester 需验证 Developer 的任务报告包含：

1. 【集成需求】部分存在
2. 提供了完整的集成代码示例
3. 声明了依赖关系
4. 代码示例符合项目规范

如果缺少集成需求声明：
```
## 报告不完整

### 任务: T4-3 骰子 Tools

### 问题
任务报告缺少【集成需求】部分。

### 要求
请补充以下内容：
1. 需要在 main.go 中添加的代码
2. 依赖的其他服务或组件
3. 初始化顺序建议
```
```

### Blackbox Tester 检查项

```
【重要】Blackbox Tester 在里程碑集成完成后执行，验证所有功能可通过 API 访问。

在执行测试前，检查：

1. 使用真实二进制
   - 必须从源码构建（集成后重新构建）
   - 不能使用旧版本

2. 真实服务启动
   - 必须启动新进程
   - 健康检查必须通过

3. 真实网络请求
   - 不能使用 Mock
   - 必须验证响应内容

4. 【核心】验证 main 函数集成正确
   - 检查所有注册的 Tool 是否在 /mcp/tools 返回
   - 检查所有注册的路由是否可访问
   - 检查所有注册的服务是否正常工作
   - 检查所有注册的中间件是否生效

5. 【核心】验证功能端到端可用
   - 每个任务的功能必须能通过真实 API 调用
   - 不允许"代码存在但无法访问"的情况

发现问题 → 拒绝测试 → 报告集成缺失
```

### E2E 测试要求

**核心原则：E2E 测试必须拉起真实服务，验证功能端到端可用。**

```
E2E 测试流程：

1. 构建二进制
   - 从源码编译：go build -o bin/dnd-server.exe ./cmd/server
   - 验证构建成功

2. 启动服务
   - 运行二进制：./bin/dnd-server.exe
   - 等待服务就绪（健康检查通过）
   - 记录服务 PID

3. 验证 main 函数注册
   - 检查所有 HTTP 路由已注册
   - 检查所有服务依赖已注入
   - 检查所有中间件已挂载

4. 执行功能测试
   - 发送真实 HTTP 请求
   - 验证响应状态码和内容
   - 验证业务逻辑正确性

5. 清理
   - 停止服务进程
   - 清理测试数据
```

#### main 函数注册验证清单

```
对于每个注册项，E2E 测试必须验证：

1. 路由注册
   - [ ] GET /api/xxx 返回正确响应
   - [ ] POST /api/xxx 返回正确响应
   - [ ] 未知路由返回 404

2. 服务注册
   - [ ] 依赖注入的服务可正常调用
   - [ ] 服务初始化顺序正确
   - [ ] 服务关闭时清理正确

3. 中间件注册
   - [ ] 日志中间件生效
   - [ ] 认证中间件生效（如有）
   - [ ] 错误恢复中间件生效

4. 存储连接
   - [ ] Redis 连接正常
   - [ ] PostgreSQL 连接正常（如有）
```

## 执行流程

### Step 1: 解析计划

```
1. 读取开发计划文档
2. 解析里程碑和任务定义
3. 分析任务依赖关系
4. 分析文件冲突
5. 生成执行计划
```

### Step 2: 执行任务（不涉及 main.go）

```
对于每个任务：

1. 检查前置条件
2. Spawn Developer teammate
3. 等待开发完成（只实现功能代码，不修改 main.go）
4. Developer 提交集成需求声明
5. Spawn Whitebox Tester
6. 处理测试结果：
   - 通过 → 继续
   - 失败 → 进入修复循环
7. 任务标记为"功能完成，待集成"
```

**注意**：此阶段只执行单元测试和集成测试，不执行 E2E 测试。

### Step 3: 里程碑集成（由 Team Lead 执行）

```
当所有任务功能完成后：

1. 收集所有任务的集成需求
2. 统一更新 main.go：
   a. 汇总所有需要初始化的 Service
   b. 汇总所有需要注册的 Tool
   c. 汇总所有需要挂载的 Route/Middleware
   d. 按依赖顺序排列初始化代码
   e. 一次性修改 main.go
3. 构建验证：确保编译通过
4. 启动验证：确保服务能正常启动
```

### Step 4: E2E 测试

```
集成完成后：

1. Spawn Blackbox Tester
2. 执行 E2E 测试：
   - 验证所有注册的 Tool 可通过 API 调用
   - 验证所有注册的路由可访问
   - 验证服务间依赖正常工作
3. 处理测试结果：
   - 通过 → 进入里程碑完成
   - 失败 → 进入集成修复循环
```

### Step 5: 修复循环

```
当测试失败时：

1. 生成失败报告
2. 通知 Developer
3. 等待修复
4. 重新测试
5. 如果仍失败，重复 1-4
6. 最多 3 次后仍失败：
   - 标记为阻塞
   - 请求人工介入
```

### Step 6: 里程碑完成

```
当所有任务完成且集成验证通过：

1. 服务启动验证
2. 检查无作弊记录
3. 生成里程碑报告
4. 通知体验 Team
```

## 进度报告格式

```markdown
## 开发进度报告

### 组件: Server/Client
### 当前里程碑/任务: M1 / Step1

### 任务状态
| ID | 任务 | 开发 | 白盒 | 集成需求 | 状态 |
|----|------|:----:|:----:|:--------:|------|
| T1-1 | 会话创建 | ✅ | ✅ | ✅ | 🔜 待集成 |
| T1-2 | 会话管理 | ✅ | ✅ | ✅ | 🔜 待集成 |
| T1-3 | 消息发送 | 🔄 | - | - | 🔄 开发中 |

### 集成需求汇总
| 任务 | 需要集成的组件 | 依赖 |
|------|---------------|------|
| T1-1 | SessionService, SessionStore | Redis |
| T1-2 | SessionHandler | SessionService |

### 里程碑集成状态
- [ ] 所有任务功能完成
- [ ] 集成需求汇总完成
- [ ] main.go 更新完成
- [ ] E2E 测试通过

### 修复记录
| Task | 问题 | 修复次数 | 结果 |
|------|------|:--------:|------|
| T1-2 | 空指针异常 | 1 | 通过 |

### 阻塞问题
| Task | 问题 | 尝试 | 需要 |
|------|------|:----:|------|
| - | - | - | - |

### 统计
- 完成任务(待集成): 2/5
- 总修复次数: 1
- 作弊记录: 0
- 测试通过率: 100%
```

## 里程碑完成报告

```markdown
## 里程碑完成报告

### 组件: Server/Client
### 里程碑: M1

### 集成阶段总结
- 集成的 Service 数量: 3
- 集成的 Tool 数量: 5
- 集成的 Route 数量: 10
- main.go 修改次数: 1 (统一修改)

### 质量指标
- 测试覆盖率: 87%
- 首次通过率: 80%
- 平均修复次数: 0.5
- 作弊记录: 0

### 测试结果
- 单元测试: 25 passed, 0 failed
- 集成测试: 10 passed, 0 failed
- E2E 测试: 5 passed, 0 failed

### 集成验证
- [x] main.go 已更新（包含所有组件注册）
- [x] 二进制构建成功
- [x] 服务启动成功
- [x] 健康检查通过

### API 可达性验证
- [x] 所有注册的 Tool 可通过 /mcp/tools 访问
- [x] 所有注册的路由可正常调用
- [x] 所有功能端到端可用

### main 函数注册验证
- [x] 路由注册：所有 API 端点可访问
- [x] 服务注册：依赖注入正常工作
- [x] 中间件注册：中间件链生效
- [x] 存储连接：Redis/PostgreSQL 连接正常

### 集成详情
```go
// M1 集成的组件
Services:
  - SessionService (依赖: SessionStore)
  - DiceService (依赖: 无)

Tools:
  - roll_dice
  - roll_check
  - roll_save

Routes:
  - GET  /api/sessions
  - POST /api/sessions
  - ...
```

### 下一步
里程碑已完成，可执行：
/exp-team start <component> M1
```

## 异常处理

### Developer 未提交集成需求

```
Whitebox Tester 发现任务报告缺少集成需求：

Lead 回应模板：
"任务报告缺少【集成需求】部分。
请在报告中补充：
1. 需要在 main.go 中添加的代码
2. 依赖的其他服务或组件
3. 初始化顺序建议

这是里程碑集成阶段的必要信息。"
```

### 集成阶段发现依赖冲突

```
当收集的集成需求存在依赖冲突时：

1. 分析依赖关系图
2. 确定正确的初始化顺序
3. 如果存在循环依赖：
   - 标记为阻塞
   - 请求架构调整
4. 生成最终的集成代码
```

### E2E 测试发现功能不可达

```
当 E2E 测试发现功能无法通过 API 访问时：

1. 检查 main.go 是否正确注册
2. 检查依赖服务是否正常初始化
3. 检查路由/工具定义是否正确
4. 生成详细报告：
   - 哪个功能不可达
   - 可能的原因
   - 建议的修复方案
5. 要求修复并重新测试
```

### Developer 请求跳过测试

```
❌ 拒绝所有跳过请求

Lead 回应模板：
"测试不能被跳过。
测试失败说明代码存在问题，必须修复。
我可以帮助你分析问题，但不能跳过测试。"
```

### 3 次修复后仍失败

```
1. 标记任务为阻塞
2. 生成详细报告：
   - 问题完整描述
   - 已尝试的所有修复
   - 每次修复的测试结果
3. 请求人工介入
```

### 发现作弊行为

```
1. 立即停止任务
2. 记录作弊行为
3. 要求立即纠正
4. 如果拒绝 → 报告用户
```

---

## 初始化流程

```
1. 解析 $ARGUMENTS:
   - 格式: <mode> <component> [target]
   - 示例: milestone server M1
   - 示例: task client Step1

2. 加载组件配置:
   - server → docs/server/plan/README.md + packages/server/
   - client → docs/client/重构计划.md + packages/client/

3. 读取计划文档，解析里程碑和任务

4. 开始执行开发流程...
```

---

**当前参数**: $ARGUMENTS

开始执行开发流程...
