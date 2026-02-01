package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ChatCompletionRequest OpenAI Chat Completion请求格式
type ChatCompletionRequest struct {
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  any       `json:"tool_choice,omitempty"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message 消息格式
type Message struct {
	Role       string      `json:"role"` // system, user, assistant, tool
	Content    string      `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
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

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage Token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse OpenAI Chat Completion响应格式
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

var requestCount int

func main() {
	http.HandleFunc("/v1/chat/completions", handleChatCompletion)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	port := ":9001"
	fmt.Printf("🤖 Mock LLM Server starting on http://localhost%s\n", port)
	fmt.Println("   模拟 OpenAI API 格式")
	fmt.Println("   支持的端点:")
	fmt.Println("   - POST /v1/chat/completions")
	fmt.Println("   - GET  /health")
	fmt.Println()

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleChatCompletion 处理聊天完成请求
func handleChatCompletion(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "invalid_request_error")
		return
	}

	// 解析请求
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_request_error")
		return
	}

	requestCount++
	log.Printf("[Request #%d] Model: %s, Messages: %d, Tools: %d",
		requestCount, req.Model, len(req.Messages), len(req.Tools))

	// 生成响应
	response := generateResponse(req)

	// 记录响应
	if len(response.Choices) > 0 {
		msg := response.Choices[0].Message
		if len(msg.ToolCalls) > 0 {
			log.Printf("[Response #%d] Tool Calls: %d", requestCount, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				log.Printf("   - %s: %s", tc.Function.Name, tc.Function.Arguments)
			}
		} else {
			log.Printf("[Response #%d] Text: %s", requestCount, truncate(msg.Content, 100))
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateResponse 根据请求生成响应
func generateResponse(req ChatCompletionRequest) ChatCompletionResponse {
	// 获取最后一条用户消息
	var lastUserMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMessage = req.Messages[i].Content
			break
		}
	}

	if lastUserMessage == "" {
		// 没有用户消息,返回简单对话
		return simpleResponse("你好!我是你的地下城主DM。有什么可以帮你的吗?", 50, 20)
	}

	content := strings.ToLower(lastUserMessage)

	// 根据消息内容决定响应类型
	switch {
	case containsKeywords(content, []string{"攻击", "战斗", "打"}):
		// 返回攻击工具调用
		return toolCallResponse("resolve_attack", map[string]any{
			"attacker_id":  "char-001",
			"target_id":    "goblin-001",
			"attack_type":  "melee",
			"weapon_damage": "1d8+3",
		}, 60, 30)

	case containsKeywords(content, []string{"投骰", "roll", "骰子"}):
		// 返回投骰工具调用
		return toolCallResponse("roll_dice", map[string]any{
			"dice_type": "d20",
			"reason":    "attack_roll",
			"modifier":  5,
		}, 40, 20)

	case containsKeywords(content, []string{"移动", "走", "前往"}):
		// 返回移动工具调用
		return toolCallResponse("move_character", map[string]any{
			"character_id": "char-001",
			"new_location": "underground_entrance",
		}, 50, 25)

	case containsKeywords(content, []string{"创建角色", "新建角色", "生成角色"}):
		// 返回创建角色工具调用
		return toolCallResponse("create_character", map[string]any{
			"name":      "新角色",
			"race":      "人类",
			"class":     "战士",
			"level":     1,
			"hp_max":    20,
			"hp":        20,
			"strength":  16,
			"dexterity": 14,
		}, 70, 35)

	case containsKeywords(content, []string{"查看状态", "状态", "当前情况"}):
		// 返回查询状态工具调用
		return toolCallResponse("get_state", map[string]any{
			"session_id": "test-session-001",
		}, 30, 15)

	default:
		// 简单对话
		responses := []string{
			"很好,冒险者!请继续你的行动。",
			"我明白了,你想要做什么?",
			"作为一名地下城主,我会协助你进行这个冒险。",
			"有趣的选择!接下来会发生什么?",
			"我在听,请告诉我你的下一步行动。",
		}
		responseText := responses[(requestCount-1)%len(responses)]

		// 根据上下文定制响应
		if strings.Contains(content, "你好") || strings.Contains(content, "hi") {
			responseText = "你好,勇敢的冒险者!欢迎来到被遗忘的国度。我是你的地下城主DM。"
		} else if strings.Contains(content, "地下城") {
			responseText = "这座地下城充满了神秘和危险。你准备好了吗?"
		} else if strings.Contains(content, "怪物") || strings.Contains(content, "敌人") {
			responseText = "当心!前方可能有危险的生物。"
		}

		return simpleResponse(responseText, 50, 30+len(responseText)/2)
	}
}

// simpleResponse 生成简单的文本响应
func simpleResponse(content string, promptTokens, completionTokens int) ChatCompletionResponse {
	return ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-mock-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4-mock",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// toolCallResponse 生成工具调用响应
func toolCallResponse(toolName string, args map[string]any, promptTokens, completionTokens int) ChatCompletionResponse {
	argsJSON, _ := json.Marshal(args)

	return ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-mock-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4-mock",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []ToolCall{
						{
							ID:   fmt.Sprintf("call_%d", time.Now().UnixNano()),
							Type: "function",
							Function: FunctionCall{
								Name:      toolName,
								Arguments: string(argsJSON),
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
}

// containsKeywords 检查内容是否包含关键词
func containsKeywords(content string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, status int, message, errorType string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    any    `json:"code"`
		}{
			Message: message,
			Type:    errorType,
			Code:    nil,
		},
	})
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
