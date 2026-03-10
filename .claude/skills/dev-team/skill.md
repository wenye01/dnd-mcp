---
name: dev-team
description: 启动开发 Team，按设计文档和开发计划进行开发。支持里程碑模式和单任务模式。测试失败必须修复，不允许作弊。
argument-hint: <mode> <target> [task]
---

# 开发 Team

组织 Developer、Whitebox Tester、Blackbox Tester 协同工作，实现代码并通过双重测试验证。

## 核心原则

详见 `agents/dev/lead.md`

- **零容忍作弊**: 测试失败必须修复代码，不能绕过
- **必须集成到 main.go**: 代码实现了不等于完成，必须集成才能真正使用

## 角色索引

| 角色 | 文件 | 职责 |
|------|------|------|
| **Lead** | `agents/dev/lead.md` | 任务分配、冲突避免、进度监控、集成协调、质量把关 |
| **Developer** | `agents/dev/developer.md` | 代码实现、修复问题、声明集成需求 |
| **Whitebox Tester** | `agents/dev/whitebox-tester.md` | 单元+集成测试、编写测试用例、反作弊检查 |
| **Blackbox Tester** | `agents/dev/blackbox-tester.md` | E2E测试、部署服务、更新测试脚本 |

## 测试类型说明

### Whitebox Tester (白盒测试)

**职责**: 根据原始需求和代码编写测试用例

```
1. 读取原始需求 (docs/{component}/需求-*.md)
2. 读取设计文档 (docs/{component}/详细设计.md)
3. 分析代码逻辑
4. 编写单元测试用例 - 覆盖核心逻辑
5. 编写集成测试用例 - 覆盖模块协作
6. 执行测试并验证
```

详见 `agents/dev/whitebox-tester.md`

### Blackbox Tester (黑盒测试)

**职责**: 根据原始需求实际使用服务进行测试

```
1. 部署服务 (构建 + 启动)
2. 根据原始需求编写 E2E 测试用例
3. 实际使用服务进行端到端验证
4. 验证 API 可达性
5. 更新 E2E 测试脚本 (tests/e2e/)
```

详见 `agents/dev/blackbox-tester.md` - 包含部署脚本和 E2E 测试脚本

## 模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `milestone` | 执行整个里程碑 | `/dev-team milestone server M1` |
| `task` | 执行单个任务 | `/dev-team task server T1-1` |
| `continue` | 继续上次中断的任务 | `/dev-team continue` |

### 使用示例

```bash
/dev-team milestone server M1     # 执行 M1 里程碑
/dev-team task server T1-1        # 执行单个任务
/dev-team continue                # 继续中断任务
```

## 组件

| 组件 | 计划文档 | 设计文档 | 工作目录 |
|------|----------|----------|----------|
| **server** | docs/server/plan/README.md | docs/server/详细设计.md | packages/server/ |
| **client** | docs/client/重构计划.md | docs/client/详细设计.md | packages/client/ |

## 工作流程

```
1. Lead 读取计划文档，解析里程碑和任务
2. Lead 分配任务给 Developer
3. Developer 实现代码，声明集成需求
4. Whitebox Tester 编写并执行单元+集成测试
5. 测试失败 → Developer 修复 → 重新测试 (最多3次)
6. 所有任务完成后，Lead 统一集成到 main.go
7. Blackbox Tester 部署服务并执行 E2E 测试
8. 里程碑完成，通知 Exp Team
```

## 进度报告

详见 `agents/dev/lead.md`

---

**当前参数**: $ARGUMENTS

开始执行开发流程...
