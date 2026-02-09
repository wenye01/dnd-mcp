// Package monitor 提供系统监控功能
package monitor

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// StatsCollector 统计收集器接口
type StatsCollector interface {
	Collect(ctx context.Context) map[string]interface{}
	Name() string
}

// SystemStats 系统统计信息
type SystemStats struct {
	Uptime       int64                  `json:"uptime_seconds"`
	StartTime    time.Time              `json:"start_time"`
	Version      string                 `json:"version,omitempty"`
	RequestCount int64                  `json:"request_count"`
	ErrorCount   int64                  `json:"error_count"`
	Components   map[string]interface{} `json:"components"`
}

// StatsMonitor 统计监控器
type StatsMonitor struct {
	startTime  time.Time
	version    string
	collectors map[string]StatsCollector
	mu         sync.RWMutex

	// 计数器
	requestCount atomic.Int64
	errorCount   atomic.Int64
}

// NewStatsMonitor 创建统计监控器
func NewStatsMonitor(version string) *StatsMonitor {
	return &StatsMonitor{
		startTime:  time.Now(),
		version:    version,
		collectors: make(map[string]StatsCollector),
	}
}

// Register 注册统计收集器
func (s *StatsMonitor) Register(collector StatsCollector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collectors[collector.Name()] = collector
}

// Unregister 注销统计收集器
func (s *StatsMonitor) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.collectors, name)
}

// Collect 收集统计信息
func (s *StatsMonitor) Collect(ctx context.Context) *SystemStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SystemStats{
		Uptime:       int64(time.Since(s.startTime).Seconds()),
		StartTime:    s.startTime,
		Version:      s.version,
		RequestCount: s.requestCount.Load(),
		ErrorCount:   s.errorCount.Load(),
		Components:   make(map[string]interface{}),
	}

	// 收集所有组件的统计信息
	for name, collector := range s.collectors {
		stats.Components[name] = collector.Collect(ctx)
	}

	return stats
}

// IncrementRequestCount 增加请求计数
func (s *StatsMonitor) IncrementRequestCount() {
	s.requestCount.Add(1)
}

// IncrementErrorCount 增加错误计数
func (s *StatsMonitor) IncrementErrorCount() {
	s.errorCount.Add(1)
}

// GetRequestCount 获取请求计数
func (s *StatsMonitor) GetRequestCount() int64 {
	return s.requestCount.Load()
}

// GetErrorCount 获取错误计数
func (s *StatsMonitor) GetErrorCount() int64 {
	return s.errorCount.Load()
}

// RedisStatsCollector Redis 统计收集器
type RedisStatsCollector struct {
	client RedisStatsGetter
}

// RedisStatsGetter Redis 统计获取接口
type RedisStatsGetter interface {
	DBSize(ctx context.Context) (int64, error)
	Info(ctx context.Context, section string) (string, error)
}

// NewRedisStatsCollector 创建 Redis 统计收集器
func NewRedisStatsCollector(client RedisStatsGetter) StatsCollector {
	return &RedisStatsCollector{client: client}
}

// Collect 收集 Redis 统计信息
func (c *RedisStatsCollector) Collect(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	if c.client == nil {
		stats["error"] = "Redis client not available"
		stats["available"] = false
		return stats
	}

	// 获取数据库大小（key数量）
	if dbSize, err := c.client.DBSize(ctx); err != nil {
		stats["error"] = err.Error()
		stats["available"] = false
	} else {
		stats["key_count"] = dbSize
		stats["available"] = true
	}

	// 获取 Redis 信息
	if info, err := c.client.Info(ctx, "memory"); err == nil {
		stats["memory_info"] = info
	}

	return stats
}

// Name 返回收集器名称
func (c *RedisStatsCollector) Name() string {
	return "redis"
}

// PostgreSQLStatsCollector PostgreSQL 统计收集器
type PostgreSQLStatsCollector struct {
	client PostgresStatsGetter
}

// PostgresStatsGetter PostgreSQL 统计获取接口
type PostgresStatsGetter interface {
	Ping(ctx context.Context) error
}

