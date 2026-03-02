# Start PostgreSQL for Development (Windows PowerShell)
# -*- coding: utf-8 -*-

param(
    [string]$Action = "start",
    [switch]$Reset
)

# Load .env configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ServerDir = Split-Path -Parent $ScriptDir
$EnvFile = Join-Path $ServerDir ".env"

if (Test-Path $EnvFile) {
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match "^([^#][^=]+)=(.*)$") {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            if ($value -match '^"(.*)"$') { $value = $matches[1] }
            if ($value -match "^'(.*)'$") { $value = $matches[1] }
            Set-Item -Path "env:$key" -Value $value -Force
        }
    }
}

$PostgresPort = if ($env:POSTGRES_PORT) { [int]$env:POSTGRES_PORT } else { 5432 }
$PostgresUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" }
$PostgresPassword = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" }
$PostgresDb = "postgres"

$PossiblePgPaths = @(
    "C:\Program Files\PostgreSQL\18\bin",
    "C:\Program Files\PostgreSQL\17\bin",
    "C:\Program Files\PostgreSQL\16\bin",
    "C:\Program Files\PostgreSQL\15\bin",
    "C:\Program Files\PostgreSQL\14\bin",
    "C:\Tools\pgsql\bin",
    "C:\tools\postgresql\bin"
)

$PgBinPath = $null
foreach ($path in $PossiblePgPaths) {
    if (Test-Path "$path\pg_ctl.exe") {
        $PgBinPath = $path
        break
    }
}

$DataDir = "$env:APPDATA\dnd-mcp\postgres-data"
$ServiceName = "postgresql-x64-18"

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Test-PostgresService {
    $service = Get-Service -Name "*postgresql*" -ErrorAction SilentlyContinue
    if ($service) {
        return $service.Status -eq "Running"
    }
    return $false
}

function Test-LocalPostgres {
    if ($null -ne $PgBinPath) {
        return $true
    }
    return $false
}

function Initialize-PostgresData {
    Write-Step "Initializing PostgreSQL data directory..."

    if (-not (Test-Path $DataDir)) {
        New-Item -ItemType Directory -Path $DataDir -Force | Out-Null
    }

    $env:PGPASSWORD = $PostgresPassword
    & "$PgBinPath\initdb.exe" -D "$DataDir" -U $PostgresUser -A trust 2>&1 | Out-Null

    Write-Host "PostgreSQL data directory initialized" -ForegroundColor Green
}

function Start-LocalPostgres {
    Write-Step "Checking PostgreSQL..."

    # First check if PostgreSQL service is running
    if (Test-PostgresService) {
        Write-Host "PostgreSQL Windows service is running" -ForegroundColor Green
        return $true
    }

    # Check if PostgreSQL process is running
    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Host "PostgreSQL is already running (PID: $($pgProc.Id))" -ForegroundColor Green
        return $true
    }

    if (-not (Test-LocalPostgres)) {
        Write-Host "PostgreSQL not found in any of these locations:" -ForegroundColor Red
        foreach ($path in $PossiblePgPaths) {
            Write-Host "  - $path" -ForegroundColor Yellow
        }
        Write-Host "Please install PostgreSQL or add it to your PATH" -ForegroundColor Yellow
        return $false
    }

    if (-not (Test-Path $DataDir)) {
        Initialize-PostgresData
    }

    $env:PGPASSWORD = $PostgresPassword

    Write-Host "Starting PostgreSQL..." -ForegroundColor Cyan
    $result = & "$PgBinPath\pg_ctl.exe" start -D "$DataDir" -o "-p $PostgresPort" -l "$DataDir\postgres.log" -w 2>&1

    Start-Sleep -Seconds 3

    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Host "PostgreSQL started successfully (PID: $($pgProc.Id))" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "Failed to start PostgreSQL, checking log..." -ForegroundColor Red
        if (Test-Path "$DataDir\postgres.log") {
            Get-Content "$DataDir\postgres.log" -Tail 10
        }
        return $false
    }
}

function Stop-LocalPostgres {
    Write-Step "Stopping local PostgreSQL..."

    # Try to stop using pg_ctl first
    if ($null -ne $PgBinPath -and (Test-Path $DataDir)) {
        $env:PGPASSWORD = $PostgresPassword
        & "$PgBinPath\pg_ctl.exe" stop -D "$DataDir" -m fast -w 2>&1 | Out-Null
        Start-Sleep -Seconds 2
    }

    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Host "Stopping PostgreSQL processes..." -ForegroundColor Yellow
        taskkill /F /IM postgres.exe 2>$null | Out-Null
        Start-Sleep -Seconds 1
        Write-Host "PostgreSQL stopped" -ForegroundColor Green
    }
    else {
        Write-Host "PostgreSQL is not running" -ForegroundColor Yellow
    }
}

