package sendto

import (
	"os"
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

func TestExtractExePath(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`"C:\Program Files\Videopress\videopress.exe" "%1"`, `C:\Program Files\Videopress\videopress.exe`},
		{`"D:\tools\videopress.exe"`, `D:\tools\videopress.exe`},
		{`D:\tools\videopress.exe "%1"`, `D:\tools\videopress.exe`},
		{``, ``},
	}

	for _, c := range cases {
		got := extractExePath(c.input)
		if got != c.expected {
			t.Errorf("extractExePath(%q) = %q; want %q", c.input, got, c.expected)
		}
	}
}

func TestRefreshLnkIfStale(t *testing.T) {
	tmpDir := t.TempDir()
	lnkFile := filepath.Join(tmpDir, "test.lnk")

	var createdLnkPath, createdTarget, createdArgs string
	origLnkCreator := lnkCreator
	origLnkTargetReader := lnkTargetReader
	defer func() {
		lnkCreator = origLnkCreator
		lnkTargetReader = origLnkTargetReader
	}()

	lnkCreator = func(lnkPath string, targetPath string, arguments string) error {
		createdLnkPath = lnkPath
		createdTarget = targetPath
		createdArgs = arguments
		return nil
	}

	// 1. 文件不存在（未安装） -> 不应创建
	err := refreshLnkIfStale(lnkFile, `C:\app\new.exe`, "--arg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdLnkPath != "" {
		t.Fatalf("expected not to create lnk when uninstalled, but created: %s", createdLnkPath)
	}

	// 创建占位 lnk 文件
	if err := os.WriteFile(lnkFile, []byte("dummy"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// 2. 目标路径已一致 -> 不应触发重写
	lnkTargetReader = func(path string) (string, error) {
		return `C:\app\new.exe`, nil
	}
	err = refreshLnkIfStale(lnkFile, `C:\app\new.exe`, "--arg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdLnkPath != "" {
		t.Fatalf("expected not to recreate lnk when target matches, got: %s", createdLnkPath)
	}

	// 3. 目标路径过时 -> 触发重写
	lnkTargetReader = func(path string) (string, error) {
		return `C:\app\old.exe`, nil
	}
	err = refreshLnkIfStale(lnkFile, `C:\app\new.exe`, "--arg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdLnkPath != lnkFile || createdTarget != `C:\app\new.exe` || createdArgs != "--arg" {
		t.Fatalf("expected recreation of stale lnk, got: %s, %s, %s", createdLnkPath, createdTarget, createdArgs)
	}
}

func TestReadLnkTargetPathChinesePath(t *testing.T) {
	tmpDir := t.TempDir()
	lnkFile := filepath.Join(tmpDir, "测试快捷方式.lnk")
	targetExe := filepath.Join(tmpDir, "中文目录", "Videopress.exe")
	if err := os.MkdirAll(filepath.Dir(targetExe), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(targetExe, []byte("exe"), 0o644); err != nil {
		t.Fatalf("write exe: %v", err)
	}

	if err := createLnk(lnkFile, targetExe, "--sendto"); err != nil {
		t.Fatalf("createLnk failed: %v", err)
	}

	got, err := readLnkTargetPath(lnkFile)
	if err != nil {
		t.Fatalf("readLnkTargetPath failed: %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(targetExe) {
		t.Fatalf("readLnkTargetPath = %q; want %q", got, targetExe)
	}
	if !isLnkTargetValid(lnkFile) {
		t.Fatalf("expected isLnkTargetValid(lnkFile) to be true for existing Chinese target")
	}
}
