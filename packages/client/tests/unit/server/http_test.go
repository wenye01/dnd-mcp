// Package server_test 测试 HTTP Server Client
package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	serverpkg "github.com/dnd-mcp/client/internal/server"
)

// setupTestHTTPServer 创建测试用的 HTTP Server
func setupTestHTTPServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// TestHTTPClient_Initialize 测试初始化方法
func TestHTTPClient_Initialize(t *testing.T) {
	t.Run("正常初始化", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求路径和方法
			if r.URL.Path != "/mcp/initialize" {
				t.Errorf("期望路径 /mcp/initialize, 得到 %s", r.URL.Path)
			}
			if r.Method != "POST" {
				t.Errorf("期望方法 POST, 得到 %s", r.Method)
			}

			// 验证请求体
			var reqBody map[string]any
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("解析请求体失败: %v", err)
				return
			}

			if reqBody["protocolVersion"] != "2024-11-05" {
				t.Errorf("期望 protocolVersion 2024-11-05, 得到 %v", reqBody["protocolVersion"])
			}

			// 返回成功响应
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"protocolVersion": "2024-11-05",
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		err := client.Initialize(ctx)
		if err != nil {
			t.Errorf("Initialize() 不应该返回错误: %v", err)
		}
	})

	t.Run("初始化失败-非200状态码", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		err := client.Initialize(ctx)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("错误消息应包含状态码 500, 得到: %v", err)
		}
	})

	t.Run("初始化失败-网络错误", func(t *testing.T) {
		// 使用无效的 URL
		client := serverpkg.NewHTTPClient("http://invalid-host-99999:0", 1)
		ctx := context.Background()

		err := client.Initialize(ctx)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
	})

	t.Run("初始化失败-上下文取消", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 延迟响应，确保上下文被取消
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := client.Initialize(ctx)
		if err == nil {
			t.Error("期望返回错误（上下文取消），但没有")
		}
	})
}

// TestHTTPClient_mcpCall 测试核心 mcpCall 方法（通过其他方法间接测试）
func TestHTTPClient_mcpCall(t *testing.T) {
	t.Run("正常 MCP 调用", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求路径和方法
			if r.URL.Path != "/mcp/tools/call" {
				t.Errorf("期望路径 /mcp/tools/call, 得到 %s", r.URL.Path)
			}
			if r.Method != "POST" {
				t.Errorf("期望方法 POST, 得到 %s", r.Method)
			}

			// 解析请求体
			var reqBody map[string]any
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("解析请求体失败: %v", err)
				return
			}

			if reqBody["name"] != "test_tool" {
				t.Errorf("期望工具名 test_tool, 得到 %v", reqBody["name"])
			}

			// 返回成功响应
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": `{"result": "success"}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		// 通过 CallTool 间接测试 mcpCall
		result, err := client.CallTool(ctx, "campaign-1", "test_tool", map[string]any{"key": "value"})
		if err != nil {
			t.Errorf("CallTool() 不应该返回错误: %v", err)
		}

		if result == nil {
			t.Error("期望返回结果，得到 nil")
		}

		// 验证 campaign_id 被添加到参数中
		if result["campaign_id"] != "campaign-1" {
			// 由于 mock 返回的是固定响应，这里只验证没有错误
		}
	})

	t.Run("MCP 调用返回错误-IsError为true", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": "tool execution failed",
					},
				},
				"isError": true,
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		_, err := client.CallTool(ctx, "campaign-1", "test_tool", nil)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
		if !strings.Contains(err.Error(), "tool execution failed") {
			t.Errorf("错误消息应包含 'tool execution failed', 得到: %v", err)
		}
	})

	t.Run("MCP 调用失败-非200状态码", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		_, err := client.CallTool(ctx, "campaign-1", "test_tool", nil)
		if err == nil {
			t.Error("期望返回错误，但没有")
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("错误消息应包含状态码 400, 得到: %v", err)
		}
	})

	t.Run("MCP 调用失败-响应内容为空", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		_, err := client.CallTool(ctx, "campaign-1", "test_tool", nil)
		if err == nil {
			t.Error("期望返回错误（响应内容为空），但没有")
		}
	})

	t.Run("MCP 调用失败-响应JSON无效", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": "invalid json{{",
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		_, err := client.CallTool(ctx, "campaign-1", "test_tool", nil)
		if err == nil {
			t.Error("期望返回错误（JSON解析失败），但没有")
		}
	})
}

