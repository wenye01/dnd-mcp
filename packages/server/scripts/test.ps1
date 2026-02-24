# DND MCP Server Quick Test Script (Windows PowerShell)
# Runs unit tests and integration tests
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

Write-Host "Running DND MCP Server Tests..." -ForegroundColor Green

# Color functions
function Print-Success {
    param([string]$Message)
    Write-Host "  OK $Message" -ForegroundColor Green
}

function Print-Error {
    param([string]$Message)
    Write-Host "  FAIL $Message" -ForegroundColor Red
}

function Print-Section {
    param([string]$Message)
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Yellow
}

# Test results
$allPassed = $true

# Run unit tests
Print-Section "Unit Tests - Models"
$unitModels = go test -v ./tests/unit/models/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Models unit tests passed"
} else {
    Print-Error "Models unit tests failed"
    Write-Host $unitModels
    $allPassed = $false
}

Print-Section "Unit Tests - Service"
$unitService = go test -v ./tests/unit/service/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Service unit tests passed"
} else {
    Print-Error "Service unit tests failed"
    Write-Host $unitService
    $allPassed = $false
}

Print-Section "Unit Tests - MCP"
$unitMCP = go test -v ./tests/unit/mcp/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "MCP unit tests passed"
} else {
    Print-Error "MCP unit tests failed"
    Write-Host $unitMCP
    $allPassed = $false
}

Print-Section "Unit Tests - Rules"
$unitRules = go test -v ./tests/unit/rules/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Rules unit tests passed"
} else {
    Print-Error "Rules unit tests failed"
    Write-Host $unitRules
    $allPassed = $false
}

# Run integration tests (requires PostgreSQL)
Print-Section "Integration Tests - Store"
$integrationStore = go test -v ./tests/integration/store/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Store integration tests passed"
} else {
    $output = $integrationStore | Out-String
    if ($output -match "Database not available") {
        Write-Host "  SKIP Database not available" -ForegroundColor Yellow
    } else {
        Print-Error "Store integration tests failed"
        Write-Host $integrationStore
        $allPassed = $false
    }
}

Print-Section "Integration Tests - Tools"
$integrationTools = go test -v ./tests/integration/tools/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Tools integration tests passed"
} else {
    $output = $integrationTools | Out-String
    if ($output -match "Database not available") {
        Write-Host "  SKIP Database not available" -ForegroundColor Yellow
    } else {
        Print-Error "Tools integration tests failed"
        Write-Host $integrationTools
        $allPassed = $false
    }
}

Print-Section "Integration Tests - MCP"
$integrationMCP = go test -v ./tests/integration/mcp/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "MCP integration tests passed"
} else {
    Print-Error "MCP integration tests failed"
    Write-Host $integrationMCP
    $allPassed = $false
}

# Summary
Write-Host ""
if ($allPassed) {
    Write-Host "All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit 1
}
