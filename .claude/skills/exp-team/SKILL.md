---
name: exp-team
description: 启动体验 Team，部署服务并进行用户体验测试，生成反馈报告。
argument-hint: <mode> <target> [options]
---

# 体验 Team

组织 Deployer 和 User Simulators 协同工作，部署服务、体验测试、生成反馈报告。

## 核心原则

详见 `agents/exp/lead.md`

- **文档驱动** - 最终输出是文档报告，不是直接交给开发 Team
- **多角色测试** - 新用户/高级用户/边界用户并行测试

## 角色索引

| 角色 | 文件 | 职责 |
|------|------|------|
| **Lead** | `agents/exp/lead.md` | 协调部署、组织测试、综合反馈、生成报告 |
| **Deployer** | `agents/exp/deployer.md` | 构建服务、启动服务、健康检查 |
| **User Simulator** | `agents/exp/user-simulator.md` | 模拟用户测试、记录体验、产出反馈 |

## 测试目标

| 目标 | 说明 | 端口 | 依赖 |
|------|------|------|------|
| `client` | MCP Client | 8080 | Redis |
| `server` | MCP Server | 8081 | PostgreSQL |
| `integrated` | Client + Server | 8080, 8081 | Redis, PostgreSQL |

## 模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `start` | 部署并完整测试 | `/exp-team start client` |
| `deploy` | 仅部署服务 | `/exp-team deploy server` |
| `test` | 仅执行体验测试 | `/exp-team test integrated` |
| `feedback` | 生成反馈报告 | `/exp-team feedback` |
| `stop` | 停止服务 | `/exp-team stop client` |

### 使用示例

```bash
/exp-team start client      # Client 体验测试
/exp-team start server      # Server 体验测试
/exp-team start integrated  # 集成体验测试
/exp-team deploy server     # 仅部署
/exp-team test integrated   # 仅测试（服务需已部署）
/exp-team stop all         # 停止所有服务
```

## 工作流程

### start 模式

```
1. Deployer 部署服务
2. User Simulator 执行体验测试
3. Lead 综合反馈，生成报告
4. 触发设计 Team 迭代（如需）
```

详见 `agents/exp/lead.md`

## 输出文档

详见 `agents/exp/lead.md`

```
体验报告: docs/体验报告-{日期}.md
```

## 与其他 Team 的关系

```
开发 Team ──> 体验 Team (输出文档报告) ──> 设计 Team
              │
              └─────── 触发迭代（如需）
```

---

**当前任务**: $ARGUMENTS

开始执行体验流程...