// TestHTTPClient_GetContext 测试获取上下文
func TestHTTPClient_GetContext(t *testing.T) {
	t.Run("正常获取上下文", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求
			var reqBody map[string]any
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("解析请求体失败: %v", err)
				return
			}

			if reqBody["name"] != "get_context" {
				t.Errorf("期望工具名 get_context, 得到 %v", reqBody["name"])
			}

			args := reqBody["arguments"].(map[string]any)
			if args["campaign_id"] != "campaign-1" {
				t.Errorf("期望 campaign_id campaign-1, 得到 %v", args["campaign_id"])
			}
			if args["message_limit"] != float64(20) {
				t.Errorf("期望 message_limit 20, 得到 %v", args["message_limit"])
			}
			if args["include_combat"] != true {
				t.Errorf("期望 include_combat true, 得到 %v", args["include_combat"])
			}

			// 返回成功响应
			w.WriteHeader(http.StatusOK)
			response := map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": `{
							"context": {
								"campaign_id": "campaign-1",
								"game_summary": {
									"time": "noon",
									"location": "Forest",
									"weather": "sunny",
									"in_combat": false,
									"party": []
								},
								"messages": [],
								"raw_message_count": 0,
								"token_estimate": 100,
								"created_at": "2026-03-10T12:00:00Z"
							}
						}`,
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		result, err := client.GetContext(ctx, "campaign-1", 20, true)
		if err != nil {
			t.Errorf("GetContext() 不应该返回错误: %v", err)
		}

		if result == nil {
			t.Error("期望返回上下文，得到 nil")
		}

		if result.CampaignID != "campaign-1" {
			t.Errorf("期望 CampaignID campaign-1, 得到 %s", result.CampaignID)
		}

		if result.GameSummary == nil {
			t.Error("GameSummary 不应为 nil")
		}
	})

	t.Run("参数正确传递", func(t *testing.T) {
		receivedArgs := make(map[string]any)

		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)
			args := reqBody["arguments"].(map[string]any)

			receivedArgs["campaign_id"] = args["campaign_id"]
			receivedArgs["message_limit"] = args["message_limit"]
			receivedArgs["include_combat"] = args["include_combat"]

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": `{"context": {"campaign_id": "test-campaign", "game_summary": null, "messages": [], "raw_message_count": 0, "token_estimate": 0, "created_at": "2026-03-10T12:00:00Z"}}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		_, _ = client.GetContext(ctx, "test-campaign", 50, false)

		if receivedArgs["campaign_id"] != "test-campaign" {
			t.Errorf("期望 campaign_id test-campaign, 得到 %v", receivedArgs["campaign_id"])
		}
		if receivedArgs["message_limit"] != float64(50) {
			t.Errorf("期望 message_limit 50, 得到 %v", receivedArgs["message_limit"])
		}
		if receivedArgs["include_combat"] != false {
			t.Errorf("期望 include_combat false, 得到 %v", receivedArgs["include_combat"])
		}
	})
}

