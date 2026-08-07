package notify

// Notify 发送系统桌面通知。
// Windows 下为 Toast 通知，要求开始菜单快捷方式已带 AppUserModelID（调用方需先 EnsureShortcut），
// 其他平台为空实现。
func Notify(title, message string) error {
	return notify(title, message)
}
