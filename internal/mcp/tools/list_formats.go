package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gongahkia/salja/internal/registry"
)

// ListFormatsTool returns the MCP tool definition for list_formats.
func ListFormatsTool() mcp.Tool {
	return mcp.NewTool("list_formats",
		mcp.WithDescription("List all supported calendar/task formats"),
	)
}

type formatInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	HasParser  bool     `json:"has_parser"`
	HasWriter  bool     `json:"has_writer"`
	Events     bool     `json:"supports_events"`
	Tasks      bool     `json:"supports_tasks"`
}

// HandleListFormats handles the list_formats tool call.
func HandleListFormats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	allFmts := registry.AllFormats()
	formats := make([]formatInfo, 0, len(allFmts))

	for id, entry := range allFmts {
		formats = append(formats, formatInfo{
			ID:         id,
			Name:       entry.Name,
			Extensions: entry.Extensions,
			HasParser:  entry.NewParser != nil,
			HasWriter:  entry.NewWriter != nil,
			Events:     entry.Capabilities.SupportsEvents,
			Tasks:      entry.Capabilities.SupportsTasks,
		})
	}

	data, _ := json.MarshalIndent(formats, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

var _ server.ToolHandlerFunc = HandleListFormats
