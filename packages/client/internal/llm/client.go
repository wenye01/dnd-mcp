// Package llm 提供 LLM 客户端接口和实现
package llm

import (
	"context"
	"fmt"

	"github.com/dnd-mcp/client/pkg/config"
)

// LLMClient LLM 客户端接口
type LLMClient interface {
	// Chat 聊天对话，返回 AI 响应
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`                 // 模型名称
	Messages    []Message `json:"messages"`              // 消息列表
	Tools       []Tool    `json:"tools,omitempty"`       // 工具定义(可选)
	ToolChoice  any       `json:"tool_choice,omitempty"` // 工具选择策略(可选)
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Type     string      `json:"type"` // "function"
	Function FunctionDef `json:"function"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"` // stop, length, tool_calls, content_filter
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Message 消息
type Message struct {
	Role       string     `json:"role"` // system, user, assistant, tool
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // 工具调用
	ToolCallID string     `json:"tool_call_id,omitempty"` // tool 角色的响应
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// NewClient 创建 LLM 客户端
func NewClient(cfg *config.LLMConfig) (LLMClient, error) {
	switch cfg.Provider {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key 不能为空")
		}
		return NewOpenAIClient(cfg), nil

	case "mock":
		return NewMockLLMClient(), nil

	default:
		return nil, fmt.Errorf("不支持的 LLM provider: %s", cfg.Provider)
	}
}
