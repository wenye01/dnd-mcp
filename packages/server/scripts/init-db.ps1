# Initialize Database for DND MCP Server (Local PostgreSQL, No Docker)
# -*- coding: utf-8 -*-

param(
    [string]$PostgresHost = "",
    [int]$PostgresPort = 0,
    [string]$AdminUser = "",
    [string]$AdminPassword = "",
    [string]$DbName = "",
    [string]$DbUser = "",
    [string]$DbPassword = "",
    [switch]$Force
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

# Set defaults from environment or hardcoded defaults
if ([string]::IsNullOrEmpty($PostgresHost)) { $PostgresHost = if ($env:POSTGRES_HOST) { $env:POSTGRES_HOST } else { "localhost" } }
if ($PostgresPort -eq 0) { $PostgresPort = if ($env:POSTGRES_PORT) { [int]$env:POSTGRES_PORT } else { 5432 } }
if ([string]::IsNullOrEmpty($AdminUser)) { $AdminUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "postgres" } }
if ([string]::IsNullOrEmpty($AdminPassword)) { $AdminPassword = if ($env:POSTGRES_PASSWORD) { $env:POSTGRES_PASSWORD } else { "postgres" } }
if ([string]::IsNullOrEmpty($DbName)) { $DbName = if ($env:POSTGRES_DBNAME) { $env:POSTGRES_DBNAME } else { "dnd_server" } }
if ([string]::IsNullOrEmpty($DbUser)) { $DbUser = "dnd" }
if ([string]::IsNullOrEmpty($DbPassword)) { $DbPassword = "password" }

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

function Write-Warn {
    param([string]$Message)
    Write-Host "  [WARN] $Message" -ForegroundColor Yellow
}

function Test-PostgresConnection {
    $env:PGPASSWORD = $AdminPassword
    try {
        $result = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "SELECT 1" 2>&1
        if ($LASTEXITCODE -eq 0) {
            return $true
        }
    }
    catch {
        # Connection failed
    }
    return $false
}

function Test-PostgresRunning {
    $pgProc = Get-Process -Name "postgres" -ErrorAction SilentlyContinue
    return ($null -ne $pgProc)
}

function Initialize-Database {
    Write-Step "Initializing DND MCP Database..."

    # Check if PostgreSQL binaries are available
    if ($null -eq $PgBinPath) {
        Write-Error "PostgreSQL binaries not found in any of these locations:"
        foreach ($path in $PossiblePgPaths) {
            Write-Host "  - $path" -ForegroundColor Yellow
        }
        Write-Host "Please install PostgreSQL or add it to your PATH" -ForegroundColor Yellow
        Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
        exit 1
    }

    Write-Success "Found PostgreSQL at: $PgBinPath"

    # Check if PostgreSQL is running
    if (-not (Test-PostgresRunning)) {
        Write-Error "PostgreSQL is not running"
        Write-Host "Please start PostgreSQL first:" -ForegroundColor Yellow
        Write-Host "  .\scripts\start-postgres.ps1" -ForegroundColor White
        exit 1
    }

    # Wait for PostgreSQL to be ready
    Write-Step "Checking PostgreSQL readiness..."
    $maxRetries = 10
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        if (Test-PostgresConnection) {
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

    $env:PGPASSWORD = $AdminPassword

    # Check if database exists
    Write-Step "Checking if database '$DbName' exists..."
    $dbExists = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DbName'" 2>$null

    if ($dbExists -eq "1") {
        if ($Force) {
            Write-Host "  Database exists, dropping due to -Force flag..." -ForegroundColor Yellow
            # Terminate connections first
            $terminateSql = "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$DbName' AND pid <> pg_backend_pid();"
            & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c $terminateSql 2>$null | Out-Null
            & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "DROP DATABASE $DbName;" 2>$null | Out-Null
            Write-Success "Database dropped"
            $dbExists = $null
        }
        else {
            Write-Success "Database '$DbName' already exists"
        }
    }

    if ($dbExists -ne "1") {
        Write-Host "  Creating database '$DbName'..." -ForegroundColor Cyan
        $result = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "CREATE DATABASE $DbName;" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database '$DbName' created"
        }
        else {
            Write-Error "Failed to create database: $result"
            exit 1
        }
    }

    # Check if user exists (only if DbUser is different from AdminUser)
    if ($DbUser -ne $AdminUser) {
        Write-Step "Checking if user '$DbUser' exists..."
        $userExists = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DbUser'" 2>$null

        if ($userExists -eq "1") {
            if ($Force) {
                Write-Host "  User exists, updating password..." -ForegroundColor Yellow
                & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "ALTER USER $DbUser WITH PASSWORD '$DbPassword';" 2>$null | Out-Null
            }
            Write-Success "User '$DbUser' already exists"
        }
        else {
            Write-Host "  Creating user '$DbUser'..." -ForegroundColor Cyan
            $result = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "CREATE USER $DbUser WITH PASSWORD '$DbPassword';" 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Success "User '$DbUser' created"
            }
            else {
                Write-Error "Failed to create user: $result"
                exit 1
            }
        }
    }

    # Grant privileges
    Write-Step "Granting privileges..."
    $targetUser = if ($DbUser -ne $AdminUser) { $DbUser } else { $AdminUser }

    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DbName TO $targetUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d $DbName -c "GRANT ALL PRIVILEGES ON SCHEMA public TO $targetUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d $DbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $targetUser;" 2>$null | Out-Null
    & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $AdminUser -d $DbName -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $targetUser;" 2>$null | Out-Null
    Write-Success "Privileges granted"

    # Test connection
    Write-Step "Testing connection with application user..."
    $env:PGPASSWORD = $DbPassword
    $testResult = & "$PgBinPath\psql.exe" -h $PostgresHost -p $PostgresPort -U $targetUser -d $DbName -c "SELECT 1;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Connection test passed"
    }
    else {
        Write-Warn "Connection test failed, but database is initialized"
        Write-Host "  This might be due to pg_hba.conf configuration" -ForegroundColor Yellow
    }

    # Summary
    Write-Step "Database Initialization Complete"
    Write-Host ""
    Write-Host "Connection Details:" -ForegroundColor Cyan
    Write-Host "  Host:     $PostgresHost" -ForegroundColor White
    Write-Host "  Port:     $PostgresPort" -ForegroundColor White
    Write-Host "  Database: $DbName" -ForegroundColor White
    Write-Host "  User:     $targetUser" -ForegroundColor White
    Write-Host "  Password: $DbPassword" -ForegroundColor White
    Write-Host ""
    Write-Host "Connection String:" -ForegroundColor Cyan
    Write-Host "  host=$PostgresHost port=$PostgresPort user=$targetUser password=$DbPassword dbname=$DbName sslmode=disable" -ForegroundColor White
    Write-Host ""
    Write-Host "Environment Variables:" -ForegroundColor Cyan
    Write-Host "  `$env:POSTGRES_HOST=`"$PostgresHost`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_PORT=`"$PostgresPort`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_USER=`"$targetUser`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_PASSWORD=`"$DbPassword`"" -ForegroundColor White
    Write-Host "  `$env:POSTGRES_DBNAME=`"$DbName`"" -ForegroundColor White
    Write-Host ""
    Write-Host "PSQL Command:" -ForegroundColor Cyan
    Write-Host "  & `"$PgBinPath\psql.exe`" -h $PostgresHost -p $PostgresPort -U $targetUser -d $DbName" -ForegroundColor White
}

# Run initialization
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP Database Initializer" -ForegroundColor Green
Write-Host "  (Local PostgreSQL)" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

Initialize-Database
