package sendto

import (
	"path/filepath"
	"testing"
)

func TestInstallAtWritesLauncherScript(t *testing.T) {
	var calledLnkPath string
	var calledTargetPath string
	var calledArguments string

	path, err := InstallAt(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo`, `C:\tools\videopress.exe`, func(lnkPath string, targetPath string, arguments string) error {
		calledLnkPath = lnkPath
		calledTargetPath = targetPath
		calledArguments = arguments
		return nil
	})
	if err != nil {
		t.Fatalf("InstallAt returned error: %v", err)
	}

	expectedPath := filepath.Clean(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo\快速压缩视频.lnk`)
	if filepath.Clean(path) != expectedPath {
		t.Fatalf("expected launcher path %s, got %s", expectedPath, path)
	}
	if filepath.Clean(calledLnkPath) != expectedPath {
		t.Fatalf("expected create path %s, got %s", expectedPath, calledLnkPath)
	}
	if calledTargetPath != `C:\tools\videopress.exe` {
		t.Fatalf("expected targetPath, got %s", calledTargetPath)
	}
	if calledArguments != "--sendto" {
		t.Fatalf("expected arguments, got %s", calledArguments)
	}
}

func TestUninstallAtRemovesLauncherScript(t *testing.T) {
	var removed []string
	err := UninstallAt(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo`, func(path string) error {
		removed = append(removed, filepath.Clean(path))
		return nil
	})
	if err != nil {
		t.Fatalf("UninstallAt returned error: %v", err)
	}

	expectedLnk := filepath.Clean(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo\快速压缩视频.lnk`)
	expectedCmd := filepath.Clean(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo\快速压缩视频.cmd`)

	if len(removed) != 2 {
		t.Fatalf("expected 2 files removed, got %d", len(removed))
	}
	if removed[0] != expectedLnk {
		t.Errorf("expected removed lnk %s, got %s", expectedLnk, removed[0])
	}
	if removed[1] != expectedCmd {
		t.Errorf("expected removed cmd %s, got %s", expectedCmd, removed[1])
	}
}
