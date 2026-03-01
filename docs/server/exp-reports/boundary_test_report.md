# M6 地图系统边界用户视角测试报告

**测试人员**: User Simulator C (边界用户视角)
**测试日期**: 2026-03-01
**Server 地址**: http://localhost:8080
**测试 Campaign ID**: `b3447d41-6bc8-4f62-b431-0e20571da635`

---

## 测试场景概览

### 场景 1: 非法输入测试
- 空字符串作为 campaign_id
- 不存在的 ID 查询
- 非法坐标（负数、超大值）
- 非法速度参数

### 场景 2: 边界条件测试
- 坐标边界（0, 0）和最大值
- 超长名称和描述
- 特殊字符输入

### 场景 3: 状态冲突测试
- 在世界地图状态下尝试战斗地图操作
- 在没有战役的情况下调用地图工具
- 重复创建相同资源

### 场景 4: 错误恢复测试
- 观察错误响应格式
- 检查错误后系统是否稳定
- 验证后续操作是否正常

---

## 测试执行结果

### 场景 1: 非法输入测试

#### 1.1 空字符串作为 campaign_id
**测试**: `get_world_map` with empty campaign_id
```json
{"name":"get_world_map","arguments":{"campaign_id":""}}
```
**结果**: ✅ PASS - 正确拒绝并返回错误
```json
{"content":[{"type":"text","text":"campaign ID is required"}],"isError":true}
```

#### 1.2 不存在的 ID 查询
**测试**: `get_world_map` with non-existent campaign_id
```json
{"name":"get_world_map","arguments":{"campaign_id":"non-existent-id-12345"}}
```
**结果**: ⚠️ PARTIAL - 拒绝了请求但错误信息暴露了数据库内部错误
```json
{"content":[{"type":"text","text":"failed to get campaign: failed to scan campaign: 错误: 无效的类型 uuid 输入语法: \"non-existent-id-12345\" (SQLSTATE 22P02)"}],"isError":true}
```
**问题**: 错误信息泄露了SQLSTATE和内部实现细节，应使用更通用的错误消息。

#### 1.3 负数坐标
**测试**: `move_to` with negative coordinates
```json
{"name":"move_to","arguments":{"campaign_id":"...","x":-100,"y":-50}}
```
**结果**: ⚠️ UNKNOWN - 请求被接受但未返回错误（可能因为世界地图不存在）

#### 1.4 极大坐标值
**测试**: `move_to` with extremely large coordinates
```json
{"name":"move_to","arguments":{"campaign_id":"...","x":999999,"y":999999}}
```
**结果**: ⚠️ UNKNOWN - 未验证坐标上限

#### 1.5 负数速度参数
**测试**: `move_token` with negative speed
```json
{"name":"move_token","arguments":{"speed":-10,...}}
```
**结果**: ⚠️ UNKNOWN - 负速度未被验证

---

### 场景 2: 边界条件测试

#### 2.1 归一化坐标边界验证
**测试**: `create_visual_location` with position_x outside [0,1]
```json
{"name":"create_visual_location","arguments":{"position_x":-0.5,...}}
{"name":"create_visual_location","arguments":{"position_x":1.5,...}}
```
**结果**: ✅ PASS - 正确验证并拒绝
```json
{"content":[{"type":"text","text":"position_x must be between 0 and 1"}],"isError":true}
```

#### 2.2 空名称验证
**测试**: `create_visual_location` with empty name
```json
{"name":"create_visual_location","arguments":{"name":"","type":"town",...}}
```
**结果**: ✅ PASS - 正确验证并拒绝
```json
{"content":[{"type":"text","text":"name is required"}],"isError":true}
```

#### 2.3 边界坐标值 (0,0) 和 (1,1)
**测试**: `create_visual_location` with boundary coordinates
```json
{"name":"create_visual_location","arguments":{"position_x":0,"position_y":0,...}}
{"name":"create_visual_location","arguments":{"position_x":1,"position_y":1,...}}
```
**结果**: ✅ PASS - 边界值被正确接受

#### 2.4 超长名称测试
**测试**: `create_visual_location` with extremely long name (300+ characters)
**结果**: ⚠️ UNKNOWN - 超长名称被接受，未验证长度限制

#### 2.5 特殊字符和XSS注入
**测试**: `create_visual_location` with XSS payload
```json
{"name":"create_visual_location","arguments":{"name":"Test<script>alert(\"XSS\")</script>",...}}
```
**结果**: ⚠️ ACCEPTED - XSS payload被接受，应在前端显示时进行转义

#### 2.6 中文字符支持
**测试**: `create_visual_location` with Chinese characters
```json
{"name":"create_visual_location","arguments":{"name":"测试地点中文",...}}
```
**结果**: ✅ PASS - 中文字符被正确接受

---

### 场景 3: 状态冲突测试

#### 3.1 在非战斗状态下获取战斗地图
**测试**: `get_battle_map` when not in battle
```json
{"name":"get_battle_map","arguments":{"campaign_id":"..."}}
```
**结果**: ✅ PASS - 正确检测并返回错误
```json
{"content":[{"type":"text","text":"not currently in a battle map"}],"isError":true}
```

#### 3.2 在非战斗状态下退出战斗地图
**测试**: `exit_battle_map` when not in battle
```json
{"name":"exit_battle_map","arguments":{"campaign_id":"..."}}
```
**结果**: ✅ PASS - 正确检测并返回错误
```json
{"content":[{"type":"text","text":"not currently in a battle map"}],"isError":true}
```

