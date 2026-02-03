#!/bin/bash

# DND MCP Client 测试脚本

set -e

echo "🧪 开始运行测试..."

# 运行所有测试
echo "🔍 运行单元测试..."
go test -v ./tests/unit/... -cover

echo "🔍 运行集成测试..."
go test -v ./tests/integration/... -tags=integration -cover

echo "✅ 测试完成!"
