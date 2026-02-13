package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// AuthStatusTool returns the MCP tool definition for auth_status.
func AuthStatusTool() mcp.Tool {
	return mcp.NewTool("auth_status",
		mcp.WithDescription("Check authentication status for all cloud services"),
	)
}

type authStatusEntry struct {
	Service       string `json:"service"`
	Authenticated bool   `json:"authenticated"`
	TokenExpiry   string `json:"token_expiry,omitempty"`
}

// HandleAuthStatus handles the auth_status tool call.
func HandleAuthStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	services := []string{"google", "microsoft", "todoist", "ticktick", "notion"}
	statuses := make([]authStatusEntry, 0, len(services))

	for _, svc := range services {
		// TODO: check actual keyring/token status
		statuses = append(statuses, authStatusEntry{
			Service:       svc,
			Authenticated: false,
		})
	}

	data, _ := json.MarshalIndent(statuses, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

var _ server.ToolHandlerFunc = HandleAuthStatus
