---
name: dev-team
description: 启动开发 Team，按设计文档和开发计划进行开发。支持里程碑模式和单任务模式。测试失败必须修复，不允许作弊。
allowed-tools: Read, Glob, Grep, Write, Edit, Bash
argument-hint: <mode> <target> [task]
---

# 开发 Team

组织 Developer、Whitebox Tester、Blackbox Tester 协同工作，实现代码并通过双重测试验证。

## 核心原则：零容忍作弊

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
└───────────────┬─────────────────────────┘
                │
                ▼
        ┌───────────────┐
        │   Developer   │
        │   实现代码     │
        │   修复问题     │
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │ Whitebox      │
        │ Tester        │
        │ 单元+集成测试  │
        │ 反作弊检查    │
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

发现问题 → 拒绝测试 → 报告作弊行为
```

### Blackbox Tester 检查项

```
在执行测试前，检查：

1. 使用真实二进制
   - 必须从源码构建
   - 不能使用旧版本

2. 真实服务启动
   - 必须启动新进程
   - 健康检查必须通过

3. 真实网络请求
   - 不能使用 Mock
   - 必须验证响应内容

4. 验证 main 函数注册功能
   - 检查所有注册的路由是否可用
   - 检查所有注册的服务是否正常
   - 检查所有注册的中间件是否生效

发现问题 → 拒绝测试 → 报告作弊行为
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

### Step 2: 执行任务

```
对于每个任务：

1. 检查前置条件
2. Spawn Developer teammate
3. 等待开发完成
4. Spawn Whitebox Tester
5. 处理测试结果：
   - 通过 → 继续
   - 失败 → 进入修复循环
6. Spawn Blackbox Tester
7. 处理测试结果：
   - 通过 → 任务完成
   - 失败 → 进入修复循环
```

### Step 3: 修复循环

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

### Step 4: 里程碑完成

```
当所有任务完成：

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
| ID | 任务 | 开发 | 白盒 | 黑盒 | 修复 | 状态 |
|----|------|:----:|:----:|:----:|:----:|------|
| T1-1 | 会话创建 | ✅ | ✅ | ✅ | 0 | ✅ 完成 |
| T1-2 | 会话管理 | ✅ | ❌→✅ | ✅ | 1 | ✅ 完成 |
| T1-3 | 消息发送 | 🔄 | - | - | 0 | 🔄 开发中 |

### 修复记录
| Task | 问题 | 修复次数 | 结果 |
|------|------|:--------:|------|
| T1-2 | 空指针异常 | 1 | 通过 |

### 阻塞问题
| Task | 问题 | 尝试 | 需要 |
|------|------|:----:|------|
| - | - | - | - |

### 统计
- 完成任务: 2/5
- 总修复次数: 1
- 作弊记录: 0
- 测试通过率: 100%
```

## 里程碑完成报告

```markdown
## 里程碑完成报告

### 组件: Server/Client
### 里程碑: M1

### 质量指标
- 测试覆盖率: 87%
- 首次通过率: 80%
- 平均修复次数: 0.5
- 作弊记录: 0

### 测试结果
- 单元测试: 25 passed, 0 failed
- 集成测试: 10 passed, 0 failed
- E2E 测试: 5 passed, 0 failed

### 服务验证
- [x] 二进制构建成功
- [x] 服务启动成功
- [x] 健康检查通过
- [x] 所有功能可用

### main 函数注册验证
- [ ] 路由注册：所有 API 端点可访问
- [ ] 服务注册：依赖注入正常工作
- [ ] 中间件注册：中间件链生效
- [ ] 存储连接：Redis/PostgreSQL 连接正常

### 下一步
里程碑已完成，可执行：
/exp-team start <component> M1
```

## 异常处理

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
