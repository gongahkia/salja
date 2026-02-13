package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/registry"
)

// ValidateTool returns the MCP tool definition for validate.
func ValidateTool() mcp.Tool {
	return mcp.NewTool("validate",
		mcp.WithDescription("Validate a calendar/task file"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to file to validate")),
	)
}

type validateResponse struct {
	ItemCount      int            `json:"item_count"`
	FormatDetected string         `json:"format_detected"`
	FieldCoverage  map[string]int `json:"field_coverage"`
	ParseErrors    []string       `json:"parse_errors"`
}

// HandleValidate handles the validate tool call.
func HandleValidate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	allFmts := registry.AllFormats()
	for id, entry := range allFmts {
		if entry.NewParser == nil {
			continue
		}
		parser := entry.NewParser()
		col, err := parser.ParseFile(ctx, filePath)
		if err != nil || len(col.Items) == 0 {
			continue
		}

		coverage := map[string]int{
			"title": 0, "description": 0, "start_time": 0, "due_date": 0, "tags": 0,
		}
		for _, item := range col.Items {
			if item.Title != "" {
				coverage["title"]++
			}
			if item.Description != "" {
				coverage["description"]++
			}
			if item.StartTime != nil {
				coverage["start_time"]++
			}
			if item.DueDate != nil {
				coverage["due_date"]++
			}
			if len(item.Tags) > 0 {
				coverage["tags"]++
			}
		}

		events, tasks := 0, 0
		for _, item := range col.Items {
			switch item.ItemType {
			case model.ItemTypeEvent:
				events++
			case model.ItemTypeTask:
				tasks++
			}
		}

		resp := validateResponse{
			ItemCount:      len(col.Items),
			FormatDetected: id,
			FieldCoverage:  coverage,
			ParseErrors:    []string{},
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return mcp.NewToolResultError(fmt.Sprintf("no parser could read %s", filePath)), nil
}

var _ server.ToolHandlerFunc = HandleValidate
