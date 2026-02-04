// Package config_test 提供配置管理单元测试
package config_test

import (
	"os"
	"testing"

	"github.com/dnd-mcp/client/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	// 保存原始环境变量
	origHost := os.Getenv("REDIS_HOST")
	origPassword := os.Getenv("REDIS_PASSWORD")
	origDB := os.Getenv("REDIS_DB")
	origPoolSize := os.Getenv("REDIS_POOL_SIZE")
	origMinIdle := os.Getenv("REDIS_MIN_IDLE_CONNS")
	origLogLevel := os.Getenv("LOG_LEVEL")
	origLogFormat := os.Getenv("LOG_FORMAT")
	origServerHost := os.Getenv("SERVER_HOST")
	origServerPort := os.Getenv("SERVER_PORT")

	// 测试结束后恢复环境变量
	defer func() {
		os.Setenv("REDIS_HOST", origHost)
		os.Setenv("REDIS_PASSWORD", origPassword)
		os.Setenv("REDIS_DB", origDB)
		os.Setenv("REDIS_POOL_SIZE", origPoolSize)
		os.Setenv("REDIS_MIN_IDLE_CONNS", origMinIdle)
		os.Setenv("LOG_LEVEL", origLogLevel)
		os.Setenv("LOG_FORMAT", origLogFormat)
		os.Setenv("SERVER_HOST", origServerHost)
		os.Setenv("SERVER_PORT", origServerPort)
	}()

	// 清空环境变量,使用默认值
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("REDIS_POOL_SIZE")
	os.Unsetenv("REDIS_MIN_IDLE_CONNS")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证默认值
	if cfg.Redis.Host != "localhost:6379" {
		t.Errorf("Redis.Host 默认值错误: got %s, want localhost:6379", cfg.Redis.Host)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Redis.DB 默认值错误: got %d, want 0", cfg.Redis.DB)
	}
	if cfg.Redis.PoolSize != 10 {
		t.Errorf("Redis.PoolSize 默认值错误: got %d, want 10", cfg.Redis.PoolSize)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level 默认值错误: got %s, want info", cfg.Log.Level)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port 默认值错误: got %d, want 8080", cfg.Server.Port)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// 保存原始环境变量
	origHost := os.Getenv("REDIS_HOST")
	origDB := os.Getenv("REDIS_DB")
	origPoolSize := os.Getenv("REDIS_POOL_SIZE")

	// 测试结束后恢复环境变量
	defer func() {
		os.Setenv("REDIS_HOST", origHost)
		os.Setenv("REDIS_DB", origDB)
		os.Setenv("REDIS_POOL_SIZE", origPoolSize)
	}()

	// 设置环境变量
	os.Setenv("REDIS_HOST", "redis.example.com:6380")
	os.Setenv("REDIS_DB", "2")
	os.Setenv("REDIS_POOL_SIZE", "20")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证环境变量值
	if cfg.Redis.Host != "redis.example.com:6380" {
		t.Errorf("Redis.Host 环境变量错误: got %s, want redis.example.com:6380", cfg.Redis.Host)
	}
	if cfg.Redis.DB != 2 {
		t.Errorf("Redis.DB 环境变量错误: got %d, want 2", cfg.Redis.DB)
	}
	if cfg.Redis.PoolSize != 20 {
		t.Errorf("Redis.PoolSize 环境变量错误: got %d, want 20", cfg.Redis.PoolSize)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "localhost:6379",
					DB:       0,
					PoolSize: 10,
				},
				Postgres: config.PostgresConfig{
					Host:            "localhost:5432",
					User:            "dnd",
					Password:        "password",
					DBName:          "dnd_client",
					PoolSize:        5,
					MaxConnLifetime: 3600,
					MaxConnIdleTime: 1800,
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Server: config.ServerConfig{
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "无效的 Redis Host",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "",
					DB:       0,
					PoolSize: 10,
				},
				Postgres: config.PostgresConfig{
					Host:            "localhost:5432",
					User:            "dnd",
					Password:        "password",
					DBName:          "dnd_client",
					PoolSize:        5,
					MaxConnLifetime: 3600,
					MaxConnIdleTime: 1800,
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Server: config.ServerConfig{
					Port: 8080,
				},
			},
			wantErr: true,
		},
		{
			name: "无效的 Pool Size",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "localhost:6379",
					DB:       0,
					PoolSize: -1,
				},
				Postgres: config.PostgresConfig{
					Host:            "localhost:5432",
					User:            "dnd",
					Password:        "password",
					DBName:          "dnd_client",
					PoolSize:        5,
					MaxConnLifetime: 3600,
					MaxConnIdleTime: 1800,
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Server: config.ServerConfig{
					Port: 8080,
				},
			},
			wantErr: true,
		},
		{
			name: "无效的 Log Level",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "localhost:6379",
					DB:       0,
					PoolSize: 10,
				},
				Postgres: config.PostgresConfig{
					Host:            "localhost:5432",
					User:            "dnd",
					Password:        "password",
					DBName:          "dnd_client",
					PoolSize:        5,
					MaxConnLifetime: 3600,
					MaxConnIdleTime: 1800,
				},
				Log: config.LogConfig{
					Level: "invalid",
				},
				Server: config.ServerConfig{
					Port: 8080,
				},
			},
			wantErr: true,
		},
		{
			name: "无效的 Server Port",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:     "localhost:6379",
					DB:       0,
					PoolSize: 10,
				},
				Postgres: config.PostgresConfig{
					Host:            "localhost:5432",
					User:            "dnd",
					Password:        "password",
					DBName:          "dnd_client",
					PoolSize:        5,
					MaxConnLifetime: 3600,
					MaxConnIdleTime: 1800,
				},
				Log: config.LogConfig{
					Level: "info",
				},
				Server: config.ServerConfig{
					Port: 70000,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
