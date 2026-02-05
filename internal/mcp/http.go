// Package mcp 提供 HTTP MCP Client 实现
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient HTTP MCP 客户端
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	sessionID  string
	timeout    time.Duration
}

// NewHTTPClient 创建 HTTP MCP 客户端
func NewHTTPClient(baseURL string, timeoutSeconds int) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// Initialize 初始化 MCP 连接
func (c *HTTPClient) Initialize(ctx context.Context, sessionID, serverURL string) error {
	c.sessionID = sessionID
	c.baseURL = serverURL

	// 调用 MCP Server 的 initialize 接口
	req := map[string]interface{}{
		"session_id":  sessionID,
		"capabilities": []string{"tools", "events"},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/initialize",
		bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送初始化请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("初始化失败 (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CallTool 调用工具
func (c *HTTPClient) CallTool(ctx context.Context, sessionID string, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"session_id": sessionID,
		"tool":       toolName,
		"arguments":  arguments,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/tools/call",
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("工具调用失败 (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return result, nil
}

// SubscribeEvents 订阅事件(HTTP 轮询模拟)
func (c *HTTPClient) SubscribeEvents(ctx context.Context, sessionID string, eventTypes []string) (<-chan Event, error) {
	eventChan := make(chan Event, 10)

	go func() {
		defer close(eventChan)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 轮询获取事件
				events, err := c.pollEvents(ctx, sessionID, eventTypes)
				if err != nil {
					continue
				}

				for _, event := range events {
					select {
					case eventChan <- event:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// pollEvents 轮询事件
func (c *HTTPClient) pollEvents(ctx context.Context, sessionID string, eventTypes []string) ([]Event, error) {
	// 构建查询参数
	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/events?session_id="+sessionID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取事件失败: status %d", resp.StatusCode)
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

// Close 关闭连接
func (c *HTTPClient) Close(ctx context.Context) error {
	// HTTP 无状态,无需关闭
	return nil
}
