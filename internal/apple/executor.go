package apple

import (
"bytes"
"fmt"
"os/exec"
"strings"
)

func RunAppleScript(script string) (string, error) {
cmd := exec.Command("osascript", "-e", script)
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

err := cmd.Run()
if err != nil {
errMsg := strings.TrimSpace(stderr.String())
if strings.Contains(errMsg, "Not authorized") || strings.Contains(errMsg, "assistive") {
return "", fmt.Errorf("permission denied: grant Automation/Accessibility access in System Preferences > Security & Privacy. Error: %s", errMsg)
}
return "", fmt.Errorf("osascript error: %s (stderr: %s)", err, errMsg)
}

return strings.TrimSpace(stdout.String()), nil
}

func CheckPermissions() error {
_, err := RunAppleScript(`tell application "System Events" to return name of first process`)
if err != nil {
return fmt.Errorf("AppleScript permissions not granted. Go to System Preferences > Security & Privacy > Privacy > Automation and grant access to Terminal/your IDE. Detail: %w", err)
}
return nil
}
