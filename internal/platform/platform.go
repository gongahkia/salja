package platform

import (
	"os/exec"
	"runtime"
)

type PlatformInfo struct {
	OS                    string
	Arch                  string
	AppleScriptAvailable  bool
}

func DetectOS() string   { return runtime.GOOS }
func DetectArch() string { return runtime.GOARCH }
func IsMacOS() bool      { return runtime.GOOS == "darwin" }
func IsLinux() bool      { return runtime.GOOS == "linux" }
func IsWindows() bool    { return runtime.GOOS == "windows" }

func HasAppleScriptSupport() bool {
	if !IsMacOS() {
		return false
	}
	_, err := exec.LookPath("osascript")
	return err == nil
}

func Summary() PlatformInfo {
	return PlatformInfo{
		OS:                   DetectOS(),
		Arch:                 DetectArch(),
		AppleScriptAvailable: HasAppleScriptSupport(),
	}
}
