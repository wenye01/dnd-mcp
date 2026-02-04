// Package llm 提供 Mock LLM 实现
package llm

import (
	"context"
	"fmt"
)

// MockLLMClient Mock LLM 客户端
type MockLLMClient struct {
	Response string // 预设响应
}

// NewMockLLMClient 创建 Mock LLM 客户端
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		Response: "这是 Mock LLM 的响应",
	}
}

// SetResponse 设置预设响应
func (m *MockLLMClient) SetResponse(response string) {
	m.Response = response
}

// Chat 实现聊天接口
func (m *MockLLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// 返回预设响应
	return &ChatResponse{
		Message: Message{
			Role:    "assistant",
			Content: m.Response,
		},
	}, nil
}
