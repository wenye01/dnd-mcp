// Package server 提供 HTTP Server Client 实现
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// mcpContent MCP 响应内容
type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// mcpCallToolResponse MCP 工具调用响应
type mcpCallToolResponse struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// HTTPClient HTTP Server 客户端
type HTTPClient struct {
	baseURL     string
	httpClient  *http.Client
	timeout     time.Duration
	mu          sync.RWMutex
	initialized bool
}

// NewHTTPClient 创建 HTTP Server 客户端
func NewHTTPClient(baseURL string, timeoutSeconds int) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		timeout:     time.Duration(timeoutSeconds) * time.Second,
		initialized: false,
	}
}

// mcpCall 是核心方法，所有操作都通过它调用 MCP Tool
func (c *HTTPClient) mcpCall(ctx context.Context, toolName string, args map[string]any) (*mcpCallToolResponse, error) {
	url := fmt.Sprintf("%s/mcp/tools/call", c.baseURL)

	reqBody := map[string]any{
		"name":      toolName,
		"arguments": args,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP 调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCP 调用失败 (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result mcpCallToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// parseMCPResponse 解析 MCP 响应中的 JSON 内容
func parseMCPResponse[T any](resp *mcpCallToolResponse) (*T, error) {
	if resp.IsError {
		if len(resp.Content) > 0 {
			return nil, fmt.Errorf("工具调用错误: %s", resp.Content[0].Text)
		}
		return nil, fmt.Errorf("工具调用错误")
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("响应内容为空")
	}

	var result T
	if err := json.Unmarshal([]byte(resp.Content[0].Text), &result); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	return &result, nil
}

// Initialize 执行 MCP 握手
func (c *HTTPClient) Initialize(ctx context.Context) error {
	url := fmt.Sprintf("%s/mcp/initialize", c.baseURL)

	reqBody := map[string]any{
		"protocolVersion": "2024-11-05",
		"clientInfo": map[string]any{
			"name":    "dnd-mcp-client",
			"version": "1.0.0",
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建初始化请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("MCP 初始化失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MCP 初始化失败 (status %d): %s", resp.StatusCode, string(respBody))
	}

	// 设置 initialized 标志
	c.mu.Lock()
	c.initialized = true
	c.mu.Unlock()

	return nil
}

// GetContext 获取压缩后的上下文
func (c *HTTPClient) GetContext(ctx context.Context, campaignID string, limit int, includeCombat bool) (*Context, error) {
	args := map[string]any{
		"campaign_id":    campaignID,
		"message_limit":  limit,
		"include_combat": includeCombat,
	}

	resp, err := c.mcpCall(ctx, "get_context", args)
	if err != nil {
		return nil, err
	}

	// MCP 响应格式: {content: [{type: "text", text: "{\"context\": {...}}"}]}
	parsed, err := parseMCPResponse[struct {
		Context *Context `json:"context"`
	}](resp)
	if err != nil {
		return nil, err
	}

	return parsed.Context, nil
}

// GetRawContext 获取原始上下文（完整模式）
func (c *HTTPClient) GetRawContext(ctx context.Context, campaignID string) (*RawContext, error) {
	args := map[string]any{
		"campaign_id": campaignID,
	}

	resp, err := c.mcpCall(ctx, "get_raw_context", args)
	if err != nil {
		return nil, err
	}

	return parseMCPResponse[RawContext](resp)
}

// SaveMessage 保存消息到 Server
func (c *HTTPClient) SaveMessage(ctx context.Context, campaignID string, msg *Message) error {
	args := map[string]any{
		"campaign_id": campaignID,
		"role":        string(msg.Role),
		"content":     msg.Content,
	}

	if msg.PlayerID != "" {
		args["player_id"] = msg.PlayerID
	}

	if len(msg.ToolCalls) > 0 {
		toolCalls := make([]map[string]any, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			toolCalls[i] = map[string]any{
				"id":        tc.ID,
				"name":      tc.Name,
				"arguments": tc.Arguments,
			}
		}
		args["tool_calls"] = toolCalls
	}

	resp, err := c.mcpCall(ctx, "save_message", args)
	if err != nil {
		return err
	}

	// 检查错误但忽略响应内容
	if resp.IsError {
		if len(resp.Content) > 0 {
			return fmt.Errorf("保存消息失败: %s", resp.Content[0].Text)
		}
		return fmt.Errorf("保存消息失败")
	}

	return nil
}

// CallTool 调用 Server MCP Tool
func (c *HTTPClient) CallTool(ctx context.Context, campaignID, toolName string, args map[string]any) (map[string]any, error) {
	// 确保 campaign_id 在参数中
	if args == nil {
		args = make(map[string]any)
	}
	args["campaign_id"] = campaignID

	resp, err := c.mcpCall(ctx, toolName, args)
	if err != nil {
		return nil, err
	}

	result, err := parseMCPResponse[map[string]any](resp)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// Close 关闭连接
func (c *HTTPClient) Close(ctx context.Context) error {
	// HTTP 无状态,无需关闭
	return nil
}
