@echo off
setlocal EnableDelayedExpansion

REM Default target
set TARGET=build

REM Parse command line arguments
if "%1" neq "" set TARGET=%1

REM Build
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

REM Run
if "%TARGET%"=="run" (
    echo Running MCP Client...
    go run cmd\client\main.go
)

REM Test
if "%TARGET%"=="test" (
    echo Running tests...
    go test -v ./...
)

REM Database migrations
if "%TARGET%"=="migrate-up" (
    echo Running database migrations...
    go run scripts\migrate\main.go -action=up
)

if "%TARGET%"=="migrate-down" (
    echo Rolling back database migrations...
    go run scripts\migrate\main.go -action=down
)

REM Dependencies
if "%TARGET%"=="deps" (
    echo Downloading dependencies...
    go mod download
    go mod tidy
)

REM Code check
if "%TARGET%"=="lint" (
    echo Running code checks...
    go vet ./...
)

REM Format
if "%TARGET%"=="fmt" (
    echo Formatting code...
    go fmt ./...
)

REM Clean
if "%TARGET%"=="clean" (
    echo Cleaning build files...
    if exist bin rmdir /s /q bin
)

REM Help
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
