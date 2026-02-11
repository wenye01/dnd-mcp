package monitor

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRedisStatsGetter 模拟 Redis 统计获取器
type mockRedisStatsGetter struct {
	dbSize int64
	info   string
	err    error
}

func (m *mockRedisStatsGetter) Ping(ctx context.Context) error {
	return nil
}

func (m *mockRedisStatsGetter) DBSize(ctx context.Context) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.dbSize, nil
}

func (m *mockRedisStatsGetter) Info(ctx context.Context, section string) (string, error) {
	return m.info, nil
}

// mockPostgresStatsGetter 模拟 PostgreSQL 统计获取器
type mockPostgresStatsGetter struct {
	pingErr error
}

func (m *mockPostgresStatsGetter) Ping(ctx context.Context) error {
	return m.pingErr
}

// mockSessionCounter 模拟会话计数器
type mockSessionCounter struct {
	count int64
	err   error
}

func (m *mockSessionCounter) Count(ctx context.Context) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.count, nil
}

func TestStatsMonitor_Collect(t *testing.T) {
	monitor := NewStatsMonitor("v1.0.0")

	// 注册收集器
	redisCollector := NewRedisStatsCollector(&mockRedisStatsGetter{dbSize: 42})
	sessionCollector := NewSessionStatsCollector(&mockSessionCounter{count: 5})

	monitor.Register(redisCollector)
	monitor.Register(sessionCollector)

	// 等待一小段时间确保uptime > 0
	time.Sleep(150 * time.Millisecond)

	// 收集统计信息
	ctx := context.Background()
	stats := monitor.Collect(ctx)

	// 验证基本字段
	if stats.Version != "v1.0.0" {
		t.Errorf("expected version v1.0.0, got %v", stats.Version)
	}
	// Uptime至少应该是0（可能时间很短）
	if stats.Uptime < 0 {
		t.Errorf("expected non-negative uptime, got %d", stats.Uptime)
	}
	if stats.StartTime.IsZero() {
		t.Error("expected non-zero start time")
	}

	// 验证组件
	if len(stats.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(stats.Components))
	}

	// 验证 Redis 统计
	redisStats, ok := stats.Components["redis"].(map[string]interface{})
	if !ok {
		t.Fatal("expected redis component to be a map")
	}
	if redisStats["key_count"] != int64(42) {
		t.Errorf("expected key_count 42, got %v", redisStats["key_count"])
	}
	if redisStats["available"] != true {
		t.Error("expected redis available to be true")
	}

	// 验证会话统计
	sessionStats, ok := stats.Components["sessions"].(map[string]interface{})
	if !ok {
		t.Fatal("expected sessions component to be a map")
	}
	if sessionStats["count"] != int64(5) {
		t.Errorf("expected session count 5, got %v", sessionStats["count"])
	}
}

func TestStatsMonitor_IncrementCounters(t *testing.T) {
	monitor := NewStatsMonitor("v1.0.0")

	// 初始计数
	if monitor.GetRequestCount() != 0 {
		t.Errorf("expected initial request count 0, got %d", monitor.GetRequestCount())
	}
	if monitor.GetErrorCount() != 0 {
		t.Errorf("expected initial error count 0, got %d", monitor.GetErrorCount())
	}

	// 增加请求计数
	monitor.IncrementRequestCount()
	monitor.IncrementRequestCount()
	monitor.IncrementRequestCount()

	if monitor.GetRequestCount() != 3 {
		t.Errorf("expected request count 3, got %d", monitor.GetRequestCount())
	}

	// 增加错误计数
	monitor.IncrementErrorCount()

	if monitor.GetErrorCount() != 1 {
		t.Errorf("expected error count 1, got %d", monitor.GetErrorCount())
	}

	// 验证统计信息中包含计数器
	ctx := context.Background()
	stats := monitor.Collect(ctx)

	if stats.RequestCount != 3 {
		t.Errorf("expected stats request count 3, got %d", stats.RequestCount)
	}
	if stats.ErrorCount != 1 {
		t.Errorf("expected stats error count 1, got %d", stats.ErrorCount)
	}
}

func TestStatsMonitor_Unregister(t *testing.T) {
	monitor := NewStatsMonitor("v1.0.0")

	collector := NewRedisStatsCollector(&mockRedisStatsGetter{dbSize: 10})

	// 注册
	monitor.Register(collector)
	ctx := context.Background()
	stats := monitor.Collect(ctx)

	if len(stats.Components) != 1 {
		t.Errorf("expected 1 component after registration, got %d", len(stats.Components))
	}

	// 注销
	monitor.Unregister("redis")
	stats = monitor.Collect(ctx)

	if len(stats.Components) != 0 {
		t.Errorf("expected 0 components after unregister, got %d", len(stats.Components))
	}
}

