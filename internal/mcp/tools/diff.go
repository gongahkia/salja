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

// DiffTool returns the MCP tool definition for diff.
func DiffTool() mcp.Tool {
	return mcp.NewTool("diff",
		mcp.WithDescription("Compare two calendar/task files"),
		mcp.WithString("file_a", mcp.Required(), mcp.Description("Path to first file")),
		mcp.WithString("file_b", mcp.Required(), mcp.Description("Path to second file")),
		mcp.WithString("output_format", mcp.Description("Output format: json, table (default: json)")),
	)
}

type diffResponse struct {
	FileA   string   `json:"file_a"`
	FileB   string   `json:"file_b"`
	Summary diffStat `json:"summary"`
}

type diffStat struct {
	EventsA int `json:"events_a"`
	EventsB int `json:"events_b"`
	TasksA  int `json:"tasks_a"`
	TasksB  int `json:"tasks_b"`
	ItemsA  int `json:"items_a"`
	ItemsB  int `json:"items_b"`
}

// HandleDiff handles the diff tool call.
func HandleDiff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	fileA, _ := args["file_a"].(string)
	fileB, _ := args["file_b"].(string)

	if fileA == "" || fileB == "" {
		return mcp.NewToolResultError("file_a and file_b are required"), nil
	}

	allFmts := registry.AllFormats()
	for _, entry := range allFmts {
		if entry.NewParser == nil {
			continue
		}
		parser := entry.NewParser()
		colA, errA := parser.ParseFile(ctx, fileA)
		colB, errB := parser.ParseFile(ctx, fileB)
		if errA != nil || errB != nil {
			continue
		}

		countTypes := func(items []model.CalendarItem) (events, tasks int) {
			for _, item := range items {
				switch item.ItemType {
				case model.ItemTypeEvent:
					events++
				case model.ItemTypeTask:
					tasks++
				}
			}
			return
		}

		eA, tA := countTypes(colA.Items)
		eB, tB := countTypes(colB.Items)

		resp := diffResponse{
			FileA: fileA,
			FileB: fileB,
			Summary: diffStat{
				EventsA: eA, EventsB: eB,
				TasksA: tA, TasksB: tB,
				ItemsA: len(colA.Items), ItemsB: len(colB.Items),
			},
		}

		data, _ := json.MarshalIndent(resp, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	return mcp.NewToolResultError(fmt.Sprintf("could not parse both files: %s and %s", fileA, fileB)), nil
}

var _ server.ToolHandlerFunc = HandleDiff
