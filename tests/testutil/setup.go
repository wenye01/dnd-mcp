package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnv 测试环境
type TestEnv struct {
	DB             *sql.DB
	DatabaseURL    string
	MCPServerURL   string // Mock MCP Server URL (如果启动)
	LLMServerURL   string // Mock LLM Server URL (如果启动)
	Containers     []testcontainers.Container
}

// SetupTestDB 设置测试数据库
// 使用环境变量 DATABASE_URL 或创建临时的测试数据库
func SetupTestDB() (*sql.DB, string, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	// 如果没有设置环境变量，使用默认值
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/dnd_mcp_test?sslmode=disable"
	}

	// 测试数据库连接
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, "", fmt.Errorf("failed to ping database: %w", err)
	}

	// 运行迁移（如果需要）
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, "", fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, databaseURL, nil
}

// SetupTestDBWithContainer 使用TestContainers创建测试数据库
func SetupTestDBWithContainer() (*sql.DB, string, error) {
	ctx := context.Background()

	// 启动PostgreSQL容器
	req := testcontainers.ContainerRequest{
		Image:        "postgres:18.1",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %w", err)
	}

	// 获取数据库连接信息
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container port: %w", err)
	}

	databaseURL := fmt.Sprintf("postgres://test:test@%s:%s/test_db?sslmode=disable", host, port.Port())

	// 等待数据库就绪
	time.Sleep(2 * time.Second)

	// 连接数据库
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		container.Terminate(ctx)
		db.Close()
		return nil, "", fmt.Errorf("failed to ping database: %w", err)
	}

	// 运行迁移
	if err := runMigrations(db); err != nil {
		container.Terminate(ctx)
		db.Close()
		return nil, "", fmt.Errorf("failed to run migrations: %w", err)
	}

	// 存储container引用以便后续清理
	env := &TestEnv{
		DB:         db,
		DatabaseURL: databaseURL,
		Containers: []testcontainers.Container{container},
	}

	return db, databaseURL, nil
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// CleanupTestEnv 清理测试环境（包括容器）
func CleanupTestEnv(env *TestEnv) {
	if env.DB != nil {
		env.DB.Close()
	}

	ctx := context.Background()
	for _, container := range env.Containers {
		if err := container.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate container: %v\n", err)
		}
	}
}

// runMigrations 运行数据库迁移
func runMigrations(db *sql.DB) error {
	// 创建sessions表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY,
			campaign_name VARCHAR(255) NOT NULL,
			creator_id UUID NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			max_players INTEGER NOT NULL DEFAULT 5,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// 创建messages表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY,
			session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			player_id UUID,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	return nil
}

// TruncateTables 清空所有测试表
func TruncateTables(db *sql.DB) error {
	ctx := context.Background()

	// 由于外键约束，需要先删除messages
	_, err := db.ExecContext(ctx, "TRUNCATE TABLE messages CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate messages: %w", err)
	}

	_, err = db.ExecContext(ctx, "TRUNCATE TABLE sessions CASCADE")
	if err != nil {
		return fmt.Errorf("failed to truncate sessions: %w", err)
	}

	return nil
}
