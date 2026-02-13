package resources

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type authState struct {
	Service       string `json:"service"`
	Authenticated bool   `json:"authenticated"`
	TokenExpiry   string `json:"token_expiry,omitempty"`
}

// HandleAuth returns authentication state for a specific service.
func HandleAuth(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	// Extract service from salja://auth/{service}
	service := strings.TrimPrefix(uri, "salja://auth/")

	// TODO: check actual keyring/token status for the service
	state := authState{
		Service:       service,
		Authenticated: false,
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
