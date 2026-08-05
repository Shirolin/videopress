//go:build windows

package util

import (
	"os/exec"
	"syscall"
)

// HideConsoleWindow 隐藏子进程控制台窗口，避免 GUI 模式下弹出黑色命令行窗口。
func HideConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000 // CREATE_NO_WINDOW
}
