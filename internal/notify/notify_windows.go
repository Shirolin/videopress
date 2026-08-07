//go:build windows

package notify

import "github.com/gen2brain/beeep"

// AppID 是 Windows Toast 通知的 AppUserModelID，须与开始菜单快捷方式的
// System.AppUserModel.ID 属性一致。
const AppID = "Shirolin.Videopress"

func notify(title, message string) error {
	beeep.AppName = AppID
	return beeep.Notify(title, message, "")
}
