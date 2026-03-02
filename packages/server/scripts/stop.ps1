# Stop DND MCP Server (Windows PowerShell)
# -*- coding: utf-8 -*-

param(
    [switch]$All,
    [switch]$Db,
    [switch]$Clean
)

$AppName = "dnd-server"
$ClientAppName = "dnd-client"
$HttpPort = 8081
$ClientHttpPort = 8080

$PostgresBinPaths = @(
    "C:\Program Files\PostgreSQL\18\bin",
    "C:\Program Files\PostgreSQL\17\bin",
    "C:\Program Files\PostgreSQL\16\bin",
    "C:\Program Files\PostgreSQL\15\bin",
    "C:\Program Files\PostgreSQL\14\bin",
    "C:\Tools\pgsql\bin",
    "C:\tools\postgresql\bin"
)
$PostgresDataDir = "$env:APPDATA\dnd-mcp\postgres-data"

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

    $process = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($process) {
        $process | Stop-Process -Force
        Write-Success "Server process stopped"
    }
    else {
        Write-Host "  No running server process found" -ForegroundColor Yellow
    }

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

    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        $pgBinPath = $null
        foreach ($path in $PostgresBinPaths) {
            if (Test-Path "$path\pg_ctl.exe") {
                $pgBinPath = $path
                break
            }
        }

        if ($pgBinPath) {
            $env:PGPASSWORD = "postgres"
            & "$pgBinPath\pg_ctl.exe" stop -D $PostgresDataDir -m fast -w 2>$null
        }

        Start-Sleep -Seconds 1

        $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
        if ($pgProc) {
            $pgProc | Stop-Process -Force
        }

        Write-Success "PostgreSQL stopped"
    }
    else {
        Write-Host "  PostgreSQL is not running" -ForegroundColor Yellow
    }
}

function Stop-Redis {
    Write-Step "Stopping Redis..."

    $redisProc = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisProc) {
        $redisProc | Stop-Process -Force
        Write-Success "Redis stopped"
    }
    else {
        Write-Host "  Redis is not running" -ForegroundColor Yellow
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

    $serverProcess = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($serverProcess) {
        Write-Host "  Server:  Running (PID: $($serverProcess.Id))" -ForegroundColor Green
    }
    else {
        Write-Host "  Server:  Stopped" -ForegroundColor Yellow
    }

    $clientProcess = Get-Process -Name $ClientAppName -ErrorAction SilentlyContinue
    if ($clientProcess) {
        Write-Host "  Client:  Running (PID: $($clientProcess.Id))" -ForegroundColor Green
    }
    else {
        Write-Host "  Client:  Stopped" -ForegroundColor Yellow
    }

    $postgresRunning = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($postgresRunning) {
        Write-Host "  PostgreSQL: Running" -ForegroundColor Green
    }
    else {
        Write-Host "  PostgreSQL: Stopped" -ForegroundColor Yellow
    }

    $redisRunning = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue
    if ($redisRunning) {
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
