// Package postgres 提供 PostgreSQL 存储实现
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/dnd-mcp/client/pkg/config"
)

// Client PostgreSQL 客户端
type Client struct {
	pool *pgxpool.Pool
}

// NewClient 创建 PostgreSQL 客户端
func NewClient(cfg config.PostgresConfig) (*Client, error) {
	// 构建连接字符串
	connStr := buildConnectionString(cfg)

	// 配置连接池
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("解析连接字符串失败: %w", err)
	}

	// 设置连接池参数
	poolConfig.MaxConns = int32(cfg.PoolSize)
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Duration(cfg.MaxConnLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// 配置日志（可选，用于调试）
	poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &pgxLogger{},
		LogLevel: tracelog.LogLevelWarn, // 生产环境使用 Warn
	}

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("创建连接池失败: %w", err)
	}

	return &Client{pool: pool}, nil
}

// Ping 检查连接状态
func (c *Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Close 关闭连接
func (c *Client) Close() error {
	c.pool.Close()
	return nil
}

// Pool 返回连接池
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// buildConnectionString 构建连接字符串
func buildConnectionString(cfg config.PostgresConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
}

// pgxLogger 实现 pgx tracelog.Logger 接口
type pgxLogger struct{}

func (l *pgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	// 简单的日志实现，实际项目可以使用更完善的日志库
	// 这里为了不依赖具体日志库，暂时使用简单的日志格式
	/*
		switch level {
		case tracelog.LogLevelTrace, tracelog.LogLevelDebug, tracelog.LogLevelInfo:
			// 可以使用 log.Debug 或 log.Info
		case tracelog.LogLevelWarn:
			// 可以使用 log.Warn
		case tracelog.LogLevelError:
			// 可以使用 log.Error
		}
	*/
	_ = ctx // 暂时不使用
	_ = level
	_ = msg
	_ = data
}
