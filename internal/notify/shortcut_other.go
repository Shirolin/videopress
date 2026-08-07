//go:build !windows

package notify

// EnsureShortcut 非 Windows 平台为空实现。
func EnsureShortcut(executablePath string) error { return nil }
