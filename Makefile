.PHONY: run build clean test help migrate-up migrate-down deps

# 默认目标
help:
	@echo "Available targets:"
	@echo "  run         - Run the server"
	@echo "  build       - Build the binary"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  deps        - Download dependencies"
	@echo "  migrate-up  - Run database migrations"
	@echo "  migrate-down- Rollback database migrations"

# 运行服务器
run:
	go run cmd/client/main.go

# 构建二进制文件
build:
	go build -o bin/dnd-mcp-client cmd/client/main.go

# 清理构建产物
clean:
	rm -rf bin/

# 运行测试
test:
	go test -v ./...

# 数据库迁移
migrate-up:
	go run scripts/migrate/main.go -action=up

migrate-down:
	go run scripts/migrate/main.go -action=down

# 下载依赖
deps:
	go mod download
	go mod tidy
