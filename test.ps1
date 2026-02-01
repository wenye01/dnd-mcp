# MCP Client Testing Script
# UTF-8 Encoding

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================"
Write-Host "MCP Client Test Suite"
Write-Host "========================================"
Write-Host ""

# Set database password
$env:TEST_DB_PASSWORD = "070831"
$env:PGPASSWORD = "070831"

# Check Go
Write-Host "[1/5] Checking test environment..."
$null = go version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go not installed"
    exit 1
}
Write-Host "[OK] Go is installed"

# Check database
Write-Host "Checking database..."
$null = & psql -h localhost -U postgres -d postgres -c "SELECT 1" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: Cannot connect to database"
} else {
    Write-Host "[OK] Database is available"
}

# Clean old data
Write-Host ""
Write-Host "[2/5] Cleaning old test data..."
if (!(Test-Path "test-reports")) {
    New-Item -ItemType Directory -Path "test-reports" | Out-Null
}
Remove-Item "test-reports\*" -Force -ErrorAction SilentlyContinue

# Drop existing test tables to ensure clean schema
Write-Host "Dropping existing test tables..."
& psql -h localhost -U postgres -d dnd_mcp_test -c "DROP TABLE IF EXISTS messages CASCADE; DROP TABLE IF EXISTS sessions CASCADE;" 2>&1 | Out-Null
Write-Host "[OK] Cleaned test data"

# Run unit tests
Write-Host ""
Write-Host "[3/5] Running unit tests..."
Write-Host ""
Write-Host "Testing: internal/store"
go test -v ./internal/store/... 2>&1 | Tee-Object -FilePath "test-reports\test_output.txt" -Append
$storeResult = $LASTEXITCODE

Write-Host ""
Write-Host "Testing: internal/client/llm"
go test -v ./internal/client/llm/... 2>&1 | Tee-Object -FilePath "test-reports\test_output.txt" -Append
$llmResult = $LASTEXITCODE

Write-Host ""
Write-Host "Testing: internal/api/handler"
go test -v ./internal/api/handler/... 2>&1 | Tee-Object -FilePath "test-reports\test_output.txt" -Append
$handlerResult = $LASTEXITCODE

# Run integration tests
Write-Host ""
Write-Host "[4/5] Running integration tests..."
Write-Host ""
Write-Host "Testing: tests/integration"
go test -v ./tests/integration/... 2>&1 | Tee-Object -FilePath "test-reports\test_output.txt" -Append
$integrationResult = $LASTEXITCODE

# Generate coverage
Write-Host ""
Write-Host "[5/5] Generating coverage report..."
go test -coverprofile=test-reports\coverage.out -covermode=atomic ./... 2>&1 | Out-Null
if (Test-Path "test-reports\coverage.out") {
    go tool cover -html=test-reports\coverage.out -o test-reports\coverage.html 2>&1 | Out-Null
    Write-Host "[OK] Coverage report generated"

    $coverage = go tool cover -func=test-reports\coverage.out | Select-String "total:"
    if ($coverage) {
        $parts = $coverage.Line.Split()
        Write-Host "Total Coverage: $($parts[$parts.Length - 2])"
    }
}

# Summary
Write-Host ""
Write-Host "========================================"
Write-Host "Test Summary"
Write-Host "========================================"

$allPassed = $true

if ($storeResult -ne 0) {
    Write-Host "[FAIL] internal/store tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] internal/store tests" -ForegroundColor Green
}

if ($llmResult -ne 0) {
    Write-Host "[FAIL] internal/client/llm tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] internal/client/llm tests" -ForegroundColor Green
}

if ($handlerResult -ne 0) {
    Write-Host "[FAIL] internal/api/handler tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] internal/api/handler tests" -ForegroundColor Green
}

if ($integrationResult -ne 0) {
    Write-Host "[FAIL] integration tests" -ForegroundColor Red
    $allPassed = $false
} else {
    Write-Host "[PASS] integration tests" -ForegroundColor Green
}

Write-Host ""

if ($allPassed) {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit 1
}
