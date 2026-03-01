---
name: exp-team
description: 启动体验 Team，部署服务并进行用户体验测试，生成反馈报告。
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
argument-hint: <mode> <target> [options]
---

# 体验 Team

组织 Deployer 和 User Simulators 协同工作，部署服务、体验测试、生成反馈报告。

## 测试目标

| 目标 | 说明 | 默认端口 | 依赖 |
|------|------|---------|------|
| `client` | 仅测试 MCP Client | 8080 | Redis |
| `server` | 仅测试 MCP Server | 8081 | Redis, PostgreSQL |
| `integrated` | 测试 Client + Server 集成 | 8080, 8081 | Redis, PostgreSQL |

## 模式

| 模式 | 说明 | 示例 |
|------|------|------|
| `start` | 部署指定目标并完整测试 | `/exp-team start client` |
| `deploy` | 仅部署指定目标服务 | `/exp-team deploy server` |
| `test` | 仅执行体验测试（服务需已部署） | `/exp-team test integrated` |
| `feedback` | 生成反馈报告 | `/exp-team feedback` |
| `stop` | 停止指定目标服务 | `/exp-team stop client` |

## 使用示例

```
# Client 体验测试（会话管理、对话历史、前端 API）
/exp-team start client

# Server 体验测试（游戏规则、状态管理、地图系统）
/exp-team start server

# 集成体验测试（完整 D&D 游戏流程）
/exp-team start integrated

# 仅部署 Server（用于手动测试）
/exp-team deploy server

# 集成环境已部署，执行完整体验测试
/exp-team test integrated

# 生成并提交反馈报告
/exp-team feedback

# 停止所有服务
/exp-team stop all
```

## Team 结构

```
┌─────────────────────────────────────────┐
│           Exp Team Lead                  │
│         (协调 + 反馈综合)                 │
└───────────────┬─────────────────────────┘
                │
    ┌───────────┼───────────┬───────────┐
    │           │           │           │
    ▼           ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│Deployer │ │  User   │ │  User   │ │  User   │
│         │ │Sim A    │ │Sim B    │ │Sim C    │
│         │ │(新用户) │ │(高级)   │ │(边界)   │
└─────────┘ └─────────┘ └─────────┘ └─────────┘
```

## 用户角色说明

| 角色 | 关注点 | 测试场景 |
|------|--------|---------|
| **新用户** | 首次使用体验、引导清晰度、错误提示 | 基础功能、简单流程 |
| **高级用户** | 效率、高级功能、性能 | 复杂场景、并发操作 |
| **边界用户** | 异常处理、边界条件、稳定性 | 极端输入、异常操作 |

## 执行流程

### Step 1: 部署服务

```
Spawn Deployer teammate:
1. 加载 .claude/agents/exp/deployer.md
2. 检查环境（Redis、端口等）
3. 构建二进制文件
4. 启动服务
5. 健康检查
6. 功能验证
7. 报告部署状态
```

### Step 2: 体验测试

```
并行 Spawn User Simulator teammates:

1. User Simulator A (新用户视角)
   - 加载 .claude/agents/exp/user-simulator.md
   - 执行基础场景测试
   - 记录首次使用体验
   - 评分和反馈

2. User Simulator B (高级用户视角)
   - 执行复杂场景测试
   - 测试性能和效率
   - 评分和反馈

3. User Simulator C (边界用户视角)
   - 执行边界场景测试
   - 测试异常处理
   - 评分和反馈
```

### Step 3: 综合反馈

```
1. 收集所有 User Simulators 的报告
2. 整合问题：
   - 去重
   - 分类（Bug/体验/新需求）
   - 优先级排序
3. 综合建议
4. 生成最终体验报告
```

### Step 4: 触发迭代

```
如果需要迭代：
1. 将体验报告传递给设计 Team
2. 执行 /design-team iterate [报告文件]

否则：
1. 标记版本可发布
2. 清理测试环境
```

## 测试场景

### 按目标分类的场景

#### Client 测试场景

```markdown
核心功能:
1. 会话管理
   - 创建会话 (POST /api/sessions)
   - 查询会话列表 (GET /api/sessions)
   - 获取会话详情 (GET /api/sessions/:id)
   - 更新会话 (PATCH /api/sessions/:id)
   - 删除会话 (DELETE /api/sessions/:id)

2. 消息管理
   - 发送消息 (POST /api/sessions/:id/chat)
   - 获取消息历史 (GET /api/sessions/:id/messages)
   - 分页查询
   - WebSocket 实时消息

3. 系统监控
   - 健康检查 (GET /api/system/health)
   - 系统统计 (GET /api/system/stats)

用户角色场景:
- 新用户: 首次创建会话、发送第一条消息
- 高级用户: 批量会话操作、并发消息
- 边界用户: 超长消息、特殊字符、快速重复发送
```

#### Server 测试场景

