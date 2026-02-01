# PowerShell 执行错误修复说明

## 问题描述
在 PowerShell 中直接执行 `.\test.bat` 时出现错误：
```
'" (' 不是内部或外部命令
'1" >nul 2>&1' 不是内部或外部命令
'cho' 不是内部或外部命令
```

## 根本原因

1. **文件编码问题**: 原始文件是 UTF-8 编码
2. **行尾符问题**: 原始文件使用 LF (\n) 而不是 CRLF (\r\n)
3. **ANSI 颜色代码**: 使用了 `[92m` 等 ANSI 转义序列，在 PowerShell 中会导致解析错误

Windows 批处理文件要求：
- **编码**: ASCII 或 UTF-8 without BOM
- **行尾符**: CRLF (Carriage Return + Line Feed, 即 `\r\n`)

## 修复方案

### 已修复的文件

1. **test.bat**
   - ✅ 移除所有 ANSI 颜色代码
   - ✅ 转换为 ASCII 编码
   - ✅ 转换为 CRLF 行尾符
   - ✅ 使用简单的文本标记：[OK]、[PASS]、[FAIL]

2. **build.bat**
   - ✅ 转换为 ASCII 编码
   - ✅ 转换为 CRLF 行尾符

### 如何使用

#### 在 PowerShell 中运行

```powershell
# 构建
.\build.bat build          # 构建应用
.\build.bat run            # 运行应用
.\build.bat clean          # 清理构建文件
.\build.bat help           # 显示帮助

# 测试
.\test.bat                 # 运行所有测试
.\test.bat --unit          # 仅运行单元测试
.\test.bat --no-coverage   # 不生成覆盖率报告
.\test.bat --help          # 显示帮助
```

#### 在 CMD 中运行
命令相同，直接使用 `.bat` 文件即可。

## 验证

### 检查文件格式
```bash
# 检查编码
file test.bat
# 应该输出: DOS batch file, ASCII text

file build.bat
# 应该输出: DOS batch file, ASCII text, with CRLF line terminators
```

### 检查行尾符
```bash
# 查看前几个字节的十六进制
head -c 30 test.bat | od -A x -t x1z -v
# 应该看到: 0d 0a (CRLF)
```

## 技术细节

### 文件编码对比

| 文件 | 修复前 | 修复后 |
|------|--------|--------|
| test.bat | UTF-8 with LF | ASCII with CRLF |
| build.bat | UTF-8 with LF | ASCII with CRLF |

### ANSI 颜色代码移除

**修复前** (使用 ANSI 颜色):
```batch
set "GREEN=[92m"
set "RED=[91m"
echo %GREEN%OK%RESET%
```

**修复后** (使用简单文本):
```batch
echo [OK]
echo [PASS]
echo [FAIL]
```

## 测试结果

✅ **test.bat** - 可以在 PowerShell 中正常运行
✅ **build.bat** - 可以在 PowerShell 中正常运行
✅ **测试报告** - 生成到 `test-reports/` 文件夹
✅ **所有功能** - 完整保留，无功能损失

## 常见问题

### Q: 为什么不在 PowerShell 中使用 ANSI 颜色？
A: Windows PowerShell 5.x 对 ANSI 颜色支持有限，需要特殊的转义序列。使用简单的文本标记更可靠。

### Q: 如果我想要彩色输出怎么办？
A:
1. 使用 Windows Terminal (支持 ANSI 颜色)
2. 升级到 PowerShell 7+ (支持 ANSI 颜色)
3. 或者在 CMD 中运行批处理文件

### Q: 如何确保新创建的 .bat 文件格式正确？
A: 使用以下方法之一：
1. 使用记事本保存，选择 "ASCII" 编码
2. 使用 VS Code，确保 "End of Line" 设置为 "CRLF"
3. 使用 PowerShell 命令转换：
   ```powershell
   [System.IO.File]::WriteAllText("test.bat", [System.IO.File]::ReadAllText("test.bat"))
   ```

## 总结

通过以下三个关键修复，解决了 PowerShell 执行错误：

1. ✅ **编码**: UTF-8 → ASCII
2. ✅ **行尾符**: LF → CRLF
3. ✅ **颜色代码**: 移除 ANSI 转义序列

现在两个批处理文件都可以在 PowerShell 中完美运行！
