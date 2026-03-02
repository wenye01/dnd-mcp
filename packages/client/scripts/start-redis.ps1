# Start Redis for Development (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

param(
    [string]$Action = "start",
    [switch]$Reset
)

$RedisPath = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-server.exe"
$RedisCliPath = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
$RedisPort = 6379
$ServiceName = "Redis"

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Test-LocalRedis {
    if (Test-Path $RedisPath) {
        return $true
    }
    return $false
}

function Start-LocalRedis {
    Write-Step "Starting local Redis..."

    if (-not (Test-LocalRedis)) {
        Write-Host "Redis not found at: $RedisPath" -ForegroundColor Red
        Write-Host "Please install Redis or update the path in this script" -ForegroundColor Yellow
        exit 1
    }

    $redisProc = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisProc) {
        Write-Host "Redis is already running (PID: $($redisProc.Id))" -ForegroundColor Green
        return $true
    }

    Start-Process -FilePath $RedisPath -ArgumentList "--port $RedisPort" -WindowStyle Hidden
    
    Start-Sleep -Seconds 2

    $redisProc = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisProc) {
        Write-Host "Redis started successfully (PID: $($redisProc.Id))" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "Failed to start Redis" -ForegroundColor Red
        return $false
    }
}

function Stop-LocalRedis {
    Write-Step "Stopping local Redis..."

    $redisProc = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisProc) {
        $redisProc | Stop-Process -Force
        Write-Host "Redis stopped" -ForegroundColor Green
    }
    else {
        Write-Host "Redis is not running" -ForegroundColor Yellow
    }
}

function Test-RedisConnection {
    Write-Step "Testing Redis connection..."

    if (-not (Test-Path $RedisCliPath)) {
        Write-Host "redis-cli not found at: $RedisCliPath" -ForegroundColor Red
        return $false
    }

    Start-Sleep -Seconds 1

    $maxRetries = 5
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        try {
            $result = & $RedisCliPath -p $RedisPort ping 2>$null
            if ($result -eq "PONG") {
                Write-Host "Redis connection successful!" -ForegroundColor Green
                Write-Host "Redis address: localhost:$RedisPort" -ForegroundColor Cyan
                return $true
            }
        }
        catch {
            # Continue retrying
        }

        $retryCount++
        Write-Host "Waiting for Redis to be ready... ($retryCount/$maxRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 1
    }

    Write-Host "Redis connection test failed after $maxRetries attempts" -ForegroundColor Red
    return $false
}

function Show-Status {
    Write-Step "Redis Status"

    $redisProc = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisProc) {
        Write-Host "Status: Running" -ForegroundColor Green
        Write-Host "PID: $($redisProc.Id)" -ForegroundColor Cyan
        Write-Host "Port: $RedisPort" -ForegroundColor Cyan
    }
    else {
        Write-Host "Status: Not Running" -ForegroundColor Red
    }
}

# Main logic
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP Redis Manager (Local)" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

switch ($Action.ToLower()) {
    "start" {
        if (Start-LocalRedis) {
            Test-RedisConnection
            Write-Host ""
            Write-Host "Common commands:" -ForegroundColor Cyan
            Write-Host "  Stop:     .\scripts\start-redis.ps1 -Action stop" -ForegroundColor White
            Write-Host "  Status:   .\scripts\start-redis.ps1 -Action status" -ForegroundColor White
            Write-Host "  Connect:  & `"$RedisCliPath`" -p $RedisPort" -ForegroundColor White
        }
        else {
            Write-Host "Redis startup failed" -ForegroundColor Red
            exit 1
        }
    }
    "stop" {
        Stop-LocalRedis
    }
    "status" {
        Show-Status
    }
    "restart" {
        Stop-LocalRedis
        Start-Sleep -Seconds 1
        Start-LocalRedis
        Test-RedisConnection
    }
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor Red
        Write-Host "Usage: .\scripts\start-redis.ps1 -Action [start|stop|status|restart]" -ForegroundColor Cyan
        exit 1
    }
}
