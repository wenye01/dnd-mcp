# Test-E2E.ps1 - 运行端到端测试
# 此脚本会:
# 1. 检查并清理Redis
# 2. 构建项目
# 3. 启动服务器(后台)
# 4. 运行E2E测试
# 5. 停止服务器
# 6. 清理测试数据

param(
    [switch]$SkipBuild = $false,
    [switch]$Verbose = $false
)

$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Step { Write-ColorOutput Cyan "==> $args" }
function Write-Success { Write-ColorOutput Green "✓ $args" }
function Write-Error { Write-ColorOutput Red "✗ $args" }
function Write-Warning { Write-ColorOutput Yellow "⚠ $args" }

# 保存原始环境
$originalEnv = @{}
$envVars = @("LOG_LEVEL", "REDIS_HOST", "HTTP_PORT")
foreach ($var in $envVars) {
    $originalEnv[$var] = [Environment]::GetEnvironmentVariable($var)
}

# 清理函数
function Cleanup {
    Write-Step "清理..."

    # 停止服务器
    if ($serverProcess -ne $null) {
        try {
            Stop-Process -Id $serverProcess.Id -Force -ErrorAction Stop
            Write-Success "已停止服务器"
        } catch {
            Write-Warning "停止服务器失败: $_"
        }
    }

    # 恢复环境变量
    foreach ($var in $envVars) {
        if ($originalEnv[$var] -ne $null) {
            [Environment]::SetEnvironmentVariable($var, $originalEnv[$var])
        } else {
            [Environment]::SetEnvironmentVariable($var, "")
        }
    }

    # 清理Redis测试数据
    $redisCli = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
    if (Test-Path $redisCli) {
        try {
            & $redisCli FLUSHALL > $null
            Write-Success "已清理Redis数据"
        } catch {
            Write-Warning "清理Redis失败: $_"
        }
    }
}

# 注册清理
$serverProcess = $null
trap {
    Cleanup
    exit 1
}

# 步骤 1: 检查Redis
Write-Step "步骤 1: 检查Redis连接"

$redisCli = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
if (-not (Test-Path $redisCli)) {
    Write-Error "Redis客户端未找到: $redisCli"
    exit 1
}

try {
    $result = & $redisCli PING 2>&1
    if ($LASTEXITCODE -eq 0 -and $result -eq "PONG") {
        Write-Success "Redis运行正常"
    } else {
        Write-Error "Redis未运行或连接失败"
        exit 1
    }
} catch {
    Write-Error "无法连接到Redis: $_"
    exit 1
}

# 步骤 2: 清理Redis
Write-Step "步骤 2: 清理Redis数据"

try {
    & $redisCli FLUSHALL > $null
    Write-Success "已清理Redis"
} catch {
    Write-Warning "清理Redis失败: $_"
}

# 步骤 3: 构建项目
if (-not $SkipBuild) {
    Write-Step "步骤 3: 构建项目"

    try {
        go build -o bin/dnd-api.exe ./cmd/api
        if ($LASTEXITCODE -eq 0) {
            Write-Success "构建成功"
        } else {
            Write-Error "构建失败"
            exit 1
        }
    } catch {
        Write-Error "构建时出错: $_"
        exit 1
    }
} else {
    Write-Warning "跳过构建(使用 -SkipBuild 标志)"
}

# 步骤 4: 启动服务器
Write-Step "步骤 4: 启动服务器"

$env:LOG_LEVEL = "info"
$env:REDIS_HOST = "localhost:6379"
$env:HTTP_PORT = "8080"

if (-not (Test-Path "bin\dnd-api.exe")) {
    Write-Error "可执行文件未找到: bin\dnd-api.exe"
    Write-Warning "请先运行构建或使用 -SkipBuild 标志"
    exit 1
}

try {
    $serverProcess = Start-Process -FilePath "bin\dnd-api.exe" -ArgumentList "-l", "info" -PassThru -NoNewWindow
    Write-Success "服务器已启动 (PID: $($serverProcess.Id))"

    # 等待服务器启动
    Write-Output "等待服务器启动..."
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
            # 继续等待
        }
        Start-Sleep -Seconds 1
        $waited++
        Write-Output "." -NoNewline
    }
    Write-Output ""

    if (-not $ready) {
        Write-Error "服务器未能在${maxWait}秒内启动"
        Cleanup
        exit 1
    }

    Write-Success "服务器已就绪"
} catch {
    Write-Error "启动服务器时出错: $_"
    Cleanup
    exit 1
}

# 步骤 5: 运行E2E测试
Write-Step "步骤 5: 运行E2E测试"

try {
    if ($Verbose) {
        go test -v ./tests/e2e/... -timeout 60s
    } else {
        go test ./tests/e2e/... -timeout 60s
    }

    if ($LASTEXITCODE -eq 0) {
        Write-Success "E2E测试通过"
    } else {
        Write-Error "E2E测试失败"
        Cleanup
        exit 1
    }
} catch {
    Write-Error "运行E2E测试时出错: $_"
    Cleanup
    exit 1
}

# 清理
Cleanup

Write-Success "E2E测试完成!"
Write-Output ""
Write-Output "测试总结:"
Write-Output "  - 服务器启动: ✓"
Write-Output "  - E2E测试: ✓"
Write-Output "  - 清理完成: ✓"
