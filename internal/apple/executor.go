package apple

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	salerr "github.com/gongahkia/salja/internal/errors"
)

// ErrNotMacOS is returned when Apple-specific features are used on non-macOS systems.
var ErrNotMacOS = fmt.Errorf("this feature requires macOS (current platform: %s/%s)", runtime.GOOS, runtime.GOARCH)

// ScriptRunner abstracts AppleScript execution for testability.
type ScriptRunner interface {
	Run(script string) (string, error)
}

// scriptRunnerFn is the package-level function used to run AppleScript.
// Override in tests to mock AppleScript execution.
var scriptRunnerFn = RunAppleScript

func RunAppleScript(script string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", ErrNotMacOS
	}

	cmd := exec.Command("osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if strings.Contains(errMsg, "Not authorized") || strings.Contains(errMsg, "assistive") {
			return "", &salerr.PermissionError{Resource: "AppleScript Automation", Message: "grant Automation/Accessibility access in System Preferences > Security & Privacy", Err: fmt.Errorf("%s", errMsg)}
		}
		return "", fmt.Errorf("osascript error: %s (stderr: %s)", err, errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

func CheckPermissions() error {
	if runtime.GOOS != "darwin" {
		return ErrNotMacOS
	}
	_, err := RunAppleScript(`tell application "System Events" to return name of first process`)
	if err != nil {
		return &salerr.PermissionError{Resource: "AppleScript Automation", Message: "Go to System Preferences > Security & Privacy > Privacy > Automation and grant access to Terminal/your IDE", Err: err}
	}
	return nil
}
