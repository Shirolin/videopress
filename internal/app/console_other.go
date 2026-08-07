//go:build !windows

package app

// AttachSendToConsole 非 Windows 平台为空实现。
func AttachSendToConsole() bool { return false }

// DetachConsole 非 Windows 平台为空实现。
func DetachConsole() {}
