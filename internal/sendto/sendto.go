package sendto

import (
	"fmt"
	"os"
	"path/filepath"
)

const launcherName = "快速压缩视频.cmd"

type WriteFileFunc func(name string, data []byte, perm os.FileMode) error
type RemoveFunc func(name string) error

func Install(executablePath string) (string, error) {
	sendToDir, err := sendToDir()
	if err != nil {
		return "", err
	}
	return InstallAt(sendToDir, executablePath, os.WriteFile)
}

func Uninstall() error {
	sendToDir, err := sendToDir()
	if err != nil {
		return err
	}
	return UninstallAt(sendToDir, os.Remove)
}

func InstallAt(sendToDir string, executablePath string, writeFile WriteFileFunc) (string, error) {
	launcherPath := filepath.Join(sendToDir, launcherName)
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", executablePath)
	if err := writeFile(launcherPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return launcherPath, nil
}

func UninstallAt(sendToDir string, remove RemoveFunc) error {
	launcherPath := filepath.Join(sendToDir, launcherName)
	if err := remove(launcherPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sendToDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA is not set")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "SendTo"), nil
}
