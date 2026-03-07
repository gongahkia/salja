package platform

import (
	"runtime"
	"testing"
)

func TestDetectOS(t *testing.T) {
	got := DetectOS()
	if got == "" {
		t.Fatal("DetectOS returned empty string")
	}
	if got != runtime.GOOS {
		t.Fatalf("DetectOS=%q, want %q", got, runtime.GOOS)
	}
}

func TestDetectArch(t *testing.T) {
	got := DetectArch()
	if got == "" {
		t.Fatal("DetectArch returned empty string")
	}
	if got != runtime.GOARCH {
		t.Fatalf("DetectArch=%q, want %q", got, runtime.GOARCH)
	}
}

func TestSummary(t *testing.T) {
	s := Summary()
	if s.OS == "" || s.Arch == "" {
		t.Fatalf("Summary has empty fields: %+v", s)
	}
}

func TestPlatformBooleans(t *testing.T) {
	switch runtime.GOOS {
	case "darwin":
		if !IsMacOS() {
			t.Fatal("expected IsMacOS=true on darwin")
		}
	case "linux":
		if !IsLinux() {
			t.Fatal("expected IsLinux=true on linux")
		}
	case "windows":
		if !IsWindows() {
			t.Fatal("expected IsWindows=true on windows")
		}
	}
}

func TestHasAppleScriptSupport(t *testing.T) {
	got := HasAppleScriptSupport()
	if runtime.GOOS != "darwin" && got {
		t.Fatal("HasAppleScriptSupport should be false on non-darwin")
	}
}
