//go:build windows

package app


var (
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
)

func getSystemLanguage() string {
	r, _, _ := procGetUserDefaultUILanguage.Call()
	langID := uint16(r)
	// 中文 (LANG_CHINESE) 的 Primary Language ID 为 0x04
	if (langID & 0x03ff) == 0x04 {
		return "zh"
	}
	return "en"
}