// NewPostgreSQLStatsCollector 创建 PostgreSQL 统计收集器
func NewPostgreSQLStatsCollector(client PostgresStatsGetter) StatsCollector {
	return &PostgreSQLStatsCollector{client: client}
}

// Collect 收集 PostgreSQL 统计信息
func (c *PostgreSQLStatsCollector) Collect(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	if c.client == nil {
		stats["error"] = "PostgreSQL client not available"
		stats["available"] = false
		return stats
	}

	if err := c.client.Ping(ctx); err != nil {
		stats["error"] = err.Error()
		stats["available"] = false
	} else {
		stats["available"] = true
	}

	return stats
}

// Name 返回收集器名称
func (c *PostgreSQLStatsCollector) Name() string {
	return "postgres"
}

// SessionStatsCollector 会话统计收集器
type SessionStatsCollector struct {
	sessionStore SessionCounter
}

// SessionCounter 会话计数接口
type SessionCounter interface {
	Count(ctx context.Context) (int64, error)
}

// NewSessionStatsCollector 创建会话统计收集器
func NewSessionStatsCollector(sessionStore SessionCounter) StatsCollector {
	return &SessionStatsCollector{sessionStore: sessionStore}
}

// Collect 收集会话统计信息
func (c *SessionStatsCollector) Collect(ctx context.Context) map[string]interface{} {
	stats := make(map[string]interface{})

	if c.sessionStore == nil {
		stats["error"] = "session store not available"
		stats["count"] = 0
		return stats
	}

	if count, err := c.sessionStore.Count(ctx); err != nil {
		stats["error"] = err.Error()
		stats["count"] = 0
	} else {
		stats["count"] = count
	}

	return stats
}

// Name 返回收集器名称
func (c *SessionStatsCollector) Name() string {
	return "sessions"
}

// MiddlewareStats 中间件统计
type MiddlewareStats struct {
	requestCount  atomic.Int64
	errorCount    atomic.Int64
	totalLatency  atomic.Int64
	mu            sync.RWMutex
	pathCounts    map[string]*atomic.Int64
	pathErrors    map[string]*atomic.Int64
	pathLatencies map[string]*atomic.Int64
}

// NewMiddlewareStats 创建中间件统计
func NewMiddlewareStats() *MiddlewareStats {
	return &MiddlewareStats{
		pathCounts:    make(map[string]*atomic.Int64),
		pathErrors:    make(map[string]*atomic.Int64),
		pathLatencies: make(map[string]*atomic.Int64),
	}
}

// RecordRequest 记录请求
func (m *MiddlewareStats) RecordRequest(path string, latency time.Duration, isError bool) {
	m.requestCount.Add(1)
	m.totalLatency.Add(int64(latency.Milliseconds()))

	if isError {
		m.errorCount.Add(1)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pathCounts[path]; !exists {
		m.pathCounts[path] = &atomic.Int64{}
		m.pathErrors[path] = &atomic.Int64{}
		m.pathLatencies[path] = &atomic.Int64{}
	}

	m.pathCounts[path].Add(1)
	m.pathLatencies[path].Add(int64(latency.Milliseconds()))
	if isError {
		m.pathErrors[path].Add(1)
	}
}

// GetStats 获取统计信息
func (m *MiddlewareStats) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requestCount := m.requestCount.Load()
	errorCount := m.errorCount.Load()
	totalLatency := m.totalLatency.Load()

	avgLatency := float64(0)
	if requestCount > 0 {
		avgLatency = float64(totalLatency) / float64(requestCount)
	}

	pathStats := make(map[string]interface{})
	for path := range m.pathCounts {
		count := m.pathCounts[path].Load()
		errors := m.pathErrors[path].Load()
		latency := m.pathLatencies[path].Load()

		pathAvgLatency := float64(0)
		if count > 0 {
			pathAvgLatency = float64(latency) / float64(count)
		}

		pathStats[path] = map[string]interface{}{
			"count":       count,
			"errors":      errors,
			"avg_latency": pathAvgLatency,
		}
	}

	return map[string]interface{}{
		"total_requests": requestCount,
		"total_errors":   errorCount,
		"avg_latency_ms": avgLatency,
		"paths":          pathStats,
	}
}
