package llm

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// RetryableClient 带重试功能的LLM客户端
type RetryableClient struct {
	client    Client
	maxRetries int
}

// NewRetryableClient 创建带重试功能的客户端
func NewRetryableClient(client Client, maxRetries int) *RetryableClient {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	return &RetryableClient{
		client:    client,
		maxRetries: maxRetries,
	}
}

// ChatCompletion 带重试的聊天完成
func (r *RetryableClient) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避时间(指数退避)
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			log.Printf("LLM request failed, retrying in %v (attempt %d/%d)", backoff, attempt, r.maxRetries)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := r.client.ChatCompletion(ctx, req)
		if err == nil {
			// 成功,返回响应
			if attempt > 0 {
				log.Printf("LLM request succeeded on retry %d", attempt)
			}
			return resp, nil
		}

		lastErr = err

		// 检查是否应该重试
		if !isRetryableError(err) {
			// 不可重试的错误,直接返回
			log.Printf("LLM request failed with non-retryable error: %v", err)
			return nil, err
		}

		log.Printf("LLM request failed with retryable error (attempt %d/%d): %v", attempt+1, r.maxRetries+1, err)
	}

	// 所有重试都失败
	return nil, fmt.Errorf("LLM request failed after %d attempts: %w", r.maxRetries+1, lastErr)
}

// StreamCompletion 带重试的流式聊天完成(暂未实现)
func (r *RetryableClient) StreamCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error) {
	return nil, fmt.Errorf("stream completion not implemented")
}

// isRetryableError 判断错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// 可重试的HTTP状态码
	retryableCodes := []string{
		"429", // Too Many Requests
		"500", // Internal Server Error
		"502", // Bad Gateway
		"503", // Service Unavailable
		"504", // Gateway Timeout
	}

	for _, code := range retryableCodes {
		if contains(errMsg, code) {
			return true
		}
	}

	// 可重试的错误消息
	retryableMessages := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"temporary failure",
		"rate limit",
	}

	for _, msg := range retryableMessages {
		if contains(errMsg, msg) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子串(忽略大小写)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(s) > len(substr) && (
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
