//go:build windows

package locale

import "syscall"

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
)

// System 返回系统 UI 语言（zh 或 en）。
func System() string {
	r, _, _ := procGetUserDefaultUILanguage.Call()
	langID := uint16(r)
	if (langID & 0x03ff) == 0x04 {
		return "zh"
	}
	return "en"
}
