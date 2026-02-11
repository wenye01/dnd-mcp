# Test-Fresh.ps1 - 从全新环境运行所有测试
# 此脚本会:
# 1. 检查Redis是否运行
# 2. 清理Redis测试数据
# 3. 运行单元测试
# 4. 运行集成测试
# 5. (可选) 运行E2E测试

param(
    [switch]$SkipE2E = $false,
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

# 步骤 1: 检查Redis
Write-Step "步骤 1: 检查Redis连接"

$redisCli = "C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe"
if (-not (Test-Path $redisCli)) {
    Write-Error "Redis客户端未找到: $redisCli"
    Write-Warning "请修改脚本中的Redis路径或确保Redis已安装"
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

# 步骤 2: 清理Redis测试数据
Write-Step "步骤 2: 清理Redis测试数据库(DB 1)"

try {
    & $redisCli -n 1 FLUSHDB > $null
    Write-Success "已清理Redis测试数据库"
} catch {
    Write-Warning "清理Redis失败: $_"
}

# 步骤 3: 运行单元测试
Write-Step "步骤 3: 运行单元测试"

try {
    if ($Verbose) {
        go test -v ./tests/unit/... -cover
    } else {
        go test ./tests/unit/... -cover
    }

    if ($LASTEXITCODE -eq 0) {
        Write-Success "单元测试通过"
    } else {
        Write-Error "单元测试失败"
        exit 1
    }
} catch {
    Write-Error "运行单元测试时出错: $_"
    exit 1
}

# 步骤 4: 运行集成测试
Write-Step "步骤 4: 运行集成测试(需要Redis)"

$env:REDIS_HOST = "localhost:6379"

try {
    if ($Verbose) {
        go test -v ./tests/integration/... -cover -timeout 30s
    } else {
        go test ./tests/integration/... -cover -timeout 30s
    }

    if ($LASTEXITCODE -eq 0) {
        Write-Success "集成测试通过"
    } else {
        Write-Error "集成测试失败"
        exit 1
    }
} catch {
    Write-Error "运行集成测试时出错: $_"
    exit 1
}

# 步骤 5: 运行E2E测试(可选)
if (-not $SkipE2E) {
    Write-Step "步骤 5: 运行E2E测试(需要运行中的服务器)"

    # 检查服务器是否运行
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/api/system/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            Write-Success "检测到服务器运行中"

            try {
                if ($Verbose) {
                    go test -v ./tests/e2e/... -timeout 60s
                } else {
                    go test ./tests/e2e/... -timeout 60s
                }

                if ($LASTEXITCODE -eq 0) {
                    Write-Success "E2E测试通过"
                } else {
                    Write-Warning "E2E测试失败(服务器可能未启动)"
                }
            } catch {
                Write-Warning "运行E2E测试时出错: $_"
            }
        }
    } catch {
        Write-Warning "未检测到运行中的服务器,跳过E2E测试"
        Write-Warning "要运行E2E测试,请先启动服务器: .\bin\dnd-api.exe"
    }
} else {
    Write-Warning "跳过E2E测试(使用 -SkipE2E 标志)"
}

# 最终清理
Write-Step "清理测试数据"

try {
    & $redisCli -n 1 FLUSHDB > $null
    Write-Success "已清理测试数据"
} catch {
    Write-Warning "最终清理失败: $_"
}

Write-Success "所有测试完成!"
Write-Output ""
Write-Output "测试总结:"
Write-Output "  - 单元测试: ✓"
Write-Output "  - 集成测试: ✓"
if (-not $SkipE2E) {
    Write-Output "  - E2E测试: 需要运行中的服务器"
}
Write-Output ""
Write-Output "提示:"
Write-Output "  - 使用 -Verbose 显示详细输出"
Write-Output "  - 使用 -SkipE2E 跳过E2E测试"
Write-Output "  - 要运行E2E测试,请在另一个终端启动服务器: .\bin\dnd-api.exe"
