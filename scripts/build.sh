#!/bin/bash

# DND MCP Client 构建脚本

set -e

echo "🔨 开始构建 DND MCP Client..."

# 设置变量
APP_NAME="dnd-client"
BUILD_DIR="bin"
VERSION=${VERSION:-"0.1.0"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'"

# 创建构建目录
echo "📁 创建构建目录..."
mkdir -p "$BUILD_DIR"

# 构建主程序
echo "🔨 构建主程序..."
go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/$APP_NAME" ./cmd/client

echo "✅ 构建完成!"
echo "📍 输出: $BUILD_DIR/$APP_NAME"
