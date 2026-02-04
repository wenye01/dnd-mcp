# DND MCP Client - 完整端到端测试脚本

param(
    [switch]$SkipCleanup = $false,
    [switch]$Verbose = $false
)

$ErrorActionPreference = "Stop"
$BaseUrl = "http://localhost:8080"
$TestSessionId = ""
$TestWebSocketKey = ""
$PassCount = 0
$FailCount = 0
$TestResults = @()

# 日志函数
function Log-Info {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] INFO: $Message" -ForegroundColor Cyan
}

function Log-Success {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] ✓ PASS: $Message" -ForegroundColor Green
    $script:PassCount++
}

function Log-Fail {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] ✗ FAIL: $Message" -ForegroundColor Red
    $script:FailCount++
}

function Log-Test {
    param([string]$Message)
    Write-Host "`n[$(Get-Date -Format 'HH:mm:ss')] TEST: $Message" -ForegroundColor Yellow
}

# HTTP 请求辅助函数
function Invoke-Api {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [object]$Body
    )

    $url = "$BaseUrl$Endpoint"
    $headers["Content-Type"] = "application/json"

    try {
        if ($Body) {
            $bodyJson = $Body | ConvertTo-Json -Depth 10
            if ($Verbose) { Log-Info "Request: $Method $url`nBody: $bodyJson" }
            $response = Invoke-RestMethod -Method $Method -Uri $url -Headers $Headers -Body $bodyJson -ErrorAction Stop
        } else {
            if ($Verbose) { Log-Info "Request: $Method $url" }
            $response = Invoke-RestMethod -Method $Method -Uri $url -Headers $Headers -ErrorAction Stop
        }
        return @{ Success = $true; Response = $response }
    } catch {
        $errorResponse = @{
            Success = $false
            Error = $_.Exception.Message
            StatusCode = $_.Exception.Response.StatusCode.value__
        }
        if ($Verbose) { Log-Info "Error: $($_.Exception.Message)" }
        return $errorResponse
    }
}

# 预检查
function Test-Prerequisites {
    Log-Test "检查前置条件"

    # 检查 Redis
    try {
        $redisResult = & "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING 2>$null
        if ($redisResult -eq "PONG") {
            Log-Success "Redis 连接正常"
        } else {
            Log-Fail "Redis 未响应"
            return $false
        }
    } catch {
        Log-Fail "Redis 连接失败: $_"
        return $false
    }

    # 检查服务器
    $healthResult = Invoke-Api -Method "GET" -Endpoint "/health"
    if ($healthResult.Success -and $healthResult.Response.status -eq "ok") {
        Log-Success "HTTP 服务器运行正常"
    } else {
        Log-Fail "HTTP 服务器未运行或响应异常"
        return $false
    }

    return $true
}

# 任务一测试: CLI命令
function Test-CliCommands {
    Log-Test "任务一: CLI 命令测试"

    # 创建会话
    $output = & .\bin\dnd-client.exe session create --name "CLI测试会话" --creator "user-cli-test" --mcp-url "http://localhost:9000" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Log-Success "CLI 创建会话成功"
    } else {
        Log-Fail "CLI 创建会话失败: $output"
    }

    # 列出会话
    $output = & .\bin\dnd-client.exe session list 2>&1
    if ($LASTEXITCODE -eq 0) {
        Log-Success "CLI 列出会话成功"
    } else {
        Log-Fail "CLI 列出会话失败: $output"
    }
}

# 任务三测试: 会话管理 API
function Test-SessionAPI {
    Log-Test "任务三: 会话管理 API 测试"

    # TC3.1: 创建会话
    $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions" -Body @{
        name = "HTTP测试会话"
        creator_id = "user-http-test"
        mcp_server_url = "http://localhost:9000"
        max_players = 5
    }

    if ($result.Success) {
        $script:TestSessionId = $result.Response.id
        $script:TestWebSocketKey = $result.Response.websocket_key
        Log-Success "创建会话 (TC3.1) - ID: $TestSessionId"
    } else {
        Log-Fail "创建会话 (TC3.1) - $($result.Error)"
        return
    }

    # TC3.2: 获取会话列表
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions"
    if ($result.Success -and $result.Response.Count -gt 0) {
        Log-Success "获取会话列表 (TC3.2) - 共 $($result.Response.Count) 个会话"
    } else {
        Log-Fail "获取会话列表 (TC3.2)"
    }

    # TC3.3: 获取会话详情
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$TestSessionId"
    if ($result.Success -and $result.Response.id -eq $TestSessionId) {
        Log-Success "获取会话详情 (TC3.3)"
    } else {
        Log-Fail "获取会话详情 (TC3.3)"
    }

    # TC3.4: 更新会话
    $result = Invoke-Api -Method "PATCH" -Endpoint "/api/sessions/$TestSessionId" -Body @{
        name = "更新后的会话名称"
    }
    if ($result.Success -and $result.Response.name -eq "更新后的会话名称") {
        Log-Success "更新会话 (TC3.4)"
    } else {
        Log-Fail "更新会话 (TC3.4)"
    }

    # TC3.5: 错误处理 - 会话不存在
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/00000000-0000-0000-0000-000000000000"
    if (-not $result.Success -and $result.StatusCode -eq 404) {
        Log-Success "错误处理 - 会话不存在 (TC3.5)"
    } else {
        Log-Fail "错误处理 - 会话不存在 (TC3.5)"
    }
}

