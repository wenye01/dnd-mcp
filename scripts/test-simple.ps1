# DND MCP Client - Simple Test Script
param(
    [switch]$SkipCleanup = $false
)

$BaseUrl = "http://localhost:8080"
$TestSessionId = ""
$PassCount = 0
$FailCount = 0

function Log-Test([string]$msg) {
    Write-Host "`n[TEST] $msg" -ForegroundColor Yellow
}

function Log-Pass([string]$msg) {
    Write-Host "[PASS] $msg" -ForegroundColor Green
    $script:PassCount++
}

function Log-Fail([string]$msg) {
    Write-Host "[FAIL] $msg" -ForegroundColor Red
    $script:FailCount++
}

function Invoke-API([string]$Method, [string]$Url, [object]$Body) {
    $fullUrl = "$BaseUrl$Url"
    $headers = @{"Content-Type" = "application/json"}

    try {
        if ($Body) {
            $json = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Method $Method -Uri $fullUrl -Headers $headers -Body $json -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Method $Method -Uri $fullUrl -Headers $headers -ErrorAction Stop
        }
        return @{Success = $true; Response = $response}
    } catch {
        return @{Success = $false; Error = $_.Exception.Message; StatusCode = $_.Exception.Response.StatusCode.value__}
    }
}

# Prerequisites
Log-Test "Checking Prerequisites"
try {
    $ping = & "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING 2>$null
    if ($ping -eq "PONG") {
        Log-Pass "Redis is running"
    } else {
        Log-Fail "Redis not responding"
        exit 1
    }

    $health = Invoke-API -Method "GET" -Url "/health"
    if ($health.Success -and $health.Response.status -eq "ok") {
        Log-Pass "HTTP Server is running"
    } else {
        Log-Fail "HTTP Server not running"
        exit 1
    }
} catch {
    Log-Fail "Prerequisites check failed: $_"
    exit 1
}

# Test 1: Create Session
Log-Test "TC1: Create Session"
$result = Invoke-API -Method "POST" -Url "/api/sessions" -Body @{
    name = "Test Session"
    creator_id = "user-001"
    mcp_server_url = "http://localhost:9000"
    max_players = 5
}

if ($result.Success) {
    $script:TestSessionId = $result.Response.id
    Log-Pass "Session created - ID: $($result.Response.id)"
} else {
    Log-Fail "Session creation failed: $($result.Error)"
    exit 1
}

# Test 2: List Sessions
Log-Test "TC2: List Sessions"
$result = Invoke-API -Method "GET" -Url "/api/sessions"
if ($result.Success -and $result.Response.Count -gt 0) {
    Log-Pass "Sessions listed - Count: $($result.Response.Count)"
} else {
    Log-Fail "List sessions failed"
}

# Test 3: Get Session
Log-Test "TC3: Get Session"
$result = Invoke-API -Method "GET" -Url "/api/sessions/$TestSessionId"
if ($result.Success -and $result.Response.id -eq $TestSessionId) {
    Log-Pass "Session retrieved"
} else {
    Log-Fail "Get session failed"
}

# Test 4: Update Session
Log-Test "TC4: Update Session"
$result = Invoke-API -Method "PATCH" -Url "/api/sessions/$TestSessionId" -Body @{
    name = "Updated Session Name"
}
if ($result.Success -and $result.Response.name -eq "Updated Session Name") {
    Log-Pass "Session updated"
} else {
    Log-Fail "Update session failed"
}

# Test 5: Send Message
Log-Test "TC5: Send Message"
$result = Invoke-API -Method "POST" -Url "/api/sessions/$TestSessionId/chat" -Body @{
    content = "Hello, this is a test message"
    player_id = "player-001"
}
if ($result.Success -and $result.Response.role -eq "assistant") {
    Log-Pass "Message sent - Response: $($result.Response.content)"
} else {
    Log-Fail "Send message failed"
}

# Test 6: Get Messages
Log-Test "TC6: Get Messages"
$result = Invoke-API -Method "GET" -Url "/api/sessions/$TestSessionId/messages?limit=10"
if ($result.Success -and $result.Response.Count -ge 2) {
    Log-Pass "Messages retrieved - Count: $($result.Response.Count)"
} else {
    Log-Fail "Get messages failed"
}

# Test 7: Broadcast Event
Log-Test "TC7: Broadcast Event"
$result = Invoke-API -Method "POST" -Url "/api/sessions/$TestSessionId/broadcast" -Body @{
    type = "state_changed"
    data = @{
        location = "Test Location"
        game_time = "Day 1"
    }
}
if ($result.Success) {
    Log-Pass "Event broadcasted - ID: $($result.Response.event_id)"
} else {
    Log-Fail "Broadcast failed"
}

# Test 8: Error Handling - Session Not Found
Log-Test "TC8: Error Handling - Session Not Found"
$result = Invoke-API -Method "GET" -Url "/api/sessions/00000000-0000-0000-0000-000000000000"
if (-not $result.Success -and $result.StatusCode -eq 404) {
    Log-Pass "Correct error response"
} else {
    Log-Fail "Error handling failed"
}

# Cleanup
if (-not $SkipCleanup -and $TestSessionId -ne "") {
    Log-Test "Cleanup"
    $result = Invoke-API -Method "DELETE" -Url "/api/sessions/$TestSessionId"
    if ($result.Success -or $result.StatusCode -eq 204) {
        Log-Pass "Test session deleted"
    }
}

# Summary
Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "Test Summary" -ForegroundColor Magenta
Write-Host "========================================" -ForegroundColor Magenta
Write-Host "Passed: $PassCount" -ForegroundColor Green
Write-Host "Failed: $FailCount" -ForegroundColor Red
Write-Host "========================================`n" -ForegroundColor Magenta

if ($FailCount -eq 0) {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "$FailCount test(s) failed" -ForegroundColor Red
    exit 1
}
