# Clean up test database completely
# This script removes the dnd_mcp_test database completely

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host ""
Write-Host "========================================" -ForegroundColor Yellow
Write-Host "Database Cleanup Script" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Yellow
Write-Host ""

$env:PGPASSWORD = "070831"

Write-Host "[1/2] Dropping tables (if any exist)..." -ForegroundColor Cyan
try {
    $null = & psql -U postgres -d dnd_mcp_test -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" 2>&1
    Write-Host "[OK] Tables dropped" -ForegroundColor Green
} catch {
    Write-Host "[INFO] No tables to drop or database does not exist" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[2/2] Dropping database..." -ForegroundColor Cyan
try {
    $output = & psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS dnd_mcp_test;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Database 'dnd_mcp_test' dropped successfully" -ForegroundColor Green
    } else {
        Write-Host "[INFO] Database did not exist" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to drop database: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Cleanup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Verify
Write-Host "Verifying..." -ForegroundColor Cyan
$databases = & psql -U postgres -d postgres -c "\l" 2>&1 | Select-String "dnd_mcp_test"
if ($databases) {
    Write-Host "[WARNING] Database still exists!" -ForegroundColor Red
} else {
    Write-Host "[SUCCESS] Database completely removed" -ForegroundColor Green
}

Write-Host ""
Write-Host "Database is now in a clean state (no database, no tables)" -ForegroundColor Cyan
Write-Host "Run setup-and-test.ps1 to recreate and test" -ForegroundColor Yellow
Write-Host ""
