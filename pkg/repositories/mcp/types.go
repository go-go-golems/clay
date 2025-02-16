package mcp

import (
	"context"
	"encoding/json"
)

// XXX(manuel, 2025-02-16) This is a temporary type to be used in the MCP repository.

type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

type ToolResult struct {
	Content []ToolContent
	IsError bool
}

type ToolContent struct {
	Type     string
	Text     string
	Data     string
	MimeType string
	Resource *ResourceContent
}

type ResourceContent struct {
	URI      string
	MimeType string
	Text     string
	Blob     string
}

type ToolProvider interface {
	// ListTools returns a list of available tools with optional pagination
	ListTools(ctx context.Context, cursor string) ([]Tool, string, error)

	// CallTool invokes a specific tool with the given arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error)
}
