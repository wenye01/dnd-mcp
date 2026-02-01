@echo off
setlocal EnableDelayedExpansion

REM 默认目标
set TARGET=build

REM 解析命令行参数
if "%1" neq "" set TARGET=%1

REM 构建
if "%TARGET%"=="build" (
    echo Building MCP Client...
    if not exist bin mkdir bin
    go build -o bin\dnd-mcp-client.exe cmd\client\main.go
    if %errorlevel% equ 0 (
        echo Build successful: bin\dnd-mcp-client.exe
    ) else (
        echo Build failed!
        exit /b 1
    )
)

REM 运行
if "%TARGET%"=="run" (
    echo Running MCP Client...
    go run cmd\client\main.go
)

REM 测试
if "%TARGET%"=="test" (
    echo Running tests...
    go test -v ./...
)

REM 数据库迁移
if "%TARGET%"=="migrate-up" (
    echo Running database migrations...
    go run scripts\migrate\main.go -action=up
)

if "%TARGET%"=="migrate-down" (
    echo Rolling back database migrations...
    go run scripts\migrate\main.go -action=down
)

REM 依赖
if "%TARGET%"=="deps" (
    echo Downloading dependencies...
    go mod download
    go mod tidy
)

REM 代码检查
if "%TARGET%"=="lint" (
    echo Running code checks...
    go vet ./...
)

REM 格式化
if "%TARGET%"=="fmt" (
    echo Formatting code...
    go fmt ./...
)

REM 清理
if "%TARGET%"=="clean" (
    echo Cleaning build files...
    if exist bin rmdir /s /q bin
)

REM 帮助
if "%TARGET%"=="help" (
    echo Available targets:
    echo   build       - Build the application
    echo   run         - Run the application
    echo   test        - Run unit tests
    echo   migrate-up  - Run database migrations
    echo   migrate-down- Rollback database migrations
    echo   deps        - Download dependencies
    echo   lint        - Run code checks
    echo   fmt         - Format code
    echo   clean       - Clean build files
    echo   help        - Show this help
)

endlocal
