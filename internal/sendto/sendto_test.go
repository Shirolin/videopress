package sendto

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAtWritesLauncherScript(t *testing.T) {
	var writtenPath string
	var writtenContent []byte

	path, err := InstallAt(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo`, `C:\tools\videopress.exe`, func(path string, data []byte, perm os.FileMode) error {
		writtenPath = path
		writtenContent = bytes.Clone(data)
		return nil
	})
	if err != nil {
		t.Fatalf("InstallAt returned error: %v", err)
	}

	expectedPath := filepath.Clean(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo\快速压缩视频.cmd`)
	if filepath.Clean(path) != expectedPath {
		t.Fatalf("expected launcher path %s, got %s", expectedPath, path)
	}
	if filepath.Clean(writtenPath) != expectedPath {
		t.Fatalf("expected write path %s, got %s", expectedPath, writtenPath)
	}
	content := string(writtenContent)
	if !strings.Contains(content, `"C:\tools\videopress.exe" %*`) {
		t.Fatalf("expected launcher to call executable, got %s", content)
	}
}

func TestUninstallAtRemovesLauncherScript(t *testing.T) {
	var removed string
	err := UninstallAt(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo`, func(path string) error {
		removed = path
		return nil
	})
	if err != nil {
		t.Fatalf("UninstallAt returned error: %v", err)
	}

	expectedPath := filepath.Clean(`C:\Users\demo\AppData\Roaming\Microsoft\Windows\SendTo\快速压缩视频.cmd`)
	if filepath.Clean(removed) != expectedPath {
		t.Fatalf("expected remove path %s, got %s", expectedPath, removed)
	}
}
