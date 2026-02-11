// Package redis 提供 Redis 客户端实现
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/dnd-mcp/client/pkg/config"
	"github.com/redis/go-redis/v9"
)

// Client Redis 客户端接口
type Client interface {
	// Ping 检查连接状态
	Ping(ctx context.Context) error

	// Close 关闭连接
	Close() error

	// Client 返回原生客户端
	Client() *redis.Client

	// DBSize 获取数据库大小（key数量）
	DBSize(ctx context.Context) (int64, error)

	// Info 获取 Redis 信息
	Info(ctx context.Context, section string) (string, error)
}

// redisClient Redis 客户端实现
type redisClient struct {
	client *redis.Client
}

// NewClient 创建 Redis 客户端
func NewClient(cfg *config.RedisConfig) (Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Host,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis 连接失败: %w", err)
	}

	return &redisClient{client: client}, nil
}

// Ping 检查连接状态
func (c *redisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close 关闭连接
func (c *redisClient) Close() error {
	return c.client.Close()
}

// Client 返回原生客户端
func (c *redisClient) Client() *redis.Client {
	return c.client
}

// DBSize 获取数据库大小（key数量）
func (c *redisClient) DBSize(ctx context.Context) (int64, error) {
	return c.client.DBSize(ctx).Result()
}

// Info 获取 Redis 信息
func (c *redisClient) Info(ctx context.Context, section string) (string, error) {
	return c.client.Info(ctx, section).Result()
}