#### 3.3 进入不存在的位置
**测试**: `enter_battle_map` with non-existent location_id
```json
{"name":"enter_battle_map","arguments":{"location_id":"non-existent-location",...}}
```
**结果**: ⚠️ PARTIAL - 返回了错误但错误消息较为通用
```json
{"content":[{"type":"text","text":"failed to get world map: record not found"}],"isError":true}
```

#### 3.4 更新不存在的位置
**测试**: `update_location` with non-existent location_id
```json
{"name":"update_location","arguments":{"location_id":"non-existent-location",...}}
```
**结果**: ⚠️ UNKNOWN - 请求被静默接受（未返回响应）

---

### 场景 4: 错误恢复测试

#### 4.1 错误响应格式一致性
**观察**: 所有错误响应都遵循MCP标准格式
```json
{"content":[{"type":"text","text":"error message"}],"isError":true}
```
**结果**: ✅ PASS - 错误格式统一且符合MCP规范

#### 4.2 错误后系统稳定性
**测试**: 连续发送多个无效请求后验证正常操作
**结果**: ✅ PASS - 系统在错误后继续正常运行
- `get_world_map` 在错误后正常工作
- `get_campaign` 返回正确的数据
- 服务器健康检查返回 `{"status":"healthy"}`

#### 4.3 并发请求处理
**测试**: 同时发送5个相同请求
**结果**: ✅ PASS - 所有5个请求都返回200状态码，系统正确处理并发

#### 4.4 幂等性测试
**测试**: 发送相同的`move_to`请求两次
**结果**: ✅ PASS - 相同请求被正确处理，没有副作用

---

## 安全测试

#### SQL注入测试
**测试**: `create_visual_location` with SQL injection payload
```json
{"name":"create_visual_location","arguments":{"name":"TestLocation\"DROP TABLE campaigns;--",...}}
```
**结果**: ✅ PASS - 使用参数化查询，SQL注入被防御

#### JSON注入测试
**测试**: `create_visual_location` with nested JSON in name
```json
{"name":"create_visual_location","arguments":{"name":"{\"nested\":\"json\"}",...}}
```
**结果**: ✅ PASS - JSON被正确转义为字符串

#### Null字节注入
**测试**: `create_visual_location` with null byte in name
```json
{"name":"create_visual_location","arguments":{"name":"\u0000",...}}
```
**结果**: ✅ PASS - 请求被拒绝，名称被视为空

#### 格式错误的JSON
**测试**: 发送无效的JSON请求体
```json
{invalid json}
```
**结果**: ✅ PASS - 正确拒绝并返回JSON解析错误
```json
{"error":"invalid character 'i' looking for beginning of object key string"}
```

---

## 必需参数验证测试

#### move_token 参数验证
**测试**: 各种空参数组合
| 参数 | 测试值 | 结果 |
|------|--------|------|
| campaign_id | `""` | ✅ "campaign ID is required" |
| map_id | `""` | ✅ "map ID is required" |
| token_id | `""` | ✅ "token ID is required" |

---

## 发现的问题

### 严重问题 (无)
- 无严重问题发现

### 中等问题

1. **BUG-M6-BOUNDARY-001**: UUID格式错误时暴露数据库内部错误
   - **位置**: `get_world_map`工具
   - **问题**: 错误信息包含SQLSTATE和内部实现细节
   - **建议**: 使用UUID验证中间件，返回更通用的"invalid UUID format"错误

2. **BUG-M6-BOUNDARY-002**: 缺少坐标上限验证
   - **位置**: `move_to`工具
   - **问题**: 未验证坐标是否有合理的上限
   - **建议**: 添加坐标范围验证（基于地图尺寸）

3. **BUG-M6-BOUNDARY-003**: 缺少字符串长度验证
   - **位置**: `create_visual_location`工具
   - **问题**: 超长名称（300+字符）被接受
   - **建议**: 添加名称长度限制（如255字符）

### 轻微问题

1. **BUG-M6-BOUNDARY-004**: XSS负载未被过滤
   - **位置**: `create_visual_location`工具
   - **问题**: `<script>`标签被存储，依赖前端转义
   - **建议**: 在服务端进行输入清理或HTML转义

2. **BUG-M6-BOUNDARY-005**: 负速度参数未验证
   - **位置**: `move_token`工具
   - **问题**: 负数速度可能通过验证
   - **建议**: 添加速度参数验证（必须为正数）

---

## 评分

| 评估维度 | 得分 | 说明 |
|----------|------|------|
| 错误处理 | 4/5 | 错误被正确捕获，但部分错误信息过于详细 |
| 系统稳定性 | 5/5 | 错误后系统完全稳定，并发处理良好 |
| 边界处理 | 4/5 | 大部分边界条件被妥善处理 |
| 安全性 | 4/5 | SQL注入防御良好，但XSS防护依赖前端 |
| 参数验证 | 4/5 | 必需参数验证完善，可选参数验证可改进 |

**总体评分**: 4.2/5

---

## 修复建议优先级

### 高优先级
1. 为UUID格式验证添加中间件，避免暴露数据库内部错误
2. 为`move_to`添加坐标范围验证
3. 为字符串字段添加长度限制

### 中优先级
4. 为速度参数添加正数验证
5. 统一错误消息格式，避免暴露实现细节

### 低优先级
6. 考虑添加输入清理层处理特殊字符
7. 为前端提供HTML转义的建议或API

---

## 结论

M6地图系统在边界用户视角下的测试表现良好。系统展现了优秀的稳定性和错误恢复能力，大部分边界条件得到妥善处理。主要改进空间在于：
1. 更严格的参数验证（特别是数值范围和字符串长度）
2. 更友好的错误消息（避免暴露内部实现）
3. 更完善的安全防护（服务端XSS防护）

系统已准备好进行正常使用，但建议在正式发布前修复高优先级问题。
