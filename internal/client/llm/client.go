package llm

import (
	"context"
)

// Client LLM客户端接口
type Client interface {
	// ChatCompletion 发送聊天完成请求
	ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// StreamCompletion 发送流式聊天完成请求(可选)
	StreamCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error)
}

// ChatCompletionRequest 聊天完成请求
type ChatCompletionRequest struct {
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  any       `json:"tool_choice,omitempty"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message 消息
type Message struct {
	Role       string     `json:"role"` // system, user, assistant, tool
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse 聊天完成响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择项
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"` // stop, length, tool_calls, content_filter
}

// Usage Token使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice 流式选择项
type ChunkChoice struct {
	Index        int          `json:"index"`
	Delta        MessageDelta `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

// MessageDelta 消息增量
type MessageDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Config LLM客户端配置
type Config struct {
	Provider     string  // openai, mock, etc.
	APIKey       string  // API密钥
	BaseURL      string  // API基础URL
	Model        string  // 模型名称
	MaxRetries   int     // 最大重试次数
	Timeout      int     // 请求超时(秒)
	Temperature  float64 // 温度参数
	MaxTokens    int     // 最大token数
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Provider:    "mock",
		BaseURL:     "http://localhost:9001",
		Model:       "gpt-4",
		MaxRetries:  3,
		Timeout:     30,
		Temperature: 0.7,
		MaxTokens:   4096,
	}
}
