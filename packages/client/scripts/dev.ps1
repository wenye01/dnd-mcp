# DND MCP API - Quick Start Development Environment (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host ""
Write-Host "DND MCP API - Development Environment" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check Go installation
Write-Host "Checking development environment..." -ForegroundColor Cyan

try {
    $goVersion = go version
    Write-Host "  Go: $goVersion" -ForegroundColor Green
}
catch {
    Write-Host "  ERROR: Go is not installed!" -ForegroundColor Red
    Write-Host "     Download from: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

# Check Redis installation
$RedisPath = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-server.exe"
if (Test-Path $RedisPath) {
    Write-Host "  Redis: Found at $RedisPath" -ForegroundColor Green
}
else {
    Write-Host "  WARNING: Redis not found!" -ForegroundColor Yellow
    Write-Host "     Please install Redis or update path in this script" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Starting Redis..." -ForegroundColor Cyan

# Start Redis using local installation
& $PSScriptRoot\start-redis.ps1 -Action start

Write-Host ""
Write-Host "Building project..." -ForegroundColor Cyan

# Download dependencies
Write-Host "  Downloading dependencies..." -ForegroundColor Yellow
go mod tidy | Out-Null

# Build project
Write-Host "  Compiling..." -ForegroundColor Yellow
go build -o bin/dnd-api.exe ./cmd/api

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Build successful!" -ForegroundColor Green
}
else {
    Write-Host "  Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Running tests..." -ForegroundColor Cyan

$testResult = go test ./pkg/... ./internal/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  Tests passed!" -ForegroundColor Green
}
else {
    Write-Host "  Some tests failed" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Development environment ready!" -ForegroundColor Green
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor Cyan
Write-Host "  Start API:   .\bin\dnd-api.exe" -ForegroundColor White
Write-Host "  Start Redis: .\scripts\start-redis.ps1" -ForegroundColor White
Write-Host "  Build:       .\scripts\build.ps1" -ForegroundColor White
Write-Host "  Test:        .\scripts\test.ps1" -ForegroundColor White
Write-Host ""
Write-Host "API Endpoints:" -ForegroundColor Cyan
Write-Host "  Health:      http://localhost:8080/api/system/health" -ForegroundColor White
Write-Host "  Stats:       http://localhost:8080/api/system/stats" -ForegroundColor White
Write-Host "  Sessions:    http://localhost:8080/api/sessions" -ForegroundColor White
Write-Host ""
Write-Host "Documentation:" -ForegroundColor Cyan
Write-Host "  README:      README.md" -ForegroundColor White
Write-Host "  Design Doc:  ../../docs/client/详细设计.md" -ForegroundColor White
Write-Host ""
