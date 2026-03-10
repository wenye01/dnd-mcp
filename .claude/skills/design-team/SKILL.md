---
name: design-team
description: 启动设计 Team，从需求到设计文档和开发计划。支持初始设计、反馈处理、设计调整四种模式。自动识别目标组件（Client/Server）。
argument-hint: <mode> [input]
---

# 设计 Team

组织多角色协同工作，产出高质量设计包。

## 核心原则

详见 `agents/design/lead.md`

- **人类决策优先** - 所有非 Bug 类反馈需要人类决策
- **新增而非修改** - 增量计划独立于现有计划

## 角色索引

| 角色 | 文件 | 职责 |
|------|------|------|
| **Lead** | `agents/design/lead.md` | 任务协调、一致性保证、反馈处理、人类决策请求 |
| **Analyst** | `agents/design/analyst.md` | 需求澄清与分析、目标组件识别 |
| **Architect-System** | `agents/design/architect.md` | 系统架构设计、组件划分、技术选型 |
| **Architect-Domain** | `agents/design/architect.md` | 领域模型设计、D&D 5e 规则实现 |
| **Architect-Integration** | `agents/design/architect.md` | 接口设计、数据流、外部系统集成 |
| **Planner** | `agents/design/planner.md` | 任务分解、里程碑划分、开发计划 |
| **Reviewer** | `agents/design/reviewer.md` | 质量审查、一致性审查 |

## 模式

| 模式 | 输入 | 输出 | 说明 |
|------|------|------|------|
| `init` | 需求描述 | 完整设计包 | 初始设计流程 |
| `feedback` | 反馈报告 | 决策点文档 | 反馈分析，等待人类决策 |
| `implement` | 决策确认 | 增量计划 | 根据人类决策生成计划 |
| `adjust` | 需求描述 | 设计调整方案 | 根据新需求调整/扩展现有设计 |

### 使用示例

```bash
/design-team init 实现一个战斗系统
/design-team feedback docs/体验报告.md
/design-team implement 决策点-骰子.md --decision A
/design-team adjust 添加借机攻击规则
```

## 目标组件识别

| 目标 | 关键词 | 文档路径 |
|------|--------|----------|
| **Server** | 战斗、骰子、角色、怪物、规则、地图 | `docs/server/` |
| **Client** | 会话、对话、消息、LLM、WebSocket | `docs/client/` |

详见 `agents/design/lead.md`

## 工作流程

### init 模式

```
1. Lead 接收需求
2. Analyst 识别目标组件，产出需求规格
3. 3名架构师并行设计 (System/Domain/Integration)
4. Planner 产出开发计划
5. Reviewer 审查
6. 交付审核
```

### feedback 模式

```
1. Lead 接收反馈报告
2. 分类: Bug → 转开发，其他 → 继续
3. Analyst 分析反馈
4. 架构师团队讨论
5. 生成决策点文档
6. 【停止】等待人类决策
```

详见 `agents/design/lead.md` - 反馈处理流程

### implement 模式

```
1. 读取决策点文档
2. 确认人类决策
3. 生成增量设计 + 增量计划
```

### adjust 模式

```
1. Analyst 识别目标组件，分析需求类型
2. 读取现有设计
3. 架构师团队讨论调整方案
4. 生成设计调整文档
5. 【停止】等待人类确认
```

## 输出路径

| 场景 | 输出 | 路径 |
|------|------|------|
| 初始设计 | 需求规格 | `docs/{component}/需求-{主题}.md` |
| 初始设计 | 详细设计 | `docs/{component}/详细设计.md` |
| 初始设计 | 开发计划 | `docs/{component}/plan/M{N}-{主题}.md` |
| 反馈处理 | 决策点文档 | `docs/决策点-{主题}.md` |
| 设计调整 | 调整方案 | `docs/{component}/设计调整-{主题}.md` |

`{component}` = `server` 或 `client`

---

**当前任务**: $ARGUMENTS

开始执行设计流程...