function Test-PostgresConnection {
    Write-Step "Testing PostgreSQL connection..."

    if ($null -eq $PgBinPath) {
        Write-Host "PostgreSQL binaries not found" -ForegroundColor Red
        return $false
    }

    $maxRetries = 10
    $retryCount = 0

    $env:PGPASSWORD = $PostgresPassword

    while ($retryCount -lt $maxRetries) {
        try {
            $result = & "$PgBinPath\psql.exe" -h localhost -p $PostgresPort -U $PostgresUser -d postgres -c "SELECT 1" 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "PostgreSQL connection successful!" -ForegroundColor Green
                Write-Host "PostgreSQL address: localhost:$PostgresPort" -ForegroundColor Cyan
                Write-Host "Username: $PostgresUser" -ForegroundColor Cyan
                return $true
            }
        }
        catch {
            # Continue retrying
        }

        $retryCount++
        Write-Host "Waiting for PostgreSQL to be ready... ($retryCount/$maxRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 1
    }

    Write-Host "PostgreSQL connection test failed after $maxRetries attempts" -ForegroundColor Red
    return $false
}

function Create-Database {
    Write-Step "Creating database if not exists..."

    if ($null -eq $PgBinPath) {
        Write-Host "PostgreSQL binaries not found" -ForegroundColor Red
        return $false
    }

    $env:PGPASSWORD = $PostgresPassword

    $dbExists = & "$PgBinPath\psql.exe" -h localhost -p $PostgresPort -U $PostgresUser -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = 'dnd_server'" 2>$null

    if ($dbExists -ne "1") {
        & "$PgBinPath\psql.exe" -h localhost -p $PostgresPort -U $PostgresUser -c "CREATE DATABASE dnd_server" 2>$null
        Write-Host "Database 'dnd_server' created" -ForegroundColor Green
    }
    else {
        Write-Host "Database 'dnd_server' already exists" -ForegroundColor Cyan
    }

    return $true
}

function Show-Status {
    Write-Step "PostgreSQL Status"

    # Check Windows service
    $service = Get-Service -Name "*postgresql*" -ErrorAction SilentlyContinue
    if ($service) {
        Write-Host "Windows Service: $($service.DisplayName)" -ForegroundColor Cyan
        Write-Host "Service Status: $($service.Status)" -ForegroundColor $(if ($service.Status -eq "Running") { "Green" } else { "Red" })
    }

    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Host "Process Status: Running" -ForegroundColor Green
        Write-Host "PID: $($pgProc.Id)" -ForegroundColor Cyan
        Write-Host "Port: $PostgresPort" -ForegroundColor Cyan
    }
    else {
        Write-Host "Process Status: Not Running" -ForegroundColor Red
    }

    if (Test-Path $DataDir) {
        Write-Host "Data Dir: $DataDir" -ForegroundColor Cyan
    }
}

# Main logic
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP PostgreSQL Manager (Local)" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

switch ($Action.ToLower()) {
    "start" {
        if ((Start-LocalPostgres) -and (Test-PostgresConnection)) {
            Create-Database
            Write-Host ""
            Write-Host "Common commands:" -ForegroundColor Cyan
            Write-Host "  Stop:     .\scripts\start-postgres.ps1 -Action stop" -ForegroundColor White
            Write-Host "  Status:   .\scripts\start-postgres.ps1 -Action status" -ForegroundColor White
            Write-Host "  Reset:    .\scripts\start-postgres.ps1 -Action reset" -ForegroundColor White
            Write-Host "  Connect:  psql -h localhost -p $PostgresPort -U $PostgresUser" -ForegroundColor White
        }
        else {
            Write-Host "PostgreSQL startup failed" -ForegroundColor Red
            exit 1
        }
    }
    "stop" {
        Stop-LocalPostgres
    }
    "status" {
        Show-Status
    }
    "reset" {
        Stop-LocalPostgres
        if (Test-Path $DataDir) {
            Remove-Item -Recurse -Force $DataDir
        }
        Start-Sleep -Seconds 1
        Start-LocalPostgres
        Test-PostgresConnection
    }
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor Red
        Write-Host "Usage: .\scripts\start-postgres.ps1 -Action [start|stop|status|reset]" -ForegroundColor Cyan
        exit 1
    }
}
