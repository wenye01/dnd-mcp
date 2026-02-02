# Quick Setup and Test Script for Fresh Environment
# 此脚本用于在全新环境中一键设置并运行测试

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "DND MCP Client - Quick Setup & Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查Go
Write-Host "[1/7] 检查Go环境..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    Write-Host "[OK] $goVersion" -ForegroundColor Green
} catch {
    Write-Host "[错误] 未找到Go，请先安装Go 1.25+" -ForegroundColor Red
    exit 1
}

# 检查PostgreSQL
Write-Host ""
Write-Host "[2/7] 检查PostgreSQL..." -ForegroundColor Yellow
try {
    $null = & psql -U postgres -d postgres -c "SELECT 1" 2>&1
    Write-Host "[OK] PostgreSQL连接成功" -ForegroundColor Green
} catch {
    Write-Host "[错误] 无法连接到PostgreSQL" -ForegroundColor Red
    Write-Host "请确保PostgreSQL已安装并运行在localhost:5432" -ForegroundColor Red
    exit 1
}

# 创建测试数据库
Write-Host ""
Write-Host "[3/7] 创建测试数据库..." -ForegroundColor Yellow
$null = & psql -U postgres -c "SELECT 1 FROM pg_database WHERE datname='dnd_mcp_test'" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[信息] 数据库已存在，将重新创建" -ForegroundColor Yellow
    $null = & psql -U postgres -c "DROP DATABASE dnd_mcp_test;" 2>&1
}

$null = & psql -U postgres -c "CREATE DATABASE dnd_mcp_test;" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] 测试数据库创建成功" -ForegroundColor Green
} else {
    Write-Host "[错误] 无法创建测试数据库" -ForegroundColor Red
    exit 1
}

# 安装依赖
Write-Host ""
Write-Host "[4/7] 安装Go依赖..." -ForegroundColor Yellow
$null = go mod download 2>&1
Write-Host "[OK] 依赖安装完成" -ForegroundColor Green

# 运行数据库迁移
Write-Host ""
Write-Host "[5/7] 运行数据库迁移..." -ForegroundColor Yellow
$env:PGPASSWORD = "070831"
$null = go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] 数据库迁移完成" -ForegroundColor Green
} else {
    Write-Host "[错误] 数据库迁移失败" -ForegroundColor Red
    exit 1
}

# 运行测试
Write-Host ""
Write-Host "[6/7] 运行测试..." -ForegroundColor Yellow
$env:TEST_DB_PASSWORD = "070831"

Write-Host ""
Write-Host "运行 Store 测试..." -ForegroundColor Cyan
$storeResult = go test ./internal/store/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Store 测试通过" -ForegroundColor Green
} else {
    Write-Host "[失败] Store 测试失败" -ForegroundColor Red
    Write-Host $storeResult
}

Write-Host "运行 LLM 测试..." -ForegroundColor Cyan
$llmResult = go test ./internal/client/llm/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] LLM 测试通过" -ForegroundColor Green
} else {
    Write-Host "[失败] LLM 测试失败" -ForegroundColor Red
    Write-Host $llmResult
}

Write-Host "运行 Handler 测试..." -ForegroundColor Cyan
$handlerResult = go test ./internal/api/handler/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Handler 测试通过" -ForegroundColor Green
} else {
    Write-Host "[失败] Handler 测试失败" -ForegroundColor Red
    Write-Host $handlerResult
}

Write-Host "运行集成测试..." -ForegroundColor Cyan
$integrationResult = go test ./tests/integration/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] 集成测试通过" -ForegroundColor Green
} else {
    Write-Host "[失败] 集成测试失败" -ForegroundColor Red
    Write-Host $integrationResult
}

# 生成覆盖率报告
Write-Host ""
Write-Host "[7/7] 生成覆盖率报告..." -ForegroundColor Yellow
if (!(Test-Path "tests/reports")) {
    New-Item -ItemType Directory -Path "tests/reports" | Out-Null
}

$null = go test -coverprofile=tests/reports/coverage.out -covermode=atomic ./internal/... ./tests/integration/... 2>&1
if (Test-Path "tests/reports/coverage.out") {
    $null = go tool cover -html=tests/reports/coverage.out -o tests/reports/coverage.html 2>&1
    Write-Host "[OK] 覆盖率报告生成: tests/reports/coverage.html" -ForegroundColor Green

    $coverage = go tool cover -func=tests/reports/coverage.out | Select-String "total:"
    if ($coverage) {
        $parts = $coverage.Line.Split()
        Write-Host "总体覆盖率: $($parts[$parts.Length - 2])" -ForegroundColor Cyan
    }
}

# 总结
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "设置完成！" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$allPassed = ($LASTEXITCODE -eq 0) -and ($storeResult -eq $null -or $storeResult -match "ok") -and ($llmResult -eq $null -or $llmResult -match "ok") -and ($handlerResult -eq $null -or $handlerResult -match "ok") -and ($integrationResult -eq $null -or $integrationResult -match "ok")

if ($allPassed) {
    Write-Host "[SUCCESS] 所有测试通过！" -ForegroundColor Green
    Write-Host ""
    Write-Host "下一步:" -ForegroundColor Yellow
    Write-Host "1. 查看覆盖率报告: tests/reports/coverage.html" -ForegroundColor White
    Write-Host "2. 运行完整测试套件: .\scripts\test.ps1" -ForegroundColor White
    Write-Host "3. 开始开发: 查看 doc/ 目录了解项目结构" -ForegroundColor White
    exit 0
} else {
    Write-Host "[FAILURE] 部分测试失败" -ForegroundColor Red
    Write-Host "请查看上方错误信息" -ForegroundColor Red
    exit 1
}
