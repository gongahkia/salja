//go:build darwin

package appdetect

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func DetectAll() []DetectedApp {
	home, _ := os.UserHomeDir()
	var apps []DetectedApp
	apps = append(apps, DetectedApp{
		Name: "Apple Calendar", FormatName: "apple-calendar",
		Installed: true, DataPaths: existingPaths(filepath.Join(home, "Library", "Calendars")),
	})
	apps = append(apps, DetectedApp{
		Name: "Apple Reminders", FormatName: "apple-reminders",
		Installed: true,
	})
	apps = append(apps, checkBundleID("Todoist", "todoist", "com.todoist.mac")...)
	apps = append(apps, checkBundleID("TickTick", "ticktick", "com.TickTick.task.mac")...)
	apps = append(apps, checkBundleID("Notion", "notion", "notion.id")...)
	apps = append(apps, checkBundleID("OmniFocus", "omnifocus", "com.omnigroup.OmniFocus3")...)
	apps = append(apps, DetectedApp{
		Name: "Google Calendar", FormatName: "gcal", Installed: false,
	})
	apps = append(apps, DetectedApp{
		Name: "Outlook", FormatName: "outlook",
		Installed: bundleExists("com.microsoft.Outlook"),
	})
	return apps
}

func checkBundleID(name, format, bundleID string) []DetectedApp {
	return []DetectedApp{{
		Name: name, FormatName: format, Installed: bundleExists(bundleID),
	}}
}

func bundleExists(bundleID string) bool {
	out, err := exec.Command("mdfind", "kMDItemCFBundleIdentifier == '"+bundleID+"'").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func existingPaths(paths ...string) []string {
	var found []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}
