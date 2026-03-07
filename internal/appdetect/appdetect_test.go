package appdetect

import (
	"testing"
)

func TestDetectAllReturnsSlice(t *testing.T) {
	apps := DetectAll()
	if apps == nil {
		t.Fatal("DetectAll returned nil")
	}
	for _, app := range apps {
		if app.Name == "" {
			t.Error("DetectedApp has empty Name")
		}
		if app.FormatName == "" {
			t.Error("DetectedApp has empty FormatName")
		}
	}
}
