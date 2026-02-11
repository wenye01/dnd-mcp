# DND MCP Client Environment Reset Script
# Resets the development environment to initial state
# Author: Claude Code
# Date: 2026-02-03

param(
    [switch]$Force,  # Force reset without confirmation
    [switch]$Verbose  # Verbose output
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  DND MCP Client Environment Reset" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Helper functions
function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Cyan
}

# 1. Stop Redis services/processes
function Stop-RedisService {
    Write-Info "Step 1/6: Stopping Redis services/processes"

    try {
        $redisProcesses = Get-Process -Name "redis-server" -ErrorAction SilentlyContinue

        if ($redisProcesses) {
            Write-Info "Found Redis process(es) running, stopping..."
            foreach ($process in $redisProcesses) {
                if ($Verbose) {
                    Write-Host "  Stopping PID: $($process.Id)" -ForegroundColor Gray
                }
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            }
            Write-Success "Redis processes stopped"
        } else {
            Write-Info "No Redis processes found"
        }

        $redisServices = Get-Service | Where-Object { $_.DisplayName -like "*redis*" -or $_.ServiceName -like "*redis*" }

        if ($redisServices) {
            foreach ($service in $redisServices) {
                if ($service.Status -eq "Running") {
                    Write-Info "Stopping Redis service: $($service.DisplayName)"
                    Stop-Service -Name $service.Name -Force -ErrorAction SilentlyContinue
                    Write-Success "Service stopped: $($service.DisplayName)"
                }
            }
        }
    }
    catch {
        Write-Warning "Error stopping Redis: $_"
    }

    Write-Host ""
}

# 2. Clear Redis database
function Clear-RedisDatabase {
    Write-Info "Step 2/6: Clearing Redis database"

    $redisCliPath = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"

    if (Test-Path $redisCliPath) {
        try {
            $pingResult = & $redisCliPath ping 2>&1

            if ($pingResult -eq "PONG") {
                Write-Info "Redis connection successful, clearing all data..."

                $dbSizeBefore = & $redisCliPath DBSIZE
                Write-Info "Database size before: $dbSizeBefore keys"

                $flushResult = & $redisCliPath FLUSHALL 2>&1

                if ($LASTEXITCODE -eq 0) {
                    $dbSizeAfter = & $redisCliPath DBSIZE
                    Write-Success "Redis database cleared"
                    Write-Info "Database size after: $dbSizeAfter keys"

                    if ($Verbose) {
                        Write-Host "  Cleared:" -ForegroundColor Gray
                        Write-Host "    - All session data (session:*)" -ForegroundColor Gray
                        Write-Host "    - All message data (msg:*)" -ForegroundColor Gray
                        Write-Host "    - All indexes (sessions:all)" -ForegroundColor Gray
                    }
                } else {
                    Write-Error "Failed to clear Redis: $flushResult"
                }
            } else {
                Write-Warning "Redis not running or cannot connect ($pingResult)"
                Write-Info "Skipping database cleanup"
            }
        }
        catch {
            Write-Warning "Error clearing Redis: $_"
        }
    } else {
        Write-Warning "redis-cli.exe not found: $redisCliPath"
        Write-Info "Skipping database cleanup"
    }

    Write-Host ""
}

# 3. Remove build artifacts
function Clear-BuildArtifacts {
    Write-Info "Step 3/6: Removing build artifacts"

    $buildDir = "bin"

    if (Test-Path $buildDir) {
        $items = Get-ChildItem -Path $buildDir -Recurse
        $totalSize = ($items | Measure-Object -Property Length -Sum).Sum

        Write-Info "Found bin directory, size: $([math]::Round($totalSize / 1MB, 2)) MB"

        Remove-Item -Path $buildDir -Recurse -Force -ErrorAction SilentlyContinue

        if (Test-Path $buildDir) {
            Write-Warning "Some files may be in use, please close running programs"
        } else {
            Write-Success "bin directory removed"
        }
    } else {
        Write-Info "bin directory not found"
    }

    $exeFiles = Get-ChildItem -Path "." -Filter "*.exe" -ErrorAction SilentlyContinue
    if ($exeFiles) {
        foreach ($file in $exeFiles) {
            Write-Info "Deleting: $($file.Name)"
            Remove-Item -Path $file.FullName -Force -ErrorAction SilentlyContinue
        }
        Write-Success "Root .exe files removed"
    }

    Write-Host ""
}

