# One-Click Deploy Script for DND MCP Server (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

param(
    [switch]$SkipDb,
    [switch]$SkipBuild,
    [switch]$Force,
    [string]$LogLevel = "info"
)

# Configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ServerDir = Split-Path -Parent $ScriptDir
$ProjectRoot = Split-Path -Parent $ServerDir
$AppName = "dnd-server"
$HttpPort = 8081
$ContainerName = "dnd-postgres"

# Database configuration
$PostgresHost = "localhost"
$PostgresPort = 5432
$PostgresUser = "dnd"
$PostgresPassword = "password"
$PostgresDbName = "dnd_server"

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

function Test-Docker {
    try {
        docker --version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Test-PostgresRunning {
    $running = docker ps --filter "name=$ContainerName" --format "{{.Names}}"
    return ($running -eq $ContainerName)
}

function Start-Postgres {
    Write-Step "Starting PostgreSQL..."

    if (-not (Test-Docker)) {
        Write-Fail "Docker is not installed"
        Write-Host "Please install Docker Desktop: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
        return $false
    }

    if (Test-PostgresRunning) {
        Write-Success "PostgreSQL is already running"
        return $true
    }

    # Check if container exists but stopped
    $existing = docker ps -a --filter "name=$ContainerName" --format "{{.Names}}"
    if ($existing -eq $ContainerName) {
        Write-Host "  Starting existing container..." -ForegroundColor Cyan
        docker start $ContainerName | Out-Null
    }
    else {
        Write-Host "  Creating new container..." -ForegroundColor Cyan
        docker run -d `
            --name $ContainerName `
            -e POSTGRES_USER=postgres `
            -e POSTGRES_PASSWORD=postgres `
            -p "${PostgresPort}:5432" `
            postgres:16-alpine | Out-Null
    }

    if ($LASTEXITCODE -ne 0) {
        Write-Fail "Failed to start PostgreSQL container"
        return $false
    }

    # Wait for PostgreSQL to be ready
    Write-Host "  Waiting for PostgreSQL to be ready..." -ForegroundColor Cyan
    $maxRetries = 15
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        $ready = docker exec $ContainerName pg_isready -U postgres 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PostgreSQL is ready"
            return $true
        }
        $retryCount++
        Start-Sleep -Seconds 1
    }

    Write-Fail "PostgreSQL failed to start within $maxRetries seconds"
    return $false
}

function Initialize-Database {
    Write-Step "Initializing database..."

    # Check if database exists
    $dbExists = docker exec $ContainerName psql -U postgres -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$PostgresDbName'" 2>$null

    if ($dbExists -eq "1" -and -not $Force) {
        Write-Success "Database '$PostgresDbName' already exists"
    }
    else {
        if ($dbExists -eq "1") {
            Write-Host "  Dropping existing database (Force mode)..." -ForegroundColor Yellow
            docker exec $ContainerName psql -U postgres -d postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$PostgresDbName' AND pid <> pg_backend_pid();" 2>$null | Out-Null
            docker exec $ContainerName psql -U postgres -d postgres -c "DROP DATABASE $PostgresDbName;" 2>$null | Out-Null
        }

        Write-Host "  Creating database..." -ForegroundColor Cyan
        docker exec $ContainerName psql -U postgres -d postgres -c "CREATE DATABASE $PostgresDbName;" 2>$null | Out-Null
    }

    # Check if user exists
    $userExists = docker exec $ContainerName psql -U postgres -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$PostgresUser'" 2>$null

    if ($userExists -eq "1") {
        Write-Success "User '$PostgresUser' already exists"
    }
    else {
        Write-Host "  Creating user..." -ForegroundColor Cyan
        docker exec $ContainerName psql -U postgres -d postgres -c "CREATE USER $PostgresUser WITH PASSWORD '$PostgresPassword';" 2>$null | Out-Null
    }

    # Grant privileges
    docker exec $ContainerName psql -U postgres -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $PostgresDbName TO $PostgresUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U postgres -d $PostgresDbName -c "GRANT ALL PRIVILEGES ON SCHEMA public TO $PostgresUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U postgres -d $PostgresDbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $PostgresUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U postgres -d $PostgresDbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $PostgresUser;" 2>$null | Out-Null

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
    $BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
    $Ldflags = "-X 'main.Version=$Version' -X 'main.BuildTime=$BuildTime'"

    New-Item -ItemType Directory -Force -Path "bin" | Out-Null
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
            $pid = $conn.OwningProcess
            Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 1
        }
    }
}

function Start-Server {
    Write-Step "Starting server..."

    Set-Location $ServerDir

    # Stop existing server
    Stop-ExistingServer

    # Set environment variables
    $env:POSTGRES_HOST = $PostgresHost
    $env:POSTGRES_PORT = $PostgresPort
    $env:POSTGRES_USER = $PostgresUser
    $env:POSTGRES_PASSWORD = $PostgresPassword
    $env:POSTGRES_DBNAME = $PostgresDbName
    $env:POSTGRES_SSLMODE = "disable"
    $env:HTTP_HOST = "0.0.0.0"
    $env:HTTP_PORT = $HttpPort
    $env:LOG_LEVEL = $LogLevel
    $env:LOG_FORMAT = "text"

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

    $maxRetries = 10
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:$HttpPort/api/system/health" -TimeoutSec 2 -ErrorAction Stop
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
        Get-Content "$ServerDir\server-error.log" -Tail 20
    }
    if (Test-Path "$ServerDir\server.log") {
        Get-Content "$ServerDir\server.log" -Tail 20
    }
    return $false
}

function Show-Summary {
    Write-Header "Deployment Summary"

    Write-Host ""
    Write-Host "Server Status:" -ForegroundColor Cyan
    Write-Host "  URL:      http://localhost:$HttpPort" -ForegroundColor White
    Write-Host "  Health:   http://localhost:$HttpPort/api/system/health" -ForegroundColor White
    Write-Host "  Log:      $ServerDir\server.log" -ForegroundColor White
    Write-Host "  Error:    $ServerDir\server-error.log" -ForegroundColor White
    Write-Host ""
    Write-Host "Database:" -ForegroundColor Cyan
    Write-Host "  Host:     $PostgresHost:$PostgresPort" -ForegroundColor White
    Write-Host "  Database: $PostgresDbName" -ForegroundColor White
    Write-Host "  User:     $PostgresUser" -ForegroundColor White
    Write-Host ""
    Write-Host "Useful Commands:" -ForegroundColor Cyan
    Write-Host "  View logs:    Get-Content $ServerDir\server.log -Tail 50 -Wait" -ForegroundColor White
    Write-Host "  Stop server:  Get-Process -Name `"$AppName`" | Stop-Process" -ForegroundColor White
    Write-Host "  Test API:     Invoke-WebRequest http://localhost:$HttpPort/api/system/health" -ForegroundColor White
    Write-Host ""
}

# Main deployment flow
Write-Header "DND MCP Server Deployment"

$success = $true

# Step 1: Start PostgreSQL
if (-not $SkipDb) {
    if (-not (Start-Postgres)) {
        $success = $false
    }
}
else {
    Write-Step "Skipping database setup (-SkipDb)"
}

# Step 2: Initialize database
if ($success -and -not $SkipDb) {
    if (-not (Initialize-Database)) {
        $success = $false
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
