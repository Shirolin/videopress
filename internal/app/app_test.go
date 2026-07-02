package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordedCall struct {
	name string
	args []string
}

func accessibleInput(_ string) bool { return true }

func TestExecuteUsesStandardPresetByDefault(t *testing.T) {
	var calls []recordedCall
	var createdDirs []string
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{`C:\videos\clip.mov`}, Dependencies{
		ExecutableDir:  `C:\tools`,
		ExecutablePath: `C:\tools\videopress.exe`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			calls = append(calls, recordedCall{name: name, args: args})
			return nil
		},
		MkdirAll: func(path string, perm os.FileMode) error {
			createdDirs = append(createdDirs, path)
			return nil
		},
		PathExists:      func(path string) bool { return false },
		InputAccessible: accessibleInput,
		Stdout:          stdout,
		Stderr:          stderr,
	})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", exitCode, stderr.String())
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 ffmpeg call, got %d", len(calls))
	}
	if calls[0].name != `C:\ffmpeg\bin\ffmpeg.exe` {
		t.Fatalf("expected ffmpeg binary path, got %s", calls[0].name)
	}

	joined := strings.Join(calls[0].args, " ")
	if !strings.Contains(joined, "-crf 27") {
		t.Fatalf("expected standard preset args, got %s", joined)
	}
	expectedDir := filepath.Clean(`C:\videos\compressed`)
	if len(createdDirs) != 1 || filepath.Clean(createdDirs[0]) != expectedDir {
		t.Fatalf("expected created dir %s, got %+v", expectedDir, createdDirs)
	}
}

func TestExecuteSkipsNonVideoFilesButKeepsGoing(t *testing.T) {
	var calls []recordedCall
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{`C:\videos\readme.txt`, `C:\videos\clip.mp4`}, Dependencies{
		ExecutableDir:  `C:\tools`,
		ExecutablePath: `C:\tools\videopress.exe`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			calls = append(calls, recordedCall{name: name, args: args})
			return nil
		},
		MkdirAll:        func(path string, perm os.FileMode) error { return nil },
		PathExists:      func(path string) bool { return false },
		InputAccessible: accessibleInput,
		Stdout:          stdout,
		Stderr:     stderr,
	})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 ffmpeg call, got %d", len(calls))
	}
	if !strings.Contains(stdout.String(), "跳过非视频文件") {
		t.Fatalf("expected skip message, got %s", stdout.String())
	}
}

func TestExecuteReturnsNonZeroWhenCompressionFails(t *testing.T) {
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{`C:\videos\clip.mp4`}, Dependencies{
		ExecutableDir:  `C:\tools`,
		ExecutablePath: `C:\tools\videopress.exe`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			return errors.New("ffmpeg failed")
		},
		MkdirAll:        func(path string, perm os.FileMode) error { return nil },
		PathExists:      func(path string) bool { return false },
		InputAccessible: accessibleInput,
		Stdout:          &bytes.Buffer{},
		Stderr:     stderr,
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "ffmpeg failed") {
		t.Fatalf("expected failure message, got %s", stderr.String())
	}
}

func TestExecutePrintsVersion(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{"--version"}, Dependencies{
		Stdout: stdout,
		Stderr: stderr,
	})

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if strings.TrimSpace(stdout.String()) != Version {
		t.Fatalf("expected version %s, got %q", Version, stdout.String())
	}
}

func TestExecutePrintsHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	for _, arg := range []string{"-h", "--help"} {
		stdout.Reset()
		stderr.Reset()
		exitCode := Execute([]string{arg}, Dependencies{
			Stdout: stdout,
			Stderr: stderr,
		})
		if exitCode != 0 {
			t.Fatalf("expected exit code 0 for %s, got %d", arg, exitCode)
		}
		if !strings.Contains(stdout.String(), "用法: videopress.exe") {
			t.Fatalf("expected help output for %s, got %q", arg, stdout.String())
		}
	}
}

func TestExecuteUnknownFlagShowsHelp(t *testing.T) {
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{"--unknown"}, Dependencies{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "未知选项") {
		t.Fatalf("expected unknown flag message, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "用法: videopress.exe") {
		t.Fatalf("expected help in stderr, got %s", stderr.String())
	}
}

func TestExecuteReturnsNonZeroWhenAllInputsSkipped(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{`C:\videos\readme.txt`}, Dependencies{
		ExecutableDir:  `C:\tools`,
		ExecutablePath: `C:\tools\videopress.exe`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		},
		Stdout: stdout,
		Stderr: stderr,
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "跳过非视频文件") {
		t.Fatalf("expected skip message, got %s", stdout.String())
	}
}

func TestExecuteFailsWhenInputNotAccessible(t *testing.T) {
	stderr := &bytes.Buffer{}

	exitCode := Execute([]string{`C:\videos\clip.mp4`}, Dependencies{
		ExecutableDir:  `C:\tools`,
		ExecutablePath: `C:\tools\videopress.exe`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		},
		InputAccessible: func(path string) bool { return false },
		Stdout:          &bytes.Buffer{},
		Stderr:          stderr,
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "输入文件不存在或不可读") {
		t.Fatalf("expected inaccessible message, got %s", stderr.String())
	}
}
