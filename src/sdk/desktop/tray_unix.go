// +build !windows

package desktop

import (
	"fmt"
	"os/exec"
	"runtime"
)

// PlatformTrayIcon implements TrayIcon for Unix-like systems
type PlatformTrayIcon struct {
	tooltip string
	icon    string
	menu    []MenuItem
	onClick func()
	onQuit  func()
}

// NewPlatformTrayIcon creates a new tray icon
func NewPlatformTrayIcon() (TrayIcon, error) {
	return &PlatformTrayIcon{}, nil
}

func (t *PlatformTrayIcon) Initialize() error {
	return nil
}

func (t *PlatformTrayIcon) SetTooltip(text string) {
	t.tooltip = text
}

func (t *PlatformTrayIcon) SetIcon(iconPath string) {
	t.icon = iconPath
}

func (t *PlatformTrayIcon) ShowNotification(title, message string) {
	// Use notify-send on Linux
	if runtime.GOOS == "linux" {
		exec.Command("notify-send", title, message).Run()
	}
}

func (t *PlatformTrayIcon) SetMenu(items []MenuItem) {
	t.menu = items
}

func (t *PlatformTrayIcon) OnClick(handler func()) {
	t.onClick = handler
}

func (t *PlatformTrayIcon) OnQuit(handler func()) {
	t.onQuit = handler
}

func (t *PlatformTrayIcon) Run() error {
	// Simple implementation - in real app would use systray library
	select {}
}

func (t *PlatformTrayIcon) Stop() {}

// OpenURL opens a URL in the default browser
func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenFile opens a file with the default application
func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}