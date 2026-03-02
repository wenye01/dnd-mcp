# One-Click Deploy Script for DND MCP Server (Local PostgreSQL, No Docker)
# -*- coding: utf-8 -*-

param(
    [switch]$SkipDb,
    [switch]$SkipBuild,
    [switch]$Force,
    [string]$LogLevel = "info"
)

# Paths
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ServerDir = Split-Path -Parent $ScriptDir
$ProjectRoot = Split-Path -Parent $ServerDir
$AppName = "dnd-server"
$HttpPort = 8081

# Load .env configuration
$EnvFile = Join-Path $ServerDir ".env"
if (Test-Path $EnvFile) {
    Write-Host "Loading configuration from .env..." -ForegroundColor Cyan
    Get-Content $EnvFile | ForEach-Object {
        if ($_ -match "^([^#][^=]+)=(.*)$") {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            # Remove quotes if present
            if ($value -match '^"(.*)"$') { $value = $matches[1] }
            if ($value -match "^'(.*)'$") { $value = $matches[1] }
            Set-Item -Path "env:$key" -Value $value -Force
        }
    }
}

# Database configuration from environment or defaults
$PostgresHost = if ($env:POSTGRES_HOST) { $env:POSTGRES_HOST } else { "localhost" }
$PostgresPort = if ($env:POSTGRES_PORT) { [int]$env:POSTGRES_PORT } else { 5432 }
$PostgresUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" }
$PostgresPassword = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" }
$PostgresDbName = if ($env:POSTGRES_DBNAME) { $env:POSTGRES_DBNAME } else { "dnd_server" }

# PostgreSQL binary paths
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
    if (Test-Path "$path\psql.exe") {
        $PgBinPath = $path
        break
    }
}

function Write-Header {
    param([string]$Title)
    Write-Host ""
    Write-Host "=====================================" -ForegroundColor Green
    Write-Host "  $Title" -ForegroundColor Green
    Write-Host "=====================================" -ForegroundColor Green
}

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host "  [FAIL] $Message" -ForegroundColor Red
}

function Write-Warn {
    param([string]$Message)
    Write-Host "  [WARN] $Message" -ForegroundColor Yellow
}

function Test-PostgresConnection {
    if ($null -eq $PgBinPath) {
        Write-Fail "PostgreSQL binaries not found"
        return $false
    }

    $env:PGPASSWORD = $PostgresPassword
    try {
        $result = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -c "SELECT 1" 2>&1
        if ($LASTEXITCODE -eq 0) {
            return $true
        }
    }
    catch {
        # Connection failed
    }
    return $false
}

function Start-Postgres {
    Write-Step "Checking PostgreSQL..."

    if ($null -eq $PgBinPath) {
        Write-Fail "PostgreSQL binaries not found in any of these locations:"
        foreach ($path in $PossiblePgPaths) {
            Write-Host "  - $path" -ForegroundColor Yellow
        }
        Write-Host "Please install PostgreSQL or add it to your PATH" -ForegroundColor Yellow
        Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
        return $false
    }

    # Check if PostgreSQL is running
    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Success "PostgreSQL is already running (PID: $($pgProc.Id))"
        return $true
    }

    # Try to start PostgreSQL using pg_ctl
    $DataDir = "$env:APPDATA\dnd-mcp\postgres-data"

    if (-not (Test-Path $DataDir)) {
        Write-Host "  Initializing PostgreSQL data directory..." -ForegroundColor Cyan
        New-Item -ItemType Directory -Path $DataDir -Force | Out-Null
        & "$PgBinPath\initdb.exe" -D "$DataDir" -U $PostgresUser -A trust 2>&1 | Out-Null
    }

    Write-Host "  Starting PostgreSQL..." -ForegroundColor Cyan
    & "$PgBinPath\pg_ctl.exe" start -D "$DataDir" -o "-p $PostgresPort" -l "$DataDir\postgres.log" -w 2>&1 | Out-Null

    Start-Sleep -Seconds 2

    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    if ($pgProc) {
        Write-Success "PostgreSQL started (PID: $($pgProc.Id))"
        return $true
    }
    else {
        Write-Fail "Failed to start PostgreSQL"
        if (Test-Path "$DataDir\postgres.log") {
            Write-Host "Log output:" -ForegroundColor Yellow
            Get-Content "$DataDir\postgres.log" -Tail 10
        }
        return $false
    }
}

