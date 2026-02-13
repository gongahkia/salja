package resources

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gongahkia/salja/internal/registry"
)

type formatMeta struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Extensions   []string `json:"extensions"`
	HasParser    bool     `json:"has_parser"`
	HasWriter    bool     `json:"has_writer"`
	Events       bool     `json:"supports_events"`
	Tasks        bool     `json:"supports_tasks"`
	Recurrence   bool     `json:"supports_recurrence"`
	Subtasks     bool     `json:"supports_subtasks"`
}

// HandleFormats returns all registered format metadata as JSON.
func HandleFormats(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	allFmts := registry.AllFormats()
	formats := make([]formatMeta, 0, len(allFmts))

	for id, entry := range allFmts {
		formats = append(formats, formatMeta{
			ID:         id,
			Name:       entry.Name,
			Extensions: entry.Extensions,
			HasParser:  entry.NewParser != nil,
			HasWriter:  entry.NewWriter != nil,
			Events:     entry.Capabilities.SupportsEvents,
			Tasks:      entry.Capabilities.SupportsTasks,
			Recurrence: entry.Capabilities.SupportsRecurrence,
			Subtasks:   entry.Capabilities.SupportsSubtasks,
		})
	}

	data, _ := json.MarshalIndent(formats, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "salja://formats",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
