# Drop Database Script
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$env:PGPASSWORD = "070831"

Write-Host "Dropping database dnd_mcp_test..." -ForegroundColor Cyan

try {
    $output = & psql -U postgres -d postgres -c "DROP DATABASE IF EXISTS dnd_mcp_test;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Database dropped successfully" -ForegroundColor Green
    } else {
        Write-Host "[INFO] Database did not exist" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
