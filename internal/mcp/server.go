package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gongahkia/salja/internal/mcp/resources"
	"github.com/gongahkia/salja/internal/mcp/tools"
)

// NewServer creates and configures the MCP server with all tools and resources.
func NewServer(version string) *server.MCPServer {
	srv := server.NewMCPServer(
		"salja",
		version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
	)

	// Register tools
	srv.AddTool(tools.ConvertTool(), tools.HandleConvert)
	srv.AddTool(tools.ValidateTool(), tools.HandleValidate)
	srv.AddTool(tools.DiffTool(), tools.HandleDiff)
	srv.AddTool(tools.ListFormatsTool(), tools.HandleListFormats)
	srv.AddTool(tools.SyncPushTool(), tools.HandleSyncPush)
	srv.AddTool(tools.SyncPullTool(), tools.HandleSyncPull)
	srv.AddTool(tools.AuthStatusTool(), tools.HandleAuthStatus)

	// Register resources
	srv.AddResource(
		mcp.NewResource("salja://formats", "Supported Formats",
			mcp.WithResourceDescription("All registered format metadata"),
			mcp.WithMIMEType("application/json"),
		),
		resources.HandleFormats,
	)
	srv.AddResource(
		mcp.NewResource("salja://config", "Configuration",
			mcp.WithResourceDescription("Current salja configuration"),
			mcp.WithMIMEType("application/json"),
		),
		resources.HandleConfig,
	)
	srv.AddResourceTemplate(
		mcp.NewResourceTemplate("salja://auth/{service}", "Auth Status",
			mcp.WithTemplateDescription("Authentication status for a service"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		resources.HandleAuth,
	)

	return srv
}
