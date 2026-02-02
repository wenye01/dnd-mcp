# Complete Clean and Test Script
# 从零开始，一次性完成数据库初始化和测试
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Complete Clean and Test Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 设置环境变量
$env:PGPASSWORD = "070831"
$env:TEST_DB_PASSWORD = "070831"

# 步骤 1: 完全清理
Write-Host "[1/7] Cleaning up..." -ForegroundColor Yellow
go clean -cache -testcache 2>&1 | Out-Null

# 断开数据库连接
$termOutput = psql -U postgres -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'dnd_mcp_test' AND pid <> pg_backend_pid();" 2>&1 | Out-String
Start-Sleep -Seconds 1

# 删除数据库
$dropOutput = psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS dnd_mcp_test;" 2>&1 | Out-String
Write-Host "[OK] Cleanup complete" -ForegroundColor Green

# 步骤 2: 创建数据库
Write-Host ""
Write-Host "[2/7] Creating database..." -ForegroundColor Yellow
$createOutput = psql -U postgres -d postgres -c "CREATE DATABASE dnd_mcp_test;" 2>&1 | Out-String
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Database created" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to create database" -ForegroundColor Red
    exit 1
}

# 步骤 3: 运行迁移
Write-Host ""
Write-Host "[3/7] Running migrations..." -ForegroundColor Yellow
$null = go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Migrations complete" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Migrations failed" -ForegroundColor Red
    exit 1
}

# 步骤 4: 运行 Store 测试
Write-Host ""
Write-Host "[4/7] Running Store tests..." -ForegroundColor Yellow
$storeResult = go test -v ./tests/unit/store/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Store tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Store tests failed" -ForegroundColor Red
    Write-Host $storeResult
    exit 1
}

# 步骤 5: 运行 LLM 测试
Write-Host ""
Write-Host "[5/7] Running LLM tests..." -ForegroundColor Yellow
$llmResult = go test -v ./tests/unit/client/llm/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] LLM tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] LLM tests failed" -ForegroundColor Red
    Write-Host $llmResult
    exit 1
}

# 步骤 6: 运行 Handler 测试
Write-Host ""
Write-Host "[6/7] Running Handler tests..." -ForegroundColor Yellow
$handlerResult = go test -v ./tests/unit/api/handler/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Handler tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Handler tests failed" -ForegroundColor Red
    Write-Host $handlerResult
    exit 1
}

# 步骤 7: 运行集成测试
Write-Host ""
Write-Host "[7/7] Running Integration tests..." -ForegroundColor Yellow
$integrationResult = go test -v ./tests/integration/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Integration tests passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Integration tests failed" -ForegroundColor Red
    Write-Host $integrationResult
    exit 1
}

# 总结
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "All Tests Passed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "  - Database created and migrated" -ForegroundColor White
Write-Host "  - All unit tests passed" -ForegroundColor White
Write-Host "  - All integration tests passed" -ForegroundColor White
Write-Host ""
