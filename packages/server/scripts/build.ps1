# DND MCP Server Build Script (Windows PowerShell)
# -*- coding: utf-8 -*-

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "Building DND MCP Server..." -ForegroundColor Green

# Set variables
$AppName = "dnd-server"
$BuildDir = "bin"
$Version = if ($env:VERSION) { $env:VERSION } else { "0.1.0" }
$BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
$Ldflags = "-X 'main.Version=$Version' -X 'main.BuildTime=$BuildTime'"

# Create build directory
Write-Host "Creating build directory..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

# Build main program
Write-Host "Building..." -ForegroundColor Cyan
go build -ldflags $Ldflags -o "$BuildDir\$AppName.exe" ./cmd/server

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build completed!" -ForegroundColor Green
    Write-Host "Output: $BuildDir\$AppName.exe" -ForegroundColor Cyan

    # Show file info
    $FileInfo = Get-Item "$BuildDir\$AppName.exe" -ErrorAction SilentlyContinue
    if ($FileInfo) {
        Write-Host "File size: $([math]::Round($FileInfo.Length / 1MB, 2)) MB" -ForegroundColor Cyan
    }
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
