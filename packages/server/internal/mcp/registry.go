// Package mcp provides MCP (Model Context Protocol) types and server implementation
package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages tool registration and lookup
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]*ToolInfo
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*ToolInfo),
	}
}

// Register registers a tool with its handler
func (r *Registry) Register(tool Tool, handler ToolHandler) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}

	r.tools[tool.Name] = &ToolInfo{
		Tool:    tool,
		Handler: handler,
	}

	return nil
}

// MustRegister registers a tool and panics on error
func (r *Registry) MustRegister(tool Tool, handler ToolHandler) {
	if err := r.Register(tool, handler); err != nil {
		panic(err)
	}
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*ToolInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.tools[name]
	return info, ok
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, info := range r.tools {
		tools = append(tools, info.Tool)
	}
	return tools
}

// Call executes a tool by name
func (r *Registry) Call(ctx context.Context, req ToolRequest) ToolResponse {
	r.mu.RLock()
	info, ok := r.tools[req.ToolName]
	r.mu.RUnlock()

	if !ok {
		return NewErrorResponse(fmt.Errorf("unknown tool: %s", req.ToolName))
	}

	return info.Handler(ctx, req)
}

// Has checks if a tool is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.tools[name]
	return ok
}

// Remove removes a tool from the registry
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tools, name)
}

// Clear removes all tools from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]*ToolInfo)
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Helper functions for creating common tools

// NewTool creates a new tool definition
func NewTool(name, description string, schema InputSchema) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: schema,
	}
}

// NewObjectSchema creates an object input schema
func NewObjectSchema(properties map[string]Property, required []string) InputSchema {
	return InputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

// Prop creates a property with type and description
func Prop(typ, desc string) Property {
	return Property{
		Type:        typ,
		Description: desc,
	}
}

// PropWithEnum creates a property with enum values
func PropWithEnum(desc string, enum ...string) Property {
	return Property{
		Type:        "string",
		Description: desc,
		Enum:        enum,
	}
}

// PropWithDefault creates a property with a default value
func PropWithDefault(typ, desc string, def interface{}) Property {
	return Property{
		Type:        typ,
		Description: desc,
		Default:     def,
	}
}

// Required marks property names as required
func Required(names ...string) []string {
	return names
}

// Common property types
var (
	StringProp = func(desc string) Property { return Prop("string", desc) }
	IntProp    = func(desc string) Property { return Prop("integer", desc) }
	BoolProp   = func(desc string) Property { return Prop("boolean", desc) }
	ObjectProp = func(desc string) Property { return Prop("object", desc) }
	ArrayProp  = func(desc string) Property { return Prop("array", desc) }
)
