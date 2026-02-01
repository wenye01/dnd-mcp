package fixtures

import "github.com/dnd-mcp/client/internal/client/llm"

// 简单对话响应
var SimpleChatResponse = &llm.ChatCompletionResponse{
	ID:     "chatcmpl-test-simple",
	Object: "chat.completion",
	Model:  "gpt-4-mock",
	Choices: []llm.Choice{
		{
			Index: 0,
			Message: llm.Message{
				Role:    "assistant",
				Content: "你好,冒险者!有什么可以帮你的吗?",
			},
			FinishReason: "stop",
		},
	},
	Usage: llm.Usage{
		PromptTokens:     50,
		CompletionTokens: 20,
		TotalTokens:      70,
	},
}

// 攻击工具调用响应
var AttackToolCallResponse = &llm.ChatCompletionResponse{
	ID:     "chatcmpl-test-attack",
	Object: "chat.completion",
	Model:  "gpt-4-mock",
	Choices: []llm.Choice{
		{
			Index: 0,
			Message: llm.Message{
				Role:    "assistant",
				Content: "",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_test_attack",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "resolve_attack",
							Arguments: `{"attacker_id":"char-001","target_id":"goblin-001","attack_type":"melee"}`,
						},
					},
				},
			},
			FinishReason: "tool_calls",
		},
	},
	Usage: llm.Usage{
		PromptTokens:     60,
		CompletionTokens: 30,
		TotalTokens:      90,
	},
}

// 投骰工具调用响应
var DiceToolCallResponse = &llm.ChatCompletionResponse{
	ID:     "chatcmpl-test-dice",
	Object: "chat.completion",
	Model:  "gpt-4-mock",
	Choices: []llm.Choice{
		{
			Index: 0,
			Message: llm.Message{
				Role:    "assistant",
				Content: "",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_test_dice",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "roll_dice",
							Arguments: `{"dice_type":"d20","reason":"attack_roll","modifier":5}`,
						},
					},
				},
			},
			FinishReason: "tool_calls",
		},
	},
	Usage: llm.Usage{
		PromptTokens:     40,
		CompletionTokens: 20,
		TotalTokens:      60,
	},
}

// 多工具调用响应
var MultiToolCallResponse = &llm.ChatCompletionResponse{
	ID:     "chatcmpl-test-multi",
	Object: "chat.completion",
	Model:  "gpt-4-mock",
	Choices: []llm.Choice{
		{
			Index: 0,
			Message: llm.Message{
				Role:    "assistant",
				Content: "",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_test_multi_1",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "roll_dice",
							Arguments: `{"dice_type":"d20","reason":"initiative"}`,
						},
					},
					{
						ID:   "call_test_multi_2",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "resolve_attack",
							Arguments: `{"attacker_id":"char-001","target_id":"goblin-001"}`,
						},
					},
				},
			},
			FinishReason: "tool_calls",
		},
	},
	Usage: llm.Usage{
		PromptTokens:     80,
		CompletionTokens: 50,
		TotalTokens:      130,
	},
}

// 示例消息列表
var SampleMessages = []llm.Message{
	{
		Role:    "system",
		Content: "你是一个DND地下城主DM。",
	},
	{
		Role:    "user",
		Content: "你好",
	},
}

// 示例工具定义
var SampleTools = []llm.Tool{
	{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "resolve_attack",
			Description: "结算攻击",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"attacker_id": map[string]any{
						"type":        "string",
						"description": "攻击者ID",
					},
					"target_id": map[string]any{
						"type":        "string",
						"description": "目标ID",
					},
				},
				"required": []string{"attacker_id", "target_id"},
			},
		},
	},
	{
		Type: "function",
		Function: llm.ToolFunction{
			Name:        "roll_dice",
			Description: "投骰子",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"dice_type": map[string]any{
						"type":        "string",
						"description": "骰子类型 (d4, d6, d8, d10, d12, d20)",
					},
					"reason": map[string]any{
						"type":        "string",
						"description": "投骰原因",
					},
				},
				"required": []string{"dice_type"},
			},
		},
	},
}
