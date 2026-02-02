# MCP Client Testing Script
# UTF-8 Encoding

# Set working directory to project root
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location (Split-Path -Parent $scriptPath)

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================"
Write-Host "MCP Client Test Suite"
Write-Host "========================================"
Write-Host ""

# Set database password
$env:TEST_DB_PASSWORD = "070831"
$env:PGPASSWORD = "070831"

# Check Go
Write-Host "[1/6] Checking test environment..."
$null = go version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go not installed" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Go is installed" -ForegroundColor Green

# Check database connection
Write-Host "Checking database connection..."
$null = & psql -h localhost -U postgres -d postgres -c "SELECT 1" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Cannot connect to database" -ForegroundColor Yellow
} else {
    Write-Host "[OK] Database is available" -ForegroundColor Green
}

# Create test reports directory
Write-Host ""
Write-Host "[2/6] Setting up test environment..."
if (!(Test-Path "tests/reports")) {
    New-Item -ItemType Directory -Path "tests/reports" | Out-Null
}
Remove-Item "tests/reports\*" -Force -ErrorAction SilentlyContinue
Write-Host "[OK] Test reports directory created" -ForegroundColor Green

# Setup database - create if not exists, then run migrations
Write-Host ""
Write-Host "[3/6] Setting up test database..."

# Check if database exists
$dbExists = $false
try {
    $null = & psql -U postgres -d dnd_mcp_test -c "SELECT 1" 2>&1
    if ($LASTEXITCODE -eq 0) {
        $dbExists = $true
    }
} catch {
    $dbExists = $false
}

if ($dbExists) {
    Write-Host "Database exists, dropping old schema..."
    $null = go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action down 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Old schema dropped" -ForegroundColor Green
    }
} else {
    Write-Host "Database does not exist, creating..."
    $null = & psql -U postgres -d postgres -c "CREATE DATABASE dnd_mcp_test;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Database created" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Failed to create database" -ForegroundColor Red
        exit 1
    }
}

Write-Host "Creating schema..."
$null = go run scripts/migrate/main.go -dsn "postgres://postgres:070831@localhost:5432/dnd_mcp_test?sslmode=disable" -action up 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[OK] Schema created" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to create database schema" -ForegroundColor Red
    exit 1
}

# Run unit tests
Write-Host ""
Write-Host "[4/6] Running unit tests..."
Write-Host ""

Write-Host "Testing: tests/unit/store" -ForegroundColor Cyan
go test -v ./tests/unit/store/... 2>&1 | Tee-Object -FilePath "tests/reports/store_tests.txt" -Append
$storeResult = $LASTEXITCODE

Write-Host ""
Write-Host "Testing: tests/unit/client/llm" -ForegroundColor Cyan
go test -v ./tests/unit/client/llm/... 2>&1 | Tee-Object -FilePath "tests/reports/llm_tests.txt" -Append
$llmResult = $LASTEXITCODE

Write-Host ""
Write-Host "Testing: tests/unit/api/handler" -ForegroundColor Cyan
go test -v ./tests/unit/api/handler/... 2>&1 | Tee-Object -FilePath "tests/reports/handler_tests.txt" -Append
$handlerResult = $LASTEXITCODE

# Run integration tests
Write-Host ""
Write-Host "[5/6] Running integration tests..."
Write-Host ""
Write-Host "Testing: tests/integration" -ForegroundColor Cyan
go test -v ./tests/integration/... 2>&1 | Tee-Object -FilePath "tests/reports/integration_tests.txt" -Append
$integrationResult = $LASTEXITCODE

# Check for race conditions
Write-Host ""
Write-Host "[6/6] Checking for race conditions..."
Write-Host "Testing with -race flag (this may take a while)..."
go test -race ./tests/unit/... ./tests/integration/... 2>&1 | Tee-Object -FilePath "tests/reports/race_tests.txt" -Append
$raceResult = $LASTEXITCODE

# Generate coverage
Write-Host ""
Write-Host "Generating coverage report..."
$coverageOutput = go test -coverprofile=tests/reports/coverage.out -covermode=atomic ./tests/unit/... ./tests/integration/... 2>&1
if (Test-Path "tests/reports/coverage.out") {
    go tool cover -html=tests/reports/coverage.out -o tests/reports/coverage.html 2>&1 | Out-Null
    Write-Host "[OK] Coverage report generated: tests/reports/coverage.html" -ForegroundColor Green

    $coverage = go tool cover -func=tests/reports/coverage.out | Select-String "total:"
    if ($coverage) {
        $parts = $coverage.Line.Split()
        Write-Host "Total Coverage: $($parts[$parts.Length - 2])" -ForegroundColor Cyan
    }
} else {
    Write-Host "WARNING: Could not generate coverage report" -ForegroundColor Yellow
}

# Summary
Write-Host ""
Write-Host "========================================"
Write-Host "Test Summary"
Write-Host "========================================"
Write-Host ""

$allPassed = $true

if ($storeResult -ne 0) {
    Write-Host "[FAIL] tests/unit/store tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] tests/unit/store tests" -ForegroundColor Green
}

if ($llmResult -ne 0) {
    Write-Host "[FAIL] tests/unit/client/llm tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] tests/unit/client/llm tests" -ForegroundColor Green
}

if ($handlerResult -ne 0) {
    Write-Host "[FAIL] tests/unit/api/handler tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] tests/unit/api/handler tests" -ForegroundColor Green
}

if ($integrationResult -ne 0) {
    Write-Host "[FAIL] integration tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] integration tests" -ForegroundColor Green
}

if ($raceResult -ne 0) {
    Write-Host "[WARN] Race condition detected" -ForegroundColor Yellow
    Write-Host "       Check tests/reports/race_tests.txt for details" -ForegroundColor Yellow
} else {
    Write-Host "[PASS] No race conditions detected" -ForegroundColor Green
}

Write-Host ""

if ($allPassed) {
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "All tests passed!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    exit 0
} else {
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "Some tests failed!" -ForegroundColor Red
    Write-Host "Check the reports in tests/reports/ for details" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    exit 1
}
