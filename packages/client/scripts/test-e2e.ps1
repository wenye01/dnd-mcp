# Test-E2E.ps1 - Run E2E tests
# This script will:
# 1. Load .env file
# 2. Check and clean Redis
# 3. Build project
# 4. Start server (background)
# 5. Run E2E tests
# 6. Stop server
# 7. Clean up test data

param(
    [switch]$SkipBuild = $false,
    [switch]$Verbose = $false
)

$ErrorActionPreference = "Stop"

# Color output functions
function Write-Step { Write-Host "==> $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[OK] $args" -ForegroundColor Green }
function Write-Error { Write-Host "[FAIL] $args" -ForegroundColor Red }
function Write-Warning { Write-Host "[WARN] $args" -ForegroundColor Yellow }

# Step 0: Load .env file
Write-Step "Step 0: Loading .env file"

$envFile = ".env"
if (-not (Test-Path $envFile)) {
    Write-Error ".env file not found: $envFile"
    Write-Output "Please create .env file from .env.example"
    exit 1
}

try {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        # Skip empty lines and comments
        if ($line -and -not $line.StartsWith("#")) {
            $idx = $line.IndexOf("=")
            if ($idx -gt 0) {
                $key = $line.Substring(0, $idx).Trim()
                $value = $line.Substring($idx + 1).Trim()
                [Environment]::SetEnvironmentVariable($key, $value)
            }
        }
    }
    Write-Success ".env file loaded"
    Write-Output "  LLM_PROVIDER: $env:LLM_PROVIDER"
    Write-Output "  LLM_MODEL: $env:LLM_MODEL"
    Write-Output "  LLM_BASE_URL: $env:LLM_BASE_URL"
} catch {
    Write-Error "Failed to load .env file: $_"
    exit 1
}

# Validate required environment variables
if (-not $env:LLM_API_KEY) {
    Write-Error "LLM_API_KEY not set in .env file"
    exit 1
}

# Save original environment
$originalEnv = @{}
$envVars = @("LOG_LEVEL", "REDIS_HOST", "HTTP_PORT", "LLM_PROVIDER", "LLM_BASE_URL", "LLM_MODEL", "LLM_API_KEY", "SERVER_URL", "MCP_SERVER_URL")
foreach ($var in $envVars) {
    $originalEnv[$var] = [Environment]::GetEnvironmentVariable($var)
}

# Server process
$serverProcess = $null

# Cleanup function
function Cleanup {
    Write-Step "Cleaning up..."

    # Stop server
    if ($serverProcess -ne $null) {
        try {
            Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
            Write-Success "Server stopped"
        } catch {
            Write-Warning "Failed to stop server: $_"
        }
    }

    # Restore environment variables
    foreach ($var in $envVars) {
        if ($originalEnv[$var] -ne $null) {
            [Environment]::SetEnvironmentVariable($var, $originalEnv[$var])
        } else {
            [Environment]::SetEnvironmentVariable($var, $null)
        }
    }

    # Clean Redis test data
    $redisCli = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
    if (Test-Path $redisCli) {
        try {
            & $redisCli FLUSHALL > $null 2>&1
            Write-Success "Redis data cleaned"
        } catch {
            Write-Warning "Failed to clean Redis: $_"
        }
    }
}

# Register cleanup
trap {
    Cleanup
    exit 1
}

# Step 1: Check Redis
Write-Step "Step 1: Checking Redis connection"

$redisCli = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
if (-not (Test-Path $redisCli)) {
    Write-Error "Redis CLI not found: $redisCli"
    exit 1
}

try {
    $result = & $redisCli PING 2>&1
    if ($LASTEXITCODE -eq 0 -and $result -eq "PONG") {
        Write-Success "Redis is running"
    } else {
        Write-Error "Redis not running or connection failed"
        exit 1
    }
} catch {
    Write-Error "Cannot connect to Redis: $_"
    exit 1
}

# Step 2: Clean Redis
Write-Step "Step 2: Cleaning Redis data"

try {
    & $redisCli FLUSHALL > $null 2>&1
    Write-Success "Redis cleaned"
} catch {
    Write-Warning "Failed to clean Redis: $_"
}

# Step 3: Build project
if (-not $SkipBuild) {
    Write-Step "Step 3: Building project"

    try {
        $buildOutput = go build -o bin/dnd-api.exe ./cmd/api 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Build successful"
        } else {
            Write-Error "Build failed: $buildOutput"
            exit 1
        }
    } catch {
        Write-Error "Build error: $_"
        exit 1
    }
} else {
    Write-Warning "Skipping build (using -SkipBuild flag)"
}

# Step 4: Start server
Write-Step "Step 4: Starting server"

# Set additional environment variables for server
$env:LOG_LEVEL = "info"
$env:REDIS_HOST = "localhost:6379"
$env:HTTP_PORT = "8080"
$env:SERVER_URL = "mock://"
$env:MCP_SERVER_URL = "mock://"

if (-not (Test-Path "bin\dnd-api.exe")) {
    Write-Error "Executable not found: bin\dnd-api.exe"
    Write-Warning "Please run build first or use -SkipBuild flag"
    exit 1
}

try {
    $serverProcess = Start-Process -FilePath "bin\dnd-api.exe" -PassThru -NoNewWindow
    Write-Success "Server started (PID: $($serverProcess.Id))"

    # Wait for server to start
    Write-Output "Waiting for server to start..."
    $maxWait = 10
    $waited = 0
    $ready = $false

    while ($waited -lt $maxWait) {
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/api/system/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                $ready = $true
                break
            }
        } catch {
            # Continue waiting
        }
        Start-Sleep -Seconds 1
        $waited++
        Write-Output "." -NoNewline
    }
    Write-Output ""

    if (-not $ready) {
        Write-Error "Server failed to start within ${maxWait} seconds"
        Cleanup
        exit 1
    }

    Write-Success "Server is ready"
} catch {
    Write-Error "Failed to start server: $_"
    Cleanup
    exit 1
}

# Step 5: Run E2E tests
Write-Step "Step 5: Running E2E tests"

try {
    if ($Verbose) {
        go test -v ./tests/e2e/... -timeout 120s
    } else {
        go test ./tests/e2e/... -timeout 120s
    }

    if ($LASTEXITCODE -eq 0) {
        Write-Success "E2E tests passed"
    } else {
        Write-Error "E2E tests failed"
        Cleanup
        exit 1
    }
} catch {
    Write-Error "Error running E2E tests: $_"
    Cleanup
    exit 1
}

# Cleanup
Cleanup

Write-Success "E2E tests completed!"
Write-Output ""
Write-Output "Test summary:"
Write-Output "  - Server started: OK"
Write-Output "  - E2E tests: OK"
Write-Output "  - Cleanup: OK"
