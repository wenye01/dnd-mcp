# Complete Database Initialization Script
# 此脚本一次性完成所有数据库设置
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Database Initialization Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 设置密码
$env:PGPASSWORD = "070831"
$env:TEST_DB_PASSWORD = "070831"

# 1. 检查 PostgreSQL 连接
Write-Host "[1/5] Checking PostgreSQL connection..." -ForegroundColor Yellow
try {
    $null = & psql -U postgres -d postgres -c "SELECT 1" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] PostgreSQL connection successful" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Cannot connect to PostgreSQL" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "[ERROR] Cannot connect to PostgreSQL" -ForegroundColor Red
    exit 1
}

# 2. 删除旧数据库（如果存在）
Write-Host ""
Write-Host "[2/5] Dropping old database (if exists)..." -ForegroundColor Yellow

# 先断开所有连接
$terminateOutput = & psql -U postgres -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'dnd_mcp_test' AND pid <> pg_backend_pid();" 2>&1 | Out-String
Start-Sleep -Milliseconds 500

# 删除数据库
$dropOutput = & psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS dnd_mcp_test;" 2>&1 | Out-String
Write-Host "[OK] Old database dropped" -ForegroundColor Green

# 3. 创建新数据库
Write-Host ""
Write-Host "[3/5] Creating new database..." -ForegroundColor Yellow
$createOutput = & psql -U postgres -d postgres -c "CREATE DATABASE dnd_mcp_test;" 2>&1 | Out-String
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Database created successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to create database" -ForegroundColor Red
    Write-Host $createOutput
    exit 1
}

# 4. 运行迁移
Write-Host ""
Write-Host "[4/5] Running database migrations..." -ForegroundColor Yellow
$migrateOutput = & {
    $ErrorActionPreference = "Continue"
    go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up 2>&1
} | Out-String

# go run 会将日志写入 stderr，但我们需要检查退出代码
$migrateExitCode = $LASTEXITCODE

if ($migrateExitCode -eq 0) {
    Write-Host "[OK] Migrations completed successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Migration failed (exit code: $migrateExitCode)" -ForegroundColor Red
    exit 1
}

# 5. 验证数据库
Write-Host ""
Write-Host "[5/5] Verifying database schema..." -ForegroundColor Yellow

# 检查表是否存在
$tables = & psql -U postgres -d dnd_mcp_test -t -c "SELECT tablename FROM pg_tables WHERE schemaname='public';" 2>&1
$tables = $tables | Where-Object { $_.Trim() -ne "" }

$expectedTables = @("sessions", "messages")
$allTablesExist = $true

foreach ($table in $expectedTables) {
    if ($tables -contains $table) {
        Write-Host "[OK] Table '$table' exists" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Table '$table' missing" -ForegroundColor Red
        $allTablesExist = $false
    }
}

if (-not $allTablesExist) {
    Write-Host ""
    Write-Host "[ERROR] Database verification failed" -ForegroundColor Red
    exit 1
}

# 总结
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Database Initialization Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Database: dnd_mcp_test" -ForegroundColor Cyan
Write-Host "Tables: sessions, messages" -ForegroundColor Cyan
Write-Host ""
Write-Host "You can now run tests with:" -ForegroundColor Yellow
Write-Host "  .\scripts\test.ps1" -ForegroundColor White
Write-Host ""
