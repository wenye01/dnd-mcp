# DND MCP Client Test Script (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "Running tests..." -ForegroundColor Green

# Run unit tests
Write-Host "Running unit tests..." -ForegroundColor Cyan
go test -v ./tests/unit/... -cover

if ($LASTEXITCODE -ne 0) {
    Write-Host "Unit tests failed!" -ForegroundColor Red
    exit 1
}

# Run integration tests
Write-Host "Running integration tests..." -ForegroundColor Cyan
$env:INTEGRATION_TEST = "1"
go test -v ./tests/integration/... -tags=integration -cover

if ($LASTEXITCODE -eq 0) {
    Write-Host "Tests completed!" -ForegroundColor Green
} else {
    Write-Host "Integration tests failed (Redis may be required)" -ForegroundColor Yellow
}
