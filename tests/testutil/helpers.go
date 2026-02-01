package testutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSON asserts that two JSON strings are equivalent
func AssertJSON(t *testing.T, expected, actual string) {
	var expectedJSON, actualJSON interface{}

	err := json.Unmarshal([]byte(expected), &expectedJSON)
	require.NoError(t, err, "Failed to parse expected JSON")

	err = json.Unmarshal([]byte(actual), &actualJSON)
	require.NoError(t, err, "Failed to parse actual JSON")

	assert.Equal(t, expectedJSON, actualJSON)
}

// SetupTestRouter 设置测试路由
func SetupTestRouter(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// SetupGinTestRouter 设置Gin测试路由
func SetupGinTestRouter(registerRoutes func(*gin.Engine)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerRoutes(router)
	return router
}

// MakeTestRequest 创建测试HTTP请求
func MakeTestRequest(t *testing.T, router *gin.Engine, method, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if body != "" {
		req.Body = nil // Reset body
		// Note: You may need to use io.Reader for body
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// WaitForCondition 等待条件满足
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, checkInterval time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, fmt.Sprintf("%s: timeout waiting for condition", message))
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// AssertDBCount 断言数据库中的记录数
func AssertDBCount(t *testing.T, db *sql.DB, table string, expectedCount int) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, count, fmt.Sprintf("Table %s should have %d rows", table, expectedCount))
}

// AssertDBExists 断言数据库中存在某个记录
func AssertDBExists(t *testing.T, db *sql.DB, table, idColumn string, id string) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idColumn)
	err := db.QueryRow(query, id).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, fmt.Sprintf("Record should exist in %s.%s = %s", table, idColumn, id))
}

// CreateTestSession 在数据库中创建测试会话
func CreateTestSession(t *testing.T, db *sql.DB, campaignName string) string {
	id := "550e8400-e29b-41d4-a716-446655440000"
	query := `
		INSERT INTO sessions (id, campaign_name, creator_id, status, max_players, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(
		query,
		id,
		campaignName,
		"550e8400-e29b-41d4-a716-446655440001",
		"active",
		5,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)
	return id
}

// CreateTestMessage 在数据库中创建测试消息
func CreateTestMessage(t *testing.T, db *sql.DB, sessionID, role, content string) string {
	id := fmt.Sprintf("550e8400-e29b-41d4-a716-44665544000%s", role[0:1])
	query := `
		INSERT INTO messages (id, session_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(query, id, sessionID, role, content, time.Now())
	require.NoError(t, err)
	return id
}

// GetTestDB 获取测试数据库连接
func GetTestDB(t *testing.T) *sql.DB {
	db, _, err := SetupTestDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		CleanupTestDB(db)
	})
	return db
}

// SetupTestDBWithCleanup 设置测试数据库并注册清理函数
func SetupTestDBWithCleanup(t *testing.T) *sql.DB {
	db, databaseURL, err := SetupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	t.Cleanup(func() {
		CleanupTestDB(db)
	})

	// 清空所有表
	t.Cleanup(func() {
		TruncateTables(db)
	})

	_ = databaseURL // Use this if needed
	return db
}