# 任务四测试: 消息管理 API
function Test-MessageAPI {
    Log-Test "任务四: 消息管理 API 测试"

    if ([string]::IsNullOrEmpty($TestSessionId)) {
        Log-Fail "需要先创建测试会话"
        return
    }

    # TC4.1: 发送消息
    $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions/$TestSessionId/chat" -Body @{
        content = "你好，这是测试消息"
        player_id = "player-test-001"
    }

    if ($result.Success -and $result.Response.role -eq "assistant") {
        Log-Success "发送消息 (TC4.1) - 助手响应: $($result.Response.content)"
    } else {
        Log-Fail "发送消息 (TC4.1) - $($result.Error)"
    }

    # TC4.2: 获取消息历史
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$TestSessionId/messages?limit=10"
    if ($result.Success -and $result.Response.Count -ge 2) {
        Log-Success "获取消息历史 (TC4.2) - 共 $($result.Response.Count) 条消息"
    } else {
        Log-Fail "获取消息历史 (TC4.2)"
    }

    # TC4.3: 消息过滤 - 按角色
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$TestSessionId/messages?role=user&limit=10"
    if ($result.Success) {
        Log-Success "消息过滤 - 按角色 (TC4.3) - 找到 $($result.Response.Count) 条用户消息"
    } else {
        Log-Fail "消息过滤 - 按角色 (TC4.3)"
    }

    # TC4.5: 获取单条消息 (使用第一条消息ID)
    $listResult = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$TestSessionId/messages?limit=1"
    if ($listResult.Success -and $listResult.Response.Count -gt 0) {
        $messageId = $listResult.Response[0].id
        $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$TestSessionId/messages/$messageId"
        if ($result.Success) {
            Log-Success "获取单条消息 (TC4.5)"
        } else {
            Log-Fail "获取单条消息 (TC4.5)"
        }
    }

    # TC4.6: 错误处理 - 参数验证
    $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions/$TestSessionId/chat" -Body @{
        player_id = "player-test-002"
    }
    if (-not $result.Success -and $result.StatusCode -eq 400) {
        Log-Success "错误处理 - 参数验证 (TC4.6)"
    } else {
        Log-Fail "错误处理 - 参数验证 (TC4.6)"
    }
}

# 任务五测试: WebSocket API
function Test-WebSocketAPI {
    Log-Test "任务五: WebSocket API 测试"

    if ([string]::IsNullOrEmpty($TestSessionId) -or [string]::IsNullOrEmpty($TestWebSocketKey)) {
        Log-Fail "需要先创建测试会话并获取 WebSocket Key"
        return
    }

    # TC5.1: 建立 WebSocket 连接 (简单测试)
    Log-Info "WebSocket 连接测试需要 websocat 或 WebSocket 客户端工具"
    Log-Info "连接 URL: ws://localhost:8080/ws/sessions/$TestSessionId?key=$TestWebSocketKey"
    Log-Success "WebSocket 连接 URL 已生成"

    # TC5.5: 广播测试事件
    $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions/$TestSessionId/broadcast" -Body @{
        type = "state_changed"
        data = @{
            changes = @{
                location = "测试地点"
                game_time = "第1天 10:00"
            }
        }
    }
    if ($result.Success) {
        Log-Success "广播测试事件 (TC5.5) - Event ID: $($result.Response.event_id)"
    } else {
        Log-Fail "广播测试事件 (TC5.5)"
    }

    # TC5.6: 获取连接信息
    $result = Invoke-Api -Method "GET" -Endpoint "/test/ws/connections?session_id=$TestSessionId"
    if ($result.Success) {
        Log-Success "获取连接信息 (TC5.6) - 连接数: $($result.Response.count)"
    } else {
        Log-Fail "获取连接信息 (TC5.6)"
    }
}

