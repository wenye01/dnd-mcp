// Package monitor 提供系统监控功能
package monitor

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus 健康状态
type HealthStatus string

const (
	// HealthStatusHealthy 健康
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy 不健康
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusDegraded 降级
	HealthStatusDegraded HealthStatus = "degraded"
)

// String 返回健康状态的字符串表示
func (h HealthStatus) String() string {
	return string(h)
}

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	LatencyMs float64      `json:"latency_ms,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check(ctx context.Context) ComponentHealth
	Name() string
}

// HealthMonitor 健康监控器
type HealthMonitor struct {
	checkers map[string]HealthChecker
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		checkers: make(map[string]HealthChecker),
	}
}

// Register 注册健康检查器
func (h *HealthMonitor) Register(checker HealthChecker) {
	h.checkers[checker.Name()] = checker
}

// Unregister 注销健康检查器
func (h *HealthMonitor) Unregister(name string) {
	delete(h.checkers, name)
}

// Check 执行健康检查
func (h *HealthMonitor) Check(ctx context.Context) *HealthResponse {
	start := time.Now()

	response := &HealthResponse{
		Timestamp:  start,
		Components: make(map[string]ComponentHealth),
	}

	// 默认状态
	overallStatus := HealthStatusHealthy

	// 检查所有组件
	for name, checker := range h.checkers {
		health := checker.Check(ctx)
		response.Components[name] = health

		// 更新整体状态
		if health.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if health.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	response.Status = overallStatus
	return response
}

// SimpleHealthChecker 简单健康检查器
type SimpleHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) ComponentHealth
}

// NewSimpleHealthChecker 创建简单健康检查器
func NewSimpleHealthChecker(name string, checkFunc func(ctx context.Context) ComponentHealth) HealthChecker {
	return &SimpleHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

// Check 执行检查
func (c *SimpleHealthChecker) Check(ctx context.Context) ComponentHealth {
	if c.checkFunc == nil {
		return ComponentHealth{
			Status:  HealthStatusHealthy,
			Message: "no check function",
		}
	}
	return c.checkFunc(ctx)
}

// Name 返回检查器名称
func (c *SimpleHealthChecker) Name() string {
	return c.name
}

// RedisHealthChecker Redis 健康检查器
type RedisHealthChecker struct {
	client RedisPinger
}

// RedisPinger Redis Ping 接口
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// NewRedisHealthChecker 创建 Redis 健康检查器
func NewRedisHealthChecker(client RedisPinger) HealthChecker {
	return NewSimpleHealthChecker("redis", func(ctx context.Context) ComponentHealth {
		start := time.Now()
		err := client.Ping(ctx)
		latency := float64(time.Since(start).Milliseconds())

		if err != nil {
			return ComponentHealth{
				Status:    HealthStatusUnhealthy,
				Message:   fmt.Sprintf("Redis ping failed: %v", err),
				LatencyMs: latency,
			}
		}

		return ComponentHealth{
			Status:    HealthStatusHealthy,
			Message:   "Redis connection OK",
			LatencyMs: latency,
		}
	})
}

// PostgreSQLHealthChecker PostgreSQL 健康检查器
type PostgreSQLHealthChecker struct {
	client PostgresPinger
}

// PostgresPinger PostgreSQL Ping 接口
type PostgresPinger interface {
	Ping(ctx context.Context) error
}

// NewPostgreSQLHealthChecker 创建 PostgreSQL 健康检查器
func NewPostgreSQLHealthChecker(client PostgresPinger) HealthChecker {
	return NewSimpleHealthChecker("postgres", func(ctx context.Context) ComponentHealth {
		start := time.Now()
		err := client.Ping(ctx)
		latency := float64(time.Since(start).Milliseconds())

		if err != nil {
			return ComponentHealth{
				Status:    HealthStatusUnhealthy,
				Message:   fmt.Sprintf("PostgreSQL ping failed: %v", err),
				LatencyMs: latency,
			}
		}

		return ComponentHealth{
			Status:    HealthStatusHealthy,
			Message:   "PostgreSQL connection OK",
			LatencyMs: latency,
		}
	})
}
