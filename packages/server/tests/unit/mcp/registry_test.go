package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dnd-mcp/server/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := mcp.NewRegistry()
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
}

func TestRegistry_Register(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: mcp.InputSchema{Type: "object"},
	}

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse("ok")
	}

	// Test successful registration
	err := r.Register(tool, handler)
	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())

	// Test duplicate registration
	err = r.Register(tool, handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegistry_Register_EmptyName(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "",
		Description: "A test tool",
	}

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse("ok")
	}

	err := r.Register(tool, handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestRegistry_Register_NilHandler(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}

	err := r.Register(tool, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestRegistry_MustRegister(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse("ok")
	}

	// Should not panic
	assert.NotPanics(t, func() {
		r.MustRegister(tool, handler)
	})

	// Should panic on duplicate
	assert.Panics(t, func() {
		r.MustRegister(tool, handler)
	})
}

func TestRegistry_Get(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse("ok")
	}

	// Get non-existent tool
	info, ok := r.Get("test_tool")
	assert.False(t, ok)
	assert.Nil(t, info)

	// Register tool
	r.Register(tool, handler)

	// Get existing tool
	info, ok = r.Get("test_tool")
	assert.True(t, ok)
	assert.NotNil(t, info)
	assert.Equal(t, "test_tool", info.Tool.Name)
}

func TestRegistry_List(t *testing.T) {
	r := mcp.NewRegistry()

	// Empty registry
	tools := r.List()
	assert.Empty(t, tools)

	// Add tools
	tool1 := mcp.Tool{Name: "tool1", Description: "Tool 1"}
	tool2 := mcp.Tool{Name: "tool2", Description: "Tool 2"}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse { return mcp.ToolResponse{} }

	r.Register(tool1, handler)
	r.Register(tool2, handler)

	tools = r.List()
	assert.Len(t, tools, 2)
}

func TestRegistry_Call(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{
		Name:        "echo",
		Description: "Echo tool",
	}

	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse {
		return mcp.NewTextResponse(string(req.Arguments))
	}

	r.Register(tool, handler)

	// Call existing tool
	req := mcp.ToolRequest{
		ToolName:  "echo",
		Arguments: json.RawMessage(`{"message": "hello"}`),
	}

	resp := r.Call(context.Background(), req)
	assert.False(t, resp.IsError)
	assert.Len(t, resp.Content, 1)
	assert.Equal(t, `{"message": "hello"}`, resp.Content[0].Text)

	// Call non-existent tool
	req = mcp.ToolRequest{
		ToolName: "unknown",
	}
	resp = r.Call(context.Background(), req)
	assert.True(t, resp.IsError)
	assert.Contains(t, resp.Content[0].Text, "unknown tool")
}

func TestRegistry_Has(t *testing.T) {
	r := mcp.NewRegistry()

	assert.False(t, r.Has("test"))

	tool := mcp.Tool{Name: "test"}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse { return mcp.ToolResponse{} }
	r.Register(tool, handler)

	assert.True(t, r.Has("test"))
}

func TestRegistry_Remove(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{Name: "test"}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse { return mcp.ToolResponse{} }
	r.Register(tool, handler)

	assert.True(t, r.Has("test"))
	r.Remove("test")
	assert.False(t, r.Has("test"))
}

func TestRegistry_Clear(t *testing.T) {
	r := mcp.NewRegistry()

	tool := mcp.Tool{Name: "test"}
	handler := func(ctx context.Context, req mcp.ToolRequest) mcp.ToolResponse { return mcp.ToolResponse{} }
	r.Register(tool, handler)
	r.Register(mcp.Tool{Name: "test2"}, handler)

	assert.Equal(t, 2, r.Count())
	r.Clear()
	assert.Equal(t, 0, r.Count())
}

func TestNewTextResponse(t *testing.T) {
	resp := mcp.NewTextResponse("hello world")
	assert.False(t, resp.IsError)
	assert.Len(t, resp.Content, 1)
	assert.Equal(t, "text", resp.Content[0].Type)
	assert.Equal(t, "hello world", resp.Content[0].Text)
}

func TestNewErrorResponse(t *testing.T) {
	resp := mcp.NewErrorResponse(assert.AnError)
	assert.True(t, resp.IsError)
	assert.Len(t, resp.Content, 1)
	assert.Equal(t, "text", resp.Content[0].Type)
	assert.NotEmpty(t, resp.Content[0].Text)
}

func TestNewJSONResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := mcp.NewJSONResponse(data)
	assert.False(t, resp.IsError)
	assert.Len(t, resp.Content, 1)
	assert.Contains(t, resp.Content[0].Text, `"key"`)
	assert.Contains(t, resp.Content[0].Text, `"value"`)
}

func TestNewTool(t *testing.T) {
	schema := mcp.NewObjectSchema(
		map[string]mcp.Property{
			"name": mcp.StringProp("The name"),
		},
		mcp.Required("name"),
	)

	tool := mcp.NewTool("test", "A test tool", schema)
	assert.Equal(t, "test", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.Equal(t, "object", tool.InputSchema.Type)
}

func TestNewObjectSchema(t *testing.T) {
	props := map[string]mcp.Property{
		"name":  mcp.StringProp("Name"),
		"count": mcp.IntProp("Count"),
	}
	required := []string{"name"}

	schema := mcp.NewObjectSchema(props, required)
	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 2)
	assert.Len(t, schema.Required, 1)
}

func TestPropHelpers(t *testing.T) {
	tests := []struct {
		name    string
		prop    mcp.Property
		expType string
		expDesc string
	}{
		{"StringProp", mcp.StringProp("string desc"), "string", "string desc"},
		{"IntProp", mcp.IntProp("int desc"), "integer", "int desc"},
		{"BoolProp", mcp.BoolProp("bool desc"), "boolean", "bool desc"},
		{"ObjectProp", mcp.ObjectProp("object desc"), "object", "object desc"},
		{"ArrayProp", mcp.ArrayProp("array desc"), "array", "array desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expType, tt.prop.Type)
			assert.Equal(t, tt.expDesc, tt.prop.Description)
		})
	}
}

func TestPropWithEnum(t *testing.T) {
	prop := mcp.PropWithEnum("status", "active", "paused", "finished")
	assert.Equal(t, "string", prop.Type)
	assert.Equal(t, "status", prop.Description)
	assert.Equal(t, []string{"active", "paused", "finished"}, prop.Enum)
}

func TestPropWithDefault(t *testing.T) {
	prop := mcp.PropWithDefault("integer", "count", 10)
	assert.Equal(t, "integer", prop.Type)
	assert.Equal(t, "count", prop.Description)
	assert.Equal(t, 10, prop.Default)
}
