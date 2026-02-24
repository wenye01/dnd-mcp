# DND MCP Server Complete Test Script (Windows PowerShell)
# Runs full test suite with coverage
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

# Color functions
function Print-Success {
    param([string]$Message)
    Write-Host "  OK $Message" -ForegroundColor Green
}

function Print-Error {
    param([string]$Message)
    Write-Host "  FAIL $Message" -ForegroundColor Red
}

function Print-Info {
    param([string]$Message)
    Write-Host "  $Message" -ForegroundColor Cyan
}

function Print-Section {
    param([string]$Message)
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Yellow
}

function Print-Header {
    param([string]$Message)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  $Message" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
}

# Test results tracking
$testResults = @{
    Build = $false
    UnitModels = $false
    UnitService = $false
    UnitMCP = $false
    UnitRules = $false
    IntegrationStore = $false
    IntegrationTools = $false
    IntegrationMCP = $false
}

Print-Header "DND MCP Server Complete Test Suite"

# Step 1: Clean
Print-Section "Step 1/7: Cleaning Build Artifacts"
try {
    if (Test-Path "bin") {
        Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue
        Print-Info "Removed bin/"
    }
    go clean -cache -testcache
    Print-Success "Go cache cleaned"
} catch {
    Print-Error "Failed to clean: $_"
}

# Step 2: Build
Print-Section "Step 2/7: Building Project"
try {
    New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    go build -o bin/dnd-server.exe ./cmd/server
    if ($LASTEXITCODE -eq 0) {
        Print-Success "Build successful"
        $testResults.Build = $true
    } else {
        Print-Error "Build failed"
        exit 1
    }
} catch {
    Print-Error "Build failed: $_"
    exit 1
}

# Step 3: Unit Tests - Models
Print-Section "Step 3/7: Unit Tests - Models"
$unitModels = go test -v -cover ./tests/unit/models/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Models unit tests passed"
    $testResults.UnitModels = $true
} else {
    Print-Error "Models unit tests failed"
    Write-Host $unitModels
}

# Step 4: Unit Tests - Service
Print-Section "Step 4/7: Unit Tests - Service"
$unitService = go test -v -cover ./tests/unit/service/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Service unit tests passed"
    $testResults.UnitService = $true
} else {
    Print-Error "Service unit tests failed"
    Write-Host $unitService
}

# Step 5: Unit Tests - MCP & Rules
Print-Section "Step 5/7: Unit Tests - MCP & Rules"
$unitMCP = go test -v -cover ./tests/unit/mcp/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "MCP unit tests passed"
    $testResults.UnitMCP = $true
} else {
    Print-Error "MCP unit tests failed"
    Write-Host $unitMCP
}

$unitRules = go test -v -cover ./tests/unit/rules/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Rules unit tests passed"
    $testResults.UnitRules = $true
} else {
    Print-Error "Rules unit tests failed"
    Write-Host $unitRules
}

# Step 6: Integration Tests
Print-Section "Step 6/7: Integration Tests"

# Store integration tests (requires PostgreSQL)
$integrationStore = go test -v -cover ./tests/integration/store/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Store integration tests passed"
    $testResults.IntegrationStore = $true
} else {
    $output = $integrationStore | Out-String
    if ($output -match "Database not available") {
        Write-Host "  SKIP Store tests (PostgreSQL not available)" -ForegroundColor Yellow
        $testResults.IntegrationStore = $true  # Skip is OK
    } else {
        Print-Error "Store integration tests failed"
        Write-Host $integrationStore
    }
}

# Tools integration tests (requires PostgreSQL)
$integrationTools = go test -v -cover ./tests/integration/tools/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "Tools integration tests passed"
    $testResults.IntegrationTools = $true
} else {
    $output = $integrationTools | Out-String
    if ($output -match "Database not available") {
        Write-Host "  SKIP Tools tests (PostgreSQL not available)" -ForegroundColor Yellow
        $testResults.IntegrationTools = $true  # Skip is OK
    } else {
        Print-Error "Tools integration tests failed"
        Write-Host $integrationTools
    }
}

# MCP integration tests
$integrationMCP = go test -v -cover ./tests/integration/mcp/... -timeout 60s 2>&1
if ($LASTEXITCODE -eq 0) {
    Print-Success "MCP integration tests passed"
    $testResults.IntegrationMCP = $true
} else {
    Print-Error "MCP integration tests failed"
    Write-Host $integrationMCP
}

# Step 7: Coverage Report
Print-Section "Step 7/7: Coverage Report"
try {
    go test -coverprofile=coverage.out ./tests/unit/... ./tests/integration/... 2>&1 | Out-Null
    if (Test-Path "coverage.out") {
        $coverageOutput = go tool cover -func=coverage.out 2>&1
        $totalLine = $coverageOutput | Select-String "total:"
        if ($totalLine) {
            Print-Info "Coverage: $totalLine"
        }
    }
} catch {
    Print-Info "Could not generate coverage report"
}

# Summary
Print-Header "Test Results Summary"

Write-Host "Build:              " -NoNewline
if ($testResults.Build) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Unit - Models:      " -NoNewline
if ($testResults.UnitModels) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Unit - Service:     " -NoNewline
if ($testResults.UnitService) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Unit - MCP:         " -NoNewline
if ($testResults.UnitMCP) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Unit - Rules:       " -NoNewline
if ($testResults.UnitRules) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Integration - Store:" -NoNewline
if ($testResults.IntegrationStore) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Integration - Tools:" -NoNewline
if ($testResults.IntegrationTools) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host "Integration - MCP:  " -NoNewline
if ($testResults.IntegrationMCP) { Write-Host "PASSED" -ForegroundColor Green } else { Write-Host "FAILED" -ForegroundColor Red }

Write-Host ""

$allPassed = $testResults.Values -notcontains $false
if ($allPassed) {
    Write-Host "Overall: ALL TESTS PASSED" -ForegroundColor Green
    Write-Host ""
    exit 0
} else {
    Write-Host "Overall: SOME TESTS FAILED" -ForegroundColor Red
    Write-Host ""
    exit 1
}
