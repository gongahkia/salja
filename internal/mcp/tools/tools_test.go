package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func makeRequest(name string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

func TestConvertMissingParams(t *testing.T) {
	req := makeRequest("convert", map[string]any{})
	result, err := HandleConvert(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing params")
	}
}

func TestConvertInvalidFormat(t *testing.T) {
	req := makeRequest("convert", map[string]any{
		"input_path":  "/tmp/test.ics",
		"output_path": "/tmp/out.csv",
		"from_format": "nonexistent",
		"to_format":   "ics",
	})
	result, err := HandleConvert(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid format")
	}
}

func TestValidateMissingPath(t *testing.T) {
	req := makeRequest("validate", map[string]any{})
	result, err := HandleValidate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing file_path")
	}
}

func TestValidateNonexistentFile(t *testing.T) {
	req := makeRequest("validate", map[string]any{
		"file_path": "/tmp/nonexistent_salja_test_file_xyz_99999.nosuchext",
	})
	result, err := HandleValidate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for nonexistent file")
	}
}

func TestDiffMissingParams(t *testing.T) {
	req := makeRequest("diff", map[string]any{"file_a": "/tmp/a.ics"})
	result, err := HandleDiff(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing file_b")
	}
}

func TestListFormatsReturnsResults(t *testing.T) {
	req := makeRequest("list_formats", map[string]any{})
	result, err := HandleListFormats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := ""
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			text = tc.Text
		}
	}
	if !strings.Contains(text, "ics") {
		t.Error("expected list_formats to include ics format")
	}
}

func TestAuthStatusReturnsAllServices(t *testing.T) {
	req := makeRequest("auth_status", map[string]any{})
	result, err := HandleAuthStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success result")
	}
	text := ""
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			text = tc.Text
		}
	}
	for _, svc := range []string{"google", "microsoft", "todoist", "ticktick", "notion"} {
		if !strings.Contains(text, svc) {
			t.Errorf("expected auth_status to include %s", svc)
		}
	}
}

func TestSyncPushMissingService(t *testing.T) {
	req := makeRequest("sync_push", map[string]any{})
	result, err := HandleSyncPush(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing service")
	}
}

func TestSyncPullMissingService(t *testing.T) {
	req := makeRequest("sync_pull", map[string]any{})
	result, err := HandleSyncPull(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing service")
	}
}
