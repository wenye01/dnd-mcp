# 需求规格: Client-Server 整体协同

## 文档信息

- **版本**: v1.0
- **创建日期**: 2026-03-01
- **目标组件**: Both (Client + Server)
- **状态**: 设计中

---

## 1. 目标组件声明

**目标组件**: Both (Client + Server)

本需求涉及 Client 和 Server 的整体协同工作，需要两个组件配合完成。

---

## 2. 背景与目标

### 2.1 当前状态

| 组件 | 状态 | 说明 |
|------|------|------|
| **MCP Server** | M5 已完成 | 基础设施、战役管理、角色管理、骰子系统、战斗系统 |
| **MCP Client** | 基础功能完成 | 会话管理、对话历史、WebSocket |
| **集成状态** | 未集成 | Client 和 Server 尚未协同工作 |

### 2.2 目标

实现 Client 和 Server 的整体部署和协同工作，使系统能够：
1. 用户通过自然语言与 AI DM 交互
2. AI 调用 Server 工具执行游戏规则
3. 完整的端到端游戏体验

---

## 3. 核心需求

### 3.1 整体部署架构需求

| 需求ID | 需求描述 | 优先级 |
|--------|----------|--------|
| REQ-001 | Client 和 Server 可在同一主机上部署运行 | P0 |
| REQ-002 | 提供统一的环境配置方案 | P0 |
| REQ-003 | 提供启动脚本和部署文档 | P1 |
| REQ-004 | 支持 Redis/PostgreSQL 数据存储 | P0 |

### 3.2 GLM-4.7-Flash 集成需求

| 需求ID | 需求描述 | 优先级 |
|--------|----------|--------|
| REQ-010 | Client 集成 GLM-4.7-Flash 作为 LLM | P0 |
| REQ-011 | 支持工具调用（Function Calling） | P0 |
| REQ-012 | 支持流式输出 | P1 |
| REQ-013 | 支持深度思考模式（thinking） | P2 |

**GLM-4.7-Flash 配置**:
```
API 端点: https://open.bigmodel.cn/api/paas/v4/chat/completions
模型名: glm-4.7-flash
最大 Token: 65536
```

### 3.3 MCP 协议对接需求

| 需求ID | 需求描述 | 优先级 |
|--------|----------|--------|
| REQ-020 | Client 通过 MCP 协议连接 Server | P0 |
| REQ-021 | Client 调用 Server 提供的工具 | P0 |
| REQ-022 | 支持 Server M1-M5 已实现的工具 | P0 |

### 3.4 端到端游戏流程需求

| 需求ID | 需求描述 | 优先级 |
|--------|----------|--------|
| REQ-030 | 用户可创建战役并开始游戏 | P0 |
| REQ-031 | 用户可创建角色 | P0 |
| REQ-032 | 用户可进行战斗 | P0 |
| REQ-033 | 用户可投掷骰子 | P0 |

---

## 4. 场景分析

### 4.1 主要使用场景

#### 场景 1: 创建战役并开始游戏

```
用户 → Client: "创建一个名为'失落矿坑'的战役"
Client → LLM: 发送用户消息 + System Prompt
LLM → Client: 返回 Tool Call: create_campaign
Client → Server: 调用 create_campaign
Server → Client: 返回战役 ID
Client → 用户: "战役'失落矿坑'已创建，你可以开始冒险了"
```

#### 场景 2: 创建角色

```
用户 → Client: "我想创建一个精灵游侠，名字叫 Legolas"
Client → LLM: 发送用户消息 + 上下文
LLM → Client: 返回 Tool Call: create_character
Client → Server: 调用 create_character
Server → Client: 返回角色数据
Client → 用户: "Legolas，精灵游侠，已加入队伍"
```

#### 场景 3: 战斗流程

```
用户 → Client: "我攻击哥布林"
Client → LLM: 发送用户消息 + 战斗上下文
LLM → Client: 返回 Tool Call: roll_attack
Client → Server: 调用 roll_attack
Server → Client: 返回攻击结果（命中/伤害）
Client → 用户: "你射出一箭，命中哥布林造成 8 点伤害"
```

### 4.2 边界场景

| 场景 | 处理方式 |
|------|----------|
| LLM 返回无效 Tool Call | Client 返回错误提示，请求 LLM 重试 |
| Server 工具调用失败 | Client 向用户展示错误信息 |
| LLM API 超时 | Client 返回超时提示，支持重试 |
| Server 未启动 | Client 启动时检查连接，提示用户 |

---

## 5. 非功能性需求

### 5.1 性能需求

| 指标 | 要求 |
|------|------|
| LLM 响应时间 | < 10s（首字节） |
| Server 工具调用 | < 500ms |
| WebSocket 延迟 | < 100ms |

### 5.2 可用性需求

| 指标 | 要求 |
|------|------|
| 服务可用性 | 99%（单机部署） |
| 错误恢复 | 支持断线重连 |

### 5.3 安全性需求

| 需求 | 说明 |
|------|------|
| API Key 保护 | 环境变量存储，不硬编码 |
| 数据隔离 | 不同战役数据隔离 |

---

## 6. 与现有功能的关系

### 6.1 Server M1-M5 功能

| 里程碑 | 功能 | 集成状态 |
|--------|------|----------|
| M1 | 项目基础设施 | 需集成 |
| M2 | 战役管理 | 需集成 |
| M3 | 角色管理 | 需集成 |
| M4 | 骰子系统 | 需集成 |
| M5 | 战斗系统 | 需集成 |

### 6.2 Client 现有功能

| 功能 | 调整需求 |
|------|----------|
| 会话管理 | 需适配 Server 战役 |
| 对话历史 | 通过 Server API 管理 |
| LLM 调用 | 集成 GLM-4.7-Flash |
| WebSocket | 保持现有实现 |

---

## 7. 约束条件

1. **Server 优先**: Server 已完成 M1-M5，Client 需适配 Server
2. **MCP 协议**: Client 通过 MCP 协议调用 Server
3. **GLM-4.7-Flash**: 使用指定的 LLM 模型
4. **单机部署**: 首要支持单机部署场景

---

## 8. 验收标准

- [ ] Client 可启动并连接 Server
- [ ] 用户可通过自然语言创建战役
- [ ] 用户可通过自然语言创建角色
- [ ] 用户可通过自然语言进行战斗
- [ ] 用户可通过自然语言投掷骰子
- [ ] 流式输出正常工作
- [ ] 错误情况有友好提示
