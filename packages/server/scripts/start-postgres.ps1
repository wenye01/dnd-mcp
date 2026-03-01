# Start PostgreSQL for Development (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

param(
    [string]$Action = "start",
    [switch]$Reset
)

$ContainerName = "dnd-postgres"
$PostgresUser = "postgres"
$PostgresPassword = "postgres"
$PostgresPort = 5432

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Test-Docker {
    try {
        docker --version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Start-PostgresContainer {
    Write-Step "Checking PostgreSQL container..."

    $existingContainer = docker ps -a --filter "name=$ContainerName" --format "{{.Names}}"

    if ($existingContainer -eq $ContainerName) {
        $runningContainer = docker ps --filter "name=$ContainerName" --format "{{.Names}}"

        if ($runningContainer -eq $ContainerName) {
            Write-Host "PostgreSQL container is already running" -ForegroundColor Green
        }
        else {
            if ($Reset) {
                Write-Host "Reset flag set, removing existing container..." -ForegroundColor Yellow
                docker rm -f $ContainerName | Out-Null
                return Create-PostgresContainer
            }
            Write-Host "Starting existing PostgreSQL container..." -ForegroundColor Cyan
            docker start $ContainerName | Out-Null
            Write-Host "PostgreSQL container started" -ForegroundColor Green
        }
    }
    else {
        return Create-PostgresContainer
    }

    return $true
}

function Create-PostgresContainer {
    Write-Host "Creating and starting new PostgreSQL container..." -ForegroundColor Cyan

    docker run -d `
        --name $ContainerName `
        -e POSTGRES_USER=$PostgresUser `
        -e POSTGRES_PASSWORD=$PostgresPassword `
        -p "${PostgresPort}:5432" `
        postgres:16-alpine

    if ($LASTEXITCODE -eq 0) {
        Write-Host "PostgreSQL container created and started" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "Failed to create PostgreSQL container" -ForegroundColor Red
        return $false
    }
}

function Stop-PostgresContainer {
    Write-Step "Stopping PostgreSQL container..."
    docker stop $ContainerName 2>$null | Out-Null
    Write-Host "PostgreSQL container stopped" -ForegroundColor Green
}

function Remove-PostgresContainer {
    Write-Step "Removing PostgreSQL container..."
    docker rm -f $ContainerName 2>$null | Out-Null
    Write-Host "PostgreSQL container removed" -ForegroundColor Green
}

function Test-PostgresConnection {
    Write-Step "Testing PostgreSQL connection..."

    Start-Sleep -Seconds 2

    $maxRetries = 10
    $retryCount = 0

    while ($retryCount -lt $maxRetries) {
        try {
            $result = docker exec $ContainerName pg_isready -U $PostgresUser 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "PostgreSQL connection successful!" -ForegroundColor Green
                Write-Host "PostgreSQL address: localhost:$PostgresPort" -ForegroundColor Cyan
                Write-Host "Username: $PostgresUser" -ForegroundColor Cyan
                return $true
            }
        }
        catch {
            # Continue retrying
        }

        $retryCount++
        Write-Host "Waiting for PostgreSQL to be ready... ($retryCount/$maxRetries)" -ForegroundColor Yellow
        Start-Sleep -Seconds 1
    }

    Write-Host "PostgreSQL connection test failed after $maxRetries attempts" -ForegroundColor Red
    return $false
}

function Show-Status {
    Write-Step "PostgreSQL Status"

    $runningContainer = docker ps --filter "name=$ContainerName" --format "{{.Names}}"

    if ($runningContainer -eq $ContainerName) {
        Write-Host "Status: Running" -ForegroundColor Green
        $info = docker inspect $ContainerName --format "{{.NetworkSettings.IPAddress}} {{.State.Status}}"
        Write-Host "Info: $info" -ForegroundColor Cyan
    }
    else {
        $existingContainer = docker ps -a --filter "name=$ContainerName" --format "{{.Names}}"
        if ($existingContainer -eq $ContainerName) {
            Write-Host "Status: Stopped" -ForegroundColor Yellow
        }
        else {
            Write-Host "Status: Not Created" -ForegroundColor Red
        }
    }
}

# Main logic
Write-Host "=====================================" -ForegroundColor Green
Write-Host "  DND MCP PostgreSQL Manager" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green

if (-not (Test-Docker)) {
    Write-Host "Docker is not installed. Please install Docker Desktop" -ForegroundColor Red
    Write-Host "Download: https://www.docker.com/products/docker-desktop" -ForegroundColor Cyan
    exit 1
}

switch ($Action.ToLower()) {
    "start" {
        if ((Start-PostgresContainer) -and (Test-PostgresConnection)) {
            Write-Host ""
            Write-Host "Common commands:" -ForegroundColor Cyan
            Write-Host "  Stop:     .\scripts\start-postgres.ps1 -Action stop" -ForegroundColor White
            Write-Host "  Status:   .\scripts\start-postgres.ps1 -Action status" -ForegroundColor White
            Write-Host "  Remove:   .\scripts\start-postgres.ps1 -Action remove" -ForegroundColor White
            Write-Host "  Reset:    .\scripts\start-postgres.ps1 -Reset" -ForegroundColor White
            Write-Host "  Connect:  docker exec -it $ContainerName psql -U $PostgresUser" -ForegroundColor White
        }
        else {
            Write-Host "PostgreSQL startup failed" -ForegroundColor Red
            exit 1
        }
    }
    "stop" {
        Stop-PostgresContainer
    }
    "remove" {
        Remove-PostgresContainer
    }
    "status" {
        Show-Status
    }
    "reset" {
        Remove-PostgresContainer
        Start-Sleep -Seconds 1
        Start-PostgresContainer
        Test-PostgresConnection
    }
    default {
        Write-Host "Unknown action: $Action" -ForegroundColor Red
        Write-Host "Usage: .\scripts\start-postgres.ps1 -Action [start|stop|remove|status|reset]" -ForegroundColor Cyan
        exit 1
    }
}
