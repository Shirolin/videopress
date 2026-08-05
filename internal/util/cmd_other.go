//go:build !windows

package util

import "os/exec"

// HideConsoleWindow 非 Windows 平台为空实现。
func HideConsoleWindow(cmd *exec.Cmd) {}
