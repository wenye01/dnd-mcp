// Package llm 提供 Mock LLM 实现
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

	// 获取最后一条用户消息
	lastUserMessage := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMessage = req.Messages[i].Content
			break
		}
	}

	// 如果用户消息包含"投掷"或"dice",返回tool_calls
	if strings.Contains(lastUserMessage, "投掷") || strings.Contains(lastUserMessage, "dice") || strings.Contains(lastUserMessage, "d20") {
		// 构造 tool_calls
		arguments := map[string]interface{}{
			"formula": "1d20+5",
		}
		argumentsJSON, _ := json.Marshal(arguments)

		return &ChatResponse{
			ID:      "mock-response-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "mock-model",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:      "assistant",
						Content:   "",
						ToolCalls: []ToolCall{
							{
								ID:   "call_mock_001",
								Type: "function",
								Function: FunctionCall{
									Name:      "roll_dice",
									Arguments: string(argumentsJSON),
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}, nil
	}

	// 检查是否是工具调用后的后续请求
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "tool" {
			// 这是工具调用后的请求,返回最终响应
			return &ChatResponse{
				ID:      "mock-response-id",
				Object:  "chat.completion",
				Created: time.Now().Unix(),
				Model:   "mock-model",
				Choices: []Choice{
					{
						Index: 0,
						Message: Message{
							Role:    "assistant",
							Content: fmt.Sprintf("投掷完成!结果是 18(15+3)。%s", m.Response),
						},
						FinishReason: "stop",
					},
				},
				Usage: Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
			}, nil
		}
	}

	// 返回预设响应（兼容新格式）
	return &ChatResponse{
		ID:      "mock-response-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-model",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: m.Response,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}