function Initialize-Database {
    Write-Step "Initializing database..."

    if ($null -eq $PgBinPath) {
        Write-Fail "PostgreSQL binaries not found"
        return $false
    }

    $env:PGPASSWORD = $PostgresPassword

    # Test connection first
    if (-not (Test-PostgresConnection)) {
        Write-Fail "Cannot connect to PostgreSQL at ${PostgresHost}:${PostgresPort}"
        Write-Host "Please ensure PostgreSQL is running" -ForegroundColor Yellow
        Write-Host "You may need to run: .\scripts\start-postgres.ps1" -ForegroundColor Yellow
        return $false
    }

    Write-Success "PostgreSQL connection successful"

    # Drop database if exists and Force flag is set
    if ($Force) {
        Write-Host "  Force mode: Dropping existing database..." -ForegroundColor Yellow

        # Terminate connections
        $terminateSql = "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$PostgresDbName' AND pid <> pg_backend_pid();"
        & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -c $terminateSql 2>$null | Out-Null

        # Drop database
        & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -c "DROP DATABASE IF EXISTS $PostgresDbName;" 2>$null | Out-Null
        Write-Success "Database dropped"
    }

    # Check if database exists
    $dbExists = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$PostgresDbName'" 2>$null

    if ($dbExists -eq "1") {
        Write-Success "Database '$PostgresDbName' already exists"
    }
    else {
        Write-Host "  Creating database '$PostgresDbName'..." -ForegroundColor Cyan
        $result = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -c "CREATE DATABASE $PostgresDbName;" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database '$PostgresDbName' created"
        }
        else {
            Write-Fail "Failed to create database: $result"
            return $false
        }
    }

    # Grant privileges
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $PostgresDbName TO $PostgresUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d $PostgresDbName -c "GRANT ALL ON SCHEMA public TO $PostgresUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d $PostgresDbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $PostgresUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d $PostgresDbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $PostgresUser;" 2>$null | Out-Null

    Write-Success "Database initialized"
    return $true
}

function Build-Server {
    Write-Step "Building server..."

    Set-Location $ServerDir

    # Clean old build
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin" -ErrorAction SilentlyContinue
    }

    # Download dependencies
    Write-Host "  Downloading dependencies..." -ForegroundColor Cyan
    go mod tidy 2>&1 | Out-Null

    # Build
    Write-Host "  Compiling..." -ForegroundColor Cyan
    $Version = "0.1.0"
    $GitCommit = git rev-parse --short HEAD 2>$null
    if ($LASTEXITCODE -ne 0) { $GitCommit = "unknown" }
    $BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"

    New-Item -ItemType Directory -Force -Path "bin" | Out-Null

    $Ldflags = "-X main.version=$Version -X main.gitCommit=$GitCommit -X main.buildTime=$BuildTime"
    go build -ldflags $Ldflags -o "bin/$AppName.exe" ./cmd/server 2>&1

    if ($LASTEXITCODE -ne 0 -or -not (Test-Path "bin/$AppName.exe")) {
        Write-Fail "Build failed"
        return $false
    }

    $fileInfo = Get-Item "bin/$AppName.exe"
    Write-Success "Build completed ($([math]::Round($fileInfo.Length / 1MB, 2)) MB)"
    return $true
}

function Test-PortAvailable {
    param([int]$Port)

    $connection = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
    return ($null -eq $connection)
}

function Stop-ExistingServer {
    Write-Host "  Checking for existing server process..." -ForegroundColor Cyan

    # Stop by process name
    $process = Get-Process -Name $AppName -ErrorAction SilentlyContinue
    if ($process) {
        Write-Host "  Stopping existing server process..." -ForegroundColor Yellow
        $process | Stop-Process -Force
        Start-Sleep -Seconds 1
    }

    # Check port and kill if needed
    if (-not (Test-PortAvailable -Port $HttpPort)) {
        Write-Host "  Port $HttpPort is in use, attempting to free..." -ForegroundColor Yellow
        $conn = Get-NetTCPConnection -LocalPort $HttpPort -ErrorAction SilentlyContinue
        if ($conn) {
            $procId = $conn.OwningProcess
            Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 1
        }
    }
}

