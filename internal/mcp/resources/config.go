package resources

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gongahkia/salja/internal/config"
)

// HandleConfig returns the current salja configuration as JSON.
func HandleConfig(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "salja://config",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
