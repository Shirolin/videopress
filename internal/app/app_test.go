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
		PathExists: func(path string) bool { return false },
		Stdout:     stdout,
		Stderr:     stderr,
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
		MkdirAll:   func(path string, perm os.FileMode) error { return nil },
		PathExists: func(path string) bool { return false },
		Stdout:     stdout,
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
		MkdirAll:   func(path string, perm os.FileMode) error { return nil },
		PathExists: func(path string) bool { return false },
		Stdout:     &bytes.Buffer{},
		Stderr:     stderr,
	})

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "ffmpeg failed") {
		t.Fatalf("expected failure message, got %s", stderr.String())
	}
}
