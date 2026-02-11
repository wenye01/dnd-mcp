# Start Redis for Development (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "Starting Redis for development..." -ForegroundColor Green

# Check if Docker is installed
try {
    docker --version | Out-Null
    Write-Host "Docker is installed" -ForegroundColor Green
}
catch {
    Write-Host "Docker is not installed. Please install Docker Desktop" -ForegroundColor Red
    Write-Host "Download: https://www.docker.com/products/docker-desktop" -ForegroundColor Cyan
    exit 1
}

# Check if Redis container exists
$existingContainer = docker ps -a --filter "name=dnd-redis" --format "{{.Names}}"

if ($existingContainer -eq "dnd-redis") {
    Write-Host "Redis container exists, starting..." -ForegroundColor Cyan
    docker start dnd-redis | Out-Null
    Write-Host "Redis container started" -ForegroundColor Green
}
else {
    Write-Host "Creating and starting new Redis container..." -ForegroundColor Cyan
    docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine
    Write-Host "Redis container created and started" -ForegroundColor Green
}

# Test connection
Write-Host "Testing Redis connection..." -ForegroundColor Cyan
Start-Sleep -Seconds 2

try {
    docker exec dnd-redis redis-cli ping | Out-Null
    Write-Host "Redis connection successful!" -ForegroundColor Green
    Write-Host "Redis address: localhost:6379" -ForegroundColor Cyan
}
catch {
    Write-Host "Redis connection test failed, but container should be running" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Common commands:" -ForegroundColor Cyan
Write-Host "  Stop Redis:  docker stop dnd-redis" -ForegroundColor White
Write-Host "  Start Redis: docker start dnd-redis" -ForegroundColor White
Write-Host "  Remove:      docker rm -f dnd-redis" -ForegroundColor White
Write-Host "  Logs:        docker logs dnd-redis" -ForegroundColor White
