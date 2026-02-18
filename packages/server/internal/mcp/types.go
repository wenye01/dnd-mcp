// Package mcp provides MCP (Model Context Protocol) types and server implementation
package mcp

import (
	"context"
	"encoding/json"
)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema represents the JSON Schema for tool input
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// Property represents a property in the input schema
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// ToolRequest represents an incoming tool call request
type ToolRequest struct {
	ToolName string          `json:"tool_name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents content in a tool response
type Content struct {
	Type string `json:"type"` // "text" or "image"
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"` // base64 for images
	MIMEType string `json:"mimeType,omitempty"`
}

// NewTextResponse creates a text response
func NewTextResponse(text string) ToolResponse {
	return ToolResponse{
		Content: []Content{
			{Type: "text", Text: text},
		},
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) ToolResponse {
	return ToolResponse{
		Content: []Content{
			{Type: "text", Text: err.Error()},
		},
		IsError: true,
	}
}

// NewJSONResponse creates a JSON response from any value
func NewJSONResponse(v interface{}) ToolResponse {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return NewErrorResponse(err)
	}
	return NewTextResponse(string(data))
}

// ToolHandler is the function signature for tool handlers
type ToolHandler func(ctx context.Context, req ToolRequest) ToolResponse

// Context provides context for tool execution
type Context struct {
	CampaignID string
	UserID     string
}

// ToolInfo contains tool metadata and handler
type ToolInfo struct {
	Tool    Tool
	Handler ToolHandler
}

// ListToolsResponse represents the response for listing tools
type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// CallToolRequest represents a tool call request (MCP protocol format)
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResponse represents a tool call response (MCP protocol format)
type CallToolResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// InitializeRequest represents an MCP initialize request
type InitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities,omitempty"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo contains client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResponse represents an MCP initialize response
type InitializeResponse struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerCapabilities     `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerCapabilities describes server capabilities
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability describes tool capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo contains server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
