package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SyncPushTool returns the MCP tool definition for sync_push.
func SyncPushTool() mcp.Tool {
	return mcp.NewTool("sync_push",
		mcp.WithDescription("Push local calendar/task data to a cloud service"),
		mcp.WithString("service", mcp.Required(), mcp.Description("Service name: google, microsoft, todoist, ticktick, notion")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to file to push")),
		mcp.WithString("start_date", mcp.Description("Start date filter (ISO 8601)")),
		mcp.WithString("end_date", mcp.Description("End date filter (ISO 8601)")),
		mcp.WithBoolean("dry_run", mcp.Description("If true, simulate without pushing")),
	)
}

// SyncPullTool returns the MCP tool definition for sync_pull.
func SyncPullTool() mcp.Tool {
	return mcp.NewTool("sync_pull",
		mcp.WithDescription("Pull calendar/task data from a cloud service"),
		mcp.WithString("service", mcp.Required(), mcp.Description("Service name: google, microsoft, todoist, ticktick, notion")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to write pulled data")),
		mcp.WithString("start_date", mcp.Description("Start date filter (ISO 8601)")),
		mcp.WithString("end_date", mcp.Description("End date filter (ISO 8601)")),
		mcp.WithBoolean("dry_run", mcp.Description("If true, simulate without pulling")),
	)
}

// HandleSyncPush handles the sync_push tool call.
func HandleSyncPush(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	service, _ := args["service"].(string)
	if service == "" {
		return mcp.NewToolResultError("service is required"), nil
	}

	// TODO: wire to internal/api sync push
	return mcp.NewToolResultError("sync_push not yet implemented for " + service), nil
}

// HandleSyncPull handles the sync_pull tool call.
func HandleSyncPull(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	service, _ := args["service"].(string)
	if service == "" {
		return mcp.NewToolResultError("service is required"), nil
	}

	// TODO: wire to internal/api sync pull
	return mcp.NewToolResultError("sync_pull not yet implemented for " + service), nil
}

var _ server.ToolHandlerFunc = HandleSyncPush
var _ server.ToolHandlerFunc = HandleSyncPull
