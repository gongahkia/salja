//go:build linux

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
	apps = append(apps, checkDesktop("Todoist", "todoist", "todoist")...)
	apps = append(apps, checkDesktop("TickTick", "ticktick", "ticktick")...)
	apps = append(apps, checkDesktop("Notion", "notion", "notion")...)
	apps = append(apps, checkDesktopWithData("OmniFocus", "omnifocus", "omnifocus",
		filepath.Join(home, ".config", "omnifocus"))...)
	apps = append(apps, DetectedApp{
		Name: "Google Calendar", FormatName: "gcal", Installed: false,
	})
	apps = append(apps, checkDesktop("Outlook", "outlook", "outlook")...)
	return apps
}

func checkDesktop(name, format, keyword string) []DetectedApp {
	return []DetectedApp{{
		Name: name, FormatName: format, Installed: desktopFileExists(keyword) || flatpakExists(keyword) || snapExists(keyword),
	}}
}

func checkDesktopWithData(name, format, keyword, dataPath string) []DetectedApp {
	app := DetectedApp{
		Name: name, FormatName: format,
		Installed: desktopFileExists(keyword) || flatpakExists(keyword) || snapExists(keyword),
	}
	if _, err := os.Stat(dataPath); err == nil {
		app.DataPaths = []string{dataPath}
	}
	return []DetectedApp{app}
}

func desktopFileExists(keyword string) bool {
	dirs := []string{"/usr/share/applications"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "share", "applications"))
	}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), keyword) && strings.HasSuffix(e.Name(), ".desktop") {
				return true
			}
		}
	}
	return false
}

func flatpakExists(keyword string) bool {
	out, err := exec.Command("flatpak", "list", "--app", "--columns=application").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), keyword)
}

func snapExists(keyword string) bool {
	out, err := exec.Command("snap", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), keyword)
}
