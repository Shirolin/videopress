package sendto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const launcherName = "快速压缩视频.lnk"
const appLnkName = "Videopress.lnk"
const startMenuSubDir = "Videopress"

type CreateLnkFunc func(lnkPath string, targetPath string, arguments string) error
type RemoveFunc func(name string) error

func Install(executablePath string) (string, error) {
	sendToDir, err := sendToDir()
	if err != nil {
		return "", err
	}
	return InstallAt(sendToDir, executablePath, createLnk)
}

func Uninstall() error {
	sendToDir, err := sendToDir()
	if err != nil {
		return err
	}
	return UninstallAt(sendToDir, os.Remove)
}

func InstallAt(sendToDir string, executablePath string, createLnk CreateLnkFunc) (string, error) {
	launcherPath := filepath.Join(sendToDir, launcherName)
	if err := createLnk(launcherPath, executablePath, "--sendto"); err != nil {
		return "", err
	}
	return launcherPath, nil
}

func UninstallAt(sendToDir string, remove RemoveFunc) error {
	lnkPath := filepath.Join(sendToDir, launcherName)
	_ = remove(lnkPath)

	// 兼容老版本清除
	oldCmdPath := filepath.Join(sendToDir, "快速压缩视频.cmd")
	_ = remove(oldCmdPath)

	return nil
}

// InstallDesktop 创建桌面快捷方式
func InstallDesktop(executablePath string) error {
	desktopDir, err := getDesktopDir()
	if err != nil {
		return err
	}
	lnkPath := filepath.Join(desktopDir, appLnkName)
	return createLnk(lnkPath, executablePath, "")
}

// UninstallDesktop 卸载桌面快捷方式
func UninstallDesktop() error {
	desktopDir, err := getDesktopDir()
	if err != nil {
		return err
	}
	lnkPath := filepath.Join(desktopDir, appLnkName)
	if _, err := os.Stat(lnkPath); err == nil {
		return os.Remove(lnkPath)
	}
	return nil
}

// InstallStartMenu 创建开始菜单快捷方式
func InstallStartMenu(executablePath string) error {
	startMenuDir, err := getStartMenuProgramsDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(startMenuDir, startMenuSubDir)
	_ = os.MkdirAll(appDir, 0755)

	lnkPath := filepath.Join(appDir, appLnkName)
	return createLnk(lnkPath, executablePath, "")
}

// UninstallStartMenu 卸载开始菜单快捷方式
func UninstallStartMenu() error {
	startMenuDir, err := getStartMenuProgramsDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(startMenuDir, startMenuSubDir)
	return os.RemoveAll(appDir)
}

// RegisterContextMenu 注册右键直接压缩菜单 (注册表 HKCU，免管理员)
func RegisterContextMenu(executablePath string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("创建右键注册表项失败: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue("MUIVerb", "使用 Videopress 压缩"); err != nil {
		return fmt.Errorf("设置 MUIVerb 失败: %w", err)
	}
	if err := k.SetStringValue("Icon", executablePath); err != nil {
		return fmt.Errorf("设置 Icon 失败: %w", err)
	}

	cmdK, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("创建 Command 子项失败: %w", err)
	}
	defer cmdK.Close()

	formattedCmd := fmt.Sprintf(`"%s" "%%1"`, executablePath)
	if err := cmdK.SetStringValue("", formattedCmd); err != nil {
		return fmt.Errorf("设置默认值失败: %w", err)
	}
	return nil
}

// UnregisterContextMenu 卸载右键菜单
func UnregisterContextMenu() error {
	// 先删除子项 command
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress\command`)
	// 再删除主项
	err := registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress`)
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("删除右键菜单注册表项失败: %w", err)
	}
	return nil
}

func createLnk(lnkPath string, targetPath string, arguments string) error {
	psCmd := fmt.Sprintf(
		`$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%s'); $Shortcut.TargetPath = '%s'; $Shortcut.Arguments = '%s'; $Shortcut.Save()`,
		lnkPath, targetPath, arguments,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd.Run()
}

func sendToDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "SendTo"), nil
}

func getDesktopDir() (string, error) {
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return "", fmt.Errorf("未设置 USERPROFILE 环境变量")
	}
	return filepath.Join(userProfile, "Desktop"), nil
}

func getStartMenuProgramsDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs"), nil
}

// IsSendToInstalled 检查 SendTo 快捷方式是否存在
func IsSendToInstalled() bool {
	dir, err := sendToDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, launcherName))
	return err == nil
}

// IsDesktopInstalled 检查桌面快捷方式是否存在
func IsDesktopInstalled() bool {
	dir, err := getDesktopDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, appLnkName))
	return err == nil
}

// IsStartMenuInstalled 检查开始菜单快捷方式是否存在
func IsStartMenuInstalled() bool {
	dir, err := getStartMenuProgramsDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, startMenuSubDir, appLnkName))
	return err == nil
}

// IsContextMenuInstalled 检查右键注册表项是否存在
func IsContextMenuInstalled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\*\shell\Videopress`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	return true
}

