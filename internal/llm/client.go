// Package llm 提供 LLM 客户端接口和实现
package llm

import (
	"context"
)

// LLMClient LLM 客户端接口
type LLMClient interface {
	// Chat 聊天对话，返回 AI 响应
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages []Message `json:"messages"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Message Message `json:"message"`
}
