# Initialize Database for DND MCP Server (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

param(
    [string]$PostgresHost = "localhost",
    [int]$PostgresPort = 5432,
    [string]$AdminUser = "postgres",
    [string]$AdminPassword = "postgres",
    [string]$DbName = "dnd_server",
    [string]$DbUser = "dnd",
    [string]$DbPassword = "password",
    [switch]$Force
)

$ContainerName = "dnd-postgres"

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "  [ERROR] $Message" -ForegroundColor Red
}

function Invoke-PostgresSql {
    param([string]$Sql, [string]$Database = "postgres")

    $result = docker exec -i $ContainerName psql -U $AdminUser -d $Database -c "$Sql" 2>&1
    return $result
}

function Test-PostgresRunning {
    $running = docker ps --filter "name=$ContainerName" --format "{{.Names}}"
    return ($running -eq $ContainerName)
}

function Initialize-Database {
    Write-Step "Initializing DND MCP Database..."

    # Check if PostgreSQL is running
    if (-not (Test-PostgresRunning)) {
        Write-Error "PostgreSQL container is not running"
        Write-Host "Please run: .\scripts\start-postgres.ps1" -ForegroundColor Yellow
        exit 1
    }

    # Wait for PostgreSQL to be ready
    Write-Step "Checking PostgreSQL readiness..."
    $maxRetries = 10
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        $ready = docker exec $ContainerName pg_isready -U $AdminUser 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PostgreSQL is ready"
            break
        }
        $retryCount++
        Start-Sleep -Seconds 1
    }

    if ($retryCount -eq $maxRetries) {
        Write-Error "PostgreSQL is not ready after $maxRetries seconds"
        exit 1
    }

    # Check if database exists
    Write-Step "Checking if database '$DbName' exists..."
    $dbExists = docker exec $ContainerName psql -U $AdminUser -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DbName'"

    if ($dbExists -eq "1") {
        if ($Force) {
            Write-Host "  Database exists, dropping due to -Force flag..." -ForegroundColor Yellow
            # Terminate connections first
            docker exec $ContainerName psql -U $AdminUser -d postgres -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$DbName' AND pid <> pg_backend_pid();" 2>$null | Out-Null
            docker exec $ContainerName psql -U $AdminUser -d postgres -c "DROP DATABASE $DbName;" 2>$null | Out-Null
            Write-Success "Database dropped"
        }
        else {
            Write-Success "Database '$DbName' already exists"
        }
    }
    else {
        Write-Host "  Creating database '$DbName'..." -ForegroundColor Cyan
        $result = docker exec $ContainerName psql -U $AdminUser -d postgres -c "CREATE DATABASE $DbName;" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database '$DbName' created"
        }
        else {
            Write-Error "Failed to create database: $result"
            exit 1
        }
    }

    # Check if user exists
    Write-Step "Checking if user '$DbUser' exists..."
    $userExists = docker exec $ContainerName psql -U $AdminUser -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DbUser'"

    if ($userExists -eq "1") {
        if ($Force) {
            Write-Host "  User exists, updating password..." -ForegroundColor Yellow
            docker exec $ContainerName psql -U $AdminUser -d postgres -c "ALTER USER $DbUser WITH PASSWORD '$DbPassword';" 2>$null | Out-Null
        }
        Write-Success "User '$DbUser' already exists"
    }
    else {
        Write-Host "  Creating user '$DbUser'..." -ForegroundColor Cyan
        $result = docker exec $ContainerName psql -U $AdminUser -d postgres -c "CREATE USER $DbUser WITH PASSWORD '$DbPassword';" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "User '$DbUser' created"
        }
        else {
            Write-Error "Failed to create user: $result"
            exit 1
        }
    }

    # Grant privileges
    Write-Step "Granting privileges..."
    docker exec $ContainerName psql -U $AdminUser -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DbName TO $DbUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U $AdminUser -d $DbName -c "GRANT ALL PRIVILEGES ON SCHEMA public TO $DbUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U $AdminUser -d $DbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DbUser;" 2>$null | Out-Null
    docker exec $ContainerName psql -U $AdminUser -d $DbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DbUser;" 2>$null | Out-Null
    Write-Success "Privileges granted"

    # Test connection
    Write-Step "Testing connection with new user..."
    $testResult = docker exec $ContainerName psql -U $DbUser -d $DbName -h localhost -c "SELECT 1;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Connection test passed"
    }
    else {
        Write-Host "  Warning: Connection test failed, but database is initialized" -ForegroundColor Yellow
        Write-Host "  This might be due to pg_hba.conf configuration" -ForegroundColor Yellow
    }

    # Summary
    Write-Step "Database Initialization Complete"
    Write-Host ""
    Write-Host "Connection Details:" -ForegroundColor Cyan
    Write-Host "  Host:     $PostgresHost" -ForegroundColor White
    Write-Host "  Port:     $PostgresPort" -ForegroundColor White
    Write-Host "  Database: $DbName" -ForegroundColor White
    Write-Host "  User:     $DbUser" -ForegroundColor White
    Write-Host "  Password: $DbPassword" -ForegroundColor White
    Write-Host ""
    Write-Host "Connection String:" -ForegroundColor Cyan
    Write-Host "  host=$PostgresHost port=$PostgresPort user=$DbUser password=$DbPassword dbname=$DbName sslmode=disable" -ForegroundColor White
    Write-Host ""
    Write-Host "Environment Variables:" -ForegroundColor Cyan
    Write-Host "  `$env:POSTGRES_HOST=`"$PostgresHost`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_PORT=`"$PostgresPort`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_USER=`"$DbUser`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_PASSWORD=`"$DbPassword`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_DBNAME=`"$DbName`"" -ForegroundColor White
}

# Run initialization
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP Database Initializer" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

Initialize-Database
