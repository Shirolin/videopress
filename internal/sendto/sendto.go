package sendto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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
	psCmd := fmt.Sprintf(
		`$Path = 'HKCU:\Software\Classes\*\shell\Videopress'; New-Item -Path $Path -Force | Out-Null; Set-ItemProperty -Path $Path -Name 'MUIVerb' -Value '使用 Videopress 压缩' -Force; Set-ItemProperty -Path $Path -Name 'Icon' -Value '%s' -Force; $CommandPath = "$Path\command"; New-Item -Path $CommandPath -Force | Out-Null; Set-Item -Path $CommandPath -Value '"%s" "%%1"' -Force`,
		executablePath, executablePath,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	return cmd.Run()
}

// UnregisterContextMenu 卸载右键菜单
func UnregisterContextMenu() error {
	psCmd := `Remove-Item -Path 'HKCU:\Software\Classes\*\shell\Videopress' -Recurse -Force -ErrorAction SilentlyContinue`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	return cmd.Run()
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
	psCmd := `Get-Item -Path 'HKCU:\Software\Classes\*\shell\Videopress' -ErrorAction Stop`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	return cmd.Run() == nil
}

