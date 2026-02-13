package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcpserver "github.com/gongahkia/salja/internal/mcp"
)

func TestMCPIntegration(t *testing.T) {
	srv := mcpserver.NewServer("test")

	testClient, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("failed to create in-process client: %v", err)
	}
	defer testClient.Close()

	ctx := context.Background()
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1.0"}

	_, err = testClient.Initialize(ctx, initReq)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	t.Run("list_formats", func(t *testing.T) {
		req := mcp.CallToolRequest{}
		req.Params.Name = "list_formats"
		result, err := testClient.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
		if result.IsError {
			t.Error("expected success")
		}
		text := extractText(t, result)
		if !strings.Contains(text, "ics") {
			t.Error("expected ics in formats")
		}
	})

	t.Run("validate_missing_param", func(t *testing.T) {
		req := mcp.CallToolRequest{}
		req.Params.Name = "validate"
		result, err := testClient.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for missing params")
		}
	})

	t.Run("auth_status", func(t *testing.T) {
		req := mcp.CallToolRequest{}
		req.Params.Name = "auth_status"
		result, err := testClient.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
		if result.IsError {
			t.Error("expected success")
		}
		text := extractText(t, result)
		var statuses []map[string]any
		if err := json.Unmarshal([]byte(text), &statuses); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(statuses) != 5 {
			t.Errorf("expected 5 services, got %d", len(statuses))
		}
	})

	t.Run("convert_invalid_format", func(t *testing.T) {
		req := mcp.CallToolRequest{}
		req.Params.Name = "convert"
		req.Params.Arguments = map[string]any{
			"input_path":  "/tmp/in.ics",
			"output_path": "/tmp/out.csv",
			"from_format": "nosuchformat",
			"to_format":   "ics",
		}
		result, err := testClient.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for invalid format")
		}
	})

	t.Run("diff_missing_params", func(t *testing.T) {
		req := mcp.CallToolRequest{}
		req.Params.Name = "diff"
		req.Params.Arguments = map[string]any{"file_a": "/tmp/a.ics"}
		result, err := testClient.CallTool(ctx, req)
		if err != nil {
			t.Fatalf("call failed: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for missing file_b")
		}
	})
}

func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatal("no text content found")
	return ""
}
