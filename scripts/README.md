# Scripts - Build and Test Scripts

This directory contains build and test scripts for both Windows and Linux/Mac environments.

## Windows PowerShell Scripts

All Windows scripts use English to avoid encoding issues and include UTF-8 encoding declarations.

### Quick Start

```powershell
# One command to setup entire development environment
.\scripts\dev.ps1
```

### Individual Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `dev.ps1` | Complete dev environment setup | `.\scripts\dev.ps1` |
| `build.ps1` | Build the project | `.\scripts\build.ps1` |
| `test.ps1` | Run tests | `.\scripts\test.ps1` |
| `start-redis.ps1` | Start Redis container | `.\scripts\start-redis.ps1` |
| `reset-env.ps1` | Reset environment to initial state | `.\scripts\reset-env.ps1 -Force` |

### Features

- ✅ No encoding issues (uses English)
- ✅ Color-coded output
- ✅ Error handling
- ✅ Automatic dependency checking
- ✅ UTF-8 encoding support

## Linux/Mac Bash Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `build.sh` | Build the project | `./scripts/build.sh` |
| `test.sh` | Run tests | `./scripts/test.sh` |

### Usage

```bash
# Make scripts executable
chmod +x scripts/*.sh

# Build
./scripts/build.sh

# Test
./scripts/test.sh
```

## Docker Commands

### Redis Management

```powershell
# Windows
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine
docker start dnd-redis
docker stop dnd-redis
docker logs dnd-redis
```

```bash
# Linux/Mac
docker run -d --name dnd-redis -p 6379:6379 redis:7-alpine
docker start dnd-redis
docker stop dnd-redis
docker logs dnd-redis
```

## Common Issues

### PowerShell Execution Policy

**Error**: "cannot be loaded because running scripts is disabled on this system"

**Solution**:

```powershell
# Temporary (recommended)
powershell -ExecutionPolicy Bypass -File scripts\build.ps1

# Permanent
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Docker Not Running

**Error**: "error during connect" or "docker daemon is not running"

**Solution**: Start Docker Desktop

### Go Command Not Found

**Error**: 'go' is not recognized as an internal or external command

**Solution**: Add Go to system PATH or install Go from https://golang.org/dl/

## Encoding Fix

All Windows PowerShell scripts were updated to:
1. Use English instead of Chinese (avoid encoding issues)
2. Include UTF-8 encoding declaration
3. Set console output encoding

```powershell
# -*- coding: utf-8 -*-
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
```

This ensures scripts work correctly on all Windows systems regardless of locale settings.

## Environment Reset Script

### Overview

The `reset-env.ps1` script completely resets your development environment to the initial state.

### Usage

```powershell
# Interactive mode (will ask for confirmation)
.\scripts\reset-env.ps1

# Force mode (execute without confirmation)
.\scripts\reset-env.ps1 -Force

# Verbose mode (show all operation details)
.\scripts\reset-env.ps1 -Verbose

# Combined
.\scripts\reset-env.ps1 -Force -Verbose
```

### What It Does

1. **Stops Redis Services/Processes**
   - Stops all `redis-server` processes
   - Stops Redis Windows services (if any)

2. **Clears Redis Databases**
   - Executes `FLUSHALL` to clear all databases
   - Shows database size before/after cleanup

3. **Removes Build Artifacts**
   - Deletes the `bin/` directory
   - Deletes `*.exe` files in root directory

4. **Cleans Test Cache**
   - Runs `go clean -testcache`

5. **Removes Temporary Files**
   - Deletes `*.log`, `*.tmp`, `*.temp` files
   - Excludes `.git` and `node_modules` directories

6. **Cleans Redis Logs**
   - Removes Redis log files from Redis installation directory

### Use Cases

- Before starting a new development task
- When environment becomes cluttered
- Before running tests for clean results
- Before deployment

### Example Workflow

```powershell
# Step 1: Reset environment
.\scripts\reset-env.ps1 -Force

# Step 2: Start fresh development environment
.\scripts\dev.ps1

# Step 3: Test CLI
.\bin\client.exe session list
```

### Warnings

⚠️ **Data Loss**: This script will clear ALL Redis databases - this operation cannot be undone!
⚠️ **Process Termination**: All Redis processes will be stopped
⚠️ **File Deletion**: Build artifacts and temporary files will be deleted

### Troubleshooting

#### Redis won't stop
```powershell
# Manually find and stop Redis processes
Get-Process redis*
Stop-Process -Name redis-server -Force
```

#### Can't clear Redis database
```powershell
# Check if Redis is running
& 'C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe' ping

# Manually flush database
& 'C:\Tools\Redis-8.4.0-Windows-x64-msys2-with-Service\redis-cli.exe' FLUSHALL
```

#### Files can't be deleted
```powershell
# Check for running programs
Get-Process | Where-Object {$_.Path -like "*dnd-mcp*"}

# Close programs and retry
.\scripts\reset-env.ps1 -Force
```
