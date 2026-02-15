---
name: test
description: 执行完整的测试流程。包括编译构建、单元测试、集成测试和端到端测试。会检查现有测试脚本并确保其适配当前代码。
disable-model-invocation: true
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
argument-hint: [test-type|all]
---

# 测试技能

执行完善的项目测试流程，确保代码质量和功能正确性。

## 测试参数

**测试类型**: $ARGUMENTS

可选值：
- `all` 或留空 - 执行完整测试流程
- `build` - 仅编译构建
- `unit` - 仅单元测试
- `integration` - 仅集成测试
- `e2e` - 仅端到端测试

## 测试流程

### 第一步：环境重置（确保干净环境）

**必须先执行，确保测试可在全新环境复现：**

```
1. 清理构建产物
   - 删除 bin/ 目录
   - 删除临时文件
   - Go: go clean -cache -testcache

2. 重置依赖服务
   - Redis: FLUSHALL 或使用独立测试数据库
   - PostgreSQL: 重建测试数据库或运行迁移重置

3. 终止残留进程
   - 检查并终止占用端口的进程（如 8080）
   - 终止僵尸测试服务器进程

4. 重置测试数据
   - 清空测试用 Redis DB（如 DB 1）
   - 重置测试数据库到初始状态

5. 验证环境就绪
   - 确认依赖服务可连接
   - 确认端口未被占用
   - 确认无残留测试数据
```

### 第二步：环境检查

```
1. 检查项目语言和技术栈（查看 go.mod、package.json 等）
2. 检查 scripts/ 目录下的测试脚本
3. 检查 tests/ 目录结构
4. 确认依赖服务状态（Redis、PostgreSQL 等）
```

### 第三步：测试脚本审计

对于发现的每个测试脚本：

```
1. 读取脚本内容
2. 检查是否覆盖以下场景：
   - 正常路径（Happy Path）
   - 边界条件
   - 错误处理
3. 检查是否适配最新代码：
   - API 端点是否匹配当前路由
   - 数据结构是否匹配当前模型
   - 配置是否匹配当前环境
4. 检查是否有环境重置逻辑
5. 如有问题，报告需要更新的内容
```

### 第四步：执行测试

按以下顺序执行：

```
1. 编译构建 (build)
   - Go: go build ./...
   - 检查编译错误

2. 单元测试 (unit)
   - Go: go test -v ./tests/unit/... -cover
   - 报告覆盖率

3. 集成测试 (integration)
   - Go: go test -v ./tests/integration/... -cover -timeout 30s
   - 确保测试数据库隔离

4. 端到端测试 (e2e)
   - 启动测试服务器
   - Go: go test -v ./tests/e2e/... -cover
   - 清理测试环境
```

### 第五步：结果报告

```
## 测试报告

### 环境状态
- [ ] 构建产物已清理
- [ ] 依赖服务已重置
- [ ] 无残留进程

### 构建状态
- [ ] 编译通过
- 错误信息（如有）

### 单元测试
- [ ] 通过 / 失败
- 覆盖率: XX%
- 失败用例列表

### 集成测试
- [ ] 通过 / 失败
- 失败用例列表

### 端到端测试
- [ ] 通过 / 失败
- 失败用例列表

### 建议的改进
- 需要添加的测试用例
- 需要更新的测试脚本
```

## 测试脚本缺失处理

如果缺少必要的测试脚本：

1. **确认项目结构和测试框架**
2. **创建符合项目规范的测试脚本**：
   - 放在 `scripts/` 目录
   - 包含环境重置逻辑
   - 遵循现有脚本风格
   - 包含适当的错误处理
3. **运行新创建的脚本验证**

## 项目特定配置

针对此 Go 项目：

```
环境重置命令:
  PowerShell:
    - Remove-Item -Recurse -Force bin/ -ErrorAction SilentlyContinue
    - go clean -cache -testcache
    - redis-cli FLUSHALL (或 redis-cli -n 1 FLUSHDB 仅清空测试库)

脚本位置: scripts/
  - reset-env.ps1   - 环境重置
  - build.ps1       - 构建脚本
  - test.ps1        - 快速测试
  - test-all.ps1    - 完整测试
  - test-e2e.ps1    - E2E 测试

测试目录: tests/
  - unit/           - 单元测试
  - integration/    - 集成测试
  - e2e/            - 端到端测试

依赖服务:
  - Redis: localhost:6379 (测试用 DB 1)
  - PostgreSQL: localhost:5432 (可选)

端口检查:
  - 8080: API 服务器
```

## 清理操作

测试完成后：
- 清理临时文件
- 重置测试数据库
- 终止测试服务器进程
- 恢复环境到测试前状态

---

开始执行测试流程...
