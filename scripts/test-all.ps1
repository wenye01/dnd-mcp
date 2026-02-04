# DND MCP Complete Test Script
# Runs all tests including unit tests, integration tests, and API tests
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$ErrorActionPreference = "Stop"

# Color functions
function Print-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Print-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
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
    UnitTests = $false
    IntegrationTests = $false
    APITests = $false
    AllPassed = $false
}

# Main script
Print-Header "DND MCP Complete Test Suite"

# 1. Stop all services
Print-Section "Step 1/6: Stopping Services"
try {
    # Stop Redis
    $redisProcess = Get-Process -Name redis-server -ErrorAction SilentlyContinue
    if ($redisProcess) {
        & "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" SHUTDOWN NOSAVE 2>$null | Out-Null
        Start-Sleep -Seconds 1
        Print-Success "Redis stopped"
    } else {
        Print-Info "Redis not running"
    }

    # Stop dnd-client server
    $clientProcess = Get-Process -Name dnd-client -ErrorAction SilentlyContinue
    if ($clientProcess) {
        Stop-Process -Name dnd-client -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
        Print-Success "DND client server stopped"
    } else {
        Print-Info "DND client server not running"
    }
} catch {
    Print-Error "Error stopping services: $_"
}

# 2. Clear Redis database
Print-Section "Step 2/6: Clearing Redis Database"
try {
    & "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" FLUSHALL 2>$null | Out-Null
    Print-Success "Redis database cleared"
} catch {
    Print-Error "Failed to clear Redis: $_"
}

# 3. Start Redis
Print-Section "Step 3/6: Starting Redis"
try {
    $redisRunning = & "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe" PING 2>$null
    if ($redisRunning -eq "PONG") {
        Print-Info "Redis already running"
    } else {
        Start-Process -FilePath "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-server.exe" -WindowStyle Hidden
        Start-Sleep -Seconds 2
        Print-Success "Redis started"
    }
} catch {
    Print-Error "Failed to start Redis: $_"
}

# 4. Build project
Print-Section "Step 4/6: Building Project"
try {
    if (Test-Path "bin") {
        Remove-Item -Path "bin" -Recurse -Force -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    go build -o bin/dnd-client.exe cmd/client/main.go
    Print-Success "Build successful"
} catch {
    Print-Error "Build failed: $_"
    exit 1
}

# 5. Run unit and integration tests
Print-Section "Step 5/6: Running Unit & Integration Tests"
try {
    Write-Host ""
    Write-Host "Running Service Layer Unit Tests..." -ForegroundColor Cyan
    $unitResult = go test -v ./tests/unit/service/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Success "Service layer unit tests passed"
        $testResults.UnitTests = $true
    } else {
        Print-Error "Service layer unit tests failed"
        Write-Host $unitResult
    }

    Write-Host ""
    Write-Host "Running API Integration Tests..." -ForegroundColor Cyan
    $integrationResult = go test -v ./tests/integration/api/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Success "API integration tests passed"
        $testResults.IntegrationTests = $true
    } else {
        Print-Error "API integration tests failed"
        Write-Host $integrationResult
    }

    Write-Host ""
    Write-Host "Running Store Integration Tests..." -ForegroundColor Cyan
    $storeResult = go test -v ./tests/integration/store/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Success "Store integration tests passed"
    } else {
        Print-Error "Store integration tests failed"
        Write-Host $storeResult
    }
} catch {
    Print-Error "Test execution failed: $_"
}

