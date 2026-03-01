# Stop DND MCP Server (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

param(
    [switch]$All,
    [switch]$Db,
    [switch]$Clean
)

$AppName = "dnd-server"
$ClientAppName = "dnd-client"
$PostgresContainer = "dnd-postgres"
$RedisContainer = "dnd-redis"
$HttpPort = 8081
$ClientHttpPort = 8080

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Stop-Server {
    Write-Step "Stopping DND MCP Server..."

    # Stop by process name
    $process = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($process) {
        $process | Stop-Process -Force
        Write-Success "Server process stopped"
    }
    else {
        Write-Host "  No running server process found" -ForegroundColor Yellow
    }

    # Check port and kill if needed
    $conn = Get-NetTCPConnection -LocalPort $HttpPort -ErrorAction SilentlyContinue
    if ($conn) {
        $pid = $conn.OwningProcess
        Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
        Write-Success "Process on port $HttpPort stopped"
    }
}

function Stop-Client {
    Write-Step "Stopping DND MCP Client..."

    $process = Get-Process -Name $ClientAppName -ErrorAction SilentlyContinue
    if ($process) {
        $process | Stop-Process -Force
        Write-Success "Client process stopped"
    }
    else {
        Write-Host "  No running client process found" -ForegroundColor Yellow
    }

    $conn = Get-NetTCPConnection -LocalPort $ClientHttpPort -ErrorAction SilentlyContinue
    if ($conn) {
        $pid = $conn.OwningProcess
        Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
        Write-Success "Process on port $ClientHttpPort stopped"
    }
}

function Stop-Postgres {
    Write-Step "Stopping PostgreSQL..."

    $running = docker ps --filter "name=$PostgresContainer" --format "{{.Names}}"
    if ($running -eq $PostgresContainer) {
        docker stop $PostgresContainer | Out-Null
        Write-Success "PostgreSQL container stopped"
    }
    else {
        Write-Host "  PostgreSQL container is not running" -ForegroundColor Yellow
    }
}

function Stop-Redis {
    Write-Step "Stopping Redis..."

    $running = docker ps --filter "name=$RedisContainer" --format "{{.Names}}"
    if ($running -eq $RedisContainer) {
        docker stop $RedisContainer | Out-Null
        Write-Success "Redis container stopped"
    }
    else {
        Write-Host "  Redis container is not running" -ForegroundColor Yellow
    }
}

function Clear-Logs {
    Write-Step "Cleaning up logs..."

    $serverDir = "$PSScriptRoot\.."
    $clientDir = "$serverDir\..\client"

    Remove-Item "$serverDir\*.log" -ErrorAction SilentlyContinue
    Remove-Item "$clientDir\*.log" -ErrorAction SilentlyContinue

    Write-Success "Log files cleaned"
}

function Show-Status {
    Write-Host ""
    Write-Host "Status:" -ForegroundColor Cyan

    # Server status
    $serverProcess = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($serverProcess) {
        Write-Host "  Server:  Running (PID: $($serverProcess.Id))" -ForegroundColor Green
    }
    else {
        Write-Host "  Server:  Stopped" -ForegroundColor Yellow
    }

    # Client status
    $clientProcess = Get-Process -Name $ClientAppName -ErrorAction SilentlyContinue
    if ($clientProcess) {
        Write-Host "  Client:  Running (PID: $($clientProcess.Id))" -ForegroundColor Green
    }
    else {
        Write-Host "  Client:  Stopped" -ForegroundColor Yellow
    }

    # PostgreSQL status
    $postgresRunning = docker ps --filter "name=$PostgresContainer" --format "{{.Names}}"
    if ($postgresRunning -eq $PostgresContainer) {
        Write-Host "  PostgreSQL: Running" -ForegroundColor Green
    }
    else {
        Write-Host "  PostgreSQL: Stopped" -ForegroundColor Yellow
    }

    # Redis status
    $redisRunning = docker ps --filter "name=$RedisContainer" --format "{{.Names}}"
    if ($redisRunning -eq $RedisContainer) {
        Write-Host "  Redis:   Running" -ForegroundColor Green
    }
    else {
        Write-Host "  Redis:   Stopped" -ForegroundColor Yellow
    }
}

# Main logic
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP Stop Utility" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

if ($All) {
    Stop-Server
    Stop-Client
    Stop-Postgres
    Stop-Redis
    if ($Clean) {
        Clear-Logs
    }
}
elseif ($Db) {
    Stop-Postgres
    Stop-Redis
}
else {
    Stop-Server
    if ($Clean) {
        Clear-Logs
    }
}

Show-Status
