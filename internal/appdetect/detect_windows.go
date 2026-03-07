//go:build windows

package appdetect

import (
	"os/exec"
	"strings"
)

func DetectAll() []DetectedApp {
	var apps []DetectedApp
	apps = append(apps, checkWinApp("Todoist", "todoist", "Todoist")...)
	apps = append(apps, checkWinApp("TickTick", "ticktick", "TickTick")...)
	apps = append(apps, checkWinApp("Notion", "notion", "Notion")...)
	apps = append(apps, checkWinApp("Outlook", "outlook", "Outlook")...)
	apps = append(apps, DetectedApp{
		Name: "Google Calendar", FormatName: "gcal", Installed: false,
	})
	return apps
}

func checkWinApp(name, format, displayName string) []DetectedApp {
	return []DetectedApp{{
		Name: name, FormatName: format, Installed: winAppInstalled(displayName),
	}}
}

func winAppInstalled(displayName string) bool {
	out, err := exec.Command("reg", "query",
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		"/s", "/f", displayName).Output()
	if err == nil && strings.Contains(string(out), displayName) {
		return true
	}
	out, err = exec.Command("reg", "query",
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		"/s", "/f", displayName).Output()
	return err == nil && strings.Contains(string(out), displayName)
}