# 6. Run API functional tests
Print-Section "Step 6/6: Running API Functional Tests"
try {
    # Start server in background
    Print-Info "Starting HTTP server..."
    $serverProcess = Start-Process -FilePath ".\bin\dnd-client.exe" -ArgumentList "server" -WindowStyle Hidden -PassThru
    Start-Sleep -Seconds 2

    # Test health check
    Write-Host ""
    Write-Host "Testing health endpoint..." -ForegroundColor Cyan
    try {
        $healthCheck = & curl -s http://localhost:8080/health 2>$null
        if ($healthCheck -eq '{"status":"ok"}') {
            Print-Success "Health check passed"
        } else {
            Print-Error "Health check failed"
        }
    } catch {
        Print-Error "Health check failed: $_"
    }

    # Test create session
    Write-Host ""
    Write-Host "Testing session creation..." -ForegroundColor Cyan
    try {
        $body = '{"name":"Test Session","creator_id":"testuser","mcp_server_url":"http://test.com","max_players":4}'
        $createSession = & curl -s -X POST http://localhost:8080/api/sessions `
            -H "Content-Type: application/json" `
            -d $body 2>$null

    if ($createSession -match '"id":"[a-f0-9-]+"') {
        Print-Success "Session creation passed"

        # Extract session ID
        if ($createSession -match '"id":"([a-f0-9-]+)"') {
            $sessionId = $matches[1]

            # Test get session
            Write-Host ""
            Write-Host "Testing get session..." -ForegroundColor Cyan
            try {
                $getSession = & curl -s http://localhost:8080/api/sessions/$sessionId 2>$null
                if ($getSession -match $sessionId) {
                    Print-Success "Get session passed"
                } else {
                    Print-Error "Get session failed"
                }
            } catch {
                Print-Error "Get session failed: $_"
            }

            # Test update session
            Write-Host ""
            Write-Host "Testing update session..." -ForegroundColor Cyan
            try {
                $updateBody = '{"name":"Updated Session"}'
                $updateSession = & curl -s -X PATCH http://localhost:8080/api/sessions/$sessionId `
                    -H "Content-Type: application/json" `
                    -d $updateBody 2>$null

                if ($updateSession -match 'Updated Session') {
                    Print-Success "Update session passed"
                } else {
                    Print-Error "Update session failed"
                }
            } catch {
                Print-Error "Update session failed: $_"
            }

            # Test delete session
            Write-Host ""
            Write-Host "Testing delete session..." -ForegroundColor Cyan
            try {
                $deleteSession = & curl -s -X DELETE http://localhost:8080/api/sessions/$sessionId 2>$null
                if ($deleteSession -eq "" -or $deleteSession -match '204') {
                    Print-Success "Delete session passed"
                } else {
                    Print-Error "Delete session failed"
                }
            } catch {
                Print-Error "Delete session failed: $_"
            }
        }

        # Test list sessions
        Write-Host ""
        Write-Host "Testing list sessions..." -ForegroundColor Cyan
        try {
            $listSessions = & curl -s http://localhost:8080/api/sessions 2>$null
        if ($listSessions -match '\[') {
            Print-Success "List sessions passed"
            $testResults.APITests = $true
        } else {
            Print-Error "List sessions failed"
        }
    } catch {
        Print-Error "List sessions failed: $_"
    }
    } else {
        Print-Error "Session creation failed"
    }

    # Stop server
    Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
    Print-Success "Server stopped"
} catch {
    Print-Error "API functional tests failed: $_"
    # Try to stop server
    try {
        Stop-Process -Name dnd-client -Force -ErrorAction SilentlyContinue
    } catch {}
}

# Final results
Print-Header "Test Results Summary"

Write-Host ""
Write-Host "Unit Tests:           " -NoNewline
if ($testResults.UnitTests) {
    Write-Host "PASSED" -ForegroundColor Green
} else {
    Write-Host "FAILED" -ForegroundColor Red
}

Write-Host "Integration Tests:    " -NoNewline
if ($testResults.IntegrationTests) {
    Write-Host "PASSED" -ForegroundColor Green
} else {
    Write-Host "FAILED" -ForegroundColor Red
}

Write-Host "API Functional Tests: " -NoNewline
if ($testResults.APITests) {
    Write-Host "PASSED" -ForegroundColor Green
} else {
    Write-Host "FAILED" -ForegroundColor Red
}

Write-Host ""

# Calculate overall result
$testResults.AllPassed = $testResults.UnitTests -and $testResults.IntegrationTests -and $testResults.APITests

if ($testResults.AllPassed) {
    Write-Host "Overall Result: " -NoNewline
    Write-Host "ALL TESTS PASSED ✓" -ForegroundColor Green
    Write-Host ""
    exit 0
} else {
    Write-Host "Overall Result: " -NoNewline
    Write-Host "SOME TESTS FAILED ✗" -ForegroundColor Red
    Write-Host ""
    exit 1
}
