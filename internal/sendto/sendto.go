package sendto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const launcherName = "快速压缩视频.lnk"

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

func createLnk(lnkPath string, targetPath string, arguments string) error {
	psCmd := fmt.Sprintf(
		`$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%s'); $Shortcut.TargetPath = '%s'; $Shortcut.Arguments = '%s'; $Shortcut.IconLocation = 'shell32.dll,216'; $Shortcut.Save()`,
		lnkPath, targetPath, arguments,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	return cmd.Run()
}

func sendToDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("未设置 APPDATA 环境变量")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "SendTo"), nil
}