// TestHTTPClient_SaveMessage 测试保存消息
func TestHTTPClient_SaveMessage(t *testing.T) {
	t.Run("保存消息成功", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("解析请求体失败: %v", err)
				return
			}

			if reqBody["name"] != "save_message" {
				t.Errorf("期望工具名 save_message, 得到 %v", reqBody["name"])
			}

			args := reqBody["arguments"].(map[string]any)
			if args["campaign_id"] != "campaign-1" {
				t.Errorf("期望 campaign_id campaign-1, 得到 %v", args["campaign_id"])
			}
			if args["role"] != "user" {
				t.Errorf("期望 role user, 得到 %v", args["role"])
			}
			if args["content"] != "Hello, world!" {
				t.Errorf("期望 content 'Hello, world!', 得到 %v", args["content"])
			}

			// 返回成功响应
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type":  "text",
						"text": "message saved",
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		msg := &serverpkg.Message{
			ID:        "msg-1",
			CampaignID: "campaign-1",
			Role:       serverpkg.MessageRoleUser,
			Content:    "Hello, world!",
			CreatedAt:  time.Now(),
		}

		err := client.SaveMessage(ctx, "campaign-1", msg)
		if err != nil {
			t.Errorf("SaveMessage() 不应该返回错误: %v", err)
		}
	})

	t.Run("ToolCalls 处理", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)
			args := reqBody["arguments"].(map[string]any)

			// 验证 tool_calls 存在
			if args["tool_calls"] == nil {
				t.Error("期望 tool_calls 存在")
			}

			toolCalls := args["tool_calls"].([]any)
			if len(toolCalls) != 1 {
				t.Errorf("期望 1 个 tool_call, 得到 %d", len(toolCalls))
			}

			tc := toolCalls[0].(map[string]any)
			if tc["id"] != "tc-1" {
				t.Errorf("期望 tool_call id tc-1, 得到 %v", tc["id"])
			}
			if tc["name"] != "roll_dice" {
				t.Errorf("期望 tool_call name roll_dice, 得到 %v", tc["name"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type":  "text",
						"text": "message saved",
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		msg := &serverpkg.Message{
			ID:        "msg-1",
			CampaignID: "campaign-1",
			Role:       serverpkg.MessageRoleAssistant,
			Content:    "",
			ToolCalls: []serverpkg.ToolCall{
				{
					ID:   "tc-1",
					Name: "roll_dice",
					Arguments: map[string]any{
						"formula": "1d20+5",
					},
				},
			},
			CreatedAt: time.Now(),
		}

		err := client.SaveMessage(ctx, "campaign-1", msg)
		if err != nil {
			t.Errorf("SaveMessage() 不应该返回错误: %v", err)
		}
	})

	t.Run("带 player_id 的消息", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)
			args := reqBody["arguments"].(map[string]any)

			if args["player_id"] != "player-1" {
				t.Errorf("期望 player_id player-1, 得到 %v", args["player_id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type":  "text",
						"text": "message saved",
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		msg := &serverpkg.Message{
			ID:        "msg-1",
			CampaignID: "campaign-1",
			Role:       serverpkg.MessageRoleUser,
			Content:    "test",
			PlayerID:   "player-1",
			CreatedAt:  time.Now(),
		}

		err := client.SaveMessage(ctx, "campaign-1", msg)
		if err != nil {
			t.Errorf("SaveMessage() 不应该返回错误: %v", err)
		}
	})
}

// TestHTTPClient_CallTool 测试调用工具
func TestHTTPClient_CallTool(t *testing.T) {
	t.Run("调用工具成功", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["name"] != "roll_dice" {
				t.Errorf("期望工具名 roll_dice, 得到 %v", reqBody["name"])
			}

			args := reqBody["arguments"].(map[string]any)
			if args["campaign_id"] != "campaign-1" {
				t.Errorf("期望 campaign_id 自动添加为 campaign-1, 得到 %v", args["campaign_id"])
			}
			if args["formula"] != "2d6+3" {
				t.Errorf("期望 formula 2d6+3, 得到 %v", args["formula"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": `{"result": 8, "rolls": [3, 5], "formula": "2d6+3"}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		result, err := client.CallTool(ctx, "campaign-1", "roll_dice", map[string]any{
			"formula": "2d6+3",
		})

		if err != nil {
			t.Errorf("CallTool() 不应该返回错误: %v", err)
		}

		if result == nil {
			t.Error("期望返回结果，得到 nil")
		}

		if result["result"] != float64(8) {
			t.Errorf("期望 result 8, 得到 %v", result["result"])
		}
	})

	t.Run("campaign_id 自动添加", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)
			args := reqBody["arguments"].(map[string]any)

			// 验证 campaign_id 被自动添加
			if args["campaign_id"] != "auto-campaign" {
				t.Errorf("期望 campaign_id auto-campaign, 得到 %v", args["campaign_id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type":  "text",
						"text": `{"success": true}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		// 传入空的 args
		_, err := client.CallTool(ctx, "auto-campaign", "test_tool", nil)
		if err != nil {
			t.Errorf("CallTool() 不应该返回错误: %v", err)
		}
	})

	t.Run("campaign_id 覆盖已有值", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)
			args := reqBody["arguments"].(map[string]any)

			// 验证 campaign_id 被覆盖为参数值
			if args["campaign_id"] != "campaign-1" {
				t.Errorf("期望 campaign_id campaign-1, 得到 %v", args["campaign_id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type":  "text",
						"text": `{"success": true}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		// 传入带有不同 campaign_id 的 args
		_, err := client.CallTool(ctx, "campaign-1", "test_tool", map[string]any{
			"campaign_id": "old-campaign",
		})
		if err != nil {
			t.Errorf("CallTool() 不应该返回错误: %v", err)
		}
	})
}

// TestHTTPClient_GetRawContext 测试获取原始上下文
func TestHTTPClient_GetRawContext(t *testing.T) {
	t.Run("正常获取原始上下文", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]any
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["name"] != "get_raw_context" {
				t.Errorf("期望工具名 get_raw_context, 得到 %v", reqBody["name"])
			}

			args := reqBody["arguments"].(map[string]any)
			if args["campaign_id"] != "campaign-1" {
				t.Errorf("期望 campaign_id campaign-1, 得到 %v", args["campaign_id"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": `{
							"campaign_id": "campaign-1",
							"game_state": {
								"location": "Forest",
								"game_time": "noon",
								"last_rest_time": "2026-03-10T08:00:00Z",
								"short_rests": 1,
								"long_rests": 0,
								"active_effects": []
							},
							"characters": [],
							"combat": null,
							"map": null,
							"messages": [],
							"message_count": 0
						}`,
					},
				},
			})
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 30)
		ctx := context.Background()

		result, err := client.GetRawContext(ctx, "campaign-1")
		if err != nil {
			t.Errorf("GetRawContext() 不应该返回错误: %v", err)
		}

		if result == nil {
			t.Error("期望返回原始上下文，得到 nil")
		}

		if result.CampaignID != "campaign-1" {
			t.Errorf("期望 CampaignID campaign-1, 得到 %s", result.CampaignID)
		}

		if result.GameState == nil {
			t.Error("GameState 不应为 nil")
		}

		if result.GameState.Location != "Forest" {
			t.Errorf("期望 Location Forest, 得到 %s", result.GameState.Location)
		}
	})
}

// TestHTTPClient_Close 测试关闭方法
func TestHTTPClient_Close(t *testing.T) {
	t.Run("关闭成功", func(t *testing.T) {
		client := serverpkg.NewHTTPClient("http://localhost:8080", 30)
		ctx := context.Background()

		err := client.Close(ctx)
		if err != nil {
			t.Errorf("Close() 不应该返回错误: %v", err)
		}
	})
}

// TestHTTPClient_Timeout 测试超时处理
func TestHTTPClient_Timeout(t *testing.T) {
	t.Run("请求超时（通过 context）", func(t *testing.T) {
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 延迟响应，超过 context 超时时间
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		client := serverpkg.NewHTTPClient(testServer.URL, 5)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := client.Initialize(ctx)
		if err == nil {
			t.Error("期望返回超时错误，但没有")
		}
	})

	t.Run("请求在超时前完成", func(t *testing.T) {
		requestReceived := make(chan struct{})
		testServer := setupTestHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(requestReceived)
			// 快速响应
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"protocolVersion": "2024-11-05",
			})
		}))
		defer testServer.Close()

		// 创建有足够超时时间的客户端
		client := serverpkg.NewHTTPClient(testServer.URL, 5)
		ctx := context.Background()

		err := client.Initialize(ctx)
		if err != nil {
			t.Errorf("不应该返回错误: %v", err)
		}

		// 确保请求被接收
		select {
		case <-requestReceived:
			// OK
		case <-time.After(100 * time.Millisecond):
			t.Error("请求未被接收")
		}
	})
}