function Start-Server {
    Write-Step "Starting server..."

    Set-Location $ServerDir

    # Stop existing server
    Stop-ExistingServer

    # Set environment variables from .env (already loaded)
    # Override with command line options
    $env:HTTP_PORT = $HttpPort
    $env:LOG_LEVEL = $LogLevel
    $env:LOG_FORMAT = "text"
    $env:POSTGRES_SSLMODE = "disable"

    # Clean old logs
    Remove-Item "server.log" -ErrorAction SilentlyContinue
    Remove-Item "server-error.log" -ErrorAction SilentlyContinue

    # Start server
    Write-Host "  Starting $AppName..." -ForegroundColor Cyan
    Start-Process -FilePath ".\bin\$AppName.exe" `
        -RedirectStandardOutput "server.log" `
        -RedirectStandardError "server-error.log" `
        -WindowStyle Hidden

    # Wait for startup
    Start-Sleep -Seconds 3

    return $true
}

function Test-ServerHealth {
    Write-Step "Checking server health..."

    $maxRetries = 15
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$HttpPort/health" -TimeoutSec 2 -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                Write-Success "Server is healthy"
                return $true
            }
        }
        catch {
            # Continue retrying
        }

        $retryCount++
        Write-Host "  Waiting for server... ($retryCount/$maxRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 1
    }

    Write-Fail "Server health check failed"
    Write-Host "`nServer logs:" -ForegroundColor Yellow
    if (Test-Path "$ServerDir\server-error.log") {
        Get-Content "$ServerDir\server-error.log" -Tail 30
    }
    if (Test-Path "$ServerDir\server.log") {
        Write-Host "`nStandard output:" -ForegroundColor Yellow
        Get-Content "$ServerDir\server.log" -Tail 30
    }
    return $false
}

function Show-Summary {
    Write-Header "Deployment Summary"

    Write-Host ""
    Write-Host "Configuration:" -ForegroundColor Cyan
    Write-Host "  Config File: $EnvFile" -ForegroundColor White
    Write-Host ""
    Write-Host "Server Status:" -ForegroundColor Cyan
    Write-Host "  URL:      http://localhost:$HttpPort" -ForegroundColor White
    Write-Host "  Health:   http://localhost:$HttpPort/health" -ForegroundColor White
    Write-Host "  Log:      $ServerDir\server.log" -ForegroundColor White
    Write-Host "  Error:    $ServerDir\server-error.log" -ForegroundColor White
    Write-Host ""
    Write-Host "Database:" -ForegroundColor Cyan
    Write-Host "  Host:     ${PostgresHost}:${PostgresPort}" -ForegroundColor White
    Write-Host "  Database: $PostgresDbName" -ForegroundColor White
    Write-Host "  User:     $PostgresUser" -ForegroundColor White
    Write-Host ""
    Write-Host "Useful Commands:" -ForegroundColor Cyan
    Write-Host "  View logs:    Get-Content $ServerDir\server.log -Tail 50 -Wait" -ForegroundColor White
    Write-Host "  Stop server:  Get-Process -Name $AppName | Stop-Process" -ForegroundColor White
    Write-Host "  Test API:     Invoke-WebRequest http://localhost:$HttpPort/health" -ForegroundColor White
    if ($PgBinPath) {
        Write-Host "  PSQL:         & `"$PgBinPath\psql.exe`" -h $PostgresHost -p $PostgresPort -U $PostgresUser -d $PostgresDbName" -ForegroundColor White
    }
    Write-Host ""
}

# Main deployment flow
Write-Header "DND MCP Server Deployment (Local)"

$success = $true

# Step 1: Start PostgreSQL
if (-not $SkipDb) {
    if (-not (Start-Postgres)) {
        $success = $false
    }
}

# Step 2: Initialize database
if ($success -and -not $SkipDb) {
    if (-not (Initialize-Database)) {
        $success = $false
    }
}
else {
    if ($SkipDb) {
        Write-Step "Skipping database setup (-SkipDb)"
    }
}

# Step 3: Build server
if ($success -and -not $SkipBuild) {
    if (-not (Build-Server)) {
        $success = $false
    }
}
elseif ($SkipBuild) {
    Write-Step "Skipping build (-SkipBuild)"
}

# Step 4: Start server
if ($success) {
    if (-not (Start-Server)) {
        $success = $false
    }
}

# Step 5: Health check
if ($success) {
    if (-not (Test-ServerHealth)) {
        $success = $false
    }
}

# Show summary
if ($success) {
    Show-Summary
    Write-Host "Deployment completed successfully!" -ForegroundColor Green
    exit 0
}
else {
    Write-Host "`nDeployment failed. Check the logs above for details." -ForegroundColor Red
    exit 1
}
