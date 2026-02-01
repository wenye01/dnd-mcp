package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient OpenAI客户端实现
type OpenAIClient struct {
	config     *Config
	httpClient *http.Client
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(config *Config) (*OpenAIClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	client := &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}

	return client, nil
}

// ChatCompletion 发送聊天完成请求
func (c *OpenAIClient) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// 构建请求URL
	url := fmt.Sprintf("%s/v1/chat/completions", c.config.BaseURL)

	// 序列化请求体
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var completionResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &completionResp, nil
}

// StreamCompletion 发送流式聊天完成请求(暂未实现)
func (c *OpenAIClient) StreamCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error) {
	// TODO: 实现流式响应
	return nil, fmt.Errorf("stream completion not implemented yet")
}