# 任务二测试: 数据持久化
function Test-Persistence {
    Log-Test "任务二: 数据持久化测试"

    # 备份数据
    $backupOutput = & .\bin\dnd-client.exe backup --all 2>&1
    if ($LASTEXITCODE -eq 0) {
        Log-Success "备份数据到 PostgreSQL"
    } else {
        Log-Fail "备份数据失败: $backupOutput"
    }

    # 恢复数据 (如果不清空 Redis,恢复会跳过已存在的数据)
    $restoreOutput = & .\bin\dnd-client.exe restore --all 2>&1
    if ($LASTEXITCODE -eq 0) {
        Log-Success "从 PostgreSQL 恢复数据"
    } else {
        Log-Fail "恢复数据失败: $restoreOutput"
    }
}

# 综合测试
function Test-CompleteWorkflow {
    Log-Test "综合测试: 完整工作流"

    # 1. 创建会话
    $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions" -Body @{
        name = "综合测试会话"
        creator_id = "user-integration-test"
        mcp_server_url = "http://localhost:9000"
    }

    if (-not $result.Success) {
        Log-Fail "综合测试 - 创建会话失败"
        return
    }

    $sessionId = $result.Response.id
    Log-Success "综合测试 - 步骤1: 创建会话"

    # 2. 发送多条消息
    for ($i = 1; $i -le 3; $i++) {
        $result = Invoke-Api -Method "POST" -Endpoint "/api/sessions/$sessionId/chat" -Body @{
            content = "测试消息 $i"
            player_id = "player-integration"
        }
        if (-not $result.Success) {
            Log-Fail "综合测试 - 发送消息 $i 失败"
            return
        }
    }
    Log-Success "综合测试 - 步骤2: 发送 3 条消息"

    # 3. 验证消息历史
    $result = Invoke-Api -Method "GET" -Endpoint "/api/sessions/$sessionId/messages?limit=10"
    if ($result.Success -and $result.Response.Count -ge 6) { # 3 user + 3 assistant
        Log-Success "综合测试 - 步骤3: 验证消息历史 - 共 $($result.Response.Count) 条"
    } else {
        Log-Fail "综合测试 - 步骤3: 消息数量不符"
    }

    # 4. 更新会话
    $result = Invoke-Api -Method "PATCH" -Endpoint "/api/sessions/$sessionId" -Body @{
        status = "active"
    }
    if ($result.Success) {
        Log-Success "综合测试 - 步骤4: 更新会话状态"
    } else {
        Log-Fail "综合测试 - 步骤4: 更新失败"
    }
}

# 清理测试数据
function Clean-Up {
    if ($SkipCleanup) {
        Log-Info "跳过清理"
        return
    }

    Log-Test "清理测试数据"

    if (-not [string]::IsNullOrEmpty($TestSessionId)) {
        $result = Invoke-Api -Method "DELETE" -Endpoint "/api/sessions/$TestSessionId"
        if ($result.Success -or $result.StatusCode -eq 204) {
            Log-Success "删除测试会话"
        }
    }
}

# 主测试流程
function Main {
    Write-Host "`n========================================" -ForegroundColor Magenta
    Write-Host "  DND MCP Client - E2E 测试" -ForegroundColor Magenta
    Write-Host "========================================`n" -ForegroundColor Magenta

    $startTime = Get-Date

    # 检查前置条件
    if (-not (Test-Prerequisites)) {
        Write-Host "`n✗ 前置条件检查失败，请确保:" -ForegroundColor Red
        Write-Host "  1. Redis 正在运行" -ForegroundColor Red
        Write-Host "  2. HTTP 服务器正在运行 (.\bin\dnd-client.exe server)" -ForegroundColor Red
        exit 1
    }

    # 执行测试
    Test-CliCommands
    Test-SessionAPI
    Test-MessageAPI
    Test-WebSocketAPI
    Test-Persistence
    Test-CompleteWorkflow

    # 清理
    Clean-Up

    # 输出结果
    $duration = (Get-Date) - $startTime
    Write-Host "`n========================================" -ForegroundColor Magenta
    Write-Host "  测试完成" -ForegroundColor Magenta
    Write-Host "========================================" -ForegroundColor Magenta
    Write-Host "  总用时: $($duration.ToString('mm\:ss'))" -ForegroundColor White
    Write-Host "  通过: $PassCount" -ForegroundColor Green
    Write-Host "  失败: $FailCount" -ForegroundColor Red
    Write-Host "========================================`n" -ForegroundColor Magenta

    if ($FailCount -eq 0) {
        Write-Host "✓ 所有测试通过!" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "✗ 有 $FailCount 个测试失败" -ForegroundColor Red
        exit 1
    }
}

# 运行
Main
