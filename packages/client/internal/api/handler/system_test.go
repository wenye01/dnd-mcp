package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dnd-mcp/client/internal/monitor"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockPersistenceManager 模拟持久化管理器接口
type mockPersistenceManager struct {
	triggerErr error
}

func (m *mockPersistenceManager) Trigger(ctx context.Context) error {
	return m.triggerErr
}

// 确保mock实现了必要的接口
var _ interface{ Trigger(context.Context) error } = (*mockPersistenceManager)(nil)

// mockHealthChecker 模拟健康检查器
type mockHealthChecker struct {
	name   string
	status monitor.HealthStatus
}

func (m *mockHealthChecker) Check(ctx context.Context) monitor.ComponentHealth {
	return monitor.ComponentHealth{
		Status:  m.status,
		Message: m.name + " is " + string(m.status),
	}
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func TestSystemHandler_Health(t *testing.T) {
	t.Run("healthy system", func(t *testing.T) {
		healthMonitor := monitor.NewHealthMonitor()
		healthMonitor.Register(&mockHealthChecker{name: "component1", status: monitor.HealthStatusHealthy})
		healthMonitor.Register(&mockHealthChecker{name: "component2", status: monitor.HealthStatusHealthy})

		handler := NewSystemHandler(nil, healthMonitor, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/system/health", nil)

		handler.Health(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		// 验证响应包含必要的字段
		body := w.Body.String()
		if body == "" {
			t.Error("expected non-empty response body")
		}
	})

	t.Run("unhealthy system", func(t *testing.T) {
		healthMonitor := monitor.NewHealthMonitor()
		healthMonitor.Register(&mockHealthChecker{name: "component1", status: monitor.HealthStatusUnhealthy})

		handler := NewSystemHandler(nil, healthMonitor, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/system/health", nil)

		handler.Health(c)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
	})

	t.Run("no health monitor configured", func(t *testing.T) {
		handler := NewSystemHandler(nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/system/health", nil)

		handler.Health(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestSystemHandler_Stats(t *testing.T) {
	t.Run("normal stats", func(t *testing.T) {
		statsMonitor := monitor.NewStatsMonitor("v1.0.0")
		statsMonitor.IncrementRequestCount()
		statsMonitor.IncrementRequestCount()
		statsMonitor.IncrementErrorCount()

		handler := NewSystemHandler(nil, nil, statsMonitor)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/system/stats", nil)

		handler.Stats(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if body == "" {
			t.Error("expected non-empty response body")
		}
	})

	t.Run("no stats monitor configured", func(t *testing.T) {
		handler := NewSystemHandler(nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/system/stats", nil)

		handler.Stats(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestSystemHandler_TriggerPersistence(t *testing.T) {
	t.Run("successful trigger", func(t *testing.T) {
		persistenceManager := &mockPersistenceManager{triggerErr: nil}
		handler := NewSystemHandler(persistenceManager, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/system/persistence/trigger", nil)

		handler.TriggerPersistence(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("trigger error", func(t *testing.T) {
		persistenceManager := &mockPersistenceManager{
			triggerErr: errors.New("persistence failed"),
		}
		handler := NewSystemHandler(persistenceManager, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/system/persistence/trigger", nil)

		handler.TriggerPersistence(c)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("no persistence manager configured", func(t *testing.T) {
		handler := NewSystemHandler(nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/system/persistence/trigger", nil)

		handler.TriggerPersistence(c)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
	})
}