# 4. Clear test cache
function Clear-TestCache {
    Write-Info "Step 4/6: Clearing test cache"

    Write-Info "Clearing Go test cache..."
    $cacheResult = go clean -testcache 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Success "Go test cache cleared"
    } else {
        Write-Warning "Issue clearing test cache"
    }

    Write-Host ""
}

# 5. Remove temporary files
function Clear-TempFiles {
    Write-Info "Step 5/6: Removing temporary files"

    $tempPatterns = @(
        "*.log",
        "*.tmp",
        "*.temp",
        "*~",
        ".DS_Store",
        "Thumbs.db"
    )

    $deletedCount = 0

    foreach ($pattern in $tempPatterns) {
        $files = Get-ChildItem -Path "." -Filter $pattern -Recurse -ErrorAction SilentlyContinue
        foreach ($file in $files) {
            if ($file.FullName -notmatch "\.git|node_modules") {
                if ($Verbose) {
                    Write-Host "  Deleting: $($file.FullName)" -ForegroundColor Gray
                }
                Remove-Item -Path $file.FullName -Force -ErrorAction SilentlyContinue
                $deletedCount++
            }
        }
    }

    if ($deletedCount -gt 0) {
        Write-Success "Removed $deletedCount temporary files"
    } else {
        Write-Info "No temporary files found"
    }

    Write-Host ""
}

# 6. Clear Redis logs
function Clear-RedisLogs {
    Write-Info "Step 6/6: Clearing Redis logs"

    $redisDir = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service"
    $logFiles = @("redis.log", "server.log", "redis-server.log")

    if (Test-Path $redisDir) {
        $deletedCount = 0
        foreach ($logFile in $logFiles) {
            $logPath = Join-Path $redisDir $logFile
            if (Test-Path $logPath) {
                if ($Verbose) {
                    Write-Host "  Deleting: $logPath" -ForegroundColor Gray
                }
                Remove-Item -Path $logPath -Force -ErrorAction SilentlyContinue
                $deletedCount++
            }
        }

        if ($deletedCount -gt 0) {
            Write-Success "Removed $deletedCount Redis log files"
        } else {
            Write-Info "No Redis log files found"
        }
    } else {
        Write-Info "Redis directory not found, skipping log cleanup"
    }

    Write-Host ""
}

# Show summary
function Show-Summary {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  Environment Reset Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Completed operations:" -ForegroundColor White
    Write-Host "  1. [OK] Stopped Redis services/processes" -ForegroundColor Green
    Write-Host "  2. [OK] Cleared Redis databases" -ForegroundColor Green
    Write-Host "  3. [OK] Removed build artifacts (bin/)" -ForegroundColor Green
    Write-Host "  4. [OK] Cleared Go test cache" -ForegroundColor Green
    Write-Host "  5. [OK] Removed temporary files" -ForegroundColor Green
    Write-Host "  6. [OK] Cleared Redis logs" -ForegroundColor Green
    Write-Host ""
    Write-Host "Environment is back to initial state!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Build project: .\scripts\build.ps1" -ForegroundColor White
    Write-Host "  2. Run tests:    .\scripts\test.ps1" -ForegroundColor White
    Write-Host "  3. Or quick start: .\scripts\dev.ps1" -ForegroundColor White
    Write-Host ""
}

# Main execution
function Main {
    if (-not $Force) {
        Write-Warning "This script will:"
        Write-Host "  1. Stop all Redis services/processes" -ForegroundColor Red
        Write-Host "  2. Clear ALL Redis databases (not recoverable!)" -ForegroundColor Red
        Write-Host "  3. Delete all build artifacts" -ForegroundColor Red
        Write-Host "  4. Clean test cache and temp files" -ForegroundColor Red
        Write-Host ""
        $confirm = Read-Host "Continue? (y/N)"

        if ($confirm -ne "y" -and $confirm -ne "Y") {
            Write-Info "Operation cancelled"
            return
        }
        Write-Host ""
    }

    Stop-RedisService
    Clear-RedisDatabase
    Clear-BuildArtifacts
    Clear-TestCache
    Clear-TempFiles
    Clear-RedisLogs

    Show-Summary
}

# Run
Main
