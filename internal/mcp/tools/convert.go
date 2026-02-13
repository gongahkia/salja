package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gongahkia/salja/internal/registry"
)

// ConvertTool returns the MCP tool definition for convert.
func ConvertTool() mcp.Tool {
	return mcp.NewTool("convert",
		mcp.WithDescription("Convert between calendar/task formats"),
		mcp.WithString("input_path", mcp.Required(), mcp.Description("Path to input file")),
		mcp.WithString("output_path", mcp.Required(), mcp.Description("Path to output file")),
		mcp.WithString("from_format", mcp.Required(), mcp.Description("Source format ID")),
		mcp.WithString("to_format", mcp.Required(), mcp.Description("Target format ID")),
		mcp.WithString("conflict_strategy", mcp.Description("Conflict strategy: prefer-source, prefer-target, skip, fail")),
		mcp.WithBoolean("dry_run", mcp.Description("If true, parse only without writing")),
	)
}

// HandleConvert handles the convert tool call.
func HandleConvert(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	inputPath, _ := args["input_path"].(string)
	outputPath, _ := args["output_path"].(string)
	fromFormat, _ := args["from_format"].(string)
	toFormat, _ := args["to_format"].(string)
	dryRun, _ := args["dry_run"].(bool)

	if inputPath == "" || outputPath == "" || fromFormat == "" || toFormat == "" {
		return mcp.NewToolResultError("input_path, output_path, from_format, and to_format are required"), nil
	}

	parser, err := registry.GetParser(fromFormat)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	writer, err := registry.GetWriter(toFormat)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	col, err := parser.ParseFile(ctx, inputPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse error: %v", err)), nil
	}

	if dryRun {
		return mcp.NewToolResultText(fmt.Sprintf("Dry run: parsed %d items from %s", len(col.Items), fromFormat)), nil
	}

	if err := writer.WriteFile(ctx, col, outputPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Converted %d items from %s to %s -> %s", len(col.Items), fromFormat, toFormat, outputPath)), nil
}

// ensure HandleConvert matches the expected type
var _ server.ToolHandlerFunc = HandleConvert
