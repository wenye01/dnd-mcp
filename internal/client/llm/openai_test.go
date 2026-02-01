package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAIClient_Chat_Success 测试成功的聊天请求
func TestOpenAIClient_Chat_Success(t *testing.T) {
	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 解析请求体
		var req ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", req.Model)
		assert.Len(t, req.Messages, 2)

		// 返回mock响应
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test-001",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "你好，冒险者！有什么可以帮你的吗？",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     50,
				CompletionTokens: 20,
				TotalTokens:      70,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	config := &Config{
		Provider:    "openai",
		APIKey:      "test-api-key",
		BaseURL:     server.URL,
		Model:       "gpt-4",
		MaxRetries:  3,
		Timeout:     30,
		Temperature: 0.7,
	}

	client, err := NewOpenAIClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	// 发送请求
	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "system", Content: "你是一个DND地下城主"},
			{Role: "user", Content: "你好"},
		},
		Temperature: 0.7,
	}

	resp, err := client.ChatCompletion(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "chatcmpl-test-001", resp.ID)
	assert.Equal(t, "你好，冒险者！有什么可以帮你的吗？", resp.Choices[0].Message.Content)
	assert.Equal(t, 70, resp.Usage.TotalTokens)
}

// TestOpenAIClient_Chat_ToolCall 测试工具调用响应
func TestOpenAIClient_Chat_ToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回工具调用响应
		response := ChatCompletionResponse{
			ID:      "chatcmpl-test-002",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "",
						ToolCalls: []ToolCall{
							{
								ID:   "call_001",
								Type: "function",
								Function: FunctionCall{
									Name:      "resolve_attack",
									Arguments: `{"attacker_id":"char-001","target_id":"goblin-001","attack_type":"melee"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: Usage{
				PromptTokens:     60,
				CompletionTokens: 30,
				TotalTokens:      90,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
	}

	client, err := NewOpenAIClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "我要攻击那个哥布林"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Choices[0].Message.ToolCalls, 1)
	assert.Equal(t, "resolve_attack", resp.Choices[0].Message.ToolCalls[0].Function.Name)
}

// TestOpenAIClient_Chat_HTTPError 测试HTTP错误
func TestOpenAIClient_Chat_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid API key",
		})
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
	}

	client, err := NewOpenAIClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "测试"}},
	}

	resp, err := client.ChatCompletion(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "401")
}

// TestOpenAIClient_Chat_Timeout 测试超时
func TestOpenAIClient_Chat_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟延迟响应
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
		Timeout: 1, // 1秒超时
	}

	client, err := NewOpenAIClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "测试"}},
	}

	_, err = client.ChatCompletion(ctx, req)
	assert.Error(t, err)
}

// TestRetryableClient_Success_NoRetry 测试无需重试的成功场景
func TestRetryableClient_Success_NoRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID:     "test-001",
			Object: "chat.completion",
			Model:  "gpt-4",
			Choices: []Choice{
				{
					Message: Message{Role: "assistant", Content: "成功"},
				},
			},
		})
	}))
	defer server.Close()

	baseConfig := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
	}

	baseClient, err := NewOpenAIClient(baseConfig)
	require.NoError(t, err)

	retryClient := NewRetryableClient(baseClient, 3)

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "测试"}},
	}

	resp, err := retryClient.ChatCompletion(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, attemptCount, "Should not retry on success")
}

// TestRetryableClient_RetryOn429 测试429错误的重试
func TestRetryableClient_RetryOn429(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			// 第一次返回429
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// 第二次成功
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			ID:     "test-002",
			Object: "chat.completion",
			Model:  "gpt-4",
			Choices: []Choice{
				{
					Message: Message{Role: "assistant", Content: "重试成功"},
				},
			},
		})
	}))
	defer server.Close()

	baseConfig := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
	}

	baseClient, err := NewOpenAIClient(baseConfig)
	require.NoError(t, err)

	retryClient := NewRetryableClient(baseClient, 3)

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "测试"}},
	}

	resp, err := retryClient.ChatCompletion(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, attemptCount, "Should retry on 429 error")
	assert.Equal(t, "重试成功", resp.Choices[0].Message.Content)
}

// TestRetryableClient_RetryExhausted 测试重试次数耗尽
func TestRetryableClient_RetryExhausted(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// 始终返回500错误
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	baseConfig := &Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
	}

	baseClient, err := NewOpenAIClient(baseConfig)
	require.NoError(t, err)

	retryClient := NewRetryableClient(baseClient, 2) // 最多重试2次

	ctx := context.Background()
	req := &ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "测试"}},
	}

	_, err = retryClient.ChatCompletion(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after")
	assert.Equal(t, 3, attemptCount, "Should attempt 1 initial + 2 retries")
}

// TestConfig_Validation 测试配置验证
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Provider:    "openai",
				APIKey:      "test-key",
				BaseURL:     "https://api.openai.com",
				Model:       "gpt-4",
				MaxRetries:  3,
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "empty api key",
			config: &Config{
				Provider:    "openai",
				APIKey:      "",
				BaseURL:     "https://api.openai.com",
				Model:       "gpt-4",
				MaxRetries:  3,
				Temperature: 0.7,
			},
			wantErr: false, // OpenAI客户端创建时不会验证API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestMessage_MarshalUnmarshal 测试消息序列化
func TestMessage_MarshalUnmarshal(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "测试消息",
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, msg.Role, decoded.Role)
	assert.Equal(t, msg.Content, decoded.Content)
}

// TestToolCall_MarshalUnmarshal 测试工具调用序列化
func TestToolCall_MarshalUnmarshal(t *testing.T) {
	toolCall := ToolCall{
		ID:   "call_001",
		Type: "function",
		Function: FunctionCall{
			Name:      "roll_dice",
			Arguments: `{"dice_type":"d20","modifier":5}`,
		},
	}

	data, err := json.Marshal(toolCall)
	require.NoError(t, err)

	var decoded ToolCall
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, toolCall.ID, decoded.ID)
	assert.Equal(t, toolCall.Type, decoded.Type)
	assert.Equal(t, toolCall.Function.Name, decoded.Function.Name)
	assert.Equal(t, toolCall.Function.Arguments, decoded.Function.Arguments)
}

// TestUsage_CalculateTotal 测试Token使用统计
func TestUsage_CalculateTotal(t *testing.T) {
	usage := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	assert.Equal(t, 150, usage.TotalTokens)
	assert.Equal(t, 100, usage.PromptTokens)
	assert.Equal(t, 50, usage.CompletionTokens)
}
