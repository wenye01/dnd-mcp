# DND MCP Client - Quick Start Development Environment (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host ""
Write-Host "DND MCP Client - Development Environment" -ForegroundColor Green
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

# Check Docker installation
try {
    $dockerVersion = docker --version
    Write-Host "  Docker: $dockerVersion" -ForegroundColor Green
}
catch {
    Write-Host "  ERROR: Docker is not installed!" -ForegroundColor Red
    Write-Host "     Download from: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Starting Redis..." -ForegroundColor Cyan

# Start Redis
$existingContainer = docker ps -a --filter "name=dnd-redis" --format "{{.Names}}" 2>$null

if ($existingContainer -eq "dnd-redis") {
    Write-Host "  Starting existing Redis container..." -ForegroundColor Yellow
    docker start dnd-redis | Out-Null
    Write-Host "  Redis started" -ForegroundColor Green
}
else {
    Write-Host "  Creating new Redis container..." -ForegroundColor Yellow
    docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine | Out-Null
    Write-Host "  Redis created and started" -ForegroundColor Green
}

Write-Host ""
Write-Host "Building project..." -ForegroundColor Cyan

# Download dependencies
Write-Host "  Downloading dependencies..." -ForegroundColor Yellow
go mod tidy | Out-Null

# Build project
Write-Host "  Compiling..." -ForegroundColor Yellow
go build ./cmd/client

if ($LASTEXITCODE -eq 0) {
    Write-Host "  Build successful!" -ForegroundColor Green
}
else {
    Write-Host "  Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Running tests..." -ForegroundColor Cyan

$testResult = go test ./tests/unit/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  Unit tests passed!" -ForegroundColor Green
}
else {
    Write-Host "  Unit tests failed" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Development environment ready!" -ForegroundColor Green
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor Cyan
Write-Host "  Help:        .\client.exe --help" -ForegroundColor White
Write-Host "  Start Redis: .\scripts\start-redis.ps1" -ForegroundColor White
Write-Host "  Build:       .\scripts\build.ps1" -ForegroundColor White
Write-Host "  Test:        .\scripts\test.ps1" -ForegroundColor White
Write-Host ""
Write-Host "Examples:" -ForegroundColor Cyan
Write-Host "  Create session: .\client.exe session create --name 'Test' --creator 'user-123' --mcp-url 'http://localhost:9000'" -ForegroundColor White
Write-Host "  List sessions: .\client.exe session list" -ForegroundColor White
Write-Host ""
Write-Host "Documentation:" -ForegroundColor Cyan
Write-Host "  User Guide:      doc\usage-guide.md" -ForegroundColor White
Write-Host "  Task Summary:    doc\task-summary.md" -ForegroundColor White
Write-Host "  Windows Guide:   doc\windows-guide.md" -ForegroundColor White
Write-Host ""