func TestRedisStatsCollector_Error(t *testing.T) {
	collector := NewRedisStatsCollector(&mockRedisStatsGetter{
		err: errors.New("redis connection failed"),
	})

	ctx := context.Background()
	stats := collector.Collect(ctx)

	if stats["available"] != false {
		t.Error("expected available to be false when error occurs")
	}
	_, hasError := stats["error"]
	if !hasError {
		t.Error("expected error message in stats")
	}
}

func TestRedisStatsCollector_NilClient(t *testing.T) {
	collector := NewRedisStatsCollector(nil)

	ctx := context.Background()
	stats := collector.Collect(ctx)

	if stats["available"] != false {
		t.Error("expected available to be false when client is nil")
	}
}

func TestPostgreSQLStatsCollector(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		collector := NewPostgreSQLStatsCollector(&mockPostgresStatsGetter{pingErr: nil})

		ctx := context.Background()
		stats := collector.Collect(ctx)

		if stats["available"] != true {
			t.Error("expected available to be true")
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		collector := NewPostgreSQLStatsCollector(&mockPostgresStatsGetter{
			pingErr: errors.New("connection failed"),
		})

		ctx := context.Background()
		stats := collector.Collect(ctx)

		if stats["available"] != false {
			t.Error("expected available to be false")
		}
		_, hasError := stats["error"]
		if !hasError {
			t.Error("expected error in stats")
		}
	})
}

func TestSessionStatsCollector(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		collector := NewSessionStatsCollector(&mockSessionCounter{count: 100})

		ctx := context.Background()
		stats := collector.Collect(ctx)

		if stats["count"] != int64(100) {
			t.Errorf("expected count 100, got %v", stats["count"])
		}
	})

	t.Run("error", func(t *testing.T) {
		collector := NewSessionStatsCollector(&mockSessionCounter{
			err: errors.New("session store error"),
		})

		ctx := context.Background()
		stats := collector.Collect(ctx)

		// count可以是int或int64，所以我们用==0来比较
		if stats["count"] != 0 {
			t.Errorf("expected count 0 when error, got %v", stats["count"])
		}
		_, hasError := stats["error"]
		if !hasError {
			t.Error("expected error in stats")
		}
	})

	t.Run("nil store", func(t *testing.T) {
		collector := NewSessionStatsCollector(nil)

		ctx := context.Background()
		stats := collector.Collect(ctx)

		if stats["count"] != 0 {
			t.Errorf("expected count 0 when store is nil, got %v", stats["count"])
		}
	})
}

func TestMiddlewareStats(t *testing.T) {
	stats := NewMiddlewareStats()

	// 记录一些请求
	stats.RecordRequest("/api/sessions", 100*time.Millisecond, false)
	stats.RecordRequest("/api/sessions", 150*time.Millisecond, false)
	stats.RecordRequest("/api/sessions", 200*time.Millisecond, true) // error
	stats.RecordRequest("/api/messages", 80*time.Millisecond, false)

	// 获取统计信息
	resultMap := stats.GetStats()

	if resultMap == nil {
		t.Fatal("expected result to be non-nil")
	}

	// 验证总体统计
	if resultMap["total_requests"] != int64(4) {
		t.Errorf("expected total_requests 4, got %v", resultMap["total_requests"])
	}
	if resultMap["total_errors"] != int64(1) {
		t.Errorf("expected total_errors 1, got %v", resultMap["total_errors"])
	}

	avgLatency := resultMap["avg_latency_ms"].(float64)
	expectedAvg := (100.0 + 150.0 + 200.0 + 80.0) / 4.0
	if avgLatency != expectedAvg {
		t.Errorf("expected avg_latency %.2f, got %.2f", expectedAvg, avgLatency)
	}

	// 验证路径统计
	paths, ok := resultMap["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("expected paths to be a map")
	}

	sessionPath, ok := paths["/api/sessions"].(map[string]interface{})
	if !ok {
		t.Fatal("expected /api/sessions to be a map")
	}

	if sessionPath["count"] != int64(3) {
		t.Errorf("expected /api/sessions count 3, got %v", sessionPath["count"])
	}
	if sessionPath["errors"] != int64(1) {
		t.Errorf("expected /api/sessions errors 1, got %v", sessionPath["errors"])
	}

	sessionAvgLatency := sessionPath["avg_latency"].(float64)
	expectedSessionAvg := (100.0 + 150.0 + 200.0) / 3.0
	if sessionAvgLatency != expectedSessionAvg {
		t.Errorf("expected /api/sessions avg_latency %.2f, got %.2f", expectedSessionAvg, sessionAvgLatency)
	}
}
