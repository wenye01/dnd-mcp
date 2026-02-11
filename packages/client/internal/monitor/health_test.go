package monitor

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRedisPinger 模拟 Redis Pinger
type mockRedisPinger struct {
	pingErr error
}

func (m *mockRedisPinger) Ping(ctx context.Context) error {
	if m.pingErr != nil {
		return m.pingErr
	}
	time.Sleep(10 * time.Millisecond) // 模拟延迟
	return nil
}

// mockPostgresPinger 模拟 PostgreSQL Pinger
type mockPostgresPinger struct {
	pingErr error
}

func (m *mockPostgresPinger) Ping(ctx context.Context) error {
	if m.pingErr != nil {
		return m.pingErr
	}
	time.Sleep(15 * time.Millisecond) // 模拟延迟
	return nil
}

func TestHealthMonitor_Check_AllHealthy(t *testing.T) {
	monitor := NewHealthMonitor()

	// 注册健康的检查器
	redisChecker := NewRedisHealthChecker(&mockRedisPinger{pingErr: nil})
	postgresChecker := NewPostgreSQLHealthChecker(&mockPostgresPinger{pingErr: nil})

	monitor.Register(redisChecker)
	monitor.Register(postgresChecker)

	// 执行检查
	ctx := context.Background()
	result := monitor.Check(ctx)

	// 验证结果
	if result.Status != HealthStatusHealthy {
		t.Errorf("expected status %v, got %v", HealthStatusHealthy, result.Status)
	}

	if len(result.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(result.Components))
	}

	// 验证 Redis 组件
	redisHealth, ok := result.Components["redis"]
	if !ok {
		t.Fatal("expected redis component")
	}
	if redisHealth.Status != HealthStatusHealthy {
		t.Errorf("expected redis status %v, got %v", HealthStatusHealthy, redisHealth.Status)
	}
	if redisHealth.LatencyMs <= 0 {
		t.Error("expected positive latency")
	}

	// 验证 PostgreSQL 组件
	postgresHealth, ok := result.Components["postgres"]
	if !ok {
		t.Fatal("expected postgres component")
	}
	if postgresHealth.Status != HealthStatusHealthy {
		t.Errorf("expected postgres status %v, got %v", HealthStatusHealthy, postgresHealth.Status)
	}
}

func TestHealthMonitor_Check_OneUnhealthy(t *testing.T) {
	monitor := NewHealthMonitor()

	// Redis 健康，PostgreSQL 不健康
	redisChecker := NewRedisHealthChecker(&mockRedisPinger{pingErr: nil})
	postgresChecker := NewPostgreSQLHealthChecker(&mockPostgresPinger{pingErr: errors.New("connection failed")})

	monitor.Register(redisChecker)
	monitor.Register(postgresChecker)

	// 执行检查
	ctx := context.Background()
	result := monitor.Check(ctx)

	// 验证结果
	if result.Status != HealthStatusUnhealthy {
		t.Errorf("expected status %v, got %v", HealthStatusUnhealthy, result.Status)
	}

	// 验证 PostgreSQL 组件状态
	postgresHealth := result.Components["postgres"]
	if postgresHealth.Status != HealthStatusUnhealthy {
		t.Errorf("expected postgres status %v, got %v", HealthStatusUnhealthy, postgresHealth.Status)
	}
	if postgresHealth.Message == "" {
		t.Error("expected error message in postgres component")
	}
}

func TestHealthMonitor_Check_Empty(t *testing.T) {
	monitor := NewHealthMonitor()

	// 没有注册任何检查器
	ctx := context.Background()
	result := monitor.Check(ctx)

	// 验证结果
	if result.Status != HealthStatusHealthy {
		t.Errorf("expected status %v when no checkers registered, got %v", HealthStatusHealthy, result.Status)
	}

	if len(result.Components) != 0 {
		t.Errorf("expected 0 components, got %d", len(result.Components))
	}
}

func TestHealthMonitor_Register_Unregister(t *testing.T) {
	monitor := NewHealthMonitor()

	checker := NewRedisHealthChecker(&mockRedisPinger{})

	// 注册
	monitor.Register(checker)
	ctx := context.Background()
	result := monitor.Check(ctx)

	if len(result.Components) != 1 {
		t.Errorf("expected 1 component after registration, got %d", len(result.Components))
	}

	// 注销
	monitor.Unregister("redis")
	result = monitor.Check(ctx)

	if len(result.Components) != 0 {
		t.Errorf("expected 0 components after unregister, got %d", len(result.Components))
	}
}

func TestSimpleHealthChecker(t *testing.T) {
	customChecker := NewSimpleHealthChecker("custom", func(ctx context.Context) ComponentHealth {
		return ComponentHealth{
			Status:    HealthStatusHealthy,
			Message:   "custom check passed",
			LatencyMs: 5.5,
		}
	})

	ctx := context.Background()
	result := customChecker.Check(ctx)

	if result.Status != HealthStatusHealthy {
		t.Errorf("expected status %v, got %v", HealthStatusHealthy, result.Status)
	}
	if result.Message != "custom check passed" {
		t.Errorf("expected message 'custom check passed', got %v", result.Message)
	}
	if result.LatencyMs != 5.5 {
		t.Errorf("expected latency 5.5, got %v", result.LatencyMs)
	}

	if customChecker.Name() != "custom" {
		t.Errorf("expected name 'custom', got %v", customChecker.Name())
	}
}

func TestHealthStatus_String(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusDegraded, "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status.String())
			}
		})
	}
}
