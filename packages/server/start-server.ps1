# MCP Server 启动脚本
$env:POSTGRES_HOST="localhost"
$env:POSTGRES_PORT="5432"
$env:POSTGRES_USER="postgres"
$env:POSTGRES_PASSWORD="postgres"
$env:POSTGRES_DBNAME="dnd_server"
$env:POSTGRES_SSLMODE="disable"
$env:HTTP_HOST="0.0.0.0"
$env:HTTP_PORT="8081"
$env:LOG_LEVEL="info"
$env:LOG_FORMAT="text"

Write-Host "Starting MCP Server on port 8081..."
Write-Host "Environment variables set."
Start-Process -FilePath "./bin/dnd-server.exe" -ArgumentList "-server" -WindowStyle Normal
Write-Host "Server started."