```markdown
核心功能:
1. 游戏规则
   - 骰子投掷 (POST /api/tools/dice)
   - 属性检定
   - 技能检定

2. 战斗系统
   - 创建战斗 (POST /api/combat)
   - 加入战斗 (POST /api/combat/:id/join)
   - 回合管理 (POST /api/combat/:id/turn)
   - 执行动作 (POST /api/combat/:id/action)
   - 结束战斗 (POST /api/combat/:id/end)

3. 角色管理
   - 创建角色 (POST /api/characters)
   - 更新状态 (PATCH /api/characters/:id)
   - 获取属性 (GET /api/characters/:id)

4. 地图系统
   - 创建地图 (POST /api/maps)
   - 更新位置 (PATCH /api/maps/:id/position)
   - 视野计算 (GET /api/maps/:id/visibility)

用户角色场景:
- 新用户: 创建角色、基础骰子投掷
- 高级用户: 完整战斗流程、复杂规则组合
- 边界用户: 极端属性值、非法战斗动作
```

#### Integrated 测试场景

```markdown
端到端流程:
1. 完整游戏流程
   - Client 创建会话 → Server 创建角色
   - Client 发送消息 → Server 处理游戏逻辑
   - Server 返回结果 → Client 更新对话历史
   - WebSocket 实时同步

2. 战斗场景
   - DM 发起战斗 → Server 创建战斗实例
   - 玩家加入 → Server 处理先攻
   - 回合执行 → Server 验证规则
   - 结果同步 → 所有客户端收到更新

3. 跨服务交互
   - Client LLM 调用 → Server 规则验证
   - Server 状态变更 → Client 通知前端
   - 数据一致性验证

用户角色场景:
- 新用户: 完整新手引导流程
- 高级用户: 多角色并发游戏
- 边界用户: 跨服务异常传播
```

### 角色特定场景

```markdown
新用户:
- 首次使用引导
- 错误提示清晰度
- 文档/帮助可用性
- 基础流程完成度

高级用户:
- 批量操作效率
- 并发请求处理
- 性能响应时间
- 高级功能可用性

边界用户:
- 异常输入处理
- 边界条件测试
- 错误恢复能力
- 系统稳定性
```

## 输出格式

### 部署报告

```markdown
## 部署报告 - [target]

### 环境状态
- [ ] Redis 可用
- [ ] PostgreSQL 可用 (仅 server/integrated)
- [ ] 端口空闲 (根据目标列出)
- [ ] 构建成功

### 服务状态
| 服务 | 端口 | 状态 | 健康检查 |
|------|------|:----:|----------|
| Client | 8080 | ✅ | /api/system/health |
| Server | 8081 | ✅ | /api/system/health |

### 可用端点
#### Client 端点 (8080)
| 端点 | 状态 |
|------|:----:|
| /api/system/health | ✅ |
| /api/sessions | ✅ |

#### Server 端点 (8081)
| 端点 | 状态 |
|------|:----:|
| /api/system/health | ✅ |
| /api/tools/dice | ✅ |
| /api/combat | ✅ |
```

### 体验报告

```markdown
## 体验测试总报告 - [target]

### 测试概况
- 测试目标: [client/server/integrated]
- 测试版本: [版本号/Commit]
- 测试时间: [日期]
- 参与角色: 新用户、高级用户、边界用户
- 测试场景数: X 个

### 评分汇总
| 角色 | 评分 | 评语 |
|------|:----:|------|
| 新用户 | 4/5 | 入门简单 |
| 高级用户 | 3/5 | 功能完善 |
| 边界用户 | 3/5 | 大部分处理得当 |
| **平均** | **3.3/5** | |

### 问题汇总

#### Bug (需立即修复)
| ID | 问题 | 严重程度 | 来源 | 组件 |
|----|------|:--------:|------|------|
| B1 | 超长文本崩溃 | P0 | 边界用户 | client |

#### 体验问题 (需设计改进)
| ID | 问题 | 严重程度 | 建议 | 组件 |
|----|------|:--------:|------|------|
| E1 | 错误提示不友好 | P1 | 增加具体提示 | server |

#### 新需求想法
| ID | 需求 | 价值 | 组件 |
|----|------|------|------|
| F1 | 批量导出 | 高 | client |

### 组件评分 (仅 integrated 模式)
| 组件 | 评分 | 主要问题 |
|------|:----:|---------|
| Client | 4/5 | 响应速度可提升 |
| Server | 3/5 | 规则提示不清晰 |
| 集成 | 4/5 | 数据同步良好 |

### 建议优先级
1. [P0] 修复超长文本崩溃问题
2. [P1] 改进错误提示
3. [新需求] 考虑添加批量导出功能

### 迭代建议
- [ ] 需要迭代 (有 P0/P1 问题)
- [ ] 可以发布 (无严重问题)
```

## 与其他 Team 的协作

```
开发 Team ──> 体验 Team ──> 设计 Team
                 │              │
                 └──────────────┘ (反馈循环)
```

## 迭代触发条件

```
触发迭代的条件：
1. 有 P0 级别的 Bug
2. 有 2 个以上 P1 级别问题
3. 平均评分低于 3 分
4. 有高价值新需求

满足任一条件 → 执行 /design-team iterate
```

## 清理流程

测试完成后：

```
1. 停止服务
2. 清理测试数据
3. 清理日志文件
4. 保存体验报告
```

## 注意事项

1. **必须先部署成功**: 服务不可用时不能进行体验测试
2. **多角色并行测试**: 提高效率，获得多角度反馈
3. **反馈必须结构化**: 便于设计 Team 处理
4. **严重问题立即反馈**: 不等待完整测试结束

---

**当前任务**: $ARGUMENTS

开始执行体验流程...
